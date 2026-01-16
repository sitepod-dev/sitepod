# SitePod Operations Manual

## Architecture

SitePod runs as a **single binary** containing:

- **Caddy** (port 80/443) - HTTP server, automatic HTTPS, static file serving
- **PocketBase** (embedded) - API, auth, SQLite database

```
Client ──► Caddy (80/443) ──┬─► /api/v1/* ──► Embedded PocketBase API
                           └─► Static files (from storage)
```

---

## Quick Start

```bash
docker run -d \
  --name sitepod \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

Caddy handles automatic HTTPS via Let's Encrypt.

---

## DNS Configuration

SitePod can be deployed on a root domain or subdomain:

### Root Domain (e.g., `sitepod.example.com`)

```
A    sitepod.example.com         → <server-ip>  (Console + API)
A    *.sitepod.example.com       → <server-ip>  (User sites)
```

### Subdomain (e.g., `pods.sitepod.example.com`)

```
A    pods.sitepod.example.com         → <server-ip>  (Console + API)
A    *.pods.sitepod.example.com       → <server-ip>  (User sites)
```

Set `SITEPOD_DOMAIN=pods.sitepod.example.com` and project URLs become:
- Production: `myapp.pods.sitepod.example.com`
- Beta: `myapp-beta.pods.sitepod.example.com`

> Note: Beta uses `-beta` suffix, not subdomain, so only 2 DNS records needed.

---

## Docker Image Targets

```bash
# Full image (default) - Caddy with embedded API
docker build --target full -t sitepod .

# CLI only
docker build --target cli -t sitepod-cli .
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SITEPOD_DOMAIN` | Yes | `localhost:8080` | Base domain (can be subdomain) |
| `SITEPOD_STORAGE_TYPE` | No | `local` | `local`, `s3`, `oss`, `r2` |
| `SITEPOD_DATA_DIR` | No | `/data` | Data directory |
| `SITEPOD_GC_ENABLED` | No | `true` | Garbage collection |
| `SITEPOD_ADMIN_EMAIL` | No | `admin@sitepod.local` | Default admin email |
| `SITEPOD_ADMIN_PASSWORD` | No | `sitepod123` | Default admin password |

> **Security Note**: Change the default admin credentials in production by setting `SITEPOD_ADMIN_EMAIL` and `SITEPOD_ADMIN_PASSWORD` environment variables. The admin account is only created on first startup if no admin exists.

### For S3/R2/OSS Storage

| Variable | Description |
|----------|-------------|
| `SITEPOD_S3_BUCKET` | Bucket name |
| `SITEPOD_S3_REGION` | Region (`auto` for R2) |
| `SITEPOD_S3_ENDPOINT` | Custom endpoint URL |
| `AWS_ACCESS_KEY_ID` | Access key |
| `AWS_SECRET_ACCESS_KEY` | Secret key |

---

## Storage Backends

### Local (Default)

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

### Cloudflare R2

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_STORAGE_TYPE=r2 \
  -e SITEPOD_S3_BUCKET=sitepod-data \
  -e SITEPOD_S3_REGION=auto \
  -e SITEPOD_S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com \
  -e AWS_ACCESS_KEY_ID=xxx \
  -e AWS_SECRET_ACCESS_KEY=xxx \
  ghcr.io/sitepod/sitepod:latest
```

### AWS S3

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_STORAGE_TYPE=s3 \
  -e SITEPOD_S3_BUCKET=sitepod-data \
  -e SITEPOD_S3_REGION=us-east-1 \
  ghcr.io/sitepod/sitepod:latest
```

---

## SSL/TLS Options

### Option 1: Direct (SitePod manages SSL)

Caddy automatically obtains Let's Encrypt certificates:

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

Requirements:
- Ports 80 and 443 accessible from internet
- DNS points to your server
- For wildcards, use DNS challenge (see below)

### Option 2: Behind Cloudflare Proxy (Recommended)

Cloudflare handles SSL, SitePod runs HTTP only:

```bash
# Download example Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.cloudflare

docker run -d \
  -p 80:80 \
  -v sitepod-data:/data \
  -v ./Caddyfile.cloudflare:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

See `server/examples/Caddyfile.cloudflare` for the full config.

**Cloudflare Settings:**
- SSL/TLS mode: **Flexible** (or Full with origin cert)
- DNS: Orange cloud (proxied) for all records
- Edge Certificates: Enable Universal SSL

### Option 3: Behind Reverse Proxy (Traefik, Nginx, Coolify)

When running behind another reverse proxy:

```bash
# Download example Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.proxy

docker run -d \
  -p 8080:8080 \
  -v sitepod-data:/data \
  -v ./Caddyfile.proxy:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

See `server/examples/Caddyfile.proxy` for the full config.

Configure your reverse proxy to:
1. Route `*.sitepod.example.com` → SitePod container:8080
2. Handle SSL termination
3. Pass `Host` header correctly

### Option 4: Wildcard Certificates (DNS Challenge)

For wildcards with direct SSL:

```bash
# Download example Caddyfile
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.wildcard

