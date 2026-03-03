# Frontend Implementation Guide
## Code Examples & Architecture for Enhanced UI

**Date**: March 3, 2026
**Scope**: Implementation details, code structure, and component patterns

---

## Part 1: New Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── common/              # Shared components
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Breadcrumb.tsx
│   │   │   ├── PageWrapper.tsx
│   │   │   └── NotificationCenter.tsx
│   │   │
│   │   ├── charts/              # Chart components
│   │   │   ├── LineChart.tsx
│   │   │   ├── BarChart.tsx
│   │   │   ├── GaugeChart.tsx
│   │   │   ├── HeatmapChart.tsx
│   │   │   └── SankeyChart.tsx
│   │   │
│   │   ├── tables/              # Data table components
│   │   │   ├── AdvancedDataTable.tsx
│   │   │   ├── AlertsTable.tsx
│   │   │   ├── LocksTable.tsx
│   │   │   └── ConnectionsTable.tsx
│   │   │
│   │   ├── cards/               # Card/widget components
│   │   │   ├── MetricCard.tsx
│   │   │   ├── StatusCard.tsx
│   │   │   ├── AlertCard.tsx
│   │   │   └── RecommendationCard.tsx
│   │   │
│   │   ├── modals/              # Modal dialogs
│   │   │   ├── ConfirmationModal.tsx
│   │   │   ├── DetailsPanel.tsx
│   │   │   ├── SettingsModal.tsx
│   │   │   └── RunbookModal.tsx
│   │   │
│   │   └── (existing)
│   │       ├── CollectorForm.tsx
│   │       ├── LoginForm.tsx
│   │       └── ...
│   │
│   ├── pages/
│   │   ├── Dashboard.tsx              # Existing collector management
│   │   ├── AuthPage.tsx               # Existing auth
│   │   ├── OverviewDashboard.tsx      # NEW
│   │   ├── AlertsIncidents.tsx        # NEW
│   │   ├── QueryPerformance.tsx       # NEW
│   │   ├── LockContention.tsx         # NEW
│   │   ├── BloatAnalysis.tsx          # NEW
│   │   ├── ConnectionManagement.tsx   # NEW
│   │   ├── CachePerformance.tsx       # NEW
│   │   ├── SchemaExplorer.tsx         # NEW
│   │   ├── ReplicationStatus.tsx      # NEW
│   │   ├── DatabaseHealth.tsx         # NEW
│   │   ├── ExtensionsConfig.tsx       # NEW
│   │   ├── CollectorsManagement.tsx   # Enhanced
│   │   └── Settings.tsx               # NEW
│   │
│   ├── hooks/
│   │   ├── useCollectors.ts           # Existing
│   │   ├── useAlerts.ts               # NEW
│   │   ├── useIncidents.ts            # NEW
│   │   ├── useMetrics.ts              # NEW
│   │   ├── useLocks.ts                # NEW
│   │   ├── useQueries.ts              # NEW
│   │   ├── useConnections.ts          # NEW
│   │   ├── useSchemaData.ts           # NEW
│   │   └── useNotifications.ts        # NEW
│   │
│   ├── store/                    # State management (Zustand)
│   │   ├── alertStore.ts         # Alert state
│   │   ├── uiStore.ts            # UI state (sidebar, theme, etc)
│   │   └── notificationStore.ts  # Notification preferences
│   │
│   ├── types/
│   │   ├── index.ts              # Existing
│   │   ├── alerts.ts             # NEW
│   │   ├── incidents.ts          # NEW
│   │   ├── metrics.ts            # NEW
│   │   └── api.ts                # NEW
│   │
│   ├── services/
│   │   ├── api.ts                # Existing
│   │   ├── alertService.ts       # NEW
│   │   ├── incidentService.ts    # NEW
│   │   └── metricsService.ts     # NEW
│   │
│   ├── utils/
│   │   ├── formatting.ts         # Number, date formatting
│   │   ├── calculations.ts       # Health score, bloat ratio, etc
│   │   ├── colors.ts             # Color helpers
│   │   └── constants.ts          # App constants
│   │
│   ├── styles/
│   │   ├── index.css             # Existing
│   │   ├── colors.css            # Color variables
│   │   ├── animations.css        # Custom animations
│   │   └── charts.css            # Chart specific styles
│   │
│   ├── App.tsx                   # Existing
│   └── main.tsx                  # Existing
│
├── public/
│   ├── logo-pganalytics.svg
│   ├── icons/
│   └── ...
│
└── (existing config files)
    ├── tailwind.config.js
    ├── vite.config.ts
    ├── tsconfig.json
    └── package.json
