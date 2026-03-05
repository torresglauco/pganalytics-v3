import React from 'react';

interface LoadingSkeletonProps {
  variant?: 'text' | 'card' | 'table' | 'dashboard';
  count?: number;
}

/**
 * Reusable loading skeleton component
 */
export const LoadingSkeleton: React.FC<LoadingSkeletonProps> = ({
  variant = 'text',
  count = 1,
}) => {
  const shimmerClass =
    'bg-gradient-to-r from-gray-200 to-gray-300 dark:from-gray-700 dark:to-gray-800 animate-pulse';

  if (variant === 'text') {
    return (
      <div className="space-y-2">
        {Array(count)
          .fill(null)
          .map((_, i) => (
            <div key={i} className={`h-4 rounded w-full ${shimmerClass}`} />
          ))}
      </div>
    );
  }

  if (variant === 'card') {
    return (
      <div className="space-y-4">
        {Array(count)
          .fill(null)
          .map((_, i) => (
            <div
              key={i}
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
            >
              <div className={`h-5 rounded w-1/2 mb-3 ${shimmerClass}`} />
              <div className={`h-4 rounded w-full mb-2 ${shimmerClass}`} />
              <div className={`h-4 rounded w-2/3 ${shimmerClass}`} />
            </div>
          ))}
      </div>
    );
  }

  if (variant === 'table') {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        {/* Table header */}
        <div className="grid grid-cols-5 gap-4 p-4 bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-700">
          {Array(5)
            .fill(null)
            .map((_, i) => (
              <div key={i} className={`h-4 rounded ${shimmerClass}`} />
            ))}
        </div>

        {/* Table rows */}
        {Array(count)
          .fill(null)
          .map((_, i) => (
            <div
              key={i}
              className="grid grid-cols-5 gap-4 p-4 border-b border-gray-200 dark:border-gray-700 last:border-b-0"
            >
              {Array(5)
                .fill(null)
                .map((_, j) => (
                  <div key={j} className={`h-4 rounded ${shimmerClass}`} />
                ))}
            </div>
          ))}
      </div>
    );
  }

  if (variant === 'dashboard') {
    return (
      <div className="space-y-6">
        {/* KPI Cards */}
        <div className="grid grid-cols-5 gap-4">
          {Array(5)
            .fill(null)
            .map((_, i) => (
              <div
                key={i}
                className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
              >
                <div className={`h-4 rounded w-1/2 mb-2 ${shimmerClass}`} />
                <div className={`h-8 rounded w-3/4 ${shimmerClass}`} />
              </div>
            ))}
        </div>

        {/* Charts/Metrics */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 h-64">
          <div className={`h-full rounded ${shimmerClass}`} />
        </div>

        {/* Table */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="grid grid-cols-6 gap-4 p-4 bg-gray-50 dark:bg-gray-700">
            {Array(6)
              .fill(null)
              .map((_, i) => (
                <div key={i} className={`h-4 rounded ${shimmerClass}`} />
              ))}
          </div>
          {Array(5)
            .fill(null)
            .map((_, i) => (
              <div
                key={i}
                className="grid grid-cols-6 gap-4 p-4 border-b border-gray-200 dark:border-gray-700"
              >
                {Array(6)
                  .fill(null)
                  .map((_, j) => (
                    <div key={j} className={`h-4 rounded ${shimmerClass}`} />
                  ))}
              </div>
            ))}
        </div>
      </div>
    );
  }

  return null;
};

export default LoadingSkeleton;
