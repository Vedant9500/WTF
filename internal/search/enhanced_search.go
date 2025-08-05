// Package search provides enhanced search capabilities with better fuzzy matching and NLP
package search

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/Vedant9500/WTF/internal/database"
)

// LevenshteinDistance calculates the edit distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// FuzzyMatch represents a fuzzy match result
type FuzzyMatch struct {
	Text     string
	Score    float64
	Distance int
}

// EnhancedSearcher provides improved search capabilities with caching
type EnhancedSearcher struct {
	db          *database.Database
	queryCache  map[string][]SearchResult
	maxCacheSize int
}

// NewEnhancedSearcher creates a new enhanced searcher with caching
func NewEnhancedSearcher(db *database.Database) *EnhancedSearcher {
	return &EnhancedSearcher{
		db:           db,
		queryCache:   make(map[string][]SearchResult),
		maxCacheSize: 100, // Cache up to 100 queries
	}
}

// CommonTypos maps common typos to correct spellings
var CommonTypos = map[string]string{
	"gti":      "git",
	"comit":    "commit",
	"comitt":   "commit",
	"committ":  "commit",
	"finde":    "find",
	"creete":   "create",
	"directry": "directory",
	"coppy":    "copy",
	"convet":   "convert",
	"changez":  "changes",
	"remot":    "remote",
	"lsit":     "list",
	"mkdri":    "mkdir",
	"mkidr":    "mkdir",
	"rmdir":    "rmdir",
	"chnage":   "change",
	"permision": "permission",
	"permisions": "permissions",
	"recusrive": "recursive",
	"recursiv": "recursive",
	"bakup":    "backup",
	"restor":   "restore",
	"databse":  "database",
	"databas":  "database",
	"compres":  "compress",
	"archiv":   "archive",
	"extrac":   "extract",
	"downlaod": "download",
	"donwload": "download",
	"netwrok":  "network",
	"netowrk":  "network",
	"moniter":  "monitor",
	"bandwith": "bandwidth",
	"trafic":   "traffic",
	"traffik":  "traffic",
}

// StopWords are common words that don't add search value
var StopWords = map[string]bool{
	"how": true, "do": true, "i": true, "to": true, "a": true, "an": true, "the": true,
	"and": true, "or": true, "but": true, "in": true, "on": true, "at": true, "by": true,
	"for": true, "with": true, "from": true, "into": true, "of": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "have": true, "has": true, "had": true,
	"will": true, "would": true, "could": true, "should": true, "can": true, "may": true,
	"might": true, "must": true, "shall": true, "this": true, "that": true, "these": true,
	"those": true, "it": true, "its": true, "you": true, "your": true, "my": true, "me": true,
	"we": true, "us": true, "our": true, "they": true, "them": true, "their": true,
}

// KeywordSynonyms maps natural language terms to technical keywords
var KeywordSynonyms = map[string][]string{
	// File format conversion
	"convert":    {"convert", "transform", "yq", "jq"},
	"converts":   {"convert", "transform", "yq", "jq"},
	"json":       {"json", "yq", "jq"},
	"yaml":       {"yaml", "yq"},
	"pretty":     {"format", "pretty", "yq"},
	"printed":    {"format", "pretty", "yq"},
	
	// Text processing and counting
	"count":      {"count", "wc"},
	"counts":     {"count", "wc"},
	"words":      {"words", "wc"},
	"lines":      {"lines", "wc", "head", "tail"},
	"characters": {"characters", "wc", "tr"},
	"bytes":      {"bytes", "wc"},
	
	// Deduplication
	"duplicate":   {"duplicate", "uniq"},
	"duplicates":  {"duplicate", "uniq"},
	"remove":      {"remove", "rm", "uniq", "tr"},
	"removes":     {"remove", "rm", "uniq", "tr"},
	"unique":      {"unique", "uniq"},
	"sorted":      {"sorted", "sort", "uniq"},
	
	// Text manipulation
	"extract":     {"extract", "sed", "awk", "grep"},
	"extracts":    {"extract", "sed", "awk", "grep"},
	"between":     {"between", "sed", "awk"},
	"patterns":    {"patterns", "grep", "sed", "regex"},
	"pattern":     {"pattern", "grep", "sed", "regex"},
	"matching":    {"matching", "grep", "sed"},
	"uppercase":   {"uppercase", "tr"},
	"lowercase":   {"lowercase", "tr"},
	"whitespace":  {"whitespace", "tr", "sed"},
	"leading":     {"leading", "sed", "tr"},
	"trailing":    {"trailing", "sed", "tr"},
	"printable":   {"printable", "tr"},
	
	// File operations
	"split":       {"split"},
	"splits":      {"split"},
	"chunks":      {"chunks", "split"},
	"smaller":     {"smaller", "split"},
	"sort":        {"sort"},
	"sorts":       {"sort"},
	"alphabetically": {"alphabetically", "sort"},
	"numerically": {"numerically", "sort"},
	"first":       {"first", "head"},
	"head":        {"head"},
	"tail":        {"tail"},
	"last":        {"last", "tail"},
	
	// Pattern matching
	"find":        {"find", "grep", "locate"},
	"finds":       {"find", "grep", "locate"},
	"match":       {"match", "grep"},
	"matches":     {"match", "grep"},
	"regex":       {"regex", "grep", "sed"},
	
	// Output redirection
	"standard":    {"standard", "tee"},
	"output":      {"output", "tee"},
	"stdout":      {"stdout", "tee"},
	"terminal":    {"terminal", "tee"},
	"display":     {"display", "tee", "cat"},
	"save":        {"save", "tee", "cp"},
	"copy":        {"copy", "cp", "tee"},
	
	// Downloads
	"download":    {"download", "wget", "curl"},
	"downloads":   {"download", "wget", "curl"},
	"url":         {"url", "wget", "curl"},
	
	// File system
	"file":        {"file", "files"},
	"files":       {"files", "ls", "find"},
	"subdirectories": {"subdirectories", "ls", "find"},
	"folder":      {"folder", "directory", "ls"},
	"inside":      {"inside", "ls", "find"},
	
	// Legacy mappings
	"compress":    {"compress", "zip", "tar", "gzip", "archive", "pack"},
	"create":      {"create", "make", "mkdir", "touch", "new"},
	"delete":      {"delete", "remove", "rm", "del", "erase"},
	"move":        {"move", "mv", "rename", "relocate"},
	"list":        {"list", "ls", "show", "display", "dir"},
	"edit":        {"edit", "modify", "change", "update", "vim", "nano"},
	"upload":      {"upload", "push", "send", "put"},
	"backup":      {"backup", "save", "export", "dump"},
	"restore":     {"restore", "import", "load", "recover"},
	"install":     {"install", "setup", "add", "mount"},
	"monitor":     {"monitor", "watch", "track", "observe", "top", "ps"},
	"permission":  {"permission", "chmod", "chown", "access", "rights"},
	"network":     {"network", "ping", "ssh", "curl", "wget", "nc"},
	"process":     {"process", "kill", "ps", "top", "jobs"},
	"directory":   {"directory", "folder", "dir", "directories", "folders"},
	"multiple":    {"multiple", "many", "several", "all"},
	"single":      {"single", "one", "into"},
	"archive":     {"archive", "zip", "tar", "compressed"},
}

