import React from 'react';
import { ArrowUp, ArrowDown, Minus } from 'lucide-react';
import { LineChart, Line, ResponsiveContainer } from 'recharts';
import type { HostMetrics } from '../../types/host';

interface HostMetricsCardProps {
  title: string;
  value: number | string;
  unit?: string;
  trend?: 'up' | 'down' | 'stable';
  trendValue?: number;
  data?: HostMetrics[];
  dataKey?: keyof HostMetrics;
  color?: string;
  warningThreshold?: number;
  criticalThreshold?: number;
  inverseThresholds?: boolean;
}

export const HostMetricsCard: React.FC<HostMetricsCardProps> = ({
  title,
  value,
  unit,
  trend,
  trendValue,
  data,
  dataKey,
  color = '#06b6d4',
  warningThreshold,
  criticalThreshold,
  inverseThresholds = false,
}) => {
  const getValueColor = () => {
    if (typeof value !== 'number') return 'text-gray-900';
    if (warningThreshold === undefined || criticalThreshold === undefined) {
      return 'text-gray-900';
    }

    if (inverseThresholds) {
      // Lower is worse (e.g., idle CPU)
      if (value <= criticalThreshold) return 'text-red-600';
      if (value <= warningThreshold) return 'text-amber-600';
      return 'text-emerald-600';
    } else {
      // Higher is worse (e.g., used percent)
      if (value >= criticalThreshold) return 'text-red-600';
      if (value >= warningThreshold) return 'text-amber-600';
      return 'text-emerald-600';
    }
  };

  const getTrendIcon = () => {
    if (!trend) return null;
    switch (trend) {
      case 'up':
        return <ArrowUp size={14} className="text-red-500" />;
      case 'down':
        return <ArrowDown size={14} className="text-emerald-500" />;
      default:
        return <Minus size={14} className="text-gray-400" />;
    }
  };

  const chartData = data && dataKey
    ? data.map((m, index) => ({
        name: index.toString(),
        value: m[dataKey] as number,
      }))
    : null;

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-2">
        <h3 className="text-sm font-medium text-gray-600">{title}</h3>
        {trendValue !== undefined && (
          <div className="flex items-center gap-1 text-xs">
            {getTrendIcon()}
            <span className="text-gray-500">{trendValue}%</span>
          </div>
        )}
      </div>

      <div className="flex items-end justify-between">
        <div>
          <span className={`text-2xl font-bold ${getValueColor()}`}>
            {typeof value === 'number' ? value.toFixed(1) : value}
          </span>
          {unit && <span className="text-sm text-gray-500 ml-1">{unit}</span>}
        </div>

        {chartData && chartData.length > 0 && (
          <div className="w-20 h-10">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData}>
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke={color}
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      {warningThreshold !== undefined && criticalThreshold !== undefined && (
        <div className="mt-2 flex gap-2 text-xs text-gray-500">
          <span>Warning: {warningThreshold}%</span>
          <span>Critical: {criticalThreshold}%</span>
        </div>
      )}
    </div>
  );
};

export default HostMetricsCard;