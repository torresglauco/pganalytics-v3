-- Phase 4 Advanced UI Features Schema Migration
-- Creates tables for alert silences, escalation policies, and state tracking
-- This migration establishes the foundation for advanced alert management

SET search_path TO pganalytics, public;

-- ============================================================================
-- Alert Silences Table
-- Allows users to temporarily suppress alert notifications
-- ============================================================================

CREATE TABLE IF NOT EXISTS alert_silences (
    id BIGSERIAL PRIMARY KEY,

    -- References
    alert_rule_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id) ON DELETE CASCADE,

    -- Silence configuration
    silenced_until TIMESTAMP WITH TIME ZONE NOT NULL,
    silence_type VARCHAR(50) NOT NULL, -- temporary, permanent, schedule-based
    reason TEXT,

    -- Audit fields
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for efficient lookup of active silences
CREATE INDEX IF NOT EXISTS idx_alert_silences_active
    ON alert_silences(alert_rule_id, instance_id, silenced_until DESC)
    WHERE silenced_until > NOW();

CREATE INDEX IF NOT EXISTS idx_alert_silences_instance
    ON alert_silences(instance_id, silenced_until DESC);

-- ============================================================================
-- Escalation Policies Table
-- Defines escalation workflow configurations
-- ============================================================================

CREATE TABLE IF NOT EXISTS escalation_policies (
    id BIGSERIAL PRIMARY KEY,

    -- Configuration
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Audit fields
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for policy lookup
CREATE INDEX IF NOT EXISTS idx_escalation_policies_active
    ON escalation_policies(is_active, name)
    WHERE is_active = true;

-- ============================================================================
-- Escalation Policy Steps Table
-- Defines individual steps within an escalation policy
-- ============================================================================

CREATE TABLE IF NOT EXISTS escalation_policy_steps (
    id BIGSERIAL PRIMARY KEY,

    -- References
    policy_id BIGINT NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,

    -- Step configuration
    step_order INTEGER NOT NULL,
    channel_type VARCHAR(100) NOT NULL, -- email, slack, webhook, pagerduty, sms
    channel_config JSONB NOT NULL, -- Channel-specific configuration
    delay_minutes INTEGER NOT NULL DEFAULT 0, -- Wait time before escalating
    requires_acknowledgment BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT escalation_policy_steps_unique UNIQUE(policy_id, step_order)
);

-- Indexes for step lookup
CREATE INDEX IF NOT EXISTS idx_escalation_policy_steps_policy
    ON escalation_policy_steps(policy_id, step_order);

-- ============================================================================
-- Alert Rule Escalation Policies Linking Table
-- Links alert rules to escalation policies
-- ============================================================================

CREATE TABLE IF NOT EXISTS alert_rule_escalation_policies (
    id BIGSERIAL PRIMARY KEY,

    -- References
    alert_rule_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    escalation_policy_id BIGINT NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,

    -- Configuration
    is_primary BOOLEAN NOT NULL DEFAULT false, -- Primary escalation policy for this rule
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT alert_rule_escalation_policies_unique UNIQUE(alert_rule_id, escalation_policy_id)
);

-- Indexes for lookups
CREATE INDEX IF NOT EXISTS idx_alert_rule_escalation_policies_rule
    ON alert_rule_escalation_policies(alert_rule_id);

CREATE INDEX IF NOT EXISTS idx_alert_rule_escalation_policies_policy
    ON alert_rule_escalation_policies(escalation_policy_id);

-- ============================================================================
-- Escalation State Table
-- Tracks the current state of escalation for alert triggers
-- ============================================================================

CREATE TABLE IF NOT EXISTS escalation_state (
    id BIGSERIAL PRIMARY KEY,

    -- References
    alert_trigger_id BIGINT NOT NULL REFERENCES alert_triggers(id) ON DELETE CASCADE,
    policy_id BIGINT NOT NULL REFERENCES escalation_policies(id) ON DELETE CASCADE,

    -- Current escalation state
    current_step INTEGER NOT NULL DEFAULT 0,
    ack_received BOOLEAN NOT NULL DEFAULT false,
    ack_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ack_at TIMESTAMP WITH TIME ZONE,

    -- Escalation timing
    last_escalated_at TIMESTAMP WITH TIME ZONE,
    next_escalation_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, resolved, acknowledged, failed

    -- Metadata
    metadata JSONB, -- Additional state information
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT escalation_state_unique UNIQUE(alert_trigger_id, policy_id)
);

-- Indexes for state lookup and escalation processing
CREATE INDEX IF NOT EXISTS idx_escalation_state_trigger
    ON escalation_state(alert_trigger_id);

CREATE INDEX IF NOT EXISTS idx_escalation_state_next_escalation
    ON escalation_state(next_escalation_at)
    WHERE status = 'active' AND next_escalation_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_escalation_state_status
    ON escalation_state(status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_escalation_state_policy
    ON escalation_state(policy_id, current_step);

-- ============================================================================
-- Comments for documentation
-- ============================================================================

COMMENT ON TABLE alert_silences IS 'Stores alert silence configurations to suppress notifications';
COMMENT ON COLUMN alert_silences.silence_type IS 'Type of silence: temporary (expires), permanent (until manually removed), or schedule-based';

COMMENT ON TABLE escalation_policies IS 'Defines escalation workflows for alert handling';
COMMENT ON COLUMN escalation_policies.is_active IS 'Whether this policy is available for use';

COMMENT ON TABLE escalation_policy_steps IS 'Individual steps in an escalation policy with notification channels';
COMMENT ON COLUMN escalation_policy_steps.channel_config IS 'JSONB configuration specific to the channel type (recipient, webhook URL, etc.)';
COMMENT ON COLUMN escalation_policy_steps.delay_minutes IS 'Minutes to wait before escalating to this step';

COMMENT ON TABLE alert_rule_escalation_policies IS 'Associates alert rules with escalation policies';
COMMENT ON COLUMN alert_rule_escalation_policies.is_primary IS 'Primary escalation policy takes precedence when multiple are defined';

COMMENT ON TABLE escalation_state IS 'Tracks real-time escalation state for triggered alerts';
COMMENT ON COLUMN escalation_state.current_step IS 'Current step number in the escalation policy (0-based)';
COMMENT ON COLUMN escalation_state.status IS 'Current state: active, resolved, acknowledged, or failed';
