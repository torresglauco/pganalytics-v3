import { useState, useCallback, useEffect } from 'react'
import { apiClient } from '../services/api'
import type { Collector, PaginatedResponse, ApiError } from '../types'

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

  const deleteCollector = useCallback(async (id: string) => {
    try {
      await apiClient.deleteCollector(id)
      setCollectors((prev) => prev.filter((c) => c.id !== id))
    } catch (err) {
      setError(err as ApiError)
      throw err
    }
  }, [])

  return {
    collectors,
    loading,
    error,
    pagination,
    fetchCollectors,
    deleteCollector,
  }
}
