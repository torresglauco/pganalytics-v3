# Phase 06: Query Optimization Foundation - Research

**Gathered:** 2026-05-11
**Status:** Complete (derived from milestone research)

---

## Phase Boundary

This phase establishes the foundation for all performance optimizations:
- Migrate from lib/pq to pgx v5 for 2-3x query performance
- Implement connection pooling with read/write separation
- Add slow query identification and timeline visualization
- Enable performance monitoring (pprof, Prometheus metrics)

**Depends on:** Phase 05 (CI/CD Infrastructure complete)

**Enables:** Phase 07 (Caching), Phase 08 (Dashboard Optimization), Phase 09 (Index Intelligence)

---

## Technical Approach

### 1. pgx v5 Migration (API-02, API-03, API-04)

**Current State:**
- Using lib/pq (maintenance mode, no active development)
- Single connection pool with MaxOpenConns=100, MaxIdleConns=20
- TimescaleDB pool: MaxOpenConns=25, MaxIdleConns=5

**Target State:**
- pgx v5.9.2 with pgxpool for native connection pooling
- Read-only pool for dashboard queries (separate from write pool)
- Pool metrics exposed via Prometheus

**Migration Steps:**
1. Add pgx v5 to go.mod alongside lib/pq
2. Create pgxpool wrapper in `internal/storage/`
3. Update database connection in `postgres.go`
4. Add read-only pool configuration
5. Expose pool metrics via `db.Stat()`

**Key Files:**
- `backend/internal/storage/postgres.go` - main connection
- `backend/internal/storage/timescaledb.go` - TimescaleDB connection
- `backend/go.mod` - dependency update

### 2. Slow Query Identification (QRY-01, QRY-02, QRY-05)

**Approach:**
- Query `pg_stat_statements` for slow queries by mean_time
- Store historical data in existing `query_plans` table
- Build timeline from `query_performance_timeline` table

**Existing Infrastructure:**
- Migration 024: `query_plans`, `query_issues`, `query_performance_timeline` tables
- Handler skeleton: `handlers_query_performance.go`
- Service skeleton: `services/query_performance.go`

**Implementation:**
1. Enable `pg_stat_statements` extension (check if enabled)
2. Create slow query API endpoint
3. Add timeline query with time_bucket aggregation
4. Build frontend component for visualization

**Key Files:**
- `backend/internal/handlers/handlers_query_performance.go`
- `backend/internal/services/query_performance.go`
- `frontend/src/pages/QueryPerformancePage.tsx`

### 3. Index Usage Statistics (IDX-01)

**Approach:**
- Query `pg_stat_user_indexes` for usage statistics
- Identify unused indexes (idx_scan = 0)
- Expose via API endpoint

**Implementation:**
1. Add index stats query to index advisor service
2. Create API endpoint for index statistics
3. Add frontend component for display

**Key Files:**
- `backend/internal/handlers/handlers_index_advisor.go`
- `backend/internal/services/index_advisor.go`

### 4. Performance Monitoring (MON-01, MON-02, MON-03)

**pprof Integration (MON-01):**
- Add `import _ "net/http/pprof"` to main server
- Profile endpoints available at `/debug/pprof/`
- Zero-effort addition since HTTP server exists

**Prometheus Metrics (MON-02, MON-03):**
- Add histogram for query_duration_seconds
- Add gauge for pool stats (open, idle, in-use)
- Extend existing Prometheus client usage

**Implementation:**
1. Add pprof import to `cmd/server/main.go`
2. Extend `internal/metrics/` with new histograms
3. Add pool stats collection in storage layer

**Key Files:**
- `backend/cmd/server/main.go`
- `backend/internal/metrics/metrics.go`
- `backend/internal/storage/postgres.go`

---

## Validation Architecture

### Dimension 1: Query Performance

**Test Strategy:**
- Benchmark before/after with pgx migration
- Compare query execution times with production data volumes
- Measure connection pool efficiency under load

**Acceptance Criteria:**
- Query latency P50 < 50ms, P95 < 200ms
- Connection pool utilization < 80% under normal load
- No connection starvation with 500+ collectors

### Dimension 2: API Response Times

**Test Strategy:**
- Load test slow query endpoint
- Measure dashboard API response times
- Verify pool metrics accuracy

**Acceptance Criteria:**
- Slow query API < 500ms for top 100 queries
- Pool metrics accurate to within 1 second
- No blocking on pool acquisition

### Dimension 3: Monitoring Accuracy

**Test Strategy:**
- Verify pprof endpoints return valid profiles
- Check Prometheus metrics match db.Stat()
- Compare histogram percentiles with actual query times

**Acceptance Criteria:**
- pprof CPU profile captures > 95% of samples
- Prometheus gauges match database pool state
- Histogram buckets cover P50-P99 range

---

## Integration Points

### Backend
- `internal/storage/postgres.go` - pgx migration
- `internal/handlers/handlers_query_performance.go` - slow queries API
- `internal/handlers/handlers_index_advisor.go` - index stats API
- `internal/metrics/metrics.go` - Prometheus metrics
- `cmd/server/main.go` - pprof import

### Frontend
- `src/pages/QueryPerformancePage.tsx` - slow queries UI
- `src/components/QueryTimeline.tsx` - timeline visualization
- `src/api/queryPerformanceApi.ts` - API client

### Database
- `pg_stat_statements` extension
- `pg_stat_user_indexes` view
- Existing migrations 024-027

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| pgx migration breaks existing queries | High | Run full test suite, use staging with production data |
| Connection pool misconfiguration | High | Load test with 500+ collectors, monitor db.Stats() |
| pg_stat_statements not enabled | Medium | Check extension, provide setup instructions |
| Performance regression in specific queries | Medium | Benchmark critical paths, EXPLAIN ANALYZE before/after |

---

## Dependencies

**New Dependencies:**
- `github.com/jackc/pgx/v5` v5.9.2 - PostgreSQL driver
- `github.com/jackc/pgx/v5/pgxpool` - Connection pooling

**Existing Dependencies (extend):**
- `github.com/prometheus/client_golang` - Metrics
- `net/http/pprof` - Profiling (stdlib)

---

*Research completed: 2026-05-11*
*Source: Milestone research (.planning/research/)*