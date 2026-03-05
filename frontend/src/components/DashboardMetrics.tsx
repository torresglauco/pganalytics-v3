import React from 'react';
import { TrendingUp, TrendingDown, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import type { AlertStats } from '../types/alertDashboard';

interface DashboardMetricsProps {
  stats: AlertStats;
}

export const DashboardMetrics: React.FC<DashboardMetricsProps> = ({ stats }) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
      {/* Total Alerts */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-gray-600">Total Alerts</h3>
          <AlertCircle size={20} className="text-gray-400" />
        </div>
        <p className="text-3xl font-bold text-gray-900">{stats.total_alerts}</p>
        {stats.alert_rate_per_hour && (
          <p className="text-xs text-gray-600 mt-1">
            {stats.alert_rate_per_hour.toFixed(1)}/hour
          </p>
        )}
      </div>

      {/* Firing Alerts */}
      <div className="bg-white rounded-lg border border-red-200 bg-red-50 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-red-600">Firing</h3>
          <AlertCircle size={20} className="text-red-500" />
        </div>
        <p className="text-3xl font-bold text-red-600">{stats.firing}</p>
        <div className="flex gap-2 mt-2 text-xs text-red-600">
          <span>
            {stats.by_severity.critical} critical
          </span>
          <span>
            {stats.by_severity.high} high
          </span>
        </div>
      </div>

      {/* Acknowledged */}
      <div className="bg-white rounded-lg border border-yellow-200 bg-yellow-50 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-yellow-600">Acknowledged</h3>
          <Clock size={20} className="text-yellow-500" />
        </div>
        <p className="text-3xl font-bold text-yellow-600">{stats.acknowledged}</p>
        <p className="text-xs text-yellow-600 mt-1">Awaiting resolution</p>
      </div>

      {/* Resolved */}
      <div className="bg-white rounded-lg border border-green-200 bg-green-50 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-green-600">Resolved</h3>
          <CheckCircle size={20} className="text-green-500" />
        </div>
        <p className="text-3xl font-bold text-green-600">{stats.resolved}</p>
        <p className="text-xs text-green-600 mt-1">Today</p>
      </div>

      {/* MTTR */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-gray-600">Avg MTTR</h3>
          <TrendingDown size={20} className="text-gray-400" />
        </div>
        <p className="text-3xl font-bold text-gray-900">
          {stats.avg_time_to_resolve_minutes
            ? `${stats.avg_time_to_resolve_minutes.toFixed(0)}m`
            : '-'}
        </p>
        <p className="text-xs text-gray-600 mt-1">Mean time to resolve</p>
      </div>

      {/* Severity Breakdown */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 lg:col-span-2">
        <h3 className="text-sm font-medium text-gray-600 mb-3">By Severity</h3>
        <div className="space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600">Critical</span>
            <div className="flex items-center gap-2">
              <div className="w-16 h-2 bg-red-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-red-600"
                  style={{
                    width: `${
                      stats.total_alerts > 0
                        ? (stats.by_severity.critical / stats.total_alerts) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-sm font-semibold text-red-600">
                {stats.by_severity.critical}
              </span>
            </div>
          </div>

          <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600">High</span>
            <div className="flex items-center gap-2">
              <div className="w-16 h-2 bg-orange-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-orange-600"
                  style={{
                    width: `${
                      stats.total_alerts > 0
                        ? (stats.by_severity.high / stats.total_alerts) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-sm font-semibold text-orange-600">
                {stats.by_severity.high}
              </span>
            </div>
          </div>

          <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600">Medium</span>
            <div className="flex items-center gap-2">
              <div className="w-16 h-2 bg-yellow-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-yellow-600"
                  style={{
                    width: `${
                      stats.total_alerts > 0
                        ? (stats.by_severity.medium / stats.total_alerts) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-sm font-semibold text-yellow-600">
                {stats.by_severity.medium}
              </span>
            </div>
          </div>

          <div className="flex justify-between items-center">
            <span className="text-sm text-gray-600">Low</span>
            <div className="flex items-center gap-2">
              <div className="w-16 h-2 bg-blue-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-600"
                  style={{
                    width: `${
                      stats.total_alerts > 0
                        ? (stats.by_severity.low / stats.total_alerts) * 100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <span className="text-sm font-semibold text-blue-600">
                {stats.by_severity.low}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Alert Sources */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 lg:col-span-2">
        <h3 className="text-sm font-medium text-gray-600 mb-3">By Source</h3>
        <div className="space-y-2">
          {Object.entries(stats.by_source).map(([source, count]) => (
            <div key={source} className="flex justify-between items-center">
              <span className="text-sm text-gray-600 capitalize">{source}</span>
              <span className="text-sm font-semibold text-gray-900">{count}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default DashboardMetrics;
