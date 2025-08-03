package database

import (
	"sort"
	"strings"
	"runtime"

	"github.com/Vedant9500/WTF/internal/constants"
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
		options.Limit = constants.DefaultSearchLimit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	results := make([]SearchResult, 0, min(len(db.Commands), options.Limit*constants.ResultsBufferMultiplier))

	currentPlatform := getCurrentPlatform()

	for i := range db.Commands {
		cmd := &db.Commands[i]
		
		// Apply platform filtering and calculate score
		if result := db.calculateCommandScore(cmd, queryWords, options.ContextBoosts, currentPlatform); result != nil {
			results = append(results, *result)
		}
	}

	return db.sortAndLimitResults(results, options.Limit)
}

// SearchWithPipelineOptions performs search with pipeline-specific enhancements
func (db *Database) SearchWithPipelineOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = constants.DefaultSearchLimit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	results := make([]SearchResult, 0, min(len(db.Commands), options.Limit*constants.ResultsBufferMultiplier))

	for i := range db.Commands {
		cmd := &db.Commands[i]

		// If PipelineOnly is true, skip non-pipeline commands
		if options.PipelineOnly && !isPipelineCommand(cmd) {
			continue
		}

		score := calculateScore(cmd, queryWords, options.ContextBoosts)

		// Apply pipeline boost
		if isPipelineCommand(cmd) && options.PipelineBoost > 0 {
			score *= options.PipelineBoost
		}

		if score > 0 {
			results = append(results, SearchResult{
				Command: cmd,
				Score:   score,
			})
		}
	}

	return db.sortAndLimitResults(results, options.Limit)
}

// sortAndLimitResults sorts results by score and applies limit
func (db *Database) sortAndLimitResults(results []SearchResult, limit int) []SearchResult {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// isPipelineCommand checks if a command is likely a pipeline
func isPipelineCommand(cmd *Command) bool {
	if cmd.Pipeline {
		return true
	}
	
	command := cmd.Command
	return strings.Contains(command, "|") ||
		strings.Contains(strings.ToLower(command), "pipe") ||
		strings.Contains(command, "&&") ||
		strings.Contains(command, ">>")
}

// calculateCommandScore handles platform filtering and score calculation for a single command
func (db *Database) calculateCommandScore(cmd *Command, queryWords []string, contextBoosts map[string]float64, currentPlatform string) *SearchResult {
	// Platform filtering: handle new cross-platform designation
	if len(cmd.Platform) > 0 {
		isCrossPlatform := false
		platformMatch := false
		
		for _, p := range cmd.Platform {
			if strings.EqualFold(p, "cross-platform") {
				isCrossPlatform = true
				break
			}
			if strings.EqualFold(p, currentPlatform) {
				platformMatch = true
				break
			}
		}
		
		// Skip if platform-specific and doesn't match current platform
		if !isCrossPlatform && !platformMatch {
			// Check if this is a legacy cross-platform tool (for backward compatibility)
			if isCrossPlatformTool(cmd.Command) {
				// Apply small penalty for legacy cross-platform tools
				score := calculateScore(cmd, queryWords, contextBoosts) * constants.CrossPlatformPenalty
				if score > 0 {
					return &SearchResult{Command: cmd, Score: score}
				}
			}
			// Skip platform-specific tools that don't match
			return nil
		}
	}
	
	score := calculateScore(cmd, queryWords, contextBoosts)
	if score > 0 {
		return &SearchResult{Command: cmd, Score: score}
	}
	return nil
}

// calculateWordScore computes the relevance score for a single word against a command
func calculateWordScore(word string, cmd *Command) float64 {
	cmdLower := cmd.CommandLower
	descLower := cmd.DescriptionLower
	keywordsLower := cmd.KeywordsLower
	tagsLower := cmd.TagsLower
	
	var wordScore float64

	// HIGHEST PRIORITY: Exact match in command name
	if strings.Contains(cmdLower, word) {
		if strings.HasPrefix(cmdLower, word) || cmdLower == word {
			wordScore += constants.DirectCommandMatchScore
		} else {
			wordScore += constants.CommandMatchScore
		}
	}

	// HIGH PRIORITY: Domain-specific command matching
	if isDomainSpecificMatch(word, cmd) {
		wordScore += constants.DomainSpecificScore
	}

	// MEDIUM-HIGH PRIORITY: Exact match in description
	if strings.Contains(descLower, word) {
		wordScore += constants.DescriptionMatchScore
	}

	// MEDIUM-HIGH PRIORITY: Exact match in tags
	for _, tag := range tagsLower {
		if tag == word {
			wordScore += constants.TagExactScore
			break
		}
	}

	// MEDIUM PRIORITY: Exact match in keywords
	for _, keyword := range keywordsLower {
		if keyword == word {
			wordScore += constants.KeywordExactScore
			break
		}
	}

	// LOW-MEDIUM PRIORITY: Partial match in tags (if no exact tag match)
	if wordScore < constants.TagExactScore {
		for _, tag := range tagsLower {
			if strings.Contains(tag, word) {
				wordScore += constants.TagPartialScore
				break
			}
		}
	}

	// LOW PRIORITY: Partial match in keywords (only if no exact match)
	if wordScore == 0 {
		for _, keyword := range keywordsLower {
			if strings.Contains(keyword, word) {
				wordScore += constants.KeywordPartialScore
				break
			}
		}
	}

	return wordScore
}

// calculateScore computes relevance score for a command based on query words and context
func calculateScore(cmd *Command, queryWords []string, contextBoosts map[string]float64) float64 {
	var score float64

	for _, word := range queryWords {
		// Skip very short words
		if len(word) < constants.MinWordLength {
			continue
		}

		wordScore := calculateWordScore(word, cmd)

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
			score *= (1.0 + boost*constants.NicheBoostFactor)
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
		options.Limit = constants.DefaultSearchLimit
	}

	// First try exact search
	exactOptions := options
	exactOptions.Limit = options.Limit * constants.FuzzySearchMultiplier
	exactOptions.UseFuzzy = false
	exactResults := db.SearchWithOptions(query, exactOptions)

	// If we have good exact results, return them
	if len(exactResults) >= options.Limit && exactResults[0].Score > constants.FuzzyScoreThreshold {
		return db.limitResults(exactResults, options.Limit)
	}

	// If exact search doesn't yield good results, try fuzzy search
	if options.UseFuzzy {
		fuzzyResults := db.performFuzzySearch(query, options)
		return db.combineAndDeduplicateResults(exactResults, fuzzyResults, options.Limit)
	}

	// Return exact results if fuzzy is disabled
	return db.limitResults(exactResults, options.Limit)
}

// limitResults applies limit to results slice
func (db *Database) limitResults(results []SearchResult, limit int) []SearchResult {
	if len(results) > limit {
		return results[:limit]
	}
	return results
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
		normalizedScore := float64(match.Score+int(constants.FuzzyNormalizationBase)) / constants.FuzzyNormalizationBase
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
		maxSuggestions = constants.DefaultMaxResults
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
		if match.Score >= constants.FuzzySuggestionThreshold {
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

	// Apply intent-specific boosts
	boost *= applyIntentBoost(cmdLower, descLower, pq.Intent)
	
	// Apply action and target boosts
	boost *= applyActionBoosts(cmdLower, descLower, pq.Actions)
	boost *= applyTargetBoosts(cmdLower, descLower, pq.Targets)

	return boost
}

// applyIntentBoost applies boost based on detected intent
func applyIntentBoost(cmdLower, descLower string, intent nlp.QueryIntent) float64 {
	switch intent {
	case nlp.IntentFind:
		if containsAny(cmdLower, []string{"find", "search", "ls", "grep"}) {
			return 2.0
		}
	case nlp.IntentCreate:
		if containsAny(cmdLower, []string{"mkdir", "touch", "create", "make"}) {
			boost := 2.0
			// Penalize package creation tools for simple "create directory" queries
			if strings.Contains(cmdLower, "makepkg") && !strings.Contains(descLower, "package") {
				boost *= 0.3
			}
			return boost
		}
	case nlp.IntentDelete:
		if containsAny(cmdLower, []string{"rm", "del", "delete", "remove"}) {
			return 2.0
		}
	case nlp.IntentInstall:
		if containsAny(cmdLower, []string{"install", "add", "setup"}) || strings.Contains(descLower, "install") {
			return 2.0
		}
	case nlp.IntentRun:
		if containsAny(cmdLower, []string{"run", "exec", "start", "launch"}) {
			return 2.0
		}
	case nlp.IntentConfigure:
		if containsAny(cmdLower, []string{"config", "set", "configure"}) || strings.Contains(descLower, "config") {
			return 2.0
		}
	}
	return 1.0
}

// applyActionBoosts applies boosts based on detected actions
func applyActionBoosts(cmdLower, descLower string, actions []string) float64 {
	boost := 1.0
	for _, action := range actions {
		if strings.Contains(cmdLower, action) {
			boost *= 1.5
		} else if strings.Contains(descLower, action) {
			boost *= 1.3
		}

		// Special handling for compression actions
		if action == "compress" || action == "archive" {
			if containsAny(cmdLower, []string{"tar", "zip", "gzip"}) {
				boost *= 2.5
			}
			if containsAny(cmdLower, []string{"find", "locate"}) {
				boost *= 0.2
			}
		}
	}
	return boost
}

// applyTargetBoosts applies boosts based on detected targets
func applyTargetBoosts(cmdLower, descLower string, targets []string) float64 {
	boost := 1.0
	for _, target := range targets {
		if strings.Contains(cmdLower, target) {
			boost *= 1.4
		} else if strings.Contains(descLower, target) {
			boost *= 1.2
		}
	}
	return boost
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
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

// crossPlatformTools contains tools that work on Windows, macOS, and Linux
var crossPlatformTools = map[string]bool{
	// Version control and development tools
	"git": true, "docker": true, "node": true, "npm": true, "yarn": true,
	"python": true, "pip": true, "go": true, "cargo": true, "rustc": true,
	"java": true, "javac": true, "mvn": true, "gradle": true,
	
	// Network and file tools (available via Git Bash, WSL, MSYS2 on Windows)
	"curl": true, "wget": true, "ssh": true, "scp": true, "rsync": true,
	"mv": true, "cp": true, "rm": true, "ls": true, "cat": true,
	"grep": true, "find": true, "sed": true, "awk": true,
	
	// Editors and utilities
	"code": true, "vim": true, "nano": true, "tar": true, "gzip": true,
	"unzip": true, "zip": true, "7z": true, "ffmpeg": true, "imagemagick": true,
	"convert": true,
	
	// Cloud and DevOps tools
	"heroku": true, "aws": true, "gcloud": true, "az": true, "kubectl": true,
	"helm": true, "terraform": true, "ansible": true, "vagrant": true,
	
	// Language-specific tools
	"composer": true, "php": true, "ruby": true, "gem": true, "bundle": true,
	"rails": true, "dotnet": true, "nuget": true, "flutter": true, "dart": true,
	
	// Frontend tools
	"ionic": true, "cordova": true, "electron": true, "ng": true, "vue": true,
	"react": true, "create-react-app": true, "webpack": true, "babel": true,
	"eslint": true, "prettier": true,
	
	// Testing tools
	"jest": true, "mocha": true, "cypress": true, "playwright": true, "selenium": true,
}

// isCrossPlatformTool checks if a command is from a cross-platform tool
func isCrossPlatformTool(command string) bool {
	cmdLower := strings.ToLower(command)
	
	// Check if the command starts with any of the cross-platform tools
	for tool := range crossPlatformTools {
		if strings.HasPrefix(cmdLower, tool+" ") || cmdLower == tool {
			return true
		}
	}
	
	return false
}
