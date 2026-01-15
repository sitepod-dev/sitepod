---
title: sitepod history
description: View deployment history
---

The `history` command shows your deployment history.

## Usage

```bash
sitepod history [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--limit <n>` | Number of entries to show (default: 10) |
| `--env <env>` | Filter by environment |

## Examples

### View recent history

```bash
sitepod history
```

```
Version  Status   Time           Git Commit  Message
v5       current  2 minutes ago  a1b2c3d     feat: add dark mode
v4       -        1 hour ago     e4f5g6h     fix: login bug
v3       -        yesterday      i7j8k9l     update styles
v2       -        2 days ago     m2n3o4p     add contact page
v1       -        1 week ago     q5r6s7t     initial release
```

### Show more entries

```bash
sitepod history --limit 20
```

### Filter by environment

```bash
sitepod history --env beta
```

## Output columns

| Column | Description |
|--------|-------------|
| Version | Deployment version number |
| Status | `current` if this is the active deployment |
| Time | When the deployment was created |
| Git Commit | Short commit hash (if deployed from git repo) |
| Message | Deployment message (from `--message` flag) |

## Integration with rollback

Use history to find a version, then rollback:

```bash
# View history
sitepod history

# Rollback to specific version
sitepod rollback --to v3
```

## See also

- [sitepod rollback](/docs/cli/rollback/) - Rollback to previous version
- [sitepod deploy](/docs/cli/deploy/) - Create new deployment
