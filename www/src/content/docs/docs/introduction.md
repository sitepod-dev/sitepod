---
title: Introduction
description: What is SitePod and why you should use it
---

SitePod is an open source self-hosted platform for releasing and rolling back static sites. It consists of an HTTP server (Caddy), API, and SQLite database packed into a single binary.

You upload a build directory (`dist/`, `build/`, `out/`). SitePod snapshots the files into an immutable **Pod**, then points an environment (`beta`, `prod`) at that Pod. Rollback means moving the pointer — no file copying, no rebuild.

```
Upload files → Create Pod (snapshot) → Point environment to Pod
                                              ↓
                                      Rollback = move pointer
```

The easiest way to get started is to deploy with the CLI:

```bash
npx sitepod deploy
```

On first run the CLI will prompt for credentials, detect your build directory, and deploy to the `beta` environment.

SitePod does not build your app. Run your build step first (`npm run build`, `vite build`, etc.), then deploy the output.

For self-hosting, see the [Self-Hosting guide](/docs/self-hosting/overview/). For architecture details, see [Core Concepts](/docs/concepts/).
