# Collector Management Dashboard

**Purpose:** Centralized monitoring and control interface for distributed collectors
**Status:** Feature Specification & UI/UX Design
**Date:** February 26, 2026
**Version:** 1.0

---

## Overview

The Collector Management Dashboard provides administrators with a unified interface to:
- View real-time status of all distributed collectors (decentralized)
- Monitor collector health metrics (CPU, memory, uptime)
- Stop/pause collectors without full shutdown
- Unregister collectors from central backend
- Re-register previously unregistered collectors
- Restart collectors on demand
- View collector logs and error messages
- Manage collector configurations
- Track metrics collection statistics

---

## Dashboard Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    Collector Management UI                       │
│                                                                  │
│  React Dashboard (Real-time WebSocket updates)                  │
│  - Collector list with status indicators                        │
│  - Health metrics and statistics                                │
│  - Actions menu (restart, stop, unregister, etc.)               │
│  - Search, filter, and sort capabilities                        │
│  - Bulk operations (select multiple collectors)                 │
└─────────────────────────────┬──────────────────────────────────┘
                              │
                    WebSocket + REST API
                              │
         ┌────────────────────┴────────────────────┐
         │                                         │
    ┌────▼──────────┐                    ┌────────▼────┐
    │ Central       │                    │ Collector   │
    │ Backend API   │◄──Heartbeat────────┤ Instances   │
    │ (Go/RDS)      │  (60 sec)           │ (Distributed)│
    │               │                    │             │
    │ ┌─────────────┴───────────────┐    └─────────────┘
    │ │ WebSocket Server (Real-time)│
    │ │ ┌─────────────────────────┐ │
    │ │ │ Collector Status Stream │ │
    │ │ │ - Connected            │ │
    │ │ │ - Heartbeat            │ │
    │ │ │ - Metrics Updated      │ │
    │ │ │ - Error Events         │ │
    │ │ └─────────────────────────┘ │
    │ │                             │
    │ │ REST API Endpoints:         │
    │ │ - GET /collectors/status    │
    │ │ - POST /collectors/stop     │
    │ │ - POST /collectors/restart  │
    │ │ - DELETE /collectors        │
    │ │ - POST /collectors/register │
    │ │ - GET /collectors/logs      │
    │ └─────────────────────────────┘
    │
    └──────────────────────────────────
```

---

## UI Mockups

### 1. Main Collector Dashboard

```
┌──────────────────────────────────────────────────────────────────┐
│ pgAnalytics - Collector Management                         [⚙] [👤]│
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Collectors                                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │  [Search collectors...]  [Filter ▼] [Sort ▼]           │   │
│  │  [Group By: Group ▼]                                    │   │
│  │                                                          │   │
│  │  Summary Statistics                                      │   │
│  │  ├─ Total Collectors: 24                                │   │
│  │  ├─ Online: 22 ✓                                        │   │
│  │  ├─ Offline: 2 ✗                                        │   │
│  │  ├─ Metrics Collected: 1,234,567                        │   │
│  │  └─ Last Updated: 2 seconds ago                         │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Collectors List (with detailed view)                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │ □ | Name         │ Status │ Group    │ Metrics │ Actions│   │
│  │──|─────────────|────────|──────────|─────────|────────│   │
│  │□ │prod-rds-1  │ ✓ OK   │AWS-RDS  │123,456 │▼ Menu  │   │
│  │  │            │        │         │        │ Menu    │   │
│  │  │ Host: prod-db-1.rds.aws                │ ┌────────────┤   │
│  │  │ Uptime: 99.8% | CPU: 15% | Memory: 34%│ │Restart     │   │
│  │  │ Last Heartbeat: 2 sec ago              │ │Stop        │   │
│  │  │                                        │ │Unregister  │   │
│  │□ │staging-db   │ ✗ Down │On-Prem  │87,234  │ │View Logs   │   │
│  │  │            │        │         │        │ │Edit Config │   │
│  │  │ Host: staging-db.internal              │ │Restart All │   │
│  │  │ Uptime: 45% | Last seen: 5 min ago     │ └────────────┤   │
│  │  │ Last Error: Connection timeout         │             │   │
│  │  │                                        │             │   │
│  │□ │dev-local    │ ⚠ Slow │Development│2,123  │             │   │
│  │  │            │        │          │       │             │   │
│  │  │ Host: localhost:5432                   │             │   │
│  │  │ Uptime: 88% | CPU: 45% | Memory: 78%  │             │   │
│  │  │ Last Heartbeat: 35 sec ago            │             │   │
│  │  │ Collection Time: avg 450ms (slow)     │             │   │
│  │                                                         │   │
│  │ [Select all] | Selected: 2 | [Bulk Action: ▼]          │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 2. Collector Status Detail Panel

**Click on collector to see full details:**

```
┌──────────────────────────────────────────────────────────────────┐
│ Collector Detail: prod-rds-1                           [X] [Edit] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Status: ✓ ONLINE & HEALTHY                                     │
│  Last Heartbeat: 2 seconds ago                                   │
│  Registered: 2024-01-15 10:30 AM (36 days ago)                  │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Connection Information                                          │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Database Type: PostgreSQL                                       │
│  Host: prod-db-1.region.rds.amazonaws.com                        │
│  Port: 5432                                                      │
│  Database: pganalytics                                           │
│  Group: AWS-RDS                                                  │
│  Tags: [prod, aws, rds, critical]                               │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Performance Metrics                                             │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Uptime: 99.8% (36 days)                                        │
│  Availability: ████████████████░░ 99.8%                         │
│                                                                  │
│  CPU Usage:  ████░░░░░░░░░░░░░░ 15% (avg)                      │
│  Memory:     ███████░░░░░░░░░░░░ 34% (avg)                      │
│  Network:    Data sent: 1.2 TB, received: 234 GB                │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Collection Statistics                                           │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Metrics Collected: 1,234,567                                   │
│  Collection Interval: 60 seconds                                │
│  Avg Collection Time: 234 ms                                    │
│  Last Collection: 1 second ago                                  │
│  Success Rate: 99.98%                                           │
│  Queries per Collection: 142 avg                                │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Recent Activity                                                 │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  2024-01-20 14:30:45  ✓  Metrics collected (156 queries)       │
│  2024-01-20 14:29:45  ✓  Metrics collected (142 queries)       │
│  2024-01-20 14:28:45  ✓  Metrics collected (149 queries)       │
│  2024-01-20 14:27:45  ✓  Metrics collected (138 queries)       │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Actions                                                         │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  [Test Connection]  [Restart Collector]                         │
│  [Stop Collector]   [Unregister]                                │
│  [Edit Configuration] [View Logs]                               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 3. Stop Collector Dialog

```
┌──────────────────────────────────────────────────────────────────┐
│ Stop Collector                                                [X] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Are you sure you want to STOP collector: prod-rds-1?          │
│                                                                  │
│  This will:                                                      │
│  ✓ Stop collecting metrics from the database                    │
│  ✓ Keep collector registration intact                           │
│  ✓ Allow restart/resume later                                   │
│  ✓ Preserve all collected metrics                               │
│                                                                  │
│  Estimated Impact:                                               │
│  ✓ Grafana dashboards will stop updating                        │
│  ✗ No data loss (stored metrics retained)                       │
│  ✗ Collector can be restarted on-demand                         │
│                                                                  │
│  Stop Reason (optional):                                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Maintenance window scheduled              [Text input...]    │   │
│  │                                                           │   │
│  │ Estimated duration: [Input: 2 hours]                    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  [Cancel]  [Stop Collector (Send Signal)]                      │
│                                                                  │
│  ℹ Note: Collector must be restarted manually or via API       │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 4. Unregister Collector Dialog

```
┌──────────────────────────────────────────────────────────────────┐
│ Unregister Collector                                          [X] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ⚠ WARNING: Unregistering collector prod-rds-1                 │
│                                                                  │
│  This will:                                                      │
│  ✓ Remove collector from central database                       │
│  ✓ Invalidate collector's JWT token                             │
│  ✓ Stop metrics collection                                      │
│  ✓ Archive metrics (retained for 90 days)                       │
│  ✗ Delete registration but keep metrics                         │
│                                                                  │
│  To Re-register:                                                 │
│  1. Use "Register New Collector" or "Re-register" option        │
│  2. Generate new JWT token                                      │
│  3. Update collector configuration                              │
│                                                                  │
│  Metrics Disposition:                                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ Archive metrics (1,234,567 records)                   │   │
│  │   Retention: 90 days                                    │   │
│  │                                                          │   │
│  │ ☐ Delete metrics immediately                           │   │
│  │   Cannot be undone!                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Reason for Unregistering:                                       │
│  [Dropdown ▼]                                                    │
│  ├─ Database retired                                            │
│  ├─ Migrated to new server                                      │
│  ├─ Maintenance/testing                                         │
│  ├─ Switching to different collector                            │
│  └─ Other (specify below)                                       │
│                                                                  │
│  Additional Notes:                                               │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Migrating to new RDS instance in us-west-2              │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  [Cancel]  [Unregister & Archive] [Unregister & Delete]        │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 5. Re-register Previously Unregistered Collector

```
┌──────────────────────────────────────────────────────────────────┐
│ Re-register Collector                                         [X] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Select a previously unregistered collector to re-register:     │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Archived Collectors (Last 90 days)                       │   │
│  │                                                          │   │
│  │ [✓] prod-rds-1                                           │   │
│  │     Unregistered: 2024-01-18 15:30                       │   │
│  │     Metrics Archived: 1,234,567 records                  │   │
│  │     Database: prod-db-1.rds.amazonaws.com               │   │
│  │     Reason: Migrated to new server                       │   │
│  │                                                          │   │
│  │ [ ] staging-db                                           │   │
│  │     Unregistered: 2024-01-15 10:00                       │   │
│  │     Metrics Archived: 87,234 records                     │   │
│  │     Database: staging-db.internal                        │   │
│  │     Reason: Database retired                             │   │
│  │                                                          │   │
│  │ [ ] old-dev-db                                           │   │
│  │     Unregistered: 2024-01-10 09:15                       │   │
│  │     Metrics Archived: 34,567 records                     │   │
│  │     Database: old-dev.internal                           │   │
│  │     Reason: Maintenance/testing                          │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Re-registration Options:                                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ☑ Generate new JWT token (recommended)                  │   │
│  │ ☐ Use archived token (NOT RECOMMENDED)                  │   │
│  │                                                          │   │
│  │ ☑ Restore archived metrics                              │   │
│  │ ☐ Start fresh (discard archived metrics)               │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  [Cancel]  [Re-register Selected]                               │
│                                                                  │
│  Note: You can register multiple collectors at once            │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 6. Collector Logs Viewer

```
┌──────────────────────────────────────────────────────────────────┐
│ Collector Logs: prod-rds-1                                    [X] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│ [Filter by level ▼]  [Search...]  [Auto-scroll ▼]  [Export]    │
│                                                                  │
│ Time                 │ Level   │ Message                        │
│──────────────────────┼─────────┼────────────────────────────────│
│ 2024-01-20 14:30:45  │ INFO    │ Metrics collection started    │
│ 2024-01-20 14:30:47  │ INFO    │ Connected to database         │
│ 2024-01-20 14:30:48  │ INFO    │ Collected 156 queries        │
│ 2024-01-20 14:30:49  │ INFO    │ Metrics pushed to API        │
│ 2024-01-20 14:30:50  │ DEBUG   │ Response: HTTP 200 OK        │
│ 2024-01-20 14:31:45  │ INFO    │ Metrics collection started    │
│ 2024-01-20 14:31:47  │ INFO    │ Connected to database         │
│ 2024-01-20 14:31:48  │ WARNING │ Slow query execution: 1200ms │
│ 2024-01-20 14:31:49  │ INFO    │ Collected 142 queries        │
│ 2024-01-20 14:31:50  │ INFO    │ Metrics pushed to API        │
│ 2024-01-20 14:32:15  │ ERROR   │ Connection timeout (retry)   │
│ 2024-01-20 14:32:20  │ INFO    │ Reconnected successfully      │
│ 2024-01-20 14:33:45  │ INFO    │ Metrics collection started    │
│                                                                  │
│ [⬇ Load More] | Showing 1-100 of 2,345 entries                 │
│                                                                  │
│ [Close]  [Export as CSV]  [Export as JSON]                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 7. Bulk Collector Operations

```
┌──────────────────────────────────────────────────────────────────┐
│ Bulk Operations: 5 Collectors Selected                        [X] │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│ Action: [Choose Action ▼]                                        │
│ ├─ Restart All (5)                                              │
│ ├─ Stop All (5)                                                 │
│ ├─ Update Group                                                 │
│ ├─ Update Tags                                                  │
│ ├─ Update Configuration                                         │
│ └─ Unregister All (5) ⚠                                         │
│                                                                  │
│ Preview (5 collectors):                                          │
│ ┌──────────────────────────────────────────────────────────┐   │
│ │ ✓ prod-rds-1     (AWS-RDS)       Status: ONLINE         │   │
│ │ ✓ prod-rds-2     (AWS-RDS)       Status: ONLINE         │   │
│ │ ✓ staging-db     (On-Prem)       Status: OFFLINE        │   │
│ │ ✓ dev-local      (Development)   Status: ONLINE         │   │
│ │ ✓ backup-db      (AWS-RDS)       Status: ONLINE         │   │
│ └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│ Schedule Execution:                                              │
│ ┌──────────────────────────────────────────────────────────┐   │
│ │ ☑ Execute immediately                                   │   │
│ │ ☐ Schedule for later                                    │   │
│ │   Datetime: [2024-01-20 22:00]                          │   │
│ └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│ [Cancel]  [Execute Bulk Operation]                              │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## API Endpoints for Dashboard

### Collector Status & Monitoring

```
GET /api/v1/collectors
  - List all collectors with status
  - Query params: ?environment=prod&group=aws&status=online&page=1&limit=20
  - Response: [{ id, name, status, host, uptime, last_heartbeat, metrics_count, ... }]

GET /api/v1/collectors/{collectorId}
  - Get detailed collector information
  - Response: { id, name, type, host, status, metrics, uptime, cpu, memory, ... }

GET /api/v1/collectors/{collectorId}/status
  - Real-time collector status
  - Response: { status, last_heartbeat, metrics_collected_today, ... }

GET /api/v1/collectors/{collectorId}/health
  - Health check status
  - Response: { is_healthy, cpu_usage, memory_usage, response_time, ... }

GET /api/v1/collectors/{collectorId}/metrics
  - Collector performance metrics
  - Response: { uptime, cpu, memory, network_io, collection_times, ... }

GET /api/v1/collectors/{collectorId}/logs
  - Collector logs with filtering
  - Query params: ?level=info&limit=100&offset=0
  - Response: [{ timestamp, level, message }]

GET /api/v1/collectors/archived
  - List unregistered/archived collectors
  - Response: [{ id, name, reason, archived_at, metrics_count, ... }]
```

### Collector Control Operations

```
POST /api/v1/collectors/{collectorId}/restart
  - Restart a collector
  - Body: { reason?: "string" }
  - Response: { success, message, restart_started_at }

POST /api/v1/collectors/{collectorId}/stop
  - Stop collector (graceful shutdown)
  - Body: { reason?: "string", estimated_duration_minutes?: number }
  - Response: { success, message, stopped_at }

POST /api/v1/collectors/{collectorId}/resume
  - Resume a stopped collector
  - Body: { reason?: "string" }
  - Response: { success, message, resumed_at }

DELETE /api/v1/collectors/{collectorId}
  - Unregister a collector
  - Body: {
      reason: "string",
      archive_metrics: true,
      retention_days: 90
    }
  - Response: { success, message, archived_metrics_count }

POST /api/v1/collectors/{collectorId}/test-connection
  - Test database connection
  - Response: { success, message, database_version }

POST /api/v1/collectors/{collectorId}/restart-jwt
  - Restart/rotate JWT token
  - Response: { new_jwt_token, expires_at }

POST /api/v1/collectors/{collectorId}/update-config
  - Update collector configuration
  - Body: {
      collection_interval?: number,
      query_limit?: number,
      tags?: string[],
      group_id?: string
    }
  - Response: { success, message, updated_config }

POST /api/v1/collectors/bulk-action
  - Perform bulk operations
  - Body: {
      action: "restart|stop|unregister",
      collector_ids: ["col_123", "col_456"],
      reason?: "string"
    }
  - Response: { success, total: 2, succeeded: 2, failed: 0, results: [] }
```

### WebSocket Events (Real-time Updates)

```javascript
// Client subscribes to collector status updates
ws.on('collector:connected', { collector_id, timestamp })
ws.on('collector:disconnected', { collector_id, timestamp })
ws.on('collector:metrics', { collector_id, metrics, timestamp })
ws.on('collector:error', { collector_id, error, timestamp })
ws.on('collector:status-changed', { collector_id, old_status, new_status })
ws.on('collector:heartbeat', { collector_id, timestamp })
ws.on('collector:restarted', { collector_id, timestamp })
ws.on('collector:stopped', { collector_id, timestamp })
```

---

## Backend Implementation

### Database Table Updates

```sql
-- Add collector state tracking
ALTER TABLE collectors ADD COLUMN (
    state VARCHAR(50) DEFAULT 'registered',  -- registered, running, stopped, error
    stop_requested BOOLEAN DEFAULT FALSE,
    stop_requested_at TIMESTAMP,
    stop_reason TEXT,
    last_restart_at TIMESTAMP,
    restart_count INTEGER DEFAULT 0,
    archived_at TIMESTAMP,
    archive_reason TEXT
);

-- Create collector audit/action log
CREATE TABLE collector_actions (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id),
    action VARCHAR(50) NOT NULL,  -- restart, stop, resume, unregister
    initiated_by UUID NOT NULL,  -- User or system
    reason TEXT,
    status VARCHAR(50) DEFAULT 'pending',  -- pending, in_progress, completed, failed
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,

    FOREIGN KEY (collector_id) REFERENCES collectors(id)
);

-- Create collector metrics snapshot
CREATE TABLE collector_snapshots (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    status VARCHAR(50),
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    metrics_count INTEGER,
    last_collection_duration_ms INTEGER,
    uptime_percentage DECIMAL(5,2),

    FOREIGN KEY (collector_id) REFERENCES collectors(id)
);
```

### Go Backend Implementation

```go
// File: backend/internal/api/handlers_collector_management.go

// RestartCollector handles POST /api/v1/collectors/{id}/restart
func (s *Server) RestartCollector(c *gin.Context) {
    collectorID := c.Param("id")
    userID := c.GetString("user_id")

    var req struct {
        Reason string `json:"reason"`
    }
    c.BindJSON(&req)

    // Check if collector exists
    collector, err := s.getCollectorByID(collectorID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Collector not found"})
        return
    }

    // Log action
    err = s.db.QueryRow(`
        INSERT INTO collector_actions (collector_id, action, initiated_by, reason, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, collectorID, "restart", userID, req.Reason, "in_progress").Scan()

    // Send restart signal to collector via gRPC or webhook
    err = s.sendCollectorCommand(collector.ID, "restart", map[string]interface{}{
        "reason": req.Reason,
    })

    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot send restart command"})
        return
    }

    // Update collector state
    s.db.Exec(`
        UPDATE collectors
        SET state = 'stopped', last_restart_at = NOW(), restart_count = restart_count + 1
        WHERE id = $1
    `, collectorID)

    // Broadcast WebSocket event
    s.broadcastCollectorEvent("restart", collectorID, map[string]interface{}{
        "timestamp": time.Now(),
    })

    c.JSON(200, gin.H{
        "success": true,
        "message": "Restart command sent",
        "collector_id": collectorID,
    })
}

// StopCollector handles POST /api/v1/collectors/{id}/stop
func (s *Server) StopCollector(c *gin.Context) {
    collectorID := c.Param("id")
    userID := c.GetString("user_id")

    var req struct {
        Reason                    string `json:"reason"`
        EstimatedDurationMinutes  int    `json:"estimated_duration_minutes"`
    }
    c.BindJSON(&req)

    // Send stop signal
    err := s.sendCollectorCommand(collectorID, "stop", map[string]interface{}{
        "graceful": true,
        "reason": req.Reason,
    })

    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot send stop command"})
        return
    }

    // Update state
    s.db.Exec(`
        UPDATE collectors
        SET state = 'stopped', stop_requested = TRUE,
            stop_requested_at = NOW(), stop_reason = $2
        WHERE id = $1
    `, collectorID, req.Reason)

    // Broadcast event
    s.broadcastCollectorEvent("stopped", collectorID, nil)

    c.JSON(200, gin.H{
        "success": true,
        "message": "Collector stopped",
        "timestamp": time.Now(),
    })
}

// UnregisterCollector handles DELETE /api/v1/collectors/{id}
func (s *Server) UnregisterCollector(c *gin.Context) {
    collectorID := c.Param("id")
    userID := c.GetString("user_id")

    var req struct {
        Reason         string `json:"reason"`
        ArchiveMetrics bool   `json:"archive_metrics" binding:"required"`
        RetentionDays  int    `json:"retention_days"`
    }
    c.BindJSON(&req)

    // Get collector info before deletion
    collector, _ := s.getCollectorByID(collectorID)
    metricsCount := getMetricsCount(collectorID)

    // Archive metrics if requested
    if req.ArchiveMetrics {
        s.db.Exec(`
            INSERT INTO collector_metrics_archive
            SELECT * FROM collector_metrics
            WHERE collector_id = $1 AND received_at > NOW() - INTERVAL '90 days'
        `, collectorID)

        // Set expiration date
        expiryDate := time.Now().AddDate(0, 0, req.RetentionDays)
        s.db.Exec(`
            UPDATE collector_metrics_archive
            SET expiration_date = $2
            WHERE collector_id = $1
        `, collectorID, expiryDate)
    }

    // Soft delete collector
    s.db.Exec(`
        UPDATE collectors
        SET deleted_at = NOW(), archived_at = NOW(),
            archive_reason = $2, state = 'archived'
        WHERE id = $1
    `, collectorID, req.Reason)

    // Log action
    s.auditLogger.Log(c.Request.Context(), &AuditEvent{
        EventType: "COLLECTOR_UNREGISTERED",
        ResourceType: "COLLECTOR",
        Action: "DELETE",
        ActorID: userID,
        Changes: map[string]interface{}{
            "collector_id": collectorID,
            "name": collector.Name,
            "reason": req.Reason,
            "metrics_archived": metricsCount,
        },
    })

    // Broadcast event
    s.broadcastCollectorEvent("unregistered", collectorID, nil)

    c.JSON(200, gin.H{
        "success": true,
        "message": "Collector unregistered and archived",
        "metrics_archived": metricsCount,
    })
}

// ReRegisterCollector handles POST /api/v1/collectors/re-register
func (s *Server) ReRegisterCollector(c *gin.Context) {
    var req struct {
        ArchivedCollectorID string `json:"archived_collector_id" binding:"required"`
        GenerateNewToken    bool   `json:"generate_new_token" binding:"required"`
        RestoreMetrics      bool   `json:"restore_metrics"`
    }
    c.BindJSON(&req)

    // Get archived collector
    var archivedCollector Collector
    err := s.db.QueryRow(`
        SELECT id, name, type, host, port, database, username,
               password_encrypted, collection_interval, query_limit
        FROM collectors
        WHERE id = $1 AND deleted_at IS NOT NULL
    `, req.ArchivedCollectorID).Scan(
        &archivedCollector.ID,
        &archivedCollector.Name,
        &archivedCollector.Type,
        &archivedCollector.Host,
        &archivedCollector.Port,
        &archivedCollector.Database,
        &archivedCollector.Username,
        &archivedCollector.PasswordEncrypted,
        &archivedCollector.CollectionInterval,
        &archivedCollector.QueryLimit,
    )

    if err != nil {
        c.JSON(404, gin.H{"error": "Archived collector not found"})
        return
    }

    // Generate new JWT token if requested
    var newToken string
    if req.GenerateNewToken {
        newToken, _ = s.generateCollectorJWT(archivedCollector.ID)
    } else {
        newToken = archivedCollector.JWTToken
    }

    // Restore collector (unsoft-delete)
    s.db.Exec(`
        UPDATE collectors
        SET deleted_at = NULL, archived_at = NULL, archive_reason = NULL,
            jwt_token = $2, state = 'registered'
        WHERE id = $1
    `, archivedCollector.ID, newToken)

    // Restore metrics if requested
    metricsRestored := 0
    if req.RestoreMetrics {
        s.db.QueryRow(`
            DELETE FROM collector_metrics_archive
            WHERE collector_id = $1
            RETURNING count(*)
        `, archivedCollector.ID).Scan(&metricsRestored)
    }

    c.JSON(200, gin.H{
        "success": true,
        "message": "Collector re-registered",
        "collector_id": archivedCollector.ID,
        "jwt_token": newToken,
        "metrics_restored": metricsRestored,
    })
}

