package wizard

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
)

// TypeStateModel is a Bubble Tea model that uses the Type State Pattern
// This provides compile-time safety while maintaining the existing UI
type TypeStateModel struct {
	// Current state - one of the type-safe states
	currentState WizardStateBase
	
	// UI state
	styles       *Styles
	windowWidth  int
	windowHeight int
	showHelp     bool
	err          error
}

// NewTypeStateModel creates a new wizard model using type-safe states
func NewTypeStateModel() TypeStateModel {
	return TypeStateModel{
		currentState: NewWelcomeState(),
		styles:       NewStyles(),
	}
}

// Init implements the Bubble Tea init method
func (m TypeStateModel) Init() tea.Cmd {
	return nil
}

// Update implements the Bubble Tea update method
func (m TypeStateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMessage(msg)
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View implements the Bubble Tea view method
func (m TypeStateModel) View() string {
	if m.windowWidth < 80 || m.windowHeight < 24 {
		return m.renderMinimumSizeWarning()
	}

	var content string

	// Header
	content += m.renderHeader()
	content += "\n"

	// Main content based on current state
	content += m.renderCurrentState()
	content += "\n"

	// Footer
	content += m.renderFooter()

	// Error display
	if m.err != nil {
		content += "\n"
		content += m.styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error()))
	}

	// Help overlay
	if m.showHelp {
		content += "\n"
		content += m.renderHelpOverlay()
	}

	// Apply border and return
	return m.styles.BorderStyle.
		Width(m.windowWidth - 4).
		Height(m.windowHeight - 2).
		Render(content)
}

// handleKeyMessage processes keyboard input using type-safe state transitions
func (m TypeStateModel) handleKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys available on all screens
	switch msg.String() {
	case "ctrl+c", "q":
		if m.currentState.GetCurrentScreen() != ProgressScreen {
			return m, tea.Quit
		}
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Handle state-specific key input using type assertions
	// This preserves type safety while allowing screen-specific behavior
	switch currentState := m.currentState.(type) {
	case *WelcomeState:
		return m.handleWelcomeKeys(msg, currentState)
	case *SourceDirectorySelectionState:
		return m.handleSourceDirectoryKeys(msg, currentState)
	case *DestinationDirectorySelectionState:
		return m.handleDestinationDirectoryKeys(msg, currentState)
	case *SyncOptionsState:
		return m.handleSyncOptionsKeys(msg, currentState)
	case *DirectoryFilterState:
		return m.handleDirectoryFilterKeys(msg, currentState)
	case *ExclusionPatternsState:
		return m.handleExclusionPatternsKeys(msg, currentState)
	case *PreviewState:
		return m.handlePreviewKeys(msg, currentState)
	case *ProgressState:
		return m.handleProgressKeys(msg, currentState)
	case *CompletedState:
		return m.handleCompletedKeys(msg, currentState)
	}

	return m, nil
}

// State-specific key handlers that use type-safe transitions

