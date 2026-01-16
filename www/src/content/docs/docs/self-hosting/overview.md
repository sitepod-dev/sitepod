---
title: Self-Hosting Overview
description: Run SitePod on your own infrastructure
---

SitePod is designed to be self-hosted. This guide covers running your own SitePod server.

## Architecture

SitePod runs as a **single binary** containing:

- **Caddy** - HTTP server, automatic HTTPS, static file serving
- **PocketBase** (embedded) - API handlers, auth, SQLite database

```
Client → Caddy (80/443) ─┬─► /api/v1/* → Embedded API
                         └─► Static files (from storage)
```

## Requirements

### Server

| Item | Minimum | Recommended |
|------|---------|-------------|
| CPU | 1 core | 2 cores |
| RAM | 512 MB | 1 GB |
| Disk | 10 GB | 20 GB+ |
| OS | Linux (amd64/arm64) | Ubuntu 22.04 LTS |

### Network

- Public IP address
- Ports 80 (HTTP) and 443 (HTTPS) open
- A domain name with DNS pointing to your server

## Deployment options

Choose based on your setup:

| Scenario | Guide |
|----------|-------|
| Fresh VPS | [VPS Deployment](/docs/self-hosting/vps/) |
| Docker Compose | [Docker Compose](/docs/self-hosting/docker-compose/) |
| Kubernetes | [Kubernetes](/docs/self-hosting/kubernetes/) |
| Behind Nginx/Caddy | [Reverse Proxy](/docs/self-hosting/reverse-proxy/) |

## Quick Start

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

Then point your DNS to the server:

```
A    sitepod.example.com         → <server-ip>  (Console + API)
A    *.sitepod.example.com       → <server-ip>  (User sites)
```

> Note: Only 2 DNS records needed. Beta uses `-beta` suffix (e.g., `myapp-beta.sitepod.example.com`).

## Configuration

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SITEPOD_DOMAIN` | Yes | `localhost:8080` | Base domain |
| `SITEPOD_STORAGE_TYPE` | No | `local` | Storage backend |
| `SITEPOD_DATA_DIR` | No | `/data` | Data directory |
| `SITEPOD_GC_ENABLED` | No | `true` | Garbage collection |
| `SITEPOD_ADMIN_EMAIL` | No | `admin@sitepod.local` | Admin email |
| `SITEPOD_ADMIN_PASSWORD` | No | `sitepod123` | Admin password |

:::caution[Security]
Change the default admin credentials in production!
:::

### Storage backends

- **Local** (default): Files stored on disk
- **S3**: Amazon S3 or compatible (R2, MinIO)

See [Storage Backends](/docs/self-hosting/storage/) for configuration.

## Data layout

```
/data/
├── blobs/           # Content-addressed files
├── refs/            # Environment pointers
├── previews/        # Preview deployments
├── routing/         # Domain routing index
└── pb_data/         # SQLite database
```

## Admin UI

Access the PocketBase admin panel at:

```
https://sitepod.example.com/_/
```

Use your admin credentials to:
- View users and projects
- Manage settings
- Access database

## Next steps

- [VPS Deployment](/docs/self-hosting/vps/) - Step-by-step guide
- [SSL/TLS Options](/docs/self-hosting/ssl/) - HTTPS configuration
- [Storage Backends](/docs/self-hosting/storage/) - S3/R2 setup
