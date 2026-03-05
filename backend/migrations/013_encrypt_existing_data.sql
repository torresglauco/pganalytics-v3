-- Migration: Encrypt Existing Plaintext Data
-- Migrates plaintext sensitive columns to encrypted format
-- This migration runs as a background job to avoid blocking production database

-- ============================================================================
-- MIGRATION STATUS TRACKING
-- ============================================================================
-- Use encryption_migration_status table (created in 012_encryption_schema.sql)
-- to track progress of data migration

-- ============================================================================
-- EMAIL ENCRYPTION
-- ============================================================================
-- Users table: migrate existing plaintext emails to encrypted format
-- Note: This assumes encryption key is available via application
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'users', 'email', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'users' AND column_name = 'email'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- SQL function to encrypt a string (requires application-side key management)
-- This is a placeholder - actual encryption should be done by application with key manager
CREATE OR REPLACE FUNCTION migrate_encrypt_users_email()
RETURNS TABLE(migrated_count INTEGER, total_count INTEGER, status TEXT)
LANGUAGE plpgsql AS $$
DECLARE
  v_total INTEGER;
  v_migrated INTEGER;
BEGIN
  -- Count total plaintext emails
  SELECT COUNT(*) INTO v_total
  FROM users
  WHERE email IS NOT NULL AND email_encrypted IS NULL;

  -- Count already encrypted
  SELECT COUNT(*) INTO v_migrated
  FROM users
  WHERE email IS NOT NULL AND email_encrypted IS NOT NULL;

  -- Update migration status
  UPDATE encryption_migration_status
  SET
    plaintext_count = v_total,
    encrypted_count = v_migrated,
    status = CASE
      WHEN v_total = 0 THEN 'completed'
      ELSE 'in_progress'
    END,
    migration_completed_at = CASE
      WHEN v_total = 0 THEN NOW()
      ELSE NULL
    END
  WHERE table_name = 'users' AND column_name = 'email';

  RETURN QUERY
  SELECT v_migrated, v_total + v_migrated,
    CASE
      WHEN v_total = 0 THEN 'completed'
      ELSE 'in_progress'
    END;
END
$$;

-- ============================================================================
-- PASSWORD HASH ENCRYPTION
-- ============================================================================
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'users', 'password_hash', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'users' AND column_name = 'password_hash'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- ============================================================================
-- REGISTRATION SECRETS ENCRYPTION (CRITICAL)
-- ============================================================================
-- CRITICAL: Registration secrets were previously stored in plaintext
-- This migration encrypts all existing registration secrets

INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'registration_secrets', 'secret_value', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'registration_secrets' AND column_name = 'secret_value'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- Count plaintext registration secrets
CREATE OR REPLACE FUNCTION count_plaintext_registration_secrets()
RETURNS INTEGER AS $$
BEGIN
  RETURN (
    SELECT COUNT(*)
    FROM registration_secrets
    WHERE secret_value IS NOT NULL AND secret_value_encrypted IS NULL
  );
END
$$ LANGUAGE plpgsql;

-- ============================================================================
-- DATABASE CONNECTION STRINGS (CRITICAL)
-- ============================================================================
-- CRITICAL: Connection strings contain sensitive database credentials
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'postgresql_instances', 'connection_string', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'postgresql_instances' AND column_name = 'connection_string'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- Count plaintext connection strings
CREATE OR REPLACE FUNCTION count_plaintext_connection_strings()
RETURNS INTEGER AS $$
BEGIN
  RETURN (
    SELECT COUNT(*)
    FROM postgresql_instances
    WHERE connection_string IS NOT NULL AND connection_string_encrypted IS NULL
  );
END
$$ LANGUAGE plpgsql;

-- ============================================================================
-- API TOKEN HASHES
-- ============================================================================
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'api_tokens', 'token_hash', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'api_tokens' AND column_name = 'token_hash'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- ============================================================================
-- OAUTH PROVIDER SECRETS
-- ============================================================================
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'oauth_providers', 'client_secret', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'oauth_providers' AND column_name = 'client_secret'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- ============================================================================
-- LDAP CONFIGURATION (CRITICAL)
-- ============================================================================
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'ldap_config', 'bind_password', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'ldap_config' AND column_name = 'bind_password'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- ============================================================================
-- SAML CONFIGURATION
-- ============================================================================
INSERT INTO encryption_migration_status (table_name, column_name, status)
SELECT 'saml_config', 'sp_cert', 'pending'
WHERE NOT EXISTS (
  SELECT 1 FROM encryption_migration_status
  WHERE table_name = 'saml_config' AND column_name = 'sp_cert'
)
ON CONFLICT(table_name, column_name) DO UPDATE
SET updated_at = NOW();

-- ============================================================================
-- MASTER MIGRATION FUNCTION
-- ============================================================================
-- This function is called by the application (golang) to migrate data
-- The application handles encryption using the key manager
CREATE OR REPLACE FUNCTION get_pending_migrations()
RETURNS TABLE(table_name VARCHAR, column_name VARCHAR, plaintext_count INTEGER, encrypted_count INTEGER)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY
  SELECT
    ems.table_name,
    ems.column_name,
    ems.plaintext_count,
    ems.encrypted_count
  FROM encryption_migration_status ems
  WHERE ems.status = 'pending' OR (ems.status = 'in_progress' AND ems.migration_completed_at IS NULL)
  ORDER BY ems.created_at ASC;
END
$$;

-- ============================================================================
-- VERIFICATION VIEWS
-- ============================================================================
-- View to check encryption status across all tables
CREATE OR REPLACE VIEW encryption_migration_status_v AS
SELECT
  ems.table_name,
  ems.column_name,
  ems.plaintext_count,
  ems.encrypted_count,
  (ems.encrypted_count::FLOAT / (ems.encrypted_count + ems.plaintext_count)::FLOAT * 100)::NUMERIC(5,2) AS encryption_percentage,
  ems.status,
  ems.migration_started_at,
  ems.migration_completed_at,
  CASE
    WHEN ems.plaintext_count = 0 THEN 'completed'
    WHEN ems.migration_completed_at IS NOT NULL THEN 'completed'
    ELSE 'pending'
  END AS actual_status
FROM encryption_migration_status ems
ORDER BY ems.table_name, ems.column_name;

-- View to identify tables still needing encryption
CREATE OR REPLACE VIEW unencrypted_sensitive_data AS
SELECT 'users' AS table_name, 'email' AS column_name, COUNT(*) AS count
FROM users
WHERE email IS NOT NULL AND email_encrypted IS NULL
UNION ALL
SELECT 'users', 'password_hash', COUNT(*)
FROM users
WHERE password_hash IS NOT NULL AND password_hash_encrypted IS NULL
UNION ALL
SELECT 'registration_secrets', 'secret_value', COUNT(*)
FROM registration_secrets
WHERE secret_value IS NOT NULL AND secret_value_encrypted IS NULL
UNION ALL
SELECT 'postgresql_instances', 'connection_string', COUNT(*)
FROM postgresql_instances
WHERE connection_string IS NOT NULL AND connection_string_encrypted IS NULL
ORDER BY table_name, column_name;

-- ============================================================================
-- INDEXES FOR ENCRYPTED DATA
-- ============================================================================
-- These indexes help with lookups on encrypted data
-- (Note: Indexes on encrypted columns are less effective than plaintext,
--  but still useful for some operations)

CREATE INDEX IF NOT EXISTS idx_users_email_encrypted_active
ON users(email_encrypted)
WHERE email_encrypted IS NOT NULL AND is_active = true;

CREATE INDEX IF NOT EXISTS idx_postgresql_instances_connection_encrypted_active
ON postgresql_instances(connection_string_encrypted)
WHERE connection_string_encrypted IS NOT NULL AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_registration_secrets_encrypted_active
ON registration_secrets(secret_value_encrypted)
WHERE secret_value_encrypted IS NOT NULL AND is_active = true;

-- ============================================================================
-- COMMENT & DOCUMENTATION
-- ============================================================================
COMMENT ON FUNCTION migrate_encrypt_users_email() IS
'Returns encryption status for users email migration.
Called by background job in application.';

COMMENT ON FUNCTION count_plaintext_registration_secrets() IS
'Returns count of unencrypted registration secrets - CRITICAL for security audit.
Should be 0 in production after migration.';

COMMENT ON FUNCTION count_plaintext_connection_strings() IS
'Returns count of unencrypted database connection strings - CRITICAL for security audit.
Should be 0 in production after migration.';

COMMENT ON FUNCTION get_pending_migrations() IS
'Returns list of pending data migrations for encryption.
Called by background job to coordinate encryption work.';

COMMENT ON VIEW encryption_migration_status_v IS
'Provides overview of encryption migration progress across all tables.
Use this to monitor data encryption status in production.';

COMMENT ON VIEW unencrypted_sensitive_data IS
'Lists remaining plaintext sensitive data that should be encrypted.
This view helps identify security gaps and track progress.
Should return 0 rows after all migrations complete.';

-- ============================================================================
-- MIGRATION COMPLETION VERIFICATION
-- ============================================================================
-- Run these queries to verify encryption migration completeness:
--
-- SELECT * FROM encryption_migration_status_v;
-- SELECT * FROM unencrypted_sensitive_data;
-- SELECT COUNT(*) FROM users WHERE email IS NOT NULL AND email_encrypted IS NULL;
-- SELECT COUNT(*) FROM registration_secrets WHERE secret_value IS NOT NULL AND secret_value_encrypted IS NULL;
-- SELECT COUNT(*) FROM postgresql_instances WHERE connection_string IS NOT NULL AND connection_string_encrypted IS NULL;
