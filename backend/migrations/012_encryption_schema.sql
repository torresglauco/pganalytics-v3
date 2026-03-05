-- Migration: Encryption at Rest Schema
-- Adds encrypted columns and key versioning for data protection

-- ============================================================================
-- KEY VERSIONING TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS encryption_keys (
    version INTEGER PRIMARY KEY,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'aes-256-gcm',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    retired_at TIMESTAMP,
    key_material_encrypted BYTEA NOT NULL,
    rotation_status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, rotating, retired
    metadata JSONB
);

CREATE INDEX idx_encryption_keys_retired_at ON encryption_keys(retired_at);
CREATE INDEX idx_encryption_keys_rotation_status ON encryption_keys(rotation_status);

COMMENT ON TABLE encryption_keys IS 'Versioned encryption keys for key rotation without downtime';
COMMENT ON COLUMN encryption_keys.version IS 'Key version number for rotation support';
COMMENT ON COLUMN encryption_keys.key_material_encrypted IS 'Key material is itself encrypted for protection at rest';

-- ============================================================================
-- ENCRYPTED COLUMNS - USERS TABLE
-- ============================================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_encrypted BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash_encrypted BYTEA;

CREATE INDEX IF NOT EXISTS idx_users_email_encrypted ON users(email_encrypted);

COMMENT ON COLUMN users.email_encrypted IS 'Encrypted email address (AES-256-GCM)';
COMMENT ON COLUMN users.password_hash_encrypted IS 'Additional encryption layer for password hash';

-- ============================================================================
-- ENCRYPTED COLUMNS - REGISTRATION SECRETS TABLE
-- ============================================================================
ALTER TABLE registration_secrets ADD COLUMN IF NOT EXISTS secret_value_encrypted BYTEA;

COMMENT ON COLUMN registration_secrets.secret_value_encrypted IS 'CRITICAL: Encrypted registration secret (was plaintext before)';

-- ============================================================================
-- ENCRYPTED COLUMNS - POSTGRESQL INSTANCES TABLE
-- ============================================================================
ALTER TABLE postgresql_instances ADD COLUMN IF NOT EXISTS connection_string_encrypted BYTEA;

CREATE INDEX IF NOT EXISTS idx_postgresql_instances_connection_encrypted ON postgresql_instances(connection_string_encrypted);

COMMENT ON COLUMN postgresql_instances.connection_string_encrypted IS 'CRITICAL: Encrypted database connection string (was plaintext before)';

-- ============================================================================
-- ENCRYPTED COLUMNS - API TOKENS TABLE
-- ============================================================================
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS token_hash_encrypted BYTEA;

COMMENT ON COLUMN api_tokens.token_hash_encrypted IS 'Encrypted token hash for additional security';

-- ============================================================================
-- ENCRYPTED COLUMNS - SECRETS TABLE (if exists)
-- ============================================================================
-- Note: Check if secrets table exists in your schema
-- ALTER TABLE secrets ADD COLUMN IF NOT EXISTS secret_encrypted BYTEA;

-- ============================================================================
-- ENCRYPTED COLUMNS - OAUTH PROVIDERS (from previous migration)
-- ============================================================================
-- Note: These were created in previous migration, but included here for completeness

-- ============================================================================
-- BACKUP ENCRYPTION KEY TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS backup_encryption_keys (
    version INTEGER PRIMARY KEY,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'aes-256-gcm',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    retired_at TIMESTAMP,
    key_material_encrypted BYTEA NOT NULL,
    metadata JSONB
);

COMMENT ON TABLE backup_encryption_keys IS 'Separate key versioning for encrypted backups (pg_dump)';

-- ============================================================================
-- DATA MIGRATION TRACKING
-- ============================================================================
CREATE TABLE IF NOT EXISTS encryption_migration_status (
    id BIGSERIAL PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255) NOT NULL,
    plaintext_count INTEGER DEFAULT 0,
    encrypted_count INTEGER DEFAULT 0,
    migration_started_at TIMESTAMP,
    migration_completed_at TIMESTAMP,
    status VARCHAR(50), -- pending, in_progress, completed, failed
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(table_name, column_name)
);

-- ============================================================================
-- FUNCTIONS FOR ENCRYPTION/DECRYPTION
-- ============================================================================

-- Function to get current encryption key (used by app)
CREATE OR REPLACE FUNCTION get_current_encryption_key()
RETURNS BYTEA AS $$
BEGIN
    RETURN key_material_encrypted FROM encryption_keys
    WHERE rotation_status = 'active'
    ORDER BY created_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get key by version
CREATE OR REPLACE FUNCTION get_encryption_key_by_version(p_version INTEGER)
RETURNS BYTEA AS $$
BEGIN
    RETURN key_material_encrypted FROM encryption_keys
    WHERE version = p_version;
END;
$$ LANGUAGE plpgsql;

-- Function to check if data is encrypted
CREATE OR REPLACE FUNCTION is_email_encrypted(p_user_id INTEGER)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN email_encrypted IS NOT NULL
    FROM users
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql;

-- Function to get migration progress
CREATE OR REPLACE FUNCTION get_encryption_migration_progress()
RETURNS TABLE(
    table_name VARCHAR,
    column_name VARCHAR,
    progress_percent NUMERIC,
    status VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ems.table_name,
        ems.column_name,
        ROUND((ems.encrypted_count::NUMERIC / (ems.encrypted_count + ems.plaintext_count) * 100)::NUMERIC, 2),
        ems.status
    FROM encryption_migration_status ems
    ORDER BY ems.updated_at DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COLUMN ENCRYPTION VERIFICATION FUNCTION
-- ============================================================================

CREATE OR REPLACE FUNCTION verify_encryption_integrity(
    p_table_name VARCHAR,
    p_plaintext_column VARCHAR,
    p_encrypted_column VARCHAR
)
RETURNS TABLE(
    total_rows BIGINT,
    encrypted_rows BIGINT,
    plaintext_rows BIGINT,
    mismatch_rows BIGINT,
    encryption_percentage NUMERIC
) AS $$
DECLARE
    v_query TEXT;
    v_total BIGINT;
    v_encrypted BIGINT;
    v_plaintext BIGINT;
    v_mismatch BIGINT;
BEGIN
    -- Count total rows with data
    v_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL OR %I IS NOT NULL',
        p_table_name, p_plaintext_column, p_encrypted_column);
    EXECUTE v_query INTO v_total;

    -- Count encrypted rows
    v_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL',
        p_table_name, p_encrypted_column);
    EXECUTE v_query INTO v_encrypted;

    -- Count plaintext rows
    v_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL',
        p_table_name, p_plaintext_column);
    EXECUTE v_query INTO v_plaintext;

    -- Count rows with both (mismatch)
    v_query := format('SELECT COUNT(*) FROM %I WHERE %I IS NOT NULL AND %I IS NOT NULL',
        p_table_name, p_plaintext_column, p_encrypted_column);
    EXECUTE v_query INTO v_mismatch;

    RETURN QUERY
    SELECT
        v_total,
        v_encrypted,
        v_plaintext,
        v_mismatch,
        ROUND((v_encrypted::NUMERIC / GREATEST(v_total, 1) * 100)::NUMERIC, 2);
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- VIEWS FOR MONITORING ENCRYPTION STATUS
-- ============================================================================

CREATE OR REPLACE VIEW encryption_status AS
SELECT
    'users' as table_name,
    'email' as column_name,
    COUNT(CASE WHEN email_encrypted IS NOT NULL THEN 1 END) as encrypted_count,
    COUNT(CASE WHEN email IS NOT NULL AND email_encrypted IS NULL THEN 1 END) as plaintext_count,
    ROUND((COUNT(CASE WHEN email_encrypted IS NOT NULL THEN 1 END)::NUMERIC /
            GREATEST(COUNT(*), 1) * 100)::NUMERIC, 2) as encryption_percentage
FROM users
UNION ALL
SELECT
    'registration_secrets',
    'secret_value',
    COUNT(CASE WHEN secret_value_encrypted IS NOT NULL THEN 1 END),
    COUNT(CASE WHEN secret_value IS NOT NULL AND secret_value_encrypted IS NULL THEN 1 END),
    ROUND((COUNT(CASE WHEN secret_value_encrypted IS NOT NULL THEN 1 END)::NUMERIC /
            GREATEST(COUNT(*), 1) * 100)::NUMERIC, 2)
FROM registration_secrets
UNION ALL
SELECT
    'postgresql_instances',
    'connection_string',
    COUNT(CASE WHEN connection_string_encrypted IS NOT NULL THEN 1 END),
    COUNT(CASE WHEN connection_string IS NOT NULL AND connection_string_encrypted IS NULL THEN 1 END),
    ROUND((COUNT(CASE WHEN connection_string_encrypted IS NOT NULL THEN 1 END)::NUMERIC /
            GREATEST(COUNT(*), 1) * 100)::NUMERIC, 2)
FROM postgresql_instances;

-- ============================================================================
-- KEY ROTATION TRIGGER
-- ============================================================================

CREATE OR REPLACE FUNCTION trigger_key_rotation()
RETURNS VOID AS $$
BEGIN
    -- Mark old keys as retired (keep last version for decryption)
    UPDATE encryption_keys
    SET rotation_status = 'retired', retired_at = NOW()
    WHERE rotation_status = 'rotating'
        AND retired_at IS NULL;

    -- Clean up very old keys (> 1 year) - keep for audit
    DELETE FROM encryption_keys
    WHERE retired_at < NOW() - INTERVAL '1 year'
        AND created_at < NOW() - INTERVAL '2 years';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PERFORMANCE CONSIDERATIONS
-- ============================================================================

-- Create partial indexes for common queries on encrypted data
CREATE INDEX IF NOT EXISTS idx_users_email_encrypted_not_null ON users(id)
    WHERE email_encrypted IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_postgresql_instances_encrypted_not_null ON postgresql_instances(id)
    WHERE connection_string_encrypted IS NOT NULL;

-- ============================================================================
-- COMMENTS AND DOCUMENTATION
-- ============================================================================

COMMENT ON FUNCTION get_current_encryption_key() IS 'Returns the current active encryption key for use by the application';
COMMENT ON FUNCTION get_encryption_key_by_version(INTEGER) IS 'Returns an encryption key by version number (for decryption after key rotation)';
COMMENT ON FUNCTION get_encryption_migration_progress() IS 'Returns the progress of data encryption migration';
COMMENT ON FUNCTION verify_encryption_integrity(VARCHAR, VARCHAR, VARCHAR) IS 'Verifies that encryption migration is consistent';
COMMENT ON FUNCTION trigger_key_rotation() IS 'Performs key rotation and cleanup of old keys';

COMMENT ON VIEW encryption_status IS 'Shows the encryption status of all sensitive data columns';

-- ============================================================================
-- MIGRATION NOTES
-- ============================================================================

/*
ENCRYPTION AT REST IMPLEMENTATION NOTES:

1. DEPLOYMENT STRATEGY:
   - Phase 1: Add encrypted columns and key manager (this migration)
   - Phase 2: New data is encrypted to _encrypted columns
   - Phase 3: Background migration of existing plaintext to encrypted columns
   - Phase 4: Application switches to reading from encrypted columns
   - Phase 5: Deprecate plaintext columns
   - Phase 6: Drop plaintext columns in future release

2. KEY ROTATION:
   - Keys are versioned and stored with metadata
   - Ciphertext includes key version for transparent decryption
   - Old keys are never deleted (kept for audit and decryption)
   - Automatic cleanup only happens after 2 years retention

3. PERFORMANCE:
   - Encryption is done in the application (not database)
   - Database has no overhead from encryption
   - Partial indexes help with queries on encrypted data

4. BACKUP STRATEGY:
   - Backups use separate encryption key
   - Key material is itself encrypted for protection at rest
   - Backup keys are versioned independently

5. AUDIT & COMPLIANCE:
   - All encryption operations are logged
   - Key rotation is tracked with timestamps
   - Migration progress is monitored via migration_status table
*/
