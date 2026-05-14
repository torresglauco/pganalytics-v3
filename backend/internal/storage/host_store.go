package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// HOST STATUS OPERATIONS (HOST-01)
// ============================================================================

// GetHostStatus retrieves the status of a single host based on collector last_seen
func (p *PostgresDB) GetHostStatus(ctx context.Context, collectorID uuid.UUID, thresholdSeconds int) (*models.HostStatus, error) {
	// Use default threshold of 5 minutes if not specified
	if thresholdSeconds <= 0 {
		thresholdSeconds = 300
	}

	query := `
		SELECT id, hostname, last_seen
		FROM collectors
		WHERE id = $1
	`

	var collectorIDResult uuid.UUID
	var hostname string
	var lastSeen *time.Time

	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(&collectorIDResult, &hostname, &lastSeen)
	if err != nil {
		return nil, apperrors.DatabaseError("query collector status", err.Error())
	}

	status := &models.HostStatus{
		CollectorID:                collectorIDResult,
		Hostname:                   hostname,
		LastSeen:                   lastSeen,
		ConfiguredThresholdSeconds: thresholdSeconds,
	}

	// Calculate status based on last_seen
	if lastSeen == nil {
		status.Status = "unknown"
		status.IsHealthy = false
		status.UnresponsiveForSeconds = 0
	} else {
		timeSinceLastSeen := time.Since(*lastSeen)
		threshold := time.Duration(thresholdSeconds) * time.Second

		if timeSinceLastSeen < threshold {
			status.Status = "up"
			status.IsHealthy = true
			status.UnresponsiveForSeconds = 0
		} else {
			status.Status = "down"
			status.IsHealthy = false
			status.UnresponsiveForSeconds = int64(timeSinceLastSeen.Seconds())
		}
	}

	return status, nil
}

