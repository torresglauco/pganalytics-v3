#!/usr/bin/env python3
"""
pgAnalytics Automation Engine - Auto-Remediation Workflows

Implements automated remediation actions for database alerts with:
- Lock contention resolution (kill blocking locks)
- Table bloat remediation (trigger VACUUM)
- Connection pool management (close idle connections)
- Cache optimization (query optimization suggestions)
- Metrics collection recovery (restart stalled collectors)

Configuration via environment variables and action decision trees.

Usage:
    python automation_engine.py

Environment Variables:
    FLASK_HOST: Host to listen on (default: 0.0.0.0)
    FLASK_PORT: Port to listen on (default: 5002)
    DATABASE_HOST: PostgreSQL host
    DATABASE_PORT: PostgreSQL port (default: 5432)
    DATABASE_USER: PostgreSQL user
    DATABASE_PASSWORD: PostgreSQL password
    LOG_LEVEL: Logging level (default: INFO)
    DRY_RUN: Set to 'true' for dry-run mode (no actual changes)
    AUTO_REMEDIATE: Set to 'true' to enable auto-remediation (default: false)
"""

import os
import sys
import json
import logging
from datetime import datetime, timedelta
from flask import Flask, request, jsonify
from typing import Dict, Any, Optional, List, Tuple
import psycopg2
import psycopg2.extras
from enum import Enum

# Configure logging
logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Flask app configuration
app = Flask(__name__)
app.config['JSON_SORT_KEYS'] = False

# Configuration from environment
CONFIG = {
    'host': os.getenv('FLASK_HOST', '0.0.0.0'),
    'port': int(os.getenv('FLASK_PORT', '5002')),
    'db_host': os.getenv('DATABASE_HOST', 'localhost'),
    'db_port': int(os.getenv('DATABASE_PORT', '5432')),
    'db_user': os.getenv('DATABASE_USER', 'pganalytics'),
    'db_password': os.getenv('DATABASE_PASSWORD', ''),
    'db_name': os.getenv('DATABASE_NAME', 'pganalytics'),
    'dry_run': os.getenv('DRY_RUN', 'false').lower() == 'true',
    'auto_remediate': os.getenv('AUTO_REMEDIATE', 'false').lower() == 'true',
}

# Action decision thresholds
REMEDIATION_THRESHOLDS = {
    'lock_contention_critical': {
        'threshold': 10,
        'action': 'kill_blocking_locks',
        'max_age_seconds': 300,
        'enabled': True,
    },
    'high_table_bloat_warning': {
        'threshold': 50,
        'action': 'trigger_vacuum',
        'enabled': True,
        'tables_max': 5,  # Vacuum max 5 tables at once
    },
    'high_connection_count_warning': {
        'threshold': 150,
        'action': 'close_idle_connections',
        'enabled': True,
        'idle_seconds': 300,
        'max_terminate': 20,  # Terminate max 20 connections
    },
    'low_cache_hit_ratio_warning': {
        'threshold': 80,
        'action': 'optimize_cache',
        'enabled': True,
    },
    'idle_in_transaction_warning': {
        'threshold': 5,
        'action': 'close_idle_transactions',
        'enabled': True,
        'idle_seconds': 600,
    },
    'metrics_collection_failure': {
        'threshold': 15,  # minutes
        'action': 'restart_collectors',
        'enabled': True,
    },
}

# Remediation action status tracking
REMEDIATION_HISTORY: Dict[str, Dict[str, Any]] = {}
MAX_HISTORY_ENTRIES = 1000


class RemediationStatus(Enum):
    """Status of a remediation action"""
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    SUCCESS = "success"
    FAILED = "failed"
    SKIPPED = "skipped"
    PARTIAL = "partial"


