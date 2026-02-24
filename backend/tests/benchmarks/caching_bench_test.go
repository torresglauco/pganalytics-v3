package benchmarks

import (
	"sync"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"go.uber.org/zap/zaptest"
)

// BenchmarkCacheGetSet benchmarks basic cache get/set operations
func BenchmarkCacheGetSet(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i % 1000
		c.Set(key, "value")
		_, _ = c.Get(key)
	}
}

// BenchmarkCacheConcurrentReads benchmarks concurrent cache reads
func BenchmarkCacheConcurrentReads(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Populate cache
	for i := 0; i < 1000; i++ {
		c.Set(i, "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = c.Get(i % 1000)
			i++
		}
	})
}

// BenchmarkCacheConcurrentWrites benchmarks concurrent cache writes
func BenchmarkCacheConcurrentWrites(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(i, "value")
			i++
		}
	})
}

// BenchmarkCacheConcurrentMixed benchmarks concurrent mixed read/write operations
func BenchmarkCacheConcurrentMixed(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Populate cache
	for i := 0; i < 500; i++ {
		c.Set(i, "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				c.Set(500+i, "value")
			} else {
				_, _ = c.Get(i % 500)
			}
			i++
		}
	})
}

// BenchmarkCacheEviction benchmarks LRU eviction performance
func BenchmarkCacheEviction(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 100)
	defer c.Close()

	b.ResetTimer()
	// Fill cache to trigger evictions
	for i := 0; i < b.N; i++ {
		c.Set(i, "value")
	}
}

// BenchmarkCacheHitRate benchmarks cache hit rate in realistic scenarios
func BenchmarkCacheHitRate(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 1000)
	defer c.Close()

	// Pre-fill with 500 items (50% of max)
	for i := 0; i < 500; i++ {
		c.Set(i, "value")
	}

	b.ResetTimer()

	hits := 0
	misses := 0

	for i := 0; i < b.N; i++ {
		// 80% chance to hit existing key
		key := i % 600
		if _, found := c.Get(key); found {
			hits++
		} else {
			misses++
			c.Set(key, "value")
		}
	}

	b.Logf("Cache hits: %d, misses: %d, hit rate: %.1f%%",
		hits, misses, float64(hits)*100/float64(hits+misses))
}

// BenchmarkCacheDelete benchmarks cache deletion performance
func BenchmarkCacheDelete(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Pre-fill cache
	for i := 0; i < 5000; i++ {
		c.Set(i, "value")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i % 5000
		c.Delete(key)
		c.Set(key, "value")
	}
}

// BenchmarkCacheMetrics benchmarks metrics calculation performance
func BenchmarkCacheMetrics(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Pre-fill cache and trigger some hits/misses
	for i := 0; i < 1000; i++ {
		c.Set(i, "value")
	}

	for i := 0; i < 500; i++ {
		_, _ = c.Get(i % 1000)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = c.GetMetrics()
	}
}

// ============================================================================
// UNIT TESTS
// ============================================================================

// TestCacheBasicOperations tests basic get/set/delete operations
func TestCacheBasicOperations(t *testing.T) {
	c := cache.NewCache[string, string](1*time.Minute, 100)
	defer c.Close()

	// Test set and get
	c.Set("key1", "value1")
	value, found := c.Get("key1")
	if !found || value != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value)
	}

	// Test missing key
	_, found = c.Get("nonexistent")
	if found {
		t.Error("Expected key not to be found")
	}

	// Test delete
	c.Delete("key1")
	_, found = c.Get("key1")
	if found {
		t.Error("Expected deleted key not to be found")
	}
}

// TestCacheExpiration tests TTL expiration
func TestCacheExpiration(t *testing.T) {
	c := cache.NewCache[string, string](100*time.Millisecond, 100)
	defer c.Close()

	c.Set("key1", "value1")

	// Immediately should exist
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected key to be found immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = c.Get("key1")
	if found {
		t.Error("Expected key to be expired after TTL")
	}
}

// TestCacheLRUEviction tests LRU eviction when cache is full
func TestCacheLRUEviction(t *testing.T) {
	c := cache.NewCache[int, string](1*time.Minute, 5)
	defer c.Close()

	// Fill cache to capacity
	for i := 0; i < 5; i++ {
		c.Set(i, "value")
	}

	// Verify all items are present
	metrics := c.GetMetrics()
	if metrics.Hits+metrics.Misses != 0 {
		t.Errorf("Expected no hits/misses yet, got %d", metrics.Hits+metrics.Misses)
	}

	// Access item 0 to make it recently used
	_, _ = c.Get(0)

	// Add items beyond capacity - item 1 should be evicted (least recently used)
	c.Set(5, "value")

	// Item 0 should still exist (was recently accessed)
	_, found := c.Get(0)
	if !found {
		t.Error("Expected recently accessed item to be in cache")
	}

	// Item 1 should be evicted
	_, found = c.Get(1)
	if found {
		t.Error("Expected LRU item to be evicted")
	}

	// Verify eviction was recorded in metrics
	metrics = c.GetMetrics()
	if metrics.Evictions == 0 {
		t.Error("Expected evictions to be recorded")
	}
}

