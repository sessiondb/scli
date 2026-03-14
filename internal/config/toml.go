// Package config: single config.toml as source of truth for scli and backend.
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ConfigTOMLPath returns the path to config.toml inside the config directory.
func ConfigTOMLPath(configDir string) string {
	return filepath.Join(configDir, "config.toml")
}

// TomlConfig is the single-file config used by scli init and consumed by run/deploy/backend.
// Sections: server, database, redis, jwt, secrets, auth (default_logins), ui.
type TomlConfig struct {
	Server   TomlServer   `toml:"server"`
	Database TomlDatabase `toml:"database"`
	Redis    TomlRedis    `toml:"redis"`
	JWT      TomlJWT      `toml:"jwt"`
	Secrets  TomlSecrets  `toml:"secrets"`
	Auth     TomlAuth     `toml:"auth"`
	UI       TomlUI       `toml:"ui"`
}

type TomlServer struct {
	Port string `toml:"port"`
	Mode string `toml:"mode"`
}

type TomlDatabase struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
	SSLMode  string `toml:"ssl_mode"`
}

type TomlRedis struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type TomlJWT struct {
	Secret        string `toml:"secret"`
	ExpiryHours   int    `toml:"expiry_hours"`
	RefreshExpiry int    `toml:"refresh_expiry"`
}

type TomlSecrets struct {
	DBCredentialEncryptionKey string `toml:"db_credential_encryption_key"`
	MigrateToken             string `toml:"migrate_token"`
}

type TomlAuth struct {
	DefaultLogins []TomlDefaultLogin `toml:"default_logins"`
}

type TomlDefaultLogin struct {
	Email    string `toml:"email"`
	Password string `toml:"password"`
	RoleKey  string `toml:"role_key"`
}

type TomlUI struct {
	APIURL string `toml:"api_url"`
}

// LoadConfigTOML reads config.toml from path. Returns nil, nil if file does not exist.
func LoadConfigTOML(path string) (*TomlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c TomlConfig
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// WriteConfigTOML writes the TOML config to path. Creates parent dirs. Mode 0600 for secrets.
func WriteConfigTOML(path string, c *TomlConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DefaultAuthLogins returns default [[auth.default_logins]] entries (admin, guest).
func DefaultAuthLogins() []TomlDefaultLogin {
	return []TomlDefaultLogin{
		{Email: "admin@sessiondb.internal", Password: "admin123", RoleKey: "super_admin"},
		{Email: "guest@sessiondb.internal", Password: "guest123", RoleKey: "analyst"},
	}
}

const defaultAPIURL = "http://localhost:8080/v1"

// EnvConfigToToml converts EnvConfig (from prompts) to TomlConfig, with default auth and ui.
func EnvConfigToToml(cfg *EnvConfig) *TomlConfig {
	if cfg == nil {
		return nil
	}
	db := 0
	if cfg.RedisDB != "" {
		if n, err := strconv.Atoi(cfg.RedisDB); err == nil {
			db = n
		}
	}
	jwtExp, jwtRef := 24, 720
	if cfg.JWTExpiryHours != "" {
		if n, err := strconv.Atoi(cfg.JWTExpiryHours); err == nil {
			jwtExp = n
		}
	}
	if cfg.JWTRefreshExpiry != "" {
		if n, err := strconv.Atoi(cfg.JWTRefreshExpiry); err == nil {
			jwtRef = n
		}
	}
	return &TomlConfig{
		Server:   TomlServer{Port: cfg.ServerPort, Mode: cfg.ServerMode},
		Database: TomlDatabase{Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser, Password: cfg.DBPassword, Name: cfg.DBName, SSLMode: cfg.DBSSLMode},
		Redis:    TomlRedis{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: db},
		JWT:      TomlJWT{Secret: cfg.JWTSecret, ExpiryHours: jwtExp, RefreshExpiry: jwtRef},
		Secrets:  TomlSecrets{DBCredentialEncryptionKey: cfg.DBCredentialEncryptionKey, MigrateToken: cfg.MigrateToken},
		Auth:     TomlAuth{DefaultLogins: DefaultAuthLogins()},
		UI:       TomlUI{APIURL: defaultAPIURL},
	}
}

// TomlToEnvConfig converts TomlConfig to EnvConfig for .env generation and existing callers.
func TomlToEnvConfig(c *TomlConfig) *EnvConfig {
	if c == nil {
		return nil
	}
	cfg := &EnvConfig{
		ServerPort:            c.Server.Port,
		ServerMode:             c.Server.Mode,
		DBHost:                 c.Database.Host,
		DBPort:                 c.Database.Port,
		DBUser:                 c.Database.User,
		DBPassword:             c.Database.Password,
		DBName:                 c.Database.Name,
		DBSSLMode:              c.Database.SSLMode,
		RedisAddr:              c.Redis.Addr,
		RedisPassword:          c.Redis.Password,
		RedisDB:                strconv.Itoa(c.Redis.DB),
		JWTSecret:              c.JWT.Secret,
		JWTExpiryHours:         strconv.Itoa(c.JWT.ExpiryHours),
		JWTRefreshExpiry:       strconv.Itoa(c.JWT.RefreshExpiry),
		DBCredentialEncryptionKey: c.Secrets.DBCredentialEncryptionKey,
		MigrateToken:           c.Secrets.MigrateToken,
	}
	return cfg
}

// EnvSliceFromToml returns KEY=value strings for exec.Cmd.Env from TomlConfig.
// Includes backend vars and API_URL for the UI server.
func EnvSliceFromToml(base []string, c *TomlConfig) []string {
	if c == nil {
		return base
	}
	cfg := TomlToEnvConfig(c)
	m := map[string]string{
		"SERVER_PORT":                    cfg.ServerPort,
		"SERVER_MODE":                    cfg.ServerMode,
		"DB_HOST":                        cfg.DBHost,
		"DB_PORT":                        cfg.DBPort,
		"DB_USER":                        cfg.DBUser,
		"DB_PASSWORD":                    cfg.DBPassword,
		"DB_NAME":                        cfg.DBName,
		"DB_SSLMODE":                     cfg.DBSSLMode,
		"REDIS_ADDR":                     cfg.RedisAddr,
		"REDIS_PASSWORD":                 cfg.RedisPassword,
		"REDIS_DB":                       cfg.RedisDB,
		"JWT_SECRET":                     cfg.JWTSecret,
		"JWT_EXPIRY_HOURS":               cfg.JWTExpiryHours,
		"JWT_REFRESH_EXPIRY":             cfg.JWTRefreshExpiry,
		"DB_CREDENTIAL_ENCRYPTION_KEY":  cfg.DBCredentialEncryptionKey,
		"MIGRATE_TOKEN":                  cfg.MigrateToken,
		"API_URL":                        c.UI.APIURL,
	}
	seen := make(map[string]bool)
	for _, s := range base {
		if idx := strings.IndexByte(s, '='); idx > 0 {
			seen[s[:idx]] = true
		}
	}
	out := append([]string(nil), base...)
	for k, v := range m {
		if v != "" && !seen[k] {
			out = append(out, k+"="+v)
			seen[k] = true
		}
	}
	return out
}
