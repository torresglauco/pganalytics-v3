# Incident Response Runbook: Lock Contention

**Severity**: CRITICAL
**On-Call Team**: Database Engineering
**Response Time SLA**: < 5 minutes
**Runbook Version**: 1.0
**Last Updated**: March 3, 2026

---

## Quick Summary

Lock contention occurs when transactions block each other, preventing progress. This runbook guides you through identification, diagnosis, and resolution.

**Key Indicators**:
- Active locks > 10 for 5+ minutes
- Lock wait time > 300 seconds
- Blocked transactions in `pg_stat_activity`
- Application timeout errors

---

## Alert Definitions

### Critical Alert: Lock Contention - Critical

```
Trigger Condition: Active locks > 10 for 5 minutes
Severity: CRITICAL
Channels: PagerDuty, Slack #critical-alerts, Email DBA
Action: Page on-call DBA immediately
Auto-Remediation: Kill blocking locks (if enabled)
```

### Warning Alert: Blocking Transaction

```
Trigger Condition: Lock wait time > 300 seconds for 5 minutes
Severity: CRITICAL
Channels: PagerDuty, Slack #critical-alerts
Action: Page on-call DBA
Auto-Remediation: Kill blocking locks (if enabled)
```

### Warning Alert: Max Lock Age

```
Trigger Condition: Lock age > 300 seconds for 5 minutes
Severity: WARNING
Channels: Slack #database-alerts, Email
Action: Alert DBA team for investigation
Auto-Remediation: None (manual only)
```

---

## Step 1: Immediate Response (0-2 minutes)

### 1.1 Confirm Alert

```bash
# Check if alert is genuine
curl -s http://localhost:5002/automation/health | jq .

# Confirm database connectivity
pg_isready -h production.db.internal -p 5432
```

### 1.2 Assess Severity

**Critical Indicators** (Needs immediate action):
- Active locks = 20+
- Wait time = 600+ seconds
- Multiple blocked transactions
- Application errors reported

**Moderate Indicators** (Watch and investigate):
- Active locks = 10-20
- Wait time = 300-600 seconds
- Single blocked transaction
- No application errors yet

### 1.3 Notify Team

```bash
# If you didn't receive automated page, notify:
# - Slack: @database-oncall
# - Phone: [on-call escalation number]
# - Chat: [team communication channel]

echo "Lock contention alert - starting investigation"
```

---

## Step 2: Diagnosis (2-5 minutes)

### 2.1 Query Current Lock Status

```sql
-- Connect to affected database
psql -h production.db.internal -U pganalytics -d pganalytics

-- View active locks
SELECT
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.application_name AS blocked_app,
    blocking_activity.application_name AS blocking_app,
    EXTRACT(EPOCH FROM (NOW() - blocked_activity.query_start)) AS blocked_seconds,
    blocked_activity.query AS blocked_query,
    blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted
ORDER BY blocked_activity.query_start;
```

### 2.2 Identify Blocking Transaction

```sql
-- Get detailed info on blocking transaction
SELECT
    pid,
    usename,
    application_name,
    state,
    query_start,
    EXTRACT(EPOCH FROM (NOW() - query_start)) AS seconds_running,
    query
FROM pg_stat_activity
WHERE pid = [blocking_pid from above]
\x on
```

### 2.3 Check Blocking Query

```sql
-- If query is long, see first 100 chars
SELECT
    pid,
    query_start,
    SUBSTRING(query, 1, 100) AS query_preview,
    EXTRACT(EPOCH FROM (NOW() - query_start)) AS seconds_running
FROM pg_stat_activity
WHERE pid = [blocking_pid];
```

### 2.4 Count Blocked Transactions

```sql
-- See how many transactions are blocked
SELECT
    COUNT(DISTINCT blocked_activity.pid) AS blocked_count,
    COUNT(DISTINCT blocking_activity.pid) AS blocking_count
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

---

## Step 3: Decision Tree

```
Are there active locks?
├─ NO: False alarm, end investigation
└─ YES: Continue

Is wait time > 300 seconds?
├─ NO: Monitor, don't act yet
└─ YES: Continue

Is blocking process:
  A) Application process (vs. system/vacuum)?
  ├─ YES: Can safely kill
  └─ NO: Investigate further

B) Long-running transaction?
  ├─ YES: Check with app owner first
  └─ NO: Safe to terminate

C) Idle in transaction?
  ├─ YES: Safe to terminate
  └─ NO: Verify before killing

