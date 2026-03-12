-- ============================================================================
-- MIGRATION: 017_anomaly_detection.sql
-- PURPOSE: Schema for Phase 5 Anomaly Detection & Alerting System
-- VERSION: 1.0
-- DATE: 2026-03-05
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gin;

-- ============================================================================
-- ANOMALY DETECTION SCHEMA
-- ============================================================================

-- Query Baselines: Statistical baseline metrics for each query
CREATE TABLE IF NOT EXISTS query_baselines (
    id BIGSERIAL PRIMARY KEY,
    database_id INTEGER NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_id INTEGER NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    metric_name VARCHAR(255) NOT NULL,

    -- Baseline statistics (calculated every hour)
    baseline_mean NUMERIC(15, 4) NOT NULL,
    baseline_stddev NUMERIC(15, 4) NOT NULL,
    baseline_min NUMERIC(15, 4) NOT NULL,
    baseline_max NUMERIC(15, 4) NOT NULL,
    baseline_median NUMERIC(15, 4) NOT NULL,

    -- Percentiles for better anomaly detection
    baseline_p25 NUMERIC(15, 4),
    baseline_p75 NUMERIC(15, 4),
    baseline_p90 NUMERIC(15, 4),
    baseline_p95 NUMERIC(15, 4),
    baseline_p99 NUMERIC(15, 4),

    -- Baseline period info
    baseline_window_hours INTEGER NOT NULL DEFAULT 168, -- 7 days rolling window
    baseline_data_points INTEGER NOT NULL DEFAULT 0,
    baseline_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Metadata
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_query_baseline UNIQUE(database_id, query_id, metric_name),
    INDEX idx_baseline_query_id (query_id),
    INDEX idx_baseline_database_id (database_id),
    INDEX idx_baseline_metric (metric_name),
    INDEX idx_baseline_updated (updated_at)
);

-- Query Anomalies: Detected anomalies for each query metric
CREATE TABLE IF NOT EXISTS query_anomalies (
    id BIGSERIAL PRIMARY KEY,
    database_id INTEGER NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_id INTEGER NOT NULL REFERENCES queries(id) ON DELETE CASCADE,
    baseline_id BIGINT NOT NULL REFERENCES query_baselines(id) ON DELETE CASCADE,

    -- Anomaly details
    metric_name VARCHAR(255) NOT NULL,
    current_value NUMERIC(15, 4) NOT NULL,
    baseline_value NUMERIC(15, 4) NOT NULL,

    -- Z-score calculation: (value - mean) / stddev
    z_score NUMERIC(10, 2) NOT NULL,
    deviation_percent NUMERIC(10, 2) NOT NULL, -- (current - baseline) / baseline * 100

    -- Severity classification
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),

    -- Detection info
    anomaly_type VARCHAR(50) NOT NULL, -- 'statistical', 'trend', 'seasonal', 'pattern'
    detection_method VARCHAR(100), -- Algorithm used

    -- State tracking
    is_active BOOLEAN DEFAULT TRUE,
    acknowledged_at TIMESTAMP,
    acknowledged_by_user_id INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,

    -- Timestamps
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_severity_level CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    INDEX idx_anomaly_query_id (query_id),
    INDEX idx_anomaly_database_id (database_id),
    INDEX idx_anomaly_severity (severity),
    INDEX idx_anomaly_active (is_active),
    INDEX idx_anomaly_detected (detected_at),
    INDEX idx_anomaly_baseline (baseline_id)
);

