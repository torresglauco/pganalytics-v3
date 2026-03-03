#!/usr/bin/env python3
"""
pgAnalytics Webhook Receiver for Incident Tracking

Receives alerts from Grafana and forwards them to incident tracking system.
Handles incident creation, updates, and resolution tracking.

Usage:
    python webhook_incident_receiver.py

Environment Variables:
    FLASK_HOST: Host to listen on (default: 0.0.0.0)
    FLASK_PORT: Port to listen on (default: 5000)
    INCIDENT_TRACKING_URL: URL of incident tracking API
    INCIDENT_TRACKING_TOKEN: Authentication token for incident tracking
    LOG_LEVEL: Logging level (default: INFO)
"""

import os
import sys
import json
import logging
from datetime import datetime, timedelta
from flask import Flask, request, jsonify
from typing import Dict, Any, Optional
import requests
from requests.auth import HTTPBearerAuth
from urllib.parse import urljoin

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
    'port': int(os.getenv('FLASK_PORT', '5000')),
    'incident_api_url': os.getenv('INCIDENT_TRACKING_URL', 'http://localhost:8080/api/incidents'),
    'incident_api_token': os.getenv('INCIDENT_TRACKING_TOKEN', ''),
    'grafana_url': os.getenv('GRAFANA_URL', 'http://localhost:3000'),
}

# In-memory incident cache (for deduplication)
INCIDENT_CACHE: Dict[str, Dict[str, Any]] = {}
CACHE_EXPIRY = 3600  # 1 hour


class IncidentManager:
    """Manage incident creation and updates"""

    @staticmethod
    def create_incident(alert_data: Dict[str, Any]) -> Optional[str]:
        """Create incident in tracking system"""
        try:
            incident = {
                "title": alert_data.get('title', 'Unknown Alert'),
                "description": alert_data.get('description', ''),
                "severity": IncidentManager.map_severity(alert_data.get('severity', 'warning')),
                "service": "PostgreSQL Monitoring",
                "database": alert_data.get('database', 'unknown'),
                "source": "pgAnalytics",
                "tags": [
                    "postgresql",
                    "monitoring",
                    alert_data.get('severity', 'warning').lower()
                ],
                "external_refs": {
                    "dashboard": alert_data.get('dashboard_url', ''),
                    "runbook": alert_data.get('runbook_url', ''),
                    "alert_rule": alert_data.get('alert_rule', '')
                },
                "metadata": {
                    "alert_value": alert_data.get('value', 'N/A'),
                    "threshold": alert_data.get('threshold', 'N/A'),
                    "timestamp": datetime.utcnow().isoformat(),
                }
            }

            # Filter out empty values
            incident = {k: v for k, v in incident.items() if v}

            # Send to incident tracking system
            headers = {
                'Authorization': f'Bearer {CONFIG["incident_api_token"]}',
                'Content-Type': 'application/json'
            }

            logger.info(f"Creating incident for: {incident['title']}")

            response = requests.post(
                CONFIG['incident_api_url'],
                json=incident,
                headers=headers,
                timeout=10
            )

            if response.status_code in [200, 201]:
                incident_id = response.json().get('id') or response.json().get('key')
                logger.info(f"Incident created: {incident_id}")

                # Cache the incident
                cache_key = f"{alert_data.get('alert_name', 'unknown')}_{alert_data.get('database', 'unknown')}"
                INCIDENT_CACHE[cache_key] = {
                    'incident_id': incident_id,
                    'created_at': datetime.utcnow(),
                    'severity': alert_data.get('severity', 'warning')
                }

                return incident_id
            else:
                logger.error(f"Failed to create incident: {response.status_code} - {response.text}")
                return None

        except Exception as e:
            logger.error(f"Error creating incident: {str(e)}")
            return None

    @staticmethod
    def map_severity(severity: str) -> str:
        """Map Grafana severity to incident system severity"""
        severity_map = {
            'critical': 'critical',
            'warning': 'high',
            'info': 'medium',
        }
        return severity_map.get(severity.lower(), 'medium')

    @staticmethod
    def get_cached_incident(alert_name: str, database: str) -> Optional[str]:
        """Get cached incident ID for deduplication"""
        cache_key = f"{alert_name}_{database}"
        cached = INCIDENT_CACHE.get(cache_key)

        if cached:
            # Check if cache entry is expired
            if datetime.utcnow() - cached['created_at'] < timedelta(seconds=CACHE_EXPIRY):
                return cached['incident_id']
            else:
                # Remove expired entry
                del INCIDENT_CACHE[cache_key]

        return None


