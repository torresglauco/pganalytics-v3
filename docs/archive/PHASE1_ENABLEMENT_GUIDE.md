# Phase 1 Collector Enablement Guide

**Quick Reference**: How to enable and use the 6 new metric collectors

---

## Enabling Individual Collectors

### 1. SchemaCollector (`pg_schema`)

**Purpose**: Track database schema changes (tables, columns, constraints)

```toml
# In collector/config.toml

[pg_schema]
enabled = true
interval = 300  # Every 5 minutes
```

**Useful for**:
- Schema change auditing
- Finding orphaned columns
- Understanding table relationships
- Compliance tracking

**Requirements**: PostgreSQL 8.0+, no special permissions

---

### 2. LockCollector (`pg_locks`)

**Purpose**: Real-time lock monitoring and blocking detection

```toml
# In collector/config.toml

[pg_locks]
enabled = true
interval = 60   # Every minute
```

**Useful for**:
- Detecting blocking queries
- Troubleshooting deadlocks
- Understanding contention
- Performance investigations

**Requirements**: PostgreSQL 8.1+, Superuser or pg_monitor role

---

### 3. BloatCollector (`pg_bloat`)

**Purpose**: Analyze table and index bloat

```toml
# In collector/config.toml

[pg_bloat]
enabled = true
interval = 300  # Every 5 minutes
```

**Useful for**:
- Identifying bloated tables
- Planning VACUUM/REINDEX
- Storage optimization
- Cost reduction

**Requirements**: PostgreSQL 8.2+, no special permissions

---

### 4. CacheHitCollector (`pg_cache`)

**Purpose**: Monitor buffer pool and cache efficiency

```toml
# In collector/config.toml

[pg_cache]
enabled = true
interval = 60   # Every minute
```

**Useful for**:
- Cache hit ratio trending
- Buffer pool tuning
- Performance optimization
- Identifying hot/cold data

**Requirements**: PostgreSQL 8.1+, no special permissions

---

### 5. ConnectionCollector (`pg_connections`)

**Purpose**: Detailed connection and session tracking

```toml
# In collector/config.toml

[pg_connections]
enabled = true
interval = 60   # Every minute
```

**Useful for**:
- Connection pool monitoring
- Detecting idle transactions
- Long-running query tracking
- Resource exhaustion prevention

**Requirements**: PostgreSQL 9.0+, Superuser for full query text

---

### 6. ExtensionCollector (`pg_extensions`)

**Purpose**: Track installed extensions and versions

```toml
# In collector/config.toml

[pg_extensions]
enabled = true
interval = 300  # Every 5 minutes
```

**Useful for**:
- Extension inventory management
- Dependency tracking
- Version compliance
- Compatibility monitoring

**Requirements**: PostgreSQL 9.1+, no special permissions

---

## Enable All Collectors at Once

Replace the collector sections in `collector/config.toml`:

```toml
[pg_stats]
enabled = true
interval = 60

[sysstat]
enabled = true
interval = 60

[pg_log]
enabled = true
interval = 300

[disk_usage]
enabled = true
interval = 300

[pg_replication]
enabled = true
interval = 60

# Phase 1 - New Collectors
[pg_schema]
enabled = true
interval = 300

[pg_locks]
enabled = true
interval = 60

[pg_bloat]
enabled = true
interval = 300

[pg_cache]
enabled = true
interval = 60

[pg_connections]
enabled = true
interval = 60

[pg_extensions]
enabled = true
interval = 300
```

---

## Building with New Collectors

### 1. Rebuild the collector binary

```bash
cd collector
mkdir -p build
cd build
cmake ..
make
```

### 2. Verify compilation

All 12 plugins (6 existing + 6 new) should compile without errors.

```bash
make install
```

### 3. Test the build

```bash
./pganalytics --version
```

---

## Monitoring Enablement Progress

### Check which collectors are registered

When the collector starts with verbose logging:

