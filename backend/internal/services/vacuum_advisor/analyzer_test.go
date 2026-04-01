package vacuum_advisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockPostgresDB mocks the database interface for testing
type MockPostgresDB struct {
	vacuumMetrics    map[string]*VacuumMetrics
	autovacuumConfig map[string]*AutovacuumConfig
	recommendations  []VacuumRecommendation
	err              error
}

func NewMockPostgresDB() *MockPostgresDB {
	return &MockPostgresDB{
		vacuumMetrics:    make(map[string]*VacuumMetrics),
		autovacuumConfig: make(map[string]*AutovacuumConfig),
		recommendations:  make([]VacuumRecommendation, 0),
	}
}

// TestVacuumAnalyzerCreation verifies VacuumAnalyzer can be created
func TestVacuumAnalyzerCreation(t *testing.T) {
	db := NewMockPostgresDB()
	analyzer := NewVacuumAnalyzer(db)
	assert.NotNil(t, analyzer)
}

// TestAnalyzeDatabaseWithHighDeadTuples identifies tables needing vacuum
func TestAnalyzeDatabaseWithHighDeadTuples(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())

	ctx := context.Background()

	// Test detection of high dead tuple ratio (> 20%)
	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "orders",
		TableSize:         10000000,
		DeadTuples:        3000000, // 30% dead
		LiveTuples:        7000000,
		DeadTuplesRatio:   30.0,
		AutovacuumEnabled: true,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.NotNil(t, rec)
	assert.Equal(t, "full_vacuum", rec.RecommendationType)
	assert.Greater(t, rec.EstimatedGain, float64(0))
}

// TestAnalyzeDatabaseWithLowDeadTuples identifies tables not needing vacuum
func TestAnalyzeDatabaseWithLowDeadTuples(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())

	ctx := context.Background()

	// Test detection of low dead tuple ratio (< 5%)
	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "users",
		TableSize:         1000000,
		DeadTuples:        20000, // 2% dead
		LiveTuples:        980000,
		DeadTuplesRatio:   2.0,
		AutovacuumEnabled: true,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.NotNil(t, rec)
	// Low ratio should result in analyze_only or no action
	assert.True(t, rec.RecommendationType == "analyze_only" || rec.EstimatedGain < 100000)
}

// TestAnalyzeDatabaseWithDisabledAutovacuum identifies tables with autovacuum disabled
func TestAnalyzeDatabaseWithDisabledAutovacuum(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())

	ctx := context.Background()

	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "archived_data",
		TableSize:         50000000,
		DeadTuples:        1000000, // 2% but autovacuum disabled
		LiveTuples:        49000000,
		DeadTuplesRatio:   2.0,
		AutovacuumEnabled: false,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.NotNil(t, rec)
	assert.Equal(t, "tune_autovacuum", rec.RecommendationType)
}

// TestCalculateEstimatedGain verifies gain calculation is accurate
func TestCalculateEstimatedGain(t *testing.T) {
	analyzer := &VacuumAnalyzer{
		costCalculator: NewCostCalculator(),
	}

	ctx := context.Background()

	gain := analyzer.calculateEstimatedGain(ctx, 10000000, 1000000, 10.0)
	assert.Greater(t, gain, float64(0))
	// Dead space (1MB) should be recoverable
	assert.Greater(t, gain, 900000.0)
}

// TestDetectRapidTupleChurn identifies tables with high churn rates
func TestDetectRapidTupleChurn(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())

	ctx := context.Background()

	// Table with high dead/live ratio indicates heavy writes/deletes
	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "events",
		TableSize:         100000000,
		DeadTuples:        40000000, // 40% dead
		LiveTuples:        60000000,
		DeadTuplesRatio:   40.0,
		AutovacuumEnabled: true,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.NotNil(t, rec)
	assert.Equal(t, "full_vacuum", rec.RecommendationType)
	// High churn should result in larger estimated gain
	assert.Greater(t, rec.EstimatedGain, 10000000.0)
}

