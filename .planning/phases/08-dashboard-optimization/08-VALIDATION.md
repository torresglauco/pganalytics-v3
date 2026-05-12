---
phase: 08
slug: dashboard-optimization
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-12
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (backend) |
| **Config file** | `backend/go.mod` |
| **Quick run command** | `cd backend && go test ./internal/timescale/... ./internal/jobs/... -short` |
| **Full suite command** | `cd backend && go test ./... -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd backend && go test ./internal/timescale/... ./internal/jobs/... -short`
- **After every plan wave:** Run full backend test suite
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 08-01-01 | 01 | 1 | DASH-02 | unit | `go test ./internal/timescale/... -run TestContinuousAggregates` | ⬜ W0 |
| 08-01-02 | 01 | 1 | DASH-02 | integration | `go test ./tests/integration/... -run TestAggregateViews` | ⬜ W0 |
| 08-02-01 | 02 | 2 | DASH-04 | unit | `go test ./internal/jobs/... -run TestDashboardAggregationWorker` | ⬜ W0 |
| 08-02-02 | 02 | 2 | DASH-01 | integration | `go test ./tests/integration/... -run TestDashboardMetrics` | ⬜ W0 |
| 08-02-03 | 02 | 2 | DASH-03 | integration | `go test ./tests/integration/... -run TestHistoricalMetrics` | ⬜ W0 |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/internal/timescale/aggregates_test.go` — tests for continuous aggregate management
- [ ] `backend/internal/timescale/aggregate_queries_test.go` — tests for pre-computed queries
- [ ] `backend/internal/jobs/dashboard_aggregation_worker_test.go` — tests for background worker
- [ ] `backend/tests/integration/dashboard_aggregates_test.go` — integration tests with testcontainers

*Existing test infrastructure (testcontainers, Go testing) covers framework requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Dashboard load speed in production | DASH-01 | Requires TimescaleDB with data | Query dashboard endpoint, verify <100ms response |
| Continuous aggregate refresh timing | DASH-02 | Requires running TimescaleDB | Check `timescaledb_information.jobs` for refresh status |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending