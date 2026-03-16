package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sessiondb/scli/internal/config"
)

// runDeploy generates systemd unit file(s) for bare metal. Component: api (backend), ui (frontend server), or all (both).
func runDeploy(configDir string, platform string, outputPath string, component string) error {
	if platform != "baremetal" && platform != "" {
		return fmt.Errorf("unsupported platform: %s (only 'baremetal' is supported)", platform)
	}
	if component != "" && !validComponent(component) {
		return fmt.Errorf("invalid component %q; use api, ui, or all", component)
	}
	if component == "" {
		component = ComponentAll
	}
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}
	configDir, _ = filepath.Abs(configDir)
	tomlPath := config.ConfigTOMLPath(configDir)
	envPath := config.EnvPath(configDir)
	configYAMLPath := config.ConfigYAMLPath(configDir)
	installRoot := getInstallRoot("")
	installRoot, _ = filepath.Abs(installRoot)
	workDir := filepath.Join(installRoot, "current")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if tomlCfg, loadErr := config.LoadConfigTOML(tomlPath); loadErr == nil && tomlCfg != nil {
			if writeErr := config.WriteEnv(envPath, config.TomlToEnvConfig(tomlCfg)); writeErr != nil {
				return fmt.Errorf("generate .env from %s: %w", tomlPath, writeErr)
			}
			fmt.Printf("Generated %s from %s\n", envPath, tomlPath)
		} else if cfg, loadErr := config.LoadConfigYAML(configYAMLPath); loadErr == nil {
			if writeErr := config.WriteEnv(envPath, cfg); writeErr != nil {
				return fmt.Errorf("recreate .env from %s: %w", configYAMLPath, writeErr)
			}
			fmt.Printf("Recreated %s from %s\n", envPath, configYAMLPath)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: no config found. Run 'scli init' first.\n")
		}
	}

	writeUnit := func(outPath, desc, execPath string) error {
		// SESSIONDB_CONFIG_DIR so backend and UI read config.toml directly; .env still generated for override.
		envLine := "SESSIONDB_CONFIG_DIR=" + configDir
		unit := `[Unit]
Description=` + desc + `
After=network.target

[Service]
Type=simple
WorkingDirectory=` + workDir + `
Environment=` + envLine + `
EnvironmentFile=` + envPath + `
ExecStart=` + execPath + `
Restart=on-failure
RestartSec=5
TimeoutStartSec=30

[Install]
WantedBy=multi-user.target
`
		return os.WriteFile(outPath, []byte(unit), 0644)
	}

	var generated []string
	if component == ComponentAPI || component == ComponentAll {
		apiBin := filepath.Join(installRoot, "current", "server", "sessiondb-server")
		if _, err := os.Stat(apiBin); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: API binary not found at %s. Run 'scli install' first.\n", apiBin)
		}
		outPath := outputPath
		if outPath == "" {
			outPath = systemdAPIServiceName
		}
		if err := writeUnit(outPath, "SessionDB API Server", apiBin); err != nil {
			return err
		}
		generated = append(generated, outPath)
	}
	if component == ComponentUI || component == ComponentAll {
		uiBin := filepath.Join(installRoot, "current", "ui", uiBinaryName)
		if _, err := os.Stat(uiBin); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: UI binary not found at %s. Install a release that includes sessiondb-ui-<os>-<arch>.\n", uiBin)
		}
		outPath := systemdUIServiceName
		if component == ComponentUI && outputPath != "" {
			outPath = outputPath
		}
		if err := writeUnit(outPath, "SessionDB UI Server", uiBin); err != nil {
			return err
		}
		generated = append(generated, outPath)
	}

	fmt.Printf("Generated %v\n", generated)
	fmt.Println()
	fmt.Println("To install and start:")
	fmt.Println("  sudo cp " + systemdAPIServiceName + " " + systemdUIServiceName + " /etc/systemd/system/")
	fmt.Println("  sudo systemctl daemon-reload")
	fmt.Println("  sudo systemctl enable sessiondb sessiondb-ui")
	fmt.Println("  sudo systemctl start sessiondb sessiondb-ui")
	fmt.Println()
	fmt.Println("After upgrading: sudo systemctl restart sessiondb sessiondb-ui")
	return nil
}
