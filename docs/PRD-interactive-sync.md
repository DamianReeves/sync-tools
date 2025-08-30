# Product Requirements Document: Two-Phased Interactive Sync

## Executive Summary

This PRD defines a two-phased interactive synchronization feature for sync-tools that provides users with fine-grained control over file synchronization operations. Similar to git's interactive rebase, users can review and modify sync operations before execution, choosing direction and selective inclusion for each file or directory.

## Problem Statement

Current synchronization tools operate in an all-or-nothing manner:
- Users cannot preview and selectively modify sync operations before execution
- No ability to change sync direction on a per-file basis
- Difficult to handle complex scenarios where some files should sync one way and others another way
- No audit trail of what sync decisions were made and why

## Target Personas

1. **DevOps Engineers**: Need precise control over deployment synchronization with audit trails
2. **Data Scientists**: Selectively sync large datasets and model files between environments
3. **System Administrators**: Carefully manage configuration file synchronization across servers
4. **Developers**: Fine-tune project file synchronization between development environments

## Solution Overview

A two-phased synchronization workflow:

**Phase 1 - Plan Generation**: Analyze differences and generate an editable sync plan file
**Phase 2 - Plan Execution**: Apply the reviewed and modified sync plan

## Detailed Requirements

### Phase 1: Sync Plan Generation

#### Command Structure
```bash
# Generate sync plan
sync-tools sync --plan [output-file] [options]
sync-tools sync --interactive [options]  # Opens in $EDITOR immediately

# Examples
sync-tools sync --source ./src --dest ./backup --plan sync-plan.txt
sync-tools sync --source ./dev --dest ./prod --plan deploy.plan --mode two-way
sync-tools sync --config deploy.toml --plan review.plan
sync-tools syncfile --plan staging.plan  # Use SyncFile as base configuration
```

#### Sync Plan File Format

The sync plan uses a line-based format similar to git rebase:

```
# Sync Plan Generated: 2025-08-30 10:45:00
# Generated from: sync-tools sync --config deploy.toml --exclude "*.log" --include-changes updates,conflicts --plan review.plan
# Source: /home/user/project
# Destination: /backup/project
# Mode: two-way
# Config: deploy.toml (layered with CLI overrides)
# Change filter: updates, conflicts (excludes new-in-source, new-in-dest, deletions, unchanged)
#
# Commands:
#   s2d, sync-to-dest, <<    - Sync from source to destination (source >> dest)
#   d2s, dest-to-source, >>  - Sync from destination to source (dest >> source)  
#   bid, bidirectional, <>   - Sync in both directions (bidirectional)
#   skip                     - Skip this item (commented out)
#
# Conflicts are marked with [CONFLICT] and require explicit resolution
# Lines starting with # are comments and will be ignored
#
# Format: <command> <item-type> <path> [size] [modified] [flags]
#
# Visual aliases make direction intuitive:
#   << = source flows to dest (like << redirection) 
#   >> = dest flows to source (like >> redirection)
#   <> = bidirectional flow (like <-> but shorter)

<< file   config/app.yml                 2.3K  2025-08-30T10:30:00  [update: newer-in-source]
>> file   config/database.yml            1.8K  2025-08-30T09:45:00  [update: newer-in-dest]
<> file   tests/integration.test.js      8.9K  2025-08-30T10:25:00  [CONFLICT: both-modified]

# Filtered out (not included due to --include-changes updates,conflicts):
# << file   src/main.js                  15.7K  [new-in-source]
# >> file   docs/README.md               4.5K   [new-in-dest] 
# skip file  logs/debug.log               2.1G   [deletion]

# Summary:
# Files matching filter: 3 (2 updates + 1 conflict)
# Filtered out: 3 (1 new-in-source + 1 new-in-dest + 1 deletion)
# Conflicts requiring resolution: 1
# Estimated transfer size: 12.5K
```

#### SyncFile Multi-Operation Plans

When generating plans from SyncFiles with multiple `SYNC` operations, the plan file includes sections for each operation:

