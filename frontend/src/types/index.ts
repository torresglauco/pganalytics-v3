export interface Collector {
  id: string
  hostname: string
  status: 'active' | 'inactive' | 'error'
  created_at: string
  last_heartbeat?: string
  metrics_count?: number
  uptime?: number
}

export interface CollectorRegisterRequest {
  hostname: string
  environment?: string
  group?: string
  description?: string
}

export interface CollectorRegisterResponse {
  collector_id: string
  status: string
  token: string
  created_at: string
}

export interface ApiError {
  message: string
  details?: string
  status_code?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}
