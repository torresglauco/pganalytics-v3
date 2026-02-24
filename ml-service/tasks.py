"""
Celery async tasks for ML service
"""

import logging
import time
from datetime import datetime
from typing import Dict, Any, Optional

import numpy as np
from celery import Celery, Task
from celery.exceptions import SoftTimeLimitExceeded

from config import Config
from models.performance_predictor import PerformanceModel
from utils.db_connection import DatabaseConnection
from utils.feature_engineer import FeatureEngineer

logger = logging.getLogger(__name__)

# Initialize Celery
celery_app = Celery(__name__)

# Configure Celery
celery_app.conf.update(
    broker_url=Config.CELERY_BROKER,
    result_backend=Config.CELERY_BACKEND,
    task_serializer='json',
    accept_content=['json'],
    result_serializer='json',
    timezone='UTC',
    enable_utc=True,
    task_track_started=True,
    task_time_limit=600,  # 10 minutes hard limit
    task_soft_time_limit=540,  # 9 minutes soft limit
    result_expires=3600,  # Results expire after 1 hour
)


class DatabaseTask(Task):
    """Base task class with database connection"""

    def on_retry(self, exc, task_id, args, kwargs, einfo):
        """Called when task is retried"""
        logger.warning(f"Task {task_id} retrying: {exc}")

    def on_failure(self, exc, task_id, args, kwargs, einfo):
        """Called when task fails"""
        logger.error(f"Task {task_id} failed: {exc}")

    def on_success(self, result, task_id, args, kwargs):
        """Called when task succeeds"""
        logger.info(f"Task {task_id} completed successfully")


celery_app.Task = DatabaseTask


@celery_app.task(bind=True, max_retries=3, default_retry_delay=60)
def train_performance_model(
    self,
    database_url: str,
    lookback_days: int = 90,
    model_type: str = 'random_forest',
    job_id: Optional[str] = None
) -> Dict[str, Any]:
    """
    Async task to train performance model

    Args:
        database_url: PostgreSQL connection string
        lookback_days: Days of historical data to use
        model_type: Type of model to train
        job_id: Job identifier for tracking

    Returns:
        Dictionary with training results
    """
    try:
        logger.info(f"Starting training job {job_id}: model_type={model_type}, lookback_days={lookback_days}")

        # Initialize database connection
        db = DatabaseConnection(database_url)
        if not db.initialize():
            logger.error(f"Failed to initialize database connection for job {job_id}")
            raise Exception("Database connection failed")

        try:
            # Extract training data
            logger.info(f"Extracting training data for job {job_id}")
            X_train, y_train = db.extract_training_data(lookback_days=lookback_days)

            if X_train is None or len(X_train) == 0:
                logger.warning(f"Insufficient training data for job {job_id}")
                raise Exception(f"Insufficient training data: need at least {Config.MIN_TRAINING_SAMPLES} samples")

            # Validate features
            if not FeatureEngineer.validate_features(X_train):
                logger.error(f"Feature validation failed for job {job_id}")
                raise Exception("Feature validation failed")

            # Handle missing values
            X_train = FeatureEngineer.handle_missing_values(X_train, strategy='zero')

            # Clip outliers
            X_train = FeatureEngineer.clip_outliers(X_train, percentile=99.0)

            logger.info(f"Training data ready: {X_train.shape[0]} samples, {X_train.shape[1]} features")

            # Create and train model
            logger.info(f"Training {model_type} model for job {job_id}")
            model = PerformanceModel(model_type=model_type, model_name=f"{job_id}-{model_type}")

            training_metrics = model.train(X_train, y_train)

            logger.info(f"Model training complete for job {job_id}: R²={training_metrics['r_squared']:.4f}")

            # Save model metadata to database
            logger.info(f"Saving model metadata for job {job_id}")
            success = db.save_model_metadata(
                model_id=model.model_id,
                model_type=model.model_type,
                model_name=model.model_name,
                r_squared=model.r_squared,
                rmse=model.rmse,
                mae=model.mae,
                training_samples=model.training_samples
            )

            if not success:
                logger.warning(f"Failed to save model metadata for job {job_id}")

            return {
                'job_id': job_id,
                'status': 'completed',
                'model_id': model.model_id,
                'model_type': model_type,
                'r_squared': training_metrics['r_squared'],
                'rmse': training_metrics['rmse'],
                'mae': training_metrics['mae'],
                'training_samples': training_metrics['training_samples'],
                'completed_at': datetime.utcnow().isoformat(),
            }

        finally:
            db.close()

    except SoftTimeLimitExceeded:
        logger.error(f"Training task {job_id} exceeded time limit")
        raise

    except Exception as e:
        logger.error(f"Training failed for job {job_id}: {e}")

        # Retry with exponential backoff
        try:
            raise self.retry(exc=e, countdown=60 * (2 ** self.request.retries))
        except Exception:
            return {
                'job_id': job_id,
                'status': 'failed',
                'error': str(e),
                'failed_at': datetime.utcnow().isoformat(),
            }


