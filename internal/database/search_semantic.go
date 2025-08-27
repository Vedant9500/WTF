package database

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// SemanticSearcher provides intelligent search capabilities beyond keyword matching
type SemanticSearcher struct {
	db *Database
}

// NewSemanticSearcher creates a new semantic searcher
func NewSemanticSearcher(db *Database) *SemanticSearcher {
	return &SemanticSearcher{db: db}
}

// SemanticSearchOptions extends SearchOptions with semantic-specific settings
type SemanticSearchOptions struct {
	SearchOptions
	SemanticThreshold float64 // Minimum semantic similarity score
	UseFuzzyFallback  bool    // Use fuzzy matching when exact matches fail
	UseIntentBoost    bool    // Boost results based on detected intent
}

// SmartSearch performs intelligent search using multiple strategies
func (ss *SemanticSearcher) SmartSearch(query string, options SemanticSearchOptions) []SearchResult {
	// Strategy 1: Always try the proven BM25F search first - this is our baseline
	results := ss.db.SearchUniversal(query, options.SearchOptions)

	// If BM25F found good results, return them (they're already well-scored)
	if len(results) > 0 {
		return results
	}

	// Strategy 2: If BM25F didn't find anything, try semantic search
	semanticResults := ss.semanticSearch(query, options)
	if len(semanticResults) > 0 {
		return semanticResults
	}

	// Strategy 3: Last resort - fuzzy matching for partial matches
	if options.UseFuzzyFallback {
		fuzzyResults := ss.fuzzySearch(query, options)
		return fuzzyResults
	}

	return nil
}

// semanticSearch finds commands based on semantic similarity rather than exact keyword matching
func (ss *SemanticSearcher) semanticSearch(query string, options SemanticSearchOptions) []SearchResult {
	queryWords := extractMeaningfulWords(query)
	if len(queryWords) == 0 {
		return nil
	}

	var results []SearchResult

	// Score each command based on semantic similarity
	for i, cmd := range ss.db.Commands {
		score := ss.calculateSemanticScore(queryWords, cmd, query)

		if score > options.SemanticThreshold {
			// Apply platform filtering
			if len(cmd.Platform) > 0 {
				currentPlatform := getCurrentPlatform()
				if !isPlatformCompatible(cmd.Platform, currentPlatform) && !isCrossPlatformTool(cmd.Command) {
					continue
				}
			}

			results = append(results, SearchResult{
				Command: &ss.db.Commands[i],
				Score:   score,
			})
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > options.Limit {
		results = results[:options.Limit]
	}

	return results
}

// calculateSemanticScore computes a semantic similarity score between query and command
func (ss *SemanticSearcher) calculateSemanticScore(queryWords []string, cmd Command, originalQuery string) float64 {
	var score float64

	// Create searchable text from command
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description)
	keywordText := strings.ToLower(strings.Join(cmd.Keywords, " "))
	fullText := cmdText + " " + keywordText

	// 1. Direct word overlap (highest weight)
	directOverlap := calculateWordOverlap(queryWords, strings.Fields(fullText))
	score += directOverlap * 100.0

	// 2. Semantic word similarity (medium weight)
	semanticSim := calculateSemanticSimilarity(queryWords, fullText)
	score += semanticSim * 50.0

	// 3. Intent matching (medium weight)
	intentScore := calculateIntentScore(originalQuery, cmd)
	score += intentScore * 40.0

	// 4. Contextual relevance (lower weight)
	contextScore := calculateContextualRelevance(originalQuery, cmd)
	score += contextScore * 30.0

	// 5. Command popularity boost (very low weight)
	popularityBoost := calculatePopularityBoost(cmd)
	score += popularityBoost * 10.0

	return score
}

// calculateWordOverlap computes the overlap between query words and command text
func calculateWordOverlap(queryWords []string, textWords []string) float64 {
	textWordsSet := make(map[string]bool)
	for _, word := range textWords {
		textWordsSet[word] = true
	}

	matches := 0
	for _, qword := range queryWords {
		if textWordsSet[qword] {
			matches++
		}
	}

	if len(queryWords) == 0 {
		return 0
	}

	return float64(matches) / float64(len(queryWords))
}

// calculateSemanticSimilarity uses simple but effective semantic similarity heuristics
func calculateSemanticSimilarity(queryWords []string, text string) float64 {
	var similarity float64

	// Define semantic word groups for common IT concepts
	semanticGroups := map[string][]string{
		"network": {"ip", "interface", "connection", "config", "configuration", "network", "adapter", "ethernet", "wifi", "dns", "dhcp"},
		"file":    {"file", "directory", "folder", "path", "document", "content", "data", "text", "read", "write"},
		"process": {"process", "service", "daemon", "task", "job", "running", "execute", "kill", "start", "stop"},
		"system":  {"system", "os", "windows", "linux", "mac", "computer", "machine", "hardware", "software"},
		"manage":  {"manage", "control", "configure", "setup", "modify", "change", "edit", "update", "admin"},
		"view":    {"view", "show", "display", "see", "look", "read", "cat", "list", "print", "output"},
		"search":  {"find", "search", "locate", "discover", "lookup", "query", "filter", "grep"},
	}

	// Check for semantic group matches
	for _, qword := range queryWords {
		for _, groupWords := range semanticGroups {
			// If query word is in a semantic group
			if contains(groupWords, qword) {
				// Check if text contains other words from the same group
				for _, groupWord := range groupWords {
					if strings.Contains(text, groupWord) && groupWord != qword {
						similarity += 0.3 // Semantic relationship found
					}
				}
			}
		}
	}

	return math.Min(similarity, 1.0)
}

// calculateIntentScore boosts commands that match the detected intent
func calculateIntentScore(query string, cmd Command) float64 {
	queryLower := strings.ToLower(query)
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description)

	var score float64

	// Intent patterns and their associated keywords
	intentPatterns := map[string][]string{
		"view":    {"what is", "show me", "display", "see", "view", "read", "cat", "look at"},
		"find":    {"find", "search for", "locate", "where is", "which command", "what command"},
		"manage":  {"manage", "control", "configure", "setup", "modify", "change"},
		"network": {"ip", "network", "interface", "connection", "dns", "dhcp", "ping"},
		"windows": {"windows", "win", "cmd", "powershell", "dos"},
	}

	for _, patterns := range intentPatterns {
		queryMatches := 0
		cmdMatches := 0

		for _, pattern := range patterns {
			if strings.Contains(queryLower, pattern) {
				queryMatches++
			}
			if strings.Contains(cmdText, pattern) {
				cmdMatches++
			}
		}

		if queryMatches > 0 && cmdMatches > 0 {
			score += float64(queryMatches*cmdMatches) * 0.2
		}
	}

	return math.Min(score, 1.0)
}

