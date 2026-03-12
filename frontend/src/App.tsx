import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Dashboard } from './components/dashboard/Dashboard'
import { LoginPage } from './components/auth/LoginPage'
import { LogsPage } from './pages/LogsPage'
import { MetricsPage } from './pages/MetricsPage'
import { AlertsPage } from './pages/AlertsPage'
import { ChannelsPage } from './pages/ChannelsPage'
import { useAuthStore } from './stores/authStore'
import { apiClient } from './services/api'
import { LoadingSpinner } from './components/ui/LoadingSpinner'
import './styles/index.css'

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const { isAuthenticated, setAuthenticated } = useAuthStore()

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
