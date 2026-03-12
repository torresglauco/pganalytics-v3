import { Badge } from '../ui/Badge'

interface Alert {
  id: string
  name: string
  description: string
  enabled: boolean
  conditions: any[]
  last_triggered?: string
}

interface AlertsTableProps {
  alerts: Alert[]
  loading: boolean
  onView: (alert: Alert) => void
}

export const AlertsTable: React.FC<AlertsTableProps> = ({ alerts, loading, onView }) => {
  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-slate-500">Loading alerts...</div>
      </div>
    )
  }

  if (!alerts || alerts.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-slate-500 mb-4">No alerts configured</div>
        <p className="text-sm text-slate-400">Create your first alert to get started</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto border border-slate-200 dark:border-slate-700 rounded-lg">
      <table className="w-full">
        <thead className="bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
          <tr>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Name
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Status
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Conditions
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Last Triggered
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Actions
            </th>
          </tr>
        </thead>
        <tbody>
          {alerts.map((alert) => (
            <tr
              key={alert.id}
              className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800"
            >
              <td className="px-6 py-3">
                <div>
                  <div className="text-sm font-medium text-slate-900 dark:text-white">
                    {alert.name}
                  </div>
                  <div className="text-xs text-slate-600 dark:text-slate-400 mt-1">
                    {alert.description}
                  </div>
                </div>
              </td>
              <td className="px-6 py-3">
                <Badge variant={alert.enabled ? 'success' : 'default'}>
                  {alert.enabled ? 'Active' : 'Inactive'}
                </Badge>
              </td>
              <td className="px-6 py-3 text-sm text-slate-600 dark:text-slate-400">
                {alert.conditions.length} condition(s)
              </td>
              <td className="px-6 py-3 text-sm text-slate-600 dark:text-slate-400">
                {alert.last_triggered
                  ? new Date(alert.last_triggered).toLocaleDateString()
                  : 'Never'}
              </td>
              <td className="px-6 py-3">
                <button
                  onClick={() => onView(alert)}
                  className="text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 text-sm"
                >
                  View
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
