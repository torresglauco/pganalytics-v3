# Load Test Execution Guide

**Date**: February 22, 2026
**Status**: ✅ READY FOR EXECUTION
**Target**: Validate 100,000+ collector capacity with JSON vs Binary protocols

---

## Quick Start

### Fastest Way to Run Tests

```bash
cd /Users/glauco.torres/git/pganalytics-v3

# Quick test (10 collectors, 15 minutes total)
./run-load-tests.sh --quick

# Full test suite (10, 50, 100, 500 collectors)
./run-load-tests.sh --full

# Test single protocol
./run-load-tests.sh --full --binary
./run-load-tests.sh --full --json
```

---

## Pre-Test Checklist

### 1. Ensure Backend is Running

```bash
# Check if services are running
docker-compose ps

# Expected output:
# pganalytics-postgres   running
# pganalytics-timescale  running
# pganalytics-backend    running
# pganalytics-collector-demo running
# pganalytics-grafana    running

# If not running, start them:
docker-compose up -d
```

### 2. Verify Backend Health

```bash
# Test backend API
curl -k https://localhost:8080/api/v1/health

# Expected response:
# {"status":"healthy","version":"3.0.0"}
```

### 3. Verify Database Connectivity

```bash
# Test PostgreSQL
docker-compose exec postgres psql -U postgres -c "SELECT 1"

# Test TimescaleDB
docker-compose exec timescale psql -U postgres -d metrics -c "SELECT 1"
```

### 4. Clear Previous Test Data (Optional)

```bash
# Backup current metrics
docker-compose exec postgres \
  pg_dump -U postgres metrics > metrics_backup_$(date +%s).sql

# Clear metrics table (warning: deletes data)
docker-compose exec postgres \
  psql -U postgres -d metrics -c "TRUNCATE TABLE metrics"
```

---

## Load Test Execution

### Option 1: Quick Test (Recommended First Run)

**Duration**: ~30 minutes
**Collectors**: 10
**Scenarios**: JSON protocol, then Binary protocol

```bash
./run-load-tests.sh --quick

# Output:
# [INFO] Starting quick load test (10 collectors only)
# [✓] Load test script found
# [✓] Backend is reachable
#
# [INFO] Testing json protocol...
# [INFO] Running test: 10 collectors, json protocol
#
# [After 15 minutes]
#
# ════════════════════════════════════════════════════════════════════════════
#                            LOAD TEST REPORT
# ════════════════════════════════════════════════════════════════════════════
#
# TEST CONFIGURATION
#   Collectors:          10
#   Protocol:            json
#   Duration:            900s
#   Collection Interval: 60s
#   Metrics/Collector:   50
#
# RESULTS SUMMARY
#   Total Collections:   150
#   Successful:          150 (100.00%)
#   Errors:              0 (0.00%)
#   Actual Duration:     901.23s
#
# PERFORMANCE METRICS
#   Avg Latency:         234.56ms
#   Min Latency:         45.23ms
#   Max Latency:         1234.56ms
#   P95 Latency:         789.12ms
#   Throughput:          8.33 metrics/sec
#   Total Metrics:       7500 (30000/hour)
#
# BANDWIDTH ANALYSIS
#   Bytes Sent:          3,750,000 bytes
#
# ════════════════════════════════════════════════════════════════════════════
```

### Option 2: Full Test Suite

**Duration**: ~5-6 hours (depending on system)
**Collectors**: 10 → 50 → 100 → 500
**Scenarios**: JSON and Binary protocols

```bash
./run-load-tests.sh --full

# Estimated timeline:
# 00:00 - Start
# 00:30 - 10 collectors (JSON + Binary)
# 01:00 - 50 collectors (JSON + Binary)
# 02:00 - 100 collectors (JSON + Binary)
# 04:00 - 500 collectors (JSON + Binary)
# 05:00 - Complete
```

### Option 3: Protocol Comparison Only

Test only specific protocol:

```bash
# Test only binary protocol (faster)
./run-load-tests.sh --full --binary

# Test only JSON protocol
./run-load-tests.sh --full --json
```

### Option 4: Custom Backend URL

If backend is running on different host:

```bash
./run-load-tests.sh --full --backend https://api.example.com:8080
```

---

## Monitoring During Tests

### Real-Time Monitoring

While tests are running, monitor system resources:

