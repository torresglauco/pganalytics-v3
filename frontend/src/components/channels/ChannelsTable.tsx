import { Badge } from '../ui/Badge'

interface Channel {
  id: string
  name: string
  type: string
  enabled: boolean
}

interface ChannelsTableProps {
  channels: Channel[]
  loading: boolean
  onView: (channel: Channel) => void
}

export const ChannelsTable: React.FC<ChannelsTableProps> = ({
  channels,
  loading,
  onView,
}) => {
  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-slate-500">Loading channels...</div>
      </div>
    )
  }

  if (!channels || channels.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-slate-500 mb-4">No notification channels configured</div>
        <p className="text-sm text-slate-400">Create your first channel to enable alert notifications</p>
      </div>
    )
  }

  const typeColors: Record<string, any> = {
    email: 'info',
    slack: 'success',
    pagerduty: 'warning',
    webhook: 'default',
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
              Type
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Status
            </th>
            <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900 dark:text-white">
              Actions
            </th>
          </tr>
        </thead>
        <tbody>
          {channels.map((channel) => (
            <tr
              key={channel.id}
              className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800"
            >
              <td className="px-6 py-3 text-sm font-medium text-slate-900 dark:text-white">
                {channel.name}
              </td>
              <td className="px-6 py-3">
                <Badge variant={typeColors[channel.type]}>
                  {channel.type.toUpperCase()}
                </Badge>
              </td>
              <td className="px-6 py-3">
                <Badge variant={channel.enabled ? 'success' : 'default'}>
                  {channel.enabled ? 'Active' : 'Inactive'}
                </Badge>
              </td>
              <td className="px-6 py-3">
                <button
                  onClick={() => onView(channel)}
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
