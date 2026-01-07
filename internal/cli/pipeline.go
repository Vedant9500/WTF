package cli

import (
	"fmt"
	"strings"

	"github.com/Vedant9500/WTF/internal/config"
	"github.com/Vedant9500/WTF/internal/context"
	"github.com/Vedant9500/WTF/internal/database"

	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline [query]",
	Short: "Search for command pipelines and multi-step recipes",
	Long: `Search for command pipelines - sequences of commands connected with pipes.
Perfect for finding complex workflows and multi-step operations.

Examples:
  wtf pipeline "process text"     # Find text processing pipelines
  wtf pipeline "find and replace" # Find search-and-replace pipelines
  wtf pipeline "data extraction"  # Find data extraction workflows
  wtf pipeline "log analysis"     # Find log analysis pipelines`,
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
		fmt.Printf("🔍 Searching for pipelines: %s\n\n", query)

		// Prepare search options with context boosts and pipeline focus
		searchOptions := database.SearchOptions{
			Limit:         cfg.MaxResults,
			PipelineOnly:  true, // New option to focus on pipelines
			PipelineBoost: 2.0,  // Boost pipeline commands
		}
		if projectContext != nil {
			searchOptions.ContextBoosts = projectContext.GetContextBoosts()
		}

		// Perform pipeline-focused search
		results := db.SearchWithPipelineOptions(query, searchOptions)

		if len(results) == 0 {
			fmt.Println("No pipeline commands found matching your query.")
			fmt.Println("\n💡 Try broader terms like:")
			fmt.Println("   • 'text processing'")
			fmt.Println("   • 'file manipulation'")
			fmt.Println("   • 'data extraction'")
			fmt.Println("   • 'log analysis'")
			return
		}

		// Display pipeline results with enhanced formatting
		fmt.Printf("📋 Found %d pipeline command(s):\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, formatPipelineCommand(result.Command.Command))
			fmt.Printf("   📝 %s\n", result.Command.Description)

			if len(result.Command.Keywords) > 0 && verbose {
				fmt.Printf("   🏷️  Keywords: %s\n", strings.Join(result.Command.Keywords, ", "))
			}
			if result.Command.Niche != "" {
				fmt.Printf("   📂 Category: %s\n", result.Command.Niche)
			}
			if len(result.Command.Platform) > 0 && verbose {
				fmt.Printf("   💻 Platforms: %s\n", strings.Join(result.Command.Platform, ", "))
			}
			if verbose {
				fmt.Printf("   ⭐ Relevance Score: %.1f\n", result.Score)
			}

			// Show pipeline breakdown
			if strings.Contains(result.Command.Command, "|") {
				fmt.Printf("   🔗 Pipeline steps:\n")
				steps := strings.Split(result.Command.Command, "|")
				for j, step := range steps {
					fmt.Printf("      %d. %s\n", j+1, strings.TrimSpace(step))
				}
			}

			fmt.Println()
		}

		// Show pipeline tips
		if verbose {
			fmt.Println("💡 Pipeline Tips:")
			fmt.Println("   • Use | to connect commands")
			fmt.Println("   • Each command processes output from the previous")
			fmt.Println("   • Test each step individually first")
			fmt.Println("   • Use 'tee' to save intermediate results")
		}
	},
}

var savePipelineCmd = &cobra.Command{
	Use:   "save-pipeline <name> <command>",
	Short: "Save a custom pipeline to your personal notebook",
	Long: `Save a custom command pipeline to your personal command notebook.
This will mark the command as a pipeline and make it easier to find with pipeline searches.

Examples:
  wtf save-pipeline "text-stats" "cat file.txt | wc -l | awk '{print \"Lines:\" $1}'"
  wtf save-pipeline "top-files" "find . -type f | head -10 | sort"
  wtf save-pipeline "log-errors" "grep ERROR /var/log/app.log | tail -20 | sort"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]

		// Get flags
		keywords, _ := cmd.Flags().GetStringSlice("keywords")
		niche, _ := cmd.Flags().GetString("category")
		platforms, _ := cmd.Flags().GetStringSlice("platforms")

		// Auto-detect pipeline characteristics
		pipelineSteps := strings.Split(command, "|")
		stepCount := len(pipelineSteps)

		// Auto-generate description if not provided
		description := fmt.Sprintf("%s - %d-step pipeline", name, stepCount)
		if desc, _ := cmd.Flags().GetString("description"); desc != "" {
			description = desc
		}

		// Auto-generate keywords based on pipeline
		autoKeywords := []string{"pipeline", "workflow"}
		if strings.Contains(command, "grep") {
			autoKeywords = append(autoKeywords, "search", "filter")
		}
		if strings.Contains(command, "awk") || strings.Contains(command, "sed") {
			autoKeywords = append(autoKeywords, "text", "processing")
		}
		if strings.Contains(command, "sort") {
			autoKeywords = append(autoKeywords, "sort", "order")
		}
		if strings.Contains(command, "find") {
			autoKeywords = append(autoKeywords, "find", "search")
		}

		// Merge user keywords with auto-detected ones
		allKeywords := append(autoKeywords, keywords...)

		// Create command entry
		entry := database.Command{
			Command:     command,
			Description: description,
			Keywords:    allKeywords,
			Niche:       niche,
			Platform:    platforms,
			Pipeline:    true, // Mark as pipeline
		}

		// Get personal database path
		cfg := config.DefaultConfig()
		personalDBPath := cfg.GetPersonalDatabasePath()

		// Save to personal database using existing save functionality
		err := saveToPersonalDatabase(personalDBPath, entry)
		if err != nil {
			fmt.Printf("Error saving pipeline: %v\n", err)
			return
		}

		fmt.Printf("Pipeline saved successfully!\n")
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Command: %s\n", command)
		fmt.Printf("Description: %s\n", description)
		fmt.Printf("Steps: %d\n", stepCount)

		if len(allKeywords) > 0 {
			fmt.Printf("Keywords: %s\n", strings.Join(allKeywords, ", "))
		}
		if niche != "" {
			fmt.Printf("Category: %s\n", niche)
		}
		if len(platforms) > 0 {
			fmt.Printf("Platforms: %s\n", strings.Join(platforms, ", "))
		}

		fmt.Println("\nYour pipeline will now appear in searches!")
		fmt.Printf("Try: wtf pipeline \"%s\"\n", name)
	},
}

func init() {
	// Add flags to save-pipeline command
	savePipelineCmd.Flags().StringSliceP("keywords", "k", nil, "Keywords for the pipeline (comma-separated)")
	savePipelineCmd.Flags().StringP("category", "c", "", "Category/niche for the pipeline")
	savePipelineCmd.Flags().StringSliceP("platforms", "p", nil, "Supported platforms (comma-separated)")
	savePipelineCmd.Flags().String("description", "", "Custom description for the pipeline")
}

func formatPipelineCommand(command string) string {
	// Add visual formatting to make pipeline commands more readable
	if !strings.Contains(command, "|") {
		return command
	}

	// Add colored pipe symbols and better spacing
	formatted := strings.ReplaceAll(command, "|", " │ ")
	return formatted
}
