---
title: Core Concepts
description: Understanding SitePod's architecture and terminology
---

## Pods

A Pod is an immutable snapshot of your static files, created on every deploy. Files are content-addressed by BLAKE3 hash, so identical files across Pods are stored once.

```
Pod v1: { index.html, app.js, style.css }
Pod v2: { index.html, app.js (updated), style.css }
Pod v3: { index.html, app.js, style.css, new-page.html }
```

## Environments

Environments are named pointers to Pods. Built-in: `beta` and `prod`.

```
beta  → Pod v3 (latest)
prod  → Pod v2 (stable)
```

Rollback moves the pointer:

```
# Before
prod → Pod v3

# After: sitepod rollback --to v2
prod → Pod v2
```

Atomic update, no files copied.

## Images

An Image is the metadata record of a Pod — a mapping from paths to content hashes:

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

Creating an Image doesn't duplicate blobs. It records which ones belong together.

## Projects

A Project groups deployments for a single site. Each project has a subdomain (e.g., `my-site.sitepod.dev`), version history, environment refs, and optionally custom domains.

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

The serving path reads from `refs/` and `blobs/` only — no database dependency. If the DB goes down, sites keep working.

The management path (auth, audit logs, history) uses SQLite via PocketBase. It is not in the critical serving path.

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

Ref files are cached with a 5s TTL.
