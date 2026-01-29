---
title: Quick Start
description: Deploy, preview, and rollback in under a minute
---

You need Node.js 18+ and a built static site (`dist/`, `build/`, `out/`). SitePod does not build your app.

## Deploy

```bash
cd your-project
npx sitepod deploy
```

On first run, the CLI will prompt for credentials, detect your build directory, and deploy to `beta`:

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
npx sitepod preview    # shareable URL, expires in 24h
npx sitepod rollback   # switch to a previous version
```

## Deploy to production

```bash
npx sitepod deploy --prod
```

## Configuration

The CLI creates a `sitepod.toml` in your project root:

```toml
[project]
name = "my-site"

[build]
directory = "./dist"
```

You can optionally install globally with `npm install --global sitepod` and drop `npx`.

## Using your own server

The examples above use the public instance. To deploy to your own:

```bash
npx sitepod login --endpoint https://your-server.com
npx sitepod deploy
```

See [Self-Hosting](/docs/self-hosting/overview/) for setup.
