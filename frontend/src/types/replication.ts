/**
 * TypeScript types for replication monitoring
 * Mirrors backend models from backend/pkg/models/replication_models.go
 */

// ============================================================================
// REPLICATION METRICS TYPES
// ============================================================================

/**
 * Represents streaming replication status from pg_stat_replication
 */
export interface ReplicationStatus {
  id: string;
  collector_id: string;
  timestamp: string;
  server_pid: number;
  usename: string;
  application_name: string;
  state: string; // streaming, catchup, etc.
  sync_state: string; // sync, async, potential, quorum
  write_lsn: string;
  flush_lsn: string;
  replay_lsn: string;
  write_lag_ms: number;
  flush_lag_ms: number;
  replay_lag_ms: number;
  behind_by_mb: number;
  client_addr: string;
  backend_start: string | null;
}

/**
 * Represents a replication slot from pg_replication_slots
 */
export interface ReplicationSlot {
  id: string;
  collector_id: string;
  database_name: string;
  timestamp: string;
  slot_name: string;
  slot_type: 'physical' | 'logical';
  active: boolean;
  restart_lsn: string;
  confirmed_flush_lsn: string;
  wal_retained_mb: number;
  backend_pid: number;
  bytes_retained: number;
}

/**
 * Response containing all replication-related metrics
 */
export interface ReplicationMetricsResponse {
  replication_status?: ReplicationStatus[];
  replication_slots?: ReplicationSlot[];
  wal_status?: Record<string, unknown>;
}

// ============================================================================
// LOGICAL REPLICATION TYPES
// ============================================================================

/**
 * Represents a logical replication subscription from pg_stat_subscription
 */
export interface LogicalSubscription {
  time: string;
  collector_id: string;
  database_name: string;
  sub_name: string;
  sub_state: string; // ready, syncing, error, disabled
  sub_recv_lsn: string;
  sub_latest_end_lsn: string;
  sub_last_msg_receipt_time: string | null;
  sub_last_msg_send_time: string | null;
  sub_worker_pid: number;
  sub_worker_count: number;
}

/**
 * Represents a logical replication publication from pg_publication
 */
export interface Publication {
  time: string;
  collector_id: string;
  database_name: string;
  pub_name: string;
  pub_owner: string;
  pub_all_tables: boolean;
  pub_insert: boolean;
  pub_update: boolean;
  pub_delete: boolean;
  pub_truncate: boolean;
}

/**
 * Represents WAL receiver status from pg_stat_wal_receiver
 */
export interface WalReceiver {
  time: string;
  collector_id: string;
  status: string; // streaming, catching up, etc.
  sender_host: string;
  sender_port: number;
  received_lsn: string;
  latest_end_lsn: string;
  slot_name: string;
  conn_info: string;
}

// ============================================================================
// REPLICATION TOPOLOGY TYPES
// ============================================================================

/**
 * Represents a downstream node in the replication topology
 */
export interface TopologyNode {
  collector_id: string;
  application_name: string;
  client_addr: string;
  state: string;
  sync_state: string;
  replay_lag_ms: number;
}

/**
 * Represents the complete replication topology for a node
 */
export interface ReplicationTopology {
  collector_id: string;
  node_role: 'primary' | 'standby' | 'cascading_standby';
  upstream_collector_id?: string;
  upstream_host?: string;
  upstream_port?: number;
  downstream_count: number;
  downstream_nodes?: TopologyNode[];
}

// ============================================================================
// TOPOLOGY GRAPH TYPES (for @xyflow/react)
// ============================================================================

/**
 * Custom node data for topology graph visualization
 */
export interface TopologyNodeData {
  label: string;
  role: 'primary' | 'standby' | 'cascading_standby';
  status: 'streaming' | 'catchup' | 'down';
  lagMs: number;
  applicationName: string;
  clientAddr: string;
  collectorId: string;
}

/**
 * Custom edge data for topology graph visualization
 */
export interface TopologyEdgeData {
  lagMs: number;
  syncState: string;
  state: string;
}