package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// run ensures the version is present under workDir (calls get if needed), loads sessiondb.yaml,
// injects env vars from it, and runs setup.sh. Supports install-root layout (versions/<tag>, current symlink) and legacy (workDir/sessiondb).
func run(version string, workDir string) error {
	workDir, _ = filepath.Abs(workDir)
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

	cmd := exec.Command(setup)
	cmd.Dir = sessiondbDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	configPath := filepath.Join(sessiondbDir, "sessiondb.yaml")
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	if cfg != nil && len(cfg.Env) > 0 {
		cmd.Env = envSlice(os.Environ(), cfg.Env)
	}

	return cmd.Run()
}
