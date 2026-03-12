import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface MetricsParams {
  instance_id?: string
  time_range?: string
}

export const useMetrics = (params: MetricsParams = {}) => {
  const [metrics, setMetrics] = useState<any>(null)
  const [errorTrend, setErrorTrend] = useState<any>(null)
  const [distribution, setDistribution] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchMetrics = async () => {
    try {
      setLoading(true)
      setError(null)

      const [metricsData, trendData, distData] = await Promise.all([
        apiClient.getMetrics(params),
        apiClient.getErrorTrend({ instance_id: params.instance_id, hours: 24 }),
        apiClient.getLogDistribution(params),
      ])

      setMetrics(metricsData)
      setErrorTrend(trendData)
      setDistribution(distData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch metrics')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMetrics()
  }, [JSON.stringify(params)])

  return { metrics, errorTrend, distribution, loading, error, fetchMetrics }
}
