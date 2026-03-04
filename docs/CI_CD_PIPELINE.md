# CI/CD Pipeline Documentation

## Overview

pgAnalytics v3 uses GitHub Actions for continuous integration and deployment, providing automated testing, quality checks, and security scanning on every push and pull request.

## Table of Contents

1. [Workflows](#workflows)
2. [Event Triggers](#event-triggers)
3. [Quality Checks](#quality-checks)
4. [Testing Strategy](#testing-strategy)
5. [Security Scanning](#security-scanning)
6. [Artifact Management](#artifact-management)
7. [Status Badges](#status-badges)
8. [Local Development](#local-development)

---

## Workflows

### 1. E2E Tests Workflow (`e2e-tests.yml`)

**Purpose**: Run end-to-end tests across multiple browsers
**Triggered**: On push/PR to main/develop branches affecting frontend/
**Duration**: ~15 minutes per browser

**Jobs**:
- **e2e-tests**: Multi-browser test execution
  - Matrix: Chromium, Firefox, WebKit
  - Runs 65+ test cases
  - Tests all major features and user flows

**Test Coverage**:
- Authentication (login/logout)
- Collector management (registration, editing)
- Dashboard visualization
- Alert management
- User management
- Permissions and access control

**Artifacts**:
- Playwright HTML reports (per browser)
- Test videos (on failure)
- Test results JSON (for parsing)

**Steps**:
```
1. Checkout code
2. Setup Node.js (v18)
3. Install frontend dependencies (npm ci)
4. Install Playwright browsers
5. Start Docker services (API, PostgreSQL, TimescaleDB)
6. Wait for backend health check
7. Run E2E tests with Playwright
8. Upload reports and videos
9. Post results to PR (if applicable)
10. Cleanup Docker services
```

### 2. Frontend Quality Workflow (`frontend-quality.yml`)

**Purpose**: Lint, type-check, unit test, and build frontend
**Triggered**: On push/PR to main/develop branches affecting frontend/
**Duration**: ~8 minutes

**Jobs**:
- **lint**: ESLint code style check
  - Runs: `npm run lint`
  - Checks TypeScript and JSX syntax
  - Reports style violations

- **type-check**: TypeScript type validation
  - Runs: `npm run type-check`
  - Validates all type annotations
  - Reports type errors

- **unit-tests**: Jest unit tests
  - Runs: `npm test`
  - Tests components in isolation
  - Uploads coverage reports
  - Coverage threshold: 70%

- **build**: Production build verification
  - Runs: `npm run build`
  - Creates optimized bundle
  - Checks bundle size (warns if >500KB)
  - Uploads build artifacts

**Artifacts**:
- Coverage reports
- Frontend build (dist/)

### 3. Backend Tests Workflow (`backend-tests.yml`)

**Purpose**: Test, lint, and build backend API
**Triggered**: On push/PR to main/develop branches affecting backend/
**Duration**: ~10 minutes

**Jobs**:
- **test**: Unit and integration tests
  - Services: PostgreSQL, TimescaleDB
  - Runs: `go test -v -race -coverprofile=coverage.out ./...`
  - Coverage reporting to Codecov
  - Threshold: 70%

- **lint**: Go linting
  - Tool: golangci-lint
  - Checks code style and quality
  - Timeout: 5 minutes

- **security**: GoSec security scan
  - Tool: gosec
  - Generates JSON and SARIF reports
  - Uploads to GitHub Security tab

- **build**: Build verification
  - Builds optimized Linux binary
  - Checks binary size
  - Uploads build artifact

**Artifacts**:
- Test coverage reports
- Security reports (SARIF)
- Compiled binary (pganalytics-api)

### 4. Security Scanning Workflow (`security.yml`)

**Purpose**: Comprehensive security vulnerability scanning
**Triggered**: On push/PR, and weekly schedule (Sunday 00:00 UTC)
**Duration**: ~15 minutes

**Jobs**:
- **npm-audit**: Frontend dependency vulnerabilities
  - Tool: npm audit
  - Severity: Moderate and above
  - Reports vulnerable packages

- **snyk-scan**: Advanced vulnerability scanning
  - Tool: Snyk (requires token)
  - Scans npm dependencies
  - Severity: High and above
  - Provides fix recommendations

- **gosec-scan**: Go security issues
  - Tool: gosec
  - Scans backend code
  - Reports security patterns

- **dependency-check**: License compliance
  - Tool: license-checker
  - Checks for problematic licenses
  - Generates license reports

- **container-scan**: Docker image scanning
  - Tool: Trivy
  - Scans built container images
  - Reports CRITICAL/HIGH vulnerabilities

- **secrets-scan**: Credential detection
  - Tool: TruffleHog
  - Scans for accidentally committed secrets
  - Prevents credential leaks

**Artifacts**:
- npm audit reports
- License reports
- Security scan results (SARIF)

---

## Event Triggers

### On Push to Main/Develop
All workflows run automatically when code is pushed to main or develop branches.

### On Pull Requests
All workflows run on PRs to ensure quality before merge.

### Path-based Triggering
Workflows only run if relevant files changed:
- **e2e-tests**: `frontend/**`, `.github/workflows/e2e-tests.yml`
- **frontend-quality**: `frontend/**`, `.github/workflows/frontend-quality.yml`
- **backend-tests**: `backend/**`, `.github/workflows/backend-tests.yml`

### Scheduled Runs
Security scanning runs on a schedule:
- **Weekly**: Every Sunday at 00:00 UTC
- Catches newly discovered vulnerabilities

---

## Quality Checks

### Frontend Quality Gates

| Check | Tool | Threshold | Action |
|-------|------|-----------|--------|
| Linting | ESLint | 0 warnings | ❌ Fail if violations |
| Type Checking | TypeScript | Strict | ❌ Fail if errors |
| Unit Tests | Vitest | 70% coverage | ⚠️ Warn if below |
| Build | Vite | <500KB | ⚠️ Warn if larger |

### Backend Quality Gates

| Check | Tool | Threshold | Action |
|-------|------|-----------|--------|
| Unit Tests | Go test | 70% coverage | ⚠️ Warn if below |
| Linting | golangci-lint | 0 issues | ❌ Fail if violations |
| Security | gosec | 0 high/critical | ⚠️ Log if found |
| Build | go build | Success | ❌ Fail if errors |

### Security Gates

| Check | Tool | Threshold | Action |
|-------|------|-----------|--------|
| npm Audit | npm audit | No moderate+ | ❌ Fail if found |
| Secrets | TruffleHog | 0 secrets | ❌ Fail if found |
| Container | Trivy | No critical+ | ⚠️ Report if found |

---

## Testing Strategy

### E2E Test Execution

**Multi-Browser Testing**:
- Chromium (default Chrome/Edge engine)
- Firefox (Mozilla engine)
- WebKit (Safari engine)
- Ensures cross-browser compatibility

**Test Environment**:
```
Backend: http://localhost:8080
Frontend: http://localhost:3000
Database: PostgreSQL + TimescaleDB
```

**Test Organization**:
```
frontend/e2e/
├── playwright.config.ts         # Configuration
├── pages/                        # Page Object Models
│   ├── LoginPage.ts
│   ├── DashboardPage.ts
│   ├── CollectorPage.ts
│   ├── AlertsPage.ts
│   └── UsersPage.ts
└── tests/
    ├── 01-login-logout.spec.ts
    ├── 02-collector-registration.spec.ts
    ├── 03-dashboard.spec.ts
    ├── 04-alert-management.spec.ts
    ├── 05-user-management.spec.ts
    └── 06-permissions-access-control.spec.ts
```

**Running E2E Tests Locally**:
```bash
# Install dependencies
cd frontend
npm install
npx playwright install

# Run all tests
npx playwright test

# Run specific browser
npx playwright test --project=chromium

# Debug mode
npx playwright test --debug

# Watch mode
npx playwright test --watch

# UI mode
npx playwright test --ui
```

### Unit Test Execution

**Frontend Unit Tests**:
```bash
cd frontend
npm test                  # Run all tests
npm test -- --coverage   # With coverage
npm test -- --watch      # Watch mode
```

**Backend Unit Tests**:
```bash
cd backend
go test ./...                              # Run all tests
go test -v ./...                           # Verbose output
go test -cover ./...                       # Show coverage
go test -coverprofile=coverage.out ./...   # Coverage file
go tool cover -html=coverage.out           # HTML report
```

---

## Security Scanning

### Vulnerability Severity Levels

| Level | Definition | Action |
|-------|-----------|--------|
| CRITICAL | Immediate risk | Fail build |
| HIGH | Significant risk | Fail or warn |
| MEDIUM | Important risk | Warn |
| LOW | Minor risk | Log |

### npm Audit

Scans frontend dependencies for known vulnerabilities.

**Config**:
- Audit level: Moderate+
- Auto-fix: Available but reviewed manually
- Schedule: Every PR and weekly

**Viewing Results**:
```bash
cd frontend
npm audit               # Text report
npm audit fix           # Apply fixes (review first!)
npm audit fix --dry-run # Preview fixes
```

**Fixing Vulnerabilities**:
1. Review npm audit report
2. Check fix recommendations
3. Test changes thoroughly
4. Commit fixes with explanation

### GoSec

Scans backend Go code for security patterns.

**Checks**:
- SQL injection patterns
- Hardcoded credentials
- Weak cryptography
- Insecure random generation
- CWE mappings

**Config**:
- Excludes: Test files, examples
- Output: JSON and SARIF formats
- GitHub integration: Automatic upload to Security tab

### Container Image Scanning (Trivy)

Scans Docker images for OS package vulnerabilities.

**Scope**:
- Base image vulnerabilities
- Installed package versions
- Severity: CRITICAL, HIGH

**Handling Results**:
1. Update base image if needed
2. Patch vulnerable packages
3. Rebuild and rescan
4. Document accepted risks

### Secrets Detection (TruffleHog)

Prevents accidentally committing secrets.

**Detects**:
- API keys and tokens
- AWS credentials
- Private keys
- Database passwords
- OAuth tokens

**Actions on Detection**:
1. Reject commit
2. Review code for secrets
3. Rotate credentials
4. Remove from commit history
5. Use `.gitignore` for secret files

---

## Artifact Management

### Retention Policies

| Artifact | Retention | Usage |
|----------|-----------|-------|
| Test reports | 30 days | Analysis |
| Test videos | 7 days | Debug failures |
| Coverage reports | 30 days | Tracking trends |
| Build artifacts | 7 days | Download if needed |
| Security reports | 30 days | Compliance |

### Accessing Artifacts

**From GitHub Actions UI**:
1. Go to Actions tab
2. Select workflow run
3. Scroll to Artifacts section
4. Click to download

**From CLI**:
```bash
# List artifacts
gh run list --workflow e2e-tests.yml

# Download artifact
gh run download <run-id> -n playwright-report-chromium
```

### Test Report Viewing

**Playwright Reports**:
```bash
# Extract and view
unzip playwright-report-chromium.zip
npx playwright show-report playwright-report
```

---

## Status Badges

Add these badges to README.md for status visibility:

```markdown
## Status

[![E2E Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml)
[![Frontend Quality](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml)
[![Backend Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml)
[![Security](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml)
```

---

## Local Development

### Running Full CI Locally

Use `act` tool to run GitHub Actions locally:

```bash
# Install act (macOS)
brew install act

# Run specific workflow
act --job e2e-tests

# Run with specific event
act push

# Run with secrets
act --secret SNYK_TOKEN=<token>
```

### Pre-commit Checks

Create `.git/hooks/pre-commit` to run checks before committing:

```bash
#!/bin/bash

# Frontend checks
cd frontend
npm run lint || exit 1
npm run type-check || exit 1
npm test -- --coverage || exit 1

# Backend checks
cd ../backend
go fmt ./...
golangci-lint run ./... || exit 1
go test ./... || exit 1

echo "✅ All pre-commit checks passed"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

### Pre-push Checks

Create `.git/hooks/pre-push` to ensure quality before pushing:

```bash
#!/bin/bash

echo "Running pre-push checks..."

# Backend
cd backend
go test -race ./... || exit 1
golangci-lint run ./... || exit 1

# Frontend
cd ../frontend
npm run type-check || exit 1
npm run lint || exit 1
npm test || exit 1
npm run build || exit 1

echo "✅ Ready to push"
```

Make executable:
```bash
chmod +x .git/hooks/pre-push
```

---

## Troubleshooting

### E2E Tests Fail Randomly

**Causes**:
- Flaky network conditions
- Backend not ready
- Race conditions in tests

**Solutions**:
1. Increase timeouts in `playwright.config.ts`
2. Add explicit waits for backend health
3. Check test logs for specific errors
4. Rerun failed tests

### Coverage Below Threshold

**Causes**:
- New code without tests
- Incomplete test coverage

**Solutions**:
1. Add tests for new code
2. Update coverage threshold if justified
3. Review untested lines: `go tool cover -html=coverage.out`

### Security Scan False Positives

**Handling**:
1. Review each finding carefully
2. Document false positives in `SECURITY.md`
3. Configure tools to ignore (if justified)
4. Add suppressions with comments

### GitHub Actions Timeouts

**Default**: 20-30 minutes per job

**Solutions**:
1. Optimize test execution
2. Parallelize tests across workers
3. Cache dependencies
4. Remove unnecessary checks

---

## Best Practices

### Writing Tests for CI

1. **Deterministic**: Tests always pass/fail consistently
2. **Independent**: No test depends on another
3. **Isolated**: Use test databases, mocks
4. **Fast**: Complete in reasonable time
5. **Descriptive**: Clear failure messages

### Optimizing CI/CD

1. **Cache aggressively**: Dependencies, build artifacts
2. **Parallelize**: Run independent jobs in parallel
3. **Fail fast**: Stop on first critical failure
4. **Clean artifacts**: Remove old reports regularly
5. **Monitor costs**: Keep eye on GitHub Actions minutes

### Security Best Practices

1. **Never commit secrets**: Use GitHub Secrets for sensitive data
2. **Review dependencies**: Check licenses and vulnerabilities
3. **Keep images updated**: Regularly update base images
4. **Scan containers**: Before deploying to production
5. **Monitor compliance**: Track security metrics over time

---

## Conclusion

The CI/CD pipeline ensures:
- ✅ Code quality on every commit
- ✅ Test coverage maintenance
- ✅ Security vulnerability detection
- ✅ Cross-browser compatibility
- ✅ Production-ready builds

Run CI/CD checks locally before pushing to catch issues early and keep the main branch stable.
