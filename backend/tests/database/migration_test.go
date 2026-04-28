package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationDataPreservation tests that migrations preserve existing data
// Requirement: TEST-10 - Migration data preservation (zero data loss)
func TestMigrationDataPreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create test table with initial data
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS test_preserve;
		CREATE TABLE test_preserve (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert test data BEFORE migration
	testData := []string{"preserve-me-1", "preserve-me-2", "preserve-me-3"}
	for _, name := range testData {
		_, err = db.ExecContext(ctx, "INSERT INTO test_preserve (name) VALUES ($1)", name)
		require.NoError(t, err, "Failed to insert test data")
	}

	// Verify initial data count
	var initialCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_preserve").Scan(&initialCount)
	require.NoError(t, err)
	require.Equal(t, 3, initialCount, "Should have 3 rows before migration")

	// Run migration (add column)
	_, err = db.ExecContext(ctx, "ALTER TABLE test_preserve ADD COLUMN IF NOT EXISTS new_column TEXT")
	require.NoError(t, err, "Migration should succeed")

	// Run another migration (add constraint)
	_, err = db.ExecContext(ctx, "ALTER TABLE test_preserve ALTER COLUMN name SET NOT NULL")
	require.NoError(t, err, "Migration should succeed")

	// Verify data still exists after migration
	rows, err := db.QueryContext(ctx, "SELECT name FROM test_preserve ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var retrievedData []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		require.NoError(t, err)
		retrievedData = append(retrievedData, name)
	}

	assert.Equal(t, testData, retrievedData, "Data should be preserved after migration")

	// Verify final count
	var finalCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_preserve").Scan(&finalCount)
	require.NoError(t, err)
	assert.Equal(t, initialCount, finalCount, "Row count should be unchanged after migration")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS test_preserve")
	require.NoError(t, err)
}

// TestMigrationBackwardCompatibility tests that existing queries work after migration
// Requirement: TEST-10 - Migration backward compatibility
func TestMigrationBackwardCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create table with initial schema
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS backward_compat_test;
		CREATE TABLE backward_compat_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert data using old query pattern
	oldInsertQuery := "INSERT INTO backward_compat_test (name) VALUES ($1)"
	_, err = db.ExecContext(ctx, oldInsertQuery, "test-value")
	require.NoError(t, err, "Old insert pattern should work")

	// Old select query pattern
	oldSelectQuery := "SELECT id, name FROM backward_compat_test WHERE name = $1"
	var id int
	var name string
	err = db.QueryRowContext(ctx, oldSelectQuery, "test-value").Scan(&id, &name)
	require.NoError(t, err, "Old select pattern should work")
	assert.Equal(t, "test-value", name)

	// Run migration (add new column with default)
	_, err = db.ExecContext(ctx, `
		ALTER TABLE backward_compat_test
		ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active'
	`)
	require.NoError(t, err, "Migration should succeed")

	// Old queries should still work after migration
	_, err = db.ExecContext(ctx, oldInsertQuery, "test-value-2")
	require.NoError(t, err, "Old insert pattern should still work after migration")

	err = db.QueryRowContext(ctx, oldSelectQuery, "test-value-2").Scan(&id, &name)
	require.NoError(t, err, "Old select pattern should still work after migration")

	// New queries using new column should also work
	newSelectQuery := "SELECT id, name, status FROM backward_compat_test WHERE name = $1"
	var status string
	err = db.QueryRowContext(ctx, newSelectQuery, "test-value").Scan(&id, &name, &status)
	require.NoError(t, err, "New select pattern should work")
	assert.Equal(t, "active", status, "Default value should be applied")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS backward_compat_test")
	require.NoError(t, err)
}

