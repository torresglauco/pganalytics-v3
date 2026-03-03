# Frontend Metrics Calculations & Analysis Functions
## Formulas, Algorithms, and Business Logic for Visualizations

**Date**: March 3, 2026
**Scope**: All calculations needed for dashboard pages and analytics

---

## Part 1: Health Score Calculations

### Overall Database Health Score

```typescript
// frontend/src/utils/calculations.ts

interface HealthMetrics {
  lockHealth: number;           // 0-100
  bloatHealth: number;          // 0-100
  queryHealth: number;          // 0-100
  cacheHealth: number;          // 0-100
  connectionHealth: number;     // 0-100
  replicationHealth: number;    // 0-100
}

export function calculateOverallHealth(metrics: HealthMetrics): number {
  const weights = {
    lockHealth: 0.15,
    bloatHealth: 0.20,
    queryHealth: 0.15,
    cacheHealth: 0.20,
    connectionHealth: 0.15,
    replicationHealth: 0.15,
  };

  return Math.round(
    metrics.lockHealth * weights.lockHealth +
    metrics.bloatHealth * weights.bloatHealth +
    metrics.queryHealth * weights.queryHealth +
    metrics.cacheHealth * weights.cacheHealth +
    metrics.connectionHealth * weights.connectionHealth +
    metrics.replicationHealth * weights.replicationHealth
  );
}

// ============================================================================
// LOCK HEALTH
// ============================================================================

interface LockMetrics {
  activeLocksCount: number;
  blockedTransactions: number;
  maxWaitTime: number;  // in seconds
}

/**
 * Lock Health Score (0-100)
 *
 * Perfect:    No locks or waits
 * Healthy:    Some locks but no blocking
 * Warning:    Blocking detected, wait time < 5min
 * Critical:   Blocking > 5min or many blocked transactions
 */
export function calculateLockHealth(metrics: LockMetrics): number {
  let score = 100;

  // Penalize for blocked transactions (each -10 points)
  score -= Math.min(metrics.blockedTransactions * 10, 50);

  // Penalize for long wait times
  if (metrics.maxWaitTime > 300) {  // > 5 minutes
    score -= 30;
  } else if (metrics.maxWaitTime > 60) {  // > 1 minute
    score -= 15;
  } else if (metrics.maxWaitTime > 10) {  // > 10 seconds
    score -= 5;
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// BLOAT HEALTH
// ============================================================================

interface BloatMetrics {
  tableCount: number;
  bloatedTableCount: number;      // tables > 30% bloat
  severeBloatCount: number;       // tables > 50% bloat
  avgBloatRatio: number;          // 0-100%
  totalReclaimableBytes: number;
}

/**
 * Bloat Health Score (0-100)
 *
 * Perfect:    No bloat (< 5%)
 * Healthy:    < 20% average bloat
 * Warning:    20-40% bloat, some tables > 30%
 * Critical:   > 40% bloat or severe (> 50%) bloat detected
 */
export function calculateBloatHealth(metrics: BloatMetrics): number {
  let score = 100;

  // Penalize for high average bloat ratio
  if (metrics.avgBloatRatio > 50) {
    score -= 40;
  } else if (metrics.avgBloatRatio > 30) {
    score -= 25;
  } else if (metrics.avgBloatRatio > 20) {
    score -= 15;
  } else if (metrics.avgBloatRatio > 5) {
    score -= 5;
  }

  // Additional penalty for severely bloated tables
  const severePenalty = Math.min(metrics.severeBloatCount * 15, 30);
  score -= severePenalty;

  // Additional penalty for many bloated tables (> 10% of total)
  const bloatRatio = metrics.tableCount > 0
    ? metrics.bloatedTableCount / metrics.tableCount
    : 0;
  if (bloatRatio > 0.1) {
    score -= Math.min(bloatRatio * 20, 20);
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// QUERY PERFORMANCE HEALTH
// ============================================================================

interface QueryMetrics {
  slowQueryCount: number;        // queries > 1s average
  verySlowQueryCount: number;    // queries > 10s average
  avgExecutionTime: number;      // in milliseconds
  highIOQueryCount: number;      // queries with seq scans
}

/**
 * Query Health Score (0-100)
 *
 * Perfect:    All queries < 100ms
 * Healthy:    Some queries 100-500ms
 * Warning:    Queries 500ms-2s or some > 1s
 * Critical:   Many slow queries or queries > 10s
 */
export function calculateQueryHealth(metrics: QueryMetrics): number {
  let score = 100;

  // Penalize for very slow queries
  score -= Math.min(metrics.verySlowQueryCount * 20, 40);

  // Penalize for slow queries
  score -= Math.min(metrics.slowQueryCount * 5, 30);

  // Penalize for average execution time
  if (metrics.avgExecutionTime > 2000) {
    score -= 15;
  } else if (metrics.avgExecutionTime > 500) {
    score -= 10;
  } else if (metrics.avgExecutionTime > 100) {
    score -= 5;
  }

  // Penalize for high-IO queries (candidates for indexing)
  score -= Math.min(metrics.highIOQueryCount * 3, 20);

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// CACHE HEALTH
// ============================================================================

interface CacheHealthMetrics {
  overallHitRatio: number;       // 0-100%
  tableHitRatio: number;         // 0-100%
  indexHitRatio: number;         // 0-100%
  lowHitRatioTableCount: number; // tables < 50% hit ratio
  lowHitRatioIndexCount: number; // indexes < 70% hit ratio
}

/**
 * Cache Health Score (0-100)
 *
 * Perfect:    Overall > 99%, tables > 95%, indexes > 98%
 * Healthy:    Overall > 90%, tables > 80%, indexes > 90%
 * Warning:    Overall > 70%, some tables < 50%
 * Critical:   Overall < 70% or many low-hit tables
 */
export function calculateCacheHealth(metrics: CacheHealthMetrics): number {
  let score = 100;

  // Overall cache hit is most important (40% weight)
  if (metrics.overallHitRatio < 70) {
    score -= (70 - metrics.overallHitRatio) * 0.4;
  } else if (metrics.overallHitRatio < 90) {
    score -= (90 - metrics.overallHitRatio) * 0.15;
  } else if (metrics.overallHitRatio < 99) {
    score -= (99 - metrics.overallHitRatio) * 0.05;
  }

  // Table hit ratio (30% weight)
  if (metrics.tableHitRatio < 80) {
    score -= (80 - metrics.tableHitRatio) * 0.3;
  } else if (metrics.tableHitRatio < 95) {
    score -= (95 - metrics.tableHitRatio) * 0.1;
  }

  // Index hit ratio (30% weight)
  if (metrics.indexHitRatio < 90) {
    score -= (90 - metrics.indexHitRatio) * 0.3;
  } else if (metrics.indexHitRatio < 98) {
    score -= (98 - metrics.indexHitRatio) * 0.05;
  }

  // Penalize for low-hit tables/indexes
  score -= Math.min(metrics.lowHitRatioTableCount * 2, 15);
  score -= Math.min(metrics.lowHitRatioIndexCount * 1, 10);

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// CONNECTION POOL HEALTH
// ============================================================================

interface ConnectionHealthMetrics {
  activeConnections: number;
  idleConnections: number;
  idleInTransactionConnections: number;
  maxConnections: number;
  connectionUsagePercent: number;
  idleOverDuration: number;  // count of connections idle > 5 minutes
  averageConnectionAge: number;  // in seconds
}

/**
 * Connection Health Score (0-100)
 *
 * Perfect:    < 50% utilization, no idle-txn
 * Healthy:    50-80% utilization
 * Warning:    80-95% utilization or some idle-txn
 * Critical:   > 95% utilization or many idle-txn connections
 */
export function calculateConnectionHealth(metrics: ConnectionHealthMetrics): number {
  let score = 100;

  // Penalize for high connection usage
  if (metrics.connectionUsagePercent > 95) {
    score -= 40;
  } else if (metrics.connectionUsagePercent > 90) {
    score -= 30;
  } else if (metrics.connectionUsagePercent > 80) {
    score -= 15;
  } else if (metrics.connectionUsagePercent > 70) {
    score -= 5;
  }

  // Heavily penalize for idle-in-transaction (connection leak indicator)
  if (metrics.idleInTransactionConnections > 5) {
    score -= Math.min(metrics.idleInTransactionConnections * 5, 30);
  }

  // Penalize for too many old idle connections
  if (metrics.idleOverDuration > 10) {
    score -= Math.min(metrics.idleOverDuration * 2, 15);
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// REPLICATION HEALTH
// ============================================================================

interface ReplicationHealthMetrics {
  isReplicating: boolean;
  replicationLagBytes: number;
  replicationLagSeconds: number;
  connectedStandbyCount: number;
  expectedStandbyCount: number;
  walArchiveStatus: 'healthy' | 'warning' | 'critical';
}

/**
 * Replication Health Score (0-100)
 *
 * Perfect:    No lag, all standbys connected, WAL healthy
 * Healthy:    < 1MB lag, all standbys connected
 * Warning:    1-100MB lag or missing standbys
 * Critical:   > 100MB lag or standbys disconnected
 */
export function calculateReplicationHealth(metrics: ReplicationHealthMetrics): number {
  if (!metrics.isReplicating) {
    return 100;  // Replication not configured, not unhealthy
  }

  let score = 100;

  // Penalize for replication lag
  if (metrics.replicationLagBytes > 100 * 1024 * 1024) {  // > 100MB
    score -= 40;
  } else if (metrics.replicationLagBytes > 10 * 1024 * 1024) {  // > 10MB
    score -= 20;
  } else if (metrics.replicationLagBytes > 1 * 1024 * 1024) {  // > 1MB
    score -= 10;
  } else if (metrics.replicationLagBytes > 0) {
    score -= 3;
  }

  // Penalize for disconnected standbys
  const disconnected = metrics.expectedStandbyCount - metrics.connectedStandbyCount;
  if (disconnected > 0) {
    score -= disconnected * 20;
  }

  // Penalize for WAL archive problems
  if (metrics.walArchiveStatus === 'critical') {
    score -= 30;
  } else if (metrics.walArchiveStatus === 'warning') {
    score -= 15;
  }

  return Math.max(0, Math.min(100, score));
}
```

---

## Part 2: Table & Index Analysis Functions

### Bloat Ratio Calculations

```typescript
// ============================================================================
// TABLE BLOAT ANALYSIS
// ============================================================================

interface TableBloatData {
  schemaName: string;
  tableName: string;
  liveRows: number;
  deadTuples: number;
  pages: number;
  avgRowSize: number;
}

export interface TableBloatResult {
  schemaName: string;
  tableName: string;
  bloatRatio: number;             // 0-100
  bloatPercentage: string;        // "25.3%"
  reclaimableBytes: number;
  reclaimableMB: string;
  recommendation: 'none' | 'vacuum' | 'vacuum_full' | 'analyze';
  severity: 'healthy' | 'warning' | 'critical';
}

/**
 * Calculate table bloat ratio and recommendations
 *
 * Bloat occurs when dead tuples accumulate and aren't reclaimed by VACUUM.
 * Formula: (dead_tuples / (live_tuples + dead_tuples)) * 100
 */
export function analyzeTableBloat(data: TableBloatData): TableBloatResult {
  const totalTuples = data.liveRows + data.deadTuples;

  // Calculate bloat ratio
  const bloatRatio = totalTuples > 0
    ? (data.deadTuples / totalTuples) * 100
    : 0;

  // Estimate reclaimable bytes
  // Assuming each dead tuple takes avg_row_size + overhead (24 bytes per tuple in PostgreSQL)
  const tupleOverhead = 24;
  const reclaimablePerTuple = data.avgRowSize + tupleOverhead;
  const reclaimableBytes = Math.round(data.deadTuples * reclaimablePerTuple);

  // Determine severity
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  if (bloatRatio > 50) {
    severity = 'critical';
  } else if (bloatRatio > 30) {
    severity = 'warning';
  }

  // Recommendation
  let recommendation: 'none' | 'vacuum' | 'vacuum_full' | 'analyze' = 'none';
  if (bloatRatio > 50) {
    recommendation = 'vacuum_full';  // Schedule during maintenance window
  } else if (bloatRatio > 20) {
    recommendation = 'vacuum';  // Regular maintenance
  } else if (bloatRatio > 5) {
    recommendation = 'analyze';  // Just analyze
  }

  return {
    schemaName: data.schemaName,
    tableName: data.tableName,
    bloatRatio: Math.round(bloatRatio * 10) / 10,
    bloatPercentage: `${(bloatRatio).toFixed(1)}%`,
    reclaimableBytes,
    reclaimableMB: `${(reclaimableBytes / (1024 * 1024)).toFixed(2)} MB`,
    recommendation,
    severity,
  };
}

// ============================================================================
// INDEX ANALYSIS
// ============================================================================

interface IndexMetrics {
  indexName: string;
  tableName: string;
  sizeBytes: number;
  indexScans: number;
  indexTuples: number;
  isUnused: boolean;
  isDuplicate: boolean;
}

export interface IndexAnalysisResult {
  indexName: string;
  tableName: string;
  sizeMB: string;
  usageScore: number;          // 0-100
  efficiency: 'excellent' | 'good' | 'poor' | 'unused';
  recommendation: 'keep' | 'monitor' | 'consider_drop' | 'duplicate';
  severity: 'healthy' | 'warning' | 'critical';
}

/**
 * Analyze index usage and effectiveness
 *
 * Factors:
 * - Is it used? (high scans = good)
 * - Is it duplicated? (another index does same work)
 * - Size vs benefit? (large unused = bad)
 */
export function analyzeIndex(metrics: IndexMetrics): IndexAnalysisResult {
  let usageScore = 100;
  let recommendation: 'keep' | 'monitor' | 'consider_drop' | 'duplicate' = 'keep';
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  let efficiency: 'excellent' | 'good' | 'poor' | 'unused' = 'excellent';

  if (metrics.isDuplicate) {
    recommendation = 'duplicate';
    severity = 'warning';
    usageScore = 50;
    efficiency = 'poor';
  } else if (metrics.isUnused) {
    recommendation = 'consider_drop';
    severity = 'warning';
    usageScore = 0;
    efficiency = 'unused';
  } else if (metrics.indexScans === 0) {
    recommendation = 'monitor';
    severity = 'critical';
    usageScore = 10;
    efficiency = 'unused';
  } else {
    // Index is being used, calculate efficiency
    const scansPerMB = metrics.indexScans / (metrics.sizeBytes / (1024 * 1024));

    if (scansPerMB < 1) {
      efficiency = 'poor';
      usageScore = 30;
      recommendation = 'consider_drop';
    } else if (scansPerMB < 10) {
      efficiency = 'good';
      usageScore = 70;
    } else {
      efficiency = 'excellent';
      usageScore = 100;
    }
  }

  return {
    indexName: metrics.indexName,
    tableName: metrics.tableName,
    sizeMB: `${(metrics.sizeBytes / (1024 * 1024)).toFixed(2)} MB`,
    usageScore,
    efficiency,
    recommendation,
    severity,
  };
}
```

