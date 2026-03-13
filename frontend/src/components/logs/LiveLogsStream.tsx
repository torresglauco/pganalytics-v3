import React, { useState, useEffect, useRef } from 'react'
import { useRealtime } from '../../hooks/useRealtime'

interface PostgreSQLLog {
  id: number
  timestamp: string
  level: string
  message: string
  instance_id: number
}

interface DisplayLog {
  id: string
  timestamp: string
  level: string
  message: string
  rawData: PostgreSQLLog
}

interface LiveLogsStreamProps {
  instanceId: number
  onLogClick?: (log: PostgreSQLLog) => void
}

export const LiveLogsStream: React.FC<LiveLogsStreamProps> = ({
  instanceId,
  onLogClick,
}) => {
  const [logs, setLogs] = useState<DisplayLog[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const { connected, subscribe, unsubscribe } = useRealtime()

  useEffect(() => {
    const handleNewLog = (data: PostgreSQLLog) => {
      // Only add logs for the current instance
      if (data.instance_id !== instanceId) {
        return
      }

      const newLog: DisplayLog = {
        id: String(data.id),
        timestamp: data.timestamp,
        level: data.level,
        message: data.message,
        rawData: data,
      }

      setLogs((prev) => {
        // Prepend new log (newest first)
        const updated = [newLog, ...prev]
        // Keep only last 50 logs (50 newest)
        return updated.slice(0, 50)
      })

      // Auto-scroll if enabled
      if (autoScroll && scrollRef.current) {
        setTimeout(() => {
          if (scrollRef.current) {
            scrollRef.current.scrollTop = 0
          }
        }, 0)
      }
    }

    subscribe('log:new', handleNewLog)
    return () => unsubscribe('log:new', handleNewLog)
  }, [instanceId, autoScroll, subscribe, unsubscribe])

  const getLevelStyles = (level: string) => {
    switch (level) {
      case 'ERROR':
        return 'bg-red-50 dark:bg-red-950/30 border-l-4 border-red-500 text-red-900 dark:text-red-100'
      case 'SLOW_QUERY':
        return 'bg-amber-50 dark:bg-amber-950/30 border-l-4 border-amber-500 text-amber-900 dark:text-amber-100'
      default:
        return 'bg-slate-50 dark:bg-slate-900/50 border-l-4 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100'
    }
  }

  const getLevelBadgeStyles = (level: string) => {
    switch (level) {
      case 'ERROR':
        return 'bg-red-200 dark:bg-red-900 text-red-800 dark:text-red-100'
      case 'SLOW_QUERY':
        return 'bg-amber-200 dark:bg-amber-900 text-amber-800 dark:text-amber-100'
      default:
        return 'bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-100'
    }
  }

  return (
    <div className="space-y-4">
      {/* Title */}
      <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
        Live Stream
      </h2>

      {/* Header with status and controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {connected ? (
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="text-sm font-medium text-green-700 dark:text-green-400">
                LIVE
              </span>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-yellow-500 rounded-full" />
              <span className="text-sm font-medium text-yellow-700 dark:text-yellow-400">
                CONNECTING...
              </span>
            </div>
          )}
        </div>

        <button
          onClick={() => setAutoScroll(!autoScroll)}
          className={`px-3 py-1 text-sm rounded font-medium transition ${
            autoScroll
              ? 'bg-blue-500 text-white hover:bg-blue-600'
              : 'bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600'
          }`}
        >
          {autoScroll ? '⬇ Pause' : '▶ Resume'}
        </button>
      </div>

      {/* Logs container */}
      <div
        ref={scrollRef}
        className="h-96 overflow-y-auto border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-950 space-y-1 p-2"
      >
        {logs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-slate-500 dark:text-slate-400">
            <p>Waiting for logs...</p>
          </div>
        ) : (
          logs.map((log) => (
            <div
              key={log.id}
              onClick={() => onLogClick?.(log.rawData)}
              className={`p-3 rounded cursor-pointer hover:shadow-md transition ${getLevelStyles(
                log.level
              )}`}
            >
              <div className="flex items-start gap-3">
                <span
                  className={`px-2 py-1 rounded text-xs font-semibold whitespace-nowrap ${getLevelBadgeStyles(
                    log.level
                  )}`}
                >
                  {log.level}
                </span>
                <div className="flex-1 min-w-0">
                  <div className="text-xs text-slate-500 dark:text-slate-400 mb-1">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </div>
                  <div className="text-sm break-words line-clamp-2">
                    {log.message}
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Log count footer */}
      <div className="text-xs text-slate-500 dark:text-slate-400">
        {logs.length === 0
          ? 'No logs'
          : `Showing ${logs.length} logs`}
      </div>
    </div>
  )
}
