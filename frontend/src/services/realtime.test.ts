import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { RealtimeClient, realtimeClient } from './realtime'

// Mock WebSocket
class MockWebSocket {
  url: string
  readyState: number = WebSocket.CONNECTING
  send = vi.fn()
  close = vi.fn()
  addEventListener = vi.fn((event: string, handler: EventListener) => {
    ;(this as any)[`_on${event}`] = handler
  })
  removeEventListener = vi.fn()

  constructor(url: string) {
    this.url = url
  }

  // Helper methods to trigger events in tests
  triggerOpen() {
    this.readyState = WebSocket.OPEN
    if ((this as any)._onopen) {
      ;(this as any)._onopen(new Event('open'))
    }
  }

  triggerError() {
    if ((this as any)._onerror) {
      ;(this as any)._onerror(new Event('error'))
    }
  }

  triggerMessage(data: string) {
    if ((this as any)._onmessage) {
      ;(this as any)._onmessage(new MessageEvent('message', { data }))
    }
  }
}

let mockWebSocketInstance: MockWebSocket | null = null
const originalWebSocket = global.WebSocket

beforeEach(() => {
  mockWebSocketInstance = null
  vi.useFakeTimers()
  vi.clearAllMocks()

  // Mock global WebSocket
  global.WebSocket = vi.fn((url: string) => {
    mockWebSocketInstance = new MockWebSocket(url)
    return mockWebSocketInstance as any
  }) as any
})

afterEach(() => {
  global.WebSocket = originalWebSocket
  vi.runOnlyPendingTimers()
  vi.useRealTimers()
})

