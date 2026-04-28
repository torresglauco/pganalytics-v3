# Phase 3: Database Testing - Research

**Researched:** 2026-04-28
**Domain:** PostgreSQL database operations testing (transactions, connection pools, migrations, time-series)
**Confidence:** HIGH

## Summary

Phase 3 focuses on validating the database layer's reliability through comprehensive testing of transaction handling, query validation, connection pool management, schema migrations, and time-series data operations. The codebase already has foundational database infrastructure (`internal/storage/postgres.go`, `internal/timescale/timescale.go`, `internal/storage/migrations.go`) but lacks dedicated test coverage.

The existing transaction pattern uses `BeginTx` with deferred rollback (see `metrics_store.go`), connection pools are configured with `SetMaxOpenConns/SetMaxIdleConns`, and migrations use a custom runner with version tracking. Testing requires a real PostgreSQL instance - the project uses `github.com/lib/pq` v1.10.9 as the PostgreSQL driver.

**Primary recommendation:** Use testcontainers-go for isolated PostgreSQL test instances, with testify for assertions. Create database-specific test helpers for transaction, pool, migration, and time-series testing.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TEST-07 | Transaction handling (commits, rollbacks, nested) | Existing `BeginTx` pattern in `metrics_store.go` (lines 23-107, 153-199, 245-291, 336-382, 427-492); defer rollback pattern established |
| TEST-08 | Query validation (edge cases, null values, large datasets) | TimescaleDB query patterns in `timescale.go`; null handling with `sql.NullFloat64` (lines 277-278); large dataset handling needed |
| TEST-09 | Connection pool management under load | Pool configuration in `postgres.go` (lines 39-56): MaxOpenConns=100, MaxIdleConns=20; environment variable override support |
| TEST-10 | Schema migrations validation | `migrations.go` has MigrationRunner with version tracking; 27 migration files in `backend/migrations/` |
| TEST-11 | Time-series data handling | `timescale.go` has `QueryMetricsRange`, `AggregateMetrics`, `GetMetricsCount` using `time_bucket` function |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/lib/pq | v1.10.9 | PostgreSQL driver | Already in use; pure Go; supports prepared statements, transactions |
| github.com/stretchr/testify | v1.11.1 | Test assertions and mocks | Already in use; industry standard; `assert`, `require`, `mock` packages |
| github.com/testcontainers/testcontainers-go | latest | Isolated PostgreSQL instances for tests | Industry standard for integration tests; Docker-based; supports PostgreSQL |
| go.uber.org/zap | v1.27.0 | Structured logging | Already in use for migration runner |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/testcontainers/testcontainers-go/modules/postgres | latest | PostgreSQL module for testcontainers | All database integration tests |
| github.com/google/uuid | v1.6.0 | UUID generation | Already in use for collector IDs |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| testcontainers-go | Docker Compose | Docker Compose requires external setup; testcontainers is self-contained per test |
| testcontainers-go | Embedded PostgreSQL | Embedded requires CGO; testcontainers works cross-platform |
| lib/pq | pgx | pgx has better performance but lib/pq is already integrated and stable |

**Installation:**
```bash
cd backend
go get github.com/testcontainers/testcontainers-go github.com/testcontainers/testcontainers-go/modules/postgres
```

**Version verification:**
```
github.com/lib/pq v1.10.9 (verified 2026-04-28)
github.com/stretchr/testify v1.11.1 (verified 2026-04-28)
go.uber.org/zap v1.27.0 (verified 2026-04-28)
```

## Architecture Patterns

### Recommended Project Structure
```
backend/tests/database/
├── testutil/
│   ├── container.go       # testcontainers setup
│   ├── fixtures.go        # test data factories
│   └── helpers.go         # assertion helpers
├── transaction_test.go    # TEST-07
├── query_test.go          # TEST-08
├── connection_pool_test.go # TEST-09
├── migration_test.go      # TEST-10
└── timeseries_test.go     # TEST-11
```

