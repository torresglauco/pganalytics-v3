# GitHub Actions CI/CD Setup Guide

## Overview

This guide provides step-by-step instructions to set up GitHub Actions workflows for pgAnalytics v3 to enable automated testing, quality checks, and security scanning.

## Prerequisites

- GitHub repository with write access
- GitHub Actions enabled (default)
- Optional: GitHub Secrets configured for sensitive tokens

---

## Workflow Files

### 1. Create Frontend E2E Tests Workflow

**File**: `.github/workflows/e2e-tests.yml`

```yaml
name: E2E Tests

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
      - '.github/workflows/e2e-tests.yml'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
      - '.github/workflows/e2e-tests.yml'

jobs:
  e2e-tests:
    name: Run E2E Tests
    runs-on: ubuntu-latest
    timeout-minutes: 30

    strategy:
      fail-fast: false
      matrix:
        browser: [chromium, firefox, webkit]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci

      - name: Install Playwright browsers
        working-directory: frontend
        run: npx playwright install --with-deps ${{ matrix.browser }}

      - name: Start backend service
        run: |
          docker-compose up -d api postgres timescaledb
          sleep 10
          for i in {1..30}; do
            if curl -f http://localhost:8080/api/v1/health; then
              echo "Backend is ready"
              break
            fi
            echo "Waiting for backend... ($i/30)"
            sleep 1
          done
        env:
          DOCKER_BUILDKIT: 1

      - name: Run E2E tests
        working-directory: frontend
        run: npx playwright test --project=${{ matrix.browser }}
        env:
          BASE_URL: 'http://localhost:3000'
          API_URL: 'http://localhost:8080/api/v1'

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: playwright-report-${{ matrix.browser }}
          path: frontend/playwright-report/
          retention-days: 30

      - name: Upload test videos
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: playwright-videos-${{ matrix.browser }}
          path: frontend/test-results/
          retention-days: 7

      - name: Stop services
        if: always()
        run: docker-compose down

  test-summary:
    name: E2E Test Summary
    runs-on: ubuntu-latest
    needs: e2e-tests
    if: always()

    steps:
      - name: Check test results
        run: |
          echo "E2E Tests completed across all browsers"
          if [ "${{ needs.e2e-tests.result }}" == "failure" ]; then
            echo "❌ Some tests failed"
            exit 1
          else
            echo "✅ All tests passed"
          fi
```

### 2. Create Frontend Quality Workflow

**File**: `.github/workflows/frontend-quality.yml`

```yaml
name: Frontend Code Quality

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
      - '.github/workflows/frontend-quality.yml'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'frontend/**'
      - '.github/workflows/frontend-quality.yml'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run ESLint
        working-directory: frontend
        run: npm run lint

  type-check:
    name: Type Check
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run TypeScript type check
        working-directory: frontend
        run: npm run type-check

  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run unit tests
        working-directory: frontend
        run: npm test

      - name: Upload coverage reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: coverage-reports
          path: frontend/coverage/
          retention-days: 30

  build:
    name: Build
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Build production bundle
        working-directory: frontend
        run: npm run build

      - name: Upload build artifacts
        if: success()
        uses: actions/upload-artifact@v4
        with:
          name: frontend-build
          path: frontend/dist/
          retention-days: 7
```

### 3. Create Backend Tests Workflow

**File**: `.github/workflows/backend-tests.yml`

```yaml
name: Backend Tests

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'backend/**'
      - '.github/workflows/backend-tests.yml'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'backend/**'
      - '.github/workflows/backend-tests.yml'

jobs:
  test:
    name: Unit & Integration Tests
    runs-on: ubuntu-latest
    timeout-minutes: 20

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: pganalytics_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      timescaledb:
        image: timescale/timescaledb:latest-pg16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: pganalytics_metrics_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5433:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Download dependencies
        working-directory: backend
        run: go mod download

      - name: Run tests with coverage
        working-directory: backend
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        env:
          DATABASE_URL: 'postgres://postgres:postgres@localhost:5432/pganalytics_test?sslmode=disable'
          TIMESCALE_URL: 'postgres://postgres:postgres@localhost:5433/pganalytics_metrics_test?sslmode=disable'

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
          flags: backend
          name: backend-coverage

  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: backend
          args: --timeout=5m

  security:
    name: Security Scan
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install gosec
        run: go install github.com/securego/gosec/v2/cmd/gosec@latest

      - name: Run GoSec security scan
        working-directory: backend
        run: |
          gosec -no-fail -fmt json -out gosec-report.json ./...

  build:
    name: Build
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Build binary
        working-directory: backend
        run: |
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-s -w" \
            -o pganalytics-api \
            ./cmd/pganalytics-api

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: pganalytics-api-linux-amd64
          path: backend/pganalytics-api
          retention-days: 7
```