```bash
# Monitor Docker container resources
docker stats pganalytics-backend pganalytics-postgres

# Expected output:
# CONTAINER                CPU %   MEM USAGE / LIMIT
# pganalytics-backend      35%     256MB / 1GB
# pganalytics-postgres     42%     512MB / 2GB

# Monitor processes
top -p $(docker inspect -f '{{.State.Pid}}' pganalytics-backend)
```

### Monitor Collector Activity

```bash
# View collector logs
docker-compose logs -f pganalytics-collector-demo

# Check metrics being received
docker-compose exec postgres \
  psql -U postgres -d metrics -c \
  "SELECT COUNT(*) FROM metrics WHERE timestamp > NOW() - interval '5 minutes'"
```

### Monitor Backend Performance

```bash
# View backend logs
docker-compose logs -f pganalytics-backend

# Check backend health
curl -s -k https://localhost:8080/api/v1/metrics | jq .
```

---

## Load Test Scenarios Explained

### Scenario 1: 10 Collectors Baseline

**Purpose**: Establish baseline performance
**Duration**: 15 minutes
**Load**:
- 10 collectors
- 50 metrics each
- 60-second intervals
- Total: 500 metrics/minute = 30,000/hour

**Success Criteria**:
- [ ] All collections successful (100%)
- [ ] Average latency <500ms
- [ ] Max latency <2000ms
- [ ] No errors

### Scenario 2: 50 Collectors

**Purpose**: Validate linear scaling
**Duration**: 30 minutes
**Load**:
- 50 collectors
- 50 metrics each
- 60-second intervals
- Total: 2,500 metrics/minute = 150,000/hour

**Success Criteria**:
- [ ] Success rate >99.9%
- [ ] Average latency <1000ms
- [ ] Throughput scales linearly
- [ ] Resource usage scales proportionally

### Scenario 3: 100 Collectors

**Purpose**: Test production-like load
**Duration**: 60 minutes
**Load**:
- 100 collectors
- 50 metrics each
- 60-second intervals
- Total: 5,000 metrics/minute = 300,000/hour

**Success Criteria**:
- [ ] Success rate >99.8%
- [ ] Average latency <2000ms
- [ ] Throughput >500 metrics/sec
- [ ] No connection exhaustion
- [ ] Stable performance over 1 hour

### Scenario 4: 500 Collectors

**Purpose**: Test extreme scale
**Duration**: 120 minutes (2 hours)
**Load**:
- 500 collectors
- 50 metrics each
- 120-second intervals (doubled for scale)
- Total: 250 metrics/minute = 15,000/hour

**Success Criteria**:
- [ ] Success rate >99.5%
- [ ] Identifiable bottleneck(s)
- [ ] Database still responsive
- [ ] No crashes or OOM

---

## Interpreting Results

### Latency Analysis

```
Avg Latency: 245ms
  - Good for 10-50 collectors
  - Acceptable for 100 collectors
  - May need optimization for 500+

P95 Latency: 1,200ms
  - Should scale with collector count
  - Indicates queue depth at peak

Max Latency: 3,400ms
  - Outliers acceptable if rare
  - Investigate if frequent
```

### Throughput

```
Metrics/sec: 523
Expected for 100 collectors:
- 100 collectors × 50 metrics = 5,000 metrics per collection cycle
- 60-second interval = 5,000/60 = 83 metrics/sec baseline
- With batching and optimization: 500+ metrics/sec typical

Scaling:
- 10 collectors: ~83 metrics/sec
- 50 collectors: ~415 metrics/sec
- 100 collectors: ~830 metrics/sec (if linear)
- Actual may be lower due to serialization overhead
```

### Bandwidth Comparison

```
JSON Protocol:
  45.2 Mbps typical
  = 3,414 MB/hour
  = 81.9 GB/day

Binary Protocol:
  18.1 Mbps typical
  = 1,365 MB/hour
  = 32.8 GB/day

Savings:
  60% bandwidth reduction
  = 49.1 GB/day saved per 100 collectors
```

### Error Analysis

```
< 0.1% Error Rate:
  ✅ Excellent - production ready

0.1% - 0.5% Error Rate:
  ⚠️  Acceptable - monitor for patterns

> 0.5% Error Rate:
  ❌ Investigate - may indicate issues with:
     - Backend overload
     - Network issues
     - Connection pool exhaustion
     - Database lock contention
```

---

## Troubleshooting

### Issue: "Connection refused"

**Cause**: Backend not running or unreachable
**Solution**:
```bash
# Start services
docker-compose up -d

# Wait for health
docker-compose exec backend curl http://localhost:8080/api/v1/health
```

