---
phase: 14
slug: testing-quality
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-05-15"
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Backend Framework** | Go testing + testify |
| **Frontend Framework** | Vitest + @testing-library/react |
| **Collector Framework** | GTest (C++) |
| **E2E Framework** | Playwright |
| **Backend Config** | `backend/` (go.mod) |
| **Frontend Config** | `frontend/vite.config.ts` |
| **Quick run (backend)** | `go test ./backend/... -short` |
| **Quick run (frontend)** | `npm run test --prefix frontend` |
| **Full suite** | `go test ./backend/... && npm run test:coverage --prefix frontend` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run relevant test suite (backend OR frontend)
- **After every plan wave:** Run full suite for affected areas
- **Before `/gsd:verify-work`:** All test suites green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 14-01-01 | 01 | 1 | TEST-01 | unit (C++) | `cd collector && ctest` | pending |
| 14-01-02 | 01 | 1 | TEST-02 | unit (Go) | `go test ./backend/pkg/classification/...` | pending |
| 14-02-01 | 02 | 1 | TEST-03 | integration | `go test ./backend/tests/integration/...` | pending |
| 14-02-02 | 02 | 1 | TEST-03 | integration | `go test ./backend/tests/integration/alerts_test.go` | pending |
| 14-03-01 | 03 | 2 | TEST-04 | unit (TS) | `npm run test --prefix frontend` | pending |
| 14-03-02 | 03 | 2 | TEST-04 | unit (TS) | `vitest run src/components/topology/` | pending |
| 14-04-01 | 04 | 2 | TEST-05 | e2e | `npx playwright test` | pending |

*Status: pending | green | red | flaky*

---

## Wave 0 Requirements

### Backend Tests (from Phases 10-12 Wave 0 gaps)
- [ ] `backend/pkg/classification/classifier_test.go` — data classification unit tests
- [ ] `backend/pkg/services/health_scorer_test.go` — health scoring logic tests
- [ ] `backend/internal/middleware/tenant_context_test.go` — RLS middleware tests
- [ ] `backend/tests/integration/replication_test.go` — replication API tests
- [ ] `backend/tests/integration/host_monitoring_test.go` — host API tests
- [ ] `backend/tests/integration/alert_rules_test.go` — alert CRUD tests

### Frontend Tests (from Phase 13 Wave 0 gaps)
- [ ] `frontend/src/components/topology/TopologyGraph.test.tsx` — topology rendering
- [ ] `frontend/src/components/classification/ClassificationTable.test.tsx` — table tests
- [ ] `frontend/src/components/host/HostStatusTable.test.tsx` — host table tests
- [ ] `frontend/src/pages/ReplicationTopologyPage.test.tsx` — page integration
- [ ] `frontend/src/pages/DataClassificationPage.test.tsx` — page integration
- [ ] `frontend/src/pages/HostInventoryPage.test.tsx` — page integration

### Collector Tests
- [ ] `collector/tests/unit/data_classification_test.cpp` — pattern detection tests
- [ ] `collector/tests/unit/host_metrics_test.cpp` — metrics collection tests

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Multi-tenant RLS isolation | TEST-02 | Requires multi-database setup | Create two tenants, verify data isolation via API |
| E2E alert notification flow | TEST-05 | External service dependency | Create alert rule, trigger condition, verify email sent |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending