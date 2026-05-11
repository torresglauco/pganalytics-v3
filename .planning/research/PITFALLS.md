# Pitfalls Research

**Domain:** PostgreSQL Monitoring Platform - Performance Optimization
**Researched:** 2026-05-11
**Confidence:** HIGH (based on codebase analysis and established PostgreSQL patterns)

## Critical Pitfalls

### Pitfall 1: Index Recommendation Breaks Existing Query Patterns

**What goes wrong:**
Creating recommended indexes changes PostgreSQL's query planner behavior. An index that improves one query can cause the planner to choose a worse execution plan for other queries that previously performed well. The planner may now use an index scan with many random I/O operations instead of a sequential scan that was actually faster.

**Why it happens:**
- PostgreSQL's cost-based optimizer estimates index scan costs, but these estimates can be wrong for specific data distributions
- Adding an index doesn't just affect queries that use it directly - it affects all queries the planner considers might benefit
- The existing system has query patterns that are implicitly optimized for current index state

**Consequences:**
- Existing dashboard queries suddenly slow down
- Query analysis features return stale results
- Index advisor recommendations may contradict each other
- Production users see degraded performance in previously fast operations

**How to avoid:**
1. Use `SET enable_seqscan = off` and `SET enable_indexscan = on` in test sessions to validate index impact before production
2. Run `EXPLAIN ANALYZE` on ALL existing queries (not just the target query) after adding indexes
3. Use partial indexes to limit scope of impact
4. Implement index changes in a staging environment with production-like data volumes first
5. Create a rollback script for every index change
6. Add tests that verify query execution plans haven't degraded

**Warning signs:**
- Query execution times increase after index creation
- `pg_stat_user_indexes.idx_scan` shows indexes being used but queries are slower
- `EXPLAIN` shows different node types than before (e.g., Hash Join changed to Nested Loop)
- Dashboard load times increase after optimization deployment

**Phase to address:**
Phase 1 (Query Optimization) - Must validate index impact across all query patterns before deployment

---

### Pitfall 2: Cache Invalidation Causes Stale Dashboard Data

**What goes wrong:**
Adding caching layers without proper invalidation logic causes dashboards to display outdated metrics. In a monitoring system, stale data is worse than slow data - users may make incorrect operational decisions based on outdated information.

**Why it happens:**
- Time-series data is continuously arriving from collectors
- Cache keys don't include all relevant parameters (time ranges, collector IDs, database filters)
- TTL-based expiration doesn't align with data freshness requirements
- No cache invalidation triggered when new metrics arrive
- The existing Manager has 5 cache types with different TTLs but no unified invalidation strategy

**Consequences:**
- Dashboard shows metrics from minutes/hours ago when real issues are occurring
- Alert conditions evaluated against stale cached data
- Users lose trust in monitoring accuracy
- Cache hit metrics look good while data quality degrades

**How to avoid:**
1. Include data freshness markers in cache keys (e.g., last metric timestamp)
2. Implement proactive cache invalidation when new metrics are inserted
3. Use cache-through pattern: populate cache only after write confirms success
4. Set TTLs based on metric collection intervals (observed in collector config)
5. Add cache hit/miss timing to metrics to detect stale cache hits
6. Implement cache warming that validates data freshness
7. Never cache real-time alert conditions - only historical aggregations

**Warning signs:**
- Dashboard data older than expected collection interval
- Cache hit rate near 100% during active incidents (should drop as users refresh)
- Metrics in database show timestamps newer than cached values
- Users report "data doesn't match" between views

**Phase to address:**
Phase 2 (API Response Optimization) - Cache invalidation strategy must be designed before caching is expanded

---

### Pitfall 3: Connection Pool Tuning Starves Concurrent Collectors

**What goes wrong:**
Adjusting connection pool settings (MaxOpenConns, MaxIdleConns, ConnMaxLifetime) causes connection exhaustion under load. The system currently supports 500+ collectors, but pool changes can break this scalability.

**Why it happens:**
- Current config: MaxOpenConns=100, MaxIdleConns=20 for main DB; TimescaleDB: 25/5
- Each collector needs periodic connections for metric push
- Reducing MaxOpenConns creates a bottleneck
- Increasing MaxIdleConns wastes resources
- Changing ConnMaxLifetime causes connection churn during peak times
- Connection establishment is expensive (TLS handshake, authentication)

**Consequences:**
- Collectors timeout when pushing metrics
- `database/sql` WaitCount increases
- Connection queue builds up, causing cascade delays
- Metrics gaps in time-series data
- Alert rules miss conditions due to missing data points

**How to avoid:**
1. Load test with realistic collector count (500+) before changing pool settings
2. Monitor `db.Stats()` before and after changes: WaitCount, WaitDuration, MaxOpenConnections
3. Keep MaxIdleConns at least 20-25% of MaxOpenConns to avoid connection thrash
4. Set ConnMaxLifetime > 5 minutes to avoid frequent reconnection storms
5. Use separate connection pools for different query types (OLTP vs time-series aggregation)
6. Implement connection pool metrics export for visibility
7. Never test with local single-collector setup and apply to production

**Warning signs:**
- `db.Stats().WaitCount` > 0 during normal operation
- Connection timeouts in logs during peak metric collection
- Latency spikes correlate with connection pool changes
- `pg_stat_activity` shows many idle connections but app reports exhaustion

**Phase to address:**
Phase 2 (API Response Optimization) - Pool changes must be validated with load testing before production

---

### Pitfall 4: Performance Measurement False Positives

**What goes wrong:**
Benchmark results show improvements that don't translate to real-world performance. Optimization appears successful in tests but provides no user-perceptible benefit.

**Why it happens:**
- Tests run with warm caches (data already in PostgreSQL shared_buffers)
- Single-threaded benchmarks miss lock contention effects
- Small data volumes don't represent production scale
- Query plans differ between test and production environments
- Missing `b.ResetTimer()` causes setup time to bias results
- Tests don't account for network latency between collector and backend
- Production has concurrent write traffic that tests don't simulate

**Consequences:**
- Time spent optimizing non-bottleneck code paths
- False confidence in performance improvements
- Regression only discovered in production
- Wasted engineering effort on low-impact changes

**How to avoid:**
1. Run benchmarks with `go test -run=^$ -bench=. -count=5` for statistical significance
2. Use `b.ResetTimer()` after setup, `b.ReportAllocs()` for memory analysis
3. Include concurrent benchmarks (see existing `BenchmarkCacheConcurrentReads` pattern)
4. Test with production-scale data volumes (copy from production anonymized)
5. Measure P50, P95, P99 latencies, not just average
6. Use `EXPLAIN (ANALYZE, BUFFERS)` to see actual I/O patterns
7. Run A/B tests in staging with realistic traffic patterns
8. Instrument actual user-facing endpoints, not just synthetic benchmarks

**Warning signs:**
- Benchmark shows 50% improvement but dashboard feels the same
- Production metrics show no improvement after optimization deployment
- Test queries execute in milliseconds, production queries take seconds
- Memory allocations in benchmarks differ from production profiles

**Phase to address:**
All phases - Each optimization must include measurement methodology

---

### Pitfall 5: TimescaleDB Aggregation Optimization Breaks Time Buckets

**What goes wrong:**
Optimizing time-series queries with TimescaleDB features (compression, continuous aggregates) breaks existing `time_bucket` queries used by dashboards.

**Why it happens:**
- TimescaleDB compression changes table storage format
- Continuous aggregates materialize data differently than ad-hoc aggregations
- `time_bucket` intervals must match between query and compression segment_by
- Timezone handling differences between raw and compressed data
- Existing queries assume uncompressed hypertable structure

**Consequences:**
- Dashboard charts show incorrect time boundaries
- Aggregation queries return errors on compressed chunks
- Gap-filling logic (`time_bucket_gapfill`) stops working
- Historical data queries fail or return partial results

**How to avoid:**
1. Test all existing `time_bucket` queries against compressed data before enabling compression
2. Create continuous aggregates only for specific, documented query patterns
3. Use `segment_by` and `order_by` compression settings that match query patterns
4. Keep uncompressed recent data (compression after 7+ days)
5. Add integration tests for all dashboard time-range queries
6. Document which time_bucket intervals are supported
7. Test timezone transitions (DST) with compressed data

**Warning signs:**
- Dashboard shows gaps at chunk boundaries
- Queries against old data suddenly slower (compression segment mismatch)
- `time_bucket` returns different results for same time range
- Error: "cannot compress chunk with foreign keys"

**Phase to address:**
Phase 3 (Dashboard Optimization) - TimescaleDB optimization requires careful compatibility testing

---

### Pitfall 6: Breaking Existing Tests During Optimization

**What goes wrong:**
Performance optimization changes break the existing comprehensive test suite (200+ backend tests, 38 database tests, 60+ frontend tests from v1.1).

**Why it happens:**
- Query plan changes cause assertion failures
- Timing assumptions in tests (timeouts, delays) no longer valid
- Cache behavior changes affect test isolation
- Connection pool changes affect test database cleanup
- Mock implementations don't match optimized code paths

**Consequences:**
- CI pipeline fails, blocking deployment
- Tests pass locally but fail in CI (different timing)
- Test flakiness increases
- Engineering time wasted debugging test failures vs. real bugs
- Regression in previously working features

**How to avoid:**
1. Run full test suite before and after each optimization change
2. Use testcontainers for isolated database tests (already in use)
3. Make tests wait for conditions rather than fixed timeouts
4. Reset caches between tests (`cache.Clear()` or `cache.Manager.Clear()`)
5. Use separate test database with consistent data setup
6. Parameterize timing values so tests can be adjusted without code changes
7. Add performance-specific tests that verify optimizations don't regress
8. Never skip failing tests - fix the underlying issue

**Warning signs:**
- Test failures after optimization that pass on main branch
- Intermittent test failures in CI
- Tests requiring multiple runs to pass
- Timeout errors in previously stable tests

**Phase to address:**
All phases - Every optimization PR must pass full test suite

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip load testing pool changes | Faster deployment | Production connection exhaustion | Never |
| Cache without TTL validation | Faster response times | Stale data, incorrect alerts | Never |
| Add index without query analysis | Quick fix for one slow query | Planner confusion, slower other queries | Only in emergency with rollback plan |
| Disable compression for speed | Faster writes | Storage explosion, slow historical queries | Only for testing, never production |
| Skip `EXPLAIN ANALYZE` on all queries | Faster optimization iteration | Hidden performance regressions | Never |
| Use `SELECT *` in cached queries | Simpler code | Memory waste, cache bloat | Never |
| Hardcode cache TTLs | Simpler implementation | Can't tune per-endpoint | MVP only, fix before production |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| PostgreSQL connection pool | Setting MaxOpenConns = MaxIdleConns (no scaling headroom) | Keep MaxIdleConns ~20-25% of MaxOpenConns |
| TimescaleDB queries | Using `time_bucket` without index on time column | Always have index on `time DESC` for hypertables |
| Cache layer | Caching query results without including all WHERE clause parameters | Include all filter values in cache key |
| Collector metrics | Assuming metric push order | Handle out-of-order timestamps in aggregations |
| WebSocket real-time updates | Sending updates for cached unchanged data | Compare before push, throttle update frequency |
| Database migrations | Running migrations concurrently with queries | Use advisory locks, run migrations in maintenance windows |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| In-memory cache without size limits | Memory exhaustion, OOM kills | Use LRU eviction (already implemented in cache.go) | ~10K items depending on value size |
| Connection per request | High latency, connection exhaustion | Use connection pool (already implemented) | ~100 concurrent users |
| Sequential metric inserts | Slow write throughput during peak | Use batch inserts, COPY command | ~1000 metrics/second |
| Single time_bucket interval | Slow queries on large time ranges | Use adaptive bucketing (hourly for week, daily for month) | ~1 year of data |
| Full table scans for count(*) | Slow dashboard load | Use estimated counts from pg_class for approximate | ~10M rows |
| N+1 queries in dashboard | Slow dashboard with many servers | Batch queries, use JOINs | ~50 servers |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Logging query parameters with sensitive data | Credential exposure in logs | Sanitize logs, use parameterized queries (already using) |
| Caching user-specific query results | Data leakage between users | Include user ID in cache key, clear on permission change |
| Index on encrypted column | Useless index, wasted storage | Index before encryption or use deterministic encryption |
| Cache timing attacks | Information disclosure | Use constant-time cache operations |
| Performance metrics in public API | System information disclosure | Rate limit, authenticate, redact sensitive details |

## UX Pitfalls

