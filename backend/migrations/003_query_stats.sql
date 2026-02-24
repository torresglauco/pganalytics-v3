-- Phase 4: Query Performance Monitoring - Query Statistics Table
-- This migration creates the necessary tables and structures for storing pg_stat_statements data
-- Note: TimescaleDB extension is optional; this version uses regular PostgreSQL tables

-- Ensure schema_versions table exists for migration tracking
CREATE TABLE IF NOT EXISTS schema_versions (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Enable required extensions (timescaledb is optional - will be skipped if not available)
-- CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Create hypertable for query statistics
CREATE TABLE IF NOT EXISTS metrics_pg_stats_query (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    user_name TEXT NOT NULL,
    query_hash BIGINT NOT NULL,
    query_text TEXT NOT NULL,
    calls BIGINT NOT NULL,
    total_time FLOAT8 NOT NULL,              -- milliseconds
    mean_time FLOAT8 NOT NULL,               -- milliseconds
    min_time FLOAT8 NOT NULL,                -- milliseconds
    max_time FLOAT8 NOT NULL,                -- milliseconds
    stddev_time FLOAT8 NOT NULL,             -- milliseconds
    rows BIGINT NOT NULL,
    shared_blks_hit BIGINT NOT NULL,
    shared_blks_read BIGINT NOT NULL,
    shared_blks_dirtied BIGINT NOT NULL,
    shared_blks_written BIGINT NOT NULL,
    local_blks_hit BIGINT NOT NULL,
    local_blks_read BIGINT NOT NULL,
    local_blks_dirtied BIGINT NOT NULL,
    local_blks_written BIGINT NOT NULL,
    temp_blks_read BIGINT NOT NULL,
    temp_blks_written BIGINT NOT NULL,
    blk_read_time FLOAT8 NOT NULL,           -- milliseconds
    blk_write_time FLOAT8 NOT NULL,          -- milliseconds
    wal_records BIGINT,                      -- PG13+ optional
    wal_fpi BIGINT,                          -- PG13+ optional
    wal_bytes BIGINT,                        -- PG13+ optional
    query_plan_time FLOAT8,                  -- PG13+ optional
    query_exec_time FLOAT8                   -- PG13+ optional
);

-- Create hypertable if not already one (optional - timescaledb specific)
-- SELECT create_hypertable(
--     'metrics_pg_stats_query',
--     'time',
--     if_not_exists => TRUE,
--     chunk_time_interval => INTERVAL '1 day'
-- );

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_query_stats_collector_time
ON metrics_pg_stats_query (collector_id, time DESC)
INCLUDE (query_hash, query_text, max_time, total_time);

CREATE INDEX IF NOT EXISTS idx_query_stats_database_time
ON metrics_pg_stats_query (database_name, time DESC)
INCLUDE (query_hash, max_time);

CREATE INDEX IF NOT EXISTS idx_query_stats_query_hash
ON metrics_pg_stats_query (query_hash, time DESC)
INCLUDE (max_time, mean_time, calls);

CREATE INDEX IF NOT EXISTS idx_query_stats_max_time
ON metrics_pg_stats_query (time DESC, max_time DESC)
WHERE max_time > 1000;  -- Index slow queries

-- Set data retention policy: 30 days (higher than other metrics due to importance)
-- Using regular PostgreSQL cleanup instead of TimescaleDB retention policy
-- SELECT add_retention_policy(
--     'metrics_pg_stats_query',
--     INTERVAL '30 days',
--     if_not_exists => TRUE
-- );

-- Create continuous aggregate for hourly rollups (used by dashboards)
-- Using regular materialized view instead of TimescaleDB continuous aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_pg_stats_query_1h AS
SELECT
    date_trunc('hour', time) AS hour,
    collector_id,
    database_name,
    user_name,
    query_hash,
    MAX(query_text) as query_text,
    SUM(calls) as total_calls,
    AVG(mean_time) as avg_mean_time,
    MAX(max_time) as max_max_time,
    MIN(min_time) as min_min_time,
    SUM(rows) as total_rows,
    AVG(shared_blks_hit) as avg_cache_hits,
    AVG(shared_blks_read) as avg_cache_reads
FROM metrics_pg_stats_query
GROUP BY date_trunc('hour', time), collector_id, database_name, user_name, query_hash;

-- Create index on continuous aggregate
CREATE INDEX IF NOT EXISTS idx_query_stats_1h_collector_time
ON metrics_pg_stats_query_1h (collector_id, hour DESC);

-- Refresh policy for continuous aggregate (optional - timescaledb specific)
-- SELECT add_continuous_aggregate_policy(
--     'metrics_pg_stats_query_1h',
--     start_offset => INTERVAL '30 days',
--     end_offset => INTERVAL '2 hours',
--     schedule_interval => INTERVAL '1 hour',
--     if_not_exists => TRUE
-- );

-- Create view for top slow queries (past 24 hours)
CREATE OR REPLACE VIEW v_top_slow_queries_24h AS
SELECT
    collector_id,
    database_name,
    query_hash,
    query_text,
    calls,
    total_time,
    mean_time,
    max_time,
    min_time,
    rows,
    shared_blks_hit,
    shared_blks_read,
    blk_read_time,
    blk_write_time
FROM metrics_pg_stats_query
WHERE time >= NOW() - INTERVAL '24 hours'
ORDER BY max_time DESC
LIMIT 100;

-- Create view for top frequent queries (past 24 hours)
CREATE OR REPLACE VIEW v_top_frequent_queries_24h AS
SELECT
    collector_id,
    database_name,
    query_hash,
    query_text,
    calls,
    total_time,
    mean_time,
    max_time,
    rows
FROM metrics_pg_stats_query
WHERE time >= NOW() - INTERVAL '24 hours'
ORDER BY calls DESC
LIMIT 100;

-- Grant access to monitoring role
GRANT SELECT ON metrics_pg_stats_query TO pg_monitor;
GRANT SELECT ON metrics_pg_stats_query_1h TO pg_monitor;
GRANT SELECT ON v_top_slow_queries_24h TO pg_monitor;
GRANT SELECT ON v_top_frequent_queries_24h TO pg_monitor;

-- Create system catalog table for query metadata (for auditing)
CREATE TABLE IF NOT EXISTS query_metadata (
    query_hash BIGINT PRIMARY KEY,
    query_text TEXT NOT NULL,
    first_seen TIMESTAMPTZ DEFAULT NOW(),
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    is_slow BOOLEAN DEFAULT FALSE,
    notes TEXT
);

GRANT SELECT ON query_metadata TO pg_monitor;

-- Record migration completion
INSERT INTO schema_versions (version, description, executed_at)
VALUES ('003_query_stats', 'Add query statistics hypertable and views', NOW())
ON CONFLICT (version) DO NOTHING;
