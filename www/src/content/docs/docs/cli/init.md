---
title: sitepod init
description: Initialize a new SitePod project
---

The `init` command creates a `sitepod.toml` configuration file in your project.

## Usage

```bash
sitepod init [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--name <name>` | Project name (skips prompt) |
| `--dir <path>` | Build directory (skips prompt) |

## Examples

### Interactive initialization

```bash
sitepod init
```

```
? Project name: my-blog
? Build directory (detected dist/): dist
✓ Created sitepod.toml

  Subdomain: my-blog.sitepod.dev
```

### Non-interactive

```bash
sitepod init --name my-blog --dir ./dist
```

## Auto-detection

SitePod tries to detect your build directory:

| Framework | Detected Directory |
|-----------|-------------------|
| Vite | `dist/` |
| Create React App | `build/` |
| Next.js (export) | `out/` |
| Astro | `dist/` |
| Hugo | `public/` |
| Docusaurus | `build/` |

## Generated configuration

```toml
# sitepod.toml
[project]
name = "my-blog"

[build]
directory = "./dist"
```

### Subdomain mode (default)

Your site will be available at:
- Prod: `https://my-blog.sitepod.dev`
- Beta: `https://my-blog-beta.sitepod.dev`

### Path mode

For multiple projects sharing a domain:

```toml
[project]
name = "blog-admin"
routing_mode = "path"

[build]
directory = "./dist"

[deploy.routing]
domain = "h5.company.com"
slug = "/blog-admin"
```

Your site will be at `https://h5.company.com/blog-admin/`

## Subdomain conflicts

If your chosen subdomain is taken:

```
? Project name: my-blog
✗ my-blog.sitepod.dev is already taken
? Project name: my-awesome-blog
✓ my-awesome-blog.sitepod.dev is available
```

Use a dash (`-`) to get a random suffix:

```
? Project name: my-blog
✗ my-blog.sitepod.dev is already taken
? Project name: -
✓ Using: my-blog-7x3k.sitepod.dev
```

## See also

- [sitepod deploy](/docs/cli/deploy/) - Deploy your site
- [Custom Domains](/docs/guides/custom-domains/) - Use your own domain
