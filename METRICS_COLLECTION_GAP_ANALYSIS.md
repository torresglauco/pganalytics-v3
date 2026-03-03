# PostgreSQL Metrics Collection Gap Analysis
## pgAnalytics v3 vs pganalyze Collector

**Date**: 2026-03-03
**Purpose**: Identify metrics collected by pganalyze that pgAnalytics v3 does not currently collect

---

## Executive Summary

pganalyze Collector is more **comprehensive** in schema analysis and query normalization, while pgAnalytics v3 focuses on **operational metrics and health monitoring**. This analysis identifies specific gaps for potential future enhancement.

---

## 1. Metrics We CURRENTLY Collect (pgAnalytics v3)

### Database-Level Metrics
```sql
FROM pg_stat_database:
✅ Database size (bytes)
✅ Transactions committed
✅ Transactions rolled back
✅ Tuples returned
✅ Tuples fetched
✅ Tuples inserted
✅ Tuples updated
✅ Tuples deleted
```

### Table-Level Metrics
```sql
FROM pg_stat_user_tables:
✅ Table size
✅ Live tuples count (n_live_tup)
✅ Dead tuples count (n_dead_tup)
✅ Last vacuum timestamp
✅ Last autovacuum timestamp
✅ Vacuum counts
✅ Autovacuum counts
```

### Index-Level Metrics
```sql
FROM pg_stat_user_indexes:
✅ Index size
✅ Index scan count
✅ Index tuple read count
✅ Index tuple fetch count
```

### Query Performance (pg_stat_statements)
```sql
✅ Query ID
✅ Query text
✅ Call count
✅ Total execution time
✅ Mean execution time
✅ Min/Max execution time
✅ Standard deviation
✅ Rows returned
✅ Shared/Local/Temp block statistics
✅ Read/Write time metrics
✅ WAL records and bytes
```

### System Metrics (sysstat plugin)
```
✅ CPU usage (user, system, idle)
✅ Memory usage
✅ Disk I/O metrics
✅ Network statistics
```

### Replication Metrics (Our Implementation)
```sql
✅ Replication slots status
✅ Streaming replication lag (write, flush, replay)
✅ WAL retention metrics
✅ Logical replication subscriptions
✅ Replica lag summary statistics
```

### Database Logs (pg_log plugin)
```
✅ Log entries collection
✅ Error tracking
✅ Query logging
```

### Disk Usage Metrics
```
✅ Filesystem usage
✅ Database directory sizes
✅ Partition information
```

---

## 2. Metrics pganalyze Collects That We DON'T

### A. SCHEMA INFORMATION (Major Gap)
**Category**: Database Structure Analysis

pganalyze collects detailed schema information:

#### 1. Table Schema Details
```sql
NOT COLLECTED BY US:
❌ Column definitions (data types, constraints)
❌ Column constraints (PRIMARY KEY, UNIQUE, FOREIGN KEY)
❌ Column defaults
❌ Column nullability
❌ Constraint types and definitions
❌ Foreign key relationships
❌ Table inheritance hierarchy
```

**Sample Queries Missing**:
```sql
-- Column information
SELECT
    table_name, column_name, data_type,
    is_nullable, column_default
FROM information_schema.columns
WHERE table_schema NOT IN ('pg_catalog', 'information_schema');

-- Table constraints
SELECT
    constraint_name, constraint_type,
    table_name, column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name;

-- Foreign keys
SELECT
    constraint_name, table_name, column_name,
    referenced_table_name, referenced_column_name
FROM information_schema.referential_constraints;
```

**Impact**: 🔴 **HIGH** - Schema tracking is critical for:
- Database evolution tracking
- Breaking change detection
- Migration planning
- DDL change auditing

---

#### 2. Index Information
```sql
NOT COLLECTED BY US:
❌ Index definition (columns, types)
❌ Index expression (for expression indexes)
❌ Index predicate (for partial indexes)
❌ Index uniqueness
❌ Index method (btree, hash, gist, gin, etc.)
❌ Index bloat percentage
```

**Sample Query Missing**:
```sql
SELECT
    schemaname, tablename, indexname, indexdef,
    idx_scan, idx_tup_read, idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE WHEN idx_scan = 0 THEN 'UNUSED'
         WHEN idx_scan < 100 THEN 'RARELY_USED'
         ELSE 'ACTIVE' END as usage_status
FROM pg_stat_user_indexes
JOIN pg_indexes ON pg_stat_user_indexes.indexrelname = pg_indexes.indexname;
```

