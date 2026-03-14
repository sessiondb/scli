package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runPrune removes all SessionDB artifacts managed by scli from the host:
// install root (all versions/current/runtime files) and config dir (.env/config.yaml/logs/pid).
// Use force=true to confirm destructive cleanup.
func runPrune(configDir string, force bool) error {
	if !force {
		return fmt.Errorf("prune is destructive; re-run with --yes to remove all installed versions and config data")
	}

	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)

	installRoot := getInstallRoot("")
	installRoot, _ = filepath.Abs(installRoot)

	if err := removePathIfExists(installRoot); err != nil {
		return fmt.Errorf("remove install root %s: %w", installRoot, err)
	}
	fmt.Printf("Removed install root: %s\n", installRoot)

	if err := removePathIfExists(configDir); err != nil {
		return fmt.Errorf("remove config dir %s: %w", configDir, err)
	}
	fmt.Printf("Removed config directory: %s\n", configDir)

	return nil
}

// removePathIfExists removes path recursively and succeeds if path does not exist.
func removePathIfExists(path string) error {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
