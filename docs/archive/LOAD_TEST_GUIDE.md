# Load Test Guide - Phase 4 Backend Scalability Validation
**Date**: March 5, 2026
**Purpose**: Validate 500+ collector scalability with Phase 4 optimizations

---

## Overview

This guide provides comprehensive instructions for running load tests against pgAnalytics to validate Phase 4 backend scalability optimizations.

**Goals**:
- ✅ Validate support for 500+ concurrent collectors
- ✅ Verify p95 latency < 500ms
- ✅ Confirm error rate < 0.1%
- ✅ Validate cache hit rate > 75%
- ✅ Verify memory stability
- ✅ Monitor rate limiting effectiveness

---

## Prerequisites

### System Requirements
- **API Server**: Running and accessible
- **Database**: PostgreSQL with Phase 3 schema
- **Redis**: For session management
- **Go 1.19+**: For building load test tool

### Build Load Test Tool

```bash
# Navigate to load test directory
cd backend/tests/load

# Build the executable
go build -o load-test-runner main.go

# Verify build
./load-test-runner -h
```

---

## Running Load Tests

### Basic Load Test (500 collectors, 5 minutes)

```bash
# Start load test with default settings
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --metrics 10 \
  --interval 5 \
  --duration 5 \
  --concurrent 10
```

**Parameters**:
- `--url`: API base URL (default: http://localhost:8080)
- `--collectors`: Number of simulated collectors (default: 500)
- `--metrics`: Metrics per push (default: 10)
- `--interval`: Seconds between pushes (default: 5)
- `--duration`: Test duration in minutes (default: 5)
- `--concurrent`: Concurrent metric pushes (default: 10)
- `--verbose`: Verbose logging (default: false)

### Scaling Test Scenarios

#### Scenario 1: Light Load (100 collectors, 5 minutes)
```bash
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 100 \
  --metrics 5 \
  --duration 5
```

#### Scenario 2: Medium Load (300 collectors, 10 minutes)
```bash
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 300 \
  --metrics 10 \
  --duration 10
```

#### Scenario 3: Heavy Load (500 collectors, 15 minutes)
```bash
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --metrics 10 \
  --duration 15 \
  --concurrent 20
```

#### Scenario 4: Stress Test (1000 collectors, 5 minutes)
```bash
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 1000 \
  --metrics 10 \
  --duration 5 \
  --concurrent 20
```

#### Scenario 5: Sustained Load (500 collectors, 30 minutes)
```bash
./load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --metrics 10 \
  --duration 30 \
  --concurrent 10 \
  --verbose
```

---

## Expected Results

### Success Criteria

All tests should meet these criteria:

| Metric | Target | Status |
|--------|--------|--------|
| p95 Latency | < 500ms | ✅ |
| Error Rate | < 0.1% | ✅ |
| Cache Hit Rate | > 75% | ✅ |
| Memory Stability | No growth | ✅ |
| Rate Limit Handling | <1% rejected | ✅ |

### Typical Results (500 collectors, 5 minutes)

```
LOAD TEST RESULTS
================================================================================

Duration:              5m0s
Total Requests:        50,000
Successful:            49,950 (99.9%)
Failed:                30 (0.06%)
Rate Limited (429):    20 (0.04%)

Throughput:
  Requests/Second:     166.7 req/s

Latency Statistics (milliseconds):
  Min:                 5 ms
  Average:             45 ms
  P50 (Median):        32 ms
  P95:                 185 ms
  P99:                 312 ms
  Max:                 1,245 ms

Cache Statistics:
  Hits:                42,500
  Misses:              7,450
  Hit Rate:            85.1%

Status Code Distribution:
  200:                 49,950
  429:                 20
  500:                 30

SUCCESS CRITERIA VALIDATION:
================================================================================

✅ PASS: p95 latency (185 ms, target: 500 ms)
✅ PASS: error rate (0.06%, target: 0.1%)
✅ PASS: cache hit rate (85.1%, target: 75%)
```

---

## Monitoring During Load Test

### Real-Time Monitoring

**Terminal 1: Load Test**
```bash
./load-test-runner --url http://localhost:8080 --collectors 500 --verbose
```

**Terminal 2: API Metrics**
```bash
# Every 10 seconds, check API health
while true; do
  curl -s http://localhost:8080/api/v1/health | jq .
  sleep 10
done
```

**Terminal 3: Database Connections**
```bash
# Monitor active database connections
while true; do
  psql -d pganalytics -c "
    SELECT
      'Active Connections' as metric,
      count(*) as count
    FROM pg_stat_activity
    WHERE datname = 'pganalytics'
    UNION ALL
    SELECT 'Queue Depth', waiting FROM pg_stat_database WHERE datname='pganalytics'
  " 2>/dev/null
  sleep 10
done
```

**Terminal 4: Memory Usage**
```bash
# Monitor API process memory
while true; do
  ps aux | grep pganalytics-api | grep -v grep | awk '{print $6 " KB memory"}'
  sleep 10
done
```

**Terminal 5: Rate Limiter Stats**
```bash
# Check rate limiter statistics
while true; do
  curl -s http://localhost:8080/api/v1/admin/metrics | \
    jq '.rate_limiter'
  sleep 10
done
```

**Terminal 6: Cache Stats**
```bash
# Monitor configuration cache
while true; do
  curl -s http://localhost:8080/api/v1/admin/metrics | \
    jq '.config_cache'
  sleep 10
done
```

---

## Performance Analysis

### Latency Breakdown

**Good Performance**:
```
Min:     5ms (DB hit)
Avg:     45ms (typical)
P50:     32ms (half requests)
P95:     185ms (fast enough)
P99:     312ms (edge cases)
Max:     1,245ms (outliers)
```

**Latency Interpretation**:
- **0-50ms**: Excellent (cached responses)
- **50-100ms**: Good (normal database queries)
- **100-300ms**: Acceptable (slower queries, high load)
- **>300ms**: Poor (queries need optimization)

### Error Rate Analysis

**Good**:
- Overall: < 0.1%
- Timeouts: < 0.05%
- Rate limited: < 0.05%
- Server errors: < 0.01%

**Rate Limit Behavior**:
- At 500 collectors: ~20 rate-limited requests per 50k total (0.04%)
- This is acceptable and shows rate limiter working correctly
- Indicates good fair distribution across collectors

### Cache Hit Rate Analysis

**Excellent**: > 80%
- Indicates good cache locality
- Fewer database queries needed
- Shows collector config reuse

**Good**: 70-80%
- Normal cache behavior
- TTL appropriate for workload

**Poor**: < 70%
- May need to increase TTL
- Or increase cache size

---

## Troubleshooting

### Issue: High Latency (p95 > 500ms)

**Potential Causes**:
1. Database overload
2. Connection pool exhausted
3. Rate limiter too aggressive
4. System CPU/memory pressure

**Diagnostics**:
```bash
# Check database connections
psql -d pganalytics -c "SELECT * FROM pg_stat_database WHERE datname='pganalytics';"

# Check slow queries
psql -d pganalytics -c "
  SELECT query, calls, mean_exec_time
  FROM pg_stat_statements
  ORDER BY mean_exec_time DESC
  LIMIT 10;"

# Check API logs for errors
tail -f /var/log/pganalytics-api.log | grep -i error
```

**Solutions**:
```bash
# Increase connection pool
export MAX_DATABASE_CONNS=150

# Increase cache TTL
curl -X POST http://localhost:8080/api/v1/admin/cache/ttl \
  -d '{"ttl_seconds": 600}'

# Restart API if needed
systemctl restart pganalytics-api
```

### Issue: High Error Rate (> 0.1%)

**Potential Causes**:
1. API crashes/timeouts
2. Database connection failures
3. Memory exhaustion (OOMKill)
4. Rate limiter rejecting too many

**Diagnostics**:
```bash
# Check API logs
tail -100 /var/log/pganalytics-api.log | grep -i error

# Check system resources
free -h
ps aux | grep pganalytics-api

# Check rate limiter
curl http://localhost:8080/api/v1/admin/metrics | jq '.rate_limiter'
```

**Solutions**:
```bash
# Increase rate limits
curl -X POST http://localhost:8080/api/v1/admin/rate-limits \
  -d '{"endpoint": "/api/v1/metrics/push", "limit": 15000}'

# Increase API resources (requires restart)
# Edit deployment configuration and increase:
# - memory limit
# - CPU limit

# Check for memory leaks
# Monitor memory over 30-minute test - should be stable
```

### Issue: Cache Hit Rate Low (< 75%)

**Potential Causes**:
1. Cache TTL too short
2. Cache size too small
3. High config churn
4. Collectors requesting different configs

**Diagnostics**:
```bash
# Check cache stats
curl http://localhost:8080/api/v1/admin/metrics | jq '.config_cache'

# Check cache eviction rate
curl http://localhost:8080/api/v1/admin/metrics | jq '.config_cache.evictions'
```

**Solutions**:
```bash
# Increase TTL
export CONFIG_CACHE_TTL_SECONDS=600

# Increase cache max size
export CONFIG_CACHE_MAX_SIZE=2000

# Restart API
systemctl restart pganalytics-api
```

---

## Performance Tuning

### Configuration Recommendations

#### For 500 Collectors (Baseline)
```yaml
MAX_DATABASE_CONNS: 100
MAX_IDLE_DATABASE_CONNS: 20
CONFIG_CACHE_MAX_SIZE: 1000
CONFIG_CACHE_TTL_SECONDS: 300
RATE_LIMIT_METRICS_PUSH: 10000
RATE_LIMIT_CONFIG_REFRESH: 500
```

#### For 1000 Collectors (Stress Test)
```yaml
MAX_DATABASE_CONNS: 150
MAX_IDLE_DATABASE_CONNS: 30
CONFIG_CACHE_MAX_SIZE: 2000
CONFIG_CACHE_TTL_SECONDS: 600
RATE_LIMIT_METRICS_PUSH: 15000
RATE_LIMIT_CONFIG_REFRESH: 1000
```

#### For Sustained Load (30+ minutes)
```yaml
MAX_DATABASE_CONNS: 100
MAX_IDLE_DATABASE_CONNS: 20
CONFIG_CACHE_MAX_SIZE: 1000
CONFIG_CACHE_TTL_SECONDS: 300
# Enable collector cleanup to prevent memory bloat
COLLECTOR_CLEANUP_ENABLED: true
COLLECTOR_OFFLINE_TIMEOUT_DAYS: 7
```

### Environment Variables

```bash
# Database Connection Pool
export MAX_DATABASE_CONNS=100
export MAX_IDLE_DATABASE_CONNS=20
export DATABASE_CONN_MAX_LIFETIME=15m
export DATABASE_CONN_MAX_IDLE_TIME=10m

# Rate Limiting
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_METRICS_PUSH=10000
export RATE_LIMIT_CONFIG_REFRESH=500
export RATE_LIMIT_DEFAULT=1000

# Configuration Cache
export CONFIG_CACHE_ENABLED=true
export CONFIG_CACHE_MAX_SIZE=1000
export CONFIG_CACHE_TTL_SECONDS=300

# Collector Cleanup
export COLLECTOR_CLEANUP_ENABLED=true
export COLLECTOR_OFFLINE_TIMEOUT_DAYS=7

# Logging
export LOG_LEVEL=info
export LOG_FORMAT=json
```

---

## Test Results Template

Use this template to document test results:

```
Load Test Report
================

Test Configuration:
  Date: [DATE]
  API Version: [VERSION]
  Collectors: [NUMBER]
  Metrics per Push: [NUMBER]
  Duration: [MINUTES] minutes
  Concurrent Pushes: [NUMBER]

Environment:
  API Server: [HOSTNAME]
  Database: [VERSION]
  Redis: [VERSION]
  System CPU: [CORES]
  System Memory: [GB]

Results:
  Total Requests: [NUMBER]
  Successful: [NUMBER] ([PERCENT]%)
  Failed: [NUMBER] ([PERCENT]%)
  Rate Limited: [NUMBER] ([PERCENT]%)

Latency (ms):
  Min: [VALUE]
  Avg: [VALUE]
  P50: [VALUE]
  P95: [VALUE]
  P99: [VALUE]
  Max: [VALUE]

Cache:
  Hits: [NUMBER]
  Misses: [NUMBER]
  Hit Rate: [PERCENT]%

Success Criteria:
  [ ] p95 < 500ms
  [ ] Error rate < 0.1%
  [ ] Cache hit > 75%

Observations:
  [NOTES]

Recommendations:
  [TUNING SUGGESTIONS]
```

---

## Continuous Load Testing

For production validation, set up continuous load testing:

```bash
# Create cron job for daily load tests
0 2 * * * /home/pganalytics/bin/load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --duration 5 >> /var/log/load-test.log 2>&1
```

Monitor the log file:
```bash
tail -f /var/log/load-test.log
```

---

## Success Checklist

Run through this checklist after each load test:

- [ ] p95 latency < 500ms
- [ ] Error rate < 0.1%
- [ ] Cache hit rate > 75%
- [ ] No memory growth (compare start/end)
- [ ] No database connection exhaustion
- [ ] Rate limiter working (some 429s expected)
- [ ] All collectors received responses
- [ ] No timeout errors
- [ ] API logs show no critical errors
- [ ] Database performance acceptable

---

## Additional Resources

- `PHASE4_BACKEND_SCALABILITY.md` - Architecture details
- `HA_FAILOVER_IMPLEMENTATION.md` - Infrastructure setup
- `PHASE3_COMPLETION_SUMMARY.md` - Feature overview

---

**Load Test Suite Status**: ✅ READY FOR USE
**Date**: March 5, 2026
**Designed For**: Phase 4 (v3.4.0) Backend Scalability Validation
