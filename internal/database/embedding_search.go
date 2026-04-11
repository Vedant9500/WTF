package database

import (
	"math"
	"sort"
	"strings"

	"github.com/Vedant9500/WTF/internal/embedding"
	"github.com/Vedant9500/WTF/internal/nlp"
	"github.com/Vedant9500/WTF/internal/queryexpansion"
)

// embeddingSearcher implements embedding-based search with ANN indexing.
type embeddingSearcher struct {
	index      *embedding.EnhancedIndex
	nlpProc    *nlp.QueryProcessor
	useANN     bool
	topK       int
	minScore   float64
}

// newEmbeddingSearcher creates a new embedding-based searcher.
func newEmbeddingSearcher(enhancedIdx *embedding.EnhancedIndex, nlpProcessor *nlp.QueryProcessor) *embeddingSearcher {
	return &embeddingSearcher{
		index:    enhancedIdx,
		nlpProc: nlpProcessor,
		useANN:   true,
		topK:     50, // Return top 50 candidates for further ranking
		minScore: 0.01, // Minimum cosine similarity threshold (raw scale: -1 to 1)
	}
}

// EmbeddingSearchResult represents a search result with metadata.
type EmbeddingSearchResult struct {
	CommandIndex int
	Score        float64
	Similarity   float64
	FieldScores  map[string]float64
	QueryTerms   []string
}

// Search performs embedding-based search with NLP enhancement and query expansion.
func (es *embeddingSearcher) Search(query string, limit int) []EmbeddingSearchResult {
	if es.index == nil || query == "" {
		return nil
	}

	// Step 1: Process query with NLP
	processedQuery := es.nlpProc.ProcessQuery(query)
	queryTerms := es.extractQueryTerms(processedQuery)

	if len(queryTerms) == 0 {
		return nil
	}

	// Step 2: Expand query with domain synonyms for better recall
	expandedTerms := queryexpansion.ExpandQueryTerms(queryTerms)

	// Step 3: Compute intent weights for query embedding
	intentWeights := es.computeIntentWeights(processedQuery)

	// Step 4: Generate query embedding (use expanded query for better representation)
	expandedQuery := strings.Join(expandedTerms, " ")
	queryEmbedding := es.index.EmbedQueryWithIntent(expandedQuery, intentWeights)
	if queryEmbedding == nil {
		// Fallback to original query if expanded query fails
		queryEmbedding = es.index.EmbedQueryWithIntent(query, intentWeights)
		if queryEmbedding == nil {
			return nil
		}
	}

	// Step 5: Search for similar commands
	candidates := es.index.EnhancedSearch(queryEmbedding, es.topK, es.useANN)

	// Step 6: Score and rank candidates
	results := es.scoreCandidates(candidates, expandedTerms, processedQuery)

	// Step 7: Apply post-scoring boosts
	results = es.applyEmbeddingBoosts(results, processedQuery)

	// Step 8: Filter and limit results
	results = es.filterAndLimit(results, limit)

	return results
}

// extractQueryTerms extracts meaningful terms from processed query.
func (es *embeddingSearcher) extractQueryTerms(pq *nlp.ProcessedQuery) []string {
	termSet := make(map[string]bool)

	// Add actions
	for _, action := range pq.Actions {
		if len(action) >= 2 {
			termSet[action] = true
		}
	}

	// Add targets
	for _, target := range pq.Targets {
		if len(target) >= 2 {
			termSet[target] = true
		}
	}

	// Add keywords
	for _, keyword := range pq.Keywords {
		if len(keyword) >= 2 {
			termSet[keyword] = true
		}
	}

	// Convert to slice
	terms := make([]string, 0, len(termSet))
	for term := range termSet {
		terms = append(terms, term)
	}

	return terms
}

