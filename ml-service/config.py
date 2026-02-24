"""
Configuration management for ML Service
"""

import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    """Base configuration"""

    # Flask settings
    DEBUG = False
    TESTING = False
    SECRET_KEY = os.getenv('SECRET_KEY', 'dev-secret-key-change-in-production')

    # Database configuration
    DATABASE_URL = os.getenv(
        'DATABASE_URL',
        'postgresql://pganalytics:password@localhost:5432/pganalytics'
    )
    DB_POOL_SIZE = int(os.getenv('DB_POOL_SIZE', '10'))
    DB_POOL_RECYCLE = int(os.getenv('DB_POOL_RECYCLE', '3600'))

    # ML Model settings
    MODEL_TYPE = os.getenv('MODEL_TYPE', 'linear_regression')
    LOOKBACK_DAYS = int(os.getenv('LOOKBACK_DAYS', '90'))
    MIN_TRAINING_SAMPLES = int(os.getenv('MIN_TRAINING_SAMPLES', '100'))
    CROSS_VALIDATION_FOLDS = int(os.getenv('CROSS_VALIDATION_FOLDS', '5'))
    TEST_TRAIN_SPLIT = float(os.getenv('TEST_TRAIN_SPLIT', '0.2'))

    # Model training constraints
    MAX_OUTLIER_STD_DEVS = float(os.getenv('MAX_OUTLIER_STD_DEVS', '3.0'))
    MIN_FEATURE_VARIANCE = float(os.getenv('MIN_FEATURE_VARIANCE', '0.01'))

    # Async job settings (Celery)
    CELERY_BROKER = os.getenv('CELERY_BROKER', 'redis://localhost:6379/0')
    CELERY_BACKEND = os.getenv('CELERY_BACKEND', 'redis://localhost:6379/1')
    CELERY_TASK_TIMEOUT = int(os.getenv('CELERY_TASK_TIMEOUT', '3600'))

    # Logging configuration
    LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO')
    LOG_FORMAT = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'

    # API configuration
    API_PORT = int(os.getenv('API_PORT', '8081'))
    API_HOST = os.getenv('API_HOST', '0.0.0.0')
    REQUEST_TIMEOUT = int(os.getenv('REQUEST_TIMEOUT', '30'))

    # CORS configuration
    CORS_ORIGINS = os.getenv(
        'CORS_ORIGINS',
        'http://localhost:8080,http://localhost:3000'
    ).split(',')

    # Feature engineering
    FEATURE_SCALER_TYPE = os.getenv('FEATURE_SCALER_TYPE', 'standard')  # 'standard' or 'minmax'
    HANDLE_MISSING_VALUES = os.getenv('HANDLE_MISSING_VALUES', 'mean')  # 'mean', 'median', 'drop'

    # Prediction settings
    PREDICTION_CONFIDENCE_METHOD = os.getenv(
        'PREDICTION_CONFIDENCE_METHOD',
        'combined'  # 'std_dev', 'r_squared', or 'combined'
    )
    MIN_PREDICTION_CONFIDENCE = float(os.getenv('MIN_PREDICTION_CONFIDENCE', '0.5'))


class DevelopmentConfig(Config):
    """Development configuration"""
    DEBUG = True
    LOG_LEVEL = 'DEBUG'
    TESTING = False


class ProductionConfig(Config):
    """Production configuration"""
    DEBUG = False
    LOG_LEVEL = 'INFO'
    TESTING = False


class TestingConfig(Config):
    """Testing configuration"""
    DEBUG = True
    TESTING = True
    LOG_LEVEL = 'DEBUG'
    DATABASE_URL = 'postgresql://pganalytics:password@localhost:5432/pganalytics_test'
    CELERY_BROKER = 'memory://'
    CELERY_BACKEND = 'cache+memory://'
    LOOKBACK_DAYS = 30  # Use shorter lookback for tests
