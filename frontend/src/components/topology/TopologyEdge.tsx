/**
 * Custom topology edge component for replication links
 * Displays lag metrics with color coding
 */

import React from 'react';
import {
  BaseEdge,
  getBezierPath,
  type EdgeProps,
} from '@xyflow/react';
import type { TopologyEdgeData } from '../../types/replication';

/**
 * Get edge color based on lag severity
 */
const getLagColor = (lagMs: number): string => {
  if (lagMs < 1000) {
    // Low lag (< 1s) - green
    return '#10b981';
  } else if (lagMs < 10000) {
    // Medium lag (1-10s) - amber/yellow
    return '#f59e0b';
  } else {
    // High lag (> 10s) - red
    return '#ef4444';
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
 * Custom topology edge component for replication links
 */
export const TopologyEdge: React.FC<EdgeProps<TopologyEdgeData>> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  style = {},
}) => {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  });

  const lagMs = data?.lagMs ?? 0;
  const edgeColor = getLagColor(lagMs);

  return (
    <>
      {/* The edge path */}
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          ...style,
          stroke: edgeColor,
          strokeWidth: 2,
        }}
      />

      {/* Lag label */}
      {lagMs > 0 && (
        <g transform={`translate(${labelX}, ${labelY})`}>
          <rect
            x="-25"
            y="-10"
            width="50"
            height="20"
            rx="4"
            fill="white"
            stroke={edgeColor}
            strokeWidth="1"
            className="dark:fill-slate-800"
          />
          <text
            textAnchor="middle"
            dominantBaseline="middle"
            className="text-xs font-medium fill-slate-600 dark:fill-slate-300"
            style={{ fontSize: '10px' }}
          >
            {formatLag(lagMs)}
          </text>
        </g>
      )}
    </>
  );
};

TopologyEdge.displayName = 'TopologyEdge';

export default TopologyEdge;