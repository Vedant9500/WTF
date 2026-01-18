package database

import (
	"github.com/Vedant9500/WTF/internal/nlp"
)

// applyPostScoringBoosts applies NLP reranking, cascading boosts, and semantic boosts
func (db *Database) applyPostScoringBoosts(
	results []SearchResult, pq *nlp.ProcessedQuery, query string, options SearchOptions,
) []SearchResult {
	// Optional NLP-based reranking
	if options.UseNLP && db.tfidf != nil {
		results = db.rerankWithNLP(results, query, options)
	}

	// Cascading boost: apply weighted boosts based on query token types
	// This replaces word vector semantic search with more targeted boosting
	if options.UseNLP && pq != nil && len(results) > 0 {
		results = db.cascadingBoost(results, pq)
	}

	// Semantic boost: blend embedding similarity into scores if embeddings are loaded
	if db.HasEmbeddings() && len(results) > 0 {
		results = db.applySemanticBoost(results, query)
	}

	return results
}
