import React, { useState, useEffect } from 'react'
import { Dashboard } from './pages/Dashboard'
import { AuthPage } from './pages/AuthPage'
import { apiClient } from './services/api'
import './styles/index.css'

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    // Auto-login with admin credentials
    const autoLogin = async () => {
      try {
        setIsLoading(true)
        setError(null)

        // Try to login with admin credentials
        const response = await apiClient.login('admin', 'admin')
        if (response) {
          // Login successful
          setIsAuthenticated(true)
          setIsLoading(false)
        }
      } catch (err) {
        console.error('Auto-login failed:', err)
        setError('Failed to authenticate. Please try logging in manually.')
        setIsLoading(false)
      }
    }

    // Only auto-login if not already authenticated
    if (!apiClient.isAuthenticated()) {
      autoLogin()
    } else {
      setIsAuthenticated(true)
      setIsLoading(false)
    }
  }, [])

  const handleLogin = async (username: string, password: string) => {
    try {
      setIsLoading(true)
      setError(null)
      const response = await apiClient.login(username, password)
      if (response) {
        setIsAuthenticated(true)
        setIsLoading(false)
      }
    } catch (err) {
      console.error('Login failed:', err)
      setError(err instanceof Error ? err.message : 'Login failed')
      setIsLoading(false)
    }
  }

  const handleLogout = () => {
    apiClient.logout()
    setIsAuthenticated(false)
    setError(null)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading application...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <AuthPage onLogin={handleLogin} error={error} />
  }

  return <Dashboard onLogout={handleLogout} />
}

export default App
