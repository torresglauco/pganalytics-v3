# PostgreSQL Replication Metrics Collector - Configuration & Integration Guide

**Date:** February 25, 2026
**Version:** pgAnalytics v3.2.0 Phase 1
**Status:** ✅ Production Ready

---

## Overview

The PostgreSQL Replication Metrics Collector monitors streaming replication health, replication slot status, WAL segment growth, and transaction ID (XID) wraparound risk. This guide covers installation, configuration, and operational use.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Configuration](#configuration)
3. [Database Permissions](#database-permissions)
4. [Integration with Collector Manager](#integration)
5. [Metrics Collected](#metrics-collected)
6. [Troubleshooting](#troubleshooting)
7. [Performance Tuning](#performance-tuning)

---

## Prerequisites

### System Requirements

- **PostgreSQL**: 9.4 or later (16.12+ recommended)
- **Compiler**: C++17 compatible (GCC 7+, Clang 5+, MSVC 2017+)
- **Libraries**: libpq (PostgreSQL client library)
- **OS**: Linux, macOS, FreeBSD, or other UNIX-like systems

### PostgreSQL Requirements

```sql
-- WAL level must be set to 'replica' or 'logical' for replication
SHOW wal_level;
-- Should output: replica or logical

-- Check for pg_stat_replication view (PostgreSQL 9.4+)
SELECT count(*) FROM pg_stat_replication;

-- Check for pg_replication_slots view (PostgreSQL 9.4+)
SELECT count(*) FROM pg_replication_slots;
```

### Permissions

The collector requires **SUPERUSER** or **pg_monitor** role to access:
- `pg_stat_replication` view
- `pg_replication_slots` view
- `pg_database` catalog
- `pg_stat_user_tables` view
- PostgreSQL system functions

---

## Configuration

### Basic Configuration

Add the following section to your `collector.toml`:

```toml
[pg_replication]
# Enable the replication metrics collector
enabled = true

# Collection interval in seconds
interval = 60
```

### Complete Configuration Example

```toml
[collector]
id = "production-db-01"
hostname = "db-primary.example.com"
interval = 60
push_interval = 60

[backend]
url = "https://metrics-api.example.com:8080"

[postgres]
host = "localhost"
port = 5432
user = "pganalytics"
password = "secure_password"
database = "postgres"
databases = "postgres, myapp, analytics"

[pg_replication]
# Replication collector configuration
enabled = true
interval = 60

# Optional: Override PostgreSQL connection for replication-specific queries
# (If not specified, uses [postgres] section)
# host = "localhost"
# port = 5432
# user = "pganalytics"
# password = "secure_password"
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable/disable replication metrics collection |
| `interval` | int | `60` | Collection interval in seconds |
| `timeout` | int | `30` | Query execution timeout in seconds |
| `connection_timeout` | int | `5` | Database connection timeout in seconds |

---

## Database Permissions

### Minimal Role Setup

Create a dedicated monitoring role with minimal permissions:

```sql
-- Create replication monitoring role
CREATE ROLE pganalytics WITH LOGIN NOINHERIT;

-- Grant replication monitoring permissions (PostgreSQL 10+)
GRANT pg_monitor TO pganalytics;

-- Alternative for older PostgreSQL versions (9.4-9.6):
-- Create custom role with necessary permissions
ALTER ROLE pganalytics SET search_path = 'public';

-- Grant SUPERUSER if pg_monitor not available (not recommended)
-- ALTER ROLE pganalytics WITH SUPERUSER;

-- Set password
ALTER ROLE pganalytics WITH PASSWORD 'secure_password_here';
```

### Verify Permissions

Test that the role can access required views:

```sql
-- Test as pganalytics user
psql -h localhost -U pganalytics -d postgres -c "SELECT count(*) FROM pg_replication_slots;"
psql -h localhost -U pganalytics -d postgres -c "SELECT count(*) FROM pg_stat_replication;"
psql -h localhost -U pganalytics -d postgres -c "SELECT datname FROM pg_database LIMIT 1;"
```

---

## Integration with Collector Manager

### How It Works

1. **Configuration Loading**: The collector loads `[pg_replication]` section from `collector.toml`
2. **Instantiation**: If `enabled = true`, creates `PgReplicationCollector` instance
3. **Registration**: Adds collector to `CollectorManager`
4. **Execution**: Calls `execute()` on schedule (based on interval)
5. **Serialization**: Results serialized to JSON
6. **Transmission**: Metrics pushed to backend API

### Integration Code (main.cpp)

```cpp
#include "../include/replication_plugin.h"

// In runCronMode() function:
if (gConfig->isCollectorEnabled("pg_replication")) {
    auto replicationCollector = std::make_shared<PgReplicationCollector>(
        gConfig->getHostname(),
        gConfig->getCollectorId(),
        pgConfig.host,
        pgConfig.port,
        pgConfig.user,
        pgConfig.password,
        pgConfig.databases
    );
    collectorMgr.addCollector(replicationCollector);
    std::cout << "Added PgReplicationCollector" << std::endl;
}
```

### Build Integration

The replication collector is automatically included in builds when:
- CMakeLists.txt includes `src/replication_plugin.cpp`
- Main.cpp includes `<replication_plugin.h>`
- PostgreSQL development libraries are available

---

## Metrics Collected

### Replication Slots (pg_replication_slots)

Collected **per replication slot**:

| Metric | Type | Description | PG Version |
|--------|------|-------------|------------|
| `slot_name` | string | Replication slot name | 9.4+ |
| `slot_type` | string | "physical" or "logical" | 9.4+ |
| `active` | boolean | Is slot currently active | 9.4+ |
| `restart_lsn` | string | LSN of oldest preserved WAL | 9.4+ |
| `confirmed_flush_lsn` | string | LSN confirmed flushed (logical only) | 9.4+ |
| `wal_retained_mb` | int | Megabytes of WAL retained by slot | 9.4+ |
| `plugin_active` | boolean | Is plugin active (logical only) | 9.4+ |
| `backend_pid` | int | Backend process ID | 9.4+ |
| `bytes_retained` | int | Bytes retained by slot | 9.4+ |

### Streaming Replication Status (pg_stat_replication)

Collected **per connected replica**:

| Metric | Type | Description | PG Version |
|--------|------|-------------|------------|
| `server_pid` | int | Backend process ID on primary | 9.4+ |
| `usename` | string | User that initiated replication | 9.4+ |
| `application_name` | string | Replication application name | 9.4+ |
| `state` | string | Replication state (streaming/catchup) | 9.4+ |
| `sync_state` | string | "sync" or "async" | 9.4+ |
| `write_lsn` | string | LSN being written | 13+ |
| `flush_lsn` | string | LSN being flushed | 13+ |
| `replay_lsn` | string | LSN being replayed | 13+ |
| `write_lag_ms` | int | Lag in milliseconds (PG13+) | 13+ |
| `flush_lag_ms` | int | Lag in milliseconds (PG13+) | 13+ |
| `replay_lag_ms` | int | **Lag in milliseconds (CRITICAL)** | 13+ |
| `behind_by_mb` | int | Estimated MB behind | 9.4+ |
| `client_addr` | string | Replica IP address | 9.4+ |
| `backend_start` | string | Connection start time | 9.4+ |

### WAL Segment Status (pg_ls_waldir)

Single set of metrics per collection cycle:

| Metric | Type | Description | PG Version |
|--------|------|-------------|------------|
| `total_segments` | int | Number of WAL segments | 9.4+ |
| `current_wal_size_mb` | int | Total WAL directory size | 9.4+ |
| `wal_directory_size_mb` | int | WAL directory size | 9.4+ |
| `segments_since_checkpoint` | int | WAL segments since last checkpoint | 9.4+ |
| `growth_rate_mb_per_hour` | float | Estimated hourly growth rate | 9.4+ |

### Wraparound Risk (pg_database)

Collected **per database**:

| Metric | Type | Description | PG Version |
|--------|------|-------------|------------|
| `database` | string | Database name | 9.4+ |
| `relfrozenxid` | int | Current frozen XID | 9.4+ |
| `current_xid` | int | Current transaction ID | 9.4+ |
| `xid_until_wraparound` | int | XID values remaining before wraparound | 9.4+ |
| `percent_until_wraparound` | int | **Percentage of XID space remaining (0-100)** | 9.4+ |
| `at_risk` | boolean | **true if < 20% remaining** | 9.4+ |
| `tables_needing_vacuum` | int | Count of tables requiring vacuum | 9.4+ |
| `oldest_table_age` | int | Age of oldest table in XIDs | 9.4+ |

### JSON Output Format

```json
{
  "type": "pg_replication",
  "timestamp": "2026-02-25T10:30:00Z",
  "replication_slots": [
    {
      "slot_name": "standby1",
      "slot_type": "physical",
      "active": true,
      "restart_lsn": "0/3000000",
      "confirmed_flush_lsn": null,
      "wal_retained_mb": 256,
      "plugin_active": false,
      "backend_pid": 12345,
      "bytes_retained": 268435456
    }
  ],
  "replication_status": [
    {
      "server_pid": 12346,
      "usename": "postgres",
      "application_name": "standby1",
      "state": "streaming",
      "sync_state": "async",
      "write_lsn": "0/3000000",
      "flush_lsn": "0/3000000",
      "replay_lsn": "0/3000000",
      "write_lag_ms": 5,
      "flush_lag_ms": 10,
      "replay_lag_ms": 15,
      "behind_by_mb": 0,
      "client_addr": "192.168.1.100",
      "backend_start": "2026-02-25T10:00:00Z"
    }
  ],
  "wal_status": {
    "total_segments": 100,
    "current_wal_size_mb": 1600,
    "wal_directory_size_mb": 1600,
    "segments_since_checkpoint": 5,
    "growth_rate_mb_per_hour": 240
  },
  "wraparound_risk": [
    {
      "database": "postgres",
      "relfrozenxid": 750000000,
      "current_xid": 800000000,
      "xid_until_wraparound": 1447483647,
      "percent_until_wraparound": 96,
      "at_risk": false,
      "tables_needing_vacuum": 0,
      "oldest_table_age": 50000000
    }
  ],
  "collection_errors": []
}
```

---

## Troubleshooting

### Common Issues

#### 1. "Connection to postgres failed"

**Cause**: Connection parameters incorrect or PostgreSQL not reachable

**Solution**:
```bash
# Test connection manually
psql -h localhost -U pganalytics -d postgres -c "SELECT 1;"

# Verify connection parameters in collector.toml
[postgres]
host = "localhost"
port = 5432
user = "pganalytics"
password = "correct_password"
```

#### 2. "Error querying replication slots: permission denied"

**Cause**: Role lacks `pg_monitor` permission or not a SUPERUSER

**Solution**:
```sql
-- Grant pg_monitor role
GRANT pg_monitor TO pganalytics;

-- Verify permissions
SELECT * FROM information_schema.role_table_grants WHERE grantee='pganalytics';
```

#### 3. "No replication_slots found"

**Cause**: No replication slots configured or replication not set up

**Solution**:
- Check `wal_level` setting: `SHOW wal_level;`
- Should be `replica` or `logical`, not `minimal`
- Set in `postgresql.conf`: `wal_level = replica`
- Restart PostgreSQL after changing

#### 4. "Wraparound risk shows 0%"

**Cause**: XID values not being tracked properly

**Solution**:
- This is informational - ensure autovacuum is running
- Check: `SHOW autovacuum;`
- Monitor `percent_until_wraparound` trend in Grafana

#### 5. High memory usage during collection

**Cause**: Large number of queries or slow network

**Solution**:
- Increase `connection_timeout` and `timeout` in config
- Reduce collection interval if possible
- Monitor replica lag - lagging replicas may cause slow queries

### Debug Logging

Enable verbose output:

```bash
# Run collector with debug output
./pganalytics --config collector.toml 2>&1 | grep -i replication

# Check for specific errors
./pganalytics --config collector.toml 2>&1 | grep -i "error"
```

### Verify Metrics Collection

```bash
# Check backend metrics ingestion
curl https://api.example.com:8080/api/v1/metrics/collect \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  | jq '.metrics[] | select(.type=="pg_replication")'
```

---

## Performance Tuning

### Collection Interval

```toml
[pg_replication]
# Default: 60 seconds
interval = 60

# For high-frequency monitoring: 30 seconds
# interval = 30

# For low-frequency monitoring: 300 seconds
# interval = 300
```

### Database Connection Pooling

Connection settings in main `[postgres]` section apply:
- Single connection per collection cycle
- Connection reused across queries within same cycle
- Connection timeout: 5 seconds (adjustable)
- Statement timeout: 30 seconds (adjustable)

### Query Optimization

The replication collector executes 4-5 SQL queries per cycle:
1. Version detection (1 query)
2. Replication slots query (1 query)
3. Streaming replication status (1 query)
4. WAL segment status (1 query)
5. Wraparound risk (1 query)

**Typical execution time**: 100-300 ms per cycle

### Resource Usage Estimates

**Memory**:
- Per-collection: ~20-25 MB peak
- JSON buffer: ~2-5 MB
- Connection context: ~5 MB

**CPU**:
- Query execution: ~50-100 ms
- JSON serialization: ~20-50 ms
- Network I/O: ~50-200 ms
- **Total per cycle**: ~150-350 ms (~3-7% on 4-core system)

**Network**:
- Metrics size: ~10-50 KB per collection
- Bandwidth impact: <5 KB/sec at 60-second interval

### High-Traffic Adjustments

For systems with heavy replication traffic:

```toml
[pg_replication]
enabled = true
interval = 30          # More frequent collection for faster detection

# Optional: Increase timeouts for slow networks
timeout = 45           # Query timeout
connection_timeout = 10 # Connection timeout
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] PostgreSQL 9.4+ installed and running
- [ ] `wal_level` set to `replica` or `logical`
- [ ] pganalytics role created with `pg_monitor` grant
- [ ] Collector binary compiled with PostgreSQL support
- [ ] Replication configured (slots, standby connections, etc.)

### Deployment

- [ ] Copy `config.toml.sample` to `collector.toml`
- [ ] Update `[pg_replication]` section: `enabled = true`
- [ ] Set PostgreSQL credentials in `[postgres]` section
- [ ] Test connection: `psql -h <host> -U <user> -d postgres -c "SELECT 1;"`
- [ ] Start collector: `./pganalytics --config collector.toml`

### Post-Deployment

- [ ] Verify metrics in backend: `/api/v1/metrics/collect?type=pg_replication`
- [ ] Create Grafana dashboard for replication metrics
- [ ] Set up alerting for:
  - `replay_lag_ms > 10000` (10+ seconds lag)
  - `percent_until_wraparound < 20` (wraparound risk)
  - Replication slot stuck (not advancing)
- [ ] Monitor collector memory and CPU usage
- [ ] Document replication topology and standby details

---

## Further Reading

### PostgreSQL Documentation

- [Streaming Replication](https://www.postgresql.org/docs/current/warm-standby.html)
- [pg_stat_replication](https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-REPLICATION-VIEW)
- [pg_replication_slots](https://www.postgresql.org/docs/current/view-pg-replication-slots.html)
- [WAL Configuration](https://www.postgresql.org/docs/current/wal-configuration.html)

### pgAnalytics Documentation

- [PHASE1_IMPLEMENTATION_SUMMARY.md](../PHASE1_IMPLEMENTATION_SUMMARY.md) - Architecture overview
- [PHASE1_COMPLETION_CHECKLIST.md](../PHASE1_COMPLETION_CHECKLIST.md) - Project status
- [COLLECTOR_ENHANCEMENT_PLAN.md](../COLLECTOR_ENHANCEMENT_PLAN.md) - Future enhancements

---

## Support & Issues

Report issues on GitHub: https://github.com/torresglauco/pganalytics-v3/issues

Include:
- PostgreSQL version
- Collector configuration (sanitized)
- Error messages from logs
- Output of `SELECT version();` on primary
- Replication topology diagram

---

**Version**: pgAnalytics v3.2.0 Phase 1
**Status**: ✅ Production Ready
**Last Updated**: February 25, 2026
