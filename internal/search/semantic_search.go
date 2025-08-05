// Package search provides semantic search capabilities using dynamic similarity algorithms
package search

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/Vedant9500/WTF/internal/database"
)

// SemanticSearcher provides intelligent search without hardcoded dictionaries
type SemanticSearcher struct {
	db           *database.Database
	wordVectors  map[string][]float64
	commandIndex map[string]*database.Command
}

// NewSemanticSearcher creates a new semantic searcher
func NewSemanticSearcher(db *database.Database) *SemanticSearcher {
	ss := &SemanticSearcher{
		db:           db,
		wordVectors:  make(map[string][]float64),
		commandIndex: make(map[string]*database.Command),
	}
	
	// Build dynamic word vectors from the database itself
	ss.buildDynamicWordVectors()
	ss.buildCommandIndex()
	
	return ss
}

// buildDynamicWordVectors creates word embeddings from command descriptions and keywords
func (ss *SemanticSearcher) buildDynamicWordVectors() {
	// Extract all words from commands, descriptions, and keywords
	wordFreq := make(map[string]int)
	wordCooccurrence := make(map[string]map[string]int)
	
	for _, cmd := range ss.db.Commands {
		// Tokenize all text associated with the command
		allText := strings.Join([]string{
			cmd.Command,
			cmd.Description,
			strings.Join(cmd.Keywords, " "),
			strings.Join(cmd.Tags, " "),
		}, " ")
		
		words := ss.tokenize(allText)
		
		// Count word frequencies
		for _, word := range words {
			wordFreq[word]++
			if wordCooccurrence[word] == nil {
				wordCooccurrence[word] = make(map[string]int)
			}
		}
		
		// Count co-occurrences (words appearing together)
		for i, word1 := range words {
			for j, word2 := range words {
				if i != j && len(word1) > 2 && len(word2) > 2 {
					wordCooccurrence[word1][word2]++
				}
			}
		}
	}
	
	// Build simple word vectors based on co-occurrence patterns
	vectorSize := 50
	for word := range wordFreq {
		if wordFreq[word] < 2 { // Skip rare words
			continue
		}
		
		vector := make([]float64, vectorSize)
		
		// Create vector based on co-occurrence patterns
		cooccurWords := wordCooccurrence[word]
		i := 0
		for _, count := range cooccurWords {
			if i >= vectorSize {
				break
			}
			// Normalize by frequency
			vector[i] = float64(count) / float64(wordFreq[word])
			i++
		}
		
		ss.wordVectors[word] = vector
	}
}

// buildCommandIndex creates an index for fast command lookup
func (ss *SemanticSearcher) buildCommandIndex() {
	for i := range ss.db.Commands {
		cmd := &ss.db.Commands[i]
		// Index by command name and major keywords
		ss.commandIndex[strings.ToLower(cmd.Command)] = cmd
		
		for _, keyword := range cmd.Keywords {
			if len(keyword) > 3 {
				ss.commandIndex[strings.ToLower(keyword)] = cmd
			}
		}
	}
}

// tokenize breaks text into meaningful tokens
func (ss *SemanticSearcher) tokenize(text string) []string {
	text = strings.ToLower(text)
	
	// Split on non-alphanumeric characters
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
	
	// Filter out very short words and common stop words
	var filtered []string
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	}
	
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}
	
	return filtered
}

// calculateSemanticSimilarity computes similarity between two words using vectors
func (ss *SemanticSearcher) calculateSemanticSimilarity(word1, word2 string) float64 {
	vec1, exists1 := ss.wordVectors[word1]
	vec2, exists2 := ss.wordVectors[word2]
	
	if !exists1 || !exists2 {
		// Fallback to string similarity
		return ss.calculateStringSimilarity(word1, word2)
	}
	
	// Calculate cosine similarity
	return ss.cosineSimilarity(vec1, vec2)
}

// cosineSimilarity calculates cosine similarity between two vectors
func (ss *SemanticSearcher) cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}
	
	var dotProduct, norm1, norm2 float64
	
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}
	
	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// calculateStringSimilarity provides fallback string similarity
func (ss *SemanticSearcher) calculateStringSimilarity(word1, word2 string) float64 {
	// Jaccard similarity based on character n-grams
	ngrams1 := ss.getNGrams(word1, 2)
	ngrams2 := ss.getNGrams(word2, 2)
	
	intersection := 0
	union := make(map[string]bool)
	
	// Add all n-grams to union
	for ngram := range ngrams1 {
		union[ngram] = true
	}
	for ngram := range ngrams2 {
		union[ngram] = true
	}
	
	// Count intersection
	for ngram := range ngrams1 {
		if ngrams2[ngram] {
			intersection++
		}
	}
	
	if len(union) == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(len(union))
}

// getNGrams generates character n-grams for a word
func (ss *SemanticSearcher) getNGrams(word string, n int) map[string]bool {
	ngrams := make(map[string]bool)
	
	if len(word) < n {
		ngrams[word] = true
		return ngrams
	}
	
	for i := 0; i <= len(word)-n; i++ {
		ngrams[word[i:i+n]] = true
	}
	
	return ngrams
}

