package database

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// ModernSearchEngine implements state-of-the-art search techniques inspired by
// ColBERT, SPLADE, and hybrid sparse+dense retrieval models
type ModernSearchEngine struct {
	db *Database
	// Context vectors for semantic similarity (simplified embeddings)
	termVectors map[string][]float64
	// Pre-computed semantic similarity matrix
	semanticGraph map[string][]string
}

// NewModernSearchEngine creates an advanced search engine with semantic capabilities
func NewModernSearchEngine(db *Database) *ModernSearchEngine {
	engine := &ModernSearchEngine{
		db:            db,
		termVectors:   make(map[string][]float64),
		semanticGraph: make(map[string][]string),
	}

	// Build semantic understanding during initialization
	engine.buildSemanticGraph()
	return engine
}

// SmartSearch implements a hybrid approach combining multiple modern IR techniques
func (mse *ModernSearchEngine) SmartSearch(query string, options SearchOptions) []SearchResult {
	// Step 1: Parse and understand the query intent
	queryContext := mse.analyzeQueryIntent(query)

	// Step 2: Multi-stage retrieval pipeline
	// Stage 1: Sparse retrieval (BM25F) - fast first-pass
	sparseResults := mse.db.SearchUniversal(query, options)

	// Stage 2: Dense semantic retrieval - handles intent and context
	semanticResults := mse.semanticRetrieval(query, queryContext, options)

	// Stage 3: Hybrid fusion - combine and re-rank
	fusedResults := mse.hybridFusion(sparseResults, semanticResults, queryContext, options.Limit)

	// Stage 4: Intent-based re-ranking
	finalResults := mse.intentBasedReranking(fusedResults, queryContext, options.Limit)

	return finalResults
}

// QueryContext represents the understood intent and semantic context of a query
type QueryContext struct {
	Intent         string   // find, manage, view, configure, etc.
	Domain         string   // network, file, process, system, etc.
	Platform       string   // windows, linux, mac, etc.
	ActionType     string   // read, write, modify, delete, etc.
	TargetEntities []string // ip, file, process, service, etc.
	Confidence     float64  // confidence in intent detection
	ExpandedTerms  []string // semantically related terms
}

// analyzeQueryIntent uses modern NLP techniques to understand query intent
func (mse *ModernSearchEngine) analyzeQueryIntent(query string) QueryContext {
	context := QueryContext{
		Intent:     "general",
		Domain:     "general",
		Platform:   "general",
		ActionType: "general",
		Confidence: 0.5,
	}

	queryLower := strings.ToLower(query)
	words := mse.extractMeaningfulWords(query)

	// Intent detection using pattern matching and semantic understanding
	context.Intent = mse.detectIntent(queryLower, words)
	context.Domain = mse.detectDomain(queryLower, words)
	context.Platform = mse.detectPlatform(queryLower, words)
	context.ActionType = mse.detectActionType(queryLower, words)
	context.TargetEntities = mse.extractTargetEntities(queryLower, words)

	// Semantic expansion using our graph
	context.ExpandedTerms = mse.expandQueryTerms(words)

	// Calculate confidence based on how many aspects we detected
	detectedAspects := 0
	if context.Intent != "general" {
		detectedAspects++
	}
	if context.Domain != "general" {
		detectedAspects++
	}
	if context.Platform != "general" {
		detectedAspects++
	}
	if len(context.TargetEntities) > 0 {
		detectedAspects++
	}

	context.Confidence = math.Min(1.0, float64(detectedAspects)*0.25+0.25)

	return context
}

// detectIntent identifies what the user wants to do
func (mse *ModernSearchEngine) detectIntent(query string, words []string) string {
	// Question patterns strongly indicate "find" intent
	questionPatterns := []string{"what", "how", "which command", "what command", "how to", "show me"}
	for _, pattern := range questionPatterns {
		if strings.Contains(query, pattern) {
			return "find"
		}
	}

	// Action word detection
	intentMap := map[string]string{
		"manage":    "configure",
		"configure": "configure",
		"config":    "configure",
		"setup":     "configure",
		"view":      "view",
		"show":      "view",
		"display":   "view",
		"see":       "view",
		"read":      "view",
		"find":      "find",
		"search":    "find",
		"locate":    "find",
		"create":    "create",
		"make":      "create",
		"install":   "install",
		"run":       "execute",
		"execute":   "execute",
		"start":     "execute",
	}

	for _, word := range words {
		if intent, exists := intentMap[word]; exists {
			return intent
		}
	}

	return "general"
}

