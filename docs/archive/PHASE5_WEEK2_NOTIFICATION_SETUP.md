# Phase 5 Week 2: Notification Channel Setup & Integration

**Date**: March 3, 2026 (Starting Week 2)
**Phase**: 5 - Alerting & Automation
**Week**: 2 - Notification Integration
**Status**: IN PROGRESS

---

## Overview

Week 2 focuses on setting up and testing all 9 notification channels configured in Week 1. This includes Slack webhooks, PagerDuty integration, email configuration, and webhook endpoints.

### Week 2 Objectives

✅ Setup Slack notification channels
✅ Configure Slack webhooks in Grafana
✅ Test Slack alert delivery
✅ Setup PagerDuty integration
✅ Configure escalation policies
✅ Test PagerDuty incident creation
✅ Configure email/SMTP settings
✅ Test email notifications
✅ Setup webhook endpoints
✅ Test webhook delivery
✅ Verify all notification paths

---

## 1. Slack Integration Setup

### 1.1 Create Slack Workspace (If Not Exists)

```bash
# If you don't have a Slack workspace:
# Go to: https://slack.com
# Click "Create a new workspace"
# Follow setup wizard
```

### 1.2 Create Slack Channels

Create three channels for pgAnalytics alerts:

#### Channel 1: #critical-alerts

```bash
# In Slack Workspace:
# 1. Click "+" next to "Channels"
# 2. Select "Create a new channel"
# 3. Name: critical-alerts
# 4. Description: "Critical pgAnalytics alerts - requires immediate action"
# 5. Make Private: No
# 6. Create Channel

# Set Channel Topic:
# /topic Critical database alerts requiring immediate action
```

#### Channel 2: #database-alerts

```bash
# 1. Create new channel
# 2. Name: database-alerts
# 3. Description: "Database warning alerts - action required"
# 4. Make Public
# 5. Create Channel

# Set Channel Topic:
# /topic PostgreSQL database warning alerts
```

#### Channel 3: #database-info

```bash
# 1. Create new channel
# 2. Name: database-info
# 3. Description: "Database information and notifications"
# 4. Make Public
# 5. Create Channel

# Set Channel Topic:
# /topic Database information, schema changes, maintenance notices
```

### 1.3 Create Slack Webhook URLs

#### For #critical-alerts Channel

```bash
# 1. Go to: https://api.slack.com/apps
# 2. Click "Create New App"
# 3. Choose "From scratch"
# 4. App name: "pgAnalytics Critical Alerts"
# 5. Workspace: [Your workspace]
# 6. Create App

# 7. In App settings:
#    - Left sidebar: "Incoming Webhooks"
#    - Toggle "Activate Incoming Webhooks" to On
#    - Click "Add New Webhook to Workspace"
#    - Select Channel: #critical-alerts
#    - Authorize

# 8. Copy the Webhook URL
# Format: https://hooks.slack.com/services/T.../B.../...
# Save as: SLACK_WEBHOOK_CRITICAL
```

#### For #database-alerts Channel

```bash
# Repeat same process:
# 1. Create another webhook for #database-alerts
# 2. Save as: SLACK_WEBHOOK_WARNING
```

#### For #database-info Channel

```bash
# Repeat same process:
# 1. Create another webhook for #database-info
# 2. Save as: SLACK_WEBHOOK_INFO
```

### 1.4 Store Webhook URLs Securely

```bash
# Option A: Environment Variables (Recommended)
export SLACK_WEBHOOK_CRITICAL="https://hooks.slack.com/services/T.../B.../..."
export SLACK_WEBHOOK_WARNING="https://hooks.slack.com/services/T.../B.../..."
export SLACK_WEBHOOK_INFO="https://hooks.slack.com/services/T.../B.../..."

# Option B: Grafana Secrets
# Grafana → Settings → Environment variables
# Add each webhook URL

# Option C: .env File (for local testing)
# Create .env file with:
# SLACK_WEBHOOK_CRITICAL=...
# SLACK_WEBHOOK_WARNING=...
# SLACK_WEBHOOK_INFO=...
```

