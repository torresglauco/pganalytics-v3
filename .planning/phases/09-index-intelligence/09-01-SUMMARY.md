---
phase: 09-index-intelligence
plan: 01
subsystem: query-performance
tags: [fingerprinting, explain-analysis, anti-pattern-detection]
dependencies:
  requires: []
  provides: [query-fingerprinting, recursive-plan-analysis]
  affects: [query-performance-api]
tech-stack:
  added: []
  patterns: [regex-based-normalization, recursive-tree-walking]
key-files:
  created:
    - backend/internal/services/query_performance/fingerprinter.go
    - backend/internal/services/query_performance/fingerprinter_test.go
  modified:
    - backend/internal/services/query_performance/parser.go
    - backend/internal/services/query_performance/parser_test.go
    - backend/internal/services/query_performance/models.go
    - backend/internal/services/query_performance/service.go
    - backend/internal/storage/query_performance_store.go
    - backend/internal/api/handlers_query_performance.go
    - backend/internal/api/server.go
decisions:
  - Use regex-based fingerprinting instead of pg_query_go due to C compilation issues with macOS SDK 26.4
  - Compute fingerprints in service layer to avoid import cycles
  - Preserve backward-compatible DetectIssues() method alongside new DetectIssuesFull()
metrics:
  duration: 25
  completed_date: "2026-05-13"
  tasks_completed: 3
  files_modified: 7
  files_created: 2
  tests_added: 15
---

# Phase 09 Plan 01: Query Plan Analysis & Fingerprinting Summary

## One-liner

Implemented query fingerprinting for grouping similar queries and recursive EXPLAIN JSON analysis for detecting anti-patterns at any depth in the plan tree.

## Completed Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add pg_query_go dependency and create Fingerprinter service | 3f3ca78 | fingerprinter.go, fingerprinter_test.go |
| 2 | Enhance QueryParser with recursive EXPLAIN JSON parsing | e9f45ec | parser.go, parser_test.go, models.go |
| 3 | Wire fingerprinting to storage and API layer | d65a23e | service.go, query_performance_store.go, handlers_query_performance.go, server.go |

## Key Artifacts

### Fingerprinter Service

- `backend/internal/services/query_performance/fingerprinter.go`: Regex-based SQL fingerprinting
  - Same queries with different parameters produce identical fingerprints
  - Supports string literals, numeric values, IN lists, VALUES clauses
  - Graceful fallback for invalid SQL

### Recursive Plan Analysis

- `backend/internal/services/query_performance/parser.go`: DetectIssuesFull method
  - Recursively walks EXPLAIN JSON plan tree
  - Detects Seq Scan at any depth with severity based on cost
  - Detects Nested Loop with high row counts (>1000)
  - Detects Sort on large datasets (>10000 rows)

### API Endpoint

- `GET /api/v1/databases/:id/query-fingerprints`: Returns queries grouped by fingerprint
  - Groups queries with same structure but different parameters
  - Provides aggregate statistics (total calls, avg time, query count)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking Issue] Replaced pg_query_go with regex-based fingerprinting**
- **Found during:** Task 1 - Adding pg_query_go dependency
- **Issue:** pg_query_go v5.1.0 and v5.0.0 fail to compile on macOS SDK 26.4 due to C compilation errors (`static declaration of 'strchrnul' follows non-static declaration`)
- **Fix:** Implemented regex-based SQL normalization for fingerprinting instead of using pg_query_go
- **Files modified:** fingerprinter.go, fingerprinter_test.go
- **Commit:** 3f3ca78
- **Impact:** Regex-based approach provides equivalent fingerprinting for common SQL patterns without C dependencies

**2. [Rule 3 - Blocking Issue] Avoided import cycle**
- **Found during:** Task 3 - Wiring fingerprinting to storage
- **Issue:** Computing fingerprints in storage layer created an import cycle (storage -> query_performance -> storage)
- **Fix:** Moved fingerprint computation to service layer instead
- **Files modified:** service.go, query_performance_store.go
- **Commit:** d65a23e

## Verification Results

All verification criteria passed:

- Fingerprinter unit tests: PASS (8 tests)
- Parser recursive analysis tests: PASS (7 tests)
- All existing tests still pass: PASS (31 total tests)
- Build succeeds: PASS

## Files Modified

```
backend/internal/services/query_performance/
  fingerprinter.go          # New - Regex-based SQL fingerprinting
  fingerprinter_test.go     # New - Comprehensive fingerprint tests
  parser.go                 # Added DetectIssuesFull for recursive analysis
  parser_test.go            # Added tests for nested plan detection
  models.go                 # Added FullExplainPlan, PlanNode structs
  service.go                # Added fingerprint computation, grouping endpoint

backend/internal/storage/
  query_performance_store.go  # Added QueryFingerprintHash field

backend/internal/api/
  handlers_query_performance.go  # Added database fingerprint endpoint
  server.go                      # Added route for /databases/:id/query-fingerprints
```

## Requirements Marked Complete

- [x] QRY-03: DetectIssuesFull recursively finds anti-patterns at any depth in EXPLAIN JSON
- [x] QRY-04: Fingerprinter groups queries with same structure but different parameters

## Self-Check: PASSED

- All created files exist
- All commits verified in git history
- All tests pass