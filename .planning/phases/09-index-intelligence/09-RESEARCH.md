# Phase 09: Index Intelligence - Research

**Researched:** 2026-05-12
**Domain:** PostgreSQL query plan analysis, index recommendation, and fingerprinting
**Confidence:** HIGH

## Summary

Phase 09 implements intelligent index recommendations with impact estimation. The existing codebase has foundational components (basic EXPLAIN parsing, cost calculator, database schemas) that need enhancement. The primary technical challenges are: (1) comprehensive EXPLAIN JSON parsing to detect anti-patterns, (2) query fingerprinting to group similar queries, (3) unused index detection from pg_stat_user_indexes, and (4) hypothetical index testing for impact estimation.

**Primary recommendation:** Use `pg_query_go/v5` for query fingerprinting, enhance the existing `QueryParser` to recursively walk EXPLAIN JSON, and leverage PostgreSQL's `hypopg` extension for safe index impact estimation without creating actual indexes.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QRY-03 | User receives automated detection of query plan anti-patterns (Seq Scan, nested loops) | Enhance existing `QueryParser.DetectIssues()` with recursive JSON traversal. Add detection for: Seq Scan, Nested Loop with high rows, Hash Join spills, Sort with large work_mem. |
| QRY-04 | User can view grouped similar queries with different parameters (fingerprinting) | Use `pg_query_go/v5` library with `Fingerprint()` function to normalize queries. Store fingerprint hash in `query_plans.query_fingerprint_hash`. |
| IDX-02 | User can see unused indexes that are candidates for removal | Query `pg_stat_user_indexes` for `idx_scan = 0`. Exclude primary keys, unique constraints, and foreign keys. Order by `pg_relation_size()` to prioritize large indexes. |
| IDX-03 | User receives index impact estimation before creating new indexes | Use `hypopg` PostgreSQL extension for hypothetical indexes. Run `EXPLAIN` with and without hypothetical index to compare costs. |
| IDX-04 | User can view recommended indexes with estimated benefit scores | Combine cost improvement (from hypopg) with query frequency (from pg_stat_statements) to calculate benefit score. Store in `index_recommendations` table. |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| pg_query_go/v5 | v5.1.0 | PostgreSQL query parsing and fingerprinting | Official Go bindings for libpg_query. Provides `Parse()`, `Fingerprint()`, `Normalize()` functions. Actively maintained by pganalyze. |
| pgx/v5 | v5.9.2 (already installed) | PostgreSQL driver with native connection pooling | Already in use. Use for all database operations. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | EXPLAIN JSON parsing | Parse PostgreSQL EXPLAIN (FORMAT JSON) output |
| github.com/lib/pq | v1.10.9 (already installed) | pq.Array for TEXT[] support | Keep for array parameter handling in index column names |

### Database Extensions
| Extension | Purpose | Installation |
|-----------|---------|--------------|
| pg_stat_statements | Query statistics (already used) | CREATE EXTENSION pg_stat_statements; |
| hypopg | Hypothetical indexes for impact testing | CREATE EXTENSION hypopg; (requires superuser) |

**Installation:**
```bash
go get github.com/pganalyze/pg_query_go/v5@v5.1.0
```

**Version verification:**
```
pg_query_go/v5: v5.1.0 (verified via go list -m -versions)
pgx/v5: v5.9.2 (already installed)
```

## Architecture Patterns

### Recommended Project Structure
```
backend/internal/
├── services/
│   ├── query_performance/
│   │   ├── parser.go           # Enhance: recursive EXPLAIN JSON parsing
│   │   ├── fingerprint.go      # NEW: query fingerprinting with pg_query_go
│   │   └── analyzer.go         # Enhance: severity scoring
│   └── index_advisor/
│       ├── analyzer.go         # Enhance: full implementation
│       ├── cost_calculator.go  # Keep: benefit calculations
│       ├── hypo_index.go       # NEW: hypopg integration
│       └── unused_detector.go  # NEW: unused index detection
├── storage/
│   ├── query_performance_store.go  # Enhance: fingerprint queries
│   └── index_recommendation_store.go # NEW: storage for recommendations
└── api/
    ├── handlers_query_performance.go # Enhance: fingerprint endpoint
    └── handlers_index_advisor.go     # Enhance: impact estimation
```

### Pattern 1: Query Fingerprinting
**What:** Normalize queries by removing literal values, replacing with placeholders
**When to use:** Grouping similar queries with different parameters for aggregate analysis

**Example:**
```go
// Source: pg_query_go documentation
import "github.com/pganalyze/pg_query_go/v5"

func FingerprintQuery(queryText string) (string, error) {
    result, err := pg_query.Fingerprint(queryText)
    if err != nil {
        return "", err
    }
    return result, nil
}

// Example:
// SELECT * FROM users WHERE id = 123
// SELECT * FROM users WHERE id = 456
// Both fingerprint to: SELECT * FROM users WHERE id = $1
```

**Implementation:**
```go
// backend/internal/services/query_performance/fingerprint.go
package query_performance

import (
    "crypto/sha256"
    "encoding/hex"
    "github.com/pganalyze/pg_query_go/v5"
)

type Fingerprinter struct{}

func NewFingerprinter() *Fingerprinter {
    return &Fingerprinter{}
}

// Fingerprint returns a normalized hash of the query
func (f *Fingerprinter) Fingerprint(queryText string) (string, error) {
    fingerprint, err := pg_query.Fingerprint(queryText)
    if err != nil {
        return "", err
    }
    hash := sha256.Sum256([]byte(fingerprint))
    return hex.EncodeToString(hash[:16]), nil // 32-char hex
}

// Normalize returns the parameterized query text
func (f *Fingerprinter) Normalize(queryText string) (string, error) {
    return pg_query.Normalize(queryText)
}
```

### Pattern 2: Recursive EXPLAIN Plan Parsing
**What:** Walk the entire EXPLAIN JSON tree to detect anti-patterns at any level
**When to use:** Comprehensive query plan analysis for QRY-03

**Example:**
```go
// backend/internal/services/query_performance/parser.go
package query_performance

import "encoding/json"

// FullExplainPlan represents the complete EXPLAIN (FORMAT JSON) output
type FullExplainPlan struct {
    Plan         *PlanNode `json:"Plan"`
    PlanningTime float64   `json:"Planning Time"`
    ExecutionTime float64  `json:"Execution Time"`
}

// PlanNode represents a single node in the plan tree
type PlanNode struct {
    NodeType          string      `json:"Node Type"`
    TotalCost         float64     `json:"Total Cost"`
    StartupCost       float64     `json:"Startup Cost"`
    PlanRows          int64       `json:"Plan Rows"`
    PlanWidth         int         `json:"Plan Width"`
    ActualRows        int64       `json:"Actual Rows"`
    ActualLoops       int64       `json:"Actual Loops"`
    RelationName      string      `json:"Relation Name"`
    IndexName         string      `json:"Index Name"`
    Filter            string      `json:"Filter"`
    HashCond          string      `json:"Hash Cond"`
    JoinType          string      `json:"Join Type"`
    Plans             []*PlanNode `json:"Plans"`
}

// DetectIssues recursively finds all anti-patterns
func (qp *QueryParser) DetectIssuesFull(plan *FullExplainPlan) []QueryIssue {
    var issues []QueryIssue
    qp.walkPlan(plan.Plan, &issues)
    return issues
}

func (qp *QueryParser) walkPlan(node *PlanNode, issues *[]QueryIssue) {
    if node == nil {
        return
    }

    // Detect sequential scans on large tables
    if node.NodeType == "Seq Scan" && node.TotalCost > 100 {
        severity := "medium"
        if node.TotalCost > 1000 {
            severity = "high"
        }
        *issues = append(*issues, QueryIssue{
            Type:           "sequential_scan",
            Severity:       severity,
            AffectedNode:   node.RelationName,
            Description:    "Sequential scan on table with high cost",
            Recommendation: "Consider creating an index on filtered columns",
            EstimatedBenefit: node.TotalCost * 0.3,
        })
    }

    // Detect nested loops with high row counts
    if node.NodeType == "Nested Loop" && node.ActualRows > 1000 {
        *issues = append(*issues, QueryIssue{
            Type:           "nested_loop_high_rows",
            Severity:       "high",
            AffectedNode:   "Nested Loop",
            Description:    "Nested loop iterating over many rows",
            Recommendation: "Consider hash join or merge join via index optimization",
            EstimatedBenefit: float64(node.ActualRows) * 0.01,
        })
    }

    // Detect sorts with large datasets
    if node.NodeType == "Sort" && node.PlanRows > 10000 {
        *issues = append(*issues, QueryIssue{
            Type:           "large_sort",
            Severity:       "medium",
            AffectedNode:   "Sort",
            Description:    "Sorting large dataset without index",
            Recommendation: "Consider creating index on ORDER BY columns",
            EstimatedBenefit: node.TotalCost * 0.2,
        })
    }

    // Recursively walk child plans
    for _, child := range node.Plans {
        qp.walkPlan(child, issues)
    }
}
```

