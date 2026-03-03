import { useState, useEffect } from 'react';
import { HealthScore } from '../types/metrics';
import { calculateOverallHealth } from '../utils/healthCalculations';

interface HealthScoreState {
  data: HealthScore | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export const useHealthScore = (collectorId: string | null): HealthScoreState => {
  const [data, setData] = useState<HealthScore | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHealthScore = async () => {
    if (!collectorId) {
      setData(null);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // TODO: Replace with actual API call
      // const response = await fetch(`/api/v1/collectors/${collectorId}/health-score`, {
      //   headers: { 'Authorization': `Bearer ${token}` }
      // });
      // const result = await response.json();

      // Mock data for development
      const mockData: HealthScore = {
        overall: 85,
        lock_health: 80,
        bloat_health: 45,
        query_health: 90,
        cache_health: 60,
        connection_health: 85,
        replication_health: 92,
        timestamp: new Date(),
      };

      setData(mockData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch health score');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealthScore();
    const interval = setInterval(fetchHealthScore, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, [collectorId]);

  return {
    data,
    loading,
    error,
    refetch: fetchHealthScore,
  };
};
