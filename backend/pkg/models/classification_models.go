package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// DATA CLASSIFICATION MODELS (DATA-01, DATA-02, DATA-03, DATA-05)
// ============================================================================

// DataClassificationResult represents a detected sensitive data pattern in a column
// DATA-01: User can view PII detection results for sensitive data patterns
// DATA-02: User can view PCI detection results for credit card numbers
// DATA-03: User can view LGPD/GDPR regulated data identification
type DataClassificationResult struct {
	Time              time.Time           `json:"time" db:"time"`
	CollectorID       uuid.UUID           `json:"collector_id" db:"collector_id"`
	DatabaseName      string              `json:"database_name" db:"database_name"`
	SchemaName        string              `json:"schema_name" db:"schema_name"`
	TableName         string              `json:"table_name" db:"table_name"`
	ColumnName        string              `json:"column_name" db:"column_name"`
	PatternType       string              `json:"pattern_type" db:"pattern_type"`             // CPF, CNPJ, EMAIL, PHONE, CREDIT_CARD, CUSTOM
	Category          string              `json:"category" db:"category"`                     // PII, PCI, SENSITIVE, CUSTOM
	Confidence        float64             `json:"confidence" db:"confidence"`                 // 0.0 to 1.0
	MatchCount        int64               `json:"match_count" db:"match_count"`               // Number of matching rows
	SampleValues      []string            `json:"sample_values" db:"sample_values"`           // Masked sample values (up to 5)
	RegulationMapping map[string][]string `json:"regulation_mapping" db:"regulation_mapping"` // LGPD/GDPR article references
}

// CustomPattern represents a user-defined detection pattern for custom sensitive data
// DATA-04: User can configure custom detection patterns
type CustomPattern struct {
	ID                  int           `json:"id" db:"id"`
	TenantID            uuid.NullUUID `json:"tenant_id" db:"tenant_id"` // NULL for global patterns
	PatternName         string        `json:"pattern_name" db:"pattern_name"`
	PatternRegex        string        `json:"pattern_regex" db:"pattern_regex"`
	Category            string        `json:"category" db:"category"`                         // PII, PCI, SENSITIVE, CUSTOM
	ValidationAlgorithm string        `json:"validation_algorithm" db:"validation_algorithm"` // Luhn, Mod11, None
	Description         string        `json:"description" db:"description"`
	Enabled             bool          `json:"enabled" db:"enabled"`
	CreatedAt           *time.Time    `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt           *time.Time    `json:"updated_at,omitempty" db:"updated_at"`
}

// ClassificationReportResponse contains aggregated classification results
// DATA-05: User can view data classification reports by database/table
type ClassificationReportResponse struct {
	TotalDatabases    int                             `json:"total_databases"`
	TotalTables       int                             `json:"total_tables"`
	TotalColumns      int                             `json:"total_columns"`
	PiiColumns        int                             `json:"pii_columns"`
	PciColumns        int                             `json:"pci_columns"`
	SensitiveColumns  int                             `json:"sensitive_columns"`
	CustomColumns     int                             `json:"custom_columns"`
	PatternBreakdown  map[string]int64                `json:"pattern_breakdown"`  // Pattern type -> count
	CategoryBreakdown map[string]int64                `json:"category_breakdown"` // Category -> count
	DatabaseSummary   []DatabaseClassificationSummary `json:"database_summary,omitempty"`
}

// DatabaseClassificationSummary contains classification summary for a single database
type DatabaseClassificationSummary struct {
	DatabaseName     string `json:"database_name"`
	SchemaCount      int    `json:"schema_count"`
	TableCount       int    `json:"table_count"`
	PiiColumnCount   int    `json:"pii_column_count"`
	PciColumnCount   int    `json:"pci_column_count"`
	SensitiveCount   int    `json:"sensitive_count"`
	HighestRiskTable string `json:"highest_risk_table,omitempty"`
}

// ClassificationFilter represents query filters for classification results
type ClassificationFilter struct {
	DatabaseName *string `json:"database_name,omitempty"`
	SchemaName   *string `json:"schema_name,omitempty"`
	TableName    *string `json:"table_name,omitempty"`
	PatternType  *string `json:"pattern_type,omitempty"`
	Category     *string `json:"category,omitempty"`
	TimeRange    string  `json:"time_range"` // 1h, 24h, 7d, 30d
	Limit        int     `json:"limit"`
	Offset       int     `json:"offset"`
}

// ClassificationMetricsResponse contains classification data with metadata
type ClassificationMetricsResponse struct {
	MetricType string                      `json:"metric_type"`
	Count      int                         `json:"count"`
	TimeRange  string                      `json:"time_range"`
	Data       []*DataClassificationResult `json:"data"`
}

// CustomPatternResponse contains custom patterns with metadata
type CustomPatternResponse struct {
	Count    int              `json:"count"`
	Patterns []*CustomPattern `json:"patterns"`
}
