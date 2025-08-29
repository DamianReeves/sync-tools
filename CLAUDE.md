# Claude Code Guidelines for sync-tools Project

## Project Context
This is the **sync-tools** repository - a Go CLI wrapper around rsync that provides fast directory synchronization with advanced filtering capabilities. The project focuses on providing a great experience for synchronizing directories with features like .syncignore files, gitignore import, whitelist mode, and sophisticated filtering rules.

## Target Personas
- **DevOps Engineers**: Use sync-tools for automated deployment and backup workflows
- **Developers**: Synchronize project files between development environments and remote servers
- **System Administrators**: Manage file synchronization across multiple systems with complex filtering needs
- **Data Managers**: Handle large-scale directory synchronization with precise inclusion/exclusion control

## Conversation and Interaction Style
- Direct conversation without unnecessary praise or pleasantries
- Peer-to-peer collaboration between skilled developers
- Offer pointed ideas about design decisions and technical trade-offs
- Focus on substance over social conventions
- Treat the human as an expert who values substantive technical input

## Problem-Solving Approach
- Lead with architectural thinking and design considerations
- Break down complex problems into testable, composable components
- Consider integration implications from the start
- Document architectural decisions and reasoning
- Create visual representations (Mermaid diagrams) for complex systems

## Code Assistance Style
- Provide working, production-ready code that prioritizes testability
- Design for composition and modularity
- Include comprehensive test examples alongside implementation
- Explain architectural trade-offs and design implications
- Point out integration concerns and external system dependencies

## Documentation and Communication
- Create design documents in markdown format
- Use Mermaid diagrams for system architecture, data flow, and component relationships
- Maintain architectural decision records (ADRs) for significant choices
- Explain the "why" behind architectural recommendations
- Focus on making design thinking shareable with other developers

## Technical Preferences
- Design all components with testing as a first-class concern
- Favor composition over inheritance and complex hierarchies
- Consider external system integration points early in design
- Value explicit, testable interfaces over implicit coupling
- Think about how components can be independently verified and composed

## Architecture Preferences
- Design for testability from the ground up
- Use composition patterns to build flexible, reusable components
- Implement clear integration boundaries for external systems
- Consider how each design decision affects testing and composition
- Document architectural decisions in ADR format
- **Multi-persona interfaces**: Design UX/APIs that serve Product Managers, Developers, and Compliance Auditors
- **Executable specifications**: Support BDD/Gherkin as a first-class specification format
- **Audit trails**: Build comprehensive logging and traceability for compliance requirements

## Testing Approach - MANDATORY BDD/TDD DISCIPLINE

### BDD for Integration and End-to-End Flows (PREFERRED)
- **Start with Gherkin Features**: All new integration features MUST begin with Gherkin scenarios
- **Red/Green/Refactor Cycle**: Follow strict TDD discipline:
  1. **Red**: Write failing BDD scenarios first using Gherkin
  2. **Green**: Implement minimal code to make scenarios pass
  3. **Refactor**: Clean up code while maintaining green tests
- **Feature Files First**: Create `.feature` files before writing any implementation code
- **Step Definitions**: Implement step definitions that fail initially (red phase)
- **Living Documentation**: Gherkin scenarios serve as executable specifications

### TDD for Unit-Level Development
- **Test First**: Write failing unit tests before implementing functionality
- **Small Increments**: Make minimal changes to achieve green tests
- **Comprehensive Coverage**: Ensure all code paths are tested
- **Fast Feedback**: Unit tests must run quickly for rapid development cycles

### Testing Requirements
- **BDD-First Approach**: End-to-end BDD tests for all user-facing features using Gherkin
- **TDD Discipline**: Unit tests written before implementation code
- **Red/Green/Refactor**: Strict adherence to TDD cycle
- **Integration tests** for runtime behavior and cross-component interaction
- **Performance benchmarks** for critical paths
- **Test Isolation**: Each test must be independent and repeatable
- **Living Documentation**: Tests serve as the primary source of behavioral documentation

