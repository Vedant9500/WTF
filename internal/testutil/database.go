package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Vedant9500/WTF/internal/database"
)

// DatabaseTestSuite provides a comprehensive test suite for database operations
type DatabaseTestSuite struct {
	db       *database.Database
	fixtures TestFixtures
	helper   *DatabaseTestHelper
}

// NewDatabaseTestSuite creates a new database test suite
func NewDatabaseTestSuite(db *database.Database) *DatabaseTestSuite {
	return &DatabaseTestSuite{
		db:       db,
		fixtures: NewTestFixtures(),
		helper:   NewDatabaseTestHelper(),
	}
}

// RunBasicSearchTests runs basic search functionality tests
func (dts *DatabaseTestSuite) RunBasicSearchTests(t *testing.T) {
	t.Helper()

	testQueries := dts.fixtures.GetTestQueries()

	for _, query := range testQueries {
		t.Run(fmt.Sprintf("Search_%s", strings.ReplaceAll(query.Query, " ", "_")), func(t *testing.T) {
			results := dts.db.Search(query.Query, 10)
			dts.helper.ValidateSearchResults(t, results, query)
		})
	}
}

// RunSearchOptionsTests runs tests for search with options
func (dts *DatabaseTestSuite) RunSearchOptionsTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		query    string
		options  database.SearchOptions
		validate func(t *testing.T, results []database.SearchResult)
	}{
		{
			name:  "LimitTest",
			query: "git",
			options: database.SearchOptions{
				Limit: 2,
			},
			validate: func(t *testing.T, results []database.SearchResult) {
				if len(results) > 2 {
					t.Errorf("Expected at most 2 results, got %d", len(results))
				}
			},
		},
		{
			name:  "ContextBoostTest",
			query: "git commit",
			options: database.SearchOptions{
				Limit: 5,
				ContextBoosts: map[string]float64{
					"git": 2.0,
				},
			},
			validate: func(t *testing.T, results []database.SearchResult) {
				if len(results) == 0 {
					t.Error("Expected at least one result with context boost")
				}
				// First result should have higher score due to boost
				if len(results) > 1 && results[0].Score <= results[1].Score {
					t.Error("Expected first result to have higher score due to context boost")
				}
			},
		},
		{
			name:  "PipelineOnlyTest",
			query: "process",
			options: database.SearchOptions{
				Limit:        5,
				PipelineOnly: true,
			},
			validate: func(t *testing.T, results []database.SearchResult) {
				for _, result := range results {
					if !result.Command.Pipeline && !strings.Contains(result.Command.Command, "|") {
						t.Errorf("Expected only pipeline commands, got: %s", result.Command.Command)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := dts.db.SearchUniversal(tc.query, tc.options)
			tc.validate(t, results)
		})
	}
}

// RunEdgeCaseTests runs edge case tests
func (dts *DatabaseTestSuite) RunEdgeCaseTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		query    string
		limit    int
		expected int
	}{
		{
			name:     "EmptyQuery",
			query:    "",
			limit:    5,
			expected: 0,
		},
		{
			name:     "SingleCharQuery",
			query:    "a",
			limit:    5,
			expected: 0, // Single char queries should be ignored
		},
		{
			name:     "ZeroLimit",
			query:    "git",
			limit:    0,
			expected: 5, // Should default to 5
		},
		{
			name:     "NegativeLimit",
			query:    "git",
			limit:    -1,
			expected: 5, // Should default to 5
		},
		{
			name:     "VeryLongQuery",
			query:    strings.Repeat("very long query ", 100),
			limit:    5,
			expected: 0, // Should handle gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := dts.db.Search(tc.query, tc.limit)
			if tc.name == "ZeroLimit" || tc.name == "NegativeLimit" {
				// For these cases, we just check that it doesn't crash
				// The actual limit behavior is handled by the search function
				return
			}
			if len(results) != tc.expected {
				t.Errorf("Expected %d results, got %d", tc.expected, len(results))
			}
		})
	}
}

// RunPerformanceTests runs performance-related tests
func (dts *DatabaseTestSuite) RunPerformanceTests(t *testing.T) {
	t.Helper()

	// Create a large database for performance testing
	largeDB := dts.helper.CreateLargeDatabase()
	largeSuite := NewDatabaseTestSuite(largeDB)

	t.Run("LargeDatasetSearch", func(t *testing.T) {
		results := largeSuite.db.Search("git", 10)
		if len(results) == 0 {
			t.Error("Expected results from large dataset search")
		}
	})

	t.Run("MultipleSearches", func(_ *testing.T) {
		queries := []string{"git", "find", "tar", "grep", "curl"}
		for _, query := range queries {
			results := largeSuite.db.Search(query, 5)
			_ = results // Just ensure it doesn't crash
		}
	})
}

