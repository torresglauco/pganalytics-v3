#!/usr/bin/env python3
"""
pgAnalytics Incident Correlation Engine

Correlates related alerts into single incidents and groups them for better
incident management. Implements:

- Alert correlation (group related alerts)
- Incident deduplication
- Alert aggregation with root cause analysis
- Escalation path management
- Incident lifecycle tracking

Configuration via environment variables.

Usage:
    python incident_correlation_engine.py

Environment Variables:
    FLASK_HOST: Host to listen on (default: 0.0.0.0)
    FLASK_PORT: Port to listen on (default: 5003)
    CORRELATION_WINDOW: Time window for correlation in seconds (default: 300)
    INCIDENT_TRACKING_URL: Incident tracking API endpoint
    INCIDENT_TRACKING_TOKEN: Bearer token for incident tracking
    LOG_LEVEL: Logging level (default: INFO)
"""

import os
import sys
import json
import logging
from datetime import datetime, timedelta
from flask import Flask, request, jsonify
from typing import Dict, Any, Optional, List, Set
from enum import Enum
import hashlib

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
    'port': int(os.getenv('FLASK_PORT', '5003')),
    'correlation_window': int(os.getenv('CORRELATION_WINDOW', '300')),  # 5 minutes
    'incident_api_url': os.getenv('INCIDENT_TRACKING_URL', 'http://localhost:8080/api/incidents'),
    'incident_api_token': os.getenv('INCIDENT_TRACKING_TOKEN', ''),
}

# Correlation rules - define which alerts should be grouped together
CORRELATION_RULES = {
    'lock_group': {
        'name': 'Lock Contention',
        'alerts': ['lock_contention_critical', 'blocking_transaction_critical', 'max_lock_age_warning'],
        'description': 'PostgreSQL locking issues causing transaction delays',
        'priority': 'critical',
    },
    'performance_group': {
        'name': 'Performance Issues',
        'alerts': ['low_cache_hit_ratio_warning', 'high_table_bloat_warning'],
        'description': 'Performance degradation from cache misses and table bloat',
        'priority': 'high',
    },
    'connection_group': {
        'name': 'Connection Pool Issues',
        'alerts': ['high_connection_count_warning', 'idle_in_transaction_warning'],
        'description': 'Connection pool saturation and resource leaks',
        'priority': 'high',
    },
    'system_group': {
        'name': 'System Health',
        'alerts': ['metrics_collection_failure'],
        'description': 'Database monitoring system health issues',
        'priority': 'critical',
    },
}

# Incident correlation state - tracks active correlated incidents
CORRELATED_INCIDENTS: Dict[str, Dict[str, Any]] = {}
INCIDENT_ALERT_MAP: Dict[str, str] = {}  # Maps alert_id to incident_id
MAX_INCIDENT_HISTORY = 500

# Severity levels for escalation
SEVERITY_LEVELS = {
    'critical': 1,
    'high': 2,
    'medium': 3,
    'low': 4,
}


class IncidentState(Enum):
    """State of a correlated incident"""
    ACTIVE = "active"
    ESCALATED = "escalated"
    RESOLVED = "resolved"
    SUPPRESSED = "suppressed"


