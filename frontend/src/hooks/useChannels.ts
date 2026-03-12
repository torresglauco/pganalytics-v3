import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'

export const useChannels = () => {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchChannels = async () => {
    try {
      setLoading(true)
      setError(null)
      const result = await apiClient.getChannels()
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch channels')
    } finally {
      setLoading(false)
    }
  }

  const createChannel = async (channelData: any) => {
    try {
      const result = await apiClient.createChannel(channelData)
      await fetchChannels()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create channel'
      setError(errorMsg)
      throw err
    }
  }

  const updateChannel = async (channelId: string, channelData: any) => {
    try {
      const result = await apiClient.updateChannel(channelId, channelData)
      await fetchChannels()
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update channel'
      setError(errorMsg)
      throw err
    }
  }

  const deleteChannel = async (channelId: string) => {
    try {
      await apiClient.deleteChannel(channelId)
      await fetchChannels()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete channel'
      setError(errorMsg)
      throw err
    }
  }

  const testChannel = async (channelId: string) => {
    try {
      const result = await apiClient.testChannel(channelId)
      return result
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to test channel'
      setError(errorMsg)
      throw err
    }
  }

  useEffect(() => {
    fetchChannels()
  }, [])

  return { data, loading, error, fetchChannels, createChannel, updateChannel, deleteChannel, testChannel }
}
