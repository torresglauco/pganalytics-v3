---
phase: 10-collector-backend-foundation
plan: 03
subsystem: host-monitoring
tags:
  - host-status
  - os-metrics
  - host-inventory
  - c++-collector
requires:
  - 10-01 (replication models pattern)
  - 10-02 (logical replication routes pattern)
provides:
  - Host status detection (up/down based on collector last_seen)
  - OS metrics collection (CPU, memory, disk I/O, load average)
  - Host inventory (OS version, hardware specs, PostgreSQL config)
affects:
  - API endpoints /hosts/*
  - TimescaleDB tables metrics_host_metrics, metrics_host_inventory
  - C++ collector host_inventory_plugin
tech-stack:
  added:
    - Go host_store.go with 6 database operations
    - Go handlers_host.go with 3 HTTP handlers
    - C++ host_inventory_plugin.cpp with OS/PostgreSQL collection
  patterns:
    - Threshold-based status detection (default 5 minutes)
    - Time range filtering for metrics queries
    - Graceful PostgreSQL connection fallback
key-files:
  created:
    - backend/pkg/models/host_models.go
    - backend/internal/storage/host_store.go
    - backend/internal/api/handlers_host.go
    - collector/include/host_inventory_plugin.h
    - collector/src/host_inventory_plugin.cpp
  modified:
    - backend/migrations/031_replication_tables.sql
    - backend/internal/api/server.go
decisions:
  - Use collectors.last_seen for status detection (no separate table)
  - Default threshold 5 minutes for host down detection
  - Retain host metrics 90 days, inventory 365 days
  - C++ plugin gracefully handles missing PostgreSQL connection
metrics:
  duration: 15 min
  tasks: 6
  files_created: 5
  files_modified: 2
  commits: 3
---

# Phase 10 Plan 03: Host Monitoring Backend Summary

## One-liner

Host monitoring backend with status detection, OS metrics, and inventory collection via C++ plugin.

## Must-Haves Delivered

### Truths Implemented

1. **User can view host up/down status based on collector last_seen timestamp**
   - GET /api/v1/hosts returns all host statuses with is_healthy boolean
   - GET /api/v1/hosts/:id/status returns single host with configurable threshold
   - Status calculated from collectors.last_seen with default 5-minute threshold

2. **User can view OS metrics (CPU, memory, disk I/O, load average) from sysstat collector**
   - GET /api/v1/hosts/:id/metrics returns time-series metrics
   - Supports time_range filter: 1h, 24h, 7d, 30d
   - Data stored in metrics_host_metrics TimescaleDB hypertable

3. **User can view host inventory (OS version, hardware specs, PostgreSQL configuration)**
   - GET /api/v1/hosts/:id/inventory returns static host configuration
   - Includes OS name/version/kernel, CPU model/cores/MHz, memory/disk totals
   - PostgreSQL version, port, data directory, and key settings

4. **Host status endpoint returns is_healthy boolean with configurable threshold**
   - threshold query parameter (default 300 seconds)
   - is_healthy = true if last_seen within threshold
   - Status values: "up", "down", "unknown"

### Artifacts Created

| Artifact | Provides | Exports |
|----------|----------|---------|
| backend/pkg/models/host_models.go | Data structures for host monitoring | HostStatus, HostMetrics, HostInventory |
| backend/internal/storage/host_store.go | Database operations for host metrics | GetHostStatus, StoreHostMetrics, GetHostMetrics, StoreHostInventory, GetHostInventory, GetAllHostStatuses |
| backend/internal/api/handlers_host.go | HTTP handlers for host endpoints | handleGetHostStatus, handleGetHostMetrics, handleGetHostInventory |
| collector/src/host_inventory_plugin.cpp | C++ collection of host inventory data | collectOsInfo, collectCpuInfo, collectMemoryInfo, collectDiskInfo, collectPostgresConfig |

## Deviations from Plan

None - plan executed exactly as written.

## Implementation Details

### Host Status Detection (HOST-01)

- Queries collectors table for last_seen timestamp
- Compares against current time minus threshold (default 5 minutes)
- Returns status: "up" (healthy), "down" (unhealthy), or "unknown" (null last_seen)
- Calculates unresponsive_for_seconds for down hosts

### Host Metrics Storage (HOST-02)

- metrics_host_metrics table with TimescaleDB hypertable
- Columns: CPU percentages, load averages, memory stats, disk stats, I/O ops
- Retention: 90 days
- Indexes on (collector_id, time DESC) for efficient queries

### Host Inventory Collection (HOST-03)

- metrics_host_inventory table with TimescaleDB hypertable
- Static configuration data changes infrequently
- Retention: 365 days
- C++ plugin collects from:
  - /etc/os-release for OS name/version
  - uname syscall for kernel version
  - /proc/cpuinfo for CPU model/cores/MHz
  - /proc/meminfo for memory total
  - statvfs for disk capacity
  - pg_settings for PostgreSQL configuration

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/hosts | Get all host statuses |
| GET | /api/v1/hosts/:id/status | Get single host status |
| GET | /api/v1/hosts/:id/metrics | Get host OS metrics |
| GET | /api/v1/hosts/:id/inventory | Get host inventory |

## Commits

| Commit | Message |
|--------|---------|
| 4b912f4 | feat(10-03): create host store for status, metrics, and inventory |
| 24147bd | feat(10-03): create host API handlers and wire routes |
| eab0572 | feat(10-03): create C++ host inventory collector plugin |

## Verification

- Backend compiles: `go build ./backend/pkg/... ./backend/internal/...` - PASSED
- Host routes registered in server.go - PASSED (4 references)
- Host tables in migration - PASSED (2 tables)
- C++ plugin methods verified - PASSED (6 references)

## Self-Check: PASSED

- [x] All files exist at specified paths
- [x] All commits verified in git history
- [x] Backend compiles successfully
- [x] API routes properly wired with auth middleware