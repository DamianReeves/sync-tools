package wizard

import (
	"time"
)

// WizardScreen represents the current screen in the wizard flow
type WizardScreen int

const (
	WelcomeScreen WizardScreen = iota
	SourceDirectoryScreen
	DestinationDirectoryScreen
	SyncOptionsScreen
	DirectoryFilterScreen
	ExclusionPatternsScreen
	PreviewScreen
	ProgressScreen
	CompletedScreen
)

// SyncMode represents the synchronization mode
type SyncMode int

const (
	OneWaySync SyncMode = iota
	TwoWaySync // Future release
)

// String returns the display name for the sync mode
func (s SyncMode) String() string {
	switch s {
	case OneWaySync:
		return "One-way sync (source → destination)"
	case TwoWaySync:
		return "Two-way sync (source ↔ destination)"
	default:
		return "Unknown"
	}
}

// Description returns a detailed description of the sync mode
func (s SyncMode) Description() string {
	switch s {
	case OneWaySync:
		return "One-way sync copies files from source to destination only.\nFiles in destination that don't exist in source will be preserved."
	case TwoWaySync:
		return "Two-way sync keeps both directories synchronized.\nConflicts will be resolved using configured strategy."
	default:
		return ""
	}
}

// WizardState holds all the configuration collected through the wizard
type WizardState struct {
	// Navigation
	CurrentScreen WizardScreen
	PreviousScreen WizardScreen

	// Configuration
	Mode                     SyncMode
	SourcePath               string
	DestinationPath          string
	CreateDestinationDir     bool

	// Options
	DryRun                   bool
	Verbose                  bool
	DeleteExtraFiles         bool
	UseGitignore             bool
	FollowSymlinks           bool
	PreserveTimestamps       bool
	ExcludeHiddenDirs        bool
	ParallelTransfers        int
	TransferTimeout          int

	// Directory filtering
	SelectedFolders          map[string]bool    // folder path -> selected
	FolderStats              map[string]FolderInfo
	
	// Exclusion patterns
	ExclusionPatterns        []string

	// Directory browser state
	SourceDirectoryTree      *DirectoryTree
	DestinationDirectoryTree *DirectoryTree

	// Progress tracking
	SyncInProgress           bool
	SyncComplete             bool
	SyncError                error
	FilesProcessed           int
	FilesTotal               int
	CurrentFile              string
	BytesTransferred         int64
	BytesTotal               int64
	SyncStartTime            time.Time
	TransferSpeed            float64

	// UI state
	WindowWidth              int
	WindowHeight             int
	ShowHelp                 bool
}

// FolderInfo contains metadata about a directory
type FolderInfo struct {
	Name       string
	Path       string
	FileCount  int
	Size       int64
	SizeString string
	IsSelected bool
}

// NewWizardState creates a new wizard state with sensible defaults
func NewWizardState() *WizardState {
	return &WizardState{
		CurrentScreen:         WelcomeScreen,
		Mode:                 OneWaySync,
		DryRun:               true,  // Safe default
		Verbose:              true,
		UseGitignore:         true,
		ExcludeHiddenDirs:    true,
		ParallelTransfers:    4,
		TransferTimeout:      300,
		SelectedFolders:      make(map[string]bool),
		FolderStats:          make(map[string]FolderInfo),
		ExclusionPatterns:    []string{
			"*.log",
			"*.tmp", 
			".DS_Store",
			"__pycache__/",
			"*.pyc",
		},
	}
}

// CanGoBack returns true if the user can navigate back from the current screen
func (s *WizardState) CanGoBack() bool {
	return s.CurrentScreen > WelcomeScreen && s.CurrentScreen < ProgressScreen
}

// CanGoNext returns true if the user can navigate forward from the current screen
func (s *WizardState) CanGoNext() bool {
	switch s.CurrentScreen {
	case WelcomeScreen:
		return true
	case SourceDirectoryScreen:
		return s.SourcePath != ""
	case DestinationDirectoryScreen:
		return s.DestinationPath != ""
	case SyncOptionsScreen, DirectoryFilterScreen, ExclusionPatternsScreen:
		return true
	case PreviewScreen:
		return true
	default:
		return false
	}
}

// GoBack navigates to the previous screen
func (s *WizardState) GoBack() {
	if !s.CanGoBack() {
		return
	}
	
	s.PreviousScreen = s.CurrentScreen
	s.CurrentScreen--
}

// GoNext navigates to the next screen
func (s *WizardState) GoNext() {
	if !s.CanGoNext() {
		return
	}
	
	s.PreviousScreen = s.CurrentScreen
	s.CurrentScreen++
}

// GetSelectedFolderCount returns the number of selected folders and their total size
func (s *WizardState) GetSelectedFolderCount() (int, int, int64) {
	count := 0
	totalFiles := 0
	totalSize := int64(0)
	
	for path, selected := range s.SelectedFolders {
		if selected {
			count++
			if info, exists := s.FolderStats[path]; exists {
				totalFiles += info.FileCount
				totalSize += info.Size
			}
		}
	}
	
	return count, totalFiles, totalSize
}

// GetExcludedFolderCount returns the number of unselected folders and their total size
func (s *WizardState) GetExcludedFolderCount() (int, int, int64) {
	count := 0
	totalFiles := 0
	totalSize := int64(0)
	
	for path, selected := range s.SelectedFolders {
		if !selected {
			count++
			if info, exists := s.FolderStats[path]; exists {
				totalFiles += info.FileCount
				totalSize += info.Size
			}
		}
	}
	
	return count, totalFiles, totalSize
}