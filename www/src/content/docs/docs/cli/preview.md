---
title: sitepod preview
description: Create a temporary preview deployment
---

The `preview` command creates a temporary deployment for sharing work-in-progress.

## Usage

```bash
sitepod preview [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--slug <name>` | Custom preview identifier |
| `--ttl <duration>` | Time to live (default: 24h) |

## Examples

### Create a preview

```bash
sitepod preview
```

```
Scanning ./dist... 89 files
Uploading... 12/12 (new files only)
✓ Created preview

  https://my-site--abc123.preview.sitepod.dev
  Expires: 24 hours
```

### Custom slug

```bash
sitepod preview --slug pr-123
```

```
✓ Created preview

  https://my-site--pr-123.preview.sitepod.dev
  Expires: 24 hours
```

### Custom expiry

```bash
sitepod preview --ttl 72h
```

## Preview URLs

Preview URLs follow the pattern:

```
https://{project}--{slug}.preview.sitepod.dev
```

Examples:
- `https://my-site--abc123.preview.sitepod.dev`
- `https://my-site--pr-42.preview.sitepod.dev`
- `https://my-site--feat-login.preview.sitepod.dev`

## Use cases

### Code review

Share a preview link in your PR:

```bash
sitepod preview --slug pr-$PR_NUMBER
```

Add the link to your PR description for reviewers.

### Stakeholder review

Share with product/design for feedback:

```bash
sitepod preview --slug demo-v2
```

### A/B testing

Create multiple previews to compare:

```bash
sitepod preview --slug variant-a
sitepod preview --slug variant-b
```

## CI/CD integration

In GitHub Actions:

```yaml
- name: Create preview
  run: |
    sitepod preview --slug pr-${{ github.event.pull_request.number }}
  env:
    SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}

- name: Comment on PR
  uses: actions/github-script@v6
  with:
    script: |
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: 'Preview: https://my-site--pr-${{ github.event.pull_request.number }}.preview.sitepod.dev'
      })
```

## Automatic cleanup

Previews are automatically deleted after expiry (default 24h).

The preview's files may be retained if they're shared with other deployments (content-addressed storage), but the preview URL will stop working.

## See also

- [sitepod deploy](/docs/cli/deploy/) - Deploy to beta/prod
- [CI/CD Integration](/docs/guides/ci-cd/) - Automate previews
