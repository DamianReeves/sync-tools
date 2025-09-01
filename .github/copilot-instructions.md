# sync-tools Development Instructions

**Always follow these instructions first and reference [CLAUDE.md](../CLAUDE.md) for comprehensive code guidelines and architectural principles.**

sync-tools is a **Go CLI wrapper** around rsync that provides advanced directory synchronization with .syncignore files, whitelist mode, filter layering, one-way/two-way sync, conflict preservation, and SyncFile post-sync actions. The project has been fully migrated from Python to Go for better performance and single-binary distribution.

## Core Guidelines Reference

**Primary Guidelines**: See [CLAUDE.md](../CLAUDE.md) for comprehensive code guidelines including:
- BDD/TDD mandatory discipline (Red/Green/Refactor cycle)
- Architecture preferences (composition, testability, modularity)
- Testing approach (Gherkin features first, comprehensive coverage)
- Development tracker maintenance requirements
- Technical preferences and commit standards

## Current Architecture (Go Implementation)

### Project Structure
```
sync-tools/
├── cmd/sync-tools/         # Main CLI application entry point
├── internal/
│   ├── cmd/               # Cobra CLI commands (sync, syncfile)
│   └── rsync/             # Core rsync wrapper and sync logic
├── pkg/syncfile/          # SyncFile parsing and post-sync actions
├── features/              # BDD test scenarios (Godog/Gherkin)
├── test/bdd/              # BDD step definitions and test context
├── docs/                  # Documentation and PRDs
├── CLAUDE.md              # Primary code guidelines and standards
├── DEVELOPMENT-TRACKER.md # Mandatory progress tracking
└── go.mod                 # Go module dependencies
```

### Development Commands (Go)
```bash
# Build application
go build -o sync-tools ./cmd/sync-tools

# Run BDD tests (primary test suite)
go test ./test/bdd -tags bdd -v

# Run all tests
go test ./...

# Run application
go run ./cmd/sync-tools [command]

# Install locally
go install ./cmd/sync-tools

# Lint code
golangci-lint run
```

## Working Effectively

### Bootstrap Environment
```bash
# Navigate to repository
cd /path/to/sync-tools

# Verify Go is installed (1.19+ required)
go version

# Verify rsync is available (3.1+ required)
rsync --version

# Build the application
go build -o sync-tools ./cmd/sync-tools

# Verify basic functionality
./sync-tools --help
```

### Testing and Validation (MANDATORY BDD/TDD)

**CRITICAL**: Follow strict BDD/TDD discipline as outlined in [CLAUDE.md](../CLAUDE.md)

```bash
# Run BDD tests (primary integration tests) - MUST pass 100%
go test ./test/bdd -tags bdd -v
# Expected: All scenarios pass, current: 57/57 scenarios

# Run full test suite
go test ./...

# Run specific feature tests
go test ./test/bdd -tags bdd -run 'APPEND'

# BDD Red/Green/Refactor cycle for new features:
# 1. RED: Create .feature file with failing scenarios
# 2. GREEN: Implement minimal code to pass
# 3. REFACTOR: Clean up while maintaining green tests
```

### Running the CLI
```bash
# Method 1: Direct execution after build
./sync-tools sync --help
./sync-tools syncfile --help

# Method 2: Go run (development)
go run ./cmd/sync-tools sync --help
go run ./cmd/sync-tools syncfile --help

# Method 3: After install
sync-tools sync --help
```

### Manual Validation Test
```bash
# Create test directories
mkdir -p /tmp/test_sync/{source,dest}
echo "test content" > /tmp/test_sync/source/file.txt

# Test dry run
./sync-tools sync --source /tmp/test_sync/source --dest /tmp/test_sync/dest --dry-run

# Test actual sync
./sync-tools sync --source /tmp/test_sync/source --dest /tmp/test_sync/dest

# Verify result
ls -la /tmp/test_sync/dest/  # Should contain file.txt
```

## Validation Status

**FULLY VALIDATED** (Go implementation):
- ✅ **BDD Test Suite**: 57 scenarios across multiple features, all passing
- ✅ **Go CLI Framework**: Complete Cobra-based command structure
- ✅ **Core Sync Operations**: One-way/two-way sync with conflict resolution  
- ✅ **Filter System**: .syncignore, .gitignore import, whitelist mode, pattern matching
- ✅ **SyncFile Format**: Dockerfile-like syntax with SYNC, APPEND, PREPEND instructions
- ✅ **Post-Sync Actions**: APPEND and PREPEND actions with comprehensive flag support
- ✅ **Cross-Platform**: Linux, macOS, Windows compatibility
- ✅ **Performance**: Single-binary distribution, efficient rsync integration

