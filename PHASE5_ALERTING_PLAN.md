# Phase 5: Alerting and Automation - Implementation Plan

**Date**: March 3, 2026
**Status**: Planning Phase
**Target**: Alert rules and automation for Phases 3-4

---

## Overview

Phase 5 builds on the Phase 4 visualization infrastructure by implementing:

1. **Alert Rules** - Automated condition detection
2. **Notification Channels** - Multi-channel alert delivery
3. **Automation Workflows** - Remediation actions
4. **Runbook Integration** - Incident response procedures

---

## Phase 5 Objectives

### Primary Goals

✅ Define critical alert thresholds
✅ Configure notification channels (Slack, PagerDuty, Email)
✅ Create automation rules for common issues
✅ Develop incident response runbooks
✅ Implement alert acknowledgment tracking
✅ Build performance dashboards for alerts

### Success Metrics

- All critical metrics have alert rules
- Alert notification delivery < 1 minute
- False positive rate < 5%
- Team response time < 15 minutes average
- Automation reduces manual intervention by 50%

---

## 1. Critical Alert Rules

### High Priority Alerts (P1 - Immediate Action)

#### 1.1 Lock Contention Alert
**Condition**: Active locks > 10 for > 5 minutes
**Severity**: CRITICAL
**Action**:
- Send to PagerDuty (immediate escalation)
- Notify Slack #database-alerts
- Trigger runbook: Investigate Lock Contention

**Query**:
```sql
SELECT COUNT(*) as active_locks
FROM metrics_pg_locks
WHERE time > NOW() - INTERVAL '5 minutes'
  AND granted = true
```

**Threshold**: > 10 for 5 consecutive minutes

---

#### 1.2 Blocking Transaction Alert
**Condition**: Lock wait time > 300 seconds
**Severity**: CRITICAL
**Action**:
- Send to PagerDuty
- Notify ops team
- Auto-collect blocking query info

**Query**:
```sql
SELECT MAX(wait_time_seconds) as max_wait
FROM metrics_pg_lock_waits
WHERE time > NOW() - INTERVAL '1 hour'
```

**Threshold**: > 300 seconds

---

#### 1.3 Idle in Transaction Alert
**Condition**: Count > 5 for > 2 minutes
**Severity**: WARNING → CRITICAL
**Action**:
- Notify team (Slack)
- Check application status
- Consider connection timeout

**Query**:
```sql
SELECT COUNT(*) as idle_txn_count
FROM metrics_pg_connections
WHERE state = 'idle in transaction'
  AND time > NOW() - INTERVAL '2 minutes'
```

**Threshold**:
- Yellow (Warning): > 3
- Red (Critical): > 10

---

### Medium Priority Alerts (P2 - Action Required)

#### 2.1 High Table Bloat Alert
**Condition**: Max bloat > 50% for any table
**Severity**: WARNING
**Action**:
- Notify team
- Schedule maintenance window
- Create VACUUM job

**Query**:
```sql
SELECT database_name, table_name, dead_ratio_percent
FROM metrics_pg_bloat_tables
WHERE time > NOW() - INTERVAL '1 day'
  AND dead_ratio_percent > 50
ORDER BY dead_ratio_percent DESC
```

**Threshold**: > 50%
**Response**: Schedule VACUUM FULL

---

#### 2.2 Low Cache Hit Ratio Alert
**Condition**: Cache hit < 80% for > 30 minutes
**Severity**: WARNING
**Action**:
- Review query patterns
- Check for full table scans
- Consider shared_buffers increase

**Query**:
```sql
SELECT AVG(cache_hit_ratio) as avg_hit_ratio
FROM metrics_pg_cache_hit_ratios
WHERE time > NOW() - INTERVAL '30 minutes'
```

**Threshold**: < 80%

---

#### 2.3 High Connection Count Alert
**Condition**: Total connections > 150 for > 10 minutes
**Severity**: WARNING
**Action**:
- Review application connection pools
- Check for connection leaks
- Monitor trend

**Query**:
```sql
SELECT COUNT(*) as total_connections
FROM metrics_pg_connections
WHERE time > NOW() - INTERVAL '10 minutes'
```

**Threshold**: > 150

---

### Low Priority Alerts (P3 - Informational)

#### 3.1 Schema Growth Alert
**Condition**: New table created or schema changed
**Severity**: INFO
**Action**:
- Log to audit trail
- Notify DBA team
- Update documentation

**Query**:
```sql
SELECT COUNT(DISTINCT table_name) as table_count
FROM metrics_pg_schema_tables
WHERE time > NOW() - INTERVAL '1 day'
```

