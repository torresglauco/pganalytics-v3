export interface TimeSeriesPoint {
  timestamp: Date;
  value: number;
  collector_id: string;
}

export interface MetricsSummary {
  lockCount: number;
  bloatRatio: number;
  cacheHitRatio: number;
  connectionCount: number;
  replicationLag: number;
}

export interface QueryMetrics {
  query_id: string;
  query_text: string;
  calls: number;
  total_time: number;
  mean_time: number;
  rows: number;
  100pct_time: number;
  database: string;
}

export interface LockMetrics {
  lock_id: string;
  blocking_pid: number;
  blocked_pid: number;
  lock_type: string;
  granted: boolean;
  duration_ms: number;
  blocking_query?: string;
}

export interface BloatMetrics {
  table_name: string;
  dead_tuples: number;
  live_tuples: number;
  bloat_ratio: number;
  reclaimable_bytes: number;
  last_vacuum: Date;
}

export interface CacheMetrics {
  table_name: string;
  heap_blks_hit: number;
  heap_blks_read: number;
  hit_ratio: number;
}

export interface ConnectionMetrics {
  pid: number;
  usename: string;
  state: 'active' | 'idle' | 'idle in transaction';
  query: string;
  query_start: Date;
  duration_ms: number;
}

export interface HealthScore {
  overall: number;
  lock_health: number;
  bloat_health: number;
  query_health: number;
  cache_health: number;
  connection_health: number;
  replication_health: number;
  timestamp: Date;
}