// PreprocessQuery cleans, corrects typos, and enhances the query with synonyms
func (es *EnhancedSearcher) PreprocessQuery(query string) string {
	words := strings.Fields(strings.ToLower(query))
	var processedWords []string
	
	// Step 1: Clean and correct typos
	for _, word := range words {
		// Remove punctuation
		cleanWord := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}, word)

		if len(cleanWord) == 0 {
			continue
		}

		// Skip stop words
		if StopWords[cleanWord] {
			continue
		}

		// Check for common typos
		if correction, exists := CommonTypos[cleanWord]; exists {
			processedWords = append(processedWords, correction)
		} else {
			processedWords = append(processedWords, cleanWord)
		}
	}

	// Step 2: Add synonyms for key terms
	var enhancedWords []string
	enhancedWords = append(enhancedWords, processedWords...) // Keep original words
	
	for _, word := range processedWords {
		if synonyms, exists := KeywordSynonyms[word]; exists {
			// Add the most relevant synonyms (limit to avoid query explosion)
			for i, synonym := range synonyms {
				if i >= 2 { // Limit to 2 synonyms per word
					break
				}
				if synonym != word { // Don't add the same word
					enhancedWords = append(enhancedWords, synonym)
				}
			}
		}
	}

	return strings.Join(enhancedWords, " ")
}

// SearchResult represents an enhanced search result
type SearchResult struct {
	Command     *database.Command
	Score       float64
	MatchReason string
	Distance    int
}

// SearchOptions holds platform filtering options
type SearchOptions struct {
	Limit            int
	PlatformFilter   []string // Empty means all platforms
	IncludeCrossPlatform bool
	ShowAllPlatforms bool    // Override platform filtering entirely
}

// FastAdaptiveSearch performs optimized search with caching for speed
func (es *EnhancedSearcher) FastAdaptiveSearch(query string, limit int) []SearchResult {
	return es.FastAdaptiveSearchWithOptions(query, SearchOptions{
		Limit:            limit,
		PlatformFilter:   []string{}, // Default: no filtering
		IncludeCrossPlatform: true,
		ShowAllPlatforms: false,
	})
}

// FastAdaptiveSearchWithOptions performs search with platform filtering options
func (es *EnhancedSearcher) FastAdaptiveSearchWithOptions(query string, options SearchOptions) []SearchResult {
	if options.Limit <= 0 {
		options.Limit = 5
	}

	originalQuery := strings.ToLower(query)
	
	// Create cache key including platform options
	cacheKey := fmt.Sprintf("%s:%d:%v:%v:%v", originalQuery, options.Limit, 
		options.PlatformFilter, options.IncludeCrossPlatform, options.ShowAllPlatforms)
	if cached, exists := es.queryCache[cacheKey]; exists {
		return cached
	}
	
	var allResults []SearchResult

	// Strategy 1: Fast exact and fuzzy search (primary - should be instant)
	exactResults := es.exactSearchWithPlatform(query, options)
	for i := range exactResults {
		exactResults[i].Score *= 1.5 // Boost exact matches
	}
	allResults = append(allResults, exactResults...)

	// Strategy 2: Fast fuzzy word search
	fuzzyResults := es.fuzzyWordSearchWithPlatform(query, options)
	allResults = append(allResults, fuzzyResults...)

	// Strategy 3: Fast intent-based search (using lightweight patterns)
	intentResults := es.fastIntentSearchWithPlatform(query, options)
	for i := range intentResults {
		intentResults[i].Score *= 1.3
	}
	allResults = append(allResults, intentResults...)

	// Strategy 4: Fast partial search as fallback
	if len(allResults) < options.Limit {
		partialResults := es.partialSearchWithPlatform(query, options)
		for i := range partialResults {
			partialResults[i].Score *= 0.9
		}
		allResults = append(allResults, partialResults...)
	}

	// Fast deduplication and ranking
	results := es.fastRanking(allResults, options.Limit, originalQuery)
	
	// Cache the results
	if len(es.queryCache) >= es.maxCacheSize {
		// Simple cache eviction - clear half the cache
		for k := range es.queryCache {
			delete(es.queryCache, k)
			if len(es.queryCache) <= es.maxCacheSize/2 {
				break
			}
		}
	}
	es.queryCache[cacheKey] = results
	
	return results
}

// fastIntentSearch uses lightweight intent detection without heavy computation
func (es *EnhancedSearcher) fastIntentSearch(query string) []SearchResult {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	
	// Lightweight intent patterns (no complex computation)
	intentKeywords := map[string][]string{
		"convert": {"convert", "transform", "json", "yaml", "format"},
		"count": {"count", "number", "words", "lines", "characters"},
		"remove": {"remove", "delete", "duplicate", "clean"},
		"extract": {"extract", "get", "between", "pattern"},
		"split": {"split", "divide", "chunk", "break"},
		"download": {"download", "fetch", "url", "get"},
		"compress": {"compress", "zip", "archive", "tar"},
		"search": {"find", "search", "grep", "locate"},
		"display": {"display", "show", "view", "print"},
		"calendar": {"calendar", "cal", "date", "time"},
		"create_directory": {"create", "new", "make", "directory", "folder", "dir"},
	}
	
	// Quick keyword matching
	for intent, keywords := range intentKeywords {
		for _, keyword := range keywords {
			if strings.Contains(queryLower, keyword) {
				// Fast command matching for this intent
				intentResults := es.quickIntentMatch(intent, queryLower)
				results = append(results, intentResults...)
				break
			}
		}
	}
	
	return results
}

// quickIntentMatch provides fast intent-based matching
func (es *EnhancedSearcher) quickIntentMatch(intent, query string) []SearchResult {
	var results []SearchResult
	
	// Pre-defined quick matchers for performance
	quickMatchers := map[string]func(string, *database.Command) float64{
		"convert": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.Contains(cmdLower, "yq") || strings.Contains(cmdLower, "jq") {
				score += 15.0
			}
			if strings.Contains(descLower, "json") || strings.Contains(descLower, "yaml") {
				score += 10.0
			}
			if strings.Contains(descLower, "convert") {
				score += 8.0
			}
			return score
		},
		"count": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.HasPrefix(cmdLower, "wc") {
				score += 15.0
			}
			if strings.Contains(descLower, "count") {
				score += 10.0
			}
			return score
		},
		"remove": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.Contains(cmdLower, "uniq") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "rm") {
				score += 12.0
			}
			if strings.Contains(descLower, "remove") || strings.Contains(descLower, "duplicate") {
				score += 10.0
			}
			return score
		},
		"extract": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.Contains(cmdLower, "sed") {
				score += 15.0
			}
			if strings.Contains(descLower, "extract") || strings.Contains(descLower, "between") {
				score += 10.0
			}
			return score
		},
		"download": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.Contains(cmdLower, "wget") || strings.Contains(cmdLower, "curl") {
				score += 15.0
			}
			if strings.Contains(descLower, "download") {
				score += 10.0
			}
			return score
		},
		"compress": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			// Boost tar commands significantly for compression queries
			if strings.HasPrefix(cmdLower, "tar ") || cmdLower == "tar" {
				score += 25.0
			}
			// Boost zip commands significantly
			if strings.HasPrefix(cmdLower, "zip ") || cmdLower == "zip" {
				score += 25.0
			}
			// Boost gzip commands
			if strings.Contains(cmdLower, "gzip") {
				score += 20.0
			}
			// Other compression tools
			if strings.Contains(cmdLower, "7z") {
				score += 18.0
			}
			// General compression mentions
			if strings.Contains(descLower, "compress") || strings.Contains(descLower, "archive") {
				score += 10.0
			}
			return score
		},
		"search": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			if strings.Contains(cmdLower, "grep") || strings.Contains(cmdLower, "find") {
				score += 15.0
			}
			if strings.Contains(descLower, "search") || strings.Contains(descLower, "find") {
				score += 10.0
			}
			return score
		},
		"display": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			// Calendar/date display commands
			if strings.Contains(cmdLower, "cal") && strings.Contains(query, "calendar") {
				score += 20.0
			}
			if strings.Contains(cmdLower, "date") && (strings.Contains(query, "date") || strings.Contains(query, "time")) {
				score += 20.0
			}
			
			// General display commands
			if strings.Contains(descLower, "display") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "show") || strings.Contains(descLower, "show") {
				score += 10.0
			}
			
			return score
		},
		"calendar": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			// Direct calendar commands
			if cmdLower == "cal" || cmdLower == "calendar" {
				score += 25.0
			}
			if strings.Contains(cmdLower, "cal") {
				score += 20.0
			}
			if strings.Contains(descLower, "calendar") {
				score += 15.0
			}
			if strings.Contains(descLower, "date") && strings.Contains(query, "calendar") {
				score += 12.0
			}
			
			return score
		},
		"create_directory": func(query string, cmd *database.Command) float64 {
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			score := 0.0
			
			// Direct mkdir command gets highest score
			if cmdLower == "mkdir" {
				score += 30.0
			}
			// Commands that start with mkdir
			if strings.HasPrefix(cmdLower, "mkdir ") {
				score += 25.0
			}
			// Description mentions creating directories
			if strings.Contains(descLower, "create") && (strings.Contains(descLower, "directory") || strings.Contains(descLower, "folder")) {
				score += 20.0
			}
			// General directory creation mentions
			if strings.Contains(descLower, "create directory") {
				score += 18.0
			}
			
			return score
		},
	}
	
	if matcher, exists := quickMatchers[intent]; exists {
		for i := range es.db.Commands {
			cmd := &es.db.Commands[i]
			score := matcher(query, cmd)
			
			if score > 5.0 {
				results = append(results, SearchResult{
					Command:     cmd,
					Score:       score,
					MatchReason: "fast intent: " + intent,
					Distance:    -1,
				})
			}
		}
	}
	
	return results
}

// fastRanking provides quick ranking without heavy computation
func (es *EnhancedSearcher) fastRanking(results []SearchResult, limit int, originalQuery string) []SearchResult {
	seen := make(map[string]*SearchResult)
	
	// Quick deduplication
	for _, result := range results {
		key := result.Command.Command + "|" + result.Command.Description
		if existing, exists := seen[key]; exists {
			if result.Score > existing.Score {
				seen[key] = &result
			}
		} else {
			seen[key] = &result
		}
	}

	// Convert to slice
	var deduplicated []SearchResult
	for _, result := range seen {
		// Quick score adjustment
		adjustedResult := *result
		adjustedResult.Score = es.quickScoreAdjustment(adjustedResult, originalQuery)
		deduplicated = append(deduplicated, adjustedResult)
	}

	// Fast sort
	sort.Slice(deduplicated, func(i, j int) bool {
		return deduplicated[i].Score > deduplicated[j].Score
	})

	// Apply limit
	if len(deduplicated) > limit {
		deduplicated = deduplicated[:limit]
	}

	return deduplicated
}

// shouldIncludeCommand checks if a command should be included based on platform filtering
func (es *EnhancedSearcher) shouldIncludeCommand(cmd *database.Command, options SearchOptions) bool {
	// If ShowAllPlatforms is true, include everything
	if options.ShowAllPlatforms {
		return true
	}
	
	// If no platform filter specified, include all
	if len(options.PlatformFilter) == 0 {
		return true
	}
	
	// If command has no platform specified, include it
	if len(cmd.Platform) == 0 {
		return true
	}
	
	// Check if command matches any of the requested platforms
	for _, cmdPlatform := range cmd.Platform {
		// Always include cross-platform commands if IncludeCrossPlatform is true
		if options.IncludeCrossPlatform && strings.EqualFold(cmdPlatform, "cross-platform") {
			return true
		}
		
		// Check against specific platform filters
		for _, filterPlatform := range options.PlatformFilter {
			if strings.EqualFold(cmdPlatform, filterPlatform) {
				return true
			}
		}
	}
	
	return false
}

// quickScoreAdjustment provides fast score adjustments
func (es *EnhancedSearcher) quickScoreAdjustment(result SearchResult, originalQuery string) float64 {
	score := result.Score
	cmd := result.Command
	
	// Quick boosts based on simple criteria
	if strings.Contains(result.MatchReason, "exact") {
		score *= 1.2
	}
	
	// Boost popular commands (simple heuristic)
	if len(cmd.Keywords) > 3 {
		score *= 1.1
	}
	
	// Boost cross-platform
	for _, platform := range cmd.Platform {
		if platform == "cross-platform" {
			score *= 1.05
			break
		}
	}
	
	return score
}

// exactSearchWithPlatform performs exact search with platform filtering
func (es *EnhancedSearcher) exactSearchWithPlatform(query string, options SearchOptions) []SearchResult {
	results := es.exactSearch(query)
	return es.filterResultsByPlatform(results, options)
}

// fuzzyWordSearchWithPlatform performs fuzzy search with platform filtering
func (es *EnhancedSearcher) fuzzyWordSearchWithPlatform(query string, options SearchOptions) []SearchResult {
	results := es.fuzzyWordSearch(query)
	return es.filterResultsByPlatform(results, options)
}

// fastIntentSearchWithPlatform performs intent search with platform filtering
func (es *EnhancedSearcher) fastIntentSearchWithPlatform(query string, options SearchOptions) []SearchResult {
	results := es.fastIntentSearch(query)
	return es.filterResultsByPlatform(results, options)
}

// partialSearchWithPlatform performs partial search with platform filtering
func (es *EnhancedSearcher) partialSearchWithPlatform(query string, options SearchOptions) []SearchResult {
	results := es.partialSearch(query)
	return es.filterResultsByPlatform(results, options)
}

// filterResultsByPlatform filters search results based on platform options
func (es *EnhancedSearcher) filterResultsByPlatform(results []SearchResult, options SearchOptions) []SearchResult {
	if options.ShowAllPlatforms || len(options.PlatformFilter) == 0 {
		return results
	}
	
	var filtered []SearchResult
	for _, result := range results {
		if es.shouldIncludeCommand(result.Command, options) {
			filtered = append(filtered, result)
		}
	}
	
	return filtered
}

// AdaptiveSearch now uses the fast version by default
func (es *EnhancedSearcher) AdaptiveSearch(query string, limit int) []SearchResult {
	return es.FastAdaptiveSearch(query, limit)
}

