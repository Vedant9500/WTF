package database

import (
	"runtime"
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/constants"
	"github.com/Vedant9500/WTF/internal/nlp"
	"github.com/Vedant9500/WTF/internal/utils"
	"github.com/sahilm/fuzzy"
)

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
	TopTermsCap    int     // Cap for top-IDF term selection in universal search (0 = default)
}

// Search performs a basic keyword-based search
// Deprecated: Use SearchUniversal for better performance and accuracy
func (db *Database) Search(query string, limit int) []SearchResult {
	return db.SearchUniversal(query, SearchOptions{
		Limit: limit,
	})
}

// SearchWithOptions performs search with advanced options including context awareness and platform filtering
// Deprecated: Use SearchUniversal for better BM25F-based ranking and NLP integration
func (db *Database) SearchWithOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = constants.DefaultSearchLimit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	results := make([]SearchResult, 0, utils.Min(len(db.Commands), options.Limit*constants.ResultsBufferMultiplier))

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
// Deprecated: Use SearchUniversal with PipelineOnly=true and PipelineBoost options
func (db *Database) SearchWithPipelineOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = constants.DefaultSearchLimit
	}

	queryWords := strings.Fields(strings.ToLower(query))
	results := make([]SearchResult, 0, utils.Min(len(db.Commands), options.Limit*constants.ResultsBufferMultiplier))

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

		if !isCrossPlatform && !platformMatch {
			if isCrossPlatformTool(cmd.Command) {
				score := calculateScore(cmd, queryWords, contextBoosts) * constants.CrossPlatformPenalty
				if score > 0 {
					return &SearchResult{Command: cmd, Score: score}
				}
			}
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
	var wordScore float64

	// Calculate scores from different sources
	wordScore += calculateCommandScore(word, cmd.CommandLower)
	wordScore += calculateDomainScore(word, cmd)
	wordScore += calculateKeywordScore(word, cmd.KeywordsLower)
	wordScore += calculateDescriptionScore(word, cmd.DescriptionLower)
	wordScore += calculateTagScore(word, cmd.TagsLower)

	return wordScore
}

// calculateCommandScore computes score based on command name matching
func calculateCommandScore(word, cmdLower string) float64 {
	// Exact command match
	if cmdLower == word {
		return constants.DirectCommandMatchScore * constants.ExactCommandMatchMultiplier
	}

	// Command starts with the word
	if strings.HasPrefix(cmdLower, word+" ") || strings.HasPrefix(cmdLower, word) {
		return constants.DirectCommandMatchScore * constants.PrefixCommandMatchMultiplier
	}

	// Word appears as a separate word in command
	if strings.Contains(cmdLower, " "+word+" ") || strings.Contains(cmdLower, " "+word) {
		return constants.CommandMatchScore
	}

	// Word appears anywhere in command
	if strings.Contains(cmdLower, word) {
		return constants.CommandMatchScore * constants.ContainsMatchMultiplier
	}

	return 0
}

// calculateDomainScore computes score for domain-specific matching
func calculateDomainScore(word string, cmd *Command) float64 {
	if isDomainSpecificMatch(word, cmd) {
		return constants.DomainSpecificScore
	}
	return 0
}

// calculateKeywordScore computes score based on keyword matching
func calculateKeywordScore(word string, keywordsLower []string) float64 {
	// Check for exact match first
	for _, keyword := range keywordsLower {
		if keyword == word {
			return constants.KeywordExactScore * constants.KeywordExactMatchMultiplier
		}
	}

	// Check for partial match if no exact match
	for _, keyword := range keywordsLower {
		if strings.Contains(keyword, word) {
			return constants.KeywordPartialScore
		}
	}

	return 0
}

// calculateDescriptionScore computes score based on description matching
func calculateDescriptionScore(word, descLower string) float64 {
	// Complete word match in description
	if strings.Contains(descLower, " "+word+" ") || strings.HasPrefix(descLower, word+" ") || strings.HasSuffix(descLower, " "+word) {
		return constants.DescriptionMatchScore
	}

	// Partial match in description
	if strings.Contains(descLower, word) {
		return constants.DescriptionMatchScore * constants.PartialMatchScoreMultiplier
	}

	return 0
}

// calculateTagScore computes score based on tag matching
func calculateTagScore(word string, tagsLower []string) float64 {
	// Check for exact match first
	for _, tag := range tagsLower {
		if tag == word {
			return constants.TagExactScore
		}
	}

	// Check for partial match if no exact match
	for _, tag := range tagsLower {
		if strings.Contains(tag, word) {
			return constants.TagPartialScore
		}
	}

	return 0
}

