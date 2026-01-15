---
title: sitepod delete-account
description: Delete your SitePod account and all projects
---

The `delete-account` command permanently deletes your account and all associated projects.

## Usage

```bash
sitepod delete-account [options]
```

## Options

| Option | Description |
|--------|-------------|
| `-y, --yes` | Skip confirmation prompt |

## Examples

### Interactive deletion

```bash
$ sitepod delete-account
⚠ This will delete your account and all projects

? Delete account? (y/N) y

◐ Deleting account
✓ Account deleted
  projects: 3
  config: cleared
```

### CI/CD or scripted deletion

```bash
sitepod delete-account --yes
```

## What happens

1. **Confirmation**: Prompts for confirmation (unless `--yes` is used)
2. **Server deletion**: Removes account and all projects from server
3. **Local cleanup**: Removes saved credentials from `~/.sitepod/config.toml`

## Warning

This action is **irreversible**. All your projects and deployments will be permanently deleted.

## See also

- [sitepod login](/docs/cli/login/) - Create a new account
