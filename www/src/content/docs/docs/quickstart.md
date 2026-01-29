---
title: Quick Start
description: Deploy, preview, and rollback in under a minute
---

## Prerequisites

- Node.js 18+
- A built static site (`dist/`, `build/`, `out/`)
- SitePod does **not** build your app — run your build first

## Deploy

```bash
cd your-project
npx sitepod deploy
```

First run prompts for email/password, detects your build directory, and deploys to beta:

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

## Preview and rollback

```bash
# Shareable preview URL (expires in 24h)
npx sitepod preview

# Roll back to a previous version
npx sitepod rollback
```

## Deploy to production

```bash
npx sitepod deploy --prod
```

## Install globally (optional)

```bash
npm install --global sitepod
```

Then drop `npx` from all commands.

## Configuration

SitePod creates `sitepod.toml` in your project root:

```toml
[project]
name = "my-site"

[build]
directory = "./dist"
```

Commit this file.

## Using your own server

The examples above use the public SitePod instance. To deploy to your own:

```bash
npx sitepod login --endpoint https://your-server.com
npx sitepod deploy
```

See [Self-Hosting](/docs/self-hosting/overview/) to set up a server.

## Next steps

- [Add a custom domain](/docs/guides/custom-domains/)
- [Set up CI/CD](/docs/guides/ci-cd/)
- [CLI reference](/docs/cli/deploy/)
