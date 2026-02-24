# ML Service for pganalytics

Python-based machine learning microservice for query execution time prediction and optimization recommendations.

## Architecture

The ML Service is a Flask-based REST API that provides:
- Model training on historical query metrics
- Query execution time predictions with confidence intervals
- Model versioning and management
- Integration with pganalytics PostgreSQL database

## Features

- **Performance Prediction**: Predict query execution time using scikit-learn models
- **Multiple Model Types**: Support for linear regression, decision trees, and random forests
- **Confidence Scoring**: Return prediction uncertainty alongside predictions
- **Model Versioning**: Store and activate different model versions
- **Async Training**: Background model training with job status tracking

## Project Structure

```
ml-service/
├── app.py                          # Flask application factory
├── config.py                       # Configuration management
├── requirements.txt                # Python dependencies
├── Dockerfile                      # Container image
├── docker-compose.yml              # Local development environment
├── api/
│   ├── __init__.py
│   ├── routes.py                   # API endpoint definitions
│   └── handlers.py                 # Endpoint implementation
├── models/
│   ├── __init__.py
│   └── performance_predictor.py    # Main ML model class
├── utils/
│   ├── __init__.py
│   ├── db_connection.py            # Database utilities
│   └── feature_engineer.py         # Feature extraction
├── tests/
│   ├── __init__.py
│   ├── test_models.py              # Model unit tests
│   └── test_api.py                 # API integration tests
└── README.md                       # This file
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Navigate to ml-service directory
cd ml-service

# Start all services (API, PostgreSQL, Redis)
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f ml-service

# Stop services
docker-compose down
```

### Local Development

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export FLASK_APP=app.py
export FLASK_ENV=development
export DATABASE_URL=postgresql://pganalytics:password@localhost:5432/pganalytics

# Run development server
python -m flask run --host=0.0.0.0 --port=8081

# Run tests
pytest tests/ -v

# With coverage
pytest tests/ -v --cov=. --cov-report=html
```

## API Endpoints

### Health Check
```bash
GET /health
# Returns: {"status": "healthy", "service": "ml-service"}
```

### Service Status
```bash
GET /api/status
# Returns service health and statistics
```

### Model Training
```bash
# Start async training
POST /api/train/performance-model
{
  "database_name": "pganalytics",
  "lookback_days": 30,
  "model_type": "linear_regression",  # or "decision_tree", "random_forest"
  "force_retrain": false
}
# Returns: {"job_id": "...", "status": "training"}

# Check training status
GET /api/train/performance-model/{job_id}
# Returns: {"job_id": "...", "status": "...", "r_squared": 0.78}
```

### Predictions
```bash
# Make prediction
POST /api/predict/query-execution
{
  "query_hash": 4001,
  "parameters": {},
  "scenario": "current"  # or "optimized"
}
# Returns: {
#   "query_hash": 4001,
#   "predicted_execution_time_ms": 125.5,
#   "confidence_score": 0.87,
#   "confidence_interval": {
#     "lower_bound_ms": 95.3,
#     "upper_bound_ms": 155.7,
#     "std_dev_ms": 15.2
#   }
# }

# Record actual execution time (for accuracy tracking)
POST /api/validate/prediction
{
  "prediction_id": "pred-001",
  "query_hash": 4001,
  "predicted_execution_time_ms": 125.5,
  "actual_execution_time_ms": 118.2,
  "model_version": "v1.2"
}
# Returns: {
#   "prediction_id": "pred-001",
#   "error_percent": 6.2,
#   "accuracy_score": 0.938,
#   "within_confidence_interval": true
# }
```

### Model Management
```bash
# Get active model
GET /api/models/latest
# Returns: Model metadata

# Get specific model
GET /api/models/{model_id}
# Returns: Model metadata

# List all models
GET /api/models
# Returns: {"models": [...], "total_models": 2, "active_model": "..."}

# Activate model
POST /api/models/{model_id}/activate
# Returns: {"model_id": "...", "status": "activated"}
```

## Configuration

Configuration is managed through environment variables. See `config.py` for all available options.

### Common Variables
- `FLASK_ENV`: development, production, testing
- `DATABASE_URL`: PostgreSQL connection string
- `CELERY_BROKER`: Redis connection for async tasks
- `CELERY_BACKEND`: Redis connection for task results
- `LOG_LEVEL`: DEBUG, INFO, WARNING, ERROR

### Example `.env`
```bash
FLASK_ENV=development
DATABASE_URL=postgresql://pganalytics:password@postgres:5432/pganalytics
CELERY_BROKER=redis://redis:6379/0
CELERY_BACKEND=redis://redis:6379/1
LOG_LEVEL=INFO
```

## Machine Learning Models

### Supported Models

1. **Linear Regression**
   - Best for: Linear relationships between features and execution time
   - Training time: Fast
   - Interpretability: High
   - Use case: Baseline model, stable predictions

2. **Decision Tree**
   - Best for: Non-linear patterns, feature interactions
   - Training time: Medium
   - Interpretability: Medium
   - Use case: Feature importance analysis

3. **Random Forest**
   - Best for: Complex patterns, robust predictions
   - Training time: Slow
   - Interpretability: Low
   - Use case: Production model, highest accuracy

### Features Used

The models use 12 features extracted from query metrics:

1. `query_calls_per_hour` - Query execution frequency
2. `mean_table_size_mb` - Average table size
3. `index_count` - Number of available indexes
4. `has_seq_scan` - Sequential scan indicator (0/1)
5. `has_nested_loop` - Nested loop join indicator (0/1)
6. `subquery_depth` - Nesting depth of subqueries
7. `concurrent_queries_avg` - Average concurrent executions
8. `available_memory_pct` - Available system memory percentage
9. `std_dev_calls` - Standard deviation of call frequency
10. `peak_hour_calls` - Maximum calls in single hour
11. `table_row_count` - Total rows in accessed tables
12. `avg_row_width_bytes` - Average row width

### Model Training

```python
from models.performance_predictor import PerformanceModel
from utils.db_connection import DatabaseConnection

