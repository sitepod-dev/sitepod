# =============================================================================
# SitePod Multi-stage Dockerfile
# =============================================================================
# Single binary architecture: Caddy embeds PocketBase API
#
# Targets (last stage is default):
#   - cli:   CLI binary only
#   - full:  Complete server (Caddy + embedded API)
#   - caddy: Alias for full (DEFAULT)
#
# Usage:
#   docker build -t sitepod .              # builds server (default)
#   docker build -t sitepod --target cli . # builds CLI only
# =============================================================================

# -----------------------------------------------------------------------------
# Stage: Build Caddy with embedded SitePod API
# -----------------------------------------------------------------------------
FROM golang:1.21-alpine AS caddy-builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .

# Build Caddy with sitepod module (embeds PocketBase)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o caddy-sitepod \
    ./cmd/caddy

# -----------------------------------------------------------------------------
# Stage: Build Rust CLI
# -----------------------------------------------------------------------------
FROM rust:1.83-alpine AS cli-builder

WORKDIR /app

RUN apk add --no-cache musl-dev openssl-dev openssl-libs-static pkgconfig

COPY cli/Cargo.toml cli/Cargo.lock* ./

# Create dummy main.rs to cache dependencies
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release || true

# Copy actual source and rebuild
COPY cli/src ./src
RUN touch src/main.rs && cargo build --release

# -----------------------------------------------------------------------------
# Target: CLI - CLI binary only
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS cli

RUN apk add --no-cache ca-certificates

COPY --from=cli-builder /app/target/release/sitepod /usr/local/bin/sitepod

ENTRYPOINT ["sitepod"]

# -----------------------------------------------------------------------------
# Target: Full - Complete server (DEFAULT - must be last)
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS full

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

# Copy binaries
COPY --from=caddy-builder /app/caddy-sitepod /usr/local/bin/caddy
COPY --from=cli-builder /app/target/release/sitepod /usr/local/bin/sitepod

# Copy Caddyfile
COPY server/Caddyfile /etc/caddy/Caddyfile

# Create directories
RUN mkdir -p /data /caddy-data /caddy-config

# Environment
ENV SITEPOD_DATA_DIR=/data
ENV SITEPOD_STORAGE_TYPE=local
ENV XDG_DATA_HOME=/caddy-data
ENV XDG_CONFIG_HOME=/caddy-config

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1/api/v1/health || exit 1

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]

# -----------------------------------------------------------------------------
# Target: Caddy - Alias for full (also default)
# -----------------------------------------------------------------------------
FROM full AS caddy
