import { useState } from 'react'
import { Badge } from '../ui/Badge'
import { LogDetailsModal } from './LogDetailsModal'

interface LogEntry {
  id: string
  log_timestamp: string
  log_level: string
  log_message: string
  source_location?: string
  user_name?: string
}

interface LogsTableProps {
  logs: LogEntry[]
  loading: boolean
  page: number
  pageSize: number
  onPageChange: (page: number) => void
}

const levelColorMap: Record<string, string> = {
  DEBUG: 'default',
  INFO: 'info',
  NOTICE: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  FATAL: 'error',
  PANIC: 'error',
}

export const LogsTable: React.FC<LogsTableProps> = ({
  logs,
  loading,
  page,
  onPageChange,
}) => {
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-slate-500">Loading logs...</div>
      </div>
    )
  }

  if (!logs || logs.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-slate-500">No logs found</div>
      </div>
    )
  }

  return (
    <>
      <div className="overflow-x-auto border border-slate-200 dark:border-slate-700 rounded-lg">
        <table className="w-full">
          <thead className="bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
                Timestamp
              </th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
                Level
              </th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
                Message
              </th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
                User
              </th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => (
              <tr
                key={log.id}
                className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 cursor-pointer"
                onClick={() => setSelectedLog(log)}
              >
                <td className="px-6 py-3 text-sm text-slate-600 dark:text-slate-400">
                  {new Date(log.log_timestamp).toLocaleString()}
                </td>
                <td className="px-6 py-3">
                  <Badge variant={levelColorMap[log.log_level] as any}>
                    {log.log_level}
                  </Badge>
                </td>
                <td className="px-6 py-3 text-sm text-slate-700 dark:text-slate-300 max-w-xs truncate">
                  {log.log_message}
                </td>
                <td className="px-6 py-3 text-sm text-slate-600 dark:text-slate-400">
                  {log.user_name || '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-4 flex items-center justify-between">
        <div className="text-sm text-slate-600 dark:text-slate-400">
          Page {page} • {logs.length} logs shown
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => onPageChange(page - 1)}
            disabled={page === 1}
            className="px-3 py-2 text-sm border border-slate-300 rounded-lg disabled:opacity-50 dark:border-slate-600"
          >
            Previous
          </button>
          <button
            onClick={() => onPageChange(page + 1)}
            className="px-3 py-2 text-sm border border-slate-300 rounded-lg dark:border-slate-600"
          >
            Next
          </button>
        </div>
      </div>

      {selectedLog && (
        <LogDetailsModal log={selectedLog} onClose={() => setSelectedLog(null)} />
      )}
    </>
  )
}
