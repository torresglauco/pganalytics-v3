# pgAnalytics v3 - Full Regression Test Plan

## Overview

This regression test suite validates pgAnalytics v3 with a comprehensive infrastructure deployment of:
- **40 PostgreSQL target instances** (simulating databases to be monitored)
- **40 Collectors** (one monitoring each target)
- **20 Managed Instances** (from the first 20 collectors)
- Complete end-to-end validation of recent fixes

## Recent Fixes Being Validated

1. **Collector ID Persistence**: Collector IDs persist to `/var/lib/pganalytics/collector.id`
2. **Auto-registration Only on First Startup**: Registration only happens once per collector
3. **Registration Secret Tracking**: Secrets properly track usage counts
4. **Dedicated Token Refresh**: Prevents re-registration on token expiry
5. **No Duplicate Registrations**: Each collector registers exactly once

## Test Infrastructure

### File Structure

```
pganalytics-v3/
├── docker-compose-load-test.yml          # Main composition for 40+40 setup
├── cleanup-and-start-load-test.sh        # Phase 1: Clean & start infrastructure
├── test-setup-managed-instances.sh       # Phase 2: Register managed instances
├── verify-regression-tests.sh            # Phase 3: Validate everything
└── REGRESSION_TEST_README.md             # This file
```

### Network Architecture

```
172.20.0.0/16 (pganalytics bridge network)
├── Core Services (172.20.0.x)
│   ├── 172.20.0.10    PostgreSQL (metadata)
│   ├── 172.20.0.11    TimescaleDB (metrics)
│   ├── 172.20.0.20    Backend API
│   └── 172.20.0.60    Frontend UI
├── Target PostgreSQL (172.20.1.101-140)
│   ├── 172.20.1.101   target-postgres-001 (port 5450)
│   ├── 172.20.1.102   target-postgres-002 (port 5451)
│   └── ...
│   └── 172.20.1.140   target-postgres-040 (port 5489)
└── Collectors (172.20.1.201-240)
    ├── 172.20.1.201   collector-001 → target-postgres-001
    ├── 172.20.1.202   collector-002 → target-postgres-002
    └── ...
    └── 172.20.1.240   collector-040 → target-postgres-040
```

### Credentials

- **PostgreSQL Metadata DB**: `postgres:pganalytics` @ localhost:5432
- **TimescaleDB**: `postgres:pganalytics` @ localhost:5433
- **Target Instances**: `postgres:pganalytics` (all targets)
- **API Admin User**: `admin:admin`
- **Registration Secret**: `test-registration-secret-12345`

## Execution Steps

### Step 1: Clean and Start Infrastructure

```bash
./cleanup-and-start-load-test.sh
```

This script:
- ✓ Stops all existing containers (both compose files)
- ✓ Removes all pganalytics volumes and orphaned images
- ✓ Builds fresh images with docker-compose
- ✓ Starts 1 core PostgreSQL + 1 TimescaleDB + 1 Backend + 1 Frontend
- ✓ Starts 40 target PostgreSQL instances (ports 5450-5489)
- ✓ Starts 40 collectors with auto-registration enabled
- ✓ Waits for core services to be healthy

**Duration**: ~10-15 minutes (mostly Docker image building)

**Expected Output**:
```
✓ Infrastructure started
Waiting for core services to be healthy...
✓ PostgreSQL OK
✓ TimescaleDB OK
✓ Backend OK
✓ Frontend OK
```

**Verification**:
```bash
# Check running containers
docker ps | grep pganalytics | wc -l  # Should show ~84 (1+1+1+1+40+40)

# Check volumes
docker volume ls | grep pganalytics | wc -l  # Should show 42 (postgres, timescale, collector_data_001-040)

# Test API health
curl http://localhost:8080/api/v1/health

# Access Frontend
open http://localhost:4000
```

### Step 2: Setup Managed Instances

Wait 1-2 minutes for all collectors to register automatically, then run:

```bash
./test-setup-managed-instances.sh
```

This script:
- ✓ Authenticates with admin credentials
- ✓ Creates managed instance entries for collectors 001-020
- ✓ Each managed instance connects to target-postgres-001 through target-postgres-020
- ✓ Generates setup report with instance IDs

**Duration**: ~1-2 minutes

**Expected Output**:
```
Phase 1: Waiting for all services to be healthy...
  Waiting for Backend... OK

Phase 2: Authenticating with backend...
✓ Authenticated successfully

Phase 3: Registering 20 managed instances...
  Registering managed instance 1/20 (Target PostgreSQL 001)... OK (ID: 1)
  Registering managed instance 2/20 (Target PostgreSQL 002)... OK (ID: 2)
  ...
  Registering managed instance 20/20 (Target PostgreSQL 020)... OK (ID: 20)

Registration Summary:
  Total Registered: 20 / 20
  Total Failed: 0 / 20
  Total Collectors in backend: 40 / 40

✓ Report generated: regression-test-setup-report.txt
```

### Step 3: Verify System

Wait 2-3 minutes for metrics to be collected, then run:

```bash
./verify-regression-tests.sh
```

This script validates:
- ✓ **Authentication**: Can login to API
- ✓ **Collector Count**: Exactly 40 collectors registered
- ✓ **Duplicate Check**: All 40 collectors have unique UUIDs
- ✓ **Managed Instances**: Exactly 20 managed instances created
- ✓ **Collector Status**: All 40 have "registered" status
- ✓ **Heartbeats**: All collectors sending heartbeats
- ✓ **Metrics Collection**: At least one collector collecting metrics
- ✓ **ID Persistence**: Collector IDs persisted to volumes
- ✓ **Registration Secrets**: All used correct secret
- ✓ **Frontend**: Accessible and responsive

**Duration**: ~2-3 minutes (includes metric collection wait)

**Expected Output**:
```
Phase 1: Authentication
  ✓ Authenticated successfully

Phase 2: Collector Registration Validation
  ✓ Collector count equals 40
  ✓ No duplicate collector UUIDs

Phase 3: Managed Instances Validation
  ✓ Managed instance count equals 20

Phase 4: Collector Status Validation
  ✓ All collectors have 'registered' status
  ✓ Collectors sending heartbeats (40/40)

Phase 5: Metrics Collection Validation
  ✓ Collectors collecting metrics

Phase 6: Collector ID Persistence Validation
  ✓ Collector ID persisted to volume

Phase 7: Registration Secret Validation
  ✓ All collectors registered with correct secret

Phase 8: Frontend Accessibility
  ✓ Frontend is accessible

=== ALL TESTS PASSED ===
Regression test completed successfully!
The system is ready for extended validation.

Report file: regression-test-report.txt
```

## Success Criteria

All of the following must be true:

| Criterion | Expected | Check |
|-----------|----------|-------|
| Collectors Registered | 40 | `curl http://localhost:8080/api/v1/collectors -H "Authorization: Bearer $TOKEN" \| jq '. \| length'` |
| Collector UUIDs Unique | 40 unique | Check no duplicates in response |
| Managed Instances | 20 | `curl http://localhost:8080/api/v1/rds-instances -H "Authorization: Bearer $TOKEN" \| jq '. \| length'` |
| Collector Status | All "registered" | Check all status fields |
| Metrics Collected | >100 per collector | `curl http://localhost:8080/api/v1/collectors/col_001/metrics?limit=10` |
| Frontend Accessible | HTTP 200 | `curl http://localhost:4000` |
| No Duplicates | 0 duplicates | Check UUID list for exact count |
| Registration Secret | Used 40x | Check backend database |
| Heartbeats | All present | All collectors should have last_heartbeat |
| System Stable | No errors | Check logs for 5+ minutes |

## Monitoring During Test

### Real-time Logs

