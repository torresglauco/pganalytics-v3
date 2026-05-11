---
phase: 06-query-optimization-foundation
plan: 02
subsystem: api, database
tags: [pg_stat_statements, slow-queries, timeline, index-stats, query-performance]

# Dependency graph
requires:
  - phase: 06-01
    provides: pgx v5 connection pooling with dedicated read-only pool
provides:
  - Query performance store with pg_stat_statements queries
  - Query performance service layer with pagination and statistics
  - API endpoints for slow queries, timeline, and index stats
  - Migration for pg_stat_statements setup helper
affects: [phase-07, phase-08, phase-09]

# Tech tracking
tech-stack:
  added: [github.com/DATA-DOG/go-sqlmock v1.5.2 for testing]
  patterns: [Store pattern for database access, Service layer for business logic, TDD with mock stores]

key-files:
  created:
    - backend/internal/storage/query_performance_store.go
    - backend/internal/storage/query_performance_store_test.go
    - backend/internal/services/query_performance/service.go
    - backend/internal/services/query_performance/service_test.go
    - backend/internal/api/handlers_query_performance_new_test.go
    - backend/migrations/028_pg_stat_statements_setup.sql
  modified:
    - backend/internal/api/handlers_query_performance.go
    - backend/internal/api/server.go

key-decisions:
  - "Create separate QueryPerformanceStore instead of extending PostgresDB for cleaner separation of concerns"
  - "Use Store interface in service to enable mocking in tests"
  - "Name database-specific handlers with 'Database' prefix to avoid collision with collector-specific handlers"
  - "Gracefully handle missing pg_stat_statements extension by returning empty results, not errors"

patterns-established:
  - "Store pattern: Database access encapsulated in dedicated store structs with context and error handling"
  - "Service layer pattern: Business logic in service structs that call stores, with response DTOs"
  - "Index categorization: unused (0 scans), low (<100), normal (<10000), high (>=10000)"

requirements-completed: [QRY-01, QRY-02, QRY-05, IDX-01]

# Metrics
duration: 73min
completed: 2026-05-11
---

# Phase 06 Plan 02: Slow Query Identification and Timeline Summary

**Slow query identification, timeline visualization, and index usage statistics using pg_stat_statements with graceful handling of missing extension**

## Performance

- **Duration:** 73 min
- **Started:** 2026-05-11T17:58:41Z
- **Completed:** 2026-05-11T19:11:44Z
- **Tasks:** 4
- **Files modified:** 8

## Accomplishments

- Query performance store with GetSlowQueries, GetQueryTimeline, and GetIndexStats methods
- Service layer with pagination metadata, timeline statistics, and index usage categorization
- API endpoints for slow queries, timeline, and index stats integrated with existing handlers
- Migration for pg_stat_statements setup helper with graceful missing extension handling

## Task Commits

Each task was committed atomically:

1. **Task 1: Create query performance store with pg_stat_statements queries** - `e996e00` (test)
2. **Task 2: Implement query performance service layer** - `e42822b` (feat)
3. **Task 3: Implement API handlers for slow queries, timeline, and index stats** - `7fa6f5e` (feat)
4. **Task 4: Create migration for pg_stat_statements setup helper** - `ded34fa` (feat)

## Files Created/Modified

- `backend/internal/storage/query_performance_store.go` - Query performance database access with pg_stat_statements
- `backend/internal/storage/query_performance_store_test.go` - Store tests with sqlmock
- `backend/internal/services/query_performance/service.go` - Service layer with pagination and statistics
- `backend/internal/services/query_performance/service_test.go` - Service tests with mock store
- `backend/internal/api/handlers_query_performance.go` - Updated with new handlers for database endpoints
- `backend/internal/api/handlers_query_performance_new_test.go` - Handler tests
- `backend/internal/api/server.go` - Route registration for new endpoints
- `backend/migrations/028_pg_stat_statements_setup.sql` - Migration for pg_stat_statements helper

## Decisions Made

- Created separate QueryPerformanceStore rather than adding methods to PostgresDB for better separation of concerns
- Used Store interface in service to enable dependency injection and easier testing
- Named database-specific handlers with "Database" prefix (handleGetDatabaseSlowQueries) to avoid collision with existing collector-specific handlers
- Graceful handling of missing pg_stat_statements extension - returns empty results instead of errors

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Handler name collision with existing handlers**
- **Found during:** Task 3 (API handlers implementation)
- **Issue:** handleGetSlowQueries and handleGetQueryTimeline already existed in handlers.go for collector-specific endpoints
- **Fix:** Renamed new handlers to handleGetDatabaseSlowQueries, handleGetDatabaseQueryTimeline, handleGetDatabaseIndexStats to indicate database-specific endpoints
- **Files modified:** backend/internal/api/handlers_query_performance.go, backend/internal/api/server.go, backend/internal/api/handlers_query_performance_new_test.go
- **Verification:** All tests pass, no build errors
- **Committed in:** 7fa6f5e (Task 3 commit)

**2. [Rule 3 - Blocking] Go formatting issues in pre-commit hooks**
- **Found during:** Multiple task commits
- **Issue:** gofmt not formatting files correctly before commit
- **Fix:** Ran gofmt -w on all modified files before final commits
- **Files modified:** Multiple Go files (query_performance_store.go, service.go, service_test.go, handlers)
- **Verification:** golangci-lint passes
- **Committed in:** Each task commit included formatting fixes

---

**Total deviations:** 2 auto-fixed (1 blocking handler collision, 1 formatting)
**Impact on plan:** Both auto-fixes were necessary for successful execution. Handler rename was essential for compilation.

## Issues Encountered

- sqlmock dependency not initially present - installed during Task 1
- Pre-commit golangci-lint timeout exceeded on first attempt - resolved by running gofmt manually

## User Setup Required

**pg_stat_statements extension must be enabled for full functionality:**

1. Add to postgresql.conf: `shared_preload_libraries = 'pg_stat_statements'`
2. Restart PostgreSQL
3. Run: `CREATE EXTENSION pg_stat_statements;` in target database

The system handles missing extension gracefully - returns empty results with no errors.

## Next Phase Readiness

- Query performance infrastructure complete
- Ready for Phase 07 (Caching Infrastructure) to optimize repeated queries
- Ready for Phase 09 (Index Intelligence) which uses the index stats from this plan

---
*Phase: 06-query-optimization-foundation*
*Completed: 2026-05-11*