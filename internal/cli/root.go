package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cmd-finder [query]",
	Short: "A CLI tool to find shell commands using natural language queries",
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
}