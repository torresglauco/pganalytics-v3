# Managed Instance Health Check Scheduler

## Overview

Automatic health check scheduler for managed PostgreSQL instances that updates connection status periodically without requiring manual user interaction.

## Problem Solved

Previously, the managed instance connection status only updated when the user clicked the "Test" button. This required manual intervention for every check and didn't reflect real-time system status.

**Now:**
- ✅ Automatic periodic health checks (every 30 seconds)
- ✅ Randomized staggered execution (prevents backend overload)
- ✅ Limited concurrency (max 3 simultaneous checks)
- ✅ Real-time UI updates via API polling
- ✅ Automatic error/success status tracking

## Architecture

### Components

#### 1. Health Check Scheduler (`backend/internal/jobs/health_check_scheduler.go`)

Background service that manages periodic health checks:

```go
// Creates a new scheduler
scheduler := jobs.NewHealthCheckScheduler(postgresDB, logger)

// Starts the scheduler
scheduler.Start()

// Stops gracefully
scheduler.Stop(timeout)
```

**Features:**
- Checks all active managed instances every 30 seconds
- Randomizes order to prevent pattern-based behavior
- Adds random jitter (0-30% delay) to each check to prevent thundering herd
- Limits to 3 concurrent checks to avoid overloading backend
- Graceful shutdown with timeout

#### 2. Database Storage Updates (`backend/internal/storage/managed_instance_store.go`)

New methods for health check operations:

```go
// List all active managed instances for health checking
instances, err := db.ListManagedInstancesForHealthCheck(ctx)

// Update status after check completes
err := db.UpdateManagedInstanceStatus(ctx, instanceID, "connected", nil)
err := db.UpdateManagedInstanceStatus(ctx, instanceID, "error", &errorMsg)
```

**Status Values:**
- `connected` - Instance is reachable
- `error` - Connection test failed
- `unknown` - Not yet checked (initial state)

#### 3. Integration (`backend/cmd/pganalytics-api/main.go`)

Scheduler starts with the API server:

```go
// Initialize scheduler
healthCheckScheduler := jobs.NewHealthCheckScheduler(postgresDB, logger)
healthCheckScheduler.Start()

// Graceful shutdown
healthCheckScheduler.Stop(30 * time.Second)
```

## How It Works

### Periodic Check Cycle

```
30-second interval
  ├─ List all active managed instances
  ├─ Randomize order (prevent pattern)
  └─ For each instance (staggered with jitter):
      ├─ Add random delay (0-30% of interval)
      ├─ Test PostgreSQL connection (with SSL fallback)
      │   ├─ Try: require
      │   ├─ Try: prefer
      │   └─ Try: disable
      └─ Update status in database:
          ├─ Success → last_connection_status = 'connected'
          └─ Failure → last_connection_status = 'error' + error message
```

### Concurrency Control

Uses semaphore pattern to limit concurrent checks:

```
Max Concurrency: 3 checks at once
├─ Check 1 (started immediately)
├─ Check 2 (started immediately)
├─ Check 3 (waits for slot to free)
└─ Check 4+ (queue until slots available)
```

**Benefits:**
- Prevents backend database connection exhaustion
- Distributes load across time
- Responsive to database changes

### Randomization Strategy

1. **Order Randomization**: Shuffles instance list each cycle
   - Prevents always checking same instances first
   - Ensures fair distribution of fresh data

2. **Delay Jitter**: Adds random delay to each check
   - Formula: `random delay = 0 to (30 seconds × 0.3) = 0-9 seconds`
   - Spreads out database/network load
   - Prevents synchronized thundering herd

### Connection Testing

For each instance, attempts connection with fallback SSL modes:

```go
sslModes := []string{"require", "prefer", "disable"}

for _, sslMode := range sslModes {
    if err := testConnection(host, port, user, password, sslMode) {
        // Success - use this mode
        break
    }
}
```

**Features:**
- 5-second connection timeout (prevents hanging)
- 10-second total operation timeout
- Handles various SSL configurations
- Graceful error handling

## Database Schema

Existing columns (already in place):
- `last_connection_status` - Current status: 'connected', 'error', 'unknown', 'invalid_credentials'
- `last_heartbeat` - Timestamp of last successful check
- `last_error_message` - Error details if connection failed
- `last_error_time` - When the error occurred

