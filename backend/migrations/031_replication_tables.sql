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

-- ============================================================================
-- TABLE INVENTORY (INV-01)
-- ============================================================================

-- Create table for database table inventory with sizes and row counts
CREATE TABLE IF NOT EXISTS metrics_table_inventory (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    table_type VARCHAR(50),        -- BASE TABLE, VIEW
    row_count BIGINT,
    total_size_mb BIGINT,
    table_size_mb BIGINT,
    index_size_mb BIGINT,
    toast_size_mb BIGINT,
    has_oids BOOLEAN,
    table_oid BIGINT,
    PRIMARY KEY (time, collector_id, table_oid)
);

-- Create TimescaleDB hypertable for table inventory
SELECT create_hypertable('metrics_table_inventory', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_table_inventory_collector
    ON metrics_table_inventory (collector_id, time DESC);

-- Create index for filtering by database and schema
CREATE INDEX IF NOT EXISTS idx_table_inventory_db_schema
    ON metrics_table_inventory (collector_id, database_name, schema_name, time DESC);

-- ============================================================================
-- COLUMN INVENTORY (INV-02)
-- ============================================================================

-- Create table for database column inventory with types
CREATE TABLE IF NOT EXISTS metrics_column_inventory (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    column_name VARCHAR(255),
    data_type VARCHAR(100),
    is_nullable BOOLEAN,
    column_default TEXT,
    ordinal_position INT,
    character_max_length INT,
    numeric_precision INT,
    numeric_scale INT,
    is_primary_key BOOLEAN,
    is_foreign_key BOOLEAN,
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, column_name)
);

-- Create TimescaleDB hypertable for column inventory
SELECT create_hypertable('metrics_column_inventory', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_column_inventory_collector
    ON metrics_column_inventory (collector_id, time DESC);

-- Create index for filtering by database and table
CREATE INDEX IF NOT EXISTS idx_column_inventory_db_table
    ON metrics_column_inventory (collector_id, database_name, table_name, time DESC);

-- ============================================================================
-- INDEX INVENTORY (INV-03)
-- ============================================================================

-- Create table for database index inventory with usage statistics
CREATE TABLE IF NOT EXISTS metrics_index_inventory (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    index_name VARCHAR(255),
    index_definition TEXT,
    index_size_mb BIGINT,
    idx_scan BIGINT,
    idx_tup_read BIGINT,
    idx_tup_fetch BIGINT,
    usage_status VARCHAR(50),      -- UNUSED, RARELY_USED, ACTIVE
    is_primary BOOLEAN,
    is_unique BOOLEAN,
    index_oid BIGINT,
    PRIMARY KEY (time, collector_id, index_oid)
);

-- Create TimescaleDB hypertable for index inventory
SELECT create_hypertable('metrics_index_inventory', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_index_inventory_collector
    ON metrics_index_inventory (collector_id, time DESC);

-- Create index for filtering by database and table
CREATE INDEX IF NOT EXISTS idx_index_inventory_db_table
    ON metrics_index_inventory (collector_id, database_name, table_name, time DESC);

-- ============================================================================
-- EXTENSION INVENTORY (INV-04)
-- ============================================================================

-- Create table for database extension inventory with versions
CREATE TABLE IF NOT EXISTS metrics_extension_inventory (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    extension_name VARCHAR(255),
    extension_version VARCHAR(100),
    extension_owner VARCHAR(255),
    extension_schema VARCHAR(255),
    is_relocatable BOOLEAN,
    description TEXT,
    PRIMARY KEY (time, collector_id, database_name, extension_name)
);

-- Create TimescaleDB hypertable for extension inventory
SELECT create_hypertable('metrics_extension_inventory', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_extension_inventory_collector
    ON metrics_extension_inventory (collector_id, time DESC);

-- ============================================================================
-- SCHEMA VERSIONS (INV-05)
-- ============================================================================

-- Create table for schema change tracking
CREATE TABLE IF NOT EXISTS metrics_schema_versions (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    version_hash VARCHAR(64),      -- MD5 hash of all object signatures
    change_type VARCHAR(50),       -- INIT, TABLE_ADDED, TABLE_REMOVED, TABLE_MODIFIED, etc.
    object_type VARCHAR(50),       -- TABLE, COLUMN, INDEX, EXTENSION
    object_name VARCHAR(255),
    change_details TEXT,           -- JSON describing the change
    previous_value TEXT,
    new_value TEXT,
    PRIMARY KEY (time, collector_id, database_name, change_type, object_name)
);

-- Create TimescaleDB hypertable for schema versions
SELECT create_hypertable('metrics_schema_versions', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_schema_versions_collector
    ON metrics_schema_versions (collector_id, time DESC);

-- Create index for filtering by database and change type
CREATE INDEX IF NOT EXISTS idx_schema_versions_db_type
    ON metrics_schema_versions (collector_id, database_name, change_type, time DESC);

-- ============================================================================
-- INVENTORY RETENTION POLICIES
-- ============================================================================

-- Set retention policy for inventory tables (keep for 90 days)
SELECT add_retention_policy('metrics_table_inventory', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_column_inventory', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_index_inventory', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_extension_inventory', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_schema_versions', INTERVAL '365 days', if_not_exists => TRUE);  -- Keep schema changes longer

COMMIT;