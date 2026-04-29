import React, { useState } from 'react';
import type {
  RuleCondition,
  ConditionType,
  ThresholdCondition,
  AnomalyCondition,
  ChangeCondition,
} from '../types/alertRules';

interface RuleConditionBuilderProps {
  condition: RuleCondition | null;
  onChange: (condition: RuleCondition) => void;
  databaseId: string;
}

export const RuleConditionBuilder: React.FC<RuleConditionBuilderProps> = ({
  condition,
  onChange,
  databaseId: _databaseId,
}) => {
  const [conditionType, setConditionType] = useState<ConditionType>(
    condition?.type || 'threshold'
  );

  /**
   * Get available metrics (mock - would come from backend)
   */
  const getMetrics = () => [
    'cpu_percent',
    'memory_percent',
    'connections',
    'queries_per_second',
    'cache_hit_ratio',
    'disk_usage_percent',
    'replication_lag_seconds',
    'transaction_rate',
  ];

  /**
   * Handle threshold condition
   */
  const handleThresholdChange = (field: string, value: any) => {
    const threshold = (condition as ThresholdCondition) || {
      type: 'threshold' as const,
      metric_name: '',
      operator: '>' as const,
      threshold_value: 0,
      window_seconds: 300,
    };

    const updated: ThresholdCondition = {
      ...threshold,
      [field]: value,
    };

    onChange(updated);
  };

  /**
   * Handle anomaly condition
   */
  const handleAnomalyChange = (field: string, value: any) => {
    const anomaly = (condition as AnomalyCondition) || {
      type: 'anomaly' as const,
      metric_name: '',
      sensitivity: 'medium' as const,
      baseline_days: 7,
    };

    const updated: AnomalyCondition = {
      ...anomaly,
      [field]: value,
    };

    onChange(updated);
  };

  /**
   * Handle change condition
   */
  const handleChangeChange = (field: string, value: any) => {
    const changeCondition = (condition as ChangeCondition) || {
      type: 'change' as const,
      metric_name: '',
      change_type: 'increase' as const,
      change_percent: 50,
      compare_to: 'previous' as const,
    };

    const updated: ChangeCondition = {
      ...changeCondition,
      [field]: value,
    };

    onChange(updated);
  };

  const metrics = getMetrics();

  return (
    <div className="space-y-6">
      {/* Condition Type Selector */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Condition Type *
        </label>
        <div className="grid grid-cols-2 gap-3">
          {[
            { value: 'threshold', label: 'Threshold', desc: 'Trigger when metric exceeds value' },
            { value: 'anomaly', label: 'Anomaly', desc: 'Trigger when metric deviates from baseline' },
            { value: 'change', label: 'Change', desc: 'Trigger when metric changes significantly' },
            { value: 'composite', label: 'Composite', desc: 'Combine multiple conditions' },
          ].map((type) => (
            <button
              key={type.value}
              onClick={() => {
                setConditionType(type.value as ConditionType);
                // Reset condition when type changes
                if (type.value === 'threshold') {
                  onChange({
                    type: 'threshold',
                    metric_name: '',
                    operator: '>',
                    threshold_value: 0,
                  });
                } else if (type.value === 'anomaly') {
                  onChange({
                    type: 'anomaly',
                    metric_name: '',
                    sensitivity: 'medium',
                  });
                } else if (type.value === 'change') {
                  onChange({
                    type: 'change',
                    metric_name: '',
                    change_type: 'increase',
                    change_percent: 50,
                    compare_to: 'previous',
                  });
                }
              }}
              className={`p-3 border-2 rounded-lg text-left transition ${
                conditionType === type.value
                  ? 'border-blue-600 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <div className="font-medium text-gray-900">{type.label}</div>
              <div className="text-xs text-gray-600">{type.desc}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Threshold Condition */}
      {conditionType === 'threshold' && (
        <div className="space-y-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h4 className="font-medium text-gray-900">Threshold Condition</h4>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Metric *
            </label>
            <select
              value={(condition as ThresholdCondition)?.metric_name || ''}
              onChange={(e) => handleThresholdChange('metric_name', e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Select metric</option>
              {metrics.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Operator *
              </label>
              <select
                value={(condition as ThresholdCondition)?.operator || '>'}
                onChange={(e) => handleThresholdChange('operator', e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value=">">Greater than (&gt;)</option>
                <option value="<">Less than (&lt;)</option>
                <option value=">=">&gt;= Greater or equal</option>
                <option value="<=">&lt;= Less or equal</option>
                <option value="=">Equals (=)</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Value *
              </label>
              <input
                type="number"
                value={(condition as ThresholdCondition)?.threshold_value || 0}
                onChange={(e) =>
                  handleThresholdChange('threshold_value', parseFloat(e.target.value))
                }
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Aggregation
              </label>
              <select
                value={(condition as ThresholdCondition)?.aggregation || 'avg'}
                onChange={(e) => handleThresholdChange('aggregation', e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="avg">Average</option>
                <option value="max">Maximum</option>
                <option value="min">Minimum</option>
                <option value="sum">Sum</option>
                <option value="count">Count</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Evaluation Window (seconds)
            </label>
            <input
              type="number"
              value={(condition as ThresholdCondition)?.window_seconds || 300}
              onChange={(e) =>
                handleThresholdChange('window_seconds', parseInt(e.target.value))
              }
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <p className="text-xs text-gray-600 mt-1">
              How long the condition must be true before triggering
            </p>
          </div>
        </div>
      )}

      {/* Anomaly Condition */}
      {conditionType === 'anomaly' && (
        <div className="space-y-4 p-4 bg-green-50 border border-green-200 rounded-lg">
          <h4 className="font-medium text-gray-900">Anomaly Condition</h4>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Metric *
            </label>
            <select
              value={(condition as AnomalyCondition)?.metric_name || ''}
              onChange={(e) => handleAnomalyChange('metric_name', e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Select metric</option>
              {metrics.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Sensitivity *
            </label>
            <select
              value={(condition as AnomalyCondition)?.sensitivity || 'medium'}
              onChange={(e) => handleAnomalyChange('sensitivity', e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="low">Low (±2.5 sigma) - Fewer alerts</option>
              <option value="medium">Medium (±2 sigma) - Balanced</option>
              <option value="high">High (±1.5 sigma) - More alerts</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Baseline Period (days)
            </label>
            <input
              type="number"
              value={(condition as AnomalyCondition)?.baseline_days || 7}
              onChange={(e) =>
                handleAnomalyChange('baseline_days', parseInt(e.target.value))
              }
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>
      )}

      {/* Change Condition */}
      {conditionType === 'change' && (
        <div className="space-y-4 p-4 bg-orange-50 border border-orange-200 rounded-lg">
          <h4 className="font-medium text-gray-900">Change Condition</h4>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Metric *
            </label>
            <select
              value={(condition as ChangeCondition)?.metric_name || ''}
              onChange={(e) => handleChangeChange('metric_name', e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Select metric</option>
              {metrics.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Change Direction *
              </label>
              <select
                value={(condition as ChangeCondition)?.change_type || 'increase'}
                onChange={(e) => handleChangeChange('change_type', e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="increase">Increase</option>
                <option value="decrease">Decrease</option>
                <option value="both">Either</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Minimum Change (%)
              </label>
              <input
                type="number"
                value={(condition as ChangeCondition)?.change_percent || 50}
                onChange={(e) =>
                  handleChangeChange('change_percent', parseFloat(e.target.value))
                }
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Compare To *
            </label>
            <select
              value={(condition as ChangeCondition)?.compare_to || 'previous'}
              onChange={(e) => handleChangeChange('compare_to', e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="previous">Previous reading</option>
              <option value="1h_ago">1 hour ago</option>
              <option value="1d_ago">1 day ago</option>
            </select>
          </div>
        </div>
      )}

      {/* Info Box */}
      <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg text-sm text-blue-700">
        <strong>Tip:</strong> You can test your condition before saving the rule to ensure it triggers correctly.
      </div>
    </div>
  );
};

export default RuleConditionBuilder;
