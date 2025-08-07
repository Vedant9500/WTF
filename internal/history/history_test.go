package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSearchHistory(t *testing.T) {
	filePath := "/tmp/test_history.json"
	maxSize := 50

	history := NewSearchHistory(filePath, maxSize)

	if history == nil {
		t.Fatal("NewSearchHistory returned nil")
	}

	if history.FilePath != filePath {
		t.Errorf("Expected FilePath '%s', got '%s'", filePath, history.FilePath)
	}

	if history.MaxSize != maxSize {
		t.Errorf("Expected MaxSize %d, got %d", maxSize, history.MaxSize)
	}

	if len(history.Entries) != 0 {
		t.Error("Expected empty entries slice")
	}
}

func TestNewSearchHistoryWithZeroMaxSize(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 0)

	if history.MaxSize != 100 {
		t.Errorf("Expected default MaxSize 100, got %d", history.MaxSize)
	}
}

func TestDefaultHistoryPath(t *testing.T) {
	path := DefaultHistoryPath()

	if path == "" {
		t.Error("DefaultHistoryPath returned empty string")
	}

	// Should contain wtf directory
	if !filepath.IsAbs(path) {
		t.Error("DefaultHistoryPath should return absolute path")
	}

	// Should end with search_history.json
	if filepath.Base(path) != "search_history.json" {
		t.Errorf("Expected filename 'search_history.json', got '%s'", filepath.Base(path))
	}
}

func TestAddEntry(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	query := "git commit"
	resultsCount := 5
	context := "project1"
	duration := 50 * time.Millisecond

	history.AddEntry(query, resultsCount, context, duration)

	if len(history.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(history.Entries))
	}

	entry := history.Entries[0]
	if entry.Query != query {
		t.Errorf("Expected Query '%s', got '%s'", query, entry.Query)
	}

	if entry.ResultsCount != resultsCount {
		t.Errorf("Expected ResultsCount %d, got %d", resultsCount, entry.ResultsCount)
	}

	if entry.Context != context {
		t.Errorf("Expected Context '%s', got '%s'", context, entry.Context)
	}

	if entry.Duration != duration.Milliseconds() {
		t.Errorf("Expected Duration %d, got %d", duration.Milliseconds(), entry.Duration)
	}

	// Check timestamp is recent
	if time.Since(entry.Timestamp) > time.Second {
		t.Error("Expected recent timestamp")
	}
}

func TestAddEntryDuplicate(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	query := "git commit"

	// Add same query twice
	history.AddEntry(query, 3, "ctx1", 30*time.Millisecond)
	history.AddEntry(query, 5, "ctx2", 50*time.Millisecond)

	// Should only have one entry (updated)
	if len(history.Entries) != 1 {
		t.Errorf("Expected 1 entry after duplicate, got %d", len(history.Entries))
	}

	entry := history.Entries[0]
	if entry.ResultsCount != 5 {
		t.Errorf("Expected updated ResultsCount 5, got %d", entry.ResultsCount)
	}

	if entry.Context != "ctx2" {
		t.Errorf("Expected updated Context 'ctx2', got '%s'", entry.Context)
	}
}

func TestAddEntryMaxSize(t *testing.T) {
	maxSize := 3
	history := NewSearchHistory("/tmp/test.json", maxSize)

	// Add more entries than max size
	for i := 0; i < 5; i++ {
		query := fmt.Sprintf("query%d", i)
		history.AddEntry(query, i, "", time.Duration(i)*time.Millisecond)
	}

	// Should only keep the last maxSize entries
	if len(history.Entries) != maxSize {
		t.Errorf("Expected %d entries, got %d", maxSize, len(history.Entries))
	}

	// Should have the last 3 entries (query2, query3, query4)
	expectedQueries := []string{"query2", "query3", "query4"}
	for i, entry := range history.Entries {
		if entry.Query != expectedQueries[i] {
			t.Errorf("Expected entry %d to be '%s', got '%s'", i, expectedQueries[i], entry.Query)
		}
	}
}

