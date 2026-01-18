package database

import (
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/nlp"
)

// boostContext holds all data needed for boost calculations
type boostContext struct {
	actionTerms  []string
	targetTerms  []string
	keywordTerms []string
	commandHints []string
	contexts     []string
	intent       nlp.QueryIntent
}

// cascadingBoost applies weighted boosts based on query token types.
// It uses NLP to classify tokens as actions/targets/keywords and boosts
// commands that match in order of importance: action > context > target.
func (db *Database) cascadingBoost(results []SearchResult, pq *nlp.ProcessedQuery) []SearchResult {
	if pq == nil || len(results) == 0 {
		return results
	}

	ctx := db.buildBoostContext(pq)

	for i, r := range results {
		boost := db.calculateBoostForCommand(r.Command, ctx)
		results[i].Score *= boost
	}

	// Re-sort by boosted scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// buildBoostContext creates the context needed for boost calculations
func (db *Database) buildBoostContext(pq *nlp.ProcessedQuery) boostContext {
	processor := nlp.NewQueryProcessor()
	return boostContext{
		actionTerms:  expandWithSynonyms(pq.Actions, processor),
		targetTerms:  expandWithSynonyms(pq.Targets, processor),
		keywordTerms: expandWithSynonyms(pq.Keywords, processor),
		commandHints: pq.GetEnhancedKeywords(),
		contexts:     extractContexts(pq.Keywords),
		intent:       pq.Intent,
	}
}

// calculateBoostForCommand calculates the total boost for a single command
func (db *Database) calculateBoostForCommand(cmd *Command, ctx boostContext) float64 {
	const (
		actionBoost  = 3.0
		targetBoost  = 2.0
		keywordBoost = 1.5
		contextBoost = 2.5
		cmdHintBoost = 6.0
	)

	boost := 1.0
	searchText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))

	boost += calcHintBoost(cmd.Command, ctx.commandHints, cmdHintBoost)
	boost += calcTermBoost(searchText, ctx.actionTerms, actionBoost)
	boost += calcContextBoost(cmd.Command, searchText, ctx.contexts, contextBoost)
	boost += calcTermBoost(searchText, ctx.targetTerms, targetBoost)
	boost += calcTermBoost(searchText, ctx.keywordTerms, keywordBoost)
	boost += getIntentBoost(ctx.intent, searchText)

	return boost
}

// calcHintBoost returns boost if command matches any hint
func calcHintBoost(command string, hints []string, boostVal float64) float64 {
	cmdLower := strings.ToLower(command)
	cmdBase := getCommandBase(cmdLower)
	for _, hint := range hints {
		hintLower := strings.ToLower(hint)
		if cmdBase == hintLower || cmdLower == hintLower {
			return boostVal
		}
	}
	return 0
}

// calcTermBoost returns boost if any term is found in searchText
func calcTermBoost(searchText string, terms []string, boostVal float64) float64 {
	for _, term := range terms {
		if containsWord(searchText, term) {
			return boostVal
		}
	}
	return 0
}

// calcContextBoost returns boost if any context is found in command or searchText
func calcContextBoost(command, searchText string, contexts []string, boostVal float64) float64 {
	for _, ctx := range contexts {
		if containsWord(command, ctx) || containsWord(searchText, ctx) {
			return boostVal
		}
	}
	return 0
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
