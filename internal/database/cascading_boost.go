package database

import (
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/nlp"
)

// cascadingBoost applies weighted boosts based on query token types.
// It uses NLP to classify tokens as actions/targets/keywords and boosts
// commands that match in order of importance: action > context > target.
func (db *Database) cascadingBoost(results []SearchResult, pq *nlp.ProcessedQuery) []SearchResult {
	if pq == nil || len(results) == 0 {
		return results
	}

	// Get synonyms for expansion
	processor := nlp.NewQueryProcessor()

	// Build boost maps with synonyms
	actionTerms := expandWithSynonyms(pq.Actions, processor)
	targetTerms := expandWithSynonyms(pq.Targets, processor)
	keywordTerms := expandWithSynonyms(pq.Keywords, processor)

	// Get command hints from NLP (e.g., "create folder" → ["mkdir"])
	commandHints := pq.GetEnhancedKeywords()

	// Boost weights (action matches are most important)
	const actionBoost = 3.0
	const targetBoost = 2.0
	const keywordBoost = 1.5
	const contextBoost = 2.5 // For known command contexts (git, docker, etc.)
	const cmdHintBoost = 6.0 // Strong boost for commands matching NLP hints

	// Extract context from query (known tool names)
	contexts := extractContexts(pq.Keywords)

	for i, r := range results {
		boost := 1.0
		cmd := r.Command

		// Combine searchable text
		searchText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))

		// Check if command name matches any NLP command hints (highest priority)
		cmdLower := strings.ToLower(cmd.Command)
		cmdBase := getCommandBase(cmdLower) // Extract base command (e.g., "mkdir" from "mkdir -p")
		for _, hint := range commandHints {
			hintLower := strings.ToLower(hint)
			if cmdBase == hintLower || cmdLower == hintLower {
				boost += cmdHintBoost
				break
			}
		}

		// Check action matches (high priority)
		for _, action := range actionTerms {
			if containsWord(searchText, action) {
				boost += actionBoost
				break // Only count once per category
			}
		}

		// Check context matches (known tools like git, docker)
		for _, ctx := range contexts {
			if containsWord(cmd.Command, ctx) || containsWord(searchText, ctx) {
				boost += contextBoost
				break
			}
		}

		// Check target matches
		for _, target := range targetTerms {
			if containsWord(searchText, target) {
				boost += targetBoost
				break
			}
		}

		// Check keyword matches
		for _, kw := range keywordTerms {
			if containsWord(searchText, kw) {
				boost += keywordBoost
				break
			}
		}

		// Apply intent-specific boosts
		boost += getIntentBoost(pq.Intent, searchText)

		results[i].Score *= boost
	}

	// Re-sort by boosted scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// getCommandBase extracts the base command from a command string
// e.g., "mkdir -p" → "mkdir", "git commit" → "git"
func getCommandBase(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) > 0 {
		return parts[0]
	}
	return cmd
}

// expandWithSynonyms adds synonyms for each term if available
func expandWithSynonyms(terms []string, processor *nlp.QueryProcessor) []string {
	expanded := make([]string, 0, len(terms)*2)
	seen := make(map[string]bool)

	for _, term := range terms {
		term = strings.ToLower(term)
		if !seen[term] {
			expanded = append(expanded, term)
			seen[term] = true
		}

		// Get synonyms from the processor
		synonyms := processor.GetSynonyms(term)
		for _, syn := range synonyms {
			syn = strings.ToLower(syn)
			if !seen[syn] {
				expanded = append(expanded, syn)
				seen[syn] = true
			}
		}
	}

	return expanded
}

// extractContexts finds known tool/command names in keywords
func extractContexts(keywords []string) []string {
	knownContexts := map[string]bool{
		"git": true, "docker": true, "npm": true, "pip": true,
		"apt": true, "yum": true, "pacman": true, "brew": true,
		"kubectl": true, "terraform": true, "ansible": true,
		"ssh": true, "rsync": true, "tar": true, "grep": true,
		"sed": true, "awk": true, "find": true, "chmod": true,
		"chown": true, "curl": true, "wget": true, "systemctl": true,
		"journalctl": true, "nginx": true, "apache": true,
		"mysql": true, "postgres": true, "redis": true, "mongo": true,
		"python": true, "node": true, "go": true, "rust": true,
		"cargo": true, "yarn": true, "composer": true, "gem": true,
		"arch": true, "ubuntu": true, "debian": true, "centos": true,
		"windows": true, "macos": true, "linux": true,
		"ipconfig": true, "ifconfig": true, "netstat": true, "ip": true,
	}

	var contexts []string
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		if knownContexts[kwLower] {
			contexts = append(contexts, kwLower)
		}
	}
	return contexts
}

// intentKeywords maps each intent to its associated keywords for boosting
var intentKeywords = map[nlp.QueryIntent][]string{
	nlp.IntentView:    {"show", "display", "list", "view", "cat", "less"},
	nlp.IntentDelete:  {"delete", "remove", "rm", "uninstall"},
	nlp.IntentCreate:  {"create", "make", "new", "mkdir", "touch"},
	nlp.IntentInstall: {"install", "setup", "add"},
	nlp.IntentModify:  {"modify", "change", "edit", "update", "undo", "revert", "reset"},
}

// getIntentBoost returns additional boost based on detected intent
func getIntentBoost(intent nlp.QueryIntent, searchText string) float64 {
	keywords, ok := intentKeywords[intent]
	if !ok {
		return 0.0
	}
	for _, kw := range keywords {
		if containsWord(searchText, kw) {
			return 1.5
		}
	}
	return 0.0
}

// containsWord checks if text contains a whole word (not substring)
func containsWord(text, word string) bool {
	text = " " + text + " "
	word = " " + word + " "
	return strings.Contains(text, word)
}
