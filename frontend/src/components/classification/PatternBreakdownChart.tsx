import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import type { PatternType } from '../../types/classification';
import { PATTERN_COLORS, PATTERN_LABELS } from '../../types/classification';

interface PatternBreakdownChartProps {
  data: Record<string, number>;
  isLoading?: boolean;
  onSegmentClick?: (patternType: PatternType) => void;
}

interface ChartDataItem {
  name: string;
  value: number;
  color: string;
  patternType: PatternType;
}

export const PatternBreakdownChart: React.FC<PatternBreakdownChartProps> = ({
  data,
  isLoading = false,
  onSegmentClick,
}) => {
  const chartData: ChartDataItem[] = Object.entries(data)
    .filter(([, count]) => count > 0)
    .map(([patternType, count]) => ({
      name: PATTERN_LABELS[patternType as PatternType] || patternType,
      value: count,
      color: PATTERN_COLORS[patternType as PatternType] || '#6b7280',
      patternType: patternType as PatternType,
    }))
    .sort((a, b) => b.value - a.value);

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Pattern Breakdown</h3>
        <div className="h-64 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </div>
    );
  }

  if (chartData.length === 0) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Pattern Breakdown</h3>
        <div className="h-64 flex items-center justify-center text-gray-500">
          No pattern data available
        </div>
      </div>
    );
  }

  const total = chartData.reduce((sum, item) => sum + item.value, 0);

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const item = payload[0].payload as ChartDataItem;
      const percentage = ((item.value / total) * 100).toFixed(1);
      return (
        <div className="bg-white px-3 py-2 shadow-lg rounded-lg border border-gray-200">
          <p className="font-medium text-gray-900">{item.name}</p>
          <p className="text-sm text-gray-600">
            {item.value.toLocaleString()} matches ({percentage}%)
          </p>
        </div>
      );
    }
    return null;
  };

  const CustomLegend = ({ payload }: any) => (
    <div className="flex flex-wrap justify-center gap-3 mt-4">
      {payload.map((entry: any, index: number) => {
        const item = chartData.find((d) => d.name === entry.value);
        const percentage = item ? ((item.value / total) * 100).toFixed(1) : '0.0';
        return (
          <button
            key={`legend-${index}`}
            className={`flex items-center gap-2 px-2 py-1 rounded text-sm ${
              onSegmentClick ? 'cursor-pointer hover:bg-gray-100' : ''
            }`}
            onClick={() => item && onSegmentClick?.(item.patternType)}
          >
            <span
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-gray-700">{entry.value}</span>
            <span className="text-gray-500">({percentage}%)</span>
          </button>
        );
      })}
    </div>
  );

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">Pattern Breakdown</h3>
      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={90}
              paddingAngle={2}
              dataKey="value"
              onClick={(data) => onSegmentClick?.(data.patternType)}
              className={onSegmentClick ? 'cursor-pointer' : ''}
            >
              {chartData.map((entry, index) => (
                <Cell
                  key={`cell-${index}`}
                  fill={entry.color}
                  stroke="white"
                  strokeWidth={2}
                />
              ))}
            </Pie>
            <Tooltip content={<CustomTooltip />} />
            <Legend content={<CustomLegend />} />
          </PieChart>
        </ResponsiveContainer>
      </div>
      <div className="text-center mt-2 text-sm text-gray-500">
        Total: {total.toLocaleString()} matches detected
      </div>
    </div>
  );
};

export default PatternBreakdownChart;