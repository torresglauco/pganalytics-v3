-- Migration 012: Lock Metrics Tables
-- Stores database lock information, wait chains, and blocking queries

BEGIN;

-- Create hypertable for active locks
CREATE TABLE IF NOT EXISTS metrics_pg_locks (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    pid INTEGER NOT NULL,
    locktype TEXT NOT NULL,  -- 'relation', 'extend', 'page', 'tuple', 'virtualxid', 'transactionid', etc.
    mode TEXT NOT NULL,  -- AccessShare, RowShare, RowExclusive, etc.
    granted BOOLEAN NOT NULL,
    relation_id INTEGER,
    page_number INTEGER,
    tuple_id INTEGER,
    username TEXT,
    session_state TEXT,  -- idle, active, etc.
    lock_age_seconds DOUBLE PRECISION,
    query TEXT,
    PRIMARY KEY (time, collector_id, database_name, pid, locktype)
);

SELECT create_hypertable('metrics_pg_locks', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for lock wait chains
CREATE TABLE IF NOT EXISTS metrics_pg_lock_waits (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    blocked_pid INTEGER NOT NULL,
    blocking_pid INTEGER NOT NULL,
    blocked_username TEXT,
    blocking_username TEXT,
    blocked_query TEXT,
    blocking_query TEXT,
    wait_time_seconds DOUBLE PRECISION,
    blocked_application TEXT,
    blocking_application TEXT,
    PRIMARY KEY (time, collector_id, database_name, blocked_pid, blocking_pid)
);

SELECT create_hypertable('metrics_pg_lock_waits', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for blocking queries summary
CREATE TABLE IF NOT EXISTS metrics_pg_blocking_queries (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    pid INTEGER NOT NULL,
    username TEXT NOT NULL,
    state TEXT NOT NULL,
    query TEXT,
    query_start TIMESTAMPTZ,
    duration_seconds DOUBLE PRECISION,
    application_name TEXT,
    client_address INET,
    PRIMARY KEY (time, collector_id, database_name, pid)
);

SELECT create_hypertable('metrics_pg_blocking_queries', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_locks_collector_db_time ON metrics_pg_locks (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_locks_pid ON metrics_pg_locks (pid, time DESC);
CREATE INDEX IF NOT EXISTS idx_lock_waits_collector_db ON metrics_pg_lock_waits (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_blocking_queries_collector_db ON metrics_pg_blocking_queries (collector_id, database_name, time DESC);

-- Set retention policy (keep for 30 days - lock info is short-lived)
SELECT add_retention_policy('metrics_pg_locks', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_lock_waits', INTERVAL '30 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_blocking_queries', INTERVAL '30 days', if_not_exists => TRUE);

COMMIT;