### Pattern 1: Transaction Handling Test Pattern
**What:** Test commit success, rollback recovery, and nested transaction isolation
**When to use:** All database operations that modify data
**Example:**
```go
// Source: Existing pattern in metrics_store.go, adapted for testing
func TestTransactionCommit(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)
    defer tx.Rollback() // Safe rollback if commit fails

    _, err = tx.ExecContext(ctx, "INSERT INTO test_table (id) VALUES ($1)", 1)
    require.NoError(t, err)

    err = tx.Commit()
    require.NoError(t, err)

    // Verify data persisted
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 1, count)
}

func TestTransactionRollback(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)

    _, err = tx.ExecContext(ctx, "INSERT INTO test_table (id) VALUES ($1)", 1)
    require.NoError(t, err)

    err = tx.Rollback()
    require.NoError(t, err)

    // Verify data was NOT persisted
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 0, count)
}
```

### Pattern 2: Connection Pool Test Pattern
**What:** Test pool behavior under concurrent load, verify no leaks
**When to use:** Validating pool configuration before production
**Example:**
```go
func TestConnectionPoolUnderLoad(t *testing.T) {
    db := testutil.NewTestDB(t)
    defer db.Close()

    // Configure pool for testing
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)

    var wg sync.WaitGroup
    errors := make(chan error, 100)

    // Simulate 100 concurrent connections (more than pool size)
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

            var result int
            err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
            if err != nil {
                errors <- err
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    // All queries should succeed despite pool limit
    for err := range errors {
        t.Errorf("Query failed: %v", err)
    }

    // Verify no connection leaks
    stats := db.Stats()
    assert.Equal(t, 0, stats.WaitCount, "No connections should be waiting")
}
```

### Pattern 3: Migration Validation Pattern
**What:** Test migrations don't lose data and maintain backward compatibility
**When to use:** Before deploying schema changes
**Example:**
```go
func TestMigrationNoDataLoss(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    // Insert test data before migration
    _, err := db.ExecContext(ctx,
        "INSERT INTO pganalytics.databases (name, host, port) VALUES ($1, $2, $3)",
        "testdb", "localhost", 5432)
    require.NoError(t, err)

    // Run migration
    runner := storage.NewMigrationRunner(db, zap.NewNop())
    err = runner.Run(ctx)
    require.NoError(t, err)

    // Verify data still exists
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM pganalytics.databases WHERE name = $1", "testdb").Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 1, count, "Data should be preserved after migration")
}
```

### Anti-Patterns to Avoid
- **Testing without cleanup:** Always use `t.Cleanup()` or deferred cleanup to prevent test pollution
- **Hardcoded connection strings:** Use testcontainers or environment variables
- **Ignoring `sql.ErrNoRows`:** Handle empty result sets explicitly, not as errors
- **Not testing concurrent access:** Database tests must verify thread safety
- **Skipping rollback tests:** Rollback is critical for error recovery; always test it

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Test database setup | Custom Docker scripts | testcontainers-go | Automatic lifecycle, port mapping, cleanup |
| SQL statement parsing | Custom parser | Existing `splitSQLStatements` in migrations.go | Already handles dollar-quoted strings |
| Connection pool configuration | Custom pool | `db.SetMaxOpenConns/SetMaxIdleConns` | Built into database/sql |
| Migration versioning | Custom version table | `schema_versions` table in migrations.go | Already implemented |
| Time-series aggregation | Custom bucketing | TimescaleDB `time_bucket` function | Already in use |

**Key insight:** The codebase already has solid infrastructure. Focus testing on behavior validation, not infrastructure reimplementation.

## Common Pitfalls

### Pitfall 1: Connection Pool Exhaustion
**What goes wrong:** Tests leave connections open, pool fills, subsequent tests fail
**Why it happens:** Not closing rows, statements, or transactions
**How to avoid:** Always `defer rows.Close()`, `defer stmt.Close()`, use `t.Cleanup()` for DB teardown
**Warning signs:** `pq: sorry, too many clients already` errors

### Pitfall 2: Transaction Rollback Not Tested
**What goes wrong:** Errors in transactions don't roll back, leaving partial data
**Why it happens:** Developers assume rollback works; defer pattern obscures rollback logic
**How to avoid:** Explicit rollback tests with assertions that data was NOT persisted
**Warning signs:** Tests pass but production has orphaned records after failures

