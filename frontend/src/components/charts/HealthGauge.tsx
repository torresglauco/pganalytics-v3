import React from 'react';
import { ArrowUp, ArrowDown } from 'lucide-react';
import { getHealthColor, getHealthStatus } from '../../utils/calculations';

interface HealthGaugeProps {
  score: number;
  title?: string;
  size?: 'sm' | 'md' | 'lg';
  showTrend?: boolean;
  trendValue?: number;
  showDetails?: boolean;
}

export const HealthGauge: React.FC<HealthGaugeProps> = ({
  score,
  title = 'Health Score',
  size = 'md',
  showTrend = false,
  trendValue,
  showDetails = true,
}) => {
  const percentage = Math.min(Math.max(score, 0), 100);
  const status = getHealthStatus(score);
  const statusColor = getHealthColor(score);

  const sizeMap = {
    sm: { radius: 45, width: 120, height: 120, fontSize: 'text-xl' },
    md: { radius: 60, width: 180, height: 180, fontSize: 'text-3xl' },
    lg: { radius: 75, width: 220, height: 220, fontSize: 'text-4xl' },
  };

  const s = sizeMap[size];
  const circumference = 2 * Math.PI * s.radius;
  const offset = circumference - (percentage / 100) * circumference;

  return (
    <div className="flex flex-col items-center">
      {title && <h3 className="text-lg font-semibold text-pg-dark mb-4">{title}</h3>}

      <div className="relative" style={{ width: s.width, height: s.height }}>
        <svg
          width={s.width}
          height={s.height}
          className="transform -rotate-90"
        >
          {/* Background circle */}
          <circle
            cx={s.width / 2}
            cy={s.height / 2}
            r={s.radius}
            fill="none"
            stroke="#e2e8f0"
            strokeWidth="8"
          />

          {/* Progress circle */}
          <circle
            cx={s.width / 2}
            cy={s.height / 2}
            r={s.radius}
            fill="none"
            stroke={statusColor.replace('text-', '#').substring(0, 7) || '#06b6d4'}
            strokeWidth="8"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            style={{
              transition: 'stroke-dashoffset 0.5s ease-in-out',
            }}
          />
        </svg>

        {/* Center content */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <div className={`font-bold ${s.fontSize} ${statusColor}`}>
            {Math.round(percentage)}
          </div>
          {showDetails && (
            <div className="text-xs text-pg-slate mt-1">/ 100</div>
          )}
        </div>
      </div>

      {/* Status label */}
      <div className="mt-4 text-center">
        <p className={`text-sm font-semibold ${statusColor}`}>
          {status.charAt(0).toUpperCase() + status.slice(1)}
        </p>
      </div>

      {/* Trend */}
      {showTrend && trendValue !== undefined && (
        <div className="mt-2 flex items-center gap-1 text-sm">
          {trendValue > 0 ? (
            <ArrowUp className="w-4 h-4 text-pg-success" />
          ) : trendValue < 0 ? (
            <ArrowDown className="w-4 h-4 text-pg-danger" />
          ) : null}
          <span className={trendValue > 0 ? 'text-pg-success' : trendValue < 0 ? 'text-pg-danger' : 'text-pg-slate'}>
            {trendValue > 0 ? '+' : ''}{trendValue}%
          </span>
        </div>
      )}
    </div>
  );
};