// SemanticSearch performs intelligent search using semantic similarity
func (ss *SemanticSearcher) SemanticSearch(query string, limit int) []SearchResult {
	if limit <= 0 {
		limit = 5
	}
	
	queryWords := ss.tokenize(query)
	if len(queryWords) == 0 {
		return []SearchResult{}
	}
	
	var results []SearchResult
	
	// Score each command based on semantic similarity
	for i := range ss.db.Commands {
		cmd := &ss.db.Commands[i]
		score := ss.calculateCommandScore(cmd, queryWords)
		
		if score > 0.1 { // Minimum threshold
			results = append(results, SearchResult{
				Command:     cmd,
				Score:       score,
				MatchReason: "semantic similarity",
				Distance:    -1,
			})
		}
	}
	
	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Apply limit
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// calculateCommandScore computes semantic similarity score for a command
func (ss *SemanticSearcher) calculateCommandScore(cmd *database.Command, queryWords []string) float64 {
	// Get all words associated with the command
	cmdWords := ss.tokenize(strings.Join([]string{
		cmd.Command,
		cmd.Description,
		strings.Join(cmd.Keywords, " "),
		strings.Join(cmd.Tags, " "),
	}, " "))
	
	if len(cmdWords) == 0 {
		return 0.0
	}
	
	var totalScore float64
	matchCount := 0
	
	// For each query word, find the best matching command word
	for _, queryWord := range queryWords {
		bestSimilarity := 0.0
		
		for _, cmdWord := range cmdWords {
			similarity := ss.calculateSemanticSimilarity(queryWord, cmdWord)
			if similarity > bestSimilarity {
				bestSimilarity = similarity
			}
		}
		
		if bestSimilarity > 0.3 { // Similarity threshold
			totalScore += bestSimilarity
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	// Normalize by number of query words and boost for multiple matches
	normalizedScore := totalScore / float64(len(queryWords))
	
	// Boost commands that match more query words
	coverageBoost := float64(matchCount) / float64(len(queryWords))
	
	return normalizedScore * (1.0 + coverageBoost)
}

// DynamicTypoCorrection suggests corrections without hardcoded dictionaries
func (ss *SemanticSearcher) DynamicTypoCorrection(word string, maxSuggestions int) []string {
	if maxSuggestions <= 0 {
		maxSuggestions = 3
	}
	
	type suggestion struct {
		word  string
		score float64
	}
	
	var suggestions []suggestion
	
	// Check against all words in our vocabulary
	for vocabWord := range ss.wordVectors {
		if len(vocabWord) < 2 {
			continue
		}
		
		// Calculate edit distance and similarity
		editDist := LevenshteinDistance(word, vocabWord)
		maxLen := len(word)
		if len(vocabWord) > maxLen {
			maxLen = len(vocabWord)
		}
		
		// Only consider reasonable edit distances
		if editDist <= maxLen/2 && editDist <= 3 {
			// Combine edit distance with semantic similarity
			editSimilarity := 1.0 - float64(editDist)/float64(maxLen)
			semanticSimilarity := ss.calculateSemanticSimilarity(word, vocabWord)
			
			// Weighted combination
			combinedScore := 0.7*editSimilarity + 0.3*semanticSimilarity
			
			suggestions = append(suggestions, suggestion{
				word:  vocabWord,
				score: combinedScore,
			})
		}
	}
	
	// Sort by score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].score > suggestions[j].score
	})
	
	// Return top suggestions
	var result []string
	for i, s := range suggestions {
		if i >= maxSuggestions {
			break
		}
		result = append(result, s.word)
	}
	
	return result
}

// AnalyzeQueryIntent dynamically determines query intent without hardcoded patterns
func (ss *SemanticSearcher) AnalyzeQueryIntent(query string) map[string]float64 {
	queryWords := ss.tokenize(query)
	intents := make(map[string]float64)
	
	// Define intent seed words (minimal set)
	intentSeeds := map[string][]string{
		"file_operation": {"file", "directory", "folder"},
		"text_processing": {"text", "line", "word", "character"},
		"network": {"download", "upload", "url", "http"},
		"system": {"process", "system", "monitor"},
		"conversion": {"convert", "transform", "format"},
		"search": {"find", "search", "locate", "grep"},
	}
	
	// Calculate intent scores based on semantic similarity to seed words
	for intent, seeds := range intentSeeds {
		var intentScore float64
		
		for _, queryWord := range queryWords {
			bestSeedSimilarity := 0.0
			
			for _, seed := range seeds {
				similarity := ss.calculateSemanticSimilarity(queryWord, seed)
				if similarity > bestSeedSimilarity {
					bestSeedSimilarity = similarity
				}
			}
			
			intentScore += bestSeedSimilarity
		}
		
		if intentScore > 0 {
			intents[intent] = intentScore / float64(len(queryWords))
		}
	}
	
	return intents
}