Common user experience mistakes in performance optimization.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Loading state not shown during optimization | Users think app is broken | Show skeleton/spinner immediately |
| Dashboard loads fast but data is stale | Users make wrong decisions | Show data timestamp, warn if stale |
| Pagination breaks with cached counts | Users see wrong total | Cache counts with expiration |
| Real-time updates stop after optimization | Users miss critical alerts | Maintain WebSocket heartbeat |
| Filter changes reset on navigation | Users lose context | Persist filters in URL state |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Index Optimization:** Often missing `ANALYZE` after index creation - verify `ANALYZE` was run on affected tables
- [ ] **Cache Implementation:** Often missing invalidation on write - verify cache is cleared/updated on data changes
- [ ] **Connection Pool Tuning:** Often missing monitoring - verify `db.Stats()` is exported and alerted on
- [ ] **Query Optimization:** Often missing production validation - verify `EXPLAIN ANALYZE` matches production plan
- [ ] **Dashboard Optimization:** Often missing mobile responsiveness - verify performance on smaller screens
- [ ] **Load Testing:** Often missing sustained load test - verify behavior over 30+ minutes, not just burst
- [ ] **TimescaleDB Optimization:** Often missing compression policy - verify compression job is scheduled

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Index broke queries | MEDIUM | 1. `DROP INDEX CONCURRENTLY` (safe), 2. Run `ANALYZE` on affected tables, 3. Verify query plans restored |
| Cache serving stale data | LOW | 1. Clear cache immediately (`cache.Clear()`), 2. Add invalidation logic, 3. Deploy with monitoring |
| Connection pool exhaustion | HIGH | 1. Restart affected services, 2. Reduce MaxOpenConns temporarily, 3. Add connection timeout alerts |
| Performance regression | MEDIUM | 1. Rollback deployment, 2. Profile production to find bottleneck, 3. Re-optimize with better measurement |
| TimescaleDB compression broke queries | HIGH | 1. Decompress affected chunks, 2. Fix query to work with compressed data, 3. Re-enable compression |
| Tests broken by optimization | LOW | 1. Identify timing assumptions, 2. Fix test to be timing-agnostic, 3. Re-run full suite |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Index breaks existing queries | Phase 1: Query Optimization | Run all existing queries with `EXPLAIN ANALYZE`, compare plans |
| Cache invalidation failures | Phase 2: API Response Optimization | Test: insert metric, verify cache cleared for affected dashboard |
| Connection pool starvation | Phase 2: API Response Optimization | Load test with 500+ concurrent collectors, monitor `db.Stats()` |
| Performance measurement false positives | All phases | Require `EXPLAIN (ANALYZE, BUFFERS)` output in PR description |
| TimescaleDB optimization breaks dashboards | Phase 3: Dashboard Optimization | Run all dashboard queries against compressed test data |
| Breaking existing tests | All phases | CI must pass full test suite (200+ backend, 38 DB, 60+ frontend) |

## Phase-Specific Warnings

### Phase 1: Query Optimization
- **Riskiest change:** Adding indexes to production tables
- **Required mitigation:** Test in staging with production data volume
- **Rollback plan:** Have `DROP INDEX CONCURRENTLY` scripts ready
- **Success metric:** P95 query latency improvement with no query plan regressions

### Phase 2: API Response Optimization
- **Riskiest change:** Connection pool and cache TTL adjustments
- **Required mitigation:** Load test with realistic collector count
- **Rollback plan:** Revert config changes via environment variables
- **Success metric:** P95 API latency improvement with no timeout errors

### Phase 3: Dashboard Optimization
- **Riskiest change:** TimescaleDB compression policies
- **Required mitigation:** Test all time_bucket queries with compressed chunks
- **Rollback plan:** Decompress affected chunks (`SELECT decompress_chunk()`)
- **Success metric:** Dashboard load time improvement with correct data display

## Sources

- PostgreSQL Official Documentation: Index Usage and Planner Behavior (postgresql.org/docs/current/indexes.html)
- TimescaleDB Documentation: Compression and Query Performance (docs.timescale.com)
- Go database/sql Connection Pooling Best Practices (pkg.go.dev/database/sql)
- Existing codebase patterns: `backend/internal/cache/`, `backend/internal/storage/postgres.go`, `backend/tests/database/connection_pool_test.go`
- Benchmark patterns: `backend/tests/benchmarks/caching_bench.go`
- Connection pool configuration: `backend/internal/storage/postgres.go` lines 39-77

---

*Pitfalls research for: Performance Optimization on PostgreSQL Monitoring Platform*
*Researched: 2026-05-11*
*Context: v1.2 Performance Optimization milestone for production system with 500+ collectors*