package timescale

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	_ "github.com/lib/pq"
)

// TimescaleDB wraps a TimescaleDB connection for time-series metrics
type TimescaleDB struct {
	db *sql.DB
}

// NewTimescaleDB creates a new TimescaleDB connection
func NewTimescaleDB(connString string) (*TimescaleDB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, apperrors.DatabaseError("open timescale connection", err.Error())
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, apperrors.DatabaseError("ping timescale", err.Error())
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &TimescaleDB{db: db}, nil
}

// Close closes the database connection
func (t *TimescaleDB) Close() error {
	return t.db.Close()
}

// Health checks the TimescaleDB health
func (t *TimescaleDB) Health(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return t.db.PingContext(ctx) == nil
}

// ============================================================================
// METRICS INSERTION
// ============================================================================

// MetricsPayload represents the structure of incoming metrics
type MetricsPayload struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// InsertPgStatsTableMetrics inserts PostgreSQL table statistics
func (t *TimescaleDB) InsertPgStatsTableMetrics(ctx context.Context, collectorID, serverID, databaseID *string, timestamp time.Time, metrics interface{}) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return apperrors.BadRequest("Invalid metrics format", err.Error())
	}

	// Extract fields from JSON for insertion
	// This is a simplified example - in production you'd deserialize properly
	_, err = t.db.ExecContext(
		ctx,
		`INSERT INTO metrics.metrics_pg_stats_table (time, collector_id, server_id, database_id)
		 VALUES ($1, $2, $3, $4)`,
		timestamp, collectorID, serverID, databaseID,
	)

	if err != nil {
		return apperrors.DatabaseError("insert pg_stats metrics", err.Error())
	}

	return nil
}

// InsertSysstatMetrics inserts system statistics
func (t *TimescaleDB) InsertSysstatMetrics(ctx context.Context, collectorID, serverID *string, timestamp time.Time, metrics interface{}) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return apperrors.BadRequest("Invalid metrics format", err.Error())
	}

	// TODO: Parse metrics and insert into sysstat table
	_ = metricsJSON

	_, err = t.db.ExecContext(
		ctx,
		`INSERT INTO metrics.metrics_sysstat (time, collector_id, server_id)
		 VALUES ($1, $2, $3)`,
		timestamp, collectorID, serverID,
	)

	if err != nil {
		return apperrors.DatabaseError("insert sysstat metrics", err.Error())
	}

	return nil
}

// InsertDiskUsageMetrics inserts disk usage statistics
func (t *TimescaleDB) InsertDiskUsageMetrics(ctx context.Context, collectorID, serverID *string, timestamp time.Time, metrics interface{}) error {
	_, err := t.db.ExecContext(
		ctx,
		`INSERT INTO metrics.metrics_disk_usage (time, collector_id, server_id)
		 VALUES ($1, $2, $3)`,
		timestamp, collectorID, serverID,
	)

	if err != nil {
		return apperrors.DatabaseError("insert disk usage metrics", err.Error())
	}

	return nil
}

// InsertPgLogMetrics inserts PostgreSQL log entries
func (t *TimescaleDB) InsertPgLogMetrics(ctx context.Context, collectorID, serverID, databaseID *string, timestamp time.Time, severity, message string) error {
	_, err := t.db.ExecContext(
		ctx,
		`INSERT INTO metrics.metrics_pg_log (time, collector_id, server_id, database_id, severity, message)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		timestamp, collectorID, serverID, databaseID, severity, message,
	)

	if err != nil {
		return apperrors.DatabaseError("insert pg_log metrics", err.Error())
	}

	return nil
}

// ============================================================================
// METRICS QUERYING
// ============================================================================

// QueryMetricsRange queries metrics for a time range
func (t *TimescaleDB) QueryMetricsRange(
	ctx context.Context,
	metricsType string,
	serverID int,
	startTime, endTime time.Time,
	limit int,
) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT * FROM metrics.metrics_%s
		WHERE server_id = $1 AND time >= $2 AND time <= $3
		ORDER BY time DESC
		LIMIT $4
	`, metricsType)

	rows, err := t.db.QueryContext(ctx, query, serverID, startTime, endTime, limit)
	if err != nil {
		return nil, apperrors.DatabaseError(fmt.Sprintf("query %s metrics", metricsType), err.Error())
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, apperrors.DatabaseError("get column names", err.Error())
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		// Create a slice to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the values
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, apperrors.DatabaseError("scan metrics row", err.Error())
		}

		// Create a map of the values
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}

		results = append(results, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("iterate metrics", err.Error())
	}

	return results, nil
}

// QueryLatestMetrics queries the latest metrics for a server
func (t *TimescaleDB) QueryLatestMetrics(
	ctx context.Context,
	metricsType string,
	serverID int,
) (map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT * FROM metrics.metrics_%s
		WHERE server_id = $1
		ORDER BY time DESC
		LIMIT 1
	`, metricsType)

	row := t.db.QueryRowContext(ctx, query, serverID)

	// Get column info
	// This is a simplified approach - in production you'd want proper type handling
	result := make(map[string]interface{})

	// Note: This is a placeholder implementation
	// Real implementation would depend on the specific metrics schema

	err := row.Scan()
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.DatabaseError("query latest metrics", err.Error())
	}

	return result, nil
}

// AggregateMetrics aggregates metrics over a time range (for dashboards)
func (t *TimescaleDB) AggregateMetrics(
	ctx context.Context,
	metricsType string,
	serverID int,
	startTime, endTime time.Time,
	bucketInterval string,
) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) as bucket,
			AVG(CAST(value as FLOAT)) as avg_value,
			MAX(CAST(value as FLOAT)) as max_value,
			MIN(CAST(value as FLOAT)) as min_value,
			COUNT(*) as count
		FROM metrics.metrics_%s
		WHERE server_id = $1 AND time >= $2 AND time <= $3
		GROUP BY bucket
		ORDER BY bucket DESC
	`, bucketInterval, metricsType)

	rows, err := t.db.QueryContext(ctx, query, serverID, startTime, endTime)
	if err != nil {
		return nil, apperrors.DatabaseError(fmt.Sprintf("aggregate %s metrics", metricsType), err.Error())
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var bucket time.Time
		var avgValue, maxValue, minValue sql.NullFloat64
		var count int

		if err := rows.Scan(&bucket, &avgValue, &maxValue, &minValue, &count); err != nil {
			return nil, apperrors.DatabaseError("scan aggregated metrics", err.Error())
		}

		result := map[string]interface{}{
			"bucket": bucket,
			"count":  count,
		}

		if avgValue.Valid {
			result["avg_value"] = avgValue.Float64
		}
		if maxValue.Valid {
			result["max_value"] = maxValue.Float64
		}
		if minValue.Valid {
			result["min_value"] = minValue.Float64
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetMetricsCount returns the count of metrics in a time range
func (t *TimescaleDB) GetMetricsCount(
	ctx context.Context,
	metricsType string,
	serverID int,
	startTime, endTime time.Time,
) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM metrics.metrics_%s
		WHERE server_id = $1 AND time >= $2 AND time <= $3
	`, metricsType)

	var count int
	err := t.db.QueryRowContext(ctx, query, serverID, startTime, endTime).Scan(&count)
	if err != nil {
		return 0, apperrors.DatabaseError(fmt.Sprintf("count %s metrics", metricsType), err.Error())
	}

	return count, nil
}
