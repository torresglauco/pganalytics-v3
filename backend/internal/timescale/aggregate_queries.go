package timescale

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
)

// DatabaseStatsAggregate represents aggregated database statistics
type DatabaseStatsAggregate struct {
	Bucket         time.Time `json:"bucket"`
	DatabaseName   string    `json:"database_name"`
	AvgBackends    float64   `json:"avg_backends"`
	MaxBackends    float64   `json:"max_backends"`
	TotalCommits   int64     `json:"total_commits"`
	TotalRollbacks int64     `json:"total_rollbacks"`
	TotalBlksRead  int64     `json:"total_blks_read"`
	TotalBlksHit   int64     `json:"total_blks_hit"`
	AvgDbSize      float64   `json:"avg_db_size"`
}

// TableStatsAggregate represents aggregated table statistics
type TableStatsAggregate struct {
	Bucket       time.Time `json:"bucket"`
	DatabaseName string    `json:"database_name"`
	TableName    string    `json:"table_name"`
	AvgSeqScan   float64   `json:"avg_seq_scan"`
	MaxSeqScan   float64   `json:"max_seq_scan"`
	AvgIdxScan   float64   `json:"avg_idx_scan"`
	MaxIdxScan   float64   `json:"max_idx_scan"`
	AvgLiveTup   float64   `json:"avg_live_tup"`
	MaxLiveTup   float64   `json:"max_live_tup"`
}

// SysstatAggregate represents aggregated system statistics
type SysstatAggregate struct {
	Bucket        time.Time `json:"bucket"`
	AvgCpuUser    float64   `json:"avg_cpu_user"`
	AvgCpuSystem  float64   `json:"avg_cpu_system"`
	AvgCpuIdle    float64   `json:"avg_cpu_idle"`
	AvgCpuIowait  float64   `json:"avg_cpu_iowait"`
	AvgLoad1m     float64   `json:"avg_load_1m"`
	AvgLoad5m     float64   `json:"avg_load_5m"`
	AvgLoad15m    float64   `json:"avg_load_15m"`
	AvgMemoryUsed float64   `json:"avg_memory_used"`
	MaxMemoryUsed float64   `json:"max_memory_used"`
}

// DashboardMetricsResponse contains all dashboard metrics
type DashboardMetricsResponse struct {
	DatabaseStats []DatabaseStatsAggregate `json:"database_stats"`
	TableStats    []TableStatsAggregate    `json:"table_stats"`
	SystemStats   []SysstatAggregate       `json:"system_stats"`
	TimeRange     string                   `json:"time_range"`
	QueriedAt     time.Time                `json:"queried_at"`
}

// GetDashboardDatabaseStats queries pre-computed database statistics
func (t *TimescaleDB) GetDashboardDatabaseStats(
	ctx context.Context,
	collectorID uuid.UUID,
	timeRange string,
) ([]DatabaseStatsAggregate, error) {
	// Select appropriate aggregate view based on time range
	var viewName string
	var intervalDuration string

	switch timeRange {
	case "1h":
		viewName = "metrics.db_stats_5m"
		intervalDuration = "1 hour"
	case "24h":
		viewName = "metrics.db_stats_5m"
		intervalDuration = "24 hours"
	case "7d":
		viewName = "metrics.db_stats_1h"
		intervalDuration = "7 days"
	case "30d":
		viewName = "metrics.db_stats_1h"
		intervalDuration = "30 days"
	default:
		viewName = "metrics.db_stats_5m"
		intervalDuration = "1 hour"
	}

	query := fmt.Sprintf(`
		SELECT bucket, database_name,
			   avg_backends, max_backends,
			   total_commits, total_rollbacks,
			   total_blks_read, total_blks_hit,
			   avg_db_size
		FROM %s
		WHERE collector_id = $1
		  AND bucket >= NOW() - $2::INTERVAL
		ORDER BY bucket DESC
		LIMIT 1000
	`, viewName)

	rows, err := t.db.QueryContext(ctx, query, collectorID, intervalDuration)
	if err != nil {
		return nil, apperrors.DatabaseError("query dashboard database stats", err.Error())
	}
	defer rows.Close()

	var results []DatabaseStatsAggregate
	for rows.Next() {
		var stats DatabaseStatsAggregate
		var avgBackends, maxBackends, avgDbSize sql.NullFloat64

		if err := rows.Scan(
			&stats.Bucket, &stats.DatabaseName,
			&avgBackends, &maxBackends,
			&stats.TotalCommits, &stats.TotalRollbacks,
			&stats.TotalBlksRead, &stats.TotalBlksHit,
			&avgDbSize,
		); err != nil {
			return nil, apperrors.DatabaseError("scan database stats row", err.Error())
		}

		if avgBackends.Valid {
			stats.AvgBackends = avgBackends.Float64
		}
		if maxBackends.Valid {
			stats.MaxBackends = maxBackends.Float64
		}
		if avgDbSize.Valid {
			stats.AvgDbSize = avgDbSize.Float64
		}

		results = append(results, stats)
	}

	return results, rows.Err()
}

