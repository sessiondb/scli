package main

import (
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runRestart stops then starts the requested component(s). Component: api, ui, or all.
func runRestart(version string, workDir string, configDir string, component string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	if workDir == "" {
		workDir = getInstallRoot("")
	}
	workDir, _ = filepath.Abs(workDir)
	if component == "" {
		component = ComponentAPI
	}
	if err := runStop(configDir); err != nil {
		return err
	}
	return runStartCommand(version, workDir, configDir, component)
}
