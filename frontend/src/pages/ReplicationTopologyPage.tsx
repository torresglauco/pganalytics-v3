/**
 * Replication Topology Page
 * Displays interactive replication topology visualization for a collector
 */

import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { RefreshCw, AlertCircle, Network } from 'lucide-react';
import type { ReplicationTopology } from '../types/replication';
import { getTopology } from '../api/replicationApi';
import { TopologyGraph, TopologyLegend } from '../components/topology';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';

export const ReplicationTopologyPage: React.FC = () => {
  const { collectorId } = useParams<{ collectorId: string }>();
  const [topology, setTopology] = useState<ReplicationTopology | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  /**
   * Load topology data
   */
  const loadTopology = useCallback(async () => {
    if (!collectorId) {
      setError('No collector ID provided');
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      const data = await getTopology(collectorId);
      setTopology(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load topology');
    } finally {
      setIsLoading(false);
    }
  }, [collectorId]);

  /**
   * Load topology on mount and when collectorId changes
   */
  useEffect(() => {
    loadTopology();
  }, [loadTopology]);

  /**
   * Loading state
   */
  if (isLoading) {
    return (
      <div className="h-[calc(100vh-200px)] flex items-center justify-center">
        <LoadingSpinner size="lg" message="Loading replication topology..." />
      </div>
    );
  }

  /**
   * Error state
   */
  if (error) {
    return (
      <div className="p-6">
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-300 flex gap-2">
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div>
            <p className="font-medium">Failed to load topology</p>
            <p className="text-sm mt-1">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  /**
   * Empty state - no topology data
   */
  if (!topology) {
    return (
      <div className="p-6">
        <div className="text-center py-12">
          <Network size={48} className="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
          <h3 className="text-lg font-medium text-slate-900 dark:text-slate-100">
            No Topology Data
          </h3>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-2">
            No replication topology found for this collector.
          </p>
          <p className="text-sm text-slate-500 dark:text-slate-400">
            This may indicate the instance is not configured for streaming replication.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-100px)] flex flex-col">
      {/* Header */}
      <div className="flex justify-between items-start px-6 py-4 border-b border-slate-200 dark:border-slate-700">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            Replication Topology
          </h1>
          <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
            Visualize cascading replication: primary {">"} standby {">"} standby
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={loadTopology}
            className="flex items-center gap-2 px-4 py-2 border border-slate-300 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 font-medium transition-colors"
            disabled={isLoading}
          >
            <RefreshCw size={18} className={isLoading ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex">
        {/* Graph container */}
        <div className="flex-1 min-h-[400px]">
          <TopologyGraph topology={topology} />
        </div>

        {/* Legend sidebar */}
        <div className="w-72 p-4 border-l border-slate-200 dark:border-slate-700 overflow-y-auto">
          <TopologyLegend />
        </div>
      </div>
    </div>
  );
};

export default ReplicationTopologyPage;