```
Starting collector in cron mode...
Configuration loaded successfully
Collector ID: collector-001

Added PgStatsCollector
Added SysstatCollector
Added DiskUsageCollector
Added PgLogCollector
Added PgReplicationCollector
Added PgSchemaCollector          # NEW
Added PgLockCollector             # NEW
Added PgBloatCollector            # NEW
Added PgCacheHitCollector         # NEW
Added PgConnectionCollector       # NEW
Added PgExtensionCollector        # NEW
```

### Recommended Rollout Strategy

#### Week 1: Soft Launch
Enable minimal impact collectors:
```toml
[pg_schema]
enabled = true
interval = 600  # Every 10 minutes

[pg_extensions]
enabled = true
interval = 600

[pg_bloat]
enabled = true
interval = 600
```

#### Week 2: Add Monitoring
Add real-time collectors:
```toml
[pg_locks]
enabled = true
interval = 60

[pg_cache]
enabled = true
interval = 60
```

#### Week 3: Full Deployment
Enable all collectors:
```toml
[pg_connections]
enabled = true
interval = 60
```

---

## Performance Impact

### Individual Collector Impact
| Collector | Query Time | Frequency | Total/Hour |
|-----------|-----------|-----------|-----------|
| SchemaCollector | 200ms | 5m | 12 × 200ms = 2.4s |
| LockCollector | 50ms | 1m | 60 × 50ms = 3s |
| BloatCollector | 500ms | 5m | 12 × 500ms = 6s |
| CacheHitCollector | 500ms | 1m | 60 × 500ms = 30s |
| ConnectionCollector | 50ms | 1m | 60 × 50ms = 3s |
| ExtensionCollector | 20ms | 5m | 12 × 20ms = 240ms |
| **Total** | - | - | **~44.6 seconds/hour** |

### Impact on Collection Cycle
- With all 6 new collectors: ~1-3 seconds per collection cycle
- Existing 6 collectors: ~2-4 seconds per collection cycle
- Total: ~3-7 seconds per full cycle
- Well within 5-second SLA

---

## Troubleshooting

### Collector not appearing in logs
1. Check `enabled = true` in config
2. Verify PostgreSQL connection works
3. Check PostgreSQL version requirements
4. Verify user permissions (especially for locks)

### High query execution time
1. Increase interval (e.g., 300s instead of 60s)
2. Check PostgreSQL server load
3. Verify indexes exist on queried tables
4. Check if tables are very large (> 1M rows)

### Missing data in backend
1. Verify migrations ran successfully
2. Check backend logs for errors
3. Verify collector is sending data with correct type
4. Check network connectivity to backend

---

## Disabling Specific Collectors

If a collector causes issues, disable it:

```toml
[pg_locks]
enabled = false
# The collector will still compile but won't execute
```

---

## Default Configuration (Safe Defaults)

All new collectors default to `enabled = false` for safe rollout:

```toml
[pg_schema]
enabled = false  # Safe to enable

[pg_locks]
enabled = false  # Safe to enable (requires superuser)

[pg_bloat]
enabled = false  # Safe to enable

[pg_cache]
enabled = false  # Safe to enable

[pg_connections]
enabled = false  # Safe to enable (requires superuser for query text)

[pg_extensions]
enabled = false  # Safe to enable
```

Users must explicitly enable collectors they want to use.

---

## Verification Checklist

After enabling collectors:

- [ ] Build completes without errors
- [ ] Collector starts without crashes
- [ ] Configuration is read correctly
- [ ] Collectors appear in startup logs
- [ ] Metrics are collected (check with curl to /metrics endpoint)
- [ ] No performance degradation observed
- [ ] Data appears in backend storage
- [ ] Retention policies are working

---

## Next Steps

1. **Phase 2**: Backend API endpoints will be created to retrieve metrics
2. **Phase 3**: Dashboard widgets will display new metrics
3. **Future**: Advanced analysis and recommendations based on metrics

For questions or issues, see `METRICS_IMPLEMENTATION_PHASE1_COMPLETE.md` for full technical documentation.
