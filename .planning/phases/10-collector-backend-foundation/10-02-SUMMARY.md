---
phase: 10
plan: 02
status: completed
completed_at: 2026-05-14T15:30:00Z
requirements:
  - REP-02
  - REP-03
---

# Plan 10-02 Summary: Logical Replication & Topology

## Completed Tasks

### Task 1: Extend replication models ✅
- Added `LogicalSubscription` struct with state tracking
- Added `Publication` struct with table ownership
- Added `WalReceiver` struct for PG 9.6+ wal status
- Added `ReplicationTopology` struct for cascading chains

### Task 2: Extend migration for logical tables ✅
- Added `metrics_logical_subscriptions` table
- Added `metrics_publications` table
- Added `metrics_wal_receivers` table
- All with TimescaleDB hypertables and indexes

### Task 3: Extend replication store ✅
- Implemented `StoreLogicalSubscriptions` and `GetLogicalSubscriptions`
- Implemented `StorePublications` and `GetPublications`
- Implemented `StoreWalReceivers` and `GetWalReceivers`
- Implemented `GetReplicationTopology` for cascading detection

### Task 4: Extend API handlers ✅
- Implemented `handleGetLogicalSubscriptions`
- Implemented `handleGetPublications`
- Implemented `handleGetReplicationTopology`
- Added Swagger annotations for all endpoints

### Task 5: Wire logical replication routes ✅
- Registered routes in server.go:
  - `GET /:id/logical-subscriptions`
  - `GET /:id/publications`
  - `GET /:id/topology`

### Task 6: Verify C++ collector plugins ✅
- Verified `replication_plugin.cpp` has version detection
- Verified `logical_replication_plugin.cpp` exists
- Both use version-adaptive queries

## Artifacts Created/Modified

| File | Purpose |
|------|---------|
| `backend/pkg/models/replication_models.go` | Extended with logical types |
| `backend/migrations/031_replication_tables.sql` | Extended with logical tables |
| `backend/internal/storage/replication_store.go` | Extended with logical operations |
| `backend/internal/api/handlers_replication.go` | Extended with logical endpoints |

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /:id/logical-subscriptions` | Subscription state and LSN |
| `GET /:id/publications` | Publication details with tables |
| `GET /:id/topology` | Cascading replication topology |

## Verification

- Build: `go build ./backend/...` ✅
- All 6 routes registered ✅
- Store functions implemented ✅

## Dependencies

- Plan 10-01 (streaming replication backend) completed first