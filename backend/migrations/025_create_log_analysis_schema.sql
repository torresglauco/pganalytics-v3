-- Migration: Create Log Analysis Schema (v3.2.0)
-- Description: Creates tables for log analysis including logs, log_patterns, and log_anomalies
-- Tables: logs, log_patterns, log_anomalies

CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    log_timestamp TIMESTAMP NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    duration FLOAT,
    table_affected VARCHAR(255),
    query_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE log_patterns (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    pattern_name VARCHAR(255) NOT NULL,
    pattern_regex TEXT NOT NULL,
    frequency BIGINT DEFAULT 0,
    severity_avg FLOAT,
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, pattern_name)
);

CREATE TABLE log_anomalies (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    pattern_id BIGINT REFERENCES log_patterns(id),
    anomaly_timestamp TIMESTAMP NOT NULL,
    anomaly_score FLOAT NOT NULL,
    deviation_from_baseline FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_logs_database_timestamp ON logs(database_id, log_timestamp DESC);
CREATE INDEX idx_logs_category ON logs(category);
CREATE INDEX idx_log_patterns_database ON log_patterns(database_id);
CREATE INDEX idx_log_anomalies_timestamp ON log_anomalies(anomaly_timestamp DESC);
