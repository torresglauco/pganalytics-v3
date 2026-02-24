package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Client represents an HTTP client for the ML service
type Client struct {
	baseURL        string
	timeout        time.Duration
	circuitBreaker *CircuitBreaker
	httpClient     *http.Client
	logger         *zap.Logger
	mu             sync.RWMutex
}

// TrainingRequest represents a request to start model training
type TrainingRequest struct {
	DatabaseURL  string `json:"database_url"`
	LookbackDays int    `json:"lookback_days"`
	ModelType    string `json:"model_type"`
	JobID        string `json:"job_id,omitempty"`
	ForceRetrain bool   `json:"force_retrain,omitempty"`
}

// TrainingResponse represents a response from training
type TrainingResponse struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// TrainingStatusResponse represents the status of a training job
type TrainingStatusResponse struct {
	JobID             string                 `json:"job_id"`
	Status            string                 `json:"status"`
	ModelID           int64                  `json:"model_id,omitempty"`
	RSquared          float64                `json:"r_squared,omitempty"`
	TrainingSamples   int                    `json:"training_samples,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	Error             string                 `json:"error,omitempty"`
	FeatureCount      int                    `json:"feature_count,omitempty"`
	FeatureImportance map[string]interface{} `json:"feature_importance,omitempty"`
}

// PredictionRequest represents a request to make a prediction
type PredictionRequest struct {
	QueryHash  int64                  `json:"query_hash"`
	Features   map[string]interface{} `json:"features,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Scenario   string                 `json:"scenario,omitempty"`
	ModelID    *int64                 `json:"model_id,omitempty"`
}

// PredictionResponse represents a prediction response
type PredictionResponse struct {
	QueryHash              int64                  `json:"query_hash"`
	PredictedExecutionMs   float64                `json:"predicted_execution_time_ms"`
	ConfidenceScore        float64                `json:"confidence"`
	Range                  PredictionRange        `json:"range"`
	ModelVersion           *string                `json:"model_version,omitempty"`
	Features               map[string]interface{} `json:"features,omitempty"`
	Timestamp              time.Time              `json:"timestamp"`
	RecommendedIndexes     []string               `json:"recommended_indexes,omitempty"`
	OptimizationSuggestion string                 `json:"optimization_suggestion,omitempty"`
}

// PredictionRange represents the confidence range for a prediction
type PredictionRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ValidationRequest represents a request to validate prediction accuracy
type ValidationRequest struct {
	PredictionID             string  `json:"prediction_id"`
	QueryHash                int64   `json:"query_hash"`
	PredictedExecutionTimeMs float64 `json:"predicted_execution_time_ms"`
	ActualExecutionTimeMs    float64 `json:"actual_execution_time_ms"`
	ModelVersion             *string `json:"model_version,omitempty"`
}

// ValidationResponse represents the result of prediction validation
type ValidationResponse struct {
	PredictionID          string    `json:"prediction_id"`
	ErrorPercent          float64   `json:"error_percent"`
	AccuracyScore         float64   `json:"accuracy_score"`
	WithinConfidenceRange bool      `json:"within_confidence_interval"`
	Message               string    `json:"message"`
	Timestamp             time.Time `json:"timestamp"`
}

// PatternRequest represents a request to detect workload patterns
type PatternRequest struct {
	DatabaseURL  string `json:"database_url"`
	LookbackDays int    `json:"lookback_days"`
}

