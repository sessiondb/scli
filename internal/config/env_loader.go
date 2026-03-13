package config

import (
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
