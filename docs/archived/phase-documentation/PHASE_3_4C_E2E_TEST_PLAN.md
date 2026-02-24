# Phase 3.4c - End-to-End Testing Plan

**Status**: Planning Phase
**Date**: February 19, 2026
**Duration**: 2-3 weeks
**Objective**: Validate full metrics collection → transmission → storage → visualization pipeline with real backend

---

## Overview

Phase 3.4c focuses on end-to-end (E2E) testing with the actual pgAnalytics v3 backend running in docker-compose. Unlike Phase 3.4b (mock backend tests), E2E tests will:

1. **Start the full stack** (PostgreSQL, TimescaleDB, Backend API, Grafana, Collector)
2. **Exercise real HTTP communication** (TLS 1.3 + mTLS + JWT with actual certificates)
3. **Verify metrics storage** in TimescaleDB
4. **Validate dashboard visibility** in Grafana
5. **Test configuration reload** from backend API
6. **Measure performance** under realistic load

---

## Part 1: Infrastructure Setup

### 1.1 Docker Compose E2E Environment

**File**: `docker-compose.e2e.yml` (new)
**Purpose**: Complete stack for E2E testing with additional services

**Services**:
- ✅ PostgreSQL 16 (metadata storage)
- ✅ TimescaleDB 16 (metrics storage)
- ✅ Backend API (Go, port 8080)
- ✅ Collector (C/C++, containerized)
- ✅ Grafana (port 3000)
- ✅ Redis (caching, port 6379)
- 🆕 Init Service (database setup)
- 🆕 Test Runner Container (execute E2E tests)

**Key Differences from docker-compose.yml**:
```yaml
services:
  # Same as main, but with:
  backend:
    # Add test mode environment variables
    TESTING_MODE: "true"
    MOCK_EXTERNAL_API: "false"

  collector:
    # For E2E: Use test configuration
    environment:
      CONFIG_MODE: "e2e_test"
      COLLECTOR_ID: "e2e_col_001"

  test_runner:
    # New service: Runs E2E tests against stack
    image: e2e-tests:latest
    depends_on:
      backend:
        condition: service_healthy
      collector:
        condition: service_running
    volumes:
      - ./e2e/tests:/tests
      - ./e2e/results:/results
```

### 1.2 Database Initialization

**File**: `backend/migrations/e2e_init.sql` (new)
**Purpose**: Initialize schema for E2E testing

**Schema Setup**:
```sql
-- Create pganalytics schema
CREATE SCHEMA IF NOT EXISTS pganalytics;

-- Create collector_registry table
CREATE TABLE IF NOT EXISTS pganalytics.collector_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    hostname VARCHAR(255),
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'active'
);

-- Create api_tokens table
CREATE TABLE IF NOT EXISTS pganalytics.api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collector_id VARCHAR(255) REFERENCES pganalytics.collector_registry(collector_id),
    token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create TimescaleDB hypertables
CREATE TABLE IF NOT EXISTS pganalytics.metrics_pg_stats (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    collector_id VARCHAR(255),
    database VARCHAR(255),
    schema VARCHAR(255),
    table_name VARCHAR(255),
    rows BIGINT,
    size_bytes BIGINT
) WITH (timescaledb.compress_orderby = 'time DESC');

SELECT create_hypertable('pganalytics.metrics_pg_stats', 'time', if_not_exists => TRUE);
SELECT add_compress_policy('pganalytics.metrics_pg_stats', INTERVAL '7 days', if_not_exists => TRUE);

-- Similar for other metric types:
-- metrics_pg_log, metrics_sysstat, metrics_disk_usage
```

### 1.3 TLS Certificate Generation

**File**: `scripts/generate_e2e_certs.sh` (new)
**Purpose**: Generate self-signed certificates for E2E testing

**Generates**:
```bash
# CA Certificate
ca.crt, ca.key

# Server Certificate (Backend API)
server.crt, server.key

# Client Certificate (Collector)
client.crt, client.key

# Store in: tls/ directory (mounted in docker-compose)
```

**Certificate Details**:
- **Validity**: 365 days
- **Key Size**: 2048 bits (RSA)
- **Algorithms**: TLS 1.3 compatible
- **SAN**: localhost, 127.0.0.1, backend (DNS name in compose)