// detectDomain identifies the technical domain
func (mse *ModernSearchEngine) detectDomain(query string, words []string) string {
	domainKeywords := map[string][]string{
		"network": {"ip", "network", "interface", "connection", "dns", "dhcp", "ping", "route", "adapter"},
		"file":    {"file", "directory", "folder", "path", "document", "content", "text"},
		"process": {"process", "service", "daemon", "task", "job", "running", "pid"},
		"system":  {"system", "os", "hardware", "memory", "cpu", "disk"},
		"package": {"install", "package", "software", "application", "program"},
		"git":     {"git", "repository", "commit", "branch", "version", "control"},
	}

	for domain, keywords := range domainKeywords {
		for _, word := range words {
			for _, keyword := range keywords {
				if word == keyword {
					return domain
				}
			}
		}
	}

	return "general"
}

// detectPlatform identifies the target platform
func (mse *ModernSearchEngine) detectPlatform(query string, words []string) string {
	platformKeywords := map[string][]string{
		"windows": {"windows", "win", "cmd", "powershell", "dos"},
		"linux":   {"linux", "unix", "bash", "shell"},
		"mac":     {"mac", "macos", "darwin", "osx"},
	}

	for platform, keywords := range platformKeywords {
		for _, word := range words {
			for _, keyword := range keywords {
				if word == keyword {
					return platform
				}
			}
		}
	}

	return "general"
}

// detectActionType identifies the type of action
func (mse *ModernSearchEngine) detectActionType(query string, words []string) string {
	actionMap := map[string]string{
		"read":      "read",
		"view":      "read",
		"show":      "read",
		"display":   "read",
		"cat":       "read",
		"write":     "write",
		"create":    "write",
		"make":      "write",
		"edit":      "write",
		"modify":    "write",
		"change":    "write",
		"delete":    "delete",
		"remove":    "delete",
		"destroy":   "delete",
		"control":   "manage",
		"manage":    "manage",
		"configure": "manage",
	}

	for _, word := range words {
		if action, exists := actionMap[word]; exists {
			return action
		}
	}

	return "general"
}

// extractTargetEntities identifies key entities in the query
func (mse *ModernSearchEngine) extractTargetEntities(query string, words []string) []string {
	entities := []string{}

	// Technical entities that are often search targets
	entityKeywords := []string{
		"ip", "network", "interface", "file", "directory", "process", "service",
		"configuration", "config", "password", "user", "permission", "port",
		"connection", "server", "database", "application", "package",
	}

	for _, word := range words {
		for _, entity := range entityKeywords {
			if word == entity {
				entities = append(entities, entity)
			}
		}
	}

	return mse.deduplicateStrings(entities)
}

// semanticRetrieval performs dense semantic retrieval for intent understanding
func (mse *ModernSearchEngine) semanticRetrieval(query string, context QueryContext, options SearchOptions) []SearchResult {
	var results []SearchResult

	// Build enhanced query using context
	enhancedQuery := mse.buildEnhancedQuery(query, context)

	// Score each command based on semantic similarity
	for i, cmd := range mse.db.Commands {
		score := mse.calculateSemanticRelevance(enhancedQuery, context, cmd)

		if score > 0.1 { // Threshold for semantic relevance
			// Apply platform filtering
			if len(cmd.Platform) > 0 {
				currentPlatform := getCurrentPlatform()
				if !isPlatformCompatible(cmd.Platform, currentPlatform) && !isCrossPlatformTool(cmd.Command) {
					continue
				}
			}

			results = append(results, SearchResult{
				Command: &mse.db.Commands[i],
				Score:   score * 100, // Scale to match BM25F scores
			})
		}
	}

	// Sort by semantic relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > options.Limit*2 {
		results = results[:options.Limit*2]
	}

	return results
}

// buildEnhancedQuery creates an enhanced query using semantic expansion
func (mse *ModernSearchEngine) buildEnhancedQuery(query string, context QueryContext) string {
	words := mse.extractMeaningfulWords(query)
	enhanced := make([]string, 0, len(words)+len(context.ExpandedTerms))

	// Add original words with higher weight
	enhanced = append(enhanced, words...)

	// Add semantically related terms
	enhanced = append(enhanced, context.ExpandedTerms...)

	// Add domain-specific terms based on context
	if context.Domain != "general" {
		enhanced = append(enhanced, context.Domain)
	}

	// Add target entities
	enhanced = append(enhanced, context.TargetEntities...)

	return strings.Join(mse.deduplicateStrings(enhanced), " ")
}

