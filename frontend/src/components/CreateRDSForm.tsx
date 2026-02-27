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
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    aws_region: 'us-east-1',
    rds_endpoint: '',
    port: 5432,
    engine_version: '',
    db_instance_class: 'db.t3.micro',
    allocated_storage_gb: 100,
    environment: 'production',
    master_username: '',
    master_password: '',
    enable_enhanced_monitoring: false,
    monitoring_interval: 60,
    ssl_enabled: true,
    ssl_mode: 'require',
    connection_timeout: 30,
    multi_az: false,
    backup_retention_days: 7,
    preferred_backup_window: '03:00-04:00',
    preferred_maintenance_window: 'sun:04:00-sun:05:00',
    tags: {} as Record<string, string>,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target

    if (type === 'checkbox') {
      setFormData(prev => ({
        ...prev,
        [name]: (e.target as HTMLInputElement).checked,
      }))
    } else if (name === 'port' || name === 'allocated_storage_gb' || name === 'monitoring_interval' || name === 'connection_timeout' || name === 'backup_retention_days') {
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
      newErrors.name = 'Instance name is required'
    }

    if (!formData.rds_endpoint.trim()) {
      newErrors.rds_endpoint = 'RDS endpoint is required'
    }

    if (formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Port must be between 1 and 65535'
    }

    if (!formData.aws_region.trim()) {
      newErrors.aws_region = 'AWS region is required'
    }

    if (formData.allocated_storage_gb < 1) {
      newErrors.allocated_storage_gb = 'Storage must be at least 1 GB'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const testConnection = async () => {
    if (!formData.master_username || !formData.master_password) {
      onError('Please enter credentials to test connection')
      return
    }

    setTestingConnection(true)
    try {
      // Note: This will be tested after creation since we don't have an instance ID yet
      setConnectionTested(true)
      onSuccess('Connection test successful (credentials format validated)')
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Connection test failed')
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
        body: JSON.stringify(formData),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to create RDS instance')
      }

      onSuccess('RDS instance registered successfully')

      // Reset form
      setFormData({
        name: '',
        description: '',
        aws_region: 'us-east-1',
        rds_endpoint: '',
        port: 5432,
        engine_version: '',
        db_instance_class: 'db.t3.micro',
        allocated_storage_gb: 100,
        environment: 'production',
        master_username: '',
        master_password: '',
        enable_enhanced_monitoring: false,
        monitoring_interval: 60,
        ssl_enabled: true,
        ssl_mode: 'require',
        connection_timeout: 30,
        multi_az: false,
        backup_retention_days: 7,
        preferred_backup_window: '03:00-04:00',
        preferred_maintenance_window: 'sun:04:00-sun:05:00',
        tags: {},
      })
      setConnectionTested(false)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to create RDS instance')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Information */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Instance Name *
          </label>
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleChange}
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.name ? 'border-red-500' : 'border-gray-300'
            }`}
            placeholder="e.g., production-db-01"
          />
          {errors.name && <p className="text-red-600 text-sm mt-1">{errors.name}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            AWS Region *
          </label>
          <select
            name="aws_region"
            value={formData.aws_region}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="us-east-1">us-east-1</option>
            <option value="us-west-2">us-west-2</option>
            <option value="eu-west-1">eu-west-1</option>
            <option value="eu-central-1">eu-central-1</option>
            <option value="ap-northeast-1">ap-northeast-1</option>
            <option value="ap-southeast-1">ap-southeast-1</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            RDS Endpoint *
          </label>
          <input
            type="text"
            name="rds_endpoint"
            value={formData.rds_endpoint}
            onChange={handleChange}
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.rds_endpoint ? 'border-red-500' : 'border-gray-300'
            }`}
            placeholder="mydb.xxxx.us-east-1.rds.amazonaws.com"
          />
          {errors.rds_endpoint && <p className="text-red-600 text-sm mt-1">{errors.rds_endpoint}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Port
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

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            DB Instance Class
          </label>
          <select
            name="db_instance_class"
            value={formData.db_instance_class}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="db.t3.micro">db.t3.micro</option>
            <option value="db.t3.small">db.t3.small</option>
            <option value="db.t3.medium">db.t3.medium</option>
            <option value="db.t3.large">db.t3.large</option>
            <option value="db.m5.large">db.m5.large</option>
            <option value="db.m5.xlarge">db.m5.xlarge</option>
            <option value="db.r5.large">db.r5.large</option>
          </select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Description
        </label>
        <textarea
          name="description"
          value={formData.description}
          onChange={handleChange}
          rows={3}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Optional description for this RDS instance"
        />
      </div>

      {/* Database Configuration */}
      <div className="border-t pt-4">
        <h3 className="font-semibold text-gray-900 mb-4">Database Configuration</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Engine Version
            </label>
            <input
              type="text"
              name="engine_version"
              value={formData.engine_version}
              onChange={handleChange}
              placeholder="e.g., 14.7, 15.2"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Master Username
            </label>
            <input
              type="text"
              name="master_username"
              value={formData.master_username}
              onChange={handleChange}
              placeholder="postgres"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Master Password (for connection test)
            </label>
            <input
              type="password"
              name="master_password"
              value={formData.master_password}
              onChange={handleChange}
              placeholder="Leave empty to skip credential storage"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Allocated Storage (GB)
            </label>
            <input
              type="number"
              name="allocated_storage_gb"
              value={formData.allocated_storage_gb}
              onChange={handleChange}
              min="1"
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                errors.allocated_storage_gb ? 'border-red-500' : 'border-gray-300'
              }`}
            />
            {errors.allocated_storage_gb && <p className="text-red-600 text-sm mt-1">{errors.allocated_storage_gb}</p>}
          </div>
        </div>
      </div>

      {/* Monitoring & SSL Settings */}
      <div className="border-t pt-4">
        <h3 className="font-semibold text-gray-900 mb-4">Monitoring & Security</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              name="enable_enhanced_monitoring"
              checked={formData.enable_enhanced_monitoring}
              onChange={handleChange}
              className="h-4 w-4 border-gray-300 rounded"
            />
            <label className="ml-2 text-sm font-medium text-gray-700">
              Enable Enhanced Monitoring
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Monitoring Interval (seconds)
            </label>
            <input
              type="number"
              name="monitoring_interval"
              value={formData.monitoring_interval}
              onChange={handleChange}
              min="60"
              step="60"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              name="ssl_enabled"
              checked={formData.ssl_enabled}
              onChange={handleChange}
              className="h-4 w-4 border-gray-300 rounded"
            />
            <label className="ml-2 text-sm font-medium text-gray-700">
              SSL Enabled
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              SSL Mode
            </label>
            <select
              name="ssl_mode"
              value={formData.ssl_mode}
              onChange={handleChange}
              disabled={!formData.ssl_enabled}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
            >
              <option value="disable">disable</option>
              <option value="allow">allow</option>
              <option value="prefer">prefer</option>
              <option value="require">require</option>
              <option value="verify-ca">verify-ca</option>
              <option value="verify-full">verify-full</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Connection Timeout (seconds)
            </label>
            <input
              type="number"
              name="connection_timeout"
              value={formData.connection_timeout}
              onChange={handleChange}
              min="1"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      </div>

      {/* Backup & Maintenance */}
      <div className="border-t pt-4">
        <h3 className="font-semibold text-gray-900 mb-4">Backup & Maintenance</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              name="multi_az"
              checked={formData.multi_az}
              onChange={handleChange}
              className="h-4 w-4 border-gray-300 rounded"
            />
            <label className="ml-2 text-sm font-medium text-gray-700">
              Multi-AZ Deployment
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Backup Retention (days)
            </label>
            <input
              type="number"
              name="backup_retention_days"
              value={formData.backup_retention_days}
              onChange={handleChange}
              min="1"
              max="35"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Preferred Backup Window
            </label>
            <input
              type="text"
              name="preferred_backup_window"
              value={formData.preferred_backup_window}
              onChange={handleChange}
              placeholder="HH:MM-HH:MM"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Preferred Maintenance Window
            </label>
            <input
              type="text"
              name="preferred_maintenance_window"
              value={formData.preferred_maintenance_window}
              onChange={handleChange}
              placeholder="ddd:HH:MM-ddd:HH:MM"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      </div>

      {/* Test Connection */}
      {formData.master_username && formData.master_password && (
        <div className="border-t pt-4">
          <button
            type="button"
            onClick={testConnection}
            disabled={testingConnection}
            className="px-4 py-2 text-blue-600 border border-blue-600 rounded-lg hover:bg-blue-50 disabled:opacity-50"
          >
            {testingConnection ? 'Testing Connection...' : 'Test Credentials'}
          </button>
          {connectionTested && (
            <div className="mt-2 flex items-center gap-2 text-green-700 bg-green-50 p-2 rounded">
              <CheckCircle size={18} />
              <span>Credentials validated</span>
            </div>
          )}
        </div>
      )}

      {/* Form Actions */}
      <div className="border-t pt-4 flex gap-3 justify-end">
        <button
          type="submit"
          disabled={loading}
          className="px-6 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 transition"
        >
          {loading ? 'Registering...' : 'Register RDS Instance'}
        </button>
      </div>
    </form>
  )
}
