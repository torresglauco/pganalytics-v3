# Architecture Research

**Domain:** PostgreSQL Monitoring Platform Performance Optimization
**Researched:** 2026-05-11
**Confidence:** HIGH (based on comprehensive codebase analysis)

## Existing Architecture Overview

### Current System Structure

```
+------------------------------------------------------------------+
|                          Frontend (React)                         |
|  +-------------+  +-------------+  +-------------+  +----------+ |
|  |  Dashboard  |  | Query       |  | Index       |  |  Metrics | |
|  |  Pages      |  | Analysis    |  | Advisor     |  |  Views   | |
|  +------+------+  +------+------+  +------+------+  +-----+----+ |
|         |                |                |                |     |
+---------+----------------+----------------+----------------+------+
          | REST API calls                    | WebSocket
+---------v----------------+------------------v----------------------+
|                          API Server (Gin)                         |
|  +----------------------------------------------------------------+|
|  |                     Middleware Layer                           ||
|  |  Auth (JWT) | Rate Limit | Request Validation | CORS           ||
|  +----------------------------------------------------------------+|
|  +------------------+  +------------------+  +------------------+  |
|  | handlers_metrics |  | handlers_query   |  | handlers_index   |  |
|  |                  |  | _performance     |  | _advisor         |  |
|  +--------+---------+  +--------+---------+  +--------+---------+  |
|           |                     |                     |            |
+-----------+---------------------+---------------------+------------+
            |                     |                     |
+-----------v-----------+---------v----------+---------v-----------+
|                       | Services Layer     |                     |
|  +------------------+ | +----------------+ | +------------------+|
|  | query_performance | | | log_analysis  | | | index_advisor    ||
|  | (collector,       | | | (collector,   | | | (analyzer,       ||
|  |  analyzer,parser) | | |  parser)      | | |  cost_calculator)||
|  +--------+---------+ | +-------+--------+ | +--------+---------+|
|           |           |         |          |          |          |
+-----------+-----------+---------+----------+----------+----------+
            |                     |                     |
+-----------v---------------------v---------------------v-----------+
|                        Storage Layer                              |
|  +------------------+  +------------------+  +------------------+ |
|  | PostgresDB       |  | TimescaleDB     |  | Cache Manager    | |
|  | (metadata,       |  | (time-series    |  | (in-memory       | |
|  |  config,metrics) |  |  metrics)       |  |  LRU with TTL)   | |
|  +--------+---------+  +--------+---------+  +--------+---------+ |
|           |                     |                     |           |
+-----------+---------------------+---------------------+-----------+
            |                     |                     |
+-----------v---------------------v---------------------v-----------+
|                    PostgreSQL / TimescaleDB                       |
|  +------------------+  +------------------+  +------------------+ |
|  | Users, Auth,     |  | metrics_* tables |  | Hypertables      | |
|  | Collectors       |  | with time-series |  | with compression | |
|  | Config tables    |  | data             |  | and retention    | |
+  +------------------+  +------------------+  +------------------+ |
```

### Component Responsibilities

| Component | Responsibility | Current Implementation |
|-----------|----------------|------------------------|
| `handlers_metrics.go` | HTTP endpoints for metrics retrieval | Direct DB queries, no caching |
| `handlers_query_performance.go` | Query performance analysis endpoints | Basic aggregation, calls storage layer |
| `handlers_index_advisor.go` | Index recommendation endpoints | Database lookups, no caching |
| `query_performance/collector` | Collects query stats from monitored DBs | Batch inserts to storage |
| `query_performance/analyzer` | Analyzes query issues and severity | Stateless, no caching |
| `index_advisor/analyzer` | Finds missing indexes from query plans | Stateless analysis |
| `PostgresDB` | PostgreSQL data access with connection pool | Pool configured: 100 max/20 idle |
| `TimescaleDB` | Time-series metrics storage and queries | Pool: 25 max/5 idle, 5min lifetime |
| `Cache[K,V]` | Generic in-memory cache with TTL and LRU | Already implemented, not fully utilized |

## Current Project Structure

