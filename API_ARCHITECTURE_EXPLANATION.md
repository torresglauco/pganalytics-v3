# API Architecture Explanation: Original vs New Metrics

**Date**: 2026-03-03
**Question**: Why are new metrics (Phase 3) different from original metrics?
**Answer**: Different architectural patterns for different use cases

---

## Original Metrics Architecture

### Original 6 Collectors
1. **pg_stats** - Query statistics
2. **sysstat** - System statistics
3. **disk_usage** - Disk usage metrics
4. **pg_log** - PostgreSQL logs
5. **pg_replication** - Replication info
6. **pg_query_stats** - Query performance stats

### Original API Endpoints

#### Query Statistics Endpoints
```
GET /api/v1/collectors/{id}/queries/slow
  - Returns slow queries (query_time > threshold)
  - Accessible directly from collector

GET /api/v1/collectors/{id}/queries/frequent
  - Returns frequent queries (execution count > threshold)
  - Accessible directly from collector
```

#### Server-Level Endpoints
```
GET /api/v1/servers
  - Returns all monitored servers
  - Aggregated from all collectors

GET /api/v1/servers/{id}
  - Returns specific server details
  - Cross-collector aggregation

GET /api/v1/servers/{id}/metrics
  - Returns server performance metrics
  - Composite from multiple sources (sysstat, disk_usage, etc.)
```

#### Alert Endpoints
```
GET /api/v1/alerts
  - Returns all alerts
  - System-wide alerts

GET /api/v1/alerts/{id}
  - Returns specific alert
  - Alert details and history

POST /api/v1/alerts/{id}/acknowledge
  - Acknowledges alert
  - Alert management
```

### Original Architecture Pattern
```
Original Collectors → Database → Aggregation Layer → REST API
                                  (server-level)
```

---

## New Metrics Architecture (Phase 3)

### New 6 Collectors
1. **pg_schema** - Database schema information
2. **pg_locks** - Database lock monitoring
3. **pg_bloat** - Table/index bloat analysis
4. **pg_cache** - Cache hit ratio metrics
5. **pg_connections** - Connection tracking
6. **pg_extensions** - Extension inventory

### New API Endpoints (Collector-Specific)

#### Schema Metrics Endpoint
```
GET /api/v1/collectors/{id}/schema
  - Returns schema information for specific collector
  - Database filtering, pagination
  - Direct collector access (no aggregation)
```

#### Lock Metrics Endpoint
```
GET /api/v1/collectors/{id}/locks
  - Returns lock information for specific collector
  - Database filtering, pagination
  - Direct collector access
```

#### Bloat Metrics Endpoint
```
GET /api/v1/collectors/{id}/bloat
  - Returns bloat analysis for specific collector
  - Database filtering, pagination
  - Direct collector access
```

#### Cache Metrics Endpoint
```
GET /api/v1/collectors/{id}/cache-hits
  - Returns cache metrics for specific collector
  - Database filtering, pagination
  - Direct collector access
```

#### Connection Metrics Endpoint
```
GET /api/v1/collectors/{id}/connections
  - Returns connection info for specific collector
  - Database filtering, pagination
  - Direct collector access
```

#### Extensions Metrics Endpoint
```
GET /api/v1/collectors/{id}/extensions
  - Returns extension list for specific collector
  - Database filtering, pagination
  - Direct collector access
```

### New Architecture Pattern
```
New Collectors → Database → Direct Storage Access → REST API
                            (no aggregation)
```

---

## Key Differences Explained

### 1. Aggregation Layer

**Original Metrics**:
- Use aggregation layer (servers, alerts)
- Cross-collector data combination
- Higher-level abstraction
- Less granular control

**New Metrics**:
- Direct collector access (no aggregation)
- Individual per-collector access
- More granular control
- Closer to raw data

**Why?**:
- Original metrics are system-wide (server health, overall alerts)
- New metrics are database-specific (schema per DB, locks per collector)
- Different use cases require different patterns

### 2. Query Patterns

**Original Metrics**:
- Server-level queries: `GET /api/v1/servers`
- Returns aggregated data from multiple collectors
- Cross-collector joins in database

**New Metrics**:
- Collector-specific queries: `GET /api/v1/collectors/{id}/schema`
- Returns data for specific collector only
- Single collector focus

**Why?**:
- Server endpoints need to aggregate data from all collectors
- New metrics are collected per-database per-collector
- Aggregation would be redundant (each DB is collected separately)

### 3. Data Granularity

**Original Metrics** (Aggregated):
```
Server Level
  ├─ Overall Performance
  ├─ System Health
  ├─ Replication Status
  └─ Query Trends (aggregated from all collectors)
```

**New Metrics** (Per-Collector):
```
Collector Level
  ├─ Database 1
  │   ├─ Schema Info
  │   ├─ Lock Status
  │   ├─ Bloat Analysis
  │   └─ Cache Metrics
  └─ Database 2
      ├─ Schema Info
      ├─ Lock Status
      ├─ Bloat Analysis
      └─ Cache Metrics
```

