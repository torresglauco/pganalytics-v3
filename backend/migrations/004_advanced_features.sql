-- Phase 4.4: Advanced Query Performance Features
-- Tables and functions for fingerprinting, EXPLAIN plans, index recommendations, anomaly detection, and snapshots

-- ============================================================================
-- 1. Query Fingerprinting & Normalization
-- ============================================================================

-- Function to normalize SQL query (remove literals, extra whitespace)
CREATE OR REPLACE FUNCTION normalize_query_text(query_text TEXT) RETURNS TEXT AS $$
BEGIN
    -- Remove leading/trailing whitespace
    query_text := TRIM(query_text);

    -- Remove multiple spaces
    query_text := regexp_replace(query_text, '\s+', ' ', 'g');

    -- Remove numeric literals (but keep table.column references)
    query_text := regexp_replace(query_text, '\b\d+\b', '?', 'g');

    -- Remove string literals ('...')
    query_text := regexp_replace(query_text, '''[^'']*''', '''?''', 'g');

    -- Remove hex literals (0x...)
    query_text := regexp_replace(query_text, '\b0x[0-9A-Fa-f]+\b', '?', 'g');

    -- Normalize boolean values
    query_text := regexp_replace(query_text, '\b(true|false)\b', '?', 'gi');

    -- Normalize NULL
    query_text := regexp_replace(query_text, '\bNULL\b', '?', 'gi');

    -- Convert to uppercase for consistency
    query_text := UPPER(query_text);

    RETURN query_text;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to compute fingerprint hash
CREATE OR REPLACE FUNCTION fingerprint_query_hash(normalized_text TEXT) RETURNS BIGINT AS $$
BEGIN
    -- Use MD5 hash converted to bigint for consistent fingerprinting
    RETURN ('x' || SUBSTR(MD5(normalized_text), 1, 16))::BIT(64)::BIGINT;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Table: query_fingerprints
CREATE TABLE IF NOT EXISTS query_fingerprints (
    id BIGSERIAL PRIMARY KEY,
    fingerprint_hash BIGINT NOT NULL,
    normalized_text TEXT NOT NULL,
    sample_query_text TEXT,
    collector_id UUID,
    database_name VARCHAR(63),
    total_calls BIGINT DEFAULT 0,
    avg_execution_time FLOAT DEFAULT 0,
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_collector FOREIGN KEY (collector_id) REFERENCES collectors(id) ON DELETE SET NULL,
    UNIQUE(fingerprint_hash, database_name)
);

CREATE INDEX IF NOT EXISTS idx_query_fingerprint_hash ON query_fingerprints(fingerprint_hash);
CREATE INDEX IF NOT EXISTS idx_query_fingerprint_db ON query_fingerprints(database_name, fingerprint_hash);
CREATE INDEX IF NOT EXISTS idx_query_fingerprint_last_seen ON query_fingerprints(last_seen DESC);

-- ============================================================================
-- 2. EXPLAIN PLAN Integration
-- ============================================================================

