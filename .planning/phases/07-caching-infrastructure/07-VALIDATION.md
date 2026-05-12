---
phase: 07
slug: caching-infrastructure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-12
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (backend) |
| **Config file** | `backend/go.mod` |
| **Quick run command** | `cd backend && go test ./... -short` |
| **Full suite command** | `cd backend && go test ./... -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd backend && go test ./internal/middleware/... ./internal/metrics/... -short`
- **After every plan wave:** Run full backend test suite
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 07-01-01 | 01 | 1 | API-01 | unit | `go test ./internal/middleware/... -run TestCacheMiddleware` | ⬜ W0 |
| 07-01-02 | 01 | 1 | API-01 | integration | `go test ./internal/api/... -run TestCachedEndpoints` | ⬜ W0 |
| 07-02-01 | 02 | 2 | MON-04 | unit | `go test ./internal/metrics/... -run TestCacheMetrics` | ⬜ W0 |
| 07-02-02 | 02 | 2 | MON-04 | integration | `go test ./internal/api/... -run TestCacheMetricsEndpoint` | ⬜ W0 |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/internal/middleware/cache_middleware_test.go` — tests for cache middleware
- [ ] `backend/internal/metrics/cache_metrics_test.go` — tests for cache metrics

*Existing cache implementation tests exist in `internal/cache/cache_test.go`.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cache hit rate in production | MON-04 | Requires traffic load | Generate load, check Prometheus metrics |
| Cache invalidation timing | API-01 | Requires running server | Modify data, verify cache clears |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending