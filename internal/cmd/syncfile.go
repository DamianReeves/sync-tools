package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DamianReeves/sync-tools/internal/logging"
	"github.com/DamianReeves/sync-tools/internal/rsync"
	"github.com/DamianReeves/sync-tools/pkg/syncfile"
	"github.com/spf13/cobra"
)

// syncfileCmd represents the syncfile command
var syncfileCmd = &cobra.Command{
	Use:   "syncfile [SYNCFILE]",
	Short: "Execute a SyncFile (Dockerfile-like syntax for sync operations)",
	Long: `Execute a SyncFile with Dockerfile-like syntax for declarative sync operations.

A SyncFile uses a simple, declarative syntax inspired by Dockerfiles to define
multiple sync operations, filters, and configurations in a single file.

Example SyncFile:
  # Multi-project sync configuration
  VAR PROJECT_ROOT=/home/user/projects
  VAR BACKUP_ROOT=/backup

  # Sync documentation
  SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs
  MODE one-way
  EXCLUDE *.tmp
  INCLUDE !important.tmp

  # Sync source code with two-way sync
  SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
  MODE two-way
  GITIGNORE true
  ONLY *.go
  ONLY *.py

Available Instructions:
  SYNC source dest [options] - Define a sync operation
  MODE one-way|two-way       - Set sync mode
  EXCLUDE pattern            - Exclude files/folders matching pattern
  INCLUDE pattern            - Include files (unignore pattern)
  ONLY pattern               - Whitelist mode - only sync matching files
  DRYRUN true|false         - Enable/disable dry run mode
  GITIGNORE true|false      - Use source .gitignore patterns
  HIDDENDIRS exclude|include - Exclude or include hidden directories
  VAR name=value            - Define a variable
  ENV name=value            - Define an environment variable
  RUN command               - Execute command (pre/post sync hooks)
  # comment                 - Comments

Variables can be referenced using ${name} or $name syntax.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSyncfile,
}

var (
	flagSyncfileDryRun bool
	flagSyncfileList   bool
)

func init() {
	rootCmd.AddCommand(syncfileCmd)

	syncfileCmd.Flags().BoolVar(&flagSyncfileDryRun, "dry-run", false, "Override all SYNC operations to use dry-run mode")
	syncfileCmd.Flags().BoolVar(&flagSyncfileList, "list", false, "List sync operations without executing")
}

func runSyncfile(cmd *cobra.Command, args []string) error {
	// Determine SyncFile path
	syncfilePath := "SyncFile"
	if len(args) > 0 {
		syncfilePath = args[0]
	}

	// Look for common SyncFile names if not found
	if syncfilePath == "SyncFile" {
		candidates := []string{"SyncFile", "Syncfile", "syncfile", ".syncfile"}
		found := false
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				syncfilePath = candidate
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no SyncFile found. Looked for: %v", candidates)
		}
	}

	// Parse SyncFile
	sf, err := syncfile.ParseSyncFile(syncfilePath)
	if err != nil {
		return fmt.Errorf("error parsing SyncFile: %w", err)
	}

	// Convert to rsync options
	optsList, err := sf.ToRsyncOptions()
	if err != nil {
		return fmt.Errorf("error converting SyncFile to rsync options: %w", err)
	}

	// Setup logging
	verbosity, _ := cmd.Flags().GetCount("verbose")
	logger, err := logging.Setup("INFO", "", "text", verbosity)
	if err != nil {
		return fmt.Errorf("error setting up logging: %w", err)
	}

	logger.Infof("Executing SyncFile: %s", syncfilePath)
	logger.Infof("Found %d sync operations", len(optsList))

	// Override dry-run if flag is set
	if flagSyncfileDryRun {
		for _, opts := range optsList {
			opts.DryRun = true
		}
		logger.Info("Dry-run mode enabled for all operations")
	}

	// List operations if requested
	if flagSyncfileList {
		for i, opts := range optsList {
			logger.Infof("Operation %d:", i+1)
			logger.Infof("  Source: %s", opts.Source)
			logger.Infof("  Dest:   %s", opts.Dest)
			logger.Infof("  Mode:   %s", opts.Mode)
			logger.Infof("  DryRun: %v", opts.DryRun)
			if len(opts.IgnoreSrc) > 0 {
				logger.Infof("  Filters: %v", opts.IgnoreSrc)
			}
			if len(opts.Only) > 0 {
				logger.Infof("  Whitelist: %v", opts.Only)
			}
		}
		return nil
	}

	// Execute sync operations
	runner := rsync.NewRunner(logger)
	
	for i, opts := range optsList {
		logger.Infof("Executing sync operation %d/%d", i+1, len(optsList))
		logger.Infof("  %s -> %s", opts.Source, opts.Dest)

		// Resolve paths relative to SyncFile location
		syncfileDir := filepath.Dir(syncfilePath)
		if !filepath.IsAbs(opts.Source) {
			opts.Source = filepath.Join(syncfileDir, opts.Source)
		}
		if !filepath.IsAbs(opts.Dest) {
			opts.Dest = filepath.Join(syncfileDir, opts.Dest)
		}

		if err := runner.Sync(opts); err != nil {
			return fmt.Errorf("sync operation %d failed: %w", i+1, err)
		}
	}

	logger.Info("All sync operations completed successfully")
	return nil
}