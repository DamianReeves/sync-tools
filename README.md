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
- **📊 Multiple Output Formats**: Text, JSON logging, and Markdown reports

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
```

## 📜 SyncFile Format

sync-tools introduces a powerful **SyncFile format** — a Dockerfile-inspired declarative syntax for complex sync operations:

### Example SyncFile

```dockerfile
# Multi-project sync configuration
VAR PROJECT_ROOT=/home/user/projects
VAR BACKUP_ROOT=/backup

# Sync documentation
SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs
MODE one-way
EXCLUDE *.tmp
EXCLUDE .DS_Store
INCLUDE !important.tmp

# Sync source code with two-way sync
SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
MODE two-way
GITIGNORE true
HIDDENDIRS exclude
ONLY *.go
ONLY *.py
ONLY *.js

# Sync configuration files
SYNC ${PROJECT_ROOT}/config ${BACKUP_ROOT}/config
DRYRUN false
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

### Logging & Output

```bash
# JSON logging to file
sync-tools sync --source ./src --dest ./dst --log-format json --log-file sync.log

# Verbose output
sync-tools sync --source ./src --dest ./dst -vv

# Generate Markdown report
sync-tools sync --source ./src --dest ./dst --report sync-report.md

# Dump rsync commands to JSON
sync-tools sync --source ./src --dest ./dst --dump-commands commands.json
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