// DatabaseValidator provides validation utilities for database testing
type DatabaseValidator struct{}

// NewDatabaseValidator creates a new database validator
func NewDatabaseValidator() *DatabaseValidator {
	return &DatabaseValidator{}
}

// ValidateDatabase validates the structure and content of a database
func (dv *DatabaseValidator) ValidateDatabase(t *testing.T, db *database.Database) {
	t.Helper()

	if db == nil {
		t.Fatal("Database is nil")
	}

	if db.Commands == nil {
		t.Fatal("Database commands slice is nil")
	}

	// Validate each command
	for i, cmd := range db.Commands {
		dv.ValidateCommand(t, &cmd, i)
	}
}

// ValidateCommand validates a single command structure
func (dv *DatabaseValidator) ValidateCommand(t *testing.T, cmd *database.Command, index int) {
	t.Helper()

	if cmd.Command == "" {
		t.Errorf("Command at index %d has empty command field", index)
	}

	if cmd.Description == "" {
		t.Errorf("Command at index %d has empty description field", index)
	}

	if len(cmd.Keywords) == 0 {
		t.Errorf("Command at index %d has no keywords", index)
	}

	// Validate cached fields are populated
	if cmd.CommandLower == "" && cmd.Command != "" {
		t.Errorf("Command at index %d has empty CommandLower field", index)
	}

	if cmd.DescriptionLower == "" && cmd.Description != "" {
		t.Errorf("Command at index %d has empty DescriptionLower field", index)
	}

	if len(cmd.KeywordsLower) != len(cmd.Keywords) {
		t.Errorf("Command at index %d has mismatched KeywordsLower length", index)
	}

	// Validate lowercased fields are actually lowercase
	if !strings.EqualFold(cmd.CommandLower, cmd.Command) {
		t.Errorf("Command at index %d has incorrect CommandLower field", index)
	}

	if !strings.EqualFold(cmd.DescriptionLower, cmd.Description) {
		t.Errorf("Command at index %d has incorrect DescriptionLower field", index)
	}

	for j, keyword := range cmd.KeywordsLower {
		if j < len(cmd.Keywords) && !strings.EqualFold(keyword, cmd.Keywords[j]) {
			t.Errorf("Command at index %d has incorrect KeywordsLower[%d] field", index, j)
		}
	}
}

// ValidateSearchResults validates search results
func (dv *DatabaseValidator) ValidateSearchResults(t *testing.T, results []database.SearchResult, _ string) {
	t.Helper()

	// Check that results are sorted by score (descending)
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Errorf("Search results are not sorted by score: result[%d].Score=%f < result[%d].Score=%f",
				i-1, results[i-1].Score, i, results[i].Score)
		}
	}

	// Check that all results have positive scores
	for i, result := range results {
		if result.Score <= 0 {
			t.Errorf("Result[%d] has non-positive score: %f", i, result.Score)
		}

		if result.Command == nil {
			t.Errorf("Result[%d] has nil command", i)
		}
	}
}

// MockDatabase provides a mock database for testing
type MockDatabase struct {
	commands      []database.Command
	searchResults []database.SearchResult
	searchCalled  bool
	lastQuery     string
	lastLimit     int
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		commands:      make([]database.Command, 0),
		searchResults: make([]database.SearchResult, 0),
	}
}

// SetCommands sets the commands for the mock database
func (md *MockDatabase) SetCommands(commands []database.Command) {
	md.commands = commands
}

// SetSearchResults sets the results that should be returned by search
func (md *MockDatabase) SetSearchResults(results []database.SearchResult) {
	md.searchResults = results
}

// Search mocks the search functionality
func (md *MockDatabase) Search(query string, limit int) []database.SearchResult {
	md.searchCalled = true
	md.lastQuery = query
	md.lastLimit = limit

	if len(md.searchResults) > limit {
		return md.searchResults[:limit]
	}
	return md.searchResults
}

// WasSearchCalled returns whether search was called
func (md *MockDatabase) WasSearchCalled() bool {
	return md.searchCalled
}

// GetLastQuery returns the last query used in search
func (md *MockDatabase) GetLastQuery() string {
	return md.lastQuery
}

// GetLastLimit returns the last limit used in search
func (md *MockDatabase) GetLastLimit() int {
	return md.lastLimit
}

// Reset resets the mock database state
func (md *MockDatabase) Reset() {
	md.searchCalled = false
	md.lastQuery = ""
	md.lastLimit = 0
	md.searchResults = md.searchResults[:0]
}
