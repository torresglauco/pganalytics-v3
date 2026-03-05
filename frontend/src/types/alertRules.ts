/**
 * Alert Rules Type Definitions
 * Comprehensive types for alert rule creation, management, and execution
 */

/**
 * Rule Condition Types
 */
export type ConditionType = 'threshold' | 'anomaly' | 'change' | 'composite' | 'query';

/**
 * Comparison operators
 */
export type ComparisonOperator = '>' | '<' | '>=' | '<=' | '=' | '!=' | 'matches' | 'contains';

/**
 * Aggregation functions
 */
export type AggregationType = 'avg' | 'sum' | 'min' | 'max' | 'count' | 'percentile' | 'stddev';

/**
 * Composite condition operators
 */
export type CompositeOperator = 'AND' | 'OR';

/**
 * Alert severity levels
 */
export type AlertSeverity = 'low' | 'medium' | 'high' | 'critical';

/**
 * Rule enabled/disabled state
 */
export type RuleStatus = 'enabled' | 'disabled' | 'testing' | 'archived';

/**
 * Threshold condition - trigger when metric crosses threshold
 */
export interface ThresholdCondition {
  type: 'threshold';
  metric_name: string;
  operator: ComparisonOperator;
  threshold_value: number;
  aggregation?: AggregationType;
  window_seconds?: number; // Evaluation window (default 300)
  count_threshold?: number; // Trigger after N consecutive evaluations
}

/**
 * Anomaly condition - trigger when metric deviates from baseline
 */
export interface AnomalyCondition {
  type: 'anomaly';
  metric_name: string;
  sensitivity: 'low' | 'medium' | 'high'; // Z-score thresholds
  baseline_days?: number; // Days to calculate baseline (default 7)
  min_deviation_percent?: number; // Minimum % deviation to trigger
}

/**
 * Change condition - trigger when metric changes significantly
 */
export interface ChangeCondition {
  type: 'change';
  metric_name: string;
  change_type: 'increase' | 'decrease' | 'both';
  change_percent: number; // % change to trigger
  compare_to: 'previous' | '1h_ago' | '1d_ago'; // What to compare against
}

/**
 * Composite condition - combine multiple conditions with AND/OR
 */
export interface CompositeCondition {
  type: 'composite';
  operator: CompositeOperator;
  conditions: RuleCondition[];
}

/**
 * Custom query condition - trigger based on SQL query result
 */
export interface QueryCondition {
  type: 'query';
  query: string; // SQL query
  operator: ComparisonOperator;
  expected_value: number | string;
  query_timeout_seconds?: number;
}

/**
 * Union type of all condition types
 */
export type RuleCondition =
  | ThresholdCondition
  | AnomalyCondition
  | ChangeCondition
  | CompositeCondition
  | QueryCondition;

/**
 * Notification configuration for rules
 */
export interface NotificationConfig {
  channel_id: string;
  channel_type: 'slack' | 'email' | 'webhook' | 'pagerduty' | 'jira';
  notify_on: 'all' | 'initial' | 'escalation'; // When to notify
  include_context?: boolean; // Include metric context in notification
}

/**
 * Alert rule definition
 */
export interface AlertRule {
  id: string;
  name: string;
  description?: string;
  database_id: string; // Target database
  status: RuleStatus;
  severity: AlertSeverity;
  condition: RuleCondition;
  notifications: NotificationConfig[];

  // Timing
  evaluation_interval_seconds?: number; // How often to evaluate (default 300)
  for_duration_seconds?: number; // Must be true for N seconds before firing (default 60)

  // Behavior
  resolve_behavior?: 'auto' | 'manual'; // Auto-resolve when condition clears
  resolve_timeout_minutes?: number; // Auto-resolve after N minutes
  dedup_window_minutes?: number; // Dedup notifications for N minutes (default 5)

  // Metadata
  created_at: string;
  updated_at: string;
  created_by: string;
  last_fired_at?: string;
  fire_count?: number;

  // Tags for organization
  tags?: string[];
  team?: string;
  runbook_url?: string;
}

/**
 * Create rule request
 */
export interface CreateRuleRequest {
  name: string;
  description?: string;
  database_id: string;
  severity: AlertSeverity;
  condition: RuleCondition;
  notifications: NotificationConfig[];
  evaluation_interval_seconds?: number;
  for_duration_seconds?: number;
  resolve_behavior?: 'auto' | 'manual';
  tags?: string[];
  team?: string;
  runbook_url?: string;
}

/**
 * Update rule request
 */
export interface UpdateRuleRequest extends Partial<CreateRuleRequest> {
  id: string;
}

/**
 * Rule test response
 */
export interface RuleTestResult {
  rule_id: string;
  condition_met: boolean;
  metric_value: number;
  threshold_value?: number;
  evaluation_time_ms: number;
  last_data_point: {
    timestamp: string;
    value: number;
  };
  sample_metrics?: Array<{
    timestamp: string;
    value: number;
  }>;
}

/**
 * Rule template for quick creation
 */
export interface RuleTemplate {
  id: string;
  name: string;
  description: string;
  category: 'performance' | 'security' | 'availability' | 'custom';
  condition: RuleCondition;
  default_severity: AlertSeverity;
  icon?: string;
  tags: string[];
}

/**
 * Rule validation result
 */
export interface RuleValidationResult {
  valid: boolean;
  errors: Array<{
    field: string;
    message: string;
  }>;
  warnings?: Array<{
    field: string;
    message: string;
  }>;
}

/**
 * Rule execution statistics
 */
export interface RuleStats {
  rule_id: string;
  last_evaluation: string;
  evaluation_count: number;
  fire_count: number;
  avg_evaluation_time_ms: number;
  last_error?: string;
}

/**
 * Bulk rule action request
 */
export interface BulkRuleAction {
  action: 'enable' | 'disable' | 'delete' | 'update_severity';
  rule_ids: string[];
  new_value?: AlertSeverity | boolean; // For severity update or enable/disable
}

/**
 * Rule import/export format
 */
export interface RuleExport {
  version: string;
  export_date: string;
  rules: AlertRule[];
}
