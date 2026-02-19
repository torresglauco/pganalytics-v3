-- pgAnalytics v3.0 - TimescaleDB Setup for Time-Series Metrics
-- This migration is for the separate TimescaleDB instance (metrics database)

-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS metrics;
SET search_path TO metrics, public;

-- Create extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ============================================================================
-- TIME-SERIES TABLES (Hypertables)
-- ============================================================================

-- PostgreSQL Statistics - Table-level metrics
CREATE TABLE IF NOT EXISTS metrics_pg_stats_table (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    instance_id INTEGER,
    database_id INTEGER,
    database_name TEXT,
    schema_name TEXT,
    table_name TEXT,
    -- Table metrics
    seq_scan BIGINT,
    seq_tup_read BIGINT,
    idx_scan BIGINT,
    idx_tup_fetch BIGINT,
    n_tup_ins BIGINT,
    n_tup_upd BIGINT,
    n_tup_del BIGINT,
    n_live_tup BIGINT,
    n_dead_tup BIGINT,
    n_mod_since_analyze BIGINT,
    last_vacuum TIMESTAMP WITH TIME ZONE,
    last_autovacuum TIMESTAMP WITH TIME ZONE,
    last_analyze TIMESTAMP WITH TIME ZONE,
    last_autoanalyze TIMESTAMP WITH TIME ZONE,
    -- Size metrics (bytes)
    heap_blks_read BIGINT,
    heap_blks_hit BIGINT,
    idx_blks_read BIGINT,
    idx_blks_hit BIGINT,
    toast_blks_read BIGINT,
    toast_blks_hit BIGINT,
    tidx_blks_read BIGINT,
    tidx_blks_hit BIGINT,
    table_size BIGINT,
    total_size BIGINT
);

-- Create hypertable
SELECT create_hypertable('metrics_pg_stats_table', 'time', if_not_exists => TRUE);

-- Add indexes for common queries
CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_table_server_time
    ON metrics_pg_stats_table (server_id, time DESC)
    WHERE server_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_table_db_time
    ON metrics_pg_stats_table (database_id, time DESC)
    WHERE database_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_table_collector
    ON metrics_pg_stats_table (collector_id, time DESC);

-- ============================================================================

-- PostgreSQL Statistics - Index-level metrics
CREATE TABLE IF NOT EXISTS metrics_pg_stats_index (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    instance_id INTEGER,
    database_id INTEGER,
    database_name TEXT,
    schema_name TEXT,
    table_name TEXT,
    index_name TEXT,
    -- Index metrics
    idx_scan BIGINT,
    idx_tup_read BIGINT,
    idx_tup_fetch BIGINT,
    pg_relation_size BIGINT,
    is_unique BOOLEAN,
    is_valid BOOLEAN,
    is_primary BOOLEAN
);

