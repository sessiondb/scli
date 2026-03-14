package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sessiondb/scli/internal/config"
)

// runLogsWithComponent shows logs for the given component (api or ui).
// Uses journalctl when systemd unit is present, otherwise tails the run log file.
func runLogsWithComponent(lines int, follow bool, component string) error {
	if component == "" {
		component = ComponentAPI
	}
	if component != ComponentAPI && component != ComponentUI {
		return fmt.Errorf("invalid component %q for logs; use api or ui", component)
	}
	if runtime.GOOS == "linux" {
		unit := systemdUnitName(component)
		if systemdUnitInstalled(unit) {
			args := []string{"-u", unit, "--no-pager"}
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
	}
	configDir := config.DefaultConfigDir()
	configDir, _ = filepath.Abs(configDir)
	logFile := runLogFileAPI
	if component == ComponentUI {
		logFile = runLogFileUI
	}
	logPath := filepath.Join(configDir, "logs", logFile)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found at %s (start with 'scli run --component %s' first)", logPath, component)
	}
	if !follow {
		data, err := os.ReadFile(logPath)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(data)
		return nil
	}
	return tailLogFile(logPath)
}

// runLogs is backward compatible: no component means API logs.
func runLogs(lines int, follow bool) error {
	return runLogsWithComponent(lines, follow, ComponentAPI)
}

