package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/DamianReeves/sync-tools/internal/config"
	"github.com/DamianReeves/sync-tools/internal/logging"
	"github.com/DamianReeves/sync-tools/internal/rsync"
	"github.com/DamianReeves/sync-tools/pkg/tui"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Perform a sync between SOURCE and DEST using rsync with layered filters",
	Long: `Perform a sync between SOURCE and DEST using rsync with layered filters.

You can specify defaults in a TOML config and override them on the command line.

Examples:
  sync-tools sync --source ./project --dest ./backup --dry-run
  sync-tools sync --config sync.toml --mode two-way
  sync-tools sync --source ./src --dest ./dst --only docs/ --report report.md`,
	RunE: runSync,
}

// Sync command flags
var (
	flagSource           string
	flagDest             string
	flagMode             string
	flagDryRun           bool
	flagUseSourceGitignore bool
	flagExcludeHiddenDirs bool
	flagOnlySyncignore    bool
	flagIgnoreSrc         []string
	flagIgnoreDest        []string
	flagOnly              []string
	flagLogLevel          string
	flagLogFile           string
	flagLogFormat         string
	flagDumpCommands      string
	flagReport            string
	flagListFiltered      string
	flagInteractive       bool
	flagPatch             string
	flagApplyPatch        bool
	flagYes               bool
	flagPreview           bool
	// Interactive sync plan flags
	flagPlan              string
	flagApplyPlan         string
	flagIncludeChanges    []string
	flagExcludeChanges    []string
	flagEditor            string
	// Conflict resolution flags
	flagConflictStrategy  string
	flagSkipConflicts     bool
	flagGenerateConflictPlan string
)

func init() {
	rootCmd.AddCommand(syncCmd)

	// Required flags
	syncCmd.Flags().StringVar(&flagSource, "source", "", "Source directory path")
	syncCmd.Flags().StringVar(&flagDest, "dest", "", "Destination directory path")

	// Mode flags
	syncCmd.Flags().StringVar(&flagMode, "mode", "one-way", "Sync mode: one-way or two-way")
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be synced without making changes")
	syncCmd.Flags().BoolVar(&flagInteractive, "interactive", false, "Use interactive Bubble Tea interface")

	// Filter flags
	syncCmd.Flags().BoolVar(&flagUseSourceGitignore, "use-source-gitignore", false, "Include .gitignore patterns from source")
	syncCmd.Flags().BoolVar(&flagExcludeHiddenDirs, "exclude-hidden-dirs", false, "Exclude all hidden directories (starting with .)")
	syncCmd.Flags().BoolVar(&flagOnlySyncignore, "only-syncignore", false, "Only use .syncignore files, ignore other filters")
	syncCmd.Flags().StringSliceVar(&flagIgnoreSrc, "ignore-src", nil, "Source-side ignore patterns")
	syncCmd.Flags().StringSliceVar(&flagIgnoreDest, "ignore-dest", nil, "Destination-side ignore patterns")
	syncCmd.Flags().StringSliceVar(&flagOnly, "only", nil, "Whitelist mode - only sync these paths")

	// Output flags
	syncCmd.Flags().StringVar(&flagLogLevel, "log-level", "", "Log level: DEBUG, INFO, WARNING, ERROR, CRITICAL")
	syncCmd.Flags().StringVar(&flagLogFile, "log-file", "", "Path to write logs")
	syncCmd.Flags().StringVar(&flagLogFormat, "log-format", "text", "Log format: text or json")
	syncCmd.Flags().StringVar(&flagDumpCommands, "dump-commands", "", "Write rsync command and filters to JSON file")
	syncCmd.Flags().StringVar(&flagReport, "report", "", "Write a sync report to this path (format detected from extension: .md/.markdown for markdown, .patch for patch)")
	syncCmd.Flags().StringVar(&flagListFiltered, "list-filtered", "", "List items that would be filtered: src, dst, or both")
	syncCmd.Flags().StringVar(&flagPatch, "patch", "", "Generate git patch file instead of syncing")
	syncCmd.Flags().BoolVar(&flagApplyPatch, "apply-patch", false, "Apply the generated patch after creation (with confirmation)")
	syncCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Automatically confirm patch application (skip confirmation prompt)")
	syncCmd.Flags().BoolVar(&flagPreview, "preview", false, "Show a colored diff preview of changes (with paging)")

	// Interactive sync plan flags
	syncCmd.Flags().StringVar(&flagPlan, "plan", "", "Generate a sync plan file instead of executing")
	syncCmd.Flags().StringVar(&flagApplyPlan, "apply-plan", "", "Execute operations from a sync plan file")
	syncCmd.Flags().StringSliceVar(&flagIncludeChanges, "include-changes", []string{}, "Include only these change types: new-in-source, new-in-dest, updates, conflicts, deletions, unchanged")
	syncCmd.Flags().StringSliceVar(&flagExcludeChanges, "exclude-changes", []string{}, "Exclude these change types: new-in-source, new-in-dest, updates, conflicts, deletions, unchanged")
	syncCmd.Flags().StringVar(&flagEditor, "editor", "", "Editor to use for interactive plan editing (overrides EDITOR env var)")
	
	// Conflict resolution flags
	syncCmd.Flags().StringVar(&flagConflictStrategy, "conflict-strategy", "", "Default conflict resolution strategy: newest-wins, source-wins, dest-wins, backup")
	syncCmd.Flags().BoolVar(&flagSkipConflicts, "skip-conflicts", false, "Skip conflicting files during plan execution")
	syncCmd.Flags().StringVar(&flagGenerateConflictPlan, "generate-conflict-plan", "", "Generate a separate plan file containing only conflicts")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	verbosity, _ := cmd.Flags().GetCount("verbose")
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Merge CLI flags with config
	opts := mergeOptionsWithConfig(cfg)

	// Setup logging
	logger, err := logging.Setup(opts.LogLevel, opts.LogFile, opts.LogFormat, verbosity)
	if err != nil {
		return fmt.Errorf("error setting up logging: %w", err)
	}

	logger.Debugf("CLI options after merge: source=%s dest=%s mode=%s dry-run=%v", 
		opts.Source, opts.Dest, opts.Mode, opts.DryRun)

	// Validate required options (unless we're applying a plan file)
	if opts.ApplyPlan == "" && (opts.Source == "" || opts.Dest == "") {
		return fmt.Errorf("source and dest must be provided either via CLI or config file")
	}

	// Resolve paths (if provided)
	if opts.Source != "" {
		sourcePath, err := filepath.Abs(opts.Source)
		if err != nil {
			return fmt.Errorf("error resolving source path: %w", err)
		}
		opts.Source = sourcePath
	}

	if opts.Dest != "" {
		destPath, err := filepath.Abs(opts.Dest)
		if err != nil {
			return fmt.Errorf("error resolving dest path: %w", err)
		}
		opts.Dest = destPath
	}

	// Check if source exists (unless we're applying a plan file)
	if opts.ApplyPlan == "" && opts.Source != "" {
		if _, err := os.Stat(opts.Source); os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", opts.Source)
		}
	}

	// Handle plan operations
	if opts.Plan != "" {
		return runPlanGeneration(opts, logger)
	}
	
	if opts.ApplyPlan != "" {
		return runPlanExecution(opts, logger)
	}

	// Check if using interactive mode
	if opts.Interactive {
		return runInteractiveSync(opts, logger)
	}

	// Run traditional CLI sync
	return runTraditionalSync(opts, logger)
}

