---
title: Rollback
description: Instantly revert to a previous deployment
---

One of SitePod's key features is instant rollback to any previous deployment.

## How rollback works

SitePod stores each deployment as an immutable snapshot (Pod). Environments are pointers to these snapshots.

```
Before: prod → Pod v3
After:  prod → Pod v2  (instant!)
```

Rollback just moves the pointer. No file copying, no rebuild.

## Interactive rollback

```bash
sitepod rollback
```

```
? Select version to rollback to:
  > v3 (current) - 2 minutes ago - add dark mode
    v2 - 1 hour ago - fix login bug
    v1 - yesterday - initial release

? Confirm rollback to v2? Yes
✓ Rolled back to v2

  https://my-site.sitepod.dev
  Rollback time: 0.3s
```

## Direct rollback

Skip the prompt:

```bash
sitepod rollback --to v2
```

## Rollback beta environment

```bash
sitepod rollback --env beta --to v1
```

## View version history

```bash
sitepod history
```

```
Version  Status   Time           Git Commit  Message
v5       current  2 minutes ago  a1b2c3d     add dark mode
v4       -        1 hour ago     e4f5g6h     fix login bug
v3       -        yesterday      i7j8k9l     update styles
v2       -        2 days ago     m2n3o4p     add contact page
v1       -        1 week ago     q5r6s7t     initial release
```

## When to rollback

Common scenarios:

### Bad deployment
```bash
# Oops, broke production
sitepod rollback --to v4
# Fixed in 0.3 seconds
```

### Testing previous version
```bash
# Need to reproduce a bug from v2
sitepod rollback --to v2

# Done testing, back to current
sitepod rollback --to v5
```

### Emergency recovery
```bash
# 3 AM incident, roll back immediately
sitepod rollback --to v3
# Investigate tomorrow
```

## Rollback vs. redeploy

| Action | What happens | When to use |
|--------|-------------|-------------|
| Rollback | Move pointer to existing Pod | Reverting, emergency |
| Redeploy | Create new Pod from files | New changes |

Rollback is **instant** because no files are copied.

## Roll forward

After a rollback, you can:

1. **Rollback again** to a newer version
2. **Deploy** to create a new version

```bash
# Rolled back to v2, now want v4
sitepod rollback --to v4

# Or deploy new changes
sitepod deploy --prod
```

## Retention

By default, SitePod keeps:
- All versions referenced by environments
- Recent versions based on garbage collection settings

Configure retention in server settings or via `SITEPOD_GC_*` environment variables.

## Best practices

1. **Add deployment messages** for easier identification:
   ```bash
   sitepod deploy --prod --message "v1.2.0 release"
   ```

2. **Test in beta first**:
   ```bash
   sitepod deploy        # beta
   # verify
   sitepod deploy --prod # production
   ```

3. **Know your versions**: Run `sitepod history` before rolling back

## See also

- [sitepod rollback](/docs/cli/rollback/) - CLI reference
- [sitepod history](/docs/cli/history/) - View deployment history
- [Core Concepts](/docs/concepts/) - Understanding Pods and refs
