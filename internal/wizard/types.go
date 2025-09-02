package wizard

import tea "github.com/charmbracelet/bubbletea"

// WizardState represents the type-safe state of the wizard
// This implements the Type State Pattern for compile-time safety
type WizardState interface {
	// Type method ensures compile-time type safety
	wizardState()
}

// Modal represents a UI overlay that can be displayed on top of a base state
type Modal interface {
	// BaseState returns the underlying wizard state this modal overlays
	BaseState() WizardState
	// Render composes the modal content with the base state content
	Render(baseContent string) string
	// HandleInput processes input specific to this modal
	HandleInput(msg tea.KeyMsg) (Modal, tea.Cmd, bool) // returns (newModal, cmd, handled)
	// IsComplete returns whether the modal should be closed and optionally an updated base state
	IsComplete() (bool, WizardState)
	// ModalType returns a string identifier for this modal type
	ModalType() string
}

// WizardUIState contains both domain state and UI overlay state
type WizardUIState struct {
	DomainState WizardState
	ModalStack  []Modal
}

// ActiveModal returns the top modal from the stack, or nil if no modals are active
func (ui *WizardUIState) ActiveModal() Modal {
	if len(ui.ModalStack) == 0 {
		return nil
	}
	return ui.ModalStack[len(ui.ModalStack)-1]
}

// PushModal adds a new modal to the stack
func (ui *WizardUIState) PushModal(modal Modal) {
	ui.ModalStack = append(ui.ModalStack, modal)
}

// PopModal removes the top modal from the stack
func (ui *WizardUIState) PopModal() Modal {
	if len(ui.ModalStack) == 0 {
		return nil
	}

	modal := ui.ModalStack[len(ui.ModalStack)-1]
	ui.ModalStack = ui.ModalStack[:len(ui.ModalStack)-1]
	return modal
}

// HasModals returns true if there are any active modals
func (ui *WizardUIState) HasModals() bool {
	return len(ui.ModalStack) > 0
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
	SourcePath       string
	DestinationPath  string
	Mode             string
	DryRun           bool
	HiddenDirs       bool
	UseGitIgnore     bool
	ConflictStrategy string
	Editor           *SyncOptionsEditor
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
	Monitor         *ProgressMonitor
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
	CurrentFile      string
	FilesProcessed   int
	TotalFiles       int
	TransferSpeed    string
	ProgressPercent  int
	BytesTransferred int64
	TotalBytes       int64
}

// Config represents wizard configuration
type Config struct {
	PrefilledSource string
	PrefilledMode   string
	TestMode        bool             // When true, runs non-interactively for testing
	TestOptions     *TestModeOptions // Test mode configuration
}

// TestModeOptions configures how the wizard behaves in test mode
type TestModeOptions struct {
	SourceDir         string
	DestinationDir    string
	Mode              string
	ExclusionPatterns []string
	EnableGitIgnore   bool
	DryRun            bool
}

// WizardModel represents the complete wizard state machine
type WizardModel struct {
	CurrentState WizardState
	Config       *Config
	Error        string
}
