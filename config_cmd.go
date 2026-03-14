package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runConfigView prints a human-friendly view of the current configuration from .env.
func runConfigView(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	envPath := config.EnvPath(configDir)

	cfg, err := config.LoadEnvConfig(envPath)
	if err != nil {
		return fmt.Errorf("load config from %s: %w", envPath, err)
	}

	fmt.Println("Config directory:", configDir)
	fmt.Println(".env:")
	fmt.Printf("  SERVER_PORT=%s\n", cfg.ServerPort)
	fmt.Printf("  SERVER_MODE=%s\n", cfg.ServerMode)
	fmt.Printf("  DB_HOST=%s\n", cfg.DBHost)
	fmt.Printf("  DB_PORT=%s\n", cfg.DBPort)
	fmt.Printf("  DB_USER=%s\n", cfg.DBUser)
	fmt.Printf("  DB_NAME=%s\n", cfg.DBName)
	fmt.Printf("  DB_SSLMODE=%s\n", cfg.DBSSLMode)
	fmt.Printf("  REDIS_ADDR=%s\n", cfg.RedisAddr)
	fmt.Printf("  REDIS_DB=%s\n", cfg.RedisDB)
	fmt.Printf("  JWT_EXPIRY_HOURS=%s\n", cfg.JWTExpiryHours)
	fmt.Printf("  JWT_REFRESH_EXPIRY=%s\n", cfg.JWTRefreshExpiry)
	if cfg.MigrateToken != "" {
		fmt.Println("  MIGRATE_TOKEN=(set)")
	} else {
		fmt.Println("  MIGRATE_TOKEN=(not set)")
	}
	if cfg.DBCredentialEncryptionKey != "" {
		fmt.Println("  DB_CREDENTIAL_ENCRYPTION_KEY=(set)")
	} else {
		fmt.Println("  DB_CREDENTIAL_ENCRYPTION_KEY=(not set)")
	}

	return nil
}

// runConfigEdit opens the .env file in the user's editor (EDITOR/ VISUAL / nano / vi).
func runConfigEdit(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	envPath := config.EnvPath(configDir)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if err := os.WriteFile(envPath, []byte{}, 0o644); err != nil {
			return fmt.Errorf("create env file: %w", err)
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, envPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	// Basic sanity check: ensure essential fields are at least present (may be empty).
	if _, err := config.LoadEnvConfig(envPath); err != nil {
		return fmt.Errorf("after editing, failed to parse %s: %w", envPath, err)
	}

	return nil
}

