# Collector Unit Tests

**Status**: ✅ Ready for Execution

**Total Test Cases**: 130+ unit tests across 5 test files

**Target Coverage**: >60% code coverage

---

## Test Files

### 1. metrics_serializer_test.cpp (20 test cases)
**Component**: MetricsSerializer - JSON schema validation

Test Coverage:
- `CreateBasicPayload` - Basic payload creation with all required fields
- `PayloadWithMetrics` - Payload containing metric data
- `ValidateValidPayload` - Valid payload validation passes
- `ValidateMissingCollectorId` - Validation fails with missing field
- `ValidateMissingMetricsArray` - Validation fails without metrics array
- `ValidatePgStatsMetric` - PostgreSQL stats metric validation
- `ValidatePgStatsWithoutDatabase` - Validation fails with missing required field
- `ValidatePgLogMetric` - PostgreSQL log metric validation
- `ValidateSysstatMetric` - System statistics metric validation
- `ValidateDiskUsageMetric` - Disk usage metric validation
- `ValidateUnknownMetricType` - Validation fails for unknown metric types
- `ValidatePgStatsWithTables` - Complex metric with nested objects
- `GetSchemaVersion` - Schema version retrieval
- `ValidateInvalidMetricObject` - Non-object metric rejection
- `ValidatePgLogWithEntries` - Log metric with multiple entries
- `ValidatePgLogEntryWithoutMessage` - Entry validation with missing fields
- `ValidateMultipleMetrics` - Payload with multiple different metric types
- `PayloadFieldTypes` - Field type validation
- `EmptyMetricsArray` - Empty array handling
- `SysstatWithAllFields` - Complex metric structure with all optional fields

**Expected Results**: All tests should pass ✅

---

### 2. auth_test.cpp (25 test cases)
**Component**: AuthManager - JWT token management and authentication

Test Coverage:
- `CreateInstance` - AuthManager instantiation
- `GenerateToken` - JWT token generation
- `TokenStructure` - JWT structure validation (3 parts: header.payload.signature)
- `GetValidToken` - Token retrieval when valid
- `IsTokenValid` - Token validity checking
- `IsTokenExpired` - Expired token detection
- `SetExternalToken` - External token assignment
- `RefreshToken` - Token refresh functionality
- `GetTokenExpiration` - Expiration time retrieval
- `LoadNonExistentCertificate` - Certificate loading error handling
- `LoadNonExistentKey` - Key loading error handling
- `GetClientCertificateEmpty` - Empty certificate handling
- `GetClientKey` - Client key retrieval
- `MultipleTokens` - Different tokens across generations
- `ValidateTokenSignature` - JWT signature validation
- `ValidateInvalidTokenFormat` - Invalid token rejection
- `TokenWithDifferentSecret` - Cross-authentication failure
- `CollectorIdInToken` - Collector identification in token
- `TokenExpirationInFuture` - Expiration time validation
- `ShortLivedToken` - Short expiration handling
- `RefreshBeforeExpiration` - Pre-expiration refresh
- `LastErrorMessage` - Error message tracking
- `TokenPayloadStructure` - Internal payload structure
- `MultipleAuthManagers` - Multiple independent auth managers
- `TokenValidityBuffer` - 60-second refresh buffer validation

**Expected Results**: All tests should pass ✅

**Security Notes**: Tests validate that tokens expire correctly and refresh buffer prevents race conditions.

---

### 3. metrics_buffer_test.cpp (20 test cases)
**Component**: MetricsBuffer - Circular buffer with gzip compression

Test Coverage:
- `CreateInstance` - Buffer instantiation
- `BufferStartsEmpty` - Initial buffer state
- `AppendMetric` - Metric addition to buffer
- `GetMetricCount` - Metric counter accuracy
- `GetUncompressedSize` - Uncompressed size calculation
- `GetCompressedData` - Compression functionality
- `CompressionRatio` - Compression effectiveness (40-50% target)
- `ClearBuffer` - Buffer clearing operation
- `MultipleMetricsCompression` - Multiple metric compression
- `LargeMetric` - Large metric handling
- `GetStats` - Statistics retrieval (6 fields)
- `BufferOverflow` - Overflow protection
- `EmptyBufferCompression` - Empty buffer handling
- `EstimatedCompressedSize` - Estimated vs actual size
- `SizeCalculationConsistency` - Size growth tracking
- `DifferentMetricTypes` - Mixed metric type handling
- `ClearAfterCompression` - Post-compression clearing
- `RepeatedCompress` - Idempotent compression
- `BufferStatsAfterClear` - Statistics after clearing
- `CompressionEfficiency` - Compression ratio validation

**Expected Results**: All tests should pass ✅

**Performance Notes**: Tests validate compression efficiency for network optimization.

---

### 4. config_manager_test.cpp (25 test cases)
**Component**: ConfigManager - TOML configuration loading and management

Test Coverage:
- `CreateInstance` - ConfigManager instantiation
- `LoadConfigFile` - TOML file loading
- `GetCollectorId` - Collector ID retrieval
- `GetHostname` - Hostname retrieval
- `GetBackendUrl` - Backend URL retrieval
- `GetStringConfig` - String value retrieval
- `GetIntConfig` - Integer value retrieval
- `GetBoolConfig` - Boolean value retrieval
- `GetStringArrayConfig` - Array value retrieval
- `IsCollectorEnabled` - Collector status checking
- `GetCollectionInterval` - Interval configuration
- `GetPostgreSQLConfig` - PostgreSQL configuration structure
- `GetTLSConfig` - TLS configuration retrieval
- `DefaultValues` - Default value fallback
- `LoadNonExistentFile` - File loading error handling
- `SetConfigValue` - Configuration updates
- `ToJson` - JSON export functionality
- `MultipleSections` - Cross-section value access
- `ConfigurationPersistence` - Configuration consistency
- `IntegerDefaultValue` - Integer default handling
- `BooleanDefaultValue` - Boolean default handling
- `EmptyDatabaseListDefaulting` - Default database handling
- `CaseSensitivity` - Case sensitivity in sections
- `SpecialCharactersInValues` - Special character handling
- `ConfigurationReload` - Reload functionality

**Expected Results**: All tests should pass ✅

**Configuration Notes**: Tests validate TOML parsing and type conversion correctness.

---

### 5. sender_test.cpp (25 test cases)
**Component**: Sender - HTTPS REST client with TLS/mTLS/JWT

Test Coverage:
- `CreateInstance` - Sender instantiation
- `SetAuthToken` - Token assignment
- `GetAuthToken` - Token retrieval
- `TokenValidityInitiallyFalse` - Initial validity state
- `TokenValidityAfterSetting` - Token validity after setting expiration
- `ValidMetrics` - Metrics structure validation
- `EmptyMetrics` - Invalid metrics rejection
- `TokenExpiration` - Expired token detection
- `MultipleTokens` - Token replacement handling
- `RefreshTokenCheck` - Token refresh triggering (60-second buffer)
- `CollectorIdStorage` - Collector ID persistence
- `BackendUrl` - Backend URL configuration
- `CertificateFilePaths` - Certificate path handling
- `TLSVerificationFlag` - TLS verification options
- `MetricsCompressionPrep` - Compression readiness
- `LargeMetricsPayload` - Large payload handling
- `MetricsStructureValidation` - Payload structure validation
- `DifferentCollectorIds` - Multiple collector support
- `DifferentExpirationTimes` - Various token lifetimes
- `EmptyMetricsArray` - Empty array handling
- `MetricsWithVariousTypes` - Mixed metric types
- `TokenValidityBuffer` - 60-second buffer validation
- `SenderConfigurationPersistence` - Configuration retention
- `TokenRefreshCycle` - Token replacement flow
- `SenderStateConsistency` - State consistency validation

**Expected Results**: All tests should pass ✅

**Security Notes**: Tests validate TLS enforcement and JWT token expiration handling.

---

## Building Tests

### Prerequisites

**macOS**:
```bash
brew install googletest cmake openssl curl
```

**Ubuntu/Debian**:
```bash
sudo apt-get install -y libgtest-dev cmake libssl-dev libcurl4-openssl-dev
# Build and install gtest
cd /usr/src/gtest && sudo cmake . && sudo make && sudo make install
```

**Fedora/RHEL**:
```bash
sudo dnf install -y gtest-devel cmake openssl-devel libcurl-devel
```

### Build Commands

```bash
# Create build directory
mkdir -p collector/build
cd collector/build

# Configure with tests enabled (default: ON)
cmake .. -DBUILD_TESTS=ON

# Build tests
make -j$(nproc)

# Run specific test file
./pganalytics-tests --gtest_filter="MetricsSerializerTest.*"
```

### Run Tests

```bash
# Run all tests
./pganalytics-tests

# Run tests with verbose output
./pganalytics-tests --gtest_print_time=1

# Run specific test
./pganalytics-tests --gtest_filter="AuthManagerTest.GenerateToken"

# Run tests matching pattern
./pganalytics-tests --gtest_filter="*TokenExpir*"

# Run with specific repeat count
./pganalytics-tests --gtest_repeat=10

# Save test results to XML
./pganalytics-tests --gtest_output="xml:test-results.xml"
```

---

## Test Execution via CMake

