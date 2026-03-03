import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';

interface GaugeChartProps {
  value: number;
  min?: number;
  max?: number;
  title?: string;
  unit?: string;
  size?: 'sm' | 'md' | 'lg';
  color?: string;
  thresholds?: {
    warning?: number;
    critical?: number;
  };
}

export const GaugeChart: React.FC<GaugeChartProps> = ({
  value,
  min = 0,
  max = 100,
  title,
  unit = '%',
  size = 'md',
  color = '#06b6d4',
  thresholds = { warning: 75, critical: 50 },
}) => {
  const percentage = ((value - min) / (max - min)) * 100;
  const clampedPercentage = Math.min(Math.max(percentage, 0), 100);

  // Determine color based on thresholds
  let gaugeColor = color;
  if (thresholds.critical && value <= thresholds.critical) {
    gaugeColor = '#f43f5e'; // danger
  } else if (thresholds.warning && value <= thresholds.warning) {
    gaugeColor = '#f59e0b'; // warning
  }

  const data = [
    { name: 'Used', value: clampedPercentage },
    { name: 'Remaining', value: 100 - clampedPercentage },
  ];

  const sizes = {
    sm: 120,
    md: 180,
    lg: 240,
  };

  const height = sizes[size];
  const width = height;
  const radius = height / 2 - 20;

  // Position text based on size
  const textSizeClass = size === 'sm' ? 'text-2xl' : size === 'md' ? 'text-4xl' : 'text-5xl';

  return (
    <div className="flex flex-col items-center gap-4">
      {title && <h3 className="text-sm font-medium text-pg-dark">{title}</h3>}

      <div className="relative" style={{ width, height }}>
        <ResponsiveContainer width={width} height={height}>
          <PieChart margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
            <Pie
              data={data}
              cx={width / 2}
              cy={height / 2}
              innerRadius={radius - 15}
              outerRadius={radius}
              startAngle={180}
              endAngle={0}
              dataKey="value"
            >
              <Cell fill={gaugeColor} />
              <Cell fill="#e2e8f0" />
            </Pie>
          </PieChart>
        </ResponsiveContainer>

        {/* Center Text */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <div className={`font-bold ${textSizeClass} text-pg-dark`}>
            {Math.round(value)}
          </div>
          <div className="text-xs text-pg-slate">{unit}</div>
        </div>
      </div>

      {/* Legend */}
      <div className="text-xs text-pg-slate text-center">
        <div>
          <span className="inline-block w-2 h-2 bg-current rounded-full mr-1"></span>
          {value.toFixed(1)} {unit}
        </div>
      </div>
    </div>
  );
};