// computeIntentWeights computes weights for each term based on intent.
func (es *embeddingSearcher) computeIntentWeights(pq *nlp.ProcessedQuery) map[string]float64 {
	weights := make(map[string]float64)

	// Actions get highest weight
	for _, action := range pq.Actions {
		weights[action] = 1.0
	}

	// Targets get medium weight
	for _, target := range pq.Targets {
		if _, exists := weights[target]; !exists {
			weights[target] = 0.7
		}
	}

	// Keywords get base weight
	for _, keyword := range pq.Keywords {
		if _, exists := weights[keyword]; !exists {
			weights[keyword] = 0.5
		}
	}

	// Boost intent-specific terms
	switch pq.Intent {
	case nlp.IntentFind:
		// Boost location/search terms
		for term := range weights {
			if strings.Contains(term, "find") || strings.Contains(term, "search") {
				weights[term] *= 1.2
			}
		}
	case nlp.IntentCreate:
		// Boost creation terms
		for term := range weights {
			if strings.Contains(term, "create") || strings.Contains(term, "make") {
				weights[term] *= 1.2
			}
		}
	case nlp.IntentView:
		// Boost viewing terms
		for term := range weights {
			if strings.Contains(term, "view") || strings.Contains(term, "show") {
				weights[term] *= 1.2
			}
		}
	}

	return weights
}

// scoreCandidates scores candidates based on query terms and NLP information.
func (es *embeddingSearcher) scoreCandidates(
	candidates []embedding.SearchResult,
	queryTerms []string,
	pq *nlp.ProcessedQuery,
) []EmbeddingSearchResult {
	results := make([]EmbeddingSearchResult, 0, len(candidates))

	for _, candidate := range candidates {
		if candidate.Score < es.minScore {
			continue
		}

		result := EmbeddingSearchResult{
			CommandIndex: candidate.CommandIndex,
			Score:        candidate.Score,
			Similarity:   candidate.Similarity,
			FieldScores:  make(map[string]float64),
			QueryTerms:   queryTerms,
		}

		// Compute field-aware scores if available
		if len(es.index.CmdFieldEmbeds) > candidate.CommandIndex {
			queryEmbedding := es.index.EmbedQueryWithIntent(pq.Original, nil)
			if queryEmbedding != nil {
				fieldScore := es.index.ComputeFieldAwareScore(queryEmbedding, candidate.CommandIndex)
				if fieldScore > 0 {
					result.FieldScores["field_aware"] = fieldScore
					// Blend base score with field-aware score (both now on same scale)
					result.Score = candidate.Score*0.6 + fieldScore*0.4
				}
			}
		}

		results = append(results, result)
	}

	return results
}

// applyEmbeddingBoosts applies additional boosts specific to embedding search.
func (es *embeddingSearcher) applyEmbeddingBoosts(
	results []EmbeddingSearchResult,
	pq *nlp.ProcessedQuery,
) []EmbeddingSearchResult {
	for i := range results {
		cmdIdx := results[i].CommandIndex
		if cmdIdx >= len(es.index.CmdMetadata) {
			continue
		}

		meta := es.index.CmdMetadata[cmdIdx]
		boost := 1.0

		// Boost exact command name matches
		cmdName := strings.ToLower(meta.Command)
		for _, term := range pq.Actions {
			if strings.Contains(cmdName, strings.ToLower(term)) {
				boost *= 1.3
			}
		}

		for _, term := range pq.Targets {
			if strings.Contains(cmdName, strings.ToLower(term)) {
				boost *= 1.2
			}
		}

		// Boost pipeline commands if query suggests it
		if meta.IsPipeline && pq.Intent == nlp.IntentRun {
			boost *= 1.1
		}

		// Apply niche boost if relevant
		if meta.Niche != "" {
			for _, keyword := range pq.Keywords {
				if strings.Contains(strings.ToLower(meta.Niche), strings.ToLower(keyword)) {
					boost *= 1.15
					break
				}
			}
		}

		results[i].Score *= boost
	}

	return results
}

