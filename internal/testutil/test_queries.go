package testutil

import (
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
)

// TestQuerySets provides various predefined test query sets
type TestQuerySets struct{}

// NewTestQuerySets creates a new test query sets provider
func NewTestQuerySets() *TestQuerySets {
	return &TestQuerySets{}
}

// GetBasicQueries returns basic search queries for testing
func (tqs *TestQuerySets) GetBasicQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "git commit",
			ExpectedResults: 1,
			MinScore:        15.0,
			MaxScore:        50.0,
			ShouldContain:   []string{"git", "commit"},
			ShouldNotContain: []string{"find", "tar", "zip"},
		},
		{
			Query:           "find files",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"find", "files"},
			ShouldNotContain: []string{"git", "commit"},
		},
		{
			Query:           "compress archive",
			ExpectedResults: 2, // tar and zip commands
			MinScore:        8.0,
			MaxScore:        35.0,
			ShouldContain:   []string{"compress", "archive"},
			ShouldNotContain: []string{"git", "find"},
		},
		{
			Query:           "search pattern",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"grep", "search"},
			ShouldNotContain: []string{"git", "tar"},
		},
		{
			Query:           "download file",
			ExpectedResults: 2, // curl and wget
			MinScore:        8.0,
			MaxScore:        35.0,
			ShouldContain:   []string{"download"},
			ShouldNotContain: []string{"git", "find"},
		},
	}
}

// GetAdvancedQueries returns advanced search queries for testing
func (tqs *TestQuerySets) GetAdvancedQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "git repository clone",
			ExpectedResults: 1,
			MinScore:        20.0,
			MaxScore:        60.0,
			ShouldContain:   []string{"git", "clone", "repository"},
			ShouldNotContain: []string{"find", "tar"},
		},
		{
			Query:           "create directory folder",
			ExpectedResults: 1,
			MinScore:        15.0,
			MaxScore:        45.0,
			ShouldContain:   []string{"mkdir", "create", "directory"},
			ShouldNotContain: []string{"git", "download"},
		},
		{
			Query:           "remote ssh connect",
			ExpectedResults: 1,
			MinScore:        18.0,
			MaxScore:        50.0,
			ShouldContain:   []string{"ssh", "remote", "connect"},
			ShouldNotContain: []string{"git", "find"},
		},
		{
			Query:           "extract uncompress archive",
			ExpectedResults: 2, // tar and unzip
			MinScore:        12.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"extract", "archive"},
			ShouldNotContain: []string{"git", "ssh"},
		},
		{
			Query:           "list files directory detailed",
			ExpectedResults: 1,
			MinScore:        15.0,
			MaxScore:        45.0,
			ShouldContain:   []string{"ls", "list", "files"},
			ShouldNotContain: []string{"git", "download"},
		},
	}
}

// GetEdgeCaseQueries returns edge case queries for testing
func (tqs *TestQuerySets) GetEdgeCaseQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "",
			ExpectedResults: 0,
			MinScore:        0.0,
			MaxScore:        0.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "a",
			ExpectedResults: 0, // Single character queries should be ignored
			MinScore:        0.0,
			MaxScore:        0.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "nonexistent command that does not exist",
			ExpectedResults: 0,
			MinScore:        0.0,
			MaxScore:        0.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "UPPERCASE QUERY",
			ExpectedResults: 0, // Should handle case insensitivity
			MinScore:        0.0,
			MaxScore:        10.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "query with many words that should not match anything specific",
			ExpectedResults: 0,
			MinScore:        0.0,
			MaxScore:        5.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "special!@#$%^&*()characters",
			ExpectedResults: 0,
			MinScore:        0.0,
			MaxScore:        0.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
	}
}

// GetPipelineQueries returns pipeline-specific queries for testing
func (tqs *TestQuerySets) GetPipelineQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "process grep",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"ps", "grep", "process"},
			ShouldNotContain: []string{"git", "find"},
		},
		{
			Query:           "sort unique",
			ExpectedResults: 1,
			MinScore:        8.0,
			MaxScore:        35.0,
			ShouldContain:   []string{"sort", "uniq"},
			ShouldNotContain: []string{"git", "ssh"},
		},
		{
			Query:           "history search",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"history", "grep"},
			ShouldNotContain: []string{"git", "tar"},
		},
		{
			Query:           "find delete log",
			ExpectedResults: 1,
			MinScore:        12.0,
			MaxScore:        45.0,
			ShouldContain:   []string{"find", "rm", "log"},
			ShouldNotContain: []string{"git", "ssh"},
		},
	}
}

// GetPlatformSpecificQueries returns platform-specific queries for testing
func (tqs *TestQuerySets) GetPlatformSpecificQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "windows dir list",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"dir", "list"},
			ShouldNotContain: []string{"ls", "git"},
		},
		{
			Query:           "windows copy file",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"copy", "file"},
			ShouldNotContain: []string{"cp", "git"},
		},
		{
			Query:           "windows delete file",
			ExpectedResults: 1,
			MinScore:        10.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"del", "delete"},
			ShouldNotContain: []string{"rm", "git"},
		},
		{
			Query:           "linux unix ls",
			ExpectedResults: 1,
			MinScore:        8.0,
			MaxScore:        35.0,
			ShouldContain:   []string{"ls", "list"},
			ShouldNotContain: []string{"dir", "windows"},
		},
	}
}