-- Table: explain_plans
CREATE TABLE IF NOT EXISTS explain_plans (
    id BIGSERIAL PRIMARY KEY,
    query_hash INT64 NOT NULL,
    query_fingerprint_hash BIGINT,
    collected_at TIMESTAMP DEFAULT NOW(),
    plan_json JSONB NOT NULL,
    plan_text TEXT,
    rows_expected BIGINT,
    rows_actual BIGINT,
    plan_duration_ms FLOAT,
    execution_duration_ms FLOAT,
    has_seq_scan BOOLEAN DEFAULT FALSE,
    has_index_scan BOOLEAN DEFAULT FALSE,
    has_bitmap_scan BOOLEAN DEFAULT FALSE,
    has_nested_loop BOOLEAN DEFAULT FALSE,
    total_buffers_read BIGINT,
    total_buffers_hit BIGINT,
    CONSTRAINT fk_query_hash FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_explain_query_hash ON explain_plans(query_hash, collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_explain_collected_at ON explain_plans(collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_explain_fingerprint ON explain_plans(query_fingerprint_hash, collected_at DESC);

-- Materialized view for latest explain plans
CREATE MATERIALIZED VIEW IF NOT EXISTS v_latest_explain_plans AS
SELECT DISTINCT ON (query_hash)
    query_hash,
    query_fingerprint_hash,
    collected_at,
    plan_json,
    has_seq_scan,
    has_index_scan,
    rows_expected,
    rows_actual,
    plan_duration_ms
FROM explain_plans
ORDER BY query_hash, collected_at DESC;

CREATE INDEX IF NOT EXISTS idx_v_latest_explain_query ON v_latest_explain_plans(query_hash);

-- ============================================================================
-- 3. Index Recommendations
-- ============================================================================

-- Table: index_recommendations
CREATE TABLE IF NOT EXISTS index_recommendations (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID,
    database_name VARCHAR(63),
    schema_name VARCHAR(63) DEFAULT 'public',
    table_name VARCHAR(63),
    column_names TEXT[] NOT NULL,
    column_names_str VARCHAR(255), -- Denormalized for indexing
    create_statement TEXT NOT NULL,
    estimated_improvement_percent FLOAT DEFAULT 0,
    affected_query_count BIGINT DEFAULT 0,
    affected_total_time_ms FLOAT DEFAULT 0,
    frequency_score FLOAT DEFAULT 0, -- How often Seq Scan occurs (0-1)
    impact_score FLOAT DEFAULT 0, -- Estimated time savings (0-1)
    confidence_score FLOAT DEFAULT 0.8, -- Overall recommendation confidence
    dismissed BOOLEAN DEFAULT FALSE,
    dismissed_at TIMESTAMP,
    dismissed_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_collector_idx FOREIGN KEY (collector_id) REFERENCES collectors(id) ON DELETE SET NULL,
    UNIQUE(database_name, schema_name, table_name, column_names_str)
);

CREATE INDEX IF NOT EXISTS idx_recommendations_db ON index_recommendations(database_name, dismissed);
CREATE INDEX IF NOT EXISTS idx_recommendations_confidence ON index_recommendations(confidence_score DESC);
CREATE INDEX IF NOT EXISTS idx_recommendations_active ON index_recommendations(dismissed, created_at DESC) WHERE NOT dismissed;

-- Function to analyze EXPLAIN plans for index recommendations
CREATE OR REPLACE FUNCTION analyze_explain_for_indexes() RETURNS TABLE(
    database_name VARCHAR,
    table_name VARCHAR,
    columns TEXT[],
    create_statement TEXT,
    frequency BIGINT,
    impact_estimate FLOAT
) AS $$
BEGIN
    -- Extract Seq Scan nodes from recent EXPLAIN plans
    -- This function will be executed periodically to generate recommendations

    -- Implementation: Parse EXPLAIN JSON to find:
    -- 1. Seq Scan nodes (missing indexes)
    -- 2. Filter conditions being applied at scan time
    -- 3. Frequency of this pattern
    -- 4. Estimated improvement from index

    RETURN QUERY
    SELECT
        'to_be_populated'::VARCHAR,
        'to_be_populated'::VARCHAR,
        ARRAY[]::TEXT[],
        'to_be_populated'::TEXT,
        0::BIGINT,
        0.0::FLOAT
    WHERE FALSE; -- Placeholder - actual implementation populated during Phase 4.4.3
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 4. Anomaly Detection
-- ============================================================================

-- Table: query_anomalies
CREATE TABLE IF NOT EXISTS query_anomalies (
    id BIGSERIAL PRIMARY KEY,
    query_hash INT64 NOT NULL,
    query_fingerprint_hash BIGINT,
    anomaly_type VARCHAR(50) NOT NULL, -- 'execution_time_spike', 'cache_degradation', 'io_increase', 'row_count_anomaly'
    severity VARCHAR(20) NOT NULL DEFAULT 'low', -- 'low', 'medium', 'high'
    detected_at TIMESTAMP DEFAULT NOW(),
    metric_name VARCHAR(100), -- Which metric triggered (execution_time, cache_ratio, etc.)
    metric_value FLOAT,
    baseline_value FLOAT,
    deviation_stddev FLOAT,
    z_score FLOAT, -- Standard deviations from mean
    raw_metrics_json JSONB,
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,
    CONSTRAINT fk_query_anomaly FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_anomaly_query ON query_anomalies(query_hash, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_anomaly_severity ON query_anomalies(severity, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_anomaly_resolved ON query_anomalies(resolved, detected_at DESC) WHERE NOT resolved;
CREATE INDEX IF NOT EXISTS idx_anomaly_type ON query_anomalies(anomaly_type, detected_at DESC);

-- Table: query_baselines (for anomaly detection)
CREATE TABLE IF NOT EXISTS query_baselines (
    id BIGSERIAL PRIMARY KEY,
    query_hash INT64 NOT NULL,
    metric_name VARCHAR(100) NOT NULL, -- 'mean_time', 'max_time', 'shared_blks_hit', 'blk_read_time', etc.
    baseline_value FLOAT NOT NULL,
    stddev_value FLOAT NOT NULL,
    baseline_period_days INT DEFAULT 7,
    last_updated TIMESTAMP DEFAULT NOW(),
    min_value FLOAT,
    max_value FLOAT,
    CONSTRAINT fk_query_baseline FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE,
    UNIQUE(query_hash, metric_name)
);

CREATE INDEX IF NOT EXISTS idx_baseline_query ON query_baselines(query_hash);

-- Function to calculate baselines and detect anomalies
CREATE OR REPLACE FUNCTION calculate_baselines_and_anomalies() RETURNS VOID AS $$
BEGIN
    -- This function will:
    -- 1. Calculate 7-day rolling baseline for each query metric
    -- 2. Detect anomalies using stddev-based detection (>2x stddev = anomaly)
    -- 3. Insert into query_anomalies table

    -- Implementation executed during Phase 4.4.4
    NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 5. Performance Snapshots for Historical Comparison
-- ============================================================================

-- Table: performance_snapshots
CREATE TABLE IF NOT EXISTS performance_snapshots (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    snapshot_type VARCHAR(50) DEFAULT 'manual', -- 'manual', 'scheduled', 'pre_deploy', 'post_deploy'
    created_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(255),
    metadata_json JSONB,
    UNIQUE(name, created_at)
);

CREATE INDEX IF NOT EXISTS idx_snapshot_created ON performance_snapshots(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snapshot_type ON performance_snapshots(snapshot_type, created_at DESC);

-- Table: query_performance_snapshots
CREATE TABLE IF NOT EXISTS query_performance_snapshots (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL,
    query_hash INT64 NOT NULL,
    query_fingerprint_hash BIGINT,
    database_name VARCHAR(63),
    calls BIGINT,
    total_time FLOAT,
    mean_time FLOAT,
    max_time FLOAT,
    min_time FLOAT,
    stddev_time FLOAT,
    rows BIGINT,
    shared_blks_hit BIGINT,
    shared_blks_read BIGINT,
    blk_read_time FLOAT,
    blk_write_time FLOAT,
    -- Optional PG13+ fields
    wal_records BIGINT,
    wal_fpi BIGINT,
    wal_bytes BIGINT,
    query_plan_time FLOAT,
    query_exec_time FLOAT,
    CONSTRAINT fk_snapshot FOREIGN KEY (snapshot_id) REFERENCES performance_snapshots(id) ON DELETE CASCADE,
    CONSTRAINT fk_snapshot_query FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE,
    UNIQUE(snapshot_id, query_hash)
);

CREATE INDEX IF NOT EXISTS idx_snapshot_data_snapshot ON query_performance_snapshots(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_data_query ON query_performance_snapshots(query_hash, snapshot_id);

-- Function to create a snapshot of current query metrics
CREATE OR REPLACE FUNCTION create_performance_snapshot(
    p_name VARCHAR,
    p_description TEXT,
    p_snapshot_type VARCHAR,
    p_created_by VARCHAR
) RETURNS BIGINT AS $$
DECLARE
    v_snapshot_id BIGINT;
BEGIN
    -- Insert snapshot metadata
    INSERT INTO performance_snapshots (name, description, snapshot_type, created_at, created_by)
    VALUES (p_name, p_description, p_snapshot_type, NOW(), p_created_by)
    RETURNING id INTO v_snapshot_id;

    -- Capture current query metrics from metrics_pg_stats_query
    INSERT INTO query_performance_snapshots (
        snapshot_id, query_hash, query_fingerprint_hash, database_name,
        calls, total_time, mean_time, max_time, min_time, stddev_time,
        rows, shared_blks_hit, shared_blks_read, blk_read_time, blk_write_time,
        wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
    )
    SELECT
        v_snapshot_id, mq.query_hash,
        fingerprint_query_hash(normalize_query_text(mq.query_text)),
        mq.database_name,
        mq.calls, mq.total_time, mq.mean_time, mq.max_time, mq.min_time, mq.stddev_time,
        mq.rows, mq.shared_blks_hit, mq.shared_blks_read, mq.blk_read_time, mq.blk_write_time,
        mq.wal_records, mq.wal_fpi, mq.wal_bytes, mq.query_plan_time, mq.query_exec_time
    FROM (
        SELECT DISTINCT ON (query_hash)
            query_hash, database_name, query_text,
            calls, total_time, mean_time, max_time, min_time, stddev_time,
            rows, shared_blks_hit, shared_blks_read, blk_read_time, blk_write_time,
            wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
        FROM metrics_pg_stats_query
        WHERE time >= NOW() - INTERVAL '24 hours'
        ORDER BY query_hash, time DESC
    ) mq;

    RETURN v_snapshot_id;
END;
$$ LANGUAGE plpgsql;

-- Function to compare two snapshots
CREATE OR REPLACE FUNCTION compare_snapshots(
    p_before_snapshot_id BIGINT,
    p_after_snapshot_id BIGINT
) RETURNS TABLE(
    query_hash INT64,
    database_name VARCHAR,
    before_calls BIGINT,
    after_calls BIGINT,
    calls_change BIGINT,
    calls_change_percent FLOAT,
    before_mean_time FLOAT,
    after_mean_time FLOAT,
    mean_time_change FLOAT,
    mean_time_change_percent FLOAT,
    before_max_time FLOAT,
    after_max_time FLOAT,
    max_time_change FLOAT,
    before_cache_hits BIGINT,
    after_cache_hits BIGINT,
    before_cache_reads BIGINT,
    after_cache_reads BIGINT,
    improvement_status VARCHAR -- 'improved', 'degraded', 'unchanged'
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        b.query_hash,
        b.database_name,
        b.calls AS before_calls,
        a.calls AS after_calls,
        COALESCE(a.calls, 0) - COALESCE(b.calls, 0) AS calls_change,
        CASE
            WHEN b.calls IS NULL OR b.calls = 0 THEN NULL
            ELSE ((COALESCE(a.calls, 0) - COALESCE(b.calls, 0))::FLOAT / b.calls) * 100
        END AS calls_change_percent,
        b.mean_time AS before_mean_time,
        a.mean_time AS after_mean_time,
        COALESCE(a.mean_time, 0) - COALESCE(b.mean_time, 0) AS mean_time_change,
        CASE
            WHEN b.mean_time IS NULL OR b.mean_time = 0 THEN NULL
            ELSE ((COALESCE(a.mean_time, 0) - COALESCE(b.mean_time, 0))::FLOAT / b.mean_time) * 100
        END AS mean_time_change_percent,
        b.max_time AS before_max_time,
        a.max_time AS after_max_time,
        COALESCE(a.max_time, 0) - COALESCE(b.max_time, 0) AS max_time_change,
        b.shared_blks_hit AS before_cache_hits,
        a.shared_blks_hit AS after_cache_hits,
        b.shared_blks_read AS before_cache_reads,
        a.shared_blks_read AS after_cache_reads,
        CASE
            WHEN a.mean_time < b.mean_time THEN 'improved'
            WHEN a.mean_time > b.mean_time THEN 'degraded'
            ELSE 'unchanged'
        END AS improvement_status
    FROM query_performance_snapshots b
    FULL OUTER JOIN query_performance_snapshots a ON b.query_hash = a.query_hash
    WHERE b.snapshot_id = p_before_snapshot_id
        AND a.snapshot_id = p_after_snapshot_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Migration Completion
-- ============================================================================

-- Grant permissions
GRANT SELECT ON query_fingerprints TO pg_monitor;
GRANT SELECT ON explain_plans TO pg_monitor;
GRANT SELECT ON v_latest_explain_plans TO pg_monitor;
GRANT SELECT ON index_recommendations TO pg_monitor;
GRANT SELECT ON query_anomalies TO pg_monitor;
GRANT SELECT ON query_baselines TO pg_monitor;
GRANT SELECT ON performance_snapshots TO pg_monitor;
GRANT SELECT ON query_performance_snapshots TO pg_monitor;

GRANT EXECUTE ON FUNCTION normalize_query_text TO pg_monitor;
GRANT EXECUTE ON FUNCTION fingerprint_query_hash TO pg_monitor;
GRANT EXECUTE ON FUNCTION analyze_explain_for_indexes TO pg_monitor;
GRANT EXECUTE ON FUNCTION calculate_baselines_and_anomalies TO pg_monitor;
GRANT EXECUTE ON FUNCTION create_performance_snapshot TO pg_monitor;
GRANT EXECUTE ON FUNCTION compare_snapshots TO pg_monitor;

-- Record migration completion
INSERT INTO schema_migrations (version, description, executed_at)
VALUES (4, 'Phase 4.4: Advanced query performance features', NOW())
ON CONFLICT DO NOTHING;