**Impact**: 🟡 **MEDIUM** - Important for:
- Index health monitoring
- Identifying unused indexes
- Index bloat detection
- Performance tuning

---

#### 3. Trigger Information
```sql
NOT COLLECTED BY US:
❌ Trigger definitions
❌ Trigger timing (BEFORE/AFTER/INSTEAD OF)
❌ Trigger events (INSERT/UPDATE/DELETE)
❌ Trigger function references
❌ Trigger execution status
```

**Impact**: 🟡 **MEDIUM** - Useful for:
- Understanding database automation
- Identifying performance impacts
- Auditing trigger changes

---

#### 4. Sequence Information
```sql
NOT COLLECTED BY US:
❌ Sequence current value
❌ Sequence increment
❌ Sequence min/max values
❌ Sequence cycle status
```

**Sample Query Missing**:
```sql
SELECT
    sequence_schema, sequence_name,
    last_value, increment_by, min_value, max_value,
    CASE WHEN cycle THEN 'YES' ELSE 'NO' END as cycles
FROM information_schema.sequences
WHERE sequence_schema NOT IN ('pg_catalog', 'information_schema');
```

---

### B. ADVANCED QUERY ANALYSIS (Significant Gap)

pganalyze performs sophisticated query analysis:

#### 1. Query Normalization & Grouping
```sql
NOT COLLECTED BY US:
❌ Query fingerprinting/normalization
❌ Query grouping by pattern
❌ Parameter handling
❌ Literal value removal
❌ Duplicate query detection (functionally equivalent)
```

**What This Means**:
```
pganalyze sees:
  SELECT * FROM users WHERE id = 1;
  SELECT * FROM users WHERE id = 2;
  SELECT * FROM users WHERE id = 3;

As ONE normalized query:
  SELECT * FROM users WHERE id = ?;

We collect them separately (3 different entries)
```

**Impact**: 🔴 **HIGH** - Critical for:
- Query aggregation
- Performance trending
- Identifying hotspot queries
- Database load analysis

---

#### 2. Slow Query Detection & Analysis
```sql
NOT COLLECTED BY US:
❌ Slow query ranking
❌ Query complexity scoring
❌ Cardinality estimation errors
❌ Index missing detection
❌ Query plan analysis
```

**Impact**: 🟡 **MEDIUM** - Important for:
- Proactive performance monitoring
- Automated bottleneck detection
- Index recommendation

---

#### 3. Query Execution Plans
```sql
NOT COLLECTED BY US:
❌ EXPLAIN plan analysis
❌ Query cost estimation
❌ Execution node details
❌ Plan node metrics
❌ Sequential vs Index scan decisions
```

---

### C. TABLE & INDEX ADVANCED METRICS (Moderate Gap)

#### 1. Index Bloat Analysis
```sql
NOT COLLECTED BY US:
❌ Index bloat ratio
❌ Index dead space percentage
❌ Estimated index rebuild benefit
```

**Sample Query Missing**:
```sql
SELECT
    indexrelname,
    ROUND(100 * (CASE WHEN otta > 0
        THEN sml.relpages - otta
        ELSE 0 END) / NULLIF(sml.relpages, 0), 2) as bloat_ratio,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes;
```

---

#### 2. Table Bloat Analysis
```sql
NOT COLLECTED BY US:
❌ Table bloat percentage
❌ Dead tuple ratio
❌ Estimated space reclamable
❌ Vacuum effectiveness
```

**Impact**: 🟡 **MEDIUM** - Useful for:
- Storage optimization
- Vacuum scheduling
- Space reclamation planning

---

#### 3. Sequential Scans vs Index Scans
```sql
NOT COLLECTED BY US:
❌ Sequential scan count
❌ Sequential scan blocks read
❌ Ratio of seq scans to index scans
❌ Seq scan efficiency metrics
```

**We collect some of this but not comprehensive tracking:**
```sql
Currently:
✅ Index scans (idx_scan from pg_stat_user_indexes)
❌ Sequential scans (seq_scan from pg_stat_user_tables) - PARTIALLY
❌ Seq vs Index scan ratio analysis
```

