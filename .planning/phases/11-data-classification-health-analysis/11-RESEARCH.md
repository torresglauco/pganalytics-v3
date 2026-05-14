# Phase 11: Data Classification & Health Analysis - Research

**Researched:** 2026-05-14
**Domain:** PII/PCI Data Detection, Host Health Scoring, Multi-Tenancy Isolation, Version-Specific Health Checks
**Confidence:** HIGH (based on existing codebase patterns)

## Summary

Phase 11 extends the collector-backend architecture from Phase 10 to add sensitive data classification (PII/PCI detection), host health scoring, and multi-tenancy scalability infrastructure. The collector (C++) will gain a new `data_classification_plugin` for pattern-based detection of sensitive data. The backend (Go) will add health score calculation, data classification storage/reporting, and tenant isolation mechanisms.

**Primary recommendation:** Create new collector plugin for data classification following the schema_plugin pattern. Implement host health scoring as a calculated metric from existing HostMetrics. Add tenant_id to collectors table for multi-tenancy isolation with row-level security policies.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DATA-01 | View PII detection results (CPF, CNPJ, email, phone, names) | New `data_classification_plugin.cpp` using regex patterns; backend models for classification results |
| DATA-02 | View PCI detection results (credit card numbers) | Same plugin with Luhn algorithm validation for credit card detection |
| DATA-03 | View LGPD/GDPR regulated data identification | Classification categories with regulation mapping (LGPD Article 5, GDPR Article 9) |
| DATA-04 | Configure custom detection patterns | JSON configuration for custom regex patterns stored in database |
| DATA-05 | View data classification reports by database/table | Aggregation queries on classification results table |
| HOST-04 | View host health score based on resource utilization | Calculated score from existing HostMetrics (CPU, memory, disk, load) with weighted formula |
| VER-03 | View version-specific health checks | Extension of VersionCapabilities with health check queries per version |
| SCALE-01 | Support 2000+ PostgreSQL clusters | Multi-tenancy with tenant_id column, RLS policies, connection pool scaling |
| SCALE-02 | Support 5000+ monitored hosts | Same infrastructure, partition by tenant_id for query performance |
| SCALE-03 | Support sharding/partitioning by tenant/cluster | PostgreSQL declarative partitioning on tenant_id |
| SCALE-04 | Support multi-tenancy with logical isolation | Row-Level Security (RLS) policies on all metric tables |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| pgx v5 | 5.9.2 | PostgreSQL driver with connection pooling | Already in use, provides 2-3x performance over lib/pq |
| Gin | 1.10.0 | HTTP web framework | Standard across all backend handlers |
| libpq | PG 11-17 | PostgreSQL client library for C++ collector | Standard for native PG connectivity |
| nlohmann/json | latest | JSON serialization for C++ | Already used in all collector plugins |
| regex (C++11) | standard | Pattern matching for PII/PCI detection | Built-in, no external dependency |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| TimescaleDB | 2.15.0-pg16 | Time-series storage for metrics | For historical classification/health data |
| uuid | google/uuid v1.6.0 | UUID generation | Tenant IDs, classification run IDs |
| zap | 1.27.0 | Structured logging | Backend logging |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| C++ regex | PCRE2 | PCRE2 adds external dependency; std::regex sufficient for pattern matching |
| Calculated health score | ML-based scoring | ML adds complexity, maintenance burden; weighted formula is explainable and tunable |
| RLS for isolation | Schema-per-tenant | Schema-per-tenant requires more migrations; RLS is simpler to manage |

**Installation:**
```bash
# Go dependencies already in go.mod
go mod download

# C++ collector dependencies managed via CMake and vcpkg
# No new dependencies needed for regex (C++11 std::regex)
```

**Version verification:**
```bash
# Go modules
go list -m github.com/jackc/pgx/v5  # v5.9.2
go list -m github.com/gin-gonic/gin  # v1.10.0
```

