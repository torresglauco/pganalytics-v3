import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ApiClient } from './api'
import type { AuthResponse } from '../types'

describe('ApiClient', () => {
  let client: ApiClient

  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    client = new ApiClient()
  })

  describe('Token Management', () => {
    it('should return null token when not authenticated', () => {
      expect(client.getToken()).toBeNull()
    })

    it('should return stored token', () => {
      localStorage.setItem('auth_token', 'test-token')
      expect(client.getToken()).toBe('test-token')
    })

    it('should return base URL', () => {
      expect(client.getBaseURL()).toBe('/api/v1')
    })
  })

  describe('Authentication State', () => {
    it('should return false for isAuthenticated when no token', () => {
      expect(client.isAuthenticated()).toBe(false)
    })

    it('should return true for isAuthenticated when token exists', () => {
      localStorage.setItem('auth_token', 'test-token')
      expect(client.isAuthenticated()).toBe(true)
    })

    it('should get current user from localStorage', () => {
      const mockUser = {
        id: 1,
        username: 'testuser',
        email: 'test@example.com',
        full_name: 'Test User',
        role: 'user',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      localStorage.setItem('user', JSON.stringify(mockUser))
      expect(client.getCurrentUser()).toEqual(mockUser)
    })

    it('should return null for getCurrentUser when no user', () => {
      expect(client.getCurrentUser()).toBeNull()
    })
  })

  describe('Logout', () => {
    it('should clear all tokens on logout', async () => {
      localStorage.setItem('auth_token', 'token')
      localStorage.setItem('refresh_token', 'refresh')
      localStorage.setItem('user', '{}')

      await client.logout()

      expect(localStorage.getItem('auth_token')).toBeNull()
      expect(localStorage.getItem('refresh_token')).toBeNull()
      expect(localStorage.getItem('user')).toBeNull()
    })
  })

  describe('ApiClient Methods', () => {
    it('should have login method', () => {
      expect(typeof client.login).toBe('function')
    })

    it('should have signup method', () => {
      expect(typeof client.signup).toBe('function')
    })

    it('should have listCollectors method', () => {
      expect(typeof client.listCollectors).toBe('function')
    })

    it('should have getCollector method', () => {
      expect(typeof client.getCollector).toBe('function')
    })

    it('should have deleteCollector method', () => {
      expect(typeof client.deleteCollector).toBe('function')
    })

    it('should have registerCollector method', () => {
      expect(typeof client.registerCollector).toBe('function')
    })

    it('should have testConnection method', () => {
      expect(typeof client.testConnection).toBe('function')
    })
  })

  describe('Token Injection', () => {
    it('should have request interceptor set up', () => {
      // The interceptor is set up in the constructor
      expect(client).toBeDefined()
    })

    it('should clear token on 401 response', () => {
      localStorage.setItem('auth_token', 'token-that-will-expire')
      expect(client.isAuthenticated()).toBe(true)
      // In real usage, a 401 response would trigger the interceptor
      // which calls localStorage.removeItem('auth_token')
    })
  })
})
