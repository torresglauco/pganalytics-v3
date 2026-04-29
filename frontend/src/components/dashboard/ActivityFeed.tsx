import React from 'react'
import { Card, CardHeader, CardTitle } from '../ui/Card'

interface Activity {
  id: string
  type: 'error' | 'warning' | 'info' | 'success'
  title: string
  description: string
  timestamp: string
}

interface ActivityFeedProps {
  activities: Activity[]
}

const getTypeIcon = (type: string) => {
  switch (type) {
    case 'error':
      return '🔴'
    case 'warning':
      return '🟡'
    case 'success':
      return '🟢'
    default:
      return '🔵'
  }
}

export const ActivityFeed: React.FC<ActivityFeedProps> = ({ activities }) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>

      <div className="space-y-3">
        {activities.length === 0 ? (
          <p className="text-sm text-slate-600 dark:text-slate-400">
            No recent activity
          </p>
        ) : (
          activities.map((activity) => (
            <div key={activity.id} className="flex gap-3 py-2">
              <span className="text-xl">{getTypeIcon(activity.type)}</span>
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-2">
                  <h4 className="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {activity.title}
                  </h4>
                  <span className="text-xs text-slate-500 dark:text-slate-400 whitespace-nowrap">
                    {activity.timestamp}
                  </span>
                </div>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  {activity.description}
                </p>
              </div>
            </div>
          ))
        )}
      </div>

      {activities.length > 0 && (
        <a
          href="/logs"
          className="inline-block mt-4 text-sm text-primary-600 hover:text-primary-700 font-medium"
        >
          View all logs →
        </a>
      )}
    </Card>
  )
}

ActivityFeed.displayName = 'ActivityFeed'
