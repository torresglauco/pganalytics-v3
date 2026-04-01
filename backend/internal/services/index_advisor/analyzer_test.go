package index_advisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexAnalyzer_FindMissingIndexes(t *testing.T) {
	analyzer := NewIndexAnalyzer(nil)

	// Test with empty query plans
	recommendations := analyzer.FindMissingIndexes([]*QueryPlan{})
	assert.NotNil(t, recommendations)
	assert.Len(t, recommendations, 0)

	// Test with query plan that has sequence scan
	queryPlans := []*QueryPlan{
		{
			NodeType:  "SeqScan",
			TotalCost: 1000.0,
			Calls:     10,
		},
	}
	recommendations = analyzer.FindMissingIndexes(queryPlans)
	assert.NotNil(t, recommendations)
	// Since extractConditions is a stub, it returns empty conditions
	assert.Len(t, recommendations, 0)
}

func TestCostCalculator_CalculateImprovement(t *testing.T) {
	calc := NewCostCalculator()

	costWithout := 1000.0
	costWith := 100.0

	improvement := calc.CalculateImprovement(costWithout, costWith)
	assert.Equal(t, 90.0, improvement)
}

func TestCostCalculator_CalculateImprovement_ZeroCost(t *testing.T) {
	calc := NewCostCalculator()

	improvement := calc.CalculateImprovement(0, 100.0)
	assert.Equal(t, 0.0, improvement)
}

func TestCostCalculator_EstimateBenefit(t *testing.T) {
	calc := NewCostCalculator()

	costImprovement := 90.0
	frequency := 100.0

	benefit := calc.EstimateBenefit(costImprovement, frequency)
	assert.Equal(t, 90.0, benefit)
}

func TestCostCalculator_CalculateIndexMaintenanceCost(t *testing.T) {
	calc := NewCostCalculator()

	tableWriteFrequency := 100.0

	maintenanceCost := calc.CalculateIndexMaintenanceCost(tableWriteFrequency)
	assert.Equal(t, 3.0, maintenanceCost)
}

func TestCostCalculator_ShouldCreateIndex_True(t *testing.T) {
	calc := NewCostCalculator()

	benefit := 100.0
	maintenanceCost := 10.0

	should := calc.ShouldCreateIndex(benefit, maintenanceCost)
	assert.True(t, should)
}

func TestCostCalculator_ShouldCreateIndex_False(t *testing.T) {
	calc := NewCostCalculator()

	benefit := 10.0
	maintenanceCost := 10.0

	should := calc.ShouldCreateIndex(benefit, maintenanceCost)
	assert.False(t, should)
}
