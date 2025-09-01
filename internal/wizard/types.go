package wizard

// WizardState represents the type-safe state of the wizard
// This implements the Type State Pattern for compile-time safety
type WizardState interface {
	// Type method ensures compile-time type safety
	wizardState()
}

// InitialState - wizard just started
type InitialState struct{}

func (InitialState) wizardState() {}

// SourceSelectionState - selecting source directory
type SourceSelectionState struct {
	CurrentPath string
	Directories []DirectoryInfo
	Browser     *DirectoryBrowser
}

func (SourceSelectionState) wizardState() {}

// DestinationSelectionState - selecting destination directory
type DestinationSelectionState struct {
	SourcePath  string
	CurrentPath string
	Directories []DirectoryInfo
	Browser     *DirectoryBrowser
}

func (DestinationSelectionState) wizardState() {}

// SyncOptionsState - configuring sync options
type SyncOptionsState struct {
	SourcePath      string
	DestinationPath string
	Mode            string
	DryRun          bool
	HiddenDirs      bool
	UseGitIgnore    bool
	ConflictStrategy string
}

func (SyncOptionsState) wizardState() {}

// ExclusionPatternsState - managing exclusion patterns
type ExclusionPatternsState struct {
	SourcePath      string
	DestinationPath string
	SyncOptions     SyncOptionsState
	Patterns        []ExclusionPattern
}

func (ExclusionPatternsState) wizardState() {}

// DirectoryFilterState - selecting which directories to sync
type DirectoryFilterState struct {
	SourcePath      string
	DestinationPath string
	SyncOptions     SyncOptionsState
	Patterns        []ExclusionPattern
	Directories     []SelectableDirectory
}

func (DirectoryFilterState) wizardState() {}

// ProgressState - showing sync progress
type ProgressState struct {
	SourcePath      string
	DestinationPath string
	SyncOptions     SyncOptionsState
	Patterns        []ExclusionPattern
	Directories     []SelectableDirectory
	Progress        ProgressInfo
}

func (ProgressState) wizardState() {}

// CompleteState - wizard completed
type CompleteState struct {
	SyncFilePath string
	Success      bool
	Error        string
}

func (CompleteState) wizardState() {}

// DirectoryInfo represents a directory entry with metadata
type DirectoryInfo struct {
	Path  string
	Name  string
	Files int
	Size  int64
}

// ExclusionPattern represents a filter pattern
type ExclusionPattern struct {
	Pattern string
	Source  string // "default", "user", ".gitignore", etc.
	Valid   bool
	Error   string
}

// SelectableDirectory represents a directory that can be selected for syncing
type SelectableDirectory struct {
	DirectoryInfo
	Selected bool
}

// ProgressInfo represents current sync progress
type ProgressInfo struct {
	CurrentFile     string
	FilesProcessed  int
	TotalFiles      int
	TransferSpeed   string
	ProgressPercent int
	BytesTransferred int64
	TotalBytes      int64
}

// Config represents wizard configuration
type Config struct {
	PrefilledSource string
	PrefilledMode   string
}

// WizardModel represents the complete wizard state machine
type WizardModel struct {
	CurrentState WizardState
	Config       *Config
	Error        string
}