# SitePod

> [!WARNING]
> Early stage. Not production-ready. APIs may change.

Self-hosted static site deployment with instant rollback.

Every deploy creates an immutable snapshot (Pod). Environments are just pointers to Pods. Rollback = move the pointer. No rebuild.

SitePod does not build your app. Bring your own `dist/`.

## Features

- `sitepod deploy --prod` and you're live
- Preview URLs for sharing WIP
- Instant rollback (pointer swap, not rebuild)
- Incremental uploads (only changed files)
- Self-hosted, single binary

## Quick Start

```bash
make quick-start
make run

# In another terminal
./bin/sitepod login --endpoint http://localhost:8080
cd examples/simple-site
../../bin/sitepod deploy

open http://demo-site-beta.localhost:8080
```

### Install CLI

```bash
curl -fsSL https://get.sitepod.dev | sh

# Or from source
cd cli && cargo build --release
```

### Deploy

```bash
sitepod init
sitepod deploy          # beta
sitepod deploy --prod   # production
sitepod preview         # shareable preview URL
sitepod rollback        # interactive rollback
```

## Self-Hosting

```bash
docker-compose up -d

# Or directly
docker run -d \
  -p 80:8080 -p 443:8443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

## Architecture

Single binary: Caddy + PocketBase + SQLite.

```
Control Plane (SQLite)     Data Plane (filesystem)
  auth, audit, history       refs/{project}/{env}.json
  GC roots                   blobs/{hash[0:2]}/{hash}
         │                          ▲
         └── writes refs ──────────┘
                                    │
                              Caddy serves
                              (reads storage only)
```

Caddy never touches the DB. If the DB goes down, sites keep serving.

## Configuration

### Server

```toml
[server]
http_addr = ":80"
https_addr = ":443"

[domain]
primary = "example.com"
acme_email = "admin@example.com"

[storage]
type = "local"  # local | s3 | oss | r2
path = "/data/blobs"
```

### Project (`sitepod.toml`)

```toml
[project]
name = "my-app"

[build]
directory = "./dist"

[deploy]
ignore = ["**/*.map", ".*", "node_modules/**"]
```

## CLI

| Command | |
|---|---|
| `sitepod login` | Authenticate |
| `sitepod init` | Create sitepod.toml |
| `sitepod deploy` | Deploy to beta |
| `sitepod deploy --prod` | Deploy to production |
| `sitepod preview` | Create preview URL |
| `sitepod rollback` | Rollback (interactive) |
| `sitepod history` | View deploy history |

## API

| Endpoint | |
|---|---|
| `POST /api/v1/auth/login` | Register or login |
| `POST /api/v1/plan` | Submit manifest, get upload URLs |
| `POST /api/v1/upload/{plan_id}/{hash}` | Upload blob |
| `POST /api/v1/commit` | Finalize upload |
| `POST /api/v1/release` | Release to environment |
| `POST /api/v1/rollback` | Rollback |
| `POST /api/v1/preview` | Create preview |
| `GET /api/v1/history` | Deploy history |
| `GET /api/v1/health` | Health check |

## Development

```bash
make deps     # install dependencies
make run      # start server
make build    # build everything
make test     # run tests
```

## License

MIT