## Architecture Patterns

### Recommended Project Structure
```
collector/
├── include/
│   ├── data_classification_plugin.h  # NEW for Phase 11
│   └── ...
├── src/
│   ├── data_classification_plugin.cpp  # NEW for Phase 11
│   └── ...
└── CMakeLists.txt

backend/
├── internal/
│   ├── api/
│   │   ├── handlers_data_classification.go  # NEW for Phase 11
│   │   ├── handlers_health_score.go         # NEW for Phase 11
│   │   └── handlers_tenant.go               # NEW for Phase 11
│   ├── storage/
│   │   ├── classification_store.go         # NEW for Phase 11
│   │   ├── health_store.go                 # NEW for Phase 11
│   │   └── tenant_store.go                 # NEW for Phase 11
│   └── services/
│       └── health_score_calculator.go      # NEW for Phase 11
├── pkg/models/
│   ├── classification_models.go            # NEW for Phase 11
│   ├── health_models.go                    # NEW for Phase 11
│   └── tenant_models.go                    # NEW for Phase 11
└── migrations/
    └── 032_data_classification_tables.sql  # NEW for Phase 11
```

### Pattern 1: Data Classification Collector Plugin (C++)
**What:** C++ collector plugin that scans column values for sensitive data patterns
**When to use:** All PII/PCI detection requirements (DATA-01, DATA-02, DATA-03)
**Example:**
```cpp
// Source: Pattern from collector/src/schema_plugin.cpp
// NEW FILE: collector/src/data_classification_plugin.cpp

class PgDataClassificationCollector : public Collector {
public:
    json execute() override {
        json result = {
            {"type", "pg_data_classification"},
            {"timestamp", getCurrentTimestamp()},
            {"collector_id", collectorId_},
            {"databases", json::object()}
        };

        for (const auto& dbname : databases_) {
            result["databases"][dbname] = classifyDatabase(dbname);
        }
        return result;
    }

private:
    json classifyDatabase(const std::string& dbname) {
        // Get all tables and columns from schema
        json tables = getTableColumns(dbname);
        json classifications = json::array();

        for (const auto& table : tables) {
            for (const auto& column : table["columns"]) {
                json detected = detectSensitiveData(
                    dbname, table["schema"], table["name"],
                    column["name"], column["data_type"]
                );
                if (!detected.empty()) {
                    classifications.push_back(detected);
                }
            }
        }
        return classifications;
    }

    json detectSensitiveData(
        const std::string& dbname,
        const std::string& schema,
        const std::string& table,
        const std::string& column,
        const std::string& dataType
    ) {
        json result = json::object();
        std::string patternType;
        std::string category;

        // Sample column values for pattern detection
        std::string query = "SELECT " + column + " FROM " + schema + "." + table +
                           " WHERE " + column + " IS NOT NULL LIMIT 100";

        // Run pattern detection
        if (detectCPF(query)) {
            patternType = "CPF";
            category = "PII";
        } else if (detectCNPJ(query)) {
            patternType = "CNPJ";
            category = "PII";
        } else if (detectCreditCard(query)) {
            patternType = "CREDIT_CARD";
            category = "PCI";
        } else if (detectEmail(query)) {
            patternType = "EMAIL";
            category = "PII";
        } else if (detectPhone(query)) {
            patternType = "PHONE";
            category = "PII";
        }

        if (!patternType.empty()) {
            result["database"] = dbname;
            result["schema"] = schema;
            result["table"] = table;
            result["column"] = column;
            result["pattern_type"] = patternType;
            result["category"] = category;
            result["confidence"] = calculateConfidence(patternType);
        }

        return result;
    }

    // Brazilian CPF validation (11 digits with check digits)
    bool detectCPF(const std::string& query) {
        // Regex: \d{3}\.?\d{3}\.?\d{3}-?\d{2}
        // Plus validate check digits algorithm
        static const std::regex cpf_regex(R"(\d{3}\.?\d{3}\.?\d{3}-?\d{2})");
        // Implementation: sample values and validate CPF algorithm
        return false; // placeholder
    }

    // Brazilian CNPJ validation (14 digits with check digits)
    bool detectCNPJ(const std::string& query) {
        // Regex: \d{2}\.?\d{3}\.?\d{3}/?\d{4}-?\d{2}
        static const std::regex cnpj_regex(R"(\d{2}\.?\d{3}\.?\d{3}/?\d{4}-?\d{2})");
        return false; // placeholder
    }

    // Credit card detection with Luhn validation
    bool detectCreditCard(const std::string& query) {
        // Regex patterns for major card types
        // Visa: 4[0-9]{12}(?:[0-9]{3})?
        // MasterCard: 5[1-5][0-9]{14}
        // Amex: 3[47][0-9]{13}
        // Plus Luhn algorithm validation
        return false; // placeholder
    }

    bool validateLuhn(const std::string& number) {
        // Luhn algorithm implementation
        int sum = 0;
        bool alternate = false;
        for (int i = number.length() - 1; i >= 0; --i) {
            int digit = number[i] - '0';
            if (alternate) {
                digit *= 2;
                if (digit > 9) digit -= 9;
            }
            sum += digit;
            alternate = !alternate;
        }
        return (sum % 10) == 0;
    }
};
```