Decision:
├─ Safe to kill: Terminate immediately
├─ Unsafe to kill: Manual investigation needed
└─ Borderline: Notify application owner first
```

---

## Step 4: Resolution Options

### Option 1: Auto-Remediation (Recommended)

**If AUTO_REMEDIATE is enabled**, the system already killed blocking locks:

```bash
# Check remediation history
curl -s http://localhost:5002/automation/history?alert_name=lock_contention_critical | jq '.results[0]'

# Should show:
# - status: success
# - locks_killed: [number]
# - pids_terminated: [list of PIDs]
```

**Verify resolution**:
```bash
# Alert should clear within 1-2 minutes
# Check incident in correlation engine:
curl -s http://localhost:5003/correlation/incidents | jq '.incidents[0]'
```

### Option 2: Manual Termination (If AUTO_REMEDIATE disabled)

```bash
# Log the decision in team chat
echo "Manually terminating blocking PID [pid] - [reason]"

# Connect to database
psql -h production.db.internal -U pganalytics -d pganalytics

# Verify one more time (critical!)
SELECT pid, application_name, query FROM pg_stat_activity WHERE pid = [pid];

# Terminate blocking process
SELECT pg_terminate_backend([pid]);
-- Should return: true

# Verify it's gone
SELECT * FROM pg_stat_activity WHERE pid = [pid];
-- Should return: (no rows)

# Check if lock is resolved
SELECT COUNT(*) FROM pg_locks WHERE NOT granted;
-- Should be 0 or much lower
```

### Option 3: Connection Restart (Nuclear Option)

**Only if locks don't clear after termination**:

```bash
# Close the application connection pool
# (contact application owner to restart app connection pool)

# Verify all blocking connections are gone
SELECT COUNT(*) FROM pg_stat_activity WHERE state != 'idle';
```

---

## Step 5: Verification

### 5.1 Confirm Locks Cleared

```sql
-- No more blocked locks
SELECT COUNT(*) FROM pg_locks WHERE NOT granted;
-- Result: 0

-- No more waiting transactions
SELECT COUNT(*) FROM pg_stat_activity
WHERE wait_event IS NOT NULL
AND wait_event != 'Client';
-- Result: 0

-- Wait times are normal
SELECT
    MAX(EXTRACT(EPOCH FROM (NOW() - query_start))) AS max_query_age_seconds
FROM pg_stat_activity
WHERE state != 'idle';
-- Result: < 60 seconds
```

### 5.2 Monitor Application

```bash
# Check if application is recovering
# - Monitor error rates (should decrease)
# - Monitor response times (should return to normal)
# - Check application logs for correlation

# Wait 2-5 minutes for full recovery
sleep 300
```

### 5.3 Verify Alert Clears

```bash
# Alert should auto-resolve within 5 minutes
# Check Grafana: Should show alert clearing

# Check incident status
curl -s http://localhost:5003/correlation/incidents?state=active | jq

# Once resolved, incident should move to RESOLVED state
curl -s http://localhost:5003/correlation/incident/[incident_id] | jq '.state'
```

---

## Step 6: Root Cause Analysis

### 6.1 Analyze Blocking Query

```sql
-- Get execution plan of blocking query
EXPLAIN ANALYZE [blocking_query];

-- Or if complex:
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) [blocking_query];
```

### 6.2 Check Query Patterns

```sql
-- Find similar expensive queries
SELECT
    query,
    calls,
    mean_exec_time,
    max_exec_time,
    stddev_exec_time
FROM pg_stat_statements
WHERE query LIKE '%[table_name]%'
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### 6.3 Check Transaction History

```sql
-- See if locks are recurring
SELECT
    DATE_TRUNC('hour', query_start)::date AS date_hour,
    COUNT(*) AS lock_count
FROM pg_stat_activity
WHERE state = 'active'
GROUP BY DATE_TRUNC('hour', query_start)
ORDER BY date_hour DESC
LIMIT 24;
```

---

## Step 7: Long-Term Prevention

### 7.1 Identify Root Cause

**Common Causes**:

1. **Long-running batch jobs**
   - Solution: Add connection timeout
   - Implement: Job scheduling with timeouts

2. **Application connection leak**
   - Solution: Restart application
   - Implement: Connection pool monitoring

3. **Missing transaction commit/rollback**
   - Solution: Add error handling
   - Implement: Application code review

4. **Poorly tuned queries**
   - Solution: Add indexes, rewrite query
   - Implement: Query optimization

5. **Deadlock loops**
   - Solution: Change query order
   - Implement: Application logic review

### 7.2 Implement Fix

**Example: Add Query Timeout**:
```sql
ALTER SYSTEM SET statement_timeout = '30min';
SELECT pg_reload_conf();
```

**Example: Add Connection Timeout**:
```bash
# In connection string:
psql "postgresql://user@host/db?connect_timeout=10&statement_timeout=1800000"
```

