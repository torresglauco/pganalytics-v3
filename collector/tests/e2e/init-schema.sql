-- E2E Test Schema Initialization
-- Initializes PostgreSQL database for E2E testing

CREATE SCHEMA IF NOT EXISTS pganalytics;

-- Collector Registry Table
CREATE TABLE IF NOT EXISTS pganalytics.collector_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    hostname VARCHAR(255),
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'active'
);

-- API Tokens Table
CREATE TABLE IF NOT EXISTS pganalytics.api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id VARCHAR(255) REFERENCES pganalytics.collector_registry(collector_id) ON DELETE CASCADE,
    token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Collector Configuration Table
CREATE TABLE IF NOT EXISTS pganalytics.collector_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id VARCHAR(255) REFERENCES pganalytics.collector_registry(collector_id) ON DELETE CASCADE,
    config_toml TEXT,
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_collector_id ON pganalytics.collector_registry(collector_id);
CREATE INDEX IF NOT EXISTS idx_token_collector ON pganalytics.api_tokens(collector_id);
CREATE INDEX IF NOT EXISTS idx_config_collector ON pganalytics.collector_config(collector_id);

-- Grant permissions (if needed)
GRANT ALL PRIVILEGES ON SCHEMA pganalytics TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA pganalytics TO postgres;

