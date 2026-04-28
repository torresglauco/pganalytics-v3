package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getQueryTestDB returns a database connection for testing
func getQueryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := getTestDBURL()

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

// TestEmptyResultSet verifies that sql.ErrNoRows is returned for empty results
// Requirement: TEST-08 - Empty result set handling
func TestEmptyResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create table with no data
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_empty_results (
			id SERIAL PRIMARY KEY,
			value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Ensure table is empty
	_, err = db.ExecContext(ctx, "DELETE FROM test_empty_results")
	require.NoError(t, err, "Failed to clean test table")

	// Test: Query returns sql.ErrNoRows for single row query
	var value string
	err = db.QueryRowContext(ctx, "SELECT value FROM test_empty_results WHERE id = $1", 999).Scan(&value)
	assert.Equal(t, sql.ErrNoRows, err, "Empty result should return sql.ErrNoRows")

	// Test: Query with QueryContext returns no rows (not an error)
	rows, err := db.QueryContext(ctx, "SELECT value FROM test_empty_results WHERE id = $1", 999)
	require.NoError(t, err, "Query should not error for empty result")
	defer rows.Close()

	assert.False(t, rows.Next(), "Rows.Next() should return false for empty result")

	// Test: Empty result with aggregation returns NULL/zero
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_empty_results").Scan(&count)
	require.NoError(t, err, "COUNT should work on empty table")
	assert.Equal(t, 0, count, "COUNT should return 0 for empty table")

	// Test: Empty result with MAX/MIN returns NULL
	var maxValue sql.NullFloat64
	err = db.QueryRowContext(ctx, "SELECT MAX(id) FROM test_empty_results").Scan(&maxValue)
	require.NoError(t, err, "MAX should work on empty table")
	assert.False(t, maxValue.Valid, "MAX should return NULL for empty table")
}

// TestNullValueHandling verifies proper handling of NULL values with sql.NullFloat64
// Requirement: TEST-08 - NULL value handling
func TestNullValueHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create table with NULL values
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_null_values (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			value FLOAT,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_null_values")
	require.NoError(t, err, "Failed to clean test table")

	// Insert row with NULL in value column
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_null_values (name, value, description)
		VALUES ($1, NULL, NULL)
	`, "null-test-item")
	require.NoError(t, err, "Failed to insert row with NULL values")

	// Query with sql.NullFloat64
	var id int
	var name string
	var value sql.NullFloat64
	var description sql.NullString

	err = db.QueryRowContext(ctx, `
		SELECT id, name, value, description
		FROM test_null_values
		WHERE name = $1
	`, "null-test-item").Scan(&id, &name, &value, &description)
	require.NoError(t, err, "Query should succeed")

	// Verify NULL detection
	assert.Equal(t, "null-test-item", name, "Name should be present")
	assert.False(t, value.Valid, "NULL value should have Valid=false")
	assert.False(t, description.Valid, "NULL description should have Valid=false")

	// Accessing Float64 on NULL should return 0 without panic
	assert.Equal(t, float64(0), value.Float64, "Float64 on NULL returns 0")
	assert.Equal(t, "", description.String, "String on NULL returns empty string")

	// Insert row with non-NULL value
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_null_values (name, value, description)
		VALUES ($1, $2, $3)
	`, "non-null-item", 42.5, "test description")
	require.NoError(t, err, "Failed to insert row with non-NULL values")

	// Query non-NULL row
	var id2 int
	var name2 string
	var value2 sql.NullFloat64
	var description2 sql.NullString

	err = db.QueryRowContext(ctx, `
		SELECT id, name, value, description
		FROM test_null_values
		WHERE name = $1
	`, "non-null-item").Scan(&id2, &name2, &value2, &description2)
	require.NoError(t, err, "Query should succeed")

	// Verify non-NULL detection
	assert.True(t, value2.Valid, "Non-NULL value should have Valid=true")
	assert.True(t, description2.Valid, "Non-NULL description should have Valid=true")
	assert.Equal(t, 42.5, value2.Float64, "Float64 should return actual value")
	assert.Equal(t, "test description", description2.String, "String should return actual value")
}