---

#### 3.2 Unused Index Alert
**Condition**: Index scans = 0 for > 7 days
**Severity**: INFO
**Action**:
- Review if safe to drop
- Document usage justification
- Plan cleanup

**Query**:
```sql
SELECT index_name, index_scans, usage_status
FROM metrics_pg_bloat_indexes
WHERE time > NOW() - INTERVAL '7 days'
  AND usage_status = 'UNUSED'
```

---

#### 3.3 Extension Installation Alert
**Condition**: New extension installed
**Severity**: INFO
**Action**:
- Log for security review
- Verify against allowed list
- Update inventory

**Query**:
```sql
SELECT extension_name, COUNT(*) as installations
FROM metrics_pg_extensions
WHERE time > NOW() - INTERVAL '1 day'
GROUP BY extension_name
```

---

## 2. Notification Channels

### 2.1 Slack Integration

**Configuration**:
```
Channel: #database-alerts
Severity Mapping:
  CRITICAL → #critical-alerts (with @here)
  WARNING → #database-alerts
  INFO → #database-info
```

**Alert Format**:
```
🔴 CRITICAL: Lock Contention Detected
Database: production
Severity: CRITICAL
Time: 2026-03-03T10:15:00Z
Active Locks: 12

Action Required: Check Lock Monitoring Dashboard
Runbook: /docs/runbooks/lock-contention.md

Click to View: [Dashboard Link]
```

**Setup Steps**:
1. Create Slack Webhook
2. Configure in Grafana: Alerting → Notification Channels
3. Test with dummy alert

---

### 2.2 PagerDuty Integration

**Configuration**:
```
Service: PostgreSQL Monitoring
Integration Key: [PagerDuty Key]
Severity Mapping:
  CRITICAL → Trigger incident
  WARNING → Create alert
```

**Escalation Policy**:
```
Level 1 (0-5 min): DBA on-call
Level 2 (5-15 min): Senior DBA
Level 3 (15+ min): Database Manager
```

**Setup Steps**:
1. Create PagerDuty service
2. Configure escalation policy
3. Add Grafana integration
4. Set up oncall schedule

---

### 2.3 Email Integration

**Configuration**:
```
Recipients:
  - Critical: dba-team@company.com
  - Warning: database-team@company.com
  - Info: database-log@company.com

Subject Template:
[{severity}] {dashboard} - {metric}
```

**Setup Steps**:
1. Configure SMTP settings
2. Create notification channel
3. Add recipient list
4. Test email delivery

---

## 3. Automation Workflows

### 3.1 Automated Lock Investigation

**Trigger**: Lock wait > 300 seconds
**Action Sequence**:

```json
{
  "name": "Auto-investigate-lock",
  "trigger": "lock_wait_seconds > 300",
  "actions": [
    {
      "type": "collect_data",
      "queries": [
        "SELECT blocking query from pg_locks",
        "SELECT query plan from EXPLAIN",
        "SELECT connection info from pg_stat_activity"
      ]
    },
    {
      "type": "notify",
      "channel": "pagerduty",
      "severity": "critical"
    },
    {
      "type": "collect_metrics",
      "metrics": ["cpu", "memory", "io"]
    },
    {
      "type": "create_incident",
      "title": "Lock contention detected",
      "context": "blocking_query_info"
    }
  ]
}
```

---

### 3.2 Automated Bloat Management

**Trigger**: Table bloat > 50%
**Action Sequence**:

```json
{
  "name": "Auto-schedule-vacuum",
  "trigger": "bloat_percent > 50",
  "actions": [
    {
      "type": "identify_tables",
      "filter": "bloat_percent > 50"
    },
    {
      "type": "schedule_job",
      "job_type": "VACUUM FULL",
      "schedule": "next_maintenance_window",
      "priority": "high"
    },
    {
      "type": "notify",
      "channel": "slack",
      "message": "VACUUM FULL scheduled for ${table_name}"
    },
    {
      "type": "create_ticket",
      "system": "jira",
      "project": "DB-OPS",
      "type": "Task",
      "title": "Execute VACUUM FULL on ${table_name}"
    }
  ]
}
```

---

### 3.3 Automated Connection Pool Management

**Trigger**: Idle connections > 25
**Action Sequence**:

