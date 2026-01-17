# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SitePod is a self-hosted static site deployment platform with a **single-binary Go server** (Caddy + embedded PocketBase) and Rust CLI.

**Core value**: Deploy once, rollback in seconds. Treats deployments as immutable "Pods" (content-addressed snapshots) with environments as ref pointers.

## Local Development

### Quick Start

```bash
# Build everything (first time)
make quick-start

# Start server (http://localhost:8080)
make run

# In another terminal, login and deploy
./bin/sitepod login --endpoint http://localhost:8080
# Enter your email and password (creates account if new)

cd examples/simple-site
../../bin/sitepod deploy

# Visit http://demo-site-beta.localhost:8080 (beta uses -beta suffix)
```

### Clean Start (Reset Everything)

```bash
# Stop server if running
pkill sitepod-server

# Remove all data and build artifacts
make clean

# Rebuild and start fresh
make build
make run
```

Or manually:

```bash
# Only reset data (keep binaries)
rm -rf data/

# Restart server
make run
```

### Data Locations

| Path | Contents |
|------|----------|
| `data/data.db` | SQLite database (users, projects, images) |
| `data/blobs/` | Deployed static files (content-addressed) |
| `data/refs/` | Environment pointers (prod/beta) |
| `bin/` | Compiled binaries |

## Make Commands

### Build & Run
```bash
make build          # Build server + CLI
make build-server   # Build server only
make build-cli      # Build CLI only
make run            # Run server (localhost:8080)
make quick-start    # First time setup: build + create data dir
```

### Testing
```bash
make test           # Run all tests
make test-server    # Go tests only
make test-cli       # Rust tests only
./test-e2e.sh       # End-to-end tests
```

### Docker
```bash
make docker-build   # Build image (sitepod:latest)
make docker-run     # Run container
make docker-stop    # Stop container
make docker-logs    # View logs
make docker-push    # Push to ghcr.io/sitepod-dev/sitepod:latest
```

### Release & Publish
```bash
make bump-patch     # Bump version x.y.Z
make bump-minor     # Bump version x.Y.0
make bump-major     # Bump version X.0.0
make release        # Create git tag and push
make npm-publish    # Publish CLI to npm
```

### Cleanup
```bash
make clean          # Remove bin/, data/, and build artifacts
```

## Project Structure

```
sitepod.dev/
├── server/                    # Go server
│   ├── cmd/caddy/            # Entry point (Caddy + embedded API)
│   ├── internal/
│   │   ├── caddy/            # Caddy module (API + static serving)
│   │   ├── storage/          # Storage backends (local, S3)
│   │   └── gc/               # Garbage collection
│   └── migrations/           # Database migrations
├── cli/                       # Rust CLI
│   ├── src/
│   │   ├── main.rs           # Entry point with clap commands
│   │   ├── api.rs            # API client
│   │   ├── config.rs         # Configuration handling
│   │   ├── hash.rs           # BLAKE3/SHA256 hashing
│   │   ├── scanner.rs        # File discovery
│   │   └── commands/         # Command implementations
│   └── Cargo.toml
├── docs/                      # Planning documents
│   ├── prd.md                # Product requirements
│   ├── tdd.md                # Technical design
│   ├── ops.md                # Operations manual
│   └── brand.md              # Brand guidelines
├── Dockerfile
└── Makefile
```

## Architecture

### Single Binary Design

SitePod runs as a **single binary** containing:
- **Caddy** - HTTP server, automatic HTTPS, static file serving
- **PocketBase** (embedded) - API handlers, auth, SQLite database

All requests go through Caddy:
- `/api/v1/*` → Embedded PocketBase API handlers
- `/*` → Static file serving for deployed sites

### Control Plane vs Data Plane Separation

- **Data Plane (SSOT)**: `refs/{project}/{env}.json` files in Storage - Caddy reads these directly, no DB dependency
- **Control Plane**: SQLite (via PocketBase) - handles auth, audit logs, history, GC roots
- **Key invariant**: Caddy serves requests by reading only Storage, never the DB. DB failure doesn't affect live sites.

### Storage Model

```
data/
├── refs/{project}/{env}.json    # Ref files containing manifest
├── blobs/{hash[0:2]}/{hash}     # Content-addressed blobs (2-char prefix sharding)
├── previews/{project}/{slug}.json
├── routing/index.json           # Domain routing index
└── pb_data/                     # PocketBase database
```