describe('RealtimeClient', () => {
  let client: RealtimeClient

  beforeEach(() => {
    client = new RealtimeClient('http://localhost:8000')
  })

  describe('Constructor', () => {
    it('should convert http to ws', () => {
      const client = new RealtimeClient('http://localhost:8000')
      expect(client['url']).toBe('ws://localhost:8000')
    })

    it('should convert https to wss', () => {
      const client = new RealtimeClient('https://example.com')
      expect(client['url']).toBe('wss://example.com')
    })

    it('should initialize empty listeners map', () => {
      expect(client['listeners']).toBeInstanceOf(Map)
      expect(client['listeners'].size).toBe(0)
    })

    it('should initialize empty message queue', () => {
      expect(client['messageQueue']).toEqual([])
    })

    it('should initialize reconnect attempts to 0', () => {
      expect(client['reconnectAttempts']).toBe(0)
    })
  })

  describe('connect', () => {
    it('should establish WebSocket connection', async () => {
      const connectPromise = client.connect('test-token')
      expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8000?token=test-token')

      // Trigger open event to resolve promise
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise
    })

    it('should store token', async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise
      expect(client['token']).toBe('test-token')
    })

    it('should setup heartbeat on connect', async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      // Check that ping timer was set
      expect(client['pingTimer']).toBeTruthy()
    })

    it('should reset reconnect attempts on successful connect', async () => {
      client['reconnectAttempts'] = 3
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      expect(client['reconnectAttempts']).toBe(0)
    })
  })

  describe('disconnect', () => {
    beforeEach(async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise
    })

    it('should close WebSocket connection', () => {
      client.disconnect()
      expect(mockWebSocketInstance?.close).toHaveBeenCalled()
    })

    it('should clear ping timer', () => {
      client.disconnect()
      expect(client['pingTimer']).toBeNull()
    })

    it('should clear reconnect timer', () => {
      client.disconnect()
      expect(client['reconnectTimer']).toBeNull()
    })

    it('should clear WebSocket reference', () => {
      client.disconnect()
      expect(client['ws']).toBeNull()
    })

    it('should reset reconnect attempts', () => {
      client['reconnectAttempts'] = 5
      client.disconnect()
      expect(client['reconnectAttempts']).toBe(0)
    })
  })

  describe('Event Listeners', () => {
    it('should register event listener with on', () => {
      const callback = vi.fn()
      client.on('test', callback)

      expect(client['listeners'].has('test')).toBe(true)
      expect(client['listeners'].get('test')?.has(callback)).toBe(true)
    })

    it('should support multiple listeners for same event', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()

      client.on('test', callback1)
      client.on('test', callback2)

      const listeners = client['listeners'].get('test')
      expect(listeners?.size).toBe(2)
    })

    it('should remove specific listener with off', () => {
      const callback = vi.fn()
      client.on('test', callback)
      client.off('test', callback)

      const listeners = client['listeners'].get('test')
      expect(listeners).toBeUndefined()
    })

    it('should remove all listeners when callback not specified', () => {
      const callback1 = vi.fn()
      const callback2 = vi.fn()

      client.on('test', callback1)
      client.on('test', callback2)
      client.off('test')

      expect(client['listeners'].get('test')).toBeUndefined()
    })

    it('should call listeners when message received', async () => {
      const callback = vi.fn()
      client.on('log:new', callback)

      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerMessage(JSON.stringify({ type: 'log:new', data: { id: '123' } }))
      }

      expect(callback).toHaveBeenCalledWith({ id: '123' })
    })

    it('should handle errors in listeners gracefully', async () => {
      const callback = vi.fn(() => {
        throw new Error('Listener error')
      })
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      client.on('test', callback)

      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerMessage(JSON.stringify({ type: 'test', data: {} }))
      }

      expect(callback).toHaveBeenCalled()
      expect(consoleSpy).toHaveBeenCalled()
      consoleSpy.mockRestore()
    })
  })

  describe('Message Queue', () => {
    it('should queue messages while offline', () => {
      client.emit('test:event', { foo: 'bar' })

      expect(client['messageQueue'].length).toBe(1)
      expect(client['messageQueue'][0]).toEqual({
        type: 'test:event',
        data: { foo: 'bar' },
      })
    })

    it('should send queued messages after connecting', async () => {
      client.emit('test:event', { foo: 'bar' })

      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      expect(mockWebSocketInstance?.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'test:event', data: { foo: 'bar' } })
      )
    })

    it('should clear queue after sending', async () => {
      client.emit('test:event', { foo: 'bar' })

      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      expect(client['messageQueue'].length).toBe(0)
    })

    it('should send messages directly when connected', async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      client.emit('test:event', { foo: 'bar' })

      expect(mockWebSocketInstance?.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'test:event', data: { foo: 'bar' } })
      )
      expect(client['messageQueue'].length).toBe(0)
    })
  })

  describe('Heartbeat', () => {
    it('should send ping every 30 seconds', async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      vi.advanceTimersByTime(30000)

      expect(mockWebSocketInstance?.send).toHaveBeenCalledWith(
        JSON.stringify({ type: 'ping', data: {} })
      )
    })

    it('should not send ping if disconnected', async () => {
      const connectPromise = client.connect('test-token')
      if (mockWebSocketInstance) {
        mockWebSocketInstance.triggerOpen()
      }
      await connectPromise

      client.disconnect()
      vi.advanceTimersByTime(30000)

      // Should not have been called for ping after disconnect
      const pingCalls = (mockWebSocketInstance?.send as any).mock.calls.filter(
        (call: any[]) => call[0]?.includes('ping')
      )
      expect(pingCalls.length).toBe(0)
    })
  })

  describe('Exponential Backoff', () => {
    it('should retry with exponential backoff delays', async () => {
      let connectAttempts = 0
      global.WebSocket = vi.fn(() => {
        connectAttempts++
        const ws = new MockWebSocket('ws://localhost:8000?token=test-token')
        // Only succeed on 3rd attempt
        if (connectAttempts < 3) {
          // Trigger error after a microtask
          Promise.resolve().then(() => ws.triggerError())
        }
        return ws as any
      })

      const connectPromise = client.connect('test-token')

      // First attempt immediate
      expect(connectAttempts).toBe(1)

      // Advance past first retry delay (1s)
      vi.advanceTimersByTime(1000)
      await vi.runOnlyPendingTimersAsync()
      expect(connectAttempts).toBe(2)

      // Advance past second retry delay (2s)
      vi.advanceTimersByTime(2000)
      await vi.runOnlyPendingTimersAsync()

      // Let the promise resolve
      mockWebSocketInstance?.triggerOpen()
      await connectPromise

      expect(connectAttempts).toBeGreaterThanOrEqual(2)
    })
  })

  describe('Error Handling', () => {
    it('should emit error event on connection failure', async () => {
      const errorHandler = vi.fn()
      client.on('error', errorHandler)

      global.WebSocket = vi.fn(() => {
        const ws = new MockWebSocket('ws://localhost:8000?token=test-token')
        Promise.resolve().then(() => ws.triggerError())
        return ws as any
      })

      const connectPromise = client.connect('test-token')

      // Run through all retry attempts
      for (let i = 0; i < 6; i++) {
        vi.advanceTimersByTime(40000) // Enough to cover all retries
        await vi.runOnlyPendingTimersAsync()
      }

      try {
        await connectPromise
      } catch (e) {
        // Expected to fail
      }

      // After max retries, error should be emitted
      expect(errorHandler).toHaveBeenCalled()
    })
  })

  describe('Singleton Export', () => {
    it('should export a realtimeClient instance', () => {
      expect(realtimeClient).toBeInstanceOf(RealtimeClient)
    })

    it('should have correct base URL', () => {
      expect(realtimeClient['url']).toMatch(/^wss?:\/\//)
    })
  })
})
