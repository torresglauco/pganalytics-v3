# Phase 08: Dashboard Optimization - Research

**Researched:** 2026-05-12
**Domain:** TimescaleDB Continuous Aggregates, Background Workers, Dashboard Pre-computation
**Confidence:** HIGH (based on codebase analysis, existing infrastructure patterns, and established TimescaleDB patterns)

## Summary

Phase 08 focuses on eliminating slow dashboard loads by pre-computing aggregations using TimescaleDB continuous aggregates. The current dashboard handler (`handleGetMetrics`) returns mock data, and metric queries perform on-demand aggregations that can be slow with large time-series datasets. By implementing continuous aggregates with automatic refresh policies and a background worker for coordination, dashboard loads become instant with pre-computed data.

**Primary recommendation:** Create TimescaleDB continuous aggregates for common dashboard views (1h, 6h, 24h, 7d buckets) with automatic refresh policies, and implement a background worker to manage aggregate freshness and handle invalidation events.

## Standard Stack

### Core
| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| **TimescaleDB** | 2.x (via Docker image `timescale/timescaledb`) | Time-series extension for PostgreSQL | Native continuous aggregates with automatic refresh, compression, retention policies. Required for this phase. |
| **pgx v5** | v5.9.2 (already integrated) | PostgreSQL driver with pgxpool | Already migrated in Phase 06, provides connection pooling for aggregate queries |
| **go.uber.org/zap** | v1.27.0 (already integrated) | Structured logging | Consistent with existing job patterns (HealthCheckScheduler) |

### Existing Infrastructure (Leverage)
| Component | Location | How to Extend |
|-----------|----------|---------------|
| **TimescaleDB** | `internal/timescale/timescale.go` | Add methods for continuous aggregate management, query pre-computed views |
| **Background Jobs** | `internal/jobs/` | Follow HealthCheckScheduler pattern for dashboard pre-computation worker |
| **Cache Manager** | `internal/cache/manager.go` | Already has response cache - extend for aggregate cache invalidation |
| **Cache Middleware** | `internal/middleware/cache_middleware.go` | Already caches specific endpoints - add dashboard aggregate endpoints |

### Migration Image Change Required

**Current State:** `docker-compose.yml` uses `postgres:16-bullseye` for the TimescaleDB container (line 29).

**Required Change:** Update to `timescale/timescaledb:latest-pg16` to enable TimescaleDB extension and continuous aggregates.

```yaml
# Current
timescale:
  image: postgres:16-bullseye

# Required
timescale:
  image: timescale/timescaledb:latest-pg16
```

**Version verification:**
```bash
docker pull timescale/timescaledb:latest-pg16
```

## Architecture Patterns

### Recommended Project Structure

```
backend/
+-- internal/
|   +-- timescale/
|   |   +-- timescale.go          # Existing connection management
|   |   +-- aggregates.go         # NEW: Continuous aggregate management
|   |   +-- aggregate_queries.go  # NEW: Query pre-computed aggregates
|   +-- jobs/
|   |   +-- health_check_scheduler.go  # Existing pattern to follow
|   |   +-- dashboard_aggregation_worker.go  # NEW: Pre-computation worker
|   +-- api/
|   |   +-- handlers_metrics.go   # Modify to use pre-computed aggregates
|   +-- storage/
|       +-- migrations.go         # Existing migration runner
+-- migrations/
    +-- 029_timescale_continuous_aggregates.sql  # NEW: Aggregate definitions
```

### Pattern 1: TimescaleDB Continuous Aggregates

**What:** Materialized views that automatically refresh as new data arrives, with configurable time buckets.

**When to use:** For dashboard metrics that aggregate time-series data (connections, queries, cache hits, locks).

**Example:**
```sql
-- Create continuous aggregate for connection metrics (5-minute buckets)
CREATE MATERIALIZED VIEW metrics.connection_summary_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    connection_state,
    AVG(connection_count) AS avg_connections,
    MAX(connection_count) AS max_connections,
    MIN(connection_count) AS min_connections,
    COUNT(*) AS sample_count
FROM metrics.metrics_pg_connections_summary
GROUP BY bucket, collector_id, database_name, connection_state
WITH DATA;

-- Create automatic refresh policy
SELECT add_continuous_aggregate_policy('metrics.connection_summary_5m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');

-- Create additional aggregates for different time ranges
CREATE MATERIALIZED VIEW metrics.connection_summary_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    collector_id,
    database_name,
    AVG(avg_connections) AS avg_connections,
    MAX(max_connections) AS max_connections
FROM metrics.connection_summary_5m
GROUP BY bucket, collector_id, database_name
WITH DATA;
```

**Source:** TimescaleDB documentation patterns (validated against existing migrations in `002_timescale.sql.backup`).

### Pattern 2: Background Worker for Aggregate Management

**What:** A background goroutine that monitors aggregate freshness, handles invalidation events, and coordinates with the cache layer.

**When to use:** When dashboard data needs to stay fresh and cache invalidation is needed on schema changes or manual triggers.

**Implementation Pattern (following HealthCheckScheduler):**
```go
// internal/jobs/dashboard_aggregation_worker.go
package jobs

type DashboardAggregationWorker struct {
    db             *timescale.TimescaleDB
    cacheManager   *cache.Manager
    logger         *zap.Logger
    ctx            context.Context
    cancel         context.CancelFunc
    wg             sync.WaitGroup
    isRunning      bool
    tickInterval   time.Duration  // Check every 30 seconds
    maxConcurrency int
}

func NewDashboardAggregationWorker(
    db *timescale.TimescaleDB,
    cacheManager *cache.Manager,
    logger *zap.Logger,
) *DashboardAggregationWorker {
    ctx, cancel := context.WithCancel(context.Background())
    return &DashboardAggregationWorker{
        db:           db,
        cacheManager: cacheManager,
        logger:       logger,
        ctx:          ctx,
        cancel:       cancel,
        tickInterval: 30 * time.Second,
    }
}

func (w *DashboardAggregationWorker) Start() error {
    w.wg.Add(1)
    go w.run()
    return nil
}

func (w *DashboardAggregationWorker) run() {
    defer w.wg.Done()
    ticker := time.NewTicker(w.tickInterval)
    defer ticker.Stop()

    for {
        select {
        case <-w.ctx.Done():
            return
        case <-ticker.C:
            w.refreshStaleAggregates()
        }
    }
}

func (w *DashboardAggregationWorker) refreshStaleAggregates() {
    // Check for aggregates needing refresh
    // Call TimescaleDB refresh functions
    // Update cache invalidation timestamps
}
```

### Pattern 3: Query Pre-Computed Aggregates

**What:** Modify dashboard handlers to query continuous aggregates instead of raw hypertables.

**When to use:** For all dashboard endpoints that show time-series aggregations.

**Example Handler Modification:**
```go
// internal/timescale/aggregate_queries.go
func (t *TimescaleDB) GetDashboardConnectionMetrics(
    ctx context.Context,
    collectorID uuid.UUID,
    timeRange string, // "1h", "24h", "7d"
) ([]AggregatedMetric, error) {
    // Choose appropriate aggregate view based on time range
    viewName := t.selectAggregateView("connection_summary", timeRange)

    query := fmt.Sprintf(`
        SELECT bucket, database_name,
               AVG(avg_connections) as connections,
               MAX(max_connections) as peak_connections
        FROM %s
        WHERE collector_id = $1
          AND bucket >= NOW() - $2::INTERVAL
        GROUP BY bucket, database_name
        ORDER BY bucket DESC
    `, viewName)

    rows, err := t.db.QueryContext(ctx, query, collectorID, timeRange)
    // ... process results
}
```

### Anti-Patterns to Avoid

