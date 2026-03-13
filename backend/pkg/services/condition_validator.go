// backend/pkg/services/condition_validator.go
package services

import (
	"fmt"
	"strings"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ConditionValidator validates alert conditions before persistence
type ConditionValidator struct {
	validMetrics   map[string]bool
	validOperators map[string]bool
}

// NewConditionValidator creates a new condition validator with initialized valid values
func NewConditionValidator() *ConditionValidator {
	return &ConditionValidator{
		validMetrics: map[string]bool{
			"error_count":         true,
			"slow_query_count":    true,
			"connection_count":    true,
			"transaction_count":   true,
			"cache_hit_ratio":     true,
			"query_latency_p95":   true,
			"query_latency_p99":   true,
			"replication_lag":     true,
			"cpu_usage":           true,
			"memory_usage":        true,
			"disk_usage":          true,
		},
		validOperators: map[string]bool{
			">":  true,
			"<":  true,
			"==": true,
			"!=": true,
			">=": true,
			"<=": true,
		},
	}
}

// Validate checks if an AlertCondition is valid
func (cv *ConditionValidator) Validate(condition models.AlertCondition) error {
	// Check MetricType is not empty
	if strings.TrimSpace(condition.MetricType) == "" {
		return fmt.Errorf("metric_type cannot be empty")
	}

	// Check MetricType is in valid list
	if !cv.validMetrics[condition.MetricType] {
		validMetrics := cv.getValidMetricsString()
		return fmt.Errorf("invalid metric_type '%s'. Valid metrics are: %s", condition.MetricType, validMetrics)
	}

	// Check Operator is in valid list
	if !cv.validOperators[condition.Operator] {
		return fmt.Errorf("invalid operator '%s'. Valid operators are: >, <, ==, !=, >=, <=", condition.Operator)
	}

	// Check Threshold is non-negative
	if condition.Threshold < 0 {
		return fmt.Errorf("threshold cannot be negative, got %.2f", condition.Threshold)
	}

	// Check TimeWindow is valid - needs to be parsed and validated
	timeWindowMinutes, err := parseTimeWindow(condition.TimeWindow)
	if err != nil {
		return fmt.Errorf("invalid time_window format: %w", err)
	}

	if timeWindowMinutes <= 0 {
		return fmt.Errorf("time_window must be positive (> 0 minutes), got %d minutes", timeWindowMinutes)
	}

	// Check TimeWindow is <= 10080 minutes (7 days)
	const maxTimeWindowMinutes = 10080
	if timeWindowMinutes > maxTimeWindowMinutes {
		return fmt.Errorf("time_window cannot exceed 10080 minutes (7 days), got %d minutes", timeWindowMinutes)
	}

	// Check Duration is non-negative
	if condition.Duration < 0 {
		return fmt.Errorf("duration cannot be negative, got %d seconds", condition.Duration)
	}

	return nil
}

// ToDisplayText converts an AlertCondition to human-readable format
func (cv *ConditionValidator) ToDisplayText(condition models.AlertCondition) string {
	metricDisplay := cv.getMetricDisplayName(condition.MetricType)
	operatorDisplay := condition.Operator
	thresholdDisplay := formatThreshold(condition.Threshold, condition.MetricType)

	// Parse time window for display
	timeWindowMinutes, _ := parseTimeWindow(condition.TimeWindow)
	timeWindowDisplay := formatTimeWindow(timeWindowMinutes)

	return fmt.Sprintf("Alert when %s %s %s over %s", metricDisplay, operatorDisplay, thresholdDisplay, timeWindowDisplay)
}

// getMetricDisplayName returns a human-readable name for a metric type
func (cv *ConditionValidator) getMetricDisplayName(metricType string) string {
	displayNames := map[string]string{
		"error_count":       "Error Count",
		"slow_query_count":  "Slow Query Count",
		"connection_count":  "Connection Count",
		"transaction_count": "Transaction Count",
		"cache_hit_ratio":   "Cache Hit Ratio",
		"query_latency_p95": "Query Latency (p95)",
		"query_latency_p99": "Query Latency (p99)",
		"replication_lag":   "Replication Lag",
		"cpu_usage":         "CPU Usage",
		"memory_usage":      "Memory Usage",
		"disk_usage":        "Disk Usage",
	}

	if name, exists := displayNames[metricType]; exists {
		return name
	}
	return metricType
}

// getValidMetricsString returns a comma-separated list of valid metrics
func (cv *ConditionValidator) getValidMetricsString() string {
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
	return strings.Join(metrics, ", ")
}

// parseTimeWindow converts time window string to minutes
// Supports formats like "5m", "1h", "1d"
func parseTimeWindow(timeWindow string) (int, error) {
	timeWindow = strings.TrimSpace(timeWindow)

	if timeWindow == "" {
		return 0, fmt.Errorf("time_window cannot be empty")
	}

	// Remove unit suffix and parse the number
	var multiplier int
	var numStr string

	if strings.HasSuffix(timeWindow, "m") {
		multiplier = 1
		numStr = strings.TrimSuffix(timeWindow, "m")
	} else if strings.HasSuffix(timeWindow, "h") {
		multiplier = 60
		numStr = strings.TrimSuffix(timeWindow, "h")
	} else if strings.HasSuffix(timeWindow, "d") {
		multiplier = 1440 // 24 * 60
		numStr = strings.TrimSuffix(timeWindow, "d")
	} else {
		// Try parsing as a number (assumed to be in minutes)
		multiplier = 1
		numStr = timeWindow
	}

	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("time_window format invalid: no number found")
	}

	var minutes int
	_, err := fmt.Sscanf(numStr, "%d", &minutes)
	if err != nil {
		return 0, fmt.Errorf("time_window value must be an integer, got '%s'", numStr)
	}

	return minutes * multiplier, nil
}

// formatThreshold formats a threshold value with appropriate units based on metric type
func formatThreshold(threshold float64, metricType string) string {
	// Add units based on metric type
	switch metricType {
	case "cache_hit_ratio":
		return fmt.Sprintf("%.2f%%", threshold)
	case "cpu_usage", "memory_usage", "disk_usage":
		return fmt.Sprintf("%.2f%%", threshold)
	case "replication_lag":
		return fmt.Sprintf("%.2f ms", threshold)
	case "query_latency_p95", "query_latency_p99":
		return fmt.Sprintf("%.2f ms", threshold)
	default:
		return fmt.Sprintf("%.2f", threshold)
	}
}

// formatTimeWindow converts minutes to a readable time format
func formatTimeWindow(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%d minute(s)", minutes)
	} else if minutes < 1440 {
		hours := minutes / 60
		remainder := minutes % 60
		if remainder == 0 {
			return fmt.Sprintf("%d hour(s)", hours)
		}
		return fmt.Sprintf("%d hour(s) %d minute(s)", hours, remainder)
	} else {
		days := minutes / 1440
		remainder := minutes % 1440
		if remainder == 0 {
			return fmt.Sprintf("%d day(s)", days)
		}
		return fmt.Sprintf("%d day(s) %d minute(s)", days, remainder)
	}
}
