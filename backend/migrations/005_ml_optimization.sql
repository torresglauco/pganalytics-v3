-- Phase 4.5: ML-Based Query Optimization Suggestions
-- Tables and functions for workload pattern detection, query rewrite suggestions,
-- parameter optimization, performance prediction, and optimization workflow

-- ============================================================================
-- 1. Workload Pattern Detection
-- ============================================================================

CREATE TABLE IF NOT EXISTS workload_patterns (
    id BIGSERIAL PRIMARY KEY,
    database_name VARCHAR(63) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL, -- 'hourly_peak', 'daily_cycle', 'weekly_pattern', 'batch_job'
    pattern_metadata JSONB NOT NULL, -- {peak_hour: 8, variance: 0.15, confidence: 0.92, affected_queries: 42}
    detection_timestamp TIMESTAMP DEFAULT NOW(),
    description TEXT,
    affected_query_count INTEGER DEFAULT 0,
    CONSTRAINT uk_pattern_database_type UNIQUE(database_name, pattern_type)
);

CREATE INDEX IF NOT EXISTS idx_workload_pattern_database ON workload_patterns(database_name, pattern_type);
CREATE INDEX IF NOT EXISTS idx_workload_pattern_detection ON workload_patterns(detection_timestamp DESC);

-- ============================================================================
-- 2. Query Rewrite Suggestions
-- ============================================================================

CREATE TABLE IF NOT EXISTS query_rewrite_suggestions (
    id BIGSERIAL PRIMARY KEY,
    query_hash BIGINT NOT NULL,
    fingerprint_hash BIGINT,
    suggestion_type VARCHAR(100) NOT NULL, -- 'n_plus_one_detected', 'subquery_optimization', 'join_reorder', 'missing_limit'
    description TEXT NOT NULL,
    original_query TEXT,
    suggested_rewrite TEXT NOT NULL,
    reasoning TEXT,
    estimated_improvement_percent FLOAT NOT NULL DEFAULT 0,
    confidence_score FLOAT NOT NULL DEFAULT 0.8, -- 0-1 confidence
    dismissed BOOLEAN DEFAULT FALSE,
    implemented BOOLEAN DEFAULT FALSE,
    implementation_notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_rewrite_query_hash FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE,
    CONSTRAINT uk_rewrite_suggestion UNIQUE(query_hash, suggestion_type)
);

