---
title: API Endpoints
description: Detailed API endpoint reference
---

Complete reference for all SitePod API endpoints.

## Health & Monitoring

### GET /health

Check server health status.

```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "database": "ok",
  "storage": "ok",
  "uptime": "24h15m"
}
```

### GET /metrics

Prometheus-format metrics endpoint.

```http
GET /api/v1/metrics
```

**Response:**
```
# HELP sitepod_up SitePod is running
# TYPE sitepod_up gauge
sitepod_up 1
# HELP sitepod_uptime_seconds Uptime in seconds
# TYPE sitepod_uptime_seconds counter
sitepod_uptime_seconds 3600
```

## Authentication

### POST /auth/login

Login or register with email and password. Creates account if email doesn't exist.

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "you@example.com",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "abc123"
}
```

### GET /auth/info

Get current user information.

```http
GET /api/v1/auth/info
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "abc123",
  "email": "you@example.com",
  "is_admin": false
}
```

### DELETE /account

Delete your account and all projects.

```http
DELETE /api/v1/account
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Account deleted successfully",
  "deleted_projects": 3
}
```

## Projects

### GET /projects

List your projects.

```http
GET /api/v1/projects
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "abc123",
    "name": "my-site",
    "subdomain": "my-site",
    "owner_id": "user123",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T12:00:00Z"
  }
]
```

### GET /subdomain/check

Check if a subdomain is available.

```http
GET /api/v1/subdomain/check?subdomain=my-site
```

**Response (available):**
```json
{
  "available": true
}
```

**Response (taken):**
```json
{
  "available": false,
  "suggestion": "my-site-a1b2"
}
```

## Deployments

### POST /plan

Start a deployment by submitting file manifest.

```http
POST /api/v1/plan
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "files": [
    {
      "path": "index.html",
      "blake3": "abc123...",
      "sha256": "def456...",
      "size": 1024,
      "content_type": "text/html"
    }
  ],
  "git": {
    "commit": "abc1234",
    "branch": "main",
    "message": "Deploy v1.0"
  }
}
```

**Response:**
```json
{
  "plan_id": "plan_abc123",
  "content_hash": "xyz789...",
  "upload_mode": "direct",
  "missing": [
    {
      "path": "app.js",
      "hash": "abc123...",
      "size": 45678,
      "upload_url": "/api/v1/upload/plan_abc123/abc123..."
    }
  ],
  "reusable": 10
}
```

### POST /upload/{plan_id}/{hash}

Upload a file blob (direct mode).

```http
POST /api/v1/upload/plan_abc123/abc123...
Content-Type: application/octet-stream
Authorization: Bearer <token>

<binary file content>
```

**Response:** `200 OK`

### POST /commit

Finalize deployment and create image.

```http
POST /api/v1/commit
Content-Type: application/json
Authorization: Bearer <token>

{
  "plan_id": "plan_abc123"
}
```

**Response:**
```json
{
  "image_id": "img_xyz789",
  "content_hash": "abc123...",
  "files_count": 42,
  "total_size": 1234567,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### POST /release

Point environment to an image.

```http
POST /api/v1/release
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "env": "prod",
  "image_id": "img_xyz789"
}
```

**Response:**
```json
{
  "project": "my-site",
  "env": "prod",
  "image_id": "img_xyz789",
  "url": "https://my-site.example.com"
}
```

### POST /rollback

Switch environment to a previous image.

```http
POST /api/v1/rollback
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "env": "prod",
  "image_id": "img_abc456"
}
```

**Response:**
```json
{
  "project": "my-site",
  "env": "prod",
  "image_id": "img_abc456",
  "url": "https://my-site.example.com"
}
```

### POST /preview

Create a preview deployment.

```http
POST /api/v1/preview
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "image_id": "img_xyz789",
  "slug": "pr-42",
  "expires_in": 86400
}
```

**Response:**
```json
{
  "url": "https://my-site--pr-42.preview.example.com",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

## Information

### GET /current

Get current deployment info.

```http
GET /api/v1/current?project=my-site&env=prod
Authorization: Bearer <token>
```

**Response:**
```json
{
  "project": "my-site",
  "env": "prod",
  "image_id": "img_xyz789",
  "deployed_at": "2024-01-15T10:30:00Z"
}
```

### GET /history

Get deployment history.

```http
GET /api/v1/history?project=my-site&limit=10
Authorization: Bearer <token>
```

### GET /images

Get images (deployments) for a project.

```http
GET /api/v1/images?project=my-site
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "img_xyz789",
    "content_hash": "abc123...",
    "files_count": 42,
    "total_size": 1234567,
    "deployed_to": ["prod", "beta"],
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

## Domains

### GET /domains

List project domains.

```http
GET /api/v1/domains?project=my-site
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "dom_123",
    "domain": "my-site.example.com",
    "slug": "/",
    "type": "system",
    "status": "active",
    "is_primary": true
  },
  {
    "id": "dom_456",
    "domain": "blog.custom.com",
    "slug": "/",
    "type": "custom",
    "status": "pending",
    "is_primary": false
  }
]
```

### POST /domains

Add a custom domain.

```http
POST /api/v1/domains
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "domain": "www.example.com",
  "slug": "/"
}
```

**Response:**
```json
{
  "id": "dom_789",
  "domain": "www.example.com",
  "status": "pending",
  "verification_token": "sitepod-verify-abc123"
}
```

### POST /domains/verify

Verify domain ownership via DNS.

```http
POST /api/v1/domains/verify
Content-Type: application/json
Authorization: Bearer <token>

{
  "domain": "www.example.com"
}
```

**Response:**
```json
{
  "domain": "www.example.com",
  "verified": true,
  "status": "active"
}
```

### DELETE /domains

Remove a custom domain.

```http
DELETE /api/v1/domains
Content-Type: application/json
Authorization: Bearer <token>

{
  "domain": "www.example.com"
}
```

**Response:**
```json
{
  "deleted": true
}
```

### PUT /domains/rename

Rename project subdomain.

```http
PUT /api/v1/domains/rename
Content-Type: application/json
Authorization: Bearer <token>

{
  "project": "my-site",
  "new_subdomain": "my-new-site"
}
```

**Response:**
```json
{
  "old_subdomain": "my-site",
  "new_subdomain": "my-new-site",
  "domain": "my-new-site.example.com"
}
```

## Maintenance (No Auth Required)

These endpoints are intended for cron jobs or admin scripts. Protect with firewall in production.

### POST /cleanup

Clean up expired preview deployments.

```http
POST /api/v1/cleanup
```

**Response:**
```json
{
  "expired_previews_deleted": 12,
  "errors": []
}
```

### POST /gc

Garbage collect unreferenced blobs.

```http
POST /api/v1/gc
```

**Response:**
```json
{
  "referenced_blobs": 150,
  "total_blobs": 180,
  "deleted_blobs": 30,
  "freed_bytes": 52428800,
  "errors": []
}
```
