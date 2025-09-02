package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/DamianReeves/sync-tools/internal/wizard"
	"github.com/spf13/cobra"
)

// wizardCmd represents the wizard command
var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive wizard for sync configuration",
	Long: `Interactive wizard for sync configuration.

The wizard provides a guided interface to configure complex sync operations
with directory browsing, filter management, and real-time sync monitoring.

Examples:
  sync-tools wizard
  sync-tools wizard --source ./project --mode two-way`,
	RunE: runWizard,
}

// Wizard command flags
var (
	flagWizardSource string
	flagWizardMode   string
	flagWizardTest   bool
)

func init() {
	// Hide command from help unless explicitly enabled via env var
	wizardCmd.Hidden = !isWizardEnabled()
	rootCmd.AddCommand(wizardCmd)

	// Optional pre-fill flags
	wizardCmd.Flags().StringVar(&flagWizardSource, "source", "", "Pre-fill source directory")
	wizardCmd.Flags().StringVar(&flagWizardMode, "mode", "", "Pre-fill sync mode (one-way, two-way)")
	wizardCmd.Flags().BoolVar(&flagWizardTest, "test", false, "Run in test mode (non-interactive)")
}

func runWizard(cmd *cobra.Command, args []string) error {
	// Gate interactive execution behind feature flag. Allow --test regardless.
	if !flagWizardTest && !isWizardEnabled() {
		return fmt.Errorf("wizard is disabled by default. Set SYNC_TOOLS_ENABLE_WIZARD=1 to enable.")
	}
	// Create wizard configuration
	config := &wizard.Config{
		PrefilledSource: flagWizardSource,
		PrefilledMode:   flagWizardMode,
		TestMode:        flagWizardTest,
	}

	// For test mode, add basic test options
	if flagWizardTest {
		testSourceDir := flagWizardSource
		// Use the provided source directory as-is for test mode
		if testSourceDir == "" {
			testSourceDir = "test_source"
		}

		config.TestOptions = &wizard.TestModeOptions{
			SourceDir:      testSourceDir,
			DestinationDir: "test_dest", // default for test mode
			Mode:           flagWizardMode,
			DryRun:         true,
		}

		// Set default mode if not provided
		if config.TestOptions.Mode == "" {
			config.TestOptions.Mode = "one-way"
		}
	}

	// Start the wizard (interactive or test mode based on config)
	w := wizard.New(config)
	if err := w.Run(); err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}

	return nil
}

// isWizardEnabled returns true if the wizard feature flag is enabled.
func isWizardEnabled() bool {
	v := strings.TrimSpace(os.Getenv("SYNC_TOOLS_ENABLE_WIZARD"))
	v = strings.ToLower(v)
	switch v {
	case "1", "true", "yes", "on", "enabled":
		return true
	default:
		return false
	}
}
