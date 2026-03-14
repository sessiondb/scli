package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// runLogs is a thin wrapper around journalctl for bare-metal installs.
// It shows logs for the sessiondb systemd service.
func runLogs(lines int, follow bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("logs command is only supported on Linux/systemd installations")
	}

	args := []string{"-u", "sessiondb.service", "--no-pager"}
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

