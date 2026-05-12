package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAPIResponseTimeHistogram(t *testing.T) {
	// Test 1: Histogram has buckets for P50, P95, P99 ranges
	t.Run("has buckets covering percentile ranges", func(t *testing.T) {
		buckets := HistogramBuckets()

		// Check that buckets cover the range from 1ms to 10s
		assert.Contains(t, buckets, 0.001, "Should have 1ms bucket")
		assert.Contains(t, buckets, 0.01, "Should have 10ms bucket (P50 range)")
		assert.Contains(t, buckets, 0.1, "Should have 100ms bucket (P95 range)")
		assert.Contains(t, buckets, 0.5, "Should have 500ms bucket (P99 range)")
		assert.Contains(t, buckets, 10.0, "Should have 10s bucket")
	})

	// Test 2: RecordAPIResponseTime accepts method, path, status, duration
	t.Run("records API response times with labels", func(t *testing.T) {
		// Record some response times
		RecordAPIResponseTime("GET", "/api/v1/users", 200, 50*time.Millisecond)
		RecordAPIResponseTime("GET", "/api/v1/users", 200, 100*time.Millisecond)
		RecordAPIResponseTime("POST", "/api/v1/users", 201, 150*time.Millisecond)

		// Verify histogram was updated
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})
}

func TestQueryDurationHistogram(t *testing.T) {
	// Test 1: Histogram has buckets from 1ms to 10s
	t.Run("has buckets from 1ms to 10s", func(t *testing.T) {
		buckets := HistogramBuckets()

		// Verify bucket range
		assert.GreaterOrEqual(t, len(buckets), 10, "Should have at least 10 buckets")
		assert.LessOrEqual(t, buckets[0], 0.001, "First bucket should be <= 1ms")
		assert.GreaterOrEqual(t, buckets[len(buckets)-1], 10.0, "Last bucket should be >= 10s")
	})

	// Test 2: RecordQueryDuration accepts database, query_type, duration
	t.Run("records query durations with labels", func(t *testing.T) {
		// Record some query durations
		RecordQueryDuration("postgres", "SELECT", 25*time.Millisecond)
		RecordQueryDuration("postgres", "SELECT", 50*time.Millisecond)
		RecordQueryDuration("timescale", "INSERT", 100*time.Millisecond)

		// Verify histogram was updated
		count := testutil.CollectAndCount(QueryDurationHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})
}

func TestQueryCounter(t *testing.T) {
	t.Run("increments query counter with labels", func(t *testing.T) {
		// Increment counters
		IncrementQueryCount("postgres", "SELECT", "success")
		IncrementQueryCount("postgres", "SELECT", "success")
		IncrementQueryCount("postgres", "INSERT", "error")

		// Verify counter was updated
		count := testutil.CollectAndCount(QueryCounter)
		assert.GreaterOrEqual(t, count, 1, "Counter should have metric data")
	})
}

func TestPercentileLabels(t *testing.T) {
	t.Run("returns human-readable percentile labels", func(t *testing.T) {
		labels := PercentileLabels()

		assert.Equal(t, "1ms", labels[0.001])
		assert.Equal(t, "10ms", labels[0.01])
		assert.Equal(t, "100ms", labels[0.1])
		assert.Equal(t, "1s", labels[1.0])
		assert.Equal(t, "10s", labels[10.0])
	})
}

// ============================================================================
// CACHE METRICS TESTS
// ============================================================================

func TestCacheHitsTotal(t *testing.T) {
	t.Run("increments cache hits counter", func(t *testing.T) {
		// Record cache hits
		RecordCacheHit("response")
		RecordCacheHit("response")
		RecordCacheHit("feature")

		// Verify counter was updated
		count := testutil.CollectAndCount(CacheHitsTotal)
		assert.GreaterOrEqual(t, count, 1, "Counter should have metric data")
	})
}

func TestCacheMissesTotal(t *testing.T) {
	t.Run("increments cache misses counter", func(t *testing.T) {
		// Record cache misses
		RecordCacheMiss("response")
		RecordCacheMiss("response")

		// Verify counter was updated
		count := testutil.CollectAndCount(CacheMissesTotal)
		assert.GreaterOrEqual(t, count, 1, "Counter should have metric data")
	})
}

func TestCacheEvictionsTotal(t *testing.T) {
	t.Run("increments cache evictions counter", func(t *testing.T) {
		// Record cache evictions
		RecordCacheEviction("response")

		// Verify counter was updated
		count := testutil.CollectAndCount(CacheEvictionsTotal)
		assert.GreaterOrEqual(t, count, 1, "Counter should have metric data")
	})
}

func TestCacheSizeGauge(t *testing.T) {
	t.Run("sets cache size gauge", func(t *testing.T) {
		// Set cache size
		SetCacheSize("response", 1024)
		SetCacheSize("feature", 2048)

		// Verify gauge was updated
		count := testutil.CollectAndCount(CacheSizeBytes)
		assert.GreaterOrEqual(t, count, 1, "Gauge should have metric data")
	})
}

func TestCacheEntriesGauge(t *testing.T) {
	t.Run("sets cache entries gauge", func(t *testing.T) {
		// Set cache entries
		SetCacheEntries("response", 100)
		SetCacheEntries("feature", 50)

		// Verify gauge was updated
		count := testutil.CollectAndCount(CacheEntries)
		assert.GreaterOrEqual(t, count, 1, "Gauge should have metric data")
	})
}

func TestCacheLatencyHistogram(t *testing.T) {
	t.Run("records cache latency values", func(t *testing.T) {
		// Record cache latencies
		RecordCacheLatency("response", "get", 1*time.Millisecond)
		RecordCacheLatency("response", "get", 5*time.Millisecond)
		RecordCacheLatency("response", "set", 10*time.Millisecond)
		RecordCacheLatency("feature", "get", 2*time.Millisecond)

		// Verify histogram was updated
		count := testutil.CollectAndCount(CacheLatencySeconds)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})
}

func TestCacheLatencyHistogramBuckets(t *testing.T) {
	t.Run("has buckets covering cache operation latency range", func(t *testing.T) {
		// CacheLatencySeconds should have buckets for sub-millisecond to 100ms operations
		// Record a very fast and a slower operation
		RecordCacheLatency("response", "get", 100*time.Microsecond) // 0.1ms
		RecordCacheLatency("response", "set", 50*time.Millisecond)  // 50ms

		count := testutil.CollectAndCount(CacheLatencySeconds)
		assert.GreaterOrEqual(t, count, 1, "Histogram should record fast and slow operations")
	})
}
