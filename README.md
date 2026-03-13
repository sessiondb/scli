# SessionDB CLI (scli)

CLI for installing and running SessionDB: interactive configuration, secret generation, install, migrate, and deploy.

## Install (so you can run `scli` from anywhere)

**Option A — Go install (recommended)**  
From this directory:
```bash
go install .
```
The binary is placed in `$GOPATH/bin` or `$HOME/go/bin`. Ensure that directory is on your PATH, e.g. in `~/.zshrc` or `~/.bashrc`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Option B — Copy to a PATH directory**
```bash
go build -o scli .
sudo mv scli /usr/local/bin/   # or: cp scli ~/bin/ if ~/bin is on PATH
```

Then run `scli init`, `scli install v1.0.1`, etc. from any directory.

## Build (local only)

```bash
go build -o scli .
# or: go build -o sessiondb .
```

## Commands

| Command | Description |
|--------|-------------|
| **scli init** | Interactive prompts for DB/Redis, generates `DB_CREDENTIAL_ENCRYPTION_KEY` and `MIGRATE_TOKEN`, saves `.env` and `config.yaml` |
| **scli install \<version\>** | Download and extract SessionDB binaries (e.g. `scli install v1.0.1`) |
| **scli get \<version\> [workdir]** | Same as install; extracts to `workdir/sessiondb/` (default: current dir) |
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
