---
title: SSL/TLS Options
description: Configure HTTPS for SitePod
---

SitePod uses Caddy which provides automatic HTTPS. This guide covers different SSL/TLS configurations.

## Option 1: Direct (default)

Caddy automatically obtains Let's Encrypt certificates:

```bash
docker run -d \
  --name sitepod \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

**Requirements:**
- Ports 80 and 443 accessible from internet
- DNS pointing to your server
- For wildcard subdomains, use DNS challenge (see below)

**How it works:**
1. First request triggers certificate issuance
2. Caddy handles renewal automatically
3. HTTP-01 challenge validates domain ownership

## Option 2: Behind Cloudflare (recommended)

Let Cloudflare handle SSL. SitePod runs HTTP only:

```bash
# Download Cloudflare Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.cloudflare

docker run -d \
  --name sitepod \
  -p 80:80 \
  -v sitepod-data:/data \
  -v ./Caddyfile.cloudflare:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

**Cloudflare settings:**
- SSL/TLS mode: **Flexible** (or Full with origin cert)
- DNS: Orange cloud (proxied) for all records
- Edge Certificates: Enable Universal SSL

**Benefits:**
- No certificate management on server
- Cloudflare handles HTTPS
- Built-in DDoS protection
- Wildcard support included

## Option 3: Wildcard certificates

For wildcard certificates (`*.sitepod.example.com`), use DNS-01 challenge:

```bash
# Download wildcard Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.wildcard

docker run -d \
  --name sitepod \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -v ./Caddyfile.wildcard:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e CF_API_TOKEN=your-cloudflare-token \
  ghcr.io/sitepod-dev/sitepod:latest
```

**DNS providers supported:**
- Cloudflare (recommended)
- Route53
- Google Cloud DNS
- And more via Caddy plugins

## Option 4: Behind reverse proxy

When running behind Nginx/Traefik, SitePod runs HTTP only:

```bash
# Download proxy Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.proxy

docker run -d \
  --name sitepod \
  -p 127.0.0.1:8080:8080 \
  -v sitepod-data:/data \
  -v ./Caddyfile.proxy:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

Your reverse proxy handles SSL. See [Reverse Proxy guide](/docs/self-hosting/reverse-proxy/).

## Caddyfile examples

### HTTP only (behind proxy)

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

### Wildcard with Cloudflare DNS

```caddyfile
{
    admin off
    order sitepod first
}

sitepod.example.com, *.sitepod.example.com {
    tls {
        dns cloudflare {$CF_API_TOKEN}
    }

    sitepod {
        storage_path /data
        data_dir /data
        domain sitepod.example.com
    }
}
```

## Troubleshooting

### Certificate not issued

1. Check ports are accessible: `curl -v http://your-domain.com`
2. Verify DNS: `dig your-domain.com`
3. Check Caddy logs: `docker logs sitepod`

### Rate limits

Let's Encrypt has rate limits:
- 50 certificates per domain per week
- 5 duplicate certificates per week
- 5 failed validations per hour

If you hit limits, wait or use staging:

```caddyfile
{
    acme_ca https://acme-staging-v02.api.letsencrypt.org/directory
}
```

### Wildcard not working

Wildcard certificates require DNS-01 challenge. HTTP-01 cannot validate wildcards.

Options:
1. Use Cloudflare proxy (handles wildcards automatically)
2. Configure DNS-01 with your DNS provider
3. Use separate certificates per subdomain (not recommended)

## Next steps

- [Reverse Proxy](/docs/self-hosting/reverse-proxy/) - Nginx/Caddy setup
- [Storage Backends](/docs/self-hosting/storage/) - S3 configuration
