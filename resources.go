package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runResources prints key SessionDB resources for bare-metal installs: install root, current version,
// server binary, UI dist, config files, and systemd unit if present.
func runResources(configDir string, installRootOverride string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	tomlPath := config.ConfigTOMLPath(configDir)
	envPath := config.EnvPath(configDir)
	configYAMLPath := config.ConfigYAMLPath(configDir)

	installRoot := installRootOverride
	if installRoot == "" {
		installRoot = getInstallRoot("")
	}
	installRoot, _ = filepath.Abs(installRoot)

	currentPath := filepath.Join(installRoot, "current")
	currentInfo, _ := os.Lstat(currentPath)
	version := ""
	if currentInfo != nil && (currentInfo.Mode()&os.ModeSymlink) != 0 {
		if target, err := os.Readlink(currentPath); err == nil {
			version = filepath.Base(target)
		}
	}

	fmt.Println("Install root:", installRoot)
	if version != "" {
		fmt.Println("Current version:", version)
	} else {
		fmt.Println("Current version: (not set; current symlink missing)")
	}
	fmt.Println()

	serverBin := filepath.Join(currentPath, "server", "sessiondb-server")
	uiDist := filepath.Join(currentPath, "ui", "dist", "index.html")
	uiBin := filepath.Join(currentPath, "ui", "sessiondb-ui")
	setupScript := filepath.Join(currentPath, "setup.sh")
	unitPath := "/etc/systemd/system/sessiondb.service"

	fmt.Println("Backend:")
	fmt.Printf("  server binary: %s", serverBin)
	if _, err := os.Stat(serverBin); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Printf("  setup script:  %s", setupScript)
	if _, err := os.Stat(setupScript); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("Frontend:")
	fmt.Printf("  ui dist:       %s", uiDist)
	if _, err := os.Stat(uiDist); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Printf("  ui binary:     %s", uiBin)
	if _, err := os.Stat(uiBin); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("Config:")
	fmt.Println("  config dir:    " + configDir)
	fmt.Printf("  config.toml:   %s", tomlPath)
	if _, err := os.Stat(tomlPath); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Printf("  .env:          %s (generated for systemd/backend)", envPath)
	if _, err := os.Stat(envPath); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Printf("  config.yaml:   %s (legacy)", configYAMLPath)
	if _, err := os.Stat(configYAMLPath); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("Systemd:")
	fmt.Printf("  unit file:     %s", unitPath)
	if _, err := os.Stat(unitPath); err != nil {
		fmt.Printf(" (missing: %v)", err)
	}
	fmt.Println()

	return nil
}
