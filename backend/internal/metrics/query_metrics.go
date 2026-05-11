package metrics

import (
	"sort"
	"sync"
	"time"
)

// QueryMetrics tracks query performance metrics with a sliding window
type QueryMetrics struct {
	mu            sync.RWMutex
	durations     []time.Duration
	maxSamples    int
	totalQueries  int64
	totalDuration time.Duration
	minDuration   time.Duration
	maxDuration   time.Duration
}

// NewQueryMetrics creates a new query metrics tracker
func NewQueryMetrics(maxSamples int) *QueryMetrics {
	if maxSamples <= 0 {
		maxSamples = 10000 // Default: keep last 10k queries
	}
	return &QueryMetrics{
		durations:   make([]time.Duration, 0, maxSamples),
		maxSamples:  maxSamples,
		minDuration: time.Duration(1<<63 - 1), // Max duration
	}
}

// RecordQuery records a query duration
func (qm *QueryMetrics) RecordQuery(duration time.Duration) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Update totals
	qm.totalQueries++
	qm.totalDuration += duration

	// Update min/max
	if duration < qm.minDuration {
		qm.minDuration = duration
	}
	if duration > qm.maxDuration {
		qm.maxDuration = duration
	}

	// Add to sliding window
	qm.durations = append(qm.durations, duration)
	if len(qm.durations) > qm.maxSamples {
		// Remove oldest entry (FIFO)
		qm.durations = qm.durations[1:]
	}
}

// QueryStats contains query performance statistics
type QueryStats struct {
	Count       int64         `json:"count"`
	MinDuration time.Duration `json:"min_duration"`
	MaxDuration time.Duration `json:"max_duration"`
	AvgDuration time.Duration `json:"avg_duration"`
	P50         time.Duration `json:"p50"`
	P95         time.Duration `json:"p95"`
	P99         time.Duration `json:"p99"`
}

// GetStats returns comprehensive query statistics
func (qm *QueryMetrics) GetStats() QueryStats {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	stats := QueryStats{
		Count:       qm.totalQueries,
		MinDuration: qm.minDuration,
		MaxDuration: qm.maxDuration,
	}

	if qm.totalQueries > 0 {
		stats.AvgDuration = time.Duration(int64(qm.totalDuration) / qm.totalQueries)
	}

	if len(qm.durations) > 0 {
		stats.P50 = qm.percentile(0.50)
		stats.P95 = qm.percentile(0.95)
		stats.P99 = qm.percentile(0.99)
	}

	return stats
}

// percentile calculates the given percentile (0.0 to 1.0)
// Assumes qm.mu is already locked
func (qm *QueryMetrics) percentile(p float64) time.Duration {
	if len(qm.durations) == 0 {
		return 0
	}

	// Make a copy for sorting
	sorted := make([]time.Duration, len(qm.durations))
	copy(sorted, qm.durations)

	// Sort the durations
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate index
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

// Global query metrics instance
var globalQueryMetrics = NewQueryMetrics(10000)

// RecordGlobalQuery records to the global metrics instance
func RecordGlobalQuery(duration time.Duration) {
	globalQueryMetrics.RecordQuery(duration)
}

// GetGlobalQueryStats returns stats from the global instance
func GetGlobalQueryStats() QueryStats {
	return globalQueryMetrics.GetStats()
}
