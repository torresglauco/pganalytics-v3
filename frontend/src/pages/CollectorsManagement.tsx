import React, { useMemo, useState } from 'react';
import {
  Plus,
  Trash2,
  RefreshCw,
  Eye,
  EyeOff,
  Copy,
  Check,
  AlertTriangle,
  AlertCircle,
} from 'lucide-react';
import { MainLayout } from '../components/layout/MainLayout';
import { PageWrapper } from '../components/common/PageWrapper';
import { MetricCard } from '../components/cards/MetricCard';
import { StatusBadge } from '../components/cards/StatusBadge';
import { DataTable, Column } from '../components/tables/DataTable';
import { formatTimeAgo } from '../utils/formatting';
import { useCollectors } from '../hooks/useCollectors';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';
import type { Collector } from '../types';

interface NewCollectorForm {
  name: string;
  host: string;
  port: string;
  database: string;
  username: string;
  password: string;
}

// Transform API collector to display format
function transformCollectorForDisplay(collector: Collector) {
  return {
    ...collector,
    name: collector.hostname || 'Unknown',
    host: collector.hostname || 'N/A',
    port: 5432, // Default, would come from config
    database: 'N/A', // Not in API response
    status: (collector.status === 'active' ? 'online' : collector.status === 'inactive' ? 'offline' : 'error') as 'online' | 'offline' | 'error',
    health_score: 85, // Would come from metrics
    last_heartbeat: collector.last_seen ? new Date(collector.last_seen) : new Date(),
    metrics_collected: collector.metrics_count_total || 0,
    collection_interval: 60,
    version: collector.version || 'unknown',
    created_at: new Date(collector.created_at),
  };
}

interface DisplayCollector extends Collector {
  status: 'online' | 'offline' | 'error';
  health_score: number;
  last_heartbeat: Date;
  metrics_collected: number;
  collection_interval: number;
  database: string;
  port: number;
}

