# Phase 3: Real-Time Features & Data Integration - Implementation Guide

## Overview

Phase 3 implements comprehensive real-time features for pgAnalytics v3, enabling live log streaming, real-time metrics updates, and instant alert notifications. This phase transforms the platform from a dashboard-based monitoring solution into a real-time data streaming platform.

**Key Features Implemented:**
- Real-time log ingestion and streaming via WebSocket
- Live metrics updates for active instances
- Instant alert notifications across multiple channels (Email, Slack, PagerDuty, Webhooks)
- Automatic alert deduplication and retry logic
- Connection resilience with exponential backoff

**Tech Stack:**
- **Backend**: Go, PostgreSQL, WebSocket (github.com/gorilla/websocket)
- **Frontend**: React, TypeScript, Zustand, WebSocket API
- **Database**: PostgreSQL (primary), TimescaleDB (metrics)
- **Authentication**: JWT tokens for WebSocket connections
- **Message Format**: JSON-based event system

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                      pgAnalytics v3 System                       │
└─────────────────────────────────────────────────────────────────┘

1. External Data Sources
   ├── pg_stat_statements Collector (HTTP POST)
   ├── Log Collector (syslog, journalctl)
   └── Metrics Collector (custom metrics)
                    │
                    ▼
   ┌──────────────────────────────┐
   │  Backend API (Go)            │
   │  ┌────────────────────────┐  │
   │  │ Log Ingest Handler     │  │
   │  │ POST /api/v1/logs/...  │  │
   │  └────────────────────────┘  │
   │  ┌────────────────────────┐  │
   │  │ WebSocket Handler      │  │
   │  │ GET /api/v1/ws         │  │
   │  └────────────────────────┘  │
   │  ┌────────────────────────┐  │
   │  │ Alert Worker (60s)     │  │
   │  │ Evaluates rules        │  │
   │  └────────────────────────┘  │
   │  ┌────────────────────────┐  │
   │  │ Notification Worker    │  │
   │  │ (5s) Sends alerts      │  │
   │  └────────────────────────┘  │
   └──────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   PostgreSQL  TimescaleDB  Cache (Redis)
        │           │
        └───────────┼───────────┘
                    │
        ┌───────────┼──────────────┐
        ▼           ▼              ▼
   WebSocket  EventBus  Notification Services
   Broadcast  (Memory)  (Email, Slack, PagerDuty)
        │
        ▼
   Frontend (React)
   ┌────────────────────────────┐
   │ RealtimeClient             │
   │ WebSocket Connection Mgmt  │
   ├────────────────────────────┤
   │ useRealtimeStore (Zustand) │
   │ Event subscriptions        │
   ├────────────────────────────┤
   │ useRealtime Hook           │
   │ Component integration      │
   ├────────────────────────────┤
   │ LiveLogsStream Component   │
   │ Real-time log display      │
   ├────────────────────────────┤
   │ RealtimeStatus Component   │
   │ Connection indicator       │
   └────────────────────────────┘
```

### Data Flow Diagrams

#### Log Ingestion & Streaming Flow
```
Collector Process
     │
     ├─ Reads PostgreSQL logs from:
     │   • postgres.log (PostgreSQL log file)
     │   • systemd journal
     │   • syslog
     │
     ├─ Parses and filters logs
     │   • Extract timestamp, level, message
     │   • Add instance_id context
     │
     └─ Sends HTTP POST to backend
        HTTP POST /api/v1/logs/ingest
        Authorization: Bearer {API_TOKEN}
        Content-Type: application/json
        │
        ▼
Backend API Handler (logs.go)
        │
        ├─ Validate API token
        ├─ Parse request body
        ├─ Enrich log data
        │   • Add received_at timestamp
        │   • Validate instance_id exists
        │
        └─ Store in PostgreSQL
           INSERT INTO postgresql_logs (...)
           │
           ▼
Database Persistence
           │
           ├─ Index by instance_id
           ├─ Index by timestamp DESC
           └─ Retention: 30 days (configurable)
           │
           ▼
WebSocket Broadcast
           │
        ┌──┴──┬──┬──┐
        ▼     ▼  ▼  ▼
   User1  User2 User3 ... (if user has instance access)
        │
        ▼
Frontend RealtimeClient
        │
        ├─ Receive "log:new" event
        ├─ Update Zustand store
        └─ LiveLogsStream component re-renders
           │
           ▼
User sees new log in real-time (~50-100ms latency)
```

#### Alert Evaluation & Notification Flow
```
Alert Worker (60-second ticker)
        │
        └─ Every 60 seconds:
           │
           ├─ SELECT * FROM alert_rules WHERE active=true
           │   (fetch all active alert rules)
           │
           ├─ FOR EACH rule:
           │   │
           │   ├─ Evaluate condition against last 60 seconds of data
           │   │   • Query PostgreSQL logs table
           │   │   • Count errors, query latency, etc.
           │   │   • Compare against threshold
           │   │
           │   ├─ IF condition matches:
           │   │   │
           │   │   ├─ Check deduplication (5-min window)
           │   │   │   • Don't fire same alert within 5 minutes
           │   │   │
           │   │   ├─ CREATE alert_trigger record
           │   │   │   INSERT INTO alert_triggers (alert_id, instance_id, triggered_at)
           │   │   │
           │   │   ├─ Broadcast "alert:triggered" event
           │   │   │   WebSocket → all connected users with access
           │   │   │
           │   │   └─ Fetch notification channels for this alert
           │   │       SELECT * FROM notification_channels WHERE alert_id = ?
           │   │
           │   └─ FOR EACH channel:
           │       │
           │       └─ CREATE notification record (status='pending')
           │           INSERT INTO notifications (alert_trigger_id, channel_id, status)
           │
           └─ Emit metrics (alert_evaluated, alert_triggered)

Notification Worker (5-second ticker)
        │
        └─ Every 5 seconds:
           │
           ├─ SELECT * FROM notifications WHERE status='pending'
           │   (fetch all pending notifications)
           │
           ├─ FOR EACH notification:
           │   │
           │   ├─ Load channel configuration (email, slack, webhook, etc)
           │   │
           │   ├─ TRY to send notification:
           │   │   │
           │   │   ├─ IF channel_type == 'email'
           │   │   │   └─ Send SMTP email to recipients
           │   │   │
           │   │   ├─ ELSE IF channel_type == 'slack'
           │   │   │   └─ POST to Slack webhook URL
           │   │   │
           │   │   ├─ ELSE IF channel_type == 'pagerduty'
           │   │   │   └─ POST to PagerDuty Events API
           │   │   │
           │   │   └─ ELSE IF channel_type == 'webhook'
           │   │       └─ POST to custom webhook URL
           │   │
           │   ├─ IF send successful:
           │   │   │
           │   │   └─ UPDATE notifications SET status='delivered', sent_at=NOW()
           │   │
           │   └─ IF send failed:
           │       │
           │       ├─ Increment retry_count
           │       ├─ IF retry_count < 3:
           │       │   └─ Leave status='pending' (will retry later)
           │       │   └─ Calculate backoff: 5s * 2^retry_count
           │       │
           │       └─ ELSE (max retries exceeded):
           │           └─ UPDATE notifications SET status='failed'
           │           └─ Log error for manual review
           │
           └─ Emit metrics (notification_sent, notification_failed)

Frontend User
        │
        ├─ Receives "alert:triggered" event
        ├─ Zustand store updates
        └─ Toast notification displayed
           │
           └─ User sees alert in real-time (~100-200ms)

Alert Recipient (Email/Slack/etc)
        │
        └─ Receives notification
           Email: in inbox within seconds
           Slack: message posted to channel immediately
           PagerDuty: incident created
           Webhook: POST sent to callback URL
```

#### WebSocket Connection Flow
```
Frontend User Authenticates
        │
        ├─ User logs in with credentials
        ├─ Backend returns JWT token
        └─ Token stored in localStorage/sessionStorage
           │
           ▼
