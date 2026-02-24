# Phase 4.5.6: Database Integration and Async Tasks - Implementation Guide

**Date**: February 20, 2026
**Status**: Implementation In Progress
**Components**: Celery tasks, job management, enhanced handlers, database integration

---

## Overview

Phase 4.5.6 adds production-ready database integration and asynchronous task processing to the Python ML service. This phase replaces mock responses with real database queries and implements background job processing with Celery.

---

## New Files Created

### 1. Celery Tasks Module (210 lines)
**File**: `ml-service/tasks.py`

Implements async background tasks:

#### `train_performance_model()` - Async Model Training
```python
@celery_app.task(bind=True, max_retries=3)
def train_performance_model(
    self,
    database_url: str,
    lookback_days: int,
    model_type: str,
    job_id: str
) -> Dict[str, Any]
```

**Features**:
- Extract historical data from database
- Validate features and handle missing values
- Train scikit-learn model with cross-validation
- Save model metadata to database
- Return training metrics (R², RMSE, MAE)
- Retry logic with exponential backoff (max 3 retries)
- Time limits: 10 minutes hard, 9 minutes soft
- Error handling and detailed logging

**Result Dictionary**:
```python
{
    'job_id': 'train-20260220-001',
    'status': 'completed',
    'model_id': 'model-rf-abc123',
    'model_type': 'random_forest',
    'r_squared': 0.78,
    'rmse': 45.2,
    'mae': 32.1,
    'training_samples': 1500,
    'completed_at': '2026-02-20T10:30:45.123456'
}
```

#### `validate_prediction()` - Async Validation
```python
@celery_app.task(bind=True, max_retries=2)
def validate_prediction(
    self,
    database_url: str,
    prediction_id: str,
    query_hash: int,
    predicted_ms: float,
    actual_ms: float,
    model_version: str
) -> Dict[str, Any]
```

**Features**:
- Calculate prediction error metrics
- Compare against confidence intervals
- Track accuracy for model learning loop
- Async job with 2-retry limit
- Uses actual execution time to validate predictions

**Result Dictionary**:
```python
{
    'prediction_id': 'pred-001',
    'query_hash': 4001,
    'error_percent': 6.2,
    'accuracy_score': 0.938,
    'within_confidence_interval': True,
    'validated_at': '2026-02-20T10:30:45.123456'
}
```

#### `collect_prediction_metrics()` - Async Metrics Aggregation
```python
@celery_app.task(bind=True, max_retries=2)
def collect_prediction_metrics(
    self,
    database_url: str,
    job_id: str
) -> Dict[str, Any]
```

**Purpose**: Aggregate prediction validation results for monitoring and analysis

#### `cleanup_old_models()` - Async Model Cleanup
```python
@celery_app.task
def cleanup_old_models(
    database_url: str,
    keep_count: int = 5
) -> Dict[str, Any]
```

**Purpose**: Remove old model versions, keeping only recent ones

#### `health_check()` - Celery Health
```python
@celery_app.task
def health_check() -> Dict[str, Any]
```

**Purpose**: Verify Celery worker is responsive

### Celery Configuration
```python
celery_app.conf.update(
    broker_url=Config.CELERY_BROKER,
    result_backend=Config.CELERY_BACKEND,
    task_time_limit=600,        # 10 minutes hard limit
    task_soft_time_limit=540,   # 9 minutes soft limit
    result_expires=3600,         # Results expire after 1 hour
    task_track_started=True,     # Track when task starts
)
```

### 2. Job Manager Utilities (250 lines)
**File**: `ml-service/utils/job_manager.py`

Manages async job tracking and status:

#### `JobManager` Class
```python
class JobManager:
    # In-memory job store (Redis/database in production)

    @staticmethod
    def create_job(job_type: str, **kwargs) -> Dict
    @staticmethod
    def get_job(job_id: str) -> Optional[Dict]
    @staticmethod
    def update_job(job_id: str, **kwargs) -> Optional[Dict]
    @staticmethod
    def set_status(job_id: str, status: str) -> Optional[Dict]
    @staticmethod
    def set_result(job_id: str, result: Dict) -> Optional[Dict]
    @staticmethod
    def set_error(job_id: str, error: str) -> Optional[Dict]
    @staticmethod
    def list_jobs(job_type: str, status: str) -> List[Dict]
    @staticmethod
    def clear_old_jobs(max_age_hours: int) -> int
```

#### `JobStatus` Constants
```python
class JobStatus:
    PENDING = 'pending'
    TRAINING = 'training'
    COMPLETED = 'completed'
    FAILED = 'failed'
```