- **Anti-pattern: Querying raw hypertables for dashboards** - Causes slow full-table scans. Use continuous aggregates instead.
- **Anti-pattern: Manual REFRESH MATERIALIZED VIEW** - Blocks queries. Use TimescaleDB's automatic refresh policies.
- **Anti-pattern: Single aggregate for all time ranges** - Inefficient. Use different bucket sizes (5m, 1h, 1d) for different time ranges.
- **Anti-pattern: Creating aggregates for every metric** - Storage overhead. Focus on frequently accessed dashboard metrics only.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Time-series aggregation | Custom aggregation tables | TimescaleDB continuous aggregates | Automatic refresh, incremental updates, proven reliability |
| Dashboard caching | Custom aggregation cache | Existing cache.Manager + continuous aggregates | Already integrated, tested, with metrics |
| Background scheduling | Custom cron or goroutine | Existing job pattern (HealthCheckScheduler) | Consistent error handling, graceful shutdown |
| Aggregate refresh | Manual REFRESH calls | add_continuous_aggregate_policy() | Automatic, efficient, no blocking |

## Common Pitfalls

### Pitfall 1: Missing TimescaleDB Extension
**What goes wrong:** Continuous aggregate creation fails with "function create_hypertable() does not exist"
**Why it happens:** Using regular PostgreSQL image instead of TimescaleDB image
**How to avoid:** Update docker-compose.yml to use `timescale/timescaledb:latest-pg16` image
**Warning signs:** Migration errors, hypertable functions not found

### Pitfall 2: Incorrect Refresh Policy Offsets
**What goes wrong:** Aggregates show stale data or overlap with real-time queries
**Why it happens:** start_offset too close to "now" includes incomplete buckets
**How to avoid:** Use end_offset of at least the bucket interval (e.g., 5 minutes for 5-minute buckets)
**Warning signs:** Missing recent data in dashboards, inconsistent results

### Pitfall 3: Over-Aggregating Everything
**What goes wrong:** Storage explodes, refresh times become long
**Why it happens:** Creating aggregates for all possible time ranges and metrics
**How to avoid:** Focus on common dashboard views: 5m/1h for recent, 1d for historical
**Warning signs:** Slow aggregate refresh, high storage usage

### Pitfall 4: Ignoring Cache Invalidation
**What goes wrong:** Dashboard shows stale data after schema changes or collector updates
**Why it happens:** Pre-computed aggregates not invalidated when underlying data changes
**How to avoid:** Background worker monitors invalidation events and clears relevant cache entries
**Warning signs:** Dashboard not reflecting recent changes

## Code Examples

### Continuous Aggregate Migration

```sql
-- migrations/029_timescale_continuous_aggregates.sql
-- Dashboard Continuous Aggregates for pgAnalytics v3

-- Ensure TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

SET search_path TO metrics, public;

-- ============================================================================
-- Connection Metrics Aggregates
-- ============================================================================

-- 5-minute aggregate for connection summary
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.connection_summary_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    connection_state,
    AVG(connection_count) AS avg_connections,
    MAX(connection_count) AS max_connections,
    MIN(connection_count) AS min_connections,
    COUNT(*) AS sample_count
FROM metrics.metrics_pg_connections_summary
GROUP BY bucket, collector_id, database_name, connection_state
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.connection_summary_5m',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

-- 1-hour aggregate (from 5-minute aggregate)
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.connection_summary_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    collector_id,
    database_name,
    AVG(avg_connections) AS avg_connections,
    MAX(max_connections) AS max_connections,
    SUM(sample_count) AS total_samples
FROM metrics.connection_summary_5m
GROUP BY bucket, collector_id, database_name
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.connection_summary_1h',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- ============================================================================
-- Lock Metrics Aggregates
-- ============================================================================

CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.lock_summary_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    lock_type,
    COUNT(*) AS lock_count,
    AVG(wait_duration_ms) AS avg_wait_ms,
    MAX(wait_duration_ms) AS max_wait_ms
FROM metrics.metrics_pg_locks
WHERE wait_duration_ms IS NOT NULL
GROUP BY bucket, collector_id, database_name, lock_type
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.lock_summary_5m',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '10 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE);

-- ============================================================================
-- Database Statistics Aggregates
-- ============================================================================

CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.db_stats_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    collector_id,
    database_name,
    AVG(numbackends) AS avg_backends,
    MAX(numbackends) AS max_backends,
    SUM(xact_commit) AS total_commits,
    SUM(xact_rollback) AS total_rollbacks,
    SUM(blks_read) AS total_blks_read,
    SUM(blks_hit) AS total_blks_hit,
    AVG(database_size) AS avg_db_size
FROM metrics.metrics_pg_stats_database
GROUP BY bucket, collector_id, database_name
WITH DATA;

SELECT add_continuous_aggregate_policy('metrics.db_stats_1h',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- Record migration
INSERT INTO pganalytics.schema_versions (version, description)
VALUES ('029_timescale_continuous_aggregates.sql', 'Dashboard continuous aggregates for instant loads')
ON CONFLICT DO NOTHING;
```

