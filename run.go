package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/sessiondb/scli/internal/config"
)

// pidFileName is the name of the PID file for the server started by "scli run" (non-systemd).
const pidFileName = "sessiondb.pid"

// runLogFileName is the name of the log file when running server in background via "scli run".
const runLogFileName = "sessiondb.log"

// run starts the requested/current version in background and follows logs.
// It uses init-generated global config (.env preferred, config.yaml fallback).
// Ctrl+C only stops the log tail; use "scli stop" to stop the server.
func run(version string, workDir string, configDir string) error {
	return start(version, workDir, configDir, true)
}

// runStart starts the requested/current version in background without following logs.
// Use this for automation or detached starts.
func runStart(version string, workDir string, configDir string) error {
	return start(version, workDir, configDir, false)
}

// start ensures the version is present under workDir (calls get if needed), loads init config,
// starts the server in background, and optionally tails logs.
func start(version string, workDir string, configDir string, followLogs bool) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	defaultRoot := getInstallRoot("")
	defaultRoot, _ = filepath.Abs(defaultRoot)

	workDir, _ = filepath.Abs(workDir)
	// If systemd is installed and caller uses the default install root/current version,
	// prefer systemd as the single runtime manager.
	if version == "" && workDir == defaultRoot && systemdUnitInstalled() {
		if err := startSystemdService(); err != nil {
			return err
		}
		if followLogs {
			return runLogs(100, true)
		}
		return nil
	}
	var sessiondbDir string
	if version != "" {
		version = normalizeTag(version)
		sessiondbDir = filepath.Join(workDir, "versions", version)
	} else {
		currentLink := filepath.Join(workDir, "current")
		if info, err := os.Lstat(currentLink); err == nil && (info.Mode()&os.ModeSymlink) != 0 {
			resolved, err := filepath.EvalSymlinks(currentLink)
			if err == nil {
				sessiondbDir = resolved
			} else {
				sessiondbDir = currentLink
			}
		} else {
			sessiondbDir = filepath.Join(workDir, "sessiondb")
		}
	}

	setup := filepath.Join(sessiondbDir, "setup.sh")
	if _, err := os.Stat(setup); os.IsNotExist(err) {
		if version != "" {
			if err := get(version, workDir); err != nil {
				return err
			}
			sessiondbDir = filepath.Join(workDir, "versions", version)
			setup = filepath.Join(sessiondbDir, "setup.sh")
		} else {
			return fmt.Errorf("no setup.sh at %s; run scli install <version> first", sessiondbDir)
		}
	}

	// Use init-generated global config as source of truth:
	// .env preferred; if missing, fallback to config.yaml and recreate .env.
	envSlice, err := config.EnvSliceFromInitFiles(os.Environ(), configDir)
	if err != nil {
		return err
	}

	pidPath := filepath.Join(configDir, pidFileName)
	logsDir := filepath.Join(configDir, "logs")
	logPath := filepath.Join(logsDir, runLogFileName)

	// If PID file exists and process is running, do not start again.
	if pidBytes, err := os.ReadFile(pidPath); err == nil {
		if pid, err := strconv.Atoi(string(pidBytes)); err == nil && processExists(pid) {
			fmt.Fprintf(os.Stderr, "Server already running (PID %d). Use 'scli stop' to stop or 'scli logs -f' to follow logs.\n", pid)
			fmt.Fprintf(os.Stderr, "Log file: %s\n", logPath)
			if followLogs {
				// Still run tail -f so user can watch logs.
				return tailLogFile(logPath)
			}
			return nil
		}
		_ = os.Remove(pidPath)
	}

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	cmd := exec.Command(setup)
	cmd.Dir = sessiondbDir
	cmd.Env = envSlice
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start server: %w", err)
	}

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		logFile.Close()
		_ = cmd.Process.Kill()
		return fmt.Errorf("write PID file: %w", err)
	}
	// Close the log file in this process so the server keeps writing; tail will read independently.
	logFile.Close()

	fmt.Fprintf(os.Stderr, "Server started in background (PID %d). Logs: %s\n", cmd.Process.Pid, logPath)
	if !followLogs {
		fmt.Fprintln(os.Stderr, "Use 'scli logs -f' to follow service logs, or 'tail -f' on the run log file.")
		return nil
	}
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop following logs (server keeps running). Use 'scli stop' to stop the server.")
	// Follow logs; Ctrl+C stops only the tail.
	return tailLogFile(logPath)
}

// tailLogFile runs "tail -f" on the given path (Unix). On Windows or if tail fails, opens and reads the file.
func tailLogFile(logPath string) error {
	if runtime.GOOS == "windows" {
		// Windows has no tail -f; just inform.
		fmt.Fprintf(os.Stderr, "Log file: %s\n", logPath)
		return nil
	}
	cmd := exec.Command("tail", "-f", logPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// processExists returns true if the given PID exists (Unix). On Windows, assumes exists if PID file present.
func processExists(pid int) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	// kill -0 checks process existence without sending a terminating signal.
	cmd := exec.Command("kill", "-0", strconv.Itoa(pid))
	return cmd.Run() == nil
}
