-- Migration 035: Version-Specific Health Checks
-- Stores health check definitions for PostgreSQL versions 11-17
-- Includes adaptive queries, severity levels, and remediation suggestions

BEGIN;

-- ============================================================================
-- POSTGRES HEALTH CHECKS TABLE (VER-03)
-- ============================================================================

-- Create table for version-specific health check definitions
CREATE TABLE IF NOT EXISTS postgres_health_checks (
    id SERIAL PRIMARY KEY,
    min_version INT NOT NULL,                -- Minimum PostgreSQL major version
    max_version INT,                         -- Maximum PostgreSQL major version (NULL = no upper limit)
    check_name VARCHAR(255) NOT NULL,        -- Unique check identifier
    check_query TEXT NOT NULL,               -- SQL query to execute on monitored database
    expected_result TEXT,                    -- Description of expected result
    severity VARCHAR(20) NOT NULL,           -- critical, warning, info
    description TEXT,                        -- What this check does
    remediation TEXT,                        -- How to fix issues
    category VARCHAR(50) DEFAULT 'configuration', -- performance, security, configuration, replication, monitoring
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(check_name, min_version)
);

-- Create indexes for version-based lookups
CREATE INDEX IF NOT EXISTS idx_health_checks_version
    ON postgres_health_checks (min_version, max_version);

CREATE INDEX IF NOT EXISTS idx_health_checks_category
    ON postgres_health_checks (category);

CREATE INDEX IF NOT EXISTS idx_health_checks_severity
    ON postgres_health_checks (severity);

-- ============================================================================
-- SEED DATA: PostgreSQL 11-12 (EOL VERSIONS - MIGRATION WARNINGS)
-- ============================================================================

-- EOL warning for PostgreSQL 11-12
INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, expected_result, severity, description, remediation, category) VALUES
(11, 12, 'version_eol_warning',
 'SELECT current_setting(''server_version_num'')::int',
 'Version is End of Life',
 'critical', 'PostgreSQL 11-12 are End of Life versions',
 'Upgrade to PostgreSQL 15 or 16 for continued security updates',
 'security'),
(11, 12, 'wal_keep_segments_deprecated',
 'SELECT setting FROM pg_settings WHERE name = ''wal_keep_segments''',
 'Setting exists but deprecated in PG13+',
 'warning', 'wal_keep_segments is deprecated in PostgreSQL 13+',
 'Migrate to wal_keep_size before upgrading to PG13+',
 'configuration'),
(11, 12, 'pg_stat_replication_sync_state',
 'SELECT count(*) FROM pg_stat_replication WHERE sync_state = ''sync''',
 'Synchronous replicas count',
 'info', 'Check synchronous replication setup',
 'Ensure appropriate number of synchronous standbys',
 'replication');

-- ============================================================================
-- SEED DATA: PostgreSQL 13+ (ACTIVE VERSIONS)
-- ============================================================================

INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, expected_result, severity, description, remediation, category) VALUES
(13, NULL, 'wal_keep_size',
 'SELECT setting FROM pg_settings WHERE name = ''wal_keep_size''',
 'WAL retention size in MB',
 'warning', 'Check WAL retention size for replication',
 'Ensure sufficient WAL retention for standbys',
 'configuration'),
(13, NULL, 'wal_compression',
 'SELECT setting FROM pg_settings WHERE name = ''wal_compression''',
 'on or lz4 or zstd',
 'info', 'WAL compression can reduce I/O',
 'Enable wal_compression for better performance',
 'performance');

-- ============================================================================
-- SEED DATA: PostgreSQL 14+
-- ============================================================================

INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, expected_result, severity, description, remediation, category) VALUES
(14, NULL, 'pg_stat_wal_available',
 'SELECT count(*) FROM pg_stat_wal',
 'WAL statistics available',
 'info', 'PostgreSQL 14+ provides pg_stat_wal view',
 'Monitor WAL generation and I/O patterns',
 'monitoring'),
(14, NULL, 'logical_replication_workers',
 'SELECT count(*) FROM pg_stat_activity WHERE backend_type = ''logical replication worker''',
 'Logical replication workers count',
 'info', 'Check parallel apply workers in PG14+',
 'Monitor logical replication worker usage',
 'replication');

-- ============================================================================
-- SEED DATA: PostgreSQL 15+
-- ============================================================================

INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, expected_result, severity, description, remediation, category) VALUES
(15, NULL, 'logical_decoding_work_mem',
 'SELECT setting FROM pg_settings WHERE name = ''logical_decoding_work_mem''',
 'Memory for logical decoding',
 'info', 'PG15 allows tuning logical decoding memory',
 'Adjust for large transactions in logical replication',
 'configuration');

-- ============================================================================
-- SEED DATA: PostgreSQL 17+
-- ============================================================================

INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, expected_result, severity, description, remediation, category) VALUES
(17, NULL, 'parallel_apply_workers',
 'SELECT count(*) FROM pg_stat_activity WHERE backend_type LIKE ''parallel apply%''',
 'Parallel apply workers count',
 'info', 'PG17 supports parallel logical replication apply',
 'Monitor parallel apply for performance gains',
 'replication');

-- ============================================================================
-- HEALTH CHECK RESULTS TABLE (OPTIONAL - FOR STORING EXECUTION HISTORY)
-- ============================================================================

-- Table to store health check execution results
CREATE TABLE IF NOT EXISTS postgres_health_check_results (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    check_id INT NOT NULL REFERENCES postgres_health_checks(id),
    passed BOOLEAN NOT NULL,
    actual_result TEXT,
    checked_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for querying results by collector
CREATE INDEX IF NOT EXISTS idx_health_check_results_collector
    ON postgres_health_check_results (collector_id, checked_at DESC);

-- Create index for querying by check
CREATE INDEX IF NOT EXISTS idx_health_check_results_check
    ON postgres_health_check_results (check_id, checked_at DESC);

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE postgres_health_checks IS 'Version-specific health check definitions for PostgreSQL 11-17';
COMMENT ON TABLE postgres_health_check_results IS 'History of health check execution results per collector';

COMMENT ON COLUMN postgres_health_checks.min_version IS 'Minimum PostgreSQL major version this check applies to';
COMMENT ON COLUMN postgres_health_checks.max_version IS 'Maximum PostgreSQL major version (NULL means no upper limit)';
COMMENT ON COLUMN postgres_health_checks.check_query IS 'SQL query to execute on the monitored database';
COMMENT ON COLUMN postgres_health_checks.severity IS 'Severity level: critical, warning, or info';
COMMENT ON COLUMN postgres_health_checks.category IS 'Check category: performance, security, configuration, replication, or monitoring';

COMMIT;