// semanticFuzzySearch combines fuzzy matching with semantic similarity
func (es *EnhancedSearcher) semanticFuzzySearch(query string, semanticSearcher *SemanticSearcher) []SearchResult {
	var results []SearchResult
	queryWords := strings.Fields(strings.ToLower(query))
	
	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		
		// Get all words from command
		cmdText := strings.Join([]string{
			cmd.Command,
			cmd.Description,
			strings.Join(cmd.Keywords, " "),
		}, " ")
		cmdWords := strings.Fields(strings.ToLower(cmdText))
		
		var totalScore float64
		matchCount := 0
		
		for _, queryWord := range queryWords {
			bestScore := 0.0
			
			for _, cmdWord := range cmdWords {
				// Combine edit distance with semantic similarity
				editSim := 1.0 - float64(LevenshteinDistance(queryWord, cmdWord))/float64(max(len(queryWord), len(cmdWord)))
				semanticSim := semanticSearcher.calculateSemanticSimilarity(queryWord, cmdWord)
				
				// Weighted combination
				combinedScore := 0.6*editSim + 0.4*semanticSim
				
				if combinedScore > bestScore {
					bestScore = combinedScore
				}
			}
			
			if bestScore > 0.4 { // Threshold for meaningful similarity
				totalScore += bestScore
				matchCount++
			}
		}
		
		if matchCount > 0 {
			// Normalize and boost for coverage
			normalizedScore := totalScore / float64(len(queryWords))
			coverageBoost := float64(matchCount) / float64(len(queryWords))
			finalScore := normalizedScore * (1.0 + coverageBoost)
			
			if finalScore > 0.3 {
				results = append(results, SearchResult{
					Command:     cmd,
					Score:       finalScore,
					MatchReason: "semantic fuzzy match",
					Distance:    -1,
				})
			}
		}
	}
	
	return results
}

// findCommandsByDynamicIntent uses learned patterns to find commands by intent
func (es *EnhancedSearcher) findCommandsByDynamicIntent(intent string, confidence float64, patternLearner *PatternLearner) []SearchResult {
	var results []SearchResult
	
	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		
		// Use pattern learner to calculate intent score
		intentScore := patternLearner.GetDynamicIntentScore(intent, cmd)
		
		if intentScore > 0.2 {
			finalScore := intentScore * confidence
			
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       finalScore,
				MatchReason: "dynamic intent: " + intent,
				Distance:    -1,
			})
		}
	}
	
	return results
}

// traditionalSearch falls back to the original enhanced search methods
func (es *EnhancedSearcher) traditionalSearch(query string, limit int) []SearchResult {
	correctedQuery := es.PreprocessQuery(query)
	
	var results []SearchResult
	
	// Exact search
	exactResults := es.exactSearch(correctedQuery)
	results = append(results, exactResults...)
	
	// Fuzzy search
	fuzzyResults := es.fuzzyWordSearch(correctedQuery)
	results = append(results, fuzzyResults...)
	
	return results
}

// adaptiveRanking provides intelligent ranking based on multiple factors
func (es *EnhancedSearcher) adaptiveRanking(results []SearchResult, limit int, originalQuery string, patternLearner *PatternLearner) []SearchResult {
	seen := make(map[string]*SearchResult)
	
	// Deduplicate, keeping the highest score for each command
	for _, result := range results {
		key := result.Command.Command + "|" + result.Command.Description
		if existing, exists := seen[key]; exists {
			if result.Score > existing.Score {
				seen[key] = &result
			}
		} else {
			seen[key] = &result
		}
	}

	// Convert back to slice and apply adaptive scoring
	var deduplicated []SearchResult
	for _, result := range seen {
		adaptiveResult := *result
		adaptiveResult.Score = es.calculateAdaptiveScore(adaptiveResult, originalQuery, patternLearner)
		deduplicated = append(deduplicated, adaptiveResult)
	}

	// Sort by adaptive score
	sort.Slice(deduplicated, func(i, j int) bool {
		// Primary sort by score
		if math.Abs(deduplicated[i].Score-deduplicated[j].Score) > 0.01 {
			return deduplicated[i].Score > deduplicated[j].Score
		}
		// Secondary sort by command simplicity (shorter commands often better)
		return len(deduplicated[i].Command.Command) < len(deduplicated[j].Command.Command)
	})

	// Apply limit
	if len(deduplicated) > limit {
		deduplicated = deduplicated[:limit]
	}

	return deduplicated
}

// calculateAdaptiveScore applies advanced scoring based on multiple factors
func (es *EnhancedSearcher) calculateAdaptiveScore(result SearchResult, originalQuery string, patternLearner *PatternLearner) float64 {
	score := result.Score
	cmd := result.Command
	
	// Boost based on match reason
	switch {
	case strings.Contains(result.MatchReason, "semantic"):
		score *= 1.3 // Semantic matches are high quality
	case strings.Contains(result.MatchReason, "learned patterns"):
		score *= 1.2 // Pattern-based matches are reliable
	case strings.Contains(result.MatchReason, "dynamic intent"):
		score *= 1.1 // Intent-based matches are good
	}
	
	// Boost for command popularity (estimated by keyword count and description quality)
	if len(cmd.Keywords) > 3 {
		score *= 1.1
	}
	
	// Boost for well-documented commands
	if len(cmd.Description) > 20 && len(cmd.Description) < 150 {
		score *= 1.05
	}
	
	// Boost for cross-platform commands
	for _, platform := range cmd.Platform {
		if platform == "cross-platform" {
			score *= 1.03
			break
		}
	}
	
	// Penalize overly complex commands for simple queries
	queryWords := strings.Fields(originalQuery)
	if len(queryWords) <= 3 && len(cmd.Command) > 50 {
		score *= 0.9
	}
	
	return score
}

// EnhancedSearch maintains backward compatibility while using adaptive search
func (es *EnhancedSearcher) EnhancedSearch(query string, limit int) []SearchResult {
	return es.AdaptiveSearch(query, limit)
}

// exactSearch performs exact string matching with improved natural language handling
func (es *EnhancedSearcher) exactSearch(query string) []SearchResult {
	var results []SearchResult
	queryWords := strings.Fields(strings.ToLower(query))
	
	// Filter out stop words but keep important ones
	filteredWords := es.filterQueryWords(queryWords)
	
	// Add explicit compression tool matching for compression queries
	compressionQuery := false
	directoryCreationQuery := false
	for _, word := range filteredWords {
		if word == "compress" || word == "archive" || word == "zip" {
			compressionQuery = true
		}
		if word == "create" || word == "new" || word == "make" {
			for _, w2 := range filteredWords {
				if w2 == "directory" || w2 == "folder" || w2 == "dir" {
					directoryCreationQuery = true
					break
				}
			}
		}
	}

	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		score := 0.0
		matchReasons := []string{}

		// Check command name (highest priority)
		for _, word := range filteredWords {
			if strings.Contains(cmd.CommandLower, word) {
				if cmd.CommandLower == word {
					score += 25.0 // Boost exact command matches
					matchReasons = append(matchReasons, "exact command")
				} else if strings.HasPrefix(cmd.CommandLower, word) {
					score += 20.0
					matchReasons = append(matchReasons, "command prefix")
				} else {
					score += 15.0
					matchReasons = append(matchReasons, "command contains")
				}
			}
		}
		
		// Special boost for common compression commands when query contains compression terms
		if compressionQuery {
			if strings.HasPrefix(cmd.CommandLower, "tar ") || cmd.CommandLower == "tar" {
				score += 50.0 // Major boost for tar commands
				matchReasons = append(matchReasons, "compression tool")
			}
			if strings.HasPrefix(cmd.CommandLower, "zip ") || cmd.CommandLower == "zip" {
				score += 50.0 // Major boost for zip commands
				matchReasons = append(matchReasons, "compression tool")
			}
			if strings.Contains(cmd.CommandLower, "gzip") {
				score += 40.0
				matchReasons = append(matchReasons, "compression tool")
			}
			if strings.Contains(cmd.CommandLower, "7z") {
				score += 35.0
				matchReasons = append(matchReasons, "compression tool")
			}
		}
		
		// Special boost for directory creation commands
		if directoryCreationQuery {
			if cmd.CommandLower == "mkdir" {
				score += 60.0 // Major boost for mkdir
				matchReasons = append(matchReasons, "directory creation")
			}
			if strings.HasPrefix(cmd.CommandLower, "mkdir ") {
				score += 55.0
				matchReasons = append(matchReasons, "directory creation")
			}
			if strings.Contains(cmd.DescriptionLower, "create") && 
			   (strings.Contains(cmd.DescriptionLower, "directory") || strings.Contains(cmd.DescriptionLower, "folder")) {
				score += 30.0
				matchReasons = append(matchReasons, "directory creation")
			}
		}

		// Check description with better scoring
		descriptionMatches := 0
		for _, word := range filteredWords {
			if strings.Contains(cmd.DescriptionLower, word) {
				score += 10.0 // Increased description weight
				descriptionMatches++
			}
		}
		if descriptionMatches > 0 {
			matchReasons = append(matchReasons, "description")
			// Bonus for multiple description matches
			if descriptionMatches > 1 {
				score += float64(descriptionMatches) * 2.0
			}
		}

		// Check keywords with improved matching
		keywordMatches := 0
		for _, word := range filteredWords {
			for _, keyword := range cmd.KeywordsLower {
				if keyword == word {
					score += 15.0 // Increased keyword weight
					keywordMatches++
				} else if strings.Contains(keyword, word) {
					score += 8.0
					keywordMatches++
				}
				// Handle word variations (compress/compression, file/files)
				if (word == "compress" && strings.Contains(keyword, "compression")) ||
				   (word == "compression" && strings.Contains(keyword, "compress")) ||
				   (word == "files" && keyword == "file-system") ||
				   (word == "file" && keyword == "files") ||
				   (word == "files" && keyword == "file") {
					score += 12.0
					keywordMatches++
				}
			}
		}
		if keywordMatches > 0 {
			matchReasons = append(matchReasons, "keywords")
			// Bonus for multiple keyword matches
			if keywordMatches > 1 {
				score += float64(keywordMatches) * 1.5
			}
		}

		// Boost score based on query coverage
		if len(filteredWords) > 0 {
			coverage := float64(descriptionMatches + keywordMatches) / float64(len(filteredWords))
			score *= (1.0 + coverage)
		}

		if score > 0 {
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       score,
				MatchReason: strings.Join(matchReasons, ", "),
				Distance:    0,
			})
		}
	}

	return results
}

