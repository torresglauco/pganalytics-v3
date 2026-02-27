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

export interface SignupRequest {
  username: string
  email: string
  password: string
  full_name: string
}

export interface User {
  id: number
  username: string
  email: string
  full_name: string
  role: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  refresh_token: string
  expires_at: string
  user: User
}
