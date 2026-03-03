# Phase 5 Week 3 - Automation Implementation Guide

**Date**: March 3, 2026
**Phase**: 5 - Alerting & Automation (Week 3)
**Status**: Automation Engine Implementation

---

## Overview

Phase 5 Week 3 implements comprehensive automation and auto-remediation workflows for pgAnalytics v3. The system automatically responds to database alerts with intelligent remediation actions and incident correlation.

### Key Components

1. **Automation Engine** - Decides and executes remediation actions
2. **Incident Correlation Engine** - Groups related alerts into incidents
3. **Decision Trees** - Rules for when/how to remediate
4. **Remediation Actions** - Specific database operations
5. **Monitoring & Tracking** - History and status of automations

---

## Architecture Overview

```
Alert Stream (from Grafana)
    ↓
┌─────────────────────────────────────┐
│ Incident Correlation Engine         │
│ - Groups related alerts             │
│ - Tracks incident state             │
│ - Analyzes root causes              │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│ Automation Engine                   │
│ - Decides if action needed          │
│ - Executes remediation              │
│ - Tracks remediation history        │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│ Remediation Actions (6 types)       │
│ - Kill blocking locks               │
│ - Trigger VACUUM                    │
│ - Close idle connections            │
│ - Close idle-in-transaction         │
│ - Analyze cache issues              │
│ - Restart collectors                │
└─────────────────────────────────────┘
```

---

## Component 1: Automation Engine

### Purpose

The Automation Engine receives alerts and decides whether to automatically remediate based on:
- Alert type and severity
- Current alert frequency
- Previous remediation actions
- Database state

### Key Features

#### 1.1 Decision Logic

**Decision Tree for Lock Contention**:
```
IF lock_count > 10 AND duration > 5m
  AND no remediation in progress
  AND lock_age > 300s
THEN
  Execute: kill_blocking_locks
  Log: Blocking PIDs terminated
  Track: Remediation history
```

**Decision Tree for Table Bloat**:
```
IF bloat_ratio > 50%
  AND bloat_duration > 10m
  AND no VACUUM in progress
THEN
  Execute: trigger_vacuum
  Vacuum max 5 tables
  Track: Tables vacuumed
```

**Decision Tree for Connections**:
```
IF active_connections > 150
  AND idle_connections > 50
THEN
  Execute: close_idle_connections
  Max terminate 20 connections
  Preserve critical operations
```

#### 1.2 Remediation Actions

**Action 1: Kill Blocking Locks**
```python
Purpose: Terminate transactions blocking others
When: lock_contention_critical fires
Step 1: Query pg_locks for blocking relationships
Step 2: Identify blocking PIDs (not blocked ones)
Step 3: pg_terminate_backend() on blocking PIDs
Step 4: Log result and count
Status: Can kill up to all blocking processes
```

**Action 2: Trigger VACUUM**
```python
Purpose: Reclaim dead tuples and reduce bloat
When: high_table_bloat_warning fires
Step 1: Find tables with dead_ratio > 50%
Step 2: Order by dead space (descending)
Step 3: VACUUM ANALYZE on top 5 tables
Step 4: Commit transaction
Status: Success if at least 1 table vacuumed
```

**Action 3: Close Idle Connections**
```python
Purpose: Reduce connection pool pressure
When: high_connection_count_warning fires
Step 1: Find idle connections > 5 minutes old
Step 2: Order by idle start time (oldest first)
Step 3: pg_terminate_backend() on idle PIDs
Step 4: Max terminate 20 connections per action
Status: Success if at least 1 connection closed
```

**Action 4: Close Idle-in-Transaction**
```python
Purpose: Release locks held by idle transactions
When: idle_in_transaction_warning fires
Step 1: Find idle-in-transaction > 10 minutes
Step 2: Query and log transaction info
Step 3: pg_terminate_backend() on idle-txn PIDs
Step 4: Track terminated connections
Status: Success if at least 1 connection closed
```

**Action 5: Analyze Cache Optimization**
```python
Purpose: Analyze cache metrics and provide suggestions
When: low_cache_hit_ratio_warning fires
Step 1: Calculate overall cache hit ratio
Step 2: Find tables with low hit ratios
Step 3: Generate recommendations
Step 4: Log suggestions
Status: Always SUCCESS (no direct remediation)
```

