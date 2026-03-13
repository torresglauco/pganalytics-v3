/**
 * Alerts Service
 * Handles API calls for alert silence, escalation policies, and acknowledgment
 */

const API_BASE = '/api/v1';

/**
 * Helper function to make authenticated API calls
 */
async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('access_token') || localStorage.getItem('auth_token');

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
 * Create a silence for an alert rule
 */
export async function createSilence(
  ruleId: string,
  silenceData: {
    duration: number; // minutes
    reason?: string;
    silenceType?: 'alert' | 'rule' | 'all';
  }
): Promise<{
  id: string;
  alert_rule_id: string;
  expires_at: string;
  reason: string;
}> {
  return apiCall(`/alerts/${ruleId}/silence`, {
    method: 'POST',
    body: JSON.stringify({
      duration_minutes: silenceData.duration,
      reason: silenceData.reason || 'Temporarily silenced',
      silence_type: silenceData.silenceType || 'rule',
    }),
  });
}

/**
 * Create a new escalation policy
 */
export async function createEscalationPolicy(
  policy: {
    name: string;
    description?: string;
    steps: Array<{
      step_number: number;
      wait_minutes: number;
      notification_channel: string;
      channel_config: Record<string, string>;
      requires_acknowledgment?: boolean;
    }>;
  }
): Promise<{
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  steps: Array<any>;
  created_at: string;
}> {
  return apiCall('/escalation-policies', {
    method: 'POST',
    body: JSON.stringify(policy),
  });
}

/**
 * Update an existing escalation policy
 */
export async function updateEscalationPolicy(
  policyId: string,
  policy: {
    name?: string;
    description?: string;
    steps?: Array<{
      step_number: number;
      wait_minutes: number;
      notification_channel: string;
      channel_config: Record<string, string>;
      requires_acknowledgment?: boolean;
    }>;
    is_active?: boolean;
  }
): Promise<{
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  steps: Array<any>;
  updated_at: string;
}> {
  return apiCall(`/escalation-policies/${policyId}`, {
    method: 'PUT',
    body: JSON.stringify(policy),
  });
}

/**
 * Acknowledge an alert
 */
export async function acknowledgeAlert(
  triggerOrAlertId: string,
  options?: {
    note?: string;
  }
): Promise<{
  id: string;
  acknowledged: boolean;
  acknowledged_at: string;
  acknowledged_by?: string;
}> {
  return apiCall(`/alerts/${triggerOrAlertId}/acknowledge`, {
    method: 'POST',
    body: JSON.stringify({
      note: options?.note || 'Acknowledged',
    }),
  });
}

/**
 * Get all escalation policies
 */
export async function getEscalationPolicies(options?: {
  active_only?: boolean;
  limit?: number;
  offset?: number;
}): Promise<{
  policies: Array<{
    id: string;
    name: string;
    description?: string;
    is_active: boolean;
    steps: Array<any>;
    created_at: string;
    updated_at: string;
  }>;
  total: number;
}> {
  const params = new URLSearchParams();
  if (options?.active_only) params.append('active_only', 'true');
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/escalation-policies${params.toString() ? '?' + params.toString() : ''}`);
}

/**
 * Get a specific escalation policy
 */
export async function getEscalationPolicy(policyId: string): Promise<{
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  steps: Array<{
    step_number: number;
    wait_minutes: number;
    notification_channel: string;
    channel_config: Record<string, string>;
    requires_acknowledgment?: boolean;
  }>;
  created_at: string;
  updated_at: string;
}> {
  return apiCall(`/escalation-policies/${policyId}`);
}

/**
 * Delete an escalation policy
 */
export async function deleteEscalationPolicy(policyId: string): Promise<void> {
  await apiCall(`/escalation-policies/${policyId}`, {
    method: 'DELETE',
  });
}

/**
 * Link an escalation policy to an alert rule
 */
export async function linkEscalationPolicy(
  ruleId: string,
  policyId: string
): Promise<{
  rule_id: string;
  policy_id: string;
  linked_at: string;
}> {
  return apiCall(`/alert-rules/${ruleId}/escalation-policies`, {
    method: 'POST',
    body: JSON.stringify({ policy_id: policyId }),
  });
}

/**
 * Get silences for an alert rule
 */
export async function getSilences(
  ruleId: string,
  options?: {
    active_only?: boolean;
    limit?: number;
    offset?: number;
  }
): Promise<{
  silences: Array<{
    id: string;
    alert_rule_id: string;
    reason: string;
    expires_at: string;
    created_at: string;
  }>;
  total: number;
}> {
  const params = new URLSearchParams();
  params.append('alert_rule_id', ruleId);
  if (options?.active_only) params.append('active_only', 'true');
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alert-silences?${params.toString()}`);
}

/**
 * Delete (deactivate) a silence
 */
export async function deleteSilence(silenceId: string): Promise<void> {
  await apiCall(`/alert-silences/${silenceId}`, {
    method: 'DELETE',
  });
}

/**
 * Get acknowledgment history for an alert
 */
export async function getAlertAcknowledgments(
  alertId: string,
  options?: {
    limit?: number;
    offset?: number;
  }
): Promise<{
  acknowledgments: Array<{
    id: string;
    alert_id: string;
    acknowledged_by: string;
    acknowledged_at: string;
    note?: string;
  }>;
  total: number;
}> {
  const params = new URLSearchParams();
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/alerts/${alertId}/acknowledgments?${params.toString()}`);
}
