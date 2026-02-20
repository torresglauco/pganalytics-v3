# Integration Tests - Phase 3.4b

## Overview

This directory contains integration tests for pgAnalytics Phase 3.4b (Integration Testing). These tests verify the full metrics collection and transmission pipeline end-to-end, with focus on HTTP communication, TLS/mTLS validation, and collector plugin integration.

**Status**: Milestone 1 Complete - Infrastructure Created
**Test Files Created**: 5 test files + 3 infrastructure files
**Planned Tests**: 50-70 integration tests across all files

---

## Directory Structure

```
integration/
├── README.md (this file)
│
├── Infrastructure Files
│   ├── mock_backend_server.h       (200+ lines) - Mock backend server declaration
│   ├── mock_backend_server.cpp     (400+ lines) - HTTP server implementation
│   ├── fixtures.h                  (300+ lines) - Test data and fixtures
│   └── test_certificates.h         (TBD)        - mTLS certificate fixtures
│
├── Test Files
│   ├── sender_integration_test.cpp     (15-20 tests) - HTTP client tests
│   ├── collector_flow_test.cpp         (15-20 tests) - End-to-end pipeline
│   ├── auth_integration_test.cpp       (10-15 tests) - JWT + mTLS auth
│   ├── config_integration_test.cpp     (8-12 tests)  - Configuration loading
│   └── error_handling_test.cpp         (12-18 tests) - Error scenarios
│
└── certs/ (generated at runtime)
    ├── server.crt                  - Self-signed server certificate
    ├── server.key                  - Server private key
    ├── client.crt                  - Client certificate for mTLS
    └── client.key                  - Client private key
```

---

## Infrastructure Components

### 1. Mock Backend Server (`mock_backend_server.h/cpp`)

Simulates the pgAnalytics backend API for testing collector communication.

**Key Features**:
- HTTP/HTTPS server on localhost:8443
- Accepts POST `/api/v1/metrics/push` (gzipped JSON metrics)
- Accepts POST `/api/v1/collectors/register` (collector registration)
- Accepts GET `/api/v1/config/{collector_id}` (configuration pull)
- JWT token validation in Authorization header
- gzip decompression of metric payloads
- Configurable response scenarios (200, 201, 400, 401, 500)
- Thread-safe request tracking for test assertions

**Main Class Methods**:

```cpp
// Control server lifecycle
bool start();                          // Start server in background thread
bool stop();                           // Stop the server
bool isRunning() const;                // Check if running

// Configure test scenarios
void setNextResponseStatus(int status);           // Return custom HTTP status
void setTokenValid(bool valid);                   // Enable/disable JWT validation
void setResponseDelay(int milliseconds);          // Simulate network latency
void setRejectMetricsWithError(const std::string& error);  // Return error

// Query received data (for test assertions)
int getReceivedMetricsCount() const;              // Count of metric payloads
json getLastReceivedMetrics() const;              // Last metrics payload
std::vector<json> getAllReceivedMetrics() const;  // All metrics payloads
std::vector<std::string> getAllReceivedTokens() const;  // Tokens sent
bool wasEndpointAccessed(const std::string& endpoint) const;  // Endpoint access
std::string getLastAuthorizationHeader() const;   // Last auth header
bool wasLastPayloadGzipped() const;               // Check compression
std::string getBaseUrl() const;                   // Get server base URL
```

**Implementation Details**:
- Socket-based HTTP server (no external HTTP library dependency)
- Non-blocking socket with timeout handling
- Parses HTTP request method, path, headers, body
- Extracts Authorization header and validates JWT format
- Decompresses gzip using zlib's inflate() function
- Thread-safe metrics/token tracking using std::mutex
- Proper HTTP response formatting with headers and Content-Length

**Example Usage**:
```cpp
MockBackendServer server(8443);
ASSERT_TRUE(server.start());

// Configure test scenario
server.setNextResponseStatus(401);  // Simulate auth failure

// Send request (collector code)
// ...

// Verify behavior
EXPECT_EQ(server.getLastResponseStatus(), 401);
EXPECT_EQ(server.getReceivedMetricsCount(), 0);

server.stop();
```

### 2. Test Fixtures (`fixtures.h`)

Reusable test data to reduce code duplication across 50+ tests.

**Configuration Fixtures** (returns std::string):
- `getBasicConfigToml()` - Minimal valid TOML with all required sections
- `getFullConfigToml()` - Complete config with optional fields enabled
- `getNoTlsConfigToml()` - Configuration without TLS
- `getInvalidConfigToml()` - Malformed TOML for error testing

