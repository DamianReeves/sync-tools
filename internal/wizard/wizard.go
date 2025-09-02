package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Wizard represents the interactive sync wizard
type Wizard struct {
	config *Config
	model  *WizardModel
}

// New creates a new wizard with the given configuration
func New(config *Config) *Wizard {
	if config == nil {
		config = &Config{}
	}

	// Initialize with appropriate starting state
	var initialState WizardState = InitialState{}
	
	// If source is pre-filled, start at destination selection
	if config.PrefilledSource != "" {
		browser := NewDirectoryBrowser(".")
		initialState = DestinationSelectionState{
			SourcePath:  config.PrefilledSource,
			CurrentPath: ".",
			Directories: []DirectoryInfo{},
			Browser:     browser,
		}
	}

	return &Wizard{
		config: config,
		model: &WizardModel{
			CurrentState: initialState,
			Config:       config,
		},
	}
}

// Run starts the interactive wizard or runs in test mode
func (w *Wizard) Run() error {
	// Check if running in test mode
	if w.config.TestMode && w.config.TestOptions != nil {
		return w.runTestMode()
	}
	
	// Create the Bubble Tea model
	teaModel := NewBubbleTeaModel(w.model)
	
	// Start the Bubble Tea program
	p := tea.NewProgram(teaModel, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run wizard: %w", err)
	}

	// Check if the wizard completed successfully
	if wizModel, ok := finalModel.(*BubbleTeaModel); ok {
		if wizModel.Error != "" {
			return fmt.Errorf("wizard error: %s", wizModel.Error)
		}
		
		// Check if we reached the complete state
		if _, isComplete := wizModel.Model.CurrentState.(CompleteState); !isComplete {
			// User cancelled or quit
			return nil
		}
	}

	return nil
}

// runTestMode executes the wizard in non-interactive test mode
func (w *Wizard) runTestMode() error {
	testOpts := w.config.TestOptions
	
	// Create a progress state with the test configuration
	progressState := &ProgressState{
		SourcePath:      testOpts.SourceDir,
		DestinationPath: testOpts.DestinationDir,
		SyncOptions: SyncOptionsState{
			Mode:             testOpts.Mode,
			DryRun:           testOpts.DryRun,
			HiddenDirs:       true, // Default
			UseGitIgnore:     testOpts.EnableGitIgnore,
			ConflictStrategy: "newest", // Default
		},
		Patterns: make([]ExclusionPattern, 0),
	}
	
	// Add exclusion patterns
	for _, pattern := range testOpts.ExclusionPatterns {
		progressState.Patterns = append(progressState.Patterns, ExclusionPattern{
			Pattern: pattern,
			Valid:   true,
			Source:  "test",
		})
	}
	
	// Generate SyncFile content
	syncFileContent, err := GenerateSyncFile(progressState)
	if err != nil {
		return fmt.Errorf("failed to generate SyncFile: %w", err)
	}
	
	// Output the generated SyncFile content for the BDD test to capture
	fmt.Print(syncFileContent)
	
	return nil
}