### Pattern 2: Host Health Score Calculation (Go)
**What:** Weighted scoring formula based on resource utilization metrics
**When to use:** HOST-04 host health score requirement
**Example:**
```go
// Source: Pattern from backend/internal/storage/host_store.go
// NEW FILE: backend/internal/services/health_score_calculator.go

package services

import (
    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// HealthScoreWeights defines the weight of each metric in the overall score
type HealthScoreWeights struct {
    CPU          float64 // 30% weight
    Memory       float64 // 25% weight
    Disk         float64 // 25% weight
    LoadAverage  float64 // 20% weight
}

// DefaultHealthScoreWeights provides standard weights
var DefaultHealthScoreWeights = HealthScoreWeights{
    CPU:         0.30,
    Memory:      0.25,
    Disk:        0.25,
    LoadAverage: 0.20,
}

// CalculateHostHealthScore computes a health score from 0-100
func CalculateHostHealthScore(metrics *models.HostMetrics, weights HealthScoreWeights) int {
    // CPU Score: lower usage = higher score
    cpuScore := 100.0 - metrics.CpuUser - metrics.CpuSystem - metrics.CpuIowait
    if cpuScore < 0 {
        cpuScore = 0
    }

    // Memory Score: lower usage percentage = higher score
    memoryScore := 100.0 - metrics.MemoryUsedPercent
    if memoryScore < 0 {
        memoryScore = 0
    }

    // Disk Score: lower usage percentage = higher score
    diskScore := 100.0 - metrics.DiskUsedPercent
    if diskScore < 0 {
        diskScore = 0
    }

    // Load Average Score: based on CPU cores
    // Load > cores = degraded, load > 2*cores = critical
    loadScore := 100.0
    if metrics.CpuLoad1m > float64(metrics.CpuCores) {
        loadScore = 50.0
    }
    if metrics.CpuLoad1m > float64(metrics.CpuCores)*2 {
        loadScore = 0
    }

    // Weighted average
    totalScore := (cpuScore * weights.CPU) +
                  (memoryScore * weights.Memory) +
                  (diskScore * weights.Disk) +
                  (loadScore * weights.LoadAverage)

    return int(totalScore)
}

// GetHealthStatus returns a status string based on score
func GetHealthStatus(score int) string {
    switch {
    case score >= 80:
        return "healthy"
    case score >= 60:
        return "degraded"
    case score >= 40:
        return "warning"
    default:
        return "critical"
    }
}
```

