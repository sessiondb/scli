package config

import (
	"github.com/AlecAivazis/survey/v2"
)

// Default values for prompts (match sessiondb .env).
const (
	DefaultDBHost        = "localhost"
	DefaultDBUser        = "sessiondb"
	DefaultDBName        = "sessiondb"
	DefaultDBPort        = "5432"
	DefaultDBSSLMode     = "disable"
	DefaultRedisAddr     = "localhost:6379"
	DefaultRedisDB       = "0"
	DefaultServerPort    = "8080"
	DefaultServerMode    = "debug"
	DefaultJWTSecret     = "your_secret_key_change_in_production"
	DefaultJWTExpiry     = "24"
	DefaultJWTRefreshExp = "720"
)

// RunPrompts asks the user for DB and Redis configuration; returns an EnvConfig with user input only (secrets left empty).
func RunPrompts() (*EnvConfig, error) {
	cfg := &EnvConfig{}
	qs := []*survey.Question{
		{
			Name:   "DBHost",
			Prompt: &survey.Input{Message: "DB Host:", Default: DefaultDBHost},
			Transform: survey.TransformString(func(s string) string {
				if s == "" {
					return DefaultDBHost
				}
				return s
			}),
		},
		{
			Name:   "DBUser",
			Prompt: &survey.Input{Message: "DB User:", Default: DefaultDBUser},
			Transform: survey.TransformString(func(s string) string {
				if s == "" {
					return DefaultDBUser
				}
				return s
			}),
		},
		{
			Name:   "DBPassword",
			Prompt: &survey.Password{Message: "DB Password:"},
		},
		{
			Name:   "DBName",
			Prompt: &survey.Input{Message: "DB Name:", Default: DefaultDBName},
			Transform: survey.TransformString(func(s string) string {
				if s == "" {
					return DefaultDBName
				}
				return s
			}),
		},
		{
			Name:   "DBPort",
			Prompt: &survey.Input{Message: "DB Port:", Default: DefaultDBPort},
			Transform: survey.TransformString(func(s string) string {
				if s == "" {
					return DefaultDBPort
				}
				return s
			}),
		},
		{
			Name:   "RedisAddr",
			Prompt: &survey.Input{Message: "Redis Address:", Default: DefaultRedisAddr},
			Transform: survey.TransformString(func(s string) string {
				if s == "" {
					return DefaultRedisAddr
				}
				return s
			}),
		},
		{
			Name:   "RedisPassword",
			Prompt: &survey.Password{Message: "Redis Password: (leave empty if none)"},
		},
	}
	if err := survey.Ask(qs, cfg); err != nil {
		return nil, err
	}
	// Set defaults for non-prompted fields (match sessiondb .env)
	if cfg.ServerPort == "" {
		cfg.ServerPort = DefaultServerPort
	}
	if cfg.ServerMode == "" {
		cfg.ServerMode = DefaultServerMode
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = DefaultDBSSLMode
	}
	if cfg.RedisDB == "" {
		cfg.RedisDB = DefaultRedisDB
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = DefaultJWTSecret
	}
	if cfg.JWTExpiryHours == "" {
		cfg.JWTExpiryHours = DefaultJWTExpiry
	}
	if cfg.JWTRefreshExpiry == "" {
		cfg.JWTRefreshExpiry = DefaultJWTRefreshExp
	}
	return cfg, nil
}
