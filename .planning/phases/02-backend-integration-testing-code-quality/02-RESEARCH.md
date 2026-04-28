# Phase 2: Backend Integration Testing & Code Quality - Research

**Researched:** 2026-04-28
**Domain:** Go backend testing, API integration testing, code quality tooling
**Confidence:** HIGH

## Summary

Phase 2 focuses on establishing comprehensive API testing and code quality baseline for the Go backend. The project already has significant testing infrastructure in place from Week 1 (Phase 1), including 2,734+ lines of boundary integration tests covering auth, collectors, instances, users, and validation endpoints. The primary work involves completing test coverage for remaining endpoints, establishing golangci-lint configuration, and implementing secret scanning.

**Primary recommendation:** Build upon existing test infrastructure using testify assertions and httptest patterns. Create a `.golangci.yml` configuration for consistent linting. Install gitleaks for secret scanning integration.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| testing | Go 1.24.0 (stdlib) | Test framework | Built-in, no external dependencies |
| stretchr/testify | v1.11.1 | Assertions and mocks | Industry standard for Go testing |
| gin-gonic/gin | v1.10.0 | HTTP router and testing | Project's web framework |
| net/http/httptest | Go 1.24.0 (stdlib) | HTTP test utilities | Standard for API testing |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| google/uuid | v1.6.0 | UUID generation/validation | Collector IDs, resource identifiers |
| stretchr/testify/require | v1.11.1 | Fatal assertions | When test cannot proceed on failure |
| lib/pq | v1.10.9 | PostgreSQL driver | Integration tests with real DB |

### Code Quality Tools
| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| golangci-lint | v2.11.4 | Go linter aggregator | Already installed |
| gitleaks | latest | Secret scanning | `brew install gitleaks` |
| go fmt | Go 1.24.0 (stdlib) | Code formatting | Built-in |
| go vet | Go 1.24.0 (stdlib) | Static analysis | Built-in |

**Version verification:**
```bash
# Already verified in project
golangci-lint --version
# golangci-lint has version 2.11.4 built with go1.26.1

go version
# go1.24.0 (from go.mod)
```

## Architecture Patterns

### Recommended Test Structure
```
backend/
├── tests/
│   ├── integration/           # Integration tests (API boundary tests)
│   │   ├── boundary_auth_test.go
│   │   ├── boundary_collectors_test.go
│   │   ├── boundary_instances_test.go
│   │   ├── boundary_users_test.go
│   │   ├── boundary_validation_test.go
│   │   ├── testhelpers.go     # Shared test utilities
│   │   └── boundary_test_helpers.go
│   ├── mocks/                 # Mock implementations
│   │   └── ml_service_mock.go
│   ├── unit/                  # Unit tests
│   └── security/              # Security-focused tests
└── internal/
    └── */                     # Package-level *_test.go files
        └── handlers_test.go
```

### Pattern 1: Boundary Testing with httptest
**What:** Test API endpoints at HTTP layer with request/response validation
**When to use:** All API integration tests (TEST-01 through TEST-06)
**Example:**
```go
// Source: Existing project pattern - backend/tests/integration/boundary_auth_test.go
func TestLoginBoundary_EmptyUsername(t *testing.T) {
    router, _, _ := newTestEnv(t)

    loginReq := models.LoginRequest{
        Username: "",
        Password: "password123",
    }

    body, _ := json.Marshal(loginReq)
    req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Empty username should be rejected
    assert.Equal(t, http.StatusBadRequest, w.Code, "Empty username should return 400")
}
```

### Pattern 2: Test Environment Setup
**What:** Create isolated test environment with mock stores
**When to use:** All integration tests that need database-like behavior
**Example:**
```go
// Source: backend/tests/integration/boundary_test_helpers.go
func newTestEnv(t *testing.T) (*gin.Engine, *TestUserStore, *TestCollectorStore) {
    t.Helper()

    userStore := NewTestUserStore()
    collectorStore := NewTestCollectorStore()
    tokenStore := NewTestTokenStore()
    _, router := createTestServer(userStore, collectorStore, tokenStore)

    return router, userStore, collectorStore
}
```

### Pattern 3: Mock Store Implementation
**What:** In-memory implementations of data stores for testing
**When to use:** Testing without real database (fast, isolated tests)
**Example:**
```go
// Source: backend/tests/integration/testhelpers.go
type TestUserStore struct {
    users map[string]*models.User
}

func NewTestUserStore() *TestUserStore {
    pm := auth.NewPasswordManager()
    hash, _ := pm.HashPassword("password123")

    return &TestUserStore{
        users: map[string]*models.User{
            "testuser": {
                ID:           1,
                Username:     "testuser",
                Email:        "test@example.com",
                PasswordHash: hash,
                FullName:     "Test User",
                Role:         "user",
                IsActive:     true,
            },
        },
    }
}
```

