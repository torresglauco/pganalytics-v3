/**
 * Health Score Calculations for pgAnalytics v3
 * Calculates component-specific and overall health scores
 */

export interface HealthMetrics {
  lockHealth: number;
  bloatHealth: number;
  queryHealth: number;
  cacheHealth: number;
  connectionHealth: number;
  replicationHealth: number;
}

export function calculateOverallHealth(metrics: HealthMetrics): number {
  const weights = {
    lockHealth: 0.15,
    bloatHealth: 0.2,
    queryHealth: 0.15,
    cacheHealth: 0.2,
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
// LOCK HEALTH (0-100)
// ============================================================================

interface LockMetricsData {
  activeLocksCount: number;
  blockedTransactions: number;
  maxWaitTime: number;
}

export function calculateLockHealth(metrics: LockMetricsData): number {
  let score = 100;

  score -= Math.min(metrics.blockedTransactions * 10, 50);

  if (metrics.maxWaitTime > 300) {
    score -= 30;
  } else if (metrics.maxWaitTime > 60) {
    score -= 15;
  } else if (metrics.maxWaitTime > 10) {
    score -= 5;
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// BLOAT HEALTH (0-100)
// ============================================================================

interface BloatMetricsData {
  tableCount: number;
  bloatedTableCount: number;
  severeBloatCount: number;
  avgBloatRatio: number;
}

export function calculateBloatHealth(metrics: BloatMetricsData): number {
  let score = 100;

  if (metrics.avgBloatRatio > 50) {
    score -= 40;
  } else if (metrics.avgBloatRatio > 30) {
    score -= 25;
  } else if (metrics.avgBloatRatio > 20) {
    score -= 15;
  } else if (metrics.avgBloatRatio > 5) {
    score -= 5;
  }

  const severePenalty = Math.min(metrics.severeBloatCount * 15, 30);
  score -= severePenalty;

  const bloatRatio = metrics.tableCount > 0
    ? metrics.bloatedTableCount / metrics.tableCount
    : 0;

  if (bloatRatio > 0.1) {
    score -= Math.min(bloatRatio * 20, 20);
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// CACHE HEALTH (0-100)
// ============================================================================

interface CacheHealthMetricsData {
  overallHitRatio: number;
  tableHitRatio: number;
  indexHitRatio: number;
  lowHitRatioTableCount: number;
  lowHitRatioIndexCount: number;
}

export function calculateCacheHealth(metrics: CacheHealthMetricsData): number {
  let score = 100;

  if (metrics.overallHitRatio < 70) {
    score -= (70 - metrics.overallHitRatio) * 0.4;
  } else if (metrics.overallHitRatio < 90) {
    score -= (90 - metrics.overallHitRatio) * 0.15;
  } else if (metrics.overallHitRatio < 99) {
    score -= (99 - metrics.overallHitRatio) * 0.05;
  }

  if (metrics.tableHitRatio < 80) {
    score -= (80 - metrics.tableHitRatio) * 0.3;
  } else if (metrics.tableHitRatio < 95) {
    score -= (95 - metrics.tableHitRatio) * 0.1;
  }

  if (metrics.indexHitRatio < 90) {
    score -= (90 - metrics.indexHitRatio) * 0.3;
  } else if (metrics.indexHitRatio < 98) {
    score -= (98 - metrics.indexHitRatio) * 0.05;
  }

  score -= Math.min(metrics.lowHitRatioTableCount * 2, 15);
  score -= Math.min(metrics.lowHitRatioIndexCount * 1, 10);

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// CONNECTION HEALTH (0-100)
// ============================================================================

interface ConnectionHealthMetricsData {
  connectionUsagePercent: number;
  idleInTransactionConnections: number;
  idleOverDuration: number;
}

export function calculateConnectionHealth(metrics: ConnectionHealthMetricsData): number {
  let score = 100;

  if (metrics.connectionUsagePercent > 95) {
    score -= 40;
  } else if (metrics.connectionUsagePercent > 90) {
    score -= 30;
  } else if (metrics.connectionUsagePercent > 80) {
    score -= 15;
  } else if (metrics.connectionUsagePercent > 70) {
    score -= 5;
  }

  if (metrics.idleInTransactionConnections > 5) {
    score -= Math.min(metrics.idleInTransactionConnections * 5, 30);
  }

  if (metrics.idleOverDuration > 10) {
    score -= Math.min(metrics.idleOverDuration * 2, 15);
  }

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// QUERY HEALTH (0-100)
// ============================================================================

interface QueryHealthMetricsData {
  slowQueryCount: number;
  verySlowQueryCount: number;
  avgExecutionTime: number;
  highIOQueryCount: number;
}

export function calculateQueryHealth(metrics: QueryHealthMetricsData): number {
  let score = 100;

  score -= Math.min(metrics.verySlowQueryCount * 20, 40);
  score -= Math.min(metrics.slowQueryCount * 5, 30);

  if (metrics.avgExecutionTime > 2000) {
    score -= 15;
  } else if (metrics.avgExecutionTime > 500) {
    score -= 10;
  } else if (metrics.avgExecutionTime > 100) {
    score -= 5;
  }

  score -= Math.min(metrics.highIOQueryCount * 3, 20);

  return Math.max(0, Math.min(100, score));
}

// ============================================================================
// REPLICATION HEALTH (0-100)
// ============================================================================

interface ReplicationHealthMetricsData {
  isReplicating: boolean;
  replicationLagBytes: number;
  connectedStandbyCount: number;
  expectedStandbyCount: number;
  walArchiveStatus: 'healthy' | 'warning' | 'critical';
}

export function calculateReplicationHealth(metrics: ReplicationHealthMetricsData): number {
  if (!metrics.isReplicating) {
    return 100;
  }

  let score = 100;

  if (metrics.replicationLagBytes > 100 * 1024 * 1024) {
    score -= 40;
  } else if (metrics.replicationLagBytes > 10 * 1024 * 1024) {
    score -= 20;
  } else if (metrics.replicationLagBytes > 1 * 1024 * 1024) {
    score -= 10;
  } else if (metrics.replicationLagBytes > 0) {
    score -= 3;
  }

  const disconnected = metrics.expectedStandbyCount - metrics.connectedStandbyCount;
  if (disconnected > 0) {
    score -= disconnected * 20;
  }

  if (metrics.walArchiveStatus === 'critical') {
    score -= 30;
  } else if (metrics.walArchiveStatus === 'warning') {
    score -= 15;
  }

  return Math.max(0, Math.min(100, score));
}
