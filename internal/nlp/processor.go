// Package nlp provides natural language processing capabilities for query understanding.
package nlp

import (
	"regexp"
	"strings"
)

// QueryProcessor handles natural language query preprocessing
type QueryProcessor struct {
	stopWords   map[string]bool
	synonyms    map[string][]string
	actionWords map[string][]string
	targetWords map[string][]string
}

// NewQueryProcessor creates a new query processor with NLP capabilities
func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{
		stopWords:   buildStopWords(),
		synonyms:    buildSynonyms(),
		actionWords: buildActionWords(),
		targetWords: buildTargetWords(),
	}
}

// ProcessedQuery represents a query after NLP processing
type ProcessedQuery struct {
	Original  string
	Cleaned   string
	Actions   []string
	Targets   []string
	Keywords  []string
	Intent    QueryIntent
	Modifiers []string
}

// QueryIntent represents the type of intent detected in the query
type QueryIntent string

const (
	IntentFind      QueryIntent = "find"
	IntentCreate    QueryIntent = "create"
	IntentDelete    QueryIntent = "delete"
	IntentModify    QueryIntent = "modify"
	IntentView      QueryIntent = "view"
	IntentRun       QueryIntent = "run"
	IntentInstall   QueryIntent = "install"
	IntentConfigure QueryIntent = "configure"
	IntentGeneral   QueryIntent = "general"
)

const (
	actionShow    = "show"
	actionSetup   = "setup"
	actionDelete  = "delete"
	actionInstall = "install"
)

// ProcessQuery analyzes and enhances a natural language query
func (qp *QueryProcessor) ProcessQuery(query string) *ProcessedQuery {
	pq := &ProcessedQuery{
		Original:  query,
		Actions:   []string{},
		Targets:   []string{},
		Keywords:  []string{},
		Modifiers: []string{},
		Intent:    IntentGeneral,
	}

	// Clean and normalize the query
	cleaned := qp.cleanQuery(query)
	pq.Cleaned = cleaned

	// Extract words
	words := strings.Fields(strings.ToLower(cleaned))

	// Detect context clues for better intent detection
	queryLower := strings.ToLower(query)
	hasViewContext := strings.Contains(queryLower, "see") || strings.Contains(queryLower, "view") ||
		strings.Contains(queryLower, "show") || strings.Contains(queryLower, "display") ||
		strings.Contains(queryLower, "read") || strings.Contains(queryLower, "look")
	hasWithoutOpening := strings.Contains(queryLower, "without opening") || strings.Contains(queryLower, "without editing")

	// Remove stop words and categorize remaining words
	for _, word := range words {
		qp.processWord(pq, word)
	}

	// Apply context-based action enhancement
	if hasViewContext && hasWithoutOpening {
		// This is clearly about viewing file contents
		pq.Actions = append(pq.Actions, "view", "show", "display")
	}

	// Detect intent
	pq.Intent = qp.detectIntent(pq.Actions, pq.Keywords)

	// Remove duplicates
	pq.Actions = removeDuplicates(pq.Actions)
	pq.Targets = removeDuplicates(pq.Targets)
	pq.Keywords = removeDuplicates(pq.Keywords)

	return pq
}

func (qp *QueryProcessor) processWord(pq *ProcessedQuery, word string) {
	normalized := normalizeToken(word)
	if qp.isStopWord(word, normalized) {
		return
	}

	if qp.tryAppendActions(pq, word, normalized) {
		return
	}

	if qp.tryAppendTargets(pq, word, normalized) {
		return
	}

	qp.appendKeywordsAndSynonyms(pq, word, normalized)
}

func (qp *QueryProcessor) isStopWord(word, normalized string) bool {
	return qp.stopWords[word] || qp.stopWords[normalized]
}

func (qp *QueryProcessor) tryAppendActions(pq *ProcessedQuery, word, normalized string) bool {
	if actions, found := qp.actionWords[word]; found {
		pq.Actions = append(pq.Actions, actions...)
		return true
	}
	if normalized != word {
		if actions, found := qp.actionWords[normalized]; found {
			pq.Actions = append(pq.Actions, actions...)
			return true
		}
	}
	return false
}

func (qp *QueryProcessor) tryAppendTargets(pq *ProcessedQuery, word, normalized string) bool {
	if targets, found := qp.targetWords[word]; found {
		pq.Targets = append(pq.Targets, targets...)
		// Keep the original token as keyword to preserve direct user signal.
		pq.Keywords = append(pq.Keywords, word)
		return true
	}
	if normalized != word {
		if targets, found := qp.targetWords[normalized]; found {
			pq.Targets = append(pq.Targets, targets...)
			pq.Keywords = append(pq.Keywords, word, normalized)
			return true
		}
	}
	return false
}

