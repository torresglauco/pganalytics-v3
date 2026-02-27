import React, { useState } from 'react'
import { Loader, Trash2, RefreshCw, AlertCircle } from 'lucide-react'
import { useCollectors } from '../hooks/useCollectors'
import type { ApiError } from '../types'

export const CollectorList: React.FC = () => {
  const { collectors, loading, error, pagination, fetchCollectors, deleteCollector } = useCollectors()
  const [deleting, setDeleting] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<ApiError | null>(null)

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this collector?')) return

    setDeleting(id)
    setDeleteError(null)
    try {
      await deleteCollector(id)
    } catch (err) {
      setDeleteError(err as ApiError)
    } finally {
      setDeleting(null)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader className="animate-spin text-blue-600" size={32} />
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-gap-3">
        <AlertCircle className="text-red-600 flex-shrink-0" size={20} />
        <div>
          <h3 className="text-sm font-medium text-red-900">Error loading collectors</h3>
          <p className="text-sm text-red-700 mt-1">{error.message}</p>
        </div>
      </div>
    )
  }

  if (collectors.length === 0) {
    return (
      <div className="text-center py-12">
        <h3 className="text-lg font-medium text-gray-900">No collectors registered</h3>
        <p className="text-gray-600 mt-1">Register your first collector to get started</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-lg font-semibold text-gray-900">
          Registered Collectors ({pagination.total})
        </h2>
        <button
          onClick={() => fetchCollectors()}
          className="flex items-center gap-2 px-3 py-2 text-sm bg-gray-200 hover:bg-gray-300 rounded"
        >
          <RefreshCw size={16} />
          Refresh
        </button>
      </div>

      {deleteError && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-700">{deleteError.message}</p>
        </div>
      )}

      <div className="overflow-x-auto border border-gray-200 rounded-lg">
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Hostname</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Status</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Created</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Last Heartbeat</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Metrics</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody>
            {collectors.map((collector) => (
              <tr key={collector.id} className="border-b border-gray-200 hover:bg-gray-50">
                <td className="px-6 py-3 text-sm font-medium text-gray-900">{collector.hostname}</td>
                <td className="px-6 py-3 text-sm">
                  <span
                    className={`px-2 py-1 rounded-full text-xs font-medium ${
                      collector.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : collector.status === 'error'
                          ? 'bg-red-100 text-red-800'
                          : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {collector.status}
                  </span>
                </td>
                <td className="px-6 py-3 text-sm text-gray-600">
                  {new Date(collector.created_at).toLocaleDateString()}
                </td>
                <td className="px-6 py-3 text-sm text-gray-600">
                  {collector.last_seen
                    ? new Date(collector.last_seen).toLocaleString()
                    : 'Never'}
                </td>
                <td className="px-6 py-3 text-sm text-gray-600">
                  {collector.metrics_count_24h ?? 0}
                </td>
                <td className="px-6 py-3 text-sm">
                  <button
                    onClick={() => handleDelete(collector.id)}
                    disabled={deleting === collector.id}
                    className="text-red-600 hover:text-red-800 disabled:text-gray-400"
                  >
                    {deleting === collector.id ? (
                      <Loader size={16} className="animate-spin" />
                    ) : (
                      <Trash2 size={16} />
                    )}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination.totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          {Array.from({ length: pagination.totalPages }, (_, i) => i + 1).map((page) => (
            <button
              key={page}
              onClick={() => fetchCollectors(page)}
              className={`px-3 py-1 rounded ${
                page === pagination.page
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 hover:bg-gray-300'
              }`}
            >
              {page}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
