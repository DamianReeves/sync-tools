package wizard

import (
	"fmt"
	"os"
	"path/filepath"
)

// Implementation of state transition methods for each state

// WelcomeState implementations
func (s *WelcomeState) GetCurrentScreen() WizardScreen {
	return WelcomeScreen
}

func (s *WelcomeState) CanGoBack() bool {
	return false // Welcome is the first screen
}

func (s *WelcomeState) Render() string {
	return fmt.Sprintf("Welcome Screen - Current Mode: %s", s.syncMode.String())
}

func (s *WelcomeState) SelectSyncMode(mode SyncMode) *SourceDirectorySelectionState {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "/"
	}
	
	return &SourceDirectorySelectionState{
		syncMode:      mode,
		directoryTree: NewDirectoryTree(homeDir),
	}
}

// SourceDirectorySelectionState implementations
func (s *SourceDirectorySelectionState) GetCurrentScreen() WizardScreen {
	return SourceDirectoryScreen
}

func (s *SourceDirectorySelectionState) CanGoBack() bool {
	return true
}

func (s *SourceDirectorySelectionState) Render() string {
	return fmt.Sprintf("Source Directory Selection - Mode: %s", s.syncMode.String())
}

func (s *SourceDirectorySelectionState) GoBack() *WelcomeState {
	return &WelcomeState{
		syncMode: s.syncMode,
	}
}

func (s *SourceDirectorySelectionState) SetSourceDirectory(path string) *DestinationDirectorySelectionState {
	// Validate that the path exists and is a directory
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		// In a real implementation, this would return an error
		// For now, we'll create the state anyway for testing
	}
	
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "/"
	}
	
	return &DestinationDirectorySelectionState{
		syncMode:      s.syncMode,
		sourcePath:    path,
		directoryTree: NewDirectoryTree(homeDir),
	}
}

// DestinationDirectorySelectionState implementations
func (s *DestinationDirectorySelectionState) GetCurrentScreen() WizardScreen {
	return DestinationDirectoryScreen
}

func (s *DestinationDirectorySelectionState) CanGoBack() bool {
	return true
}

func (s *DestinationDirectorySelectionState) Render() string {
	return fmt.Sprintf("Destination Directory Selection - Mode: %s, Source: %s", 
		s.syncMode.String(), s.sourcePath)
}

func (s *DestinationDirectorySelectionState) GoBack() *SourceDirectorySelectionState {
	return &SourceDirectorySelectionState{
		syncMode:      s.syncMode,
		directoryTree: s.directoryTree, // Preserve directory tree state
	}
}

func (s *DestinationDirectorySelectionState) SetDestinationDirectory(path string, createIfNotExists bool) *SyncOptionsState {
	// Validate destination path or check if it can be created
	if createIfNotExists {
		// Validate that parent directory exists
		parentDir := filepath.Dir(path)
		if _, err := os.Stat(parentDir); err != nil {
			// In a real implementation, this would return an error
		}
	} else {
		// Validate that the path exists and is a directory
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			// In a real implementation, this would return an error
		}
	}
	
	return &SyncOptionsState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   path,
		createDestination: createIfNotExists,
	}
}

// SyncOptionsState implementations
func (s *SyncOptionsState) GetCurrentScreen() WizardScreen {
	return SyncOptionsScreen
}

func (s *SyncOptionsState) CanGoBack() bool {
	return true
}

func (s *SyncOptionsState) Render() string {
	return fmt.Sprintf("Sync Options - Mode: %s, Source: %s, Dest: %s", 
		s.syncMode.String(), s.sourcePath, s.destinationPath)
}

func (s *SyncOptionsState) GoBack() *DestinationDirectorySelectionState {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "/"
	}
	
	return &DestinationDirectorySelectionState{
		syncMode:        s.syncMode,
		sourcePath:      s.sourcePath,
		directoryTree:   NewDirectoryTree(homeDir),
	}
}

func (s *SyncOptionsState) ConfigureOptions(config SyncOptionsConfig) *DirectoryFilterState {
	// Scan source directory to build folder stats
	folderStats := scanDirectoryForStats(s.sourcePath)
	
	return &DirectoryFilterState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
		syncOptions:       config,
		folderStats:       folderStats,
	}
}

// DirectoryFilterState implementations
func (s *DirectoryFilterState) GetCurrentScreen() WizardScreen {
	return DirectoryFilterScreen
}

func (s *DirectoryFilterState) CanGoBack() bool {
	return true
}

func (s *DirectoryFilterState) Render() string {
	return fmt.Sprintf("Directory Filter - %d folders available", len(s.folderStats))
}

func (s *DirectoryFilterState) GoBack() *SyncOptionsState {
	return &SyncOptionsState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
	}
}

func (s *DirectoryFilterState) SetSelectedFolders(folders map[string]bool) *ExclusionPatternsState {
	return &ExclusionPatternsState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
		syncOptions:       s.syncOptions,
		selectedFolders:   folders,
	}
}

// ExclusionPatternsState implementations
func (s *ExclusionPatternsState) GetCurrentScreen() WizardScreen {
	return ExclusionPatternsScreen
}

func (s *ExclusionPatternsState) CanGoBack() bool {
	return true
}

func (s *ExclusionPatternsState) Render() string {
	return fmt.Sprintf("Exclusion Patterns - %d folders selected", len(s.selectedFolders))
}

func (s *ExclusionPatternsState) GoBack() *DirectoryFilterState {
	// Rebuild folder stats since we're going back
	folderStats := scanDirectoryForStats(s.sourcePath)
	
	return &DirectoryFilterState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
		syncOptions:       s.syncOptions,
		folderStats:       folderStats,
	}
}

func (s *ExclusionPatternsState) SetExclusionPatterns(patterns []string) *PreviewState {
	// Calculate estimates based on selected folders and exclusion patterns
	estimatedFiles, estimatedSize := calculateEstimates(s.sourcePath, s.selectedFolders, patterns)
	
	return &PreviewState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
		syncOptions:       s.syncOptions,
		selectedFolders:   s.selectedFolders,
		exclusionPatterns: patterns,
		estimatedFiles:    estimatedFiles,
		estimatedSize:     estimatedSize,
	}
}

// PreviewState implementations
func (s *PreviewState) GetCurrentScreen() WizardScreen {
	return PreviewScreen
}

func (s *PreviewState) CanGoBack() bool {
	return true
}

func (s *PreviewState) Render() string {
	return fmt.Sprintf("Preview - %d files (~%d bytes) to sync", s.estimatedFiles, s.estimatedSize)
}

func (s *PreviewState) GoBack() *ExclusionPatternsState {
	return &ExclusionPatternsState{
		syncMode:          s.syncMode,
		sourcePath:        s.sourcePath,
		destinationPath:   s.destinationPath,
		createDestination: s.createDestination,
		syncOptions:       s.syncOptions,
		selectedFolders:   s.selectedFolders,
	}
}

func (s *PreviewState) StartSync() *ProgressState {
	config := s.GetCompleteConfig()
	
	return &ProgressState{
		config:           config,
		filesTotal:       s.estimatedFiles,
		bytesTotal:       s.estimatedSize,
		filesProcessed:   0,
		bytesTransferred: 0,
		canCancel:        true,
	}
}

func (s *PreviewState) SaveAsSyncFile() (string, error) {
	config := s.GetCompleteConfig()
	
	// Generate SyncFile content
	syncFileContent := generateSyncFileContent(config)
	
	// Save to file
	filename := "WizardConfig.sync"
	if err := os.WriteFile(filename, []byte(syncFileContent), 0644); err != nil {
		return "", fmt.Errorf("failed to save SyncFile: %w", err)
	}
	
	return filename, nil
}

// ProgressState implementations
func (s *ProgressState) GetCurrentScreen() WizardScreen {
	return ProgressScreen
}

func (s *ProgressState) CanGoBack() bool {
	return false // Cannot go back during sync
}

func (s *ProgressState) Render() string {
	progress := float64(s.filesProcessed) / float64(s.filesTotal) * 100
	return fmt.Sprintf("Progress - %.1f%% (%d/%d files)", progress, s.filesProcessed, s.filesTotal)
}

func (s *ProgressState) WaitForCompletion() *CompletedState {
	// In a real implementation, this would wait for actual sync completion
	// For now, simulate completion
	return &CompletedState{
		config:       s.config,
		successful:   true,
		filesSync:    s.filesTotal,
		errors:       nil,
		syncDuration: "2m 30s",
	}
}

func (s *ProgressState) CancelSync() (*CompletedState, error) {
	if !s.canCancel {
		return nil, fmt.Errorf("sync cannot be cancelled at this time")
	}
	
	return &CompletedState{
		config:       s.config,
		successful:   false,
		filesSync:    s.filesProcessed,
		errors:       []error{fmt.Errorf("sync cancelled by user")},
		syncDuration: "cancelled",
	}, nil
}

// CompletedState implementations
func (s *CompletedState) GetCurrentScreen() WizardScreen {
	return CompletedScreen
}

func (s *CompletedState) CanGoBack() bool {
	return false // Completed state is terminal
}

func (s *CompletedState) Render() string {
	if s.successful {
		return fmt.Sprintf("Sync Complete - %d files synced in %s", s.filesSync, s.syncDuration)
	}
	return fmt.Sprintf("Sync Failed - %d files processed, %d errors", s.filesSync, len(s.errors))
}

func (s *CompletedState) Exit() bool {
	return true // Signal that wizard should exit
}

// Helper functions

// scanDirectoryForStats builds folder statistics for the directory filter
func scanDirectoryForStats(sourcePath string) map[string]FolderInfo {
	stats := make(map[string]FolderInfo)
	
	// In a real implementation, this would scan the directory tree
	// For now, return empty stats
	return stats
}

// calculateEstimates estimates files and size based on selections and exclusions
func calculateEstimates(sourcePath string, selectedFolders map[string]bool, exclusionPatterns []string) (int, int64) {
	// In a real implementation, this would calculate actual estimates
	// For now, return mock data
	return 100, 1024 * 1024 // 100 files, 1MB
}

// generateSyncFileContent creates SyncFile content from wizard configuration
func generateSyncFileContent(config *CompleteWizardConfig) string {
	content := fmt.Sprintf("SYNC %s %s\n", config.SourcePath, config.DestinationPath)
	content += fmt.Sprintf("MODE %s\n", config.SyncMode.String())
	content += fmt.Sprintf("DRYRUN %t\n", config.SyncOptions.DryRun)
	content += fmt.Sprintf("VERBOSE %t\n", config.SyncOptions.Verbose)
	content += fmt.Sprintf("GITIGNORE %t\n", config.SyncOptions.UseGitignore)
	
	// Add selected folders as ONLY directives
	for folder, selected := range config.SelectedFolders {
		if selected {
			content += fmt.Sprintf("ONLY %s\n", folder)
		}
	}
	
	// Add exclusion patterns
	for _, pattern := range config.ExclusionPatterns {
		content += fmt.Sprintf("EXCLUDE %s\n", pattern)
	}
	
	return content
}