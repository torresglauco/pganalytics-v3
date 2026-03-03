-- Migration 013: Bloat Metrics Tables
-- Stores table and index bloat analysis data

BEGIN;

-- Create hypertable for table bloat metrics
CREATE TABLE IF NOT EXISTS metrics_pg_bloat_tables (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    dead_tuples BIGINT NOT NULL,
    live_tuples BIGINT NOT NULL,
    dead_ratio_percent DOUBLE PRECISION,  -- Percentage of dead tuples
    table_size TEXT,  -- Human-readable size
    space_wasted_percent DOUBLE PRECISION,  -- Estimated wasted space percentage
    last_vacuum TIMESTAMPTZ,
    last_autovacuum TIMESTAMPTZ,
    vacuum_count BIGINT,
    autovacuum_count BIGINT,
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name)
);

SELECT create_hypertable('metrics_pg_bloat_tables', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for index bloat metrics
CREATE TABLE IF NOT EXISTS metrics_pg_bloat_indexes (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    index_name TEXT NOT NULL,
    index_scans BIGINT,
    tuples_read BIGINT,
    tuples_fetched BIGINT,
    index_size TEXT,  -- Human-readable size
    usage_status TEXT,  -- 'UNUSED', 'RARELY_USED', 'ACTIVE'
    recommendation TEXT,  -- 'CONSIDER_DROPPING', 'IN_USE'
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, index_name)
);

SELECT create_hypertable('metrics_pg_bloat_indexes', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_bloat_tables_collector_db ON metrics_pg_bloat_tables (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_bloat_tables_wasted_space ON metrics_pg_bloat_tables (space_wasted_percent DESC, time DESC);
CREATE INDEX IF NOT EXISTS idx_bloat_indexes_collector_db ON metrics_pg_bloat_indexes (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_bloat_indexes_unused ON metrics_pg_bloat_indexes (recommendation, time DESC);

-- Set retention policy (keep for 90 days - useful for trend analysis)
SELECT add_retention_policy('metrics_pg_bloat_tables', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_bloat_indexes', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;
