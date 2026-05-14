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
-- LOGICAL SUBSCRIPTIONS TABLE
-- ============================================================================

-- Create table for logical subscriptions (from pg_stat_subscription)
CREATE TABLE IF NOT EXISTS metrics_logical_subscriptions (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    sub_name VARCHAR(255),
    sub_state VARCHAR(50),             -- ready, syncing, error, disabled
    sub_recv_lsn VARCHAR(50),
    sub_latest_end_lsn VARCHAR(50),
    sub_last_msg_receipt_time TIMESTAMPTZ,
    sub_worker_pid BIGINT,
    PRIMARY KEY (time, collector_id, sub_name)
);

-- Create TimescaleDB hypertable for logical subscriptions
SELECT create_hypertable('metrics_logical_subscriptions', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_logical_subscriptions_collector
    ON metrics_logical_subscriptions (collector_id, time DESC);

-- ============================================================================
-- PUBLICATIONS TABLE
-- ============================================================================

-- Create table for publications (from pg_publication)
CREATE TABLE IF NOT EXISTS metrics_publications (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    pub_name VARCHAR(255),
    pub_owner VARCHAR(255),
    pub_all_tables BOOLEAN,
    pub_insert BOOLEAN,
    pub_update BOOLEAN,
    pub_delete BOOLEAN,
    pub_truncate BOOLEAN,
    PRIMARY KEY (time, collector_id, pub_name)
);

-- Create TimescaleDB hypertable for publications
SELECT create_hypertable('metrics_publications', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_publications_collector
    ON metrics_publications (collector_id, time DESC);

-- ============================================================================
-- WAL RECEIVERS TABLE
-- ============================================================================

-- Create table for WAL receivers (from pg_stat_wal_receiver)
CREATE TABLE IF NOT EXISTS metrics_wal_receivers (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    status VARCHAR(50),               -- streaming, catching up, etc.
    sender_host VARCHAR(255),
    sender_port INT,
    received_lsn VARCHAR(50),
    latest_end_lsn VARCHAR(50),
    slot_name VARCHAR(255),
    conn_info TEXT,
    PRIMARY KEY (time, collector_id)
);

-- Create TimescaleDB hypertable for WAL receivers
SELECT create_hypertable('metrics_wal_receivers', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_wal_receivers_collector
    ON metrics_wal_receivers (collector_id, time DESC);

-- ============================================================================
-- RETENTION POLICIES
-- ============================================================================

-- Set retention policy (keep for 90 days)
SELECT add_retention_policy('metrics_replication_status', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_replication_slots', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_logical_subscriptions', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_publications', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_wal_receivers', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;