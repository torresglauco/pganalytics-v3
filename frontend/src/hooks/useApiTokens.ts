import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

export interface ApiToken {
  id: string
  name: string
  token: string
  created_at: string | Date
  last_used: string | Date | null
  expires_at: string | Date | null
}

export const useApiTokens = () => {
  const [data, setData] = useState<ApiToken[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchTokens = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.listApiTokens()
      setData(Array.isArray(result) ? result : [])
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to fetch tokens'
      setError(errorMsg)
      console.error('Error fetching tokens:', err)
    } finally {
      setLoading(false)
    }
  }

  const createToken = async (tokenData: { name: string; expires_at?: string }) => {
    try {
      setError(null)
      const result = await apiClient.createApiToken(tokenData)
      await fetchTokens()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create token'
      setError(errorMsg)
      throw err
    }
  }

  const deleteToken = async (tokenId: string) => {
    try {
      setError(null)
      await apiClient.deleteApiToken(tokenId)
      await fetchTokens()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete token'
      setError(errorMsg)
      throw err
    }
  }

  const updateToken = async (tokenId: string, tokenData: any) => {
    try {
      setError(null)
      const result = await apiClient.updateApiToken(tokenId, tokenData)
      await fetchTokens()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update token'
      setError(errorMsg)
      throw err
    }
  }

  useEffect(() => {
    fetchTokens()
  }, [])

  return {
    data,
    loading,
    error,
    fetchTokens,
    createToken,
    deleteToken,
    updateToken
  }
}