// GetDashboardTableStats queries pre-computed table statistics
func (t *TimescaleDB) GetDashboardTableStats(
	ctx context.Context,
	collectorID uuid.UUID,
	timeRange string,
	limit int,
) ([]TableStatsAggregate, error) {
	if limit <= 0 {
		limit = 100
	}

	var viewName string
	var intervalDuration string

	switch timeRange {
	case "1h":
		viewName = "metrics.table_stats_5m"
		intervalDuration = "1 hour"
	case "24h":
		viewName = "metrics.table_stats_5m"
		intervalDuration = "24 hours"
	case "7d":
		viewName = "metrics.table_stats_1h"
		intervalDuration = "7 days"
	case "30d":
		viewName = "metrics.table_stats_1h"
		intervalDuration = "30 days"
	default:
		viewName = "metrics.table_stats_5m"
		intervalDuration = "1 hour"
	}

	query := fmt.Sprintf(`
		SELECT bucket, database_name, table_name,
			   avg_seq_scan, max_seq_scan,
			   avg_idx_scan, max_idx_scan,
			   avg_live_tup, max_live_tup
		FROM %s
		WHERE collector_id = $1
		  AND bucket >= NOW() - $2::INTERVAL
		ORDER BY bucket DESC
		LIMIT $3
	`, viewName)

	rows, err := t.db.QueryContext(ctx, query, collectorID, intervalDuration, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("query dashboard table stats", err.Error())
	}
	defer rows.Close()

	var results []TableStatsAggregate
	for rows.Next() {
		var stats TableStatsAggregate
		var avgSeqScan, maxSeqScan, avgIdxScan, maxIdxScan, avgLiveTup, maxLiveTup sql.NullFloat64

		if err := rows.Scan(
			&stats.Bucket, &stats.DatabaseName, &stats.TableName,
			&avgSeqScan, &maxSeqScan,
			&avgIdxScan, &maxIdxScan,
			&avgLiveTup, &maxLiveTup,
		); err != nil {
			return nil, apperrors.DatabaseError("scan table stats row", err.Error())
		}

		if avgSeqScan.Valid {
			stats.AvgSeqScan = avgSeqScan.Float64
		}
		if maxSeqScan.Valid {
			stats.MaxSeqScan = maxSeqScan.Float64
		}
		if avgIdxScan.Valid {
			stats.AvgIdxScan = avgIdxScan.Float64
		}
		if maxIdxScan.Valid {
			stats.MaxIdxScan = maxIdxScan.Float64
		}
		if avgLiveTup.Valid {
			stats.AvgLiveTup = avgLiveTup.Float64
		}
		if maxLiveTup.Valid {
			stats.MaxLiveTup = maxLiveTup.Float64
		}

		results = append(results, stats)
	}

	return results, rows.Err()
}

// GetDashboardSysstat queries pre-computed system statistics
func (t *TimescaleDB) GetDashboardSysstat(
	ctx context.Context,
	collectorID uuid.UUID,
	timeRange string,
) ([]SysstatAggregate, error) {
	var intervalDuration string

	switch timeRange {
	case "1h":
		intervalDuration = "1 hour"
	case "24h":
		intervalDuration = "24 hours"
	case "7d":
		intervalDuration = "7 days"
	case "30d":
		intervalDuration = "30 days"
	default:
		intervalDuration = "1 hour"
	}

	query := `
		SELECT bucket,
			   avg_cpu_user, avg_cpu_system, avg_cpu_idle, avg_cpu_iowait,
			   avg_load_1m, avg_load_5m, avg_load_15m,
			   avg_memory_used, max_memory_used
		FROM metrics.sysstat_5m
		WHERE collector_id = $1
		  AND bucket >= NOW() - $2::INTERVAL
		ORDER BY bucket DESC
		LIMIT 1000
	`

	rows, err := t.db.QueryContext(ctx, query, collectorID, intervalDuration)
	if err != nil {
		return nil, apperrors.DatabaseError("query dashboard sysstat", err.Error())
	}
	defer rows.Close()

	var results []SysstatAggregate
	for rows.Next() {
		var stats SysstatAggregate
		var avgCpuUser, avgCpuSystem, avgCpuIdle, avgCpuIowait sql.NullFloat64
		var avgLoad1m, avgLoad5m, avgLoad15m sql.NullFloat64
		var avgMemoryUsed, maxMemoryUsed sql.NullFloat64

		if err := rows.Scan(
			&stats.Bucket,
			&avgCpuUser, &avgCpuSystem, &avgCpuIdle, &avgCpuIowait,
			&avgLoad1m, &avgLoad5m, &avgLoad15m,
			&avgMemoryUsed, &maxMemoryUsed,
		); err != nil {
			return nil, apperrors.DatabaseError("scan sysstat row", err.Error())
		}

		if avgCpuUser.Valid {
			stats.AvgCpuUser = avgCpuUser.Float64
		}
		if avgCpuSystem.Valid {
			stats.AvgCpuSystem = avgCpuSystem.Float64
		}
		if avgCpuIdle.Valid {
			stats.AvgCpuIdle = avgCpuIdle.Float64
		}
		if avgCpuIowait.Valid {
			stats.AvgCpuIowait = avgCpuIowait.Float64
		}
		if avgLoad1m.Valid {
			stats.AvgLoad1m = avgLoad1m.Float64
		}
		if avgLoad5m.Valid {
			stats.AvgLoad5m = avgLoad5m.Float64
		}
		if avgLoad15m.Valid {
			stats.AvgLoad15m = avgLoad15m.Float64
		}
		if avgMemoryUsed.Valid {
			stats.AvgMemoryUsed = avgMemoryUsed.Float64
		}
		if maxMemoryUsed.Valid {
			stats.MaxMemoryUsed = maxMemoryUsed.Float64
		}

		results = append(results, stats)
	}

	return results, rows.Err()
}