```

---

## Part 2: Component Patterns & Code Examples

### 1. Enhanced Tailwind Configuration

**File**: `frontend/tailwind.config.js`

```javascript
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'pg-blue': '#1e3a8a',        // Primary
        'pg-cyan': '#06b6d4',         // Accent
        'pg-success': '#10b981',      // Success/Healthy
        'pg-warning': '#f59e0b',      // Warning/Caution
        'pg-danger': '#f43f5e',       // Danger/Critical
        'pg-slate': '#64748b',        // Neutral/Text
        'pg-dark': '#0f172a',         // Dark mode bg
      },
      fontFamily: {
        'sans': ['Inter', 'sans-serif'],
        'mono': ['Fira Code', 'monospace'],
      },
      animation: {
        'pulse-subtle': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'slide-in': 'slideIn 0.3s ease-out',
        'fade-in': 'fadeIn 0.2s ease-out',
      },
      keyframes: {
        slideIn: {
          '0%': { transform: 'translateX(-100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms'),
  ],
}
```

### 2. Type Definitions

**File**: `frontend/src/types/alerts.ts`

```typescript
export type AlertSeverity = 'critical' | 'warning' | 'info';
export type AlertStatus = 'active' | 'resolved' | 'muted';
export type AlertType =
  | 'lock_contention'
  | 'table_bloat'
  | 'cache_miss'
  | 'connection_pool'
  | 'idle_transaction'
  | 'replication_lag'
  | 'metrics_collection_failure';

export interface Alert {
  id: string;
  collector_id: string;
  alert_type: AlertType;
  severity: AlertSeverity;
  status: AlertStatus;
  title: string;
  description: string;
  value?: number;
  threshold?: number;
  unit?: string;
  fired_at: Date;
  resolved_at?: Date;
  incident_id?: string;
  runbook_link?: string;
}

export interface Incident {
  id: string;
  group_name: string;
  state: 'active' | 'acknowledged' | 'resolved';
  severity: AlertSeverity;
  alerts: Alert[];
  root_cause?: string;
  confidence: number;
  suggested_actions: string[];
  created_at: Date;
  updated_at: Date;
  resolved_at?: Date;
}

export interface SuppressionRule {
  id: string;
  name: string;
  alert_type: AlertType;
  collector_id?: string;
  enabled: boolean;
  time_based?: {
    start_hour: number;
    end_hour: number;
    days: string[];
  };
}
```

**File**: `frontend/src/types/metrics.ts`

```typescript
export interface TimeSeriesPoint {
  timestamp: Date;
  value: number;
  collector_id: string;
}

export interface MetricsSummary {
  lockCount: number;
  bloatRatio: number;
  cacheHitRatio: number;
  connectionCount: number;
  replicationLag: number;
}

export interface QueryMetrics {
  query_id: string;
  query_text: string;
  calls: number;
  total_time: number;
  mean_time: number;
  rows: number;
  100pct_time: number;
  database: string;
}

export interface LockMetrics {
  lock_id: string;
  blocking_pid: number;
  blocked_pid: number;
  lock_type: string;
  granted: boolean;
  duration_ms: number;
  blocking_query?: string;
}

export interface BloatMetrics {
  table_name: string;
  dead_tuples: number;
  live_tuples: number;
  bloat_ratio: number;
  reclaimable_bytes: number;
  last_vacuum: Date;
}

export interface CacheMetrics {
  table_name: string;
  heap_blks_hit: number;
  heap_blks_read: number;
  hit_ratio: number;
}

