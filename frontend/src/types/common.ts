export interface ApiError {
  status: number
  message: string
  code?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type LogLevel = 'DEBUG' | 'INFO' | 'NOTICE' | 'WARNING' | 'ERROR' | 'FATAL' | 'PANIC'

export type CollectorStatus = 'OK' | 'SLOW' | 'DOWN'

export interface Collector {
  id: string
  hostname: string
  environment: string
  group?: string
  description?: string
  status: CollectorStatus
  last_heartbeat: string
  created_at: string
  updated_at: string
}

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

export interface NotificationPreferences {
  email_enabled: boolean
  browser_enabled: boolean
  sound_enabled: boolean
}
