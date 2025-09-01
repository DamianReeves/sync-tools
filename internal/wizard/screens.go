package wizard

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen 1: Welcome & Mode Selection
func (m Model) renderWelcomeScreen() string {
	var content strings.Builder
	
	content.WriteString(m.styles.HeaderStyle.Render("Welcome & Mode Selection"))
	content.WriteString("\n\n")
	
	content.WriteString("Welcome! This wizard will guide you through setting up a sync operation.\n\n")
	
	content.WriteString("Select sync mode:\n\n")
	
	// Mode options
	modes := []SyncMode{OneWaySync, TwoWaySync}
	for _, mode := range modes {
		prefix := "○"
		style := m.styles.UnselectedStyle
		
		if mode == m.state.Mode {
			prefix = "●"
			style = m.styles.SelectedStyle
		}
		
		line := fmt.Sprintf("%s %s", prefix, mode.String())
		
		if mode == TwoWaySync {
			line += "  " + m.styles.WarningStyle.Render("[Coming in future release]")
		}
		
		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	content.WriteString(m.state.Mode.Description())
	
	return content.String()
}

func (m Model) handleWelcomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.state.Mode == TwoWaySync {
			m.state.Mode = OneWaySync
		}
	case "down", "j":
		// For now, don't allow selecting two-way sync
		// if m.state.Mode == OneWaySync {
		//     m.state.Mode = TwoWaySync
		// }
	case "enter", "right", "l":
		if m.state.CanGoNext() {
			m.state.GoNext()
		}
	case "left", "h", "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	
	return m, nil
}

// Screen 2: Source Directory Selection
func (m Model) renderSourceDirectoryScreen() string {
	var content strings.Builder
	
	content.WriteString(m.styles.HeaderStyle.Render("Select Source Directory"))
	content.WriteString("\n\n")
	
	content.WriteString("Choose the directory to sync FROM:\n\n")
	
	if m.state.SourcePath != "" {
		content.WriteString(fmt.Sprintf("Current: %s\n\n", 
			m.styles.HighlightStyle.Render(m.state.SourcePath)))
	}
	
	// Initialize directory tree if not exists
	if m.state.SourceDirectoryTree == nil {
		homeDir, _ := os.UserHomeDir()
		if homeDir == "" {
			homeDir = "/"
		}
		m.state.SourceDirectoryTree = NewDirectoryTree(homeDir)
		m.state.SourceDirectoryTree.SetViewHeight(10)
		m.state.SourceDirectoryTree.Refresh()
	}
	
	// Render directory tree
	content.WriteString(m.state.SourceDirectoryTree.Render())
	content.WriteString("\n\n")
	
	// Path input box
	selectedPath := m.state.SourceDirectoryTree.GetCurrentPath()
	content.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	content.WriteString(fmt.Sprintf("│ Path: %-53s │\n", selectedPath))
	content.WriteString("└─────────────────────────────────────────────────────────────┘")
	
	return content.String()
}

func (m Model) handleSourceDirectoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state.SourceDirectoryTree == nil {
		return m, nil
	}
	
	switch msg.String() {
	case "up", "k":
		m.state.SourceDirectoryTree.MoveUp()
	case "down", "j":
		m.state.SourceDirectoryTree.MoveDown()
	case "right", "l":
		m.state.SourceDirectoryTree.ExpandSelected()
	case "left", "h":
		m.state.SourceDirectoryTree.CollapseSelected()
	case "enter":
		if item := m.state.SourceDirectoryTree.GetSelectedItem(); item != nil && item.IsDir {
			m.state.SourcePath = item.Path
			if m.state.CanGoNext() {
				m.state.GoNext()
			}
		}
	case "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	
	return m, nil
}

// Screen 3: Destination Directory Selection
func (m Model) renderDestinationDirectoryScreen() string {
	var content strings.Builder
	
	content.WriteString(m.styles.HeaderStyle.Render("Select Destination Directory"))
	content.WriteString("\n\n")
	
	content.WriteString("Choose the directory to sync TO:\n\n")
	
	content.WriteString(fmt.Sprintf("Source: %s\n", 
		m.styles.HighlightStyle.Render(m.state.SourcePath)))
		
	if m.state.DestinationPath != "" {
		content.WriteString(fmt.Sprintf("Current: %s\n\n", 
			m.styles.HighlightStyle.Render(m.state.DestinationPath)))
	} else {
		content.WriteString("\n")
	}
	
	// Initialize destination directory tree if not exists
	if m.state.DestinationDirectoryTree == nil {
		homeDir, _ := os.UserHomeDir()
		if homeDir == "" {
			homeDir = "/"
		}
		m.state.DestinationDirectoryTree = NewDirectoryTree(homeDir)
		m.state.DestinationDirectoryTree.SetViewHeight(8)
		m.state.DestinationDirectoryTree.Refresh()
	}
	
	// Render directory tree
	content.WriteString(m.state.DestinationDirectoryTree.Render())
	content.WriteString("\n\n")
	
	// Path input box
	selectedPath := m.state.DestinationDirectoryTree.GetCurrentPath()
	content.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	content.WriteString(fmt.Sprintf("│ Path: %-53s │\n", selectedPath))
	content.WriteString("└─────────────────────────────────────────────────────────────┘\n\n")
	
	// Create destination option
	checkbox := "□"
	if m.state.CreateDestinationDir {
		checkbox = "☑"
	}
	content.WriteString(fmt.Sprintf("%s Create destination directory if it doesn't exist", checkbox))
	
	return content.String()
}

func (m Model) handleDestinationDirectoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state.DestinationDirectoryTree == nil {
		return m, nil
	}
	
	switch msg.String() {
	case "up", "k":
		m.state.DestinationDirectoryTree.MoveUp()
	case "down", "j":
		m.state.DestinationDirectoryTree.MoveDown()
	case "right", "l":
		m.state.DestinationDirectoryTree.ExpandSelected()
	case "left", "h":
		m.state.DestinationDirectoryTree.CollapseSelected()
	case "enter":
		if item := m.state.DestinationDirectoryTree.GetSelectedItem(); item != nil && item.IsDir {
			m.state.DestinationPath = item.Path
			if m.state.CanGoNext() {
				m.state.GoNext()
			}
		}
	case "space":
		m.state.CreateDestinationDir = !m.state.CreateDestinationDir
	case "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	
	return m, nil
}

