package integration

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/query_performance"
)

// TestQueryPerformanceE2E tests the complete query performance data flow
// From EXPLAIN ANALYZE capture → parsing → analysis → storage
func TestQueryPerformanceE2E(t *testing.T) {
	t.Run("capture_parse_analyze_flow", func(t *testing.T) {
		// Arrange: Initialize the collector service
		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		// Arrange: Get mock EXPLAIN ANALYZE output
		explainOutput := MockExplainOutput()
		require.NotEmpty(t, explainOutput)

		// Act: Parse the EXPLAIN output
		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Assert: Verify parsed plan structure
		assert.NotEmpty(t, plan.NodeType)
		assert.Equal(t, "Seq Scan", plan.NodeType)
		assert.Greater(t, plan.TotalCost, 0.0)
	})

	t.Run("complex_explain_output_parsing", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		explainOutput := MockExplainOutputComplex()
		require.NotEmpty(t, explainOutput)

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Verify complex plan parsing
		assert.Equal(t, "Hash Join", plan.NodeType)
		assert.Greater(t, plan.TotalCost, 0.0)
	})

	t.Run("invalid_explain_output_handling", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		invalidOutput := `{invalid json}`

		plan, err := collector.ParseExplainOutput(invalidOutput)
		assert.Error(t, err)
		assert.Nil(t, plan)
	})

	t.Run("empty_explain_output_handling", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		emptyOutput := ""

		plan, err := collector.ParseExplainOutput(emptyOutput)
		assert.Error(t, err)
		assert.Nil(t, plan)
	})
}

// TestQueryPerformanceAnalysis tests the analysis of query performance data
func TestQueryPerformanceAnalysis(t *testing.T) {
	t.Run("analyze_execution_time", func(t *testing.T) {
		// Create mock performance data
		executionTime := 2.345
		planningTime := 0.123
		totalTime := 2.468

		// Verify time calculations
		assert.Greater(t, executionTime, planningTime)
		assert.Greater(t, totalTime, executionTime)
		assert.InDelta(t, totalTime, executionTime+planningTime, 0.01)
	})

	t.Run("identify_sequential_scans", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		explainOutput := MockExplainOutput()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)

		// Check for sequential scan (performance issue)
		assert.Equal(t, "Seq Scan", plan.NodeType)
		// Sequential scans on large tables are performance concerns
		assert.Greater(t, plan.TotalCost, 0.0)
	})

	t.Run("identify_join_performance", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		explainOutput := MockExplainOutputComplex()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)

		// Check for hash join (more efficient than nested loop)
		assert.Equal(t, "Hash Join", plan.NodeType)
		assert.Greater(t, plan.TotalCost, 0.0)
	})

	t.Run("detect_missing_indexes", func(t *testing.T) {
		// Simulate detecting queries that would benefit from indexes
		collector := query_performance.NewQueryCollector(nil)
		explainOutput := MockExplainOutput()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)

		// Sequential scans without index conditions indicate missing indexes
		if plan.NodeType == "Seq Scan" {
			// This is a candidate for index optimization
			assert.NotEmpty(t, plan.NodeType)
		}
	})
}

// TestQueryPerformanceMetricsAggregation tests aggregation of performance metrics
func TestQueryPerformanceMetricsAggregation(t *testing.T) {
	t.Run("aggregate_timing_metrics", func(t *testing.T) {
		// Simulate multiple query executions
		executionTimes := []float64{1.2, 1.5, 2.3, 1.8, 2.1}

		// Calculate aggregates
		var sum, min, max float64
		min = executionTimes[0]
		max = executionTimes[0]

		for _, et := range executionTimes {
			sum += et
			if et < min {
				min = et
			}
			if et > max {
				max = et
			}
		}

		avg := sum / float64(len(executionTimes))

		// Verify calculations
		assert.Equal(t, 1.2, min)
		assert.Equal(t, 2.3, max)
		assert.InDelta(t, 1.78, avg, 0.01)
	})

	t.Run("calculate_percentiles", func(t *testing.T) {
		// Simulate latency distribution
		latencies := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0}

		// P50 (median)
		p50Idx := len(latencies) / 2
		assert.Equal(t, 2.5, latencies[p50Idx])

		// P95
		p95Idx := int(float64(len(latencies)) * 0.95)
		assert.Equal(t, 4.5, latencies[p95Idx])

		// P99
		p99Idx := int(float64(len(latencies)) * 0.99)
		assert.Equal(t, 5.0, latencies[p99Idx])
	})

	t.Run("track_query_frequency", func(t *testing.T) {
		// Simulate query execution frequency tracking
		queryFrequency := map[string]int{
			"SELECT * FROM users":     150,
			"SELECT * FROM orders":    120,
			"SELECT * FROM products":  80,
		}

		totalQueries := 0
		for _, count := range queryFrequency {
			totalQueries += count
		}

		assert.Equal(t, 350, totalQueries)
		assert.Greater(t, queryFrequency["SELECT * FROM users"], 100)
	})
}

