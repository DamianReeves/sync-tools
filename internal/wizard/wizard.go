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
		initialState = DestinationSelectionState{
			SourcePath: config.PrefilledSource,
			CurrentPath: ".",
			Directories: []DirectoryInfo{}, // Will be populated by the UI
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

// Run starts the interactive wizard
func (w *Wizard) Run() error {
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