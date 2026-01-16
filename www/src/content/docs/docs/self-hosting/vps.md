---
title: VPS Deployment
description: Deploy SitePod on a standalone VPS server
---

This guide walks through deploying SitePod on a fresh VPS.

## Prerequisites

- A VPS with Ubuntu 22.04 (or similar Linux)
- SSH access with sudo privileges
- A domain name

## Step 1: Install Docker

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and back in for group changes
exit
```

## Step 2: Open firewall ports

**Ubuntu (UFW):**

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw reload
```

**Cloud provider:**

In your cloud console (AWS, GCP, DigitalOcean, etc.), ensure your security group allows inbound TCP on ports 80 and 443.

## Step 3: Configure DNS

Point your domain to your server. Example for `sitepod.example.com`:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| A | sitepod | `<server-ip>` | Console + API |
| A | *.sitepod | `<server-ip>` | User sites |

> Beta uses `-beta` suffix (e.g., `myapp-beta.sitepod.example.com`), so only 2 records needed.

Verify DNS propagation:

```bash
dig sitepod.example.com
dig test.sitepod.example.com
```

Both should return your server IP.

## Step 4: Create data directory

```bash
sudo mkdir -p /opt/sitepod
```

## Step 5: Start SitePod

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  ghcr.io/sitepod-dev/sitepod:latest
```

:::caution
Use a strong admin password! The default is insecure.
:::

## Step 6: Verify deployment

Check container status:

```bash
docker ps
docker logs sitepod
```

Health check:

```bash
curl http://localhost/api/v1/health
# {"status":"healthy","database":"ok","storage":"ok"}
```

Access admin UI:

```
https://sitepod.example.com/_/
```

## Step 7: Deploy your first site

On your local machine:

```bash
sitepod login --endpoint https://sitepod.example.com
cd your-project
sitepod deploy
```

## Updating

```bash
# Pull latest image
docker pull ghcr.io/sitepod-dev/sitepod:latest

# Restart container
docker stop sitepod
docker rm sitepod

# Start with same config (use your original command)
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 -p 443:443 \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  ghcr.io/sitepod-dev/sitepod:latest
```

## Backup

```bash
# Stop container (optional, for consistency)
docker stop sitepod

# Backup data
tar -czvf sitepod-backup-$(date +%Y%m%d).tar.gz /opt/sitepod

# Restart
docker start sitepod
```

## Monitoring

Create a health check script at `/opt/sitepod/healthcheck.sh`:

```bash
#!/bin/bash
response=$(curl -s http://localhost/api/v1/health)
status=$(echo $response | jq -r '.status')

if [ "$status" != "healthy" ]; then
    echo "SitePod is unhealthy: $response"
    exit 1
fi
echo "SitePod is healthy"
```

Add to crontab:

```bash
# Check every 5 minutes
*/5 * * * * /opt/sitepod/healthcheck.sh >> /var/log/sitepod-health.log 2>&1
```

## Troubleshooting

### SSL certificate issues

- Ensure ports 80 and 443 are accessible from the internet
- Wait a few minutes for Caddy to obtain certificates
- Check logs: `docker logs sitepod | grep -i cert`

### Subdomain not working

- Verify wildcard DNS: `dig test.sitepod.example.com`
- Check `SITEPOD_DOMAIN` matches your DNS
- Check container logs for errors

### Database locked

- Ensure only one SitePod instance is running
- Check disk space: `df -h`

## Next steps

- [Docker Compose](/docs/self-hosting/docker-compose/) - For easier management
- [SSL/TLS Options](/docs/self-hosting/ssl/) - Cloudflare, wildcards
- [Storage Backends](/docs/self-hosting/storage/) - S3/R2 storage
