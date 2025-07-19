package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SearchEntry represents a single search in the history
type SearchEntry struct {
	Query        string    `json:"query"`
	Timestamp    time.Time `json:"timestamp"`
	ResultsCount int       `json:"results_count"`
	Context      string    `json:"context,omitempty"`  // Project context if any
	Duration     int64     `json:"duration,omitempty"` // Search duration in milliseconds
}

// SearchHistory manages the search history
type SearchHistory struct {
	Entries  []SearchEntry `json:"entries"`
	MaxSize  int           `json:"max_size"`
	FilePath string        `json:"-"` // Don't serialize the file path
}

// NewSearchHistory creates a new search history manager
func NewSearchHistory(filePath string, maxSize int) *SearchHistory {
	if maxSize <= 0 {
		maxSize = 100 // Default max size
	}

	return &SearchHistory{
		Entries:  make([]SearchEntry, 0),
		MaxSize:  maxSize,
		FilePath: filePath,
	}
}

// DefaultHistoryPath returns the default path for search history
func DefaultHistoryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".wtf", "search_history.json")
	}
	return filepath.Join(configDir, "wtf", "search_history.json")
}

// Load loads search history from file
func (sh *SearchHistory) Load() error {
	if _, err := os.Stat(sh.FilePath); os.IsNotExist(err) {
		// File doesn't exist, start with empty history
		return nil
	}

	data, err := os.ReadFile(sh.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	if len(data) == 0 {
		// Empty file, start with empty history
		return nil
	}

	err = json.Unmarshal(data, sh)
	if err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	return nil
}

// Save saves search history to file
func (sh *SearchHistory) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(sh.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	data, err := json.MarshalIndent(sh, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	err = os.WriteFile(sh.FilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// AddEntry adds a new search entry to the history
func (sh *SearchHistory) AddEntry(query string, resultsCount int, context string, duration time.Duration) {
	entry := SearchEntry{
		Query:        query,
		Timestamp:    time.Now(),
		ResultsCount: resultsCount,
		Context:      context,
		Duration:     duration.Milliseconds(),
	}

	// Check if this is a duplicate of the most recent entry
	if len(sh.Entries) > 0 && sh.Entries[len(sh.Entries)-1].Query == query {
		// Update the existing entry instead of adding a duplicate
		sh.Entries[len(sh.Entries)-1] = entry
		return
	}

	sh.Entries = append(sh.Entries, entry)

	// Trim to max size if needed
	if len(sh.Entries) > sh.MaxSize {
		sh.Entries = sh.Entries[len(sh.Entries)-sh.MaxSize:]
	}
}

// GetRecentQueries returns the most recent queries
func (sh *SearchHistory) GetRecentQueries(limit int) []string {
	if limit <= 0 {
		limit = 10
	}

	// Get unique queries in reverse chronological order
	seen := make(map[string]bool)
	var queries []string

	for i := len(sh.Entries) - 1; i >= 0 && len(queries) < limit; i-- {
		query := sh.Entries[i].Query
		if !seen[query] {
			seen[query] = true
			queries = append(queries, query)
		}
	}

	return queries
}

// GetEntriesByPattern returns entries matching a pattern
func (sh *SearchHistory) GetEntriesByPattern(pattern string) []SearchEntry {
	var matches []SearchEntry

	for _, entry := range sh.Entries {
		if containsIgnoreCase(entry.Query, pattern) {
			matches = append(matches, entry)
		}
	}

	// Sort by timestamp (most recent first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Timestamp.After(matches[j].Timestamp)
	})

	return matches
}

// GetTopQueries returns the most frequently searched queries
func (sh *SearchHistory) GetTopQueries(limit int) []QueryFrequency {
	if limit <= 0 {
		limit = 10
	}

	frequency := make(map[string]int)
	lastSeen := make(map[string]time.Time)

	for _, entry := range sh.Entries {
		frequency[entry.Query]++
		if entry.Timestamp.After(lastSeen[entry.Query]) {
			lastSeen[entry.Query] = entry.Timestamp
		}
	}

	var queryFreqs []QueryFrequency
	for query, count := range frequency {
		queryFreqs = append(queryFreqs, QueryFrequency{
			Query:    query,
			Count:    count,
			LastUsed: lastSeen[query],
		})
	}

	// Sort by frequency, then by recency
	sort.Slice(queryFreqs, func(i, j int) bool {
		if queryFreqs[i].Count == queryFreqs[j].Count {
			return queryFreqs[i].LastUsed.After(queryFreqs[j].LastUsed)
		}
		return queryFreqs[i].Count > queryFreqs[j].Count
	})

	if len(queryFreqs) > limit {
		queryFreqs = queryFreqs[:limit]
	}

	return queryFreqs
}

// QueryFrequency represents a query with its usage frequency
type QueryFrequency struct {
	Query    string    `json:"query"`
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

// GetStats returns usage statistics
func (sh *SearchHistory) GetStats() HistoryStats {
	if len(sh.Entries) == 0 {
		return HistoryStats{}
	}

	stats := HistoryStats{
		TotalSearches: len(sh.Entries),
		UniqueQueries: len(sh.getUniqueQueries()),
		OldestEntry:   sh.Entries[0].Timestamp,
		NewestEntry:   sh.Entries[len(sh.Entries)-1].Timestamp,
	}

	// Calculate average results per search
	totalResults := 0
	totalDuration := int64(0)
	for _, entry := range sh.Entries {
		totalResults += entry.ResultsCount
		totalDuration += entry.Duration
	}

	stats.AvgResultsPerSearch = float64(totalResults) / float64(len(sh.Entries))
	if totalDuration > 0 {
		stats.AvgSearchDuration = float64(totalDuration) / float64(len(sh.Entries))
	}

	return stats
}

// HistoryStats represents usage statistics
type HistoryStats struct {
	TotalSearches       int       `json:"total_searches"`
	UniqueQueries       int       `json:"unique_queries"`
	AvgResultsPerSearch float64   `json:"avg_results_per_search"`
	AvgSearchDuration   float64   `json:"avg_search_duration_ms"`
	OldestEntry         time.Time `json:"oldest_entry"`
	NewestEntry         time.Time `json:"newest_entry"`
}

// getUniqueQueries returns a map of unique queries
func (sh *SearchHistory) getUniqueQueries() map[string]bool {
	unique := make(map[string]bool)
	for _, entry := range sh.Entries {
		unique[entry.Query] = true
	}
	return unique
}

// Clear removes all entries from history
func (sh *SearchHistory) Clear() error {
	sh.Entries = make([]SearchEntry, 0)
	return sh.Save()
}

// containsIgnoreCase checks if s contains substr (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
