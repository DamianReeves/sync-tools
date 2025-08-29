package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/DamianReeves/sync-tools/internal/rsync"
	"github.com/DamianReeves/sync-tools/internal/logging"
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	progressStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true)
)

// Model represents the Bubble Tea model for interactive sync
type Model struct {
	opts     *rsync.Options
	logger   logging.Logger
	state    syncState
	progress string
	error    string
	result   string
	quitting bool
}

type syncState int

const (
	stateIdle syncState = iota
	stateSyncing
	stateComplete
	stateError
)

// syncMsg is sent when sync operation completes
type syncMsg struct {
	err error
}

// syncProgressMsg is sent during sync operation
type syncProgressMsg struct {
	message string
}

// NewModel creates a new interactive sync model
func NewModel(opts *rsync.Options, logger logging.Logger) Model {
	return Model{
		opts:   opts,
		logger: logger,
		state:  stateIdle,
	}
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter", " ":
			if m.state == stateIdle {
				m.state = stateSyncing
				m.progress = "Starting sync..."
				return m, m.performSync()
			}
			if m.state == stateComplete || m.state == stateError {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case syncProgressMsg:
		m.progress = msg.message
		return m, nil

	case syncMsg:
		if msg.err != nil {
			m.state = stateError
			m.error = msg.err.Error()
		} else {
			m.state = stateComplete
			m.result = "Sync completed successfully!"
		}
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🔄 Interactive Sync"))
	content.WriteString("\n\n")

	// Sync configuration info
	info := fmt.Sprintf(
		"Source: %s\nDest:   %s\nMode:   %s\nDry Run: %v",
		m.opts.Source,
		m.opts.Dest,
		m.opts.Mode,
		m.opts.DryRun,
	)
	content.WriteString(infoStyle.Render(info))
	content.WriteString("\n\n")

	// State-specific content
	switch m.state {
	case stateIdle:
		content.WriteString("Press [Enter] or [Space] to start sync\n")
		content.WriteString("Press [q] or [Ctrl+C] to quit\n")

	case stateSyncing:
		content.WriteString(progressStyle.Render("⏳ " + m.progress))
		content.WriteString("\n\nSyncing in progress... Press [Ctrl+C] to cancel\n")

	case stateComplete:
		content.WriteString(successStyle.Render("✅ " + m.result))
		content.WriteString("\n\nPress [Enter] or [q] to exit\n")

	case stateError:
		content.WriteString(errorStyle.Render("❌ Sync failed:"))
		content.WriteString("\n" + m.error)
		content.WriteString("\n\nPress [Enter] or [q] to exit\n")
	}

	return content.String()
}

// performSync executes the sync operation in the background
func (m Model) performSync() tea.Cmd {
	return func() tea.Msg {
		// Create a new runner
		runner := rsync.NewRunner(m.logger)

		// Execute sync
		err := runner.Sync(m.opts)

		return syncMsg{err: err}
	}
}