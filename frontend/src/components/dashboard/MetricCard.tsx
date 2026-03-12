import React from 'react'
import { Card, CardTitle } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface MetricCardProps {
  title: string
  value: string | number
  unit?: string
  trend?: {
    direction: 'up' | 'down' | 'neutral'
    percentage: number
    period: string
  }
  icon?: string
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit,
  trend,
  icon,
}) => {
  const getTrendColor = () => {
    if (!trend) return 'default'
    return trend.direction === 'up' ? 'error' : 'success'
  }

  const getTrendIcon = () => {
    if (!trend) return ''
    return trend.direction === 'up' ? '📈' : '📉'
  }

  return (
    <Card>
      <div className="flex items-start justify-between mb-3">
        <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
          {title}
        </CardTitle>
        {icon && <span className="text-2xl">{icon}</span>}
      </div>

      <div className="mb-3">
        <div className="text-3xl font-bold text-slate-900 dark:text-slate-100">
          {value}
          {unit && <span className="text-lg text-slate-500 ml-1">{unit}</span>}
        </div>
      </div>

      {trend && (
        <Badge variant={getTrendColor()} size="sm">
          {getTrendIcon()} {Math.abs(trend.percentage)}% in {trend.period}
        </Badge>
      )}
    </Card>
  )
}

MetricCard.displayName = 'MetricCard'
