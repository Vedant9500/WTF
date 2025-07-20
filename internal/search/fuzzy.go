// Package search provides advanced search capabilities for command discovery.
//
// This package implements fuzzy search functionality using the Levenshtein distance
// algorithm to handle typos and approximate matches. It provides:
//   - Command name fuzzy matching
//   - Description fuzzy matching
//   - Combined search across commands and descriptions
//   - Configurable match scoring and limits
package search

import (
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
	"github.com/sahilm/fuzzy"
)

// FuzzySearcher handles fuzzy matching for commands
type FuzzySearcher struct {
	commands []database.Command
}

// NewFuzzySearcher creates a new fuzzy searcher
func NewFuzzySearcher(commands []database.Command) *FuzzySearcher {
	return &FuzzySearcher{
		commands: commands,
	}
}

// FuzzyMatch represents a fuzzy search result
type FuzzyMatch struct {
	Command *database.Command
	Score   int
}

// Search performs fuzzy search on commands and descriptions
func (fs *FuzzySearcher) Search(query string, limit int) []FuzzyMatch {
	if limit <= 0 {
		limit = 5
	}

	// Create search targets combining command and description
	targets := make([]string, len(fs.commands))
	for i, cmd := range fs.commands {
		// Combine command and description for better matching
		targets[i] = cmd.Command + " " + cmd.Description
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Convert to our result format
	var results []FuzzyMatch
	for i, match := range matches {
		if i >= limit {
			break
		}

		results = append(results, FuzzyMatch{
			Command: &fs.commands[match.Index],
			Score:   match.Score,
		})
	}

	return results
}

// SearchCommand performs fuzzy search specifically on command names
func (fs *FuzzySearcher) SearchCommand(query string, limit int) []FuzzyMatch {
	if limit <= 0 {
		limit = 5
	}

	// Create search targets using only command names
	targets := make([]string, len(fs.commands))
	for i, cmd := range fs.commands {
		targets[i] = cmd.Command
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Convert to our result format
	var results []FuzzyMatch
	for i, match := range matches {
		if i >= limit {
			break
		}

		results = append(results, FuzzyMatch{
			Command: &fs.commands[match.Index],
			Score:   match.Score,
		})
	}

	return results
}

// SearchDescription performs fuzzy search specifically on descriptions
func (fs *FuzzySearcher) SearchDescription(query string, limit int) []FuzzyMatch {
	if limit <= 0 {
		limit = 5
	}

	// Create search targets using only descriptions
	targets := make([]string, len(fs.commands))
	for i, cmd := range fs.commands {
		targets[i] = cmd.Description
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Convert to our result format
	var results []FuzzyMatch
	for i, match := range matches {
		if i >= limit {
			break
		}

		results = append(results, FuzzyMatch{
			Command: &fs.commands[match.Index],
			Score:   match.Score,
		})
	}

	return results
}

// SuggestCorrections provides "Did you mean?" suggestions for typos
func (fs *FuzzySearcher) SuggestCorrections(query string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}

	// Extract common words from commands and descriptions
	wordSet := make(map[string]bool)
	for _, cmd := range fs.commands {
		// Split command into words
		cmdWords := strings.Fields(cmd.Command)
		for _, word := range cmdWords {
			if len(word) > 2 { // Ignore very short words
				wordSet[strings.ToLower(word)] = true
			}
		}

		// Split description into words
		descWords := strings.Fields(cmd.Description)
		for _, word := range descWords {
			cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:"))
			if len(cleanWord) > 2 { // Ignore very short words
				wordSet[cleanWord] = true
			}
		}
	}

	// Convert to slice for fuzzy matching
	words := make([]string, 0, len(wordSet))
	for word := range wordSet {
		words = append(words, word)
	}

	// Find fuzzy matches for the query
	matches := fuzzy.Find(query, words)

	var suggestions []string
	for i, match := range matches {
		if i >= maxSuggestions {
			break
		}
		suggestions = append(suggestions, words[match.Index])
	}

	return suggestions
}

// IsTypo determines if a query might be a typo based on fuzzy match scores
func (fs *FuzzySearcher) IsTypo(query string, exactMatches int) bool {
	// If we have exact matches, it's probably not a typo
	if exactMatches > 0 {
		return false
	}

	// Check if fuzzy search returns significantly better results
	fuzzyResults := fs.Search(query, 1)
	if len(fuzzyResults) > 0 && fuzzyResults[0].Score > 0 {
		// If fuzzy search finds something but exact search doesn't,
		// and the fuzzy score is reasonable, it might be a typo
		return fuzzyResults[0].Score >= -10 // Adjust threshold as needed
	}

	return false
}
