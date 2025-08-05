// Package search provides pattern learning capabilities that adapt to the database
package search

import (
	"math"
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
)

// PatternLearner learns patterns from the database to improve search accuracy
type PatternLearner struct {
	db                *database.Database
	wordAssociations  map[string]map[string]float64 // word -> related words with scores
	commandPatterns   map[string][]string           // command -> associated words
	descriptionWords  map[string][]string           // common words in descriptions
	functionalGroups  map[string][]string           // groups of similar commands
}

// NewPatternLearner creates a new pattern learner
func NewPatternLearner(db *database.Database) *PatternLearner {
	pl := &PatternLearner{
		db:               db,
		wordAssociations: make(map[string]map[string]float64),
		commandPatterns:  make(map[string][]string),
		descriptionWords: make(map[string][]string),
		functionalGroups: make(map[string][]string),
	}
	
	pl.learnPatterns()
	return pl
}

// learnPatterns analyzes the database to learn word associations and command patterns
func (pl *PatternLearner) learnPatterns() {
	pl.learnWordAssociations()
	pl.learnCommandPatterns()
	pl.learnFunctionalGroups()
}

// learnWordAssociations discovers which words commonly appear together
func (pl *PatternLearner) learnWordAssociations() {
	// Track word co-occurrences across all commands
	cooccurrence := make(map[string]map[string]int)
	wordFreq := make(map[string]int)
	
	for _, cmd := range pl.db.Commands {
		// Combine all text sources
		allText := strings.Join([]string{
			cmd.Command,
			cmd.Description,
			strings.Join(cmd.Keywords, " "),
		}, " ")
		
		words := pl.extractMeaningfulWords(allText)
		
		// Count individual word frequencies
		for _, word := range words {
			wordFreq[word]++
		}
		
		// Count co-occurrences
		for i, word1 := range words {
			if cooccurrence[word1] == nil {
				cooccurrence[word1] = make(map[string]int)
			}
			
			for j, word2 := range words {
				if i != j {
					cooccurrence[word1][word2]++
				}
			}
		}
	}
	
	// Convert co-occurrence counts to association scores using PMI (Pointwise Mutual Information)
	totalWords := 0
	for _, freq := range wordFreq {
		totalWords += freq
	}
	
	for word1, cooccurMap := range cooccurrence {
		if pl.wordAssociations[word1] == nil {
			pl.wordAssociations[word1] = make(map[string]float64)
		}
		
		for word2, cooccurCount := range cooccurMap {
			if wordFreq[word1] > 1 && wordFreq[word2] > 1 { // Skip rare words
				// Calculate PMI: log(P(word1,word2) / (P(word1) * P(word2)))
				pWord1 := float64(wordFreq[word1]) / float64(totalWords)
				pWord2 := float64(wordFreq[word2]) / float64(totalWords)
				pBoth := float64(cooccurCount) / float64(totalWords)
				
				if pBoth > 0 && pWord1 > 0 && pWord2 > 0 {
					pmi := math.Log(pBoth / (pWord1 * pWord2))
					if pmi > 0 { // Only keep positive associations
						pl.wordAssociations[word1][word2] = pmi
					}
				}
			}
		}
	}
}

// learnCommandPatterns identifies patterns in how commands are described
func (pl *PatternLearner) learnCommandPatterns() {
	for _, cmd := range pl.db.Commands {
		cmdName := strings.Fields(strings.ToLower(cmd.Command))[0] // Get base command
		
		// Extract meaningful words from description
		descWords := pl.extractMeaningfulWords(cmd.Description)
		pl.commandPatterns[cmdName] = descWords
		
		// Group by description patterns
		for _, word := range descWords {
			pl.descriptionWords[word] = append(pl.descriptionWords[word], cmdName)
		}
	}
}

// learnFunctionalGroups discovers commands that serve similar functions
func (pl *PatternLearner) learnFunctionalGroups() {
	// Group commands by shared keywords and description patterns
	commandSimilarity := make(map[string]map[string]float64)
	
	for i, cmd1 := range pl.db.Commands {
		cmdName1 := strings.Fields(strings.ToLower(cmd1.Command))[0]
		if commandSimilarity[cmdName1] == nil {
			commandSimilarity[cmdName1] = make(map[string]float64)
		}
		
		for j, cmd2 := range pl.db.Commands {
			if i != j {
				cmdName2 := strings.Fields(strings.ToLower(cmd2.Command))[0]
				similarity := pl.calculateCommandSimilarity(&cmd1, &cmd2)
				
				if similarity > 0.3 { // Similarity threshold
					commandSimilarity[cmdName1][cmdName2] = similarity
				}
			}
		}
	}
	
	// Create functional groups based on similarity
	processed := make(map[string]bool)
	
	for cmdName, similarities := range commandSimilarity {
		if processed[cmdName] {
			continue
		}
		
		// Find all similar commands
		group := []string{cmdName}
		for similarCmd, similarity := range similarities {
			if similarity > 0.5 && !processed[similarCmd] {
				group = append(group, similarCmd)
				processed[similarCmd] = true
			}
		}
		
		if len(group) > 1 {
			// Use the most common word in descriptions as group name
			groupName := pl.findGroupName(group)
			if groupName != "" {
				pl.functionalGroups[groupName] = group
			}
		}
		
		processed[cmdName] = true
	}
}