### Pattern 3: Multi-Tenancy with Row-Level Security (SQL)
**What:** PostgreSQL RLS policies for tenant isolation
**When to use:** SCALE-01, SCALE-02, SCALE-03, SCALE-04 multi-tenancy requirements
**Example:**
```sql
-- Source: Pattern for multi-tenancy isolation
-- NEW FILE: backend/migrations/032_data_classification_tables.sql

-- Add tenant_id to collectors table (SCALE-04)
ALTER TABLE collectors ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Create index for tenant-based queries
CREATE INDEX IF NOT EXISTS idx_collectors_tenant ON collectors(tenant_id);

-- Enable RLS on all metric tables
ALTER TABLE metrics_host_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics_host_inventory ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics_replication_status ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics_table_inventory ENABLE ROW LEVEL SECURITY;

-- Create RLS policy: users can only see data from their tenant's collectors
CREATE POLICY tenant_isolation_policy ON metrics_host_metrics
    USING (collector_id IN (
        SELECT id FROM collectors WHERE tenant_id = current_setting('app.current_tenant')::uuid
    ));

-- Set tenant context per session (called by backend after authentication)
-- SET app.current_tenant = 'tenant-uuid-here';
```

### Pattern 4: Data Classification Storage (Go)
**What:** Backend storage for classification results with confidence scoring
**When to use:** DATA-01 to DATA-05 requirements
**Example:**
```go
// Source: Pattern from backend/internal/storage/host_store.go
// NEW FILE: backend/internal/storage/classification_store.go

package storage

import (
    "context"
    "github.com/google/uuid"
    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// StoreClassificationResults inserts classification results from collector
func (p *PostgresDB) StoreClassificationResults(ctx context.Context, results []*models.DataClassificationResult) error {
    if len(results) == 0 {
        return nil
    }

    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return apperrors.DatabaseError("begin transaction", err.Error())
    }
    defer func() { _ = tx.Rollback() }()

    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO metrics_data_classification (
            time, collector_id, database_name, schema_name, table_name, column_name,
            pattern_type, category, confidence, match_count, sample_values
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (time, collector_id, database_name, schema_name, table_name, column_name)
        DO UPDATE SET
            pattern_type = EXCLUDED.pattern_type,
            category = EXCLUDED.category,
            confidence = EXCLUDED.confidence,
            match_count = EXCLUDED.match_count
    `)
    if err != nil {
        return apperrors.DatabaseError("prepare classification insert", err.Error())
    }
    defer func() { _ = stmt.Close() }()

    for _, r := range results {
        _, err := stmt.ExecContext(ctx,
            r.Time, r.CollectorID, r.DatabaseName, r.SchemaName, r.TableName, r.ColumnName,
            r.PatternType, r.Category, r.Confidence, r.MatchCount, r.SampleValues,
        )
        if err != nil {
            return apperrors.DatabaseError("insert classification result", err.Error())
        }
    }

    return tx.Commit()
}
```

### Anti-Patterns to Avoid
- **Don't scan entire tables:** Use LIMIT and sampling for classification to avoid performance impact
- **Don't store raw sensitive values:** Store only metadata (pattern type, confidence, count)
- **Don't use application-level tenant filtering:** Use RLS for defense in depth
- **Don't hardcode detection patterns:** Load from configuration table for flexibility (DATA-04)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Credit card validation | Custom checksum logic | Luhn algorithm (standard implementation) | Industry standard, handles all card types |
| CPF/CNPJ validation | Custom digit check | Brazilian government algorithm | Official validation rules, avoids false positives |
| Tenant isolation | Application-level filtering | PostgreSQL Row-Level Security | Database-enforced, cannot be bypassed |
| Health score formula | Complex ML model | Weighted average formula | Explainable, tunable, maintainable |
| Regex pattern matching | String parsing | C++11 std::regex | Built-in, no dependencies, well-tested |

**Key insight:** Use established algorithms for data validation (Luhn, CPF/CNPJ check digits). Use PostgreSQL features for isolation (RLS). Keep health scoring simple and explainable.

## Common Pitfalls

### Pitfall 1: Classification Performance Impact
**What goes wrong:** Full table scans for pattern detection slow down production database
**Why it happens:** Scanning every row and column for patterns is I/O intensive
**How to avoid:**
- Use LIMIT clauses (sample 100-1000 rows per column)
- Run classification during low-traffic periods
- Exclude known safe columns (primary keys, timestamps, IDs)
**Warning signs:** Query duration > 1 second, CPU spike on monitored database

### Pitfall 2: False Positives in Pattern Detection
**What goes wrong:** Valid data incorrectly flagged as sensitive (e.g., sequential IDs matching credit card regex)
**Why it happens:** Regex patterns are not foolproof without validation algorithms
**How to avoid:**
- Always validate with check digit algorithms (Luhn for credit cards, mod 11 for CPF/CNPJ)
- Use confidence scoring (100% for validated, 50% for pattern-only)
- Allow users to mark false positives
**Warning signs:** High volume of classifications, user complaints about false positives

### Pitfall 3: RLS Performance Overhead
**What goes wrong:** RLS policies cause slow queries on large metric tables
**Why it happens:** Subquery in RLS policy runs for each row
**How to avoid:**
- Create functional index on `current_setting('app.current_tenant')`
- Cache tenant-collector mapping in application layer
- Use materialized views for frequently accessed aggregations
**Warning signs:** Query times > 500ms for dashboard queries

### Pitfall 4: Tenant Context Not Set
**What goes wrong:** Users see no data or wrong data after RLS enabled
**Why it happens:** Application didn't set `app.current_tenant` session variable
**How to avoid:**
- Set tenant context immediately after authentication in middleware
- Add audit logging for tenant context changes
- Return 403 Forbidden if tenant context missing
**Warning signs:** Empty dashboard, "permission denied" errors

## Code Examples

### PII Detection Patterns (C++)
```cpp
// Source: Standard patterns for Brazilian PII detection
// CPF: XXX.XXX.XXX-XX (11 digits, mod 11 validation)
// CNPJ: XX.XXX.XXX/XXXX-XX (14 digits, mod 11 validation)

