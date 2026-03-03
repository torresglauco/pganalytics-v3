export interface ApiResponse<T> {
  data?: T;
  error?: string;
  status: number;
  timestamp: Date;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface Collector {
  id: string;
  hostname: string;
  environment?: string;
  group?: string;
  description?: string;
  status: 'ok' | 'slow' | 'down';
  created_at: Date;
  last_heartbeat: Date;
  metrics_count?: number;
}

export interface User {
  id: string;
  username: string;
  email?: string;
  role: 'admin' | 'user' | 'viewer';
  created_at: Date;
  last_login?: Date;
}

export interface RegistrationSecret {
  secret: string;
  created_at: Date;
  expires_at?: Date;
}

export interface ApiError {
  status: number;
  message: string;
  code?: string;
  details?: unknown;
}