### 1.5 Configure Slack Webhooks in Grafana

#### Step 1: Access Notification Channels

```
Grafana → Settings (gear icon) → Alerting → Notification channels
```

#### Step 2: Create Critical Alerts Channel

```
1. Click "New channel"
2. Name: "Slack - Critical Alerts"
3. Type: "Slack"
4. Webhook URL: ${SLACK_WEBHOOK_CRITICAL}
5. Channel: #critical-alerts
6. Username: pgAnalytics Alerts
7. Icon Emoji: :warning:
8. Mention Users: @here
9. Message Template:

   Title: [🔴 CRITICAL] {{ .AlertTitle }}

   Text: {{ .AlertDescription }}
   Database: {{ .Labels.database }}
   Severity: {{ .Labels.severity }}
   Time: {{ .StartsAt }}
   Value: {{ .ValueString }}

   <{{ .DashboardURL }}|View Dashboard> | <{{ .RulesURL }}|View Runbook>

10. Send reminder: Yes (every 15 minutes)
11. Upload image: Yes
12. Save
```

#### Step 3: Create Warning Alerts Channel

```
1. Click "New channel"
2. Name: "Slack - Warning Alerts"
3. Type: "Slack"
4. Webhook URL: ${SLACK_WEBHOOK_WARNING}
5. Channel: #database-alerts
6. Username: pgAnalytics Alerts
7. Icon Emoji: :warning:
8. Message Template:

   Title: [🟠 WARNING] {{ .AlertTitle }}

   Text: {{ .AlertDescription }}
   Database: {{ .Labels.database }}
   Severity: {{ .Labels.severity }}
   Value: {{ .ValueString }}

   <{{ .DashboardURL }}|View Dashboard> | <{{ .RulesURL }}|View Runbook>

9. Send reminder: Yes (every 1 hour)
10. Save
```

#### Step 4: Create Info Notifications Channel

```
1. Click "New channel"
2. Name: "Slack - Info Notifications"
3. Type: "Slack"
4. Webhook URL: ${SLACK_WEBHOOK_INFO}
5. Channel: #database-info
6. Username: pgAnalytics Info
7. Icon Emoji: :information_source:
8. Message Template:

   Title: [ℹ️ INFO] {{ .AlertTitle }}

   Text: {{ .AlertDescription }}
   Database: {{ .Labels.database }}

   <{{ .DashboardURL }}|View Dashboard>

9. Send reminder: No
10. Save
```

### 1.6 Test Slack Integration

#### Test Critical Alert Channel

```bash
# In Grafana:
# Settings → Alerting → Notification channels → "Slack - Critical Alerts" → Test

# Expected:
# - Message appears in #critical-alerts within 5 seconds
# - Message includes [🔴 CRITICAL] prefix
# - Message is red colored
# - Dashboard link is included
```

#### Test Warning Alert Channel

```bash
# Settings → Alerting → Notification channels → "Slack - Warning Alerts" → Test

# Expected:
# - Message appears in #database-alerts
# - Message includes [🟠 WARNING] prefix
# - Message is orange colored
```

#### Test Info Channel

```bash
# Settings → Alerting → Notification channels → "Slack - Info Notifications" → Test

# Expected:
# - Message appears in #database-info
# - Message is blue colored
# - No action mention
```

---

## 2. PagerDuty Integration Setup

### 2.1 Create PagerDuty Service

#### Step 1: Login to PagerDuty

```
Go to: https://www.pagerduty.com
Login to your account (or create free trial)
```

#### Step 2: Create New Service

```
1. Services → New Service
2. Name: "PostgreSQL Monitoring"
3. Description: "pgAnalytics alerts for PostgreSQL databases"
4. Escalation Policy: [Create new or select existing]
5. Status: Active
6. Create Service
```

#### Step 3: Add Grafana Integration

```
1. On Service page: "Integrations" tab
2. Click "Add an integration"
3. Integration Type: Select "Grafana" (or Webhook if not available)
4. Integration Name: "Grafana Alerts"
5. Create Integration
6. Copy Integration Key (API key)
7. Save as: PAGERDUTY_INTEGRATION_KEY
```

### 2.2 Create Escalation Policies

#### Critical Escalation Policy

```
1. Settings → Escalation Policies
2. Click "New Escalation Policy"
3. Name: "Database - Critical"
4. First Escalation Rule:
   - Escalate after: 5 minutes
   - Escalate to: [On-call DBA user or schedule]
   - If no one acknowledges

5. Second Escalation Rule:
   - Escalate after: 15 minutes (from start)
   - Escalate to: [Senior DBA or team lead]
   - If no one acknowledges

6. Third Escalation Rule:
   - Escalate after: 30 minutes (from start)
   - Escalate to: [Database Manager]
   - If no one acknowledges

7. Create Policy
```

#### Standard Escalation Policy

```
1. Create new Escalation Policy
2. Name: "Database - Standard"
3. First Escalation Rule:
   - Escalate after: 15 minutes
   - Escalate to: [On-call DBA]

4. Second Escalation Rule:
   - Escalate after: 30 minutes
   - Escalate to: [Team Lead]

5. Create Policy
```

### 2.3 Configure PagerDuty in Grafana

#### Create Critical Channel

```
1. Grafana → Settings → Alerting → Notification channels
2. Click "New channel"
3. Name: "PagerDuty - Critical"
4. Type: "PagerDuty"
5. Integration Key: ${PAGERDUTY_INTEGRATION_KEY}
6. Severity: "critical"
7. Custom Details:
   {
     "alert": "{{ .AlertTitle }}",
     "description": "{{ .AlertDescription }}",
     "database": "{{ .Labels.database }}",
     "value": "{{ .ValueString }}",
     "dashboard": "{{ .DashboardURL }}"
   }
8. Client: "pgAnalytics"
9. Client URL: https://grafana.internal/d/system-overview
10. Save
```

#### Create Warning Channel

```
1. Click "New channel"
2. Name: "PagerDuty - Warning"
3. Type: "PagerDuty"
4. Integration Key: ${PAGERDUTY_INTEGRATION_KEY}
5. Severity: "warning"
6. Save
```

### 2.4 Test PagerDuty Integration

```bash
# In Grafana:
# Settings → Alerting → Notification channels → "PagerDuty - Critical" → Test

# Expected:
# - PagerDuty incident created in service
# - Severity: high/critical
# - Assignment: follows escalation policy
# - Alert details visible in incident

# Check:
# 1. PagerDuty incident appears
# 2. Correct escalation policy assigned
# 3. On-call user notified
# 4. Can acknowledge incident in PagerDuty
```

---

## 3. Email Configuration

### 3.1 Configure SMTP Settings

#### Edit Grafana Configuration

```bash
# File: /etc/grafana/grafana.ini (or grafana.ini in your installation)

[smtp]
enabled = true
host = smtp.your-domain.com:587
user = alerts@your-domain.com
password = your_secure_password
from_address = alerts@your-domain.com
from_name = pgAnalytics Alerts
skip_verify = false
startTLS_policy = "MandatoryStartTLS"

[emails]
templates_pattern = emails/*.html
content_types = text/html

# Restart Grafana
systemctl restart grafana-server
```

#### Common SMTP Settings

**Gmail**:
```
host = smtp.gmail.com:587
user = your-email@gmail.com
password = your-app-password  # Use app password, not Gmail password
```

**Office 365**:
```
host = smtp.office365.com:587
user = your-email@company.com
password = your-password
```

**AWS SES**:
```
host = email-smtp.region.amazonaws.com:587
user = your-smtp-username
password = your-smtp-password
```

**Custom Mail Server**:
```
host = mail.your-company.com:587
user = alerts@your-company.com
password = your-password
```

