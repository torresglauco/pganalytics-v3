# pgAnalytics v3 Testing Guide

## Overview

pgAnalytics v3 maintains comprehensive test coverage across all components with unit tests, integration tests, and end-to-end tests. This guide covers testing strategies, commands, and best practices.

## Test Coverage by Component

### Backend (Go)
- **Unit Tests**: 233+ tests
- **Integration Tests**: 40+ tests
- **E2E Tests**: Full workflow tests
- **Coverage**: >90%
- **Location**: `backend/internal/**/tests`, `backend/pkg/**/tests`

Key areas:
- API endpoints and handlers
- Database operations and migrations
- Authentication and authorization
- Metric processing and aggregation
- WebSocket communication

### Frontend (Node.js)
- **Unit Tests**: 386+ tests
- **Component Tests**: 120+ tests
- **E2E Tests (Playwright)**: 6+ tests
- **Coverage**: >85%
- **Location**: `frontend/src/**/__tests__`, `frontend/e2e`

Key areas:
- React components
- State management (Redux)
- API integration
- UI interactions
- Form validation

### Collector (C++)
- **Unit Tests**: 228+ tests
- **Integration Tests**: Full pipeline tests
- **Coverage**: >80%
- **Location**: `collector/tests`

Key areas:
- Metric collection logic
- PostgreSQL connection pooling
- Data serialization
- Error handling and retry logic

### CLI (Go)
- **Unit Tests**: 6+ tests
- **Integration Tests**: 3+ tests
- **E2E Tests**: Full workflow tests
- **Location**: `backend/cmd/pganalytics-cli/tests`

Key areas:
- Command parsing
- Configuration management
- API client operations
- Output formatting

### MCP Server (Go)
- **Unit Tests**: 76+ tests
- **Validation Tests**: 12 tests
- **Integration Tests**: 25 tests
- **E2E Tests**: 22 tests
- **Coverage**: >45%
- **Location**: `backend/internal/mcp/tests`

Key areas:
- Tool registration and invocation
- stdio JSON-RPC protocol
- Database query execution
- Error handling

## Running Tests

### Backend Tests

#### All Tests
```bash
cd backend
go test ./...
```

#### Unit Tests Only
```bash
go test ./... -short
```

#### With Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Specific Package
```bash
go test ./internal/api/handlers -v
```

#### Using Mise (Task Runner)
```bash
mise run test:backend       # All tests
mise run test:backend:unit  # Unit only
mise run test:integration   # Integration only
mise run test:backend:race  # Race condition detection
```

### Frontend Tests

#### All Tests
```bash
cd frontend
npm test
```

#### Unit Tests Only
```bash
npm run test:unit
```

#### E2E Tests (Playwright)
```bash
npm run test:e2e
npm run test:e2e:ui    # Interactive mode
npm run test:e2e:debug # Debug mode
```

#### Coverage Report
```bash
npm test -- --coverage
```

#### Watch Mode (Development)
```bash
npm test -- --watch
```

#### Using Mise
```bash
mise run test:frontend      # All
mise run test:frontend:unit # Unit only
mise run test:frontend:e2e  # E2E only
```

### Collector Tests

#### C++ Tests
```bash
cd collector
cmake -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build --target test
```

#### Using Mise
```bash
mise run test:collector
```

### CLI Tests

#### Run CLI Tests
```bash
cd backend
go test ./cmd/pganalytics-cli/tests -v
```

### MCP Tests

#### Run MCP Tests
```bash
cd backend
go test ./internal/mcp/tests -v
```

#### Specific Test Category
```bash
go test ./internal/mcp/tests -run TestTools -v
go test ./internal/mcp/tests -run TestValidation -v
go test ./internal/mcp/tests -run TestIntegration -v
```

## Test Organization

### Naming Conventions

**Unit Tests**: `*_test.go` files in same package
```go
package handlers

import "testing"

func TestGetMetrics(t *testing.T) {
    // Test implementation
}
```

**Integration Tests**: `*_integration_test.go` files
```go
func TestGetMetrics_Integration(t *testing.T) {
    // Integration test with database
}
```

**E2E Tests**: Playwright tests in `frontend/e2e`
```javascript
test('should display metrics dashboard', async ({ page }) => {
    // E2E test
});
```

### Test Structure

Each test should follow AAA pattern:
- **Arrange**: Set up test data and mocks
- **Act**: Execute the function/component
- **Assert**: Verify the results

```go
func TestProcessMetric(t *testing.T) {
    // Arrange
    metric := createTestMetric()
    processor := NewMetricProcessor()

    // Act
    result := processor.Process(metric)

    // Assert
    if result.Status != "processed" {
        t.Errorf("expected processed, got %s", result.Status)
    }
}
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- **Push to main**: Full test suite
- **Pull Requests**: Full test suite + coverage checks
- **Nightly**: Extended tests including compatibility checks

**Workflow File**: `.github/workflows/test.yml`

Configuration:
```yaml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: |
          mise run test:backend
          mise run test:frontend
          mise run test:collector
