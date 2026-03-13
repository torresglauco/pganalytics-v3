import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useRealtime } from './useRealtime'
import { useRealtimeStore } from '../stores/realtimeStore'

vi.mock('../stores/realtimeStore')

describe('useRealtime', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('should return correct initial state', () => {
      // Mock the store to return initial state
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      expect(result.current.connected).toBe(false)
      expect(result.current.lastUpdate).toBeNull()
      expect(result.current.error).toBeNull()
      expect(result.current.subscribe).toBeDefined()
      expect(result.current.unsubscribe).toBeDefined()
    })
  })

  describe('state tracking', () => {
    it('should return updated connected state', () => {
      const mockStore = {
        connected: true,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      expect(result.current.connected).toBe(true)
    })

    it('should return updated lastUpdate state', () => {
      const timestamp = '2024-03-13T12:00:00Z'
      const mockStore = {
        connected: false,
        lastUpdate: timestamp,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      expect(result.current.lastUpdate).toBe(timestamp)
    })

    it('should return error state', () => {
      const errorMsg = 'Connection failed'
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: errorMsg,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      expect(result.current.error).toBe(errorMsg)
    })
  })

  describe('subscribe method', () => {
    it('should call store subscribe method with event and callback', () => {
      const mockSubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: mockSubscribe,
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())
      const callback = vi.fn()

      act(() => {
        result.current.subscribe('test-event', callback)
      })

      expect(mockSubscribe).toHaveBeenCalledWith('test-event', callback)
    })

    it('should allow subscribing to multiple events', () => {
      const mockSubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: mockSubscribe,
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())
      const callback1 = vi.fn()
      const callback2 = vi.fn()

      act(() => {
        result.current.subscribe('event-1', callback1)
        result.current.subscribe('event-2', callback2)
      })

      expect(mockSubscribe).toHaveBeenCalledTimes(2)
      expect(mockSubscribe).toHaveBeenNthCalledWith(1, 'event-1', callback1)
      expect(mockSubscribe).toHaveBeenNthCalledWith(2, 'event-2', callback2)
    })
  })

  describe('unsubscribe method', () => {
    it('should call store unsubscribe method with event and callback', () => {
      const mockUnsubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: mockUnsubscribe,
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())
      const callback = vi.fn()

      act(() => {
        result.current.unsubscribe('test-event', callback)
      })

      expect(mockUnsubscribe).toHaveBeenCalledWith('test-event', callback)
    })

    it('should call store unsubscribe method without callback', () => {
      const mockUnsubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: mockUnsubscribe,
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      act(() => {
        result.current.unsubscribe('test-event')
      })

      expect(mockUnsubscribe).toHaveBeenCalledWith('test-event', undefined)
    })

    it('should allow unsubscribing from multiple events', () => {
      const mockUnsubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: mockUnsubscribe,
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())
      const callback1 = vi.fn()
      const callback2 = vi.fn()

      act(() => {
        result.current.unsubscribe('event-1', callback1)
        result.current.unsubscribe('event-2', callback2)
      })

      expect(mockUnsubscribe).toHaveBeenCalledTimes(2)
      expect(mockUnsubscribe).toHaveBeenNthCalledWith(1, 'event-1', callback1)
      expect(mockUnsubscribe).toHaveBeenNthCalledWith(2, 'event-2', callback2)
    })
  })

  describe('memoization', () => {
    it('should memoize subscribe method to prevent unnecessary updates', () => {
      const mockSubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: mockSubscribe,
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result, rerender } = renderHook(() => useRealtime())
      const callback = vi.fn()
      const firstSubscribe = result.current.subscribe

      // Call subscribe
      act(() => {
        result.current.subscribe('test-event', callback)
      })

      // Rerender with same store state
      rerender()
      const secondSubscribe = result.current.subscribe

      // Subscribe should be the same reference (memoized)
      expect(firstSubscribe).toBe(secondSubscribe)
    })

    it('should memoize unsubscribe method to prevent unnecessary updates', () => {
      const mockUnsubscribe = vi.fn()
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: mockUnsubscribe,
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result, rerender } = renderHook(() => useRealtime())
      const firstUnsubscribe = result.current.unsubscribe

      // Rerender with same store state
      rerender()
      const secondUnsubscribe = result.current.unsubscribe

      // Unsubscribe should be the same reference (memoized)
      expect(firstUnsubscribe).toBe(secondUnsubscribe)
    })
  })

  describe('selector pattern', () => {
    it('should use selector pattern to get connected state', () => {
      const mockStore = {
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      const selectorMock = vi.fn((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      ;(useRealtimeStore as any).mockImplementation(selectorMock)

      renderHook(() => useRealtime())

      // Verify selectors were called for each state property
      expect(selectorMock).toHaveBeenCalled()
    })
  })

  describe('integration test', () => {
    it('should provide complete interface for components', () => {
      const mockStore = {
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      const { result } = renderHook(() => useRealtime())

      // Verify all properties exist
      expect(result.current).toHaveProperty('connected')
      expect(result.current).toHaveProperty('lastUpdate')
      expect(result.current).toHaveProperty('error')
      expect(result.current).toHaveProperty('subscribe')
      expect(result.current).toHaveProperty('unsubscribe')

      // Verify types
      expect(typeof result.current.connected).toBe('boolean')
      expect(['object', 'string']).toContain(typeof result.current.lastUpdate) // null or string
      expect(['object', 'string']).toContain(typeof result.current.error) // null or string
      expect(typeof result.current.subscribe).toBe('function')
      expect(typeof result.current.unsubscribe).toBe('function')
    })

    it('should not take any parameters', () => {
      const mockStore = {
        connected: false,
        lastUpdate: null,
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      }

      ;(useRealtimeStore as any).mockImplementation((selector: any) => {
        if (typeof selector === 'function') {
          return selector(mockStore)
        }
        return mockStore
      })

      // Should not throw when called without parameters
      const { result } = renderHook(() => useRealtime())

      expect(result.current).toBeDefined()
    })
  })
})
