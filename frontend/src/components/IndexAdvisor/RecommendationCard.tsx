import React, { useState } from 'react'
import { AlertCircle, CheckCircle, Loader } from 'lucide-react'

export interface IndexRecommendation {
  id: number
  table_name: string
  column_names: string[]
  estimated_benefit: number
  weighted_cost_improvement: number
  status: 'recommended' | 'created' | 'rejected'
}

interface RecommendationCardProps {
  recommendation: IndexRecommendation
  onCreateIndex: (id: number) => Promise<void>
  onError?: (error: string) => void
}

export const RecommendationCard: React.FC<RecommendationCardProps> = ({
  recommendation,
  onCreateIndex,
  onError,
}) => {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleCreate = async () => {
    setLoading(true)
    setError(null)
    try {
      await onCreateIndex(recommendation.id)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      setError(errorMessage)
      onError?.(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const isCreated = recommendation.status === 'created'

  return (
    <div className="bg-white dark:bg-slate-800 border-l-4 border-blue-500 dark:border-blue-400 rounded-lg shadow p-4 hover:shadow-lg transition-shadow">
      <div className="flex justify-between items-start gap-4">
        <div className="flex-1">
          <h3 className="font-bold text-lg text-slate-900 dark:text-white">
            {recommendation.table_name}
          </h3>
          <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
            Columns: {recommendation.column_names.join(', ')}
          </p>

          <div className="mt-3 grid grid-cols-2 gap-4">
            <div>
              <span className="text-gray-600 dark:text-slate-400 text-sm block mb-1">
                Estimated Benefit
              </span>
              <div className="text-lg font-bold text-green-600 dark:text-green-400">
                {recommendation.estimated_benefit.toFixed(1)}%
              </div>
            </div>
            <div>
              <span className="text-gray-600 dark:text-slate-400 text-sm block mb-1">
                Cost Improvement
              </span>
              <div className="text-lg font-bold text-blue-600 dark:text-blue-400">
                {recommendation.weighted_cost_improvement.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>

        <div className="flex flex-col items-end gap-2">
          <button
            onClick={handleCreate}
            disabled={loading || isCreated}
            className={`px-4 py-2 rounded font-medium transition-colors flex items-center gap-2 whitespace-nowrap ${
              isCreated
                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 cursor-default'
                : loading
                  ? 'bg-blue-600 dark:bg-blue-700 text-white cursor-wait'
                  : 'bg-blue-600 dark:bg-blue-700 text-white hover:bg-blue-700 dark:hover:bg-blue-600'
            }`}
          >
            {loading ? (
              <>
                <Loader className="w-4 h-4 animate-spin" />
                Creating...
              </>
            ) : isCreated ? (
              <>
                <CheckCircle className="w-4 h-4" />
                Created
              </>
            ) : (
              'Create Index'
            )}
          </button>

          {error && (
            <div className="flex items-center gap-1 text-red-600 dark:text-red-400 text-xs bg-red-50 dark:bg-red-900/20 px-2 py-1 rounded max-w-xs">
              <AlertCircle className="w-3 h-3 flex-shrink-0" />
              <span className="truncate">{error}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
