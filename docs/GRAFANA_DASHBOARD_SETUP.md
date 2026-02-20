# Grafana Dashboard Setup Guide - Query Performance Monitoring

## Overview

The Query Performance Monitoring dashboard provides comprehensive visualization of PostgreSQL query execution patterns, performance trends, and resource usage. It enables DBAs and DevOps teams to identify slow queries, optimize workloads, and track performance improvements.

## Dashboard Features

### 1. Quick Statistics Panels (Top Right)
- **Total Query Count**: Aggregate query executions in the selected time period
- **Peak Query Execution Time**: Maximum query time observed
- **Overall Cache Hit Ratio**: Buffer cache efficiency percentage
- **Unique Queries Count**: Number of distinct queries

### 2. Top Queries Analysis

#### Top 10 Queries by Total Execution Time (Pie Chart)
- Visualizes which queries consume the most cumulative time
- Color-coded by time spent
- Helps identify optimization priorities
- Interactive legend with metrics

#### Top 20 Slowest Queries (Table)
- Detailed breakdown of slowest queries
- Columns:
  - **Query**: Truncated SQL query text (first 80 chars)
  - **Calls**: Number of times executed
  - **Max Time**: Maximum single execution time
  - **Avg Time**: Average execution time
  - **Min Time**: Minimum execution time
  - **StdDev**: Standard deviation (consistency indicator)
- Sorted by maximum execution time
- Color coding: Green (good), Yellow (warning), Red (alert)

### 3. Execution Frequency Trend
- **Query Execution Frequency Trend**: Shows query volume over time
- Bar chart with stacked execution counts
- Helps identify peak query periods
- 24-hour view with hourly granularity

### 4. Performance Metrics

#### Average Query Execution Time Trend
- Line chart showing execution time trends
- Color-coded thresholds (green < 1000ms, red > 1000ms)
- Includes mean and max values in legend
- Identifies performance degradation

#### Top 20 Most Frequent Queries (Table)
- Queries executed most often
- Columns:
  - **Query**: SQL query text
  - **Calls**: Total executions
  - **Rows Processed**: Total rows affected
  - **Cache Hit %**: Cache efficiency for this query
  - **I/O Read Time**: Total block read time
- Useful for optimization (high-frequency = high impact)

### 5. Cache Efficiency Monitoring

#### Buffer Cache Hit Ratio Trend
- Shows shared buffer pool efficiency over time
- Line chart with threshold thresholds
- Target: >95% (green), >80% (yellow), <80% (red)
- Smooth interpolation for clarity
- Identifies cache misses and potential disk I/O issues

### 6. I/O Performance Metrics

#### I/O Read Time Trend
- Block read time over 24 hours
- Identifies I/O bottlenecks
- Color thresholds: Green <100ms, Yellow 100-500ms, Red >500ms
- Mean and max values displayed

#### I/O Write Time Trend
- Block write time monitoring
- Complements read time analysis
- Important for write-heavy workloads
- Same thresholds as read time

## Dashboard Navigation

### Time Range Selection
- **1 hour**: Real-time monitoring, detailed view
- **6 hours**: Short-term trends, recent performance
- **24 hours** (default): Daily analysis, overnight behavior
- **7 days**: Weekly trends, pattern identification
- **30 days**: Monthly trends, long-term analysis

### Interactive Features
- **Click pie chart slices**: Drill down to specific queries
- **Table sorting**: Click column headers to re-sort
- **Hover over panels**: See exact values and details
- **Zoom graphs**: Select time range on any chart for detailed view
- **Legend toggles**: Click legend items to show/hide series

## Data Sources

All panels query the TimescaleDB `metrics_pg_stats_query` table with these key data:

### Raw Metrics (metrics_pg_stats_query)
- Per-collection query statistics
- 30-day retention
- All pg_stat_statements fields

