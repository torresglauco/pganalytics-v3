import { useCallback, useMemo } from 'react'
import { useRealtimeStore } from '../stores/realtimeStore'

type EventCallback = (data: any) => void

export interface UseRealtimeReturn {
  connected: boolean
  lastUpdate: string | null
  error: string | null
  subscribe: (event: string, callback: EventCallback) => void
  unsubscribe: (event: string, callback?: EventCallback) => void
}

export const useRealtime = (): UseRealtimeReturn => {
  const connected = useRealtimeStore((state) => state.connected)
  const lastUpdate = useRealtimeStore((state) => state.lastUpdate)
  const error = useRealtimeStore((state) => state.error)
  const { subscribe, unsubscribe } = useRealtimeStore()

  const memoizedSubscribe = useCallback(
    (event: string, callback: EventCallback) => {
      subscribe(event, callback)
    },
    [subscribe]
  )

  const memoizedUnsubscribe = useCallback(
    (event: string, callback?: EventCallback) => {
      unsubscribe(event, callback)
    },
    [unsubscribe]
  )

  return useMemo(
    () => ({
      connected,
      lastUpdate,
      error,
      subscribe: memoizedSubscribe,
      unsubscribe: memoizedUnsubscribe,
    }),
    [connected, lastUpdate, error, memoizedSubscribe, memoizedUnsubscribe]
  )
}
