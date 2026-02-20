# Phase 3.4c - End-to-End Testing

This directory contains comprehensive end-to-end (E2E) tests for the pgAnalytics v3 system. These tests validate the complete metrics collection → transmission → storage → visualization pipeline with a real running backend.

## Overview

The E2E test suite verifies:
- **Collector Registration**: Collector can register with backend and obtain credentials
- **Metrics Ingestion**: Metrics flow from collector to PostgreSQL/TimescaleDB
- **Configuration Management**: Collector can pull and apply configuration from backend
- **Dashboard Visibility**: Metrics are visible in Grafana dashboards
- **Performance**: System performs within acceptable latency/throughput targets
- **Failure Recovery**: System gracefully handles and recovers from failures

## Directory Structure

```
collector/tests/e2e/
├── docker-compose.e2e.yml        # Docker Compose environment
├── init-schema.sql               # PostgreSQL schema initialization
├── init-timescale.sql            # TimescaleDB hypertable setup
├── collector-config.toml         # Collector configuration for E2E tests
├── fixtures.h                    # Test data and fixtures
│
├── e2e_harness.h/cpp            # Docker Compose lifecycle management
├── http_client.h/cpp            # HTTPS client for backend API calls
├── database_helper.h/cpp        # Database query utilities
├── grafana_helper.h/cpp         # Grafana API utilities
│
├── 1_collector_registration_test.cpp    # Registration tests (10 tests)
├── 2_metrics_ingestion_test.cpp         # Metrics flow tests (12 tests)
├── 3_configuration_test.cpp             # Configuration tests (8 tests)
├── 4_dashboard_visibility_test.cpp      # Dashboard tests (6 tests)
├── 5_performance_test.cpp               # Performance tests (5 tests)
├── 6_failure_recovery_test.cpp          # Failure recovery tests (8 tests)
│
└── README.md                     # This file
```

## Prerequisites

- **Docker & Docker Compose**: v20.10+ and v2.0+
- **PostgreSQL/psql**: For manual database inspection (optional)
- **curl**: For testing HTTP endpoints (optional)
- **C++ Compiler**: GCC 9+ or Clang 10+ (for building tests)
- **CMake**: 3.25+ (for building tests)

## Setup

### 1. Generate TLS Certificates

TLS certificates are required for HTTPS communication between collector and backend.

```bash
# Generate self-signed certificates for E2E testing
cd collector/tests/e2e
../../scripts/generate_e2e_certs.sh

# Verify certificates were created
ls -la ../../tls/
```

### 2. Build E2E Tests

```bash
cd collector/build
cmake .. -DBUILD_E2E_TESTS=ON
make -j4 e2e_tests
```

### 3. Start E2E Environment

```bash
cd collector/tests/e2e
docker-compose -f docker-compose.e2e.yml up -d

# Wait for services to become healthy
docker-compose -f docker-compose.e2e.yml ps

# Check logs if needed
docker-compose -f docker-compose.e2e.yml logs -f backend
```

## Running E2E Tests

### Run All E2E Tests

```bash
./tests/e2e/pganalytics-e2e-tests
```

### Run Specific Test Suite

```bash
# Collector registration tests
./tests/e2e/pganalytics-e2e-tests --gtest_filter="E2ERegistration*"

# Metrics ingestion tests
./tests/e2e/pganalytics-e2e-tests --gtest_filter="E2EMetrics*"

# Configuration tests
./tests/e2e/pganalytics-e2e-tests --gtest_filter="E2EConfig*"
```

### Run with Verbose Output

```bash
./tests/e2e/pganalytics-e2e-tests --gtest_filter="*" -v
```

## Test Categories

### 1. Collector Registration (10 tests)
Tests the collector registration flow with the backend API.

**Key Tests**:
- `RegisterNewCollector` - Submit registration request
- `RegistrationValidation` - Validate JWT token structure
- `CertificatePersistence` - Certificate saved to filesystem
- `TokenExpiration` - Token has correct expiration
- `MultipleRegistrations` - Different collectors can register independently

