# Phase 1 Deployment Guide - pgAnalytics v3.3.0

**Date**: February 26, 2026
**Version**: 3.3.0 with Phase 1 Critical Fixes
**Status**: ✅ READY FOR PRODUCTION DEPLOYMENT

---

## Table of Contents

1. [Pre-Deployment Checklist](#pre-deployment-checklist)
2. [Deployment Steps](#deployment-steps)
3. [Configuration](#configuration)
4. [Validation](#validation)
5. [Rollback Plan](#rollback-plan)
6. [Performance Monitoring](#performance-monitoring)
7. [Troubleshooting](#troubleshooting)

---

## Pre-Deployment Checklist

### Environment Requirements

- **PostgreSQL**: 13+ (connection pooling requires modern libpq)
- **Operating System**: Linux (tested), macOS (tested), Windows (requires WSL2)
- **C++ Compiler**: C++17 support (GCC 7+, Clang 5+, MSVC 2017+)
- **CMake**: 3.22+
- **Memory**: Minimum 512MB per collector instance
- **CPU**: Minimum 2 cores recommended

### Pre-Flight Checks

```bash
# 1. Verify compiler version
g++ --version  # or clang++ --version
# Expected: C++17 support

# 2. Verify CMake version
cmake --version
# Expected: 3.22+

# 3. Verify PostgreSQL connectivity
psql -h <host> -U <user> -d <database> -c "SELECT version();"
# Expected: PostgreSQL 13+

# 4. Verify disk space
df -h /opt/pganalytics
# Expected: At least 2GB free

# 5. Check existing pgAnalytics version (if upgrading)
/opt/pganalytics/bin/pganalytics --version
# Expected: Version string
```

---

## Deployment Steps

### Step 1: Backup Current Installation

```bash
#!/bin/bash
BACKUP_DIR="/backup/pganalytics-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup current binary
if [ -f /opt/pganalytics/bin/pganalytics ]; then
    cp /opt/pganalytics/bin/pganalytics "$BACKUP_DIR/"
fi

# Backup configuration
if [ -f /etc/pganalytics/collector.toml ]; then
    cp /etc/pganalytics/collector.toml "$BACKUP_DIR/"
fi

# Backup logs
if [ -d /var/log/pganalytics ]; then
    cp -r /var/log/pganalytics "$BACKUP_DIR/logs"
fi

echo "Backup created: $BACKUP_DIR"
```

### Step 2: Download and Extract Release

```bash
#!/bin/bash
RELEASE_VERSION="v3.3.0"
RELEASE_URL="https://github.com/torresglauco/pganalytics-v3/releases/download/${RELEASE_VERSION}/pganalytics-v3.3.0-linux-x64.tar.gz"

# Download
cd /tmp
wget "$RELEASE_URL" -O pganalytics-v3.3.0.tar.gz

# Verify checksum
echo "Verifying checksum..."
sha256sum pganalytics-v3.3.0.tar.gz

# Extract
mkdir -p pganalytics-v3.3.0
tar xzf pganalytics-v3.3.0.tar.gz -C pganalytics-v3.3.0/

echo "Release extracted to /tmp/pganalytics-v3.3.0"
```

### Step 3: Build from Source (Alternative)

```bash
#!/bin/bash
# Clone repository
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3

# Checkout version tag
git checkout v3.3.0

# Build collector
cd collector
mkdir -p build
cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
make -j4

# Binary location: /tmp/pganalytics-v3/collector/build/src/pganalytics
```

### Step 4: Stop Current Collector

```bash
#!/bin/bash
# If running as systemd service
sudo systemctl stop pganalytics-collector.service

# If running as cron
# Remove from crontab
crontab -e
# Remove line: */1 * * * * /opt/pganalytics/bin/pganalytics

# Wait for graceful shutdown
sleep 5

# Verify it stopped
ps aux | grep pganalytics | grep -v grep
# Should return empty
```

### Step 5: Install New Binary

```bash
#!/bin/bash
INSTALL_DIR="/opt/pganalytics"

# Create installation directory
sudo mkdir -p "$INSTALL_DIR/bin"
sudo mkdir -p "$INSTALL_DIR/lib"

# Install binary
sudo cp /tmp/pganalytics-v3.3.0/collector/build/src/pganalytics \
    "$INSTALL_DIR/bin/pganalytics-v3.3.0"

# Create symlink for easy access
sudo ln -sf "$INSTALL_DIR/bin/pganalytics-v3.3.0" \
    "$INSTALL_DIR/bin/pganalytics"

# Set permissions
sudo chmod 755 "$INSTALL_DIR/bin/pganalytics"
sudo chown pganalytics:pganalytics "$INSTALL_DIR/bin/pganalytics"

# Verify installation
ls -lah "$INSTALL_DIR/bin/pganalytics"
"$INSTALL_DIR/bin/pganalytics" --version
```

### Step 6: Update Configuration

```bash
# Copy new configuration template
sudo cp /tmp/pganalytics-v3.3.0/collector/config.toml \
    /etc/pganalytics/collector.toml.new

# Compare with existing (if upgrading)
diff /etc/pganalytics/collector.toml \
     /etc/pganalytics/collector.toml.new

# Review new options manually
# Key Phase 1 additions:
# [postgresql]
# query_stats_limit = 100        (NEW)
# pool_min_size = 2              (NEW)
# pool_max_size = 10             (NEW)
# [collector_threading]
# thread_pool_size = 4           (NEW)

# Merge configurations (preserve custom settings)
# sudo cp /etc/pganalytics/collector.toml \
#     /etc/pganalytics/collector.toml.backup
# # Manual merge or use vimdiff
# vimdiff /etc/pganalytics/collector.toml \
#         /etc/pganalytics/collector.toml.new

# For new installations
sudo cp /etc/pganalytics/collector.toml.new \
    /etc/pganalytics/collector.toml
```

### Step 7: Verify Configuration

```bash
# Validate TOML syntax
/opt/pganalytics/bin/pganalytics --validate-config

# Expected output:
# Configuration loaded successfully
# Collector ID: collector-001
# Backend URL: https://backend:8080
```

### Step 8: Start Collector

```bash
#!/bin/bash
# Option 1: Direct execution
/opt/pganalytics/bin/pganalytics cron

# Option 2: Systemd service
sudo systemctl start pganalytics-collector.service

# Option 3: Cron job
# Add to crontab:
# */1 * * * * /opt/pganalytics/bin/pganalytics cron

# Verify it's running
sleep 5
ps aux | grep pganalytics | grep -v grep
```

### Step 9: Monitor Initial Execution

```bash
#!/bin/bash
# Check logs for first execution
tail -f /var/log/pganalytics/collector.log

# Expected log output:
# Starting collection loop (collect every 60s, push every 60s, config pull every 300s)
# Collecting metrics...
# CollectorManager thread pool initialized with 4 threads
# DEBUG: Collecting query stats for database: postgres
# Parallel collection completed in XXXms
# Pushing XX metrics to backend...
# Metrics pushed successfully
```

---

## Configuration

### Phase 1 New Configuration Options

Add these sections to `/etc/pganalytics/collector.toml`:

```toml
[postgresql]
# Query statistics collection limit
# Controls how many top queries are collected from pg_stat_statements
#
# Recommendations:
#  - Development (< 100 QPS):     100  (full collection)
#  - Small Prod (100-1K QPS):     500  (5% sampling)
#  - Medium Prod (1K-10K QPS):    1000 (1-10% sampling)
#  - Large Prod (10K+ QPS):       5000 (0.1-1% sampling)
#
# Default: 100 (backward compatible)
# Min: 10, Max: 10000
query_stats_limit = 100

# Connection pool settings
# Controls how many persistent connections to keep open

# Minimum connections to keep in pool (default: 2)
# Set higher if you have bursty collection patterns
pool_min_size = 2

# Maximum connections to allow (default: 10)
# Increase if you have many collector threads
pool_max_size = 10

# Connection idle timeout in seconds (default: 300)
# Close connections idle for this long
pool_idle_timeout = 300

[collector_threading]
# Thread pool configuration for parallel collector execution

# Number of worker threads for collector execution (default: 4)
# Recommended: number of CPU cores
# Set to 1 for sequential execution (backward compatible)
thread_pool_size = 4
```

### Configuration Tuning by Scale

#### Small Deployments (10-25 collectors)

```toml
[postgresql]
query_stats_limit = 100
pool_min_size = 1
pool_max_size = 5

[collector_threading]
thread_pool_size = 2
```

#### Medium Deployments (25-50 collectors)

```toml
[postgresql]
query_stats_limit = 500
pool_min_size = 2
pool_max_size = 10

[collector_threading]
thread_pool_size = 4
```

#### Large Deployments (50-100 collectors)

```toml
[postgresql]
query_stats_limit = 1000
pool_min_size = 3
pool_max_size = 15

[collector_threading]
thread_pool_size = 8
```

#### Enterprise Deployments (100+ collectors per instance)

```toml
[postgresql]
query_stats_limit = 5000
pool_min_size = 5
pool_max_size = 20

[collector_threading]
thread_pool_size = 16
```

---

## Validation

### Functional Validation

```bash
#!/bin/bash
# Test 1: Verify binary execution
/opt/pganalytics/bin/pganalytics cron &
COLLECTOR_PID=$!
sleep 5
kill $COLLECTOR_PID

# Expected: Process runs, collects metrics, exits cleanly

# Test 2: Verify configuration loading
/opt/pganalytics/bin/pganalytics --validate-config

# Expected output: Configuration loaded successfully

# Test 3: Check collector.log for errors
tail -100 /var/log/pganalytics/collector.log | grep -i error

# Expected: No critical errors (warnings OK)
```

### Performance Validation

```bash
#!/bin/bash
# Run 3 collection cycles and measure performance
echo "Starting performance validation (3 cycles)..."

TIMES=()
for i in {1..3}; do
    START=$(date +%s%N)
    /opt/pganalytics/bin/pganalytics cron > /dev/null 2>&1
    END=$(date +%s%N)
    ELAPSED_MS=$(( (END - START) / 1000000 ))
    TIMES+=($ELAPSED_MS)
    echo "Cycle $i: ${ELAPSED_MS}ms"
done

# Calculate average
AVG=$(( (${TIMES[0]} + ${TIMES[1]} + ${TIMES[2]}) / 3 ))
echo ""
echo "Average cycle time: ${AVG}ms"
echo ""
echo "Expected:"
echo "  - 10-25 collectors: 1-4 seconds"
echo "  - 25-50 collectors: 4-9 seconds"
echo "  - 50-100 collectors: 9-15 seconds"

# If > 15 seconds, check:
# 1. PostgreSQL connection status
# 2. Network latency to backend
# 3. Collector.log for warnings
# 4. System CPU and memory
```

### Metrics Validation

```bash
#!/bin/bash
# Verify metrics are being collected and pushed
echo "Checking backend for metrics..."

# Query backend API (adjust credentials)
curl -s -H "Authorization: Bearer $COLLECTOR_TOKEN" \
    https://backend:8080/api/v1/metrics \
    -H "Content-Type: application/json" | jq '.metrics | length'

# Expected: > 0 (metrics present)

# Check specific metric types
curl -s -H "Authorization: Bearer $COLLECTOR_TOKEN" \
    https://backend:8080/api/v1/metrics \
    -H "Content-Type: application/json" | \
    jq '.metrics[] | .type' | sort | uniq -c

# Expected output should include:
# pg_query_stats
# pg_stats
# sysstat
# disk_usage
```

### Query Stats Collection Validation

```bash
#!/bin/bash
# Verify query statistics are being collected correctly
tail -50 /var/log/pganalytics/collector.log | grep -E "pg_query_stats|sampling_percent"

# Expected output:
# DEBUG: Collecting query stats for database: postgres
# DEBUG: Successfully collected 100 queries from postgres
# Pool metrics - acquisitions: X, reuses: Y, active: Z/10
# Parallel collection completed in XXXms
```

---

## Rollback Plan

### Emergency Rollback (If Issues Detected)

```bash
#!/bin/bash
BACKUP_DIR="/backup/pganalytics-backup-YYYYMMDD-HHMMSS"

# 1. Stop current instance
sudo systemctl stop pganalytics-collector.service

# 2. Restore previous binary
sudo cp "$BACKUP_DIR/pganalytics" /opt/pganalytics/bin/pganalytics

# 3. Restore configuration
sudo cp "$BACKUP_DIR/collector.toml" /etc/pganalytics/collector.toml

# 4. Restart with previous version
sudo systemctl start pganalytics-collector.service

# 5. Verify it's running
sleep 5
ps aux | grep pganalytics | grep -v grep

# 6. Check logs for errors
tail -20 /var/log/pganalytics/collector.log
```

### Rollback Verification

```bash
# Verify rollback successful
/opt/pganalytics/bin/pganalytics --version

# Should show previous version (e.g., v3.2.0)

# Check that old behavior is restored
tail -20 /var/log/pganalytics/collector.log

# Should NOT show:
# "thread pool initialized"
# "connection pool"
# "parallel collection"
```

---

## Performance Monitoring

### Key Metrics to Monitor

#### 1. Collector Cycle Time

```bash
# Extract cycle times from logs
tail -1000 /var/log/pganalytics/collector.log | \
    grep "Parallel collection completed" | \
    grep -oE "[0-9]+ms" | \
    awk '{sum+=$1; count++} END {print "Average:", sum/count, "ms"}'

# Expected:
# 10 collectors:  1-2 seconds
# 25 collectors:  2-5 seconds
# 50 collectors:  5-10 seconds
# 100 collectors: 9-15 seconds (Phase 1 target: <15s)
```

#### 2. CPU Utilization

```bash
# Monitor CPU during collection
while true; do
    echo "CPU at $(date '+%H:%M:%S'):"
    top -b -n 1 | grep pganalytics | awk '{print $9 "%"}'
    sleep 60
done

# Expected:
# 10 collectors:  < 5% CPU
# 50 collectors:  < 20% CPU
# 100 collectors: < 30% CPU (Phase 1 target: <50%)
```

#### 3. Memory Usage

```bash
# Monitor memory during collection
watch -n 5 'ps aux | grep pganalytics | grep -v grep | awk "{print \$6 \" MB\"}"'

# Expected:
# Stable memory usage, no continuous growth
# ~100-150 MB per collector instance
```

#### 4. Connection Pool Statistics

```bash
# Extract pool metrics from logs
tail -1000 /var/log/pganalytics/collector.log | \
    grep "Pool metrics" | \
    tail -1

# Expected output:
# Pool metrics - acquisitions: 100, reuses: 95, active: 2/10
# Shows connections are being reused (reuses should be high)
```

#### 5. Query Sampling Metrics

```bash
# Check sampling percentage
tail -1000 /var/log/pganalytics/collector.log | \
    grep "sampling_percent" | \
    awk -F'[=%]' '{sum+=$NF; count++} END {print "Average sampling:", sum/count "%"}'

# Expected:
# Default (limit=100): 1-100% depending on QPS
# Configured (limit=500): 5-100% at high QPS
```

### Grafana Dashboard Queries

Add these Prometheus queries to monitor Phase 1 metrics:

```promql
# Cycle time trend
rate(pganalytics_collection_cycle_ms[5m])

# CPU utilization
process_cpu_seconds_total{job="pganalytics-collector"}

# Memory usage
process_resident_memory_bytes{job="pganalytics-collector"}

# Connection pool reuse rate
rate(pganalytics_connection_pool_reuses[5m]) /
rate(pganalytics_connection_pool_acquisitions[5m])
```

---

## Troubleshooting

### Issue 1: Collector Won't Start

**Symptoms**:
- Binary exits immediately
- No logs generated

**Solutions**:

```bash
# 1. Check configuration syntax
/opt/pganalytics/bin/pganalytics --validate-config

# 2. Test PostgreSQL connection
psql -h postgres.host -U collector -d postgres -c "SELECT 1;"

# 3. Check thread pool initialization
# Add to config: thread_pool_size = 1 (disable threading for diagnosis)

# 4. Run with verbose logging
/opt/pganalytics/bin/pganalytics cron 2>&1 | tee /tmp/debug.log

# 5. Check system resources
free -m
df -h
```

### Issue 2: High Cycle Times (> 15s @ 100 collectors)

**Symptoms**:
- Collector taking too long per cycle
- Missing collection windows

**Solutions**:

```bash
# 1. Check thread pool is initialized
grep "thread pool initialized" /var/log/pganalytics/collector.log

# Expected: "CollectorManager thread pool initialized with X threads"

# 2. Verify connection pooling is active
grep "Connection pool" /var/log/pganalytics/collector.log

# Expected: "Connection pool initialized with X threads"

# 3. Check PostgreSQL query performance
psql -h postgres.host -c "SELECT version();"
psql -h postgres.host -c "SELECT count(*) FROM pg_stat_statements;"

# 4. Monitor PostgreSQL connections
ps aux | grep postgres | grep -i idle
# Should have multiple idle connections (from pool)

# 5. Check network latency
ping -c 5 backend.host
# Should be < 50ms

# 6. Reduce thread pool size if too many context switches
# Update config: thread_pool_size = 2
```

### Issue 3: Connection Pool Exhaustion

**Symptoms**:
- "Failed to acquire connection from pool" in logs
- Increasing query latency

**Solutions**:

```bash
# 1. Check pool metrics
grep "Pool metrics" /var/log/pganalytics/collector.log | tail -5

# If active connections == max connections consistently:
# Increase pool_max_size in config

# 2. Check for hung connections
psql -h postgres.host -c "SELECT * FROM pg_stat_activity WHERE state = 'idle';"

# 3. Reduce idle timeout to disconnect stale connections
# Update config: pool_idle_timeout = 120 (was 300)

# 4. Increase pool_min_size to warm up more connections
# Update config: pool_min_size = 5 (was 2)
```

### Issue 4: Memory Leak (Growing Memory Usage)

**Symptoms**:
- Memory usage increases over time
- Eventually causes OOM

**Solutions**:

```bash
# 1. Verify no continuous allocations
# Monitor over 1 hour
for i in {1..60}; do
    ps aux | grep pganalytics | grep -v grep | awk '{print $6}' >> /tmp/memory.log
    sleep 60
done
# Check if line is constantly growing

# 2. If memory grows: likely buffer issue
# Check buffer usage
grep "buffer" /var/log/pganalytics/collector.log

# 3. Reduce query limit (queries consume buffer space)
# Update config: query_stats_limit = 50 (was 100)

# 4. Reduce pool_max_size (each connection has overhead)
# Update config: pool_max_size = 5 (was 10)

# 5. Reduce thread_pool_size (threads have stack overhead)
# Update config: thread_pool_size = 2 (was 4)
```

### Issue 5: Metrics Not Appearing in Backend

**Symptoms**:
- Collector logs show "Metrics pushed successfully"
- But metrics not visible in backend

**Solutions**:

```bash
# 1. Verify authentication token
grep "AUTH\|token" /var/log/pganalytics/collector.log

# 2. Check backend connectivity
curl -v https://backend:8080/api/v1/health

# 3. Verify metrics format
tail -50 /var/log/pganalytics/collector.log | grep -A 5 "Pushing"

# 4. Check backend logs
tail -100 /var/log/backend/api.log | grep -i "collector\|metrics"

# 5. Retry with small query limit
# Update config: query_stats_limit = 10
# Restart collector
# Check if metrics appear
```

---

## Health Check Script

```bash
#!/bin/bash
# Place in /opt/pganalytics/bin/health-check.sh

echo "=== pgAnalytics Health Check ==="
echo "Timestamp: $(date)"
echo ""

# 1. Check if running
if pgrep -f "pganalytics cron" > /dev/null; then
    echo "✅ Collector is running"
else
    echo "❌ Collector is NOT running"
    exit 1
fi

# 2. Check cycle time
LAST_CYCLE=$(grep "Parallel collection completed" \
    /var/log/pganalytics/collector.log | tail -1 | \
    grep -oE "[0-9]+ms")
echo "✅ Last cycle time: $LAST_CYCLE"

# 3. Check metrics push
LAST_PUSH=$(grep "Metrics pushed successfully" \
    /var/log/pganalytics/collector.log | tail -1 | \
    awk '{print $(NF-3) " " $(NF-2) " " $(NF-1) " " $NF}')
echo "✅ Last metrics push: $LAST_PUSH"

# 4. Check for errors
ERROR_COUNT=$(grep "ERROR\|CRITICAL" \
    /var/log/pganalytics/collector.log | wc -l)
if [ $ERROR_COUNT -gt 0 ]; then
    echo "⚠️  Found $ERROR_COUNT errors in logs"
else
    echo "✅ No errors in logs"
fi

# 5. Check memory
MEMORY=$(ps aux | grep pganalytics | grep -v grep | awk '{print $6}')
echo "✅ Memory usage: ${MEMORY}MB"

echo ""
echo "Health check complete"
```

---

## Post-Deployment Validation (24-48 Hours)

### Daily Monitoring

- ✅ Collector running continuously
- ✅ Metrics flowing to backend
- ✅ No errors in logs
- ✅ Cycle time stable (< 15s for 100 collectors)
- ✅ CPU utilization stable (< 50% for 100 collectors)
- ✅ Memory stable (no growth)
- ✅ Connection pool healthy (good reuse rate)

### Weekly Monitoring

- ✅ Query sampling metrics valid
- ✅ Pool health checks passing
- ✅ No memory leaks detected
- ✅ Performance consistent across time
- ✅ Backend receiving all expected metrics

---

## Support & Issues

### If You Encounter Issues

1. **Check logs first**:
   ```bash
   tail -200 /var/log/pganalytics/collector.log
   ```

2. **Validate configuration**:
   ```bash
   /opt/pganalytics/bin/pganalytics --validate-config
   ```

3. **Run health check**:
   ```bash
   /opt/pganalytics/bin/health-check.sh
   ```

4. **Gather diagnostics**:
   ```bash
   # Create diagnostic bundle
   mkdir -p /tmp/pganalytics-diag
   cp /var/log/pganalytics/collector.log /tmp/pganalytics-diag/
   cp /etc/pganalytics/collector.toml /tmp/pganalytics-diag/
   ps aux | grep pganalytics > /tmp/pganalytics-diag/processes.txt
   free -m > /tmp/pganalytics-diag/memory.txt
   df -h > /tmp/pganalytics-diag/disk.txt
   ```

5. **Report issue** with:
   - `/tmp/pganalytics-diag/` contents
   - `pganalytics --version`
   - System information (OS, CPU, memory)
   - Collector count
   - Expected vs actual cycle time

---

## Rollback Decision Tree

```
Is the deployment working?
│
├─ YES: Metrics flowing, no errors
│   └─ Monitor for 24-48 hours
│
└─ NO: Issues detected
    │
    ├─ Collector won't start
    │   └─ ROLLBACK: Restore previous binary
    │
    ├─ Cycle time too high (> 20s @ 100 col)
    │   ├─ Check PostgreSQL performance first
    │   ├─ If still high after 1 hour: ROLLBACK
    │   └─ If issue resolves: CONTINUE
    │
    ├─ Memory leak (growing memory)
    │   ├─ Try config adjustments (reduce limits)
    │   ├─ If resolved: CONTINUE
    │   └─ If continues: ROLLBACK
    │
    └─ Metrics not appearing
        ├─ Check backend connectivity
        ├─ Verify authentication
        ├─ If resolved: CONTINUE
        └─ If continues: ROLLBACK
```

---

## Deployment Complete Checklist

After deployment, verify:

- [ ] Binary installed and executable
- [ ] Configuration updated with Phase 1 options
- [ ] Collector starts without errors
- [ ] Metrics appear in backend within 2 minutes
- [ ] Cycle time < 15 seconds (100 collectors)
- [ ] CPU utilization < 50% (100 collectors)
- [ ] Memory stable (no growth)
- [ ] Connection pool initialized
- [ ] Thread pool with 4 workers
- [ ] Logs show query stats collection
- [ ] Sampling percentage metrics present
- [ ] Pool reuse rate high (>80%)
- [ ] 24-hour monitoring shows stability
- [ ] Team notified of successful deployment

---

**Deployment Guide Version**: 1.0
**Last Updated**: February 26, 2026
**Status**: ✅ READY FOR USE
