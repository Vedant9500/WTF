package main

import (
	"fmt"
	"strings"

	"github.com/Vedant9500/WTF/internal/nlp"
)

func main() {
	query := "what is the command to manage ip in windows"

	// Test NLP processing
	processor := nlp.NewQueryProcessor()
	pq := processor.ProcessQuery(query)

	fmt.Printf("Original query: %s\n", pq.Original)
	fmt.Printf("Cleaned: %s\n", pq.Cleaned)
	fmt.Printf("Intent: %s\n", pq.Intent)
	fmt.Printf("Actions: %v\n", pq.Actions)
	fmt.Printf("Targets: %v\n", pq.Targets)
	fmt.Printf("Keywords: %v\n", pq.Keywords)
	fmt.Printf("Enhanced keywords: %v\n", pq.GetEnhancedKeywords())

	// Test basic normalization
	basicTerms := normalizeAndTokenize(query)
	fmt.Printf("\nBasic normalization: %v\n", basicTerms)

	// Test combined approach
	enh := pq.GetEnhancedKeywords()
	if len(enh) > 0 {
		// Create a set to avoid duplicates
		termSet := make(map[string]bool)

		// Add original terms first (they have priority)
		for _, term := range basicTerms {
			termSet[term] = true
		}

		// Add enhanced terms that aren't already present
		for _, enhTerm := range enh {
			termSet[enhTerm] = true
		}

		// Convert back to slice
		combinedTerms := make([]string, 0, len(termSet))

		// Prioritize original terms by adding them first
		for _, term := range basicTerms {
			combinedTerms = append(combinedTerms, term)
		}

		// Then add the new enhanced terms
		for _, enhTerm := range enh {
			found := false
			for _, origTerm := range basicTerms {
				if origTerm == enhTerm {
					found = true
					break
				}
			}
			if !found {
				combinedTerms = append(combinedTerms, enhTerm)
			}
		}

		fmt.Printf("Combined terms: %v\n", combinedTerms)
	}
}

func normalizeAndTokenize(s string) []string {
	// Apply NLP normalization
	normalized := nlp.NormalizeText(s)
	words := strings.Fields(strings.ToLower(normalized))

	// Remove stop words
	stopWords := nlp.StopWords()
	var result []string
	for _, word := range words {
		if !stopWords[word] {
			result = append(result, word)
		}
	}
	return result
}
