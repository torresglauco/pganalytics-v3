# Phase 4.5.5: Python ML Service - Testing Guide

**Date**: February 20, 2026
**Status**: Testing Procedures Ready
**Duration**: Estimated 2-3 hours for full test suite

---

## Quick Start (10 minutes)

### Docker Setup

```bash
# Navigate to ml-service directory
cd ml-service

# Start services with docker-compose
docker-compose up -d

# Verify services are running
docker-compose ps
# Expected: ml-service, postgres, redis all healthy

# Check logs
docker-compose logs -f ml-service
```

### Quick API Test

```bash
# Health check
curl http://localhost:8081/health
# Expected: {"status": "healthy", "service": "ml-service"}

# Get service status
curl http://localhost:8081/api/status
# Expected: {"service": "ml-service", "status": "healthy", "version": "1.0.0"}

# Get latest model
curl http://localhost:8081/api/models/latest
# Expected: Model metadata (mock response for now)

# Make prediction
curl -X POST http://localhost:8081/api/predict/query-execution \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001, "parameters": {}}'
# Expected: Prediction with confidence interval
```

---

## Unit Tests

### Test 1: Feature Extraction

**File**: `tests/test_models.py`

```python
def test_feature_extraction():
    """Test feature extraction from query metrics"""
    model = PerformanceModel('linear_regression')

    # Create sample query metrics
    query_metrics = {
        'query_calls_per_hour': 100,
        'mean_table_size_mb': 512,
        'index_count': 3,
        'has_seq_scan': 1,
        'has_nested_loop': 0,
        'subquery_depth': 0,
        'concurrent_queries_avg': 5,
        'available_memory_pct': 60,
        'std_dev_calls': 20,
        'peak_hour_calls': 150,
        'table_row_count': 50000,
        'avg_row_width_bytes': 256,
    }

    features = model.extract_features(query_metrics)

    # Verify feature shape
    assert features.shape == (1, 12)

    # Verify all features are numeric
    assert all(isinstance(f, (int, float)) for f in features[0])
```

---

### Test 2: Model Training

**File**: `tests/test_models.py`

```python
def test_model_training():
    """Test model training on synthetic data"""
    model = PerformanceModel('linear_regression')

    # Generate synthetic training data
    n_samples = 200
    n_features = 12
    X_train = np.random.randn(n_samples, n_features)
    # Y = 50 + 5*X1 + 3*X2 + noise
    y_train = 50 + 5*X_train[:, 0] + 3*X_train[:, 1] + np.random.randn(n_samples) * 5

    # Train model
    metrics = model.train(X_train, y_train)

    # Verify training metrics
    assert 'model_id' in metrics
    assert 'r_squared' in metrics
    assert 'rmse' in metrics
    assert metrics['r_squared'] > 0.5  # Synthetic data should have good fit
    assert metrics['training_samples'] == n_samples
```

---

### Test 3: Prediction Generation

**File**: `tests/test_models.py`

```python
def test_prediction():
    """Test model prediction with confidence interval"""
    # Train model first
    model = PerformanceModel('linear_regression')
    X_train = np.random.randn(200, 12)
    y_train = 50 + 5*X_train[:, 0] + 3*X_train[:, 1] + np.random.randn(200) * 5
    model.train(X_train, y_train)

    # Make prediction
    X_test = np.random.randn(1, 12)
    prediction = model.predict(X_test, return_confidence=True)

    # Verify prediction structure
    assert 'predicted_execution_time_ms' in prediction
    assert 'confidence_score' in prediction
    assert 'confidence_interval' in prediction

    # Verify values are reasonable
    assert prediction['predicted_execution_time_ms'] > 0
    assert 0 <= prediction['confidence_score'] <= 1
    assert prediction['confidence_interval']['lower_bound_ms'] < prediction['confidence_interval']['upper_bound_ms']
```

---

### Test 4: Model Serialization

**File**: `tests/test_models.py`

```python
def test_model_serialization():
    """Test model saving and loading"""
    import tempfile
    import os

    # Train model
    model = PerformanceModel('linear_regression', 'test-model')
    X_train = np.random.randn(200, 12)
    y_train = 50 + 5*X_train[:, 0] + np.random.randn(200) * 5
    model.train(X_train, y_train)

    # Save model
    with tempfile.NamedTemporaryFile(suffix='.pkl', delete=False) as f:
        filepath = f.name

    try:
        model.save(filepath)
        assert os.path.exists(filepath)

        # Load model
        loaded_model = PerformanceModel.load(filepath)

        # Verify loaded model is same
        assert loaded_model.model_name == 'test-model'
        assert loaded_model.r_squared == model.r_squared

        # Make prediction with loaded model
        X_test = np.random.randn(1, 12)
        prediction = loaded_model.predict(X_test)
        assert 'predicted_execution_time_ms' in prediction
    finally:
        os.unlink(filepath)
```

---

### Test 5: Different Model Types

**File**: `tests/test_models.py`

```python
def test_different_model_types():
    """Test training with different model types"""
    X_train = np.random.randn(200, 12)
    y_train = 50 + 5*X_train[:, 0] + 3*X_train[:, 1] + np.random.randn(200) * 5

    for model_type in ['linear_regression', 'decision_tree', 'random_forest']:
        model = PerformanceModel(model_type)
        metrics = model.train(X_train, y_train)

        assert metrics['model_type'] == model_type
        assert metrics['r_squared'] > 0.3

        # Make prediction
        X_test = np.random.randn(1, 12)
        prediction = model.predict(X_test)
        assert 'predicted_execution_time_ms' in prediction
```

---

## API Integration Tests

### Test 1: Health Check Endpoint

```bash
curl -X GET http://localhost:8081/health

# Expected response (200 OK):
# {
#   "status": "healthy",
#   "service": "ml-service"
# }
```

---

### Test 2: Training Endpoint

```bash
# Start training (returns 202 Accepted)
curl -X POST http://localhost:8081/api/train/performance-model \
  -H "Content-Type: application/json" \
  -d '{
    "database_name": "pganalytics",
    "lookback_days": 30,
    "model_type": "linear_regression",
    "force_retrain": false
  }'

# Expected response:
# {
#   "job_id": "train-20260220-001",
#   "status": "training",
#   "message": "Model training started in background"
# }
```

---

### Test 3: Training Status Endpoint

```bash
# Get training status (replace job_id with actual ID)
curl -X GET http://localhost:8081/api/train/performance-model/train-20260220-001

# Expected response (200 OK):
# {
#   "job_id": "train-20260220-001",
#   "status": "completed",
#   "model_id": "model-linear-001",
#   "r_squared": 0.78,
#   "rmse": 45.2,
#   "mae": 32.1,
#   "completed_at": "2026-02-20T..."
# }
```

---

### Test 4: Prediction Endpoint

```bash
curl -X POST http://localhost:8081/api/predict/query-execution \
  -H "Content-Type: application/json" \
  -d '{
    "query_hash": 4001,
    "parameters": {"param1": "value"},
    "scenario": "current"
  }'

# Expected response (200 OK):
# {
#   "query_hash": 4001,
#   "predicted_execution_time_ms": 125.5,
#   "confidence_score": 0.87,
#   "confidence_interval": {
#     "lower_bound_ms": 95.3,
#     "upper_bound_ms": 155.7,
#     "std_dev_ms": 15.2
#   },
#   "model_version": "v1.2",
#   "prediction_timestamp": "2026-02-20T..."
# }
```

---

### Test 5: Prediction Validation Endpoint

```bash
curl -X POST http://localhost:8081/api/validate/prediction \
  -H "Content-Type: application/json" \
  -d '{
    "prediction_id": "pred-001",
    "query_hash": 4001,
    "predicted_execution_time_ms": 125.5,
    "actual_execution_time_ms": 118.2,
    "model_version": "v1.2"
  }'

# Expected response (200 OK):
# {
#   "prediction_id": "pred-001",
#   "error_percent": 6.2,
#   "accuracy_score": 0.938,
#   "within_confidence_interval": true,
#   "message": "Prediction validation recorded"
# }
```

---

### Test 6: Model List Endpoint

```bash
curl -X GET http://localhost:8081/api/models

# Expected response (200 OK):
# {
#   "models": [
#     {
#       "model_id": "model-linear-001",
#       "model_type": "linear_regression",
#       "training_date": "2026-02-20T...",
#       "r_squared": 0.78,
#       "is_active": true
#     }
#   ],
#   "total_models": 2,
#   "active_model": "model-linear-001"
# }
```

---

### Test 7: Model Activation Endpoint

