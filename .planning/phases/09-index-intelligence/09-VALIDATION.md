---
phase: 09
slug: index-intelligence
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-12
---

# Phase 09 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (backend) |
| **Config file** | `backend/go.mod` |
| **Quick run command** | `cd backend && go test ./internal/services/query_performance/... ./internal/services/index_advisor/... -short` |
| **Full suite command** | `cd backend && go test ./... -v` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd backend && go test ./internal/services/query_performance/... ./internal/services/index_advisor/... -short`
- **After every plan wave:** Run full backend test suite
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 09-01-01 | 01 | 1 | QRY-03 | unit | `go test ./internal/services/query_performance/... -run TestDetectSeqScan` | ⬜ W0 |
| 09-01-02 | 01 | 1 | QRY-03 | unit | `go test ./internal/services/query_performance/... -run TestDetectNestedLoop` | ⬜ W0 |
| 09-01-03 | 01 | 1 | QRY-04 | unit | `go test ./internal/services/query_performance/... -run TestFingerprint` | ⬜ W0 |
| 09-02-01 | 02 | 2 | IDX-02 | unit | `go test ./internal/services/index_advisor/... -run TestUnusedIndexes` | ⬜ W0 |
| 09-02-02 | 02 | 2 | IDX-03 | unit | `go test ./internal/services/index_advisor/... -run TestIndexImpact` | ⬜ W0 |
| 09-02-03 | 02 | 2 | IDX-04 | unit | `go test ./internal/services/index_advisor/... -run TestBenefitScore` | Partial |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/internal/services/query_performance/fingerprinter_test.go` — tests for query fingerprinting
- [ ] `backend/internal/services/query_performance/plan_analyzer_test.go` — tests for recursive EXPLAIN analysis
- [ ] `backend/internal/services/index_advisor/unused_finder_test.go` — tests for unused index detection
- [ ] `backend/internal/services/index_advisor/impact_estimator_test.go` — tests for index impact estimation

*Existing test infrastructure covers basic patterns. Extensions needed for new functionality.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| hypopg extension availability | IDX-03 | Requires superuser access | Check `SELECT * FROM pg_extension WHERE extname = 'hypopg'` |
| Index recommendation accuracy | IDX-04 | Requires production workload | Compare recommendations with DBA analysis |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending