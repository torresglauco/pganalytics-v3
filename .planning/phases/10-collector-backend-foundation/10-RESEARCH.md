# Phase 10: Collector & Backend Foundation - Research

**Researched:** 2026-05-13
**Domain:** PostgreSQL Monitoring (replication, host metrics, inventory), C++ Collector, Go Backend
**Confidence:** HIGH

## Summary

Phase 10 extends the existing pgAnalytics collector-backend architecture to support replication monitoring, host status/inventory, and database inventory tracking. The collector (C++) already has foundational plugins for replication (`replication_plugin.cpp`) and system stats (`sysstat_plugin.cpp`). The backend (Go with pgx v5) has established patterns for metrics storage, API handlers, and database operations that must be followed.

**Primary recommendation:** Extend existing collector plugins and backend storage/handlers following established patterns. Do NOT rewrite - only ADD new capabilities. Leverage existing version detection in collector for PG 11-17 compatibility.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| pgx v5 | 5.9.2 | PostgreSQL driver with native connection pooling | Already in use, provides 2-3x performance over lib/pq |
| Gin | 1.10.0 | HTTP web framework | Standard across all backend handlers |
| libpq | PG 11-17 | PostgreSQL client library for C++ collector | Standard for native PG connectivity |
| nlohmann/json | latest | JSON serialization for C++ | Already used in all collector plugins |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| TimescaleDB | 2.15.0-pg16 | Time-series storage for metrics | For historical replication/host metrics |
| uuid | google/uuid v1.6.0 | UUID generation | Collector IDs, entity identification |
| zap | 1.27.0 | Structured logging | Backend logging |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| libpq in C++ | nanodbc | nanodbc adds ODBC abstraction layer - unnecessary for PostgreSQL-only |
| Custom version detection | pg_catalog queries | Current approach using `current_setting('server_version_num')` works reliably |

**Installation:**
```bash
# Go dependencies already in go.mod
go mod download

# C++ collector dependencies managed via CMake and vcpkg
# See collector/CMakeLists.txt
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
├── include/           # Header files for all plugins
│   ├── replication_plugin.h
│   ├── sysstat_plugin.h
│   ├── host_inventory_plugin.h  # NEW for Phase 10
│   └── ...
├── src/               # Implementation files
│   ├── replication_plugin.cpp
│   ├── sysstat_plugin.cpp
│   └── ...
└── CMakeLists.txt

backend/
├── internal/
│   ├── api/
│   │   ├── handlers_replication.go    # NEW for Phase 10
│   │   ├── handlers_host.go           # NEW for Phase 10
│   │   └── handlers_inventory.go      # NEW for Phase 10
│   ├── storage/
│   │   ├── replication_store.go       # NEW for Phase 10
│   │   ├── host_store.go              # NEW for Phase 10
│   │   └── inventory_store.go         # NEW for Phase 10
│   └── timescale/
│       └── aggregates_replication.go  # NEW for Phase 10
├── pkg/models/
│   ├── replication_models.go          # NEW for Phase 10
│   ├── host_models.go                 # NEW for Phase 10
│   └── inventory_models.go            # NEW for Phase 10
└── migrations/
    └── 031_replication_tables.sql     # NEW for Phase 10
```

### Pattern 1: Collector Plugin Implementation
**What:** C++ collector plugin that inherits from `Collector` base class
**When to use:** All new metric collection types
**Example:**
```cpp
// Source: collector/include/replication_plugin.h (existing pattern)
class PgReplicationCollector : public Collector {
public:
    explicit PgReplicationCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_replication"; }
    bool isEnabled() const override { return enabled_; }

private:
    int detectPostgresVersion();  // Key for version-adaptive queries
    // ... collection methods
};
```

