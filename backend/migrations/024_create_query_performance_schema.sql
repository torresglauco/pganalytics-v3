-- pgAnalytics v3.1.0 - Query Performance Analysis Tables
-- This migration creates tables for analyzing query execution plans, identifying performance issues,
-- and tracking query performance metrics over time

-- Ensure we're using the pganalytics schema
SET search_path TO pganalytics, public;

-- Create query_plans table
-- Stores parsed query execution plans and aggregate statistics
CREATE TABLE query_plans (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_hash VARCHAR(64) NOT NULL,
    query_text TEXT NOT NULL,
    plan_json JSONB NOT NULL,
    mean_time FLOAT NOT NULL,
    total_time FLOAT NOT NULL,
    calls BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, query_hash)
);

-- Create query_issues table
-- Stores identified performance issues within query execution plans
CREATE TABLE query_issues (
    id BIGSERIAL PRIMARY KEY,
    query_plan_id BIGINT NOT NULL REFERENCES query_plans(id) ON DELETE CASCADE,
    issue_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    affected_node_id INT,
    description TEXT,
    recommendation TEXT,
    estimated_benefit FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create query_performance_timeline table
-- Stores time-series metrics for query performance tracking
CREATE TABLE query_performance_timeline (
    id BIGSERIAL PRIMARY KEY,
    query_plan_id BIGINT NOT NULL REFERENCES query_plans(id) ON DELETE CASCADE,
    metric_timestamp TIMESTAMP NOT NULL,
    avg_duration FLOAT NOT NULL,
    max_duration FLOAT NOT NULL,
    executions BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for optimal query performance
CREATE INDEX idx_query_plans_database ON query_plans(database_id);
CREATE INDEX idx_query_issues_query_plan ON query_issues(query_plan_id);
CREATE INDEX idx_timeline_query_timestamp ON query_performance_timeline(query_plan_id, metric_timestamp);
