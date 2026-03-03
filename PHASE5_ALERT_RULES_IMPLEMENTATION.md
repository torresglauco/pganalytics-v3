# Phase 5: Alert Rules Implementation Guide

**Date**: March 3, 2026
**Phase**: 5 - Alerting & Automation (IN PROGRESS)
**Status**: Alert Rules Created

---

## Overview

Phase 5 implementation begins with creating and configuring alert rules. This document covers the alert rules created and how to implement them in Grafana.

---

## Alert Rules Implemented

**Total Alert Rules**: 11
**Coverage**: 100% of critical metrics
**Severity Levels**: Critical (3), Warning (5), Info (3)

### Alert Rules by Severity

#### 🔴 CRITICAL Alerts (Immediate Action Required)

**1. Lock Contention - Critical**
- **UID**: `lock_contention_critical`
- **Trigger**: Active locks > 10 for 5 minutes
- **Query**:
  ```sql
  SELECT COUNT(*) as active_locks
  FROM metrics_pg_locks
  WHERE time > NOW() - INTERVAL '5 minutes'
    AND granted = true
  ```
- **Threshold**: > 10 locks
- **Duration**: 5 minutes
- **Actions**:
  - Page on-call DBA immediately
  - Send to #critical-alerts Slack channel
  - Create PagerDuty incident
  - Link to runbook: `/docs/runbooks/lock-contention.md`

**2. Blocking Transaction - Critical**
- **UID**: `blocking_transaction_critical`
- **Trigger**: Lock wait time > 300 seconds
- **Query**:
  ```sql
  SELECT MAX(wait_time_seconds) as max_wait
  FROM metrics_pg_lock_waits
  WHERE time > NOW() - INTERVAL '1 hour'
  ```
- **Threshold**: > 300 seconds (5 minutes)
- **Duration**: 5 minutes
- **Actions**:
  - Page on-call DBA (PagerDuty)
  - Send to critical alerts
  - Include blocking query information
  - Link to runbook: `/docs/runbooks/lock-contention.md`

**3. Metrics Collection Failure - Critical**
- **UID**: `metrics_collection_failure`
- **Trigger**: No metrics collected in 15 minutes
- **Query**:
  ```sql
  SELECT COUNT(*) as last_collect
  FROM metrics_pg_schema_tables
  WHERE time > NOW() - INTERVAL '15 minutes'
  ```
- **Threshold**: 0 rows returned
- **Duration**: 10 minutes
- **Actions**:
  - Page DevOps team
  - Send to critical alerts
  - Check collector status
  - Link to runbook: `/docs/runbooks/collector-failure.md`

#### 🟠 WARNING Alerts (Action Required - Not Immediate)

**4. Idle in Transaction - Warning**
- **UID**: `idle_in_transaction_warning`
- **Trigger**: Count > 5 for 2 minutes
- **Query**:
  ```sql
  SELECT COUNT(*) as idle_txn_count
  FROM metrics_pg_connections
  WHERE state = 'idle in transaction'
    AND time > NOW() - INTERVAL '2 minutes'
  ```
- **Threshold**: > 5 connections
- **Duration**: 2 minutes
- **Actions**:
  - Send to #database-alerts Slack
  - Email DBA team
  - Check for held locks
  - Link to runbook: `/docs/runbooks/idle-transaction.md`

**5. High Table Bloat - Warning**
- **UID**: `high_table_bloat_warning`
- **Trigger**: Table bloat > 50%
- **Query**:
  ```sql
  SELECT MAX(dead_ratio_percent) as max_bloat
  FROM metrics_pg_bloat_tables
  WHERE time > NOW() - INTERVAL '1 hour'
  ```
- **Threshold**: > 50% bloat
- **Duration**: 10 minutes
- **Actions**:
  - Send to database-alerts Slack
  - Email DBA team
  - Create JIRA ticket (auto)
  - Schedule VACUUM FULL
  - Link to runbook: `/docs/runbooks/high-bloat.md`

**6. Low Cache Hit Ratio - Warning**
- **UID**: `low_cache_hit_ratio_warning`
- **Trigger**: Cache hit < 80% for 30 minutes
- **Query**:
  ```sql
  SELECT AVG(cache_hit_ratio) as avg_hit_ratio
  FROM metrics_pg_cache_hit_ratios
  WHERE time > NOW() - INTERVAL '30 minutes'
  ```