```json
{
  "name": "Auto-manage-connections",
  "trigger": "idle_connection_count > 25",
  "actions": [
    {
      "type": "collect_data",
      "queries": [
        "SELECT application, count(*) FROM pg_stat_activity GROUP BY application"
      ]
    },
    {
      "type": "notify",
      "channel": "slack",
      "template": "connection_pool_warning"
    },
    {
      "type": "auto_action",
      "condition": "idle_duration > 3600",
      "action": "terminate_idle_connection"
    },
    {
      "type": "create_ticket",
      "title": "Review ${app_name} connection pool settings"
    }
  ]
}
```

---

### 3.4 Automated Cache Tuning Alert

**Trigger**: Cache hit ratio < 80% for 1 hour
**Action Sequence**:

```json
{
  "name": "Auto-analyze-cache",
  "trigger": "cache_hit_ratio < 80 for 60 minutes",
  "actions": [
    {
      "type": "collect_diagnostics",
      "queries": [
        "SELECT table_name, cache_hit_ratio FROM metrics",
        "SELECT query, cache_hit_ratio FROM slow_queries"
      ]
    },
    {
      "type": "notify",
      "channel": "slack",
      "severity": "warning",
      "recommendations": [
        "Increase shared_buffers",
        "Optimize expensive queries",
        "Review table scan patterns"
      ]
    },
    {
      "type": "generate_report",
      "format": "html",
      "destination": "email"
    }
  ]
}
```

---

## 4. Incident Response Runbooks

### 4.1 Lock Contention Runbook

**Filename**: `/docs/runbooks/lock-contention.md`

**Problem**: High number of active locks with blocking chains

**Detection**:
- Lock Monitoring Dashboard: Active Locks > 10
- PagerDuty Incident triggered
- Slack alert in #critical-alerts

**Investigation Steps**:

1. **View Current Locks**
   ```sql
   SELECT * FROM metrics_pg_lock_waits
   WHERE time > NOW() - INTERVAL '1 hour'
   ORDER BY wait_time_seconds DESC
   LIMIT 10;
   ```

2. **Identify Blocking Query**
   ```sql
   SELECT blocking_query, wait_time_seconds
   FROM metrics_pg_lock_waits
   WHERE blocking_username IS NOT NULL
   ORDER BY wait_time_seconds DESC;
   ```

3. **Check Query Performance**
   ```sql
   EXPLAIN ANALYZE [blocking query];
   ```

4. **Review Connection Info**
   ```sql
   SELECT * FROM pg_stat_activity
   WHERE pid = [blocking_pid];
   ```

**Resolution Options**:

- **Option 1**: Wait for query to complete (if < 5 minutes remaining)
- **Option 2**: KILL blocking query
  ```sql
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE query = '[blocking query]';
  ```
- **Option 3**: Optimize query and rerun
- **Option 4**: Scale up resources (if CPU/memory constrained)

**Follow-up**:
- Document in incident ticket
- Review query execution plan
- Update slow query log
- Schedule performance optimization

---

### 4.2 High Bloat Runbook

**Filename**: `/docs/runbooks/high-bloat.md`

**Problem**: Table bloat > 50%

**Detection**:
- Bloat Analysis Dashboard alert
- Slack notification
- Email to DBA team

**Investigation Steps**:

1. **Identify Bloated Tables**
   ```sql
   SELECT database_name, table_name, dead_ratio_percent, space_wasted_percent
   FROM metrics_pg_bloat_tables
   WHERE dead_ratio_percent > 50
   ORDER BY dead_ratio_percent DESC;
   ```

2. **Estimate Space Recovery**
   ```sql
   SELECT pg_size_pretty(pg_total_relation_size('[table_name]')) as current_size;
   ```

3. **Check Table Activity**
   ```sql
   SELECT seq_scan, seq_tup_read, idx_scan
   FROM pg_stat_user_tables
   WHERE relname = '[table_name]';
   ```

**Resolution Options**:

- **Option 1**: VACUUM (minimal downtime)
  ```sql
  VACUUM ANALYZE [table_name];
  ```

- **Option 2**: VACUUM FULL (requires lock, longer)
  ```sql
  VACUUM FULL ANALYZE [table_name];
  ```

- **Option 3**: REINDEX (rebuild indexes)
  ```sql
  REINDEX TABLE [table_name];
  ```

- **Option 4**: Scheduled maintenance window
  - Plan VACUUM FULL
  - Reduce application load
  - Monitor during execution

**Prevention**:
- Increase AUTOVACUUM frequency
- Review table write patterns
- Monitor bloat trends

---

### 4.3 Low Cache Hit Ratio Runbook

**Filename**: `/docs/runbooks/cache-hit-ratio.md`

**Problem**: Cache hit ratio < 80%