func (qp *QueryProcessor) appendKeywordsAndSynonyms(pq *ProcessedQuery, word, normalized string) {
	pq.Keywords = append(pq.Keywords, word)
	if normalized != word {
		pq.Keywords = append(pq.Keywords, normalized)
	}

	if synonyms, found := qp.synonyms[word]; found {
		if len(synonyms) > 0 {
			pq.Keywords = append(pq.Keywords, synonyms[0])
		}
		return
	}

	if synonyms, found := qp.synonyms[normalized]; found && len(synonyms) > 0 {
		pq.Keywords = append(pq.Keywords, synonyms[0])
	}
}

// cleanQuery removes unnecessary characters and normalizes the text
func (qp *QueryProcessor) cleanQuery(query string) string {
	// Remove special characters but keep spaces and common punctuation
	re := regexp.MustCompile(`[^\w\s\-.]`)
	cleaned := re.ReplaceAllString(query, " ")

	// Normalize multiple spaces
	re = regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}

// NormalizeText applies the same cleanup used in the NLP query processor.
// It removes special characters (keeping spaces, hyphens, and dots) and collapses whitespace.
func NormalizeText(s string) string {
	re := regexp.MustCompile(`[^\w\s\-.]`)
	cleaned := re.ReplaceAllString(s, " ")
	re = regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}

