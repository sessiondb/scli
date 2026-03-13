package main

import (
	"fmt"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
	"github.com/sessiondb/scli/internal/utils"
)

func runInit(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	envPath := config.EnvPath(configDir)
	configYAMLPath := config.ConfigYAMLPath(configDir)

	fmt.Println("Initializing SessionDB configuration...")
	fmt.Println()
	fmt.Println("Enter database configuration")
	fmt.Println()

	cfg, err := config.RunPrompts()
	if err != nil {
		return err
	}

	// Load existing .env if present to avoid regenerating secrets
	existing, _ := config.ReadEnv(envPath)
	encKeyExists := existing["DB_CREDENTIAL_ENCRYPTION_KEY"] != ""
	tokenExists := existing["MIGRATE_TOKEN"] != ""

	if !encKeyExists || !tokenExists {
		fmt.Println("Generating secure keys...")
		fmt.Println()
	}
	if !encKeyExists {
		key, err := utils.GenerateEncryptionKey()
		if err != nil {
			return fmt.Errorf("generate encryption key: %w", err)
		}
		cfg.DBCredentialEncryptionKey = key
		fmt.Println("✓ DB_CREDENTIAL_ENCRYPTION_KEY generated")
	} else {
		cfg.DBCredentialEncryptionKey = existing["DB_CREDENTIAL_ENCRYPTION_KEY"]
	}
	if !tokenExists {
		token, err := utils.GenerateToken()
		if err != nil {
			return fmt.Errorf("generate migrate token: %w", err)
		}
		cfg.MigrateToken = token
		fmt.Println("✓ MIGRATE_TOKEN generated")
	} else {
		cfg.MigrateToken = existing["MIGRATE_TOKEN"]
	}
	if !encKeyExists || !tokenExists {
		fmt.Println()
	}

	if err := config.WriteEnv(envPath, cfg); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}
	if err := config.WriteConfigYAML(configYAMLPath, cfg); err != nil {
		return fmt.Errorf("write config.yaml: %w", err)
	}

	fmt.Println("Configuration saved to:")
	fmt.Println()
	fmt.Println("  " + envPath)
	fmt.Println("  " + configYAMLPath)
	fmt.Println()
	return nil
}
