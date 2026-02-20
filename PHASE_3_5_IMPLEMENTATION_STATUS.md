# Phase 3.5: C/C++ Collector Modernization - Implementation Status

**Date**: February 19, 2026
**Status**: IN PROGRESS - Starting implementation
**Branch**: `feature/phase3-collector-modernization`

---

## Current Situation

### What's Completed
- ✅ **Phase 3.4**: Complete Testing Suite (272 tests, 91.6% passing)
- ✅ **Project Structure**: All directories and CMakeLists.txt configured
- ✅ **Header Files**: All interfaces defined
- ✅ **Unit Tests**: 20+ test cases, 100% passing (metrics_serializer, auth, buffer, config)
- ✅ **Config System**: TOML parsing with hot-reload support
- ✅ **Metrics Buffer**: Circular buffer with gzip compression implemented
- ✅ **Metrics Serializer**: JSON schema validation for Phase 2 backend
- ✅ **Authentication**: JWT token generation, validation, mTLS cert loading
- ✅ **Sender**: HTTP REST client with libcurl and TLS 1.3 setup

### What Needs Implementation

#### Priority 1: Complete Plugin Implementations
1. **postgres_plugin.cpp** (STUB - needs libpq integration)
   - [ ] Implement PostgreSQL connections with libpq
   - [ ] Query table statistics, index statistics, database stats
   - [ ] Handle multiple databases per config
   - [ ] Parse results into JSON matching backend schema
   - [ ] Reuse v2 SQL logic from `coletor/src/collectors.cpp`
   - **Target**: Full functional implementation with database access

2. **sysstat_plugin.cpp** (PARTIAL - needs /proc parsing)
   - [ ] Implement /proc/stat, /proc/meminfo, /proc/diskstats parsing
   - [ ] Calculate CPU%, memory usage, iops, throughput
   - [ ] Generate JSON output matching backend schema
   - **Target**: Full functional /proc parsing

3. **log_plugin.cpp** (STUB - needs log parsing)
   - [ ] Read PostgreSQL log files with state tracking
   - [ ] Parse log lines into structured format
   - [ ] Filter by log level (LOG, WARNING, ERROR, FATAL)
   - [ ] Generate JSON output matching backend schema
   - **Target**: Functional log file parsing

4. **disk_usage_plugin.cpp** (STUB - needs df command)
   - [ ] Execute `df` command and parse output
   - [ ] Calculate usage percentages
   - [ ] Generate JSON output matching backend schema
   - **Target**: Functional disk usage collection

#### Priority 2: Complete Main Loop Integration
1. **main.cpp improvements**
   - [ ] Implement config pull from backend (every 5 minutes)
   - [ ] Implement SIGHUP hot-reload handler
   - [ ] Add structured JSON logging with spdlog
   - [ ] Implement performance metrics collection
   - [ ] Add graceful shutdown with timeout

#### Priority 3: Complete Sender Implementation
1. **sender.cpp improvements**
   - [ ] Test HTTP status code handling (200/201 success, 401 token refresh)
   - [ ] Implement exponential backoff retry (max 3 retries)
   - [ ] Verify TLS 1.3 enforcement (no TLS 1.2 fallback)
   - [ ] Verify mTLS certificate validation
   - [ ] Add connection pooling / keep-alive

#### Priority 4: Complete Collector Registrations
1. **collector.cpp**
   - [ ] Fix compiler warnings about unused fields
   - [ ] Implement plugin registration system
   - [ ] Implement plugin interface inheritance

#### Priority 5: Comprehensive Testing & Validation
1. **Unit Tests**
   - [ ] Add postgres_plugin tests (mock PostgreSQL)
   - [ ] Add sysstat_plugin tests
   - [ ] Add log_plugin tests
   - [ ] Add disk_usage_plugin tests
   - **Target**: 60+ total unit tests

2. **Integration Tests**
   - [ ] Test mock backend communication
   - [ ] Test metrics format validation
   - [ ] Test error handling and retries
   - **Target**: 20+ integration scenarios

3. **E2E Tests**
   - [ ] Verify with real backend (docker-compose)
   - [ ] Test full registration flow
   - [ ] Test metrics push cycles
   - [ ] Test config pull and reload
   - **Target**: All E2E tests passing with Docker

#### Priority 6: Documentation & Finalization
1. **Documentation**
   - [ ] Create collector/README.md with build/install/config instructions
   - [ ] Create COLLECTOR-ARCHITECTURE.md
   - [ ] Create COLLECTOR-MIGRATION.md (v2 → v3 mapping)
   - [ ] Create SECURITY-COLLECTOR.md (TLS, mTLS, JWT)

2. **Validation**
   - [ ] Verify TLS 1.3 enforcement
   - [ ] Verify mTLS validation
   - [ ] Verify JWT in all API calls
   - [ ] Verify gzip compression >40%
   - [ ] Verify no hardcoded credentials
   - [ ] Verify config hot-reload
   - [ ] Performance: <100ms collect, <50ms serialize, <500ms push
   - [ ] Memory stability after 1000+ cycles

