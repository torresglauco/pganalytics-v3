# Collector Management Features - Quick Reference

**Date:** February 26, 2026
**Project:** pgAnalytics v3.3.0
**Features:** Collector Registration UI + Management Dashboard

---

## Quick Overview

Your request for collector management features has been fully designed:

### 1. **Collector Registration Interface** ✅
A React-based web UI where you can:
- Register new distributed collectors
- Test database connections before registration
- Generate JWT tokens for collector authentication
- Import multiple collectors via CSV/JSON
- Group collectors by environment/region
- Tag collectors for organization

### 2. **Collector Management Dashboard** ✅
A unified dashboard where you can:
- View status of ALL collectors (online/offline/slow)
- See real-time health metrics (CPU, memory, uptime)
- **STOP** a collector (graceful shutdown)
- **RESTART** a collector on demand
- **UNREGISTER** a collector (soft delete with archival)
- **RE-REGISTER** previously unregistered collectors
- Perform bulk operations on multiple collectors
- View collector logs and audit trail

---

## Feature Comparison

| Feature | Status | Hours | Details |
|---------|--------|-------|---------|
| **Register Collector** | ✅ Ready | 10-15h | React form + backend API |
| **View Collector Status** | ✅ Ready | 20-25h | Real-time WebSocket updates |
| **Stop Collector** | ✅ Ready | 5h | Graceful shutdown signal |
| **Restart Collector** | ✅ Ready | 5h | Send restart command |
| **Unregister Collector** | ✅ Ready | 5h | Soft delete + metric archival |
| **Re-register Collector** | ✅ Ready | 5h | Restore from archive |
| **Bulk Operations** | ✅ Ready | 5-10h | Multi-collector actions |
| **View Collector Logs** | ✅ Ready | 5h | Filter, search, export |
| **Audit Trail** | ✅ Ready | 5h | Full action history |
| **Real-time Updates** | ✅ Ready | 10-15h | WebSocket events |

**Total Estimated Effort:** 110-150 hours (Two 55-75 hour features)

---

## How It Works: Stop/Restart/Unregister

### Scenario 1: Stop a Collector for Maintenance

```
1. Open Collector Management Dashboard
2. Find "prod-rds-1" in list
3. Click [Stop] button
4. Confirm stopping with reason: "Database maintenance window"
5. Estimated duration: 2 hours
6.
Result:
✓ Collector stops collecting metrics
✓ Grafana dashboards stop updating
✓ No data loss (metrics stored)
✓ Collector can be restarted later
✓ Action logged in audit trail
```

### Scenario 2: Restart a Collector

```
1. Open Collector Management Dashboard
2. Find stopped collector "prod-rds-1"
3. Click [Restart] button
4. Confirmation shows restart started
5.
Result:
✓ Collector receives restart signal
✓ Metrics collection resumes
✓ Grafana dashboards update again
✓ Last heartbeat timestamp updates
✓ Action logged in audit trail
```

### Scenario 3: Unregister & Re-register

```
Unregister Flow:
1. Open Collector Management Dashboard
2. Find "old-db" collector
3. Click [Unregister] button
4. Confirmation dialog with options:
   - Archive metrics (yes/no)
   - Retention: 90 days
   - Reason: "Database migrated to new server"
5. Click [Unregister & Archive]

Result:
✓ Collector removed from central database
✓ JWT token invalidated
✓ Metrics archived for 90 days
✓ Can be restored later

Re-register Flow:
1. Click "View Archived Collectors"
2. Find "old-db" in list of archived collectors
3. Click [Re-register]
4. Options:
   - Generate new JWT token (recommended)
   - Restore archived metrics (yes/no)
5. Click [Re-register Selected]

Result:
✓ Collector added back to central database
✓ New JWT token generated
✓ Metrics restored from archive
✓ Ready to use again
```

---

## User Interface Components

