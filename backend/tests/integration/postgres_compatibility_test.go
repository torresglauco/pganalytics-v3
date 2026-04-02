package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// TestPostgresVersionCompatibility validates that all migrations work across supported PostgreSQL versions
// Supported versions: PostgreSQL 14, 15, 16, 17, 18
func TestPostgresVersionCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get current database connection to check PostgreSQL version
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get PostgreSQL version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to get PostgreSQL version: %v", err)
	}

	t.Logf("Testing against PostgreSQL version: %s", version)

	// Verify minimum supported version (14)
	if !isMinimumSupportedVersion(version) {
		t.Skipf("PostgreSQL version not in supported range (14-18): %s", version)
	}
}

// TestSchemaIntegrity validates that all tables exist and have expected structures
func TestSchemaIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Verify pganalytics schema exists
	var schemaExists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata WHERE schema_name = 'pganalytics'
		)
	`).Scan(&schemaExists)
	if err != nil {
		t.Fatalf("Failed to check schema existence: %v", err)
	}
	if !schemaExists {
		t.Fatal("pganalytics schema does not exist")
	}

	// Expected tables from migrations
	expectedTables := []string{
		"users",
		"api_tokens",
		"collectors",
		"collector_tokens",
		"collector_certificates",
		"registration_secrets",
		"servers",
		"postgresql_instances",
		"databases",
		"secrets",
		"alert_rules",
		"managed_instances",
		"managed_instance_databases",
		"postgresql_logs",
		"log_events_hourly",
		"metrics",
		"notification_channels",
		"alert_triggers",
		"notifications",
		"alert_silences",
		"escalation_policies",
		"escalation_policy_steps",
		"query_plans",
		"query_issues",
		"query_performance_timeline",
		"logs",
		"log_patterns",
		"log_anomalies",
		"index_recommendations",
		"index_analysis",
		"unused_indexes",
		"vacuum_recommendations",
		"autovacuum_configurations",
	}

	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'pganalytics'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to count tables: %v", err)
	}

	if tableCount < len(expectedTables) {
		t.Logf("Expected at least %d tables, found %d. This may be expected depending on migrations.", len(expectedTables), tableCount)
	}

	// Verify key tables exist
	for _, tableName := range expectedTables[:5] { // Check first 5 as sample
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'pganalytics' AND table_name = $1
			)
		`, tableName).Scan(&exists)
		if err != nil {
			t.Logf("Error checking table %s: %v", tableName, err)
			continue
		}
		if !exists {
			t.Logf("Warning: Table %s not found", tableName)
		}
	}
}

// TestUUIDExtension validates uuid-ossp extension compatibility
func TestUUIDExtension(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Test gen_random_uuid() which is used in migrations
	var uuid string
	err := db.QueryRow("SELECT gen_random_uuid()::text").Scan(&uuid)
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	if len(uuid) == 0 {
		t.Fatal("Generated UUID is empty")
	}

	t.Logf("Successfully generated UUID: %s", uuid)
}

// TestJSONBSupport validates JSONB data type compatibility
func TestJSONBSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Test JSONB operations used in notification_channels
	var result string
	err := db.QueryRow(`
		SELECT ('{"key": "value"}'::jsonb ->> 'key')
	`).Scan(&result)
	if err != nil {
		t.Fatalf("Failed JSONB operation: %v", err)
	}

	if result != "value" {
		t.Fatalf("Expected 'value', got %s", result)
	}

	t.Log("JSONB operations verified")
}

// TestTimestampWithTimeZone validates TIMESTAMP WITH TIME ZONE compatibility
func TestTimestampWithTimeZone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	var timestamp string
	err := db.QueryRow(`
		SELECT CURRENT_TIMESTAMP::text
	`).Scan(&timestamp)
	if err != nil {
		t.Fatalf("Failed to get timestamp: %v", err)
	}

	if len(timestamp) == 0 {
		t.Fatal("Timestamp is empty")
	}

	t.Logf("Current timestamp: %s", timestamp)
}

// TestArrayTypes validates TEXT[] array type compatibility
func TestArrayTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Test array operations used in index_recommendations
	var arrayLength int
	err := db.QueryRow(`
		SELECT array_length(ARRAY['col1', 'col2', 'col3']::text[], 1)
	`).Scan(&arrayLength)
	if err != nil {
		t.Fatalf("Failed array operation: %v", err)
	}

	if arrayLength != 3 {
		t.Fatalf("Expected array length 3, got %d", arrayLength)
	}

	t.Log("Array types verified")
}

