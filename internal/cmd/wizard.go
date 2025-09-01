package cmd

import (
	"fmt"

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
)

func init() {
	rootCmd.AddCommand(wizardCmd)

	// Optional pre-fill flags
	wizardCmd.Flags().StringVar(&flagWizardSource, "source", "", "Pre-fill source directory")
	wizardCmd.Flags().StringVar(&flagWizardMode, "mode", "", "Pre-fill sync mode (one-way, two-way)")
}

func runWizard(cmd *cobra.Command, args []string) error {
	// Create wizard configuration
	config := &wizard.Config{
		PrefilledSource: flagWizardSource,
		PrefilledMode:   flagWizardMode,
	}

	// Start the interactive wizard
	w := wizard.New(config)
	if err := w.Run(); err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}

	return nil
}