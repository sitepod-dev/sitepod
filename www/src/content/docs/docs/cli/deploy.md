---
title: sitepod deploy
description: Deploy your static site to SitePod
---

The `deploy` command uploads your static files and creates a new deployment.

## Usage

```bash
sitepod deploy [source] [options]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `[source]` | Source directory to deploy (default: current directory) |

## Options

| Option | Description |
|--------|-------------|
| `--prod` | Deploy to production environment (default: beta) |
| `-n, --name <name>` | Project name (overrides config) |
| `-c, --concurrent <num>` | Number of concurrent uploads (default: 20) |
| `-y, --yes` | Skip confirmation prompts (for CI/CD) |

## Examples

### Deploy to beta

```bash
sitepod deploy
```

### Deploy to production

```bash
sitepod deploy --prod
```

### Specify source directory

```bash
sitepod deploy ./build
```

### CI/CD usage (skip prompts)

```bash
sitepod deploy --prod --yes
```

## What happens

1. **Scan**: Reads all files in your build directory
2. **Hash**: Computes BLAKE3 hash for each file
3. **Plan**: Sends manifest to server, gets list of missing files
4. **Upload**: Uploads only new/changed files
5. **Commit**: Creates new Image (deployment record)
6. **Release**: Points environment ref to new Image

## First-time setup

If you haven't logged in, `deploy` will prompt you:

```bash
$ sitepod deploy

# Not logged in? Prompts for credentials
? Email: you@example.com
? Password: ********
✓ Logged in

# No sitepod.toml? Prompts for project setup
? Project name: my-site
? Build directory (detected dist/): dist
✓ Created sitepod.toml

# Then deploys
◐ Scanning ./dist
✓ Found 42 files
◐ Uploading 42 files
✓ Upload complete
◐ Releasing to beta
✓ Released to beta

  url: https://my-site-beta.sitepod.dev
```

## Incremental uploads

SitePod only uploads files that changed:

```bash
$ sitepod deploy
◐ Scanning ./dist
✓ Found 42 files
◐ Uploading 3 files  # Only 3 files changed
✓ Upload complete
◐ Releasing to beta
✓ Released to beta
```

This makes subsequent deployments fast.

## Git integration

If deploying from a git repository, SitePod captures:
- Current commit hash
- Branch name

This information appears in deployment history for traceability.

## Configuration

Deploy reads from `sitepod.toml`:

```toml
[project]
name = "my-site"

[build]
directory = "./dist"
```

Override with `--dir` flag if needed.

## See also

- [sitepod init](/docs/cli/init/) - Initialize project
- [sitepod preview](/docs/cli/preview/) - Create preview deployment
- [sitepod history](/docs/cli/history/) - View deployment history
