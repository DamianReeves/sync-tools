package rsync

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DamianReeves/sync-tools/internal/filters"
	"github.com/DamianReeves/sync-tools/internal/logging"
)

// Options holds all the options for running rsync
type Options struct {
	Source              string
	Dest                string
	Mode                string
	DryRun              bool
	UseSourceGitignore  bool
	ExcludeHiddenDirs   bool
	OnlySyncignore      bool
	IgnoreSrc           []string
	IgnoreDest          []string
	Only                []string
	LogLevel            string
	LogFile             string
	LogFormat           string
	DumpCommands        string
	Report              string
	ListFiltered        string
	Interactive         bool
	Patch               string
	ApplyPatch          bool
	Yes                 bool
	Preview             bool
	// Interactive sync plan fields
	Plan                string
	ApplyPlan           string
	IncludeChanges      []string
	ExcludeChanges      []string
	Editor              string // Custom editor for interactive plan editing
	// Conflict resolution fields
	ConflictStrategy        string // Default conflict resolution strategy: newest-wins, source-wins, dest-wins, backup
	SkipConflicts          bool   // Skip conflicting files during plan execution  
	GenerateConflictPlan   string // Generate a separate plan file containing only conflicts
}

// Runner handles rsync operations
type Runner struct {
	logger logging.Logger
}

// NewRunner creates a new rsync runner
func NewRunner(logger logging.Logger) *Runner {
	return &Runner{
		logger: logger,
	}
}

// Sync performs the synchronization operation
func (r *Runner) Sync(opts *Options) error {
	// Check if preview mode is requested
	if opts.Preview {
		return r.showPreview(opts)
	}
	
	// Check if patch mode is requested (either via --patch flag or --report with .patch extension)
	if opts.Patch != "" {
		r.logger.Infof("Starting patch generation: %s -> %s (output: %s, dry-run: %v)",
			opts.Source, opts.Dest, opts.Patch, opts.DryRun)
		return r.generatePatch(opts)
	}
	
	// Check if report with patch format is requested (based on file extension)
	if opts.Report != "" && (strings.HasSuffix(strings.ToLower(opts.Report), ".patch") || strings.HasSuffix(strings.ToLower(opts.Report), ".diff")) {
		r.logger.Infof("Starting patch report generation: %s -> %s (output: %s, dry-run: %v)",
			opts.Source, opts.Dest, opts.Report, opts.DryRun)
		// Use the report path as patch path
		patchOpts := *opts
		patchOpts.Patch = opts.Report
		return r.generatePatch(&patchOpts)
	}
	
	// Check if markdown report is requested
	if opts.Report != "" && (strings.HasSuffix(strings.ToLower(opts.Report), ".md") || strings.HasSuffix(strings.ToLower(opts.Report), ".markdown")) {
		r.logger.Infof("Starting markdown report generation: %s -> %s (output: %s, dry-run: %v)",
			opts.Source, opts.Dest, opts.Report, opts.DryRun)
		return r.generateMarkdownReport(opts)
	}

	r.logger.Infof("Starting sync: %s -> %s (mode: %s, dry-run: %v)",
		opts.Source, opts.Dest, opts.Mode, opts.DryRun)

	switch opts.Mode {
	case "one-way":
		return r.runOneWay(opts)
	case "two-way":
		return r.runTwoWay(opts)
	default:
		return fmt.Errorf("unsupported mode: %s", opts.Mode)
	}
}

// runOneWay performs one-way synchronization
func (r *Runner) runOneWay(opts *Options) error {
	// Build filter files
	sourceFilter, err := r.buildSourceFilter(opts)
	if err != nil {
		return fmt.Errorf("error building source filter: %w", err)
	}
	defer r.cleanupTempFile(sourceFilter)

	var destFilter string
	if len(opts.IgnoreDest) > 0 {
		destFilter, err = r.buildDestFilter(opts)
		if err != nil {
			return fmt.Errorf("error building dest filter: %w", err)
		}
		defer r.cleanupTempFile(destFilter)
	}

	// Build rsync command
	cmd := r.buildRsyncCommand(opts, sourceFilter, destFilter)

	// Execute rsync
	return r.executeRsync(cmd, opts)
}

// runTwoWay performs two-way synchronization
func (r *Runner) runTwoWay(opts *Options) error {
	// Two-way sync involves conflict detection and resolution
	r.logger.Info("Performing two-way sync with conflict detection")

	// First, detect conflicts
	conflicts, err := r.detectConflicts(opts)
	if err != nil {
		return fmt.Errorf("error detecting conflicts: %w", err)
	}

	if len(conflicts) > 0 {
		r.logger.Warnf("Found %d conflicts, preserving destination versions as conflict files", len(conflicts))
		if err := r.preserveConflicts(conflicts, opts); err != nil {
			return fmt.Errorf("error preserving conflicts: %w", err)
		}
	}

	// Then perform one-way sync
	return r.runOneWay(opts)
}

// buildSourceFilter creates the source-side filter file
func (r *Runner) buildSourceFilter(opts *Options) (string, error) {
	var patterns []string

	// Add default exclusions
	patterns = append(patterns, "/.git/")

	if opts.ExcludeHiddenDirs {
		patterns = append(patterns, "/.*")
	}

	// Add .syncignore patterns from source
	if !opts.OnlySyncignore {
		syncignoreFile := filepath.Join(opts.Source, ".syncignore")
		if _, err := os.Stat(syncignoreFile); err == nil {
			ignorePatterns, err := r.readIgnoreFile(syncignoreFile)
			if err != nil {
				return "", err
			}
			patterns = append(patterns, ignorePatterns...)
		}

		// Add .gitignore patterns if requested
		if opts.UseSourceGitignore {
			gitignoreFile := filepath.Join(opts.Source, ".gitignore")
			if _, err := os.Stat(gitignoreFile); err == nil {
				ignorePatterns, err := r.readIgnoreFile(gitignoreFile)
				if err != nil {
					return "", err
				}
				patterns = append(patterns, ignorePatterns...)
			}
		}
	}

	// Add CLI ignore patterns
	patterns = append(patterns, opts.IgnoreSrc...)

	// Handle whitelist mode
	if len(opts.Only) > 0 {
		return filters.BuildOnlyFilter(opts.Only)
	}

	return filters.BuildExcludeFilter(patterns)
}

// buildDestFilter creates the destination-side filter file
func (r *Runner) buildDestFilter(opts *Options) (string, error) {
	return filters.BuildExcludeFilter(opts.IgnoreDest)
}

// buildRsyncCommand constructs the rsync command
func (r *Runner) buildRsyncCommand(opts *Options, sourceFilter, destFilter string) *exec.Cmd {
	// Ensure source path ends with / for proper rsync behavior
	source := opts.Source
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	
	// Build rsync command using unified builder
	cmdOpts := &RsyncCommandOptions{
		UseChecksum:      true, // Always use checksum for consistency
		UseDelete:        true, // Remove files from dest that don't exist in source
		UseVerbose:       true,
		UseHumanReadable: true,
		SourceFilter:     sourceFilter,
		DestFilter:       destFilter,
		Source:           source,
		Dest:             opts.Dest,
		DryRun:           opts.DryRun,
	}
	
	args := r.buildRsyncArgs(cmdOpts)
	args = append(args, source, opts.Dest)

	return exec.Command("rsync", args...)
}

