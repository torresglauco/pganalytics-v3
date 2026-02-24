# Phase 4.5.5: Python ML Service - Implementation Specification

**Date**: February 20, 2026
**Status**: Implementation In Progress
**Duration**: 3-4 days estimated
**Lines of Code**: 600+ (Python service + configuration)

---

## Overview

Phase 4.5.5 implements a Python-based machine learning microservice that:

1. **Trains** performance prediction models on query historical data
2. **Predicts** query execution time with confidence intervals
3. **Validates** predictions against actual results
4. **Refines** model accuracy through learning loops
5. **Provides** REST API for Go backend integration

---

## Architecture

### System Design

```
┌─────────────────────────────────────────────────────────────┐
│ Go Backend API Server (pganalytics-api)                     │
│ - Handles HTTP requests                                     │
│ - Query parameter validation                                │
│ - Authentication & authorization                            │
└─────────────────────────────────────────────────────────────┘
                          ↕ HTTP/JSON
                   (localhost:8081)
┌─────────────────────────────────────────────────────────────┐
│ Python ML Service (ml-service)                              │
│ - scikit-learn models                                       │
│ - Feature engineering                                       │
│ - Model training & inference                                │
│ - Prediction confidence calculation                         │
└─────────────────────────────────────────────────────────────┘
                          ↕ psycopg2
                   PostgreSQL Database
┌─────────────────────────────────────────────────────────────┐
│ Feature Storage & Model Persistence                         │
│ - Query metrics (training data)                             │
│ - Model parameters (JSONB)                                  │
│ - Prediction history (for validation)                       │
└─────────────────────────────────────────────────────────────┘
```

### Service Structure

```
ml-service/
├── app.py                    # Flask app initialization
├── config.py                 # Configuration management
├── requirements.txt          # Python dependencies
├── Dockerfile                # Container image
├── docker-compose.yml        # Local development
│
├── models/
│   ├── __init__.py
│   ├── performance_predictor.py     # Main ML model class
│   ├── pattern_detector.py          # Pattern detection (future)
│   └── model_validator.py           # Cross-validation & metrics
│
├── api/
│   ├── __init__.py
│   ├── routes.py             # API endpoint definitions
│   └── handlers.py           # Endpoint handler logic
│
├── utils/
│   ├── __init__.py
│   ├── db_connection.py      # PostgreSQL connection pool
│   ├── feature_engineer.py   # Feature extraction & transformation
│   ├── model_storage.py      # Model serialization/deserialization
│   └── logger.py             # Logging configuration
│
└── tests/
    ├── __init__.py
    ├── test_models.py        # Model training tests
    ├── test_api.py           # API endpoint tests
    └── test_features.py      # Feature engineering tests
```

---

## Core Components

### 1. PerformanceModel Class

**Purpose**: Main ML model for execution time prediction

**Methods**:

```python
class PerformanceModel:
    def __init__(self, model_type='linear_regression', model_name=None):
        """Initialize model"""
        self.model_type = model_type  # 'linear_regression', 'decision_tree', 'random_forest'
        self.model_name = model_name
        self.model = None
        self.scaler = None  # StandardScaler for feature normalization
        self.feature_names = None
        self.r_squared = None
        self.training_date = None

    def extract_features(self, query_hash: int, from_db=True) -> Dict:
        """
        Extract features for a query from database

        Features:
        - query_characteristics: fingerprint_hash, scan_type, join_type, subquery_depth
        - table_statistics: row_count, table_size_mb, index_count
        - historical_patterns: avg_calls_per_hour, peak_hour_impact
        - system_state: concurrent_queries_avg, available_memory_pct
        - optimization_flags: has_index_scan, has_bitmap_scan, has_hash_join

        Returns: Feature vector as dict/array
        """

    def train(self, queries_lookback_days=90) -> Dict:
        """
        Train model on historical query data

        Steps:
        1. Extract features for all queries in lookback window
        2. Get target variable (mean_exec_time_ms)
        3. Handle missing values and outliers
        4. Split into train/test (80/20)
        5. Scale features (StandardScaler)
        6. Train model (linear regression, decision tree, or random forest)
        7. Perform cross-validation (5-fold)
        8. Evaluate on test set (R², RMSE, MAE)
        9. Store model and metrics

        Returns: {
            'model_id': uuid,
            'model_type': 'linear_regression',
            'training_samples': 1500,
            'r_squared': 0.78,
            'rmse': 45.2,
            'mae': 32.1,
            'training_date': '2026-02-20T...',
            'feature_count': 12
        }
        """

    def predict(self, features: Dict, return_confidence=True) -> Dict:
        """
        Predict execution time for query

        Steps:
        1. Validate input features
        2. Apply feature scaling
        3. Get model prediction
        4. Calculate prediction standard deviation
        5. Calculate confidence interval (±σ, ±2σ)
        6. Generate confidence score

        Returns: {
            'predicted_execution_time_ms': 125.5,
            'confidence_score': 0.87,
            'confidence_interval': {
                'lower_bound': 95.3,
                'upper_bound': 155.7,
                'std_dev': 15.2
            },
            'model_version': 'v1.2',
            'prediction_timestamp': '2026-02-20T...'
        }
        """

    def save_to_db(self, db_connection):
        """Serialize and store model in PostgreSQL query_performance_models table"""

    def load_from_db(self, db_connection, model_id: str):
        """Retrieve and deserialize model from database"""

    def evaluate(self, X_test, y_test) -> Dict:
        """
        Evaluate model performance

        Metrics:
        - R² Score: coefficient of determination (0-1, higher is better)
        - RMSE: root mean squared error (execution time units)
        - MAE: mean absolute error (execution time units)
        - Prediction errors: histogram of prediction - actual

        Returns: {
            'r_squared': 0.78,
            'rmse': 45.2,
            'mae': 32.1,
            'median_error': 15.0,
            'prediction_errors': {...}
        }
        """
```

---

## API Endpoints

### 1. Training Endpoints

#### POST /api/train/performance-model

**Purpose**: Trigger async model training on historical data

**Request**:
```json
{
  "database_name": "pganalytics",
  "lookback_days": 90,
  "model_type": "linear_regression",
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
  "model_type": "linear_regression",
  "message": "Model training started in background"
}
```

**Implementation**:
- Launch async Celery task for model training
- Return immediately with job_id
- Train on historical query metrics (30-90 day window)
- Store trained model in database
- Log training progress

---

#### GET /api/train/performance-model/{job_id}

**Purpose**: Check status of training job

**Response** (200 OK):
```json
{
  "job_id": "train-20260220-001",
  "status": "completed",
  "model_id": "model-linear-001",
  "model_type": "linear_regression",
  "training_samples": 1500,
  "r_squared": 0.78,
  "rmse": 45.2,
  "mae": 32.1,
  "training_duration_seconds": 234,
  "completed_at": "2026-02-20T15:45:30Z"
}
```

---

### 2. Prediction Endpoints

#### POST /api/predict/query-execution

**Purpose**: Predict query execution time

**Request**:
```json
{
  "query_hash": 4001,
  "parameters": {
    "param1": "value1",
    "param2": 100
  },
  "scenario": "current"
}
```

**Response** (200 OK):
```json
{
  "query_hash": 4001,
  "predicted_execution_time_ms": 125.5,
  "confidence_score": 0.87,
  "confidence_interval": {
    "lower_bound_ms": 95.3,
    "upper_bound_ms": 155.7,
    "std_dev_ms": 15.2
  },
  "model_version": "v1.2",
  "prediction_timestamp": "2026-02-20T15:30:00Z",
  "message": "Prediction successful"
}
```

**Implementation**:
- Extract features from query and parameters
- Load latest trained model
- Generate prediction with confidence interval
- Return prediction with uncertainty bounds

---

#### POST /api/validate/prediction

**Purpose**: Record actual result and validate prediction accuracy

**Request**:
```json
{
  "prediction_id": "pred-20260220-001",
  "query_hash": 4001,
  "predicted_execution_time_ms": 125.5,
  "actual_execution_time_ms": 118.2,
  "model_version": "v1.2"
}
```

**Response** (200 OK):
```json
{
  "prediction_id": "pred-20260220-001",
  "error_percent": 6.2,
  "accuracy_score": 0.938,
  "within_confidence_interval": true,
  "message": "Prediction validation recorded"
}
```

**Implementation**:
- Calculate error: ABS(predicted - actual) / actual × 100%
- Calculate accuracy: 1 - (error / 100)
- Track prediction accuracy over time
- Use for model drift detection

---

### 3. Model Management Endpoints

#### GET /api/models/latest

**Purpose**: Get latest trained model metadata

**Response** (200 OK):
```json
{
  "model_id": "model-linear-001",
  "model_type": "linear_regression",
  "model_name": "Q-Exec-Predictor-v1.2",
  "training_samples": 1500,
  "training_date": "2026-02-20T14:30:00Z",
  "r_squared": 0.78,
  "rmse": 45.2,
  "mae": 32.1,
  "feature_count": 12,
  "feature_names": [
    "query_calls_per_hour",
    "mean_table_size_mb",
    "index_count",
    ...
  ],
  "prediction_error_std": 12.5,
  "is_active": true
}
```

---

#### GET /api/models/{model_id}

**Purpose**: Get specific model metadata

---

#### GET /api/models

**Purpose**: List all trained models with versions

**Response** (200 OK):
```json
{
  "models": [
    {
      "model_id": "model-linear-001",
      "model_type": "linear_regression",
      "training_date": "2026-02-20T14:30:00Z",
      "r_squared": 0.78,
      "is_active": true
    },
    {
      "model_id": "model-tree-001",
      "model_type": "decision_tree",
      "training_date": "2026-02-20T15:00:00Z",
      "r_squared": 0.75,
      "is_active": false
    }
  ],
  "total_models": 2,
  "active_model": "model-linear-001"
}
```

---

## Feature Engineering

### Query Features Extracted

```python
FEATURE_GROUPS = {
    'query_characteristics': [
        'fingerprint_hash',           # Query pattern ID
        'scan_type_seq_scan',         # Boolean: has sequential scan
        'scan_type_index_scan',       # Boolean: has index scan
        'scan_type_bitmap_scan',      # Boolean: has bitmap scan
        'join_type_nested_loop',      # Boolean: has nested loop
        'join_type_hash_join',        # Boolean: has hash join
        'join_type_merge_join',       # Boolean: has merge join
        'subquery_depth',             # Nesting level of subqueries
    ],
    'table_statistics': [
        'row_count',                  # Number of rows in table
        'table_size_mb',              # Physical size in MB
        'index_count',                # Number of indexes
        'avg_row_width_bytes',        # Average row size
    ],
    'historical_patterns': [
        'avg_calls_per_hour',         # Query execution frequency
        'peak_hour_calls',            # Max calls in any hour
        'std_dev_calls',              # Variability in call frequency
        'calls_per_day_coefficient',  # Day-to-day variance
    ],
    'system_state': [
        'concurrent_queries_avg',     # Average concurrent queries
        'available_memory_pct',       # Available system memory %
        'cpu_utilization_avg',        # Average CPU usage
    ],
    'optimization_indicators': [
        'has_index_scan',             # Fast index access
        'has_bitmap_scan',            # Efficient multi-index access
        'has_hash_join',              # Efficient join method
        'missing_index_indicator',    # Seq scan where index would help
    ]
}
```

### Feature Extraction Process

```python
def extract_features(query_hash: int, from_db=True) -> np.ndarray:
    """Extract feature vector for query"""

    # Step 1: Get query metadata
    query = get_query_metadata(query_hash)  # From metrics_pg_stats_query

    # Step 2: Parse EXPLAIN plan
    explain_plan = parse_explain_plan(query.plan_json)

    # Step 3: Extract table statistics
    table_stats = get_table_statistics(query.table_names)

    # Step 4: Get historical patterns
    patterns = get_historical_patterns(query_hash, lookback_days=30)

    # Step 5: Get system state
    sys_state = get_system_state()

    # Step 6: Engineer features
    features = {
        # Query characteristics
        'scan_type_seq_scan': explain_plan.has_seq_scan,
        'scan_type_index_scan': explain_plan.has_index_scan,
        'join_type_nested_loop': explain_plan.has_nested_loop,
        'subquery_depth': count_subqueries(query.text),

        # Table statistics
        'row_count': table_stats[0].row_count,
        'table_size_mb': table_stats[0].size_mb,
        'index_count': len(table_stats[0].indexes),

        # Historical patterns
        'avg_calls_per_hour': patterns.avg_calls,
        'peak_hour_calls': patterns.peak_calls,
        'std_dev_calls': patterns.std_dev,

        # System state
        'concurrent_queries_avg': sys_state.concurrent_queries,
        'available_memory_pct': sys_state.memory_available,
    }

    return np.array([features[name] for name in FEATURE_ORDER])
```

---

## Model Training Pipeline

### Data Preparation

```python
def prepare_training_data(lookback_days=90) -> Tuple[np.ndarray, np.ndarray]:
    """
    Prepare training data from historical metrics

    Steps:
    1. Query database for metrics from last N days
    2. Filter: remove outliers, handle missing values
    3. Extract features for each query
    4. Get target variable (mean_exec_time_ms)
    5. Remove any samples with missing features
    6. Normalize features (StandardScaler)

    Returns: (X_train, y_train) feature and target arrays
    """
```

### Model Types Supported

1. **Linear Regression** (Default)
   - Fast training
   - Interpretable coefficients
   - Good baseline model
   - Accuracy: ~75-80% R²

2. **Decision Tree Regressor**
   - Handles non-linearity
   - Auto feature selection
   - Prone to overfitting
   - Accuracy: ~70-75% R²

3. **Random Forest Regressor**
   - Ensemble of decision trees
   - More robust than single tree
   - Better handling of non-linearity
   - Accuracy: ~80-85% R²

4. **Gradient Boosting** (Optional)
   - XGBoost for advanced predictions
   - Best accuracy but slower training
   - Accuracy: ~85-90% R²

---

## Confidence Calculation

### Method 1: Standard Deviation Based

```python
def calculate_confidence(prediction: float, std_dev: float) -> float:
    """
    Calculate confidence based on prediction uncertainty

    confidence = 1 - (std_dev / prediction)

    Range: 0-1
    High std_dev relative to prediction → lower confidence
    """
```

### Method 2: R² Based

```python
def calculate_confidence_from_r_squared(r_squared: float) -> float:
    """
    confidence = r_squared

    Directly maps model accuracy to prediction confidence
    """
```

### Method 3: Combined Score

```python
def calculate_confidence_combined(
    r_squared: float,
    std_dev: float,
    prediction: float,
    prediction_error_history: List[float]
) -> float:
    """
    confidence = (r_squared + 0.5) × (1 - normalized_std_dev)

    Takes into account:
    - Model quality (R²)
    - Prediction uncertainty (std_dev)
    - Historical prediction errors
    """
```

---

## Database Integration

### Feature Extraction Queries

```sql
-- Get query metrics for feature engineering
SELECT
    query_hash, mean_exec_time_ms, calls, rows,
    p95_exec_time_ms, p99_exec_time_ms,
    fingerprint_hash, plan_json
FROM metrics_pg_stats_query
WHERE collected_at > NOW() - INTERVAL '90 days'
ORDER BY collected_at DESC;

-- Get table statistics
SELECT
    table_name, row_count, total_bytes,
    (SELECT COUNT(*) FROM pg_indexes WHERE tablename = t.tablename) as index_count
FROM pg_tables t
WHERE schemaname = 'public';
```

### Model Storage

```sql
-- Store trained model
INSERT INTO query_performance_models (
    model_type, model_name, model_binary, model_json,
    feature_names, training_sample_size, r_squared,
    training_date
)
VALUES (
    'linear_regression',
    'Q-Exec-Predictor-v1.2',
    E'\\x...',  -- Pickled model binary
    '{"coefficients": {...}, "intercept": 45.2}',
    ARRAY['feature1', 'feature2', ...],
    1500,
    0.78,
    NOW()
);
```

---

## Configuration

### config.py

```python
import os
from dotenv import load_dotenv

class Config:
    """Base configuration"""
    DEBUG = False
    TESTING = False

    # Database
    DATABASE_URL = os.getenv(
        'DATABASE_URL',
        'postgresql://user:password@localhost/pganalytics'
    )
    DB_POOL_SIZE = int(os.getenv('DB_POOL_SIZE', 10))

    # ML Models
    MODEL_TYPE = os.getenv('MODEL_TYPE', 'linear_regression')
    LOOKBACK_DAYS = int(os.getenv('LOOKBACK_DAYS', 90))
    MIN_TRAINING_SAMPLES = int(os.getenv('MIN_TRAINING_SAMPLES', 100))

    # Async Jobs
    CELERY_BROKER = os.getenv('CELERY_BROKER', 'redis://localhost:6379/0')
    CELERY_BACKEND = os.getenv('CELERY_BACKEND', 'redis://localhost:6379/1')

    # Logging
    LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO')
    LOG_FORMAT = os.getenv(
        'LOG_FORMAT',
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    # API
    API_PORT = int(os.getenv('API_PORT', 8081))
    API_HOST = os.getenv('API_HOST', '0.0.0.0')

    # CORS
    CORS_ORIGINS = os.getenv('CORS_ORIGINS', 'http://localhost:8080').split(',')

class DevelopmentConfig(Config):
    """Development configuration"""
    DEBUG = True
    LOG_LEVEL = 'DEBUG'

class ProductionConfig(Config):
    """Production configuration"""
    DEBUG = False
    LOG_LEVEL = 'INFO'

class TestingConfig(Config):
    """Testing configuration"""
    TESTING = True
    DATABASE_URL = 'postgresql://user:password@localhost/pganalytics_test'
    CELERY_BROKER = 'memory://'
    CELERY_BACKEND = 'cache+memory://'
```

---

## Docker Deployment

### Dockerfile

```dockerfile
FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Set environment variables
ENV FLASK_APP=app.py
ENV PYTHONUNBUFFERED=1

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8081/health')"

# Run application
CMD ["gunicorn", "--bind", "0.0.0.0:8081", "--workers", "4", "app:app"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  ml-service:
    build: .
    ports:
      - "8081:8081"
    environment:
      DATABASE_URL: postgresql://pganalytics:password@postgres:5432/pganalytics
      CELERY_BROKER: redis://redis:6379/0
      CELERY_BACKEND: redis://redis:6379/1
      FLASK_ENV: development
      LOG_LEVEL: DEBUG
    depends_on:
      - postgres
      - redis
    volumes:
      - .:/app

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: pganalytics
      POSTGRES_PASSWORD: password
      POSTGRES_DB: pganalytics
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

## Testing Strategy

### Unit Tests

```python
# test_models.py
def test_feature_extraction():
    """Test feature extraction from query"""
    features = model.extract_features(query_hash=4001)
    assert len(features) == EXPECTED_FEATURE_COUNT
    assert all(isinstance(f, (int, float)) for f in features)

def test_model_training():
    """Test model training on sample data"""
    model = PerformanceModel('linear_regression')
    metrics = model.train(queries_lookback_days=30)
    assert metrics['r_squared'] > 0.5
    assert metrics['training_samples'] > 100

def test_prediction():
    """Test model prediction"""
    model = PerformanceModel.load_latest()
    prediction = model.predict(features)
    assert 'predicted_execution_time_ms' in prediction
    assert 'confidence_score' in prediction
    assert 0 <= prediction['confidence_score'] <= 1
```

### Integration Tests

```python
# test_api.py
def test_train_model_endpoint():
    """Test model training via API"""
    response = client.post('/api/train/performance-model', json={
        'database_name': 'pganalytics',
        'lookback_days': 30
    })
    assert response.status_code == 202
    assert 'job_id' in response.json

def test_predict_endpoint():
    """Test prediction via API"""
    response = client.post('/api/predict/query-execution', json={
        'query_hash': 4001,
        'parameters': {}
    })
    assert response.status_code == 200
    assert 'predicted_execution_time_ms' in response.json
    assert 'confidence_score' in response.json
```

---

## Success Criteria

✅ **Criteria 1**: Model training works on historical data
- Can train linear regression model
- Training completes successfully
- R² score > 0.70

✅ **Criteria 2**: Features extracted correctly
- 12+ features extracted per query
- Features normalized properly
- Missing values handled

✅ **Criteria 3**: Predictions generated with confidence
- Predictions within reasonable bounds
- Confidence scores 0.6-0.95
- Confidence intervals calculated

✅ **Criteria 4**: API endpoints functional
- Training endpoint accepts requests
- Prediction endpoint returns predictions
- Model management endpoints work

✅ **Criteria 5**: Models stored and retrieved
- Models serialized to database
- Models can be loaded and used
- Multiple model versions supported

✅ **Criteria 6**: Validation/learning loop
- Actual results recorded
- Accuracy scores calculated
- Model drift detected

✅ **Criteria 7**: Performance acceptable
- Training completes in < 5 minutes
- Predictions in < 500ms
- Memory usage < 500MB

✅ **Criteria 8**: Integration with Go backend
- Go backend can call prediction endpoint
- Fallback behavior when service unavailable
- Error handling working

---

## Dependencies

### requirements.txt

```
Flask==2.3.2
Flask-CORS==4.0.0
gunicorn==21.2.0
psycopg2-binary==2.9.6
python-dotenv==1.0.0
scikit-learn==1.2.2
numpy==1.24.3
pandas==2.0.3
joblib==1.2.0
requests==2.31.0
celery==5.3.1
redis==4.5.5
prometheus-client==0.17.1
```

---

## Monitoring & Logging

### Metrics Tracked

```python
# Prometheus metrics
model_training_duration_seconds = Histogram(
    'ml_model_training_duration_seconds',
    'Time taken to train model'
)

prediction_latency_ms = Histogram(
    'ml_prediction_latency_ms',
    'Prediction inference latency'
)

model_accuracy_r_squared = Gauge(
    'ml_model_accuracy_r_squared',
    'Model R² score'
)

prediction_error_percent = Histogram(
    'ml_prediction_error_percent',
    'Prediction error percentage'
)
```

### Logging Configuration

```python
logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('logs/ml_service.log'),
        logging.StreamHandler()
    ]
)
```

---

**Status**: Ready for implementation

