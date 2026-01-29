---
title: Introduction
description: What is SitePod and why you should use it
---

SitePod is a **self-hosted static release & rollback platform**. Single binary. Deploy a directory, roll back in seconds.

## What it does

You give SitePod a build directory (`dist/`, `build/`, `out/`). It snapshots the files into an immutable **Pod**, then points an environment (beta, prod) at that Pod. Rollback means moving the pointer — no file copying, instant.

```
Upload files → Create Pod (snapshot) → Point environment to Pod
                                              ↓
                                      Rollback = move pointer
```

One executable contains the HTTP server (Caddy), API, and SQLite database. No Docker, no external dependencies.

## Where it fits

Think Surge, but self-hosted.

| | Directory upload | Git-driven |
|---|---|---|
| **Platform-hosted** | Surge | GitHub Pages / Cloudflare Pages |
| **Self-hosted** | **SitePod** | DIY CI + OSS/CDN |

SitePod focuses on **release and rollback**. Bring your own build system.

## What you get

- `sitepod deploy` — one command, done
- Immutable versioned snapshots with content deduplication
- Built-in beta and prod environments
- Preview URLs for sharing work-in-progress
- Local or S3 storage backends
- Automatic HTTPS via Let's Encrypt
- API tokens for CI/CD
- Runs on a $5/mo VPS

## When to use something else

- You want managed builds (Vercel, Netlify)
- You want Git-push-to-deploy without running a server

## Next steps

- [Quick Start](/docs/quickstart/) — deploy your first site
- [Core Concepts](/docs/concepts/) — how Pods and environments work
- [Self-Hosting](/docs/self-hosting/overview/) — run your own server
