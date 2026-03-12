-- Migration 016: Extension Metrics Tables
-- Stores database extensions and modules information

BEGIN;

-- Create hypertable for extension inventory
CREATE TABLE IF NOT EXISTS metrics_pg_extensions (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    extension_name TEXT NOT NULL,
    extension_version TEXT NOT NULL,
    extension_owner TEXT,
    extension_schema TEXT,
    is_relocatable BOOLEAN,
    description TEXT,
    PRIMARY KEY (time, collector_id, database_name, extension_name)
);

SELECT create_hypertable('metrics_pg_extensions', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_extensions_collector_db ON metrics_pg_extensions (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_extensions_name ON metrics_pg_extensions (extension_name, time DESC);

-- Set retention policy (keep for 90 days - extensions change infrequently)
SELECT add_retention_policy('metrics_pg_extensions', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;
