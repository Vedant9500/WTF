package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/constants"
)

// SearchResult represents a cached search result (avoiding circular import)
type SearchResult struct {
	Command interface{} `json:"command"`
	Score   float64     `json:"score"`
}

// SearchOptions represents search options for cache key generation
type SearchOptions struct {
	Limit          int                `json:"limit"`
	ContextBoosts  map[string]float64 `json:"context_boosts,omitempty"`
	PipelineOnly   bool               `json:"pipeline_only,omitempty"`
	PipelineBoost  float64            `json:"pipeline_boost,omitempty"`
	UseFuzzy       bool               `json:"use_fuzzy,omitempty"`
	FuzzyThreshold int                `json:"fuzzy_threshold,omitempty"`
	UseNLP         bool               `json:"use_nlp,omitempty"`
}

// SearchCache provides caching for search results
type SearchCache struct {
	cache     *LRUCache
	enabled   bool
	keyPrefix string
}

// NewSearchCache creates a new search result cache
func NewSearchCache(capacity int, ttl time.Duration) *SearchCache {
	return &SearchCache{
		cache:     NewLRUCache(capacity, ttl),
		enabled:   true,
		keyPrefix: "search:",
	}
}

// Get retrieves cached search results
func (sc *SearchCache) Get(query string, options SearchOptions) ([]SearchResult, bool) {
	if !sc.enabled {
		return nil, false
	}

	key := sc.generateCacheKey(query, options)
	if value, found := sc.cache.Get(key); found {
		if results, ok := value.([]SearchResult); ok {
			return results, true
		}
	}

	return nil, false
}

// Put stores search results in cache
func (sc *SearchCache) Put(query string, options SearchOptions, results []SearchResult) {
	if !sc.enabled || len(results) == 0 {
		return
	}

	key := sc.generateCacheKey(query, options)

	// Create a copy of results to avoid reference issues
	cachedResults := make([]SearchResult, len(results))
	copy(cachedResults, results)

	sc.cache.Put(key, cachedResults)
}

// Invalidate removes all cached results (called when database is updated)
func (sc *SearchCache) Invalidate() {
	sc.cache.Clear()
}

// InvalidatePattern removes cached results matching a pattern
func (sc *SearchCache) InvalidatePattern(pattern string) int {
	keys := sc.cache.Keys()
	removed := 0

	for _, key := range keys {
		if strings.Contains(key, pattern) {
			if sc.cache.Delete(key) {
				removed++
			}
		}
	}

	return removed
}

// Enable enables or disables the cache
func (sc *SearchCache) Enable(enabled bool) {
	sc.enabled = enabled
}

// IsEnabled returns whether the cache is enabled
func (sc *SearchCache) IsEnabled() bool {
	return sc.enabled
}

// Stats returns cache statistics
func (sc *SearchCache) Stats() CacheStats {
	return sc.cache.Stats()
}

// Size returns the current cache size
func (sc *SearchCache) Size() int {
	return sc.cache.Size()
}

// CleanupExpired removes expired entries
func (sc *SearchCache) CleanupExpired() int {
	return sc.cache.CleanupExpired()
}

// generateCacheKey creates a unique cache key for the query and options
func (sc *SearchCache) generateCacheKey(query string, options SearchOptions) string {
	// Normalize query for consistent caching
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	// Create a deterministic key that includes all relevant options
	keyData := struct {
		Query   string        `json:"query"`
		Options SearchOptions `json:"options"`
	}{
		Query:   normalizedQuery,
		Options: options,
	}

	// Serialize to JSON for consistent key generation
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		// Fallback to simple key if JSON marshaling fails
		return fmt.Sprintf("%s%s:%d", sc.keyPrefix, normalizedQuery, options.Limit)
	}

	// Generate SHA256 hash for compact key (more secure than MD5)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%s%x", sc.keyPrefix, hash)
}

// CacheManager manages multiple cache instances
type CacheManager struct {
	searchCache *SearchCache
	enabled     bool
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		searchCache: NewSearchCache(
			constants.DefaultCacheCapacity,
			constants.DefaultCacheTTL,
		),
		enabled: true,
	}
}

// GetSearchCache returns the search cache instance
func (cm *CacheManager) GetSearchCache() *SearchCache {
	return cm.searchCache
}

// Enable enables or disables all caches
func (cm *CacheManager) Enable(enabled bool) {
	cm.enabled = enabled
	cm.searchCache.Enable(enabled)
}

// IsEnabled returns whether caching is enabled
func (cm *CacheManager) IsEnabled() bool {
	return cm.enabled
}

// InvalidateAll clears all caches
func (cm *CacheManager) InvalidateAll() {
	cm.searchCache.Invalidate()
}

// GetStats returns statistics for all caches
func (cm *CacheManager) GetStats() map[string]CacheStats {
	return map[string]CacheStats{
		"search": cm.searchCache.Stats(),
	}
}

// CleanupExpired removes expired entries from all caches
func (cm *CacheManager) CleanupExpired() map[string]int {
	return map[string]int{
		"search": cm.searchCache.CleanupExpired(),
	}
}
