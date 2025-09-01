package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the wizard UI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// BubbleTeaModel wraps WizardModel for Bubble Tea
type BubbleTeaModel struct {
	Model    *WizardModel
	Error    string
	quitting bool
}

// NewBubbleTeaModel creates a new Bubble Tea model
func NewBubbleTeaModel(model *WizardModel) *BubbleTeaModel {
	return &BubbleTeaModel{
		Model: model,
	}
}

// Init initializes the Bubble Tea model
func (m *BubbleTeaModel) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages
func (m *BubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		
		case "?", "h":
			// TODO: Show help screen
			return m, nil
			
		default:
			return m.handleStateSpecificKeys(msg)
		}
	
	case tea.WindowSizeMsg:
		// Handle window resize if needed
		return m, nil
	}

	return m, nil
}

// View renders the current wizard state
func (m *BubbleTeaModel) View() string {
	if m.quitting {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🧙 Sync Wizard"))
	content.WriteString("\n\n")

	// Show error if present
	if m.Error != "" {
		content.WriteString(errorStyle.Render("❌ Error: " + m.Error))
		content.WriteString("\n\n")
	}

	// Render state-specific content
	content.WriteString(m.renderCurrentState())

	// Help text
	content.WriteString("\n" + helpStyle.Render("Press 'q' to quit, '?' for help"))

	return content.String()
}

// renderCurrentState renders the UI for the current state
func (m *BubbleTeaModel) renderCurrentState() string {
	switch state := m.Model.CurrentState.(type) {
	case InitialState:
		return m.renderInitialState()
	case SourceSelectionState:
		return m.renderSourceSelectionState(state)
	case DestinationSelectionState:
		return m.renderDestinationSelectionState(state)
	case SyncOptionsState:
		return m.renderSyncOptionsState(state)
	case ExclusionPatternsState:
		return m.renderExclusionPatternsState(state)
	case DirectoryFilterState:
		return m.renderDirectoryFilterState(state)
	case ProgressState:
		return m.renderProgressState(state)
	case CompleteState:
		return m.renderCompleteState(state)
	default:
		return "Unknown state"
	}
}

// handleStateSpecificKeys handles keyboard input based on current state
func (m *BubbleTeaModel) handleStateSpecificKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch state := m.Model.CurrentState.(type) {
	case InitialState:
		return m.handleInitialStateKeys(msg)
	case SourceSelectionState:
		return m.handleSourceSelectionKeys(msg, state)
	case DestinationSelectionState:
		return m.handleDestinationSelectionKeys(msg, state)
	case SyncOptionsState:
		return m.handleSyncOptionsKeys(msg, state)
	case ExclusionPatternsState:
		return m.handleExclusionPatternsKeys(msg, state)
	case DirectoryFilterState:
		return m.handleDirectoryFilterKeys(msg, state)
	case ProgressState:
		return m.handleProgressKeys(msg, state)
	case CompleteState:
		return m.handleCompleteKeys(msg, state)
	}
	
	return m, nil
}

// Placeholder render methods - will be implemented in subsequent phases
func (m *BubbleTeaModel) renderInitialState() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Welcome to the Sync Wizard"))
	content.WriteString("\n\n")
	content.WriteString("This wizard will help you configure a sync operation.\n")
	content.WriteString("\nPress [Enter] to start by selecting a source directory")
	return content.String()
}

// renderSourceSelectionState renders the source directory selection screen
func (m *BubbleTeaModel) renderSourceSelectionState(state SourceSelectionState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Select source directory"))
	content.WriteString("\n\n")
	
	// Initialize directory browser if not exists
	if state.Browser == nil {
		// This should be handled in the state initialization
		content.WriteString("Initializing directory browser...\n")
		return content.String()
	}
	
	content.WriteString(fmt.Sprintf("Current path: %s\n\n", state.Browser.GetCurrentPath()))
	
	// Render directory entries
	entries := state.Browser.GetEntries()
	selectedIndex := state.Browser.GetSelectedIndex()
	
	for i, entry := range entries {
		prefix := "  "
		if i == selectedIndex {
			prefix = "▶ "
		}
		
		icon := "📁"
		if !entry.IsDir {
			icon = "📄"
		}
		
		sizeInfo := ""
		if entry.IsDir && entry.FileCount > 0 {
			sizeInfo = fmt.Sprintf(" (%d files, %s)", entry.FileCount, FormatSize(entry.Size))
		} else if !entry.IsDir {
			sizeInfo = fmt.Sprintf(" (%s)", FormatSize(entry.Size))
		}
		
		content.WriteString(fmt.Sprintf("%s%s %s%s\n", prefix, icon, entry.Name, sizeInfo))
	}
	
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate, →: Enter directory, ←: Go up, Enter: Select, Esc: Cancel"))
	return content.String()
}