// TestLargeDatasetStreaming verifies streaming of 10,000+ rows without memory issues
// Requirement: TEST-08 - Large dataset streaming
func TestLargeDatasetStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create test table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_large_dataset (
			id SERIAL PRIMARY KEY,
			value TEXT,
			counter INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_large_dataset")
	require.NoError(t, err, "Failed to clean test table")

	// Insert 10,000 rows in batches
	batchSize := 1000
	totalRows := 10000

	for batch := 0; batch < totalRows/batchSize; batch++ {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err, "Failed to begin transaction for batch %d", batch)

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO test_large_dataset (value, counter)
			VALUES ($1, $2)
		`)
		require.NoError(t, err, "Failed to prepare statement for batch %d", batch)

		for i := 0; i < batchSize; i++ {
			id := batch*batchSize + i
			_, err = stmt.ExecContext(ctx, fmt.Sprintf("value-%d", id), id)
			require.NoError(t, err, "Failed to insert row %d", id)
		}

		stmt.Close()
		err = tx.Commit()
		require.NoError(t, err, "Failed to commit batch %d", batch)
	}

	// Verify all rows inserted
	var insertedCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_large_dataset").Scan(&insertedCount)
	require.NoError(t, err, "Failed to count rows")
	require.Equal(t, totalRows, insertedCount, "Should have inserted 10,000 rows")

	// Stream results using rows.Next()
	rows, err := db.QueryContext(ctx, "SELECT id, value, counter FROM test_large_dataset ORDER BY id")
	require.NoError(t, err, "Failed to query large dataset")
	defer rows.Close()

	processedCount := 0
	lastID := -1

	for rows.Next() {
		var id, counter int
		var value string
		err := rows.Scan(&id, &value, &counter)
		require.NoError(t, err, "Failed to scan row")

		// Verify ordering
		assert.Greater(t, id, lastID, "Rows should be in ascending order")
		lastID = id

		processedCount++
	}
	require.NoError(t, rows.Err(), "Rows iteration should not error")

	// Verify all rows processed
	assert.Equal(t, totalRows, processedCount, "Should have processed all 10,000 rows")

	// Test streaming with selective query
	rows2, err := db.QueryContext(ctx, `
		SELECT id, value FROM test_large_dataset
		WHERE counter % 100 = 0
		ORDER BY id
	`)
	require.NoError(t, err, "Failed to query subset")
	defer rows2.Close()

	selectiveCount := 0
	for rows2.Next() {
		var id int
		var value string
		err := rows2.Scan(&id, &value)
		require.NoError(t, err, "Failed to scan selective row")
		assert.Equal(t, 0, id%100, "ID should be multiple of 100")
		selectiveCount++
	}
	require.NoError(t, rows2.Err(), "Selective rows iteration should not error")
	assert.Equal(t, 100, selectiveCount, "Should have 100 rows (every 100th row)")
}

// TestMultipleNullColumns verifies independent handling of multiple NULL columns
// Requirement: TEST-08 - Multiple NULL columns handling
func TestMultipleNullColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create table with multiple nullable columns
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_multiple_nulls (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			col1 FLOAT,
			col2 TEXT,
			col3 INTEGER,
			col4 TIMESTAMP,
			col5 BOOLEAN,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_multiple_nulls")
	require.NoError(t, err, "Failed to clean test table")

	// Insert row with all NULL
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_multiple_nulls (name, col1, col2, col3, col4, col5)
		VALUES ('all-null', NULL, NULL, NULL, NULL, NULL)
	`)
	require.NoError(t, err, "Failed to insert all-NULL row")

	// Insert row with some NULL, some non-NULL
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_multiple_nulls (name, col1, col2, col3, col4, col5)
		VALUES ('mixed-null', 1.5, NULL, 42, NULL, true)
	`)
	require.NoError(t, err, "Failed to insert mixed-NULL row")

	// Insert row with no NULL
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_multiple_nulls (name, col1, col2, col3, col4, col5)
		VALUES ('no-null', 2.5, 'text', 100, '2026-01-01 00:00:00', false)
	`)
	require.NoError(t, err, "Failed to insert no-NULL row")

	// Query and verify each column's NULL state independently
	rows, err := db.QueryContext(ctx, `
		SELECT name, col1, col2, col3, col4, col5
		FROM test_multiple_nulls
		ORDER BY name
	`)
	require.NoError(t, err, "Failed to query multiple nulls")
	defer rows.Close()

	type rowData struct {
		name string
		col1 sql.NullFloat64
		col2 sql.NullString
		col3 sql.NullInt32
		col4 sql.NullTime
		col5 sql.NullBool
	}

	results := make(map[string]rowData)
	for rows.Next() {
		var r rowData
		err := rows.Scan(&r.name, &r.col1, &r.col2, &r.col3, &r.col4, &r.col5)
		require.NoError(t, err, "Failed to scan row")
		results[r.name] = r
	}
	require.NoError(t, rows.Err(), "Rows iteration should not error")

	// Verify all-null row
	allNull := results["all-null"]
	assert.False(t, allNull.col1.Valid, "col1 should be NULL")
	assert.False(t, allNull.col2.Valid, "col2 should be NULL")
	assert.False(t, allNull.col3.Valid, "col3 should be NULL")
	assert.False(t, allNull.col4.Valid, "col4 should be NULL")
	assert.False(t, allNull.col5.Valid, "col5 should be NULL")

	// Verify mixed-null row
	mixedNull := results["mixed-null"]
	assert.True(t, mixedNull.col1.Valid, "col1 should not be NULL")
	assert.Equal(t, 1.5, mixedNull.col1.Float64)
	assert.False(t, mixedNull.col2.Valid, "col2 should be NULL")
	assert.True(t, mixedNull.col3.Valid, "col3 should not be NULL")
	assert.Equal(t, int32(42), mixedNull.col3.Int32)
	assert.False(t, mixedNull.col4.Valid, "col4 should be NULL")
	assert.True(t, mixedNull.col5.Valid, "col5 should not be NULL")
	assert.True(t, mixedNull.col5.Bool)

	// Verify no-null row
	noNull := results["no-null"]
	assert.True(t, noNull.col1.Valid, "col1 should not be NULL")
	assert.True(t, noNull.col2.Valid, "col2 should not be NULL")
	assert.True(t, noNull.col3.Valid, "col3 should not be NULL")
	assert.True(t, noNull.col4.Valid, "col4 should not be NULL")
	assert.True(t, noNull.col5.Valid, "col5 should not be NULL")
	assert.Equal(t, 2.5, noNull.col1.Float64)
	assert.Equal(t, "text", noNull.col2.String)
	assert.Equal(t, int32(100), noNull.col3.Int32)
	assert.True(t, noNull.col5.Bool)
}

