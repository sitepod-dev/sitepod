---
title: Behind Reverse Proxy
description: Run SitePod behind Nginx, Caddy, or Traefik
---

When you have an existing reverse proxy (Nginx, Caddy, Traefik), SitePod runs on an internal port and your proxy handles SSL termination.

## Architecture

```
Internet → Nginx/Caddy/Traefik (80/443) → SitePod (8080)
                    ↓
              SSL termination
              Routing to services
```

## SitePod configuration

Create a Caddyfile for SitePod (HTTP only):

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

Start SitePod on internal port:

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v /opt/sitepod/data:/data \
  -v /opt/sitepod/Caddyfile:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

Note: `-p 127.0.0.1:8080:8080` binds only to localhost.

## Nginx configuration

Create `/etc/nginx/sites-available/sitepod`:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name sitepod.example.com *.sitepod.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name sitepod.example.com *.sitepod.example.com;

    ssl_certificate /etc/letsencrypt/live/sitepod.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sitepod.example.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Upload size limit
        client_max_body_size 100M;
    }
}
```

Enable and reload:

```bash
sudo ln -s /etc/nginx/sites-available/sitepod /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Caddy configuration

Add to your Caddyfile:

```caddyfile
sitepod.example.com, *.sitepod.example.com {
    reverse_proxy localhost:8080
}
```

Reload Caddy:

```bash
sudo systemctl reload caddy
```

## Traefik configuration

### Labels (Docker)

```yaml
services:
  sitepod:
    image: ghcr.io/sitepod/sitepod:latest
    volumes:
      - ./data:/data
      - ./Caddyfile:/etc/caddy/Caddyfile
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.sitepod.rule=HostRegexp(`sitepod.example.com`, `{subdomain:[a-z0-9-]+}.sitepod.example.com`)"
      - "traefik.http.routers.sitepod.entrypoints=websecure"
      - "traefik.http.routers.sitepod.tls.certresolver=letsencrypt"
      - "traefik.http.services.sitepod.loadbalancer.server.port=8080"
```

## Wildcard certificates

Wildcard certificates require DNS-01 challenge. Here's how to get them:

### Certbot + Cloudflare

```bash
# Install certbot with cloudflare plugin
sudo apt install certbot python3-certbot-dns-cloudflare

# Create credentials file
cat > ~/.cloudflare.ini << EOF
dns_cloudflare_api_token = YOUR_CLOUDFLARE_API_TOKEN
EOF
chmod 600 ~/.cloudflare.ini

# Get wildcard certificate
sudo certbot certonly \
  --dns-cloudflare \
  --dns-cloudflare-credentials ~/.cloudflare.ini \
  -d sitepod.example.com \
  -d "*.sitepod.example.com"
```

### acme.sh + Other DNS providers

```bash
# Install acme.sh
curl https://get.acme.sh | sh

# Example with Cloudflare
export CF_Token="your-token"
acme.sh --issue --dns dns_cf \
  -d sitepod.example.com \
  -d "*.sitepod.example.com"

# Install to Nginx
acme.sh --install-cert -d sitepod.example.com \
  --key-file /etc/nginx/ssl/sitepod.key \
  --fullchain-file /etc/nginx/ssl/sitepod.crt \
  --reloadcmd "systemctl reload nginx"
```

## Coolify

Coolify uses Traefik. You have two options:

### Option A: Coolify manages SSL

1. Add Docker resource with image `ghcr.io/sitepod/sitepod:latest`
2. Set port to `8080`
3. Mount custom Caddyfile
4. Add domains in Coolify:
   - `sitepod.example.com`
   - `*.sitepod.example.com`

### Option B: SitePod manages SSL (host network)

1. Deploy with network mode `host`
2. Disable Coolify's proxy for this service
3. SitePod binds directly to 80/443

## Next steps

- [SSL/TLS Options](/docs/self-hosting/ssl/) - More SSL configurations
- [VPS Deployment](/docs/self-hosting/vps/) - Direct deployment
