---
title: Quick Start
description: Deploy, preview, and rollback in under a minute
---

Get your static site released in under a minute.

## Prerequisites

- Node.js 18+ (for npx)
- A static site with a build directory (e.g., `dist/`, `build/`, `out/`)
- SitePod does **not** build your app — run your build first

## Deploy in 30 seconds

```bash
cd your-project
npx sitepod deploy
```

No install required. If you install the CLI, you can drop `npx` from all commands.

That's it! On first run, SitePod will:

1. Prompt for email & password (creates account if new)
2. Auto-detect your build directory
3. Generate a subdomain
4. Deploy to beta environment

```
$ npx sitepod deploy

? Email: you@example.com
? Password: ********
✓ Logged in

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

  url: https://my-site-beta.sitepod.dev
```

## 60-second demo: deploy → preview → rollback

After your first deploy, try the full release flow:

```bash
# Create a 24h preview URL
npx sitepod preview

# Roll back to a previous version (interactive)
npx sitepod rollback
```

Preview gives you a shareable, expiring URL. Rollback switches the environment ref to a previous Pod — no rebuild needed.

## Install the CLI (optional)

```bash
npm install --global sitepod
```

## Deploy to production

Once you've verified your beta deployment:

```bash
npx sitepod deploy --prod
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
npx sitepod login --endpoint https://your-server.com
npx sitepod deploy
```

See [Self-Hosting](/docs/self-hosting/overview/) to run your own SitePod server.