// filterAndLimit filters results below threshold and limits to top K.
func (es *embeddingSearcher) filterAndLimit(results []EmbeddingSearchResult, limit int) []EmbeddingSearchResult {
	// Filter by minimum score
	filtered := make([]EmbeddingSearchResult, 0, len(results))
	for _, result := range results {
		if result.Score >= es.minScore {
			filtered = append(filtered, result)
		}
	}

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// HybridSearch combines embedding-based search with BM25F for optimal results.
func (db *Database) HybridSearch(query string, limit int, useNLP bool) []SearchResult {
	// Try embedding search first if available
	if db.enhancedEmbeddingIndex != nil {
		embeddingResults := db.embeddingSearch(query, limit)
		if len(embeddingResults) > 0 {
			// If we have good embedding results, blend with BM25F
			if db.uIndex != nil {
				bm25Results := db.searchBM25F(query, limit*2, useNLP)
				return db.hybridFuseResults(embeddingResults, bm25Results, limit)
			}
			return db.convertEmbeddingToSearchResults(embeddingResults)
		}
	}

	// Fallback to BM25F
	if db.uIndex != nil {
		return db.searchBM25F(query, limit, useNLP)
	}

	return nil
}

// embeddingSearch performs pure embedding-based search.
func (db *Database) embeddingSearch(query string, limit int) []EmbeddingSearchResult {
	if db.embeddingSearcher == nil || db.enhancedEmbeddingIndex == nil {
		return nil
	}

	return db.embeddingSearcher.Search(query, limit)
}

// searchBM25F performs BM25F-based search (wrapper for existing implementation).
func (db *Database) searchBM25F(query string, limit int, useNLP bool) []SearchResult {
	opts := SearchOptions{
		Limit:   limit,
		UseNLP:  useNLP,
		UseFuzzy: true,
	}
	return db.SearchUniversal(query, opts)
}

// hybridFuseResults fuses embedding and BM25F results using Reciprocal Rank Fusion (RRF).
// RRF formula: score(doc) = Σ 1/(k + rank_i) where k=60 (empirically optimal)
func (db *Database) hybridFuseResults(
	embeddingResults []EmbeddingSearchResult,
	bm25Results []SearchResult,
	limit int,
) []SearchResult {
	// RRF constant (research shows k=60 is optimal for most use cases)
	k := 60.0

	// Track fused scores and which list each doc came from
	scoreMap := make(map[int]float64)
	embedRanks := make(map[int]int)
	bm25Ranks := make(map[int]int)

	// Score embedding results using RRF
	for rank, result := range embeddingResults {
		reciprocalRank := 1.0 / (k + float64(rank+1))
		scoreMap[result.CommandIndex] += reciprocalRank
		embedRanks[result.CommandIndex] = rank + 1
	}

	// Score BM25F results using RRF
	for rank, result := range bm25Results {
		if cmdIdx, ok := db.cmdIndex[result.Command]; ok {
			reciprocalRank := 1.0 / (k + float64(rank+1))
			scoreMap[cmdIdx] += reciprocalRank
			bm25Ranks[cmdIdx] = rank + 1
		}
	}

	// Convert back to results sorted by RRF score
	type scoredResult struct {
		cmdIdx     int
		rrfScore   float64
		embedRank  int
		bm25Rank   int
	}

	fused := make([]scoredResult, 0, len(scoreMap))
	for cmdIdx, rrfScore := range scoreMap {
		fused = append(fused, scoredResult{
			cmdIdx:    cmdIdx,
			rrfScore:  rrfScore,
			embedRank: embedRanks[cmdIdx],
			bm25Rank:  bm25Ranks[cmdIdx],
		})
	}

	// Sort by RRF score descending
	sort.Slice(fused, func(i, j int) bool {
		return fused[i].rrfScore > fused[j].rrfScore
	})

	// Convert to SearchResults and limit
	results := make([]SearchResult, 0, limit)
	for i := 0; i < len(fused) && i < limit; i++ {
		cmdIdx := fused[i].cmdIdx
		if cmdIdx < len(db.Commands) {
			cmd := &db.Commands[cmdIdx]
			// Normalize RRF score to 0-100 range for consistency
			// Max possible RRF score = 2/(k+1) ≈ 0.0328, so multiply by ~3000
			normalizedScore := math.Min(fused[i].rrfScore*3000.0, 100.0)
			results = append(results, SearchResult{
				Command: cmd,
				Score:   normalizedScore,
			})
		}
	}

	return results
}

// convertEmbeddingToSearchResults converts embedding results to standard format.
func (db *Database) convertEmbeddingToSearchResults(
	embeddingResults []EmbeddingSearchResult,
) []SearchResult {
	results := make([]SearchResult, 0, len(embeddingResults))

	for _, er := range embeddingResults {
		if er.CommandIndex < len(db.Commands) {
			cmd := &db.Commands[er.CommandIndex]
			results = append(results, SearchResult{
				Command: cmd,
				Score:   er.Score,
			})
		}
	}

	return results
}

// Initialize embedding searcher for the database.
func (db *Database) initializeEmbeddingSearcher() {
	if db.enhancedEmbeddingIndex != nil && db.nlpProcessor != nil {
		db.embeddingSearcher = newEmbeddingSearcher(db.enhancedEmbeddingIndex, db.nlpProcessor)
	}
}
