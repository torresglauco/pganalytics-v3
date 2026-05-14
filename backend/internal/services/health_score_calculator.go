package services

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// CalculateHostHealthScore computes a health score from 0-100 based on resource utilization
// using weighted formula: CPU (30%) + Memory (25%) + Disk (25%) + Load (20%)
func CalculateHostHealthScore(metrics *models.HostMetrics, weights models.HealthScoreWeights) int {
	if metrics == nil {
		return 0
	}

	// CPU Score: 100 - (user + system + iowait), clamp to 0-100
	cpuScore := 100.0 - (metrics.CpuUser + metrics.CpuSystem + metrics.CpuIowait)
	cpuScore = math.Max(0, math.Min(100, cpuScore))

	// Memory Score: 100 - used_percent, clamp to 0-100
	memoryScore := 100.0 - metrics.MemoryUsedPercent
	memoryScore = math.Max(0, math.Min(100, memoryScore))

	// Disk Score: 100 - used_percent, clamp to 0-100
	diskScore := 100.0 - metrics.DiskUsedPercent
	diskScore = math.Max(0, math.Min(100, diskScore))

	// Load Score: based on CPU load thresholds
	// load <= 1.0: score = 100 (light load)
	// load > 1.0 AND load <= 2.0: score = 50 (moderate load)
	// load > 2.0: score = 0 (heavy load)
	// Note: For accurate load assessment relative to CPU cores,
	// use HostInventory.CpuCores in a combined calculation
	loadScore := 100.0
	if metrics.CpuLoad1m > 1.0 {
		if metrics.CpuLoad1m > 2.0 {
			loadScore = 0
		} else {
			loadScore = 50
		}
	}

	// Weighted average
	totalScore := (cpuScore * weights.CPU) +
		(memoryScore * weights.Memory) +
		(diskScore * weights.Disk) +
		(loadScore * weights.LoadAverage)

	return int(math.Round(totalScore))
}

// GetHealthStatus returns a status string based on score
// score >= 80: healthy
// score >= 60: degraded
// score >= 40: warning
// score < 40: critical
func GetHealthStatus(score int) string {
	switch {
	case score >= 80:
		return "healthy"
	case score >= 60:
		return "degraded"
	case score >= 40:
		return "warning"
	default:
		return "critical"
	}
}

// CalculateHealthScoreWithDetails calculates the health score with component breakdown and details
func CalculateHealthScoreWithDetails(metrics *models.HostMetrics, weights models.HealthScoreWeights) *models.HealthScore {
	if metrics == nil {
		return nil
	}

	// Calculate individual component scores
	cpuScore := 100.0 - (metrics.CpuUser + metrics.CpuSystem + metrics.CpuIowait)
	cpuScore = math.Max(0, math.Min(100, cpuScore))

	memoryScore := 100.0 - metrics.MemoryUsedPercent
	memoryScore = math.Max(0, math.Min(100, memoryScore))

	diskScore := 100.0 - metrics.DiskUsedPercent
	diskScore = math.Max(0, math.Min(100, diskScore))

	// Load score calculation (threshold-based without CPU cores)
	loadScore := 100.0
	if metrics.CpuLoad1m > 1.0 {
		if metrics.CpuLoad1m > 2.0 {
			loadScore = 0
		} else {
			loadScore = 50
		}
	}

	// Calculate overall score
	totalScore := CalculateHostHealthScore(metrics, weights)
	status := GetHealthStatus(totalScore)

	// Build calculation details
	details := map[string]interface{}{
		"cpu_user":        metrics.CpuUser,
		"cpu_system":      metrics.CpuSystem,
		"cpu_iowait":      metrics.CpuIowait,
		"cpu_idle":        metrics.CpuIdle,
		"cpu_load_1m":     metrics.CpuLoad1m,
		"memory_used_pct": metrics.MemoryUsedPercent,
		"disk_used_pct":   metrics.DiskUsedPercent,
		"weights": map[string]float64{
			"cpu":          weights.CPU,
			"memory":       weights.Memory,
			"disk":         weights.Disk,
			"load_average": weights.LoadAverage,
		},
	}

	return &models.HealthScore{
		Time:               time.Now(),
		CollectorID:        metrics.CollectorID,
		HealthScore:        totalScore,
		Status:             status,
		CpuScore:           cpuScore,
		MemoryScore:        memoryScore,
		DiskScore:          diskScore,
		LoadScore:          loadScore,
		CalculationDetails: details,
	}
}

// NewHealthScoreWithCollectorID creates a health score with a specific collector ID and time
// Useful when calculating scores for storage
func NewHealthScoreWithCollectorID(score int, collectorID uuid.UUID, timestamp time.Time, cpuScore, memoryScore, diskScore, loadScore float64, details map[string]interface{}) *models.HealthScore {
	status := GetHealthStatus(score)

	return &models.HealthScore{
		Time:               timestamp,
		CollectorID:        collectorID,
		HealthScore:        score,
		Status:             status,
		CpuScore:           cpuScore,
		MemoryScore:        memoryScore,
		DiskScore:          diskScore,
		LoadScore:          loadScore,
		CalculationDetails: details,
	}
}
