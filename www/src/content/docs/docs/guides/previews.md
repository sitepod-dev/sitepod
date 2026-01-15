---
title: Preview Deployments
description: Share work-in-progress with temporary preview links
---

Preview deployments let you share work-in-progress before going to production.

## Create a preview

```bash
sitepod preview
```

```
Scanning ./dist... 89 files
Uploading... 12/12 (new files)
✓ Created preview

  https://my-site--abc123.preview.sitepod.dev
  Expires: 24 hours
```

## Custom slug

Name your preview for easy identification:

```bash
sitepod preview --slug feature-login
```

```
✓ Created preview

  https://my-site--feature-login.preview.sitepod.dev
  Expires: 24 hours
```

## Custom expiry

Change the default 24-hour expiry:

```bash
sitepod preview --ttl 72h    # 3 days
sitepod preview --ttl 7d     # 1 week
```

## URL patterns

Preview URLs follow this pattern:

```
https://{project}--{slug}.preview.sitepod.dev
```

Examples:
- `https://my-site--abc123.preview.sitepod.dev`
- `https://my-site--pr-42.preview.sitepod.dev`
- `https://my-site--v2-redesign.preview.sitepod.dev`

## Use cases

### Pull request reviews

Create a preview for each PR:

```bash
sitepod preview --slug pr-$PR_NUMBER
```

Reviewers can see changes without checking out the branch.

### Stakeholder demos

Share with product managers or clients:

```bash
sitepod preview --slug demo-q4 --ttl 7d
```

A dedicated URL they can bookmark.

### A/B comparisons

Compare two versions:

```bash
# Build variant A
npm run build
sitepod preview --slug variant-a

# Build variant B (different config)
npm run build
sitepod preview --slug variant-b
```

## CI/CD integration

Automatically create previews on PRs. See [CI/CD Integration](/docs/guides/ci-cd/) for full examples.

Quick GitHub Actions snippet:

```yaml
- name: Create preview
  run: |
    sitepod preview --slug pr-${{ github.event.pull_request.number }}
  env:
    SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}
```

## Automatic cleanup

Previews are automatically deleted after expiry:

- Default: 24 hours
- Files may persist if shared with other deployments (deduplication)
- Preview URL stops working at expiry

## Limitations

- Preview URLs are public (no authentication)
- Maximum TTL depends on server configuration
- Counts toward storage quota

## vs. Beta environment

| Feature | Preview | Beta |
|---------|---------|------|
| Purpose | Temporary sharing | Ongoing staging |
| Expiry | Yes (24h default) | No |
| URL | `--slug.preview.` | `.beta.` |
| History | Not tracked | Full history |

Use **previews** for temporary shares (PRs, demos).
Use **beta** for ongoing staging environment.

## See also

- [sitepod preview](/docs/cli/preview/) - CLI reference
- [CI/CD Integration](/docs/guides/ci-cd/) - Automate previews