---

## Part 3: Performance Metrics

### Cache Hit Ratio Calculations

```typescript
// ============================================================================
// CACHE HIT RATIO ANALYSIS
// ============================================================================

interface TableCacheMetrics {
  tableName: string;
  heapBlksHit: number;      // successful cache hits in table blocks
  heapBlksRead: number;     // disk reads of table blocks
  idxBlksHit: number;       // successful cache hits in index blocks
  idxBlksRead: number;      // disk reads of index blocks
}

export interface CacheAnalysisResult {
  tableName: string;
  tableHitRatio: number;      // 0-100
  tableHitRatioPercent: string;
  indexHitRatio: number;      // 0-100
  indexHitRatioPercent: string;
  totalAccesses: number;
  totalCacheHits: number;
  severity: 'healthy' | 'warning' | 'critical';
  recommendation: string;
}

/**
 * Calculate cache hit ratios for tables and indexes
 *
 * Formula: (hits / (hits + reads)) * 100
 *
 * Interpretation:
 * - 99%+: Excellent, data is in cache almost always
 * - 90-99%: Good, most requests served from cache
 * - 70-90%: Fair, significant disk I/O still happening
 * - <70%: Poor, frequent disk reads, consider indexing
 */
export function analyzeCacheHitRatio(metrics: TableCacheMetrics): CacheAnalysisResult {
  // Calculate hit ratios
  const tableTotal = metrics.heapBlksHit + metrics.heapBlksRead;
  const tableHitRatio = tableTotal > 0
    ? (metrics.heapBlksHit / tableTotal) * 100
    : 100;

  const indexTotal = metrics.idxBlksHit + metrics.idxBlksRead;
  const indexHitRatio = indexTotal > 0
    ? (metrics.idxBlksHit / indexTotal) * 100
    : 100;

  // Determine severity
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  if (tableHitRatio < 70 || indexHitRatio < 80) {
    severity = 'critical';
  } else if (tableHitRatio < 85 || indexHitRatio < 90) {
    severity = 'warning';
  }

  // Generate recommendation
  let recommendation = '';
  if (tableHitRatio < 70) {
    recommendation = 'Table has high disk I/O. Consider: (1) Adding index on frequently filtered columns, (2) Analyzing query plans, (3) Reviewing WHERE clauses for sargability';
  } else if (tableHitRatio < 85) {
    recommendation = 'Table cache could be improved. Review query patterns and consider additional indexes.';
  } else {
    recommendation = 'Good cache hit ratio. No immediate action needed.';
  }

  return {
    tableName: metrics.tableName,
    tableHitRatio: Math.round(tableHitRatio * 10) / 10,
    tableHitRatioPercent: `${tableHitRatio.toFixed(1)}%`,
    indexHitRatio: Math.round(indexHitRatio * 10) / 10,
    indexHitRatioPercent: `${indexHitRatio.toFixed(1)}%`,
    totalAccesses: tableTotal + indexTotal,
    totalCacheHits: metrics.heapBlksHit + metrics.idxBlksHit,
    severity,
    recommendation,
  };
}

// ============================================================================
// SEQUENTIAL SCAN ANALYSIS
// ============================================================================

interface TableScanMetrics {
  tableName: string;
  sequentialScans: number;
  indexScans: number;
  rowsSelectedBySeqScan: number;
  rowsSelectedByIndexScan: number;
}

export interface ScanAnalysisResult {
  tableName: string;
  seqScanRatio: number;         // 0-100, % of all scans that are sequential
  seqScanPercent: string;
  rowsPerSeqScan: number;       // Average rows returned per sequential scan
  recommendation: string;
  severity: 'healthy' | 'warning' | 'critical';
}

/**
 * Analyze sequential vs index scan patterns
 *
 * Sequential scans are not inherently bad - they're optimal for:
 * - Full table scans
 * - Returning large % of table
 * - Small tables
 *
 * Warning signs:
 * - Frequent seq scans with low row selectivity
 * - Table should have index but doesn't
 * - Seq scans instead of index range scans
 */
export function analyzeTableScans(metrics: TableScanMetrics): ScanAnalysisResult {
  const totalScans = metrics.sequentialScans + metrics.indexScans;
  const seqScanRatio = totalScans > 0
    ? (metrics.sequentialScans / totalScans) * 100
    : 0;

  const totalRowsSelected = metrics.rowsSelectedBySeqScan + metrics.rowsSelectedByIndexScan;
  const rowsPerSeqScan = metrics.sequentialScans > 0
    ? Math.round(metrics.rowsSelectedBySeqScan / metrics.sequentialScans)
    : 0;

  // Determine severity
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  let recommendation = 'Sequential scan usage is within normal parameters.';

  // High sequential scan ratio might indicate missing indexes
  if (seqScanRatio > 80 && metrics.sequentialScans > 100) {
    severity = 'warning';
    recommendation = `This table is frequently accessed with sequential scans (${seqScanRatio.toFixed(0)}%). Consider adding an index on frequently filtered columns.`;
  }

  // If sequential scans return few rows, it's inefficient
  if (rowsPerSeqScan < 10 && metrics.sequentialScans > 50) {
    severity = 'critical';
    recommendation = `Sequential scans are returning very few rows (avg ${rowsPerSeqScan}). An index would significantly improve query performance.`;
  }

  return {
    tableName: metrics.tableName,
    seqScanRatio: Math.round(seqScanRatio * 10) / 10,
    seqScanPercent: `${seqScanRatio.toFixed(1)}%`,
    rowsPerSeqScan,
    recommendation,
    severity,
  };
}
```

