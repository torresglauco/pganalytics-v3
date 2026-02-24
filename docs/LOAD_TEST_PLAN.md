# Load Test Plan - 100+ Simulated Collectors

**Date**: February 22, 2026
**Status**: ✅ READY FOR EXECUTION
**Target**: Validate performance with 100,000+ concurrent collectors

---

## Load Test Overview

This document outlines the comprehensive load test strategy for validating the pganalytics-v3 collector with binary protocol support under realistic production-like conditions.

### Test Objectives

1. **Scalability Validation**
   - Test with 10, 50, 100, 500 collectors
   - Identify breaking points
   - Validate 100,000+ capacity claim

2. **Protocol Comparison**
   - JSON protocol performance baseline
   - Binary protocol improvements
   - Bandwidth reduction verification
   - CPU/Memory impact analysis

3. **Performance Benchmarking**
   - Metrics collection latency
   - Backend ingestion latency
   - Network bandwidth usage
   - Resource utilization (CPU, memory, disk)

4. **Stability & Reliability**
   - Long-running stability (24+ hours)
   - Error rate validation (<0.1%)
   - Recovery from failures
   - Data integrity verification

---

## Load Test Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Test Framework                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Load Test Generator (Python/Go)                           │
│  ├─ Spawn N collector processes                            │
│  ├─ Configure each with unique ID                          │
│  ├─ Control collection interval                            │
│  └─ Monitor resource usage                                 │
│                                                             │
│  ┌──────────────────────────────────────────────┐           │
│  │ 10 Collectors  │ 50 Collectors │ 100+ Collectors        │
│  ├────────────────┼───────────────┼─────────────────       │
│  │ Per-collector: │ Per-collector:│ Per-collector:         │
│  │ • 60s interval │ • 60s interval│ • 60s interval         │
│  │ • 50 metrics   │ • 50 metrics  │ • 50 metrics           │
│  │ • 2s latency   │ • 2s latency  │ • 2s latency           │
│  └──────────────────────────────────────────────┘           │
│           ↓              ↓              ↓                   │
│  ┌─────────────────────────────────────────────────┐        │
│  │         Test Backend & Database                │        │
│  ├─────────────────────────────────────────────────┤        │
│  │ PostgreSQL + TimescaleDB + Backend API         │        │
│  └─────────────────────────────────────────────────┘        │
│           ↓              ↓              ↓                   │
│  ┌─────────────────────────────────────────────────┐        │
│  │  Monitoring & Metrics Collection               │        │
│  ├─────────────────────────────────────────────────┤        │
│  │ • Prometheus metrics                           │        │
│  │ • Grafana dashboards                           │        │
│  │ • Performance reports                          │        │
│  └─────────────────────────────────────────────────┘        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Test Scenarios

### Scenario 1: Baseline JSON Protocol (10 collectors)

**Duration**: 15 minutes
**Collectors**: 10
**Collection Interval**: 60 seconds
**Metrics per Collector**: 50 per collection
**Total Metrics/Minute**: 500 (10 × 50)
**Total Metrics/Hour**: 30,000

**Metrics to Track**:
- Average response time per collection
- Metrics ingestion rate (metrics/second)
- Backend CPU/memory usage
- Database CPU/memory usage
- Network bandwidth
- Error rate

**Expected Results**:
- Response time: <500ms per collection
- Ingestion rate: >60 metrics/second
- Errors: <0.1%

### Scenario 2: Binary Protocol Optimized (10 collectors)

**Duration**: 15 minutes
**Collectors**: 10 (same as Scenario 1)
**Protocol**: Binary + Zstd compression
**Collection Interval**: 60 seconds
**Metrics per Collector**: 50

**Metrics to Track**:
- Comparison with Scenario 1
- Bandwidth reduction percentage
- CPU reduction percentage
- Memory reduction percentage
- Serialization time

**Expected Results**:
- 20-30% faster (binary protocol)
- 60% less bandwidth (Zstd compression)
- 10-30% less CPU usage

### Scenario 3: Scale Test - 50 Collectors

**Duration**: 30 minutes
**Collectors**: 50
**Collection Interval**: 60 seconds
**Metrics per Collector**: 50 per collection
**Total Metrics/Minute**: 2,500 (50 × 50)
**Total Metrics/Hour**: 150,000

**Performance Targets**:
- Response time: <1000ms
- Ingestion rate: >300 metrics/second
- Errors: <0.1%
- Memory increase: Linear with collector count

### Scenario 4: Heavy Load - 100 Collectors

**Duration**: 60 minutes
**Collectors**: 100
**Collection Interval**: 60 seconds
**Metrics per Collector**: 50 per collection
**Total Metrics/Minute**: 5,000 (100 × 50)
**Total Metrics/Hour**: 300,000

**Performance Targets**:
- Response time: <2000ms
- Ingestion rate: >500 metrics/second
- Errors: <0.2%
- CPU scaling: Should be sub-linear

### Scenario 5: Maximum Scale - 500+ Collectors

**Duration**: 120 minutes (2 hours)
**Collectors**: 500
**Collection Interval**: 120 seconds (doubled for scale)
**Metrics per Collector**: 50 per collection
**Total Metrics/Minute**: 250 (500 × 50 / 120s)
**Total Metrics/Hour**: 15,000

**Performance Targets**:
- Response time: <5000ms
- Ingestion rate: >100 metrics/second
- Errors: <0.5%
- Identify bottlenecks at scale

### Scenario 6: Protocol Comparison at Scale (100 collectors)

**Duration**: 60 minutes
**Collectors**: 100 (50 JSON, 50 Binary)
**Metrics**: 300,000/hour total

**Comparison Points**:
- Bandwidth per protocol
- CPU usage per protocol
- Memory usage per protocol
- Error rates per protocol
- Latency per protocol

---

## Load Test Implementation

### Test Framework (Go/Python)

```go
// Simulated collector process
type SimulatedCollector struct {
    ID              string
    BackendURL      string
    Protocol        string  // "json" or "binary"
    Interval        int     // seconds
    MetricsCount    int

    // Metrics
    CollectionTime  time.Duration
    IngestionTime   time.Duration
    ErrorCount      int
    SuccessCount    int
}

func (c *SimulatedCollector) Collect() error {
    // Generate synthetic metrics
    metrics := generateMetrics(c.MetricsCount)

    // Measure collection time
    start := time.Now()

    // Send via specified protocol
    if c.Protocol == "binary" {
        err := c.sendBinary(metrics)
        if err != nil {
            c.ErrorCount++
            return err
        }
    } else {
        err := c.sendJSON(metrics)
        if err != nil {
            c.ErrorCount++
            return err
        }
    }

    c.CollectionTime = time.Since(start)
    c.SuccessCount++
    return nil
}

func (c *SimulatedCollector) Run(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.Collect()
        }
    }
}
```

---

## Performance Metrics to Collect

### Per Collector

```
collection_time_ms           Time to collect metrics from simulated DB
ingestion_latency_ms         Time for backend to process submission
roundtrip_time_ms            Total time from collection to ingestion
error_count                  Number of failed transmissions
success_count                Number of successful transmissions
success_rate                 Percentage of successful submissions
bytes_sent                   Network bytes sent (protocol dependent)
cpu_time_ms                  CPU time spent on this collector
```

### Backend Metrics

```
metrics_received_total       Total metrics ingested
metrics_received_rate        Metrics per second
avg_ingestion_latency_ms     Average time to process metrics
p50_latency_ms               50th percentile latency
p95_latency_ms               95th percentile latency
p99_latency_ms               99th percentile latency
error_rate                   Percentage of failed submissions
protocol_distribution        Count by protocol (json/binary)
```

### System Metrics

```
backend_cpu_percent          CPU usage of backend process
backend_memory_mb            Memory usage of backend
database_cpu_percent         PostgreSQL CPU usage
database_memory_mb           PostgreSQL memory usage
timescaledb_cpu_percent      TimescaleDB CPU usage
disk_io_mb_s                 Disk I/O throughput
network_bandwidth_mbps       Network bandwidth in use
connection_count             Database connection count
```

---

## Test Execution Steps

### Phase 1: Environment Preparation

1. **Start Backend Stack**
   ```bash
   docker-compose up -d postgres timescale backend
   ```

2. **Verify Services**
   ```bash
   docker-compose exec backend curl http://localhost:8080/api/v1/health
   docker-compose exec postgres psql -U postgres -c "SELECT 1"
   ```

3. **Set Up Monitoring**
   - Enable Prometheus metrics collection
   - Configure Grafana dashboards
   - Prepare log aggregation

### Phase 2: Baseline Testing (Scenario 1)

1. Deploy 10 JSON protocol collectors
2. Run for 15 minutes
3. Collect baseline metrics
4. Analyze results

### Phase 3: Protocol Comparison (Scenario 2)

1. Stop Scenario 1 collectors
2. Deploy 10 Binary protocol collectors
3. Run for 15 minutes
4. Compare with Scenario 1 results

### Phase 4: Scale Testing (Scenarios 3-5)

1. Deploy 50 collectors → Run 30 min
2. Deploy 100 collectors → Run 60 min
3. Deploy 500 collectors → Run 120 min
4. Identify performance degradation points

### Phase 5: Mixed Protocol Testing (Scenario 6)

1. Deploy 50 JSON + 50 Binary collectors
2. Run for 60 minutes
3. Compare protocol performance at scale

### Phase 6: Long-Running Stability (Optional)

1. Deploy 50-100 collectors
2. Run for 24+ hours
3. Monitor for:
   - Memory leaks
   - Connection exhaustion
   - Performance degradation over time
   - Error accumulation

---

## Success Criteria

### Phase 1: 10 Collectors (JSON)
- [ ] Response time <500ms
- [ ] Error rate <0.1%
- [ ] Ingestion rate >60 metrics/sec
- [ ] No memory leaks

### Phase 2: 10 Collectors (Binary)
- [ ] 20-30% faster than JSON
- [ ] 60% bandwidth reduction
- [ ] Same or better error rate
- [ ] Lower CPU usage

### Phase 3: 50 Collectors
- [ ] Response time <1000ms
- [ ] Error rate <0.1%
- [ ] Ingestion rate >300 metrics/sec
- [ ] Linear memory scaling
- [ ] Throughput scales with collector count

### Phase 4: 100 Collectors
- [ ] Response time <2000ms
- [ ] Error rate <0.2%
- [ ] Ingestion rate >500 metrics/sec
- [ ] No connection exhaustion
- [ ] Backend handles concurrent load

### Phase 5: 500+ Collectors
- [ ] Response time <5000ms
- [ ] Error rate <0.5%
- [ ] Maintains stability
- [ ] Identifies scaling bottlenecks
- [ ] Database can handle volume

### Phase 6: Protocol Comparison
- [ ] Binary >20% improvement over JSON
- [ ] Bandwidth reduction validates 60% target
- [ ] Both protocols scale linearly

---

## Load Test Monitoring

### Real-Time Dashboard

Monitor during load test:

```
┌─────────────────────────────────────────────────────────────┐
│                  LOAD TEST DASHBOARD                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ COLLECTORS STATUS                                           │
│ Active: 100/100  │  Healthy: 100  │  Errors: 0            │
│                                                             │
│ THROUGHPUT                                                  │
│ Metrics/sec: 523  │  Mbps: 45.2  │  Errors: 0.08%        │
│                                                             │
│ LATENCY (ms)                                                │
│ Min: 12  │  Avg: 245  │  P95: 1200  │  Max: 3400          │
│                                                             │
│ RESOURCE USAGE                                              │
│ Backend CPU: 35%  │  Memory: 256MB                         │
│ DB CPU: 42%       │  Memory: 512MB                         │
│                                                             │
│ NETWORK                                                     │
│ JSON: 45.2 Mbps   │  Binary: 18.1 Mbps  │  Saved: 27.1 Mbps│
│                                                             │
│ DATABASE                                                    │
│ Connections: 45   │  Queries/sec: 1200  │  Write: 500/sec  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Grafana Dashboards

Create dashboards for:
1. **Collector Health** - Active collectors, error rates
2. **Performance** - Latency, throughput, bandwidth
3. **Resources** - CPU, memory, disk I/O
4. **Database** - Connections, queries, inserts
5. **Protocol Comparison** - JSON vs Binary metrics

---

## Load Test Tools

### Tool Options

1. **Custom Go Program** (Recommended)
   - Direct control over collector behavior
   - Precise timing and metrics
   - Easy to configure and extend
   - Minimal overhead

2. **Apache JMeter**
   - Familiar to QA teams
   - Built-in reporting
   - Scalability limitations
   - Higher overhead

3. **Locust (Python)**
   - Python-based, easy to modify
   - Good for quick tests
   - Less efficient than Go

### Recommended Approach

Use Go for load test generator with built-in metrics:

```bash
# Build load test tool
go build -o load-test ./tools/load-test/main.go

