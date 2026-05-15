-- Migration 036: Alert Rules Multi-Tenancy
-- Adds tenant_id column to alert-related tables with Row-Level Security policies
-- Ensures tenant isolation for alert rules, triggers, silences, escalation policies, and channels

BEGIN;

SET search_path TO pganalytics, public;

-- ============================================================================
-- ADD TENANT_ID TO ALERT_RULES TABLE
-- ============================================================================

-- Add tenant_id column to alert_rules
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based alert rule queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id ON alert_rules(tenant_id) WHERE tenant_id IS NOT NULL;

-- Create index for combined tenant + enabled status queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_enabled ON alert_rules(tenant_id, is_enabled) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- ADD TENANT_ID TO ALERT_TRIGGERS TABLE
-- ============================================================================

-- Add tenant_id column to alert_triggers
ALTER TABLE alert_triggers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based alert trigger queries
CREATE INDEX IF NOT EXISTS idx_alert_triggers_tenant_id ON alert_triggers(tenant_id) WHERE tenant_id IS NOT NULL;

-- Create index for combined tenant + status queries
CREATE INDEX IF NOT EXISTS idx_alert_triggers_tenant_status ON alert_triggers(tenant_id, status) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- ADD TENANT_ID TO NOTIFICATION_CHANNELS TABLE
-- ============================================================================

-- Add tenant_id column to notification_channels
ALTER TABLE notification_channels ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based notification channel queries
CREATE INDEX IF NOT EXISTS idx_notification_channels_tenant_id ON notification_channels(tenant_id) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- ADD TENANT_ID TO ALERT_SILENCES TABLE
-- ============================================================================

-- Add tenant_id column to alert_silences
ALTER TABLE alert_silences ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based silence queries
CREATE INDEX IF NOT EXISTS idx_alert_silences_tenant_id ON alert_silences(tenant_id) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- ADD TENANT_ID TO ESCALATION_POLICIES TABLE
-- ============================================================================

-- Add tenant_id column to escalation_policies
ALTER TABLE escalation_policies ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based escalation policy queries
CREATE INDEX IF NOT EXISTS idx_escalation_policies_tenant_id ON escalation_policies(tenant_id) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- ADD TENANT_ID TO ESCALATION_STATE TABLE
-- ============================================================================

-- Add tenant_id column to escalation_state
ALTER TABLE escalation_state ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant-based escalation state queries
CREATE INDEX IF NOT EXISTS idx_escalation_state_tenant_id ON escalation_state(tenant_id) WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- UPDATE EXISTING RECORDS WITH DEFAULT TENANT
-- ============================================================================

-- Set default tenant for existing alert-related records (backward compatibility)
DO $$
DECLARE
    default_tenant_id UUID;
BEGIN
    -- Get default tenant ID
    SELECT id INTO default_tenant_id FROM tenants WHERE slug = 'default' LIMIT 1;

    IF default_tenant_id IS NOT NULL THEN
        -- Update alert_rules without tenant
        UPDATE alert_rules SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

        -- Update alert_triggers without tenant
        UPDATE alert_triggers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

        -- Update notification_channels without tenant
        UPDATE notification_channels SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

        -- Update alert_silences without tenant
        UPDATE alert_silences SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

        -- Update escalation_policies without tenant
        UPDATE escalation_policies SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

        -- Update escalation_state without tenant
        UPDATE escalation_state SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    END IF;
END $$;

-- ============================================================================
-- ENABLE ROW-LEVEL SECURITY ON ALERT TABLES
-- ============================================================================

-- Enable RLS on alert_rules
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;

-- Enable RLS on alert_triggers
ALTER TABLE alert_triggers ENABLE ROW LEVEL SECURITY;

-- Enable RLS on notification_channels
ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;

-- Enable RLS on alert_silences
ALTER TABLE alert_silences ENABLE ROW LEVEL SECURITY;

-- Enable RLS on escalation_policies
ALTER TABLE escalation_policies ENABLE ROW LEVEL SECURITY;

-- Enable RLS on escalation_state
ALTER TABLE escalation_state ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- CREATE RLS POLICIES FOR TENANT ISOLATION
-- ============================================================================

-- Alert Rules: Tenant isolation policy
CREATE POLICY alert_rules_tenant_isolation ON alert_rules
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);  -- Allow NULL for backward compatibility

-- Alert Rules: Superuser bypass
CREATE POLICY superuser_bypass_alert_rules ON alert_rules
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Alert Triggers: Tenant isolation policy
CREATE POLICY alert_triggers_tenant_isolation ON alert_triggers
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);

-- Alert Triggers: Superuser bypass
CREATE POLICY superuser_bypass_alert_triggers ON alert_triggers
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Notification Channels: Tenant isolation policy
CREATE POLICY notification_channels_tenant_isolation ON notification_channels
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);

-- Notification Channels: Superuser bypass
CREATE POLICY superuser_bypass_notification_channels ON notification_channels
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Alert Silences: Tenant isolation policy
CREATE POLICY alert_silences_tenant_isolation ON alert_silences
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);

-- Alert Silences: Superuser bypass
CREATE POLICY superuser_bypass_alert_silences ON alert_silences
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Escalation Policies: Tenant isolation policy
CREATE POLICY escalation_policies_tenant_isolation ON escalation_policies
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);

-- Escalation Policies: Superuser bypass
CREATE POLICY superuser_bypass_escalation_policies ON escalation_policies
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- Escalation State: Tenant isolation policy
CREATE POLICY escalation_state_tenant_isolation ON escalation_state
    USING (tenant_id = current_setting('app.current_tenant', TRUE)::uuid
           OR tenant_id IS NULL);

-- Escalation State: Superuser bypass
CREATE POLICY superuser_bypass_escalation_state ON escalation_state
    USING (pg_has_role(current_user, 'pg_write_all_data', 'member'));

-- ============================================================================
-- TRIGGER TO AUTO-POPULATE TENANT_ID ON INSERT
-- ============================================================================

-- Function to auto-set tenant_id from session variable
CREATE OR REPLACE FUNCTION auto_set_tenant_id() RETURNS TRIGGER AS $$
BEGIN
    -- If tenant_id is NULL, try to get it from session variable
    IF NEW.tenant_id IS NULL THEN
        BEGIN
            NEW.tenant_id := current_setting('app.current_tenant', TRUE)::uuid;
        EXCEPTION WHEN OTHERS THEN
            -- If session variable is not set, leave as NULL
            NEW.tenant_id := NULL;
        END;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to alert_rules
DROP TRIGGER IF EXISTS trigger_alert_rules_tenant_id ON alert_rules;
CREATE TRIGGER trigger_alert_rules_tenant_id
    BEFORE INSERT ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- Apply trigger to alert_triggers
DROP TRIGGER IF EXISTS trigger_alert_triggers_tenant_id ON alert_triggers;
CREATE TRIGGER trigger_alert_triggers_tenant_id
    BEFORE INSERT ON alert_triggers
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- Apply trigger to notification_channels
DROP TRIGGER IF EXISTS trigger_notification_channels_tenant_id ON notification_channels;
CREATE TRIGGER trigger_notification_channels_tenant_id
    BEFORE INSERT ON notification_channels
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- Apply trigger to alert_silences
DROP TRIGGER IF EXISTS trigger_alert_silences_tenant_id ON alert_silences;
CREATE TRIGGER trigger_alert_silences_tenant_id
    BEFORE INSERT ON alert_silences
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- Apply trigger to escalation_policies
DROP TRIGGER IF EXISTS trigger_escalation_policies_tenant_id ON escalation_policies;
CREATE TRIGGER trigger_escalation_policies_tenant_id
    BEFORE INSERT ON escalation_policies
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- Apply trigger to escalation_state
DROP TRIGGER IF EXISTS trigger_escalation_state_tenant_id ON escalation_state;
CREATE TRIGGER trigger_escalation_state_tenant_id
    BEFORE INSERT ON escalation_state
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_tenant_id();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON COLUMN alert_rules.tenant_id IS 'Tenant ID for multi-tenant isolation';
COMMENT ON COLUMN alert_triggers.tenant_id IS 'Tenant ID for multi-tenant isolation';
COMMENT ON COLUMN notification_channels.tenant_id IS 'Tenant ID for multi-tenant isolation';
COMMENT ON COLUMN alert_silences.tenant_id IS 'Tenant ID for multi-tenant isolation';
COMMENT ON COLUMN escalation_policies.tenant_id IS 'Tenant ID for multi-tenant isolation';
COMMENT ON COLUMN escalation_state.tenant_id IS 'Tenant ID for multi-tenant isolation';

COMMENT ON POLICY alert_rules_tenant_isolation ON alert_rules IS 'Restricts access to alert rules belonging to the current tenant';
COMMENT ON POLICY alert_triggers_tenant_isolation ON alert_triggers IS 'Restricts access to alert triggers belonging to the current tenant';
COMMENT ON POLICY notification_channels_tenant_isolation ON notification_channels IS 'Restricts access to notification channels belonging to the current tenant';
COMMENT ON POLICY alert_silences_tenant_isolation ON alert_silences IS 'Restricts access to silences belonging to the current tenant';
COMMENT ON POLICY escalation_policies_tenant_isolation ON escalation_policies IS 'Restricts access to escalation policies belonging to the current tenant';
COMMENT ON POLICY escalation_state_tenant_isolation ON escalation_state IS 'Restricts access to escalation state belonging to the current tenant';

COMMENT ON FUNCTION auto_set_tenant_id() IS 'Auto-populates tenant_id from session variable on INSERT';

COMMIT;