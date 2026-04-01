// VACUUM Advisor types for frontend

export interface VacuumRecommendation {
  id: number;
  database_id: number;
  table_name: string;
  table_size: number;
  dead_tuples_count: number;
  dead_tuples_ratio: number;
  autovacuum_enabled: boolean;
  autovacuum_naptime?: string;
  last_vacuum?: string;
  last_autovacuum?: string;
  recommendation_type: 'full_vacuum' | 'analyze_only' | 'tune_autovacuum';
  estimated_gain: number;
  created_at: string;
  updated_at: string;
}

export interface AutovacuumConfig {
  id: number;
  database_id: number;
  table_name: string;
  setting_name: string;
  current_value: string;
  recommended_value: string;
  impact: 'high' | 'medium' | 'low';
  created_at: string;
}

export interface VacuumRecommendationsResponse {
  database_id: number;
  recommendations: VacuumRecommendation[];
  count: number;
  limit: number;
}

export interface VacuumTableRecommendationResponse {
  database_id: number;
  table_name: string;
  recommendation: VacuumRecommendation;
  autovacuum_config: AutovacuumConfig[];
}

export interface AutovacuumConfigResponse {
  database_id: number;
  configurations: AutovacuumConfig[];
  total_tables: number;
}

export interface VacuumExecutionResponse {
  status: string;
  executed_at: string;
  tables_affected: number;
}

export interface AutovacuumTuningSuggestion {
  table_name: string;
  parameter: string;
  current_value: string;
  recommended_value: string;
  rationale: string;
  expected_improvement: number;
}

export interface VacuumTuningSuggestionsResponse {
  database_id: number;
  suggestions: AutovacuumTuningSuggestion[];
  estimated_improvement: number;
}

// Summary metrics for dashboard
export interface VacuumAdvisorSummary {
  total_tables: number;
  tables_needing_vacuum: number;
  total_dead_space_bytes: number;
  average_dead_ratio: number;
  autovacuum_disabled_count: number;
  estimated_total_gain: number;
}

// Filter and sorting options
export interface VacuumFilter {
  recommendation_type?: 'full_vacuum' | 'analyze_only' | 'tune_autovacuum';
  min_dead_ratio?: number;
  max_dead_ratio?: number;
  autovacuum_enabled?: boolean;
}

export interface VacuumSort {
  field: 'dead_ratio' | 'estimated_gain' | 'table_size' | 'last_vacuum';
  order: 'asc' | 'desc';
}
