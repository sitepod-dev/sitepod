---
title: Introduction
description: What is SitePod and why you should use it
---

SitePod is a **self-hosted static release & rollback platform** for static sites. Deploy once, rollback in seconds.

## What is SitePod?

SitePod provides a simple way to release static websites with:

- **Single binary** - One executable contains the HTTP server (Caddy), API, and SQLite database
- **Instant rollbacks** - Switch between any previous deployment version instantly
- **Directory-first** - Upload your build output (`dist/`, `build/`, `out/`)
- **Self-hosted** - Run on your own infrastructure with full control

## How it works

SitePod uses a **Pod-based architecture**:

1. **Pods** are immutable, content-addressed snapshots of your static files
2. **Environments** (beta, prod) are just pointers to Pods
3. **Rollback** means updating a pointer - no file copying, instant

```
Upload files → Create Pod (snapshot) → Point environment to Pod
                                              ↓
                                      Rollback = move pointer
```

## Where SitePod fits

Think of static deployment along two axes:

- **Upload model**: directory upload vs Git-driven
- **Hosting**: platform-hosted vs self-hosted

| | Directory upload | Git-driven |
|---|---|---|
| **Platform-hosted** | Surge | GitHub Pages / Cloudflare Pages |
| **Self-hosted** | **SitePod (primary path)** | DIY CI + OSS/CDN |

If you've used Surge, the CLI flow will feel familiar — but SitePod deploys to **your** storage and domains.

## Key Features

### For Developers

- **One command deploy**: `sitepod deploy`
- **Preview deployments**: Share work-in-progress with teammates
- **Multiple environments**: Built-in beta and production
- **CI/CD friendly**: API tokens for automation

### For Operations

- **Single binary**: No complex infrastructure
- **Local or S3 storage**: Choose your storage backend
- **Automatic HTTPS**: Let's Encrypt integration via Caddy
- **Low resource usage**: Runs on minimal VPS

## When SitePod is a fit

- You need **self-hosting** for compliance, latency, or control
- You already have a build step and want **versioned releases + fast rollback**
- You want **multi-environment** releases without building a custom pipeline

## When another option may be better

- You want a **managed build service** (e.g., Vercel/Netlify)
- You prefer **Git-driven auto-deploy** without self-hosting

SitePod focuses on **release and rollback**. Bring your own build system (npm, Vite, Next export, etc.).

## Next Steps

- [Quick Start](/docs/quickstart/) - Deploy your first site
- [Core Concepts](/docs/concepts/) - Understand the architecture
- [Self-Hosting](/docs/self-hosting/overview/) - Run your own SitePod server
