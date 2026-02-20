-- E2E TimescaleDB Schema Initialization
-- Initializes TimescaleDB for metrics storage in E2E tests

-- Create hypertables for different metric types

-- PostgreSQL Statistics Metrics
CREATE TABLE IF NOT EXISTS metrics_pg_stats (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id VARCHAR(255),
    database VARCHAR(255),
    schema VARCHAR(255),
    table_name VARCHAR(255),
    rows BIGINT,
    size_bytes BIGINT,
    last_vacuum TIMESTAMP WITH TIME ZONE,
    last_analyze TIMESTAMP WITH TIME ZONE,
    seq_scan BIGINT,
    idx_scan BIGINT
);

SELECT create_hypertable('metrics_pg_stats', 'time', if_not_exists => TRUE);

-- Add compression policy (compress data older than 7 days)
SELECT add_compress_policy('metrics_pg_stats', INTERVAL '7 days', if_not_exists => TRUE);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_collector_time
    ON metrics_pg_stats (collector_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_pg_stats_database_time
    ON metrics_pg_stats (database, time DESC);

-- PostgreSQL Log Metrics
CREATE TABLE IF NOT EXISTS metrics_pg_log (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id VARCHAR(255),
    database VARCHAR(255),
    log_level VARCHAR(50),
    message TEXT,
    duration_ms INTEGER,
    query TEXT
);

SELECT create_hypertable('metrics_pg_log', 'time', if_not_exists => TRUE);
SELECT add_compress_policy('metrics_pg_log', INTERVAL '7 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_pg_log_collector_time
    ON metrics_pg_log (collector_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_pg_log_level_time
    ON metrics_pg_log (log_level, time DESC);

-- System Statistics Metrics
CREATE TABLE IF NOT EXISTS metrics_sysstat (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id VARCHAR(255),
    cpu_user NUMERIC,
    cpu_system NUMERIC,
    cpu_idle NUMERIC,
    cpu_load_1m NUMERIC,
    cpu_load_5m NUMERIC,
    cpu_load_15m NUMERIC,
    mem_total_mb BIGINT,
    mem_used_mb BIGINT,
    mem_cached_mb BIGINT,
    mem_free_mb BIGINT
);

SELECT create_hypertable('metrics_sysstat', 'time', if_not_exists => TRUE);
SELECT add_compress_policy('metrics_sysstat', INTERVAL '7 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_sysstat_collector_time
    ON metrics_sysstat (collector_id, time DESC);

-- Disk Usage Metrics
CREATE TABLE IF NOT EXISTS metrics_disk_usage (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id VARCHAR(255),
    mount_point VARCHAR(255),
    device VARCHAR(255),
    total_gb BIGINT,
    used_gb BIGINT,
    free_gb BIGINT,
    percent_used NUMERIC
);

SELECT create_hypertable('metrics_disk_usage', 'time', if_not_exists => TRUE);
SELECT add_compress_policy('metrics_disk_usage', INTERVAL '7 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_metrics_disk_usage_collector_time
    ON metrics_disk_usage (collector_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_disk_usage_mount_time
    ON metrics_disk_usage (mount_point, time DESC);

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;

