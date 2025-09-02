package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Modal styles
var (
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Margin(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)

	errorModalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// ManualPathEntryModal handles manual path input
type ManualPathEntryModal struct {
	baseState    WizardState
	pathInput    string
	pathError    string
	suggestions  []string
	history      []string
	historyIndex int
}

// NewManualPathEntryModal creates a new manual path entry modal
func NewManualPathEntryModal(baseState WizardState, initialPath string) *ManualPathEntryModal {
	return &ManualPathEntryModal{
		baseState:    baseState,
		pathInput:    initialPath,
		pathError:    "",
		suggestions:  []string{},
		history:      []string{},
		historyIndex: -1,
	}
}

func (m *ManualPathEntryModal) BaseState() WizardState {
	return m.baseState
}

func (m *ManualPathEntryModal) ModalType() string {
	return "manual-path-entry"
}

func (m *ManualPathEntryModal) Render(baseContent string) string {
	var content strings.Builder

	// Modal title
	content.WriteString(modalTitleStyle.Render("📝 Enter Path Manually"))
	content.WriteString("\n\n")

	// Path input
	pathDisplay := m.pathInput
	if pathDisplay == "" {
		pathDisplay = "(type path here)"
	}
	content.WriteString("Path: ")
	content.WriteString(inputStyle.Render(pathDisplay))
	content.WriteString("\n")

	// Error display
	if m.pathError != "" {
		content.WriteString(errorModalStyle.Render("❌ " + m.pathError))
		content.WriteString("\n")
	}

	// Suggestions
	if len(m.suggestions) > 0 {
		content.WriteString("\n" + suggestionStyle.Render("Suggestions:"))
		for _, suggestion := range m.suggestions {
			content.WriteString("\n  " + suggestionStyle.Render("• "+suggestion))
		}
		content.WriteString("\n")
	}

	// Help
	content.WriteString("\n" + helpStyle.Render("Type path | Enter: Confirm | Esc: Cancel | ~: Home | Tab: Complete | ↑↓: History"))

	modalContent := modalStyle.Render(content.String())

	// Compose with base content
	return baseContent + "\n" + modalContent
}

func (m *ManualPathEntryModal) HandleInput(msg tea.KeyMsg) (Modal, tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		// Validate and complete path entry
		if m.pathInput == "" {
			m.pathError = "Path cannot be empty"
			return m, nil, true
		}

		expandedPath := m.expandPath(m.pathInput)
		if err := m.validatePath(expandedPath); err != nil {
			m.pathError = err.Error()
			return m, nil, true
		}

		// Add to history
		m.addToHistory(m.pathInput)

		// Update base state with the selected path
		m.updateBaseStateWithPath(expandedPath)

		// Signal completion
		return m, nil, true // Modal will be popped by caller

	case "escape":
		// Cancel path entry
		return m, nil, true // Signal handled, modal will be popped

	case "backspace":
		// Remove last character
		if len(m.pathInput) > 0 {
			m.pathInput = m.pathInput[:len(m.pathInput)-1]
			m.pathError = ""
			m.updateSuggestions()
		}
		return m, nil, true

	case "tab":
		// Path completion
		m.performPathCompletion()
		return m, nil, true

	case "up":
		// Previous history
		m.navigateHistory(-1)
		return m, nil, true

	case "down":
		// Next history
		m.navigateHistory(1)
		return m, nil, true

	case "ctrl+u":
		// Clear entire path
		m.pathInput = ""
		m.pathError = ""
		m.suggestions = []string{}
		return m, nil, true

	default:
		// Handle character input and special keys
		switch msg.String() {
		case "~":
			// Home directory shortcut
			if homeDir, err := os.UserHomeDir(); err == nil {
				m.pathInput = homeDir
				m.pathError = ""
				m.updateSuggestions()
			}
		case "/":
			// Root directory shortcut
			m.pathInput = "/"
			m.pathError = ""
			m.updateSuggestions()
		default:
			// Regular character input
			if len(msg.String()) == 1 {
				m.pathInput += msg.String()
				m.pathError = ""
				m.updateSuggestions()
			}
		}
		return m, nil, true
	}
}

func (m *ManualPathEntryModal) IsComplete() (bool, WizardState) {
	return false, m.baseState // Modal handles its own completion
}

// expandPath expands ~ and environment variables
func (m *ManualPathEntryModal) expandPath(path string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	} else if path == "~" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = homeDir
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Clean the path
	return filepath.Clean(path)
}

// validatePath checks if the path is valid and accessible
func (m *ManualPathEntryModal) validatePath(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("cannot access path: %v", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// updateSuggestions generates path completion suggestions
func (m *ManualPathEntryModal) updateSuggestions() {
	m.suggestions = []string{}

	if m.pathInput == "" {
		// Suggest common paths
		if homeDir, err := os.UserHomeDir(); err == nil {
			m.suggestions = append(m.suggestions, "~ ("+homeDir+")")
		}
		m.suggestions = append(m.suggestions, "/ (root)")
		return
	}

	// Get directory for completion
	dir := filepath.Dir(m.expandPath(m.pathInput))
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && len(m.suggestions) < 5 {
				fullPath := filepath.Join(dir, entry.Name())
				if strings.HasPrefix(strings.ToLower(fullPath), strings.ToLower(m.expandPath(m.pathInput))) {
					m.suggestions = append(m.suggestions, fullPath)
				}
			}
		}
	}
}

