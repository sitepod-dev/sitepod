# SitePod

**Self-hosted static deployments. Immutable deploys, instant rollback.**

SitePod treats every deployment as an immutable **Pod** — a content-addressed snapshot of your site. Environments (prod, beta, preview) are just refs pointing to pods. Switch versions in milliseconds, not minutes.

- 🚀 One command to deploy: `sitepod deploy --prod`
- ⚡ Instant rollback: switch refs, not rebuild
- 📦 Incremental uploads: only upload what changed
- 🔒 Self-hosted: your data, your infrastructure

## Quick Start

### Local Testing (30 seconds)

```bash
# Build everything
make quick-start

# Start server (in terminal 1)
make run

# Login (in terminal 2) - creates account if new email
./bin/sitepod login --endpoint http://localhost:8080
# Enter your email and password

# Deploy example site
cd examples/simple-site
../../bin/sitepod deploy

# Visit your site
open http://demo-site-beta.localhost:8080
```

### Install CLI

```bash
# macOS/Linux
curl -fsSL https://get.sitepod.dev | sh

# Or build from source
cd cli && cargo build --release
```

### Deploy Your Site

```bash
# Initialize project
sitepod init

# Deploy to beta
sitepod deploy

# Deploy to production
sitepod deploy --prod

# Create preview
sitepod preview

# Rollback
sitepod rollback
```

## Self-Hosting

### Docker

```bash
# Start with docker-compose
docker-compose up -d

# Or use the image directly
docker run -d \
  -p 80:8080 -p 443:8443 -p 8090:8090 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

### Binary

```bash
# Download and run
./sitepod-server serve --http=:8080 --data=/var/sitepod-data
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Control Plane                            │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                      PocketBase                            │  │
│  │  Auth │ API │ Admin UI │ GC/Cleanup                       │  │
│  │                    ↓                                       │  │
│  │  SQLite (Control Plane SSOT): auth, audit, history, GC    │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                          Data Plane                              │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                   Storage Backend                          │  │
│  │   refs/{project}/{env}.json  ← Caddy reads directly        │  │
│  │   blobs/{hash[0:2]}/{hash}   ← File content                │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              ↑                                   │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                      Caddy Server                          │  │
│  │   TLS (ACME) │ SitePod Module │ Reverse Proxy             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**Key invariant**: Caddy serves requests by reading Storage only, never the DB. DB failure doesn't affect live sites.

## Configuration

### Server (`/etc/sitepod/config.toml`)

```toml
[server]
http_addr = ":80"
https_addr = ":443"
admin_addr = ":8090"

[domain]
primary = "example.com"
acme_email = "admin@example.com"

[storage]
type = "local"  # local | s3 | oss | r2
path = "/data/blobs"

[gc]
enabled = true
interval = "24h"
grace_period = "1h"
min_versions = 5
```

### CLI (`sitepod.toml`)

```toml
[project]
name = "my-app"

[build]
directory = "./dist"

[deploy]
ignore = ["**/*.map", ".*", "node_modules/**"]
concurrent = 20
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `sitepod login` | Login to server |
| `sitepod init` | Initialize project configuration |
| `sitepod deploy` | Deploy to beta (default) |
| `sitepod deploy --prod` | Deploy to production |
| `sitepod preview` | Create preview deployment |
| `sitepod rollback` | Rollback to previous version |
| `sitepod history` | View deployment history |

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /api/v1/auth/login` | Register or login with email/password |
| `GET /api/v1/auth/info` | Get current user info |
| `POST /api/v1/plan` | Submit file manifest, get upload URLs |
| `POST /api/v1/upload/{plan_id}/{hash}` | Upload blob (direct mode) |
| `POST /api/v1/commit` | Confirm upload, create image |
| `POST /api/v1/release` | Release image to environment |
| `POST /api/v1/rollback` | Rollback to previous image |
| `POST /api/v1/preview` | Create preview deployment |
| `GET /api/v1/history` | Get deployment history |
| `GET /api/v1/current` | Get current deployment |
| `GET /api/v1/health` | Health check |
| `GET /api/v1/metrics` | Prometheus metrics |

## Development

```bash
# Install dependencies
make deps

# Run server
make run

# Build everything
make build

# Run tests
make test

# Docker build
make docker-build
```

## Documentation

- [Product Requirements](./prd.md)
- [Technical Design](./tdd.md)
- [Operations Manual](./docs/ops.md)
- [Brand Guidelines](./brand.md)

## License

MIT
