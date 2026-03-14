package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// systemdUnitName is the name of the systemd service unit.
const systemdUnitName = "sessiondb.service"

// stopSystemdService runs "systemctl stop sessiondb.service" on Linux. Returns nil if stopped or not present.
func stopSystemdService() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd only supported on Linux")
	}
	cmd := exec.Command("systemctl", "stop", systemdUnitName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// startSystemdService runs "systemctl start sessiondb.service" on Linux.
func startSystemdService() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd only supported on Linux")
	}
	cmd := exec.Command("systemctl", "start", systemdUnitName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// systemdServiceActive returns true if the sessiondb service is active (running).
func systemdServiceActive() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	cmd := exec.Command("systemctl", "is-active", "--quiet", systemdUnitName)
	err := cmd.Run()
	return err == nil
}

// systemdUnitInstalled returns true when sessiondb.service is installed on this host.
func systemdUnitInstalled() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	cmd := exec.Command("systemctl", "cat", systemdUnitName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// tryRestartSystemdAfterInstall stops the sessiondb service if active, then starts it.
// Call after updating the "current" symlink so the new version runs. No-op if not Linux.
// If the unit is not installed, start will fail and a warning is printed.
func tryRestartSystemdAfterInstall() {
	if runtime.GOOS != "linux" {
		return
	}
	if !systemdUnitInstalled() {
		return
	}
	active := systemdServiceActive()
	if !active {
		// Do not auto-start a stopped unit after install; keep existing service state.
		return
	}
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Installed new version; restart service with: sudo systemctl restart %s\n", systemdUnitName)
		return
	}
	if active {
		if err := stopSystemdService(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not stop sessiondb service: %v\n", err)
			return
		}
	}
	if err := startSystemdService(); err != nil {
		// Unit may not be installed; only warn if it was active (we tried to restart).
		if active {
			fmt.Fprintf(os.Stderr, "Warning: could not start sessiondb service: %v\n", err)
		}
		return
	}
	if active {
		fmt.Println("SessionDB service restarted with new version.")
	} else {
		fmt.Println("SessionDB service started.")
	}
}