### 4. Authentication & Authorization

**Original Metrics**:
- Server-level access control
- Broader permissions
- System-wide visibility

**New Metrics**:
- Collector-level access control
- Per-database visibility
- More fine-grained security

---

## Why Different Endpoints?

### Technical Reasons

1. **Data Structure**
   - Original metrics: System-wide aggregation needed
   - New metrics: Database-specific, per-collector

2. **Query Patterns**
   - Original: Cross-collector joins
   - New: Single collector queries

3. **Performance**
   - Original: Aggregation at API layer
   - New: Direct query from storage

4. **Use Cases**
   - Original: System health dashboard
   - New: Database-specific monitoring

### Architectural Reasons

1. **Separation of Concerns**
   - Server endpoints handle system-level concerns
   - Collector endpoints handle database-level concerns

2. **Scalability**
   - Direct collector access avoids expensive joins
   - Each collector maintains own metric history

3. **Flexibility**
   - New collectors can add endpoints without breaking existing patterns
   - Server endpoints remain unchanged

4. **Future Extensibility**
   - Easy to add more collector-specific metrics
   - Pattern established for future phases

---

## Could We Create Server-Level Endpoints for New Metrics?

### Yes, but...

**Possible Approach**:
```
GET /api/v1/servers/{server_id}/schema
  - Would aggregate schema from all collectors
  - Combine data from multiple databases
  - Requires cross-collector joins
```

**Why Not Done**:
1. **Complexity**: Schema is per-database, not per-server
2. **Performance**: Aggregation overhead for each schema call
3. **Redundancy**: Each collector already provides this
4. **Design**: Keep endpoints close to data source

**Future Option**:
- Phase 4 could add server-level aggregation
- Create derived dashboards combining data
- Aggregate in Grafana rather than API

---

## API Endpoint Summary

### Collector-Specific (New Metrics - Phase 3)
```
GET /api/v1/collectors/{id}/schema          ✅ Database schema
GET /api/v1/collectors/{id}/locks           ✅ Lock information
GET /api/v1/collectors/{id}/bloat           ✅ Bloat analysis
GET /api/v1/collectors/{id}/cache-hits      ✅ Cache performance
GET /api/v1/collectors/{id}/connections     ✅ Connection info
GET /api/v1/collectors/{id}/extensions      ✅ Extension list
```

### Collector-Specific (Original Metrics)
```
GET /api/v1/collectors/{id}/queries/slow    ✅ Query statistics
GET /api/v1/collectors/{id}/queries/frequent ✅ Frequent queries
```

### Server-Level (Original Architecture)
```
GET /api/v1/servers                         ✅ All servers
GET /api/v1/servers/{id}                    ✅ Server details
GET /api/v1/servers/{id}/metrics            ✅ Server metrics
```

### System-Level
```
GET /api/v1/alerts                          ✅ All alerts
GET /api/v1/alerts/{id}                     ✅ Alert details
POST /api/v1/alerts/{id}/acknowledge        ✅ Alert management
```

---

## Recommendation for Complete API Coverage

### Current State
✅ New metrics have individual collector endpoints
✅ Original metrics accessible via server endpoints
⚠️ No server-level aggregation for new metrics

### Option 1: Keep as-is (Recommended for Now)
- Pro: Simple, direct access to data
- Pro: High performance (no aggregation)
- Pro: Follows same pattern as original query endpoints
- Con: Requires knowing collector ID

### Option 2: Add Server-Level Endpoints (Phase 4+)
- Pro: Unified server view
- Pro: Easier for dashboards
- Con: Additional aggregation layer
- Con: Extra database queries

### Option 3: Aggregate in Grafana
- Pro: Flexible visualization
- Pro: No API changes needed
- Pro: Better dashboard control
- Con: Requires dashboard configuration

---

## Conclusion

### Why Different Patterns?

The original metrics use **server-level aggregation** because:
- They're system-wide (server health, alerts, overall performance)
- They need cross-collector data combination
- Server is the natural aggregation point

The new metrics use **collector-specific endpoints** because:
- They're database-specific (schema, locks, etc.)
- Each collector monitors different databases
- Aggregation would be redundant
- Direct access is more efficient

### Architecture is Consistent

Both patterns follow the same principle:
> **Endpoints match the natural data boundaries**

- Server endpoints aggregate at server level
- Collector endpoints provide per-collector access
- Both patterns are optimal for their use cases

### Future Flexibility

If needed, Phase 4 could add:
- Server-level aggregation endpoints
- Grafana dashboards combining data
- Custom aggregation queries

But for now, collector-specific endpoints are:
✅ **Simple** - Direct data access
✅ **Efficient** - No expensive aggregation
✅ **Scalable** - Works with many collectors
✅ **Consistent** - Matches query endpoint pattern

---

**Summary**: Different endpoints for different data patterns. Both are correct for their use cases. ✅