**Detection**:
- Cache Performance Dashboard alert
- Slack notification

**Investigation Steps**:

1. **Check Current Cache Hit Ratio**
   ```sql
   SELECT AVG(cache_hit_ratio) as avg_hit_ratio
   FROM metrics_pg_cache_hit_ratios
   WHERE time > NOW() - INTERVAL '1 hour';
   ```

2. **Identify Low-Hit Tables**
   ```sql
   SELECT table_name, cache_hit_ratio, heap_blks_read, heap_blks_hit
   FROM metrics_pg_cache_hit_ratios
   WHERE cache_hit_ratio < 80
   ORDER BY cache_hit_ratio ASC;
   ```

3. **Check Query Patterns**
   ```sql
   SELECT query, calls, total_time, mean_time
   FROM pg_stat_statements
   WHERE mean_time > 1000
   ORDER BY total_time DESC
   LIMIT 10;
   ```

4. **Review Memory Usage**
   ```sql
   SHOW shared_buffers;
   SHOW effective_cache_size;
   ```

**Resolution Options**:

- **Option 1**: Increase shared_buffers
  ```
  # In postgresql.conf
  shared_buffers = [increased value]

  # Restart PostgreSQL
  systemctl restart postgresql
  ```

- **Option 2**: Optimize expensive queries
  - Add missing indexes
  - Rewrite queries
  - Use EXPLAIN ANALYZE

- **Option 3**: Review application caching
  - Add application-level cache
  - Use connection pooling
  - Batch queries

- **Option 4**: Adjust work_mem
  ```sql
  SET work_mem = '256MB';
  ```

**Monitoring**:
- Track cache hit ratio daily
- Monitor after each change
- Alert if declining trend

---

### 4.4 Idle Transaction Runbook

**Filename**: `/docs/runbooks/idle-transaction.md`

**Problem**: Idle-in-transaction connections > 5

**Detection**:
- Connection Tracking Dashboard alert
- Slack notification

**Investigation Steps**:

1. **Find Idle Transactions**
   ```sql
   SELECT pid, usename, database, state, query_start, query
   FROM pg_stat_activity
   WHERE state = 'idle in transaction'
   ORDER BY query_start ASC;
   ```

2. **Check Duration**
   ```sql
   SELECT pid, EXTRACT(EPOCH FROM (NOW() - query_start)) as duration_seconds
   FROM pg_stat_activity
   WHERE state = 'idle in transaction'
   ORDER BY duration_seconds DESC;
   ```

3. **Identify Application**
   ```sql
   SELECT application_name, count(*) as idle_count
   FROM pg_stat_activity
   WHERE state = 'idle in transaction'
   GROUP BY application_name;
   ```

**Resolution Options**:

- **Option 1**: Connection timeout (if allowed)
  ```sql
  ALTER ROLE [user] SET idle_in_transaction_session_timeout = '5min';
  ```

- **Option 2**: Application notification
  - Alert application team
  - Request code review
  - Add transaction timeout

- **Option 3**: Manual cleanup
  ```sql
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE state = 'idle in transaction'
    AND query_start < NOW() - INTERVAL '30 minutes';
  ```

**Prevention**:
- Code review for transaction handling
- Use connection pooling with timeout
- Monitor application logs
- Add APM tracing

---

## 5. Alert Configuration Files

### 5.1 Grafana Alert Rules (JSON)

**File**: `monitoring/grafana-alerts.json`

```json
{
  "alert_rules": [
    {
      "name": "HighLockCount",
      "dashboard_id": "lock-monitoring",
      "panel_id": 2,
      "condition": "value > 10",
      "for": "5m",
      "annotations": {
        "description": "Active locks exceed threshold",
        "runbook_url": "/docs/runbooks/lock-contention.md"
      },
      "labels": {
        "severity": "critical",
        "team": "database"
      }
    },
    {
      "name": "HighTableBloat",
      "dashboard_id": "bloat-analysis",
      "panel_id": 1,
      "condition": "value > 50",
      "for": "10m",
      "annotations": {
        "description": "Table bloat exceeds 50%",
        "runbook_url": "/docs/runbooks/high-bloat.md"
      },
      "labels": {
        "severity": "warning",
        "team": "database"
      }
    },
    {
      "name": "LowCacheHitRatio",
      "dashboard_id": "cache-performance",
      "panel_id": 1,
      "condition": "value < 80",
      "for": "30m",
      "annotations": {
        "description": "Cache hit ratio below target",
        "runbook_url": "/docs/runbooks/cache-hit-ratio.md"
      },
      "labels": {
        "severity": "warning",
        "team": "database"
      }
    }
  ]
}
```

