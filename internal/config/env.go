// Package config provides env schema, prompts, and .env read/write for scli.
package config

import (
	"os"
	"path/filepath"
)

// EnvConfig holds all configuration: user-provided (DB, Redis, server, JWT) and auto-generated secrets.
// Used for .env and config.yaml storage. Matches sessiondb backend .env.
type EnvConfig struct {
	// Server
	ServerPort string `yaml:"server_port" env:"SERVER_PORT"`
	ServerMode string `yaml:"server_mode" env:"SERVER_MODE"`

	// Database (prompted)
	DBHost     string `yaml:"db_host" env:"DB_HOST"`
	DBPort     string `yaml:"db_port" env:"DB_PORT"`
	DBUser     string `yaml:"db_user" env:"DB_USER"`
	DBPassword string `yaml:"db_password" env:"DB_PASSWORD"`
	DBName     string `yaml:"db_name" env:"DB_NAME"`
	DBSSLMode  string `yaml:"db_sslmode" env:"DB_SSLMODE"`

	// Redis (prompted; RedisPassword prompted)
	RedisAddr     string `yaml:"redis_addr" env:"REDIS_ADDR"`
	RedisPassword string `yaml:"redis_password" env:"REDIS_PASSWORD"`
	RedisDB       string `yaml:"redis_db" env:"REDIS_DB"`

	// JWT
	JWTSecret        string `yaml:"jwt_secret" env:"JWT_SECRET"`
	JWTExpiryHours   string `yaml:"jwt_expiry_hours" env:"JWT_EXPIRY_HOURS"`
	JWTRefreshExpiry string `yaml:"jwt_refresh_expiry" env:"JWT_REFRESH_EXPIRY"`

	// Auto-generated (first install only)
	DBCredentialEncryptionKey string `yaml:"db_credential_encryption_key" env:"DB_CREDENTIAL_ENCRYPTION_KEY"`
	MigrateToken              string `yaml:"migrate_token" env:"MIGRATE_TOKEN"`
}

// DefaultConfigDir returns the directory for .env and config.yaml.
// Uses SESSIONDB_CONFIG_DIR if set; otherwise /opt/sessiondb when running as root, else $HOME/.config/sessiondb.
func DefaultConfigDir() string {
	if d := os.Getenv("SESSIONDB_CONFIG_DIR"); d != "" {
		return d
	}
	// Prefer /opt/sessiondb when running as root (e.g. system install)
	if uid := os.Getuid(); uid == 0 {
		return "/opt/sessiondb"
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		return filepath.Join(home, ".config", "sessiondb")
	}
	return "./sessiondb-config"
}

// EnvPath returns the path to the .env file inside the config directory.
func EnvPath(configDir string) string {
	return filepath.Join(configDir, ".env")
}

// ConfigYAMLPath returns the path to config.yaml inside the config directory.
func ConfigYAMLPath(configDir string) string {
	return filepath.Join(configDir, "config.yaml")
}
