/**
 * Health Score Calculations for pgAnalytics Dashboard
 */

export interface HealthMetrics {
  lockHealth: number;
  bloatHealth: number;
  queryHealth: number;
  cacheHealth: number;
  connectionHealth: number;
  replicationHealth: number;
}

/**
 * Calculate overall database health score (0-100)
 */
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

/**
 * Get health color based on score
 */
export function getHealthColor(score: number): string {
  if (score >= 80) return 'text-pg-success';
  if (score >= 60) return 'text-pg-warning';
  return 'text-pg-danger';
}

/**
 * Get health background color
 */
export function getHealthBgColor(score: number): string {
  if (score >= 80) return 'bg-pg-success/10 border-pg-success/20';
  if (score >= 60) return 'bg-pg-warning/10 border-pg-warning/20';
  return 'bg-pg-danger/10 border-pg-danger/20';
}

/**
 * Get health status label
 */
export function getHealthStatus(score: number): 'healthy' | 'warning' | 'critical' {
  if (score >= 80) return 'healthy';
  if (score >= 60) return 'warning';
  return 'critical';
}

/**
 * Calculate bloat health (0-100)
 * Perfect: < 5% bloat
 * Healthy: < 20% bloat
 * Warning: 20-40% bloat
 * Critical: > 40% bloat
 */
export function calculateBloatHealth(avgBloatRatio: number, severeBloatCount: number): number {
  let score = 100;

  if (avgBloatRatio > 50) {
    score -= 40;
  } else if (avgBloatRatio > 30) {
    score -= 25;
  } else if (avgBloatRatio > 20) {
    score -= 15;
  } else if (avgBloatRatio > 5) {
    score -= 5;
  }

  score -= Math.min(severeBloatCount * 15, 30);

  return Math.max(0, Math.min(100, score));
}

/**
 * Calculate cache health (0-100)
 * Excellent: > 99%
 * Healthy: > 90%
 * Warning: > 70%
 * Critical: < 70%
 */
export function calculateCacheHealth(
  overallHitRatio: number,
  tableHitRatio: number,
  indexHitRatio: number
): number {
  let score = 100;

  // Overall hit ratio (40% weight)
  if (overallHitRatio < 70) {
    score -= (70 - overallHitRatio) * 0.4;
  } else if (overallHitRatio < 90) {
    score -= (90 - overallHitRatio) * 0.15;
  } else if (overallHitRatio < 99) {
    score -= (99 - overallHitRatio) * 0.05;
  }

  // Table hit ratio (30% weight)
  if (tableHitRatio < 80) {
    score -= (80 - tableHitRatio) * 0.3;
  } else if (tableHitRatio < 95) {
    score -= (95 - tableHitRatio) * 0.1;
  }

  // Index hit ratio (30% weight)
  if (indexHitRatio < 90) {
    score -= (90 - indexHitRatio) * 0.3;
  } else if (indexHitRatio < 98) {
    score -= (98 - indexHitRatio) * 0.05;
  }

  return Math.max(0, Math.min(100, score));
}

/**
 * Calculate connection health (0-100)
 * Perfect: < 50% utilization
 * Healthy: 50-80% utilization
 * Warning: 80-95% utilization
 * Critical: > 95% utilization
 */
export function calculateConnectionHealth(
  connectionUsagePercent: number,
  idleInTransactionCount: number
): number {
  let score = 100;

  if (connectionUsagePercent > 95) {
    score -= 40;
  } else if (connectionUsagePercent > 90) {
    score -= 30;
  } else if (connectionUsagePercent > 80) {
    score -= 15;
  } else if (connectionUsagePercent > 70) {
    score -= 5;
  }

  if (idleInTransactionCount > 5) {
    score -= Math.min(idleInTransactionCount * 5, 30);
  }

  return Math.max(0, Math.min(100, score));
}

/**
 * Calculate lock health (0-100)
 * Perfect: No locks
 * Healthy: Some locks, no blocking
 * Warning: Blocking detected < 5 min
 * Critical: Blocking > 5 min
 */
export function calculateLockHealth(
  blockedTransactions: number,
  maxWaitTimeSeconds: number
): number {
  let score = 100;

  score -= Math.min(blockedTransactions * 10, 50);

  if (maxWaitTimeSeconds > 300) {
    score -= 30;
  } else if (maxWaitTimeSeconds > 60) {
    score -= 15;
  } else if (maxWaitTimeSeconds > 10) {
    score -= 5;
  }

  return Math.max(0, Math.min(100, score));
}

/**
 * Calculate query performance health (0-100)
 */
export function calculateQueryHealth(
  avgExecutionTimeMs: number,
  slowQueryCount: number,
  verySlowQueryCount: number
): number {
  let score = 100;

  score -= Math.min(verySlowQueryCount * 20, 40);
  score -= Math.min(slowQueryCount * 5, 30);

  if (avgExecutionTimeMs > 2000) {
    score -= 15;
  } else if (avgExecutionTimeMs > 500) {
    score -= 10;
  } else if (avgExecutionTimeMs > 100) {
    score -= 5;
  }

  return Math.max(0, Math.min(100, score));
}

/**
 * Analyze trend direction based on historical data
 */
export function analyzeTrend(
  values: number[]
): { trend: 'increasing' | 'decreasing' | 'stable'; strength: number } {
  if (values.length < 2) {
    return { trend: 'stable', strength: 0 };
  }

  const n = values.length;
  const xs = Array.from({ length: n }, (_, i) => i);
  const ys = values;

  const xMean = xs.reduce((a, b) => a + b, 0) / n;
  const yMean = ys.reduce((a, b) => a + b, 0) / n;

  const numerator = xs.reduce((sum, x, i) => sum + (x - xMean) * (ys[i] - yMean), 0);
  const denominator = xs.reduce((sum, x) => sum + Math.pow(x - xMean, 2), 0);

  const slope = denominator !== 0 ? numerator / denominator : 0;
  const minVal = Math.min(...ys);
  const maxVal = Math.max(...ys);
  const range = maxVal - minVal || 1;

  const strength = Math.min(Math.abs(slope / range) * 100, 100);

  return {
    trend: slope > 0.1 ? 'increasing' : slope < -0.1 ? 'decreasing' : 'stable',
    strength: Math.round(strength),
  };
}
