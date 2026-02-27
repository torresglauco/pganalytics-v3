import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi } from 'vitest'
import type { Collector, User, AuthResponse } from '../types'

// Mock API Client Factory
export const createMockApiClient = (overrides = {}) => ({
  login: vi.fn(),
  signup: vi.fn(),
  logout: vi.fn(),
  registerCollector: vi.fn(),
  listCollectors: vi.fn(),
  getCollector: vi.fn(),
  deleteCollector: vi.fn(),
  testConnection: vi.fn(),
  getCurrentUser: vi.fn(),
  getToken: vi.fn(),
  getBaseURL: vi.fn(),
  isAuthenticated: vi.fn(),
  handleError: vi.fn(),
  ...overrides,
})

// Test Data Generators
export const mockUser: User = {
  id: 1,
  username: 'testuser',
  email: 'test@example.com',
  full_name: 'Test User',
  role: 'user',
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

export const mockAdminUser: User = {
  id: 2,
  username: 'admin',
  email: 'admin@example.com',
  full_name: 'Admin User',
  role: 'admin',
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

export const mockAuthResponse: AuthResponse = {
  token: 'mock-token-12345',
  refresh_token: 'mock-refresh-token-12345',
  expires_at: '2024-12-31T00:00:00Z',
  user: mockUser,
}

export const mockCollector: Collector = {
  id: 'collector-1',
  hostname: 'localhost',
  status: 'active',
  created_at: '2024-01-01T00:00:00Z',
  last_heartbeat: '2024-01-01T12:00:00Z',
  metrics_count: 150,
  uptime: 99.9,
}

export const mockCollector2: Collector = {
  id: 'collector-2',
  hostname: 'prod.example.com',
  status: 'active',
  created_at: '2024-01-02T00:00:00Z',
  last_heartbeat: '2024-01-02T12:00:00Z',
  metrics_count: 200,
  uptime: 99.95,
}

// Custom Render with Router
interface AllTheProvidersProps {
  children: React.ReactNode
}

const AllTheProviders: React.FC<AllTheProvidersProps> = ({ children }) => {
  return React.createElement(BrowserRouter, null, children)
}

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options })

export * from '@testing-library/react'
export { customRender as render }