// TestMixedNullAndNonNullValues verifies correct handling of mixed NULL patterns across multiple rows
// Requirement: TEST-08 - Mixed NULL and non-NULL values
func TestMixedNullAndNonNullValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_mixed_values (
			id SERIAL PRIMARY KEY,
			category TEXT,
			amount FLOAT,
			status TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_mixed_values")
	require.NoError(t, err, "Failed to clean test table")

	// Insert rows with varying NULL patterns
	testData := []struct {
		category string
		amount   sql.NullFloat64
		status   sql.NullString
	}{
		{"electronics", sql.NullFloat64{Float64: 100.0, Valid: true}, sql.NullString{String: "active", Valid: true}},
		{"books", sql.NullFloat64{Float64: 0, Valid: false}, sql.NullString{String: "pending", Valid: true}},
		{"clothing", sql.NullFloat64{Float64: 50.0, Valid: true}, sql.NullString{String: "", Valid: false}},
		{"food", sql.NullFloat64{Float64: 0, Valid: false}, sql.NullString{String: "", Valid: false}},
		{"toys", sql.NullFloat64{Float64: 25.0, Valid: true}, sql.NullString{String: "active", Valid: true}},
	}

	for _, data := range testData {
		var amount interface{}
		var status interface{}

		if data.amount.Valid {
			amount = data.amount.Float64
		}
		if data.status.Valid {
			status = data.status.String
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO test_mixed_values (category, amount, status)
			VALUES ($1, $2, $3)
		`, data.category, amount, status)
		require.NoError(t, err, "Failed to insert test data")
	}

	// Query all rows and verify NULL indicators
	rows, err := db.QueryContext(ctx, `
		SELECT category, amount, status
		FROM test_mixed_values
		ORDER BY id
	`)
	require.NoError(t, err, "Failed to query mixed values")
	defer rows.Close()

	results := make(map[string]struct {
		amount sql.NullFloat64
		status sql.NullString
	})

	for rows.Next() {
		var category string
		var amount sql.NullFloat64
		var status sql.NullString

		err := rows.Scan(&category, &amount, &status)
		require.NoError(t, err, "Failed to scan row")
		results[category] = struct {
			amount sql.NullFloat64
			status sql.NullString
		}{amount, status}
	}
	require.NoError(t, rows.Err(), "Rows iteration should not error")

	// Verify each row's NULL indicators match expected
	for _, expected := range testData {
		result, ok := results[expected.category]
		require.True(t, ok, "Category %s should exist", expected.category)

		assert.Equal(t, expected.amount.Valid, result.amount.Valid,
			"Amount NULL state should match for %s", expected.category)
		assert.Equal(t, expected.status.Valid, result.status.Valid,
			"Status NULL state should match for %s", expected.category)

		if expected.amount.Valid {
			assert.Equal(t, expected.amount.Float64, result.amount.Float64,
				"Amount value should match for %s", expected.category)
		}
		if expected.status.Valid {
			assert.Equal(t, expected.status.String, result.status.String,
				"Status value should match for %s", expected.category)
		}
	}

	// Test aggregation with NULL handling
	var totalAmount sql.NullFloat64
	var countWithAmount int
	err = db.QueryRowContext(ctx, `
		SELECT SUM(amount), COUNT(amount)
		FROM test_mixed_values
	`).Scan(&totalAmount, &countWithAmount)
	require.NoError(t, err, "Aggregation should succeed")
	assert.True(t, totalAmount.Valid, "SUM should be valid")
	assert.Equal(t, 100.0+50.0+25.0, totalAmount.Float64, "SUM should only include non-NULL values")
	assert.Equal(t, 3, countWithAmount, "COUNT should only count non-NULL values")
}

// TestNullJSONHandling verifies NULL values in JSON columns
// Requirement: TEST-08 - NULL in JSON handling
func TestNullJSONHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Setup: Create table with JSON column
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test_null_json (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			data JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Cleanup before test
	_, err = db.ExecContext(ctx, "DELETE FROM test_null_json")
	require.NoError(t, err, "Failed to clean test table")

	// Insert row with NULL JSON
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_null_json (name, data) VALUES ('null-json', NULL)
	`)
	require.NoError(t, err, "Failed to insert NULL JSON")

	// Insert row with valid JSON
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_null_json (name, data) VALUES ('valid-json', '{"key": "value"}')
	`)
	require.NoError(t, err, "Failed to insert valid JSON")

	// Insert row with empty JSON object
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_null_json (name, data) VALUES ('empty-json', '{}')
	`)
	require.NoError(t, err, "Failed to insert empty JSON")

	// Query and verify NULL vs empty JSON
	rows, err := db.QueryContext(ctx, `
		SELECT name, data FROM test_null_json ORDER BY name
	`)
	require.NoError(t, err, "Failed to query JSON data")
	defer rows.Close()

	for rows.Next() {
		var name string
		var data sql.NullString

		err := rows.Scan(&name, &data)
		require.NoError(t, err, "Failed to scan row")

		switch name {
		case "empty-json":
			assert.True(t, data.Valid, "Empty JSON should be valid")
			assert.Equal(t, "{}", data.String, "Empty JSON should be {}")
		case "null-json":
			assert.False(t, data.Valid, "NULL JSON should not be valid")
		case "valid-json":
			assert.True(t, data.Valid, "Valid JSON should be valid")
			assert.Contains(t, data.String, "key", "Valid JSON should contain key")
		}
	}
	require.NoError(t, rows.Err(), "Rows iteration should not error")
}

