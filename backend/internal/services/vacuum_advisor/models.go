package vacuum_advisor

import (
	"time"
)

// VacuumRecommendation represents a single VACUUM recommendation
type VacuumRecommendation struct {
	ID                 int64
	DatabaseID         int64
	TableName          string
	TableSize          int64
	DeadTuplesCount    int64
	DeadTuplesRatio    float64
	AutovacuumEnabled  bool
	AutovacuumNaptime  string
	LastVacuum         *time.Time
	LastAutovacuum     *time.Time
	RecommendationType string // 'full_vacuum', 'analyze_only', 'tune_autovacuum'
	EstimatedGain      float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// AutovacuumConfig represents current and recommended autovacuum settings
type AutovacuumConfig struct {
	ID               int64
	DatabaseID       int64
	TableName        string
	SettingName      string
	CurrentValue     string
	RecommendedValue string
	Impact           string // 'high', 'medium', 'low'
	CreatedAt        time.Time
}

// AutovacuumTuning represents a tuning recommendation for autovacuum
type AutovacuumTuning struct {
	TableName           string
	Parameter           string
	CurrentValue        string
	RecommendedValue    string
	Rationale           string
	ExpectedImprovement float64 // percentage improvement
}

// VacuumMetrics represents VACUUM-related metrics for a table
type VacuumMetrics struct {
	DatabaseID        int64
	TableName         string
	TableSize         int64
	DeadTuples        int64
	LiveTuples        int64
	DeadTuplesRatio   float64
	LastVacuum        *time.Time
	LastAutovacuum    *time.Time
	VacuumFrequency   string // estimated frequency
	AutovacuumEnabled bool
}
