package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sessiondb/scli/internal/config"
)

// runInstall downloads the given version and extracts to workDir/sessiondb/.
// If workDir is empty, uses configDir/sessiondb-install or current dir.
func runInstall(version string, workDir string, configDir string) error {
	if version == "" {
		return fmt.Errorf("version required (e.g. 1.0.1 or v1.0.1)")
	}
	version = strings.TrimPrefix(version, "v")
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
