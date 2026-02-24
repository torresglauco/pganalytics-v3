"""
Integration tests for ML Service API
"""

import pytest
import json
import numpy as np
from pathlib import Path

# Add parent directory to path for imports
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from app import create_app
from config import Config


@pytest.fixture
def app():
    """Create Flask app for testing"""
    test_config = Config()
    test_config.TESTING = True
    app = create_app(test_config)
    return app


@pytest.fixture
def client(app):
    """Create test client"""
    return app.test_client()


class TestHealthCheck:
    """Tests for health check endpoint"""

    def test_health_check(self, client):
        """Test /health endpoint"""
        response = client.get('/health')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'healthy'
        assert data['service'] == 'ml-service'


class TestServiceStatus:
    """Tests for service status endpoint"""

    def test_service_status(self, client):
        """Test /api/status endpoint"""
        response = client.get('/api/status')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'service' in data
        assert 'status' in data
        assert 'version' in data


class TestTrainingEndpoint:
    """Tests for model training endpoint"""

    def test_train_performance_model(self, client):
        """Test POST /api/train/performance-model"""
        request_data = {
            'database_name': 'pganalytics',
            'lookback_days': 30,
            'model_type': 'linear_regression',
            'force_retrain': False
        }

        response = client.post(
            '/api/train/performance-model',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 202
        data = json.loads(response.data)
        assert 'job_id' in data
        assert data['status'] == 'training'
        assert 'message' in data

    def test_train_with_invalid_model_type(self, client):
        """Test training with invalid model type"""
        request_data = {
            'model_type': 'invalid_model'
        }

        response = client.post(
            '/api/train/performance-model',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400
        data = json.loads(response.data)
        assert 'error' in data

    def test_train_missing_content_type(self, client):
        """Test training without content type"""
        response = client.post('/api/train/performance-model')

        assert response.status_code == 415

    def test_get_training_status(self, client):
        """Test GET /api/train/performance-model/{job_id}"""
        job_id = 'train-20260220-001'

        response = client.get(f'/api/train/performance-model/{job_id}')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['job_id'] == job_id
        assert 'status' in data


class TestPredictionEndpoint:
    """Tests for prediction endpoint"""

    def test_predict_query_execution(self, client):
        """Test POST /api/predict/query-execution"""
        request_data = {
            'query_hash': 4001,
            'parameters': {}
        }

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'predicted_execution_time_ms' in data
        assert 'confidence_score' in data
        assert 'confidence_interval' in data
        assert data['confidence_interval']['lower_bound_ms'] < data['confidence_interval']['upper_bound_ms']

    def test_predict_missing_query_hash(self, client):
        """Test prediction with missing query_hash"""
        request_data = {'parameters': {}}

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400

    def test_predict_with_parameters(self, client):
        """Test prediction with parameter values"""
        request_data = {
            'query_hash': 4001,
            'parameters': {'limit': 1000},
            'scenario': 'optimized'
        }

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'query_hash' in data
        assert data['query_hash'] == 4001


class TestValidationEndpoint:
    """Tests for prediction validation endpoint"""

    def test_validate_prediction(self, client):
        """Test POST /api/validate/prediction"""
        request_data = {
            'prediction_id': 'pred-001',
            'query_hash': 4001,
            'predicted_execution_time_ms': 125.5,
            'actual_execution_time_ms': 118.2,
            'model_version': 'v1.2'
        }

        response = client.post(
            '/api/validate/prediction',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'prediction_id' in data
        assert 'error_percent' in data
        assert 'accuracy_score' in data

    def test_validate_prediction_missing_fields(self, client):
        """Test validation with missing fields"""
        request_data = {
            'prediction_id': 'pred-001'
        }

        response = client.post(
            '/api/validate/prediction',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400


class TestModelEndpoints:
    """Tests for model management endpoints"""

    def test_get_latest_model(self, client):
        """Test GET /api/models/latest"""
        response = client.get('/api/models/latest')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'model_id' in data
        assert 'model_type' in data
        assert 'r_squared' in data

    def test_get_specific_model(self, client):
        """Test GET /api/models/{model_id}"""
        model_id = 'model-linear-001'

        response = client.get(f'/api/models/{model_id}')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['model_id'] == model_id

    def test_list_all_models(self, client):
        """Test GET /api/models"""
        response = client.get('/api/models')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'models' in data
        assert isinstance(data['models'], list)
        assert 'total_models' in data
        assert 'active_model' in data

    def test_activate_model(self, client):
        """Test POST /api/models/{model_id}/activate"""
        model_id = 'model-linear-001'

        response = client.post(f'/api/models/{model_id}/activate')

        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'activated'


class TestErrorHandling:
    """Tests for error handling"""

    def test_invalid_json(self, client):
        """Test handling of invalid JSON"""
        response = client.post(
            '/api/predict/query-execution',
            data='invalid json',
            content_type='application/json'
        )

        assert response.status_code == 400
        data = json.loads(response.data)
        assert 'error' in data

    def test_invalid_content_type(self, client):
        """Test handling of unsupported content type"""
        response = client.post(
            '/api/train/performance-model',
            data='some data',
            content_type='text/plain'
        )

        assert response.status_code == 415

    def test_method_not_allowed(self, client):
        """Test method not allowed (GET on POST-only endpoint)"""
        response = client.get('/api/train/performance-model')

        assert response.status_code == 405

    def test_not_found(self, client):
        """Test 404 for non-existent endpoint"""
        response = client.get('/api/nonexistent')

        assert response.status_code == 404


class TestRequestValidation:
    """Tests for request validation"""

    def test_predict_query_hash_type(self, client):
        """Test query_hash must be integer"""
        request_data = {
            'query_hash': 'not_an_int',
            'parameters': {}
        }

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400

    def test_predict_negative_query_hash(self, client):
        """Test query_hash must be positive"""
        request_data = {
            'query_hash': -1,
            'parameters': {}
        }

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400

    def test_lookback_days_positive(self, client):
        """Test lookback_days must be positive"""
        request_data = {
            'lookback_days': -1
        }

        response = client.post(
            '/api/train/performance-model',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        assert response.status_code == 400


class TestResponseFormats:
    """Tests for response format consistency"""

    def test_prediction_response_format(self, client):
        """Test prediction response has required fields"""
        request_data = {
            'query_hash': 4001,
            'parameters': {}
        }

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        data = json.loads(response.data)

        # Check required fields
        assert isinstance(data.get('query_hash'), int)
        assert isinstance(data.get('predicted_execution_time_ms'), (int, float))
        assert 0 <= data.get('confidence_score', 0) <= 1
        assert isinstance(data.get('model_version'), str)
        assert isinstance(data.get('prediction_timestamp'), str)

        # Check confidence interval structure
        ci = data.get('confidence_interval', {})
        assert isinstance(ci.get('lower_bound_ms'), (int, float))
        assert isinstance(ci.get('upper_bound_ms'), (int, float))
        assert isinstance(ci.get('std_dev_ms'), (int, float))

    def test_error_response_format(self, client):
        """Test error responses have consistent format"""
        request_data = {'parameters': {}}

        response = client.post(
            '/api/predict/query-execution',
            data=json.dumps(request_data),
            content_type='application/json'
        )

        data = json.loads(response.data)
        assert 'error' in data
        assert isinstance(data['error'], str)
