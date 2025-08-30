package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

  # Preview changes before syncing
  SYNC ${PROJECT_ROOT}/docs ${BACKUP_ROOT}/docs
  MODE one-way
  PREVIEW true
  EXCLUDE *.tmp
  INCLUDE !important.tmp

  # Generate and apply patch for source code
  SYNC ${PROJECT_ROOT}/src ${BACKUP_ROOT}/src
  MODE two-way
  PATCH src-changes.patch
  APPLYPATCH true
  AUTOCONFIRM false  # Prompt for confirmation
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
  PATCH filename            - Generate git patch file instead of syncing
  APPLYPATCH true|false     - Apply generated patch after creation
  PREVIEW true|false        - Show colored diff preview before sync
  AUTOCONFIRM true|false    - Auto-confirm patch application (like -y)
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
	flagSyncfilePlan   string
)

func init() {
	rootCmd.AddCommand(syncfileCmd)

	syncfileCmd.Flags().BoolVar(&flagSyncfileDryRun, "dry-run", false, "Override all SYNC operations to use dry-run mode")
	syncfileCmd.Flags().BoolVar(&flagSyncfileList, "list", false, "List sync operations without executing")
	syncfileCmd.Flags().StringVar(&flagSyncfilePlan, "plan", "", "Generate a sync plan file instead of executing")
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

	// Change to SyncFile directory for relative path resolution
	syncfileDir := filepath.Dir(syncfilePath)
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	
	err = os.Chdir(syncfileDir)
	if err != nil {
		return fmt.Errorf("error changing to SyncFile directory: %w", err)
	}
	
	// Ensure we return to original directory
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Convert to rsync options (now relative to SyncFile directory)
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
			if opts.Patch != "" {
				logger.Infof("  Patch: %s", opts.Patch)
				logger.Infof("  ApplyPatch: %v", opts.ApplyPatch)
				logger.Infof("  AutoConfirm: %v", opts.Yes)
			}
			if opts.Preview {
				logger.Infof("  Preview: %v", opts.Preview)
			}
			if len(opts.IgnoreSrc) > 0 {
				logger.Infof("  Filters: %v", opts.IgnoreSrc)
			}
			if len(opts.Only) > 0 {
				logger.Infof("  Whitelist: %v", opts.Only)
			}
		}
		return nil
	}

	// Handle plan generation if requested
	if flagSyncfilePlan != "" {
		logger.Infof("Generating sync plan: %s", flagSyncfilePlan)
		return generateSyncfilePlan(optsList, flagSyncfilePlan, syncfilePath, logger)
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

// generateSyncfilePlan generates a plan file from SyncFile operations
func generateSyncfilePlan(optsList []*rsync.Options, planFile string, syncfilePath string, logger logging.Logger) error {
	runner := rsync.NewRunner(logger)
	
	// For SyncFile plans, we need to generate combined plan content from all operations
	var planContent strings.Builder
	
	// Generate header
	planContent.WriteString(fmt.Sprintf("# Sync Plan Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	planContent.WriteString(fmt.Sprintf("# Generated from: sync-tools syncfile %s --plan %s\n", syncfilePath, planFile))
	planContent.WriteString(fmt.Sprintf("# SyncFile: %s\n", syncfilePath))
	
	// Process each sync operation
	for i, opts := range optsList {
		logger.Infof("Analyzing sync operation %d/%d", i+1, len(optsList))
		
		// Set plan file for this operation
		tempPlan := fmt.Sprintf("%s.temp.%d", planFile, i)
		opts.Plan = tempPlan
		
		// Generate plan for this operation
		if err := runner.GeneratePlan(opts); err != nil {
			logger.Warnf("Failed to generate plan for operation %d: %v", i+1, err)
			continue
		}
		
		// Read the temporary plan file
		tempContent, err := os.ReadFile(tempPlan)
		if err != nil {
			logger.Warnf("Failed to read temporary plan for operation %d: %v", i+1, err)
			continue
		}
		
		// Add operation header
		if i > 0 {
			planContent.WriteString("\n")
		}
		planContent.WriteString(fmt.Sprintf("# Operation %d: %s -> %s\n", i+1, opts.Source, opts.Dest))
		planContent.WriteString(fmt.Sprintf("# Source: %s\n", opts.Source))
		planContent.WriteString(fmt.Sprintf("# Destination: %s\n", opts.Dest))
		planContent.WriteString(fmt.Sprintf("# Mode: %s\n", opts.Mode))
		
		// Add filters info if any
		if len(opts.IgnoreSrc) > 0 || len(opts.Only) > 0 {
			filters := []string{}
			if len(opts.IgnoreSrc) > 0 {
				filters = append(filters, fmt.Sprintf("EXCLUDE %s", strings.Join(opts.IgnoreSrc, ", ")))
			}
			if opts.UseSourceGitignore {
				filters = append(filters, "GITIGNORE true")
			}
			if len(opts.Only) > 0 {
				filters = append(filters, fmt.Sprintf("ONLY %s", strings.Join(opts.Only, ", ")))
			}
			planContent.WriteString(fmt.Sprintf("# Filters: %s\n", strings.Join(filters, ", ")))
		}
		planContent.WriteString("#\n")
		
		// Extract operation lines from temp plan (skip headers and comments)
		tempLines := strings.Split(string(tempContent), "\n")
		for _, line := range tempLines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				planContent.WriteString(line + "\n")
			}
		}
		
		// Clean up temporary file
		os.Remove(tempPlan)
	}
	
	// Write the combined plan file
	err := os.WriteFile(planFile, []byte(planContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}
	
	logger.Infof("SyncFile plan generated successfully: %s", planFile)
	return nil
}