export const CollectorsManagement: React.FC = () => {
  const { collectors, loading, error, fetchCollectors, createCollector, deleteCollector } = useCollectors();
  const [showNewForm, setShowNewForm] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [visibleSecrets, setVisibleSecrets] = useState<Set<string>>(new Set());
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [formData, setFormData] = useState<NewCollectorForm>({
    name: '',
    host: '',
    port: '5432',
    database: '',
    username: '',
    password: '',
  });

  // Transform collectors for display
  const displayCollectors: DisplayCollector[] = useMemo(() => {
    return collectors.map(c => transformCollectorForDisplay(c)) as DisplayCollector[];
  }, [collectors]);

  // Calculate stats
  const stats = useMemo(() => {
    return {
      total: displayCollectors.length,
      online: displayCollectors.filter((c) => c.status === 'online').length,
      offline: displayCollectors.filter((c) => c.status === 'offline').length,
      error: displayCollectors.filter((c) => c.status === 'error').length,
      avgHealth:
        displayCollectors.length > 0
          ? Math.round(
              displayCollectors.reduce((sum, c) => sum + c.health_score, 0) / displayCollectors.length
            )
          : 0,
    };
  }, [displayCollectors]);

  const toggleSecretVisibility = (id: string) => {
    const newSet = new Set(visibleSecrets);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setVisibleSecrets(newSet);
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleAddCollector = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setSubmitError(null);

    try {
      await createCollector({
        name: formData.name,
        host: formData.host,
        port: parseInt(formData.port) || 5432,
        database: formData.database,
        username: formData.username,
        password: formData.password,
      });
      setShowNewForm(false);
      setFormData({
        name: '',
        host: '',
        port: '5432',
        database: '',
        username: '',
        password: '',
      });
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create collector';
      setSubmitError(errorMsg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteCollector = async (id: string) => {
    try {
      await deleteCollector(id);
      setDeleteConfirm(null);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete collector';
      setSubmitError(errorMsg);
    }
  };

  const columns: Column<DisplayCollector>[] = [
    {
      key: 'name',
      label: 'Collector',
      sortable: true,
      render: (value, row) => (
        <div className="space-y-1">
          <div className="font-medium text-pg-dark">{String(value)}</div>
          <div className="text-xs text-pg-slate">
            {row.host}:{row.port}/{row.database}
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      width: '100px',
      render: (value) => {
        const status = value === 'online' ? 'success' : value === 'error' ? 'error' : 'warning';
        return (
          <StatusBadge
            status={status}
            label={String(value).toUpperCase()}
            size="sm"
          />
        );
      },
    },
    {
      key: 'health_score',
      label: 'Health',
      width: '70px',
      render: (value) => {
        const score = value as number;
        const color =
          score >= 80 ? 'text-pg-success' : score >= 60 ? 'text-pg-warning' : 'text-pg-danger';
        return <span className={`font-semibold ${color}`}>{score}</span>;
      },
    },
    {
      key: 'metrics_collected',
      label: 'Metrics',
      width: '80px',
      render: (value) => <span className="text-sm text-pg-slate">{String(value)}</span>,
    },
    {
      key: 'last_heartbeat',
      label: 'Last Heartbeat',
      width: '120px',
      render: (value) => <span className="text-sm">{formatTimeAgo(value as Date)}</span>,
    },
    {
      key: 'version',
      label: 'Version',
      width: '80px',
      render: (value) => <span className="text-sm text-pg-slate">v{String(value)}</span>,
    },
  ];

  if (loading && displayCollectors.length === 0) {
    return <LoadingSpinner fullScreen message="Loading collectors..." />;
  }

  return (
    <MainLayout>
      <PageWrapper
        title="Collectors Management"
        description="Manage PostgreSQL database collectors and monitor their health"
      >
      {/* Error Messages */}
      {error && (
        <div className="mb-6 p-4 bg-pg-danger/10 border border-pg-danger/20 rounded-lg flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-pg-danger flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="font-semibold text-pg-dark">Error</h3>
            <p className="text-sm text-pg-dark/70">{error.message}</p>
          </div>
        </div>
      )}

      {submitError && (
        <div className="mb-6 p-4 bg-pg-danger/10 border border-pg-danger/20 rounded-lg flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-pg-danger flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="font-semibold text-pg-dark">Form Error</h3>
            <p className="text-sm text-pg-dark/70">{submitError}</p>
          </div>
        </div>
      )}

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-3 mb-8">
        <MetricCard
          title="Total"
          value={stats.total}
          status="info"
        />
        <MetricCard
          title="Online"
          value={stats.online}
          status={stats.online === stats.total ? 'healthy' : 'warning'}
          icon={<Check className="w-5 h-5" />}
        />
        <MetricCard
          title="Offline"
          value={stats.offline}
          status={stats.offline > 0 ? 'warning' : 'healthy'}
        />
        <MetricCard
          title="Errors"
          value={stats.error}
          status={stats.error > 0 ? 'critical' : 'healthy'}
        />
        <MetricCard
          title="Avg Health"
          value={stats.avgHealth}
          unit="/100"
          status={stats.avgHealth >= 80 ? 'healthy' : stats.avgHealth >= 60 ? 'warning' : 'critical'}
        />
      </div>

      {/* Add New Collector Form */}
      {showNewForm && (
        <div className="bg-white rounded-lg shadow p-6 mb-6 border-l-4 border-pg-blue">
          <h3 className="text-lg font-semibold text-pg-dark mb-4">Register New Collector</h3>
          <form onSubmit={handleAddCollector}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">
                  Collector Name
                </label>
                <input
                  type="text"
                  placeholder="e.g., Production Primary"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Database</label>
                <input
                  type="text"
                  placeholder="e.g., maindb"
                  value={formData.database}
                  onChange={(e) => setFormData({ ...formData, database: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Host</label>
                <input
                  type="text"
                  placeholder="e.g., localhost"
                  value={formData.host}
                  onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Port</label>
                <input
                  type="number"
                  placeholder="5432"
                  value={formData.port}
                  onChange={(e) => setFormData({ ...formData, port: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Username</label>
                <input
                  type="text"
                  placeholder="e.g., pganalytics"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-pg-dark mb-2">Password</label>
                <input
                  type="password"
                  placeholder="••••••••"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
                  required
                  disabled={isSubmitting}
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={isSubmitting}
                className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
              >
                {isSubmitting ? 'Registering...' : 'Register Collector'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowNewForm(false);
                  setSubmitError(null);
                }}
                disabled={isSubmitting}
                className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Collectors Table */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-pg-dark">Active Collectors</h3>
          <div className="flex items-center gap-2">
            <button
              onClick={() => fetchCollectors()}
              disabled={loading}
              className="p-2 hover:bg-pg-slate/10 disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors"
              title="Refresh"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={() => setShowNewForm(!showNewForm)}
              className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Add Collector
            </button>
          </div>
        </div>
        <DataTable
          columns={columns}
          data={displayCollectors}
          searchable={true}
          emptyMessage={displayCollectors.length === 0 ? 'No collectors found' : ''}
        />
      </div>

      {/* Collector Actions */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h3 className="text-lg font-semibold text-pg-dark mb-4">Collector Actions</h3>
        {displayCollectors.length === 0 ? (
          <p className="text-sm text-pg-slate">No collectors registered. Add a collector from the table above.</p>
        ) : (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {displayCollectors.map((collector) => (
              <div key={collector.id} className="flex items-center justify-between p-3 bg-pg-slate/5 rounded-lg">
                <div>
                  <p className="text-sm font-medium text-pg-dark">{collector.name}</p>
                  <p className="text-xs text-pg-slate">{collector.host}:{collector.port}</p>
                </div>
                <div className="flex gap-2">
                  {deleteConfirm === collector.id ? (
                    <>
                      <button
                        onClick={() => handleDeleteCollector(collector.id)}
                        disabled={isSubmitting}
                        className="px-3 py-1 bg-pg-danger text-white rounded text-xs hover:bg-pg-danger/90 disabled:opacity-50 transition-colors"
                      >
                        {isSubmitting ? 'Deleting...' : 'Confirm Delete'}
                      </button>
                      <button
                        onClick={() => setDeleteConfirm(null)}
                        className="px-3 py-1 border border-pg-slate/20 text-pg-dark rounded text-xs hover:bg-pg-slate/5 transition-colors"
                      >
                        Cancel
                      </button>
                    </>
                  ) : (
                    <button
                      onClick={() => setDeleteConfirm(collector.id)}
                      className="flex items-center gap-2 px-3 py-1 text-pg-danger border border-pg-danger/20 rounded text-xs hover:bg-pg-danger/5 transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                      Delete Collector
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Registration Secrets */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h3 className="text-lg font-semibold text-pg-dark mb-4">Registration Secrets</h3>
        {displayCollectors.length === 0 ? (
          <p className="text-sm text-pg-slate">No collectors registered yet</p>
        ) : (
          <div className="space-y-3">
            {displayCollectors.map((collector) => (
              <div key={collector.id} className="flex items-center justify-between p-4 bg-pg-slate/5 rounded-lg">
                <div>
                  <h4 className="font-medium text-pg-dark">{collector.name || collector.hostname}</h4>
                  <p className="text-xs text-pg-slate">{collector.id}</p>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type={visibleSecrets.has(collector.id) ? 'text' : 'password'}
                    value={'sk_' + collector.id.slice(0, 16) + '...'}
                    readOnly
                    className="px-3 py-2 bg-white border border-pg-slate/20 rounded text-sm font-mono w-64"
                  />
                  <button
                    onClick={() => toggleSecretVisibility(collector.id)}
                    className="p-2 hover:bg-pg-slate/10 rounded transition-colors"
                    title={visibleSecrets.has(collector.id) ? 'Hide' : 'Show'}
                  >
                    {visibleSecrets.has(collector.id) ? (
                      <EyeOff className="w-4 h-4 text-pg-slate" />
                    ) : (
                      <Eye className="w-4 h-4 text-pg-slate" />
                    )}
                  </button>
                  <button
                    onClick={() =>
                      copyToClipboard(
                        'sk_' + collector.id.slice(0, 16),
                        collector.id
                      )
                    }
                    className="p-2 hover:bg-pg-slate/10 rounded transition-colors"
                    title="Copy to clipboard"
                  >
                    {copiedId === collector.id ? (
                      <Check className="w-4 h-4 text-pg-success" />
                    ) : (
                      <Copy className="w-4 h-4 text-pg-slate" />
                    )}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Danger Zone */}
      <div className="bg-pg-danger/5 rounded-lg p-6 border-l-4 border-pg-danger">
        <h3 className="text-lg font-semibold text-pg-dark mb-4 flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-pg-danger" />
          Danger Zone
        </h3>
        {displayCollectors.length === 0 ? (
          <p className="text-sm text-pg-slate">No collectors to delete</p>
        ) : (
          <div className="space-y-3">
            {displayCollectors.map((collector) => (
              <div
                key={collector.id}
                className="flex items-center justify-between p-4 bg-white rounded-lg border border-pg-danger/20"
              >
                <div>
                  <h4 className="font-medium text-pg-dark">{collector.name || collector.hostname}</h4>
                  <p className="text-xs text-pg-slate">{collector.id}</p>
                </div>
                {deleteConfirm === collector.id ? (
                  <div className="flex items-center gap-2">
                    <p className="text-sm text-pg-danger font-medium">Are you sure?</p>
                    <button
                      onClick={() => handleDeleteCollector(collector.id)}
                      className="px-3 py-1 bg-pg-danger text-white rounded text-sm hover:bg-pg-danger/90 transition-colors"
                    >
                      Delete
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(null)}
                      className="px-3 py-1 border border-pg-slate/20 text-pg-dark rounded text-sm hover:bg-pg-slate/5 transition-colors"
                    >
                      Cancel
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setDeleteConfirm(collector.id)}
                    className="flex items-center gap-2 px-3 py-2 text-pg-danger border border-pg-danger/20 rounded-lg hover:bg-pg-danger/5 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                    Delete Collector
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </PageWrapper>
    </MainLayout>
  );
};
