package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runReset removes scli local state: legacy sessiondb-install dir and release-based install root (versions/, current).
// If all is true, also removes .env and config.yaml from the config directory.
func runReset(configDir string, all bool) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	legacyInstallDir := filepath.Join(configDir, "sessiondb-install")
	if err := os.RemoveAll(legacyInstallDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove install dir %s: %w", legacyInstallDir, err)
	}
	fmt.Printf("Removed install directory: %s\n", legacyInstallDir)

	installRoot := getInstallRoot("")
	versionsDir := filepath.Join(installRoot, "versions")
	currentLink := filepath.Join(installRoot, "current")
	if err := os.RemoveAll(versionsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove versions dir %s: %w", versionsDir, err)
	}
	if err := os.Remove(currentLink); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove current symlink: %w", err)
	}
	fmt.Printf("Removed release install root: %s (versions/, current)\n", installRoot)

	if all {
		envPath := config.EnvPath(configDir)
		yamlPath := config.ConfigYAMLPath(configDir)
		for _, p := range []string{envPath, yamlPath} {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", p, err)
			}
			fmt.Printf("Removed: %s\n", p)
		}
	}

	return nil
}