**Metric Payload Fixtures** (returns JSON):
- `getPgStatsMetric()` - PostgreSQL statistics (tables, indexes, databases)
- `getSysstatMetric()` - System stats (CPU, memory, disk IO)
- `getPgLogMetric()` - PostgreSQL log entries
- `getDiskUsageMetric()` - Filesystem usage per mount point
- `getBasicMetricsPayload()` - Complete payload with all 4 metric types
- `getLargeMetricsPayload()` - 400 metrics for load testing (10 MB+)
- `getInvalidMetricsPayload()` - Missing required fields for validation
- `getMultipleMetricsPayload()` - Duplicate metrics for dedup testing

**Helper Functions**:
- `getTestCollectorId()` - Returns "test-collector-001"
- `getTestHostname()` - Returns "test-host"
- `getTestJwtToken()` - Returns valid JWT token string
- `getTestExpiredJwtToken()` - Returns expired JWT token
- `getCurrentTimestamp()` - Returns ISO8601 timestamp

**Example Usage**:
```cpp
auto config = fixtures::getBasicConfigToml();
auto payload = fixtures::getBasicMetricsPayload();
auto invalid = fixtures::getInvalidMetricsPayload();
```

### 3. Test Certificates (TBD)

Will contain self-signed certificates for mTLS testing:
- Server certificate and key
- Client certificate and key (for collector mTLS)
- Helper functions to load certificates

---

## Test Files Overview

### 1. sender_integration_test.cpp (15-20 tests)

Tests HTTP client (Sender) communication with mock backend.

**Test Categories**:
1. **Basic Transmission** (5 tests)
   - SendMetricsSuccess: Valid metrics → 200 OK
   - SendMetricsCreated: Valid metrics → 201 Created
   - ValidatePayloadFormat: Gzip compression, headers
   - AuthorizationHeaderPresent: Bearer token in header
   - ContentTypeJson: Content-Type header validation

2. **Token Management** (5 tests)
   - TokenExpiredRetry: 401 → refresh → retry
   - SuccessAfterTokenRefresh: New token works
   - MaxRetriesExceeded: Give up after N failures
   - TokenValidityBuffer: 60-second buffer

3. **Error Handling** (4 tests)
   - MalformedPayload: 400 response
   - ServerError: 500 response
   - ConnectionRefused: Network unavailable
   - RequestTimeout: Timeout after N seconds

4. **TLS Verification** (4 tests)
   - TlsRequired: HTTPS enforced
   - CertificateValidation: Self-signed cert accepted
   - MtlsCertificatePresent: Client cert sent
   - InvalidCertificateRejected: Bad cert fails

5. **Large Payloads** (2+ tests)
   - LargeMetricsTransmission: 10 MB payload
   - CompressionRatio: Compression >40%
   - PartialBufferTransmission: Partial buffer

### 2. collector_flow_test.cpp (15-20 tests)

Tests end-to-end metric collection and transmission pipeline.

**Test Categories**:
1. **Collection Pipeline** (4 tests)
   - CollectAndSerialize: Collect → serialize → validate
   - BufferAppendAndCompress: Buffer and compression
   - PayloadCreation: Payload structure validation
   - PayloadSerialization: Format matches backend

2. **Transmission Flow** (4 tests)
   - CollectAndTransmit: Full pipeline flow
   - MultipleMetricTypes: All 4 types in payload
   - MetricsTimestamps: Timestamp correctness
   - CollectorIdIncluded: Collector ID in payload

3. **Configuration** (4 tests)
   - ConfigLoadAndApply: Config loading
   - EnabledMetricsOnly: Respect enabled/disabled
   - CollectionIntervals: Interval enforcement
   - TlsConfigApplied: TLS settings

4. **Buffer Management** (4 tests)
   - BufferClearAfterSend: Clear on success
   - BufferOverflow: Handle overflow
   - PartialBufferRetain: Retain on failure
   - CompressionEfficiency: Realistic compression

5. **State Transitions** (3 tests)
   - IdleToCollecting: State transition
   - CollectingToTransmitting: Transmission state
   - ErrorRecovery: Recover from errors
   - ConfigReload: Hot reload config

6. **Data Integrity** (3 tests)
   - NoDataLoss: All metrics transmitted
   - NoDataDuplication: No duplicates
   - MetadataPreserved: ID, hostname, version

### 3. auth_integration_test.cpp (10-15 tests)

Tests JWT token and mTLS certificate handling.

**Test Categories**:
1. **Token Generation & Validation** (4 tests)
   - GenerateAndValidateToken: Backend validates
   - TokenSignatureVerified: JWT signature check
   - TokenExpirationEnforced: Expired token rejected
   - TokenPayloadStructure: Correct JWT claims

2. **Token Refresh** (4 tests)
   - TokenRefreshFlow: Refresh works correctly
   - RefreshBuffer: 60-second buffer
   - MultipleRefreshes: Multiple refresh cycles
   - RefreshOnExpiration: Auto-refresh on expiration

