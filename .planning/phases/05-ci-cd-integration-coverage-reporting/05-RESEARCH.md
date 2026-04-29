# Phase 5: CI/CD Integration & Coverage Reporting - Research

**Researched:** 2026-04-29
**Domain:** CI/CD automation, coverage reporting, quality gates, unused code detection
**Confidence:** HIGH (project-specific analysis) / MEDIUM (external tooling versions)

## Summary

Phase 5 requires integrating existing test infrastructure with coverage reporting and quality gates. The project already has GitHub Actions workflows for backend tests, frontend quality checks, E2E tests, and security scanning. However, these workflows lack: (1) unified coverage reporting to Codecov, (2) strict quality gates that block PRs on failures, (3) test performance tracking, and (4) automated unused code cleanup enforcement.

Current coverage baselines are low: backend at 9.1%, frontend tests partially passing (393/410 tests pass). The 80% target requires significant test expansion, but this phase focuses on CI/CD infrastructure to track and enforce coverage, not writing additional tests.

**Primary recommendation:** Enhance existing workflows with coverage thresholds, Codecov integration, and branch protection rules rather than creating new workflows from scratch.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QUAL-05 | Coverage tracking and reporting (80%+ target) | Vitest v8 coverage, Go coverprofile, Codecov YAML for thresholds |
| QUAL-06 | Unused code cleanup | golangci-lint `unused` linter, ESLint `no-unused-vars` rule (already configured) |
| TEST-17 | CI/CD pipeline test execution | Existing workflows in `.github/workflows/`, need consolidation |
| TEST-18 | Test failures block deployment | GitHub branch protection rules, required status checks |
| TEST-19 | Coverage reports published | Codecov action integration, artifact uploads |
| TEST-20 | Performance tracking (test execution time) | Go `-json` flag with custom parser, Vitest `--reporter=verbose`, Playwright HTML reports |

</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| GitHub Actions | N/A | CI/CD platform | Native GitHub integration, already in use |
| Codecov | v4 | Coverage reporting | Industry standard, free for open source, supports both Go and JavaScript |
| golangci-lint | v2.11.4 | Go linting with unused detection | Already installed, `unused` linter available |
| Vitest | v1.0.0 | Frontend unit testing + coverage | Already configured with v8 coverage provider |
| Playwright | v1.59.1 | E2E testing | Already configured with multi-browser support |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| action-coverage | v3 | Codecov GitHub Action | Coverage upload to codecov.io |
| golangci-lint-action | v3 | Run golangci-lint in CI | Already in use for linting |
| actions/upload-artifact | v4 | Store test artifacts | Test reports, coverage HTML |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Codecov | Coveralls | Coveralls has less detailed PR comments, similar free tier |
| Vitest coverage | Istanbul/nyc | Vitest's v8 provider is faster, already integrated |
| golangci-lint unused | deadcode standalone | golangci-lint integrates multiple linters, already configured |

**Installation:**

```bash
# No new dependencies needed - all tools already installed
# Codecov configuration file needed:
touch codecov.yml
```

**Version verification:**

```
golangci-lint: v2.11.4 (verified 2026-04-29)
Vitest: v1.0.0 (from package.json)
Playwright: v1.59.1 (from package.json)
```

## Architecture Patterns

### Recommended CI Workflow Structure

```
.github/workflows/
├── ci.yml              # Unified CI gate (NEW - combines test + quality)
├── backend-tests.yml   # Existing - keep for backend-specific triggers
├── frontend-quality.yml # Existing - keep for frontend-specific triggers
├── e2e-tests.yml       # Existing - keep for E2E-specific triggers
└── security.yml        # Existing - keep for weekly security scans
```

### Pattern 1: Unified CI Quality Gate

**What:** Single workflow that must pass before PR merge, combining all quality checks.

**When to use:** For branch protection rules - one check to mark as required.

**Example:**