```
# Sync Plan Generated: 2025-08-30 10:45:00
# Generated from: sync-tools syncfile DeploymentFile --plan multi-deploy.plan
# SyncFile: /home/user/project/DeploymentFile
#
# === Operation 1: Frontend Assets ===
# Source: ./frontend/dist
# Destination: /var/www/assets
# Mode: one-way
# Filters: EXCLUDE logs/, GITIGNORE true

<< file   css/main.css                   45.2K 2025-08-30T10:30:00  [newer]
<< file   js/bundle.min.js               890K  2025-08-30T10:32:00  [newer]
<< dir    images/                        12.4M 2025-08-30T10:25:00  [modified]

# === Operation 2: Configuration Sync ===  
# Source: ./config
# Destination: /etc/myapp
# Mode: two-way
# Filters: EXCLUDE *.local, ONLY *.yml

<> file   app.yml                        2.3K  2025-08-30T10:30:00  [newer]
>> file   database.yml                   1.8K  2025-08-30T09:45:00  [newer-in-dest]
# skip file app.local.yml                 892B  2025-08-30T10:15:00  [excluded]

# Summary:
# Operations: 2
# Total files to sync: 5
# Total estimated transfer: 13.2M
```

#### Plan Generation Logic

1. **Difference Analysis**
   - Compare source and destination using rsync dry-run
   - Identify new, modified, deleted, and conflicting files
   - Calculate sizes and modification times

2. **Intelligent Defaults**
   - Newer files default to sync from newer to older (auto-select `<<` or `>>`)
   - New files default to sync from where they exist (`<<` for new-in-source, `>>` for new-in-dest)
   - Deleted files prompt for confirmation (commented by default)
   - Conflicts are marked with `<>` but left uncommented for user decision

3. **Command Syntax and Aliases**
   
   | Command | Aliases | Visual | Meaning | Use Case |
   |---------|---------|--------|---------|----------|
   | `s2d`, `sync-to-dest` | `<<` | Source → Dest | Push to destination | Deployments, backups |
   | `d2s`, `dest-to-source` | `>>` | Source ← Dest | Pull from destination | Config retrieval, downloads |
   | `bid`, `bidirectional` | `<>` | Source ↔ Dest | Sync both ways | Development environments |
   | `skip` | `# <any command>` | Commented | Ignore this item | Temporary files, large assets |
   
   The visual aliases (`<<`, `>>`, `<>`) are recommended for their clarity and editing convenience.

4. **Metadata Collection**
   - File/directory type
   - Size (human-readable) 
   - Modification timestamp
   - Status flags: [new], [modified], [deleted], [newer], [older], [conflict], [large]

### Phase 2: Sync Plan Execution

#### Command Structure
```bash
# Execute sync plan
sync-tools sync --apply-plan <plan-file> [options]

# Examples
sync-tools sync --apply-plan sync-plan.txt
sync-tools sync --apply-plan deploy.plan --dry-run  # Preview execution
sync-tools sync --apply-plan sync-plan.txt --verbose --config base.toml
sync-tools sync --apply-plan multi.plan --exclude "*.tmp"  # Override plan settings
```

#### Execution Process

1. **Plan Validation**
   - Verify plan file syntax
   - Check source/destination paths still exist
   - Validate file states haven't changed critically
   - Warn about any discrepancies

2. **Operation Execution**
   - Process operations in order
   - Group operations by direction for efficiency
   - Show progress for each operation
   - Log all operations to audit file

3. **Conflict Handling**
   - For bidirectional conflicts, use newest-wins by default
   - Support conflict resolution strategies: newest-wins, largest-wins, source-wins, dest-wins
   - Create .conflict backups when configured
   - **Interactive merge tool integration** for file-level conflicts

4. **Configuration Precedence During Execution**
   - Plan file settings (embedded in plan header)
   - Runtime config file (`--config` during apply-plan)
   - Runtime CLI flags (`--verbose`, `--exclude`, etc. during apply-plan)
   - Plan files are self-contained but can be overridden for safety (dry-run, verbose, etc.)

### Editor Integration

#### Workflow
```bash
# Open in default editor
sync-tools sync --interactive --source ./src --dest ./backup

# Workflow:
# 1. Generates plan to temporary file
# 2. Opens in $EDITOR (or --editor flag)
# 3. User edits and saves
# 4. On editor exit, validates plan
# 5. Prompts for confirmation
# 6. Executes plan
```

