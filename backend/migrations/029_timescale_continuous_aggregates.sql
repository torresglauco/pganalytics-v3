-- pgAnalytics v3.0 - TimescaleDB Continuous Aggregates
-- Dashboard pre-computed aggregations for instant loads

-- Ensure TimescaleDB extension (graceful handling)
CREATE EXTENSION IF NOT EXISTS timescaledb;

SET search_path TO metrics, public;

-- ============================================================================
-- Connection Metrics Aggregates (5-minute and 1-hour buckets)
-- ============================================================================

-- 5-minute aggregate for metrics_pg_stats_database
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.db_stats_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    AVG(numbackends) AS avg_backends,
    MAX(numbackends) AS max_backends,
    SUM(xact_commit) AS total_commits,
    SUM(xact_rollback) AS total_rollbacks,
    SUM(blks_read) AS total_blks_read,
    SUM(blks_hit) AS total_blks_hit,
    AVG(database_size) AS avg_db_size,
    COUNT(*) AS sample_count
FROM metrics.metrics_pg_stats_database
GROUP BY bucket, collector_id, database_name
WITH DATA;

-- Refresh policy for 5-minute aggregate
SELECT add_continuous_aggregate_policy('metrics.db_stats_5m',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

-- 1-hour aggregate (from 5-minute aggregate)
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.db_stats_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    collector_id,
    database_name,
    AVG(avg_backends) AS avg_backends,
    MAX(max_backends) AS max_backends,
    SUM(total_commits) AS total_commits,
    SUM(total_rollbacks) AS total_rollbacks,
    SUM(total_blks_read) AS total_blks_read,
    SUM(total_blks_hit) AS total_blks_hit,
    AVG(avg_db_size) AS avg_db_size
FROM metrics.db_stats_5m
GROUP BY bucket, collector_id, database_name
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.db_stats_1h',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- ============================================================================
-- Table Metrics Aggregates
-- ============================================================================

-- 5-minute aggregate for table statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.table_stats_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    table_name,
    AVG(seq_scan) AS avg_seq_scan,
    MAX(seq_scan) AS max_seq_scan,
    AVG(idx_scan) AS avg_idx_scan,
    MAX(idx_scan) AS max_idx_scan,
    AVG(n_live_tup) AS avg_live_tup,
    MAX(n_live_tup) AS max_live_tup,
    AVG(table_size) AS avg_table_size,
    MAX(table_size) AS max_table_size,
    COUNT(*) AS sample_count
FROM metrics.metrics_pg_stats_table
GROUP BY bucket, collector_id, database_name, table_name
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.table_stats_5m',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

-- 1-hour aggregate for table statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.table_stats_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    collector_id,
    database_name,
    table_name,
    AVG(avg_seq_scan) AS avg_seq_scan,
    MAX(max_seq_scan) AS max_seq_scan,
    AVG(avg_idx_scan) AS avg_idx_scan,
    MAX(max_idx_scan) AS max_idx_scan,
    AVG(avg_live_tup) AS avg_live_tup,
    MAX(max_live_tup) AS max_live_tup
FROM metrics.table_stats_5m
GROUP BY bucket, collector_id, database_name, table_name
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.table_stats_1h',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- ============================================================================
-- System Statistics Aggregates
-- ============================================================================

-- 5-minute aggregate for system metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.sysstat_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    AVG(cpu_user) AS avg_cpu_user,
    AVG(cpu_system) AS avg_cpu_system,
    AVG(cpu_idle) AS avg_cpu_idle,
    AVG(cpu_iowait) AS avg_cpu_iowait,
    AVG(load_1m) AS avg_load_1m,
    AVG(load_5m) AS avg_load_5m,
    AVG(load_15m) AS avg_load_15m,
    AVG(memory_used) AS avg_memory_used,
    MAX(memory_used) AS max_memory_used,
    AVG(memory_cached) AS avg_memory_cached,
    COUNT(*) AS sample_count
FROM metrics.metrics_sysstat
GROUP BY bucket, collector_id
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.sysstat_5m',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

-- ============================================================================
-- Index for aggregate queries
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_db_stats_5m_collector
    ON metrics.db_stats_5m (collector_id, bucket DESC);

CREATE INDEX IF NOT EXISTS idx_db_stats_1h_collector
    ON metrics.db_stats_1h (collector_id, bucket DESC);

CREATE INDEX IF NOT EXISTS idx_table_stats_5m_collector
    ON metrics.table_stats_5m (collector_id, bucket DESC);

CREATE INDEX IF NOT EXISTS idx_sysstat_5m_collector
    ON metrics.sysstat_5m (collector_id, bucket DESC);

-- ============================================================================
-- Record migration
-- ============================================================================

INSERT INTO pganalytics.schema_versions (version, description)
VALUES ('029_timescale_continuous_aggregates.sql', 'Dashboard continuous aggregates for instant loads')
ON CONFLICT DO NOTHING;