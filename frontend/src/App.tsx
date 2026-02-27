import React, { useState, useEffect } from 'react'
import { Dashboard } from './pages/Dashboard'
import { AuthPage } from './pages/AuthPage'
import { apiClient } from './services/api'
import './styles/index.css'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(apiClient.isAuthenticated())

  useEffect(() => {
    // Check authentication on mount
    setIsAuthenticated(apiClient.isAuthenticated())
  }, [])

  const handleAuthSuccess = () => {
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    apiClient.logout()
    setIsAuthenticated(false)
  }

  if (!isAuthenticated) {
    return <AuthPage onAuthSuccess={handleAuthSuccess} />
  }

  return <Dashboard onLogout={handleLogout} />
}

export default App
