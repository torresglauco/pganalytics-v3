package integration

import (
	"context"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"github.com/torresglauco/pganalytics-v3/backend/tests/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestMLClientHealthCheck tests health check functionality
func TestMLClientHealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthy := client.IsHealthy(ctx)
	if !healthy {
		t.Errorf("Expected service to be healthy")
	}
}

// TestMLClientHealthCheckUnhealthy tests health check when service is down
func TestMLClientHealthCheckUnhealthy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	mockService.SetShouldFail(true)
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthy := client.IsHealthy(ctx)
	if healthy {
		t.Errorf("Expected service to be unhealthy")
	}
}

// TestMLClientTrainPerformanceModel tests model training request
func TestMLClientTrainPerformanceModel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &ml.TrainingRequest{
		DatabaseURL:  "postgresql://user:pass@localhost/db",
		LookbackDays: 90,
		ModelType:    "random_forest",
	}

	resp, err := client.TrainPerformanceModel(ctx, req)
	if err != nil {
		t.Fatalf("Expected successful training request, got error: %v", err)
	}

	if resp == nil {
		t.Fatalf("Expected response, got nil")
	}

	if resp.Status != "training" {
		t.Errorf("Expected status 'training', got %s", resp.Status)
	}

	if resp.JobID == "" {
		t.Errorf("Expected job ID in response")
	}
}

// TestMLClientTrainPerformanceModelFailure tests training failure handling
func TestMLClientTrainPerformanceModelFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	mockService.SetHTTPStatusCode(500)
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &ml.TrainingRequest{
		DatabaseURL:  "postgresql://user:pass@localhost/db",
		LookbackDays: 90,
		ModelType:    "random_forest",
	}

	resp, err := client.TrainPerformanceModel(ctx, req)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response on error, got %v", resp)
	}
}

// TestMLClientGetTrainingStatus tests getting training job status
func TestMLClientGetTrainingStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, create a training job
	trainReq := &ml.TrainingRequest{
		DatabaseURL:  "postgresql://user:pass@localhost/db",
		LookbackDays: 90,
		ModelType:    "random_forest",
	}

	trainResp, err := client.TrainPerformanceModel(ctx, trainReq)
	if err != nil {
		t.Fatalf("Failed to create training job: %v", err)
	}

	// Now get the status
	statusResp, err := client.GetTrainingStatus(ctx, trainResp.JobID)
	if err != nil {
		t.Fatalf("Expected successful status request, got error: %v", err)
	}

	if statusResp == nil {
		t.Fatalf("Expected response, got nil")
	}

	if statusResp.JobID != trainResp.JobID {
		t.Errorf("Expected job ID %s, got %s", trainResp.JobID, statusResp.JobID)
	}
}

// TestMLClientPredictQueryExecution tests prediction request
func TestMLClientPredictQueryExecution(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features: map[string]interface{}{
			"mean_execution_time_ms": 125.5,
			"calls_per_minute":       100.0,
		},
	}

	resp, err := client.PredictQueryExecution(ctx, req)
	if err != nil {
		t.Fatalf("Expected successful prediction, got error: %v", err)
	}

	if resp == nil {
		t.Fatalf("Expected response, got nil")
	}

	if resp.QueryHash != 4001 {
		t.Errorf("Expected query hash 4001, got %d", resp.QueryHash)
	}

	if resp.PredictedExecutionMs <= 0 {
		t.Errorf("Expected positive predicted execution time, got %f", resp.PredictedExecutionMs)
	}

	if resp.ConfidenceScore <= 0 || resp.ConfidenceScore > 1 {
		t.Errorf("Expected confidence score between 0 and 1, got %f", resp.ConfidenceScore)
	}
}

// TestMLClientValidatePrediction tests prediction validation
func TestMLClientValidatePrediction(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &ml.ValidationRequest{
		PredictionID:             "pred-001",
		QueryHash:                4001,
		PredictedExecutionTimeMs: 125.5,
		ActualExecutionTimeMs:    118.2,
	}

	resp, err := client.ValidatePrediction(ctx, req)
	if err != nil {
		t.Fatalf("Expected successful validation, got error: %v", err)
	}

	if resp == nil {
		t.Fatalf("Expected response, got nil")
	}

	if resp.PredictionID != "pred-001" {
		t.Errorf("Expected prediction ID 'pred-001', got %s", resp.PredictionID)
	}

	if resp.ErrorPercent < 0 {
		t.Errorf("Expected non-negative error percent, got %f", resp.ErrorPercent)
	}
}

// TestMLClientDetectWorkloadPatterns tests pattern detection
func TestMLClientDetectWorkloadPatterns(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &ml.PatternRequest{
		DatabaseURL:  "postgresql://user:pass@localhost/db",
		LookbackDays: 30,
	}

	resp, err := client.DetectWorkloadPatterns(ctx, req)
	if err != nil {
		t.Fatalf("Expected successful pattern detection, got error: %v", err)
	}

	if resp == nil {
		t.Fatalf("Expected response, got nil")
	}

	if resp.PatternsDetected <= 0 {
		t.Errorf("Expected patterns detected > 0, got %d", resp.PatternsDetected)
	}

	if len(resp.Patterns) != resp.PatternsDetected {
		t.Errorf("Expected %d patterns, got %d", resp.PatternsDetected, len(resp.Patterns))
	}
}

// TestMLClientContextTimeout tests timeout handling
func TestMLClientContextTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	mockService.SetResponseDelay(2 * time.Second)
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 100*time.Millisecond, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	_, err := client.PredictQueryExecution(ctx, req)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

// TestMLClientCircuitBreakerIntegration tests circuit breaker with client
func TestMLClientCircuitBreakerIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check initial state
	state := client.GetCircuitBreakerState()
	if state != "closed" {
		t.Errorf("Expected initial circuit breaker state 'closed', got %s", state)
	}

	// Make successful request
	req := &ml.PredictionRequest{
		QueryHash: 4001,
		Features:  make(map[string]interface{}),
	}

	_, err := client.PredictQueryExecution(ctx, req)
	if err != nil {
		t.Errorf("Expected successful request, got error: %v", err)
	}

	// State should still be closed
	state = client.GetCircuitBreakerState()
	if state != "closed" {
		t.Errorf("Expected circuit breaker state 'closed' after success, got %s", state)
	}
}

// TestMLClientMultipleRequests tests sequential requests
func TestMLClientMultipleRequests(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := mocks.NewMockMLService()
	defer mockService.Close()

	client := ml.NewClient(mockService.URL(), 5*time.Second, logger)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req := &ml.PredictionRequest{
			QueryHash: int64(4000 + i),
			Features:  make(map[string]interface{}),
		}

		resp, err := client.PredictQueryExecution(ctx, req)
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
			continue
		}

		if resp.QueryHash != int64(4000+i) {
			t.Errorf("Request %d: expected query hash %d, got %d", i, 4000+i, resp.QueryHash)
		}
	}
}