docker run -d \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -v ./Caddyfile.wildcard:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e CF_API_TOKEN=your-cloudflare-token \
  ghcr.io/sitepod/sitepod:latest
```

See `server/examples/Caddyfile.wildcard` for the full config.

---

## Platform Deployment

### Standalone VPS

Direct deployment with SitePod managing SSL:

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  ghcr.io/sitepod/sitepod:latest
```

### Coolify

Coolify uses Traefik for SSL. Choose one approach:

#### Approach A: Coolify Manages SSL (Simple)

1. Add new resource → Docker Image
2. Image: `ghcr.io/sitepod/sitepod:latest`
3. **Port**: `8080` (internal only, not 80/443)
4. Environment variables:
   - `SITEPOD_DOMAIN=pods.example.com`
5. Volume: `/data`
6. Custom Caddyfile mount: `/etc/caddy/Caddyfile`

Use `Caddyfile.proxy` (see above) to run on port 8080.

In Coolify domains, add:
- `pods.example.com`
- `*.pods.example.com` (requires Coolify wildcard support)

#### Approach B: SitePod Manages SSL (Host Network)

For full control, bypass Coolify's proxy:

1. Add new resource → Docker Image
2. Image: `ghcr.io/sitepod/sitepod:latest`
3. **Network Mode**: `host`
4. Environment variables:
   - `SITEPOD_DOMAIN=pods.example.com`
5. Volume: `/data`
6. **Disable** Coolify's domain/proxy for this service

SitePod will bind directly to ports 80/443 on the host.

#### Approach C: Cloudflare + Coolify

Let Cloudflare handle SSL, Coolify routes traffic:

1. Deploy SitePod on port 8080 (Approach A)
2. Cloudflare DNS: proxy enabled (orange cloud)
3. Cloudflare SSL mode: Flexible
4. Coolify handles routing, Cloudflare handles SSL

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sitepod
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sitepod
  template:
    metadata:
      labels:
        app: sitepod
    spec:
      containers:
      - name: sitepod
        image: ghcr.io/sitepod/sitepod:latest
        ports:
        - containerPort: 8080
        env:
        - name: SITEPOD_DOMAIN
          value: sitepod.example.com
        volumeMounts:
        - name: data
          mountPath: /data
        - name: caddyfile
          mountPath: /etc/caddy/Caddyfile
          subPath: Caddyfile
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sitepod-data
      - name: caddyfile
        configMap:
          name: sitepod-caddyfile
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: sitepod-caddyfile
data:
  Caddyfile: |
    {
        admin off
        auto_https off
        order sitepod first
    }
    :8080 {
        sitepod {
            storage_path /data
            data_dir /data
            domain sitepod.example.com
        }
    }
---
apiVersion: v1
kind: Service
metadata:
  name: sitepod
spec:
  selector:
    app: sitepod
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sitepod
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - "sitepod.example.com"
    - "*.sitepod.example.com"
    secretName: sitepod-tls
  rules:
  - host: "sitepod.example.com"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sitepod
            port:
              number: 80
  - host: "*.sitepod.example.com"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sitepod
            port:
              number: 80
```

### Docker Compose

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod/sitepod:latest
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_STORAGE_TYPE=local
    volumes:
      - sitepod-data:/data
      - caddy-data:/caddy-data
      - caddy-config:/caddy-config

volumes:
  sitepod-data:
  caddy-data:
  caddy-config:
```

---

## Data Layout

```
/data/
├── blobs/{hash[0:2]}/{hash}    # Content-addressed files
├── refs/{project}/{env}.json   # Environment pointers
├── routing/index.json          # Domain routing index
├── previews/{project}/         # Preview deployments
└── pb_data/                    # PocketBase database
```

---

## Monitoring

### Health Check

```bash
curl http://localhost/api/v1/health
# or behind proxy:
curl http://localhost:8080/api/v1/health
```

```json
{"status":"healthy","database":"ok","storage":"ok","uptime":"1h23m"}
```

### Prometheus Metrics

```bash
curl http://localhost/api/v1/metrics
```

---

## Backup

### Local Storage

```bash
docker run --rm \
  -v sitepod-data:/data \
  -v $(pwd):/backup \
  alpine tar -czvf /backup/sitepod-backup.tar.gz /data
```

### S3/R2

```bash
rclone sync r2:sitepod-data ./backup/
```

---

## Troubleshooting

### SSL Certificate Issues

1. Ensure ports 80/443 are accessible (for direct mode)
2. Check Caddy logs: `docker logs sitepod`
3. For wildcards, verify DNS challenge token is correct
4. Behind proxy? Use HTTP-only Caddyfile

### Subdomain Not Working

1. Check DNS: `dig myapp.sitepod.example.com`
2. Verify `SITEPOD_DOMAIN` matches your DNS
3. Check container logs
4. Verify wildcard DNS is configured

### Port Conflicts (Coolify/Traefik)

If Traefik occupies 80/443:
- Use `Caddyfile.proxy` on port 8080
- Or use host network mode
- Or use Cloudflare for SSL termination

### Database Locked

- Ensure only one instance is running
- Check disk space
- For HA, use S3/R2 storage (SQLite still single-instance)
