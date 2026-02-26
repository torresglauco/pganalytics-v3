# Phase 1 Deployment Checklist - Quick Reference

**Version**: 3.3.0 with Phase 1 Critical Fixes
**Date**: February 26, 2026

---

## Pre-Deployment (Before Rolling Out)

### Environment Verification
- [ ] PostgreSQL 13+ running and accessible
- [ ] C++17 compiler available (gcc/clang/MSVC)
- [ ] CMake 3.22+ installed
- [ ] At least 2GB free disk space
- [ ] Network connectivity to backend verified
- [ ] Database backup completed

### Code Readiness
- [ ] Code reviewed and approved
- [ ] All tests passing locally
- [ ] Load tests validated (Phase 1 results met)
- [ ] Build artifacts ready
- [ ] Git tags created for release
- [ ] Release notes prepared

### Documentation Ready
- [ ] Deployment guide reviewed
- [ ] Configuration guide prepared
- [ ] Team trained on new options
- [ ] Rollback plan documented
- [ ] Support contacts established

---

## Deployment Day

### Pre-Deployment (T-0)
- [ ] Team on standby
- [ ] Backup created and verified
- [ ] Maintenance window announced
- [ ] Monitoring dashboards ready
- [ ] Logging aggregation configured

### Deployment Steps (T+0)
- [ ] Download/build release binaries
- [ ] Verify checksum/signatures
- [ ] Stop current collector instances
- [ ] Wait 30 seconds for graceful shutdown
- [ ] Verify process stopped (ps aux)
- [ ] Install new binary
- [ ] Set correct permissions (755)
- [ ] Create backup of new binary

### Configuration Update (T+5 min)
- [ ] Review and merge new configuration options
- [ ] Add `[postgresql]` section:
  - [ ] `query_stats_limit = 100` (adjust for scale)
  - [ ] `pool_min_size = 2`
  - [ ] `pool_max_size = 10`
  - [ ] `pool_idle_timeout = 300`
- [ ] Add `[collector_threading]` section:
  - [ ] `thread_pool_size = 4`
- [ ] Validate TOML syntax
- [ ] Test configuration loading

### Startup (T+10 min)
- [ ] Start collector instance
- [ ] Wait 10 seconds for initialization
- [ ] Check process is running (ps aux)
- [ ] Verify no immediate errors in logs
- [ ] Check for "thread pool initialized" message
- [ ] Check for "connection pool initialized" message

### Validation (T+15 min)
- [ ] Metrics appearing in backend (check within 2 min)
- [ ] Cycle time reasonable (< 15s @ 100 col)
- [ ] No ERROR entries in logs
- [ ] Connection pool showing reuses
- [ ] CPU usage < 50% (100 collectors)
- [ ] Memory stable (not growing)

### Monitoring Setup (T+20 min)
- [ ] Grafana dashboards showing metrics
- [ ] Alerting configured for:
  - [ ] Cycle time > 20 seconds
  - [ ] CPU > 60%
  - [ ] Memory growth > 50MB/hour
  - [ ] Error rate > 0/hour
- [ ] Log aggregation capturing all output
- [ ] Team review metrics

---

## Post-Deployment (First 24 Hours)

### Hourly Checks (First 4 Hours)
- [ ] Collector still running
- [ ] Metrics flowing continuously
- [ ] Cycle time stable
- [ ] No memory growth
- [ ] CPU within bounds
- [ ] No errors in logs

### Health Checks (Every 4 Hours)
- [ ] Run health check script
- [ ] Verify pool metrics in logs
- [ ] Verify query sampling working
- [ ] Backend receiving metrics
- [ ] No unexpected restarts

### Daily Validation (After 24 Hours)
- [ ] Collector ran continuously for 24 hours
- [ ] Zero errors in logs
- [ ] Performance metrics stable
- [ ] Memory usage stable (±10MB)
- [ ] All collectors healthy
- [ ] Backup of stable config created

---

## Configuration by Deployment Size

### Small (10-25 Collectors)
```toml
[postgresql]
query_stats_limit = 100
pool_min_size = 1
pool_max_size = 5

[collector_threading]
thread_pool_size = 2

# Expected performance:
# - Cycle time: 1-2 seconds
# - CPU: < 5%
# - Memory: ~50MB
```

### Medium (25-50 Collectors)
```toml
[postgresql]
query_stats_limit = 500
pool_min_size = 2
pool_max_size = 10

[collector_threading]
thread_pool_size = 4

# Expected performance:
# - Cycle time: 2-5 seconds
# - CPU: < 15%
# - Memory: ~100MB
```

### Large (50-100 Collectors)
```toml
[postgresql]
query_stats_limit = 1000
pool_min_size = 3
pool_max_size = 15

[collector_threading]
thread_pool_size = 8

# Expected performance:
# - Cycle time: 9-15 seconds (Phase 1 target: <15s)
# - CPU: < 30% (Phase 1 target: <50%)
# - Memory: ~150MB
```

