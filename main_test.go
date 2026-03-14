package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Test_getInstallRoot_envOverride ensures SESSIONDB_INSTALL_ROOT is honored.
func Test_getInstallRoot_envOverride(t *testing.T) {
	t.Setenv("SESSIONDB_INSTALL_ROOT", "/custom/root")
	got := getInstallRoot("")
	if got != "/custom/root" {
		t.Fatalf("expected install root /custom/root, got %s", got)
	}
}

// Test_runDeploy_usesInstallRootEnv ensures runDeploy uses SESSIONDB_INSTALL_ROOT in ExecStart.
func Test_runDeploy_usesInstallRootEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("systemd unit generation not relevant on windows")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "conf")
	output := filepath.Join(tmpDir, "sessiondb.service")

	// Create a dummy .env so config.DefaultConfigDir()/EnvPath works if needed.
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configDir: %v", err)
	}
	envPath := filepath.Join(configDir, ".env")
	if err := os.WriteFile(envPath, []byte("SERVER_PORT=8080\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	t.Setenv("SESSIONDB_INSTALL_ROOT", "/opt/custom-sessiondb")

	if err := runDeploy(configDir, "baremetal", output); err != nil {
		t.Fatalf("runDeploy error: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read unit: %v", err)
	}
	if string(data) == "" {
		t.Fatal("expected non-empty unit file")
	}
	if want := "ExecStart=/opt/custom-sessiondb/current/server/sessiondb-server"; !containsLine(string(data), want) {
		t.Fatalf("expected ExecStart line %q in unit, got:\n%s", want, string(data))
	}
}

// containsLine reports whether text contains a line starting with the given prefix.
func containsLine(text, prefix string) bool {
	lines := 0
	start := 0
	for i := 0; i <= len(text); i++ {
		if i == len(text) || text[i] == '\n' {
			line := text[start:i]
			if len(line) > 0 && line[:min(len(line), len(prefix))] == prefix {
				return true
			}
			start = i + 1
			lines++
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test_runPrune_removesInstallAndConfig verifies prune removes install root and config directory.
func Test_runPrune_removesInstallAndConfig(t *testing.T) {
	tmpDir := t.TempDir()
	installRoot := filepath.Join(tmpDir, "install-root")
	configDir := filepath.Join(tmpDir, "config-dir")

	if err := os.MkdirAll(filepath.Join(installRoot, "versions", "v1.0.0"), 0o755); err != nil {
		t.Fatalf("mkdir install versions: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "versions", "v1.0.0", "marker.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write install marker: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(configDir, "logs"), 0o755); err != nil {
		t.Fatalf("mkdir config logs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, ".env"), []byte("SERVER_PORT=8080\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("server_port: \"8080\"\n"), 0o644); err != nil {
		t.Fatalf("write config yaml: %v", err)
	}

	t.Setenv("SESSIONDB_INSTALL_ROOT", installRoot)
	if err := runPrune(configDir, true); err != nil {
		t.Fatalf("runPrune error: %v", err)
	}

	if _, err := os.Stat(installRoot); !os.IsNotExist(err) {
		t.Fatalf("expected install root to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("expected config dir to be removed, stat err=%v", err)
	}
}
