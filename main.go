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
		verboseFlag := fs.Bool("verbose", false, "Print detailed logs")
		verboseShort := fs.Bool("v", false, "Print detailed logs (short)")
		_ = fs.Parse(args)
		verbose = *verboseFlag || *verboseShort
		version := ""
		if fs.NArg() > 0 {
			version = fs.Arg(0)
		}
		err = runInstall(version, *workDir, *configDir)
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
		err = get(version, workDir)
	case "run":
		version := ""
		workDir := "."
		if len(args) >= 1 {
			version = args[0]
		}
		if len(args) >= 2 {
			workDir = args[1]
		}
		if version == "" {
			fmt.Fprintln(os.Stderr, "Usage: scli run <version> [workdir]")
			os.Exit(1)
		}
		workDir, _ = filepath.Abs(workDir)
		err = run(version, workDir)
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
		_ = fs.Parse(args)
		err = runDeploy(*configDir, *platform, *output)
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
  init              Interactive configuration + generate secrets, save .env and config.yaml
  install <version> Download and extract SessionDB binaries (e.g. scli install v1.0.1)
  get <version>     Same as install, extract to workdir/sessiondb (default workdir: .)
  run <version>     Run server+UI (get if needed, inject env from sessiondb.yaml)
  migrate           Run migrations (uses MIGRATE_TOKEN from config)
  status            Check if server is reachable
  deploy            Generate systemd unit for bare metal (--platform baremetal)

Examples:
  scli init
  scli install v1.0.1
  scli migrate
  scli status`)
}