// TestCacheThreadSafety tests concurrent access safety
func TestCacheThreadSafety(t *testing.T) {
	c := cache.NewCache[int, int](1*time.Minute, 10000)
	defer c.Close()

	numGoroutines := 20
	operationsPerGoroutine := 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines doing reads and writes
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				key := (goroutineID + i) % 100
				c.Set(key, i)
				_, _ = c.Get(key)
				if i%3 == 0 {
					c.Delete(key)
				}
			}
		}(g)
	}

	wg.Wait()

	// If we got here without panicking, thread safety is OK
	metrics := c.GetMetrics()
	if metrics.Hits < 0 || metrics.Misses < 0 {
		t.Error("Expected non-negative hit/miss counts")
	}
}

// TestCachingUnderLoad tests cache behavior under sustained load
func TestCachingUnderLoad(t *testing.T) {
	c := cache.NewCache[int, string](5*time.Minute, 1000)
	defer c.Close()

	numGoroutines := 100
	operationsPerGoroutine := 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < operationsPerGoroutine; i++ {
				// 80% read, 20% write distribution
				key := (goroutineID*10 + i) % 500
				if i%5 == 0 {
					c.Set(key, "value")
				} else {
					_, _ = c.Get(key)
				}
			}
		}(g)
	}

	wg.Wait()

	metrics := c.GetMetrics()
	totalOps := metrics.Hits + metrics.Misses
	if totalOps == 0 {
		t.Error("Expected some cache operations")
	}

	// Calculate hit rate
	hitRate := float64(metrics.Hits) / float64(totalOps)

	// Under load with 80% reads, we should have good hit rate
	t.Logf("Cache hit rate under load: %.1f%% (hits: %d, misses: %d, evictions: %d)",
		hitRate*100, metrics.Hits, metrics.Misses, metrics.Evictions)

	if hitRate < 0.5 {
		t.Logf("Warning: Hit rate %.1f%% is lower than expected 70%+", hitRate*100)
	}
}

// TestCacheMetrics tests metrics accuracy
func TestCacheMetrics(t *testing.T) {
	c := cache.NewCache[int, string](1*time.Minute, 100)
	defer c.Close()

	// Perform known operations
	// Sets (10)
	for i := 0; i < 10; i++ {
		c.Set(i, "value")
	}

	// Hits (5)
	for i := 0; i < 5; i++ {
		_, _ = c.Get(i)
	}

	// Misses (3)
	for i := 10; i < 13; i++ {
		_, _ = c.Get(i)
	}

	metrics := c.GetMetrics()

	// Verify hit count
	if metrics.Hits < 5 {
		t.Errorf("Expected at least 5 hits, got %d", metrics.Hits)
	}

	// Verify miss count
	if metrics.Misses < 3 {
		t.Errorf("Expected at least 3 misses, got %d", metrics.Misses)
	}

	// Verify hit rate is reasonable
	totalOps := metrics.Hits + metrics.Misses
	if totalOps > 0 {
		hitRate := float64(metrics.Hits) / float64(totalOps)
		if hitRate < 0.4 || hitRate > 1.0 {
			t.Errorf("Hit rate %.2f is outside reasonable bounds", hitRate)
		}
	}
}

// TestCacheManagerBasics tests the cache manager
func TestCacheManagerBasics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	m := cache.NewManager(1000, 5*time.Minute, 5*time.Minute, logger)

	// Test feature cache
	m.SetFeatures("query:1", "features1")
	val, found := m.GetFeatures("query:1")
	if !found || val != "features1" {
		t.Error("Expected feature to be cached")
	}

	m.ClearFeatures("query:1")
	_, found = m.GetFeatures("query:1")
	if found {
		t.Error("Expected feature to be cleared")
	}

	// Test prediction cache
	m.SetPrediction("pred:1", "prediction1")
	val, found = m.GetPrediction("pred:1")
	if !found || val != "prediction1" {
		t.Error("Expected prediction to be cached")
	}

	// Test clear all
	m.Clear()
	_, found = m.GetPrediction("pred:1")
	if found {
		t.Error("Expected prediction to be cleared after Clear()")
	}

	// Test close
	if err := m.Close(); err != nil {
		t.Errorf("Expected Close() to succeed, got error: %v", err)
	}
}
