# Phase 03: Database Testing - Validation Strategy

**Phase:** 03 - Database Testing
**Goal:** Database operations are verified to handle edge cases and concurrent access reliably
**Created:** 2026-04-28

---

## Validation Architecture

This phase introduces database-level testing infrastructure using testcontainers. Validation strategy tracks test infrastructure setup and progressive verification of database reliability.

### Wave 0: Test Infrastructure Preparation

**Gap 1: Testcontainers Integration**
- **File:** `backend/tests/database/testdb_helpers.go`
- **Requirement:** Foundation for all DB tests
- **Why Important:** Isolated, reproducible database environment for all test phases
- **Validation Method:** `testdb.NewPostgresContainer()` creates isolated database with migrations applied
- **Success Criteria:** Container starts, migrations run, connection pool configured

**Gap 2: Fixture Management**
- **File:** `backend/tests/database/fixtures.go`
- **Requirement:** Seed data for consistent test scenarios
- **Why Important:** Enables repeatable test execution across developer machines
- **Validation Method:** Fixtures load without errors, data can be queried
- **Success Criteria:** InsertUser, InsertCollector, InsertMetrics work correctly

**Gap 3: Test Database Utilities**
- **File:** `backend/tests/database/assertions.go`
- **Requirement:** Assertion helpers for database verification
- **Why Important:** Makes database assertions explicit and readable
- **Validation Method:** AssertTransactionCommitted, AssertNoConnectionLeaks, etc. work as expected
- **Success Criteria:** Test assertions execute without panic, provide clear failure messages

### Wave 1: Database Layer Testing

**Dimension 1: Transaction Handling (TEST-07)**
- Transaction test suite created with testcontainers
- Tests validate: commit success, rollback recovery, nested transaction isolation
- Savepoint support for nested transactions
- Verification: All transaction scenarios pass with explicit assertions

**Dimension 2: Query Validation (TEST-08)**
- Edge case tests for empty results, null values, large datasets
- Query validation tests with 10,000+ row datasets
- Null byte handling and malformed data scenarios
- Verification: All query edge cases pass without errors

**Dimension 3: Connection Pool Management (TEST-09)**
- Concurrent load test simulating 100+ simultaneous connections
- Connection leak detection and pool exhaustion prevention
- Idle timeout and connection recycling validation
- Verification: Load test completes without connection errors

**Dimension 4: Schema Migrations (TEST-10)**
- Migration validation tests for all active migration files
- Zero data loss verification during schema changes
- Backward compatibility checks for schema evolution
- Verification: All active migrations pass, rollback works

**Dimension 5: Time-Series Data (TEST-11)**
- Time-series query tests with timezone conversions (UTC, PST, EST)
- Timestamp ordering validation across time zones
- TimescaleDB time_bucket function verification
- Verification: Time-series data maintains ordering and accuracy

### Verification Gates

**Gate 1: Infrastructure (Wave 0 → Wave 1)**
- [ ] Testcontainers Go module installed and working
- [ ] Database container starts successfully with migrations
- [ ] Test fixtures load without errors
- [ ] Assertion helpers available and functional
- **Pass Criteria:** All Wave 0 infrastructure components operational

**Gate 2: Transaction Testing (Wave 1 Part A)**
- [ ] Transaction test suite created and passing
- [ ] Commit/rollback scenarios verified
- [ ] Nested transaction savepoints working
- [ ] No uncommitted transaction leaks
- **Pass Criteria:** All transaction tests passing with explicit assertions

**Gate 3: Query & Connection Testing (Wave 1 Part B)**
- [ ] Query validation tests for edge cases passing
- [ ] Large dataset queries (10,000+ rows) working
- [ ] Connection pool load test completing successfully
- [ ] No connection leaks detected under 100+ concurrent connections
- **Pass Criteria:** All query and connection tests passing

**Gate 4: Migration & Time-Series Testing (Wave 1 Part C)**
- [ ] All active migration files validated
- [ ] Schema changes cause zero data loss
- [ ] Time-series queries working across timezones
- [ ] Timestamp ordering correct in all time zones
- **Pass Criteria:** All migration and time-series tests passing

**Gate 5: Final Verification (Post-Wave 1)**
- [ ] Full database test suite passes: `go test ./backend/tests/database -v`
- [ ] No flaky tests (run suite 3x to verify)
- [ ] All 5 requirements verified as PASS
- [ ] Coverage includes edge cases and failure scenarios
- **Pass Criteria:** Green test suite, no flaky tests, all requirements met

---

## Risk Mitigation

**Risk 1: Testcontainers startup overhead**
- **Mitigation:** Reuse container across test suites within a test run
- **Fallback:** Single dedicated test database (slower but simpler)

**Risk 2: Timezone conversion complexity**
- **Mitigation:** Use Go's time package functions (UTC, In(), Format())
- **Fallback:** Skip timezone tests, validate single timezone

**Risk 3: Large dataset performance**
- **Mitigation:** Use indexes in migration fixtures, limit dataset to 10,000 rows
- **Fallback:** Reduce dataset size or skip performance assertions

**Risk 4: Migration backward compatibility**
- **Mitigation:** Test down-migration (if available) or verify data structure
- **Fallback:** Accept forward-only migrations, document assumptions

---

## Success Criteria by Dimension

| Dimension | Criteria | Verification |
|-----------|----------|--------------|
| **Infrastructure** | testcontainers working, fixtures load | `go test ./backend/tests/database -v` |
| **Transactions** | Commits, rollbacks, savepoints working | Transaction test assertions |
| **Queries** | Edge cases, large datasets passing | Query validation test results |
| **Connections** | Pool handles 100+ concurrent connections | Load test completion |
| **Migrations** | Zero data loss, backward compatible | Migration validation tests |
| **Time-Series** | Correct ordering across timezones | Time-series assertion results |

---

## Handoff to Phase 4

Phase 4 (Frontend Integration Testing & Quality) depends on:
- ✓ Phase 3 complete with all database tests passing
- ✓ Database infrastructure stable and reliable
- ✓ Schema migrations verified
- ✓ Connection pooling working correctly

Phase 4 will:
- Test frontend components against verified database layer
- Implement frontend integration tests
- Maintain overall code quality standards

---

*Validation Strategy Created: 2026-04-28*
*Phase 03 Blocker Resolution: Nyquist compliance documentation*
