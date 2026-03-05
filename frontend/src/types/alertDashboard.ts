/**
 * Alert Dashboard Type Definitions
 * Types for alert dashboard, metrics, and real-time updates
 */

/**
 * Alert status
 */
export type AlertStatus = 'firing' | 'resolved' | 'acknowledged';

/**
 * Alert source types
 */
export type AlertSourceType = 'rule' | 'anomaly' | 'manual' | 'integration';

/**
 * Alert incident
 */
export interface AlertIncident {
  id: string;
  rule_id: string;
  database_id: string;
  alert_rule_name: string;
  status: AlertStatus;
  severity: 'low' | 'medium' | 'high' | 'critical';
  source_type: AlertSourceType;

  // Content
  title: string;
  description?: string;
  metric_name?: string;
  metric_value?: number;
  threshold_value?: number;

  // Timing
  fired_at: string;
  acknowledged_at?: string;
  acknowledged_by?: string;
  resolved_at?: string;
  resolved_by?: string;

  // Metadata
  tags?: string[];
  runbook_url?: string;
  related_alerts?: string[]; // IDs of related alerts
  fire_count?: number; // Number of times this alert has fired

  // Environment
  environment?: string;
  service?: string;
  team?: string;
}

/**
 * Alert metric for aggregation
 */
export interface AlertMetric {
  metric_id: string;
  name: string;
  value: number;
  unit?: string;
  timestamp: string;
  tags?: Record<string, string>;
}

/**
 * Dashboard KPI
 */
export interface DashboardKPI {
  label: string;
  value: number | string;
  unit?: string;
  trend?: 'up' | 'down' | 'stable';
  trend_percent?: number;
  color?: 'green' | 'yellow' | 'red';
}

/**
 * Alert statistics
 */
export interface AlertStats {
  total_alerts: number;
  firing: number;
  acknowledged: number;
  resolved: number;
  by_severity: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  by_source: Record<AlertSourceType, number>;
  avg_time_to_resolve_minutes?: number;
  alert_rate_per_hour?: number;
}

/**
 * Time series data point
 */
export interface TimeSeriesPoint {
  timestamp: string;
  value: number;
  label?: string;
}

/**
 * Alert event for timeline
 */
export interface AlertEvent {
  id: string;
  timestamp: string;
  event_type: 'fired' | 'acknowledged' | 'resolved' | 'escalated';
  actor?: string;
  message: string;
  metadata?: Record<string, any>;
}

/**
 * Alert group/incident
 */
export interface AlertGroup {
  id: string;
  name: string;
  description?: string;
  related_alert_ids: string[];
  created_at: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: AlertStatus;
  fire_count: number;
  last_update: string;
}

/**
 * Filter options for alerts
 */
export interface AlertFilters {
  status?: AlertStatus[];
  severity?: Array<'low' | 'medium' | 'high' | 'critical'>;
  rule_id?: string;
  database_id?: string;
  source_type?: AlertSourceType[];
  tags?: string[];
  date_range?: {
    start: string;
    end: string;
  };
  search_term?: string;
  team?: string;
  environment?: string;
}

/**
 * Alert list response
 */
export interface AlertListResponse {
  alerts: AlertIncident[];
  total: number;
  limit: number;
  offset: number;
  filters_applied: AlertFilters;
}

/**
 * Real-time alert update message
 */
export interface AlertUpdateMessage {
  type: 'alert_fired' | 'alert_acknowledged' | 'alert_resolved' | 'stats_updated';
  alert?: AlertIncident;
  stats?: AlertStats;
  timestamp: string;
}

/**
 * Correlation suggestion
 */
export interface CorrelationSuggestion {
  alert_ids: string[];
  confidence: number; // 0-1
  reason: string;
  suggested_group_name?: string;
}

/**
 * Alert action (acknowledge, resolve, etc)
 */
export interface AlertAction {
  alert_ids: string[];
  action: 'acknowledge' | 'resolve' | 'reopen' | 'escalate' | 'snooze';
  notes?: string;
  snooze_minutes?: number; // For snooze action
}

/**
 * Bulk alert action result
 */
export interface BulkAlertActionResult {
  succeeded: number;
  failed: number;
  errors?: Array<{
    alert_id: string;
    error: string;
  }>;
}

/**
 * Alert suggestion
 */
export interface AlertSuggestion {
  type: 'correlation' | 'action' | 'insight';
  title: string;
  description: string;
  recommended_action?: string;
  confidence: number;
}

/**
 * SLA metrics
 */
export interface SLAMetrics {
  sla_name: string;
  total_incidents: number;
  compliant_incidents: number;
  compliance_percentage: number;
  avg_time_to_acknowledge_minutes: number;
  avg_time_to_resolve_minutes: number;
  target_acknowledge_minutes: number;
  target_resolve_minutes: number;
}