### Aggregated Metrics (metrics_pg_stats_query_1h)
- Hourly rollups for efficiency
- Reduces query load by 60x
- Ideal for dashboard visualization
- Automatically refreshed by TimescaleDB

## Query Examples

### Get Slowest Query Details
```sql
SELECT
  query_hash,
  query_text,
  calls,
  total_time,
  mean_time,
  max_time,
  min_time,
  stddev_time
FROM metrics_pg_stats_query
WHERE time >= NOW() - INTERVAL '24 hours'
ORDER BY max_time DESC
LIMIT 1;
```

### Identify High-Frequency Queries
```sql
SELECT
  query_text,
  SUM(calls) as total_calls,
  AVG(mean_time) as avg_time,
  SUM(rows) as total_rows
FROM metrics_pg_stats_query
WHERE time >= NOW() - INTERVAL '24 hours'
GROUP BY query_hash, query_text
ORDER BY SUM(calls) DESC
LIMIT 20;
```

### Analyze Cache Efficiency
```sql
SELECT
  time,
  SUM(shared_blks_hit) as hits,
  SUM(shared_blks_read) as reads,
  ROUND(100.0 * SUM(shared_blks_hit) /
    NULLIF(SUM(shared_blks_hit) + SUM(shared_blks_read), 0), 2) as hit_ratio
FROM metrics_pg_stats_query_1h
WHERE time >= NOW() - INTERVAL '24 hours'
GROUP BY time
ORDER BY time;
```

### Find I/O Hotspots
```sql
SELECT
  SUBSTRING(query_text, 1, 80) as query,
  SUM(blk_read_time) as total_read_time,
  SUM(blk_write_time) as total_write_time,
  SUM(calls) as executions
FROM metrics_pg_stats_query
WHERE time >= NOW() - INTERVAL '24 hours'
GROUP BY query_hash, query_text
ORDER BY (SUM(blk_read_time) + SUM(blk_write_time)) DESC
LIMIT 20;
```

## Interpretation Guide

### What to Look For

#### Performance Red Flags
- **Spikes in max query time**: Query degradation, missing indexes
- **Increasing execution frequency**: Popular slow queries
- **Dropping cache hit ratio**: Insufficient buffer pool, large scans
- **High I/O times**: Disk I/O saturation, sequential scans

#### Healthy Patterns
- **Stable execution times**: Consistent performance
- **High cache hit ratio (>95%)**: Efficient buffer usage
- **Low I/O times**: Query optimization working
- **Predictable frequency**: Normal workload patterns

### Optimization Workflow

1. **Identify Problem Queries**
   - Look at "Top 20 Slowest Queries" table
   - Check "Average Execution Time Trend" for degradation
   - Note unique query characteristics

2. **Analyze Query Characteristics**
   - Check "Cache Hit %" in top frequent queries
   - Review I/O times from slow queries table
   - Look for high row counts relative to returned rows

3. **Apply Optimizations**
   - Add missing indexes (suggested from cache misses)
   - Rewrite queries (especially sequential scans)
   - Adjust parameters (work_mem, shared_buffers)

4. **Monitor Results**
   - Watch execution time trends
   - Track cache hit ratio improvements
   - Verify I/O reduction

## Dashboard Configuration

### Refresh Rate
- Default: 30 seconds
- Configurable in Grafana dashboard settings
- Lower for real-time monitoring (10-15 seconds)
- Higher for historical analysis (1-5 minutes)

### Color Scheme
- **Dark theme** (default): Better for NOC/war rooms
- Switchable to light theme in Grafana settings

### Panel Organization
- **Top row**: Summary pie chart and trend analysis
- **Middle sections**: Detailed query tables
- **Lower sections**: Frequency and I/O trends
- **Bottom row**: Key metrics and statistics

## Setting Up Alerts

### Recommended Alert Rules

#### 1. Slow Query Alert
```
IF: max query execution time > 5000ms (last 5 minutes)
THEN: Notify ops team
CONDITION: ON any query
```