bool validateCPF(const std::string& cpf) {
    // Remove non-digits
    std::string digits;
    for (char c : cpf) {
        if (std::isdigit(c)) digits += c;
    }
    if (digits.length() != 11) return false;

    // Validate check digits (mod 11 algorithm)
    int sum1 = 0, sum2 = 0;
    for (int i = 0; i < 9; i++) {
        int digit = digits[i] - '0';
        sum1 += digit * (10 - i);
        sum2 += digit * (11 - i);
    }
    int check1 = (sum1 * 10) % 11;
    if (check1 == 10) check1 = 0;

    sum2 += check1 * 2;
    int check2 = (sum2 * 10) % 11;
    if (check2 == 10) check2 = 0;

    return (check1 == digits[9] - '0') && (check2 == digits[10] - '0');
}
```

### Credit Card Detection with Luhn (C++)
```cpp
// Source: Luhn algorithm (ISO/IEC 7812)
bool validateCreditCard(const std::string& number) {
    std::string digits;
    for (char c : number) {
        if (std::isdigit(c)) digits += c;
    }
    if (digits.length() < 13 || digits.length() > 19) return false;

    // Luhn algorithm
    int sum = 0;
    bool alternate = false;
    for (int i = digits.length() - 1; i >= 0; --i) {
        int digit = digits[i] - '0';
        if (alternate) {
            digit *= 2;
            if (digit > 9) digit -= 9;
        }
        sum += digit;
        alternate = !alternate;
    }
    return (sum % 10) == 0;
}

