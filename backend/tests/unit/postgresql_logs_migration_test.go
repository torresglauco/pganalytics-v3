package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// TestPostgresqlLogsTableExists verifies the postgresql_logs table was created
func TestPostgresqlLogsTableExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'pganalytics' AND table_name = 'postgresql_logs'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}

	if !exists {
		t.Error("postgresql_logs table does not exist")
	}
}

// TestLogEventsHourlyTableExists verifies the log_events_hourly table was created
func TestLogEventsHourlyTableExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'pganalytics' AND table_name = 'log_events_hourly'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}

	if !exists {
		t.Error("log_events_hourly table does not exist")
	}
}

// TestLogStatsHourlyViewExists verifies the log_stats_hourly view was created
func TestLogStatsHourlyViewExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views
			WHERE table_schema = 'pganalytics' AND table_name = 'log_stats_hourly'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check view existence: %v", err)
	}

	if !exists {
		t.Error("log_stats_hourly view does not exist")
	}
}

// TestPostgresqlLogsIndexes verifies all expected indexes were created
func TestPostgresqlLogsIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expectedIndexes := []string{
		"idx_postgresql_logs_collector_timestamp",
		"idx_postgresql_logs_level_timestamp",
		"idx_postgresql_logs_instance_timestamp",
		"idx_postgresql_logs_database_timestamp",
	}

	for _, indexName := range expectedIndexes {
		var exists bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.statistics
				WHERE table_schema = 'pganalytics'
				AND table_name = 'postgresql_logs'
				AND index_name = $1
			)
		`, indexName).Scan(&exists)

		if err != nil {
			t.Fatalf("Failed to check index existence for %s: %v", indexName, err)
		}

		if !exists {
			t.Errorf("Expected index %s does not exist", indexName)
		}
	}
}

// TestInsertPostgresqlLog tests inserting a single log entry
func TestInsertPostgresqlLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create test data
	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	log := &models.PostgreSQLLog{
		CollectorID:    collectorID,
		InstanceID:     instanceID,
		DatabaseID:     nil,
		LogTimestamp:   time.Now(),
		LogLevel:       "INFO",
		LogMessage:     "Test log message",
		SourceLocation: stringPtr("test.c:100"),
		ProcessID:      intPtr(1234),
		QueryText:      stringPtr("SELECT 1"),
		QueryHash:      int64Ptr(12345),
		ErrorCode:      nil,
		UserName:       stringPtr("postgres"),
		ConnectionFrom: stringPtr("127.0.0.1:54321"),
		SessionID:      stringPtr("session-001"),
	}

	result, err := db.InsertPostgresqlLog(ctx, log)
	if err != nil {
		t.Fatalf("Failed to insert log: %v", err)
	}

	if result.ID == 0 {
		t.Error("Expected non-zero ID after insertion")
	}

	if result.CollectorID != collectorID {
		t.Errorf("Expected collector ID %s, got %s", collectorID, result.CollectorID)
	}

	if result.LogLevel != "INFO" {
		t.Errorf("Expected log level INFO, got %s", result.LogLevel)
	}

	if result.LogMessage != "Test log message" {
		t.Errorf("Expected log message 'Test log message', got '%s'", result.LogMessage)
	}
}

// TestInsertMultipleLogs tests inserting multiple log entries
func TestInsertMultipleLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	logLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
	var insertedLogs []*models.PostgreSQLLog

	for i, level := range logLevels {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now().Add(time.Duration(-i) * time.Hour),
			LogLevel:     level,
			LogMessage:   "Test message for " + level,
			UserName:     stringPtr("postgres"),
		}

		result, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert log level %s: %v", level, err)
		}

		insertedLogs = append(insertedLogs, result)
	}

	if len(insertedLogs) != len(logLevels) {
		t.Errorf("Expected %d inserted logs, got %d", len(logLevels), len(insertedLogs))
	}
}

// TestGetPostgresqlLogs tests retrieving logs
func TestGetPostgresqlLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	// Insert test logs
	for i := 0; i < 5; i++ {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now().Add(time.Duration(-i) * time.Hour),
			LogLevel:     "INFO",
			LogMessage:   "Test message",
		}

		_, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert log: %v", err)
		}
	}

	// Retrieve logs
	logs, err := db.GetPostgresqlLogs(ctx, instanceID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(logs))
	}

	// Verify ordering (DESC)
	for i := 0; i < len(logs)-1; i++ {
		if logs[i].LogTimestamp.Before(logs[i+1].LogTimestamp) {
			t.Error("Logs are not ordered by timestamp descending")
			break
		}
	}
}

// TestGetPostgresqlLogsByLevel tests filtering logs by level
func TestGetPostgresqlLogsByLevel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	// Insert logs with different levels
	for i := 0; i < 3; i++ {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now().Add(time.Duration(-i) * time.Hour),
			LogLevel:     "INFO",
			LogMessage:   "Info message",
		}
		_, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert INFO log: %v", err)
		}
	}

	for i := 0; i < 2; i++ {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now().Add(time.Duration(-(i + 3)) * time.Hour),
			LogLevel:     "ERROR",
			LogMessage:   "Error message",
		}
		_, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert ERROR log: %v", err)
		}
	}

	// Get ERROR logs
	errorLogs, err := db.GetPostgresqlLogsByLevel(ctx, instanceID, "ERROR", 10, 0)
	if err != nil {
		t.Fatalf("Failed to get error logs: %v", err)
	}

	if len(errorLogs) != 2 {
		t.Errorf("Expected 2 ERROR logs, got %d", len(errorLogs))
	}

	for _, log := range errorLogs {
		if log.LogLevel != "ERROR" {
			t.Errorf("Expected log level ERROR, got %s", log.LogLevel)
		}
	}
}

// TestGetErrorLogs tests retrieving critical error logs
func TestGetErrorLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	logLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL", "PANIC"}
	expectedErrorCount := 0

	for _, level := range logLevels {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now(),
			LogLevel:     level,
			LogMessage:   "Test message",
		}

		_, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert log: %v", err)
		}

		if level == "ERROR" || level == "FATAL" || level == "PANIC" {
			expectedErrorCount++
		}
	}

	errorLogs, err := db.GetErrorLogs(ctx, instanceID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get error logs: %v", err)
	}

	if len(errorLogs) != expectedErrorCount {
		t.Errorf("Expected %d error logs, got %d", expectedErrorCount, len(errorLogs))
	}

	for _, log := range errorLogs {
		if log.LogLevel != "ERROR" && log.LogLevel != "FATAL" && log.LogLevel != "PANIC" {
			t.Errorf("Expected critical error log, got level %s", log.LogLevel)
		}
	}
}

// TestLogPagination tests pagination of log retrieval
func TestLogPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	// Insert 25 test logs
	for i := 0; i < 25; i++ {
		log := &models.PostgreSQLLog{
			CollectorID:  collectorID,
			InstanceID:   instanceID,
			LogTimestamp: time.Now().Add(time.Duration(-i) * time.Hour),
			LogLevel:     "INFO",
			LogMessage:   "Test message",
		}

		_, err := db.InsertPostgresqlLog(ctx, log)
		if err != nil {
			t.Fatalf("Failed to insert log: %v", err)
		}
	}

	// Get first page (10 items)
	page1, err := db.GetPostgresqlLogs(ctx, instanceID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}

	if len(page1) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(page1))
	}

	// Get second page (10 items)
	page2, err := db.GetPostgresqlLogs(ctx, instanceID, 10, 10)
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}

	if len(page2) != 10 {
		t.Errorf("Expected 10 items on page 2, got %d", len(page2))
	}

	// Verify no overlap between pages
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Error("Found duplicate log ID between pages")
			}
		}
	}
}

// TestErrorCodeStorage tests storage of error codes and details
func TestErrorCodeStorage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	log := &models.PostgreSQLLog{
		CollectorID:  collectorID,
		InstanceID:   instanceID,
		LogTimestamp: time.Now(),
		LogLevel:     "ERROR",
		LogMessage:   "Test error",
		ErrorCode:    stringPtr("42P01"),
		ErrorDetail:  stringPtr("Relation not found"),
		ErrorHint:    stringPtr("Check table exists"),
		ErrorContext: stringPtr("During query execution"),
	}

	result, err := db.InsertPostgresqlLog(ctx, log)
	if err != nil {
		t.Fatalf("Failed to insert log with error details: %v", err)
	}

	if result.ErrorCode == nil || *result.ErrorCode != "42P01" {
		t.Error("Error code not stored correctly")
	}

	if result.ErrorDetail == nil || *result.ErrorDetail != "Relation not found" {
		t.Error("Error detail not stored correctly")
	}

	if result.ErrorHint == nil || *result.ErrorHint != "Check table exists" {
		t.Error("Error hint not stored correctly")
	}

	if result.ErrorContext == nil || *result.ErrorContext != "During query execution" {
		t.Error("Error context not stored correctly")
	}
}

// TestQueryMetadataStorage tests storage of query-related metadata
func TestQueryMetadataStorage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collectorID := uuid.New()
	instanceID := createTestInstance(t, db, collectorID)

	log := &models.PostgreSQLLog{
		CollectorID:    collectorID,
		InstanceID:     instanceID,
		LogTimestamp:   time.Now(),
		LogLevel:       "INFO",
		LogMessage:     "Query executed",
		QueryText:      stringPtr("SELECT * FROM users WHERE id = $1"),
		QueryHash:      int64Ptr(123456789),
		SourceLocation: stringPtr("backend.c:1234"),
		ProcessID:      intPtr(9999),
		UserName:       stringPtr("app_user"),
		ConnectionFrom: stringPtr("10.0.0.1:5432"),
		SessionID:      stringPtr("sess-abc123"),
	}

	result, err := db.InsertPostgresqlLog(ctx, log)
	if err != nil {
		t.Fatalf("Failed to insert query log: %v", err)
	}

	if result.QueryText == nil || *result.QueryText != "SELECT * FROM users WHERE id = $1" {
		t.Error("Query text not stored correctly")
	}

	if result.QueryHash == nil || *result.QueryHash != 123456789 {
		t.Error("Query hash not stored correctly")
	}

	if result.ProcessID == nil || *result.ProcessID != 9999 {
		t.Error("Process ID not stored correctly")
	}

	if result.SessionID == nil || *result.SessionID != "sess-abc123" {
		t.Error("Session ID not stored correctly")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// setupTestDB initializes test database connection
func setupTestDB(t *testing.T) *storage.PostgresDB {
	// This assumes we have a test database setup
	// In real environment, this would connect to test PostgreSQL instance
	connStr := getTestConnectionString()
	db, err := storage.NewPostgresDB(connStr)
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}
	return db
}

// getTestConnectionString returns test database connection string
func getTestConnectionString() string {
	// Use environment variable or default test connection
	// Format: postgres://user:password@localhost:5432/pganalytics_test
	return "postgres://postgres:postgres@localhost:5432/pganalytics_test?sslmode=disable"
}

// createTestInstance creates a test PostgreSQL instance for log storage
func createTestInstance(t *testing.T, db *storage.PostgresDB, collectorID uuid.UUID) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create test server first
	serverID := createTestServer(t, db)

	// Create test PostgreSQL instance
	var instanceID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.postgresql_instances (
			server_id, name, port, maintenance_database, monitoring_role, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, serverID, "test-instance", 5432, "postgres", "pganalytics", true).Scan(&instanceID)

	if err != nil {
		t.Skipf("Could not create test instance: %v", err)
	}

	return instanceID
}

// createTestServer creates a test server for log storage
func createTestServer(t *testing.T, db *storage.PostgresDB) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var serverID int
	err := db.QueryRowContext(ctx, `
		INSERT INTO pganalytics.servers (
			name, hostname, address, environment, is_active
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, "test-server", "localhost", "127.0.0.1", "test", true).Scan(&serverID)

	if err != nil {
		t.Skipf("Could not create test server: %v", err)
	}

	return serverID
}

// Helper pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
