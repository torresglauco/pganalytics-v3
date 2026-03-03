import React from 'react';
import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';

export interface BarChartDataPoint {
  name: string;
  [key: string]: string | number;
}

export interface BarDefinition {
  key: string;
  name?: string;
  fill: string;
  stackId?: string;
}

interface BarChartProps {
  data: BarChartDataPoint[];
  bars: BarDefinition[];
  height?: number;
  width?: string | number;
  layout?: 'vertical' | 'horizontal';
  stacked?: boolean;
  showGrid?: boolean;
  showLegend?: boolean;
  showTooltip?: boolean;
  xAxisLabel?: string;
  yAxisLabel?: string;
  colors?: string[];
}

export const BarChart: React.FC<BarChartProps> = ({
  data,
  bars,
  height = 300,
  width = '100%',
  layout = 'vertical',
  stacked = false,
  showGrid = true,
  showLegend = true,
  showTooltip = true,
  xAxisLabel,
  yAxisLabel,
  colors = ['#06b6d4', '#10b981', '#f59e0b', '#f43f5e'],
}) => {
  // Use provided colors or cycle through default colors
  const getBarsWithColors = () => {
    return bars.map((bar, idx) => ({
      ...bar,
      fill: bar.fill || colors[idx % colors.length],
    }));
  };

  const barsWithColors = getBarsWithColors();

  return (
    <ResponsiveContainer width={width} height={height}>
      <RechartsBarChart
        data={data}
        layout={layout}
        margin={{
          top: 5,
          right: 30,
          left: 20,
          bottom: 5,
        }}
      >
        {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />}

        {layout === 'vertical' ? (
          <>
            <XAxis type="number" />
            <YAxis dataKey="name" type="category" width={80} />
          </>
        ) : (
          <>
            <XAxis dataKey="name" />
            <YAxis />
          </>
        )}

        {showTooltip && (
          <Tooltip
            contentStyle={{
              backgroundColor: '#ffffff',
              border: '1px solid #e2e8f0',
              borderRadius: '6px',
              boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
            }}
            labelStyle={{ color: '#1e293b' }}
          />
        )}

        {showLegend && <Legend />}

        {barsWithColors.map((bar) => (
          <Bar
            key={bar.key}
            dataKey={bar.key}
            fill={bar.fill}
            name={bar.name || bar.key}
            stackId={stacked ? bar.stackId || 'stack' : undefined}
            radius={[4, 4, 0, 0]}
          />
        ))}
      </RechartsBarChart>
    </ResponsiveContainer>
  );
};