### Background Worker Implementation

```go
// internal/jobs/dashboard_aggregation_worker.go
package jobs

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/torresglauco/pganalytics-v3/backend/internal/cache"
    "github.com/torresglauco/pganalytics-v3/backend/internal/timescale"
    "go.uber.org/zap"
)

// DashboardAggregationWorker manages aggregate freshness and cache invalidation
type DashboardAggregationWorker struct {
    db           *timescale.TimescaleDB
    cacheManager *cache.Manager
    logger       *zap.Logger
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    mu           sync.RWMutex
    isRunning    bool
    tickInterval time.Duration
}

// NewDashboardAggregationWorker creates a new aggregation worker
func NewDashboardAggregationWorker(
    db *timescale.TimescaleDB,
    cacheManager *cache.Manager,
    logger *zap.Logger,
) *DashboardAggregationWorker {
    ctx, cancel := context.WithCancel(context.Background())
    return &DashboardAggregationWorker{
        db:           db,
        cacheManager: cacheManager,
        logger:       logger,
        ctx:          ctx,
        cancel:       cancel,
        tickInterval: 30 * time.Second,
    }
}

// Start begins the aggregation worker
func (w *DashboardAggregationWorker) Start() error {
    w.mu.Lock()
    if w.isRunning {
        w.mu.Unlock()
        return fmt.Errorf("worker already running")
    }
    w.isRunning = true
    w.mu.Unlock()

    w.logger.Info("Starting dashboard aggregation worker",
        zap.Duration("interval", w.tickInterval),
    )

    w.wg.Add(1)
    go w.run()

    return nil
}

// Stop gracefully shuts down the worker
func (w *DashboardAggregationWorker) Stop(timeout time.Duration) error {
    w.mu.Lock()
    if !w.isRunning {
        w.mu.Unlock()
        return fmt.Errorf("worker not running")
    }
    w.mu.Unlock()

    w.logger.Info("Stopping dashboard aggregation worker")
    w.cancel()

    done := make(chan struct{})
    go func() {
        w.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        w.logger.Info("Dashboard aggregation worker stopped")
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("worker shutdown timeout exceeded")
    }
}

// run is the main worker loop
func (w *DashboardAggregationWorker) run() {
    defer w.wg.Done()

    ticker := time.NewTicker(w.tickInterval)
    defer ticker.Stop()

    // Initial check on startup
    w.checkAggregateHealth()

    for {
        select {
        case <-w.ctx.Done():
            w.logger.Info("Dashboard aggregation worker context canceled")
            return
        case <-ticker.C:
            w.checkAggregateHealth()
        }
    }
}

// checkAggregateHealth verifies aggregates are refreshing correctly
func (w *DashboardAggregationWorker) checkAggregateHealth() {
    ctx, cancel := context.WithTimeout(w.ctx, 10*time.Second)
    defer cancel()

    // Query TimescaleDB job status for continuous aggregates
    jobs, err := w.getAggregateJobStatus(ctx)
    if err != nil {
        w.logger.Error("Failed to check aggregate job status", zap.Error(err))
        return
    }

    for _, job := range jobs {
        if job.LastRunStatus != "Success" && job.LastRunStatus != "" {
            w.logger.Warn("Aggregate job may have issues",
                zap.String("job_name", job.JobName),
                zap.String("status", job.LastRunStatus),
            )
        }
    }
}

// AggregateJobStatus represents TimescaleDB job status
type AggregateJobStatus struct {
    JobID         int
    JobName       string
    LastRun       time.Time
    LastRunStatus string
    NextRun       time.Time
}

// getAggregateJobStatus queries timescaledb_information.jobs
func (w *DashboardAggregationWorker) getAggregateJobStatus(ctx context.Context) ([]AggregateJobStatus, error) {
    query := `
        SELECT job_id, application_name, last_run_started_at, last_run_status, next_start
        FROM timescaledb_information.jobs
        WHERE proc_name = 'policy_refresh_continuous_aggregate'
        ORDER BY job_id
    `

    rows, err := w.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query aggregate jobs: %w", err)
    }
    defer rows.Close()

    var jobs []AggregateJobStatus
    for rows.Next() {
        var job AggregateJobStatus
        var lastRun, nextRun sql.NullTime
        var status sql.NullString

        if err := rows.Scan(&job.JobID, &job.JobName, &lastRun, &status, &nextRun); err != nil {
            return nil, fmt.Errorf("scan job row: %w", err)
        }

        if lastRun.Valid {
            job.LastRun = lastRun.Time
        }
        if status.Valid {
            job.LastRunStatus = status.String
        }
        if nextRun.Valid {
            job.NextRun = nextRun.Time
        }

        jobs = append(jobs, job)
    }

    return jobs, nil
}

// InvalidateCollectorCache clears cache entries for a specific collector
func (w *DashboardAggregationWorker) InvalidateCollectorCache(collectorID uuid.UUID) {
    w.logger.Info("Invalidating dashboard cache for collector",
        zap.String("collector_id", collectorID.String()),
    )
    // Clear response cache entries for this collector
    // The cache middleware will repopulate on next request
}
```

