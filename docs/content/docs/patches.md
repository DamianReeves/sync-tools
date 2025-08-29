---
title: "Git Patch Features"
linkTitle: "Patches"
weight: 3
description: >
  Discover patch generation, preview, and application workflows.
---

# Git Patch Features

sync-tools provides comprehensive git patch support for review and manual application workflows instead of automatic synchronization.

## Patch Generation

Generate git-format patch files for review and manual application:

### Basic Patch Generation

```bash
# Generate patch file from differences (using --patch flag)
sync-tools sync --source ./src --dest ./dst --patch changes.patch

# Generate patch file from differences (using --report flag with .patch extension)
sync-tools sync --source ./src --dest ./dst --report changes.patch

# Preview what would be patched (dry-run)
sync-tools sync --source ./src --dest ./dst --patch preview.patch --dry-run
```

### Two Ways to Generate Patches

1. **Dedicated flag**: Use `--patch filename.patch` for explicit patch generation
2. **Report flag**: Use `--report filename.patch` where format is auto-detected from extension (.patch/.diff)

## Preview Changes

Use the `--preview` flag to see colored diff output before operations:

```bash
# Preview changes with colored diff (uses less pager, press 'q' to quit)
sync-tools sync --source ./src --dest ./dst --preview

# Preview changes before generating patch
sync-tools sync --source ./src --dest ./dst --preview --patch changes.patch
```

## Patch Application

### Apply with Confirmation

```bash
# Generate and apply patch with confirmation prompt
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch

# The system will:
# 1. Generate the patch file
# 2. Show a preview of changes
# 3. Prompt for confirmation
# 4. Apply the patch if confirmed
```

### Auto-Apply (No Confirmation)

```bash
# Generate and auto-apply without confirmation (Unix-style -y flag)
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch -y
```

## Patch Generation Behavior

When using patch generation, sync-tools:

- **Creates git-format patches** using `git diff --no-index` for proper formatting
- **Respects all filters** (.syncignore, --only, --ignore-*, --use-source-gitignore)
- **Shows file additions** (new files in source) as `+++ /dev/null`
- **Shows file deletions** (files only in destination) as `--- /dev/null`
- **Shows modifications** with unified diff format showing line changes
- **Includes metadata** header with source, destination, and generation timestamp
- **Dry-run mode** shows what would be patched without creating the file
- **No actual syncing** occurs - only the patch file is generated

## Applying Generated Patches

Generated patches can be applied using standard git tools:

```bash
git apply changes.patch    # Apply the patch
git apply --check changes.patch    # Validate patch without applying
```

## Advanced Patch Examples

### Filtered Patch Generation

```bash
# Generate patch with filtering (respects .syncignore)
sync-tools sync --source ./project --dest ./backup --patch filtered.patch --use-source-gitignore

# Patch with whitelist mode (only specific files)
sync-tools sync --source ./docs --dest ./backup --patch docs-only.patch --only "*.md" --only "*.txt"
```

### Two-way Patch Generation

```bash
# Two-way patch generation shows differences in both directions
sync-tools sync --source ./local --dest ./remote --patch bidirectional.patch --mode two-way
```

## Patch Recipes

Common use cases for git patch generation:

### Code Review Workflow

```bash
# Generate patch of changes for review (using --patch flag)
sync-tools sync --source ./feature-branch --dest ./main-branch --patch review.patch

# Alternative: using --report with .patch extension
sync-tools sync --source ./feature-branch --dest ./main-branch --report review.patch

git apply review.patch  # Apply the patch for testing
```

### Deployment Preparation

```bash
# Create deployment patch with only production files (using --patch flag)
sync-tools sync --source ./local --dest ./production \
  --patch deployment.patch \
  --only "src/" --only "config/prod.conf" \
  --ignore-src "*.test" --ignore-src "dev/"

# Alternative: using --report with .patch extension
sync-tools sync --source ./local --dest ./production \
  --report deployment.patch \
  --only "src/" --only "config/prod.conf" \
  --ignore-src "*.test" --ignore-src "dev/"
```

### Documentation Updates

```bash
# Generate patch for documentation changes only
sync-tools sync --source ./docs-new --dest ./docs-current \
  --patch docs-update.patch \
  --only "*.md" --only "*.rst" --only "images/"
```

### Selective Updates

```bash
# Preview changes before applying to sensitive environment
sync-tools sync --source ./staging --dest ./production \
  --patch staging-to-prod.patch --dry-run

# Review the patch, then apply manually:
# git apply staging-to-prod.patch
```

### Backup Validation

```bash
# Generate patch to see what would be backed up
sync-tools sync --source ./project --dest ./backup \
  --patch backup-preview.patch \
  --use-source-gitignore --exclude-hidden-dirs
```

## SyncFile Patch Support

All patch features are available in SyncFile format:

```dockerfile
# Example SyncFile with patch workflow
VAR PROJECT=/home/user/project

# Preview changes first
SYNC ${PROJECT}/src ./backup/src
PREVIEW true
MODE one-way

# Generate and apply patch with confirmation
SYNC ${PROJECT}/config ./backup/config
PATCH config-changes.patch
APPLYPATCH true
AUTOCONFIRM false  # Will prompt user

# Auto-apply critical updates without confirmation
SYNC ${PROJECT}/critical ./backup/critical
PATCH critical-updates.patch
APPLYPATCH true
AUTOCONFIRM true  # No prompts, auto-apply
```

See the [SyncFile documentation]({{< relref "/docs/syncfile" >}}) for more details on declarative patch workflows.