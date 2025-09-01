package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model implements the main wizard model for Bubble Tea
type Model struct {
	state  *WizardState
	styles *Styles
	err    error
}

// Styles contains all the styling for the wizard UI
type Styles struct {
	// Layout styles
	BorderStyle        lipgloss.Style
	TitleStyle         lipgloss.Style
	HeaderStyle        lipgloss.Style
	ContentStyle       lipgloss.Style
	FooterStyle        lipgloss.Style

	// Interactive elements
	SelectedStyle      lipgloss.Style
	UnselectedStyle    lipgloss.Style
	ActiveStyle        lipgloss.Style
	InactiveStyle      lipgloss.Style
	CheckboxSelected   lipgloss.Style
	CheckboxUnselected lipgloss.Style

	// Text styles
	HighlightStyle     lipgloss.Style
	ErrorStyle         lipgloss.Style
	SuccessStyle       lipgloss.Style
	WarningStyle       lipgloss.Style
	HelpStyle          lipgloss.Style

	// Progress styles
	ProgressBarStyle   lipgloss.Style
	ProgressTextStyle  lipgloss.Style
}

// NewModel creates a new wizard model
func NewModel() Model {
	return Model{
		state:  NewWizardState(),
		styles: NewStyles(),
	}
}

// NewStyles creates the default styling for the wizard
func NewStyles() *Styles {
	return &Styles{
		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),

		TitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Align(lipgloss.Center),

		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),

		ContentStyle: lipgloss.NewStyle().
			Padding(1, 2),

		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),

		SelectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),

		UnselectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		ActiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Bold(true),

		InactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		CheckboxSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true),

		CheckboxUnselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		HighlightStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		SuccessStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true),

		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true),

		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),

		ProgressBarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),

		ProgressTextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true),
	}
}

// Init implements the Bubble Tea init method
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements the Bubble Tea update method
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMessage(msg)
	case tea.WindowSizeMsg:
		m.state.WindowWidth = msg.Width
		m.state.WindowHeight = msg.Height
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// handleKeyMessage processes keyboard input
func (m Model) handleKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys available on all screens
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state.CurrentScreen != ProgressScreen {
			return m, tea.Quit
		}
	case "?":
		m.state.ShowHelp = !m.state.ShowHelp
		return m, nil
	}

	// Screen-specific key handling
	switch m.state.CurrentScreen {
	case WelcomeScreen:
		return m.handleWelcomeKeys(msg)
	case SourceDirectoryScreen:
		return m.handleSourceDirectoryKeys(msg)
	case DestinationDirectoryScreen:
		return m.handleDestinationDirectoryKeys(msg)
	case SyncOptionsScreen:
		return m.handleSyncOptionsKeys(msg)
	case DirectoryFilterScreen:
		return m.handleDirectoryFilterKeys(msg)
	case ExclusionPatternsScreen:
		return m.handleExclusionPatternsKeys(msg)
	case PreviewScreen:
		return m.handlePreviewKeys(msg)
	case ProgressScreen:
		return m.handleProgressKeys(msg)
	case CompletedScreen:
		return m.handleCompletedKeys(msg)
	}

	return m, nil
}

// View implements the Bubble Tea view method
func (m Model) View() string {
	if m.state.WindowWidth < 80 || m.state.WindowHeight < 24 {
		return m.renderMinimumSizeWarning()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// Main content based on current screen
	switch m.state.CurrentScreen {
	case WelcomeScreen:
		content.WriteString(m.renderWelcomeScreen())
	case SourceDirectoryScreen:
		content.WriteString(m.renderSourceDirectoryScreen())
	case DestinationDirectoryScreen:
		content.WriteString(m.renderDestinationDirectoryScreen())
	case SyncOptionsScreen:
		content.WriteString(m.renderSyncOptionsScreen())
	case DirectoryFilterScreen:
		content.WriteString(m.renderDirectoryFilterScreen())
	case ExclusionPatternsScreen:
		content.WriteString(m.renderExclusionPatternsScreen())
	case PreviewScreen:
		content.WriteString(m.renderPreviewScreen())
	case ProgressScreen:
		content.WriteString(m.renderProgressScreen())
	case CompletedScreen:
		content.WriteString(m.renderCompletedScreen())
	}

	content.WriteString("\n")

	// Footer
	content.WriteString(m.renderFooter())

	// Error display
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(m.styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
	}

	// Help overlay
	if m.state.ShowHelp {
		content.WriteString("\n")
		content.WriteString(m.renderHelpOverlay())
	}

	// Apply border and return
	return m.styles.BorderStyle.
		Width(m.state.WindowWidth - 4).
		Height(m.state.WindowHeight - 2).
		Render(content.String())
}

// renderHeader renders the wizard title and progress indicator
func (m Model) renderHeader() string {
	title := "sync-tools Interactive Wizard"
	
	// Progress indicator
	screenNames := []string{
		"Welcome", "Source", "Destination", "Options", 
		"Folders", "Exclusions", "Preview", "Progress", "Complete",
	}
	
	var progress strings.Builder
	for i, name := range screenNames {
		if i == int(m.state.CurrentScreen) {
			progress.WriteString(m.styles.SelectedStyle.Render(name))
		} else if i < int(m.state.CurrentScreen) {
			progress.WriteString(m.styles.SuccessStyle.Render(name))
		} else {
			progress.WriteString(m.styles.UnselectedStyle.Render(name))
		}
		
		if i < len(screenNames)-1 {
			if i < int(m.state.CurrentScreen) {
				progress.WriteString(m.styles.SuccessStyle.Render(" → "))
			} else {
				progress.WriteString(m.styles.UnselectedStyle.Render(" → "))
			}
		}
	}

	return fmt.Sprintf("%s\n\n%s",
		m.styles.TitleStyle.Render(title),
		progress.String(),
	)
}

// renderFooter renders the navigation help and current screen info
func (m Model) renderFooter() string {
	var shortcuts []string
	
	// Navigation shortcuts
	if m.state.CanGoBack() {
		shortcuts = append(shortcuts, "←/b Back")
	}
	if m.state.CanGoNext() {
		shortcuts = append(shortcuts, "→/Enter Next")
	}
	
	// Screen-specific shortcuts
	switch m.state.CurrentScreen {
	case WelcomeScreen:
		shortcuts = append(shortcuts, "↑↓ Navigate", "Enter Select")
	case SourceDirectoryScreen, DestinationDirectoryScreen:
		shortcuts = append(shortcuts, "↑↓ Navigate", "→ Expand", "← Collapse", "/ Search")
	case SyncOptionsScreen:
		shortcuts = append(shortcuts, "↑↓ Navigate", "Space Toggle", "←→ Adjust")
	case DirectoryFilterScreen:
		shortcuts = append(shortcuts, "↑↓ Navigate", "Space Toggle", "a Select All", "n Select None")
	case ExclusionPatternsScreen:
		shortcuts = append(shortcuts, "↑↓ Navigate", "Enter Add", "Del Remove")
	case PreviewScreen:
		shortcuts = append(shortcuts, "s Save as SyncFile")
	}
	
	// Global shortcuts
	shortcuts = append(shortcuts, "? Help", "q Quit")
	
	return m.styles.FooterStyle.Render(strings.Join(shortcuts, " • "))
}

// renderMinimumSizeWarning shows a warning when terminal is too small
func (m Model) renderMinimumSizeWarning() string {
	return m.styles.WarningStyle.Render(fmt.Sprintf(
		"Terminal too small (%dx%d)\nMinimum size required: 80x24\nPlease resize your terminal and try again.",
		m.state.WindowWidth, m.state.WindowHeight,
	))
}

// renderHelpOverlay renders context-sensitive help
func (m Model) renderHelpOverlay() string {
	help := "Wizard Help\n\n"
	
	switch m.state.CurrentScreen {
	case WelcomeScreen:
		help += "Select your sync mode:\n"
		help += "• One-way: Copy files from source to destination only\n"
		help += "• Two-way: Keep both directories synchronized (future release)\n"
	case SourceDirectoryScreen:
		help += "Choose the directory to sync FROM:\n"
		help += "• Use arrow keys to navigate the directory tree\n"
		help += "• Press → to expand folders, ← to collapse\n"
		help += "• Type / to search for directories\n"
	case DestinationDirectoryScreen:
		help += "Choose the directory to sync TO:\n"
		help += "• Navigate the same way as source selection\n"
		help += "• Check 'Create directory' if it doesn't exist\n"
	default:
		help += "Context-sensitive help for this screen will be displayed here."
	}
	
	return m.styles.HelpStyle.Render(help)
}

// Screen-specific key handling and rendering methods are implemented in screens.go