#### Editor Features
- **Syntax highlighting rules** (for vim, emacs, VS Code)
- **Visual direction indicators**: `<<`, `>>`, `<>` are immediately recognizable
- **Comment/uncomment shortcuts**: Toggle lines with `#`
- **Bulk operations**: 
  - Comment all: `:%s/^/<#/g` (vim)
  - Change all to source→dest: `:%s/^>>/<</ g` (vim)
  - Change all to bidirectional: `:%s/^<</</g` (vim)
- **Quick editing patterns**:
  - `<<` + `>>` + `<>` are easy to type and visually distinct
  - No need to remember long command names
  - Direction is immediately clear when scanning the file
- **Validation on save**

#### Interactive Merge Tool Integration

For files marked with conflicts, sync-tools can launch external merge tools for resolution:

```bash
# Enable interactive merge resolution during plan execution
sync-tools sync --apply-plan conflicts.plan --interactive-merge

# Configure merge tool (respects git config)
sync-tools sync --apply-plan conflicts.plan --merge-tool vimdiff
sync-tools sync --apply-plan conflicts.plan --merge-tool vscode  # VS Code merge editor
sync-tools sync --apply-plan conflicts.plan --merge-tool meld    # GUI merge tool
```

##### Merge Tool Integration Workflow

1. **Conflict Detection**: When executing a plan with `<>` (bidirectional) conflicts
2. **Tool Launch**: Opens configured merge tool with three-way merge:
   - **Base**: Last known common version (if available) or empty
   - **Source**: Current source file version
   - **Dest**: Current destination file version
3. **User Resolution**: User resolves conflicts in familiar merge interface
4. **Result Handling**: Resolved file is applied to both source and destination
5. **Backup Creation**: Original conflicting files saved as `.conflict-source` and `.conflict-dest`

##### Supported Merge Tools

| Tool | Command Template | Description |
|------|------------------|-------------|
| `vimdiff` | `vim -d {source} {dest}` | Vim's built-in diff mode |
| `nvim` | `nvim -d {source} {dest}` | Neovim diff mode |  
| `vscode` | `code --diff {source} {dest}` | VS Code merge editor |
| `meld` | `meld {source} {dest}` | Cross-platform GUI merge tool |
| `kdiff3` | `kdiff3 {source} {dest}` | KDE merge tool |
| `p4merge` | `p4merge {source} {dest}` | Perforce visual merge tool |
| `beyond-compare` | `bcomp {source} {dest}` | Beyond Compare |
| `git` | Uses `git config merge.tool` setting | Respects git configuration |

##### Plan File Integration

Conflicts can specify preferred resolution methods:

```
# Standard conflict - will prompt for merge tool
<> file   config/settings.json          2.1K  2025-08-30T10:25:00  [CONFLICT: both-modified]

# Auto-resolve with newest-wins strategy  
<> file   logs/debug.log                15.2M 2025-08-30T10:30:00  [CONFLICT: both-modified, auto:newest]

# Force specific merge tool for this conflict
<> file   src/complex.js                8.4K  2025-08-30T10:20:00  [CONFLICT: both-modified, merge-tool:vscode]
```

##### Configuration Options

```toml
[merge]
default_tool = "vimdiff"
auto_backup = true
backup_suffix = ".conflict-backup"
prompt_before_merge = true
timeout_seconds = 300  # Auto-skip if merge takes too long

[merge.strategies]
text_files = "interactive"      # Use merge tool for text files
binary_files = "newest-wins"    # Auto-resolve binary conflicts
large_files = "prompt"          # Ask user for large files (>10MB)
```

##### Advanced Merge Features

###### Three-Way Merge Support
When available, use common ancestor for better conflict resolution:
```bash
# Enable three-way merge with git integration
sync-tools sync --apply-plan conflicts.plan --interactive-merge --use-git-base
```

###### Batch Conflict Resolution
```bash
# Process all conflicts with same tool
sync-tools sync --apply-plan conflicts.plan --merge-tool meld --batch-conflicts

# Skip conflicts and generate resolution plan
sync-tools sync --apply-plan conflicts.plan --skip-conflicts --generate-conflict-plan resolve.plan
```

###### Integration with Git Workflow
```bash
# For git repositories, respect .gitattributes merge settings
sync-tools sync --apply-plan conflicts.plan --respect-git-attributes

# Use git's configured merge tool
sync-tools sync --apply-plan conflicts.plan --merge-tool git
```

### Advanced Features

#### Configuration Integration

The two-phased sync inherits all existing sync-tools configuration capabilities:

