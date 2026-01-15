---
title: sitepod domain
description: Manage custom domains for your projects
---

The `domain` command manages custom domains for your SitePod projects.

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `add` | Add a custom domain |
| `list` | List all domains for a project |
| `verify` | Verify domain ownership via DNS |
| `remove` | Remove a domain |
| `rename` | Rename the project subdomain |

---

## sitepod domain add

Add a custom domain to a project.

### Usage

```bash
sitepod domain add <domain> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `<domain>` | Domain to add (e.g., `www.example.com`) |

### Options

| Option | Description |
|--------|-------------|
| `-p, --project <name>` | Project name (uses current directory if omitted) |
| `-s, --slug <path>` | Path slug (default: `/`) |

### Example

```bash
$ sitepod domain add www.example.com --project my-site
✓ Domain added
  domain: www.example.com
  project: my-site

⚠ Verification required
Add DNS TXT record:

  sitepod-verify=abc123xyz

ℹ Run sitepod domain verify www.example.com to verify
```

---

## sitepod domain list

List all domains configured for a project.

### Usage

```bash
sitepod domain list [options]
```

### Options

| Option | Description |
|--------|-------------|
| `-p, --project <name>` | Project name (uses current directory if omitted) |

### Example

```bash
$ sitepod domain list --project my-site
Domains: my-site

  active my-site.sitepod.dev → / [system] (primary)
  pending www.example.com → / [custom]
```

---

## sitepod domain verify

Verify domain ownership by checking DNS TXT record.

### Usage

```bash
sitepod domain verify <domain>
```

### Arguments

| Argument | Description |
|----------|-------------|
| `<domain>` | Domain to verify |

### Example

```bash
$ sitepod domain verify www.example.com
✓ Domain verified
  domain: www.example.com
```

If verification fails:

```bash
$ sitepod domain verify www.example.com
✗ Domain not verified
  domain: www.example.com

TXT record not found. Add: sitepod-verify=abc123xyz
```

---

## sitepod domain remove

Remove a custom domain from a project.

### Usage

```bash
sitepod domain remove <domain>
```

### Arguments

| Argument | Description |
|----------|-------------|
| `<domain>` | Domain to remove |

### Example

```bash
$ sitepod domain remove www.example.com
✓ Domain removed
  domain: www.example.com
```

---

## sitepod domain rename

Rename the default subdomain for a project.

### Usage

```bash
sitepod domain rename <new-subdomain> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `<new-subdomain>` | New subdomain name |

### Options

| Option | Description |
|--------|-------------|
| `-p, --project <name>` | Project name (uses current directory if omitted) |

### Example

```bash
$ sitepod domain rename my-new-site --project my-site
✓ Subdomain updated
  subdomain: my-new-site
```

Your site will now be available at `my-new-site.sitepod.dev` instead of `my-site.sitepod.dev`.

---

## DNS Configuration

When adding a custom domain, you need to configure DNS:

### For apex domains (example.com)

Add an A record pointing to your SitePod server IP.

### For subdomains (www.example.com)

Add a CNAME record pointing to your project subdomain:

```
CNAME www → my-site.sitepod.dev
```

### Verification

Add a TXT record with the verification token provided by `domain add`:

```
TXT _sitepod-verify → sitepod-verify=abc123xyz
```

## See also

- [Custom Domains Guide](/docs/guides/custom-domains/) - Detailed domain setup guide
- [sitepod deploy](/docs/cli/deploy/) - Deploy your site
