import React, { useState } from 'react'
import { AlertCircle, CheckCircle } from 'lucide-react'
import { SignupForm } from '../components/SignupForm'
import { LoginForm } from '../components/LoginForm'
import type { ApiError } from '../types'

interface AuthPageProps {
  onAuthSuccess: () => void
}

export const AuthPage: React.FC<AuthPageProps> = ({ onAuthSuccess }) => {
  const [isSignup, setIsSignup] = useState(true)
  const [successMessage, setSuccessMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  const handleSuccess = (message: string) => {
    setSuccessMessage(message)
    setErrorMessage('')
    // Redirect after 2 seconds
    setTimeout(() => {
      onAuthSuccess()
    }, 2000)
  }

  const handleError = (error: Error | ApiError) => {
    let message = 'An error occurred'

    if ('message' in error) {
      message = error.message
    } else if (error instanceof Error) {
      message = error.message
    }

    setErrorMessage(message)
    setSuccessMessage('')
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">pgAnalytics</h1>
          <p className="text-gray-600">PostgreSQL Performance Analytics</p>
        </div>

        {/* Success Message */}
        {successMessage && (
          <div className="mb-6 bg-green-50 border border-green-200 rounded-lg p-4 flex gap-3">
            <CheckCircle className="text-green-600 flex-shrink-0 mt-0.5" size={20} />
            <div>
              <h3 className="font-medium text-green-900">Success!</h3>
              <p className="text-sm text-green-700 mt-1">{successMessage}</p>
            </div>
          </div>
        )}

        {/* Error Message */}
        {errorMessage && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4 flex gap-3">
            <AlertCircle className="text-red-600 flex-shrink-0 mt-0.5" size={20} />
            <div>
              <h3 className="font-medium text-red-900">Error</h3>
              <p className="text-sm text-red-700 mt-1">{errorMessage}</p>
            </div>
          </div>
        )}

        {/* Auth Card */}
        <div className="bg-white rounded-lg shadow-lg p-8">
          {/* Tab Header */}
          <div className="flex gap-4 mb-6 border-b border-gray-200">
            <button
              onClick={() => {
                setIsSignup(false)
                setErrorMessage('')
                setSuccessMessage('')
              }}
              className={`px-4 py-2 font-medium text-sm border-b-2 transition ${
                !isSignup
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-600 hover:text-gray-900'
              }`}
            >
              Log In
            </button>
            <button
              onClick={() => {
                setIsSignup(true)
                setErrorMessage('')
                setSuccessMessage('')
              }}
              className={`px-4 py-2 font-medium text-sm border-b-2 transition ${
                isSignup
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-600 hover:text-gray-900'
              }`}
            >
              Sign Up
            </button>
          </div>

          {/* Form Content */}
          {isSignup ? (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Create Account</h2>
              <SignupForm
                onSuccess={handleSuccess}
                onError={handleError}
                onSwitchToLogin={() => setIsSignup(false)}
              />
            </div>
          ) : (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Welcome Back</h2>
              <LoginForm
                onSuccess={handleSuccess}
                onError={handleError}
                onSwitchToSignup={() => setIsSignup(true)}
              />
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-600">
          <p>© 2026 pgAnalytics. All rights reserved.</p>
          <p className="mt-1">v3.3.0 - PostgreSQL Performance Analytics Platform</p>
        </div>
      </div>
    </div>
  )
}
