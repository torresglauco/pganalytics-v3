import React, { useState, useEffect } from 'react'
import { Trash2, Lock, Unlock, Shield, User, AlertCircle, RotateCcw, Copy, CheckCircle, Plus, X } from 'lucide-react'
import { CreateUserForm } from './CreateUserForm'
import { apiClient } from '../services/api'

interface User {
  id: number
  username: string
  email: string
  full_name: string
  role: 'admin' | 'user'
  is_active: boolean
  created_at: string
}

interface UserManagementTableProps {
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

export const UserManagementTable: React.FC<UserManagementTableProps> = ({ onSuccess, onError }) => {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [togglingStatus, setTogglingStatus] = useState<number | null>(null)
  const [changingRole, setChangingRole] = useState<number | null>(null)
  const [resettingPassword, setResettingPassword] = useState<number | null>(null)
  const [tempPasswordData, setTempPasswordData] = useState<{ username: string; password: string } | null>(null)
  const [copiedPassword, setCopiedPassword] = useState(false)
  const [showCreateForm, setShowCreateForm] = useState(false)

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/users', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        throw new Error('Failed to load users')
      }

      const data = await response.json()
      setUsers(Array.isArray(data) ? data : data.users || [])
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }

  const isDefaultAdmin = (user: User): boolean => {
    return user.username === 'admin'
  }

  const canModifyUser = (user: User): boolean => {
    return !isDefaultAdmin(user)
  }

  const deleteUser = async (userId: number, username: string) => {
    if (!confirm(`Are you sure you want to delete user "${username}"?`)) {
      return
    }

    setDeleting(userId)
    try {
      const response = await fetch(`/api/v1/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to delete user')
      }

      setUsers(users.filter(u => u.id !== userId))
      onSuccess(`User "${username}" deleted successfully`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to delete user')
    } finally {
      setDeleting(null)
    }
  }

  const toggleUserStatus = async (user: User) => {
    setTogglingStatus(user.id)
    try {
      const response = await fetch(`/api/v1/users/${user.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({
          is_active: !user.is_active,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to update user status')
      }

      const updatedUser = await response.json()
      setUsers(users.map(u => (u.id === user.id ? updatedUser : u)))
      const action = updatedUser.is_active ? 'enabled' : 'disabled'
      onSuccess(`User "${user.username}" ${action} successfully`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to update user status')
    } finally {
      setTogglingStatus(null)
    }
  }

  const changeRole = async (userId: number, username: string, newRole: 'admin' | 'user') => {
    if (!confirm(`Change "${username}" role to ${newRole === 'admin' ? 'Administrator' : 'User'}?`)) {
      return
    }

    setChangingRole(userId)
    try {
      const response = await fetch(`/api/v1/users/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({
          role: newRole,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to change user role')
      }

      const updatedUser = await response.json()
      setUsers(users.map(u => (u.id === userId ? updatedUser : u)))
      onSuccess(`User "${username}" role changed to ${newRole === 'admin' ? 'Administrator' : 'User'}`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to change user role')
    } finally {
      setChangingRole(null)
    }
  }

  const resetPassword = async (userId: number, username: string) => {
    if (!confirm(`Reset password for user "${username}"? They will receive a temporary password.`)) {
      return
    }

    setResettingPassword(userId)
    try {
      const response = await fetch(`/api/v1/users/${userId}/reset-password`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Failed to reset user password')
      }

      const data = await response.json()
      setTempPasswordData({ username: data.username, password: data.temp_password })
      onSuccess(`Password reset for "${username}"`)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to reset user password')
    } finally {
      setResettingPassword(null)
    }
  }

  const copyPasswordToClipboard = () => {
    if (tempPasswordData?.password) {
      navigator.clipboard.writeText(tempPasswordData.password)
      setCopiedPassword(true)
      setTimeout(() => setCopiedPassword(false), 2000)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
          <p className="text-gray-600">Loading users...</p>
        </div>
      </div>
    )
  }

  return (
    <div>
      {/* Header with Create User Button */}
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Manage Users</h2>
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
              Create User
            </>
          )}
        </button>
      </div>

      {/* Create User Form */}
      {showCreateForm && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <CreateUserForm
            onSuccess={(message) => {
              setShowCreateForm(false)
              loadUsers()
              onSuccess(message)
            }}
            onError={onError}
          />
        </div>
      )}

      {users.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-600">No users found</p>
        </div>
      ) : null}

      <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 border border-gray-200 rounded-lg">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Username
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Email
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Full Name
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Role
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Status
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {users.map((user) => (
            <tr key={user.id} className="hover:bg-gray-50">
              <td className="px-6 py-4 whitespace-nowrap">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-900">{user.username}</span>
                  {isDefaultAdmin(user) && (
                    <Shield size={16} className="text-yellow-600" title="Default Admin" />
                  )}
                </div>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                {user.email}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                {user.full_name || '-'}
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                  user.role === 'admin'
                    ? 'bg-purple-100 text-purple-800'
                    : 'bg-blue-100 text-blue-800'
                }`}>
                  {user.role === 'admin' ? 'Administrator' : 'User'}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap">
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                  user.is_active
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                }`}>
                  {user.is_active ? 'Active' : 'Disabled'}
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                <div className="flex gap-2">
                  {/* Toggle Status */}
                  <button
                    onClick={() => toggleUserStatus(user)}
                    disabled={isDefaultAdmin(user) || togglingStatus === user.id}
                    className={`p-2 rounded transition ${
                      isDefaultAdmin(user)
                        ? 'text-gray-300 cursor-not-allowed'
                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                    }`}
                    title={isDefaultAdmin(user) ? 'Cannot disable default admin' : (user.is_active ? 'Disable user' : 'Enable user')}
                  >
                    {user.is_active ? (
                      <Lock size={18} />
                    ) : (
                      <Unlock size={18} />
                    )}
                  </button>

                  {/* Reset Password */}
                  <button
                    onClick={() => resetPassword(user.id, user.username)}
                    disabled={resettingPassword === user.id}
                    className="p-2 text-orange-600 hover:text-orange-900 hover:bg-orange-100 rounded transition disabled:opacity-50"
                    title="Reset password"
                  >
                    <RotateCcw size={18} />
                  </button>

                  {/* Change Role */}
                  <button
                    onClick={() => changeRole(user.id, user.username, user.role === 'admin' ? 'user' : 'admin')}
                    disabled={isDefaultAdmin(user) || changingRole === user.id}
                    className={`p-2 rounded transition ${
                      isDefaultAdmin(user)
                        ? 'text-gray-300 cursor-not-allowed'
                        : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                    }`}
                    title={isDefaultAdmin(user) ? 'Cannot change default admin role' : (user.role === 'admin' ? 'Make User' : 'Make Administrator')}
                  >
                    {user.role === 'admin' ? (
                      <User size={18} />
                    ) : (
                      <Shield size={18} />
                    )}
                  </button>

                  {/* Delete User */}
                  <button
                    onClick={() => deleteUser(user.id, user.username)}
                    disabled={isDefaultAdmin(user) || deleting === user.id}
                    className={`p-2 rounded transition ${
                      isDefaultAdmin(user)
                        ? 'text-gray-300 cursor-not-allowed'
                        : 'text-red-600 hover:text-red-900 hover:bg-red-100'
                    }`}
                    title={isDefaultAdmin(user) ? 'Cannot delete default admin' : 'Delete user'}
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {users.some(u => isDefaultAdmin(u)) && (
        <div className="mt-4 flex gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
          <AlertCircle size={18} className="flex-shrink-0 mt-0.5" />
          <div>
            <p className="font-medium">Default Admin Account Protection</p>
            <p>
              The default admin user (username: <strong>admin</strong>) is protected for security reasons:
            </p>
            <ul className="list-disc list-inside mt-1">
              <li>Cannot be deleted</li>
              <li>Cannot be disabled</li>
              <li>Cannot have role changed</li>
              <li>Can have password reset (only)</li>
            </ul>
          </div>
        </div>
      )}
      </div>

      {/* Temporary Password Modal */}
      {tempPasswordData && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-md w-full p-6">
            <div className="flex items-center gap-2 mb-4">
              <CheckCircle className="text-green-600" size={24} />
              <h3 className="text-lg font-semibold text-gray-900">Password Reset Successfully</h3>
            </div>

            <p className="text-gray-600 mb-4">
              Temporary password has been generated for user <strong>{tempPasswordData.username}</strong>.
              Share this password securely. The user should change it on their next login.
            </p>

            <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
              <label className="text-sm font-medium text-gray-700 block mb-2">Temporary Password</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={tempPasswordData.password}
                  readOnly
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white font-mono text-sm"
                />
                <button
                  onClick={copyPasswordToClipboard}
                  className={`px-3 py-2 rounded-md transition ${
                    copiedPassword
                      ? 'bg-green-500 text-white'
                      : 'bg-blue-600 text-white hover:bg-blue-700'
                  }`}
                >
                  {copiedPassword ? (
                    <CheckCircle size={18} />
                  ) : (
                    <Copy size={18} />
                  )}
                </button>
              </div>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mb-4 text-sm text-yellow-800">
              <strong>Important:</strong> This password is shown only once. Store it securely before closing this dialog.
            </div>

            <button
              onClick={() => setTempPasswordData(null)}
              className="w-full px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
