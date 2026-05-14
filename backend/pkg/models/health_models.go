package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// HEALTH SCORE MODELS (HOST-04)
// ============================================================================

// HealthScoreWeights defines the weight of each metric in the overall health score
type HealthScoreWeights struct {
	CPU         float64 `json:"cpu" db:"cpu"`                   // Default: 0.30 (30%)
	Memory      float64 `json:"memory" db:"memory"`             // Default: 0.25 (25%)
	Disk        float64 `json:"disk" db:"disk"`                 // Default: 0.25 (25%)
	LoadAverage float64 `json:"load_average" db:"load_average"` // Default: 0.20 (20%)
}

// DefaultHealthScoreWeights provides standard weights for health score calculation
var DefaultHealthScoreWeights = HealthScoreWeights{
	CPU:         0.30,
	Memory:      0.25,
	Disk:        0.25,
	LoadAverage: 0.20,
}

// HealthScore represents a calculated host health score with component breakdown
type HealthScore struct {
	Time               time.Time              `json:"time" db:"time"`
	CollectorID        uuid.UUID              `json:"collector_id" db:"collector_id"`
	HealthScore        int                    `json:"health_score" db:"health_score"`               // 0-100
	Status             string                 `json:"status" db:"status"`                           // healthy, degraded, warning, critical
	CpuScore           float64                `json:"cpu_score" db:"cpu_score"`                     // Component score 0-100
	MemoryScore        float64                `json:"memory_score" db:"memory_score"`               // Component score 0-100
	DiskScore          float64                `json:"disk_score" db:"disk_score"`                   // Component score 0-100
	LoadScore          float64                `json:"load_score" db:"load_score"`                   // Component score 0-100
	CalculationDetails map[string]interface{} `json:"calculation_details" db:"calculation_details"` // Contributing factors
}

// HealthScoreResponse contains health score data with context metrics
type HealthScoreResponse struct {
	HealthScore   *HealthScore `json:"health_score"`
	LatestMetrics *HostMetrics `json:"latest_metrics"` // Snapshot of metrics used for calculation
}

// HealthScoreHistoryResponse contains historical health scores with pagination
type HealthScoreHistoryResponse struct {
	Scores     []*HealthScore   `json:"scores"`
	Pagination PaginationParams `json:"pagination"`
}