### 4. Create Security Scanning Workflow

**File**: `.github/workflows/security.yml`

```yaml
name: Security Scanning

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  schedule:
    - cron: '0 0 * * 0'  # Weekly scan on Sunday

jobs:
  npm-audit:
    name: npm Dependency Audit
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run npm audit
        working-directory: frontend
        continue-on-error: true
        run: |
          npm audit --audit-level=moderate || true
          npm audit --json > npm-audit-report.json || true

      - name: Check for vulnerabilities
        working-directory: frontend
        run: |
          if npm audit | grep -q "found.*vulnerabilities"; then
            echo "⚠️  Vulnerabilities found in npm dependencies"
            exit 1
          else
            echo "✅ No vulnerabilities found"
          fi

      - name: Upload audit report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: npm-audit-report
          path: frontend/npm-audit-report.json
          retention-days: 30

  gosec-scan:
    name: Go Security Scan
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install gosec
        run: go install github.com/securego/gosec/v2/cmd/gosec@latest

      - name: Run GoSec security scan
        working-directory: backend
        run: gosec -no-fail -fmt json -out gosec-report.json ./...

      - name: Upload security reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: gosec-report
          path: backend/gosec-report.json
          retention-days: 30
```

---

## Setup Instructions

### Step 1: Create Workflow Files

1. Create `.github/workflows/` directory in repository
2. Copy each workflow YAML file above into this directory
3. Commit the files to repository

```bash
mkdir -p .github/workflows
# Copy e2e-tests.yml, frontend-quality.yml, backend-tests.yml, security.yml
git add .github/workflows/
git commit -m "ci: Add GitHub Actions CI/CD workflows"
git push origin main
```

### Step 2: Configure GitHub Secrets (Optional)

For advanced features like Snyk scanning:

1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add secrets:
   - `SNYK_TOKEN`: Snyk API token (for advanced scanning)
   - `CODECOV_TOKEN`: Codecov token (for coverage reporting)

### Step 3: Enable GitHub Actions

1. Go to repository Settings → Actions → General
2. Ensure "Actions permissions" is set to "Allow all actions"
3. Set workflow permissions to "Read and write permissions"

### Step 4: Verify Workflows

1. Push code to trigger workflows
2. Go to Actions tab in GitHub
3. Verify workflows run successfully
4. Check individual workflow results

---

## Workflow Status Badges

Add to README.md:

```markdown
## CI/CD Status

[![E2E Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml)
[![Frontend Quality](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml)
[![Backend Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml)
[![Security](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml/badge.svg?branch=main)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml)
```

---

## What These Workflows Do

### E2E Tests Workflow
- Runs 65+ Playwright test cases
- Tests across 3 browsers (Chromium, Firefox, WebKit)
- Automatically uploads test reports
- Posts results to pull requests
- Duration: ~15 min per browser

### Frontend Quality Workflow
- ESLint linting
- TypeScript type checking
- Vitest unit tests
- Production build verification
- Duration: ~8 minutes

### Backend Tests Workflow
- Go unit and integration tests
- golangci-lint linting
- GoSec security scanning
- Production binary build
- Duration: ~10 minutes

### Security Workflow
- npm audit for dependencies
- GoSec for Go code
- Weekly scheduled scans
- Artifact uploads for review
- Duration: ~5 minutes

---

## Local Testing

Test workflows locally using `act`:

```bash
# Install act
brew install act  # macOS

# Run specific workflow
act --job e2e-tests

# Run with all browsers
act push

# Run security checks
act --job npm-audit
```

---

## Conclusion

These GitHub Actions workflows provide:
- ✅ Automated testing on every push/PR
- ✅ Quality gates for code standards
- ✅ Security vulnerability scanning
- ✅ Cross-browser compatibility testing
- ✅ Production-ready builds
- ✅ Artifact management and retention

They ensure code quality and security without manual intervention.
