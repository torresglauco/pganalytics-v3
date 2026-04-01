package vacuum_advisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCostCalculatorCreation verifies CostCalculator instantiation
func TestCostCalculatorCreation(t *testing.T) {
	calc := NewCostCalculator()
	assert.NotNil(t, calc)
	assert.Equal(t, 1.0, calc.SeqScanCostPerPage)
	assert.Equal(t, 4.0, calc.RandomAccessCostPerTuple)
}

// TestEstimateVacuumDuration returns reasonable time estimates
func TestEstimateVacuumDuration(t *testing.T) {
	calc := NewCostCalculator()

	// Small table (1MB)
	duration := calc.EstimateVacuumDuration(1_000_000, 10_000)
	assert.Greater(t, duration, 0.0)
	assert.Less(t, duration, 1.0) // Should be very fast

	// Large table (1GB)
	duration = calc.EstimateVacuumDuration(1_000_000_000, 10_000_000)
	assert.Greater(t, duration, 1.0) // Should take longer than 1 second
}

// TestEstimateVacuumDurationWithZeroSize handles edge cases
func TestEstimateVacuumDurationWithZeroSize(t *testing.T) {
	calc := NewCostCalculator()

	duration := calc.EstimateVacuumDuration(0, 0)
	assert.Equal(t, 0.0, duration)
}

// TestEstimateVacuumImpact calculates realistic impact metrics
func TestEstimateVacuumImpact(t *testing.T) {
	calc := NewCostCalculator()

	// Small table impact (1% of DB)
	impact := calc.EstimateVacuumImpact(1_000_000_000, 10_000_000)
	assert.Greater(t, impact.QuerySlowdownFactor, 1.0)
	assert.Greater(t, impact.DiskIOIncrease, 1.0)

	// Large table impact (20% of DB) - should have higher impact
	impact = calc.EstimateVacuumImpact(1_000_000_000, 200_000_000)
	assert.Greater(t, impact.DiskIOIncrease, 1.0) // Definitely > 1.0
	assert.Greater(t, impact.QuerySlowdownFactor, 1.0) // Definitely > 1.0
	assert.True(t, impact.DiskIOIncrease >= 2.0 || impact.QuerySlowdownFactor >= 1.1) // At least one is significant
}

// TestCalculateOptimalSchedule provides reasonable schedule recommendations
func TestCalculateOptimalSchedule(t *testing.T) {
	calc := NewCostCalculator()

	// Small table - should be runnable anytime
	rec := calc.CalculateOptimalSchedule(1_000_000, 8)
	assert.NotEmpty(t, rec.RecommendedWindow)
	assert.Greater(t, rec.EstimatedDuration, 0.0)

	// Large table - should require maintenance window
	rec = calc.CalculateOptimalSchedule(10_000_000_000, 8)
	assert.NotEmpty(t, rec.RecommendedWindow)
	assert.NotEmpty(t, rec.Rationale)
}

// TestCalculateRecoverableSpace estimates realistic space recovery
func TestCalculateRecoverableSpace(t *testing.T) {
	calc := NewCostCalculator()

	// No dead tuples
	space := calc.CalculateRecoverableSpace(1_000_000, 0)
	assert.Equal(t, int64(0), space)

	// 10% dead tuples
	space = calc.CalculateRecoverableSpace(1_000_000, 10.0)
	assert.Greater(t, space, int64(0))
	assert.Less(t, space, int64(100_000)) // Can't recover more than dead space

	// 50% dead tuples
	space = calc.CalculateRecoverableSpace(1_000_000, 50.0)
	assert.Greater(t, space, int64(0))
}

// TestCalculateIndexBlowup estimates index bloat from dead tuples
func TestCalculateIndexBlowup(t *testing.T) {
	calc := NewCostCalculator()

	// No dead tuples
	bloat := calc.CalculateIndexBlowup(1_000_000, 0, 5)
	assert.Equal(t, int64(0), bloat)

	// 10k dead tuples across 5 indexes
	bloat = calc.CalculateIndexBlowup(1_000_000, 10_000, 5)
	assert.Greater(t, bloat, int64(0))

	// More indexes should increase bloat
	bloatWith10Indexes := calc.CalculateIndexBlowup(1_000_000, 10_000, 10)
	assert.Greater(t, bloatWith10Indexes, bloat)
}

// TestCalculateAutovacuumEfficiency calculates efficiency metrics
func TestCalculateAutovacuumEfficiency(t *testing.T) {
	calc := NewCostCalculator()

	// Perfect autovacuum - matches expected churn
	efficiency := calc.CalculateAutovacuumEfficiency(
		1_000_000,      // table size
		10.0,           // 10% dead tuples
		1.0,            // vacuum 1 day ago
		10_000,         // 10k tuples/day churn
	)
	assert.Greater(t, efficiency, 0.0)
	assert.LessOrEqual(t, efficiency, 200.0)

	// Falling behind - actual dead > expected
	efficiency = calc.CalculateAutovacuumEfficiency(
		1_000_000,      // table size
		50.0,           // 50% dead (behind)
		1.0,            // vacuum 1 day ago
		10_000,         // expected only 10k
	)
	assert.Less(t, efficiency, 100.0) // Less efficient than ideal

	// Ahead of schedule
	efficiency = calc.CalculateAutovacuumEfficiency(
		1_000_000,      // table size
		1.0,            // only 1% dead (ahead)
		1.0,            // vacuum 1 day ago
		100_000,        // high churn expected
	)
	assert.Greater(t, efficiency, 100.0) // More efficient than expected
}

// TestCalculateRecoverableSpaceWithHighBloat handles high bloat percentage
func TestCalculateRecoverableSpaceWithHighBloat(t *testing.T) {
	calc := NewCostCalculator()

	// 80% bloat - high bloat scenario
	space := calc.CalculateRecoverableSpace(1_000_000, 80.0)
	assert.Greater(t, space, int64(600_000)) // Should recover most of 800KB
	assert.Less(t, space, int64(800_000))    // But not all (due to overhead)
}

// TestEstimateVacuumDurationScales with table size
func TestEstimateVacuumDurationScales(t *testing.T) {
	calc := NewCostCalculator()

	smallDuration := calc.EstimateVacuumDuration(1_000_000, 10_000)
	largeDuration := calc.EstimateVacuumDuration(1_000_000_000, 10_000_000)

	// Larger table should take longer
	assert.Greater(t, largeDuration, smallDuration)

	// Should scale roughly linearly with size
	ratio := largeDuration / smallDuration
	assert.Greater(t, ratio, 100.0) // 1000x size increase should scale significantly
}