# Run test scenarios
./load-test -collectors 10 -duration 15m -protocol json
./load-test -collectors 10 -duration 15m -protocol binary
./load-test -collectors 50 -duration 30m -protocol binary
./load-test -collectors 100 -duration 60m -protocol binary
```

---

## Expected Results Summary

### Performance Improvements (Binary Protocol)

| Metric | JSON | Binary | Improvement |
|--------|------|--------|-------------|
| Bandwidth (KB/min) | 450 | 180 | 60% reduction |
| Serialization (ms) | 15 | 5 | 3x faster |
| CPU (%) | 35% | 25% | 10-30% reduction |
| Memory (MB) | 256 | 128 | 47% reduction |
| Latency (ms) | 250 | 180 | 30% faster |

### Scalability Results

| Collectors | Metrics/Hour | Response Time | Error Rate | Status |
|------------|--------------|---------------|-----------|--------|
| 10 | 30,000 | <500ms | <0.1% | ✅ |
| 50 | 150,000 | <1000ms | <0.1% | ✅ |
| 100 | 300,000 | <2000ms | <0.2% | ✅ |
| 500 | 1,500,000 | <5000ms | <0.5% | ✅ |

### System Resource Usage

At 100 collectors:
- Backend CPU: 30-40%
- Backend Memory: 256-512MB
- Database CPU: 40-50%
- Database Memory: 512-1024MB
- Network Bandwidth: 20-50 Mbps (varies by protocol)

---

## Test Report Template

After each test run, generate report:

```
═════════════════════════════════════════════════════════════════
                      LOAD TEST REPORT
                    [Date/Time Range]
═════════════════════════════════════════════════════════════════

TEST CONFIGURATION
  Collectors:        100
  Protocol:          Binary
  Duration:          60 minutes
  Collection Interval: 60 seconds
  Metrics/Collector:  50

RESULTS SUMMARY
  Total Metrics:     300,000
  Success Rate:      99.92%
  Errors:            240
  Avg Latency:       245ms
  P95 Latency:       1,200ms
  P99 Latency:       2,800ms

PERFORMANCE METRICS
  Ingestion Rate:    500 metrics/sec
  Throughput:        3,000 metrics/min
  Bandwidth:         18.1 Mbps (binary), 45.2 Mbps (JSON equivalent)
  Compression Ratio: 60% reduction

RESOURCE USAGE
  Backend CPU:       35% avg, 42% peak
  Backend Memory:    256MB avg, 312MB peak
  Database CPU:      42% avg, 48% peak
  Database Memory:   512MB avg, 614MB peak

ANALYSIS
  ✅ Binary protocol delivered 60% bandwidth reduction
  ✅ Latency remained under target at scale
  ✅ Error rate acceptable at 0.08%
  ✅ Resource scaling was sub-linear
  ✅ No memory leaks detected

RECOMMENDATIONS
  1. Deploy with binary protocol for production
  2. Monitor at 200+ collectors for further optimization
  3. Consider connection pooling tuning for 1000+ collectors
  4. Implement metric batching for extreme scale

═════════════════════════════════════════════════════════════════
```

---

## Next Steps

1. ✅ Prepare load test environment
2. ✅ Deploy load test tool
3. **→ Execute Scenario 1: 10 JSON collectors**
4. → Execute Scenario 2: 10 Binary collectors
5. → Execute Scenario 3: 50 collectors
6. → Execute Scenario 4: 100 collectors
7. → Execute Scenario 5: 500+ collectors
8. → Generate comprehensive report

---

## Timeline

| Phase | Duration | Start | End |
|-------|----------|-------|-----|
| Preparation | 30 min | Now | +30min |
| Scenario 1 (10 JSON) | 15 min | +30min | +45min |
| Scenario 2 (10 Binary) | 15 min | +45min | +60min |
| Scenario 3 (50) | 30 min | +60min | +90min |
| Scenario 4 (100) | 60 min | +90min | +150min |
| Scenario 5 (500) | 120 min | +150min | +270min |
| **Total** | **~4.5 hours** | | |

---

## Support & Documentation

For detailed implementation:
- DEPLOYMENT_GUIDE.md - Setup instructions
- BINARY_PROTOCOL_USAGE_GUIDE.md - Protocol details
- deploy.sh - Deployment automation

---

**Generated**: February 22, 2026
**Project**: pganalytics-v3 (torresglauco)
**Status**: ✅ LOAD TEST PLAN READY
