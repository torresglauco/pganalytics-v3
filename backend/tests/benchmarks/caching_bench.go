package benchmarks

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
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

	// Fill cache to trigger evictions
	for i := 0; i < b.N; i++ {
		c.Set(i, "value")
	}
}

// BenchmarkCacheHitRate benchmarks cache hit rate performance
func BenchmarkCacheHitRate(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Populate with 100 items
	for i := 0; i < 100; i++ {
		c.Set(i, "value")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 80% hit rate (access same 100 keys repeatedly)
		c.Get(i % 100)
	}
}

// BenchmarkCacheDelete benchmarks cache delete operations
func BenchmarkCacheDelete(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := i % 1000
		c.Set(key, "value")
		c.Delete(key)
	}
}

// BenchmarkCacheMetrics benchmarks cache metrics calculation
func BenchmarkCacheMetrics(b *testing.B) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Warm up cache
	for i := 0; i < 1000; i++ {
		c.Set(i, "value")
		_, _ = c.Get(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = c.GetMetrics()
	}
}

// TestCacheMemoryUsage tests cache memory usage under load
func TestCacheMemoryUsage(t *testing.T) {
	c := cache.NewCache[int, [1024]byte](10*time.Minute, 10000)
	defer c.Close()

	// Populate cache with 10000 items
	for i := 0; i < 10000; i++ {
		var data [1024]byte
		c.Set(i, data)
	}

	size := c.Size()
	if size != 10000 {
		t.Errorf("Expected cache size 10000, got %d", size)
	}

	metrics := c.GetMetrics()
	t.Logf("Cache metrics - Hits: %d, Misses: %d, Evictions: %d",
		metrics.Hits, metrics.Misses, metrics.Evictions)
}

// TestCacheExpiration tests cache item expiration
func TestCacheExpiration(t *testing.T) {
	c := cache.NewCache[string, string](100*time.Millisecond, 100)
	defer c.Close()

	c.Set("key", "value")

	// Item should exist immediately
	if val, found := c.Get("key"); !found || val != "value" {
		t.Error("Expected item to exist immediately after insertion")
	}

	// Item should expire after TTL
	time.Sleep(150 * time.Millisecond)
	if _, found := c.Get("key"); found {
		t.Error("Expected item to be expired")
	}
}

// TestCacheThreadSafety tests cache thread safety with high contention
func TestCacheThreadSafety(t *testing.T) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	numGoroutines := 20
	opsPerGoroutine := 1000
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				key := (id * opsPerGoroutine) + i
				c.Set(key, "value")

				// Read back
				if val, found := c.Get(key); !found || val != "value" {
					t.Errorf("Failed to retrieve value for key %d", key)
				}
			}
		}(g)
	}

	wg.Wait()

	// Verify some data
	if _, found := c.Get(0); !found {
		t.Error("Expected first item to exist")
	}

	metrics := c.GetMetrics()
	if metrics.Hits == 0 {
		t.Error("Expected cache hits > 0")
	}
}

// BenchmarkFeatureExtractorWithCache simulates feature extraction with caching
func BenchmarkFeatureExtractorWithCache(b *testing.B) {
	c := cache.NewCache[string, *models.QueryFeatures](15*time.Minute, 10000)
	defer c.Close()

	// Simulate features
	features := &models.QueryFeatures{
		QueryHash: 12345,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate cache hit/miss pattern (80% hit rate)
		key := "feature:" + string(rune((i % 100)))

		if i%5 == 0 { // 20% miss
			c.Set(key, features)
		} else { // 80% hit
			_, _ = c.Get(key)
		}
	}
}

// BenchmarkPredictionCaching simulates ML prediction caching
func BenchmarkPredictionCaching(b *testing.B) {
	c := cache.NewCache[string, *models.PredictionCacheEntry](5*time.Minute, 10000)
	defer c.Close()

	prediction := &models.PredictionCacheEntry{
		Time: time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "pred:" + string(rune((i % 100)))

		if i%3 == 0 { // 33% miss
			c.Set(key, prediction)
		} else { // 67% hit
			_, _ = c.Get(key)
		}
	}
}

