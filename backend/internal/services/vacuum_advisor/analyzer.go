package vacuum_advisor

import (
	"context"
	"math"
	"time"
)

// VacuumAnalyzer provides VACUUM recommendations and analysis
type VacuumAnalyzer struct {
	db             interface{} // Would be *storage.PostgresDB in production
	costCalculator *CostCalculator
}

// NewVacuumAnalyzer creates a new VACUUM analyzer
func NewVacuumAnalyzer(db interface{}) *VacuumAnalyzer {
	return &VacuumAnalyzer{
		db:             db,
		costCalculator: NewCostCalculator(),
	}
}

// AnalyzeDatabase returns VACUUM recommendations for all tables in a database
func (va *VacuumAnalyzer) AnalyzeDatabase(ctx context.Context, databaseID int64) ([]VacuumRecommendation, error) {
	var recommendations []VacuumRecommendation

	// In production, this would query pg_stat_user_tables
	// For now, return empty slice
	if databaseID <= 0 {
		return recommendations, nil
	}

	return recommendations, nil
}

// AnalyzeTable returns VACUUM recommendation for a specific table
func (va *VacuumAnalyzer) AnalyzeTable(ctx context.Context, databaseID int64, tableName string) (*VacuumRecommendation, error) {
	if databaseID <= 0 || tableName == "" {
		return nil, nil
	}

	// In production, this would query pg_stat_user_tables for the specific table
	// Return a base recommendation
	return &VacuumRecommendation{
		DatabaseID: databaseID,
		TableName:  tableName,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// analyzeTableMetrics analyzes a table's VACUUM metrics and returns a recommendation
func (va *VacuumAnalyzer) analyzeTableMetrics(ctx context.Context, metrics *VacuumMetrics) *VacuumRecommendation {
	if metrics == nil {
		return nil
	}

	rec := &VacuumRecommendation{
		DatabaseID:        metrics.DatabaseID,
		TableName:         metrics.TableName,
		TableSize:         metrics.TableSize,
		DeadTuplesCount:   metrics.DeadTuples,
		DeadTuplesRatio:   metrics.DeadTuplesRatio,
		AutovacuumEnabled: metrics.AutovacuumEnabled,
		LastVacuum:        metrics.LastVacuum,
		LastAutovacuum:    metrics.LastAutovacuum,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Determine recommendation type
	rec.RecommendationType = va.selectRecommendationType(metrics)

	// Calculate estimated gain (space to be recovered)
	rec.EstimatedGain = va.calculateEstimatedGain(ctx, metrics.TableSize, metrics.DeadTuples, metrics.DeadTuplesRatio)

	return rec
}

// selectRecommendationType determines the best VACUUM action for a table
func (va *VacuumAnalyzer) selectRecommendationType(metrics *VacuumMetrics) string {
	// If autovacuum is disabled, recommend tuning it
	if !metrics.AutovacuumEnabled {
		return "tune_autovacuum"
	}

	// If dead tuple ratio is very high (>20%), recommend full vacuum
	if metrics.DeadTuplesRatio > 20.0 {
		return "full_vacuum"
	}

	// If dead tuple ratio is moderate (5-20%), recommend full vacuum
	if metrics.DeadTuplesRatio > 5.0 {
		return "full_vacuum"
	}

	// If dead tuple ratio is low (<5%), analyze only
	return "analyze_only"
}

// calculateEstimatedGain calculates the approximate space that could be recovered
func (va *VacuumAnalyzer) calculateEstimatedGain(ctx context.Context, tableSize int64, deadTuples int64, deadRatio float64) float64 {
	if tableSize <= 0 || deadTuples <= 0 {
		return 0.0
	}

	// Estimate space per tuple (8KB for header + data, varies by table)
	avgTupleSize := 8192.0 // Average PostgreSQL tuple size in bytes

	// Calculate recoverable space
	// Dead tuples can be reused by VACUUM
	recoverable := float64(deadTuples) * avgTupleSize

	// Apply a realistic recovery factor (not all space may be recoverable due to bloat)
	recoveryFactor := 0.85 // 85% of dead space typically recoverable
	if deadRatio < 10.0 {
		recoveryFactor = 0.7 // Lower recovery rate for tables with less bloat
	}

	estimated := recoverable * recoveryFactor

	// Cap at table size to avoid overestimation
	if estimated > float64(tableSize) {
		estimated = float64(tableSize) * 0.9
	}

	return math.Max(estimated, 0.0)
}

// GetAutovacuumConfig returns current autovacuum configuration for a database
func (va *VacuumAnalyzer) GetAutovacuumConfig(ctx context.Context, databaseID int64) []AutovacuumConfig {
	var configs []AutovacuumConfig

	if databaseID <= 0 {
		return configs
	}

	// In production, this would query pg_settings and pg_class
	// Return empty slice for now
	return configs
}

// TuneAutovacuum provides autovacuum parameter recommendations for a table
func (va *VacuumAnalyzer) TuneAutovacuum(ctx context.Context, databaseID int64, tableName string) []AutovacuumTuning {
	var tunings []AutovacuumTuning

	if databaseID <= 0 || tableName == "" {
		return tunings
	}

	// Recommend autovacuum_naptime adjustment
	tunings = append(tunings, AutovacuumTuning{
		TableName:           tableName,
		Parameter:           "autovacuum_naptime",
		CurrentValue:        "1min",
		RecommendedValue:    "30s",
		Rationale:           "More frequent autovacuum helps prevent table bloat",
		ExpectedImprovement: 15.0,
	})

	// Recommend scale factor adjustment for high-churn tables
	tunings = append(tunings, AutovacuumTuning{
		TableName:           tableName,
		Parameter:           "autovacuum_vacuum_scale_factor",
		CurrentValue:        "0.1",
		RecommendedValue:    "0.05",
		Rationale:           "Lower scale factor triggers vacuum more frequently",
		ExpectedImprovement: 20.0,
	})

	// Recommend threshold adjustment
	tunings = append(tunings, AutovacuumTuning{
		TableName:           tableName,
		Parameter:           "autovacuum_vacuum_threshold",
		CurrentValue:        "50",
		RecommendedValue:    "1000",
		Rationale:           "Higher threshold reduces unnecessary autovacuum runs",
		ExpectedImprovement: 10.0,
	})

	return tunings
}

// GetHighPriorityTables returns tables that need immediate VACUUM attention
func (va *VacuumAnalyzer) GetHighPriorityTables(ctx context.Context, databaseID int64) []VacuumRecommendation {
	var recommendations []VacuumRecommendation

	if databaseID <= 0 {
		return recommendations
	}

	// In production, would query tables with:
	// - Dead tuple ratio > 30%
	// - No vacuum in > 7 days
	// - Autovacuum disabled
	// - Rapid tuple churn detected

	return recommendations
}
