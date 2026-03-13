package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/sessiondb/scli/internal/config"
)

// runMigrate calls POST /v1/migrate with X-Migrate-Token from config.
func runMigrate(configDir string, baseURL string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	envPath := config.EnvPath(configDir)
	cfg, err := config.LoadEnvConfig(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found at %s (run 'scli init' first)", envPath)
		}
		return err
	}
	if cfg.MigrateToken == "" {
		return fmt.Errorf("MIGRATE_TOKEN not set in %s", envPath)
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	url := baseURL + "/v1/migrate"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(nil))
	if err != nil {
		return err
	}
	req.Header.Set("X-Migrate-Token", cfg.MigrateToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("migrate request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("migrate failed: HTTP %d", resp.StatusCode)
	}
	fmt.Println("✓ Migrations completed successfully")
	return nil
}
