package log_analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogCollector_IngestLogs(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{
		{
			"timestamp": "2026-03-30T10:00:00Z",
			"message":   "FATAL: database mydb does not exist",
			"severity":  "FATAL",
		},
	}

	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.NoError(t, err)
}

func TestLogCollector_IngestLogs_MultipleLogs(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{
		{
			"timestamp": "2026-03-30T10:00:00Z",
			"message":   "FATAL: database mydb does not exist",
			"severity":  "FATAL",
		},
		{
			"timestamp": "2026-03-30T10:01:00Z",
			"message":   "ERROR: syntax error at or near 'SELECT'",
			"severity":  "ERROR",
		},
		{
			"timestamp": "2026-03-30T10:02:00Z",
			"message":   "LOG: duration: 1234.567 ms",
			"severity":  "LOG",
		},
	}

	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.NoError(t, err)
}

func TestLogCollector_NewLogCollector(t *testing.T) {
	collector := NewLogCollector(nil)
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.parser)
	assert.Nil(t, collector.db)
}

func TestLogCollector_ClassifyLogs(t *testing.T) {
	collector := NewLogCollector(nil)

	tests := []struct {
		message  string
		expected LogCategory
	}{
		{"FATAL: database mydb does not exist", CategoryDatabaseError},
		{"Connection refused", CategoryConnectionError},
		{"ERROR: syntax error at or near", CategorySyntaxError},
		{"UNIQUE constraint violation", CategoryConstraintError},
		{"duration: 5000.123 ms", CategorySlowQuery},
	}

	for _, tt := range tests {
		category := collector.parser.ClassifyLog(tt.message)
		assert.Equal(t, tt.expected, category, "Failed for message: %s", tt.message)
	}
}

func TestLogCollector_ExtractMetadata(t *testing.T) {
	collector := NewLogCollector(nil)

	tests := []struct {
		message string
		checks  func(t *testing.T, metadata map[string]interface{})
	}{
		{
			message: "duration: 1234.567 ms",
			checks: func(t *testing.T, metadata map[string]interface{}) {
				assert.Equal(t, "1234.567", metadata["duration"])
			},
		},
		{
			message: "relation \"users\" not found",
			checks: func(t *testing.T, metadata map[string]interface{}) {
				assert.Equal(t, "users", metadata["table"])
			},
		},
		{
			message: "no metadata here",
			checks: func(t *testing.T, metadata map[string]interface{}) {
				assert.Empty(t, metadata)
			},
		},
	}

	for _, tt := range tests {
		metadata := collector.parser.ExtractMetadata(tt.message)
		tt.checks(t, metadata)
	}
}

func TestLogCollector_StreamLogs_ContextCancellation(t *testing.T) {
	collector := NewLogCollector(nil)
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan map[string]interface{})

	// Cancel immediately to test context handling
	cancel()

	// Should return context error
	err := collector.StreamLogs(ctx, "1", ch)
	assert.Equal(t, context.Canceled, err)
}

func TestLogCollector_StreamLogs_ContextTimeout(t *testing.T) {
	collector := NewLogCollector(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := make(chan map[string]interface{})

	// Should timeout when no database is available
	err := collector.StreamLogs(ctx, "1", ch)
	// Either timeout or database error is acceptable
	assert.NotNil(t, err)
}

func TestLogCollector_IngestLogs_WithNilDB(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{
		{
			"timestamp": "2026-03-30T10:00:00Z",
			"message":   "test log",
			"severity":  "INFO",
		},
	}

	// Should handle nil database gracefully by skipping inserts
	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.NoError(t, err)
}

func TestLogCollector_IngestLogs_InvalidTimestamp(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{
		{
			"timestamp": "invalid-timestamp",
			"message":   "test log",
			"severity":  "INFO",
		},
	}

	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.Error(t, err)
}

func TestLogCollector_IngestLogs_MissingFields(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{
		{
			"timestamp": "2026-03-30T10:00:00Z",
			// missing message and severity
		},
	}

	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.Error(t, err)
}

func TestLogCollector_IngestLogs_EmptyLogs(t *testing.T) {
	collector := NewLogCollector(nil)

	logs := []map[string]interface{}{}

	err := collector.IngestLogs(context.Background(), "1", logs)
	assert.NoError(t, err)
}

func TestLogCollector_Parser_Integration(t *testing.T) {
	collector := NewLogCollector(nil)

	logMessage := "FATAL: database mydb does not exist"
	category := collector.parser.ClassifyLog(logMessage)
	metadata := collector.parser.ExtractMetadata(logMessage)

	assert.Equal(t, CategoryDatabaseError, category)
	assert.Empty(t, metadata)
}

func TestLogCollector_Parser_WithMetadata(t *testing.T) {
	collector := NewLogCollector(nil)

	logMessage := "ERROR in relation \"products\": duration: 2500.456 ms"
	category := collector.parser.ClassifyLog(logMessage)
	metadata := collector.parser.ExtractMetadata(logMessage)

	assert.NotNil(t, category)
	assert.Contains(t, metadata, "table")
	assert.Equal(t, "products", metadata["table"])
	assert.Contains(t, metadata, "duration")
	assert.Equal(t, "2500.456", metadata["duration"])
}

func TestLogCollector_IngestLogs_ContextCancellation(t *testing.T) {
	collector := NewLogCollector(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logs := []map[string]interface{}{
		{
			"timestamp": "2026-03-30T10:00:00Z",
			"message":   "test log",
			"severity":  "INFO",
		},
	}

	err := collector.IngestLogs(ctx, "1", logs)
	// Should handle nil database gracefully by skipping inserts
	// Context cancellation only matters when database is available
	assert.NoError(t, err)
}
