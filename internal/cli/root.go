package cli

import (
	"cmd-finder/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "cmd-finder [query]",
	Short:   "A CLI tool to find shell commands using natural language queries",
	Version: version.Version,
	Long: `cmd-finder helps you discover shell commands by searching through a curated database
of common command-line tools and their usage examples. Simply describe what you want to do
in natural language, and cmd-finder will suggest relevant commands.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand specified, default to search
		searchCmd.Run(cmd, args)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(searchCmd)
	
	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("database", "d", "", "Path to custom database file")
	rootCmd.PersistentFlags().IntP("limit", "l", 0, "Maximum number of results to display (default: 5)")
}