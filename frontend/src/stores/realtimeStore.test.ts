import { describe, it, expect, beforeEach } from 'vitest'
import { useRealtimeStore } from './realtimeStore'

describe('realtimeStore', () => {
  beforeEach(() => {
    useRealtimeStore.setState({
      connected: false,
      lastUpdate: null,
      error: null,
      subscriptions: new Map(),
    })
  })

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const store = useRealtimeStore.getState()
      expect(store.connected).toBe(false)
      expect(store.lastUpdate).toBeNull()
      expect(store.error).toBeNull()
      expect(store.subscriptions).toEqual(new Map())
    })
  })

  describe('setConnected', () => {
    it('should set connected to true', () => {
      const { setConnected } = useRealtimeStore.getState()
      setConnected(true)
      expect(useRealtimeStore.getState().connected).toBe(true)
    })

    it('should set connected to false', () => {
      useRealtimeStore.setState({ connected: true })
      const { setConnected } = useRealtimeStore.getState()
      setConnected(false)
      expect(useRealtimeStore.getState().connected).toBe(false)
    })
  })

  describe('setLastUpdate', () => {
    it('should set lastUpdate to a timestamp', () => {
      const timestamp = '2024-03-13T12:00:00Z'
      const { setLastUpdate } = useRealtimeStore.getState()
      setLastUpdate(timestamp)
      expect(useRealtimeStore.getState().lastUpdate).toBe(timestamp)
    })

    it('should update lastUpdate when called multiple times', () => {
      const { setLastUpdate } = useRealtimeStore.getState()
      setLastUpdate('2024-03-13T12:00:00Z')
      setLastUpdate('2024-03-13T12:00:01Z')
      expect(useRealtimeStore.getState().lastUpdate).toBe('2024-03-13T12:00:01Z')
    })
  })

  describe('setError', () => {
    it('should set error message', () => {
      const errorMsg = 'Connection failed'
      const { setError } = useRealtimeStore.getState()
      setError(errorMsg)
      expect(useRealtimeStore.getState().error).toBe(errorMsg)
    })

    it('should clear error by setting to null', () => {
      useRealtimeStore.setState({ error: 'Some error' })
      const { setError } = useRealtimeStore.getState()
      setError(null)
      expect(useRealtimeStore.getState().error).toBeNull()
    })
  })

  describe('subscribe and emit', () => {
    it('should add callback to subscriptions', () => {
      const callback = vi.fn()
      const { subscribe } = useRealtimeStore.getState()
      subscribe('test-event', callback)

      const { subscriptions } = useRealtimeStore.getState()
      expect(subscriptions.has('test-event')).toBe(true)
      expect(subscriptions.get('test-event')?.has(callback)).toBe(true)
    })

    it('should call listener when event is emitted', () => {
      const callback = vi.fn()
      const { subscribe, emit } = useRealtimeStore.getState()
      subscribe('test-event', callback)

      const testData = { message: 'hello' }
      emit('test-event', testData)

      expect(callback).toHaveBeenCalledWith(testData)
      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('should call multiple listeners for same event', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, emit } = useRealtimeStore.getState()

      subscribe('test-event', callback1)
      subscribe('test-event', callback2)

      const testData = { message: 'hello' }
      emit('test-event', testData)

      expect(callback1).toHaveBeenCalledWith(testData)
      expect(callback2).toHaveBeenCalledWith(testData)
    })

    it('should not call listeners for different events', () => {
      const callback = vi.fn()
      const { subscribe, emit } = useRealtimeStore.getState()
      subscribe('event-1', callback)

      emit('event-2', { data: 'test' })

      expect(callback).not.toHaveBeenCalled()
    })

    it('should handle multiple events independently', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, emit } = useRealtimeStore.getState()

      subscribe('event-1', callback1)
      subscribe('event-2', callback2)

      emit('event-1', { data: 'first' })
      emit('event-2', { data: 'second' })

      expect(callback1).toHaveBeenCalledWith({ data: 'first' })
      expect(callback2).toHaveBeenCalledWith({ data: 'second' })
    })

    it('should emit data of any type', () => {
      const callback = vi.fn()
      const { subscribe, emit } = useRealtimeStore.getState()
      subscribe('test', callback)

      emit('test', 'string')
      expect(callback).toHaveBeenCalledWith('string')

      emit('test', 123)
      expect(callback).toHaveBeenCalledWith(123)

      emit('test', { nested: { data: true } })
      expect(callback).toHaveBeenCalledWith({ nested: { data: true } })

      emit('test', ['array', 'data'])
      expect(callback).toHaveBeenCalledWith(['array', 'data'])
    })
  })

  describe('unsubscribe', () => {
    it('should remove specific callback from subscription', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, unsubscribe, emit } = useRealtimeStore.getState()

      subscribe('test-event', callback1)
      subscribe('test-event', callback2)

      unsubscribe('test-event', callback1)

      emit('test-event', { data: 'test' })

      expect(callback1).not.toHaveBeenCalled()
      expect(callback2).toHaveBeenCalledWith({ data: 'test' })
    })

    it('should remove all listeners for event when callback is not specified', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, unsubscribe, emit } = useRealtimeStore.getState()

      subscribe('test-event', callback1)
      subscribe('test-event', callback2)

      unsubscribe('test-event')

      emit('test-event', { data: 'test' })

      expect(callback1).not.toHaveBeenCalled()
      expect(callback2).not.toHaveBeenCalled()
    })

    it('should handle unsubscribe for non-existent event', () => {
      const callback = vi.fn()
      const { unsubscribe } = useRealtimeStore.getState()

      // Should not throw
      unsubscribe('non-existent-event', callback)
      expect(true).toBe(true)
    })

    it('should not affect other events when unsubscribing', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, unsubscribe, emit } = useRealtimeStore.getState()

      subscribe('event-1', callback1)
      subscribe('event-2', callback2)

      unsubscribe('event-1', callback1)

      emit('event-1', { data: 'first' })
      emit('event-2', { data: 'second' })

      expect(callback1).not.toHaveBeenCalled()
      expect(callback2).toHaveBeenCalledWith({ data: 'second' })
    })

    it('should clean up empty event subscriptions after removing all callbacks', () => {
      const callback = vi.fn()
      const { subscribe, unsubscribe } = useRealtimeStore.getState()

      subscribe('test-event', callback)
      unsubscribe('test-event', callback)

      const { subscriptions } = useRealtimeStore.getState()
      expect(subscriptions.has('test-event')).toBe(false)
    })
  })

  describe('clear', () => {
    it('should reset all state to initial values', () => {
      const callback = vi.fn()
      const { subscribe, setConnected, setError, setLastUpdate, clear } = useRealtimeStore.getState()

      subscribe('test-event', callback)
      setConnected(true)
      setError('Some error')
      setLastUpdate('2024-03-13T12:00:00Z')

      clear()

      const state = useRealtimeStore.getState()
      expect(state.connected).toBe(false)
      expect(state.error).toBeNull()
      expect(state.lastUpdate).toBeNull()
      expect(state.subscriptions.size).toBe(0)
    })

    it('should clear all subscriptions', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()
      const { subscribe, clear, emit } = useRealtimeStore.getState()

      subscribe('event-1', callback1)
      subscribe('event-2', callback2)

      clear()

      emit('event-1', { data: 'test' })
      emit('event-2', { data: 'test' })

      expect(callback1).not.toHaveBeenCalled()
      expect(callback2).not.toHaveBeenCalled()
    })

    it('should allow subscription after clear', () => {
      const callback = vi.fn()
      const { subscribe, clear, emit } = useRealtimeStore.getState()

      clear()
      subscribe('new-event', callback)
      emit('new-event', { data: 'test' })

      expect(callback).toHaveBeenCalledWith({ data: 'test' })
    })
  })

  describe('concurrent operations', () => {
    it('should handle multiple subscriptions and unsubscriptions', () => {
      const callbacks = Array.from({ length: 5 }, () => vi.fn())
      const { subscribe, unsubscribe, emit } = useRealtimeStore.getState()

      callbacks.forEach((cb, i) => {
        subscribe(`event-${i}`, cb)
      })

      unsubscribe('event-1')
      unsubscribe('event-3')

      emit('event-0', { data: 'test' })
      emit('event-1', { data: 'test' })
      emit('event-2', { data: 'test' })
      emit('event-3', { data: 'test' })
      emit('event-4', { data: 'test' })

      expect(callbacks[0]).toHaveBeenCalled()
      expect(callbacks[1]).not.toHaveBeenCalled()
      expect(callbacks[2]).toHaveBeenCalled()
      expect(callbacks[3]).not.toHaveBeenCalled()
      expect(callbacks[4]).toHaveBeenCalled()
    })

    it('should handle rapid state updates', () => {
      const { setConnected, setError, setLastUpdate } = useRealtimeStore.getState()

      for (let i = 0; i < 10; i++) {
        setConnected(i % 2 === 0)
        setError(i % 3 === 0 ? `Error ${i}` : null)
        setLastUpdate(`2024-03-13T12:00:${String(i).padStart(2, '0')}Z`)
      }

      const state = useRealtimeStore.getState()
      expect(state.connected).toBe(false) // Last iteration (9) is odd
      expect(state.error).toBe('Error 9') // Last iteration (9) is divisible by 3
      expect(state.lastUpdate).toBe('2024-03-13T12:00:09Z')
    })
  })

  describe('Zustand hook integration', () => {
    it('should work with getState and setState', () => {
      const { setConnected, setError } = useRealtimeStore.getState()

      setConnected(true)
      let state = useRealtimeStore.getState()
      expect(state.connected).toBe(true)

      setError('Test error')
      state = useRealtimeStore.getState()
      expect(state.error).toBe('Test error')
    })

    it('should expose store as Zustand hook', () => {
      // Verify that useRealtimeStore is a valid Zustand hook
      expect(useRealtimeStore).toBeDefined()
      expect(useRealtimeStore.getState).toBeDefined()
      expect(useRealtimeStore.setState).toBeDefined()
      expect(useRealtimeStore.subscribe).toBeDefined()

      // Verify we can get current state
      const state = useRealtimeStore.getState()
      expect(state).toHaveProperty('connected')
      expect(state).toHaveProperty('lastUpdate')
      expect(state).toHaveProperty('error')
      expect(state).toHaveProperty('subscriptions')
      expect(state).toHaveProperty('setConnected')
      expect(state).toHaveProperty('setError')
    })
  })
})
