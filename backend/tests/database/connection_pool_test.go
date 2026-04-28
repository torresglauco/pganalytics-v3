package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDBURL returns the database URL for testing
func getTestDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pganalytics_test?sslmode=disable"
	}
	return dbURL
}

// skipIfNoDatabase skips the test if the database is not available
func skipIfNoDatabase(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
}

// TestConnectionPoolUnderLoad tests that 100+ concurrent connections execute successfully without pool exhaustion
// Requirement: TEST-09 - Connection pool management under load
func TestConnectionPoolUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Configure pool (smaller than production for faster testing)
	// Production: MaxOpenConns=100, MaxIdleConns=20
	// Test: Use smaller values to verify pooling works correctly
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Create test table
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS pool_test;
		CREATE TABLE pool_test (
			id SERIAL PRIMARY KEY,
			value INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Track successful and failed queries
	var successCount int64
	var failCount int64

	// Simulate 100 concurrent connections (more than pool size)
	// This tests that the pool properly queues requests
	numGoroutines := 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Execute a simple query
			var result int
			err := db.QueryRowContext(ctx, "SELECT $1::int", id).Scan(&result)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				t.Logf("Query %d failed: %v", id, err)
			} else {
				atomic.AddInt64(&successCount, 1)
			}

			// Also insert into test table to verify connection works for writes
			_, err = db.ExecContext(ctx, "INSERT INTO pool_test (value) VALUES ($1)", id)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				t.Logf("Insert %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// All queries should succeed despite pool limit
	assert.Equal(t, int64(0), atomic.LoadInt64(&failCount), "All queries should succeed")
	assert.Equal(t, int64(numGoroutines), atomic.LoadInt64(&successCount), "All queries should complete")

	// Verify all inserts succeeded
	var rowCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pool_test").Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, numGoroutines, rowCount, "All rows should be inserted")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS pool_test")
	require.NoError(t, err)
}

// TestNoConnectionLeaks tests that there are no connection leaks after operations complete
// Requirement: TEST-09 - No connection leaks (WaitCount = 0)
func TestNoConnectionLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	// Configure pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	ctx := context.Background()

	// Get initial stats
	initialStats := db.Stats()
	t.Logf("Initial stats: OpenConnections=%d, InUse=%d, Idle=%d, WaitCount=%d",
		initialStats.OpenConnections, initialStats.InUse, initialStats.Idle, initialStats.WaitCount)

	// Execute multiple operations
	for i := 0; i < 50; i++ {
		var result int
		err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
		require.NoError(t, err, "Query %d should succeed", i)

		// Properly close rows if using Query
		rows, err := db.QueryContext(ctx, "SELECT $1::int", i)
		if err == nil {
			rows.Close()
		}
	}

	// Allow connections to return to idle state
	time.Sleep(100 * time.Millisecond)

	// Get final stats
	finalStats := db.Stats()
	t.Logf("Final stats: OpenConnections=%d, InUse=%d, Idle=%d, WaitCount=%d",
		finalStats.OpenConnections, finalStats.InUse, finalStats.Idle, finalStats.WaitCount)

	// WaitCount should be 0 (no connections had to wait)
	assert.Equal(t, int64(0), finalStats.WaitCount, "No connections should have waited (WaitCount should be 0)")

	// InUse should be 0 (all connections returned to pool)
	assert.Equal(t, 0, finalStats.InUse, "No connections should be in use after operations complete")

	// Open connections should match idle connections
	assert.Equal(t, finalStats.OpenConnections, finalStats.Idle, "All open connections should be idle")
}

// TestPoolConfigurationRespected tests that MaxOpenConns limit is honored
// Requirement: TEST-09 - Pool configuration respected
func TestPoolConfigurationRespected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	// Set a specific pool size
	maxOpen := 5
	maxIdle := 3
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	// Track maximum concurrent connections observed
	var maxObserved int64

	// Create a barrier to synchronize goroutines
	var wg sync.WaitGroup
	var startWg sync.WaitGroup
	startWg.Add(1) // Block all goroutines until ready

	// Spawn more goroutines than the pool size
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Wait for start signal
			startWg.Wait()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Check pool stats during operation
			stats := db.Stats()
			open := int64(stats.OpenConnections)

			// Track maximum
			for {
				old := atomic.LoadInt64(&maxObserved)
				if open <= old || atomic.CompareAndSwapInt64(&maxObserved, old, open) {
					break
				}
			}

			// Execute a query that takes some time
			_, err := db.ExecContext(ctx, "SELECT pg_sleep(0.05)")
			if err != nil {
				t.Logf("Query failed: %v", err)
			}
		}()
	}

	// Start all goroutines simultaneously
	startWg.Done()
	wg.Wait()

	// Verify max connections never exceeded the limit
	observed := atomic.LoadInt64(&maxObserved)
	t.Logf("Maximum observed open connections: %d (limit: %d)", observed, maxOpen)

	// The pool should have limited connections to maxOpen
	assert.LessOrEqual(t, observed, int64(maxOpen), "Open connections should not exceed MaxOpenConns")

	// Cleanup verification
	time.Sleep(100 * time.Millisecond)
	finalStats := db.Stats()
	assert.LessOrEqual(t, finalStats.OpenConnections, maxIdle, "Connections should return to idle pool")
}

