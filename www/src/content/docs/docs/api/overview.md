---
title: API Overview
description: SitePod REST API reference
---

SitePod provides a REST API for programmatic deployments and integrations.

## Base URL

```
https://your-sitepod-server.com/api/v1
```

## Authentication

All API requests require a token in the `Authorization` header:

```http
Authorization: Bearer sitepod_xxxxxxxxxxxx
```

### Get a token

Login with email and password to get a token:

```bash
sitepod login
```

Or use the API directly:

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "you@example.com",
  "password": "your-password"
}
```

## Endpoints

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/login` | POST | Login or register with email/password |
| `/auth/info` | GET | Get current user info |
| `/account` | DELETE | Delete account |

### Deployments

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/plan` | POST | Start deployment, get upload URLs |
| `/upload/{plan_id}/{hash}` | POST | Upload a file blob |
| `/commit` | POST | Complete deployment |
| `/release` | POST | Point environment to image |
| `/rollback` | POST | Switch to previous image |

### Information

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/current` | GET | Get current deployment info |
| `/history` | GET | List deployment history |
| `/health` | GET | Health check |

### Domains

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/domains` | GET | List domains |
| `/domains` | POST | Add domain |
| `/domains/{domain}/verify` | POST | Verify domain |
| `/domains/{domain}` | DELETE | Remove domain |

## Deployment flow

The typical deployment flow:

```
1. POST /plan         → Get upload URLs for missing files
2. POST /upload/...   → Upload each missing file
3. POST /commit       → Create image from manifest
4. POST /release      → Point environment to image
```

### 1. Plan

```http
POST /api/v1/plan
Content-Type: application/json
Authorization: Bearer sitepod_xxx

{
  "project": "my-site",
  "files": {
    "/index.html": "a1b2c3d4e5f6...",
    "/app.js": "f6e5d4c3b2a1...",
    "/style.css": "1a2b3c4d5e6f..."
  }
}
```

Response:
```json
{
  "plan_id": "plan_abc123",
  "missing": ["f6e5d4c3b2a1...", "1a2b3c4d5e6f..."],
  "upload_urls": {
    "f6e5d4c3b2a1...": "/api/v1/upload/plan_abc123/f6e5d4c3b2a1...",
    "1a2b3c4d5e6f...": "/api/v1/upload/plan_abc123/1a2b3c4d5e6f..."
  }
}
```

### 2. Upload missing files

```http
POST /api/v1/upload/plan_abc123/f6e5d4c3b2a1...
Content-Type: application/octet-stream
Authorization: Bearer sitepod_xxx

<file contents>
```

### 3. Commit

```http
POST /api/v1/commit
Content-Type: application/json
Authorization: Bearer sitepod_xxx

{
  "plan_id": "plan_abc123",
  "message": "v1.2.0 release"
}
```

Response:
```json
{
  "image_id": "img_xyz789",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 4. Release

```http
POST /api/v1/release
Content-Type: application/json
Authorization: Bearer sitepod_xxx

{
  "project": "my-site",
  "environment": "prod",
  "image_id": "img_xyz789"
}
```

## Error responses

Errors return appropriate HTTP status codes with JSON body:

```json
{
  "error": "invalid_token",
  "message": "The provided token is invalid or expired"
}
```

Common status codes:
- `400` - Bad request (invalid parameters)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (no permission)
- `404` - Not found
- `429` - Rate limited
- `500` - Server error

## Rate limits

Default limits:
- 100 requests per minute per token
- 10 concurrent uploads per plan

Rate limit headers:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705312800
```

## See also

- [CI/CD Integration](/docs/guides/ci-cd/) - Using the API in pipelines
- [Endpoints Reference](/docs/api/endpoints/) - Detailed endpoint docs