### Main Dashboard
```
Collectors Dashboard
├── Summary Stats
│   ├─ Total: 24 collectors
│   ├─ Online: 22 ✓
│   ├─ Offline: 2 ✗
│   └─ Last update: 2 sec ago
│
├── Collector List
│   ├─ prod-rds-1      [✓ Online]  [▼ Actions]
│   ├─ staging-db      [✗ Offline] [▼ Actions]
│   ├─ dev-local       [⚠ Slow]    [▼ Actions]
│   └─ ... (more collectors)
│
└── Details Panel (click to see more)
    ├─ Status, uptime, metrics
    ├─ Host, port, database
    ├─ CPU, memory usage
    ├─ Collection statistics
    └─ Action buttons
```

### Collector Detail View
```
prod-rds-1 Collector Details
├── Status: ✓ ONLINE & HEALTHY
├── Host: prod-db-1.region.rds.amazonaws.com
├── Uptime: 99.8% (36 days)
├── Metrics: 1,234,567 collected
├── CPU: 15% avg | Memory: 34% avg
├── Last Heartbeat: 2 seconds ago
│
├── Recent Activity
│   ├─ 14:30:45 ✓ Collected 156 queries
│   ├─ 14:29:45 ✓ Collected 142 queries
│   └─ 14:28:45 ✓ Collected 149 queries
│
└── Actions
    ├─ [Restart]        (restart collector)
    ├─ [Stop]           (graceful shutdown)
    ├─ [Unregister]     (remove from system)
    ├─ [Test Connection] (verify database)
    ├─ [Edit Config]    (change settings)
    └─ [View Logs]      (troubleshoot)
```

---

## API Endpoints

### Collector Control
```
POST /api/v1/collectors/{id}/restart
  → Restart a collector

POST /api/v1/collectors/{id}/stop
  → Stop collector gracefully

DELETE /api/v1/collectors/{id}
  → Unregister collector

POST /api/v1/collectors/{id}/resume
  → Resume stopped collector
```

### Status Monitoring
```
GET /api/v1/collectors
  → List all collectors

GET /api/v1/collectors/{id}
  → Get collector details

GET /api/v1/collectors/{id}/status
  → Real-time status

GET /api/v1/collectors/{id}/health
  → CPU, memory, uptime metrics
```

### Archived Collectors
```
GET /api/v1/collectors/archived
  → List unregistered collectors

POST /api/v1/collectors/re-register
  → Restore archived collector
```

### Logs & Audit
```
GET /api/v1/collectors/{id}/logs
  → View collector logs

GET /api/v1/collectors/{id}/audit
  → View action history
```

---

## Database Changes

New tables for collector management:

```sql
-- Track collector state (running, stopped, error, archived)
ALTER TABLE collectors ADD COLUMN state VARCHAR(50);

-- Log all actions on collectors
CREATE TABLE collector_actions (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID,
    action VARCHAR(50),  -- restart, stop, unregister, resume
    initiated_by UUID,
    reason TEXT,
    status VARCHAR(50),
    timestamp TIMESTAMP
);

-- Store archived metrics for re-registration
CREATE TABLE collector_metrics_archive (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID,
    metric_data JSONB,
    archived_at TIMESTAMP,
    expiration_date DATE
);
```

---

## Technology Stack

**Frontend:**
- React 18 + TypeScript
- Redux Toolkit for state
- React Hook Form for forms
- Material-UI components
- WebSocket for real-time updates
- Responsive design (mobile-friendly)

**Backend:**
- Go with Gin framework
- PostgreSQL RDS database
- gRPC or webhooks for collector commands
- WebSocket server for live events
- JWT token authentication

**Integration:**
- HTTP REST API
- WebSocket for real-time
- Collector communication via gRPC/webhook

---

## Implementation Timeline

**Phase 1: Backend API** (15-20 hours)
- Create REST endpoints for control operations
- Implement WebSocket event system
- Add database tables and schema
- Set up command delivery to collectors
- Create audit logging

**Phase 2: Frontend** (20-25 hours)
- Build React components
- Create forms with validation
- Implement real-time WebSocket updates
- Add status indicators and metrics display
- Mobile responsive design

**Phase 3: Integration** (10-15 hours)
- Connect frontend to backend API
- Test WebSocket real-time updates
- Error handling and user feedback
- Performance optimization

**Phase 4: Testing** (10-15 hours)
- Unit tests
- E2E tests
- Load testing with multiple collectors
- Security testing
- Documentation