**Action 6: Restart Collectors**
```python
Purpose: Restart stalled metrics collectors
When: metrics_collection_failure fires
Step 1: Queue restart request to collector manager
Step 2: Notify collector orchestration system
Step 3: Wait for collectors to restart
Step 4: Verify metrics collection resumes
Status: SUCCESS when restart request sent
```

#### 1.3 Configuration Options

```yaml
lock_contention_critical:
  threshold: 10 locks
  action: kill_blocking_locks
  max_age_seconds: 300
  enabled: true

high_table_bloat_warning:
  threshold: 50% bloat
  action: trigger_vacuum
  enabled: true
  tables_max: 5

high_connection_count_warning:
  threshold: 150 connections
  action: close_idle_connections
  enabled: true
  idle_seconds: 300
  max_terminate: 20

idle_in_transaction_warning:
  threshold: 5 connections
  action: close_idle_transactions
  enabled: true
  idle_seconds: 600

low_cache_hit_ratio_warning:
  threshold: 80%
  action: optimize_cache
  enabled: true

metrics_collection_failure:
  threshold: 15 minutes
  action: restart_collectors
  enabled: true
```

#### 1.4 Safety Features

**Dry-Run Mode**:
```bash
DRY_RUN=true python automation_engine.py
# Will log intended actions but not execute them
```

**Enable/Disable Auto-Remediation**:
```bash
AUTO_REMEDIATE=true python automation_engine.py
# Enable auto-remediation (default: false)
```

**Prevent Concurrent Remediation**:
- Check if remediation already in progress for alert
- Skip if same alert/database has active remediation
- Track remediation state in history

**Preserve Critical Operations**:
- Check connection user and application
- Don't kill critical system connections
- Skip remediation if risk is too high

---

## Component 2: Incident Correlation Engine

### Purpose

The Incident Correlation Engine groups related alerts into single incidents for better tracking and root cause analysis.

### Correlation Groups

**Group 1: Lock Contention** (Critical)
```
Alerts: lock_contention_critical
        blocking_transaction_critical
        max_lock_age_warning

Description: PostgreSQL locking issues causing transaction delays

Root Cause Analysis:
- Long-running transaction holding locks
- Unhandled exception leaving transaction open
- Application deadlock or lock escalation
- Batch operation without transaction timeout

Recommended Actions:
- Identify long-running transactions
- Check application logs for deadlocks
- Review query execution plans
- Implement application-level lock timeouts
```

**Group 2: Performance Issues** (High)
```
Alerts: low_cache_hit_ratio_warning
        high_table_bloat_warning

Description: Performance degradation from cache misses and table bloat

Root Cause Analysis:
- Missing indexes on frequently accessed tables
- Table autovacuum lagging behind write load
- Query plan changes due to statistics update
- Shared buffer pressure or cache thrashing

Recommended Actions:
- Run ANALYZE to update statistics
- Identify missing indexes
- Consider aggressive autovacuum tuning
- Profile top queries with pg_stat_statements
```

**Group 3: Connection Pool Issues** (High)
```
Alerts: high_connection_count_warning
        idle_in_transaction_warning

Description: Connection pool saturation and resource leaks

Root Cause Analysis:
- Connection pool not returning connections
- Application holding connections open
- Connection leak in ORM or driver
- Long-running transactions blocking connection release

Recommended Actions:
- Monitor connection state distribution
- Check application for connection pool leaks
- Verify application retry/timeout settings
- Review database session durations
```

**Group 4: System Health** (Critical)
```
Alerts: metrics_collection_failure

Description: Database monitoring system health issues

Root Cause Analysis:
- Collector process crashed or hung
- Database connectivity issue
- Collector resource exhaustion
- Database under heavy load

Recommended Actions:
- Check collector process status
- Verify database connection
- Monitor collector resource usage
- Review database slow query log
```

### Correlation Mechanism

**Signature Calculation**:
```
signature = MD5(database + severity + group_id)
```

**Incident Deduplication**:
- Alerts within 5-minute window (configurable) are correlated
- Same database + severity + group = same incident
- New alerts added to existing incident
- Incident timestamp updated on each new alert
- Alert count incremented

