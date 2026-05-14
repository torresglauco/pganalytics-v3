-- Migration 034: Multi-Tenancy Infrastructure with Row-Level Security
-- Implements tenant isolation for SaaS multi-tenancy
-- Supports 2000+ PostgreSQL clusters with logical data separation

BEGIN;

-- ============================================================================
-- TENANTS TABLE (SCALE-04)
-- ============================================================================

-- Create table for tenant management
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Create index for slug lookups
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);

-- Create index for active tenant queries
CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(is_active) WHERE is_active = TRUE;

-- ============================================================================
-- TENANT_USERS TABLE (SCALE-04)
-- ============================================================================

-- Create junction table for user-tenant membership
CREATE TABLE IF NOT EXISTS tenant_users (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'viewer',  -- admin, editor, viewer
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);

-- Create index for user-to-tenant lookup (most common query)
CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id);

-- Create index for tenant membership queries
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant ON tenant_users(tenant_id);

-- ============================================================================
-- ADD TENANT_ID TO COLLECTORS TABLE (SCALE-01)
-- ============================================================================

-- Add tenant_id column to collectors table
ALTER TABLE collectors ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);

-- Create index for tenant-based collector queries
CREATE INDEX IF NOT EXISTS idx_collectors_tenant ON collectors(tenant_id) WHERE tenant_id IS NOT NULL;

-- Create a default tenant for existing collectors (if any exist)
-- This ensures backward compatibility for single-tenant deployments
DO $$
DECLARE
    default_tenant_id UUID;
    collector_count INT;
BEGIN
    SELECT COUNT(*) INTO collector_count FROM collectors WHERE tenant_id IS NULL;

    IF collector_count > 0 THEN
        -- Check if default tenant exists
        SELECT id INTO default_tenant_id FROM tenants WHERE slug = 'default' LIMIT 1;

        IF default_tenant_id IS NULL THEN
            -- Create default tenant
            INSERT INTO tenants (name, slug, is_active)
            VALUES ('Default Tenant', 'default', TRUE)
            RETURNING id INTO default_tenant_id;
        END IF;

        -- Assign collectors without tenant to default tenant
        UPDATE collectors SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    END IF;
END $$;

-- ============================================================================
-- ENABLE ROW-LEVEL SECURITY ON METRIC TABLES (SCALE-02)
-- ============================================================================

-- Enable RLS on host metrics table
ALTER TABLE metrics_host_metrics ENABLE ROW LEVEL SECURITY;

-- Enable RLS on host inventory table
ALTER TABLE metrics_host_inventory ENABLE ROW LEVEL SECURITY;

-- Enable RLS on replication status table
ALTER TABLE metrics_replication_status ENABLE ROW LEVEL SECURITY;

-- Enable RLS on replication slots table
ALTER TABLE metrics_replication_slots ENABLE ROW LEVEL SECURITY;

-- Enable RLS on table inventory table
ALTER TABLE metrics_table_inventory ENABLE ROW LEVEL SECURITY;

-- Enable RLS on data classification table
ALTER TABLE metrics_data_classification ENABLE ROW LEVEL SECURITY;

-- Enable RLS on host health scores table
ALTER TABLE metrics_host_health_scores ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- CREATE RLS POLICIES FOR TENANT ISOLATION (SCALE-02, SCALE-03)
-- ============================================================================

-- Helper function to set tenant context for the current session
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID) RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_tenant', tenant_uuid::text, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create RLS policies for metrics_host_metrics
CREATE POLICY tenant_isolation_policy ON metrics_host_metrics
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_host_metrics ON metrics_host_metrics
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_host_inventory
CREATE POLICY tenant_isolation_policy ON metrics_host_inventory
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_host_inventory ON metrics_host_inventory
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_replication_status
CREATE POLICY tenant_isolation_policy ON metrics_replication_status
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_replication_status ON metrics_replication_status
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_replication_slots
CREATE POLICY tenant_isolation_policy ON metrics_replication_slots
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_replication_slots ON metrics_replication_slots
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_table_inventory
CREATE POLICY tenant_isolation_policy ON metrics_table_inventory
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_table_inventory ON metrics_table_inventory
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_data_classification
CREATE POLICY tenant_isolation_policy ON metrics_data_classification
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_data_classification ON metrics_data_classification
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Create RLS policies for metrics_host_health_scores
CREATE POLICY tenant_isolation_policy ON metrics_host_health_scores
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant', TRUE)::uuid
    ));

CREATE POLICY superuser_bypass_host_health_scores ON metrics_host_health_scores
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- ============================================================================
-- RLS POLICIES FOR COLLECTORS TABLE
-- ============================================================================

-- Enable RLS on collectors table
ALTER TABLE collectors ENABLE ROW LEVEL SECURITY;

-- Policy for tenant-isolated collector access
CREATE POLICY tenant_isolation_policy ON collectors
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);  -- Allow NULL for backward compatibility

-- Superuser bypass for collectors
CREATE POLICY superuser_bypass_collectors ON collectors
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- ============================================================================
-- RLS POLICIES FOR TENANTS TABLE
-- ============================================================================

-- Enable RLS on tenants table
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see tenants they belong to
CREATE POLICY tenant_membership_policy ON tenants
    USING (id IN (
        SELECT tenant_id FROM tenant_users WHERE user_id = current_setting('app.current_user_id', TRUE)::uuid
    ));

-- Superuser bypass for tenants
CREATE POLICY superuser_bypass_tenants ON tenants
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE tenants IS 'Multi-tenant organizations for SaaS isolation';
COMMENT ON TABLE tenant_users IS 'Junction table for user-tenant membership with roles';

COMMENT ON COLUMN tenants.slug IS 'URL-friendly unique identifier for tenant (e.g., acme-corp)';
COMMENT ON COLUMN tenant_users.role IS 'User role within tenant: admin, editor, viewer';

COMMENT ON FUNCTION set_tenant_context(UUID) IS 'Sets the app.current_tenant session variable for RLS policies';

COMMENT ON POLICY tenant_isolation_policy ON metrics_host_metrics IS 'Restricts access to metrics from tenant''s collectors only';

COMMIT;