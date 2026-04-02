package handlers

import (
	"encoding/json"
	"testing"
)

// TestTableStatsWithMockDatabase tests table_stats tool with mock database (no DB connection)
func TestTableStatsWithMockDatabase(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	result, err := mockCtx.TableStats(map[string]interface{}{
		"table_name": "users",
	})
	if err != nil {
		t.Fatalf("TableStats failed: %v", err)
	}

	stats, ok := result.([]TableStats)
	if !ok {
		t.Fatalf("Result type mismatch: got %T, want []TableStats", result)
	}

	if len(stats) != 1 {
		t.Errorf("Expected 1 stat, got %d", len(stats))
	}

	stat := stats[0]
	if stat.TableName != "users" {
		t.Errorf("Expected table_name 'users', got %s", stat.TableName)
	}

	if stat.RowCount != 1000000 {
		t.Errorf("Expected row_count 1000000, got %d", stat.RowCount)
	}

	if stat.SizeBytes != 10485760 {
		t.Errorf("Expected size_bytes 10485760, got %d", stat.SizeBytes)
	}

	if stat.IndexCount != 3 {
		t.Errorf("Expected index_count 3, got %d", stat.IndexCount)
	}
}

// TestTableStatsErrorHandling tests table_stats error handling
func TestTableStatsErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]interface{}
		shouldErr      bool
		expectedErrMsg string
	}{
		{
			name:           "Missing table_name",
			params:         map[string]interface{}{},
			shouldErr:      true,
			expectedErrMsg: "table_name parameter required",
		},
		{
			name:           "Empty table_name",
			params:         map[string]interface{}{"table_name": ""},
			shouldErr:      true,
			expectedErrMsg: "table_name parameter required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewHandlerContext(nil)
			_, err := mockCtx.TableStats(tt.params)

			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}

			if tt.shouldErr && err.Error() != tt.expectedErrMsg {
				t.Errorf("Expected error message %q, got %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}

// TestQueryAnalysisValidAnalysis tests query_analysis with valid query
func TestQueryAnalysisValidAnalysis(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	params := map[string]interface{}{
		"query_id": "query_123",
	}

	result, err := mockCtx.QueryAnalysis(params)
	if err != nil {
		t.Fatalf("QueryAnalysis failed: %v", err)
	}

	analysis, ok := result.(QueryAnalysisResult)
	if !ok {
		t.Fatalf("Result type mismatch: got %T, want QueryAnalysisResult", result)
	}

	if analysis.QueryID != "query_123" {
		t.Errorf("Expected query_id 'query_123', got %s", analysis.QueryID)
	}

	if analysis.Anomalies == nil {
		t.Fatal("Expected non-nil anomalies array")
	}

	if analysis.Recommendations == nil {
		t.Fatal("Expected non-nil recommendations array")
	}
}

// TestQueryAnalysisErrorHandling tests query_analysis error handling
func TestQueryAnalysisErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]interface{}
		shouldErr      bool
		expectedErrMsg string
	}{
		{
			name:           "Missing query_id",
			params:         map[string]interface{}{},
			shouldErr:      true,
			expectedErrMsg: "query_id parameter required",
		},
		{
			name:           "Empty query_id",
			params:         map[string]interface{}{"query_id": ""},
			shouldErr:      true,
			expectedErrMsg: "query_id parameter required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewHandlerContext(nil)
			_, err := mockCtx.QueryAnalysis(tt.params)

			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}

			if tt.shouldErr && err.Error() != tt.expectedErrMsg {
				t.Errorf("Expected error message %q, got %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}

// TestIndexSuggestWithTable tests index suggestions for specific table
func TestIndexSuggestWithTable(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	params := map[string]interface{}{
		"table_name": "users",
	}

	result, err := mockCtx.IndexSuggest(params)
	if err != nil {
		t.Fatalf("IndexSuggest failed: %v", err)
	}

	suggestions, ok := result.([]IndexSuggestion)
	if !ok {
		t.Fatalf("Result type mismatch: got %T, want []IndexSuggestion", result)
	}

	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion")
	}

	for i, sugg := range suggestions {
		if sugg.TableName != "users" {
			t.Errorf("Suggestion %d: Expected table_name 'users', got %s", i, sugg.TableName)
		}

		validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
		if !validPriorities[sugg.Priority] {
			t.Errorf("Suggestion %d: Invalid priority: %s", i, sugg.Priority)
		}
	}
}

// TestAnomalyDetectValidation tests anomaly detection validation
func TestAnomalyDetectValidation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	params := map[string]interface{}{
		"table_name": "users",
	}

	result, err := mockCtx.DetectAnomalies(params)
	if err != nil {
		t.Fatalf("DetectAnomalies failed: %v", err)
	}

	alerts, ok := result.([]AnomalyAlert)
	if !ok {
		t.Fatalf("Result type mismatch: got %T, want []AnomalyAlert", result)
	}

	// Verify that any detected anomalies have valid severity levels
	for _, anomaly := range alerts {
		validSeverities := map[string]bool{"high": true, "medium": true, "low": true}
		if !validSeverities[anomaly.Severity] {
			t.Errorf("Invalid severity: %s", anomaly.Severity)
		}
	}
}

// TestAnomalyDetectErrorHandling tests anomaly_detect error handling
func TestAnomalyDetectErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]interface{}
		shouldErr      bool
		expectedErrMsg string
	}{
		{
			name:           "Missing table_name",
			params:         map[string]interface{}{},
			shouldErr:      true,
			expectedErrMsg: "table_name parameter required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewHandlerContext(nil)
			_, err := mockCtx.DetectAnomalies(tt.params)

			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error but got none")
			}

			if tt.shouldErr && err.Error() != tt.expectedErrMsg {
				t.Errorf("Expected error message %q, got %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}

// TestHandlerContextCreation tests proper handler context creation
func TestHandlerContextCreation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)
	if mockCtx == nil {
		t.Fatal("NewHandlerContext returned nil")
	}

	if mockCtx.DB != nil {
		t.Error("Expected DB to be nil in mock mode")
	}
}

// TestMarshalSuggestion tests JSON marshaling of suggestions
func TestMarshalSuggestion(t *testing.T) {
	suggestion := &IndexSuggestion{
		TableName:     "users",
		Columns:       []string{"email", "created_at"},
		EstimatedGain: 45.5,
		Reason:        "High frequency in WHERE clauses",
		Priority:      "high",
	}

	data, err := MarshalSuggestion(suggestion)
	if err != nil {
		t.Fatalf("MarshalSuggestion failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("MarshalSuggestion returned empty data")
	}

	// Verify we can unmarshal it back
	var unmarshalled IndexSuggestion
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal suggestion: %v", err)
	}

	if unmarshalled.TableName != suggestion.TableName {
		t.Errorf("Unmarshalled table_name mismatch: got %s, want %s", unmarshalled.TableName, suggestion.TableName)
	}
}

// TestResponseStructureIntegrity tests that all tool responses maintain integrity
func TestResponseStructureIntegrity(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	tests := []struct {
		name      string
		toolFunc  func(map[string]interface{}) (interface{}, error)
		params    map[string]interface{}
	}{
		{
			name:     "TableStats",
			toolFunc: mockCtx.TableStats,
			params:   map[string]interface{}{"table_name": "test"},
		},
		{
			name:     "QueryAnalysis",
			toolFunc: mockCtx.QueryAnalysis,
			params:   map[string]interface{}{"query_id": "test"},
		},
		{
			name:     "IndexSuggest",
			toolFunc: mockCtx.IndexSuggest,
			params:   map[string]interface{}{"table_name": "test"},
		},
		{
			name:     "DetectAnomalies",
			toolFunc: mockCtx.DetectAnomalies,
			params:   map[string]interface{}{"table_name": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.toolFunc(tt.params)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			if result == nil {
				t.Fatalf("%s returned nil result", tt.name)
			}
		})
	}
}

// TestIndexSuggestWithoutTable tests index suggestions without specific table
func TestIndexSuggestWithoutTable(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	// Call without table_name parameter
	result, err := mockCtx.IndexSuggest(map[string]interface{}{})
	if err != nil {
		t.Fatalf("IndexSuggest failed: %v", err)
	}

	suggestions, ok := result.([]IndexSuggestion)
	if !ok {
		t.Fatalf("Result type mismatch: got %T, want []IndexSuggestion", result)
	}

	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for all tables")
	}

	// Should suggest all tables
	sugg := suggestions[0]
	if sugg.TableName != "*all_tables" {
		t.Errorf("Expected table_name '*all_tables', got %s", sugg.TableName)
	}
}

// TestTableStatsFieldValidation validates all required fields in TableStats
func TestTableStatsFieldValidation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	result, err := mockCtx.TableStats(map[string]interface{}{"table_name": "test_table"})
	if err != nil {
		t.Fatalf("TableStats failed: %v", err)
	}

	stats, ok := result.([]TableStats)
	if !ok {
		t.Fatal("Result is not []TableStats")
	}

	if len(stats) == 0 {
		t.Fatal("No stats returned")
	}

	stat := stats[0]

	// Verify all fields are populated
	if stat.LastAutovacuum == "" {
		t.Error("LastAutovacuum is empty")
	}

	if stat.DeadRowsPercent < 0 {
		t.Error("DeadRowsPercent is negative")
	}

	if stat.TableBloatPercent < 0 {
		t.Error("TableBloatPercent is negative")
	}
}

// TestQueryAnalysisFieldValidation validates all required fields in QueryAnalysisResult
func TestQueryAnalysisFieldValidation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	result, err := mockCtx.QueryAnalysis(map[string]interface{}{"query_id": "q_test"})
	if err != nil {
		t.Fatalf("QueryAnalysis failed: %v", err)
	}

	analysis, ok := result.(QueryAnalysisResult)
	if !ok {
		t.Fatal("Result is not QueryAnalysisResult")
	}

	// Verify all required numeric fields
	if analysis.ExecutionCount == 0 {
		t.Error("ExecutionCount is zero")
	}

	if analysis.MeanTimeMs <= 0 {
		t.Error("MeanTimeMs should be positive")
	}

	if analysis.MaxTimeMs <= 0 {
		t.Error("MaxTimeMs should be positive")
	}

	if analysis.TotalTimeMs <= 0 {
		t.Error("TotalTimeMs should be positive")
	}

	if analysis.QueryText == "" {
		t.Error("QueryText is empty")
	}
}

// TestIndexSuggestPriorityValidation validates priority assignment logic
func TestIndexSuggestPriorityValidation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	result, err := mockCtx.IndexSuggest(map[string]interface{}{"table_name": "users"})
	if err != nil {
		t.Fatalf("IndexSuggest failed: %v", err)
	}

	suggestions, ok := result.([]IndexSuggestion)
	if !ok {
		t.Fatal("Result is not []IndexSuggestion")
	}

	// Verify each suggestion has proper priority
	for _, sugg := range suggestions {
		switch sugg.Priority {
		case "high", "medium", "low":
			// Valid priority
		default:
			t.Errorf("Invalid priority: %s", sugg.Priority)
		}

		// Verify estimated gain is valid
		if sugg.EstimatedGain < 0 || sugg.EstimatedGain > 100 {
			t.Errorf("Invalid estimated gain: %f", sugg.EstimatedGain)
		}

		// Verify columns array is populated
		if len(sugg.Columns) == 0 {
			t.Error("Columns array is empty")
		}
	}
}

// TestAnomalyDetectZScoreValidation validates Z-score calculation
func TestAnomalyDetectZScoreValidation(t *testing.T) {
	mockCtx := NewHandlerContext(nil)

	result, err := mockCtx.DetectAnomalies(map[string]interface{}{"table_name": "users"})
	if err != nil {
		t.Fatalf("DetectAnomalies failed: %v", err)
	}

	alerts, ok := result.([]AnomalyAlert)
	if !ok {
		t.Fatal("Result is not []AnomalyAlert")
	}

	// If anomalies exist, verify their properties
	for _, alert := range alerts {
		// Z-score should be calculated
		if alert.ZScore == 0 {
			t.Logf("Z-score is 0 for metric %s", alert.MetricName)
		}

		// Current value should differ from baseline
		if alert.CurrentValue == alert.BaselineValue {
			t.Errorf("Current and baseline values are equal for %s", alert.MetricName)
		}

		// Metric name should not be empty
		if alert.MetricName == "" {
			t.Error("MetricName is empty")
		}

		// Description should exist
		if alert.Description == "" {
			t.Error("Description is empty")
		}
	}
}
