/**
 * Host Monitoring Type Definitions
 * Types for host status, metrics, and inventory data
 */

/**
 * Host status - up/down status of a host based on collector last_seen
 */
export interface HostStatus {
  collector_id: string;
  hostname: string;
  status: 'up' | 'down' | 'unknown';
  is_healthy: boolean;
  last_seen?: string;
  unresponsive_for_seconds: number;
  status_changed_at?: string;
  configured_threshold_seconds: number;
}

/**
 * Host status response with metadata
 */
export interface HostStatusResponse {
  count: number;
  status: HostStatus[];
}

/**
 * Host metrics - OS-level metrics collected by sysstat collector
 */
export interface HostMetrics {
  time: string;
  collector_id: string;
  cpu_user: number;
  cpu_system: number;
  cpu_idle: number;
  cpu_iowait: number;
  cpu_load_1m: number;
  cpu_load_5m: number;
  cpu_load_15m: number;
  cpu_cores: number;
  memory_total_mb: number;
  memory_free_mb: number;
  memory_used_mb: number;
  memory_cached_mb: number;
  memory_used_percent: number;
  disk_total_gb: number;
  disk_used_gb: number;
  disk_free_gb: number;
  disk_used_percent: number;
  disk_io_read_ops: number;
  disk_io_write_ops: number;
  network_rx_bytes: number;
  network_tx_bytes: number;
}

/**
 * Host metrics response with metadata
 */
export interface HostMetricsResponse {
  metric_type: string;
  count: number;
  time_range: string;
  data: HostMetrics[];
}

/**
 * Host inventory - static host configuration and hardware specs
 */
export interface HostInventory {
  time: string;
  collector_id: string;
  os_name: string;
  os_version: string;
  os_kernel: string;
  cpu_cores: number;
  cpu_model: string;
  cpu_mhz: number;
  memory_total_mb: number;
  disk_total_gb: number;
  postgres_version: string;
  postgres_edition: string;
  postgres_port: number;
  postgres_data_dir: string;
  postgres_max_connections: number;
  postgres_shared_buffers_mb: number;
  postgres_work_mem_mb: number;
}

/**
 * Host inventory response with metadata
 */
export interface HostInventoryResponse {
  metric_type: string;
  data: HostInventory;
}

/**
 * Host summary statistics
 */
export interface HostSummaryStats {
  total_hosts: number;
  hosts_up: number;
  hosts_down: number;
  avg_cpu_percent: number;
  avg_memory_percent: number;
  avg_disk_percent: number;
}

/**
 * Time range options for metrics queries
 */
export type TimeRange = '1h' | '24h' | '7d' | '30d';

/**
 * Host status filter options
 */
export interface HostFilters {
  status?: ('up' | 'down' | 'unknown')[];
  search_term?: string;
}