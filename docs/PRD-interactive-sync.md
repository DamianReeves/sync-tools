# Product Requirements Document: Interactive Sync

## Executive Summary

This PRD defines interactive synchronization features for sync-tools that provide users with intuitive control over file synchronization operations. The system offers two complementary approaches: a **wizard-like UI** using Bubble Tea for guided sync setup, and a **plan-based workflow** similar to git's interactive rebase for advanced users who want fine-grained control over sync operations.

## Problem Statement

Current synchronization tools operate in an all-or-nothing manner:
- Users lack intuitive interfaces for configuring sync operations
- No guided workflow to help users make informed sync decisions
- Users cannot easily preview and selectively modify sync operations before execution
- No ability to change sync direction on a per-file basis
- Difficult to handle complex scenarios where some files should sync one way and others another way
- No audit trail of what sync decisions were made and why

## Target Personas

1. **DevOps Engineers**: Need precise control over deployment synchronization with audit trails
2. **Data Scientists**: Selectively sync large datasets and model files between environments  
3. **System Administrators**: Carefully manage configuration file synchronization across servers
4. **Developers**: Fine-tune project file synchronization between development environments
5. **Casual Users**: Need simple, guided sync setup without complex command-line options

## Solution Overview

The interactive sync feature provides two complementary workflows:

### Wizard-Like Interactive UI (Primary - Phase 1)
A **Bubble Tea-powered terminal UI** that guides users through sync setup with:
- **Setup Wizard**: Step-by-step sync configuration (source, destination, mode)
- **Directory Tree Selection**: Visual folder browser to select what to sync
- **Exclusion Management**: Interactive interface to exclude unwanted files/folders
- **Preview & Confirmation**: Show planned operations before execution
- **Progress Monitoring**: Real-time sync progress with visual indicators

### Plan-Based Workflow (Advanced - Phase 2)  
A **two-phased synchronization workflow** for power users:
- **Phase 1 - Plan Generation**: Analyze differences and generate an editable sync plan file
- **Phase 2 - Plan Execution**: Apply the reviewed and modified sync plan

## Detailed Requirements

## Interactive Wizard UI (Bubble Tea Implementation)

### Command Structure
```bash
# Launch interactive wizard mode
sync-tools sync --wizard
sync-tools wizard  # Dedicated wizard command (alias)

# Launch with pre-filled options
sync-tools sync --wizard --source ./src
sync-tools wizard --destination ./backup --mode one-way

# Examples
sync-tools wizard                              # Full guided setup
sync-tools sync --wizard --source ./project   # Pre-set source, wizard for rest
```

### Wizard Flow Overview

The wizard provides a step-by-step interface with the following screens:

1. **Welcome & Mode Selection**
2. **Source Directory Selection**
3. **Destination Directory Selection** 
4. **Sync Options Configuration**
5. **Directory Tree Filter Selection**
6. **Exclusion Pattern Management**
7. **Preview & Confirmation**
8. **Execution & Progress Monitoring**

### Detailed Screen Specifications

#### Screen 1: Welcome & Mode Selection
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          sync-tools Interactive Wizard                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Welcome! This wizard will guide you through setting up a sync operation.  │
│                                                                             │
│  Select sync mode:                                                          │
│                                                                             │
│  ● One-way sync    (source → destination)                                  │
│    ○ Two-way sync   (source ↔ destination)  [Coming in future release]     │
│                                                                             │
│  One-way sync copies files from source to destination only.                │
│  Files in destination that don't exist in source will be preserved.        │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Tab] Navigate • [Enter] Select • [q] Quit                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- Radio button selection for sync mode
- Clear description of selected mode
- Future modes clearly marked as "Coming soon"
- **Initial Release**: Only one-way sync available
- **Future Release**: Two-way sync support

#### Screen 2: Source Directory Selection
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Select Source Directory                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Choose the directory to sync FROM:                                        │
│                                                                             │
│  Current: /home/user/projects                                              │
│                                                                             │
│  📁 /home/user/                                                            │
│  ├── 📁 Documents/                                                         │
│  ├── 📁 Downloads/                                                         │
│  ├── 📁 projects/ ●                                                        │
│  │   ├── 📁 my-app/                                                        │
│  │   ├── 📁 sync-tools/                                                    │
│  │   └── 📁 website/                                                       │
│  └── 📁 workspace/                                                         │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────┐           │
│  │ Path: /home/user/projects                                  │           │
│  └─────────────────────────────────────────────────────────────┘           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate • [→] Expand • [←] Collapse • [Enter] Select • [/] Search   │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- Interactive directory tree browser
- Real-time path display
- Expandable/collapsible folders
- Search functionality for large directory structures
- Breadcrumb navigation
- Visual indicators (folders, current selection)

#### Screen 3: Destination Directory Selection  
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Select Destination Directory                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Choose the directory to sync TO:                                          │
│                                                                             │
│  Source: /home/user/projects                                               │
│  Current: /backup/projects                                                 │
│                                                                             │
│  📁 /backup/                                                               │
│  ├── 📁 documents/                                                         │
│  ├── 📁 projects/ ●                                                        │
│  │   ├── 📁 old-backups/                                                   │
│  │   └── 📁 archives/                                                      │
│  └── 📁 system/                                                            │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────┐           │
│  │ Path: /backup/projects                                     │           │
│  └─────────────────────────────────────────────────────────────┘           │
│                                                                             │
│  □ Create destination directory if it doesn't exist                        │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate • [→] Expand • [←] Collapse • [Enter] Select • [Space] Toggle│
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- Same directory browser as source selection
- Shows selected source path for reference
- Option to create destination directory
- Warning if destination exists and contains files
- Path validation and permission checking

#### Screen 4: Sync Options Configuration
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Sync Options                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Configure sync behavior:                                                  │
│                                                                             │
│  Basic Options:                                                            │
│  ☑ Dry run (preview only - no actual changes)                             │
│  ☑ Verbose output                                                         │
│  □ Delete files in destination that don't exist in source                 │
│                                                                             │
│  Advanced Options:                                                         │
│  ☑ Use .gitignore files to exclude files                                  │
│  □ Follow symbolic links                                                   │
│  □ Preserve file timestamps                                                │
│  ☑ Exclude hidden files and directories                                   │
│                                                                             │
│  Performance:                                                              │
│  Parallel transfers: [4     ] (1-16)                                      │
│  Transfer timeout:   [300s  ] seconds                                     │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate • [Space] Toggle • [←→] Adjust • [Tab] Next section         │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- Checkbox options for boolean settings
- Slider/input controls for numeric values
- Grouped options (Basic, Advanced, Performance)
- Helpful descriptions for each option
- Sensible defaults with dry-run enabled initially

#### Screen 5: Directory Tree Filter Selection
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       Select Folders to Sync                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Choose which folders from source should be synchronized:                   │
│                                                                             │
│  Source: /home/user/projects                                               │
│                                                                             │
│  📁 projects/                                                              │
│  ├── ☑ 📁 src/                           (156 files, 2.4 MB)             │
│  ├── ☑ 📁 docs/                          (24 files, 180 KB)              │
│  ├── ☑ 📁 config/                        (8 files, 32 KB)                │
│  ├── ☐ 📁 node_modules/                  (15,420 files, 480 MB)           │
│  ├── ☐ 📁 .git/                          (892 files, 45 MB)               │
│  ├── ☑ 📁 assets/                        (67 files, 8.2 MB)              │
│  ├── ☐ 📁 logs/                          (145 files, 23 MB)               │
│  ├── ☑ 📁 tests/                         (43 files, 256 KB)              │
│  └── ☐ 📁 tmp/                           (0 files, 0 MB)                  │
│                                                                             │
│  Selected: 298 files (11.1 MB) • Excluded: 16,457 files (548 MB)          │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate • [Space] Toggle • [a] Select All • [n] Select None         │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- **Interactive checkbox tree** showing all top-level folders from source
- **File count and size information** for each folder to help users decide
- **Real-time summary** of selected vs excluded files and sizes
- **Bulk operations**: Select all, select none, toggle selection
- **Smart defaults**: Common exclude patterns (node_modules, .git, logs, tmp) unchecked by default
- **Visual indicators**: Clear distinction between selected ☑ and unselected ☐ folders

#### Screen 6: Exclusion Pattern Management  
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        File Exclusion Patterns                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Add patterns to exclude specific files or folders:                        │
│                                                                             │
│  Current patterns:                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ *.log                                                               │   │
│  │ *.tmp                                                               │   │
│  │ .DS_Store                                                          │   │
│  │ __pycache__/                                                       │   │
│  │ *.pyc                                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Add new pattern:                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ *.env                                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Common patterns:                                                          │
│  [*.log] [*.tmp] [node_modules/] [.git/] [*.env] [coverage/]              │
│                                                                             │
│  Pattern examples:                                                         │
│  • *.log          - All log files                                         │
│  • temp/          - Entire temp directory                                 │
│  • **/cache/      - Any cache directory at any level                      │
│  • !important.*  - Don't exclude files starting with "important"          │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [↑↓] Navigate • [Enter] Add • [Delete] Remove • [Tab] Quick add patterns  │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- **Editable list** of exclusion patterns with add/remove functionality
- **Quick-add buttons** for common patterns
- **Pattern examples** and syntax help
- **Real-time validation** of pattern syntax
- **Smart suggestions** based on detected file types in source

#### Screen 7: Preview & Confirmation
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Sync Preview                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Ready to sync with these settings:                                        │
│                                                                             │
│  Mode:         One-way (source → destination)                             │
│  Source:       /home/user/projects                                         │
│  Destination:  /backup/projects                                            │
│  Dry run:      Yes (preview mode)                                         │
│                                                                             │
│  Selected folders:                                                         │
│  • src/ (156 files, 2.4 MB)                                              │
│  • docs/ (24 files, 180 KB)                                              │
│  • config/ (8 files, 32 KB)                                              │
│  • assets/ (67 files, 8.2 MB)                                            │
│  • tests/ (43 files, 256 KB)                                             │
│                                                                             │
│  Exclusions:                                                               │
│  • *.log, *.tmp, .DS_Store, __pycache__/, *.pyc, *.env                   │
│                                                                             │
│  Summary:                                                                  │
│  • Files to copy: 298 files (11.1 MB)                                    │
│  • Files to skip: 16,457 files (548 MB)                                  │
│  • Estimated time: 2-3 seconds                                           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Enter] Start Sync • [b] Back to edit • [s] Save as SyncFile • [q] Quit   │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- **Complete summary** of all sync settings
- **File count and size estimates**
- **Clear action buttons** with keyboard shortcuts
- **Option to save configuration** as a SyncFile for future use
- **Back navigation** to modify any previous settings

#### Screen 8: Execution & Progress Monitoring
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Sync Progress                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Synchronizing: /home/user/projects → /backup/projects                     │
│                                                                             │
│  Progress: ████████████████████░░░░░ 67% (198/298 files)                  │
│                                                                             │
│  Current: src/components/Header.tsx                                        │
│                                                                             │
│  Completed:                                                                │
│  ✓ src/                     156/156 files    2.4 MB                       │
│  ✓ docs/                     24/24 files     180 KB                       │
│  ✓ config/                   8/8 files       32 KB                        │
│  ⏳ assets/                  10/67 files      1.2 MB                       │
│  ⏳ tests/                   0/43 files       0 MB                         │
│                                                                             │
│  Transferred: 3.8 MB of 11.1 MB                                           │
│  Speed: 1.2 MB/s • Elapsed: 00:03 • Remaining: 00:06                      │
│                                                                             │
│  Skipped: 16,457 files (matched exclusion patterns)                       │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Ctrl+C] Cancel • [Space] Pause • [l] View logs                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Features:**
- **Real-time progress bar** with percentage and file count
- **Current file being processed**
- **Per-folder progress breakdown**
- **Transfer speed and time estimates**  
- **Pause/resume functionality**
- **Access to detailed logs**
- **Cancellation with confirmation**

### Technical Implementation Requirements

#### Type State Pattern Architecture

**Design Philosophy**: Use the **Type State Pattern** (popularized in Rust) to create compile-time guarantees about wizard flow and prevent invalid state transitions.