### Pattern 2: Backend Handler Implementation
**What:** Go HTTP handler following Gin framework conventions
**When to use:** All new API endpoints
**Example:**
```go
// Source: backend/internal/api/handlers_metrics.go (existing pattern)
// @Summary Get Replication Metrics
// @Description Get replication status for a collector
// @Tags Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Success 200 {object} models.MetricsResponse
// @Router /api/v1/collectors/{collector_id}/replication [get]
func (s *Server) handleGetReplicationMetrics(c *gin.Context) {
    collectorIDStr := c.Param("collector_id")
    collectorID, err := uuid.Parse(collectorIDStr)
    if err != nil {
        errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
        c.JSON(errResp.StatusCode, errResp)
        return
    }

    ctx := c.Request.Context()
    metrics, err := s.postgres.GetReplicationMetrics(ctx, collectorID)
    if err != nil {
        c.JSON(err.(*apperrors.AppError).StatusCode, err)
        return
    }

    resp := &models.MetricsResponse{
        MetricType: "pg_replication",
        Count:      len(metrics.ReplicationStatus),
        Timestamp:  time.Now(),
        Data:       metrics,
    }

    c.JSON(http.StatusOK, resp)
}
```

### Pattern 3: Backend Storage Implementation
**What:** Go database operations using pgx v5 with prepared statements
**When to use:** All database read/write operations
**Example:**
```go
// Source: backend/internal/storage/metrics_store.go (existing pattern)
func (p *PostgresDB) StoreReplicationMetrics(ctx context.Context, status []*models.ReplicationStatus, slots []*models.ReplicationSlot) error {
    if len(status) == 0 && len(slots) == 0 {
        return nil
    }

    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return apperrors.DatabaseError("begin transaction", err.Error())
    }
    defer func() { _ = tx.Rollback() }()

    // Insert replication status
    if len(status) > 0 {
        stmt, err := tx.PrepareContext(ctx, `
            INSERT INTO metrics_replication_status (time, collector_id, application_name, state, sync_state,
                write_lsn, flush_lsn, replay_lsn, write_lag_ms, flush_lag_ms, replay_lag_ms)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
            ON CONFLICT DO NOTHING
        `)
        // ... execute for each item
    }
    return tx.Commit()
}
```

### Anti-Patterns to Avoid
- **Don't create new Collector base class:** Extend existing `Collector` class from `collector.h`
- **Don't use lib/pq in new code:** Use pgx v5 exclusively for new backend code
- **Don't skip version detection:** All PG queries must check version and adapt (see `detectPostgresVersion()` in replication_plugin.cpp)
- **Don't ignore collector auth:** All collector-to-backend communication must use mTLS or token auth

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| PostgreSQL connection management | Custom connection pooling | pgxpool (pgx v5) | Built-in health checks, connection limits, idle timeout |
| Version detection | Custom version parsing | `SELECT current_setting('server_version_num')::int` | Reliable integer comparison, matches PG internal version |
| LSN parsing | String manipulation | `parseLsn()` in replication_plugin.cpp | Handles hex format X/XXXXXXXX correctly |
| System stats on Linux | Manual /proc parsing | Existing `SysstatCollector` class | Already handles /proc/stat, /proc/meminfo, /proc/diskstats |

**Key insight:** The collector already has robust implementations for replication stats and system stats. The primary work is backend storage and API handlers, plus extending collector with new inventory capabilities.

## Common Pitfalls

### Pitfall 1: Version-Specific Query Incompatibility
**What goes wrong:** Query works on PG 15 but fails on PG 11 due to missing columns/views
**Why it happens:** PostgreSQL adds new system columns/views in each major version
**How to avoid:** Always check `postgres_version_major_` and branch query logic
**Warning signs:** "column X does not exist" errors on older PG versions

**Example from existing code:**
```cpp
// collector/src/replication_plugin.cpp lines 256-299
if (postgres_version_major_ >= 13) {
    query = R"(
        SELECT ... write_lag, flush_lag, replay_lag ...  -- PG13+ columns
        FROM pg_stat_replication
    )";
} else {
    query = R"(
        SELECT ... 0 as write_lag_ms, 0 as flush_lag_ms, 0 as replay_lag_ms
        FROM pg_stat_replication
    )";
}
```

