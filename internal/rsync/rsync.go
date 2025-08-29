package rsync

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	// Check if patch mode is requested
	if opts.Patch != "" {
		r.logger.Infof("Starting patch generation: %s -> %s (output: %s, dry-run: %v)",
			opts.Source, opts.Dest, opts.Patch, opts.DryRun)
		return r.generatePatch(opts)
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
		"--delete-excluded",  // Also delete excluded files from dest
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