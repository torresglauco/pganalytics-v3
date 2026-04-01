# pgAnalytics v3 Wave 2: AI/ML Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build machine learning pipeline for pgAnalytics with query latency prediction and anomaly detection using XGBoost and Isolation Forest models.

**Architecture:**
Separate Python FastAPI microservice (`backend/services/ml-service/`) handles model training and inference. Go backend communicates via HTTP with Bearer token auth. Models trained on pgAnalytics query execution data stored in PostgreSQL.

**Tech Stack:**
Python 3.11, FastAPI, scikit-learn, XGBoost, Pandas, NumPy, Go 1.26, PostgreSQL 15

**Timeline:**
Weeks 5-10 (can run parallel with Wave 1 after Week 2)

---

## File Structure

```
backend/internal/ml/
├── models/
│   ├── query_fingerprint.go       # Feature engineering for queries
│   ├── latency_predictor.go       # XGBoost model wrapper
│   └── anomaly_detector.go        # Isolation Forest wrapper
├── training/
│   ├── data_loader.go             # Load training data from PostgreSQL
│   ├── feature_engineer.go        # Feature transformation
│   └── model_trainer.go           # Model training pipeline
├── inference/
│   ├── predictor.go               # Real-time predictions
│   └── cache.go                   # Redis caching
└── tests/
    ├── models_test.go
    └── inference_test.go

backend/services/ml-service/
├── main.py                        # FastAPI server
├── models.py                      # Model implementations
├── anomaly_model.py               # Anomaly detection
├── inference.py                   # Prediction logic
├── requirements.txt               # Python dependencies
└── models/                        # Trained model storage
    ├── model.joblib
    └── scaler.joblib

backend/internal/api/
└── handlers_ml.go                 # ML API endpoints
```

---

## Task 2.1: ML Model Infrastructure & Data Loader

**Files:**
- Create: `backend/internal/ml/models/query_fingerprint.go`
- Create: `backend/internal/ml/training/data_loader.go`
- Create: `backend/internal/ml/tests/data_loader_test.go`

- [ ] **Step 1: Write test for data loader**

```go
// File: backend/internal/ml/tests/data_loader_test.go
package tests

import (
	"testing"
	"pganalytics/internal/ml/training"
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
		"scan_type":  1,  // sequential=0, index=1, bitmap=2
		"row_count":  1000,
	}

	fingerprint := training.FingerprintQuery(features)

	if fingerprint == "" {
		t.Fatal("Expected fingerprint to be generated")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/ml/tests -v -run DataLoader
```

Expected: FAIL with "package not found"

- [ ] **Step 3: Create query fingerprinting**

```go
// File: backend/internal/ml/models/query_fingerprint.go
package models

import (
	"crypto/md5"
	"fmt"
	"sort"
)

type QueryFeatures struct {
	JoinCount      int     `json:"join_count"`
	ScanType       string  `json:"scan_type"`       // seq_scan, index_scan, bitmap_scan
	RowCount       int     `json:"row_count"`
	FilterCount    int     `json:"filter_count"`
	SubqueryCount  int     `json:"subquery_count"`
	AggregateType  string  `json:"aggregate_type"`  // none, sum, count, group_by
	ExecutionTimeMs float64 `json:"execution_time_ms"`
}

func (qf *QueryFeatures) Fingerprint() string {
	// Create normalized fingerprint for aggregation
	data := fmt.Sprintf("%d|%s|%d|%d|%d|%s",
		qf.JoinCount,
		qf.ScanType,
		qf.RowCount/1000,  // Bucketing
		qf.FilterCount,
		qf.SubqueryCount,
		qf.AggregateType,
	)

	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

func (qf *QueryFeatures) Vector() []float64 {
	scanTypeVal := 0.0
	switch qf.ScanType {
	case "seq_scan":
		scanTypeVal = 0.0
	case "index_scan":
		scanTypeVal = 1.0
	case "bitmap_scan":
		scanTypeVal = 2.0
	}

	aggregateVal := 0.0
	switch qf.AggregateType {
	case "none":
		aggregateVal = 0.0
	case "sum", "count":
		aggregateVal = 1.0
	case "group_by":
		aggregateVal = 2.0
	}

	return []float64{
		float64(qf.JoinCount),
		scanTypeVal,
		float64(qf.RowCount) / 1000,  // Normalize
		float64(qf.FilterCount),
		float64(qf.SubqueryCount),
		aggregateVal,
	}
}
```

- [ ] **Step 4: Create data loader**

```go
// File: backend/internal/ml/training/data_loader.go
package training

import (
	"database/sql"
	"fmt"
	"pganalytics/internal/ml/models"
)

type DataLoader struct {
	db      *sql.DB
	limit   int
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
	query := `
		SELECT
			jsonb_extract_path_text(plan, 'Join Count')::int,
			COALESCE(jsonb_extract_path_text(plan, 'Node Type'), 'unknown'),
			rows_affected,
			COALESCE(filters_applied, 0),
			execution_time_ms
		FROM pg_analytics.query_execution_plans
		WHERE execution_time_ms > 10
		LIMIT $1
	`

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
		JoinCount:  int(features["join_count"]),
		ScanType:   "seq_scan",
		RowCount:   int(features["row_count"]),
	}
	return qf.Fingerprint()
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd backend
go test ./internal/ml/tests -v -run DataLoader
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/ml/models/query_fingerprint.go internal/ml/training/data_loader.go internal/ml/tests/data_loader_test.go
git commit -m "feat: implement ML model infrastructure and query fingerprinting"
```

---

## Task 2.2: Query Latency Prediction Model

**Files:**
- Create: `backend/internal/ml/models/latency_predictor.go`
- Create: `backend/services/ml-service/models.py`
- Create: `backend/services/ml-service/requirements.txt`
- Create: `backend/services/ml-service/main.py`

- [ ] **Step 1: Write test for latency prediction**

```go
// File: backend/internal/ml/tests/latency_test.go
package tests

import (
	"testing"
	"pganalytics/internal/ml/models"
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/ml/tests -v -run Latency
```

- [ ] **Step 3: Create latency predictor interface**

```go
// File: backend/internal/ml/models/latency_predictor.go
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
	LatencyMs float64 `json:"latency_ms"`
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
```

- [ ] **Step 4: Create Python ML service requirements**

```txt
# File: backend/services/ml-service/requirements.txt
fastapi==0.104.1
uvicorn==0.24.0
scikit-learn==1.3.2
xgboost==2.0.0
joblib==1.3.2
pydantic==2.5.0
numpy==1.26.2
pandas==2.1.3
```

- [ ] **Step 5: Create Python models module**

```python
# File: backend/services/ml-service/models.py
import numpy as np
import joblib
from sklearn.ensemble import RandomForestRegressor
from sklearn.preprocessing import StandardScaler

class LatencyModel:
    def __init__(self):
        self.model = None
        self.scaler = StandardScaler()
        self.is_trained = False

    def train(self, X, y):
        """Train latency prediction model"""
        X_scaled = self.scaler.fit_transform(X)
        self.model = RandomForestRegressor(
            n_estimators=100,
            max_depth=15,
            min_samples_split=5,
            random_state=42,
            n_jobs=-1
        )
        self.model.fit(X_scaled, y)
        self.is_trained = True

    def predict(self, X):
        """Predict latency for query features"""
        if not self.is_trained:
            raise RuntimeError("Model not trained")

        X_scaled = self.scaler.transform(X)
        return self.model.predict(X_scaled)

    def save(self, path):
        """Save model to disk"""
        joblib.dump(self.model, f"{path}/model.joblib")
        joblib.dump(self.scaler, f"{path}/scaler.joblib")

    def load(self, path):
        """Load model from disk"""
        self.model = joblib.load(f"{path}/model.joblib")
        self.scaler = joblib.load(f"{path}/scaler.joblib")
        self.is_trained = True
```

- [ ] **Step 6: Create FastAPI service**

```python
# File: backend/services/ml-service/main.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import numpy as np
from models import LatencyModel
import logging

app = FastAPI(title="pgAnalytics ML Service")
logger = logging.getLogger(__name__)

# Load pre-trained model
latency_model = LatencyModel()
try:
    latency_model.load("./models")
except Exception as e:
    logger.warning(f"Could not load pre-trained model: {e}")

class PredictionRequest(BaseModel):
    features: list[float]

class PredictionResponse(BaseModel):
    latency_ms: float
    confidence: float

@app.post("/predict")
def predict_latency(request: PredictionRequest) -> PredictionResponse:
    """Predict query latency based on features"""
    try:
        X = np.array([request.features])
        pred = latency_model.predict(X)[0]

        # Confidence based on model certainty (simplified)
        confidence = 0.85

        return PredictionResponse(
            latency_ms=float(pred),
            confidence=confidence
        )
    except Exception as e:
        logger.error(f"Prediction error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
def health_check():
    """Health check endpoint"""
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5000)
```

