-- PostgreSQL Logs Schema Migration
-- Stores PostgreSQL database logs for monitoring and alerting
-- This migration creates tables for collecting and analyzing PostgreSQL logs

SET search_path TO pganalytics, public;

-- ============================================================================
-- PostgreSQL Logs Table (19 columns)
-- Stores complete PostgreSQL log entries with structured fields
-- ============================================================================

CREATE TABLE IF NOT EXISTS postgresql_logs (
    id BIGSERIAL PRIMARY KEY,

    -- Core identifiers and relationships
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE NOT NULL,
    instance_id INTEGER REFERENCES postgresql_instances(id) ON DELETE CASCADE NOT NULL,
    database_id INTEGER REFERENCES databases(id) ON DELETE SET NULL,

    -- Log timestamp and classification
    log_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    log_level VARCHAR(20) NOT NULL,  -- DEBUG, INFO, NOTICE, WARNING, ERROR, FATAL, PANIC
    log_message TEXT NOT NULL,

    -- Source location in PostgreSQL code
    source_location VARCHAR(255),  -- file:line or function
    process_id INTEGER,

    -- Query-related fields
    query_text TEXT,  -- NULL for non-query logs
    query_hash BIGINT,  -- For grouping similar queries

    -- Error and exception details
    error_code VARCHAR(5),  -- PostgreSQL error code (e.g., '42P01')
    error_detail TEXT,
    error_hint TEXT,
    error_context TEXT,

    -- User and connection information
    user_name VARCHAR(255),
    connection_from VARCHAR(255),  -- IP address or socket
    session_id VARCHAR(255),

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_postgresql_logs_collector_timestamp
    ON postgresql_logs(collector_id, log_timestamp DESC)
    WHERE log_timestamp > CURRENT_TIMESTAMP - INTERVAL '7 days';

CREATE INDEX IF NOT EXISTS idx_postgresql_logs_level_timestamp
    ON postgresql_logs(log_level, log_timestamp DESC)
    WHERE log_level IN ('ERROR', 'FATAL', 'PANIC');

CREATE INDEX IF NOT EXISTS idx_postgresql_logs_instance_timestamp
    ON postgresql_logs(instance_id, log_timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_postgresql_logs_database_timestamp
    ON postgresql_logs(database_id, log_timestamp DESC)
    WHERE database_id IS NOT NULL;

-- ============================================================================
-- Log Events Hourly Table (Aggregated data)
-- Hourly aggregation for efficient time-series analysis
-- ============================================================================

CREATE TABLE IF NOT EXISTS log_events_hourly (
    id BIGSERIAL PRIMARY KEY,

    -- Time bucket (rounded to hour)
    hour_bucket TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Dimensions for grouping
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE NOT NULL,
    instance_id INTEGER REFERENCES postgresql_instances(id) ON DELETE CASCADE NOT NULL,
    database_id INTEGER REFERENCES databases(id) ON DELETE SET NULL,
    log_level VARCHAR(20) NOT NULL,

    -- Aggregated metrics
    event_count INTEGER NOT NULL DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    unique_sessions INTEGER DEFAULT 0,

    -- Statistics
    error_count INTEGER DEFAULT 0,
    warning_count INTEGER DEFAULT 0,
    fatal_count INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT log_events_hourly_unique UNIQUE(hour_bucket, collector_id, instance_id, database_id, log_level)
);

CREATE INDEX IF NOT EXISTS idx_log_events_hourly_bucket_level
    ON log_events_hourly(hour_bucket DESC, log_level);

CREATE INDEX IF NOT EXISTS idx_log_events_hourly_instance
    ON log_events_hourly(instance_id, hour_bucket DESC);

-- ============================================================================
-- Log Statistics Hourly View
-- Provides high-level statistics for dashboard and monitoring
-- ============================================================================

CREATE OR REPLACE VIEW log_stats_hourly AS
SELECT
    hour_bucket,
    collector_id,
    instance_id,
    database_id,
    SUM(event_count) as total_events,
    SUM(error_count) as total_errors,
    SUM(warning_count) as total_warnings,
    SUM(fatal_count) as total_fatals,
    SUM(unique_users) as total_unique_users,
    SUM(unique_sessions) as total_unique_sessions,
    COUNT(DISTINCT log_level) as log_level_variety
FROM log_events_hourly
GROUP BY hour_bucket, collector_id, instance_id, database_id
ORDER BY hour_bucket DESC, collector_id, instance_id;

-- ============================================================================
-- Comments for documentation
-- ============================================================================

COMMENT ON TABLE postgresql_logs IS 'Stores PostgreSQL database log entries for monitoring and alerting';
COMMENT ON COLUMN postgresql_logs.log_level IS 'PostgreSQL log level: DEBUG, INFO, NOTICE, WARNING, ERROR, FATAL, PANIC';
COMMENT ON COLUMN postgresql_logs.error_code IS 'PostgreSQL error code for classification and alerting';
COMMENT ON TABLE log_events_hourly IS 'Hourly aggregated log event metrics for efficient time-series analysis';
COMMENT ON VIEW log_stats_hourly IS 'High-level statistics view for dashboard and monitoring';
