package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DamianReeves/sync-tools/internal/config"
	"github.com/DamianReeves/sync-tools/internal/logging"
	"github.com/DamianReeves/sync-tools/internal/rsync"
	"github.com/spf13/cobra"
)

// syncFromCmd represents the "sync from" subcommand
var syncFromCmd = &cobra.Command{
	Use:   "from SOURCE_DIR",
	Short: "Sync from source directory to the current directory",
	Long: `Sync from a source directory to the current working directory.
This is a convenience command that automatically uses the current directory as the destination.

Examples:
  # Sync from a source to current directory
  sync-tools sync from ~/projects/myapp
  
  # With dry-run to preview changes
  sync-tools sync from ~/backup --dry-run
  
  # With markdown report
  sync-tools sync from ~/data --report sync_report.md
  
  # With filters
  sync-tools sync from ~/source --only "*.go" --exclude-hidden-dirs
  
  # Preview changes with colored diff
  sync-tools sync from ~/documents --preview`,
	Args: cobra.ExactArgs(1),
	RunE: runSyncFrom,
}

// Sync from command flags (subset of sync flags relevant for this use case)
var (
	fromFlagMode               string
	fromFlagDryRun             bool
	fromFlagUseSourceGitignore bool
	fromFlagExcludeHiddenDirs  bool
	fromFlagOnlySyncignore     bool
	fromFlagIgnoreSrc          []string
	fromFlagOnly               []string
	fromFlagLogLevel           string
	fromFlagLogFile            string
	fromFlagLogFormat          string
	fromFlagDumpCommands       string
	fromFlagReport             string
	fromFlagListFiltered       string
	fromFlagPatch              string
	fromFlagApplyPatch         bool
	fromFlagYes                bool
	fromFlagPreview            bool
)

func init() {
	// Add "from" as a subcommand of "sync"
	syncCmd.AddCommand(syncFromCmd)

	// Mode flags
	syncFromCmd.Flags().StringVar(&fromFlagMode, "mode", "one-way", "Sync mode: one-way or two-way")
	syncFromCmd.Flags().BoolVar(&fromFlagDryRun, "dry-run", false, "Show what would be synced without making changes")

	// Filter flags
	syncFromCmd.Flags().BoolVar(&fromFlagUseSourceGitignore, "use-source-gitignore", false, "Include .gitignore patterns from source")
	syncFromCmd.Flags().BoolVar(&fromFlagExcludeHiddenDirs, "exclude-hidden-dirs", false, "Exclude all hidden directories (starting with .)")
	syncFromCmd.Flags().BoolVar(&fromFlagOnlySyncignore, "only-syncignore", false, "Only use .syncignore files, ignore other filters")
	syncFromCmd.Flags().StringSliceVar(&fromFlagIgnoreSrc, "ignore-src", nil, "Source-side ignore patterns")
	syncFromCmd.Flags().StringSliceVar(&fromFlagOnly, "only", nil, "Whitelist mode - only sync these paths")

	// Output flags
	syncFromCmd.Flags().StringVar(&fromFlagLogLevel, "log-level", "", "Log level: DEBUG, INFO, WARNING, ERROR, CRITICAL")
	syncFromCmd.Flags().StringVar(&fromFlagLogFile, "log-file", "", "Path to write logs")
	syncFromCmd.Flags().StringVar(&fromFlagLogFormat, "log-format", "text", "Log format: text or json")
	syncFromCmd.Flags().StringVar(&fromFlagDumpCommands, "dump-commands", "", "Write rsync command and filters to JSON file")
	syncFromCmd.Flags().StringVar(&fromFlagReport, "report", "", "Write a sync report to this path (format detected from extension: .md/.markdown for markdown, .patch for patch)")
	syncFromCmd.Flags().StringVar(&fromFlagListFiltered, "list-filtered", "", "List items that would be filtered: src, dst, or both")
	syncFromCmd.Flags().StringVar(&fromFlagPatch, "patch", "", "Generate git patch file instead of syncing")
	syncFromCmd.Flags().BoolVar(&fromFlagApplyPatch, "apply-patch", false, "Apply the generated patch after creation (with confirmation)")
	syncFromCmd.Flags().BoolVarP(&fromFlagYes, "yes", "y", false, "Automatically confirm patch application (skip confirmation prompt)")
	syncFromCmd.Flags().BoolVar(&fromFlagPreview, "preview", false, "Show a colored diff preview of changes (with paging)")
}

func runSyncFrom(cmd *cobra.Command, args []string) error {
	// Get current working directory as destination
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	verbosity, _ := cmd.Flags().GetCount("verbose")
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Get the source directory (we know there's exactly one from Args validation)
	sourceArg := args[0]
	
	// Resolve source path
	sourcePath, err := filepath.Abs(sourceArg)
	if err != nil {
		return fmt.Errorf("error resolving source path '%s': %w", sourceArg, err)
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourcePath)
	}

	// Prevent syncing to itself
	if sourcePath == currentDir {
		return fmt.Errorf("cannot sync directory to itself: %s", sourcePath)
	}

	// Build options
	opts := &rsync.Options{
		Source:              sourcePath,
		Dest:                currentDir,
		Mode:                fromFlagMode,
		DryRun:              fromFlagDryRun,
		UseSourceGitignore:  fromFlagUseSourceGitignore,
		ExcludeHiddenDirs:   fromFlagExcludeHiddenDirs,
		OnlySyncignore:      fromFlagOnlySyncignore,
		IgnoreSrc:           fromFlagIgnoreSrc,
		Only:                fromFlagOnly,
		LogLevel:            fromFlagLogLevel,
		LogFile:             fromFlagLogFile,
		LogFormat:           fromFlagLogFormat,
		DumpCommands:        fromFlagDumpCommands,
		Report:              fromFlagReport,
		ListFiltered:        fromFlagListFiltered,
		Patch:               fromFlagPatch,
		ApplyPatch:          fromFlagApplyPatch,
		Yes:                 fromFlagYes,
		Preview:             fromFlagPreview,
	}

	// Merge with config values (config provides defaults)
	if cfg != nil {
		if opts.Mode == "one-way" && cfg.Mode != "" {
			opts.Mode = cfg.Mode
		}
		if !opts.DryRun && cfg.DryRun {
			opts.DryRun = cfg.DryRun
		}
		if !opts.UseSourceGitignore && cfg.UseSourceGitignore {
			opts.UseSourceGitignore = cfg.UseSourceGitignore
		}
		if !opts.ExcludeHiddenDirs && cfg.ExcludeHiddenDirs {
			opts.ExcludeHiddenDirs = cfg.ExcludeHiddenDirs
		}
		if !opts.OnlySyncignore && cfg.OnlySyncignore {
			opts.OnlySyncignore = cfg.OnlySyncignore
		}
		if len(opts.IgnoreSrc) == 0 && len(cfg.IgnoreSrc) > 0 {
			opts.IgnoreSrc = cfg.IgnoreSrc
		}
		if len(opts.Only) == 0 && len(cfg.Only) > 0 {
			opts.Only = cfg.Only
		}
	}

	// Setup logging
	logger, err := logging.Setup(opts.LogLevel, opts.LogFile, opts.LogFormat, verbosity)
	if err != nil {
		return fmt.Errorf("error setting up logging: %w", err)
	}

	// Log the operation
	logger.Infof("Syncing from: %s -> %s (current directory)", 
		sourcePath, currentDir)

	// Create rsync runner and execute sync
	runner := rsync.NewRunner(logger)
	if err := runner.Sync(opts); err != nil {
		return fmt.Errorf("error syncing from %s: %w", sourcePath, err)
	}

	return nil
}