package main

import (
	"path/filepath"
)

// runInstall downloads the given version (or latest if empty) from GitHub Releases and installs under install root.
// workDir: if set, used as install root (versions/ and current go here); otherwise getInstallRoot() is used.
func runInstall(version string, workDir string, _ string) error {
	if workDir == "" {
		workDir = getInstallRoot("")
	}
	workDir, _ = filepath.Abs(workDir)
	return get(version, workDir)
}
