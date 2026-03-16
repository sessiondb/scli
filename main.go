// SessionDB CLI: init, install, run, migrate, status, deploy.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// verbose is set by -v/--verbose on install and get; enables detailed logs.
var verbose bool

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory (default: $HOME/.config/sessiondb or /opt/sessiondb)")
		_ = fs.Parse(args)
		err = runInit(*configDir)
	case "install":
		fs := flag.NewFlagSet("install", flag.ExitOnError)
		workDir := fs.String("workdir", "", "Extract to this directory (default: <config-dir>/sessiondb-install)")
		configDir := fs.String("config-dir", "", "Config directory for default workdir")
		force := fs.Bool("force", false, "Reinstall even if version already installed (also downloads UI binary if missing)")
		verboseFlag := fs.Bool("verbose", false, "Print detailed logs")
		verboseShort := fs.Bool("v", false, "Print detailed logs (short)")
		_ = fs.Parse(args)
		verbose = *verboseFlag || *verboseShort
		version := ""
		if fs.NArg() > 0 {
			version = fs.Arg(0)
		}
		err = runInstall(version, *workDir, *configDir, *force)
	case "get":
		fs := flag.NewFlagSet("get", flag.ExitOnError)
		verboseFlag := fs.Bool("verbose", false, "Print detailed logs")
		verboseShort := fs.Bool("v", false, "Print detailed logs (short)")
		_ = fs.Parse(args)
		verbose = *verboseFlag || *verboseShort
		version := ""
		workDir := "."
		if fs.NArg() >= 1 {
			version = fs.Arg(0)
		}
		if fs.NArg() >= 2 {
			workDir = fs.Arg(1)
		}
		if version == "" {
			fmt.Fprintln(os.Stderr, "Usage: scli get <version> [workdir]")
			os.Exit(1)
		}
		workDir, _ = filepath.Abs(workDir)
		err = get(version, workDir, false)
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory (default: $HOME/.config/sessiondb)")
		component := fs.String("component", "api", "Component to run: api, ui, or all")
		_ = fs.Parse(args)
		version := ""
		workDir := ""
		if fs.NArg() >= 1 {
			version = fs.Arg(0)
		}
		if fs.NArg() >= 2 {
			workDir = fs.Arg(1)
		}
		if workDir == "" {
			workDir = getInstallRoot("")
		}
		workDir, _ = filepath.Abs(workDir)
		err = run(version, workDir, *configDir, *component)
	case "start":
		fs := flag.NewFlagSet("start", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory (default: $HOME/.config/sessiondb)")
		component := fs.String("component", "api", "Component to start: api, ui, or all")
		_ = fs.Parse(args)
		version := ""
		workDir := ""
		if fs.NArg() >= 1 {
			version = fs.Arg(0)
		}
		if fs.NArg() >= 2 {
			workDir = fs.Arg(1)
		}
		if workDir == "" {
			workDir = getInstallRoot("")
		}
		workDir, _ = filepath.Abs(workDir)
		err = runStartCommand(version, workDir, *configDir, *component)
	case "stop":
		fs := flag.NewFlagSet("stop", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		_ = fs.Parse(args)
		err = runStop(*configDir)
	case "restart":
		fs := flag.NewFlagSet("restart", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory (default: $HOME/.config/sessiondb)")
		component := fs.String("component", "api", "Component to restart: api, ui, or all")
		_ = fs.Parse(args)
		version := ""
		workDir := ""
		if fs.NArg() >= 1 {
			version = fs.Arg(0)
		}
		if fs.NArg() >= 2 {
			workDir = fs.Arg(1)
		}
		if workDir == "" {
			workDir = getInstallRoot("")
		}
		workDir, _ = filepath.Abs(workDir)
		err = runRestart(version, workDir, *configDir, *component)
	case "migrate":
		fs := flag.NewFlagSet("migrate", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		baseURL := fs.String("url", "http://localhost:8080", "Server base URL")
		_ = fs.Parse(args)
		err = runMigrate(*configDir, *baseURL)
	case "status":
		fs := flag.NewFlagSet("status", flag.ExitOnError)
		baseURL := fs.String("url", "http://localhost:8080", "Server base URL")
		_ = fs.Parse(args)
		err = runStatus(*baseURL)
	case "deploy":
		fs := flag.NewFlagSet("deploy", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		platform := fs.String("platform", "baremetal", "Deploy platform (baremetal)")
		output := fs.String("output", "sessiondb.service", "Output path for systemd unit")
		component := fs.String("component", "all", "Component to deploy: api, ui, or all")
		_ = fs.Parse(args)
		err = runDeploy(*configDir, *platform, *output, *component)
	case "reset":
		fs := flag.NewFlagSet("reset", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		all := fs.Bool("all", false, "Also remove config.toml and generated .env")
		_ = fs.Parse(args)
		err = runReset(*configDir, *all)
	case "prune":
		fs := flag.NewFlagSet("prune", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		yes := fs.Bool("yes", false, "Confirm destructive cleanup")
		_ = fs.Parse(args)
		err = runPrune(*configDir, *yes)
	case "update":
		err = runUpdate()
	case "resources":
		fs := flag.NewFlagSet("resources", flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		installRoot := fs.String("install-root", "", "Install root (default: SESSIONDB_INSTALL_ROOT or /opt/sessiondb when root)")
		_ = fs.Parse(args)
		err = runResources(*configDir, *installRoot)
	case "logs":
		fs := flag.NewFlagSet("logs", flag.ExitOnError)
		lines := fs.Int("n", 100, "Number of log lines to show")
		follow := fs.Bool("f", false, "Follow logs (like tail -f)")
		component := fs.String("component", "api", "Component logs: api or ui")
		_ = fs.Parse(args)
		err = runLogsWithComponent(*lines, *follow, *component)
	case "config":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Usage: scli config <view|edit> [--config-dir DIR]")
			os.Exit(1)
		}
		sub := args[0]
		subArgs := args[1:]
		fs := flag.NewFlagSet("config "+sub, flag.ExitOnError)
		configDir := fs.String("config-dir", "", "Config directory")
		_ = fs.Parse(subArgs)
		switch sub {
		case "view":
			err = runConfigView(*configDir)
		case "edit":
			err = runConfigEdit(*configDir)
		default:
			fmt.Fprintf(os.Stderr, "Unknown config subcommand: %s\n", sub)
			fmt.Fprintln(os.Stderr, "Usage: scli config <view|edit> [--config-dir DIR]")
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: scli <command> [options]

Commands:
  init                Interactive configuration + generate secrets, save config.toml (and .env for systemd)
  install [version]   Download from GitHub Releases (default: latest). Restarts systemd service if active.
  get <version>       Same as install, extract to workdir (default workdir: install root)
  run [version]       Start server in background and follow logs (uses init config). Ctrl+C stops tail only.
  start [version]     Start latest/current version in background (or systemd if installed)
  stop                Stop all SessionDB processes (PID file, systemd, leftovers)
  restart [version]   Stop then start (same options as start; use --component api|ui|all)
  migrate             Run migrations (uses MIGRATE_TOKEN from config)
  status              Check if server is reachable
  deploy              Generate systemd unit for bare metal (config from config.toml / .env, version via current symlink)
  reset               Remove SessionDB install dir (use --all to also remove config.toml and .env)
  prune               Remove all SessionDB install + config data from host (requires --yes)
  update              Fetch last 5 scli versions, pick one to install (self-update)
  resources           Show installed SessionDB resources (binaries, UI, config, unit)
  logs                Show SessionDB service logs (systemd/journalctl wrapper)
  config view|edit    View or edit SessionDB configuration (config.toml or .env)

Examples:
  scli init
  scli install v1.0.1
  scli run
  scli start
  scli restart --component all
  scli run v1.0.1
  scli stop
  scli deploy --platform baremetal
  scli reset
  scli reset --all
  scli prune --yes
  scli update
  scli migrate
  scli status`)
}
