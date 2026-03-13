import { create } from 'zustand'

type Callback = (data: any) => void

interface RealtimeStore {
  connected: boolean
  lastUpdate: string | null
  error: string | null
  subscriptions: Map<string, Set<Callback>>

  // Actions
  setConnected: (connected: boolean) => void
  setLastUpdate: (timestamp: string) => void
  setError: (error: string | null) => void
  subscribe: (event: string, callback: Callback) => void
  unsubscribe: (event: string, callback?: Callback) => void
  emit: (event: string, data: any) => void
  clear: () => void
}

export const useRealtimeStore = create<RealtimeStore>((set, get) => ({
  connected: false,
  lastUpdate: null,
  error: null,
  subscriptions: new Map(),

  setConnected: (connected) => set({ connected }),

  setLastUpdate: (timestamp) => set({ lastUpdate: timestamp }),

  setError: (error) => set({ error }),

  subscribe: (event, callback) => {
    const { subscriptions } = get()
    if (!subscriptions.has(event)) {
      subscriptions.set(event, new Set())
    }
    subscriptions.get(event)!.add(callback)
  },

  unsubscribe: (event, callback) => {
    const { subscriptions } = get()
    if (callback) {
      subscriptions.get(event)?.delete(callback)
      // Clean up empty subscription sets
      if (subscriptions.get(event)?.size === 0) {
        subscriptions.delete(event)
      }
    } else {
      subscriptions.delete(event)
    }
  },

  emit: (event, data) => {
    const { subscriptions } = get()
    subscriptions.get(event)?.forEach((callback) => {
      try {
        callback(data)
      } catch (error) {
        console.error(`Error in listener for event "${event}":`, error)
      }
    })
  },

  clear: () =>
    set({
      connected: false,
      lastUpdate: null,
      error: null,
      subscriptions: new Map(),
    }),
}))
