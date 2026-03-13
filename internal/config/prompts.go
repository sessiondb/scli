package config

import (
	"github.com/AlecAivazis/survey/v2"
)

// Default values for prompts.
const (
	DefaultDBHost    = "localhost"
	DefaultDBUser    = "sessiondb"
	DefaultDBName    = "sessiondb"
	DefaultDBPort    = "5432"
	DefaultRedisAddr = "localhost:6379"
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
	}
	if err := survey.Ask(qs, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
