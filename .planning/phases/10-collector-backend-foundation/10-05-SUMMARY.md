---
phase: 10-collector-backend-foundation
plan: 05
subsystem: multi-version-support
tags: [version-detection, capabilities, eol-tracking, collector-modes, decentralized, centralized, cpp-collector]

# Dependency graph
requires:
  - phase: 10-01
    provides: TimescaleDB infrastructure and collector authentication
provides:
  - PostgreSQL version detection and capability tracking for 11-17
  - EOL status tracking for version lifecycle management
  - Collector deployment mode detection (decentralized vs centralized)
  - TLS configuration support for secure connections
affects: [alerting, dashboard, capacity-planning, replication, inventory]

# Tech tracking
tech-stack:
  added: []
  patterns: [version-capabilities, eol-tracking, mode-detection, prepared-statements]

key-files:
  created:
    - backend/pkg/models/version_models.go
    - backend/internal/storage/version_store.go
    - backend/internal/api/handlers_version.go
    - backend/tests/integration/version_test.go
  modified:
    - backend/internal/api/server.go
    - collector/src/host_inventory_plugin.cpp
    - collector/include/host_inventory_plugin.h

key-decisions:
  - "Used version_num (integer) for parsing - matches PostgreSQL's server_version_num format"
  - "Capability flags per version for feature detection without runtime queries"
  - "Mode detection based on connection type (localhost/Unix socket = decentralized)"
  - "EOL dates sourced from official PostgreSQL versioning policy"

patterns-established:
  - "Version capabilities as boolean flags for compile-time feature detection"
  - "Mode detection from postgres_host connection string analysis"
  - "C++ collector version detection using server_version_num GUC"

requirements-completed: [VER-01, VER-02, VER-04, COLL-01, COLL-02, COLL-03, COLL-04, COLL-05]

# Metrics
duration: 35min
completed: 2026-05-14
---

# Phase 10 Plan 05: Multi-Version Support Summary

**Multi-version PostgreSQL support with version detection, capability tracking, EOL status, and collector deployment mode detection for decentralized and centralized architectures**

## Performance

- **Duration:** 35 min
- **Started:** 2026-05-14T15:15:00Z
- **Completed:** 2026-05-14T15:50:00Z
- **Tasks:** 6
- **Files modified:** 6

## Accomplishments
- PostgreSQL version detection for versions 11-17 with major/minor parsing
- Version-specific capability tracking (HasWriteLagColumns, HasLogicalReplication, HasPgStatWal, etc.)
- EOL status tracking for version lifecycle management
- Collector deployment mode detection (decentralized vs centralized)
- TLS configuration support for secure connections
- 3 REST API endpoints for version and mode queries
- Version detection in all 3 C++ collectors (replication, logical_replication, host_inventory)

## Task Commits

Each task was committed atomically:

1. **Task 1: Create version models** - `ba7f5ee` (feat)
2. **Task 2: Create version store** - `11af6cb` (feat)
3. **Task 3: Create version API handlers** - `814baff` (feat)
4. **Task 4: Wire version routes and create integration tests** - `814baff`, `ee53e02` (feat, test)
5. **Task 5: Verify and enhance version-adaptive queries in C++ collectors** - `e24fbc5` (feat)
6. **Task 6: Verify multi-version support and collector modes** - Human-verified checkpoint passed

## Files Created/Modified
- `backend/pkg/models/version_models.go` - 6 structs for version, capabilities, and mode config
- `backend/internal/storage/version_store.go` - 7 functions for version detection and capabilities
- `backend/internal/api/handlers_version.go` - 3 HTTP handlers with Swagger annotations
- `backend/tests/integration/version_test.go` - 6 test functions for version detection
- `backend/internal/api/server.go` - 3 version routes registered
- `collector/src/host_inventory_plugin.cpp` - Added version detection following existing pattern
- `collector/include/host_inventory_plugin.h` - Added version member variables

## Decisions Made
- Used `server_version_num` GUC for reliable version detection in C++ collectors
- Capability flags derived from PostgreSQL feature matrix documentation
- Mode detection based on postgres_host analysis (localhost/Unix socket vs remote)
- EOL dates tracked for PostgreSQL 11-17 from official versioning policy

## Deviations from Plan

None - plan executed exactly as written.

## API Endpoints Added

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/versions/supported` | List all supported PostgreSQL versions with EOL dates |
| GET | `/api/v1/collectors/:id/version` | Get version info and capabilities for a collector |
| GET | `/api/v1/collectors/:id/mode` | Get deployment mode for a collector |

## Version Capabilities Tracked

| Capability | Minimum Version | Description |
|------------|-----------------|-------------|
| HasWriteLagColumns | PG 13 | write_lag, flush_lag, replay_lag columns in pg_stat_replication |
| HasWalReceiver | PG 10 | pg_stat_wal_receiver view |
| HasLogicalReplication | PG 10 | Native logical replication support |
| HasPublication | PG 10 | CREATE PUBLICATION support |
| HasStandbySignal | PG 12 | standby.signal file support |
| HasPgStatWal | PG 14 | pg_stat_wal view |
| HasPgStatSubscription | PG 10 | pg_stat_subscription view |

## Test Results

All version detection tests pass:
```
=== RUN   TestPostgreSQLVersionDetection
--- PASS: TestPostgreSQLVersionDetection (0.00s)
=== RUN   TestPostgreSQLVersionParsingEdgeCases
--- PASS: TestPostgreSQLVersionParsingEdgeCases (0.00s)
=== RUN   TestIsVersionSupported
--- PASS: TestIsVersionSupported (0.00s)
=== RUN   TestGetAllSupportedVersions
--- PASS: TestGetAllSupportedVersions (0.00s)
```

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Multi-version support complete, ready for version-adaptive queries
- Collector mode detection enables centralized monitoring architecture
- EOL tracking enables proactive version upgrade planning

## Self-Check: PASSED

All files and commits verified:
- backend/pkg/models/version_models.go - FOUND
- backend/internal/storage/version_store.go - FOUND
- backend/internal/api/handlers_version.go - FOUND
- backend/tests/integration/version_test.go - FOUND
- .planning/phases/10-collector-backend-foundation/10-05-SUMMARY.md - FOUND
- Commits ba7f5ee, 11af6cb, 814baff, ee53e02, e24fbc5 - ALL FOUND

---
*Phase: 10-collector-backend-foundation*
*Completed: 2026-05-14*