# Phase 3.4b Milestone 2 - Integration Tests Implementation Plan

**Phase**: 3.4b (Integration Testing)
**Milestone**: 2 (Test Case Implementation)
**Status**: 🚀 STARTING
**Date**: February 19, 2026

---

## Overview

Milestone 2 focuses on implementing the 85+ test case bodies that were created as placeholders in Milestone 1. This involves replacing TODO comments with actual test logic, integrating with real Collector and Sender components, and verifying that the mock backend server works correctly.

**Infrastructure Status**: ✅ Complete (Milestone 1)
**Test Framework**: ✅ Ready (112 test classes registered)
**Build System**: ✅ Verified (all tests compiling)
**Performance**: ✅ Baseline (218 ms avg per test)

---

## Milestone 2 Scope

### Test Implementation by Category

**1. SenderIntegrationTest (20 tests)**
- Currently: 20 test cases with TODO placeholder bodies
- Goal: Implement with actual HTTP client testing
- Dependencies: Sender class, mock_backend_server
- Estimated complexity: Medium

**2. CollectorFlowTest (23 tests)**
- Currently: 23 test cases with TODO placeholder bodies
- Goal: Implement end-to-end pipeline tests
- Dependencies: Collector, Sender, ConfigManager, MetricsBuffer
- Estimated complexity: High

**3. AuthIntegrationTest (19 tests)**
- Currently: 19 test cases with TODO placeholder bodies
- Goal: Implement JWT and mTLS authentication tests
- Dependencies: AuthManager, mTLS certificates
- Estimated complexity: Medium-High

**4. ConfigIntegrationTest (22 tests)**
- Currently: 22 test cases with TODO placeholder bodies
- Goal: Implement configuration loading and application tests
- Dependencies: ConfigManager, backend config pull
- Estimated complexity: Medium

**5. ErrorHandlingTest (28 tests)**
- Currently: 28 test cases with TODO placeholder bodies
- Goal: Implement error scenarios and recovery testing
- Dependencies: Mock server error simulation, retry logic
- Estimated complexity: High

### Total Test Cases to Implement: 112

---

## Implementation Strategy

### Phase 1: Sender Integration Tests (Days 1-2)

**Approach:**
1. Implement basic transmission tests first (foundation)
2. Build token management tests
3. Implement error handling
4. Add TLS verification
5. Test large payloads

**Key Files:**
- `collector/tests/integration/sender_integration_test.cpp`
- Use: `Sender` class, `fixtures.h` data
- Mock: `MockBackendServer` for HTTP endpoints

**Test Structure Example:**
```cpp
TEST_F(SenderIntegrationTest, SendMetricsSuccess) {
    // Arrange: Create Sender and metrics
    Sender sender("test-collector-001", mock_server.getBaseUrl());
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Push metrics to mock server
    bool success = sender.pushMetrics(payload);

    // Assert: Verify response
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
    EXPECT_EQ(mock_server.getLastResponseStatus(), 200);
}
```

### Phase 2: Collector Flow Tests (Days 3-5)

**Approach:**
1. Implement collection pipeline tests
2. Add transmission flow tests
3. Implement configuration application tests
4. Test buffer management
5. Verify state transitions

**Key Files:**
- `collector/tests/integration/collector_flow_test.cpp`
- Use: `Collector`, `MetricsBuffer`, `ConfigManager`, `Sender`
- Dependencies: Mock PostgreSQL or fixture data

**Complexity:**
- May need to mock PostgreSQL connection
- Requires understanding of collector plugin architecture
- Tests full pipeline from collection to transmission

### Phase 3: Auth Integration Tests (Days 6-7)

**Approach:**
1. Implement token generation and validation tests
2. Add token refresh scenario tests
3. Implement certificate management tests
4. Test authorization error handling

**Key Files:**
- `collector/tests/integration/auth_integration_test.cpp`
- Use: `AuthManager`, `MockBackendServer`
- Need: mTLS certificates (self-signed)

**Dependencies:**
- May need to create test_certificates.h file
- Generate self-signed certificates for testing

### Phase 4: Config Integration Tests (Days 8-9)

**Approach:**
1. Implement file loading tests
2. Add validation tests
3. Test configuration application
4. Implement dynamic reload tests

**Key Files:**
- `collector/tests/integration/config_integration_test.cpp`
- Use: `ConfigManager`, `Sender`, `MockBackendServer`
- Test data: Fixture configurations

### Phase 5: Error Handling Tests (Days 10-12)

**Approach:**
1. Implement network error tests
2. Add backend error response tests
3. Test payload error scenarios
4. Implement retry and recovery tests
5. Test auth error handling

**Key Files:**
- `collector/tests/integration/error_handling_test.cpp`
- Use: Mock server error simulation features
- Complex: Retry logic, exponential backoff validation

---

## Implementation Guidelines

### General Test Pattern

Each test should follow this pattern:

```cpp
TEST_F(TestSuite, TestCase) {
    // 1. ARRANGE - Set up test data and mock responses
    mock_server.setNextResponseStatus(200);
    auto test_data = fixtures::getSomeFixture();
    auto expected_value = /* calculated value */;

    // 2. ACT - Perform the action being tested
    auto result = component->doSomething(test_data);

    // 3. ASSERT - Verify the outcome
    EXPECT_EQ(result, expected_value);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}
```

### Mock Server Usage

All tests have access to `MockBackendServer`:

```cpp
// Configure mock server behavior
mock_server.setNextResponseStatus(401);        // Return 401 error
mock_server.setTokenValid(false);              // Reject tokens
mock_server.setResponseDelay(1000);            // 1 second delay
mock_server.setRejectMetricsWithError("error"); // Custom error

// Query results after test
int count = mock_server.getReceivedMetricsCount();
json metrics = mock_server.getLastReceivedMetrics();
bool gzipped = mock_server.wasLastPayloadGzipped();
std::string auth = mock_server.getLastAuthorizationHeader();
```

### Test Data Usage

Use fixtures for consistent test data:

```cpp
// Configurations
auto config = fixtures::getBasicConfigToml();

// Metrics
auto payload = fixtures::getBasicMetricsPayload();
auto pg_stats = fixtures::getPgStatsMetric();
auto invalid = fixtures::getInvalidMetricsPayload();

// Helper values
auto collector_id = fixtures::getTestCollectorId();
auto token = fixtures::getTestJwtToken();
auto timestamp = fixtures::getCurrentTimestamp();
```

### Assertions & Expectations

Use meaningful assertions:

```cpp
// HTTP status verification
EXPECT_EQ(mock_server.getLastResponseStatus(), 200);

// Metrics received
EXPECT_EQ(mock_server.getReceivedMetricsCount(), expected_count);

// Authentication
EXPECT_TRUE(mock_server.getLastAuthorizationHeader().find("Bearer") != std::string::npos);

// Compression
EXPECT_TRUE(mock_server.wasLastPayloadGzipped());

// JSON structure
auto received = mock_server.getLastReceivedMetrics();
EXPECT_TRUE(received.contains("collector_id"));
EXPECT_EQ(received["collector_id"], "test-collector-001");
```

---

## Dependencies & Integration Points

### Collector Components (Required)

1. **Sender** (`sender.h/cpp`)
   - `pushMetrics(json payload)` - Push to backend
   - `setAuthToken(token, expiration)` - Set JWT token
   - Used in: SenderIntegrationTest, CollectorFlowTest

2. **AuthManager** (`auth.h/cpp`)
   - `generateToken(expiresIn)` - Create JWT
   - `validateTokenSignature(token)` - Verify JWT
   - `loadClientCertificate(path)` - Load mTLS cert
   - Used in: AuthIntegrationTest

3. **ConfigManager** (`config_manager.h/cpp`)
   - `loadFromFile(path)` - Load TOML
   - `getCollectorId()` - Get collector ID
   - `getBackendUrl()` - Get backend URL
   - Used in: ConfigIntegrationTest, CollectorFlowTest

4. **MetricsBuffer** (`metrics_buffer.h/cpp`)
   - `append(metric)` - Add metric
   - `getCompressedData()` - Get gzip data
   - `clear()` - Clear buffer
   - Used in: CollectorFlowTest, SenderIntegrationTest

5. **MetricsSerializer** (`metrics_serializer.h/cpp`)
   - `serialize(data)` - Convert to JSON
   - `validate(json)` - Validate schema
   - Used in: CollectorFlowTest

6. **Collector** (`collector.h/cpp`)
   - `collectMetrics()` - Gather metrics
   - `configure(config)` - Apply config
   - Used in: CollectorFlowTest

### Mock Infrastructure (Built in M1)

- **MockBackendServer** - HTTP server simulation ✅
- **fixtures.h** - Test data ✅
- **CMakeLists.txt** - Build configuration ✅

### To Create (M2)

- **test_certificates.h** (optional) - mTLS certificates for testing
  - Can use fixtures to generate or include self-signed certs
  - Used in: AuthIntegrationTest

---

## Test Implementation Priority

### High Priority (Do First)
1. SenderIntegrationTest.SendMetricsSuccess - Basic happy path
2. SenderIntegrationTest.TokenExpiredRetry - Token refresh
3. ConfigIntegrationTest.LoadValidConfiguration - Config loading
4. ErrorHandlingTest.ServerError500 - Error handling baseline