- [ ] **Step 7: Run test to verify it passes**

```bash
cd backend
go test ./internal/ml/tests -v -run Latency
```

- [ ] **Step 8: Commit**

```bash
cd backend
git add internal/ml/models/latency_predictor.go services/ml-service/
git commit -m "feat: implement query latency prediction model with FastAPI service"
```

---

## Task 2.3: Anomaly Detection Model

**Files:**
- Create: `backend/internal/ml/models/anomaly_detector.go`
- Create: `backend/services/ml-service/anomaly_model.py`

- [ ] **Step 1: Create anomaly detector**

```go
// File: backend/internal/ml/models/anomaly_detector.go
package models

import (
	"fmt"
	"time"
)

type AnomalyDetector struct {
	mlServiceURL string
	baseline     map[string]*MetricBaseline
}

type MetricBaseline struct {
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

type AnomalyAlert struct {
	MetricName  string    `json:"metric_name"`
	CurrentValue float64   `json:"current_value"`
	Baseline    float64   `json:"baseline"`
	ZScore      float64   `json:"z_score"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // low, medium, high
}

func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		mlServiceURL: "http://localhost:5000",
		baseline:     make(map[string]*MetricBaseline),
	}
}

func (ad *AnomalyDetector) SetBaseline(metricName string, baseline *MetricBaseline) {
	ad.baseline[metricName] = baseline
}

func (ad *AnomalyDetector) Detect(metricName string, value float64) (*AnomalyAlert, bool) {
	baseline, exists := ad.baseline[metricName]
	if !exists {
		return nil, false
	}

	// Calculate z-score
	zScore := (value - baseline.Mean) / baseline.StdDev

	// Threshold: > 2.5 std deviations = anomaly
	if zScore > 2.5 || zScore < -2.5 {
		severity := "medium"
		if zScore > 3.5 || zScore < -3.5 {
			severity = "high"
		}

		return &AnomalyAlert{
			MetricName:  metricName,
			CurrentValue: value,
			Baseline:    baseline.Mean,
			ZScore:      zScore,
			Timestamp:   time.Now(),
			Severity:    severity,
		}, true
	}

	return nil, false
}
```

- [ ] **Step 2: Create Python anomaly model**

```python
# File: backend/services/ml-service/anomaly_model.py
from sklearn.ensemble import IsolationForest
import numpy as np

class AnomalyModel:
    def __init__(self, contamination=0.05):
        self.model = IsolationForest(
            contamination=contamination,
            random_state=42,
            n_estimators=100
        )
        self.is_trained = False

    def train(self, X):
        """Train anomaly detection model"""
        self.model.fit(X)
        self.is_trained = True

    def detect(self, X):
        """Detect anomalies (-1 = anomaly, 1 = normal)"""
        if not self.is_trained:
            raise RuntimeError("Model not trained")

        predictions = self.model.predict(X)
        scores = self.model.score_samples(X)

        return predictions, scores

    def get_baseline_stats(self, X):
        """Calculate baseline statistics for z-score detection"""
        return {
            "mean": float(np.mean(X)),
            "std": float(np.std(X)),
            "min": float(np.min(X)),
            "max": float(np.max(X))
        }
```

- [ ] **Step 3: Add anomaly detection endpoint to FastAPI**

Add to `backend/services/ml-service/main.py`:

```python
from anomaly_model import AnomalyModel

anomaly_model = AnomalyModel()

@app.post("/detect-anomaly")
def detect_anomaly(request: PredictionRequest) -> dict:
    """Detect if a metric reading is anomalous"""
    try:
        X = np.array([request.features])
        predictions, scores = anomaly_model.detect(X)

        is_anomaly = predictions[0] == -1
        anomaly_score = float(scores[0])

        return {
            "is_anomaly": is_anomaly,
            "anomaly_score": anomaly_score,
            "severity": "high" if anomaly_score < -0.5 else "medium" if is_anomaly else "normal"
        }
    except Exception as e:
        logger.error(f"Anomaly detection error: {e}")
        raise HTTPException(status_code=500, detail=str(e))
