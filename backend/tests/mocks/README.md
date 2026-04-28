# Mock Libraries for Integration Testing

This directory contains mock implementations for external dependencies used in integration tests.

## Available Mocks

### MockMLService

**Location:** `ml_service_mock.go`

**Purpose:** Mock HTTP server simulating the ML prediction service.

**Usage:**
```go
func TestMyFeature(t *testing.T) {
    mockML := NewMockMLService()
    defer mockML.Close()

    // Configure behavior
    mockML.SetShouldFail(true)  // Simulate service failure
    mockML.SetResponseDelay(100 * time.Millisecond)  // Add latency

    // Use URL in configuration
    config.MLServiceURL = mockML.URL()

    // Run tests...
}
```

**Methods:**
- `NewMockMLService() *MockMLService` - Create new mock server
- `URL() string` - Get mock server URL
- `SetShouldFail(bool)` - Configure failure simulation
- `SetHTTPStatusCode(int)` - Set custom HTTP status code for responses
- `SetResponseDelay(time.Duration)` - Add artificial delay
- `GetTrainingJob(jobID string) *TrainingJob` - Retrieve a training job by ID
- `GetPrediction(queryHash int64) *PredictionResponse` - Retrieve a prediction by query hash
- `GetRequestCount() int` - Get the number of requests made
- `Close()` - Shutdown mock server

**Endpoints Mocked:**
- `/api/health` - Health check endpoint
- `/api/train/performance-model` - Training endpoint
- `/api/train/performance-model/{jobID}` - Training status endpoint
- `/api/predict/query-execution` - Prediction endpoint
- `/api/validate/prediction` - Validation endpoint
- `/api/detect/patterns` - Pattern detection endpoint

---

### TestUserStore

**Location:** `../integration/handlers_test.go`

**Purpose:** In-memory user storage for authentication tests.

**Usage:**
```go
func TestAuth(t *testing.T) {
    userStore := NewTestUserStore()
    // Default user: "testuser" / "password123" with role "user"

    // Add custom users
    userStore.users["admin"] = &models.User{
        Username: "admin",
        Role:     "admin",
        // ...
    }
}
```

**Default Data:**
- Username: `testuser`
- Password: `password123`
- Role: `user`
- Email: `test@example.com`

**Methods:**
- `NewTestUserStore() *TestUserStore` - Create store with default test user
- `GetUserByUsername(username string) (*models.User, error)` - Lookup by username
- `GetUserByID(id int) (*models.User, error)` - Lookup by ID
- `UpdateUserLastLogin(userID int, timestamp time.Time) error` - Update last login

---

### TestCollectorStore

**Location:** `../integration/handlers_test.go`

**Purpose:** In-memory collector storage for collector endpoint tests.

**Usage:**
```go
func TestCollectors(t *testing.T) {
    collectorStore := NewTestCollectorStore()
    // Empty by default

    // Create collectors using the API or manually
    collector := &models.Collector{
        ID:       uuid.New(),
        Name:     "test-collector",
        Hostname: "db-server-01",
    }
    collectorStore.CreateCollector(collector)
}
```

**Methods:**
- `NewTestCollectorStore() *TestCollectorStore` - Create empty store
- `CreateCollector(collector *models.Collector) (uuid.UUID, error)` - Add new collector
- `GetCollectorByID(id uuid.UUID) (*models.Collector, error)` - Lookup by UUID
- `UpdateCollectorStatus(id uuid.UUID, status string) error` - Update status
- `UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error` - Update certificate

---

### TestTokenStore

**Location:** `../integration/handlers_test.go`

**Purpose:** In-memory API token storage for token validation tests.

**Usage:**
```go
func TestTokens(t *testing.T) {
    tokenStore := NewTestTokenStore()
    // Configure tokens as needed

    token := &models.APIToken{
        TokenHash: "hash-of-token",
        Name:      "test-token",
    }
    tokenStore.CreateAPIToken(token)
}
```

**Methods:**
- `NewTestTokenStore() *TestTokenStore` - Create empty store
- `CreateAPIToken(token *models.APIToken) (int, error)` - Add new token
- `GetAPITokenByHash(hash string) (*models.APIToken, error)` - Lookup by hash
- `UpdateAPITokenLastUsed(id int, timestamp time.Time) error` - Update last used

---

## Test Environment Helpers

