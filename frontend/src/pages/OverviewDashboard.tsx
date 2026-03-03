import React, { useMemo } from 'react';
import {
  AlertTriangle,
  Activity,
  TrendingUp,
  Clock,
} from 'lucide-react';
import { PageWrapper } from '../components/common/PageWrapper';
import { MetricCard } from '../components/cards/MetricCard';
import { StatusBadge } from '../components/cards/StatusBadge';
import { HealthGauge } from '../components/charts/HealthGauge';
import { LineChart, ChartDataPoint } from '../components/charts/LineChart';
import { DataTable, Column } from '../components/tables/DataTable';
import { calculateOverallHealth } from '../utils/calculations';
import { formatTimeAgo } from '../utils/formatting';

// Mock data for demonstration
const mockHealthMetrics = {
  lockHealth: 85,
  bloatHealth: 60,
  queryHealth: 75,
  cacheHealth: 70,
  connectionHealth: 90,
  replicationHealth: 95,
};

const mockAlerts = [
  {
    id: '1',
    severity: 'critical' as const,
    title: 'Lock Contention',
    status: 'active' as const,
    fired_at: new Date(Date.now() - 5 * 60000), // 5 minutes ago
  },
  {
    id: '2',
    severity: 'warning' as const,
    title: 'Table Bloat',
    status: 'active' as const,
    fired_at: new Date(Date.now() - 15 * 60000), // 15 minutes ago
  },
  {
    id: '3',
    severity: 'info' as const,
    title: 'Replication Lag',
    status: 'resolved' as const,
    fired_at: new Date(Date.now() - 2 * 3600000), // 2 hours ago
  },
];

const mockHealthHistory: ChartDataPoint[] = [
  { name: '12:00', value: 78 },
  { name: '13:00', value: 80 },
  { name: '14:00', value: 82 },
  { name: '15:00', value: 85 },
  { name: '16:00', value: 84 },
  { name: '17:00', value: 85 },
  { name: '18:00', value: 86 },
];

export const OverviewDashboard: React.FC = () => {
  const overallHealth = useMemo(
    () => calculateOverallHealth(mockHealthMetrics),
    []
  );

  const criticalCount = mockAlerts.filter(a => a.severity === 'critical').length;
  const warningCount = mockAlerts.filter(a => a.severity === 'warning').length;
  const activeCount = mockAlerts.filter(a => a.status === 'active').length;

  const alertColumns: Column<typeof mockAlerts[0]>[] = [
    {
      key: 'severity',
      label: 'Severity',
      render: (value) => (
        <StatusBadge
          status={value === 'critical' ? 'error' : value === 'warning' ? 'warning' : 'info'}
          label={String(value).toUpperCase()}
          size="sm"
        />
      ),
    },
    {
      key: 'title',
      label: 'Alert',
      sortable: true,
    },
    {
      key: 'status',
      label: 'Status',
      render: (value) => (
        <StatusBadge
          status={value === 'active' ? 'error' : 'success'}
          label={String(value).toUpperCase()}
          size="sm"
        />
      ),
    },
    {
      key: 'fired_at',
      label: 'Fired',
      render: (value) => formatTimeAgo(value as Date),
    },
  ];

  return (
    <PageWrapper title="Dashboard" description="System health and recent activity at a glance">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <MetricCard
          title="Health Score"
          value={overallHealth}
          unit="/100"
          icon={<Activity className="w-6 h-6" />}
          status={overallHealth >= 80 ? 'healthy' : overallHealth >= 60 ? 'warning' : 'critical'}
          trend="up"
          trendValue="+5%"
        />
        <MetricCard
          title="Critical Alerts"
          value={criticalCount}
          icon={<AlertTriangle className="w-6 h-6" />}
          status={criticalCount > 0 ? 'critical' : 'healthy'}
          onClick={() => console.log('View critical alerts')}
        />
        <MetricCard
          title="Warnings"
          value={warningCount}
          icon={<TrendingUp className="w-6 h-6" />}
          status={warningCount > 0 ? 'warning' : 'healthy'}
        />
        <MetricCard
          title="Active Issues"
          value={activeCount}
          icon={<Clock className="w-6 h-6" />}
          status={activeCount > 0 ? 'warning' : 'healthy'}
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Health Gauge */}
        <div className="lg:col-span-1 bg-white rounded-lg shadow p-6">
          <HealthGauge
            score={overallHealth}
            title="Overall Health"
            size="md"
            showTrend={true}
            trendValue={5}
          />
        </div>

        {/* Health Score Breakdown */}
        <div className="lg:col-span-2 bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-pg-dark mb-4">Health Breakdown</h3>
          <div className="grid grid-cols-2 gap-4">
            {[
              { label: 'Lock Health', value: mockHealthMetrics.lockHealth },
              { label: 'Bloat Health', value: mockHealthMetrics.bloatHealth },
              { label: 'Query Health', value: mockHealthMetrics.queryHealth },
              { label: 'Cache Health', value: mockHealthMetrics.cacheHealth },
              { label: 'Connection Health', value: mockHealthMetrics.connectionHealth },
              { label: 'Replication Health', value: mockHealthMetrics.replicationHealth },
            ].map((item) => (
              <div key={item.label} className="space-y-2">
                <div className="flex justify-between items-center text-sm">
                  <span className="text-pg-slate">{item.label}</span>
                  <span className="font-semibold text-pg-dark">{item.value}</span>
                </div>
                <div className="bg-pg-slate/10 rounded-full h-2 overflow-hidden">
                  <div
                    className={`h-full transition-all ${
                      item.value >= 80
                        ? 'bg-pg-success'
                        : item.value >= 60
                        ? 'bg-pg-warning'
                        : 'bg-pg-danger'
                    }`}
                    style={{ width: `${item.value}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Health Score Timeline */}
      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <h3 className="text-lg font-semibold text-pg-dark mb-4">Health Score History</h3>
        <LineChart
          data={mockHealthHistory}
          height={250}
          lines={[
            {
              key: 'value',
              stroke: '#06b6d4',
              name: 'Health Score',
            },
          ]}
        />
      </div>

      {/* Recent Alerts */}
      <div className="bg-white rounded-lg shadow p-6">
        <DataTable
          title="Recent Alerts"
          columns={alertColumns}
          data={mockAlerts.map(alert => ({
            ...alert,
            fired_at: alert.fired_at,
          }))}
          searchable={true}
          emptyMessage="No alerts found"
        />
      </div>
    </PageWrapper>
  );
};