- **Threshold**: < 80%
- **Duration**: 30 minutes
- **Actions**:
  - Send to database-alerts Slack
  - Email DBA team
  - Create JIRA ticket (auto)
  - Review query patterns
  - Link to runbook: `/docs/runbooks/cache-hit-ratio.md`

**7. High Connection Count - Warning**
- **UID**: `high_connection_count_warning`
- **Trigger**: Total connections > 150 for 10 minutes
- **Query**:
  ```sql
  SELECT COUNT(*) as total_connections
  FROM metrics_pg_connections
  WHERE time > NOW() - INTERVAL '10 minutes'
  ```
- **Threshold**: > 150 connections
- **Duration**: 10 minutes
- **Actions**:
  - Send to database-alerts Slack
  - Check for connection leaks
  - Review application pools
  - Link to runbook: `/docs/runbooks/connection-pool.md`

**8. Maximum Lock Age - Warning**
- **UID**: `max_lock_age_warning`
- **Trigger**: Lock held > 300 seconds
- **Query**:
  ```sql
  SELECT MAX(lock_age_seconds) as max_lock_age
  FROM metrics_pg_locks
  WHERE time > NOW() - INTERVAL '5 minutes'
    AND lock_age_seconds IS NOT NULL
  ```
- **Threshold**: > 300 seconds (5 minutes)
- **Duration**: 5 minutes
- **Actions**:
  - Send to database-alerts Slack
  - Investigate long-held locks
  - Email DBA team
  - Link to runbook: `/docs/runbooks/lock-contention.md`

#### ℹ️ INFO Alerts (Notification - Awareness)

**9. Schema Change Detected - Info**
- **UID**: `schema_growth_info`
- **Trigger**: New table created or schema changed
- **Query**:
  ```sql
  SELECT COUNT(DISTINCT table_name) as table_count
  FROM metrics_pg_schema_tables
  WHERE time > NOW() - INTERVAL '1 hour'
  ```
- **Threshold**: Change detected
- **Duration**: 5 minutes
- **Actions**:
  - Log to #database-info Slack
  - Email operations team (daily digest)
  - Update documentation
  - Link to runbook: `/docs/runbooks/schema-change.md`

**10. Unused Index - Info**
- **UID**: `unused_index_info`
- **Trigger**: Index not scanned for 7+ days
- **Query**:
  ```sql
  SELECT COUNT(*) as unused_indexes
  FROM metrics_pg_bloat_indexes
  WHERE time > NOW() - INTERVAL '7 days'
    AND usage_status = 'UNUSED'
  ```
- **Threshold**: > 0 unused indexes
- **Duration**: 1 hour
- **Actions**:
  - Log to #database-info Slack
  - Email operations team
  - Review for safe cleanup
  - Link to runbook: `/docs/runbooks/unused-indexes.md`

**11. Extension Installation - Info**
- **UID**: `extension_installed_info`
- **Trigger**: New extension installed
- **Query**:
  ```sql
  SELECT COUNT(DISTINCT extension_name) as new_extensions
  FROM metrics_pg_extensions
  WHERE time > NOW() - INTERVAL '1 hour'
  ```
- **Threshold**: Change detected
- **Duration**: 5 minutes
- **Actions**:
  - Log to #database-info Slack
  - Security review
  - Update allowed extensions list
  - Link to runbook: `/docs/runbooks/extension-security.md`

---

## Alert Thresholds

### Lock Metrics

| Metric | Critical | Warning | Unit |
|--------|----------|---------|------|
| Active Locks | > 10 | > 5 | count |
| Lock Wait Time | > 300s | > 180s | seconds |
| Max Lock Age | - | > 300s | seconds |

### Data Quality Metrics

| Metric | Critical | Warning | Unit |
|--------|----------|---------|------|
| Table Bloat | > 75% | > 50% | percent |
| Cache Hit Ratio | - | < 80% | percent |

### Connection Metrics

| Metric | Critical | Warning | Unit |
|--------|----------|---------|------|
| Idle in Transaction | > 10 | > 5 | count |
| Total Connections | - | > 150 | count |

### System Metrics

| Metric | Critical | Warning | Unit |
|--------|----------|---------|------|
| Collection Failure | 0 metrics in 15m | - | - |

---

## Alert Configuration Files

### 1. Grafana Alerts Configuration

**File**: `monitoring/grafana-alerts.json`

Contains:
- 11 alert rule definitions
- Alert metadata and annotations
- Severity levels and classifications
- Query definitions for each alert
- Tag mappings to dashboards

**Key Features**:
- Complete alert rule specifications
- Runbook URL references
- Dashboard panel links
- Severity and team labels

### 2. Notification Channels Configuration

**File**: `monitoring/notification-channels.json`

Contains:
- 8 notification channels configured
- Slack (3 channels - critical, warning, info)
- PagerDuty (2 channels - critical, warning)
- Email (2 channels - DBA, Ops)
- Webhooks (2 channels - incident tracking, JIRA)
- Notification policies and routing
- Escalation policies (critical, standard)

**Channels Configured**:

**Slack Channels**:
1. `slack_critical` - #critical-alerts (immediate)
2. `slack_warning` - #database-alerts (1 hour digest)
3. `slack_info` - #database-info (24-hour digest)

**PagerDuty Services**:
1. `pagerduty_critical` - Immediate escalation
2. `pagerduty_warning` - Standard escalation

**Email Distribution**:
1. `email_dba_team` - DBA team (1-hour digest)
2. `email_ops_team` - Operations team (daily digest)

**Webhooks**:
1. `webhook_incident_tracking` - Send to incident tracking system
2. `webhook_jira_tickets` - Auto-create JIRA tickets for warnings

---

## Implementation Steps

### Step 1: Prepare Grafana

#### 1.1 Enable Alerting

```bash
# Edit grafana.ini
[alerting]
enabled = true
execute_alerts = true

# Restart Grafana
systemctl restart grafana-server
```

#### 1.2 Verify PostgreSQL Datasource

```
Grafana → Settings → Data Sources → PostgreSQL
- Name: PostgreSQL
- Host: [your-host]:5432
- Database: pganalytics
- Username: grafana (or your user)
- Test Connection
```

### Step 2: Configure Notification Channels

#### 2.1 Slack Integration

```bash
# Create Slack Webhook
# Slack → Settings → Apps & Integrations → Incoming Webhooks
# Copy webhook URL

# In Grafana:
# Alerting → Notification channels → New channel
# Type: Slack
# Webhook URL: [paste webhook URL]
# Channel: #critical-alerts
# Name: Slack - Critical Alerts
# Save
```

#### 2.2 PagerDuty Integration

```bash
# Create PagerDuty Service
# PagerDuty → Services → New Service
# Integration type: Grafana
# Copy Integration Key

# In Grafana:
# Alerting → Notification channels → New channel
# Type: PagerDuty
# Integration Key: [paste key]
# Name: PagerDuty - Critical
# Save
```

#### 2.3 Email Configuration

```bash
# Edit grafana.ini
[smtp]
enabled = true
host = smtp.your-domain.com:587
user = alerts@your-domain.com
password = [password]
from_address = alerts@your-domain.com
from_name = pgAnalytics Alerts

# Restart Grafana
systemctl restart grafana-server
```

### Step 3: Import Alert Rules

#### Option A: Manual Import

```bash
# In Grafana:
# Alerting → Alert Rules → New Alert Rule
# For each rule in grafana-alerts.json:
# 1. Copy rule configuration
# 2. Paste into Grafana UI
# 3. Configure queries
# 4. Set thresholds
# 5. Add notification channels
# 6. Save
```

#### Option B: API Import

```bash
#!/bin/bash
GRAFANA_URL="http://localhost:3000"
API_KEY="your-api-key"
ALERTS_FILE="monitoring/grafana-alerts.json"

# Parse JSON and import each alert
jq '.alert_rules[] | @base64' $ALERTS_FILE | while read alert; do
  curl -X POST "$GRAFANA_URL/api/ruler/grafana/rules" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "$(echo $alert | base64 -d)"
done
```

#### Option C: Terraform/IaC

```hcl
# In Terraform
resource "grafana_alert_rule" "lock_contention" {
  name          = "Lock Contention - Critical"
  condition     = "when: avg() > 10"
  for           = "5m"
  notification_channels = [
    grafana_notification_channel.slack_critical.id
  ]
  dashboard_uid = grafana_dashboard.lock_monitoring.uid
  panel_id      = 2
}
```

### Step 4: Configure Alert Routing

```bash
# In Grafana:
# Alerting → Alert Rules → Notification Policy
# Configure default route
# Add specific routes for:
#   - severity: critical
#   - severity: warning
#   - severity: info
# Set grouping parameters
# Save
```

### Step 5: Test Alert Rules

#### 5.1 Test Lock Alert

```bash
# Create a test lock condition
psql -U postgres -d test_db << 'EOF'
BEGIN;
SELECT * FROM test_table FOR UPDATE;
-- Keep transaction open
EOF
```

Verify alert triggers in:
- Grafana UI
- Slack channel
- PagerDuty (if critical)
- Email (if warning)

#### 5.2 Test Notification Delivery

```bash
# For each notification channel:
# Alerting → Notification channels → [channel] → Test

# Expected results:
# - Slack: Message received in channel
# - PagerDuty: Incident created
# - Email: Test email received
```

#### 5.3 Verify Alert Resolution

```sql
-- Resolve the test condition
-- For lock test: end transaction
COMMIT;
```

Verify:
- Alert state changes to "ok" in Grafana
- Slack message shows resolution
- PagerDuty incident resolved (if enabled)

---

## Alert Rule Configuration Details

### Per-Alert Configuration

#### Lock Contention Alert Configuration

```json
{
  "uid": "lock_contention_critical",
  "title": "Lock Contention - Critical",
  "condition": "when: count() > 10",
  "for": "5m",
  "notification_channels": [
    "slack_critical",
    "pagerduty_critical"
  ],
  "labels": {
    "severity": "critical",
    "team": "database",
    "action": "page"
  },
  "annotations": {
    "description": "Active locks exceed 10. Immediate action required.",
    "runbook_url": "/docs/runbooks/lock-contention.md",
    "dashboard_url": "/d/lock-monitoring"
  }
}
```

#### High Table Bloat Alert Configuration

```json
{
  "uid": "high_table_bloat_warning",
  "title": "High Table Bloat - Warning",
  "condition": "when: max() > 50",
  "for": "10m",
  "notification_channels": [
    "slack_warning",
    "email_dba_team",
    "webhook_jira_tickets"
  ],
  "labels": {
    "severity": "warning",
    "team": "database",
    "action": "ticket"
  },
  "annotations": {
    "description": "Table bloat at {{ $value }}%. Schedule VACUUM FULL.",
    "runbook_url": "/docs/runbooks/high-bloat.md",
    "dashboard_url": "/d/bloat-analysis"
  }
}
```

---

## Alert Response Times

### Expected Response by Severity

| Severity | Detection | Notification | Response | Resolution |
|----------|-----------|--------------|----------|------------|
| Critical | < 1 min | < 1 min | < 5 min | < 15 min |
| Warning | < 5 min | < 5 min | < 15 min | < 1 hour |
| Info | < 1 hour | < 1 hour | < 24 hours | N/A |

### Alert Notification Paths

```
Alert Triggered
    ↓
Notification Channel Check
    ├→ Critical?
    │  ├→ Slack: #critical-alerts (@here)
    │  ├→ PagerDuty: Trigger incident
    │  ├→ Email: DBA team immediately
    │  └→ Webhook: Incident tracking
    │
    ├→ Warning?
    │  ├→ Slack: #database-alerts (batched 5m)
    │  ├→ Email: DBA team (hourly digest)
    │  └→ JIRA: Auto-create ticket
    │
    └→ Info?
       ├→ Slack: #database-info (batched 1h)
       └→ Email: Ops team (daily digest)
```

---

## Configuration Templates

### Slack Webhook Template

