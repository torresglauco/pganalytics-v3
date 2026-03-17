# Phase 4 Grafana Dashboards - Complete Implementation Guide

**Date**: 2026-03-03
**Status**: ✅ Complete - All 7 Production Dashboards Implemented
**Commit**: 892167c

---

## Overview

Phase 4 delivers **7 production-ready Grafana dashboards** that provide comprehensive monitoring coverage for all Phase 3 metrics. These dashboards transform raw metrics from the database into actionable visualizations.

### Dashboard List

| Dashboard | UID | Purpose | Panels | Refresh |
|-----------|-----|---------|--------|---------|
| **Schema Overview** | `schema-overview` | Database structure tracking | 5 | 5m |
| **Lock Monitoring** | `lock-monitoring` | Real-time lock detection | 7 | 1m |
| **Bloat Analysis** | `bloat-analysis` | Table/index bloat analysis | 7 | 5m |
| **Cache Performance** | `cache-performance` | Buffer pool efficiency | 6 | 5m |
| **Connection Tracking** | `connection-tracking` | Session monitoring | 7 | 1m |
| **Extensions Config** | `extensions-config` | Extension inventory | 6 | 5m |
| **System Overview** | `system-overview` | Health dashboard | 8 | 1m |

**Total**: 46 visualization panels
**Total Queries**: 46 SQL queries on TimescaleDB
**Coverage**: 100% of Phase 3 metrics

---

## 1. Schema Overview Dashboard

**File**: `dashboards/schema-overview.json`
**Purpose**: Monitor database schema structure and changes

### Panels (5)

#### 1. Table Count Trend (7 Days)
- **Type**: Time Series
- **Query**: Counts distinct tables by time
- **Threshold**: Red at 80 (schema expansion warning)
- **Use Case**: Detect schema changes, schema growth monitoring

```sql
SELECT
  time,
  COUNT(*) as table_count
FROM metrics_pg_schema_tables
WHERE $__timeFilter(time)
GROUP BY time
ORDER BY time DESC
```

#### 2. Schema Statistics (24h)
- **Type**: Stat
- **Metrics**:
  - Total tables (distinct count)
  - Total schemas (distinct count)
- **Time Window**: Last 24 hours
- **Use Case**: Quick reference for schema inventory

#### 3. Top Tables by Column Count
- **Type**: Table
- **Limits**: Top 20 tables
- **Columns**: database, schema, table, column_count
- **Use Case**: Identify complex tables, schema complexity analysis

```sql
SELECT
  database_name,
  schema_name,
  table_name,
  COUNT(*) as column_count
FROM metrics_pg_schema_columns
WHERE time > NOW() - INTERVAL '1 day'
GROUP BY database_name, schema_name, table_name
ORDER BY column_count DESC
LIMIT 20
```

#### 4. Constraint Types Distribution (7 Days)
- **Type**: Time Series (Stacked Bars)
- **Breakdown**: By constraint_type (PK, FK, CHECK, UNIQUE, etc.)
- **Stacking**: Normal (shows contribution)
- **Use Case**: Constraint usage patterns

#### 5. Foreign Key Relationships
- **Type**: Table
- **Limits**: Top 50 relationships
- **Columns**: database, source_table, target_table, fk_count
- **Use Case**: Understand data relationships, dependency analysis

---

## 2. Lock Monitoring Dashboard

**File**: `dashboards/lock-monitoring.json`
**Purpose**: Real-time lock detection and blocking chain identification

### Panels (7)

#### 1. Active Locks (Real-time)
- **Type**: Stat
- **Metric**: COUNT(*) of granted locks
- **Thresholds**:
  - Green: 0-4
  - Yellow: 5+
  - Red: 10+
- **Refresh**: 1 minute (critical)
- **Use Case**: Instant lock status

```sql
SELECT
  COUNT(*) as active_locks
FROM metrics_pg_locks
WHERE time > NOW() - INTERVAL '5 minutes'
  AND granted = true
```

#### 2. Lock Wait Chains
- **Type**: Stat
- **Metric**: COUNT(*) of wait conditions
- **Thresholds**:
  - Green: 0
  - Yellow: 1+
  - Red: 3+