```yaml
# .github/workflows/ci.yml
name: CI Quality Gate

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  backend:
    name: Backend Tests + Coverage
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: pganalytics_test
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: Run tests with coverage
        working-directory: backend
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/pganalytics_test?sslmode=disable
      - name: Check coverage threshold
        working-directory: backend
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "::error::Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi
      - name: Upload to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./backend/coverage.out
          flags: backend
          fail_ci_if_error: true
          token: ${{ secrets.CODECOV_TOKEN }}

  frontend:
    name: Frontend Tests + Coverage
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      - name: Run tests with coverage
        working-directory: frontend
        run: npm run test:coverage -- --run
      - name: Upload to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./frontend/coverage/coverage-final.json
          flags: frontend
          fail_ci_if_error: true
          token: ${{ secrets.CODECOV_TOKEN }}

  lint:
    name: Lint (Unused Code Check)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: golangci-lint with unused
        uses: golangci/golangci-lint-action@v3
        with:
          version: v2.11.4
          working-directory: backend
          args: --timeout=5m
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Frontend lint
        working-directory: frontend
        run: npm ci && npm run lint

  # Summary job for branch protection
  ci-passed:
    name: CI Passed
    runs-on: ubuntu-latest
    needs: [backend, frontend, lint]
    if: always()
    steps:
      - name: Check all jobs
        run: |
          if [[ "${{ needs.backend.result }}" != "success" || \
                "${{ needs.frontend.result }}" != "success" || \
                "${{ needs.lint.result }}" != "success" ]]; then
            echo "CI failed"
            exit 1
          fi
          echo "CI passed"
```

### Pattern 2: Codecov YAML Configuration

**What:** Project-level coverage threshold enforcement.

**When to use:** To block PRs that decrease coverage below threshold.

**Example:**

```yaml
# codecov.yml
coverage:
  precision: 2
  round: down
  range: "70...100"
  status:
    project:
      default:
        target: 80%
        threshold: 1%  # Allow 1% decrease without failing
        if_ci_failed: error
    patch:
      default:
        target: 80%
        threshold: 5%

comment:
  layout: "header, diff, files, footer"
  behavior: default
  require_changes: true

flags:
  backend:
    paths:
      - backend/
    carryforward: true
  frontend:
    paths:
      - frontend/src/
    carryforward: true

ignore:
  - "backend/cmd/**"
  - "backend/tests/**"
  - "frontend/src/test/**"
  - "frontend/e2e/**"
  - "**/*.test.*"
  - "**/*.spec.*"
```

### Pattern 3: Test Performance Tracking

**What:** Capture and report test execution times.

**When to use:** To identify slow tests (> 5 seconds) for optimization.

**Example (Go):**

```bash
# Run tests with JSON output for timing analysis
go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | "\(.Package) \(.Elapsed)s"' | sort -k2 -n -r | head -20
```

**Example (Vitest):**

```typescript
// vitest.config.ts - add reporter
export default defineConfig({
  test: {
    reporters: ['verbose'],
    // ... existing config
  }
})
```

### Anti-Patterns to Avoid

- **Separate workflows for each check without unified gate:** Makes branch protection complex with many required checks.
- **Coverage as warning only:** Does not enforce quality improvement; should fail CI.
- **Ignoring integration tests in coverage:** Integration tests often cover critical paths; include them.
- **Using `continue-on-error: true` for quality checks:** Defeats the purpose of CI gates.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Coverage threshold checking | Shell script parsing coverage output | Codecov YAML with `target` setting | Codecov handles edge cases, trend analysis, and PR comments |
| Test timing analysis | Custom timing scripts | Built-in `-json` flag (Go), `--reporter=verbose` (Vitest) | Native tools handle timing accurately |
| Branch protection | Manual PR review requirements | GitHub required status checks | Automation prevents human error |
| Unused code detection | Regex-based grep patterns | golangci-lint `unused`, ESLint `no-unused-vars` | Static analysis handles imports, dead code, shadowed variables |

**Key insight:** The CI/CD ecosystem has mature tools for all Phase 5 requirements. Custom solutions introduce maintenance burden and miss edge cases.

## Common Pitfalls

### Pitfall 1: Coverage Threshold Too Strict Initially

**What goes wrong:** Setting 80% threshold immediately blocks all PRs due to current 9.1% backend coverage.

**Why it happens:** Current coverage is far below target; immediate enforcement would halt development.

**How to avoid:** Use incremental threshold approach:
- Week 1: Set threshold to current coverage (9%) - ensures no regression
- Week 2-4: Increase by 10% per week until 80% reached
- Or use Codecov's `threshold` setting to allow gradual improvement

**Warning signs:** CI failing on every PR with coverage errors; developers disabling CI.

### Pitfall 2: Missing CODECOV_TOKEN Secret

**What goes wrong:** Coverage uploads fail silently or with authentication errors.

**Why it happens:** Codecov requires token for private repos and some features in public repos.

**How to avoid:**
1. Add `CODECOV_TOKEN` to GitHub repository secrets
2. Test with `codecov/codecov-action@v4` which provides clear error messages
3. Use `fail_ci_if_error: true` to make failures visible

**Warning signs:** Coverage reports not appearing in PR comments; "failed to upload" errors in CI logs.

### Pitfall 3: Excluding Too Much From Coverage

**What goes wrong:** Coverage appears high but critical code paths are untested.

**Why it happens:** Developers exclude "hard to test" code from coverage.

**How to avoid:** Only exclude:
- Test files themselves (`*_test.go`, `*.test.*`)
- Generated code
- Configuration files
- Main entry points (minimal logic)

**Warning signs:** Codecov `ignore` list contains source directories; coverage drops when ignore patterns removed.

### Pitfall 4: Not Running Integration Tests in CI

**What goes wrong:** Integration tests pass locally but fail in CI due to missing service dependencies.

**Why it happens:** CI environment differs from local development environment.

**How to avoid:** Use GitHub Actions `services` for PostgreSQL, TimescaleDB:
```yaml
services:
  postgres:
    image: postgres:16
    env:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: pganalytics_test
    ports:
      - 5432:5432
    options: --health-cmd pg_isready --health-interval 10s
```

**Warning signs:** Integration tests skipped in CI; "database connection refused" errors.

## Code Examples

### Coverage Threshold Check Script (Go)

```bash
# Source: Standard Go tooling pattern
#!/bin/bash
# check-coverage.sh - Fail if coverage below threshold

THRESHOLD=80
COVERAGE_FILE=coverage.out

cd backend

# Generate coverage
go test -coverprofile=$COVERAGE_FILE ./...

# Extract total coverage
COVERAGE=$(go tool cover -func=$COVERAGE_FILE | grep total | awk '{print $3}' | sed 's/%//')

echo "Coverage: ${COVERAGE}%"

# Compare using bc for floating point
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
  echo "::error::Coverage ${COVERAGE}% is below ${THRESHOLD}% threshold"
  exit 1
fi

echo "Coverage check passed"
```

### Vitest Coverage with Thresholds

```typescript
// vite.config.ts - add coverage thresholds
export default defineConfig({
  test: {
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      // Thresholds for enforcement
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 70,
        statements: 80,
        // Fail build if thresholds not met
        '100': false, // Set to true for 100% coverage enforcement
      },
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData.ts',
      ]
    }
  }
})
```

### Enable Unused Linter in golangci.yml

```yaml
# .golangci.yml - add unused to enabled linters
version: 2

run:
  timeout: 5m
  modules-download-mode: readonly
  skip-dirs:
    - tools
    - backend/cmd  # Main entry points

linters:
  enable:
    - govet
    - ineffassign
    - misspell
    - unused  # NEW: Detect unused constants, variables, functions, types

  disable:
    - staticcheck  # Keep disabled for now
    - errcheck     # Keep disabled for now
    # ... rest of disabled linters
```

### GitHub Branch Protection Rule (via API)

