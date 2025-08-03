// Package search provides enhanced search capabilities with better fuzzy matching and NLP
package search

import (
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

// EnhancedSearcher provides improved search capabilities
type EnhancedSearcher struct {
	db *database.Database
}

// NewEnhancedSearcher creates a new enhanced searcher
func NewEnhancedSearcher(db *database.Database) *EnhancedSearcher {
	return &EnhancedSearcher{db: db}
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

// PreprocessQuery cleans and corrects common typos in the query
func (es *EnhancedSearcher) PreprocessQuery(query string) string {
	words := strings.Fields(strings.ToLower(query))
	correctedWords := make([]string, len(words))

	for i, word := range words {
		// Remove punctuation
		cleanWord := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}, word)

		// Check for common typos
		if correction, exists := CommonTypos[cleanWord]; exists {
			correctedWords[i] = correction
		} else {
			correctedWords[i] = cleanWord
		}
	}

	return strings.Join(correctedWords, " ")
}

// SearchResult represents an enhanced search result
type SearchResult struct {
	Command     *database.Command
	Score       float64
	MatchReason string
	Distance    int
}

// EnhancedSearch performs comprehensive search with multiple strategies
func (es *EnhancedSearcher) EnhancedSearch(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 5
	}

	// Preprocess query to fix common typos
	correctedQuery := es.PreprocessQuery(query)
	
	var allResults []SearchResult

	// Strategy 1: Exact matching on corrected query
	exactResults := es.exactSearch(correctedQuery)
	allResults = append(allResults, exactResults...)

	// Strategy 2: Fuzzy matching on individual words
	fuzzyResults := es.fuzzyWordSearch(correctedQuery)
	allResults = append(allResults, fuzzyResults...)

	// Strategy 3: Partial matching with high tolerance
	partialResults := es.partialSearch(correctedQuery)
	allResults = append(allResults, partialResults...)

	// Strategy 4: Intent-based matching
	intentResults := es.intentBasedSearch(correctedQuery)
	allResults = append(allResults, intentResults...)

	// Deduplicate and sort by score
	return es.deduplicateAndSort(allResults, limit)
}