// GetAllHostStatuses retrieves the status of all hosts
func (p *PostgresDB) GetAllHostStatuses(ctx context.Context, thresholdSeconds int) ([]*models.HostStatus, error) {
	// Use default threshold of 5 minutes if not specified
	if thresholdSeconds <= 0 {
		thresholdSeconds = 300
	}

	query := `
		SELECT id, hostname, last_seen
		FROM collectors
		ORDER BY hostname
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.DatabaseError("query all collectors status", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var statuses []*models.HostStatus
	threshold := time.Duration(thresholdSeconds) * time.Second

	for rows.Next() {
		var collectorID uuid.UUID
		var hostname string
		var lastSeen *time.Time

		if err := rows.Scan(&collectorID, &hostname, &lastSeen); err != nil {
			return nil, apperrors.DatabaseError("scan collector status", err.Error())
		}

		status := &models.HostStatus{
			CollectorID:                collectorID,
			Hostname:                   hostname,
			LastSeen:                   lastSeen,
			ConfiguredThresholdSeconds: thresholdSeconds,
		}

		// Calculate status based on last_seen
		if lastSeen == nil {
			status.Status = "unknown"
			status.IsHealthy = false
			status.UnresponsiveForSeconds = 0
		} else {
			timeSinceLastSeen := time.Since(*lastSeen)

			if timeSinceLastSeen < threshold {
				status.Status = "up"
				status.IsHealthy = true
				status.UnresponsiveForSeconds = 0
			} else {
				status.Status = "down"
				status.IsHealthy = false
				status.UnresponsiveForSeconds = int64(timeSinceLastSeen.Seconds())
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// ============================================================================
// HOST METRICS OPERATIONS (HOST-02)
// ============================================================================

// StoreHostMetrics inserts host metrics into the database
func (p *PostgresDB) StoreHostMetrics(ctx context.Context, metrics []*models.HostMetrics) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_host_metrics (
			time, collector_id,
			cpu_user, cpu_system, cpu_idle, cpu_iowait,
			cpu_load_1m, cpu_load_5m, cpu_load_15m,
			memory_total_mb, memory_free_mb, memory_used_mb, memory_cached_mb, memory_used_percent,
			disk_total_gb, disk_used_gb, disk_free_gb, disk_used_percent,
			disk_io_read_ops, disk_io_write_ops,
			network_rx_bytes, network_tx_bytes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare host metrics insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, m := range metrics {
		_, err := stmt.ExecContext(ctx,
			m.Time, m.CollectorID,
			m.CpuUser, m.CpuSystem, m.CpuIdle, m.CpuIowait,
			m.CpuLoad1m, m.CpuLoad5m, m.CpuLoad15m,
			m.MemoryTotalMb, m.MemoryFreeMb, m.MemoryUsedMb, m.MemoryCachedMb, m.MemoryUsedPercent,
			m.DiskTotalGb, m.DiskUsedGb, m.DiskFreeGb, m.DiskUsedPercent,
			m.DiskIoReadOps, m.DiskIoWriteOps,
			m.NetworkRxBytes, m.NetworkTxBytes,
		)
		if err != nil {
			return apperrors.DatabaseError("insert host metrics", err.Error())
		}
	}

	return tx.Commit()
}

// GetHostMetrics retrieves host metrics for a collector
func (p *PostgresDB) GetHostMetrics(ctx context.Context, collectorID uuid.UUID, timeRange string, limit int) ([]*models.HostMetrics, error) {
	// Build time filter based on timeRange
	var timeFilter string
	switch timeRange {
	case "1h":
		timeFilter = "AND time > NOW() - INTERVAL '1 hour'"
	case "7d":
		timeFilter = "AND time > NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "AND time > NOW() - INTERVAL '30 days'"
	default: // "24h"
		timeFilter = "AND time > NOW() - INTERVAL '24 hours'"
	}

	query := fmt.Sprintf(`
		SELECT time, collector_id,
			cpu_user, cpu_system, cpu_idle, cpu_iowait,
			cpu_load_1m, cpu_load_5m, cpu_load_15m,
			memory_total_mb, memory_free_mb, memory_used_mb, memory_cached_mb, memory_used_percent,
			disk_total_gb, disk_used_gb, disk_free_gb, disk_used_percent,
			disk_io_read_ops, disk_io_write_ops,
			network_rx_bytes, network_tx_bytes
		FROM metrics_host_metrics
		WHERE collector_id = $1 %s
		ORDER BY time DESC
		LIMIT $2
	`, timeFilter)

	rows, err := p.db.QueryContext(ctx, query, collectorID, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("query host metrics", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var metricsList []*models.HostMetrics
	for rows.Next() {
		m := &models.HostMetrics{}
		err := rows.Scan(
			&m.Time, &m.CollectorID,
			&m.CpuUser, &m.CpuSystem, &m.CpuIdle, &m.CpuIowait,
			&m.CpuLoad1m, &m.CpuLoad5m, &m.CpuLoad15m,
			&m.MemoryTotalMb, &m.MemoryFreeMb, &m.MemoryUsedMb, &m.MemoryCachedMb, &m.MemoryUsedPercent,
			&m.DiskTotalGb, &m.DiskUsedGb, &m.DiskFreeGb, &m.DiskUsedPercent,
			&m.DiskIoReadOps, &m.DiskIoWriteOps,
			&m.NetworkRxBytes, &m.NetworkTxBytes,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan host metrics", err.Error())
		}
		metricsList = append(metricsList, m)
	}

	return metricsList, nil
}

// ============================================================================
// HOST INVENTORY OPERATIONS (HOST-03)
// ============================================================================

// StoreHostInventory inserts host inventory into the database
func (p *PostgresDB) StoreHostInventory(ctx context.Context, inventory []*models.HostInventory) error {
	if len(inventory) == 0 {
		return nil
	}

	stmt, err := p.db.PrepareContext(ctx, `
		INSERT INTO metrics_host_inventory (
			time, collector_id,
			os_name, os_version, os_kernel,
			cpu_cores, cpu_model, cpu_mhz,
			memory_total_mb, disk_total_gb,
			postgres_version, postgres_edition,
			postgres_port, postgres_data_dir,
			postgres_max_connections, postgres_shared_buffers_mb, postgres_work_mem_mb
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (time, collector_id) DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare host inventory insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, inv := range inventory {
		_, err := stmt.ExecContext(ctx,
			inv.Time, inv.CollectorID,
			inv.OsName, inv.OsVersion, inv.OsKernel,
			inv.CpuCores, inv.CpuModel, inv.CpuMHz,
			inv.MemoryTotalMb, inv.DiskTotalGb,
			inv.PostgresVersion, inv.PostgresEdition,
			inv.PostgresPort, inv.PostgresDataDir,
			inv.PostgresMaxConnections, inv.PostgresSharedBuffersMb, inv.PostgresWorkMemMb,
		)
		if err != nil {
			return apperrors.DatabaseError("insert host inventory", err.Error())
		}
	}

	return nil
}

// GetHostInventory retrieves the latest host inventory for a collector
func (p *PostgresDB) GetHostInventory(ctx context.Context, collectorID uuid.UUID) (*models.HostInventory, error) {
	query := `
		SELECT time, collector_id,
			os_name, os_version, os_kernel,
			cpu_cores, cpu_model, cpu_mhz,
			memory_total_mb, disk_total_gb,
			postgres_version, postgres_edition,
			postgres_port, postgres_data_dir,
			postgres_max_connections, postgres_shared_buffers_mb, postgres_work_mem_mb
		FROM metrics_host_inventory
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT 1
	`

	inv := &models.HostInventory{}
	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(
		&inv.Time, &inv.CollectorID,
		&inv.OsName, &inv.OsVersion, &inv.OsKernel,
		&inv.CpuCores, &inv.CpuModel, &inv.CpuMHz,
		&inv.MemoryTotalMb, &inv.DiskTotalGb,
		&inv.PostgresVersion, &inv.PostgresEdition,
		&inv.PostgresPort, &inv.PostgresDataDir,
		&inv.PostgresMaxConnections, &inv.PostgresSharedBuffersMb, &inv.PostgresWorkMemMb,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("query host inventory", err.Error())
	}

	return inv, nil
}
