import { useState } from 'react'
import { useMetrics } from '../../hooks/useMetrics'
import { MetricsControls } from './MetricsControls'
import { ErrorTrendChart } from './ErrorTrendChart'
import { LogDistributionChart } from './LogDistributionChart'
import { TopErrorCodesChart } from './TopErrorCodesChart'

export const MetricsViewer: React.FC = () => {
  const [timeRange, setTimeRange] = useState('24h')

  const { metrics, errorTrend, distribution, loading, error, fetchMetrics } = useMetrics({
    time_range: timeRange,
  })

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
        <div className="text-red-800 dark:text-red-200">Error: {error}</div>
        <button
          onClick={() => fetchMetrics()}
          className="mt-2 px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <MetricsControls timeRange={timeRange} onTimeRangeChange={setTimeRange} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ErrorTrendChart data={errorTrend || []} loading={loading} />
        <LogDistributionChart data={distribution || []} loading={loading} />
      </div>

      <div className="grid grid-cols-1 gap-6">
        <TopErrorCodesChart data={metrics?.topErrors || []} loading={loading} />
      </div>
    </div>
  )
}
