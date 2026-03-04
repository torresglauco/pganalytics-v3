# GitHub Actions Manual Setup Guide

## Overview

Due to GitHub Personal Access Token scope limitations (workflow scope required for direct commits), this guide provides step-by-step instructions to manually create the GitHub Actions workflows through the GitHub UI.

**Note**: The complete workflow YAML code is available in `docs/GITHUB_ACTIONS_SETUP.md` and locally in `.github/workflows/` directory.

---

## Setup Instructions

### Step 1: Navigate to Workflows

1. Go to your GitHub repository: https://github.com/torresglauco/pganalytics-v3
2. Click on the **Actions** tab at the top
3. Click **"New workflow"** or **"Set up a workflow yourself"**

### Step 2: Create E2E Tests Workflow

1. In the Actions tab, click **New workflow**
2. Choose **"Set up a workflow yourself"**
3. Name the file: `e2e-tests.yml`
4. Copy the complete YAML from `.github/workflows/e2e-tests.yml` (141 lines)
5. Click **"Start commit"** → **"Commit new file"**

**What this workflow does:**
- Runs multi-browser E2E tests (Chromium, Firefox, WebKit) on frontend changes
- Starts backend services for testing
- Uploads test reports and videos
- Comments PR with test results
- Retains artifacts for 30 days

### Step 3: Create Frontend Quality Workflow

1. Click **New workflow** → **"Set up a workflow yourself"**
2. Name the file: `frontend-quality.yml`
3. Copy the complete YAML from `.github/workflows/frontend-quality.yml` (167 lines)
4. Click **"Start commit"** → **"Commit new file"**

**What this workflow does:**
- Runs ESLint with 0 violations gate
- Strict TypeScript type checking
- Unit tests with coverage reporting
- Production build verification
- Bundle size monitoring (warns >500KB)

### Step 4: Create Backend Tests Workflow

1. Click **New workflow** → **"Set up a workflow yourself"**
2. Name the file: `backend-tests.yml`
3. Copy the complete YAML from `.github/workflows/backend-tests.yml` (206 lines)
4. Click **"Start commit"** → **"Commit new file"**

**What this workflow does:**
- Runs Go unit and integration tests
- Sets up PostgreSQL and TimescaleDB services
- Uploads coverage to Codecov
- Runs golangci-lint for code quality
- GoSec security scanning
- Builds production binary
- Verifies binary size

### Step 5: Create Security Scanning Workflow

1. Click **New workflow** → **"Set up a workflow yourself"**
2. Name the file: `security.yml`
3. Copy the complete YAML from `.github/workflows/security.yml` (206 lines)
4. Click **"Start commit"** → **"Commit new file"**

**What this workflow does:**
- npm audit with moderate severity threshold
- Snyk vulnerability scanning
- GoSec security scanning
- License compliance checking
- Container image scanning (Trivy)
- Secrets detection (TruffleHog)
- Weekly scheduled scans (Sunday at midnight)

---

## Configure GitHub Secrets (Optional)

Some workflows use GitHub Secrets for sensitive tokens:

### For Snyk Scanning

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **"New repository secret"**
3. Name: `SNYK_TOKEN`
4. Value: Your Snyk API token (from https://app.snyk.io/account/settings/api)
5. Click **"Add secret"**

### For Codecov (Optional)

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **"New repository secret"**
3. Name: `CODECOV_TOKEN`
4. Value: Your Codecov token (from https://codecov.io/account/gh/torresglauco/pganalytics-v3)
5. Click **"Add secret"**

---

## Verify Workflows Are Active

1. Go to the **Actions** tab
2. You should see the 4 workflows listed:
   - ✅ E2E Tests
   - ✅ Frontend Code Quality
   - ✅ Backend Tests
   - ✅ Security Scanning

3. For each workflow, you should see:
   - **Status**: Active (green checkmark)
   - **Triggers**: Configured for push/pull_request

---

## Test Workflows

### Trigger E2E Tests
```bash
# Make a frontend change and push
git checkout -b test/workflows
echo "// test" >> frontend/src/App.tsx
git add frontend/src/App.tsx
git commit -m "test: trigger workflows"
git push origin test/workflows
# Go to GitHub → Pull Requests → Create PR
# Watch Actions tab for workflow execution
```

### Trigger Backend Tests
```bash
# Make a backend change and push
echo "// test" >> backend/main.go
git add backend/main.go
git commit -m "test: trigger backend workflow"
git push origin test/workflows
```

---

## Monitor Workflow Execution

### View Workflow Runs
1. Go to **Actions** tab
2. Click on the workflow name (e.g., "E2E Tests")
3. See the list of recent runs with status

### View Workflow Logs
1. Click on a specific run
2. Click on a job name (e.g., "Run E2E Tests")
3. Expand steps to see detailed logs

### View Workflow Artifacts
1. Click on a specific run
2. Scroll down to "Artifacts" section
3. Download test reports, coverage reports, etc.

---

## Add Status Badges to README

To display workflow status in your README, add these badges:

```markdown
## CI/CD Status

[![E2E Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml/badge.svg)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/e2e-tests.yml)
[![Frontend Quality](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml/badge.svg)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/frontend-quality.yml)
[![Backend Tests](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml/badge.svg)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/backend-tests.yml)
[![Security Scanning](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml/badge.svg)](https://github.com/torresglauco/pganalytics-v3/actions/workflows/security.yml)
```

---

## Troubleshooting

### Workflow Not Triggering

**Problem**: Workflow doesn't run on push/PR

**Solution**:
1. Check workflow file is in `.github/workflows/` directory
2. Verify branch name matches (e.g., `main` vs `master`)
3. Check file paths filter (some workflows only trigger on specific paths)
4. Go to **Actions** → workflow name → **Edit**

### Workflow Fails on First Run

**Problem**: Tests fail because services aren't ready

**Solution**:
- Workflows have built-in retry logic and wait times
- First run may take longer (dependencies installing)
- Check logs for actual error messages
- Common issues:
  - Port 5432 (PostgreSQL) already in use locally
  - Port 3000 (Frontend) already in use locally
  - Docker not available in GitHub Actions runner
  - Node.js cache issues

### Artifact Not Available

**Problem**: "Artifacts not found" when trying to download

**Solution**:
- Artifacts expire after retention-days (default: 30 days)
- Only successful steps upload artifacts
- Check if previous jobs failed
- View logs to see what happened

### Secret Not Working

**Problem**: Snyk or other secrets fail

**Solution**:
1. Verify secret is configured in Settings
2. Verify secret name is spelled correctly
3. Check it's in correct repository (not organization level)
4. Test token is valid:
   - Snyk: https://app.snyk.io/account/settings/api
   - Codecov: https://codecov.io/account

---

## Local Testing with Act

To test workflows locally before pushing, use the `act` tool:

```bash
# Install act (macOS)
brew install act

# List workflows
act -l

# Run specific workflow
act -j e2e-tests

# Run with specific event
act pull_request

# View workflow logs
act -v
```

---

## Next Steps

1. **Create workflows** using the manual setup steps above
2. **Configure secrets** if using Snyk or Codecov
3. **Test workflows** by creating a test PR or pushing changes
4. **Monitor runs** in the Actions tab
5. **Review logs** if any steps fail
6. **Add badges** to README for status visibility

---

## Complete Workflow YAML

All workflow files are available in the repository:

- `docs/GITHUB_ACTIONS_SETUP.md` - Complete documentation with all YAML code
- `.github/workflows/e2e-tests.yml` - E2E testing workflow (141 lines)
- `.github/workflows/frontend-quality.yml` - Frontend quality checks (167 lines)
- `.github/workflows/backend-tests.yml` - Backend testing (206 lines)
- `.github/workflows/security.yml` - Security scanning (206 lines)

Copy the YAML directly from these files into the GitHub UI when creating workflows.

---

## Support

For issues with workflow setup:

1. Check the logs in the Actions tab
2. Review the documentation in `docs/GITHUB_ACTIONS_SETUP.md`
3. See `docs/FAQ_AND_TROUBLESHOOTING.md` for common issues
4. Open a GitHub issue with workflow logs attached

---

**Last Updated**: March 29, 2026
**Version**: 1.0
**Status**: Complete and Ready