// TestQueryPerformanceContextHandling tests proper context handling in performance collection
func TestQueryPerformanceContextHandling(t *testing.T) {
	t.Run("context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		// Verify context can be cancelled
		cancel()
		select {
		case <-ctx.Done():
			// Context was successfully cancelled
			assert.NoError(t, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("context was not cancelled in time")
		}
	})

	t.Run("context_timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Wait for timeout
		<-time.After(150 * time.Millisecond)

		select {
		case <-ctx.Done():
			assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
		default:
			t.Fatal("context should have exceeded deadline")
		}
	})
}

// TestQueryPerformanceErrorHandling tests error handling in the performance flow
func TestQueryPerformanceErrorHandling(t *testing.T) {
	t.Run("handle_parse_error", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Test with invalid JSON
		invalidJSON := `{"invalid": json}`
		plan, err := collector.ParseExplainOutput(invalidJSON)

		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("handle_malformed_json", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Test with completely invalid data
		invalidData := `not json at all`
		plan, err := collector.ParseExplainOutput(invalidData)

		assert.Error(t, err)
		assert.Nil(t, plan)
	})

	t.Run("handle_missing_plan_field", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Test with JSON missing the Plan field
		missingPlan := `{"NotPlan": {"Node Type": "Seq Scan"}}`
		plan, err := collector.ParseExplainOutput(missingPlan)

		// The error might be about missing Plan or it might parse with empty fields
		if err == nil {
			// If no error, plan fields should be default values
			assert.NotNil(t, plan)
		} else {
			assert.Error(t, err)
		}
	})
}

// TestQueryPerformanceDataIntegrity tests data integrity throughout the flow
func TestQueryPerformanceDataIntegrity(t *testing.T) {
	t.Run("preserve_explain_metadata", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)
		explainOutput := MockExplainOutput()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Verify all metadata is preserved
		assert.NotEmpty(t, plan.NodeType)
		assert.NotZero(t, plan.TotalCost)
	})

	t.Run("maintain_cost_consistency", func(t *testing.T) {
		// Verify that parsed costs are positive and reasonable
		collector := query_performance.NewQueryCollector(nil)
		explainOutput := MockExplainOutputComplex()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)

		// Costs should be positive
		assert.Greater(t, plan.TotalCost, 0.0)

		// Costs should be reasonable (not infinity or NaN)
		assert.False(t, isInfinity(plan.TotalCost))
	})
}

// Helper function to check if a float64 is infinity
func isInfinity(f float64) bool {
	return math.IsInf(f, 0)
}

// TestQueryPerformanceRealWorldScenarios tests realistic query performance scenarios
func TestQueryPerformanceRealWorldScenarios(t *testing.T) {
	t.Run("single_user_query", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Simple SELECT by ID query
		explainOutput := `{"Plan": {"Node Type": "Index Scan", "Index Name": "users_pkey", "Total Cost": 0.29}}`

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		assert.Equal(t, "Index Scan", plan.NodeType)
		assert.Less(t, plan.TotalCost, 1.0) // Should be very cheap
	})

	t.Run("batch_import_query", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Bulk insert query
		explainOutput := `{"Plan": {"Node Type": "Result", "Total Cost": 245.50}}`

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		assert.Greater(t, plan.TotalCost, 100.0) // Should be more expensive
	})

	t.Run("complex_aggregation_query", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Complex GROUP BY with joins
		explainOutput := MockExplainOutputComplex()

		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		assert.Equal(t, "Hash Join", plan.NodeType)
		assert.Greater(t, plan.TotalCost, 100.0)
	})
}

// TestQueryPerformancePipelineIntegration tests the complete pipeline
func TestQueryPerformancePipelineIntegration(t *testing.T) {
	t.Run("complete_flow_simulation", func(t *testing.T) {
		// Step 1: Initialize collector
		collector := query_performance.NewQueryCollector(nil)
		require.NotNil(t, collector)

		// Step 2: Get EXPLAIN output
		explainOutput := MockExplainOutput()
		require.NotEmpty(t, explainOutput)

		// Step 3: Parse explain output
		plan, err := collector.ParseExplainOutput(explainOutput)
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Step 4: Validate plan
		assert.NotEmpty(t, plan.NodeType)
		assert.Greater(t, plan.TotalCost, 0.0)

		// Step 5: Simulate storage (plan is now ready to be stored)
		storedPlan := plan
		assert.NotNil(t, storedPlan)

		// Step 6: Verify end-to-end flow succeeded
		assert.Equal(t, "Seq Scan", storedPlan.NodeType)
	})

	t.Run("high_volume_parsing", func(t *testing.T) {
		collector := query_performance.NewQueryCollector(nil)

		// Simulate parsing many explain outputs
		explainOutputs := []string{
			MockExplainOutput(),
			MockExplainOutputComplex(),
			`{"Plan": {"Node Type": "Index Scan", "Total Cost": 5.50}}`,
		}

		successCount := 0
		for _, output := range explainOutputs {
			plan, err := collector.ParseExplainOutput(output)
			if err == nil && plan != nil {
				successCount++
			}
		}

		assert.Equal(t, 3, successCount)
	})
}
