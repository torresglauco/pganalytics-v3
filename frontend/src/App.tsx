import React, { useState, useEffect } from 'react'
import { Dashboard } from './pages/Dashboard'
import { AuthPage } from './pages/AuthPage'
import { apiClient } from './services/api'
import './styles/index.css'

// Test: GitHub Actions workflow verification
// This change triggers E2E Tests, Frontend Quality, and Security Scanning workflows

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    // Check if user is already authenticated via stored token
    const checkAuthentication = async () => {
      try {
        setIsLoading(true)
        setError(null)

        if (apiClient.isAuthenticated()) {
          // Token exists, user is authenticated
          setIsAuthenticated(true)
        } else {
          // No token, user needs to login
          setIsAuthenticated(false)
        }
      } catch (err) {
        console.error('Auth check failed:', err)
        setIsAuthenticated(false)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuthentication()
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
      <div className="flex items-center justify-center min-h-screen bg-pg-light">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-pg-cyan mx-auto mb-4"></div>
          <p className="text-pg-slate">Loading application...</p>
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