### Pattern 4: Table-Driven Tests for HTTP Status Codes
**What:** Test multiple scenarios in one test function
**When to use:** TEST-06 (HTTP status codes coverage)
**Example:**
```go
func TestHTTPStatusCodes(t *testing.T) {
    tests := []struct {
        name         string
        method       string
        endpoint     string
        body         interface{}
        expectedCode int
    }{
        {"Health check returns 200", "GET", "/api/v1/health", nil, http.StatusOK},
        {"Invalid login returns 401", "POST", "/api/v1/auth/login", models.LoginRequest{}, http.StatusUnauthorized},
        {"Protected endpoint without auth returns 401", "GET", "/api/v1/users", nil, http.StatusUnauthorized},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            router, _, _ := newTestEnv(t)
            // Execute request and assert status code
        })
    }
}
```

### Anti-Patterns to Avoid
- **Silent error catching:** `try { test } catch { console.log }` - Test can pass while failing. Use explicit assertions instead.
- **Testing implementation details:** Focus on HTTP behavior, not internal function calls.
- **Hardcoded test data in multiple places:** Use test helpers and fixtures.
- **Skipping authentication in tests:** Use proper auth tokens or test unauthenticated paths explicitly.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Mock HTTP server | Custom httptest setup | httptest.NewServer | Built-in, handles lifecycle |
| Test assertions | Manual if/else checks | stretchr/testify | Clear failure messages, type-safe |
| Mock stores | Partial interfaces | Full interface implementation | Ensures all methods work |
| Table tests | Duplicated test functions | Single table-driven test | Easier to add new cases |
| Secret detection | Regex patterns | gitleaks | Covers 100+ secret types, low false positives |
| Linting config | Manual linter calls | golangci-lint | Aggregates 50+ linters, single config |

**Key insight:** The project already has excellent test helper patterns established. Extend them rather than creating new patterns.

## Common Pitfalls

### Pitfall 1: Tests Pass Despite Failures (Silent Catches)
**What goes wrong:** Tests with try-catch blocks that log errors but don't fail
**Why it happens:** Week 1 E2E tests had silent error catching that hid failures
**How to avoid:** Always use explicit assertions; never catch and ignore errors in tests
**Warning signs:** Tests with `if err != nil { t.Log(err) }` instead of `require.NoError(t, err)`

### Pitfall 2: Inconsistent Boundary Testing
**What goes wrong:** Some endpoints have boundary tests, others don't
**Why it happens:** Ad-hoc test writing without systematic coverage plan
**How to avoid:** Use the existing boundary test files as templates; ensure all API categories have boundary test coverage
**Warning signs:** New endpoints added without corresponding boundary tests

### Pitfall 3: Linter Config Drift
**What goes wrong:** No `.golangci.yml` means default config changes between versions
**Why it happens:** Relying on golangci-lint defaults without explicit configuration
**How to avoid:** Create `.golangci.yml` at project root with explicit linter settings
**Warning signs:** Different linting behavior on different machines or CI runs

### Pitfall 4: Missing Secret Scanner
**What goes wrong:** Hardcoded credentials slip into commits
**Why it happens:** No automated secret detection in workflow
**How to avoid:** Install gitleaks, add to pre-commit hooks and CI pipeline
**Warning signs:** Credentials found in git history (already fixed in Week 1)

### Pitfall 5: Test Helper Not Using t.Helper()
**What goes wrong:** Error messages point to helper function instead of test
**Why it happens:** Forgetting to mark helpers
**How to avoid:** Always call `t.Helper()` at start of test helper functions
**Warning signs:** Stack traces don't show actual failing test line

## Code Examples

### HTTP Status Code Test Pattern (TEST-06)
```go
// Pattern for comprehensive HTTP status code coverage
func TestAPIEndpoints_StatusCodes(t *testing.T) {
    router, _, _ := newTestEnv(t)

    statusTests := []struct {
        name       string
        method     string
        path       string
        body       interface{}
        wantStatus int
    }{
        // 200 OK - Successful requests
        {"Health endpoint returns 200", "GET", "/api/v1/health", nil, http.StatusOK},

        // 400 Bad Request - Validation errors
        {"Empty username returns 400", "POST", "/api/v1/auth/login",
            models.LoginRequest{Username: "", Password: "pass"}, http.StatusBadRequest},

        // 401 Unauthorized - Missing/invalid auth
        {"Protected endpoint without token returns 401", "GET", "/api/v1/users", nil, http.StatusUnauthorized},

        // 403 Forbidden - Insufficient permissions
        {"Non-admin creating user returns 403", "POST", "/api/v1/users",
            models.CreateUserRequest{}, http.StatusForbidden},

        // 404 Not Found - Resource doesn't exist
        {"Invalid collector UUID returns 404", "GET", "/api/v1/collectors/"+uuid.New().String(), nil, http.StatusNotFound},
    }

    for _, tt := range statusTests {
        t.Run(tt.name, func(t *testing.T) {
            var body bytes.Buffer
            if tt.body != nil {
                json.NewEncoder(&body).Encode(tt.body)
            }

            req := httptest.NewRequest(tt.method, tt.path, &body)
            if tt.body != nil {
                req.Header.Set("Content-Type", "application/json")
            }

            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code, "Status code mismatch")
        })
    }
}
```

### golangci-lint Configuration (QUAL-01)
```yaml
# .golangci.yml - Recommended configuration for Phase 2
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    # Default linters plus these additions for production code
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Report suspicious constructs
    - ineffassign   # Detect unused assignments
    - staticcheck   # Go static analysis
    - unused        # Find unused code
    - bodyclose     # Check HTTP response body is closed
    - gosec         # Security checks
    - gofmt         # Check formatting
    - goimports     # Check imports
    - misspell      # Find misspelled words
    - revive        # General purpose linter
    - sqlclosecheck # Check sql.Rows and sql.Stmt are closed

linters-settings:
  gosec:
    excludes:
      - G104 # Audit errors (will be handled by errcheck)
  errcheck:
    check-type-assertions: true
    check-blank: true

issues:
  exclude-rules:
    # Exclude test files from some linters
    - path: _test\.go
      linters:
        - errcheck
```

### Secret Scanning with Gitleaks (QUAL-03)
```bash
# Install gitleaks
brew install gitleaks

# Scan entire repository
gitleaks detect --source . --verbose

# Scan staged changes (for pre-commit)
gitleaks detect --source . --staged

# Generate report
gitleaks detect --source . --report-path gitleaks-report.json
```

