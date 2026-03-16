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
	tomlPath := config.ConfigTOMLPath(configDir)
	envPath := config.EnvPath(configDir)

	fmt.Println("Initializing SessionDB configuration...")
	fmt.Println()
	fmt.Println("Enter database configuration")
	fmt.Println()

	cfg, err := config.RunPrompts()
	if err != nil {
		return err
	}

	// Preserve existing secrets from config.toml or .env
	existingToml, _ := config.LoadConfigTOML(tomlPath)
	existingEnv, _ := config.ReadEnv(envPath)
	encKeyExists := (existingToml != nil && existingToml.Secrets.DBCredentialEncryptionKey != "") || existingEnv["DB_CREDENTIAL_ENCRYPTION_KEY"] != ""
	tokenExists := (existingToml != nil && existingToml.Secrets.MigrateToken != "") || existingEnv["MIGRATE_TOKEN"] != ""

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
		if existingToml != nil {
			cfg.DBCredentialEncryptionKey = existingToml.Secrets.DBCredentialEncryptionKey
		} else {
			cfg.DBCredentialEncryptionKey = existingEnv["DB_CREDENTIAL_ENCRYPTION_KEY"]
		}
	}
	if !tokenExists {
		token, err := utils.GenerateToken()
		if err != nil {
			return fmt.Errorf("generate migrate token: %w", err)
		}
		cfg.MigrateToken = token
		fmt.Println("✓ MIGRATE_TOKEN generated")
	} else {
		if existingToml != nil {
			cfg.MigrateToken = existingToml.Secrets.MigrateToken
		} else {
			cfg.MigrateToken = existingEnv["MIGRATE_TOKEN"]
		}
	}
	if !encKeyExists || !tokenExists {
		fmt.Println()
	}

	tomlCfg := config.EnvConfigToToml(cfg)
	if err := config.WriteConfigTOML(tomlPath, tomlCfg); err != nil {
		return fmt.Errorf("write config.toml: %w", err)
	}
	// Generate .env from TOML so systemd (deploy) and backend keep working
	if err := config.WriteEnv(envPath, config.TomlToEnvConfig(tomlCfg)); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	fmt.Println("Configuration saved to:")
	fmt.Println()
	fmt.Println("  " + tomlPath)
	fmt.Println("  ( .env generated for systemd / backend )")
	fmt.Println()
	fmt.Println("When you use 'scli run' or 'scli start', SESSIONDB_CONFIG_DIR is set automatically to this directory.")
	fmt.Println("You do not need to set SESSIONDB_CONFIG_DIR manually. Deploy-generated systemd units set it too.")
	fmt.Println()
	return nil
}
