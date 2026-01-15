# Caddyfile Examples

Choose the right Caddyfile based on your deployment scenario:

| File | Use Case |
|------|----------|
| `Caddyfile.direct` | Direct deployment, SitePod manages SSL via Let's Encrypt |
| `Caddyfile.cloudflare` | Behind Cloudflare proxy (orange cloud), Cloudflare handles SSL |
| `Caddyfile.proxy` | Behind reverse proxy (Traefik, Nginx, Coolify), proxy handles SSL |
| `Caddyfile.wildcard` | Direct deployment with wildcard certificates via DNS challenge |

## Usage

Mount the appropriate Caddyfile when running:

```bash
docker run -d \
  -v ./Caddyfile.proxy:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

## Environment Variables

All Caddyfiles use these environment variables:

- `SITEPOD_DOMAIN` - Your base domain (required)
- `SITEPOD_DATA_DIR` - Data directory (default: `/data`)
- `CF_API_TOKEN` - Cloudflare API token (only for `Caddyfile.wildcard`)
