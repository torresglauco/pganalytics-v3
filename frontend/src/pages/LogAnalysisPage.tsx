import React from 'react'
import { useParams } from 'react-router-dom'
import { MainLayout } from '../components/layout/MainLayout'
import { LogStream } from '../components/LogInsights/LogStream'

export const LogAnalysisPage: React.FC = () => {
  const { databaseId } = useParams<{ databaseId: string }>()

  if (!databaseId) {
    return (
      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
              Log Analysis
            </h1>
            <p className="mt-2 text-slate-600 dark:text-slate-400">
              Real-time PostgreSQL log insights
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

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
            Log Analysis
          </h1>
          <p className="mt-2 text-slate-600 dark:text-slate-400">
            Real-time PostgreSQL log insights
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6">
          <div className="bg-white dark:bg-slate-800 p-6 rounded-lg shadow">
            <h2 className="text-xl font-bold mb-4 text-slate-900 dark:text-white">
              Live Log Stream
            </h2>
            <LogStream databaseId={databaseId} />
          </div>
        </div>
      </div>
    </MainLayout>
  )
}
