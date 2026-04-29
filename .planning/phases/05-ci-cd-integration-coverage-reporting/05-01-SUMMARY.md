---
phase: 05-ci-cd-integration-coverage-reporting
plan: 01
subsystem: ci-cd
tags: [github-actions, codecov, coverage, quality-gate]
dependency_graph:
  requires: []
  provides: [QUAL-05, TEST-17, TEST-19, TEST-20]
  affects: [branch-protection, pr-workflow]
tech_stack:
  added:
    - codecov/codecov-action@v4
    - Codecov YAML configuration
  patterns:
    - Unified CI quality gate with summary job
    - Coverage threshold enforcement via Codecov
    - Per-test timing via -v flag (Go) and verbose reporter (Vitest)
key_files:
  created:
    - .github/workflows/ci.yml (174 lines)
    - codecov.yml (38 lines)
  modified:
    - frontend/vite.config.ts (coverage thresholds added)
decisions:
  - Use unified CI workflow with ci-passed summary job for branch protection
  - Use Codecov YAML for threshold enforcement instead of inline scripts
  - Set 80% coverage target with 1% threshold to allow gradual improvement
  - Branch coverage set to 70% (lower than lines/functions due to conditional logic difficulty)
metrics:
  duration: 5min
  tasks: 3
  files: 3
  completed_date: 2026-04-29
---

# Phase 05 Plan 01: CI Quality Gate with Coverage Reporting Summary

**One-liner:** Unified GitHub Actions workflow with Codecov integration for backend and frontend coverage reporting, quality gates, and test performance tracking.

## What Was Built

### 1. Unified CI Quality Gate Workflow (`.github/workflows/ci.yml`)

Created a comprehensive CI workflow with 4 jobs:

- **backend job**: PostgreSQL 16 + TimescaleDB service containers, Go tests with race detection and coverage, Codecov upload with `fail_ci_if_error: true`
- **frontend job**: Node 18 setup, Vitest coverage with lcov reporter, Codecov upload
- **lint job**: golangci-lint v2.11.4 for backend, ESLint for frontend
- **ci-passed job**: Summary job for branch protection - single status check that aggregates all quality gates

### 2. Codecov Configuration (`codecov.yml`)

Project-level coverage enforcement:

- **Project target**: 80% coverage with 1% threshold (allows gradual improvement)
- **Patch target**: 80% coverage with 5% threshold for PR changes
- **Flags**: Separate `backend` and `frontend` flags for independent tracking
- **Exclusions**: Test files, E2E tests, main entry points excluded from coverage

### 3. Vitest Coverage Thresholds (`frontend/vite.config.ts`)

Added coverage thresholds to frontend test configuration:

- Lines: 80%
- Functions: 80%
- Branches: 70% (lower due to conditional logic complexity)
- Statements: 80%
- Added `lcov` reporter for Codecov compatibility

## Key Technical Decisions

| Decision | Rationale |
|----------|-----------|
| Use `ci-passed` summary job | Single status check for branch protection simplifies GitHub settings |
| Use Codecov YAML for thresholds | Better than inline scripts - handles edge cases, trend analysis, PR comments |
| Set threshold to 1% | Allows gradual improvement without blocking every PR |
| Branch coverage at 70% | Branch coverage is harder to achieve, especially with error handling paths |
| Use `fail_ci_if_error: true` | Ensures coverage upload failures are visible and block CI |

## Requirements Satisfied

| Requirement | Status | Evidence |
|-------------|--------|----------|
| QUAL-05 | COMPLETE | Codecov configuration with 80% target, lcov reporter for frontend |
| TEST-17 | COMPLETE | Unified CI workflow runs all tests on push/PR |
| TEST-19 | COMPLETE | Codecov uploads for both backend and frontend with flags |
| TEST-20 | COMPLETE | `-v` flag for Go tests, Vitest verbose output shows timing |

## Deviations from Plan

None - plan executed exactly as written.

## Files Modified

```
.github/workflows/ci.yml    +174 lines (NEW)
codecov.yml                 +38 lines (NEW)
frontend/vite.config.ts      +8 lines (MODIFIED)
```

## Next Steps

1. Add `CODECOV_TOKEN` secret to GitHub repository settings
2. Configure branch protection rule to require `ci-passed` status check
3. Monitor coverage trends on Codecov dashboard
4. Plan test expansion to reach 80% coverage target

## Self-Check: PASSED

- All files exist and verified
- All commits created successfully
- YAML syntax valid
- TypeScript config valid