#### `TrainingJobManager` Helper
```python
class TrainingJobManager:
    @staticmethod
    def create_training_job(...) -> Dict
    @staticmethod
    def get_training_job(job_id: str) -> Optional[Dict]
    @staticmethod
    def mark_training_started(job_id: str) -> Optional[Dict]
    @staticmethod
    def mark_training_completed(job_id: str, model_id: str, metrics: Dict)
    @staticmethod
    def mark_training_failed(job_id: str, error: str)
```

**Job Record Structure**:
```python
{
    'job_id': 'train-20260220-abc123',
    'job_type': 'training',
    'status': 'training',        # pending, training, completed, failed
    'created_at': '2026-02-20T10:00:00',
    'updated_at': '2026-02-20T10:05:00',
    'result': {...},             # Set when completed
    'error': None,               # Set if failed
    'database_name': 'pganalytics',
    'lookback_days': 90,
    'model_type': 'random_forest',
    'force_retrain': False,
}
```

### 3. Enhanced API Handlers (Updated)
**File**: `ml-service/api/handlers.py` (Updated)

**Major Changes**:

#### Training Handler Enhancement
```python
def handle_train_performance_model(request):
    # 1. Create job record with TrainingJobManager
    job = TrainingJobManager.create_training_job(...)

    # 2. Launch async Celery task
    if CELERY_AVAILABLE:
        task = train_performance_model.delay(
            database_url=database_url,
            lookback_days=lookback_days,
            model_type=model_type,
            job_id=job_id
        )

    # 3. Return 202 Accepted with job_id
    return jsonify({'job_id': job_id, 'status': 'training'}), 202
```

#### Training Status Handler Enhancement
```python
def handle_get_training_status(job_id):
    # Retrieve job from JobManager
    job = TrainingJobManager.get_training_job(job_id)

    # Return job status and result if available
    response = {
        'job_id': job_id,
        'status': job['status'],
        'created_at': job['created_at'],
    }

    if job['result']:
        response.update(job['result'])  # Add metrics when completed

    return jsonify(response)
```

#### Prediction Handler Enhancement
```python
def handle_predict_query_execution(request):
    # 1. Get database URL from environment
    database_url = os.environ.get('DATABASE_URL', ...)

    # 2. Try real database feature extraction
    db = DatabaseConnection(database_url)
    query_metrics = db.extract_features_for_query(query_hash)

    # 3. Return real or estimated prediction
    # Falls back to mock if database unavailable
```

#### Validation Handler Enhancement
```python
def handle_validate_prediction(request):
    # 1. Extract and validate prediction data

    # 2. Launch async validation task
    if CELERY_AVAILABLE:
        task = validate_prediction.delay(
            database_url=database_url,
            prediction_id=prediction_id,
            query_hash=query_hash,
            predicted_ms=predicted_ms,
            actual_ms=actual_ms,
            model_version=model_version
        )

    # 3. Return validation results immediately
    # Actual recording happens asynchronously
```

#### Model Status Handler Enhancement
```python
def handle_service_status():
    # Report job queue status
    {
        'service': 'ml-service',
        'status': 'healthy',
        'database_connected': True,
        'celery_available': True,
        'pending_jobs': 2,
        'training_jobs': 1,
        ...
    }
```

### 4. Celery Worker Entry Point
**File**: `ml-service/celery_worker.py`

```python
def run_worker():
    """Run Celery worker with configuration"""
    celery_app.worker_main([
        'worker',
        '--loglevel=info',
        '--concurrency=4',
        '--pool=threads',
    ])
```

**Usage**:
```bash
# Start worker in separate terminal
python celery_worker.py

# Or with specific queue
celery -A tasks worker --loglevel=info
```

---

## Database Integration Architecture

### Feature Extraction Flow
```
API Handler
    ↓
DatabaseConnection.get_connection() [context manager]
    ↓
extract_features_for_query(query_hash)
    ↓
Query: SELECT features FROM metrics_pg_stats_query WHERE query_hash=?
    ↓
FeatureEngineer.extract_from_metrics()
    ↓
NumPy array of 12 features
    ↓
Model.predict()
```

### Training Task Flow
```
API Handler (202 Accepted)
    ↓
train_performance_model.delay() [Celery]
    ↓
Worker receives task
    ↓
DatabaseConnection.extract_training_data()
    ↓
Query: SELECT features FROM metrics_pg_stats_query (30-90 day window)
    ↓
FeatureEngineer.validate_features()
    ↓
PerformanceModel.train()
    ↓
DatabaseConnection.save_model_metadata()
    ↓
Insert: query_performance_models table
    ↓
JobManager.set_result() [Update job status]
```

### Validation Flow
```
API Handler
    ↓
validate_prediction.delay() [Celery]
    ↓
Worker receives task
    ↓
Calculate error metrics
    ↓
DatabaseConnection.record_prediction()
    ↓
Return validation results
```

---

## Configuration

### Environment Variables

**Required**:
```bash
DATABASE_URL=postgresql://pganalytics:password@postgres:5432/pganalytics
CELERY_BROKER=redis://redis:6379/0
CELERY_BACKEND=redis://redis:6379/1
```

**Optional**:
```bash
FLASK_ENV=development
LOG_LEVEL=DEBUG
ML_SERVICE_PORT=8081
CELERY_WORKERS=4
CELERY_SOFT_TIMEOUT=540     # 9 minutes
CELERY_HARD_TIMEOUT=600     # 10 minutes
```

### Docker Compose Updates Needed
```yaml
services:
  ml-service:
    # ... existing config ...
    depends_on:
      - postgres
      - redis
    command: python -m flask run --host=0.0.0.0 --port=8081

  celery-worker:
    build: ./ml-service
    command: python celery_worker.py
    environment:
      - DATABASE_URL=postgresql://pganalytics:password@postgres:5432/pganalytics
      - CELERY_BROKER=redis://redis:6379/0
      - CELERY_BACKEND=redis://redis:6379/1
    depends_on:
      - postgres
      - redis
    networks:
      - pganalytics
```

---

## API Behavior Changes

### Training Endpoint
**Request**:
```bash
POST /api/train/performance-model
{
  "database_name": "pganalytics",
  "lookback_days": 90,
  "model_type": "random_forest",
  "force_retrain": false
}
```

**Response** (202 Accepted):
```json
{
  "job_id": "train-20260220-001",
  "status": "training",
  "database_name": "pganalytics",
  "lookback_days": 90,
  "model_type": "random_forest",
  "message": "Model training started in background"
}
```

**Check Status**:
```bash
GET /api/train/performance-model/train-20260220-001
```

**While Training** (202):
```json
{
  "job_id": "train-20260220-001",
  "status": "training",
  "created_at": "2026-02-20T10:00:00",
  "updated_at": "2026-02-20T10:05:30"
}
```

**When Complete** (200):
```json
{
  "job_id": "train-20260220-001",
  "status": "completed",
  "model_id": "model-rf-abc123",
  "model_type": "random_forest",
  "r_squared": 0.78,
  "rmse": 45.2,
  "mae": 32.1,
  "training_samples": 1500,
  "completed_at": "2026-02-20T10:10:00"
}
```

### Prediction Endpoint
**Behavior Changes**:
- Attempts real database feature extraction first
- Falls back to mock if database unavailable
- Includes 'source' field: 'feature-based-estimation' or 'mock'
- Database connection is optional for API availability

### Validation Endpoint
**Behavior Changes**:
- Queues async validation task if Celery available
- Returns immediate response with calculated metrics
- Validation record happens asynchronously
- Supports learning loop for model improvement

---

## Testing with Database Integration

### Unit Tests for Job Manager
```python
def test_create_training_job():
    job = TrainingJobManager.create_training_job(
        database_name='pganalytics',
        lookback_days=90,
        model_type='random_forest'
    )
    assert job['job_id'].startswith('train-')
    assert job['status'] == 'pending'

def test_job_status_transitions():
    job = JobManager.create_job('training')
    job_id = job['job_id']

    JobManager.set_status(job_id, 'training')
    assert JobManager.get_job(job_id)['status'] == 'training'

    JobManager.set_result(job_id, {'model_id': 'abc'})
    assert JobManager.get_job(job_id)['status'] == 'completed'
```

### Integration Tests for Handlers
```python
def test_training_endpoint_returns_job_id(client):
    response = client.post('/api/train/performance-model', json={
        'lookback_days': 30,
        'model_type': 'linear_regression'
    })

    assert response.status_code == 202
    data = response.get_json()
    assert 'job_id' in data
    assert data['status'] == 'training'

def test_training_status_endpoint(client):
    # Start training
    train_resp = client.post('/api/train/performance-model', json={
        'lookback_days': 30
    })
    job_id = train_resp.get_json()['job_id']

    # Check status
    status_resp = client.get(f'/api/train/performance-model/{job_id}')
    assert status_resp.status_code == 200
    assert status_resp.get_json()['job_id'] == job_id
```

### End-to-End Testing with Celery
```python
def test_training_with_database(celery_worker, postgres_db):
    # Start training
    response = client.post('/api/train/performance-model', json={
        'lookback_days': 30,
        'model_type': 'random_forest'
    })

    job_id = response.get_json()['job_id']

    # Wait for Celery task to complete
    time.sleep(5)

    # Check final status
    status = client.get(f'/api/train/performance-model/{job_id}')
    assert status.get_json()['status'] == 'completed'
    assert 'model_id' in status.get_json()
    assert 'r_squared' in status.get_json()
```

---

## Error Handling

### Database Connection Errors
```python
# In handler:
try:
    db = DatabaseConnection(database_url)
    if db.initialize():
        # ... use database ...
except Exception as e:
    logger.warning(f"Database error: {e}")
    # Fallback to mock response
    return fallback_response()
```

### Celery Task Failures
```python
# In task:
@celery_app.task(bind=True, max_retries=3)
def train_performance_model(self, ...):
    try:
        # ... training logic ...
    except SoftTimeLimitExceeded:
        logger.error("Task exceeded time limit")
        raise
    except Exception as e:
        # Retry with exponential backoff
        raise self.retry(exc=e, countdown=60 * (2 ** self.request.retries))
```

### Graceful Degradation
```python
# Service status reports health
{
    'status': 'healthy',              # Still healthy even if DB down
    'database_connected': False,      # Report individual component status
    'celery_available': True,
    'pending_jobs': 2,
    'training_jobs': 0
}
```

---

## Performance Considerations

### Celery Concurrency
- Default: 4 concurrent workers
- Configurable via `CELERY_WORKERS` environment variable
- Thread pool for I/O-bound tasks (database, network)

### Job Storage
- **Development**: In-memory (test/prototype only)
- **Production**: Redis or database (next phase)
- Job TTL: 1 hour default
- Old jobs cleared every 24 hours

### Database Connection Pooling
- Min connections: 2
- Max connections: 10
- Automatic cleanup on context exit
- Timeout: 30 seconds per query

### Timeout Configuration
```python
# Celery hard limit: 10 minutes
# Celery soft limit: 9 minutes (graceful shutdown attempt)
# Model training target: 2-5 minutes for 1000 samples
# Large datasets (5000+ samples): 5-8 minutes
```

---

## Future Enhancements (Phase 4.5.7+)

1. **Redis Job Storage**
   - Replace in-memory with Redis HASH storage
   - Persistent job history
   - Cluster-safe job tracking

2. **Database Job Storage**
   - Create `celery_jobs` table
   - Full job audit trail
   - Integration with monitoring dashboards

3. **Model Persistence**
   - Store trained models in database (BYTEA)
   - Version tracking and rollback
   - Active model selection per database

4. **Prediction Caching**
   - Cache predictions in Redis
   - TTL-based cache invalidation
   - Cache hit/miss metrics

5. **Metrics Collection**
   - Prometheus metrics for tasks
   - Training duration histograms
   - Prediction accuracy tracking
   - Database query performance

6. **Scheduled Jobs**
   - Daily model retraining
   - Weekly metrics aggregation
   - Cleanup old jobs/models

7. **Go Backend Integration**
   - HTTP calls to ML service for predictions
   - Circuit breaker pattern
   - Timeout and retry logic
   - Request queuing if overloaded

---

## Summary of Changes

### Files Created
1. **tasks.py** (210 lines) - Celery async tasks
2. **utils/job_manager.py** (250 lines) - Job tracking
3. **celery_worker.py** (25 lines) - Worker entry point

### Files Modified
1. **api/handlers.py** (Enhanced - 50+ line changes)
   - Database integration in predict/train handlers
   - Celery task launching
   - Job status tracking
   - Error handling with fallbacks

### New Features
- ✅ Async model training with Celery
- ✅ Job status tracking
- ✅ Database feature extraction
- ✅ Graceful fallbacks for database unavailability
- ✅ Retry logic for failed tasks
- ✅ Time limits and timeout handling
- ✅ Comprehensive error logging

### API Compatibility
- ✅ All endpoints maintain same signatures
- ✅ No breaking changes
- ✅ Backward compatible with Phase 4.5.5

---

**Status**: Phase 4.5.6 Implementation In Progress

**Next Steps**:
1. Complete database method implementations in DatabaseConnection
2. Add Redis job storage
3. Implement model persistence
4. Add comprehensive integration tests
5. Deploy and performance test