##### Command Line Options
```bash
# All existing flags work with plan generation
sync-tools sync --plan output.plan --only "*.js" --only "*.ts"
sync-tools sync --plan output.plan --exclude "node_modules/" --use-source-gitignore
sync-tools sync --plan output.plan --dry-run --verbose  # Preview what would be planned
sync-tools syncfile MySyncFile --plan deploy.plan --exclude "logs/"  # Override SyncFile settings
```

##### Change-Type Filtering

Plan generation can be filtered to include only specific types of changes:

```bash
# Filter by change type
sync-tools sync --plan conflicts.plan --include-changes conflicts
sync-tools sync --plan new-files.plan --include-changes new-in-source
sync-tools sync --plan updates.plan --include-changes updates,conflicts
sync-tools sync --plan review.plan --exclude-changes new-in-dest  # Skip files only in destination

# Available change types:
#   new-in-source    - Files that exist in source but not in destination
#   new-in-dest      - Files that exist in destination but not in source  
#   updates          - Files that exist in both but differ (newer/modified)
#   conflicts        - Files with bidirectional conflicts (both modified)
#   deletions        - Files deleted from source (for cleanup review)
#   unchanged        - Files that are identical (rarely needed)
#   all             - All changes (default)

# Practical examples
sync-tools sync --plan new-only.plan --include-changes new-in-source  # Only new files to deploy
sync-tools sync --plan conflicts-only.plan --include-changes conflicts  # Focus on conflicts
sync-tools sync --plan cleanup.plan --include-changes new-in-dest,deletions  # Review removals
sync-tools syncfile --plan deploy-new.plan --include-changes new-in-source,updates  # Deployment focus
```

##### Config File Integration
```bash
# Use TOML config as base configuration
sync-tools sync --config production.toml --plan prod-deploy.plan

# Override config settings with CLI flags
sync-tools sync --config base.toml --mode one-way --plan override.plan
```

##### SyncFile Integration
```bash
# Generate plan from SyncFile operations
sync-tools syncfile --plan review.plan  # Default SyncFile
sync-tools syncfile MySyncFile --plan custom.plan  # Specific SyncFile
sync-tools syncfile --list --plan preview.plan  # Preview operations in plan format

# Multiple SyncFile operations become multiple plan sections
sync-tools syncfile MultiOpSyncFile --plan multi.plan
```

##### Configuration Layering
The feature respects sync-tools' configuration hierarchy:
1. **Built-in defaults**
2. **Config file settings** (`--config file.toml`)  
3. **SyncFile instructions** (`EXCLUDE`, `MODE`, `GITIGNORE`, etc.)
4. **Command-line overrides** (`--exclude`, `--mode`, `--only`, etc.)

Plan generation uses the final resolved configuration to determine defaults and filtering.

#### Change-Type Filtering Use Cases

Different filtering strategies serve specific workflow needs:

##### Deployment-Focused Plans
```bash
# Focus only on new features and updates (skip cleanup)
sync-tools sync --plan deploy.plan --include-changes new-in-source,updates
# Result: Only files being added or updated, no deletions or conflicts to worry about
```

##### Conflict Resolution Plans
```bash
# Show only items that need human decision
sync-tools sync --plan resolve.plan --include-changes conflicts
# Result: Focused plan with only bidirectional conflicts requiring manual resolution
```

##### Cleanup Review Plans  
```bash
# Review what will be removed from destination
sync-tools sync --plan cleanup.plan --include-changes new-in-dest,deletions
# Result: Files that exist only in destination or were deleted from source
```

##### Update-Only Plans
```bash
# Focus on files that changed, ignore new/deleted files
sync-tools sync --plan updates.plan --include-changes updates
# Result: Only existing files that have newer versions in source or dest
```

##### Comprehensive Review Plans
```bash
# Everything except unchanged files (reduces noise)
sync-tools sync --plan full-review.plan --exclude-changes unchanged
# Result: All meaningful changes, skip identical files
```

##### Combined Filtering
```bash
# Complex deployment: new files and updates, but not conflicts or deletions
sync-tools sync --plan safe-deploy.plan --include-changes new-in-source,updates \
  --exclude-changes conflicts,deletions
# Result: Safe deployment that avoids destructive operations and conflicts
```

#### Filter Precedence Rules

