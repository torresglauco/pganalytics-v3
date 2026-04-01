-- pgAnalytics v3.3.0 - Index Advisor Tables
-- This migration creates tables for index recommendation and analysis
-- Part of Wave 2: Advanced Optimization Features

-- Ensure we're using the pganalytics schema
SET search_path TO pganalytics, public;

-- Create index_recommendations table
-- Stores index recommendations with their expected benefits and status
CREATE TABLE index_recommendations (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    column_names TEXT[] NOT NULL,
    index_type VARCHAR(50) DEFAULT 'btree',
    estimated_benefit FLOAT NOT NULL,
    weighted_cost_improvement FLOAT NOT NULL,
    status VARCHAR(20) DEFAULT 'recommended',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, table_name, column_names)
);

-- Create index_analysis table
-- Stores detailed analysis of index candidates per query
CREATE TABLE index_analysis (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_id BIGINT REFERENCES query_plans(id) ON DELETE SET NULL,
    index_candidate VARCHAR(255),
    cost_without_index FLOAT,
    cost_with_index FLOAT,
    benefit_score FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create unused_indexes table
-- Stores information about indexes that are not being used
CREATE TABLE unused_indexes (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    index_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    size_bytes BIGINT,
    last_used TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for optimal query performance
CREATE INDEX idx_recommendations_database ON index_recommendations(database_id);
CREATE INDEX idx_recommendations_status ON index_recommendations(status);
CREATE INDEX idx_analysis_database ON index_analysis(database_id);
CREATE INDEX idx_analysis_query ON index_analysis(query_id);
CREATE INDEX idx_analysis_timestamp ON index_analysis(created_at);
CREATE INDEX idx_unused_indexes_database ON unused_indexes(database_id);
CREATE INDEX idx_unused_indexes_table ON unused_indexes(table_name);
