# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0-M1] - 2025-09-01

### Added
- Automatic binary building in BDD test setup
- Comprehensive GoReleaser v2 configuration for automated releases
- Homebrew tap integration with DamianReeves/homebrew-tap
- Cross-platform build support (Linux, macOS x64/ARM64, Windows)
- Enhanced GitHub Actions workflow with BDD test validation
- Complete documentation overhaul in docs/user_guide.adoc
- HOMEBREW-TAP-SETUP-GUIDE.md for tap repository setup
- Copilot instructions integration with CLAUDE.md guidelines
- @wip tag filtering in BDD test runner

### Changed
- Migrated from Python-based documentation to Go architecture
- Updated Makefile VERSION from 0.2.0 to 0.4.0-M1
- Enhanced merge flags and conflict resolution logging
- Improved BDD test framework with automatic binary compilation
- Updated copilot instructions to reflect current Go implementation

### Fixed
- BDD test setup now builds sync-tools binary automatically
- Removed dependency on pre-existing checked-in binary
- Fixed conflict resolution test validation logic
- Resolved "fork/exec sync-tools: no such file or directory" errors

### Removed
- Outdated Python-related Release tasks
- sync-tools binary from git tracking (now built on-demand)

## [0.3.0] - 2025-08-30

### Added
- SyncFile APPEND and PREPEND post-sync actions
- Complete BDD test coverage with 57 passing scenarios
- Two-phased interactive sync with plan generation and execution
- Git patch generation capabilities
- Advanced filtering with .syncignore and .gitignore integration
- Whitelist mode with "only" patterns for precise sync control
- Conflict resolution strategies (newest-wins, oldest-wins, interactive)
- Comprehensive merge tool integration
- Plan validation and syntax checking

### Changed
- Enhanced sync operation workflow with plan-and-execute model
- Improved conflict handling with automatic conflict file generation
- Better error messaging and user feedback

## [0.2.1] - 2025-08-29

### Fixed
- Minor bug fixes and stability improvements

## [0.2.0] - 2025-08-28

### Added
- Initial Go implementation of sync-tools CLI
- Basic one-way and two-way directory synchronization
- rsync integration with advanced filtering capabilities
- .syncignore file support
- Command-line interface with comprehensive options

### Changed
- Complete rewrite from previous implementation
- Modern Go architecture with clean separation of concerns

## [0.1.1rc1] - 2025-08-25

### Added
- Initial release candidate
- Basic synchronization functionality
- Foundation for advanced filtering features

[Unreleased]: https://github.com/DamianReeves/sync-tools/compare/v0.4.0-M1...HEAD
[0.4.0-M1]: https://github.com/DamianReeves/sync-tools/compare/v0.3.0...v0.4.0-M1
[0.3.0]: https://github.com/DamianReeves/sync-tools/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/DamianReeves/sync-tools/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/DamianReeves/sync-tools/compare/v0.1.1rc1...v0.2.0
[0.1.1rc1]: https://github.com/DamianReeves/sync-tools/releases/tag/v0.1.1rc1
