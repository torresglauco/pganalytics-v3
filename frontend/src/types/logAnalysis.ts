export interface LogEntry {
  id: number
  database_id: number
  log_timestamp: string
  category: string
  severity: 'INFO' | 'WARNING' | 'ERROR' | 'FATAL'
  message: string
  duration?: number
  table_affected?: string
}

export interface LogPattern {
  id: number
  pattern_name: string
  frequency: number
  severity_avg: number
  last_seen: string
}

export interface LogAnomaly {
  id: number
  pattern_id: number
  anomaly_timestamp: string
  anomaly_score: number
  deviation_from_baseline: number
}