// exactSearch performs exact string matching
func (es *EnhancedSearcher) exactSearch(query string) []SearchResult {
	var results []SearchResult
	queryWords := strings.Fields(query)

	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		score := 0.0
		matchReasons := []string{}

		// Check command name
		for _, word := range queryWords {
			if strings.Contains(cmd.CommandLower, word) {
				if cmd.CommandLower == word {
					score += 20.0
					matchReasons = append(matchReasons, "exact command")
				} else if strings.HasPrefix(cmd.CommandLower, word) {
					score += 15.0
					matchReasons = append(matchReasons, "command prefix")
				} else {
					score += 10.0
					matchReasons = append(matchReasons, "command contains")
				}
			}
		}

		// Check description
		for _, word := range queryWords {
			if strings.Contains(cmd.DescriptionLower, word) {
				score += 8.0
				matchReasons = append(matchReasons, "description")
			}
		}

		// Check keywords
		for _, word := range queryWords {
			for _, keyword := range cmd.KeywordsLower {
				if keyword == word {
					score += 12.0
					matchReasons = append(matchReasons, "exact keyword")
				} else if strings.Contains(keyword, word) {
					score += 6.0
					matchReasons = append(matchReasons, "keyword contains")
				}
			}
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

// intentBasedSearch performs intent-based matching
func (es *EnhancedSearcher) intentBasedSearch(query string) []SearchResult {
	var results []SearchResult
	
	// Define intent patterns
	intents := map[string][]string{
		"create": {"create", "make", "new", "mkdir", "touch"},
		"delete": {"delete", "remove", "rm", "del", "erase"},
		"copy":   {"copy", "cp", "duplicate", "clone"},
		"move":   {"move", "mv", "rename", "relocate"},
		"list":   {"list", "ls", "show", "display", "dir"},
		"find":   {"find", "search", "locate", "grep"},
		"edit":   {"edit", "modify", "change", "update", "vim", "nano"},
		"compress": {"compress", "zip", "tar", "archive", "gzip"},
		"extract":  {"extract", "unzip", "untar", "decompress"},
		"download": {"download", "fetch", "get", "wget", "curl"},
		"upload":   {"upload", "push", "send", "put"},
		"backup":   {"backup", "save", "export", "dump"},
		"restore":  {"restore", "import", "load", "recover"},
		"install":  {"install", "setup", "add", "mount"},
		"monitor":  {"monitor", "watch", "track", "observe"},
	}

	queryLower := strings.ToLower(query)
	
	for intent, keywords := range intents {
		for _, keyword := range keywords {
			if strings.Contains(queryLower, keyword) {
				// Find commands that match this intent
				intentResults := es.findCommandsByIntent(intent, keyword)
				results = append(results, intentResults...)
				break
			}
		}
	}

	return results
}

// findCommandsByIntent finds commands that match a specific intent
func (es *EnhancedSearcher) findCommandsByIntent(intent, keyword string) []SearchResult {
	var results []SearchResult

	for i := range es.db.Commands {
		cmd := &es.db.Commands[i]
		score := 0.0

		// Check if command matches the intent
		switch intent {
		case "create":
			if strings.Contains(cmd.CommandLower, "mkdir") || 
			   strings.Contains(cmd.CommandLower, "touch") ||
			   strings.Contains(cmd.CommandLower, "create") ||
			   strings.Contains(cmd.DescriptionLower, "create") {
				score = 8.0
			}
		case "delete":
			if strings.Contains(cmd.CommandLower, "rm") || 
			   strings.Contains(cmd.CommandLower, "del") ||
			   strings.Contains(cmd.CommandLower, "remove") ||
			   strings.Contains(cmd.DescriptionLower, "delete") ||
			   strings.Contains(cmd.DescriptionLower, "remove") {
				score = 8.0
			}
		case "copy":
			if strings.Contains(cmd.CommandLower, "cp") || 
			   strings.Contains(cmd.CommandLower, "copy") ||
			   strings.Contains(cmd.DescriptionLower, "copy") {
				score = 8.0
			}
		case "list":
			if strings.Contains(cmd.CommandLower, "ls") || 
			   strings.Contains(cmd.CommandLower, "dir") ||
			   strings.Contains(cmd.CommandLower, "list") ||
			   strings.Contains(cmd.DescriptionLower, "list") {
				score = 8.0
			}
		case "find":
			if strings.Contains(cmd.CommandLower, "find") || 
			   strings.Contains(cmd.CommandLower, "grep") ||
			   strings.Contains(cmd.CommandLower, "search") ||
			   strings.Contains(cmd.DescriptionLower, "find") ||
			   strings.Contains(cmd.DescriptionLower, "search") {
				score = 8.0
			}
		case "compress":
			if strings.Contains(cmd.CommandLower, "tar") || 
			   strings.Contains(cmd.CommandLower, "zip") ||
			   strings.Contains(cmd.CommandLower, "gzip") ||
			   strings.Contains(cmd.DescriptionLower, "compress") ||
			   strings.Contains(cmd.DescriptionLower, "archive") {
				score = 8.0
			}
		// Add more intent matching as needed
		}

		if score > 0 {
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       score,
				MatchReason: "intent: " + intent,
				Distance:    -1,
			})
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

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GenerateSuggestions creates "Did you mean?" suggestions for failed searches
func (es *EnhancedSearcher) GenerateSuggestions(query string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}

	queryWords := strings.Fields(strings.ToLower(query))
	var suggestions []string
	suggestionSet := make(map[string]bool)

	// Collect all command words and keywords
	var allWords []string
	for _, cmd := range es.db.Commands {
		// Add command words
		cmdWords := strings.Fields(cmd.CommandLower)
		allWords = append(allWords, cmdWords...)
		
		// Add keywords
		allWords = append(allWords, cmd.KeywordsLower...)
	}

	// Find close matches for each query word
	for _, queryWord := range queryWords {
		if len(queryWord) < 2 {
			continue
		}

		var bestMatches []FuzzyMatch
		for _, word := range allWords {
			if len(word) < 2 {
				continue
			}

			distance := LevenshteinDistance(queryWord, word)
			maxLen := max(len(queryWord), len(word))
			
			// Only suggest if the distance is reasonable
			if distance <= maxLen/2 && distance <= 3 {
				similarity := 1.0 - float64(distance)/float64(maxLen)
				bestMatches = append(bestMatches, FuzzyMatch{
					Text:     word,
					Score:    similarity,
					Distance: distance,
				})
			}
		}

		// Sort by similarity
		sort.Slice(bestMatches, func(i, j int) bool {
			return bestMatches[i].Score > bestMatches[j].Score
		})

		// Add top suggestions
		for i, match := range bestMatches {
			if i >= 2 { // Max 2 suggestions per word
				break
			}
			if !suggestionSet[match.Text] && match.Text != queryWord {
				suggestions = append(suggestions, match.Text)
				suggestionSet[match.Text] = true
			}
		}
	}

	// Limit total suggestions
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	return suggestions
}