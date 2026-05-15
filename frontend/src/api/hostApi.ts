/**
 * Host Monitoring API Client
 * Handles all API calls for host monitoring endpoints
 */

import type {
  HostStatus,
  HostStatusResponse,
  HostMetrics,
  HostMetricsResponse,
  HostInventory,
  HostInventoryResponse,
  TimeRange,
} from '../types/host';

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
 * Get all host statuses
 */
export async function getAllHostStatuses(threshold = 300): Promise<HostStatus[]> {
  const response = await apiCall<HostStatusResponse>(
    `/hosts?threshold=${threshold}`
  );
  return response.status;
}

/**
 * Get status for a specific host
 */
export async function getHostStatus(
  collectorId: string,
  threshold = 300
): Promise<HostStatus> {
  const response = await apiCall<HostStatusResponse>(
    `/hosts/${collectorId}/status?threshold=${threshold}`
  );
  return response.status[0];
}

/**
 * Get host metrics for a time range
 */
export async function getHostMetrics(
  collectorId: string,
  timeRange: TimeRange = '24h',
  limit?: number
): Promise<HostMetrics[]> {
  const params = new URLSearchParams();
  params.append('time_range', timeRange);
  if (limit) {
    params.append('limit', limit.toString());
  }

  const response = await apiCall<HostMetricsResponse>(
    `/hosts/${collectorId}/metrics?${params.toString()}`
  );
  return response.data;
}

/**
 * Get host inventory (static configuration and specs)
 */
export async function getHostInventory(collectorId: string): Promise<HostInventory> {
  const response = await apiCall<HostInventoryResponse>(
    `/hosts/${collectorId}/inventory`
  );
  return response.data;
}

/**
 * Host API object with all methods
 */
export const hostApi = {
  getAllHostStatuses,
  getHostStatus,
  getHostMetrics,
  getHostInventory,
};

export default hostApi;