// calculateScore computes relevance score for a command based on query words and context
func calculateScore(cmd *Command, queryWords []string, contextBoosts map[string]float64) float64 {
	var score float64
	var maxWordScore float64
	matchedWords := 0

	for _, word := range queryWords {
		if len(word) < constants.MinWordLength {
			continue
		}

		wordScore := calculateWordScore(word, cmd)

		if wordScore > maxWordScore {
			maxWordScore = wordScore
		}

		if wordScore > 0 {
			matchedWords++
		}

		if contextBoosts != nil {
			if boost, exists := contextBoosts[word]; exists {
				wordScore *= boost
			}
		}

		score += wordScore
	}

	if len(queryWords) > 1 && matchedWords > 1 {
		completenessBonus := float64(matchedWords) / float64(len(queryWords))
		score *= (1.0 + completenessBonus*0.5)
	}

	if maxWordScore >= constants.DirectCommandMatchScore {
		score *= constants.DirectCommandMatchBonus
	} else if maxWordScore >= constants.CommandMatchScore {
		score *= constants.CommandMatchBonus
	}

	score *= getCategoryRelevanceBoost(cmd, queryWords)

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

	domainMappings := map[string][]string{
		"compress":   {"tar", "gzip", "zip", "bzip", "7z", "compress", "archive"},
		"archive":    {"tar", "gzip", "zip", "bzip", "7z", "compress", "archive", "unzip"},
		"extract":    {"tar", "unzip", "gunzip", "extract", "unarchive"},
		"directory":  {"mkdir", "rmdir", "ls", "dir", "cd", "pwd"},
		"folder":     {"mkdir", "rmdir", "ls", "dir", "cd", "pwd"},
		"create":     {"mkdir", "touch", "make", "new"},
		"file":       {"cp", "mv", "rm", "touch", "cat", "less", "more"},
		"search":     {"grep", "find", "locate", "ag", "rg"},
		"download":   {"wget", "curl", "fetch", "download"},
		"git":        {"git", "clone", "commit", "push", "pull", "branch"},
		"package":    {"apt", "yum", "dnf", "pkg", "brew", "pip", "npm"},
		"process":    {"ps", "kill", "top", "htop", "jobs"},
		"network":    {"ping", "ssh", "scp", "rsync", "nc", "nmap"},
		"edit":       {"vim", "nano", "emacs", "edit", "sed", "awk"},
		"permission": {"chmod", "chown", "chgrp", "sudo"},
		"new":        {"mkdir", "touch", "create", "make"},
	}

	// Check if the command belongs to the word's domain
	if commands, exists := domainMappings[word]; exists {
		for _, domainCmd := range commands {
			// Check for exact command match or command starting with the domain command
			if cmdLower == domainCmd || strings.HasPrefix(cmdLower, domainCmd+" ") {
				return true
			}
		}
	}

	return false
}

// getCategoryRelevanceBoost applies category-based relevance scoring
func getCategoryRelevanceBoost(cmd *Command, queryWords []string) float64 {
	boost := 1.0
	cmdLower := strings.ToLower(cmd.Command)

	for _, word := range queryWords {
		categoryBoost := getCategoryBoostForWord(word, cmdLower)
		boost *= categoryBoost
	}

	return boost
}

// getCategoryBoostForWord returns the boost factor for a specific word and command
func getCategoryBoostForWord(word, cmdLower string) float64 {
	switch word {
	case "compress", "archive":
		return getCompressionBoost(cmdLower)
	case "zip":
		return getZipBoost(cmdLower)
	case "tar":
		return getTarBoost(cmdLower)
	case "directory", "folder":
		return getDirectoryBoost(cmdLower)
	case "create":
		return getCreateBoost(cmdLower)
	case "new":
		return getNewBoost(cmdLower)
	case "search", "find":
		return getSearchBoost(cmdLower)
	case "download", "get":
		return getDownloadBoost(cmdLower)
	default:
		return 1.0
	}
}

// getCompressionBoost returns boost for compression-related queries
func getCompressionBoost(cmdLower string) float64 {
	if isCompressionTool(cmdLower) {
		return constants.CategoryBoostSpecialCompression
	}
	if isSearchTool(cmdLower) {
		return constants.CategoryBoostSearchPenalty
	}
	return 1.0
}

// getZipBoost returns boost for zip-specific queries
func getZipBoost(cmdLower string) float64 {
	if cmdLower == "zip" || strings.HasPrefix(cmdLower, "zip ") {
		return 3.0
	}
	if strings.Contains(cmdLower, "bzip") || strings.Contains(cmdLower, "gzip") {
		return 0.3
	}
	return 1.0
}

// getTarBoost returns boost for tar-specific queries
func getTarBoost(cmdLower string) float64 {
	if cmdLower == "tar" || strings.HasPrefix(cmdLower, "tar ") {
		return 3.0
	}
	return 1.0
}

// getDirectoryBoost returns boost for directory-related queries
func getDirectoryBoost(cmdLower string) float64 {
	if isMkdirCommand(cmdLower) {
		return constants.CategoryBoostDirectory * 2.0
	}
	if isPackageCreationTool(cmdLower) {
		return 0.2
	}
	return 1.0
}

// getCreateBoost returns boost for create-related queries
func getCreateBoost(cmdLower string) float64 {
	if isMkdirCommand(cmdLower) {
		return constants.CategoryBoostDirectory * 1.8
	}
	return 1.0
}

