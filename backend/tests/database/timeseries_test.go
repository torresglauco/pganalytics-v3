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

// TestTimezoneOrderingUTC tests that timestamps are ordered correctly in UTC
// Requirement: TEST-11 - Timestamps ordered correctly across timezones (UTC)
func TestTimezoneOrderingUTC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create time-series table
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS ts_test;
		CREATE TABLE ts_test (
			time TIMESTAMPTZ NOT NULL,
			value FLOAT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert timestamps in UTC
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	timestamps := []time.Time{
		baseTime.Add(1 * time.Hour),
		baseTime.Add(2 * time.Hour),
		baseTime.Add(3 * time.Hour),
		baseTime.Add(4 * time.Hour),
		baseTime.Add(5 * time.Hour),
	}

	for i, ts := range timestamps {
		_, err = db.ExecContext(ctx, "INSERT INTO ts_test (time, value) VALUES ($1, $2)", ts, float64(i)*10.0)
		require.NoError(t, err, "Failed to insert timestamp %d", i)
	}

	// Query ordered by time ascending
	rows, err := db.QueryContext(ctx, "SELECT time, value FROM ts_test ORDER BY time ASC")
	require.NoError(t, err)
	defer rows.Close()

	var prevTime time.Time
	var count int
	for rows.Next() {
		var ts time.Time
		var value float64
		err := rows.Scan(&ts, &value)
		require.NoError(t, err)

		if count > 0 {
			assert.True(t, ts.After(prevTime), "Timestamps should be ordered ascending: %s should be after %s", ts, prevTime)
		}
		prevTime = ts
		count++
	}

	assert.Equal(t, 5, count, "Should have 5 timestamps")

	// Query ordered by time descending
	rows2, err := db.QueryContext(ctx, "SELECT time FROM ts_test ORDER BY time DESC")
	require.NoError(t, err)
	defer rows2.Close()

	prevTime = time.Time{}
	count = 0
	for rows2.Next() {
		var ts time.Time
		err := rows2.Scan(&ts)
		require.NoError(t, err)

		if count > 0 {
			assert.True(t, ts.Before(prevTime), "Timestamps should be ordered descending: %s should be before %s", ts, prevTime)
		}
		prevTime = ts
		count++
	}

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS ts_test")
	require.NoError(t, err)
}

