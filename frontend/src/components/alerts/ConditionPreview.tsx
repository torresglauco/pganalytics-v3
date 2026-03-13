import React from 'react'
import type { AlertCondition } from '../../types/alerts'

interface ConditionPreviewProps {
  condition: AlertCondition
}

const getMetricLabel = (metricType: string): string => {
  const labels: Record<string, string> = {
    error_count: 'Error Count',
    slow_query_count: 'Slow Query Count',
    connection_count: 'Connection Count',
    cache_hit_ratio: 'Cache Hit Ratio',
  }
  return labels[metricType] || metricType
}

const formatThreshold = (threshold: number, metricType: string): string => {
  if (metricType === 'cache_hit_ratio') {
    return `${(threshold * 100).toFixed(1)}%`
  }
  return threshold.toString()
}

export const ConditionPreview: React.FC<ConditionPreviewProps> = ({ condition }) => {
  const metricLabel = getMetricLabel(condition.metricType)
  const thresholdValue = formatThreshold(condition.threshold, condition.metricType)
  const durationText = condition.duration
    ? `for ${condition.duration} minute${condition.duration > 1 ? 's' : ''}`
    : ''

  return (
    <div className="px-3 py-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-200 dark:border-slate-700">
      <p className="text-sm text-slate-700 dark:text-slate-300">
        <span className="font-medium">{metricLabel}</span>
        {' '}
        <span className="text-slate-500 dark:text-slate-400">{condition.operator}</span>
        {' '}
        <span className="font-medium">{thresholdValue}</span>
        {' '}
        <span className="text-slate-500 dark:text-slate-400">
          in last {condition.timeWindow} minute{condition.timeWindow > 1 ? 's' : ''}
        </span>
        {durationText && (
          <>
            {' '}
            <span className="text-slate-500 dark:text-slate-400">{durationText}</span>
          </>
        )}
      </p>
    </div>
  )
}
