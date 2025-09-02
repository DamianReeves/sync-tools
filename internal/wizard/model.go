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

// BubbleTeaModel wraps StateMachine for Bubble Tea
type BubbleTeaModel struct {
	StateMachine *StateMachine
	Config       *Config
	Error        string
	quitting     bool
}

// NewBubbleTeaModel creates a new Bubble Tea model
func NewBubbleTeaModel(stateMachine *StateMachine, config *Config) *BubbleTeaModel {
	return &BubbleTeaModel{
		StateMachine: stateMachine,
		Config:       config,
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
			// Show context-specific help
			return m.showContextHelp()

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
	currentState := m.StateMachine.CurrentState()
	switch state := currentState.(type) {
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
	currentState := m.StateMachine.CurrentState()
	switch state := currentState.(type) {
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

	// Manual path entry mode
	if state.ManualEntry {
		content.WriteString("Enter path manually:\n")
		pathDisplay := state.PathInput
		if pathDisplay == "" {
			pathDisplay = "(type path here)"
		}
		content.WriteString(fmt.Sprintf("Path: %s\n", pathDisplay))

		if state.PathError != "" {
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s\n", state.PathError)))
		}

		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Type path, Enter: Confirm, Esc: Back to browser, t: Toggle mode"))
		return content.String()
	}

	// Browser mode
	if state.Browser == nil {
		content.WriteString("Initializing directory browser...\n")
		return content.String()
	}

	content.WriteString(fmt.Sprintf("Current path: %s\n\n", state.Browser.GetCurrentPath()))

	// Show scroll indicators
	if state.Browser.HasMoreAbove() {
		content.WriteString(helpStyle.Render("    ↑ More entries above...\n"))
	}

	// Render visible directory entries only
	visibleEntries := state.Browser.GetVisibleEntries()
	visibleSelectedIndex := state.Browser.GetVisibleSelectedIndex()

	for i, entry := range visibleEntries {
		prefix := "  "
		if i == visibleSelectedIndex {
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

	// Show scroll indicators
	if state.Browser.HasMoreBelow() {
		content.WriteString(helpStyle.Render("    ↓ More entries below...\n"))
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate, →: Enter, ←: Up, Enter: Select, t: Type path, ?: Help, Esc: Cancel"))
	return content.String()
}

// renderDestinationSelectionState renders the destination directory selection screen
func (m *BubbleTeaModel) renderDestinationSelectionState(state DestinationSelectionState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Select destination directory"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Source: %s\n", state.SourcePath))

	// Manual path entry mode
	if state.ManualEntry {
		content.WriteString("\nEnter path manually:\n")
		pathDisplay := state.PathInput
		if pathDisplay == "" {
			pathDisplay = "(type path here)"
		}
		content.WriteString(fmt.Sprintf("Path: %s\n", pathDisplay))

		if state.PathError != "" {
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s\n", state.PathError)))
		}

		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Type path, Enter: Confirm, Esc: Back to browser, t: Toggle mode"))
		return content.String()
	}

	// Browser mode
	if state.Browser == nil {
		content.WriteString("Initializing directory browser...\n")
		return content.String()
	}

	content.WriteString(fmt.Sprintf("Current path: %s\n\n", state.Browser.GetCurrentPath()))

	// Show scroll indicators
	if state.Browser.HasMoreAbove() {
		content.WriteString(helpStyle.Render("    ↑ More entries above...\n"))
	}

	// Render visible directory entries only
	visibleEntries := state.Browser.GetVisibleEntries()
	visibleSelectedIndex := state.Browser.GetVisibleSelectedIndex()

	for i, entry := range visibleEntries {
		prefix := "  "
		if i == visibleSelectedIndex {
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

	// Show scroll indicators
	if state.Browser.HasMoreBelow() {
		content.WriteString(helpStyle.Render("    ↓ More entries below...\n"))
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate, →: Enter, ←: Up, Enter: Select, t: Type path, ?: Help, Esc: Back"))
	return content.String()
}

// renderSyncOptionsState renders the sync options configuration screen
func (m *BubbleTeaModel) renderSyncOptionsState(state SyncOptionsState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Configure sync options"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Source: %s\n", state.SourcePath))
	content.WriteString(fmt.Sprintf("Destination: %s\n\n", state.DestinationPath))

	// Initialize editor if not exists
	if state.Editor == nil {
		state.Editor = NewSyncOptionsEditor(&state)
		_ = m.StateMachine.TransitionTo(state)
		return content.String()
	}

	// Render all option fields
	content.WriteString(state.Editor.RenderAllFields())

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate, Space/Enter: Toggle, ←→: Change value, Tab: Continue, Esc: Back"))
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

// renderProgressState renders the sync progress screen
func (m *BubbleTeaModel) renderProgressState(state ProgressState) string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Sync Progress"))
	content.WriteString("\n\n")

	// Initialize monitor if needed
	if state.Monitor == nil {
		state.Monitor = NewProgressMonitor(&state)
		_ = state.Monitor.StartSync(&state)
		_ = m.StateMachine.TransitionTo(state)
		return content.String()
	}

	progress := state.Monitor.GetProgress()

	// Current operation
	content.WriteString(fmt.Sprintf("Current file: %s\n", progress.CurrentFile))
	content.WriteString(fmt.Sprintf("Progress: %d/%d files\n", progress.FilesProcessed, progress.TotalFiles))

	// Progress bar
	progressBar := RenderProgressBar(progress.ProgressPercent, 40)
	content.WriteString(fmt.Sprintf("%s\n\n", progressBar))

	// Transfer statistics
	if progress.TransferSpeed != "" {
		content.WriteString(fmt.Sprintf("Transfer speed: %s\n", progress.TransferSpeed))
	}

	if progress.BytesTransferred > 0 {
		content.WriteString(fmt.Sprintf("Transferred: %s", FormatSize(progress.BytesTransferred)))
		if progress.TotalBytes > 0 {
			content.WriteString(fmt.Sprintf(" / %s", FormatSize(progress.TotalBytes)))
		}
		content.WriteString("\n")
	}

	// Status-specific content
	if state.Monitor.IsActive() {
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Sync in progress... Press Ctrl+C to cancel"))
	} else if state.Monitor.IsComplete() {
		content.WriteString("\n✅ Sync completed successfully!\n")

		// Generate SyncFile
		if syncFileContent, err := GenerateSyncFile(&state); err == nil {
			content.WriteString("\n📄 Generated SyncFile:\n")
			content.WriteString(infoStyle.Render(syncFileContent))
		}

		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Press Enter to finish, Esc to go back"))
	} else if state.Monitor.IsFailed() {
		content.WriteString(fmt.Sprintf("\n❌ Sync failed: %v\n", state.Monitor.GetError()))
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Press Enter to finish, Esc to go back"))
	}

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

// Placeholder key handlers - now using type-safe state machine
func (m *BubbleTeaModel) handleInitialStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Use type-safe operations
		if ops, err := m.StateMachine.GetInitialOperations(); err == nil {
			if err := ops.StartSourceSelection(); err != nil {
				m.Error = fmt.Sprintf("Failed to start source selection: %v", err)
			}
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleSourceSelectionKeys(msg tea.KeyMsg, state SourceSelectionState) (tea.Model, tea.Cmd) {
	// Handle manual path entry mode
	if state.ManualEntry {
		return m.handleManualPathEntry(msg, &state.ManualEntry, &state.PathInput, &state.PathError, func(path string) error {
			ops, err := m.StateMachine.GetSourceSelectionOperations()
			if err != nil {
				return err
			}
			return ops.SelectSource(path)
		}, func() {
			state.ManualEntry = false
			state.PathInput = ""
			state.PathError = ""
			_ = m.StateMachine.TransitionTo(state)
		})
	}

	if state.Browser == nil {
		// Initialize browser if missing
		state.Browser = NewDirectoryBrowser(state.CurrentPath)
		_ = m.StateMachine.TransitionTo(state)
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		state.Browser.MoveUp()
		_ = m.StateMachine.TransitionTo(state)

	case "down", "j":
		state.Browser.MoveDown()
		_ = m.StateMachine.TransitionTo(state)

	case "right", "l":
		// Enter selected directory
		if state.Browser.EnterDirectory() {
			_ = m.StateMachine.TransitionTo(state)
		}

	case "left", "h":
		// Go to parent directory
		if state.Browser.GoUp() {
			_ = m.StateMachine.TransitionTo(state)
		}

	case "t":
		// Toggle to manual path entry mode
		state.ManualEntry = true
		state.PathInput = state.Browser.GetCurrentPath()
		state.PathError = ""
		_ = m.StateMachine.TransitionTo(state)

	case "enter":
		// Select current directory as source using type-safe operations
		selectedPath := state.Browser.GetCurrentPath()
		if ops, err := m.StateMachine.GetSourceSelectionOperations(); err == nil {
			if err := ops.SelectSource(selectedPath); err != nil {
				m.Error = fmt.Sprintf("Failed to select source: %v", err)
			}
		}

	case "escape":
		// Navigate back using state machine
		if m.StateMachine.CanGoBack() {
			_ = m.StateMachine.GoBack()
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleDestinationSelectionKeys(msg tea.KeyMsg, state DestinationSelectionState) (tea.Model, tea.Cmd) {
	// Handle manual path entry mode
	if state.ManualEntry {
		return m.handleManualPathEntry(msg, &state.ManualEntry, &state.PathInput, &state.PathError, func(path string) error {
			ops, err := m.StateMachine.GetDestinationSelectionOperations()
			if err != nil {
				return err
			}
			return ops.SelectDestination(path)
		}, func() {
			state.ManualEntry = false
			state.PathInput = ""
			state.PathError = ""
			_ = m.StateMachine.TransitionTo(state)
		})
	}

	if state.Browser == nil {
		// Initialize browser if missing
		state.Browser = NewDirectoryBrowser(state.CurrentPath)
		_ = m.StateMachine.TransitionTo(state)
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		state.Browser.MoveUp()
		_ = m.StateMachine.TransitionTo(state)

	case "down", "j":
		state.Browser.MoveDown()
		_ = m.StateMachine.TransitionTo(state)

	case "right", "l":
		// Enter selected directory
		if state.Browser.EnterDirectory() {
			_ = m.StateMachine.TransitionTo(state)
		}

	case "left", "h":
		// Go to parent directory
		if state.Browser.GoUp() {
			_ = m.StateMachine.TransitionTo(state)
		}

	case "t":
		// Toggle to manual path entry mode
		state.ManualEntry = true
		state.PathInput = state.Browser.GetCurrentPath()
		state.PathError = ""
		_ = m.StateMachine.TransitionTo(state)

	case "enter":
		// Select current directory as destination using type-safe operations
		selectedPath := state.Browser.GetCurrentPath()
		if ops, err := m.StateMachine.GetDestinationSelectionOperations(); err == nil {
			if err := ops.SelectDestination(selectedPath); err != nil {
				m.Error = fmt.Sprintf("Failed to select destination: %v", err)
			}
		}

	case "escape":
		// Navigate back using state machine
		if m.StateMachine.CanGoBack() {
			_ = m.StateMachine.GoBack()
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleSyncOptionsKeys(msg tea.KeyMsg, state SyncOptionsState) (tea.Model, tea.Cmd) {
	if state.Editor == nil {
		// Initialize editor if missing
		state.Editor = NewSyncOptionsEditor(&state)
		_ = m.StateMachine.TransitionTo(state)
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		state.Editor.MoveUp()
		_ = m.StateMachine.TransitionTo(state)

	case "down", "j":
		state.Editor.MoveDown()
		_ = m.StateMachine.TransitionTo(state)

	case "left", "h":
		state.Editor.ChangeValue(-1)
		_ = m.StateMachine.TransitionTo(state)

	case "right", "l":
		state.Editor.ChangeValue(1)
		_ = m.StateMachine.TransitionTo(state)

	case "space", "enter":
		state.Editor.ToggleValue()
		_ = m.StateMachine.TransitionTo(state)

	case "tab":
		// Proceed to exclusion patterns
		newState := ExclusionPatternsState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state,
			Patterns:        []ExclusionPattern{{Pattern: ".git/", Source: "default", Valid: true}},
		}
		_ = m.StateMachine.TransitionTo(newState)

	case "escape":
		// Navigate back using state machine
		if m.StateMachine.CanGoBack() {
			_ = m.StateMachine.GoBack()
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleExclusionPatternsKeys(msg tea.KeyMsg, state ExclusionPatternsState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to directory filter
		newState := DirectoryFilterState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state.SyncOptions,
			Patterns:        state.Patterns,
			Directories:     []SelectableDirectory{},
		}
		_ = m.StateMachine.TransitionTo(newState)
	case "escape":
		// Navigate back using state machine
		if m.StateMachine.CanGoBack() {
			_ = m.StateMachine.GoBack()
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleDirectoryFilterKeys(msg tea.KeyMsg, state DirectoryFilterState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Transition to progress
		newState := ProgressState{
			SourcePath:      state.SourcePath,
			DestinationPath: state.DestinationPath,
			SyncOptions:     state.SyncOptions,
			Patterns:        state.Patterns,
			Directories:     state.Directories,
			Progress:        ProgressInfo{},
			Monitor:         nil, // Will be initialized when rendered
		}
		_ = m.StateMachine.TransitionTo(newState)
	case "escape":
		// Navigate back using state machine
		if m.StateMachine.CanGoBack() {
			_ = m.StateMachine.GoBack()
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleProgressKeys(msg tea.KeyMsg, state ProgressState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Cancel sync if active
		if state.Monitor != nil && state.Monitor.IsActive() {
			_ = state.Monitor.Cancel()
			_ = m.StateMachine.TransitionTo(state)
		} else {
			// Quit if not active
			m.quitting = true
			return m, tea.Quit
		}

	case "enter":
		// Finish wizard if sync is complete or failed
		if state.Monitor != nil && (state.Monitor.IsComplete() || state.Monitor.IsFailed()) {
			completeState := CompleteState{
				SyncFilePath: "SyncFile", // TODO: Save to actual file
				Success:      state.Monitor.IsComplete(),
				Error:        "",
			}
			if state.Monitor.IsFailed() && state.Monitor.GetError() != nil {
				completeState.Error = state.Monitor.GetError().Error()
			}
			_ = m.StateMachine.TransitionTo(completeState)
		}

	case "escape":
		// Navigate back if sync is not active
		if state.Monitor == nil || !state.Monitor.IsActive() {
			if m.StateMachine.CanGoBack() {
				_ = m.StateMachine.GoBack()
			}
		}
	}
	return m, nil
}

func (m *BubbleTeaModel) handleCompleteKeys(msg tea.KeyMsg, _ CompleteState) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handleManualPathEntry handles text input for manual path entry
func (m *BubbleTeaModel) handleManualPathEntry(
	msg tea.KeyMsg,
	_ *bool,
	pathInput *string,
	pathError *string,
	selectPath func(string) error,
	exitManualEntry func(),
) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Validate and select the path
		if *pathInput == "" {
			*pathError = "Path cannot be empty"
			return m, nil
		}

		// Try to select the path
		if err := selectPath(*pathInput); err != nil {
			*pathError = err.Error()
			return m, nil
		}
		// Success - path was selected and wizard moved to next state
		return m, nil

	case "escape", "t":
		// Exit manual entry mode
		exitManualEntry()
		return m, nil

	case "backspace":
		// Remove last character
		if len(*pathInput) > 0 {
			*pathInput = (*pathInput)[:len(*pathInput)-1]
			*pathError = "" // Clear error when editing
		}
		return m, nil

	default:
		// Add character to path input
		if len(msg.String()) == 1 {
			*pathInput += msg.String()
			*pathError = "" // Clear error when editing
		}
	}
	return m, nil
}

// showContextHelp shows help information based on current state
func (m *BubbleTeaModel) showContextHelp() (tea.Model, tea.Cmd) {
	// For now, just show a simple help message
	// In a full implementation, this would show a proper help screen
	currentState := m.StateMachine.CurrentState()

	var helpText string
	switch currentState.(type) {
	case SourceSelectionState:
		helpText = "Source Selection Help:\n" +
			"↑↓/jk: Navigate directories\n" +
			"→/l: Enter directory\n" +
			"←/h: Go to parent directory\n" +
			"Enter: Select current directory\n" +
			"t: Toggle manual path entry\n" +
			"Esc: Cancel selection\n" +
			"q: Quit wizard"

	case DestinationSelectionState:
		helpText = "Destination Selection Help:\n" +
			"↑↓/jk: Navigate directories\n" +
			"→/l: Enter directory\n" +
			"←/h: Go to parent directory\n" +
			"Enter: Select current directory\n" +
			"t: Toggle manual path entry\n" +
			"Esc: Go back to source selection\n" +
			"q: Quit wizard"

	default:
		helpText = "General Help:\n" +
			"q: Quit wizard\n" +
			"Esc: Go back to previous step\n" +
			"?: Show this help"
	}

	// For now, set error to show help (in a real implementation, you'd show a modal or separate screen)
	m.Error = helpText
	return m, nil
}