### Medium Priority (Do Next)
1. CollectorFlowTest.CollectAndTransmit - Full pipeline
2. AuthIntegrationTest.GenerateAndValidateToken - Auth baseline
3. ErrorHandlingTest.ConnectionRefused - Network errors
4. SenderIntegrationTest.LargeMetricsTransmission - Large payloads

### Lower Priority (Do Last)
1. Timing-sensitive tests (token refresh timing)
2. Complex multi-step scenarios
3. Edge case tests
4. Performance validation tests

---

## Success Criteria for Milestone 2

### Code Quality
- [ ] All 112 test cases have implementation (no TODO comments remain)
- [ ] Clear, readable test code
- [ ] Proper use of fixtures and mock server
- [ ] Meaningful assertions in each test
- [ ] No magic numbers or hardcoded values

### Functionality
- [ ] All 112 tests compile without errors
- [ ] All 112 tests execute
- [ ] Target: 100% pass rate (or identify real bugs)
- [ ] No memory leaks
- [ ] Proper resource cleanup in all tests

### Integration
- [ ] Tests properly integrate with real Collector components
- [ ] Mock server handles all test scenarios correctly
- [ ] Configuration fixtures work as expected
- [ ] Error handling properly tested

### Documentation
- [ ] Test code is self-documenting
- [ ] Comments explain non-obvious logic
- [ ] Test names clearly describe what's being tested
- [ ] README updated with implementation status

### Performance
- [ ] Average test time < 250 ms (baseline was 218 ms)
- [ ] No significant performance regressions
- [ ] Build time reasonable (~2 minutes)
- [ ] Memory usage stable

---

## Implementation Checklist

### Phase 1: Sender Tests
- [ ] SendMetricsSuccess
- [ ] SendMetricsCreated
- [ ] ValidatePayloadFormat
- [ ] AuthorizationHeaderPresent
- [ ] ContentTypeJson
- [ ] TokenExpiredRetry
- [ ] SuccessAfterTokenRefresh
- [ ] MaxRetriesExceeded
- [ ] TokenValidityBuffer
- [ ] MalformedPayload
- [ ] ServerError
- [ ] ConnectionRefused
- [ ] RequestTimeout
- [ ] TlsRequired
- [ ] CertificateValidation
- [ ] MtlsCertificatePresent
- [ ] InvalidCertificateRejected
- [ ] LargeMetricsTransmission
- [ ] CompressionRatio
- [ ] PartialBufferTransmission

### Phase 2: Collector Flow Tests
- [ ] CollectAndSerialize
- [ ] BufferAppendAndCompress
- [ ] PayloadCreation
- [ ] PayloadSerialization
- [ ] CollectAndTransmit
- [ ] MultipleMetricTypes
- [ ] MetricsTimestamps
- [ ] CollectorIdIncluded
- [ ] ConfigLoadAndApply
- [ ] EnabledMetricsOnly
- [ ] CollectionIntervals
- [ ] TlsConfigApplied
- [ ] BufferClearAfterSend
- [ ] BufferOverflow
- [ ] PartialBufferRetain
- [ ] CompressionEfficiency
- [ ] IdleToCollecting
- [ ] CollectingToTransmitting
- [ ] ErrorRecovery
- [ ] ConfigReload
- [ ] NoDataLoss
- [ ] NoDataDuplication
- [ ] MetadataPreserved

### Phase 3: Auth Tests
- [ ] GenerateAndValidateToken
- [ ] TokenSignatureVerified
- [ ] TokenExpirationEnforced
- [ ] TokenPayloadStructure
- [ ] TokenRefreshFlow
- [ ] RefreshBuffer
- [ ] MultipleRefreshes
- [ ] RefreshOnExpiration
- [ ] ClientCertificateRequired
- [ ] CertificateLoadError
- [ ] InvalidCertificateFormat
- [ ] UnauthorizedResponse
- [ ] ForbiddenResponse
- [ ] ExpiredTokenRejected
- [ ] InvalidSignatureRejected
- [ ] TokenCaching
- [ ] TokenExpirationTime
- [ ] MultipleAuthManagers
- [ ] TokenValidityCheck

### Phase 4: Config Tests
- [ ] LoadValidConfiguration
- [ ] MissingConfigFile
- [ ] InvalidTomlSyntax
- [ ] DefaultValuesApplied
- [ ] RequiredFieldsPresent
- [ ] InvalidBackendUrl
- [ ] InvalidPostgresqlConfig
- [ ] TlsConfigValidation
- [ ] ConfigApplyToCollector
- [ ] MetricsEnabled
- [ ] CollectionIntervalsApplied
- [ ] BackendUrlApplied
- [ ] TlsSettingsApplied
- [ ] PostgresqlConfigApplied
- [ ] ConfigReloadFromBackend
- [ ] ConfigVersionTracking
- [ ] ConfigHotReload
- [ ] ConfigChangeNotification
- [ ] ConfigurationPersistence
- [ ] MultipleSections
- [ ] SpecialCharactersInValues
- [ ] CaseSensitivity

