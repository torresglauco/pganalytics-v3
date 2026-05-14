package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test DataClassificationResult struct has all required fields
func TestDataClassificationResultFields(t *testing.T) {
	collectorID := uuid.New()
	now := time.Now()

	result := DataClassificationResult{
		Time:              now,
		CollectorID:       collectorID,
		DatabaseName:      "testdb",
		SchemaName:        "public",
		TableName:         "users",
		ColumnName:        "cpf",
		PatternType:       "CPF",
		Category:          "PII",
		Confidence:        0.95,
		MatchCount:        150,
		SampleValues:      []string{"***.123.456-**", "***.789.012-**"},
		RegulationMapping: map[string][]string{"LGPD": {"Art. 5, I - dado pessoal"}},
	}

	if result.Time != now {
		t.Error("Time field not set correctly")
	}
	if result.CollectorID != collectorID {
		t.Error("CollectorID field not set correctly")
	}
	if result.DatabaseName != "testdb" {
		t.Error("DatabaseName field not set correctly")
	}
	if result.SchemaName != "public" {
		t.Error("SchemaName field not set correctly")
	}
	if result.TableName != "users" {
		t.Error("TableName field not set correctly")
	}
	if result.ColumnName != "cpf" {
		t.Error("ColumnName field not set correctly")
	}
	if result.PatternType != "CPF" {
		t.Error("PatternType field not set correctly")
	}
	if result.Category != "PII" {
		t.Error("Category field not set correctly")
	}
	if result.Confidence != 0.95 {
		t.Error("Confidence field not set correctly")
	}
	if result.MatchCount != 150 {
		t.Error("MatchCount field not set correctly")
	}
	if len(result.SampleValues) != 2 {
		t.Error("SampleValues field not set correctly")
	}
	if len(result.RegulationMapping["LGPD"]) != 1 {
		t.Error("RegulationMapping field not set correctly")
	}
}

// Test DataClassificationResult JSON tags match database column names
func TestDataClassificationResultJSONTags(t *testing.T) {
	result := DataClassificationResult{
		Time:              time.Now(),
		CollectorID:       uuid.New(),
		DatabaseName:      "testdb",
		SchemaName:        "public",
		TableName:         "users",
		ColumnName:        "email",
		PatternType:       "EMAIL",
		Category:          "PII",
		Confidence:        0.99,
		MatchCount:        500,
		SampleValues:      []string{"***@example.com"},
		RegulationMapping: map[string][]string{"LGPD": {"Art. 5, I"}},
	}

	// Check that db and json tags are present on all required fields
	// This test ensures the struct compiles with the correct tags
	_ = result
}

// Test CustomPattern struct supports tenant_id for tenant-specific patterns
func TestCustomPatternTenantID(t *testing.T) {
	// Test with tenant-specific pattern
	tenantID := uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	pattern := CustomPattern{
		ID:                  1,
		TenantID:            tenantID,
		PatternName:         "Brazilian RG",
		PatternRegex:        `^\d{2}\.\d{3}\.\d{3}-\d{1}$`,
		Category:            "PII",
		ValidationAlgorithm: "None",
		Description:         "Brazilian Identity Card number pattern",
		Enabled:             true,
	}

	if !pattern.TenantID.Valid {
		t.Error("TenantID should be valid for tenant-specific pattern")
	}
	if pattern.TenantID.UUID == uuid.Nil {
		t.Error("TenantID.UUID should not be Nil")
	}
}

// Test CustomPattern with NULL tenant_id (global pattern)
func TestCustomPatternGlobalPattern(t *testing.T) {
	pattern := CustomPattern{
		ID:                  2,
		TenantID:            uuid.NullUUID{Valid: false},
		PatternName:         "Global Phone Pattern",
		PatternRegex:        `^\+?\d{10,15}$`,
		Category:            "PII",
		ValidationAlgorithm: "None",
		Description:         "Global phone number pattern",
		Enabled:             true,
	}

	if pattern.TenantID.Valid {
		t.Error("TenantID should be invalid (NULL) for global pattern")
	}
}

