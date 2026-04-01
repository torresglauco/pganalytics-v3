import { useState, useCallback, useEffect } from 'react';
import {
  VacuumRecommendation,
  AutovacuumConfig,
  VacuumRecommendationsResponse,
  VacuumTableRecommendationResponse,
  AutovacuumConfigResponse,
  AutovacuumTuningSuggestion,
  VacuumTuningSuggestionsResponse,
  VacuumFilter,
  VacuumSort,
} from '../types/vacuumAdvisor';

interface UseVacuumAdvisorReturn {
  // Recommendations
  recommendations: VacuumRecommendation[];
  recommendationsLoading: boolean;
  recommendationsError: string | null;
  fetchRecommendations: (databaseId: number, limit?: number) => Promise<void>;

  // Table recommendation
  tableRecommendation: VacuumRecommendation | null;
  tableRecommendationLoading: boolean;
  tableRecommendationError: string | null;
  fetchTableRecommendation: (databaseId: number, tableName: string) => Promise<void>;

  // Autovacuum config
  autovacuumConfigs: AutovacuumConfig[];
  autovacuumConfigLoading: boolean;
  autovacuumConfigError: string | null;
  fetchAutovacuumConfig: (databaseId: number) => Promise<void>;

  // Tuning suggestions
  tuningSuggestions: AutovacuumTuningSuggestion[];
  tuningSuggestionsLoading: boolean;
  tuningSuggestionsError: string | null;
  fetchTuningSuggestions: (databaseId: number) => Promise<void>;

  // Vacuum execution
  executeVacuum: (recommendationId: number) => Promise<void>;
  vacuumExecuting: boolean;
  vacuumExecutionError: string | null;

  // Filtering and sorting
  filter: VacuumFilter;
  setFilter: (filter: VacuumFilter) => void;
  sort: VacuumSort;
  setSort: (sort: VacuumSort) => void;
  filteredAndSortedRecommendations: VacuumRecommendation[];
}

