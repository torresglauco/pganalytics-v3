# pgAnalytics v3.3.0 - Performance Optimization Roadmap
**Date**: February 26, 2026
**Status**: Ready for Implementation
**Based On**: LOAD_TEST_REPORT_FEB_2026.md

---

## Overview

This roadmap outlines the implementation strategy to fix 6 identified bottlenecks and enable pgAnalytics to scale from 10-25 collectors (current) to 500+ collectors (optimized).

### Timeline & Effort

- **Phase 1 (CRITICAL)**: 30-36 hours - Fix fundamental bottlenecks
- **Phase 2 (HIGH)**: 22-30 hours - Enable scale to 100+ collectors
- **Phase 3 (MEDIUM)**: 28-38 hours - Optimize protocol and operations
- **Total**: 80-104 hours

### Impact

```
Current State (10-25 collectors)
  ↓ Phase 1 (CRITICAL fixes)
Improved State (25-100 collectors)
  ↓ Phase 2 (HIGH priority)
Scaled State (100-200 collectors)
  ↓ Phase 3 (MEDIUM optimization)
Enterprise State (500+ collectors with horizontal scaling)
```

---

## Phase 1: CRITICAL Bottleneck Fixes (30-36 hours)

### Task 1.1: Implement Thread Pool for Collectors (20-30 hours)

**Current Issue**: Single-threaded main loop prevents parallel collector execution

**Solution**: Use std::thread_pool with 4-8 worker threads

**Files to Modify**:
- `collector/src/main.cpp` - Rewrite main collection loop
- `collector/include/collector.h` - Add ThreadPool class
- `collector/src/collector.cpp` - Implement thread pool management

**Implementation Steps**:

1. **Create ThreadPool Utility** (4 hours)
   ```cpp
   // collector/include/thread_pool.h
   class ThreadPool {
       std::vector<std::thread> workers;
       std::queue<std::function<void()>> tasks;
       std::mutex queue_mutex;
       std::condition_variable condition;

   public:
       explicit ThreadPool(size_t threads) { ... }
       template<class F, class... Args>
       void enqueue(F&& f, Args&&... args) { ... }
       ~ThreadPool();
   };
   ```

2. **Modify CollectorManager** (6 hours)
   ```cpp
   // Changes to executeAllCollectors()
   void CollectorManager::executeAllCollectors() {
       ThreadPool pool(4);

       for (auto& collector : collectors_) {
           pool.enqueue([collector]() {
               collector->execute();
           });
       }

       pool.wait_all();  // Wait for all to complete
   }
   ```

3. **Update Main Loop** (6 hours)
   ```cpp
   // In runCronMode():
   // Loop remains same, but collectors execute in parallel
   while (!shouldExit) {
       auto start = std::chrono::steady_clock::now();

       collectorMgr.executeAllCollectors();  // Now parallelized

       // ... rest of loop
   }
   ```

4. **Add Thread Pool Monitoring** (4 hours)
   - Thread pool size metrics
   - Queue depth monitoring
   - Task execution time tracking
   - Thread utilization reporting

**Expected Impact**:
- Sequential: 100 collectors × 577ms = 57.7s per cycle
- Parallel (4 threads): 100 collectors ÷ 4 = 25 cycles = ~14.4s per cycle
- **Net improvement: 75% reduction in collection time**

**Testing**:
- Unit test: ThreadPool with 10 dummy tasks
- Integration test: 4 collectors with various latencies
- Load test: 100 collectors with thread pool vs without

**Acceptance Criteria**:
- ✅ 4 collectors execute in parallel
- ✅ Total cycle time < 15 seconds for 100 collectors
- ✅ No race conditions or deadlocks
- ✅ Thread pool metrics available
- ✅ All existing tests pass

---

### Task 1.2: Make Query Limit Configurable (2-4 hours)

**Current Issue**: Hard-coded LIMIT 100 in pg_stat_statements query

**Solution**: Read from configuration file with intelligent defaults

**Files to Modify**:
- `collector/src/query_stats_plugin.cpp` - Query generation
- `collector/config/collector.conf` - Configuration template
- `collector/src/config_manager.cpp` - Config loading

**Implementation Steps**:

1. **Update Configuration Schema** (1 hour)
   ```ini
   # collector/config/collector.conf
   [postgresql]
   # Query statistics collection limit
   # At 1000 QPS: query_stats_limit=100 → 10% collection
   # At 10000 QPS: query_stats_limit=500 → 5% collection
   # At 100000 QPS: query_stats_limit=1000 → 1% collection
   query_stats_limit=100
   ```

2. **Modify Query Generation** (1 hour)
   ```cpp
   // In query_stats_plugin.cpp
   int limit = gConfig->getInt("postgresql", "query_stats_limit", 100);

   std::string query = fmt::format(
       "SELECT {} FROM pg_stat_statements "
       "ORDER BY total_time DESC "
       "LIMIT {} ",
       select_clause, limit
   );
   ```

3. **Add Validation** (1 hour)
   - Min value: 10 (don't allow < 10)
   - Max value: 10000 (don't allow > 10000)
   - Warn if limit < total unique queries
   - Log actual vs limit ratio

4. **Add Monitoring** (1 hour)
   ```cpp
   // Metrics to emit:
   // - collector_query_stats_limit
   // - collector_query_stats_collected
   // - collector_query_stats_sampling_percent
   ```

**Expected Impact**:
- Allows adaptive sampling based on database load
- Production database can set limit=5000
- Low-volume databases can set limit=100
- **Enables accurate data collection at scale**

**Testing**:
- Unit test: Config validation (10, 100, 10000, invalid values)
- Integration test: Query with limit=500 vs limit=100
- Load test: Monitor sampling % at various limits

**Acceptance Criteria**:
- ✅ Config option read successfully
- ✅ Values 10-10000 accepted, others rejected
- ✅ Default remains 100
- ✅ Query executed with correct LIMIT value
- ✅ Metrics show sampling percentage

---

### Task 1.3: Implement Connection Pooling (8-12 hours)

**Current Issue**: Each collection creates fresh PostgreSQL connection (100-500ms overhead)

**Solution**: Persistent connection pool with min/max connections

**Files to Modify**:
- `collector/include/connection_pool.h` - New header
- `collector/src/connection_pool.cpp` - Implementation
- `collector/src/query_stats_plugin.cpp` - Use pool instead of new conn
- `collector/src/main.cpp` - Initialize pool at startup

**Implementation Steps**:

1. **Create ConnectionPool Class** (4 hours)
   ```cpp
   // collector/include/connection_pool.h
   class ConnectionPool {
       std::vector<std::unique_ptr<pqxx::connection>> available;
       std::queue<pqxx::connection*> idle_queue;
       std::mutex pool_mutex;
       std::condition_variable condition;

       const std::string connection_string;
       const int min_connections = 2;
       const int max_connections = 10;

   public:
       std::unique_ptr<pqxx::connection> acquire();  // RAII borrow
       void release(pqxx::connection* conn);
       void close_all();
   };
   ```

2. **Implement RAII Connection Borrower** (2 hours)
   ```cpp
   class ConnectionBorrow {
       ConnectionPool& pool;
       pqxx::connection* conn;
   public:
       ConnectionBorrow(ConnectionPool& p) : pool(p), conn(p.acquire()) {}
       ~ConnectionBorrow() { if(conn) pool.release(conn); }

       pqxx::connection* operator->() { return conn; }
   };
   ```

3. **Integrate with Collectors** (4 hours)
   ```cpp
   // In query_stats_plugin.cpp
   // OLD: pqxx::connection conn(connStr);
   // NEW: ConnectionBorrow conn(pool_);

   for (const auto& db : databases_) {
       ConnectionBorrow conn(pool_);
       auto txn = conn->prepare("query").exec();
   }
   ```

4. **Add Pool Monitoring** (2 hours)
   - Available connections in pool
   - Idle vs active ratio
   - Connection age (timeout stale)
   - Pool size growth/shrinkage

**Expected Impact**:
- Connection overhead: 100-500ms → 5-10ms (reuse)
- **50% cycle time reduction**
- At 50 collectors: 28.85s → 14.4s per cycle

**Testing**:
- Unit test: Pool size management (acquire/release)
- Integration test: 10 concurrent requests on pool of 5
- Load test: Monitor pool metrics under 100 collector load
- Stress test: Exhaust pool, verify queueing

**Acceptance Criteria**:
- ✅ Min 2 connections available
- ✅ Max 10 connections (configurable)
- ✅ Acquire with timeout (don't block forever)
- ✅ Auto-reconnect on stale connection
- ✅ No connection leak
- ✅ Pool metrics exposed

---

### Phase 1 Summary

**Total Effort**: 30-46 hours
**Core Work**: Rewrite main collection loop architecture
**Measurable Impact**:

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| 100 collectors cycle time | 57.7s | 14.4s | 75% ↓ |
| CPU usage at 100 collectors | 96% | 36% | 62% ↓ |
| Connection overhead | 200-400ms | 5-10ms | 95% ↓ |
| Query sampling at 10K QPS | 1% | 5-10% | 5-10x ↑ |

**Outcome**: System becomes viable for 100-200 collectors

---

## Phase 2: HIGH Priority Enhancements (22-30 hours)

### Task 2.1: Eliminate Triple JSON Serialization (12-16 hours)

**Current Issue**: JSON serialized 3 times before transmission (75-150ms CPU)

**Solution**: Binary intermediate format, serialize once

**Files to Modify**:
- `collector/src/metrics_serializer.cpp` - Add binary format
- `collector/src/metrics_buffer.cpp` - Store binary, not JSON
- `collector/src/sender.cpp` - Deserialize binary → JSON for API

**Implementation Steps**:

1. **Define Binary Format** (4 hours)
   ```cpp
   struct BinaryMetric {
       uint16_t type_id;           // Metric type
       uint32_t timestamp;         // Unix timestamp
       uint16_t field_count;       // Number of fields
       struct Field {
           uint8_t type;           // Field type (int, float, string)
           uint16_t name_len;
           char name[];
           union {
               int64_t i64;
               double f64;
               struct { uint16_t len; char data[]; } str;
           } value;
       } fields[];
   };
   ```

2. **Implement Binary Serialization** (4 hours)
   - Collector outputs BinaryMetric directly
   - No intermediate JSON
   - Buffer stores binary data
   - Saves 25-50ms per cycle

3. **Update Buffer** (2 hours)
   - Change from JSON array to binary stream
   - Support both formats temporarily
   - Add version header for future compatibility

4. **Update Sender** (2 hours)
   - Deserialize binary back to JSON for REST API
   - Keep backward compatibility with API
   - Could use binary protocol in future

**Expected Impact**:
- Serialization time: 75-150ms → 15-30ms (80% reduction)
- CPU savings: 30-50ms per cycle
- No network bandwidth change (still compress)

**Testing**:
- Unit test: Binary serialization roundtrip
- Integration test: 50 metrics serialized/deserialized
- Load test: Measure serialization CPU impact
- Compatibility test: Old JSON format still accepted

**Acceptance Criteria**:
- ✅ JSON serialization reduced by 75%
- ✅ No data loss in binary format
- ✅ Can serialize 1000 metrics/second
- ✅ Backward compatible with existing API
- ✅ All tests pass

---

### Task 2.2: Add Buffer Overflow Monitoring (4-6 hours)

**Current Issue**: Silent metric discarding when buffer full

**Solution**: Logging, metrics, and graceful degradation

**Files to Modify**:
- `collector/src/metrics_buffer.cpp` - Add monitoring
- `collector/src/main.cpp` - Log warnings
- `collector/src/sender.cpp` - Emit metrics about overflow

**Implementation Steps**:

1. **Add Metrics Tracking** (2 hours)
   ```cpp
   class MetricsBuffer {
       size_t max_size_reached = 0;
       size_t metrics_discarded = 0;

   public:
       bool append(const json& metrics) {
           if (will_overflow) {
               metrics_discarded++;
               std::cerr << "[WARNING] Buffer full, discarding metric";
               return false;  // Still fail, but now we tracked it
           }
           // ...
       }
   };
   ```

2. **Add Logging** (1 hour)
   - WARNING when buffer > 80% capacity
   - ERROR when metrics discarded
   - DEBUG: track individual metric sizes

3. **Emit Metrics** (1 hour)
   ```cpp
   {
       "type": "collector_metrics",
       "buffer_used_bytes": 40000000,
       "buffer_max_bytes": 50000000,
       "buffer_used_percent": 80,
       "metrics_discarded_count": 5,
       "timestamp": "2026-02-26T15:30:00Z"
   }
   ```

4. **Optional: Priority-Based Eviction** (2 hours - if time permits)
   - Keep high-priority metrics (CPU, memory)
   - Discard low-priority metrics (disk I/O detail)
   - Configurable priority levels

**Expected Impact**:
- Visibility into data loss
- Can alert on overflow events
- **Prevent silent failures**

**Testing**:
- Unit test: Overflow detection
- Integration test: Fill buffer, verify logging
- Load test: Monitor discarded metrics count
- Verify metrics emitted correctly

**Acceptance Criteria**:
- ✅ Buffer overflow detected
- ✅ Warnings logged to stderr
- ✅ Metrics emitted about overflow
- ✅ Can track total discarded metrics
- ✅ No performance regression

---

### Task 2.3: Implement Rate Limiting (6-8 hours)

**Current Issue**: No rate limiting on metrics push endpoint (operational risk)

**Solution**: Token bucket rate limiting per collector

**Files to Modify**:
- `backend/internal/api/middleware.go` - Implement RateLimitMiddleware
- `backend/internal/config/config.go` - Rate limit config
- `backend/internal/api/handlers.go` - Apply middleware to routes

**Implementation Steps**:

1. **Define Rate Limit Config** (1 hour)
   ```go
   type RateLimitConfig struct {
       Requests     int    // Requests per window
       WindowSize   int    // Window in seconds
       BurstSize    int    // Allow bursts up to this
       BackoffTime  int    // Seconds to wait after limit hit
   }
   ```

2. **Implement Token Bucket** (3 hours)
   ```go
   type TokenBucket struct {
       capacity    float64
       tokens      float64
       refillRate  float64
       lastRefill  time.Time
       mutex       sync.Mutex
   }

   func (tb *TokenBucket) TryAcquire(tokens float64) bool {
       tb.refill()
       if tb.tokens >= tokens {
           tb.tokens -= tokens
           return true
       }
       return false
   }
   ```

3. **Implement Middleware** (2 hours)
   ```go
   func RateLimitMiddleware(limits RateLimitConfig) http.Middleware {
       buckets := make(map[string]*TokenBucket)

       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               collectorID := getCollectorID(r)
               bucket := getBucket(buckets, collectorID)

               if !bucket.TryAcquire(1) {
                   w.Header().Set("Retry-After", "60")
                   http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                   return
               }

               next.ServeHTTP(w, r)
           })
       }
   }
   ```

4. **Apply to Routes** (2 hours)
   - `/api/v1/metrics/push` - 100 req/min per collector
   - `/api/v1/collectors/register` - 5 req/min per IP
   - `/api/v1/health` - 1000 req/min (public endpoint)

**Expected Impact**:
- Prevents thundering herd
- Protects backend from metric floods
- **Enables stable operation at 500+ collectors**

**Testing**:
- Unit test: Token bucket acquire/refill
- Integration test: Exceed limit, verify 429 response
- Load test: 100 collectors all pushing simultaneously
- Verify Retry-After header sent

**Acceptance Criteria**:
- ✅ Token bucket works correctly
- ✅ 100 requests/min limit enforced per collector
- ✅ Bursts allowed up to window size
- ✅ Metrics available about rate limit hits
- ✅ Graceful degradation when limit hit

---

### Phase 2 Summary

**Total Effort**: 22-30 hours
**Focus**: Enable safe operation at scale
**Impact**:

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Serialization CPU | 75-150ms | 15-30ms | 80% ↓ |
| Data loss visibility | None | Tracked | ✓ |
| Thundering herd risk | High | Low | ✓ |
| Safe collector count | 100 | 200 | 2x ↑ |

**Outcome**: System stable for 100-200 collectors with full visibility

---

## Phase 3: MEDIUM Optimizations (28-38 hours)

### Task 3.1: Binary Protocol Implementation (16-20 hours)
- Switch from JSON+gzip to binary+zstd
- 60% bandwidth reduction
- Compatible with existing collectors

### Task 3.2: Connection Pool Monitoring (4-6 hours)
- Visualize pool utilization in Grafana
- Alert on stale connections
- Monitor connection age

### Task 3.3: Metrics Prioritization (8-12 hours)
- Mark metrics as critical/important/optional
- Preserve critical on overflow
- Discard optional under pressure

**Phase 3 Total**: 28-38 hours
**Outcome**: Enterprise-ready for 500+ collectors with horizontal scaling

---

## Implementation Schedule

### Week 1-2 (40 hours)
- **Phase 1**: All 3 CRITICAL tasks
  - Thread pool (20-30h)
  - Query limit config (2-4h)
  - Connection pooling (8-12h)
- **Start Phase 2**: Task 2.1 (JSON serialization)

### Week 3 (40 hours)
- **Phase 2**: Complete remaining tasks
  - Finish JSON serialization (remaining 4-12h)
  - Buffer overflow monitoring (4-6h)
  - Rate limiting (6-8h)
- **Load testing**: Validate improvements

### Week 4+ (40+ hours)
- **Phase 3**: Protocol optimization and monitoring
- **Horizontal scaling**: Multi-instance deployment

---

## Success Metrics

### Performance

| Metric | Current | Target | By When |
|--------|---------|--------|---------|
| CPU @ 100 collectors | 96% | 36% | Week 2 |
| Cycle time @ 100 col | 57.7s | 14.4s | Week 2 |
| Query sampling @ 10K QPS | 1% | 5-10% | Week 1 |
| Connection overhead | 200-400ms | 5-10ms | Week 1 |
| Serialization CPU | 75-150ms | 15-30ms | Week 3 |

### Scalability

- ✅ 10-25 collectors (current) → STABLE
- ✅ 25-100 collectors (after Phase 1) → STABLE
- ✅ 100-200 collectors (after Phase 2) → STABLE
- ✅ 200-500 collectors (after Phase 3) → STABLE

### Quality

- ✅ Zero data loss (overflow monitoring)
- ✅ 95% test coverage
- ✅ All existing tests passing
- ✅ Load test scenarios successful

---

## Risk Assessment

### High Risk Areas

1. **Thread Pool Complexity**
   - Mitigation: Start with small pool (2 threads), test extensively
   - Fallback: Keep single-threaded path as option

2. **Connection Pool Stale Connections**
   - Mitigation: Add health checks, timeout old connections
   - Fallback: Disable pooling if issues detected

3. **Binary Format Compatibility**
   - Mitigation: Version the format, support both temporarily
   - Fallback: Revert to JSON if binary causes issues

### Medium Risk Areas

4. **Rate Limiting False Positives**
   - Mitigation: Start with generous limits, reduce gradually
   - Monitoring: Track rate limit hits, adjust if needed

5. **Buffer Overflow Handling**
   - Mitigation: Test with synthetic overflow scenarios
   - Fallback: Increase buffer to 100MB temporary

---

## Sign-Off & Approval

- [ ] Project Lead Approval
- [ ] Architecture Review
- [ ] Security Review
- [ ] Performance Team Review

---

**Roadmap Created**: February 26, 2026
**Ready for**: Implementation Planning & Team Assignment
**Next Step**: Begin Phase 1 implementation (Week of 3/3/2026)

