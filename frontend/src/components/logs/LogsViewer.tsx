import { useState } from 'react'
import { useLogs } from '../../hooks/useLogs'
import { SearchBar } from './SearchBar'
import { LogFilters, FilterState } from './LogFilters'
import { LogsTable } from './LogsTable'
import { LiveLogsStream } from './LiveLogsStream'
import { RealtimeStatus } from '../common/RealtimeStatus'

export const LogsViewer: React.FC = () => {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [filters, setFilters] = useState<FilterState>({
    level: null,
    fromDate: '',
    toDate: '',
    instanceId: null,
  })

  const { data, loading, error, fetchLogs } = useLogs({
    page,
    page_size: 25,
    level: filters.level || undefined,
    search: search || undefined,
    instance_id: filters.instanceId ? String(filters.instanceId) : undefined,
    from_time: filters.fromDate || undefined,
    to_time: filters.toDate || undefined,
  })

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
        <div className="text-red-800 dark:text-red-200">Error: {error}</div>
        <button
          onClick={() => fetchLogs()}
          className="mt-2 px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4">
        <div className="lg:col-span-3">
          <SearchBar onSearch={setSearch} />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1">
          <LogFilters onFiltersChange={setFilters} />
        </div>

        <div className="lg:col-span-3">
          {/* Live Stream Section - only show when instance is selected */}
          {filters.instanceId && (
            <div className="mb-6 border-b pb-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
                  Live Stream
                </h2>
                <RealtimeStatus showTimestamp={true} />
              </div>
              <LiveLogsStream instanceId={filters.instanceId} />
            </div>
          )}

          {/* Historical Logs Section */}
          <div>
            <h2 className="text-lg font-semibold mb-4 text-slate-900 dark:text-white">
              Historical Logs
            </h2>
            <LogsTable
              logs={data?.logs || []}
              loading={loading}
              page={page}
              pageSize={25}
              onPageChange={setPage}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
