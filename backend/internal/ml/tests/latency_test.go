package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml/models"
)

func TestLatencyPrediction(t *testing.T) {
	features := []float64{2, 1, 1000, 2, 0, 0} // [joins, scan_type, rows, filters, subqueries, agg]

	// Mock the ML service
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"latency_ms": 150.5, "confidence": 0.95}`))
	}))
	defer mockServer.Close()

	// Create predictor with mock service URL
	predictor := models.NewLatencyPredictor()
	predictor.SetMLServiceURL(mockServer.URL)

	pred, err := predictor.Predict(features)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if pred <= 0 {
		t.Errorf("Expected positive prediction, got %f", pred)
	}

	if pred != 150.5 {
		t.Errorf("Expected prediction of 150.5, got %f", pred)
	}
}
