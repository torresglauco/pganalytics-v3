/**
 * Data Classification API Client
 * Handles all API calls for classification and custom pattern operations
 */

import type {
  DataClassificationResult,
  ClassificationReportResponse,
  ClassificationFilter,
  ClassificationMetricsResponse,
  CustomPattern,
  CustomPatternResponse,
  PatternType,
  Category,
} from '../types/classification';

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
 * Get classification results for a collector with filtering
 */
export async function getClassificationResults(
  collectorId: string,
  filters: ClassificationFilter = {}
): Promise<ClassificationMetricsResponse> {
  const params = new URLSearchParams();

  if (filters.database) params.append('database', filters.database);
  if (filters.schema) params.append('schema', filters.schema);
  if (filters.table) params.append('table', filters.table);
  if (filters.pattern_type) params.append('pattern_type', filters.pattern_type);
  if (filters.category) params.append('category', filters.category);
  if (filters.time_range) params.append('time_range', filters.time_range);
  if (filters.limit) params.append('limit', filters.limit.toString());
  if (filters.offset) params.append('offset', filters.offset.toString());

  const queryString = params.toString();
  const endpoint = `/collectors/${collectorId}/classification${queryString ? `?${queryString}` : ''}`;

  return apiCall<ClassificationMetricsResponse>(endpoint);
}

/**
 * Get aggregated classification report for a collector
 */
export async function getClassificationReport(
  collectorId: string
): Promise<ClassificationReportResponse> {
  return apiCall<ClassificationReportResponse>(
    `/collectors/${collectorId}/classification/report`
  );
}

/**
 * Get custom detection patterns (global and tenant-specific)
 */
export async function getCustomPatterns(): Promise<CustomPatternResponse> {
  return apiCall<CustomPatternResponse>('/classification/patterns');
}

/**
 * Create a new custom detection pattern
 */
export async function createCustomPattern(
  pattern: Omit<CustomPattern, 'id' | 'created_at' | 'updated_at'>
): Promise<CustomPattern> {
  return apiCall<CustomPattern>('/classification/patterns', {
    method: 'POST',
    body: JSON.stringify(pattern),
  });
}

/**
 * Update an existing custom detection pattern
 */
export async function updateCustomPattern(
  id: number,
  pattern: Partial<Omit<CustomPattern, 'id' | 'created_at' | 'updated_at'>>
): Promise<CustomPattern> {
  return apiCall<CustomPattern>(`/classification/patterns/${id}`, {
    method: 'PUT',
    body: JSON.stringify(pattern),
  });
}

/**
 * Delete a custom detection pattern
 */
export async function deleteCustomPattern(id: number): Promise<void> {
  await apiCall<void>(`/classification/patterns/${id}`, {
    method: 'DELETE',
  });
}

/**
 * Classification API object for convenience
 */
export const classificationApi = {
  getClassificationResults,
  getClassificationReport,
  getCustomPatterns,
  createCustomPattern,
  updateCustomPattern,
  deleteCustomPattern,
};

export default classificationApi;