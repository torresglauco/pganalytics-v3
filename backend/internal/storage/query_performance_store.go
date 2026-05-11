package storage

import (
	"context"
	"strings"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
)

// QueryPerformanceStore handles query performance data
type QueryPerformanceStore struct {
	db *PostgresDB
}

// NewQueryPerformanceStore creates a new query performance store
func NewQueryPerformanceStore(db *PostgresDB) *QueryPerformanceStore {
	return &QueryPerformanceStore{db: db}
}

// SlowQuery represents a slow query from pg_stat_statements
type SlowQuery struct {
	QueryID      int64   `json:"query_id"`
	QueryHash    string  `json:"query_hash"`
	QueryText    string  `json:"query_text"`
	Calls        int64   `json:"calls"`
	TotalTime    float64 `json:"total_time_ms"`
	MeanTime     float64 `json:"mean_time_ms"`
	MinTime      float64 `json:"min_time_ms"`
	MaxTime      float64 `json:"max_time_ms"`
	Rows         int64   `json:"rows"`
	SharedBlks   int64   `json:"shared_blks_hit"`
	DatabaseName string  `json:"database_name"`
}

// QueryTimelinePoint represents a single point in query performance timeline
type QueryTimelinePoint struct {
	Timestamp   time.Time `json:"timestamp"`
	AvgDuration float64   `json:"avg_duration_ms"`
	MaxDuration float64   `json:"max_duration_ms"`
	Executions  int64     `json:"executions"`
}

// IndexUsageStats represents index usage statistics
type IndexUsageStats struct {
	SchemaName  string `json:"schema_name"`
	TableName   string `json:"table_name"`
	IndexName   string `json:"index_name"`
	IdxScan     int64  `json:"idx_scan"`
	IdxTupRead  int64  `json:"idx_tup_read"`
	IdxTupFetch int64  `json:"idx_tup_fetch"`
	SizeBytes   int64  `json:"size_bytes"`
	IsUnused    bool   `json:"is_unused"`
}

// GetSlowQueries retrieves top N slow queries by mean execution time
func (s *QueryPerformanceStore) GetSlowQueries(ctx context.Context, databaseID int, limit int) ([]SlowQuery, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get database name from databaseID first
	var dbName string
	err := s.db.QueryRowContext(ctx,
		"SELECT name FROM databases WHERE id = $1",
		databaseID,
	).Scan(&dbName)
	if err != nil {
		return nil, apperrors.DatabaseError("get database name", err.Error())
	}

	query := `
		SELECT
			queryid as query_id,
			encode(sha256(query::bytea), 'hex')::substring(1, 16) as query_hash,
			query as query_text,
			calls,
			total_exec_time as total_time_ms,
			mean_exec_time as mean_time_ms,
			min_exec_time as min_time_ms,
			max_exec_time as max_time_ms,
			rows,
			shared_blks_hit,
			datname as database_name
		FROM pg_stat_statements
		JOIN pg_database ON pg_stat_statements.dbid = pg_database.oid
		WHERE dbid = (SELECT oid FROM pg_database WHERE datname = $1)
		ORDER BY mean_exec_time DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, dbName, limit)
	if err != nil {
		// Check if pg_stat_statements is missing
		if isPgStatStatementsMissing(err) {
			return []SlowQuery{}, nil // Return empty, not error
		}
		return nil, apperrors.DatabaseError("get slow queries", err.Error())
	}
	defer rows.Close()

	var queries []SlowQuery
	for rows.Next() {
		var q SlowQuery
		err := rows.Scan(
			&q.QueryID, &q.QueryHash, &q.QueryText,
			&q.Calls, &q.TotalTime, &q.MeanTime,
			&q.MinTime, &q.MaxTime, &q.Rows,
			&q.SharedBlks, &q.DatabaseName,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan slow query", err.Error())
		}
		queries = append(queries, q)
	}

	return queries, nil
}

// GetQueryTimeline retrieves historical performance for a query
func (s *QueryPerformanceStore) GetQueryTimeline(ctx context.Context, queryHash string, since time.Time) ([]QueryTimelinePoint, error) {
	query := `
		SELECT
			metric_timestamp,
			avg_duration,
			max_duration,
			executions
		FROM query_performance_timeline qpt
		JOIN query_plans qp ON qpt.query_plan_id = qp.id
		WHERE qp.query_hash = $1
		AND qpt.metric_timestamp >= $2
		ORDER BY qpt.metric_timestamp ASC
	`

	rows, err := s.db.QueryContext(ctx, query, queryHash, since)
	if err != nil {
		return nil, apperrors.DatabaseError("get query timeline", err.Error())
	}
	defer rows.Close()

	var timeline []QueryTimelinePoint
	for rows.Next() {
		var p QueryTimelinePoint
		err := rows.Scan(&p.Timestamp, &p.AvgDuration, &p.MaxDuration, &p.Executions)
		if err != nil {
			return nil, apperrors.DatabaseError("scan timeline point", err.Error())
		}
		timeline = append(timeline, p)
	}

	return timeline, nil
}

// GetIndexStats retrieves index usage statistics for a database
func (s *QueryPerformanceStore) GetIndexStats(ctx context.Context, databaseID int) ([]IndexUsageStats, error) {
	var dbName string
	err := s.db.QueryRowContext(ctx,
		"SELECT name FROM databases WHERE id = $1",
		databaseID,
	).Scan(&dbName)
	if err != nil {
		return nil, apperrors.DatabaseError("get database name", err.Error())
	}

	query := `
		SELECT
			schemaname,
			tablename,
			indexname,
			COALESCE(idx_scan, 0) as idx_scan,
			COALESCE(idx_tup_read, 0) as idx_tup_read,
			COALESCE(idx_tup_fetch, 0) as idx_tup_fetch,
			pg_relation_size(schemaname || '.' || indexname) as size_bytes
		FROM pg_stat_user_indexes
		ORDER BY idx_scan ASC, pg_relation_size(schemaname || '.' || indexname) DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.DatabaseError("get index stats", err.Error())
	}
	defer rows.Close()

	var stats []IndexUsageStats
	for rows.Next() {
		var s IndexUsageStats
		err := rows.Scan(
			&s.SchemaName, &s.TableName, &s.IndexName,
			&s.IdxScan, &s.IdxTupRead, &s.IdxTupFetch, &s.SizeBytes,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan index stat", err.Error())
		}
		s.IsUnused = s.IdxScan == 0
		stats = append(stats, s)
	}

	return stats, nil
}

// isPgStatStatementsMissing checks if the error is due to missing extension
func isPgStatStatementsMissing(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "pg_stat_statements") ||
		strings.Contains(err.Error(), "does not exist"))
}
