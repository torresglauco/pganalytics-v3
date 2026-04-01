package tests

import (
	"testing"
	"github.com/torresglauco/pganalytics-v3/backend/internal/ml/models"
)

func TestLatencyPrediction(t *testing.T) {
	features := []float64{2, 1, 1000, 2, 0, 0}  // [joins, scan_type, rows, filters, subqueries, agg]

	predictor := models.NewLatencyPredictor()
	pred, err := predictor.Predict(features)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if pred <= 0 {
		t.Errorf("Expected positive prediction, got %f", pred)
	}
}
