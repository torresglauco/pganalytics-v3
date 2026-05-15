/**
 * Replication API Client
 * Handles all API calls for replication monitoring operations
 */

import type {
  ReplicationTopology,
  ReplicationStatus,
  ReplicationSlot,
  LogicalSubscription,
  Publication,
  ReplicationMetricsResponse,
} from '../types/replication';

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
 * API response wrapper for metrics endpoints
 */
interface MetricsResponse<T> {
  metric_type: string;
  count: number;
  timestamp: string;
  data: T;
}

// ============================================================================
// REPLICATION TOPOLOGY
// ============================================================================

/**
 * Get replication topology for a collector
 * Shows cascading replication: primary -> standby -> standby
 */
export async function getTopology(collectorId: string): Promise<ReplicationTopology> {
  const response = await apiCall<{ topology: ReplicationTopology }>(
    `/collectors/${collectorId}/topology`
  );
  return response.topology;
}

// ============================================================================
// STREAMING REPLICATION
// ============================================================================

/**
 * Get streaming replication metrics with lag information
 */
export async function getReplicationMetrics(
  collectorId: string,
  options?: {
    limit?: number;
    offset?: number;
  }
): Promise<ReplicationMetricsResponse> {
  const params = new URLSearchParams();

  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  const response = await apiCall<MetricsResponse<ReplicationMetricsResponse>>(
    `/collectors/${collectorId}/replication?${params.toString()}`
  );

  return response.data;
}

/**
 * Get replication slots with WAL retention information
 */
export async function getReplicationSlots(
  collectorId: string,
  options?: {
    limit?: number;
    offset?: number;
  }
): Promise<ReplicationSlot[]> {
  const params = new URLSearchParams();

  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  const response = await apiCall<MetricsResponse<ReplicationSlot[]>>(
    `/collectors/${collectorId}/replication-slots?${params.toString()}`
  );

  return response.data;
}

// ============================================================================
// LOGICAL REPLICATION
// ============================================================================

/**
 * Get logical replication subscriptions
 */
export async function getLogicalSubscriptions(
  collectorId: string,
  options?: {
    database?: string;
    limit?: number;
    offset?: number;
  }
): Promise<LogicalSubscription[]> {
  const params = new URLSearchParams();

  if (options?.database) params.append('database', options.database);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  const response = await apiCall<MetricsResponse<LogicalSubscription[]>>(
    `/collectors/${collectorId}/logical-subscriptions?${params.toString()}`
  );

  return response.data;
}

/**
 * Get logical replication publications
 */
export async function getPublications(
  collectorId: string,
  options?: {
    database?: string;
    limit?: number;
    offset?: number;
  }
): Promise<Publication[]> {
  const params = new URLSearchParams();

  if (options?.database) params.append('database', options.database);
  if (options?.limit) params.append('limit', options.limit.toString());
  if (options?.offset) params.append('offset', options.offset.toString());

  const response = await apiCall<MetricsResponse<Publication[]>>(
    `/collectors/${collectorId}/publications?${params.toString()}`
  );

  return response.data;
}

/**
 * Export replication API as an object for consistent usage pattern
 */
export const replicationApi = {
  getTopology,
  getReplicationMetrics,
  getReplicationSlots,
  getLogicalSubscriptions,
  getPublications,
};

export default replicationApi;