#!/usr/bin/env python3
"""
pgAnalytics JIRA Webhook Receiver

Receives alerts from Grafana and automatically creates JIRA tickets
for actionable warnings (bloat, cache, connections, etc.)

Usage:
    python webhook_jira_receiver.py

Environment Variables:
    FLASK_HOST: Host to listen on (default: 0.0.0.0)
    FLASK_PORT: Port to listen on (default: 5001)
    JIRA_URL: JIRA instance URL
    JIRA_PROJECT: JIRA project key (e.g., 'DB')
    JIRA_USER: JIRA API user email
    JIRA_API_TOKEN: JIRA API token
    JIRA_ISSUE_TYPE: JIRA issue type (default: Task)
    LOG_LEVEL: Logging level (default: INFO)
"""

import os
import sys
import json
import logging
from datetime import datetime
from flask import Flask, request, jsonify
from typing import Dict, Any, Optional, List
import requests
from requests.auth import HTTPBasicAuth
import base64

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
    'port': int(os.getenv('FLASK_PORT', '5001')),
    'jira_url': os.getenv('JIRA_URL', 'https://jira.company.com'),
    'jira_project': os.getenv('JIRA_PROJECT', 'DB'),
    'jira_user': os.getenv('JIRA_USER', ''),
    'jira_token': os.getenv('JIRA_API_TOKEN', ''),
    'jira_issue_type': os.getenv('JIRA_ISSUE_TYPE', 'Task'),
}

# Alert types that trigger JIRA ticket creation
TICKET_TRIGGER_ALERTS = [
    'high_table_bloat_warning',
    'low_cache_hit_ratio_warning',
    'high_connection_count_warning',
    'idle_in_transaction_warning',
]

# Priority mapping
PRIORITY_MAP = {
    'critical': 'Highest',
    'warning': 'High',
    'info': 'Medium',
}