#### 2. Cache Hit Ratio Alert
```
IF: cache hit ratio < 85% (last 30 minutes)
THEN: Notify DBA
CONDITION: Check buffer pool config
```

#### 3. High I/O Time Alert
```
IF: avg I/O read time > 500ms (last 15 minutes)
THEN: Investigate disk I/O
CONDITION: Check I/O bottlenecks
```

#### 4. Query Frequency Spike
```
IF: query executions > 3x daily average
THEN: Possible N+1 query pattern
CONDITION: Review application logs
```

## Dashboard Installation

### Docker Compose Method
```yaml
version: '3.9'
services:
  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: false
```

### Manual Installation
1. Open Grafana UI (http://localhost:3000)
2. Go to **Dashboards** → **Import**
3. Upload `query-performance.json` file
4. Select **TimescaleDB** as data source
5. Click **Import**

### Kubernetes Helm
```yaml
grafana:
  dashboards:
    default:
      query-performance:
        url: https://raw.githubusercontent.com/.../query-performance.json
```

## Data Source Configuration

### TimescaleDB Connection
```
Host: timescaledb.default.svc.cluster.local
Port: 5432
Database: pganalytics
User: pg_monitor
Password: (from secrets)
SSL Mode: disable (or enable for production)
```

### Health Check
```sql
SELECT 1 FROM metrics_pg_stats_query LIMIT 1;
```

## Troubleshooting

### No Data Appearing
- **Check 1**: Verify pg_stat_statements enabled on PostgreSQL
- **Check 2**: Confirm collector has pg_query_stats enabled
- **Check 3**: Verify TimescaleDB has data: `SELECT COUNT(*) FROM metrics_pg_stats_query;`
- **Check 4**: Check Grafana datasource connection

### Queries Slow/Timeout
- **Issue**: Large time ranges with fine granularity
- **Solution**: Use hourly aggregates (`metrics_pg_stats_query_1h`) for >7 days
- **Alternative**: Reduce number of queries on dashboard

### Missing Optional Fields (PG13+)
- **Issue**: WAL metrics showing NULL
- **Cause**: PostgreSQL version < 13
- **Solution**: Queries handle NULL values gracefully

### Cache Hit Ratio Always Low
- **Check**: Is shared buffer pool sized correctly?
- **Check**: Are there full table scans?
- **Check**: Is working memory adequate?

## Best Practices

### Monitoring Workflow
1. **Baseline**: Establish normal performance (week 1)
2. **Alert**: Set thresholds 20% above baseline
3. **Review**: Weekly dashboard review for trends
4. **Act**: Investigate anomalies within 24 hours
5. **Report**: Monthly performance summaries

### Optimization Priorities
1. **High-frequency + high-time**: Biggest impact queries
2. **High-I/O**: Disk bottleneck queries
3. **Low-cache-hit**: Memory access optimization
4. **Increasing-time**: Performance regression queries

### Dashboard Maintenance
- Review alert thresholds monthly
- Archive performance reports quarterly
- Update dashboard based on workload changes
- Share insights with application teams

## Future Enhancements

### Planned Additions
- EXPLAIN PLAN visualization
- Index recommendations panel
- Query fingerprinting and grouping
- Anomaly detection alerts
- Historical comparison views
- ML-based query optimization suggestions

### Integration Opportunities
- Slack/Teams notifications
- PagerDuty incident creation
- Jira ticket auto-creation for slow queries
- Application performance correlation

## Support & References

- **PostgreSQL pg_stat_statements**: https://postgresql.org/docs/current/pgstatstatements.html
- **TimescaleDB Continuous Aggregates**: https://docs.timescale.com/latest/using-timescaledb/continuous-aggregates/
- **Grafana Dashboard Documentation**: https://grafana.com/docs/grafana/latest/dashboards/
- **Query Optimization Guide**: See adjacent documentation

---

**Version**: 1.0
**Last Updated**: February 20, 2026
**Status**: Production Ready