// StopWords exposes the default stop words used by the NLP processor for consistency.
func StopWords() map[string]bool {
	// return a copy to avoid external mutation
	src := buildStopWords()
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// detectIntent analyzes actions and keywords to determine user intent
func (qp *QueryProcessor) detectIntent(actions, keywords []string) QueryIntent {
	// Score intents from both actions and keywords to avoid early generic-view
	// matches dominating stronger task-specific verbs in conversational queries.
	scores := map[QueryIntent]float64{}

	for _, action := range actions {
		for intent, w := range qp.intentVotesFromAction(action) {
			scores[intent] += w
		}
	}

	for _, keyword := range keywords {
		for intent, w := range qp.intentVotesFromKeyword(keyword, actions) {
			scores[intent] += w
		}
	}

	best := IntentGeneral
	bestScore := 0.0
	for intent, score := range scores {
		if score > bestScore {
			best = intent
			bestScore = score
		}
	}

	if bestScore <= 0 {
		return IntentGeneral
	}

	return best
}

func (qp *QueryProcessor) intentVotesFromAction(action string) map[QueryIntent]float64 {
	votes := map[QueryIntent]float64{}
	if action == "" {
		return votes
	}

	switch action {
	case "find", "search", "locate", "list":
		votes[IntentFind] += 1.0
	case actionShow, "display", "view", "see", "read", "cat", "check":
		votes[IntentView] += 0.35
	case "create", "make", "build", "generate", "new":
		votes[IntentCreate] += 1.0
	case actionDelete, "remove", "destroy", "clean", "clear":
		votes[IntentDelete] += 1.0
	case "modify", "change", "edit", "update", "alter", "save":
		votes[IntentModify] += 1.0
	case actionInstall, "add", "download", "fetch", "retrieve":
		votes[IntentInstall] += 1.0
	case "run", "execute", "start", "launch", "kill", "stop", "terminate":
		votes[IntentRun] += 1.0
	case "configure", "config", actionSetup, "set":
		votes[IntentConfigure] += 1.0
	}

	return votes
}

func (qp *QueryProcessor) intentVotesFromKeyword(keyword string, actions []string) map[QueryIntent]float64 {
	votes := map[QueryIntent]float64{}
	if keyword == "" {
		return votes
	}

	switch keyword {
	case "contents", "content", "inside", "text":
		if qp.isViewContext(actions) {
			votes[IntentView] += 0.4
		}
	case actionInstall, "installation", "download", "fetch", "url", "https", "http":
		votes[IntentInstall] += 0.6
	case "config", "configuration", actionSetup:
		votes[IntentConfigure] += 0.6
	case "running", "execution", "processes":
		votes[IntentFind] += 0.4
	case "permissions", "permission", "chmod":
		votes[IntentModify] += 0.4
	}

	return votes
}

func (qp *QueryProcessor) isViewContext(actions []string) bool {
	hasViewAction := false
	hasClearAction := false
	for _, action := range actions {
		switch action {
		case "view", "show", "see", "read", "display":
			hasViewAction = true
		case "clear", "empty", "delete", "remove":
			hasClearAction = true
		}
	}
	return hasViewAction || (!hasClearAction && len(actions) == 0)
}

// GetEnhancedKeywords returns expanded keywords for better search
func (pq *ProcessedQuery) GetEnhancedKeywords() []string {
	var enhanced []string

	// Add original keywords first
	enhanced = append(enhanced, pq.Keywords...)

	// Add actions as keywords (medium priority) - but only the most relevant ones
	enhanced = append(enhanced, pq.getRelevantActions()...)

	// Add targets as keywords (medium priority) - but only the most relevant ones
	enhanced = append(enhanced, pq.getRelevantTargets()...)

	// Only add intent-specific keywords if we have very few other keywords
	if len(enhanced) < 4 {
		enhanced = append(enhanced, pq.getIntentKeywords()...)
	}

	return removeDuplicates(enhanced)
}

// buildStopWords creates a map of common English stop words
func buildStopWords() map[string]bool {
	words := []string{
		"a", "an", "and", "are", "as", "at", "be", "by", "for", "from",
		"has", "he", "in", "is", "it", "its", "of", "on", "that", "the",
		"to", "was", "will", "with", "this", "but", "they", "have",
		"had", "what", "said", "each", "which", "she", "how", "their",
		"if", "up", "out", "many", "then", "them", "these", "so", "some", "her",
		"would", "like", "into", "him", "time", "two", "go", "no",
		"way", "could", "my", "than", "first", "been", "call", "who", "oil", "sit",
		"now", "down", "day", "did", "get", "come", "made", "may", "part",
		"command", "commands",
	}

	stopWords := make(map[string]bool)
	for _, word := range words {
		stopWords[word] = true
	}
	return stopWords
}

// buildSynonyms creates a map of word synonyms for better matching
func buildSynonyms() map[string][]string {
	return map[string][]string{
		// File operations
		"file":     {"document", "data", "content"},
		"files":    {"documents", "data", "content"},
		"folder":   {"directory", "dir", "path"},
		"folders":  {"directories", "dirs", "paths"},
		"contents": {"content", "data", "text", "inside"},
		"content":  {"contents", "data", "text", "inside"},

		// Viewing/Reading actions - more comprehensive
		"see":     {"view", "show", "display", "read", "cat", "less", "more"},
		"view":    {"see", "show", "display", "read", "cat", "less", "more"},
		"show":    {"view", "see", "display", "read", "cat", "less"},
		"display": {"view", "see", "show", "read", "cat", "less"},
		"read":    {"view", "see", "show", "display", "cat", "less"},
		"look":    {"view", "see", "show", "display", "cat"},
		"check":   {"view", "see", "show", "display", "cat"},
		"print":   {"cat", "echo", "printf", "show"},
		"cat":     {"view", "show", "display", "less", "more", "head", "tail"},

		// Actions
		"find":   {"search", "locate", "discover", "lookup"},
		"create": {"make", "build", "generate", "new"},
		"delete": {"remove", "destroy", "erase", "clean"},
		"copy":   {"duplicate", "clone", "backup"},
		"move":   {"relocate", "transfer", "shift"},

		// Compression - bi-directional mappings
		"compress":   {"zip", "archive", "pack", "bundle", "tar"},
		"extract":    {"unzip", "unpack", "decompress", "expand", "tar -x"},
		"decompress": {"unzip", "extract", "unpack", "expand", "gunzip"},
		"unpack":     {"unzip", "extract", "decompress", "expand"},
		"unarchive":  {"unzip", "extract", "untar", "expand"},

		// Network
		"download":  {"fetch", "get", "pull", "retrieve"},
		"upload":    {"push", "send", "transfer", "post"},
		"ip":        {"network", "interface", "address", "configuration", "config"},
		"network":   {"ip", "interface", "connection", "config"},
		"interface": {"ip", "network", "adapter", "config"},

		// System
		"process":   {"task", "job", "service", "daemon"},
		"processes": {"tasks", "jobs", "services"},
		"running":   {"active", "executing", "live"},
		"kill":      {"stop", "terminate", "end"},
		"start":     {"run", "launch", "execute", "begin"},

		// Permissions
		"permission":  {"permissions", "access", "rights", "chmod"},
		"permissions": {"permission", "access", "rights", "chmod"},
		"change":      {"modify", "alter", "update", "edit"},

		// Development
		"compile": {"build", "make", "assemble"},
		"deploy":  {"release", "publish", "ship"},
		"test":    {"check", "verify", "validate"},
	}
}

// buildActionWords maps natural language actions to standardized verbs
func buildActionWords() map[string][]string {
	return map[string][]string{
		// Finding/Searching
		"find":   {"find", "search", "locate"},
		"search": {"find", "search", "locate"},
		"locate": {"find", "search", "locate"},
		"list":   {"list", "show", "display"},

		// Viewing/Reading
		"show":    {"show", "display", "view"},
		"display": {"show", "display", "view"},
		"view":    {"view", "show", "display"},
		"see":     {"view", "show", "display"},
		"read":    {"view", "show", "display"},
		"look":    {"view", "show", "display"},
		"check":   {"view", "show", "display"},

		// Creating
		"create":   {"create", "make", "build"},
		"make":     {"create", "make", "build"},
		"build":    {"build", "create", "make"},
		"generate": {"create", "make", "build"},
		"new":      {"create", "make", "build"},

		// Modifying
		"edit":   {"edit", "modify", "change"},
		"modify": {"edit", "modify", "change"},
		"change": {"edit", "modify", "change"},
		"update": {"update", "modify", "change"},
		"manage": {"manage", "modify", "configure", "control"},

		// Removing
		"delete":  {"delete", "remove", "destroy"},
		"remove":  {"delete", "remove", "destroy"},
		"destroy": {"delete", "remove", "destroy"},
		"clean":   {"clean", "delete", "remove"},

		// Running
		"run":     {"run", "execute", "start"},
		"execute": {"run", "execute", "start"},
		"start":   {"start", "run", "execute"},
		"launch":  {"start", "run", "execute"},

		// Installing
		"install":   {"install", "setup", "add"},
		"setup":     {"setup", "install", "configure"},
		"configure": {"configure", "config", "setup", "set"},
		"config":    {"configure", "config", "setup", "set"},
		"add":       {"add", "install", "setup"},
		"download":  {"download", "fetch", "retrieve", "save"},
		"fetch":     {"download", "fetch", "retrieve"},
		"retrieve":  {"download", "fetch", "retrieve"},
		"save":      {"save", "write", "output"},

		// Compression
		"compress":   {"compress", "archive", "zip", "tar"},
		"extract":    {"extract", "unzip", "unpack", "tar"},
		"archive":    {"archive", "compress", "zip", "tar"},
		"pack":       {"compress", "archive", "zip", "tar"},
		"unpack":     {"extract", "unzip", "unpack", "tar"},
		"decompress": {"extract", "unzip", "decompress", "gunzip"},

		// File operations
		"copy":   {"copy", "cp", "duplicate"},
		"move":   {"move", "mv", "relocate"},
		"rename": {"rename", "mv", "move"},

		// Process operations
		"kill":      {"kill", "stop", "terminate"},
		"stop":      {"stop", "kill", "terminate"},
		"terminate": {"terminate", "kill", "stop"},
	}
}

// buildTargetWords maps natural language targets to standardized objects
func buildTargetWords() map[string][]string {
	return map[string][]string{
		// File system
		"file":        {"file", "document"},
		"files":       {"files", "documents"},
		"folder":      {"directory", "folder"},
		"directory":   {"directory", "folder"},
		"directories": {"directories", "folders"},
		"path":        {"path", "location"},
		"contents":    {"contents", "content", "data"},
		"content":     {"content", "contents", "data"},

		// Archives
		"archive":  {"archive", "zip", "tar"},
		"archives": {"archives", "zip", "tar"},
		"zip":      {"zip", "archive"},
		"tar":      {"tar", "archive"},

		// Processes
		"process":   {"process", "task"},
		"processes": {"processes", "tasks"},
		"service":   {"service", "daemon"},
		"services":  {"services", "daemons"},
		"daemon":    {"daemon", "service"},
		"task":      {"task", "process"},
		"tasks":     {"tasks", "processes"},

		// Network
		"server":     {"server", "host"},
		"port":       {"port", "socket"},
		"connection": {"connection", "link"},
		"url":        {"url", "link", "address"},
		"website":    {"website", "url", "site"},
		"ip":         {"network", "interface", "address"},
		"network":    {"ip", "interface", "connection"},
		"interface":  {"ip", "network", "adapter"},

		// Development
		"project":    {"project", "repo", "repository"},
		"repository": {"repository", "repo"},
		"repo":       {"repo", "repository"},
		"branch":     {"branch", "ref"},
		"commit":     {"commit", "revision"},
		"code":       {"code", "source", "program"},

		// System
		"permission":  {"permission", "permissions", "access"},
		"permissions": {"permissions", "permission", "access"},
		"user":        {"user", "account"},
		"group":       {"group", "users"},
	}
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetSynonyms returns synonyms for a given word from the processor's synonym map
func (qp *QueryProcessor) GetSynonyms(word string) []string {
	word = strings.ToLower(word)
	if synonyms, found := qp.synonyms[word]; found {
		return synonyms
	}
	return nil
}