func TestGetRecentQueries(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	// Add some entries
	queries := []string{"git commit", "find files", "git push", "find files", "tar compress"}
	for _, query := range queries {
		history.AddEntry(query, 1, "", time.Millisecond)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Get recent queries (should be unique and in reverse order)
	recent := history.GetRecentQueries(3)

	expected := []string{"tar compress", "find files", "git push"}
	if len(recent) != len(expected) {
		t.Errorf("Expected %d recent queries, got %d", len(expected), len(recent))
	}

	for i, query := range recent {
		if query != expected[i] {
			t.Errorf("Expected recent query %d to be '%s', got '%s'", i, expected[i], query)
		}
	}
}

func TestGetRecentQueriesWithLimit(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	// Add entries
	for i := 0; i < 5; i++ {
		history.AddEntry(fmt.Sprintf("query%d", i), 1, "", time.Millisecond)
	}

	// Test with limit 0 (should default to 10)
	recent := history.GetRecentQueries(0)
	if len(recent) != 5 { // All 5 entries
		t.Errorf("Expected 5 recent queries with limit 0, got %d", len(recent))
	}

	// Test with specific limit
	recent = history.GetRecentQueries(2)
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent queries, got %d", len(recent))
	}
}

func TestGetEntriesByPattern(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	// Add entries
	queries := []string{"git commit", "git push", "find files", "grep pattern", "git log"}
	for _, query := range queries {
		history.AddEntry(query, 1, "", time.Millisecond)
		time.Sleep(time.Millisecond)
	}

	// Search for "git" pattern
	matches := history.GetEntriesByPattern("git")

	expectedCount := 3 // git commit, git push, git log
	if len(matches) != expectedCount {
		t.Errorf("Expected %d matches for 'git', got %d", expectedCount, len(matches))
	}

	// Should be sorted by timestamp (most recent first)
	expectedOrder := []string{"git log", "git push", "git commit"}
	for i, entry := range matches {
		if entry.Query != expectedOrder[i] {
			t.Errorf("Expected match %d to be '%s', got '%s'", i, expectedOrder[i], entry.Query)
		}
	}
}

func TestGetTopQueries(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	// Add entries with different frequencies
	queries := []string{"git commit", "find files", "git commit", "git push", "git commit", "find files"}
	for _, query := range queries {
		history.AddEntry(query, 1, "", time.Millisecond)
		time.Sleep(time.Millisecond)
	}

	topQueries := history.GetTopQueries(3)

	// git commit should be first (3 times), then find files (2 times), then git push (1 time)
	expected := []QueryFrequency{
		{Query: "git commit", Count: 3},
		{Query: "find files", Count: 2},
		{Query: "git push", Count: 1},
	}

	if len(topQueries) != len(expected) {
		t.Errorf("Expected %d top queries, got %d", len(expected), len(topQueries))
	}

	for i, qf := range topQueries {
		if qf.Query != expected[i].Query {
			t.Errorf("Expected top query %d to be '%s', got '%s'", i, expected[i].Query, qf.Query)
		}
		if qf.Count != expected[i].Count {
			t.Errorf("Expected count %d for query '%s', got %d", expected[i].Count, qf.Query, qf.Count)
		}
	}
}

