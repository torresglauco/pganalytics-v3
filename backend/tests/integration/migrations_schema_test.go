package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaMigrations validates all 4 advanced feature schema migrations
func TestSchemaMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pganalytics_test"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Verify database connection
	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}

	ctx := context.Background()

	t.Run("QueryPerformanceSchema", func(t *testing.T) {
		testQueryPerformanceSchema(t, db, ctx)
	})

	t.Run("LogAnalysisSchema", func(t *testing.T) {
		testLogAnalysisSchema(t, db, ctx)
	})

	t.Run("IndexAdvisorSchema", func(t *testing.T) {
		testIndexAdvisorSchema(t, db, ctx)
	})

	t.Run("VacuumAdvisorSchema", func(t *testing.T) {
		testVacuumAdvisorSchema(t, db, ctx)
	})

	t.Run("ForeignKeyConstraints", func(t *testing.T) {
		testForeignKeyConstraints(t, db, ctx)
	})

	t.Run("AllIndexesExist", func(t *testing.T) {
		testAllIndexesExist(t, db, ctx)
	})
}

// testQueryPerformanceSchema validates v3.1.0 query performance schema
func testQueryPerformanceSchema(t *testing.T, db *sql.DB, ctx context.Context) {
	// Verify query_plans table
	rows, err := db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'query_plans'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "query_plans table should exist")
	rows.Close()

	// Verify query_issues table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'query_issues'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "query_issues table should exist")
	rows.Close()

	// Verify query_performance_timeline table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'query_performance_timeline'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "query_performance_timeline table should exist")
	rows.Close()

	// Verify columns in query_plans
	columns := []string{"id", "database_id", "query_hash", "query_text", "plan_json", "mean_time", "total_time", "calls", "created_at"}
	for _, col := range columns {
		rows, err := db.QueryContext(ctx, `
			SELECT column_name FROM information_schema.columns
			WHERE table_schema = 'pganalytics' AND table_name = 'query_plans' AND column_name = $1
		`, col)
		require.NoError(t, err)
		assert.True(t, rows.Next(), "Column %s should exist in query_plans", col)
		rows.Close()
	}
}

// testLogAnalysisSchema validates v3.2.0 log analysis schema
func testLogAnalysisSchema(t *testing.T, db *sql.DB, ctx context.Context) {
	// Verify logs table
	rows, err := db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'logs'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "logs table should exist")
	rows.Close()

	// Verify log_patterns table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'log_patterns'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "log_patterns table should exist")
	rows.Close()

	// Verify log_anomalies table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'log_anomalies'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "log_anomalies table should exist")
	rows.Close()

	// Verify columns in logs
	columns := []string{"id", "database_id", "log_timestamp", "category", "severity", "message", "duration", "table_affected", "query_text", "created_at"}
	for _, col := range columns {
		rows, err := db.QueryContext(ctx, `
			SELECT column_name FROM information_schema.columns
			WHERE table_schema = 'pganalytics' AND table_name = 'logs' AND column_name = $1
		`, col)
		require.NoError(t, err)
		assert.True(t, rows.Next(), "Column %s should exist in logs", col)
		rows.Close()
	}
}

// testIndexAdvisorSchema validates v3.3.0 index advisor schema
func testIndexAdvisorSchema(t *testing.T, db *sql.DB, ctx context.Context) {
	// Verify index_recommendations table
	rows, err := db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'index_recommendations'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "index_recommendations table should exist")
	rows.Close()

	// Verify index_analysis table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'index_analysis'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "index_analysis table should exist")
	rows.Close()

	// Verify unused_indexes table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'unused_indexes'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "unused_indexes table should exist")
	rows.Close()

	// Verify columns in index_recommendations
	columns := []string{"id", "database_id", "table_name", "column_names", "index_type", "estimated_benefit", "weighted_cost_improvement", "status", "created_at"}
	for _, col := range columns {
		rows, err := db.QueryContext(ctx, `
			SELECT column_name FROM information_schema.columns
			WHERE table_schema = 'pganalytics' AND table_name = 'index_recommendations' AND column_name = $1
		`, col)
		require.NoError(t, err)
		assert.True(t, rows.Next(), "Column %s should exist in index_recommendations", col)
		rows.Close()
	}
}