---

## Part 2: E2E Test Suite

### 2.1 Test Structure

**Location**: `collector/tests/e2e/`
**Framework**: Google Test + Docker Compose integration
**Language**: C++ with shell scripts

```
collector/tests/e2e/
├── CMakeLists.txt              # Build configuration
├── docker-compose.yml           # Spin up services
├── fixtures.h                   # E2E test data
├── helpers.cpp/h               # Utility functions
│
├── 1_collector_registration_test.cpp    (10 tests)
├── 2_metrics_ingestion_test.cpp         (12 tests)
├── 3_configuration_test.cpp             (8 tests)
├── 4_dashboard_visibility_test.cpp      (6 tests)
├── 5_performance_test.cpp               (5 tests)
├── 6_failure_recovery_test.cpp          (8 tests)
│
└── run_e2e_tests.sh             # Master test runner script
```

### 2.2 Test Categories & Scope

#### Category 1: Collector Registration (10 tests)

**File**: `1_collector_registration_test.cpp`
**Purpose**: Verify collector can register with backend and obtain credentials

**Tests**:
1. `RegisterNewCollector` - Submit registration, receive JWT token + cert
2. `RegistrationValidation` - Validate token claims (exp, collector_id, iss)
3. `CertificatePersistence` - Certificate saved to filesystem
4. `TokenExpiration` - Token has correct expiration time (15 min default)
5. `MultipleRegistrations` - Different collectors can register independently
6. `RegistrationFailure` - Handle invalid input gracefully
7. `DuplicateRegistration` - Same collector ID fails/overwrites
8. `CertificateFormat` - Certificate is valid X.509 format
9. `PrivateKeyProtection` - Private key file has restricted permissions (600)
10. `RegistrationAudit` - Registration logged with timestamp and details

**Key Validations**:
- Backend accepts registration request
- JWT token structure: `header.payload.signature`
- Certificate chain validation
- Database records created in `pganalytics.collector_registry`

#### Category 2: Metrics Ingestion (12 tests)

**File**: `2_metrics_ingestion_test.cpp`
**Purpose**: Verify metrics flow from collector to TimescaleDB

**Tests**:
1. `SendMetricsSuccess` - POST /api/v1/metrics/push returns 200
2. `MetricsStored` - Metrics appear in TimescaleDB hypertables
3. `MetricsSchema` - Columns match expected schema
4. `TimestampAccuracy` - Collector timestamp preserved
5. `MetricTypes` - All 4 metric types (pg_stats, pg_log, sysstat, disk) stored
6. `PayloadCompression` - Metrics compressed with gzip (validate Content-Encoding)
7. `MetricsCount` - Correct number of metrics rows inserted
8. `DataIntegrity` - No missing or corrupted data
9. `ConcurrentPushes` - Multiple metrics pushes don't interfere
10. `LargePayload` - Handle 10MB+ compressed payload
11. `PartialFailure` - Some metrics stored even if some fail
12. `MetricsQuery` - Can query stored metrics via backend API

**Key Validations**:
- HTTP 200/201 response codes
- Database inserts verified with SQL queries
- Row counts match expected
- Timestamps in ISO8601 format
- gzip compression effective (>50% reduction)

#### Category 3: Configuration Management (8 tests)

**File**: `3_configuration_test.cpp`
**Purpose**: Verify config pull, parsing, and hot-reload

**Tests**:
1. `ConfigPullOnStartup` - Collector pulls config from backend on startup
2. `ConfigValidation` - Invalid config rejected with clear error
3. `ConfigApplication` - Collector applies config to metric collectors
4. `HotReload` - Config change on backend picked up by collector (5 min interval)
5. `ConfigVersionTracking` - Backend tracks config version
6. `CollectionIntervals` - Collector respects configured intervals (60s push)
7. `EnabledMetrics` - Only enabled metric types collected
8. `ConfigPersistence` - Local config cache survives restart

**Key Validations**:
- GET /api/v1/config/{collector_id} returns valid config
- Configuration parsed correctly from TOML
- Metric collection intervals respected
- Config versions tracked
- Hot-reload without losing metrics

#### Category 4: Dashboard Visibility (6 tests)

**File**: `4_dashboard_visibility_test.cpp`
**Purpose**: Verify metrics visible in Grafana dashboards

