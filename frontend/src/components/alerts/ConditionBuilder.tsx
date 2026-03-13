import React from 'react'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import type { AlertCondition, MetricType, ComparisonOperator } from '../../types/alerts'
import { ConditionPreview } from './ConditionPreview'

interface ConditionBuilderProps {
  conditions: AlertCondition[]
  onAddCondition: (condition: AlertCondition) => void
  onRemoveCondition: (index: number) => void
  onUpdateCondition?: (index: number, condition: AlertCondition) => void
}

const METRICS: MetricType[] = [
  'error_count',
  'slow_query_count',
  'connection_count',
  'cache_hit_ratio',
]

const OPERATORS: ComparisonOperator[] = ['>', '<', '==', '!=', '>=', '<=']

const getMetricLabel = (metric: MetricType): string => {
  const labels: Record<MetricType, string> = {
    error_count: 'Error Count',
    slow_query_count: 'Slow Query Count',
    connection_count: 'Connection Count',
    cache_hit_ratio: 'Cache Hit Ratio',
  }
  return labels[metric]
}

const ConditionBlock: React.FC<{
  condition: AlertCondition
  index: number
  onRemove: (index: number) => void
  onUpdate?: (index: number, condition: AlertCondition) => void
}> = ({ condition, index, onRemove, onUpdate }) => {
  const handleChange = (field: keyof AlertCondition, value: any) => {
    if (onUpdate) {
      onUpdate(index, {
        ...condition,
        [field]: value,
      })
    }
  }

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
        {/* Metric Type */}
        <div>
          <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
            Metric
          </label>
          <select
            value={condition.metricType}
            onChange={(e) => handleChange('metricType', e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white text-sm"
          >
            {METRICS.map((metric) => (
              <option key={metric} value={metric}>
                {getMetricLabel(metric as MetricType)}
              </option>
            ))}
          </select>
        </div>

        {/* Operator */}
        <div>
          <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
            Operator
          </label>
          <select
            value={condition.operator}
            onChange={(e) => handleChange('operator', e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white text-sm"
          >
            {OPERATORS.map((op) => (
              <option key={op} value={op}>
                {op}
              </option>
            ))}
          </select>
        </div>

        {/* Threshold */}
        <div>
          <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
            Threshold
          </label>
          <Input
            type="number"
            value={condition.threshold}
            onChange={(e) => handleChange('threshold', parseFloat(e.target.value))}
            placeholder="e.g., 5"
            className="text-sm"
          />
        </div>

        {/* Time Window */}
        <div>
          <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
            Time Window (min)
          </label>
          <Input
            type="number"
            value={condition.timeWindow}
            onChange={(e) => handleChange('timeWindow', parseInt(e.target.value, 10))}
            placeholder="e.g., 10"
            className="text-sm"
          />
        </div>

        {/* Duration */}
        <div>
          <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
            Duration (min)
          </label>
          <Input
            type="number"
            value={condition.duration || ''}
            onChange={(e) => handleChange('duration', e.target.value ? parseInt(e.target.value, 10) : undefined)}
            placeholder="Optional"
            className="text-sm"
          />
        </div>
      </div>

      {/* Remove Button */}
      <div className="flex justify-end">
        <button
          onClick={() => onRemove(index)}
          className="text-sm px-3 py-1 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
        >
          Remove
        </button>
      </div>

      {/* Preview */}
      <ConditionPreview condition={condition} />
    </div>
  )
}

export const ConditionBuilder: React.FC<ConditionBuilderProps> = ({
  conditions,
  onAddCondition,
  onRemoveCondition,
  onUpdateCondition,
}) => {
  const addNewCondition = () => {
    const newCondition: AlertCondition = {
      id: `condition-${Date.now()}`,
      metricType: 'error_count',
      operator: '>',
      threshold: 5,
      timeWindow: 10,
    }
    onAddCondition(newCondition)
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
          Alert Conditions
        </label>
        <span className="text-xs text-slate-500 dark:text-slate-400">
          {conditions.length} condition{conditions.length !== 1 ? 's' : ''}
        </span>
      </div>

      {conditions.length === 0 ? (
        <div className="p-4 border border-dashed border-slate-300 dark:border-slate-600 rounded-lg bg-slate-50 dark:bg-slate-800">
          <p className="text-sm text-slate-500 dark:text-slate-400 text-center">
            No conditions added yet
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {conditions.map((condition, index) => (
            <div key={condition.id || index} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4">
              {index > 0 && (
                <div className="mb-3 text-sm font-medium text-slate-600 dark:text-slate-400">
                  AND
                </div>
              )}
              <ConditionBlock
                condition={condition}
                index={index}
                onRemove={onRemoveCondition}
                onUpdate={onUpdateCondition}
              />
            </div>
          ))}
        </div>
      )}

      <Button
        variant="secondary"
        size="sm"
        onClick={addNewCondition}
      >
        + Add Condition
      </Button>
    </div>
  )
}