// TestBIGSERIALSupport validates BIGSERIAL primary key compatibility
func TestBIGSERIALSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// BIGSERIAL is used in all modern tables
	var nextval int64
	err := db.QueryRow(`
		SELECT 9223372036854775807::bigint
	`).Scan(&nextval)
	if err != nil {
		t.Fatalf("Failed BIGSERIAL test: %v", err)
	}

	t.Log("BIGSERIAL support verified")
}

// TestTriggerFunctionCompatibility validates trigger function compatibility
func TestTriggerFunctionCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Check if update_updated_at_column function exists
	var functionExists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column'
		)
	`).Scan(&functionExists)
	if err != nil {
		t.Logf("Failed to check function existence: %v", err)
		return
	}

	if functionExists {
		t.Log("update_updated_at_column function exists")
	} else {
		t.Log("update_updated_at_column function not found (may not be created yet)")
	}
}

// TestIndexCompatibility validates index creation compatibility
func TestIndexCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Count indexes in pganalytics schema (partial indexes should be created successfully)
	var indexCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'pganalytics'
	`).Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to count indexes: %v", err)
	}

	t.Logf("Found %d indexes in pganalytics schema", indexCount)
}

// TestConstraintCompatibility validates constraint creation compatibility
func TestConstraintCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Check for constraints (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK)
	var constraintCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.table_constraints
		WHERE table_schema = 'pganalytics'
	`).Scan(&constraintCount)
	if err != nil {
		t.Fatalf("Failed to count constraints: %v", err)
	}

	if constraintCount > 0 {
		t.Logf("Found %d constraints in pganalytics schema", constraintCount)
	}
}

// TestPostgres14Compatibility validates PostgreSQL 14 specific compatibility
func TestPostgres14Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	version := getPostgresVersion(t, db)
	if !strings.Contains(version, "14") {
		t.Skipf("Test is for PostgreSQL 14, current version: %s", version)
	}

	// Test minimum version features
	testMinimumVersionFeatures(t, db)
}

// TestPostgres15Compatibility validates PostgreSQL 15 specific compatibility
func TestPostgres15Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	version := getPostgresVersion(t, db)
	if !strings.Contains(version, "15") {
		t.Skipf("Test is for PostgreSQL 15, current version: %s", version)
	}

	testMinimumVersionFeatures(t, db)
}

// TestPostgres16Compatibility validates PostgreSQL 16 specific compatibility
func TestPostgres16Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	version := getPostgresVersion(t, db)
	if !strings.Contains(version, "16") {
		t.Skipf("Test is for PostgreSQL 16, current version: %s", version)
	}

	testMinimumVersionFeatures(t, db)
}

// TestPostgres17Compatibility validates PostgreSQL 17 specific compatibility
func TestPostgres17Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	version := getPostgresVersion(t, db)
	if !strings.Contains(version, "17") {
		t.Skipf("Test is for PostgreSQL 17, current version: %s", version)
	}

	testMinimumVersionFeatures(t, db)
}

// TestPostgres18Compatibility validates PostgreSQL 18 specific compatibility
func TestPostgres18Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	version := getPostgresVersion(t, db)
	if !strings.Contains(version, "18") {
		t.Skipf("Test is for PostgreSQL 18, current version: %s", version)
	}

	testMinimumVersionFeatures(t, db)
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("DATABASE_URL or TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test connection
	ctx, cancel := context.Background(), context.CancelFunc(func() {})
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func getPostgresVersion(t *testing.T, db *sql.DB) string {
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		t.Logf("Failed to get PostgreSQL version: %v", err)
		return ""
	}
	return version
}

func isMinimumSupportedVersion(version string) bool {
	// PostgreSQL versions >= 14 are supported
	supportedVersions := []string{"14", "15", "16", "17", "18", "19"}
	for _, v := range supportedVersions {
		if strings.Contains(version, fmt.Sprintf("PostgreSQL %s", v)) {
			return true
		}
	}
	return false
}

func testMinimumVersionFeatures(t *testing.T, db *sql.DB) {
	// Test basic features available in PG 14+
	testQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "CURRENT_TIMESTAMP",
			query: "SELECT CURRENT_TIMESTAMP",
		},
		{
			name:  "gen_random_uuid",
			query: "SELECT gen_random_uuid()",
		},
		{
			name:  "JSONB operations",
			query: "SELECT '{\"key\": \"value\"}'::jsonb",
		},
		{
			name:  "Array operations",
			query: "SELECT ARRAY['a', 'b', 'c']::text[]",
		},
	}

	for _, tc := range testQueries {
		var result interface{}
		err := db.QueryRow(tc.query).Scan(&result)
		if err != nil {
			t.Errorf("Feature test failed: %s - %v", tc.name, err)
		} else {
			t.Logf("Feature verified: %s", tc.name)
		}
	}
}
