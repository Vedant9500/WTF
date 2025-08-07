package database

import (
	"sort"
	"strings"
	"sync"

	"github.com/Vedant9500/WTF/internal/constants"
)

// SearchResultPool is an object pool for SearchResult slices to reduce allocations
var searchResultPool = sync.Pool{
	New: func() interface{} {
		return make([]SearchResult, 0, constants.DefaultSearchLimit*constants.ResultsBufferMultiplier)
	},
}

// StringSlicePool is an object pool for string slices to reduce allocations
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 10) // Typical query has ~3-5 words
	},
}

// getSearchResults gets a SearchResult slice from the pool
func getSearchResults() []SearchResult {
	return searchResultPool.Get().([]SearchResult)
}

// putSearchResults returns a SearchResult slice to the pool
func putSearchResults(results []SearchResult) {
	if cap(results) > constants.DefaultSearchLimit*constants.ResultsBufferMultiplier*2 {
		// Don't return overly large slices to the pool
		return
	}
	results = results[:0] // Reset length but keep capacity
	searchResultPool.Put(results)
}

// getStringSlice gets a string slice from the pool
func getStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// putStringSlice returns a string slice to the pool
func putStringSlice(slice []string) {
	if cap(slice) > 20 {
		// Don't return overly large slices to the pool
		return
	}
	slice = slice[:0] // Reset length but keep capacity
	stringSlicePool.Put(slice)
}

// OptimizedSearch performs search with memory and CPU optimizations
func (db *Database) OptimizedSearch(query string, limit int) []SearchResult {
	return db.OptimizedSearchWithOptions(query, SearchOptions{
		Limit: limit,
	})
}

// OptimizedSearchWithOptions performs optimized search with advanced options
func (db *Database) OptimizedSearchWithOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = constants.DefaultSearchLimit
	}

	// Optimized query parsing - avoid repeated string operations
	queryLower := strings.ToLower(query)
	queryWords := make([]string, 0, 8) // Most queries have < 8 words
	queryWords = parseQueryWords(queryLower, queryWords)

	if len(queryWords) == 0 {
		return nil
	}

	// Pre-allocate results slice with reasonable capacity
	expectedResults := min(len(db.Commands)/5, options.Limit*constants.ResultsBufferMultiplier)
	results := make([]SearchResult, 0, expectedResults)

	currentPlatform := getCurrentPlatform()

	// Optimized search loop with early termination for performance
	for i := range db.Commands {
		cmd := &db.Commands[i]

		// Quick platform filter first (cheapest check)
		if !db.isPlatformMatch(cmd, currentPlatform) {
			continue
		}

		// Calculate score using optimized algorithm
		if score := db.calculateOptimizedScore(cmd, queryWords, options.ContextBoosts); score > 0 {
			results = append(results, SearchResult{
				Command: cmd,
				Score:   score,
			})
		}
	}

	// Sort and limit results
	return db.sortAndLimitResultsOptimized(results, options.Limit)
}

// parseQueryWords optimally parses query into words, reusing provided slice
func parseQueryWords(queryLower string, words []string) []string {
	if len(queryLower) == 0 {
		return words[:0]
	}

	// Manual parsing to avoid strings.Fields allocation
	start := 0
	inWord := false

	for i, r := range queryLower {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if inWord {
				word := queryLower[start:i]
				if len(word) >= constants.MinWordLength {
					words = append(words, word)
				}
				inWord = false
			}
		} else {
			if !inWord {
				start = i
				inWord = true
			}
		}
	}

	// Handle last word
	if inWord {
		word := queryLower[start:]
		if len(word) >= constants.MinWordLength {
			words = append(words, word)
		}
	}

	return words
}

// isPlatformMatch performs optimized platform matching
func (db *Database) isPlatformMatch(cmd *Command, currentPlatform string) bool {
	if len(cmd.Platform) == 0 {
		return true // No platform restriction
	}

	// Check for cross-platform or current platform match
	for _, p := range cmd.Platform {
		if len(p) == 0 {
			continue
		}

		// Fast case-insensitive comparison using first character
		firstChar := p[0] | 0x20 // Convert to lowercase
		if firstChar == 'c' && strings.EqualFold(p, "cross-platform") {
			return true
		}

		// Check current platform match
		if strings.EqualFold(p, currentPlatform) {
			return true
		}
	}

	// Check legacy cross-platform tools
	return isCrossPlatformTool(cmd.Command)
}

