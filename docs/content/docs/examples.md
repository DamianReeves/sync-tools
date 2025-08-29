---
title: "Examples"
linkTitle: "Examples"
weight: 4
description: >
  Real-world usage scenarios and recipes for sync-tools.
---

# Examples and Recipes

This page contains real-world examples and recipes for common sync-tools usage scenarios.

## Basic Synchronization

### Project Backup

```bash
# Simple backup with preview
sync-tools sync --source ./my-project --dest ./backup/my-project --preview

# Actual backup excluding temporary files
sync-tools sync --source ./my-project --dest ./backup/my-project \
  --ignore-src "*.tmp" --ignore-src "node_modules/" \
  --use-source-gitignore
```

### Configuration Sync

```bash
# Sync only configuration files
sync-tools sync --source ./app --dest ./app-backup \
  --only "*.conf" --only "*.yaml" --only "*.json" \
  --exclude "secrets.json"
```

## Development Workflows

### Development to Staging

```bash
# Preview changes before deploying
sync-tools sync --source ./dev-app --dest ./staging-app --preview

# Deploy with filtering
sync-tools sync --source ./dev-app --dest ./staging-app \
  --only "src/" --only "public/" --only "config/staging.conf" \
  --ignore-src "*.log" --ignore-src "debug/"
```

### Code Review Preparation

```bash
# Generate patch for code review
sync-tools sync --source ./feature-branch --dest ./main-branch \
  --patch code-review.patch \
  --use-source-gitignore

# Apply patch for testing (manual step)
git apply code-review.patch
```

## Patch Workflows

### Safe Deployment with Patches

```bash
# Generate deployment patch with preview
sync-tools sync --source ./staging --dest ./production \
  --patch deployment.patch --preview \
  --only "app/" --only "config/prod.conf"

# Review the patch file manually, then apply
sync-tools sync --source ./staging --dest ./production \
  --patch deployment.patch --apply-patch
```

### Automated Patch Application

```bash
# Auto-apply patches in trusted environments
sync-tools sync --source ./updates --dest ./current \
  --patch auto-update.patch --apply-patch -y \
  --ignore-src "*.backup"
```

## SyncFile Examples

### Multi-Environment SyncFile

```dockerfile
# multi-env.sf - Development to Production Pipeline
VAR APP_NAME=myapp
VAR DEV_ROOT=/home/dev/${APP_NAME}
VAR STAGING_ROOT=/opt/staging/${APP_NAME}
VAR PROD_ROOT=/opt/production/${APP_NAME}

# Step 1: Dev to Staging with preview
SYNC ${DEV_ROOT} ${STAGING_ROOT}
MODE one-way
PREVIEW true
GITIGNORE true
EXCLUDE *.log
EXCLUDE debug/
EXCLUDE node_modules/

# Step 2: Staging to Production with patch review
SYNC ${STAGING_ROOT} ${PROD_ROOT}
PATCH staging-to-prod.patch
APPLYPATCH true
AUTOCONFIRM false  # Require manual confirmation
ONLY src/
ONLY public/
ONLY config/production.conf
```

Execute with:
```bash
sync-tools syncfile multi-env.sf
```

### Backup Strategy SyncFile

```dockerfile
# backup-strategy.sf - Comprehensive backup with patches
VAR HOME_DIR=/home/user
VAR BACKUP_ROOT=/mnt/backup

# Full project backup with exclusions
SYNC ${HOME_DIR}/projects ${BACKUP_ROOT}/projects
MODE one-way
GITIGNORE true
EXCLUDE node_modules/
EXCLUDE .git/
EXCLUDE *.log

# Critical config backup with patch tracking
SYNC ${HOME_DIR}/.config ${BACKUP_ROOT}/config
PATCH config-backup.patch
APPLYPATCH true
AUTOCONFIRM true  # Auto-backup configs
EXCLUDE cache/
INCLUDE !important-config.yaml

# Documents with preview
SYNC ${HOME_DIR}/Documents ${BACKUP_ROOT}/Documents
PREVIEW true
ONLY *.pdf
ONLY *.doc*
ONLY *.txt
```

### Development Workflow SyncFile

```dockerfile
# dev-workflow.sf - Complete development sync workflow
VAR PROJECT_ROOT=/workspace/myapp

# Database sync with backup
SYNC ${PROJECT_ROOT}/database/current ${PROJECT_ROOT}/database/backup
PATCH db-backup.patch
APPLYPATCH true
AUTOCONFIRM true
EXCLUDE logs/

# Frontend assets with preview
SYNC ${PROJECT_ROOT}/frontend/dist ${PROJECT_ROOT}/public/assets
PREVIEW true
MODE one-way
ONLY *.js
ONLY *.css
ONLY *.woff*
ONLY images/

# Configuration sync with manual confirmation
SYNC ${PROJECT_ROOT}/config/templates ${PROJECT_ROOT}/config/active
PATCH config-update.patch
APPLYPATCH true
AUTOCONFIRM false  # Confirm config changes
EXCLUDE *.example
INCLUDE !production.conf.example
```

## Advanced Filtering

### Complex Ignore Patterns

```bash
# Multiple source and destination filters
sync-tools sync --source ./complex-project --dest ./backup \
  --ignore-src "*.tmp" --ignore-src "node_modules/" --ignore-src "*.log" \
  --ignore-dest "old-*" --ignore-dest "cache/" \
  --use-source-gitignore
```

### Whitelist with Exceptions

```bash
# Sync only specific files but exclude some patterns
sync-tools sync --source ./documents --dest ./public-docs \
  --only "*.md" --only "*.pdf" --only "images/" \
  --ignore-src "internal-*.md" --ignore-src "draft-*.pdf"
```

## Two-Way Sync Examples

### Conflict Resolution

```bash
# Two-way sync with conflict preservation
sync-tools sync --source ./local --dest ./remote --mode two-way
# Conflicts are preserved as .conflict-timestamp files
```

### Bidirectional Patch Generation

```bash
# Generate patch showing differences in both directions
sync-tools sync --source ./local --dest ./remote \
  --patch bidirectional.patch --mode two-way
```

## Interactive Mode Examples

### Safe Interactive Sync

```bash
# Use interactive mode for careful synchronization
sync-tools sync --source ./important-data --dest ./backup \
  --interactive --use-source-gitignore
```

## Logging and Reports

### Comprehensive Logging

```bash
# JSON logging with report generation
sync-tools sync --source ./app --dest ./backup \
  --log-format json --log-file sync.log \
  --report sync-report.md \
  -vv  # Verbose output
```

### Audit Trail

```bash
# Generate patch for audit trail
sync-tools sync --source ./production --dest ./archive \
  --patch audit-$(date +%Y%m%d).patch \
  --use-source-gitignore
```

## Testing and Validation

### Dry-Run Everything

```bash
# Preview all changes before any operations
sync-tools sync --source ./staging --dest ./production \
  --dry-run --preview --patch test.patch
```

### Validate Before Sync

```bash
# Check what would be synced
sync-tools sync --source ./source --dest ./dest --dry-run

# Then actually sync
sync-tools sync --source ./source --dest ./dest
```