### 3.2 Create Email Notification Channels

#### DBA Team Channel

```
1. Grafana → Settings → Alerting → Notification channels
2. Click "New channel"
3. Name: "Email - DBA Team"
4. Type: "Email"
5. Email addresses:
   - dba-team@company.com
   - Lead DBA email (optional)
6. Single email: No (sends individual emails)
7. Include image: Yes (includes dashboard screenshot)
8. Message Template:

Subject: [{{ .GroupLabels.severity | upper }}] {{ .GroupLabels.alertname }}

Body:
Alert: {{ .AlertTitle }}
Database: {{ .Labels.database }}
Severity: {{ .Labels.severity }}
Time: {{ .StartsAt }}
Value: {{ .ValueString }}

Description:
{{ .AlertDescription }}

Actions:
- View Dashboard: {{ .DashboardURL }}
- View Runbook: {{ .RulesURL }}

9. Send reminder: Yes (every 1 hour)
10. Save
```

#### Operations Team Channel

```
1. Click "New channel"
2. Name: "Email - Operations Team"
3. Type: "Email"
4. Email addresses:
   - ops-team@company.com
   - operations@company.com
5. Single email: Yes (batches into one email)
6. Message Template:

Subject: pgAnalytics Daily Alert Summary - {{ .Date }}

Body:
Daily Alert Summary

The following alerts were triggered in the last 24 hours:

{{ range .Alerts }}
- {{ .AlertTitle }}: {{ .AlertDescription }}
  Database: {{ .Labels.database }}
  Time: {{ .StartsAt }}
{{ end }}

For detailed information, see the dashboard:
{{ .DashboardURL }}

7. Send reminder: No (daily digest only)
8. Save
```

### 3.3 Test Email Configuration

#### Test SMTP Connection

```bash
# In Grafana:
# Settings → Alerting → Notification channels → "Email - DBA Team" → Test

# Expected:
# - Test email appears in inbox within 2 minutes
# - Subject includes database name
# - Dashboard screenshot attached (if enabled)
# - All links working
```

#### Verify Email Delivery

```bash
# Check:
# 1. Email arrives in inbox (not spam)
# 2. All formatting correct
# 3. Links clickable and working
# 4. Images load properly
# 5. Sender address correct
```

---

## 4. Webhook Endpoints Setup

### 4.1 Incident Tracking Webhook

#### Create Webhook Receiver

```python
# File: monitoring/webhook_receiver.py
# Simple Flask app to receive webhooks

from flask import Flask, request, jsonify
import json
import requests
from datetime import datetime

app = Flask(__name__)

# Configuration
INCIDENT_TRACKING_URL = "https://incident-tracking.internal/api/incidents"
INCIDENT_TRACKING_TOKEN = "your-api-token"

@app.route('/webhook/incident', methods=['POST'])
def receive_incident_webhook():
    """Receive alert webhook and create incident"""

    try:
        data = request.json

        # Parse alert data
        incident = {
            "title": data.get('title', 'Unknown Alert'),
            "description": data.get('description', ''),
            "severity": data.get('severity', 'warning'),
            "service": "PostgreSQL",
            "tags": data.get('tags', []),
            "external_url": data.get('dashboard_url', ''),
            "timestamp": datetime.utcnow().isoformat(),
            "source": "pgAnalytics"
        }

        # Send to incident tracking system
        headers = {
            "Authorization": f"Bearer {INCIDENT_TRACKING_TOKEN}",
            "Content-Type": "application/json"
        }

        response = requests.post(
            INCIDENT_TRACKING_URL,
            json=incident,
            headers=headers
        )

        if response.status_code in [200, 201]:
            return jsonify({"status": "success", "incident_id": response.json().get('id')}), 201
        else:
            return jsonify({"status": "error", "message": response.text}), response.status_code

    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 500

@app.route('/webhook/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "healthy"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
```

