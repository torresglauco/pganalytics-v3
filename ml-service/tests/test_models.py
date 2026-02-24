"""
Unit tests for ML models
"""

import pytest
import numpy as np
import tempfile
import os
from pathlib import Path

# Add parent directory to path for imports
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from models.performance_predictor import PerformanceModel
from utils.feature_engineer import FeatureEngineer


class TestPerformanceModel:
    """Test cases for PerformanceModel class"""

    @pytest.fixture
    def sample_metrics(self):
        """Fixture providing sample query metrics"""
        return {
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

    @pytest.fixture
    def synthetic_data(self):
        """Fixture providing synthetic training data"""
        n_samples = 200
        n_features = 12
        X_train = np.random.randn(n_samples, n_features)
        # Y = 50 + 5*X1 + 3*X2 + noise
        y_train = 50 + 5*X_train[:, 0] + 3*X_train[:, 1] + np.random.randn(n_samples) * 5
        return X_train, y_train

    def test_model_initialization(self):
        """Test PerformanceModel initialization"""
        model = PerformanceModel('linear_regression')

        assert model.model_type == 'linear_regression'
        assert model.model_name is not None
        assert model.model is None  # Not trained yet
        assert model.r_squared is None
        assert model.rmse is None

    def test_invalid_model_type(self):
        """Test initialization with invalid model type"""
        with pytest.raises(ValueError, match="Invalid model_type"):
            PerformanceModel('invalid_model')

    def test_feature_extraction(self, sample_metrics):
        """Test feature extraction from query metrics"""
        model = PerformanceModel('linear_regression')
        features = model.extract_features(sample_metrics)

        assert features.shape == (1, 12)
        assert all(isinstance(f, (int, float, np.number)) for f in features[0])

    def test_feature_extraction_missing_values(self):
        """Test feature extraction handles missing values"""
        model = PerformanceModel('linear_regression')
        incomplete_metrics = {
            'query_calls_per_hour': 100,
            'mean_table_size_mb': 512,
            # Missing other fields
        }

        features = model.extract_features(incomplete_metrics)

        assert features.shape == (1, 12)
        # Missing values should be 0.0
        assert features[0, 2] == 0.0  # index_count missing

    def test_model_training_linear_regression(self, synthetic_data):
        """Test model training with linear regression"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data

        metrics = model.train(X_train, y_train)

        assert 'model_id' in metrics
        assert 'r_squared' in metrics
        assert 'rmse' in metrics
        assert 'mae' in metrics
        assert metrics['model_type'] == 'linear_regression'
        assert metrics['r_squared'] > 0.5  # Synthetic data should fit well
        assert model.model is not None

    def test_model_training_decision_tree(self, synthetic_data):
        """Test model training with decision tree"""
        model = PerformanceModel('decision_tree')
        X_train, y_train = synthetic_data

        metrics = model.train(X_train, y_train)

        assert metrics['model_type'] == 'decision_tree'
        assert 'r_squared' in metrics
        assert model.model is not None

    def test_model_training_random_forest(self, synthetic_data):
        """Test model training with random forest"""
        model = PerformanceModel('random_forest')
        X_train, y_train = synthetic_data

        metrics = model.train(X_train, y_train)

        assert metrics['model_type'] == 'random_forest'
        assert 'r_squared' in metrics
        assert model.model is not None

    def test_model_training_with_test_set(self, synthetic_data):
        """Test model training with test set evaluation"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data

        # Split data
        X_test = X_train[:50]
        y_test = y_train[:50]
        X_train = X_train[50:]
        y_train = y_train[50:]

        metrics = model.train(X_train, y_train, X_test, y_test)

        assert 'test_r_squared' in metrics
        assert metrics['test_r_squared'] is not None

    def test_prediction_requires_training(self, sample_metrics):
        """Test that prediction fails if model not trained"""
        model = PerformanceModel('linear_regression')
        features = model.extract_features(sample_metrics)

        with pytest.raises(ValueError, match="Model not trained"):
            model.predict(features)

    def test_prediction_with_confidence(self, synthetic_data, sample_metrics):
        """Test prediction with confidence interval"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data
        model.train(X_train, y_train)

        features = model.extract_features(sample_metrics)
        prediction = model.predict(features, return_confidence=True)

        assert 'predicted_execution_time_ms' in prediction
        assert 'confidence_score' in prediction
        assert 'confidence_interval' in prediction
        assert prediction['predicted_execution_time_ms'] > 0
        assert 0 <= prediction['confidence_score'] <= 1
        assert 'lower_bound_ms' in prediction['confidence_interval']
        assert 'upper_bound_ms' in prediction['confidence_interval']

    def test_prediction_without_confidence(self, synthetic_data, sample_metrics):
        """Test prediction without confidence interval"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data
        model.train(X_train, y_train)

        features = model.extract_features(sample_metrics)
        prediction = model.predict(features, return_confidence=False)

        assert 'predicted_execution_time_ms' in prediction
        assert 'confidence_score' not in prediction
        assert 'confidence_interval' not in prediction

    def test_model_evaluation(self, synthetic_data):
        """Test model evaluation on test data"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data

        # Split data
        X_test = X_train[:50]
        y_test = y_train[:50]
        X_train = X_train[50:]
        y_train = y_train[50:]

        model.train(X_train, y_train)
        eval_metrics = model.evaluate(X_test, y_test)

        assert 'r_squared' in eval_metrics
        assert 'rmse' in eval_metrics
        assert 'mae' in eval_metrics
        assert 'mean_error_percent' in eval_metrics
        assert all(isinstance(v, (int, float)) for v in eval_metrics.values())

    def test_model_serialization(self, synthetic_data):
        """Test saving and loading model"""
        model = PerformanceModel('linear_regression', 'test-model')
        X_train, y_train = synthetic_data
        model.train(X_train, y_train)

        with tempfile.NamedTemporaryFile(suffix='.pkl', delete=False) as f:
            filepath = f.name

        try:
            # Save model
            model.save(filepath)
            assert os.path.exists(filepath)

            # Load model
            loaded_model = PerformanceModel.load(filepath)

            assert loaded_model.model_name == 'test-model'
            assert loaded_model.model_type == model.model_type
            assert loaded_model.r_squared == model.r_squared
            assert loaded_model.rmse == model.rmse

            # Verify loaded model can predict
            features = np.random.randn(1, 12)
            prediction = loaded_model.predict(features)
            assert 'predicted_execution_time_ms' in prediction
        finally:
            os.unlink(filepath)

    def test_confidence_calculation(self, synthetic_data):
        """Test confidence score calculation"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data
        model.train(X_train, y_train)

        confidence = model._calculate_confidence()

        assert 0.5 <= confidence <= 0.95
        # With good training data, confidence should be reasonably high
        assert confidence >= 0.7

    def test_confidence_with_few_samples(self):
        """Test confidence score with limited training samples"""
        model = PerformanceModel('linear_regression')

        # Train with very few samples
        X_train = np.random.randn(50, 12)
        y_train = 50 + 5*X_train[:, 0] + np.random.randn(50) * 5

        model.train(X_train, y_train)
        confidence = model._calculate_confidence()

        # With fewer samples, confidence should be reduced
        assert confidence < 0.8

    def test_model_representation(self, synthetic_data):
        """Test model string representation"""
        model = PerformanceModel('linear_regression')
        X_train, y_train = synthetic_data
        model.train(X_train, y_train)

        repr_str = repr(model)

        assert 'PerformanceModel' in repr_str
        assert 'linear_regression' in repr_str
        assert model.model_id in repr_str


class TestFeatureEngineer:
    """Test cases for FeatureEngineer utility"""

    @pytest.fixture
    def sample_metrics(self):
        """Fixture providing sample query metrics"""
        return {
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

    def test_extract_from_metrics(self, sample_metrics):
        """Test feature extraction"""
        features = FeatureEngineer.extract_from_metrics(sample_metrics)

        assert features.shape == (1, 12)
        assert features[0, 0] == 100.0  # query_calls_per_hour

    def test_validate_features(self):
        """Test feature validation"""
        valid_features = np.random.randn(1, 12)
        assert FeatureEngineer.validate_features(valid_features) is True

        # Invalid: wrong shape
        invalid_features = np.random.randn(1, 10)
        assert FeatureEngineer.validate_features(invalid_features) is False

        # Invalid: contains NaN
        nan_features = np.full((1, 12), np.nan)
        assert FeatureEngineer.validate_features(nan_features) is False

    def test_handle_missing_values_zero(self):
        """Test missing value handling with zero strategy"""
        features = np.array([[1, 2, np.nan, 4, 5, 6, 7, 8, 9, 10, 11, 12]], dtype=float)
        cleaned = FeatureEngineer.handle_missing_values(features, strategy='zero')

        assert not np.any(np.isnan(cleaned))
        assert cleaned[0, 2] == 0.0

    def test_get_feature_descriptions(self):
        """Test feature descriptions"""
        descriptions = FeatureEngineer.get_feature_descriptions()

        assert len(descriptions) == 12
        assert 'query_calls_per_hour' in descriptions
        assert isinstance(descriptions['query_calls_per_hour'], str)

    def test_normalize_features(self):
        """Test feature normalization"""
        features = np.array([
            [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
            [2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24]
        ], dtype=float)

        normalized, mean, std = FeatureEngineer.normalize_features(features)

        # Check that mean of normalized data is ~0 and std is ~1
        assert np.allclose(np.mean(normalized, axis=0), 0, atol=1e-10)
        assert np.allclose(np.std(normalized, axis=0), 1, atol=1e-10)

    def test_create_feature_report(self):
        """Test feature statistics reporting"""
        features = np.random.randn(100, 12) * 100 + 50
        report = FeatureEngineer.create_feature_report(features)

        assert report['num_samples'] == 100
        assert report['num_features'] == 12
        assert len(report['feature_stats']) == 12

        # Check first feature stats
        first_feature_stats = report['feature_stats']['query_calls_per_hour']
        assert 'mean' in first_feature_stats
        assert 'std' in first_feature_stats
        assert 'min' in first_feature_stats
        assert 'max' in first_feature_stats
        assert 'median' in first_feature_stats

    def test_clip_outliers(self):
        """Test outlier clipping"""
        features = np.array([[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 1000]], dtype=float)
        clipped = FeatureEngineer.clip_outliers(features, percentile=95)

        # Last value should be clipped to a lower value
        assert clipped[0, -1] < 1000