// TestRecommendationTypeSelection verifies correct type is selected
func TestRecommendationTypeSelection(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())
	ctx := context.Background()

	testCases := []struct {
		name               string
		deadRatio          float64
		autovacuumEnabled  bool
		expectedType       string
	}{
		{
			name:              "High dead ratio with autovacuum",
			deadRatio:         25.0,
			autovacuumEnabled: true,
			expectedType:      "full_vacuum",
		},
		{
			name:              "Disabled autovacuum",
			deadRatio:         5.0,
			autovacuumEnabled: false,
			expectedType:      "tune_autovacuum",
		},
		{
			name:              "Low dead ratio with autovacuum",
			deadRatio:         3.0,
			autovacuumEnabled: true,
			expectedType:      "analyze_only",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := &VacuumMetrics{
				DatabaseID:        1,
				TableName:         "test_table",
				TableSize:         1000000,
				DeadTuples:        int64(float64(1000000) * tc.deadRatio / 100),
				LiveTuples:        int64(float64(1000000) * (100 - tc.deadRatio) / 100),
				DeadTuplesRatio:   tc.deadRatio,
				AutovacuumEnabled: tc.autovacuumEnabled,
			}

			rec := analyzer.analyzeTableMetrics(ctx, metrics)
			assert.Equal(t, tc.expectedType, rec.RecommendationType)
		})
	}
}

// TestAutovacuumTuning generates reasonable tuning recommendations
func TestAutovacuumTuning(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())
	ctx := context.Background()

	tunings := analyzer.TuneAutovacuum(ctx, 1, "events")
	assert.NotNil(t, tunings)
	assert.Greater(t, len(tunings), 0)

	// Each tuning should have valid fields
	for _, tuning := range tunings {
		assert.NotEmpty(t, tuning.Parameter)
		assert.NotEmpty(t, tuning.CurrentValue)
		assert.NotEmpty(t, tuning.RecommendedValue)
		assert.NotEmpty(t, tuning.Rationale)
	}
}

// TestAutovacuumConfigRetrieval returns current autovacuum settings
func TestAutovacuumConfigRetrieval(t *testing.T) {
	db := NewMockPostgresDB()
	analyzer := NewVacuumAnalyzer(db)
	ctx := context.Background()

	configs := analyzer.GetAutovacuumConfig(ctx, 1)
	// Should return slice (could be empty)
	assert.IsType(t, []AutovacuumConfig{}, configs)
}

// TestAnalyzeTableWithoutVacuumHistory detects stale tables
func TestAnalyzeTableWithoutVacuumHistory(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())
	ctx := context.Background()

	// Table never vacuumed - very old or new
	twoWeeksAgo := time.Now().Add(-time.Hour * 24 * 14)
	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "stale_table",
		TableSize:         5000000,
		DeadTuples:        100000,
		LiveTuples:        4900000,
		DeadTuplesRatio:   2.0,
		LastVacuum:        &twoWeeksAgo,
		LastAutovacuum:    &twoWeeksAgo,
		AutovacuumEnabled: true,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.NotNil(t, rec)
	assert.True(t, rec.RecommendationType == "full_vacuum" || rec.RecommendationType == "analyze_only")
}

// TestGetAutovacuumConfigForSpecificTable returns config for specific table
func TestGetAutovacuumConfigForSpecificTable(t *testing.T) {
	db := NewMockPostgresDB()
	analyzer := NewVacuumAnalyzer(db)
	ctx := context.Background()

	// Should handle missing table gracefully
	configs := analyzer.GetAutovacuumConfig(ctx, 1)
	assert.IsType(t, []AutovacuumConfig{}, configs)
}

// TestRecommendationMetadataPopulation verifies all recommendation fields are set
func TestRecommendationMetadataPopulation(t *testing.T) {
	analyzer := NewVacuumAnalyzer(NewMockPostgresDB())
	ctx := context.Background()

	metrics := &VacuumMetrics{
		DatabaseID:        123,
		TableName:         "test_table",
		TableSize:         10000000,
		DeadTuples:        500000,
		LiveTuples:        9500000,
		DeadTuplesRatio:   5.0,
		AutovacuumEnabled: true,
	}

	rec := analyzer.analyzeTableMetrics(ctx, metrics)
	assert.Equal(t, int64(123), rec.DatabaseID)
	assert.Equal(t, "test_table", rec.TableName)
	assert.Equal(t, int64(10000000), rec.TableSize)
	assert.Equal(t, int64(500000), rec.DeadTuplesCount)
	assert.Equal(t, 5.0, rec.DeadTuplesRatio)
	assert.True(t, rec.AutovacuumEnabled)
	assert.NotZero(t, rec.CreatedAt)
	assert.NotZero(t, rec.UpdatedAt)
}