## Core Domain Model
- **Sync Operations**: One-way and two-way directory synchronization with conflict resolution
- **Filter System**: Layered filtering with .syncignore, .gitignore import, and CLI overrides
- **Whitelist Mode**: Explicit path inclusion with "only" patterns for precise sync control
- **Remote Endpoints**: Support for rsync remote syntax (user@host:/path) alongside local paths
- **Conflict Preservation**: Two-way sync with automatic conflict file generation

## Current Architecture Status
- **Go CLI Framework**: Command-line interface built with modern Go patterns
- **Filter Engine**: Sophisticated pattern matching and exclusion/inclusion logic
- **Rsync Wrapper**: Efficient integration with rsync for actual file operations
- **Configuration System**: Support for both config files and pure CLI usage
- **Cross-Platform**: Built for Linux, macOS, and Windows compatibility

## Development Commands
- Build: `go build -o sync-tools ./cmd/sync-tools`
- Test: `go test ./...`
- Run: `go run ./cmd/sync-tools [command]`
- Install: `go install ./cmd/sync-tools`
- Lint: `golangci-lint run`

## Commit Standards
- All changes must compile successfully before commit (`go build`)
- **BDD tests must pass at 100% success rate** (`go test ./...`)
- **TDD discipline**: No production code without failing tests first
- **Linting must pass** (`golangci-lint run`)
- Include meaningful commit messages with context
- Use feature branches for significant changes
- Reference BDD scenarios in commit messages when applicable

## Response Format
- Use proper markdown formatting for all documentation
- Include Mermaid diagrams when illustrating system relationships
- Provide step-by-step architectural reasoning
- Show both implementation and corresponding test code
- Reference relevant patterns and integration considerations

## Development Tracker Maintenance (MANDATORY)

### Always Update DEVELOPMENT-TRACKER.md
- **Track Progress**: Update task status as work progresses (In Progress → Pending → Refined → Backlog)
- **Document Completions**: Move completed work to Changelog with completion dates
- **Record Decisions**: Document architectural decisions and trade-offs made during implementation
- **Maintain Accuracy**: Keep current status, priorities, and next steps accurate
- **Update on Every Session**: Refresh relevant sections whenever tasks are completed or new work begins

### DEVELOPMENT-TRACKER.md Structure
The tracker uses these sections:
- **TASKS**:
  - **In Progress**: Currently active work (limit 1-2 items)
  - **Pending**: Ready to start, clearly defined
  - **Refined**: Analyzed and planned, waiting for capacity
  - **Backlog**: Future considerations, less detailed
- **Changelog**: Completed work with dates and outcomes
- **Current Architecture Status**: Live snapshot of system state

### Priority System for Tasks
All tasks must include priority levels:
- **P0 - Critical**: Blocking issues, security vulnerabilities, system failures
- **P1 - High**: Important features, architecture improvements, performance issues
- **P2 - Medium**: Enhancements, documentation, developer experience improvements
- **P3 - Low**: Nice-to-have features, code quality improvements
- **P4 - Future**: Long-term considerations, research items

Format: `**Task Name** [Priority: P0 - Critical]`

### Required Updates
When completing tasks:
1. Move completed items from TASKS to Changelog
2. Update Current Architecture Status if architecture changed
3. Add new discoveries or follow-up tasks to appropriate TASKS sections
4. Update "Last Updated" date and overall status

## Next Development Priorities
1. **Go Implementation** - Migrate from Python to Go for better performance and single-binary distribution
2. **Enhanced Filter Engine** - More sophisticated pattern matching and conflict resolution
3. **Remote Sync Optimization** - Better handling of SSH connections and remote endpoints
4. **Documentation & Developer Experience** improvements

## Key Integration Points
- Cross-platform compatibility (Linux, macOS, Windows)
- CI/CD pipeline integration for automated deployments
- Development workflow integration for multi-environment sync
- System administration tools and backup solutions
- Remote server management and file distribution systems