```
backend/
+-- cmd/
|   +-- pganalytics-api/main.go      # API server entry point
|   +-- pganalytics-cli/             # CLI tool
|   +-- pganalytics-mcp-server/      # MCP server
+-- internal/
|   +-- api/                         # HTTP handlers and routing
|   |   +-- handlers*.go             # Feature-specific handlers
|   |   +-- server.go                # Server setup and route registration
|   |   +-- middleware.go            # Auth, rate limiting, validation
|   +-- auth/                        # JWT, password, MFA, SAML, LDAP
|   +-- cache/                       # In-memory cache implementation
|   |   +-- cache.go                 # Generic LRU cache with TTL
|   |   +-- manager.go               # Cache manager for features
|   +-- config/                      # Configuration loading
|   +-- jobs/                        # Scheduled jobs (health checks)
|   +-- metrics/                     # Cache metrics tracking
|   +-- services/                    # Business logic services
|   |   +-- query_performance/       # Query analysis service
|   |   +-- index_advisor/           # Index recommendation service
|   |   +-- log_analysis/            # Log parsing service
|   |   +-- vacuum_advisor/          # Vacuum analysis service
|   +-- storage/                     # Data access layer
|   |   +-- postgres.go              # PostgreSQL connection and queries
|   |   +-- metrics_store.go         # Metrics storage operations
|   |   +-- timescale.go             # TimescaleDB integration
|   +-- timescale/                   # TimescaleDB-specific operations
+-- pkg/
|   +-- handlers/                    # Shared handlers (alerts, conditions)
|   +-- models/                      # Data models
|   +-- services/                    # Shared services (alerts, notifications)
|   +-- errors/                      # Error types
+-- tests/
    +-- benchmarks/                  # Performance benchmarks
    +-- integration/                 # Integration tests
    +-- load/                        # Load testing
```

### Structure Rationale

