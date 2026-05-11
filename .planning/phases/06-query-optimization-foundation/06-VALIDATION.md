---
phase: 06
slug: query-optimization-foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-11
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (backend), Vitest (frontend) |
| **Config file** | `backend/go.mod`, `frontend/vitest.config.ts` |
| **Quick run command** | `cd backend && go test ./... -short` |
| **Full suite command** | `cd backend && go test ./... && cd frontend && npm test` |
| **Estimated runtime** | ~60 seconds (backend), ~30 seconds (frontend) |

---

## Sampling Rate

- **After every task commit:** Run `cd backend && go test ./... -short`
- **After every plan wave:** Run full suite (backend + frontend)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | API-02 | integration | `go test ./internal/storage/...` | ✅ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | API-03 | integration | `go test ./internal/storage/...` | ✅ W0 | ⬜ pending |
| 06-01-03 | 01 | 1 | API-04 | unit | `go test ./internal/storage/postgres_test.go` | ✅ W0 | ⬜ pending |
| 06-02-01 | 02 | 1 | QRY-01 | integration | `go test ./internal/handlers/... -run TestQueryPerformance` | ⬜ W0 | ⬜ pending |
| 06-02-02 | 02 | 1 | QRY-02 | integration | `go test ./internal/handlers/... -run TestQueryTimeline` | ⬜ W0 | ⬜ pending |
| 06-02-03 | 02 | 1 | QRY-05 | integration | `go test ./internal/handlers/... -run TestQueryStats` | ⬜ W0 | ⬜ pending |
| 06-02-04 | 02 | 1 | IDX-01 | integration | `go test ./internal/handlers/... -run TestIndexStats` | ⬜ W0 | ⬜ pending |
| 06-03-01 | 03 | 2 | MON-01 | manual | `curl localhost:8080/debug/pprof/` | N/A | ⬜ pending |
| 06-03-02 | 03 | 2 | MON-02 | unit | `go test ./internal/metrics/...` | ✅ W0 | ⬜ pending |
| 06-03-03 | 03 | 2 | MON-03 | unit | `go test ./internal/metrics/... -run TestHistograms` | ⬜ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/internal/storage/pgx_test.go` — tests for pgx pool integration
- [ ] `backend/internal/handlers/query_performance_test.go` — tests for slow query endpoints
- [ ] `backend/internal/metrics/performance_test.go` — tests for Prometheus histograms
- [ ] Extend `backend/tests/database/connection_pool_test.go` — test new pool metrics

*Existing infrastructure covers most phase requirements. Wave 0 adds specific test files.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| pprof endpoint accessibility | MON-01 | Requires running server | Start server, curl `localhost:8080/debug/pprof/`, verify HTML response |
| Prometheus metrics format | MON-02 | Requires metrics endpoint | Curl `localhost:8080/metrics`, verify `query_duration_seconds` histogram |
| Pool metrics accuracy | API-04 | Requires live database | Compare `db.Stats()` output with Prometheus gauge values |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending