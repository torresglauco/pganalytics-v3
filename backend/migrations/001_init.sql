-- pgAnalytics v3.0 - Initial Database Schema
-- Main PostgreSQL instance for metadata and user/collector management

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Schema for pgAnalytics
CREATE SCHEMA IF NOT EXISTS pganalytics;
SET search_path TO pganalytics, public;

-- ============================================================================
-- USERS & AUTHENTICATION
-- ============================================================================

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer')),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- ============================================================================
-- COLLECTORS
-- ============================================================================

CREATE TABLE collectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    hostname VARCHAR(255) NOT NULL,
    address INET,
    version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'registered' CHECK (status IN ('registered', 'active', 'offline', 'error', 'deregistered')),
    last_seen TIMESTAMP WITH TIME ZONE,
    certificate_thumbprint VARCHAR(255),
    certificate_expires_at TIMESTAMP WITH TIME ZONE,
    config_version INTEGER DEFAULT 0,
    metrics_count_total BIGINT DEFAULT 0,
    metrics_count_24h BIGINT DEFAULT 0,
    health_check_interval INTEGER DEFAULT 300,  -- 5 minutes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_collectors_hostname ON collectors(hostname);
CREATE INDEX idx_collectors_status ON collectors(status);
CREATE INDEX idx_collectors_last_seen ON collectors(last_seen);

-- ============================================================================
-- COLLECTOR CONFIGURATION
-- ============================================================================

CREATE TABLE collector_config (
    id SERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    config JSONB NOT NULL,  -- Configuration as JSON (intervals, enabled collectors, etc)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER REFERENCES users(id)
);

CREATE INDEX idx_collector_config_collector ON collector_config(collector_id, version DESC);

-- ============================================================================
-- SERVERS & INSTANCES
-- ============================================================================

CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    address INET NOT NULL,
    environment VARCHAR(50) DEFAULT 'production' CHECK (environment IN ('production', 'staging', 'development', 'test')),
    collector_id UUID REFERENCES collectors(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_servers_hostname ON servers(hostname);
CREATE INDEX idx_servers_collector ON servers(collector_id);

CREATE TABLE postgresql_instances (
    id SERIAL PRIMARY KEY,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50),
    port INTEGER DEFAULT 5432,
    connection_string VARCHAR(2048),
    maintenance_database VARCHAR(255) DEFAULT 'postgres',
    monitoring_role VARCHAR(255) DEFAULT 'pganalytics_monitor',
    is_active BOOLEAN DEFAULT true,
    last_connected TIMESTAMP WITH TIME ZONE,
    replication_role VARCHAR(50) CHECK (replication_role IN ('primary', 'standby', 'unknown')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_instances_server ON postgresql_instances(server_id);
CREATE INDEX idx_instances_active ON postgresql_instances(is_active);

CREATE TABLE databases (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    size_bytes BIGINT,
    is_template BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    last_analyzed TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_databases_instance ON databases(instance_id);
CREATE INDEX idx_databases_name ON databases(name);
CREATE UNIQUE INDEX idx_databases_instance_name ON databases(instance_id, name);

-- ============================================================================
-- AUTHENTICATION & SECRETS
-- ============================================================================

CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    description VARCHAR(255),
    last_used TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_tokens_collector ON api_tokens(collector_id);
CREATE INDEX idx_api_tokens_user ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_expires ON api_tokens(expires_at);

CREATE TABLE secrets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    secret_encrypted BYTEA NOT NULL,  -- Encrypted with pgcrypto
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- METRICS METADATA
-- ============================================================================

CREATE TABLE metric_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,  -- e.g., 'pg_stats', 'sysstat', 'pg_log', 'disk_usage'
    description TEXT,
    schema_definition JSONB,  -- JSON schema for validation
    retention_days INTEGER DEFAULT 7,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO metric_types (name, description, retention_days) VALUES
    ('pg_stats', 'PostgreSQL table/index/database statistics', 7),
    ('pg_log', 'PostgreSQL server logs', 7),
    ('sysstat', 'System CPU, memory, I/O statistics', 7),
    ('disk_usage', 'Disk space usage per filesystem', 30),
    ('replication', 'Replication lag and status', 7)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- ALERTS
-- ============================================================================

CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_type VARCHAR(255) NOT NULL,
    condition_type VARCHAR(50) CHECK (condition_type IN ('threshold', 'change', 'anomaly')),
    condition_value VARCHAR(255),
    severity VARCHAR(50) DEFAULT 'warning' CHECK (severity IN ('info', 'warning', 'critical')),
    enabled BOOLEAN DEFAULT true,
    notification_channel VARCHAR(255),
    evaluation_interval INTEGER DEFAULT 300,  -- 5 minutes
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    collector_id UUID REFERENCES collectors(id),
    rule_id INTEGER REFERENCES alert_rules(id),
    server_id INTEGER REFERENCES servers(id),
    database_id INTEGER REFERENCES databases(id),
    metric_type VARCHAR(255),
    metric_value VARCHAR(255),
    severity VARCHAR(50) CHECK (severity IN ('info', 'warning', 'critical')),
    message TEXT,
    is_acknowledged BOOLEAN DEFAULT false,
    acknowledged_by INTEGER REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_alerts_collector ON alerts(collector_id);
CREATE INDEX idx_alerts_server ON alerts(server_id);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_created ON alerts(created_at DESC);
CREATE INDEX idx_alerts_resolved ON alerts(resolved_at) WHERE resolved_at IS NULL;

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255),
    resource_id VARCHAR(255),
    changes JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Create default admin user (password: admin / should be changed on first login)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES (
    'admin',
    'admin@pganalytics.local',
    crypt('admin', gen_salt('bf')),
    'Administrator',
    'admin'
) ON CONFLICT DO NOTHING;

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Function to update 'updated_at' timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_collectors_updated_at BEFORE UPDATE ON collectors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_servers_updated_at BEFORE UPDATE ON servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_postgresql_instances_updated_at BEFORE UPDATE ON postgresql_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_databases_updated_at BEFORE UPDATE ON databases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_secrets_updated_at BEFORE UPDATE ON secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- GRANTS (Role-based access)
-- ============================================================================

-- Create application roles
CREATE ROLE pganalytics_app_master WITH LOGIN;
CREATE ROLE pganalytics_app_user WITH LOGIN;
CREATE ROLE pganalytics_app_readonly WITH LOGIN;

-- Grant schema privileges
GRANT USAGE ON SCHEMA pganalytics TO pganalytics_app_master, pganalytics_app_user, pganalytics_app_readonly;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_master;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_user;
GRANT SELECT ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_readonly;

-- Default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT ALL ON TABLES TO pganalytics_app_master;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT, INSERT, UPDATE ON TABLES TO pganalytics_app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT ON TABLES TO pganalytics_app_readonly;
