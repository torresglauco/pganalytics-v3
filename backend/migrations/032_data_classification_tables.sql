-- Migration 032: Data Classification Tables
-- Stores PII/PCI detection results and custom patterns for data classification
-- Includes TimescaleDB hypertables for time-series queries

BEGIN;

-- ============================================================================
-- DATA CLASSIFICATION RESULTS TABLE (DATA-01, DATA-02, DATA-03)
-- ============================================================================

-- Create table for data classification results from column analysis
CREATE TABLE IF NOT EXISTS metrics_data_classification (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    column_name VARCHAR(255),
    pattern_type VARCHAR(50),        -- CPF, CNPJ, EMAIL, PHONE, CREDIT_CARD, CUSTOM
    category VARCHAR(20),            -- PII, PCI, SENSITIVE, CUSTOM
    confidence FLOAT,                -- 0.0 to 1.0 confidence score
    match_count BIGINT,              -- Number of rows matching the pattern
    sample_values JSONB,             -- Array of masked sample values (up to 5)
    regulation_mapping JSONB,        -- Map of regulation -> article references
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, column_name)
);

-- Create TimescaleDB hypertable for data classification
SELECT create_hypertable('metrics_data_classification', 'time',
    if_not_exists => TRUE,
    migrate_data => FALSE);

-- Create index for efficient querying by collector
CREATE INDEX IF NOT EXISTS idx_data_classification_collector
    ON metrics_data_classification (collector_id, time DESC);

-- Create index for filtering by pattern type and category
CREATE INDEX IF NOT EXISTS idx_data_classification_pattern_category
    ON metrics_data_classification (pattern_type, category);

-- Create index for filtering by database/schema/table
CREATE INDEX IF NOT EXISTS idx_data_classification_location
    ON metrics_data_classification (collector_id, database_name, schema_name, table_name);

-- ============================================================================
-- CUSTOM PATTERNS TABLE (DATA-04)
-- ============================================================================

-- Create table for custom detection patterns (tenant-specific or global)
CREATE TABLE IF NOT EXISTS data_classification_patterns (
    id SERIAL PRIMARY KEY,
    tenant_id UUID,                  -- NULL for global patterns, UUID for tenant-specific
    pattern_name VARCHAR(255) NOT NULL,
    pattern_regex TEXT NOT NULL,
    category VARCHAR(20) NOT NULL,   -- PII, PCI, SENSITIVE, CUSTOM
    validation_algorithm VARCHAR(50), -- Luhn, Mod11, None
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    enabled BOOLEAN DEFAULT TRUE
);

-- Create index for tenant-filtered pattern queries
CREATE INDEX IF NOT EXISTS idx_patterns_tenant
    ON data_classification_patterns (tenant_id) WHERE tenant_id IS NOT NULL;

-- Create index for enabled patterns
CREATE INDEX IF NOT EXISTS idx_patterns_enabled
    ON data_classification_patterns (enabled) WHERE enabled = TRUE;

-- ============================================================================
-- REGULATION MAPPINGS SEED DATA
-- ============================================================================

-- Insert regulation mappings for common patterns
-- These are reference values used during classification

CREATE TABLE IF NOT EXISTS regulation_mappings (
    pattern_type VARCHAR(50) PRIMARY KEY,
    regulations JSONB NOT NULL
);

INSERT INTO regulation_mappings (pattern_type, regulations) VALUES
('CPF', '{"LGPD": ["Art. 5, I - dado pessoal", "Art. 11 - consentimento"]}'::jsonb),
('CNPJ', '{"LGPD": ["Art. 5, I - dado pessoal"]}'::jsonb),
('EMAIL', '{"LGPD": ["Art. 5, I - dado pessoal"], "GDPR": ["Art. 4(1) - personal data"]}'::jsonb),
('PHONE', '{"LGPD": ["Art. 5, I - dado pessoal"], "GDPR": ["Art. 4(1) - personal data"]}'::jsonb),
('CREDIT_CARD', '{"PCI-DSS": ["Req. 3 - Protect stored cardholder data", "Req. 4 - Encrypt transmission"], "LGPD": ["Art. 5, I - dado pessoal"], "GDPR": ["Art. 4(1) - personal data"]}'::jsonb)
ON CONFLICT (pattern_type) DO NOTHING;

-- ============================================================================
-- RETENTION POLICIES
-- ============================================================================

-- Set retention policy for data classification results (keep for 90 days)
SELECT add_retention_policy('metrics_data_classification', INTERVAL '90 days', if_not_exists => TRUE);

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE metrics_data_classification IS 'Stores data classification results from PII/PCI detection scans';
COMMENT ON TABLE data_classification_patterns IS 'Custom detection patterns for tenant-specific sensitive data';
COMMENT ON TABLE regulation_mappings IS 'Reference mapping of pattern types to regulatory articles (LGPD, GDPR, PCI-DSS)';

COMMENT ON COLUMN metrics_data_classification.pattern_type IS 'Type of sensitive data pattern: CPF, CNPJ, EMAIL, PHONE, CREDIT_CARD, CUSTOM';
COMMENT ON COLUMN metrics_data_classification.category IS 'Category of sensitive data: PII, PCI, SENSITIVE, CUSTOM';
COMMENT ON COLUMN metrics_data_classification.confidence IS 'Detection confidence score from 0.0 to 1.0';
COMMENT ON COLUMN metrics_data_classification.sample_values IS 'JSONB array of masked sample values (up to 5 samples)';
COMMENT ON COLUMN metrics_data_classification.regulation_mapping IS 'JSONB map of regulation name to applicable articles';

COMMENT ON COLUMN data_classification_patterns.tenant_id IS 'NULL for global patterns, UUID for tenant-specific patterns';
COMMENT ON COLUMN data_classification_patterns.validation_algorithm IS 'Optional validation: Luhn (credit cards), Mod11 (CPF/CNPJ), None';

COMMIT;