---
title: Core Concepts
description: Understanding SitePod's architecture and terminology
---

This page explains the key concepts behind SitePod.

## Pods (Immutable Snapshots)

A **Pod** is an immutable snapshot of your static files at a point in time.

- Created on every deployment
- Content-addressed (files are deduplicated by hash)
- Never modified after creation
- Referenced by a unique ID

```
Pod v1: { index.html, app.js, style.css }
Pod v2: { index.html, app.js (updated), style.css }
Pod v3: { index.html, app.js, style.css, new-page.html }
```

## Environments (Refs)

**Environments** are named pointers to Pods.

Built-in environments:
- `beta` - For testing/staging
- `prod` - For production

```
beta  → Pod v3 (latest)
prod  → Pod v2 (stable)
```

### How rollback works

Rollback simply moves the pointer to a previous Pod:

```
# Before rollback
prod → Pod v3

# After: sitepod rollback --to v2
prod → Pod v2
```

No files are copied. The pointer update is atomic. Your site is instantly serving the previous version.

## Content-Addressed Storage

Files are stored by their content hash (BLAKE3).

```
data/blobs/
├── a1/a1b2c3d4...  (index.html)
├── e5/e5f6g7h8...  (app.js)
└── i9/i9j0k1l2...  (style.css)
```

Benefits:
- **Deduplication**: Same file uploaded once, referenced many times
- **Fast uploads**: Only new/changed files are transferred
- **Integrity**: Content verified by hash

## Images (Manifests)

An **Image** is the metadata record of a Pod:

```json
{
  "id": "img_abc123",
  "files": {
    "/index.html": "a1b2c3d4...",
    "/app.js": "e5f6g7h8...",
    "/style.css": "i9j0k1l2..."
  },
  "created_at": "2024-01-15T10:30:00Z",
  "git_commit": "abc1234"
}
```

The Image points to blobs in storage. Creating a new Image doesn't duplicate files.

## Projects

A **Project** groups deployments for a single site:

- Has a unique subdomain (e.g., `my-site.sitepod.dev`)
- Contains multiple Images (version history)
- Has environment refs (beta, prod)
- Can have custom domains

## Storage Layout

```
data/
├── blobs/           # Content-addressed files
│   ├── a1/
│   │   └── a1b2c3d4...
│   └── e5/
│       └── e5f6g7h8...
├── refs/            # Environment pointers
│   └── my-project/
│       ├── beta.json
│       └── prod.json
├── previews/        # Temporary preview deployments
└── pb_data/         # SQLite database
```

## Control Plane vs Data Plane

SitePod separates concerns:

**Data Plane** (serving requests):
- Reads only from `refs/` and `blobs/`
- No database dependency
- If DB goes down, sites keep working

**Control Plane** (management):
- SQLite database via PocketBase
- Handles auth, audit logs, history
- Not in the critical path for serving

This separation means a database issue won't take down your sites.

## Request Flow

```
Request: https://my-site.sitepod.dev/app.js
                    │
                    ▼
┌─────────────────────────────────────┐
│ 1. Parse hostname → project: my-site│
│ 2. Read refs/my-site/prod.json      │
│ 3. Look up /app.js in manifest      │
│ 4. Serve blobs/{hash}               │
└─────────────────────────────────────┘
```

The ref file is cached (5s TTL) for performance while allowing quick updates.

## Next Steps

- [CLI Reference](/docs/cli/deploy/) - Available commands
- [Self-Hosting](/docs/self-hosting/overview/) - Architecture details
- [API Reference](/docs/api/overview/) - Build integrations
