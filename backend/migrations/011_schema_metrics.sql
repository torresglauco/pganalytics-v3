-- Migration 011: Schema Metrics Tables
-- Stores detailed schema information for all monitored databases
-- Includes tables, columns, constraints, foreign keys, indexes, and triggers

BEGIN;

-- Create hypertable for table schema information
CREATE TABLE IF NOT EXISTS metrics_pg_schema_tables (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    table_type TEXT NOT NULL,  -- 'BASE TABLE', 'VIEW', etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name)
);

SELECT create_hypertable('metrics_pg_schema_tables', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for column information
CREATE TABLE IF NOT EXISTS metrics_pg_schema_columns (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    is_nullable BOOLEAN NOT NULL,
    column_default TEXT,
    ordinal_position INTEGER NOT NULL,
    character_max_length INTEGER,
    numeric_precision INTEGER,
    numeric_scale INTEGER,
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, column_name)
);

SELECT create_hypertable('metrics_pg_schema_columns', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for constraints
CREATE TABLE IF NOT EXISTS metrics_pg_schema_constraints (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    constraint_name TEXT NOT NULL,
    constraint_type TEXT NOT NULL,  -- PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK
    columns TEXT,  -- Comma-separated column names
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, constraint_name)
);

SELECT create_hypertable('metrics_pg_schema_constraints', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create hypertable for foreign keys
CREATE TABLE IF NOT EXISTS metrics_pg_schema_foreign_keys (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL,
    database_name TEXT NOT NULL,
    source_schema TEXT NOT NULL,
    source_table TEXT NOT NULL,
    source_column TEXT NOT NULL,
    target_schema TEXT NOT NULL,
    target_table TEXT NOT NULL,
    target_column TEXT NOT NULL,
    update_rule TEXT NOT NULL,
    delete_rule TEXT NOT NULL,
    PRIMARY KEY (time, collector_id, database_name, source_schema, source_table, source_column)
);

SELECT create_hypertable('metrics_pg_schema_foreign_keys', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_schema_tables_collector_db ON metrics_pg_schema_tables (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_schema_columns_collector_db ON metrics_pg_schema_columns (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_schema_constraints_collector_db ON metrics_pg_schema_constraints (collector_id, database_name, time DESC);
CREATE INDEX IF NOT EXISTS idx_schema_fk_collector_db ON metrics_pg_schema_foreign_keys (collector_id, database_name, time DESC);

-- Set retention policy (keep for 90 days)
SELECT add_retention_policy('metrics_pg_schema_tables', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_schema_columns', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_schema_constraints', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_pg_schema_foreign_keys', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;
