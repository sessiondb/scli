package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// stopSystemdServiceUnit runs "systemctl stop <unit>" on Linux.
func stopSystemdServiceUnit(unit string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd only supported on Linux")
	}
	cmd := exec.Command("systemctl", "stop", unit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// startSystemdServiceUnit runs "systemctl start <unit>" on Linux.
func startSystemdServiceUnit(unit string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("systemd only supported on Linux")
	}
	cmd := exec.Command("systemctl", "start", unit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// systemdServiceActive returns true if the given unit is active (running).
func systemdServiceActive(unit string) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	cmd := exec.Command("systemctl", "is-active", "--quiet", unit)
	return cmd.Run() == nil
}

// systemdUnitInstalled returns true when the given unit is installed on this host.
func systemdUnitInstalled(unit string) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	cmd := exec.Command("systemctl", "cat", unit)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// stopSystemdService stops the API systemd service (backward compatible).
func stopSystemdService() error {
	return stopSystemdServiceUnit(systemdAPIServiceName)
}

// startSystemdService starts the API systemd service (backward compatible).
func startSystemdService() error {
	return startSystemdServiceUnit(systemdAPIServiceName)
}

// systemdUnitInstalledAPI returns true when sessiondb.service is installed.
func systemdUnitInstalledAPI() bool {
	return systemdUnitInstalled(systemdAPIServiceName)
}

// tryRestartSystemdAfterInstall stops API and UI services if active, then starts them.
// No-op if not Linux or not root; non-root gets a message to restart manually.
func tryRestartSystemdAfterInstall() {
	if runtime.GOOS != "linux" {
		return
	}
	units := []string{systemdAPIServiceName, systemdUIServiceName}
	var toRestart []string
	for _, unit := range units {
		if !systemdUnitInstalled(unit) {
			continue
		}
		if systemdServiceActive(unit) {
			toRestart = append(toRestart, unit)
		}
	}
	if len(toRestart) == 0 {
		return
	}
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Installed new version; restart with: sudo systemctl restart %s", systemdAPIServiceName)
		if systemdUnitInstalled(systemdUIServiceName) {
			fmt.Fprintf(os.Stderr, " %s", systemdUIServiceName)
		}
		fmt.Fprintln(os.Stderr)
		return
	}
	for _, unit := range toRestart {
		_ = stopSystemdServiceUnit(unit)
	}
	for _, unit := range toRestart {
		if err := startSystemdServiceUnit(unit); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not start %s: %v\n", unit, err)
		}
	}
	fmt.Println("SessionDB service(s) restarted with new version.")
}