// TestRowsClosePreventsLeaks verifies proper rows.Close() usage prevents connection leaks
// Requirement: TEST-08 - Connection leak prevention with proper cleanup
func TestRowsClosePreventsLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	db := getQueryTestDB(t)

	// Configure pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Get initial stats
	initialStats := db.Stats()

	// Execute multiple queries with proper rows.Close()
	for i := 0; i < 50; i++ {
		rows, err := db.QueryContext(ctx, "SELECT $1::int", i)
		require.NoError(t, err, "Query %d should succeed", i)

		// Immediately close rows (pattern: defer rows.Close())
		rows.Close()
	}

	// Allow connections to return to idle
	time.Sleep(100 * time.Millisecond)

	finalStats := db.Stats()

	// InUse should be back to initial
	assert.Equal(t, initialStats.InUse, finalStats.InUse,
		"InUse connections should return to initial level after proper cleanup")

	// No wait count
	assert.Equal(t, int64(0), finalStats.WaitCount,
		"WaitCount should be 0 with proper cleanup")
}

// TestQueryTimeout verifies query timeout handling
// Requirement: TEST-08 - Query timeout handling
func TestQueryTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := getQueryTestDB(t)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute a slow query
	start := time.Now()
	var result int
	err := db.QueryRowContext(ctx, "SELECT pg_sleep(0.2)").Scan(&result)
	elapsed := time.Since(start)

	// Should timeout
	require.Error(t, err, "Query should timeout")
	assert.Contains(t, err.Error(), "context deadline exceeded", "Should return context deadline exceeded")
	assert.Less(t, elapsed, 200*time.Millisecond, "Query should fail fast on timeout")
}