// calculateContextualRelevance uses contextual clues to boost relevance
func calculateContextualRelevance(query string, cmd Command) float64 {
	queryLower := strings.ToLower(query)
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description)

	var score float64

	// Platform context
	if strings.Contains(queryLower, "windows") && strings.Contains(cmdText, "windows") {
		score += 0.3
	}

	// Question patterns
	questionPatterns := []string{"what is", "how to", "which command", "what command"}
	for _, pattern := range questionPatterns {
		if strings.Contains(queryLower, pattern) {
			score += 0.2
			break
		}
	}

	// Command specificity
	if strings.Contains(queryLower, "command") && strings.Contains(cmdText, cmd.Command) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// calculatePopularityBoost gives slight boost to commonly used commands
func calculatePopularityBoost(cmd Command) float64 {
	commonCommands := map[string]float64{
		"ls":       0.9,
		"cd":       0.9,
		"cat":      0.8,
		"grep":     0.8,
		"find":     0.8,
		"ipconfig": 0.7,
		"ping":     0.7,
		"ps":       0.7,
		"top":      0.6,
		"chmod":    0.6,
	}

	if boost, exists := commonCommands[strings.ToLower(cmd.Command)]; exists {
		return boost * 0.1 // Very small boost
	}

	return 0
}

// fuzzySearch provides fuzzy matching for partial word matches
func (ss *SemanticSearcher) fuzzySearch(query string, options SemanticSearchOptions) []SearchResult {
	queryWords := extractMeaningfulWords(query)
	if len(queryWords) == 0 {
		return nil
	}

	var results []SearchResult

	for i, cmd := range ss.db.Commands {
		score := ss.calculateFuzzyScore(queryWords, cmd)

		if score > 0.3 { // Lower threshold for fuzzy matching
			results = append(results, SearchResult{
				Command: &ss.db.Commands[i],
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

// calculateFuzzyScore computes fuzzy similarity score
func (ss *SemanticSearcher) calculateFuzzyScore(queryWords []string, cmd Command) float64 {
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))

	var totalScore float64

	for _, qword := range queryWords {
		bestMatch := 0.0
		for _, word := range strings.Fields(cmdText) {
			similarity := calculateLevenshteinSimilarity(qword, word)
			if similarity > bestMatch {
				bestMatch = similarity
			}
		}
		totalScore += bestMatch
	}

	return totalScore / float64(len(queryWords))
}

// Helper functions
func extractMeaningfulWords(query string) []string {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true,
		"by": true, "for": true, "from": true, "has": true, "he": true, "in": true, "is": true,
		"it": true, "its": true, "of": true, "on": true, "that": true, "the": true, "to": true,
		"was": true, "will": true, "with": true, "what": true, "how": true, "which": true,
		"where": true, "when": true, "why": true, "who": true,
	}

	words := strings.FieldsFunc(strings.ToLower(query), func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	var meaningful []string
	for _, word := range words {
		if len(word) > 1 && !stopWords[word] {
			meaningful = append(meaningful, word)
		}
	}

	return meaningful
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func calculateLevenshteinSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple implementation - for better results, use a proper Levenshtein distance library
	if strings.Contains(s2, s1) || strings.Contains(s1, s2) {
		return 0.8
	}
	if strings.HasPrefix(s2, s1) || strings.HasPrefix(s1, s2) {
		return 0.6
	}
	if strings.HasSuffix(s2, s1) || strings.HasSuffix(s1, s2) {
		return 0.4
	}

	return 0.0
}