---

## Part 4: Query Performance Analysis

```typescript
// ============================================================================
// QUERY PERFORMANCE ANALYSIS
// ============================================================================

export enum QueryDuration {
  FAST = '< 10ms',
  NORMAL = '10ms - 100ms',
  SLOW = '100ms - 1s',
  VERY_SLOW = '1s - 10s',
  CRITICAL = '> 10s',
}

interface QueryMetricsData {
  queryId: string;
  queryText: string;
  calls: number;
  totalTime: number;           // milliseconds
  meanTime: number;            // milliseconds
  minTime: number;             // milliseconds
  maxTime: number;             // milliseconds
  stddev: number;              // standard deviation
  rows: number;                // average rows returned
}

export interface QueryAnalysisResult {
  queryId: string;
  queryText: string;
  durationCategory: QueryDuration;
  calls: number;
  totalTime: string;           // "1.23s"
  meanTime: string;            // "10.5ms"
  medianTime: string;          // estimated
  variance: number;            // stddev / mean (consistency)
  totalRowsReturned: number;
  severity: 'healthy' | 'warning' | 'critical';
  recommendations: string[];
}

export function analyzeQueryPerformance(metrics: QueryMetricsData): QueryAnalysisResult {
  let durationCategory: QueryDuration;
  if (metrics.meanTime < 10) {
    durationCategory = QueryDuration.FAST;
  } else if (metrics.meanTime < 100) {
    durationCategory = QueryDuration.NORMAL;
  } else if (metrics.meanTime < 1000) {
    durationCategory = QueryDuration.SLOW;
  } else if (metrics.meanTime < 10000) {
    durationCategory = QueryDuration.VERY_SLOW;
  } else {
    durationCategory = QueryDuration.CRITICAL;
  }

  // Calculate variance (consistency)
  const variance = metrics.stddev / metrics.meanTime;

  // Determine severity
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  if (metrics.meanTime > 5000) {
    severity = 'critical';
  } else if (metrics.meanTime > 500) {
    severity = 'warning';
  }

  // Generate recommendations
  const recommendations: string[] = [];

  if (metrics.meanTime > 500 && metrics.calls > 100) {
    recommendations.push('This query is executed frequently and is slow. Optimize with indexes or query rewrite.');
  }

  if (variance > 2) {
    recommendations.push('High performance variance detected. Some executions are much slower than others. Check for plan changes or resource contention.');
  }

  if (metrics.meanTime > 5000) {
    recommendations.push('Query is very slow (>5s). Investigate query plan, missing indexes, or data volume. Use EXPLAIN ANALYZE.');
  }

  if (metrics.calls > 1000 && metrics.meanTime > 10) {
    recommendations.push('Query is called very frequently. Consider caching or materialized views.');
  }

  return {
    queryId: metrics.queryId,
    queryText: metrics.queryText,
    durationCategory,
    calls: metrics.calls,
    totalTime: formatDuration(metrics.totalTime),
    meanTime: formatDuration(metrics.meanTime),
    medianTime: formatDuration(metrics.meanTime), // Approximation
    variance: Math.round(variance * 100) / 100,
    totalRowsReturned: metrics.rows * metrics.calls,
    severity,
    recommendations,
  };
}

function formatDuration(milliseconds: number): string {
  if (milliseconds < 1) return `${(milliseconds * 1000).toFixed(0)}µs`;
  if (milliseconds < 1000) return `${milliseconds.toFixed(2)}ms`;
  return `${(milliseconds / 1000).toFixed(2)}s`;
}
```

