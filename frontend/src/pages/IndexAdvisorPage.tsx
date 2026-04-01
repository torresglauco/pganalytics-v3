import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { MainLayout } from '../components/layout/MainLayout'
import { RecommendationCard, IndexRecommendation } from '../components/IndexAdvisor/RecommendationCard'
import { LoadingSpinner } from '../components/ui/LoadingSpinner'

export const IndexAdvisorPage: React.FC = () => {
  const { databaseId } = useParams<{ databaseId: string }>()
  const [recommendations, setRecommendations] = useState<IndexRecommendation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchRecommendations = async () => {
    if (!databaseId) {
      setError('Invalid database ID')
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const response = await fetch(`/api/v1/index-advisor/database/${databaseId}/recommendations`)

      if (!response.ok) {
        throw new Error(`Failed to fetch recommendations: ${response.statusText}`)
      }

      const data = await response.json()
      setRecommendations(data.recommendations || [])
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred'
      setError(errorMessage)
      console.error('Error fetching recommendations:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchRecommendations()
  }, [databaseId])

  const handleCreateIndex = async (id: number) => {
    try {
      const response = await fetch(
        `/api/v1/index-advisor/recommendation/${id}/create`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )

      if (!response.ok) {
        throw new Error(`Failed to create index: ${response.statusText}`)
      }

      // Update the recommendation status to 'created'
      setRecommendations((prevRecs) =>
        prevRecs.map((rec) => (rec.id === id ? { ...rec, status: 'created' as const } : rec))
      )

      // Optionally refresh the list after a short delay
      setTimeout(() => {
        fetchRecommendations()
      }, 1000)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      console.error('Error creating index:', err)
      throw new Error(errorMessage)
    }
  }

  if (!databaseId) {
    return (
      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
              Index Advisor
            </h1>
            <p className="mt-2 text-slate-600 dark:text-slate-400">
              Find missing and unused indexes to optimize query performance
            </p>
          </div>
          <div className="bg-red-50 dark:bg-red-900/20 p-6 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-red-800 dark:text-red-300">
              Invalid database ID. Please select a database first.
            </p>
          </div>
        </div>
      </MainLayout>
    )
  }

  if (loading) {
    return (
      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
              Index Advisor
            </h1>
            <p className="mt-2 text-slate-600 dark:text-slate-400">
              Find missing and unused indexes to optimize query performance
            </p>
          </div>
          <LoadingSpinner fullScreen message="Loading recommendations..." />
        </div>
      </MainLayout>
    )
  }

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            Index Advisor
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            Find missing and unused indexes to optimize query performance
          </p>
        </div>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 p-6 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-red-800 dark:text-red-300">
              Error: {error}
            </p>
            <button
              onClick={fetchRecommendations}
              className="mt-4 px-4 py-2 bg-red-600 dark:bg-red-700 text-white rounded hover:bg-red-700 dark:hover:bg-red-600 transition-colors"
            >
              Retry
            </button>
          </div>
        )}

        <div className="space-y-4">
          {recommendations.length === 0 ? (
            <div className="bg-white dark:bg-slate-800 p-8 rounded-lg text-center text-slate-600 dark:text-slate-400 shadow">
              {error ? (
                <div>No recommendations could be loaded</div>
              ) : (
                <div>
                  <div className="text-lg font-medium mb-2">No index recommendations at this time</div>
                  <p className="text-sm">Your indexes appear to be well-optimized!</p>
                </div>
              )}
            </div>
          ) : (
            recommendations.map((rec) => (
              <RecommendationCard
                key={rec.id}
                recommendation={rec}
                onCreateIndex={handleCreateIndex}
              />
            ))
          )}
        </div>
      </div>
    </MainLayout>
  )
}
