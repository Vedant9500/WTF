package database

import (
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/constants"
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
	boost := 1.0
	searchText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))

	matches := boostMatches{}

	if hasHintMatch(cmd.Command, ctx.commandHints) {
		matches.hint = true
		boost += constants.CascadingHintBoost
	}
	if hasTermMatch(searchText, ctx.actionTerms) {
		matches.action = true
		boost += constants.CascadingActionBoost
	}
	if hasContextMatch(cmd.Command, searchText, ctx.contexts) {
		matches.context = true
		boost += constants.CascadingContextBoost
	}
	if hasTermMatch(searchText, ctx.targetTerms) {
		matches.target = true
		boost += constants.CascadingTargetBoost
	}
	if hasTermMatch(searchText, ctx.keywordTerms) {
		matches.keyword = true
		boost += constants.CascadingKeywordBoost
	}
	if hasIntentMatch(ctx.intent, searchText) {
		matches.intent = true
		boost += constants.CascadingIntentBoost
	}

	if !shouldApplyCascadingBoost(matches) {
		return 1.0
	}

	if boost > constants.CascadingMaxMultiplier {
		return constants.CascadingMaxMultiplier
	}

	return boost
}

type boostMatches struct {
	hint    bool
	action  bool
	target  bool
	keyword bool
	context bool
	intent  bool
}

func shouldApplyCascadingBoost(m boostMatches) bool {
	count := 0
	if m.hint {
		count++
	}
	if m.action {
		count++
	}
	if m.target {
		count++
	}
	if m.keyword {
		count++
	}
	if m.context {
		count++
	}
	if m.intent {
		count++
	}

	if count >= constants.CascadingMinSignalCount {
		return true
	}

	if constants.CascadingAllowSingleSignalWithHint && m.hint {
		return true
	}

	return false
}

// hasHintMatch returns true if command matches any hint
func hasHintMatch(command string, hints []string) bool {
	cmdLower := strings.ToLower(command)
	cmdBase := getCommandBase(cmdLower)
	for _, hint := range hints {
		hintLower := strings.ToLower(hint)
		if cmdBase == hintLower || cmdLower == hintLower {
			return true
		}
	}
	return false
}

// hasTermMatch returns true if any term is found in searchText
func hasTermMatch(searchText string, terms []string) bool {
	for _, term := range terms {
		if containsWord(searchText, term) {
			return true
		}
	}
	return false
}

// hasContextMatch returns true if any context is found in command or searchText
func hasContextMatch(command, searchText string, contexts []string) bool {
	for _, ctx := range contexts {
		if containsWord(command, ctx) || containsWord(searchText, ctx) {
			return true
		}
	}
	return false
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

// hasIntentMatch returns whether command text matches intent-associated keywords.
func hasIntentMatch(intent nlp.QueryIntent, searchText string) bool {
	keywords, ok := intentKeywords[intent]
	if !ok {
		return false
	}
	for _, kw := range keywords {
		if containsWord(searchText, kw) {
			return true
		}
	}
	return false
}

// containsWord checks if text contains a whole word (not substring)
func containsWord(text, word string) bool {
	text = " " + text + " "
	word = " " + word + " "
	return strings.Contains(text, word)
}