### Pattern 3: Unused Index Detection
**What:** Query pg_stat_user_indexes to find indexes never used for scans
**When to use:** IDX-02 requirement for index cleanup recommendations

**Implementation:**
```go
// backend/internal/services/index_advisor/unused_detector.go
package index_advisor

import (
    "context"
    "database/sql"
)

type UnusedIndexDetector struct {
    db *sql.DB
}

type UnusedIndex struct {
    SchemaName string
    TableName  string
    IndexName  string
    SizeBytes  int64
    IdxScan    int64
    IsPrimary  bool
    IsUnique   bool
}

func NewUnusedIndexDetector(db *sql.DB) *UnusedIndexDetector {
    return &UnusedIndexDetector{db: db}
}

func (d *UnusedIndexDetector) FindUnused(ctx context.Context) ([]UnusedIndex, error) {
    query := `
    SELECT
        schemaname,
        tablename,
        indexrelname as indexname,
        pg_relation_size(indexrelid) as size_bytes,
        COALESCE(idx_scan, 0) as idx_scan,
        contype = 'p' as is_primary,
        contype = 'u' as is_unique
    FROM pg_stat_user_indexes psui
    LEFT JOIN pg_constraint c ON c.conindid = psui.indexrelid
    WHERE idx_scan = 0
      AND contype IS NULL  -- Exclude primary keys, unique constraints
    ORDER BY pg_relation_size(indexrelid) DESC
    LIMIT 50
    `
    // Implementation follows existing patterns in query_performance_store.go
    return []UnusedIndex{}, nil
}
```

### Pattern 4: Hypothetical Index Impact Estimation
**What:** Use hypopg extension to test index impact without creating it
**When to use:** IDX-03 requirement for safe impact estimation

**Implementation:**
```go
// backend/internal/services/index_advisor/hypo_index.go
package index_advisor

import (
    "context"
    "database/sql"
    "fmt"
)

type HypoIndexTester struct {
    db *sql.DB
}

func NewHypoIndexTester(db *sql.DB) *HypoIndexTester {
    return &HypoIndexTester{db: db}
}

type IndexImpact struct {
    IndexName       string
    TableName       string
    Columns         []string
    CostWithout     float64
    CostWith        float64
    ImprovementPct  float64
    QueryCount      int
}

// EstimateImpact creates a hypothetical index and measures improvement
func (t *HypoIndexTester) EstimateImpact(
    ctx context.Context,
    queryText string,
    tableName string,
    columns []string,
) (*IndexImpact, error) {
    // 1. Get baseline cost
    var baselineCost float64
    err := t.db.QueryRowContext(ctx,
        fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", queryText),
    ).Scan(&baselineCost)
    if err != nil {
        return nil, err
    }

    // 2. Create hypothetical index
    colList := strings.Join(columns, ", ")
    hypoQuery := fmt.Sprintf(
        "SELECT hypopg_create_index('CREATE INDEX ON %s (%s)')",
        tableName, colList,
    )
    var indexName string
    err = t.db.QueryRowContext(ctx, hypoQuery).Scan(&indexName)
    if err != nil {
        // hypopg extension may not be installed
        return nil, fmt.Errorf("hypopg extension required: %w", err)
    }
    defer t.db.ExecContext(ctx, "SELECT hypopg_drop_index($1)", indexName)

    // 3. Get cost with hypothetical index
    var indexCost float64
    err = t.db.QueryRowContext(ctx,
        fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", queryText),
    ).Scan(&indexCost)

    // 4. Calculate improvement
    improvement := ((baselineCost - indexCost) / baselineCost) * 100

    return &IndexImpact{
        TableName:      tableName,
        Columns:        columns,
        CostWithout:    baselineCost,
        CostWith:       indexCost,
        ImprovementPct: improvement,
    }, nil
}
```

### Anti-Patterns to Avoid
- **Running EXPLAIN ANALYZE in production:** Use EXPLAIN (without ANALYZE) to avoid executing queries. ANALYZE executes the query which can be dangerous.
- **Creating indexes without testing:** Always use hypopg for estimation first. Never recommend indexes without impact analysis.
- **Recommending index removal for primary keys/unique constraints:** These are structural requirements, not performance indexes.
- **Ignoring query frequency:** An index that improves one query executed once is less valuable than one that improves a query executed 1000 times daily.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Query normalization/fingerprinting | Custom regex-based parameter removal | pg_query_go Fingerprint() | Handles all PostgreSQL syntax edge cases, including subqueries, CTEs, arrays |
| EXPLAIN JSON parsing | Manual string parsing | encoding/json with proper struct hierarchy | EXPLAIN output is complex nested JSON; manual parsing is error-prone |
| Index creation testing | Actually creating indexes in production | hypopg extension | Safe testing without disk I/O or lock contention |
| Cost calculation | Manual formula | EXPLAIN output Total Cost | PostgreSQL's planner knows more about data distribution |

