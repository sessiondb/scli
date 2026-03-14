package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/sessiondb/scli/internal/config"
)

// runStop stops all currently running SessionDB processes:
// 1) process tracked by PID file from "scli run"/"scli start",
// 2) systemd service process (if active),
// 3) any leftover sessiondb-server processes.
func runStop(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	pidPath := filepath.Join(configDir, pidFileName)
	installRoot := getInstallRoot("")
	installRoot, _ = filepath.Abs(installRoot)
	stoppedAny := false

	// Stop the process started by "scli run"/"scli start" if PID file exists.
	if data, err := os.ReadFile(pidPath); err == nil {
		pid, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("invalid PID in %s: %w", pidPath, err)
		}
		if !processExists(pid) {
			os.Remove(pidPath)
			fmt.Println("Server was not running (stale PID file removed).")
			return nil
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("find process %d: %w", pid, err)
		}
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("stop server (PID %d): %w", pid, err)
		}
		stoppedAny = true
		if err := os.Remove(pidPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove PID file %s: %v\n", pidPath, err)
		}
		fmt.Printf("Stopped server (PID %d).\n", pid)
	}

	// Stop systemd unit if active.
	if err := stopSystemdService(); err == nil {
		fmt.Println("Stopped sessiondb systemd service.")
		stoppedAny = true
	}

	// Stop any leftover SessionDB server processes.
	// We match paths under install root and generic sessiondb-server process names.
	if count, err := killMatchingProcesses([]string{
		filepath.Join(installRoot, "current", "server", "sessiondb-server"),
		filepath.Join(installRoot, "versions"),
		"sessiondb-server",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: process cleanup encountered an error: %v\n", err)
	} else if count > 0 {
		stoppedAny = true
		fmt.Printf("Stopped %d additional SessionDB process(es).\n", count)
	}

	if !stoppedAny {
		fmt.Println("No SessionDB server process is currently running.")
	}
	return nil
}

// killMatchingProcesses finds and kills processes matching each pattern using pgrep/kill.
// Returns count of successfully killed processes.
func killMatchingProcesses(patterns []string) (int, error) {
	total := 0
	for _, pattern := range patterns {
		// pgrep -f returns one PID per line and non-zero when no process is found.
		out, err := exec.Command("pgrep", "-f", pattern).Output()
		if err != nil {
			continue
		}
		lines := splitLines(string(out))
		for _, line := range lines {
			pid, convErr := strconv.Atoi(line)
			if convErr != nil || pid <= 0 {
				continue
			}
			proc, findErr := os.FindProcess(pid)
			if findErr != nil {
				continue
			}
			if killErr := proc.Kill(); killErr == nil {
				total++
			}
		}
	}
	return total, nil
}

// splitLines splits output into non-empty trimmed lines.
func splitLines(s string) []string {
	out := make([]string, 0)
	cur := ""
	for _, r := range s {
		if r == '\n' || r == '\r' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