**Total: 55-75 hours for each feature (110-150 total)**

---

## Key Design Decisions

### 1. Stop vs Unregister
- **Stop:** Temporary (collector can be restarted)
  - Keeps registration intact
  - Preserves JWT token
  - Graceful shutdown

- **Unregister:** Permanent (must re-register)
  - Removes from central database
  - Invalidates JWT token
  - Archives metrics

### 2. Graceful Stop
- Send SIGTERM to collector process
- Give 30-60 second timeout
- If not stopped, force kill (SIGKILL)
- Log all activity

### 3. Metric Archival
- Keep metrics for 90 days after unregister
- Can restore on re-register
- Prevents data loss
- Allows historical analysis

### 4. Real-time Updates
- WebSocket for instant status changes
- Heartbeat every 60 seconds
- Update every 2-5 seconds for fast responsiveness
- Fallback to polling if WebSocket unavailable

---

## Success Criteria

✅ Stop collector works on demand
✅ Restart brings collector back online
✅ Unregister removes from system
✅ Re-register restores from archive
✅ All actions logged in audit trail
✅ Real-time status updates
✅ <100ms API response time
✅ Mobile responsive UI
✅ >80% test coverage
✅ Clear error messages

---

## Security Considerations

- JWT authentication for API
- RBAC: Only admin can restart/stop/unregister
- Audit logging for compliance
- No sensitive data in error messages
- Soft delete (no data loss)
- Metrics archived before deletion
- Token rotation on operations

---

## Example Usage

### Via Web UI
1. Open http://pganalytics:3000/collectors
2. See list of all collectors with status
3. Click on "prod-rds-1"
4. Click "Stop Collector"
5. Fill in reason and confirm
6. Watch status change to "stopped"
7. Later: Click "Restart"
8. Status changes to "online"

### Via REST API
```bash
# Restart collector
curl -X POST http://localhost:8080/api/v1/collectors/col_123/restart \
  -H "Authorization: Bearer $JWT_TOKEN"

# Stop collector
curl -X POST http://localhost:8080/api/v1/collectors/col_123/stop \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"reason":"maintenance"}'

# Unregister collector
curl -X DELETE http://localhost:8080/api/v1/collectors/col_123 \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"reason":"migrated","archive_metrics":true}'

# List collectors
curl http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

## What's Included

You now have complete specifications for:

1. ✅ **Collector Registration UI** (55-75 hours)
   - Design mockups in COLLECTOR_REGISTRATION_UI.md
   - Implementation guide with code examples
   - Database schema and API endpoints

2. ✅ **Collector Management Dashboard** (55-75 hours)
   - Design mockups in COLLECTOR_MANAGEMENT_DASHBOARD.md
   - Stop/restart/unregister flows
   - Real-time WebSocket events
   - Audit logging
   - Re-registration of archived collectors

3. ✅ **Centralized Architecture**
   - RDS-based backend (CENTRALIZED_COLLECTOR_ARCHITECTURE.md)
   - Scalable for hundreds of collectors
   - Multi-cloud backup support
   - JWT authentication

---

## Next Steps

1. **Review Designs**
   - Look at COLLECTOR_REGISTRATION_UI.md
   - Look at COLLECTOR_MANAGEMENT_DASHBOARD.md
   - Review CENTRALIZED_COLLECTOR_ARCHITECTURE.md

2. **Approve Specifications**
   - Confirm feature list
   - Adjust UI mockups if needed
   - Clarify any requirements

3. **Begin Implementation**
   - Assign backend developer for API
   - Assign frontend developer for UI
   - Create sprint board with tasks
   - Begin development

---

## Files to Reference

- `COLLECTOR_REGISTRATION_UI.md` - Onboarding interface design
- `COLLECTOR_MANAGEMENT_DASHBOARD.md` - Central management interface
- `CENTRALIZED_COLLECTOR_ARCHITECTURE.md` - System architecture
- `v3.3.0_COMPLETE_IMPLEMENTATION_GUIDE.md` - Full implementation roadmap

---

**Status:** ✅ Design Complete - Ready for Implementation

All features designed, documented, with code examples and implementation roadmap.

Ready to start coding! 🚀