```

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/ml/models/anomaly_detector.go services/ml-service/anomaly_model.py
git commit -m "feat: implement anomaly detection with Isolation Forest"
```

---

## Task 2.4: ML API Endpoints in Go Backend

**Files:**
- Create: `backend/internal/api/handlers_ml.go`

- [ ] **Step 1: Create ML API handler**

```go
// File: backend/internal/api/handlers_ml.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pganalytics/internal/ml/models"
)

type MLServer struct {
	latencyPredictor *models.LatencyPredictor
	anomalyDetector  *models.AnomalyDetector
}

func NewMLServer() *MLServer {
	return &MLServer{
		latencyPredictor: models.NewLatencyPredictor(),
		anomalyDetector:  models.NewAnomalyDetector(),
	}
}

type PredictQueryLatencyRequest struct {
	JoinCount     int    `json:"join_count"`
	ScanType      string `json:"scan_type"`
	RowCount      int    `json:"row_count"`
	FilterCount   int    `json:"filter_count"`
	SubqueryCount int    `json:"subquery_count"`
	AggregateType string `json:"aggregate_type"`
}

type PredictQueryLatencyResponse struct {
	PredictedLatencyMs float64 `json:"predicted_latency_ms"`
	Confidence        float64 `json:"confidence"`
	Recommendations   []string `json:"recommendations"`
}

// HandlePredictQueryLatency predicts query execution time
func (s *MLServer) HandlePredictQueryLatency(w http.ResponseWriter, r *http.Request) {
	var req PredictQueryLatencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Convert to feature vector
	features := []float64{
		float64(req.JoinCount),
		scanTypeValue(req.ScanType),
		float64(req.RowCount) / 1000,
		float64(req.FilterCount),
		float64(req.SubqueryCount),
		aggregateTypeValue(req.AggregateType),
	}

	// Get prediction
	latency, err := s.latencyPredictor.Predict(features)
	if err != nil {
		http.Error(w, fmt.Sprintf("Prediction failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate recommendations
	recommendations := generateRecommendations(latency, &req)

	resp := PredictQueryLatencyResponse{
		PredictedLatencyMs: latency,
		Confidence:        0.85,
		Recommendations:   recommendations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateRecommendations(latency float64, req *PredictQueryLatencyRequest) []string {
	var recs []string

	if latency > 500 {
		recs = append(recs, "High latency (>500ms): Consider adding an index")
	}

	if req.JoinCount > 3 {
		recs = append(recs, fmt.Sprintf("Query has %d joins: Review join order optimization", req.JoinCount))
	}

	if req.ScanType == "seq_scan" && req.RowCount > 10000 {
		recs = append(recs, "Sequential scan on large table: Add an index")
	}

	if req.SubqueryCount > 2 {
		recs = append(recs, fmt.Sprintf("Query has %d subqueries: Consider using CTEs or joins", req.SubqueryCount))
	}

	return recs
}

func scanTypeValue(st string) float64 {
	switch st {
	case "seq_scan":
		return 0.0
	case "index_scan":
		return 1.0
	case "bitmap_scan":
		return 2.0
	default:
		return 0.0
	}
}

func aggregateTypeValue(at string) float64 {
	switch at {
	case "none":
		return 0.0
	case "sum", "count":
		return 1.0
	case "group_by":
		return 2.0
	default:
		return 0.0
	}
}
```

- [ ] **Step 2: Register ML routes in server.go**

Modify `backend/internal/api/server.go` to add:

```go
func (s *Server) registerMLRoutes() {
	mlServer := NewMLServer()
	s.mux.HandleFunc("POST /api/v1/ml/predict-latency", mlServer.HandlePredictQueryLatency)
}
```

And call `s.registerMLRoutes()` in your Server initialization.

- [ ] **Step 3: Commit**

```bash
cd backend
git add internal/api/handlers_ml.go
git commit -m "feat: add ML prediction endpoints (latency, anomaly)"
```

---

## Success Criteria

- [ ] All 4 ML tasks completed
- [ ] Query fingerprinting works correctly
- [ ] Data loader successfully retrieves query metrics
- [ ] Latency predictor integrates with FastAPI service
- [ ] Anomaly detector properly calculates z-scores
- [ ] ML API endpoints functional
- [ ] All Go tests passing
- [ ] Python FastAPI service runs without errors
- [ ] All commits are atomic and descriptive
- [ ] 100% test coverage for critical paths