```bash
# All services
docker-compose -f docker-compose-load-test.yml logs -f

# Backend only
docker-compose -f docker-compose-load-test.yml logs -f backend

# Frontend only
docker-compose -f docker-compose-load-test.yml logs -f frontend

# Specific collector (001-040)
docker-compose -f docker-compose-load-test.yml logs -f collector-001

# Specific target (001-040)
docker-compose -f docker-compose-load-test.yml logs -f target-postgres-001
```

### Expected Log Patterns

**Collector Registration (appears once per collector)**:
```
[collector-001] Auto-registering collector for the first time...
[collector-001] Successfully registered with backend
[collector-001] Collector ID persisted to /var/lib/pganalytics/collector.id
```

**Collector Running (repeats every 60 seconds)**:
```
[collector-001] Collector already registered with ID: col_001 (skipping registration)
[collector-001] Collecting metrics from target-postgres-001...
[collector-001] Pushing metrics to backend...
[collector-001] Successfully pushed X metrics
```

**Backend Processing**:
```
[backend] Registered collector: col_001 (UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
[backend] Received metrics from collector col_001 (X metrics)
[backend] Metrics stored in TimescaleDB (X records)
```

## Extended Testing (Optional)

After verification passes, you can run extended tests:

### 1. Collector Restart Test

```bash
# Restart a collector
docker-compose -f docker-compose-load-test.yml restart collector-001

# Check logs - should NOT re-register
docker-compose -f docker-compose-load-test.yml logs collector-001 | tail -20

# Verify same UUID
curl http://localhost:8080/api/v1/collectors/col_001 \
  -H "Authorization: Bearer $TOKEN" | jq '.uuid'
```

**Expected**: UUID remains the same, no new registration attempt

### 2. Metrics Volume Test

```bash
# Wait 1 hour (60 collection cycles)
# Check metrics growth
psql -h localhost -U postgres -d metrics -c \
  "SELECT COUNT(*) as metric_count FROM metrics \
   WHERE collector_id = 'col_001';"

# Should show >600 metrics (40 collectors × 15+ metrics per cycle)
```

**Expected**: Metrics count increases monotonically

### 3. Service Restart Resilience

```bash
# Restart backend
docker-compose -f docker-compose-load-test.yml restart backend

# Collectors should continue sending metrics and retry
# Check that metrics eventually appear again

# Restart target PostgreSQL
docker-compose -f docker-compose-load-test.yml restart target-postgres-001

# Collector-001 should log connection error but keep running
docker-compose -f docker-compose-load-test.yml logs collector-001 | grep -i error
```

**Expected**: System recovers from temporary failures

### 4. Concurrent Startup Test

```bash
# Start all 40 collectors simultaneously (already done)
# Verify no duplicate registrations with race conditions

# Count collectors multiple times
for i in {1..5}; do
  echo "Check $i:"
  curl -s http://localhost:8080/api/v1/collectors \
    -H "Authorization: Bearer $TOKEN" | jq '.[] | .id' | sort | uniq | wc -l
  sleep 5
done

# All should show exactly 40
```

**Expected**: Consistent count of 40 with no race condition issues

## Cleanup

To remove the load test infrastructure:

```bash
# Stop all containers and remove volumes
docker-compose -f docker-compose-load-test.yml down -v

# Or use the cleanup script (keeps standard docker-compose running)
./cleanup-and-start-load-test.sh  # Choose "n" to skip restart
```

## Troubleshooting

### Collectors Not Registering

```bash
# Check backend is healthy
curl http://localhost:8080/api/v1/health

# Check collector logs
docker-compose -f docker-compose-load-test.yml logs collector-001 | grep -i error

# Verify network connectivity
docker-compose -f docker-compose-load-test.yml exec collector-001 \
  ping -c 1 backend

# Check registration secret matches
docker-compose -f docker-compose-load-test.yml exec collector-001 \
  cat /etc/pganalytics/collector.toml | grep secret
```

### Metrics Not Collecting

