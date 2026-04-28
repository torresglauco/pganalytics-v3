package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// CreateTestDatabase inserts a test database record
func CreateTestDatabase(t *testing.T, db *sql.DB, ctx context.Context, name string) int64 {
	var id int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.databases (name, host, port, username, password_encrypted, ssl_mode)
		VALUES ($1, 'localhost', 5432, 'postgres', 'encrypted', 'disable')
		RETURNING id
	`, name).Scan(&id)
	require.NoError(t, err, "Failed to create test database")
	return id
}

// CreateTestCollector inserts a test collector record
func CreateTestCollector(t *testing.T, db *sql.DB, ctx context.Context, dbID int64) uuid.UUID {
	id := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pganalytics.collectors (id, database_id, name, status)
		VALUES ($1, $2, 'test-collector', 'active')
	`, id, dbID)
	require.NoError(t, err, "Failed to create test collector")
	return id
}

// CreateTestInstance inserts a test instance record
func CreateTestInstance(t *testing.T, db *sql.DB, ctx context.Context, name string) int64 {
	var id int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.instances (name, host, port, username, ssl_mode, engine_version, status)
		VALUES ($1, 'localhost', 5432, 'postgres', 'disable', '16.0', 'active')
		RETURNING id
	`, name).Scan(&id)
	require.NoError(t, err, "Failed to create test instance")
	return id
}

// CreateTestUser inserts a test user record
func CreateTestUser(t *testing.T, db *sql.DB, ctx context.Context, email string) int64 {
	var id int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.users (email, password_hash, role)
		VALUES ($1, 'hashed_password', 'user')
		RETURNING id
	`, email).Scan(&id)
	require.NoError(t, err, "Failed to create test user")
	return id
}

// CleanupTable deletes all rows from a table (use in test cleanup)
func CleanupTable(t *testing.T, db *sql.DB, ctx context.Context, schema, table string) {
	_, err := db.ExecContext(ctx, "DELETE FROM "+schema+"."+table)
	if err != nil {
		t.Logf("Warning: failed to cleanup table %s.%s: %v", schema, table, err)
	}
}