**Expected Results**: All collectors successfully register and receive valid credentials.

### 2. Metrics Ingestion (12 tests)
Tests metrics flowing from collector to TimescaleDB storage.

**Key Tests**:
- `SendMetricsSuccess` - Metrics pushed and acknowledged (HTTP 200/201)
- `MetricsStored` - Metrics appear in TimescaleDB within expected time
- `MetricsSchema` - Columns match expected schema
- `PayloadCompression` - Metrics compressed with gzip
- `ConcurrentPushes` - Multiple metrics pushes work correctly

**Expected Results**: Metrics successfully transmitted and stored.

### 3. Configuration Management (8 tests)
Tests configuration pull and hot-reload capabilities.

**Key Tests**:
- `ConfigPullOnStartup` - Collector pulls config on startup
- `ConfigValidation` - Invalid configs rejected gracefully
- `HotReload` - Config changes propagated without restart
- `CollectionIntervals` - Intervals respected from config

**Expected Results**: Configuration system fully functional.

### 4. Dashboard Visibility (6 tests)
Tests that metrics are visible in Grafana dashboards.

**Key Tests**:
- `GrafanaDatasource` - PostgreSQL datasource working
- `DashboardLoads` - Pre-built dashboards load
- `MetricsVisible` - Dashboard panels show metrics
- `AlertsConfigured` - Alert rules exist and are active

**Expected Results**: Metrics visible in Grafana UI.

### 5. Performance (5 tests)
Tests performance within acceptable targets.

**Key Tests**:
- `MetricCollectionLatency` - Collection <500ms
- `MetricsTransmissionLatency` - Transmission <1 second
- `ThroughputSustained` - 100+ metrics/sec sustainable
- `MemoryStability` - No memory leaks over time

**Expected Results**: All performance targets met.

### 6. Failure Recovery (8 tests)
Tests handling of failures and proper recovery.

**Key Tests**:
- `BackendUnavailable` - Retries when backend down
- `NetworkPartition` - Metrics buffered during outage
- `TokenExpiration` - Auto-refresh of expired tokens
- `DatabaseDown` - Graceful handling of DB unavailability

**Expected Results**: System recovers cleanly from failures.

## Manual Testing

### Check Backend Health

```bash
# Backend API health check
curl -k https://localhost:8080/api/v1/health

# Expected response: {"status": "healthy"}
```

### Check Database

```bash
# Connect to PostgreSQL
PGPASSWORD=pganalytics psql -h localhost -U postgres -d pganalytics

# Check collectors table
SELECT * FROM pganalytics.collector_registry;

# Connect to TimescaleDB (port 5433)
PGPASSWORD=pganalytics psql -h localhost -p 5433 -U postgres -d metrics

# Check metrics
SELECT COUNT(*) FROM metrics_pg_stats;
```

### Check Grafana

Access Grafana dashboard: http://localhost:3000
- Username: admin
- Password: admin

### Check Logs

```bash
# Backend logs
docker-compose -f docker-compose.e2e.yml logs backend

# Collector logs
docker-compose -f docker-compose.e2e.yml logs collector

# View in real-time
docker-compose -f docker-compose.e2e.yml logs -f
```

## Cleanup

### Stop E2E Environment

```bash
docker-compose -f docker-compose.e2e.yml down
```

### Remove All Volumes (WARNING: Deletes Data)

```bash
docker-compose -f docker-compose.e2e.yml down -v
```

### Clean Build Artifacts

```bash
cd collector/build
make clean
```

## Troubleshooting

### Services Fail to Start

Check service health and logs:
```bash
docker-compose -f docker-compose.e2e.yml ps
docker-compose -f docker-compose.e2e.yml logs [service-name]
```

### Connection Refused

