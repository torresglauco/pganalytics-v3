-- pgAnalytics v3.0 - Complete Database Schema
-- COMPREHENSIVE SCHEMA THAT ENSURES ALL TABLES EXIST
-- This migration runs first (000_) and creates all necessary infrastructure

-- Extensions
CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION "pgcrypto";
CREATE EXTENSION "pg_trgm";
CREATE EXTENSION "btree_gin";

-- Create schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS pganalytics;

SET search_path TO pganalytics, public;

-- Note: schema_versions table and pganalytics schema are created by the migration runner

-- Users & Authentication
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user', 'viewer')),
    is_active BOOLEAN DEFAULT true,
    password_changed BOOLEAN DEFAULT false,
    last_login TIMESTAMP WITH TIME ZONE,
    last_password_change TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_password_changed ON users(password_changed) WHERE password_changed = false;
CREATE INDEX idx_users_role ON users(role);

-- API Tokens  
CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX idx_api_tokens_user ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_expires ON api_tokens(expires_at);
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash);

-- Collectors (Monitoring Agents)
CREATE TABLE collectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'registered',
    last_seen TIMESTAMP WITH TIME ZONE,
    metrics_count_24h BIGINT,
    config_version INTEGER DEFAULT 0,
    health_check_interval INTEGER DEFAULT 300,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_collectors_hostname ON collectors(hostname);
CREATE INDEX idx_collectors_status ON collectors(status);
CREATE INDEX idx_collectors_last_seen ON collectors(last_seen);
CREATE INDEX idx_collectors_active ON collectors(status) WHERE status != 'offline';

-- Collector Tokens
CREATE TABLE collector_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX idx_collector_tokens_collector ON collector_tokens(collector_id);
CREATE INDEX idx_collector_tokens_hash ON collector_tokens(token_hash);
CREATE INDEX idx_collector_tokens_expires ON collector_tokens(expires_at);

-- Registration Secrets
CREATE TABLE registration_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    secret_value VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    total_registrations INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    max_registrations INTEGER,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_registration_secrets_secret_value ON registration_secrets(secret_value);
CREATE INDEX idx_registration_secrets_active ON registration_secrets(active);
CREATE INDEX idx_registration_secrets_expires ON registration_secrets(expires_at);

-- Registration Secret Audit
CREATE TABLE registration_secret_audit (
    id BIGSERIAL PRIMARY KEY,
    secret_id UUID NOT NULL REFERENCES registration_secrets(id),
    collector_id UUID NOT NULL REFERENCES collectors(id),
    used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_registration_secret_audit_secret_id ON registration_secret_audit(secret_id);
CREATE INDEX idx_registration_secret_audit_used_at ON registration_secret_audit(used_at);

-- Collector Config
CREATE TABLE collector_config (
    id SERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    version INTEGER DEFAULT 1,
    config_json TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_collector_config_collector ON collector_config(collector_id);

-- Managed Instances (RDS/Aurora)
CREATE TABLE managed_instances (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    aws_region VARCHAR(50) NOT NULL,
    rds_endpoint VARCHAR(255) UNIQUE NOT NULL,
    port INTEGER DEFAULT 5432,
    engine_version VARCHAR(50),
    db_instance_class VARCHAR(50),
    ssl_enabled BOOLEAN DEFAULT true,
    ssl_mode VARCHAR(20) DEFAULT 'require',
    is_active BOOLEAN DEFAULT true,
    last_connection_status VARCHAR(50) DEFAULT 'unknown',
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    last_error_message TEXT,
    environment VARCHAR(50) DEFAULT 'production',
    multi_az BOOLEAN DEFAULT false,
    backup_retention_days INTEGER,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_managed_instances_region ON managed_instances(aws_region);
CREATE INDEX idx_managed_instances_active ON managed_instances(is_active);
CREATE INDEX idx_managed_instances_environment ON managed_instances(environment);
CREATE INDEX idx_managed_instances_status ON managed_instances(last_connection_status);

-- Managed Instance Databases
CREATE TABLE managed_instance_databases (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES managed_instances(id) ON DELETE CASCADE,
    database_name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    encoding VARCHAR(50),
    "collation" VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_managed_instance_databases_instance ON managed_instance_databases(instance_id);
CREATE INDEX idx_managed_instance_databases_name ON managed_instance_databases(database_name);

-- Servers
CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    collector_id UUID REFERENCES collectors(id),
    os_type VARCHAR(50),
    os_version VARCHAR(100),
    cpu_count INTEGER,
    memory_bytes BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_servers_hostname ON servers(hostname);
CREATE INDEX idx_servers_collector ON servers(collector_id);
CREATE INDEX idx_servers_active ON servers(is_active);

-- PostgreSQL Instances
CREATE TABLE postgresql_instances (
    id SERIAL PRIMARY KEY,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    port INTEGER NOT NULL,
    version VARCHAR(50),
    data_directory VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_instances_server ON postgresql_instances(server_id);
CREATE INDEX idx_instances_active ON postgresql_instances(is_active);

-- Databases
CREATE TABLE databases (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    encoding VARCHAR(50),
    "collation" VARCHAR(100),
    size_bytes BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_databases_instance ON databases(instance_id);
CREATE INDEX idx_databases_name ON databases(name);

-- Secrets (Encrypted Credentials)
CREATE TABLE secrets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    secret_type VARCHAR(50),
    encrypted_value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_secrets_name ON secrets(name);

-- Alert Rules
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_type VARCHAR(100),
    condition VARCHAR(50),
    threshold NUMERIC(10, 2),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX idx_alert_rules_metric ON alert_rules(metric_type);

-- Alerts
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    alert_rule_id INTEGER REFERENCES alert_rules(id),
    severity VARCHAR(50),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_alerts_collector ON alerts(collector_id);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_created ON alerts(created_at DESC);
CREATE INDEX idx_alerts_resolved ON alerts(resolved_at) WHERE resolved_at IS NULL;

-- Audit Log
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id INTEGER,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_action ON audit_log(action);

-- Metric Types
CREATE TABLE metric_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    retention_days INTEGER DEFAULT 7
);
INSERT INTO metric_types (name, description, retention_days) VALUES
    ('pg_stats', 'PostgreSQL table/index/database statistics', 7),
    ('pg_log', 'PostgreSQL server logs', 3),
    ('system_metrics', 'System CPU, memory, disk metrics', 7),
    ('query_metrics', 'Query performance metrics', 14)
ON CONFLICT DO NOTHING;

-- Role-Based Access Control
CREATE ROLE pganalytics_app_master WITH LOGIN;
CREATE ROLE pganalytics_app_user WITH LOGIN;
CREATE ROLE pganalytics_app_readonly WITH LOGIN;

GRANT USAGE ON SCHEMA pganalytics TO pganalytics_app_master, pganalytics_app_user, pganalytics_app_readonly;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_master;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_user;
GRANT SELECT ON ALL TABLES IN SCHEMA pganalytics TO pganalytics_app_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT ALL ON TABLES TO pganalytics_app_master;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT, INSERT, UPDATE ON TABLES TO pganalytics_app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA pganalytics GRANT SELECT ON TABLES TO pganalytics_app_readonly;