// renderDestinationSelectionState renders the destination directory selection screen
func (m *BubbleTeaModel) renderDestinationSelectionState(state DestinationSelectionState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Select destination directory"))
	content.WriteString("\n\n")
	
	content.WriteString(fmt.Sprintf("Source: %s\n", state.SourcePath))
	
	// Initialize directory browser if not exists
	if state.Browser == nil {
		content.WriteString("Initializing directory browser...\n")
		return content.String()
	}
	
	content.WriteString(fmt.Sprintf("Current path: %s\n\n", state.Browser.GetCurrentPath()))
	
	// Render directory entries
	entries := state.Browser.GetEntries()
	selectedIndex := state.Browser.GetSelectedIndex()
	
	for i, entry := range entries {
		prefix := "  "
		if i == selectedIndex {
			prefix = "▶ "
		}
		
		icon := "📁"
		if !entry.IsDir {
			icon = "📄"
		}
		
		sizeInfo := ""
		if entry.IsDir && entry.FileCount > 0 {
			sizeInfo = fmt.Sprintf(" (%d files, %s)", entry.FileCount, FormatSize(entry.Size))
		} else if !entry.IsDir {
			sizeInfo = fmt.Sprintf(" (%s)", FormatSize(entry.Size))
		}
		
		content.WriteString(fmt.Sprintf("%s%s %s%s\n", prefix, icon, entry.Name, sizeInfo))
	}
	
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate, →: Enter directory, ←: Go up, Enter: Select, Esc: Back"))
	return content.String()
}

func (m *BubbleTeaModel) renderSyncOptionsState(state SyncOptionsState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Configure sync options"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Source: %s\n", state.SourcePath))
	content.WriteString(fmt.Sprintf("Destination: %s\n", state.DestinationPath))
	content.WriteString(fmt.Sprintf("Mode: %s\n", state.Mode))
	content.WriteString(fmt.Sprintf("Dry Run: %v\n", state.DryRun))
	content.WriteString("\n[Placeholder: Interactive controls will be implemented here]\n")
	content.WriteString("\nPress [Enter] to continue, [Esc] to go back")
	return content.String()
}

func (m *BubbleTeaModel) renderExclusionPatternsState(state ExclusionPatternsState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Manage exclusion patterns"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Patterns: %d configured\n", len(state.Patterns)))
	content.WriteString("\n[Placeholder: Pattern editor will be implemented here]\n")
	content.WriteString("\nPress [Enter] to continue, [Esc] to go back")
	return content.String()
}

func (m *BubbleTeaModel) renderDirectoryFilterState(state DirectoryFilterState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Select directories to sync"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Directories: %d found\n", len(state.Directories)))
	content.WriteString("\n[Placeholder: Directory selection will be implemented here]\n")
	content.WriteString("\nPress [Enter] to continue, [Esc] to go back")
	return content.String()
}

func (m *BubbleTeaModel) renderProgressState(state ProgressState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Sync Progress"))
	content.WriteString("\n\n")
	content.WriteString("[Placeholder: Real-time progress will be implemented here]\n")
	content.WriteString(fmt.Sprintf("Current file: %s\n", state.Progress.CurrentFile))
	content.WriteString(fmt.Sprintf("Progress: %d%%\n", state.Progress.ProgressPercent))
	return content.String()
}

