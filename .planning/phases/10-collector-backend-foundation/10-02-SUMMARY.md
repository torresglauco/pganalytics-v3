---
phase: 10-collector-backend-foundation
plan: 02
subsystem: database
tags: [postgresql, replication, logical-replication, topology, timescaledb, cpp-collector]

# Dependency graph
requires:
  - phase: 10-collector-backend-foundation
    plan: 01
    provides: Streaming replication models and storage (created alongside this plan)
provides:
  - Logical replication subscription monitoring via pg_stat_subscription
  - Publication monitoring via pg_publication
  - WAL receiver status for topology detection via pg_stat_wal_receiver
  - Cascading replication topology view
  - C++ collector plugin for logical replication metrics
affects: [frontend-replication-ui, alerting-replication]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Version-adaptive PostgreSQL queries (PG 9.6+, 10+)
    - Multi-database iteration for per-database metrics
    - Cascading topology detection from wal_receiver + replication_status

key-files:
  created:
    - backend/internal/storage/replication_store.go
    - backend/internal/api/handlers_replication.go
    - collector/include/logical_replication_plugin.h
    - collector/src/logical_replication_plugin.cpp
  modified:
    - backend/pkg/models/replication_models.go
    - backend/migrations/031_replication_tables.sql
    - backend/internal/api/server.go

key-decisions:
  - "Created replication_store.go with both streaming and logical replication functions in one file for cohesion"
  - "Topology detection uses combination of wal_receiver (incoming) and replication_status (outgoing) to determine node role"
  - "C++ collector iterates databases for subscriptions/publications which are per-database objects"

patterns-established:
  - "Pattern: Store and Get functions follow prepared statement pattern with ON CONFLICT DO NOTHING"
  - "Pattern: Version detection in collector enables backward-compatible queries"
  - "Pattern: Topology node role: primary (no wal_receiver), standby (has wal_receiver), cascading_standby (has both)"

requirements-completed:
  - REP-02
  - REP-03

# Metrics
duration: 18min
completed: 2026-05-14
---
# Phase 10 Plan 02: Logical Replication & Topology Summary

**Logical replication monitoring with subscriptions, publications, WAL receiver status, and cascading topology detection for PostgreSQL 10+**

## Performance

- **Duration:** 18 min
- **Started:** 2026-05-14T14:28:27Z
- **Completed:** 2026-05-14T14:46:15Z
- **Tasks:** 6
- **Files modified:** 6

## Accomplishments
- Extended replication models with LogicalSubscription, Publication, WalReceiver, and ReplicationTopology structs
- Created migration with 3 new TimescaleDB hypertables for logical replication data
- Implemented replication_store.go with 7 storage functions for streaming and logical replication
- Created 5 API handlers for replication metrics, slots, subscriptions, publications, and topology
- Wired 5 new routes with authentication middleware
- Built C++ logical_replication_plugin with version-adaptive queries and multi-database iteration

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend replication models for logical replication and topology** - `3c9f351` (feat)
2. **Task 2: Extend database migration for logical replication tables** - `fc84edc` (feat)
3. **Task 3: Extend replication store for logical replication** - `3e4b265` (feat)
4. **Task 4: Extend replication API handlers for logical replication** - `4850759` (feat)
5. **Task 5: Wire logical replication routes in server** - `4850759` (feat)
6. **Task 6: Create C++ logical replication collector plugin** - `7bc188b` (feat)

## Files Created/Modified
- `backend/pkg/models/replication_models.go` - Added 5 new structs: LogicalSubscription, Publication, WalReceiver, TopologyNode, ReplicationTopology
- `backend/migrations/031_replication_tables.sql` - Added 3 new tables: metrics_logical_subscriptions, metrics_publications, metrics_wal_receivers
- `backend/internal/storage/replication_store.go` - New file with 7 functions for replication storage
- `backend/internal/api/handlers_replication.go` - New file with 5 API handlers
- `backend/internal/api/server.go` - Added 5 routes for replication endpoints
- `collector/include/logical_replication_plugin.h` - New header for C++ collector
- `collector/src/logical_replication_plugin.cpp` - New implementation with 3 collection methods

## Decisions Made
- Combined streaming and logical replication storage in one file (replication_store.go) for logical cohesion
- Topology detection determines node role from wal_receiver (incoming) and replication_status (outgoing) presence
- C++ collector iterates databases since subscriptions and publications are per-database objects
- Used prepared statements with ON CONFLICT DO NOTHING for idempotent inserts

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Created replication_store.go from scratch**
- **Found during:** Task 3 (Extend replication store)
- **Issue:** Plan expected file from 10-01 but it did not exist
- **Fix:** Created complete file with both streaming and logical replication functions following existing patterns from metrics_store.go
- **Files modified:** backend/internal/storage/replication_store.go
- **Verification:** go build ./backend/internal/storage passes
- **Committed in:** 3e4b265 (Task 3 commit)

**2. [Rule 3 - Blocking] Created handlers_replication.go from scratch**
- **Found during:** Task 4 (Extend replication API handlers)
- **Issue:** Plan expected file from 10-01 but it did not exist
- **Fix:** Created complete file with all 5 handlers following existing patterns from handlers_metrics.go
- **Files modified:** backend/internal/api/handlers_replication.go
- **Verification:** go build ./backend/internal/api passes
- **Committed in:** 4850759 (Task 4 commit)

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both auto-fixes were necessary to unblock task execution. The plan's dependency on 10-01 was implicit - the files needed to be created regardless of which plan created them.

## Issues Encountered
- Pre-existing test build errors in backend/tests/integration unrelated to this plan's changes - not in scope to fix
- gofmt formatting issues resolved during commit hooks

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Logical replication API endpoints ready for frontend consumption
- C++ collector ready for integration with collector manager
- Topology detection logic ready for visualization in UI
- Migration ready for database deployment

---
*Phase: 10-collector-backend-foundation*
*Completed: 2026-05-14*