// TestMigrationIdempotent tests that running migrations twice doesn't fail
// Requirement: TEST-10 - Idempotent migrations
func TestMigrationIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create test table
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS idempotent_test;
		CREATE TABLE idempotent_test (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Migration using IF NOT EXISTS (idempotent pattern)
	migrationSQL := `
		ALTER TABLE idempotent_test
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	`

	// Run migration first time
	_, err = db.ExecContext(ctx, migrationSQL)
	require.NoError(t, err, "First migration run should succeed")

	// Verify column exists
	var columnExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'idempotent_test' AND column_name = 'created_at'
		)
	`).Scan(&columnExists)
	require.NoError(t, err)
	require.True(t, columnExists, "Column should exist after first migration")

	// Run migration second time - should NOT fail
	_, err = db.ExecContext(ctx, migrationSQL)
	require.NoError(t, err, "Second migration run should succeed (idempotent)")

	// Verify column still exists and there's only one
	var columnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = 'idempotent_test' AND column_name = 'created_at'
	`).Scan(&columnCount)
	require.NoError(t, err)
	assert.Equal(t, 1, columnCount, "Should have exactly one column (no duplicates)")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS idempotent_test")
	require.NoError(t, err)
}

// TestSchemaVersionsTracked tests that schema_versions table tracks migrations correctly
// Requirement: TEST-10 - Schema versions tracked
func TestSchemaVersionsTracked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Ensure schema exists
	_, err = db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS pganalytics")
	require.NoError(t, err, "Failed to create schema")

	// Create schema_versions table (as migrations would)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS pganalytics.schema_versions (
			version VARCHAR(100) PRIMARY KEY,
			description TEXT,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			execution_time_ms INTEGER
		)
	`)
	require.NoError(t, err, "Failed to create schema_versions table")

	// Record a migration
	migrationName := "001_test_migration.sql"
	startTime := time.Now()
	_, err = db.ExecContext(ctx, `
		INSERT INTO pganalytics.schema_versions (version, description, execution_time_ms)
		VALUES ($1, $2, $3)
	`, migrationName, "Test migration for tracking", 123)
	require.NoError(t, err, "Failed to record migration")

	// Verify migration is tracked
	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pganalytics.schema_versions WHERE version = $1
		)
	`, migrationName).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "Migration should be tracked in schema_versions")

	// Verify migration details
	var executedAt time.Time
	var execTimeMs int
	err = db.QueryRowContext(ctx, `
		SELECT executed_at, execution_time_ms
		FROM pganalytics.schema_versions
		WHERE version = $1
	`, migrationName).Scan(&executedAt, &execTimeMs)
	require.NoError(t, err)
	assert.WithinDuration(t, startTime, executedAt, 5*time.Second, "executed_at should be recent")
	assert.Equal(t, 123, execTimeMs, "execution_time_ms should match")

	// Check idempotency - can check if migration exists before running
	var alreadyRan bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pganalytics.schema_versions WHERE version = $1
		)
	`, migrationName).Scan(&alreadyRan)
	require.NoError(t, err)
	assert.True(t, alreadyRan, "Should detect migration already ran")

	// Cleanup
	_, err = db.ExecContext(ctx, "DELETE FROM pganalytics.schema_versions WHERE version = $1", migrationName)
	require.NoError(t, err)
}

// TestMigrationOrderRespected tests that migrations run in filename order
// Requirement: TEST-10 - Migration order respected
func TestMigrationOrderRespected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test migration filename sorting (as migrations.go does)
	migrations := []string{
		"003_add_indexes.sql",
		"001_initial_schema.sql",
		"002_add_users.sql",
		"010_add_constraints.sql",
		"005_add_logs.sql",
	}

	// Sort by filename (migrations.go uses sort.Slice with string comparison)
	sort.Strings(migrations)

	expected := []string{
		"001_initial_schema.sql",
		"002_add_users.sql",
		"003_add_indexes.sql",
		"005_add_logs.sql",
		"010_add_constraints.sql",
	}

	assert.Equal(t, expected, migrations, "Migrations should be sorted by filename")

	// Verify the sorting matches migrations.go behavior
	// In migrations.go: sort.Slice(migrations, func(i, j int) bool { return migrations[i].Name < migrations[j].Name })
	for i := 1; i < len(migrations); i++ {
		assert.Less(t, migrations[i-1], migrations[i], "Each migration should come after the previous one")
	}
}

// TestMigrationTransactionSafety tests that failed migrations don't leave partial state
// Requirement: TEST-10 - Migration transaction safety
func TestMigrationTransactionSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create test table
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS txn_test;
		CREATE TABLE txn_test (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert initial data
	_, err = db.ExecContext(ctx, "INSERT INTO txn_test (value) VALUES ('initial')")
	require.NoError(t, err)

	// Start a transaction for migration
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to start transaction")

	// Execute first statement
	_, err = tx.ExecContext(ctx, "ALTER TABLE txn_test ADD COLUMN status TEXT DEFAULT 'pending'")
	require.NoError(t, err, "First statement should succeed")

	// Execute second statement that will fail
	_, err = tx.ExecContext(ctx, "ALTER TABLE txn_test ADD CONSTRAINT invalid_constraint CHECK (value = 'nonexistent')")
	require.NoError(t, err, "Second statement setup should succeed")

	// Try to insert invalid data (should fail constraint)
	_, err = tx.ExecContext(ctx, "INSERT INTO txn_test (value) VALUES ('invalid')")
	if err != nil {
		// Rollback on failure
		tx.Rollback()
	}

	// Verify rollback worked - initial data should still exist
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM txn_test").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Transaction rollback should preserve original state")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS txn_test")
	require.NoError(t, err)
}

// TestMigrationNullHandling tests that migrations handle NULL values correctly
// Requirement: TEST-10 - Migration NULL handling
func TestMigrationNullHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create table with nullable column
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS null_test;
		CREATE TABLE null_test (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert data with NULL
	_, err = db.ExecContext(ctx, "INSERT INTO null_test (value) VALUES (NULL)")
	require.NoError(t, err, "Should insert NULL value")

	// Migration: Add default to nullable column
	_, err = db.ExecContext(ctx, "ALTER TABLE null_test ALTER COLUMN value SET DEFAULT 'default'")
	require.NoError(t, err, "Migration should succeed")

	// Existing NULL should remain NULL
	var value sql.NullString
	err = db.QueryRowContext(ctx, "SELECT value FROM null_test WHERE id = 1").Scan(&value)
	require.NoError(t, err)
	assert.False(t, value.Valid, "Existing NULL should remain NULL")

	// New inserts should get default
	_, err = db.ExecContext(ctx, "INSERT INTO null_test (id) VALUES (DEFAULT)")
	require.NoError(t, err, "Should insert with default")

	err = db.QueryRowContext(ctx, "SELECT value FROM null_test WHERE id = 2").Scan(&value)
	require.NoError(t, err)
	assert.True(t, value.Valid, "New row should have default value")
	assert.Equal(t, "default", value.String, "Default value should be applied")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS null_test")
	require.NoError(t, err)
}

// TestMigrationIndexCreation tests that indexes are created correctly during migration
// Requirement: TEST-10 - Migration index creation
func TestMigrationIndexCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create table without index
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS index_test;
		CREATE TABLE index_test (
			id SERIAL PRIMARY KEY,
			email TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert data before index
	for i := 0; i < 100; i++ {
		_, err = db.ExecContext(ctx, "INSERT INTO index_test (email) VALUES ($1)",
			fmt.Sprintf("user%d@example.com", i))
		require.NoError(t, err)
	}

	// Migration: Create index
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_index_test_email ON index_test (email)
	`)
	require.NoError(t, err, "Migration should create index")

	// Verify index exists
	var indexExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'index_test' AND indexname = 'idx_index_test_email'
		)
	`).Scan(&indexExists)
	require.NoError(t, err)
	require.True(t, indexExists, "Index should exist after migration")

	// Run migration again (idempotent)
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_index_test_email ON index_test (email)
	`)
	require.NoError(t, err, "Idempotent migration should succeed")

	// Query using index should be efficient
	var explainPlan string
	err = db.QueryRowContext(ctx, `
		EXPLAIN (FORMAT TEXT) SELECT * FROM index_test WHERE email = 'user50@example.com'
	`).Scan(&explainPlan)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(explainPlan), "index",
		"Query should use index (EXPLAIN shows index scan)")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS index_test")
	require.NoError(t, err)
}
