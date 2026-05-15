# Phase 14: Testing & Quality - Research

**Researched:** 2026-05-15
**Domain:** Test Coverage, Unit Testing, Integration Testing, E2E Testing, Multi-Language Test Infrastructure
**Confidence:** HIGH

## Summary

Phase 14 establishes comprehensive test coverage for all features added in v1.3 (Phases 10-13). The project has mature test infrastructure already in place: Go testing with testify for backend, Vitest with @testing-library/react for frontend, and Google Test for C++ collector. The primary work is filling coverage gaps for new features: replication monitoring, host monitoring, data classification, alerting, and frontend UI components.

**Primary recommendation:** Follow established test patterns from existing code. Focus on filling Wave 0 gaps identified in previous phases. Target 80% coverage for new code, integrate with existing CI/CD pipeline.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TEST-01 | All new collector plugins have C++ unit tests | GTest framework configured; existing patterns in `collector/tests/unit/` |
| TEST-02 | All new backend services have Go unit tests | Go testing + testify; existing patterns in `backend/internal/*_test.go` |
| TEST-03 | All new API endpoints have integration tests | Integration test patterns in `backend/tests/integration/` |
| TEST-04 | All new frontend components have tests | Vitest + @testing-library/react; existing patterns in `frontend/src/**/*.test.tsx` |
| TEST-05 | End-to-end tests cover critical user flows | Playwright configured; existing E2E tests in `frontend/e2e/tests/` |

</phase_requirements>

## Standard Stack

### Core (Backend - Go)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| testing | stdlib | Go test framework | Built-in, zero dependencies |
| testify | v1.9.0 | Assertions and mocking | Already used throughout backend |
| go test | 1.22 | Test runner with coverage | Standard Go toolchain |

### Core (Frontend - TypeScript)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vitest | v1.0.0 | Unit test framework | Fast, Vite-native, Jest-compatible API |
| @testing-library/react | v14.1.2 | React component testing | Standard for testing React components |
| @testing-library/jest-dom | v6.1.5 | DOM matchers | Extended assertions for DOM |
| @vitest/coverage-v8 | v1.0.0 | Code coverage | V8-native coverage, fast |

### Core (Collector - C++)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Google Test | 1.14+ | C++ test framework | Industry standard, already configured via CMake |
| nlohmann/json | 3.2.0+ | JSON assertions | Already used in collector for test data |

### Core (E2E)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Playwright | v1.59.1 | E2E browser testing | Cross-browser, trace viewer, auto-wait |
| @playwright/test | v1.59.1 | Test runner | Parallel execution, fixtures support |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| jsdom | v23.0.1 | DOM simulation for Vitest | Frontend unit tests |
| msw | Latest | API mocking | Frontend integration tests |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Vitest | Jest | Jest slower, requires Babel config; Vitest native to Vite |
| testify | gomock | testify simpler assertions; gomock for complex mock generation |
| Playwright | Cypress | Cypress slower, less reliable cross-browser; Playwright more modern |

**Installation:**
```bash
# Backend (Go) - already in go.mod
go mod download

# Frontend - already in package.json
cd frontend && npm install

# Collector (C++) - already in CMakeLists.txt
cd collector && cmake -B build && cmake --build build

# E2E browsers
cd frontend && npx playwright install
```

**Version verification:**
```bash
# Go
go version  # 1.22+
go list -m github.com/stretchr/testify  # v1.9.0

# Frontend
cd frontend && npm list vitest @testing-library/react

# Collector
cd collector/build && ctest --version

# E2E
cd frontend && npx playwright --version  # 1.59.1
```

## Architecture Patterns

