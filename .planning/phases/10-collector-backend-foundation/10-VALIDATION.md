---
phase: 10
slug: collector-backend-foundation
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-05-13
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: testing + testify; C++: Catch2 (via CMake) |
| **Config file** | Go: none (tests self-contained); C++: collector/tests/CMakeLists.txt |
| **Quick run command** | `go build ./backend/...` (compilation check) |
| **Full suite command** | `go test ./... -cover -race && cd collector/build && ctest -V` |
| **Estimated runtime** | ~90 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./backend/...` (compilation verification)
- **After every plan wave:** Run `go test ./... -cover -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 10-01-01 | 01 | 1 | REP-02 | build | `go build ./backend/pkg/models` | ✅ |
| 10-01-02 | 01 | 1 | REP-03 | build | `grep -c "metrics_replication_status" backend/migrations/031_replication_tables.sql` | ✅ |
| 10-01-03 | 01 | 1 | REP-01 | build | `go build ./backend/internal/storage` | ✅ |
| 10-01-04 | 01 | 1 | REP-01 | build | `go build ./backend/internal/api` | ✅ |
| 10-01-05 | 01 | 1 | REP-04 | build | `grep -c "handleGetReplicationMetrics" backend/internal/api/server.go` | ✅ |
| 10-02-01 | 02 | 1 | REP-02 | build | `go build ./backend/pkg/models` | ✅ |
| 10-02-02 | 02 | 1 | REP-03 | build | `grep -c "CREATE TABLE.*metrics_logical" backend/migrations/031_replication_tables.sql` | ✅ |
| 10-02-03 | 02 | 1 | REP-02/03 | build | `go build ./backend/internal/storage` | ✅ |
| 10-02-04 | 02 | 1 | REP-02/03 | build | `go build ./backend/internal/api` | ✅ |
| 10-02-05 | 02 | 1 | REP-02/03 | build | `grep -c "handleGetLogicalSubscriptions" backend/internal/api/server.go` | ✅ |
| 10-02-06 | 02 | 1 | REP-02/03 | build | `grep -c "collectLogicalSubscriptions" collector/src/logical_replication_plugin.cpp` | ✅ |
| 10-03-01 | 03 | 2 | HOST-01 | build | `go build ./backend/pkg/models` | ✅ |
| 10-03-02 | 03 | 2 | HOST-02 | build | `grep -c "metrics_host" backend/migrations/031_replication_tables.sql` | ✅ |
| 10-03-03 | 03 | 2 | HOST-01/02/03 | build | `go build ./backend/internal/storage` | ✅ |
| 10-03-04 | 03 | 2 | HOST-01/02/03 | build | `go build ./backend/internal/api` | ✅ |
| 10-03-05 | 03 | 2 | HOST-01/02/03 | build | `grep -c "handleGetHostStatus" backend/internal/api/server.go` | ✅ |
| 10-03-06 | 03 | 2 | HOST-03 | build | `grep -c "collectOsInfo" collector/src/host_inventory_plugin.cpp` | ✅ |
| 10-04-01 | 04 | 2 | INV-01 | build | `go build ./backend/pkg/models` | ✅ |
| 10-04-02 | 04 | 2 | INV-01-05 | build | `grep -c "metrics_table_inventory" backend/migrations/031_replication_tables.sql` | ✅ |
| 10-04-03 | 04 | 2 | INV-01-05 | build | `go build ./backend/internal/storage` | ✅ |
| 10-04-04 | 04 | 2 | INV-01-04 | build | `go build ./backend/internal/api` | ✅ |
| 10-04-05 | 04 | 2 | INV-01-04 | build | `grep -c "handleGetTableInventory" backend/internal/api/server.go` | ✅ |
| 10-04-06 | 04 | 2 | INV-01 | build | `grep -c "pg_total_relation_size" collector/src/schema_plugin.cpp` | ✅ |
| 10-05-01 | 05 | 3 | VER-01 | build | `go build ./backend/pkg/models` | ✅ |
| 10-05-02 | 05 | 3 | VER-02 | build | `go build ./backend/internal/storage` | ✅ |
| 10-05-03 | 05 | 3 | VER-01/02 | build | `go build ./backend/internal/api` | ✅ |
| 10-05-04 | 05 | 3 | VER-04 | build | `go build ./backend/tests/integration` | ✅ |
| 10-05-05 | 05 | 3 | VER-01/02 | build | `grep -l "postgres_version_major_" collector/src/*.cpp | wc -l` | ✅ |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

All tasks use build-based verification that does not require pre-existing test infrastructure.

**Integration tests** (created in Plan 05, Task 4):
- `backend/tests/integration/version_test.go` — Created during execution
- Tests for version detection, capabilities, and collector modes

**Existing infrastructure covers**: version detection, streaming replication plugin, sysstat plugin, schema plugin.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cascading topology accuracy | REP-03 | Requires multi-node PG cluster | Deploy primary → standby → standby chain, verify topology shows all nodes |
| Collector TLS handshake | COLL-05 | Requires cert generation | Run collector with mTLS, verify server accepts connection |
| Host down detection threshold | HOST-01 | Requires timing verification | Stop collector, wait >5min, verify status shows "down" |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 90s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending