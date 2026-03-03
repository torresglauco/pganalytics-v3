import React from 'react';
import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

export interface ChartDataPoint {
  name: string;
  value: number;
  [key: string]: string | number;
}

interface LineChartProps {
  data: ChartDataPoint[];
  title?: string;
  height?: number;
  lines?: Array<{
    key: string;
    stroke?: string;
    name?: string;
  }>;
  tooltip?: (payload: any) => React.ReactNode;
}

export const LineChart: React.FC<LineChartProps> = ({
  data,
  title,
  height = 300,
  lines = [{ key: 'value', stroke: '#06b6d4', name: 'Value' }],
  tooltip,
}) => {
  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center bg-pg-slate/5 rounded-lg border border-pg-slate/10 p-6" style={{ height }}>
        <p className="text-pg-slate">No data available</p>
      </div>
    );
  }

  return (
    <div className="w-full">
      {title && <h3 className="text-lg font-semibold text-pg-dark mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <RechartsLineChart
          data={data}
          margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
          <XAxis dataKey="name" stroke="#64748b" style={{ fontSize: '12px' }} />
          <YAxis stroke="#64748b" style={{ fontSize: '12px' }} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#ffffff',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
            }}
            formatter={tooltip}
          />
          <Legend />
          {lines.map((line) => (
            <Line
              key={line.key}
              type="monotone"
              dataKey={line.key}
              stroke={line.stroke || '#06b6d4'}
              name={line.name || line.key}
              dot={false}
              strokeWidth={2}
              isAnimationActive={true}
            />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  );
};
