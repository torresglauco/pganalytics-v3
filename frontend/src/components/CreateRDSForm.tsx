import React, { useState } from 'react'
import { AlertCircle, CheckCircle } from 'lucide-react'

interface CreateRDSFormProps {
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

export const CreateRDSForm: React.FC<CreateRDSFormProps> = ({ onSuccess, onError }) => {
  const [loading, setLoading] = useState(false)
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionTested, setConnectionTested] = useState(false)
  const [connectionError, setConnectionError] = useState('')
  const [formData, setFormData] = useState({
    name: '',
    rds_endpoint: '',
    port: 5432,
    environment: 'production',
    master_username: '',
    master_password: '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target

    if (type === 'number') {
      setFormData(prev => ({
        ...prev,
        [name]: parseInt(value) || 0,
      }))
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value,
      }))
    }

    // Clear error for this field
    if (errors[name]) {
      setErrors(prev => {
        const newErrors = { ...prev }
        delete newErrors[name]
        return newErrors
      })
    }
  }

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Cluster name is required'
    }

    if (!formData.rds_endpoint.trim()) {
      newErrors.rds_endpoint = 'RDS endpoint is required'
    }

    if (formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Port must be between 1 and 65535'
    }

    if (!formData.master_username.trim()) {
      newErrors.master_username = 'Username is required'
    }

    if (!formData.master_password.trim()) {
      newErrors.master_password = 'Password is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const testConnection = async () => {
    if (!formData.rds_endpoint || !formData.master_username || !formData.master_password) {
      onError('Please enter RDS endpoint, username, and password to test connection')
      return
    }

    setTestingConnection(true)
    setConnectionError('')
    try {
      const token = localStorage.getItem('auth_token')
      if (!token) {
        onError('Authentication required')
        setTestingConnection(false)
        return
      }

      const response = await fetch('/api/v1/rds-instances/test-connection-direct', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          rds_endpoint: formData.rds_endpoint,
          port: formData.port,
          username: formData.master_username,
          password: formData.master_password,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Connection test failed')
      }

      const data = await response.json()
      if (data.success) {
        setConnectionTested(true)
        setConnectionError('')
        onSuccess('✓ Connection successful! Credentials are valid.')
      } else {
        setConnectionError(data.error || 'Connection test failed')
        onError(`Connection test failed: ${data.error || 'Unknown error'}`)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Connection test failed'
      setConnectionError(errorMsg)
      onError(errorMsg)
    } finally {
      setTestingConnection(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    setLoading(true)
    try {
      const token = localStorage.getItem('auth_token')
      if (!token) {
        onError('Authentication required')
        setLoading(false)
        return
      }

      const response = await fetch('/api/v1/rds-instances', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          ...formData,
          aws_region: 'us-east-1', // Default region extracted from RDS endpoint
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to register RDS instance')
      }

      onSuccess('RDS instance registered successfully')

      // Reset form
      setFormData({
        name: '',
        rds_endpoint: '',
        port: 5432,
        environment: 'production',
        master_username: '',
        master_password: '',
      })
      setConnectionTested(false)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to register RDS instance')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <p className="text-sm text-gray-600 mb-4">
        Enter your RDS PostgreSQL connection details to begin monitoring.
      </p>

      {/* Cluster Name */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Cluster Name *
        </label>
        <input
          type="text"
          name="name"
          value={formData.name}
          onChange={handleChange}
          placeholder="e.g., production-db, staging-db"
          className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
            errors.name ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.name && <p className="text-red-600 text-sm mt-1">{errors.name}</p>}
      </div>

      {/* RDS Endpoint */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          RDS Endpoint *
        </label>
        <input
          type="text"
          name="rds_endpoint"
          value={formData.rds_endpoint}
          onChange={handleChange}
          placeholder="mydb.xxxx.us-east-1.rds.amazonaws.com"
          className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
            errors.rds_endpoint ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.rds_endpoint && <p className="text-red-600 text-sm mt-1">{errors.rds_endpoint}</p>}
      </div>

      {/* Port */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Port *
        </label>
        <input
          type="number"
          name="port"
          value={formData.port}
          onChange={handleChange}
          min="1"
          max="65535"
          className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
            errors.port ? 'border-red-500' : 'border-gray-300'
          }`}
        />
        {errors.port && <p className="text-red-600 text-sm mt-1">{errors.port}</p>}
      </div>

      {/* Environment */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Environment
        </label>
        <select
          name="environment"
          value={formData.environment}
          onChange={handleChange}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="production">Production</option>
          <option value="staging">Staging</option>
          <option value="development">Development</option>
          <option value="test">Test</option>
        </select>
      </div>

      {/* Username and Password */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Username *
          </label>
          <input
            type="text"
            name="master_username"
            value={formData.master_username}
            onChange={handleChange}
            placeholder="postgres"
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.master_username ? 'border-red-500' : 'border-gray-300'
            }`}
          />
          {errors.master_username && <p className="text-red-600 text-sm mt-1">{errors.master_username}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Password *
          </label>
          <input
            type="password"
            name="master_password"
            value={formData.master_password}
            onChange={handleChange}
            placeholder="••••••••"
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.master_password ? 'border-red-500' : 'border-gray-300'
            }`}
          />
          {errors.master_password && <p className="text-red-600 text-sm mt-1">{errors.master_password}</p>}
        </div>
      </div>

      {/* Test Connection Button */}
      <div className="flex gap-3 items-center">
        <button
          type="button"
          onClick={testConnection}
          disabled={testingConnection || !formData.master_username || !formData.master_password}
          className="px-4 py-2 text-blue-600 border border-blue-600 rounded-lg hover:bg-blue-50 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {testingConnection ? 'Testing...' : 'Test Connection'}
        </button>
        {connectionTested && (
          <div className="flex items-center gap-2 text-green-700">
            <CheckCircle size={18} />
            <span className="text-sm">Credentials valid</span>
          </div>
        )}
        {connectionError && (
          <div className="flex items-center gap-2 text-red-700">
            <AlertCircle size={18} />
            <span className="text-sm">{connectionError}</span>
          </div>
        )}
      </div>

      {/* Form Actions */}
      <div className="border-t pt-4 flex gap-3 justify-end">
        <button
          type="submit"
          disabled={loading}
          className="px-6 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 transition"
        >
          {loading ? 'Registering...' : 'Register Instance'}
        </button>
      </div>
    </form>
  )
}
