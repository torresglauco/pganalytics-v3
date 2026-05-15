/**
 * Alert Rules API Client
 * Handles all API calls for alert rule management
 */

import type {
  AlertRule,
  CreateRuleRequest,
  UpdateRuleRequest,
  RuleTestResult,
  RuleValidationResult,
  RuleStats,
  BulkRuleAction,
} from '../types/alertRules';

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
 * List all alert rules for a database
 */
export async function listAlertRules(
  databaseId: string,
  options?: {
    status?: string;
    severity?: string;
    tag?: string;
    search?: string;
    limit?: number;
    offset?: number;
  }
): Promise<{
  rules: AlertRule[];
  total: number;
  limit: number;
  offset: number;
}> {
  const params = new URLSearchParams();
  params.append('database_id', databaseId);
  if (options?.status) params.append('status', options.status);
  if (options?.severity) params.append('severity', options.severity);
  if (options?.tag) params.append('tag', options.tag);
  if (options?.search) params.append('search', options.search);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alert-rules?${params.toString()}`);
}

/**
 * Get a specific alert rule
 */
export async function getAlertRule(ruleId: string): Promise<AlertRule> {
  return apiCall(`/alert-rules/${ruleId}`);
}

/**
 * Create a new alert rule
 */
export async function createAlertRule(
  request: CreateRuleRequest
): Promise<AlertRule> {
  return apiCall('/alert-rules', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Update an existing alert rule
 */
export async function updateAlertRule(request: UpdateRuleRequest): Promise<AlertRule> {
  const { id, ...data } = request;
  return apiCall(`/alert-rules/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/**
 * Delete an alert rule
 */
export async function deleteAlertRule(ruleId: string): Promise<void> {
  await apiCall(`/alert-rules/${ruleId}`, {
    method: 'DELETE',
  });
}

/**
 * Enable/disable an alert rule
 */
export async function toggleAlertRule(
  ruleId: string,
  enabled: boolean
): Promise<AlertRule> {
  return apiCall(`/alert-rules/${ruleId}/toggle`, {
    method: 'POST',
    body: JSON.stringify({ enabled }),
  });
}

/**
 * Test an alert rule condition
 */
export async function testAlertRule(
  ruleId: string | null,
  request: CreateRuleRequest
): Promise<RuleTestResult> {
  const endpoint = ruleId ? `/alert-rules/${ruleId}/test` : '/alert-rules/test';
  return apiCall(endpoint, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Validate an alert rule before saving
 */
export async function validateAlertRule(
  request: CreateRuleRequest
): Promise<RuleValidationResult> {
  return apiCall('/alert-rules/validate', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Get rule execution statistics
 */
export async function getRuleStats(ruleId: string): Promise<RuleStats> {
  return apiCall(`/alert-rules/${ruleId}/stats`);
}

/**
 * Get rule execution history
 */
export async function getRuleHistory(
  ruleId: string,
  options?: {
    limit?: number;
    offset?: number;
  }
): Promise<{
  events: Array<{
    timestamp: string;
    event_type: 'fired' | 'resolved' | 'triggered';
    metric_value: number;
    condition_met: boolean;
  }>;
  total: number;
}> {
  const params = new URLSearchParams();
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alert-rules/${ruleId}/history?${params.toString()}`);
}

/**
 * Perform bulk actions on rules
 */
export async function bulkRuleAction(action: BulkRuleAction): Promise<{
  succeeded: number;
  failed: number;
  errors?: Array<{
    rule_id: string;
    error: string;
  }>;
}> {
  return apiCall('/alert-rules/bulk-action', {
    method: 'POST',
    body: JSON.stringify(action),
  });
}

/**
 * Export rules as JSON or CSV
 */
export async function exportRules(
  databaseId: string,
  format: 'json' | 'csv' = 'json'
): Promise<Blob> {
  const response = await fetch(
    `${API_BASE}/alert-rules/export?database_id=${databaseId}&format=${format}`,
    {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('access_token')}`,
      },
    }
  );

  if (!response.ok) {
    throw new Error('Export failed');
  }

  return response.blob();
}

/**
 * Import rules from JSON or CSV
 */
export async function importRules(
  databaseId: string,
  file: File
): Promise<{
  imported: number;
  skipped: number;
  errors: Array<{
    line: number;
    error: string;
  }>;
}> {
  const formData = new FormData();
  formData.append('database_id', databaseId);
  formData.append('file', file);

  const response = await fetch(`${API_BASE}/alert-rules/import`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('access_token')}`,
    },
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || 'Import failed');
  }

  return response.json();
}

/**
 * Clone an existing rule
 */
export async function cloneAlertRule(
  ruleId: string,
  overrides?: Partial<CreateRuleRequest>
): Promise<AlertRule> {
  return apiCall(`/alert-rules/${ruleId}/clone`, {
    method: 'POST',
    body: JSON.stringify(overrides || {}),
  });
}

/**
 * Get suggested rules based on database metrics
 */
export async function getSuggestedRules(databaseId: string): Promise<{
  suggestions: Array<{
    name: string;
    description: string;
    condition: any;
    severity: string;
    reason: string;
  }>;
}> {
  return apiCall(`/alert-rules/suggestions?database_id=${databaseId}`);
}

/**
 * Get alert history across all rules
 */
export async function getAlertHistory(options?: {
  database_id?: string;
  rule_id?: string;
  severity?: string;
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<{
  triggers: Array<{
    id: string;
    rule_id: string;
    rule_name: string;
    triggered_at: string;
    severity: string;
    status: string;
    metric_value: number;
    acknowledged_at?: string;
    acknowledged_by?: string;
  }>;
  total: number;
}> {
  const params = new URLSearchParams();
  if (options?.database_id) params.append('database_id', options.database_id);
  if (options?.rule_id) params.append('rule_id', options.rule_id);
  if (options?.severity) params.append('severity', options.severity);
  if (options?.status) params.append('status', options.status);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alerts/history?${params.toString()}`);
}

/**
 * Acknowledge an alert trigger
 */
export async function acknowledgeAlert(triggerId: string): Promise<{
  id: string;
  status: string;
  acknowledged_at: string;
}> {
  return apiCall(`/alerts/${triggerId}/acknowledge`, {
    method: 'POST',
  });
}
