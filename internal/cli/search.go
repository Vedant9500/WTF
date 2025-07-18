package cli

import (
	"fmt"
	"strings"

	"cmd-finder/internal/config"
	"cmd-finder/internal/database"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for commands using natural language",
	Long: `Search the command database using natural language queries.
	
Examples:
  cmd-finder search "compress a directory"
  cmd-finder search "find files by name"
  cmd-finder search "git commit changes"
  cmd-finder search --limit 10 "docker commands"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		
		// Get flags
		limit, _ := cmd.Flags().GetInt("limit")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dbPath, _ := cmd.Flags().GetString("database")
		
		// Load configuration
		cfg := config.DefaultConfig()
		if limit > 0 {
			cfg.MaxResults = limit
		}
		if dbPath != "" {
			cfg.DatabasePath = dbPath
		}
		
		// Load database
		dbFilePath := cfg.GetDatabasePath()
		db, err := database.LoadDatabase(dbFilePath)
		if err != nil {
			fmt.Printf("Error loading database from %s: %v\n", dbFilePath, err)
			fmt.Println("Make sure the commands.yml file exists in the current directory.")
			return
		}
		
		if verbose {
			fmt.Printf("Loaded %d commands from database: %s\n", db.Size(), dbFilePath)
		}
		fmt.Printf("Searching for: %s\n\n", query)
		
		// Perform search
		results := db.Search(query, cfg.MaxResults)
		
		if len(results) == 0 {
			fmt.Println("No commands found matching your query.")
			return
		}
		
		// Display results
		fmt.Printf("Found %d matching command(s):\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, result.Command.Command)
			fmt.Printf("   Description: %s\n", result.Command.Description)
			if len(result.Command.Keywords) > 0 && verbose {
				fmt.Printf("   Keywords: %s\n", strings.Join(result.Command.Keywords, ", "))
			}
			if result.Command.Niche != "" {
				fmt.Printf("   Category: %s\n", result.Command.Niche)
			}
			if len(result.Command.Platform) > 0 && verbose {
				fmt.Printf("   Platforms: %s\n", strings.Join(result.Command.Platform, ", "))
			}
			if verbose {
				fmt.Printf("   Relevance Score: %.1f\n", result.Score)
			}
			fmt.Println()
		}
	},
}

func init() {
	// Flags are inherited from parent (root) command as persistent flags
}