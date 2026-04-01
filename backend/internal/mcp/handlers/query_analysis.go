package handlers

import "fmt"

func (ctx *HandlerContext) QueryAnalysis(params map[string]interface{}) (interface{}, error) {
	queryID, ok := params["query_id"].(string)
	if !ok || queryID == "" {
		return nil, fmt.Errorf("query_id parameter required")
	}

	if ctx.DB == nil {
		// Return mock data for testing
		return QueryAnalysisResult{
			QueryID:         queryID,
			QueryText:       "SELECT * FROM users WHERE id = $1",
			ExecutionCount:  1500,
			MeanTimeMs:      45.2,
			MaxTimeMs:       250.5,
			TotalTimeMs:     67800,
			Anomalies:       []string{"high variance in execution time"},
			Recommendations: []string{"add index on id column", "consider query rewrite"},
		}, nil
	}

	// Real implementation would query database
	query := `
		SELECT
			query_id,
			query_text,
			execution_count,
			mean_time_ms,
			max_time_ms,
			total_time_ms
		FROM pg_analytics.query_stats
		WHERE query_id = $1
	`

	var result QueryAnalysisResult
	err := ctx.DB.QueryRow(query, queryID).Scan(
		&result.QueryID,
		&result.QueryText,
		&result.ExecutionCount,
		&result.MeanTimeMs,
		&result.MaxTimeMs,
		&result.TotalTimeMs,
	)

	if err != nil {
		return nil, err
	}

	// Initialize slices to prevent nil pointer dereferences
	if result.Anomalies == nil {
		result.Anomalies = make([]string, 0)
	}
	if result.Recommendations == nil {
		result.Recommendations = make([]string, 0)
	}

	// Add anomaly detection logic with bounds checking
	if result.MeanTimeMs > 0 && result.MaxTimeMs > result.MeanTimeMs*5 {
		result.Anomalies = append(result.Anomalies, "high variance in execution time")
		result.Recommendations = append(result.Recommendations, "investigate query performance spikes")
	}

	return result, nil
}
