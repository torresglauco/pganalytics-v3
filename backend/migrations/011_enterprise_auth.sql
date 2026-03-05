-- Migration: Enterprise Authentication Schema
-- Adds tables and functions for LDAP, SAML, OAuth, MFA, and session management

-- ============================================================================
-- USER MFA METHODS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_mfa_methods (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- totp, sms, email
    secret_encrypted BYTEA, -- Base32-encoded TOTP secret or encrypted phone number
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    enabled BOOLEAN DEFAULT FALSE,
    enabled_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, type)
);

CREATE INDEX idx_user_mfa_methods_user_id ON user_mfa_methods(user_id);
CREATE INDEX idx_user_mfa_methods_enabled ON user_mfa_methods(user_id, enabled);

-- ============================================================================
-- USER BACKUP CODES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_backup_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, code_hash)
);

CREATE INDEX idx_user_backup_codes_user_id ON user_backup_codes(user_id);
CREATE INDEX idx_user_backup_codes_used ON user_backup_codes(user_id, used);

-- ============================================================================
-- USER SESSIONS TABLE (Distributed Sessions)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(session_token_hash);

-- Auto-cleanup of expired sessions (daily)
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM user_sessions
    WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- OAUTH PROVIDER CONFIGURATION
-- ============================================================================
CREATE TABLE IF NOT EXISTS oauth_providers (
    id BIGSERIAL PRIMARY KEY,
    provider_name VARCHAR(100) NOT NULL UNIQUE, -- google, github, azure_ad, custom
    client_id_encrypted BYTEA NOT NULL,
    client_secret_encrypted BYTEA NOT NULL,
    config_json JSONB,
    enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- LDAP CONFIGURATION (encrypted in secrets)
-- ============================================================================
CREATE TABLE IF NOT EXISTS ldap_config (
    id BIGSERIAL PRIMARY KEY,
    server_url_encrypted BYTEA NOT NULL,
    bind_dn_encrypted BYTEA,
    bind_password_encrypted BYTEA,
    user_search_base VARCHAR(255),
    group_search_base VARCHAR(255),
    group_to_role_mapping JSONB,
    enabled BOOLEAN DEFAULT FALSE,
    tested_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- SAML CONFIGURATION
-- ============================================================================
CREATE TABLE IF NOT EXISTS saml_config (
    id BIGSERIAL PRIMARY KEY,
    idp_url VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    cert_data_encrypted BYTEA NOT NULL,
    key_data_encrypted BYTEA NOT NULL,
    metadata_xml TEXT,
    enabled BOOLEAN DEFAULT FALSE,
    tested_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- AUTHENTICATION EVENTS (Audit Trail for Security)
-- ============================================================================
CREATE TABLE IF NOT EXISTS auth_events (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL, -- login_success, login_failed, mfa_setup, mfa_verify, logout, password_change, token_refresh
    ip_address INET,
    user_agent TEXT,
    auth_method VARCHAR(50), -- ldap, saml, oauth, jwt, password
    provider VARCHAR(100), -- For OAuth: google, github, azure_ad
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_events_user_id ON auth_events(user_id);
CREATE INDEX idx_auth_events_created_at ON auth_events(created_at);
CREATE INDEX idx_auth_events_event_type ON auth_events(event_type);
CREATE INDEX idx_auth_events_ip_address ON auth_events(ip_address);

-- ============================================================================
-- ACTIVE SESSIONS TRACKING
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_active_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name VARCHAR(255),
    device_type VARCHAR(50), -- web, mobile, api
    ip_address INET,
    user_agent TEXT,
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_active_sessions_user_id ON user_active_sessions(user_id);
CREATE INDEX idx_user_active_sessions_last_seen ON user_active_sessions(last_seen);

-- ============================================================================
-- TOKEN BLACKLIST (For logout and revocation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS token_blacklist (
    id BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_type VARCHAR(50), -- access_token, refresh_token
    reason VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_token_blacklist_token_hash ON token_blacklist(token_hash);
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- ============================================================================
-- LOGIN ATTEMPTS TRACKING (Brute force protection)
-- ============================================================================
CREATE TABLE IF NOT EXISTS login_attempts (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    success BOOLEAN NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_attempts_username_created ON login_attempts(username, created_at);
CREATE INDEX idx_login_attempts_ip_created ON login_attempts(ip_address, created_at);

-- Create function to check brute force
CREATE OR REPLACE FUNCTION check_login_attempts_brute_force(p_username VARCHAR, p_ip INET, p_window_minutes INTEGER DEFAULT 15, p_max_attempts INTEGER DEFAULT 5)
RETURNS BOOLEAN AS $$
DECLARE
    v_attempts INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO v_attempts
    FROM login_attempts
    WHERE username = p_username
        AND ip_address = p_ip
        AND success = FALSE
        AND created_at > NOW() - (p_window_minutes || ' minutes')::INTERVAL;

    RETURN v_attempts >= p_max_attempts;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- USER AUTHENTICATION PROVIDER MAPPING
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_auth_providers (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL, -- ldap, saml, oauth_google, oauth_github, oauth_azure
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    provider_attributes JSONB,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_user_auth_providers_user_id ON user_auth_providers(user_id);
CREATE INDEX idx_user_auth_providers_provider ON user_auth_providers(provider, provider_user_id);

-- ============================================================================
-- CLEANUP TRIGGERS AND FUNCTIONS
-- ============================================================================

-- Clean up expired sessions daily (via cron job)
CREATE OR REPLACE FUNCTION trigger_cleanup_expired_sessions()
RETURNS VOID AS $$
BEGIN
    DELETE FROM user_sessions
    WHERE expires_at < NOW() - INTERVAL '1 day';

    DELETE FROM token_blacklist
    WHERE expires_at < NOW();

    DELETE FROM login_attempts
    WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PERMISSIONS AND SECURITY
-- ============================================================================

-- Ensure sensitive data is not accidentally exposed
COMMENT ON TABLE oauth_providers IS 'OAuth provider credentials - handle with care';
COMMENT ON TABLE ldap_config IS 'LDAP configuration - credentials encrypted';
COMMENT ON TABLE saml_config IS 'SAML configuration - credentials encrypted';
COMMENT ON COLUMN user_mfa_methods.secret_encrypted IS 'TOTP secrets are encrypted';
COMMENT ON COLUMN oauth_providers.client_secret_encrypted IS 'Client secret is always encrypted';

-- ============================================================================
-- INITIAL SETUP QUERIES (Run separately if needed)
-- ============================================================================

-- View for user MFA status (useful for queries)
CREATE OR REPLACE VIEW user_mfa_status AS
SELECT
    u.id as user_id,
    u.username,
    COUNT(DISTINCT CASE WHEN umm.enabled = true THEN umm.type END) as enabled_mfa_count,
    MAX(CASE WHEN umm.type = 'totp' THEN umm.enabled ELSE false END) as has_totp,
    MAX(CASE WHEN umm.type = 'sms' THEN umm.enabled ELSE false END) as has_sms,
    MAX(CASE WHEN umm.type = 'email' THEN umm.enabled ELSE false END) as has_email,
    COUNT(DISTINCT CASE WHEN ubc.used = false THEN ubc.id END) as backup_codes_remaining
FROM users u
LEFT JOIN user_mfa_methods umm ON u.id = umm.user_id
LEFT JOIN user_backup_codes ubc ON u.id = ubc.user_id
GROUP BY u.id, u.username;

-- View for active sessions
CREATE OR REPLACE VIEW active_user_sessions AS
SELECT
    us.id,
    us.user_id,
    u.username,
    us.ip_address,
    us.user_agent,
    us.last_activity,
    us.created_at,
    EXTRACT(EPOCH FROM (us.expires_at - NOW())) as seconds_until_expiry
FROM user_sessions us
JOIN users u ON us.user_id = u.id
WHERE us.expires_at > NOW()
ORDER BY us.last_activity DESC;

-- ============================================================================
-- GRANTS (adjust based on your role structure)
-- ============================================================================

-- Example: If you have read-only and admin roles
-- GRANT SELECT ON user_mfa_methods TO "read_only_role";
-- GRANT ALL ON user_mfa_methods TO "admin_role";
-- GRANT SELECT ON auth_events TO "audit_role";