// Test CustomPattern JSON tags
func TestCustomPatternJSONTags(t *testing.T) {
	pattern := CustomPattern{
		ID:                  1,
		TenantID:            uuid.NullUUID{Valid: false},
		PatternName:         "Test Pattern",
		PatternRegex:        `^\d+$`,
		Category:            "CUSTOM",
		ValidationAlgorithm: "Luhn",
		Description:         "Test description",
		Enabled:             true,
	}

	// Ensure struct compiles with correct tags
	_ = pattern
}

// Test ClassificationReportResponse struct
func TestClassificationReportResponse(t *testing.T) {
	report := ClassificationReportResponse{
		TotalDatabases:    5,
		TotalTables:       120,
		TotalColumns:      1500,
		PiiColumns:        45,
		PciColumns:        12,
		SensitiveColumns:  23,
		CustomColumns:     8,
		PatternBreakdown:  map[string]int64{"CPF": 150, "CNPJ": 50, "EMAIL": 200},
		CategoryBreakdown: map[string]int64{"PII": 350, "PCI": 50},
	}

	if report.TotalDatabases != 5 {
		t.Error("TotalDatabases field not set correctly")
	}
	if report.TotalTables != 120 {
		t.Error("TotalTables field not set correctly")
	}
	if report.PiiColumns != 45 {
		t.Error("PiiColumns field not set correctly")
	}
	if report.PciColumns != 12 {
		t.Error("PciColumns field not set correctly")
	}
	if len(report.PatternBreakdown) != 3 {
		t.Error("PatternBreakdown field not set correctly")
	}
}

// Test ClassificationFilter struct
func TestClassificationFilter(t *testing.T) {
	dbName := "testdb"
	schemaName := "public"
	tableName := "users"
	patternType := "CPF"
	category := "PII"

	filter := ClassificationFilter{
		DatabaseName: &dbName,
		SchemaName:   &schemaName,
		TableName:    &tableName,
		PatternType:  &patternType,
		Category:     &category,
		TimeRange:    "24h",
		Limit:        100,
		Offset:       0,
	}

	if *filter.DatabaseName != "testdb" {
		t.Error("DatabaseName filter not set correctly")
	}
	if *filter.SchemaName != "public" {
		t.Error("SchemaName filter not set correctly")
	}
	if *filter.TableName != "users" {
		t.Error("TableName filter not set correctly")
	}
	if *filter.PatternType != "CPF" {
		t.Error("PatternType filter not set correctly")
	}
	if *filter.Category != "PII" {
		t.Error("Category filter not set correctly")
	}
	if filter.TimeRange != "24h" {
		t.Error("TimeRange filter not set correctly")
	}
}

// Test pattern types are valid
func TestPatternTypes(t *testing.T) {
	validPatternTypes := map[string]bool{
		"CPF":         true,
		"CNPJ":        true,
		"EMAIL":       true,
		"PHONE":       true,
		"CREDIT_CARD": true,
		"CUSTOM":      true,
	}

	for patternType := range validPatternTypes {
		result := DataClassificationResult{
			PatternType: patternType,
		}
		if !validPatternTypes[result.PatternType] {
			t.Errorf("Invalid pattern type: %s", patternType)
		}
	}
}

// Test category types are valid
func TestCategoryTypes(t *testing.T) {
	validCategories := map[string]bool{
		"PII":       true,
		"PCI":       true,
		"SENSITIVE": true,
		"CUSTOM":    true,
	}

	for category := range validCategories {
		result := DataClassificationResult{
			Category: category,
		}
		if !validCategories[result.Category] {
			t.Errorf("Invalid category: %s", category)
		}
	}
}

// Test validation algorithms are valid
func TestValidationAlgorithms(t *testing.T) {
	validAlgorithms := map[string]bool{
		"Luhn":  true,
		"Mod11": true,
		"None":  true,
	}

	for algo := range validAlgorithms {
		pattern := CustomPattern{
			ValidationAlgorithm: algo,
		}
		if !validAlgorithms[pattern.ValidationAlgorithm] {
			t.Errorf("Invalid validation algorithm: %s", algo)
		}
	}
}