**Benefits:**
- **Compile-time Safety**: Impossible states become unrepresentable in the type system
- **Clear State Transitions**: Each state can only transition to valid next states
- **Type-safe Data Access**: Screen-specific data is only accessible when in that screen's state
- **Maintainable Code**: Adding new screens requires explicit handling of all transitions
- **Runtime Performance**: Zero-cost state validation - all checks happen at compile time

**Go Implementation Strategy:**
```go
// Type state pattern using Go generics and interfaces
type WizardState[T any] interface {
    GetCommonData() *CommonWizardData
    GetSpecificData() T
    Render(styles *Styles) string
    HandleInput(key tea.KeyMsg) WizardTransition
}

// Each screen state has its own type
type WelcomeState struct {
    common     *CommonWizardData
    selectedMode SyncMode
}

type SourceDirectoryState struct {
    common        *CommonWizardData
    directoryTree *DirectoryTree
    currentPath   string
}

type DestinationDirectoryState struct {
    common         *CommonWizardData
    directoryTree  *DirectoryTree
    currentPath    string
    createIfMissing bool
    // Source path is guaranteed to exist by type system
    sourcePath     string // Carried from previous state
}

// Transitions are explicitly typed
type WizardTransition interface {
    Apply() WizardState[any]
}

type ToSourceDirectory struct {
    mode SyncMode
}

func (t ToSourceDirectory) Apply() WizardState[any] {
    return &SourceDirectoryState{
        common: &CommonWizardData{
            selectedMode: t.mode,
        },
        directoryTree: NewDirectoryTree("/"),
    }
}

// Invalid transitions are impossible to create
```

**Common Wizard Problems Solved:**

1. **The "Null/Empty Field" Problem**:
   ```go
   // ❌ TRADITIONAL: Runtime null checks everywhere
   if wizard.destinationPath == "" {
       return errors.New("destination not set")
   }
   
   // ✅ TYPE STATE: Impossible to access unset fields
   func (s *DestinationDirectoryState) GetDestinationPath() string {
       return s.destinationPath // Always set by type construction
   }
   ```

2. **The "Impossible State" Problem**:
   ```go
   // ❌ TRADITIONAL: Can be in progress AND complete simultaneously
   type BadWizardState struct {
       inProgress bool
       complete   bool
       error      error  // What if all three are true?
   }
   
   // ✅ TYPE STATE: Mutually exclusive states
   type ProgressState struct { startTime time.Time }
   type CompleteState struct { result SyncResult }
   type ErrorState struct { error error }
   ```

3. **The "Screen Validation" Problem**:
   ```go
   // ❌ TRADITIONAL: Runtime validation required everywhere
   func canGoToNextScreen(current ScreenType) bool {
       switch current {
       case SourceScreen:
           return wizard.sourcePath != ""
       case DestScreen: 
           return wizard.sourcePath != "" && wizard.destPath != ""
       // ... complex validation logic
       }
   }
   
   // ✅ TYPE STATE: Transitions encode validation
   func (s *SourceDirectoryState) SelectDirectory(path string) (*DestinationDirectoryState, error) {
       if !isValidDirectory(path) {
           return nil, errors.New("invalid directory")
       }
       return &DestinationDirectoryState{
           sourcePath: path, // Guaranteed valid by construction
           directoryTree: NewDirectoryTree("/"),
       }, nil
   }
   ```

4. **The "Back Button" Problem**:
   ```go
   // ❌ TRADITIONAL: Lose data when going back
   func goBack() {
       wizard.currentScreen--
       // Data might be lost or inconsistent
   }
   
   // ✅ TYPE STATE: Preserve data in transition history
   type DestinationDirectoryState struct {
       sourcePath    string           // From previous state
       previousState *SourceDirectoryState  // Can reconstruct
   }
   ```

**State Flow Guarantees:**
- **Welcome → Source**: Must have selected sync mode
- **Source → Destination**: Must have valid source directory path
- **Destination → Options**: Must have both source and destination paths
- **Options → Filters**: Must have valid sync configuration
- **Filters → Exclusions**: Must have folder selections
- **Exclusions → Preview**: Must have exclusion patterns (can be empty)
- **Preview → Progress**: Must have confirmed configuration
- **Progress → Complete**: Must have finished sync operation

