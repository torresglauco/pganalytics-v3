import React from 'react';
import { ArrowUp, ArrowDown, TrendingUp } from 'lucide-react';

interface MetricCardProps {
  title: string;
  value: number | string;
  unit?: string;
  icon: React.ReactNode;
  trend?: 'up' | 'down' | 'stable';
  trendValue?: string;
  status?: 'healthy' | 'warning' | 'critical';
  onClick?: () => void;
  loading?: boolean;
}

const statusColors = {
  healthy: 'bg-pg-success/10 border-pg-success/20 hover:border-pg-success/40',
  warning: 'bg-pg-warning/10 border-pg-warning/20 hover:border-pg-warning/40',
  critical: 'bg-pg-danger/10 border-pg-danger/20 hover:border-pg-danger/40',
};

const statusTextColors = {
  healthy: 'text-pg-success',
  warning: 'text-pg-warning',
  critical: 'text-pg-danger',
};

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit = '',
  icon,
  trend,
  trendValue,
  status = 'healthy',
  onClick,
  loading = false,
}) => {
  return (
    <div
      className={`
        p-6 rounded-lg border-2 transition-all
        ${statusColors[status]}
        ${onClick ? 'cursor-pointer hover:shadow-lg hover:scale-105' : ''}
      `}
      onClick={onClick}
    >
      <div className="flex justify-between items-start mb-4">
        <div className={`p-2 rounded-lg ${statusTextColors[status]}`}>
          {icon}
        </div>
        {trend && (
          <div className="flex items-center gap-1 text-sm">
            {trend === 'up' && <ArrowUp className="w-4 h-4 text-pg-success" />}
            {trend === 'down' && <ArrowDown className="w-4 h-4 text-pg-danger" />}
            {trend === 'stable' && <TrendingUp className="w-4 h-4 text-pg-slate" />}
            {trendValue && <span className="text-xs text-pg-slate">{trendValue}</span>}
          </div>
        )}
      </div>

      <h3 className="text-sm font-medium text-pg-slate mb-2">{title}</h3>

      {loading ? (
        <div className="bg-pg-slate/10 h-8 rounded animate-pulse" />
      ) : (
        <div className="flex items-baseline gap-1">
          <span className="text-3xl font-bold text-pg-dark">{value}</span>
          {unit && <span className="text-lg text-pg-slate">{unit}</span>}
        </div>
      )}
    </div>
  );
};