SELECT create_hypertable('metrics_pg_stats_index', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_index_db_time
    ON metrics_pg_stats_index (database_id, time DESC)
    WHERE database_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_index_collector
    ON metrics_pg_stats_index (collector_id, time DESC);

-- ============================================================================

-- PostgreSQL Statistics - Database-level metrics
CREATE TABLE IF NOT EXISTS metrics_pg_stats_database (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    instance_id INTEGER,
    database_id INTEGER,
    database_name TEXT,
    -- Database metrics
    numbackends INTEGER,
    xact_commit BIGINT,
    xact_rollback BIGINT,
    blks_read BIGINT,
    blks_hit BIGINT,
    tup_returned BIGINT,
    tup_fetched BIGINT,
    tup_inserted BIGINT,
    tup_updated BIGINT,
    tup_deleted BIGINT,
    -- Connection metrics
    conflicts BIGINT,
    deadlocks BIGINT,
    temp_files BIGINT,
    temp_bytes BIGINT,
    -- Size
    database_size BIGINT
);

SELECT create_hypertable('metrics_pg_stats_database', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_database_time
    ON metrics_pg_stats_database (server_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_database_collector
    ON metrics_pg_stats_database (collector_id, time DESC);

-- ============================================================================

-- System Statistics - CPU, Memory, I/O
CREATE TABLE IF NOT EXISTS metrics_sysstat (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    -- CPU metrics (percentages)
    cpu_user FLOAT8,
    cpu_system FLOAT8,
    cpu_idle FLOAT8,
    cpu_iowait FLOAT8,
    cpu_steal FLOAT8,
    -- Load average
    load_1m FLOAT8,
    load_5m FLOAT8,
    load_15m FLOAT8,
    -- Memory (bytes)
    memory_total BIGINT,
    memory_used BIGINT,
    memory_cached BIGINT,
    memory_buffers BIGINT,
    memory_free BIGINT,
    -- Swap (bytes)
    swap_total BIGINT,
    swap_used BIGINT,
    swap_free BIGINT,
    -- I/O operations per second
    io_read_iops FLOAT8,
    io_write_iops FLOAT8,
    io_read_mb_s FLOAT8,
    io_write_mb_s FLOAT8,
    -- Context switches per second
    context_switches FLOAT8
);

SELECT create_hypertable('metrics_sysstat', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_sysstat_server_time
    ON metrics_sysstat (server_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_sysstat_collector
    ON metrics_sysstat (collector_id, time DESC);

-- ============================================================================

-- Disk Usage Metrics
CREATE TABLE IF NOT EXISTS metrics_disk_usage (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    mountpoint TEXT,
    device TEXT,
    filesystem TEXT,
    -- Size metrics (bytes/GB)
    total_bytes BIGINT,
    used_bytes BIGINT,
    free_bytes BIGINT,
    available_bytes BIGINT,
    -- Percentages
    used_percent FLOAT8,
    inode_total BIGINT,
    inode_used BIGINT,
    inode_free BIGINT
);

SELECT create_hypertable('metrics_disk_usage', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_disk_usage_server_time
    ON metrics_disk_usage (server_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_disk_usage_collector
    ON metrics_disk_usage (collector_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_disk_usage_mount
    ON metrics_disk_usage (server_id, mountpoint, time DESC);

-- ============================================================================

-- PostgreSQL Logs (structured)
CREATE TABLE IF NOT EXISTS metrics_pg_log (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    instance_id INTEGER,
    database_id INTEGER,
    database_name TEXT,
    user_name TEXT,
    -- Log content
    severity TEXT,
    message TEXT,
    detail TEXT,
    hint TEXT,
    context TEXT,
    query TEXT,
    -- SQL state (5-char error code)
    sqlstate TEXT,
    -- Duration for queries
    duration_ms FLOAT8,
    -- Checkpoint info
    checkpoint_redo_blocks BIGINT,
    checkpoint_buffer_dirty BIGINT,
    checkpoint_written BIGINT,
    checkpoint_duration_ms FLOAT8
);

SELECT create_hypertable('metrics_pg_log', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_log_database_time
    ON metrics_pg_log (database_id, time DESC)
    WHERE database_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_metrics_pg_log_severity_time
    ON metrics_pg_log (severity, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_log_collector
    ON metrics_pg_log (collector_id, time DESC);

-- ============================================================================

-- Replication Metrics
CREATE TABLE IF NOT EXISTS metrics_replication (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id UUID NOT NULL,
    server_id INTEGER,
    instance_id INTEGER,
    standby_pid INTEGER,
    usename TEXT,
    application_name TEXT,
    client_addr INET,
    state TEXT,
    sync_state TEXT,
    sync_priority INTEGER,
    -- LSN positions (as text for compatibility)
    write_lsn TEXT,
    flush_lsn TEXT,
    replay_lsn TEXT,
    -- Lag metrics
    write_lag_bytes BIGINT,
    flush_lag_bytes BIGINT,
    replay_lag_bytes BIGINT,
    reply_time TIMESTAMP WITH TIME ZONE
);

SELECT create_hypertable('metrics_replication', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_replication_server_time
    ON metrics_replication (server_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_replication_collector
    ON metrics_replication (collector_id, time DESC);

-- ============================================================================
-- CONTINUOUS AGGREGATES (for fast rollups) - Optional, for high-volume scenarios
-- ============================================================================

-- Hourly aggregates for table metrics (useful for dashboards)
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_pg_stats_table_1h AS
SELECT
    time_bucket('1 hour', time) as time,
    collector_id,
    server_id,
    instance_id,
    database_id,
    database_name,
    schema_name,
    table_name,
    avg(seq_scan) as avg_seq_scan,
    max(seq_scan) as max_seq_scan,
    avg(idx_scan) as avg_idx_scan,
    max(idx_scan) as max_idx_scan,
    avg(n_live_tup) as avg_live_tuples,
    max(n_live_tup) as max_live_tuples,
    avg(table_size) as avg_table_size,
    max(table_size) as max_table_size
FROM metrics_pg_stats_table
WHERE time > now() - INTERVAL '30 days'
GROUP BY 1, 2, 3, 4, 5, 6, 7, 8;

CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_table_1h_time
    ON metrics_pg_stats_table_1h (server_id, time DESC);

-- ============================================================================
-- RETENTION POLICIES
-- ============================================================================

-- Drop existing policies (if any) to avoid errors
DO $$
BEGIN
    EXECUTE 'SELECT drop_chunks(interval ''7 days'', ''metrics_pg_stats_table'')';
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Set retention to 7 days for high-resolution tables
SELECT add_retention_policy('metrics_pg_stats_table', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_stats_index', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_stats_database', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_sysstat', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_disk_usage', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_log', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_replication', INTERVAL '7 days', if_not_exists => TRUE);

-- ============================================================================
-- GRANTS
-- ============================================================================

GRANT USAGE ON SCHEMA metrics TO pganalytics_app_master, pganalytics_app_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA metrics TO pganalytics_app_master;
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA metrics TO pganalytics_app_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA metrics GRANT ALL ON TABLES TO pganalytics_app_master;
ALTER DEFAULT PRIVILEGES IN SCHEMA metrics GRANT SELECT, INSERT ON TABLES TO pganalytics_app_user;