**Type Safety Examples:**
```go
// ✅ VALID: Can only access source path after it's been set
func (s *DestinationDirectoryState) GetSourcePath() string {
    return s.sourcePath // Guaranteed to exist
}

// ❌ INVALID: Cannot access destination path from source state
func (s *SourceDirectoryState) GetDestinationPath() string {
    // This method doesn't exist - compile error!
}

// ✅ VALID: Sync can only start from preview state with complete config
func (s *PreviewState) StartSync() *ProgressState {
    return &ProgressState{
        config: s.BuildSyncConfig(), // All data guaranteed present
        startTime: time.Now(),
    }
}

// ❌ INVALID: Cannot start sync from incomplete state
func (s *WelcomeState) StartSync() *ProgressState {
    // This method doesn't exist - compile error!
}
```

#### Bubble Tea Components
- **Navigation**: Consistent keyboard navigation across all screens
- **State Management**: Type-safe state machine using type state pattern
- **Data Validation**: Compile-time validation with runtime error messages  
- **Responsive Layout**: Adapts to different terminal sizes (minimum 80x24)
- **Accessibility**: Screen reader compatible, clear visual hierarchy

#### Directory Tree Browser
- **Lazy Loading**: Only load directory contents when needed for performance
- **Caching**: Cache directory structure to avoid repeated filesystem calls
- **Error Handling**: Graceful handling of permission denied, symlinks, etc.
- **Search**: Fast substring search across directory tree
- **Filtering**: Hide/show hidden files, filter by type

#### File System Integration  
- **Permission Checking**: Validate read/write access before starting sync
- **Space Calculation**: Accurate disk space requirements calculation
- **Path Validation**: Handle edge cases (long paths, special characters, etc.)
- **Real-time Updates**: Refresh file counts if source changes during wizard

#### Progress Monitoring
- **Non-blocking UI**: Progress updates don't freeze the interface
- **Detailed Logging**: Structured logs for debugging and audit trails
- **Error Recovery**: Handle partial failures gracefully
- **Performance Metrics**: Track and display transfer speeds, file counts

### Integration with Existing Features

#### SyncFile Generation
Users can save wizard configurations as SyncFile format:
```bash
# Generated SyncFile from wizard
SYNC /home/user/projects /backup/projects
MODE one-way
EXCLUDE *.log
EXCLUDE *.tmp
EXCLUDE .DS_Store
EXCLUDE __pycache__/
EXCLUDE *.pyc
EXCLUDE *.env
ONLY src/
ONLY docs/
ONLY config/
ONLY assets/
ONLY tests/
DRYRUN false
GITIGNORE true
VERBOSE true
```

#### Plan Mode Integration
Advanced users can generate plans from wizard output:
```bash
# Wizard saves temp config, then generates plan
sync-tools wizard --save-config temp.syncfile
sync-tools syncfile temp.syncfile --plan review.plan
```

## Plan-Based Workflow (Advanced Users)

### Advanced Plan Generation

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

### Phase 1: Wizard UI Success Metrics
1. **User Adoption**: 50% of new users try wizard mode within first week
2. **Completion Rate**: 80% of users who start wizard complete the full flow
3. **Error Reduction**: 60% reduction in sync misconfigurations compared to CLI-only usage
4. **User Satisfaction**: 4.5+ star rating for wizard experience
5. **Learning Curve**: Average time to first successful sync < 5 minutes for new users

### Type State Pattern Benefits Metrics
1. **Development Velocity**: 40% reduction in wizard-related bug reports after implementation
2. **Code Maintainability**: 100% test coverage achievable due to impossible states being unrepresentable
3. **Onboarding Time**: New developers can contribute to wizard screens 50% faster due to compile-time constraints
4. **Runtime Errors**: 90% reduction in wizard state-related runtime errors
5. **Refactoring Safety**: Major wizard refactors can be done with 100% confidence due to type safety

### Phase 2: Plan-Based Success Metrics  
1. **Power User Adoption**: 30% of advanced users adopt plan-based mode within 3 months
2. **Reduced Sync Errors**: 50% reduction in accidental overwrites during deployments
3. **Time Saved**: Average 20% reduction in sync-related troubleshooting
4. **Workflow Integration**: 40% of plan users save and reuse configurations

### Overall Success Metrics
1. **User Base Growth**: 25% increase in active users within 6 months
2. **Feature Usage**: 70% of syncs use interactive features (wizard or plan-based)
3. **Support Reduction**: 40% decrease in sync-related support tickets
4. **Retention**: 85% user retention rate for interactive features after 30 days

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

### Phase 1: Interactive Wizard (MVP)
**Target**: User-friendly sync setup with type-safe state management

**Core Features:**
- Type state pattern wizard architecture with compile-time guarantees
- Bubble Tea wizard UI with all 8 screens
- One-way sync mode only
- Source/destination directory browser with lazy loading
- Folder selection with file count/size display
- Basic exclusion pattern management
- Preview and confirmation screen
- Real-time progress monitoring
- Save configuration as SyncFile

**Technical Requirements:**
- **Type State Implementation**: Complete state machine with typed transitions
- **Bubble Tea Integration**: Screen components with type-safe state binding
- **Directory Tree Browser**: Lazy-loading filesystem browser with caching
- **File System Scanning**: Async file counting and size calculation
- **Progress Monitoring**: Non-blocking rsync integration with pause/resume
- **State Persistence**: SyncFile generation from type-safe wizard state

**Implementation Architecture:**
```go
// Core wizard driver with type-safe state management
type WizardModel struct {
    currentState WizardState[any]
    history      []WizardState[any] // For back navigation
    styles       *Styles
    windowSize   tea.WindowSizeMsg
}

// State machine handles all transitions
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    transition := m.currentState.HandleInput(msg)
    if transition != nil {
        m.history = append(m.history, m.currentState)
        m.currentState = transition.Apply()
    }
    return m, nil
}

// Rendering delegates to current state
func (m WizardModel) View() string {
    return m.currentState.Render(m.styles)
}
```

### Phase 2: Plan-Based Workflow (Advanced)
**Target**: Power users who need fine-grained control

**Core Features:**
- Plan generation from wizard output
- Text-based plan editing with visual aliases (`<<`, `>>`, `<>`)
- Plan execution with validation
- Basic conflict detection and resolution
- Editor integration ($EDITOR support)
- Configuration layering (config files + CLI overrides)

**Integration:**
- Wizard → Plan workflow
- SyncFile → Plan generation
- Plan templates and reuse

### Phase 3: Two-Way Sync & Advanced Features
**Target**: Complex synchronization scenarios

**Core Features:**
- Two-way sync mode in wizard
- Bidirectional conflict resolution
- Interactive merge tool integration
- Advanced plan filtering (change types)
- Audit logging and compliance features
- Bulk operations and plan templates

**Advanced Integration:**
- Git integration for merge tools
- Custom conflict resolution strategies
- Performance optimizations for large directories

## Example Use Cases

### 1. New User Project Backup (Wizard Workflow)
Casual user wants to backup their project to external drive:

**Scenario**: Developer backing up project for the first time

