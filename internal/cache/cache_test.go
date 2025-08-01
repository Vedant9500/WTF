package cache

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewCache(ttl)

	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	if cache.ttl != ttl {
		t.Errorf("Expected TTL %v, got %v", ttl, cache.ttl)
	}

	if cache.items == nil {
		t.Error("Cache items map is nil")
	}

	if len(cache.items) != 0 {
		t.Error("New cache should be empty")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache(time.Hour)

	// Test setting and getting a value
	key := "test_key"
	value := "test_value"

	cache.Set(key, value)

	retrieved, exists := cache.Get(key)
	if !exists {
		t.Error("Expected key to exist")
	}

	if retrieved != value {
		t.Errorf("Expected value %v, got %v", value, retrieved)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := NewCache(time.Hour)

	value, exists := cache.Get("nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}

	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache(50 * time.Millisecond)

	key := "expiring_key"
	value := "expiring_value"

	cache.Set(key, value)

	// Should exist immediately
	retrieved, exists := cache.Get(key)
	if !exists {
		t.Error("Expected key to exist immediately after setting")
	}
	if retrieved != value {
		t.Errorf("Expected value %v, got %v", value, retrieved)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not exist after expiration
	retrieved, exists = cache.Get(key)
	if exists {
		t.Error("Expected key to be expired")
	}
	if retrieved != nil {
		t.Errorf("Expected nil value for expired key, got %v", retrieved)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache(time.Hour)

	key := "delete_key"
	value := "delete_value"

	cache.Set(key, value)

	// Verify it exists
	_, exists := cache.Get(key)
	if !exists {
		t.Error("Expected key to exist before deletion")
	}

	// Delete it
	cache.Delete(key)

	// Verify it's gone
	_, exists = cache.Get(key)
	if exists {
		t.Error("Expected key to not exist after deletion")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(time.Hour)

	// Add multiple items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Clear the cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}

	// Verify items are gone
	_, exists := cache.Get("key1")
	if exists {
		t.Error("Expected key1 to not exist after clear")
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewCache(time.Hour)

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache size 0, got %d", cache.Size())
	}

	cache.Set("key1", "value1")
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	cache.Set("key2", "value2")
	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}

	cache.Delete("key1")
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after deletion, got %d", cache.Size())
	}
}

func TestCacheCleanup(t *testing.T) {
	cache := NewCache(50 * time.Millisecond)

	// Add items that will expire
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Add item that won't expire (we'll update its TTL)
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Before cleanup, expired items are still in the map
	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3 before cleanup, got %d", cache.Size())
	}

	// Run cleanup
	cache.Cleanup()

	// After cleanup, expired items should be removed
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after cleanup, got %d", cache.Size())
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(time.Hour)

	// Test concurrent access
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("key")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we get here without deadlock, the test passes
}

func TestCacheOverwrite(t *testing.T) {
	cache := NewCache(time.Hour)

	key := "overwrite_key"
	value1 := "value1"
	value2 := "value2"

	// Set initial value
	cache.Set(key, value1)

	retrieved, exists := cache.Get(key)
	if !exists || retrieved != value1 {
		t.Errorf("Expected initial value %v, got %v", value1, retrieved)
	}

	// Overwrite with new value
	cache.Set(key, value2)

	retrieved, exists = cache.Get(key)
	if !exists || retrieved != value2 {
		t.Errorf("Expected overwritten value %v, got %v", value2, retrieved)
	}

	// Size should still be 1
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after overwrite, got %d", cache.Size())
	}
}

func TestCacheWithDifferentTypes(t *testing.T) {
	cache := NewCache(time.Hour)

	// Test with different value types
	cache.Set("string", "test")
	cache.Set("int", 42)
	cache.Set("slice", []string{"a", "b", "c"})
	cache.Set("map", map[string]int{"key": 123})

	// Retrieve and verify types
	strVal, exists := cache.Get("string")
	if !exists || strVal.(string) != "test" {
		t.Error("String value not stored/retrieved correctly")
	}

	intVal, exists := cache.Get("int")
	if !exists || intVal.(int) != 42 {
		t.Error("Int value not stored/retrieved correctly")
	}

	sliceVal, exists := cache.Get("slice")
	if !exists {
		t.Error("Slice value not found")
	} else {
		slice := sliceVal.([]string)
		if len(slice) != 3 || slice[0] != "a" {
			t.Error("Slice value not stored/retrieved correctly")
		}
	}

	mapVal, exists := cache.Get("map")
	if !exists {
		t.Error("Map value not found")
	} else {
		m := mapVal.(map[string]int)
		if m["key"] != 123 {
			t.Error("Map value not stored/retrieved correctly")
		}
	}
}