### Issue: High Error Rate

**Cause**: Backend or database overwhelmed
**Solution**:
```bash
# Check backend logs
docker-compose logs pganalytics-backend | grep -i error

# Check database connections
docker-compose exec postgres \
  psql -U postgres -c "SELECT count(*) FROM pg_stat_activity"

# If too many connections, increase pool:
# In docker-compose.yml, increase connection limits
```

### Issue: Memory Growing Unbounded

**Cause**: Potential memory leak
**Solution**:
```bash
# Monitor memory during test
docker stats pganalytics-backend --no-stream

# If growing, may indicate:
# - Connection leak
# - Unbounded queue
# - Metric buffering issue

# Restart and retry
docker-compose restart pganalytics-backend
```

### Issue: Latency Increasing Over Time

**Cause**: Metrics table growing large, queries slowing
**Solution**:
```bash
# Check table size
docker-compose exec postgres \
  psql -U postgres -d metrics -c \
  "SELECT pg_size_pretty(pg_total_relation_size('metrics'))"

# Create index if missing
docker-compose exec postgres \
  psql -U postgres -d metrics -c \
  "CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)"

# Vacuum table
docker-compose exec postgres \
  psql -U postgres -d metrics -c "VACUUM metrics"
```

---

## Results Analysis

### After Each Test Run

1. **Check Summary**
   ```bash
   cat load-test-results/summary.txt
   ```

2. **Review Logs**
   ```bash
   ls -lh load-test-results/
   tail -100 load-test-results/test_*.log
   ```

3. **Compare Results**
   ```bash
   # Extract key metrics from logs
   grep "Throughput" load-test-results/test_*json*.log
   grep "Throughput" load-test-results/test_*binary*.log
   ```

### Generating Comparison Report

```bash
#!/bin/bash
# Compare JSON vs Binary results

echo "PROTOCOL COMPARISON"
echo "==================="
echo ""

for collectors in 10 50 100 500; do
    echo "Collectors: $collectors"

    json_log=$(ls load-test-results/test_${collectors}c_json*.log | tail -1)
    binary_log=$(ls load-test-results/test_${collectors}c_binary*.log | tail -1)

    echo "  JSON:"
    grep -E "Throughput|Bytes Sent|Latency" "$json_log" | head -5

    echo "  Binary:"
    grep -E "Throughput|Bytes Sent|Latency" "$binary_log" | head -5

    echo ""
done
```

---

## Performance Expectations

### Baseline Performance (10 Collectors)

| Metric | JSON | Binary | Improvement |
|--------|------|--------|-------------|
| Avg Latency | 250ms | 180ms | 28% faster |
| Throughput | 83 metrics/sec | 110 metrics/sec | 32% faster |
| Bandwidth | 450 KB/min | 180 KB/min | 60% reduction |
| Success Rate | >99.9% | >99.9% | Same |

### Scaling Performance (100 Collectors)

| Metric | Target | Expected | Status |
|--------|--------|----------|--------|
| Response Time | <2000ms | <2000ms | ✅ |
| Success Rate | >99.8% | >99.8% | ✅ |
| Throughput | >500 metrics/sec | 500+ metrics/sec | ✅ |
| Error Rate | <0.2% | <0.2% | ✅ |

---

## Next Steps

1. **Run Quick Test**
   ```bash
   ./run-load-tests.sh --quick
   ```
   Expected time: 30-40 minutes

2. **Run Full Test Suite**
   ```bash
   ./run-load-tests.sh --full
   ```
   Expected time: 5-6 hours

3. **Analyze Results**
   - Compare JSON vs Binary protocol
   - Identify performance bottlenecks
   - Validate 100,000+ collector capacity

4. **Performance Tuning** (if needed)
   - Increase connection pool size
   - Optimize batch sizes
   - Tune database indexes
   - Adjust collection intervals

5. **Production Readiness**
   - Document performance baseline
   - Set up monitoring/alerts
   - Plan deployment strategy
   - Conduct security review

---

## Support

For issues or questions:
- See LOAD_TEST_PLAN.md for detailed test plan
- Check DEPLOYMENT_GUIDE.md for backend setup
- Review BINARY_PROTOCOL_USAGE_GUIDE.md for protocol details

---

**Generated**: February 22, 2026
**Project**: pganalytics-v3 (torresglauco)
**Status**: ✅ LOAD TEST READY TO EXECUTE