func (m *BubbleTeaModel) renderCompleteState(state CompleteState) string {
	var content strings.Builder
	if state.Success {
		content.WriteString(headerStyle.Render("✅ Wizard Complete!"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("SyncFile generated: %s\n", state.SyncFilePath))
	} else {
		content.WriteString(headerStyle.Render("❌ Wizard Failed"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("Error: %s\n", state.Error))
	}
	content.WriteString("\nPress [Enter] to exit")
	return content.String()
}

// Placeholder key handlers - will be implemented with real logic
func (m *BubbleTeaModel) handleInitialStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to source selection with directory browser
		browser := NewDirectoryBrowser(".")
		m.Model.CurrentState = SourceSelectionState{
			CurrentPath: ".",
			Directories: []DirectoryInfo{},
			Browser:     browser,
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleSourceSelectionKeys(msg tea.KeyMsg, state SourceSelectionState) (tea.Model, tea.Cmd) {
	if state.Browser == nil {
		// Initialize browser if missing
		state.Browser = NewDirectoryBrowser(state.CurrentPath)
		m.Model.CurrentState = state
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		state.Browser.MoveUp()
		m.Model.CurrentState = state
		
	case "down", "j":
		state.Browser.MoveDown()
		m.Model.CurrentState = state
		
	case "right", "l":
		// Enter selected directory
		if state.Browser.EnterDirectory() {
			m.Model.CurrentState = state
		}
		
	case "left", "h":
		// Go to parent directory
		if state.Browser.GoUp() {
			m.Model.CurrentState = state
		}
		
	case "enter":
		// Select current directory as source
		selectedPath := state.Browser.GetCurrentPath()
		browser := NewDirectoryBrowser(".")
		m.Model.CurrentState = DestinationSelectionState{
			SourcePath:  selectedPath,
			CurrentPath: ".",
			Directories: []DirectoryInfo{},
			Browser:     browser,
		}
		
	case "escape":
		m.Model.CurrentState = InitialState{}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleDestinationSelectionKeys(msg tea.KeyMsg, state DestinationSelectionState) (tea.Model, tea.Cmd) {
	if state.Browser == nil {
		// Initialize browser if missing
		state.Browser = NewDirectoryBrowser(state.CurrentPath)
		m.Model.CurrentState = state
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		state.Browser.MoveUp()
		m.Model.CurrentState = state
		
	case "down", "j":
		state.Browser.MoveDown()
		m.Model.CurrentState = state
		
	case "right", "l":
		// Enter selected directory
		if state.Browser.EnterDirectory() {
			m.Model.CurrentState = state
		}
		
	case "left", "h":
		// Go to parent directory
		if state.Browser.GoUp() {
			m.Model.CurrentState = state
		}
		
	case "enter":
		// Select current directory as destination and move to sync options
		selectedPath := state.Browser.GetCurrentPath()
		m.Model.CurrentState = SyncOptionsState{
			SourcePath:      state.SourcePath,
			DestinationPath: selectedPath,
			Mode:           "one-way",
			DryRun:         false,
			HiddenDirs:     true,
			UseGitIgnore:   false,
			ConflictStrategy: "newest-wins",
		}
		
	case "escape":
		// Go back to source selection
		browser := NewDirectoryBrowser(state.SourcePath)
		m.Model.CurrentState = SourceSelectionState{
			CurrentPath: state.SourcePath,
			Directories: []DirectoryInfo{},
			Browser:     browser,
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleSyncOptionsKeys(msg tea.KeyMsg, state SyncOptionsState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to exclusion patterns
		m.Model.CurrentState = ExclusionPatternsState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state,
			Patterns:        []ExclusionPattern{{Pattern: ".git/", Source: "default", Valid: true}},
		}
	case "escape":
		m.Model.CurrentState = DestinationSelectionState{
			SourcePath:  state.SourcePath,
			CurrentPath: state.DestinationPath,
			Directories: []DirectoryInfo{},
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleExclusionPatternsKeys(msg tea.KeyMsg, state ExclusionPatternsState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to directory filter
		m.Model.CurrentState = DirectoryFilterState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state.SyncOptions,
			Patterns:        state.Patterns,
			Directories:     []SelectableDirectory{},
		}
	case "escape":
		m.Model.CurrentState = state.SyncOptions
	}
	return m, nil
}

func (m *BubbleTeaModel) handleDirectoryFilterKeys(msg tea.KeyMsg, state DirectoryFilterState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to progress
		m.Model.CurrentState = ProgressState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state.SyncOptions,
			Patterns:        state.Patterns,
			Directories:     state.Directories,
			Progress:        ProgressInfo{},
		}
		// TODO: Start actual sync operation
	case "escape":
		m.Model.CurrentState = state.SyncOptions
	}
	return m, nil
}

func (m *BubbleTeaModel) handleProgressKeys(msg tea.KeyMsg, state ProgressState) (tea.Model, tea.Cmd) {
	// Progress state is usually non-interactive except for cancellation
	return m, nil
}

func (m *BubbleTeaModel) handleCompleteKeys(msg tea.KeyMsg, state CompleteState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}