// RsyncCommandOptions configures rsync command building
type RsyncCommandOptions struct {
	UseChecksum    bool
	UseDelete      bool 
	UseVerbose     bool
	UseHumanReadable bool
	SourceFilter   string
	DestFilter     string
	Source         string
	Dest           string
	DryRun         bool
}

// buildRsyncArgs creates a unified rsync command args array
func (r *Runner) buildRsyncArgs(opts *RsyncCommandOptions) []string {
	args := []string{"--archive"} // Always use archive mode
	
	if opts.UseChecksum {
		args = append(args, "--checksum")
	}
	if opts.UseVerbose {
		args = append(args, "--verbose")
	}
	if opts.UseHumanReadable {
		args = append(args, "--human-readable") 
	}
	if opts.UseDelete {
		args = append(args, "--delete")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	
	// Add filter files
	if opts.SourceFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", opts.SourceFilter))
	}
	if opts.DestFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", opts.DestFilter))
	}
	
	return args
}

// executeRsync runs the rsync command
func (r *Runner) executeRsync(cmd *exec.Cmd, opts *Options) error {
	r.logger.Debugf("Executing rsync command: %s", strings.Join(cmd.Args, " "))

	// Set up output capturing
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Read and log output
	go r.logOutput(stdout, "STDOUT")
	go r.logOutput(stderr, "STDERR")

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("rsync command failed: %w", err)
	}

	r.logger.Info("Sync completed successfully")
	return nil
}

// logOutput logs command output line by line
func (r *Runner) logOutput(reader io.ReadCloser, prefix string) {
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			r.logger.Infof("[%s] %s", prefix, line)
		}
	}
}

// readIgnoreFile reads patterns from an ignore file
func (r *Runner) readIgnoreFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// detectConflicts finds files that have changed on both sides
func (r *Runner) detectConflicts(opts *Options) ([]string, error) {
	// This is a simplified conflict detection
	// In practice, this would compare file modification times and checksums
	r.logger.Debug("Detecting conflicts (simplified implementation)")
	return []string{}, nil
}

// preserveConflicts creates conflict copies of files
func (r *Runner) preserveConflicts(conflicts []string, opts *Options) error {
	timestamp := time.Now().Unix()
	for _, conflict := range conflicts {
		conflictName := fmt.Sprintf("%s.conflict-%d", conflict, timestamp)
		r.logger.Infof("Creating conflict file: %s", conflictName)
		// Implementation would copy the file with the conflict name
	}
	return nil
}

// cleanupTempFile removes temporary filter files
func (r *Runner) cleanupTempFile(filename string) {
	if filename != "" {
		if err := os.Remove(filename); err != nil {
			r.logger.Debugf("Failed to remove temp file %s: %v", filename, err)
		}
	}
}

// generatePatch creates a git patch file instead of syncing
func (r *Runner) generatePatch(opts *Options) error {
	if opts.DryRun {
		r.logger.Infof("Would generate patch file: %s", opts.Patch)
		r.logger.Infof("Would include changes from %s to %s", opts.Source, opts.Dest)
		return nil
	}

	// Build filter files to respect ignore patterns
	sourceFilter, err := r.buildSourceFilter(opts)
	if err != nil {
		return fmt.Errorf("error building source filter: %w", err)
	}
	defer r.cleanupTempFile(sourceFilter)

	// Create the patch file
	patchFile, err := os.Create(opts.Patch)
	if err != nil {
		return fmt.Errorf("error creating patch file: %w", err)
	}
	defer patchFile.Close()

	// Write patch header
	fmt.Fprintf(patchFile, "# Git patch generated by sync-tools\n")
	fmt.Fprintf(patchFile, "# Source: %s\n", opts.Source)
	fmt.Fprintf(patchFile, "# Destination: %s\n", opts.Dest)
	fmt.Fprintf(patchFile, "# Generated: %s\n\n", time.Now().Format(time.RFC3339))

	// Use git diff to generate the patch
	cmd := exec.Command("git", "diff", "--no-index", "--no-prefix", opts.Dest, opts.Source)
	cmd.Dir = filepath.Dir(opts.Source)
	
	output, err := cmd.CombinedOutput()
	// git diff returns exit code 1 when there are differences, which is expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// This is expected when there are differences
		} else {
			r.logger.Debugf("git diff command failed: %v", err)
			// Fall back to a simple diff if git is not available or fails
			return r.generateSimplePatch(opts, patchFile)
		}
	}

	// Write the diff output to the patch file
	if _, err := patchFile.Write(output); err != nil {
		return fmt.Errorf("error writing patch content: %w", err)
	}

	r.logger.Infof("Patch file generated: %s", opts.Patch)
	
	// Apply patch if requested
	if opts.ApplyPatch {
		return r.applyPatchWithConfirmation(opts)
	}
	
	return nil
}

// generateSimplePatch creates a basic patch when git is not available
func (r *Runner) generateSimplePatch(opts *Options, patchFile *os.File) error {
	fmt.Fprintf(patchFile, "# Simple patch (git not available)\n")
	fmt.Fprintf(patchFile, "# Files would be synchronized from %s to %s\n", opts.Source, opts.Dest)
	
	// Just create a basic listing of files that would be synced
	err := filepath.Walk(opts.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(opts.Source, path)
			if err != nil {
				return err
			}
			fmt.Fprintf(patchFile, "# Would sync: %s\n", relPath)
		}
		return nil
	})
	
	return err
}

// applyPatchWithConfirmation applies the generated patch with user confirmation
func (r *Runner) applyPatchWithConfirmation(opts *Options) error {
	patchPath := opts.Patch
	if opts.Report != "" && opts.Patch == "" {
		// When using --report with patch format, the patch path is in Report field
		patchPath = opts.Report
	}
	
	// Check if patch file exists
	if _, err := os.Stat(patchPath); os.IsNotExist(err) {
		return fmt.Errorf("patch file does not exist: %s", patchPath)
	}
	
	// Show patch preview
	r.logger.Info("Patch contents:")
	previewCmd := exec.Command("git", "apply", "--stat", patchPath)
	previewOutput, err := previewCmd.CombinedOutput()
	if err != nil {
		r.logger.Warnf("Could not preview patch stats: %v", err)
		// Fallback to showing first few lines of the patch
		if patchContent, readErr := os.ReadFile(patchPath); readErr == nil {
			lines := strings.Split(string(patchContent), "\n")
			maxLines := 10
			if len(lines) > maxLines {
				lines = lines[:maxLines]
				lines = append(lines, "... (truncated)")
			}
			r.logger.Info(strings.Join(lines, "\n"))
		}
	} else {
		r.logger.Info(string(previewOutput))
	}
	
	// Get confirmation unless --yes flag is used
	if !opts.Yes {
		if !r.confirmPatchApplication(patchPath) {
			r.logger.Info("Patch application cancelled by user")
			return nil
		}
	} else {
		r.logger.Info("Auto-confirming patch application (--yes flag)")
	}
	
	// Apply the patch
	r.logger.Infof("Applying patch: %s", patchPath)
	
	// Convert to absolute path if needed
	absPatchPath, err := filepath.Abs(patchPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for patch: %w", err)
	}
	
	applyCmd := exec.Command("git", "apply", absPatchPath)
	applyCmd.Dir = opts.Dest
	
	if output, err := applyCmd.CombinedOutput(); err != nil {
		r.logger.Errorf("Failed to apply patch: %v", err)
		r.logger.Errorf("Git apply output: %s", string(output))
		return fmt.Errorf("failed to apply patch: %w", err)
	}
	
	r.logger.Info("Patch applied successfully!")
	return nil
}

