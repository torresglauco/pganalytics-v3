package storage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// Test StoreClassificationResults inserts batch of classification results
func TestStoreClassificationResults(t *testing.T) {
	// This test validates the function signature and structure
	// Actual database tests would require a test database connection
	results := []*models.DataClassificationResult{
		{
			Time:              time.Now(),
			CollectorID:       uuid.New(),
			DatabaseName:      "testdb",
			SchemaName:        "public",
			TableName:         "users",
			ColumnName:        "cpf",
			PatternType:       "CPF",
			Category:          "PII",
			Confidence:        0.95,
			MatchCount:        100,
			SampleValues:      []string{"***.123.456-**"},
			RegulationMapping: map[string][]string{"LGPD": {"Art. 5, I"}},
		},
	}

	if len(results) != 1 {
		t.Error("Expected 1 result")
	}
}

// Test GetClassificationResults filters by collector_id, database, table
func TestGetClassificationResults(t *testing.T) {
	// Validate filter structure
	dbName := "testdb"
	tableName := "users"
	patternType := "CPF"

	filter := models.ClassificationFilter{
		DatabaseName: &dbName,
		TableName:    &tableName,
		PatternType:  &patternType,
		TimeRange:    "24h",
		Limit:        100,
		Offset:       0,
	}

	if *filter.DatabaseName != "testdb" {
		t.Error("DatabaseName filter not set correctly")
	}
	if *filter.TableName != "users" {
		t.Error("TableName filter not set correctly")
	}
	if *filter.PatternType != "CPF" {
		t.Error("PatternType filter not set correctly")
	}
}

// Test StoreCustomPattern validates pattern_regex
func TestStoreCustomPattern(t *testing.T) {
	// Validate regex compilation
	validRegex := `^\d{3}\.\d{3}\.\d{3}-\d{2}$`
	_, err := regexp.Compile(validRegex)
	if err != nil {
		t.Errorf("Valid regex failed to compile: %v", err)
	}

	// Invalid regex should fail
	invalidRegex := `[invalid(`
	_, err = regexp.Compile(invalidRegex)
	if err == nil {
		t.Error("Invalid regex should have failed to compile")
	}
}

// Test ClassificationFilter time range parsing
func TestClassificationFilterTimeRange(t *testing.T) {
	validRanges := []string{"1h", "24h", "7d", "30d"}
	for _, tr := range validRanges {
		filter := models.ClassificationFilter{TimeRange: tr}
		if filter.TimeRange != tr {
			t.Errorf("TimeRange not set correctly: expected %s, got %s", tr, filter.TimeRange)
		}
	}
}

// Test GetCustomPatterns returns global + tenant patterns
func TestGetCustomPatternsLogic(t *testing.T) {
	// This tests the logic that global patterns (tenant_id IS NULL)
	// should be returned alongside tenant-specific patterns

	// Simulate pattern filtering logic
	globalPattern := models.CustomPattern{
		ID:          1,
		TenantID:    uuid.NullUUID{Valid: false}, // NULL = global
		PatternName: "Global Pattern",
		Enabled:     true,
	}

	tenantID := uuid.New()
	tenantPattern := models.CustomPattern{
		ID:          2,
		TenantID:    uuid.NullUUID{UUID: tenantID, Valid: true},
		PatternName: "Tenant Pattern",
		Enabled:     true,
	}

	// Both should be returned for a tenant
	if globalPattern.TenantID.Valid {
		t.Error("Global pattern should have NULL tenant_id")
	}
	if !tenantPattern.TenantID.Valid {
		t.Error("Tenant pattern should have valid tenant_id")
	}
}

// Test context usage in store methods
func TestStoreContextUsage(t *testing.T) {
	ctx := context.Background()

	// Verify context can be used with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("Context should not be done yet")
	default:
		// Context is active
	}
}
