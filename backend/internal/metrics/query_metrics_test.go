package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryMetrics(t *testing.T) {
	// Test 1: NewQueryMetrics creates metrics tracker with default window
	t.Run("creates tracker with default window", func(t *testing.T) {
		qm := NewQueryMetrics(0)
		assert.NotNil(t, qm)
		assert.Equal(t, 10000, qm.maxSamples, "Default should be 10000 samples")
	})

	t.Run("creates tracker with custom window", func(t *testing.T) {
		qm := NewQueryMetrics(5000)
		assert.NotNil(t, qm)
		assert.Equal(t, 5000, qm.maxSamples)
	})
}

func TestQueryMetrics_RecordQuery(t *testing.T) {
	// Test 2: RecordQuery adds query duration to sliding window
	t.Run("records queries correctly", func(t *testing.T) {
		qm := NewQueryMetrics(100)

		qm.RecordQuery(10 * time.Millisecond)
		qm.RecordQuery(20 * time.Millisecond)
		qm.RecordQuery(30 * time.Millisecond)

		stats := qm.GetStats()
		assert.Equal(t, int64(3), stats.Count)
		assert.Equal(t, 10*time.Millisecond, stats.MinDuration)
		assert.Equal(t, 30*time.Millisecond, stats.MaxDuration)
	})
}

func TestQueryMetrics_GetStats(t *testing.T) {
	// Test 3: GetPercentiles returns P50, P95, P99 from recorded data
	t.Run("calculates percentiles correctly", func(t *testing.T) {
		qm := NewQueryMetrics(10000)

		// Add 100 samples from 1ms to 100ms
		for i := 1; i <= 100; i++ {
			qm.RecordQuery(time.Duration(i) * time.Millisecond)
		}

		stats := qm.GetStats()

		// P50 should be around 50ms
		assert.GreaterOrEqual(t, stats.P50, 45*time.Millisecond, "P50 should be around 50ms")
		assert.LessOrEqual(t, stats.P50, 55*time.Millisecond, "P50 should be around 50ms")

		// P95 should be around 95ms
		assert.GreaterOrEqual(t, stats.P95, 90*time.Millisecond, "P95 should be around 95ms")
		assert.LessOrEqual(t, stats.P95, 98*time.Millisecond, "P95 should be around 95ms")

		// P99 should be around 99ms
		assert.GreaterOrEqual(t, stats.P99, 97*time.Millisecond, "P99 should be around 99ms")
		assert.LessOrEqual(t, stats.P99, 100*time.Millisecond, "P99 should be around 99ms")
	})

	// Test 4: GetStats returns comprehensive statistics
	t.Run("returns comprehensive statistics", func(t *testing.T) {
		qm := NewQueryMetrics(10000)

		// Add some samples
		for i := 1; i <= 10; i++ {
			qm.RecordQuery(time.Duration(i*10) * time.Millisecond)
		}

		stats := qm.GetStats()

		assert.Equal(t, int64(10), stats.Count)
		assert.Equal(t, 10*time.Millisecond, stats.MinDuration)
		assert.Equal(t, 100*time.Millisecond, stats.MaxDuration)
		assert.Equal(t, 55*time.Millisecond, stats.AvgDuration) // (10+20+...+100)/10 = 55
		assert.Greater(t, stats.P50, time.Duration(0))
		assert.Greater(t, stats.P95, time.Duration(0))
		assert.Greater(t, stats.P99, time.Duration(0))
	})
}

func TestQueryMetrics_SlidingWindow(t *testing.T) {
	t.Run("maintains sliding window of samples", func(t *testing.T) {
		qm := NewQueryMetrics(5) // Small window for testing

		// Add 10 samples, only last 5 should be kept
		for i := 1; i <= 10; i++ {
			qm.RecordQuery(time.Duration(i) * time.Millisecond)
		}

		stats := qm.GetStats()

		// Total count should be 10
		assert.Equal(t, int64(10), stats.Count)

		// But percentiles should be based on last 5 samples (6-10ms)
		// P50 should be around 8ms
		assert.GreaterOrEqual(t, stats.P50, 7*time.Millisecond)
		assert.LessOrEqual(t, stats.P50, 9*time.Millisecond)
	})
}

func TestQueryMetrics_GlobalFunctions(t *testing.T) {
	t.Run("global functions work correctly", func(t *testing.T) {
		// Record to global instance
		RecordGlobalQuery(50 * time.Millisecond)
		RecordGlobalQuery(100 * time.Millisecond)

		stats := GetGlobalQueryStats()
		require.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.Count, int64(2))
	})
}

func TestQueryMetrics_ThreadSafety(t *testing.T) {
	t.Run("concurrent writes are safe", func(t *testing.T) {
		qm := NewQueryMetrics(10000)

		// Launch multiple goroutines writing concurrently
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					qm.RecordQuery(time.Duration(j) * time.Millisecond)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		stats := qm.GetStats()
		assert.Equal(t, int64(1000), stats.Count)
	})
}