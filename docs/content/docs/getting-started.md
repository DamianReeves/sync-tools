---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 1
description: >
  Install and use sync-tools with basic examples.
---

# Getting Started with sync-tools

This guide will help you install sync-tools and get started with basic synchronization operations.

## Installation

### Download from Releases (Recommended)

```bash
# Download the latest binary for your platform
curl -L https://github.com/DamianReeves/sync-tools/releases/latest/download/sync-tools-linux-amd64 -o sync-tools
chmod +x sync-tools
sudo mv sync-tools /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/DamianReeves/sync-tools.git
cd sync-tools
make build
```

## Basic Usage

### Simple One-way Sync

```bash
# Basic one-way sync with dry-run
sync-tools sync --source ./project --dest ./backup --dry-run

# Actual sync (removes --dry-run)
sync-tools sync --source ./project --dest ./backup
```

### Interactive Mode

Launch the beautiful terminal interface:

```bash
sync-tools sync --source ./project --dest ./backup --interactive
```

### Two-way Sync

```bash
# Two-way sync with conflict detection
sync-tools sync --source ./local --dest ./remote --mode two-way
```

## Preview Changes

Use the `--preview` flag to see what changes will be made:

```bash
# Preview changes with colored diff (uses less pager, press 'q' to quit)
sync-tools sync --source ./src --dest ./dst --preview
```

## Filtering and Patterns

### Using .gitignore patterns

```bash
# Use source .gitignore patterns
sync-tools sync --source ./code --dest ./backup --use-source-gitignore
```

### Custom ignore patterns

```bash
# Custom source and destination filters
sync-tools sync --source ./src --dest ./dst \
  --ignore-src "*.log" --ignore-src "temp/" \
  --ignore-dest "cache/"
```

### Whitelist mode

```bash
# Only sync specific file types
sync-tools sync --source ./docs --dest ./backup \
  --only "*.md" --only "*.txt" --only "images/"
```

## Git Patch Generation

Generate git-format patch files instead of syncing:

```bash
# Generate patch file from differences
sync-tools sync --source ./src --dest ./dst --patch changes.patch

# Preview what would be patched (dry-run)
sync-tools sync --source ./src --dest ./dst --patch preview.patch --dry-run
```

### Apply Patches

```bash
# Generate and apply patch with confirmation
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch

# Generate and auto-apply without confirmation
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch -y
```

## Configuration Files

Create a `sync.toml` or `.sync.toml` file for default settings:

```toml
source = "./project"
dest = "./backup"
mode = "one-way"
dry_run = true
use_source_gitignore = true
exclude_hidden_dirs = false
ignore_src = ["*.tmp", "node_modules/"]
ignore_dest = []
only = []
log_level = "INFO"
log_format = "text"
```

## Next Steps

- Learn about the [SyncFile format]({{< relref "/docs/syncfile" >}}) for declarative configurations
- Explore [Git Patch features]({{< relref "/docs/patches" >}}) for advanced workflows
- Check out [Examples]({{< relref "/docs/examples" >}}) for real-world usage scenarios