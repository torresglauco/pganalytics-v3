export interface DashboardMetric {
  label: string
  value: string | number
  unit?: string
  trend?: {
    direction: 'up' | 'down' | 'neutral'
    percentage: number
    period: string
  }
  color?: 'success' | 'warning' | 'error' | 'info'
}

export interface ActivityEvent {
  id: string
  type: 'error' | 'warning' | 'info' | 'success'
  title: string
  description: string
  timestamp: string
  action_url?: string
}

export interface DrillDownOption {
  id: string
  icon: string
  title: string
  description: string
  link: string
}

export interface CollectorStatusRow {
  id: string
  hostname: string
  environment: string
  status: 'OK' | 'SLOW' | 'DOWN'
  last_heartbeat: string
  error_count_24h: number
}
