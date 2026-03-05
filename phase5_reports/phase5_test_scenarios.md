# Phase 5 Extended Load Test Scenarios

## Scenario 1: Baseline Validation
- Collectors: 100
- Metrics/Push: 10
- Interval: 5 seconds
- Duration: 5 minutes
- Expected Throughput: ~50 requests/sec
- **Target Success Rate:** >99.9%
- **Target p95 Latency:** <185ms (Phase 4 baseline)

## Scenario 2: Medium Load (3x Scaling)
- Collectors: 300
- Metrics/Push: 10
- Interval: 5 seconds
- Duration: 10 minutes
- Expected Throughput: ~150 requests/sec
- **Target Success Rate:** >99.9%
- **Target p95 Latency:** <250ms

## Scenario 3: Full-Scale Load
- Collectors: 500
- Metrics/Push: 10
- Interval: 5 seconds
- Duration: 30 minutes
- Expected Throughput: ~250 requests/sec
- **Target Success Rate:** >99.9%
- **Target p95 Latency:** <350ms

## Scenario 4: Sustained Load (Memory Leak Detection)
- Collectors: 500
- Metrics/Push: 10
- Interval: 5 seconds
- Duration: 60 minutes
- Expected Throughput: ~250 requests/sec
- **Target Success Rate:** >99.9%
- **Target p95 Latency:** stable <350ms
- **Target Memory Growth:** <0.2%/minute

## Feature Validation During Tests
1. Anomaly Detection: Ensure baselines computed correctly
2. Alert Rules: Verify rules evaluate without errors
3. Notifications: Check notification queueing
4. Cache Performance: Monitor hit rates (target >75%)
5. Circuit Breaker: Verify resilience under load