### Recommended Project Structure
```
backend/
├── internal/
│   ├── auth/
│   │   └── service_test.go        # Unit tests for auth service
│   ├── api/
│   │   ├── handlers_test.go        # Handler unit tests
│   │   └── handlers_replication_test.go  # NEW for Phase 14
│   └── storage/
│       └── replication_store_test.go    # NEW for Phase 14
├── pkg/
│   ├── handlers/
│   │   ├── alerts_test.go          # Existing pattern
│   │   └── escalations_test.go     # Existing pattern
│   └── services/
│       └── condition_validator_test.go  # Existing pattern
├── tests/
│   ├── integration/
│   │   ├── alert_flow_test.go      # Existing
│   │   ├── replication_test.go     # NEW for Phase 14
│   │   └── classification_test.go  # NEW for Phase 14
│   └── security/
│       └── sql_injection_test.go   # Existing

frontend/
├── src/
│   ├── pages/
│   │   ├── ReplicationTopologyPage.test.tsx    # NEW
│   │   ├── DataClassificationPage.test.tsx     # NEW
│   │   └── HostInventoryPage.test.tsx          # NEW
│   ├── components/
│   │   ├── topology/
│   │   │   └── ReplicationGraph.test.tsx       # NEW
│   │   ├── classification/
│   │   │   └── ClassificationTable.test.tsx    # NEW
│   │   └── host/
│   │       └── HostStatusCard.test.tsx         # NEW
│   └── test/
│       └── setup.ts                 # Vitest setup (existing)
├── e2e/
│   ├── tests/
│   │   ├── 12-replication-topology.spec.ts     # NEW
│   │   ├── 13-data-classification.spec.ts      # NEW
│   │   └── 14-host-monitoring.spec.ts          # NEW
│   └── pages/
│       ├── ReplicationTopologyPage.ts          # NEW
│       └── DataClassificationPage.ts           # NEW
└── playwright.config.ts              # Existing config

collector/
├── tests/
│   ├── unit/
│   │   ├── replication_collector_test.cpp      # Existing
│   │   ├── host_inventory_test.cpp             # NEW
│   │   └── data_classification_test.cpp        # NEW
│   ├── integration/
│   │   └── multi_version_support_test.cpp      # Existing
│   └── CMakeLists.txt                 # Test configuration
```

### Pattern 1: Go Unit Test with Mock Store
**What:** Unit test using mock implementations of storage interfaces
**When to use:** Testing service layer logic without database
**Example:**
```go
// Source: backend/internal/auth/service_test.go (existing pattern)
func TestAuthService_LoginUser_Success(t *testing.T) {
    jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
    pm := NewPasswordManager()
    cm, _ := NewCertificateManager("", "")
    userStore := NewMockUserStore()
    collectorStore := NewMockCollectorStore()
    tokenStore := NewMockTokenStore()

    authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

    // Setup test data
    hashedPassword, err := pm.HashPassword("password123")
    require.NoError(t, err)
    user, _ := userStore.GetUserByUsername("testuser")
    user.PasswordHash = hashedPassword

    // Test
    resp, err := authService.LoginUser("testuser", "password123")

    require.NoError(t, err)
    assert.NotNil(t, resp)
    assert.NotEmpty(t, resp.Token)
}
```

### Pattern 2: Go Integration Test with Test Database
**What:** Integration test with real PostgreSQL container
**When to use:** Testing database operations, API handlers with DB
**Example:**
```go
// Source: backend/tests/integration/alert_flow_test.go (existing pattern)
func TestAlertFlow_EndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    db := setupTestDatabase(t)
    defer db.Close()

    // Create test data
    collector := createTestCollector(t, db)
    rule := createTestAlertRule(t, db, collector.ID)

    // Test alert trigger flow
    engine := jobs.NewAlertRuleEngine(db, nil)
    err := engine.EvaluateRule(context.Background(), rule.ID)
    require.NoError(t, err)

    // Verify alert created
    triggers, err := db.GetAlertTriggers(context.Background(), collector.ID)
    require.NoError(t, err)
    assert.Len(t, triggers, 1)
}
```

### Pattern 3: Vitest Component Test
**What:** React component test with Vitest and testing-library
**When to use:** Testing UI components, user interactions
**Example:**
```typescript
// Source: frontend/src/App.test.tsx (existing pattern)
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ReplicationTopologyPage } from './pages/ReplicationTopologyPage'
import { api } from './services/api'

vi.mock('./services/api')

describe('ReplicationTopologyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render replication topology graph', async () => {
    const mockData = {
      nodes: [
        { id: 'primary', type: 'primary', label: 'Primary' },
        { id: 'standby1', type: 'standby', label: 'Standby 1' },
      ],
      edges: [
        { source: 'primary', target: 'standby1' },
      ],
    }
    vi.mocked(api.getReplicationTopology).mockResolvedValue(mockData)

    render(<ReplicationTopologyPage />)

    await waitFor(() => {
      expect(screen.getByText('Primary')).toBeInTheDocument()
    })
    expect(screen.getByText('Standby 1')).toBeInTheDocument()
  })

  it('should display lag metrics on node hover', async () => {
    const user = userEvent.setup()
    vi.mocked(api.getReplicationTopology).mockResolvedValue(mockData)

    render(<ReplicationTopologyPage />)

    await waitFor(() => {
      expect(screen.getByText('Standby 1')).toBeInTheDocument()
    })

    await user.hover(screen.getByText('Standby 1'))

    await waitFor(() => {
      expect(screen.getByText(/replay_lag/i)).toBeInTheDocument()
    })
  })
})
```

### Pattern 4: Playwright E2E Test
**What:** End-to-end browser test with Playwright
**When to use:** Testing complete user flows, cross-browser validation
**Example:**
```typescript
// Source: frontend/e2e/tests/01-login-logout.spec.ts (existing pattern)
import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/LoginPage'
import { ReplicationTopologyPage } from '../pages/ReplicationTopologyPage'

test.describe('Replication Topology', () => {
  let loginPage: LoginPage
  let replicationPage: ReplicationTopologyPage

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page)
    replicationPage = new ReplicationTopologyPage(page)

    // Login first
    await loginPage.goto()
    await loginPage.login('admin', 'admin')
  })

  test('should display replication topology graph', async ({ page }) => {
    await replicationPage.goto()

    // Wait for topology to load
    await expect(page.locator('[data-testid="topology-graph"]')).toBeVisible()

    // Verify nodes rendered
    await expect(page.locator('.react-flow__node')).toHaveCount(3)
  })

  test('should show lag metrics on node click', async ({ page }) => {
    await replicationPage.goto()

    // Click on standby node
    await page.locator('[data-testid="node-standby-1"]').click()

    // Verify metrics panel appears
    await expect(page.locator('[data-testid="node-metrics-panel"]')).toBeVisible()
    await expect(page.locator('text=replay_lag')).toBeVisible()
  })
})
```

### Pattern 5: C++ Unit Test with GTest
**What:** C++ unit test with Google Test framework
**When to use:** Testing collector plugins, data processing logic
**Example:**
```cpp
// Source: collector/tests/unit/replication_collector_test.cpp (existing pattern)
#include "gtest/gtest.h"
#include "../../include/replication_plugin.h"
#include <nlohmann/json.hpp>

using json = nlohmann::json;

class ReplicationCollectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        hostname_ = "test-collector";
        collector_id_ = "test-replication-001";
        postgres_host_ = "localhost";
        postgres_port_ = 5432;
    }

    std::string hostname_;
    std::string collector_id_;
    std::string postgres_host_;
    int postgres_port_;
};

TEST_F(ReplicationCollectorTest, ConstructorInitializesCorrectly) {
    PgReplicationCollector collector(
        hostname_, collector_id_, postgres_host_, postgres_port_
    );

    EXPECT_EQ(collector.getType(), "pg_replication");
    EXPECT_TRUE(collector.isEnabled());
}

TEST_F(ReplicationCollectorTest, ParseLsnConvertsCorrectly) {
    // Test LSN parsing (hex X/XXXXXXXX format)
    // Uses internal helper or public method
}

// Integration test requiring real PostgreSQL
TEST_F(ReplicationCollectorTest, DISABLED_ExecuteReturnsValidJson) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(/* params */);
    json result = collector.execute();

    EXPECT_TRUE(result.contains("type"));
    EXPECT_EQ(result["type"], "pg_replication");
    EXPECT_TRUE(result["replication_slots"].is_array());
}
```

### Anti-Patterns to Avoid
- **Don't test implementation details:** Test behavior, not internal state (frontend)
- **Don't use `any` or type assertions:** Lose type safety in tests
- **Don't skip tests without reason:** Use build tags or environment checks
- **Don't mock what you don't own:** Mock external APIs, not your own code
- **Don't use sleep in tests:** Use proper async waiting with `waitFor`

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Test database setup | Manual container management | Docker Compose or testcontainers | Consistent environment, cleanup handled |
| Mock API responses | Custom fetch interceptors | MSW (Mock Service Worker) | Standard approach, network-level mocking |
| Assertion matchers | Custom expect functions | @testing-library/jest-dom | Well-tested, good error messages |
| Test fixtures | Hardcoded test data | Factory functions or fixtures library | Reusable, maintainable |

**Key insight:** Leverage existing test infrastructure. The project has mature patterns for all three languages. Focus on coverage gaps, not framework changes.

## Common Pitfalls

### Pitfall 1: Missing `t.Parallel()` in Go Tests
**What goes wrong:** Tests run sequentially, slow test suite
**Why it happens:** Developers forget to enable parallel execution
**How to avoid:** Add `t.Parallel()` at start of each test function
**Warning signs:** Test suite takes > 2 minutes to run

### Pitfall 2: Playwright Tests Don't Wait for API
**What goes wrong:** Tests fail intermittently due to timing issues
**Why it happens:** Not using Playwright's auto-waiting properly
**How to avoid:**
- Use `await expect(locator).toBeVisible()` not `await page.waitForTimeout()`
- Use `await page.waitForResponse()` for API calls
- Set proper test timeouts in config
**Warning signs:** Flaky tests, "element not found" errors

### Pitfall 3: Mocking Without Verification
**What goes wrong:** Tests pass but code doesn't work in production
**Why it happens:** Mocks return different data than real implementations
**How to avoid:**
- Use contract tests between mocks and real implementations
- Use MSW to mock at network level instead of module level
- Verify mock calls with `expect(mock).toHaveBeenCalledWith(...)`
**Warning signs:** Tests pass but integration fails

### Pitfall 4: C++ Tests Require Running PostgreSQL
**What goes wrong:** CI tests fail due to missing database
**Why it happens:** Integration tests need real PostgreSQL connection
**How to avoid:**
- Use `DISABLED_` prefix for tests requiring external resources
- Use `GTEST_SKIP()` when environment variable `CI` is set
- Separate unit tests from integration tests with CMake
**Warning signs:** Tests pass locally but fail in CI

### Pitfall 5: Frontend Tests Rely on Implementation Details
**What goes wrong:** Tests break on refactoring even though behavior unchanged
**Why it happens:** Testing component state or CSS classes instead of user-visible behavior
**How to avoid:**
- Use `getByRole`, `getByText`, `getByLabelText` instead of `getByTestId`
- Test user interactions, not component state
- Use `@testing-library/user-event` for realistic interactions
**Warning signs:** Tests fail after minor CSS or refactoring changes

## Code Examples

### Data Classification Unit Test (C++)
```cpp
// NEW FILE: collector/tests/unit/data_classification_test.cpp
#include "gtest/gtest.h"
#include "../../include/data_classification_plugin.h"
#include <nlohmann/json.hpp>
#include <regex>

using json = nlohmann::json;

class DataClassificationTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Setup test patterns
    }
};

// CPF validation algorithm test
TEST_F(DataClassificationTest, ValidateCPFWithValidNumber) {
    std::string valid_cpf = "123.456.789-09"; // Valid check digits
    EXPECT_TRUE(validateCPF(valid_cpf));
}

TEST_F(DataClassificationTest, ValidateCPFWithInvalidNumber) {
    std::string invalid_cpf = "123.456.789-00"; // Invalid check digits
    EXPECT_FALSE(validateCPF(invalid_cpf));
}

// Luhn algorithm test for credit cards
TEST_F(DataClassificationTest, ValidateCreditCardWithLuhn) {
    // Valid test card number (passes Luhn)
    std::string valid_card = "4111111111111111";
    EXPECT_TRUE(validateLuhn(valid_card));
}

TEST_F(DataClassificationTest, DetectCardTypeVisa) {
    EXPECT_EQ(detectCardType("4111111111111111"), "VISA");
}

// CNPJ validation test
TEST_F(DataClassificationTest, ValidateCNPJWithValidNumber) {
    std::string valid_cnpj = "11.222.333/0001-81";
    EXPECT_TRUE(validateCNPJ(valid_cnpj));
}
```

### Alert Rule Handler Integration Test (Go)
```go
// NEW FILE: backend/tests/integration/alert_rules_test.go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

func TestAlertRules_CRUD(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    db := setupTestDatabase(t)
    defer db.Close()

    ctx := context.Background()

    // Create
    rule := &models.AlertRule{
        Name:          "High Replication Lag",
        RuleType:      "threshold",
        MetricName:    "replication_lag_ms",
        Condition:     json.RawMessage(`{"operator": ">", "value": 5000}`),
        AlertSeverity: "high",
        IsEnabled:     true,
    }

    id, err := db.CreateAlertRule(ctx, rule)
    require.NoError(t, err)
    assert.NotZero(t, id)

    // Read
    fetched, err := db.GetAlertRule(ctx, id)
    require.NoError(t, err)
    assert.Equal(t, "High Replication Lag", fetched.Name)

    // Update
    fetched.AlertSeverity = "critical"
    err = db.UpdateAlertRule(ctx, fetched)
    require.NoError(t, err)

    // Delete
    err = db.DeleteAlertRule(ctx, id)
    require.NoError(t, err)

    _, err = db.GetAlertRule(ctx, id)
    assert.Error(t, err) // Should not exist
}

func TestAlertRules_Evaluation(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    // Create rule and test data
    rule := createTestAlertRule(t, db, "threshold", `{"operator": ">", "value": 1000}`)

    // Simulate metric data
    metric := &models.HostMetrics{
        CollectorID: rule.CollectorID,
        CpuUser:     50.0,
    }
    err := db.StoreHostMetrics(context.Background(), metric)
    require.NoError(t, err)

    // Evaluate rule
    engine := NewAlertRuleEngine(db, nil)
    result, err := engine.EvaluateRule(context.Background(), rule.ID)
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Replication Topology Component Test (TypeScript)
```typescript
// NEW FILE: frontend/src/pages/ReplicationTopologyPage.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ReplicationTopologyPage } from './ReplicationTopologyPage'
import { api } from '../services/api'

vi.mock('../services/api')

const mockTopology = {
  nodes: [
    { id: 'primary', type: 'primary', data: { label: 'Primary DB', lag: 0 } },
    { id: 'standby1', type: 'standby', data: { label: 'Standby 1', lag: 150 } },
    { id: 'standby2', type: 'standby', data: { label: 'Standby 2', lag: 320 } },
  ],
  edges: [
    { id: 'e1', source: 'primary', target: 'standby1' },
    { id: 'e2', source: 'primary', target: 'standby2' },
  ],
}

describe('ReplicationTopologyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.getReplicationTopology).mockResolvedValue(mockTopology)
  })

  it('should render topology graph with nodes', async () => {
    render(<ReplicationTopologyPage />)

    await waitFor(() => {
      expect(screen.getByText('Primary DB')).toBeInTheDocument()
    })

    expect(screen.getByText('Standby 1')).toBeInTheDocument()
    expect(screen.getByText('Standby 2')).toBeInTheDocument()
  })

  it('should show loading state initially', () => {
    vi.mocked(api.getReplicationTopology).mockImplementation(() => new Promise(() => {}))

    render(<ReplicationTopologyPage />)

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
  })

  it('should display error message on API failure', async () => {
    vi.mocked(api.getReplicationTopology).mockRejectedValue(new Error('API Error'))

    render(<ReplicationTopologyPage />)

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument()
    })
  })

  it('should show node details on click', async () => {
    const user = userEvent.setup()
    render(<ReplicationTopologyPage />)

    await waitFor(() => {
      expect(screen.getByText('Standby 1')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Standby 1'))

    await waitFor(() => {
      const detailsPanel = screen.getByTestId('node-details-panel')
      expect(within(detailsPanel).getByText('150 ms')).toBeInTheDocument()
    })
  })
})
```

### E2E Test for Data Classification (Playwright)
```typescript
// NEW FILE: frontend/e2e/tests/12-data-classification.spec.ts
import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/LoginPage'
import { DataClassificationPage } from '../pages/DataClassificationPage'

test.describe('Data Classification', () => {
  let loginPage: LoginPage
  let classificationPage: DataClassificationPage

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page)
    classificationPage = new DataClassificationPage(page)

    await loginPage.goto()
    await loginPage.login('admin', 'admin')
  })

  test('should display classification results', async ({ page }) => {
    await classificationPage.goto()

    // Wait for table to load
    await expect(page.locator('[data-testid="classification-table"]')).toBeVisible()

    // Verify columns exist
    await expect(page.locator('th', { hasText: 'Database' })).toBeVisible()
    await expect(page.locator('th', hasText: 'Pattern' })).toBeVisible()
    await expect(page.locator('th', { hasText: 'Category' })).toBeVisible()
  })

  test('should filter by pattern type', async ({ page }) => {
    await classificationPage.goto()

    // Select CPF filter
    await page.selectOption('[data-testid="pattern-filter"]', 'CPF')

    // Verify only CPF results shown
    await expect(page.locator('td', { hasText: 'CPF' })).toHaveCount(
      await page.locator('tbody tr').count()
    )
  })

  test('should drill down to column details', async ({ page }) => {
    await classificationPage.goto()

    // Click on a classification row
    await page.locator('tbody tr').first().click()

    // Verify detail panel opens
    await expect(page.locator('[data-testid="column-detail-panel"]')).toBeVisible()

    // Verify sample values shown (masked)
    await expect(page.locator('[data-testid="sample-values"]')).toBeVisible()
  })

  test('should export classification report', async ({ page }) => {
    await classificationPage.goto()

    // Start waiting for download
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.click('[data-testid="export-report"]'),
    ])

    // Verify download started
    expect(download.suggestedFilename()).toMatch(/classification.*\.csv/)
  })
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Jest | Vitest | Phase 05 | Faster test execution, native ESM |
| Manual E2E tests | Playwright automated | Phase 03 | Cross-browser, reliable |
| No coverage tracking | Codecov with thresholds | Phase 05 | Quality visibility |
| In-memory database | Docker PostgreSQL | Phase 02 | Real database behavior |

**Deprecated/outdated:**
- Jest: Replaced by Vitest in frontend
- Karma: Not used, Vitest handles browser testing
- Manual test scripts: Use Playwright instead

## Open Questions

1. **Coverage Threshold per Component**
   - What we know: Current threshold is 80% lines/functions, 70% branches
   - What's unclear: Should collector C++ have same threshold as backend Go?
   - Recommendation: Apply 70% threshold to C++ (harder to achieve high coverage with GTest), 80% to Go and TypeScript

2. **E2E Test Data Management**
   - What we know: Tests use seeded test database
   - What's unclear: How to reset database state between E2E tests
   - Recommendation: Use database transactions with rollback, or Docker container restart for full isolation