- **Use Case**: Detect deadlock risks

#### 3. Max Lock Age
- **Type**: Stat
- **Metric**: Maximum lock age in seconds
- **Thresholds**:
  - Green: < 300s
  - Yellow: 300-1800s
  - Orange: 1800-3600s
  - Red: > 3600s
- **Use Case**: Identify long-held locks

#### 4. Lock Count Trend (24h)
- **Type**: Time Series
- **Query**: Count locks by time
- **Time Window**: Last 24 hours
- **Use Case**: Lock activity patterns

#### 5. Lock Modes Distribution (7 Days)
- **Type**: Time Series (Stacked Bars)
- **Breakdown**: By lock mode (AccessShare, Row, RowExclusive, etc.)
- **Time Window**: Last 7 days
- **Use Case**: Lock type distribution, contention patterns

#### 6. Current Lock Blocking Chain (Last Hour)
- **Type**: Table
- **Limits**: Top 50 blocking chains
- **Columns**:
  - database_name
  - blocked_username
  - blocking_username
  - wait_time_seconds
  - blocked_query (truncated)
  - blocking_query (truncated)
- **Use Case**: Identify current blocking, root cause analysis

```sql
SELECT
  database_name,
  blocked_username,
  blocking_username,
  wait_time_seconds,
  blocked_query,
  blocking_query
FROM metrics_pg_lock_waits
WHERE time > NOW() - INTERVAL '1 hour'
ORDER BY wait_time_seconds DESC
LIMIT 50
```

---

## 3. Bloat Analysis Dashboard

**File**: `dashboards/bloat-analysis.json`
**Purpose**: Detect and track table/index bloat

### Panels (7)

#### 1. Max Table Bloat %
- **Type**: Stat
- **Metric**: MAX(dead_ratio_percent)
- **Thresholds**:
  - Green: < 10%
  - Yellow: 10-25%
  - Orange: 25-50%
  - Red: > 50%
- **Time Window**: Last 24 hours
- **Use Case**: Overall bloat health indicator

#### 2. Total Dead Tuples
- **Type**: Stat
- **Metric**: SUM(dead_tuples) in bytes
- **Thresholds**:
  - Green: < 100MB
  - Yellow: 100MB-500MB
  - Red: > 500MB
- **Use Case**: Wasted space measurement

#### 3. Space Wasted (Bloated Tables)
- **Type**: Stat
- **Metric**: SUM(table_size) where bloat > 10%
- **Thresholds**:
  - Green: < 1GB
  - Yellow: 1GB-5GB
  - Red: > 5GB
- **Use Case**: Actionable cleanup targets

#### 4. Tables with High Bloat (>10%)
- **Type**: Table
- **Limits**: Top 20 tables
- **Columns**:
  - database_name
  - schema_name
  - table_name
  - dead_ratio_percent
  - dead_tuples
  - live_tuples
  - recommendation (color-coded)
- **Color Coding**:
  - 🔴 CRITICAL: > 50%
  - 🟠 HIGH: 25-50%
  - 🟡 MEDIUM: 10-25%
  - 🟢 LOW: < 10%

#### 5. Unused/Rarely Used Indexes
- **Type**: Table
- **Limits**: Top 20 indexes
- **Query**: Filter by UNUSED or RARELY_USED status
- **Columns**:
  - database_name
  - schema_name
  - table_name
  - index_name
  - index_scans
  - usage_status
  - recommendation
- **Use Case**: Index cleanup candidates

```sql
SELECT
  database_name,
  schema_name,
  table_name,
  index_name,
  index_scans,
  usage_status,
  recommendation
FROM metrics_pg_bloat_indexes
WHERE time > NOW() - INTERVAL '1 day'
  AND usage_status IN ('UNUSED', 'RARELY_USED')
ORDER BY index_scans ASC
LIMIT 20
```

#### 6. Average Table Bloat Trend (7 days)
- **Type**: Time Series
- **Metric**: AVG(dead_ratio_percent)
- **Time Window**: Last 7 days
- **Use Case**: Bloat growth/shrinkage trend

#### 7. Bloat Analysis Details
- **Type**: Table (Combined metrics)
- **Purpose**: Comprehensive bloat overview

