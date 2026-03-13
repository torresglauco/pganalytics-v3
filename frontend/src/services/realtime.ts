type EventListener = (data: any) => void

interface Message {
  type: string
  data: any
}

export class RealtimeClient {
  private ws: WebSocket | null = null
  private url: string
  private token: string = ''
  private listeners: Map<string, Set<EventListener>> = new Map()
  private messageQueue: Message[] = []
  private reconnectAttempts: number = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private pingTimer: NodeJS.Timeout | null = null
  private maxReconnectAttempts: number = 5

  constructor(baseURL: string) {
    this.url = baseURL
      .replace(/^https/, 'wss')
      .replace(/^http/, 'ws')
  }

  async connect(token: string): Promise<void> {
    this.token = token
    return this.attemptConnect()
  }

  private async attemptConnect(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      try {
        const wsUrl = `${this.url}?token=${this.token}`
        this.ws = new WebSocket(wsUrl)

        let isResolved = false

        const handleOpen = () => {
          this.reconnectAttempts = 0
          this.setupHeartbeat()
          this.flushMessageQueue()
          cleanup()
          if (!isResolved) {
            isResolved = true
            resolve()
          }
        }

        const handleError = () => {
          cleanup()
          if (!isResolved) {
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
              this.attemptReconnect()
              isResolved = true
              resolve() // Continue attempting in background
            } else {
              this.emitError(
                new Error('Max reconnection attempts reached, falling back to polling')
              )
              isResolved = true
              reject(new Error('Failed to establish WebSocket connection'))
            }
          }
        }

        const handleClose = () => {
          cleanup()
          if (!isResolved) {
            // Connection was closed before opening
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
              this.attemptReconnect()
              isResolved = true
              resolve()
            }
          }
        }

        const cleanup = () => {
          if (this.ws) {
            this.ws.removeEventListener('open', handleOpen)
            this.ws.removeEventListener('error', handleError)
            this.ws.removeEventListener('close', handleClose)
          }
        }

        this.ws.addEventListener('open', handleOpen)
        this.ws.addEventListener('error', handleError)
        this.ws.addEventListener('close', handleClose)
        this.ws.addEventListener('message', (event) => this.handleMessage(event))
      } catch (error) {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.attemptReconnect()
          resolve()
        } else {
          this.emitError(error as Error)
          reject(error)
        }
      }
    })
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.pingTimer) {
      clearInterval(this.pingTimer)
      this.pingTimer = null
    }

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }

    this.reconnectAttempts = 0
  }

  on(event: string, callback: EventListener): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  off(event: string, callback?: EventListener): void {
    if (!callback) {
      // Remove all listeners for this event
      this.listeners.delete(event)
      return
    }

    const listeners = this.listeners.get(event)
    if (listeners) {
      listeners.delete(callback)
      if (listeners.size === 0) {
        this.listeners.delete(event)
      }
    }
  }

  emit(event: string, data: any): void {
    const message: Message = { type: event, data }

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      this.messageQueue.push(message)
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message = JSON.parse(event.data) as Message

      // Dispatch to listeners
      const listeners = this.listeners.get(message.type)
      if (listeners) {
        listeners.forEach((callback) => {
          try {
            callback(message.data)
          } catch (error) {
            console.error(`Error in listener for ${message.type}:`, error)
          }
        })
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error)
    }
  }

  private setupHeartbeat(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer)
    }

    this.pingTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.emit('ping', {})
      }
    }, 30000) as NodeJS.Timeout
  }

  private attemptReconnect(): void {
    if (this.reconnectTimer) {
      return // Already scheduled
    }

    this.reconnectAttempts++

    // Exponential backoff: 1s, 2s, 4s, 8s, 30s (max)
    const delays = [1000, 2000, 4000, 8000, 30000]
    const delay = delays[Math.min(this.reconnectAttempts - 1, delays.length - 1)]

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.attemptConnect().catch(() => {
        // Continue attempting to reconnect
      })
    }, delay) as NodeJS.Timeout
  }

  private flushMessageQueue(): void {
    while (this.messageQueue.length > 0 && this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = this.messageQueue.shift()!
      this.ws.send(JSON.stringify(message))
    }
  }

  private emitError(error: Error): void {
    const listeners = this.listeners.get('error')
    if (listeners) {
      listeners.forEach((callback) => {
        try {
          callback({ message: error.message })
        } catch (err) {
          console.error('Error in error listener:', err)
        }
      })
    } else {
      console.error('RealtimeClient error:', error)
    }
  }
}

export const realtimeClient = new RealtimeClient(
  import.meta.env.VITE_API_URL || 'http://localhost:8000'
)