**Workflow:**
```bash
# Launch wizard
sync-tools wizard

# Screen 1: Welcome & Mode Selection
# → User selects "One-way sync" 

# Screen 2: Source Directory Selection  
# → User navigates to /home/user/my-project
# → Wizard shows folder structure, user selects project root

# Screen 3: Destination Directory Selection
# → User navigates to /media/backup-drive/projects
# → Selects ☑ "Create destination directory if it doesn't exist"

# Screen 4: Sync Options
# → Keeps defaults: ☑ Dry run, ☑ Verbose, ☑ Use .gitignore
# → Unchecks "Delete files in destination that don't exist in source" (safe default)

# Screen 5: Directory Tree Filter Selection
# → Wizard shows project folders with file counts:
#   ☑ src/ (245 files, 1.2 MB)
#   ☑ docs/ (15 files, 89 KB)  
#   ☑ config/ (5 files, 12 KB)
#   ☐ node_modules/ (8,431 files, 156 MB)  [auto-unchecked]
#   ☐ .git/ (423 files, 12 MB)             [auto-unchecked] 
#   ☐ build/ (89 files, 2.1 MB)           [auto-unchecked]
#   ☑ assets/ (34 files, 8.9 MB)
# → User keeps smart defaults

# Screen 6: Exclusion Patterns
# → Pre-populated with: *.log, *.tmp, .DS_Store, __pycache__/, *.pyc
# → User adds: *.env (clicks quick-add button)

# Screen 7: Preview & Confirmation
# → Shows: "299 files (10.3 MB) will be copied"
# → User clicks "Start Sync"

# Screen 8: Progress
# → Real-time progress bar
# → Shows current file being copied
# → Completes successfully

# Wizard offers to save config for future use
# → Saves as ProjectBackup.syncfile
```

**User Benefits:**
- **No command-line knowledge required**
- **Visual feedback** shows exactly what will be copied
- **Smart defaults** prevent common mistakes
- **File size awareness** helps user understand storage impact
- **Reusable configuration** for future backups

### 2. Advanced Deployment with Plan Review (Hybrid Workflow)
DevOps engineer combines wizard convenience with plan-based precision:

**Workflow:**
```bash
# Start with wizard for convenience
sync-tools wizard --source ./app --destination /var/www/production

# Wizard generates initial SyncFile with user selections
# → Saves as production-deploy.syncfile

# Generate plan for review
sync-tools syncfile production-deploy.syncfile --plan deploy-review.plan --exclude-changes unchanged

# Edit plan file to:
# - Skip large asset files that haven't changed
# - Pull critical config from production (>> commands)
# - Add comments for manual verification steps

# Execute with audit trail
sync-tools sync --apply-plan deploy-review.plan --audit-log deploy-$(date +%Y%m%d).log
```

**Benefits:**
- **Best of both worlds**: Wizard convenience + plan precision
- **Audit compliance**: Full trail of what was deployed
- **Team collaboration**: Plan files can be reviewed and version-controlled

### 3. Legacy Deployment with SyncFile Base
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

## Architectural Decision: Type State Pattern

### Comparison with Alternative Approaches

#### Traditional State Machine Approach
```go
// ❌ TRADITIONAL: Single state struct with validation everywhere
type WizardState struct {
    currentScreen     ScreenType
    sourcePath        string  // Could be empty
    destinationPath   string  // Could be empty  
    syncOptions       SyncOptions // Could be invalid
    selectedFolders   []string    // Could be empty when required
}

func (w *WizardState) CanProceed() bool {
    switch w.currentScreen {
    case SourceScreen:
        return w.sourcePath != ""
    case DestScreen:
        return w.sourcePath != "" && w.destinationPath != ""
    // ... 50+ lines of validation logic
    }
}
```

**Problems:**
- Runtime validation required everywhere
- Impossible states are representable (sourcePath set but on WelcomeScreen)
- Easy to forget validation checks
- Testing requires covering all invalid state combinations
- Back navigation loses data or creates inconsistencies

#### Type State Pattern Approach
```go
// ✅ TYPE STATE: Each state can only contain valid data
type SourceDirectoryChosen struct {
    path          string           // Always valid (constructor validates)
    directoryTree *DirectoryTree   // Always initialized
    syncMode      SyncMode         // Inherited from previous state
}

type DestinationDirectoryChosen struct {
    sourcePath      string         // Guaranteed from previous state
    destinationPath string         // Guaranteed valid by constructor
    createIfMissing bool           // Explicit choice made
    syncMode        SyncMode       // Carried through states
}

func NewDestinationDirectoryChosen(
    prev *SourceDirectoryChosen, 
    destPath string,
    createIfMissing bool,
) (*DestinationDirectoryChosen, error) {
    if !isValidDirectory(destPath) {
        return nil, errors.New("invalid destination")
    }
    return &DestinationDirectoryChosen{
        sourcePath:      prev.path,          // Type safety ensures this exists
        destinationPath: destPath,
        createIfMissing: createIfMissing,
        syncMode:        prev.syncMode,
    }, nil
}
```

