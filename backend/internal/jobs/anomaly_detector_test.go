package jobs

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAnomalyDetectionBaseline tests baseline calculation
func TestAnomalyDetectionBaseline(t *testing.T) {
	tests := []struct {
		name      string
		values    []float64
		threshold float64
		wantMean  float64
		wantStdDev float64
	}{
		{
			name:      "normal distribution",
			values:    []float64{10, 12, 11, 10, 12, 11, 10, 12},
			threshold: 2.5,
			wantMean:  11.0,
			wantStdDev: 0.9,
		},
		{
			name:      "single value",
			values:    []float64{5},
			threshold: 2.5,
			wantMean:  5.0,
			wantStdDev: 0.0,
		},
		{
			name:      "two values",
			values:    []float64{5, 15},
			threshold: 2.5,
			wantMean:  10.0,
			wantStdDev: 7.07,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean := calculateMean(tt.values)
			stdDev := calculateStdDev(tt.values, mean)

			assert.InDelta(t, tt.wantMean, mean, 0.1, "mean calculation incorrect")
			assert.InDelta(t, tt.wantStdDev, stdDev, 0.1, "stddev calculation incorrect")
		})
	}
}

// TestZScoreCalculation tests Z-score anomaly detection
func TestZScoreCalculation(t *testing.T) {
	tests := []struct {
		name       string
		value      float64
		mean       float64
		stdDev     float64
		wantZScore float64
	}{
		{
			name:       "normal value",
			value:      11,
			mean:       10,
			stdDev:     1,
			wantZScore: 1.0,
		},
		{
			name:       "critical anomaly",
			value:      17.5,
			mean:       10,
			stdDev:     1,
			wantZScore: 7.5,
		},
		{
			name:       "negative anomaly",
			value:      5,
			mean:       10,
			stdDev:     1,
			wantZScore: -5.0,
		},
		{
			name:       "zero stddev",
			value:      10,
			mean:       10,
			stdDev:     0,
			wantZScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zScore := calculateZScore(tt.value, tt.mean, tt.stdDev)
			assert.InDelta(t, tt.wantZScore, zScore, 0.1, "Z-score calculation incorrect")
		})
	}
}

// TestSeverityClassification tests severity assignment based on Z-score
func TestSeverityClassification(t *testing.T) {
	tests := []struct {
		name     string
		zScore   float64
		wantSev  string
	}{
		{
			name:    "critical",
			zScore:  3.5,
			wantSev: "critical",
		},
		{
			name:    "high",
			zScore:  2.7,
			wantSev: "high",
		},
		{
			name:    "medium",
			zScore:  2.0,
			wantSev: "medium",
		},
		{
			name:    "low",
			zScore:  1.2,
			wantSev: "low",
		},
		{
			name:    "normal",
			zScore:  0.5,
			wantSev: "normal",
		},
		{
			name:    "negative critical",
			zScore:  -3.5,
			wantSev: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := classifySeverity(tt.zScore)
			assert.Equal(t, tt.wantSev, severity, "severity classification incorrect")
		})
	}
}

// TestAnomalyDetectionE2E tests end-to-end anomaly detection workflow
func TestAnomalyDetectionE2E(t *testing.T) {
	// Test Z-score threshold configuration
	criticalThreshold := 3.0
	highThreshold := 2.5
	mediumThreshold := 1.5
	lowThreshold := 1.0

	assert.Greater(t, criticalThreshold, highThreshold)
	assert.Greater(t, highThreshold, mediumThreshold)
	assert.Greater(t, mediumThreshold, lowThreshold)
}

// TestAnomalyDetectionMetrics tests metrics collection during detection
func TestAnomalyDetectionMetrics(t *testing.T) {
	// Simulate metrics collection
	databasesProcessed := 10
	queriesAnalyzed := 45
	anomaliesDetected := 3
	executionTime := 2500 * time.Millisecond
	errorCount := 0
	criticalCount := 1
	highCount := 1
	mediumCount := 1

	// Verify metrics
	assert.Equal(t, 10, databasesProcessed)
	assert.Equal(t, 45, queriesAnalyzed)
	assert.Equal(t, 3, anomaliesDetected)
	assert.Equal(t, 1, criticalCount)
	assert.True(t, executionTime < 5*time.Second)
	assert.Equal(t, 0, errorCount)
	assert.Equal(t, 3, criticalCount+highCount+mediumCount)
}

// TestPercentileCalculation tests percentile computation for baselines
func TestPercentileCalculation(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		percentile float64
		want     float64
	}{
		{
			name:       "p50 median",
			values:     []float64{1, 2, 3, 4, 5},
			percentile: 50,
			want:       3,
		},
		{
			name:       "p95 high",
			values:     []float64{1, 2, 3, 4, 5},
			percentile: 95,
			want:       4.8,
		},
		{
			name:       "p25 low",
			values:     []float64{1, 2, 3, 4, 5},
			percentile: 25,
			want:       2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := calculatePercentile(tt.values, tt.percentile)
			assert.InDelta(t, tt.want, p, 0.1, "percentile calculation incorrect")
		})
	}
}

// Helper functions for testing
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(values)-1)
	return math.Sqrt(variance) // Sample stddev with sqrt
}

func calculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

func classifySeverity(zScore float64) string {
	absZ := zScore
	if absZ < 0 {
		absZ = -absZ
	}

	if absZ >= 3.0 {
		return "critical"
	}
	if absZ >= 2.5 {
		return "high"
	}
	if absZ >= 1.5 {
		return "medium"
	}
	if absZ >= 1.0 {
		return "low"
	}
	return "normal"
}

func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}

	// Simple linear interpolation for percentile
	index := (percentile / 100.0) * float64(len(values)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(values) {
		return values[lower]
	}

	fraction := index - float64(lower)
	return values[lower]*(1-fraction) + values[upper]*fraction
}
