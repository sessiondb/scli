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

// run starts the requested/current version in background and follows logs.
// Component: api, ui, or all. Uses init-generated global config.
func run(version string, workDir string, configDir string, component string) error {
	return start(version, workDir, configDir, true, component)
}

// runStart starts the requested/current version in background without following logs.
func runStart(version string, workDir string, configDir string, component string) error {
	return start(version, workDir, configDir, false, component)
}

// start ensures the version is present, loads init config, starts the requested component(s)
// in background, and optionally tails logs.
func start(version string, workDir string, configDir string, followLogs bool, component string) error {
	if component == "" {
		component = ComponentAPI
	}
	if !validComponent(component) {
		return fmt.Errorf("invalid component %q; use api, ui, or all", component)
	}
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	defaultRoot := getInstallRoot("")
	defaultRoot, _ = filepath.Abs(defaultRoot)
	workDir, _ = filepath.Abs(workDir)

	// Prefer systemd when installed and default install root.
	if version == "" && workDir == defaultRoot && runtime.GOOS == "linux" {
		if component == ComponentAPI && systemdUnitInstalled(systemdAPIServiceName) {
			if err := startSystemdServiceUnit(systemdAPIServiceName); err != nil {
				return err
			}
			if followLogs {
				return runLogsWithComponent(100, true, ComponentAPI)
			}
			return nil
		}
		if component == ComponentUI && systemdUnitInstalled(systemdUIServiceName) {
			if err := startSystemdServiceUnit(systemdUIServiceName); err != nil {
				return err
			}
			if followLogs {
				return runLogsWithComponent(100, true, ComponentUI)
			}
			return nil
		}
		if component == ComponentAll {
			if systemdUnitInstalled(systemdAPIServiceName) {
				_ = startSystemdServiceUnit(systemdAPIServiceName)
			}
			if systemdUnitInstalled(systemdUIServiceName) {
				_ = startSystemdServiceUnit(systemdUIServiceName)
			}
			if followLogs {
				return runLogsWithComponent(100, true, ComponentAPI)
			}
			return nil
		}
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
	uiBin := filepath.Join(sessiondbDir, "ui", uiBinaryName)
	if component == ComponentAPI || component == ComponentAll {
		if _, err := os.Stat(setup); os.IsNotExist(err) {
			if version != "" {
				if err := get(version, workDir, false); err != nil {
					return err
				}
				sessiondbDir = filepath.Join(workDir, "versions", version)
				setup = filepath.Join(sessiondbDir, "setup.sh")
				uiBin = filepath.Join(sessiondbDir, "ui", uiBinaryName)
			} else {
				return fmt.Errorf("no setup.sh at %s; run scli install <version> first", sessiondbDir)
			}
		}
	}
	if component == ComponentUI || component == ComponentAll {
		if _, err := os.Stat(uiBin); os.IsNotExist(err) {
			return fmt.Errorf("UI binary not found at %s; install a release that includes sessiondb-ui-<os>-<arch>", uiBin)
		}
	}

	envSlice, err := config.EnvSliceFromInitFiles(os.Environ(), configDir)
	if err != nil {
		return err
	}
	logsDir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	startAPI := component == ComponentAPI || component == ComponentAll
	startUI := component == ComponentUI || component == ComponentAll

	if startAPI {
		pidPath := filepath.Join(configDir, pidFileAPI)
		logPath := filepath.Join(logsDir, runLogFileAPI)
		if pidBytes, err := os.ReadFile(pidPath); err == nil {
			if pid, err := strconv.Atoi(string(pidBytes)); err == nil && processExists(pid) {
				fmt.Fprintf(os.Stderr, "API already running (PID %d). Use 'scli stop' or 'scli logs -f --component api'.\n", pid)
				if followLogs {
					return tailLogFile(logPath)
				}
				startAPI = false
			} else {
				_ = os.Remove(pidPath)
			}
		}
		if startAPI {
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("open API log file: %w", err)
			}
			cmd := exec.Command(setup)
			cmd.Dir = sessiondbDir
			cmd.Env = envSlice
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			if err := cmd.Start(); err != nil {
				logFile.Close()
				return fmt.Errorf("start API: %w", err)
			}
			_ = os.WriteFile(filepath.Join(configDir, pidFileAPI), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
			logFile.Close()
			fmt.Fprintf(os.Stderr, "API started (PID %d). Logs: %s\n", cmd.Process.Pid, logPath)
		}
	}
	if startUI {
		pidPath := filepath.Join(configDir, pidFileUI)
		logPath := filepath.Join(logsDir, runLogFileUI)
		if pidBytes, err := os.ReadFile(pidPath); err == nil {
			if pid, err := strconv.Atoi(string(pidBytes)); err == nil && processExists(pid) {
				fmt.Fprintf(os.Stderr, "UI already running (PID %d). Use 'scli stop' or 'scli logs -f --component ui'.\n", pid)
				if followLogs && !startAPI {
					return tailLogFile(logPath)
				}
				startUI = false
			} else {
				_ = os.Remove(pidPath)
			}
		}
		if startUI {
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("open UI log file: %w", err)
			}
			cmd := exec.Command(uiBin)
			cmd.Dir = sessiondbDir
			cmd.Env = envSlice
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			if err := cmd.Start(); err != nil {
				logFile.Close()
				return fmt.Errorf("start UI: %w", err)
			}
			_ = os.WriteFile(filepath.Join(configDir, pidFileUI), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
			logFile.Close()
			fmt.Fprintf(os.Stderr, "UI started (PID %d). Logs: %s\n", cmd.Process.Pid, logPath)
		}
	}

	if !followLogs {
		fmt.Fprintln(os.Stderr, "Use 'scli logs -f --component api' or 'scli logs -f --component ui' to follow logs.")
		return nil
	}
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop following logs. Use 'scli stop' to stop.")
	if component == ComponentUI {
		return tailLogFile(filepath.Join(logsDir, runLogFileUI))
	}
	return tailLogFile(filepath.Join(logsDir, runLogFileAPI))
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