// performPathCompletion attempts to complete the current path
func (m *ManualPathEntryModal) performPathCompletion() {
	if len(m.suggestions) == 1 {
		// Single suggestion - use it
		m.pathInput = m.suggestions[0]
		m.updateSuggestions()
	} else if len(m.suggestions) > 1 {
		// Multiple suggestions - find common prefix
		commonPrefix := m.findCommonPrefix(m.suggestions)
		if len(commonPrefix) > len(m.pathInput) {
			m.pathInput = commonPrefix
			m.updateSuggestions()
		}
	}
}

// findCommonPrefix finds the common prefix among suggestions
func (m *ManualPathEntryModal) findCommonPrefix(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	prefix := suggestions[0]
	for _, s := range suggestions[1:] {
		for i := 0; i < len(prefix) && i < len(s); i++ {
			if prefix[i] != s[i] {
				prefix = prefix[:i]
				break
			}
		}
	}
	return prefix
}

// addToHistory adds a path to the history
func (m *ManualPathEntryModal) addToHistory(path string) {
	// Avoid duplicates
	for i, h := range m.history {
		if h == path {
			// Move to end
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}

	m.history = append(m.history, path)

	// Limit history size
	if len(m.history) > 20 {
		m.history = m.history[1:]
	}

	m.historyIndex = len(m.history)
}

// navigateHistory moves through path history
func (m *ManualPathEntryModal) navigateHistory(direction int) {
	if len(m.history) == 0 {
		return
	}

	m.historyIndex += direction

	if m.historyIndex < 0 {
		m.historyIndex = 0
	} else if m.historyIndex >= len(m.history) {
		m.historyIndex = len(m.history)
		m.pathInput = ""
		return
	}

	m.pathInput = m.history[m.historyIndex]
	m.pathError = ""
	m.updateSuggestions()
}

// updateBaseStateWithPath updates the base state with the selected path
func (m *ManualPathEntryModal) updateBaseStateWithPath(path string) {
	switch state := m.baseState.(type) {
	case SourceSelectionState:
		// Create new browser for the path
		state.Browser = NewDirectoryBrowser(path)
		state.CurrentPath = path
		m.baseState = state

	case DestinationSelectionState:
		// Create new browser for the path
		state.Browser = NewDirectoryBrowser(path)
		state.CurrentPath = path
		m.baseState = state
	}
}

// HelpModal displays context-sensitive help
type HelpModal struct {
	baseState WizardState
	helpText  string
}

// NewHelpModal creates a new help modal with context-specific help
func NewHelpModal(baseState WizardState) *HelpModal {
	helpText := generateContextualHelp(baseState)
	return &HelpModal{
		baseState: baseState,
		helpText:  helpText,
	}
}

func (h *HelpModal) BaseState() WizardState {
	return h.baseState
}

func (h *HelpModal) ModalType() string {
	return "help"
}

func (h *HelpModal) Render(baseContent string) string {
	var content strings.Builder

	// Modal title
	content.WriteString(modalTitleStyle.Render("❓ Help"))
	content.WriteString("\n\n")

	// Help content
	content.WriteString(h.helpText)
	content.WriteString("\n\n")

	// Close instruction
	content.WriteString(helpStyle.Render("Press any key to close help"))

	modalContent := modalStyle.Render(content.String())

	// Compose with base content
	return baseContent + "\n" + modalContent
}

func (h *HelpModal) HandleInput(msg tea.KeyMsg) (Modal, tea.Cmd, bool) {
	// Any key closes help
	return h, nil, true
}

func (h *HelpModal) IsComplete() (bool, WizardState) {
	return false, h.baseState
}

// generateContextualHelp creates help text based on the current state
func generateContextualHelp(state WizardState) string {
	switch state.(type) {
	case SourceSelectionState:
		return `Source Directory Selection:

Navigation:
  ↑↓ or j/k    Navigate directories
  → or l       Enter selected directory
  ← or h       Go to parent directory
  Enter        Select current directory as source

Quick Actions:
  t            Manual path entry mode
  ~            Jump to home directory (in manual mode)
  1-9          Quick bookmarks (if configured)

General:
  ?            Show this help
  Esc          Cancel/go back
  q            Quit wizard`

	case DestinationSelectionState:
		return `Destination Directory Selection:

Navigation:
  ↑↓ or j/k    Navigate directories
  → or l       Enter selected directory
  ← or h       Go to parent directory
  Enter        Select current directory as destination

Quick Actions:
  t            Manual path entry mode
  ~            Jump to home directory (in manual mode)
  1-9          Quick bookmarks (if configured)

General:
  ?            Show this help
  Esc          Go back to source selection
  q            Quit wizard`

	default:
		return `Sync Wizard Help:

General Navigation:
  ↑↓           Move up/down
  Enter        Select/confirm
  Esc          Go back/cancel
  q            Quit wizard
  ?            Show help (this screen)

Path Entry:
  t            Toggle manual path entry
  ~            Home directory shortcut
  Tab          Path completion
  ↑↓           Browse path history`
	}
}

// HomeNavigationModal provides quick access to common directories
type HomeNavigationModal struct {
	baseState     WizardState
	bookmarks     []Bookmark
	selectedIndex int
}

// Bookmark represents a quick-access directory bookmark
type Bookmark struct {
	Name        string
	Path        string
	Description string
	Icon        string
}

// NewHomeNavigationModal creates a new home navigation modal
func NewHomeNavigationModal(baseState WizardState) *HomeNavigationModal {
	bookmarks := generateBookmarks()
	return &HomeNavigationModal{
		baseState:     baseState,
		bookmarks:     bookmarks,
		selectedIndex: 0,
	}
}

func (h *HomeNavigationModal) BaseState() WizardState {
	return h.baseState
}

func (h *HomeNavigationModal) ModalType() string {
	return "home-navigation"
}

func (h *HomeNavigationModal) Render(baseContent string) string {
	var content strings.Builder

	// Modal title
	content.WriteString(modalTitleStyle.Render("🏠 Quick Navigation"))
	content.WriteString("\n\n")

	// Bookmarks list
	for i, bookmark := range h.bookmarks {
		prefix := "  "
		if i == h.selectedIndex {
			prefix = "▶ "
		}

		content.WriteString(fmt.Sprintf("%s%s %s", prefix, bookmark.Icon, bookmark.Name))
		if bookmark.Description != "" {
			content.WriteString(fmt.Sprintf(" - %s", suggestionStyle.Render(bookmark.Description)))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑↓: Navigate | Enter: Select | 1-9: Quick select | Esc: Cancel"))

	modalContent := modalStyle.Render(content.String())

	// Compose with base content
	return baseContent + "\n" + modalContent
}

func (h *HomeNavigationModal) HandleInput(msg tea.KeyMsg) (Modal, tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		if h.selectedIndex > 0 {
			h.selectedIndex--
		}
		return h, nil, true

	case "down", "j":
		if h.selectedIndex < len(h.bookmarks)-1 {
			h.selectedIndex++
		}
		return h, nil, true

	case "enter":
		// Select current bookmark
		if h.selectedIndex >= 0 && h.selectedIndex < len(h.bookmarks) {
			bookmark := h.bookmarks[h.selectedIndex]
			h.updateBaseStateWithBookmark(bookmark)
		}
		return h, nil, true

	case "escape":
		// Cancel navigation
		return h, nil, true

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick number selection
		index := int(msg.String()[0] - '1')
		if index >= 0 && index < len(h.bookmarks) {
			bookmark := h.bookmarks[index]
			h.updateBaseStateWithBookmark(bookmark)
		}
		return h, nil, true

	default:
		return h, nil, false // Not handled
	}
}

func (h *HomeNavigationModal) IsComplete() (bool, WizardState) {
	return false, h.baseState
}

// generateBookmarks creates a list of common directory bookmarks
func generateBookmarks() []Bookmark {
	bookmarks := []Bookmark{}

	// Home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		bookmarks = append(bookmarks, Bookmark{
			Name:        "Home",
			Path:        homeDir,
			Description: homeDir,
			Icon:        "🏠",
		})

		// Common subdirectories
		commonDirs := []struct {
			name, subpath, icon string
		}{
			{"Desktop", "Desktop", "🖥️"},
			{"Documents", "Documents", "📄"},
			{"Downloads", "Downloads", "⬇️"},
			{"Pictures", "Pictures", "📸"},
			{"Music", "Music", "🎵"},
			{"Videos", "Videos", "🎬"},
		}

		for _, dir := range commonDirs {
			fullPath := filepath.Join(homeDir, dir.subpath)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				bookmarks = append(bookmarks, Bookmark{
					Name:        dir.name,
					Path:        fullPath,
					Description: fullPath,
					Icon:        dir.icon,
				})
			}
		}
	}

	// Root directory
	bookmarks = append(bookmarks, Bookmark{
		Name:        "Root",
		Path:        "/",
		Description: "System root directory",
		Icon:        "💾",
	})

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		bookmarks = append(bookmarks, Bookmark{
			Name:        "Current Dir",
			Path:        cwd,
			Description: cwd,
			Icon:        "📁",
		})
	}

	return bookmarks
}

// updateBaseStateWithBookmark updates the base state with the selected bookmark path
func (h *HomeNavigationModal) updateBaseStateWithBookmark(bookmark Bookmark) {
	switch state := h.baseState.(type) {
	case SourceSelectionState:
		state.Browser = NewDirectoryBrowser(bookmark.Path)
		state.CurrentPath = bookmark.Path
		h.baseState = state

	case DestinationSelectionState:
		state.Browser = NewDirectoryBrowser(bookmark.Path)
		state.CurrentPath = bookmark.Path
		h.baseState = state
	}
}