// calculateSemanticRelevance computes relevance based on semantic understanding
func (mse *ModernSearchEngine) calculateSemanticRelevance(enhancedQuery string, context QueryContext, cmd Command) float64 {
	var score float64

	// Create searchable text
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))
	queryWords := strings.Fields(strings.ToLower(enhancedQuery))

	// 1. Direct term matching (high weight)
	directMatches := 0
	for _, qword := range queryWords {
		if strings.Contains(cmdText, qword) {
			directMatches++
		}
	}
	score += float64(directMatches) / float64(len(queryWords)) * 0.4

	// 2. Intent-command alignment (high weight)
	intentScore := mse.calculateIntentAlignment(context, cmd)
	score += intentScore * 0.3

	// 3. Domain relevance (medium weight)
	domainScore := mse.calculateDomainRelevance(context, cmd)
	score += domainScore * 0.2

	// 4. Semantic similarity using our graph (medium weight)
	semanticScore := mse.calculateGraphSimilarity(queryWords, cmdText)
	score += semanticScore * 0.1

	return score
}

// calculateIntentAlignment checks if command matches user intent
func (mse *ModernSearchEngine) calculateIntentAlignment(context QueryContext, cmd Command) float64 {
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description)

	// Intent-specific keywords that should appear in relevant commands
	intentKeywords := map[string][]string{
		"find":      {"show", "list", "display", "get", "view", "find", "search"},
		"configure": {"config", "set", "setup", "configure", "manage", "control"},
		"view":      {"cat", "view", "show", "display", "read", "print", "less", "more"},
		"create":    {"create", "make", "new", "generate", "build"},
		"execute":   {"run", "execute", "start", "launch", "invoke"},
		"install":   {"install", "add", "setup", "download"},
	}

	if keywords, exists := intentKeywords[context.Intent]; exists {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(cmdText, keyword) {
				matches++
			}
		}
		return float64(matches) / float64(len(keywords))
	}

	return 0.0
}

// calculateDomainRelevance checks domain-specific relevance
func (mse *ModernSearchEngine) calculateDomainRelevance(context QueryContext, cmd Command) float64 {
	if context.Domain == "general" {
		return 0.0
	}

	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description + " " + strings.Join(cmd.Keywords, " "))

	domainKeywords := map[string][]string{
		"network": {"network", "ip", "interface", "connection", "dns", "dhcp", "ping", "route", "adapter", "ethernet", "wifi"},
		"file":    {"file", "directory", "folder", "path", "document", "content", "text", "read", "write"},
		"process": {"process", "service", "daemon", "task", "job", "running", "pid", "kill", "start", "stop"},
		"system":  {"system", "os", "hardware", "memory", "cpu", "disk", "mount", "device"},
	}

	if keywords, exists := domainKeywords[context.Domain]; exists {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(cmdText, keyword) {
				matches++
			}
		}
		return float64(matches) / float64(len(keywords))
	}

	return 0.0
}

// calculateGraphSimilarity uses the semantic graph for similarity scoring
func (mse *ModernSearchEngine) calculateGraphSimilarity(queryWords []string, cmdText string) float64 {
	var similarity float64

	for _, qword := range queryWords {
		if relatedTerms, exists := mse.semanticGraph[qword]; exists {
			for _, related := range relatedTerms {
				if strings.Contains(cmdText, related) {
					similarity += 0.1 // Small boost for semantic relationships
				}
			}
		}
	}

	return math.Min(similarity, 1.0)
}

// hybridFusion combines sparse and dense retrieval results
func (mse *ModernSearchEngine) hybridFusion(sparseResults, semanticResults []SearchResult, context QueryContext, limit int) []SearchResult {
	// Create a map to track commands and their best scores
	commandScores := make(map[*Command]float64)

	// Weight sparse vs semantic results based on query confidence
	sparseWeight := 0.7 - (context.Confidence * 0.3) // Less sparse weight for high-confidence semantic queries
	semanticWeight := 0.3 + (context.Confidence * 0.3)

	// Add sparse results with weighting
	for _, result := range sparseResults {
		commandScores[result.Command] = result.Score * sparseWeight
	}

	// Add or boost with semantic results
	for _, result := range semanticResults {
		if existingScore, exists := commandScores[result.Command]; exists {
			// Combine scores for commands found by both methods
			commandScores[result.Command] = existingScore + (result.Score * semanticWeight)
		} else {
			// Add new commands found only by semantic search
			commandScores[result.Command] = result.Score * semanticWeight
		}
	}

	// Convert back to SearchResult slice
	var fusedResults []SearchResult
	for cmd, score := range commandScores {
		fusedResults = append(fusedResults, SearchResult{
			Command: cmd,
			Score:   score,
		})
	}

	// Sort by combined score
	sort.Slice(fusedResults, func(i, j int) bool {
		return fusedResults[i].Score > fusedResults[j].Score
	})

	if len(fusedResults) > limit*2 {
		fusedResults = fusedResults[:limit*2]
	}

	return fusedResults
}

// intentBasedReranking applies final intent-based adjustments
func (mse *ModernSearchEngine) intentBasedReranking(results []SearchResult, context QueryContext, limit int) []SearchResult {
	// Apply intent-specific boosts
	for i := range results {
		intentBoost := mse.calculateIntentBoost(context, *results[i].Command)
		results[i].Score *= (1.0 + intentBoost)
	}

	// Final sort
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// calculateIntentBoost provides final intent-based score adjustment
func (mse *ModernSearchEngine) calculateIntentBoost(context QueryContext, cmd Command) float64 {
	cmdText := strings.ToLower(cmd.Command + " " + cmd.Description)

	// Strong boosts for exact intent matches
	if context.Intent == "find" && len(context.TargetEntities) > 0 && len(context.TargetEntities[0]) > 0 {
		if strings.Contains(context.TargetEntities[0], "ip") {
			if strings.Contains(cmd.Command, "ipconfig") || strings.Contains(cmd.Command, "ip") {
				return 0.5 // 50% boost for IP-related commands when looking for IP management
			}
		}
	}

	// Platform-specific boosts
	if context.Platform == "windows" && strings.Contains(cmdText, "windows") {
		return 0.2
	}

	// Domain-specific boosts
	if context.Domain == "network" && (strings.Contains(cmdText, "network") || strings.Contains(cmdText, "ip")) {
		return 0.3
	}

	return 0.0
}

// Helper functions
func (mse *ModernSearchEngine) extractMeaningfulWords(query string) []string {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true,
		"by": true, "for": true, "from": true, "has": true, "he": true, "in": true, "is": true,
		"it": true, "its": true, "of": true, "on": true, "that": true, "the": true, "to": true,
		"was": true, "will": true, "with": true, "what": true, "how": true, "which": true,
		"where": true, "when": true, "why": true, "who": true,
	}

	words := strings.FieldsFunc(strings.ToLower(query), func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	var meaningful []string
	for _, word := range words {
		if len(word) > 1 && !stopWords[word] {
			meaningful = append(meaningful, word)
		}
	}

	return meaningful
}

func (mse *ModernSearchEngine) deduplicateStrings(slice []string) []string {
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

// buildSemanticGraph creates semantic relationships between terms
func (mse *ModernSearchEngine) buildSemanticGraph() {
	// Build a semantic graph of related terms for better query expansion
	mse.semanticGraph = map[string][]string{
		// Network domain
		"ip":        {"network", "interface", "address", "configuration", "config", "ipconfig", "ifconfig"},
		"network":   {"ip", "interface", "connection", "adapter", "ethernet", "wifi"},
		"interface": {"ip", "network", "adapter", "ethernet", "connection"},
		"manage":    {"configure", "control", "setup", "modify", "change", "admin"},
		"configure": {"config", "setup", "manage", "set", "modify"},
		"config":    {"configure", "configuration", "setup", "settings"},

		// File domain
		"file":      {"document", "content", "data", "text"},
		"directory": {"folder", "path", "dir"},
		"folder":    {"directory", "path", "dir"},

		// Action words
		"show":    {"display", "view", "list", "print", "cat"},
		"display": {"show", "view", "list", "print"},
		"view":    {"show", "display", "see", "read", "cat"},
		"find":    {"search", "locate", "discover", "get"},
		"search":  {"find", "locate", "discover", "query"},

		// Platform specific
		"windows": {"win", "cmd", "powershell", "dos"},
		"linux":   {"unix", "bash", "shell"},

		// Command specific expansions
		"command": {"cmd", "tool", "utility", "program"},
	}
}

// expandQueryTerms uses the semantic graph to expand query terms
func (mse *ModernSearchEngine) expandQueryTerms(words []string) []string {
	var expanded []string

	for _, word := range words {
		if related, exists := mse.semanticGraph[word]; exists {
			// Add the most relevant related terms (limit to avoid noise)
			for i, term := range related {
				if i < 3 { // Limit to top 3 related terms per word
					expanded = append(expanded, term)
				}
			}
		}
	}

	return mse.deduplicateStrings(expanded)
}
