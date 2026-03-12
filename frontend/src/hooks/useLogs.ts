import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface LogsParams {
  page?: number
  page_size?: number
  level?: string
  search?: string
  instance_id?: string
  from_time?: string
  to_time?: string
}

export const useLogs = (params: LogsParams = {}) => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchLogs = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getLogs(params)
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
  }, [JSON.stringify(params)])

  const getLogDetails = async (logId: string) => {
    try {
      const result = await apiClient.getLogDetails(logId)
      return result
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch log details')
      return null
    }
  }

  return { data, loading, error, fetchLogs, getLogDetails }
}
