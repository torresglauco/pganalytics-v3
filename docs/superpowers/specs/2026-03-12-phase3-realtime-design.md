# pgAnalytics Phase 3: Real-Time Features & Data Integration

> **For agentic workers:** Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement real-time log ingestion, WebSocket streaming, and alert execution for PostgreSQL observability.

**Architecture:**
- Collectors send ERROR and SLOW_QUERY logs via HTTP POST to backend
- Backend ingests logs, persists to PostgreSQL, broadcasts via WebSocket
- Frontend receives real-time logs and displays in live stream
- Background worker evaluates alert rules every 60 seconds
- When alert triggers, async worker delivers notifications via configured channels

**Tech Stack:** React 18, TypeScript, WebSocket, Go backend, PostgreSQL, Zustand, Recharts

---

## Phase 3: Real-Time Features & Data Integration Design

### 1. Frontend Real-Time Architecture

#### 1.1 WebSocket Client Service

**File:** `src/services/realtime.ts`

Core WebSocket client that manages connection lifecycle and event handling.

**Responsibilities:**
- Establish WebSocket connection with JWT authentication
- Handle connection/disconnection events
- Emit and listen for real-time events (log:new, metric:update, alert:triggered)
- Auto-reconnect with exponential backoff
- Queue messages while offline, sync on reconnect

**Events:**
- `log:new` - New ERROR or SLOW_QUERY log arrived
- `metric:update` - Aggregated metrics updated
- `alert:triggered` - Alert rule condition satisfied

**Error Handling:**
- Connection lost? Auto-reconnect (backoff: 1s → 2s → 4s → 8s → 30s max)
- Connection fails after 5 retries? Emit error event, fall back to polling
- Server closes connection? Clear queue, reset state

#### 1.2 Real-Time Hook & Zustand Store

**File:** `src/hooks/useRealtime.ts`

React hook for consuming real-time data in components.

```typescript
export const useRealtime = () => {
  const { connected, lastUpdate } = useRealtimeStore()
  const { subscribe, unsubscribe } = useRealtimeStore()

  return {
    connected,        // boolean
    lastUpdate,       // ISO timestamp
    subscribe,        // (event: string, callback) => void
    unsubscribe,      // (event: string, callback) => void
  }
}
```

**File:** `src/stores/realtimeStore.ts`

Zustand store managing connection state and event subscriptions.

**State:**
- `connected: boolean` - WebSocket connected
- `lastUpdate: string | null` - ISO timestamp of last event
- `error: string | null` - Connection error message
- `subscriptions: Map<string, Set<Callback>>` - Event listeners

#### 1.3 Frontend Components - Live Logs

**File:** `src/components/logs/LiveLogsStream.tsx`

New component displaying real-time log stream for current instance.

**Features:**
- Displays last 50 logs with auto-scroll
- "LIVE" badge when connected, "CONNECTING..." when offline
- Auto-scroll toggle (pause to read, resume to follow)
- Color-coded by level (ERROR = red, SLOW_QUERY = orange)
- Click log to open details modal

**Update:** `src/components/logs/LogsViewer.tsx`

Add new section for live stream above historical logs table:

```tsx
<div className="space-y-6">
  {/* NEW: Live Stream Section */}
  <div className="border-t pt-6">
    <h2 className="text-lg font-semibold mb-4">Live Stream</h2>
    <LiveLogsStream />
  </div>

  {/* EXISTING: Historical Logs Table */}
  <div>
    <h2 className="text-lg font-semibold mb-4">Historical Logs</h2>
    <LogsTable ... />
  </div>
</div>
```

#### 1.4 Frontend Components - Real-Time Metrics

**Update:** `src/components/metrics/MetricsViewer.tsx`

Subscribe to `metric:update` events and refresh charts when metrics change.

```typescript
const MetricsViewer: React.FC = () => {
  const { connected } = useRealtime()

  useEffect(() => {
    const handleMetricUpdate = (data: any) => {
      // Update Recharts data with smooth transition
      setMetrics(prev => ({
        ...prev,
        ...data
      }))
    }

    subscribe('metric:update', handleMetricUpdate)
    return () => unsubscribe('metric:update', handleMetricUpdate)
  }, [])

  return (
    <div>
      <div className="text-sm text-slate-500">
        {connected ? '🟢 Live' : '⚪ Polling'}
        Last update: {lastUpdate}
      </div>
      {/* Charts update automatically */}
    </div>
  )
}
```

#### 1.5 Frontend Components - Alert Notifications

**Update:** `src/stores/notificationStore.ts`

When `alert:triggered` event received, add notification to toast queue.

```typescript
const handleAlertTriggered = (alertData: any) => {
  notificationStore.addToast({
    type: 'warning',
    title: `Alert: ${alertData.alert_name}`,
    message: alertData.description,
    action: {
      label: 'View',
      onClick: () => navigate(`/alerts/${alertData.alert_id}`)
    }
  })
}

subscribe('alert:triggered', handleAlertTriggered)
```

**New Component:** `src/components/common/RealtimeStatus.tsx`

Visual indicator in header showing real-time connection status.

```tsx
<div className="flex items-center gap-2">
  {connected ? (
    <>
      <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
      <span className="text-xs text-green-600 dark:text-green-400">Live</span>
    </>
  ) : (
    <>
      <div className="w-2 h-2 bg-yellow-500 rounded-full" />
      <span className="text-xs text-yellow-600 dark:text-yellow-400">Polling</span>
    </>
  )}
</div>
```

---

### 2. Backend Log Ingestion (Go)

#### 2.1 Log Ingest Handler

**File:** `pkg/handlers/logs.go`

HTTP endpoint for collectors to submit logs.

```
POST /api/v1/logs/ingest
Authorization: Bearer {API_TOKEN}

Request Body:
{
  "collector_id": "uuid",
  "instance_id": 1,
  "logs": [
    {
      "timestamp": "2026-03-12T18:00:00Z",
      "level": "ERROR",
      "message": "Syntax error in query",
      "source_location": "src/query.c:1234",
      "process_id": 12345,
      "user_name": "postgres",
      "error_code": "42601",
      "error_detail": "syntax error at or near SELEC"
    },
    {
      "timestamp": "2026-03-12T18:00:01Z",
      "level": "SLOW_QUERY",
      "message": "SELECT * FROM large_table",
      "duration_ms": 5000,
      "query_text": "SELECT * FROM large_table WHERE id > 1000000"
    }
  ]
}

Response:
{
  "success": true,
  "ingested": 2,
  "errors": []
}
```

**Handler Steps:**
1. Extract JWT token from Authorization header
2. Validate token and get collector/user
3. Validate instance_id belongs to collector
4. For each log:
   - Validate timestamp (not in future, not older than 24h)
   - Validate level is ERROR or SLOW_QUERY
   - Insert into postgresql_logs table
5. Publish WebSocket event for each log
6. Return 200 with count

**Error Handling:**
- Missing token → 401 Unauthorized
- Invalid token → 401 Unauthorized
- Invalid instance_id → 403 Forbidden
- Malformed body → 400 Bad Request
- Database error → 500 Internal Server Error

#### 2.2 WebSocket Handler

**File:** `pkg/handlers/realtime.go`

WebSocket endpoint for real-time event streaming.

```
GET /api/v1/ws
Authorization: Bearer {JWT_TOKEN}
Upgrade: websocket
Connection: Upgrade

Events sent to client:
{
  "type": "log:new",
  "data": {
    "id": 12345,
    "timestamp": "2026-03-12T18:00:00Z",
    "level": "ERROR",
    "message": "...",
    "instance_id": 1
  }
}

{
  "type": "metric:update",
  "data": {
    "instance_id": 1,
    "error_count": 42,
    "slow_query_count": 5,
    "timestamp": "2026-03-12T18:00:00Z"
  }
}

{
  "type": "alert:triggered",
  "data": {
    "alert_id": 123,
    "alert_name": "High Error Rate",
    "instance_id": 1,
    "triggered_at": "2026-03-12T18:00:00Z"
  }
}
```

**Handler Steps:**
1. Validate JWT token
2. Get user and their accessible instances
3. Upgrade HTTP connection to WebSocket
4. Register connection in connection manager
5. Listen for messages (heartbeat/ping)
6. Send events filtered by user's instances
7. Clean up connection on close

