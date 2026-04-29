import { useState, useCallback, useEffect } from 'react'
import { apiClient } from '../services/api'
import type { Collector, PaginatedResponse, ApiError } from '../types'

export interface CreateCollectorData {
  name: string
  host: string
  port: number | string
  database: string
  username: string
  password: string
}

export interface UpdateCollectorData {
  name?: string
  host?: string
  port?: number | string
  database?: string
}

export function useCollectors() {
  const [collectors, setCollectors] = useState<Collector[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ApiError | null>(null)
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  })

  const fetchCollectors = useCallback(async (page: number = 1, pageSize: number = 20) => {
    setLoading(true)
    setError(null)
    try {
      const response = await apiClient.listCollectors(page, pageSize)
      setCollectors(response.data)
      setPagination({
        page: response.page,
        pageSize: response.page_size,
        total: response.total,
        totalPages: response.total_pages,
      })
    } catch (err) {
      setError(err as ApiError)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCollectors()
  }, [fetchCollectors])

  const createCollector = useCallback(async (data: CreateCollectorData) => {
    try {
      setError(null)
      // WHY: Transform frontend form data to API format.
      // The form uses 'host' and 'name' fields, but the API expects
      // 'hostname' and 'description'. We set defaults for required fields.
      const payload = {
        hostname: data.host,
        environment: 'production',
        group: 'default',
        description: data.name,
      }
      const response = await apiClient.registerCollector(payload, 'sk_default_secret')
      // WHY: Refetch after creation to get the server-generated ID and
      // ensure UI is in sync with backend state.
      await fetchCollectors()
      return response
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [fetchCollectors])

  const deleteCollector = useCallback(async (id: string) => {
    try {
      setError(null)
      await apiClient.deleteCollector(id)
      // WHY: We use local state filtering instead of refetching after delete
      // to provide immediate UI feedback and reduce API calls. If the delete
      // fails, the error is thrown and the UI can handle it appropriately.
      setCollectors((prev) => prev.filter((c) => c.id !== id))
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [])

  const updateCollector = useCallback(async (id: string, data: UpdateCollectorData) => {
    try {
      setError(null)
      // Transform frontend form data to API format if needed
      const payload = {
        hostname: data.host || undefined,
        ...data,
      }
      // Note: Update endpoint may not be fully implemented in backend yet
      // This is a placeholder for future implementation
      console.log('Update collector:', id, payload)
      // Refetch to get latest data
      await fetchCollectors()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError)
      throw apiError
    }
  }, [fetchCollectors])

  return {
    collectors,
    loading,
    error,
    pagination,
    fetchCollectors,
    createCollector,
    deleteCollector,
    updateCollector,
  }
}
