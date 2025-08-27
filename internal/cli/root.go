// Package cli provides the command-line interface for the WTF application.
//
// This package implements all CLI commands and their associated functionality using
// the Cobra CLI framework. It includes:
//   - Root command with global flags and configuration
//   - Search command for finding commands by natural language queries
//   - History management for tracking search queries
//   - Setup and configuration commands
//   - Pipeline-specific search functionality
//   - Command aliasing and saving capabilities
//
// The Execute function is the main entry point for the CLI application.
package cli

import (
	"github.com/Vedant9500/WTF/internal/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "wtf [query]",
	Short:   "What's The Function - A CLI tool to find shell commands using natural language",
	Version: version.Version,
	Long: `WTF (What's The Function) helps you discover shell commands by searching through a curated database
of common command-line tools and their usage examples. Simply describe what you want to do
in natural language, and WTF will suggest relevant commands.

When you can't remember a command, you think "What's The Function I need?" - that's WTF! 😄`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand specified, default to search
		searchCmd.Run(cmd, args)
	},
}

// Execute starts the CLI application
func Execute() error {
	return rootCmd.Execute()
}

// NewRootCommand creates a new root command for testing
func NewRootCommand() *cobra.Command {
	// Create a new command similar to rootCmd but for testing
	testRootCmd := &cobra.Command{
		Use:     "wtf [query]",
		Short:   "What's The Function - A CLI tool to find shell commands using natural language",
		Version: version.Version,
		Long: `WTF (What's The Function) helps you discover shell commands by searching through a curated database
of common command-line tools and their usage examples. Simply describe what you want to do
in natural language, and WTF will suggest relevant commands.

When you can't remember a command, you think "What's The Function I need?" - that's WTF! 😄`,
		Args: cobra.MinimumNArgs(1),
		Run:  rootCmd.Run, // Use the same run function
	}

	// Add the search command to the test root command
	testRootCmd.AddCommand(searchCmd)

	return testRootCmd
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(wizardCmd)
	rootCmd.AddCommand(pipelineCmd)
	rootCmd.AddCommand(savePipelineCmd)
	rootCmd.AddCommand(historyCmd)

	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("database", "d", "", "Path to custom database file")
	rootCmd.PersistentFlags().IntP("limit", "l", 0, "Maximum number of results to display (default: 5)")
	rootCmd.PersistentFlags().StringSliceP("platform", "p", []string{}, "Filter by platform (linux, macos, windows, cross-platform)")
	rootCmd.PersistentFlags().BoolP("all-platforms", "a", false, "Show commands from all platforms (ignore platform filtering)")
	rootCmd.PersistentFlags().BoolP("no-cross-platform", "", false, "Exclude cross-platform commands when using platform filter")

	// Output formatting flags
	rootCmd.PersistentFlags().String("format", "list", "Output format: list|table|json")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output (or set NO_COLOR env)")
}