App.tsx Initialization
        │
        ├─ On app load, check if user is authenticated
        ├─ RealtimeClient.connect(jwtToken) called
        └─ WebSocket connection initiated
           │
           ▼
WebSocket Upgrade
        │
        GET /api/v1/ws?token={JWT_TOKEN}
        Upgrade: websocket
        │
        ▼
Backend Handler (handlers_realtime.go)
        │
        ├─ Extract JWT token from query param
        ├─ Validate token signature
        ├─ Verify token not expired
        ├─ Extract user_id and instance_ids from claims
        └─ Accept WebSocket connection
           │
           ▼
WebSocket Manager Registration
        │
        ├─ Create connection object
        ├─ Store in connections map keyed by user_id
        ├─ Subscribe to events for user's instances
        └─ Send "connected" message to client
           │
           ▼
Frontend RealtimeClient
        │
        ├─ Receive "connected" message
        ├─ Reset reconnection counter
        ├─ Update Zustand store: setConnected(true)
        ├─ Emit "connected" event
        └─ Application can now send/receive real-time events
           │
           ▼
Event Streaming Loop
        │
        ├─ Backend sends messages:
        │   • "log:new" - new log entry
        │   • "metric:update" - metric value changed
        │   • "alert:triggered" - alert fired
        │   • "connection:pong" - heartbeat response
        │
        ├─ Frontend receives and emits events
        └─ Components subscribed to events update state
           │
           ▼
Graceful Disconnect
        │
        ├─ User logs out OR
        ├─ Browser tab closed OR
        ├─ Network disconnected
        │
        ▼
RealtimeClient Detects Disconnect
        │
        ├─ on('close') event fired
        ├─ Update Zustand store: setConnected(false)
        └─ Attempt automatic reconnect (with backoff)
           │
           ├─ Attempt 1: wait 1 second
           ├─ Attempt 2: wait 2 seconds
           ├─ Attempt 3: wait 4 seconds
           ├─ Attempt 4: wait 8 seconds
           ├─ Attempt 5: wait 30 seconds
           │
           └─ If all attempts fail, emit error event
              Frontend shows "Offline" status
              Falls back to polling if configured
