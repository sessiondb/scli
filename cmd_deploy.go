package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runDeploy generates a systemd unit file for bare metal that uses EnvironmentFile=/opt/sessiondb/.env.
func runDeploy(configDir string, platform string, outputPath string) error {
	if platform != "baremetal" && platform != "" {
		return fmt.Errorf("unsupported platform: %s (only 'baremetal' is supported)", platform)
	}
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	envPath := config.EnvPath(configDir)
	// Default deploy path for systemd
	if outputPath == "" {
		outputPath = "sessiondb.service"
	}
	unit := `[Unit]
Description=SessionDB Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
EnvironmentFile=` + envPath + `
ExecStart=%s/sessiondb
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`
	installDir := "/opt/sessiondb"
	if d := os.Getenv("SESSIONDB_INSTALL_DIR"); d != "" {
		installDir = d
	}
	unit = fmt.Sprintf(unit, installDir)
	if err := os.WriteFile(outputPath, []byte(unit), 0644); err != nil {
		return err
	}
	fmt.Printf("Generated %s\n", outputPath)
	fmt.Println()
	fmt.Println("To install:")
	fmt.Println("  sudo cp " + outputPath + " /etc/systemd/system/")
	fmt.Println("  sudo systemctl daemon-reload")
	fmt.Println("  sudo systemctl enable sessiondb")
	fmt.Println("  sudo systemctl start sessiondb")
	fmt.Println()
	fmt.Println("Ensure the server binary is at " + installDir + "/sessiondb and EnvironmentFile points to your .env")
	return nil
}
