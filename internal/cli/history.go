package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/history"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [pattern]",
	Short: "View and manage search history",
	Long: `View your search history and get quick access to recent searches.
	
Examples:
  wtf history                    # Show recent searches
  wtf history docker             # Show searches containing "docker"
  wtf history --top              # Show most frequent searches
  wtf history --stats            # Show usage statistics
  wtf history --clear            # Clear all history`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		showTop, _ := cmd.Flags().GetBool("top")
		showStats, _ := cmd.Flags().GetBool("stats")
		clearHistory, _ := cmd.Flags().GetBool("clear")
		limit, _ := cmd.Flags().GetInt("limit")

		// Initialize search history
		historyPath := history.DefaultHistoryPath()
		searchHistory := history.NewSearchHistory(historyPath, 100)

		err := searchHistory.Load()
		if err != nil {
			fmt.Printf("Error loading search history: %v\n", err)
			return
		}

		// Handle clear flag
		if clearHistory {
			err := searchHistory.Clear()
			if err != nil {
				fmt.Printf("Error clearing history: %v\n", err)
				return
			}
			fmt.Println("Search history cleared successfully.")
			return
		}

		// Handle stats flag
		if showStats {
			stats := searchHistory.GetStats()
			fmt.Println("📊 Search History Statistics")
			fmt.Println("=" + strings.Repeat("=", 28))
			fmt.Printf("Total searches: %d\n", stats.TotalSearches)
			fmt.Printf("Unique queries: %d\n", stats.UniqueQueries)
			fmt.Printf("Average results per search: %.1f\n", stats.AvgResultsPerSearch)
			if stats.AvgSearchDuration > 0 {
				fmt.Printf("Average search duration: %.1fms\n", stats.AvgSearchDuration)
			}
			if !stats.OldestEntry.IsZero() {
				fmt.Printf("First search: %s\n", stats.OldestEntry.Format("2006-01-02 15:04"))
				fmt.Printf("Last search: %s\n", stats.NewestEntry.Format("2006-01-02 15:04"))
			}
			return
		}

		// Handle top queries flag
		if showTop {
			topQueries := searchHistory.GetTopQueries(limit)
			if len(topQueries) == 0 {
				fmt.Println("No search history found.")
				return
			}

			fmt.Println("🔥 Most Frequent Searches")
			fmt.Println("=" + strings.Repeat("=", 24))
			for i, qf := range topQueries {
				fmt.Printf("%d. \"%s\" (%d times, last used: %s)\n",
					i+1, qf.Query, qf.Count, qf.LastUsed.Format("Jan 2 15:04"))
			}
			return
		}

		// Handle pattern search
		if len(args) > 0 {
			pattern := strings.Join(args, " ")
			entries := searchHistory.GetEntriesByPattern(pattern)

			if len(entries) == 0 {
				fmt.Printf("No searches found matching: %s\n", pattern)
				return
			}

			fmt.Printf("🔍 Searches matching \"%s\"\n", pattern)
			fmt.Println("=" + strings.Repeat("=", len(pattern)+20))

			for i, entry := range entries {
				if i >= limit {
					break
				}

				timeAgo := formatTimeAgo(time.Since(entry.Timestamp))
				fmt.Printf("%d. \"%s\" (%d results, %s)\n",
					i+1, entry.Query, entry.ResultsCount, timeAgo)

				if entry.Context != "" {
					fmt.Printf("   Context: %s\n", entry.Context)
				}
			}
			return
		}

		// Show recent searches (default behavior)
		recentQueries := searchHistory.GetRecentQueries(limit)
		if len(recentQueries) == 0 {
			fmt.Println("No search history found.")
			fmt.Println("Start searching to build your history: wtf \"your query\"")
			return
		}

		fmt.Println("📚 Recent Searches")
		fmt.Println("=" + strings.Repeat("=", 17))
		for i, query := range recentQueries {
			fmt.Printf("%d. %s\n", i+1, query)
		}

		fmt.Printf("\nTo run a search again: wtf \"%s\"\n", recentQueries[0])
		fmt.Println("For more options: wtf history --help")
	},
}

// formatTimeAgo formats a duration as a human-readable "time ago" string
func formatTimeAgo(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(d.Hours()) / 24
		if days == 1 {
			return "1 day ago"
		} else if days < 7 {
			return fmt.Sprintf("%d days ago", days)
		} else if days < 30 {
			weeks := days / 7
			if weeks == 1 {
				return "1 week ago"
			}
			return fmt.Sprintf("%d weeks ago", weeks)
		} else {
			months := days / 30
			if months == 1 {
				return "1 month ago"
			}
			return fmt.Sprintf("%d months ago", months)
		}
	}
}

func init() {
	historyCmd.Flags().BoolP("top", "t", false, "Show most frequent searches")
	historyCmd.Flags().BoolP("stats", "s", false, "Show search statistics")
	historyCmd.Flags().BoolP("clear", "c", false, "Clear all search history")
	historyCmd.Flags().IntP("limit", "l", 10, "Maximum number of entries to show")
}
