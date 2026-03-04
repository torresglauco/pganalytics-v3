# Frontend Security Scanning & npm Audit Guide

## Overview

This guide documents security scanning procedures for the pgAnalytics frontend, including npm audit configuration, vulnerability management, and dependency security best practices.

## Table of Contents

1. [npm Audit Configuration](#npm-audit-configuration)
2. [Vulnerability Management](#vulnerability-management)
3. [Dependency Maintenance](#dependency-maintenance)
4. [Security Scanning Tools](#security-scanning-tools)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

---

## npm Audit Configuration

### Package Configuration

The `frontend/package.json` includes audit configuration:

```json
{
  "name": "pganalytics-ui",
  "version": "3.3.0",
  "scripts": {
    "audit": "npm audit --audit-level=moderate",
    "audit:fix": "npm audit fix",
    "audit:fix:dry-run": "npm audit fix --dry-run",
    "audit:json": "npm audit --json > audit-report.json",
    "security:check": "npm run audit && npm run type-check && npm run lint"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0"
  }
}
```

### Audit Levels

- **low**: Minor issues, low impact
- **moderate**: Some functionality may be affected
- **high**: Significant security risk
- **critical**: Immediate security vulnerability

**Current Policy**: Fail on moderate+ severity

### Running npm Audit

```bash
# Check for vulnerabilities
npm audit

# Check with specific severity
npm audit --audit-level=high

# Generate JSON report
npm audit --json > audit-report.json

# Generate human-readable report with fixes
npm audit --production  # Only production dependencies
```

---

## Vulnerability Management

### Vulnerability Severity Matrix

| Severity | CVSS | Impact | Action |
|----------|------|--------|--------|
| CRITICAL | 9.0-10.0 | Immediate risk | Fix immediately or replace |
| HIGH | 7.0-8.9 | Significant risk | Fix in next release |
| MODERATE | 4.0-6.9 | Some functionality affected | Schedule fix |
| LOW | 0.1-3.9 | Minor issue | Document and track |

### Handling Vulnerabilities

#### Step 1: Assess the Risk

```bash
# Get detailed vulnerability info
npm audit --json | jq '.vulnerabilities'

# Check affected packages
npm list [vulnerable-package]

# Check if vulnerability affects production code
npm audit --production
```

#### Step 2: Update Dependencies

```bash
# Review fixes before applying
npm audit fix --dry-run

# Apply automatic fixes
npm audit fix

# Manual update if needed
npm update [vulnerable-package] --save
```

#### Step 3: Verify Fixes

```bash
# Confirm vulnerabilities are resolved
npm audit

# Run tests to ensure no breakage
npm test
npm run build

# Type check
npm run type-check
```

#### Step 4: Commit Changes

```bash
git add package.json package-lock.json
git commit -m "security: Update dependencies to fix audit vulnerabilities

Fixes:
- [CVE-XXXX] - [Vulnerability Description]

npm audit: 0 vulnerabilities found"

git push origin [feature-branch]
```

### Handling Unfixable Vulnerabilities

If vulnerability cannot be fixed:

1. **Document in `SECURITY.md`**:
```markdown
## Known Vulnerabilities

### CVE-XXXX: [Description]
- Affected Package: [package-name]@[version]
- Severity: [CRITICAL/HIGH]
- Status: ACCEPTED
- Reason: [Reason for acceptance]
- Mitigation: [Mitigation strategy]
- Fix Expected: [Expected fix date or PR link]
```

2. **Configure npm to ignore** (use cautiously):
```bash
# Create .npmrc
echo "audit.allow-list=CVE-XXXX" >> .npmrc
```

3. **Use dependency override** (npm 8.3+):
```json
{
  "overrides": {
    "vulnerable-package": {
      "sub-dependency": ">=1.0.0"
    }
  }
}
```

---

## Dependency Maintenance

### Regular Updates

**Schedule**:
- Weekly: Check for security updates
- Biweekly: Update minor versions
- Monthly: Update major versions (with thorough testing)

**Update Procedure**:

```bash
# Check outdated packages
npm outdated

# Update patch/minor versions safely
npm update

# Update specific package
npm install [package]@latest

# Update to latest major version (breaking changes possible)
npm install [package]@latest --save
```

### Dependency Pinning Strategy

**Current Strategy**:
- Use caret ranges (`^`) in package.json
- Lock exact versions with package-lock.json
- Never commit node_modules/

**Example**:
```json
{
  "dependencies": {
    "react": "^18.2.0",       // Allows 18.2.0 - 18.9.9
    "typescript": "^5.3.3",    // Allows 5.3.3 - 5.9.9
    "vite": "^5.0.8"          // Allows 5.0.8 - 5.9.9
  }
}
```

### Production vs Development Dependencies

**Production** (`--save`):
- React, React Router, Axios
- UI libraries (Headless UI, Lucide)
- Form handling (React Hook Form)
- Validation (Zod)
- State management (Zustand)

**Development** (`--save-dev`):
- TypeScript, ESLint, Prettier
- Testing (Vitest, Testing Library)
- Build tools (Vite, Tailwind)
- Type definitions (@types/*)

```bash
# Add to production
npm install [package] --save

# Add to development
npm install [package] --save-dev

# Remove from both
npm uninstall [package]
```

---

## Security Scanning Tools

### npm Audit

**Built-in npm vulnerability scanner**

```bash
# Basic scan
npm audit

# Continuous scan
npm audit --production

# Fix known vulnerabilities
npm audit fix

# Review before fixing
npm audit fix --dry-run

# Generate reports
npm audit --json > audit.json
npm audit --csv > audit.csv
```

**Configuration**:
```json
{
  "scripts": {
    "audit": "npm audit --audit-level=moderate"
  }
}
```

### Snyk (Recommended for CI/CD)

**Advanced vulnerability scanning**

**Setup**:
1. Sign up at https://snyk.io
2. Get API token from account settings
3. Add to GitHub Secrets: `SNYK_TOKEN`

**Installation**:
```bash
npm install -g snyk
snyk auth [API_TOKEN]
```

**Usage**:
```bash
# Scan project
snyk test

# Fix vulnerabilities
snyk fix

# Monitor for new vulnerabilities
snyk monitor

# Test in CI with high severity
snyk test --severity-threshold=high
```

**GitHub Integration**:
```yaml
- name: Run Snyk scan
  uses: snyk/actions/node@master
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
  with:
    args: --severity-threshold=high
```

### WhiteSource/Mend

**License and security scanning**

```bash
npm install -g whitesource
whitesource-fs-agent --config whitesource-fs-agent.config.json
```

### OWASP Dependency Check

**Open-source vulnerability scanner**

```bash
# Installation
npm install -g owasp-dependency-check

# Usage
dependency-check --project "pgAnalytics Frontend" \
                  --scan frontend/node_modules \
                  --format JSON \
                  --out results.json
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Frontend Security

on:
  push:
    branches: [main, develop]
    paths: ['frontend/**']
  pull_request:
    branches: [main, develop]
    paths: ['frontend/**']
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  npm-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Run npm audit
        working-directory: frontend
        run: npm audit --audit-level=moderate

      - name: Upload audit report
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: npm-audit-report
          path: audit-report.json
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running security checks..."

cd frontend

# Run npm audit
if ! npm audit --audit-level=high > /dev/null 2>&1; then
  echo "❌ npm audit failed"
  npm audit
  exit 1
fi

echo "✅ Security checks passed"
exit 0
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Best Practices

### 1. Regular Audits

- Run weekly: `npm audit`
- Schedule automated scans in CI/CD
- Review reports carefully
- Document decisions for vulnerabilities

### 2. Keep Dependencies Updated

```bash
# Check what's outdated
npm outdated

# Update safely (patch + minor)
npm update

# Full update report
npm outdated --json > outdated.json
```

### 3. Minimize Attack Surface

- Use only necessary dependencies
- Remove unused packages: `npm prune --production`
- Prefer smaller libraries
- Check package downloads and maintenance status

### 4. Review New Dependencies

Before adding a new package:

```bash
# Check package info
npm info [package]

# Check security record
npm audit --dry-run [package]

# Check size impact
npm info [package] | grep dist

# Check maintenance status
npm view [package] time
```

### 5. Supply Chain Security

- Use npm lockfiles (`package-lock.json`)
- Commit lockfiles to git
- Verify package authenticity: `npm verify`
- Use GitHub dependency scanning
- Monitor for typosquatting

```bash
# Verify packages are authentic
npm verify

# Check for suspicious activity
npm audit signatures
```

### 6. Documentation

Maintain `SECURITY.md` with:
- Known vulnerabilities
- Disclosure policy
- Fix procedures
- Contact information

---

## Troubleshooting

### npm audit Reports False Positives

**Solution**:
1. Check if vulnerability actually affects your code
2. Verify dependency chain: `npm list [package]`
3. Document false positive in SECURITY.md
4. Configure npm to ignore if necessary (last resort)

### Dependency Conflict

```bash
# Find conflicts
npm ls [package]

# Use npm audit fix with force
npm audit fix --force  # Use with caution!

# Or manually update conflicting package
npm install [package]@[version] --save
```

### Broken Dependency Updates

If update breaks functionality:

```bash
# Revert to previous version
npm install [package]@[previous-version] --save

# Create issue for compatibility problem
# Work with package maintainers on fix
```

### CI/CD Failures

**If CI/CD fails on npm audit**:

1. Check specific vulnerabilities:
```bash
npm audit --json | jq '.vulnerabilities'
```

2. Review severity and impact
3. Apply fixes: `npm audit fix`
4. Test thoroughly
5. Commit and push again

---

## Security Reporting

### Reporting Vulnerabilities

Found a security issue in pgAnalytics?

**Do NOT create public issues for security vulnerabilities.**

Email security concerns to: `security@pganalytics.local`

Include:
- Description of vulnerability
- Affected versions
- Proof of concept (if possible)
- Recommended fix

### Receiving Notifications

**GitHub Alerts**:
1. Go to repository Settings
2. Security & analysis → Enable Dependabot
3. Receive alerts for vulnerable dependencies

**npm Email Notifications**:
1. Go to https://www.npmjs.com/settings/account
2. Enable email notifications for audit advisories

---

## Reporting & Metrics

### Generate Security Reports

```bash
# npm audit report
npm audit --json > npm-audit-report.json

# Outdated packages
npm outdated --json > npm-outdated-report.json

# License compliance
npx license-checker --json > license-report.json
```

### Track Security Metrics

Maintain metrics over time:
- Number of vulnerabilities by severity
- Mean time to patch (MTTP)
- Dependency update frequency
- License compliance percentage

```json
{
  "scan_date": "2026-03-24",
  "critical": 0,
  "high": 0,
  "moderate": 2,
  "low": 5,
  "total_dependencies": 142,
  "outdated": 8,
  "compliance": "98%"
}
```

---

## Conclusion

Security scanning for frontend dependencies:
- ✅ Automated via npm audit
- ✅ Integrated in CI/CD pipeline
- ✅ Weekly scheduled scans
- ✅ Clear vulnerability response procedures
- ✅ Regular dependency updates
- ✅ Supply chain protection

Regular audits and updates keep the codebase secure and maintain a healthy dependency tree.
