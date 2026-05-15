/**
 * Topology Legend Component
 * Displays legend for role colors, status indicators, and lag thresholds
 */

import React from 'react';
import clsx from 'clsx';

interface LegendItemProps {
  color: string;
  label: string;
  description?: string;
}

const LegendItem: React.FC<LegendItemProps> = ({ color, label, description }) => (
  <div className="flex items-center gap-2">
    <div className={clsx('w-3 h-3 rounded', color)} />
    <div>
      <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
        {label}
      </span>
      {description && (
        <span className="text-xs text-slate-500 dark:text-slate-400 ml-1">
          {description}
        </span>
      )}
    </div>
  </div>
);

interface LegendSectionProps {
  title: string;
  children: React.ReactNode;
}

const LegendSection: React.FC<LegendSectionProps> = ({ title, children }) => (
  <div className="space-y-2">
    <h4 className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
      {title}
    </h4>
    <div className="space-y-1.5">{children}</div>
  </div>
);

/**
 * Topology legend component showing role colors, status indicators, and lag thresholds
 */
export const TopologyLegend: React.FC = () => {
  return (
    <div className="bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 p-4 space-y-4">
      <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
        Legend
      </h3>

      {/* Node Roles */}
      <LegendSection title="Node Roles">
        <LegendItem
          color="bg-emerald-500"
          label="Primary"
          description="Master/leader node"
        />
        <LegendItem
          color="bg-blue-500"
          label="Standby"
          description="Streaming replica"
        />
        <LegendItem
          color="bg-amber-500"
          label="Cascading"
          description="Connected to standby"
        />
      </LegendSection>

      {/* Status Indicators */}
      <LegendSection title="Status">
        <LegendItem
          color="bg-emerald-500"
          label="Streaming"
          description="Actively receiving WAL"
        />
        <LegendItem
          color="bg-amber-500"
          label="Catchup"
          description="Falling behind"
        />
        <LegendItem
          color="bg-red-500"
          label="Down"
          description="Not connected"
        />
      </LegendSection>

      {/* Lag Thresholds */}
      <LegendSection title="Replication Lag">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-emerald-500 rounded" />
            <span className="text-sm text-slate-600 dark:text-slate-400">
              Low (&lt; 1s)
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-amber-500 rounded" />
            <span className="text-sm text-slate-600 dark:text-slate-400">
              Medium (1-10s)
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-red-500 rounded" />
            <span className="text-sm text-slate-600 dark:text-slate-400">
              High (&gt; 10s)
            </span>
          </div>
        </div>
      </LegendSection>

      {/* Sync State */}
      <LegendSection title="Sync State">
        <LegendItem
          color="bg-blue-500"
          label="Sync"
          description="Synchronous replica"
        />
        <LegendItem
          color="bg-slate-400"
          label="Async"
          description="Asynchronous replica"
        />
        <LegendItem
          color="bg-purple-500"
          label="Potential"
          description="Potential sync candidate"
        />
      </LegendSection>

      {/* Tips */}
      <div className="pt-3 border-t border-slate-200 dark:border-slate-700">
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Tip: Zoom and pan to explore large topologies. Click nodes for details.
        </p>
      </div>
    </div>
  );
};

TopologyLegend.displayName = 'TopologyLegend';

export default TopologyLegend;