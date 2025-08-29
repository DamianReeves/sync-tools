# sync-tools — Fast directory sync with Go, Cobra, and Bubble Tea

[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

sync-tools is a powerful, modern Go CLI wrapper around rsync that provides:

## ✨ Features

- **🚀 Fast & Efficient**: Built with Go for high performance and cross-platform support
- **🎯 One-way or two-way** directory synchronization
- **📁 Gitignore-style** `.syncignore` files (source and destination)
- **🔗 Optional import** of `SOURCE/.gitignore` patterns
- **🎨 Interactive Mode**: Beautiful terminal UI with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **📜 SyncFile Format**: Dockerfile-like declarative sync configuration
- **⚡ Per-side ignore** files and inline patterns (with `!` unignore)
- **📋 "Whitelist" mode** to sync only specified paths
- **⚙️ Flexible Configuration**: TOML config files OR pure CLI usage
- **🔍 Smart Defaults**: Excludes `.git/`, optional hidden directory exclusion
- **🎭 Dry-run previews** and detailed change output
- **📊 Multiple Output Formats**: Text, JSON logging, Markdown reports, and git patches
- **🔧 Git Patch Generation**: Create git-format patch files instead of syncing (via --patch flag or --report with .patch/.diff extension)

## 🚀 Quick Start

### Installation

**Download from Releases** (Recommended):
```bash
# Download the latest binary for your platform
curl -L https://github.com/DamianReeves/sync-tools/releases/latest/download/sync-tools-linux-amd64 -o sync-tools
chmod +x sync-tools
sudo mv sync-tools /usr/local/bin/
```

**Build from Source**:
```bash
git clone https://github.com/DamianReeves/sync-tools.git
cd sync-tools
make build
```

### Basic Usage

```bash
# Simple one-way sync
sync-tools sync --source ./project --dest ./backup --dry-run

# Interactive mode with beautiful TUI
sync-tools sync --source ./src --dest ./dst --interactive

# Two-way sync with conflict resolution
sync-tools sync --source ./local --dest ./remote --mode two-way

# Use gitignore patterns and custom filters
sync-tools sync --source ./code --dest ./backup --use-gitignore --ignore-src "*.tmp"

# Whitelist mode (only sync specific files)
sync-tools sync --source ./docs --dest ./backup --only "*.md" --only "*.txt"

# Generate git patch instead of syncing
sync-tools sync --source ./src --dest ./dst --patch changes.patch

# Preview changes with colored diff (with paging, press 'q' to quit)
sync-tools sync --source ./src --dest ./dst --preview

# Generate patch and apply it with confirmation
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch

# Generate patch and auto-apply without confirmation
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch -y
```

## 📜 SyncFile Format

sync-tools introduces a powerful **SyncFile format** — a Dockerfile-inspired declarative syntax for complex sync operations:

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

### Execute SyncFile

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

### SyncFile Instructions

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

## 🎨 Interactive Mode

Launch the beautiful terminal interface:

```bash
sync-tools sync --source ./project --dest ./backup --interactive
```

Features:
- Real-time sync progress
- Visual confirmation before sync
- Elegant UI with styled output
- Easy abort/continue controls

## ⚙️ Configuration

### TOML Configuration File

Create a `sync.toml` or `.sync.toml` file:

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

### Configuration Priority

1. Command-line flags (highest priority)
2. TOML configuration file
3. Default values (lowest priority)

## 🔍 Advanced Usage

### Preview and Interactive Features

```bash
# Preview changes with colored diff (uses less pager, press 'q' to quit)
sync-tools sync --source ./src --dest ./dst --preview

# Preview changes before generating patch
sync-tools sync --source ./src --dest ./dst --preview --patch changes.patch

# Interactive mode with Bubble Tea UI
sync-tools sync --source ./src --dest ./dst --interactive
```

### Logging & Output

```bash
# JSON logging to file
sync-tools sync --source ./src --dest ./dst --log-format json --log-file sync.log

# Verbose output
sync-tools sync --source ./src --dest ./dst -vv

# Generate Markdown report
sync-tools sync --source ./src --dest ./dst --report sync-report.md

# Generate patch report (format auto-detected from extension)
sync-tools sync --source ./src --dest ./dst --report changes.patch

# Dump rsync commands to JSON
sync-tools sync --source ./src --dest ./dst --dump-commands commands.json
```

### Git Patch Generation

Generate git-format patch files for review and manual application instead of automatic syncing:

```bash
# Generate patch file from differences (using --patch flag)
sync-tools sync --source ./src --dest ./dst --patch changes.patch

# Generate patch file from differences (using --report flag with .patch extension)
sync-tools sync --source ./src --dest ./dst --report changes.patch

# Preview what would be patched (dry-run)
sync-tools sync --source ./src --dest ./dst --patch preview.patch --dry-run

# Generate patch with filtering (respects .syncignore)
sync-tools sync --source ./project --dest ./backup --patch filtered.patch --use-source-gitignore

# Patch with whitelist mode (only specific files)
sync-tools sync --source ./docs --dest ./backup --patch docs-only.patch --only "*.md" --only "*.txt"

# Two-way patch generation shows differences in both directions
sync-tools sync --source ./local --dest ./remote --patch bidirectional.patch --mode two-way

# Apply patch with confirmation prompt
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch

# Apply patch automatically without confirmation
sync-tools sync --source ./src --dest ./dst --patch changes.patch --apply-patch -y
```

### Filter Examples

```bash
# Use .gitignore patterns
sync-tools sync --source ./code --dest ./backup --use-gitignore

# Custom source and destination filters
sync-tools sync --source ./src --dest ./dst \
  --ignore-src "*.log" --ignore-src "temp/" \
  --ignore-dest "cache/"

# Whitelist only specific file types
sync-tools sync --source ./docs --dest ./backup \
  --only "*.md" --only "*.txt" --only "images/"

# Unignore specific files in .syncignore
echo "temp/" > .syncignore
echo "!temp/important.txt" >> .syncignore
```

### Patch Generation Behavior

When using `--patch filename.patch`, sync-tools:

- **Creates git-format patches** using `git diff --no-index` for proper formatting
- **Respects all filters** (.syncignore, --only, --ignore-*, --use-source-gitignore)  
- **Shows file additions** (new files in source) as `+++ /dev/null` 
- **Shows file deletions** (files only in destination) as `--- /dev/null`
- **Shows modifications** with unified diff format showing line changes
- **Includes metadata** header with source, destination, and generation timestamp
- **Dry-run mode** shows what would be patched without creating the file
- **No actual syncing** occurs - only the patch file is generated

**Two Ways to Generate Patches:**
1. **Dedicated flag**: Use `--patch filename.patch` for explicit patch generation
2. **Report flag**: Use `--report filename.patch` where format is auto-detected from extension (.patch/.diff)

Generated patches can be applied using standard git tools:
```bash
git apply changes.patch    # Apply the patch
git apply --check changes.patch    # Validate patch without applying
```

## 🛠️ Development

### Building

```bash
# Install dependencies and build
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run all checks (format, vet, lint, test)
make check

# Development build with all checks
make dev
```

### Project Structure

```
sync-tools/
├── cmd/sync-tools/        # Main application entry point
├── internal/
│   ├── cmd/              # Cobra command definitions
│   ├── config/           # TOML configuration handling
│   ├── rsync/            # Rsync wrapper and execution
│   ├── filters/          # Filter file generation
│   └── logging/          # Logging setup and utilities
├── pkg/
│   ├── syncfile/         # SyncFile format parser
│   └── tui/              # Bubble Tea interactive interface
├── go.mod                # Go module definition
└── Makefile.go           # Build system
```

## 📋 Migration from Python Version

sync-tools has been completely rewritten in Go from the original Python implementation. The Go version provides:

- **⚡ Better Performance**: Faster startup and execution
- **📦 Easy Deployment**: Single binary, no dependencies
- **🎨 Enhanced UX**: Interactive mode with Bubble Tea
- **📜 New Features**: SyncFile format for declarative configuration
- **🔧 Better Tooling**: Modern Go build system and toolchain

All original functionality is preserved and enhanced.

## 📋 Patch Recipes

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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [GitHub Repository](https://github.com/DamianReeves/sync-tools)
- [Issue Tracker](https://github.com/DamianReeves/sync-tools/issues)
- [Releases](https://github.com/DamianReeves/sync-tools/releases)

---

**sync-tools**: Powerful, modern directory synchronization with Go, Cobra, and Bubble Tea 🚀