### 7.3 Monitor Prevention

```sql
-- Monitor for recurring issues
SELECT
    application_name,
    COUNT(*) AS connection_count,
    AVG(EXTRACT(EPOCH FROM (NOW() - query_start))) AS avg_query_age
FROM pg_stat_activity
WHERE state != 'idle'
GROUP BY application_name
ORDER BY connection_count DESC;
```

---

## Escalation Path

### Level 1: On-Call DBA (You are here)
- **Response**: < 2 minutes
- **Actions**: Diagnose, remediate locks
- **Escalate if**: Cannot resolve in 5 minutes

### Level 2: DBA Team Lead
- **Trigger**: Locks unresolved after 5 minutes
- **Actions**: Review diagnosis, approve killing
- **Contact**: [Team lead phone/slack]

### Level 3: Database Architect
- **Trigger**: Locks recurring (> 3x per week)
- **Actions**: Root cause analysis, long-term fix
- **Contact**: [Architect phone/slack]

### Level 4: Application Owner
- **Trigger**: Application causing locks
- **Actions**: Code review, connection pool review
- **Contact**: [App owner contact]

---

## Communication Template

### Incident Notification

```
🚨 INCIDENT: Lock Contention Alert

Database: production
Severity: CRITICAL
Time: [current time]
Detection: Grafana Alert

Status: INVESTIGATING
- Identifying blocking process
- Checking blocking query
- Assessing safe termination

Updates every 2 minutes.
```

### Resolution Notification

```
✅ INCIDENT RESOLVED: Lock Contention

Database: production
Duration: [time from start to resolution]
Root Cause: [brief description]

Actions Taken:
- [action 1]
- [action 2]
- [action 3]

Prevention:
- [long-term fix]

Next Steps: Monitor for recurrence.
```

---

## Useful Commands

### Quick Diagnosis

```bash
# Everything in one go
psql -h $DB_HOST -U pganalytics -d pganalytics -c "
SELECT
    blocked.pid AS blocked,
    blocking.pid AS blocking,
    EXTRACT(EPOCH FROM (NOW() - blocked.query_start)) AS blocked_seconds,
    blocked.usename,
    SUBSTRING(blocked.query, 1, 50) AS blocked_query
FROM pg_stat_activity blocked
JOIN pg_stat_activity blocking ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
WHERE blocked.pid != blocking.pid
ORDER BY blocked.query_start;
"
```

### Kill All Blocking Locks

```bash
# Dangerous! Only use after thorough investigation
psql -h $DB_HOST -U pganalytics -d pganalytics -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE pid = ANY(
    SELECT DISTINCT blocking_locks.pid
    FROM pg_locks blocked_locks
    JOIN pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    WHERE NOT blocked_locks.granted
);
"
```

### Monitor Locks in Real-Time

```bash
# Watch command (macOS/Linux)
watch -n 2 'psql -h $DB_HOST -U pganalytics -d pganalytics -c "
SELECT COUNT(*) as active_locks FROM pg_locks WHERE NOT granted;
"'
```

---

## Troubleshooting

### Problem: Lock Status Query Hangs

**Symptom**: `pg_stat_activity` query itself is blocked

**Solution**:
1. Use different connection with `superuser` role
2. Kill the query blocking the query tool:
   ```sql
   SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query LIKE '%pg_stat_activity%';
   ```

### Problem: pg_terminate_backend Returns False

**Symptom**: Process won't die

**Solution**:
1. Process already exited - check status
2. Insufficient privileges - use superuser
3. PostgreSQL version issue - verify version

### Problem: Locks Return After Resolution

**Symptom**: Same blocking PID reappears

**Solution**:
1. Root cause not fixed (see root cause analysis)
2. Application is restarting transactions
3. Schedule meeting with application team

---

## Success Checklist

- [ ] Alert acknowledged within 2 minutes
- [ ] Diagnosis completed within 5 minutes
- [ ] Blocking locks killed within 10 minutes
- [ ] Alert resolved/cleared within 15 minutes
- [ ] Application confirmed recovered
- [ ] Incident marked as resolved
- [ ] Root cause identified
- [ ] Prevention plan documented
- [ ] Team notified of resolution
- [ ] Post-incident review scheduled (if recurring)

---

## References

- **Grafana Dashboard**: Lock Monitoring
- **Runbook**: Database Fundamentals (Lock Types)
- **FAQ**: Lock Contention Troubleshooting
- **Team Wiki**: On-Call Procedures

---

Generated: March 3, 2026
Author: Claude Opus 4.6
Last Reviewed: March 3, 2026