// getNewBoost returns boost for new-related queries
func getNewBoost(cmdLower string) float64 {
	if isMkdirCommand(cmdLower) {
		return constants.CategoryBoostDirectory * 1.5
	}
	if isPackageCreationTool(cmdLower) {
		return 0.3
	}
	return 1.0
}

// getSearchBoost returns boost for search-related queries
func getSearchBoost(cmdLower string) float64 {
	if isSearchTool(cmdLower) {
		return constants.CategoryBoostSearch
	}
	return 1.0
}

// getDownloadBoost returns boost for download-related queries
func getDownloadBoost(cmdLower string) float64 {
	if isDownloadTool(cmdLower) {
		return constants.CategoryBoostDownload
	}
	return 1.0
}

// Helper functions for command classification
func isCompressionTool(cmdLower string) bool {
	return strings.HasPrefix(cmdLower, "tar ") || cmdLower == "tar" ||
		strings.HasPrefix(cmdLower, "zip ") || cmdLower == "zip" ||
		strings.HasPrefix(cmdLower, "gzip ") || cmdLower == "gzip"
}

func isSearchTool(cmdLower string) bool {
	return strings.Contains(cmdLower, "find") || strings.Contains(cmdLower, "locate") ||
		strings.Contains(cmdLower, "grep")
}

func isMkdirCommand(cmdLower string) bool {
	return strings.HasPrefix(cmdLower, "mkdir") || cmdLower == "mkdir"
}

func isPackageCreationTool(cmdLower string) bool {
	return strings.Contains(cmdLower, "cargo") || strings.Contains(cmdLower, "conda")
}

func isDownloadTool(cmdLower string) bool {
	return strings.Contains(cmdLower, "wget") || strings.Contains(cmdLower, "curl")
}

// SearchWithFuzzy performs hybrid search combining exact matching and fuzzy search
// Deprecated: Use SearchUniversal with UseFuzzy=true option
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

	// Use shared TF-IDF searcher if available
	if db.tfidf != nil && db.cmdIndex != nil {
		tfidfResults := db.tfidf.Search(query, options.Limit*2) // Get more results for better selection

		// Convert TF-IDF results to database SearchResult format
		var results []SearchResult
		for _, tfidfResult := range tfidfResults {
			if tfidfResult.CommandIndex < len(db.Commands) {
				results = append(results, SearchResult{
					Command: &db.Commands[tfidfResult.CommandIndex],
					Score:   tfidfResult.Similarity * 100.0, // Scale similarity to match other scores
				})
			}
		}

		// Apply final limit
		if len(results) > options.Limit {
			results = results[:options.Limit]
		}

		return results
	}

	// Fallback: Create temporary TF-IDF searcher (for backward compatibility)
	nlpCommands := make([]nlp.Command, len(db.Commands))
	for i, cmd := range db.Commands {
		nlpCommands[i] = nlp.Command{
			Command:     cmd.Command,
			Description: cmd.Description,
			Keywords:    cmd.Keywords,
		}
	}

	tfidfSearcher := nlp.NewTFIDFSearcher(nlpCommands)
	tfidfResults := tfidfSearcher.Search(query, options.Limit*2)

	// Convert TF-IDF results to database SearchResult format
	var results []SearchResult
	for _, tfidfResult := range tfidfResults {
		results = append(results, SearchResult{
			Command: &db.Commands[tfidfResult.CommandIndex],
			Score:   tfidfResult.Score,
		})
	}

	// If TF-IDF doesn't find enough results, fall back to traditional search
	if len(results) < options.Limit {
		// Process query with traditional NLP
		processor := nlp.NewQueryProcessor()
		processedQuery := processor.ProcessQuery(query)

		// Use enhanced keywords for search
		enhancedKeywords := processedQuery.GetEnhancedKeywords()
		enhancedQuery := strings.Join(enhancedKeywords, " ")

		// Perform search with enhanced query
		searchOptions := options
		searchOptions.UseNLP = false                       // Prevent infinite recursion
		searchOptions.Limit = options.Limit - len(results) // Only get remaining needed results

		fallbackResults := db.SearchWithFuzzy(enhancedQuery, searchOptions)

		// Apply intent-based scoring boost to fallback results
		for i := range fallbackResults {
			intentBoost := calculateIntentBoost(fallbackResults[i].Command, processedQuery)
			fallbackResults[i].Score *= intentBoost * constants.FallbackResultPriority // Slightly lower priority than TF-IDF
		}

		// Combine results
		results = append(results, fallbackResults...)
	}

	// Re-sort by updated scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply final limit
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

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
	case nlp.IntentView:
		if containsAny(cmdLower, []string{"cat", "less", "more", "head", "tail", "view"}) {
			return 2.5 // Higher boost for view commands
		}
		if strings.Contains(descLower, "display") || strings.Contains(descLower, "show") ||
			strings.Contains(descLower, "view") || strings.Contains(descLower, "print") {
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
	case nlp.IntentModify:
		if containsAny(cmdLower, []string{"chmod", "chown", "edit", "modify", "change"}) {
			return 2.0
		}
		if strings.Contains(descLower, "permission") || strings.Contains(descLower, "modify") {
			return 1.8
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
