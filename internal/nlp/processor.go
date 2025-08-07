// Package nlp provides natural language processing capabilities for query understanding.
//
// This package implements advanced query preprocessing including:
//   - Stop word filtering and synonym expansion
//   - Intent detection (create, find, delete, etc.)
//   - Action and target word extraction
//   - Keyword enhancement for improved search relevance
//
// The QueryProcessor is the main entry point for NLP functionality.
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
		if qp.stopWords[word] {
			continue
		}

		// Check for actions
		if actions, found := qp.actionWords[word]; found {
			pq.Actions = append(pq.Actions, actions...)
			continue
		}

		// Check for targets
		if targets, found := qp.targetWords[word]; found {
			pq.Targets = append(pq.Targets, targets...)
			continue
		}

		// Add as keyword
		pq.Keywords = append(pq.Keywords, word)

		// Expand with synonyms ONLY for very specific cases
		if synonyms, found := qp.synonyms[word]; found {
			// Only add the most relevant synonym, not all of them
			if len(synonyms) > 0 {
				pq.Keywords = append(pq.Keywords, synonyms[0]) // Just the first/best synonym
			}
		}
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

// cleanQuery removes unnecessary characters and normalizes the text
func (qp *QueryProcessor) cleanQuery(query string) string {
	// Remove special characters but keep spaces and common punctuation
	re := regexp.MustCompile(`[^\w\s\-\.]`)
	cleaned := re.ReplaceAllString(query, " ")

	// Normalize multiple spaces
	re = regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}

// detectIntent analyzes actions and keywords to determine user intent
func (qp *QueryProcessor) detectIntent(actions []string, keywords []string) QueryIntent {
	// Check actions for clear intent
	for _, action := range actions {
		switch action {
		case "find", "search", "locate", "list":
			return IntentFind
		case "show", "display", "view", "see", "read", "cat":
			return IntentView
		case "create", "make", "build", "generate", "new":
			return IntentCreate
		case "delete", "remove", "destroy", "clean", "clear":
			return IntentDelete
		case "modify", "change", "edit", "update", "alter":
			return IntentModify
		case "install", "add", "download":
			return IntentInstall
		case "run", "execute", "start", "launch":
			return IntentRun
		case "configure", "config", "setup", "set":
			return IntentConfigure
		}
	}

	// Check keywords for intent hints with context
	for _, keyword := range keywords {
		switch keyword {
		case "contents", "content", "inside", "text":
			// Check if this is about viewing (see, show, read) or clearing (clear, empty)
			hasViewAction := false
			hasClearAction := false
			for _, action := range actions {
				if action == "view" || action == "show" || action == "see" || action == "read" || action == "display" {
					hasViewAction = true
				}
				if action == "clear" || action == "empty" || action == "delete" || action == "remove" {
					hasClearAction = true
				}
			}
			if hasViewAction || (!hasClearAction && len(actions) == 0) {
				return IntentView // Default to view if no clear action detected
			}
		case "install", "installation":
			return IntentInstall
		case "config", "configuration", "setup":
			return IntentConfigure
		case "running", "execution", "processes":
			return IntentFind // Finding/listing processes
		case "permissions", "permission", "chmod":
			return IntentModify
		}
	}

	return IntentGeneral
}

// GetEnhancedKeywords returns expanded keywords for better search
func (pq *ProcessedQuery) GetEnhancedKeywords() []string {
	var enhanced []string

	// Add original keywords FIRST (highest priority)
	enhanced = append(enhanced, pq.Keywords...)

	// Add actions as keywords (medium priority)
	enhanced = append(enhanced, pq.Actions...)

	// Add targets as keywords (medium priority)
	enhanced = append(enhanced, pq.Targets...)

	// Only add intent-specific keywords if we have few other keywords
	if len(enhanced) < 3 {
		switch pq.Intent {
		case IntentFind:
			enhanced = append(enhanced, "search", "find", "list")
		case IntentView:
			enhanced = append(enhanced, "cat", "view", "show", "display")
		case IntentCreate:
			enhanced = append(enhanced, "create", "make", "new")
		case IntentDelete:
			enhanced = append(enhanced, "delete", "remove")
		case IntentInstall:
			enhanced = append(enhanced, "install", "setup")
		case IntentModify:
			enhanced = append(enhanced, "chmod", "change", "modify")
		}
	}

	return removeDuplicates(enhanced)
} // buildStopWords creates a map of common English stop words
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

		// Actions
		"find":   {"search", "locate", "discover", "lookup"},
		"create": {"make", "build", "generate", "new"},
		"delete": {"remove", "destroy", "erase", "clean"},
		"copy":   {"duplicate", "clone", "backup"},
		"move":   {"relocate", "transfer", "shift"},

		// Compression
		"compress": {"zip", "archive", "pack", "bundle", "tar"},
		"extract":  {"unzip", "unpack", "decompress", "expand"},

		// Network
		"download": {"fetch", "get", "pull", "retrieve"},
		"upload":   {"push", "send", "transfer", "post"},

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
		"install": {"install", "setup", "add"},
		"setup":   {"setup", "install", "configure"},
		"add":     {"add", "install", "setup"},

		// Compression
		"compress": {"compress", "archive", "zip", "tar"},
		"extract":  {"extract", "unzip", "unpack", "tar"},
		"archive":  {"archive", "compress", "zip", "tar"},
		"pack":     {"compress", "archive", "zip", "tar"},
		"unpack":   {"extract", "unzip", "unpack", "tar"},

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
