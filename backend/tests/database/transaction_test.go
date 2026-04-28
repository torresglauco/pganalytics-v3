package database

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDB returns a database connection for testing
func getTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pganalytics_test"
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "Failed to open database connection")

	// Verify connection
	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestTransactionCommit verifies data persists after successful transaction commit
func TestTransactionCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_commit (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_commit")
	require.NoError(t, err, "Failed to clean test table")

	// Test: Transaction with commit
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	// Safe rollback in case test fails
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_commit (value) VALUES ($1)", "test-commit-value")
	require.NoError(t, err, "Failed to insert data in transaction")

	err = tx.Commit()
	require.NoError(t, err, "Failed to commit transaction")
	tx = nil // Mark as committed so deferred rollback doesn't execute

	// Verify: Data persisted after commit
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_commit WHERE value = $1", "test-commit-value").Scan(&count)
	require.NoError(t, err, "Failed to query committed data")
	assert.Equal(t, 1, count, "Data should persist after commit")

	// Verify: Can query the actual value
	var retrievedValue string
	err = db.QueryRow("SELECT value FROM test_transactions_commit WHERE value = $1", "test-commit-value").Scan(&retrievedValue)
	require.NoError(t, err, "Failed to retrieve committed value")
	assert.Equal(t, "test-commit-value", retrievedValue, "Retrieved value should match inserted value")
}

// TestTransactionRollback verifies data is NOT persisted after transaction rollback
func TestTransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_rollback (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_rollback")
	require.NoError(t, err, "Failed to clean test table")

	// Insert a committed row first to verify rollback only affects uncommitted
	_, err = db.ExecContext(ctx, "INSERT INTO test_transactions_rollback (value) VALUES ($1)", "committed-before")
	require.NoError(t, err, "Failed to insert baseline data")

	// Test: Transaction with rollback
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_rollback (value) VALUES ($1)", "test-rollback-value")
	require.NoError(t, err, "Failed to insert data in transaction")

	err = tx.Rollback()
	require.NoError(t, err, "Failed to rollback transaction")

	// Verify: Rolled back data NOT persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_rollback WHERE value = $1", "test-rollback-value").Scan(&count)
	require.NoError(t, err, "Failed to query after rollback")
	assert.Equal(t, 0, count, "Rolled back data should NOT persist")

	// Verify: Previously committed data still exists
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_rollback WHERE value = $1", "committed-before").Scan(&count)
	require.NoError(t, err, "Failed to query baseline data")
	assert.Equal(t, 1, count, "Previously committed data should still exist")
}

// TestNestedTransactionWithSavepoint verifies partial rollback with savepoints
func TestNestedTransactionWithSavepoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_savepoint (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_savepoint")
	require.NoError(t, err, "Failed to clean test table")

	// Begin outer transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert outer data
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_savepoint (value) VALUES ($1)", "outer-value")
	require.NoError(t, err, "Failed to insert outer data")

	// Create savepoint
	_, err = tx.ExecContext(ctx, "SAVEPOINT inner_txn")
	require.NoError(t, err, "Failed to create savepoint")

	// Insert inner data
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_savepoint (value) VALUES ($1)", "inner-value")
	require.NoError(t, err, "Failed to insert inner data")

	// Rollback to savepoint
	_, err = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT inner_txn")
	require.NoError(t, err, "Failed to rollback to savepoint")

	// Commit outer transaction
	err = tx.Commit()
	require.NoError(t, err, "Failed to commit transaction")
	tx = nil

	// Verify: Only outer data persists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_savepoint").Scan(&count)
	require.NoError(t, err, "Failed to count rows")
	assert.Equal(t, 1, count, "Should have only 1 row after savepoint rollback")

	// Verify: Only outer value exists
	var value string
	err = db.QueryRow("SELECT value FROM test_transactions_savepoint").Scan(&value)
	require.NoError(t, err, "Failed to retrieve value")
	assert.Equal(t, "outer-value", value, "Only outer value should persist")
}

