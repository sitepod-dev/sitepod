---
title: sitepod rollback
description: Rollback to a previous deployment
---

The `rollback` command switches your site to a previous deployment version.

## Usage

```bash
sitepod rollback [options]
```

## Options

| Option | Description |
|--------|-------------|
| `--to <version>` | Target version (e.g., v2) |
| `--env <env>` | Environment to rollback (default: prod) |

## Examples

### Interactive rollback

```bash
sitepod rollback
```

```
? Select version to rollback to:
  > v3 (current) - 2 minutes ago - feat: add dark mode
    v2 - 1 hour ago - fix: login bug
    v1 - yesterday - initial release

? Confirm rollback to v2? Yes
✓ Rolled back to v2

  https://my-site.sitepod.dev
  Rollback time: 0.3s
```

### Direct rollback

```bash
sitepod rollback --to v2
```

### Rollback beta environment

```bash
sitepod rollback --env beta --to v1
```

## How it works

Rollback updates the environment pointer to a previous Pod:

```
Before: prod → Pod v3
After:  prod → Pod v2
```

Key characteristics:
- **Instant**: Just a pointer update, no file copying
- **Atomic**: Site switches version in one operation
- **Reversible**: Roll forward by deploying or rollback again

## Version history

View available versions with `sitepod history`:

```bash
sitepod history
```

```
Version  Status   Time           Git Commit  Message
v3       current  2 minutes ago  a1b2c3d     feat: add dark mode
v2       -        1 hour ago     e4f5g6h     fix: login bug
v1       -        yesterday      i7j8k9l     initial release
```

## Rollback vs redeploy

| Action | What happens | Speed |
|--------|-------------|-------|
| Rollback | Move pointer to existing Pod | Instant |
| Redeploy | Upload files, create new Pod | Depends on changes |

Use rollback when:
- Reverting a bad deployment
- Testing a previous version
- Emergency recovery

Use redeploy when:
- You need to make new changes
- The previous version needs modifications

## See also

- [sitepod history](/docs/cli/history/) - View deployment history
- [sitepod deploy](/docs/cli/deploy/) - Create new deployment