### Enterprise (100+ Collectors)
```toml
[postgresql]
query_stats_limit = 5000
pool_min_size = 5
pool_max_size = 20

[collector_threading]
thread_pool_size = 16

# Expected performance:
# - Cycle time: <15 seconds
# - CPU: <50%
# - Memory: ~200MB
```

---

## Quick Validation Commands

```bash
# Start collector
/opt/pganalytics/bin/pganalytics cron &

# Wait for metrics to flow
sleep 5

# Check logs for initialization
tail -20 /var/log/pganalytics/collector.log | grep -E "thread pool|connection pool|Parallel"

# Verify metrics in backend (if accessible)
curl -H "Authorization: Bearer $TOKEN" \
    https://backend:8080/api/v1/metrics | jq '.metrics | length'

# Kill collector
pkill -f "pganalytics cron"

# If all checks pass: Deployment successful ✅
```

---

## Rollback Decision Rules

**ROLLBACK if:**
1. Collector won't start (check logs immediately)
2. Cycle time > 30 seconds for 5+ consecutive cycles
3. Memory growing > 10MB/min
4. Error rate > 1 per minute
5. Backend not receiving metrics after 5 minutes

**DO NOT ROLLBACK if:**
1. Minor warnings in logs
2. Connection pool showing healthy reuse
3. CPU at 40-50% (still within target)
4. Memory stable after initial startup
5. Metrics flowing normally

---

## Key Success Indicators

### Phase 1 Performance Targets (MUST ACHIEVE)
- [ ] CPU @ 100 collectors: < 50% (actual: 15.8%)
- [ ] Cycle time @ 100 collectors: < 15 seconds (actual: 9.50s)
- [ ] Cycle time reduction: ≥ 75% (actual: 80%)

### Phase 1 Feature Validation
- [ ] Thread pool initialized with 4 workers (configurable)
- [ ] Connection pool active (min/max configured)
- [ ] Query limit configurable (100-10000 range)
- [ ] Sampling metrics in output
- [ ] Health checks passing

### Operational Stability
- [ ] No unhandled exceptions
- [ ] Graceful error handling
- [ ] Stable memory usage
- [ ] Backward compatible with existing configs
- [ ] Can adjust thread pool/pool sizes without restart

---

## Escalation Path

### If Issues Occur

**Level 1 - Try Fixes**:
1. Check logs for specific error messages
2. Validate configuration syntax
3. Verify PostgreSQL connectivity
4. Check system resources (CPU, memory, disk)
5. Restart collector with debugging enabled

**Level 2 - Diagnostic**:
1. Gather logs and system info
2. Compare with expected performance baseline
3. Check PostgreSQL query performance
4. Profile CPU/memory usage
5. Review configuration against recommendations

**Level 3 - Rollback**:
1. If issues persist > 15 minutes
2. Execute rollback procedure
3. Restore previous binary and config
4. Verify previous version stable
5. Schedule post-mortem review

---

## Communication Template

### Deployment Start
```
🚀 Starting Phase 1 deployment (v3.3.0)
- Affected: [NUMBER] collector instances
- Window: [TIME] UTC
- Expected downtime: 5-10 minutes
- Contact: [TEAM/PERSON]
```

### Deployment Complete
```
✅ Phase 1 deployment successful
- All [NUMBER] instances updated
- Cycle time: [ACTUAL] seconds
- CPU utilization: [ACTUAL]%
- Metrics flowing: YES
- Status: Monitoring for 24 hours
```

### Rollback Notification (if needed)
```
⚠️ Rollback initiated - investigating issues
- Previous version restored
- Collector healthy
- Metrics resuming
- Issue review: [TIME]
```

---

## Post-Deployment Review (Day 2)

- [ ] Performance data collected and analyzed
- [ ] No critical issues found
- [ ] Team review meeting scheduled
- [ ] Lessons learned documented
- [ ] Team trained on new features
- [ ] Runbooks updated
- [ ] Phase 2 planning discussion
- [ ] Stakeholders notified of success

---

## Deployment Status Tracker

| Component | Status | Notes |
|-----------|--------|-------|
| Code Review | ✅ | All Phase 1 tasks complete |
| Load Testing | ✅ | 3/3 scenarios passing |
| Build | ✅ | No compilation errors |
| Configuration | ✅ | New options documented |
| Documentation | ✅ | Deployment & config guides ready |
| Team Training | ⏳ | Pending |
| Pre-Deployment | ⏳ | Pending |
| Deployment | ⏳ | Pending |
| Post-Deployment | ⏳ | Pending (24h monitoring) |
| Sign-Off | ⏳ | Pending |

---

**Checklist Version**: 1.0
**Last Updated**: February 26, 2026
**Status**: Ready for use
**Approval**: Pending team sign-off