// PatternResponse represents workload pattern detection response
type PatternResponse struct {
	PatternsDetected int       `json:"patterns_detected"`
	Patterns         []Pattern `json:"patterns,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// Pattern represents a detected workload pattern
type Pattern struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ErrorResponse represents an error response from ML service
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
}

// NewClient creates a new ML service client with default HTTP configuration
func NewClient(baseURL string, timeout time.Duration, logger *zap.Logger) *Client {
	return &Client{
		baseURL:        baseURL,
		timeout:        timeout,
		circuitBreaker: NewCircuitBreaker(logger),
		httpClient:     newHTTPClientWithPooling(timeout),
		logger:         logger,
	}
}

// newHTTPClientWithPooling creates an HTTP client with connection pooling optimized for performance
func newHTTPClientWithPooling(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			MaxConnsPerHost:     10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// TrainPerformanceModel starts async model training
func (c *Client) TrainPerformanceModel(ctx context.Context, req *TrainingRequest) (*TrainingResponse, error) {
	if !c.circuitBreaker.IsOpen() {
		c.logger.Debug("Circuit breaker is open for ML service")
		return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/train/performance-model", body)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.circuitBreaker.RecordFailure()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	c.circuitBreaker.RecordSuccess()

	var trainingResp TrainingResponse
	if err := json.NewDecoder(resp.Body).Decode(&trainingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &trainingResp, nil
}

// GetTrainingStatus gets the status of a training job
func (c *Client) GetTrainingStatus(ctx context.Context, jobID string) (*TrainingStatusResponse, error) {
	if !c.circuitBreaker.IsOpen() {
		return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/train/performance-model/%s", jobID), nil)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.circuitBreaker.RecordFailure()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	c.circuitBreaker.RecordSuccess()

	var statusResp TrainingStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &statusResp, nil
}

// PredictQueryExecution makes a prediction for query execution time
func (c *Client) PredictQueryExecution(ctx context.Context, req *PredictionRequest) (*PredictionResponse, error) {
	if !c.circuitBreaker.IsOpen() {
		return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/predict/query-execution", body)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.circuitBreaker.RecordFailure()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	c.circuitBreaker.RecordSuccess()

	var predResp PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &predResp, nil
}

// ValidatePrediction records actual query result and validates prediction accuracy
func (c *Client) ValidatePrediction(ctx context.Context, req *ValidationRequest) (*ValidationResponse, error) {
	if !c.circuitBreaker.IsOpen() {
		return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/validate/prediction", body)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.circuitBreaker.RecordFailure()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	c.circuitBreaker.RecordSuccess()

	var valResp ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&valResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &valResp, nil
}

// DetectWorkloadPatterns triggers workload pattern detection
func (c *Client) DetectWorkloadPatterns(ctx context.Context, req *PatternRequest) (*PatternResponse, error) {
	if !c.circuitBreaker.IsOpen() {
		return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/api/detect/patterns", body)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.circuitBreaker.RecordFailure()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	c.circuitBreaker.RecordSuccess()

	var patternResp PatternResponse
	if err := json.NewDecoder(resp.Body).Decode(&patternResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &patternResp, nil
}

// IsHealthy checks if the ML service is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	resp, err := c.doRequest(ctx, "GET", "/api/health", nil)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return false
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == http.StatusOK
	if healthy {
		c.circuitBreaker.RecordSuccess()
	} else {
		c.circuitBreaker.RecordFailure()
	}

	return healthy
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (c *Client) GetCircuitBreakerState() string {
	return c.circuitBreaker.State()
}

// doRequest performs an HTTP request to the ML service with exponential backoff retry
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	return c.doRequestWithRetry(ctx, method, path, body, 3)
}

// doRequestWithRetry performs an HTTP request with exponential backoff retry mechanism
// Only retries on transient failures (5xx, timeout, connection errors)
func (c *Client) doRequestWithRetry(
	ctx context.Context,
	method, path string,
	body []byte,
	maxRetries int,
) (*http.Response, error) {
	url := c.baseURL + path
	backoff := 100 * time.Millisecond // Initial backoff
	multiplier := 2.0                 // Exponential multiplier

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Create request
		var req *http.Request
		var err error

		if body != nil {
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequestWithContext(ctx, method, url, nil)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.logger.Debug("Calling ML service",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("attempt", attempt+1),
		)

		// Perform request
		resp, err := c.httpClient.Do(req)

		// Check if error is retriable
		if err != nil {
			if attempt < maxRetries && isRetriableError(err) {
				c.logger.Warn("Retriable ML service error, retrying",
					zap.Error(err),
					zap.String("path", path),
					zap.Int("attempt", attempt+1),
					zap.Duration("backoff", backoff),
				)

				select {
				case <-time.After(backoff):
					backoff = time.Duration(float64(backoff) * multiplier)
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			c.logger.Warn("ML service request failed",
				zap.Error(err),
				zap.String("path", path),
				zap.Int("attempt", attempt+1),
			)
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Check for transient HTTP errors (5xx)
		if resp.StatusCode >= 500 && attempt < maxRetries {
			resp.Body.Close()
			c.logger.Warn("ML service returned 5xx error, retrying",
				zap.Int("status_code", resp.StatusCode),
				zap.String("path", path),
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff),
			)

			select {
			case <-time.After(backoff):
				backoff = time.Duration(float64(backoff) * multiplier)
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded for %s", url)
}

// isRetriableError checks if an error is retriable (transient failure)
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}

	// Check for network errors
	errStr := err.Error()
	if errStr == "EOF" ||
		errStr == "connection reset by peer" ||
		errStr == "broken pipe" ||
		errStr == "i/o timeout" {
		return true
	}

	return false
}

// Close closes the client (cleanup)
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
