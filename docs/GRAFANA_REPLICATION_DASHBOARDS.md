# Grafana Replication Dashboards - Setup & User Guide

**Date:** February 25, 2026
**Version:** pgAnalytics v3.2.0 Phase 1
**Status:** ✅ Production Ready

---

## Overview

Two comprehensive Grafana dashboards have been created to visualize PostgreSQL replication metrics collected by the Phase 1 Replication Collector:

1. **PostgreSQL Replication Health Monitor** - Core replication health metrics
2. **PostgreSQL Replication Advanced Analytics** - Detailed trend analysis and insights

---

## Table of Contents

1. [Dashboard Installation](#dashboard-installation)
2. [PostgreSQL Replication Health Monitor](#postgresql-replication-health-monitor)
3. [PostgreSQL Replication Advanced Analytics](#postgresql-replication-advanced-analytics)
4. [Configuration & Data Source](#configuration--data-source)
5. [Alerts & Thresholds](#alerts--thresholds)
6. [Troubleshooting](#troubleshooting)

---

## Dashboard Installation

### Prerequisites

- **Grafana**: 7.0+ (tested with 8.x, 9.x, 10.x)
- **PostgreSQL Data Source**: Already configured in Grafana
- **Replication Metrics**: Being collected via pgAnalytics replication collector
- **Metrics Database**: PostgreSQL backend receiving metrics from collector

### Installation Steps

#### Option 1: Auto-Provisioned (Recommended)

The dashboards are automatically provisioned if using Grafana provisioning:

1. **Verify Provisioning Directory**
   ```bash
   # Default Docker location
   /var/lib/grafana/dashboards

   # Or check docker-compose volume mount
   volumes:
     - ./grafana/dashboards:/var/lib/grafana/dashboards
   ```

2. **Copy Dashboard Files**
   ```bash
   cp grafana/dashboards/replication-health-monitor.json \
      /var/lib/grafana/dashboards/
   cp grafana/dashboards/replication-advanced-analytics.json \
      /var/lib/grafana/dashboards/
   ```

3. **Restart Grafana**
   ```bash
   docker restart grafana
   # or
   systemctl restart grafana-server
   ```

4. **Verify**
   - Open Grafana: http://localhost:3000
   - Navigate to: Dashboards → pgAnalytics folder
   - Should see both replication dashboards

#### Option 2: Manual Import

1. **Log in to Grafana**
   - URL: http://localhost:3000
   - Default credentials: admin / admin

2. **Import Dashboard**
   - Click: + (Create) → Import
   - Select: Upload JSON file
   - Choose: `replication-health-monitor.json`
   - Click: Import
   - Repeat for `replication-advanced-analytics.json`

3. **Select Data Source**
   - When prompted, select your PostgreSQL datasource
   - Click: Import

---

## PostgreSQL Replication Health Monitor

**UID:** `pg-replication-health`
**Type:** Overview dashboard
**Refresh:** 30 seconds
**Time Range:** Last 6 hours (configurable)

### Dashboard Panels

#### 1. **Max Replica Replay Lag** (Stat)
- **Location:** Top-left
- **Metric:** Latest replay lag in milliseconds
- **Thresholds:**
  - Green: < 1 second (1000 ms)
  - Yellow: 1-5 seconds
  - Red: > 5 seconds
- **Alert:** Red indicates replication falling behind

#### 2. **Min XID Wraparound % Remaining** (Stat)
- **Location:** Top-center
- **Metric:** Minimum percentage of XID space remaining across all databases
- **Thresholds:**
  - Red: < 20% (at risk)
  - Yellow: 20-50%
  - Green: > 50%
- **Alert:** Red means wraparound vacuum is urgent

#### 3. **Connected Replicas** (Stat)
- **Location:** Top-right (1)
- **Metric:** Number of active replication connections
- **Interpretation:** Should match expected replica count
- **Alert:** Drop indicates connection issue

#### 4. **Replication Slots** (Stat)
- **Location:** Top-right (2)
- **Metric:** Number of replication slots
- **Interpretation:** Should match slot count
- **Alert:** Increase may indicate new slot creation

#### 5. **Replication Replay Lag Over Time** (Time Series)
- **Location:** Middle-left
- **Metric:** Replay lag trends for each connected replica
- **X-Axis:** Time
- **Y-Axis:** Lag in milliseconds
- **Legend:** Shows mean and max lag for each replica

#### 6. **WAL Directory Size Over Time** (Time Series)
- **Location:** Middle-right
- **Metric:** WAL directory size in MB over time
- **X-Axis:** Time
- **Y-Axis:** Size in MB
- **Trend:** Should be relatively stable (growth = active DML)

#### 7. **XID Wraparound Risk % by Database** (Time Series)
- **Location:** Bottom-left
- **Metric:** XID space remaining percentage for each database
- **Threshold Line:** 20% (red zone)
- **Trend:** Should remain > 50% (below 20% requires immediate action)

#### 8. **Replica Sync State Distribution** (Pie Chart)
- **Location:** Bottom-center
- **Metric:** Count of sync vs async replicas
- **Interpretation:**
  - "sync": Synchronous replicas (high durability)
  - "async": Asynchronous replicas (high performance)

#### 9. **Replication Status Details** (Table)
- **Location:** Bottom (full width)
- **Columns:**
  - Replica: Application name/identifier
  - State: streaming, catchup, or backup
  - Sync State: sync or async
  - Lag (ms): Current replay lag
  - Client IP: Replica IP address
  - Connected Since: Connection start time
- **Use:** Get detailed connection info

---

## PostgreSQL Replication Advanced Analytics

**UID:** `pg-replication-analytics`
**Type:** Analytics dashboard
**Refresh:** 30 seconds
**Time Range:** Last 7 days (detailed trend analysis)

### Dashboard Panels

#### 1. **Replication Lag Analysis** (Time Series - Stacked)
- **Location:** Top-left
- **Metrics:** Write, Flush, and Replay lag breakdown
- **Stacking:** Normal (stacked)
- **Insight:** Shows lag at different stages:
  - Write lag: Time to write to replica buffer
  - Flush lag: Time to flush to disk
  - Replay lag: Time to apply changes
- **Analysis:** Usually write_lag < flush_lag < replay_lag

#### 2. **WAL Growth Rate** (Bar Chart)
- **Location:** Top-right
- **Metric:** WAL growth in MB/hour
- **Trend:** Shows rate of data change
- **High Rate:** Active workload, high I/O
- **Low Rate:** Idle or light workload
- **Spike:** May indicate bulk operation

#### 3. **WAL Segments Analysis** (Time Series)
- **Location:** Middle-left
- **Metrics:**
  - Total Segments: Total WAL files
  - Segments Since Checkpoint: WAL generated since last checkpoint
- **Trend:** Checkpoint should reset "since checkpoint" regularly

#### 4. **Replication Slot Type Distribution** (Pie Chart)
- **Location:** Middle-right
- **Metric:** Count of physical vs logical slots
- **Interpretation:**
  - Physical: For streaming replication
  - Logical: For logical replication and subscriptions

#### 5. **XID Wraparound Risk Trend** (Time Series - All Databases)
- **Location:** Middle (full width)
- **Metric:** Percentage remaining per database over time
- **Threshold Line:** 20% (red zone)
- **Analysis:**
  - Trend should be relatively stable
  - Downward trend indicates autovacuum not keeping up
  - Below 20% requires manual vacuum

#### 6. **Replica Lag Statistics** (Table)
- **Location:** Bottom-left (full width)
- **Columns:**
  - Replica Name: Application identifier
  - Max Lag (ms): Highest lag observed
  - Avg Lag (ms): Average lag
  - Min Lag (ms): Lowest lag observed
  - Measurements: Number of data points
- **Use:** Identify chronically lagging replicas

#### 7. **WAL Status History** (Table)
- **Location:** Bottom-right (full width)
- **Columns:**
  - Current WAL Size (MB): Total WAL directory size
  - WAL Dir Size (MB): Same as above
  - Growth Rate (MB/h): Estimated hourly growth
  - Total Segments: Count of WAL segments
  - Segments Since CP: Segments since last checkpoint
- **Use:** Track WAL growth patterns

---

## Configuration & Data Source

### PostgreSQL Datasource Setup

Both dashboards require a PostgreSQL datasource pointing to the metrics backend database.

#### Datasource Configuration

1. **Data Source Name**: Must match dashboard datasource UID
   - Default UID: `P4755FD0186DF985F` (from existing dashboards)
   - Check your Grafana: Admin → Data Sources

2. **Connection Details**
   ```
   PostgreSQL Connection
   - Host: metrics-db.example.com
   - Port: 5432
   - Database: pganalytics_metrics
   - User: grafana_user (read-only)
   - SSL Mode: require (production)
   ```

3. **User Permissions**
   ```sql
   -- Create read-only user for Grafana
   CREATE ROLE grafana_user WITH LOGIN;
   GRANT CONNECT ON DATABASE pganalytics_metrics TO grafana_user;
   GRANT USAGE ON SCHEMA public TO grafana_user;
   GRANT SELECT ON TABLE metrics TO grafana_user;
   ```

### Dashboard Variables

Both dashboards use hard-coded queries. To add dynamic filtering:

1. **Add Collector Selector**
   ```
   Name: $collector
   Type: Query
   Query: SELECT DISTINCT collector_id FROM metrics
   ```

2. **Update Panel Queries**
   ```sql
   -- Add WHERE clause
   WHERE metrics->>'type' = 'pg_replication'
   AND collector_id = '$collector'
   ```

---

## Alerts & Thresholds

### Recommended Alert Rules

#### Alert 1: High Replication Lag

```yaml
Alert Name: ReplicationLagHigh
Condition: Max(replay_lag_ms) > 10000
For: 5 minutes
Severity: Warning
Action: Notify on-call team
```

**Query for Grafana Alert:**
```sql
SELECT
  EXTRACT(EPOCH FROM NOW())::BIGINT as timestamp,
  COALESCE(MAX(CAST(metrics->'replication_status'->0->>'replay_lag_ms' AS BIGINT)), 0) as value
FROM metrics
WHERE metrics->>'type' = 'pg_replication'
ORDER BY timestamp DESC LIMIT 1
```

#### Alert 2: Wraparound Risk

```yaml
Alert Name: XIDWraparoundRisk
Condition: Min(percent_until_wraparound) < 20
For: 10 minutes
Severity: Critical
Action: Page on-call, trigger emergency vacuum
```

**Query for Grafana Alert:**
```sql
SELECT
  EXTRACT(EPOCH FROM NOW())::BIGINT as timestamp,
  COALESCE(MIN(CAST(metrics->'wraparound_risk'->0->>'percent_until_wraparound' AS INT)), 100) as value
FROM metrics
WHERE metrics->>'type' = 'pg_replication'
ORDER BY timestamp DESC LIMIT 1
```

#### Alert 3: Connection Loss

```yaml
Alert Name: ReplicaConnectionLost
Condition: Connected Replicas < expected_count
For: 1 minute
Severity: Critical
Action: Page on-call immediately
```

#### Alert 4: Slot Inactive

```yaml
Alert Name: ReplicationSlotInactive
Condition: Active Slots < Total Slots
For: 5 minutes
Severity: Warning
Action: Check for stuck or failed slots
```

### Manual Alert Configuration

1. **Open Dashboard**
   - Click: Dashboard → Alerting → Create Alert

2. **Select Panel**
   - Choose one of the stat panels
   - Define threshold (e.g., 5000 ms for lag)

3. **Configure Notification**
   - Channel: Slack, PagerDuty, Email
   - Message: "Replication lag exceeded threshold"

4. **Save Alert**
   - Dashboard auto-saves
   - Alert activates immediately

---

## Troubleshooting

### Dashboard Shows "No Data"

**Cause 1: Metrics not being collected**
- Check collector logs: `docker logs pganalytics-collector`
- Verify PostgreSQL connection works
- Check: `SELECT count(*) FROM metrics WHERE type='pg_replication'`

**Solution:**
```bash
# Verify replication metrics exist
psql -h localhost -d pganalytics_metrics -c \
  "SELECT type, count(*) FROM metrics GROUP BY type"

# Expected output should include pg_replication row
```

**Cause 2: Datasource not configured correctly**
- Check datasource connection
- Verify UID matches dashboard datasource

**Solution:**
1. Admin → Data Sources
2. Select PostgreSQL datasource
3. Click "Test" to verify connection
4. Update dashboard datasource UID if needed

**Cause 3: Metrics table not created**
- Backend may not have created schema yet

**Solution:**
```bash
# Create metrics table manually if missing
psql -h localhost -d pganalytics_metrics -f \
  backend/schema/metrics.sql
```

### Lag Panel Shows Extreme Values

**Cause:** Null values from PG < 13 (before lag metrics added)

**Solution:**
- Dashboard handles this with COALESCE(... 0)
- Ensure PostgreSQL 13+ in production
- Data normalizes as old metrics age out

### Wraparound Panel Not Showing Data

**Cause:** Wraparound metrics require pg_database access

**Solution:**
1. Verify collector user has superuser or pg_monitor
   ```sql
   GRANT pg_monitor TO pganalytics;
   ```
2. Restart collector
3. Wait 1-2 collection cycles for data

### Sync State Shows No Distribution

**Cause:** No replicas currently connected

**Solution:**
1. Check replica connections: `SELECT * FROM pg_stat_replication`
2. Verify standby is configured correctly
3. Check network connectivity
4. Check replication credentials

### High CPU Usage When Viewing Dashboard

**Cause:** Complex JSONB queries on large metrics table

**Solution:**
1. **Add Index**
   ```sql
   CREATE INDEX idx_metrics_type_timestamp
   ON metrics (metrics->>'type', timestamp DESC)
   WHERE metrics->>'type' = 'pg_replication';
   ```

2. **Enable Dashboard Caching**
   - Dashboard → Cog → Cache timeout: 60 seconds

3. **Archive Old Metrics**
   ```sql
   -- Archive metrics older than 30 days
   DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '30 days';
   VACUUM ANALYZE metrics;
   ```

### Grafana Not Picking Up Provisioned Dashboards

**Solution:**
```bash
# Check Grafana log
docker logs -f grafana

# Verify provisioning path
docker exec grafana ls -la /var/lib/grafana/dashboards

# Manually trigger reload
docker exec grafana curl -X POST \
  http://admin:admin@localhost:3000/api/admin/provisioning/dashboards/reload
```

---

## Performance Optimization

### Database Queries

The dashboards use efficient JSON queries:

```sql
-- Example: Extract single value from nested JSON
CAST(metrics->'replication_status'->0->>'replay_lag_ms' AS BIGINT)

-- Example: Iterate through JSON array elements
jsonb_array_elements(metrics->'replication_status') WITH ORDINALITY
```

### Query Optimization Tips

1. **Add Indexes for Metrics**
   ```sql
   CREATE INDEX idx_metrics_type_ts ON metrics ((metrics->>'type'), timestamp DESC);
   CREATE INDEX idx_metrics_pg_rep ON metrics (timestamp DESC)
   WHERE metrics->>'type' = 'pg_replication';
   ```

2. **Partition Metrics by Date**
   ```sql
   CREATE TABLE metrics_2026_02 PARTITION OF metrics
   FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
   ```

3. **Adjust Dashboard Refresh Rate**
   - High (10s): Near real-time but high DB load
   - Medium (30s): Balanced (default)
   - Low (60s+): Low DB load, delayed visibility

4. **Archive Old Metrics**
   - Keep last 30-90 days in hot table
   - Archive to separate schema for historical analysis

---

## Advanced Customization

### Adding Custom Panels

To add custom panels to existing dashboards:

1. **Edit Dashboard**
   - Click: Dashboard → Edit

2. **Add Panel**
   - Click: + (Add panel)
   - Select: Visualization type
   - Enter: SQL query targeting metrics table

3. **Example Custom Panel: Replication Downtime**
   ```sql
   SELECT
     time_bucket('5 minutes', timestamp) as time,
     CASE WHEN COUNT(*) = 0 THEN 1 ELSE 0 END as downtime
   FROM metrics
   WHERE metrics->>'type' = 'pg_replication'
   GROUP BY time_bucket('5 minutes', timestamp)
   ORDER BY time DESC
   ```

### Creating Alert Notification Channels

```bash
# Create Slack notification channel
curl -X POST http://localhost:3000/api/alert-notifications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{
    "name": "Replication Alerts",
    "type": "slack",
    "isDefault": false,
    "settings": {
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    }
  }'
```

---

## Dashboard Metrics Reference

### Panel: Max Replica Replay Lag
- **Metric**: `metrics->'replication_status'->0->>'replay_lag_ms'`
- **Unit**: milliseconds
- **Importance**: CRITICAL (indicates replication status)

### Panel: Min XID Wraparound
- **Metric**: `metrics->'wraparound_risk'->0->>'percent_until_wraparound'`
- **Unit**: percentage (0-100%)
- **Importance**: CRITICAL (wraparound prevents all database changes)

### Panel: Connected Replicas
- **Metric**: `jsonb_array_length(metrics->'replication_status')`
- **Unit**: count
- **Importance**: HIGH (connection loss = data loss risk)

### Panel: Replication Slots
- **Metric**: `jsonb_array_length(metrics->'replication_slots')`
- **Unit**: count
- **Importance**: MEDIUM (slot health indicator)

### Panel: WAL Directory Size
- **Metric**: `metrics->'wal_status'->>'current_wal_size_mb'`
- **Unit**: megabytes
- **Importance**: MEDIUM (disk space planning)

### Panel: WAL Growth Rate
- **Metric**: `metrics->'wal_status'->>'growth_rate_mb_per_hour'`
- **Unit**: MB/hour
- **Importance**: MEDIUM (capacity planning)

---

## Support & Documentation

### Related Documentation
- [REPLICATION_COLLECTOR_GUIDE.md](REPLICATION_COLLECTOR_GUIDE.md) - Collector configuration
- [PHASE1_IMPLEMENTATION_SUMMARY.md](../PHASE1_IMPLEMENTATION_SUMMARY.md) - Architecture details
- [PostgreSQL Replication Docs](https://www.postgresql.org/docs/current/warm-standby.html)

### Dashboard Updates

To receive updates when new dashboards are released:
1. Check GitHub: https://github.com/torresglauco/pganalytics-v3/releases
2. Watch for `v3.2.0-phase2-*` releases with additional dashboards

---

**Version**: pgAnalytics v3.2.0 Phase 1
**Status**: ✅ Production Ready
**Last Updated**: February 25, 2026