```

---

## Backend Implementation

### Database Schema

#### Alert Triggers Table
```sql
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
CREATE INDEX idx_alert_triggers_created_at ON alert_triggers(created_at DESC);
```

**Indexes:**
- `idx_alert_triggers_alert_id`: Fast lookup of triggers for a specific alert
- `idx_alert_triggers_instance_id`: Find all triggers for an instance
- `idx_alert_triggers_triggered_at`: Recent triggers (for UI display)
- `idx_alert_triggers_created_at`: Newest first queries (for deduplication)

**Unique Constraint:**
- Prevents duplicate triggers for the same alert+instance on the same day
- One trigger per alert per instance per calendar day

#### Notifications Table
```sql
CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  channel_id INTEGER NOT NULL REFERENCES notification_channels(id),
  alert_trigger_id BIGINT NOT NULL REFERENCES alert_triggers(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, delivered, failed
  retry_count INTEGER DEFAULT 0,
  last_retry_at TIMESTAMP WITH TIME ZONE,
  sent_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel_id ON notifications(channel_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_alert_trigger_id ON notifications(alert_trigger_id);
```

**Status Values:**
- `pending`: Notification created, waiting to be sent
- `delivered`: Successfully sent to recipient
- `failed`: Send failed after 3 retry attempts

**Running the Migration:**
```bash
# Backend directory
cd backend

# Apply migrations (using your migration tool)
./pganalytics-api migrate

# Or manually:
psql -d pganalytics -f migrations/022_realtime_tables.sql

# Verify tables created
psql -d pganalytics -c "\dt alert_triggers notifications"
```

### API Endpoints

#### Log Ingestion Endpoint
```
POST /api/v1/logs/ingest
Content-Type: application/json
Authorization: Bearer {API_TOKEN}

Request Body:
{
  "collector_id": "123e4567-e89b-12d3-a456-426614174000",
  "instance_id": 1,
  "logs": [
    {
      "timestamp": "2026-03-13T20:00:00Z",
      "level": "ERROR",
      "message": "Deadlock detected on table public.orders",
      "context": {
        "query": "SELECT * FROM orders WHERE id = $1",
        "pid": 12345
      }
    },
    {
      "timestamp": "2026-03-13T20:00:01Z",
      "level": "WARNING",
      "message": "Slow query detected (5234ms): SELECT * FROM big_table",
      "context": {
        "query": "SELECT * FROM big_table",
        "duration_ms": 5234
      }
    }
  ]
}

Response (200 OK):
{
  "success": true,
  "ingested": 2,
  "message": "2 logs successfully ingested"
}

Error Responses:
- 400 Bad Request: Invalid JSON or missing required fields
- 401 Unauthorized: Invalid or missing API token
- 404 Not Found: Instance with given instance_id doesn't exist
- 500 Internal Server Error: Database error
```

**Implementation:**
- File: `backend/pkg/handlers/logs.go`
- Function: `HandleIngestLogs`
- Validates API token (not JWT, but collector API key)
- Validates instance_id exists
- Batch inserts logs into PostgreSQL
- Broadcasts "log:new" event to connected users
- Returns count of ingested logs

#### WebSocket Endpoint
```
GET /api/v1/ws
Authorization: Bearer {JWT_TOKEN}
Upgrade: websocket
Connection: Upgrade

Query Parameters:
- token: JWT token (can be in query param or Authorization header)

WebSocket Messages from Server:
{
  "type": "connected",
  "data": {
    "user_id": 1,
    "connected_at": "2026-03-13T20:00:00Z"
  }
}

{
  "type": "log:new",
  "data": {
    "id": 12345,
    "instance_id": 1,
    "timestamp": "2026-03-13T20:00:00Z",
    "level": "ERROR",
    "message": "Connection timeout"
  }
}

{
  "type": "metric:update",
  "data": {
    "instance_id": 1,
    "metric_type": "connections",
    "value": 42,
    "timestamp": "2026-03-13T20:00:00Z"
  }
}

{
  "type": "alert:triggered",
  "data": {
    "alert_id": 5,
    "instance_id": 1,
    "alert_name": "High error rate",
    "triggered_at": "2026-03-13T20:00:00Z"
  }
}

{
  "type": "error",
  "data": {
    "message": "JWT token expired",
    "code": "TOKEN_EXPIRED"
  }
}
```

**Implementation:**
- File: `backend/internal/api/handlers_realtime.go`
- Function: `handleWebSocket`
- JWT validation on connection
- Maintains user-to-connections mapping
- Filters logs/events by user's instance permissions
- Heartbeat mechanism (ping/pong every 30 seconds)

### Go Services

#### WebSocket Manager (`pkg/services/websocket.go`)

```go
type WebSocketManager struct {
  connections map[int]*Connection  // user_id -> connection
  broadcast   chan BroadcastMessage
  register    chan *Connection
  unregister  chan *Connection
  mu          sync.RWMutex
}

type Connection struct {
  UserID      int
  InstanceIDs []int
  WebSocket   *websocket.Conn
  Send        chan interface{}
}

type BroadcastMessage struct {
  Type       string      // "log:new", "alert:triggered", etc
  Data       interface{}
  InstanceID int         // Only send to users with access to this instance
  UserID     int         // Or only to specific user (0 = all)
}
```

**Key Methods:**
- `NewWebSocketManager()`: Create manager instance
- `Register(conn *Connection)`: Add new WebSocket connection
- `Unregister(conn *Connection)`: Remove closed connection
- `Broadcast(msg BroadcastMessage)`: Send event to users
- `Run()`: Start goroutine to process messages
- `Start()`: Initialize listeners and workers

**Features:**
- Thread-safe connection management
- Per-user instance filtering
- Efficient broadcast to multiple users
- Graceful shutdown support

#### Alert Worker (`pkg/services/alert_worker.go`)

```go
type AlertWorker struct {
  db         *sql.DB
  wsManager  *WebSocketManager
  ticker     *time.Ticker
  done       chan struct{}
  logger     *zap.Logger
}
```

**Execution Flow (every 60 seconds):**
1. Query all active alert rules from `alert_rules` table
2. For each rule:
   - Get relevant data (logs, metrics) for evaluation window
   - Evaluate condition (e.g., "error_count > 5 in 1 minute")
   - Check deduplication cache (5-minute window)
   - If condition matches AND not deduplicated:
     - Create `alert_trigger` record
     - Broadcast `alert:triggered` event to WebSocket
     - Create `notification` records for each channel
3. Update metrics (alerts_evaluated, alerts_triggered)

**Deduplication Logic:**
```go
// Don't trigger the same alert+instance more than once per 5 minutes
SELECT COUNT(*) FROM alert_triggers
WHERE alert_id = ? AND instance_id = ?
AND triggered_at > NOW() - INTERVAL '5 minutes'
```

**Error Handling:**
- If evaluation fails, log error but continue with next alert
- If broadcast fails, notification will still be created
- Worker continues running despite individual failures

#### Notification Worker (`pkg/services/notification_worker.go`)

```go
type NotificationWorker struct {
  db        *sql.DB
  channels  map[string]Channel
  ticker    *time.Ticker
  done      chan struct{}
  logger    *zap.Logger
}

type Channel interface {
  Send(ctx context.Context, notification *Notification) error
  Name() string
}
```

**Execution Flow (every 5 seconds):**
1. Query pending notifications from `notifications` table
2. For each notification:
   - Load channel configuration
   - Load alert details (name, description, etc)
   - Load alert trigger context
   - Call appropriate channel's Send() method:
     - EmailChannel: Send SMTP email
     - SlackChannel: POST to webhook
     - PagerDutyChannel: Create/resolve incident
     - WebhookChannel: POST to custom URL
   - If send successful:
     - Update status = 'delivered'
     - Set sent_at = NOW()
   - If send failed:
     - Increment retry_count
     - If retry_count < 3:
       - Keep status = 'pending' (will retry)
     - Else:
       - Update status = 'failed'
       - Log failure
3. Update metrics (notifications_sent, notifications_failed)

**Retry Logic:**
```
Attempt 1 (immediate): Send notification
Attempt 2 (5s later):  Retry if first failed
Attempt 3 (10s later): Retry if second failed
After 3 failures:      Mark as 'failed', stop retrying
```

**Supported Channels:**
- Email: SMTP-based email notifications
- Slack: Webhook URL for message posting
- PagerDuty: Events API for incident management
- Webhook: Generic POST to custom URL

### Go Models (`pkg/models/models.go`)

```go
// AlertTrigger represents when an alert rule was triggered
type AlertTrigger struct {
  ID          int64     `db:"id" json:"id"`
  AlertID     int64     `db:"alert_id" json:"alert_id"`
  InstanceID  int       `db:"instance_id" json:"instance_id"`
  TriggeredAt time.Time `db:"triggered_at" json:"triggered_at"`
  CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// Notification represents a pending/sent alert notification
type Notification struct {
  ID             int64      `db:"id" json:"id"`
  ChannelID      int64      `db:"channel_id" json:"channel_id"`
  AlertTriggerID int64      `db:"alert_trigger_id" json:"alert_trigger_id"`
  Status         string     `db:"status" json:"status"` // pending, delivered, failed
  RetryCount     int        `db:"retry_count" json:"retry_count"`
  LastRetryAt    *time.Time `db:"last_retry_at" json:"last_retry_at"`
  SentAt         *time.Time `db:"sent_at" json:"sent_at"`
  CreatedAt      time.Time  `db:"created_at" json:"created_at"`
  UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}
```

---

## Frontend Implementation

### RealtimeClient Service (`src/services/realtime.ts`)

```typescript
export class RealtimeClient {
  private ws: WebSocket | null = null
  private url: string
  private token: string = ''
  private listeners: Map<string, Set<EventListener>> = new Map()
  private messageQueue: Message[] = []
  private reconnectAttempts: number = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private maxReconnectAttempts: number = 5

  constructor(baseURL: string)
  async connect(token: string): Promise<void>
  disconnect(): void
  on(eventType: string, listener: EventListener): void
  off(eventType: string, listener: EventListener): void
  private attemptConnect(): Promise<void>
  private attemptReconnect(): void
  private handleMessage(event: MessageEvent): void
  private setupHeartbeat(): void
  private flushMessageQueue(): void
  private emitError(error: Error): void
}
```

**Key Features:**
- Auto-reconnect with exponential backoff: 1s → 2s → 4s → 8s → 30s
- Message queuing when offline (prevents message loss during disconnects)
- Event listener pattern (pub/sub)
- JWT token handling
- Heartbeat/ping mechanism for connection validation
- Automatic URL protocol conversion (http→ws, https→wss)

**Usage:**
```typescript
const client = new RealtimeClient('http://localhost:8000/api/v1/ws')

// Connect with JWT token
await client.connect(jwtToken)

// Listen to events
client.on('log:new', (logData) => {
  console.log('New log:', logData)
})

client.on('alert:triggered', (alertData) => {
  console.log('Alert fired:', alertData)
})

client.on('error', (error) => {
  console.error('Connection error:', error)
})

// Disconnect
client.disconnect()
```

### Zustand Store (`src/stores/realtimeStore.ts`)

```typescript
interface RealtimeState {
  connected: boolean
  lastUpdate: Date | null
  error: string | null
  listeners: Map<string, Set<(data: any) => void>>

  // Actions
  setConnected: (connected: boolean) => void
  setLastUpdate: (date: Date) => void
  setError: (error: string | null) => void
  subscribe: (eventType: string, listener: (data: any) => void) => void
  unsubscribe: (eventType: string, listener: (data: any) => void) => void
  emit: (eventType: string, data: any) => void
}

export const useRealtimeStore = create<RealtimeState>((set, get) => ({
  // initial state
  connected: false,
  lastUpdate: null,
  error: null,
  listeners: new Map(),

  // actions
  setConnected: (connected) => set({ connected }),
  setLastUpdate: (date) => set({ lastUpdate: date }),
  setError: (error) => set({ error }),
  // ... etc
}))
```

**State Management:**
- `connected`: Boolean flag for WebSocket connection state
- `lastUpdate`: Timestamp of last received event
- `error`: Last error message (if any)
- `listeners`: Map of event subscriptions
- Methods: `setConnected`, `setError`, `subscribe`, `emit`

**Usage:**
```typescript
const { connected, error, subscribe, unsubscribe } = useRealtimeStore()

// Subscribe to log events
const handleNewLog = (log) => { /* ... */ }
subscribe('log:new', handleNewLog)

// Later: unsubscribe
unsubscribe('log:new', handleNewLog)
```

### useRealtime Hook (`src/hooks/useRealtime.ts`)

```typescript
interface UseRealtimeReturn {
  connected: boolean
  lastUpdate: Date | null
  error: string | null
  subscribe: (eventType: string, listener: (data: any) => void) => void
  unsubscribe: (eventType: string, listener: (data: any) => void) => void
}

export const useRealtime = (): UseRealtimeReturn => {
  const store = useRealtimeStore()

  return {
    connected: store.connected,
    lastUpdate: store.lastUpdate,
    error: store.error,
    subscribe: useCallback(store.subscribe, []),
    unsubscribe: useCallback(store.unsubscribe, []),
  }
}
```

**Features:**
- Memoized methods to prevent unnecessary re-renders
- Simple, clean API for React components
- Wrapper around Zustand store

**Usage in Components:**
```typescript
function LogViewer() {
  const { connected, error, subscribe } = useRealtime()

  useEffect(() => {
    const handleNewLog = (log) => {
      console.log('Got log:', log)
    }

    subscribe('log:new', handleNewLog)
    return () => unsubscribe('log:new', handleNewLog)
  }, [subscribe, unsubscribe])

  return (
    <div>
      Status: {connected ? 'Live' : 'Offline'}
      {error && <Alert>{error}</Alert>}
    </div>
  )
}
```

### LiveLogsStream Component (`src/components/logs/LiveLogsStream.tsx`)

```typescript
interface LiveLogsStreamProps {
  instanceId: number
  maxLogs?: number  // default: 50
  autoScroll?: boolean  // default: true
}

export function LiveLogsStream({
  instanceId,
  maxLogs = 50,
  autoScroll = true
}: LiveLogsStreamProps) {
  const [logs, setLogs] = useState<Log[]>([])
  const [paused, setPaused] = useState(false)
  const { subscribe } = useRealtime()

  // Subscribe to log:new events
  useEffect(() => {
    const handleNewLog = (log: Log) => {
      if (!paused) {
        setLogs(prev => [log, ...prev].slice(0, maxLogs))
      }
    }
    subscribe('log:new', handleNewLog)
    return () => unsubscribe('log:new', handleNewLog)
  }, [paused, maxLogs])

  return (
    <div>
      <div className="header">
        <h2>Live Logs</h2>
        <button onClick={() => setPaused(!paused)}>
          {paused ? 'Resume' : 'Pause'}
        </button>
      </div>

      <div className="logs-container" ref={endRef}>
        {logs.map(log => (
          <LogEntry key={log.id} log={log} />
        ))}
      </div>
    </div>
  )
}
```

**Features:**
- Displays last 50 logs in real-time
- Auto-scroll with pause/resume functionality
- Color-coded by level (ERROR=red, WARNING=yellow, INFO=blue)
- Expandable log details
- Efficient log buffer (prevents memory bloat)

**Log Levels & Colors:**
- ERROR: Red (#dc2626)
- WARNING: Yellow (#d97706)
- INFO: Blue (#3b82f6)
- DEBUG: Gray (#6b7280)

### RealtimeStatus Component (`src/components/common/RealtimeStatus.tsx`)

```typescript
interface RealtimeStatusProps {
  showTimestamp?: boolean  // default: false
  compact?: boolean  // default: false
}

export function RealtimeStatus({
  showTimestamp = false,
  compact = false
}: RealtimeStatusProps) {
  const { connected, lastUpdate, error } = useRealtime()

  const statusColor = connected ? 'green' : 'yellow'
  const statusText = connected ? 'Live' : 'Polling'
  const statusIcon = connected ? '●' : '○'

  return (
    <div className={`realtime-status ${statusColor}`}>
      <span className="icon">{statusIcon}</span>
      <span className="text">{statusText}</span>

      {showTimestamp && lastUpdate && (
        <span className="timestamp">
          {formatTime(lastUpdate)}
        </span>
      )}

      {error && (
        <span className="error" title={error}>⚠</span>
      )}
    </div>
  )
}
```

**Display States:**
- Connected: Green dot, "Live" text, pulsing animation
- Disconnected: Yellow dot, "Polling" text, static
- Error: Red alert icon, error message in tooltip

**CSS Styling:**
```css
.realtime-status.green {
  animation: pulse 2s infinite;
  border-color: #10b981;
  color: #10b981;
}

.realtime-status.yellow {
  border-color: #f59e0b;
  color: #f59e0b;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```

### App Initialization (`src/App.tsx`)

```typescript
export function App() {
  const { user, isAuthenticated } = useAuth()
  const {
    setConnected,
    setError: setRealtimeError,
    setLastUpdate,
    emit
  } = useRealtimeStore()

  // Initialize RealtimeClient when authenticated
  useEffect(() => {
    if (!isAuthenticated || !user) {
      // User logged out
      realtimeClient.disconnect()
      setRealtimeError(null)
      return
    }

    // Connect WebSocket with JWT token
    realtimeClient
      .connect(user.jwt_token)
      .then(() => {
        // Set up event listeners
        realtimeClient.on('log:new', (log) => {
          setLastUpdate(new Date())
          emit('log:new', log)
        })

        realtimeClient.on('alert:triggered', (alert) => {
          setLastUpdate(new Date())
          emit('alert:triggered', alert)
          showNotification(`Alert: ${alert.alert_name}`)
        })

        realtimeClient.on('metric:update', (metric) => {
          setLastUpdate(new Date())
          emit('metric:update', metric)
        })

        realtimeClient.on('connected', () => {
          setConnected(true)
          setRealtimeError(null)
        })

        realtimeClient.on('disconnected', () => {
          setConnected(false)
        })

        realtimeClient.on('error', (error) => {
          const errorMessage = error?.message || 'Connection error'
          setRealtimeError(errorMessage)
          console.error('RealtimeClient error:', error)
        })
      })
      .catch((error) => {
        const errorMessage = error?.message || 'Failed to connect to realtime service'
        setRealtimeError(errorMessage)
        console.error('Failed to initialize RealtimeClient:', error)
      })

    // Cleanup on unmount or logout
    return () => {
      realtimeClient.disconnect()
    }
  }, [isAuthenticated, user])

  // ... rest of component
}
```

---

## Configuration

### Environment Variables

**Backend (.env or .env.production)**
```bash
# Database
DATABASE_URL=postgres://pganalytics:password@localhost:5432/pganalytics
TIMESCALE_URL=postgres://pganalytics:password@localhost:5432/pganalytics

# Server
API_PORT=8000
ENVIRONMENT=production

# JWT
JWT_SECRET=your-very-secret-key-change-this
JWT_EXPIRY=24h

# WebSocket
WEBSOCKET_PORT=8000
WEBSOCKET_READ_BUFFER_SIZE=1024
WEBSOCKET_WRITE_BUFFER_SIZE=1024

# Log Retention
LOG_RETENTION_DAYS=30

# Cache
CACHE_ENABLED=true
CACHE_MAX_SIZE=10000
CACHE_TTL=3600

# Workers
ALERT_WORKER_INTERVAL=60s
NOTIFICATION_WORKER_INTERVAL=5s
```

**Frontend (.env.production or .env.local)**
```bash
# API Configuration
VITE_API_URL=https://api.pganalytics.example.com
VITE_API_TIMEOUT=10000
VITE_WEBSOCKET_URL=wss://api.pganalytics.example.com/api/v1/ws

# Feature Flags
VITE_ENABLE_REALTIME=true
VITE_ENABLE_ALERTS=true

# Monitoring
VITE_SENTRY_DSN=https://key@sentry.example.com/project
```

### Docker Compose Configuration

```yaml
services:
  pganalytics-api:
    image: pganalytics-api:latest
    ports:
      - "8000:8000"
    environment:
      DATABASE_URL: postgres://pganalytics:password@postgres:5432/pganalytics
      TIMESCALE_URL: postgres://pganalytics:password@timescale:5432/pganalytics
      JWT_SECRET: ${JWT_SECRET:-change-me}
      ALERT_WORKER_INTERVAL: 60s
      NOTIFICATION_WORKER_INTERVAL: 5s
    depends_on:
      - postgres
      - timescale
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/api/v1/health"]
      interval: 10s
      timeout: 3s
      retries: 3

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: pganalytics
      POSTGRES_PASSWORD: password
    volumes:
      - postgres-data:/var/lib/postgresql/data

  timescale:
    image: timescale/timescaledb:latest-pg15
    environment:
      POSTGRES_DB: pganalytics
      POSTGRES_PASSWORD: password
    volumes:
      - timescale-data:/var/lib/postgresql/data

volumes:
  postgres-data:
  timescale-data:
```

---

## Testing

### Manual Testing Checklist

#### 1. Database Setup
- [ ] PostgreSQL running with pganalytics database
- [ ] Migration 022_realtime_tables.sql applied successfully
- [ ] alert_triggers table exists: `psql -c "\dt alert_triggers"`
- [ ] notifications table exists: `psql -c "\dt notifications"`
- [ ] Indexes created: `psql -c "\di" | grep alert_triggers`

#### 2. Backend Service
- [ ] Backend API running on port 8000
- [ ] Health check responds: `curl http://localhost:8000/api/v1/health`
- [ ] Database connection working
- [ ] WebSocket endpoint available: `curl -i http://localhost:8000/api/v1/ws`
- [ ] Alert worker running (check logs for "Alert worker started")
- [ ] Notification worker running (check logs for "Notification worker started")

#### 3. Frontend Build
- [ ] Frontend builds without errors: `npm run build`
- [ ] No TypeScript errors: `npm run type-check`
- [ ] RealtimeClient can be imported: `import { RealtimeClient } from './services/realtime'`
- [ ] Zustand store exists: `src/stores/realtimeStore.ts`
- [ ] useRealtime hook exists: `src/hooks/useRealtime.ts`
- [ ] Components exist: LiveLogsStream, RealtimeStatus

#### 4. Log Ingestion Testing

**Test with curl:**
```bash
# Get API token for collector
API_TOKEN="your-collector-api-key"
INSTANCE_ID=1

curl -X POST http://localhost:8000/api/v1/logs/ingest \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collector_id": "test-collector",
    "instance_id": '$INSTANCE_ID',
    "logs": [
      {
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "level": "ERROR",
        "message": "Test error message"
      }
    ]
  }'

# Response should be:
# {"success": true, "ingested": 1}
```

**Check database:**
```bash
# Verify logs stored
psql -d pganalytics -c "SELECT * FROM postgresql_logs ORDER BY timestamp DESC LIMIT 5;"
```

- [ ] Logs successfully POST to /api/v1/logs/ingest
- [ ] HTTP 200 response with {"success": true, "ingested": N}
- [ ] Logs appear in postgresql_logs table
- [ ] Timestamps are correct

#### 5. WebSocket Connection Testing

**Test with WebSocket client:**
```bash
# Using websocat or similar tool
websocat -v ws://localhost:8000/api/v1/ws?token=YOUR_JWT_TOKEN

# You should see connection messages:
# --- sending stdin line to server
# {"type":"connected",...}
# --- got message from server
```

**Check frontend:**
- [ ] Open browser console (F12)
- [ ] Login to frontend application
- [ ] Check console: Should see WebSocket connection established
- [ ] RealtimeStatus component shows "●Live" (green dot)
- [ ] No errors in console about connection failures

- [ ] Frontend connects to /api/v1/ws with JWT token
- [ ] Connection shows as "Live" status
- [ ] Connection persists without errors

#### 6. Real-Time Log Display Testing

**Procedure:**
1. Login to frontend
2. Verify WebSocket connected (green status)
3. In another terminal, send test log:
   ```bash
   curl -X POST http://localhost:8000/api/v1/logs/ingest \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{...log data...}'
   ```
4. Watch frontend for new log appearance

- [ ] Log appears in LiveLogsStream within 100ms
- [ ] Log displays with correct level color
- [ ] Timestamp is accurate
- [ ] Message content displays correctly
- [ ] Multiple logs appear in correct order (newest first)
- [ ] Pause button works (stops new logs)
- [ ] Resume button works (shows new logs again)

#### 7. Alert Evaluation Testing

**Setup test alert:**
```bash
# Use database or API to create an alert rule
psql -d pganalytics -c "
INSERT INTO alert_rules (instance_id, name, condition, threshold, active)
VALUES (1, 'High Error Rate', 'error_count > 2', 2, true);
"

# Get alert_id from result
```

**Procedure:**
1. Create alert rule via UI or database
2. Send 3+ ERROR logs within 60 seconds
3. Wait for alert worker to run (60s)
4. Check database for alert_trigger record

```bash
# Check alert_triggers table
psql -d pganalytics -c "SELECT * FROM alert_triggers ORDER BY created_at DESC LIMIT 5;"
```

- [ ] Alert rule active in database
- [ ] After 60 seconds, alert_trigger created
- [ ] Only one trigger per alert per day (deduplication working)
- [ ] Frontend receives "alert:triggered" event
- [ ] Alert notification displayed to user

#### 8. Notification Testing

**Email Channel:**
```bash
# Setup test email channel
psql -d pganalytics -c "
INSERT INTO notification_channels (alert_id, channel_type, config)
VALUES (
  1,
  'email',
  '{\"recipients\": [\"test@example.com\"]}'
);
"

# Trigger alert (send ERROR logs)
# Wait 5 seconds for notification worker to run
# Check email inbox or logs for SMTP debug
```

- [ ] notification records created with status='pending'
- [ ] After 5 seconds, status changes to 'delivered'
- [ ] Email sent to recipients
- [ ] Retry count = 0 for successful sends
- [ ] Failed notifications get retried (up to 3 times)

**Slack Channel:**
```bash
# Setup test Slack channel
psql -d pganalytics -c "
INSERT INTO notification_channels (alert_id, channel_type, config)
VALUES (
  1,
  'slack',
  '{\"webhook_url\": \"https://hooks.slack.com/services/YOUR/HOOK/URL\"}'
);
"

# Trigger alert
# Check Slack channel for message
```

- [ ] Slack message posted when alert triggered
- [ ] Message includes alert name, instance, timestamp
- [ ] Message formatted as Slack message (with formatting)

#### 9. Connection Resilience Testing

**Disconnect/Reconnect:**
1. Open frontend with WebSocket connected
2. In browser DevTools (F12):
   - Application → Storage → Cookies
   - Or: Network tab → filter to WebSocket
3. Simulate network issues:
   ```bash
   # Option 1: Turn off WiFi
   # Option 2: Use DevTools to throttle connection
   # Option 3: Stop backend and restart
   ```
4. Watch frontend:
   - Status should change to "Offline" (yellow)
   - After ~4 seconds: Status back to "Live" (green)
   - No error messages in console

- [ ] WebSocket disconnect detected automatically
- [ ] Status badge shows "Offline" state
- [ ] Automatic reconnect with exponential backoff
- [ ] Connection restored after 4-8 seconds
- [ ] No logs/data lost during reconnect
- [ ] Error messages clear when reconnected

#### 10. Error Handling Testing

**Invalid JWT Token:**
```bash
websocat -v ws://localhost:8000/api/v1/ws?token=invalid-token

# Should get error message
```

- [ ] Invalid token rejected with error
- [ ] Frontend shows error message
- [ ] Connection closes gracefully
- [ ] Frontend shows "Offline" status

**Database Unavailable:**
- [ ] Stop PostgreSQL service
- [ ] Send log to /api/v1/logs/ingest
- [ ] Receive 500 error
- [ ] Backend logs error
- [ ] Restart PostgreSQL
- [ ] Log ingestion works again

- [ ] Error response when database unavailable
- [ ] Graceful error handling
- [ ] Service recovery after database restart

### Integration Test Scenarios

#### Scenario 1: Real-Time Log Display (20 minutes)

**Setup:**
```bash
# Terminal 1: Start backend
cd backend && go run cmd/pganalytics-api/main.go

# Terminal 2: Start frontend
cd frontend && npm run dev

# Terminal 3: Collector simulation
```

**Steps:**
1. Login to frontend at http://localhost:5173
2. Navigate to Instances → select an instance → Logs
3. Verify LiveLogsStream visible and RealtimeStatus shows "Live"
4. Send test log:
   ```bash
   curl -X POST http://localhost:8000/api/v1/logs/ingest \
     -H "Authorization: Bearer $COLLECTOR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "collector_id": "test",
       "instance_id": 1,
       "logs": [{
         "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
         "level": "ERROR",
         "message": "Test error log"
       }]
     }'
   ```
5. Observe new log in frontend within 100ms
6. Repeat with 10 logs in rapid succession
7. Verify all logs appear without loss

**Verification:**
- [ ] Logs appear in LiveLogsStream
- [ ] Latency < 100ms from ingestion to display
- [ ] Colors match log levels
- [ ] No console errors
- [ ] Timestamp ordering correct

#### Scenario 2: Alert Notification Flow (30 minutes)

**Setup:**
1. Create alert rule via UI or database:
   ```sql
   INSERT INTO alert_rules (instance_id, name, condition, threshold, active)
   VALUES (1, 'Test Alert', 'error_count > 1', 1, true);
   ```
2. Create email notification channel:
   ```sql
   INSERT INTO notification_channels (alert_id, channel_type, config)
   VALUES (1, 'email', '{"recipients": ["test@example.com"]}');
   ```

**Steps:**
1. Verify alert rule is active
2. Send 2+ ERROR logs in quick succession
3. Wait 60 seconds for alert worker to run
4. Check database for alert_trigger:
   ```bash
   psql -d pganalytics -c "SELECT * FROM alert_triggers ORDER BY created_at DESC LIMIT 1;"
   ```
5. Check for notification records:
   ```bash
   psql -d pganalytics -c "SELECT * FROM notifications ORDER BY created_at DESC LIMIT 1;"
   ```
6. Wait 5 seconds for notification worker
7. Verify email sent (check logs or mailbox)
8. Check notification status = 'delivered':
   ```bash
   psql -d pganalytics -c "SELECT status, sent_at FROM notifications WHERE id = <id>;"
   ```

**Verification:**
- [ ] alert_trigger created when condition matched
- [ ] notification record created
- [ ] Email sent within 10 seconds of condition match
- [ ] notification status = 'delivered'
- [ ] Frontend received "alert:triggered" event
- [ ] Toast notification shown to user

#### Scenario 3: Connection Resilience Under Load (15 minutes)

**Setup:**
```bash
# Terminal 1: Start backend
cd backend && go run cmd/pganalytics-api/main.go

# Terminal 2: Start frontend
cd frontend && npm run dev

# Terminal 3: Load test script
```

**Load Test Script:**
```bash
#!/bin/bash
# Send 100 logs over 60 seconds (variable rate)

COLLECTOR_TOKEN="your-token"
INSTANCE_ID=1

for i in {1..100}; do
  curl -s -X POST http://localhost:8000/api/v1/logs/ingest \
    -H "Authorization: Bearer $COLLECTOR_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "collector_id": "test",
      "instance_id": '$INSTANCE_ID',
      "logs": [{
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "level": "'$(echo ERROR WARNING INFO | shuf -n 1)'",
        "message": "Log message '$i'"
      }]
    }' > /dev/null

  # Variable delay (0-500ms)
  sleep 0.$(($RANDOM % 5))
done

echo "Sent 100 logs"
```

**Steps:**
1. Open frontend with DevTools (F12) → Network tab
2. Filter to WebSocket connections
3. Run load test script
4. Observe:
   - WebSocket connection remains stable
   - Logs appear in LiveLogsStream as sent
   - No dropped connections
   - CPU/memory reasonable
5. While load testing, simulate network issue:
   ```bash
   # Use macOS 'ifconfig' or Linux 'ip' to disable network
   # Or use DevTools to throttle to "Slow 3G"
   ```
6. Observe reconnection during load
7. After resuming network, logs continue flowing

**Verification:**
- [ ] 100 logs all appear in frontend
- [ ] No WebSocket connection drops
- [ ] No errors in backend logs
- [ ] Graceful handling of high volume
- [ ] Reconnection works under load
- [ ] No memory leaks (memory usage stable)

---

## Troubleshooting

### Frontend Issues

#### Problem: "CONNECTING..." badge stuck indefinitely

**Symptoms:**
- RealtimeStatus shows "Polling" (yellow)
- No logs appearing in real-time
- Console shows no errors

**Diagnosis:**
```bash
# Check if backend is running
lsof -i :8000

# Check WebSocket endpoint
curl -i http://localhost:8000/api/v1/ws

# Check browser console for WebSocket errors
# F12 → Console → filter to errors
```

**Solutions:**
1. Verify backend is running:
   ```bash
   cd backend && go run cmd/pganalytics-api/main.go
   ```

2. Check JWT token validity:
   - Tokens expire after 24 hours
   - Try logging out and logging back in
   - Check localStorage: DevTools → Application → LocalStorage → token

3. Verify firewall allows WebSocket:
   ```bash
   # Test raw WebSocket connection
   websocat -v ws://localhost:8000/api/v1/ws?token=YOUR_TOKEN
   ```

4. Check browser console (F12):
   ```javascript
   // Test WebSocket manually
   const ws = new WebSocket('ws://localhost:8000/api/v1/ws?token=' + yourToken)
   ws.onopen = () => console.log('Connected')
   ws.onerror = (e) => console.error('Error:', e)
   ws.onmessage = (e) => console.log('Message:', e.data)
   ```

#### Problem: Logs not appearing in real-time

**Symptoms:**
- LiveLogsStream empty
- RealtimeStatus shows "Live"
- But logs sent via curl appear in database

**Diagnosis:**
```bash
# Verify logs are in database
psql -d pganalytics -c "SELECT COUNT(*) FROM postgresql_logs;"

# Check if WebSocket handler is broadcasting
# Look in backend logs for "Broadcasting" messages

# Check frontend subscription
# In browser console:
useRealtimeStore.getState().listeners
```

**Solutions:**
1. Verify collector sending logs:
   ```bash
   curl -X POST http://localhost:8000/api/v1/logs/ingest \
     -H "Authorization: Bearer $COLLECTOR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "collector_id": "test",
       "instance_id": 1,
       "logs": [{
         "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
         "level": "ERROR",
         "message": "Test"
       }]
     }'
   ```

2. Check backend logs for errors:
   ```bash
   # Look for:
   # - "Failed to broadcast log"
   # - Database connection errors
   # - WebSocket write errors
   ```

3. Verify instance_id matches:
   - Collector sends logs for instance_id=1
   - Frontend user has access to instance_id=1
   - User permissions in database: `SELECT * FROM user_instances WHERE user_id = ? AND instance_id = 1`

4. Clear browser cache:
   - DevTools → Application → Clear all data
   - Refresh page
   - Try again

#### Problem: Error "Cannot find module './realtime'"

**Symptoms:**
- Build error: Module not found
- Runtime error in browser console

**Solutions:**
1. Verify file exists:
   ```bash
   ls -la frontend/src/services/realtime.ts
   ls -la frontend/src/stores/realtimeStore.ts
   ls -la frontend/src/hooks/useRealtime.ts
   ```

2. Clear cache and reinstall:
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm ci
   npm run build
   ```

3. Check TypeScript config:
   - File: `frontend/tsconfig.json`
   - Should have `"baseUrl": "."`
   - Should include `"src"`

4. Check import paths:
   - Should be relative: `import { RealtimeClient } from '../services/realtime'`
   - Not absolute without baseUrl

#### Problem: High CPU or memory usage

**Symptoms:**
- Browser tab using 100% CPU
- Memory usage growing continuously
- Frontend becomes sluggish

**Diagnosis:**
```javascript
// In browser console:
// Check listeners
useRealtimeStore.getState().listeners.size

// Check log buffer size
// In LiveLogsStream: logs.length
```

**Solutions:**
1. Check for message loop:
   - Verify listener unsubscribe on component unmount
   - Use DevTools Profiler to find re-renders

2. Limit log buffer:
   - LiveLogsStream has maxLogs=50 by default
   - Reduce if needed: `<LiveLogsStream maxLogs={20} />`

3. Check WebSocket message rate:
   - DevTools → Network → WS → Messages
   - Should be 1-2 per second max
   - If higher, check alert worker/notification worker frequency

4. Disable features if needed:
   - Set `VITE_ENABLE_REALTIME=false` to test
   - Disable specific components one by one

### Backend Issues

#### Problem: WebSocket connection fails with 401 Unauthorized

**Symptoms:**
- WebSocket handshake fails
- HTTP 401 response
- Backend logs: "Invalid token"

**Solutions:**
1. Verify JWT token format:
   ```bash
   # Should be: Authorization: Bearer <token>
   # Check localStorage for token
   ```

2. Check JWT secret matches:
   - Backend: env var `JWT_SECRET`
   - Token generated during login should use same secret
   - If changed, all existing tokens are invalid

3. Check token expiration:
   ```bash
   # Decode JWT (without verification):
   # Copy token to jwt.io
   # Check "exp" claim is in future
   ```

4. Verify auth middleware:
   - Handler: `backend/internal/api/handlers_realtime.go`
   - Check JWT extraction and validation

#### Problem: Logs not persisted to database

**Symptoms:**
- curl POST returns 200 success
- But logs don't appear in postgresql_logs table
- No errors in logs

**Diagnosis:**
```bash
# Check database connection
psql -d pganalytics -c "SELECT NOW();"

# Check tables exist
psql -d pganalytics -c "\dt postgresql_logs"

# Check disk space
df -h /var/lib/postgresql

# Check PostgreSQL logs
tail -f /var/log/postgresql/postgresql.log
```

**Solutions:**
1. Verify database connection:
   - Check DATABASE_URL env var
   - Test connection: `psql $DATABASE_URL -c "SELECT 1"`

2. Verify migration ran:
   ```bash
   # Check if table exists
   psql -d pganalytics -c "SELECT * FROM information_schema.tables WHERE table_name='postgresql_logs';"

   # If not, run migration
   cd backend
   ./pganalytics-api migrate
   ```

3. Check INSERT permissions:
   ```bash
   # Verify user can write to table
   psql -d pganalytics -U pganalytics -c "INSERT INTO postgresql_logs (instance_id, timestamp, level, message) VALUES (1, NOW(), 'TEST', 'test');"
   ```

4. Check disk space:
   - If full, delete old logs
   - Or extend partition

#### Problem: Alerts not firing

**Symptoms:**
- alert_triggers table empty
- alert_rules exist
- Error logs present

**Solutions:**
1. Verify alert worker is running:
   - Backend logs should show "Alert worker started"
   - Check: `ps aux | grep pganalytics-api`

2. Check alert rule condition:
   ```bash
   # Verify rule exists and is active
   psql -d pganalytics -c "SELECT * FROM alert_rules WHERE active = true;"

   # Check last evaluation:
   psql -d pganalytics -c "SELECT * FROM alert_rules WHERE id = 1 \gx"
   ```

3. Check if condition matches:
   - Manually query logs for the condition
   - Example: Rule condition is "error_count > 5 in 1 minute"
   ```bash
   # Count errors in last 60 seconds
   psql -d pganalytics -c "
   SELECT COUNT(*) FROM postgresql_logs
   WHERE level = 'ERROR'
   AND timestamp > NOW() - INTERVAL '1 minute'
   AND instance_id = 1;"
   ```

4. Check deduplication window:
   ```bash
   # Verify alert doesn't already exist in 5-min window
   psql -d pganalytics -c "
   SELECT * FROM alert_triggers
   WHERE alert_id = 1
   AND triggered_at > NOW() - INTERVAL '5 minutes';"
   ```

5. Check backend logs for alert worker errors:
   - Look for "Alert evaluation failed"
   - Look for database errors

#### Problem: Notifications not sending

**Symptoms:**
- notification records created (status='pending')
- But status stays 'pending' forever
- No emails/Slack messages

**Solutions:**
1. Verify notification worker is running:
   - Backend logs should show "Notification worker started"
   - Check process: `ps aux | grep pganalytics-api`

2. Check notification channel config:
   ```bash
   # For email:
   psql -d pganalytics -c "
   SELECT * FROM notification_channels WHERE channel_type = 'email' \gx"

   # Verify config has recipients:
   # {"recipients": ["test@example.com"]}
   ```

3. Check SMTP/Slack config:
   - Email: verify SMTP server credentials
   - Slack: verify webhook URL is valid
   - Test manually:
     ```bash
     # Test Slack webhook
     curl -X POST -H 'Content-type: application/json' \
       --data '{"text":"Test"}' \
       https://hooks.slack.com/services/YOUR/HOOK/URL
     ```

4. Check backend logs for notification errors:
   - Look for "Failed to send notification"
   - Look for "SMTP error"
   - Look for "HTTP 400/401" responses

5. Verify retry logic:
   ```bash
   # Check notification record
   psql -d pganalytics -c "
   SELECT status, retry_count, last_retry_at
   FROM notifications
   ORDER BY created_at DESC LIMIT 1 \gx"
   ```

#### Problem: High memory usage on backend

**Symptoms:**
- Backend process using > 500MB RAM
- Memory growing over time
- Performance degradation

**Diagnosis:**
```bash
# Check goroutines
curl http://localhost:8000/api/v1/debug/pprof/goroutine

# Check memory profile
curl http://localhost:8000/api/v1/debug/pprof/heap > heap.prof

# Analyze
go tool pprof heap.prof
```

**Solutions:**
1. Check WebSocket connections not leaking:
   - Verify unregister() called on disconnect
   - Check broadcast channel not blocking

2. Reduce log retention:
   - Update LOG_RETENTION_DAYS env var
   - Delete old logs: `DELETE FROM postgresql_logs WHERE timestamp < NOW() - INTERVAL '7 days'`

3. Restart backend service:
   ```bash
   pkill -f pganalytics-api
   sleep 2
   go run cmd/pganalytics-api/main.go &
   ```

### Database Issues

#### Problem: Alert triggers not being deduplicated

**Symptoms:**
- Same alert fires multiple times in 5-minute window
- Multiple alert_trigger records for same alert+instance+date

**Diagnosis:**
```bash
# Check for duplicates
psql -d pganalytics -c "
SELECT alert_id, instance_id, DATE(triggered_at), COUNT(*)
FROM alert_triggers
GROUP BY alert_id, instance_id, DATE(triggered_at)
HAVING COUNT(*) > 1;"
```

**Solutions:**
1. Check deduplication logic in alert_worker.go:
   - Should check last 5 minutes before inserting
   - Unique constraint prevents database-level duplicates

2. Verify unique constraint exists:
   ```bash
   psql -d pganalytics -c "
   SELECT constraint_name
   FROM information_schema.table_constraints
   WHERE table_name = 'alert_triggers'
   AND constraint_type = 'UNIQUE';"
   ```

3. If constraint missing, recreate it:
   ```bash
   psql -d pganalytics -c "
   ALTER TABLE alert_triggers
   ADD CONSTRAINT alert_triggers_unique
   UNIQUE (alert_id, instance_id, DATE(triggered_at));"
   ```

---

## Performance Considerations

### Frontend Performance

**Log Buffer Management:**
- Keep last 50 logs in memory (configurable)
- Older logs discarded when buffer full
- Prevents memory bloat on long-running sessions

**WebSocket Message Size:**
- Average log event: ~500 bytes
- Average alert event: ~300 bytes
- Average metric event: ~200 bytes
- Total throughput at 1 event/sec: ~1KB/sec per connection

**Browser Resource Usage:**
- Background tab: ~50MB memory, <1% CPU
- Active tab: ~100MB memory, 1-5% CPU (depending on activity)

**Optimization Tips:**
1. Use `maxLogs={20}` if memory constrained
2. Implement virtual scrolling for large lists
3. Debounce high-frequency updates
4. Consider pagination instead of infinite scroll

### Backend Performance

**WebSocket Connection Limits:**
- Memory per connection: ~50KB
- Typical system: 10,000 concurrent connections = 500MB
- Goroutines per connection: ~3-4
- File descriptor limit important

**Database Query Performance:**
```
Alert evaluation per rule: ~10-50ms
Notification sending per channel: ~100-500ms
PostgreSQL log table query: <100ms with indexes
```

**Scaling Recommendations:**

| Metric | Scale | Recommendation |
|--------|-------|-----------------|
| Logs/sec | <100 | Single backend sufficient |
| Logs/sec | 100-1000 | Add read replica for logs |
| Logs/sec | >1000 | Separate log ingestion service |
| WebSocket connections | <1000 | Single backend sufficient |
| WebSocket connections | 1000-10000 | Load balance WebSocket |
| WebSocket connections | >10000 | Dedicated WebSocket cluster |
| Alert rules | <100 | Single alert worker |
| Alert rules | 100-1000 | Horizontal alert worker |
| Alert rules | >1000 | Distributed alert processor |

### Database Optimization

**Retention Policy:**
```sql
-- Auto-delete logs older than 30 days
SELECT cron.schedule('cleanup-old-logs', '0 2 * * *', $$
  DELETE FROM postgresql_logs
  WHERE timestamp < NOW() - INTERVAL '30 days';
$$);
```

**Index Maintenance:**
```sql
-- Regular VACUUM and ANALYZE
VACUUM ANALYZE postgresql_logs;
VACUUM ANALYZE alert_triggers;
VACUUM ANALYZE notifications;

-- Check index bloat
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read
FROM pg_stat_user_indexes
WHERE schemaname != 'pg_catalog'
ORDER BY idx_scan DESC;
```

---

## Security Considerations

### Authentication & Authorization

**JWT Tokens:**
- Issued on login, expire after 24 hours
- Sent in Authorization header for REST APIs
- Sent as query parameter for WebSocket
- Signed with HS256 algorithm

**Token Validation:**
```go
// Backend validates:
// 1. Signature matches JWT_SECRET
// 2. Token not expired (exp < now)
// 3. User exists and is not deleted
// 4. User has access to requested instances
```

**Best Practices:**
1. Store tokens in memory or secure storage (not localStorage if possible)
2. Implement token refresh flow
3. Rotate JWT_SECRET regularly (requires login)
4. Use HTTPS/WSS in production

### Instance-Level Access Control

**WebSocket Filtering:**
```go
// User can only see logs/alerts for their instances
// Backend query:
SELECT * FROM user_instances
WHERE user_id = ? AND instance_id = ?

// Only broadcast events for instances user can access
```

**API Token Security:**
- Collector API tokens separate from JWT
- Long-lived but rotatable
- Stored as hashed in database
- Rate-limited to prevent abuse

### Data Privacy

**Log Data Handling:**
- Logs stored in user's PostgreSQL instance
- No data sent to external services (unless configured)
- Retention policy deletes old logs automatically

**Sensitive Data in Logs:**
- Logs may contain query parameters with secrets
- Document policy: "Don't log sensitive data in PostgreSQL"
- Filter or mask sensitive patterns if possible

**Notification Channel Credentials:**
- Stored encrypted in database
- Never logged or exposed in API responses
- Only accessed by notification worker

### HTTPS/WSS in Production

**Configuration:**
```bash
# Backend
API_TLS_CERT=/path/to/cert.pem
API_TLS_KEY=/path/to/key.pem

# Frontend must use WSS (WebSocket Secure)
VITE_WEBSOCKET_URL=wss://api.pganalytics.example.com/api/v1/ws
```

**Testing:**
```bash
# Verify certificate is valid
openssl x509 -in cert.pem -text -noout

# Test HTTPS
curl -k https://localhost:8000/api/v1/health

# Test WSS
websocat -v wss://localhost:8000/api/v1/ws?token=...
```

---

## Files Summary

### Backend Files

| File | Purpose | Key Components |
|------|---------|-----------------|
| `backend/migrations/022_realtime_tables.sql` | Database schema | alert_triggers, notifications tables |
| `backend/pkg/models/models.go` | Go structs | AlertTrigger, Notification types |
| `backend/internal/api/handlers_realtime.go` | WebSocket handler | JWT auth, connection upgrade |
| `backend/pkg/handlers/logs.go` | Log ingest handler | API token validation, log parsing |
| `backend/pkg/services/websocket.go` | Connection manager | broadcast, register, unregister |
| `backend/pkg/services/alert_worker.go` | Alert evaluation | 60s ticker, condition evaluation |
| `backend/pkg/services/notification_worker.go` | Notification delivery | 5s ticker, multi-channel support |
| `backend/internal/api/server.go` | Route registration | /api/v1/ws, /api/v1/logs/ingest |

### Frontend Files

| File | Purpose | Key Components |
|------|---------|-----------------|
| `frontend/src/services/realtime.ts` | WebSocket client | connection mgmt, auto-reconnect |
| `frontend/src/stores/realtimeStore.ts` | Zustand store | event subscriptions, state |
| `frontend/src/hooks/useRealtime.ts` | React hook | component integration, memoized |
| `frontend/src/components/logs/LiveLogsStream.tsx` | Live logs display | real-time log rendering |
| `frontend/src/components/common/RealtimeStatus.tsx` | Status indicator | connection status badge |
| `frontend/src/App.tsx` | App initialization | RealtimeClient setup |

---

## Next Steps

### Short Term (Week 1)
1. **Testing**: Complete all manual testing scenarios
2. **Documentation**: Share this guide with team
3. **Monitoring**: Set up alerts for worker errors
   ```sql
   -- Monitor alert worker
   SELECT COUNT(*) FROM alert_triggers WHERE created_at > NOW() - INTERVAL '1 hour';

   -- Monitor notification delivery
   SELECT status, COUNT(*) FROM notifications
   GROUP BY status;
   ```

### Medium Term (Week 2-3)
1. **Load Testing**: Run with 1000+ concurrent users
   - WebSocket connection stability
   - Message throughput
   - Database query performance

2. **Performance Tuning**:
   - Optimize database indexes
   - Tune buffer sizes
   - Profile memory usage

3. **Reliability**:
   - Add circuit breakers for external services
   - Implement health checks
   - Add graceful shutdown

### Long Term (Month 2)
1. **Feature Enhancements**:
   - Custom alert conditions (UI)
   - Webhook integration
   - Alert silencing/snooze
   - Escalation policies

2. **Scaling**:
   - Separate alert worker to dedicated service
   - Separate notification worker
   - Redis for connection management
   - Kafka for event streaming

3. **Advanced**:
   - Machine learning for anomaly detection
   - Correlation across multiple alerts
   - Automated remediation
   - Integration with external monitoring systems

---

## Support & Debugging

### Getting Help

**When Something Breaks:**
1. Check this troubleshooting section
2. Review backend logs: `docker logs pganalytics-api`
3. Review frontend console: `F12 → Console tab`
4. Check database: `psql -d pganalytics -c "SELECT..."`
5. Verify configuration: `env | grep API_`

**Debug Mode:**
```bash
# Backend debug logging
export DEBUG=true
export LOG_LEVEL=debug
go run cmd/pganalytics-api/main.go

# Frontend debug logging
VITE_DEBUG=true npm run dev
```

**Useful Queries:**

```sql
-- Check alert status
SELECT a.*, COUNT(t.id) as trigger_count
FROM alert_rules a
LEFT JOIN alert_triggers t ON a.id = t.alert_id
GROUP BY a.id;

-- Check notification backlog
SELECT status, COUNT(*) as count
FROM notifications
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY status;

-- Check log volume
SELECT DATE(timestamp), COUNT(*) as log_count
FROM postgresql_logs
GROUP BY DATE(timestamp)
ORDER BY DATE DESC;

-- Check WebSocket connections (if available)
SELECT COUNT(*) FROM active_websocket_connections;
```

**Contact:**
- For issues: Review logs and this guide
- For feature requests: Create GitHub issue
- For emergency: Restart backend service

---

## Glossary

- **Alert Rule**: Configuration defining when to trigger an alert
- **Alert Trigger**: Record of an alert rule firing at a specific time
- **Notification**: Request to send alert to a channel (email, Slack, etc)
- **Channel**: Destination for notifications (email address, Slack webhook)
- **WebSocket**: Protocol for persistent, bidirectional connection
- **JWT**: JSON Web Token for authentication
- **Deduplication**: Preventing duplicate alerts within a time window
- **Broadcast**: Sending event to multiple connected users
- **Ticker**: Timer that runs code at regular intervals

---

## Change Log

- **v1.0.0** (2026-03-13): Initial Phase 3 implementation
  - Real-time log ingestion and streaming
  - Alert evaluation and triggering
  - Multi-channel notifications
  - WebSocket connection management
  - Automatic reconnection with exponential backoff
