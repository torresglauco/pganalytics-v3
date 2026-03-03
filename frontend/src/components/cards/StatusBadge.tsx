import React from 'react';
import { AlertCircle, CheckCircle, AlertTriangle, Info } from 'lucide-react';

type BadgeStatus = 'success' | 'error' | 'warning' | 'info';

interface StatusBadgeProps {
  status: BadgeStatus;
  label: string;
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
}

const statusConfig = {
  success: {
    bg: 'bg-pg-success/10',
    text: 'text-pg-success',
    border: 'border-pg-success/20',
    icon: CheckCircle,
  },
  error: {
    bg: 'bg-pg-danger/10',
    text: 'text-pg-danger',
    border: 'border-pg-danger/20',
    icon: AlertCircle,
  },
  warning: {
    bg: 'bg-pg-warning/10',
    text: 'text-pg-warning',
    border: 'border-pg-warning/20',
    icon: AlertTriangle,
  },
  info: {
    bg: 'bg-pg-cyan/10',
    text: 'text-pg-cyan',
    border: 'border-pg-cyan/20',
    icon: Info,
  },
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  label,
  size = 'md',
  showIcon = true,
}) => {
  const config = statusConfig[status];
  const Icon = config.icon;

  const sizeClasses = {
    sm: 'px-2 py-1 text-xs gap-1',
    md: 'px-3 py-1.5 text-sm gap-1.5',
    lg: 'px-4 py-2 text-base gap-2',
  };

  return (
    <div
      className={`
        inline-flex items-center rounded-full border font-semibold
        ${config.bg} ${config.text} ${config.border}
        ${sizeClasses[size]}
      `}
    >
      {showIcon && <Icon className="w-4 h-4" />}
      {label}
    </div>
  );
};