### Mock/Stub Configuration Documentation (TEST-21)
```go
// Mock documentation pattern - Document in tests/mocks/README.md
/*
# Mock Libraries Configuration

## TestUserStore
Provides in-memory user storage for authentication tests.

Usage:
    userStore := NewTestUserStore()
    // Default user: "testuser" / "password123" with role "user"

## TestCollectorStore
Provides in-memory collector storage for collector endpoint tests.

Usage:
    collectorStore := NewTestCollectorStore()
    // Empty by default, use CreateCollector() to add

## TestTokenStore
Provides in-memory API token storage for token validation tests.

## MockMLService
Provides mock ML service HTTP server for ML integration tests.
Located in: tests/mocks/ml_service_mock.go

Usage:
    mockML := NewMockMLService()
    defer mockML.Close()

    // Configure behavior
    mockML.SetShouldFail(true)
    mockML.SetResponseDelay(100 * time.Millisecond)

    // Use URL
    config.MLServiceURL = mockML.URL()
*/
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| LocalStorage for JWT tokens | httpOnly cookies | Week 1 (April 2026) | XSS protection |
| Silent test error catching | Explicit assertions | Week 1 (April 2026) | Reliable tests |
| Default golangci-lint | Custom .golangci.yml | Phase 2 (planned) | Consistent linting |
| No secret scanning | Gitleaks integration | Phase 2 (planned) | Prevent credential leaks |

**Deprecated/outdated:**
- `demo@pganalytics.com` login: Updated to `admin/admin` credentials in Week 1
- MD5 hash for UUID generation: Replaced with UUID v4 in Week 1
- CORS `Allow-Origin: *`: Replaced with whitelist in Week 1

## Open Questions

1. **Database Integration Tests vs Mock Stores**
   - What we know: Current boundary tests use mock stores (no real DB)
   - What's unclear: Whether Phase 3 database tests should use testcontainers or separate test DB
   - Recommendation: Phase 2 continues with mock stores; Phase 3 introduces real DB testing

2. **Coverage Baseline Measurement**
   - What we know: `go test -coverprofile=coverage.out ./...` generates coverage
   - What's unclear: Current baseline percentage
   - Recommendation: Run `make test-backend` to establish baseline before Phase 2 work

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) + stretchr/testify v1.11.1 |
| Config file | None (Go uses `*_test.go` convention) |
| Quick run command | `go test ./backend/... -v -short` |
| Full suite command | `make test-backend` (includes coverage) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TEST-01 | API endpoints integration tests | integration | `go test ./backend/tests/integration/... -v` | ✅ Partial (boundary tests exist) |
| TEST-02 | Authentication boundary tests | integration | `go test ./backend/tests/integration/... -run TestLoginBoundary -v` | ✅ boundary_auth_test.go |
| TEST-03 | Collector endpoints boundary | integration | `go test ./backend/tests/integration/... -run TestCollector -v` | ✅ boundary_collectors_test.go |
| TEST-04 | Instance endpoints testing | integration | `go test ./backend/tests/integration/... -run TestManagedInstance -v` | ✅ boundary_instances_test.go |
| TEST-05 | User management permissions | integration | `go test ./backend/tests/integration/... -run TestCreateUser -v` | ✅ boundary_users_test.go |
| TEST-06 | HTTP status codes coverage | integration | `go test ./backend/tests/integration/... -run TestHTTP -v` | ❌ Wave 0 (need status code suite) |
| QUAL-01 | Go linting | lint | `golangci-lint run ./backend/...` | ⚠️ No .golangci.yml config |
| QUAL-03 | No hardcoded secrets | security | `gitleaks detect --source .` | ❌ Wave 0 (gitleaks not installed) |
| TEST-21 | Mock/stub configuration | docs | N/A (documentation) | ✅ mocks/ml_service_mock.go exists |

### Sampling Rate
- **Per task commit:** `go test ./backend/... -short`
- **Per wave merge:** `make test-backend`
- **Phase gate:** Full suite green + lint pass + secret scan clean

### Wave 0 Gaps
- [ ] `backend/tests/integration/http_status_codes_test.go` — covers TEST-06 (comprehensive status code suite)
- [ ] `.golangci.yml` — QUAL-01 configuration at project root
- [ ] `gitleaks` installation — QUAL-03 secret scanning tool
- [ ] `.pre-commit-config.yaml` — Optional: hook for lint + secret scan (can defer to Phase 5)

*(If no gaps: "None — existing test infrastructure covers all phase requirements")*

## Sources

### Primary (HIGH confidence)
- Project files examined: `go.mod`, `Makefile`, `backend/tests/integration/*.go` (April 2026)
- golangci-lint installation verified: v2.11.4 (March 2026)
- Week 1 documentation: `WEEK1_FINAL_SUMMARY.md` (April 2026)

### Secondary (MEDIUM confidence)
- testify library documentation: github.com/stretchr/testify (v1.11.1 in go.mod)
- httptest patterns from Go stdlib documentation

### Tertiary (LOW confidence)
- gitleaks best practices from project knowledge (not yet verified with official docs)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Directly examined project files and dependencies
- Architecture: HIGH - Existing test patterns documented and verified
- Pitfalls: HIGH - Based on actual Week 1 issues and fixes
- Code quality tools: MEDIUM - golangci-lint verified, gitleaks needs installation

**Research date:** 2026-04-28
**Valid until:** 30 days (stable Go testing patterns, golangci-lint config stable across versions)

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TEST-01 | API endpoints integration tests (happy path + error cases) | Boundary test patterns established in boundary_*.go files; extend to all endpoints |
| TEST-02 | Authentication boundary tests | Existing boundary_auth_test.go covers login, setup, password change; add token expiration/revocation |
| TEST-03 | Collector endpoints boundary validation | Existing boundary_collectors_test.go covers registration, metrics push; add SQL injection tests |
| TEST-04 | Instance endpoints version/configuration testing | Existing boundary_instances_test.go covers port, name, status validation; add PG version tests |
| TEST-05 | User management permission boundaries | Existing boundary_users_test.go covers CRUD; add admin vs user permission tests |
| TEST-06 | HTTP status codes coverage | Need new test file: http_status_codes_test.go with table-driven tests |
| QUAL-01 | Go linting and formatting | golangci-lint installed (v2.11.4); create .golangci.yml configuration |
| QUAL-03 | No hardcoded secrets | Install gitleaks, create scanning workflow |
| TEST-21 | Mock/stub configuration | Existing mocks: TestUserStore, TestCollectorStore, TestTokenStore, MockMLService; document usage |
</phase_requirements>