---

## Part 5: Connection Pool Analysis

```typescript
// ============================================================================
// CONNECTION POOL ANALYSIS
// ============================================================================

interface ConnectionPoolMetrics {
  maxConnections: number;
  currentConnections: number;
  activeConnections: number;
  idleConnections: number;
  idleInTransactionConnections: number;
  supernuserConnections: number;
  connectionCreationRate: number;  // connections/sec
  averageConnectionAge: number;    // seconds
}

export interface ConnectionPoolAnalysis {
  utilizationPercent: number;
  utilizationCategory: 'low' | 'medium' | 'high' | 'critical';
  activePercent: number;
  idlePercent: number;
  idleInTransactionPercent: number;
  availableConnections: number;
  availablePercent: number;
  averageConnectionAgeMins: string;
  severity: 'healthy' | 'warning' | 'critical';
  recommendations: string[];
}

export function analyzeConnectionPool(metrics: ConnectionPoolMetrics): ConnectionPoolAnalysis {
  const utilizationPercent = (metrics.currentConnections / metrics.maxConnections) * 100;
  const activePercent = (metrics.activeConnections / metrics.currentConnections) * 100 || 0;
  const idlePercent = (metrics.idleConnections / metrics.currentConnections) * 100 || 0;
  const idleInTxnPercent = (metrics.idleInTransactionConnections / metrics.currentConnections) * 100 || 0;
  const availableConnections = metrics.maxConnections - metrics.currentConnections;
  const availablePercent = (availableConnections / metrics.maxConnections) * 100;

  // Determine utilization category
  let utilizationCategory: 'low' | 'medium' | 'high' | 'critical' = 'low';
  if (utilizationPercent < 50) {
    utilizationCategory = 'low';
  } else if (utilizationPercent < 75) {
    utilizationCategory = 'medium';
  } else if (utilizationPercent < 90) {
    utilizationCategory = 'high';
  } else {
    utilizationCategory = 'critical';
  }

  // Determine severity
  let severity: 'healthy' | 'warning' | 'critical' = 'healthy';
  if (utilizationPercent > 90 || idleInTxnPercent > 20) {
    severity = 'critical';
  } else if (utilizationPercent > 75 || idleInTxnPercent > 10) {
    severity = 'warning';
  }

  // Generate recommendations
  const recommendations: string[] = [];

  if (utilizationPercent > 85) {
    recommendations.push(`Connection pool is highly utilized (${utilizationPercent.toFixed(0)}%). Consider increasing max_connections or optimizing application connection pooling.`);
  }

  if (idleInTxnPercent > 15) {
    recommendations.push(`High percentage of idle-in-transaction connections (${idleInTxnPercent.toFixed(0)}%). This indicates connection leaks in the application. Add transaction timeouts.`);
  }

  if (metrics.connectionCreationRate > 10) {
    recommendations.push(`High connection creation rate (${metrics.connectionCreationRate.toFixed(1)}/sec). Enable connection pooling in application to reduce overhead.`);
  }

  if (availablePercent < 10 && utilizationPercent > 90) {
    recommendations.push('Critical: Very few available connections remain. New connection requests will fail. Immediate action needed.');
  }

  return {
    utilizationPercent: Math.round(utilizationPercent * 10) / 10,
    utilizationCategory,
    activePercent: Math.round(activePercent * 10) / 10,
    idlePercent: Math.round(idlePercent * 10) / 10,
    idleInTransactionPercent: Math.round(idleInTxnPercent * 10) / 10,
    availableConnections,
    availablePercent: Math.round(availablePercent * 10) / 10,
    averageConnectionAgeMins: formatDuration(metrics.averageConnectionAge * 1000),
    severity,
    recommendations,
  };
}
```