**Tests**:
1. `GrafanaDatasource` - PostgreSQL datasource configured and working
2. `DashboardLoads` - Pre-built dashboards load without errors
3. `MetricsVisible` - Dashboard panels show collected metrics
4. `TimeRangeQuery` - Time range selector works correctly
5. `AlertsConfigured` - Alert rules defined for metrics
6. `AlertTriggered` - Alert fires when threshold exceeded

**Key Validations**:
- Grafana API health check
- Datasources connected to PostgreSQL/TimescaleDB
- Dashboard JSON loaded from provisioning
- Metrics queries return data
- Time-series visualizations render

#### Category 5: Performance Testing (5 tests)

**File**: `5_performance_test.cpp`
**Purpose**: Measure performance under realistic load

**Tests**:
1. `MetricCollectionLatency` - Collection takes <500ms for 1000 metrics
2. `MetricsTransmissionLatency` - Transmission to backend <1 second
3. `DatabaseInsertLatency` - TimescaleDB inserts <100ms
4. `ThroughputSustained` - Maintain 100+ metrics/sec under continuous load
5. `MemoryStability` - Memory usage stable after 1000+ cycles (no leaks)

**Key Validations**:
- Latency measurements with timers
- Memory profiling before/after
- CPU usage tracking
- Network bandwidth measurement
- No memory leaks over extended run

#### Category 6: Failure Recovery (8 tests)

**File**: `6_failure_recovery_test.cpp`
**Purpose**: Verify resilience to failures and proper recovery

**Tests**:
1. `BackendUnavailable` - Collector retries when backend is down
2. `NetworkPartition` - Metrics buffered during network outage
3. `NetworkRecovery` - Buffered metrics sent after network restores
4. `TokenExpiration` - Collector refreshes expired token automatically
5. `AuthenticationFailure` - 401 error handled with retry
6. `CertificateFailure` - Invalid cert detected and logged
7. `DatabaseDown` - Graceful handling if TimescaleDB unavailable
8. `PartialDataRecovery` - Incomplete metrics recovered on retry

**Key Validations**:
- Retry logic with exponential backoff
- Buffering mechanism functional
- Token refresh triggered by 401
- Error logging with context
- No data loss during recovery

---

## Part 3: Test Execution Framework

### 3.1 Test Harness (`e2e_harness.cpp/h`)

**Purpose**: Manage docker-compose lifecycle and test execution

**Responsibilities**:
```cpp
class E2ETestHarness {
public:
    // Lifecycle management
    void startStack();           // docker-compose up -d
    void stopStack();            // docker-compose down
    void resetData();            // Clear databases

    // Health checks
    bool isBackendReady();       // Poll /health endpoint
    bool isCollectorRunning();   // Check process status
    bool isDatabaseReady();      // Test connections

    // Helper functions
    std::string getBackendUrl();        // https://backend:8080
    std::string getMetricsQuery();      // Query TimescaleDB
    json queryGrafana(endpoint);        // Query Grafana API

    // Wait utilities
    void waitForCondition(condition, timeout);
    void waitForMetrics(count, timeout);
};
```

**Example Usage**:
```cpp
class E2ETest : public ::testing::Test {
protected:
    static E2ETestHarness harness;

    static void SetUpTestSuite() {
        harness.startStack();
        harness.waitForBackendReady(60000);  // 60 second timeout
    }

    static void TearDownTestSuite() {
        harness.stopStack();
    }

    void SetUp() override {
        harness.resetData();
    }
};

TEST_F(E2ETest, MetricsFlow) {
    // Test logic
}
```

### 3.2 HTTP Client for E2E

**Purpose**: Make HTTPS requests to real backend with actual TLS

**Library**: libcurl (same as collector)
**Features**:
- TLS 1.3 enforcement
- mTLS certificate validation
- JWT token handling
- Response parsing

```cpp
class E2EHttpClient {
public:
    E2EHttpClient(const std::string& backend_url,
                  const std::string& cert_file,
                  const std::string& key_file);

    // Make authenticated requests
    json postMetrics(const json& payload);
    json getConfig(const std::string& collector_id);
    json registerCollector(const json& request);

    // Query helpers
    std::string getLastResponseStatus();
    std::string getLastResponseBody();
};
```

