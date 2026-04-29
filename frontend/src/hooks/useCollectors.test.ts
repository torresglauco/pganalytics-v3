import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useCollectors } from './useCollectors'
import { apiClient } from '../services/api'

vi.mock('../services/api')

describe('useCollectors', () => {
  const mockCollectors = [
    {
      id: 'collector-1',
      hostname: 'localhost',
      status: 'active' as const,
      created_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'collector-2',
      hostname: 'prod.example.com',
      status: 'active' as const,
      created_at: '2024-01-02T00:00:00Z',
    },
  ]

  const mockPaginatedResponse = {
    data: mockCollectors,
    total: 2,
    page: 1,
    page_size: 20,
    total_pages: 1,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch collectors on mount', async () => {
    vi.mocked(apiClient.listCollectors).mockResolvedValue(mockPaginatedResponse)

    const { result } = renderHook(() => useCollectors())

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.collectors).toEqual(mockCollectors)
    expect(result.current.pagination).toEqual({
      page: 1,
      pageSize: 20,
      total: 2,
      totalPages: 1,
    })
    expect(vi.mocked(apiClient.listCollectors)).toHaveBeenCalledWith(1, 20)
  })

  it('should handle fetch error', async () => {
    const error = {
      message: 'Failed to fetch collectors',
      status_code: 500,
    }

    vi.mocked(apiClient.listCollectors).mockRejectedValue(error)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(error)
    expect(result.current.collectors).toEqual([])
  })

  it('should fetch collectors with custom pagination', async () => {
    vi.mocked(apiClient.listCollectors).mockResolvedValue(mockPaginatedResponse)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await result.current.fetchCollectors(2, 10)

    expect(vi.mocked(apiClient.listCollectors)).toHaveBeenCalledWith(2, 10)
  })

  it('should delete collector and update state', async () => {
    vi.mocked(apiClient.listCollectors).mockResolvedValue(mockPaginatedResponse)
    vi.mocked(apiClient.deleteCollector).mockResolvedValue(undefined)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.collectors).toHaveLength(2)

    // Delete the collector
    await result.current.deleteCollector('collector-1')

    // The hook uses local state filtering, not a refetch
    // So we need to check that the delete was called and the state was updated locally
    expect(vi.mocked(apiClient.deleteCollector)).toHaveBeenCalledWith('collector-1')

    // Wait for state update
    await waitFor(() => {
      expect(result.current.collectors).toHaveLength(1)
      expect(result.current.collectors[0].id).toBe('collector-2')
    })
  })

  it('should handle delete error', async () => {
    const initialResponse = {
      ...mockPaginatedResponse,
      data: [mockCollectors[0]],
    }

    vi.mocked(apiClient.listCollectors).mockResolvedValue(initialResponse)
    const deleteError = {
      message: 'Failed to delete collector',
      status_code: 500,
    }
    vi.mocked(apiClient.deleteCollector).mockRejectedValue(deleteError)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    // The error should be set when deletion fails
    await expect(result.current.deleteCollector('collector-1')).rejects.toEqual(deleteError)

    // The error state is set in the hook
    await waitFor(() => {
      expect(result.current.error).toEqual(deleteError)
    })

    expect(result.current.collectors).toHaveLength(1)
  })

  it('should update pagination state', async () => {
    const page2Response = {
      data: [
        {
          id: 'collector-3',
          hostname: 'staging.example.com',
          status: 'active' as const,
          created_at: '2024-01-03T00:00:00Z',
        },
      ],
      total: 3,
      page: 2,
      page_size: 2,
      total_pages: 2,
    }

    vi.mocked(apiClient.listCollectors).mockResolvedValue(page2Response)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await result.current.fetchCollectors(2, 2)

    expect(result.current.pagination).toEqual({
      page: 2,
      pageSize: 2,
      total: 3,
      totalPages: 2,
    })
  })

  it('should handle network error gracefully', async () => {
    vi.mocked(apiClient.listCollectors).mockRejectedValue(new Error('Network Error'))

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBeTruthy()
    expect(result.current.error?.message).toBe('Network Error')
    expect(result.current.collectors).toEqual([])
  })

  it('should handle 401 unauthorized error', async () => {
    const authError = {
      message: 'Unauthorized',
      status_code: 401,
    }

    vi.mocked(apiClient.listCollectors).mockRejectedValue(authError)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(authError)
    expect(result.current.error?.status_code).toBe(401)
    expect(result.current.collectors).toEqual([])
  })

  it('should handle 403 forbidden error', async () => {
    const forbiddenError = {
      message: 'Access denied',
      status_code: 403,
    }

    vi.mocked(apiClient.listCollectors).mockRejectedValue(forbiddenError)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(forbiddenError)
    expect(result.current.error?.status_code).toBe(403)
    expect(result.current.collectors).toEqual([])
  })

  it('should handle 404 not found error', async () => {
    const notFoundError = {
      message: 'Resource not found',
      status_code: 404,
    }

    vi.mocked(apiClient.listCollectors).mockRejectedValue(notFoundError)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(notFoundError)
    expect(result.current.error?.status_code).toBe(404)
    expect(result.current.collectors).toEqual([])
  })

  it('should handle 400 bad request error', async () => {
    const badRequestError = {
      message: 'Invalid request parameters',
      status_code: 400,
    }

    vi.mocked(apiClient.listCollectors).mockRejectedValue(badRequestError)

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(badRequestError)
    expect(result.current.error?.status_code).toBe(400)
    expect(result.current.collectors).toEqual([])
  })
})
