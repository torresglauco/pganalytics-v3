-- pgAnalytics v3.0 - Complete Database Schema
-- COMPREHENSIVE SCHEMA THAT ENSURES ALL TABLES EXIST
-- This migration runs first (000_) and creates all necessary infrastructure
-- Subsequent migrations build on top of this foundation

SET search_path TO pganalytics, public;

-- ============================================================================
-- EXTENSIONS & SCHEMA SETUP
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For text search
CREATE EXTENSION IF NOT EXISTS "btree_gin";  -- For composite indexes

-- Create pganalytics schema
CREATE SCHEMA IF NOT EXISTS pganalytics;

-- ============================================================================
-- SCHEMA MIGRATION TRACKING (MUST EXIST FIRST)
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_versions (
    version VARCHAR(100) PRIMARY KEY,
    description TEXT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER
);

-- ============================================================================
-- USERS & AUTHENTICATION
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer')),
    is_active BOOLEAN DEFAULT true,
    password_changed BOOLEAN DEFAULT false,  -- Track if user has changed password from setup
    last_login TIMESTAMP WITH TIME ZONE,
    last_password_change TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_password_changed ON users(password_changed) WHERE password_changed = false;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ============================================================================
-- API TOKENS
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    description TEXT,
    last_used TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_tokens_user ON api_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_expires ON api_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_api_tokens_hash ON api_tokens(token_hash);

-- ============================================================================
-- COLLECTORS (Monitoring Agents)
-- ============================================================================

CREATE TABLE IF NOT EXISTS collectors (
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
    health_check_interval INTEGER DEFAULT 300,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_collectors_hostname ON collectors(hostname);
CREATE INDEX IF NOT EXISTS idx_collectors_status ON collectors(status);
CREATE INDEX IF NOT EXISTS idx_collectors_last_seen ON collectors(last_seen);
CREATE INDEX IF NOT EXISTS idx_collectors_active ON collectors(status) WHERE status = 'active';

-- ============================================================================
-- COLLECTOR REGISTRATION TOKENS
-- ============================================================================

CREATE TABLE IF NOT EXISTS collector_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    token_type VARCHAR(50) DEFAULT 'access',  -- access, refresh, registration
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_collector_tokens_collector ON collector_tokens(collector_id);
CREATE INDEX IF NOT EXISTS idx_collector_tokens_hash ON collector_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_collector_tokens_expires ON collector_tokens(expires_at);

-- ============================================================================
-- REGISTRATION SECRETS (For Collector Self-Registration)
-- ============================================================================

CREATE TABLE IF NOT EXISTS registration_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    secret_value VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    total_registrations INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    max_registrations INTEGER  -- NULL = unlimited
);

CREATE INDEX IF NOT EXISTS idx_registration_secrets_secret_value ON registration_secrets(secret_value) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_registration_secrets_active ON registration_secrets(active);
CREATE INDEX IF NOT EXISTS idx_registration_secrets_expires ON registration_secrets(expires_at);