// Card type detection
std::string detectCardType(const std::string& number) {
    std::string digits;
    for (char c : number) {
        if (std::isdigit(c)) digits += c;
    }

    // Visa: starts with 4, length 13, 16, or 19
    if (digits[0] == '4' && (digits.length() == 13 || digits.length() == 16 || digits.length() == 19)) {
        return "VISA";
    }
    // MasterCard: starts with 51-55 or 2221-2720, length 16
    if (digits.length() == 16) {
        int prefix = std::stoi(digits.substr(0, 2));
        if ((prefix >= 51 && prefix <= 55) || (prefix >= 22 && prefix <= 27)) {
            return "MASTERCARD";
        }
    }
    // Amex: starts with 34 or 37, length 15
    if ((digits.substr(0, 2) == "34" || digits.substr(0, 2) == "37") && digits.length() == 15) {
        return "AMEX";
    }
    return "UNKNOWN";
}
```

### Database Migration for Classification Tables
```sql
-- Source: Pattern from backend/migrations/031_replication_tables.sql
-- NEW FILE: backend/migrations/032_data_classification_tables.sql

BEGIN;

-- Data Classification Results (DATA-01, DATA-02, DATA-03, DATA-05)
CREATE TABLE IF NOT EXISTS metrics_data_classification (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    database_name VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    column_name VARCHAR(255),
    pattern_type VARCHAR(50),      -- CPF, CNPJ, EMAIL, PHONE, CREDIT_CARD, CUSTOM
    category VARCHAR(20),          -- PII, PCI, SENSITIVE, CUSTOM
    confidence FLOAT,              -- 0.0 to 1.0
    match_count BIGINT,            -- Number of matching rows
    sample_values JSONB,           -- Anonymized samples (masked)
    regulation_mapping JSONB,      -- {"LGPD": ["Art. 5, I"], "GDPR": ["Art. 9"]}
    PRIMARY KEY (time, collector_id, database_name, schema_name, table_name, column_name)
);

SELECT create_hypertable('metrics_data_classification', 'time',
    if_not_exists => TRUE, migrate_data => FALSE);

CREATE INDEX idx_classification_collector ON metrics_data_classification(collector_id, time DESC);
CREATE INDEX idx_classification_pattern ON metrics_data_classification(pattern_type, category);

-- Custom Detection Patterns (DATA-04)
CREATE TABLE IF NOT EXISTS data_classification_patterns (
    id SERIAL PRIMARY KEY,
    tenant_id UUID,                -- NULL for global patterns
    pattern_name VARCHAR(255) NOT NULL,
    pattern_regex TEXT NOT NULL,
    category VARCHAR(20) NOT NULL, -- PII, PCI, SENSITIVE, CUSTOM
    validation_algorithm VARCHAR(50), -- Luhn, Mod11, None
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    enabled BOOLEAN DEFAULT TRUE
);

-- Host Health Scores (HOST-04)
CREATE TABLE IF NOT EXISTS metrics_host_health_scores (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    health_score INT,             -- 0-100
    status VARCHAR(20),           -- healthy, degraded, warning, critical
    cpu_score FLOAT,
    memory_score FLOAT,
    disk_score FLOAT,
    load_score FLOAT,
    calculation_details JSONB,    -- Factors that contributed to score
    PRIMARY KEY (time, collector_id)
);

SELECT create_hypertable('metrics_host_health_scores', 'time',
    if_not_exists => TRUE, migrate_data => FALSE);

CREATE INDEX idx_health_scores_collector ON metrics_host_health_scores(collector_id, time DESC);

-- Version-Specific Health Checks (VER-03)
CREATE TABLE IF NOT EXISTS postgres_health_checks (
    id SERIAL PRIMARY KEY,
    min_version INT NOT NULL,     -- Minimum PG version
    max_version INT,              -- Maximum PG version (NULL = no upper limit)
    check_name VARCHAR(255) NOT NULL,
    check_query TEXT NOT NULL,
    expected_result TEXT,         -- Description of expected result
    severity VARCHAR(20),         -- critical, warning, info
    description TEXT,
    UNIQUE(check_name, min_version)
);