-- System Metrics Baselines: For system-level anomalies
CREATE TABLE IF NOT EXISTS system_metrics_baselines (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(100) NOT NULL, -- 'cpu', 'memory', 'disk_io', 'network', 'connections'
    metric_name VARCHAR(255) NOT NULL,

    -- Baseline statistics
    baseline_mean NUMERIC(15, 4) NOT NULL,
    baseline_stddev NUMERIC(15, 4) NOT NULL,
    baseline_min NUMERIC(15, 4) NOT NULL,
    baseline_max NUMERIC(15, 4) NOT NULL,

    -- Configuration
    baseline_window_hours INTEGER DEFAULT 168,
    baseline_data_points INTEGER DEFAULT 0,
    baseline_calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Thresholds
    warning_threshold NUMERIC(15, 4),
    critical_threshold NUMERIC(15, 4),

    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_system_metric_baseline UNIQUE(metric_type, metric_name),
    INDEX idx_system_baseline_type (metric_type)
);

-- System Anomalies: Detected system-level anomalies
CREATE TABLE IF NOT EXISTS system_anomalies (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(100) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    baseline_id BIGINT REFERENCES system_metrics_baselines(id) ON DELETE SET NULL,

    -- Anomaly details
    current_value NUMERIC(15, 4) NOT NULL,
    baseline_value NUMERIC(15, 4),
    z_score NUMERIC(10, 2),

    -- Severity
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),

    -- Detection
    anomaly_type VARCHAR(50),
    detection_method VARCHAR(100),

    -- State
    is_active BOOLEAN DEFAULT TRUE,
    acknowledged_at TIMESTAMP,
    acknowledged_by_user_id INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,

    -- Timestamps
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_system_anomaly_type (metric_type),
    INDEX idx_system_anomaly_severity (severity),
    INDEX idx_system_anomaly_active (is_active),
    INDEX idx_system_anomaly_detected (detected_at)
);

-- ============================================================================
-- ALERT RULES SCHEMA
-- ============================================================================

-- Alert Rules: User-defined rules for triggering alerts
CREATE TABLE IF NOT EXISTS alert_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Rule metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL, -- 'threshold', 'change', 'anomaly', 'composite'

    -- Target scope
    database_id INTEGER REFERENCES databases(id) ON DELETE CASCADE,
    query_id INTEGER REFERENCES queries(id) ON DELETE CASCADE,
    metric_name VARCHAR(255),

    -- Rule condition (JSON stored for flexibility)
    condition JSONB NOT NULL,
    -- Example: {"type": "threshold", "metric": "execution_time", "operator": ">", "value": 1000, "unit": "ms"}
    -- Example: {"type": "anomaly", "severity": "high"}
    -- Example: {"type": "composite", "operator": "AND", "rules": [{"type": "threshold"...}]}

    -- Severity when triggered
    alert_severity VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (alert_severity IN ('low', 'medium', 'high', 'critical')),

    -- Evaluation settings
    evaluation_interval_seconds INTEGER DEFAULT 300, -- Check rule every 5 minutes
    for_duration_seconds INTEGER DEFAULT 0, -- Trigger only if condition true for N seconds (0 = immediate)

    -- Notification settings
    notification_enabled BOOLEAN DEFAULT TRUE,
    notification_channels JSONB, -- Array of channel IDs to notify

    -- State
    is_enabled BOOLEAN DEFAULT TRUE,
    is_paused BOOLEAN DEFAULT FALSE,

    -- Lifecycle
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    CONSTRAINT check_rule_type CHECK (rule_type IN ('threshold', 'change', 'anomaly', 'composite')),
    CONSTRAINT check_alert_severity CHECK (alert_severity IN ('low', 'medium', 'high', 'critical')),
    INDEX idx_rule_user_id (user_id),
    INDEX idx_rule_database_id (database_id),
    INDEX idx_rule_query_id (query_id),
    INDEX idx_rule_enabled (is_enabled),
    INDEX idx_rule_created (created_at)
);

-- Alert Rule Evaluations: Track rule evaluation history
CREATE TABLE IF NOT EXISTS alert_rule_evaluations (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,

    -- Evaluation result
    condition_met BOOLEAN NOT NULL,
    current_value NUMERIC(15, 4),
    threshold_value NUMERIC(15, 4),

    -- Metadata
    evaluation_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER, -- How long evaluation took
    error_message TEXT,

    INDEX idx_eval_rule_id (rule_id),
    INDEX idx_eval_timestamp (evaluation_timestamp),
    CONSTRAINT idx_eval_condition CHECK (condition_met IN (TRUE, FALSE))
);

