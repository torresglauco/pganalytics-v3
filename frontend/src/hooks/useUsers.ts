import { useState, useCallback, useEffect } from 'react'
import { apiClient } from '../services/api'
import type { ApiError } from '../types'

export interface AdminUser {
  id: string | number
  email: string
  username?: string
  full_name?: string
  role: 'admin' | 'viewer' | 'operator'
  is_active?: boolean
  status?: 'active' | 'inactive'
  last_login?: Date | string | null
  created_at?: Date | string
  password_changed?: boolean
}

export interface CreateUserData {
  email: string
  full_name?: string
  username?: string
  password?: string
  role: 'admin' | 'viewer' | 'operator'
}

export interface UpdateUserData {
  role?: 'admin' | 'viewer' | 'operator'
  is_active?: boolean
  name?: string
  email?: string
}

export interface ResetPasswordResponse {
  username: string
  temp_password: string
  message: string
}

export function useUsers() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ApiError | null>(null)
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  })

  const fetchUsers = useCallback(async (page: number = 1, pageSize: number = 20) => {
    setLoading(true)
    setError(null)
    try {
      const response = await apiClient.listUsers(page, pageSize)
      setUsers(response.data || [])
      setPagination({
        page: response.page || 1,
        pageSize: response.page_size || pageSize,
        total: response.total || 0,
        totalPages: response.total_pages || 0,
      })
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      console.error('Failed to fetch users:', apiError)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const createUser = useCallback(async (data: CreateUserData) => {
    try {
      setError(null)
      const response = await apiClient.createUser(data)
      // Refetch users after successful creation
      await fetchUsers()
      return response
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [fetchUsers])

  const updateUser = useCallback(async (id: string | number, data: UpdateUserData) => {
    try {
      setError(null)
      const response = await apiClient.updateUser(String(id), data)
      // Refetch users after successful update
      await fetchUsers()
      return response
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [fetchUsers])

  const deleteUser = useCallback(async (id: string | number) => {
    try {
      setError(null)
      await apiClient.deleteUser(String(id))
      setUsers((prev) => prev.filter((u) => String(u.id) !== String(id)))
      return true
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [])

  const resetPassword = useCallback(async (id: string | number): Promise<ResetPasswordResponse> => {
    try {
      setError(null)
      const response = await apiClient.resetUserPassword(String(id))
      return response as ResetPasswordResponse
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [])

  return {
    users,
    loading,
    error,
    pagination,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    resetPassword,
  }
}