// GetPerformanceQueries returns queries for performance testing
func (tqs *TestQuerySets) GetPerformanceQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "git",
			ExpectedResults: 8, // All git commands
			MinScore:        4.0,
			MaxScore:        50.0,
			ShouldContain:   []string{"git"},
			ShouldNotContain: []string{},
		},
		{
			Query:           "file",
			ExpectedResults: 5, // Multiple file-related commands
			MinScore:        1.0,
			MaxScore:        30.0,
			ShouldContain:   []string{},
			ShouldNotContain: []string{},
		},
		{
			Query:           "archive",
			ExpectedResults: 4, // Archive-related commands
			MinScore:        4.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"archive"},
			ShouldNotContain: []string{},
		},
		{
			Query:           "download",
			ExpectedResults: 2, // curl and wget
			MinScore:        6.0,
			MaxScore:        40.0,
			ShouldContain:   []string{"download"},
			ShouldNotContain: []string{},
		},
	}
}

// GetContextBoostQueries returns queries for testing context boosts
func (tqs *TestQuerySets) GetContextBoostQueries() []TestQuery {
	return []TestQuery{
		{
			Query:           "git commit",
			ExpectedResults: 1,
			MinScore:        30.0, // Higher due to context boost
			MaxScore:        100.0,
			ShouldContain:   []string{"git", "commit"},
			ShouldNotContain: []string{"find", "tar"},
		},
		{
			Query:           "repository clone",
			ExpectedResults: 1,
			MinScore:        25.0, // Higher due to context boost
			MaxScore:        80.0,
			ShouldContain:   []string{"git", "clone"},
			ShouldNotContain: []string{"find", "tar"},
		},
		{
			Query:           "version control",
			ExpectedResults: 1,
			MinScore:        8.0,
			MaxScore:        50.0,
			ShouldContain:   []string{"git"},
			ShouldNotContain: []string{"find", "tar"},
		},
	}
}

// GetAllTestQueries returns all test queries combined
func (tqs *TestQuerySets) GetAllTestQueries() []TestQuery {
	var allQueries []TestQuery
	
	allQueries = append(allQueries, tqs.GetBasicQueries()...)
	allQueries = append(allQueries, tqs.GetAdvancedQueries()...)
	allQueries = append(allQueries, tqs.GetPipelineQueries()...)
	allQueries = append(allQueries, tqs.GetPlatformSpecificQueries()...)
	
	return allQueries
}

// GetTestQueriesByCategory returns queries filtered by category
func (tqs *TestQuerySets) GetTestQueriesByCategory(category string) []TestQuery {
	switch category {
	case "basic":
		return tqs.GetBasicQueries()
	case "advanced":
		return tqs.GetAdvancedQueries()
	case "edge":
		return tqs.GetEdgeCaseQueries()
	case "pipeline":
		return tqs.GetPipelineQueries()
	case "platform":
		return tqs.GetPlatformSpecificQueries()
	case "performance":
		return tqs.GetPerformanceQueries()
	case "context":
		return tqs.GetContextBoostQueries()
	default:
		return tqs.GetAllTestQueries()
	}
}

// GetTestQueriesForCommands returns queries that should match specific commands
func (tqs *TestQuerySets) GetTestQueriesForCommands(commands []database.Command) []TestQuery {
	var queries []TestQuery
	
	for _, cmd := range commands {
		// Create a query based on the first keyword
		if len(cmd.Keywords) > 0 {
			query := TestQuery{
				Query:           cmd.Keywords[0],
				ExpectedResults: 1,
				MinScore:        4.0,
				MaxScore:        50.0,
				ShouldContain:   []string{cmd.Keywords[0]},
				ShouldNotContain: []string{},
			}
			queries = append(queries, query)
		}
		
		// Create a query based on command name (first word)
		if cmd.Command != "" {
			cmdParts := strings.Fields(cmd.Command)
			if len(cmdParts) > 0 {
				query := TestQuery{
					Query:           cmdParts[0],
					ExpectedResults: 1,
					MinScore:        10.0,
					MaxScore:        50.0,
					ShouldContain:   []string{cmdParts[0]},
					ShouldNotContain: []string{},
				}
				queries = append(queries, query)
			}
		}
	}
	
	return queries
}

// GetRandomTestQueries generates random test queries for stress testing
func (tqs *TestQuerySets) GetRandomTestQueries(count int) []TestQuery {
	queries := make([]TestQuery, count)
	
	baseWords := []string{
		"git", "find", "grep", "tar", "zip", "curl", "wget", "ssh", "ls", "cp",
		"mv", "rm", "mkdir", "cat", "sort", "uniq", "head", "tail", "awk", "sed",
		"file", "directory", "search", "download", "upload", "compress", "extract",
		"create", "delete", "copy", "move", "list", "show", "display", "edit",
	}
	
	for i := 0; i < count; i++ {
		// Generate random query with 1-3 words
		queryWords := make([]string, 1+i%3)
		for j := range queryWords {
			queryWords[j] = baseWords[(i+j)%len(baseWords)]
		}
		
		queries[i] = TestQuery{
			Query:           strings.Join(queryWords, " "),
			ExpectedResults: i%5 + 1, // 1-5 expected results
			MinScore:        float64(i%10 + 1),
			MaxScore:        float64(i%20 + 20),
			ShouldContain:   []string{queryWords[0]},
			ShouldNotContain: []string{},
		}
	}
	
	return queries
}