### Phase 5: Error Handling Tests
- [ ] ConnectionRefused
- [ ] ConnectionTimeout
- [ ] RequestTimeout
- [ ] NetworkPartition
- [ ] ServerError500
- [ ] ServiceUnavailable503
- [ ] BadGateway502
- [ ] PartialResponse
- [ ] MalformedJson400
- [ ] MissingRequiredFields400
- [ ] InvalidMetricType400
- [ ] SizeLimit413
- [ ] EmptyPayload
- [ ] ExponentialBackoff
- [ ] MaxRetriesExceeded
- [ ] PartialBufferRetained
- [ ] SuccessfulRecovery
- [ ] RecoveryWithoutDataLoss
- [ ] CircuitBreakerPattern
- [ ] TokenExpiredRetry
- [ ] AuthenticationFailureAfterRefresh
- [ ] UnauthorizedAfterRefresh
- [ ] ErrorsLogged
- [ ] RetryLogged
- [ ] RecoveryLogged
- [ ] RapidFailures
- [ ] SlowResponses
- [ ] MixedSuccessAndFailure

---

## Build & Test Workflow

### During Implementation

```bash
# Build with tests enabled
cd collector/build
cmake .. -DBUILD_TESTS=ON
make -j4

# Run all tests
./tests/pganalytics-tests

# Run specific test suite while implementing
./tests/pganalytics-tests --gtest_filter="SenderIntegrationTest.*"

# Run single test for debugging
./tests/pganalytics-tests --gtest_filter="SenderIntegrationTest.SendMetricsSuccess" -v
```

### Progress Tracking

- Track which tests are implemented (✓)
- Track which tests are passing (✓)
- Track which tests have known failures (with reason)
- Update checklist regularly

---

## Known Considerations

### Mock Server Limitations
- Socket-based HTTP (no advanced features)
- Simple request parsing (may not handle all edge cases)
- Single-threaded request handling (but thread-safe state)
- No SSL/TLS actual implementation (test mode)

### Test Isolation
- Each test gets fresh mock server instance
- SetUp/TearDown handles server lifecycle
- No shared state between tests
- Clean resource cleanup required

### Timing Issues
- Some tests may be timing-sensitive (token expiration)
- Avoid absolute timing assertions where possible
- Use relative comparisons (before/after expiration)

### Dependencies
- Tests depend on Collector components being properly implemented
- Mock server is fully implemented and verified
- Fixtures are complete and working
- CMakeLists.txt properly configured

---

## Next Steps After M2

### Upon M2 Completion
1. Verify all 112 tests compile and execute
2. Review code coverage (target: >80% on new tests)
3. Performance validation (average < 250ms per test)
4. Documentation update

### Milestone 3: mTLS & Certificates
- Create test_certificates.h file
- Generate self-signed certificates
- Add certificate validation tests
- Verify mTLS handshake

### Milestone 4: Final Validation
- Full test suite execution
- 100% pass rate achievement
- Performance optimization
- Documentation finalization

### Milestone 5: Release Prep
- Code review and cleanup
- Final documentation
- Performance benchmarks
- Release notes

---

## Timeline Estimate

- **Phase 1 (Sender Tests)**: 2 days
- **Phase 2 (Collector Flow Tests)**: 3 days
- **Phase 3 (Auth Tests)**: 2 days
- **Phase 4 (Config Tests)**: 1.5 days
- **Phase 5 (Error Handling)**: 2.5 days
- **Buffer for fixes/adjustments**: 1 day

**Total Estimated Timeline**: 12 days (about 2 weeks)

---

## Success Definition

✅ **Milestone 2 will be complete when:**

1. All 112 test cases have implementation (no TODO comments)
2. All 112 tests compile without errors
3. All 112 tests execute successfully
4. Target: 100% pass rate (or identify and document real bugs)
5. Proper integration with Collector components
6. Mock server handles all test scenarios
7. No memory leaks or resource cleanup issues
8. Code is well-structured and maintainable
9. Performance is acceptable (< 250ms average)
10. Documentation is updated

**Status**: 🚀 Ready to begin Phase 1 implementation

---

**Last Updated**: February 19, 2026
**Phase**: 3.4b Milestone 2 (Implementation)
**Status**: STARTING