// confirmPatchApplication prompts the user for confirmation
func (r *Runner) confirmPatchApplication(patchPath string) bool {
	fmt.Printf("\nDo you want to apply the patch '%s'? [y/N]: ", patchPath)
	
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// showPreview generates and displays a colored diff preview
func (r *Runner) showPreview(opts *Options) error {
	r.logger.Infof("Generating preview: %s -> %s",
		opts.Source, opts.Dest)
	
	// Generate diff using git diff with color
	cmd := exec.Command("git", "diff", "--no-index", "--no-prefix", "--color=always", opts.Dest, opts.Source)
	cmd.Dir = filepath.Dir(opts.Source)
	
	output, err := cmd.CombinedOutput()
	// git diff returns exit code 1 when there are differences, which is expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// This is expected when there are differences
		} else {
			r.logger.Debugf("git diff command failed: %v", err)
			// Fall back to a simple diff if git is not available
			return r.showSimplePreview(opts)
		}
	}
	
	// If there's no output, there are no differences
	if len(output) == 0 {
		r.logger.Info("No differences found between source and destination")
		return nil
	}
	
	// Check if less is available for paging
	if _, err := exec.LookPath("less"); err == nil {
		// Use less with options similar to git diff
		lessCmd := exec.Command("less", "-R", "-FX")
		lessCmd.Stdin = strings.NewReader(string(output))
		lessCmd.Stdout = os.Stdout
		lessCmd.Stderr = os.Stderr
		
		r.logger.Debug("Displaying preview with pager (press 'q' to quit)")
		return lessCmd.Run()
	} else {
		// Fall back to direct output if less is not available
		r.logger.Debug("Pager not available, displaying preview directly")
		fmt.Print(string(output))
		return nil
	}
}

// showSimplePreview shows a simple preview when git is not available
func (r *Runner) showSimplePreview(opts *Options) error {
	// Use rsync's dry-run to show what would be changed
	r.logger.Info("Git not available, showing rsync dry-run preview:")
	
	// Build filter files
	sourceFilter, err := r.buildSourceFilter(opts)
	if err != nil {
		return fmt.Errorf("error building source filter: %w", err)
	}
	defer r.cleanupTempFile(sourceFilter)
	
	var destFilter string
	if len(opts.IgnoreDest) > 0 {
		destFilter, err = r.buildDestFilter(opts)
		if err != nil {
			return fmt.Errorf("error building dest filter: %w", err)
		}
		defer r.cleanupTempFile(destFilter)
	}
	
	// Build rsync command with dry-run and itemize changes
	args := []string{
		"--archive",
		"--verbose",
		"--human-readable",
		"--delete",
		"--dry-run",
		"--itemize-changes",
	}
	
	// Add filter files
	if sourceFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", sourceFilter))
	}
	if destFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", destFilter))
	}
	
	// Add source and destination
	source := opts.Source
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	args = append(args, source, opts.Dest)
	
	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(err.Error(), "exit status 23") {
		// Exit status 23 is partial transfer due to error, often from dry-run
		return fmt.Errorf("rsync preview failed: %w", err)
	}
	
	fmt.Print(string(output))
	return nil
}

// SyncChange represents a single file change in the sync operation
type SyncChange struct {
	Action   string // "create", "update", "delete"
	Path     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
}

// SyncReport contains the report data for a sync operation
type SyncReport struct {
	Source      string
	Destination string
	Mode        string
	DryRun      bool
	Timestamp   time.Time
	Changes     []SyncChange
	Stats       SyncStats
}

// SyncStats contains statistics about the sync operation
type SyncStats struct {
	FilesCreated   int
	FilesUpdated   int
	FilesDeleted   int
	DirsCreated    int
	DirsDeleted    int
	TotalSize      int64
	FilteredCount  int
}

// generateMarkdownReport creates a markdown report of the sync operation
func (r *Runner) generateMarkdownReport(opts *Options) error {
	r.logger.Debug("Generating markdown report")
	
	// Collect sync information using dry-run
	report, err := r.collectSyncInfo(opts)
	if err != nil {
		return fmt.Errorf("error collecting sync information: %w", err)
	}
	
	// Write the markdown report
	if err := r.writeMarkdownReport(report, opts.Report); err != nil {
		return fmt.Errorf("error writing markdown report: %w", err)
	}
	
	r.logger.Infof("Markdown report generated: %s", opts.Report)
	
	// If not in dry-run mode, also perform the actual sync
	if !opts.DryRun {
		r.logger.Info("Performing actual sync after report generation")
		switch opts.Mode {
		case "one-way":
			return r.runOneWay(opts)
		case "two-way":
			return r.runTwoWay(opts)
		default:
			return fmt.Errorf("unsupported mode: %s", opts.Mode)
		}
	}
	
	return nil
}

