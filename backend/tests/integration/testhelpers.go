package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestDB provides a test database connection and cleanup utilities
type TestDB struct {
	db     *sql.DB
	dbName string
}

// NewTestDB creates a new test database or uses a test PostgreSQL instance
// For testing purposes, we use an in-memory database simulation
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// For integration tests, we typically use a test PostgreSQL instance
	// This can be mocked for unit-level integration tests
	// In production, use testcontainers or docker-compose for full integration tests

	db := &TestDB{
		db:     nil, // Can be nil for service-level tests
		dbName: "test_pganalytics",
	}

	return db
}

// GetDB returns the underlying database connection
func (tdb *TestDB) GetDB() *sql.DB {
	return tdb.db
}

// Cleanup closes the database connection and cleans up resources
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()
	if tdb.db != nil {
		err := tdb.db.Close()
		if err != nil {
			t.Logf("Failed to close test database: %v", err)
		}
	}
}

// QueryHelper provides utilities for database queries in tests
type QueryHelper struct {
	db *sql.DB
}

// NewQueryHelper creates a new QueryHelper instance
func NewQueryHelper(db *sql.DB) *QueryHelper {
	return &QueryHelper{db: db}
}

// ExecuteQuery executes a query and returns the results as a map
func (qh *QueryHelper) ExecuteQuery(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := qh.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})
		for i, col := range cols {
			entry[col] = values[i]
		}
		results = append(results, entry)
	}

	return results, rows.Err()
}

// MockExplainOutput returns realistic EXPLAIN ANALYZE output for testing
func MockExplainOutput() string {
	return `{
    "Plan": {
      "Node Type": "Seq Scan",
      "Relation Name": "users",
      "Startup Cost": 0.0,
      "Total Cost": 35.50,
      "Rows": 1000,
      "Width": 200,
      "Plans": [
        {
          "Node Type": "Seq Scan",
          "Filter": "(active = true)",
          "Rows": 950
        }
      ]
    },
    "Execution Time": 2.345,
    "Planning Time": 0.123,
    "Total Time": 2.468
}`
}

// MockExplainOutputComplex returns a more complex EXPLAIN ANALYZE output
func MockExplainOutputComplex() string {
	return `{
    "Plan": {
      "Node Type": "Hash Join",
      "Join Type": "Inner",
      "Startup Cost": 45.50,
      "Total Cost": 125.75,
      "Rows": 500,
      "Width": 400,
      "Plans": [
        {
          "Node Type": "Seq Scan",
          "Relation Name": "orders",
          "Startup Cost": 0.0,
          "Total Cost": 35.50,
          "Rows": 1000
        },
        {
          "Node Type": "Hash",
          "Startup Cost": 35.50,
          "Total Cost": 35.50,
          "Rows": 250,
          "Plans": [
            {
              "Node Type": "Seq Scan",
              "Relation Name": "customers",
              "Filter": "(status = 'active')"
            }
          ]
        }
      ]
    },
    "Execution Time": 4.567,
    "Planning Time": 0.234,
    "Total Time": 4.801
}`
}

// MockPostgresLogEntries returns realistic PostgreSQL log entries
func MockPostgresLogEntries() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"message":   "duration: 123.456 ms  execute <unnamed>: SELECT * FROM users WHERE id = $1",
			"severity":  "LOG",
			"timestamp": time.Now().Add(-10 * time.Second).Format(time.RFC3339),
		},
		{
			"message":   "ERROR: duplicate key value violates unique constraint \"users_email_key\"",
			"severity":  "ERROR",
			"timestamp": time.Now().Add(-8 * time.Second).Format(time.RFC3339),
		},
		{
			"message":   "WARNING: you don't own a lock of type AccessShareLock",
			"severity":  "WARNING",
			"timestamp": time.Now().Add(-6 * time.Second).Format(time.RFC3339),
		},
		{
			"message":   "duration: 456.789 ms  statement: VACUUM ANALYZE products",
			"severity":  "LOG",
			"timestamp": time.Now().Add(-4 * time.Second).Format(time.RFC3339),
		},
		{
			"message":   "FATAL: database \"nonexistent\" does not exist",
			"severity":  "FATAL",
			"timestamp": time.Now().Add(-2 * time.Second).Format(time.RFC3339),
		},
		{
			"message":   "duration: 789.012 ms  parse <unnamed>: SELECT COUNT(*) FROM orders",
			"severity":  "LOG",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
}

// MockLogEntriesByCategory returns log entries organized by category
func MockLogEntriesByCategory() map[string][]map[string]interface{} {
	return map[string][]map[string]interface{}{
		"slow_query": {
			{
				"message":   "duration: 5000.123 ms  execute <unnamed>: SELECT * FROM large_table",
				"severity":  "LOG",
				"timestamp": time.Now().Add(-10 * time.Second).Format(time.RFC3339),
			},
		},
		"error": {
			{
				"message":   "ERROR: syntax error at or near \"SELECT\"",
				"severity":  "ERROR",
				"timestamp": time.Now().Add(-8 * time.Second).Format(time.RFC3339),
			},
			{
				"message":   "ERROR: permission denied for schema public",
				"severity":  "ERROR",
				"timestamp": time.Now().Add(-6 * time.Second).Format(time.RFC3339),
			},
		},
		"lock": {
			{
				"message":   "WARNING: you don't own a lock of type RowExclusiveLock",
				"severity":  "WARNING",
				"timestamp": time.Now().Add(-4 * time.Second).Format(time.RFC3339),
			},
		},
		"connection": {
			{
				"message":   "FATAL: database \"test\" does not exist",
				"severity":  "FATAL",
				"timestamp": time.Now().Add(-2 * time.Second).Format(time.RFC3339),
			},
		},
	}
}

// WaitForCondition waits for a condition to become true with timeout
func WaitForCondition(ctx context.Context, checkFn func() bool, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if checkFn() {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for condition")
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for condition")
			}
		}
	}
}

// AssertWithinDuration checks if a duration is within an expected range
func AssertWithinDuration(t *testing.T, actual, expected, tolerance time.Duration) {
	t.Helper()
	require.True(t,
		actual >= expected-tolerance && actual <= expected+tolerance,
		fmt.Sprintf("duration %v not within tolerance of expected %v (tolerance: %v)",
			actual, expected, tolerance),
	)
}

// AssertTimeRecent checks if a time is recent (within the last N seconds)
func AssertTimeRecent(t *testing.T, ts time.Time, maxAge time.Duration) {
	t.Helper()
	age := time.Since(ts)
	require.True(t,
		age <= maxAge,
		fmt.Sprintf("timestamp %v is too old (age: %v, max: %v)",
			ts, age, maxAge),
	)
}
