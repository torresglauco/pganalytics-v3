package handlers

import (
	"database/sql"
	"encoding/json"
)

type HandlerContext struct {
	DB *sql.DB
}

type TableStats struct {
	TableName         string  `json:"table_name"`
	RowCount          int64   `json:"row_count"`
	SizeBytes         int64   `json:"size_bytes"`
	IndexCount        int     `json:"index_count"`
	LastAutovacuum    string  `json:"last_autovacuum"`
	DeadRowsPercent   float64 `json:"dead_rows_percent"`
	TableBloatPercent float64 `json:"table_bloat_percent"`
}

type QueryAnalysisResult struct {
	QueryID         string   `json:"query_id"`
	QueryText       string   `json:"query_text"`
	ExecutionCount  int64    `json:"execution_count"`
	MeanTimeMs      float64  `json:"mean_time_ms"`
	MaxTimeMs       float64  `json:"max_time_ms"`
	TotalTimeMs     float64  `json:"total_time_ms"`
	Anomalies       []string `json:"anomalies"`
	Recommendations []string `json:"recommendations"`
}

type IndexSuggestion struct {
	TableName     string   `json:"table_name"`
	Columns       []string `json:"columns"`
	EstimatedGain float64  `json:"estimated_gain_percent"`
	Reason        string   `json:"reason"`
	Priority      string   `json:"priority"` // high, medium, low
}

func NewHandlerContext(db *sql.DB) *HandlerContext {
	return &HandlerContext{DB: db}
}

func MarshalSuggestion(s *IndexSuggestion) ([]byte, error) {
	return json.Marshal(s)
}
