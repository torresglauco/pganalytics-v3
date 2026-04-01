package vacuum_advisor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestVacuumRecommendationModelCreation verifies the model can be instantiated
func TestVacuumRecommendationModelCreation(t *testing.T) {
	now := time.Now()
	rec := &VacuumRecommendation{
		ID:                  1,
		DatabaseID:          1,
		TableName:           "test_table",
		TableSize:           1000000,
		DeadTuplesCount:     50000,
		DeadTuplesRatio:     5.0,
		AutovacuumEnabled:   true,
		RecommendationType:  "full_vacuum",
		EstimatedGain:       50000.00,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	assert.NotNil(t, rec)
	assert.Equal(t, int64(1), rec.ID)
	assert.Equal(t, "test_table", rec.TableName)
	assert.Equal(t, 5.0, rec.DeadTuplesRatio)
}

// TestAutovacuumConfigModelCreation verifies the model can be instantiated
func TestAutovacuumConfigModelCreation(t *testing.T) {
	now := time.Now()
	config := &AutovacuumConfig{
		ID:               1,
		DatabaseID:       1,
		TableName:        "test_table",
		SettingName:      "autovacuum_naptime",
		CurrentValue:     "1min",
		RecommendedValue: "30s",
		Impact:           "high",
		CreatedAt:        now,
	}

	assert.NotNil(t, config)
	assert.Equal(t, "autovacuum_naptime", config.SettingName)
	assert.Equal(t, "1min", config.CurrentValue)
	assert.Equal(t, "high", config.Impact)
}

// TestAutovaQUumTuningModelCreation verifies the tuning model can be instantiated
func TestAutovacuumTuningModelCreation(t *testing.T) {
	tuning := &AutovacuumTuning{
		TableName:           "test_table",
		Parameter:           "autovacuum_naptime",
		CurrentValue:        "1min",
		RecommendedValue:    "30s",
		Rationale:           "High write rate requires more frequent vacuuming",
		ExpectedImprovement: 25.5,
	}

	assert.NotNil(t, tuning)
	assert.Equal(t, "test_table", tuning.TableName)
	assert.Equal(t, 25.5, tuning.ExpectedImprovement)
}

// TestVacuumMetricsModelCreation verifies the metrics model can be instantiated
func TestVacuumMetricsModelCreation(t *testing.T) {
	now := time.Now()
	metrics := &VacuumMetrics{
		DatabaseID:        1,
		TableName:         "test_table",
		TableSize:         1000000,
		DeadTuples:        50000,
		LiveTuples:        950000,
		DeadTuplesRatio:   5.0,
		LastVacuum:        &now,
		LastAutovacuum:    &now,
		VacuumFrequency:   "daily",
		AutovacuumEnabled: true,
	}

	assert.NotNil(t, metrics)
	assert.Equal(t, 5.0, metrics.DeadTuplesRatio)
	assert.True(t, metrics.AutovacuumEnabled)
}

// TestVacuumRecommendationTypeValidation verifies recommendation type values
func TestVacuumRecommendationTypeValidation(t *testing.T) {
	validTypes := []string{"full_vacuum", "analyze_only", "tune_autovacuum"}

	for _, rt := range validTypes {
		rec := &VacuumRecommendation{
			RecommendationType: rt,
		}
		assert.NotEmpty(t, rec.RecommendationType)
		assert.Contains(t, validTypes, rec.RecommendationType)
	}
}

// TestAutovacuumConfigImpactValidation verifies impact level values
func TestAutovacuumConfigImpactValidation(t *testing.T) {
	validImpacts := []string{"high", "medium", "low"}

	for _, impact := range validImpacts {
		config := &AutovacuumConfig{
			Impact: impact,
		}
		assert.Contains(t, validImpacts, config.Impact)
	}
}