- **internal/api/**: HTTP layer isolated for clear request/response handling
- **internal/services/**: Business logic decoupled from HTTP concerns
- **internal/storage/**: Data access abstraction enables caching injection
- **internal/cache/**: Already has infrastructure for caching

## Architectural Patterns for Performance Optimization

### Pattern 1: Cache-Aside with Query Plan Caching

**What:** Cache query execution plans alongside results to avoid repeated EXPLAIN ANALYZE calls

**When to use:** For frequently analyzed queries in the query performance service

**Trade-offs:**
- Pros: Reduces database load, faster query analysis
- Cons: Cache invalidation complexity, memory overhead

**Current State:** Cache infrastructure exists (`internal/cache/`) but not used for query plans

**Integration Points:**
```
+----------------+      +----------------+      +----------------+
| Query          |      | Query Plan     |      | Cache          |
| Performance    +----->+ Cache Service  +----->+ (existing)     |
| Handler        |      | (NEW)          |      |                |
+-------+--------+      +-------+--------+      +-------+--------+
        |                       |                       |
        |  1. Check cache       |                       |
        | <-------------------->|                       |
        |                       |                       |
        |  2. If miss, get plan |                       |
        |---------------------->|                       |
        |                       |  3. Store in cache    |
        |                       |---------------------->|
        |                       |                       |
        |  4. Return cached/    |                       |
        |     fresh plan        |                       |
        |<----------------------|                       |
```

**Implementation:**
```go
// New component: internal/services/query_performance/plan_cache.go
type QueryPlanCache struct {
    cache *cache.Cache[string, *QueryPlan]
    db    *storage.PostgresDB
}

func (c *QueryPlanCache) GetPlan(ctx context.Context, queryHash string) (*QueryPlan, error) {
    // Check cache first
    if plan, ok := c.cache.Get(queryHash); ok {
        return plan, nil
    }

    // Fetch and cache
    plan, err := c.db.GetQueryPlan(ctx, queryHash)
    if err != nil {
        return nil, err
    }

    c.cache.Set(queryHash, plan)
    return plan, nil
}
```

### Pattern 2: Connection Pool Optimization per Service

**What:** Dedicated connection pools for different service types with tuned parameters

**When to use:** When services have different access patterns (metrics collection vs. API queries)

**Trade-offs:**
- Pros: Better resource isolation, tuned for access patterns
- Cons: More complex configuration, more connections overall

**Current State:** Single pool for PostgreSQL (100 max/20 idle), single pool for TimescaleDB (25 max/5 idle)

**Recommended Changes:**
```
+-------------------+     +-------------------+
| Metrics Handlers  |     | Query Analysis    |
| (Read-heavy,      |     | Handlers          |
|  dashboard)       |     | (Read/write,      |
+--------+----------+     |  expensive)       |
         |                +--------+----------+
         |                         |
+--------v----------+     +--------v----------+
| Read-Only Pool    |     | Read-Write Pool   |
| - MaxConns: 50    |     | - MaxConns: 30    |
| - MaxIdle: 15     |     | - MaxIdle: 10     |
| - ConnMaxLife: 5m |     | - ConnMaxLife: 10m|
| - ConnMaxIdle: 2m |     | - ConnMaxIdle: 5m |
+--------+----------+     +--------+----------+
         |                         |
         +------------+------------+
                      |
            +---------v----------+
            | PostgreSQL         |
            +--------------------+
```

**Implementation:**
```go
// Modify internal/storage/postgres.go
type PostgresDB struct {
    db          *sql.DB           // General purpose
    readDB      *sql.DB           // Read-optimized for dashboards
    analyticsDB *sql.DB           // Analytics-heavy queries
}

func NewPostgresDBWithPools(cfg PoolConfig) (*PostgresDB, error) {
    // Main pool (existing)
    mainDB, _ := sql.Open("postgres", cfg.PrimaryDSN)
    mainDB.SetMaxOpenConns(100)
    mainDB.SetMaxIdleConns(20)

    // Read replica pool (NEW) - for dashboard aggregations
    readDB, _ := sql.Open("postgres", cfg.ReadReplicaDSN)
    readDB.SetMaxOpenConns(50)
    readDB.SetMaxIdleConns(15)
    readDB.SetConnMaxLifetime(5 * time.Minute)

    return &PostgresDB{db: mainDB, readDB: readDB}, nil
}
```

### Pattern 3: Pre-Aggregated Dashboard Metrics

**What:** Pre-compute and cache dashboard aggregations instead of computing on-demand

**When to use:** For frequently accessed dashboard views with time-series data

**Trade-offs:**
- Pros: Near-instant dashboard loads, reduced database load
- Cons: Storage overhead, data freshness latency

**Current State:** `handleGetMetrics` returns mock data, aggregations computed on-demand

**Recommended Architecture:**
```
+------------------+     +------------------+     +------------------+
| Metrics          |     | Aggregation      |     | Aggregated       |
| Collector        +---->| Worker (NEW)     +---->| Metrics Cache    |
| (existing)       |     |                  |     | (NEW)            |
+--------+---------+     +--------+---------+     +--------+---------+
         |                        |                        |
         |  Raw metrics           |  Aggregated            |  Dashboard
         |  INSERT                |  REFRESH               |  SELECT
         v                        v                        v
+--------+------------------------+------------------------+---------+
|                         TimescaleDB                                 |
|  +------------------+  +------------------+  +------------------+  |
|  | Raw metrics      |  | materialized_*   |  | Continuous       |  |
|  | (hypertables)    |  | views (NEW)      |  | aggregates (NEW) |  |
|  +------------------+  +------------------+  +------------------+  |
+--------------------------------------------------------------------+
```

**Implementation:**
```sql
-- New TimescaleDB continuous aggregates for dashboard
CREATE MATERIALIZED VIEW metrics.dashboard_connection_summary
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    collector_id,
    database_name,
    connection_state,
    AVG(connection_count) AS avg_connections,
    MAX(connection_count) AS max_connections
FROM metrics.metrics_pg_connections_summary
GROUP BY bucket, collector_id, database_name, connection_state;

-- Refresh policy
SELECT add_continuous_aggregate_policy('metrics.dashboard_connection_summary',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');
```

### Pattern 4: Background Index Recommendation Refresh

**What:** Pre-compute index recommendations in background, serve from cache

**When to use:** When index advisor is accessed frequently but data changes slowly

**Trade-offs:**
- Pros: Instant recommendations, decoupled from query analysis
- Cons: Stale recommendations possible, background job overhead

**Current State:** Recommendations computed on-demand in `handleGetIndexAdvisorRecommendations`

**Recommended Flow:**
```
+------------------+     +------------------+     +------------------+
| Index Advisor    |     | Background       |     | Pre-computed     |
| Handler          |     | Analyzer Worker  |     | Recommendations  |
| (existing)       |     | (NEW)            |     | Table (NEW)      |
+--------+---------+     +--------+---------+     +--------+---------+
         |                        |                        |
         |  GET /recommendations  |  Periodic analyze      |
         |                        |                        |
         |  Read from cache       |  Write results         |
         | <--------------------- | ---------------------> |
         |                        |                        |
         v                        v                        v
+--------+------------------------+------------------------+---------+
|                         PostgresDB                                  |
|  +------------------+                                       +-----+ |
|  | index_recommenda-|<--- Periodic REFRESH <---------------|Worker| |
|  | tions (existing) |                                       +-----+ |
|  +------------------+                                               |
+--------------------------------------------------------------------+
```

### Pattern 5: Request Batching for Metrics Collection

**What:** Batch multiple metrics inserts into single transactions

**When to use:** For high-throughput metrics collection from multiple databases

**Trade-offs:**
- Pros: Reduced database round-trips, better throughput
- Cons: Transaction overhead, rollback complexity

**Current State:** `StoreSchemaMetrics`, `StoreLockMetrics`, etc. already use transactions but could batch more efficiently

**Implementation Enhancement:**
```go
// Enhance internal/storage/metrics_store.go
type MetricsBatcher struct {
    db           *PostgresDB
    batchTimeout time.Duration
    batchSize    int
    buffer       chan MetricsBatch
}

func (b *MetricsBatcher) Collect() {
    // Collect metrics for multiple collectors
    // Batch insert when buffer full or timeout
}
```

## Data Flow for Performance Features

### Dashboard Aggregation Flow (Optimized)

```
[Frontend Dashboard]
        |
        | GET /api/v1/metrics?instance_id=X&time_range=24h
        v
[handleGetMetrics] --> Check Cache Manager (existing)
        |                     |
        |                     | HIT --> Return cached aggregations
        |                     v
        |                     MISS
        |                     |
        v                     v
[Read-Optimized Pool] --> Query Materialized Views
        |                     |
        |                     v
        |                 [TimescaleDB Continuous Aggregates]
        |                     |
        v                     v
[Cache Results] <-- Store in Cache with 5-min TTL
        |
        v
[Return to Frontend]
```

### Query Plan Cache Flow

```
[Frontend Query Analysis]
        |
        | GET /api/v1/queries/:hash/performance
        v
[handleGetQueryPerformance]
        |
        v
[QueryPlanCache.GetPlan] --> Cache HIT?
        |                         |
        | YES <-------------------| NO
        |                         |
        v                         v
[Return Cached Plan]      [Execute EXPLAIN ANALYZE]
                                    |
                                    v
                            [Cache Result with 1-hour TTL]
                                    |
                                    v
                            [Return to Handler]
```

### Index Recommendation Flow (Optimized)

```
[Frontend Index Advisor]
        |
        | GET /api/v1/index-advisor/database/:id/recommendations
        v
[handleGetIndexAdvisorRecommendations]
        |
        v
[Check Pre-computed Table] --> Has fresh recommendations?
        |                              |
        | YES <------------------------| NO (stale or missing)
        |                              |
        v                              v
[Return from Table]          [Trigger Background Refresh]
        |                              |
        |                              v
        |                      [Index Analyzer Worker]
        |                              |
        |                              v
        |                      [Update recommendations table]
        |                              |
        | <----------------------------+
        v
[Return to Frontend]
```

## Integration Points Summary

### New Components Required

| Component | Location | Purpose | Integrates With |
|-----------|----------|---------|-----------------|
| `QueryPlanCache` | `internal/services/query_performance/` | Cache query execution plans | `handlers_query_performance.go`, existing `cache.Manager` |
| `AggregationWorker` | `internal/jobs/` | Pre-compute dashboard metrics | `storage/postgres.go`, `timescale/timescale.go` |
| `MetricsBatcher` | `internal/storage/` | Batch metrics inserts | Existing collectors |
| `IndexRefreshWorker` | `internal/jobs/` | Background index analysis | `services/index_advisor/`, `storage/postgres.go` |
| `ConnectionPoolManager` | `internal/storage/` | Manage multiple pools | `postgres.go`, `timescale.go` |

### Existing Components to Modify

| Component | Changes Required | Performance Impact |
|-----------|------------------|-------------------|
| `postgres.go` | Add read replica pool, optimize connection settings | Reduced query latency for dashboards |
| `cache/manager.go` | Add query plan cache, aggregation cache | Faster repeated queries |
| `timescale.go` | Add continuous aggregates, optimize queries | Instant dashboard loads |
| `handlers_metrics.go` | Use read pool, check cache first | Reduced response time |
| `handlers_query_performance.go` | Integrate query plan cache | Faster query analysis |
| `handlers_index_advisor.go` | Use pre-computed recommendations | Instant recommendations |

### API Endpoint Changes

| Endpoint | Current Behavior | Optimized Behavior |
|----------|------------------|-------------------|
| `GET /api/v1/metrics` | Returns mock data | Returns cached aggregations from materialized views |
| `GET /api/v1/queries/:hash/performance` | Direct DB query | Cache-first with query plan caching |
| `GET /api/v1/index-advisor/database/:id/recommendations` | On-demand computation | Pre-computed with background refresh |
| `GET /api/v1/collectors/:id/schema` | Direct DB query | Cache with 10-minute TTL |
| `GET /api/v1/collectors/:id/locks` | Direct DB query | Cache with 30-second TTL |

## Build Order Considering Dependencies

### Phase 1: Connection Pool Optimization
**Dependencies:** None (foundation)
**Delivers:** Better database connection management
**Files to modify:**
- `internal/storage/postgres.go` - Add pool configuration
- `internal/config/config.go` - Add pool settings

### Phase 2: Query Plan Caching
**Dependencies:** Phase 1 (needs optimized pools)
**Delivers:** Faster query performance analysis
**New files:**
- `internal/services/query_performance/plan_cache.go`
**Files to modify:**
- `internal/api/handlers_query_performance.go`
- `internal/cache/manager.go`

### Phase 3: Dashboard Aggregation Pre-computation
**Dependencies:** Phase 1 (needs optimized TimescaleDB pool)
**Delivers:** Instant dashboard loads
**New files:**
- `internal/jobs/aggregation_worker.go`
- Database migrations for materialized views
**Files to modify:**
- `internal/timescale/timescale.go`
- `internal/api/handlers_metrics.go`

### Phase 4: Index Recommendation Caching
**Dependencies:** Phase 2 (follows caching pattern)
**Delivers:** Instant index recommendations
**New files:**
- `internal/jobs/index_refresh_worker.go`
**Files to modify:**
- `internal/api/handlers_index_advisor.go`
- `internal/services/index_advisor/analyzer.go`

### Phase 5: Metrics Collection Batching
**Dependencies:** Phase 1 (needs optimized insert pool)
**Delivers:** Higher throughput metrics collection
**New files:**
- `internal/storage/metrics_batcher.go`
**Files to modify:**
- `internal/services/query_performance/collector.go`
- `internal/services/log_analysis/collector.go`

## Scaling Considerations

| Concern | At 10 collectors | At 100 collectors | At 500+ collectors |
|---------|------------------|-------------------|-------------------|
| Connection pools | Single pool (100 conns) sufficient | Add read replica pool | Multiple read replicas, pool per service type |
| Cache size | 1000 items, 5MB | 5000 items, 25MB | 10000 items, 50MB with eviction |
| Aggregation refresh | Every 5 minutes | Every 2 minutes | Every minute, parallel workers |
| Metrics batch size | 100 items | 500 items | 1000 items with parallel processing |
| Index analysis | On-demand | Background hourly | Background every 15 minutes |

## Anti-Patterns to Avoid

### Anti-Pattern 1: Caching Everything

**What people do:** Add caching to every handler without considering data freshness requirements

**Why it's wrong:** Stale data in monitoring dashboards can hide critical issues; excessive caching consumes memory

**Do this instead:**
- Cache metadata (schema, extensions) with long TTL (10+ minutes)
- Cache aggregations with medium TTL (2-5 minutes)
- Never cache real-time data (locks, active connections) or use very short TTL (< 30 seconds)

### Anti-Pattern 2: One Giant Connection Pool

**What people do:** Create a single large connection pool for all database operations

**Why it's wrong:** Dashboard queries (read-heavy, fast) compete with analytics queries (read-heavy, slow) for connections

**Do this instead:**
- Separate pools by access pattern: read-only for dashboards, read-write for mutations
- Use connection pooler like PgBouncer for additional pooling at infrastructure level

### Anti-Pattern 3: Pre-computing All Aggregations

**What people do:** Create materialized views for every possible time range and metric combination

**Why it's wrong:** Excessive storage overhead, long refresh times, maintenance burden

**Do this instead:**
- Focus on common dashboard views: 1h, 24h, 7d aggregations
- Use TimescaleDB continuous aggregates for automatic management
- Keep raw data for ad-hoc queries, pre-compute only common paths

### Anti-Pattern 4: Synchronous Expensive Operations

**What people do:** Run EXPLAIN ANALYZE or index analysis in the request path

**Why it's wrong:** Blocks the request, poor user experience, can timeout

**Do this instead:**
- Pre-compute expensive analyses in background workers
- Serve results from cache/pre-computed tables
- Return "analysis in progress" for stale data with async refresh trigger

## Sources

- Existing codebase analysis (HIGH confidence - comprehensive review performed)
- Go database/sql connection pool documentation (MEDIUM confidence - standard patterns)
- TimescaleDB continuous aggregates documentation (MEDIUM confidence - standard patterns)
- PostgreSQL query optimization best practices (MEDIUM confidence - standard patterns)

---
*Architecture research for: Performance Optimization in PostgreSQL Monitoring Platform*
*Researched: 2026-05-11*