3. **Certificate Management** (3 tests)
   - ClientCertificateRequired: mTLS validation
   - CertificateLoadError: Missing cert handling
   - InvalidCertificateFormat: Malformed cert

4. **Authorization Errors** (4 tests)
   - UnauthorizedResponse: 401 handling
   - ForbiddenResponse: 403 handling
   - ExpiredTokenRejected: Expired token rejection
   - InvalidSignatureRejected: Bad signature

### 4. config_integration_test.cpp (8-12 tests)

Tests configuration loading from files and backend.

**Test Categories**:
1. **File Loading** (4 tests)
   - LoadValidConfiguration: Valid config
   - MissingConfigFile: Missing file handling
   - InvalidTomlSyntax: Malformed TOML
   - DefaultValuesApplied: Default fallback

2. **Validation** (4 tests)
   - RequiredFieldsPresent: Required fields check
   - InvalidBackendUrl: URL validation
   - InvalidPostgresqlConfig: DB param validation
   - TlsConfigValidation: Cert path existence

3. **Application** (6 tests)
   - ConfigApplyToCollector: Apply to collector
   - MetricsEnabled: Enable/disable metrics
   - CollectionIntervalsApplied: Interval respect
   - BackendUrlApplied: Backend URL usage
   - TlsSettingsApplied: TLS settings
   - PostgresqlConfigApplied: DB connection

4. **Dynamic Config** (4 tests)
   - ConfigReloadFromBackend: Pull config from backend
   - ConfigVersionTracking: Track config version
   - ConfigHotReload: Apply without restart
   - ConfigChangeNotification: Detect changes

### 5. error_handling_test.cpp (12-18 tests)

Tests error scenarios and recovery mechanisms.

**Test Categories**:
1. **Network Errors** (4 tests)
   - ConnectionRefused: Backend unavailable
   - ConnectionTimeout: Connection timeout
   - RequestTimeout: Request timeout
   - NetworkPartition: Intermittent connectivity

2. **Backend Errors** (4 tests)
   - ServerError500: 500 error handling
   - ServiceUnavailable503: 503 error
   - BadGateway502: 502 error
   - PartialResponse: Incomplete response

3. **Payload Errors** (5 tests)
   - MalformedJson400: Invalid JSON
   - MissingRequiredFields400: Field validation
   - InvalidMetricType400: Unknown metric type
   - SizeLimit413: Payload too large
   - EmptyPayload: Empty metrics array

4. **Retry & Recovery** (5+ tests)
   - ExponentialBackoff: Backoff progression
   - MaxRetriesExceeded: Max retry limit
   - PartialBufferRetained: Metrics retained
   - SuccessfulRecovery: Recovery after failure
   - RecoveryWithoutDataLoss: No data loss

5. **Authentication Errors** (3 tests)
   - TokenExpiredRetry: 401 → refresh → retry
   - AuthenticationFailureAfterRefresh: Refresh fails
   - UnauthorizedAfterRefresh: Still unauthorized

6. **Logging & Edge Cases** (3+ tests)
   - ErrorsLogged: Error logging
   - RetryLogged: Retry logging
   - RapidFailures: Consecutive failures
   - SlowResponses: Slow but successful
   - MixedSuccessAndFailure: Alternating success/failure

---

## Build Configuration

### CMakeLists.txt Changes

Updated `collector/tests/CMakeLists.txt` to include integration tests:

```cmake
# Integration test sources
set(INTEGRATION_TEST_SOURCES
    integration/mock_backend_server.cpp
    integration/sender_integration_test.cpp
    integration/collector_flow_test.cpp
    integration/auth_integration_test.cpp
    integration/config_integration_test.cpp
    integration/error_handling_test.cpp
)

# Include directories
target_include_directories(pganalytics-tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${CMAKE_CURRENT_SOURCE_DIR}/integration    # Add fixtures.h include path
    ${GTEST_INCLUDE_DIRS}
    ${nlohmann_json_INCLUDE_DIR}
)
```

### Required Dependencies

All dependencies already available from Phase 3.4a:
- Google Test (gtest)
- OpenSSL 3.0+
- libcurl
- zlib
- nlohmann/json

No new dependencies introduced.

---

## Test Execution

### Run All Tests

```bash
cd collector/build
cmake .. -DBUILD_TESTS=ON
make -j4

# Run all unit + integration tests
./tests/pganalytics-tests

# Or using ctest
ctest --verbose
```

### Run Only Integration Tests

```bash
./tests/pganalytics-tests --gtest_filter="*Integration*"
```

### Run Specific Test Suite

```bash
# Sender tests only
./tests/pganalytics-tests --gtest_filter="SenderIntegrationTest.*"

# Collector flow tests only
./tests/pganalytics-tests --gtest_filter="CollectorFlowTest.*"

# Auth tests only
./tests/pganalytics-tests --gtest_filter="AuthIntegrationTest.*"
```

