package database

import (
	"github.com/Vedant9500/WTF/internal/cache"
)

// CachedDatabase wraps Database with caching capabilities
type CachedDatabase struct {
	*Database
	cacheManager *cache.CacheManager
}

// NewCachedDatabase creates a new database with caching
func NewCachedDatabase(db *Database) *CachedDatabase {
	return &CachedDatabase{
		Database:     db,
		cacheManager: cache.NewCacheManager(),
	}
}

// SearchWithCache performs search with result caching
func (cdb *CachedDatabase) SearchWithCache(query string, limit int) []SearchResult {
	return cdb.SearchWithOptionsAndCache(query, SearchOptions{Limit: limit})
}

// SearchWithOptionsAndCache performs search with options and caching
func (cdb *CachedDatabase) SearchWithOptionsAndCache(query string, options SearchOptions) []SearchResult {
	searchCache := cdb.cacheManager.GetSearchCache()
	if !cdb.cacheManager.IsEnabled() {
		// Cache disabled, use universal search
		return cdb.SearchUniversal(query, options)
	}

	// Convert SearchOptions to cache.SearchOptions
	cacheOptions := cache.SearchOptions{
		Limit:          options.Limit,
		ContextBoosts:  options.ContextBoosts,
		PipelineOnly:   options.PipelineOnly,
		PipelineBoost:  options.PipelineBoost,
		UseFuzzy:       options.UseFuzzy,
		FuzzyThreshold: options.FuzzyThreshold,
		UseNLP:         options.UseNLP,
	}

	// Try to get from cache first
	if cachedResults, found := searchCache.Get(query, cacheOptions); found {
		return convertCacheResults(cachedResults)
	}

	// Cache miss - perform actual search
	results := cdb.SearchUniversal(query, options)

	// Store in cache
	if len(results) > 0 {
		searchCache.Put(query, cacheOptions, convertDBResults(results))
	}

	return results
}

// InvalidateCache clears all cached results
func (cdb *CachedDatabase) InvalidateCache() {
	cdb.cacheManager.InvalidateAll()
}

// EnableCache enables or disables caching
func (cdb *CachedDatabase) EnableCache(enabled bool) {
	cdb.cacheManager.Enable(enabled)
}

// IsCacheEnabled returns whether caching is enabled
func (cdb *CachedDatabase) IsCacheEnabled() bool {
	return cdb.cacheManager.IsEnabled()
}

// GetCacheStats returns cache statistics
func (cdb *CachedDatabase) GetCacheStats() map[string]cache.CacheStats {
	return cdb.cacheManager.GetStats()
}

// CleanupExpiredCache removes expired cache entries
func (cdb *CachedDatabase) CleanupExpiredCache() map[string]int {
	return cdb.cacheManager.CleanupExpired()
}

// UpdateDatabase updates the underlying database and invalidates cache
func (cdb *CachedDatabase) UpdateDatabase(commands []Command) {
	cdb.Database.Commands = commands
	cdb.Database.BuildUniversalIndex() // Rebuild universal index
	cdb.InvalidateCache()              // Invalidate cache when database is updated
}

// SearchWithPipelineOptionsAndCache performs pipeline search with caching
func (cdb *CachedDatabase) SearchWithPipelineOptionsAndCache(query string, options SearchOptions) []SearchResult {
	searchCache := cdb.cacheManager.GetSearchCache()
	if !cdb.cacheManager.IsEnabled() {
		return cdb.SearchUniversal(query, options)
	}

	// Convert SearchOptions to cache.SearchOptions
	cacheOptions := cache.SearchOptions{
		Limit:          options.Limit,
		ContextBoosts:  options.ContextBoosts,
		PipelineOnly:   options.PipelineOnly,
		PipelineBoost:  options.PipelineBoost,
		UseFuzzy:       options.UseFuzzy,
		FuzzyThreshold: options.FuzzyThreshold,
		UseNLP:         options.UseNLP,
	}

	// Try cache first
	if cachedResults, found := searchCache.Get(query, cacheOptions); found {
		return convertCacheResults(cachedResults)
	}

	// Cache miss - perform search
	results := cdb.SearchUniversal(query, options)

	// Store in cache
	if len(results) > 0 {
		searchCache.Put(query, cacheOptions, convertDBResults(results))
	}

	return results
}

// SearchWithFuzzyAndCache performs fuzzy search with caching
func (cdb *CachedDatabase) SearchWithFuzzyAndCache(query string, options SearchOptions) []SearchResult {
	searchCache := cdb.cacheManager.GetSearchCache()
	if !cdb.cacheManager.IsEnabled() {
		return cdb.SearchUniversal(query, options)
	}

	cacheOptions := cache.SearchOptions{
		Limit:          options.Limit,
		ContextBoosts:  options.ContextBoosts,
		PipelineOnly:   options.PipelineOnly,
		PipelineBoost:  options.PipelineBoost,
		UseFuzzy:       options.UseFuzzy,
		FuzzyThreshold: options.FuzzyThreshold,
		UseNLP:         options.UseNLP,
	}

	if cachedResults, found := searchCache.Get(query, cacheOptions); found {
		return convertCacheResults(cachedResults)
	}

	results := cdb.SearchUniversal(query, options)

	if len(results) > 0 {
		searchCache.Put(query, cacheOptions, convertDBResults(results))
	}

	return results
}

// convertCacheResults converts cached results to database results.
func convertCacheResults(cached []cache.SearchResult) []SearchResult {
	results := make([]SearchResult, 0, len(cached))
	for _, c := range cached {
		if cmd, ok := c.Command.(*Command); ok {
			results = append(results, SearchResult{Command: cmd, Score: c.Score})
		}
	}
	return results
}

// convertDBResults converts database results to cache-friendly results.
func convertDBResults(results []SearchResult) []cache.SearchResult {
	out := make([]cache.SearchResult, len(results))
	for i, r := range results {
		out[i] = cache.SearchResult{Command: r.Command, Score: r.Score}
	}
	return out
}
