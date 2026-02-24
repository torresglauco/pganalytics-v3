package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
)

// MockMLService provides a mock ML service for testing
type MockMLService struct {
	server           *httptest.Server
	mu               sync.RWMutex
	trainingJobs     map[string]*TrainingJob
	predictions      map[string]*ml.PredictionResponse
	shouldFail       bool
	failureCount     int
	responseDelay    time.Duration
	httpStatusCode   int
}

// TrainingJob represents a mock training job
type TrainingJob struct {
	JobID             string
	Status            string
	ModelID           int64
	RSquared          float64
	TrainingSamples   int
	FeatureCount      int
	CompletedAt       *time.Time
	Error             string
	FeatureImportance map[string]interface{}
}

// NewMockMLService creates a new mock ML service
func NewMockMLService() *MockMLService {
	mock := &MockMLService{
		trainingJobs:   make(map[string]*TrainingJob),
		predictions:    make(map[string]*ml.PredictionResponse),
		httpStatusCode: http.StatusOK,
	}

	mux := http.NewServeMux()

	// Training endpoints
	mux.HandleFunc("/api/health", mock.handleHealth)
	mux.HandleFunc("/api/train/performance-model", mock.handleTrain)
	mux.HandleFunc("/api/train/performance-model/", mock.handleTrainingStatus)

	// Prediction endpoints
	mux.HandleFunc("/api/predict/query-execution", mock.handlePredict)
	mux.HandleFunc("/api/validate/prediction", mock.handleValidate)

	// Pattern detection
	mux.HandleFunc("/api/detect/patterns", mock.handleDetectPatterns)

	mock.server = httptest.NewServer(mux)
	return mock
}

// URL returns the mock server URL
func (m *MockMLService) URL() string {
	return m.server.URL
}

// Close closes the mock server
func (m *MockMLService) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// SetShouldFail makes the mock service fail subsequent requests
func (m *MockMLService) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

// SetHTTPStatusCode sets the HTTP status code for responses
func (m *MockMLService) SetHTTPStatusCode(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpStatusCode = code
}

// SetResponseDelay sets a delay for responses (to test timeouts)
func (m *MockMLService) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// GetTrainingJob retrieves a training job by ID
func (m *MockMLService) GetTrainingJob(jobID string) *TrainingJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trainingJobs[jobID]
}

// GetPrediction retrieves a prediction by query hash
func (m *MockMLService) GetPrediction(queryHash int64) *ml.PredictionResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, pred := range m.predictions {
		if pred.QueryHash == queryHash {
			return pred
		}
	}
	return nil
}

// GetRequestCount returns the number of requests made
func (m *MockMLService) GetRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failureCount
}

// Handler methods

func (m *MockMLService) handleHealth(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	if shouldFail {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Service unavailable"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *MockMLService) handleTrain(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	if shouldFail || statusCode >= 400 {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": "Training failed"})
		return
	}

	var req ml.TrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	jobID := fmt.Sprintf("job-%d", time.Now().Unix())
	job := &TrainingJob{
		JobID:   jobID,
		Status:  "training",
		ModelID: 1,
	}

	m.mu.Lock()
	m.trainingJobs[jobID] = job
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(ml.TrainingResponse{
		JobID:     jobID,
		Status:    "training",
		Message:   "Model training started",
		Timestamp: time.Now().UTC(),
	})
}

func (m *MockMLService) handleTrainingStatus(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	// Extract job ID from URL
	jobID := r.URL.Path[len("/api/train/performance-model/"):]
	if jobID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Job ID required"})
		return
	}

	m.mu.RLock()
	job, exists := m.trainingJobs[jobID]
	m.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Job not found"})
		return
	}

	if shouldFail || statusCode >= 400 {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": "Status check failed"})
		return
	}

	// Update job status (simulate completion)
	m.mu.Lock()
	if job.Status == "training" {
		job.Status = "completed"
		now := time.Now().UTC()
		job.CompletedAt = &now
		job.RSquared = 0.78
		job.TrainingSamples = 1500
		job.FeatureCount = 12
		job.FeatureImportance = map[string]interface{}{
			"mean_execution_time_ms": 0.35,
			"calls_per_minute":       0.28,
			"index_count":            0.18,
			"table_row_count":        0.12,
			"scan_type":              0.07,
		}
	}
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ml.TrainingStatusResponse{
		JobID:             job.JobID,
		Status:            job.Status,
		ModelID:           job.ModelID,
		RSquared:          job.RSquared,
		TrainingSamples:   job.TrainingSamples,
		CompletedAt:       job.CompletedAt,
		Error:             job.Error,
		FeatureCount:      job.FeatureCount,
		FeatureImportance: job.FeatureImportance,
	})
}

func (m *MockMLService) handlePredict(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	var req ml.PredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	if shouldFail || statusCode >= 400 {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": "Prediction failed"})
		return
	}

	// Generate mock prediction
	pred := &ml.PredictionResponse{
		QueryHash:            req.QueryHash,
		PredictedExecutionMs: 125.5,
		ConfidenceScore:      0.87,
		Range: ml.PredictionRange{
			Min: 95.3,
			Max: 155.7,
		},
		ModelVersion: stringPtr("v1.2"),
		Timestamp:    time.Now().UTC(),
	}

	m.mu.Lock()
	m.predictions[fmt.Sprintf("pred-%d", req.QueryHash)] = pred
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pred)
}

func (m *MockMLService) handleValidate(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	var req ml.ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	if shouldFail || statusCode >= 400 {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed"})
		return
	}

	// Calculate error percentage
	errorPct := ((req.PredictedExecutionTimeMs - req.ActualExecutionTimeMs) / req.ActualExecutionTimeMs) * 100
	if errorPct < 0 {
		errorPct = -errorPct
	}

	withinRange := req.ActualExecutionTimeMs >= (req.PredictedExecutionTimeMs*0.8) &&
		req.ActualExecutionTimeMs <= (req.PredictedExecutionTimeMs*1.2)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ml.ValidationResponse{
		PredictionID:          req.PredictionID,
		ErrorPercent:          errorPct,
		AccuracyScore:         1.0 - (errorPct / 100.0),
		WithinConfidenceRange: withinRange,
		Message:               "Prediction validation recorded",
		Timestamp:             time.Now().UTC(),
	})
}

func (m *MockMLService) handleDetectPatterns(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	shouldFail := m.shouldFail
	delay := m.responseDelay
	statusCode := m.httpStatusCode
	m.mu.RUnlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	var req ml.PatternRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	if shouldFail || statusCode >= 400 {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": "Pattern detection failed"})
		return
	}

	patterns := []ml.Pattern{
		{
			Type:        "hourly_peak",
			Description: "Peak load at 8 AM UTC",
			Confidence:  0.92,
			Metadata: map[string]interface{}{
				"peak_hour": 8,
				"variance":  0.15,
			},
		},
		{
			Type:        "daily_cycle",
			Description: "Daily performance variation",
			Confidence:  0.85,
			Metadata: map[string]interface{}{
				"peak_day":     0,
				"variance":     0.20,
				"pattern_type": "regular",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ml.PatternResponse{
		PatternsDetected: len(patterns),
		Patterns:         patterns,
		Timestamp:        time.Now().UTC(),
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