```bash
# Check collector can connect to target
docker-compose -f docker-compose-load-test.yml exec collector-001 \
  psql -h target-postgres-001 -U postgres -d postgres -c "SELECT 1"

# Check TimescaleDB is accessible
docker-compose -f docker-compose-load-test.yml exec backend \
  psql -h timescale -U postgres -d metrics -c "SELECT count(*) FROM metrics"

# Check collector has started collection
docker-compose -f docker-compose-load-test.yml logs collector-001 | \
  grep -i "collecting\|metrics"
```

### Frontend Not Loading

```bash
# Check frontend logs
docker-compose -f docker-compose-load-test.yml logs frontend

# Test API connectivity
curl http://localhost:8080/api/v1/health

# Check if port 4000 is in use
lsof -i :4000

# Verify frontend build completed
docker-compose -f docker-compose-load-test.yml logs frontend | tail -20
```

### API Authentication Failing

```bash
# Verify admin user exists
docker exec pganalytics-postgres psql -U postgres -d pganalytics \
  -c "SELECT username, role FROM users;"

# Test login directly
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Check JWT settings
curl http://localhost:8080/version
```

## Files Generated

After running the test scripts, you'll have:

```
regression-test-setup-report.txt      # Setup phase results
regression-test-report.txt            # Verification phase results
docker-compose-load-test.yml          # Full composition (40+40)
cleanup-and-start-load-test.sh        # Cleanup & startup script
test-setup-managed-instances.sh       # Managed instance registration
verify-regression-tests.sh            # Verification script
```

## Performance Expectations

| Metric | Value | Notes |
|--------|-------|-------|
| Infrastructure Startup | 10-15 min | Mostly Docker image building |
| Collector Registration | 1-2 min | Collectors auto-register on startup |
| Metrics Appearance | 1-3 min | First metrics within 1-2 cycles |
| Setup Phase | 1-2 min | Register 20 managed instances |
| Verification Phase | 2-3 min | Includes metric collection wait |
| **Total Test Time** | **~30-45 min** | Full regression test suite |

## Docker Resources

```
Container Count:
  - 4 core services (postgres, timescale, backend, frontend)
  - 40 target PostgreSQL instances
  - 40 collectors
  = 84 total containers

Network Connections:
  - ~2000 TCP connections (40 collectors × ~50 per collector)
  - Docker bridge can handle this easily

Disk Usage:
  - PostgreSQL data: ~500MB-1GB
  - TimescaleDB data: ~100MB-500MB (depending on metric retention)
  - Docker images: ~3-4GB (base images + builds)
  - Volumes: ~1-2GB

Memory Usage:
  - Per PostgreSQL instance: ~50-100MB
  - Per collector: ~10-20MB
  - Backend API: ~100-200MB
  - Frontend: ~50-100MB
  - Total: ~3-4GB
```

## Additional Resources

- **Backend API Docs**: http://localhost:8080/swagger/index.html
- **Metrics Schema**: Check TimescaleDB `metrics` table
- **Collector Configuration**: See `collector/entrypoint.sh`
- **Backend Migrations**: See `backend/migrations/`

## Notes

1. **SSL/TLS**: Set to disabled for testing. Not suitable for production.
2. **Database Credentials**: Change in production (currently hardcoded in compose file).
3. **Target Instances**: Simulated with basic PostgreSQL - no actual data loaded.
4. **Metrics**: Limited to system metrics (disk, CPU, connections) since targets are minimal.
5. **Duration**: Can run for hours without issues - TimescaleDB handles large metric volumes efficiently.

## Questions?

Check the logs:
```bash
docker-compose -f docker-compose-load-test.yml logs backend 2>&1 | tail -100
```

Or test the API manually:
```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Get collectors
curl http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Get managed instances
curl http://localhost:8080/api/v1/rds-instances \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Get metrics for a collector
curl "http://localhost:8080/api/v1/collectors/col_001/metrics?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```