Backend may not be ready. Wait 30 seconds and retry:
```bash
# Wait for backend to be healthy
docker-compose -f docker-compose.e2e.yml exec backend \
  curl -k https://localhost:8080/api/v1/health
```

### Database Connection Issues

Verify database is accessible:
```bash
PGPASSWORD=pganalytics psql -h localhost -U postgres -c "SELECT version();"
```

### Tests Timeout

Increase timeout values in test code or docker-compose if services are slow.

### Port Conflicts

If ports are in use, modify docker-compose.e2e.yml:
```yaml
services:
  backend:
    ports:
      - "8081:8080"  # Use 8081 instead of 8080
```

Then update E2E harness to use new port.

## Test Results

After running tests, check results:

```bash
# View test summary
./tests/e2e/pganalytics-e2e-tests --gtest_filter="*" 2>&1 | tail -50

# Export results to file
./tests/e2e/pganalytics-e2e-tests > e2e_results.txt 2>&1
```

## Performance Baseline

Expected metrics from a healthy system:

| Metric | Target | Tolerance |
|--------|--------|-----------|
| Collection latency | <500ms | ±100ms |
| Transmission latency | <1000ms | ±200ms |
| Database insert latency | <100ms | ±50ms |
| Sustained throughput | 100+ metrics/sec | ≥80 metrics/sec |
| Memory growth | <5MB/hour | Monitor for trends |

## CI/CD Integration

E2E tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run E2E Tests
  run: |
    cd collector/tests/e2e
    docker-compose -f docker-compose.e2e.yml up -d
    sleep 30  # Wait for services
    ../../../build/tests/e2e/pganalytics-e2e-tests
    docker-compose -f docker-compose.e2e.yml down
```

## Development & Debugging

### Add New E2E Test

1. Create test file (e.g., `7_new_feature_test.cpp`)
2. Include test harness header
3. Use E2ETestHarness for lifecycle management
4. Use E2EHttpClient for API calls
5. Use E2EDatabaseHelper for verification

Example:
```cpp
#include <gtest/gtest.h>
#include "e2e_harness.h"
#include "http_client.h"
#include "database_helper.h"
#include "fixtures.h"

class E2ENewFeatureTest : public ::testing::Test {
protected:
    static E2ETestHarness harness;

    static void SetUpTestSuite() {
        harness.startStack(60);
        harness.resetData();
    }

    static void TearDownTestSuite() {
        harness.stopStack();
    }
};

E2ETestHarness E2ENewFeatureTest::harness;

TEST_F(E2ENewFeatureTest, MyNewTest) {
    // Test implementation
}
```

### Enable Verbose Logging

In test code:
```cpp
E2EHttpClient http_client(harness.getBackendUrl());
http_client.setVerbose(true);
http_client.setLogFile("e2e_http.log");
```

### Database Debugging

```cpp
E2EDatabaseHelper db(
    harness.getDatabaseUrl(),
    harness.getTimescaleUrl()
);

// Execute custom queries
std::string result = db.executeQuery(
    "SELECT * FROM metrics_pg_stats LIMIT 5;",
    true
);
```

## References

- **Phase 3.4b Plan**: See `PHASE_3_4B_MILESTONE_2_PLAN.md` for mock server design
- **Backend API**: See `backend/README.md` for API documentation
- **Collector Design**: See `collector/README.md` for collector architecture
- **Docker Compose**: See main `docker-compose.yml` for production setup

## Contributing

When modifying E2E tests:

1. Follow existing test patterns (AAA: Arrange-Act-Assert)
2. Use test fixtures for consistent data
3. Add descriptive test names
4. Include comments for complex logic
5. Verify all tests pass before submitting
6. Update this README if adding new test categories

## Status

- **Current Phase**: Phase 3.4c - End-to-End Testing
- **Test Count**: 49 E2E tests planned
- **Status**: Infrastructure implementation in progress
- **Target**: All tests passing with >90% success rate

---

**Last Updated**: February 19, 2026
**Maintained By**: pgAnalytics v3 Development Team

