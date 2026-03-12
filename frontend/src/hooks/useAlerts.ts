import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

interface AlertsParams {
  page?: number
  page_size?: number
  status?: string
}

export const useAlerts = (params: AlertsParams = {}) => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchAlerts = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getAlerts(params)
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch alerts')
    } finally {
      setLoading(false)
    }
  }

  const createAlert = async (alertData: any) => {
    try {
      const result = await apiClient.createAlert(alertData)
      await fetchAlerts()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create alert'
      setError(errorMsg)
      throw err
    }
  }

  const updateAlert = async (alertId: string, alertData: any) => {
    try {
      const result = await apiClient.updateAlert(alertId, alertData)
      await fetchAlerts()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update alert'
      setError(errorMsg)
      throw err
    }
  }

  const deleteAlert = async (alertId: string) => {
    try {
      await apiClient.deleteAlert(alertId)
      await fetchAlerts()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete alert'
      setError(errorMsg)
      throw err
    }
  }

  useEffect(() => {
    fetchAlerts()
  }, [JSON.stringify(params)])

  return { data, loading, error, fetchAlerts, createAlert, updateAlert, deleteAlert }
}
