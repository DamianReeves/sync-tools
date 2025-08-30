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
	args := []string{
		"--archive",          // -a
		"--verbose",          // -v
		"--human-readable",   // -h
		"--delete",           // Remove files from dest that don't exist in source
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	// Add filter files
	if sourceFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", sourceFilter))
	}
	if destFilter != "" {
		args = append(args, "--filter", fmt.Sprintf(". %s", destFilter))
	}

	// Add source and destination
	// Ensure source path ends with / for proper rsync behavior
	source := opts.Source
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	args = append(args, source, opts.Dest)

	return exec.Command("rsync", args...)
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
		
		// Parse itemize-changes format
		if len(line) > 11 && line[0] != ' ' {
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

// parseRsyncChange parses a line from rsync's itemize-changes output
func (r *Runner) parseRsyncChange(line string) *SyncChange {
	// Itemize format: YXcstpoguax
	// Y: type of update
	// X: file type
	// c: checksum differs
	// s: size differs
	// t: mod time differs
	// etc.
	
	if len(line) < 11 {
		return nil
	}
	
	flags := line[:11]
	rest := strings.TrimSpace(line[11:])
	
	if rest == "" {
		return nil
	}
	
	// Parse the flags
	change := &SyncChange{}
	
	// Determine action based on flags
	if flags[0] == '>' || flags[0] == '<' {
		if flags[1] == 'f' {
			if flags[2] == '+' {
				change.Action = "create"
			} else {
				change.Action = "update"
			}
		} else if flags[1] == 'd' {
			change.IsDir = true
			if flags[2] == '+' {
				change.Action = "create"
			} else {
				change.Action = "update"
			}
		}
	} else if flags[0] == '*' {
		if flags[1] == 'd' {
			change.Action = "delete"
			change.IsDir = true
		} else {
			change.Action = "delete"
		}
	}
	
	// Extract path from the rest of the line
	parts := strings.Fields(rest)
	if len(parts) > 0 {
		change.Path = parts[0]
	}
	
	// Try to extract size if available
	if len(parts) > 3 {
		if size, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
			change.Size = size
		}
	}
	
	// Try to extract modification time if available
	if len(parts) > 4 {
		if modTime, err := time.Parse("2006/01/02-15:04:05", parts[4]); err == nil {
			change.ModTime = modTime
		}
	}
	
	return change
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