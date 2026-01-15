---
title: Docker Compose
description: Deploy SitePod using Docker Compose
---

Docker Compose provides easier management and configuration for SitePod.

## Basic setup

Create `/opt/sitepod/docker-compose.yml`:

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod/sitepod:latest
    container_name: sitepod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
      - SITEPOD_STORAGE_TYPE=local
    volumes:
      - ./data:/data
      - caddy-data:/caddy-data
      - caddy-config:/caddy-config

volumes:
  caddy-data:
  caddy-config:
```

Start:

```bash
cd /opt/sitepod
docker compose up -d
```

## With Cloudflare R2

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod/sitepod:latest
    container_name: sitepod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
      - SITEPOD_STORAGE_TYPE=r2
      - SITEPOD_S3_BUCKET=sitepod-data
      - SITEPOD_S3_REGION=auto
      - SITEPOD_S3_ENDPOINT=https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com
      - AWS_ACCESS_KEY_ID=${R2_ACCESS_KEY}
      - AWS_SECRET_ACCESS_KEY=${R2_SECRET_KEY}
    volumes:
      - ./data:/data
      - caddy-data:/caddy-data
      - caddy-config:/caddy-config

volumes:
  caddy-data:
  caddy-config:
```

Create `.env` file:

```bash
R2_ACCESS_KEY=your-access-key
R2_SECRET_KEY=your-secret-key
```

## Behind reverse proxy

When running behind Nginx/Traefik:

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod/sitepod:latest
    container_name: sitepod
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"  # Only bind localhost
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
    volumes:
      - ./data:/data
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
```

Create `Caddyfile`:

```caddyfile
{
    admin off
    auto_https off
    order sitepod first
}

:8080 {
    sitepod {
        storage_path /data
        data_dir /data
        domain {$SITEPOD_DOMAIN}
    }
}
```

## Commands

```bash
# Start
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down

# Update
docker compose pull
docker compose up -d

# Restart
docker compose restart
```

## Health check

Add a health check to the compose file:

```yaml
services:
  sitepod:
    # ... other config
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

## Next steps

- [Reverse Proxy](/docs/self-hosting/reverse-proxy/) - Nginx/Caddy setup
- [Storage Backends](/docs/self-hosting/storage/) - S3 configuration
