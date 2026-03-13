package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WriteEnv writes EnvConfig to the .env file at the given path. Creates parent dirs.
func WriteEnv(path string, cfg *EnvConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	writeEnvLine(w, "SERVER_PORT", cfg.ServerPort)
	writeEnvLine(w, "SERVER_MODE", cfg.ServerMode)
	writeEnvLine(w, "DB_HOST", cfg.DBHost)
	writeEnvLine(w, "DB_PORT", cfg.DBPort)
	writeEnvLine(w, "DB_USER", cfg.DBUser)
	writeEnvLine(w, "DB_PASSWORD", cfg.DBPassword)
	writeEnvLine(w, "DB_NAME", cfg.DBName)
	writeEnvLine(w, "DB_SSLMODE", cfg.DBSSLMode)
	writeEnvLine(w, "REDIS_ADDR", cfg.RedisAddr)
	writeEnvLine(w, "REDIS_PASSWORD", cfg.RedisPassword)
	writeEnvLine(w, "REDIS_DB", cfg.RedisDB)
	writeEnvLine(w, "JWT_SECRET", cfg.JWTSecret)
	writeEnvLine(w, "JWT_EXPIRY_HOURS", cfg.JWTExpiryHours)
	writeEnvLine(w, "JWT_REFRESH_EXPIRY", cfg.JWTRefreshExpiry)
	writeEnvLine(w, "DB_CREDENTIAL_ENCRYPTION_KEY", cfg.DBCredentialEncryptionKey)
	writeEnvLine(w, "MIGRATE_TOKEN", cfg.MigrateToken)
	return w.Flush()
}

func writeEnvLine(w *bufio.Writer, key, value string) {
	escaped := value
	escaped = strings.ReplaceAll(escaped, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	if strings.ContainsAny(escaped, " \t\n\"") {
		escaped = `"` + escaped + `"`
	}
	_, _ = w.WriteString(key + "=" + escaped + "\n")
}

// WriteConfigYAML writes EnvConfig to config.yaml (secrets included; file should be restricted).
func WriteConfigYAML(path string, cfg *EnvConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// EnvExists returns true if the .env file exists and contains a non-empty value for the given key.
func EnvExists(path, key string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	prefix := key + "="
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimPrefix(line, prefix)
			val = strings.Trim(val, `"`)
			if len(val) > 0 {
				return true, nil
			}
		}
	}
	return false, scanner.Err()
}

// ReadEnv reads key=value pairs from .env and returns a map (used by loader and to check existing secrets).
func ReadEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := line[:idx]
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"`)
		out[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// LoadEnvConfig reads .env from path and populates EnvConfig.
func LoadEnvConfig(path string) (*EnvConfig, error) {
	m, err := ReadEnv(path)
	if err != nil {
		return nil, err
	}
	cfg := &EnvConfig{}
	if v, ok := m["SERVER_PORT"]; ok {
		cfg.ServerPort = v
	}
	if v, ok := m["SERVER_MODE"]; ok {
		cfg.ServerMode = v
	}
	if v, ok := m["DB_HOST"]; ok {
		cfg.DBHost = v
	}
	if v, ok := m["DB_PORT"]; ok {
		cfg.DBPort = v
	}
	if v, ok := m["DB_USER"]; ok {
		cfg.DBUser = v
	}
	if v, ok := m["DB_PASSWORD"]; ok {
		cfg.DBPassword = v
	}
	if v, ok := m["DB_NAME"]; ok {
		cfg.DBName = v
	}
	if v, ok := m["DB_SSLMODE"]; ok {
		cfg.DBSSLMode = v
	}
	if v, ok := m["REDIS_ADDR"]; ok {
		cfg.RedisAddr = v
	}
	if v, ok := m["REDIS_PASSWORD"]; ok {
		cfg.RedisPassword = v
	}
	if v, ok := m["REDIS_DB"]; ok {
		cfg.RedisDB = v
	}
	if v, ok := m["JWT_SECRET"]; ok {
		cfg.JWTSecret = v
	}
	if v, ok := m["JWT_EXPIRY_HOURS"]; ok {
		cfg.JWTExpiryHours = v
	}
	if v, ok := m["JWT_REFRESH_EXPIRY"]; ok {
		cfg.JWTRefreshExpiry = v
	}
	if v, ok := m["DB_CREDENTIAL_ENCRYPTION_KEY"]; ok {
		cfg.DBCredentialEncryptionKey = v
	}
	if v, ok := m["MIGRATE_TOKEN"]; ok {
		cfg.MigrateToken = v
	}
	return cfg, nil
}
