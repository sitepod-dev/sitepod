---
title: Storage Backends
description: Configure local or cloud storage for SitePod
---

SitePod supports multiple storage backends for deployed files.

## Local storage (default)

Files stored on the local filesystem:

```bash
docker run -d \
  --name sitepod \
  -v /opt/sitepod:/data \
  -e SITEPOD_STORAGE_TYPE=local \
  ghcr.io/sitepod/sitepod:latest
```

**Data layout:**
```
/data/
├── blobs/         # Content-addressed files
├── refs/          # Environment pointers
├── previews/      # Preview deployments
└── pb_data/       # SQLite database
```

**Pros:**
- Simple setup
- No external dependencies
- Fast local access

**Cons:**
- Limited by disk size
- Requires backup strategy
- Single server only

## Cloudflare R2

S3-compatible storage with no egress fees:

```bash
docker run -d \
  --name sitepod \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_STORAGE_TYPE=r2 \
  -e SITEPOD_S3_BUCKET=sitepod-data \
  -e SITEPOD_S3_REGION=auto \
  -e SITEPOD_S3_ENDPOINT=https://ACCOUNT_ID.r2.cloudflarestorage.com \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  ghcr.io/sitepod/sitepod:latest
```

### Setting up R2

1. Go to Cloudflare dashboard → R2
2. Create a bucket (e.g., `sitepod-data`)
3. Go to **Manage R2 API Tokens**
4. Create token with **Object Read & Write** permissions
5. Copy Account ID, Access Key ID, and Secret Access Key

**Environment variables:**
| Variable | Value |
|----------|-------|
| `SITEPOD_STORAGE_TYPE` | `r2` |
| `SITEPOD_S3_BUCKET` | Your bucket name |
| `SITEPOD_S3_REGION` | `auto` |
| `SITEPOD_S3_ENDPOINT` | `https://ACCOUNT_ID.r2.cloudflarestorage.com` |
| `AWS_ACCESS_KEY_ID` | Your R2 access key |
| `AWS_SECRET_ACCESS_KEY` | Your R2 secret key |

## Amazon S3

```bash
docker run -d \
  --name sitepod \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_STORAGE_TYPE=s3 \
  -e SITEPOD_S3_BUCKET=sitepod-data \
  -e SITEPOD_S3_REGION=us-east-1 \
  ghcr.io/sitepod/sitepod:latest
```

For EC2 instances, IAM roles are recommended. Otherwise, set credentials:

```bash
-e AWS_ACCESS_KEY_ID=your-access-key \
-e AWS_SECRET_ACCESS_KEY=your-secret-key
```

### IAM policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::sitepod-data",
        "arn:aws:s3:::sitepod-data/*"
      ]
    }
  ]
}
```

## S3-compatible storage

Any S3-compatible storage works (MinIO, Backblaze B2, etc.):

```bash
docker run -d \
  --name sitepod \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_STORAGE_TYPE=s3 \
  -e SITEPOD_S3_BUCKET=sitepod-data \
  -e SITEPOD_S3_REGION=us-east-1 \
  -e SITEPOD_S3_ENDPOINT=https://s3.your-provider.com \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  ghcr.io/sitepod/sitepod:latest
```

## Environment variables reference

| Variable | Description |
|----------|-------------|
| `SITEPOD_STORAGE_TYPE` | `local`, `s3`, `r2` |
| `SITEPOD_S3_BUCKET` | Bucket name |
| `SITEPOD_S3_REGION` | AWS region (use `auto` for R2) |
| `SITEPOD_S3_ENDPOINT` | Custom endpoint URL |
| `AWS_ACCESS_KEY_ID` | Access key |
| `AWS_SECRET_ACCESS_KEY` | Secret key |

## Choosing a backend

| Need | Recommendation |
|------|----------------|
| Simple setup | Local |
| Cost-effective cloud | Cloudflare R2 (no egress fees) |
| AWS ecosystem | Amazon S3 |
| Self-hosted cloud | MinIO |

## Backup

### Local storage

```bash
tar -czvf backup.tar.gz /opt/sitepod
```

### S3/R2

Use `rclone` or native tools:

```bash
# With rclone
rclone sync r2:sitepod-data ./backup/

# With AWS CLI
aws s3 sync s3://sitepod-data ./backup/
```

## Next steps

- [VPS Deployment](/docs/self-hosting/vps/) - Complete setup guide
- [Docker Compose](/docs/self-hosting/docker-compose/) - Multi-service setup
