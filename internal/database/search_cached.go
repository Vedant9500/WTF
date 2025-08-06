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
	if !cdb.cacheManager.IsEnabled() {
		// Cache disabled, use regular search
		return cdb.OptimizedSearchWithOptions(query, options)
	}
	
	searchCache := cdb.cacheManager.GetSearchCache()
	
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
		// Convert cache.SearchResult back to database.SearchResult
		results := make([]SearchResult, len(cachedResults))
		for i, cached := range cachedResults {
			if cmd, ok := cached.Command.(*Command); ok {
				results[i] = SearchResult{
					Command: cmd,
					Score:   cached.Score,
				}
			}
		}
		return results
	}
	
	// Cache miss - perform actual search
	results := cdb.OptimizedSearchWithOptions(query, options)
	
	// Store in cache
	if len(results) > 0 {
		// Convert database.SearchResult to cache.SearchResult
		cacheResults := make([]cache.SearchResult, len(results))
		for i, result := range results {
			cacheResults[i] = cache.SearchResult{
				Command: result.Command,
				Score:   result.Score,
			}
		}
		searchCache.Put(query, cacheOptions, cacheResults)
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
	cdb.InvalidateCache() // Invalidate cache when database is updated
}

// SearchWithPipelineOptionsAndCache performs pipeline search with caching
func (cdb *CachedDatabase) SearchWithPipelineOptionsAndCache(query string, options SearchOptions) []SearchResult {
	if !cdb.cacheManager.IsEnabled() {
		return cdb.SearchWithPipelineOptions(query, options)
	}
	
	searchCache := cdb.cacheManager.GetSearchCache()
	
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
		results := make([]SearchResult, len(cachedResults))
		for i, cached := range cachedResults {
			if cmd, ok := cached.Command.(*Command); ok {
				results[i] = SearchResult{
					Command: cmd,
					Score:   cached.Score,
				}
			}
		}
		return results
	}
	
	// Cache miss - perform search
	results := cdb.SearchWithPipelineOptions(query, options)
	
	// Store in cache
	if len(results) > 0 {
		cacheResults := make([]cache.SearchResult, len(results))
		for i, result := range results {
			cacheResults[i] = cache.SearchResult{
				Command: result.Command,
				Score:   result.Score,
			}
		}
		searchCache.Put(query, cacheOptions, cacheResults)
	}
	
	return results
}

// SearchWithFuzzyAndCache performs fuzzy search with caching
func (cdb *CachedDatabase) SearchWithFuzzyAndCache(query string, options SearchOptions) []SearchResult {
	if !cdb.cacheManager.IsEnabled() {
		return cdb.SearchWithFuzzy(query, options)
	}
	
	searchCache := cdb.cacheManager.GetSearchCache()
	
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
		results := make([]SearchResult, len(cachedResults))
		for i, cached := range cachedResults {
			if cmd, ok := cached.Command.(*Command); ok {
				results[i] = SearchResult{
					Command: cmd,
					Score:   cached.Score,
				}
			}
		}
		return results
	}
	
	results := cdb.SearchWithFuzzy(query, options)
	
	if len(results) > 0 {
		cacheResults := make([]cache.SearchResult, len(results))
		for i, result := range results {
			cacheResults[i] = cache.SearchResult{
				Command: result.Command,
				Score:   result.Score,
			}
		}
		searchCache.Put(query, cacheOptions, cacheResults)
	}
	
	return results
}