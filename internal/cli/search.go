package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/config"
	"github.com/Vedant9500/WTF/internal/context"
	"github.com/Vedant9500/WTF/internal/database"
	"github.com/Vedant9500/WTF/internal/errors"
	"github.com/Vedant9500/WTF/internal/history"
	"github.com/Vedant9500/WTF/internal/recovery"
	"github.com/Vedant9500/WTF/internal/validation"

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
  
  # Platform-specific searches:
  wtf search --platform linux "list files"
  wtf search --platform windows,macos "compress files"
  wtf search --all-platforms "git commands"
  wtf search --platform linux --no-cross-platform "process management"
  
  # Or use directly without 'search':
  wtf "compress a directory"
  wtf --platform linux "find files by name"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
		query := strings.Join(args, " ")

		// Validate and sanitize query
		cleanQuery, err := validation.ValidateQuery(query)
		if err != nil {
			fmt.Printf("%s\n", errors.GetUserFriendlyMessage(err))
			if suggestions := errors.GetErrorSuggestions(err); len(suggestions) > 0 {
				fmt.Printf("\nSuggestions:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("• %s\n", suggestion)
				}
			}
			return
		}
		query = cleanQuery

		// Get flags once at the beginning
		flags := struct {
			limit            int
			verbose          bool
			dbPath           string
			platforms        []string
			allPlatforms     bool
			noCrossPlatform  bool
		}{}
		flags.limit, _ = cmd.Flags().GetInt("limit")
		flags.verbose, _ = cmd.Flags().GetBool("verbose")
		flags.dbPath, _ = cmd.Flags().GetString("database")
		flags.platforms, _ = cmd.Flags().GetStringSlice("platform")
		flags.allPlatforms, _ = cmd.Flags().GetBool("all-platforms")
		flags.noCrossPlatform, _ = cmd.Flags().GetBool("no-cross-platform")

		// Validate limit
		validLimit, err := validation.ValidateLimit(flags.limit)
		if err != nil {
			fmt.Printf("%s\n", errors.GetUserFriendlyMessage(err))
			if suggestions := errors.GetErrorSuggestions(err); len(suggestions) > 0 {
				fmt.Printf("\nSuggestions:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("• %s\n", suggestion)
				}
			}
			return
		}
		flags.limit = validLimit

		// Load configuration
		cfg := config.DefaultConfig()
		if flags.limit > 0 {
			cfg.MaxResults = flags.limit
		}
		if flags.dbPath != "" {
			cfg.DatabasePath = flags.dbPath
		}

		// Analyze current directory context
		analyzer := context.NewAnalyzer()
		projectContext, err := analyzer.AnalyzeCurrentDirectory()
		if err != nil && flags.verbose {
			fmt.Printf("Warning: Could not analyze directory context: %v\n", err)
		}

		// Load database (main + personal) with recovery mechanisms
		dbFilePath := cfg.GetDatabasePath()
		personalDBPath := cfg.GetPersonalDatabasePath()
		
		// Use database recovery for robust loading
		dbRecovery := recovery.NewDatabaseRecovery(recovery.DefaultRetryConfig())
		db, err := dbRecovery.LoadDatabaseWithFallback(dbFilePath, personalDBPath)
		if err != nil {
			// Use user-friendly error messages
			fmt.Printf("%s\n", errors.GetUserFriendlyMessage(err))
			
			// Show suggestions if available
			if suggestions := errors.GetErrorSuggestions(err); len(suggestions) > 0 {
				fmt.Printf("\nSuggestions:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("• %s\n", suggestion)
				}
			}
			return
		}

		if flags.verbose {
			fmt.Printf("Loaded %d commands from database: %s\n", db.Size(), dbFilePath)
			if projectContext != nil {
				fmt.Printf("Context detected: %s\n", projectContext.GetContextDescription())
			}
			
			// Show platform filtering info
			if flags.allPlatforms {
				fmt.Printf("Platform filter: All platforms (no filtering)\n")
			} else if len(flags.platforms) > 0 {
				fmt.Printf("Platform filter: %v", flags.platforms)
				if !flags.noCrossPlatform {
					fmt.Printf(" + cross-platform")
				}
				fmt.Printf("\n")
			} else {
				fmt.Printf("Platform filter: None (showing all platforms)\n")
			}
		}
		fmt.Printf("Searching for: %s\n\n", query)

		// Use the original database search which was working well
		searchOptions := database.SearchOptions{
			Limit:          cfg.MaxResults,
			UseFuzzy:       true,
			FuzzyThreshold: -30,
			UseNLP:         true,
		}
		if projectContext != nil {
			searchOptions.ContextBoosts = projectContext.GetContextBoosts()
		}
		
		// Use the robust database search instead of the over-engineered enhanced search
		results := db.SearchWithNLP(query, searchOptions)
		
		searchDuration := time.Since(startTime)
		
		// If search failed, try recovery mechanisms
		if len(results) == 0 {
			searchRecovery := recovery.NewSearchRecovery()
			recoveredResults, recoveryErr := searchRecovery.RecoverFromSearchFailure(query, nil, db)
			if recoveryErr == nil && len(recoveredResults) > 0 {
				results = recoveredResults
			}
		}

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
			// Use database suggestions
			suggestions := db.GetSuggestions(query, 5)
			
			fmt.Printf("No commands found matching '%s'.\n\n", query)
			
			if len(suggestions) > 0 {
				fmt.Printf("Did you mean:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("• %s\n", suggestion)
				}
				fmt.Printf("\nTry: wtf \"%s\"\n", suggestions[0])
			} else {
				fmt.Printf("Suggestions:\n")
				fmt.Printf("• Try using different keywords\n")
				fmt.Printf("• Check for typos in your query\n")
				fmt.Printf("• Be more specific or more general\n")
				fmt.Printf("• Use simpler terms (e.g., 'compress files' instead of 'how do I compress files')\n")
				if len(flags.platforms) > 0 {
					fmt.Printf("• Try --all-platforms to search across all platforms\n")
				}
			}
			return
		}

		// Display results
		fmt.Printf("Found %d matching command(s):\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, result.Command.Command)
			fmt.Printf("   Description: %s\n", result.Command.Description)
			if len(result.Command.Keywords) > 0 && flags.verbose {
				fmt.Printf("   Keywords: %s\n", strings.Join(result.Command.Keywords, ", "))
			}
			if result.Command.Niche != "" {
				fmt.Printf("   Category: %s\n", result.Command.Niche)
			}
			if len(result.Command.Platform) > 0 && flags.verbose {
				fmt.Printf("   Platforms: %s\n", strings.Join(result.Command.Platform, ", "))
			}
			if flags.verbose {
				fmt.Printf("   Relevance Score: %.1f\n", result.Score)
			}
			fmt.Println()
		}

		if flags.verbose {
			fmt.Printf("Search completed in %v\n", searchDuration)
		}
	},
}

func init() {
	// Flags are inherited from parent (root) command as persistent flags
}