## Frontend Updates

The frontend polls the API to get fresh status:

```
GET /api/v1/managed-instances
→ Returns all instances with current:
  - last_connection_status
  - last_heartbeat (when last check occurred)
  - last_error_message (if failed)
  - metrics_count_total (active collection)
```

**Recommended Polling:**
- Frontend polls every 5-10 seconds
- Shows real-time status updates
- No manual "Test" button clicks needed

## Performance Implications

**Backend Database:**
- 20 instances checked every 30 seconds = ~40 connection tests/minute max
- Each test is quick (timeout after 5 seconds)
- Staggered execution prevents spikes
- Uses existing connection pool

**Network:**
- Minimal impact: ~0.7 checks/second
- Distributed over time, not simultaneous
- Only affects managed instances being monitored

**Frontend:**
- Poll every 5-10 seconds for updates
- Lightweight API call (uses existing endpoint)
- No additional database queries needed

## Configuration

Current hardcoded settings (can be made configurable):

```go
tickInterval:    30 * time.Second  // Check every 30 seconds
jitterFactor:    0.3               // 30% randomization (0-9 seconds)
maxConcurrency:  3                 // Max 3 concurrent checks
timeout:         5 * time.Second   // Per-connection timeout
```

## Error Handling

**Connection Failures:**
- Tries all SSL modes before giving up
- Records specific error message
- Sets status to 'error' with details
- Logs for debugging

**Database Failures:**
- Logs errors but doesn't crash
- Continues checking remaining instances
- Automatic retry on next cycle

**Graceful Shutdown:**
- Waits for in-flight health checks to complete
- 30-second timeout for cleanup
- Logs completion

## Example Status Updates

### Instance Successfully Connecting

```json
{
  "id": 1,
  "name": "Target PostgreSQL 001",
  "endpoint": "target-postgres-001",
  "last_connection_status": "connected",
  "last_heartbeat": "2026-03-02T22:40:15.235Z",
  "last_error_message": null,
  "last_error_time": null
}
```

### Instance Connection Failed

```json
{
  "id": 2,
  "name": "Target PostgreSQL 002",
  "endpoint": "target-postgres-02",
  "last_connection_status": "error",
  "last_heartbeat": "2026-03-02T22:39:45.891Z",
  "last_error_message": "failed to connect to PostgreSQL: dial tcp: lookup target-postgres-02 on 127.0.0.11:53: no such host",
  "last_error_time": "2026-03-02T22:39:45.891Z"
}
```

## Monitoring

### Check Logs

Look for health check logs:

```bash
docker logs pganalytics-backend | grep "health check"

# Output:
# "Performing health check" instance_id=1
# "Health check passed" instance_id=1 ssl_mode=prefer
# "Health check failed" instance_id=2 error="connection refused"
```

### Scheduler Status

The scheduler exposes status methods:

```go
scheduler.IsRunning()        // bool
scheduler.GetActiveTaskCount() // int - current concurrent checks
```

## Files Modified

- `backend/cmd/pganalytics-api/main.go` - Initialize and manage scheduler lifecycle
- `backend/internal/storage/managed_instance_store.go` - Add health check database operations
- `backend/internal/jobs/health_check_scheduler.go` - New file with scheduler implementation

## Testing the Feature

1. **Start the system:**
   ```bash
   ./cleanup-and-start-load-test.sh
   ```

2. **Register managed instances:**
   ```bash
   ./test-setup-managed-instances.sh
   ```

3. **Watch status update automatically:**
   - Open frontend at http://localhost:4000
   - Go to Manage Instances
   - Watch "Connection Status" column update from "Unknown" → "Connected/Error"
   - No "Test" button clicks needed - updates happen automatically every 30 seconds

4. **Verify in logs:**
   ```bash
   docker logs pganalytics-backend | grep "health check"
   ```

## Future Enhancements

Potential improvements:
- Configurable check interval per instance
- Custom health check queries (not just connection test)
- Health check history/metrics
- Alerting when status changes
- Parallel multi-database checks
- Advanced SSL/certificate validation

## Conclusion

The automatic health check scheduler provides real-time monitoring of managed instances without requiring user interaction or overwhelming the backend with simultaneous requests. Status updates appear automatically in the frontend as checks complete.
