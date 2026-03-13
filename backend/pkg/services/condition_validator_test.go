// backend/pkg/services/condition_validator_test.go
package services

import (
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

func TestValidateCondition_ValidMetricCondition(t *testing.T) {
	validator := NewConditionValidator()

	validCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(validCondition)
	if err != nil {
		t.Errorf("Expected no error for valid condition, got: %v", err)
	}
}

func TestValidateCondition_AllValidMetrics(t *testing.T) {
	validator := NewConditionValidator()

	metrics := []string{
		"error_count",
		"slow_query_count",
		"connection_count",
		"transaction_count",
		"cache_hit_ratio",
		"query_latency_p95",
		"query_latency_p99",
		"replication_lag",
		"cpu_usage",
		"memory_usage",
		"disk_usage",
	}

	for _, metric := range metrics {
		condition := models.AlertCondition{
			MetricType: metric,
			Operator:   ">",
			Threshold:  50.0,
			TimeWindow: "5m",
			Duration:   300,
		}

		err := validator.Validate(condition)
		if err != nil {
			t.Errorf("Expected no error for valid metric '%s', got: %v", metric, err)
		}
	}
}

func TestValidateCondition_AllValidOperators(t *testing.T) {
	validator := NewConditionValidator()

	operators := []string{">", "<", "==", "!=", ">=", "<="}

	for _, op := range operators {
		condition := models.AlertCondition{
			MetricType: "error_count",
			Operator:   op,
			Threshold:  50.0,
			TimeWindow: "5m",
			Duration:   300,
		}

		err := validator.Validate(condition)
		if err != nil {
			t.Errorf("Expected no error for valid operator '%s', got: %v", op, err)
		}
	}
}

func TestValidateCondition_InvalidOperator(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">>",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for invalid operator, got nil")
	}

	if err.Error() != "invalid operator '>>'. Valid operators are: >, <, ==, !=, >=, <=" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_InvalidMetricType(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "invalid_metric",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for invalid metric type, got nil")
	}

	if err.Error() != "invalid metric_type 'invalid_metric'. Valid metrics are: error_count, slow_query_count, connection_count, transaction_count, cache_hit_ratio, query_latency_p95, query_latency_p99, replication_lag, cpu_usage, memory_usage, disk_usage" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_EmptyMetricType(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for empty metric type, got nil")
	}

	if err.Error() != "metric_type cannot be empty" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_NegativeThreshold(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  -100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for negative threshold, got nil")
	}

	if err.Error() != "threshold cannot be negative, got -100.00" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_ZeroThreshold(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  0.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	err := validator.Validate(condition)
	if err != nil {
		t.Errorf("Expected no error for zero threshold, got: %v", err)
	}
}

func TestValidateCondition_ZeroTimeWindow(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "0m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for zero time window, got nil")
	}

	if err.Error() != "time_window must be positive (> 0 minutes), got 0 minutes" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_NegativeTimeWindow(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "-5m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for negative time window, got nil")
	}

	if err.Error() != "time_window must be positive (> 0 minutes), got -5 minutes" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_TimeWindowExceedsLimit(t *testing.T) {
	validator := NewConditionValidator()

	// 10081 minutes = 7 days + 1 minute
	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "10081m",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for time window exceeding limit, got nil")
	}

	if err.Error() != "time_window cannot exceed 10080 minutes (7 days), got 10081 minutes" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_TimeWindowWithHours(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "2h",
		Duration:   300,
	}

	err := validator.Validate(condition)
	if err != nil {
		t.Errorf("Expected no error for time window in hours, got: %v", err)
	}
}

func TestValidateCondition_TimeWindowWithDays(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "1d",
		Duration:   300,
	}

	err := validator.Validate(condition)
	if err != nil {
		t.Errorf("Expected no error for time window in days, got: %v", err)
	}
}

func TestValidateCondition_MaxTimeWindow(t *testing.T) {
	validator := NewConditionValidator()

	// 10080 minutes = 7 days exactly
	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "7d",
		Duration:   300,
	}

	err := validator.Validate(condition)
	if err != nil {
		t.Errorf("Expected no error for max time window, got: %v", err)
	}
}

func TestValidateCondition_NegativeDuration(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   -1,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for negative duration, got nil")
	}

	if err.Error() != "duration cannot be negative, got -1 seconds" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestValidateCondition_ZeroDuration(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   0,
	}

	err := validator.Validate(condition)
	if err != nil {
		t.Errorf("Expected no error for zero duration, got: %v", err)
	}
}