// filterQueryWords removes stop words but keeps important terms
func (es *EnhancedSearcher) filterQueryWords(words []string) []string {
	// Minimal stop words - only remove the most common ones
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "to": true, "in": true, "on": true,
		"at": true, "by": true, "for": true, "with": true, "from": true,
		"i": true, "me": true, "my": true, "we": true, "us": true, "our": true,
		"you": true, "your": true, "it": true, "its": true, "this": true, "that": true,
	}
	
	var filtered []string
	for _, word := range words {
		// Keep words that are 2+ characters and not stop words
		if len(word) >= 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}
	
	// If we filtered out too much, keep the original
	if len(filtered) == 0 {
		return words
	}
	
	return filtered
}

// fuzzyWordSearch performs fuzzy matching on individual words
func (es *EnhancedSearcher) fuzzyWordSearch(query string) []SearchResult {
	var results []SearchResult
	queryWords := strings.Fields(query)

	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		totalScore := 0.0
		minDistance := math.MaxInt32
		matchReasons := []string{}

		// Check each query word against command components
		for _, queryWord := range queryWords {
			if len(queryWord) < 2 {
				continue
			}

			// Check against command name words
			cmdWords := strings.Fields(cmd.CommandLower)
			for _, cmdWord := range cmdWords {
				distance := LevenshteinDistance(queryWord, cmdWord)
				similarity := 1.0 - float64(distance)/float64(max(len(queryWord), len(cmdWord)))
				
				if similarity > 0.6 { // 60% similarity threshold
					score := similarity * 15.0
					totalScore += score
					if distance < minDistance {
						minDistance = distance
					}
					matchReasons = append(matchReasons, "fuzzy command")
				}
			}

			// Check against keywords
			for _, keyword := range cmd.KeywordsLower {
				distance := LevenshteinDistance(queryWord, keyword)
				similarity := 1.0 - float64(distance)/float64(max(len(queryWord), len(keyword)))
				
				if similarity > 0.7 { // Higher threshold for keywords
					score := similarity * 10.0
					totalScore += score
					if distance < minDistance {
						minDistance = distance
					}
					matchReasons = append(matchReasons, "fuzzy keyword")
				}
			}

			// Check against description words
			descWords := strings.Fields(cmd.DescriptionLower)
			for _, descWord := range descWords {
				if len(descWord) < 3 {
					continue
				}
				distance := LevenshteinDistance(queryWord, descWord)
				similarity := 1.0 - float64(distance)/float64(max(len(queryWord), len(descWord)))
				
				if similarity > 0.8 { // Even higher threshold for description
					score := similarity * 5.0
					totalScore += score
					if distance < minDistance {
						minDistance = distance
					}
					matchReasons = append(matchReasons, "fuzzy description")
				}
			}
		}

		if totalScore > 3.0 { // Minimum score threshold
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       totalScore,
				MatchReason: strings.Join(matchReasons, ", "),
				Distance:    minDistance,
			})
		}
	}

	return results
}

// partialSearch performs partial matching with substring search
func (es *EnhancedSearcher) partialSearch(query string) []SearchResult {
	var results []SearchResult
	queryWords := strings.Fields(query)

	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		score := 0.0
		matchReasons := []string{}

		// Partial matching in command
		for _, word := range queryWords {
			if len(word) >= 2 {
				if strings.Contains(cmd.CommandLower, word) {
					score += 5.0
					matchReasons = append(matchReasons, "partial command")
				}
			}
		}

		// Partial matching in description
		for _, word := range queryWords {
			if len(word) >= 3 {
				if strings.Contains(cmd.DescriptionLower, word) {
					score += 3.0
					matchReasons = append(matchReasons, "partial description")
				}
			}
		}

		if score > 0 {
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       score,
				MatchReason: strings.Join(matchReasons, ", "),
				Distance:    -1,
			})
		}
	}

	return results
}

// intentBasedSearch performs intent-based matching with context awareness
func (es *EnhancedSearcher) intentBasedSearch(query string) []SearchResult {
	var results []SearchResult
	
	// Analyze query for compound intents (e.g., "compress multiple files")
	queryLower := strings.ToLower(query)
	detectedIntents := es.detectIntents(queryLower)
	
	for intent, confidence := range detectedIntents {
		if confidence > 0.5 { // Only use high-confidence intents
			intentResults := es.findCommandsByIntent(intent, confidence)
			results = append(results, intentResults...)
		}
	}

	return results
}