3. **Multi-tenant Test Isolation**
   - What we know: RLS policies added in Phase 11
   - What's unclear: How to test RLS policies in integration tests
   - Recommendation: Create dedicated RLS security tests that set tenant context and verify isolation

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go: testing + testify; C++: GTest; TypeScript: Vitest; E2E: Playwright |
| Config file | Go: none; C++: CMakeLists.txt; TS: vite.config.ts; E2E: playwright.config.ts |
| Quick run command | `go test ./... -short` (Go); `npm run test` (Frontend); `ctest` (C++) |
| Full suite command | `go test ./... -race -cover` (Go); `npm run test:coverage` (Frontend); `ctest -V` (C++) |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TEST-01 | Collector plugins C++ unit tests | unit | `cd collector/build && ctest -R unit` | Partial - existing tests |
| TEST-02 | Backend services Go unit tests | unit | `go test ./internal/... ./pkg/... -v` | Partial - Wave 0 |
| TEST-03 | API endpoints integration tests | integration | `go test ./tests/integration/... -v` | Partial - Wave 0 |
| TEST-04 | Frontend component tests | unit | `npm run test` | Partial - Wave 0 |
| TEST-05 | E2E tests for critical flows | e2e | `npm run test:e2e` | Partial - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./pkg/... -short` (Go); `npm run test` (Frontend unit)
- **Per wave merge:** `go test ./... -race -cover` (Go); `npm run test:coverage` (Frontend)
- **Phase gate:** Full suite green + 80% coverage on new code

### Wave 0 Gaps
- [ ] `backend/internal/api/handlers_replication_test.go` - Replication API tests
- [ ] `backend/internal/api/handlers_host_test.go` - Host monitoring API tests
- [ ] `backend/internal/storage/replication_store_test.go` - Replication storage tests
- [ ] `backend/internal/storage/classification_store_test.go` - Classification storage tests
- [ ] `backend/internal/services/health_score_calculator_test.go` - Health score logic tests
- [ ] `backend/tests/integration/replication_test.go` - End-to-end replication tests
- [ ] `backend/tests/integration/classification_test.go` - Classification flow tests
- [ ] `backend/tests/security/rls_isolation_test.go` - Multi-tenant RLS tests
- [ ] `collector/tests/unit/data_classification_test.cpp` - Classification plugin tests
- [ ] `collector/tests/unit/host_inventory_test.cpp` - Host inventory plugin tests
- [ ] `frontend/src/pages/ReplicationTopologyPage.test.tsx` - Topology page tests
- [ ] `frontend/src/pages/DataClassificationPage.test.tsx` - Classification page tests
- [ ] `frontend/src/pages/HostInventoryPage.test.tsx` - Host inventory page tests
- [ ] `frontend/src/components/topology/ReplicationGraph.test.tsx` - Graph component tests
- [ ] `frontend/e2e/tests/12-replication-topology.spec.ts` - Replication E2E
- [ ] `frontend/e2e/tests/13-data-classification.spec.ts` - Classification E2E
- [ ] `frontend/e2e/tests/14-host-monitoring.spec.ts` - Host monitoring E2E
- [ ] `frontend/e2e/pages/ReplicationTopologyPage.ts` - Page object model
- [ ] `frontend/e2e/pages/DataClassificationPage.ts` - Page object model

## Sources

### Primary (HIGH confidence)
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/auth/service_test.go` - Go test pattern
- `/Users/glauco.torres/git/pganalytics-v3/frontend/src/App.test.tsx` - Vitest test pattern
- `/Users/glauco.torres/git/pganalytics-v3/frontend/e2e/tests/01-login-logout.spec.ts` - Playwright pattern
- `/Users/glauco.torres/git/pganalytics-v3/collector/tests/unit/replication_collector_test.cpp` - GTest pattern
- `/Users/glauco.torres/git/pganalytics-v3/frontend/vite.config.ts` - Vitest configuration
- `/Users/glauco.torres/git/pganalytics-v3/frontend/playwright.config.ts` - Playwright configuration
- `/Users/glauco.torres/git/pganalytics-v3/.github/workflows/ci.yml` - CI configuration

### Secondary (MEDIUM confidence)
- `/Users/glauco.torres/git/pganalytics-v3/backend/tests/integration/` - Integration test patterns
- `/Users/glauco.torres/git/pganalytics-v3/collector/tests/CMakeLists.txt` - CMake test configuration
- `/Users/glauco.torres/git/pganalytics-v3/.planning/phases/10-collector-backend-foundation/10-RESEARCH.md` - Wave 0 gaps from Phase 10
- `/Users/glauco.torres/git/pganalytics-v3/.planning/phases/11-data-classification-health-analysis/11-RESEARCH.md` - Wave 0 gaps from Phase 11
- `/Users/glauco.torres/git/pganalytics-v3/.planning/phases/12-alerting-system/12-RESEARCH.md` - Wave 0 gaps from Phase 12

### Tertiary (LOW confidence)
- None - all findings verified against source code and existing patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All frameworks already configured and in use
- Architecture: HIGH - Clear patterns from existing test files
- Pitfalls: HIGH - Based on common testing anti-patterns and existing codebase issues

**Research date:** 2026-05-15
**Valid until:** 30 days (stable test frameworks, established patterns)