// collectSyncInfo gathers information about what would be synced
func (r *Runner) collectSyncInfo(opts *Options) (*SyncReport, error) {
	report := &SyncReport{
		Source:      opts.Source,
		Destination: opts.Dest,
		Mode:        opts.Mode,
		DryRun:      opts.DryRun,
		Timestamp:   time.Now(),
		Changes:     []SyncChange{},
	}
	
	// Build filter files
	sourceFilter, err := r.buildSourceFilter(opts)
	if err != nil {
		return nil, fmt.Errorf("error building source filter: %w", err)
	}
	defer r.cleanupTempFile(sourceFilter)
	
	var destFilter string
	if len(opts.IgnoreDest) > 0 {
		destFilter, err = r.buildDestFilter(opts)
		if err != nil {
			return nil, fmt.Errorf("error building dest filter: %w", err)
		}
		defer r.cleanupTempFile(destFilter)
	}
	
	// Build rsync command with dry-run and itemize changes
	args := []string{
		"--archive",
		"--verbose",
		"--human-readable",
		"--delete",
		"--dry-run",
		"--itemize-changes",
		"--out-format=%i %n %L %l %t",
	}
	
	// Add filter files
	if sourceFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", sourceFilter))
	}
	if destFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", destFilter))
	}
	
	// Add source and destination
	source := opts.Source
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	args = append(args, source, opts.Dest)
	
	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(err.Error(), "exit status 23") {
		// Exit status 23 is partial transfer due to error, often from dry-run
		return nil, fmt.Errorf("rsync dry-run failed: %w", err)
	}
	
	// Parse rsync output to extract changes
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Skip rsync header and summary lines
		if strings.HasPrefix(line, "sending incremental") ||
			strings.HasPrefix(line, "sent ") ||
			strings.HasPrefix(line, "total size") ||
			strings.HasPrefix(line, "bytes/sec") {
			continue
		}
		
		// Parse itemize-changes format - lines starting with *, >, <, c, etc.
		if r.isItemizedChange(line) {
			change := r.parseRsyncChange(line)
			if change != nil {
				report.Changes = append(report.Changes, *change)
				
				// Update statistics
				switch change.Action {
				case "create":
					if change.IsDir {
						report.Stats.DirsCreated++
					} else {
						report.Stats.FilesCreated++
						report.Stats.TotalSize += change.Size
					}
				case "update":
					report.Stats.FilesUpdated++
					report.Stats.TotalSize += change.Size
				case "delete":
					if change.IsDir {
						report.Stats.DirsDeleted++
					} else {
						report.Stats.FilesDeleted++
					}
				}
			}
		}
	}
	
	return report, nil
}

// FileInfo represents information about a file
type FileInfo struct {
	RelativePath string
	AbsolutePath string
	Size         int64
	ModTime      time.Time
	IsDir        bool
}

// collectSyncInfoComprehensive performs comprehensive bi-directional analysis
func (r *Runner) collectSyncInfoComprehensive(opts *Options) (*SyncReport, error) {
	report := &SyncReport{
		Source:      opts.Source,
		Destination: opts.Dest,
		Mode:        opts.Mode,
		DryRun:      opts.DryRun,
		Timestamp:   time.Now(),
		Changes:     []SyncChange{},
	}
	
	// Get file listings from both directories
	sourceFiles, err := r.getFileList(opts.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to list source files: %w", err)
	}
	
	destFiles, err := r.getFileList(opts.Dest)
	if err != nil {
		return nil, fmt.Errorf("failed to list dest files: %w", err)
	}
	
	// Create maps for efficient lookup
	sourceMap := make(map[string]FileInfo)
	for _, file := range sourceFiles {
		sourceMap[file.RelativePath] = file
	}
	
	destMap := make(map[string]FileInfo)
	for _, file := range destFiles {
		destMap[file.RelativePath] = file
	}
	
	// Find all unique paths
	allPaths := make(map[string]bool)
	for path := range sourceMap {
		allPaths[path] = true
	}
	for path := range destMap {
		allPaths[path] = true
	}
	
	// Analyze each path
	for path := range allPaths {
		sourceFile, inSource := sourceMap[path]
		destFile, inDest := destMap[path]
		
		change := r.analyzeFileChange(path, sourceFile, destFile, inSource, inDest)
		if change != nil {
			report.Changes = append(report.Changes, *change)
		}
	}
	
	// Update statistics
	for _, change := range report.Changes {
		switch change.Action {
		case "create":
			if change.IsDir {
				report.Stats.DirsCreated++
			} else {
				report.Stats.FilesCreated++
				report.Stats.TotalSize += change.Size
			}
		case "update":
			report.Stats.FilesUpdated++
			report.Stats.TotalSize += change.Size
		case "delete":
			if change.IsDir {
				report.Stats.DirsDeleted++
			} else {
				report.Stats.FilesDeleted++
			}
		}
	}
	
	return report, nil
}

// getFileList recursively lists all files in a directory
func (r *Runner) getFileList(dirPath string) ([]FileInfo, error) {
	var files []FileInfo
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if path == dirPath {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		
		files = append(files, FileInfo{
			RelativePath: filepath.ToSlash(relPath), // Use forward slashes for consistency
			AbsolutePath: path,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsDir:        info.IsDir(),
		})
		
		return nil
	})
	
	return files, err
}

// analyzeFileChange determines what type of change a file represents
func (r *Runner) analyzeFileChange(path string, sourceFile, destFile FileInfo, inSource, inDest bool) *SyncChange {
	if !inSource && !inDest {
		return nil // Shouldn't happen
	}
	
	change := &SyncChange{
		Path: path,
	}
	
	if inSource && !inDest {
		// File only exists in source
		if sourceFile.IsDir {
			return nil // Skip directories
		}
		change.Action = "create"
		change.Size = sourceFile.Size
		change.ModTime = sourceFile.ModTime
		change.IsDir = sourceFile.IsDir
	} else if !inSource && inDest {
		// File only exists in dest
		if destFile.IsDir {
			return nil // Skip directories
		}
		change.Action = "delete"
		change.Size = destFile.Size
		change.ModTime = destFile.ModTime
		change.IsDir = destFile.IsDir
	} else {
		// File exists in both - check for differences
		change.IsDir = sourceFile.IsDir
		
		if sourceFile.IsDir && destFile.IsDir {
			// Both are directories - generally no action needed
			return nil // Skip directories for now
		} else if !sourceFile.IsDir && !destFile.IsDir {
			// Both are files - compare content and timestamps
			if r.filesAreIdentical(sourceFile.AbsolutePath, destFile.AbsolutePath) {
				// Files are identical - no change needed
				return nil
			}
			
			// Files differ - determine if it's an update or conflict
			sourceMod := sourceFile.ModTime
			destMod := destFile.ModTime
			
			// Use source file info for the change
			change.Size = sourceFile.Size
			change.ModTime = sourceMod
			
			if sourceMod.After(destMod) {
				change.Action = "update"
			} else if destMod.After(sourceMod) {
				// Dest is newer - potential conflict
				change.Action = "conflict"
			} else {
				// Same modification time but different content - conflict
				change.Action = "conflict"
			}
		} else {
			// One is file, one is directory - conflict
			change.Action = "conflict"
			change.Size = sourceFile.Size
			change.ModTime = sourceFile.ModTime
		}
	}
	
	return change
}

// filesAreIdentical checks if two files have identical content
func (r *Runner) filesAreIdentical(path1, path2 string) bool {
	// Read and compare file content
	content1, err1 := os.ReadFile(path1)
	content2, err2 := os.ReadFile(path2)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	// Compare actual content
	return string(content1) == string(content2)
}

// parseRsyncChange parses a line from rsync's itemize-changes output
func (r *Runner) parseRsyncChange(line string) *SyncChange {
	// Parse rsync output format: flags filename size timestamp
	// Examples:
	// >f+++++++++ main.js  18 2025/08/30 01:11:42
	// *deleting   config.yml  0 2025/08/30 01:11:42
	// cd+++++++++ dir/
	
	parts := strings.Fields(line)
	if len(parts) < 4 { // Need at least flags, filename, size, timestamp
		return nil
	}
	
	flags := parts[0]
	filename := parts[1]
	size := parts[2]
	// Join remaining parts for timestamp (could be date and time)
	timestamp := strings.Join(parts[3:], " ")
	
	// Parse the flags to determine action and type
	change := &SyncChange{
		Path:   filename,
		ModTime: r.parseTimestamp(timestamp),
		Size:    r.parseSize(size),
	}
	
	// Determine action and type based on flags
	if strings.HasPrefix(flags, "*deleting") {
		change.Action = "delete"
		// Check if it's a directory by file extension or other heuristics
		change.IsDir = !strings.Contains(filename, ".")
	} else if len(flags) >= 2 {
		// Regular itemize format: YXcstpoguax
		if flags[1] == 'f' {
			change.IsDir = false
			if strings.Contains(flags, "+") {
				change.Action = "create"
			} else {
				change.Action = "update"
			}
		} else if flags[1] == 'd' {
			change.IsDir = true
			if strings.Contains(flags, "+") {
				change.Action = "create"
			} else {
				change.Action = "update"
			}
		}
	} else {
		// Fallback - assume file update
		change.IsDir = false
		change.Action = "update"
	}
	
	return change
}

// isItemizedChange checks if a line looks like an itemized change from rsync
func (r *Runner) isItemizedChange(line string) bool {
	// Itemized changes start with specific characters
	// Examples: >f+++++++++, *deleting, cd+++++++++, .f...t.....
	if len(line) < 3 {
		return false
	}
	
	// Check for common itemized change patterns
	firstChar := line[0]
	return firstChar == '>' || firstChar == '<' || firstChar == '*' || 
		   firstChar == 'c' || firstChar == '.' || 
		   strings.HasPrefix(line, "*deleting")
}

// parseTimestamp converts rsync timestamp to time.Time
func (r *Runner) parseTimestamp(timestamp string) time.Time {
	// Format: "2025/08/30 01:11:42"
	if t, err := time.Parse("2006/01/02 15:04:05", timestamp); err == nil {
		return t
	}
	return time.Time{}
}

// parseSize converts size string to int64
func (r *Runner) parseSize(sizeStr string) int64 {
	if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return size
	}
	return 0
}

// writeMarkdownReport writes the sync report in markdown format
func (r *Runner) writeMarkdownReport(report *SyncReport, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	w := bufio.NewWriter(file)
	defer w.Flush()
	
	// Write header
	fmt.Fprintf(w, "# Sync Report\n\n")
	fmt.Fprintf(w, "Generated by sync-tools on %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Write configuration section
	fmt.Fprintf(w, "## Configuration\n\n")
	fmt.Fprintf(w, "| Setting | Value |\n")
	fmt.Fprintf(w, "|---------|-------|\n")
	fmt.Fprintf(w, "| Source | `%s` |\n", report.Source)
	fmt.Fprintf(w, "| Destination | `%s` |\n", report.Destination)
	fmt.Fprintf(w, "| Mode | %s |\n", report.Mode)
	fmt.Fprintf(w, "| Dry Run | %v |\n", report.DryRun)
	fmt.Fprintf(w, "\n")
	
	// Write statistics section
	fmt.Fprintf(w, "## Summary Statistics\n\n")
	fmt.Fprintf(w, "| Metric | Count |\n")
	fmt.Fprintf(w, "|--------|-------|\n")
	fmt.Fprintf(w, "| Files Created | %d |\n", report.Stats.FilesCreated)
	fmt.Fprintf(w, "| Files Updated | %d |\n", report.Stats.FilesUpdated)
	fmt.Fprintf(w, "| Files Deleted | %d |\n", report.Stats.FilesDeleted)
	fmt.Fprintf(w, "| Directories Created | %d |\n", report.Stats.DirsCreated)
	fmt.Fprintf(w, "| Directories Deleted | %d |\n", report.Stats.DirsDeleted)
	fmt.Fprintf(w, "| Total Size | %s |\n", r.formatSize(report.Stats.TotalSize))
	fmt.Fprintf(w, "| Total Changes | %d |\n", len(report.Changes))
	fmt.Fprintf(w, "\n")
	
	// Write changes section
	if len(report.Changes) > 0 {
		fmt.Fprintf(w, "## Changes\n\n")
		
		// Group changes by action
		creates := []SyncChange{}
		updates := []SyncChange{}
		deletes := []SyncChange{}
		
		for _, change := range report.Changes {
			switch change.Action {
			case "create":
				creates = append(creates, change)
			case "update":
				updates = append(updates, change)
			case "delete":
				deletes = append(deletes, change)
			}
		}
		
		// Write creates
		if len(creates) > 0 {
			fmt.Fprintf(w, "### Files/Directories to Create (%d)\n\n", len(creates))
			for _, change := range creates {
				if change.IsDir {
					fmt.Fprintf(w, "- 📁 `%s/`\n", change.Path)
				} else {
					fmt.Fprintf(w, "- 📄 `%s` (%s)\n", change.Path, r.formatSize(change.Size))
				}
			}
			fmt.Fprintf(w, "\n")
		}
		
		// Write updates
		if len(updates) > 0 {
			fmt.Fprintf(w, "### Files to Update (%d)\n\n", len(updates))
			for _, change := range updates {
				fmt.Fprintf(w, "- 🔄 `%s` (%s)\n", change.Path, r.formatSize(change.Size))
			}
			fmt.Fprintf(w, "\n")
		}
		
		// Write deletes
		if len(deletes) > 0 {
			fmt.Fprintf(w, "### Files/Directories to Delete (%d)\n\n", len(deletes))
			for _, change := range deletes {
				if change.IsDir {
					fmt.Fprintf(w, "- ❌ `%s/`\n", change.Path)
				} else {
					fmt.Fprintf(w, "- ❌ `%s`\n", change.Path)
				}
			}
			fmt.Fprintf(w, "\n")
		}
	} else {
		fmt.Fprintf(w, "## Changes\n\n")
		fmt.Fprintf(w, "No changes detected.\n\n")
	}
	
	// Write footer
	fmt.Fprintf(w, "---\n")
	fmt.Fprintf(w, "*Report generated by [sync-tools](https://github.com/DamianReeves/sync-tools)*\n")
	
	return nil
}

// formatSize formats a size in bytes to a human-readable string
func (r *Runner) formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GeneratePlan creates a sync plan file for two-phased interactive sync
func (r *Runner) GeneratePlan(opts *Options) error {
	r.logger.Infof("Generating sync plan: %s", opts.Plan)
	
	// Collect sync information using comprehensive analysis
	syncInfo, err := r.collectSyncInfoComprehensive(opts)
	if err != nil {
		return fmt.Errorf("failed to analyze sync operations: %w", err)
	}
	
	// Generate plan content from analysis
	planContent, err := r.generatePlanContent(opts, syncInfo)
	if err != nil {
		return fmt.Errorf("failed to generate plan content: %w", err)
	}
	
	// Write plan file
	err = os.WriteFile(opts.Plan, []byte(planContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}
	
	r.logger.Infof("Plan file created successfully: %s", opts.Plan)
	
	// If interactive mode is enabled, open in editor
	if opts.Interactive {
		return r.openPlanInEditor(opts)
	}
	
	return nil
}

// PlanData represents the parsed content of a plan file
type PlanData struct {
	Source     string
	Dest       string
	Mode       string
	Operations []PlanOperation
}

// PlanOperation represents a single operation from a plan file
type PlanOperation struct {
	Alias string // <<, >>, <>, s2d, d2s, bid, etc.
	Type  string // "file" or "dir"
	Path  string
	Size  string
	Time  string
	Flags string
}

// ExecutePlan executes operations from a sync plan file
func (r *Runner) ExecutePlan(opts *Options) error {
	r.logger.Infof("Executing sync plan: %s", opts.ApplyPlan)
	
	// Validate plan file first
	if err := r.validatePlanFile(opts.ApplyPlan); err != nil {
		return err // This will include "invalid plan syntax" errors
	}
	
	// Read and parse plan file
	content, err := os.ReadFile(opts.ApplyPlan)
	if err != nil {
		return fmt.Errorf("failed to read plan file: %w", err)
	}
	
	if len(content) == 0 {
		return fmt.Errorf("plan file is empty")
	}
	
	// Parse plan file to extract metadata and operations
	planData, err := r.parsePlan(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse plan file: %w", err)
	}
	
	// Override source/dest with values from plan file if not provided
	if opts.Source == "" {
		opts.Source = planData.Source
	}
	if opts.Dest == "" {
		opts.Dest = planData.Dest
	}
	if opts.Mode == "" {
		opts.Mode = planData.Mode
	}
	
	// Resolve relative paths to absolute paths
	if opts.Source != "" {
		if absSource, err := filepath.Abs(opts.Source); err == nil {
			opts.Source = absSource
		}
	}
	if opts.Dest != "" {
		if absDest, err := filepath.Abs(opts.Dest); err == nil {
			opts.Dest = absDest
		}
	}
	
	// Execute each operation in the plan
	for _, op := range planData.Operations {
		if err := r.executePlanOperation(op, opts); err != nil {
			return fmt.Errorf("failed to execute operation %s %s: %w", op.Alias, op.Path, err)
		}
	}
	
	r.logger.Infof("Plan execution completed successfully: %d operations", len(planData.Operations))
	return nil
}

// parsePlan parses a plan file content and extracts metadata and operations
func (r *Runner) parsePlan(content string) (*PlanData, error) {
	lines := strings.Split(content, "\n")
	planData := &PlanData{
		Operations: []PlanOperation{},
	}
	
	// Parse metadata from comments
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Parse metadata comments
		if strings.HasPrefix(line, "# Source:") {
			planData.Source = strings.TrimSpace(strings.TrimPrefix(line, "# Source:"))
			continue
		}
		if strings.HasPrefix(line, "# Destination:") {
			planData.Dest = strings.TrimSpace(strings.TrimPrefix(line, "# Destination:"))
			continue
		}
		if strings.HasPrefix(line, "# Mode:") {
			planData.Mode = strings.TrimSpace(strings.TrimPrefix(line, "# Mode:"))
			continue
		}
		
		// Skip other comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse operation lines (format: alias type path size time flags)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			// Check if this looks like an operation line that's malformed
			if len(fields) > 0 && (strings.HasPrefix(fields[0], "<<") || strings.HasPrefix(fields[0], ">>") || strings.HasPrefix(fields[0], "<>")) {
				return nil, fmt.Errorf("invalid plan syntax: malformed operation line: %s", line)
			}
			continue // Skip other malformed lines
		}
		
		op := PlanOperation{
			Alias: fields[0],
			Type:  fields[1],
			Path:  fields[2],
		}
		
		// Parse optional fields
		if len(fields) > 3 {
			op.Size = fields[3]
		}
		if len(fields) > 4 {
			op.Time = fields[4]
		}
		if len(fields) > 5 {
			op.Flags = strings.Join(fields[5:], " ")
		}
		
		planData.Operations = append(planData.Operations, op)
	}
	
	return planData, nil
}

// executePlanOperation executes a single plan operation
func (r *Runner) executePlanOperation(op PlanOperation, opts *Options) error {
	r.logger.Infof("Executing: %s %s %s", op.Alias, op.Type, op.Path)
	
	// Determine sync direction based on alias
	var srcPath, destPath string
	switch op.Alias {
	case "<<", "s2d", "sync-to-dest":
		// Source to destination
		srcPath = filepath.Join(opts.Source, op.Path)
		destPath = filepath.Join(opts.Dest, op.Path)
	case ">>", "d2s", "dest-to-source":
		// Destination to source  
		srcPath = filepath.Join(opts.Dest, op.Path)
		destPath = filepath.Join(opts.Source, op.Path)
	case "<>", "bid", "bidirectional":
		// Bidirectional sync - handle conflicts by using newest-wins strategy by default
		return r.syncBidirectional(op, opts)
	default:
		return fmt.Errorf("unknown operation alias: %s", op.Alias)
	}
	
	// Execute the file operation using rsync or file operations
	if err := r.syncSingleFile(srcPath, destPath, op.Type == "dir"); err != nil {
		return fmt.Errorf("failed to sync %s: %w", op.Path, err)
	}
	
	return nil
}

// syncSingleFile syncs a single file or directory using rsync
func (r *Runner) syncSingleFile(srcPath, destPath string, isDir bool) error {
	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	if isDir {
		// For directories, ensure source ends with / and create dest directory
		if !strings.HasSuffix(srcPath, "/") {
			srcPath += "/"
		}
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}
	
	// Build rsync command using unified builder
	cmdOpts := &RsyncCommandOptions{
		UseChecksum: true, // Always use checksum for plan execution 
		Source:      srcPath,
		Dest:        destPath,
	}
	args := r.buildRsyncArgs(cmdOpts)
	args = append(args, srcPath, destPath)
	cmd := exec.Command("rsync", args...)
	
	r.logger.Debugf("Executing rsync command: %s", strings.Join(cmd.Args, " "))
	
	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w, output: %s", err, string(output))
	}
	
	return nil
}

// syncBidirectional performs bidirectional sync with conflict resolution
func (r *Runner) syncBidirectional(op PlanOperation, opts *Options) error {
	r.logger.Infof("Executing bidirectional sync: %s %s", op.Type, op.Path)
	
	srcPath := filepath.Join(opts.Source, op.Path)
	destPath := filepath.Join(opts.Dest, op.Path)
	
	// Check if both files exist
	srcInfo, srcErr := os.Stat(srcPath)
	destInfo, destErr := os.Stat(destPath)
	
	if srcErr != nil && destErr != nil {
		// Neither file exists - nothing to sync
		r.logger.Warnf("Neither source nor destination file exists: %s", op.Path)
		return nil
	}
	
	if srcErr != nil && destErr == nil {
		// Only destination exists - sync from dest to source
		r.logger.Infof("File only exists in destination, syncing to source: %s", op.Path)
		return r.syncSingleFile(destPath, srcPath, destInfo.IsDir())
	}
	
	if srcErr == nil && destErr != nil {
		// Only source exists - sync from source to dest
		r.logger.Infof("File only exists in source, syncing to destination: %s", op.Path)
		return r.syncSingleFile(srcPath, destPath, srcInfo.IsDir())
	}
	
	// Both files exist - need conflict resolution
	return r.resolveConflict(op, srcPath, destPath, srcInfo, destInfo, opts)
}

// ConflictStrategy defines how to resolve bidirectional conflicts
type ConflictStrategy string

const (
	NewestWins ConflictStrategy = "newest-wins"
	LargestWins ConflictStrategy = "largest-wins"
	SourceWins ConflictStrategy = "source-wins"
	DestWins ConflictStrategy = "dest-wins"
	Merge ConflictStrategy = "merge"
	Backup ConflictStrategy = "backup"
)

// resolveConflict handles conflict resolution for bidirectional sync
func (r *Runner) resolveConflict(op PlanOperation, srcPath, destPath string, srcInfo, destInfo os.FileInfo, opts *Options) error {
	r.logger.Infof("Resolving conflict for: %s", op.Path)
	
	// Check if files are identical
	if !srcInfo.IsDir() && !destInfo.IsDir() {
		if r.filesAreIdentical(srcPath, destPath) {
			r.logger.Infof("Files are identical, no sync needed: %s", op.Path)
			return nil
		}
	}
	
	// Default strategy: newest-wins
	strategy := r.getConflictStrategy(op, opts)
	
	switch strategy {
	case NewestWins:
		return r.resolveNewestWins(srcPath, destPath, srcInfo, destInfo)
	case LargestWins:
		return r.resolveLargestWins(srcPath, destPath, srcInfo, destInfo)
	case SourceWins:
		return r.syncSingleFile(srcPath, destPath, srcInfo.IsDir())
	case DestWins:
		return r.syncSingleFile(destPath, srcPath, destInfo.IsDir())
	case Backup:
		return r.resolveWithBackup(srcPath, destPath, srcInfo, destInfo)
	default:
		// Default to newest-wins
		return r.resolveNewestWins(srcPath, destPath, srcInfo, destInfo)
	}
}

// getConflictStrategy determines the conflict resolution strategy from flags or defaults
func (r *Runner) getConflictStrategy(op PlanOperation, opts *Options) ConflictStrategy {
	// Check if strategy is specified in the operation flags
	if strings.Contains(op.Flags, "auto:newest") {
		return NewestWins
	}
	if strings.Contains(op.Flags, "auto:largest") {
		return LargestWins
	}
	if strings.Contains(op.Flags, "auto:source") {
		return SourceWins
	}
	if strings.Contains(op.Flags, "auto:dest") {
		return DestWins
	}
	if strings.Contains(op.Flags, "auto:backup") {
		return Backup
	}
	
	// TODO: Check opts for global conflict strategy setting
	// For now, default to newest-wins
	return NewestWins
}

// resolveNewestWins syncs the newer file to the older location
func (r *Runner) resolveNewestWins(srcPath, destPath string, srcInfo, destInfo os.FileInfo) error {
	if srcInfo.ModTime().After(destInfo.ModTime()) {
		r.logger.Infof("Source is newer, syncing to destination: %s", srcPath)
		return r.syncSingleFile(srcPath, destPath, srcInfo.IsDir())
	} else if destInfo.ModTime().After(srcInfo.ModTime()) {
		r.logger.Infof("Destination is newer, syncing to source: %s", destPath)
		return r.syncSingleFile(destPath, srcPath, destInfo.IsDir())
	} else {
		// Same modification time - use size as tiebreaker
		if srcInfo.Size() > destInfo.Size() {
			r.logger.Infof("Same modification time, source is larger: %s", srcPath)
			return r.syncSingleFile(srcPath, destPath, srcInfo.IsDir())
		} else {
			r.logger.Infof("Same modification time, destination is larger or same size: %s", destPath)
			return r.syncSingleFile(destPath, srcPath, destInfo.IsDir())
		}
	}
}

// resolveLargestWins syncs the larger file to the smaller location
func (r *Runner) resolveLargestWins(srcPath, destPath string, srcInfo, destInfo os.FileInfo) error {
	if srcInfo.Size() > destInfo.Size() {
		r.logger.Infof("Source is larger, syncing to destination: %s", srcPath)
		return r.syncSingleFile(srcPath, destPath, srcInfo.IsDir())
	} else if destInfo.Size() > srcInfo.Size() {
		r.logger.Infof("Destination is larger, syncing to source: %s", destPath)
		return r.syncSingleFile(destPath, srcPath, destInfo.IsDir())
	} else {
		// Same size - use modification time as tiebreaker
		return r.resolveNewestWins(srcPath, destPath, srcInfo, destInfo)
	}
}

// resolveWithBackup creates backup copies before syncing
func (r *Runner) resolveWithBackup(srcPath, destPath string, srcInfo, destInfo os.FileInfo) error {
	timestamp := time.Now().Unix()
	
	// Create backup of destination file
	destBackup := fmt.Sprintf("%s.conflict-%d", destPath, timestamp)
	r.logger.Infof("Creating backup of destination: %s", destBackup)
	if err := r.copyFile(destPath, destBackup); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Sync using newest-wins strategy
	return r.resolveNewestWins(srcPath, destPath, srcInfo, destInfo)
}

// copyFile creates a copy of a file
func (r *Runner) copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// generatePlanContent creates the plan file content from sync analysis
func (r *Runner) generatePlanContent(opts *Options, report *SyncReport) (string, error) {
	var content strings.Builder
	
	// Generate header with metadata
	content.WriteString(fmt.Sprintf("# Sync Plan Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("# Generated from: sync-tools sync --source %s --dest %s --plan %s\n", 
		opts.Source, opts.Dest, opts.Plan))
	content.WriteString(fmt.Sprintf("# Source: %s\n", opts.Source))
	content.WriteString(fmt.Sprintf("# Destination: %s\n", opts.Dest))
	content.WriteString(fmt.Sprintf("# Mode: %s\n", opts.Mode))
	
	// Add change filter info if specified
	if len(opts.IncludeChanges) > 0 {
		content.WriteString(fmt.Sprintf("# Include changes: %s\n", strings.Join(opts.IncludeChanges, ", ")))
	}
	if len(opts.ExcludeChanges) > 0 {
		content.WriteString(fmt.Sprintf("# Exclude changes: %s\n", strings.Join(opts.ExcludeChanges, ", ")))
	}
	
	content.WriteString("#\n")
	content.WriteString("# Commands:\n")
	content.WriteString("#   s2d, sync-to-dest, <<    - Sync from source to destination (source >> dest)\n")
	content.WriteString("#   d2s, dest-to-source, >>  - Sync from destination to source (dest >> source)\n")
	content.WriteString("#   bid, bidirectional, <>   - Sync in both directions (bidirectional)\n")
	content.WriteString("#   skip                     - Skip this item (commented out)\n")
	content.WriteString("#\n")
	content.WriteString("# Visual aliases make direction intuitive:\n")
	content.WriteString("#   << = source flows to dest (like << redirection)\n")  
	content.WriteString("#   >> = dest flows to source (like >> redirection)\n")
	content.WriteString("#   <> = bidirectional flow (like <-> but shorter)\n")
	content.WriteString("#\n")
	content.WriteString("# Format: <command> <item-type> <path> [size] [modified] [flags]\n")
	content.WriteString("\n")
	
	// Filter changes based on include/exclude options
	filteredChanges := r.filterChanges(report.Changes, opts)
	
	// Generate operations for each change
	for _, change := range filteredChanges {
		operation := r.determineOperation(change, opts.Mode)
		content.WriteString(fmt.Sprintf("%s %s %-30s %8s  %s  %s\n",
			operation,
			r.getItemType(change),
			change.Path,
			r.formatSize(change.Size),
			change.ModTime.Format("2006-01-02T15:04:05"),
			r.getChangeFlags(change)))
	}
	
	// Generate summary
	content.WriteString("\n# Summary:\n")
	summary := r.generateSummary(filteredChanges, len(report.Changes))
	content.WriteString(summary)
	
	return content.String(), nil
}

// filterChanges applies include/exclude change type filtering
func (r *Runner) filterChanges(changes []SyncChange, opts *Options) []SyncChange {
	if len(opts.IncludeChanges) == 0 && len(opts.ExcludeChanges) == 0 {
		return changes // No filtering
	}
	
	var filtered []SyncChange
	for _, change := range changes {
		changeType := r.getChangeType(change)
		
		// If include list is specified, only include matching types
		if len(opts.IncludeChanges) > 0 {
			included := false
			for _, include := range opts.IncludeChanges {
				if changeType == include {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		// If exclude list is specified, exclude matching types
		if len(opts.ExcludeChanges) > 0 {
			excluded := false
			for _, exclude := range opts.ExcludeChanges {
				if changeType == exclude {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		filtered = append(filtered, change)
	}
	
	return filtered
}

// determineOperation determines the visual alias operation for a change
func (r *Runner) determineOperation(change SyncChange, mode string) string {
	// Determine visual alias based on action and sync mode
	switch change.Action {
	case "create":
		return "<<" // New file from source to dest
	case "update":
		return "<<" // Update from source to dest (source newer)
	case "delete":
		return ">>" // File only in dest (shown as dest-only file)
	case "conflict":
		return "<>" // Bidirectional conflict indicator
	default:
		return "<<" // Default to source to dest
	}
}

// getItemType determines if the item is a file or directory
func (r *Runner) getItemType(change SyncChange) string {
	if change.IsDir {
		return "dir "
	}
	return "file"
}

// getChangeType maps SyncChange to change type string
func (r *Runner) getChangeType(change SyncChange) string {
	// Map Action field to change types used in filtering
	switch change.Action {
	case "create":
		return "new-in-source" // New file in source
	case "update":
		return "updates" // File updated (source newer)
	case "delete":
		return "new-in-dest" // File only in dest (will be deleted)
	case "conflict":
		return "conflicts" // Both modified or dest newer
	default:
		return "unchanged"
	}
}

// getChangeFlags generates the flag description for a change
func (r *Runner) getChangeFlags(change SyncChange) string {
	// Generate descriptive flags matching BDD expectations
	switch change.Action {
	case "create":
		return "[new-in-source]"
	case "update":
		return "[update: newer-in-source]"
	case "delete":
		return "[new-in-dest]"
	case "conflict":
		return "[CONFLICT: both-modified]"
	default:
		return fmt.Sprintf("[%s]", change.Action)
	}
}

// generateSummary creates the summary statistics
func (r *Runner) generateSummary(filteredChanges []SyncChange, totalChanges int) string {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("# Files matching filter: %d\n", len(filteredChanges)))
	
	// Count by change type
	counts := make(map[string]int)
	for _, change := range filteredChanges {
		changeType := r.getChangeType(change)
		counts[changeType]++
	}
	
	summary.WriteString(fmt.Sprintf("# New in source: %d\n", counts["new-in-source"]))
	summary.WriteString(fmt.Sprintf("# New in dest: %d\n", counts["new-in-dest"]))
	summary.WriteString(fmt.Sprintf("# Updates: %d\n", counts["updates"]))
	summary.WriteString(fmt.Sprintf("# Conflicts: %d\n", counts["conflicts"]))
	
	if len(filteredChanges) != totalChanges {
		summary.WriteString(fmt.Sprintf("# Filtered out: %d\n", totalChanges-len(filteredChanges)))
	}
	
	return summary.String()
}

// openPlanInEditor opens the generated plan file in the user's editor
func (r *Runner) openPlanInEditor(opts *Options) error {
	// Determine which editor to use
	editor := opts.Editor // Check if custom editor was specified
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "vi", "nano", "emacs", "code"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found; please set EDITOR environment variable or use --editor flag")
	}
	
	r.logger.Infof("Opening plan file in editor: %s", editor)
	
	// Get absolute path of plan file
	absPlanPath, err := filepath.Abs(opts.Plan)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of plan file: %w", err)
	}
	
	// Open editor
	cmd := exec.Command(editor, absPlanPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	// After editing, validate the plan
	r.logger.Info("Validating edited plan file...")
	if err := r.validatePlanFile(absPlanPath); err != nil {
		r.logger.Warnf("Plan validation warning: %v", err)
	}
	
	// Ask if user wants to execute the plan immediately
	if !opts.DryRun && !opts.Yes {
		if r.confirmPlanExecution(absPlanPath) {
			// Execute the plan
			opts.ApplyPlan = absPlanPath
			return r.ExecutePlan(opts)
		}
	}
	
	return nil
}

// validatePlanFile performs basic validation on a plan file
func (r *Runner) validatePlanFile(planPath string) error {
	content, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("failed to read plan file: %w", err)
	}
	
	if len(content) == 0 {
		return fmt.Errorf("plan file is empty")
	}
	
	// Parse plan to check for basic validity
	planData, err := r.parsePlan(string(content))
	if err != nil {
		return fmt.Errorf("invalid plan syntax: %w", err)
	}
	
	if len(planData.Operations) == 0 {
		return fmt.Errorf("no operations found in plan")
	}
	
	// Check for invalid aliases
	validAliases := map[string]bool{
		"<<": true, ">>": true, "<>": true,
		"s2d": true, "sync-to-dest": true,
		"d2s": true, "dest-to-source": true,
		"bid": true, "bidirectional": true,
		"skip": true,
	}
	
	for _, op := range planData.Operations {
		if !validAliases[op.Alias] {
			return fmt.Errorf("invalid operation alias: %s", op.Alias)
		}
	}
	
	r.logger.Info("Plan file is valid")
	return nil
}

// confirmPlanExecution prompts the user to confirm plan execution
func (r *Runner) confirmPlanExecution(planPath string) bool {
	fmt.Printf("\nDo you want to execute the plan now? [y/N]: ")
	
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}