// calculateCommandSimilarity computes similarity between two commands
func (pl *PatternLearner) calculateCommandSimilarity(cmd1, cmd2 *database.Command) float64 {
	// Compare keywords
	keywordSim := pl.calculateSetSimilarity(cmd1.Keywords, cmd2.Keywords)
	
	// Compare description words
	desc1Words := pl.extractMeaningfulWords(cmd1.Description)
	desc2Words := pl.extractMeaningfulWords(cmd2.Description)
	descSim := pl.calculateSetSimilarity(desc1Words, desc2Words)
	
	// Compare tags
	tagSim := pl.calculateSetSimilarity(cmd1.Tags, cmd2.Tags)
	
	// Weighted combination
	return 0.4*keywordSim + 0.4*descSim + 0.2*tagSim
}

// calculateSetSimilarity computes Jaccard similarity between two string sets
func (pl *PatternLearner) calculateSetSimilarity(set1, set2 []string) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0
	}
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0
	}
	
	// Convert to maps for faster lookup
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)
	
	for _, item := range set1 {
		map1[strings.ToLower(item)] = true
	}
	for _, item := range set2 {
		map2[strings.ToLower(item)] = true
	}
	
	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)
	
	for item := range map1 {
		union[item] = true
		if map2[item] {
			intersection++
		}
	}
	for item := range map2 {
		union[item] = true
	}
	
	return float64(intersection) / float64(len(union))
}

// findGroupName determines the best name for a functional group
func (pl *PatternLearner) findGroupName(commands []string) string {
	wordFreq := make(map[string]int)
	
	// Count words in descriptions of all commands in the group
	for _, cmdName := range commands {
		if words, exists := pl.commandPatterns[cmdName]; exists {
			for _, word := range words {
				wordFreq[word]++
			}
		}
	}
	
	// Find the most common meaningful word
	maxFreq := 0
	bestWord := "misc"
	
	for word, freq := range wordFreq {
		if freq > maxFreq && len(word) > 3 {
			maxFreq = freq
			bestWord = word
		}
	}
	
	return bestWord
}

// extractMeaningfulWords extracts important words from text
func (pl *PatternLearner) extractMeaningfulWords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var meaningful []string
	
	// Simple stop words (minimal set)
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
		"or": true, "but": true, "not": true, "can": true, "have": true, "this": true,
	}
	
	for _, word := range words {
		// Clean word
		cleaned := strings.Trim(word, ".,!?;:()[]{}\"'")
		
		// Keep meaningful words
		if len(cleaned) > 2 && !stopWords[cleaned] {
			meaningful = append(meaningful, cleaned)
		}
	}
	
	return meaningful
}

// ExpandQuery dynamically expands a query with learned associations
func (pl *PatternLearner) ExpandQuery(query string) []string {
	queryWords := pl.extractMeaningfulWords(query)
	expanded := make(map[string]bool)
	
	// Add original words
	for _, word := range queryWords {
		expanded[word] = true
	}
	
	// Add associated words based on learned patterns
	for _, word := range queryWords {
		if associations, exists := pl.wordAssociations[word]; exists {
			// Add top associated words
			type assoc struct {
				word  string
				score float64
			}
			
			var assocs []assoc
			for assocWord, score := range associations {
				assocs = append(assocs, assoc{assocWord, score})
			}
			
			// Sort by score and take top associations
			sort.Slice(assocs, func(i, j int) bool {
				return assocs[i].score > assocs[j].score
			})
			
			// Add top 2 associations per word to avoid query explosion
			for i, a := range assocs {
				if i >= 2 {
					break
				}
				if a.score > 1.0 { // Only strong associations
					expanded[a.word] = true
				}
			}
		}
	}
	
	// Convert back to slice
	var result []string
	for word := range expanded {
		result = append(result, word)
	}
	
	return result
}

// SuggestSimilarCommands finds commands similar to a given command
func (pl *PatternLearner) SuggestSimilarCommands(commandName string, maxSuggestions int) []string {
	var suggestions []string
	
	// Look for commands in the same functional group
	for _, commands := range pl.functionalGroups {
		for _, cmd := range commands {
			if strings.Contains(cmd, commandName) || strings.Contains(commandName, cmd) {
				// Add other commands from the same group
				for _, otherCmd := range commands {
					if otherCmd != cmd && len(suggestions) < maxSuggestions {
						suggestions = append(suggestions, otherCmd)
					}
				}
				break
			}
		}
		
		if len(suggestions) >= maxSuggestions {
			break
		}
	}
	
	return suggestions
}

// GetDynamicIntentScore calculates intent score based on learned patterns
func (pl *PatternLearner) GetDynamicIntentScore(query string, cmd *database.Command) float64 {
	queryWords := pl.extractMeaningfulWords(query)
	cmdWords := pl.extractMeaningfulWords(cmd.Description + " " + strings.Join(cmd.Keywords, " "))
	
	var totalScore float64
	
	for _, queryWord := range queryWords {
		bestScore := 0.0
		
		for _, cmdWord := range cmdWords {
			// Direct match
			if queryWord == cmdWord {
				bestScore = math.Max(bestScore, 1.0)
				continue
			}
			
			// Association-based match
			if associations, exists := pl.wordAssociations[queryWord]; exists {
				if score, hasAssoc := associations[cmdWord]; hasAssoc {
					bestScore = math.Max(bestScore, score/5.0) // Scale down association scores
				}
			}
		}
		
		totalScore += bestScore
	}
	
	// Normalize by query length
	if len(queryWords) > 0 {
		return totalScore / float64(len(queryWords))
	}
	
	return 0.0
}