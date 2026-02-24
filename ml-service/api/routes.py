"""
API Routes for ML Service
Defines all REST API endpoints
"""

import logging
from flask import Blueprint, request, jsonify
from . import handlers

logger = logging.getLogger(__name__)

api_blueprint = Blueprint('api', __name__)


# Training Endpoints
@api_blueprint.route('/train/performance-model', methods=['POST'])
def train_performance_model():
    """
    POST /api/train/performance-model

    Trigger async model training on historical query data

    Request body:
    {
        "database_name": "pganalytics",
        "lookback_days": 90,
        "model_type": "linear_regression",
        "force_retrain": false
    }

    Response (202 Accepted):
    {
        "job_id": "train-20260220-001",
        "status": "training",
        "message": "Model training started in background"
    }
    """
    return handlers.handle_train_performance_model(request)


@api_blueprint.route('/train/performance-model/<job_id>', methods=['GET'])
def get_training_status(job_id):
    """
    GET /api/train/performance-model/{job_id}

    Get status of model training job

    Response (200 OK):
    {
        "job_id": "train-20260220-001",
        "status": "completed",
        "model_id": "model-linear-001",
        "r_squared": 0.78,
        "training_samples": 1500,
        "completed_at": "2026-02-20T..."
    }
    """
    return handlers.handle_get_training_status(job_id)


# Prediction Endpoints
@api_blueprint.route('/predict/query-execution', methods=['POST'])
def predict_query_execution():
    """
    POST /api/predict/query-execution

    Predict query execution time with confidence interval

    Request body:
    {
        "query_hash": 4001,
        "parameters": {"param1": "value1"},
        "scenario": "current"
    }

    Response (200 OK):
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
        "prediction_timestamp": "2026-02-20T..."
    }
    """
    return handlers.handle_predict_query_execution(request)


@api_blueprint.route('/validate/prediction', methods=['POST'])
def validate_prediction():
    """
    POST /api/validate/prediction

    Record actual query result and validate prediction accuracy

    Request body:
    {
        "prediction_id": "pred-20260220-001",
        "query_hash": 4001,
        "predicted_execution_time_ms": 125.5,
        "actual_execution_time_ms": 118.2,
        "model_version": "v1.2"
    }

    Response (200 OK):
    {
        "prediction_id": "pred-20260220-001",
        "error_percent": 6.2,
        "accuracy_score": 0.938,
        "within_confidence_interval": true,
        "message": "Prediction validation recorded"
    }
    """
    return handlers.handle_validate_prediction(request)


# Model Management Endpoints
@api_blueprint.route('/models/latest', methods=['GET'])
def get_latest_model():
    """
    GET /api/models/latest

    Get metadata for latest trained model

    Response (200 OK):
    {
        "model_id": "model-linear-001",
        "model_type": "linear_regression",
        "training_date": "2026-02-20T...",
        "r_squared": 0.78,
        "feature_count": 12,
        "is_active": true
    }
    """
    return handlers.handle_get_latest_model()


@api_blueprint.route('/models/<model_id>', methods=['GET'])
def get_model(model_id):
    """
    GET /api/models/{model_id}

    Get metadata for specific model version

    Response (200 OK):
    {
        "model_id": "model-linear-001",
        "model_type": "linear_regression",
        ...
    }
    """
    return handlers.handle_get_model(model_id)


@api_blueprint.route('/models', methods=['GET'])
def list_models():
    """
    GET /api/models

    List all trained model versions

    Response (200 OK):
    {
        "models": [
            {
                "model_id": "model-linear-001",
                "model_type": "linear_regression",
                "training_date": "2026-02-20T...",
                "r_squared": 0.78,
                "is_active": true
            }
        ],
        "total_models": 2,
        "active_model": "model-linear-001"
    }
    """
    return handlers.handle_list_models()


@api_blueprint.route('/models/<model_id>/activate', methods=['POST'])
def activate_model(model_id):
    """
    POST /api/models/{model_id}/activate

    Set a model as active for predictions

    Response (200 OK):
    {
        "model_id": "model-linear-001",
        "status": "activated",
        "message": "Model activated for predictions"
    }
    """
    return handlers.handle_activate_model(model_id)


# Analytics & Query Insights Endpoints
@api_blueprint.route('/analytics/slow-queries', methods=['GET'])
def get_slow_queries():
    """
    GET /api/analytics/slow-queries

    Get slowest queries exceeding threshold

    Query parameters:
    - threshold_ms: Execution time threshold (default: 1000)
    - limit: Max results to return (default: 20)

    Response (200 OK):
    {
        "slow_queries": [
            {
                "query_hash": 4001,
                "mean_execution_time_ms": 1500,
                "calls_per_minute": 10,
                "total_impact_ms": 15000
            }
        ],
        "count": 5
    }
    """
    return handlers.handle_get_slow_queries(request)


@api_blueprint.route('/analytics/frequent-queries', methods=['GET'])
def get_frequent_queries():
    """
    GET /api/analytics/frequent-queries

    Get most frequently executed queries

    Query parameters:
    - limit: Max results to return (default: 20)

    Response (200 OK):
    {
        "frequent_queries": [
            {
                "query_hash": 4001,
                "calls_per_minute": 100,
                "mean_execution_time_ms": 125,
                "total_impact_ms": 12500
            }
        ],
        "count": 5
    }
    """
    return handlers.handle_get_frequent_queries(request)


@api_blueprint.route('/analytics/database-health', methods=['GET'])
def get_database_health():
    """
    GET /api/analytics/database-health

    Get overall database health summary

    Response (200 OK):
    {
        "total_queries": 1500,
        "avg_execution_ms": 125.5,
        "max_execution_ms": 5000,
        "seq_scan_count": 45,
        "indexed_count": 1455
    }
    """
    return handlers.handle_get_database_health(request)


@api_blueprint.route('/analytics/query/<int:query_hash>', methods=['GET'])
def get_query_analytics(query_hash):
    """
    GET /api/analytics/query/{query_hash}

    Get detailed analytics for a specific query

    Response (200 OK):
    {
        "query_hash": 4001,
        "calls_per_minute": 100,
        "mean_execution_time_ms": 125,
        "max_execution_time_ms": 500,
        "index_count": 3,
        "scan_type": "Index Scan"
    }
    """
    return handlers.handle_get_query_analytics(query_hash)


# Health & Status Endpoints
@api_blueprint.route('/status', methods=['GET'])
def service_status():
    """
    GET /api/status

    Get ML service status and statistics

    Response (200 OK):
    {
        "service": "ml-service",
        "status": "healthy",
        "version": "1.0.0",
        "active_model": "model-linear-001",
        "total_predictions": 15234,
        "avg_prediction_accuracy": 0.87,
        "uptime_seconds": 86400
    }
    """
    return handlers.handle_service_status()


# Error handling for invalid endpoints
@api_blueprint.errorhandler(405)
def method_not_allowed(error):
    """Handle 405 Method Not Allowed"""
    logger.warning(f"Method not allowed: {error}")
    return jsonify({'error': 'Method not allowed'}), 405


@api_blueprint.errorhandler(415)
def unsupported_media_type(error):
    """Handle 415 Unsupported Media Type"""
    logger.warning(f"Unsupported media type: {error}")
    return jsonify({'error': 'Unsupported media type, use application/json'}), 415
