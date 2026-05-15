package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// TestCalculateHostHealthScore_Score100WhenAllMetricsOptimal tests perfect score
func TestCalculateHostHealthScore_Score100WhenAllMetricsOptimal(t *testing.T) {
	t.Parallel()

	collectorID := uuid.New()
	metrics := &models.HostMetrics{
		CollectorID:       collectorID,
		CpuUser:           0,
		CpuSystem:         0,
		CpuIowait:         0,
		CpuIdle:           100,
		MemoryUsedPercent: 0,
		DiskUsedPercent:   0,
		CpuLoad1m:         0.5, // Load < 1.0 -> score = 100
	}

	weights := models.DefaultHealthScoreWeights
	score := CalculateHostHealthScore(metrics, weights)

	assert.Equal(t, 100, score, "Score should be 100 when all metrics are optimal")
}

// TestCalculateHostHealthScore_Score0WhenAllMetricsCritical tests worst case score
func TestCalculateHostHealthScore_Score0WhenAllMetricsCritical(t *testing.T) {
	t.Parallel()

	collectorID := uuid.New()
	metrics := &models.HostMetrics{
		CollectorID:       collectorID,
		CpuUser:           50,
		CpuSystem:         30,
		CpuIowait:         20,  // 100% CPU usage -> CPU score = 0
		MemoryUsedPercent: 100, // 100% memory -> Memory score = 0
		DiskUsedPercent:   100, // 100% disk -> Disk score = 0
		CpuLoad1m:         3.0, // Load > 2.0 -> Load score = 0
	}

	weights := models.DefaultHealthScoreWeights
	score := CalculateHostHealthScore(metrics, weights)

	assert.Equal(t, 0, score, "Score should be 0 when all metrics are at critical thresholds")
}

// TestCalculateHostHealthScore_WeightedCalculation tests the weighted formula
func TestCalculateHostHealthScore_WeightedCalculation(t *testing.T) {
	t.Parallel()

	// Test with known values
	// CPU at 70% idle (score 70), Memory at 80% free (score 80), Disk at 90% free (score 90), Load at 1.5 (score 50)
	// Expected: 70*0.30 + 80*0.25 + 90*0.25 + 50*0.20 = 21 + 20 + 22.5 + 10 = 73.5 -> 74
	collectorID := uuid.New()
	metrics := &models.HostMetrics{
		CollectorID:       collectorID,
		CpuUser:           20,
		CpuSystem:         5,
		CpuIowait:         5,   // Total CPU used = 30, so CPU score = 70
		MemoryUsedPercent: 20,  // Memory score = 80
		DiskUsedPercent:   10,  // Disk score = 90
		CpuLoad1m:         1.5, // Load > 1.0 and <= 2.0 -> Load score = 50
	}

	weights := models.DefaultHealthScoreWeights
	score := CalculateHostHealthScore(metrics, weights)

	// Expected: (70 * 0.30) + (80 * 0.25) + (90 * 0.25) + (50 * 0.20) = 21 + 20 + 22.5 + 10 = 73.5 -> 74
	assert.Equal(t, 74, score, "Score should match weighted calculation")
}

// TestCalculateHostHealthScore_NilMetricsReturns0 tests nil handling
func TestCalculateHostHealthScore_NilMetricsReturns0(t *testing.T) {
	t.Parallel()

	weights := models.DefaultHealthScoreWeights
	score := CalculateHostHealthScore(nil, weights)

	assert.Equal(t, 0, score, "Score should be 0 when metrics is nil")
}

// TestGetHealthStatus_HealthyWhenScoreGE80 tests healthy status boundary
func TestGetHealthStatus_HealthyWhenScoreGE80(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		score int
	}{
		{"score 100", 100},
		{"score 90", 90},
		{"score 80", 80},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status := GetHealthStatus(tc.score)
			assert.Equal(t, "healthy", status, "Status should be healthy for score >= 80")
		})
	}
}

