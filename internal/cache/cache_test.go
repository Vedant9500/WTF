package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestLRUCache_BasicOperations(t *testing.T) {
	cache := NewLRUCache(3, 0) // No TTL
	
	// Test Put and Get
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")
	
	if value, found := cache.Get("key1"); !found || value != "value1" {
		t.Errorf("Expected to find key1 with value1, got %v, %v", value, found)
	}
	
	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2, 0) // Capacity 2, no TTL
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3") // Should evict key1
	
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be evicted")
	}
	
	if value, found := cache.Get("key2"); !found || value != "value2" {
		t.Errorf("Expected to find key2, got %v, %v", value, found)
	}
	
	if value, found := cache.Get("key3"); !found || value != "value3" {
		t.Errorf("Expected to find key3, got %v, %v", value, found)
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(2, 0)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	
	// Access key1 to make it most recently used
	cache.Get("key1")
	
	// Add key3, should evict key2 (least recently used)
	cache.Put("key3", "value3")
	
	if _, found := cache.Get("key2"); found {
		t.Error("Expected key2 to be evicted")
	}
	
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected key1 to still be in cache")
	}
	
	if _, found := cache.Get("key3"); !found {
		t.Error("Expected key3 to be in cache")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(10, 50*time.Millisecond)
	
	cache.Put("key1", "value1")
	
	// Should be found immediately
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected to find key1 immediately")
	}
	
	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)
	
	// Should not be found after TTL
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be expired")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(10, 0)
	
	cache.Put("key1", "value1")
	cache.Put("key1", "value2") // Update
	
	if value, found := cache.Get("key1"); !found || value != "value2" {
		t.Errorf("Expected updated value2, got %v, %v", value, found)
	}
	
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after update, got %d", cache.Size())
	}
}

func TestLRUCache_Delete(t *testing.T) {
	cache := NewLRUCache(10, 0)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	
	if !cache.Delete("key1") {
		t.Error("Expected Delete to return true for existing key")
	}
	
	if cache.Delete("key1") {
		t.Error("Expected Delete to return false for non-existing key")
	}
	
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be deleted")
	}
	
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after delete, got %d", cache.Size())
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10, 0)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
	
	if _, found := cache.Get("key1"); found {
		t.Error("Expected no keys after clear")
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(10, 0)
	
	cache.Put("key1", "value1")
	cache.Get("key1")  // Hit
	cache.Get("key2")  // Miss
	
	stats := cache.Stats()
	
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
	
	if stats.Capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", stats.Capacity)
	}
	
	expectedHitRatio := 0.5 // 1 hit out of 2 total
	if stats.HitRatio != expectedHitRatio {
		t.Errorf("Expected hit ratio %.2f, got %.2f", expectedHitRatio, stats.HitRatio)
	}
}

func TestLRUCache_CleanupExpired(t *testing.T) {
	cache := NewLRUCache(10, 50*time.Millisecond)
	
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	
	// Wait for some entries to expire
	time.Sleep(60 * time.Millisecond)
	
	cache.Put("key3", "value3") // Fresh entry
	
	removed := cache.CleanupExpired()
	
	if removed != 2 {
		t.Errorf("Expected 2 expired entries removed, got %d", removed)
	}
	
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after cleanup, got %d", cache.Size())
	}
	
	if _, found := cache.Get("key3"); !found {
		t.Error("Expected key3 to remain after cleanup")
	}
}

func TestSearchCache_BasicOperations(t *testing.T) {
	searchCache := NewSearchCache(10, 0)
	
	query := "test query"
	options := SearchOptions{Limit: 5}
	results := []SearchResult{
		{Command: "test command", Score: 1.0},
	}
	
	// Test Put and Get
	searchCache.Put(query, options, results)
	
	if cachedResults, found := searchCache.Get(query, options); !found {
		t.Error("Expected to find cached results")
	} else if len(cachedResults) != 1 {
		t.Errorf("Expected 1 cached result, got %d", len(cachedResults))
	} else if cachedResults[0].Score != 1.0 {
		t.Errorf("Expected score 1.0, got %f", cachedResults[0].Score)
	}
}

func TestSearchCache_KeyGeneration(t *testing.T) {
	searchCache := NewSearchCache(10, 0)
	
	query1 := "test query"
	query2 := "TEST QUERY" // Different case
	options := SearchOptions{Limit: 5}
	results := []SearchResult{{Command: "test", Score: 1.0}}
	
	searchCache.Put(query1, options, results)
	
	// Should find with different case (normalized)
	if _, found := searchCache.Get(query2, options); !found {
		t.Error("Expected to find cached results with different case")
	}
	
	// Different options should not match
	differentOptions := SearchOptions{Limit: 10}
	if _, found := searchCache.Get(query1, differentOptions); found {
		t.Error("Expected not to find cached results with different options")
	}
}

func TestSearchCache_Invalidation(t *testing.T) {
	searchCache := NewSearchCache(10, 0)
	
	query := "test query"
	options := SearchOptions{Limit: 5}
	results := []SearchResult{{Command: "test", Score: 1.0}}
	
	searchCache.Put(query, options, results)
	
	// Verify it's cached
	if _, found := searchCache.Get(query, options); !found {
		t.Error("Expected to find cached results before invalidation")
	}
	
	// Invalidate
	searchCache.Invalidate()
	
	// Should not find after invalidation
	if _, found := searchCache.Get(query, options); found {
		t.Error("Expected not to find cached results after invalidation")
	}
}

func TestCacheManager(t *testing.T) {
	manager := NewCacheManager()
	
	if !manager.IsEnabled() {
		t.Error("Expected cache manager to be enabled by default")
	}
	
	searchCache := manager.GetSearchCache()
	if searchCache == nil {
		t.Error("Expected search cache to be available")
	}
	
	// Test enable/disable
	manager.Enable(false)
	if manager.IsEnabled() {
		t.Error("Expected cache manager to be disabled")
	}
	
	if searchCache.IsEnabled() {
		t.Error("Expected search cache to be disabled")
	}
	
	// Test stats
	stats := manager.GetStats()
	if _, exists := stats["search"]; !exists {
		t.Error("Expected search cache stats to be available")
	}
}

// Benchmark tests
func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(1000, 0)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Put(key, "value")
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, 0)
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Put(key, "value")
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Get(key)
	}
}

func BenchmarkSearchCache_Put(b *testing.B) {
	searchCache := NewSearchCache(1000, 0)
	options := SearchOptions{Limit: 10}
	results := []SearchResult{{Command: "test", Score: 1.0}}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("query%d", i%1000)
		searchCache.Put(query, options, results)
	}
}

func BenchmarkSearchCache_Get(b *testing.B) {
	searchCache := NewSearchCache(1000, 0)
	options := SearchOptions{Limit: 10}
	results := []SearchResult{{Command: "test", Score: 1.0}}
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		query := fmt.Sprintf("query%d", i)
		searchCache.Put(query, options, results)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		query := fmt.Sprintf("query%d", i%1000)
		searchCache.Get(query, options)
	}
}