import React from 'react'
import { useLogAnalysis } from '../../hooks/useLogAnalysis'

export const LogStream: React.FC<{ databaseId: string }> = ({ databaseId }) => {
  const { logs, connected, error } = useLogAnalysis(databaseId)

  const severityColors = {
    INFO: 'text-blue-600 dark:text-blue-400',
    WARNING: 'text-yellow-600 dark:text-yellow-400',
    ERROR: 'text-red-600 dark:text-red-400',
    FATAL: 'text-red-900 dark:text-red-500 font-bold',
  }

  const severityBgColors = {
    INFO: 'bg-blue-50 dark:bg-blue-900/20',
    WARNING: 'bg-yellow-50 dark:bg-yellow-900/20',
    ERROR: 'bg-red-50 dark:bg-red-900/20',
    FATAL: 'bg-red-100 dark:bg-red-900/40',
  }

  const statusIndicator = connected ? (
    <span className="inline-flex items-center gap-2">
      <span className="w-2 h-2 rounded-full bg-green-500 dark:bg-green-400 animate-pulse"></span>
      <span className="text-xs font-medium text-green-700 dark:text-green-300">Connected</span>
    </span>
  ) : (
    <span className="inline-flex items-center gap-2">
      <span className="w-2 h-2 rounded-full bg-gray-500 dark:bg-gray-400"></span>
      <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Disconnected</span>
    </span>
  )

  return (
    <div className="bg-gray-900 text-gray-50 p-4 rounded-lg font-mono text-sm max-h-96 overflow-y-auto border border-gray-700 dark:border-gray-600">
      <div className="mb-4 pb-2 border-b border-gray-700 flex items-center justify-between">
        <div>{statusIndicator}</div>
        {error && (
          <span className="text-xs text-red-400">
            Error: {error}
          </span>
        )}
      </div>

      {logs.length === 0 ? (
        <div className="text-gray-400 text-center py-8">
          {connected ? 'Waiting for logs...' : 'Connect to view logs'}
        </div>
      ) : (
        <div className="space-y-1">
          {logs.map((log) => (
            <div
              key={log.id}
              className={`p-2 rounded transition-colors ${severityBgColors[log.severity]}`}
            >
              <div className="flex gap-2 items-start">
                <span className="text-gray-500 whitespace-nowrap">
                  [{new Date(log.log_timestamp).toLocaleTimeString()}]
                </span>
                <span className={`font-bold whitespace-nowrap ${severityColors[log.severity]}`}>
                  {log.severity}
                </span>
                {log.category && (
                  <span className="text-gray-400 whitespace-nowrap">
                    [{log.category}]
                  </span>
                )}
                <span className="text-gray-200 break-all flex-1">{log.message}</span>
              </div>
              {log.duration && (
                <div className="text-xs text-gray-400 mt-1 ml-2">
                  Duration: {log.duration}ms
                  {log.table_affected && ` | Table: ${log.table_affected}`}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
