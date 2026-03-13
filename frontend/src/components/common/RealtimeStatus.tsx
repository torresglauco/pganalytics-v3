import React from 'react'
import { useRealtime } from '../../hooks/useRealtime'

interface RealtimeStatusProps {
  showTimestamp?: boolean
}

export const RealtimeStatus: React.FC<RealtimeStatusProps> = ({
  showTimestamp = false
}) => {
  const { connected, lastUpdate } = useRealtime()

  return (
    <div className="flex items-center gap-2">
      {connected ? (
        <>
          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
          <span className="text-xs font-semibold text-green-600 dark:text-green-400">
            Live
          </span>
        </>
      ) : (
        <>
          <div className="w-2 h-2 bg-yellow-500 rounded-full" />
          <span className="text-xs font-semibold text-yellow-600 dark:text-yellow-400">
            Polling
          </span>
        </>
      )}
      {showTimestamp && lastUpdate && (
        <span className="text-xs text-slate-500 dark:text-slate-400 ml-2">
          {new Date(lastUpdate).toLocaleTimeString()}
        </span>
      )}
    </div>
  )
}
