/**
 * Alert Dashboard API Client
 * Handles all API calls for alert dashboard operations
 */

import type {
  AlertIncident,
  AlertListResponse,
  AlertStats,
  AlertFilters,
  AlertEvent,
  AlertAction,
  BulkAlertActionResult,
  CorrelationSuggestion,
  AlertSuggestion,
  SLAMetrics,
} from '../types/alertDashboard';

const API_BASE = '/api/v1';

/**
 * Helper function to make authenticated API calls
 */
async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('access_token');

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `API error: ${response.status}`);
  }

  return response.json();
}

/**
 * List alerts with filtering and pagination
 */
export async function listAlerts(options?: {
  filters?: AlertFilters;
  limit?: number;
  offset?: number;
  sort_by?: 'fired_at' | 'severity' | 'status';
  sort_order?: 'asc' | 'desc';
}): Promise<AlertListResponse> {
  const params = new URLSearchParams();

  if (options?.filters) {
    if (options.filters.status)
      params.append('status', options.filters.status.join(','));
    if (options.filters.severity)
      params.append('severity', options.filters.severity.join(','));
    if (options.filters.rule_id) params.append('rule_id', options.filters.rule_id);
    if (options.filters.database_id)
      params.append('database_id', options.filters.database_id);
    if (options.filters.source_type)
      params.append('source_type', options.filters.source_type.join(','));
    if (options.filters.tags)
      params.append('tags', options.filters.tags.join(','));
    if (options.filters.team) params.append('team', options.filters.team);
    if (options.filters.environment)
      params.append('environment', options.filters.environment);
    if (options.filters.search_term)
      params.append('search', options.filters.search_term);
    if (options.filters.date_range) {
      params.append('date_start', options.filters.date_range.start);
      params.append('date_end', options.filters.date_range.end);
    }
  }

  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());
  if (options?.sort_by) params.append('sort_by', options.sort_by);
  if (options?.sort_order) params.append('sort_order', options.sort_order);

  return apiCall(`/alerts?${params.toString()}`);
}

/**
 * Get a specific alert
 */
export async function getAlert(alertId: string): Promise<AlertIncident> {
  return apiCall(`/alerts/${alertId}`);
}

/**
 * Get alert statistics
 */
export async function getAlertStats(): Promise<AlertStats> {
  return apiCall('/alerts/stats');
}

/**
 * Get alert timeline/events
 */
export async function getAlertEvents(
  alertId: string,
  options?: {
    limit?: number;
    offset?: number;
  }
): Promise<{
  events: AlertEvent[];
  total: number;
}> {
  const params = new URLSearchParams();
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alerts/${alertId}/events?${params.toString()}`);
}

/**
 * Acknowledge alerts
 */
export async function acknowledgeAlerts(
  alertIds: string[],
  notes?: string
): Promise<BulkAlertActionResult> {
  return apiCall('/alerts/acknowledge', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
      notes,
    }),
  });
}

/**
 * Resolve alerts
 */
export async function resolveAlerts(
  alertIds: string[],
  notes?: string
): Promise<BulkAlertActionResult> {
  return apiCall('/alerts/resolve', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
      notes,
    }),
  });
}

/**
 * Reopen alerts
 */
export async function reopenAlerts(alertIds: string[]): Promise<BulkAlertActionResult> {
  return apiCall('/alerts/reopen', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
    }),
  });
}

/**
 * Escalate alerts
 */
export async function escalateAlerts(
  alertIds: string[],
  escalation_level: number
): Promise<BulkAlertActionResult> {
  return apiCall('/alerts/escalate', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
      escalation_level,
    }),
  });
}

/**
 * Snooze alerts
 */
export async function snoozeAlerts(
  alertIds: string[],
  minutes: number
): Promise<BulkAlertActionResult> {
  return apiCall('/alerts/snooze', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
      snooze_duration_minutes: minutes,
    }),
  });
}

/**
 * Get correlation suggestions for alerts
 */
export async function getCorrelationSuggestions(
  alertIds: string[]
): Promise<CorrelationSuggestion[]> {
  return apiCall('/alerts/correlations', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids: alertIds,
    }),
  });
}

/**
 * Get alert suggestions and insights
 */
export async function getAlertSuggestions(): Promise<AlertSuggestion[]> {
  return apiCall('/alerts/suggestions');
}

/**
 * Get SLA metrics
 */
export async function getSLAMetrics(): Promise<SLAMetrics[]> {
  return apiCall('/alerts/sla-metrics');
}

/**
 * Export alerts
 */
export async function exportAlerts(
  filters: AlertFilters,
  format: 'csv' | 'json' = 'csv'
): Promise<Blob> {
  const params = new URLSearchParams();
  params.append('format', format);

  if (filters.status) params.append('status', filters.status.join(','));
  if (filters.severity) params.append('severity', filters.severity.join(','));
  if (filters.rule_id) params.append('rule_id', filters.rule_id);
  if (filters.database_id) params.append('database_id', filters.database_id);

  const response = await fetch(`${API_BASE}/alerts/export?${params.toString()}`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('access_token')}`,
    },
  });

  if (!response.ok) {
    throw new Error('Export failed');
  }

  return response.blob();
}

/**
 * Get time series data for a specific metric
 */
export async function getMetricTimeSeries(
  metric_name: string,
  options?: {
    start_time?: string;
    end_time?: string;
    interval?: string; // '1m', '5m', '1h', etc
  }
): Promise<{
  data: Array<{
    timestamp: string;
    value: number;
  }>;
  metric_name: string;
}> {
  const params = new URLSearchParams();
  params.append('metric', metric_name);
  if (options?.start_time) params.append('start_time', options.start_time);
  if (options?.end_time) params.append('end_time', options.end_time);
  if (options?.interval) params.append('interval', options.interval);

  return apiCall(`/alerts/metrics/timeseries?${params.toString()}`);
}

/**
 * Get database health status
 */
export async function getDatabaseHealth(databaseId: string): Promise<{
  database_id: string;
  status: 'healthy' | 'warning' | 'critical';
  last_update: string;
  active_alerts: number;
  metrics: Record<string, number>;
}> {
  return apiCall(`/alerts/databases/${databaseId}/health`);
}

/**
 * Subscribe to real-time alert updates (WebSocket)
 */
export function subscribeToAlertUpdates(
  onMessage: (data: any) => void,
  onError: (error: Error) => void
): WebSocket {
  const token = localStorage.getItem('access_token');
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/api/v1/alerts/stream?token=${token}`;

  const ws = new WebSocket(wsUrl);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (err) {
      onError(new Error('Failed to parse WebSocket message'));
    }
  };

  ws.onerror = (event) => {
    onError(new Error('WebSocket error'));
  };

  return ws;
}

/**
 * Get alert grouping recommendations
 */
export async function getGroupingRecommendations(): Promise<{
  recommendations: Array<{
    alert_ids: string[];
    group_name: string;
    confidence: number;
    reason: string;
  }>;
}> {
  return apiCall('/alerts/grouping-recommendations');
}

/**
 * Create alert group
 */
export async function createAlertGroup(
  alert_ids: string[],
  group_name: string
): Promise<{
  group_id: string;
  name: string;
  alert_ids: string[];
}> {
  return apiCall('/alerts/groups', {
    method: 'POST',
    body: JSON.stringify({
      alert_ids,
      name: group_name,
    }),
  });
}

/**
 * Get related alerts
 */
export async function getRelatedAlerts(alertId: string): Promise<AlertIncident[]> {
  return apiCall(`/alerts/${alertId}/related`);
}