**Benefits:**
- No runtime validation needed (compile-time guarantees)
- Impossible states cannot be constructed
- Exhaustive pattern matching forces handling all cases
- Back navigation preserves all data through state history
- Unit tests only need to cover valid transitions

### Implementation Complexity Comparison

| Aspect | Traditional | Type State | Winner |
|--------|-------------|------------|--------|
| **Lines of Code** | ~800 LOC | ~1200 LOC | Traditional |
| **Runtime Bugs** | High risk | Near zero | **Type State** |
| **Compile-time Safety** | None | Complete | **Type State** |
| **Refactoring Risk** | High | Very Low | **Type State** |
| **New Developer Onboarding** | Weeks | Days | **Type State** |
| **Test Coverage Required** | >95% for safety | <50% (types prevent bugs) | **Type State** |
| **Documentation Needs** | Extensive | Self-documenting | **Type State** |

### Trade-offs Analysis

**Type State Pattern Costs:**
- 50% more initial implementation time
- Steeper learning curve for Go developers unfamiliar with the pattern
- More complex type definitions
- Requires disciplined approach to state transitions

**Type State Pattern Benefits:**
- 90% fewer runtime bugs related to state management
- 100% confidence during refactoring
- Self-documenting state flow through types
- Impossible to ship wizard with invalid state transitions
- New features require explicit handling of all existing states (forces design consideration)

**Decision:** The benefits heavily outweigh the costs, especially for a complex wizard with 8+ screens where state consistency is critical for user experience.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex UI confuses users | Low adoption | Provide simple mode with sensible defaults |
| Type state pattern learning curve | Developer productivity | Provide comprehensive examples and documentation |
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
# Wizard UI settings
default_mode = "wizard"           # Default to wizard for new users
wizard_theme = "default"          # UI theme: default, compact, minimal
min_terminal_size = "80x24"       # Minimum terminal size for wizard
directory_scan_depth = 3          # How deep to scan directories initially
file_count_threshold = 10000      # Warn when directories have many files

# Directory browser
show_hidden_files = false         # Show .files in directory browser
show_file_sizes = true           # Display file/folder sizes
cache_directory_scans = true     # Cache directory contents for performance

# Smart defaults
auto_exclude_patterns = ["node_modules/", ".git/", "*.log", "*.tmp", ".DS_Store"]
auto_dry_run = true              # Start with dry-run enabled by default
auto_verbose = true              # Enable verbose output by default

# Plan-based settings (advanced users)
default_editor = "vim"
conflict_strategy = "newest-wins"
show_size_threshold = "1MB"
audit_log_dir = "~/.sync-tools/audit/"

# Performance
max_concurrent_scans = 4         # Parallel directory scanning
scan_timeout_seconds = 30       # Timeout for large directory scans
```

### Command Reference Summary

#### Wizard Commands (Phase 1)
```bash
# Launch wizard
sync-tools wizard
sync-tools sync --wizard

# Wizard with pre-filled options  
sync-tools wizard --source ./src
sync-tools sync --wizard --source ./project --destination ./backup

# Wizard shortcuts
sync-tools wizard --help             # Show wizard help
sync-tools wizard --theme compact    # Use compact UI theme
```

#### Plan Commands (Phase 2)
```bash
# Generate plan
sync-tools sync --plan <file>
sync-tools syncfile MySyncFile --plan <file>

# Interactive plan mode
sync-tools sync --interactive

# Apply plan
sync-tools sync --apply-plan <file>

# Dry run plan
sync-tools sync --apply-plan <file> --dry-run

# Plan with audit
sync-tools sync --apply-plan <file> --audit-log <log-file>
```

#### Integration Commands
```bash
# Save wizard config as SyncFile
sync-tools wizard --save-config MySyncFile

# Generate plan from wizard config
sync-tools wizard --source ./src --destination ./dest --plan review.plan

# Hybrid workflow
sync-tools wizard --generate-plan    # Wizard → Plan workflow
sync-tools syncfile MySyncFile --wizard --plan review.plan  # SyncFile → Wizard → Plan
```