// TestGetHealthStatus_DegradedWhenScoreGE60 tests degraded status boundary
func TestGetHealthStatus_DegradedWhenScoreGE60(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		score int
	}{
		{"score 79", 79},
		{"score 70", 70},
		{"score 60", 60},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status := GetHealthStatus(tc.score)
			assert.Equal(t, "degraded", status, "Status should be degraded for 60 <= score < 80")
		})
	}
}

// TestGetHealthStatus_WarningWhenScoreGE40 tests warning status boundary
func TestGetHealthStatus_WarningWhenScoreGE40(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		score int
	}{
		{"score 59", 59},
		{"score 50", 50},
		{"score 40", 40},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status := GetHealthStatus(tc.score)
			assert.Equal(t, "warning", status, "Status should be warning for 40 <= score < 60")
		})
	}
}

// TestGetHealthStatus_CriticalWhenScoreLT40 tests critical status boundary
func TestGetHealthStatus_CriticalWhenScoreLT40(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		score int
	}{
		{"score 39", 39},
		{"score 20", 20},
		{"score 0", 0},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status := GetHealthStatus(tc.score)
			assert.Equal(t, "critical", status, "Status should be critical for score < 40")
		})
	}
}

// TestCalculateHealthScoreWithDetails_ComponentScoresStored tests component scores
func TestCalculateHealthScoreWithDetails_ComponentScoresStored(t *testing.T) {
	t.Parallel()

	collectorID := uuid.New()
	metrics := &models.HostMetrics{
		CollectorID:       collectorID,
		CpuUser:           10,
		CpuSystem:         5,
		CpuIowait:         5,   // CPU score = 100 - 20 = 80
		MemoryUsedPercent: 30,  // Memory score = 100 - 30 = 70
		DiskUsedPercent:   20,  // Disk score = 100 - 20 = 80
		CpuLoad1m:         0.5, // Load score = 100
	}

	weights := models.DefaultHealthScoreWeights
	result := CalculateHealthScoreWithDetails(metrics, weights)

	require.NotNil(t, result)
	assert.Equal(t, 80.0, result.CpuScore)
	assert.Equal(t, 70.0, result.MemoryScore)
	assert.Equal(t, 80.0, result.DiskScore)
	assert.Equal(t, 100.0, result.LoadScore)
	assert.Equal(t, collectorID, result.CollectorID)
}

// TestCalculateHealthScoreWithDetails_CalculationDetailsJSONB tests details structure
func TestCalculateHealthScoreWithDetails_CalculationDetailsJSONB(t *testing.T) {
	t.Parallel()

	collectorID := uuid.New()
	metrics := &models.HostMetrics{
		CollectorID:       collectorID,
		CpuUser:           15,
		CpuSystem:         10,
		CpuIowait:         5,
		CpuIdle:           70,
		CpuLoad1m:         1.5,
		MemoryUsedPercent: 40,
		DiskUsedPercent:   30,
	}

	weights := models.DefaultHealthScoreWeights
	result := CalculateHealthScoreWithDetails(metrics, weights)

	require.NotNil(t, result)
	require.NotNil(t, result.CalculationDetails)

	// Verify calculation details contain expected fields
	details := result.CalculationDetails
	assert.Contains(t, details, "cpu_user")
	assert.Contains(t, details, "cpu_system")
	assert.Contains(t, details, "cpu_iowait")
	assert.Contains(t, details, "cpu_idle")
	assert.Contains(t, details, "cpu_load_1m")
	assert.Contains(t, details, "memory_used_pct")
	assert.Contains(t, details, "disk_used_pct")
	assert.Contains(t, details, "weights")

	// Verify weights are included
	weightsMap, ok := details["weights"].(map[string]float64)
	require.True(t, ok, "weights should be a map[string]float64")
	assert.Equal(t, 0.30, weightsMap["cpu"])
	assert.Equal(t, 0.25, weightsMap["memory"])
	assert.Equal(t, 0.25, weightsMap["disk"])
	assert.Equal(t, 0.20, weightsMap["load_average"])
}