func TestGetStats(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 10)

	if len(history.Entries) == 0 {
		stats := history.GetStats()
		if stats.TotalSearches != 0 {
			t.Error("Expected empty stats for empty history")
		}
		return
	}

	// Add some entries
	queries := []string{"git commit", "find files", "git commit"}
	totalResults := 0
	totalDuration := int64(0)

	for i, query := range queries {
		results := i + 1
		duration := time.Duration(i+1) * 10 * time.Millisecond
		history.AddEntry(query, results, "", duration)
		totalResults += results
		totalDuration += duration.Milliseconds()
		time.Sleep(time.Millisecond)
	}

	stats := history.GetStats()

	expectedTotalSearches := 2 // git commit appears twice but should be deduplicated in final count
	if stats.TotalSearches != expectedTotalSearches {
		t.Errorf("Expected TotalSearches %d, got %d", expectedTotalSearches, stats.TotalSearches)
	}

	expectedUniqueQueries := 2 // "git commit" and "find files"
	if stats.UniqueQueries != expectedUniqueQueries {
		t.Errorf("Expected UniqueQueries %d, got %d", expectedUniqueQueries, stats.UniqueQueries)
	}

	// Check that timestamps are set
	if stats.OldestEntry.IsZero() {
		t.Error("Expected OldestEntry to be set")
	}

	if stats.NewestEntry.IsZero() {
		t.Error("Expected NewestEntry to be set")
	}

	if !stats.NewestEntry.After(stats.OldestEntry) && !stats.NewestEntry.Equal(stats.OldestEntry) {
		t.Error("Expected NewestEntry to be after or equal to OldestEntry")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_history.json")

	// Create history and add entries
	history := NewSearchHistory(filePath, 10)
	history.AddEntry("git commit", 5, "project1", 50*time.Millisecond)
	history.AddEntry("find files", 3, "project2", 30*time.Millisecond)

	// Save to file
	err := history.Save()
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Load into new history instance
	newHistory := NewSearchHistory(filePath, 10)
	err = newHistory.Load()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	// Verify loaded data
	if len(newHistory.Entries) != 2 {
		t.Errorf("Expected 2 loaded entries, got %d", len(newHistory.Entries))
	}

	if newHistory.MaxSize != 10 {
		t.Errorf("Expected MaxSize 10, got %d", newHistory.MaxSize)
	}

	// Verify first entry
	entry := newHistory.Entries[0]
	if entry.Query != "git commit" {
		t.Errorf("Expected first entry query 'git commit', got '%s'", entry.Query)
	}
	if entry.ResultsCount != 5 {
		t.Errorf("Expected first entry results 5, got %d", entry.ResultsCount)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	history := NewSearchHistory("/nonexistent/path/history.json", 10)

	// Should not error when file doesn't exist
	err := history.Load()
	if err != nil {
		t.Errorf("Expected no error loading nonexistent file, got: %v", err)
	}

	// Should have empty entries
	if len(history.Entries) != 0 {
		t.Errorf("Expected empty entries after loading nonexistent file, got %d", len(history.Entries))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	// Create empty file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty_history.json")

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	file.Close()

	history := NewSearchHistory(filePath, 10)
	err = history.Load()
	if err != nil {
		t.Errorf("Expected no error loading empty file, got: %v", err)
	}

	if len(history.Entries) != 0 {
		t.Errorf("Expected empty entries after loading empty file, got %d", len(history.Entries))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Create file with invalid JSON
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid_history.json")

	err := os.WriteFile(filePath, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	history := NewSearchHistory(filePath, 10)
	err = history.Load()
	if err == nil {
		t.Error("Expected error loading invalid JSON file")
	}
}

func TestClear(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "clear_test.json")

	history := NewSearchHistory(filePath, 10)
	history.AddEntry("test query", 1, "", time.Millisecond)

	if len(history.Entries) != 1 {
		t.Error("Expected entry before clear")
	}

	err := history.Clear()
	if err != nil {
		t.Errorf("Failed to clear history: %v", err)
	}

	if len(history.Entries) != 0 {
		t.Errorf("Expected empty entries after clear, got %d", len(history.Entries))
	}

	// Verify file is updated
	newHistory := NewSearchHistory(filePath, 10)
	err = newHistory.Load()
	if err != nil {
		t.Errorf("Failed to load after clear: %v", err)
	}

	if len(newHistory.Entries) != 0 {
		t.Errorf("Expected empty entries in file after clear, got %d", len(newHistory.Entries))
	}
}

func TestJSONSerialization(t *testing.T) {
	history := NewSearchHistory("/tmp/test.json", 5)
	history.AddEntry("test query", 3, "context", 100*time.Millisecond)

	// Test JSON marshaling
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal history: %v", err)
	}

	// Test JSON unmarshaling
	var newHistory SearchHistory
	err = json.Unmarshal(data, &newHistory)
	if err != nil {
		t.Fatalf("Failed to unmarshal history: %v", err)
	}

	// Verify data
	if len(newHistory.Entries) != 1 {
		t.Errorf("Expected 1 entry after unmarshal, got %d", len(newHistory.Entries))
	}

	if newHistory.MaxSize != 5 {
		t.Errorf("Expected MaxSize 5 after unmarshal, got %d", newHistory.MaxSize)
	}

	// FilePath should not be serialized
	if newHistory.FilePath != "" {
		t.Errorf("Expected empty FilePath after unmarshal, got '%s'", newHistory.FilePath)
	}
}
