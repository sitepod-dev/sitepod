#!/bin/sh
set -e

# Select Caddyfile based on SITEPOD_PROXY_MODE
if [ "$SITEPOD_PROXY_MODE" = "1" ] || [ "$SITEPOD_PROXY_MODE" = "true" ]; then
    echo "Running in proxy mode (port 8080, no SSL)"
    CADDYFILE="/etc/caddy/Caddyfile.proxy"
else
    echo "Running in direct mode (ports 80/443, auto SSL)"
    CADDYFILE="/etc/caddy/Caddyfile"
fi

exec caddy run --config "$CADDYFILE" --adapter caddyfile
