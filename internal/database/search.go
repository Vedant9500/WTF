package database

import (
	"sort"
	"strings"
)

// SearchResult represents a command with its relevance score
type SearchResult struct {
	Command *Command
	Score   float64
}

// Search performs a basic keyword-based search
func (db *Database) Search(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 5 // default limit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	var results []SearchResult

	for i := range db.Commands {
		score := calculateScore(&db.Commands[i], queryWords)
		if score > 0 {
			results = append(results, SearchResult{
				Command: &db.Commands[i],
				Score:   score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top results
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// calculateScore computes relevance score for a command based on query words
func calculateScore(cmd *Command, queryWords []string) float64 {
	var score float64

	// Convert command text to lowercase for matching
	cmdLower := strings.ToLower(cmd.Command)
	descLower := strings.ToLower(cmd.Description)
	
	// Convert keywords to lowercase
	var keywordsLower []string
	for _, keyword := range cmd.Keywords {
		keywordsLower = append(keywordsLower, strings.ToLower(keyword))
	}

	for _, word := range queryWords {
		// Skip very short words
		if len(word) < 2 {
			continue
		}

		// Exact match in command (highest weight)
		if strings.Contains(cmdLower, word) {
			score += 10.0
		}

		// Exact match in description (high weight)
		if strings.Contains(descLower, word) {
			score += 5.0
		}

		// Exact match in keywords (medium weight)
		for _, keyword := range keywordsLower {
			if keyword == word {
				score += 3.0
				break
			}
		}

		// Partial match in keywords (low weight)
		for _, keyword := range keywordsLower {
			if strings.Contains(keyword, word) {
				score += 1.0
				break
			}
		}
	}

	return score
}