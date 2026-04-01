-- pgAnalytics v3.4.0 - VACUUM Advisor Tables
-- This migration creates tables for VACUUM recommendations and autovacuum tuning
-- Part of Wave 3: Advanced Maintenance Features

-- Ensure we're using the pganalytics schema
SET search_path TO pganalytics, public;

-- Create vacuum_recommendations table
-- Stores VACUUM recommendations with analysis results
CREATE TABLE vacuum_recommendations (
    id BIGSERIAL PRIMARY KEY,
    database_id BIGINT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    table_size BIGINT,
    dead_tuples_count BIGINT,
    dead_tuples_ratio DECIMAL(5, 2),
    autovacuum_enabled BOOLEAN DEFAULT true,
    autovacuum_naptime INTERVAL,
    last_vacuum TIMESTAMP,
    last_autovacuum TIMESTAMP,
    recommendation_type VARCHAR(50) DEFAULT 'full_vacuum',
    estimated_gain DECIMAL(15, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, table_name, created_at)
);

-- Create autovacuum_configurations table
-- Stores current and recommended autovacuum settings
CREATE TABLE autovacuum_configurations (
    id BIGSERIAL PRIMARY KEY,
    database_id BIGINT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    setting_name VARCHAR(255) NOT NULL,
    current_value VARCHAR(500),
    recommended_value VARCHAR(500),
    impact VARCHAR(20) DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, table_name, setting_name)
);

-- Create indexes for optimal query performance
CREATE INDEX idx_vacuum_recommendations_database ON vacuum_recommendations(database_id);
CREATE INDEX idx_vacuum_recommendations_table ON vacuum_recommendations(database_id, table_name);
CREATE INDEX idx_vacuum_recommendations_type ON vacuum_recommendations(recommendation_type);
CREATE INDEX idx_vacuum_recommendations_created ON vacuum_recommendations(created_at DESC);
CREATE INDEX idx_vacuum_recommendations_ratio ON vacuum_recommendations(dead_tuples_ratio DESC);
CREATE INDEX idx_autovacuum_configs_database ON autovacuum_configurations(database_id);
CREATE INDEX idx_autovacuum_configs_table ON autovacuum_configurations(database_id, table_name);
CREATE INDEX idx_autovacuum_configs_setting ON autovacuum_configurations(setting_name);