// TestCalculateHealthScoreWithDetails_NilMetricsReturnsNil tests nil handling
func TestCalculateHealthScoreWithDetails_NilMetricsReturnsNil(t *testing.T) {
	t.Parallel()

	weights := models.DefaultHealthScoreWeights
	result := CalculateHealthScoreWithDetails(nil, weights)

	assert.Nil(t, result, "Result should be nil when metrics is nil")
}

// TestCalculateHostHealthScore_LoadScoreThresholds tests load score calculation
func TestCalculateHostHealthScore_LoadScoreThresholds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		load          float64
		expectedScore int
	}{
		{"load 0.5 - light load", 0.5, 100},
		{"load 1.0 - boundary", 1.0, 100},
		{"load 1.5 - moderate load", 1.5, 50},
		{"load 2.0 - boundary", 2.0, 50},
		{"load 2.5 - heavy load", 2.5, 0},
		{"load 3.0 - heavy load", 3.0, 0},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// For load-only testing, set other metrics to optimal
			metrics := &models.HostMetrics{
				CollectorID:       uuid.New(),
				CpuUser:           0,
				CpuSystem:         0,
				CpuIowait:         0,
				MemoryUsedPercent: 0,
				DiskUsedPercent:   0,
				CpuLoad1m:         tc.load,
			}

			// Calculate the load score component
			weights := models.DefaultHealthScoreWeights
			result := CalculateHealthScoreWithDetails(metrics, weights)

			require.NotNil(t, result)
			assert.Equal(t, float64(tc.expectedScore), result.LoadScore)
		})
	}
}

// TestCalculateHostHealthScore_CPUScoreClamping tests CPU score clamping
func TestCalculateHostHealthScore_CPUScoreClamping(t *testing.T) {
	t.Parallel()

	t.Run("CPU score clamped to 0 when usage exceeds 100%", func(t *testing.T) {
		t.Parallel()

		metrics := &models.HostMetrics{
			CollectorID:       uuid.New(),
			CpuUser:           60,
			CpuSystem:         30,
			CpuIowait:         20, // Total = 110, would be -10 without clamping
			MemoryUsedPercent: 0,
			DiskUsedPercent:   0,
			CpuLoad1m:         0.5,
		}

		weights := models.DefaultHealthScoreWeights
		result := CalculateHealthScoreWithDetails(metrics, weights)

		require.NotNil(t, result)
		assert.GreaterOrEqual(t, result.CpuScore, 0.0, "CPU score should be clamped to >= 0")
		assert.LessOrEqual(t, result.CpuScore, 100.0, "CPU score should be clamped to <= 100")
	})
}

// TestNewHealthScoreWithCollectorID tests the factory function
func TestNewHealthScoreWithCollectorID(t *testing.T) {
	t.Parallel()

	collectorID := uuid.New()
	now := time.Now()

	score := NewHealthScoreWithCollectorID(
		85, // score
		collectorID,
		now,
		80.0, // cpuScore
		90.0, // memoryScore
		75.0, // diskScore
		95.0, // loadScore
		map[string]interface{}{
			"test_key": "test_value",
		},
	)

	require.NotNil(t, score)
	assert.Equal(t, 85, score.HealthScore)
	assert.Equal(t, "healthy", score.Status) // 85 >= 80
	assert.Equal(t, collectorID, score.CollectorID)
	assert.Equal(t, now, score.Time)
	assert.Equal(t, 80.0, score.CpuScore)
	assert.Equal(t, 90.0, score.MemoryScore)
	assert.Equal(t, 75.0, score.DiskScore)
	assert.Equal(t, 95.0, score.LoadScore)
	assert.Contains(t, score.CalculationDetails, "test_key")
}