#### Deploy Webhook Receiver

```bash
# Install dependencies
pip install flask requests

# Run webhook receiver
python monitoring/webhook_receiver.py

# Or use systemd service:
# Create /etc/systemd/system/pganalytics-webhook.service

[Unit]
Description=pgAnalytics Webhook Receiver
After=network.target

[Service]
Type=simple
User=postgres
WorkingDirectory=/opt/pganalytics
ExecStart=/usr/bin/python3 monitoring/webhook_receiver.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target

# Enable and start
systemctl enable pganalytics-webhook
systemctl start pganalytics-webhook
```

#### Configure Webhook in Grafana

```
1. Grafana → Settings → Alerting → Notification channels
2. Click "New channel"
3. Name: "Webhook - Incident Tracking"
4. Type: "Webhook"
5. URL: https://your-server:5000/webhook/incident
6. HTTP Method: POST
7. Username: (leave blank)
8. Password: (leave blank)
9. Custom HTTP Headers:
   Authorization: Bearer ${INCIDENT_TRACKING_TOKEN}

10. Save
```

### 4.2 JIRA Auto-Ticket Webhook

#### Create JIRA Webhook Receiver

```python
# File: monitoring/jira_webhook.py

from flask import Flask, request, jsonify
import requests
import json

app = Flask(__name__)

# Configuration
JIRA_URL = "https://jira.company.com"
JIRA_PROJECT = "DB"
JIRA_API_TOKEN = "your-api-token"
JIRA_USER = "alerts@company.com"

@app.route('/webhook/jira', methods=['POST'])
def create_jira_ticket():
    """Create JIRA ticket from alert webhook"""

    try:
        data = request.json

        # Only create for specific alerts
        alert_name = data.get('title', '')
        if not any(x in alert_name for x in ['Bloat', 'Cache', 'Connection']):
            return jsonify({"status": "skipped"}), 200

        # Prepare JIRA issue
        issue_data = {
            "fields": {
                "project": {"key": JIRA_PROJECT},
                "issuetype": {"name": "Task"},
                "summary": f"[DB-ALERT] {alert_name}",
                "description": f"{data.get('description', '')}\n\n" +
                             f"Database: {data.get('database', 'unknown')}\n" +
                             f"Value: {data.get('value', 'N/A')}\n" +
                             f"Dashboard: {data.get('dashboard_url', '')}\n" +
                             f"Runbook: {data.get('runbook_url', '')}",
                "priority": {"name": "High"},
                "labels": ["postgresql", "monitoring", data.get('severity', 'warning')]
            }
        }

        # Create issue in JIRA
        auth = (JIRA_USER, JIRA_API_TOKEN)
        headers = {"Content-Type": "application/json"}

        response = requests.post(
            f"{JIRA_URL}/rest/api/3/issue",
            json=issue_data,
            auth=auth,
            headers=headers
        )

        if response.status_code in [200, 201]:
            issue_id = response.json().get('key', 'unknown')
            return jsonify({
                "status": "success",
                "issue_id": issue_id,
                "issue_url": f"{JIRA_URL}/browse/{issue_id}"
            }), 201
        else:
            return jsonify({
                "status": "error",
                "message": response.text
            }), response.status_code

    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001)
```

#### Configure JIRA Webhook in Grafana

```
1. Grafana → Settings → Alerting → Notification channels
2. Click "New channel"
3. Name: "Webhook - JIRA Tickets"
4. Type: "Webhook"
5. URL: https://your-server:5001/webhook/jira
6. HTTP Method: POST
7. Custom HTTP Headers:
   Content-Type: application/json

8. Save
```

---

## 5. Alert Notification Routing

### 5.1 Configure Notification Policy

