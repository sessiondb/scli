package main

import (
	"fmt"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runInstall downloads the given version and extracts to workDir/sessiondb/.
// If workDir is empty, uses configDir/sessiondb-install or current dir.
func runInstall(version string, workDir string, configDir string) error {
	if version == "" {
		return fmt.Errorf("version required (e.g. 1.0.1 or v1.0.1)")
	}
	// Pass version as-is to get() so both "releases/v0.0.1/binaries/" and "releases/0.0.1/binaries/" are tried
	if workDir == "" {
		if configDir != "" {
			workDir = filepath.Join(configDir, "sessiondb-install")
		} else {
			workDir = filepath.Join(config.DefaultConfigDir(), "sessiondb-install")
		}
	}
	workDir, _ = filepath.Abs(workDir)
	return get(version, workDir)
}
