/**
 * Notifications API Client
 * Handles all API calls for notification channel management
 */

import type {
  NotificationChannel,
  CreateChannelRequest,
  UpdateChannelRequest,
  TestChannelResult,
  NotificationDelivery,
  DeliveryStats,
  BulkChannelAction,
} from '../types/notifications';

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
 * List all notification channels
 */
export async function listNotificationChannels(options?: {
  type?: string;
  status?: string;
  search?: string;
  limit?: number;
  offset?: number;
}): Promise<{
  channels: NotificationChannel[];
  total: number;
  limit: number;
  offset: number;
}> {
  const params = new URLSearchParams();
  if (options?.type) params.append('type', options.type);
  if (options?.status) params.append('status', options.status);
  if (options?.search) params.append('search', options.search);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(`/notification-channels?${params.toString()}`);
}

/**
 * Get a specific notification channel
 */
export async function getNotificationChannel(
  channelId: string
): Promise<NotificationChannel> {
  return apiCall(`/notification-channels/${channelId}`);
}

/**
 * Create a new notification channel
 */
export async function createNotificationChannel(
  request: CreateChannelRequest
): Promise<NotificationChannel> {
  return apiCall('/notification-channels', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Update a notification channel
 */
export async function updateNotificationChannel(
  request: UpdateChannelRequest
): Promise<NotificationChannel> {
  const { id, ...data } = request;
  return apiCall(`/notification-channels/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/**
 * Delete a notification channel
 */
export async function deleteNotificationChannel(channelId: string): Promise<void> {
  await apiCall(`/notification-channels/${channelId}`, {
    method: 'DELETE',
  });
}

/**
 * Test a notification channel
 */
export async function testNotificationChannel(
  channelId: string
): Promise<TestChannelResult> {
  return apiCall(`/notification-channels/${channelId}/test`, {
    method: 'POST',
  });
}

/**
 * Enable/disable a notification channel
 */
export async function toggleNotificationChannel(
  channelId: string,
  enabled: boolean
): Promise<NotificationChannel> {
  return apiCall(`/notification-channels/${channelId}/toggle`, {
    method: 'POST',
    body: JSON.stringify({ enabled }),
  });
}

/**
 * Set a channel as default
 */
export async function setDefaultChannel(channelId: string): Promise<NotificationChannel> {
  return apiCall(`/notification-channels/${channelId}/set-default`, {
    method: 'POST',
  });
}

/**
 * Get delivery statistics for a channel
 */
export async function getChannelDeliveryStats(
  channelId: string
): Promise<DeliveryStats> {
  return apiCall(`/notification-channels/${channelId}/stats`);
}

/**
 * Get delivery history for a channel
 */
export async function getChannelDeliveryHistory(
  channelId: string,
  options?: {
    status?: string;
    limit?: number;
    offset?: number;
  }
): Promise<{
  deliveries: NotificationDelivery[];
  total: number;
  limit: number;
  offset: number;
}> {
  const params = new URLSearchParams();
  if (options?.status) params.append('status', options.status);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  return apiCall(
    `/notification-channels/${channelId}/deliveries?${params.toString()}`
  );
}

/**
 * Get delivery history for an alert
 */
export async function getAlertDeliveryHistory(
  alertId: string
): Promise<NotificationDelivery[]> {
  const result = await apiCall<{
    deliveries: NotificationDelivery[];
  }>(`/alerts/${alertId}/delivery-history`);
  return result.deliveries;
}

/**
 * Retry failed delivery
 */
export async function retryDelivery(deliveryId: string): Promise<TestChannelResult> {
  return apiCall(`/notification-deliveries/${deliveryId}/retry`, {
    method: 'POST',
  });
}

/**
 * Perform bulk channel actions
 */
export async function bulkChannelAction(action: BulkChannelAction): Promise<{
  succeeded: number;
  failed: number;
  errors?: Array<{
    channel_id: string;
    error: string;
  }>;
}> {
  return apiCall('/notification-channels/bulk-action', {
    method: 'POST',
    body: JSON.stringify(action),
  });
}

/**
 * Get available templates for a channel type
 */
export async function getChannelTemplates(type: string): Promise<{
  templates: Array<{
    name: string;
    preview: string;
  }>;
}> {
  return apiCall(`/notification-channels/templates/${type}`);
}

/**
 * Validate channel configuration
 */
export async function validateChannelConfig(
  request: CreateChannelRequest
): Promise<{
  valid: boolean;
  errors?: Array<{
    field: string;
    message: string;
  }>;
}> {
  return apiCall('/notification-channels/validate', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Export channels configuration
 */
export async function exportChannels(): Promise<Blob> {
  const response = await fetch(`${API_BASE}/notification-channels/export`, {
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
 * Import channels configuration
 */
export async function importChannels(file: File): Promise<{
  imported: number;
  skipped: number;
  errors: Array<{
    line: number;
    error: string;
  }>;
}> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`${API_BASE}/notification-channels/import`, {
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
