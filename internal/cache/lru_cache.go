package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Key        string
	Value      interface{}
	CreatedAt  time.Time
	AccessedAt time.Time
	AccessCount int64
}

// LRUCache implements a thread-safe LRU cache with TTL support
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	evictList *list.List
	
	// Metrics
	hits   int64
	misses int64
	evictions int64
}

// NewLRUCache creates a new LRU cache with specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	
	return &LRUCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	element, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}
	
	entry := element.Value.(*CacheEntry)
	
	// Check TTL expiration
	if c.ttl > 0 && time.Since(entry.CreatedAt) > c.ttl {
		c.removeElement(element)
		c.misses++
		return nil, false
	}
	
	// Update access information
	entry.AccessedAt = time.Now()
	entry.AccessCount++
	
	// Move to front (most recently used)
	c.evictList.MoveToFront(element)
	
	c.hits++
	return entry.Value, true
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	
	// Check if key already exists
	if element, exists := c.items[key]; exists {
		// Update existing entry
		entry := element.Value.(*CacheEntry)
		entry.Value = value
		entry.AccessedAt = now
		entry.AccessCount++
		c.evictList.MoveToFront(element)
		return
	}
	
	// Create new entry
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		AccessedAt:  now,
		AccessCount: 1,
	}
	
	// Add to front of list
	element := c.evictList.PushFront(entry)
	c.items[key] = element
	
	// Check if we need to evict
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if element, exists := c.items[key]; exists {
		c.removeElement(element)
		return true
	}
	return false
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*list.Element)
	c.evictList.Init()
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// Size returns the current number of items in the cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Capacity returns the maximum capacity of the cache
func (c *LRUCache) Capacity() int {
	return c.capacity
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	total := c.hits + c.misses
	var hitRatio float64
	if total > 0 {
		hitRatio = float64(c.hits) / float64(total)
	}
	
	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      len(c.items),
		Capacity:  c.capacity,
		HitRatio:  hitRatio,
	}
}

// Keys returns all keys in the cache (for debugging)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// CleanupExpired removes expired entries from the cache
func (c *LRUCache) CleanupExpired() int {
	if c.ttl <= 0 {
		return 0 // No TTL configured
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	removed := 0
	
	// Walk from back to front (oldest to newest)
	for element := c.evictList.Back(); element != nil; {
		entry := element.Value.(*CacheEntry)
		
		if now.Sub(entry.CreatedAt) > c.ttl {
			next := element.Prev()
			c.removeElement(element)
			removed++
			element = next
		} else {
			// Since list is ordered by access time, we can break early
			break
		}
	}
	
	return removed
}

// evictOldest removes the least recently used item
func (c *LRUCache) evictOldest() {
	element := c.evictList.Back()
	if element != nil {
		c.removeElement(element)
		c.evictions++
	}
}

// removeElement removes an element from both the list and map
func (c *LRUCache) removeElement(element *list.Element) {
	c.evictList.Remove(element)
	entry := element.Value.(*CacheEntry)
	delete(c.items, entry.Key)
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	Capacity  int     `json:"capacity"`
	HitRatio  float64 `json:"hit_ratio"`
}

// String returns a string representation of cache stats
func (s CacheStats) String() string {
	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, Evictions=%d, Size=%d/%d, HitRatio=%.2f%%",
		s.Hits, s.Misses, s.Evictions, s.Size, s.Capacity, s.HitRatio*100)
}