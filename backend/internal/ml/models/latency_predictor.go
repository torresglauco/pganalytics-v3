package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LatencyPredictor struct {
	mlServiceURL string
	client       *http.Client
}

type PredictionRequest struct {
	Features []float64 `json:"features"`
}

type PredictionResponse struct {
	LatencyMs  float64 `json:"latency_ms"`
	Confidence float64 `json:"confidence"`
}

func NewLatencyPredictor() *LatencyPredictor {
	return &LatencyPredictor{
		mlServiceURL: "http://localhost:5000",
		client:       &http.Client{},
	}
}

func (lp *LatencyPredictor) Predict(features []float64) (float64, error) {
	req := PredictionRequest{Features: features}
	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", lp.mlServiceURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := lp.client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var prediction PredictionResponse
	if err := json.Unmarshal(respBody, &prediction); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return prediction.LatencyMs, nil
}