### 3.3 Database Query Helpers

**Purpose**: Verify metrics stored in TimescaleDB

```cpp
class E2EDatabaseHelper {
public:
    E2EDatabaseHelper(const std::string& connection_string);

    // Query metrics
    int getMetricsCount(const std::string& table);
    json getLatestMetrics(const std::string& table, int limit);

    // Verify storage
    bool metricsExist(const std::string& collector_id);
    int countMetricsByType(const std::string& type);

    // Clear data
    void clearAllMetrics();
    void clearCollectorMetrics(const std::string& collector_id);
};
```

### 3.4 Grafana API Helper

**Purpose**: Query Grafana to verify dashboards and alerts

```cpp
class E2EGrafanaHelper {
public:
    E2EGrafanaHelper(const std::string& grafana_url);

    // Datasource operations
    bool isDatasourceHealthy(const std::string& name);

    // Dashboard operations
    json getDashboard(const std::string& uid);
    json getPanelData(const std::string& dashboard_uid, int panel_id);

    // Alert operations
    std::vector<json> getAlerts();
    bool isAlertFiring(const std::string& alert_name);
};
```

---

## Part 4: Implementation Phases

### Phase 1: Infrastructure & Harness (Week 1)
- [x] docker-compose.e2e.yml setup
- [x] Certificate generation script
- [x] Database initialization
- [ ] E2E test harness implementation
- [ ] HTTP client wrapper
- [ ] Database helper class

### Phase 2: Core E2E Tests (Week 2)
- [ ] Collector registration tests (10 tests)
- [ ] Metrics ingestion tests (12 tests)
- [ ] Configuration tests (8 tests)
- [ ] Dashboard visibility tests (6 tests)

### Phase 3: Performance & Recovery (Week 3)
- [ ] Performance tests (5 tests)
- [ ] Failure recovery tests (8 tests)
- [ ] Stress testing with load
- [ ] Documentation

### Phase 4: Documentation & Validation
- [ ] E2E test README
- [ ] Performance baseline documentation
- [ ] Test results reporting
- [ ] Production readiness checklist

---

## Part 5: Critical Files to Create

**New Files** (~2500 lines):
1. `collector/tests/e2e/docker-compose.e2e.yml` (100 lines)
2. `collector/tests/e2e/CMakeLists.txt` (50 lines)
3. `collector/tests/e2e/fixtures.h` (150 lines)
4. `collector/tests/e2e/e2e_harness.cpp/h` (300 lines)
5. `collector/tests/e2e/http_client.cpp/h` (250 lines)
6. `collector/tests/e2e/database_helper.cpp/h` (200 lines)
7. `collector/tests/e2e/grafana_helper.cpp/h` (150 lines)
8. `collector/tests/e2e/1_collector_registration_test.cpp` (250 lines)
9. `collector/tests/e2e/2_metrics_ingestion_test.cpp` (300 lines)
10. `collector/tests/e2e/3_configuration_test.cpp` (200 lines)
11. `collector/tests/e2e/4_dashboard_visibility_test.cpp` (150 lines)
12. `collector/tests/e2e/5_performance_test.cpp` (200 lines)
13. `collector/tests/e2e/6_failure_recovery_test.cpp` (250 lines)
14. `scripts/generate_e2e_certs.sh` (100 lines)
15. `backend/migrations/e2e_init.sql` (150 lines)

**Modified Files**:
1. `collector/tests/CMakeLists.txt` - Add E2E test build configuration
2. `docker-compose.yml` - Add test mode configuration (optional)

**Documentation** (500 lines):
1. `collector/tests/e2e/README.md` - E2E test guide
2. `PHASE_3_4C_E2E_TEST_RESULTS.md` - Test results and metrics

---

## Part 6: Success Criteria

### Build & Compilation
- ✅ E2E tests compile without errors
- ✅ docker-compose environment starts successfully
- ✅ All services reach healthy status
- ✅ Certificates generated correctly

### Test Execution
- ✅ All 49 E2E tests execute (no crashes)
- ✅ >90% tests passing (allow for environmental issues)
- ✅ Clear error messages for failures
- ✅ Results logged to file

