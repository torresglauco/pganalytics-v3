-- Migration 015: Connection and Session Metrics Tables
-- Stores detailed connection tracking and session information

BEGIN;

-- Create hypertable for connection statistics
CREATE TABLE IF NOT EXISTS metrics_pg_connections_summary (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    connection_state TEXT NOT NULL,  -- 'active', 'idle', 'idle in transaction', 'disabled', etc.
    connection_count INTEGER NOT NULL,
    max_age_seconds DOUBLE PRECISION,
    min_age_seconds DOUBLE PRECISION,
    PRIMARY KEY (time, collector_id, database_name, connection_state)
);

SELECT create_hypertable('metrics_pg_connections_summary', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for long-running transactions
CREATE TABLE IF NOT EXISTS metrics_pg_long_running_transactions (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    pid INTEGER NOT NULL,
    username TEXT NOT NULL,
    session_state TEXT,
    query TEXT,
    query_start TIMESTAMPTZ,
    duration_seconds DOUBLE PRECISION,
    application_name TEXT,
    client_address INET,
    PRIMARY KEY (time, collector_id, database_name, pid)
);

SELECT create_hypertable('metrics_pg_long_running_transactions', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for idle transactions
CREATE TABLE IF NOT EXISTS metrics_pg_idle_transactions (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    pid INTEGER NOT NULL,
    username TEXT NOT NULL,
    query_start TIMESTAMPTZ,
    state_change TIMESTAMPTZ,
    idle_time_seconds DOUBLE PRECISION,
    application_name TEXT,
    client_address INET,
    PRIMARY KEY (time, collector_id, database_name, pid)
);

SELECT create_hypertable('metrics_pg_idle_transactions', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_connections_summary_collector_db ON metrics_pg_connections_summary (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_connections_summary_state ON metrics_pg_connections_summary (connection_state, time DESC);
CREATE INDEX IF NOT EXISTS idx_long_running_collector_db ON metrics_pg_long_running_transactions (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_long_running_duration ON metrics_pg_long_running_transactions (duration_seconds DESC, time DESC);
CREATE INDEX IF NOT EXISTS idx_idle_transactions_collector_db ON metrics_pg_idle_transactions (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_idle_transactions_idle_time ON metrics_pg_idle_transactions (idle_time_seconds DESC, time DESC);

-- Set retention policy (keep for 30 days - short-lived data)
SELECT add_retention_policy('metrics_pg_connections_summary', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_long_running_transactions', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_idle_transactions', INTERVAL '30 days', if_not_exists => TRUE);

COMMIT;