**Key insight:** Query fingerprinting has many edge cases (string escaping, numeric formats, array literals). pg_query_go uses the actual PostgreSQL parser for correctness.

## Common Pitfalls

### Pitfall 1: Incorrect Fingerprint Hashing
**What goes wrong:** Using SHA256 on raw query text instead of normalized form
**Why it happens:** Forgetting to call `pg_query.Normalize()` before hashing
**How to avoid:** Always fingerprint via `pg_query.Fingerprint()` which handles normalization internally
**Warning signs:** Same query with different parameter values shows as different fingerprints

### Pitfall 2: Missing hypopg Extension
**What goes wrong:** Code fails when hypopg is not installed on monitored database
**Why it happens:** hypopg requires superuser to install; not all databases have it
**How to avoid:** Graceful degradation - check for extension existence, fall back to cost estimation from EXPLAIN without actual index testing
**Warning signs:** Error "function hypopg_create_index does not exist"

### Pitfall 3: EXPLAIN on DDL/DML
**What goes wrong:** Running EXPLAIN on INSERT/UPDATE/DELETE can modify data
**Why it happens:** EXPLAIN ANALYZE executes the statement; ANALYZE must be avoided
**How to avoid:** Use `EXPLAIN (FORMAT JSON)` without ANALYZE for safe analysis
**Warning signs:** Data changes after running analysis

### Pitfall 4: Ignoring Index Maintenance Cost
**What goes wrong:** Recommending indexes that slow down writes more than they speed up reads
**Why it happens:** Only measuring read improvement, not write overhead
**How to avoid:** Include write frequency in benefit calculation (already in CostCalculator.CalculateIndexMaintenanceCost)
**Warning signs:** Users complain about slow INSERT/UPDATE after implementing recommendations

## Code Examples

### Query Fingerprinting Service
```go
// backend/internal/services/query_performance/fingerprint.go
package query_performance

import (
    "crypto/sha256"
    "encoding/hex"
    "github.com/pganalyze/pg_query_go/v5"
    "go.uber.org/zap"
)

type FingerprintService struct {
    logger *zap.Logger
}

func NewFingerprintService(logger *zap.Logger) *FingerprintService {
    return &FingerprintService{logger: logger}
}

// Fingerprint returns a 32-character hex hash for query grouping
func (s *FingerprintService) Fingerprint(queryText string) string {
    fp, err := pg_query.Fingerprint(queryText)
    if err != nil {
        s.logger.Debug("Failed to fingerprint query, using raw hash",
            zap.Error(err))
        fp = queryText
    }
    hash := sha256.Sum256([]byte(fp))
    return hex.EncodeToString(hash[:16])
}

// GroupQueries groups queries by fingerprint
func (s *FingerprintService) GroupQueries(queries []string) map[string][]string {
    groups := make(map[string][]string)
    for _, q := range queries {
        fp := s.Fingerprint(q)
        groups[fp] = append(groups[fp], q)
    }
    return groups
}
```

### Enhanced Query Issue Detection
```go
// backend/internal/services/query_performance/advanced_parser.go
package query_performance

import "encoding/json"

// IssueType constants for consistency
const (
    IssueTypeSeqScan    = "sequential_scan"
    IssueTypeNestedLoop = "nested_loop"
    IssueTypeLargeSort  = "large_sort"
    IssueTypeHashSpill  = "hash_spill"
)

// DetectAntiPatterns finds all performance anti-patterns in a plan
func (qp *QueryParser) DetectAntiPatterns(explainJSON string) ([]QueryIssue, error) {
    var plans []FullExplainPlan
    if err := json.Unmarshal([]byte(explainJSON), &plans); err != nil {
        return nil, err
    }

    var issues []QueryIssue
    for _, plan := range plans {
        qp.walkPlan(plan.Plan, &issues)
    }
    return issues, nil
}
```

