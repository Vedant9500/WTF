// Package search provides fast, intelligent search without sacrificing accuracy
package search

import (
	"sort"
	"strings"
	"sync"

	"github.com/Vedant9500/WTF/internal/database"
)

// FastSearcher provides lightning-fast search with pre-computed indexes
type FastSearcher struct {
	db              *database.Database
	wordIndex       map[string][]int          // word -> command indices
	ngramIndex      map[string][]int          // n-gram -> command indices  
	commandWords    [][]string                // pre-tokenized command words
	descWords       [][]string                // pre-tokenized description words
	keywordWords    [][]string                // pre-tokenized keyword words
	mutex           sync.RWMutex
	initialized     bool
}

// NewFastSearcher creates a new fast searcher with pre-computed indexes
func NewFastSearcher(db *database.Database) *FastSearcher {
	fs := &FastSearcher{
		db:           db,
		wordIndex:    make(map[string][]int),
		ngramIndex:   make(map[string][]int),
		commandWords: make([][]string, len(db.Commands)),
		descWords:    make([][]string, len(db.Commands)),
		keywordWords: make([][]string, len(db.Commands)),
	}
	
	// Pre-compute all indexes (this happens once at startup)
	fs.buildIndexes()
	
	return fs
}

// buildIndexes pre-computes all search indexes for fast lookup
func (fs *FastSearcher) buildIndexes() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	// Pre-tokenize all text for each command
	for i, cmd := range fs.db.Commands {
		// Tokenize command
		fs.commandWords[i] = fs.fastTokenize(cmd.Command)
		
		// Tokenize description
		fs.descWords[i] = fs.fastTokenize(cmd.Description)
		
		// Tokenize keywords
		fs.keywordWords[i] = cmd.Keywords
		
		// Build word index
		allWords := append(fs.commandWords[i], fs.descWords[i]...)
		allWords = append(allWords, fs.keywordWords[i]...)
		
		for _, word := range allWords {
			if len(word) > 1 {
				wordLower := strings.ToLower(word)
				fs.wordIndex[wordLower] = append(fs.wordIndex[wordLower], i)
				
				// Build n-gram index for fuzzy matching
				for _, ngram := range fs.getNGrams(wordLower, 3) {
					fs.ngramIndex[ngram] = append(fs.ngramIndex[ngram], i)
				}
			}
		}
	}
	
	fs.initialized = true
}

// fastTokenize quickly tokenizes text into words
func (fs *FastSearcher) fastTokenize(text string) []string {
	// Simple but fast tokenization
	words := strings.FieldsFunc(strings.ToLower(text), func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	})
	
	// Filter out very short words
	var filtered []string
	for _, word := range words {
		if len(word) > 1 {
			filtered = append(filtered, word)
		}
	}
	
	return filtered
}

// getNGrams generates n-grams for fuzzy matching
func (fs *FastSearcher) getNGrams(word string, n int) []string {
	if len(word) < n {
		return []string{word}
	}
	
	var ngrams []string
	for i := 0; i <= len(word)-n; i++ {
		ngrams = append(ngrams, word[i:i+n])
	}
	
	return ngrams
}