### Functional Coverage
- ✅ Collector registration flow complete
- ✅ Metrics transmitted to backend
- ✅ Metrics stored in TimescaleDB
- ✅ Metrics visible in Grafana
- ✅ Configuration reload works
- ✅ Error recovery verified

### Performance Metrics
- ✅ Metric collection: <500ms
- ✅ Transmission: <1 second
- ✅ Database insert: <100ms
- ✅ Sustained throughput: 100+ metrics/sec
- ✅ Memory stable (no growth)

### Documentation
- ✅ E2E test README
- ✅ Performance baseline established
- ✅ Failure modes documented
- ✅ Troubleshooting guide

---

## Part 7: Test Execution Commands

### Start E2E Environment
```bash
cd collector/tests/e2e
docker-compose -f docker-compose.e2e.yml up -d
docker-compose -f docker-compose.e2e.yml ps
```

### Check Service Health
```bash
# Backend API
curl -k https://localhost:8080/api/v1/health

# Grafana
curl http://localhost:3000/api/health

# Databases
psql -h localhost -U postgres -d pganalytics -c "SELECT version();"
```

### Run E2E Tests
```bash
# Build E2E tests
cmake .. -DBUILD_E2E_TESTS=ON
make -j4 e2e_tests

# Run all E2E tests
./tests/e2e/pganalytics-e2e-tests

# Run specific test suite
./tests/e2e/pganalytics-e2e-tests --gtest_filter="E2ERegistration*"

# Run with verbose output
./tests/e2e/pganalytics-e2e-tests --gtest_filter="*" -v
```

### Stop E2E Environment
```bash
docker-compose -f docker-compose.e2e.yml down
docker volume prune  # Optional: clean up volumes
```

---

## Part 8: Potential Challenges & Mitigations

| Challenge | Impact | Mitigation |
|---|---|---|
| Docker not available | Cannot run tests | Provide alternative: local backend startup script |
| Network latency | Tests flaky | Increase timeouts, add retry logic |
| Port conflicts | Services fail to start | Use random port allocation |
| Certificate issues | TLS failures | Pre-generate certs, validate before test |
| Database corruption | Tests fail | Reset schema between test suites |
| Memory exhaustion | OOM kills | Profile memory usage, set limits |
| Grafana initialization | Dashboard not ready | Add health check with longer timeout |

---

## Part 9: Integration with CI/CD

**GitHub Actions Workflow** (new):
```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:latest
    steps:
      - uses: actions/checkout@v3
      - name: Start E2E Stack
        run: docker-compose -f collector/tests/e2e/docker-compose.e2e.yml up -d
      - name: Wait for Services
        run: sleep 30  # Or implement health check
      - name: Build E2E Tests
        run: cmake .. -DBUILD_E2E_TESTS=ON && make e2e_tests
      - name: Run E2E Tests
        run: ./tests/e2e/pganalytics-e2e-tests
      - name: Upload Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: e2e-results
          path: ./e2e-results/
      - name: Cleanup
        if: always()
        run: docker-compose -f collector/tests/e2e/docker-compose.e2e.yml down
```

---

## Part 10: Success Measurement

### Test Results Report
- Total tests: 49
- Passing: >90% (target)
- Failures analyzed and documented
- Performance metrics captured

### Performance Baseline
- Collection latency: ~XXX ms (measured)
- Transmission latency: ~XXX ms (measured)
- Database latency: ~XXX ms (measured)
- Memory usage: ~XXX MB (measured)

### Coverage Validation
- All major flows tested
- Error paths exercised
- Recovery scenarios verified
- Performance validated

---

## Next Steps

**Immediate** (Phase 3.4c.1):
1. Implement E2E harness
2. Create HTTP client wrapper
3. Create database helper
4. Set up docker-compose.e2e.yml

**Short Term** (Phase 3.4c.2):
1. Implement registration tests
2. Implement metrics ingestion tests
3. Implement configuration tests

**Medium Term** (Phase 3.4c.3):
1. Implement dashboard visibility tests
2. Implement performance tests
3. Implement recovery tests

**Final** (Phase 3.4c.4):
1. Document results
2. Establish performance baselines
3. Production readiness validation

---

**Plan Created**: February 19, 2026
**Target Duration**: 2-3 weeks
**Status**: Ready for implementation
**Next Action**: Implement E2E harness infrastructure