---

## 4. Cache Performance Dashboard

**File**: `dashboards/cache-performance.json`
**Purpose**: Monitor buffer pool and cache efficiency

### Panels (6)

#### 1. Average Cache Hit Ratio (24h)
- **Type**: Stat
- **Metric**: AVG(cache_hit_ratio)
- **Thresholds**:
  - Red: < 90%
  - Yellow: 90-95%
  - Green: > 95%
- **Use Case**: Overall cache health

#### 2. Average Index Cache Hit Ratio
- **Type**: Stat
- **Metric**: AVG(index_cache_hit_ratio)
- **Thresholds**:
  - Red: < 85%
  - Yellow: 85-95%
  - Green: > 95%
- **Use Case**: Index buffer pool efficiency

#### 3. Total Heap Block Misses (24h)
- **Type**: Stat
- **Metric**: SUM(heap_blks_miss)
- **Thresholds**:
  - Green: < 1M misses
  - Yellow: 1M-10M misses
  - Red: > 10M misses
- **Use Case**: Memory pressure indicator

#### 4. Cache Hit Ratio Trend (7 days)
- **Type**: Time Series
- **Metric**: AVG(cache_hit_ratio)
- **Breakdown**: Heap cache performance over time
- **Use Case**: Cache trend analysis

#### 5. Index Cache Hit Ratio Trend (7 days)
- **Type**: Time Series
- **Metric**: AVG(index_cache_hit_ratio)
- **Use Case**: Index buffer efficiency trends

#### 6. Cache Performance by Table (24h)
- **Type**: Table
- **Limits**: Top 50 tables
- **Columns**:
  - database_name
  - schema_name
  - table_name
  - cache_hit_ratio
  - heap_hits
  - heap_misses
  - status (color-coded)
- **Status Color Coding**:
  - 🔴 CRITICAL: < 80%
  - 🟠 WARNING: 80-90%
  - 🟢 HEALTHY: > 90%

---

## 5. Connection Tracking Dashboard

**File**: `dashboards/connection-tracking.json`
**Purpose**: Monitor sessions and detect connection issues

### Panels (7)

#### 1. Total Active Connections
- **Type**: Stat
- **Metric**: COUNT(*) where state='active'
- **Thresholds**:
  - Green: 0-50
  - Yellow: 50-100
  - Orange: 100-200
  - Red: > 200
- **Refresh**: 1 minute
- **Use Case**: Connection load

#### 2. Idle Connections
- **Type**: Stat
- **Metric**: COUNT(*) where state='idle'
- **Thresholds**:
  - Green: 0-10
  - Yellow: 10-25
  - Orange: 25-50
  - Red: > 50
- **Use Case**: Connection pool efficiency

#### 3. Idle in Transaction
- **Type**: Stat
- **Metric**: COUNT(*) where state='idle in transaction'
- **Thresholds**:
  - Green: 0
  - Yellow: 1-5
  - Red: > 5
- **Use Case**: Held transaction detection (critical)

```sql
SELECT
  COUNT(*) as idle_in_txn_connections
FROM metrics_pg_connections
WHERE time > NOW() - INTERVAL '5 minutes'
  AND state = 'idle in transaction'
```

#### 4. Connection State Distribution (24h)
- **Type**: Time Series
- **Breakdown**: By state (active, idle, idle in transaction, etc.)
- **Use Case**: Connection state patterns

#### 5. Connections by Database (7d)
- **Type**: Time Series (Stacked Bars)
- **Breakdown**: By database_name
- **Time Window**: Last 7 days
- **Use Case**: Multi-database connection tracking

#### 6. Long-Running & Idle Transactions (Last Hour)
- **Type**: Table
- **Limits**: Top 50
- **Filters**:
  - state='idle in transaction' (duration > any)
  - OR duration > 60 seconds
- **Columns**:
  - database_name
  - username
  - state
  - query_duration_seconds
  - query (truncated)
  - severity (color-coded)
- **Severity Color Coding**:
  - 🔴 CRITICAL: Idle in txn > 300s
  - 🟠 WARNING: Idle in txn
  - 🔴 SLOW QUERY: Active > 3600s
  - 🟢 NORMAL: Otherwise