### newTestEnv

**Location:** `../integration/boundary_test_helpers.go`

Creates a complete test environment with router and all stores.

```go
router, userStore, collectorStore := newTestEnv(t)
```

**Returns:**
- `*gin.Engine` - Router for making test requests
- `*TestUserStore` - User store with default test user
- `*TestCollectorStore` - Empty collector store

### newTestEnvWithEmptyUsers

**Location:** `../integration/boundary_test_helpers.go`

Creates environment without default test user (for registration tests).

```go
router, userStore, collectorStore := newTestEnvWithEmptyUsers(t)
```

**Returns:**
- `*gin.Engine` - Router for making test requests
- `*TestUserStore` - Empty user store (no default user)
- `*TestCollectorStore` - Empty collector store

### createTestServer

**Location:** `../integration/handlers_test.go`

Lower-level helper for creating a test server with custom stores.

```go
server, router := createTestServer(userStore, collectorStore, tokenStore)
```

---

## Additional Test Helpers

### TestDB

**Location:** `../integration/testhelpers.go`

Provides test database connection utilities.

```go
testDB := NewTestDB(t)
defer testDB.Cleanup(t)

db := testDB.GetDB()
```

### QueryHelper

**Location:** `../integration/testhelpers.go`

Utilities for database queries in tests.

```go
qh := NewQueryHelper(db)
results, err := qh.ExecuteQuery(ctx, "SELECT * FROM users")
```

### Mock Data Functions

**Location:** `../integration/testhelpers.go`

- `MockExplainOutput() string` - Returns realistic EXPLAIN ANALYZE output
- `MockExplainOutputComplex() string` - Returns complex EXPLAIN ANALYZE output
- `MockPostgresLogEntries() []map[string]interface{}` - Returns PostgreSQL log entries
- `MockLogEntriesByCategory() map[string][]map[string]interface{}` - Log entries by category

### Assertion Helpers

**Location:** `../integration/testhelpers.go`

- `WaitForCondition(ctx, checkFn, maxWait)` - Wait for condition with timeout
- `AssertWithinDuration(t, actual, expected, tolerance)` - Duration range check
- `AssertTimeRecent(t, ts, maxAge)` - Timestamp freshness check

---

## Best Practices

1. **Always use `t.Helper()`** in custom test helpers to improve error messages
2. **Always `defer Close()`** mock services to prevent resource leaks
3. **Use `assert` from testify** for clear failure messages:
   ```go
   assert.Equal(t, http.StatusOK, w.Code)
   require.NoError(t, err)  // Stops test on failure
   ```
4. **Never use `t.Log(err)` without failing** - use `require.NoError(t, err)` instead
5. **Keep mock data minimal** - add only what the test needs
6. **Use subtests** for table-driven tests:
   ```go
   for name, tc := range tests {
       t.Run(name, func(t *testing.T) {
           // test code
       })
   }
   ```

---

## Common Patterns

### Testing Authentication

```go
func TestProtectedEndpoint(t *testing.T) {
    router, _, _ := newTestEnv(t)

    // First login to get token
    loginReq := models.LoginRequest{
        Username: "testuser",
        Password: "password123",
    }
    body, _ := json.Marshal(loginReq)
    req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    var resp models.LoginResponse
    json.Unmarshal(w.Body.Bytes(), &resp)

    // Use token for protected endpoint
    req = httptest.NewRequest("GET", "/api/v1/protected", nil)
    req.Header.Set("Authorization", "Bearer "+resp.Token)

    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
}
```

### Testing ML Service Failures

```go
func TestMLFailure(t *testing.T) {
    mockML := NewMockMLService()
    defer mockML.Close()

    // Simulate service failure
    mockML.SetShouldFail(true)

    // Or simulate slow response
    mockML.SetResponseDelay(5 * time.Second)

    // Configure your code to use mockML.URL()
}
```

### Testing Collector Registration

```go
func TestCollectorRegistration(t *testing.T) {
    router, _, collectorStore := newTestEnv(t)

    registerReq := models.CollectorRegisterRequest{
        Name:     "test-collector",
        Hostname: "db-server-01",
    }

    body, _ := json.Marshal(registerReq)
    req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Registration-Secret", "test-secret")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Verify collector was created
    assert.Equal(t, http.StatusOK, w.Code)
}
```