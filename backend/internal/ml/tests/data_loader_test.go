package tests

import (
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml/training"
)

func TestLoadQueryData(t *testing.T) {
	// Mock data loading
	loader := training.NewDataLoader(":memory:", 1000)

	if loader == nil {
		t.Fatal("Expected loader to be created")
	}
}

func TestFeatureEngineering(t *testing.T) {
	features := map[string]float64{
		"join_count": 2,
		"scan_type":  1, // sequential=0, index=1, bitmap=2
		"row_count":  1000,
	}

	fingerprint := training.FingerprintQuery(features)

	if fingerprint == "" {
		t.Fatal("Expected fingerprint to be generated")
	}
}