### Pitfall 3: Timezone Handling Issues
**What goes wrong:** Timestamps appear in wrong timezone, ordering breaks
**Why it happens:** PostgreSQL uses server timezone, Go uses local time, mixing causes bugs
**How to avoid:** Always store in UTC, use `TIMESTAMP WITH TIME ZONE`, test with explicit timezones
**Warning signs:** Time comparisons fail across DST boundaries

### Pitfall 4: Large Dataset Query Memory
**What goes wrong:** Query returns 10,000+ rows, test runs out of memory
**Why it happens:** Tests load all results into memory instead of streaming
**How to avoid:** Use cursor-based iteration, `LIMIT` in tests, or `sql.Rows` streaming
**Warning signs:** Tests slow down, OOM errors on large datasets

### Pitfall 5: Migration Order Dependency
**What goes wrong:** Tests assume migrations run in specific order, fail when new migrations added
**Why it happens:** Hardcoded migration names, not using version tracking
**How to avoid:** Use `schema_versions` table for idempotency, test from empty database state
**Warning signs:** Tests fail after adding new migration files

## Code Examples

### Transaction Handling with Savepoints (Nested Transactions)
```go
// Source: PostgreSQL savepoint pattern for nested transactions
func TestNestedTransactionRollback(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)
    defer tx.Rollback()

    // Outer insert
    _, err = tx.ExecContext(ctx, "INSERT INTO test_table (id) VALUES (1)")
    require.NoError(t, err)

    // Create savepoint
    _, err = tx.ExecContext(ctx, "SAVEPOINT inner_txn")
    require.NoError(t, err)

    // Inner insert
    _, err = tx.ExecContext(ctx, "INSERT INTO test_table (id) VALUES (2)")
    require.NoError(t, err)

    // Rollback to savepoint
    _, err = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT inner_txn")
    require.NoError(t, err)

    // Commit outer transaction
    err = tx.Commit()
    require.NoError(t, err)

    // Only outer insert should persist
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 1, count, "Only outer transaction should persist")
}
```

### Null Value Handling
```go
// Source: Pattern from timescale.go lines 277-278
func TestNullValueHandling(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    // Insert with NULL
    _, err := db.ExecContext(ctx,
        "INSERT INTO metrics (value, label) VALUES (NULL, $1)", "test")
    require.NoError(t, err)

    // Query with NULL handling
    rows, err := db.QueryContext(ctx, "SELECT value FROM metrics WHERE label = $1", "test")
    require.NoError(t, err)
    defer rows.Close()

    for rows.Next() {
        var value sql.NullFloat64
        err := rows.Scan(&value)
        require.NoError(t, err)

        assert.False(t, value.Valid, "Value should be NULL")
        // Access with value.Float64 (returns 0 if NULL)
    }
}
```

