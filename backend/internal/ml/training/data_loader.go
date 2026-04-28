package training

import (
	"database/sql"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml/models"
)

type DataLoader struct {
	db    *sql.DB
	limit int
}

type TrainingDataset struct {
	Features [][]float64
	Labels   []float64
}

func NewDataLoader(dbURL string, limit int) *DataLoader {
	return &DataLoader{
		limit: limit,
	}
}

func (dl *DataLoader) LoadQueryMetrics() (*TrainingDataset, error) {
	// Query from pg_analytics.query_execution_plans
	// query := `
	// 	SELECT
	// 		jsonb_extract_path_text(plan, 'Join Count')::int,
	// 		COALESCE(jsonb_extract_path_text(plan, 'Node Type'), 'unknown'),
	// 		rows_affected,
	// 		COALESCE(filters_applied, 0),
	// 		execution_time_ms
	// 	FROM pg_analytics.query_execution_plans
	// 	WHERE execution_time_ms > 10
	// 	LIMIT $1
	// `

	dataset := &TrainingDataset{
		Features: make([][]float64, 0),
		Labels:   make([]float64, 0),
	}

	// In real implementation, execute query and build dataset
	// For now, return empty dataset structure

	return dataset, nil
}

func (dl *DataLoader) LoadAnomalyTrainingData() (*TrainingDataset, error) {
	// Load baseline metrics for anomaly detection
	dataset := &TrainingDataset{
		Features: make([][]float64, 0),
		Labels:   make([]float64, 0),
	}

	return dataset, nil
}

func FingerprintQuery(features map[string]float64) string {
	// Simple fingerprint from features
	qf := &models.QueryFeatures{
		JoinCount: int(features["join_count"]),
		ScanType:  "seq_scan",
		RowCount:  int(features["row_count"]),
	}
	return qf.Fingerprint()
}