-- Seed version-specific health checks
INSERT INTO postgres_health_checks (min_version, max_version, check_name, check_query, severity, description) VALUES
-- PostgreSQL 11-12 (EOL versions)
(11, 12, 'wal_keep_segments',
 'SELECT setting FROM pg_settings WHERE name = ''wal_keep_segments''',
 'warning', 'Check wal_keep_segments (deprecated in PG13)'),
-- PostgreSQL 13+
(13, NULL, 'wal_keep_size',
 'SELECT setting FROM pg_settings WHERE name = ''wal_keep_size''',
 'warning', 'Check WAL retention size'),
-- PostgreSQL 14+
(14, NULL, 'pg_stat_wal',
 'SELECT wal_records, wal_fpi FROM pg_stat_wal',
 'info', 'WAL statistics available in PG14+'),
-- PostgreSQL 17+
(17, NULL, 'logical_decoding_workers',
 'SELECT count(*) FROM pg_stat_activity WHERE backend_type = ''logical replication worker''',
 'info', 'Check parallel apply workers in PG17+');

-- Retention policies
SELECT add_retention_policy('metrics_data_classification', INTERVAL '90 days', if_not_exists => TRUE);
SELECT add_retention_policy('metrics_host_health_scores', INTERVAL '90 days', if_not_exists => TRUE);

COMMIT;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Application-level tenant filtering | PostgreSQL RLS | Phase 11 | Defense in depth, cannot be bypassed |
| Manual PII audits | Automated pattern detection | Phase 11 | Continuous compliance, reduced audit cost |
| Static health checks | Version-adaptive queries | Phase 11 | Works across PG 11-17 |
| Single-tenant schema | tenant_id with RLS | Phase 11 | Scales to 2000+ clusters |

**Deprecated/outdated:**
- wal_keep_segments (PG 11-12): Use wal_keep_size in PG 13+
- procpid column (PG 9.x): Use pid in pg_stat_activity

## Open Questions

1. **Classification Sampling Strategy**
   - What we know: LIMIT 100-1000 rows per column is standard practice
   - What's unclear: Optimal sample size for confidence vs performance
   - Recommendation: Start with LIMIT 1000, make configurable per deployment

2. **Custom Pattern Configuration UI**
   - What we know: DATA-04 requires configurable custom patterns
   - What's unclear: Whether patterns are tenant-specific or global
   - Recommendation: Support both tenant-specific and global patterns

3. **Health Score Thresholds**
   - What we know: 80+ = healthy, 60-80 = degraded, <60 = warning/critical
   - What's unclear: Should thresholds be configurable per tenant
   - Recommendation: Start with global thresholds, add tenant override in future

