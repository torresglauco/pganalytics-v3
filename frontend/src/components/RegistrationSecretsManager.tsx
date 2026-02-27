import React, { useState, useEffect } from 'react'
import { Plus, Edit2, Trash2, Eye, EyeOff, Copy, Check, AlertCircle } from 'lucide-react'
import { apiClient } from '../services/api'

interface RegistrationSecret {
  id: string
  name: string
  description?: string
  active: boolean
  created_at: string
  updated_at: string
  total_registrations: number
  last_used_at?: string
  created_by_username?: string
}

interface CreateSecretResponse extends RegistrationSecret {
  secret_value: string
  message: string
}

export const RegistrationSecretsManager: React.FC = () => {
  const [secrets, setSecrets] = useState<RegistrationSecret[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingSecret, setEditingSecret] = useState<RegistrationSecret | null>(null)
  const [copiedSecretId, setCopiedSecretId] = useState<string | null>(null)
  const [newSecretValue, setNewSecretValue] = useState<string | null>(null)
  const [showNewSecret, setShowNewSecret] = useState(false)
  const [successMessage, setSuccessMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')
  const [formData, setFormData] = useState({ name: '', description: '' })
  const [editFormData, setEditFormData] = useState({ name: '', description: '', active: true })

  useEffect(() => {
    loadSecrets()
  }, [])

  const loadSecrets = async () => {
    try {
      setLoading(true)
      const response = await fetch(`${apiClient.getBaseURL()}/registration-secrets`, {
        headers: {
          'Authorization': `Bearer ${apiClient.getToken()}`,
        },
      })
      if (response.ok) {
        const data = await response.json()
        setSecrets(data.secrets || [])
      } else {
        setErrorMessage('Failed to load registration secrets')
      }
    } catch (error) {
      setErrorMessage('Error loading registration secrets')
      console.error('Error loading secrets:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateSecret = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await fetch(`${apiClient.getBaseURL()}/registration-secrets`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${apiClient.getToken()}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
        }),
      })
      if (response.ok) {
        const data: CreateSecretResponse = await response.json()
        setNewSecretValue(data.secret_value)
        setShowNewSecret(true)
        setFormData({ name: '', description: '' })
        setShowCreateModal(false)
        setSuccessMessage(data.message)
        setTimeout(() => setSuccessMessage(''), 5000)
        loadSecrets()
      } else {
        setErrorMessage('Failed to create registration secret')
      }
    } catch (error) {
      setErrorMessage('Error creating registration secret')
      console.error('Error creating secret:', error)
    }
  }

  const handleUpdateSecret = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingSecret) return

    try {
      const response = await fetch(
        `${apiClient.getBaseURL()}/registration-secrets/${editingSecret.id}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${apiClient.getToken()}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            name: editFormData.name,
            description: editFormData.description,
            active: editFormData.active,
          }),
        }
      )
      if (response.ok) {
        setSuccessMessage('Secret updated successfully')
        setTimeout(() => setSuccessMessage(''), 5000)
        setShowEditModal(false)
        setEditingSecret(null)
        loadSecrets()
      } else {
        setErrorMessage('Failed to update registration secret')
      }
    } catch (error) {
      setErrorMessage('Error updating registration secret')
      console.error('Error updating secret:', error)
    }
  }

  const handleDeleteSecret = async (secretId: string) => {
    if (!window.confirm('Are you sure you want to delete this secret? This cannot be undone.')) {
      return
    }

    try {
      const response = await fetch(
        `${apiClient.getBaseURL()}/registration-secrets/${secretId}`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${apiClient.getToken()}`,
          },
        }
      )
      if (response.ok) {
        setSuccessMessage('Secret deleted successfully')
        setTimeout(() => setSuccessMessage(''), 5000)
        loadSecrets()
      } else {
        setErrorMessage('Failed to delete registration secret')
      }
    } catch (error) {
      setErrorMessage('Error deleting registration secret')
      console.error('Error deleting secret:', error)
    }
  }

  const copyToClipboard = (text: string, secretId: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSecretId(secretId)
    setTimeout(() => setCopiedSecretId(null), 2000)
  }

  const openEditModal = (secret: RegistrationSecret) => {
    setEditingSecret(secret)
    setEditFormData({
      name: secret.name,
      description: secret.description || '',
      active: secret.active,
    })
    setShowEditModal(true)
  }

  if (loading) {
    return <div className="text-center py-12">Loading secrets...</div>
  }

  return (
    <div className="space-y-6">
      {/* Messages */}
      {successMessage && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <p className="text-green-700">{successMessage}</p>
        </div>
      )}
      {errorMessage && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-700">{errorMessage}</p>
        </div>
      )}

      {/* New Secret Display */}
      {newSecretValue && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <div className="flex gap-2 items-start">
            <AlertCircle className="text-yellow-600 flex-shrink-0 mt-0.5" size={20} />
            <div className="flex-1">
              <h3 className="font-medium text-yellow-900">Save Your Secret</h3>
              <p className="text-sm text-yellow-700 mt-1">
                This is the only time the secret will be displayed. Save it securely now.
              </p>
              <div className="mt-3 bg-white border border-yellow-300 rounded p-3 font-mono text-sm break-all">
                {showNewSecret ? newSecretValue : '••••••••••••••••••••••••••••'}
              </div>
              <div className="mt-3 flex gap-2">
                <button
                  onClick={() => setShowNewSecret(!showNewSecret)}
                  className="flex items-center gap-2 px-3 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 text-sm"
                >
                  {showNewSecret ? (
                    <>
                      <EyeOff size={16} /> Hide
                    </>
                  ) : (
                    <>
                      <Eye size={16} /> Show
                    </>
                  )}
                </button>
                <button
                  onClick={() => copyToClipboard(newSecretValue, 'new')}
                  className="flex items-center gap-2 px-3 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 text-sm"
                >
                  {copiedSecretId === 'new' ? (
                    <>
                      <Check size={16} /> Copied
                    </>
                  ) : (
                    <>
                      <Copy size={16} /> Copy
                    </>
                  )}
                </button>
                <button
                  onClick={() => setNewSecretValue(null)}
                  className="px-3 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 text-sm ml-auto"
                >
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Registration Secrets</h2>
          <p className="text-gray-600 mt-1">Manage secrets for collector self-registration</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          <Plus size={20} /> Create Secret
        </button>
      </div>

      {/* Secrets Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {secrets.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">No registration secrets yet</p>
            <p className="text-gray-400 text-sm mt-1">Create one to enable collector self-registration</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Name</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Description</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Status</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Usage</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Created</th>
                  <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {secrets.map((secret) => (
                  <tr key={secret.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3 text-sm font-medium text-gray-900">{secret.name}</td>
                    <td className="px-6 py-3 text-sm text-gray-600">
                      {secret.description || '-'}
                    </td>
                    <td className="px-6 py-3 text-sm">
                      <span
                        className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${
                          secret.active
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-700'
                        }`}
                      >
                        {secret.active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600">
                      <div className="text-sm">
                        <p className="font-medium">{secret.total_registrations} registrations</p>
                        {secret.last_used_at && (
                          <p className="text-gray-500 text-xs mt-1">
                            Last used: {new Date(secret.last_used_at).toLocaleDateString()}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600">
                      {new Date(secret.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-3 text-sm space-x-2">
                      <button
                        onClick={() => openEditModal(secret)}
                        className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-700 rounded-md hover:bg-blue-200 text-xs"
                      >
                        <Edit2 size={14} /> Edit
                      </button>
                      <button
                        onClick={() => handleDeleteSecret(secret.id)}
                        className="inline-flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-md hover:bg-red-200 text-xs"
                      >
                        <Trash2 size={14} /> Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-md w-full mx-4 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Create Registration Secret</h3>
            <form onSubmit={handleCreateSecret} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Name *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., production, staging"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="e.g., Secret for production collectors"
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && editingSecret && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-md w-full mx-4 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Edit Registration Secret</h3>
            <form onSubmit={handleUpdateSecret} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Name *
                </label>
                <input
                  type="text"
                  value={editFormData.name}
                  onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  value={editFormData.description}
                  onChange={(e) => setEditFormData({ ...editFormData, description: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="active"
                  checked={editFormData.active}
                  onChange={(e) => setEditFormData({ ...editFormData, active: e.target.checked })}
                  className="rounded border-gray-300"
                />
                <label htmlFor="active" className="text-sm font-medium text-gray-700">
                  Active
                </label>
              </div>
              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false)
                    setEditingSecret(null)
                  }}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  Update
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