// detectIntents analyzes the query to detect user intents with confidence scores
func (es *EnhancedSearcher) detectIntents(query string) map[string]float64 {
	intents := make(map[string]float64)
	
	// Define intent patterns with weights - enhanced with text processing commands
	intentPatterns := map[string]map[string]float64{
		"convert": {
			"convert": 1.0, "converts": 1.0, "transform": 0.8, "change": 0.7,
			"json": 0.9, "yaml": 0.9, "xml": 0.8, "csv": 0.8,
			"pretty-printed": 0.9, "format": 0.8, "parse": 0.7,
		},
		"count": {
			"count": 1.0, "counts": 1.0, "number of": 0.9, "how many": 0.8,
			"words": 0.9, "lines": 0.9, "characters": 0.9, "bytes": 0.8,
			"files": 0.7, "directories": 0.7,
		},
		"remove_duplicates": {
			"remove": 0.8, "removes": 0.8, "duplicate": 1.0, "duplicates": 1.0,
			"unique": 0.9, "deduplicate": 1.0, "sorted": 0.7, "text file": 0.6,
		},
		"text_processing": {
			"extract": 0.8, "extracts": 0.8, "lines between": 1.0, "matching patterns": 1.0,
			"non-printable": 1.0, "characters": 0.7, "uppercase": 0.9, "lowercase": 0.9,
			"leading": 0.8, "trailing": 0.8, "whitespace": 0.9,
		},
		"file_operations": {
			"splits": 1.0, "split": 1.0, "chunks": 0.9, "smaller": 0.7,
			"sorts": 1.0, "sort": 1.0, "alphabetically": 0.9, "numerically": 0.9,
			"first": 0.8, "lines": 0.8, "head": 0.9, "tail": 0.9,
		},
		"pattern_matching": {
			"finds": 0.9, "find": 0.9, "match": 0.9, "matches": 0.9,
			"regex": 1.0, "pattern": 1.0, "grep": 1.0,
		},
		"compress": {
			"compress": 1.0, "zip": 0.9, "tar": 0.9, "archive": 0.8, "gzip": 0.8,
			"pack": 0.7, "bundle": 0.6, "multiple files": 0.8, "single archive": 0.9,
		},
		"extract": {
			"extract": 1.0, "unzip": 0.9, "untar": 0.9, "decompress": 0.8, "unpack": 0.7,
			"open archive": 0.8, "get files from": 0.7,
		},
		"create": {
			"create": 1.0, "make": 0.8, "new": 0.7, "mkdir": 0.9, "touch": 0.9,
			"directory": 0.8, "folder": 0.8, "file": 0.6,
		},
		"delete": {
			"delete": 1.0, "remove": 0.9, "rm": 0.9, "del": 0.8, "erase": 0.7,
			"get rid": 0.6, "clean up": 0.6,
		},
		"copy": {
			"copy": 1.0, "cp": 0.9, "duplicate": 0.8, "clone": 0.7, "backup": 0.6,
			"saves": 0.7, "save": 0.7,
		},
		"move": {
			"move": 1.0, "mv": 0.9, "rename": 0.8, "relocate": 0.7, "transfer": 0.6,
		},
		"list": {
			"list": 1.0, "ls": 0.9, "show": 0.8, "display": 0.8, "dir": 0.9,
			"see": 0.6, "view": 0.7, "contents": 0.7,
		},
		"find": {
			"find": 1.0, "search": 0.9, "locate": 0.8, "grep": 0.8, "look for": 0.7,
			"where is": 0.6, "which": 0.5,
		},
		"download": {
			"download": 1.0, "downloads": 1.0, "fetch": 0.8, "get": 0.7, "wget": 0.9, "curl": 0.9,
			"pull": 0.6, "retrieve": 0.7, "url": 0.8, "from url": 0.9,
		},
		"output_redirect": {
			"standard output": 1.0, "stdout": 1.0, "copy": 0.7, "save": 0.8,
			"file": 0.6, "terminal": 0.8, "display": 0.7, "tee": 1.0,
		},
		"backup": {
			"backup": 1.0, "save": 0.8, "export": 0.7, "dump": 0.8, "preserve": 0.6,
		},
		"restore": {
			"restore": 1.0, "import": 0.8, "load": 0.7, "recover": 0.9, "bring back": 0.6,
		},
		"monitor": {
			"monitor": 1.0, "watch": 0.9, "track": 0.8, "observe": 0.7, "check": 0.6,
			"status": 0.6, "usage": 0.7, "bandwidth": 0.8, "traffic": 0.8,
		},
		"permission": {
			"permission": 1.0, "chmod": 0.9, "chown": 0.8, "access": 0.7, "rights": 0.7,
			"ownership": 0.6, "security": 0.5,
		},
	}
	
	// Score each intent based on pattern matches
	for intent, patterns := range intentPatterns {
		score := 0.0
		matches := 0
		
		for pattern, weight := range patterns {
			if strings.Contains(query, pattern) {
				score += weight
				matches++
			}
		}
		
		if matches > 0 {
			// Normalize score and boost for multiple matches
			normalizedScore := score / float64(len(patterns))
			if matches > 1 {
				normalizedScore *= 1.2 // Boost for multiple pattern matches
			}
			intents[intent] = normalizedScore
		}
	}
	
	return intents
}