### Large Dataset Streaming
```go
func TestLargeDatasetStreaming(t *testing.T) {
    ctx := context.Background()
    db := testutil.NewTestDB(t)
    defer db.Close()

    // Insert 10,000 rows
    for i := 0; i < 10000; i++ {
        _, err := db.ExecContext(ctx, "INSERT INTO large_table (id) VALUES ($1)", i)
        require.NoError(t, err)
    }

    // Stream results instead of loading all into memory
    rows, err := db.QueryContext(ctx, "SELECT id FROM large_table ORDER BY id")
    require.NoError(t, err)
    defer rows.Close()

    count := 0
    prevID := -1
    for rows.Next() {
        var id int
        err := rows.Scan(&id)
        require.NoError(t, err)
        assert.Equal(t, prevID+1, id, "IDs should be sequential")
        prevID = id
        count++
    }

    assert.Equal(t, 10000, count, "Should process all rows")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Docker Compose for test DBs | testcontainers-go | 2020+ | Self-contained, automatic cleanup |
| Custom assertion libraries | stretchr/testify | 2014+ | De facto standard |
| Mock databases for unit tests | Real DB with transactions | Ongoing | Catches actual SQL errors |
| Hardcoded timestamps | `time.Now().UTC()` | Best practice | Timezone-safe |

**Deprecated/outdated:**
- `sqlmock`: Limited for integration tests; testcontainers preferred for real DB behavior
- `pgx` mock: Not needed; use real PostgreSQL via testcontainers

## Open Questions

1. **Test database isolation strategy**
   - What we know: testcontainers supports PostgreSQL, can use ephemeral containers
   - What's unclear: Should each test get its own database or share one with transaction rollback?
   - Recommendation: Share container across tests, use separate databases for parallel tests

2. **Migration test coverage scope**
   - What we know: 27 migration files exist, some disabled
   - What's unclear: Should we test all migrations or just active ones?
   - Recommendation: Test all `.sql` files (not `.disabled`), focus on active migrations first

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify v1.11.1 |
| Config file | None (testcontainers programmatically) |
| Quick run command | `go test ./tests/database/... -short` |
| Full suite command | `go test ./tests/database/... -count=1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TEST-07 | Transaction commit success | unit | `go test ./tests/database/transaction_test.go -run TestTransactionCommit -v` | No - Wave 0 |
| TEST-07 | Transaction rollback recovery | unit | `go test ./tests/database/transaction_test.go -run TestTransactionRollback -v` | No - Wave 0 |
| TEST-07 | Nested transaction isolation | unit | `go test ./tests/database/transaction_test.go -run TestNestedTransaction -v` | No - Wave 0 |
| TEST-08 | Empty result set handling | unit | `go test ./tests/database/query_test.go -run TestEmptyResultSet -v` | No - Wave 0 |
| TEST-08 | NULL value handling | unit | `go test ./tests/database/query_test.go -run TestNullHandling -v` | No - Wave 0 |
| TEST-08 | Large dataset query (10K+ rows) | unit | `go test ./tests/database/query_test.go -run TestLargeDataset -v` | No - Wave 0 |
| TEST-09 | Connection pool under concurrent load (100+) | integration | `go test ./tests/database/connection_pool_test.go -run TestConnectionPoolLoad -v` | No - Wave 0 |
| TEST-09 | No connection leaks | integration | `go test ./tests/database/connection_pool_test.go -run TestNoLeaks -v` | No - Wave 0 |
| TEST-10 | Migration zero data loss | integration | `go test ./tests/database/migration_test.go -run TestMigrationDataPreservation -v` | No - Wave 0 |
| TEST-10 | Backward compatibility | integration | `go test ./tests/database/migration_test.go -run TestBackwardCompatibility -v` | No - Wave 0 |
| TEST-11 | Timestamp ordering UTC/PST/EST | unit | `go test ./tests/database/timeseries_test.go -run TestTimezoneOrdering -v` | No - Wave 0 |
| TEST-11 | Time bucket aggregation | unit | `go test ./tests/database/timeseries_test.go -run TestTimeBucket -v` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./tests/database/... -short` (skip slow tests)
- **Per wave merge:** `go test ./tests/database/... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `tests/database/testutil/container.go` - testcontainers PostgreSQL setup
- [ ] `tests/database/testutil/fixtures.go` - test data factories
- [ ] `tests/database/testutil/helpers.go` - assertion helpers
- [ ] `tests/database/transaction_test.go` - covers TEST-07
- [ ] `tests/database/query_test.go` - covers TEST-08
- [ ] `tests/database/connection_pool_test.go` - covers TEST-09
- [ ] `tests/database/migration_test.go` - covers TEST-10
- [ ] `tests/database/timeseries_test.go` - covers TEST-11
- [ ] Framework install: `go get github.com/testcontainers/testcontainers-go/modules/postgres`

## Sources

### Primary (HIGH confidence)
- Existing codebase: `backend/internal/storage/postgres.go` - connection pool configuration
- Existing codebase: `backend/internal/storage/metrics_store.go` - transaction patterns
- Existing codebase: `backend/internal/storage/migrations.go` - migration runner
- Existing codebase: `backend/internal/timescale/timescale.go` - time-series operations
- Existing test patterns: `backend/tests/integration/migrations_schema_test.go` - schema validation

### Secondary (MEDIUM confidence)
- Go database/sql documentation - standard patterns
- lib/pq v1.10.9 driver documentation - PostgreSQL driver features
- testcontainers-go documentation - container setup patterns

### Tertiary (LOW confidence)
- None - recommendations based on existing codebase patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use except testcontainers, which is industry standard
- Architecture: HIGH - Existing patterns in codebase provide clear guidance
- Pitfalls: HIGH - Based on common Go/database testing issues and codebase analysis

**Research date:** 2026-04-28
**Valid until:** 30 days - stable patterns, testcontainers API stable