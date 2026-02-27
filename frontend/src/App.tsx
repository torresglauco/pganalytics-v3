import React, { useState, useEffect } from 'react'
import { Dashboard } from './pages/Dashboard'
import { apiClient } from './services/api'
import './styles/index.css'

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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
          setIsLoading(false)
        }
      } catch (err) {
        console.error('Auto-login failed:', err)
        setError('Failed to authenticate. Please refresh the page.')
        setIsLoading(false)
      }
    }

    // Only auto-login if not already authenticated
    if (!apiClient.isAuthenticated()) {
      autoLogin()
    } else {
      setIsLoading(false)
    }
  }, [])

  const handleLogout = () => {
    apiClient.logout()
    // Auto-login again after logout
    window.location.reload()
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

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <div className="bg-white rounded-lg shadow-md p-6 max-w-md">
          <h1 className="text-xl font-bold text-red-600 mb-4">Authentication Error</h1>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return <Dashboard onLogout={handleLogout} />
}

export default App
