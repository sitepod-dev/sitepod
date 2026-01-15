---
title: sitepod console
description: Open the SitePod web console in your browser
---

The `console` command opens the SitePod web console in your default browser.

## Usage

```bash
sitepod console
```

## What happens

The command constructs the console URL based on your configured endpoint:

- **Local development**: `http://console.localhost:8080`
- **Production**: `https://console.sitepod.dev`

The URL is then opened in your default browser.

## Example

```bash
$ sitepod console
◐ Opening console
  url: https://console.sitepod.dev
```

## Requirements

You must be logged in first:

```bash
sitepod login
sitepod console
```

## See also

- [sitepod login](/docs/cli/login/) - Authenticate with server