**Root Cause Analysis**:
- Automatic analysis based on group type
- Suggests possible root causes
- Recommends investigation and remediation actions
- Assigns confidence score (75-90%)

### Incident Lifecycle

```
ACTIVE (receiving alerts)
  ↓
ESCALATED (if severity increases)
  ↓
RESOLVED (marked by operator or auto-resolved)
  ↓
ARCHIVED (removed after 24 hours)
```

---

## Component 3: Decision Trees

### Lock Contention Decision Tree

```
Alert: lock_contention_critical
  │
  ├─ Check: Auto-remediation enabled?
  │   ├─ NO → Skip, alert only
  │   └─ YES → Continue
  │
  ├─ Check: Remediation already in progress?
  │   ├─ YES → Skip, let it finish
  │   └─ NO → Continue
  │
  ├─ Check: Lock age > 300 seconds?
  │   ├─ NO → Monitor, don't kill yet
  │   └─ YES → Continue
  │
  └─ Action: Kill blocking locks
      ├─ Step 1: Find blocking PIDs
      ├─ Step 2: Log blocking queries
      ├─ Step 3: Terminate blocking backends
      └─ Step 4: Track terminated PIDs
```

### Bloat Decision Tree

```
Alert: high_table_bloat_warning
  │
  ├─ Check: Auto-remediation enabled?
  │   ├─ NO → Skip, alert only
  │   └─ YES → Continue
  │
  ├─ Check: Bloat ratio > 50%?
  │   ├─ NO → Skip
  │   └─ YES → Continue
  │
  ├─ Check: Duration > 10 minutes?
  │   ├─ NO → Monitor, don't vacuum yet
  │   └─ YES → Continue
  │
  └─ Action: Trigger VACUUM
      ├─ Step 1: Find high-bloat tables
      ├─ Step 2: Order by bloat amount
      ├─ Step 3: VACUUM ANALYZE top 5 tables
      └─ Step 4: Track vacuumed tables
```

### Connection Decision Tree

```
Alert: high_connection_count_warning
  │
  ├─ Check: Auto-remediation enabled?
  │   ├─ NO → Skip, alert only
  │   └─ YES → Continue
  │
  ├─ Check: Connections > 150?
  │   ├─ NO → Skip
  │   └─ YES → Continue
  │
  ├─ Check: Idle connections available?
  │   ├─ NO → Skip (all connections in use)
  │   └─ YES → Continue
  │
  └─ Action: Close idle connections
      ├─ Step 1: Find idle > 5 min
      ├─ Step 2: Max terminate 20
      ├─ Step 3: Skip critical connections
      └─ Step 4: Track closed connections
```

---

## Deployment

### Installation

```bash
# Copy automation engines
cp monitoring/automation_engine.py /opt/pganalytics/automation/
cp monitoring/incident_correlation_engine.py /opt/pganalytics/automation/

# Install Python dependencies
pip install flask psycopg2-binary

# Make executable
chmod +x /opt/pganalytics/automation/automation_engine.py
chmod +x /opt/pganalytics/automation/incident_correlation_engine.py
```

### Environment Configuration

**Automation Engine** (`/etc/pganalytics/automation.env`):
```bash
FLASK_HOST=127.0.0.1
FLASK_PORT=5002
DATABASE_HOST=postgres.internal
DATABASE_PORT=5432
DATABASE_USER=pganalytics_admin
DATABASE_PASSWORD=<secure_password>
DATABASE_NAME=pganalytics
AUTO_REMEDIATE=false
DRY_RUN=true
LOG_LEVEL=INFO
```

**Incident Correlation Engine** (`/etc/pganalytics/correlation.env`):
```bash
FLASK_HOST=127.0.0.1
FLASK_PORT=5003
CORRELATION_WINDOW=300
INCIDENT_TRACKING_URL=http://localhost:5000/api/incidents
INCIDENT_TRACKING_TOKEN=Bearer_token_here
LOG_LEVEL=INFO
```

### Systemd Services

**File**: `/etc/systemd/system/pganalytics-automation.service`

```ini
[Unit]
Description=pgAnalytics Automation Engine
After=network.target
Requires=network.target

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics/automation
EnvironmentFile=/etc/pganalytics/automation.env
ExecStart=/usr/bin/python3 /opt/pganalytics/automation/automation_engine.py
Restart=always
RestartSec=5
StandardOutput=journal

[Install]
WantedBy=multi-user.target
```

**File**: `/etc/systemd/system/pganalytics-correlation.service`

```ini
[Unit]
Description=pgAnalytics Incident Correlation Engine
After=network.target
Requires=network.target

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics/automation
EnvironmentFile=/etc/pganalytics/correlation.env
ExecStart=/usr/bin/python3 /opt/pganalytics/automation/incident_correlation_engine.py
Restart=always
RestartSec=5
StandardOutput=journal

[Install]
WantedBy=multi-user.target
```

### Startup

```bash
# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable pganalytics-automation.service
sudo systemctl enable pganalytics-correlation.service
sudo systemctl start pganalytics-automation.service
sudo systemctl start pganalytics-correlation.service

# Verify running
sudo systemctl status pganalytics-automation.service
sudo systemctl status pganalytics-correlation.service

# View logs
sudo journalctl -u pganalytics-automation.service -f
sudo journalctl -u pganalytics-correlation.service -f
```

---

## API Endpoints

### Automation Engine

**POST /automation/remediate**
- Accept alert and execute remediation
- Request: Alert data (alert_name, severity, database, etc.)
- Response: Remediation status and details

**GET /automation/history**
- Get recent remediation actions
- Query params: limit, alert_name
- Response: List of remediation results

**GET /automation/health**
- Service health check
- Response: Service status and config

**GET /automation/config**
- Get remediation configuration
- Response: Enabled actions and thresholds

### Incident Correlation Engine

**POST /correlation/correlate**
- Correlate alert into incident
- Request: Alert data
- Response: Incident ID and correlation status

**GET /correlation/incident/<id>**
- Get incident details and analysis
- Response: Full incident with root cause analysis

**POST /correlation/incident/<id>/resolve**
- Mark incident as resolved
- Request: Resolution notes (optional)
- Response: Confirmation

**GET /correlation/incidents**
- List active and recent incidents
- Query params: state, limit
- Response: List of incidents

**GET /correlation/health**
- Service health check
- Response: Service status and incident count

**GET /correlation/config**
- Get correlation configuration
- Response: Correlation groups and settings

---

## Testing & Verification

### Test Scenario 1: Lock Contention

```bash
# Start automation engine in dry-run mode
AUTO_REMEDIATE=true DRY_RUN=true python automation_engine.py &

# Send test alert
curl -X POST http://localhost:5002/automation/remediate \
  -H 'Content-Type: application/json' \
  -d '{
    "alert_name": "lock_contention_critical",
    "severity": "critical",
    "database": "production",
    "value": "15",
    "threshold": "10",
    "timestamp": "2026-03-03T12:00:00Z"
  }'

# Check logs for [DRY-RUN] messages
# Verify no actual lock termination occurred
```

### Test Scenario 2: Incident Correlation

```bash
# Start correlation engine
python incident_correlation_engine.py &

# Send first alert
curl -X POST http://localhost:5003/correlation/correlate \
  -H 'Content-Type: application/json' \
  -d '{
    "alert_name": "lock_contention_critical",
    "severity": "critical",
    "database": "production"
  }'

# Response: {"status": "success", "is_new": true, "incident_id": "INC_..."}

# Send related alert within 5 minutes
curl -X POST http://localhost:5003/correlation/correlate \
  -H 'Content-Type: application/json' \
  -d '{
    "alert_name": "blocking_transaction_critical",
    "severity": "critical",
    "database": "production"
  }'

# Response: {"status": "success", "is_new": false, "incident_id": "INC_..."}
# Same incident_id as first alert

# Get incident analysis
curl http://localhost:5003/correlation/incident/INC_...
```

### Test Scenario 3: Root Cause Analysis

```bash
# Get incident summary with root cause
curl http://localhost:5003/correlation/incident/INC_123 | jq .root_cause_analysis

# Output shows:
# - Possible root causes (with descriptions)
# - Recommended actions
# - Confidence level (75-90%)
```

---

## Monitoring & Logging

### Key Metrics to Track

1. **Remediation Success Rate**
   - % of remediation actions that succeeded
   - Target: > 90%

2. **Time to Remediate**
   - Time between alert and remediation completion
   - Target: < 2 minutes