// FastSearch performs lightning-fast search with good accuracy
func (fs *FastSearcher) FastSearch(query string, limit int) []SearchResult {
	if !fs.initialized {
		return []SearchResult{}
	}
	
	if limit <= 0 {
		limit = 5
	}
	
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	queryWords := fs.fastTokenize(query)
	if len(queryWords) == 0 {
		return []SearchResult{}
	}
	
	// Fast candidate collection using pre-computed indexes
	candidates := fs.getCandidates(queryWords)
	
	// Fast scoring of candidates
	var results []SearchResult
	for cmdIdx, score := range candidates {
		if score > 0.1 {
			results = append(results, SearchResult{
				Command:     &fs.db.Commands[cmdIdx],
				Score:       score,
				MatchReason: "fast search",
				Distance:    -1,
			})
		}
	}
	
	// Fast sort and limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// getCandidates quickly finds candidate commands using pre-computed indexes
func (fs *FastSearcher) getCandidates(queryWords []string) map[int]float64 {
	candidates := make(map[int]float64)
	
	for _, queryWord := range queryWords {
		queryLower := strings.ToLower(queryWord)
		
		// Exact word matches (highest priority)
		if cmdIndices, exists := fs.wordIndex[queryLower]; exists {
			for _, cmdIdx := range cmdIndices {
				candidates[cmdIdx] += 2.0
			}
		}
		
		// Fuzzy matches using n-grams (lower priority)
		queryNgrams := fs.getNGrams(queryLower, 3)
		ngramMatches := make(map[int]int)
		
		for _, ngram := range queryNgrams {
			if cmdIndices, exists := fs.ngramIndex[ngram]; exists {
				for _, cmdIdx := range cmdIndices {
					ngramMatches[cmdIdx]++
				}
			}
		}
		
		// Convert n-gram matches to scores
		for cmdIdx, ngramCount := range ngramMatches {
			// Score based on n-gram coverage
			coverage := float64(ngramCount) / float64(len(queryNgrams))
			if coverage > 0.3 { // Minimum threshold
				candidates[cmdIdx] += coverage * 0.8
			}
		}
	}
	
	// Boost commands that match multiple query words
	for cmdIdx, score := range candidates {
		matchedWords := 0
		for _, queryWord := range queryWords {
			if fs.commandMatchesWord(cmdIdx, queryWord) {
				matchedWords++
			}
		}
		
		if matchedWords > 1 {
			coverageBoost := float64(matchedWords) / float64(len(queryWords))
			candidates[cmdIdx] = score * (1.0 + coverageBoost)
		}
	}
	
	return candidates
}

// commandMatchesWord quickly checks if a command matches a word
func (fs *FastSearcher) commandMatchesWord(cmdIdx int, word string) bool {
	wordLower := strings.ToLower(word)
	
	// Check command words
	for _, cmdWord := range fs.commandWords[cmdIdx] {
		if strings.Contains(cmdWord, wordLower) {
			return true
		}
	}
	
	// Check description words
	for _, descWord := range fs.descWords[cmdIdx] {
		if strings.Contains(descWord, wordLower) {
			return true
		}
	}
	
	// Check keywords
	for _, keyword := range fs.keywordWords[cmdIdx] {
		if strings.Contains(strings.ToLower(keyword), wordLower) {
			return true
		}
	}
	
	return false
}

// FastTypoCorrection provides quick typo suggestions
func (fs *FastSearcher) FastTypoCorrection(word string, maxSuggestions int) []string {
	if !fs.initialized || maxSuggestions <= 0 {
		return []string{}
	}
	
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	wordLower := strings.ToLower(word)
	suggestions := make(map[string]int)
	
	// Use n-gram index for fast fuzzy matching
	wordNgrams := fs.getNGrams(wordLower, 3)
	
	for _, ngram := range wordNgrams {
		if cmdIndices, exists := fs.ngramIndex[ngram]; exists {
			// Get words from matching commands
			for _, cmdIdx := range cmdIndices {
				allWords := append(fs.commandWords[cmdIdx], fs.descWords[cmdIdx]...)
				for _, candidateWord := range allWords {
					if len(candidateWord) > 2 && candidateWord != wordLower {
						// Quick similarity check
						if fs.quickSimilarity(wordLower, candidateWord) > 0.5 {
							suggestions[candidateWord]++
						}
					}
				}
			}
		}
	}
	
	// Convert to sorted list
	type suggestion struct {
		word  string
		count int
	}
	
	var sortedSuggestions []suggestion
	for word, count := range suggestions {
		sortedSuggestions = append(sortedSuggestions, suggestion{word, count})
	}
	
	sort.Slice(sortedSuggestions, func(i, j int) bool {
		return sortedSuggestions[i].count > sortedSuggestions[j].count
	})
	
	// Return top suggestions
	var result []string
	for i, s := range sortedSuggestions {
		if i >= maxSuggestions {
			break
		}
		result = append(result, s.word)
	}
	
	return result
}

// quickSimilarity provides fast similarity calculation
func (fs *FastSearcher) quickSimilarity(word1, word2 string) float64 {
	if word1 == word2 {
		return 1.0
	}
	
	// Quick length check
	lenDiff := len(word1) - len(word2)
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	
	maxLen := len(word1)
	if len(word2) > maxLen {
		maxLen = len(word2)
	}
	
	if lenDiff > maxLen/2 {
		return 0.0
	}
	
	// Quick character overlap check
	chars1 := make(map[rune]int)
	chars2 := make(map[rune]int)
	
	for _, c := range word1 {
		chars1[c]++
	}
	for _, c := range word2 {
		chars2[c]++
	}
	
	overlap := 0
	total := 0
	
	for c, count1 := range chars1 {
		count2 := chars2[c]
		if count2 > 0 {
			if count1 < count2 {
				overlap += count1
			} else {
				overlap += count2
			}
		}
		total += count1
	}
	
	for _, count2 := range chars2 {
		total += count2
	}
	
	if total == 0 {
		return 0.0
	}
	
	return float64(overlap*2) / float64(total)
}

// SmartSearch combines fast exact matching with intelligent fallbacks
func (fs *FastSearcher) SmartSearch(query string, limit int) []SearchResult {
	// Try fast search first
	results := fs.FastSearch(query, limit)
	
	// If we got good results, return them
	if len(results) >= limit/2 && results[0].Score > 1.0 {
		return results
	}
	
	// Otherwise, try with expanded query
	queryWords := fs.fastTokenize(query)
	if len(queryWords) > 1 {
		// Try individual words
		for _, word := range queryWords {
			wordResults := fs.FastSearch(word, limit-len(results))
			for _, result := range wordResults {
				result.Score *= 0.7 // Reduce score for single-word matches
				results = append(results, result)
			}
		}
	}
	
	// Deduplicate and sort
	seen := make(map[string]bool)
	var deduped []SearchResult
	
	for _, result := range results {
		key := result.Command.Command
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, result)
		}
	}
	
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].Score > deduped[j].Score
	})
	
	if len(deduped) > limit {
		deduped = deduped[:limit]
	}
	
	return deduped
}