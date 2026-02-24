package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheMetrics tracks cache performance statistics
type CacheMetrics struct {
	Hits      int64
	Misses    int64
	Evictions int64
}

// cacheItem represents a cached value with expiration
type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache is a generic thread-safe in-memory cache with TTL support and LRU eviction
type Cache[K comparable, V any] struct {
	items    map[K]*cacheItem[V]
	mu       sync.RWMutex
	ttl      time.Duration
	maxSize  int
	metrics  *CacheMetrics
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewCache creates a new cache with the specified TTL and max size
func NewCache[K comparable, V any](ttl time.Duration, maxSize int) *Cache[K, V] {
	c := &Cache[K, V]{
		items:   make(map[K]*cacheItem[V]),
		ttl:     ttl,
		maxSize: maxSize,
		metrics: &CacheMetrics{},
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	c.wg.Add(1)
	go c.cleanupExpired()

	return c
}

// Get retrieves a value from the cache
// Returns (value, true) if found and not expired, (zero, false) otherwise
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		atomic.AddInt64(&c.metrics.Misses, 1)
		var zero V
		return zero, false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		atomic.AddInt64(&c.metrics.Misses, 1)
		var zero V
		return zero, false
	}

	atomic.AddInt64(&c.metrics.Hits, 1)
	return item.value, true
}

// Set stores a value in the cache with TTL
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	c.items[key] = &cacheItem[V]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a key from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*cacheItem[V])
}

// GetMetrics returns current cache metrics
func (c *Cache[K, V]) GetMetrics() CacheMetrics {
	return CacheMetrics{
		Hits:      atomic.LoadInt64(&c.metrics.Hits),
		Misses:    atomic.LoadInt64(&c.metrics.Misses),
		Evictions: atomic.LoadInt64(&c.metrics.Evictions),
	}
}

// Size returns the current number of items in the cache
func (c *Cache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// evictLRU removes the least recently used item (oldest expiration time)
// Must be called while holding the lock
func (c *Cache[K, V]) evictLRU() {
	if len(c.items) == 0 {
		return
	}

	// Find the item with earliest expiration time
	var oldestKey K
	var oldestTime time.Time
	first := true

	for k, item := range c.items {
		if first || item.expiresAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = item.expiresAt
			first = false
		}
	}

	delete(c.items, oldestKey)
	atomic.AddInt64(&c.metrics.Evictions, 1)
}

// cleanupExpired periodically removes expired items from the cache
func (c *Cache[K, V]) cleanupExpired() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.ttl / 2) // Cleanup every half TTL
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.removeExpiredItems()
		}
	}
}

// removeExpiredItems removes all expired items from the cache
func (c *Cache[K, V]) removeExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, k)
		}
	}
}

// Close stops the cache cleanup goroutine and clears all items
func (c *Cache[K, V]) Close() error {
	close(c.stopCh)
	c.wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[K]*cacheItem[V])
	return nil
}