**CURRENT FEATURES**:
- **Basic Sync**: One-way and two-way directory synchronization
- **Advanced Filtering**: Multi-layer filter system with sophisticated patterns
- **SyncFile Support**: Declarative sync configurations with post-sync actions
- **Interactive Mode**: Two-phased sync with plan generation and execution
- **Git Integration**: Patch generation and application capabilities
- **Report Generation**: Markdown reports with detailed sync statistics

## Development Workflow (MANDATORY)

**CRITICAL**: Always follow BDD/TDD discipline from [CLAUDE.md](../CLAUDE.md)

### For New Features:
1. **BDD First**: Create `.feature` file with Gherkin scenarios (RED phase)
2. **Run Tests**: Confirm scenarios fail (`go test ./test/bdd -tags bdd`)
3. **Implement**: Write minimal code to make tests pass (GREEN phase)
4. **Refactor**: Clean up code while maintaining green tests
5. **Update Tracker**: Update DEVELOPMENT-TRACKER.md with progress

### Before Committing:
```bash
# MANDATORY: All tests must pass at 100%
go test ./...

# MANDATORY: BDD tests must pass
go test ./test/bdd -tags bdd -v

# MANDATORY: Code must compile
go build -o sync-tools ./cmd/sync-tools

# RECOMMENDED: Run linting
golangci-lint run

# MANDATORY: Update DEVELOPMENT-TRACKER.md
# See [CLAUDE.md](../CLAUDE.md) for tracker maintenance requirements
```

### Development Tracker Maintenance
**MANDATORY**: Update `DEVELOPMENT-TRACKER.md` on every session
- Track progress: In Progress → Pending → Refined → Backlog  
- Document completions with dates and outcomes
- Record architectural decisions and trade-offs
- Maintain accurate current status and priorities
- Reference BDD scenarios in commit messages

## Key Project Components

### Core Architecture
- **CLI Framework**: Cobra-based command structure with sync and syncfile subcommands
- **Rsync Wrapper**: Efficient integration with rsync for file operations
- **Filter Engine**: Sophisticated pattern matching and exclusion/inclusion logic  
- **SyncFile Parser**: Dockerfile-like syntax processor with post-sync actions
- **PostSyncAction Framework**: Pluggable action executors (APPEND, PREPEND, future PATCH)
- **BDD Framework**: Godog integration for executable specifications

### Post-Sync Actions (Current)
- **APPEND**: Add content to end of files with inline/file-based sources
- **PREPEND**: Add content to beginning of files with headers/metadata
- **Planned**: PATCH (git diff application), REPLACE (text substitution), SCRIPT (execution)

### Testing Framework
- **BDD Primary**: Godog with Gherkin scenarios for integration testing
- **Unit Tests**: Standard Go testing for component-level verification
- **Manual Validation**: Real sync operations for end-to-end verification

## Build and Distribution

```bash
# Build single binary
go build -o sync-tools ./cmd/sync-tools

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o sync-tools-linux ./cmd/sync-tools
GOOS=windows GOARCH=amd64 go build -o sync-tools.exe ./cmd/sync-tools
GOOS=darwin GOARCH=amd64 go build -o sync-tools-macos ./cmd/sync-tools

# Install to GOPATH/bin
go install ./cmd/sync-tools
```

## Architectural Principles

**Reference [CLAUDE.md](../CLAUDE.md) for comprehensive guidelines. Key principles:**

1. **BDD/TDD Mandatory**: All features start with failing Gherkin scenarios
2. **Composition Over Inheritance**: Build flexible, reusable components
3. **Testability First**: Design all components with testing as primary concern
4. **Explicit Interfaces**: Value explicit, testable interfaces over implicit coupling
5. **Documentation as Code**: Tests serve as living documentation
6. **Multi-Persona Design**: Serve DevOps Engineers, Developers, Compliance Auditors
7. **Audit Trails**: Comprehensive logging and traceability

## Troubleshooting

### Build Issues
```bash
# Verify Go version
go version  # Requires 1.19+

# Clear module cache if needed  
go clean -modcache
go mod tidy
go build -o sync-tools ./cmd/sync-tools
```

### Test Failures
```bash
# Run specific test suite
go test ./test/bdd -tags bdd -run 'TestFeatures'

# Verbose output for debugging
go test ./test/bdd -tags bdd -v -run 'APPEND'

# Check test isolation
go test ./test/bdd -tags bdd -count=2
```

### BDD Development
```bash
# Run specific feature
go test ./test/bdd -tags bdd -run 'BasicSync'

# Debug step definitions
# Check test/bdd/steps/sync_steps.go for available steps

# Add new scenarios to features/*.feature files
# Implement step definitions in test/bdd/steps/
```

Remember: sync-tools provides reliable rsync-based synchronization with advanced filtering and post-sync actions. Always validate core functionality after changes and maintain 100% BDD test pass rate.

For comprehensive development guidelines, architectural principles, and mandatory practices, see [CLAUDE.md](../CLAUDE.md).