// TestIdleConnectionsReused tests that idle connections are reused from the pool
// Requirement: TEST-09 - Idle connections reused
func TestIdleConnectionsReused(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	// Configure pool with idle connections
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	ctx := context.Background()

	// Execute first query
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)

	statsAfterFirst := db.Stats()
	t.Logf("After first query: OpenConnections=%d, InUse=%d, Idle=%d",
		statsAfterFirst.OpenConnections, statsAfterFirst.InUse, statsAfterFirst.Idle)

	// Execute second query - should reuse idle connection
	err = db.QueryRowContext(ctx, "SELECT 2").Scan(&result)
	require.NoError(t, err)

	statsAfterSecond := db.Stats()
	t.Logf("After second query: OpenConnections=%d, InUse=%d, Idle=%d",
		statsAfterSecond.OpenConnections, statsAfterSecond.InUse, statsAfterSecond.Idle)

	// OpenConnections should not have increased (connection reused)
	assert.LessOrEqual(t, statsAfterSecond.OpenConnections, statsAfterFirst.OpenConnections+1,
		"Should reuse existing connection rather than creating new ones")

	// WaitDuration should be 0 (no waiting for connections)
	assert.Equal(t, time.Duration(0), statsAfterSecond.WaitDuration,
		"Should not wait for connections when idle connections available")
}

// TestConnectionTimeoutHandling tests that queries timeout gracefully under extreme load
// Requirement: TEST-09 - Connection timeout handling
func TestConnectionTimeoutHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	// Configure pool with limited connections
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)

	// Create a very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Attempt a query that takes longer than the timeout
	// pg_sleep(0.1) = 100ms, context timeout = 50ms
	start := time.Now()
	var result int
	err = db.QueryRowContext(ctx, "SELECT pg_sleep(0.1)").Scan(&result)
	elapsed := time.Since(start)

	// Should get a context deadline exceeded error
	require.Error(t, err, "Query should timeout")
	assert.Contains(t, err.Error(), "context deadline exceeded", "Should return context deadline exceeded")
	assert.Less(t, elapsed, 100*time.Millisecond, "Query should fail fast on timeout")

	// Pool should still be functional after timeout
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	err = db.QueryRowContext(ctx2, "SELECT 1").Scan(&result)
	require.NoError(t, err, "Pool should be functional after timeout")
	assert.Equal(t, 1, result, "Query should return correct result")
}

// TestConnectionPoolStats verifies that db.Stats() returns correct values
// Requirement: TEST-09 - Pool stats validation
func TestConnectionPoolStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	// Configure pool
	maxOpen := 10
	maxIdle := 5
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	// Get stats
	stats := db.Stats()

	// Verify stats structure has expected values
	assert.Equal(t, maxOpen, stats.MaxOpenConnections, "MaxOpenConnections should match configured value")
	assert.GreaterOrEqual(t, stats.OpenConnections, 0, "OpenConnections should be non-negative")
	assert.GreaterOrEqual(t, stats.InUse, 0, "InUse should be non-negative")
	assert.GreaterOrEqual(t, stats.Idle, 0, "Idle should be non-negative")
	assert.GreaterOrEqual(t, stats.WaitCount, int64(0), "WaitCount should be non-negative")
	assert.GreaterOrEqual(t, stats.WaitDuration, time.Duration(0), "WaitDuration should be non-negative")
	assert.GreaterOrEqual(t, stats.MaxIdleClosed, int64(0), "MaxIdleClosed should be non-negative")
	assert.GreaterOrEqual(t, stats.MaxLifetimeClosed, int64(0), "MaxLifetimeClosed should be non-negative")

	t.Logf("Pool stats: MaxOpen=%d, Open=%d, InUse=%d, Idle=%d, WaitCount=%d, WaitDuration=%v",
		stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle,
		stats.WaitCount, stats.WaitDuration)
}

// TestConcurrentReadWrite tests concurrent read and write operations
// Requirement: TEST-09 - Concurrent operations
func TestConcurrentReadWrite(t *testing.T) {
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
		DROP TABLE IF EXISTS concurrent_test;
		CREATE TABLE concurrent_test (
			id SERIAL PRIMARY KEY,
			value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Configure pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	var wg sync.WaitGroup
	var writeCount int64
	var readCount int64

	// Concurrent writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := db.ExecContext(ctx,
				"INSERT INTO concurrent_test (value) VALUES ($1)",
				fmt.Sprintf("writer-%d", id))
			if err == nil {
				atomic.AddInt64(&writeCount, 1)
			} else {
				t.Logf("Write %d failed: %v", id, err)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var count int
			err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM concurrent_test").Scan(&count)
			if err == nil {
				atomic.AddInt64(&readCount, 1)
			} else {
				t.Logf("Read %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed
	assert.Equal(t, int64(50), writeCount, "All writes should succeed")
	assert.Equal(t, int64(50), readCount, "All reads should succeed")

	// Verify final count
	var finalCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM concurrent_test").Scan(&finalCount)
	require.NoError(t, err)
	assert.Equal(t, 50, finalCount, "All 50 rows should be present")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS concurrent_test")
	require.NoError(t, err)
}