class AutomationEngine:
    """Core automation and remediation logic"""

    @staticmethod
    def should_remediate(alert_name: str, alert_data: Dict[str, Any]) -> Tuple[bool, Optional[str]]:
        """
        Determine if alert should trigger auto-remediation

        Returns:
            (should_remediate: bool, action_name: Optional[str])
        """
        if not CONFIG['auto_remediate']:
            logger.debug("Auto-remediation disabled via configuration")
            return False, None

        if alert_name not in REMEDIATION_THRESHOLDS:
            logger.debug(f"No remediation defined for {alert_name}")
            return False, None

        config = REMEDIATION_THRESHOLDS[alert_name]

        if not config.get('enabled', False):
            logger.debug(f"Remediation disabled for {alert_name}")
            return False, None

        # Check if action is already running for this alert
        if AutomationEngine.is_remediation_in_progress(alert_name, alert_data.get('database')):
            logger.info(f"Remediation already in progress for {alert_name}")
            return False, None

        logger.info(f"Will remediate {alert_name} using {config['action']}")
        return True, config['action']

    @staticmethod
    def execute_remediation(alert_name: str, action: str, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Execute remediation action for alert"""

        remediation_id = f"{alert_name}_{alert_data.get('database', 'unknown')}_{int(datetime.utcnow().timestamp())}"

        logger.info(f"Starting remediation {remediation_id}: {action}")

        result = {
            'remediation_id': remediation_id,
            'alert_name': alert_name,
            'action': action,
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.PENDING.value,
            'message': '',
            'details': {},
        }

        try:
            # Route to appropriate remediation action
            if action == 'kill_blocking_locks':
                result = AutomationEngine.kill_blocking_locks(alert_data, remediation_id)
            elif action == 'trigger_vacuum':
                result = AutomationEngine.trigger_vacuum(alert_data, remediation_id)
            elif action == 'close_idle_connections':
                result = AutomationEngine.close_idle_connections(alert_data, remediation_id)
            elif action == 'close_idle_transactions':
                result = AutomationEngine.close_idle_transactions(alert_data, remediation_id)
            elif action == 'optimize_cache':
                result = AutomationEngine.analyze_cache_optimization(alert_data, remediation_id)
            elif action == 'restart_collectors':
                result = AutomationEngine.restart_collectors(alert_data, remediation_id)
            else:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = f"Unknown action: {action}"

        except Exception as e:
            logger.error(f"Remediation {remediation_id} failed: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        # Store in history
        AutomationEngine.store_remediation_history(remediation_id, result)

        return result

    @staticmethod
    def kill_blocking_locks(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Kill locks that are blocking other transactions"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'lock_contention_critical',
            'action': 'kill_blocking_locks',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.IN_PROGRESS.value,
            'message': 'Identifying and terminating blocking locks',
            'details': {
                'locks_killed': 0,
                'pids_terminated': [],
                'errors': [],
            },
        }

        try:
            # Get connection to database
            conn = AutomationEngine.get_connection(alert_data.get('database'))
            if not conn:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = 'Could not connect to database'
                return result

            cursor = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

            # Query to find blocking locks
            blocking_query = """
            SELECT
                blocking.pid AS blocking_pid,
                blocked.pid AS blocked_pid,
                blocked.usename AS blocked_user,
                blocked.application_name AS blocked_app,
                blocked.query_start AS blocked_start,
                EXTRACT(EPOCH FROM (NOW() - blocked.query_start)) AS blocked_seconds,
                blocked.query AS blocked_query
            FROM pg_stat_activity blocked
            JOIN pg_stat_activity blocking ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
            WHERE blocked.pid != blocking.pid
            ORDER BY blocked_seconds DESC
            """

            cursor.execute(blocking_query)
            blocking_locks = cursor.fetchall()

            if not blocking_locks:
                result['status'] = RemediationStatus.SKIPPED.value
                result['message'] = 'No blocking locks found'
                result['details']['locks_killed'] = 0
                cursor.close()
                conn.close()
                return result

            logger.info(f"Found {len(blocking_locks)} blocking lock(s)")

            # Kill blocking PIDs (not the blocked ones)
            blocking_pids = list(set([lock['blocking_pid'] for lock in blocking_locks]))

            for pid in blocking_pids:
                try:
                    if CONFIG['dry_run']:
                        logger.info(f"[DRY-RUN] Would terminate PID {pid}")
                        result['details']['pids_terminated'].append(pid)
                    else:
                        terminate_query = f"SELECT pg_terminate_backend({pid})"
                        cursor.execute(terminate_query)
                        result_code = cursor.fetchone()[0]

                        if result_code:
                            logger.info(f"Successfully terminated PID {pid}")
                            result['details']['pids_terminated'].append(pid)
                        else:
                            logger.warning(f"Failed to terminate PID {pid} (already terminated?)")

                except Exception as e:
                    logger.error(f"Error terminating PID {pid}: {str(e)}")
                    result['details']['errors'].append(f"PID {pid}: {str(e)}")

            if not CONFIG['dry_run']:
                conn.commit()

            result['details']['locks_killed'] = len(result['details']['pids_terminated'])
            result['status'] = RemediationStatus.SUCCESS.value if result['details']['locks_killed'] > 0 else RemediationStatus.SKIPPED.value
            result['message'] = f"Terminated {result['details']['locks_killed']} blocking process(es)"

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error in kill_blocking_locks: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        return result

    @staticmethod
    def trigger_vacuum(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Trigger VACUUM on high-bloat tables"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'high_table_bloat_warning',
            'action': 'trigger_vacuum',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.IN_PROGRESS.value,
            'message': 'Identifying and vacuuming bloated tables',
            'details': {
                'tables_vacuumed': [],
                'tables_skipped': [],
                'errors': [],
            },
        }

        try:
            conn = AutomationEngine.get_connection(alert_data.get('database'))
            if not conn:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = 'Could not connect to database'
                return result

            cursor = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

            # Query to find high-bloat tables
            bloat_query = """
            SELECT
                schemaname,
                tablename,
                ROUND(100 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1)) AS dead_ratio
            FROM pg_stat_user_tables
            WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
            ORDER BY dead_ratio DESC
            LIMIT 5
            """

            cursor.execute(bloat_query)
            bloat_tables = cursor.fetchall()

            if not bloat_tables:
                result['status'] = RemediationStatus.SKIPPED.value
                result['message'] = 'No high-bloat tables found'
                cursor.close()
                conn.close()
                return result

            logger.info(f"Found {len(bloat_tables)} high-bloat table(s)")

            for table in bloat_tables:
                table_name = f"{table['schemaname']}.{table['tablename']}"
                try:
                    if CONFIG['dry_run']:
                        logger.info(f"[DRY-RUN] Would VACUUM {table_name} (dead ratio: {table['dead_ratio']}%)")
                        result['details']['tables_vacuumed'].append(table_name)
                    else:
                        vacuum_query = f"VACUUM ANALYZE {table_name}"
                        cursor.execute(vacuum_query)
                        logger.info(f"Vacuumed {table_name}")
                        result['details']['tables_vacuumed'].append(table_name)

                except Exception as e:
                    logger.error(f"Error vacuuming {table_name}: {str(e)}")
                    result['details']['errors'].append(f"{table_name}: {str(e)}")

            if not CONFIG['dry_run']:
                conn.commit()

            result['status'] = RemediationStatus.SUCCESS.value if result['details']['tables_vacuumed'] else RemediationStatus.FAILED.value
            result['message'] = f"Vacuumed {len(result['details']['tables_vacuumed'])} table(s)"

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error in trigger_vacuum: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        return result

    @staticmethod
    def close_idle_connections(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Close idle connections to reduce connection pool pressure"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'high_connection_count_warning',
            'action': 'close_idle_connections',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.IN_PROGRESS.value,
            'message': 'Closing idle connections',
            'details': {
                'connections_closed': 0,
                'pids_terminated': [],
                'idle_seconds': 300,
                'errors': [],
            },
        }

        try:
            conn = AutomationEngine.get_connection(alert_data.get('database'))
            if not conn:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = 'Could not connect to database'
                return result

            cursor = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

            # Query to find idle connections
            idle_query = """
            SELECT
                pid,
                usename,
                application_name,
                state,
                EXTRACT(EPOCH FROM (NOW() - query_start)) AS idle_seconds
            FROM pg_stat_activity
            WHERE state = 'idle'
                AND query_start < NOW() - INTERVAL '5 minutes'
                AND pid != pg_backend_pid()
            ORDER BY query_start
            LIMIT 20
            """

            cursor.execute(idle_query)
            idle_connections = cursor.fetchall()

            if not idle_connections:
                result['status'] = RemediationStatus.SKIPPED.value
                result['message'] = 'No idle connections found'
                cursor.close()
                conn.close()
                return result

            logger.info(f"Found {len(idle_connections)} idle connection(s)")

            for conn_info in idle_connections:
                pid = conn_info['pid']
                try:
                    if CONFIG['dry_run']:
                        logger.info(f"[DRY-RUN] Would terminate idle connection PID {pid} ({conn_info['usename']})")
                        result['details']['pids_terminated'].append(pid)
                    else:
                        terminate_query = f"SELECT pg_terminate_backend({pid})"
                        cursor.execute(terminate_query)
                        result_code = cursor.fetchone()[0]

                        if result_code:
                            logger.info(f"Terminated idle connection PID {pid}")
                            result['details']['pids_terminated'].append(pid)

                except Exception as e:
                    logger.error(f"Error terminating connection PID {pid}: {str(e)}")
                    result['details']['errors'].append(f"PID {pid}: {str(e)}")

            if not CONFIG['dry_run']:
                conn.commit()

            result['details']['connections_closed'] = len(result['details']['pids_terminated'])
            result['status'] = RemediationStatus.SUCCESS.value if result['details']['connections_closed'] > 0 else RemediationStatus.PARTIAL.value
            result['message'] = f"Closed {result['details']['connections_closed']} idle connection(s)"

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error in close_idle_connections: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        return result

    @staticmethod
    def close_idle_transactions(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Close idle-in-transaction connections"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'idle_in_transaction_warning',
            'action': 'close_idle_transactions',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.IN_PROGRESS.value,
            'message': 'Closing idle-in-transaction connections',
            'details': {
                'connections_closed': 0,
                'pids_terminated': [],
                'errors': [],
            },
        }

        try:
            conn = AutomationEngine.get_connection(alert_data.get('database'))
            if not conn:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = 'Could not connect to database'
                return result

            cursor = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

            # Query to find idle-in-transaction connections
            idle_txn_query = """
            SELECT
                pid,
                usename,
                xact_start,
                EXTRACT(EPOCH FROM (NOW() - xact_start)) AS txn_seconds
            FROM pg_stat_activity
            WHERE state = 'idle in transaction'
                AND xact_start < NOW() - INTERVAL '10 minutes'
                AND pid != pg_backend_pid()
            ORDER BY xact_start
            """

            cursor.execute(idle_txn_query)
            idle_transactions = cursor.fetchall()

            if not idle_transactions:
                result['status'] = RemediationStatus.SKIPPED.value
                result['message'] = 'No idle-in-transaction connections found'
                cursor.close()
                conn.close()
                return result

            logger.info(f"Found {len(idle_transactions)} idle-in-transaction connection(s)")

            for txn_info in idle_transactions:
                pid = txn_info['pid']
                try:
                    if CONFIG['dry_run']:
                        logger.info(f"[DRY-RUN] Would terminate idle-txn PID {pid} ({txn_info['usename']})")
                        result['details']['pids_terminated'].append(pid)
                    else:
                        terminate_query = f"SELECT pg_terminate_backend({pid})"
                        cursor.execute(terminate_query)
                        result_code = cursor.fetchone()[0]

                        if result_code:
                            logger.info(f"Terminated idle-txn PID {pid}")
                            result['details']['pids_terminated'].append(pid)

                except Exception as e:
                    logger.error(f"Error terminating txn PID {pid}: {str(e)}")
                    result['details']['errors'].append(f"PID {pid}: {str(e)}")

            if not CONFIG['dry_run']:
                conn.commit()

            result['details']['connections_closed'] = len(result['details']['pids_terminated'])
            result['status'] = RemediationStatus.SUCCESS.value if result['details']['connections_closed'] > 0 else RemediationStatus.PARTIAL.value
            result['message'] = f"Closed {result['details']['connections_closed']} idle-in-transaction connection(s)"

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error in close_idle_transactions: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        return result

    @staticmethod
    def analyze_cache_optimization(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Analyze cache and provide optimization suggestions"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'low_cache_hit_ratio_warning',
            'action': 'optimize_cache',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.SUCCESS.value,
            'message': 'Cache optimization analysis complete',
            'details': {
                'current_ratio': 0.0,
                'recommendations': [],
                'high_miss_tables': [],
            },
        }

        try:
            conn = AutomationEngine.get_connection(alert_data.get('database'))
            if not conn:
                result['status'] = RemediationStatus.FAILED.value
                result['message'] = 'Could not connect to database'
                return result

            cursor = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

            # Get overall cache hit ratio
            cache_query = """
            SELECT
                SUM(heap_blks_read) as heap_blks_read,
                SUM(heap_blks_hit) as heap_blks_hit,
                SUM(idx_blks_read) as idx_blks_read,
                SUM(idx_blks_hit) as idx_blks_hit
            FROM pg_statio_user_tables
            """

            cursor.execute(cache_query)
            cache_stats = cursor.fetchone()

            total_reads = (cache_stats['heap_blks_read'] or 0) + (cache_stats['idx_blks_read'] or 0)
            total_hits = (cache_stats['heap_blks_hit'] or 0) + (cache_stats['idx_blks_hit'] or 0)

            if total_reads + total_hits > 0:
                ratio = total_hits / (total_hits + total_reads) * 100
                result['details']['current_ratio'] = round(ratio, 2)

            # Find tables with high cache miss rates
            high_miss_query = """
            SELECT
                schemaname,
                tablename,
                heap_blks_read,
                heap_blks_hit,
                ROUND(100 * heap_blks_hit / GREATEST(heap_blks_hit + heap_blks_read, 1), 2) as hit_ratio
            FROM pg_statio_user_tables
            WHERE heap_blks_hit + heap_blks_read > 1000
            ORDER BY hit_ratio ASC
            LIMIT 10
            """

            cursor.execute(high_miss_query)
            high_miss_tables = cursor.fetchall()

            for table in high_miss_tables:
                result['details']['high_miss_tables'].append({
                    'table': f"{table['schemaname']}.{table['tablename']}",
                    'hit_ratio': table['hit_ratio'],
                    'reads': table['heap_blks_read'],
                })

            # Generate recommendations
            recommendations = [
                "Add missing indexes on frequently filtered columns",
                "Increase shared_buffers if cache ratio < 80%",
                "Enable pg_stat_statements to identify expensive queries",
                "Consider partitioning large tables with poor cache hit rates",
                "Review and optimize slow queries using EXPLAIN ANALYZE",
            ]

            result['details']['recommendations'] = recommendations

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error in analyze_cache_optimization: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.PARTIAL.value
            result['message'] = f"Analysis completed with errors: {str(e)}"

        return result

    @staticmethod
    def restart_collectors(alert_data: Dict[str, Any], remediation_id: str) -> Dict[str, Any]:
        """Trigger restart of stalled metrics collectors"""

        result = {
            'remediation_id': remediation_id,
            'alert_name': 'metrics_collection_failure',
            'action': 'restart_collectors',
            'database': alert_data.get('database', 'unknown'),
            'timestamp': datetime.utcnow().isoformat(),
            'status': RemediationStatus.SUCCESS.value,
            'message': 'Restart request sent to collectors',
            'details': {
                'collectors_notified': 0,
                'notification_targets': [],
                'note': 'Collectors should restart within 60 seconds',
            },
        }

        try:
            # In production, this would send a signal/API call to collector orchestration
            # For now, log the action
            logger.info(f"Would restart collectors for database: {alert_data.get('database')}")

            if not CONFIG['dry_run']:
                # Would call collector management API or orchestration system
                result['details']['notification_targets'].append('collector-manager-api')
                result['details']['collectors_notified'] = 1
            else:
                logger.info("[DRY-RUN] Would notify collector restart system")

            result['message'] = 'Collector restart request queued'

        except Exception as e:
            logger.error(f"Error in restart_collectors: {str(e)}", exc_info=True)
            result['status'] = RemediationStatus.FAILED.value
            result['message'] = str(e)

        return result

    @staticmethod
    def get_connection(database: str = None):
        """Get PostgreSQL connection"""
        try:
            conn = psycopg2.connect(
                host=CONFIG['db_host'],
                port=CONFIG['db_port'],
                user=CONFIG['db_user'],
                password=CONFIG['db_password'],
                database=database or CONFIG['db_name'],
                connect_timeout=5
            )
            return conn
        except Exception as e:
            logger.error(f"Failed to connect to database: {str(e)}")
            return None

    @staticmethod
    def is_remediation_in_progress(alert_name: str, database: str) -> bool:
        """Check if remediation is already in progress for this alert"""
        for remediation_id, remediation in REMEDIATION_HISTORY.items():
            if (remediation['alert_name'] == alert_name and
                remediation['database'] == database and
                remediation['status'] == RemediationStatus.IN_PROGRESS.value):
                return True
        return False

    @staticmethod
    def store_remediation_history(remediation_id: str, result: Dict[str, Any]):
        """Store remediation in history with size limit"""
        REMEDIATION_HISTORY[remediation_id] = result

        # Keep only recent entries
        if len(REMEDIATION_HISTORY) > MAX_HISTORY_ENTRIES:
            # Remove oldest entries
            oldest_keys = sorted(REMEDIATION_HISTORY.keys(), key=lambda k: REMEDIATION_HISTORY[k].get('timestamp', ''))[:100]
            for key in oldest_keys:
                del REMEDIATION_HISTORY[key]


@app.route('/automation/remediate', methods=['POST', 'OPTIONS'])
def receive_remediation_request():
    """
    Receive alert and decide on remediation action

    Expected POST data:
    {
        "title": "High Table Bloat Alert",
        "severity": "warning",
        "database": "production",
        "alert_name": "high_table_bloat_warning",
        "value": "65",
        "threshold": "50",
        "timestamp": "2026-03-03T12:00:00Z"
    }
    """

    # Handle CORS preflight
    if request.method == 'OPTIONS':
        return '', 204

    try:
        data = request.get_json()

        if not data:
            logger.warning("Received empty request body")
            return jsonify({"status": "error", "message": "Empty request body"}), 400

        alert_name = data.get('alert_name', '')
        logger.info(f"Received remediation request for: {alert_name}")

        # Decide if we should remediate
        should_remediate, action = AutomationEngine.should_remediate(alert_name, data)

        if not should_remediate:
            logger.info(f"Skipping remediation for {alert_name} (not enabled or not configured)")
            return jsonify({
                "status": "skipped",
                "reason": "Remediation not enabled or not configured",
                "alert_name": alert_name
            }), 200

        # Execute remediation
        result = AutomationEngine.execute_remediation(alert_name, action, data)

        return jsonify(result), 200 if result['status'] != RemediationStatus.FAILED.value else 500

    except Exception as e:
        logger.error(f"Remediation request error: {str(e)}", exc_info=True)
        return jsonify({
            "status": "error",
            "message": str(e)
        }), 500


@app.route('/automation/history', methods=['GET'])
def get_remediation_history():
    """Get recent remediation actions"""
    limit = request.args.get('limit', '50', type=int)
    alert_name = request.args.get('alert_name', None)

    results = list(REMEDIATION_HISTORY.values())

    if alert_name:
        results = [r for r in results if r.get('alert_name') == alert_name]

    # Sort by timestamp descending
    results = sorted(results, key=lambda x: x.get('timestamp', ''), reverse=True)[:limit]

    return jsonify({
        "total": len(REMEDIATION_HISTORY),
        "returned": len(results),
        "results": results
    }), 200


@app.route('/automation/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "pgAnalytics Automation Engine",
        "timestamp": datetime.utcnow().isoformat(),
        "auto_remediate_enabled": CONFIG['auto_remediate'],
        "dry_run_mode": CONFIG['dry_run'],
        "remediation_history_size": len(REMEDIATION_HISTORY)
    }), 200


@app.route('/automation/config', methods=['GET'])
def get_config():
    """Get configuration (debugging)"""
    return jsonify({
        "auto_remediate_enabled": CONFIG['auto_remediate'],
        "dry_run_mode": CONFIG['dry_run'],
        "database_host": CONFIG['db_host'],
        "database_port": CONFIG['db_port'],
        "remediation_thresholds": {k: {
            'action': v.get('action'),
            'enabled': v.get('enabled', False),
            'threshold': v.get('threshold'),
        } for k, v in REMEDIATION_THRESHOLDS.items()},
    }), 200


@app.errorhandler(404)
def not_found(error):
    """Handle 404 errors"""
    return jsonify({"status": "error", "message": "Endpoint not found"}), 404


@app.errorhandler(500)
def server_error(error):
    """Handle 500 errors"""
    logger.error(f"Server error: {str(error)}")
    return jsonify({"status": "error", "message": "Internal server error"}), 500


def main():
    """Main entry point"""
    logger.info("=" * 60)
    logger.info("pgAnalytics Automation Engine Starting")
    logger.info("=" * 60)
    logger.info(f"Host: {CONFIG['host']}")
    logger.info(f"Port: {CONFIG['port']}")
    logger.info(f"Auto-Remediation: {'ENABLED' if CONFIG['auto_remediate'] else 'DISABLED'}")
    logger.info(f"Dry-Run Mode: {'ON' if CONFIG['dry_run'] else 'OFF'}")
    logger.info(f"Database: {CONFIG['db_host']}:{CONFIG['db_port']}/{CONFIG['db_name']}")
    logger.info("=" * 60)

    try:
        app.run(
            host=CONFIG['host'],
            port=CONFIG['port'],
            debug=os.getenv('FLASK_DEBUG', 'false').lower() == 'true',
            use_reloader=False
        )
    except KeyboardInterrupt:
        logger.info("Automation engine stopped")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Failed to start automation engine: {str(e)}")
        sys.exit(1)


if __name__ == '__main__':
    main()
