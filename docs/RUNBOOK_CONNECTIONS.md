# Incident Response Runbook: High Connection Count

**Severity**: WARNING
**On-Call Team**: Database Engineering
**Response Time SLA**: < 15 minutes
**Runbook Version**: 1.0
**Last Updated**: March 3, 2026

---

## Quick Summary

High connection count indicates connection pool saturation, which can cause application timeouts and cascading failures. This runbook covers diagnosis and remediation.

**Key Indicators**:
- Active connections > 150
- Idle connections > 50
- Connection creation rate high
- Application timeout errors

---

## Alert Definitions

### Warning Alert: High Connection Count
```
Trigger: Connections > 150 for 10 minutes
Auto-Remediation: Close idle connections (if enabled)
Response: < 15 minutes
```

### Warning Alert: Idle-in-Transaction
```
Trigger: Idle-in-transaction connections > 5 for 2 minutes
Auto-Remediation: Terminate idle-txn (if enabled)
Response: < 10 minutes
```

---

## Quick Diagnosis

```sql
-- Connect to database
psql -h production.db.internal -U pganalytics -d pganalytics

-- Connection summary
SELECT
    state,
    COUNT(*) as count,
    COUNT(DISTINCT usename) as users,
    COUNT(DISTINCT application_name) as applications
FROM pg_stat_activity
WHERE pid != pg_backend_pid()
GROUP BY state
ORDER BY count DESC;

-- Find idle connections
SELECT
    pid,
    usename,
    application_name,
    state,
    EXTRACT(EPOCH FROM (NOW() - query_start)) AS idle_seconds,
    query_start
FROM pg_stat_activity
WHERE state = 'idle'
    AND query_start < NOW() - INTERVAL '5 minutes'
ORDER BY query_start
LIMIT 20;

-- Find idle-in-transaction (DANGEROUS!)
SELECT
    pid,
    usename,
    application_name,
    EXTRACT(EPOCH FROM (NOW() - xact_start)) AS txn_seconds,
    xact_start
FROM pg_stat_activity
WHERE state = 'idle in transaction'
    AND xact_start < NOW() - INTERVAL '10 minutes'
ORDER BY xact_start;
```

---

## Resolution Options

### Option 1: Auto-Remediation (Recommended)

```bash
# Check if auto-remediation already closed connections
curl -s http://localhost:5002/automation/history?limit=10 | jq '.results[] | select(.action == "close_idle_connections")'
```

### Option 2: Manual Termination

```sql
-- Terminate idle connections (5+ minutes)
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
    AND query_start < NOW() - INTERVAL '5 minutes'
    AND pid != pg_backend_pid();

-- Terminate idle-in-transaction (10+ minutes)
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle in transaction'
    AND xact_start < NOW() - INTERVAL '10 minutes'
    AND pid != pg_backend_pid();
```

---

## Root Cause Analysis

```sql
-- Applications with most connections
SELECT
    application_name,
    COUNT(*) as connection_count,
    COUNT(DISTINCT usename) as users,
    COUNT(CASE WHEN state = 'active' THEN 1 END) as active,
    COUNT(CASE WHEN state = 'idle' THEN 1 END) as idle,
    COUNT(CASE WHEN state = 'idle in transaction' THEN 1 END) as idle_txn
FROM pg_stat_activity
WHERE pid != pg_backend_pid()
GROUP BY application_name
ORDER BY connection_count DESC
LIMIT 15;

-- Identify connection leaks
SELECT
    application_name,
    state,
    COUNT(*) as count,
    MAX(EXTRACT(EPOCH FROM (NOW() - query_start))) AS max_age_seconds
FROM pg_stat_activity
WHERE pid != pg_backend_pid()
GROUP BY application_name, state
ORDER BY count DESC;
```

---

## Long-Term Prevention

### 1. Reduce Connection Pool Size
```bash
# Application-side connection pool configuration
# Typical: 5-20 per application
# Check current: curl http://app-host/admin/pool-stats
```

### 2. Implement Connection Timeout
```bash
# Add to connection string:
options='-c statement_timeout=30min -c idle_in_transaction_session_timeout=10min'
```

### 3. Monitor Connection Patterns
```sql
-- Daily connection trend
SELECT
    DATE(query_start) as date,
    COUNT(*) as connection_count,
    MAX(EXTRACT(EPOCH FROM (NOW() - query_start))) as max_age_seconds
FROM pg_stat_activity
WHERE pid != pg_backend_pid()
GROUP BY DATE(query_start)
ORDER BY date DESC
LIMIT 30;
```

---

## Escalation

**Level 1**: Terminate idle connections (you)
**Level 2**: Contact application owner if pattern continues
**Level 3**: Architecture review if recurring

---

## Success Checklist

- [ ] Alert acknowledged
- [ ] Connection state analyzed
- [ ] Idle connections closed
- [ ] Connection count reduced below 150
- [ ] Application functionality verified
- [ ] Root cause identified
- [ ] Prevention plan documented

---

Generated: March 3, 2026