---

### D. CACHE & BUFFER METRICS (Minor Gap)

#### 1. Cache Hit Ratios by Table/Index
```sql
NOT COLLECTED BY US:
❌ Table cache hit ratio (heap_blks_hit / total heap blocks)
❌ Index cache hit ratio
❌ Shared buffer utilization
❌ Cache efficiency metrics
```

**Sample Calculation Missing**:
```sql
SELECT
    schemaname, tablename,
    ROUND(100.0 * heap_blks_hit / (heap_blks_hit + heap_blks_read), 2) as cache_hit_ratio
FROM pg_statio_user_tables
WHERE (heap_blks_hit + heap_blks_read) > 0;
```

---

#### 2. Buffer Pool Statistics
```sql
NOT COLLECTED BY US:
❌ Checkpoint activity
❌ Dirty buffer ratio
❌ Buffer backend writes
❌ Backend page rate
```

---

### E. AUTOVACUUM & MAINTENANCE (Minor Gap)

#### 1. Advanced Vacuum Status
```sql
NOT COLLECTED BY US:
❌ Autovacuum configuration per table
❌ Autovacuum effectiveness (tuples cleaned)
❌ Last autovacuum duration
❌ Autovacuum worker count
❌ Vacuum queue status
```

**We have basic info but not detailed analysis:**
```sql
We collect:
✅ last_autovacuum timestamp
❌ autovacuum_count (detail)
❌ autovacuum_duration
❌ autovacuum_effectiveness
```

---

#### 2. XID Age Tracking
```sql
NOT COLLECTED BY US:
❌ XID wraparound risk percentage
❌ MultiXID age tracking
❌ Wraparound ETA calculation
```

**We have the capability but don't track comprehensively:**
- Our replication queries include XID analysis
- But we don't actively monitor/alert on wraparound risk

---

### F. CLUSTER-LEVEL METRICS (Minor Gap)

#### 1. Connections & Sessions
```sql
NOT COLLECTED BY US:
❌ Active connection count by database
❌ Connection state breakdown
❌ Idle transaction sessions
❌ Long-running transactions
❌ Connection wait times
```

**Sample Queries Missing**:
```sql
SELECT
    datname, state, count(*) as connection_count,
    max(NOW() - pg_stat_activity.query_start) as longest_duration
FROM pg_stat_activity
GROUP BY datname, state;

-- Long transactions
SELECT
    pid, usename, datname, state, query,
    age(now(), pg_stat_activity.query_start) as duration
FROM pg_stat_activity
WHERE state != 'idle'
  AND query_start < NOW() - INTERVAL '5 minutes'
ORDER BY query_start;
```

---

#### 2. Backend Activity Metrics
```sql
NOT COLLECTED BY US:
❌ Backend type breakdown
❌ Parallel worker status
❌ Backend resource usage
❌ Query plan caching statistics
```

---

### G. LOCK MONITORING (Not Collected)

#### 1. Database Locks
```sql
NOT COLLECTED BY US:
❌ Active lock list
❌ Lock wait chains
❌ Deadlock history
❌ Lock contention metrics
```

**Sample Query Missing**:
```sql
SELECT
    l.pid, l.mode, l.granted,
    a.usename, a.datname, a.query,
    a.query_start
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE NOT l.granted
ORDER BY a.query_start;
```

**Impact**: 🔴 **HIGH** - Critical for:
- Identifying blocking queries
- Deadlock analysis
- Transaction isolation issues

---

### H. REPLICATION ADVANCED METRICS (Partial)

pganalyze collects (we have some):

```sql
They collect more:
✅ We have: Replication slot status, lag metrics
❌ We're missing:
    - LSN position analysis (we have queries but don't track actively)
    - Replication timeline history
    - Standby feedback metrics
    - Replication conflicts
```

---

### I. EXTENSIONS & MODULES INFO (Not Collected)

```sql
NOT COLLECTED BY US:
❌ Installed extensions list
❌ Extension version tracking
❌ Custom data type definitions
❌ Custom operator definitions
```

**Sample Query Missing**:
```sql
SELECT extname, extversion, extowner::regrole,
       pg_describe_object('pg_extension'::regclass, oid, 0) as description
FROM pg_extension
ORDER BY extname;
```

---

## 3. Gap Summary Table

| Metric Category | Coverage | Importance | Effort |
|-----------------|----------|-----------|--------|
| **Schema Information** | 🔴 0% | 🔴 HIGH | 🟢 Medium |
| Query Normalization | 🔴 0% | 🔴 HIGH | 🔴 High |
| Query Plans | 🔴 0% | 🟡 MEDIUM | 🔴 High |
| Index Bloat | 🔴 0% | 🟡 MEDIUM | 🟢 Easy |
| Table Bloat | 🔴 0% | 🟡 MEDIUM | 🟢 Easy |
| Cache Hit Ratios | 🔴 0% | 🟡 MEDIUM | 🟢 Easy |
| Locks | 🔴 0% | 🔴 HIGH | 🟢 Easy |
| Connections Detail | 🟡 20% | 🟡 MEDIUM | 🟢 Easy |
| Extensions | 🔴 0% | 🟢 LOW | 🟢 Easy |
| **Collected Metrics** | 🟢 70% | - | - |

---

## 4. Impact Assessment

### HIGH Priority (Should Consider)
1. **Schema Information** (Column definitions, constraints, FK relationships)
   - Enables schema change auditing
   - Supports dependency tracking
   - Critical for compliance/governance

2. **Query Normalization**
   - Reduces noise in metrics
   - Better performance trending
   - Essential for identifying true bottlenecks

3. **Lock Monitoring**
   - Detect blocking queries
   - Troubleshoot deadlocks
   - Understand contention

### MEDIUM Priority (Good to Have)
1. Index/Table Bloat Analysis
   - Optimization insights
   - Space reclamation planning

2. Query Plan Analysis
   - Proactive tuning
   - Index recommendations

3. Cache Hit Ratios
   - Performance insights
   - Buffer tuning

### LOW Priority (Nice to Have)
1. Extensions information
2. Advanced replication metrics
3. Detailed XID tracking

---

## 5. Implementation Recommendations

### Phase 1 (Quick Wins - Easy Implementation)
1. Add cache hit ratio calculations (1-2 hours)
2. Add index bloat detection (1-2 hours)
3. Add lock monitoring (1-2 hours)
4. Enhance connection tracking (1 hour)

**Effort**: ~1 week of development
**Impact**: 20% improvement in metrics coverage

### Phase 2 (Medium Effort)
1. Add schema information collection (40-60 hours)
2. Add table bloat analysis (3-5 hours)
3. Expand lock monitoring (additional details)

**Effort**: 2-3 weeks of development
**Impact**: 40% improvement in metrics coverage

### Phase 3 (High Effort)
1. Query normalization engine (60-80 hours)
2. Query plan analysis (40-60 hours)
3. Advanced query optimization suggestions

**Effort**: 3-4 weeks of development
**Impact**: 60% improvement, significant value add

---

## 6. Current Strengths vs pganalyze

**Our Unique Advantages**:
- ✅ Automatic health check scheduling (30-second intervals)
- ✅ Managed instance connection status monitoring
- ✅ Randomized jitter (prevents thundering herd)
- ✅ Self-hosted infrastructure
- ✅ Real-time status updates (not deferred)
- ✅ Cost-effective scaling (tested to 1,029 instances)

**Their Advantages**:
- ✅ Comprehensive schema analysis
- ✅ Query normalization & deduplication
- ✅ Mature query optimization suggestions
- ✅ Advanced index recommendations
- ✅ Extensive bloat analysis

---

## 7. Conclusion

**Current State**: pgAnalytics v3 covers ~70% of operational metrics that pganalyze covers, with particular strength in:
- Health monitoring
- Replication tracking
- System metrics
- Query statistics (basic)

**Gaps**: Primarily in:
- Schema information (0%)
- Query normalization (0%)
- Bloat analysis (0%)
- Lock monitoring (0%)

**Recommendation**: Focus Phase 1 on quick-win metrics (locks, bloat, cache) to reach 80% parity while maintaining our unique advantages in health monitoring and scalability.

---

**Next Steps**:
1. Prioritize lock monitoring (high impact, easy)
2. Add bloat analysis (medium impact, easy)
3. Plan schema collection for Phase 2
4. Consider query normalization architecture for Phase 3