def parse_grafana_alert(data: Dict[str, Any]) -> Dict[str, Any]:
    """Parse Grafana alert webhook data"""

    # Handle different alert formats
    alert_data = {
        'title': data.get('title') or data.get('AlertTitle', 'Unknown Alert'),
        'description': data.get('description') or data.get('AlertDescription', ''),
        'severity': data.get('severity', data.get('Severity', 'warning')).lower(),
        'database': data.get('database', data.get('Labels', {}).get('database', 'unknown')),
        'alert_name': data.get('alert_name', data.get('AlertName', '')),
        'value': data.get('value', data.get('ValueString', 'N/A')),
        'threshold': data.get('threshold', 'N/A'),
        'alert_rule': data.get('alert_rule', ''),
        'dashboard_url': data.get('dashboard_url', data.get('DashboardURL', '')),
        'runbook_url': data.get('runbook_url', data.get('RulesURL', '')),
        'timestamp': data.get('timestamp', datetime.utcnow().isoformat()),
    }

    return alert_data


@app.route('/webhook/incident', methods=['POST', 'OPTIONS'])
def receive_alert_webhook():
    """
    Receive alert webhook from Grafana

    Expected POST data:
    {
        "title": "Lock Contention Alert",
        "description": "Active locks > 10",
        "severity": "critical",
        "database": "production",
        "alert_name": "lock_contention_critical",
        "value": "12",
        "dashboard_url": "https://grafana.internal/d/lock-monitoring",
        "runbook_url": "https://docs.internal/runbooks/lock-contention.md"
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

        logger.info(f"Received alert webhook: {data.get('title', 'Unknown')}")

        # Parse alert data
        alert_data = parse_grafana_alert(data)

        # Check for duplicate incidents
        cached_incident = IncidentManager.get_cached_incident(
            alert_data['alert_name'],
            alert_data['database']
        )

        if cached_incident:
            logger.info(f"Incident already exists in cache: {cached_incident}")
            return jsonify({
                "status": "skipped",
                "message": "Incident already exists",
                "incident_id": cached_incident
            }), 200

        # Create incident
        incident_id = IncidentManager.create_incident(alert_data)

        if incident_id:
            return jsonify({
                "status": "success",
                "incident_id": incident_id,
                "alert_title": alert_data['title'],
                "severity": alert_data['severity']
            }), 201
        else:
            return jsonify({
                "status": "error",
                "message": "Failed to create incident"
            }), 500

    except Exception as e:
        logger.error(f"Webhook error: {str(e)}", exc_info=True)
        return jsonify({
            "status": "error",
            "message": str(e)
        }), 500


@app.route('/webhook/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "pgAnalytics Webhook Receiver",
        "timestamp": datetime.utcnow().isoformat(),
        "incidents_cached": len(INCIDENT_CACHE)
    }), 200


@app.route('/webhook/metrics', methods=['GET'])
def metrics():
    """Get webhook metrics"""
    return jsonify({
        "incident_cache_size": len(INCIDENT_CACHE),
        "timestamp": datetime.utcnow().isoformat(),
        "uptime_seconds": 0,  # Calculate if needed
    }), 200


@app.route('/webhook/cache', methods=['DELETE'])
def clear_cache():
    """Clear incident cache (admin endpoint)"""
    try:
        count = len(INCIDENT_CACHE)
        INCIDENT_CACHE.clear()
        logger.info(f"Cache cleared. Removed {count} entries")
        return jsonify({
            "status": "success",
            "entries_cleared": count
        }), 200
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 500


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
    logger.info("pgAnalytics Webhook Receiver Starting")
    logger.info("=" * 60)
    logger.info(f"Host: {CONFIG['host']}")
    logger.info(f"Port: {CONFIG['port']}")
    logger.info(f"Incident API: {CONFIG['incident_api_url']}")
    logger.info("=" * 60)

    # Validate configuration
    if not CONFIG['incident_api_token']:
        logger.warning("INCIDENT_TRACKING_TOKEN not set - authentication may fail")

    try:
        app.run(
            host=CONFIG['host'],
            port=CONFIG['port'],
            debug=os.getenv('FLASK_DEBUG', 'false').lower() == 'true',
            use_reloader=False
        )
    except KeyboardInterrupt:
        logger.info("Webhook receiver stopped")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Failed to start webhook receiver: {str(e)}")
        sys.exit(1)


if __name__ == '__main__':
    main()
