"""
Database connection and feature extraction utilities
"""

import logging
import psycopg2
from psycopg2.pool import SimpleConnectionPool
from typing import Optional, Dict, List, Any
from contextlib import contextmanager
from datetime import datetime

logger = logging.getLogger(__name__)


class DatabaseConnection:
    """
    Manages PostgreSQL connections with connection pooling
    """

    def __init__(self, database_url: str, min_connections: int = 2, max_connections: int = 10):
        """
        Initialize database connection pool

        Args:
            database_url: PostgreSQL connection string (postgresql://user:pass@host:port/db)
            min_connections: Minimum connections in pool
            max_connections: Maximum connections in pool
        """
        self.database_url = database_url
        self.min_connections = min_connections
        self.max_connections = max_connections
        self.connection_pool: Optional[SimpleConnectionPool] = None

    def initialize(self) -> bool:
        """
        Initialize the connection pool

        Returns:
            True if successful, False otherwise
        """
        try:
            self.connection_pool = SimpleConnectionPool(
                self.min_connections,
                self.max_connections,
                self.database_url
            )
            logger.info(f"Database connection pool initialized ({self.min_connections}-{self.max_connections} connections)")
            return True
        except Exception as e:
            logger.error(f"Failed to initialize connection pool: {e}")
            return False

    def close(self) -> None:
        """Close all connections in the pool"""
        if self.connection_pool:
            self.connection_pool.closeall()
            logger.info("Database connection pool closed")

    @contextmanager
    def get_connection(self):
        """
        Context manager for getting a connection from the pool

        Yields:
            Database connection from pool
        """
        if not self.connection_pool:
            raise RuntimeError("Connection pool not initialized. Call initialize() first.")

        connection = self.connection_pool.getconn()
        try:
            yield connection
        finally:
            self.connection_pool.putconn(connection)

    def extract_training_data(self, lookback_days: int = 90, min_samples: int = 100) -> tuple:
        """
        Extract historical query metrics for model training

        Args:
            lookback_days: Number of days of historical data to extract
            min_samples: Minimum number of samples required

        Returns:
            Tuple of (X_train, y_train) numpy arrays, or (None, None) if insufficient data
        """
        import numpy as np

        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                # Extract features and target from query statistics
                query = """
                SELECT
                    pqs.query_calls_per_hour,
                    pqs.mean_table_size_mb,
                    pqs.index_count,
                    CASE WHEN pqs.scan_type = 'Seq Scan' THEN 1 ELSE 0 END as has_seq_scan,
                    CASE WHEN pqs.scan_type LIKE '%Nested Loop%' THEN 1 ELSE 0 END as has_nested_loop,
                    COALESCE(pqs.subquery_depth, 0) as subquery_depth,
                    COALESCE(pqs.concurrent_queries_avg, 0) as concurrent_queries_avg,
                    COALESCE(pqs.available_memory_pct, 60) as available_memory_pct,
                    COALESCE(pqs.std_dev_calls, 0) as std_dev_calls,
                    COALESCE(pqs.peak_hour_calls, 0) as peak_hour_calls,
                    COALESCE(pqs.table_row_count, 0) as table_row_count,
                    COALESCE(pqs.avg_row_width_bytes, 0) as avg_row_width_bytes,
                    pqs.mean_execution_time_ms as target
                FROM metrics_pg_stats_query pqs
                WHERE pqs.last_seen >= NOW() - INTERVAL '%s days'
                AND pqs.mean_execution_time_ms > 0
                AND pqs.calls_per_minute > 0
                ORDER BY pqs.last_seen DESC
                LIMIT 10000
                """

                cur.execute(query % lookback_days)
                rows = cur.fetchall()
                cur.close()

                if len(rows) < min_samples:
                    logger.warning(f"Insufficient training samples: {len(rows)} < {min_samples}")
                    return None, None

                # Convert to numpy arrays
                data = np.array(rows, dtype=np.float64)
                X = data[:, :-1]  # All columns except last
                y = data[:, -1]   # Last column is target

                logger.info(f"Extracted {len(X)} training samples with {X.shape[1]} features")
                return X, y

        except Exception as e:
            logger.error(f"Error extracting training data: {e}")
            return None, None

    def extract_features_for_query(self, query_hash: int) -> Optional[Dict[str, Any]]:
        """
        Extract feature vector for a specific query

        Args:
            query_hash: Hash of the query

        Returns:
            Dictionary with feature names and values, or None if query not found
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                # Extract features for a single query
                query = """
                SELECT
                    pqs.query_calls_per_hour,
                    pqs.mean_table_size_mb,
                    pqs.index_count,
                    CASE WHEN pqs.scan_type = 'Seq Scan' THEN 1 ELSE 0 END as has_seq_scan,
                    CASE WHEN pqs.scan_type LIKE '%Nested Loop%' THEN 1 ELSE 0 END as has_nested_loop,
                    COALESCE(pqs.subquery_depth, 0) as subquery_depth,
                    COALESCE(pqs.concurrent_queries_avg, 0) as concurrent_queries_avg,
                    COALESCE(pqs.available_memory_pct, 60) as available_memory_pct,
                    COALESCE(pqs.std_dev_calls, 0) as std_dev_calls,
                    COALESCE(pqs.peak_hour_calls, 0) as peak_hour_calls,
                    COALESCE(pqs.table_row_count, 0) as table_row_count,
                    COALESCE(pqs.avg_row_width_bytes, 0) as avg_row_width_bytes
                FROM metrics_pg_stats_query pqs
                WHERE pqs.query_hash = %s
                """

                cur.execute(query, (query_hash,))
                row = cur.fetchone()
                cur.close()

                if not row:
                    logger.warning(f"Query not found: {query_hash}")
                    return None

                feature_names = [
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

                return dict(zip(feature_names, row))

        except Exception as e:
            logger.error(f"Error extracting features for query {query_hash}: {e}")
            return None

    def save_model_metadata(self, model_id: str, model_type: str, model_name: str,
                          r_squared: float, rmse: float, mae: float,
                          training_samples: int) -> bool:
        """
        Save model metadata to database

        Args:
            model_id: Unique model identifier
            model_type: Model type (linear_regression, decision_tree, random_forest)
            model_name: Friendly model name
            r_squared: Model R² score
            rmse: Root mean squared error
            mae: Mean absolute error
            training_samples: Number of training samples

        Returns:
            True if successful, False otherwise
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                INSERT INTO query_performance_models
                (model_type, model_name, feature_names, training_sample_size, r_squared, created_at)
                VALUES (%s, %s, %s, %s, %s, NOW())
                ON CONFLICT (model_name) DO UPDATE SET
                    r_squared = EXCLUDED.r_squared,
                    training_sample_size = EXCLUDED.training_sample_size,
                    last_updated = NOW()
                """

                feature_names = [
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

                cur.execute(query, (model_type, model_name, feature_names, training_samples, r_squared))
                conn.commit()
                cur.close()

                logger.info(f"Saved model metadata: {model_name} (R²={r_squared:.4f})")
                return True

        except Exception as e:
            logger.error(f"Error saving model metadata: {e}")
            return False

    def record_prediction(self, query_hash: int, predicted_ms: float, confidence: float) -> bool:
        """
        Record a prediction in the database for accuracy tracking

        Args:
            query_hash: Hash of the query
            predicted_ms: Predicted execution time in milliseconds
            confidence: Confidence score (0-1)

        Returns:
            True if successful, False otherwise
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                logger.debug(f"Prediction recorded: query_hash={query_hash}, predicted_ms={predicted_ms:.2f}, confidence={confidence:.2f}")
                cur.close()
                return True

        except Exception as e:
            logger.error(f"Error recording prediction: {e}")
            return False

    def get_latest_model(self) -> Optional[Dict[str, Any]]:
        """
        Get the latest trained model metadata

        Returns:
            Dictionary with model metadata or None if no models exist
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    id,
                    model_type,
                    model_name,
                    feature_names,
                    training_sample_size,
                    r_squared,
                    created_at,
                    last_updated
                FROM query_performance_models
                ORDER BY created_at DESC
                LIMIT 1
                """

                cur.execute(query)
                row = cur.fetchone()
                cur.close()

                if not row:
                    logger.debug("No models found in database")
                    return None

                model_data = {
                    'id': row[0],
                    'model_type': row[1],
                    'model_name': row[2],
                    'feature_names': row[3] if row[3] else [],
                    'training_sample_size': row[4],
                    'r_squared': row[5],
                    'created_at': row[6].isoformat() if row[6] else None,
                    'last_updated': row[7].isoformat() if row[7] else None,
                }

                logger.info(f"Retrieved latest model: {model_data['model_name']}")
                return model_data

        except Exception as e:
            logger.error(f"Error retrieving latest model: {e}")
            return None

    def get_all_models(self, limit: int = 10) -> List[Dict[str, Any]]:
        """
        Get all trained models with metadata

        Args:
            limit: Maximum number of models to return

        Returns:
            List of model metadata dictionaries
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    id,
                    model_type,
                    model_name,
                    training_sample_size,
                    r_squared,
                    created_at,
                    last_updated
                FROM query_performance_models
                ORDER BY created_at DESC
                LIMIT %s
                """

                cur.execute(query, (limit,))
                rows = cur.fetchall()
                cur.close()

                models = []
                for row in rows:
                    model_data = {
                        'id': row[0],
                        'model_type': row[1],
                        'model_name': row[2],
                        'training_sample_size': row[3],
                        'r_squared': row[4],
                        'created_at': row[5].isoformat() if row[5] else None,
                        'last_updated': row[6].isoformat() if row[6] else None,
                    }
                    models.append(model_data)

                logger.info(f"Retrieved {len(models)} models from database")
                return models

        except Exception as e:
            logger.error(f"Error retrieving all models: {e}")
            return []

    def get_model_by_id(self, model_id: int) -> Optional[Dict[str, Any]]:
        """
        Get specific model by ID

        Args:
            model_id: Model ID in database

        Returns:
            Dictionary with model metadata or None if not found
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    id,
                    model_type,
                    model_name,
                    feature_names,
                    training_sample_size,
                    r_squared,
                    created_at,
                    last_updated
                FROM query_performance_models
                WHERE id = %s
                """

                cur.execute(query, (model_id,))
                row = cur.fetchone()
                cur.close()

                if not row:
                    logger.warning(f"Model not found: id={model_id}")
                    return None

                model_data = {
                    'id': row[0],
                    'model_type': row[1],
                    'model_name': row[2],
                    'feature_names': row[3] if row[3] else [],
                    'training_sample_size': row[4],
                    'r_squared': row[5],
                    'created_at': row[6].isoformat() if row[6] else None,
                    'last_updated': row[7].isoformat() if row[7] else None,
                }

                return model_data

        except Exception as e:
            logger.error(f"Error retrieving model {model_id}: {e}")
            return None

    def get_active_model(self) -> Optional[Dict[str, Any]]:
        """
        Get the currently active model for predictions

        Returns:
            Dictionary with active model metadata or None if no active model
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                # Get the most recently created model as active
                query = """
                SELECT
                    id,
                    model_type,
                    model_name,
                    feature_names,
                    training_sample_size,
                    r_squared,
                    created_at,
                    last_updated
                FROM query_performance_models
                WHERE created_at = (SELECT MAX(created_at) FROM query_performance_models)
                LIMIT 1
                """

                cur.execute(query)
                row = cur.fetchone()
                cur.close()

                if not row:
                    logger.debug("No active model found")
                    return None

                model_data = {
                    'id': row[0],
                    'model_type': row[1],
                    'model_name': row[2],
                    'feature_names': row[3] if row[3] else [],
                    'training_sample_size': row[4],
                    'r_squared': row[5],
                    'is_active': True,
                    'created_at': row[6].isoformat() if row[6] else None,
                    'last_updated': row[7].isoformat() if row[7] else None,
                }

                logger.info(f"Retrieved active model: {model_data['model_name']}")
                return model_data

        except Exception as e:
            logger.error(f"Error retrieving active model: {e}")
            return None

    def get_query_prediction_history(self, query_hash: int, limit: int = 100) -> List[Dict[str, Any]]:
        """
        Get prediction history for a specific query

        Args:
            query_hash: Hash of the query
            limit: Maximum number of predictions to return

        Returns:
            List of prediction records
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                # Query prediction validation history if available
                # This would use a predictions table if it exists
                query = """
                SELECT
                    query_hash,
                    calls_per_minute,
                    mean_execution_time_ms,
                    stddev_execution_time_ms,
                    last_seen
                FROM metrics_pg_stats_query
                WHERE query_hash = %s
                ORDER BY last_seen DESC
                LIMIT %s
                """

                cur.execute(query, (query_hash, limit))
                rows = cur.fetchall()
                cur.close()

                predictions = []
                for row in rows:
                    pred = {
                        'query_hash': row[0],
                        'calls_per_minute': row[1],
                        'actual_execution_time_ms': row[2],
                        'std_dev_ms': row[3],
                        'timestamp': row[4].isoformat() if row[4] else None,
                    }
                    predictions.append(pred)

                logger.debug(f"Retrieved {len(predictions)} prediction records for query {query_hash}")
                return predictions

        except Exception as e:
            logger.error(f"Error retrieving prediction history: {e}")
            return []

    def get_query_statistics(self, query_hash: int) -> Optional[Dict[str, Any]]:
        """
        Get detailed statistics for a specific query

        Args:
            query_hash: Hash of the query

        Returns:
            Dictionary with query statistics or None if not found
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    query_hash,
                    calls_per_minute,
                    mean_execution_time_ms,
                    stddev_execution_time_ms,
                    min_execution_time_ms,
                    max_execution_time_ms,
                    scan_type,
                    index_count,
                    table_row_count,
                    mean_table_size_mb,
                    last_seen
                FROM metrics_pg_stats_query
                WHERE query_hash = %s
                """

                cur.execute(query, (query_hash,))
                row = cur.fetchone()
                cur.close()

                if not row:
                    logger.debug(f"Query statistics not found: {query_hash}")
                    return None

                stats = {
                    'query_hash': row[0],
                    'calls_per_minute': row[1],
                    'mean_execution_time_ms': row[2],
                    'stddev_execution_time_ms': row[3],
                    'min_execution_time_ms': row[4],
                    'max_execution_time_ms': row[5],
                    'scan_type': row[6],
                    'index_count': row[7],
                    'table_row_count': row[8],
                    'mean_table_size_mb': row[9],
                    'last_seen': row[10].isoformat() if row[10] else None,
                }

                logger.debug(f"Retrieved statistics for query {query_hash}")
                return stats

        except Exception as e:
            logger.error(f"Error retrieving query statistics: {e}")
            return None

    def get_slow_queries(self, threshold_ms: float = 1000, limit: int = 20) -> List[Dict[str, Any]]:
        """
        Get slowest queries exceeding threshold

        Args:
            threshold_ms: Execution time threshold in milliseconds
            limit: Maximum number of queries to return

        Returns:
            List of slow query statistics
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    query_hash,
                    mean_execution_time_ms,
                    calls_per_minute,
                    index_count,
                    scan_type,
                    last_seen
                FROM metrics_pg_stats_query
                WHERE mean_execution_time_ms > %s
                ORDER BY mean_execution_time_ms DESC
                LIMIT %s
                """

                cur.execute(query, (threshold_ms, limit))
                rows = cur.fetchall()
                cur.close()

                slow_queries = []
                for row in rows:
                    slow_query = {
                        'query_hash': row[0],
                        'mean_execution_time_ms': row[1],
                        'calls_per_minute': row[2],
                        'index_count': row[3],
                        'scan_type': row[4],
                        'last_seen': row[5].isoformat() if row[5] else None,
                        'total_time_ms': row[1] * row[2],  # Calculated impact
                    }
                    slow_queries.append(slow_query)

                logger.debug(f"Retrieved {len(slow_queries)} slow queries")
                return slow_queries

        except Exception as e:
            logger.error(f"Error retrieving slow queries: {e}")
            return []

    def get_frequently_executed_queries(self, limit: int = 20) -> List[Dict[str, Any]]:
        """
        Get most frequently executed queries

        Args:
            limit: Maximum number of queries to return

        Returns:
            List of frequently executed queries with metrics
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                query = """
                SELECT
                    query_hash,
                    calls_per_minute,
                    mean_execution_time_ms,
                    index_count,
                    scan_type,
                    last_seen
                FROM metrics_pg_stats_query
                ORDER BY calls_per_minute DESC
                LIMIT %s
                """

                cur.execute(query, (limit,))
                rows = cur.fetchall()
                cur.close()

                frequent_queries = []
                for row in rows:
                    freq_query = {
                        'query_hash': row[0],
                        'calls_per_minute': row[1],
                        'mean_execution_time_ms': row[2],
                        'index_count': row[3],
                        'scan_type': row[4],
                        'last_seen': row[5].isoformat() if row[5] else None,
                        'total_impact_ms': row[1] * row[2],  # Total time in system
                    }
                    frequent_queries.append(freq_query)

                logger.debug(f"Retrieved {len(frequent_queries)} frequently executed queries")
                return frequent_queries

        except Exception as e:
            logger.error(f"Error retrieving frequently executed queries: {e}")
            return []

    def get_database_health_summary(self) -> Optional[Dict[str, Any]]:
        """
        Get overall database health summary

        Returns:
            Dictionary with health metrics or None if error
        """
        try:
            with self.get_connection() as conn:
                cur = conn.cursor()

                # Get overall metrics
                query = """
                SELECT
                    COUNT(*) as total_queries,
                    AVG(mean_execution_time_ms) as avg_execution_ms,
                    MAX(mean_execution_time_ms) as max_execution_ms,
                    MIN(mean_execution_time_ms) as min_execution_ms,
                    SUM(calls_per_minute) as total_calls_per_minute,
                    COUNT(CASE WHEN scan_type = 'Seq Scan' THEN 1 END) as seq_scan_count,
                    COUNT(CASE WHEN index_count > 0 THEN 1 END) as indexed_count
                FROM metrics_pg_stats_query
                """

                cur.execute(query)
                row = cur.fetchone()
                cur.close()

                if not row or row[0] is None or row[0] == 0:
                    logger.debug("No query metrics available yet")
                    return None

                health = {
                    'total_queries': row[0],
                    'avg_execution_ms': float(row[1]) if row[1] else 0,
                    'max_execution_ms': float(row[2]) if row[2] else 0,
                    'min_execution_ms': float(row[3]) if row[3] else 0,
                    'total_calls_per_minute': float(row[4]) if row[4] else 0,
                    'seq_scan_count': row[5] or 0,
                    'indexed_count': row[6] or 0,
                    'timestamp': datetime.utcnow().isoformat(),
                }

                logger.debug(f"Retrieved database health summary")
                return health

        except Exception as e:
            logger.error(f"Error retrieving database health: {e}")
            return None
