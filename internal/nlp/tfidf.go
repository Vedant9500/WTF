// Package nlp provides TF-IDF based natural language processing for command search
package nlp

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// Command represents a command for TF-IDF indexing
type Command struct {
	Command     string
	Description string
	Keywords    []string
}

// TFIDFSearcher implements TF-IDF based command search
type TFIDFSearcher struct {
	commands     []Command
	vocabulary   map[string]int    // word -> index mapping
	idf          []float64         // inverse document frequency for each word
	commandTF    []map[int]float64 // term frequency for each command
	commandNorms []float64         // document norms for cosine similarity
}

// NewTFIDFSearcher creates a new TF-IDF based searcher
func NewTFIDFSearcher(commands []Command) *TFIDFSearcher {
	searcher := &TFIDFSearcher{
		commands: commands,
	}
	searcher.buildIndex()
	return searcher
}

// TFIDFResult represents a search result with TF-IDF scoring
type TFIDFResult struct {
	CommandIndex int
	Score        float64
	Similarity   float64
}

// buildIndex creates the TF-IDF index from the command database
func (s *TFIDFSearcher) buildIndex() {
	// Step 1: Build vocabulary from all commands
	wordCounts := make(map[string]int)
	documents := make([][]string, len(s.commands))

	for i, cmd := range s.commands {
		// Combine command, description, and keywords into a single document
		text := strings.Join([]string{
			cmd.Command,
			cmd.Description,
			strings.Join(cmd.Keywords, " "),
		}, " ")

		words := s.tokenize(text)
		documents[i] = words

		// Count word occurrences across all documents
		wordSet := make(map[string]bool)
		for _, word := range words {
			if !wordSet[word] {
				wordCounts[word]++
				wordSet[word] = true
			}
		}
	}

	// Step 2: Build vocabulary index (only words that appear in multiple documents)
	s.vocabulary = make(map[string]int)
	vocabIndex := 0
	for word, docCount := range wordCounts {
		// Only include words that appear in at least 2 documents but not in more than 50% of documents
		if docCount >= 2 && docCount <= len(s.commands)/2 {
			s.vocabulary[word] = vocabIndex
			vocabIndex++
		}
	}

	// Step 3: Calculate IDF for each word
	s.idf = make([]float64, len(s.vocabulary))
	for word, idx := range s.vocabulary {
		docCount := wordCounts[word]
		s.idf[idx] = math.Log(float64(len(s.commands)) / float64(docCount))
	}

	// Step 4: Calculate TF for each command and document norms
	s.commandTF = make([]map[int]float64, len(s.commands))
	s.commandNorms = make([]float64, len(s.commands))

	for i, words := range documents {
		// Calculate term frequency
		termCounts := make(map[int]int)
		for _, word := range words {
			if idx, exists := s.vocabulary[word]; exists {
				termCounts[idx]++
			}
		}

		// Convert to TF-IDF and calculate norm
		s.commandTF[i] = make(map[int]float64)
		var norm float64

		for termIdx, count := range termCounts {
			tf := float64(count) / float64(len(words))
			tfidf := tf * s.idf[termIdx]
			s.commandTF[i][termIdx] = tfidf
			norm += tfidf * tfidf
		}

		s.commandNorms[i] = math.Sqrt(norm)
	}
}

// tokenize converts text into normalized tokens
func (s *TFIDFSearcher) tokenize(text string) []string {
	// Convert to lowercase and split into words
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	var tokens []string
	for _, word := range words {
		// Skip very short words and common stop words
		if len(word) >= 2 && !s.isStopWord(word) {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

// isStopWord checks if a word is a common stop word
func (s *TFIDFSearcher) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
		"you": true, "your": true, "all": true, "any": true, "can": true, "from": true,
		"not": true, "no": true, "if": true, "when": true, "where": true, "how": true,
		"what": true, "which": true, "who": true, "why": true, "use": true, "used": true,
		"using": true,
	}
	return stopWords[word]
}

// Search performs TF-IDF based search with cosine similarity
func (s *TFIDFSearcher) Search(query string, limit int) []TFIDFResult {
	// Tokenize query
	queryTokens := s.tokenize(query)
	if len(queryTokens) == 0 {
		return []TFIDFResult{}
	}

	// Build query vector
	queryVector := make(map[int]float64)
	queryTermCounts := make(map[int]int)

	for _, token := range queryTokens {
		if idx, exists := s.vocabulary[token]; exists {
			queryTermCounts[idx]++
		}
	}

	// Calculate query TF-IDF
	var queryNorm float64
	for termIdx, count := range queryTermCounts {
		tf := float64(count) / float64(len(queryTokens))
		tfidf := tf * s.idf[termIdx]
		queryVector[termIdx] = tfidf
		queryNorm += tfidf * tfidf
	}
	queryNorm = math.Sqrt(queryNorm)

	if queryNorm == 0 {
		return []TFIDFResult{}
	}

	// Calculate cosine similarity with each command
	var results []TFIDFResult
	for i := range s.commands {
		similarity := s.cosineSimilarity(queryVector, queryNorm, s.commandTF[i], s.commandNorms[i])

		if similarity > 0.01 { // Minimum similarity threshold
			results = append(results, TFIDFResult{
				CommandIndex: i,
				Score:        similarity * 100, // Scale for compatibility
				Similarity:   similarity,
			})
		}
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// cosineSimilarity calculates cosine similarity between query and document vectors
func (s *TFIDFSearcher) cosineSimilarity(queryVector map[int]float64, queryNorm float64,
	docVector map[int]float64, docNorm float64) float64 {
	if queryNorm == 0 || docNorm == 0 {
		return 0
	}

	var dotProduct float64
	for termIdx, queryTFIDF := range queryVector {
		if docTFIDF, exists := docVector[termIdx]; exists {
			dotProduct += queryTFIDF * docTFIDF
		}
	}

	return dotProduct / (queryNorm * docNorm)
}

// GetVocabularyStats returns statistics about the built vocabulary
func (s *TFIDFSearcher) GetVocabularyStats() map[string]interface{} {
	return map[string]interface{}{
		"vocabulary_size":       len(s.vocabulary),
		"total_commands":        len(s.commands),
		"avg_terms_per_command": s.getAverageTermsPerCommand(),
	}
}

// getAverageTermsPerCommand calculates the average number of terms per command
func (s *TFIDFSearcher) getAverageTermsPerCommand() float64 {
	totalTerms := 0
	for _, termMap := range s.commandTF {
		totalTerms += len(termMap)
	}
	return float64(totalTerms) / float64(len(s.commandTF))
}
