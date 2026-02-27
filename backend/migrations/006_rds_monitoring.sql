-- RDS PostgreSQL Monitoring Support
-- Extends pgAnalytics to support centralized RDS instance monitoring

SET search_path TO pganalytics, public;

-- ============================================================================
-- RDS INSTANCES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_instances (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    aws_region VARCHAR(50) NOT NULL,  -- e.g., 'us-east-1', 'us-west-2'
    rds_endpoint VARCHAR(255) NOT NULL UNIQUE,  -- e.g., 'mydb.xxxxxxxxxxxx.us-east-1.rds.amazonaws.com'
    port INTEGER DEFAULT 5432,
    engine_version VARCHAR(50),  -- e.g., '14.7', '15.2'
    db_instance_class VARCHAR(50),  -- e.g., 'db.t3.medium', 'db.r6i.xlarge'
    allocated_storage_gb INTEGER,  -- Storage in GB
    environment VARCHAR(50) DEFAULT 'production' CHECK (environment IN ('production', 'staging', 'development', 'test')),
    master_username VARCHAR(255) NOT NULL,
    -- Credentials stored encrypted in secrets table
    secret_id INTEGER REFERENCES secrets(id) ON DELETE SET NULL,
    -- Enable enhanced monitoring (requires IAM role)
    enable_enhanced_monitoring BOOLEAN DEFAULT false,
    monitoring_interval INTEGER DEFAULT 60,  -- Seconds
    -- Connection settings
    ssl_enabled BOOLEAN DEFAULT true,
    ssl_mode VARCHAR(20) DEFAULT 'require' CHECK (ssl_mode IN ('disable', 'allow', 'prefer', 'require', 'verify-ca', 'verify-full')),
    connection_timeout INTEGER DEFAULT 30,  -- Seconds
    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    last_connection_status VARCHAR(50) DEFAULT 'unknown' CHECK (last_connection_status IN ('connected', 'error', 'unknown', 'invalid_credentials')),
    last_error_message TEXT,
    last_error_time TIMESTAMP WITH TIME ZONE,
    -- AWS metadata
    multi_az BOOLEAN DEFAULT false,
    backup_retention_days INTEGER,
    preferred_backup_window VARCHAR(100),
    preferred_maintenance_window VARCHAR(100),
    -- Tags for organization
    tags JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_rds_instances_region ON rds_instances(aws_region);
CREATE INDEX idx_rds_instances_active ON rds_instances(is_active);
CREATE INDEX idx_rds_instances_environment ON rds_instances(environment);
CREATE INDEX idx_rds_instances_status ON rds_instances(last_connection_status);

-- ============================================================================
-- RDS DATABASES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_databases (
    id SERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255),
    size_bytes BIGINT,
    is_template BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    last_analyzed TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_databases_instance ON rds_databases(rds_instance_id);
CREATE INDEX idx_rds_databases_name ON rds_databases(name);
CREATE UNIQUE INDEX idx_rds_databases_instance_name ON rds_databases(rds_instance_id, name);

-- ============================================================================
-- RDS MONITORING METRICS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_metrics (
    id BIGSERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    metric_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    metric_type VARCHAR(100) NOT NULL,  -- e.g., 'cpu', 'disk_space', 'connections', 'iops', 'network'
    metric_value NUMERIC(19, 4) NOT NULL,
    metric_unit VARCHAR(50),  -- e.g., 'percent', 'bytes', 'count', 'MB/s'
    dimensions JSONB DEFAULT '{}',  -- Additional dimensions for grouping
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_metrics_instance_time ON rds_metrics(rds_instance_id, metric_timestamp DESC);
CREATE INDEX idx_rds_metrics_type_time ON rds_metrics(metric_type, metric_timestamp DESC);

-- ============================================================================
-- RDS PERFORMANCE INSIGHTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_performance_insights (
    id SERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    observation_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    observation_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    -- Top dimensions (top SQL, top waits, etc)
    top_dimensions JSONB NOT NULL,  -- Array of top contributing dimensions
    -- Performance metrics for the period
    database_load NUMERIC(10, 2),  -- Average active sessions
    max_database_load NUMERIC(10, 2),
    non_db_cpu_percent NUMERIC(5, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_perf_insights_instance ON rds_performance_insights(rds_instance_id);
CREATE INDEX idx_rds_perf_insights_period ON rds_performance_insights(observation_period_start, observation_period_end);

-- ============================================================================
-- RDS BACKUP & MAINTENANCE HISTORY
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_backup_events (
    id SERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    backup_type VARCHAR(50) NOT NULL CHECK (backup_type IN ('automated', 'manual')),
    backup_id VARCHAR(255),
    backup_window_start TIMESTAMP WITH TIME ZONE,
    backup_window_end TIMESTAMP WITH TIME ZONE,
    backup_size_bytes BIGINT,
    status VARCHAR(50) CHECK (status IN ('available', 'creating', 'deleting', 'failed', 'restored')),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_backups_instance ON rds_backup_events(rds_instance_id);
CREATE INDEX idx_rds_backups_type ON rds_backup_events(backup_type);

-- ============================================================================
-- RDS MAINTENANCE WINDOW HISTORY
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_maintenance_history (
    id SERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    maintenance_type VARCHAR(100) NOT NULL,  -- e.g., 'engine_upgrade', 'os_update', 'storage_optimization'
    maintenance_window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    maintenance_window_end TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('scheduled', 'in_progress', 'completed', 'failed', 'cancelled')),
    description TEXT,
    previous_version VARCHAR(50),
    new_version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_maint_instance ON rds_maintenance_history(rds_instance_id);
CREATE INDEX idx_rds_maint_status ON rds_maintenance_history(status);

-- ============================================================================
-- MONITORING JOBS FOR RDS
-- ============================================================================

CREATE TABLE IF NOT EXISTS rds_monitoring_jobs (
    id SERIAL PRIMARY KEY,
    rds_instance_id INTEGER NOT NULL REFERENCES rds_instances(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL,  -- e.g., 'health_check', 'metrics_collection', 'backup_status'
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    last_error TEXT,
    run_interval_seconds INTEGER DEFAULT 300,  -- Default 5 minutes
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rds_jobs_instance ON rds_monitoring_jobs(rds_instance_id);
CREATE INDEX idx_rds_jobs_next_run ON rds_monitoring_jobs(next_run) WHERE status = 'pending' AND is_enabled = true;

-- ============================================================================
-- SCHEMA MIGRATION TRACKING
-- ============================================================================

INSERT INTO schema_versions (version, description) VALUES
    ('006_rds_monitoring', 'Add RDS PostgreSQL monitoring support')
ON CONFLICT DO NOTHING;