// calculateOptimizedScore computes relevance score with optimizations
func (db *Database) calculateOptimizedScore(cmd *Command, queryWords []string, contextBoosts map[string]float64) float64 {
	var score float64
	var maxWordScore float64
	matchedWords := 0

	// Pre-compute command fields for efficiency
	cmdLower := cmd.CommandLower
	descLower := cmd.DescriptionLower
	keywordsLower := cmd.KeywordsLower
	tagsLower := cmd.TagsLower

	for _, word := range queryWords {
		wordScore := db.calculateOptimizedWordScore(word, cmdLower, descLower, keywordsLower, tagsLower, cmd)

		// Track the highest scoring word
		if wordScore > maxWordScore {
			maxWordScore = wordScore
		}

		// Count words that have some match
		if wordScore > 0 {
			matchedWords++
		}

		// Apply context boost if available
		if contextBoosts != nil {
			if boost, exists := contextBoosts[word]; exists {
				wordScore *= boost
			}
		}

		score += wordScore
	}

	// Apply completeness bonus
	if len(queryWords) > 1 && matchedWords > 1 {
		completenessBonus := float64(matchedWords) / float64(len(queryWords))
		score *= (1.0 + completenessBonus*0.5)
	}

	// Apply score boosts based on match quality
	if maxWordScore >= constants.DirectCommandMatchScore {
		score *= 1.8
	} else if maxWordScore >= constants.CommandMatchScore {
		score *= 1.4
	}

	// Apply category-based relevance boost
	score *= getCategoryRelevanceBoost(cmd, queryWords)

	// Apply niche-based context boost
	if contextBoosts != nil && cmd.Niche != "" {
		nicheLower := strings.ToLower(cmd.Niche)
		if boost, exists := contextBoosts[nicheLower]; exists {
			score *= (1.0 + boost*constants.NicheBoostFactor)
		}
	}

	return score
}

// calculateOptimizedWordScore computes word score with optimized string operations
func (db *Database) calculateOptimizedWordScore(word, cmdLower, descLower string, keywordsLower, tagsLower []string, cmd *Command) float64 {
	var wordScore float64
	wordLen := len(word)

	// HIGHEST PRIORITY: Command matching with optimized string operations
	if len(cmdLower) == wordLen && cmdLower == word {
		wordScore += constants.DirectCommandMatchScore * 2.0
	} else if len(cmdLower) > wordLen {
		// Fast prefix check
		if cmdLower[:wordLen] == word {
			// Check if it's a word boundary
			if cmdLower[wordLen] == ' ' {
				wordScore += constants.DirectCommandMatchScore * 1.5
			} else {
				wordScore += constants.CommandMatchScore * 0.7
			}
		} else {
			// Use fast contains check for word boundaries
			if idx := strings.Index(cmdLower, word); idx >= 0 {
				// Check word boundaries
				prevOK := idx == 0 || cmdLower[idx-1] == ' '
				nextOK := idx+wordLen >= len(cmdLower) || cmdLower[idx+wordLen] == ' '

				if prevOK && nextOK {
					wordScore += constants.CommandMatchScore
				} else if strings.Contains(cmdLower, word) {
					wordScore += constants.CommandMatchScore * 0.7
				}
			}
		}
	}

	// HIGH PRIORITY: Domain-specific matching (only if no command match yet)
	if wordScore < constants.CommandMatchScore && isDomainSpecificMatch(word, cmd) {
		wordScore += constants.DomainSpecificScore
	}

	// MEDIUM-HIGH PRIORITY: Keywords matching
	exactKeywordMatch := false
	for _, keyword := range keywordsLower {
		if len(keyword) == wordLen && keyword == word {
			wordScore += constants.KeywordExactScore * 1.5
			exactKeywordMatch = true
			break
		}
	}

	// Partial keyword match only if no exact match
	if !exactKeywordMatch {
		for _, keyword := range keywordsLower {
			if strings.Contains(keyword, word) {
				wordScore += constants.KeywordPartialScore
				break
			}
		}
	}

	// MEDIUM-HIGH PRIORITY: Description matching
	if idx := strings.Index(descLower, word); idx >= 0 {
		// Check word boundaries
		prevOK := idx == 0 || descLower[idx-1] == ' '
		nextOK := idx+wordLen >= len(descLower) || descLower[idx+wordLen] == ' '

		if prevOK && nextOK {
			wordScore += constants.DescriptionMatchScore
		} else {
			wordScore += constants.DescriptionMatchScore * 0.6
		}
	}

	// MEDIUM-HIGH PRIORITY: Tags matching
	for _, tag := range tagsLower {
		if len(tag) == wordLen && tag == word {
			wordScore += constants.TagExactScore
			return wordScore // Early return for exact tag match
		}
	}

	// Partial tag matching
	for _, tag := range tagsLower {
		if strings.Contains(tag, word) {
			wordScore += constants.TagPartialScore
			break
		}
	}

	return wordScore
}

