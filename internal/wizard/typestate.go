package wizard

// Type State Pattern Implementation for Wizard State Management
// This ensures compile-time safety by making invalid state transitions impossible.

// Common interfaces that all states implement
type WizardStateBase interface {
	GetCurrentScreen() WizardScreen
	CanGoBack() bool
	Render() string
}

// State transition interfaces - each state only allows valid transitions
type FromWelcome interface {
	WizardStateBase
	SelectSyncMode(mode SyncMode) *SourceDirectorySelectionState
}

type FromSourceDirectorySelection interface {
	WizardStateBase
	GoBack() *WelcomeState
	SetSourceDirectory(path string) *DestinationDirectorySelectionState
}

type FromDestinationDirectorySelection interface {
	WizardStateBase
	GoBack() *SourceDirectorySelectionState
	SetDestinationDirectory(path string, createIfNotExists bool) *SyncOptionsState
}

type FromSyncOptions interface {
	WizardStateBase
	GoBack() *DestinationDirectorySelectionState
	ConfigureOptions(config SyncOptionsConfig) *DirectoryFilterState
}

type FromDirectoryFilter interface {
	WizardStateBase
	GoBack() *SyncOptionsState
	SetSelectedFolders(folders map[string]bool) *ExclusionPatternsState
}

type FromExclusionPatterns interface {
	WizardStateBase
	GoBack() *DirectoryFilterState
	SetExclusionPatterns(patterns []string) *PreviewState
}

type FromPreview interface {
	WizardStateBase
	GoBack() *ExclusionPatternsState
	StartSync() *ProgressState
	SaveAsSyncFile() (string, error) // Returns path to created SyncFile
}

type FromProgress interface {
	WizardStateBase
	// Progress state cannot go back - only forward to completion
	WaitForCompletion() *CompletedState
	CancelSync() (*CompletedState, error)
}

type FromCompleted interface {
	WizardStateBase
	// Completed state is terminal - wizard exits
	Exit() bool
}

// Concrete state types implementing the Type State Pattern

// WelcomeState - Entry point, can only proceed to source directory selection
type WelcomeState struct {
	syncMode SyncMode
}

// SourceDirectorySelectionState - Has sync mode, can set source directory
type SourceDirectorySelectionState struct {
	syncMode SyncMode
	directoryTree *DirectoryTree
}

// DestinationDirectorySelectionState - Has sync mode and source, can set destination
type DestinationDirectorySelectionState struct {
	syncMode      SyncMode
	sourcePath    string
	directoryTree *DirectoryTree
}

// SyncOptionsState - Has sync mode, source, and destination, can configure options
type SyncOptionsState struct {
	syncMode          SyncMode
	sourcePath        string
	destinationPath   string
	createDestination bool
}

// DirectoryFilterState - Has all above plus sync options, can select folders
type DirectoryFilterState struct {
	syncMode          SyncMode
	sourcePath        string
	destinationPath   string
	createDestination bool
	syncOptions       SyncOptionsConfig
	folderStats       map[string]FolderInfo
}

// ExclusionPatternsState - Has all above plus selected folders, can set exclusions
type ExclusionPatternsState struct {
	syncMode          SyncMode
	sourcePath        string
	destinationPath   string
	createDestination bool
	syncOptions       SyncOptionsConfig
	selectedFolders   map[string]bool
}

// PreviewState - Has complete configuration, can start sync or save as SyncFile
type PreviewState struct {
	syncMode          SyncMode
	sourcePath        string
	destinationPath   string
	createDestination bool
	syncOptions       SyncOptionsConfig
	selectedFolders   map[string]bool
	exclusionPatterns []string
	estimatedFiles    int
	estimatedSize     int64
}

// ProgressState - Sync is running, shows progress
type ProgressState struct {
	config            *CompleteWizardConfig
	filesProcessed    int
	filesTotal        int
	currentFile       string
	bytesTransferred  int64
	bytesTotal        int64
	transferSpeed     float64
	canCancel         bool
}

// CompletedState - Sync is finished, shows results
type CompletedState struct {
	config       *CompleteWizardConfig
	successful   bool
	filesSync    int
	errors       []error
	syncDuration string
}

// Configuration types

// SyncOptionsConfig holds all sync behavior configuration
type SyncOptionsConfig struct {
	DryRun             bool
	Verbose            bool
	DeleteExtraFiles   bool
	UseGitignore       bool
	FollowSymlinks     bool
	PreserveTimestamps bool
	ExcludeHiddenDirs  bool
	ParallelTransfers  int
	TransferTimeout    int
}

// CompleteWizardConfig holds the final wizard configuration
type CompleteWizardConfig struct {
	SyncMode          SyncMode
	SourcePath        string
	DestinationPath   string
	CreateDestination bool
	SyncOptions       SyncOptionsConfig
	SelectedFolders   map[string]bool
	ExclusionPatterns []string
}

// Factory functions for creating initial states

// NewWelcomeState creates the initial wizard state
func NewWelcomeState() *WelcomeState {
	return &WelcomeState{
		syncMode: OneWaySync, // Default selection
	}
}

// GetCompleteConfig extracts the complete configuration from any state that has it
func (s *PreviewState) GetCompleteConfig() *CompleteWizardConfig {
	return &CompleteWizardConfig{
		SyncMode:          s.syncMode,
		SourcePath:        s.sourcePath,
		DestinationPath:   s.destinationPath,
		CreateDestination: s.createDestination,
		SyncOptions:       s.syncOptions,
		SelectedFolders:   s.selectedFolders,
		ExclusionPatterns: s.exclusionPatterns,
	}
}

func (s *ProgressState) GetCompleteConfig() *CompleteWizardConfig {
	return s.config
}

func (s *CompletedState) GetCompleteConfig() *CompleteWizardConfig {
	return s.config
}

// Default sync options
func DefaultSyncOptions() SyncOptionsConfig {
	return SyncOptionsConfig{
		DryRun:             true,  // Safe default
		Verbose:            true,
		DeleteExtraFiles:   false, // Safe default
		UseGitignore:       true,
		FollowSymlinks:     false, // Safe default
		PreserveTimestamps: true,
		ExcludeHiddenDirs:  true,
		ParallelTransfers:  4,
		TransferTimeout:    300,
	}
}