export interface ConnectionMetrics {
  pid: number;
  usename: string;
  state: 'active' | 'idle' | 'idle in transaction';
  query: string;
  query_start: Date;
  duration_ms: number;
}

export interface HealthScore {
  overall: number;
  lock_health: number;
  bloat_health: number;
  query_health: number;
  cache_health: number;
  connection_health: number;
  replication_health: number;
  timestamp: Date;
}
```

### 3. Zustand Store Example

**File**: `frontend/src/store/alertStore.ts`

```typescript
import { create } from 'zustand';
import { Alert, AlertSeverity, AlertStatus } from '../types/alerts';

interface AlertFilters {
  severity?: AlertSeverity;
  status?: AlertStatus;
  collectorId?: string;
  timeRange?: '1h' | '24h' | '7d' | '30d';
}

interface AlertStore {
  alerts: Alert[];
  filters: AlertFilters;
  loading: boolean;
  error: string | null;

  // Actions
  setAlerts: (alerts: Alert[]) => void;
  setFilters: (filters: AlertFilters) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  // Computed
  filteredAlerts: () => Alert[];
  alertCounts: () => { critical: number; warning: number; info: number; };
}

export const useAlertStore = create<AlertStore>((set, get) => ({
  alerts: [],
  filters: {},
  loading: false,
  error: null,

  setAlerts: (alerts) => set({ alerts }),
  setFilters: (filters) => set({ filters }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),

  filteredAlerts: () => {
    const { alerts, filters } = get();
    return alerts.filter(alert => {
      if (filters.severity && alert.severity !== filters.severity) return false;
      if (filters.status && alert.status !== filters.status) return false;
      if (filters.collectorId && alert.collector_id !== filters.collectorId) return false;
      return true;
    });
  },

  alertCounts: () => {
    const filtered = get().filteredAlerts();
    return {
      critical: filtered.filter(a => a.severity === 'critical').length,
      warning: filtered.filter(a => a.severity === 'warning').length,
      info: filtered.filter(a => a.severity === 'info').length,
    };
  },
}));
```

### 4. Custom Hook Example

**File**: `frontend/src/hooks/useAlerts.ts`

```typescript
import { useEffect, useState } from 'react';
import { Alert } from '../types/alerts';
import { apiClient } from '../services/api';
import { useAlertStore } from '../store/alertStore';