### Unused Index Detection Query
```sql
-- Query for detecting unused indexes (IDX-02)
-- Source: PostgreSQL documentation on pg_stat_user_indexes
SELECT
    schemaname,
    tablename,
    indexrelname AS index_name,
    pg_relation_size(indexrelid) AS size_bytes,
    COALESCE(idx_scan, 0) AS scans,
    COALESCE(idx_tup_read, 0) AS tuples_read
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelid NOT IN (
    SELECT conindid FROM pg_constraint WHERE contype IN ('p', 'u', 'f')
  )
ORDER BY pg_relation_size(indexrelid) DESC
LIMIT 20;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Regex-based query normalization | pg_query parser-based fingerprinting | pg_query_go v2+ (2022) | Handles all PostgreSQL syntax correctly |
| Manual index creation for testing | hypopg hypothetical indexes | hypopg 1.0+ (2016) | Safe testing without production impact |
| Single benefit score | Combined cost + frequency + maintenance | Industry standard | More accurate recommendations |

**Deprecated/outdated:**
- Manual string manipulation for fingerprinting: Use pg_query_go instead
- Creating test indexes in production: Always use hypopg

## Open Questions

1. **hypopg availability**
   - What we know: hypopg requires superuser to install extension
   - What's unclear: Whether all monitored databases will have hypopg installed
   - Recommendation: Implement fallback estimation using EXPLAIN cost without hypopg when extension unavailable

2. **EXPLAIN ANALYZE on SELECT with side effects**
   - What we know: Some SELECT queries can have side effects (functions, triggers)
   - What's unclear: How to safely get actual execution metrics
   - Recommendation: Use EXPLAIN (no ANALYZE) by default. Allow ANALYZE only for explicitly whitelisted safe queries.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify/assert (already in use) |
| Config file | None - tests self-contained |
| Quick run command | `go test ./internal/services/query_performance/... -short` |
| Full suite command | `go test ./... -count=1 -race` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QRY-03 | Detect Seq Scan anti-pattern | unit | `go test ./internal/services/query_performance/... -run TestDetectSeqScan -v` | Partial - parser_test.go exists |
| QRY-03 | Detect Nested Loop anti-pattern | unit | `go test ./internal/services/query_performance/... -run TestDetectNestedLoop -v` | Partial - parser_test.go exists |
| QRY-04 | Fingerprint similar queries | unit | `go test ./internal/services/query_performance/... -run TestFingerprint -v` | No - Wave 0 |
| IDX-02 | Find unused indexes | unit | `go test ./internal/services/index_advisor/... -run TestUnusedIndexes -v` | No - Wave 0 |
| IDX-03 | Estimate index impact | unit | `go test ./internal/services/index_advisor/... -run TestIndexImpact -v` | No - Wave 0 |
| IDX-04 | Calculate benefit score | unit | `go test ./internal/services/index_advisor/... -run TestBenefitScore -v` | Partial - cost_calculator.go tested |

### Sampling Rate
- **Per task commit:** `go test ./internal/services/... -short`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `backend/internal/services/query_performance/fingerprint.go` - QRY-04 fingerprinting implementation
- [ ] `backend/internal/services/query_performance/fingerprint_test.go` - fingerprint unit tests
- [ ] `backend/internal/services/index_advisor/unused_detector.go` - IDX-02 detection
- [ ] `backend/internal/services/index_advisor/unused_detector_test.go` - unused index tests
- [ ] `backend/internal/services/index_advisor/hypo_index.go` - IDX-03 impact estimation
- [ ] `backend/internal/services/index_advisor/hypo_index_test.go` - impact estimation tests
- [ ] Migration for hypopg extension check in setup

## Sources

### Primary (HIGH confidence)
- Codebase analysis (backend/internal/services/query_performance/*.go) - existing implementation patterns
- Codebase analysis (backend/internal/services/index_advisor/*.go) - existing stubs and models
- Database schema (backend/migrations/024_create_query_performance_schema.sql, 026_create_index_advisor_schema.sql) - table structures
- Go module list (go list -m -versions) - version verification for pg_query_go

### Secondary (MEDIUM confidence)
- pg_query_go GitHub repository - library capabilities (fingerprint, parse, normalize)
- PostgreSQL documentation - pg_stat_user_indexes, EXPLAIN format

### Tertiary (LOW confidence)
- hypopg documentation - extension usage patterns (not verified in current environment)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - pg_query_go is the industry standard for PostgreSQL query parsing
- Architecture: HIGH - builds on existing patterns in codebase
- Pitfalls: MEDIUM - hypopg availability varies by environment

**Research date:** 2026-05-12
**Valid until:** 2026-06-12 (1 month - stable PostgreSQL ecosystem)