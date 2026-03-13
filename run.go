package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// run ensures the version is present under workDir (calls get if needed), loads sessiondb.yaml,
// injects env vars from it into the process environment, then runs setup.sh so server and UI get those values.
func run(version string, workDir string) error {
	sessiondbDir := filepath.Join(workDir, "sessiondb")
	setup := filepath.Join(sessiondbDir, "setup.sh")
	if _, err := os.Stat(setup); os.IsNotExist(err) {
		if err := get(version, workDir); err != nil {
			return err
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
