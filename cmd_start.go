package main

import (
	"path/filepath"
	"runtime"
)

// runStartCommand starts SessionDB in the background. Component: api, ui, or all.
func runStartCommand(version string, workDir string, configDir string, component string) error {
	if component == "" {
		component = ComponentAPI
	}
	defaultRoot := getInstallRoot("")
	defaultRoot, _ = filepath.Abs(defaultRoot)
	workDir, _ = filepath.Abs(workDir)
	if version == "" && workDir == defaultRoot && runtime.GOOS == "linux" {
		if component == ComponentAPI && systemdUnitInstalled(systemdAPIServiceName) {
			return startSystemdServiceUnit(systemdAPIServiceName)
		}
		if component == ComponentUI && systemdUnitInstalled(systemdUIServiceName) {
			return startSystemdServiceUnit(systemdUIServiceName)
		}
		if component == ComponentAll {
			if systemdUnitInstalled(systemdAPIServiceName) {
				_ = startSystemdServiceUnit(systemdAPIServiceName)
			}
			if systemdUnitInstalled(systemdUIServiceName) {
				_ = startSystemdServiceUnit(systemdUIServiceName)
			}
			return nil
		}
	}
	return runStart(version, workDir, configDir, component)
}