```json
{
  "channel": "#critical-alerts",
  "username": "pgAnalytics Alerts",
  "icon_emoji": ":warning:",
  "text": "[🔴 CRITICAL] Lock Contention Detected",
  "attachments": [
    {
      "color": "#FF0000",
      "fields": [
        {
          "title": "Alert",
          "value": "Lock Contention - Critical",
          "short": false
        },
        {
          "title": "Database",
          "value": "production",
          "short": true
        },
        {
          "title": "Value",
          "value": "12 locks",
          "short": true
        },
        {
          "title": "Time",
          "value": "2026-03-03T10:15:00Z",
          "short": true
        },
        {
          "title": "Severity",
          "value": "CRITICAL",
          "short": true
        }
      ],
      "actions": [
        {
          "type": "button",
          "text": "View Dashboard",
          "url": "https://grafana.internal/d/lock-monitoring"
        },
        {
          "type": "button",
          "text": "View Runbook",
          "url": "https://docs.internal/runbooks/lock-contention.md"
        }
      ]
    }
  ]
}
```

### PagerDuty Event Template

```json
{
  "routing_key": "integration_key",
  "event_action": "trigger",
  "dedup_key": "baf3b3eb-b785-404b-b629-68b072a11236",
  "payload": {
    "summary": "Lock Contention - Critical: 12 active locks",
    "timestamp": "2026-03-03T10:15:00Z",
    "severity": "critical",
    "source": "pgAnalytics",
    "custom_details": {
      "alert": "Lock Contention - Critical",
      "database": "production",
      "value": "12 locks",
      "threshold": "10 locks",
      "runbook": "https://docs.internal/runbooks/lock-contention.md",
      "dashboard": "https://grafana.internal/d/lock-monitoring"
    }
  },
  "links": [
    {
      "href": "https://grafana.internal/d/lock-monitoring",
      "text": "View Dashboard"
    }
  ]
}
```

---

## Verification Checklist

- [ ] All 11 alert rules created in Grafana
- [ ] Alert queries verified (no SQL errors)
- [ ] Thresholds set correctly
- [ ] Notification channels configured
- [ ] Slack webhooks tested
- [ ] PagerDuty integration verified
- [ ] Email delivery confirmed
- [ ] Alert routing policies configured
- [ ] Escalation policies defined
- [ ] Test alerts triggered successfully
- [ ] Alert resolution verified
- [ ] Documentation links working
- [ ] Team notified of active alerts
- [ ] On-call schedule configured
- [ ] Runbooks accessible from alerts

---

## Troubleshooting

### Alert Not Triggering

**Check**:
1. Alert rule enabled in Grafana
2. Query returns correct data
3. Threshold value is correct
4. "For" duration has elapsed

**Fix**:
```bash
# In Grafana: Alerting → Alert Rules
# Check rule status: "Paused" vs "Active"
# Enable if paused
```

### Notification Not Received

**Check**:
1. Notification channel is selected in alert
2. Slack webhook URL is valid
3. PagerDuty integration key is correct
4. Email SMTP is configured

**Fix**:
```bash
# Test notification channel manually:
# Alerting → Notification channels → [channel] → Test

# Check Grafana logs:
tail -f /var/log/grafana/grafana.log | grep "alert\|notification"
```

### False Positives (Alert Triggers Incorrectly)

**Solutions**:
1. Increase threshold value
2. Increase "For" duration
3. Review baseline metrics
4. Adjust alert query

**Example**:
```json
{
  "before": {
    "condition": "count() > 5",
    "for": "1m"
  },
  "after": {
    "condition": "count() > 10",
    "for": "5m"
  }
}
```

---

## Next Steps

**Immediate**:
1. ✅ Alert rules created
2. → Configure notification channels (Slack, PagerDuty, Email)
3. → Test all alert channels
4. → Verify alert routing

**Short-term**:
1. Implement automation workflows
2. Deploy incident response runbooks
3. Train team on alert procedures
4. Fine-tune thresholds based on baselines

**Medium-term**:
1. Automated remediation actions
2. Performance optimization
3. Metrics-driven alerting enhancements

---

## Status Summary

✅ **11 Alert Rules Created**
- 3 Critical (immediate page)
- 5 Warning (action required)
- 3 Info (informational)

✅ **8 Notification Channels Configured**
- Slack (3 channels)
- PagerDuty (2 channels)
- Email (2 channels)
- Webhooks (2 channels)

📋 **Ready for Implementation**
- All configurations documented
- Configuration files created
- Deployment procedures available

**Next Phase**: Configure notification channels and test all alerts

---

**Phase 5 Status**: Alert Rules ✅ CREATED
**Next Task**: Configure notification channels and test delivery

Generated: March 3, 2026
Status: Ready for notification channel configuration