```

### Coverage Requirements

- **Minimum coverage**: 80% for all components
- **Backend**: >90% target
- **Frontend**: >85% target
- **Pull Requests**: Must not decrease coverage
- **Reports**: Published to Codecov

## Test Data and Fixtures

### Backend Fixtures

Location: `backend/internal/testutil`

```go
func createTestMetric() *Metric {
    return &Metric{
        AgentID: "test-agent",
        Timestamp: time.Now(),
        Values: map[string]float64{
            "cpu_usage": 45.2,
            "memory_usage": 2048,
        },
    }
}
```

### Database Test Setup

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create mock: %v", err)
    }
    return db
}
```

### Frontend Test Utilities

Location: `frontend/src/testutil`

```typescript
export const createMockMetrics = (count: number): Metric[] => {
    return Array.from({ length: count }, (_, i) => ({
        id: `metric-${i}`,
        timestamp: new Date(),
        value: Math.random() * 100,
    }));
};
```

## Mocking Strategies

### Backend Mocking

**HTTP Mocking** (httptest package):
```go
server := httptest.NewServer(handler)
defer server.Close()
```

**Database Mocking** (sqlmock):
```go
mock.ExpectQuery("SELECT.*").WillReturnRows(rows)
```

**Interface Mocking** (mockgen):
```bash
mockgen -source=service.go -destination=mocks/service.go
```

### Frontend Mocking

**Component Mocking** (Jest):
```javascript
jest.mock('../api/client', () => ({
    getMetrics: jest.fn(() => Promise.resolve(mockData))
}));
```

**API Mocking** (MSW - Mock Service Worker):
```javascript
server.use(
    rest.get('/api/v1/metrics', (req, res, ctx) => {
        return res(ctx.json(mockMetrics));
    })
);
```

## Performance Testing

### Load Testing

**Backend Load Tests**:
```bash
# Using Apache Bench
ab -n 1000 -c 10 http://localhost:8080/api/v1/metrics

# Using k6
k6 run backend/tests/load.js
```

**Frontend Performance Tests**:
```bash
npm run test:performance
```

### Benchmarking

**Go Benchmarks**:
```go
func BenchmarkProcessMetric(b *testing.B) {
    metric := createTestMetric()
    processor := NewMetricProcessor()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processor.Process(metric)
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

## Debugging Tests

### Backend Debugging

**Print Debugging**:
```go
t.Logf("debug info: %+v", variable)
```

**Running Single Test**:
```bash
go test -run TestSpecificTest -v
```

**Verbose Output**:
```bash
go test -v ./...
```

**Debugging with Delve**:
```bash
dlv test ./... -- -test.run TestName
```

### Frontend Debugging

**Debug Mode**:
```bash
npm test -- --inspect-brk --runInBand
```

**E2E Debug Mode**:
```bash
npm run test:e2e:debug
```

**Browser DevTools**:
```bash
HEADED=true npm run test:e2e
```

## Test Maintenance

### Flaky Tests

Flaky tests that intermittently fail should:
1. Be isolated and identified
2. Have root cause analyzed
3. Be refactored or removed
4. Not be skipped (t.Skip) without issue tracking

**Finding Flaky Tests**:
```bash
# Run tests multiple times
for i in {1..10}; do go test ./...; done
```

### Test Dependencies

Tests should be:
- **Independent**: No inter-test dependencies
- **Deterministic**: Same result every run
- **Fast**: Complete quickly
- **Isolated**: No side effects

### Updating Tests

When adding features:
1. Write tests first (TDD)
2. Implement feature
3. All tests pass
4. Update integration tests if needed
5. Verify E2E tests

## Best Practices

### Do's
- Test behavior, not implementation
- Use table-driven tests for multiple scenarios
- Keep tests focused and small
- Use meaningful assertion messages
- Mock external dependencies
- Test error cases
- Clean up resources in teardown

### Don'ts
- Don't skip flaky tests without fixing
- Don't have tests dependent on order
- Don't test private implementation details
- Don't use real databases in unit tests
- Don't make tests timeout-dependent
- Don't hardcode data in tests

## Coverage Report Analysis

### Viewing Coverage

**HTML Report**:
```bash
go tool cover -html=coverage.out
```

**Coverage by Function**:
```bash
go tool cover -func=coverage.out
```

### Improving Coverage

Target coverage by package:
```bash
go test -cover ./... | awk '{print $NF}' | sort -n
```

## Continuous Testing

### Pre-commit Testing

Setup git hooks:
```bash
cp scripts/hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

### Watch Mode Development

**Backend**:
```bash
go install github.com/cosmtrek/air@latest
air -c .air.toml
```

**Frontend**:
```bash
npm test -- --watch
```

## Troubleshooting

### Common Issues

**Tests Timeout**:
```bash
# Increase timeout
go test -timeout 30s ./...
```

**Database Connection Issues**:
- Ensure test database is running
- Check DATABASE_URL environment variable
- Verify migrations are applied

**Port Already in Use**:
- Use dynamic port assignment in tests
- Kill existing processes: `lsof -ti:PORT | xargs kill -9`

**Import Cycle Errors**:
- Review package dependencies
- Move shared code to util packages
- Use interfaces to break cycles

## Additional Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Jest Documentation](https://jestjs.io/)
- [Playwright Documentation](https://playwright.dev/)
- [testutil Package](./backend/internal/testutil)
