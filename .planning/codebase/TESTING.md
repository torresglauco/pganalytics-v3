# Testing Patterns

**Analysis Date:** 2026-03-30

## Test Framework

**Runner:**
- Frontend: Vitest 1.0.0
- Config: `frontend/vite.config.ts` with `test` property
- Backend: Go's built-in `testing` package

**Assertion Library:**
- Frontend: @testing-library/react with vitest (uses `expect()` from vitest)
- Backend: standard Go `testing` package with error comparisons

**Run Commands:**
```bash
npm test              # Run all tests (frontend)
npm run test:ui       # Interactive test UI (frontend)
npm run test:coverage # Generate coverage report
go test ./...         # Run all tests (backend)
go test -v ./...      # Verbose test output
```

## Test File Organization

**Location:**
- Frontend: co-located with source files (same directory)
  - Example: `frontend/src/App.test.tsx` next to `frontend/src/App.tsx`
  - Components: `frontend/src/components/CollectorForm.test.tsx` next to component
  - Services: `frontend/src/services/realtime.test.ts` next to service
  - Stores: `frontend/src/stores/realtimeStore.test.ts` next to store
- Backend: separate `tests/` directories organized by type
  - Location: `backend/tests/unit/`, `backend/tests/integration/`, `backend/tests/load/`
  - Same package: test files in same package with `_test.go` suffix

**Naming:**
- Frontend: `[SourceFile].test.tsx` or `[SourceFile].test.ts`
- Backend: `[descriptor]_test.go`

**Structure:**
```
frontend/src/
├── App.tsx
├── App.test.tsx
├── components/
│   ├── CollectorForm.tsx
│   ├── CollectorForm.test.tsx
│   └── [other components]
├── services/
│   ├── api.ts
│   ├── realtime.ts
│   └── realtime.test.ts
└── stores/
    ├── authStore.ts
    └── realtimeStore.test.ts

backend/
├── tests/
│   ├── unit/
│   │   ├── circuit_breaker_test.go
│   │   └── postgresql_logs_migration_test.go
│   ├── integration/
│   │   ├── handlers_test.go
│   │   └── ml_client_test.go
│   ├── load/
│   │   ├── load_test.go
│   │   └── phase5_load_test.go
│   └── benchmarks/
│       ├── circuit_breaker_bench.go
│       └── ml_client_bench.go
└── internal/
    ├── auth/
    │   ├── mfa.go
    │   └── mfa_test.go
    └── [other packages]
```

## Test Structure

**Suite Organization:**
```typescript
// Frontend pattern using vitest + @testing-library/react
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'

describe('ComponentName', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset state/stores
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('specific behavior group', () => {
    it('should do something specific', () => {
      // Arrange
      const mockData = { ... }

      // Act
      render(<Component {...mockData} />)

      // Assert
      expect(screen.getByTestId('element')).toBeInTheDocument()
    })
  })
})
```

```go
// Backend Go pattern
func TestFeatureName(t *testing.T) {
  tests := []struct {
    name     string
    input    string
    expected string
  }{
    {
      name:     "valid case",
      input:    "test",
      expected: "test",
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      result := FunctionUnderTest(tt.input)
      if result != tt.expected {
        t.Errorf("expected %v, got %v", tt.expected, result)
      }
    })
  }
}
```

**Patterns:**
- Setup: `beforeEach()` for test isolation (mocks cleared, state reset)
- Teardown: `afterEach()` for cleanup (mocks verified, localStorage cleared)
- Assertion: expect-based assertions with `.toBeInTheDocument()`, `.toHaveBeenCalled()`, etc.

## Mocking

**Framework:**
- Frontend: `vitest` - uses `vi.mock()`, `vi.fn()`, `vi.spyOn()`, `vi.mocked()`
- Backend: manual mock structs

**Patterns:**
```typescript
// Module mocking
vi.mock('../services/api', () => {
  const mockClient = {
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    emit: vi.fn(),
  }
  return { realtimeClient: mockClient }
})

// Function mocking and assertion
const mockOnSuccess = vi.fn()
render(<Component onSuccess={mockOnSuccess} />)
expect(mockOnSuccess).toHaveBeenCalledWith(expectedData)

// Mock resolution
vi.mocked(apiClient.testConnection).mockResolvedValue(true)
vi.mocked(apiClient.registerCollector).mockRejectedValueOnce(new Error('Failed'))
```

```go
// Backend Go mock pattern
type MockSMSProvider struct {
  messages []string
  err      error
}

func (m *MockSMSProvider) SendSMS(phoneNumber, message string) error {
  if m.err != nil {
    return m.err
  }
  m.messages = append(m.messages, message)
  return nil
}

// Usage in test
mockSMS := &MockSMSProvider{}
manager := NewMFAManager(nil, mockSMS)
```

**What to Mock:**
- External API calls (database, HTTP services)
- Browser APIs (WebSocket, localStorage, matchMedia)
- Child components (in integration tests to isolate component under test)

**What NOT to Mock:**
- Core business logic being tested
- Internal store state management (test real store behavior)
- Internal component event handlers

## Fixtures and Factories

**Test Data:**
```typescript
// From frontend/src/test/utils.ts
export const mockUser: User = {
  id: 1,
  username: 'testuser',
  email: 'test@example.com',
  full_name: 'Test User',
  role: 'user',
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

export const createMockApiClient = (overrides = {}) => ({
  login: vi.fn(),
  signup: vi.fn(),
  logout: vi.fn(),
  registerCollector: vi.fn(),
  testConnection: vi.fn(),
  ...overrides,
})

export const mockAuthResponse: AuthResponse = {
  token: 'mock-token-12345',
  refresh_token: 'mock-refresh-token-12345',
  expires_at: '2024-12-31T00:00:00Z',
  user: mockUser,
}
```

**Location:**
- Frontend: `frontend/src/test/utils.ts` - centralized test utilities and fixtures
- Backend: test files include inline mock structs or separate `tests/mocks/` directory

## Coverage

**Requirements:** Not enforced in CI but available

**View Coverage:**
```bash
npm run test:coverage
# Generates HTML report in coverage/ directory
```

**Configured in `vite.config.ts`:**
```javascript
test: {
  globals: true,
  environment: 'jsdom',
  setupFiles: ['./src/test/setup.ts'],
  coverage: {
    provider: 'v8',
    reporter: ['text', 'json', 'html'],
    exclude: [
      'node_modules/',
      'src/test/',
      '**/*.d.ts',
      '**/*.config.*',
      '**/mockData.ts',
    ]
  }
}
```

## Test Types

**Unit Tests:**
- Scope: Individual functions, hooks, stores
- Approach: Test behavior in isolation with mocked dependencies
- Example: `realtimeStore.test.ts` tests store methods (`setConnected()`, `subscribe()`, `unsubscribe()`) with state assertions
- Example: `circuit_breaker_test.go` tests state transitions without external calls

**Integration Tests:**
- Scope: Component interactions, store + components, multiple modules together
- Approach: Mock external services (API, WebSocket) but test component rendering and state changes
- Location: `backend/tests/integration/`
- Example: `handlers_test.go` tests HTTP handler behavior with mocked database
- Example: `App.test.tsx` tests app-level routing, authentication, and realtime client initialization

**E2E Tests:**
- Not formally implemented in current test structure
- Load tests exist in `backend/tests/load/` for stress testing

## Common Patterns

**Async Testing:**
```typescript
// Using waitFor for state changes
await waitFor(() => {
  expect(realtimeClient.connect).toHaveBeenCalledWith(token)
})

// Using resolvedValue and rejectedValue
vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)
vi.mocked(realtimeClient.connect).mockRejectedValueOnce(new Error('Failed'))

// Running pending timers
vi.advanceTimersByTime(1000)
await vi.runOnlyPendingTimersAsync()
```

**Error Testing:**
```typescript
// Testing error handling in components
it('should handle connection errors gracefully', async () => {
  const token = 'test-token-12345'
  const errorMessage = 'WebSocket connection failed'
  localStorage.setItem('auth_token', token)
  useAuthStore.getState().setToken(token)

  const connectError = new Error(errorMessage)
  vi.mocked(realtimeClient.connect).mockRejectedValueOnce(connectError)

  render(<App />)

  await waitFor(() => {
    expect(useRealtimeStore.getState().error).toBe(errorMessage)
  })
})

// Go testing error cases
func TestCircuitBreakerOpenOnFailures(t *testing.T) {
  cb := ml.NewCircuitBreaker(logger)

  for i := 0; i < 5; i++ {
    cb.RecordFailure()
  }

  if cb.State() != "open" {
    t.Errorf("Expected state open after 5 failures, got %s", cb.State())
  }
}
```

**Test Setup/Teardown:**
```typescript
// Frontend: store and state reset
beforeEach(() => {
  vi.clearAllMocks()
  useAuthStore.getState().reset()
  useRealtimeStore.getState().clear()
  localStorage.clear()
})

afterEach(() => {
  vi.clearAllMocks()
})
```

**Test Utilities:**
From `frontend/src/test/setup.ts`:
- Global cleanup: `afterEach(() => { cleanup() })`
- localStorage mock implementation
- matchMedia mock for responsive design testing
- @testing-library/jest-dom matchers loaded

---

*Testing analysis: 2026-03-30*