### Pitfall 2: Missing Collector Authentication
**What goes wrong:** Collector endpoints accept unauthenticated requests
**Why it happens:** Forgetting to add `s.CollectorAuthMiddleware()` or `s.MTLSMiddleware()`
**How to avoid:** All collector-initiated endpoints must use appropriate auth middleware
**Warning signs:** Security scan shows unauthenticated metrics push endpoint

### Pitfall 3: TimescaleDB Missing for Aggregates
**What goes wrong:** Dashboard queries fail with "relation does not exist"
**Why it happens:** TimescaleDB extension not installed or continuous aggregates not created
**How to avoid:** Check `s.timescale == nil` and return appropriate 503 error
**Warning signs:** `handleGetDashboardDatabaseStats` returns 503 Service Unavailable

### Pitfall 4: Cascading Replication Detection
**What goes wrong:** Only direct primary-standby pairs detected, missing cascading standbys
**Why it happens:** Only querying `pg_stat_replication` on primary
**How to avoid:** Track `recovery.conf`/`standby.signal` and cross-reference with upstream
**Warning signs:** Missing intermediate nodes in replication topology

## Code Examples

### Version-Adaptive PostgreSQL Query (C++)
```cpp
// Source: collector/src/replication_plugin.cpp pattern
// Pattern for adapting queries based on PostgreSQL version

std::string buildReplicationQuery() {
    std::string query;

    if (postgres_version_major_ >= 13) {
        // PG 13+ has write_lag, flush_lag, replay_lag columns
        query = R"(
            SELECT pid, usename, application_name, state, sync_state,
                   write_lsn, flush_lsn, replay_lsn,
                   EXTRACT(EPOCH FROM write_lag) * 1000 as write_lag_ms,
                   EXTRACT(EPOCH FROM flush_lag) * 1000 as flush_lag_ms,
                   EXTRACT(EPOCH FROM replay_lag) * 1000 as replay_lag_ms
            FROM pg_stat_replication
        )";
    } else if (postgres_version_major_ >= 10) {
        // PG 10-12 has lsn columns but not lag interval columns
        query = R"(
            SELECT pid, usename, application_name, state, sync_state,
                   sent_lsn, flush_lsn, replay_lsn,
                   0 as write_lag_ms, 0 as flush_lag_ms, 0 as replay_lag_ms
            FROM pg_stat_replication
        )";
    } else {
        // PG 9.x uses different column names
        query = R"(
            SELECT procpid as pid, usesysid, application_name, state, sync_state,
                   location as sent_lsn, flush_lsn, replay_lsn,
                   0 as write_lag_ms, 0 as flush_lag_ms, 0 as replay_lag_ms
            FROM pg_stat_replication
        )";
    }
    return query;
}
```

### System Statistics Collection (C++)
```cpp
// Source: collector/src/sysstat_plugin.cpp
// Pattern for reading Linux /proc filesystem for system metrics

json SysstatCollector::collectCpuStats() {
    json result = json::object();

    // Parse /proc/stat for CPU percentages
    std::ifstream stat_file("/proc/stat");
    if (stat_file.is_open()) {
        std::string line;
        if (std::getline(stat_file, line)) {
            std::istringstream iss(line);
            std::string cpu_label;
            unsigned long user, nice, system, idle, iowait, irq, softirq, steal;
            if (iss >> cpu_label >> user >> nice >> system >> idle >> iowait >> irq >> softirq >> steal) {
                unsigned long total = user + nice + system + idle + iowait + irq + softirq + steal;
                if (total > 0) {
                    result["user"] = (100.0 * user) / total;
                    result["system"] = (100.0 * system) / total;
                    result["idle"] = (100.0 * idle) / total;
                    result["iowait"] = (100.0 * iowait) / total;
                }
            }
        }
        stat_file.close();
    }
    return result;
}
```

### Backend Route Registration (Go)
```go
// Source: backend/internal/api/server.go lines 345-371 pattern
// Pattern for registering collector-specific metric routes

collectors := api.Group("/collectors")
{
    // Registration (no auth required)
    collectors.POST("/register", s.handleCollectorRegister)

    // Token refresh (collector auth required)
    collectors.POST("/refresh-token", s.CollectorAuthMiddleware(), s.handleRefreshCollectorToken)

    // Protected routes (user auth)
    collectors.GET("", s.AuthMiddleware(), s.handleListCollectors)
    collectors.GET("/:id", s.AuthMiddleware(), s.handleGetCollector)
    collectors.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteCollector)

    // NEW Phase 10 routes - following existing pattern
    collectors.GET("/:id/replication", s.AuthMiddleware(), s.handleGetReplicationMetrics)
    collectors.GET("/:id/host-status", s.AuthMiddleware(), s.handleGetHostStatus)
    collectors.GET("/:id/inventory/tables", s.AuthMiddleware(), s.handleGetTableInventory)
    collectors.GET("/:id/inventory/indexes", s.AuthMiddleware(), s.handleGetIndexInventory)
    collectors.GET("/:id/inventory/columns", s.AuthMiddleware(), s.handleGetColumnInventory)
}
```

