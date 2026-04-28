package testutil

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertRowCount verifies a table has expected row count
func AssertRowCount(t *testing.T, db *sql.DB, table string, expected int) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	assert.NoError(t, err, "Failed to count rows in %s", table)
	assert.Equal(t, expected, count, "Row count in %s should be %d", table, expected)
}

// AssertTableExists verifies a table exists in the schema
func AssertTableExists(t *testing.T, db *sql.DB, schema, table string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = $1 AND table_name = $2
		)
	`, schema, table).Scan(&exists)
	assert.NoError(t, err, "Failed to check table existence")
	assert.True(t, exists, "Table %s.%s should exist", schema, table)
}

// AssertColumnExists verifies a column exists in a table
func AssertColumnExists(t *testing.T, db *sql.DB, schema, table, column string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2 AND column_name = $3
		)
	`, schema, table, column).Scan(&exists)
	assert.NoError(t, err, "Failed to check column existence")
	assert.True(t, exists, "Column %s should exist in table %s.%s", column, schema, table)
}

// AssertIndexExists verifies an index exists on a table
func AssertIndexExists(t *testing.T, db *sql.DB, schema, table, indexName string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = $1 AND tablename = $2 AND indexname = $3
		)
	`, schema, table, indexName).Scan(&exists)
	assert.NoError(t, err, "Failed to check index existence")
	assert.True(t, exists, "Index %s should exist on table %s.%s", indexName, schema, table)
}

// AssertForeignKeyExists verifies a foreign key constraint exists
func AssertForeignKeyExists(t *testing.T, db *sql.DB, schema, table string) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_schema = $1 AND table_name = $2 AND constraint_type = 'FOREIGN KEY'
	`, schema, table).Scan(&count)
	assert.NoError(t, err, "Failed to check foreign key existence")
	assert.Greater(t, count, 0, "Table %s.%s should have at least one foreign key", schema, table)
}
