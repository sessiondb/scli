// Package main provides the SessionDB CLI (get, run) for versioned binders.
package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SessionDBConfig is the runtime config (sessiondb.yaml) for ports, paths, and env vars (CLI injects env into server and UI).
type SessionDBConfig struct {
	Server struct {
		Port   string `yaml:"port"`
		Binary string `yaml:"binary"`
	} `yaml:"server"`
	UI struct {
		Port int    `yaml:"port"`
		Dir  string `yaml:"dir"`
	} `yaml:"ui"`
	Env map[string]interface{} `yaml:"env"` // values can be string or number; CLI converts to string for env
}

// loadConfig reads and parses a YAML config file. Returns nil, nil if path is empty or file missing.
func loadConfig(path string) (*SessionDBConfig, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c SessionDBConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// envSlice returns KEY=value strings for exec.Cmd.Env (base is typically os.Environ(); env from config is merged).
func envSlice(base []string, env map[string]interface{}) []string {
	if len(env) == 0 {
		return base
	}
	seen := make(map[string]bool)
	for _, s := range base {
		if idx := strings.IndexByte(s, '='); idx > 0 {
			seen[s[:idx]] = true
		}
	}
	out := make([]string, 0, len(base)+len(env))
	out = append(out, base...)
	for k, v := range env {
		if !seen[k] {
			out = append(out, k+"="+envValString(v))
			seen[k] = true
		}
	}
	return out
}

func envValString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
