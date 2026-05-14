-- Migration 033: Host Health Scores Table
-- Stores calculated health scores for hosts based on resource utilization
-- Includes TimescaleDB hypertable for time-series queries
-- Requirement: HOST-04

BEGIN;

-- ============================================================================
-- HOST HEALTH SCORES TABLE
-- ============================================================================

-- Create table for host health scores (0-100 with status labels)
CREATE TABLE IF NOT EXISTS metrics_host_health_scores (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    health_score INT,             -- 0-100
    status VARCHAR(20),           -- healthy, degraded, warning, critical
    cpu_score FLOAT,
    memory_score FLOAT,
    disk_score FLOAT,
    load_score FLOAT,
    calculation_details JSONB,    -- Factors that contributed to score
    PRIMARY KEY (time, collector_id)
);

-- Create TimescaleDB hypertable for health scores
SELECT create_hypertable('metrics_host_health_scores', 'time',
    if_not_exists => TRUE, migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_health_scores_collector
    ON metrics_host_health_scores (collector_id, time DESC);

-- ============================================================================
-- RETENTION POLICY
-- ============================================================================

-- Set retention policy for health scores (keep for 90 days)
SELECT add_retention_policy('metrics_host_health_scores', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;