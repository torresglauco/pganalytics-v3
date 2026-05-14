package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// REPLICATION METRICS MODELS
// ============================================================================

// ReplicationStatus represents streaming replication status from pg_stat_replication
type ReplicationStatus struct {
	ID              string     `json:"id" db:"id"`
	CollectorID     uuid.UUID  `json:"collector_id" db:"collector_id"`
	Timestamp       time.Time  `json:"timestamp" db:"time"`
	ServerPID       int64      `json:"server_pid" db:"server_pid"`
	Usename         string     `json:"usename" db:"usename"`
	ApplicationName string     `json:"application_name" db:"application_name"`
	State           string     `json:"state" db:"state"`                 // streaming, catchup, etc.
	SyncState       string     `json:"sync_state" db:"sync_state"`       // sync, async, potential, quorum
	WriteLsn        string     `json:"write_lsn" db:"write_lsn"`         // Write LSN
	FlushLsn        string     `json:"flush_lsn" db:"flush_lsn"`         // Flush LSN
	ReplayLsn       string     `json:"replay_lsn" db:"replay_lsn"`       // Replay LSN
	WriteLagMs      int64      `json:"write_lag_ms" db:"write_lag_ms"`   // Write lag in milliseconds
	FlushLagMs      int64      `json:"flush_lag_ms" db:"flush_lag_ms"`   // Flush lag in milliseconds
	ReplayLagMs     int64      `json:"replay_lag_ms" db:"replay_lag_ms"` // Replay lag in milliseconds
	BehindByMb      int64      `json:"behind_by_mb" db:"behind_by_mb"`   // Bytes behind converted to MB
	ClientAddr      string     `json:"client_addr" db:"client_addr"`
	BackendStart    *time.Time `json:"backend_start" db:"backend_start"`
}

// ReplicationSlot represents a replication slot from pg_replication_slots
type ReplicationSlot struct {
	ID                string    `json:"id" db:"id"`
	CollectorID       uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName      string    `json:"database_name" db:"database_name"`
	Timestamp         time.Time `json:"timestamp" db:"time"`
	SlotName          string    `json:"slot_name" db:"slot_name"`
	SlotType          string    `json:"slot_type" db:"slot_type"` // physical, logical
	Active            bool      `json:"active" db:"active"`
	RestartLsn        string    `json:"restart_lsn" db:"restart_lsn"`
	ConfirmedFlushLsn string    `json:"confirmed_flush_lsn" db:"confirmed_flush_lsn"`
	WalRetainedMb     int64     `json:"wal_retained_mb" db:"wal_retained_mb"`
	BackendPid        int64     `json:"backend_pid" db:"backend_pid"`
	BytesRetained     int64     `json:"bytes_retained" db:"bytes_retained"`
}

// ReplicationMetricsResponse contains all replication-related metrics
type ReplicationMetricsResponse struct {
	ReplicationStatus []*ReplicationStatus `json:"replication_status,omitempty"`
	ReplicationSlots  []*ReplicationSlot   `json:"replication_slots,omitempty"`
	WalStatus         json.RawMessage      `json:"wal_status,omitempty"`
}

// ============================================================================
// LOGICAL REPLICATION MODELS
// ============================================================================

// LogicalSubscription represents a logical replication subscription from pg_stat_subscription
type LogicalSubscription struct {
	Time                  time.Time  `json:"time" db:"time"`
	CollectorID           uuid.UUID  `json:"collector_id" db:"collector_id"`
	DatabaseName          string     `json:"database_name" db:"database_name"`
	SubName               string     `json:"sub_name" db:"sub_name"`
	SubState              string     `json:"sub_state" db:"sub_state"` // ready, syncing, error, disabled
	SubRecvLsn            string     `json:"sub_recv_lsn" db:"sub_recv_lsn"`
	SubLatestEndLsn       string     `json:"sub_latest_end_lsn" db:"sub_latest_end_lsn"`
	SubLastMsgReceiptTime *time.Time `json:"sub_last_msg_receipt_time" db:"sub_last_msg_receipt_time"`
	SubLastMsgSendTime    *time.Time `json:"sub_last_msg_send_time" db:"sub_last_msg_send_time"`
	SubWorkerPid          int64      `json:"sub_worker_pid" db:"sub_worker_pid"`
	SubWorkerCount        int        `json:"sub_worker_count" db:"sub_worker_count"`
}

// Publication represents a logical replication publication from pg_publication
type Publication struct {
	Time         time.Time `json:"time" db:"time"`
	CollectorID  uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName string    `json:"database_name" db:"database_name"`
	PubName      string    `json:"pub_name" db:"pub_name"`
	PubOwner     string    `json:"pub_owner" db:"pub_owner"`
	PubAllTables bool      `json:"pub_all_tables" db:"pub_all_tables"`
	PubInsert    bool      `json:"pub_insert" db:"pub_insert"`
	PubUpdate    bool      `json:"pub_update" db:"pub_update"`
	PubDelete    bool      `json:"pub_delete" db:"pub_delete"`
	PubTruncate  bool      `json:"pub_truncate" db:"pub_truncate"`
}

// WalReceiver represents WAL receiver status from pg_stat_wal_receiver (detects standby status)
type WalReceiver struct {
	Time         time.Time `json:"time" db:"time"`
	CollectorID  uuid.UUID `json:"collector_id" db:"collector_id"`
	Status       string    `json:"status" db:"status"` // streaming, catching up, etc.
	SenderHost   string    `json:"sender_host" db:"sender_host"`
	SenderPort   int       `json:"sender_port" db:"sender_port"`
	ReceivedLsn  string    `json:"received_lsn" db:"received_lsn"`
	LatestEndLsn string    `json:"latest_end_lsn" db:"latest_end_lsn"`
	SlotName     string    `json:"slot_name" db:"slot_name"`
	ConnInfo     string    `json:"conn_info" db:"conn_info"`
}

// ============================================================================
// REPLICATION TOPOLOGY MODELS
// ============================================================================

// TopologyNode represents a downstream node in the replication topology
type TopologyNode struct {
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	ClientAddr      string    `json:"client_addr" db:"client_addr"`
	State           string    `json:"state" db:"state"`
	SyncState       string    `json:"sync_state" db:"sync_state"`
	ReplayLagMs     int64     `json:"replay_lag_ms" db:"replay_lag_ms"`
}

// ReplicationTopology represents the complete replication topology for a node
type ReplicationTopology struct {
	CollectorID         uuid.UUID      `json:"collector_id" db:"collector_id"`
	NodeRole            string         `json:"node_role" db:"node_role"` // primary, standby, cascading_standby
	UpstreamCollectorID *uuid.UUID     `json:"upstream_collector_id,omitempty" db:"upstream_collector_id"`
	UpstreamHost        string         `json:"upstream_host,omitempty" db:"upstream_host"`
	UpstreamPort        int            `json:"upstream_port,omitempty" db:"upstream_port"`
	DownstreamCount     int            `json:"downstream_count" db:"downstream_count"`
	DownstreamNodes     []TopologyNode `json:"downstream_nodes,omitempty" db:"downstream_nodes"`
}
