package log_analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogParser_ClassifyLog(t *testing.T) {
	parser := NewLogParser()

	tests := []struct {
		message  string
		expected LogCategory
	}{
		{"FATAL: database mydb does not exist", CategoryDatabaseError},
		{"Connection refused", CategoryConnectionError},
		{"FATAL: no pg_hba.conf entry", CategoryAuthenticationError},
		{"ERROR: syntax error at or near", CategorySyntaxError},
		{"UNIQUE constraint violation", CategoryConstraintError},
		{"duration: 5000.123 ms", CategorySlowQuery},
		{"LOG: checkpoint", CategoryCheckpoint},
		{"automatic vacuum", CategoryVacuum},
		{"long transaction", CategoryLongTransaction},
		{"lock timeout", CategoryLockTimeout},
		{"deadlock detected", CategoryDeadlock},
		{"replication lag", CategoryReplicationError},
		{"WAL error", CategoryWALError},
		{"out of memory", CategoryOutOfMemory},
		{"No space left on device", CategoryDiskFull},
		{"WARNING: something", CategoryWarning},
	}

	for _, tt := range tests {
		category := parser.ClassifyLog(tt.message)
		assert.Equal(t, tt.expected, category, "Failed for message: %s", tt.message)
	}
}

func TestLogParser_ExtractMetadata(t *testing.T) {
	parser := NewLogParser()

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
		metadata := parser.ExtractMetadata(tt.message)
		tt.checks(t, metadata)
	}
}

func TestLogParser_DefaultCategories(t *testing.T) {
	parser := NewLogParser()

	tests := []struct {
		message  string
		expected LogCategory
	}{
		{"some error message", CategoryDatabaseError},
		{"some warning message", CategoryWarning},
		{"some info message", CategoryInfo},
	}

	for _, tt := range tests {
		category := parser.ClassifyLog(tt.message)
		assert.Equal(t, tt.expected, category, "Failed for message: %s", tt.message)
	}
}