```
Grafana → Settings → Alerting → Alert rules → Notification policy

1. Default Route:
   Group by: [alertname, database, severity]
   Group wait: 10s
   Group interval: 10s
   Repeat interval: 1h

2. Critical Alert Route:
   Match: severity="critical"
   Receivers:
     - Slack - Critical Alerts
     - PagerDuty - Critical
     - Email - DBA Team
   Group wait: 0s (immediate)
   Repeat interval: 15m

3. Warning Alert Route:
   Match: severity="warning"
   Receivers:
     - Slack - Warning Alerts
     - Email - DBA Team
     - Webhook - JIRA Tickets
   Group wait: 5m
   Repeat interval: 1h

4. Info Alert Route:
   Match: severity="info"
   Receivers:
     - Slack - Info Notifications
   Group wait: 1h
   Repeat interval: 24h

5. Save Policy
```

---

## 6. Testing All Notification Channels

### 6.1 Test Slack Integration

```bash
# Test Critical Channel
curl -X POST ${SLACK_WEBHOOK_CRITICAL} \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "[🔴 TEST] Critical Alert Test",
    "attachments": [
      {
        "color": "#FF0000",
        "title": "Lock Contention Alert",
        "text": "Test message to verify critical alert channel"
      }
    ]
  }'

# Verify: Message appears in #critical-alerts within 5 seconds
```

### 6.2 Test PagerDuty Integration

```bash
# Check PagerDuty service
# 1. Go to PagerDuty dashboard
# 2. Services → PostgreSQL Monitoring
# 3. Click "Trigger Test Event"
# 4. Fill in test event details
# 5. Verify incident created and escalation triggered
```

### 6.3 Test Email Delivery

```bash
# In Grafana:
# Settings → Alerting → Notification channels
# Find "Email - DBA Team" channel
# Click "Test"

# Verify:
# 1. Email arrives in inbox within 2 minutes
# 2. Check spam folder if not found
# 3. Verify all content displays correctly
```

### 6.4 Test Webhook Integration

```bash
# Test Incident Tracking Webhook
curl -X POST https://your-server:5000/webhook/incident \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${INCIDENT_TRACKING_TOKEN}" \
  -d '{
    "title": "Test Alert - Lock Contention",
    "description": "Test alert to verify webhook delivery",
    "severity": "critical",
    "database": "production",
    "tags": ["test", "postgresql"],
    "dashboard_url": "https://grafana.internal/d/lock-monitoring"
  }'

# Expected: HTTP 201 response with incident_id
```

---

## 7. Verification Checklist

