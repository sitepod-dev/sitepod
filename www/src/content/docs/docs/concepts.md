---
title: Core Concepts
description: Understanding SitePod's architecture and terminology
---

## Pods

A Pod is an immutable snapshot of your static files. Created on every deploy, never modified after.

Files are content-addressed (BLAKE3 hash), so identical files are stored once regardless of how many Pods reference them.

```
Pod v1: { index.html, app.js, style.css }
Pod v2: { index.html, app.js (updated), style.css }
Pod v3: { index.html, app.js, style.css, new-page.html }
```

## Environments

Environments are named pointers (refs) to Pods. Built-in: `beta` and `prod`.

```
beta  → Pod v3 (latest)
prod  → Pod v2 (stable)
```

Rollback moves the pointer. That's it.

```
# Before
prod → Pod v3

# After: sitepod rollback --to v2
prod → Pod v2
```

No files copied. Atomic pointer update. Site serves the previous version immediately.

## Images (Manifests)

An Image is the metadata record of a Pod — a mapping from file paths to content hashes:

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

Creating a new Image doesn't duplicate files. It just records which blobs belong together.

## Projects

A Project groups deployments for a single site. It has:

- A subdomain (e.g., `my-site.sitepod.dev`)
- Version history (Images)
- Environment refs (beta, prod)
- Optional custom domains

## Storage layout

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

## Control plane vs data plane

The serving path (data plane) reads only from `refs/` and `blobs/`. No database dependency. If the DB goes down, sites keep working.

The management path (control plane) uses SQLite via PocketBase for auth, audit logs, and history. Not in the critical serving path.

## Request flow

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

Ref files are cached (5s TTL) — fast enough for serving, short enough for quick rollbacks.

## Next steps

- [CLI Reference](/docs/cli/deploy/) — available commands
- [Self-Hosting](/docs/self-hosting/overview/) — architecture details
- [API Reference](/docs/api/overview/) — build integrations
