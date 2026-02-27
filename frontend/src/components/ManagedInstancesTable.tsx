import React, { useState, useEffect } from 'react'
import { Trash2, Plus, X, AlertCircle, CheckCircle, Edit, Zap } from 'lucide-react'
import { CreateManagedInstanceForm } from './CreateManagedInstanceForm'

interface ManagedInstance {
  id: number
  name: string
  description: string
  aws_region: string
  endpoint: string
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
  status: string
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

interface ManagedInstancesTableProps {
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

export const ManagedInstancesTable: React.FC<ManagedInstancesTableProps> = ({ onSuccess, onError }) => {
  const [instances, setInstances] = useState<ManagedInstance[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [testingConnectionId, setTestingConnectionId] = useState<number | null>(null)
  const [editFormData, setEditFormData] = useState<Partial<ManagedInstance> | null>(null)

  useEffect(() => {
    loadInstances()
  }, [])

  const loadInstances = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/managed-instances', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        throw new Error('Failed to load managed instances')
      }

      const data = await response.json()
      setInstances(Array.isArray(data) ? data : data.instances || [])
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to load managed instances')
    } finally {
      setLoading(false)
    }
  }

  const deleteInstance = async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete managed instance "${name}"?`)) {
      return
    }

    setDeleting(id)
    try {
      const response = await fetch(`/api/v1/managed-instances/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to delete managed instance')
      }

      setInstances(instances.filter(i => i.id !== id))
      onSuccess(`Managed instance "${name}" deleted successfully`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to delete managed instance')
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

  const getInstanceStatusColor = (status: string) => {
    switch (status) {
      case 'monitoring':
        return 'bg-green-100 text-green-800'
      case 'registered':
        return 'bg-blue-100 text-blue-800'
      case 'registering':
        return 'bg-yellow-100 text-yellow-800'
      case 'paused':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const startEdit = (instance: ManagedInstance) => {
    setEditingId(instance.id)
    setEditFormData(instance)
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditFormData(null)
  }

  const saveEdit = async () => {
    if (!editFormData || !editingId) return

    try {
      const response = await fetch(`/api/v1/managed-instances/${editingId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({
          name: editFormData.name,
          aws_region: editFormData.aws_region || 'us-east-1',
          endpoint: editFormData.endpoint,
          port: editFormData.port,
          environment: editFormData.environment,
          master_username: editFormData.master_username,
          master_password: editFormData.master_username, // Keep same for edit
          description: editFormData.description,
          status: editFormData.status,
          engine_version: editFormData.engine_version || '',
          db_instance_class: editFormData.db_instance_class || '',
          allocated_storage_gb: editFormData.allocated_storage_gb || 0,
          enable_enhanced_monitoring: editFormData.enable_enhanced_monitoring || false,
          monitoring_interval: editFormData.monitoring_interval || 60,
          ssl_enabled: editFormData.ssl_enabled || true,
          ssl_mode: editFormData.ssl_mode || 'require',
          connection_timeout: editFormData.connection_timeout || 30,
          multi_az: editFormData.multi_az || false,
          backup_retention_days: editFormData.backup_retention_days || 0,
          preferred_backup_window: editFormData.preferred_backup_window || '',
          preferred_maintenance_window: editFormData.preferred_maintenance_window || '',
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to update managed instance')
      }

      onSuccess('Managed instance updated successfully')
      cancelEdit()
      loadInstances()
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to update managed instance')
    }
  }

  const testConnection = async (id: number) => {
    setTestingConnectionId(id)
    const instance = instances.find(i => i.id === id)
    if (!instance) return

    try {
      const response = await fetch(`/api/v1/managed-instances/${id}/test-connection`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({
          // Empty body - backend will use stored credentials from the managed instance record
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Connection test failed')
      }

      const data = await response.json()
      if (data.success) {
        onSuccess(`✓ Connection successful for ${instance.name}`)
        // Reload instances to update the connection status in the UI
        loadInstances()
      } else {
        onError(`${data.error || 'Connection test failed'}`)
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Failed to test connection'
      onError(`Connection test error: ${errorMsg}`)
    } finally {
      setTestingConnectionId(null)
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
        <p className="text-gray-600">Loading managed instances...</p>
      </div>
    )
  }

  return (
    <div>
      {/* Header with Create Button */}
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Managed Instances</h2>
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

      {/* Edit Modal */}
      {editingId !== null && editFormData && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Edit Managed Instance</h3>
              <button
                onClick={cancelEdit}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Instance Name</label>
                  <input
                    type="text"
                    value={editFormData.name || ''}
                    onChange={(e) => setEditFormData({...editFormData, name: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                  <select
                    value={editFormData.status || 'registered'}
                    onChange={(e) => setEditFormData({...editFormData, status: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  >
                    <option value="registering">Registering</option>
                    <option value="registered">Registered</option>
                    <option value="monitoring">Monitoring</option>
                    <option value="paused">Paused</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Environment</label>
                  <select
                    value={editFormData.environment || 'production'}
                    onChange={(e) => setEditFormData({...editFormData, environment: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  >
                    <option value="production">Production</option>
                    <option value="staging">Staging</option>
                    <option value="development">Development</option>
                    <option value="test">Test</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint</label>
                  <input
                    type="text"
                    value={editFormData.endpoint || ''}
                    onChange={(e) => setEditFormData({...editFormData, endpoint: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
                  <input
                    type="number"
                    value={editFormData.port || 5432}
                    onChange={(e) => setEditFormData({...editFormData, port: parseInt(e.target.value)})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                  <input
                    type="text"
                    value={editFormData.master_username || ''}
                    onChange={(e) => setEditFormData({...editFormData, master_username: e.target.value})}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={saveEdit}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition"
                >
                  Save Changes
                </button>
                <button
                  onClick={cancelEdit}
                  className="flex-1 px-4 py-2 bg-gray-300 text-gray-800 font-medium rounded-lg hover:bg-gray-400 transition"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create Form */}
      {showCreateForm && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Register New Managed Instance</h3>
            <button
              onClick={() => setShowCreateForm(false)}
              className="text-gray-500 hover:text-gray-700"
            >
              ✕
            </button>
          </div>
          <CreateManagedInstanceForm
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
          <p className="text-gray-600 text-lg">No managed instances registered yet</p>
          <p className="text-gray-500 text-sm mt-2">Click "Register Instance" to add a new PostgreSQL database for monitoring</p>
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
                  Instance Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Connection Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
                  Instance Class
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
                    <p className="text-sm text-gray-600 font-mono">{instance.endpoint}</p>
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
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getInstanceStatusColor(instance.status)}`}>
                      {instance.status.charAt(0).toUpperCase() + instance.status.slice(1)}
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
                    <p className="text-sm text-gray-600">{instance.db_instance_class || '-'}</p>
                    {instance.engine_version && (
                      <p className="text-xs text-gray-500">v{instance.engine_version}</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm flex gap-2">
                    <button
                      onClick={() => startEdit(instance)}
                      className="p-2 text-blue-600 hover:text-blue-900 hover:bg-blue-100 rounded transition"
                      title="Edit instance"
                    >
                      <Edit size={18} />
                    </button>
                    <button
                      onClick={() => testConnection(instance.id)}
                      disabled={testingConnectionId === instance.id}
                      className="p-2 text-yellow-600 hover:text-yellow-900 hover:bg-yellow-100 rounded transition disabled:opacity-50"
                      title="Test connection"
                    >
                      <Zap size={18} />
                    </button>
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
