package cli

import (
	"fmt"
	"strings"

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
  cmd-finder search "git commit changes"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		
		// Load database
		db, err := database.LoadDatabase("commands.yml")
		if err != nil {
			fmt.Printf("Error loading database: %v\n", err)
			return
		}
		
		fmt.Printf("Loaded %d commands from database\n", db.Size())
		fmt.Printf("Searching for: %s\n\n", query)
		
		// Perform search
		results := db.Search(query, 5)
		
		if len(results) == 0 {
			fmt.Println("No commands found matching your query.")
			return
		}
		
		// Display results
		fmt.Printf("Found %d matching command(s):\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, result.Command.Command)
			fmt.Printf("   Description: %s\n", result.Command.Description)
			if len(result.Command.Keywords) > 0 {
				fmt.Printf("   Keywords: %s\n", strings.Join(result.Command.Keywords, ", "))
			}
			if result.Command.Niche != "" {
				fmt.Printf("   Category: %s\n", result.Command.Niche)
			}
			if len(result.Command.Platform) > 0 {
				fmt.Printf("   Platforms: %s\n", strings.Join(result.Command.Platform, ", "))
			}
			fmt.Printf("   Relevance Score: %.1f\n", result.Score)
			fmt.Println()
		}
	},
}