### Database Migration Pattern
```sql
-- Pattern from backend/migrations/000_complete_schema.sql
-- TimescaleDB hypertable creation for time-series metrics

-- Replication status metrics table
CREATE TABLE metrics_replication_status (
    time TIMESTAMPTZ NOT NULL,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    application_name VARCHAR(255),
    state VARCHAR(50),
    sync_state VARCHAR(10),
    write_lsn VARCHAR(50),
    flush_lsn VARCHAR(50),
    replay_lsn VARCHAR(50),
    write_lag_ms BIGINT,
    flush_lag_ms BIGINT,
    replay_lag_ms BIGINT,
    behind_by_mb BIGINT
);

-- Create TimescaleDB hypertable for time-series queries
SELECT create_hypertable('metrics_replication_status', 'time', if_not_exists => TRUE);

-- Indexes for common query patterns
CREATE INDEX idx_replication_status_collector ON metrics_replication_status(collector_id, time DESC);
CREATE INDEX idx_replication_status_state ON metrics_replication_status(state) WHERE state != 'streaming';
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| lib/pq driver | pgx v5 with pgxpool | Phase 06 (v1.2) | 2-3x query performance, native pooling |
| Per-query connections | Connection pool reuse | Phase 06 (v1.2) | Reduced connection overhead |
| Full table scans | TimescaleDB continuous aggregates | Phase 08 (v1.2) | Instant dashboard loads |
| Manual version checks | `detectPostgresVersion()` caching | Existing in collector | Avoids repeated version queries |

**Deprecated/outdated:**
- lib/pq: Replaced by pgx v5, do NOT use in new code
- Direct SQL string concatenation: Use prepared statements to prevent SQL injection
- Procedural metrics collection: Use existing plugin architecture

## Open Questions

1. **Logical Replication Subscriptions (REP-02)**
   - What we know: `pg_stat_subscription` view available in PG 10+ for monitoring logical replication
   - What's unclear: Need to determine if collector should also collect publication definitions from `pg_publication`
   - Recommendation: Start with subscription status, add publications in follow-up task

2. **Cascading Replication Topology (REP-03)**
   - What we know: `pg_stat_wal_receiver` on standbys shows upstream connection
   - What's unclear: Best approach for building complete topology graph (single query vs multiple queries with client-side assembly)
   - Recommendation: Collect standby info via `pg_stat_wal_receiver`, build topology in backend from multiple collector reports

3. **Host Up/Down Detection (HOST-01)**
   - What we know: Can use `last_seen` timestamp from collectors table
   - What's unclear: Threshold for "down" status (5 minutes? configurable?)
   - Recommendation: Use configurable threshold (default 5 minutes), expose via API with `is_healthy` boolean

4. **Schema Change Tracking (INV-05)**
   - What we know: Current `metrics_pg_schema_*` tables store snapshots
   - What's unclear: How to detect and store changes efficiently (diff vs version history)
   - Recommendation: Add `schema_version` column, track changes via trigger or periodic comparison

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
| REP-01 | View streaming replication status with lag metrics | integration | `go test ./tests/integration -run TestReplication` | No - Wave 0 |
| REP-02 | View logical replication subscriptions | integration | `go test ./tests/integration -run TestLogicalReplication` | No - Wave 0 |
| REP-03 | View cascading replication topology | integration | `go test ./tests/integration -run TestReplicationTopology` | No - Wave 0 |
| REP-04 | View replication slots with WAL retention | integration | `go test ./tests/integration -run TestReplicationSlots` | No - Wave 0 |
| HOST-01 | View host up/down status | unit | `go test ./internal/storage -run TestHostStatus` | No - Wave 0 |
| HOST-02 | View OS metrics (CPU, memory, disk) | unit | `go test ./internal/storage -run TestHostMetrics` | No - Wave 0 |
| HOST-03 | View host inventory | unit | `go test ./internal/storage -run TestHostInventory` | No - Wave 0 |
| INV-01 | View table inventory with sizes | integration | `go test ./tests/integration -run TestTableInventory` | No - Wave 0 |
| INV-02 | View column inventory | integration | `go test ./tests/integration -run TestColumnInventory` | No - Wave 0 |
| INV-03 | View index inventory with usage | integration | `go test ./tests/integration -run TestIndexInventory` | No - Wave 0 |
| INV-04 | View extension inventory | integration | `go test ./tests/integration -run TestExtensionInventory` | No - Wave 0 |
| INV-05 | Track schema changes | integration | `go test ./tests/integration -run TestSchemaChanges` | No - Wave 0 |
| VER-01 | Support PG 13-17 | unit | `go test ./... -run TestPostgreSQLVersion` | Partial - existing version detection |
| VER-02 | Support PG 11-12 | unit | `go test ./... -run TestPostgreSQLVersion` | Partial - existing version detection |
| VER-04 | Adapt queries based on version | unit | C++ tests in collector/tests | Partial - existing tests |
| COLL-01 | Collector decentralized mode | integration | Requires live PG cluster | No - Wave 0 |
| COLL-02 | Collector centralized mode | integration | Requires live PG cluster | No - Wave 0 |
| COLL-03 | Mixed deployment | integration | Requires live PG clusters | No - Wave 0 |
| COLL-04 | Low resource footprint | benchmark | `go test -bench=. ./...` | No - Wave 0 |
| COLL-05 | Secure communication | security | `go test ./tests/security -run TestTLS` | Partial - existing TLS tests |

### Sampling Rate
- **Per task commit:** `go test ./internal/storage ./internal/api -short -v`
- **Per wave merge:** `go test ./... -cover -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `backend/internal/storage/replication_store.go` - Store/Get methods for REP-*
- [ ] `backend/internal/storage/host_store.go` - Store/Get methods for HOST-*
- [ ] `backend/internal/storage/inventory_store.go` - Store/Get methods for INV-*
- [ ] `backend/internal/api/handlers_replication.go` - HTTP handlers for replication endpoints
- [ ] `backend/internal/api/handlers_host.go` - HTTP handlers for host endpoints
- [ ] `backend/internal/api/handlers_inventory.go` - HTTP handlers for inventory endpoints
- [ ] `backend/pkg/models/replication_models.go` - Data structures for replication metrics
- [ ] `backend/pkg/models/host_models.go` - Data structures for host metrics
- [ ] `backend/pkg/models/inventory_models.go` - Data structures for inventory metrics
- [ ] `backend/migrations/031_replication_tables.sql` - Database schema for new metrics
- [ ] `backend/tests/integration/replication_test.go` - Integration tests for replication
- [ ] `backend/tests/integration/host_test.go` - Integration tests for host monitoring
- [ ] `backend/tests/integration/inventory_test.go` - Integration tests for inventory
- [ ] `collector/include/host_inventory_plugin.h` - C++ header for host inventory collection
- [ ] `collector/src/host_inventory_plugin.cpp` - C++ implementation for host inventory
- [ ] `collector/tests/test_replication_plugin.cpp` - C++ unit tests for replication plugin

