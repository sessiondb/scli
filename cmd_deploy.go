package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runDeploy generates a systemd unit file for bare metal using the global .env and a single
// version via the "current" symlink. Uses absolute paths and WorkingDirectory so the service
// starts reliably. When you install a new version, the symlink is updated; stop then start
// the service (or use "scli install" which restarts the service) so the new version runs.
func runDeploy(configDir string, platform string, outputPath string) error {
	if platform != "baremetal" && platform != "" {
		return fmt.Errorf("unsupported platform: %s (only 'baremetal' is supported)", platform)
	}
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	envPath := config.EnvPath(configDir)
	configYAMLPath := config.ConfigYAMLPath(configDir)
	installRoot := getInstallRoot("")
	installRoot, _ = filepath.Abs(installRoot)
	binaryPath := filepath.Join(installRoot, "current", "server", "sessiondb-server")
	workDir := filepath.Join(installRoot, "current")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// Fallback: recreate .env from config.yaml from init so deploy uses a single EnvironmentFile.
		cfg, loadErr := config.LoadConfigYAML(configYAMLPath)
		if loadErr == nil {
			if writeErr := config.WriteEnv(envPath, cfg); writeErr != nil {
				return fmt.Errorf("recreate .env from %s: %w", configYAMLPath, writeErr)
			}
			fmt.Printf("Recreated %s from %s\n", envPath, configYAMLPath)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: %s does not exist and %s could not be loaded. Run 'scli init' first.\n", envPath, configYAMLPath)
		}
	}
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: server binary not found at %s. Run 'scli install' first.\n", binaryPath)
	}

	if outputPath == "" {
		outputPath = "sessiondb.service"
	}

	// Use absolute paths so the unit works regardless of who runs systemctl.
	unit := `[Unit]
Description=SessionDB Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
WorkingDirectory=` + workDir + `
EnvironmentFile=` + envPath + `
ExecStart=` + binaryPath + `
Restart=on-failure
RestartSec=5
TimeoutStartSec=30

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile(outputPath, []byte(unit), 0644); err != nil {
		return err
	}
	fmt.Printf("Generated %s\n", outputPath)
	fmt.Println()
	fmt.Println("Paths used:")
	fmt.Printf("  WorkingDirectory: %s\n", workDir)
	fmt.Printf("  EnvironmentFile:  %s\n", envPath)
	fmt.Printf("  ExecStart:        %s\n", binaryPath)
	fmt.Println()
	fmt.Println("To install and start:")
	fmt.Println("  sudo cp " + outputPath + " /etc/systemd/system/")
	fmt.Println("  sudo systemctl daemon-reload")
	fmt.Println("  sudo systemctl enable sessiondb")
	fmt.Println("  sudo systemctl start sessiondb")
	fmt.Println()
	fmt.Println("After upgrading with 'scli install <version>', restart so the new version runs:")
	fmt.Println("  sudo systemctl restart sessiondb")
	return nil
}