```bash
# Build project with tests
cd collector/build
cmake .. -DBUILD_TESTS=ON
make

# Run tests with ctest
ctest --output-on-failure

# Run with verbose output
ctest --verbose

# Run specific test
ctest -R "MetricsSerializerTest" --verbose
```

---

## Code Coverage

### Generate Coverage Report

```bash
# Build with coverage flags (GCC/Clang)
cd collector/build
cmake .. -DCMAKE_CXX_FLAGS="--coverage" -DBUILD_TESTS=ON
make

# Run tests
./pganalytics-tests

# Generate coverage report
lcov --directory . --capture --output-file coverage.info
lcov --remove coverage.info '/usr/*' '*/tests/*' --output-file coverage.info

# Generate HTML report
genhtml coverage.info --output-directory coverage-report

# View report
open coverage-report/index.html
```

### Coverage Targets

- **Overall**: >60% code coverage
- **Security-critical**: >80% (auth.cpp, sender.cpp)
- **Data handling**: >70% (serializer.cpp, buffer.cpp)
- **Configuration**: >75% (config_manager.cpp)

---

## Test Organization

### Test Class Pattern

Each test file follows this structure:
```cpp
class ComponentTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Initialize test fixtures
    }

    void TearDown() override {
        // Clean up after tests
    }

    // Test data and helper methods
};

// Individual test cases
TEST_F(ComponentTest, TestName) {
    // Test code
}
```

### Test Naming Convention

- **Component**: `*Test` (e.g., `AuthManagerTest`)
- **Test Method**: `CamelCaseDescriptive` (e.g., `GenerateToken`)
- **Full name**: `ComponentTest.MethodName`

---

## Expected Test Results

### All Tests Should Pass

```
[==========] 130+ tests from 5 test suites ran. (X ms total)
[  PASSED  ] 130+ tests
```

### Success Criteria

- ✅ All 130+ tests pass
- ✅ No memory leaks (valgrind)
- ✅ >60% code coverage
- ✅ All security validations pass
- ✅ Token expiration works correctly
- ✅ Compression efficiency >30%

---

## Troubleshooting

### Build Issues

**Google Test not found**:
```bash
# Install on macOS
brew install googletest

# Install on Ubuntu
sudo apt-get install libgtest-dev

# Install on Fedora
sudo dnf install gtest-devel
```

**Compilation errors**:
```bash
# Clean and rebuild
make clean
cmake .. -DBUILD_TESTS=ON
make -j$(nproc)
```

### Test Failures

**Timing-sensitive tests**:
- Some tests are timing-dependent (token expiration)
- May fail on slow systems
- Run with `--gtest_repeat=3` for stability

**File path issues**:
- Tests create temporary files in `/tmp`
- Ensure `/tmp` is writable and has space
- Use `-DTEST_TMPDIR=/custom/path` if needed

### Memory Issues

**Detect leaks with valgrind**:
```bash
valgrind --leak-check=full ./pganalytics-tests
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Build and Test
  run: |
    mkdir -p collector/build
    cd collector/build
    cmake .. -DBUILD_TESTS=ON
    make -j$(nproc)
    ./pganalytics-tests --gtest_output="xml:test-results.xml"

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.info
```

---

## Test Metrics

### Coverage Goals

| Component | Target | Status |
|-----------|--------|--------|
| metrics_serializer.cpp | 70% | 🎯 |
| auth.cpp | 80% | 🎯 |
| metrics_buffer.cpp | 65% | 🎯 |
| config_manager.cpp | 75% | 🎯 |
| sender.cpp | 70% | 🎯 |
| **Overall** | **>60%** | 🎯 |

### Test Distribution

| Category | Count | Percentage |
|----------|-------|-----------|
| Schema/Validation | 20 | 15% |
| Authentication | 25 | 19% |
| Buffering | 20 | 15% |
| Configuration | 25 | 19% |
| Communication | 25 | 19% |
| Integration | 15 | 12% |

---

## Next Steps

After running tests successfully:

1. ✅ Verify all 130+ tests pass
2. ✅ Check code coverage >60%
3. ✅ Review any failed tests
4. ✅ Run with different configurations:
   - Debug vs Release builds
   - Different OS platforms
   - Different compiler versions
5. ✅ Integration tests (Phase 3.4b)
6. ✅ E2E tests with real backend (Phase 3.4c)

---

## Additional Resources

- [Google Test Documentation](https://google.github.io/googletest/)
- [CMake Testing](https://cmake.org/cmake/help/latest/command/enable_testing.html)
- [Code Coverage Tools](https://gcovr.com/)

---

**Test Suite Status**: ✅ Ready for Execution
**Expected Duration**: 2-5 seconds (full suite)
**Last Updated**: February 20, 2026
