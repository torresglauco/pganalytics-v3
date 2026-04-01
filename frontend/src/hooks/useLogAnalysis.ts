import { useState, useEffect } from 'react'
import { LogEntry } from '../types/logAnalysis'

export const useLogAnalysis = (databaseId: string) => {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!databaseId) {
      setConnected(false)
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/logs/stream/${databaseId}`

    let ws: WebSocket

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        setConnected(true)
        setError(null)
      }

      ws.onmessage = (event) => {
        try {
          const newLog = JSON.parse(event.data) as LogEntry
          setLogs((prev) => [newLog, ...prev].slice(0, 100))
        } catch (err) {
          console.error('Failed to parse log message:', err)
        }
      }

      ws.onerror = (event) => {
        console.error('WebSocket error:', event)
        setError('WebSocket connection error')
        setConnected(false)
      }

      ws.onclose = () => {
        setConnected(false)
      }

      return () => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.close()
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create WebSocket')
      setConnected(false)
    }
  }, [databaseId])

  return { logs, connected, error }
}