---

## 6. Extensions Configuration Dashboard

**File**: `dashboards/extensions-config.json`
**Purpose**: Extension inventory and tracking

### Panels (6)

#### 1. Total Extensions Installed
- **Type**: Stat
- **Metric**: COUNT(DISTINCT extension_name)
- **Time Window**: Last 24 hours
- **Use Case**: Extension count

#### 2. Databases with Extensions
- **Type**: Stat
- **Metric**: COUNT(DISTINCT database_name)
- **Use Case**: Extension deployment scope

#### 3. UUID-OSSP Installations
- **Type**: Stat
- **Metric**: COUNT where extension_name='uuid-ossp'
- **Thresholds**:
  - Yellow: 1+ (useful for adoption tracking)
- **Use Case**: Specific extension tracking

#### 4. Extensions Distribution (30 days)
- **Type**: Time Series (Stacked Bars)
- **Breakdown**: By extension_name
- **Time Window**: Last 30 days
- **Use Case**: Extension adoption trends

#### 5. Extension Count Trend (30 days)
- **Type**: Time Series
- **Metric**: COUNT(DISTINCT extension_name)
- **Use Case**: Overall extension growth

#### 6. Extension Inventory (24h)
- **Type**: Table
- **Limits**: Top 100 entries
- **Columns**:
  - database_name
  - extension_name
  - extension_version
  - schema_name
  - extension_owner
  - installations (count)
- **Use Case**: Complete extension reference

```sql
SELECT
  database_name,
  extension_name,
  extension_version,
  schema_name,
  extension_owner,
  COUNT(*) as installations
FROM metrics_pg_extensions
WHERE time > NOW() - INTERVAL '1 day'
GROUP BY database_name, extension_name, extension_version, schema_name, extension_owner
ORDER BY extension_name, database_name
LIMIT 100
```

---

## 7. System Overview Dashboard

**File**: `dashboards/system-overview.json`
**Purpose**: Consolidated health and key metrics

### Panels (8)

#### 1-4. Key Metrics (Stats)
Layout: Top row with 4 stat panels (6 units wide each)

- **Active Connections**: Current active session count
- **Active Locks**: Current granted locks
- **Avg Cache Hit Ratio**: 24-hour average
- **Max Table Bloat %**: Worst bloat percentage

#### 5. Connection Trend (24h)
- **Type**: Time Series
- **Metric**: COUNT(*) connections
- **Use Case**: Connection activity over time

#### 6. Cache Hit Ratio Trend (7d)
- **Type**: Time Series
- **Metric**: AVG(cache_hit_ratio)
- **Use Case**: Cache performance trend

#### 7. Database Health Summary
- **Type**: Table
- **Purpose**: One-row-per-database health status
- **Columns**:
  - database_name
  - active_connections
  - max_bloat_percent
  - avg_cache_hit_ratio
  - extension_count
  - health_status (color-coded)
- **Health Status Logic**:
  - 🔴 CRITICAL: bloat > 50%
  - 🟠 WARNING: cache_hit < 80%
  - 🟢 HEALTHY: Otherwise

```sql
SELECT
  database_name,
  COALESCE(
    (SELECT COUNT(*) FROM metrics_pg_connections
     WHERE database_name = m.database_name AND time > NOW() - INTERVAL '5 minutes'
     AND state = 'active'), 0
  ) as active_connections,
  COALESCE(
    (SELECT MAX(dead_ratio_percent) FROM metrics_pg_bloat_tables
     WHERE database_name = m.database_name AND time > NOW() - INTERVAL '24 hours'), 0
  ) as max_bloat_percent,
  COALESCE(
    (SELECT AVG(cache_hit_ratio) FROM metrics_pg_cache_hit_ratios
     WHERE database_name = m.database_name AND time > NOW() - INTERVAL '24 hours'), 0
  ) as avg_cache_hit_ratio,
  COALESCE(
    (SELECT COUNT(DISTINCT extension_name) FROM metrics_pg_extensions
     WHERE database_name = m.database_name AND time > NOW() - INTERVAL '1 day'), 0
  ) as extension_count,
  CASE
    WHEN ... THEN '🔴 CRITICAL'
    WHEN ... THEN '🟠 WARNING'
    ELSE '🟢 HEALTHY'
  END as health_status
FROM metrics_pg_schema_tables m
WHERE time > NOW() - INTERVAL '1 day'
GROUP BY database_name
ORDER BY database_name
LIMIT 50
```

