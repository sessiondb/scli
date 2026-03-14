package main

import (
	"path/filepath"
)

// runInstall downloads the given version (or latest if empty) from GitHub Releases and installs under install root.
// If the sessiondb systemd service is active, it is stopped and then started so the new version runs (version-based:
// one unit, "current" symlink points to the active version; previous version is stopped, new version runs).
// workDir: if set, used as install root (versions/ and current go here); otherwise getInstallRoot() is used.
// force: if true, reinstall even when the version is already installed (removes and re-downloads).
func runInstall(version string, workDir string, _ string, force bool) error {
	if workDir == "" {
		workDir = getInstallRoot("")
	}
	workDir, _ = filepath.Abs(workDir)
	if err := get(version, workDir, force); err != nil {
		return err
	}
	tryRestartSystemdAfterInstall()
	return nil
}