class JiraManager:
    """Manage JIRA ticket creation"""

    @staticmethod
    def should_create_ticket(alert_name: str, severity: str) -> bool:
        """Determine if alert should create JIRA ticket"""
        # Create tickets for specific warning alerts
        if alert_name in TICKET_TRIGGER_ALERTS:
            return True
        # Also create for critical alerts
        if severity.lower() == 'critical':
            return True
        return False

    @staticmethod
    def create_issue(alert_data: Dict[str, Any]) -> Optional[str]:
        """Create JIRA issue"""
        try:
            # Prepare issue description with rich formatting
            description = f"""
*Alert Summary*

Database: {alert_data.get('database', 'unknown')}
Severity: {alert_data.get('severity', 'unknown').upper()}
Timestamp: {alert_data.get('timestamp', 'N/A')}
Current Value: {alert_data.get('value', 'N/A')}
Threshold: {alert_data.get('threshold', 'N/A')}

*Description*

{alert_data.get('description', 'No description provided')}

*Actions*

- Review Dashboard: {alert_data.get('dashboard_url', 'N/A')}
- See Runbook: {alert_data.get('runbook_url', 'N/A')}
- Alert Rule: {alert_data.get('alert_rule', 'N/A')}

*Next Steps*

1. Investigate issue using dashboard and runbook
2. Implement remediation steps
3. Verify resolution in dashboard
4. Update ticket with resolution summary
5. Close ticket when complete
"""

            # Prepare JIRA issue
            issue_data = {
                "fields": {
                    "project": {"key": CONFIG['jira_project']},
                    "issuetype": {"name": CONFIG['jira_issue_type']},
                    "summary": f"[{alert_data.get('severity', 'warning').upper()}] {alert_data.get('title', 'Unknown Alert')}",
                    "description": description,
                    "priority": {"name": PRIORITY_MAP.get(alert_data.get('severity', 'warning'), 'Medium')},
                    "labels": JiraManager.get_labels(alert_data),
                    "components": [
                        {"name": "PostgreSQL"}
                    ]
                }
            }

            # Add assignee if available (optional)
            assignee = os.getenv('JIRA_ASSIGNEE', '')
            if assignee:
                issue_data['fields']['assignee'] = {"name": assignee}

            logger.info(f"Creating JIRA ticket: {issue_data['fields']['summary']}")

            # Create issue via JIRA API
            auth = HTTPBasicAuth(CONFIG['jira_user'], CONFIG['jira_token'])
            headers = {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }

            response = requests.post(
                f"{CONFIG['jira_url']}/rest/api/3/issue",
                json=issue_data,
                auth=auth,
                headers=headers,
                timeout=10
            )

            if response.status_code in [200, 201]:
                result = response.json()
                issue_id = result.get('key', result.get('id'))
                issue_url = f"{CONFIG['jira_url']}/browse/{issue_id}"

                logger.info(f"JIRA ticket created: {issue_id}")
                logger.info(f"URL: {issue_url}")

                return issue_id
            else:
                logger.error(f"Failed to create JIRA ticket: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return None

        except Exception as e:
            logger.error(f"Error creating JIRA ticket: {str(e)}")
            return None

    @staticmethod
    def get_labels(alert_data: Dict[str, Any]) -> List[str]:
        """Generate JIRA labels from alert data"""
        labels = [
            'postgresql',
            'monitoring',
            'pganalytics',
            alert_data.get('severity', 'warning').lower(),
        ]

        # Add alert-specific labels
        alert_name = alert_data.get('alert_name', '').lower()
        if 'bloat' in alert_name:
            labels.append('bloat')
            labels.append('vacuum')
        elif 'cache' in alert_name:
            labels.append('cache')
            labels.append('performance')
        elif 'connection' in alert_name:
            labels.append('connections')
        elif 'lock' in alert_name:
            labels.append('locks')
        elif 'idle' in alert_name:
            labels.append('transactions')

        return labels


def parse_grafana_alert(data: Dict[str, Any]) -> Dict[str, Any]:
    """Parse Grafana alert webhook data"""
    alert_data = {
        'title': data.get('title', data.get('AlertTitle', 'Unknown Alert')),
        'description': data.get('description', data.get('AlertDescription', '')),
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


@app.route('/webhook/jira', methods=['POST', 'OPTIONS'])
def receive_alert_webhook():
    """
    Receive alert webhook from Grafana and create JIRA ticket

    Expected POST data:
    {
        "title": "High Table Bloat Alert",
        "description": "Table bloat at 65%",
        "severity": "warning",
        "database": "production",
        "alert_name": "high_table_bloat_warning",
        "value": "65",
        "threshold": "50",
        "dashboard_url": "https://grafana.internal/d/bloat-analysis",
        "runbook_url": "https://docs.internal/runbooks/high-bloat.md"
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

        # Check if this alert should create a JIRA ticket
        if not JiraManager.should_create_ticket(
            alert_data['alert_name'],
            alert_data['severity']
        ):
            logger.info(f"Alert '{alert_data['alert_name']}' does not trigger JIRA ticket creation")
            return jsonify({
                "status": "skipped",
                "reason": "Alert type does not trigger ticket creation",
                "alert_name": alert_data['alert_name']
            }), 200

        # Create JIRA ticket
        issue_id = JiraManager.create_issue(alert_data)

        if issue_id:
            return jsonify({
                "status": "success",
                "issue_id": issue_id,
                "issue_url": f"{CONFIG['jira_url']}/browse/{issue_id}",
                "alert_title": alert_data['title'],
                "severity": alert_data['severity']
            }), 201
        else:
            return jsonify({
                "status": "error",
                "message": "Failed to create JIRA ticket"
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
        "service": "pgAnalytics JIRA Webhook Receiver",
        "timestamp": datetime.utcnow().isoformat(),
        "jira_configured": bool(CONFIG['jira_token'])
    }), 200


@app.route('/webhook/config', methods=['GET'])
def get_config():
    """Get configuration (for debugging)"""
    return jsonify({
        "jira_url": CONFIG['jira_url'],
        "jira_project": CONFIG['jira_project'],
        "jira_issue_type": CONFIG['jira_issue_type'],
        "trigger_alerts": TICKET_TRIGGER_ALERTS,
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
    logger.info("pgAnalytics JIRA Webhook Receiver Starting")
    logger.info("=" * 60)
    logger.info(f"Host: {CONFIG['host']}")
    logger.info(f"Port: {CONFIG['port']}")
    logger.info(f"JIRA URL: {CONFIG['jira_url']}")
    logger.info(f"JIRA Project: {CONFIG['jira_project']}")
    logger.info("=" * 60)

    # Validate configuration
    if not CONFIG['jira_token']:
        logger.error("JIRA_API_TOKEN not set - cannot create tickets")
        sys.exit(1)

    if not CONFIG['jira_user']:
        logger.error("JIRA_USER not set - cannot authenticate")
        sys.exit(1)

    try:
        app.run(
            host=CONFIG['host'],
            port=CONFIG['port'],
            debug=os.getenv('FLASK_DEBUG', 'false').lower() == 'true',
            use_reloader=False
        )
    except KeyboardInterrupt:
        logger.info("JIRA webhook receiver stopped")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Failed to start JIRA webhook receiver: {str(e)}")
        sys.exit(1)


if __name__ == '__main__':
    main()