### Dual Hash Strategy

- **BLAKE3**: CAS key, deduplication, content_hash (CLI scanning - fast)
- **SHA256**: S3 upload verification via `x-amz-checksum-sha256` header

## Key Design Decisions

1. **Single binary**: Caddy embeds PocketBase API - no separate processes, no IPC
2. **Plan/Commit upload flow**: CLI sends file manifest → Server returns missing blobs → CLI uploads only missing → CLI commits
3. **Atomic ref writes**: Write to temp file, then rename (local) or copy+delete (S3)
4. **Release write order**: Storage ref file first → SQLite audit log second (ref success = release success)
5. **Ref cache TTL**: 5 seconds in Caddy for fast version switching while allowing quick updates
6. **SPA fallback**: If path not in manifest, try `index.html`

## API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /api/v1/config` | Get server configuration (domain, is_demo) |
| `POST /api/v1/auth/login` | Register or login with email/password |
| `GET /api/v1/auth/info` | Get current user info (supports admin tokens) |
| `POST /api/v1/plan` | Submit file manifest, get missing blob upload URLs |
| `POST /api/v1/upload/{plan_id}/{hash}` | Upload blob (direct mode for local storage) |
| `POST /api/v1/commit` | Confirm upload completion, create image |
| `POST /api/v1/release` | Point environment ref to image |
| `POST /api/v1/rollback` | Switch ref to previous image |
| `POST /api/v1/preview` | Create temporary preview with expiry |
| `GET /api/v1/current` | Get current deployment for environment |
| `GET /api/v1/history` | Get deployment history |
| `GET /api/v1/health` | Health check |
| `POST /api/v1/domains` | Add custom domain |
| `POST /api/v1/domains/{domain}/verify` | Verify domain ownership |
| `GET /api/v1/domains` | List domains |
| `DELETE /api/v1/domains/{domain}` | Remove domain |

## CLI Commands

```
sitepod login          # Authenticate with email/password (creates account if new)
sitepod deploy         # Deploy to beta environment
sitepod deploy --prod  # Deploy to production
sitepod init           # Initialize project config (creates sitepod.toml)
sitepod preview        # Create preview deployment
sitepod rollback       # Rollback to previous version (interactive)
sitepod history        # View deployment history
sitepod domain add     # Add custom domain
sitepod domain verify  # Verify domain ownership
sitepod domain rename  # Rename system-assigned subdomain
sitepod domain list    # List domains
sitepod domain remove  # Remove domain
sitepod delete-account # Delete your account and all projects
sitepod console        # Open SitePod console in browser
```

## Naming Conventions

- CLI command: `sitepod` (not `pod`)
- Config file: `sitepod.toml`
- Environment variables: `SITEPOD_` prefix
- Docker image: `ghcr.io/sitepod-dev/sitepod`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SITEPOD_DOMAIN` | Base domain for sites | `localhost` |
| `SITEPOD_PROXY_MODE` | Run behind reverse proxy (no TLS) | Not set |
| `SITEPOD_STORAGE_TYPE` | Storage backend: `local`, `s3`, `oss`, `r2` | `local` |
| `SITEPOD_ACCESS_LOG` | Log all static file requests | Not set |
| `SITEPOD_ADMIN_EMAIL` | PocketBase admin email (PB admin UI only) | `admin@sitepod.local` |
| `SITEPOD_ADMIN_PASSWORD` | PocketBase admin password (PB admin UI only) | `sitepod123` |
| `SITEPOD_CONSOLE_ADMIN_EMAIL` | Console admin email (users.is_admin) | Not set |
| `SITEPOD_CONSOLE_ADMIN_PASSWORD` | Console admin password (users.is_admin) | Not set |
| `IS_DEMO` | Demo mode - creates demo user | Not set |

### Demo Mode

When `IS_DEMO=1`:
- Creates demo user: `demo@sitepod.dev` / `demo123`
- Creates Console admin using `SITEPOD_ADMIN_EMAIL` / `SITEPOD_ADMIN_PASSWORD` (defaults to `admin@sitepod.local` / `sitepod123`)
- Console shows demo credentials and admin credentials on login page
- PocketBase admin (`/_/`) uses same credentials as Console admin
