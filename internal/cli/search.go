package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/ai"
	"github.com/Vedant9500/WTF/internal/config"
	"github.com/Vedant9500/WTF/internal/context"
	"github.com/Vedant9500/WTF/internal/database"
	"github.com/Vedant9500/WTF/internal/errors"
	"github.com/Vedant9500/WTF/internal/history"
	"github.com/Vedant9500/WTF/internal/recovery"
	"github.com/Vedant9500/WTF/internal/validation"

	"runtime"

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
			limit           int
			verbose         bool
			dbPath          string
			platforms       []string
			allPlatforms    bool
			noCrossPlatform bool
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

		// Use original search with smart NLP enhancement
		searchOptions := database.SearchOptions{
			Limit:          cfg.MaxResults,
			UseFuzzy:       true,
			FuzzyThreshold: -30,
			UseNLP:         true, // Enable smart NLP enhancement
		}
		if projectContext != nil {
			searchOptions.ContextBoosts = projectContext.GetContextBoosts()
		}

		// Use enhanced universal search
		results := db.SearchUniversal(query, searchOptions)

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

		// Check if we should use AI
		useAI, _ := cmd.Flags().GetBool("ai")
		aiConfig := ai.GetConfigFromEnv()
		
		// Determine if we should attempt AI generation
		// Case 1:User explicitly requested --ai
		// Case 2: No results found and AI provider is configured
		attemptAI := useAI
		if !attemptAI && len(results) == 0 && aiConfig.Provider != "" {
			attemptAI = true
		}

		if attemptAI {
			if len(results) == 0 {
				fmt.Printf("No local commands found. Asking AI (%s)...\n", aiConfig.Provider)
			} else if useAI {
				fmt.Printf("Generating AI response (%s)...\n", aiConfig.Provider)
			}

			client, err := ai.NewClient(aiConfig)
			if err != nil {
				if useAI {
					fmt.Printf("Error initializing AI client: %v\n", err)
					return
				}
				// If fallback, just log invalid config and continue to standard not found msg
			} else {
				// Construct system context
				sysCtx := fmt.Sprintf("OS: %s, Shell: %s", runtime.GOOS, os.Getenv("SHELL"))
				if projectContext != nil {
					sysCtx += fmt.Sprintf(", Project: %s", projectContext.GetContextDescription())
				}

				cmdSuggestion, err := client.GenerateCommand(cmd.Context(), query, sysCtx)
				if err != nil {
					fmt.Printf("AI generation failed: %v\n", err)
				} else {
					fmt.Printf("\n✨ AI Suggested Command:\n")
					fmt.Printf("   %s\n\n", cmdSuggestion)
					
					// If we found an AI command, we can return unless we want to show local results too (which are empty if we are here in fallback)
					if len(results) == 0 || useAI {
						return
					}
				}
			}
		}

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
				if aiConfig.Provider == "" {
					fmt.Printf("• Set WTF_AI_PROVIDER (openai, gemini, ollama) to enable AI suggestions\n")
				}
			}
			return
		}

		// Output formatting options
		format, _ := cmd.Flags().GetString("format")
		noColor, _ := cmd.Flags().GetBool("no-color")

		// Detect NO_COLOR env if flag not set
		if !noColor {
			if _, ok := os.LookupEnv("NO_COLOR"); ok {
				noColor = true
			}
		}

		// ANSI color helpers
		color := func(code string) string {
			if noColor {
				return ""
			}
			return code
		}
		reset := color("\x1b[0m")
		bold := color("\x1b[1m")
		cyan := color("\x1b[36m")
		yellow := color("\x1b[33m")
		gray := color("\x1b[90m")

		// Sort by score descending to present clearly (should already be, but ensure)
		sort.SliceStable(results, func(i, j int) bool { return results[i].Score > results[j].Score })

		switch strings.ToLower(format) {
		case "json":
			// Emit stable JSON structure
			type outItem struct {
				Command     string   `json:"command"`
				Description string   `json:"description"`
				Keywords    []string `json:"keywords,omitempty"`
				Category    string   `json:"category,omitempty"`
				Platforms   []string `json:"platforms,omitempty"`
				Score       float64  `json:"score,omitempty"`
			}
			out := make([]outItem, 0, len(results))
			for _, r := range results {
				it := outItem{
					Command:     r.Command.Command,
					Description: r.Command.Description,
					Keywords:    nil,
					Category:    r.Command.Niche,
					Platforms:   nil,
				}
				if flags.verbose {
					it.Keywords = append(it.Keywords, r.Command.Keywords...)
					it.Platforms = append(it.Platforms, r.Command.Platform...)
					it.Score = r.Score
				}
				out = append(out, it)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(out)

		case "table":
			// Simple fixed-width table (no extra deps)
			// Headers
			fmt.Printf("%s%-3s %-48s %-24s %-10s%s\n", bold, "#", "Command", "Category", "Score", reset)
			fmt.Printf("%s%s%s\n", gray, strings.Repeat("-", 90), reset)
			for i, r := range results {
				score := fmt.Sprintf("%.1f", r.Score)
				if !flags.verbose {
					score = ""
				}
				cmdStr := r.Command.Command
				if len(cmdStr) > 48 {
					cmdStr = cmdStr[:45] + "..."
				}
				cat := r.Command.Niche
				if len(cat) > 24 {
					cat = cat[:21] + "..."
				}
				fmt.Printf("%-3d %-48s %-24s %-10s\n", i+1, cmdStr, cat, score)
			}

		default: // list
			fmt.Printf("Found %d matching command(s):\n\n", len(results))
			for i, result := range results {
				fmt.Printf("%s%d.%s %s%s%s\n", bold, i+1, reset, cyan, result.Command.Command, reset)
				fmt.Printf("   %sDescription:%s %s\n", yellow, reset, result.Command.Description)
				if len(result.Command.Keywords) > 0 && flags.verbose {
					fmt.Printf("   %sKeywords:%s %s\n", yellow, reset, strings.Join(result.Command.Keywords, ", "))
				}
				if result.Command.Niche != "" {
					fmt.Printf("   %sCategory:%s %s\n", yellow, reset, result.Command.Niche)
				}
				if len(result.Command.Platform) > 0 && flags.verbose {
					fmt.Printf("   %sPlatforms:%s %s\n", yellow, reset, strings.Join(result.Command.Platform, ", "))
				}
				if flags.verbose {
					fmt.Printf("   %sRelevance:%s %.1f\n", yellow, reset, result.Score)
				}
				fmt.Println()
			}
		}

		if flags.verbose {
			fmt.Printf("Search completed in %v\n", searchDuration)
		}
	},
}

func init() {
	// Flags are inherited from parent (root) command as persistent flags
}