// findCommandsByIntent finds commands that match a specific intent with confidence weighting
func (es *EnhancedSearcher) findCommandsByIntent(intent string, confidence float64) []SearchResult {
	var results []SearchResult

	// Define comprehensive intent matching patterns
	intentMatchers := map[string]func(*database.Command) float64{
		"convert": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// JSON/YAML conversion tools
			if strings.Contains(cmdLower, "yq") {
				score += 20.0 // yq is the primary JSON/YAML tool
			}
			if strings.Contains(cmdLower, "jq") {
				score += 15.0 // jq for JSON processing
			}
			if strings.Contains(descLower, "json") && strings.Contains(descLower, "yaml") {
				score += 18.0
			}
			if strings.Contains(descLower, "convert") && (strings.Contains(descLower, "json") || strings.Contains(descLower, "yaml")) {
				score += 15.0
			}
			
			return score
		},
		
		"count": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// wc is the primary counting tool
			if cmdLower == "wc" || strings.HasPrefix(cmdLower, "wc ") {
				score += 20.0
			}
			if strings.Contains(descLower, "count") && (strings.Contains(descLower, "words") || strings.Contains(descLower, "lines") || strings.Contains(descLower, "characters")) {
				score += 15.0
			}
			if strings.Contains(descLower, "number of") {
				score += 12.0
			}
			
			return score
		},
		
		"remove_duplicates": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// uniq is the primary deduplication tool
			if strings.Contains(cmdLower, "uniq") {
				score += 20.0
			}
			if strings.Contains(descLower, "duplicate") && strings.Contains(descLower, "remove") {
				score += 15.0
			}
			if strings.Contains(descLower, "unique") {
				score += 12.0
			}
			
			return score
		},
		
		"text_processing": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// sed for text processing
			if strings.Contains(cmdLower, "sed") {
				score += 18.0
			}
			// tr for character translation
			if cmdLower == "tr" || strings.HasPrefix(cmdLower, "tr ") {
				score += 18.0
			}
			// awk for text processing
			if strings.Contains(cmdLower, "awk") {
				score += 15.0
			}
			
			if strings.Contains(descLower, "non-printable") {
				score += 20.0
			}
			if strings.Contains(descLower, "uppercase") || strings.Contains(descLower, "lowercase") {
				score += 15.0
			}
			if strings.Contains(descLower, "whitespace") {
				score += 12.0
			}
			
			return score
		},
		
		"file_operations": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// split for file splitting
			if strings.Contains(cmdLower, "split") {
				score += 18.0
			}
			// sort for sorting
			if cmdLower == "sort" || strings.HasPrefix(cmdLower, "sort ") {
				score += 18.0
			}
			// head for first lines
			if strings.Contains(cmdLower, "head") {
				score += 18.0
			}
			// tail for last lines
			if strings.Contains(cmdLower, "tail") {
				score += 15.0
			}
			
			if strings.Contains(descLower, "split") && strings.Contains(descLower, "lines") {
				score += 15.0
			}
			if strings.Contains(descLower, "sort") {
				score += 12.0
			}
			if strings.Contains(descLower, "first") && strings.Contains(descLower, "lines") {
				score += 15.0
			}
			
			return score
		},
		
		"pattern_matching": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// grep for pattern matching
			if strings.Contains(cmdLower, "grep") {
				score += 20.0
			}
			if strings.Contains(descLower, "regex") || strings.Contains(descLower, "pattern") {
				score += 15.0
			}
			if strings.Contains(descLower, "match") {
				score += 12.0
			}
			
			return score
		},
		
		"output_redirect": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// tee for output redirection
			if strings.Contains(cmdLower, "tee") {
				score += 20.0
			}
			if strings.Contains(descLower, "standard output") || strings.Contains(descLower, "stdout") {
				score += 15.0
			}
			if strings.Contains(descLower, "terminal") && strings.Contains(descLower, "file") {
				score += 12.0
			}
			
			return score
		},
		
		"compress": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			// High priority matches
			if strings.Contains(cmdLower, "tar") && (strings.Contains(cmdLower, "czf") || strings.Contains(cmdLower, "czvf")) {
				score += 15.0 // tar create commands
			}
			if strings.Contains(cmdLower, "zip") && !strings.Contains(cmdLower, "unzip") {
				score += 12.0 // zip commands
			}
			if strings.Contains(cmdLower, "gzip") && !strings.Contains(cmdLower, "gunzip") {
				score += 10.0 // gzip commands
			}
			
			// Medium priority matches
			if strings.Contains(descLower, "compress") || strings.Contains(descLower, "archive") {
				score += 8.0
			}
			if strings.Contains(cmdLower, "7z") && strings.Contains(cmdLower, "a") {
				score += 10.0 // 7zip archive
			}
			
			// Keyword matches
			for _, keyword := range cmd.KeywordsLower {
				if keyword == "compress" || keyword == "archive" || keyword == "zip" {
					score += 6.0
				}
			}
			
			return score
		},
		
		"extract": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.Contains(cmdLower, "tar") && (strings.Contains(cmdLower, "xzf") || strings.Contains(cmdLower, "xvf")) {
				score += 15.0
			}
			if strings.Contains(cmdLower, "unzip") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "gunzip") {
				score += 10.0
			}
			if strings.Contains(descLower, "extract") || strings.Contains(descLower, "decompress") {
				score += 8.0
			}
			
			return score
		},
		
		"create": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.Contains(cmdLower, "mkdir") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "touch") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "create") {
				score += 10.0
			}
			if strings.Contains(descLower, "create") {
				score += 8.0
			}
			
			return score
		},
		
		"delete": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if cmdLower == "rm" || strings.HasPrefix(cmdLower, "rm ") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "del") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "remove") {
				score += 10.0
			}
			if strings.Contains(descLower, "delete") || strings.Contains(descLower, "remove") {
				score += 8.0
			}
			
			return score
		},
		
		"list": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if cmdLower == "ls" || strings.HasPrefix(cmdLower, "ls ") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "dir") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "list") {
				score += 10.0
			}
			if strings.Contains(descLower, "list") {
				score += 8.0
			}
			
			return score
		},
		
		"find": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.HasPrefix(cmdLower, "find ") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "grep") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "locate") {
				score += 10.0
			}
			if strings.Contains(descLower, "find") || strings.Contains(descLower, "search") {
				score += 8.0
			}
			
			return score
		},
		
		"download": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.Contains(cmdLower, "wget") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "curl") {
				score += 12.0
			}
			if strings.Contains(descLower, "download") {
				score += 10.0
			}
			
			return score
		},
		
		"backup": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.Contains(cmdLower, "mysqldump") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "rsync") {
				score += 12.0
			}
			if strings.Contains(descLower, "backup") {
				score += 10.0
			}
			
			return score
		},
		
		"monitor": func(cmd *database.Command) float64 {
			score := 0.0
			cmdLower := cmd.CommandLower
			descLower := cmd.DescriptionLower
			
			if strings.Contains(cmdLower, "top") || strings.Contains(cmdLower, "htop") {
				score += 15.0
			}
			if strings.Contains(cmdLower, "ps") {
				score += 12.0
			}
			if strings.Contains(cmdLower, "netstat") {
				score += 10.0
			}
			if strings.Contains(descLower, "monitor") || strings.Contains(descLower, "watch") {
				score += 8.0
			}
			
			return score
		},
	}

	// Apply the appropriate matcher
	if matcher, exists := intentMatchers[intent]; exists {
		for i := range es.db.Commands {
			cmd := &es.db.Commands[i]
			score := matcher(cmd)
			
			if score > 0 {
				// Apply confidence weighting
				finalScore := score * confidence
				
				results = append(results, SearchResult{
					Command:     cmd,
					Score:       finalScore,
					MatchReason: "intent: " + intent,
					Distance:    -1,
				})
			}
		}
	}

	return results
}

// deduplicateAndSort removes duplicates and sorts results by score
func (es *EnhancedSearcher) deduplicateAndSort(results []SearchResult, limit int) []SearchResult {
	seen := make(map[string]*SearchResult)
	
	// Deduplicate, keeping the highest score for each command
	for _, result := range results {
		key := result.Command.Command + "|" + result.Command.Description
		if existing, exists := seen[key]; exists {
			if result.Score > existing.Score {
				seen[key] = &result
			}
		} else {
			seen[key] = &result
		}
	}

	// Convert back to slice
	var deduplicated []SearchResult
	for _, result := range seen {
		deduplicated = append(deduplicated, *result)
	}

	// Sort by score (descending)
	sort.Slice(deduplicated, func(i, j int) bool {
		return deduplicated[i].Score > deduplicated[j].Score
	})

	// Apply limit
	if len(deduplicated) > limit {
		deduplicated = deduplicated[:limit]
	}

	return deduplicated
}

// deduplicateAndSortWithRanking provides enhanced ranking with query relevance
func (es *EnhancedSearcher) deduplicateAndSortWithRanking(results []SearchResult, limit int, originalQuery string) []SearchResult {
	seen := make(map[string]*SearchResult)
	
	// Deduplicate, keeping the highest score for each command
	for _, result := range results {
		key := result.Command.Command + "|" + result.Command.Description
		if existing, exists := seen[key]; exists {
			if result.Score > existing.Score {
				seen[key] = &result
			}
		} else {
			seen[key] = &result
		}
	}

	// Convert back to slice and apply enhanced scoring
	var deduplicated []SearchResult
	for _, result := range seen {
		enhancedResult := *result
		enhancedResult.Score = es.calculateEnhancedScore(enhancedResult, originalQuery)
		deduplicated = append(deduplicated, enhancedResult)
	}

	// Sort by enhanced score (descending)
	sort.Slice(deduplicated, func(i, j int) bool {
		// Primary sort by score
		if deduplicated[i].Score != deduplicated[j].Score {
			return deduplicated[i].Score > deduplicated[j].Score
		}
		// Secondary sort by command length (shorter commands often more relevant)
		return len(deduplicated[i].Command.Command) < len(deduplicated[j].Command.Command)
	})

	// Apply limit
	if len(deduplicated) > limit {
		deduplicated = deduplicated[:limit]
	}

	return deduplicated
}

