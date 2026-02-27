-- Registration secrets for collector self-registration
CREATE TABLE IF NOT EXISTS registration_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    secret_value VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_by INTEGER,  -- Will be set to NULL if users table doesn't exist
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,

    -- Track usage
    total_registrations INTEGER DEFAULT 0,
    last_used_at TIMESTAMP
);

-- Note: Foreign key constraint on users table not added due to initialization issues
-- The users table may not be properly populated during migration
-- created_by can be NULL and will be set during registration if user data is available

-- Index for lookups
CREATE INDEX idx_registration_secrets_secret_value ON registration_secrets(secret_value) WHERE active = true;
CREATE INDEX idx_registration_secrets_active ON registration_secrets(active);
CREATE INDEX idx_registration_secrets_created_by ON registration_secrets(created_by);

-- Audit table for tracking secret usage
CREATE TABLE IF NOT EXISTS registration_secret_audit (
    id BIGSERIAL PRIMARY KEY,
    secret_id UUID REFERENCES registration_secrets(id) ON DELETE CASCADE,
    collector_id UUID,
    collector_name VARCHAR(255),
    status VARCHAR(50), -- 'success', 'failed', 'expired'
    error_message TEXT,
    ip_address VARCHAR(45),
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_registration_secret_audit_secret_id ON registration_secret_audit(secret_id);
CREATE INDEX idx_registration_secret_audit_used_at ON registration_secret_audit(used_at);
