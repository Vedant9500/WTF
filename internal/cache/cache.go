// Package cache provides caching functionality for improved performance.
package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type Item struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache provides a thread-safe in-memory cache with expiration and automatic cleanup
type Cache struct {
	items         map[string]*Item
	mutex         sync.RWMutex
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	wg            sync.WaitGroup
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]*Item),
		ttl:   ttl,
	}
}

// NewCacheWithAutoCleanup creates a new cache with automatic background cleanup.
// The cleanup runs at the specified interval and removes expired items.
// Call Stop() to gracefully shut down the cleanup goroutine.
func NewCacheWithAutoCleanup(ttl, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:         make(map[string]*Item),
		ttl:           ttl,
		cleanupTicker: time.NewTicker(cleanupInterval),
		stopCleanup:   make(chan struct{}),
	}

	c.wg.Add(1)
	go c.autoCleanup()

	return c
}

// autoCleanup runs the periodic cleanup in a background goroutine
func (c *Cache) autoCleanup() {
	defer c.wg.Done()

	for {
		select {
		case <-c.cleanupTicker.C:
			c.Cleanup()
		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// Stop gracefully stops the automatic cleanup goroutine.
// This should be called when the cache is no longer needed.
func (c *Cache) Stop() {
	if c.stopCleanup != nil {
		close(c.stopCleanup)
		c.wg.Wait()
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Don't delete here to avoid write lock, let cleanup handle it
		return nil, false
	}

	return item.Value, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &Item{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*Item)
}

// Cleanup removes expired items from the cache
func (c *Cache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// IsAutoCleanupEnabled returns true if automatic cleanup is enabled
func (c *Cache) IsAutoCleanupEnabled() bool {
	return c.cleanupTicker != nil
}