// containsWord checks if a word exists as a complete word in text (optimized)
func (db *Database) containsWord(text, word string) bool {
	wordLen := len(word)
	textLen := len(text)

	if wordLen > textLen {
		return false
	}

	// Look for word boundaries
	for i := 0; i <= textLen-wordLen; i++ {
		// Check if we found the word
		if text[i:i+wordLen] == word {
			// Check word boundaries
			prevOK := i == 0 || text[i-1] == ' '
			nextOK := i+wordLen == textLen || text[i+wordLen] == ' '

			if prevOK && nextOK {
				return true
			}
		}
	}

	return false
}

// sortAndLimitResultsOptimized sorts and limits results with optimizations
func (db *Database) sortAndLimitResultsOptimized(results []SearchResult, limit int) []SearchResult {
	if len(results) == 0 {
		return nil
	}

	// Use partial sort if we have many more results than needed
	if len(results) > limit*3 {
		// Use partial sort for better performance with large result sets
		db.partialSort(results, limit)
		if len(results) > limit {
			results = results[:limit]
		}
	} else {
		// Use regular sort for smaller result sets
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		if len(results) > limit {
			results = results[:limit]
		}
	}

	// Create a new slice to return (not from pool) since caller will keep it
	finalResults := make([]SearchResult, len(results))
	copy(finalResults, results)

	return finalResults
}

// partialSort performs partial sorting to get top N results efficiently
func (db *Database) partialSort(results []SearchResult, n int) {
	if n >= len(results) {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		return
	}

	// Use selection sort for small n, quickselect for larger n
	if n <= 10 {
		// Selection sort for top n elements
		for i := 0; i < n; i++ {
			maxIdx := i
			for j := i + 1; j < len(results); j++ {
				if results[j].Score > results[maxIdx].Score {
					maxIdx = j
				}
			}
			if maxIdx != i {
				results[i], results[maxIdx] = results[maxIdx], results[i]
			}
		}
	} else {
		// For larger n, use partial sort
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
}

// BatchOptimizedSearch performs multiple searches efficiently using shared resources
func (db *Database) BatchOptimizedSearch(queries []string, limit int) [][]SearchResult {
	if len(queries) == 0 {
		return nil
	}

	results := make([][]SearchResult, len(queries))

	// Reuse query words slice across all searches
	queryWords := getStringSlice()
	defer putStringSlice(queryWords)

	// Reuse search results slice across all searches
	searchResults := getSearchResults()
	defer putSearchResults(searchResults)

	currentPlatform := getCurrentPlatform()

	for i, query := range queries {
		// Reset slices for reuse
		queryWords = queryWords[:0]
		searchResults = searchResults[:0]

		// Parse query
		queryLower := strings.ToLower(query)
		queryWords = parseQueryWords(queryLower, queryWords)

		if len(queryWords) == 0 {
			results[i] = nil
			continue
		}

		// Perform search
		for j := range db.Commands {
			cmd := &db.Commands[j]

			if !db.isPlatformMatch(cmd, currentPlatform) {
				continue
			}

			if score := db.calculateOptimizedScore(cmd, queryWords, nil); score > 0 {
				searchResults = append(searchResults, SearchResult{
					Command: cmd,
					Score:   score,
				})
			}
		}

		// Sort and limit
		results[i] = db.sortAndLimitResultsOptimized(searchResults, limit)
	}

	return results
}