func mergeOptionsWithConfig(cfg *config.Config) *rsync.Options {
	opts := &rsync.Options{
		Source:              flagSource,
		Dest:                flagDest,
		Mode:                flagMode,
		DryRun:              flagDryRun,
		UseSourceGitignore:  flagUseSourceGitignore,
		ExcludeHiddenDirs:   flagExcludeHiddenDirs,
		OnlySyncignore:      flagOnlySyncignore,
		IgnoreSrc:           flagIgnoreSrc,
		IgnoreDest:          flagIgnoreDest,
		Only:                flagOnly,
		LogLevel:            flagLogLevel,
		LogFile:             flagLogFile,
		LogFormat:           flagLogFormat,
		DumpCommands:        flagDumpCommands,
		Report:              flagReport,
		ListFiltered:        flagListFiltered,
		Interactive:         flagInteractive,
		Patch:               flagPatch,
		ApplyPatch:          flagApplyPatch,
		Yes:                 flagYes,
		Preview:             flagPreview,
		// Interactive sync plan fields
		Plan:                flagPlan,
		ApplyPlan:           flagApplyPlan,
		IncludeChanges:      flagIncludeChanges,
		ExcludeChanges:      flagExcludeChanges,
		Editor:              flagEditor,
		// Conflict resolution fields
		ConflictStrategy:    flagConflictStrategy,
		SkipConflicts:       flagSkipConflicts,
		GenerateConflictPlan: flagGenerateConflictPlan,
	}

	// Merge with config values (config provides defaults)
	if cfg != nil {
		if opts.Source == "" && cfg.Source != "" {
			opts.Source = cfg.Source
		}
		if opts.Dest == "" && cfg.Dest != "" {
			opts.Dest = cfg.Dest
		}
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
		if len(opts.IgnoreDest) == 0 && len(cfg.IgnoreDest) > 0 {
			opts.IgnoreDest = cfg.IgnoreDest
		}
		if len(opts.Only) == 0 && len(cfg.Only) > 0 {
			opts.Only = cfg.Only
		}
	}

	return opts
}

func runInteractiveSync(opts *rsync.Options, logger logging.Logger) error {
	// Create the Bubble Tea model
	model := tui.NewModel(opts, logger)

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running interactive sync: %w", err)
	}

	return nil
}

func runTraditionalSync(opts *rsync.Options, logger logging.Logger) error {
	// Create rsync runner
	runner := rsync.NewRunner(logger)

	// Execute sync
	return runner.Sync(opts)
}

func runPlanGeneration(opts *rsync.Options, logger logging.Logger) error {
	// Create rsync runner
	runner := rsync.NewRunner(logger)

	// Generate plan file
	return runner.GeneratePlan(opts)
}

func runPlanExecution(opts *rsync.Options, logger logging.Logger) error {
	// Create rsync runner
	runner := rsync.NewRunner(logger)

	// Execute plan file
	return runner.ExecutePlan(opts)
}