```bash
# Set required status checks for main branch
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["ci-passed"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"dismiss_stale_reviews":true,"require_code_owner_reviews":false}'
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Coverage as informational | Coverage as quality gate with thresholds | 2020+ | Prevents coverage regression |
| Manual PR checks | Required status checks in GitHub | 2019+ | Enforces CI before merge |
| Multiple CI workflows for same checks | Unified CI workflow with summary job | 2022+ | Simpler branch protection |
| Istanbul/nyc for JS coverage | Vitest v8 coverage provider | 2023+ | 10-20x faster coverage |

**Deprecated/outdated:**
- `codecov/codecov-action@v1-v2`: Use v4 for better token handling
- `go test -cover` without `-coverprofile`: Cannot generate reports or check thresholds
- Manual coverage badge updates: Use shields.io with Codecov integration

## Open Questions

1. **Should we enforce 80% coverage immediately or incrementally?**
   - What we know: Current backend coverage is 9.1%, far below 80% target
   - What's unclear: Timeline for test expansion, team capacity for writing tests
   - Recommendation: Start with 10% threshold (current + 1%), increase 5-10% per week

2. **Should E2E test coverage be included in the 80% target?**
   - What we know: Playwright tests exist, but coverage collection for E2E is complex
   - What's unclear: Whether E2E coverage is practical to collect
   - Recommendation: Exclude E2E from coverage target; focus on unit/integration tests

3. **What is the current frontend coverage percentage?**
   - What we know: 393/410 tests pass, coverage directory exists
   - What's unclear: Actual percentage (coverage report incomplete due to test failures)
   - Recommendation: Fix failing integration tests first, then measure baseline

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Backend Framework | Go testing + testcontainers-go v0.42.0 |
| Backend Config file | None (standard Go testing) |
| Backend Quick run command | `go test -short ./backend/...` |
| Backend Full suite command | `go test -race -coverprofile=coverage.out ./backend/...` |
| Frontend Framework | Vitest v1.0.0 + Playwright v1.59.1 |
| Frontend Config file | `/Users/glauco.torres/git/pganalytics-v3/frontend/vite.config.ts` |
| Frontend Quick run command | `npm run test -- --run` |
| Frontend Full suite command | `npm run test:coverage -- --run && npm run test:e2e` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUAL-05 | Coverage tracking 80%+ | CI/CD | `go tool cover -func=coverage.out` | CI workflow exists, threshold check needed |
| QUAL-06 | Unused code cleanup | Lint | `golangci-lint run -E unused ./backend/...` | Linter available, not enabled in config |
| TEST-17 | CI/CD test execution | CI/CD | GitHub Actions workflows | Existing in `.github/workflows/` |
| TEST-18 | Failures block deployment | CI/CD | Branch protection rules | Not configured |
| TEST-19 | Coverage reports published | CI/CD | `codecov/codecov-action@v4` | Not configured (no codecov.yml) |
| TEST-20 | Performance tracking | CI/CD | Go `-json`, Vitest verbose reporter | Not configured |

### Sampling Rate

- **Per task commit:** `go test -short ./backend/...` (backend), `npm run test -- --run` (frontend)
- **Per wave merge:** Full suite with coverage: `go test -race -coverprofile=coverage.out ./backend/...` && `npm run test:coverage -- --run`
- **Phase gate:** All CI checks green + coverage >= threshold

### Wave 0 Gaps

- [ ] `.github/workflows/ci.yml` - Unified quality gate workflow
- [ ] `codecov.yml` - Coverage threshold configuration
- [ ] `CODECOV_TOKEN` secret in GitHub repository settings
- [ ] Branch protection rule requiring `ci-passed` status check
- [ ] `.golangci.yml` - Enable `unused` linter
- [ ] `vite.config.ts` - Add coverage thresholds configuration
- [ ] Fix 17 failing frontend integration tests in `components.integration.test.tsx`

## Sources

### Primary (HIGH confidence)

- Project analysis of `.github/workflows/` - Existing CI configuration
- Go tooling documentation - Coverage collection patterns
- Vitest documentation - Coverage provider configuration
- golangci-lint v2.11.4 - Linter configuration verified

### Secondary (MEDIUM confidence)

- GitHub Actions patterns for required status checks
- Codecov documentation for threshold configuration

### Tertiary (LOW confidence)

- None - All findings based on project-specific analysis

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All tools already installed and configured in project
- Architecture: HIGH - Clear patterns from existing workflows
- Pitfalls: MEDIUM - Based on industry patterns, specific to project context
- Coverage baseline: HIGH - Measured directly (9.1% backend)

**Research date:** 2026-04-29
**Valid until:** 30 days (CI/CD tooling stable)