# Connect to database
db = DatabaseConnection(database_url)
db.initialize()

# Extract training data
X_train, y_train = db.extract_training_data(lookback_days=90)

# Train model
model = PerformanceModel('random_forest')
metrics = model.train(X_train, y_train)

# Save model
model.save('models/trained/rf_model.pkl')
```

### Making Predictions

```python
from models.performance_predictor import PerformanceModel

# Load model
model = PerformanceModel.load('models/trained/rf_model.pkl')

# Extract features for a query
from utils.feature_engineer import FeatureEngineer
query_metrics = {
    'query_calls_per_hour': 100,
    'mean_table_size_mb': 512,
    # ... other features
}
features = FeatureEngineer.extract_from_metrics(query_metrics)

# Make prediction
prediction = model.predict(features, return_confidence=True)
print(f"Predicted execution time: {prediction['predicted_execution_time_ms']:.2f}ms")
print(f"Confidence: {prediction['confidence_score']:.2f}")
```

## Testing

### Run All Tests
```bash
pytest tests/ -v
```

### Run Specific Test Class
```bash
pytest tests/test_models.py::TestPerformanceModel -v
```

### Run with Coverage
```bash
pytest tests/ -v --cov=. --cov-report=html
```

### Test Categories

- **Unit Tests** (`test_models.py`):
  - PerformanceModel class functionality
  - FeatureEngineer utilities
  - Model training and prediction
  - Feature extraction and validation

- **Integration Tests** (`test_api.py`):
  - API endpoint functionality
  - Request/response validation
  - Error handling
  - Response format consistency

## Performance Optimization

### Caching
- Models are loaded once and cached in memory
- Feature extraction queries are optimized with indexes

### Async Training
- Model training happens asynchronously via Celery
- Long-running jobs don't block API requests
- Check job status with `GET /api/train/performance-model/{job_id}`

### Prediction Latency
- Target: <500ms per prediction
- Uses pre-scaled features and vectorized NumPy operations
- Connection pooling reduces database latency

## Monitoring

### Health Checks
```bash
# Service health
curl http://localhost:8081/health

# Detailed status
curl http://localhost:8081/api/status
```

### Logs
```bash
# Real-time logs
docker-compose logs -f ml-service

# View specific lines
docker-compose logs --tail=50 ml-service
```

### Metrics
The service exposes Prometheus metrics at `/metrics` (when prometheus-client is enabled).

## Database Integration

### Feature Extraction
The service queries `metrics_pg_stats_query` table for historical metrics:
- Filters by query hash for specific query features
- Aggregates metrics over lookback period
- Handles missing values automatically

### Model Storage
Models are persisted in pickle format, with metadata stored in `query_performance_models` table:
- Model type, name, and ID
- Training metrics (R², RMSE, MAE)
- Feature names and training sample count
- Creation and update timestamps

## Troubleshooting

### Connection to PostgreSQL Failed
```
Error: could not connect to server: Connection refused
```
**Solution**: Ensure PostgreSQL is running and DATABASE_URL is correct.
```bash
docker-compose ps  # Check if postgres service is healthy
```

### Model Training Fails
```
Error: Insufficient training samples: 50 < 100
```
**Solution**: Need at least 100 historical query executions. Wait for more data or reduce `min_samples` parameter.

### Predictions Return Same Value
**Possible Causes**:
- Model not trained (using mock response)
- Model needs retraining (data distribution changed)
- Feature values are constant

**Solution**: Retrain model with latest data.

### Port Already in Use
```bash
# Find process using port 8081
lsof -i :8081
# Kill process
kill -9 <PID>
```

## Development Guidelines

### Adding New Features
1. Create model method in `performance_predictor.py`
2. Create corresponding handler in `api/handlers.py`
3. Add route in `api/routes.py`
4. Write tests in `tests/`
5. Update API documentation

### Code Style
- Follow PEP 8
- Use type hints in function signatures
- Document public methods with docstrings
- Test coverage target: >80%

## Dependencies

See `requirements.txt` for complete list. Key packages:

- **Flask** (2.3.2) - Web framework
- **scikit-learn** (1.2.2) - ML algorithms
- **NumPy/Pandas** - Data manipulation
- **psycopg2** - PostgreSQL driver
- **Celery/Redis** - Async task queue
- **pytest** - Testing framework

## Performance Characteristics

| Operation | Typical Time |
|-----------|-------------|
| Prediction (cached model) | 50-100ms |
| Model training (1000 samples) | 2-5 seconds |
| Feature extraction | 100-500ms |
| Model loading from pickle | 500ms-1s |

## Future Enhancements

- [ ] XGBoost integration for advanced models
- [ ] Real-time model performance monitoring
- [ ] Automatic model retraining on schedule
- [ ] SHAP values for prediction explanations
- [ ] A/B testing framework for model comparison
- [ ] GPU support for large-scale training

## License

See repository LICENSE file

## Support

For issues or questions:
1. Check logs: `docker-compose logs ml-service`
2. Review troubleshooting section above
3. File issue in repository with full error trace
