import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Dashboard } from './components/dashboard/Dashboard'
import { LoginPage } from './components/auth/LoginPage'
import { LogsPage } from './pages/LogsPage'
import { MetricsPage } from './pages/MetricsPage'
import { AlertsPage } from './pages/AlertsPage'
import { ChannelsPage } from './pages/ChannelsPage'
import { NotImplementedPage } from './pages/NotImplementedPage'
import { useAuthStore } from './stores/authStore'
import { useRealtimeStore } from './stores/realtimeStore'
import { realtimeClient } from './services/realtime'
import { apiClient } from './services/api'
import { LoadingSpinner } from './components/ui/LoadingSpinner'
import './styles/index.css'

// Component to redirect to external Grafana service
function GrafanaRedirect() {
  useEffect(() => {
    // Redirect to Grafana service running on port 3001
    window.location.href = 'http://localhost:3001'
  }, [])
  return <LoadingSpinner fullScreen message="Redirecting to Grafana..." />
}

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const { isAuthenticated, setAuthenticated, token } = useAuthStore()
  const { setConnected, setError: setRealtimeError, setLastUpdate, emit } = useRealtimeStore()

  useEffect(() => {
    const checkAuthentication = async () => {
      try {
        setIsLoading(true)

        if (apiClient.isAuthenticated()) {
          // Could fetch user profile here
          setAuthenticated(true)
        } else {
          setAuthenticated(false)
        }
      } catch (err) {
        console.error('Auth check failed:', err)
        setAuthenticated(false)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuthentication()
  }, [setAuthenticated])

  // Initialize RealtimeClient connection when authenticated
  useEffect(() => {
    if (!isAuthenticated || !token) {
      // Disconnect if not authenticated
      if (realtimeClient) {
        realtimeClient.disconnect()
        setConnected(false)
      }
      return
    }

    // Handler for connection established
    const handleConnected = () => {
      setConnected(true)
      setRealtimeError(null)
    }

    // Handler for disconnection
    const handleDisconnected = () => {
      setConnected(false)
    }

    // Handler for connection errors
    const handleError = (error: any) => {
      const errorMessage = error?.message || 'Connection error'
      setRealtimeError(errorMessage)
      console.error('RealtimeClient error:', error)
    }

    // Handler for new logs received
    const handleLogReceived = (logData: any) => {
      setLastUpdate(new Date().toISOString())
      emit('log:new', logData)
    }

    // Handler for metric updates
    const handleMetricUpdate = (metricData: any) => {
      setLastUpdate(new Date().toISOString())
      emit('metric:update', metricData)
    }

    // Handler for alert triggers
    const handleAlertTriggered = (alertData: any) => {
      setLastUpdate(new Date().toISOString())
      emit('alert:triggered', alertData)
    }

    // Connect to WebSocket
    realtimeClient
      .connect(token)
      .then(() => {
        setConnected(true)
        setRealtimeError(null)

        // Subscribe to events
        realtimeClient.on('log:new', handleLogReceived)
        realtimeClient.on('metric:update', handleMetricUpdate)
        realtimeClient.on('alert:triggered', handleAlertTriggered)
        realtimeClient.on('connected', handleConnected)
        realtimeClient.on('disconnected', handleDisconnected)
        realtimeClient.on('error', handleError)
      })
      .catch((error) => {
        const errorMessage = error?.message || 'Failed to connect to realtime service'
        setRealtimeError(errorMessage)
        console.error('Failed to connect RealtimeClient:', error)
      })

    // Cleanup on unmount or token change
    return () => {
      realtimeClient.off('log:new', handleLogReceived)
      realtimeClient.off('metric:update', handleMetricUpdate)
      realtimeClient.off('alert:triggered', handleAlertTriggered)
      realtimeClient.off('connected', handleConnected)
      realtimeClient.off('disconnected', handleDisconnected)
      realtimeClient.off('error', handleError)
      realtimeClient.disconnect()
      setConnected(false)
    }
  }, [isAuthenticated, token, setConnected, setRealtimeError, setLastUpdate, emit])

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading pgAnalytics..." />
  }

  return (
    <BrowserRouter>
      <Routes>
        {/* Public Routes */}
        <Route
          path="/login"
          element={isAuthenticated ? <Navigate to="/" /> : <LoginPage />}
        />

        {/* Protected Routes */}
        {isAuthenticated ? (
          <>
            <Route path="/" element={<Dashboard />} />
            <Route path="/logs" element={<LogsPage />} />
            <Route path="/metrics" element={<MetricsPage />} />
            <Route path="/alerts" element={<AlertsPage />} />
            <Route path="/channels" element={<ChannelsPage />} />
            {/* Grafana redirect to external service */}
            <Route path="/grafana" element={<GrafanaRedirect />} />
            {/* Unimplemented routes - show "Coming Soon" pages */}
            <Route
              path="/collectors"
              element={
                <NotImplementedPage
                  icon="📁"
                  title="Collectors"
                  description="Manage PostgreSQL collectors and data sources"
                />
              }
            />
            <Route
              path="/users"
              element={
                <NotImplementedPage
                  icon="👥"
                  title="User Management"
                  description="Manage system users and access permissions"
                />
              }
            />
            <Route
              path="/settings"
              element={
                <NotImplementedPage
                  icon="⚙️"
                  title="Settings"
                  description="Configure application settings and preferences"
                />
              }
            />
            {/* Catch all */}
            <Route path="*" element={<Navigate to="/" />} />
          </>
        ) : (
          <Route path="*" element={<Navigate to="/login" />} />
        )}
      </Routes>
    </BrowserRouter>
  )
}

export default App
