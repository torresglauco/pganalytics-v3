import React, { useState, useEffect } from 'react';
import {
  AlertCircle,
  CheckCircle,
  Download,
  Search,
  Filter,
  RefreshCw,
} from 'lucide-react';
import type { AlertIncident, AlertStats, AlertFilters } from '../types/alertDashboard';
import {
  listAlerts,
  getAlertStats,
  acknowledgeAlerts,
  resolveAlerts,
} from '../api/alertDashboardApi';
import AlertsTable from '../components/AlertsTable';
import AlertFiltersPanel from '../components/AlertFiltersPanel';
import DashboardMetrics from '../components/DashboardMetrics';
import AlertDetailPanel from '../components/AlertDetailPanel';
import BulkAlertActions from '../components/BulkAlertActions';

export const AlertsDashboard: React.FC = () => {
  const [alerts, setAlerts] = useState<AlertIncident[]>([]);
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // UI State
  const [selectedAlertId, setSelectedAlertId] = useState<string | null>(null);
  const [selectedAlerts, setSelectedAlerts] = useState<Set<string>>(new Set());
  const [showFilters, setShowFilters] = useState(false);

  // Filters
  const [filters, setFilters] = useState<AlertFilters>({});
  const [searchTerm, setSearchTerm] = useState('');

  // Pagination
  const [limit] = useState(50);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);

  // Auto-refresh
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval] = useState(10000); // 10 seconds

  /**
   * Load alerts and stats
   */
  useEffect(() => {
    loadAlertsAndStats();
  }, [filters, searchTerm, offset]);

  /**
   * Auto-refresh alerts
   */
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      loadAlertsAndStats();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  const loadAlertsAndStats = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [alertsResponse, statsData] = await Promise.all([
        listAlerts({
          filters: {
            ...filters,
            search_term: searchTerm || undefined,
          },
          limit,
          offset,
          sort_by: 'fired_at',
          sort_order: 'desc',
        }),
        getAlertStats(),
      ]);

      setAlerts(alertsResponse.alerts);
      setTotal(alertsResponse.total);
      setStats(statsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load alerts');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Handle alert selection
   */
  const handleSelectAlert = (alertId: string) => {
    setSelectedAlertId(alertId);
  };

  /**
   * Toggle alert checkbox
   */
  const toggleAlertSelection = (alertId: string) => {
    const newSelected = new Set(selectedAlerts);
    if (newSelected.has(alertId)) {
      newSelected.delete(alertId);
    } else {
      newSelected.add(alertId);
    }
    setSelectedAlerts(newSelected);
  };

  /**
   * Select all alerts
   */
  const toggleSelectAll = () => {
    if (selectedAlerts.size === alerts.length) {
      setSelectedAlerts(new Set());
    } else {
      setSelectedAlerts(new Set(alerts.map((a) => a.id)));
    }
  };

  /**
   * Acknowledge selected alerts
   */
  const handleAcknowledge = async (notes?: string) => {
    const alertIds = Array.from(selectedAlerts);
    if (alertIds.length === 0) return;

    try {
      await acknowledgeAlerts(alertIds, notes);
      await loadAlertsAndStats();
      setSelectedAlerts(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to acknowledge alerts');
    }
  };

  /**
   * Resolve selected alerts
   */
  const handleResolve = async (notes?: string) => {
    const alertIds = Array.from(selectedAlerts);
    if (alertIds.length === 0) return;

    try {
      await resolveAlerts(alertIds, notes);
      await loadAlertsAndStats();
      setSelectedAlerts(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resolve alerts');
    }
  };

  /**
   * Handle export
   */
  const handleExport = async () => {
    try {
      // TODO: Implement export
      setError('Export functionality coming soon');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed');
    }
  };

  if (selectedAlertId) {
    const alert = alerts.find((a) => a.id === selectedAlertId);
    if (alert) {
      return (
        <AlertDetailPanel
          alert={alert}
          onClose={() => setSelectedAlertId(null)}
          onAcknowledge={() => {
            handleAcknowledge();
            setSelectedAlertId(null);
          }}
          onResolve={() => {
            handleResolve();
            setSelectedAlertId(null);
          }}
        />
      );
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Alerts & Incidents</h1>
          <p className="text-sm text-gray-600 mt-1">
            Monitor and manage active alerts across all databases
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={loadAlertsAndStats}
            className={`flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium ${
              isLoading ? 'animate-spin' : ''
            }`}
          >
            <RefreshCw size={18} />
            Refresh
          </button>
          <button
            onClick={handleExport}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium"
          >
            <Download size={18} />
            Export
          </button>
        </div>
      </div>

      {/* Auto-refresh toggle */}
      <div className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          id="auto_refresh"
          checked={autoRefresh}
          onChange={(e) => setAutoRefresh(e.target.checked)}
          className="rounded"
        />
        <label htmlFor="auto_refresh" className="text-gray-700">
          Auto-refresh every {refreshInterval / 1000}s
        </label>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 flex gap-2">
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div className="flex-1">{error}</div>
        </div>
      )}

      {/* Dashboard Metrics */}
      {stats && <DashboardMetrics stats={stats} />}

      {/* Search and Filters */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 space-y-4">
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <Search size={18} className="absolute left-3 top-2.5 text-gray-400" />
            <input
              type="text"
              placeholder="Search alerts..."
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value);
                setOffset(0);
              }}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700"
          >
            <Filter size={18} />
            Filters
          </button>
        </div>

        {/* Filters Panel */}
        {showFilters && (
          <AlertFiltersPanel
            filters={filters}
            onChange={(newFilters) => {
              setFilters(newFilters);
              setOffset(0);
            }}
          />
        )}
      </div>

      {/* Bulk Actions */}
      {selectedAlerts.size > 0 && (
        <BulkAlertActions
          selectedCount={selectedAlerts.size}
          onAcknowledge={handleAcknowledge}
          onResolve={handleResolve}
        />
      )}

      {/* Alerts Table */}
      {isLoading ? (
        <div className="p-8 text-center text-gray-600">Loading alerts...</div>
      ) : alerts.length === 0 ? (
        <div className="p-8 text-center text-gray-600">
          <CheckCircle size={32} className="mx-auto mb-2 opacity-50" />
          <p>No active alerts</p>
          <p className="text-sm mt-2">All systems are running smoothly!</p>
        </div>
      ) : (
        <>
          <AlertsTable
            alerts={alerts}
            selectedAlerts={selectedAlerts}
            onSelectAlert={handleSelectAlert}
            onToggleSelection={toggleAlertSelection}
            onToggleSelectAll={toggleSelectAll}
          />

          {/* Pagination */}
          {total > limit && (
            <div className="flex justify-between items-center">
              <p className="text-sm text-gray-600">
                Showing {offset + 1}-{Math.min(offset + limit, total)} of {total} alerts
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                  className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() =>
                    setOffset(Math.min(offset + limit, Math.max(0, total - limit)))
                  }
                  disabled={offset + limit >= total}
                  className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default AlertsDashboard;
