package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cmd-finder/internal/config"
	"cmd-finder/internal/database"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var saveCmd = &cobra.Command{
	Use:   "save <command> <description>",
	Short: "Save a custom command to your personal notebook",
	Long: `Save a custom command to your personal command notebook.
This creates a personal database file that is automatically included in searches.

Examples:
  wtf save "tar -czf backup.tar.gz /home/user" "Create compressed backup of home directory"
  wtf save "docker ps -a --format 'table {{.Names}}\t{{.Status}}'" "Show docker containers in table format"
  wtf save "find . -name '*.go' -exec gofmt -w {} \;" "Format all Go files in current directory"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		description := args[1]

		// Get flags
		keywords, _ := cmd.Flags().GetStringSlice("keywords")
		niche, _ := cmd.Flags().GetString("category")
		platforms, _ := cmd.Flags().GetStringSlice("platforms")
		pipeline, _ := cmd.Flags().GetBool("pipeline")

		// Create command entry
		entry := database.Command{
			Command:     command,
			Description: description,
			Keywords:    keywords,
			Niche:       niche,
			Platform:    platforms,
			Pipeline:    pipeline,
		}

		// Get personal database path
		cfg := config.DefaultConfig()
		personalDBPath := cfg.GetPersonalDatabasePath()

		// Save to personal database
		err := saveToPersonalDatabase(personalDBPath, entry)
		if err != nil {
			fmt.Printf("Error saving command: %v\n", err)
			return
		}

		fmt.Printf("✅ Command saved successfully!\n")
		fmt.Printf("Command: %s\n", command)
		fmt.Printf("Description: %s\n", description)
		if len(keywords) > 0 {
			fmt.Printf("Keywords: %s\n", strings.Join(keywords, ", "))
		}
		if niche != "" {
			fmt.Printf("Category: %s\n", niche)
		}
		if len(platforms) > 0 {
			fmt.Printf("Platforms: %s\n", strings.Join(platforms, ", "))
		}
		if pipeline {
			fmt.Printf("Pipeline: true\n")
		}
		fmt.Printf("\nYour command will now appear in search results! 🎉\n")
	},
}

func init() {
	saveCmd.Flags().StringSliceP("keywords", "k", nil, "Keywords for the command (comma-separated)")
	saveCmd.Flags().StringP("category", "c", "", "Category/niche for the command")
	saveCmd.Flags().StringSliceP("platforms", "p", nil, "Supported platforms (comma-separated)")
	saveCmd.Flags().BoolP("pipeline", "", false, "Mark as a pipeline command")
}

func saveToPersonalDatabase(dbPath string, entry database.Command) error {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing personal commands if file exists
	var commands []database.Command
	if _, err := os.Stat(dbPath); err == nil {
		// File exists, load it
		data, err := os.ReadFile(dbPath)
		if err != nil {
			return fmt.Errorf("failed to read personal database: %w", err)
		}

		if err := yaml.Unmarshal(data, &commands); err != nil {
			return fmt.Errorf("failed to parse personal database: %w", err)
		}
	}

	// Check if command already exists
	for i, cmd := range commands {
		if cmd.Command == entry.Command {
			// Update existing command
			commands[i] = entry
			return writePersonalDatabase(dbPath, commands)
		}
	}

	// Add new command
	commands = append(commands, entry)
	return writePersonalDatabase(dbPath, commands)
}

func writePersonalDatabase(dbPath string, commands []database.Command) error {
	data, err := yaml.Marshal(commands)
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	err = os.WriteFile(dbPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write personal database: %w", err)
	}

	return nil
}
