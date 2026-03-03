# Incident Response Runbook: High Table Bloat

**Severity**: WARNING
**On-Call Team**: Database Engineering
**Response Time SLA**: < 30 minutes
**Runbook Version**: 1.0
**Last Updated**: March 3, 2026

---

## Quick Summary

Table bloat occurs when dead tuples accumulate, wasting disk space and degrading query performance. This runbook guides you through identification, analysis, and remediation.

**Key Indicators**:
- Table dead ratio > 50%
- Disk usage growing without data growth
- Query performance degrading
- Autovacuum lagging behind writes

---

## Alert Definition

### Warning Alert: High Table Bloat

```
Trigger Condition: Max table dead_ratio > 50% for 10 minutes
Severity: WARNING
Channels: Slack #database-alerts, Email DBA, JIRA auto-ticket
Action: Create ticket, alert DBA team
Auto-Remediation: VACUUM ANALYZE top 5 tables (if enabled)
```

---

## Step 1: Immediate Response (0-5 minutes)

### 1.1 Confirm Alert

```bash
# Check alert is genuine
curl -s http://localhost:5001/webhook/jira | jq '.status'

# Verify database connectivity
pg_isready -h production.db.internal -p 5432
```

### 1.2 Assess Severity

**High Priority** (Act immediately):
- Dead ratio > 70%
- Multiple tables affected (5+)
- Disk usage > 85%
- Query performance degrading significantly

**Medium Priority** (Investigate within 2 hours):
- Dead ratio 50-70%
- 2-4 tables affected
- Disk usage 70-85%
- Minor performance impact

**Low Priority** (Monitor):
- Dead ratio 50-60%
- Single table affected
- Disk usage < 70%
- No performance impact yet

### 1.3 Notify Team (if not automated)

```bash
echo "Table bloat alert - investigating bloated tables"
# Create JIRA ticket if auto-ticket failed
```

---

## Step 2: Diagnosis (5-15 minutes)

### 2.1 Find Bloated Tables

```sql
-- Connect to database
psql -h production.db.internal -U pganalytics -d pganalytics

-- Find tables with high bloat ratio
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename, 'main')) AS table_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) AS dead_space,
    ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) AS dead_ratio_percent,
    n_live_tup,
    n_dead_tup
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) DESC
LIMIT 20;
```

### 2.2 Check Autovacuum Status

```sql
-- See when tables were last vacuumed
SELECT
    schemaname,
    tablename,
    last_vacuum,
    last_autovacuum,
    EXTRACT(EPOCH FROM (NOW() - last_autovacuum)) / 3600 AS hours_since_autovacuum,
    n_dead_tup,
    autovacuum_count
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY last_autovacuum NULLS FIRST
LIMIT 20;
```

### 2.3 Check Autovacuum Tuning

```sql
-- Check if autovacuum is running
SELECT datname, pid, usename, state, query
FROM pg_stat_activity
WHERE query LIKE '%autovacuum%';

-- Check autovacuum settings
SHOW autovacuum;
SHOW autovacuum_analyze_scale_factor;
SHOW autovacuum_vacuum_scale_factor;
SHOW autovacuum_vacuum_cost_delay;
SHOW autovacuum_vacuum_cost_limit;
```

### 2.4 Estimate Recovery

```sql
-- Estimate how much space can be recovered
SELECT
    schemaname,
    tablename,
    pg_size_pretty(
        GREATEST(pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main'), 0)
    ) AS recoverable_space,
    ROUND(
        100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1),
        2
    ) AS recovery_percent
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
    AND schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) DESC
LIMIT 10;
```

---

## Step 3: Resolution

### Option 1: Auto-Remediation (Recommended)

**If AUTO_REMEDIATE is enabled**, the system already vacuumed top tables:

```bash
# Check remediation history
curl -s http://localhost:5002/automation/history?alert_name=high_table_bloat_warning | jq '.results[0]'

# Should show:
# - status: success
# - tables_vacuumed: [list of tables]
```

**Verify remediation**:
```sql
-- Check if bloat decreased
SELECT
    tablename,
    n_dead_tup,
    last_autovacuum
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY n_dead_tup DESC
LIMIT 5;
```

### Option 2: Manual VACUUM (If AUTO_REMEDIATE disabled)

```sql
-- Vacuum highest-bloat table
VACUUM ANALYZE schema.table_name;

-- For very large tables (non-blocking):
VACUUM (ANALYZE, PARALLEL 4) schema.table_name;

-- Monitor progress in separate connection:
SELECT
    pid,
    relname,
    phase,
    heap_blks_total,
    heap_blks_scanned,
    ROUND(100.0 * heap_blks_scanned / GREATEST(heap_blks_total, 1), 2) AS progress_percent
FROM pg_stat_progress_vacuum;
```

### Option 3: Full Maintenance (Multiple Tables)

```bash
# Script to vacuum top 5 bloated tables
psql -h production.db.internal -U pganalytics -d pganalytics <<EOF
-- Get top 5 tables
\set table1 'public.table_with_most_bloat'
\set table2 'public.second_bloated_table'
\set table3 'public.third_bloated_table'
\set table4 'public.fourth_bloated_table'
\set table5 'public.fifth_bloated_table'

-- Vacuum each
VACUUM ANALYZE :table1;
VACUUM ANALYZE :table2;
VACUUM ANALYZE :table3;
VACUUM ANALYZE :table4;
VACUUM ANALYZE :table5;
EOF
```

---

## Step 4: Verification

### 4.1 Confirm Space Recovered

```sql
-- Check bloat decreased
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS current_size,
    ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) AS dead_ratio_percent
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) DESC
LIMIT 5;
```

### 4.2 Verify Query Performance

```bash
# Run sample queries to verify performance
# Compare execution times before/after VACUUM

# Example:
psql -h production.db.internal -U pganalytics -d pganalytics -c "
EXPLAIN ANALYZE SELECT COUNT(*) FROM [bloated_table];
"
```

### 4.3 Monitor Alert

```bash
# Alert should clear within 5-10 minutes
# Check Grafana: Alert should show resolution

# Check incident status
curl -s http://localhost:5003/correlation/incidents?state=active | jq
```

---

## Step 5: Root Cause Analysis

### 5.1 Why Did Bloat Accumulate?

```sql
-- Check write patterns
SELECT
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    EXTRACT(EPOCH FROM (NOW() - last_autovacuum)) / 3600 AS hours_since_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 10000
ORDER BY n_tup_upd + n_tup_del DESC
LIMIT 10;
```

### 5.2 Is Autovacuum Sufficient?

```sql
-- Check autovacuum effectiveness
SELECT
    schemaname,
    tablename,
    autovacuum_count,
    autoanalyze_count,
    EXTRACT(EPOCH FROM (NOW() - last_autovacuum)) / 3600 AS hours_since_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 5000
ORDER BY autovacuum_count DESC;
```

### 5.3 Identify Update-Heavy Tables

```sql
-- Find tables with high update/delete ratio
SELECT
    schemaname,
    tablename,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    ROUND(100.0 * (n_tup_upd + n_tup_del) / GREATEST(n_tup_ins + n_tup_upd + n_tup_del, 1), 2) AS update_del_ratio
FROM pg_stat_user_tables
WHERE n_tup_ins + n_tup_upd + n_tup_del > 100000
ORDER BY (n_tup_upd + n_tup_del) DESC
LIMIT 10;
```

---

## Step 6: Long-Term Prevention

### 6.1 Tune Autovacuum

```sql
-- For update-heavy tables, increase autovacuum frequency:
ALTER TABLE schema.bloated_table SET (
    autovacuum_vacuum_scale_factor = 0.01,  -- 1% instead of default 5%
    autovacuum_analyze_scale_factor = 0.005, -- 0.5% instead of default 1%
    autovacuum_vacuum_cost_delay = 5  -- milliseconds
);

-- Or globally in postgresql.conf:
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.02;
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.01;
SELECT pg_reload_conf();
```