1. **Default behavior**: `--include-changes all` (includes everything)
2. **Include takes precedence**: If `--include-changes` is specified, only those types are included
3. **Exclude refines**: If `--exclude-changes` is used with include, it removes types from the include list
4. **Exclude only**: If only `--exclude-changes` is specified, it removes types from the default "all"

```bash
# These are equivalent:
sync-tools sync --plan example.plan --include-changes new-in-source,updates,conflicts
sync-tools sync --plan example.plan --exclude-changes new-in-dest,deletions,unchanged

# Include wins over exclude in conflicts:
sync-tools sync --plan example.plan --include-changes conflicts --exclude-changes conflicts
# Result: No conflicts included (exclude refines the include)
```

#### Plan Templates
```bash
# Save plan as template (with patterns, not specific files)
sync-tools sync --plan-template deploy.template

# Apply template to new sync
sync-tools sync --apply-template deploy.template --source ./new-src --dest ./new-dest
```

#### Audit Trail
```bash
# Execution generates audit log
sync-tools sync --apply-plan sync.plan --audit-log sync-audit.log

# Audit log format (JSON Lines)
{"timestamp": "2025-08-30T10:45:00Z", "action": "s2d", "file": "config/app.yml", "size": 2355, "result": "success"}
{"timestamp": "2025-08-30T10:45:01Z", "action": "d2s", "file": "config/database.yml", "size": 1843, "result": "success"}
```

## User Experience Flow

### Typical Workflow

1. **Generate Plan**
   ```bash
   sync-tools sync --source ./dev --dest ./prod --plan review.plan
   ```

2. **Review and Edit**
   ```bash
   vim review.plan
   # - Comment out large log files
   # - Change critical configs to d2s (pull from production)
   # - Resolve conflicts by choosing direction
   ```

3. **Dry Run**
   ```bash
   sync-tools sync --apply-plan review.plan --dry-run
   ```

4. **Execute**
   ```bash
   sync-tools sync --apply-plan review.plan --audit-log deploy-$(date +%Y%m%d).log
   ```

### Interactive Mode Flow

1. **Single Command**
   ```bash
   sync-tools sync --interactive --source ./dev --dest ./prod
   ```

2. **Editor Opens** with generated plan

3. **User Edits** and saves

4. **Confirmation Prompt**
   ```
   Ready to execute sync plan:
   - 45 files to sync source → dest
   - 12 files to sync dest → source  
   - 8 files bidirectional
   - 23 files skipped
   
   Continue? [y/N]
   ```

5. **Execution** with progress display

## Success Metrics

1. **Reduced Sync Errors**: 50% reduction in accidental overwrites
2. **Adoption Rate**: 30% of power users adopt interactive mode within 3 months
3. **Time Saved**: Average 20% reduction in sync-related troubleshooting
4. **User Satisfaction**: 4.5+ star rating for the feature

## Technical Considerations

### Performance
- Plan generation should complete in < 5 seconds for 10,000 files
- Plan parsing and validation < 1 second
- Minimal memory overhead for large plans

### Compatibility
- Plan files are portable text files
- Support for Windows, macOS, and Linux editors
- UTF-8 encoding for international file names

### Error Handling
- Graceful handling of editor crashes
- Recovery from partial plan execution
- Clear error messages for invalid plan syntax

## Implementation Phases

### MVP (Phase 1)
- Basic plan generation
- Simple s2d/d2s/skip commands
- Plan execution
- Basic validation

### Enhanced (Phase 2)
- Bidirectional sync
- Conflict detection and marking
- Editor integration
- Progress display

### Advanced (Phase 3)
- Plan templates
- Audit logging
- Custom conflict resolution strategies
- Bulk operations in plan files

## Example Use Cases

### 1. Deployment Sync with SyncFile Base
DevOps engineer uses SyncFile as deployment template, then reviews with plan:

**DeploymentFile:**
```
# Production deployment base
SYNC ./app /var/www/myapp
MODE one-way  
EXCLUDE logs/
EXCLUDE *.local
GITIGNORE true
```

