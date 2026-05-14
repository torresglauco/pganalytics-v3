-- Migration 031: Replication Metrics Tables
-- Stores streaming replication status and slot information for monitoring
-- Includes TimescaleDB hypertables for time-series queries

BEGIN;

-- ============================================================================
-- REPLICATION STATUS TABLE
-- ============================================================================

-- Create table for streaming replication status (from pg_stat_replication)
CREATE TABLE IF NOT EXISTS metrics_replication_status (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    server_pid BIGINT,
    usename VARCHAR(255),
    application_name VARCHAR(255),
    state VARCHAR(50),          -- streaming, catchup, etc.
    sync_state VARCHAR(20),     -- sync, async, potential, quorum
    write_lsn VARCHAR(50),
    flush_lsn VARCHAR(50),
    replay_lsn VARCHAR(50),
    write_lag_ms BIGINT,
    flush_lag_ms BIGINT,
    replay_lag_ms BIGINT,
    behind_by_mb BIGINT,
    client_addr VARCHAR(100),
    backend_start TIMESTAMPTZ,
    PRIMARY KEY (time, collector_id, server_pid)
);

-- Create TimescaleDB hypertable for replication status
SELECT create_hypertable('metrics_replication_status', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_replication_status_collector
    ON metrics_replication_status (collector_id, time DESC);

-- ============================================================================
-- REPLICATION SLOTS TABLE
-- ============================================================================

-- Create table for replication slots (from pg_replication_slots)
CREATE TABLE IF NOT EXISTS metrics_replication_slots (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    slot_name VARCHAR(255),
    slot_type VARCHAR(20),      -- physical, logical
    active BOOLEAN,
    restart_lsn VARCHAR(50),
    confirmed_flush_lsn VARCHAR(50),
    wal_retained_mb BIGINT,
    backend_pid BIGINT,
    bytes_retained BIGINT,
    PRIMARY KEY (time, collector_id, slot_name)
);

-- Create TimescaleDB hypertable for replication slots
SELECT create_hypertable('metrics_replication_slots', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_replication_slots_collector
    ON metrics_replication_slots (collector_id, time DESC);

-- ============================================================================
-- RETENTION POLICIES
-- ============================================================================

-- Set retention policy (keep for 90 days)
SELECT add_retention_policy('metrics_replication_status', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_replication_slots', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;