func (m TypeStateModel) handleWelcomeKeys(msg tea.KeyMsg, state *WelcomeState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Toggle between sync modes (in a real implementation)
		// For now, we'll transition immediately to source selection
	case "down", "j":
		// Toggle between sync modes
	case "enter", "right", "l":
		// Use type-safe transition
		newState := state.SelectSyncMode(OneWaySync)
		m.currentState = newState
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleSourceDirectoryKeys(msg tea.KeyMsg, state *SourceDirectorySelectionState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Use type-safe transition - this ensures source path is set
		// In a real implementation, this would get the selected path from the directory tree
		selectedPath := "/tmp/test-source" // Mock selection
		newState := state.SetSourceDirectory(selectedPath)
		m.currentState = newState
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleDestinationDirectoryKeys(msg tea.KeyMsg, state *DestinationDirectorySelectionState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Use type-safe transition - this ensures destination path is set
		selectedPath := "/tmp/test-dest" // Mock selection
		newState := state.SetDestinationDirectory(selectedPath, false)
		m.currentState = newState
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleSyncOptionsKeys(msg tea.KeyMsg, state *SyncOptionsState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "right":
		// Use type-safe transition with default options
		config := DefaultSyncOptions()
		newState := state.ConfigureOptions(config)
		m.currentState = newState
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleDirectoryFilterKeys(msg tea.KeyMsg, state *DirectoryFilterState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "right":
		// Use type-safe transition with selected folders
		folders := map[string]bool{"src/": true, "docs/": true} // Mock selection
		newState := state.SetSelectedFolders(folders)
		m.currentState = newState
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleExclusionPatternsKeys(msg tea.KeyMsg, state *ExclusionPatternsState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "right":
		// Use type-safe transition with exclusion patterns
		patterns := []string{"*.log", "*.tmp", ".DS_Store"} // Mock patterns
		newState := state.SetExclusionPatterns(patterns)
		m.currentState = newState
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handlePreviewKeys(msg tea.KeyMsg, state *PreviewState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Use type-safe transition to start sync
		newState := state.StartSync()
		m.currentState = newState
		return m, nil
	case "s":
		// Use type-safe SyncFile generation
		filename, err := state.SaveAsSyncFile()
		if err != nil {
			m.err = err
		} else {
			// Show success message (in a real implementation)
			_ = filename
		}
		return m, nil
	case "backspace", "left", "h":
		if state.CanGoBack() {
			newState := state.GoBack()
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleProgressKeys(msg tea.KeyMsg, state *ProgressState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Use type-safe cancellation
		newState, err := state.CancelSync()
		if err != nil {
			m.err = err
		} else {
			m.currentState = newState
		}
		return m, nil
	}
	return m, nil
}

func (m TypeStateModel) handleCompletedKeys(msg tea.KeyMsg, state *CompletedState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q":
		// Use type-safe exit
		if state.Exit() {
			return m, tea.Quit
		}
	}
	return m, nil
}

// Rendering methods

func (m TypeStateModel) renderCurrentState() string {
	// Use the state's own render method
	return m.currentState.Render()
}

func (m TypeStateModel) renderHeader() string {
	title := "sync-tools Interactive Wizard (Type-Safe)"
	
	// Progress indicator
	screenNames := []string{
		"Welcome", "Source", "Destination", "Options", 
		"Folders", "Exclusions", "Preview", "Progress", "Complete",
	}
	
	currentScreen := int(m.currentState.GetCurrentScreen())
	var progress string
	for i, name := range screenNames {
		if i == currentScreen {
			progress += m.styles.SelectedStyle.Render(name)
		} else if i < currentScreen {
			progress += m.styles.SuccessStyle.Render(name)
		} else {
			progress += m.styles.UnselectedStyle.Render(name)
		}
		
		if i < len(screenNames)-1 {
			if i < currentScreen {
				progress += m.styles.SuccessStyle.Render(" → ")
			} else {
				progress += m.styles.UnselectedStyle.Render(" → ")
			}
		}
	}

	return fmt.Sprintf("%s\n\n%s",
		m.styles.TitleStyle.Render(title),
		progress,
	)
}

func (m TypeStateModel) renderFooter() string {
	var shortcuts []string
	
	// Navigation shortcuts
	if m.currentState.CanGoBack() {
		shortcuts = append(shortcuts, "←/b Back")
	}
	// We can't easily check CanGoNext in the type state pattern without type assertions
	// This is a trade-off - we could add it to the WizardStateBase interface if needed
	
	// Global shortcuts
	shortcuts = append(shortcuts, "? Help", "q Quit")
	
	return m.styles.FooterStyle.Render(fmt.Sprintf("Type-safe wizard • %s", 
		fmt.Sprintf("Screen: %v • %s", m.currentState.GetCurrentScreen(), 
		fmt.Sprintf("Navigate with Enter/Arrow keys • %s", 
		fmt.Sprintf("%s", "Additional context-sensitive help available")))))
}

func (m TypeStateModel) renderMinimumSizeWarning() string {
	return m.styles.WarningStyle.Render(fmt.Sprintf(
		"Terminal too small (%dx%d)\nMinimum size required: 80x24\nPlease resize your terminal and try again.",
		m.windowWidth, m.windowHeight,
	))
}

func (m TypeStateModel) renderHelpOverlay() string {
	help := "Type-Safe Wizard Help\n\n"
	
	switch m.currentState.GetCurrentScreen() {
	case WelcomeScreen:
		help += "Select your sync mode with compile-time guarantees:\n"
		help += "• States ensure you cannot access unset configuration\n"
		help += "• Invalid transitions are impossible at compile time\n"
	case SourceDirectoryScreen:
		help += "Source directory selection with type safety:\n"
		help += "• Cannot proceed without selecting a valid source\n"
		help += "• Previous configuration is preserved when going back\n"
	default:
		help += "Context-sensitive help for type-safe state transitions."
	}
	
	return m.styles.HelpStyle.Render(help)
}