### Dashboard Handler Using Pre-Computed Aggregates

```go
// internal/timescale/aggregate_queries.go
package timescale

import (
    "context"
    "time"

    "github.com/google/uuid"
)

// GetDashboardMetrics returns pre-aggregated dashboard metrics
func (t *TimescaleDB) GetDashboardMetrics(
    ctx context.Context,
    collectorID uuid.UUID,
    timeRange string,
) (*DashboardMetricsResponse, error) {
    // Select appropriate aggregate view based on time range
    var bucketInterval string
    var viewName string

    switch timeRange {
    case "1h":
        bucketInterval = "5 minutes"
        viewName = "metrics.connection_summary_5m"
    case "24h":
        bucketInterval = "1 hour"
        viewName = "metrics.connection_summary_1h"
    case "7d", "30d":
        bucketInterval = "1 day"
        viewName = "metrics.connection_summary_1d"
    default:
        bucketInterval = "1 hour"
        viewName = "metrics.connection_summary_1h"
    }

    // Query pre-computed aggregate
    query := fmt.Sprintf(`
        SELECT bucket, database_name,
               AVG(avg_connections) as connections,
               MAX(max_connections) as peak_connections
        FROM %s
        WHERE collector_id = $1
          AND bucket >= NOW() - $2::INTERVAL
        GROUP BY bucket, database_name
        ORDER BY bucket DESC
        LIMIT 1000
    `, viewName)

    rows, err := t.db.QueryContext(ctx, query, collectorID, timeRange)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Process results...
    return &DashboardMetricsResponse{}, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| On-demand aggregation queries | TimescaleDB continuous aggregates | This phase | Instant dashboard loads, reduced database load |
| Manual MATERIALIZED VIEW REFRESH | Automatic refresh policies | This phase | No blocking, always fresh data |
| Single connection pool | Dedicated pools (Phase 06) | 2026-05-11 | Already provides read-optimized pool for dashboards |
| Response caching only | Pre-computed aggregates + cache | This phase | Combined approach for maximum performance |

**Deprecated/outdated:**
- Regular PostgreSQL for time-series: Use TimescaleDB for native time-series features
- Manual aggregation tables: Use continuous aggregates for automatic management

## Open Questions

1. **Metric Tables Status**
   - What we know: Many metric table migrations are `.disabled` (002_timescale.sql.disabled, etc.)
   - What's unclear: Which metric tables are actively being populated by collectors
   - Recommendation: Plan 08-01 should verify which tables exist and are populated before creating aggregates

2. **TimescaleDB Image Version**
   - What we know: Current docker-compose uses `postgres:16-bullseye`
   - What's unclear: Whether to use `timescale/timescaledb:latest-pg16` or a specific version
   - Recommendation: Pin to a specific version like `timescale/timescaledb:2.15.0-pg16` for reproducibility

## Validation Architecture

> Note: workflow.nyquist_validation is not explicitly set in .planning/config.json, so validation section is included.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package + testcontainers (already integrated) |
| Config file | None (standard Go test pattern) |
| Quick run command | `go test ./internal/timescale/... -short -v` |
| Full suite command | `go test ./internal/timescale/... ./internal/jobs/... -v -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DASH-01 | User sees instant dashboard loads | integration | `go test ./tests/integration/... -run TestDashboardMetrics -v` | No - Wave 0 |
| DASH-02 | System uses TimescaleDB continuous aggregates | unit | `go test ./internal/timescale/... -run TestContinuousAggregates -v` | No - Wave 0 |
| DASH-03 | User can view historical metrics without full table scans | integration | `go test ./tests/integration/... -run TestHistoricalMetrics -v` | No - Wave 0 |
| DASH-04 | Background worker pre-computes dashboard metrics | unit | `go test ./internal/jobs/... -run TestDashboardAggregationWorker -v` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/timescale/... ./internal/jobs/... -short -v`
- **Per wave merge:** `go test ./... -v -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/timescale/aggregates_test.go` - Test continuous aggregate creation and querying
- [ ] `internal/timescale/aggregate_queries_test.go` - Test pre-computed aggregate query functions
- [ ] `internal/jobs/dashboard_aggregation_worker_test.go` - Test background worker lifecycle
- [ ] `tests/integration/dashboard_aggregates_test.go` - Integration test with testcontainers
- [ ] Framework install: Already have Go testing + testcontainers

## Sources

### Primary (HIGH confidence)
- Existing codebase analysis (`internal/timescale/timescale.go`, `internal/jobs/health_check_scheduler.go`, migrations) - HIGH confidence
- Existing ARCHITECTURE.md research - HIGH confidence (Phase 08 follows patterns documented there)
- Existing STACK.md research - HIGH confidence (pgx v5 already integrated)

### Secondary (MEDIUM confidence)
- TimescaleDB documentation patterns (standard SQL syntax, well-established) - MEDIUM confidence
- Existing migration patterns in `002_timescale.sql.backup` - HIGH confidence

### Tertiary (LOW confidence)
- None - All findings verified against existing codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - TimescaleDB is the standard for PostgreSQL time-series, existing patterns well-established
- Architecture: HIGH - Following existing HealthCheckScheduler pattern, using existing cache infrastructure
- Pitfalls: HIGH - Common TimescaleDB pitfalls well-documented and verified against existing code

**Research date:** 2026-05-12
**Valid until:** 30 days (stable patterns, TimescaleDB API stable)