// Screen 4: Sync Options Configuration
func (m Model) renderSyncOptionsScreen() string {
	var content strings.Builder
	
	content.WriteString(m.styles.HeaderStyle.Render("Sync Options"))
	content.WriteString("\n\n")
	
	content.WriteString("Configure sync behavior:\n\n")
	
	// Basic Options
	content.WriteString(m.styles.HighlightStyle.Render("Basic Options:"))
	content.WriteString("\n")
	
	options := []struct {
		name        string
		value       *bool
		description string
	}{
		{"Dry run (preview only - no actual changes)", &m.state.DryRun, ""},
		{"Verbose output", &m.state.Verbose, ""},
		{"Delete files in destination that don't exist in source", &m.state.DeleteExtraFiles, ""},
	}
	
	for _, opt := range options {
		checkbox := "□"
		if *opt.value {
			checkbox = "☑"
		}
		content.WriteString(fmt.Sprintf("  %s %s\n", checkbox, opt.name))
	}
	
	content.WriteString("\n")
	content.WriteString(m.styles.HighlightStyle.Render("Advanced Options:"))
	content.WriteString("\n")
	
	advancedOptions := []struct {
		name        string
		value       *bool
		description string
	}{
		{"Use .gitignore files to exclude files", &m.state.UseGitignore, ""},
		{"Follow symbolic links", &m.state.FollowSymlinks, ""},
		{"Preserve file timestamps", &m.state.PreserveTimestamps, ""},
		{"Exclude hidden files and directories", &m.state.ExcludeHiddenDirs, ""},
	}
	
	for _, opt := range advancedOptions {
		checkbox := "□"
		if *opt.value {
			checkbox = "☑"
		}
		content.WriteString(fmt.Sprintf("  %s %s\n", checkbox, opt.name))
	}
	
	content.WriteString("\n")
	content.WriteString(m.styles.HighlightStyle.Render("Performance:"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  Parallel transfers: [%-5d] (1-16)\n", m.state.ParallelTransfers))
	content.WriteString(fmt.Sprintf("  Transfer timeout:   [%-5ds] seconds\n", m.state.TransferTimeout))
	
	return content.String()
}

func (m Model) handleSyncOptionsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.state.DryRun = !m.state.DryRun
	case "2":
		m.state.Verbose = !m.state.Verbose
	case "3":
		m.state.DeleteExtraFiles = !m.state.DeleteExtraFiles
	case "4":
		m.state.UseGitignore = !m.state.UseGitignore
	case "5":
		m.state.FollowSymlinks = !m.state.FollowSymlinks
	case "6":
		m.state.PreserveTimestamps = !m.state.PreserveTimestamps
	case "7":
		m.state.ExcludeHiddenDirs = !m.state.ExcludeHiddenDirs
	case "enter", "right":
		if m.state.CanGoNext() {
			m.state.GoNext()
		}
	case "left", "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	
	return m, nil
}

// Placeholder screen implementations (to be completed)
func (m Model) renderDirectoryFilterScreen() string {
	return m.styles.HeaderStyle.Render("Directory Tree Filter Selection") + "\n\n" +
		"This screen will show folders from the source directory with checkboxes\n" +
		"to select which folders should be synchronized.\n\n" +
		"[Implementation in progress...]"
}

func (m Model) renderExclusionPatternsScreen() string {
	return m.styles.HeaderStyle.Render("File Exclusion Patterns") + "\n\n" +
		"This screen will allow adding/removing exclusion patterns\n" +
		"like *.log, *.tmp, node_modules/, etc.\n\n" +
		"[Implementation in progress...]"
}

func (m Model) renderPreviewScreen() string {
	return m.styles.HeaderStyle.Render("Sync Preview") + "\n\n" +
		"This screen will show a complete summary of all sync settings\n" +
		"and allow the user to start the sync or save as SyncFile.\n\n" +
		"[Implementation in progress...]"
}

func (m Model) renderProgressScreen() string {
	return m.styles.HeaderStyle.Render("Sync Progress") + "\n\n" +
		"This screen will show real-time progress of the sync operation\n" +
		"with progress bars, current file, and transfer statistics.\n\n" +
		"[Implementation in progress...]"
}

func (m Model) renderCompletedScreen() string {
	return m.styles.HeaderStyle.Render("Sync Complete") + "\n\n" +
		"This screen will show the sync completion status\n" +
		"and options for next steps.\n\n" +
		"[Implementation in progress...]"
}

// Placeholder key handlers (to be completed)
func (m Model) handleDirectoryFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "right":
		if m.state.CanGoNext() {
			m.state.GoNext()
		}
	case "left", "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	return m, nil
}

func (m Model) handleExclusionPatternsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "right":
		if m.state.CanGoNext() {
			m.state.GoNext()
		}
	case "left", "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	return m, nil
}

func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Start sync
		m.state.GoNext()
	case "s":
		// Save as SyncFile
		// TODO: Implement SyncFile generation
	case "left", "backspace":
		if m.state.CanGoBack() {
			m.state.GoBack()
		}
	}
	return m, nil
}

func (m Model) handleProgressKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Cancel sync
		return m, tea.Quit
	case "space":
		// Pause/resume sync
		// TODO: Implement pause/resume
	}
	return m, nil
}

func (m Model) handleCompletedKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q":
		return m, tea.Quit
	}
	return m, nil
}