---

### 5.2 Notification Channel Configuration

**File**: `monitoring/notification-channels.json`

```json
{
  "channels": [
    {
      "name": "Critical Alerts",
      "type": "pagerduty",
      "settings": {
        "integrationKey": "${PAGERDUTY_KEY}"
      },
      "rules": {
        "severity": "critical"
      }
    },
    {
      "name": "Database Alerts Slack",
      "type": "slack",
      "settings": {
        "webhook_url": "${SLACK_WEBHOOK}",
        "channel": "#database-alerts"
      },
      "rules": {
        "severity": ["warning", "critical"]
      }
    },
    {
      "name": "DBA Email",
      "type": "email",
      "settings": {
        "to": "dba-team@company.com",
        "from": "alerts@postgres.internal"
      },
      "rules": {
        "severity": ["warning", "critical"]
      }
    }
  ]
}
```

---

## 6. Implementation Timeline

### Week 1: Alert Rules Setup
- ✅ Define critical thresholds
- ✅ Create Grafana alert rules
- ✅ Configure notification channels
- ✅ Test alert delivery

### Week 2: Notification Integration
- ✅ Setup Slack integration
- ✅ Configure PagerDuty
- ✅ Test email notifications
- ✅ Create escalation policies

### Week 3: Automation Implementation
- ✅ Build automation workflows
- ✅ Implement auto-remediation
- ✅ Test automation logic
- ✅ Create incident tracking

### Week 4: Runbooks & Training
- ✅ Document all runbooks
- ✅ Train team on procedures
- ✅ Conduct tabletop exercises
- ✅ Gather feedback

---

## 7. Success Criteria

### Functionality
✅ All critical metrics have alert rules
✅ Notification delivery < 1 minute
✅ Auto-remediation functional
✅ Runbooks complete and tested
✅ Escalation policies defined

### Reliability
✅ Alert false positive rate < 5%
✅ 99.9% notification delivery success
✅ Team response time tracked
✅ MTTR (Mean Time To Resolve) < 15 minutes

### Operations
✅ Team trained on all procedures
✅ On-call schedule configured
✅ Incident tracking integrated
✅ Post-incident reviews scheduled

---

## 8. Dependencies

### Required from Phase 3-4
✅ Grafana 8.0+ with alerting enabled
✅ All Phase 4 dashboards deployed
✅ Metrics flowing in database
✅ API endpoints operational

### External Services
- [ ] Slack workspace access
- [ ] PagerDuty account
- [ ] Email/SMTP configuration
- [ ] Jira/ticket system integration

### Team Resources
- [ ] Designated on-call rotation
- [ ] DBA team training
- [ ] Management approval for automation

---

## 9. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Alert fatigue | High | Careful threshold tuning, monitoring false positives |
| Auto-remediation mistakes | High | Test in staging, implement approval workflows |
| Notification channel failure | Medium | Redundant channels, fallback procedures |
| Team response delay | Medium | Clear escalation, training, accountability |
| Incomplete runbooks | Medium | Regular reviews, updates, team feedback |

---

## 10. Documentation Requirements

### Files to Create

```
/docs/runbooks/
├── lock-contention.md
├── high-bloat.md
├── cache-hit-ratio.md
├── idle-transaction.md
├── connection-pool.md
└── emergency-procedures.md

/docs/guides/
├── alert-setup.md
├── notification-channels.md
├── automation-workflows.md
├── incident-response.md
└── team-training.md

/monitoring/
├── grafana-alerts.json
├── notification-channels.json
├── automation-rules.json
└── escalation-policies.json
```

---

## Conclusion

Phase 5 will transform pgAnalytics v3 from a visualization platform into a fully automated monitoring and alerting system. With proper alert rules, automation, and runbooks, the team can achieve:

- ✅ **Proactive monitoring**: Issues detected before user impact
- ✅ **Rapid response**: Clear procedures and automation
- ✅ **Reduced MTTR**: Faster incident resolution
- ✅ **Team efficiency**: Automated routine tasks
- ✅ **Operational excellence**: Consistent processes

---

**Next Step**: Begin Phase 5 implementation with alert rules creation and notification channel setup.

**Estimated Duration**: 3-4 weeks
**Team Effort**: 1-2 DBAs, 1 SRE/DevOps engineer
**Success Metrics**: 100% critical coverage, < 5% false positive rate, < 15 min MTTR

---

Generated: March 3, 2026
Status: Planning Complete - Ready for Implementation
