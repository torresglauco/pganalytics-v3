import React, { useMemo, useState } from 'react';
import {
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  X,
  Plus,
  Filter,
} from 'lucide-react';
import { MainLayout } from '../components/layout/MainLayout';
import { PageWrapper } from '../components/common/PageWrapper';
import { MetricCard } from '../components/cards/MetricCard';
import { StatusBadge } from '../components/cards/StatusBadge';
import { DataTable, Column } from '../components/tables/DataTable';
import { formatTimeAgo } from '../utils/formatting';
import { Alert, Incident, AlertSeverity, AlertStatus } from '../types/alerts';

// Mock data for demonstration
const mockAlerts: Alert[] = [
  {
    id: '1',
    collector_id: 'prod-db-01',
    alert_type: 'lock_contention',
    severity: 'critical',
    status: 'active',
    title: 'High Lock Contention',
    description: 'Database experiencing severe lock contention on pg_stat_statements table',
    value: 850,
    threshold: 100,
    unit: 'locks/min',
    fired_at: new Date(Date.now() - 5 * 60000),
    incident_id: 'INC-001',
  },
  {
    id: '2',
    collector_id: 'prod-db-01',
    alert_type: 'table_bloat',
    severity: 'warning',
    status: 'active',
    title: 'Table Bloat Detected',
    description: 'users_archive table bloat at 45% - recommend VACUUM FULL',
    value: 45,
    threshold: 30,
    unit: '%',
    fired_at: new Date(Date.now() - 15 * 60000),
    incident_id: 'INC-001',
  },
  {
    id: '3',
    collector_id: 'staging-db-02',
    alert_type: 'cache_miss',
    severity: 'warning',
    status: 'active',
    title: 'Low Cache Hit Ratio',
    description: 'Cache hit ratio dropped to 78% on staging',
    value: 78,
    threshold: 85,
    unit: '%',
    fired_at: new Date(Date.now() - 30 * 60000),
  },
  {
    id: '4',
    collector_id: 'prod-db-01',
    alert_type: 'connection_pool',
    severity: 'info',
    status: 'resolved',
    title: 'Connection Pool Spike',
    description: 'Connection pool used 95% capacity during peak hours',
    value: 95,
    threshold: 80,
    unit: '%',
    fired_at: new Date(Date.now() - 2 * 3600000),
    resolved_at: new Date(Date.now() - 1800000),
  },
  {
    id: '5',
    collector_id: 'prod-db-02',
    alert_type: 'replication_lag',
    severity: 'critical',
    status: 'active',
    title: 'High Replication Lag',
    description: 'Replication lag exceeded 5 seconds',
    value: 7.2,
    threshold: 5,
    unit: 's',
    fired_at: new Date(Date.now() - 10 * 60000),
  },
  {
    id: '6',
    collector_id: 'staging-db-02',
    alert_type: 'idle_transaction',
    severity: 'warning',
    status: 'muted',
    title: 'Idle Transactions',
    description: 'Long-running idle transactions detected',
    value: 12,
    threshold: 5,
    unit: 'txn',
    fired_at: new Date(Date.now() - 45 * 60000),
  },
];

const mockIncidents: Incident[] = [
  {
    id: 'INC-001',
    group_name: 'Database Performance Degradation',
    state: 'active',
    severity: 'critical',
    alerts: [mockAlerts[0], mockAlerts[1]],
    root_cause: 'Schema migration locked tables for 2 hours',
    confidence: 92,
    suggested_actions: [
      'Check for long-running DDL statements',
      'Review pg_stat_statements for blocking queries',
      'Consider LOCK TIMEOUT for future migrations',
    ],
    created_at: new Date(Date.now() - 5 * 60000),
    updated_at: new Date(Date.now() - 5 * 60000),
  },
  {
    id: 'INC-002',
    group_name: 'Maintenance Window Issues',
    state: 'resolved',
    severity: 'warning',
    alerts: [mockAlerts[3]],
    root_cause: 'Scheduled batch job caused connection spike',
    confidence: 85,
    suggested_actions: [
      'Reschedule batch jobs to off-peak hours',
      'Implement connection pooling with pgBouncer',
    ],
    created_at: new Date(Date.now() - 4 * 3600000),
    updated_at: new Date(Date.now() - 1800000),
    resolved_at: new Date(Date.now() - 1800000),
  },
];

interface AlertFilters {
  severity: AlertSeverity | 'all';
  status: AlertStatus | 'all';
  alertType: string;
  collectorId: string;
}

export const AlertsIncidents: React.FC = () => {
  const [filters, setFilters] = useState<AlertFilters>({
    severity: 'all',
    status: 'all',
    alertType: 'all',
    collectorId: 'all',
  });
  const [showSuppressionForm, setShowSuppressionForm] = useState(false);

  // Filter alerts based on selected filters
  const filteredAlerts = useMemo(() => {
    return mockAlerts.filter((alert) => {
      if (filters.severity !== 'all' && alert.severity !== filters.severity) return false;
      if (filters.status !== 'all' && alert.status !== filters.status) return false;
      if (filters.alertType !== 'all' && alert.alert_type !== filters.alertType) return false;
      if (filters.collectorId !== 'all' && alert.collector_id !== filters.collectorId)
        return false;
      return true;
    });
  }, [filters]);

  // Calculate alert counts
  const alertStats = useMemo(() => {
    return {
      total: mockAlerts.length,
      critical: mockAlerts.filter((a) => a.severity === 'critical').length,
      warning: mockAlerts.filter((a) => a.severity === 'warning').length,
      info: mockAlerts.filter((a) => a.severity === 'info').length,
      active: mockAlerts.filter((a) => a.status === 'active').length,
      resolved: mockAlerts.filter((a) => a.status === 'resolved').length,
      muted: mockAlerts.filter((a) => a.status === 'muted').length,
    };
  }, []);

  // Define table columns
  const alertColumns: Column<Alert>[] = [
    {
      key: 'severity',
      label: 'Severity',
      width: '80px',
      render: (value) => {
        const status = value === 'critical' ? 'error' : value === 'warning' ? 'warning' : 'info';
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
      key: 'title',
      label: 'Alert',
      sortable: true,
      render: (value, row) => (
        <div className="space-y-1">
          <div className="font-medium text-pg-dark">{String(value)}</div>
          <div className="text-sm text-pg-slate">{row.description}</div>
        </div>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      width: '80px',
      render: (value) => {
        const status =
          value === 'active' ? 'error' : value === 'resolved' ? 'success' : 'warning';
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
      key: 'collector_id',
      label: 'Collector',
      width: '120px',
      render: (value) => <span className="text-sm text-pg-slate">{String(value)}</span>,
    },
    {
      key: 'fired_at',
      label: 'Fired',
      width: '100px',
      render: (value) => <span className="text-sm">{formatTimeAgo(value as Date)}</span>,
    },
  ];

  return (
    <MainLayout>
      <PageWrapper
        title="Alerts & Incidents"
        description="Manage database alerts, incidents, and suppression rules"
      >
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-7 gap-3 mb-8">
        <MetricCard
          title="Total"
          value={alertStats.total}
          icon={<AlertCircle className="w-5 h-5" />}
          status="info"
        />
        <MetricCard
          title="Critical"
          value={alertStats.critical}
          icon={<AlertTriangle className="w-5 h-5" />}
          status={alertStats.critical > 0 ? 'critical' : 'healthy'}
        />
        <MetricCard
          title="Warning"
          value={alertStats.warning}
          icon={<AlertTriangle className="w-5 h-5" />}
          status={alertStats.warning > 0 ? 'warning' : 'healthy'}
        />
        <MetricCard
          title="Info"
          value={alertStats.info}
          icon={<AlertCircle className="w-5 h-5" />}
          status="info"
        />
        <MetricCard
          title="Active"
          value={alertStats.active}
          icon={<AlertTriangle className="w-5 h-5" />}
          status="warning"
        />
        <MetricCard
          title="Resolved"
          value={alertStats.resolved}
          icon={<CheckCircle className="w-5 h-5" />}
          status="healthy"
        />
        <MetricCard
          title="Muted"
          value={alertStats.muted}
          icon={<X className="w-5 h-5" />}
          status="info"
        />
      </div>

      {/* Filters and Actions */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-pg-slate" />
            <h3 className="text-lg font-semibold text-pg-dark">Filters</h3>
          </div>
          <button
            onClick={() => setShowSuppressionForm(!showSuppressionForm)}
            className="flex items-center gap-2 px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            New Suppression Rule
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-pg-dark mb-2">Severity</label>
            <select
              value={filters.severity}
              onChange={(e) =>
                setFilters({ ...filters, severity: e.target.value as AlertFilters['severity'] })
              }
              className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-pg-dark mb-2">Status</label>
            <select
              value={filters.status}
              onChange={(e) =>
                setFilters({ ...filters, status: e.target.value as AlertFilters['status'] })
              }
              className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="resolved">Resolved</option>
              <option value="muted">Muted</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-pg-dark mb-2">Alert Type</label>
            <select
              value={filters.alertType}
              onChange={(e) => setFilters({ ...filters, alertType: e.target.value })}
              className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
            >
              <option value="all">All Types</option>
              <option value="lock_contention">Lock Contention</option>
              <option value="table_bloat">Table Bloat</option>
              <option value="cache_miss">Cache Miss</option>
              <option value="connection_pool">Connection Pool</option>
              <option value="replication_lag">Replication Lag</option>
              <option value="idle_transaction">Idle Transaction</option>
              <option value="metrics_collection_failure">Collection Failure</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-pg-dark mb-2">Collector</label>
            <select
              value={filters.collectorId}
              onChange={(e) => setFilters({ ...filters, collectorId: e.target.value })}
              className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
            >
              <option value="all">All Collectors</option>
              <option value="prod-db-01">prod-db-01</option>
              <option value="prod-db-02">prod-db-02</option>
              <option value="staging-db-02">staging-db-02</option>
            </select>
          </div>
        </div>
      </div>

      {/* Suppression Rule Form */}
      {showSuppressionForm && (
        <div className="bg-white rounded-lg shadow p-6 mb-6 border-l-4 border-pg-blue">
          <h3 className="text-lg font-semibold text-pg-dark mb-4">Create Suppression Rule</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-pg-dark mb-2">Rule Name</label>
              <input
                type="text"
                placeholder="e.g., Maintenance Window"
                className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-pg-dark mb-2">Alert Type</label>
              <select className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue">
                <option>All Types</option>
                <option>lock_contention</option>
                <option>table_bloat</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-pg-dark mb-2">Collector</label>
              <select className="w-full px-3 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-blue">
                <option>All Collectors</option>
                <option>prod-db-01</option>
              </select>
            </div>
          </div>
          <div className="flex gap-2">
            <button className="px-4 py-2 bg-pg-blue text-white rounded-lg hover:bg-pg-blue/90 transition-colors text-sm font-medium">
              Create Rule
            </button>
            <button
              onClick={() => setShowSuppressionForm(false)}
              className="px-4 py-2 border border-pg-slate/20 text-pg-dark rounded-lg hover:bg-pg-slate/5 transition-colors text-sm font-medium"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Active Incidents */}
      {mockIncidents.filter((i) => i.state === 'active').length > 0 && (
        <div className="bg-pg-danger/5 rounded-lg p-6 mb-6 border-l-4 border-pg-danger">
          <h3 className="text-lg font-semibold text-pg-dark mb-4">Active Incidents</h3>
          <div className="space-y-4">
            {mockIncidents
              .filter((i) => i.state === 'active')
              .map((incident) => (
                <div
                  key={incident.id}
                  className="bg-white rounded-lg p-4 border-l-4 border-pg-danger"
                >
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <h4 className="font-semibold text-pg-dark">{incident.group_name}</h4>
                      <p className="text-sm text-pg-slate">{incident.id}</p>
                    </div>
                    <StatusBadge
                      status="error"
                      label={incident.state.toUpperCase()}
                      size="sm"
                    />
                  </div>
                  {incident.root_cause && (
                    <div className="mb-3 p-3 bg-pg-danger/5 rounded text-sm text-pg-dark">
                      <strong>Root Cause:</strong> {incident.root_cause}
                    </div>
                  )}
                  <div className="text-sm text-pg-slate">
                    <strong>Confidence:</strong> {incident.confidence}%
                  </div>
                  {incident.suggested_actions.length > 0 && (
                    <div className="mt-3">
                      <strong className="text-sm text-pg-dark">Suggested Actions:</strong>
                      <ul className="text-sm text-pg-slate mt-2 list-disc list-inside space-y-1">
                        {incident.suggested_actions.map((action, idx) => (
                          <li key={idx}>{action}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              ))}
          </div>
        </div>
      )}

      {/* Alerts Table */}
      <div className="bg-white rounded-lg shadow p-6">
        <DataTable
          title={`Recent Alerts (${filteredAlerts.length})`}
          columns={alertColumns}
          data={filteredAlerts}
          searchable={true}
          emptyMessage="No alerts match the selected filters"
        />
      </div>
    </PageWrapper>
    </MainLayout>
  );
};
