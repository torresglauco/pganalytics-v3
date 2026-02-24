"""
Feature engineering utilities for query metrics
"""

import logging
import numpy as np
from typing import Dict, List, Any, Optional

logger = logging.getLogger(__name__)


class FeatureEngineer:
    """
    Handles feature extraction and transformation from query metrics
    """

    # Standard feature names in consistent order
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

    @staticmethod
    def extract_from_metrics(query_metrics: Dict[str, Any]) -> np.ndarray:
        """
        Extract feature vector from query metrics dictionary

        Args:
            query_metrics: Dictionary containing query characteristics

        Returns:
            numpy array of features in consistent order (1 x n_features)
        """
        features = []

        for feature_name in FeatureEngineer.FEATURE_NAMES:
            value = query_metrics.get(feature_name, 0.0)

            # Handle None values
            if value is None:
                value = 0.0

            # Ensure numeric type
            try:
                value = float(value)
            except (ValueError, TypeError):
                value = 0.0

            features.append(value)

        return np.array([features])

    @staticmethod
    def validate_features(features: np.ndarray) -> bool:
        """
        Validate feature array for model input

        Args:
            features: Feature array to validate

        Returns:
            True if valid, False otherwise
        """
        if features is None:
            logger.warning("Features are None")
            return False

        if not isinstance(features, np.ndarray):
            logger.warning(f"Features not numpy array: {type(features)}")
            return False

        if features.shape[1] != len(FeatureEngineer.FEATURE_NAMES):
            logger.warning(f"Feature count mismatch: {features.shape[1]} != {len(FeatureEngineer.FEATURE_NAMES)}")
            return False

        # Check for NaN or Inf values
        if np.any(np.isnan(features)) or np.any(np.isinf(features)):
            logger.warning("Features contain NaN or Inf values")
            return False

        return True

    @staticmethod
    def handle_missing_values(features: np.ndarray, strategy: str = 'zero') -> np.ndarray:
        """
        Handle missing values in features

        Args:
            features: Feature array
            strategy: 'zero' (fill with 0), 'mean' (fill with feature mean)

        Returns:
            Cleaned feature array
        """
        features = features.copy()

        if strategy == 'zero':
            # Replace NaN with 0
            features[np.isnan(features)] = 0.0
        elif strategy == 'mean':
            # Replace NaN with column mean
            col_mean = np.nanmean(features, axis=0)
            inds = np.where(np.isnan(features))
            features[inds] = np.take(col_mean, inds[1])

        return features

    @staticmethod
    def clip_outliers(features: np.ndarray, percentile: float = 99.0) -> np.ndarray:
        """
        Clip outlier values to reduce their impact on predictions

        Args:
            features: Feature array
            percentile: Percentile to use as clip threshold

        Returns:
            Array with outliers clipped
        """
        features = features.copy()

        for col in range(features.shape[1]):
            upper_bound = np.percentile(features[:, col], percentile)
            features[:, col] = np.minimum(features[:, col], upper_bound)

        return features

    @staticmethod
    def get_feature_descriptions() -> Dict[str, str]:
        """
        Get descriptions of each feature for documentation

        Returns:
            Dictionary mapping feature names to descriptions
        """
        return {
            'query_calls_per_hour': 'Average number of query executions per hour',
            'mean_table_size_mb': 'Average size of tables accessed by query in MB',
            'index_count': 'Number of indexes available on tables',
            'has_seq_scan': 'Binary: whether query uses sequential scan (1=yes, 0=no)',
            'has_nested_loop': 'Binary: whether query uses nested loop join (1=yes, 0=no)',
            'subquery_depth': 'Nesting depth of subqueries (0=none)',
            'concurrent_queries_avg': 'Average number of concurrent queries during execution',
            'available_memory_pct': 'Percentage of available system memory',
            'std_dev_calls': 'Standard deviation of calls per minute',
            'peak_hour_calls': 'Maximum calls in any single hour',
            'table_row_count': 'Total number of rows in tables accessed',
            'avg_row_width_bytes': 'Average width of rows in bytes',
        }

    @staticmethod
    def normalize_features(features: np.ndarray, mean: Optional[np.ndarray] = None,
                          std: Optional[np.ndarray] = None) -> tuple:
        """
        Normalize features using Z-score normalization

        Args:
            features: Feature array
            mean: Pre-calculated mean (if None, calculated from features)
            std: Pre-calculated std dev (if None, calculated from features)

        Returns:
            Tuple of (normalized_features, mean, std)
        """
        if mean is None:
            mean = np.mean(features, axis=0)
        if std is None:
            std = np.std(features, axis=0)

        # Avoid division by zero
        std = np.where(std == 0, 1, std)

        normalized = (features - mean) / std
        return normalized, mean, std

    @staticmethod
    def create_feature_report(features: np.ndarray) -> Dict[str, Any]:
        """
        Create a statistical report of features

        Args:
            features: Feature array

        Returns:
            Dictionary with feature statistics
        """
        report = {
            'num_samples': features.shape[0],
            'num_features': features.shape[1],
            'feature_stats': {}
        }

        for i, name in enumerate(FeatureEngineer.FEATURE_NAMES):
            col = features[:, i]
            report['feature_stats'][name] = {
                'mean': float(np.mean(col)),
                'std': float(np.std(col)),
                'min': float(np.min(col)),
                'max': float(np.max(col)),
                'median': float(np.median(col)),
            }

        return report

    @staticmethod
    def log_feature_info(features: np.ndarray) -> None:
        """
        Log feature information for debugging

        Args:
            features: Feature array to log
        """
        report = FeatureEngineer.create_feature_report(features)
        logger.info(f"Feature shapes: {report['num_samples']} samples, {report['num_features']} features")

        for name, stats in report['feature_stats'].items():
            logger.debug(f"{name}: mean={stats['mean']:.2f}, std={stats['std']:.2f}, "
                        f"min={stats['min']:.2f}, max={stats['max']:.2f}")
