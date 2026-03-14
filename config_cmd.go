package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runConfigView prints a human-friendly view of the current configuration (config.toml or .env).
func runConfigView(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	tomlPath := config.ConfigTOMLPath(configDir)
	envPath := config.EnvPath(configDir)

	tomlCfg, err := config.LoadConfigTOML(tomlPath)
	if err != nil {
		return fmt.Errorf("load config from %s: %w", tomlPath, err)
	}
	if tomlCfg != nil {
		fmt.Println("Config directory:", configDir)
		fmt.Println("config.toml:")
		fmt.Printf("  [server] port=%s mode=%s\n", tomlCfg.Server.Port, tomlCfg.Server.Mode)
		fmt.Printf("  [database] host=%s port=%s user=%s name=%s\n", tomlCfg.Database.Host, tomlCfg.Database.Port, tomlCfg.Database.User, tomlCfg.Database.Name)
		fmt.Printf("  [redis] addr=%s db=%d\n", tomlCfg.Redis.Addr, tomlCfg.Redis.DB)
		fmt.Printf("  [ui] api_url=%s\n", tomlCfg.UI.APIURL)
		fmt.Printf("  [auth] default_logins=%d\n", len(tomlCfg.Auth.DefaultLogins))
		if tomlCfg.Secrets.MigrateToken != "" {
			fmt.Println("  [secrets] migrate_token=(set)")
		} else {
			fmt.Println("  [secrets] migrate_token=(not set)")
		}
		if tomlCfg.Secrets.DBCredentialEncryptionKey != "" {
			fmt.Println("  [secrets] db_credential_encryption_key=(set)")
		} else {
			fmt.Println("  [secrets] db_credential_encryption_key=(not set)")
		}
		return nil
	}

	cfg, err := config.LoadEnvConfig(envPath)
	if err != nil {
		return fmt.Errorf("load config from %s: %w", envPath, err)
	}
	fmt.Println("Config directory:", configDir)
	fmt.Println(".env:")
	fmt.Printf("  SERVER_PORT=%s SERVER_MODE=%s\n", cfg.ServerPort, cfg.ServerMode)
	fmt.Printf("  DB_HOST=%s DB_PORT=%s DB_USER=%s DB_NAME=%s\n", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	fmt.Printf("  REDIS_ADDR=%s REDIS_DB=%s\n", cfg.RedisAddr, cfg.RedisDB)
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

// runConfigEdit opens config.toml (or .env if no config.toml) in the user's editor.
func runConfigEdit(configDir string) error {
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	tomlPath := config.ConfigTOMLPath(configDir)
	envPath := config.EnvPath(configDir)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	editPath := tomlPath
	if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
		editPath = envPath
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			if err := os.WriteFile(envPath, []byte{}, 0o644); err != nil {
				return fmt.Errorf("create env file: %w", err)
			}
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, editPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	if editPath == tomlPath {
		if _, err := config.LoadConfigTOML(tomlPath); err != nil {
			return fmt.Errorf("after editing, failed to parse %s: %w", tomlPath, err)
		}
	} else {
		if _, err := config.LoadEnvConfig(envPath); err != nil {
			return fmt.Errorf("after editing, failed to parse %s: %w", envPath, err)
		}
	}
	return nil
}

