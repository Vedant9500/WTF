package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/config"
	"github.com/Vedant9500/WTF/internal/context"
	"github.com/Vedant9500/WTF/internal/database"
	"github.com/Vedant9500/WTF/internal/history"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for commands using natural language",
	Long: `Search the command database using natural language queries.
	
Examples:
  wtf search "compress a directory"
  wtf search "find files by name"
  wtf search "git commit changes"
  wtf search --limit 10 "docker commands"
  
  # Or use directly without 'search':
  wtf "compress a directory"
  hey "find files by name"  # if you set up 'hey' as an alias`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
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

		// Analyze current directory context
		analyzer := context.NewAnalyzer()
		projectContext, err := analyzer.AnalyzeCurrentDirectory()
		if err != nil && verbose {
			fmt.Printf("Warning: Could not analyze directory context: %v\n", err)
		}

		// Load database (main + personal)
		dbFilePath := cfg.GetDatabasePath()
		personalDBPath := cfg.GetPersonalDatabasePath()
		db, err := database.LoadDatabaseWithPersonal(dbFilePath, personalDBPath)
		if err != nil {
			fmt.Printf("Error loading database from %s: %v\n", dbFilePath, err)
			fmt.Println("Make sure the commands.yml file exists in the current directory.")
			return
		}

		if verbose {
			fmt.Printf("Loaded %d commands from database: %s\n", db.Size(), dbFilePath)
			if projectContext != nil {
				fmt.Printf("Context detected: %s\n", projectContext.GetContextDescription())
			}
		}
		fmt.Printf("Searching for: %s\n\n", query)

		// Prepare search options with context boosts and fuzzy search
		searchOptions := database.SearchOptions{
			Limit:          cfg.MaxResults,
			UseFuzzy:       true,  // Enable fuzzy search for better typo handling
			FuzzyThreshold: -30,   // Reasonable threshold for fuzzy matches
		}
		if projectContext != nil {
			searchOptions.ContextBoosts = projectContext.GetContextBoosts()
		}

		// Perform context-aware search with fuzzy capabilities
		results := db.SearchWithFuzzy(query, searchOptions)
		searchDuration := time.Since(startTime)

		// Record search in history
		historyPath := history.DefaultHistoryPath()
		searchHistory := history.NewSearchHistory(historyPath, 100)
		_ = searchHistory.Load() // Ignore errors for history loading
		
		contextDesc := ""
		if projectContext != nil {
			contextDesc = projectContext.GetContextDescription()
		}
		
		searchHistory.AddEntry(query, len(results), contextDesc, searchDuration)
		_ = searchHistory.Save() // Ignore errors for history saving

		if len(results) == 0 {
			fmt.Println("No commands found matching your query.")
			
			// Provide suggestions for potential typos
			suggestions := db.GetSuggestions(query, 3)
			if len(suggestions) > 0 {
				fmt.Printf("\nDid you mean:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("  • %s\n", suggestion)
				}
				fmt.Printf("\nTry: wtf \"%s\"\n", suggestions[0])
			}
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
		
		if verbose {
			fmt.Printf("Search completed in %v\n", searchDuration)
		}
	},
}

func init() {
	// Flags are inherited from parent (root) command as persistent flags
}
