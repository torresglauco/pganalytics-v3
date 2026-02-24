"""
PerformanceModel: ML Model for Query Execution Time Prediction
"""

import logging
import uuid
import pickle
from datetime import datetime
from typing import Dict, Tuple, Optional, List
import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.linear_model import LinearRegression
from sklearn.tree import DecisionTreeRegressor
from sklearn.ensemble import RandomForestRegressor
from sklearn.metrics import r2_score, mean_squared_error, mean_absolute_error

logger = logging.getLogger(__name__)


class PerformanceModel:
    """
    Machine Learning Model for Query Execution Time Prediction

    Trains on historical query metrics and predicts execution time
    with confidence intervals for new queries.
    """

    # Feature names in consistent order
    FEATURE_NAMES = [
        'query_calls_per_hour',
        'mean_table_size_mb',
        'index_count',
        'has_seq_scan',
        'has_nested_loop',
        'subquery_depth',
        'concurrent_queries_avg',
        'available_memory_pct',
        'std_dev_calls',
        'peak_hour_calls',
        'table_row_count',
        'avg_row_width_bytes',
    ]

    MODEL_TYPES = {
        'linear_regression': LinearRegression,
        'decision_tree': DecisionTreeRegressor,
        'random_forest': RandomForestRegressor,
    }

    def __init__(self, model_type: str = 'linear_regression', model_name: Optional[str] = None):
        """
        Initialize performance model

        Args:
            model_type: 'linear_regression', 'decision_tree', or 'random_forest'
            model_name: Optional friendly name for the model
        """
        if model_type not in self.MODEL_TYPES:
            raise ValueError(f"Invalid model_type: {model_type}")

        self.model_type = model_type
        self.model_name = model_name or f"Q-Exec-Predictor-{datetime.now().strftime('%Y%m%d-%H%M%S')}"
        self.model = None
        self.scaler = StandardScaler()
        self.feature_names = self.FEATURE_NAMES.copy()
        self.r_squared = None
        self.rmse = None
        self.mae = None
        self.training_date = None
        self.training_samples = 0
        self.model_id = f"model-{model_type}-{str(uuid.uuid4())[:8]}"

        logger.info(f"Initialized PerformanceModel: {self.model_name} ({model_type})")

    def extract_features(self, query_metrics: Dict) -> np.ndarray:
        """
        Extract feature vector from query metrics

        Args:
            query_metrics: Dictionary with query characteristics from database

        Returns:
            numpy array of features in consistent order
        """
        features = []

        for feature_name in self.feature_names:
            value = query_metrics.get(feature_name, 0.0)
            # Handle missing values by using 0 or default
            if value is None:
                value = 0.0
            features.append(float(value))

        return np.array([features])

    def train(self, X_train: np.ndarray, y_train: np.ndarray,
              X_test: Optional[np.ndarray] = None,
              y_test: Optional[np.ndarray] = None) -> Dict:
        """
        Train the model on historical data

        Args:
            X_train: Feature matrix (n_samples, n_features)
            y_train: Target vector (execution times)
            X_test: Optional test features for evaluation
            y_test: Optional test targets for evaluation

        Returns:
            Dictionary with training metrics
        """
        logger.info(f"Training {self.model_type} model on {len(X_train)} samples")

        try:
            # Fit scaler on training data
            X_train_scaled = self.scaler.fit_transform(X_train)

            # Create and train model
            if self.model_type == 'linear_regression':
                self.model = LinearRegression()
            elif self.model_type == 'decision_tree':
                self.model = DecisionTreeRegressor(random_state=42, max_depth=15)
            elif self.model_type == 'random_forest':
                self.model = RandomForestRegressor(
                    n_estimators=100,
                    random_state=42,
                    n_jobs=-1,
                    max_depth=15
                )

            # Train model
            self.model.fit(X_train_scaled, y_train)

            # Evaluate on training set
            train_pred = self.model.predict(X_train_scaled)
            self.r_squared = r2_score(y_train, train_pred)
            self.rmse = np.sqrt(mean_squared_error(y_train, train_pred))
            self.mae = mean_absolute_error(y_train, train_pred)

            # Evaluate on test set if provided
            test_r_squared = None
            if X_test is not None and y_test is not None:
                X_test_scaled = self.scaler.transform(X_test)
                test_pred = self.model.predict(X_test_scaled)
                test_r_squared = r2_score(y_test, test_pred)

            self.training_date = datetime.now()
            self.training_samples = len(X_train)

            logger.info(f"Training complete. R²={self.r_squared:.4f}, RMSE={self.rmse:.2f}, MAE={self.mae:.2f}")

            return {
                'model_id': self.model_id,
                'model_type': self.model_type,
                'training_samples': self.training_samples,
                'r_squared': round(self.r_squared, 4),
                'rmse': round(self.rmse, 2),
                'mae': round(self.mae, 2),
                'test_r_squared': round(test_r_squared, 4) if test_r_squared else None,
                'training_date': self.training_date.isoformat(),
                'feature_count': len(self.feature_names)
            }

        except Exception as e:
            logger.error(f"Error during model training: {e}")
            raise

    def predict(self, X: np.ndarray, return_confidence: bool = True) -> Dict:
        """
        Predict execution time for query

        Args:
            X: Feature matrix (n_samples, n_features) or (n_features,) for single sample
            return_confidence: Whether to include confidence interval

        Returns:
            Dictionary with predictions and confidence intervals
        """
        if self.model is None:
            raise ValueError("Model not trained. Call train() first.")

        # Handle single sample
        if X.ndim == 1:
            X = X.reshape(1, -1)

        # Scale features
        X_scaled = self.scaler.transform(X)

        # Make predictions
        predictions = self.model.predict(X_scaled)

        result = {
            'predicted_execution_time_ms': float(predictions[0]),
            'model_version': self.model_name,
            'model_type': self.model_type,
            'prediction_timestamp': datetime.now().isoformat(),
        }

        if return_confidence:
            # Calculate confidence interval based on model accuracy
            confidence = self._calculate_confidence()
            std_dev = self.rmse / 2  # Rough estimate of prediction uncertainty

            result.update({
                'confidence_score': confidence,
                'confidence_interval': {
                    'lower_bound_ms': float(predictions[0] - (2 * std_dev)),
                    'upper_bound_ms': float(predictions[0] + (2 * std_dev)),
                    'std_dev_ms': float(std_dev),
                }
            })

        return result

    def _calculate_confidence(self) -> float:
        """
        Calculate prediction confidence score (0-1)

        Based on model R² score and training data quality
        """
        if self.r_squared is None:
            return 0.5

        # Confidence = R² score (model accuracy)
        # But cap at 0.95 to indicate some uncertainty always exists
        confidence = min(self.r_squared, 0.95)

        # Lower confidence if not many training samples
        if self.training_samples < 100:
            confidence *= 0.8

        return max(0.5, confidence)  # Minimum 0.5 confidence

    def evaluate(self, X_test: np.ndarray, y_test: np.ndarray) -> Dict:
        """
        Evaluate model performance on test data

        Args:
            X_test: Test feature matrix
            y_test: Test target values

        Returns:
            Dictionary with evaluation metrics
        """
        if self.model is None:
            raise ValueError("Model not trained.")

        X_test_scaled = self.scaler.transform(X_test)
        predictions = self.model.predict(X_test_scaled)

        r2 = r2_score(y_test, predictions)
        rmse = np.sqrt(mean_squared_error(y_test, predictions))
        mae = mean_absolute_error(y_test, predictions)

        errors = np.abs(predictions - y_test)
        error_percent = (errors / y_test) * 100

        return {
            'r_squared': round(r2, 4),
            'rmse': round(rmse, 2),
            'mae': round(mae, 2),
            'median_error_percent': round(np.median(error_percent), 2),
            'mean_error_percent': round(np.mean(error_percent), 2),
            'max_error_percent': round(np.max(error_percent), 2),
        }

    def save(self, filepath: str) -> None:
        """
        Save model to disk

        Args:
            filepath: Path to save model pickle
        """
        if self.model is None:
            raise ValueError("Model not trained.")

        try:
            with open(filepath, 'wb') as f:
                pickle.dump({
                    'model': self.model,
                    'scaler': self.scaler,
                    'model_type': self.model_type,
                    'model_name': self.model_name,
                    'model_id': self.model_id,
                    'r_squared': self.r_squared,
                    'rmse': self.rmse,
                    'mae': self.mae,
                    'training_date': self.training_date,
                    'training_samples': self.training_samples,
                    'feature_names': self.feature_names,
                }, f)
            logger.info(f"Model saved to {filepath}")
        except Exception as e:
            logger.error(f"Error saving model: {e}")
            raise

    @classmethod
    def load(cls, filepath: str) -> 'PerformanceModel':
        """
        Load model from disk

        Args:
            filepath: Path to saved model pickle

        Returns:
            PerformanceModel instance
        """
        try:
            with open(filepath, 'rb') as f:
                data = pickle.load(f)

            model = cls(model_type=data['model_type'], model_name=data['model_name'])
            model.model = data['model']
            model.scaler = data['scaler']
            model.model_id = data['model_id']
            model.r_squared = data['r_squared']
            model.rmse = data['rmse']
            model.mae = data['mae']
            model.training_date = data['training_date']
            model.training_samples = data['training_samples']
            model.feature_names = data['feature_names']

            logger.info(f"Model loaded from {filepath}")
            return model
        except Exception as e:
            logger.error(f"Error loading model: {e}")
            raise

    def __repr__(self) -> str:
        """String representation"""
        return (
            f"PerformanceModel(id={self.model_id}, type={self.model_type}, "
            f"r2={self.r_squared}, samples={self.training_samples})"
        )
