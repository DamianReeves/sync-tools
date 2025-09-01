# sync-tools

[![Go Report Card](https://goreportcard.com/badge/github.com/DamianReeves/sync-tools)](https://goreportcard.com/report/github.com/DamianReeves/sync-tools)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/release/DamianReeves/sync-tools.svg)](https://github.com/DamianReeves/sync-tools/releases/latest)

A powerful Go CLI wrapper around rsync that provides fast directory synchronization with advanced filtering capabilities, SyncFile post-sync actions, and sophisticated conflict resolution.

## Features

- **Fast Directory Synchronization**: Built on rsync for efficient file transfers
- **Advanced Filtering**: Support for `.syncignore` files, `.gitignore` import, and CLI overrides  
- **Whitelist Mode**: Explicit path inclusion with "only" patterns for precise sync control
- **Two-Phase Interactive Sync**: Generate and review sync plans before execution
- **SyncFile Post-Sync Actions**: APPEND and PREPEND operations for automated file modifications
- **Conflict Resolution**: Multiple strategies including newest-wins, oldest-wins, and interactive
- **Git Patch Generation**: Create patch files for version control integration
- **Cross-Platform**: Supports Linux, macOS, and Windows
- **Single Binary**: No external dependencies except rsync

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap DamianReeves/tap
brew install sync-tools
```

### Download Binaries

Download the latest release from the [releases page](https://github.com/DamianReeves/sync-tools/releases).

### Build from Source

```bash
git clone https://github.com/DamianReeves/sync-tools.git
cd sync-tools
go build -o sync-tools ./cmd/sync-tools
```

## Quick Start

### Basic Synchronization

```bash
# One-way sync (source → destination)
sync-tools sync --source ./my-project --dest ./backup

# Dry-run to preview changes
sync-tools sync --source ./my-project --dest ./backup --dry-run

# Two-way sync with conflict detection
sync-tools sync --source ./local --dest ./remote --mode two-way
```

### Interactive Sync with Plans

```bash
# Generate a sync plan for review
sync-tools sync --source ./src --dest ./dst --plan sync.plan

# Review and edit the plan file, then apply it
sync-tools sync --apply-plan sync.plan
```

### Advanced Filtering

```bash
# Use .syncignore file
echo "*.tmp" > .syncignore
echo "node_modules/" >> .syncignore
sync-tools sync --source . --dest ../backup

# Import .gitignore patterns  
sync-tools sync --source . --dest ../backup --gitignore

# Whitelist mode - only sync specific patterns
sync-tools sync --source . --dest ../backup --only "*.go" --only "*.md"
```

### SyncFile Operations

Create a `MySyncFile` configuration:

```
# MySyncFile - Sync configuration
VAR SOURCE=./my-project
VAR DEST=./backup

SYNC ${SOURCE} ${DEST}
MODE two-way
GITIGNORE true
EXCLUDE *.tmp
EXCLUDE build/

# Post-sync actions
APPEND README.md "Last synced: $(date)"
PREPEND CHANGELOG.md "## Latest Changes\n"
```

Execute the SyncFile:

```bash
sync-tools syncfile MySyncFile
```

## Configuration

### .syncignore Format

```gitignore
# Comments start with #
*.tmp
*.log
build/
node_modules/
.git/

# Negate patterns with !
!important.log
```

### SyncFile Syntax

```
# Variables
VAR NAME=value

# Sync operations  
SYNC source destination
MODE one-way|two-way
GITIGNORE true|false
EXCLUDE pattern
ONLY pattern
DRYRUN true|false

# Post-sync actions
APPEND file content
PREPEND file content
```

## Command Reference

### Main Commands

- `sync-tools sync` - Synchronize directories
- `sync-tools syncfile` - Execute SyncFile configuration
- `sync-tools version` - Show version information

### Sync Options

- `--source, -s` - Source directory path
- `--dest, -d` - Destination directory path  
- `--mode` - Sync mode: one-way (default) or two-way
- `--dry-run` - Preview changes without executing
- `--interactive, -i` - Interactive mode with confirmations
- `--gitignore` - Import .gitignore patterns
- `--exclude` - Exclude patterns (can be used multiple times)
- `--only` - Whitelist patterns (can be used multiple times)

### Plan Operations

- `--plan FILE` - Generate sync plan to file
- `--apply-plan FILE` - Execute sync plan from file
- `--include-changes TYPE` - Filter plan by change types
- `--exclude-changes TYPE` - Exclude change types from plan

### Patch Generation

- `--patch FILE` - Generate git patch instead of syncing
- `--apply-patch` - Apply patch after generation

## Development

### Prerequisites

- Go 1.21 or later
- rsync installed on the system
- golangci-lint for code quality

### Building

```bash
# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Run BDD tests
make test-bdd

# Run all checks
make check
```

### Testing

The project uses BDD (Behavior-Driven Development) with Gherkin scenarios:

```bash
# Run all tests including BDD
make test-with-bdd

# Run only BDD tests
make test-bdd

# View test coverage
make test-coverage
```

Current test coverage: **57/57 BDD scenarios passing (100%)**

### Release Process

```bash
# Create and push a release tag
make tag-release VERSION=v0.4.1

# Test release locally
make release-dry
```

## Architecture

sync-tools is built with a modular architecture focusing on:

- **Testability**: All components designed with testing as a first-class concern
- **Composition**: Flexible, reusable components over complex hierarchies  
- **BDD/TDD Discipline**: Comprehensive test coverage with executable specifications
- **Integration Boundaries**: Clear separation between core logic and external systems

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write BDD scenarios for new functionality
4. Implement code following TDD discipline
5. Ensure all tests pass (`make check`)
6. Commit changes (`git commit -m 'Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines and [DEVELOPMENT-TRACKER.md](DEVELOPMENT-TRACKER.md) for current project status.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on top of the powerful [rsync](https://rsync.samba.org/) utility
- BDD testing powered by [Godog](https://github.com/cucumber/godog)
- Release automation with [GoReleaser](https://goreleaser.com/)

---

**Target Users:**
- **DevOps Engineers**: Automated deployment and backup workflows
- **Developers**: Multi-environment file synchronization  
- **System Administrators**: Large-scale directory management
- **Data Managers**: Precise inclusion/exclusion control for data sync