// TestTimezoneConversionPST tests that Pacific time converts to UTC correctly
// Requirement: TEST-11 - PST timezone conversion
func TestTimezoneConversionPST(t *testing.T) {
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
		DROP TABLE IF EXISTS tz_pst_test;
		CREATE TABLE tz_pst_test (
			time TIMESTAMPTZ NOT NULL,
			timezone TEXT,
			original_value TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// PST is UTC-8 (or UTC-7 during daylight saving)
	// 2026-01-15 10:00:00 PST = 2026-01-15 18:00:00 UTC (standard time)
	// We use a fixed location for testing
	pst := time.FixedZone("PST", -8*60*60) // UTC-8

	pstTime := time.Date(2026, 1, 15, 10, 0, 0, 0, pst)
	expectedUTC := time.Date(2026, 1, 15, 18, 0, 0, 0, time.UTC)

	// Insert timestamp - PostgreSQL stores in UTC internally
	_, err = db.ExecContext(ctx,
		"INSERT INTO tz_pst_test (time, timezone, original_value) VALUES ($1, $2, $3)",
		pstTime, "PST", "10:00 PST")
	require.NoError(t, err, "Failed to insert PST timestamp")

	// Retrieve and verify UTC value
	var storedTime time.Time
	err = db.QueryRowContext(ctx, "SELECT time FROM tz_pst_test").Scan(&storedTime)
	require.NoError(t, err)

	// PostgreSQL returns time in UTC
	assert.True(t, storedTime.Equal(expectedUTC),
		"PST timestamp should be converted to UTC: got %s, expected %s",
		storedTime.UTC(), expectedUTC)

	// Test querying with timezone in PostgreSQL
	var timezoneConverted time.Time
	err = db.QueryRowContext(ctx, "SELECT time AT TIME ZONE 'PST' FROM tz_pst_test").Scan(&timezoneConverted)
	require.NoError(t, err)

	// AT TIME ZONE 'PST' converts UTC back to PST
	// The result is a timestamp without timezone
	t.Logf("UTC stored: %s, PST converted: %s", storedTime.UTC(), timezoneConverted)

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS tz_pst_test")
	require.NoError(t, err)
}

// TestTimezoneConversionEST tests that Eastern time converts to UTC correctly
// Requirement: TEST-11 - EST timezone conversion
func TestTimezoneConversionEST(t *testing.T) {
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
		DROP TABLE IF EXISTS tz_est_test;
		CREATE TABLE tz_est_test (
			time TIMESTAMPTZ NOT NULL,
			value FLOAT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// EST is UTC-5 (standard time)
	// 2026-01-15 10:00:00 EST = 2026-01-15 15:00:00 UTC
	est := time.FixedZone("EST", -5*60*60) // UTC-5

	estTime := time.Date(2026, 1, 15, 10, 0, 0, 0, est)
	expectedUTC := time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC)

	// Insert timestamp
	_, err = db.ExecContext(ctx,
		"INSERT INTO tz_est_test (time, value) VALUES ($1, $2)",
		estTime, 42.5)
	require.NoError(t, err, "Failed to insert EST timestamp")

	// Retrieve and verify UTC value
	var storedTime time.Time
	var value float64
	err = db.QueryRowContext(ctx, "SELECT time, value FROM tz_est_test").Scan(&storedTime, &value)
	require.NoError(t, err)

	assert.True(t, storedTime.Equal(expectedUTC),
		"EST timestamp should be converted to UTC: got %s, expected %s",
		storedTime.UTC(), expectedUTC)
	assert.Equal(t, 42.5, value, "Value should be preserved")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS tz_est_test")
	require.NoError(t, err)
}

// TestTimeBucketAggregation tests hourly/daily bucket aggregation
// Requirement: TEST-11 - Time bucket aggregation produces correct intervals
func TestTimeBucketAggregation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create table with time-series data
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS metrics_test;
		CREATE TABLE metrics_test (
			time TIMESTAMPTZ NOT NULL,
			server_id INT,
			value FLOAT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert 24 hourly data points spanning one day
	// Values: 0.0, 1.0, 2.0, ... 23.0
	baseTime := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 24; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		_, err = db.ExecContext(ctx,
			"INSERT INTO metrics_test (time, server_id, value) VALUES ($1, $2, $3)",
			ts, 1, float64(i))
		require.NoError(t, err, "Failed to insert data point %d", i)
	}

	// PostgreSQL date_trunc for hourly bucketing (similar to TimescaleDB time_bucket)
	rows, err := db.QueryContext(ctx, `
		SELECT
			date_trunc('hour', time) as bucket,
			AVG(value) as avg_value,
			MAX(value) as max_value,
			MIN(value) as min_value,
			COUNT(*) as count
		FROM metrics_test
		WHERE server_id = 1
		GROUP BY bucket
		ORDER BY bucket
	`)
	require.NoError(t, err, "Hourly aggregation should succeed")
	defer rows.Close()

	var bucketCount int
	for rows.Next() {
		var bucket time.Time
		var avgValue, maxValue, minValue float64
		var count int
		err := rows.Scan(&bucket, &avgValue, &maxValue, &minValue, &count)
		require.NoError(t, err)

		// Each hour should have 1 value
		assert.Equal(t, 1, count, "Each hourly bucket should have 1 value")
		assert.Equal(t, avgValue, maxValue, "Avg should equal max for single value")
		assert.Equal(t, avgValue, minValue, "Avg should equal min for single value")
		bucketCount++
	}

	assert.Equal(t, 24, bucketCount, "Should have 24 hourly buckets")

	// Daily aggregation (all 24 hours in one bucket)
	rows2, err := db.QueryContext(ctx, `
		SELECT
			date_trunc('day', time) as bucket,
			AVG(value) as avg_value,
			MAX(value) as max_value,
			MIN(value) as min_value,
			COUNT(*) as count
		FROM metrics_test
		WHERE server_id = 1
		GROUP BY bucket
		ORDER BY bucket
	`)
	require.NoError(t, err, "Daily aggregation should succeed")
	defer rows2.Close()

	var dailyBucket time.Time
	var dailyAvg, dailyMax, dailyMin float64
	var dailyCount int
	require.True(t, rows2.Next(), "Should have one daily bucket")
	err = rows2.Scan(&dailyBucket, &dailyAvg, &dailyMax, &dailyMin, &dailyCount)
	require.NoError(t, err)

	assert.Equal(t, 24, dailyCount, "Daily bucket should have 24 values")
	assert.Equal(t, 23.0, dailyMax, "Max value should be 23.0")
	assert.Equal(t, 0.0, dailyMin, "Min value should be 0.0")
	assert.InDelta(t, 11.5, dailyAvg, 0.01, "Avg value should be 11.5")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS metrics_test")
	require.NoError(t, err)
}

// TestTimestampRangeQuery tests start/end time filtering
// Requirement: TEST-11 - Timestamp range queries work correctly
func TestTimestampRangeQuery(t *testing.T) {
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
		DROP TABLE IF EXISTS range_test;
		CREATE TABLE range_test (
			time TIMESTAMPTZ NOT NULL,
			event_id INT,
			data TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert data spanning multiple hours
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		_, err = db.ExecContext(ctx,
			"INSERT INTO range_test (time, event_id, data) VALUES ($1, $2, $3)",
			ts, i, fmt.Sprintf("event-%d", i))
		require.NoError(t, err, "Failed to insert data point %d", i)
	}

	// Query with start/end range (inclusive)
	startTime := baseTime.Add(2 * time.Hour) // 12:00
	endTime := baseTime.Add(6 * time.Hour)   // 16:00

	rows, err := db.QueryContext(ctx, `
		SELECT time, event_id FROM range_test
		WHERE time >= $1 AND time <= $2
		ORDER BY time
	`, startTime, endTime)
	require.NoError(t, err, "Range query should succeed")
	defer rows.Close()

	var count int
	var eventIDs []int
	for rows.Next() {
		var ts time.Time
		var eventID int
		err := rows.Scan(&ts, &eventID)
		require.NoError(t, err)

		// Verify timestamp is within range
		assert.True(t, ts.After(startTime) || ts.Equal(startTime),
			"Timestamp %s should be >= start %s", ts, startTime)
		assert.True(t, ts.Before(endTime) || ts.Equal(endTime),
			"Timestamp %s should be <= end %s", ts, endTime)

		eventIDs = append(eventIDs, eventID)
		count++
	}

	assert.Equal(t, 5, count, "Should have 5 events in range (12:00-16:00)")
	assert.Equal(t, []int{2, 3, 4, 5, 6}, eventIDs, "Event IDs should be 2-6")

	// Test exclusive end boundary
	rows2, err := db.QueryContext(ctx, `
		SELECT COUNT(*) FROM range_test
		WHERE time >= $1 AND time < $2
	`, startTime, endTime)
	require.NoError(t, err)
	defer rows2.Close()

	var exclusiveCount int
	require.True(t, rows2.Next())
	err = rows2.Scan(&exclusiveCount)
	require.NoError(t, err)
	assert.Equal(t, 4, exclusiveCount, "Exclusive end should have 4 events (12:00-15:59)")

	// Test boundary conditions
	// Exact match on start
	var startMatchCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM range_test WHERE time = $1
	`, startTime).Scan(&startMatchCount)
	require.NoError(t, err)
	assert.Equal(t, 1, startMatchCount, "Should find exact match at start boundary")

	// Exact match on end
	var endMatchCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM range_test WHERE time = $1
	`, endTime).Scan(&endMatchCount)
	require.NoError(t, err)
	assert.Equal(t, 1, endMatchCount, "Should find exact match at end boundary")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS range_test")
	require.NoError(t, err)
}

