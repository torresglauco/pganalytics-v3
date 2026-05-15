/**
 * Custom topology node component for PostgreSQL instances
 * Displays role badges, status indicators, and lag metrics
 */

import React from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import clsx from 'clsx';
import type { TopologyNodeData } from '../../types/replication';

/**
 * Get border color based on node role
 */
const getRoleBorderColor = (role: TopologyNodeData['role']): string => {
  switch (role) {
    case 'primary':
      return 'border-emerald-500';
    case 'standby':
      return 'border-blue-500';
    case 'cascading_standby':
      return 'border-amber-500';
    default:
      return 'border-slate-400';
  }
};

/**
 * Get role badge variant
 */
const getRoleBadgeStyle = (role: TopologyNodeData['role']): string => {
  switch (role) {
    case 'primary':
      return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200';
    case 'standby':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
    case 'cascading_standby':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200';
    default:
      return 'bg-slate-100 text-slate-800 dark:bg-slate-700 dark:text-slate-200';
  }
};

/**
 * Get status indicator color
 */
const getStatusColor = (status: TopologyNodeData['status']): string => {
  switch (status) {
    case 'streaming':
      return 'bg-emerald-500';
    case 'catchup':
      return 'bg-amber-500';
    case 'down':
      return 'bg-red-500';
    default:
      return 'bg-slate-400';
  }
};

/**
 * Format lag milliseconds to human-readable string
 */
const formatLag = (lagMs: number): string => {
  if (lagMs < 1000) {
    return `${lagMs}ms`;
  } else if (lagMs < 60000) {
    return `${(lagMs / 1000).toFixed(1)}s`;
  } else {
    return `${(lagMs / 60000).toFixed(1)}m`;
  }
};

/**
 * Custom PostgreSQL topology node component
 */
export const TopologyNode: React.FC<NodeProps<TopologyNodeData>> = ({ data }) => {
  const borderColor = getRoleBorderColor(data.role);
  const roleBadgeStyle = getRoleBadgeStyle(data.role);
  const statusColor = getStatusColor(data.status);

  return (
    <div
      className={clsx(
        'px-4 py-3 rounded-lg border-2 bg-white dark:bg-slate-800 shadow-md min-w-[180px]',
        borderColor
      )}
    >
      {/* Target handle (left side) - for upstream connections */}
      <Handle
        type="target"
        position={Position.Left}
        className="!bg-slate-400 !w-3 !h-3"
      />

      {/* Header with status indicator and label */}
      <div className="flex items-center gap-2 mb-2">
        <div
          className={clsx('w-2.5 h-2.5 rounded-full', statusColor)}
          title={`Status: ${data.status}`}
        />
        <span className="font-semibold text-slate-900 dark:text-slate-100 truncate">
          {data.label}
        </span>
      </div>

      {/* Role badge */}
      <div className="mb-2">
        <span
          className={clsx(
            'px-2 py-0.5 text-xs font-medium rounded',
            roleBadgeStyle
          )}
        >
          {data.role === 'cascading_standby' ? 'cascading' : data.role}
        </span>
      </div>

      {/* Application name and client address */}
      {data.role !== 'primary' && (
        <div className="text-xs text-slate-500 dark:text-slate-400 space-y-1">
          <div className="truncate" title={data.applicationName}>
            App: {data.applicationName || 'unknown'}
          </div>
          <div className="truncate" title={data.clientAddr}>
            IP: {data.clientAddr || 'unknown'}
          </div>
        </div>
      )}

      {/* Lag display for non-primary nodes */}
      {data.role !== 'primary' && data.lagMs > 0 && (
        <div className="mt-2 pt-2 border-t border-slate-200 dark:border-slate-700">
          <span className="text-xs font-medium text-slate-600 dark:text-slate-300">
            Lag: {formatLag(data.lagMs)}
          </span>
        </div>
      )}

      {/* Source handle (right side) - for downstream connections */}
      <Handle
        type="source"
        position={Position.Right}
        className="!bg-slate-400 !w-3 !h-3"
      />
    </div>
  );
};

TopologyNode.displayName = 'TopologyNode';

export default TopologyNode;