```bash
curl -X POST http://localhost:8081/api/models/model-linear-001/activate

# Expected response (200 OK):
# {
#   "model_id": "model-linear-001",
#   "status": "activated",
#   "message": "Model activated for predictions"
# }
```

---

## Error Handling Tests

### Test 1: Invalid Model Type

```bash
curl -X POST http://localhost:8081/api/train/performance-model \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "invalid_model"
  }'

# Expected response (400 Bad Request):
# {"error": "Invalid model_type"}
```

---

### Test 2: Missing Required Fields

```bash
curl -X POST http://localhost:8081/api/predict/query-execution \
  -H "Content-Type: application/json" \
  -d '{}'

# Expected response (400 Bad Request):
# {"error": "query_hash must be a positive integer"}
```

---

### Test 3: Invalid Request Format

```bash
curl -X POST http://localhost:8081/api/predict/query-execution \
  -H "Content-Type: application/json" \
  -d 'invalid json'

# Expected response (400 Bad Request):
# {"error": "Invalid JSON"}
```

---

## Performance Tests

### Test 1: Prediction Latency

```bash
# Run 100 predictions and measure average latency
for i in {1..100}; do
  time curl -X POST http://localhost:8081/api/predict/query-execution \
    -H "Content-Type: application/json" \
    -d '{"query_hash": 4001}'
done

# Expected: Average latency < 500ms
```

---

### Test 2: Concurrent Predictions

```bash
# Use Apache Bench to test concurrent requests
ab -n 1000 -c 10 -p data.json -T "application/json" \
  http://localhost:8081/api/predict/query-execution

# Expected:
# - Requests per second: > 20
# - Failed requests: 0
# - Average time per request: < 500ms
```

---

## Container Health Tests

### Test 1: Container Health Check

```bash
# Check container status
docker inspect --format='{{.State.Health.Status}}' pganalytics-ml-service

# Expected: "healthy"
```

---

### Test 2: Service Connectivity

```bash
# From within container, test PostgreSQL connection
docker exec pganalytics-ml-service \
  python -c "import psycopg2; conn = psycopg2.connect(dbname='pganalytics'); print('Connected')"

# Expected: "Connected"
```

---

## Integration Tests

### Test 1: Full Prediction Workflow

```bash
# 1. Start training
JOB=$(curl -s -X POST http://localhost:8081/api/train/performance-model \
  -H "Content-Type: application/json" \
  -d '{"lookback_days": 30}' | jq -r '.job_id')

# 2. Wait for training to complete
sleep 10

# 3. Get training status
curl http://localhost:8081/api/train/performance-model/$JOB | jq .

# 4. Make prediction with trained model
curl -X POST http://localhost:8081/api/predict/query-execution \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001}' | jq .

# 5. Validate prediction
curl -X POST http://localhost:8081/api/validate/prediction \
  -H "Content-Type: application/json" \
  -d '{
    "prediction_id": "test-001",
    "query_hash": 4001,
    "predicted_execution_time_ms": 125.5,
    "actual_execution_time_ms": 118.2
  }' | jq .
```

---

### Test 2: Model Version Management

```bash
# 1. Get latest model
curl http://localhost:8081/api/models/latest | jq '.model_id'

# 2. List all models
curl http://localhost:8081/api/models | jq '.models[].model_id'

# 3. Activate different model
MODEL_ID=$(curl http://localhost:8081/api/models | jq -r '.models[1].model_id')
curl -X POST http://localhost:8081/api/models/$MODEL_ID/activate | jq .

# 4. Verify activation
curl http://localhost:8081/api/models/latest | jq '.model_id'
# Should be the newly activated model
```

---

## Success Criteria

✅ All unit tests pass
✅ All API endpoints return expected responses
✅ Prediction latency < 500ms
✅ Concurrent requests handled properly
✅ Error handling returns correct status codes
✅ Container health checks pass
✅ Database connectivity verified
✅ Model serialization/deserialization works

---

## Running Full Test Suite

```bash
# From ml-service directory

# 1. Start services
docker-compose up -d

# 2. Run unit tests
pytest tests/ -v

# 3. Run integration tests
python -m pytest tests/test_api.py -v

# 4. Performance test
ab -n 100 -c 5 http://localhost:8081/api/status

# 5. Check logs
docker-compose logs ml-service | tail -20

# 6. Clean up
docker-compose down
```

---

**Estimated Duration**: 2-3 hours for complete test suite
**Success Rate Target**: > 95% test pass rate

