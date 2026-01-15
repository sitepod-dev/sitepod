#!/bin/bash
# Build Caddy with SitePod module
#
# Usage:
#   ./build-caddy.sh         # Build for current platform
#   ./build-caddy.sh linux   # Cross-compile for linux/amd64

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/bin"

mkdir -p "$OUTPUT_DIR"

TARGET_OS="${1:-$(go env GOOS)}"
TARGET_ARCH="${2:-$(go env GOARCH)}"
OUTPUT_NAME="caddy-sitepod"

if [ "$TARGET_OS" != "$(go env GOOS)" ] || [ "$TARGET_ARCH" != "$(go env GOARCH)" ]; then
    OUTPUT_NAME="caddy-sitepod-${TARGET_OS}-${TARGET_ARCH}"
fi

echo "Building Caddy with SitePod module..."
echo "  Platform: ${TARGET_OS}/${TARGET_ARCH}"
echo "  Output: ${OUTPUT_DIR}/${OUTPUT_NAME}"

cd "$SCRIPT_DIR"

CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build \
    -ldflags="-s -w" \
    -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
    ./cmd/caddy

echo ""
echo "✓ Build complete: ${OUTPUT_DIR}/${OUTPUT_NAME}"
echo ""
echo "Run with:"
echo "  ${OUTPUT_DIR}/${OUTPUT_NAME} run --config Caddyfile"