---

## Deployment Instructions

### 1. Prerequisites

- Grafana 8.0+ installed and running
- PostgreSQL datasource configured as "PostgreSQL"
- TimescaleDB backend with Phase 3 migrations applied
- API collector running and sending metrics

### 2. Dashboard Import

#### Method A: Manual JSON Import

```bash
# For each dashboard file:
1. Go to Grafana: http://localhost:3000
2. Dashboards → Import
3. Paste JSON content or upload file
4. Select "PostgreSQL" datasource
5. Click Import
```

#### Method B: Automated Script

```bash
#!/bin/bash
GRAFANA_URL="http://localhost:3000"
API_KEY="your-grafana-api-key"
DASHBOARD_DIR="dashboards"

for dashboard in $DASHBOARD_DIR/*.json; do
  curl -X POST "$GRAFANA_URL/api/dashboards/db" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d @"$dashboard"
done
```

#### Method C: Grafana Provisioning

```yaml
# /etc/grafana/provisioning/dashboards/dashboards.yml
apiVersion: 1
providers:
  - name: 'pgAnalytics Phase 4'
    orgId: 1
    folder: 'PostgreSQL'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

Then copy JSON files to `/var/lib/grafana/dashboards/`

### 3. Datasource Configuration

Ensure PostgreSQL datasource is configured:

```
URL: postgresql://localhost:5432/pganalytics
Username: grafana
Password: [secure password]
SSL Mode: disable (or require, based on setup)
Database: pganalytics
```

### 4. Alert Rules (Optional)

Configure alert notifications in Grafana:

**For Lock Monitoring Dashboard**:
```
Alert: Active Locks > 10
Condition: stat value > 10
Action: Send to Slack/PagerDuty
```

**For Bloat Analysis Dashboard**:
```
Alert: Max Bloat > 50%
Condition: stat value > 50
Action: Send to Slack/PagerDuty
```

**For Cache Performance Dashboard**:
```
Alert: Cache Hit Ratio < 80%
Condition: time series avg < 80
Action: Send to Slack/PagerDuty
```

---

## Monitoring Best Practices

### 1. Normal Operating Ranges

| Metric | Green | Yellow | Red |
|--------|-------|--------|-----|
| Cache Hit Ratio | > 95% | 90-95% | < 90% |
| Max Lock Age | < 300s | 300-1800s | > 1800s |
| Table Bloat | < 10% | 10-25% | > 50% |
| Idle in Txn | 0 | 1-5 | > 5 |
| Active Connections | < 50 | 50-100 | > 200 |

### 2. Troubleshooting Guide

**High Lock Count?**
- Check "Lock Blocking Chain" table
- Identify blocking query
- Review query execution plan
- Consider KILL query if necessary

**Low Cache Hit Ratio?**
- Increase shared_buffers
- Check for full table scans
- Review expensive queries
- Monitor memory pressure

**High Bloat?**
- Schedule VACUUM FULL during maintenance window
- REINDEX if bloat is persistent
- Monitor after cleanup

**Idle in Transaction?**
- Find long-running transaction
- Notify application team
- Review application connection handling
- Consider connection timeout

### 3. Regular Maintenance

**Daily**:
- Monitor System Overview for health status
- Check Lock Monitoring for blocking chains
- Review Cache Performance trends

**Weekly**:
- Review Bloat Analysis for cleanup candidates
- Analyze Connection Tracking patterns
- Check Extension updates

**Monthly**:
- Review schema changes in Schema Overview
- Plan bloat cleanup operations
- Analyze lock and performance patterns

---

## Performance Considerations

### Query Optimization

All dashboard queries are optimized for performance:

- ✅ Use TimescaleDB hypertable optimization
- ✅ Time-based filtering (time > NOW() - INTERVAL '...')
- ✅ Aggregation at query level, not visualization
- ✅ Pagination with LIMIT clauses
- ✅ Index utilization for GROUP BY operations

### Refresh Rates

- **Critical Dashboards** (1m): Lock Monitoring, Connection Tracking, System Overview
  - Real-time operational awareness
  - Minimal query overhead

- **Standard Dashboards** (5m): Schema, Bloat, Cache, Extensions
  - Trend analysis focus
  - Balanced performance/accuracy

### Scaling

For large environments (1000+ databases):

1. **Create database-specific views** in Grafana
2. **Use time-based bucketing** in queries
3. **Consider data retention policies** (30-90 days)
4. **Implement dashboard variables** for filtering

---

## Integration with Alerting

### Grafana Alert Manager

Configure notifications for critical conditions:

```json
{
  "dashboard": "lock-monitoring",
  "panels": [
    {
      "title": "Active Locks",
      "alert": {
        "name": "High Lock Count",
        "message": "Active locks > 10",
        "for": "5m"
      }
    }
  ]
}
```

### External Alerting

Export alerts to external systems:

**Prometheus Export**:
```
/metrics endpoint for Prometheus scraping
```

**Webhook Integration**:
```
POST /webhook/metrics
{
  "dashboard": "name",
  "alert": "condition",
  "value": 123,
  "timestamp": "2026-03-03T10:00:00Z"
}
```

---

## Dashboard Folder Structure (Recommended)

```
PostgreSQL
├── Phase 3 Metrics (Folder)
│   ├── Schema Overview
│   ├── Lock Monitoring
│   ├── Bloat Analysis
│   ├── Cache Performance
│   ├── Connection Tracking
│   ├── Extensions Config
│   └── System Overview
├── Alerts (Folder)
│   ├── Critical Alerts
│   └── Warning Alerts
└── Reports (Folder)
    ├── Weekly Summary
    └── Monthly Analysis