#### 2.3 WebSocket Connection Manager

**File:** `pkg/services/websocket.go`

Manages active WebSocket connections and broadcasts events.

```go
type ConnectionManager struct {
  connections map[string]*Connection  // user_id → connections
  mu          sync.RWMutex
}

type Connection struct {
  userID     string
  instances  []int           // instances user can access
  conn       *websocket.Conn
  send       chan interface{}
  done       chan bool
}

// Broadcast log to all connected users with access to instance
func (cm *ConnectionManager) BroadcastLog(log *Log) {
  for _, conn := range cm.connections {
    if conn.hasAccess(log.InstanceID) {
      conn.send <- map[string]interface{}{
        "type": "log:new",
        "data": log,
      }
    }
  }
}
```

---

### 3. Alert Execution (Go)

#### 3.1 Alert Evaluation Worker

**File:** `pkg/services/alert_worker.go`

Background job that evaluates alert rules every 60 seconds.

**Execution Flow:**
1. Every 60 seconds, fetch all active alert_rules
2. For each alert:
   - Get conditions from alert_rules
   - Query logs from last 60s matching conditions
   - If match count > threshold AND not recently triggered:
     - Create alert_trigger record
     - Publish WebSocket event
     - Mark in-flight notification for worker
3. Handle errors gracefully (log and continue)

**Trigger Deduplication:**
- Check: has alert been triggered in last 5 minutes?
- If yes, skip (avoid alert fatigue)
- If no, create new trigger

**Conditions Evaluation:**
```go
// Example: "Error count > 10"
type Condition struct {
  Type      string // "error_count", "slow_query_count"
  Operator  string // "greater_than", "less_than", "equals"
  Threshold int
}

// Query: count logs where level='ERROR' in last 60s
count := db.Count("SELECT COUNT(*) FROM postgresql_logs WHERE log_level='ERROR' AND log_timestamp > now() - interval '60s'")
if count > condition.Threshold {
  // Trigger alert
}
```

#### 3.2 Notification Worker

**File:** `pkg/services/notification_worker.go`

Async worker that delivers notifications when alerts trigger.

**Execution Flow:**
1. Query notifications table for status='pending'
2. For each pending notification:
   - Get channel configuration
   - Call appropriate delivery function (sendEmail, sendSlack, sendWebhook)
   - If success: mark status='delivered', set sent_at
   - If error: increment retry_count, mark failed
3. Run continuously (check every 5 seconds)

**Delivery Implementations:**
- Email: Use SMTP (nodemailer equivalent for Go)
- Slack: HTTP POST to webhook URL
- PagerDuty: HTTP API call
- Webhook: HTTP POST to user's custom URL

**Retry Logic:**
- Max 3 retries per notification
- Exponential backoff: 5s → 30s → 300s
- After 3 failures: mark status='failed'

#### 3.3 Database Models

**File:** `pkg/models/models.go` (additions)

```go
type AlertTrigger struct {
  ID          int64
  AlertID     int64
  InstanceID  int
  TriggeredAt time.Time
  CreatedAt   time.Time
}

type Notification struct {
  ID            int64
  ChannelID     int64
  AlertTriggerID int64
  Status        string // pending, delivered, failed
  RetryCount    int
  SentAt        *time.Time
  CreatedAt     time.Time
  UpdatedAt     time.Time
}
```

#### 3.4 Database Migration

**File:** `backend/migrations/022_realtime_tables.sql`

```sql
-- Alert Triggers
CREATE TABLE IF NOT EXISTS alert_triggers (
  id BIGSERIAL PRIMARY KEY,
  alert_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
  instance_id INTEGER NOT NULL REFERENCES postgresql_instances(id),
  triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  UNIQUE(alert_id, instance_id, DATE(triggered_at))
);

CREATE INDEX idx_alert_triggers_alert_id ON alert_triggers(alert_id);
CREATE INDEX idx_alert_triggers_instance_id ON alert_triggers(instance_id);
CREATE INDEX idx_alert_triggers_triggered_at ON alert_triggers(triggered_at DESC);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  channel_id INTEGER NOT NULL REFERENCES notification_channels(id),
  alert_trigger_id BIGINT NOT NULL REFERENCES alert_triggers(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, delivered, failed
  retry_count INTEGER DEFAULT 0,
  sent_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel_id ON notifications(channel_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
```

---

### 4. API Routes & Integration

#### 4.1 New Routes

**File:** `internal/api/routes.go` (updates)

```go
// Public routes (require API token)
POST   /api/v1/logs/ingest          → handlers.IngestLogs
GET    /api/v1/ws                   → handlers.WebSocketHandler

// Authenticated routes (require JWT)
GET    /api/v1/metrics/realtime     → handlers.GetRealtimeMetrics
GET    /api/v1/alerts/recent        → handlers.GetRecentAlerts
```

---

### 5. Error Handling & Resilience

#### 5.1 Backend Resilience

| Scenario | Behavior |
|----------|----------|
| Log ingest fails | Return 500, collector retries with exponential backoff |
| WebSocket connection drops | Frontend auto-reconnects, no data loss |
| Alert worker crashes | Systemd/Docker supervisor restarts it |
| Notification delivery fails | Retry up to 3 times with backoff |
| Database unavailable | Queue requests, retry when available |

#### 5.2 Frontend Resilience

| Scenario | Behavior |
|----------|----------|
| WebSocket offline | Show "Polling" badge, fall back to GET /api/v1/metrics/realtime every 10s |
| WebSocket reconnected | Resume live stream, sync any missed data |
| Browser closed | Server keeps alerts in queue until delivered |
| Network latency | Queue updates locally, process when connection available |

#### 5.3 Data Consistency

- **Source of truth:** PostgreSQL database
- **WebSocket:** Best-effort delivery (non-critical, can miss)
- **Alerts:** Idempotent (same alert won't fire twice in 5 min window)
- **Notifications:** Persisted in database, retried until delivered

---

### 6. Implementation Files Summary

#### Frontend (7 files)
- `src/services/realtime.ts` - WebSocket client
- `src/hooks/useRealtime.ts` - React hook
- `src/stores/realtimeStore.ts` - Zustand store
- `src/components/logs/LiveLogsStream.tsx` - Live stream component
- `src/components/common/RealtimeStatus.tsx` - Status badge
- Update: `src/components/logs/LogsViewer.tsx`
- Update: `src/components/metrics/MetricsViewer.tsx`

#### Backend (8 files)
- `pkg/handlers/logs.go` - Log ingest endpoint
- `pkg/handlers/realtime.go` - WebSocket handler
- `pkg/services/websocket.go` - Connection manager
- `pkg/services/alert_worker.go` - Alert evaluation
- `pkg/services/notification_worker.go` - Notification delivery
- `pkg/models/models.go` - New models
- `backend/migrations/022_realtime_tables.sql` - Database schema
- Update: `internal/api/routes.go` - New routes

---

## Key Design Decisions

### 1. WebSocket over HTTP Polling
- **Why:** Sub-100ms latency vs 3-5s polling
- **Trade-off:** Slightly more complex, but professional UX

### 2. Batch Alert Evaluation (60s)
- **Why:** Simplicity, no event queue needed
- **Trade-off:** Alert latency up to 60s

### 3. HTTP POST for Log Ingestion
- **Why:** Simple, works with any collector language
- **Trade-off:** Slightly higher overhead than gRPC

### 4. PostgreSQL for All Persistence
- **Why:** Single database, simpler operations
- **Trade-off:** Won't scale to 100k+ logs/sec (but filtered to ERRORS + SLOW only)

### 5. Async Notification Worker
- **Why:** Doesn't block log ingestion
- **Trade-off:** Notifications delayed by up to 5 seconds

---

## Testing Strategy

**Frontend:**
- Unit tests for RealtimeClient (mock WebSocket)
- Component tests for LiveLogsStream
- Integration tests for Zustand store

**Backend:**
- Unit tests for condition evaluation logic
- Integration tests for log ingest endpoint (with test collector)
- End-to-end tests for alert flow (ingest → evaluate → notify)

---

**Plan Status:** ✅ Ready for Implementation
**Next Step:** Use superpowers:subagent-driven-development to execute