// calculateEnhancedScore applies additional ranking factors
func (es *EnhancedSearcher) calculateEnhancedScore(result SearchResult, originalQuery string) float64 {
	score := result.Score
	cmd := result.Command
	queryLower := strings.ToLower(originalQuery)
	
	// Boost for exact command name matches
	if strings.Contains(queryLower, cmd.CommandLower) {
		score *= 1.2
	}
	
	// Boost for popular/common commands
	commonCommands := map[string]float64{
		"ls": 1.1, "cd": 1.1, "pwd": 1.1, "mkdir": 1.1, "rm": 1.1, "cp": 1.1, "mv": 1.1,
		"find": 1.1, "grep": 1.1, "tar": 1.1, "zip": 1.1, "git": 1.1, "ssh": 1.1,
		"wget": 1.1, "curl": 1.1, "ps": 1.1, "top": 1.1, "kill": 1.1, "chmod": 1.1,
	}
	
	cmdName := strings.Fields(cmd.CommandLower)[0] // Get first word of command
	if boost, exists := commonCommands[cmdName]; exists {
		score *= boost
	}
	
	// Boost for commands with good descriptions
	if len(cmd.Description) > 20 && len(cmd.Description) < 100 {
		score *= 1.05 // Prefer commands with informative but not overly long descriptions
	}
	
	// Boost for commands with categories
	if cmd.Niche != "" {
		score *= 1.03
	}
	
	// Penalize very long commands (often less useful)
	if len(cmd.Command) > 80 {
		score *= 0.9
	}
	
	return score
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateDynamicSuggestions creates intelligent suggestions without hardcoded dictionaries
func (es *EnhancedSearcher) GenerateDynamicSuggestions(query string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}

	// Initialize dynamic components
	semanticSearcher := NewSemanticSearcher(es.db)
	patternLearner := NewPatternLearner(es.db)
	
	var suggestions []string
	suggestionSet := make(map[string]bool)
	
	// Strategy 1: Semantic-based typo correction
	queryWords := strings.Fields(strings.ToLower(query))
	for _, word := range queryWords {
		if len(word) > 2 {
			semanticSuggestions := semanticSearcher.DynamicTypoCorrection(word, 2)
			for _, suggestion := range semanticSuggestions {
				if !suggestionSet[suggestion] && len(suggestions) < maxSuggestions {
					suggestions = append(suggestions, suggestion)
					suggestionSet[suggestion] = true
				}
			}
		}
	}
	
	// Strategy 2: Pattern-based query expansion
	if len(suggestions) < maxSuggestions {
		expandedTerms := patternLearner.ExpandQuery(query)
		for _, term := range expandedTerms {
			if term != strings.ToLower(query) && !suggestionSet[term] && len(suggestions) < maxSuggestions {
				suggestions = append(suggestions, term)
				suggestionSet[term] = true
			}
		}
	}
	
	// Strategy 3: Similar command suggestions
	if len(suggestions) < maxSuggestions {
		// Find the most likely command the user was looking for
		partialResults := es.semanticFuzzySearch(query, semanticSearcher)
		if len(partialResults) > 0 {
			bestMatch := partialResults[0]
			cmdName := strings.Fields(strings.ToLower(bestMatch.Command.Command))[0]
			similarCommands := patternLearner.SuggestSimilarCommands(cmdName, maxSuggestions-len(suggestions))
			
			for _, similar := range similarCommands {
				if !suggestionSet[similar] && len(suggestions) < maxSuggestions {
					suggestions = append(suggestions, similar)
					suggestionSet[similar] = true
				}
			}
		}
	}
	
	// Strategy 4: Intent-based suggestions
	if len(suggestions) < maxSuggestions {
		intents := semanticSearcher.AnalyzeQueryIntent(query)
		for intent, confidence := range intents {
			if confidence > 0.3 && !suggestionSet[intent] && len(suggestions) < maxSuggestions {
				// Suggest the intent as a simpler query
				suggestions = append(suggestions, intent)
				suggestionSet[intent] = true
			}
		}
	}

	return suggestions
}

// GenerateSuggestions maintains backward compatibility
func (es *EnhancedSearcher) GenerateSuggestions(query string, maxSuggestions int) []string {
	return es.GenerateDynamicSuggestions(query, maxSuggestions)
}

// findSimilarCommands finds commands with names similar to the query
func (es *EnhancedSearcher) findSimilarCommands(query string, maxSuggestions int) []string {
	var suggestions []string
	var matches []FuzzyMatch
	
	// Check against command names
	for _, cmd := range es.db.Commands {
		cmdName := strings.Fields(cmd.CommandLower)[0] // Get first word of command
		if len(cmdName) < 2 {
			continue
		}
		
		distance := LevenshteinDistance(query, cmdName)
		maxLen := max(len(query), len(cmdName))
		
		// More lenient threshold for command names
		if distance <= maxLen/2 && distance <= 4 {
			similarity := 1.0 - float64(distance)/float64(maxLen)
			matches = append(matches, FuzzyMatch{
				Text:     cmdName,
				Score:    similarity,
				Distance: distance,
			})
		}
	}
	
	// Sort by similarity
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	
	// Add unique suggestions
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(suggestions) >= maxSuggestions {
			break
		}
		if !seen[match.Text] {
			suggestions = append(suggestions, match.Text)
			seen[match.Text] = true
		}
	}
	
	return suggestions
}

// findDescriptionBasedSuggestions suggests based on description content
func (es *EnhancedSearcher) findDescriptionBasedSuggestions(queryWords []string, maxSuggestions int) []string {
	var suggestions []string
	wordFreq := make(map[string]int)
	
	// Find common words in descriptions that match query intent
	for _, cmd := range es.db.Commands {
		descWords := strings.Fields(cmd.DescriptionLower)
		for _, descWord := range descWords {
			if len(descWord) > 3 { // Only consider meaningful words
				for _, queryWord := range queryWords {
					if strings.Contains(descWord, queryWord) || strings.Contains(queryWord, descWord) {
						wordFreq[descWord]++
					}
				}
			}
		}
	}
	
	// Sort by frequency and add suggestions
	type wordCount struct {
		word  string
		count int
	}
	
	var wordCounts []wordCount
	for word, count := range wordFreq {
		if count > 1 { // Only suggest words that appear multiple times
			wordCounts = append(wordCounts, wordCount{word, count})
		}
	}
	
	sort.Slice(wordCounts, func(i, j int) bool {
		return wordCounts[i].count > wordCounts[j].count
	})
	
	for i, wc := range wordCounts {
		if i >= maxSuggestions {
			break
		}
		suggestions = append(suggestions, wc.word)
	}
	
	return suggestions
}

// generateSimplifiedSuggestions creates simpler query alternatives
func (es *EnhancedSearcher) generateSimplifiedSuggestions(query string) []string {
	var suggestions []string
	
	// Map complex phrases to simpler alternatives
	simplifications := map[string][]string{
		"how do i": {""},
		"how to": {""},
		"i want to": {""},
		"i need to": {""},
		"help me": {""},
		"multiple files": {"files"},
		"single archive": {"archive"},
		"into a": {""},
		"from a": {""},
		"with a": {""},
	}
	
	simplified := query
	for complex, simples := range simplifications {
		if strings.Contains(simplified, complex) {
			for _, simple := range simples {
				newQuery := strings.ReplaceAll(simplified, complex, simple)
				newQuery = strings.TrimSpace(strings.Join(strings.Fields(newQuery), " ")) // Clean up spaces
				if newQuery != query && newQuery != "" {
					suggestions = append(suggestions, newQuery)
				}
			}
		}
	}
	
	// Add common command categories as suggestions
	if len(suggestions) == 0 {
		categoryKeywords := []string{"compress", "extract", "copy", "move", "find", "list", "create", "delete"}
		for _, keyword := range categoryKeywords {
			if strings.Contains(query, keyword) {
				suggestions = append(suggestions, keyword)
				break
			}
		}
	}
	
	return suggestions
}