-- ============================================================================
-- ALERTS SCHEMA
-- ============================================================================

-- Alerts: Fired alerts from rules
CREATE TABLE IF NOT EXISTS alerts (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    anomaly_id BIGINT REFERENCES query_anomalies(id) ON DELETE SET NULL,

    -- Alert info
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),

    -- Context
    database_id INTEGER REFERENCES databases(id),
    query_id INTEGER REFERENCES queries(id),

    -- Alert context data (JSON for flexibility)
    context JSONB,
    -- Example: {"metric": "execution_time", "current_value": 2500, "threshold": 1000, "unit": "ms"}

    -- State machine
    status VARCHAR(50) NOT NULL DEFAULT 'firing' CHECK (status IN ('firing', 'alerting', 'resolved', 'acknowledged')),

    -- Lifecycle
    fired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    acknowledged_by_user_id INTEGER REFERENCES users(id),
    acknowledgment_note TEXT,

    -- Notification tracking
    notification_count INTEGER DEFAULT 0,
    last_notified_at TIMESTAMP,

    -- Deduplication
    fingerprint VARCHAR(64), -- Hash of rule_id + database_id + severity for grouping

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_alert_status CHECK (status IN ('firing', 'alerting', 'resolved', 'acknowledged')),
    CONSTRAINT check_alert_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    INDEX idx_alert_rule_id (rule_id),
    INDEX idx_alert_status (status),
    INDEX idx_alert_severity (severity),
    INDEX idx_alert_fingerprint (fingerprint),
    INDEX idx_alert_fired (fired_at),
    INDEX idx_alert_resolved (resolved_at)
);

-- Alert History: Immutable log of alert state changes
CREATE TABLE IF NOT EXISTS alert_history (
    id BIGSERIAL PRIMARY KEY,
    alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,

    -- State transition
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,

    -- Change info
    changed_by_user_id INTEGER REFERENCES users(id),
    change_reason TEXT,

    -- Timestamp
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_history_alert_id (alert_id),
    INDEX idx_history_changed (changed_at)
);

-- ============================================================================
-- NOTIFICATION CHANNELS
-- ============================================================================

-- Notification Channels: User's configured notification destinations
CREATE TABLE IF NOT EXISTS notification_channels (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Channel info
    name VARCHAR(255) NOT NULL,
    channel_type VARCHAR(50) NOT NULL, -- 'slack', 'email', 'webhook', 'pagerduty', 'jira'

    -- Configuration (encrypted in application layer)
    config JSONB NOT NULL,
    -- Examples:
    -- Slack: {"webhook_url": "...", "channel": "#alerts"}
    -- Email: {"recipients": ["user@example.com"]}
    -- Webhook: {"url": "...", "method": "POST", "headers": {...}}
    -- PagerDuty: {"integration_key": "..."}
    -- Jira: {"url": "...", "project_key": "...", "auth_token": "..."}

    -- Verification
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verified_at TIMESTAMP,

    -- State
    is_enabled BOOLEAN DEFAULT TRUE,

    -- Testing
    last_test_at TIMESTAMP,
    last_test_status VARCHAR(20), -- 'success', 'failed', 'pending'
    last_test_error TEXT,

    -- Lifecycle
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_channel_type CHECK (channel_type IN ('slack', 'email', 'webhook', 'pagerduty', 'jira')),
    INDEX idx_channel_user_id (user_id),
    INDEX idx_channel_type (channel_type),
    INDEX idx_channel_enabled (is_enabled)
);

-- Notification Deliveries: Track sent notifications
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id BIGSERIAL PRIMARY KEY,
    alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,

    -- Delivery info
    delivery_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'failed', 'bounced'
    delivery_attempts INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 5,

    -- Message
    message_subject VARCHAR(255),
    message_body TEXT,

    -- Results
    delivered_at TIMESTAMP,
    last_error TEXT,

    -- Retry info
    next_retry_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_delivery_status CHECK (delivery_status IN ('pending', 'sent', 'failed', 'bounced')),
    INDEX idx_delivery_alert_id (alert_id),
    INDEX idx_delivery_channel_id (channel_id),
    INDEX idx_delivery_status (delivery_status),
    INDEX idx_delivery_next_retry (next_retry_at)
);

-- ============================================================================
-- FUNCTIONS FOR ANOMALY DETECTION
-- ============================================================================

-- Function: Calculate baseline statistics for a query metric
CREATE OR REPLACE FUNCTION calculate_query_baseline(
    p_database_id INTEGER,
    p_query_id INTEGER,
    p_metric_name VARCHAR,
    p_window_hours INTEGER DEFAULT 168
)
RETURNS TABLE(
    mean_value NUMERIC,
    stddev_value NUMERIC,
    min_value NUMERIC,
    max_value NUMERIC,
    median_value NUMERIC,
    p25_value NUMERIC,
    p75_value NUMERIC,
    p90_value NUMERIC,
    p95_value NUMERIC,
    p99_value NUMERIC,
    data_points INTEGER
) AS $$
DECLARE
    v_metric_column VARCHAR;
BEGIN
    -- Map metric names to columns in query_history
    SELECT CASE
        WHEN p_metric_name = 'execution_time' THEN 'execution_time_ms'
        WHEN p_metric_name = 'calls' THEN 'calls'
        WHEN p_metric_name = 'rows' THEN 'rows_returned'
        WHEN p_metric_name = 'rows_affected' THEN 'rows_affected'
        WHEN p_metric_name = 'mean_time' THEN 'mean_exec_time'
        ELSE p_metric_name
    END INTO v_metric_column;

    RETURN QUERY EXECUTE format(
        'SELECT
            avg(%I)::NUMERIC as mean_value,
            stddev_pop(%I)::NUMERIC as stddev_value,
            min(%I)::NUMERIC as min_value,
            max(%I)::NUMERIC as max_value,
            percentile_cont(0.5) WITHIN GROUP (ORDER BY %I)::NUMERIC as median_value,
            percentile_cont(0.25) WITHIN GROUP (ORDER BY %I)::NUMERIC as p25_value,
            percentile_cont(0.75) WITHIN GROUP (ORDER BY %I)::NUMERIC as p75_value,
            percentile_cont(0.90) WITHIN GROUP (ORDER BY %I)::NUMERIC as p90_value,
            percentile_cont(0.95) WITHIN GROUP (ORDER BY %I)::NUMERIC as p95_value,
            percentile_cont(0.99) WITHIN GROUP (ORDER BY %I)::NUMERIC as p99_value,
            count(*)::INTEGER as data_points
        FROM query_history
        WHERE database_id = $1
            AND query_id = $2
            AND collected_at > NOW() - INTERVAL ''1 hour'' * $3
            AND %I IS NOT NULL',
        v_metric_column, v_metric_column, v_metric_column, v_metric_column,
        v_metric_column, v_metric_column, v_metric_column, v_metric_column,
        v_metric_column, v_metric_column, v_metric_column
    )
    USING p_database_id, p_query_id, p_window_hours;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function: Detect anomalies using Z-score method