```

---

## Verification Checklist

After deployment, verify:

- ✅ All 7 dashboards import without errors
- ✅ PostgreSQL datasource connects successfully
- ✅ All queries execute without SQL errors
- ✅ All 46 panels render with data
- ✅ Refresh rates work correctly
- ✅ Color thresholds display properly
- ✅ Time pickers function correctly
- ✅ Tables display data with pagination
- ✅ Time series show trends over time
- ✅ Stat panels show current values

---

## Rollback Procedure

If issues occur:

1. **Stop metrics collection** (set plugin=off in config)
2. **Delete dashboards** in Grafana UI
3. **Verify data integrity** in backend database
4. **Restart Grafana** service
5. **Resume metrics collection** (set plugin=on)

---

## Success Metrics

**Phase 4 Deployment Success** when:

✅ All 7 dashboards deployed and accessible
✅ 46 visualization panels displaying data
✅ No SQL errors in dashboard queries
✅ Average query response time < 2 seconds
✅ Dashboard navigation time < 1 second
✅ Alert notifications working correctly
✅ Team can identify and resolve issues faster
✅ 95%+ uptime of dashboard availability

---

## Support & Troubleshooting

### Common Issues

**"No Data" in dashboard panels**:
- Verify metrics are being collected (check backend logs)
- Check time filter in dashboard
- Verify PostgreSQL datasource connectivity

**Slow dashboard load**:
- Check Grafana server resources
- Review query performance in PostgreSQL
- Reduce time window in queries

**Incorrect threshold colors**:
- Verify threshold values in field config
- Check metric units (percent, bytes, seconds)
- Test with known values

---

## Documentation References

- **Phase 3 Metrics**: PHASE3_TESTING_INTEGRATION_COMPLETE.md
- **API Architecture**: API_ARCHITECTURE_EXPLANATION.md
- **Quick Reference**: PHASE3_QUICK_REFERENCE.md
- **Production Plan**: PHASE4_PRODUCTION_DEPLOYMENT_PLAN.md

---

**Status**: ✅ Phase 4 Grafana Dashboards Complete and Documented

All 7 production dashboards are ready for deployment. See PHASE4_PRODUCTION_DEPLOYMENT_PLAN.md for complete rollout procedures.
