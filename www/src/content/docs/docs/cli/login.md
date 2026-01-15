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
sitepod login --endpoint https://sitepod.company.com
```

### Local development

```bash
sitepod login --endpoint http://localhost:8080
```

## Authentication methods

When you run `login`, you'll be prompted to choose:

```
? Login method:
  > Anonymous (quick start, 24h limit)
    Email & Password
```

### Email authentication

1. Enter your email address
2. Enter your password

```bash
$ sitepod login
? Login method: Email & Password
? Email: you@example.com
? Password: ********
◐ Authenticating
✓ Logged in
```

### Anonymous authentication

For quick testing without email:

```bash
$ sitepod login
? Login method: Anonymous (quick start, 24h limit)
◐ Creating anonymous session
✓ Anonymous session
  expires: 24h
  next: sitepod bind
```

Anonymous limitations:
- Account expires in 24 hours
- Deployments are deleted when account expires
- Can be upgraded with `sitepod bind`

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

- [sitepod bind](/docs/cli/bind/) - Upgrade anonymous account
- [sitepod deploy](/docs/cli/deploy/) - Deploy your site
