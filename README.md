# SessionDB CLI (scli)

CLI for installing and running SessionDB: interactive configuration, secret generation, install, migrate, and deploy.

## Install (one command — `scli` on PATH)

Run this to download the latest release and install so `scli` works everywhere (installs to `/usr/local/bin` if writable, else `~/.local/bin` and adds it to your shell config):

```bash
curl -sSL https://raw.githubusercontent.com/sessiondb/scli/main/install.sh | bash
```

To install a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/sessiondb/scli/main/install.sh | bash -s -- v1.0.0
```

After install, open a new terminal or run `source ~/.zshrc` (or `~/.bashrc`) so `scli` is on PATH. Then run `scli init`, `scli install v1.0.1`, etc.

---

**Other install options**

**Go install** (from this repo):
```bash
go install .
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Manual**: Download the right binary from [Releases](https://github.com/sessiondb/scli/releases), put it in a directory on your PATH, and name it `scli` (or `scli.exe` on Windows).

## Build (local only)

```bash
go build -o scli .
# or: go build -o sessiondb .
```

## Commands

| Command | Description |
|--------|-------------|
| **scli init** | Interactive prompts for DB/Redis, generates `DB_CREDENTIAL_ENCRYPTION_KEY` and `MIGRATE_TOKEN`, saves `.env` and `config.yaml` |
| **scli install [version]** | Download from GitHub Releases (sessiondb/service). Omit version for latest. Installs to install root (`versions/<tag>/`, `current` symlink). Use `-v` or `--verbose` for detailed logs. |
| **scli get \<version\> [workdir]** | Same as install; use `workdir` as install root (default: current dir). |
| **scli run [version] [workdir]** | Run server (uses `workdir/versions/<version>/` or `workdir/current`; injects env from sessiondb.yaml). |
| **scli migrate** | POST `/v1/migrate` with `X-Migrate-Token` from config (run after deploy) |
| **scli status** | Check if server is reachable (GET /health) |
| **scli deploy** | Generate systemd unit for bare metal (`EnvironmentFile` = your .env) |
| **scli reset** | Remove SessionDB install directory. Use `--all` to also remove `.env` and `config.yaml`. |
| **scli update** | Fetch the last 5 scli releases, prompt to select a version, then download and replace the current binary (self-update). |

## Install root and layout

- **Install root** — Default: `/opt/sessiondb` (when root) or `$HOME/.local/share/sessiondb`. Override with `SESSIONDB_INSTALL_ROOT`. Under it: `versions/<tag>/` (backend binary, frontend dist, setup.sh, sessiondb.yaml) and `current` symlink → `versions/<installed-tag>`.
- **Checksums** — If the release has `checksums.txt`, backend and frontend tarballs are verified with SHA256 before extraction.

## Reset and self-update

- **scli reset** — Removes the install root’s `versions/` and `current`, and the legacy config-dir `sessiondb-install`. Use `scli reset --all` to also remove `.env` and `config.yaml` (full config reset).
- **scli update** — Fetches the latest 5 scli versions from GitHub, shows an interactive list, and after you pick one, downloads the matching binary for your OS/arch and replaces the current `scli` binary. On Windows, the new binary is written as `scli.exe.new`; replace `scli.exe` manually after closing the terminal.

## First-time flow

```bash
scli init
scli install v1.0.1
# Copy binary to /opt/sessiondb, then:
scli deploy --output sessiondb.service
sudo cp sessiondb.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable sessiondb && sudo systemctl start sessiondb
scli migrate
scli status
```

## Configuration

- **Config directory**  
  Default: `$HOME/.config/sessiondb` (or `SESSIONDB_CONFIG_DIR`).  
  Override with `--config-dir` on init, install, migrate, deploy.

- **Secrets**  
  `DB_CREDENTIAL_ENCRYPTION_KEY` and `MIGRATE_TOKEN` are generated on first `scli init` and **reused** if .env already exists (no regeneration on re-run).

- **Bare metal**  
  `scli deploy` writes a systemd unit that uses `EnvironmentFile=/path/to/.env`. Set `SESSIONDB_INSTALL_DIR` to the directory containing the server binary (default `/opt/sessiondb`).

## License

Same as SessionDB project.
