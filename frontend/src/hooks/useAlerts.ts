import { useEffect, useState } from 'react';
import { Alert } from '../types/alerts';
import { useAlertStore } from '../store/alertStore';
import { apiClient } from '../services/api';
import { REFRESH_INTERVALS } from '../utils/constants';

export const useAlerts = (autoRefresh = true) => {
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
        `${import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'}/alerts?${params.toString()}`,
        {
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );

      if (!response.ok) throw new Error('Failed to fetch alerts');

      const data = await response.json();
      const alertsData = (data.data || []).map((alert: any) => ({
        ...alert,
        fired_at: new Date(alert.fired_at),
        resolved_at: alert.resolved_at ? new Date(alert.resolved_at) : undefined,
      }));
      setAlerts(alertsData);
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
      const interval = setInterval(fetchAlerts, REFRESH_INTERVALS.alerts);
      return () => clearInterval(interval);
    }
  }, [filters, autoRefresh]);

  const acknowledgeAlert = async (alertId: string) => {
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'}/alerts/${alertId}/acknowledge`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );

      if (!response.ok) throw new Error('Failed to acknowledge alert');
      await fetchAlerts();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to acknowledge alert');
    }
  };

  const muteAlert = async (alertId: string) => {
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'}/alerts/${alertId}/mute`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${apiClient.getToken()}` }
        }
      );

      if (!response.ok) throw new Error('Failed to mute alert');
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
