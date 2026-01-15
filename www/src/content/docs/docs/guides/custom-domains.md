---
title: Custom Domains
description: Use your own domain with SitePod
---

Bind your own domain name to your SitePod deployment.

## Add a custom domain

```bash
sitepod domain add blog.example.com
```

Output:
```
Add this DNS record to verify ownership:

  Type: CNAME
  Name: blog.example.com
  Value: my-site.sitepod.dev

After adding, run: sitepod domain verify blog.example.com
```

## Verify ownership

After adding the DNS record:

```bash
sitepod domain verify blog.example.com
```

```
✓ Domain verified

Your site is now accessible at:
  https://my-site.sitepod.dev    (system domain)
  https://blog.example.com       (custom domain)
```

## DNS configuration

### CNAME record (recommended)

Point your domain to your SitePod subdomain:

| Type | Name | Value |
|------|------|-------|
| CNAME | blog | my-site.sitepod.dev |

### A record (apex domains)

For apex domains (e.g., `example.com` without `www`), use A records:

| Type | Name | Value |
|------|------|-------|
| A | @ | `<sitepod-server-ip>` |

:::note
Some DNS providers support CNAME flattening for apex domains (Cloudflare, NS1). This is the preferred method.
:::

## Multiple domains

Add multiple domains to the same project:

```bash
sitepod domain add blog.example.com
sitepod domain add www.example.com
sitepod domain add example.com
```

All domains will serve the same content.

## List domains

```bash
sitepod domain list
```

```
Domain                  Status      Added
my-site.sitepod.dev    system      -
blog.example.com       verified    2 days ago
www.example.com        pending     1 hour ago
```

## Remove a domain

```bash
sitepod domain remove blog.example.com
```

```
? Remove blog.example.com? Yes
✓ Domain removed
```

## Rename system subdomain

Change your auto-assigned subdomain:

```bash
sitepod domain rename
```

```
Current subdomain: my-site-7x3k.sitepod.dev

? New subdomain: my-awesome-site
✓ my-awesome-site.sitepod.dev is available
✓ Subdomain renamed

Your site is now at:
  https://my-awesome-site.sitepod.dev
  https://my-awesome-site.beta.sitepod.dev
```

## SSL certificates

SitePod (via Caddy) automatically obtains SSL certificates for custom domains:

1. Domain must be verified (DNS pointing to SitePod)
2. Certificate issued on first HTTPS request
3. Automatic renewal before expiry

If using Cloudflare proxy (orange cloud), Cloudflare handles SSL instead.

## Troubleshooting

### Domain not verifying

1. Check DNS propagation: `dig blog.example.com`
2. Ensure CNAME points to your SitePod subdomain
3. Wait for DNS propagation (up to 48 hours)

### SSL certificate error

1. Ensure domain points to SitePod server
2. Check that ports 80/443 are accessible
3. View server logs for certificate errors

### Cloudflare users

If using Cloudflare proxy:
- Set SSL/TLS mode to "Full" or "Full (strict)"
- Orange cloud (proxied) is fine
- Cloudflare handles the SSL certificate

## See also

- [sitepod domain add](/docs/cli/domain/) - CLI reference
- [SSL/TLS Options](/docs/self-hosting/ssl/) - Certificate configuration
