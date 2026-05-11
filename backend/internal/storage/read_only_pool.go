package storage

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
)

// ReadOnlyPool provides a dedicated read-only connection pool for dashboard queries.
// This isolates read-heavy dashboard operations from write operations on the primary pool.
type ReadOnlyPool struct {
	pool *pgxpool.Pool
	db   *sql.DB
}

// NewReadOnlyPool creates a read-only connection pool.
// Uses DATABASE_READ_ONLY_URL if available, otherwise falls back to primary connection string.
func NewReadOnlyPool(primaryConnString string) (*ReadOnlyPool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use read-only URL if provided, otherwise use primary
	connString := os.Getenv("DATABASE_READ_ONLY_URL")
	if connString == "" {
		connString = primaryConnString
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, apperrors.DatabaseError("parse read-only config", err.Error())
	}

	// Read-only pool configuration - optimized for dashboard queries
	maxConns := int32(50)
	if v := os.Getenv("MAX_READ_ONLY_CONNS"); v != "" {
		if m, err := strconv.ParseInt(v, 10, 32); err == nil && m > 0 {
			maxConns = int32(m)
		}
	}
	config.MaxConns = maxConns
	config.MinConns = int32(10) // Lower min for read-only
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 15 * time.Minute

	// Add read-only transaction mode to connection config
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set session to read-only mode
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY")
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, apperrors.DatabaseError("create read-only pool", err.Error())
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, apperrors.DatabaseError("ping read-only pool", err.Error())
	}

	db := stdlib.OpenDBFromPool(pool)
	return &ReadOnlyPool{pool: pool, db: db}, nil
}

// GetPoolMetrics returns connection pool statistics
func (r *ReadOnlyPool) GetPoolMetrics() PoolMetrics {
	if r.pool == nil {
		return PoolMetrics{}
	}
	stat := r.pool.Stat()
	return PoolMetrics{
		OpenConns:    stat.TotalConns(),
		IdleConns:    stat.IdleConns(),
		InUseConns:   stat.AcquiredConns(),
		MaxOpenConns: stat.MaxConns(),
		WaitCount:    stat.EmptyAcquireCount(),
		WaitDuration: stat.AcquireDuration().Milliseconds(),
	}
}

// QueryContext executes a query on the read-only pool
func (r *ReadOnlyPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query row on the read-only pool
func (r *ReadOnlyPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.db.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a statement on the read-only pool (will fail due to read-only mode)
func (r *ReadOnlyPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.db.ExecContext(ctx, query, args...)
}

// Close closes the read-only pool
func (r *ReadOnlyPool) Close() error {
	if r.db != nil {
		_ = r.db.Close()
	}
	if r.pool != nil {
		r.pool.Close()
	}
	return nil
}

// GetDB returns the underlying sql.DB connection
func (r *ReadOnlyPool) GetDB() *sql.DB {
	return r.db
}

// Health checks the read-only pool health
func (r *ReadOnlyPool) Health(ctx context.Context) bool {
	if r.pool == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.pool.Ping(ctx) == nil
}

// GetAllPoolMetrics creates a map of pool metrics for both primary and read-only pools
func GetAllPoolMetrics(primary PoolMetricsProvider, readOnly *ReadOnlyPool) map[string]PoolMetrics {
	metrics := make(map[string]PoolMetrics)
	if primary != nil {
		metrics["primary"] = primary.GetPoolMetrics()
	}
	if readOnly != nil {
		metrics["read_only"] = readOnly.GetPoolMetrics()
	}
	return metrics
}
