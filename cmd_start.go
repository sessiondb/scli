package main

import "path/filepath"

// runStartCommand starts SessionDB in the background using the requested/current version.
// It does not attach a tail; use "scli logs -f" to follow logs.
func runStartCommand(version string, workDir string, configDir string) error {
	defaultRoot := getInstallRoot("")
	defaultRoot, _ = filepath.Abs(defaultRoot)
	workDir, _ = filepath.Abs(workDir)
	// Prefer systemd when installed and no explicit version/workdir is requested.
	// This ensures "start" always runs the current/latest installed version in service mode.
	if version == "" && workDir == defaultRoot && systemdUnitInstalled() {
		return startSystemdService()
	}
	return runStart(version, workDir, configDir)
}
