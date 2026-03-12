import { useState } from 'react'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'

interface LogFiltersProps {
  onFiltersChange: (filters: FilterState) => void
}

export interface FilterState {
  level: string | null
  fromDate: string
  toDate: string
}

export const LogFilters: React.FC<LogFiltersProps> = ({ onFiltersChange }) => {
  const [filters, setFilters] = useState<FilterState>({
    level: null,
    fromDate: '',
    toDate: '',
  })

  const handleLevelChange = (level: string | null) => {
    const newFilters = { ...filters, level }
    setFilters(newFilters)
    onFiltersChange(newFilters)
  }

  const handleDateChange = (key: 'fromDate' | 'toDate', value: string) => {
    const newFilters = { ...filters, [key]: value }
    setFilters(newFilters)
    onFiltersChange(newFilters)
  }

  const logLevels = ['DEBUG', 'INFO', 'NOTICE', 'WARNING', 'ERROR', 'FATAL', 'PANIC']

  return (
    <div className="space-y-4 rounded-lg border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900">
      <div>
        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
          Log Level
        </label>
        <select
          value={filters.level || ''}
          onChange={(e) => handleLevelChange(e.target.value || null)}
          className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white"
        >
          <option value="">All Levels</option>
          {logLevels.map((level) => (
            <option key={level} value={level}>
              {level}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
          From Date
        </label>
        <Input
          type="datetime-local"
          value={filters.fromDate}
          onChange={(e) => handleDateChange('fromDate', e.target.value)}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
          To Date
        </label>
        <Input
          type="datetime-local"
          value={filters.toDate}
          onChange={(e) => handleDateChange('toDate', e.target.value)}
        />
      </div>

      <Button
        variant="secondary"
        size="sm"
        fullWidth
        onClick={() => {
          setFilters({ level: null, fromDate: '', toDate: '' })
          onFiltersChange({ level: null, fromDate: '', toDate: '' })
        }}
      >
        Reset Filters
      </Button>
    </div>
  )
}
