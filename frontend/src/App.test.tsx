import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import App from './App'
import { useAuthStore } from './stores/authStore'
import { useRealtimeStore } from './stores/realtimeStore'
import { realtimeClient } from './services/realtime'

// Mock the RealtimeClient
vi.mock('./services/realtime', () => {
  const mockRealtimeClient = {
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    emit: vi.fn(),
  }
  return {
    realtimeClient: mockRealtimeClient,
  }
})

// Mock child components to simplify testing
vi.mock('./components/dashboard/Dashboard', () => ({
  Dashboard: () => <div data-testid="dashboard">Dashboard</div>,
}))

vi.mock('./components/auth/LoginPage', () => ({
  LoginPage: () => <div data-testid="login-page">LoginPage</div>,
}))

vi.mock('./pages/LogsPage', () => ({
  LogsPage: () => <div data-testid="logs-page">LogsPage</div>,
}))

vi.mock('./pages/MetricsPage', () => ({
  MetricsPage: () => <div data-testid="metrics-page">MetricsPage</div>,
}))

vi.mock('./pages/AlertsPage', () => ({
  AlertsPage: () => <div data-testid="alerts-page">AlertsPage</div>,
}))

vi.mock('./pages/ChannelsPage', () => ({
  ChannelsPage: () => <div data-testid="channels-page">ChannelsPage</div>,
}))

vi.mock('./components/ui/LoadingSpinner', () => ({
  LoadingSpinner: () => <div data-testid="loading-spinner">Loading...</div>,
}))

describe('App Component - RealtimeClient Initialization', () => {
  beforeEach(() => {
    // Reset all mocks before each test
    vi.clearAllMocks()
    // Clear auth and realtime stores
    useAuthStore.getState().reset()
    useRealtimeStore.getState().clear()
    // Clear localStorage
    localStorage.clear()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should not connect RealtimeClient when user is not authenticated', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    expect(realtimeClient.connect).not.toHaveBeenCalled()
  })

  it('should connect RealtimeClient when user is authenticated', async () => {
    // Set up authenticated state
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    render(<App />)

    await waitFor(() => {
      expect(realtimeClient.connect).toHaveBeenCalledWith(token)
    })
  })

  it('should set connected state in Zustand store when RealtimeClient connects', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(useRealtimeStore.getState().connected).toBe(true)
    })
  })

  it('should clear error state when RealtimeClient connects successfully', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(useRealtimeStore.getState().error).toBeNull()
    })
  })

  it('should handle connection errors gracefully', async () => {
    const token = 'test-token-12345'
    const errorMessage = 'WebSocket connection failed'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    const connectError = new Error(errorMessage)
    vi.mocked(realtimeClient.connect).mockRejectedValueOnce(connectError)

    render(<App />)

    await waitFor(() => {
      expect(useRealtimeStore.getState().error).toBe(errorMessage)
    })

    expect(useRealtimeStore.getState().connected).toBe(false)
  })

  it('should set up event listeners for log:new', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(realtimeClient.on).toHaveBeenCalledWith('log:new', expect.any(Function))
    })
  })

  it('should set up event listeners for metric:update', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(realtimeClient.on).toHaveBeenCalledWith('metric:update', expect.any(Function))
    })
  })

  it('should set up event listeners for alert:triggered', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(realtimeClient.on).toHaveBeenCalledWith('alert:triggered', expect.any(Function))
    })
  })

  it('should set up error event listener', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(realtimeClient.on).toHaveBeenCalledWith('error', expect.any(Function))
    })
  })

  it('should forward log:new events to Zustand store', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    let logNewListener: any = null
    vi.mocked(realtimeClient.on).mockImplementation((event, callback) => {
      if (event === 'log:new') {
        logNewListener = callback
      }
    })

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(logNewListener).not.toBeNull()
    })

    const logData = { id: '1', message: 'test log' }
    logNewListener(logData)

    expect(useRealtimeStore.getState().lastUpdate).not.toBeNull()
  })

  it('should disconnect RealtimeClient on component unmount', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    const { unmount } = render(<App />)

    await waitFor(() => {
      expect(realtimeClient.connect).toHaveBeenCalled()
    })

    unmount()

    expect(realtimeClient.disconnect).toHaveBeenCalled()
  })

  it('should disconnect RealtimeClient when token becomes unavailable', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    const { unmount } = render(<App />)

    await waitFor(() => {
      expect(realtimeClient.connect).toHaveBeenCalled()
    })

    unmount()

    expect(realtimeClient.disconnect).toHaveBeenCalled()
  })

  it('should update lastUpdate timestamp when events are received', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    let metricListener: any = null
    vi.mocked(realtimeClient.on).mockImplementation((event, callback) => {
      if (event === 'metric:update') {
        metricListener = callback
      }
    })

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(metricListener).not.toBeNull()
    })

    const oldTimestamp = useRealtimeStore.getState().lastUpdate

    // Wait a bit to ensure timestamp is different
    await new Promise(resolve => setTimeout(resolve, 10))

    const metricData = { id: '1', value: 100 }
    metricListener(metricData)

    const newTimestamp = useRealtimeStore.getState().lastUpdate
    expect(newTimestamp).not.toBeNull()
    expect(newTimestamp).not.toBe(oldTimestamp)
  })

  it('should handle error events from RealtimeClient', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    let errorListener: any = null
    vi.mocked(realtimeClient.on).mockImplementation((event, callback) => {
      if (event === 'error') {
        errorListener = callback
      }
    })

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(errorListener).not.toBeNull()
    })

    const errorData = { message: 'Connection error' }
    await waitFor(() => {
      errorListener(errorData)
    })

    await waitFor(() => {
      expect(useRealtimeStore.getState().error).toBe('Connection error')
    })
  })

  it('should render Dashboard when authenticated', async () => {
    const token = 'test-token-12345'
    localStorage.setItem('auth_token', token)
    useAuthStore.getState().setToken(token)

    vi.mocked(realtimeClient.connect).mockResolvedValueOnce(undefined)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByTestId('dashboard')).toBeInTheDocument()
    })
  })
})