---

## Part 6: Trend Analysis & Forecasting

```typescript
// ============================================================================
// TREND ANALYSIS
// ============================================================================

interface TimeSeriesPoint {
  timestamp: Date;
  value: number;
}

export interface TrendAnalysis {
  average: number;
  min: number;
  max: number;
  stddev: number;
  trend: 'increasing' | 'decreasing' | 'stable';
  trendStrength: number;  // 0-100
  forecast7d: number;    // predicted value in 7 days
  changePercent: number; // % change from start to end
}

/**
 * Analyze trend in time series data
 *
 * Uses linear regression to determine trend direction and strength
 */
export function analyzeTrend(points: TimeSeriesPoint[]): TrendAnalysis {
  if (points.length < 2) {
    return {
      average: points[0]?.value || 0,
      min: points[0]?.value || 0,
      max: points[0]?.value || 0,
      stddev: 0,
      trend: 'stable',
      trendStrength: 0,
      forecast7d: points[0]?.value || 0,
      changePercent: 0,
    };
  }

  // Calculate statistics
  const values = points.map(p => p.value);
  const average = values.reduce((a, b) => a + b, 0) / values.length;
  const min = Math.min(...values);
  const max = Math.max(...values);

  // Calculate standard deviation
  const variance = values.reduce((sum, val) => sum + Math.pow(val - average, 2), 0) / values.length;
  const stddev = Math.sqrt(variance);

  // Linear regression for trend
  const n = points.length;
  const xs = points.map((_, i) => i);
  const ys = values;

  const xMean = xs.reduce((a, b) => a + b, 0) / n;
  const yMean = ys.reduce((a, b) => a + b, 0) / n;

  const numerator = xs.reduce((sum, x, i) => sum + (x - xMean) * (ys[i] - yMean), 0);
  const denominator = xs.reduce((sum, x) => sum + Math.pow(x - xMean, 2), 0);

  const slope = denominator !== 0 ? numerator / denominator : 0;
  const intercept = yMean - slope * xMean;

  // Determine trend
  let trend: 'increasing' | 'decreasing' | 'stable' = 'stable';
  let trendStrength = 0;

  if (Math.abs(slope) > stddev * 0.1) {
    trend = slope > 0 ? 'increasing' : 'decreasing';
    trendStrength = Math.min(Math.abs(slope) / (max - min + 1) * 100, 100);
  }

  // Forecast 7 days ahead
  const forecast7d = intercept + slope * (n + 7);

  // Calculate percent change
  const changePercent = values.length > 0
    ? ((values[values.length - 1] - values[0]) / values[0]) * 100
    : 0;

  return {
    average: Math.round(average * 100) / 100,
    min,
    max,
    stddev: Math.round(stddev * 100) / 100,
    trend,
    trendStrength: Math.round(trendStrength * 10) / 10,
    forecast7d: Math.round(forecast7d * 100) / 100,
    changePercent: Math.round(changePercent * 10) / 10,
  };
}
```

---

## Part 7: Utility Functions for UI

```typescript
// frontend/src/utils/formatting.ts

export function formatBytes(bytes: number): string {
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  if (bytes === 0) return '0 B';

  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

export function formatPercentage(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatDuration(milliseconds: number): string {
  if (milliseconds < 1) return `${(milliseconds * 1000).toFixed(0)}µs`;
  if (milliseconds < 1000) return `${milliseconds.toFixed(2)}ms`;
  if (milliseconds < 60000) return `${(milliseconds / 1000).toFixed(2)}s`;
  return `${(milliseconds / 60000).toFixed(2)}m`;
}

export function formatNumber(num: number, decimals = 0): string {
  return num.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

export function getHealthColor(score: number): string {
  if (score >= 80) return 'text-pg-success';
  if (score >= 60) return 'text-pg-warning';
  return 'text-pg-danger';
}

export function getHealthBgColor(score: number): string {
  if (score >= 80) return 'bg-pg-success/10';
  if (score >= 60) return 'bg-pg-warning/10';
  return 'bg-pg-danger/10';
}
```

---

**All calculation functions ready for implementation!**

These can be directly imported and used in React components.

Generated: March 3, 2026
