import React, { useState } from 'react'
import { Eye, EyeOff, Loader } from 'lucide-react'
import { apiClient } from '../services/api'

interface LoginFormProps {
  onSuccess: (message: string) => void
  onError: (error: Error) => void
  onSwitchToSignup?: () => void
  onLogin?: (username: string, password: string) => Promise<void>
}

export const LoginForm: React.FC<LoginFormProps> = ({
  onSuccess,
  onError,
  onSwitchToSignup,
  onLogin: externalLogin,
}) => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!username.trim()) {
      newErrors.username = 'Username is required'
    }

    if (!password) {
      newErrors.password = 'Password is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    setLoading(true)
    try {
      if (externalLogin) {
        // Use external login handler (from App.tsx)
        await externalLogin(username, password)
        onSuccess(`Welcome back, ${username}!`)
      } else {
        // Use apiClient directly (legacy behavior)
        const response = await apiClient.login(username, password)
        onSuccess(`Welcome back, ${response.user.username}!`)
      }
      // Reset form
      setUsername('')
      setPassword('')
    } catch (error) {
      if (error instanceof Error) {
        onError(error)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700">
          Username
        </label>
        <input
          type="text"
          id="username"
          value={username}
          onChange={(e) => {
            setUsername(e.target.value)
            if (errors.username) {
              setErrors((prev) => {
                const newErrors = { ...prev }
                delete newErrors.username
                return newErrors
              })
            }
          }}
          placeholder="Enter your username"
          className={`mt-1 w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
            errors.username ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.username && (
          <p className="mt-1 text-sm text-red-600">{errors.username}</p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium text-gray-700">
          Password
        </label>
        <div className="mt-1 relative">
          <input
            type={showPassword ? 'text' : 'password'}
            id="password"
            value={password}
            onChange={(e) => {
              setPassword(e.target.value)
              if (errors.password) {
                setErrors((prev) => {
                  const newErrors = { ...prev }
                  delete newErrors.password
                  return newErrors
                })
              }
            }}
            placeholder="Enter your password"
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.password ? 'border-red-500' : 'border-gray-300'
            }`}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-2.5 text-gray-500 hover:text-gray-700"
          >
            {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
          </button>
        </div>
        {errors.password && (
          <p className="mt-1 text-sm text-red-600">{errors.password}</p>
        )}
      </div>

      <button
        type="submit"
        disabled={loading}
        className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed font-medium flex items-center justify-center gap-2"
      >
        {loading && <Loader size={18} className="animate-spin" />}
        {loading ? 'Logging in...' : 'Log In'}
      </button>

      {onSwitchToSignup && (
        <div className="mt-6 pt-6 border-t border-gray-200">
          <p className="text-sm text-gray-600 text-center">
            Don't have an account?{' '}
            <button
              type="button"
              onClick={onSwitchToSignup}
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Sign up here
            </button>
          </p>
        </div>
      )}
    </form>
  )
}
