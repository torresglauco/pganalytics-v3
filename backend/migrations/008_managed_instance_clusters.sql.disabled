-- Managed Instance Cluster Support
-- Allows grouping multiple RDS instances into clusters (master + replicas)

SET search_path TO pganalytics, public;

-- ============================================================================
-- Managed Instance CLUSTERS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS managed_instance_clusters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    cluster_type VARCHAR(50) NOT NULL CHECK (cluster_type IN ('single-az', 'multi-az', 'aurora', 'custom')),
    environment VARCHAR(50) DEFAULT 'production' CHECK (environment IN ('production', 'staging', 'development', 'test')),
    status VARCHAR(50) DEFAULT 'registered' CHECK (status IN ('registering', 'registered', 'monitoring', 'paused')),
    is_active BOOLEAN DEFAULT true,
    tags JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_managed_instance_clusters_environment ON managed_instance_clusters(environment);
CREATE INDEX idx_managed_instance_clusters_status ON managed_instance_clusters(status);
CREATE INDEX idx_managed_instance_clusters_active ON managed_instance_clusters(is_active);

-- ============================================================================
-- UPDATE RDS INSTANCES - ADD CLUSTER RELATIONSHIP
-- ============================================================================

ALTER TABLE managed_instances ADD COLUMN cluster_id INTEGER REFERENCES managed_instance_clusters(id) ON DELETE SET NULL;
ALTER TABLE managed_instances ADD COLUMN instance_role VARCHAR(50) DEFAULT 'standalone' CHECK (instance_role IN ('master', 'read-replica', 'standby', 'standalone'));

CREATE INDEX idx_managed_instances_cluster ON managed_instances(cluster_id);
CREATE INDEX idx_managed_instances_role ON managed_instances(instance_role);

-- ============================================================================
-- SCHEMA MIGRATION TRACKING
-- ============================================================================

INSERT INTO schema_versions (version, description) VALUES
    ('008_managed_instance_clusters', 'Add RDS cluster support with master/replica grouping')
ON CONFLICT DO NOTHING;