// testVacuumAdvisorSchema validates v3.4.0 VACUUM advisor schema
func testVacuumAdvisorSchema(t *testing.T, db *sql.DB, ctx context.Context) {
	// Verify vacuum_recommendations table
	rows, err := db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'vacuum_recommendations'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "vacuum_recommendations table should exist")
	rows.Close()

	// Verify autovacuum_configurations table
	rows, err = db.QueryContext(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'pganalytics' AND table_name = 'autovacuum_configurations'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "autovacuum_configurations table should exist")
	rows.Close()

	// Verify columns in vacuum_recommendations
	columns := []string{"id", "database_id", "table_name", "table_size", "dead_tuples_count", "dead_tuples_ratio", "autovacuum_enabled", "autovacuum_naptime", "last_vacuum", "last_autovacuum", "recommendation_type", "estimated_gain", "created_at", "updated_at"}
	for _, col := range columns {
		rows, err := db.QueryContext(ctx, `
			SELECT column_name FROM information_schema.columns
			WHERE table_schema = 'pganalytics' AND table_name = 'vacuum_recommendations' AND column_name = $1
		`, col)
		require.NoError(t, err)
		assert.True(t, rows.Next(), "Column %s should exist in vacuum_recommendations", col)
		rows.Close()
	}
}

// testForeignKeyConstraints validates all foreign key relationships
func testForeignKeyConstraints(t *testing.T, db *sql.DB, ctx context.Context) {
	// Check foreign key from query_plans to databases
	rows, err := db.QueryContext(ctx, `
		SELECT constraint_name FROM information_schema.table_constraints
		WHERE table_schema = 'pganalytics' AND table_name = 'query_plans' AND constraint_type = 'FOREIGN KEY'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "query_plans should have foreign key to databases")
	rows.Close()

	// Check foreign key from logs to databases
	rows, err = db.QueryContext(ctx, `
		SELECT constraint_name FROM information_schema.table_constraints
		WHERE table_schema = 'pganalytics' AND table_name = 'logs' AND constraint_type = 'FOREIGN KEY'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "logs should have foreign key to databases")
	rows.Close()

	// Check foreign key from index_recommendations to databases
	rows, err = db.QueryContext(ctx, `
		SELECT constraint_name FROM information_schema.table_constraints
		WHERE table_schema = 'pganalytics' AND table_name = 'index_recommendations' AND constraint_type = 'FOREIGN KEY'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "index_recommendations should have foreign key to databases")
	rows.Close()

	// Check foreign key from vacuum_recommendations to databases
	rows, err = db.QueryContext(ctx, `
		SELECT constraint_name FROM information_schema.table_constraints
		WHERE table_schema = 'pganalytics' AND table_name = 'vacuum_recommendations' AND constraint_type = 'FOREIGN KEY'
	`)
	require.NoError(t, err)
	assert.True(t, rows.Next(), "vacuum_recommendations should have foreign key to databases")
	rows.Close()
}

// testAllIndexesExist validates all database indexes
func testAllIndexesExist(t *testing.T, db *sql.DB, ctx context.Context) {
	// Expected indexes - mapping table to index names
	expectedIndexes := map[string][]string{
		"query_plans": {
			"idx_query_plans_database",
		},
		"query_issues": {
			"idx_query_issues_query_plan",
		},
		"query_performance_timeline": {
			"idx_timeline_query_timestamp",
		},
		"logs": {
			"idx_logs_database_timestamp",
			"idx_logs_category",
		},
		"log_patterns": {
			"idx_log_patterns_database",
		},
		"log_anomalies": {
			"idx_log_anomalies_timestamp",
		},
		"index_recommendations": {
			"idx_recommendations_database",
			"idx_recommendations_status",
		},
		"index_analysis": {
			"idx_analysis_database",
			"idx_analysis_query",
			"idx_analysis_timestamp",
		},
		"unused_indexes": {
			"idx_unused_indexes_database",
			"idx_unused_indexes_table",
		},
		"vacuum_recommendations": {
			"idx_vacuum_recommendations_database",
			"idx_vacuum_recommendations_table",
			"idx_vacuum_recommendations_type",
			"idx_vacuum_recommendations_created",
			"idx_vacuum_recommendations_ratio",
		},
		"autovacuum_configurations": {
			"idx_autovacuum_configs_database",
			"idx_autovacuum_configs_table",
			"idx_autovacuum_configs_setting",
		},
	}

	for table, indexes := range expectedIndexes {
		for _, indexName := range indexes {
			rows, err := db.QueryContext(ctx, `
				SELECT indexname FROM pg_indexes
				WHERE schemaname = 'pganalytics' AND tablename = $1 AND indexname = $2
			`, table, indexName)
			require.NoError(t, err)
			assert.True(t, rows.Next(), "Index %s should exist on table %s", indexName, table)
			rows.Close()
		}
	}
}

// TestDataInsertionAndRetrieval validates that data can be inserted and retrieved
func TestDataInsertionAndRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pganalytics_test"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Verify database connection
	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}

	ctx := context.Background()

	// Create test database entry
	var testDatabaseID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.databases (name, host, port, username, password_encrypted, ssl_mode)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, "test_db", "localhost", 5432, "postgres", "encrypted", "require").Scan(&testDatabaseID)
	require.NoError(t, err, "Failed to insert test database")

	t.Run("QueryPerformanceInsertRetrieve", func(t *testing.T) {
		// Insert query_plan
		var planID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO pganalytics.query_plans (database_id, query_hash, query_text, plan_json, mean_time, total_time, calls)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`, testDatabaseID, "abc123", "SELECT * FROM users", `{"Plan": {}}`, 10.5, 105.0, 10).Scan(&planID)
		require.NoError(t, err)

		// Retrieve and verify
		var queryText string
		err = db.QueryRowContext(ctx, "SELECT query_text FROM pganalytics.query_plans WHERE id = $1", planID).Scan(&queryText)
		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users", queryText)
	})

	t.Run("LogAnalysisInsertRetrieve", func(t *testing.T) {
		// Insert log
		var logID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO pganalytics.logs (database_id, log_timestamp, category, severity, message)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, testDatabaseID, "2026-03-31 12:00:00", "query", "WARNING", "Slow query detected").Scan(&logID)
		require.NoError(t, err)

		// Retrieve and verify
		var message string
		err = db.QueryRowContext(ctx, "SELECT message FROM pganalytics.logs WHERE id = $1", logID).Scan(&message)
		require.NoError(t, err)
		assert.Equal(t, "Slow query detected", message)
	})

	t.Run("IndexAdvisorInsertRetrieve", func(t *testing.T) {
		// Insert recommendation
		var recID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO pganalytics.index_recommendations (database_id, table_name, column_names, estimated_benefit, weighted_cost_improvement)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, testDatabaseID, "users", `{email}`, 50.0, 75.5).Scan(&recID)
		require.NoError(t, err)

		// Retrieve and verify
		var tableName string
		err = db.QueryRowContext(ctx, "SELECT table_name FROM pganalytics.index_recommendations WHERE id = $1", recID).Scan(&tableName)
		require.NoError(t, err)
		assert.Equal(t, "users", tableName)
	})

	t.Run("VacuumAdvisorInsertRetrieve", func(t *testing.T) {
		// Insert vacuum recommendation
		var vacID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO pganalytics.vacuum_recommendations (database_id, table_name, dead_tuples_ratio, recommendation_type, estimated_gain)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, testDatabaseID, "large_table", 25.5, "full_vacuum", 1000000).Scan(&vacID)
		require.NoError(t, err)

		// Retrieve and verify
		var tableName string
		var ratio float64
		err = db.QueryRowContext(ctx, "SELECT table_name, dead_tuples_ratio FROM pganalytics.vacuum_recommendations WHERE id = $1", vacID).Scan(&tableName, &ratio)
		require.NoError(t, err)
		assert.Equal(t, "large_table", tableName)
		assert.InDelta(t, 25.5, ratio, 0.1)
	})
}