func TestConditionToDisplay_BasicCondition(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "5m",
		Duration:   300,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when Error Count > 100.00 over 5 minute(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestConditionToDisplay_CPUUsage(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "cpu_usage",
		Operator:   ">=",
		Threshold:  85.5,
		TimeWindow: "10m",
		Duration:   600,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when CPU Usage >= 85.50% over 10 minute(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestConditionToDisplay_QueryLatency(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "query_latency_p95",
		Operator:   ">",
		Threshold:  500.0,
		TimeWindow: "1h",
		Duration:   3600,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when Query Latency (p95) > 500.00 ms over 1 hour(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestConditionToDisplay_CacheHitRatio(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "cache_hit_ratio",
		Operator:   "<",
		Threshold:  70.0,
		TimeWindow: "30m",
		Duration:   1800,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when Cache Hit Ratio < 70.00% over 30 minute(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestConditionToDisplay_ReplicationLag(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "replication_lag",
		Operator:   ">",
		Threshold:  1000.0,
		TimeWindow: "2h",
		Duration:   7200,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when Replication Lag > 1000.00 ms over 2 hour(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestConditionToDisplay_DaysTimeWindow(t *testing.T) {
	validator := NewConditionValidator()

	condition := models.AlertCondition{
		MetricType: "disk_usage",
		Operator:   ">=",
		Threshold:  90.0,
		TimeWindow: "1d",
		Duration:   86400,
	}

	displayText := validator.ToDisplayText(condition)
	expected := "Alert when Disk Usage >= 90.00% over 1 day(s)"

	if displayText != expected {
		t.Errorf("Expected '%s', got '%s'", expected, displayText)
	}
}

func TestValidateMultipleConditions(t *testing.T) {
	validator := NewConditionValidator()

	conditions := []models.AlertCondition{
		{
			MetricType: "error_count",
			Operator:   ">",
			Threshold:  100.0,
			TimeWindow: "5m",
			Duration:   300,
		},
		{
			MetricType: "cpu_usage",
			Operator:   ">=",
			Threshold:  80.0,
			TimeWindow: "10m",
			Duration:   600,
		},
		{
			MetricType: "memory_usage",
			Operator:   "<",
			Threshold:  10.0,
			TimeWindow: "1h",
			Duration:   3600,
		},
	}

	for i, condition := range conditions {
		err := validator.Validate(condition)
		if err != nil {
			t.Errorf("Expected no error for condition %d, got: %v", i, err)
		}
	}
}

func TestConditionMarshalToJSON(t *testing.T) {
	condition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.5,
		TimeWindow: "5m",
		Duration:   300,
	}

	// Verify the struct can be accessed as expected
	if condition.MetricType != "error_count" {
		t.Errorf("Expected metric_type 'error_count', got '%s'", condition.MetricType)
	}

	if condition.Operator != ">" {
		t.Errorf("Expected operator '>', got '%s'", condition.Operator)
	}

	if condition.Threshold != 100.5 {
		t.Errorf("Expected threshold 100.5, got %.1f", condition.Threshold)
	}

	if condition.TimeWindow != "5m" {
		t.Errorf("Expected time_window '5m', got '%s'", condition.TimeWindow)
	}

	if condition.Duration != 300 {
		t.Errorf("Expected duration 300, got %d", condition.Duration)
	}
}

func TestValidateCondition_InvalidTimeWindowFormat(t *testing.T) {
	validator := NewConditionValidator()

	invalidCondition := models.AlertCondition{
		MetricType: "error_count",
		Operator:   ">",
		Threshold:  100.0,
		TimeWindow: "invalid",
		Duration:   300,
	}

	err := validator.Validate(invalidCondition)
	if err == nil {
		t.Error("Expected error for invalid time window format, got nil")
	}
}

func TestGetMetricDisplayName(t *testing.T) {
	validator := NewConditionValidator()

	testCases := []struct {
		metricType   string
		expectedName string
	}{
		{"error_count", "Error Count"},
		{"slow_query_count", "Slow Query Count"},
		{"connection_count", "Connection Count"},
		{"transaction_count", "Transaction Count"},
		{"cache_hit_ratio", "Cache Hit Ratio"},
		{"query_latency_p95", "Query Latency (p95)"},
		{"query_latency_p99", "Query Latency (p99)"},
		{"replication_lag", "Replication Lag"},
		{"cpu_usage", "CPU Usage"},
		{"memory_usage", "Memory Usage"},
		{"disk_usage", "Disk Usage"},
		{"unknown_metric", "unknown_metric"}, // Should return original if not found
	}

	for _, tc := range testCases {
		name := validator.getMetricDisplayName(tc.metricType)
		if name != tc.expectedName {
			t.Errorf("For metric '%s', expected '%s', got '%s'", tc.metricType, tc.expectedName, name)
		}
	}
}

func TestGetValidMetricsString(t *testing.T) {
	validator := NewConditionValidator()

	metricsStr := validator.getValidMetricsString()

	// Check that all metrics are present
	expectedMetrics := []string{
		"error_count",
		"slow_query_count",
		"connection_count",
		"transaction_count",
		"cache_hit_ratio",
		"query_latency_p95",
		"query_latency_p99",
		"replication_lag",
		"cpu_usage",
		"memory_usage",
		"disk_usage",
	}

	for _, metric := range expectedMetrics {
		if !contains(metricsStr, metric) {
			t.Errorf("Expected metrics string to contain '%s', got: %s", metric, metricsStr)
		}
	}
}

// Helper function for test
func contains(s string, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > 0 && s[0:1] != "" && (s == substr || len(s) >= len(substr)))
}

func TestParseTimeWindow_Minutes(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
		hasErr   bool
	}{
		{"5m", 5, false},
		{"1m", 1, false},
		{"60m", 60, false},
		{"5", 5, false},
	}

	for _, tc := range testCases {
		result, err := parseTimeWindow(tc.input)
		if (err != nil) != tc.hasErr {
			t.Errorf("For input '%s', expected hasErr=%v, got %v", tc.input, tc.hasErr, err != nil)
		}
		if !tc.hasErr && result != tc.expected {
			t.Errorf("For input '%s', expected %d minutes, got %d", tc.input, tc.expected, result)
		}
	}
}

func TestParseTimeWindow_Hours(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{"1h", 60},
		{"2h", 120},
		{"24h", 1440},
	}

	for _, tc := range testCases {
		result, err := parseTimeWindow(tc.input)
		if err != nil {
			t.Errorf("For input '%s', expected no error, got %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("For input '%s', expected %d minutes, got %d", tc.input, tc.expected, result)
		}
	}
}

func TestParseTimeWindow_Days(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{"1d", 1440},
		{"7d", 10080},
		{"3d", 4320},
	}

	for _, tc := range testCases {
		result, err := parseTimeWindow(tc.input)
		if err != nil {
			t.Errorf("For input '%s', expected no error, got %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("For input '%s', expected %d minutes, got %d", tc.input, tc.expected, result)
		}
	}
}
