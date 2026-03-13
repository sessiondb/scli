package main

const setupScriptContent = `#!/usr/bin/env bash
# SessionDB server + UI launcher. Env vars from sessiondb.yaml are injected by scli run.
set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export SESSIONDB_BINARY="$ROOT/server/sessiondb-server"
export SESSIONDB_UI_DIR="$ROOT/ui/dist"
if [[ ! -x "$SESSIONDB_BINARY" ]]; then
  echo "Server binary not found: $SESSIONDB_BINARY"
  exit 1
fi
if [[ ! -d "$SESSIONDB_UI_DIR" ]]; then
  echo "UI dir not found: $SESSIONDB_UI_DIR"
  exit 1
fi
# Run server in foreground (scli run injects env)
exec "$SESSIONDB_BINARY"
`

const sessiondbYAMLContent = `# SessionDB runtime config. Edit env and run: scli run <version> (or scli run from install root with current symlink)
server:
  port: "8080"
  binary: "server/sessiondb-server"
ui:
  port: 3000
  dir: "ui/dist"
env:
  SERVER_PORT: "8080"
  SERVER_MODE: "release"
  DB_HOST: "localhost"
  DB_PORT: "5432"
  DB_USER: "sessiondb"
  DB_PASSWORD: ""
  DB_NAME: "sessiondb"
  DB_SSLMODE: "disable"
  REDIS_ADDR: "localhost:6379"
  REDIS_PASSWORD: ""
  REDIS_DB: "0"
`