export const useAlerts = (autoRefresh = true, refreshInterval = 30000) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { alerts, setAlerts, filters, setLoading: setStoreLoading } = useAlertStore();

  const fetchAlerts = async () => {
    try {
      setIsLoading(true);
      setStoreLoading(true);
      setError(null);

      // Build query parameters
      const params = new URLSearchParams();
      if (filters.severity) params.append('severity', filters.severity);
      if (filters.status) params.append('status', filters.status);
      if (filters.collectorId) params.append('collector_id', filters.collectorId);
      params.append('limit', '100');
      params.append('offset', '0');

      const response = await fetch(
        `/api/v1/alerts?${params.toString()}`,
        {
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );

      if (!response.ok) throw new Error('Failed to fetch alerts');

      const data = await response.json();
      setAlerts(data.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch alerts');
    } finally {
      setIsLoading(false);
      setStoreLoading(false);
    }
  };

  // Auto-refresh
  useEffect(() => {
    fetchAlerts();

    if (autoRefresh) {
      const interval = setInterval(fetchAlerts, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [filters, autoRefresh, refreshInterval]);

  const acknowledgeAlert = async (alertId: string) => {
    try {
      await fetch(
        `/api/v1/alerts/${alertId}/acknowledge`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );
      await fetchAlerts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to acknowledge alert');
    }
  };

  const muteAlert = async (alertId: string) => {
    try {
      await fetch(
        `/api/v1/alerts/${alertId}/mute`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );
      await fetchAlerts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mute alert');
    }
  };

  return {
    alerts,
    isLoading,
    error,
    fetchAlerts,
    acknowledgeAlert,
    muteAlert,
  };
};
```

### 5. Reusable Component Examples

**File**: `frontend/src/components/cards/MetricCard.tsx`

```typescript
import React from 'react';
import { ArrowUp, ArrowDown, TrendingUp } from 'lucide-react';

interface MetricCardProps {
  title: string;
  value: number | string;
  unit?: string;
  icon: React.ReactNode;
  trend?: 'up' | 'down' | 'stable';
  trendValue?: string;
  status?: 'healthy' | 'warning' | 'critical';
  onClick?: () => void;
}

const statusColors = {
  healthy: 'bg-pg-success/10 border-pg-success/20',
  warning: 'bg-pg-warning/10 border-pg-warning/20',
  critical: 'bg-pg-danger/10 border-pg-danger/20',
};

const statusTextColors = {
  healthy: 'text-pg-success',
  warning: 'text-pg-warning',
  critical: 'text-pg-danger',
};

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit = '',
  icon,
  trend,
  trendValue,
  status = 'healthy',
  onClick,
}) => {
  return (
    <div
      className={`
        p-6 rounded-lg border-2 cursor-pointer
        transition-all hover:shadow-lg hover:scale-105
        ${statusColors[status]}
      `}
      onClick={onClick}
    >
      <div className="flex justify-between items-start mb-4">
        <div className={`p-2 rounded-lg ${statusTextColors[status]}`}>
          {icon}
        </div>
        {trend && (
          <div className="flex items-center gap-1 text-sm">
            {trend === 'up' && <ArrowUp className="w-4 h-4 text-pg-success" />}
            {trend === 'down' && <ArrowDown className="w-4 h-4 text-pg-danger" />}
            {trend === 'stable' && <TrendingUp className="w-4 h-4 text-pg-slate" />}
            {trendValue && <span className="text-xs text-pg-slate">{trendValue}</span>}
          </div>
        )}
      </div>

      <h3 className="text-sm font-medium text-pg-slate mb-2">{title}</h3>
      <div className="text-3xl font-bold text-pg-dark">
        {value}
        {unit && <span className="text-lg text-pg-slate ml-1">{unit}</span>}
      </div>
    </div>
  );
};
```

**File**: `frontend/src/components/tables/AdvancedDataTable.tsx`

```typescript
import React, { useState } from 'react';
import { ChevronUp, ChevronDown, Search } from 'lucide-react';

interface Column<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
  width?: string;
}

interface AdvancedDataTableProps<T extends { id: string }> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  searchable?: boolean;
  selectable?: boolean;
  onRowClick?: (row: T) => void;
}

export const AdvancedDataTable = <T extends { id: string }>({
  columns,
  data,
  loading = false,
  searchable = true,
  selectable = false,
  onRowClick,
}: AdvancedDataTableProps<T>) => {
  const [sortKey, setSortKey] = useState<keyof T | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());

  // Filter data
  let filteredData = data;
  if (searchTerm) {
    filteredData = data.filter(row =>
      JSON.stringify(row).toLowerCase().includes(searchTerm.toLowerCase())
    );
  }

  // Sort data
  if (sortKey) {
    filteredData = [...filteredData].sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      const comparison = aVal > bVal ? 1 : -1;
      return sortOrder === 'asc' ? comparison : -comparison;
    });
  }

  const toggleSort = (key: keyof T) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };

  const toggleAllRows = () => {
    if (selectedRows.size === filteredData.length) {
      setSelectedRows(new Set());
    } else {
      setSelectedRows(new Set(filteredData.map(row => row.id)));
    }
  };

  const toggleRow = (id: string) => {
    const newSelected = new Set(selectedRows);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedRows(newSelected);
  };

  if (loading) {
    return <div className="p-6 text-center text-pg-slate">Loading...</div>;
  }

  return (
    <div className="space-y-4">
      {searchable && (
        <div className="relative">
          <Search className="absolute left-3 top-3 w-4 h-4 text-pg-slate" />
          <input
            type="text"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-pg-slate/20 rounded-lg focus:outline-none focus:ring-2 focus:ring-pg-cyan"
          />
        </div>
      )}

      <div className="overflow-x-auto border border-pg-slate/10 rounded-lg">
        <table className="w-full text-sm">
          <thead className="bg-pg-slate/5 border-b border-pg-slate/10">
            <tr>
              {selectable && (
                <th className="w-12 px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedRows.size === filteredData.length && filteredData.length > 0}
                    onChange={toggleAllRows}
                  />
                </th>
              )}
              {columns.map(col => (
                <th
                  key={String(col.key)}
                  className="px-4 py-3 text-left font-semibold text-pg-dark cursor-pointer hover:bg-pg-slate/10"
                  style={{ width: col.width }}
                  onClick={() => col.sortable && toggleSort(col.key)}
                >
                  <div className="flex items-center gap-2">
                    {col.label}
                    {col.sortable && sortKey === col.key && (
                      <>
                        {sortOrder === 'asc' ? (
                          <ChevronUp className="w-4 h-4" />
                        ) : (
                          <ChevronDown className="w-4 h-4" />
                        )}
                      </>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filteredData.length === 0 ? (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)} className="px-4 py-8 text-center text-pg-slate">
                  No data found
                </td>
              </tr>
            ) : (
              filteredData.map(row => (
                <tr
                  key={row.id}
                  className="border-b border-pg-slate/5 hover:bg-pg-slate/5 cursor-pointer"
                  onClick={() => onRowClick?.(row)}
                >
                  {selectable && (
                    <td className="w-12 px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedRows.has(row.id)}
                        onChange={() => toggleRow(row.id)}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </td>
                  )}
                  {columns.map(col => (
                    <td key={String(col.key)} className="px-4 py-3 text-pg-dark">
                      {col.render
                        ? col.render(row[col.key], row)
                        : String(row[col.key])}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
```

### 6. Page Example: Alerts & Incidents

**File**: `frontend/src/pages/AlertsIncidents.tsx`

```typescript
import React, { useState } from 'react';
import { AlertTriangle, Bell, Clock, Activity } from 'lucide-react';
import { useAlerts } from '../hooks/useAlerts';
import { useAlertStore } from '../store/alertStore';
import { MetricCard } from '../components/cards/MetricCard';
import { AdvancedDataTable } from '../components/tables/AdvancedDataTable';
import { PageWrapper } from '../components/common/PageWrapper';
import { Alert } from '../types/alerts';

export const AlertsIncidents: React.FC = () => {
  const { alerts, isLoading, acknowledgeAlert, muteAlert } = useAlerts();
  const { filters, setFilters } = useAlertStore();
  const counts = useAlertStore(state => state.alertCounts());

  const handleAcknowledge = async (alertId: string) => {
    await acknowledgeAlert(alertId);
  };

  const handleMute = async (alertId: string) => {
    await muteAlert(alertId);
  };

  const tableColumns = [
    {
      key: 'severity',
      label: 'Severity',
      render: (value: string) => (
        <span className={`
          px-3 py-1 rounded-full text-xs font-semibold
          ${value === 'critical' ? 'bg-pg-danger/20 text-pg-danger' : ''}
          ${value === 'warning' ? 'bg-pg-warning/20 text-pg-warning' : ''}
          ${value === 'info' ? 'bg-pg-cyan/20 text-pg-cyan' : ''}
        `}>
          {value.toUpperCase()}
        </span>
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
      render: (value: string) => (
        <span className={`
          px-3 py-1 rounded-full text-xs font-semibold
          ${value === 'active' ? 'bg-pg-danger/20 text-pg-danger' : 'bg-pg-success/20 text-pg-success'}
        `}>
          {value.toUpperCase()}
        </span>
      ),
    },
    {
      key: 'fired_at',
      label: 'Fired',
      render: (value: Date) => new Date(value).toLocaleTimeString(),
    },
  ];

  return (
    <PageWrapper title="Alerts & Incidents">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <MetricCard
          title="Critical Alerts"
          value={counts.critical}
          icon={<AlertTriangle className="w-6 h-6" />}
          status={counts.critical > 0 ? 'critical' : 'healthy'}
        />
        <MetricCard
          title="Warnings"
          value={counts.warning}
          icon={<AlertTriangle className="w-6 h-6" />}
          status={counts.warning > 0 ? 'warning' : 'healthy'}
        />
        <MetricCard
          title="Info"
          value={counts.info}
          icon={<Bell className="w-6 h-6" />}
          status="healthy"
        />
        <MetricCard
          title="Active"
          value={alerts.filter(a => a.status === 'active').length}
          icon={<Activity className="w-6 h-6" />}
        />
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6 flex gap-4">
        <select
          value={filters.severity || ''}
          onChange={(e) => setFilters({ ...filters, severity: (e.target.value as any) || undefined })}
          className="px-3 py-2 border border-pg-slate/20 rounded-lg text-sm"
        >
          <option value="">All Severities</option>
          <option value="critical">Critical</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
        </select>

        <select
          value={filters.status || ''}
          onChange={(e) => setFilters({ ...filters, status: (e.target.value as any) || undefined })}
          className="px-3 py-2 border border-pg-slate/20 rounded-lg text-sm"
        >
          <option value="">All Statuses</option>
          <option value="active">Active</option>
          <option value="resolved">Resolved</option>
          <option value="muted">Muted</option>
        </select>
      </div>

      {/* Alerts Table */}
      <div className="bg-white rounded-lg shadow">
        <div className="p-6 border-b border-pg-slate/10">
          <h2 className="text-lg font-semibold text-pg-dark">Recent Alerts</h2>
        </div>
        <div className="p-6">
          <AdvancedDataTable
            columns={tableColumns}
            data={alerts}
            loading={isLoading}
            searchable={true}
            onRowClick={(alert: Alert) => {
              // Open alert details modal
              console.log('Alert clicked:', alert);
            }}
          />
        </div>
      </div>
    </PageWrapper>
  );
};
```

---

## Part 3: Setup Instructions

### Step 1: Install Dependencies

```bash
cd frontend

npm install

# Add new libraries
npm install recharts @tanstack/react-table zustand react-query date-fns numeral
npm install framer-motion react-hot-toast --save-dev
npm install @types/numeral --save-dev
```

### Step 2: Create Directory Structure

```bash
mkdir -p src/components/{common,charts,tables,cards,modals}
mkdir -p src/pages
mkdir -p src/hooks
mkdir -p src/store
mkdir -p src/types
mkdir -p src/services
mkdir -p src/utils
mkdir -p src/styles
```

### Step 3: Copy Tailwind Configuration

Update `tailwind.config.js` with the enhanced configuration from Part 2.

### Step 4: Create Base Components

Start with:
1. `src/components/common/Header.tsx`
2. `src/components/common/Sidebar.tsx`
3. `src/components/common/PageWrapper.tsx`
4. `src/components/cards/MetricCard.tsx`
5. `src/components/tables/AdvancedDataTable.tsx`

### Step 5: Create First Page

Implement `src/pages/OverviewDashboard.tsx` as the entry point.

---

## Part 4: API Integration Checklist

Before pages can be fully functional, ensure backend has these endpoints:

### Must-Have for MVP
- [ ] `GET /api/v1/alerts` with filtering
- [ ] `GET /api/v1/collectors/{id}/metrics/timeseries`
- [ ] `GET /api/v1/collectors/{id}/schema`
- [ ] `GET /api/v1/collectors/{id}/locks`
- [ ] `GET /api/v1/collectors/{id}/bloat-analysis`

### Nice-to-Have (Phase 2)
- [ ] `GET /api/v1/incidents`
- [ ] `GET /api/v1/collectors/{id}/health-score`
- [ ] `GET /api/v1/collectors/{id}/query-performance`
- [ ] `POST /api/v1/automation/remediate`

---

## Part 5: Development Workflow

### Run in Development Mode
```bash
npm run dev
# Opens http://localhost:5173
```

### Type Checking
```bash
npm run type-check
```

### Linting
```bash
npm run lint
```

### Testing
```bash
npm run test
npm run test:coverage
```

### Build for Production
```bash
npm run build
```

---

**Ready to begin implementation!**

Next Steps:
1. Install dependencies
2. Create directory structure
3. Implement base components
4. Create first dashboard page
5. Iterate through remaining pages

Generated: March 3, 2026