3. **Finalization**
   - [ ] Create PR with comprehensive description
   - [ ] Request code review
   - [ ] Merge to main branch

---

## Implementation Order

### Session 1 (Current): Foundation & Plugins
1. Complete sysstat_plugin.cpp with /proc parsing
2. Complete log_plugin.cpp with log file parsing
3. Complete postgres_plugin.cpp with libpq integration
4. Complete disk_usage_plugin.cpp with df parsing
5. Fix compiler warnings in collector.cpp

### Session 2: Integration & Testing
1. Complete sender.cpp error handling and retry logic
2. Complete main.cpp with config pull and hot-reload
3. Implement unit tests for all plugins
4. Fix integration test failures (TLS setup in Docker)
5. Run E2E tests with docker-compose

### Session 3: Validation & Documentation
1. Validate all success criteria
2. Create comprehensive documentation
3. Create PR and prepare for review
4. Merge to main branch

---

## Testing Strategy

### Current Test Results
```
✅ Unit Tests:        20/20 (100%)
❌ Integration Tests: 0/19 (0% - libcurl TLS limitation on macOS)
⏭️  E2E Tests:        Pending docker-compose
```

### Build Status
```
✅ Compilation: SUCCESSFUL
✅ Main Binary: ./src/pganalytics (builds without errors)
✅ Test Binary: ./tests/pganalytics-tests (builds without errors)
⚠️  Warnings: ~15 unused parameters/variables (will be fixed during implementation)
```

### Success Criteria
- [ ] All 4 metric types collecting from live PostgreSQL
- [ ] JSON output matches Phase 2 backend schema exactly
- [ ] Gzip compression ratio >40%
- [ ] TLS 1.3 enforced (no TLS 1.2)
- [ ] mTLS validates certificates properly
- [ ] JWT tokens used in all API calls
- [ ] Config pull works with Authentication header
- [ ] Hot-reload updates intervals without restart
- [ ] Exponential backoff retry on network failures
- [ ] Performance targets met
- [ ] Memory stable after 1000+ cycles
- [ ] 60+ unit tests (100% passing)
- [ ] 20+ integration tests (100% passing)
- [ ] E2E tests (100% passing with docker-compose)

---

## Files to Modify/Complete

### Source Files
1. `collector/src/postgres_plugin.cpp` - libpq integration (PRIORITY)
2. `collector/src/sysstat_plugin.cpp` - /proc parsing (PRIORITY)
3. `collector/src/log_plugin.cpp` - log file parsing (PRIORITY)
4. `collector/src/disk_usage_plugin.cpp` - df parsing (PRIORITY)
5. `collector/src/collector.cpp` - fix warnings, plugin management
6. `collector/src/main.cpp` - config pull, hot-reload, logging
7. `collector/src/sender.cpp` - error handling, retries
8. `collector/src/auth.cpp` - may need enhancements

### Test Files
1. `collector/tests/unit/postgres_plugin_test.cpp` - NEW
2. `collector/tests/unit/sysstat_plugin_test.cpp` - NEW
3. `collector/tests/unit/log_plugin_test.cpp` - NEW
4. `collector/tests/unit/disk_usage_plugin_test.cpp` - NEW
5. `collector/tests/integration/*_test.cpp` - existing, may need fixes

### Documentation
1. `collector/README.md` - NEW/UPDATE
2. `docs/COLLECTOR-ARCHITECTURE.md` - NEW
3. `docs/COLLECTOR-MIGRATION.md` - NEW
4. `docs/SECURITY-COLLECTOR.md` - NEW

---

## Code Reuse from v2

### PostgreSQL Collector
- **Source**: `coletor/src/collectors.cpp::PgStatsCollector`
- **Reuse**: SQL queries, connection logic, result parsing
- **Adapt**: Output format → JSON instead of COPY

### System Statistics Collector
- **Source**: `coletor/src/collectors.cpp::SysstatCollector`
- **Reuse**: /proc file parsing, calculations
- **Adapt**: Output format → JSON

### Log Collector
- **Source**: `coletor/src/collectors.cpp::PgLogCollector`
- **Reuse**: Log file parsing, state tracking
- **Adapt**: Output format → JSON

### Disk Usage Collector
- **Source**: `coletor/src/ServerInfo.cpp`
- **Reuse**: df command execution and parsing
- **Adapt**: Output format → JSON

---

## Next Steps

### Immediate Actions
1. ✅ Create this implementation status document
2. ➡️ Implement sysstat_plugin.cpp with /proc parsing
3. ➡️ Implement log_plugin.cpp with log file parsing
4. ➡️ Implement postgres_plugin.cpp with libpq integration
5. ➡️ Implement disk_usage_plugin.cpp with df parsing

### Session Goals
- Complete all plugin implementations
- Fix compiler warnings
- Ensure all 4 metric types collect properly
- Verify JSON output format
- Run unit tests successfully

---

## References

- **Phase 2 Backend Schema**: `/backend/internal/models/metrics.go`
- **Phase 3.4 Tests**: `/collector/tests/`
- **v2 Collector Code**: `/coletor/src/collectors.cpp`
- **Configuration Schema**: `/collector/config.toml.sample`

