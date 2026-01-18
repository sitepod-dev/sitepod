#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SITEPOD_BIN="${SITEPOD_BIN:-$ROOT_DIR/bin/sitepod}"
DEMO_DIR="${DEMO_DIR:-$ROOT_DIR/examples/simple-site}"

if [ ! -x "$SITEPOD_BIN" ]; then
  echo "sitepod CLI not found at $SITEPOD_BIN"
  echo "Run: make build-cli"
  exit 1
fi

if [ ! -f "$HOME/.sitepod/config.toml" ] && [ -z "${SITEPOD_TOKEN:-}" ]; then
  echo "No SitePod credentials found."
  echo "Run: $SITEPOD_BIN login --endpoint http://localhost:8080"
  echo "Or set SITEPOD_TOKEN and SITEPOD_ENDPOINT."
  exit 1
fi

if [ ! -d "$DEMO_DIR" ]; then
  echo "Demo directory not found: $DEMO_DIR"
  exit 1
fi

cd "$DEMO_DIR"

echo "==> Deploying demo site"
"$SITEPOD_BIN" deploy

echo "==> Creating preview URL"
"$SITEPOD_BIN" preview

echo "==> Rolling back (interactive)"
"$SITEPOD_BIN" rollback