**Workflow:**
```bash
# Generate focused deployment plan - only new and updated files
sync-tools syncfile DeploymentFile --plan prod-review.plan --exclude "test/" \
  --include-changes new-in-source,updates

# Edit plan to:
# - Skip large asset files that haven't changed
# - Pull critical config from production (change to d2s)
# - Add manual verification steps as comments

vim prod-review.plan

# Separate conflict resolution plan if needed
sync-tools syncfile DeploymentFile --plan conflicts.plan --include-changes conflicts

# Execute with audit logging
sync-tools sync --apply-plan prod-review.plan --audit-log deploy-$(date +%Y%m%d).log
```

### 2. Development Environment Sync with Config Layering
Developer uses base config with CLI overrides:

**dev-sync.toml:**
```toml
mode = "two-way"
exclude_hidden_dirs = true
use_source_gitignore = true

[filters]
ignore_src = ["node_modules/", "*.log"]
only = ["src/", "config/", "docs/"]
```

**Workflow:**
```bash
# Generate plan using config + CLI overrides - focus on meaningful changes
sync-tools sync --config dev-sync.toml --source ~/laptop/project --dest ~/workstation/project \
  --exclude "coverage/" --exclude-changes unchanged --plan dev-review.plan

# Separate plan for conflict resolution
sync-tools sync --config dev-sync.toml --source ~/laptop/project --dest ~/workstation/project \
  --include-changes conflicts --plan dev-conflicts.plan

# Review dev-review.plan and selectively sync:
# - Most files: bidirectional  
# - IDE configs: skip (OS-specific)
# - Database dumps: pull from workstation only (d2s)

# Handle conflicts with interactive merge tools  
sync-tools sync --apply-plan dev-conflicts.plan --interactive-merge --merge-tool vscode

# Alternative: Edit conflict plan to specify per-file merge tools
vim dev-conflicts.plan
# Then execute normally - tools specified in plan file will be used
```

### 3. Backup with Multi-Operation SyncFile
System admin uses complex SyncFile for different backup strategies:

**BackupFile:**
```
# Critical configs - bidirectional with conflict preservation
SYNC /etc/myapp ./backup/config
MODE two-way

# Application data - one-way backup only  
SYNC /var/lib/myapp ./backup/data
MODE one-way
EXCLUDE cache/
EXCLUDE *.tmp

# Logs - selective backup of recent files
SYNC /var/log/myapp ./backup/logs  
MODE one-way
ONLY *.log
```

**Workflow:**
```bash
# Generate comprehensive backup plan - exclude identical files for faster review
sync-tools syncfile BackupFile --plan backup-review.plan --exclude "*.log.*" \
  --exclude-changes unchanged

# Generate conflict-focused plan for critical configs
sync-tools syncfile BackupFile --plan backup-conflicts.plan --include-changes conflicts

# Generate cleanup plan for space management  
sync-tools syncfile BackupFile --plan backup-cleanup.plan --include-changes new-in-dest

# Review backup-review.plan to:
# - Skip logs older than 7 days  
# - Verify critical configs are included
# - Skip cache directories larger than 1GB

# Handle config conflicts carefully
vim backup-conflicts.plan

# Review and approve cleanup operations
vim backup-cleanup.plan

# Execute in stages
sync-tools sync --apply-plan backup-review.plan --verbose
sync-tools sync --apply-plan backup-conflicts.plan --verbose  
sync-tools sync --apply-plan backup-cleanup.plan --verbose
```

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex UI confuses users | Low adoption | Provide simple mode with sensible defaults |
| Plan file corruption | Data loss | Validation before execution, backup original plan |
| Editor integration issues | Poor UX | Fallback to simple text file generation |
| Performance with large directories | User frustration | Implement pagination and filtering |

## Open Questions

1. Should we support XML/JSON format in addition to text format?
2. How to handle symbolic links and special files?
3. Should plans be reusable across different source/dest pairs?
4. Integration with version control for plan files?

## Appendix

### Sample Configuration
```toml
[interactive]
default_editor = "vim"
auto_skip_patterns = ["*.log", "*.tmp", ".DS_Store"]
conflict_strategy = "newest-wins"
show_size_threshold = "1MB"
audit_log_dir = "~/.sync-tools/audit/"
```

### Command Reference Summary
```bash
# Generate plan
sync-tools sync --plan <file>

# Interactive mode
sync-tools sync --interactive

# Apply plan
sync-tools sync --apply-plan <file>

# Dry run
sync-tools sync --apply-plan <file> --dry-run

# With audit
sync-tools sync --apply-plan <file> --audit-log <log-file>
```