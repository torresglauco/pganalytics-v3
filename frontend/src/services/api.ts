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

class ApiClient {
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
