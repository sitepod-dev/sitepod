---
title: sitepod login
description: Authenticate with a SitePod server
---

The `login` command authenticates you with a SitePod server.

## Usage

```bash
sitepod login [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--endpoint <url>` | Server URL (default: https://sitepod.dev) |

## Examples

### Login to default server

```bash
sitepod login
```

### Login to custom server

```bash
sitepod login --endpoint https://sitepod.example.com
```

### Local development

```bash
sitepod login --endpoint http://localhost:8080
```

## Authentication

When you run `login`, you'll be prompted for email and password:

```bash
$ sitepod login
? Email: you@example.com
? Password: ********
◐ Authenticating
✓ Logged in
```

If the email doesn't exist, an account will be created automatically.

## Configuration

Login saves credentials to `~/.sitepod/config.toml`:

```toml
[server]
endpoint = "https://sitepod.example.com"

[auth]
token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Multiple servers

The config file stores one endpoint at a time. To switch servers:

```bash
sitepod login --endpoint https://other-server.com
```

Or use `--endpoint` with each command:

```bash
sitepod --endpoint https://other-server.com deploy
```

## See also

- [sitepod deploy](/docs/cli/deploy/) - Deploy your site