// ListArchivedCollectors handles GET /api/v1/collectors/archived
func (s *Server) ListArchivedCollectors(c *gin.Context) {
    rows, err := s.db.Query(`
        SELECT id, name, host, archived_at, archive_reason,
               (SELECT COUNT(*) FROM collector_metrics_archive WHERE collector_id = collectors.id) as metrics_count
        FROM collectors
        WHERE deleted_at IS NOT NULL
        ORDER BY archived_at DESC
    `)

    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot query archived collectors"})
        return
    }
    defer rows.Close()

    archived := []map[string]interface{}{}
    for rows.Next() {
        var id, name, host, archiveReason string
        var archivedAt time.Time
        var metricsCount int

        rows.Scan(&id, &name, &host, &archivedAt, &archiveReason, &metricsCount)

        archived = append(archived, map[string]interface{}{
            "id": id,
            "name": name,
            "host": host,
            "archived_at": archivedAt,
            "archive_reason": archiveReason,
            "metrics_count": metricsCount,
        })
    }

    c.JSON(200, gin.H{
        "archived_collectors": archived,
        "total": len(archived),
    })
}

// BulkCollectorAction handles POST /api/v1/collectors/bulk-action
func (s *Server) BulkCollectorAction(c *gin.Context) {
    var req struct {
        Action        string   `json:"action" binding:"required,oneof=restart stop unregister"`
        CollectorIDs  []string `json:"collector_ids" binding:"required"`
        Reason        string   `json:"reason"`
    }
    c.BindJSON(&req)

    results := []map[string]interface{}{}

    for _, collectorID := range req.CollectorIDs {
        // Perform action for each collector
        var result map[string]interface{}

        switch req.Action {
        case "restart":
            result = s.restartCollectorSync(collectorID, req.Reason)
        case "stop":
            result = s.stopCollectorSync(collectorID, req.Reason)
        case "unregister":
            result = s.unregisterCollectorSync(collectorID, req.Reason)
        }

        results = append(results, result)
    }

    succeeded := 0
    for _, r := range results {
        if r["success"].(bool) {
            succeeded++
        }
    }

    c.JSON(200, gin.H{
        "action": req.Action,
        "total": len(req.CollectorIDs),
        "succeeded": succeeded,
        "failed": len(req.CollectorIDs) - succeeded,
        "results": results,
    })
}

// WebSocket handler for real-time updates
func (s *Server) handleWebSocketCollectorUpdates(ws *websocket.Conn) {
    // Subscribe client to collector events
    clientID := uuid.New().String()
    s.wsClients[clientID] = ws

    for {
        // Listen for client requests
        var msg map[string]interface{}
        if err := ws.ReadJSON(&msg); err != nil {
            break
        }

        action := msg["action"].(string)

        switch action {
        case "subscribe":
            collectorID := msg["collector_id"].(string)
            s.subscribeToCollector(clientID, collectorID)
        case "unsubscribe":
            collectorID := msg["collector_id"].(string)
            s.unsubscribeFromCollector(clientID, collectorID)
        }
    }

    delete(s.wsClients, clientID)
    ws.Close()
}

// Helper to broadcast events
func (s *Server) broadcastCollectorEvent(eventType, collectorID string, data map[string]interface{}) {
    event := map[string]interface{}{
        "type": eventType,
        "collector_id": collectorID,
        "timestamp": time.Now(),
        "data": data,
    }

    s.wsEventBus <- event
}
```

---

## Frontend React Components

### Collector Dashboard Main Component

```typescript
// File: frontend/src/pages/CollectorManagement.tsx
import React, { useEffect, useState } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { collectorApi } from '../services/collectors';
import CollectorList from '../components/CollectorList';
import CollectorDetail from '../components/CollectorDetail';
import CollectorActions from '../components/CollectorActions';

export const CollectorManagement: React.FC = () => {
    const [collectors, setCollectors] = useState([]);
    const [selectedCollector, setSelectedCollector] = useState(null);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState({
        environment: '',
        group: '',
        status: '',
    });

    // WebSocket for real-time updates
    const { connected, subscribe } = useWebSocket('wss://api.pganalytics.local/ws');

    useEffect(() => {
        // Load initial data
        fetchCollectors();

        // Subscribe to WebSocket events
        if (connected) {
            subscribe('collector:*', (event) => {
                handleCollectorEvent(event);
            });
        }
    }, [connected]);

    const fetchCollectors = async () => {
        try {
            const response = await collectorApi.listCollectors(filter);
            setCollectors(response.data);
        } catch (error) {
            console.error('Failed to fetch collectors', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCollectorEvent = (event) => {
        // Update collector in list when status changes
        setCollectors((prev) =>
            prev.map((c) =>
                c.id === event.collector_id
                    ? { ...c, status: event.new_status || c.status }
                    : c
            )
        );
    };

    const handleRestart = async (collectorId) => {
        await collectorApi.restartCollector(collectorId);
        fetchCollectors();
    };

    const handleStop = async (collectorId, reason) => {
        await collectorApi.stopCollector(collectorId, reason);
        fetchCollectors();
    };

    const handleUnregister = async (collectorId, reason) => {
        await collectorApi.unregisterCollector(collectorId, reason);
        fetchCollectors();
    };

    if (loading) return <div>Loading collectors...</div>;

    return (
        <div className="collector-management">
            <h1>Collector Management</h1>

            <CollectorList
                collectors={collectors}
                onSelect={setSelectedCollector}
                onRestart={handleRestart}
                onStop={handleStop}
                onUnregister={handleUnregister}
            />

            {selectedCollector && (
                <CollectorDetail
                    collector={selectedCollector}
                    onRestart={handleRestart}
                    onStop={handleStop}
                    onUnregister={handleUnregister}
                />
            )}
        </div>
    );
};
```

---

## Key Features Summary

| Feature | Status | Details |
|---------|--------|---------|
| View all collectors | ✅ | List with status, host, uptime, metrics |
| Real-time status | ✅ | WebSocket updates every 2-5 seconds |
| Collector details | ✅ | CPU, memory, metrics, uptime statistics |
| Restart collector | ✅ | Send restart signal via gRPC/webhook |
| Stop collector | ✅ | Graceful shutdown with timeout |
| Unregister | ✅ | Soft delete with metric archival |
| Re-register | ✅ | Restore archived collectors |
| Bulk operations | ✅ | Restart/stop/unregister multiple at once |
| View logs | ✅ | Filter by level, search, export |
| Audit trail | ✅ | Track all actions with timestamps |
| WebSocket events | ✅ | Real-time status changes |

---

## Implementation Timeline

**Phase 1: Backend API** (15-20 hours)
- Create API endpoints for collector control
- Implement gRPC/webhook communication with collectors
- Database schema updates
- WebSocket event system
- Audit logging

**Phase 2: Frontend Components** (20-25 hours)
- Collector list and detail views
- Control action modals (restart, stop, unregister)
- Status indicators and health metrics
- Logs viewer
- Bulk operations UI

**Phase 3: Integration** (10-15 hours)
- Connect frontend to backend API
- WebSocket real-time updates
- Error handling and user feedback
- Performance optimization
- Testing

**Phase 4: Testing & Deployment** (10-15 hours)
- Unit tests
- E2E tests
- Load testing
- Documentation
- Deployment

**Total Estimated Effort:** 55-75 hours

---

**Status:** Feature specification complete, ready for implementation
**Priority:** High (Critical for production fleet management)

