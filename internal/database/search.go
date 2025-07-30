package database

import (
	"sort"
	"strings"
	"runtime"

	"github.com/Vedant9500/WTF/internal/nlp"
	"github.com/sahilm/fuzzy"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SearchResult represents a command with its relevance score
type SearchResult struct {
	Command *Command
	Score   float64
}

// SearchOptions holds options for search behavior
type SearchOptions struct {
	Limit          int
	ContextBoosts  map[string]float64
	PipelineOnly   bool    // Focus only on pipeline commands
	PipelineBoost  float64 // Boost factor for pipeline commands
	UseFuzzy       bool    // Enable fuzzy search for typos
	FuzzyThreshold int     // Minimum fuzzy score threshold
	UseNLP         bool    // Enable natural language processing
}

// Search performs a basic keyword-based search
func (db *Database) Search(query string, limit int) []SearchResult {
	return db.SearchWithOptions(query, SearchOptions{
		Limit: limit,
	})
}

// SearchWithOptions performs search with advanced options including context awareness and platform filtering
func (db *Database) SearchWithOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = 5 // default limit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	results := make([]SearchResult, 0, min(len(db.Commands), options.Limit*3))

	currentPlatform := getCurrentPlatform()

	for i := range db.Commands {
		cmd := &db.Commands[i]
		// Platform filtering/penalty: if cmd.Platform is set and currentPlatform is not in it, penalize or skip
		if len(cmd.Platform) > 0 {
			found := false
			for _, p := range cmd.Platform {
				if strings.EqualFold(p, currentPlatform) {
					found = true
					break
				}
			}
			if !found {
				// Penalize score heavily (or skip entirely)
				continue // To only show platform-compatible commands, uncomment this line
				// Or, to just penalize:
				// score := calculateScore(cmd, queryWords, options.ContextBoosts) * 0.1
				// if score > 0 { results = append(results, SearchResult{Command: cmd, Score: score}) }
				// continue
			}
		}
		score := calculateScore(cmd, queryWords, options.ContextBoosts)
		if score > 0 {
			results = append(results, SearchResult{
				Command: cmd,
				Score:   score,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

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

	// Convert command text to lowercase for matching (cache these conversions)
	cmdLower := strings.ToLower(cmd.Command)
	descLower := strings.ToLower(cmd.Description)

	// Convert keywords to lowercase once and cache
	keywordsLower := make([]string, len(cmd.Keywords))
	for i, keyword := range cmd.Keywords {
		keywordsLower[i] = strings.ToLower(keyword)
	}

	for _, word := range queryWords {
		// Skip very short words
		if len(word) < 2 {
			continue
		}

		wordScore := 0.0

		// HIGHEST PRIORITY: Exact match in command name
		if strings.Contains(cmdLower, word) {
			// Even higher boost if it's a direct command match
			if strings.HasPrefix(cmdLower, word) || cmdLower == word {
				wordScore += 15.0 // Increased from 10.0
			} else {
				wordScore += 10.0
			}
		}

		// HIGH PRIORITY: Domain-specific command matching
		if isDomainSpecificMatch(word, cmd) {
			wordScore += 12.0 // New: boost for domain-specific relevance
		}

		// MEDIUM-HIGH PRIORITY: Exact match in description
		if strings.Contains(descLower, word) {
			wordScore += 6.0 // Increased from 5.0
		}

		// MEDIUM PRIORITY: Exact match in keywords
		for _, keyword := range keywordsLower {
			if keyword == word {
				wordScore += 4.0 // Increased from 3.0
				break
			}
		}

		// LOW PRIORITY: Partial match in keywords (only if no exact match)
		if wordScore == 0 {
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

	// Apply category-based relevance boost
	score *= getCategoryRelevanceBoost(cmd, queryWords)

	// Apply niche-based context boost
	if contextBoosts != nil && cmd.Niche != "" {
		nicheLower := strings.ToLower(cmd.Niche)
		if boost, exists := contextBoosts[nicheLower]; exists {
			score *= (1.0 + boost*0.2) // Moderate boost for niche match
		}
	}

	return score
}

// isDomainSpecificMatch checks if a query word has high domain relevance to a command
func isDomainSpecificMatch(word string, cmd *Command) bool {
	cmdLower := strings.ToLower(cmd.Command)

	// Define domain-specific mappings for better relevance
	domainMappings := map[string][]string{
		"compress":   {"tar", "gzip", "zip", "bzip", "7z", "compress", "archive"},
		"extract":    {"tar", "unzip", "gunzip", "extract", "unarchive"},
		"directory":  {"mkdir", "rmdir", "ls", "dir", "cd", "pwd"},
		"file":       {"cp", "mv", "rm", "touch", "cat", "less", "more"},
		"search":     {"grep", "find", "locate", "ag", "rg"},
		"download":   {"wget", "curl", "fetch", "download"},
		"git":        {"git", "clone", "commit", "push", "pull", "branch"},
		"package":    {"apt", "yum", "dnf", "pkg", "brew", "pip", "npm"},
		"process":    {"ps", "kill", "top", "htop", "jobs"},
		"network":    {"ping", "ssh", "scp", "rsync", "nc", "nmap"},
		"edit":       {"vim", "nano", "emacs", "edit", "sed", "awk"},
		"permission": {"chmod", "chown", "chgrp", "sudo"},
	}

	// Check if the command belongs to the word's domain
	if commands, exists := domainMappings[word]; exists {
		for _, domainCmd := range commands {
			if strings.Contains(cmdLower, domainCmd) {
				return true
			}
		}
	}

	return false
}

// getCategoryRelevanceBoost applies category-based relevance scoring
func getCategoryRelevanceBoost(cmd *Command, queryWords []string) float64 {
	boost := 1.0

	// Check if query suggests specific command categories
	for _, word := range queryWords {
		switch word {
		case "compress", "archive", "zip", "tar":
			if strings.Contains(strings.ToLower(cmd.Command), "tar") ||
				strings.Contains(strings.ToLower(cmd.Command), "zip") ||
				strings.Contains(strings.ToLower(cmd.Command), "gzip") {
				boost *= 1.5
			}
		case "directory", "folder", "mkdir":
			if strings.Contains(strings.ToLower(cmd.Command), "mkdir") ||
				strings.Contains(strings.ToLower(cmd.Command), "dir") {
				boost *= 1.5
			}
		case "search", "find":
			if strings.Contains(strings.ToLower(cmd.Command), "grep") ||
				strings.Contains(strings.ToLower(cmd.Command), "find") {
				boost *= 1.3
			}
		case "download", "get":
			if strings.Contains(strings.ToLower(cmd.Command), "wget") ||
				strings.Contains(strings.ToLower(cmd.Command), "curl") {
				boost *= 1.4
			}
		}
	}

	return boost
}

// SearchWithFuzzy performs hybrid search combining exact matching and fuzzy search
func (db *Database) SearchWithFuzzy(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = 5
	}

	// First try exact search
	exactResults := db.SearchWithOptions(query, SearchOptions{
		Limit:         options.Limit * 2, // Get more for better selection
		ContextBoosts: options.ContextBoosts,
		PipelineOnly:  options.PipelineOnly,
		PipelineBoost: options.PipelineBoost,
		UseFuzzy:      false, // Disable fuzzy for exact search
	})

	// If we have good exact results, return them
	if len(exactResults) >= options.Limit && exactResults[0].Score > 0.5 {
		if len(exactResults) > options.Limit {
			exactResults = exactResults[:options.Limit]
		}
		return exactResults
	}

	// If exact search doesn't yield good results, try fuzzy search
	if options.UseFuzzy {
		fuzzyResults := db.performFuzzySearch(query, options)

		// Combine and deduplicate results
		combinedResults := db.combineAndDeduplicateResults(exactResults, fuzzyResults, options.Limit)
		return combinedResults
	}

	// Return exact results if fuzzy is disabled
	if len(exactResults) > options.Limit {
		exactResults = exactResults[:options.Limit]
	}
	return exactResults
}

// performFuzzySearch conducts fuzzy search on the database
func (db *Database) performFuzzySearch(query string, options SearchOptions) []SearchResult {
	// Create search targets combining command and description
	targets := make([]string, len(db.Commands))
	var builder strings.Builder

	for i, cmd := range db.Commands {
		builder.Reset()
		builder.WriteString(cmd.Command)
		builder.WriteByte(' ')
		builder.WriteString(cmd.Description)
		targets[i] = builder.String()
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	var results []SearchResult
	for i, match := range matches {
		if i >= options.Limit*2 { // Get more for better selection
			break
		}

		// Apply fuzzy threshold
		if options.FuzzyThreshold > 0 && match.Score < options.FuzzyThreshold {
			continue
		}

		// Convert fuzzy score to our scoring system
		// Fuzzy scores are negative (better matches have higher negative values)
		// Convert to positive score between 0-1
		normalizedScore := float64(match.Score+100) / 100.0
		if normalizedScore < 0 {
			normalizedScore = 0
		}
		if normalizedScore > 1 {
			normalizedScore = 1
		}

		results = append(results, SearchResult{
			Command: &db.Commands[match.Index],
			Score:   normalizedScore,
		})
	}

	return results
}

// combineAndDeduplicateResults merges exact and fuzzy results, removing duplicates
func (db *Database) combineAndDeduplicateResults(exactResults, fuzzyResults []SearchResult, limit int) []SearchResult {
	seen := make(map[string]bool)
	var combined []SearchResult

	// Add exact results first (they have higher priority)
	for _, result := range exactResults {
		key := result.Command.Command + "|" + result.Command.Description
		if !seen[key] {
			seen[key] = true
			combined = append(combined, result)
		}
	}

	// Add fuzzy results that aren't already included
	for _, result := range fuzzyResults {
		key := result.Command.Command + "|" + result.Command.Description
		if !seen[key] {
			seen[key] = true
			// Slightly reduce fuzzy scores to prioritize exact matches
			result.Score *= 0.8
			combined = append(combined, result)
		}
	}

	// Sort by score
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Score > combined[j].Score
	})

	// Return top results
	if len(combined) > limit {
		combined = combined[:limit]
	}

	return combined
}

// GetSuggestions provides "Did you mean?" suggestions for potential typos
func (db *Database) GetSuggestions(query string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}

	// Extract unique words from commands and descriptions
	wordSet := make(map[string]bool)
	for _, cmd := range db.Commands {
		// Split command into words
		cmdWords := strings.Fields(cmd.Command)
		for _, word := range cmdWords {
			cleanWord := strings.ToLower(strings.Trim(word, "-_.[]{}()"))
			if len(cleanWord) > 2 { // Ignore very short words
				wordSet[cleanWord] = true
			}
		}

		// Split description into words
		descWords := strings.Fields(cmd.Description)
		for _, word := range descWords {
			cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:()[]{}\"'"))
			if len(cleanWord) > 2 && !isCommonWord(cleanWord) {
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
		// Only suggest if the match is reasonably good
		if match.Score >= -20 { // Adjust threshold as needed
			suggestions = append(suggestions, words[match.Index])
		}
	}

	return suggestions
}

// isCommonWord filters out very common English words that aren't useful for suggestions
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
		"you": true, "your": true, "all": true, "any": true, "can": true, "from": true,
		"not": true, "no": true, "if": true, "when": true, "where": true, "how": true,
		"what": true, "which": true, "who": true, "why": true, "use": true, "used": true,
		"using": true, "file": true, "files": true, "directory": true, "directories": true,
	}
	return commonWords[word]
}

// SearchWithNLP performs natural language search with advanced query processing
func (db *Database) SearchWithNLP(query string, options SearchOptions) []SearchResult {
	if !options.UseNLP {
		// Fall back to regular search if NLP is disabled
		return db.SearchWithFuzzy(query, options)
	}

	// Process query with NLP
	processor := nlp.NewQueryProcessor()
	processedQuery := processor.ProcessQuery(query)

	// Use enhanced keywords for search
	enhancedKeywords := processedQuery.GetEnhancedKeywords()
	enhancedQuery := strings.Join(enhancedKeywords, " ")

	// Perform search with enhanced query
	searchOptions := options
	searchOptions.UseNLP = false // Prevent infinite recursion

	results := db.SearchWithFuzzy(enhancedQuery, searchOptions)

	// Apply intent-based scoring boost
	for i := range results {
		intentBoost := calculateIntentBoost(results[i].Command, processedQuery)
		results[i].Score *= intentBoost
	}

	// Re-sort by updated scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// calculateIntentBoost applies scoring boost based on detected intent
func calculateIntentBoost(cmd *Command, pq *nlp.ProcessedQuery) float64 {
	boost := 1.0
	cmdLower := strings.ToLower(cmd.Command)
	descLower := strings.ToLower(cmd.Description)

	// Apply intent-specific boosts with stronger differentiation
	switch pq.Intent {
	case nlp.IntentFind:
		if strings.Contains(cmdLower, "find") || strings.Contains(cmdLower, "search") ||
			strings.Contains(cmdLower, "ls") || strings.Contains(cmdLower, "grep") {
			boost *= 2.0 // Increased from 1.5
		}

	case nlp.IntentCreate:
		if strings.Contains(cmdLower, "mkdir") || strings.Contains(cmdLower, "touch") ||
			strings.Contains(cmdLower, "create") || strings.Contains(cmdLower, "make") {
			boost *= 2.0 // Increased from 1.5
		}
		// Penalize package creation tools for simple "create directory" queries
		if strings.Contains(cmdLower, "makepkg") &&
			!strings.Contains(descLower, "package") {
			boost *= 0.3 // Strong penalty for mismatched tools
		}

	case nlp.IntentDelete:
		if strings.Contains(cmdLower, "rm") || strings.Contains(cmdLower, "del") ||
			strings.Contains(cmdLower, "delete") || strings.Contains(cmdLower, "remove") {
			boost *= 2.0 // Increased from 1.5
		}

	case nlp.IntentInstall:
		if strings.Contains(cmdLower, "install") || strings.Contains(cmdLower, "add") ||
			strings.Contains(cmdLower, "setup") || strings.Contains(descLower, "install") {
			boost *= 2.0 // Increased from 1.5
		}

	case nlp.IntentRun:
		if strings.Contains(cmdLower, "run") || strings.Contains(cmdLower, "exec") ||
			strings.Contains(cmdLower, "start") || strings.Contains(cmdLower, "launch") {
			boost *= 2.0 // Increased from 1.5
		}

	case nlp.IntentConfigure:
		if strings.Contains(cmdLower, "config") || strings.Contains(cmdLower, "set") ||
			strings.Contains(cmdLower, "configure") || strings.Contains(descLower, "config") {
			boost *= 2.0 // Increased from 1.5
		}
	}

	// Apply action-based boosts with stronger weights for exact matches
	for _, action := range pq.Actions {
		if strings.Contains(cmdLower, action) {
			boost *= 1.5 // Increased from 1.2 for command matches
		} else if strings.Contains(descLower, action) {
			boost *= 1.3 // Increased from 1.2 for description matches
		}

		// Special handling for compression actions
		if action == "compress" || action == "archive" {
			if strings.Contains(cmdLower, "tar") ||
				strings.Contains(cmdLower, "zip") ||
				strings.Contains(cmdLower, "gzip") {
				boost *= 2.5 // Strong boost for compression tools
			}
			// Penalize search tools for compression queries
			if strings.Contains(cmdLower, "find") ||
				strings.Contains(cmdLower, "locate") {
				boost *= 0.2 // Strong penalty
			}
		}
	}

	// Apply target-based boosts with stronger weights
	for _, target := range pq.Targets {
		if strings.Contains(cmdLower, target) {
			boost *= 1.4 // Increased from 1.2 for command matches
		} else if strings.Contains(descLower, target) {
			boost *= 1.2 // Same for description matches
		}
	}

	return boost
}

// getCurrentPlatform returns the platform string used in the command database for the current OS
func getCurrentPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}
