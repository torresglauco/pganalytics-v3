package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryPerformanceStore_GetSlowQueries(t *testing.T) {
	t.Run("returns queries sorted by mean_time descending", func(t *testing.T) {
		// Create mock database
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Wrap with PostgresDB
		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		// Mock getting database name
		rows := sqlmock.NewRows([]string{"name"}).AddRow("testdb")
		mock.ExpectQuery("SELECT name FROM databases WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		// Mock pg_stat_statements query
		statRows := sqlmock.NewRows([]string{
			"query_id", "query_hash", "query_text", "calls", "total_time_ms",
			"mean_time_ms", "min_time_ms", "max_time_ms", "rows",
			"shared_blks_hit", "database_name",
		}).
			AddRow(int64(101), "abc123", "SELECT * FROM users WHERE id = $1", int64(100), 500.0, 50.0, 10.0, 100.0, int64(1000), int64(500), "testdb").
			AddRow(int64(102), "def456", "SELECT * FROM orders WHERE user_id = $1", int64(50), 200.0, 40.0, 5.0, 80.0, int64(500), int64(300), "testdb")

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_statements").
			WithArgs("testdb", 20).
			WillReturnRows(statRows)

		// Execute
		ctx := context.Background()
		queries, err := store.GetSlowQueries(ctx, 1, 20)

		// Assert
		require.NoError(t, err)
		assert.Len(t, queries, 2)
		assert.Equal(t, int64(101), queries[0].QueryID)
		assert.Equal(t, "abc123", queries[0].QueryHash)
		assert.Equal(t, 50.0, queries[0].MeanTime)
		assert.Equal(t, int64(102), queries[1].QueryID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		// Mock getting database name
		rows := sqlmock.NewRows([]string{"name"}).AddRow("testdb")
		mock.ExpectQuery("SELECT name FROM databases WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		// Mock pg_stat_statements query with limit 10
		statRows := sqlmock.NewRows([]string{
			"query_id", "query_hash", "query_text", "calls", "total_time_ms",
			"mean_time_ms", "min_time_ms", "max_time_ms", "rows",
			"shared_blks_hit", "database_name",
		}).
			AddRow(int64(101), "abc123", "SELECT * FROM users", int64(100), 500.0, 50.0, 10.0, 100.0, int64(1000), int64(500), "testdb")

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_statements").
			WithArgs("testdb", 10).
			WillReturnRows(statRows)

		ctx := context.Background()
		queries, err := store.GetSlowQueries(ctx, 1, 10)

		require.NoError(t, err)
		assert.Len(t, queries, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles missing pg_stat_statements gracefully", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		// Mock getting database name
		rows := sqlmock.NewRows([]string{"name"}).AddRow("testdb")
		mock.ExpectQuery("SELECT name FROM databases WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		// Mock pg_stat_statements error (extension not available)
		// Error message contains "pg_stat_statements" to trigger graceful handling
		mock.ExpectQuery("SELECT(.+)FROM pg_stat_statements").
			WillReturnError(errors.New("relation \"pg_stat_statements\" does not exist"))

		ctx := context.Background()
		queries, err := store.GetSlowQueries(ctx, 1, 20)

		// Should return empty slice, not error
		require.NoError(t, err)
		assert.Empty(t, queries)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error for non-pg_stat_statements errors", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		// Mock getting database name
		rows := sqlmock.NewRows([]string{"name"}).AddRow("testdb")
		mock.ExpectQuery("SELECT name FROM databases WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		// Mock a different database error
		mock.ExpectQuery("SELECT(.+)FROM pg_stat_statements").
			WillReturnError(errors.New("connection refused"))

		ctx := context.Background()
		queries, err := store.GetSlowQueries(ctx, 1, 20)

		// Should return error for non-extension errors
		require.Error(t, err)
		assert.Empty(t, queries)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestQueryPerformanceStore_GetQueryTimeline(t *testing.T) {
	t.Run("returns time-series data for a query hash", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		now := time.Now()
		since := now.Add(-24 * time.Hour)

		// Mock timeline query
		rows := sqlmock.NewRows([]string{
			"metric_timestamp", "avg_duration", "max_duration", "executions",
		}).
			AddRow(now.Add(-2*time.Hour), 45.5, 120.0, int64(50)).
			AddRow(now.Add(-1*time.Hour), 42.0, 115.0, int64(48))

		mock.ExpectQuery("SELECT(.+)FROM query_performance_timeline").
			WithArgs("abc123", since).
			WillReturnRows(rows)

		ctx := context.Background()
		timeline, err := store.GetQueryTimeline(ctx, "abc123", since)

		require.NoError(t, err)
		assert.Len(t, timeline, 2)
		assert.Equal(t, 45.5, timeline[0].AvgDuration)
		assert.Equal(t, int64(50), timeline[0].Executions)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestQueryPerformanceStore_GetIndexStats(t *testing.T) {
	t.Run("returns index usage statistics", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &PostgresDB{db: db}
		store := NewQueryPerformanceStore(pgDB)

		// Mock getting database name
		rows := sqlmock.NewRows([]string{"name"}).AddRow("testdb")
		mock.ExpectQuery("SELECT name FROM databases WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		// Mock pg_stat_user_indexes query
		indexRows := sqlmock.NewRows([]string{
			"schemaname", "tablename", "indexname", "idx_scan",
			"idx_tup_read", "idx_tup_fetch", "size_bytes",
		}).
			AddRow("public", "users", "users_pkey", int64(1000), int64(5000), int64(5000), int64(8192)).
			AddRow("public", "users", "idx_users_email", int64(0), int64(0), int64(0), int64(4096))

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_user_indexes").
			WillReturnRows(indexRows)

		ctx := context.Background()
		stats, err := store.GetIndexStats(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, stats, 2)
		assert.Equal(t, "users_pkey", stats[0].IndexName)
		assert.Equal(t, int64(1000), stats[0].IdxScan)
		assert.False(t, stats[0].IsUnused)

		assert.Equal(t, "idx_users_email", stats[1].IndexName)
		assert.Equal(t, int64(0), stats[1].IdxScan)
		assert.True(t, stats[1].IsUnused)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