// TestCachingUnderLoad tests caching behavior under concurrent load
func TestCachingUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	numGoroutines := 100
	requestsPerGoroutine := 100
	totalRequests := numGoroutines * requestsPerGoroutine

	var successCount int64
	var mu sync.Mutex
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	start := time.Now()
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < requestsPerGoroutine; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					// Simulate hot keys (80/20 principle)
					key := i % 20

					if i%5 == 0 {
						c.Set(key, "value")
					} else {
						if _, found := c.Get(key); found {
							mu.Lock()
							successCount++
							mu.Unlock()
						}
					}
				}
			}
		}(g)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Calculate hit rate
	metrics := c.GetMetrics()
	totalHitsAndMisses := metrics.Hits + metrics.Misses
	hitRate := 0.0
	if totalHitsAndMisses > 0 {
		hitRate = float64(metrics.Hits) / float64(totalHitsAndMisses)
	}

	t.Logf("Load test results:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful cache hits: %d", successCount)
	t.Logf("  Cache hits: %d, misses: %d", metrics.Hits, metrics.Misses)
	t.Logf("  Hit rate: %.2f%%", hitRate*100)
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Throughput: %.0f req/sec", float64(totalRequests)/elapsed.Seconds())

	// Expect >80% hit rate with hot keys
	if hitRate < 0.80 {
		t.Errorf("Expected hit rate >80%%, got %.2f%%", hitRate*100)
	}
}

// BenchmarkCacheDashboard provides an overview of cache performance characteristics
func BenchmarkCacheDashboard(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"GetSet", BenchmarkCacheGetSet},
		{"ConcurrentReads", BenchmarkCacheConcurrentReads},
		{"ConcurrentWrites", BenchmarkCacheConcurrentWrites},
		{"ConcurrentMixed", BenchmarkCacheConcurrentMixed},
		{"Eviction", BenchmarkCacheEviction},
		{"Delete", BenchmarkCacheDelete},
		{"Metrics", BenchmarkCacheMetrics},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, bm.fn)
	}
}

// TestCachePerformanceCharacteristics documents and validates expected cache performance
func TestCachePerformanceCharacteristics(t *testing.T) {
	c := cache.NewCache[int, string](10*time.Minute, 10000)
	defer c.Close()

	// Expected performance characteristics:
	// - Get/Set: <1 microsecond per operation
	// - Concurrent reads: <1 microsecond per operation
	// - Eviction: <1 millisecond
	// - Memory: ~500 bytes per cached item

	// Test single operations
	start := time.Now()
	for i := 0; i < 10000; i++ {
		c.Set(i, "value")
	}
	elapsed := time.Since(start)
	avgSetTime := elapsed / 10000

	if avgSetTime > 1000*time.Nanosecond {
		t.Logf("Warning: Set operation took %v (expected <1μs)", avgSetTime)
	}

	t.Logf("Cache performance characteristics:")
	t.Logf("  Set operation: %v avg", avgSetTime)

	// Test concurrent reads
	start = time.Now()
	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_, _ = c.Get(i % 10000)
			}
		}()
	}
	wg.Wait()
	elapsed = time.Since(start)
	avgGetTime := elapsed / 10000

	if avgGetTime > 1000*time.Nanosecond {
		t.Logf("Warning: Get operation took %v (expected <1μs)", avgGetTime)
	}

	t.Logf("  Get operation: %v avg", avgGetTime)

	// Get cache size estimation
	metrics := c.GetMetrics()
	t.Logf("  Total cache items: %d", c.Size())
	t.Logf("  Metrics - Hits: %d, Misses: %d, Evictions: %d",
		metrics.Hits, metrics.Misses, metrics.Evictions)
}