// TestTimezoneAcrossDayBoundary tests timezone handling across day boundaries
// Requirement: TEST-11 - Timezone handling across day boundaries
func TestTimezoneAcrossDayBoundary(t *testing.T) {
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
		DROP TABLE IF EXISTS day_boundary_test;
		CREATE TABLE day_boundary_test (
			time TIMESTAMPTZ NOT NULL,
			day_label TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert timestamps near midnight UTC
	beforeMidnight := time.Date(2026, 1, 15, 23, 30, 0, 0, time.UTC)
	atMidnight := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)
	afterMidnight := time.Date(2026, 1, 16, 0, 30, 0, 0, time.UTC)

	_, err = db.ExecContext(ctx,
		"INSERT INTO day_boundary_test (time, day_label) VALUES ($1, $2)",
		beforeMidnight, "Jan 15")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		"INSERT INTO day_boundary_test (time, day_label) VALUES ($1, $2)",
		atMidnight, "Midnight")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		"INSERT INTO day_boundary_test (time, day_label) VALUES ($1, $2)",
		afterMidnight, "Jan 16")
	require.NoError(t, err)

	// Query by date in UTC
	var jan15Count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM day_boundary_test
		WHERE time >= '2026-01-15 00:00:00 UTC' AND time < '2026-01-16 00:00:00 UTC'
	`).Scan(&jan15Count)
	require.NoError(t, err)
	assert.Equal(t, 1, jan15Count, "Should have 1 entry on Jan 15 UTC")

	// Query by date function
	var dayCounts []struct {
		day   time.Time
		count int
	}
	rows, err := db.QueryContext(ctx, `
		SELECT date_trunc('day', time) as day, COUNT(*) as count
		FROM day_boundary_test
		GROUP BY day
		ORDER BY day
	`)
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var day time.Time
		var count int
		err := rows.Scan(&day, &count)
		require.NoError(t, err)
		dayCounts = append(dayCounts, struct {
			day   time.Time
			count int
		}{day, count})
	}

	assert.Equal(t, 2, len(dayCounts), "Should have 2 days")
	assert.Equal(t, 1, dayCounts[0].count, "Jan 15 should have 1 entry")
	assert.Equal(t, 2, dayCounts[1].count, "Jan 16 should have 2 entries")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS day_boundary_test")
	require.NoError(t, err)
}

// TestNullTimestampHandling tests handling of NULL timestamps
// Requirement: TEST-11 - NULL timestamp handling
func TestNullTimestampHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := sql.Open("postgres", getTestDBURL())
	require.NoError(t, err)
	defer db.Close()

	skipIfNoDatabase(t, db)

	ctx := context.Background()

	// Create test table with nullable timestamp
	_, err = db.ExecContext(ctx, `
		DROP TABLE IF EXISTS null_ts_test;
		CREATE TABLE null_ts_test (
			id SERIAL PRIMARY KEY,
			event_time TIMESTAMPTZ,
			data TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert rows with NULL timestamps
	_, err = db.ExecContext(ctx, "INSERT INTO null_ts_test (event_time, data) VALUES (NULL, 'no-time')")
	require.NoError(t, err)

	// Insert rows with actual timestamps
	validTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, "INSERT INTO null_ts_test (event_time, data) VALUES ($1, 'has-time')", validTime)
	require.NoError(t, err)

	// Query using IS NULL
	var nullData string
	err = db.QueryRowContext(ctx,
		"SELECT data FROM null_ts_test WHERE event_time IS NULL").Scan(&nullData)
	require.NoError(t, err)
	assert.Equal(t, "no-time", nullData, "Should find row with NULL timestamp")

	// Query using IS NOT NULL
	var notNullData string
	var ts time.Time
	err = db.QueryRowContext(ctx,
		"SELECT data, event_time FROM null_ts_test WHERE event_time IS NOT NULL").Scan(&notNullData, &ts)
	require.NoError(t, err)
	assert.Equal(t, "has-time", notNullData, "Should find row with non-NULL timestamp")
	assert.True(t, ts.Equal(validTime), "Timestamp should match")

	// Use sql.NullTime for handling nullable timestamps
	rows, err := db.QueryContext(ctx, "SELECT event_time FROM null_ts_test ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var nullCount, validCount int
	for rows.Next() {
		var nullTime sql.NullTime
		err := rows.Scan(&nullTime)
		require.NoError(t, err)

		if nullTime.Valid {
			validCount++
			assert.True(t, nullTime.Time.Equal(validTime), "Valid time should match")
		} else {
			nullCount++
		}
	}

	assert.Equal(t, 1, nullCount, "Should have 1 NULL timestamp")
	assert.Equal(t, 1, validCount, "Should have 1 valid timestamp")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS null_ts_test")
	require.NoError(t, err)
}

// TestTimestampPrecision tests microsecond/nanosecond precision handling
// Requirement: TEST-11 - Timestamp precision
func TestTimestampPrecision(t *testing.T) {
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
		DROP TABLE IF EXISTS precision_test;
		CREATE TABLE precision_test (
			time TIMESTAMPTZ NOT NULL,
			data TEXT
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// PostgreSQL stores timestamps with microsecond precision
	// Go's time.Time can have nanosecond precision
	now := time.Now().UTC()
	nowWithNanos := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		123456789, // nanoseconds (will be truncated to microseconds in PostgreSQL)
		time.UTC,
	)

	// Insert with nanosecond precision
	_, err = db.ExecContext(ctx,
		"INSERT INTO precision_test (time, data) VALUES ($1, $2)",
		nowWithNanos, "nano-test")
	require.NoError(t, err, "Failed to insert timestamp")

	// Retrieve and check precision
	var storedTime time.Time
	err = db.QueryRowContext(ctx, "SELECT time FROM precision_test").Scan(&storedTime)
	require.NoError(t, err)

	// PostgreSQL truncates to microseconds (6 decimal places)
	// The stored time should be within 1 microsecond of the original
	diff := storedTime.Sub(nowWithNanos).Abs()
	assert.LessOrEqual(t, diff.Microseconds(), int64(1),
		"Stored time should be within 1 microsecond of original")

	// Query by exact timestamp
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM precision_test WHERE time = $1", storedTime).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should find exact timestamp match")

	// Cleanup
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS precision_test")
	require.NoError(t, err)
}
