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
  addAlert: (alert: Alert) => void;
  removeAlert: (id: string) => void;
  updateAlert: (id: string, alert: Partial<Alert>) => void;

  // Computed
  filteredAlerts: () => Alert[];
  alertCounts: () => { critical: number; warning: number; info: number };
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

  addAlert: (alert) => set((state) => ({
    alerts: [alert, ...state.alerts],
  })),

  removeAlert: (id) => set((state) => ({
    alerts: state.alerts.filter((a) => a.id !== id),
  })),

  updateAlert: (id, updates) => set((state) => ({
    alerts: state.alerts.map((a) => (a.id === id ? { ...a, ...updates } : a)),
  })),

  filteredAlerts: () => {
    const { alerts, filters } = get();
    return alerts.filter((alert) => {
      if (filters.severity && alert.severity !== filters.severity) return false;
      if (filters.status && alert.status !== filters.status) return false;
      if (filters.collectorId && alert.collector_id !== filters.collectorId) return false;
      return true;
    });
  },

  alertCounts: () => {
    const filtered = get().filteredAlerts();
    return {
      critical: filtered.filter((a) => a.severity === 'critical').length,
      warning: filtered.filter((a) => a.severity === 'warning').length,
      info: filtered.filter((a) => a.severity === 'info').length,
    };
  },
}));