### Run Specific Test Case

```bash
./tests/pganalytics-tests --gtest_filter="SenderIntegrationTest.SendMetricsSuccess"
```

### Generate Test Report

```bash
./tests/pganalytics-tests --gtest_output="xml:test-results.xml"
```

---

## Test Data Usage

### Example: Using Fixtures in Tests

```cpp
TEST_F(SenderIntegrationTest, SendMetricsSuccess) {
    // Load test data
    auto payload = fixtures::getBasicMetricsPayload();
    auto config = fixtures::getBasicConfigToml();

    // Create sender and push metrics
    Sender sender("test-collector-001", mock_server.getBaseUrl());
    sender.setAuthToken(fixtures::getTestJwtToken(),
                       std::time(nullptr) + 3600);

    // Send metrics (mock_server handles the request)
    bool success = sender.pushMetrics(payload);

    // Assert with mock server state
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
    EXPECT_EQ(mock_server.getLastResponseStatus(), 200);
}
```

---

## Expected Test Results

**Current Status**: Milestone 1 Infrastructure Complete

**Test Files**: 5 created (placeholder implementations)
- sender_integration_test.cpp: 20 test cases (TODO implementations)
- collector_flow_test.cpp: 20 test cases (TODO implementations)
- auth_integration_test.cpp: 15 test cases (TODO implementations)
- config_integration_test.cpp: 12 test cases (TODO implementations)
- error_handling_test.cpp: 18 test cases (TODO implementations)

**Total Planned Tests**: 85 test cases across 5 files

**Infrastructure Files**: 3 completed
- mock_backend_server.h/cpp: Full HTTP server implementation
- fixtures.h: Complete test data fixtures

**Next Steps** (Milestone 2-5):
1. Implement test case bodies (currently contain TODO comments)
2. Add mTLS certificate generation (test_certificates.h)
3. Run compilation to verify all dependencies
4. Execute tests with actual implementations

---

## Development Notes

### Fixtures Pattern

The fixtures namespace provides reusable test data, reducing boilerplate:
- Configuration files are generated in-memory (no file I/O needed)
- Metric payloads are pre-defined JSON structures
- Helper functions return consistent test values
- All fixtures use inline functions for immediate availability

### Mock Server Architecture

The MockBackendServer is designed for simplicity and isolation:
- Socket-based (no heavy HTTP library dependency)
- Single-threaded request handler in background thread
- Non-blocking accept() to allow clean shutdown
- Thread-safe state via mutex-protected vectors
- Gzip decompression supports real-world compression validation

### Test Isolation

Each test class has:
- Independent mock server instance
- SetUp() starts server before test
- TearDown() stops server after test
- No shared state between tests
- Clean resource cleanup (no leaks)

---

## Troubleshooting

### "gtest/gtest.h" file not found

This is expected during IDE analysis. CMake build resolves paths correctly.

**Solution**: Run `cmake` configuration before building:
```bash
cd collector/build
cmake .. -DBUILD_TESTS=ON
```

### Mock Server Already in Use

If tests fail with "port 8443 in use":

**Solution**: Ensure previous test instances stopped:
```bash
# Kill any stray processes
killall pganalytics-tests 2>/dev/null

# Try different port in test
MockBackendServer server(18443);  // Use different port
```

### gzip Decompression Failures

If mock server can't decompress payloads:

**Solution**: Verify payload is actually gzipped:
- Check for magic number `0x1f 0x8b` at start of buffer
- Verify sender uses `Content-Encoding: gzip` header
- Check sender actually compresses data before sending

---

## Future Enhancements

### Milestone 2: Complete Test Implementations
- Implement test case bodies
- Add actual collector and sender integration
- Verify real HTTP communication

### Milestone 3: mTLS Certificates
- Generate self-signed certs for testing
- Add certificate validation tests
- Test certificate renewal flow

### Milestone 4: Performance Tests
- Load test with 100+ concurrent collectors
- Measure latency distributions
- Verify memory stability

### Milestone 5: Documentation
- API usage examples
- Troubleshooting guide
- Performance benchmarks

---

## References

- [Phase 3.4b Plan](../../../docs/PHASE_3_4B_PLAN.md)
- [Unit Tests Report](../../../UNIT_TESTS_EXECUTION_REPORT.md)
- [Backend API Specification](../../../docs/API.md)
- [Collector Architecture](../../../docs/COLLECTOR-ARCHITECTURE.md)

---

**Status**: ✅ Milestone 1 Complete - Infrastructure Created
**Created**: 2026-02-19
**Phase**: 3.4b (Integration Testing)
**Goal**: 50-70 integration tests for full pipeline validation
