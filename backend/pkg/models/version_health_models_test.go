package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test VersionHealthCheck struct has all required fields
func TestVersionHealthCheckFields(t *testing.T) {
	check := VersionHealthCheck{
		ID:             1,
		MinVersion:     11,
		MaxVersion:     12,
		CheckName:      "version_eol_warning",
		CheckQuery:     "SELECT current_setting('server_version_num')::int",
		ExpectedResult: "Version is End of Life",
		Severity:       "critical",
		Description:    "PostgreSQL 11-12 are End of Life versions",
		Remediation:    "Upgrade to PostgreSQL 15 or 16 for continued security updates",
		Category:       "security",
	}

	if check.ID != 1 {
		t.Error("ID field not set correctly")
	}
	if check.MinVersion != 11 {
		t.Error("MinVersion field not set correctly")
	}
	if check.MaxVersion != 12 {
		t.Error("MaxVersion field not set correctly")
	}
	if check.CheckName != "version_eol_warning" {
		t.Error("CheckName field not set correctly")
	}
	if check.CheckQuery != "SELECT current_setting('server_version_num')::int" {
		t.Error("CheckQuery field not set correctly")
	}
	if check.ExpectedResult != "Version is End of Life" {
		t.Error("ExpectedResult field not set correctly")
	}
	if check.Severity != "critical" {
		t.Error("Severity field not set correctly")
	}
	if check.Description != "PostgreSQL 11-12 are End of Life versions" {
		t.Error("Description field not set correctly")
	}
	if check.Remediation != "Upgrade to PostgreSQL 15 or 16 for continued security updates" {
		t.Error("Remediation field not set correctly")
	}
	if check.Category != "security" {
		t.Error("Category field not set correctly")
	}
}

// Test VersionHealthCheck with NULL max_version (no upper limit)
func TestVersionHealthCheckNoUpperLimit(t *testing.T) {
	check := VersionHealthCheck{
		ID:             5,
		MinVersion:     13,
		MaxVersion:     0, // 0 represents NULL (no upper limit)
		CheckName:      "wal_keep_size",
		CheckQuery:     "SELECT setting FROM pg_settings WHERE name = 'wal_keep_size'",
		ExpectedResult: "WAL retention size in MB",
		Severity:       "warning",
		Description:    "Check WAL retention size for replication",
		Remediation:    "Ensure sufficient WAL retention for standbys",
		Category:       "configuration",
	}

	if check.MinVersion != 13 {
		t.Error("MinVersion should be 13")
	}
	// MaxVersion of 0 indicates no upper limit
	if check.MaxVersion != 0 {
		t.Error("MaxVersion should be 0 (no upper limit)")
	}
}

// Test HealthCheckResult struct has all required fields
func TestHealthCheckResultFields(t *testing.T) {
	now := time.Now()

	result := HealthCheckResult{
		CheckID:        1,
		CheckName:      "version_eol_warning",
		Severity:       "critical",
		Passed:         false,
		ActualResult:   "110000",
		ExpectedResult: "Version is End of Life",
		Message:        "PostgreSQL version is End of Life",
		Remediation:    "Upgrade to PostgreSQL 15 or 16",
		CheckedAt:      now,
	}

	if result.CheckID != 1 {
		t.Error("CheckID field not set correctly")
	}
	if result.CheckName != "version_eol_warning" {
		t.Error("CheckName field not set correctly")
	}
	if result.Severity != "critical" {
		t.Error("Severity field not set correctly")
	}
	if result.Passed != false {
		t.Error("Passed field not set correctly")
	}
	if result.ActualResult != "110000" {
		t.Error("ActualResult field not set correctly")
	}
	if result.ExpectedResult != "Version is End of Life" {
		t.Error("ExpectedResult field not set correctly")
	}
	if result.Message != "PostgreSQL version is End of Life" {
		t.Error("Message field not set correctly")
	}
	if result.Remediation != "Upgrade to PostgreSQL 15 or 16" {
		t.Error("Remediation field not set correctly")
	}
	if result.CheckedAt != now {
		t.Error("CheckedAt field not set correctly")
	}
}

// Test HealthCheckResult passed scenario
func TestHealthCheckResultPassed(t *testing.T) {
	result := HealthCheckResult{
		CheckID:        4,
		CheckName:      "wal_compression",
		Severity:       "info",
		Passed:         true,
		ActualResult:   "on",
		ExpectedResult: "on or lz4 or zstd",
		Message:        "WAL compression is enabled",
		Remediation:    "Enable wal_compression for better performance",
		CheckedAt:      time.Now(),
	}

	if !result.Passed {
		t.Error("Passed should be true for successful check")
	}
	if result.Severity != "info" {
		t.Error("Severity should be info for passed check")
	}
}

// Test VersionHealthCheckResponse struct includes collector version info
func TestVersionHealthCheckResponseFields(t *testing.T) {
	collectorID := uuid.New()
	now := time.Now()

	response := VersionHealthCheckResponse{
		CollectorID:             collectorID,
		PostgreSQLVersion:       14,
		PostgreSQLVersionString: "14.2",
		Results: []*HealthCheckResult{
			{
				CheckID:   1,
				CheckName: "test_check",
				Passed:    true,
				CheckedAt: now,
			},
		},
		Summary: HealthCheckSummary{
			TotalChecks:    1,
			PassedChecks:   1,
			FailedCritical: 0,
			FailedWarning:  0,
			FailedInfo:     0,
		},
	}

	if response.CollectorID != collectorID {
		t.Error("CollectorID field not set correctly")
	}
	if response.PostgreSQLVersion != 14 {
		t.Error("PostgreSQLVersion field not set correctly")
	}
	if response.PostgreSQLVersionString != "14.2" {
		t.Error("PostgreSQLVersionString field not set correctly")
	}
	if len(response.Results) != 1 {
		t.Error("Results slice should have 1 element")
	}
	if response.Summary.TotalChecks != 1 {
		t.Error("Summary TotalChecks not set correctly")
	}
	if response.Summary.PassedChecks != 1 {
		t.Error("Summary PassedChecks not set correctly")
	}
}

// Test HealthCheckSummary struct fields
func TestHealthCheckSummaryFields(t *testing.T) {
	summary := HealthCheckSummary{
		TotalChecks:    10,
		PassedChecks:   7,
		FailedCritical: 1,
		FailedWarning:  1,
		FailedInfo:     1,
	}

	if summary.TotalChecks != 10 {
		t.Error("TotalChecks field not set correctly")
	}
	if summary.PassedChecks != 7 {
		t.Error("PassedChecks field not set correctly")
	}
	if summary.FailedCritical != 1 {
		t.Error("FailedCritical field not set correctly")
	}
	if summary.FailedWarning != 1 {
		t.Error("FailedWarning field not set correctly")
	}
	if summary.FailedInfo != 1 {
		t.Error("FailedInfo field not set correctly")
	}
}

// Test severity values are valid
func TestSeverityValues(t *testing.T) {
	validSeverities := map[string]bool{
		"critical": true,
		"warning":  true,
		"info":     true,
	}

	for severity := range validSeverities {
		check := VersionHealthCheck{
			Severity: severity,
		}
		if !validSeverities[check.Severity] {
			t.Errorf("Invalid severity: %s", severity)
		}
	}
}

// Test category values are valid
func TestHealthCheckCategoryValues(t *testing.T) {
	validCategories := map[string]bool{
		"performance":   true,
		"security":      true,
		"configuration": true,
		"replication":   true,
		"monitoring":    true,
	}

	for category := range validCategories {
		check := VersionHealthCheck{
			Category: category,
		}
		if !validCategories[check.Category] {
			t.Errorf("Invalid category: %s", category)
		}
	}
}

// Test VersionHealthCheckResponse with mixed results
func TestVersionHealthCheckResponseMixedResults(t *testing.T) {
	response := VersionHealthCheckResponse{
		CollectorID:             uuid.New(),
		PostgreSQLVersion:       12,
		PostgreSQLVersionString: "12.15",
		Results: []*HealthCheckResult{
			{CheckID: 1, CheckName: "eol_warning", Severity: "critical", Passed: false},
			{CheckID: 2, CheckName: "wal_segments", Severity: "warning", Passed: false},
			{CheckID: 3, CheckName: "replication_check", Severity: "info", Passed: true},
		},
		Summary: HealthCheckSummary{
			TotalChecks:    3,
			PassedChecks:   1,
			FailedCritical: 1,
			FailedWarning:  1,
			FailedInfo:     0,
		},
	}

	if response.PostgreSQLVersion != 12 {
		t.Error("PostgreSQLVersion should be 12")
	}
	if response.Summary.FailedCritical != 1 {
		t.Error("Should have 1 failed critical check")
	}
	if response.Summary.FailedWarning != 1 {
		t.Error("Should have 1 failed warning check")
	}
	if response.Summary.PassedChecks != 1 {
		t.Error("Should have 1 passed check")
	}
}
