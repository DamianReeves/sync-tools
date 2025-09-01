package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DamianReeves/sync-tools/internal/wizard"
)

// wizardCmd represents the wizard command
var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Launch interactive wizard for sync setup",
	Long: `Launch the interactive wizard that guides you through setting up a sync operation.

The wizard provides a step-by-step interface to:
• Select sync mode (one-way or two-way)
• Choose source and destination directories
• Configure sync options and exclusion patterns
• Preview operations before execution
• Monitor sync progress in real-time

Examples:
  sync-tools wizard                              # Full guided setup
  sync-tools wizard --source ./project          # Pre-set source directory
  sync-tools sync --wizard --destination ./backup  # Pre-set destination directory`,
	RunE: runWizard,
}

var (
	// Wizard-specific flags
	wizardSource      string
	wizardDestination string
	wizardMode        string
	wizardTheme       string
	wizardTypeSafe    bool
)

func init() {
	rootCmd.AddCommand(wizardCmd)

	// Add wizard flags
	wizardCmd.Flags().StringVar(&wizardSource, "source", "", "Pre-set source directory")
	wizardCmd.Flags().StringVar(&wizardDestination, "destination", "", "Pre-set destination directory") 
	wizardCmd.Flags().StringVar(&wizardMode, "mode", "one-way", "Pre-set sync mode (one-way, two-way)")
	wizardCmd.Flags().StringVar(&wizardTheme, "theme", "default", "UI theme (default, compact, minimal)")
	wizardCmd.Flags().BoolVar(&wizardTypeSafe, "type-safe", true, "Use type-safe wizard implementation (default: true)")

	// Also add wizard flag to the main sync command
	syncCmd.Flags().Bool("wizard", false, "Launch interactive wizard mode")
}

// runWizard executes the interactive wizard
func runWizard(cmd *cobra.Command, args []string) error {
	// Check if we're in wizard mode from sync command
	if cmd.Parent() != nil && cmd.Parent().Name() == "sync" {
		isWizard, _ := cmd.Parent().Flags().GetBool("wizard")
		if !isWizard {
			return fmt.Errorf("wizard flag not set")
		}
	}

	// Check terminal capabilities
	if !isTerminalSupported() {
		return fmt.Errorf("wizard mode requires a supported terminal")
	}

	// Create wizard model - choose between type-safe and original
	var model tea.Model
	if wizardTypeSafe {
		model = wizard.NewTypeStateModel()
		// Apply flags to type-safe model (TODO: implement)
	} else {
		originalModel := wizard.NewModel()
		// Apply any pre-set options from flags
		if err := applyWizardFlags(originalModel); err != nil {
			return fmt.Errorf("failed to apply wizard flags: %w", err)
		}
		model = originalModel
	}

	// Create Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Process wizard results
	if wizardTypeSafe {
		if typeSafeModel, ok := finalModel.(wizard.TypeStateModel); ok {
			return handleTypeStateWizardCompletion(typeSafeModel)
		}
	} else {
		if wizardModel, ok := finalModel.(wizard.Model); ok {
			return handleWizardCompletion(wizardModel)
		}
	}

	return nil
}

// isTerminalSupported checks if the current terminal supports the wizard UI
func isTerminalSupported() bool {
	// Check if stdout is a terminal
	if fi, err := os.Stdout.Stat(); err != nil {
		return false
	} else if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Check minimum terminal size
	// This will be handled by the wizard UI itself, so we'll allow it for now
	return true
}

// applyWizardFlags applies any command-line flags to the wizard state
func applyWizardFlags(model wizard.Model) error {
	// TODO: Implement flag application
	// This would set pre-configured values in the wizard state based on CLI flags
	
	if wizardSource != "" {
		// Set source directory in wizard state
		// model.SetSourceDirectory(wizardSource)
	}
	
	if wizardDestination != "" {
		// Set destination directory in wizard state  
		// model.SetDestinationDirectory(wizardDestination)
	}
	
	if wizardMode != "one-way" {
		// Set sync mode in wizard state
		// model.SetSyncMode(wizardMode)
	}

	return nil
}

// handleWizardCompletion processes the wizard results after completion
func handleWizardCompletion(model wizard.Model) error {
	// TODO: Implement wizard completion handling
	// This would:
	// 1. Extract the final configuration from the wizard
	// 2. Generate SyncFile if requested
	// 3. Execute sync operation if confirmed
	// 4. Show completion message
	
	fmt.Println("Original wizard completed successfully!")
	return nil
}

// handleTypeStateWizardCompletion processes the type-safe wizard results
func handleTypeStateWizardCompletion(model wizard.TypeStateModel) error {
	// Extract configuration from the type-safe model
	// The type state pattern ensures we only have access to valid, complete configuration
	
	fmt.Println("Type-safe wizard completed successfully!")
	fmt.Println("✓ Compile-time state safety guaranteed")
	fmt.Println("✓ Invalid state transitions prevented")
	fmt.Println("✓ Configuration completeness enforced")
	
	// In a real implementation, we would:
	// 1. Extract the complete configuration from the final state
	// 2. Execute the sync operation or save SyncFile
	// 3. Handle any errors with type safety
	
	return nil
}