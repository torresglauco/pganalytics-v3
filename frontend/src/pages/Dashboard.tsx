import React, { useState } from 'react'
import { Tab } from '@headlessui/react'
import { AlertCircle } from 'lucide-react'
import { CollectorForm } from '../components/CollectorForm'
import { CollectorList } from '../components/CollectorList'
import { UserMenuDropdown } from '../components/UserMenuDropdown'
import { ManagedInstancesTable } from '../components/ManagedInstancesTable'
import { RegistrationSecretsManager } from '../components/RegistrationSecretsManager'
import { apiClient } from '../services/api'

interface DashboardProps {
  onLogout: () => void
}

export const Dashboard: React.FC<DashboardProps> = ({ onLogout }) => {
  const [registrationSecret, setRegistrationSecret] = useState('')
  const [secretVisible, setSecretVisible] = useState(false)
  const [successMessage, setSuccessMessage] = useState('')
  const [managedInstanceMessage, setManagedInstanceMessage] = useState('')
  const [managedInstanceMessageType, setManagedInstanceMessageType] = useState<'success' | 'error' | ''>('')
  const currentUser = apiClient.getCurrentUser()
  const isAdmin = currentUser?.role === 'admin'

  const handleSecretChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setRegistrationSecret(e.target.value)
  }

  const handleSuccess = () => {
    setSuccessMessage('Collector registered successfully!')
    setTimeout(() => setSuccessMessage(''), 5000)
  }

  const isSecretValid = registrationSecret.trim().length > 0

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">pgAnalytics Collector Manager</h1>
              <p className="text-gray-600 mt-1">v3.3.0 - Manage PostgreSQL database collectors</p>
            </div>
            <UserMenuDropdown onLogout={onLogout} />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

        {/* Tabs */}
        <Tab.Group>
          <Tab.List className="flex gap-4 border-b border-gray-200 mb-6 flex-wrap">
            <Tab
              className={({ selected }) =>
                `px-4 py-2 font-medium text-sm border-b-2 transition ${
                  selected
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-600 hover:text-gray-900'
                }`
              }
            >
              Manage Collectors
            </Tab>
            {isAdmin && (
              <Tab
                className={({ selected }) =>
                  `px-4 py-2 font-medium text-sm border-b-2 transition ${
                    selected
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-600 hover:text-gray-900'
                  }`
                }
              >
                Managed Instances
              </Tab>
            )}
            {isAdmin && (
              <Tab
                className={({ selected }) =>
                  `px-4 py-2 font-medium text-sm border-b-2 transition ${
                    selected
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-600 hover:text-gray-900'
                  }`
                }
              >
                Registration Secrets
              </Tab>
            )}
          </Tab.List>

          <Tab.Panels>
            <Tab.Panel>
              <div className="space-y-6">
                {/* Registration Secret */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex gap-2 items-start">
                    <AlertCircle className="text-blue-600 flex-shrink-0 mt-0.5" size={20} />
                    <div className="flex-1">
                      <h3 className="font-medium text-blue-900">Registration Secret Required</h3>
                      <p className="text-sm text-blue-700 mt-1">
                        Enter the registration secret from your environment configuration to register new collectors
                      </p>
                      <div className="mt-3 flex gap-2">
                        <div className="flex-1">
                          <input
                            type={secretVisible ? 'text' : 'password'}
                            value={registrationSecret}
                            onChange={handleSecretChange}
                            placeholder="Enter registration secret..."
                            className="w-full px-3 py-2 border border-blue-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          />
                        </div>
                        <button
                          onClick={() => setSecretVisible(!secretVisible)}
                          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                        >
                          {secretVisible ? 'Hide' : 'Show'}
                        </button>
                      </div>
                      {!isSecretValid && (
                        <p className="text-sm text-blue-700 mt-2">
                          ℹ️ Secret is required to register collectors for security
                        </p>
                      )}
                    </div>
                  </div>
                </div>

                {successMessage && (
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <p className="text-green-700">{successMessage}</p>
                  </div>
                )}

                {/* Register Collector Form */}
                <div className="bg-white rounded-lg shadow p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Register Collector</h3>
                  {!isSecretValid ? (
                    <div className="text-center py-12">
                      <AlertCircle className="mx-auto text-yellow-600 mb-3" size={32} />
                      <h4 className="text-lg font-medium text-gray-900">Secret Required</h4>
                      <p className="text-gray-600 mt-2">
                        Please enter the registration secret above to register collectors
                      </p>
                    </div>
                  ) : (
                    <CollectorForm
                      registrationSecret={registrationSecret}
                      onSuccess={handleSuccess}
                      onError={(error) =>
                        alert(`Error: ${error.message}`)
                      }
                    />
                  )}
                </div>

                {/* Manage Collectors */}
                <div className="bg-white rounded-lg shadow p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Active Collectors</h3>
                  <CollectorList />
                </div>
              </div>
            </Tab.Panel>

            {isAdmin && (
              <Tab.Panel>
                <div className="bg-white rounded-lg shadow p-6">
                  {managedInstanceMessage && (
                    <div className={`mb-4 border rounded-lg p-4 ${
                      managedInstanceMessageType === 'success'
                        ? 'bg-green-50 border-green-200'
                        : 'bg-red-50 border-red-200'
                    }`}>
                      <p className={managedInstanceMessageType === 'success' ? 'text-green-700' : 'text-red-700'}>
                        {managedInstanceMessage}
                      </p>
                    </div>
                  )}
                  <ManagedInstancesTable
                    onSuccess={(message) => {
                      setManagedInstanceMessage(message)
                      setManagedInstanceMessageType('success')
                      setTimeout(() => setManagedInstanceMessage(''), 10000)
                    }}
                    onError={(message) => {
                      setManagedInstanceMessage(message)
                      setManagedInstanceMessageType('error')
                      setTimeout(() => setManagedInstanceMessage(''), 10000)
                    }}
                  />
                </div>
              </Tab.Panel>
            )}
            {isAdmin && (
              <Tab.Panel>
                <RegistrationSecretsManager />
              </Tab.Panel>
            )}
          </Tab.Panels>
        </Tab.Group>
      </main>
    </div>
  )
}
