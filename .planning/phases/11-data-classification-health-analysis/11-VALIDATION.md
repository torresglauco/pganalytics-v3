---
phase: 11
slug: data-classification-health-analysis
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-05-14
---

# Phase 11 — Validation Strategy

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
| 11-01-01 | 01 | 1 | DATA-01 | build | `go build ./backend/pkg/models` | ⬜ |
| 11-01-02 | 01 | 1 | DATA-01 | build | `grep -c "classification_results" backend/migrations/*.sql` | ⬜ |
| 11-01-03 | 01 | 1 | DATA-01 | build | `go build ./backend/internal/storage` | ⬜ |
| 11-01-04 | 01 | 1 | DATA-01 | build | `go build ./backend/internal/api` | ⬜ |
| 11-01-05 | 01 | 1 | DATA-01 | build | `grep -c "collectDataClassification" collector/src/data_classification_plugin.cpp` | ⬜ |
| 11-02-01 | 02 | 2 | HOST-04 | build | `go build ./backend/pkg/models` | ⬜ |
| 11-02-02 | 02 | 2 | HOST-04 | build | `grep -c "metrics_host_health" backend/migrations/*.sql` | ⬜ |
| 11-02-03 | 02 | 2 | HOST-04 | build | `go build ./backend/internal/storage` | ⬜ |
| 11-02-04 | 02 | 2 | HOST-04 | build | `go build ./backend/internal/api` | ⬜ |
| 11-03-01 | 03 | 2 | SCALE-01 | build | `grep -c "tenant_id" backend/pkg/models/models.go` | ⬜ |
| 11-03-02 | 03 | 2 | SCALE-02 | build | `grep -c "tenant_id" backend/migrations/*.sql` | ⬜ |
| 11-03-03 | 03 | 2 | SCALE-03 | build | `go build ./backend/internal/storage` | ⬜ |
| 11-03-04 | 03 | 2 | SCALE-04 | build | `go build ./backend/internal/api` | ⬜ |
| 11-04-01 | 04 | 3 | VER-03 | build | `go build ./backend/pkg/models` | ⬜ |
| 11-04-02 | 04 | 3 | VER-03 | build | `grep -c "postgres_health_checks" backend/migrations/*.sql` | ⬜ |
| 11-04-03 | 04 | 3 | VER-03 | build | `go build ./backend/internal/storage` | ⬜ |
| 11-04-04 | 04 | 3 | VER-03 | build | `go build ./backend/internal/api` | ⬜ |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

All tasks use build-based verification that does not require pre-existing test infrastructure.

**Existing infrastructure covers**: version detection tests, host status tests, inventory tests from Phase 10.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| PII detection accuracy | DATA-01 | Requires real data patterns | Insert test data with known PII patterns, verify detection results |
| Health score accuracy | HOST-04 | Requires multi-metric scenarios | Set CPU/memory/disk values, verify score calculation |
| RLS tenant isolation | SCALE-03 | Requires multi-tenant setup | Create multiple tenants, verify cross-tenant data isolation |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 90s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending