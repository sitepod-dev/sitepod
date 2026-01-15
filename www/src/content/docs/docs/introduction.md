---
title: Introduction
description: What is SitePod and why you should use it
---

SitePod is a **self-hosted static site deployment platform** that treats deployments as immutable snapshots. Deploy once, rollback in seconds.

## What is SitePod?

SitePod provides a simple way to deploy static websites with:

- **Single binary** - One executable contains the HTTP server (Caddy), API, and SQLite database
- **Instant rollbacks** - Switch between any previous deployment version instantly
- **Zero config** - Auto-detect build directories, generate subdomains, handle SSL
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

## Comparison

| Feature | SitePod | Vercel/Netlify |
|---------|---------|----------------|
| Self-hosted | Yes | No |
| Pricing | Free (your infra) | Tiered |
| Control | Full | Limited |
| Rollback | Instant | Usually instant |
| Custom domains | Yes | Yes |
| Build service | No (BYO) | Yes |

SitePod focuses on **deployment and hosting** - bring your own build system (npm, Vite, Next export, etc.).

## Next Steps

- [Quick Start](/docs/quickstart/) - Deploy your first site
- [Core Concepts](/docs/concepts/) - Understand the architecture
- [Self-Hosting](/docs/self-hosting/overview/) - Run your own SitePod server