4. **RLS Index Strategy**
   - What we know: Subquery in RLS can be slow without index
   - What's unclear: Best index strategy for tenant-collector lookup in RLS
   - Recommendation: Create index on collectors(tenant_id, id), benchmark with 2000+ clusters

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go: testing + testify; C++: Catch2 (via CMake) |
| Config file | Go: none (tests self-contained); C++: collector/tests/CMakeLists.txt |
| Quick run command | `go test ./... -short` (Go); `cd collector/build && ctest` (C++) |
| Full suite command | `go test ./... -cover -race` (Go); `cd collector/build && ctest -V` (C++) |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DATA-01 | View PII detection results (CPF, CNPJ, email, phone) | unit | `go test ./internal/storage -run TestClassification` | No - Wave 0 |
| DATA-02 | View PCI detection results (credit cards) | unit | `go test ./internal/storage -run TestPCIDetection` | No - Wave 0 |
| DATA-03 | View LGPD/GDPR regulated data identification | integration | `go test ./tests/integration -run TestRegulationMapping` | No - Wave 0 |
| DATA-04 | Configure custom detection patterns | integration | `go test ./tests/integration -run TestCustomPatterns` | No - Wave 0 |
| DATA-05 | View data classification reports by database/table | integration | `go test ./tests/integration -run TestClassificationReports` | No - Wave 0 |
| HOST-04 | View host health score | unit | `go test ./internal/services -run TestHealthScoreCalculator` | No - Wave 0 |
| VER-03 | View version-specific health checks | unit | `go test ./internal/storage -run TestVersionHealthChecks` | No - Wave 0 |
| SCALE-01 | Support 2000+ PostgreSQL clusters | integration | `go test ./tests/integration -run TestMultiTenancy` | No - Wave 0 |
| SCALE-02 | Support 5000+ monitored hosts | integration | `go test ./tests/integration -run TestHostScaling` | No - Wave 0 |
| SCALE-03 | Support sharding/partitioning by tenant | unit | `go test ./internal/storage -run TestTenantPartitioning` | No - Wave 0 |
| SCALE-04 | Support multi-tenancy with logical isolation | security | `go test ./tests/security -run TestRLSIsolation` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/storage ./internal/services -short -v`
- **Per wave merge:** `go test ./... -cover -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `backend/internal/services/health_score_calculator.go` - Health scoring logic
- [ ] `backend/internal/storage/classification_store.go` - Classification storage
- [ ] `backend/internal/storage/health_store.go` - Health score storage
- [ ] `backend/internal/storage/tenant_store.go` - Tenant management
- [ ] `backend/internal/api/handlers_data_classification.go` - Classification endpoints
- [ ] `backend/internal/api/handlers_health_score.go` - Health score endpoints
- [ ] `backend/internal/api/handlers_tenant.go` - Tenant management endpoints
- [ ] `backend/pkg/models/classification_models.go` - Classification data structures
- [ ] `backend/pkg/models/health_models.go` - Health score data structures
- [ ] `backend/pkg/models/tenant_models.go` - Tenant data structures
- [ ] `backend/migrations/032_data_classification_tables.sql` - Database schema
- [ ] `backend/tests/integration/classification_test.go` - Classification integration tests
- [ ] `backend/tests/integration/health_score_test.go` - Health score integration tests
- [ ] `backend/tests/integration/multi_tenant_test.go` - Multi-tenancy tests
- [ ] `backend/tests/security/rls_test.go` - Row-Level Security tests
- [ ] `collector/include/data_classification_plugin.h` - C++ header for classification
- [ ] `collector/src/data_classification_plugin.cpp` - C++ implementation
- [ ] `collector/tests/test_data_classification.cpp` - C++ unit tests

## Sources

### Primary (HIGH confidence)
- Existing codebase analysis: `collector/src/schema_plugin.cpp` (pattern for database scanning)
- Existing codebase analysis: `backend/internal/storage/host_store.go` (pattern for metrics storage)
- Existing codebase analysis: `backend/internal/storage/version_store.go` (pattern for version detection)
- Database schema: `backend/migrations/031_replication_tables.sql` (TimescaleDB pattern)

### Secondary (MEDIUM confidence)
- PostgreSQL documentation for Row-Level Security: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- Luhn algorithm specification: ISO/IEC 7812
- Brazilian CPF/CNPJ validation: Receita Federal specifications

### Tertiary (LOW confidence)
- Web search for multi-tenancy patterns - need to verify with TimescaleDB documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use from Phase 10
- Architecture: HIGH - Clear patterns from existing plugins and handlers
- Pitfalls: MEDIUM - Multi-tenancy RLS performance needs benchmarking
- Data classification: MEDIUM - Pattern detection confidence depends on sample size

**Research date:** 2026-05-14
**Valid until:** 30 days - PostgreSQL RLS patterns stable, CPF/CNPJ validation rules stable