*(If no gaps: "None - existing test infrastructure covers all phase requirements")*

## Sources

### Primary (HIGH confidence)
- Existing codebase analysis: `collector/src/replication_plugin.cpp`, `collector/src/sysstat_plugin.cpp`, `collector/src/schema_plugin.cpp`
- Backend patterns: `backend/internal/api/handlers_metrics.go`, `backend/internal/storage/metrics_store.go`
- Database schema: `backend/migrations/000_complete_schema.sql`
- Go modules: `go.mod` showing pgx v5.9.2, Gin 1.10.0

### Secondary (MEDIUM confidence)
- PostgreSQL 17 documentation patterns for `pg_stat_replication`, `pg_stat_subscription`, `pg_replication_slots`
- pgx v5 best practices from existing Phase 06 implementation

### Tertiary (LOW confidence)
- Web search for PostgreSQL version differences - verified against existing codebase version detection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Already implemented and working in the codebase
- Architecture: HIGH - Clear patterns from existing plugins and handlers
- Pitfalls: HIGH - Based on actual codebase analysis and existing error handling

**Research date:** 2026-05-13
**Valid until:** 30 days - PostgreSQL versions 11-17 stable, pgx v5 stable

---

## Phase Requirements Mapping

| ID | Description | Research Support |
|----|-------------|-----------------|
| REP-01 | View streaming replication status with lag metrics | Existing `replication_plugin.cpp` collects `pg_stat_replication` with version-adaptive queries for write_lag/flush_lag/replay_lag (PG13+) |
| REP-02 | View logical replication subscriptions | Need new collection of `pg_stat_subscription` (PG10+) and `pg_publication` |
| REP-03 | View cascading replication topology | Need to combine `pg_stat_replication` (primary side) with `pg_stat_wal_receiver` (standby side) |
| REP-04 | View replication slots with WAL retention | Existing `collectReplicationSlots()` in replication_plugin.cpp queries `pg_replication_slots` |
| HOST-01 | View host up/down status | Use `collectors.last_seen` timestamp with configurable threshold (default 5 min) |
| HOST-02 | View OS metrics | Existing `SysstatCollector` collects CPU, memory, disk I/O from /proc filesystem |
| HOST-03 | View host inventory | Need new host inventory plugin to collect OS version, hardware, PG config from `pg_settings` |
| INV-01 | View table inventory with sizes | Existing `schema_plugin.cpp` collects tables; add size info from `pg_total_relation_size()` |
| INV-02 | View column inventory | Existing `collectColumnInfo()` in schema_plugin.cpp |
| INV-03 | View index inventory with usage | Existing `collectIndexInfo()` needs enhancement with `pg_stat_user_indexes` usage stats |
| INV-04 | View extension inventory | Existing `extension_plugin.cpp` collects `pg_extension` |
| INV-05 | Track schema changes | Need to add version tracking to existing `metrics_pg_schema_*` tables |
| VER-01 | Support PG 13-17 | Existing `detectPostgresVersion()` handles all versions; queries adapt via version branching |
| VER-02 | Support PG 11-12 | Same version detection; queries use fallback columns for older versions |
| VER-04 | Adapt queries based on version | Pattern established in replication_plugin.cpp lines 256-299 |
| COLL-01 | Decentralized mode | Default collector mode - runs on same host as PostgreSQL |
| COLL-02 | Centralized mode | Collector can connect remotely; use `postgresHost` config for remote PG |
| COLL-03 | Mixed deployment | Both modes supported via configuration; no code changes needed |
| COLL-04 | Low resource footprint | Existing collector uses minimal memory; single thread per plugin type |
| COLL-05 | Secure communication | Existing mTLS support in auth.cpp; collector tokens with expiration |