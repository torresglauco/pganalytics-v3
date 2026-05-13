---
phase: 09-index-intelligence
plan: 02
subsystem: database
tags: [index-advisor, hypopg, unused-indexes, impact-estimation]

requires:
  - phase: 09-01
    provides: Fingerprinter service for query grouping, EXPLAIN analysis patterns
provides:
  - Unused index detection from pg_stat_user_indexes
  - Hypothetical index impact estimation using hypopg
  - Index recommendation engine with benefit scoring
  - API endpoints for unused indexes and impact estimation
affects: [index-intelligence, query-optimization, performance-advisor]

tech-stack:
  added: []
  patterns:
    - "Hypopg extension for safe hypothetical index testing"
    - "Graceful fallback when extensions unavailable"
    - "Benefit scoring combining cost improvement with query frequency"

key-files:
  created:
    - backend/internal/services/index_advisor/unused_detector.go
    - backend/internal/services/index_advisor/hypo_index.go
    - backend/internal/storage/index_recommendation_store.go
    - backend/migrations/030_add_hypopg_check.sql
  modified:
    - backend/internal/services/index_advisor/analyzer.go
    - backend/internal/api/handlers_index_advisor.go
    - backend/internal/api/server.go

key-decisions:
  - "Use hypopg extension for hypothetical index testing with graceful fallback"
  - "Query pg_stat_user_indexes with LEFT JOIN pg_constraint to exclude PK/unique/FK indexes"
  - "Calculate improvement as (costWithout - costWith) / costWithout * 100"
  - "Always drop hypothetical index after testing (defer cleanup)"
  - "Connect to monitored databases dynamically from connection_string in postgresql_instances table"

requirements-completed: [IDX-02, IDX-03, IDX-04]

duration: 30min
completed: 2026-05-13
---

# Phase 09 Plan 02: Index Recommendation Engine Summary

**Unused index detection and impact estimation using hypopg with benefit scoring**

## Performance

- **Duration:** 30 min
- **Started:** 2026-05-13T22:49:28Z
- **Completed:** 2026-05-13T23:19:56Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- UnusedIndexDetector finds indexes with zero scans, excluding constraint indexes
- HypoIndexTester estimates index impact using hypopg with automatic cleanup
- IndexAnalyzer enhanced with RecommendIndexWithImpact for benefit scoring
- API endpoints wired to real implementations (no more placeholders)
- Graceful fallback when hypopg extension not installed on monitored databases

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement unused index detection** - `55e53b6` (test/feat)
2. **Task 2: Implement hypothetical index impact estimation** - `29572fb` (test/feat)
3. **Task 3: Wire to storage and API endpoints** - `ac334aa` (feat)

**Plan metadata:** pending final commit

## Files Created/Modified

- `backend/internal/services/index_advisor/unused_detector.go` - Detects unused indexes from pg_stat_user_indexes
- `backend/internal/services/index_advisor/hypo_index.go` - Tests hypothetical indexes with hypopg
- `backend/internal/storage/index_recommendation_store.go` - Persists index recommendations
- `backend/migrations/030_add_hypopg_check.sql` - Adds hypopg_available column to databases table
- `backend/internal/services/index_advisor/analyzer.go` - Added RecommendIndexWithImpact method
- `backend/internal/api/handlers_index_advisor.go` - Real implementations for unused indexes and impact estimation
- `backend/internal/api/server.go` - Added estimate-impact route

## Decisions Made

1. **Hypopg for impact estimation**: Use hypopg extension for safe hypothetical index testing without creating real indexes. Graceful fallback when not available.

2. **Constraint exclusion via LEFT JOIN**: Query pg_stat_user_indexes with LEFT JOIN pg_constraint on conindid, filtering WHERE contype IS NULL to exclude primary keys, unique constraints, and foreign keys.

3. **Dynamic database connections**: Connect to monitored databases at request time using connection_string from postgresql_instances table, allowing per-database index analysis.

4. **Benefit scoring formula**: Combine cost improvement percentage with query frequency using CostCalculator.EstimateBenefit for weighted benefit scores.

## Deviations from Plan

None - plan executed exactly as written. Tasks 1 and 2 were already partially implemented from a previous phase, requiring only integration and wiring.

## Issues Encountered

None - all tests pass, build succeeds.

## User Setup Required

**External services require manual configuration.** See user_setup section in PLAN.md:

- Install hypopg extension on monitored databases: `CREATE EXTENSION hypopg;` (requires superuser)
- Fallback available: When hypopg unavailable, impact estimation returns improvement_pct of 0 with installation instructions

## Next Phase Readiness

- Index Intelligence features complete with detection, estimation, and recommendation capabilities
- Ready for integration testing with real PostgreSQL databases
- API endpoints functional: `/api/v1/index-advisor/database/:id/unused` and `/api/v1/index-advisor/database/:id/estimate-impact`

---
*Phase: 09-index-intelligence*
*Completed: 2026-05-13*

## Self-Check: PASSED

All key files verified on disk:
- unused_detector.go: FOUND
- hypo_index.go: FOUND
- index_recommendation_store.go: FOUND
- 030_add_hypopg_check.sql: FOUND

Commits verified:
- 55e53b6: test(09-02): add tests and implementation for unused index detection
- 29572fb: feat(09-02): implement hypothetical index impact estimation with hypopg
- ac334aa: feat(09-02): wire index advisor to storage and API endpoints