CREATE OR REPLACE FUNCTION detect_anomalies_zscore(
    p_database_id INTEGER,
    p_threshold NUMERIC DEFAULT 2.5
)
RETURNS TABLE(
    query_id INTEGER,
    metric_name VARCHAR,
    current_value NUMERIC,
    baseline_value NUMERIC,
    z_score NUMERIC,
    deviation_percent NUMERIC,
    severity VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        qb.query_id,
        qb.metric_name,
        qh.execution_time_ms::NUMERIC,
        qb.baseline_mean,
        CASE
            WHEN qb.baseline_stddev = 0 THEN 0
            ELSE ((qh.execution_time_ms - qb.baseline_mean) / qb.baseline_stddev)::NUMERIC
        END as z_score,
        CASE
            WHEN qb.baseline_mean = 0 THEN 0
            ELSE (((qh.execution_time_ms - qb.baseline_mean) / qb.baseline_mean) * 100)::NUMERIC
        END as deviation_percent,
        CASE
            WHEN ABS((qh.execution_time_ms - qb.baseline_mean) / NULLIF(qb.baseline_stddev, 0)) > p_threshold * 2 THEN 'critical'
            WHEN ABS((qh.execution_time_ms - qb.baseline_mean) / NULLIF(qb.baseline_stddev, 0)) > p_threshold THEN 'high'
            WHEN ABS((qh.execution_time_ms - qb.baseline_mean) / NULLIF(qb.baseline_stddev, 0)) > 1.5 THEN 'medium'
            ELSE 'low'
        END as severity
    FROM query_history qh
    JOIN query_baselines qb ON qh.query_id = qb.query_id AND qb.database_id = qh.database_id
    WHERE qh.database_id = p_database_id
        AND qh.collected_at > NOW() - INTERVAL '1 hour'
        AND qb.is_enabled = TRUE
        AND ABS((qh.execution_time_ms - qb.baseline_mean) / NULLIF(qb.baseline_stddev, 0)) > 1.5;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_anomalies_query_severity
    ON query_anomalies(database_id, query_id, severity)
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_alerts_status_severity
    ON alerts(status, severity)
    WHERE status IN ('firing', 'alerting');

CREATE INDEX IF NOT EXISTS idx_notifications_status_channel
    ON notification_deliveries(channel_id, delivery_status)
    WHERE delivery_status IN ('pending', 'failed');

-- GIN index for JSONB queries
CREATE INDEX IF NOT EXISTS idx_rule_condition_gin
    ON alert_rules USING GIN(condition);

CREATE INDEX IF NOT EXISTS idx_alert_context_gin
    ON alerts USING GIN(context);

-- ============================================================================
-- TRIGGER FUNCTIONS
-- ============================================================================

-- Trigger: Update updated_at on query_baselines
CREATE OR REPLACE FUNCTION update_query_baselines_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_query_baselines_updated_at
    BEFORE UPDATE ON query_baselines
    FOR EACH ROW
    EXECUTE FUNCTION update_query_baselines_updated_at();

-- Trigger: Update updated_at on alert_rules
CREATE OR REPLACE FUNCTION update_alert_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_rules_updated_at();

-- Trigger: Update updated_at on alerts
CREATE OR REPLACE FUNCTION update_alerts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_alerts_updated_at();

-- Trigger: Log alert state changes to alert_history
CREATE OR REPLACE FUNCTION log_alert_state_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status != OLD.status THEN
        INSERT INTO alert_history(alert_id, previous_status, new_status, changed_at)
        VALUES(NEW.id, OLD.status, NEW.status, CURRENT_TIMESTAMP);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_alert_state_change
    AFTER UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION log_alert_state_change();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE query_baselines IS 'Statistical baseline metrics for queries used in anomaly detection';
COMMENT ON TABLE query_anomalies IS 'Detected anomalies in query behavior with severity classification';
COMMENT ON TABLE alert_rules IS 'User-defined alert rules that trigger based on conditions';
COMMENT ON TABLE alerts IS 'Fired alerts with state management and notification tracking';
COMMENT ON TABLE notification_channels IS 'User-configured notification destinations (Slack, Email, etc)';
COMMENT ON TABLE notification_deliveries IS 'Delivery tracking for sent notifications with retry logic';

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================

-- Track migration completion
INSERT INTO schema_migrations(name, applied_at) VALUES('017_anomaly_detection', CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
