package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// HOST STATUS MODELS (HOST-01)
// ============================================================================

// HostStatus represents the up/down status of a host based on collector last_seen
type HostStatus struct {
	CollectorID                uuid.UUID  `json:"collector_id" db:"collector_id"`
	Hostname                   string     `json:"hostname" db:"hostname"`
	Status                     string     `json:"status" db:"status"` // up, down, unknown
	IsHealthy                  bool       `json:"is_healthy" db:"is_healthy"`
	LastSeen                   *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	UnresponsiveForSeconds     int64      `json:"unresponsive_for_seconds" db:"unresponsive_for_seconds"` // 0 if up, seconds since last_seen if down
	StatusChangedAt            *time.Time `json:"status_changed_at,omitempty" db:"status_changed_at"`
	ConfiguredThresholdSeconds int        `json:"configured_threshold_seconds" db:"configured_threshold_seconds"` // default 300 = 5 minutes
}

// HostStatusResponse contains host status data with metadata
type HostStatusResponse struct {
	Count  int           `json:"count"`
	Status []*HostStatus `json:"status"`
}

// ============================================================================
// HOST METRICS MODELS (HOST-02)
// ============================================================================

// HostMetrics represents OS-level metrics collected by sysstat collector
type HostMetrics struct {
	Time              time.Time `json:"time" db:"time"`
	CollectorID       uuid.UUID `json:"collector_id" db:"collector_id"`
	CpuUser           float64   `json:"cpu_user" db:"cpu_user"`         // percentage
	CpuSystem         float64   `json:"cpu_system" db:"cpu_system"`     // percentage
	CpuIdle           float64   `json:"cpu_idle" db:"cpu_idle"`         // percentage
	CpuIowait         float64   `json:"cpu_iowait" db:"cpu_iowait"`     // percentage
	CpuLoad1m         float64   `json:"cpu_load_1m" db:"cpu_load_1m"`   // load average
	CpuLoad5m         float64   `json:"cpu_load_5m" db:"cpu_load_5m"`   // load average
	CpuLoad15m        float64   `json:"cpu_load_15m" db:"cpu_load_15m"` // load average
	MemoryTotalMb     int64     `json:"memory_total_mb" db:"memory_total_mb"`
	MemoryFreeMb      int64     `json:"memory_free_mb" db:"memory_free_mb"`
	MemoryUsedMb      int64     `json:"memory_used_mb" db:"memory_used_mb"`
	MemoryCachedMb    int64     `json:"memory_cached_mb" db:"memory_cached_mb"`
	MemoryUsedPercent float64   `json:"memory_used_percent" db:"memory_used_percent"`
	DiskTotalGb       int64     `json:"disk_total_gb" db:"disk_total_gb"`
	DiskUsedGb        int64     `json:"disk_used_gb" db:"disk_used_gb"`
	DiskFreeGb        int64     `json:"disk_free_gb" db:"disk_free_gb"`
	DiskUsedPercent   float64   `json:"disk_used_percent" db:"disk_used_percent"`
	DiskIoReadOps     int64     `json:"disk_io_read_ops" db:"disk_io_read_ops"`   // cumulative
	DiskIoWriteOps    int64     `json:"disk_io_write_ops" db:"disk_io_write_ops"` // cumulative
	NetworkRxBytes    int64     `json:"network_rx_bytes" db:"network_rx_bytes"`   // cumulative
	NetworkTxBytes    int64     `json:"network_tx_bytes" db:"network_tx_bytes"`   // cumulative
}

// HostMetricsResponse contains host metrics data with metadata
type HostMetricsResponse struct {
	MetricType string         `json:"metric_type"`
	Count      int            `json:"count"`
	TimeRange  string         `json:"time_range"`
	Data       []*HostMetrics `json:"data"`
}

// ============================================================================
// HOST INVENTORY MODELS (HOST-03)
// ============================================================================

// HostInventory represents static host configuration and hardware specs
type HostInventory struct {
	Time                    time.Time `json:"time" db:"time"`
	CollectorID             uuid.UUID `json:"collector_id" db:"collector_id"`
	OsName                  string    `json:"os_name" db:"os_name"`       // e.g., "Ubuntu"
	OsVersion               string    `json:"os_version" db:"os_version"` // e.g., "22.04 LTS"
	OsKernel                string    `json:"os_kernel" db:"os_kernel"`   // e.g., "5.15.0-91-generic"
	CpuCores                int       `json:"cpu_cores" db:"cpu_cores"`
	CpuModel                string    `json:"cpu_model" db:"cpu_model"`
	CpuMHz                  float64   `json:"cpu_mhz" db:"cpu_mhz"`
	MemoryTotalMb           int64     `json:"memory_total_mb" db:"memory_total_mb"`
	DiskTotalGb             int64     `json:"disk_total_gb" db:"disk_total_gb"`
	PostgresVersion         string    `json:"postgres_version" db:"postgres_version"` // e.g., "16.2"
	PostgresEdition         string    `json:"postgres_edition" db:"postgres_edition"` // e.g., "Community", "EnterpriseDB"
	PostgresPort            int       `json:"postgres_port" db:"postgres_port"`
	PostgresDataDir         string    `json:"postgres_data_dir" db:"postgres_data_dir"`
	PostgresMaxConnections  int       `json:"postgres_max_connections" db:"postgres_max_connections"`
	PostgresSharedBuffersMb int       `json:"postgres_shared_buffers_mb" db:"postgres_shared_buffers_mb"`
	PostgresWorkMemMb       int       `json:"postgres_work_mem_mb" db:"postgres_work_mem_mb"`
}

// HostInventoryResponse contains host inventory data with metadata
type HostInventoryResponse struct {
	MetricType string         `json:"metric_type"`
	Data       *HostInventory `json:"data"`
}
