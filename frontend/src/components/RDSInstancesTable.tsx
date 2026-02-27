import React, { useState, useEffect } from 'react'
import { Trash2, Plus, X, AlertCircle, CheckCircle } from 'lucide-react'
import { CreateRDSForm } from './CreateRDSForm'

interface RDSInstance {
  id: number
  name: string
  description: string
  aws_region: string
  rds_endpoint: string
  port: number
  engine_version: string
  db_instance_class: string
  allocated_storage_gb: number
  environment: string
  master_username: string
  enable_enhanced_monitoring: boolean
  monitoring_interval: number
  ssl_enabled: boolean
  ssl_mode: string
  connection_timeout: number
  is_active: boolean
  last_heartbeat?: string
  last_connection_status: string
  last_error_message?: string
  multi_az: boolean
  backup_retention_days: number
  preferred_backup_window: string
  preferred_maintenance_window: string
  created_at: string
  updated_at: string
}

interface RDSInstancesTableProps {
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

export const RDSInstancesTable: React.FC<RDSInstancesTableProps> = ({ onSuccess, onError }) => {
  const [instances, setInstances] = useState<RDSInstance[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)

  useEffect(() => {
    loadInstances()
  }, [])

  const loadInstances = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/rds-instances', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        throw new Error('Failed to load RDS instances')
      }

      const data = await response.json()
      setInstances(Array.isArray(data) ? data : data.instances || [])
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to load RDS instances')
    } finally {
      setLoading(false)
    }
  }

  const deleteInstance = async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete RDS instance "${name}"?`)) {
      return
    }

    setDeleting(id)
    try {
      const response = await fetch(`/api/v1/rds-instances/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to delete RDS instance')
      }

      setInstances(instances.filter(i => i.id !== id))
      onSuccess(`RDS instance "${name}" deleted successfully`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to delete RDS instance')
    } finally {
      setDeleting(null)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected':
        return 'bg-green-100 text-green-800'
      case 'error':
        return 'bg-red-100 text-red-800'
      case 'unknown':
        return 'bg-gray-100 text-gray-800'
      case 'invalid_credentials':
        return 'bg-yellow-100 text-yellow-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getEnvironmentColor = (env: string) => {
    switch (env) {
      case 'production':
        return 'bg-red-100 text-red-800'
      case 'staging':
        return 'bg-yellow-100 text-yellow-800'
      case 'development':
        return 'bg-blue-100 text-blue-800'
      case 'test':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-600">Loading RDS instances...</p>
      </div>
    )
  }

  return (
    <div>
      {/* Header with Create Button */}
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">RDS Instances</h2>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition"
        >
          {showCreateForm ? (
            <>
              <X size={18} />
              Cancel
            </>
          ) : (
            <>
              <Plus size={18} />
              Register Instance
            </>
          )}
        </button>
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Register New RDS Instance</h3>
            <button
              onClick={() => setShowCreateForm(false)}
              className="text-gray-500 hover:text-gray-700"
            >
              ✕
            </button>
          </div>
          <CreateRDSForm
            onSuccess={(message) => {
              onSuccess(message)
              setTimeout(() => {
                setShowCreateForm(false)
                loadInstances()
              }, 1000)
            }}
            onError={onError}
          />
        </div>
      )}

      {instances.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
          <AlertCircle size={48} className="mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 text-lg">No RDS instances registered yet</p>
          <p className="text-gray-500 text-sm mt-2">Click "Register Instance" to add a new RDS PostgreSQL database for monitoring</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 border border-gray-200 rounded-lg">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Endpoint
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Region
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Environment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Instance Class
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Storage
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {instances.map((instance) => (
                <tr key={instance.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <p className="font-medium text-gray-900">{instance.name}</p>
                      {instance.description && (
                        <p className="text-xs text-gray-500 mt-1">{instance.description}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-gray-600 font-mono">{instance.rds_endpoint}</p>
                    <p className="text-xs text-gray-500">Port: {instance.port}</p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-gray-600">{instance.aws_region}</p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getEnvironmentColor(instance.environment)}`}>
                      {instance.environment.charAt(0).toUpperCase() + instance.environment.slice(1)}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(instance.last_connection_status)}`}>
                        {instance.last_connection_status === 'connected' ? (
                          <span className="flex items-center gap-1">
                            <CheckCircle size={14} />
                            Connected
                          </span>
                        ) : (
                          instance.last_connection_status.charAt(0).toUpperCase() + instance.last_connection_status.slice(1).replace('_', ' ')
                        )}
                      </span>
                      {instance.last_heartbeat && (
                        <p className="text-xs text-gray-500 mt-1">
                          {new Date(instance.last_heartbeat).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-gray-600">{instance.db_instance_class}</p>
                    {instance.engine_version && (
                      <p className="text-xs text-gray-500">v{instance.engine_version}</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <p className="text-sm text-gray-600">{instance.allocated_storage_gb} GB</p>
                    {instance.multi_az && (
                      <p className="text-xs text-green-700 font-medium">Multi-AZ</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <button
                      onClick={() => deleteInstance(instance.id, instance.name)}
                      disabled={deleting === instance.id}
                      className="p-2 text-red-600 hover:text-red-900 hover:bg-red-100 rounded transition disabled:opacity-50"
                      title="Delete instance"
                    >
                      <Trash2 size={18} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
