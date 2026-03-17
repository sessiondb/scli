# SessionDB CLI (scli) — Tooling Guide

This document is the main tooling reference for the SessionDB CLI (`scli`). It includes a **step-by-step workflow** and a **troubleshooting section** with common issues, how to verify them, and how to resolve or find the solution.

---

## Table of contents

1. [Step-by-step guide: Working with scli](#step-by-step-guide-working-with-scli)
2. [Issues and solutions](#issues-and-solutions)
3. [How to find the solution for a new issue](#how-to-find-the-solution-for-a-new-issue)

---

## Step-by-step guide: Working with scli

### Prerequisites

- **Network**: GitHub access for install/update (releases from `sessiondb/service`, `sessiondb/client`, and `sessiondb/scli`).
- **Permissions**: Write access to install root (default: `/opt/sessiondb` as root, or `~/.local/share/sessiondb` as user) and config directory (default: `~/.config/sessiondb` or `SESSIONDB_CONFIG_DIR`).
- **Optional**: On Linux with systemd, `systemctl` for deploy/start/stop of services.

### Step 1: Install scli

Choose one:

- **One-liner (recommended)**  
  ```bash
  curl -sSL https://raw.githubusercontent.com/sessiondb/scli/main/install.sh | bash
  ```
  For a specific version:  
  `curl -sSL ... | bash -s -- v1.0.0`

- **Go install**  
  ```bash
  go install .
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```

- **Manual**: Download the binary from [Releases](https://github.com/sessiondb/scli/releases), put it on your `PATH`, and name it `scli` (or `scli.exe` on Windows).

Then open a new terminal or run `source ~/.zshrc` (or `~/.bashrc`) so `scli` is on `PATH`. Verify:

```bash
scli
# Should print usage (init, install, run, ...).
```

### Step 2: Initialize configuration

Run the interactive wizard to set database, Redis, and server options. This creates **config.toml** and generates **.env** for the backend/systemd.

```bash
scli init
```

- Use `--config-dir DIR` if you want a custom config directory (otherwise `$HOME/.config/sessiondb` or `SESSIONDB_CONFIG_DIR`).
- Secrets (`DB_CREDENTIAL_ENCRYPTION_KEY`, `MIGRATE_TOKEN`) are generated on first run and reused if config already exists.

Verify:

```bash
scli config view
# Shows config directory and main settings (and whether secrets are set).
```

### Step 3: Install SessionDB (backend + frontend)

Download and install a SessionDB release into the install root (`versions/<tag>/`, `current` symlink).

```bash
scli install
# Or a specific version:
scli install v1.0.1
```

- Omit version for **latest**.
- Use `-v` or `--verbose` for detailed download/log output.
- Use `--force` to reinstall even if the version is already present (and to fetch UI binary if it was missing).

If the systemd units are already installed and active, `scli install` will try to restart them with the new version (on Linux, and when run as root).

Verify:

```bash
scli resources
# Shows install root, current version, paths to server binary, UI, config files, and systemd unit.
```

### Step 4a: Run locally (no systemd)

Start the server in the background and optionally follow logs:

```bash
scli run
# Or: scli run v1.0.1
# Or: scli run v1.0.1 /path/to/workdir
```

- Config is read from **config.toml** (or .env) in the config directory.
- Ctrl+C only stops the log tail; the server keeps running. Stop it with:

```bash
scli stop
```

To start without following logs:

```bash
scli start
# Then: scli logs -f --component api  (or ui)
```

To run only API or only UI:

```bash
scli run --component api
scli run --component ui
scli run --component all
```

### Step 4b: Deploy with systemd (bare metal)

Generate systemd unit file(s) and install them:

```bash
scli deploy --platform baremetal --output sessiondb.service
```

- Generates **sessiondb.service** (API) and **sessiondb-ui.service** (UI) by default (`--component all`). Use `--component api` or `--component ui` to generate only one.
- If `.env` is missing, it is generated from **config.toml** (or legacy config.yaml).

Then:

```bash
sudo cp sessiondb.service sessiondb-ui.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable sessiondb sessiondb-ui
sudo systemctl start sessiondb sessiondb-ui
```

Verify:

```bash
scli status
# Uses default URL http://localhost:8080; use --url if different.
```

### Step 5: Run migrations

After the server is running (either via `scli run`/`scli start` or systemd):

```bash
scli migrate
```

- Uses `MIGRATE_TOKEN` from config (config.toml / .env).
- Default URL is `http://localhost:8080`; override with `--url BASE_URL`.

### Step 6: Check status and logs

- **Health**:  
  `scli status` (default: `http://localhost:8080/health`).

- **Logs**:  
  - With systemd: `scli logs -n 100 -f --component api` (or `ui`) uses `journalctl -u sessiondb` / `sessiondb-ui`.
  - Without systemd: same command tails the run log file under the config directory.

### Step 7: Restart or upgrade

- **Restart (no new download)**  
  `scli restart` (or `scli restart --component all`). Uses same version and config.

- **Upgrade to a new version**  
  `scli install v1.0.2` (or `scli install` for latest). If systemd is in use and you are root, services are restarted automatically; otherwise you may need `sudo systemctl restart sessiondb sessiondb-ui`.

- **Self-update scli**  
  `scli update` — choose one of the last 5 scli releases and replace the current binary. On Windows, the new binary is written as `scli.exe.new`; replace `scli.exe` manually after closing the terminal.

### Step 8: Reset or full cleanup

- **Remove installed versions only** (keeps config):  
  `scli reset`  
  Removes install root’s `versions/` and `current` (and legacy `sessiondb-install` under config dir).

- **Remove versions and legacy config files** (.env, config.yaml):  
  `scli reset --all`  
  Note: This does **not** remove **config.toml** in the current implementation; see [Issues and solutions](#reset---all-does-not-remove-configtoml) below.

- **Remove everything** (install root + entire config directory, including config.toml, .env, logs, pid):  
  `scli prune --yes`  
  Destructive; use only when you want a full cleanup.

---

## Issues and solutions

Each entry describes the **issue**, **how to check** whether you’re affected, and **what to do** (solution or where to look for the fix).

---

### Config / init

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"config not found" or "run scli init first"** | Run `scli config view` or try `scli migrate` / `scli deploy`. | Run `scli init` (and optionally `--config-dir DIR`). Ensure `SESSIONDB_CONFIG_DIR` points to the same directory you use for init. |
| **Wrong config directory used** | Compare paths printed by `scli config view` with where you edited config. | Set `SESSIONDB_CONFIG_DIR` before running scli, or pass `--config-dir DIR` on every command that supports it (init, install, migrate, deploy, reset, prune, config, logs, stop, start, run, restart). |
| **Secrets (MIGRATE_TOKEN / DB_CREDENTIAL_ENCRYPTION_KEY) missing after re-init** | Run `scli config view` and check whether secrets are shown as “(set)”. | Init **reuses** existing secrets from config.toml or .env. If you removed those files, new secrets are generated. Back up config.toml/.env before removing them. |
| **config view fails or shows wrong format** | Run `scli config view`. | Prefer **config.toml** as single source of truth. If only .env or config.yaml exists, view/edit use that; ensure file is valid and in the expected format (see `internal/config` in the repo). |

---

### Install / get

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"release not found" or "asset not found"** | Run `scli install VERSION` or `scli get VERSION` with `-v`. | Confirm the tag exists for both **sessiondb/service** and **sessiondb/client** on GitHub. Use a tag that has backend and frontend assets for your OS/arch (e.g. `sessiondb-backend-linux-amd64.tar.gz`, frontend tarball or `sessiondb-ui-<os>-<arch>`). |
| **"UI binary not found" / "install a release that includes sessiondb-ui"** | Run `scli resources` and check “ui binary” path; or `scli run --component ui`. | Install a release that publishes `sessiondb-ui-<os>-<arch>` (or the tarball that extracts the UI server). Use `scli install VERSION --force` to re-download and add the UI binary to an existing version dir. |
| **Checksum mismatch** | Install fails with “checksum mismatch” in verbose output. | Re-download: use `--force` and run install again. If it persists, check release’s `checksums.txt` / `checksums-frontend.txt` on GitHub and open an issue in the repo that publishes the asset. |
| **Network / GitHub rate limit** | Install fails with connection or API errors. | Check network and proxy; for rate limits, wait or use a GitHub token (if the CLI supports it in the future). See `release.go` / `get.go` for where requests are made. |

---

### Run / start / stop / restart

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"no setup.sh at ... run scli install first"** | Run `scli run` or `scli start` without an existing install. | Run `scli install [version]` so that `current` points to a version that has `setup.sh` (and server binary) under `versions/<tag>/server/`. |
| **"UI binary not found at ..."** | Run `scli run --component ui` or `scli run --component all`. | Same as install: use a release that includes the UI binary and optionally `scli install VERSION --force`. |
| **"API already running (PID ...)"** | Start again without stopping. | Run `scli stop`, then `scli start` or `scli run`. If you don’t want to stop, use `scli logs -f --component api` to follow logs. |
| **scli stop doesn’t stop systemd service** | Run `scli stop` then `systemctl status sessiondb`. | On Linux, `scli stop` stops both PID-file processes and systemd units `sessiondb` and `sessiondb-ui`. If a unit is still active, run `sudo systemctl stop sessiondb sessiondb-ui`. |
| **Restart doesn’t pick up new version** | After `scli install vNEW`, run `scli restart`. | If using systemd, ensure units are installed and you’ve restarted them (`sudo systemctl restart sessiondb sessiondb-ui`). If not root, scli may only print a message to restart manually. See `lifecycle.go` (`tryRestartSystemdAfterInstall`). |

---

### Deploy

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"Warning: no config found. Run scli init first"** | Run `scli deploy` without config.toml or .env in config dir. | Run `scli init` so that config.toml (and generated .env) exist. Deploy generates .env from config.toml if .env is missing. |
| **"Warning: API binary not found at ..."** | Run `scli deploy` before install. | Run `scli install [version]` first so `current/server/sessiondb-server` exists. |
| **"Warning: UI binary not found at ..."** | Run `scli deploy --component ui` or `all` without UI binary. | Install a release that includes the UI binary; see Install section above. |
| **Unsupported platform** | Pass e.g. `--platform docker`. | Only `baremetal` is supported. Use `--platform baremetal` (default). |

---

### Migrate / status

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"config not found at ... (run scli init first)"** | Run `scli migrate`. | Run `scli init` and ensure config dir contains .env or config.toml; migrate reads from .env. |
| **"MIGRATE_TOKEN not set"** | Run `scli config view` and check secrets. | Run `scli init` to generate (or restore) MIGRATE_TOKEN, or set it in config.toml / .env and regenerate .env if needed. |
| **"migrate request failed" / connection error** | Run `scli migrate` with server down or wrong URL. | Start the server (`scli run` or systemd) and ensure it listens on the URL you use. Override with `scli migrate --url http://HOST:PORT`. |
| **"migrate failed: HTTP 401/403"** | Migrate returns non-200. | Token mismatch: ensure `X-Migrate-Token` used by the server matches `MIGRATE_TOKEN` in config. Check server logs and config (e.g. `scli config view`). |
| **"Status: unreachable"** | Run `scli status`. | Server not running or not on default `http://localhost:8080`. Start server and/or use `scli status --url http://HOST:PORT`. |

---

### Reset / prune

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **reset --all does not remove config.toml** | Run `scli reset --all` and list config dir; config.toml may still be there. | Current behavior: `reset --all` removes .env and **config.yaml** (legacy), not **config.toml**. For a full config reset either remove config.toml manually or use `scli prune --yes` (removes entire config directory). See `cmd_reset.go` for exact paths removed. |
| **"prune is destructive; re-run with --yes"** | Run `scli prune` without `--yes`. | Intended safety check. Run `scli prune --yes` only when you want to remove the entire install root and config directory. |
| **Permission denied removing install root or config dir** | Prune or reset fails with permission errors. | Run with sufficient permissions (e.g. sudo for `/opt/sessiondb` or adjust ownership). Check `SESSIONDB_INSTALL_ROOT` and `SESSIONDB_CONFIG_DIR`. |

---

### Logs / resources

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"log file not found at ... (start with scli run --component ... first)"** | Run `scli logs` or `scli logs -f --component api` without systemd and without having started via scli. | Start the component with `scli run --component api` (or ui) so the log file is created, or install systemd units and use `journalctl -u sessiondb`. |
| **resources shows "(missing)" for binary or config** | Run `scli resources`. | Install the correct version (`scli install`) and run `scli init` if config is missing. Fix paths/permissions if install root or config dir were moved. |

---

### Update (self-update)

| Issue | How to check | Solution / where to look |
|-------|----------------|---------------------------|
| **"no releases found" / "no asset ... for release"** | Run `scli update`. | Check network and GitHub access to **sessiondb/scli** releases. Ensure the release has an asset for your OS/arch (e.g. `scli-v0.1.0-linux-amd64`). |
| **Windows: binary not replaced** | Update says “Replace ... with it and restart”. | On Windows the new binary is written as `scli.exe.new`. Close all scli terminals, then replace `scli.exe` with the `.new` file manually. |

---

## How to find the solution for a new issue

1. **Reproduce and capture the exact error**  
   Run the failing command (e.g. `scli install v1.0.0`, `scli migrate`) and note the full error message and exit code.

2. **Check config and paths**  
   - `scli config view` — confirms config directory and that secrets/settings are present.  
   - `scli resources` — confirms install root, current version, and presence of server binary, UI, config files, and systemd unit.

3. **Use verbose where available**  
   For install/get: `scli install VERSION -v` (or `--verbose`) to see download and extraction steps.

4. **Verify environment**  
   - `echo $SESSIONDB_CONFIG_DIR` and `echo $SESSIONDB_INSTALL_ROOT` — must match where you ran `scli init` and where you expect installs.  
   - On Linux with systemd: `systemctl status sessiondb sessiondb-ui` and `journalctl -u sessiondb -n 50`.

5. **Match code to behavior**  
   - **Config / paths**: `internal/config/env.go`, `internal/config/toml.go`, `internal/config/env_loader.go`.  
   - **Install / get**: `get.go`, `release.go`.  
   - **Run / start / stop**: `run.go`, `cmd_stop.go`, `lifecycle.go`.  
   - **Deploy**: `cmd_deploy.go`.  
   - **Migrate / status**: `cmd_migrate.go`, `cmd_status.go`.  
   - **Reset / prune**: `cmd_reset.go`, `cmd_prune.go`.

6. **Documentation and issues**  
   - This guide: `docs/tooling.md`.  
   - Main usage: `README.md`.  
   - For bugs or feature requests: open an issue in the **sessiondb/scli** repository (or the repo that owns the failing component) with the exact error, `scli config view` / `scli resources` output (redact secrets), and steps to reproduce.

---

## Quick reference: commands and flags

| Command | Typical use |
|--------|-------------|
| `scli init [--config-dir DIR]` | First-time config (config.toml + .env). |
| `scli install [version] [-v] [--force] [--workdir DIR] [--config-dir DIR]` | Install SessionDB release; optional reinstall/UI with `--force`. |
| `scli get <version> [workdir] [-v]` | Same as install; workdir = install root. |
| `scli run [version] [workdir] [--config-dir DIR] [--component api\|ui\|all]` | Start and follow logs. |
| `scli start [version] [workdir] [--config-dir DIR] [--component api\|ui\|all]` | Start in background. |
| `scli stop [--config-dir DIR]` | Stop PID-file and systemd processes. |
| `scli restart [version] [workdir] [--config-dir DIR] [--component api\|ui\|all]` | Stop then start. |
| `scli migrate [--config-dir DIR] [--url BASE_URL]` | POST /v1/migrate with token from config. |
| `scli status [--url BASE_URL]` | GET /health. |
| `scli deploy [--config-dir DIR] [--platform baremetal] [--output FILE] [--component api\|ui\|all]` | Generate systemd unit(s). |
| `scli reset [--config-dir DIR] [--all]` | Remove install dir; `--all` also removes .env and config.yaml. |
| `scli prune [--config-dir DIR] --yes` | Remove install root and config dir. |
| `scli update` | Self-update scli binary. |
| `scli resources [--config-dir DIR] [--install-root DIR]` | Show paths and presence of binaries/config. |
| `scli logs [-n N] [-f] [--component api\|ui]` | Show/follow logs (journalctl or log file). |
| `scli config view\|edit [--config-dir DIR]` | View or edit config.toml / .env. |

Default URLs: migrate and status use `http://localhost:8080` unless overridden with `--url`.
