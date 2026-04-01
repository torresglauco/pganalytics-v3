package mcp_test

import (
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/handlers"
)

func TestTableStatsHandler(t *testing.T) {
	ctx := handlers.NewHandlerContext(nil)
	if ctx == nil {
		t.Fatal("NewHandlerContext returned nil")
	}
}

func TestQueryAnalysisHandler(t *testing.T) {
	stats := &handlers.QueryAnalysisResult{
		QueryID:        "test_query_1",
		ExecutionCount: 100,
	}
	if stats.QueryID != "test_query_1" {
		t.Fatal("QueryAnalysisResult not initialized correctly")
	}
}

func TestIndexSuggestion(t *testing.T) {
	suggestion := &handlers.IndexSuggestion{
		TableName:     "users",
		Columns:       []string{"email"},
		EstimatedGain: 45.5,
		Priority:      "high",
	}
	data, _ := handlers.MarshalSuggestion(suggestion)
	if len(data) == 0 {
		t.Fatal("MarshalSuggestion returned empty data")
	}
}
