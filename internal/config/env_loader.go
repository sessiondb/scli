package config

import (
	"fmt"
	"os"
	"strings"
)

// LoadEnvIntoProcess reads the .env file at path and sets each key=value in the current process environment.
// Use this when the CLI needs to run a child that should see the same env (e.g. backend server).
func LoadEnvIntoProcess(path string) error {
	m, err := ReadEnv(path)
	if err != nil {
		return err
	}
	for k, v := range m {
		_ = os.Setenv(k, v)
	}
	return nil
}

// EnvSlice returns KEY=value strings suitable for exec.Cmd.Env, merging existing base with .env file.
// If path is empty or file missing, returns base unchanged.
func EnvSlice(base []string, path string) ([]string, error) {
	m, err := ReadEnv(path)
	if err != nil {
		if os.IsNotExist(err) {
			return base, nil
		}
		return nil, err
	}
	seen := make(map[string]bool)
	for _, s := range base {
		if idx := strings.IndexByte(s, '='); idx > 0 {
			seen[s[:idx]] = true
		}
	}
	out := append([]string(nil), base...)
	for k, v := range m {
		if !seen[k] {
			out = append(out, k+"="+v)
			seen[k] = true
		}
	}
	return out, nil
}

// EnvSliceFromInitFiles returns KEY=value strings for exec.Cmd.Env by loading the init-generated config.
// Prefers config.toml (single source of truth); if missing, falls back to .env then config.yaml.
// Whenever config.toml is loaded, .env is regenerated from it so they stay in sync.
func EnvSliceFromInitFiles(base []string, configDir string) ([]string, error) {
	tomlPath := ConfigTOMLPath(configDir)
	tomlCfg, err := LoadConfigTOML(tomlPath)
	if err != nil {
		return nil, err
	}
	if tomlCfg != nil {
		envPath := EnvPath(configDir)
		_ = WriteEnv(envPath, TomlToEnvConfig(tomlCfg)) // keep .env in sync whenever toml is used
		return EnvSliceFromToml(base, tomlCfg), nil
	}

	envPath := EnvPath(configDir)
	out, err := EnvSlice(base, envPath)
	if err == nil {
		return out, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	cfgPath := ConfigYAMLPath(configDir)
	cfg, cfgErr := LoadConfigYAML(cfgPath)
	if cfgErr != nil {
		if os.IsNotExist(cfgErr) {
			return nil, fmt.Errorf("no config found (looked for %s, %s, %s); run 'scli init' first", tomlPath, envPath, cfgPath)
		}
		return nil, cfgErr
	}

	if writeErr := WriteEnv(envPath, cfg); writeErr != nil {
		return nil, writeErr
	}
	return EnvSlice(base, envPath)
}
