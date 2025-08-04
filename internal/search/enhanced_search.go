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

// EnhancedSearch performs comprehensive search with multiple strategies and smart ranking
func (es *EnhancedSearcher) EnhancedSearch(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 5
	}

	// Preprocess query to fix common typos and add synonyms
	correctedQuery := es.PreprocessQuery(query)
	originalQuery := strings.ToLower(query)
	
	var allResults []SearchResult

	// Strategy 1: Exact matching on corrected query (highest priority)
	exactResults := es.exactSearch(correctedQuery)
	for i := range exactResults {
		exactResults[i].Score *= 1.5 // Boost exact matches
	}
	allResults = append(allResults, exactResults...)

	// Strategy 2: Intent-based matching (high priority for natural language)
	intentResults := es.intentBasedSearch(originalQuery)
	for i := range intentResults {
		intentResults[i].Score *= 1.3 // Boost intent matches
	}
	allResults = append(allResults, intentResults...)

	// Strategy 3: Fuzzy matching on individual words
	fuzzyResults := es.fuzzyWordSearch(correctedQuery)
	allResults = append(allResults, fuzzyResults...)

	// Strategy 4: Partial matching with high tolerance
	partialResults := es.partialSearch(correctedQuery)
	for i := range partialResults {
		partialResults[i].Score *= 0.8 // Reduce partial match scores
	}
	allResults = append(allResults, partialResults...)

	// Strategy 5: Fallback search on original query if corrected query yields poor results
	if len(allResults) < limit {
		fallbackResults := es.exactSearch(originalQuery)
		for i := range fallbackResults {
			fallbackResults[i].Score *= 0.9 // Slightly reduce fallback scores
		}
		allResults = append(allResults, fallbackResults...)
	}

	// Deduplicate and sort by score with enhanced ranking
	return es.deduplicateAndSortWithRanking(allResults, limit, originalQuery)
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

// GenerateSuggestions creates intelligent "Did you mean?" suggestions for failed searches
func (es *EnhancedSearcher) GenerateSuggestions(query string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)
	
	// Strategy 1: Check for common typos first
	var suggestions []string
	suggestionSet := make(map[string]bool)
	
	// Check entire query for common typo patterns
	for typo, correction := range CommonTypos {
		if strings.Contains(queryLower, typo) {
			correctedQuery := strings.ReplaceAll(queryLower, typo, correction)
			if correctedQuery != queryLower && !suggestionSet[correctedQuery] {
				suggestions = append(suggestions, correctedQuery)
				suggestionSet[correctedQuery] = true
			}
		}
	}
	
	// Strategy 2: Find similar command names using Levenshtein distance
	commandSuggestions := es.findSimilarCommands(queryLower, maxSuggestions-len(suggestions))
	for _, suggestion := range commandSuggestions {
		if !suggestionSet[suggestion] {
			suggestions = append(suggestions, suggestion)
			suggestionSet[suggestion] = true
		}
	}
	
	// Strategy 3: Suggest based on partial matches in descriptions
	if len(suggestions) < maxSuggestions {
		descriptionSuggestions := es.findDescriptionBasedSuggestions(queryWords, maxSuggestions-len(suggestions))
		for _, suggestion := range descriptionSuggestions {
			if !suggestionSet[suggestion] {
				suggestions = append(suggestions, suggestion)
				suggestionSet[suggestion] = true
			}
		}
	}
	
	// Strategy 4: Suggest simpler alternatives
	if len(suggestions) < maxSuggestions {
		simplifiedSuggestions := es.generateSimplifiedSuggestions(queryLower)
		for _, suggestion := range simplifiedSuggestions {
			if !suggestionSet[suggestion] && len(suggestions) < maxSuggestions {
				suggestions = append(suggestions, suggestion)
				suggestionSet[suggestion] = true
			}
		}
	}

	return suggestions
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