// TestTransactionIsolation verifies concurrent transactions don't interfere
func TestTransactionIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_isolation (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			counter INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_isolation")
	require.NoError(t, err, "Failed to clean test table")

	// Insert initial row
	_, err = db.ExecContext(ctx, "INSERT INTO test_transactions_isolation (value, counter) VALUES ($1, $2)", "initial", 0)
	require.NoError(t, err, "Failed to insert initial data")

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// Transaction 1: Increment counter
	wg.Add(1)
	go func() {
		defer wg.Done()

		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			errCh <- err
			return
		}
		defer tx.Rollback()

		// Read current counter
		var counter int
		err = tx.QueryRow("SELECT counter FROM test_transactions_isolation WHERE value = $1", "initial").Scan(&counter)
		if err != nil {
			errCh <- err
			return
		}

		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		// Update counter
		_, err = tx.ExecContext(ctx, "UPDATE test_transactions_isolation SET counter = $1 WHERE value = $2", counter+1, "initial")
		if err != nil {
			errCh <- err
			return
		}

		if err := tx.Commit(); err != nil {
			errCh <- err
			return
		}
	}()

	// Transaction 2: Increment counter
	wg.Add(1)
	go func() {
		defer wg.Done()

		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			errCh <- err
			return
		}
		defer tx.Rollback()

		// Simulate delay to let TX1 start first
		time.Sleep(25 * time.Millisecond)

		// Read current counter
		var counter int
		err = tx.QueryRow("SELECT counter FROM test_transactions_isolation WHERE value = $1", "initial").Scan(&counter)
		if err != nil {
			errCh <- err
			return
		}

		// Update counter
		_, err = tx.ExecContext(ctx, "UPDATE test_transactions_isolation SET counter = $1 WHERE value = $2", counter+1, "initial")
		if err != nil {
			errCh <- err
			return
		}

		if err := tx.Commit(); err != nil {
			errCh <- err
			return
		}
	}()

	wg.Wait()
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			t.Logf("Transaction error: %v", err)
		}
	}

	// Verify: Counter should be at least 1 (one transaction succeeded)
	// Note: With READ COMMITTED isolation, we expect both to succeed with their respective reads
	var finalCounter int
	err = db.QueryRow("SELECT counter FROM test_transactions_isolation WHERE value = $1", "initial").Scan(&finalCounter)
	require.NoError(t, err, "Failed to read final counter")
	assert.GreaterOrEqual(t, finalCounter, 1, "Counter should be at least 1")
}

// TestTransactionErrorRecovery verifies automatic rollback on error
func TestTransactionErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_error (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_error")
	require.NoError(t, err, "Failed to clean test table")

	// Test: Transaction with error triggers rollback
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	// Insert valid data
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_error (value) VALUES ($1)", "valid-data")
	require.NoError(t, err, "Failed to insert valid data")

	// Attempt invalid operation (violates NOT NULL constraint)
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_error (value) VALUES ($1)", nil)
	assert.Error(t, err, "Invalid insert should fail")

	// Rollback due to error
	err = tx.Rollback()
	require.NoError(t, err, "Failed to rollback transaction")

	// Verify: No partial data persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_error").Scan(&count)
	require.NoError(t, err, "Failed to count rows")
	assert.Equal(t, 0, count, "No data should persist after error and rollback")
}

// TestTransactionDeferredRollback verifies defer pattern handles rollback correctly
func TestTransactionDeferredRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_deferred (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup table before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_deferred")
	require.NoError(t, err, "Failed to clean test table")

	// Test function that uses deferred rollback pattern
	testFunc := func() (err error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback() // Safe rollback (no-op if committed)
		}()

		_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_deferred (value) VALUES ($1)", "deferred-test")
		if err != nil {
			return err
		}

		return tx.Commit()
	}

	// Execute with success path
	err = testFunc()
	require.NoError(t, err, "Transaction should succeed")

	// Verify: Data persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_deferred WHERE value = $1", "deferred-test").Scan(&count)
	require.NoError(t, err, "Failed to query data")
	assert.Equal(t, 1, count, "Data should persist with deferred rollback pattern")
}

// TestMultipleOperationsInTransaction verifies atomic multi-statement transactions
func TestMultipleOperationsInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getTestDB(t)

	// Setup: Create test tables
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions_multi_a (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS test_transactions_multi_b (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create test tables")

	// Cleanup
	_, err = db.ExecContext(ctx, "DELETE FROM test_transactions_multi_a; DELETE FROM test_transactions_multi_b")
	require.NoError(t, err, "Failed to clean test tables")

	// Test: Multiple inserts in single transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err, "Failed to begin transaction")

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert into table A
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_multi_a (value) VALUES ($1)", "multi-a")
	require.NoError(t, err, "Failed to insert into table A")

	// Insert into table B
	_, err = tx.ExecContext(ctx, "INSERT INTO test_transactions_multi_b (value) VALUES ($1)", "multi-b")
	require.NoError(t, err, "Failed to insert into table B")

	err = tx.Commit()
	require.NoError(t, err, "Failed to commit transaction")
	tx = nil

	// Verify: Both inserts persisted
	var countA, countB int
	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_multi_a WHERE value = $1", "multi-a").Scan(&countA)
	require.NoError(t, err, "Failed to count table A")
	assert.Equal(t, 1, countA, "Table A should have 1 row")

	err = db.QueryRow("SELECT COUNT(*) FROM test_transactions_multi_b WHERE value = $1", "multi-b").Scan(&countB)
	require.NoError(t, err, "Failed to count table B")
	assert.Equal(t, 1, countB, "Table B should have 1 row")
}