-- Registration secret audit table
CREATE TABLE IF NOT EXISTS registration_secret_audit (
    id BIGSERIAL PRIMARY KEY,
    secret_id UUID REFERENCES registration_secrets(id) ON DELETE CASCADE,
    collector_id UUID REFERENCES collectors(id) ON DELETE SET NULL,
    collector_name VARCHAR(255),
    status VARCHAR(50),  -- 'success', 'failed', 'expired'
    error_message TEXT,
    ip_address VARCHAR(45),
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_registration_secret_audit_secret_id ON registration_secret_audit(secret_id);
CREATE INDEX IF NOT EXISTS idx_registration_secret_audit_used_at ON registration_secret_audit(used_at);

-- ============================================================================
-- COLLECTOR CONFIGURATION
-- ============================================================================

CREATE TABLE IF NOT EXISTS collector_config (
    id SERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE(collector_id, version)
);

CREATE INDEX IF NOT EXISTS idx_collector_config_collector ON collector_config(collector_id, version DESC);

-- ============================================================================
-- MANAGED INSTANCES (RDS/Aurora PostgreSQL)
-- ============================================================================

CREATE TABLE IF NOT EXISTS managed_instances (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    aws_region VARCHAR(50) NOT NULL,
    rds_endpoint VARCHAR(255) NOT NULL UNIQUE,
    port INTEGER DEFAULT 5432,
    engine_version VARCHAR(50),
    db_instance_class VARCHAR(50),
    allocated_storage_gb INTEGER,
    environment VARCHAR(50) DEFAULT 'production' CHECK (environment IN ('production', 'staging', 'development', 'test')),
    master_username VARCHAR(255) NOT NULL,
    secret_id INTEGER,  -- References secrets table
    enable_enhanced_monitoring BOOLEAN DEFAULT false,
    monitoring_interval INTEGER DEFAULT 60,
    ssl_enabled BOOLEAN DEFAULT true,
    ssl_mode VARCHAR(20) DEFAULT 'require',
    connection_timeout INTEGER DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    last_connection_status VARCHAR(50) DEFAULT 'unknown',
    last_error_message TEXT,
    last_error_time TIMESTAMP WITH TIME ZONE,
    multi_az BOOLEAN DEFAULT false,
    backup_retention_days INTEGER,
    preferred_backup_window VARCHAR(100),
    preferred_maintenance_window VARCHAR(100),
    tags JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_managed_instances_region ON managed_instances(aws_region);
CREATE INDEX IF NOT EXISTS idx_managed_instances_active ON managed_instances(is_active);
CREATE INDEX IF NOT EXISTS idx_managed_instances_environment ON managed_instances(environment);
CREATE INDEX IF NOT EXISTS idx_managed_instances_status ON managed_instances(last_connection_status);

-- ============================================================================
-- MANAGED INSTANCE DATABASES
-- ============================================================================

CREATE TABLE IF NOT EXISTS managed_instance_databases (
    id SERIAL PRIMARY KEY,
    managed_instance_id INTEGER NOT NULL REFERENCES managed_instances(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    size_bytes BIGINT,
    is_template BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    last_analyzed TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(managed_instance_id, name)
);

CREATE INDEX IF NOT EXISTS idx_managed_instance_databases_instance ON managed_instance_databases(managed_instance_id);
CREATE INDEX IF NOT EXISTS idx_managed_instance_databases_name ON managed_instance_databases(name);

-- ============================================================================
-- SERVERS
-- ============================================================================

CREATE TABLE IF NOT EXISTS servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    address INET NOT NULL,
    environment VARCHAR(50) DEFAULT 'production',
    collector_id UUID REFERENCES collectors(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_servers_hostname ON servers(hostname);
CREATE INDEX IF NOT EXISTS idx_servers_collector ON servers(collector_id);
CREATE INDEX IF NOT EXISTS idx_servers_active ON servers(is_active);

-- ============================================================================
-- POSTGRESQL INSTANCES
-- ============================================================================

CREATE TABLE IF NOT EXISTS postgresql_instances (
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
    replication_role VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_instances_server ON postgresql_instances(server_id);
CREATE INDEX IF NOT EXISTS idx_instances_active ON postgresql_instances(is_active);

-- ============================================================================
-- DATABASES
-- ============================================================================

CREATE TABLE IF NOT EXISTS databases (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    size_bytes BIGINT,
    is_template BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    last_analyzed TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(instance_id, name)
);

CREATE INDEX IF NOT EXISTS idx_databases_instance ON databases(instance_id);
CREATE INDEX IF NOT EXISTS idx_databases_name ON databases(name);

-- ============================================================================
-- SECRETS (Encrypted)
-- ============================================================================

CREATE TABLE IF NOT EXISTS secrets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    secret_encrypted BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);

-- ============================================================================
-- ALERTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_type VARCHAR(255) NOT NULL,
    condition_type VARCHAR(50),
    condition_value VARCHAR(255),
    severity VARCHAR(50) DEFAULT 'warning',
    enabled BOOLEAN DEFAULT true,
    notification_channel VARCHAR(255),
    evaluation_interval INTEGER DEFAULT 300,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric_type);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE,
    rule_id INTEGER REFERENCES alert_rules(id) ON DELETE CASCADE,
    server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
    database_id INTEGER REFERENCES databases(id) ON DELETE CASCADE,
    metric_type VARCHAR(255),
    metric_value VARCHAR(255),
    severity VARCHAR(50),
    message TEXT,
    is_acknowledged BOOLEAN DEFAULT false,
    acknowledged_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_alerts_collector ON alerts(collector_id);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved_at) WHERE resolved_at IS NULL;

-- ============================================================================
-- AUDIT LOG
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255),
    resource_id VARCHAR(255),
    changes JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- ============================================================================
-- METRIC TYPES (Metadata)
-- ============================================================================

CREATE TABLE IF NOT EXISTS metric_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    schema_definition JSONB,
    retention_days INTEGER DEFAULT 7,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO metric_types (name, description, retention_days) VALUES
    ('pg_stats', 'PostgreSQL table/index/database statistics', 7),
    ('pg_log', 'PostgreSQL server logs', 7),
    ('sysstat', 'System CPU, memory, I/O statistics', 7),
    ('disk_usage', 'Disk space usage per filesystem', 30),
    ('replication', 'Replication lag and status', 7),
    ('locks', 'Lock contention and deadlocks', 7),
    ('connections', 'Connection metrics and activity', 7),
    ('bloat', 'Table and index bloat estimation', 7),
    ('cache_hits', 'Buffer cache hit ratios', 7),
    ('extensions', 'PostgreSQL extension metrics', 7),
    ('schema', 'Schema object metrics', 7),
    ('query_performance', 'Query execution performance', 7)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- TRIGGER FUNCTIONS
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

CREATE TRIGGER IF NOT EXISTS trigger_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_collectors_updated_at BEFORE UPDATE ON collectors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_servers_updated_at BEFORE UPDATE ON servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_postgresql_instances_updated_at BEFORE UPDATE ON postgresql_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_databases_updated_at BEFORE UPDATE ON databases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_secrets_updated_at BEFORE UPDATE ON secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_registration_secrets_updated_at BEFORE UPDATE ON registration_secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_managed_instances_updated_at BEFORE UPDATE ON managed_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER IF NOT EXISTS trigger_managed_instance_databases_updated_at BEFORE UPDATE ON managed_instance_databases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ROLE-BASED ACCESS CONTROL
-- ============================================================================

CREATE ROLE IF NOT EXISTS pganalytics_app_master WITH LOGIN;
CREATE ROLE IF NOT EXISTS pganalytics_app_user WITH LOGIN;
CREATE ROLE IF NOT EXISTS pganalytics_app_readonly WITH LOGIN;

GRANT USAGE ON SCHEMA pganalytics TO pganalytics_app_master, pganalytics_app_user, pganalytics_app_readonly;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_master;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_user;
GRANT SELECT ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT ALL ON TABLES TO pganalytics_app_master;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT, INSERT, UPDATE ON TABLES TO pganalytics_app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT ON TABLES TO pganalytics_app_readonly;

-- ============================================================================
-- RECORD COMPLETION
-- ============================================================================

INSERT INTO schema_versions (version, description, executed_at)
VALUES ('000_complete_schema', 'Complete and comprehensive database schema - creates all necessary tables and infrastructure', NOW())
ON CONFLICT DO NOTHING;
