import React from 'react'
import { Card, CardHeader, CardTitle } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface Collector {
  id: string
  hostname: string
  environment: string
  status: 'OK' | 'SLOW' | 'DOWN'
  last_heartbeat: string
  error_count_24h: number
}

interface CollectorStatusTableProps {
  collectors: Collector[]
  isLoading?: boolean
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'OK':
      return '✓'
    case 'SLOW':
      return '⚠'
    case 'DOWN':
      return '✗'
    default:
      return '•'
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'OK':
      return 'success'
    case 'SLOW':
      return 'warning'
    case 'DOWN':
      return 'error'
    default:
      return 'default'
  }
}

export const CollectorStatusTable: React.FC<CollectorStatusTableProps> = ({
  collectors,
}) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Collector Status (Real-time)</CardTitle>
      </CardHeader>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200 dark:border-slate-700">
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Hostname
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Environment
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Status
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Last Seen
              </th>
              <th className="text-right py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Errors (24h)
              </th>
            </tr>
          </thead>
          <tbody>
            {collectors.map((collector) => (
              <tr
                key={collector.id}
                className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors"
              >
                <td className="py-3 px-4 text-slate-900 dark:text-slate-100 font-medium">
                  {collector.hostname}
                </td>
                <td className="py-3 px-4 text-slate-600 dark:text-slate-400">
                  {collector.environment}
                </td>
                <td className="py-3 px-4">
                  <Badge variant={getStatusColor(collector.status)} size="sm">
                    {getStatusIcon(collector.status)} {collector.status}
                  </Badge>
                </td>
                <td className="py-3 px-4 text-slate-600 dark:text-slate-400">
                  {collector.last_heartbeat}
                </td>
                <td className="py-3 px-4 text-right text-slate-900 dark:text-slate-100">
                  {collector.error_count_24h}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}

CollectorStatusTable.displayName = 'CollectorStatusTable'