### 6.2 Schedule Manual Maintenance

```bash
# Create maintenance window for large VACUUM operations
# Example: Schedule Sunday 2 AM for full VACUUM FULL

# In cron:
0 2 * * 0 /usr/local/bin/maintenance-vacuum.sh

# Script content:
#!/bin/bash
psql -h production.db.internal -U pganalytics -d pganalytics <<EOF
-- Full VACUUM (requires exclusive lock)
VACUUM FULL ANALYZE public.frequently_updated_table;
EOF
```

### 6.3 Monitor Prevention

```sql
-- Create ongoing monitoring query
SELECT
    schemaname,
    tablename,
    n_dead_tup,
    ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) AS bloat_ratio,
    CASE
        WHEN ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) > 70 THEN 'HIGH'
        WHEN ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) > 50 THEN 'MEDIUM'
        ELSE 'LOW'
    END AS bloat_level
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY bloat_ratio DESC;
```

---

## Escalation Path

### Level 1: On-Call DBA (You are here)
- **Response**: < 10 minutes
- **Actions**: Diagnose bloat, trigger VACUUM
- **Escalate if**: Cannot resolve in 30 minutes

### Level 2: DBA Team
- **Trigger**: Bloat recurring (> 2x per week)
- **Actions**: Tune autovacuum, analyze patterns
- **Contact**: [Team lead]

### Level 3: Database Architect
- **Trigger**: Fundamental design issue
- **Actions**: Partition design, retention policy
- **Contact**: [Architect]

---

## Communication Template

### Initial Notification

```
⚠️ ALERT: High Table Bloat Detected

Database: production
Severity: WARNING
Time: [current time]

Affected Tables:
- [table1]: 65% bloat
- [table2]: 58% bloat
- [table3]: 52% bloat

Status: VACUUMING
Recovery in progress...
```

### Resolution Notification

```
✅ RESOLVED: High Table Bloat

Database: production
Duration: [time]

Space Recovered:
- [table1]: 2.3 GB
- [table2]: 1.8 GB
- [table3]: 1.2 GB
Total: 5.3 GB

Next: Monitor autovacuum effectiveness
```

---

## Useful Commands

### Quick Bloat Check

```bash
psql -h $DB_HOST -U pganalytics -d pganalytics -c "
SELECT
    tablename,
    ROUND(100.0 * (pg_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename, 'main')) / GREATEST(pg_relation_size(schemaname||'.'||tablename), 1), 2) AS bloat_ratio,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY bloat_ratio DESC
LIMIT 10;
"
```

### Monitor VACUUM Progress

```bash
watch -n 5 'psql -h $DB_HOST -U pganalytics -d pganalytics -c "
SELECT
    pid,
    relname,
    phase,
    ROUND(100.0 * heap_blks_scanned / GREATEST(heap_blks_total, 1), 2) AS progress_percent,
    heap_blks_scanned,
    heap_blks_total
FROM pg_stat_progress_vacuum;
"'
```

---

## Troubleshooting

### Problem: VACUUM Takes Too Long

**Solution**:
1. Use non-blocking VACUUM: `VACUUM (PARALLEL 4) table_name`
2. Schedule during low traffic window
3. Consider VACUUM FULL during maintenance

### Problem: Bloat Returns Immediately

**Solution**:
1. Autovacuum not tuned enough
2. Application pattern changed
3. Consider partitioning

---

## Success Checklist

- [ ] Alert acknowledged
- [ ] Bloated tables identified
- [ ] VACUUM triggered/completed
- [ ] Space recovery verified
- [ ] Performance impact assessed
- [ ] Autovacuum tuning reviewed
- [ ] Prevention plan documented
- [ ] JIRA ticket updated/closed

---

Generated: March 3, 2026
