import React, { useState } from 'react'
import { Tab } from '@headlessui/react'
import { AlertCircle } from 'lucide-react'
import { CollectorForm } from '../components/CollectorForm'
import { CollectorList } from '../components/CollectorList'

export const Dashboard: React.FC = () => {
  const [registrationSecret, setRegistrationSecret] = useState('')
  const [secretVisible, setSecretVisible] = useState(false)
  const [successMessage, setSuccessMessage] = useState('')

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
          <h1 className="text-3xl font-bold text-gray-900">pgAnalytics Collector Manager</h1>
          <p className="text-gray-600 mt-1">v3.3.0 - Manage PostgreSQL database collectors</p>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Registration Secret */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-8">
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
          <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-8">
            <p className="text-green-700">{successMessage}</p>
          </div>
        )}

        {/* Tabs */}
        <Tab.Group>
          <Tab.List className="flex gap-4 border-b border-gray-200 mb-6">
            <Tab
              className={({ selected }) =>
                `px-4 py-2 font-medium text-sm border-b-2 transition ${
                  selected
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-600 hover:text-gray-900'
                }`
              }
            >
              Register Collector
            </Tab>
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
          </Tab.List>

          <Tab.Panels>
            <Tab.Panel>
              <div className="bg-white rounded-lg shadow p-6">
                {!isSecretValid ? (
                  <div className="text-center py-12">
                    <AlertCircle className="mx-auto text-yellow-600 mb-3" size={32} />
                    <h3 className="text-lg font-medium text-gray-900">Secret Required</h3>
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
            </Tab.Panel>

            <Tab.Panel>
              <div className="bg-white rounded-lg shadow p-6">
                <CollectorList />
              </div>
            </Tab.Panel>
          </Tab.Panels>
        </Tab.Group>
      </main>
    </div>
  )
}
