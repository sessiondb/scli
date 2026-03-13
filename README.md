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
| **scli install \<version\>** | Download and extract SessionDB binaries (e.g. `scli install v1.0.1`). Use `-v` or `--verbose` for detailed logs. |
| **scli get \<version\> [workdir]** | Same as install; extracts to `workdir/sessiondb/` (default: current dir). Use `-v` or `--verbose` for detailed logs. |
| **scli run \<version\> [workdir]** | Run server+UI (get if needed; injects env from sessiondb.yaml in the bundle) |
| **scli migrate** | POST `/v1/migrate` with `X-Migrate-Token` from config (run after deploy) |
| **scli status** | Check if server is reachable (GET /health) |
| **scli deploy** | Generate systemd unit for bare metal (`EnvironmentFile` = your .env) |

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