export function useVacuumAdvisor(databaseId?: number): UseVacuumAdvisorReturn {
  // Recommendations state
  const [recommendations, setRecommendations] = useState<VacuumRecommendation[]>([]);
  const [recommendationsLoading, setRecommendationsLoading] = useState(false);
  const [recommendationsError, setRecommendationsError] = useState<string | null>(null);

  // Table recommendation state
  const [tableRecommendation, setTableRecommendation] = useState<VacuumRecommendation | null>(null);
  const [tableRecommendationLoading, setTableRecommendationLoading] = useState(false);
  const [tableRecommendationError, setTableRecommendationError] = useState<string | null>(null);

  // Autovacuum config state
  const [autovacuumConfigs, setAutovacuumConfigs] = useState<AutovacuumConfig[]>([]);
  const [autovacuumConfigLoading, setAutovacuumConfigLoading] = useState(false);
  const [autovacuumConfigError, setAutovacuumConfigError] = useState<string | null>(null);

  // Tuning suggestions state
  const [tuningSuggestions, setTuningSuggestions] = useState<AutovacuumTuningSuggestion[]>([]);
  const [tuningSuggestionsLoading, setTuningSuggestionsLoading] = useState(false);
  const [tuningSuggestionsError, setTuningSuggestionsError] = useState<string | null>(null);

  // Vacuum execution state
  const [vacuumExecuting, setVacuumExecuting] = useState(false);
  const [vacuumExecutionError, setVacuumExecutionError] = useState<string | null>(null);

  // Filter and sort state
  const [filter, setFilter] = useState<VacuumFilter>({});
  const [sort, setSort] = useState<VacuumSort>({ field: 'dead_ratio', order: 'desc' });

  // Fetch recommendations
  const fetchRecommendations = useCallback(
    async (dbId: number, limit: number = 20) => {
      setRecommendationsLoading(true);
      setRecommendationsError(null);

      try {
        const response = await fetch(
          `/api/v1/vacuum-advisor/database/${dbId}/recommendations?limit=${limit}`
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch recommendations: ${response.statusText}`);
        }

        const data: VacuumRecommendationsResponse = await response.json();
        setRecommendations(data.recommendations || []);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        setRecommendationsError(errorMessage);
      } finally {
        setRecommendationsLoading(false);
      }
    },
    []
  );

  // Fetch table recommendation
  const fetchTableRecommendation = useCallback(
    async (dbId: number, tableName: string) => {
      setTableRecommendationLoading(true);
      setTableRecommendationError(null);

      try {
        const response = await fetch(
          `/api/v1/vacuum-advisor/database/${dbId}/table/${tableName}`
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch table recommendation: ${response.statusText}`);
        }

        const data: VacuumTableRecommendationResponse = await response.json();
        setTableRecommendation(data.recommendation);
        setAutovacuumConfigs(data.autovacuum_config || []);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        setTableRecommendationError(errorMessage);
      } finally {
        setTableRecommendationLoading(false);
      }
    },
    []
  );

  // Fetch autovacuum config
  const fetchAutovacuumConfig = useCallback(
    async (dbId: number) => {
      setAutovacuumConfigLoading(true);
      setAutovacuumConfigError(null);

      try {
        const response = await fetch(
          `/api/v1/vacuum-advisor/database/${dbId}/autovacuum-config`
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch autovacuum config: ${response.statusText}`);
        }

        const data: AutovacuumConfigResponse = await response.json();
        setAutovacuumConfigs(data.configurations || []);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        setAutovacuumConfigError(errorMessage);
      } finally {
        setAutovacuumConfigLoading(false);
      }
    },
    []
  );

  // Fetch tuning suggestions
  const fetchTuningSuggestions = useCallback(
    async (dbId: number) => {
      setTuningSuggestionsLoading(true);
      setTuningSuggestionsError(null);

      try {
        const response = await fetch(
          `/api/v1/vacuum-advisor/database/${dbId}/tune-suggestions`
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch tuning suggestions: ${response.statusText}`);
        }

        const data: VacuumTuningSuggestionsResponse = await response.json();
        setTuningSuggestions(data.suggestions || []);
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        setTuningSuggestionsError(errorMessage);
      } finally {
        setTuningSuggestionsLoading(false);
      }
    },
    []
  );

  // Execute vacuum
  const executeVacuum = useCallback(
    async (recommendationId: number) => {
      setVacuumExecuting(true);
      setVacuumExecutionError(null);

      try {
        const response = await fetch(
          `/api/v1/vacuum-advisor/recommendation/${recommendationId}/execute`,
          {
            method: 'POST',
          }
        );

        if (!response.ok) {
          throw new Error(`Failed to execute VACUUM: ${response.statusText}`);
        }

        // Refresh recommendations after execution
        if (databaseId) {
          await fetchRecommendations(databaseId);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        setVacuumExecutionError(errorMessage);
      } finally {
        setVacuumExecuting(false);
      }
    },
    [databaseId, fetchRecommendations]
  );

  // Apply filter and sort to recommendations
  const filteredAndSortedRecommendations = recommendations
    .filter((rec) => {
      if (filter.recommendation_type && rec.recommendation_type !== filter.recommendation_type) {
        return false;
      }
      if (filter.min_dead_ratio && rec.dead_tuples_ratio < filter.min_dead_ratio) {
        return false;
      }
      if (filter.max_dead_ratio && rec.dead_tuples_ratio > filter.max_dead_ratio) {
        return false;
      }
      if (filter.autovacuum_enabled !== undefined && rec.autovacuum_enabled !== filter.autovacuum_enabled) {
        return false;
      }
      return true;
    })
    .sort((a, b) => {
      let aVal: number;
      let bVal: number;

      switch (sort.field) {
        case 'dead_ratio':
          aVal = a.dead_tuples_ratio;
          bVal = b.dead_tuples_ratio;
          break;
        case 'estimated_gain':
          aVal = a.estimated_gain;
          bVal = b.estimated_gain;
          break;
        case 'table_size':
          aVal = a.table_size;
          bVal = b.table_size;
          break;
        case 'last_vacuum':
          aVal = a.last_vacuum ? new Date(a.last_vacuum).getTime() : 0;
          bVal = b.last_vacuum ? new Date(b.last_vacuum).getTime() : 0;
          break;
        default:
          return 0;
      }

      return sort.order === 'asc' ? aVal - bVal : bVal - aVal;
    });

  // Auto-fetch recommendations when databaseId changes
  useEffect(() => {
    if (databaseId) {
      fetchRecommendations(databaseId);
      fetchAutovacuumConfig(databaseId);
      fetchTuningSuggestions(databaseId);
    }
  }, [databaseId, fetchRecommendations, fetchAutovacuumConfig, fetchTuningSuggestions]);

  return {
    recommendations,
    recommendationsLoading,
    recommendationsError,
    fetchRecommendations,

    tableRecommendation,
    tableRecommendationLoading,
    tableRecommendationError,
    fetchTableRecommendation,

    autovacuumConfigs,
    autovacuumConfigLoading,
    autovacuumConfigError,
    fetchAutovacuumConfig,

    tuningSuggestions,
    tuningSuggestionsLoading,
    tuningSuggestionsError,
    fetchTuningSuggestions,

    executeVacuum,
    vacuumExecuting,
    vacuumExecutionError,

    filter,
    setFilter,
    sort,
    setSort,
    filteredAndSortedRecommendations,
  };
}