### Slack Integration
- [ ] 3 Slack channels created (#critical-alerts, #database-alerts, #database-info)
- [ ] 3 webhook URLs generated
- [ ] Webhooks added to Grafana notification channels
- [ ] Test message sent to each channel
- [ ] Messages display with correct color and format
- [ ] Links (dashboard, runbook) are clickable

### PagerDuty Integration
- [ ] PagerDuty service created
- [ ] Escalation policies configured (critical, standard)
- [ ] Integration key obtained and saved
- [ ] PagerDuty channels added to Grafana
- [ ] Test incident created
- [ ] Incident follows correct escalation policy
- [ ] On-call user received notification
- [ ] Incident can be acknowledged and resolved

### Email Configuration
- [ ] SMTP settings configured in grafana.ini
- [ ] Email channels created (DBA team, Operations)
- [ ] Test email sent successfully
- [ ] Email appears in inbox (not spam)
- [ ] All formatting and links work correctly
- [ ] Attachments (screenshots) render properly

### Webhook Integration
- [ ] Webhook receivers deployed and running
- [ ] Incident tracking webhook tested
- [ ] JIRA webhook tested
- [ ] Both receive correct POST requests
- [ ] Both create records in target systems
- [ ] Error handling works correctly

### Notification Routing
- [ ] Alert routing policy configured
- [ ] Critical alerts routed to immediate channels
- [ ] Warning alerts routed to non-immediate channels
- [ ] Info alerts routed to info channels
- [ ] Batching configured correctly
- [ ] Escalation timings appropriate

---

## 8. Troubleshooting

### Slack Webhook Not Working

**Issue**: Webhook URL invalid or expired

```bash
# Solution:
1. Verify webhook URL format: https://hooks.slack.com/services/T.../B.../...
2. Test webhook directly:
   curl -X POST <WEBHOOK_URL> \
     -H 'Content-Type: application/json' \
     -d '{"text":"Test message"}'
3. If 404: Webhook URL is invalid
4. Regenerate webhook URL in Slack app settings
```

**Issue**: Messages not appearing in channel

```bash
# Solution:
1. Verify channel name matches in Grafana config
2. Check Slack app permissions: Add to Workspace → Select correct channel
3. Check Slack app notification settings: App → Notification preferences
4. Verify webhook URL is for correct channel
```

### PagerDuty Integration Failed

**Issue**: Integration key invalid

```bash
# Solution:
1. Verify API key copied correctly
2. Check for trailing spaces in key
3. Regenerate API key in PagerDuty settings
4. Update key in Grafana
```

**Issue**: Incidents not being created

```bash
# Solution:
1. Verify service exists in PagerDuty
2. Check service has active escalation policy
3. Test event creation manually in PagerDuty UI
4. Check Grafana alert rule is firing
5. Check notification channel is selected in alert
```

### Email Not Being Delivered

**Issue**: Emails going to spam

```bash
# Solution:
1. Add pganalytics-alerts@yourdomain.com to contacts
2. Configure SPF/DKIM/DMARC records (if applicable)
3. Use corporate email server if available
4. Check mail server logs: /var/log/mail.log
```

**Issue**: SMTP connection refused

```bash
# Solution:
1. Verify SMTP host and port: telnet smtp.domain.com 587
2. Check credentials: username and password correct
3. Verify StartTLS is enabled if using port 587
4. Check firewall allows outbound SMTP
5. Verify app password (not regular password) for Gmail
```

### Webhook Receiver Errors

**Issue**: 500 error from webhook

```bash
# Solution:
1. Check webhook receiver logs
2. Verify request format matches expected JSON
3. Check authorization headers included
4. Test with curl: curl -X POST <webhook_url> -d '{"test":"data"}'
```

---

## 9. Week 2 Status Summary

### Completion Tracking

| Task | Status | Notes |
|------|--------|-------|
| Slack channels created | ⏳ IN PROGRESS | 3 channels needed |
| Slack webhooks configured | ⏳ IN PROGRESS | 3 webhooks in Grafana |
| Slack testing complete | 📋 NEXT | Test each channel |
| PagerDuty setup | ⏳ IN PROGRESS | Service + escalation |
| PagerDuty testing | 📋 NEXT | Test incident creation |
| Email SMTP config | ⏳ IN PROGRESS | Edit grafana.ini |
| Email testing | 📋 NEXT | Test delivery |
| Webhook receivers | 📋 NEXT | Deploy incident/JIRA |
| Webhook testing | 📋 NEXT | Test integration |
| Notification routing | 📋 NEXT | Configure policies |

### Week 2 Timeline

**Day 1-2**: Setup Slack (webhooks, channels, Grafana config)
**Day 3**: Setup PagerDuty (service, escalation, integration)
**Day 4**: Setup Email (SMTP, channels, testing)
**Day 5**: Setup Webhooks (deploy, configure, test)
**Day 6-7**: Test all channels, fix issues, document

---

## 10. Success Criteria

✅ All notification channels configured
✅ All channels tested successfully
✅ Critical alerts → PagerDuty (immediate)
✅ Warning alerts → Slack + Email (batched)
✅ Info alerts → Slack #database-info (hourly)
✅ Auto-escalation working
✅ JIRA tickets auto-creating
✅ Response times < target
✅ No false positives
✅ Team trained on alert procedures

---

**Week 2 Status**: IN PROGRESS
**Expected Completion**: When all notification channels tested and verified
**Next Phase**: Week 3 - Automation Implementation

---

Generated: March 3, 2026
Status: Week 2 detailed procedures and setup guide
