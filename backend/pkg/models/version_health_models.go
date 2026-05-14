package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// VERSION-SPECIFIC HEALTH CHECK MODELS (11-04)
// ============================================================================

// VersionHealthCheck represents a version-specific health check definition
type VersionHealthCheck struct {
	ID             int       `json:"id" db:"id"`
	MinVersion     int       `json:"min_version" db:"min_version"`         // Minimum PG version (e.g., 11)
	MaxVersion     int       `json:"max_version" db:"max_version"`         // Maximum PG version (0 = no upper limit)
	CheckName      string    `json:"check_name" db:"check_name"`           // Unique check identifier
	CheckQuery     string    `json:"check_query" db:"check_query"`         // SQL query to execute
	ExpectedResult string    `json:"expected_result" db:"expected_result"` // Description of expected result
	Severity       string    `json:"severity" db:"severity"`               // critical, warning, info
	Description    string    `json:"description" db:"description"`         // What this check does
	Remediation    string    `json:"remediation" db:"remediation"`         // How to fix issues
	Category       string    `json:"category" db:"category"`               // performance, security, configuration, replication, monitoring
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// HealthCheckResult represents the result of executing a version-specific health check
type HealthCheckResult struct {
	CheckID        int       `json:"check_id" db:"check_id"`
	CheckName      string    `json:"check_name" db:"check_name"`
	Severity       string    `json:"severity" db:"severity"`
	Passed         bool      `json:"passed" db:"passed"`
	ActualResult   string    `json:"actual_result" db:"actual_result"`
	ExpectedResult string    `json:"expected_result" db:"expected_result"`
	Message        string    `json:"message" db:"message"`
	Remediation    string    `json:"remediation" db:"remediation"`
	CheckedAt      time.Time `json:"checked_at" db:"checked_at"`
}

// HealthCheckSummary provides aggregate statistics for health check results
type HealthCheckSummary struct {
	TotalChecks    int `json:"total_checks" db:"total_checks"`
	PassedChecks   int `json:"passed_checks" db:"passed_checks"`
	FailedCritical int `json:"failed_critical" db:"failed_critical"`
	FailedWarning  int `json:"failed_warning" db:"failed_warning"`
	FailedInfo     int `json:"failed_info" db:"failed_info"`
}

// VersionHealthCheckResponse represents the complete response for version health checks
type VersionHealthCheckResponse struct {
	CollectorID             uuid.UUID            `json:"collector_id"`
	PostgreSQLVersion       int                  `json:"postgresql_version"`
	PostgreSQLVersionString string               `json:"postgresql_version_string"`
	Results                 []*HealthCheckResult `json:"results"`
	Summary                 HealthCheckSummary   `json:"summary"`
}
