import React, { useState, useRef, useEffect } from 'react'
import { Settings, LogOut, User, Key } from 'lucide-react'
import { ChangePasswordForm } from './ChangePasswordForm'
import { apiClient } from '../services/api'

interface UserMenuDropdownProps {
  onLogout: () => void
}

export const UserMenuDropdown: React.FC<UserMenuDropdownProps> = ({ onLogout }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [successMessage, setSuccessMessage] = useState('')
  const menuRef = useRef<HTMLDivElement>(null)

  const currentUser = apiClient.getCurrentUser()

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleLogout = () => {
    if (confirm('Are you sure you want to log out?')) {
      onLogout()
    }
  }

  return (
    <div className="relative" ref={menuRef}>
      {/* User Info Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-3 hover:bg-gray-100 px-3 py-2 rounded-lg transition"
      >
        {currentUser ? (
          <div className="text-right">
            <p className="text-sm font-medium text-gray-900">{currentUser.full_name || currentUser.username || 'User'}</p>
            <p className="text-xs text-gray-500">{currentUser.email || ''}</p>
          </div>
        ) : null}
        <Settings size={20} className="text-gray-600" />
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-50">
          {/* User Info Header */}
          <div className="px-4 py-3 border-b border-gray-200">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                <User size={20} className="text-blue-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {currentUser?.full_name || currentUser?.username || 'User'}
                </p>
                <p className="text-xs text-gray-500 truncate">{currentUser?.email || ''}</p>
              </div>
            </div>
          </div>

          {/* Menu Options */}
          <div className="py-2">
            <button
              onClick={() => {
                setShowPasswordModal(true)
                setIsOpen(false)
              }}
              className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition"
            >
              <Key size={16} className="text-gray-500" />
              Change Password
            </button>

            <button
              onClick={handleLogout}
              className="w-full px-4 py-2 text-left text-sm text-red-700 hover:bg-red-50 flex items-center gap-2 transition"
            >
              <LogOut size={16} className="text-red-500" />
              Logout
            </button>
          </div>
        </div>
      )}

      {/* Change Password Modal */}
      {showPasswordModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-lg max-w-md w-full p-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
              <Key size={24} className="text-blue-600" />
              Change Password
            </h2>

            {successMessage && (
              <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
                <p className="text-green-700 text-sm">{successMessage}</p>
              </div>
            )}

            <ChangePasswordForm
              onSuccess={(message) => {
                setSuccessMessage(message)
                setTimeout(() => {
                  setShowPasswordModal(false)
                  setSuccessMessage('')
                }, 2000)
              }}
              onError={(message) => {
                alert(`Error: ${message}`)
              }}
            />

            <button
              onClick={() => setShowPasswordModal(false)}
              className="mt-4 w-full px-4 py-2 bg-gray-200 text-gray-900 font-medium rounded-md hover:bg-gray-300 transition"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
