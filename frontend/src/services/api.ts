import axios, { AxiosInstance, AxiosError } from 'axios'
import type {
  Collector,
  CollectorRegisterRequest,
  CollectorRegisterResponse,
  PaginatedResponse,
  ApiError,
  SignupRequest,
  AuthResponse,
  User
} from '../types'

export class ApiClient {
  private client: AxiosInstance

  constructor(baseURL: string = '/api/v1') {
    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Add request interceptor for auth token
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('auth_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('auth_token')
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
  }

  async registerCollector(
    data: CollectorRegisterRequest,
    registrationSecret: string
  ): Promise<CollectorRegisterResponse> {
    try {
      const response = await this.client.post<CollectorRegisterResponse>(
        '/collectors/register',
        data,
        {
          headers: {
            'X-Registration-Secret': registrationSecret,
          },
        }
      )
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async listCollectors(
    page: number = 1,
    pageSize: number = 20
  ): Promise<PaginatedResponse<Collector>> {
    try {
      const response = await this.client.get<PaginatedResponse<Collector>>(
        '/collectors',
        {
          params: { page, page_size: pageSize },
        }
      )
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getCollector(id: string): Promise<Collector> {
    try {
      const response = await this.client.get<Collector>(`/collectors/${id}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteCollector(id: string): Promise<void> {
    try {
      await this.client.delete(`/collectors/${id}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async testConnection(hostname: string, port: number = 5432): Promise<boolean> {
    try {
      const response = await this.client.post('/collectors/test-connection', {
        hostname,
        port,
      })
      return response.status === 200
    } catch (error) {
      return false
    }
  }

  async signup(data: SignupRequest): Promise<AuthResponse> {
    try {
      const response = await this.client.post<AuthResponse>('/auth/signup', data)
      // Store auth token
      if (response.data.token) {
        localStorage.setItem('auth_token', response.data.token)
        localStorage.setItem('refresh_token', response.data.refresh_token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
      }
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async login(username: string, password: string): Promise<AuthResponse> {
    try {
      const response = await this.client.post<AuthResponse>('/auth/login', {
        username,
        password,
      })
      // Store auth token
      if (response.data.token) {
        localStorage.setItem('auth_token', response.data.token)
        localStorage.setItem('refresh_token', response.data.refresh_token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
      }
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async logout(): Promise<void> {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
  }

  getCurrentUser(): User | null {
    const user = localStorage.getItem('user')
    return user ? JSON.parse(user) : null
  }

  getToken(): string | null {
    return localStorage.getItem('auth_token')
  }

  getBaseURL(): string {
    return this.client.defaults.baseURL || '/api/v1'
  }

  isAuthenticated(): boolean {
    return !!localStorage.getItem('auth_token')
  }

  // Logs endpoints
  async getLogs(params?: {
    page?: number
    page_size?: number
    level?: string
    search?: string
    instance_id?: string
    from_time?: string
    to_time?: string
  }): Promise<any> {
    try {
      const response = await this.client.get('/logs', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getLogDetails(logId: string): Promise<any> {
    try {
      const response = await this.client.get(`/logs/${logId}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Metrics endpoints
  async getMetrics(params?: {
    instance_id?: string
    time_range?: string
  }): Promise<any> {
    try {
      const response = await this.client.get('/metrics', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getErrorTrend(params?: {
    instance_id?: string
    hours?: number
  }): Promise<any> {
    try {
      const response = await this.client.get('/metrics/error-trend', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getLogDistribution(params?: {
    instance_id?: string
    time_range?: string
  }): Promise<any> {
    try {
      const response = await this.client.get('/metrics/log-distribution', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Alert endpoints
  async getAlerts(params?: {
    page?: number
    page_size?: number
    status?: string
  }): Promise<any> {
    try {
      const response = await this.client.get('/alerts', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createAlert(data: any): Promise<any> {
    try {
      const response = await this.client.post('/alerts', data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async updateAlert(alertId: string, data: any): Promise<any> {
    try {
      const response = await this.client.put(`/alerts/${alertId}`, data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteAlert(alertId: string): Promise<void> {
    try {
      await this.client.delete(`/alerts/${alertId}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async testAlert(alertId: string): Promise<any> {
    try {
      const response = await this.client.post(`/alerts/${alertId}/test`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Channel endpoints
  async getChannels(): Promise<any> {
    try {
      const response = await this.client.get('/channels')
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createChannel(data: any): Promise<any> {
    try {
      const response = await this.client.post('/channels', data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async updateChannel(channelId: string, data: any): Promise<any> {
    try {
      const response = await this.client.put(`/channels/${channelId}`, data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteChannel(channelId: string): Promise<void> {
    try {
      await this.client.delete(`/channels/${channelId}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async testChannel(channelId: string): Promise<any> {
    try {
      const response = await this.client.post(`/channels/${channelId}/test`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Alert Rules endpoints
  async validateAlertCondition(condition: any): Promise<any> {
    try {
      const response = await this.client.post('/alert-rules/validate', condition)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createAlertRule(data: any): Promise<any> {
    try {
      const response = await this.client.post('/alert-rules', data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getAlertRules(params?: {
    page?: number
    page_size?: number
    status?: string
  }): Promise<any> {
    try {
      const response = await this.client.get('/alert-rules', { params })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getAlertRule(ruleId: string): Promise<any> {
    try {
      const response = await this.client.get(`/alert-rules/${ruleId}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async updateAlertRule(ruleId: string, data: any): Promise<any> {
    try {
      const response = await this.client.put(`/alert-rules/${ruleId}`, data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteAlertRule(ruleId: string): Promise<void> {
    try {
      await this.client.delete(`/alert-rules/${ruleId}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // User Management endpoints
  async listUsers(page: number = 1, pageSize: number = 20): Promise<any> {
    try {
      const response = await this.client.get('/users', {
        params: { page, page_size: pageSize },
      })
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createUser(data: any): Promise<any> {
    try {
      const response = await this.client.post('/users', data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getUser(id: string): Promise<any> {
    try {
      const response = await this.client.get(`/users/${id}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async updateUser(id: string, data: any): Promise<any> {
    try {
      const response = await this.client.put(`/users/${id}`, data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteUser(id: string): Promise<void> {
    try {
      await this.client.delete(`/users/${id}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async resetUserPassword(id: string): Promise<any> {
    try {
      const response = await this.client.post(`/users/${id}/reset-password`, {})
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // API Token methods
  async listApiTokens(): Promise<any[]> {
    try {
      const response = await this.client.get('/registration-secrets')
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createApiToken(data: { name: string; expires_at?: string }): Promise<any> {
    try {
      const response = await this.client.post('/registration-secrets', data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async getApiToken(id: string): Promise<any> {
    try {
      const response = await this.client.get(`/registration-secrets/${id}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async updateApiToken(id: string, data: any): Promise<any> {
    try {
      const response = await this.client.put(`/registration-secrets/${id}`, data)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async deleteApiToken(id: string): Promise<void> {
    try {
      await this.client.delete(`/registration-secrets/${id}`)
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Index Advisor endpoints
  async getIndexRecommendations(databaseId: string, limit: number = 20): Promise<any> {
    try {
      const response = await this.client.get(
        `/index-advisor/database/${databaseId}/recommendations`,
        { params: { limit } }
      )
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  async createIndexFromRecommendation(recommendationId: number): Promise<any> {
    try {
      const response = await this.client.post(
        `/index-advisor/recommendation/${recommendationId}/create`
      )
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Query Performance endpoints
  async getQueryPerformance(databaseId: string): Promise<any> {
    try {
      const response = await this.client.get(`/query-performance/database/${databaseId}`)
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  // Vacuum Advisor endpoints
  async getVacuumRecommendations(databaseId: number, limit: number = 20): Promise<any> {
    try {
      const response = await this.client.get(
        `/vacuum-advisor/database/${databaseId}/recommendations`,
        { params: { limit } }
      )
      return response.data
    } catch (error) {
      throw this.handleError(error)
    }
  }

  private handleError(error: unknown): ApiError {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<ApiError>
      return {
        message: axiosError.response?.data?.message || error.message,
        details: axiosError.response?.data?.details,
        status_code: axiosError.response?.status,
      }
    }
    return {
      message: 'An unexpected error occurred',
    }
  }
}

export const apiClient = new ApiClient()