@celery_app.task(bind=True, max_retries=2, default_retry_delay=30)
def validate_prediction(
    self,
    database_url: str,
    prediction_id: str,
    query_hash: int,
    predicted_ms: float,
    actual_ms: float,
    model_version: str
) -> Dict[str, Any]:
    """
    Async task to validate and record prediction accuracy

    Args:
        database_url: PostgreSQL connection string
        prediction_id: Unique prediction identifier
        query_hash: Hash of the query
        predicted_ms: Predicted execution time
        actual_ms: Actual execution time
        model_version: Model version used for prediction

    Returns:
        Validation results with accuracy metrics
    """
    try:
        logger.info(f"Validating prediction {prediction_id} (query_hash={query_hash})")

        # Initialize database connection
        db = DatabaseConnection(database_url)
        if not db.initialize():
            raise Exception("Database connection failed")

        try:
            # Calculate error metrics
            error_ms = abs(predicted_ms - actual_ms)
            error_percent = (error_ms / actual_ms) * 100 if actual_ms > 0 else 0
            accuracy_score = max(0, 1 - (error_percent / 100))
            within_interval = error_percent <= 30  # 30% threshold

            # Record prediction validation
            logger.debug(f"Recording validation for {prediction_id}: error={error_percent:.2f}%, accuracy={accuracy_score:.2f}")

            result = {
                'prediction_id': prediction_id,
                'query_hash': query_hash,
                'predicted_ms': predicted_ms,
                'actual_ms': actual_ms,
                'error_ms': error_ms,
                'error_percent': round(error_percent, 2),
                'accuracy_score': round(accuracy_score, 3),
                'within_confidence_interval': within_interval,
                'model_version': model_version,
                'validated_at': datetime.utcnow().isoformat(),
            }

            logger.info(f"Prediction {prediction_id} validated: accuracy={accuracy_score:.2f}")
            return result

        finally:
            db.close()

    except SoftTimeLimitExceeded:
        logger.error(f"Validation task {prediction_id} exceeded time limit")
        raise

    except Exception as e:
        logger.error(f"Validation failed for {prediction_id}: {e}")

        # Retry once
        try:
            raise self.retry(exc=e, countdown=30 * (2 ** self.request.retries))
        except Exception:
            return {
                'prediction_id': prediction_id,
                'status': 'failed',
                'error': str(e),
            }


@celery_app.task(bind=True, max_retries=2)
def collect_prediction_metrics(
    self,
    database_url: str,
    job_id: str
) -> Dict[str, Any]:
    """
    Async task to collect and aggregate prediction metrics

    Args:
        database_url: PostgreSQL connection string
        job_id: Job identifier

    Returns:
        Aggregated metrics
    """
    try:
        logger.info(f"Collecting prediction metrics for job {job_id}")

        db = DatabaseConnection(database_url)
        if not db.initialize():
            raise Exception("Database connection failed")

        try:
            # In a full implementation, this would query prediction validation table
            # and calculate aggregate metrics
            metrics = {
                'job_id': job_id,
                'total_predictions': 0,
                'mean_accuracy': 0.0,
                'median_error_percent': 0.0,
                'p95_error_percent': 0.0,
                'collected_at': datetime.utcnow().isoformat(),
            }

            logger.info(f"Metrics collected for job {job_id}")
            return metrics

        finally:
            db.close()

    except Exception as e:
        logger.error(f"Metric collection failed for {job_id}: {e}")
        try:
            raise self.retry(exc=e, countdown=60)
        except Exception:
            return {'job_id': job_id, 'status': 'failed', 'error': str(e)}


@celery_app.task
def cleanup_old_models(database_url: str, keep_count: int = 5) -> Dict[str, Any]:
    """
    Async task to clean up old model versions

    Args:
        database_url: PostgreSQL connection string
        keep_count: Number of recent models to keep

    Returns:
        Cleanup results
    """
    try:
        logger.info(f"Cleaning up old models, keeping {keep_count} recent versions")

        db = DatabaseConnection(database_url)
        if not db.initialize():
            raise Exception("Database connection failed")

        try:
            # In a full implementation, this would delete old models from database
            # keeping only the most recent versions
            return {
                'status': 'success',
                'models_deleted': 0,
                'cleaned_at': datetime.utcnow().isoformat(),
            }

        finally:
            db.close()

    except Exception as e:
        logger.error(f"Model cleanup failed: {e}")
        return {'status': 'failed', 'error': str(e)}


@celery_app.task
def health_check() -> Dict[str, Any]:
    """
    Task to verify Celery worker is healthy

    Returns:
        Health status
    """
    return {
        'status': 'healthy',
        'timestamp': datetime.utcnow().isoformat(),
    }