3. **False Positive Rate**
   - % of remediation actions that were unnecessary
   - Target: < 5%

4. **Incident Correlation Accuracy**
   - % of related alerts correctly grouped
   - Target: > 95%

### Log Monitoring

```bash
# Watch for remediation actions
sudo journalctl -u pganalytics-automation.service | grep "remediat"

# Watch for incident creation
sudo journalctl -u pganalytics-correlation.service | grep "incident"

# Watch for errors
sudo journalctl -u pganalytics-automation.service -p err
```

### Metrics Export

```bash
# Get recent remediation history
curl http://localhost:5002/automation/history?limit=100

# Get active incidents
curl http://localhost:5003/correlation/incidents?state=active

# Get correlation config
curl http://localhost:5003/correlation/config
```

---

## Safety & Compliance

### Risk Assessment

| Action | Risk Level | Mitigation |
|--------|-----------|------------|
| Kill locks | High | Test in dry-run, verify blocking PID |
| VACUUM | Medium | Only top 5 tables, analyze after |
| Close connections | Medium | Skip critical apps, max 20 per action |
| Close idle-txn | Medium | Only > 10 minutes old |
| Cache analysis | Low | Read-only, no database changes |
| Restart collectors | Low | Documented procedure |

### Approval Workflow

**Option 1: Automatic (Production)**
```
Alert → Auto-remediation enabled → Execute immediately
Time to remediate: < 2 minutes
```

**Option 2: Manual (Conservative)**
```
Alert → DRY_RUN=true → Review logs → Manual execution
Time to remediate: 10-30 minutes
```

**Option 3: Staged (Recommended)**
```
Week 1: DRY_RUN=true (only log actions)
Week 2: AUTO_REMEDIATE=false (wait for manual)
Week 3: AUTO_REMEDIATE=true (full automation)
```

---

## Troubleshooting

### Issue: Remediation Not Executing

**Symptoms**: Alerts received but no remediation

**Check**:
```bash
# 1. Verify auto-remediation is enabled
curl http://localhost:5002/automation/config | grep auto_remediate

# 2. Check if remediation already in progress
curl http://localhost:5002/automation/history | jq

# 3. Verify database connectivity
curl http://localhost:5002/automation/health

# 4. Check logs
sudo journalctl -u pganalytics-automation.service
```

### Issue: False Remediation

**Symptoms**: Remediation executed but made things worse

**Prevent**:
```bash
# 1. Always test in dry-run first
DRY_RUN=true ...

# 2. Enable dry-run initially
AUTO_REMEDIATE=false DRY_RUN=true ...

# 3. Monitor results before enabling
curl http://localhost:5002/automation/history
```

### Issue: Incident Correlation Not Working

**Symptoms**: Related alerts not grouped

**Check**:
```bash
# 1. Verify correlation groups
curl http://localhost:5003/correlation/config | jq

# 2. Check correlation window
# Default: 300 seconds (5 minutes)
# Verify alerts sent within window

# 3. Check incident state
curl http://localhost:5003/correlation/incidents | jq

# 4. Review correlation logs
sudo journalctl -u pganalytics-correlation.service
```

---

## Best Practices

1. **Start Conservative**
   - Enable `DRY_RUN=true` initially
   - Review all actions in logs
   - Only enable auto-remediation after testing

2. **Monitor Results**
   - Track remediation success/failure
   - Set up alerting on remediation failures
   - Review false positive rate

3. **Document Decisions**
   - Keep decision tree documentation updated
   - Document any custom thresholds
   - Track remediation approval chain

4. **Test Regularly**
   - Test each remediation action in staging
   - Verify database connectivity
   - Test with realistic alert data

5. **Escalate When Needed**
   - Set thresholds appropriately
   - Escalate to DBA/SRE for manual review
   - Don't rely 100% on automation

---

## Summary

Week 3 delivers:
- **Automation Engine**: 6 remediation actions (880+ lines)
- **Correlation Engine**: 4 correlation groups (750+ lines)
- **Decision Trees**: 6 detailed flowcharts for remediation
- **Deployment Guide**: Systemd, configuration, testing
- **API Documentation**: 10 endpoints across 2 services
- **Safety Features**: Dry-run mode, prevent concurrent actions

---

Generated: March 3, 2026
Status: Week 3 Implementation Complete
