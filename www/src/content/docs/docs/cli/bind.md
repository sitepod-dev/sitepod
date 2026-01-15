---
title: sitepod bind
description: Upgrade an anonymous account by binding an email
---

The `bind` command upgrades an anonymous account to a permanent account.

## Usage

```bash
sitepod bind
```

## When to use

If you started with an anonymous account:

```bash
$ sitepod login
? Login method: Anonymous (quick start, 24h limit)
◐ Creating anonymous session
✓ Anonymous session
  expires: 24h
```

Use `bind` to make it permanent:

```bash
$ sitepod bind
? Email: you@example.com
◐ Sending verification email
✓ Email sent

Next:
  - Check your inbox
  - Click the verification link
  - Account upgraded
```

## What gets preserved

When you bind an email:
- All existing deployments remain
- Project configurations are kept
- Subdomains are unchanged
- History is preserved

## Verification

Similar to login, binding requires email verification:

1. Enter your email address
2. Check inbox for verification link
3. Click link to complete binding

## See also

- [sitepod login](/docs/cli/login/) - Initial authentication
- [sitepod deploy](/docs/cli/deploy/) - Deploy your site
