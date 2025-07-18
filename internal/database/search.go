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

// SearchOptions holds options for search behavior
type SearchOptions struct {
	Limit         int
	ContextBoosts map[string]float64
	PipelineOnly  bool    // Focus only on pipeline commands
	PipelineBoost float64 // Boost factor for pipeline commands
}

// Search performs a basic keyword-based search
func (db *Database) Search(query string, limit int) []SearchResult {
	return db.SearchWithOptions(query, SearchOptions{
		Limit: limit,
	})
}

// SearchWithOptions performs search with advanced options including context awareness
func (db *Database) SearchWithOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = 5 // default limit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	var results []SearchResult

	for i := range db.Commands {
		score := calculateScore(&db.Commands[i], queryWords, options.ContextBoosts)
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
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

	return results
}

// SearchWithPipelineOptions performs search with pipeline-specific enhancements
func (db *Database) SearchWithPipelineOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = 5 // default limit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	var results []SearchResult

	for i := range db.Commands {
		cmd := &db.Commands[i]

		// If PipelineOnly is true, skip non-pipeline commands
		if options.PipelineOnly && !cmd.Pipeline && !isPipelineCommand(cmd.Command) {
			continue
		}

		score := calculateScore(cmd, queryWords, options.ContextBoosts)

		// Apply pipeline boost
		if (cmd.Pipeline || isPipelineCommand(cmd.Command)) && options.PipelineBoost > 0 {
			score *= options.PipelineBoost
		}

		if score > 0 {
			results = append(results, SearchResult{
				Command: cmd,
				Score:   score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top results
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

	return results
}

// isPipelineCommand checks if a command is likely a pipeline
func isPipelineCommand(command string) bool {
	return strings.Contains(command, "|") ||
		strings.Contains(strings.ToLower(command), "pipe") ||
		strings.Contains(command, "&&") ||
		strings.Contains(command, ">>")
}

// calculateScore computes relevance score for a command based on query words and context
func calculateScore(cmd *Command, queryWords []string, contextBoosts map[string]float64) float64 {
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

		wordScore := 0.0

		// Exact match in command (highest weight)
		if strings.Contains(cmdLower, word) {
			wordScore += 10.0
		}

		// Exact match in description (high weight)
		if strings.Contains(descLower, word) {
			wordScore += 5.0
		}

		// Exact match in keywords (medium weight)
		for _, keyword := range keywordsLower {
			if keyword == word {
				wordScore += 3.0
				break
			}
		}

		// Partial match in keywords (low weight)
		if wordScore == 0 { // Only if no exact match found
			for _, keyword := range keywordsLower {
				if strings.Contains(keyword, word) {
					wordScore += 1.0
					break
				}
			}
		}

		// Apply context boost if available
		if contextBoosts != nil {
			if boost, exists := contextBoosts[word]; exists {
				wordScore *= boost
			}
		}

		score += wordScore
	}

	// Apply niche-based context boost
	if contextBoosts != nil && cmd.Niche != "" {
		nicheLower := strings.ToLower(cmd.Niche)
		if boost, exists := contextBoosts[nicheLower]; exists {
			score *= (1.0 + boost*0.2) // Moderate boost for niche match
		}
	}

	return score
}
