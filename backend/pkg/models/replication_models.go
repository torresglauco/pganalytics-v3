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
