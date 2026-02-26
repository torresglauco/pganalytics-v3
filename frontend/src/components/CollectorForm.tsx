import React, { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { AlertCircle, CheckCircle, Loader } from 'lucide-react'
import { apiClient } from '../services/api'
import type { CollectorRegisterRequest, CollectorRegisterResponse } from '../types'

const collectorSchema = z.object({
  hostname: z.string().min(1, 'Hostname is required'),
  environment: z.string().optional(),
  group: z.string().optional(),
  description: z.string().optional(),
})

type CollectorFormData = z.infer<typeof collectorSchema>

interface CollectorFormProps {
  registrationSecret: string
  onSuccess?: (response: CollectorRegisterResponse) => void
  onError?: (error: Error) => void
}

export const CollectorForm: React.FC<CollectorFormProps> = ({
  registrationSecret,
  onSuccess,
  onError,
}) => {
  const [submitting, setSubmitting] = useState(false)
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [successResponse, setSuccessResponse] = useState<CollectorRegisterResponse | null>(null)

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
    reset,
  } = useForm<CollectorFormData>({
    resolver: zodResolver(collectorSchema),
  })

  const hostname = watch('hostname')

  const handleTestConnection = async () => {
    if (!hostname) {
      setConnectionStatus('error')
      return
    }

    setTestingConnection(true)
    setConnectionStatus('idle')

    try {
      const success = await apiClient.testConnection(hostname)
      setConnectionStatus(success ? 'success' : 'error')
    } catch (err) {
      setConnectionStatus('error')
    } finally {
      setTestingConnection(false)
    }
  }

  const onSubmit = async (data: CollectorFormData) => {
    setSubmitting(true)
    try {
      const response = await apiClient.registerCollector(
        data as CollectorRegisterRequest,
        registrationSecret
      )
      setSuccessResponse(response)
      reset()
      setConnectionStatus('idle')
      onSuccess?.(response)
    } catch (error) {
      const err = error instanceof Error ? error : new Error('Registration failed')
      onError?.(err)
    } finally {
      setSubmitting(false)
    }
  }

  if (successResponse) {
    return (
      <div className="bg-green-50 border border-green-200 rounded-lg p-6">
        <div className="flex items-center gap-3 mb-4">
          <CheckCircle className="text-green-600" size={24} />
          <h3 className="text-lg font-semibold text-green-900">Collector Registered Successfully!</h3>
        </div>
        <div className="space-y-3 mb-4">
          <div>
            <label className="text-sm font-medium text-green-700">Collector ID:</label>
            <code className="block bg-white border border-green-200 rounded p-2 mt-1 font-mono text-sm">
              {successResponse.collector_id}
            </code>
          </div>
          <div>
            <label className="text-sm font-medium text-green-700">Authentication Token:</label>
            <code className="block bg-white border border-green-200 rounded p-2 mt-1 font-mono text-sm break-all">
              {successResponse.token}
            </code>
          </div>
          <p className="text-sm text-green-700">
            Use the collector ID and token to configure your collector instance:
          </p>
          <pre className="bg-white border border-green-200 rounded p-2 text-xs overflow-auto">
{`export COLLECTOR_ID=${successResponse.collector_id}
export COLLECTOR_JWT_TOKEN=${successResponse.token}
export CENTRAL_API_URL=http://pganalytics:8080`}
          </pre>
        </div>
        <button
          onClick={() => {
            setSuccessResponse(null)
            reset()
          }}
          className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded"
        >
          Register Another Collector
        </button>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Collector Hostname/IP *
        </label>
        <div className="flex gap-2">
          <input
            {...register('hostname')}
            type="text"
            placeholder="prod-db-1.region.rds.amazonaws.com"
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="button"
            onClick={handleTestConnection}
            disabled={testingConnection || !hostname}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
          >
            {testingConnection ? <Loader size={18} className="animate-spin" /> : 'Test'}
          </button>
        </div>
        {errors.hostname && (
          <p className="text-sm text-red-600 mt-1">{errors.hostname.message}</p>
        )}
        {connectionStatus === 'success' && (
          <p className="text-sm text-green-600 mt-1 flex items-center gap-1">
            <CheckCircle size={16} /> Connection successful
          </p>
        )}
        {connectionStatus === 'error' && (
          <p className="text-sm text-red-600 mt-1 flex items-center gap-1">
            <AlertCircle size={16} /> Connection failed
          </p>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Environment
          </label>
          <select
            {...register('environment')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select environment</option>
            <option value="development">Development</option>
            <option value="staging">Staging</option>
            <option value="production">Production</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Group
          </label>
          <input
            {...register('group')}
            type="text"
            placeholder="e.g., AWS-RDS, On-Prem"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Description
        </label>
        <textarea
          {...register('description')}
          placeholder="Optional description for this collector..."
          rows={3}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <button
        type="submit"
        disabled={submitting}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 rounded-md transition"
      >
        {submitting ? 'Registering...' : 'Register Collector'}
      </button>
    </form>
  )
}