CREATE INDEX IF NOT EXISTS idx_rewrite_query_hash ON query_rewrite_suggestions(query_hash, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rewrite_confidence ON query_rewrite_suggestions(confidence_score DESC, estimated_improvement_percent DESC);
CREATE INDEX IF NOT EXISTS idx_rewrite_type ON query_rewrite_suggestions(suggestion_type, created_at DESC);

-- ============================================================================
-- 3. Parameter Tuning Suggestions
-- ============================================================================

CREATE TABLE IF NOT EXISTS parameter_tuning_suggestions (
    id BIGSERIAL PRIMARY KEY,
    query_hash BIGINT NOT NULL,
    fingerprint_hash BIGINT,
    parameter_name VARCHAR(100) NOT NULL, -- 'work_mem', 'sort_mem', 'limit', 'batch_size'
    current_value VARCHAR(255),
    recommended_value VARCHAR(255) NOT NULL,
    reasoning TEXT,
    estimated_improvement_percent FLOAT NOT NULL DEFAULT 0,
    confidence_score FLOAT NOT NULL DEFAULT 0.75, -- 0-1 confidence
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_param_query_hash FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE,
    CONSTRAINT uk_param_suggestion UNIQUE(query_hash, parameter_name)
);

CREATE INDEX IF NOT EXISTS idx_param_query_hash ON parameter_tuning_suggestions(query_hash, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_param_confidence ON parameter_tuning_suggestions(confidence_score DESC, estimated_improvement_percent DESC);

-- ============================================================================
-- 4. Optimization Recommendations (Aggregated)
-- ============================================================================

CREATE TABLE IF NOT EXISTS optimization_recommendations (
    id BIGSERIAL PRIMARY KEY,
    query_hash BIGINT NOT NULL,
    source_type VARCHAR(50) NOT NULL, -- 'index', 'rewrite', 'parameter', 'workload'
    source_id BIGINT, -- Reference to source table (rewrite_suggestions.id, parameter_tuning_suggestions.id, etc)
    recommendation_text TEXT NOT NULL,
    detailed_explanation TEXT,
    estimated_improvement_percent FLOAT NOT NULL DEFAULT 0,
    confidence_score FLOAT NOT NULL DEFAULT 0.8, -- 0-1
    urgency_score FLOAT NOT NULL DEFAULT 0.5, -- 0-1 (frequency × impact)
    roi_score FLOAT NOT NULL DEFAULT 0, -- confidence × improvement × urgency
    implementation_complexity VARCHAR(20), -- 'low', 'medium', 'high'
    dismissal_reason TEXT,
    is_dismissed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_recommendation_query FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_recommendation_roi ON optimization_recommendations(roi_score DESC, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recommendation_query_hash ON optimization_recommendations(query_hash, roi_score DESC);
CREATE INDEX IF NOT EXISTS idx_recommendation_source ON optimization_recommendations(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_dismissed ON optimization_recommendations(is_dismissed, roi_score DESC);

-- ============================================================================
-- 5. Optimization Implementation Tracking
-- ============================================================================

CREATE TABLE IF NOT EXISTS optimization_implementations (
    id BIGSERIAL PRIMARY KEY,
    recommendation_id BIGINT NOT NULL,
    query_hash BIGINT NOT NULL,
    implementation_timestamp TIMESTAMP DEFAULT NOW(),
    implementation_notes TEXT,
    pre_optimization_stats JSONB, -- Snapshot of metrics before: {mean_time: 125.5, calls: 1000, total_time: 125500}
    post_optimization_stats JSONB, -- Snapshot after: {mean_time: 95.3, calls: 1000, total_time: 95300}
    actual_improvement_percent FLOAT, -- Measured improvement
    actual_improvement_seconds FLOAT, -- Total time saved
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'implemented', 'reverted', 'failed'
    error_message TEXT,
    measured_at TIMESTAMP, -- When post-optimization metrics were captured
    CONSTRAINT fk_impl_recommendation FOREIGN KEY (recommendation_id) REFERENCES optimization_recommendations(id) ON DELETE CASCADE,
    CONSTRAINT fk_impl_query FOREIGN KEY (query_hash) REFERENCES metrics_pg_stats_query(query_hash) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_implementation_recommendation ON optimization_implementations(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_implementation_query ON optimization_implementations(query_hash, implementation_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_implementation_status ON optimization_implementations(status, implementation_timestamp DESC);

-- ============================================================================
-- 6. Query Performance Models
-- ============================================================================

CREATE TABLE IF NOT EXISTS query_performance_models (
    id BIGSERIAL PRIMARY KEY,
    model_type VARCHAR(100) NOT NULL, -- 'linear_regression', 'decision_tree', 'random_forest', 'xgboost'
    model_name VARCHAR(255),
    database_name VARCHAR(63),
    model_binary BYTEA, -- Serialized model (joblib pickle or other format)
    model_json JSONB, -- JSON representation of model coefficients/parameters
    feature_names TEXT[] NOT NULL, -- Features used in training
    training_sample_size INTEGER,
    r_squared FLOAT, -- R² score for accuracy
    rmse FLOAT, -- Root Mean Square Error
    mae FLOAT, -- Mean Absolute Error
    feature_importance JSONB, -- Feature importance scores
    training_timestamp TIMESTAMP DEFAULT NOW(),
    last_updated TIMESTAMP DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    metrics JSONB -- Additional metrics: {cross_val_scores: [...], outliers_removed: 42}
);

CREATE INDEX IF NOT EXISTS idx_model_database_active ON query_performance_models(database_name, is_active, last_updated DESC);
CREATE INDEX IF NOT EXISTS idx_model_type ON query_performance_models(model_type, is_active);

-- ============================================================================
-- 7. Utility Views
-- ============================================================================

-- Top optimization opportunities ranked by ROI
CREATE OR REPLACE VIEW v_top_optimization_recommendations AS
SELECT
    r.id,
    r.query_hash,
    r.source_type,
    r.recommendation_text,
    r.estimated_improvement_percent,
    r.confidence_score,
    r.urgency_score,
    r.roi_score,
    q.query_text,
    q.calls,
    q.total_exec_time_ms,
    q.mean_exec_time_ms,
    RANK() OVER (ORDER BY r.roi_score DESC, r.created_at DESC) as rank
FROM optimization_recommendations r
JOIN metrics_pg_stats_query q ON r.query_hash = q.query_hash
WHERE r.is_dismissed = FALSE
AND r.roi_score > 10
ORDER BY r.roi_score DESC;

-- Implementation results and measured improvements
CREATE OR REPLACE VIEW v_optimization_results AS
SELECT
    i.id as implementation_id,
    r.id as recommendation_id,
    r.query_hash,
    r.recommendation_text,
    r.estimated_improvement_percent as estimated_improvement,
    i.actual_improvement_percent as actual_improvement,
    ROUND(((i.actual_improvement_percent - r.estimated_improvement_percent) / NULLIF(r.estimated_improvement_percent, 0) * 100)::NUMERIC, 2) as prediction_error_percent,
    i.status,
    i.implementation_timestamp,
    i.measured_at,
    (i.measured_at - i.implementation_timestamp) as time_to_measurement,
    r.confidence_score,
    i.actual_improvement_seconds
FROM optimization_implementations i
JOIN optimization_recommendations r ON i.recommendation_id = r.id
ORDER BY i.implementation_timestamp DESC;

-- Pattern summary view
CREATE OR REPLACE VIEW v_workload_pattern_summary AS
SELECT
    database_name,
    pattern_type,
    COUNT(*) as occurrences,
    AVG((pattern_metadata->>'confidence')::FLOAT) as avg_confidence,
    MAX(detection_timestamp) as latest_detection,
    JSONB_AGG(pattern_metadata) as all_metadata
FROM workload_patterns
GROUP BY database_name, pattern_type;

-- ============================================================================
-- 8. PostgreSQL Functions for Analysis
-- ============================================================================

-- Function to detect workload patterns from historical data
CREATE OR REPLACE FUNCTION detect_workload_patterns(
    p_database_name VARCHAR(63),
    p_lookback_days INTEGER DEFAULT 30
) RETURNS TABLE(pattern_id BIGINT, pattern_type VARCHAR, confidence FLOAT) AS $$
BEGIN
    -- Validate input
    IF p_lookback_days < 7 THEN
        p_lookback_days := 7;
    END IF;
    IF p_lookback_days > 365 THEN
        p_lookback_days := 365;
    END IF;

    -- Step 1: Hourly pattern detection
    -- Group by hour, calculate statistics, identify peaks
    WITH hourly_stats AS (
        SELECT
            EXTRACT(HOUR FROM collected_at)::INTEGER as hour_of_day,
            DATE(collected_at) as stat_date,
            COUNT(*) as query_count,
            AVG(mean_exec_time_ms) as avg_exec_time,
            MAX(mean_exec_time_ms) as max_exec_time
        FROM metrics_pg_stats_query
        WHERE database_name = p_database_name
        AND collected_at > NOW() - (p_lookback_days || ' days')::INTERVAL
        GROUP BY EXTRACT(HOUR FROM collected_at), DATE(collected_at)
    ),

    -- Step 2: Aggregate statistics by hour
    hourly_aggregated AS (
        SELECT
            hour_of_day,
            COUNT(DISTINCT stat_date) as days_observed,
            AVG(query_count) as avg_count,
            STDDEV_POP(query_count) as stddev_count,
            AVG(avg_exec_time) as avg_time,
            STDDEV_POP(avg_exec_time) as stddev_time,
            MAX(query_count) as max_count,
            MIN(query_count) as min_count
        FROM hourly_stats
        GROUP BY hour_of_day
    ),

    -- Step 3: Calculate overall statistics
    overall_stats AS (
        SELECT
            AVG(avg_count) as overall_avg_count,
            STDDEV_POP(avg_count) as overall_stddev_count,
            AVG(avg_time) as overall_avg_time,
            STDDEV_POP(avg_time) as overall_stddev_time,
            COUNT(DISTINCT hour_of_day) as hours_with_data,
            MAX(days_observed) as max_days_observed
        FROM hourly_aggregated
    ),

    -- Step 4: Calculate z-scores and identify peaks
    peak_hours AS (
        SELECT
            ha.hour_of_day,
            ha.days_observed,
            ha.avg_count,
            ha.stddev_count,
            ha.avg_time,
            ha.stddev_time,
            os.overall_avg_count,
            os.overall_stddev_count,
            os.overall_avg_time,
            os.overall_stddev_time,
            os.max_days_observed,
            -- Z-scores (how many standard deviations from mean)
            CASE WHEN os.overall_stddev_count > 0
                 THEN (ha.avg_count - os.overall_avg_count) / os.overall_stddev_count
                 ELSE 0 END as z_score_count,
            CASE WHEN os.overall_stddev_time > 0
                 THEN (ha.avg_time - os.overall_avg_time) / os.overall_stddev_time
                 ELSE 0 END as z_score_time,
            -- Consistency score (lower stddev = higher consistency)
            CASE WHEN ha.avg_count > 0
                 THEN LEAST(1.0 - (ha.stddev_count / ha.avg_count), 1.0)
                 ELSE 0 END as consistency_score,
            -- Recurrence score (how many days showed this pattern)
            ha.days_observed::FLOAT / NULLIF(os.max_days_observed::FLOAT, 0) as recurrence_score
        FROM hourly_aggregated ha, overall_stats os
    ),

    -- Step 5: Filter for actual peaks (z-score > 1.0)
    significant_peaks AS (
        SELECT
            hour_of_day,
            days_observed,
            avg_count,
            stddev_count,
            avg_time,
            z_score_count,
            z_score_time,
            consistency_score,
            recurrence_score,
            -- Final confidence = consistency × recurrence
            ROUND(LEAST(1.0, (consistency_score * recurrence_score))::NUMERIC, 4) as confidence
        FROM peak_hours
        WHERE z_score_count > 1.0 OR z_score_time > 1.0
    ),

    -- Step 6: Insert patterns and get results
    inserted_patterns AS (
        INSERT INTO workload_patterns (
            database_name, pattern_type, pattern_metadata,
            detection_timestamp, description, affected_query_count
        )
        SELECT
            p_database_name,
            'hourly_peak',
            jsonb_build_object(
                'peak_hour', sp.hour_of_day,
                'variance', ROUND((sp.stddev_count / NULLIF(sp.avg_count, 0))::NUMERIC, 4),
                'confidence', sp.confidence,
                'affected_queries', CEIL(sp.avg_count),
                'z_score_count', ROUND(sp.z_score_count::NUMERIC, 2),
                'z_score_time', ROUND(sp.z_score_time::NUMERIC, 2),
                'days_observed', sp.days_observed,
                'consistency_score', ROUND(sp.consistency_score::NUMERIC, 4),
                'recurrence_score', ROUND(sp.recurrence_score::NUMERIC, 4)
            ),
            NOW(),
            'Peak load detected at hour ' || sp.hour_of_day::TEXT || ' UTC (' || ROUND((sp.confidence * 100)::NUMERIC, 1)::TEXT || '% confidence)',
            CEIL(sp.avg_count)
        FROM significant_peaks sp
        ON CONFLICT (database_name, pattern_type)
        DO UPDATE SET
            pattern_metadata = EXCLUDED.pattern_metadata,
            detection_timestamp = NOW(),
            description = EXCLUDED.description,
            affected_query_count = EXCLUDED.affected_query_count
        RETURNING id, 'hourly_peak'::VARCHAR, (pattern_metadata->>'confidence')::FLOAT as conf
    )
    SELECT id, pattern_type, conf FROM inserted_patterns
    ORDER BY conf DESC;

END;
$$ LANGUAGE plpgsql;

-- Function to calculate ROI score for recommendations
CREATE OR REPLACE FUNCTION calculate_roi_score(
    p_confidence FLOAT,
    p_improvement_percent FLOAT,
    p_urgency_score FLOAT
) RETURNS FLOAT AS $$
BEGIN
    RETURN ROUND((p_confidence * p_improvement_percent * p_urgency_score)::NUMERIC, 2)::FLOAT;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to calculate urgency score based on query impact
CREATE OR REPLACE FUNCTION calculate_urgency_score(
    p_query_calls BIGINT,
    p_mean_exec_time_ms FLOAT
) RETURNS FLOAT AS $$
DECLARE
    v_frequency_score FLOAT;
    v_impact_score FLOAT;
BEGIN
    -- Frequency: normalized to 0-1 (assume max is 100,000 calls/day)
    v_frequency_score := LEAST(p_query_calls::FLOAT / 100000.0, 1.0);

    -- Impact: execution time (assume max 10 seconds)
    v_impact_score := LEAST(p_mean_exec_time_ms / 10000.0, 1.0);

    -- Urgency is product of frequency and impact
    RETURN ROUND((v_frequency_score * v_impact_score)::NUMERIC, 2)::FLOAT;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get top recommendations for a query
CREATE OR REPLACE FUNCTION get_top_recommendations_for_query(
    p_query_hash BIGINT,
    p_limit INTEGER DEFAULT 10
) RETURNS TABLE(
    rec_id BIGINT,
    source_type VARCHAR,
    recommendation_text TEXT,
    estimated_improvement FLOAT,
    confidence FLOAT,
    roi_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        r.id,
        r.source_type,
        r.recommendation_text,
        r.estimated_improvement_percent,
        r.confidence_score,
        r.roi_score
    FROM optimization_recommendations r
    WHERE r.query_hash = p_query_hash
    AND r.is_dismissed = FALSE
    ORDER BY r.roi_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to record optimization implementation
CREATE OR REPLACE FUNCTION record_optimization_implementation(
    p_recommendation_id BIGINT,
    p_query_hash BIGINT,
    p_notes TEXT,
    p_pre_stats JSONB
) RETURNS TABLE(impl_id BIGINT, status VARCHAR) AS $$
DECLARE
    v_impl_id BIGINT;
BEGIN
    INSERT INTO optimization_implementations (
        recommendation_id,
        query_hash,
        implementation_notes,
        pre_optimization_stats,
        status
    )
    VALUES (
        p_recommendation_id,
        p_query_hash,
        p_notes,
        p_pre_stats,
        'pending'
    )
    RETURNING optimization_implementations.id
    INTO v_impl_id;

    RETURN QUERY
    SELECT v_impl_id, 'pending'::VARCHAR;
END;
$$ LANGUAGE plpgsql;

-- Function to update implementation results
CREATE OR REPLACE FUNCTION update_implementation_results(
    p_implementation_id BIGINT,
    p_post_stats JSONB,
    p_actual_improvement_percent FLOAT,
    p_actual_improvement_seconds FLOAT,
    p_status VARCHAR DEFAULT 'implemented'
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE optimization_implementations
    SET
        post_optimization_stats = p_post_stats,
        actual_improvement_percent = p_actual_improvement_percent,
        actual_improvement_seconds = p_actual_improvement_seconds,
        status = p_status,
        measured_at = NOW()
    WHERE id = p_implementation_id;

    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function to generate pattern metadata JSON
CREATE OR REPLACE FUNCTION create_pattern_metadata(
    p_peak_hour INT,
    p_variance FLOAT,
    p_confidence FLOAT,
    p_affected_queries INT
) RETURNS JSONB AS $$
BEGIN
    RETURN jsonb_build_object(
        'peak_hour', p_peak_hour,
        'variance', ROUND(p_variance::NUMERIC, 4),
        'confidence', ROUND(p_confidence::NUMERIC, 4),
        'affected_queries', p_affected_queries
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to generate query rewrite suggestions
-- Detects anti-patterns: N+1, inefficient joins, missing indexes, subqueries, IN vs ANY
CREATE OR REPLACE FUNCTION generate_rewrite_suggestions(
    p_query_hash BIGINT
) RETURNS TABLE(
    suggestion_id BIGINT,
    suggestion_type VARCHAR,
    confidence FLOAT
) AS $$
DECLARE
    v_query_text TEXT;
    v_fingerprint_hash BIGINT;
    v_mean_exec_time FLOAT;
    v_calls BIGINT;
    v_database_name VARCHAR(63);
    v_suggestion_count INT := 0;
BEGIN
    -- Get query details
    SELECT query_text, fingerprint_hash, mean_exec_time_ms, calls, database_name
    INTO v_query_text, v_fingerprint_hash, v_mean_exec_time, v_calls, v_database_name
    FROM metrics_pg_stats_query
    WHERE query_hash = p_query_hash
    ORDER BY collected_at DESC
    LIMIT 1;

    IF v_query_text IS NULL THEN
        RETURN;
    END IF;

    -- Detect N+1 patterns: multiple rapid calls with same fingerprint
    -- N+1 occurs when: call_count > 50 AND frequency indicates rapid execution
    IF v_calls > 50 AND v_mean_exec_time < 200 THEN
        INSERT INTO query_rewrite_suggestions (
            query_hash, fingerprint_hash, suggestion_type, description,
            original_query, suggested_rewrite, reasoning,
            estimated_improvement_percent, confidence_score, created_at, updated_at
        )
        VALUES (
            p_query_hash,
            v_fingerprint_hash,
            'n_plus_one_detected',
            'Multiple queries with identical pattern detected in rapid succession',
            v_query_text,
            'Combine into single query using IN clause or JOIN for batch processing',
            'Query called ' || v_calls || ' times with mean execution ' ||
            ROUND(v_mean_exec_time::NUMERIC, 1) || 'ms. Consider batching queries with IN clause or converting to single JOIN query.',
            ROUND((MIN(99.0, ((v_calls - 1)::FLOAT / v_calls::FLOAT) * 100))::NUMERIC, 1)::FLOAT,
            ROUND(LEAST(0.90 * (1.0 - EXP(-v_calls::FLOAT / 100.0)), 0.95)::NUMERIC, 2)::FLOAT,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, suggestion_type) DO UPDATE SET
            updated_at = NOW(),
            estimated_improvement_percent = EXCLUDED.estimated_improvement_percent,
            confidence_score = EXCLUDED.confidence_score;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Detect inefficient joins: Nested Loop when Hash Join would be better
    -- Check for Nested Loop + Seq Scan combination in EXPLAIN plan
    IF EXISTS (
        SELECT 1 FROM explain_plans ep
        WHERE ep.query_hash = p_query_hash
        AND ep.has_nested_loop = TRUE
        AND ep.has_seq_scan = TRUE
        LIMIT 1
    ) AND v_calls > 100 THEN
        INSERT INTO query_rewrite_suggestions (
            query_hash, fingerprint_hash, suggestion_type, description,
            original_query, suggested_rewrite, reasoning,
            estimated_improvement_percent, confidence_score, created_at, updated_at
        )
        VALUES (
            p_query_hash,
            v_fingerprint_hash,
            'inefficient_join_detected',
            'Nested Loop join detected where Hash Join would be more efficient',
            v_query_text,
            'Add index on join columns or reorder joins to enable Hash Join optimization',
            'EXPLAIN plan shows Nested Loop with Sequential Scans. Adding index on join columns ' ||
            'or reordering tables would allow optimizer to choose Hash Join, significantly improving performance.',
            85.0::FLOAT,
            0.80::FLOAT,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, suggestion_type) DO NOTHING;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Detect missing indexes: Sequential scan on large result set
    -- When: Seq Scan + high call frequency + high execution time
    IF EXISTS (
        SELECT 1 FROM explain_plans ep
        WHERE ep.query_hash = p_query_hash
        AND ep.has_seq_scan = TRUE
        LIMIT 1
    ) AND v_calls > 100 AND v_mean_exec_time > 100 THEN
        INSERT INTO query_rewrite_suggestions (
            query_hash, fingerprint_hash, suggestion_type, description,
            original_query, suggested_rewrite, reasoning,
            estimated_improvement_percent, confidence_score, created_at, updated_at
        )
        VALUES (
            p_query_hash,
            v_fingerprint_hash,
            'missing_index_detected',
            'Sequential scan on table that could benefit from index creation',
            v_query_text,
            'Create index on WHERE/JOIN columns: CREATE INDEX idx_table_column ON table(column)',
            'EXPLAIN shows Sequential Scan called ' || v_calls || ' times with mean execution ' ||
            ROUND(v_mean_exec_time::NUMERIC, 1) || 'ms. Adding index on filter/join columns ' ||
            'would convert to Index Scan or Index Seek, improving performance significantly.',
            80.0::FLOAT,
            0.83::FLOAT,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, suggestion_type) DO NOTHING;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Detect subquery optimization: SubPlan in EXPLAIN can often be rewritten as JOIN
    -- When: SubPlan exists + query called frequently
    IF v_query_text ILIKE '%WHERE IN (SELECT%' AND v_calls > 50 THEN
        INSERT INTO query_rewrite_suggestions (
            query_hash, fingerprint_hash, suggestion_type, description,
            original_query, suggested_rewrite, reasoning,
            estimated_improvement_percent, confidence_score, created_at, updated_at
        )
        VALUES (
            p_query_hash,
            v_fingerprint_hash,
            'subquery_optimization',
            'Inefficient subquery detected that can be rewritten as JOIN',
            v_query_text,
            'Convert WHERE IN (SELECT ...) to INNER JOIN ... ON for better optimizer control',
            'Query uses subquery in WHERE clause. JOINs often allow better query optimization and are ' ||
            'easier for the optimizer to parallelize. Rewriting as JOIN usually improves performance.',
            75.0::FLOAT,
            0.75::FLOAT,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, suggestion_type) DO NOTHING;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Detect IN vs ANY optimization: IN with many values can use parameterized ANY
    -- When: Query has IN clause with multiple values
    IF v_query_text ILIKE '%IN (%;' AND LENGTH(v_query_text) > 200 THEN
        INSERT INTO query_rewrite_suggestions (
            query_hash, fingerprint_hash, suggestion_type, description,
            original_query, suggested_rewrite, reasoning,
            estimated_improvement_percent, confidence_score, created_at, updated_at
        )
        VALUES (
            p_query_hash,
            v_fingerprint_hash,
            'in_vs_any_optimization',
            'IN clause with multiple values can be optimized using ANY with array',
            v_query_text,
            'Replace WHERE col IN (...) with WHERE col = ANY(ARRAY[...]) for parameterized queries',
            'Long IN list detected. Using ANY with array parameter allows better statement caching ' ||
            'and can improve performance with parameterized queries. Especially beneficial with ORM frameworks.',
            15.0::FLOAT,
            0.65::FLOAT,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, suggestion_type) DO NOTHING;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Return all generated suggestions for this query
    RETURN QUERY
    SELECT
        qrs.id,
        qrs.suggestion_type::VARCHAR,
        qrs.confidence_score
    FROM query_rewrite_suggestions qrs
    WHERE qrs.query_hash = p_query_hash
    AND qrs.dismissed = FALSE
    ORDER BY qrs.confidence_score DESC;

END;
$$ LANGUAGE plpgsql;

-- Function to generate parameter optimization suggestions
-- Detects: missing LIMIT, work_mem optimization, batch size opportunities
CREATE OR REPLACE FUNCTION optimize_parameters(
    p_query_hash BIGINT
) RETURNS TABLE(
    suggestion_count INT,
    parameter_types TEXT[]
) AS $$
DECLARE
    v_query_hash BIGINT;
    v_query_text TEXT;
    v_has_sort BOOLEAN;
    v_has_limit BOOLEAN;
    v_mean_exec_time FLOAT;
    v_calls BIGINT;
    v_current_work_mem VARCHAR;
    v_result_rows BIGINT;
    v_current_work_mem_mb INT;
    v_new_work_mem_mb INT;
    v_suggested_limit INT;
    v_suggestion_count INT := 0;
    i INT;
BEGIN
    -- Step 1: Get query details from metrics table
    SELECT mpsq.query_hash, mpsq.query_text, mpsq.mean_exec_time_ms, mpsq.calls, mpsq.rows
    INTO v_query_hash, v_query_text, v_mean_exec_time, v_calls, v_result_rows
    FROM metrics_pg_stats_query mpsq
    WHERE mpsq.query_hash = p_query_hash
    ORDER BY mpsq.collected_at DESC
    LIMIT 1;

    -- Step 2: Validate query exists
    IF v_query_hash IS NULL THEN
        RETURN QUERY SELECT 0::INT, ARRAY[]::TEXT[];
        RETURN;
    END IF;

    -- Step 3: Analyze query characteristics
    v_has_sort := (v_query_text ILIKE '%ORDER BY%' OR
                   v_query_text ILIKE '%GROUP BY%' OR
                   v_query_text ILIKE '%DISTINCT%');
    v_has_limit := v_query_text ILIKE '%LIMIT%';

    -- Get current work_mem setting
    BEGIN
        v_current_work_mem := current_setting('work_mem');
        -- Parse work_mem value (e.g., "4MB" -> 4)
        v_current_work_mem_mb := CAST(split_part(v_current_work_mem, 'M', 1) AS INT);
    EXCEPTION WHEN OTHERS THEN
        v_current_work_mem := '4MB';
        v_current_work_mem_mb := 4;
    END;

    -- Step 4: Generate LIMIT recommendations for large result sets
    -- Pattern: No LIMIT clause + slow execution + large result set
    IF NOT v_has_limit AND v_mean_exec_time > 100 AND v_result_rows > 100 THEN
        v_suggested_limit := GREATEST(CEIL(v_result_rows::FLOAT / 10.0)::INT, 100);

        INSERT INTO parameter_tuning_suggestions (
            query_hash, parameter_name, current_value, recommended_value,
            reasoning, estimated_improvement_percent, confidence_score,
            created_at, updated_at
        )
        VALUES (
            v_query_hash,
            'LIMIT',
            'NOT SET',
            'LIMIT ' || v_suggested_limit::TEXT,
            'Query returns ' || v_result_rows::TEXT || ' rows with mean execution ' ||
            ROUND(v_mean_exec_time::NUMERIC, 0)::TEXT || 'ms. ' ||
            'Consider adding LIMIT to reduce scanning and result processing.',
            CASE
                WHEN v_result_rows > 10000 THEN 85.0
                WHEN v_result_rows > 5000 THEN 80.0
                WHEN v_result_rows > 1000 THEN 75.0
                ELSE 50.0
            END,
            CASE
                WHEN v_result_rows > 10000 THEN 0.95
                WHEN v_result_rows > 5000 THEN 0.90
                WHEN v_result_rows > 1000 THEN 0.85
                ELSE 0.70
            END,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
            updated_at = NOW(),
            estimated_improvement_percent = EXCLUDED.estimated_improvement_percent,
            confidence_score = EXCLUDED.confidence_score;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Step 5: Generate work_mem recommendations for sort operations
    -- Pattern: Has sort nodes + slow execution + called frequently
    IF v_has_sort AND v_mean_exec_time > 200 AND v_calls > 10 THEN
        -- Recommend increase to 1.5x current work_mem
        v_new_work_mem_mb := CEIL(v_current_work_mem_mb * 1.5);

        INSERT INTO parameter_tuning_suggestions (
            query_hash, parameter_name, current_value, recommended_value,
            reasoning, estimated_improvement_percent, confidence_score,
            created_at, updated_at
        )
        VALUES (
            v_query_hash,
            'work_mem',
            v_current_work_mem,
            v_new_work_mem_mb::TEXT || 'MB',
            'Query has ORDER BY/GROUP BY/DISTINCT operations taking ' ||
            ROUND(v_mean_exec_time::NUMERIC, 0)::TEXT || 'ms with ' || v_calls::TEXT ||
            ' calls. Increasing work_mem allows faster in-memory sorting.',
            CASE
                WHEN v_mean_exec_time > 500 THEN 35.0
                WHEN v_mean_exec_time > 300 THEN 25.0
                ELSE 15.0
            END,
            CASE
                WHEN v_mean_exec_time > 500 THEN 0.90
                WHEN v_mean_exec_time > 300 THEN 0.85
                ELSE 0.80
            END,
            NOW(),
            NOW()
        )
        ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
            updated_at = NOW(),
            estimated_improvement_percent = EXCLUDED.estimated_improvement_percent,
            confidence_score = EXCLUDED.confidence_score;

        v_suggestion_count := v_suggestion_count + 1;
    END IF;

    -- Step 6: Generate batch size recommendations for N+1 patterns
    -- Pattern: High call frequency (> 100 calls)
    IF v_calls > 100 THEN
        -- Recommend batch sizes: 50, 100, 500
        FOR i IN ARRAY[50, 100, 500] LOOP
            INSERT INTO parameter_tuning_suggestions (
                query_hash, parameter_name, current_value, recommended_value,
                reasoning, estimated_improvement_percent, confidence_score,
                created_at, updated_at
            )
            VALUES (
                v_query_hash,
                'batch_size',
                '1',
                i::TEXT,
                'Query called ' || v_calls::TEXT || ' times (high frequency). ' ||
                'Consider batching with size ' || i::TEXT ||
                ' to reduce network round-trips and improve overall throughput.',
                CASE
                    WHEN i = 50 THEN 75.0
                    WHEN i = 100 THEN 72.0
                    ELSE 70.0
                END,
                CASE
                    WHEN i = 50 THEN 0.75
                    WHEN i = 100 THEN 0.73
                    ELSE 0.70
                END,
                NOW(),
                NOW()
            )
            ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
                updated_at = NOW(),
                estimated_improvement_percent = EXCLUDED.estimated_improvement_percent,
                confidence_score = EXCLUDED.confidence_score;

            v_suggestion_count := v_suggestion_count + 1;
        END LOOP;
    END IF;

    -- Step 7: Return results with parameter type list
    RETURN QUERY
    SELECT
        COUNT(*)::INT,
        ARRAY_AGG(DISTINCT parameter_name ORDER BY parameter_name)
    FROM parameter_tuning_suggestions
    WHERE query_hash = v_query_hash
    AND created_at > NOW() - INTERVAL '1 minute';

END;
$$ LANGUAGE plpgsql;

-- Function to aggregate all recommendations for a query into optimization_recommendations
-- Combines rewrite suggestions, parameter tuning, and workload patterns
CREATE OR REPLACE FUNCTION aggregate_recommendations_for_query(
    p_query_hash BIGINT
) RETURNS TABLE(
    recommendation_count INT,
    source_types TEXT[]
) AS $$
DECLARE
    v_urgency_score FLOAT;
    v_roi_score FLOAT;
    v_query_hash BIGINT;
    v_mean_exec_time FLOAT;
    v_calls BIGINT;
    v_rows BIGINT;
    v_frequency_score FLOAT;
    v_impact_score FLOAT;
    v_count INT := 0;
BEGIN
    -- Step 1: Get query metrics for urgency calculation
    SELECT mpsq.query_hash, mpsq.mean_exec_time_ms, mpsq.calls, mpsq.rows
    INTO v_query_hash, v_mean_exec_time, v_calls, v_rows
    FROM metrics_pg_stats_query mpsq
    WHERE mpsq.query_hash = p_query_hash
    ORDER BY mpsq.collected_at DESC
    LIMIT 1;

    IF v_query_hash IS NULL THEN
        RETURN QUERY SELECT 0::INT, ARRAY[]::TEXT[];
        RETURN;
    END IF;

    -- Step 2: Calculate urgency score components
    -- Frequency: 0-1 (assume max 100,000 calls/day = 1,157 calls/min avg)
    v_frequency_score := LEAST(v_calls::FLOAT / 100000.0, 1.0);

    -- Impact: 0-1 (assume max 10 seconds execution time)
    v_impact_score := LEAST(v_mean_exec_time::FLOAT / 10000.0, 1.0);

    -- Urgency: product of frequency and impact
    v_urgency_score := v_frequency_score * v_impact_score;

    -- Step 3: Aggregate rewrite suggestions
    INSERT INTO optimization_recommendations (
        query_hash, source_type, source_id, recommendation_text, detailed_explanation,
        estimated_improvement_percent, confidence_score, urgency_score, roi_score,
        implementation_complexity, created_at, updated_at
    )
    SELECT
        qrs.query_hash,
        'rewrite'::VARCHAR,
        qrs.id,
        qrs.suggestion_type || ': ' || qrs.description,
        'Original: ' || qrs.original_query || ' → Suggested: ' || qrs.suggested_rewrite,
        qrs.estimated_improvement_percent,
        qrs.confidence_score,
        v_urgency_score,
        ROUND((qrs.confidence_score * LEAST(qrs.estimated_improvement_percent / 100.0, 1.0) * v_urgency_score)::NUMERIC, 4)::FLOAT,
        CASE
            WHEN qrs.suggestion_type = 'n_plus_one_detected' THEN 'medium'::VARCHAR
            WHEN qrs.suggestion_type = 'inefficient_join_detected' THEN 'high'::VARCHAR
            WHEN qrs.suggestion_type = 'missing_index_detected' THEN 'high'::VARCHAR
            WHEN qrs.suggestion_type = 'subquery_optimization' THEN 'medium'::VARCHAR
            WHEN qrs.suggestion_type = 'in_vs_any_optimization' THEN 'low'::VARCHAR
            ELSE 'medium'::VARCHAR
        END,
        NOW(),
        NOW()
    FROM query_rewrite_suggestions qrs
    WHERE qrs.query_hash = p_query_hash
    AND qrs.dismissed = FALSE
    ON CONFLICT DO NOTHING;

    GET DIAGNOSTICS v_count = ROW_COUNT;

    -- Step 4: Aggregate parameter tuning suggestions
    INSERT INTO optimization_recommendations (
        query_hash, source_type, source_id, recommendation_text, detailed_explanation,
        estimated_improvement_percent, confidence_score, urgency_score, roi_score,
        implementation_complexity, created_at, updated_at
    )
    SELECT
        pts.query_hash,
        'parameter'::VARCHAR,
        pts.id,
        'Set ' || pts.parameter_name || ' = ' || pts.recommended_value,
        pts.reasoning || ' (current: ' || pts.current_value || ')',
        pts.estimated_improvement_percent,
        pts.confidence_score,
        v_urgency_score,
        ROUND((pts.confidence_score * LEAST(pts.estimated_improvement_percent / 100.0, 1.0) * v_urgency_score)::NUMERIC, 4)::FLOAT,
        CASE
            WHEN pts.parameter_name = 'LIMIT' THEN 'low'::VARCHAR
            WHEN pts.parameter_name = 'work_mem' THEN 'medium'::VARCHAR
            WHEN pts.parameter_name = 'sort_mem' THEN 'medium'::VARCHAR
            WHEN pts.parameter_name = 'batch_size' THEN 'high'::VARCHAR
            ELSE 'medium'::VARCHAR
        END,
        NOW(),
        NOW()
    FROM parameter_tuning_suggestions pts
    WHERE pts.query_hash = p_query_hash
    ON CONFLICT DO NOTHING;

    GET DIAGNOSTICS v_count = v_count + ROW_COUNT;

    -- Step 5: Return summary
    RETURN QUERY
    SELECT
        COUNT(*)::INT,
        ARRAY_AGG(DISTINCT source_type ORDER BY source_type)
    FROM optimization_recommendations
    WHERE query_hash = p_query_hash
    AND created_at > NOW() - INTERVAL '1 minute';

END;
$$ LANGUAGE plpgsql;

-- Function to record recommendation implementation
-- Captures pre-optimization metrics and creates implementation record
CREATE OR REPLACE FUNCTION record_recommendation_implementation(
    p_recommendation_id BIGINT,
    p_query_hash BIGINT,
    p_implementation_notes TEXT DEFAULT NULL
) RETURNS TABLE(
    impl_id BIGINT,
    status VARCHAR,
    pre_snapshot JSONB
) AS $$
DECLARE
    v_impl_id BIGINT;
    v_pre_snapshot JSONB;
BEGIN
    -- Step 1: Get current query metrics as pre-optimization snapshot
    SELECT jsonb_build_object(
        'mean_exec_time_ms', mpsq.mean_exec_time_ms,
        'calls', mpsq.calls,
        'calls_per_sec', mpsq.calls_per_sec,
        'total_time_ms', ROUND((mpsq.mean_exec_time_ms * mpsq.calls)::NUMERIC, 2),
        'rows', mpsq.rows,
        'p95_exec_time_ms', mpsq.p95_exec_time_ms,
        'p99_exec_time_ms', mpsq.p99_exec_time_ms,
        'collected_at', mpsq.collected_at
    )
    INTO v_pre_snapshot
    FROM metrics_pg_stats_query mpsq
    WHERE mpsq.query_hash = p_query_hash
    ORDER BY mpsq.collected_at DESC
    LIMIT 1;

    -- Step 2: Create implementation record
    INSERT INTO optimization_implementations (
        recommendation_id,
        query_hash,
        implementation_notes,
        pre_optimization_stats,
        status
    )
    VALUES (
        p_recommendation_id,
        p_query_hash,
        p_implementation_notes,
        v_pre_snapshot,
        'pending'
    )
    RETURNING optimization_implementations.id
    INTO v_impl_id;

    -- Step 3: Mark recommendation as dismissed (don't show again)
    UPDATE optimization_recommendations
    SET is_dismissed = TRUE
    WHERE id = p_recommendation_id;

    -- Step 4: Return results
    RETURN QUERY
    SELECT v_impl_id, 'pending'::VARCHAR, v_pre_snapshot;

END;
$$ LANGUAGE plpgsql;

-- Function to measure implementation results
-- Compares pre/post metrics and calculates actual improvement
CREATE OR REPLACE FUNCTION measure_implementation_results(
    p_implementation_id BIGINT
) RETURNS TABLE(
    impl_id BIGINT,
    actual_improvement_percent FLOAT,
    predicted_improvement_percent FLOAT,
    status VARCHAR,
    accuracy_score FLOAT
) AS $$
DECLARE
    v_pre_mean_time FLOAT;
    v_post_mean_time FLOAT;
    v_actual_improvement FLOAT;
    v_predicted_improvement FLOAT;
    v_accuracy FLOAT;
    v_query_hash BIGINT;
    v_recommendation_id BIGINT;
    v_status VARCHAR;
BEGIN
    -- Step 1: Get implementation details
    SELECT
        oi.query_hash,
        oi.recommendation_id,
        (oi.pre_optimization_stats->>'mean_exec_time_ms')::FLOAT,
        oi.status
    INTO v_query_hash, v_recommendation_id, v_pre_mean_time, v_status
    FROM optimization_implementations oi
    WHERE oi.id = p_implementation_id;

    -- Step 2: Get current (post-optimization) metrics
    SELECT mpsq.mean_exec_time_ms
    INTO v_post_mean_time
    FROM metrics_pg_stats_query mpsq
    WHERE mpsq.query_hash = v_query_hash
    ORDER BY mpsq.collected_at DESC
    LIMIT 1;

    -- Step 3: Calculate actual improvement
    v_actual_improvement := CASE
        WHEN v_pre_mean_time > 0 THEN
            GREATEST(0.0, (v_pre_mean_time - v_post_mean_time) / v_pre_mean_time * 100.0)
        ELSE 0.0
    END;

    -- Step 4: Get predicted improvement from recommendation
    SELECT orec.estimated_improvement_percent
    INTO v_predicted_improvement
    FROM optimization_recommendations orec
    WHERE orec.id = v_recommendation_id;

    -- Step 5: Calculate accuracy score
    v_accuracy := CASE
        WHEN v_predicted_improvement > 0 THEN
            GREATEST(0.0, 1.0 - ABS(v_actual_improvement - v_predicted_improvement) / v_predicted_improvement)
        ELSE 0.5
    END;

    -- Step 6: Update implementation record
    UPDATE optimization_implementations
    SET
        post_optimization_stats = jsonb_build_object(
            'mean_exec_time_ms', v_post_mean_time,
            'actual_improvement_percent', ROUND(v_actual_improvement::NUMERIC, 2),
            'measured_at', NOW()
        ),
        actual_improvement_percent = v_actual_improvement,
        status = CASE
            WHEN v_actual_improvement > 0 THEN 'implemented'
            ELSE 'no_improvement'
        END,
        measured_at = NOW()
    WHERE id = p_implementation_id;

    -- Step 7: Return results
    RETURN QUERY
    SELECT
        p_implementation_id,
        ROUND(v_actual_improvement::NUMERIC, 2)::FLOAT,
        ROUND(v_predicted_improvement::NUMERIC, 2)::FLOAT,
        CASE
            WHEN v_actual_improvement > 0 THEN 'implemented'
            ELSE 'no_improvement'
        END::VARCHAR,
        ROUND(v_accuracy::NUMERIC, 4)::FLOAT;

END;
$$ LANGUAGE plpgsql;

-- Function to get top recommendations across all queries
-- Ordered by ROI score (highest first)
CREATE OR REPLACE FUNCTION get_top_recommendations(
    p_limit INTEGER DEFAULT 20,
    p_min_impact FLOAT DEFAULT 5.0
) RETURNS TABLE(
    rec_id BIGINT,
    query_hash BIGINT,
    source_type VARCHAR,
    recommendation_text TEXT,
    estimated_improvement FLOAT,
    confidence FLOAT,
    urgency_score FLOAT,
    roi_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        orec.id,
        orec.query_hash,
        orec.source_type,
        orec.recommendation_text,
        orec.estimated_improvement_percent,
        orec.confidence_score,
        orec.urgency_score,
        orec.roi_score
    FROM optimization_recommendations orec
    WHERE orec.is_dismissed = FALSE
    AND orec.estimated_improvement_percent >= p_min_impact
    ORDER BY orec.roi_score DESC, orec.confidence_score DESC
    LIMIT p_limit;

END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 9. Grants and Permissions (if needed)
-- ============================================================================

-- Grant appropriate permissions to application role
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO app_role;

-- ============================================================================
-- 10. Indices for Performance
-- ============================================================================

-- Additional indices for complex queries
CREATE INDEX IF NOT EXISTS idx_optimization_rec_urgency ON optimization_recommendations(urgency_score DESC);
CREATE INDEX IF NOT EXISTS idx_impl_with_measurement ON optimization_implementations(measured_at DESC) WHERE status = 'implemented' AND measured_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rewrite_not_dismissed ON query_rewrite_suggestions(confidence_score DESC) WHERE dismissed = FALSE;
CREATE INDEX IF NOT EXISTS idx_param_tuning_active ON parameter_tuning_suggestions(created_at DESC) WHERE created_at > NOW() - INTERVAL '30 days';

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Summary of new tables:
-- 1. workload_patterns - Detected recurring patterns in query execution
-- 2. query_rewrite_suggestions - SQL rewrite recommendations
-- 3. parameter_tuning_suggestions - Parameter optimization recommendations
-- 4. optimization_recommendations - Aggregated recommendations with ROI scoring
-- 5. optimization_implementations - Track implementation and measure results
-- 6. query_performance_models - Store trained ML models

-- Views:
-- 1. v_top_optimization_recommendations - Top opportunities ranked by ROI
-- 2. v_optimization_results - Implementation results and measured improvements
-- 3. v_workload_pattern_summary - Pattern occurrences and statistics

-- Functions:
-- 1. detect_workload_patterns() - Analyze patterns from historical data
-- 2. calculate_roi_score() - Compute ROI for recommendations
-- 3. calculate_urgency_score() - Compute urgency from query impact
-- 4. get_top_recommendations_for_query() - Get recommendations for specific query
-- 5. record_optimization_implementation() - Record when optimization is applied
-- 6. update_implementation_results() - Update with actual improvement measurements
-- 7. create_pattern_metadata() - Generate pattern metadata JSON
