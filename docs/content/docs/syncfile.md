---
title: "SyncFile Format"
linkTitle: "SyncFile"
weight: 2
description: >
  Master the powerful SyncFile format for declarative sync operations.
---

# SyncFile Format

sync-tools introduces a powerful **SyncFile format** — a Dockerfile-inspired declarative syntax for complex sync operations.

## Basic Syntax

A SyncFile uses simple, declarative syntax to define multiple sync operations, filters, and configurations in a single file.

### Example SyncFile

```dockerfile
# Multi-project sync configuration with patch support
VAR PROJECT_ROOT=/home/user/projects
VAR BACKUP_ROOT=/backup

# Preview documentation changes before syncing
SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs
MODE one-way
PREVIEW true
EXCLUDE *.tmp
EXCLUDE .DS_Store
INCLUDE !important.tmp

# Generate and apply patch for source code with confirmation
SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
MODE two-way
PATCH src-changes.patch
APPLYPATCH true
AUTOCONFIRM false  # Will prompt for confirmation
GITIGNORE true
HIDDENDIRS exclude
ONLY *.go
ONLY *.py
ONLY *.js

# Auto-apply configuration patches without confirmation
SYNC ${PROJECT_ROOT}/config ${BACKUP_ROOT}/config
PATCH config-update.patch
APPLYPATCH true
AUTOCONFIRM true  # Like -y flag, no prompts
EXCLUDE secrets/
INCLUDE !config/main.conf
```

## Instructions Reference

| Instruction | Description | Example |
|-------------|-------------|---------|
| `SYNC source dest [options]` | Define a sync operation | `SYNC ./src ./backup --dry-run` |
| `MODE one-way\|two-way` | Set sync mode | `MODE two-way` |
| `EXCLUDE pattern` | Exclude files/folders | `EXCLUDE *.tmp` |
| `INCLUDE pattern` | Include files (unignore) | `INCLUDE !important.tmp` |
| `ONLY pattern` | Whitelist mode | `ONLY *.go` |
| `DRYRUN true\|false` | Enable/disable dry run | `DRYRUN true` |
| `PATCH filename` | Generate git patch file | `PATCH changes.patch` |
| `APPLYPATCH true\|false` | Apply patch after creation | `APPLYPATCH true` |
| `PREVIEW true\|false` | Show colored diff preview | `PREVIEW true` |
| `AUTOCONFIRM true\|false` | Auto-confirm patch application | `AUTOCONFIRM true` |
| `GITIGNORE true\|false` | Use .gitignore patterns | `GITIGNORE true` |
| `HIDDENDIRS exclude\|include` | Handle hidden directories | `HIDDENDIRS exclude` |
| `VAR name=value` | Define variable | `VAR BASE=/home/user` |
| `ENV name=value` | Environment variable | `ENV RSYNC_OPTS=--progress` |
| `# comment` | Comments | `# Sync documentation` |

Variables can be referenced using `${name}` or `$name` syntax.

## Execution

### Execute Default SyncFile

```bash
# Execute default SyncFile
sync-tools syncfile

# Execute specific SyncFile
sync-tools syncfile my-sync.sf

# List operations without executing
sync-tools syncfile --list

# Override to dry-run mode
sync-tools syncfile --dry-run
```

## Advanced Examples

### Multi-Environment Sync

```dockerfile
# Development environment sync
VAR DEV_ROOT=/home/user/dev
VAR STAGING_ROOT=/mnt/staging
VAR PROD_ROOT=/mnt/production

# Preview development to staging
SYNC ${DEV_ROOT} ${STAGING_ROOT}
MODE one-way
PREVIEW true
GITIGNORE true
EXCLUDE *.log
EXCLUDE node_modules/

# Generate patch for staging to production
SYNC ${STAGING_ROOT} ${PROD_ROOT}
PATCH staging-to-prod.patch
APPLYPATCH false  # Manual review required
ONLY src/
ONLY config/production.conf
```

### Workflow with Multiple Patches

```dockerfile
# Database sync with backup
SYNC ./database/current ./database/backup
PATCH db-backup.patch
APPLYPATCH true
AUTOCONFIRM true  # Auto-backup

# Code review preparation
SYNC ./feature-branch ./main-branch
PATCH code-review.patch
APPLYPATCH false  # Generate for review only
GITIGNORE true

# Documentation updates
SYNC ./docs-new ./docs-current
PATCH docs-update.patch
APPLYPATCH true
AUTOCONFIRM false  # Confirm doc changes
ONLY *.md
ONLY *.rst
ONLY images/
```

## Configuration Priority

When using SyncFiles with CLI flags, the priority order is:

1. Command-line flags (highest priority)
2. SyncFile instructions
3. TOML configuration file
4. Default values (lowest priority)

## SyncFile Discovery

sync-tools looks for SyncFiles in this order:

1. Specified file: `sync-tools syncfile my-file.sf`
2. `SyncFile`
3. `Syncfile`
4. `syncfile`
5. `.syncfile`