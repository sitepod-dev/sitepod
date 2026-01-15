---
title: Quick Start
description: Deploy your first static site in 30 seconds
---

Get your static site deployed in under a minute.

## Prerequisites

- Node.js 18+ (for npx)
- A static site with a build directory (e.g., `dist/`, `build/`, `out/`)

## Deploy in 30 seconds

```bash
cd your-project
npx sitepod deploy
```

That's it! On first run, SitePod will:

1. Create an anonymous session (24h)
2. Auto-detect your build directory
3. Generate a subdomain
4. Deploy to beta environment

```
$ npx sitepod deploy

◐ Creating anonymous session
✓ Anonymous session
  expires: 24h
  next: sitepod bind

? Project name: my-site
? Build directory (detected dist/): dist
✓ Created sitepod.toml

◐ Scanning ./dist
✓ Found 42 files

◐ Planning deployment
  project: my-site
  env: beta
  files: 42
✓ Plan ready
  → 12 new, 30 reused (71%)

◐ Uploading 12 files
✓ Upload complete

◐ Committing
✓ Commit ready

◐ Releasing to beta
✓ Released to beta

  url: https://my-site.beta.sitepod.dev

⚠ Anonymous session - expires in 24h
  next: sitepod bind
```

## Keep your account

Bind an email to preserve your deployments beyond 24 hours:

```bash
sitepod bind
```

```
? Email: you@example.com
◐ Sending verification email
✓ Email sent

Next:
  - Check your inbox
  - Click the verification link
  - Account upgraded
```

## Deploy to production

Once you've verified your beta deployment:

```bash
sitepod deploy --prod
```

```
✓ Released to prod

  url: https://my-site.sitepod.dev
```

## Project configuration

SitePod creates a `sitepod.toml` file in your project:

```toml
[project]
name = "my-site"

[build]
directory = "./dist"
```

Commit this file to your repository.

## Next steps

- [Add a custom domain](/docs/guides/custom-domains/)
- [Set up CI/CD](/docs/guides/ci-cd/)
- [View CLI reference](/docs/cli/deploy/)

## Using your own server

The examples above use the public SitePod instance. To use your own server:

```bash
sitepod login --endpoint https://your-server.com
sitepod deploy
```

See [Self-Hosting](/docs/self-hosting/overview/) to run your own SitePod server.
