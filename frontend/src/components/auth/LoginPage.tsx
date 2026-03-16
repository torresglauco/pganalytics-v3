import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { LoadingSpinner } from '../ui/LoadingSpinner'
import { useAuthStore } from '../../stores/authStore'
import { apiClient } from '../../services/api'

export const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { setUser, setToken, setError, setLoading, error, isLoading } = useAuthStore()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [usernameError, setUsernameError] = useState('')
  const [passwordError, setPasswordError] = useState('')

  const validateForm = () => {
    let isValid = true

    if (!username) {
      setUsernameError('Username is required')
      isValid = false
    } else {
      setUsernameError('')
    }

    if (!password) {
      setPasswordError('Password is required')
      isValid = false
    } else {
      setPasswordError('')
    }

    return isValid
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    try {
      setLoading(true)
      setError(null)

      const response = await apiClient.login(username, password)

      setToken(response.token)
      // Type assertion - the User type from both modules represents the same data
      setUser(response.user as any)
      navigate('/')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Logging in..." />
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-slate-50 dark:from-slate-900 dark:to-slate-800 flex items-center justify-center p-4">
      {/* Left Side - Brand */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-center px-12">
        <div className="mb-8">
          <div className="w-16 h-16 bg-primary-600 rounded-xl mb-4" />
          <h1 className="text-4xl font-bold text-slate-900 dark:text-slate-100 mb-2">
            pgAnalytics
          </h1>
          <p className="text-lg text-slate-600 dark:text-slate-400">
            PostgreSQL Observability Platform
          </p>
        </div>

        <div className="space-y-6">
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              Monitor in Real-Time
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Track PostgreSQL logs, metrics, and alerts in a unified dashboard
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              Deep Analysis
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Find slow queries, errors, and performance issues instantly
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              Proactive Alerting
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Get notified before problems impact your users
            </p>
          </div>
        </div>
      </div>

      {/* Right Side - Form */}
      <div className="w-full lg:w-1/2 max-w-md">
        <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 border border-slate-200 dark:border-slate-700">
          <h2 className="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">
            Sign In
          </h2>
          <p className="text-slate-600 dark:text-slate-400 mb-6">
            Welcome back to pgAnalytics
          </p>

          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Username"
              type="text"
              placeholder="admin"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              error={usernameError}
              required
            />

            <Input
              label="Password"
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              error={passwordError}
              required
            />

            <div className="flex items-center">
              <input
                type="checkbox"
                id="remember"
                className="w-4 h-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
              />
              <label htmlFor="remember" className="ml-2 text-sm text-slate-600 dark:text-slate-400">
                Remember me
              </label>
            </div>

            <Button type="submit" fullWidth size="lg">
              Sign In
            </Button>
          </form>

          <div className="mt-6">
            <div className="relative mb-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200 dark:border-slate-700" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                  Or continue with
                </span>
              </div>
            </div>

            <Button variant="secondary" fullWidth size="md">
              SSO Login
            </Button>
          </div>

          <p className="mt-6 text-center text-slate-600 dark:text-slate-400">
            Don't have an account?{' '}
            <a href="/signup" className="text-primary-600 hover:text-primary-700 font-medium">
              Sign up
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}

LoginPage.displayName = 'LoginPage'
