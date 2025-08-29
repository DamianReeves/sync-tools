package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.2.0" // Incremented from Python version
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sync-tools",
	Short: "Fast directory sync with .syncignore, unignore, whitelist, optional .gitignore import, and hidden-dir controls",
	Long: `sync-tools is a Go CLI wrapper around rsync that provides:

• One-way or two-way directory synchronization
• Gitignore-style .syncignore files (source and destination)
• Optional import of SOURCE/.gitignore patterns
• Per-side ignore files and inline patterns (with ! unignore)
• "Whitelist" mode to sync only specified paths
• Optional config file (sync.conf) OR pure CLI usage
• Default exclusion of .git/
• Optional exclusion of all hidden directories (dot-dirs)
• Dry-run previews and detailed change output
• Interactive sync mode with Bubble Tea UI
• Custom SyncFile format (Dockerfile-like syntax)`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to a TOML config file to load default options")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Verbose output (use -v, -vv, etc.)")
}