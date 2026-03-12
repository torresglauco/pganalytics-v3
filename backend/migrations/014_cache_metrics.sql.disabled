-- Migration 014: Cache Hit Ratio Metrics Tables
-- Stores buffer pool and cache performance metrics

BEGIN;

-- Create hypertable for table cache hit ratios
CREATE TABLE IF NOT EXISTS metrics_pg_cache_tables (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    heap_blks_hit BIGINT NOT NULL,
    heap_blks_read BIGINT NOT NULL,
    heap_cache_hit_ratio DOUBLE PRECISION,  -- Percentage
    idx_blks_hit BIGINT,
    idx_blks_read BIGINT,
    idx_cache_hit_ratio DOUBLE PRECISION,  -- Percentage
    toast_blks_hit BIGINT,
    toast_blks_read BIGINT,
    tidx_blks_hit BIGINT,
    tidx_blks_read BIGINT,
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name)
);

SELECT create_hypertable('metrics_pg_cache_tables', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for index cache hit ratios
CREATE TABLE IF NOT EXISTS metrics_pg_cache_indexes (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    index_name TEXT NOT NULL,
    blks_hit BIGINT NOT NULL,
    blks_read BIGINT NOT NULL,
    cache_hit_ratio DOUBLE PRECISION,  -- Percentage
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, index_name)
);

SELECT create_hypertable('metrics_pg_cache_indexes', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_cache_tables_collector_db ON metrics_pg_cache_tables (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_cache_tables_hit_ratio ON metrics_pg_cache_tables (heap_cache_hit_ratio, time DESC);
CREATE INDEX IF NOT EXISTS idx_cache_indexes_collector_db ON metrics_pg_cache_indexes (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_cache_indexes_hit_ratio ON metrics_pg_cache_indexes (cache_hit_ratio, time DESC);

-- Set retention policy (keep for 90 days)
SELECT add_retention_policy('metrics_pg_cache_tables', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_cache_indexes', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;