class IncidentCorrelationEngine:
    """Core incident correlation and grouping logic"""

    @staticmethod
    def find_correlation_group(alert_name: str) -> Optional[Dict[str, Any]]:
        """Find correlation group for an alert"""
        for group_id, group_config in CORRELATION_RULES.items():
            if alert_name in group_config['alerts']:
                return {
                    'group_id': group_id,
                    'config': group_config,
                }
        return None

    @staticmethod
    def calculate_incident_signature(alert_data: Dict[str, Any], group_id: str) -> str:
        """
        Calculate signature for incident correlation

        Signature combines database + severity + group to correlate related alerts
        """
        database = alert_data.get('database', 'unknown')
        severity = alert_data.get('severity', 'unknown')

        signature_string = f"{database}_{severity}_{group_id}"
        signature = hashlib.md5(signature_string.encode()).hexdigest()[:16]

        return signature

    @staticmethod
    def find_existing_incident(alert_data: Dict[str, Any], group_id: str) -> Optional[str]:
        """
        Find existing incident to correlate with

        Returns incident_id if found within correlation window, None otherwise
        """
        signature = IncidentCorrelationEngine.calculate_incident_signature(alert_data, group_id)
        database = alert_data.get('database', 'unknown')

        for incident_id, incident in CORRELATED_INCIDENTS.items():
            # Check if within correlation window
            incident_timestamp = datetime.fromisoformat(incident['timestamp'])
            time_diff = (datetime.utcnow() - incident_timestamp).total_seconds()

            if (time_diff < CONFIG['correlation_window'] and
                incident['database'] == database and
                incident['group_id'] == group_id and
                incident['state'] == IncidentState.ACTIVE.value):
                return incident_id

        return None

    @staticmethod
    def create_correlated_incident(alert_data: Dict[str, Any], group_config: Dict[str, Any]) -> str:
        """Create a new correlated incident"""

        incident_id = f"INC_{int(datetime.utcnow().timestamp())}_{alert_data.get('database', 'unknown')}"

        incident = {
            'incident_id': incident_id,
            'group_id': group_config['group_id'],
            'group_name': group_config['config']['name'],
            'database': alert_data.get('database', 'unknown'),
            'severity': alert_data.get('severity', 'warning'),
            'priority': group_config['config']['priority'],
            'state': IncidentState.ACTIVE.value,
            'timestamp': datetime.utcnow().isoformat(),
            'alerts': [alert_data],
            'alert_count': 1,
            'description': group_config['config']['description'],
            'root_cause': None,
            'remediation_status': 'pending',
            'related_incidents': [],
        }

        CORRELATED_INCIDENTS[incident_id] = incident
        INCIDENT_ALERT_MAP[alert_data.get('alert_name', '')] = incident_id

        logger.info(f"Created correlated incident {incident_id} for group {group_config['config']['name']}")

        return incident_id

    @staticmethod
    def add_alert_to_incident(incident_id: str, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Add alert to existing incident"""

        if incident_id not in CORRELATED_INCIDENTS:
            logger.error(f"Incident {incident_id} not found")
            return {'status': 'error', 'message': 'Incident not found'}

        incident = CORRELATED_INCIDENTS[incident_id]
        incident['alerts'].append(alert_data)
        incident['alert_count'] = len(incident['alerts'])
        incident['timestamp'] = datetime.utcnow().isoformat()

        # Update severity if new alert is more severe
        new_severity_level = SEVERITY_LEVELS.get(alert_data.get('severity', 'low'), 4)
        current_severity_level = SEVERITY_LEVELS.get(incident['severity'], 4)

        if new_severity_level < current_severity_level:
            incident['severity'] = alert_data.get('severity')

        logger.info(f"Added alert to incident {incident_id} (total alerts: {incident['alert_count']})")

        return {
            'status': 'success',
            'incident_id': incident_id,
            'alert_count': incident['alert_count'],
        }

    @staticmethod
    def correlate_alert(alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Correlate incoming alert with existing incidents

        Returns correlation result with incident_id and new/existing status
        """

        alert_name = alert_data.get('alert_name', '')
        logger.info(f"Correlating alert: {alert_name}")

        # Find correlation group for this alert
        group_info = IncidentCorrelationEngine.find_correlation_group(alert_name)

        if not group_info:
            logger.debug(f"No correlation group for {alert_name}, creating standalone incident")
            return {
                'status': 'skipped',
                'message': 'Alert has no correlation group',
                'alert_name': alert_name,
            }

        group_id = group_info['group_id']
        group_config = group_info['config']

        # Try to find existing incident
        existing_incident_id = IncidentCorrelationEngine.find_existing_incident(alert_data, group_id)

        if existing_incident_id:
            logger.info(f"Found existing incident {existing_incident_id}, correlating alert")
            result = IncidentCorrelationEngine.add_alert_to_incident(existing_incident_id, alert_data)
            result['is_new'] = False
            result['incident_id'] = existing_incident_id
            return result
        else:
            logger.info(f"Creating new correlated incident for group {group_config['name']}")
            group_with_id = {'group_id': group_id, 'config': group_config}
            incident_id = IncidentCorrelationEngine.create_correlated_incident(alert_data, group_with_id)

            return {
                'status': 'success',
                'incident_id': incident_id,
                'is_new': True,
                'group_name': group_config['name'],
                'alert_count': 1,
            }

    @staticmethod
    def analyze_root_cause(incident_id: str) -> Dict[str, Any]:
        """
        Analyze alerts in incident to determine root cause

        Uses alert sequence and characteristics to suggest root cause
        """

        if incident_id not in CORRELATED_INCIDENTS:
            return {'status': 'error', 'message': 'Incident not found'}

        incident = CORRELATED_INCIDENTS[incident_id]
        alerts = incident['alerts']

        analysis = {
            'incident_id': incident_id,
            'alert_count': len(alerts),
            'timestamp': datetime.utcnow().isoformat(),
            'possible_root_causes': [],
            'recommended_actions': [],
            'confidence': 0,
        }

        group_id = incident['group_id']

        # Lock group analysis
        if group_id == 'lock_group':
            analysis['possible_root_causes'] = [
                "Long-running transaction holding locks",
                "Unhandled exception leaving transaction open",
                "Application deadlock or lock escalation",
                "Batch operation without transaction timeout",
            ]
            analysis['recommended_actions'] = [
                "Identify long-running transactions with pg_stat_activity",
                "Check application logs for deadlocks",
                "Review query execution plans for optimization",
                "Implement application-level lock timeouts",
            ]
            analysis['confidence'] = 85

        # Performance group analysis
        elif group_id == 'performance_group':
            analysis['possible_root_causes'] = [
                "Missing indexes on frequently accessed tables",
                "Table autovacuum lagging behind write load",
                "Query plan changes due to statistics update",
                "Shared buffer pressure or cache thrashing",
            ]
            analysis['recommended_actions'] = [
                "Run ANALYZE to update table statistics",
                "Identify missing indexes using pg_stat_user_indexes",
                "Consider aggressive autovacuum tuning",
                "Profile top queries with pg_stat_statements",
            ]
            analysis['confidence'] = 75

        # Connection group analysis
        elif group_id == 'connection_group':
            analysis['possible_root_causes'] = [
                "Connection pool not returning connections",
                "Application holding connections open",
                "Connection leak in ORM or driver",
                "Long-running transactions blocking connection release",
            ]
            analysis['recommended_actions'] = [
                "Monitor connection state distribution",
                "Check application for connection pool leaks",
                "Verify application retry/timeout settings",
                "Review database session durations",
            ]
            analysis['confidence'] = 80

        # System group analysis
        elif group_id == 'system_group':
            analysis['possible_root_causes'] = [
                "Collector process crashed or hung",
                "Database connectivity issue",
                "Collector resource exhaustion",
                "Database under heavy load preventing metric collection",
            ]
            analysis['recommended_actions'] = [
                "Check collector process status",
                "Verify database connection from collector host",
                "Monitor collector resource usage",
                "Review database slow query log",
            ]
            analysis['confidence'] = 90

        return analysis

    @staticmethod
    def get_incident_summary(incident_id: str) -> Dict[str, Any]:
        """Get comprehensive summary of an incident"""

        if incident_id not in CORRELATED_INCIDENTS:
            return {'status': 'error', 'message': 'Incident not found'}

        incident = CORRELATED_INCIDENTS[incident_id]

        # Analyze root cause
        root_cause_analysis = IncidentCorrelationEngine.analyze_root_cause(incident_id)

        summary = {
            'incident_id': incident_id,
            'group_name': incident['group_name'],
            'database': incident['database'],
            'severity': incident['severity'],
            'priority': incident['priority'],
            'state': incident['state'],
            'created_at': incident['timestamp'],
            'alert_count': incident['alert_count'],
            'affected_alerts': [alert.get('alert_name', 'unknown') for alert in incident['alerts']],
            'root_cause_analysis': root_cause_analysis,
            'description': incident['description'],
        }

        return summary

    @staticmethod
    def resolve_incident(incident_id: str, resolution_notes: str = '') -> Dict[str, Any]:
        """Mark incident as resolved"""

        if incident_id not in CORRELATED_INCIDENTS:
            return {'status': 'error', 'message': 'Incident not found'}

        incident = CORRELATED_INCIDENTS[incident_id]
        incident['state'] = IncidentState.RESOLVED.value
        incident['resolved_at'] = datetime.utcnow().isoformat()
        incident['resolution_notes'] = resolution_notes

        logger.info(f"Resolved incident {incident_id}")

        return {
            'status': 'success',
            'incident_id': incident_id,
            'message': 'Incident marked as resolved',
        }

    @staticmethod
    def cleanup_old_incidents():
        """Remove resolved incidents older than 24 hours"""

        cutoff_time = datetime.utcnow() - timedelta(hours=24)
        incidents_to_remove = []

        for incident_id, incident in CORRELATED_INCIDENTS.items():
            if incident['state'] == IncidentState.RESOLVED.value:
                incident_time = datetime.fromisoformat(incident.get('resolved_at', incident['timestamp']))
                if incident_time < cutoff_time:
                    incidents_to_remove.append(incident_id)

        for incident_id in incidents_to_remove:
            del CORRELATED_INCIDENTS[incident_id]

        logger.info(f"Cleaned up {len(incidents_to_remove)} old resolved incidents")


@app.route('/correlation/correlate', methods=['POST', 'OPTIONS'])
def correlate_alert():
    """
    Receive alert and correlate with existing incidents

    Expected POST data:
    {
        "alert_name": "lock_contention_critical",
        "severity": "critical",
        "database": "production",
        "value": "15",
        "threshold": "10",
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

        logger.info(f"Correlating alert: {data.get('alert_name')}")

        # Correlate the alert
        correlation_result = IncidentCorrelationEngine.correlate_alert(data)

        return jsonify(correlation_result), 200

    except Exception as e:
        logger.error(f"Correlation error: {str(e)}", exc_info=True)
        return jsonify({
            "status": "error",
            "message": str(e)
        }), 500


@app.route('/correlation/incident/<incident_id>', methods=['GET'])
def get_incident(incident_id):
    """Get incident details and analysis"""

    try:
        summary = IncidentCorrelationEngine.get_incident_summary(incident_id)

        if summary.get('status') == 'error':
            return jsonify(summary), 404

        return jsonify(summary), 200

    except Exception as e:
        logger.error(f"Error fetching incident: {str(e)}")
        return jsonify({"status": "error", "message": str(e)}), 500


@app.route('/correlation/incident/<incident_id>/resolve', methods=['POST'])
def resolve_incident_endpoint(incident_id):
    """Mark incident as resolved"""

    try:
        data = request.get_json() or {}
        notes = data.get('resolution_notes', '')

        result = IncidentCorrelationEngine.resolve_incident(incident_id, notes)

        return jsonify(result), 200 if result['status'] == 'success' else 404

    except Exception as e:
        logger.error(f"Error resolving incident: {str(e)}")
        return jsonify({"status": "error", "message": str(e)}), 500


@app.route('/correlation/incidents', methods=['GET'])
def list_incidents():
    """List active and recent incidents"""

    state_filter = request.args.get('state', None)
    limit = request.args.get('limit', '50', type=int)

    incidents = list(CORRELATED_INCIDENTS.values())

    if state_filter:
        incidents = [i for i in incidents if i['state'] == state_filter]

    # Sort by timestamp descending
    incidents = sorted(incidents, key=lambda x: x['timestamp'], reverse=True)[:limit]

    return jsonify({
        'total': len(CORRELATED_INCIDENTS),
        'returned': len(incidents),
        'incidents': incidents
    }), 200


@app.route('/correlation/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "pgAnalytics Incident Correlation Engine",
        "timestamp": datetime.utcnow().isoformat(),
        "active_incidents": len([i for i in CORRELATED_INCIDENTS.values() if i['state'] == IncidentState.ACTIVE.value]),
        "total_incidents": len(CORRELATED_INCIDENTS)
    }), 200


@app.route('/correlation/config', methods=['GET'])
def get_config():
    """Get configuration (debugging)"""
    return jsonify({
        "correlation_window_seconds": CONFIG['correlation_window'],
        "correlation_groups": {k: {
            'name': v['name'],
            'alerts': v['alerts'],
            'priority': v['priority'],
        } for k, v in CORRELATION_RULES.items()},
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
    logger.info("pgAnalytics Incident Correlation Engine Starting")
    logger.info("=" * 60)
    logger.info(f"Host: {CONFIG['host']}")
    logger.info(f"Port: {CONFIG['port']}")
    logger.info(f"Correlation Window: {CONFIG['correlation_window']} seconds")
    logger.info(f"Correlation Groups: {len(CORRELATION_RULES)}")
    logger.info("=" * 60)

    try:
        app.run(
            host=CONFIG['host'],
            port=CONFIG['port'],
            debug=os.getenv('FLASK_DEBUG', 'false').lower() == 'true',
            use_reloader=False
        )
    except KeyboardInterrupt:
        logger.info("Correlation engine stopped")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Failed to start correlation engine: {str(e)}")
        sys.exit(1)


if __name__ == '__main__':
    main()
