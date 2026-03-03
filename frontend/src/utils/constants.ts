/**
 * Application constants and configuration
 */

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export const COLORS = {
  primary: '#1e3a8a',      // Deep Blue
  accent: '#06b6d4',       // Cyan
  success: '#10b981',      // Emerald
  warning: '#f59e0b',      // Amber
  danger: '#f43f5e',       // Rose
  neutral: '#64748b',      // Slate
  dark: '#0f172a',         // Dark Blue
  light: '#f8fafc',        // Light Gray
};

export const ALERT_SEVERITY_LEVELS = {
  critical: { color: COLORS.danger, label: 'Critical', priority: 0 },
  warning: { color: COLORS.warning, label: 'Warning', priority: 1 },
  info: { color: COLORS.accent, label: 'Info', priority: 2 },
};

export const ALERT_STATUS_LABELS = {
  active: 'Active',
  resolved: 'Resolved',
  muted: 'Muted',
};

export const HEALTH_STATUS = {
  healthy: { color: COLORS.success, label: 'Healthy', score: '80+' },
  warning: { color: COLORS.warning, label: 'Warning', score: '60-79' },
  critical: { color: COLORS.danger, label: 'Critical', score: '0-59' },
};

export const TIME_RANGES = [
  { label: '1 hour', value: '1h' as const },
  { label: '24 hours', value: '24h' as const },
  { label: '7 days', value: '7d' as const },
  { label: '30 days', value: '30d' as const },
];

export const PAGES = {
  OVERVIEW: 'overview',
  ALERTS: 'alerts',
  QUERIES: 'queries',
  LOCKS: 'locks',
  BLOAT: 'bloat',
  CONNECTIONS: 'connections',
  CACHE: 'cache',
  SCHEMA: 'schema',
  REPLICATION: 'replication',
  HEALTH: 'health',
  EXTENSIONS: 'extensions',
  COLLECTORS: 'collectors',
  SETTINGS: 'settings',
};

export const SIDEBAR_ITEMS = [
  { id: PAGES.OVERVIEW, label: 'Overview', icon: '📊' },
  { id: PAGES.ALERTS, label: 'Alerts & Incidents', icon: '🚨' },
  { id: PAGES.QUERIES, label: 'Query Performance', icon: '⚡' },
  { id: PAGES.LOCKS, label: 'Lock Contention', icon: '🔒' },
  { id: PAGES.BLOAT, label: 'Table Bloat', icon: '🧹' },
  { id: PAGES.CONNECTIONS, label: 'Connections', icon: '📡' },
  { id: PAGES.CACHE, label: 'Cache Performance', icon: '💾' },
  { id: PAGES.SCHEMA, label: 'Schema Explorer', icon: '📐' },
  { id: PAGES.REPLICATION, label: 'Replication', icon: '🔄' },
  { id: PAGES.HEALTH, label: 'Database Health', icon: '💪' },
  { id: PAGES.EXTENSIONS, label: 'Extensions & Config', icon: '⚙️' },
  { id: PAGES.COLLECTORS, label: 'Collectors', icon: '🖥️' },
  { id: PAGES.SETTINGS, label: 'Settings', icon: '⚙️' },
];

export const REFRESH_INTERVALS = {
  alerts: 30000,        // 30 seconds
  metrics: 60000,       // 1 minute
  health: 120000,       // 2 minutes
  status: 5000,         // 5 seconds for critical
};
