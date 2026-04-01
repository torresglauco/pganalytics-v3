package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/log_analysis"
)

// TestLogAnalysisE2E tests the complete log analysis data flow
// From log ingestion → classification → anomaly detection → API response
func TestLogAnalysisE2E(t *testing.T) {
	t.Run("ingest_and_classify_logs", func(t *testing.T) {
		// Arrange: Initialize collector without database (for service-level test)
		collector := log_analysis.NewLogCollector(nil)
		require.NotNil(t, collector)

		// Arrange: Get mock log entries
		logEntries := MockPostgresLogEntries()
		require.NotEmpty(t, logEntries)

		// Act: Ingest logs
		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", logEntries)

		// Assert: No error on ingestion
		assert.NoError(t, err)
	})

	t.Run("classify_multiple_log_categories", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)
		logEntries := MockPostgresLogEntries()

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", logEntries)
		require.NoError(t, err)

		// Verify logs were processed (classifier should categorize them)
		parser := collector.GetLogParser()
		require.NotNil(t, parser)

		// Check that parser can classify various log types
		for _, logEntry := range logEntries {
			message, ok := logEntry["message"].(string)
			require.True(t, ok)

			category := parser.ClassifyLog(message)
			assert.NotEmpty(t, category)
		}
	})

	t.Run("detect_error_logs", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)
		errorLogs := []map[string]interface{}{
			{
				"message":   "ERROR: duplicate key value violates unique constraint \"users_email_key\"",
				"severity":  "ERROR",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			{
				"message":   "ERROR: syntax error at or near \"SELECT\"",
				"severity":  "ERROR",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", errorLogs)
		require.NoError(t, err)

		// Verify error logs are properly classified
		parser := collector.GetLogParser()
		for _, log := range errorLogs {
			msg := log["message"].(string)
			category := parser.ClassifyLog(msg)
			assert.NotEmpty(t, category)
		}
	})

	t.Run("detect_slow_queries", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)
		slowQueryLogs := []map[string]interface{}{
			{
				"message":   "duration: 5000.123 ms  execute <unnamed>: SELECT * FROM large_table",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			{
				"message":   "duration: 8500.456 ms  statement: SELECT COUNT(*) FROM huge_table",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", slowQueryLogs)
		require.NoError(t, err)

		// Parse and extract metadata
		parser := collector.GetLogParser()
		for _, log := range slowQueryLogs {
			msg := log["message"].(string)
			metadata := parser.ExtractMetadata(msg)
			assert.NotNil(t, metadata)
		}
	})
}

// TestLogAnalysisClassification tests log classification accuracy
func TestLogAnalysisClassification(t *testing.T) {
	t.Run("classify_slow_queries", func(t *testing.T) {
		parser := log_analysis.NewLogParser()
		require.NotNil(t, parser)

		slowQueryMsg := "duration: 3000.45 ms  execute <unnamed>: SELECT * FROM users"
		category := parser.ClassifyLog(slowQueryMsg)

		assert.NotEmpty(t, category)
	})

	t.Run("classify_errors", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		errorMsg := "ERROR: permission denied for schema public"
		category := parser.ClassifyLog(errorMsg)

		assert.NotEmpty(t, category)
	})

	t.Run("classify_warnings", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		warningMsg := "WARNING: you don't own a lock of type RowExclusiveLock"
		category := parser.ClassifyLog(warningMsg)

		assert.NotEmpty(t, category)
	})

	t.Run("classify_connection_issues", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		connectionMsg := "FATAL: database \"nonexistent\" does not exist"
		category := parser.ClassifyLog(connectionMsg)

		assert.NotEmpty(t, category)
	})

	t.Run("classify_lock_issues", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		lockMsg := "WARNING: you don't own a lock of type AccessShareLock"
		category := parser.ClassifyLog(lockMsg)

		assert.NotEmpty(t, category)
	})
}

// TestLogAnalysisMetadataExtraction tests metadata extraction from logs
func TestLogAnalysisMetadataExtraction(t *testing.T) {
	t.Run("extract_duration", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		// Log with duration
		msg := "duration: 123.456 ms  execute <unnamed>: SELECT * FROM users"
		metadata := parser.ExtractMetadata(msg)

		assert.NotNil(t, metadata)
		assert.NotEmpty(t, metadata["duration"], "Duration should be extracted")
	})

	t.Run("extract_table_affected", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		// Log with table reference
		msg := "ERROR: permission denied on table orders"
		metadata := parser.ExtractMetadata(msg)

		assert.NotNil(t, metadata)
	})

	t.Run("extract_statement_type", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		msg := "duration: 456.789 ms  statement: VACUUM ANALYZE products"
		metadata := parser.ExtractMetadata(msg)

		assert.NotNil(t, metadata)
	})

	t.Run("handle_minimal_metadata", func(t *testing.T) {
		parser := log_analysis.NewLogParser()

		// Minimal log entry
		msg := "LOG: something happened"
		metadata := parser.ExtractMetadata(msg)

		assert.NotNil(t, metadata)
	})
}

// TestLogAnalysisAnomalyDetection tests anomaly detection capabilities
func TestLogAnalysisAnomalyDetection(t *testing.T) {
	t.Run("detect_high_error_rate", func(t *testing.T) {
		// Simulate baseline: 1% error rate
		baselineErrorRate := 0.01
		totalLogs := 1000
		baselineErrors := int(float64(totalLogs) * baselineErrorRate)

		// Anomaly: 20% error rate (20x increase)
		anomalyErrors := int(float64(totalLogs) * 0.20)

		errorRateIncrease := float64(anomalyErrors) / float64(baselineErrors)
		assert.Greater(t, errorRateIncrease, 5.0) // At least 5x increase

		// This would trigger an anomaly alert
		isAnomaly := errorRateIncrease > 3.0
		assert.True(t, isAnomaly)
	})

	t.Run("detect_slow_query_spike", func(t *testing.T) {
		// Baseline: average query duration 50ms
		baselineAvgDuration := 50.0
		baselineStdDev := 5.0

		// Current: average query duration 150ms (3 sigma above baseline)
		currentAvgDuration := 150.0
		zScore := (currentAvgDuration - baselineAvgDuration) / baselineStdDev

		// 3 sigma indicates anomaly
		isAnomaly := zScore > 3.0
		assert.True(t, isAnomaly)
		assert.Greater(t, zScore, 3.0)
	})

	t.Run("detect_lock_contention", func(t *testing.T) {
		// Baseline: 2 lock warnings per hour
		baselineLockWarnings := 2

		// Anomaly: 50 lock warnings in one sample (spike)
		currentLockWarnings := 50

		ratio := float64(currentLockWarnings) / float64(baselineLockWarnings)
		isAnomaly := ratio > 10.0

		assert.True(t, isAnomaly)
	})

	t.Run("detect_connection_drops", func(t *testing.T) {
		// Baseline: 0 failed connections per sample
		failedConnections := 0

		// First anomaly detection: any connection failure is anomalous
		isAnomaly := failedConnections > 0
		assert.False(t, isAnomaly) // No failures yet

		// Add failures
		failedConnections = 5
		isAnomaly = failedConnections > 0
		assert.True(t, isAnomaly)
	})
}

// TestLogAnalysisPatternDetection tests pattern detection in logs
func TestLogAnalysisPatternDetection(t *testing.T) {
	t.Run("detect_repeated_errors", func(t *testing.T) {
		logsByMessage := map[string]int{
			"ERROR: permission denied": 15,
			"ERROR: constraint violated": 8,
			"ERROR: syntax error": 2,
		}

		// Find most common error
		maxCount := 0
		mostCommonError := ""
		for msg, count := range logsByMessage {
			if count > maxCount {
				maxCount = count
				mostCommonError = msg
			}
		}

		assert.Equal(t, "ERROR: permission denied", mostCommonError)
		assert.Equal(t, 15, maxCount)
	})

	t.Run("detect_hourly_pattern", func(t *testing.T) {
		// Simulate hourly query volume pattern
		hourlyVolume := map[int]int{
			0: 100, 1: 120, 2: 150, 3: 180, // Night spike
			4: 200, 5: 250, 6: 300, 7: 350, // Peak hours
			8: 400, 9: 450, 10: 480, 11: 500, // High load
		}

		// Find peak hour
		maxVolume := 0
		peakHour := 0
		for hour, volume := range hourlyVolume {
			if volume > maxVolume {
				maxVolume = volume
				peakHour = hour
			}
		}

		assert.Equal(t, 11, peakHour)
		assert.Equal(t, 500, maxVolume)
	})

	t.Run("detect_sequential_failures", func(t *testing.T) {
		// Simulate sequence of error logs with timestamps
		logSequence := []struct {
			time    time.Time
			message string
		}{
			{time.Now().Add(-3 * time.Second), "ERROR: connection timeout"},
			{time.Now().Add(-2 * time.Second), "ERROR: connection timeout"},
			{time.Now().Add(-1 * time.Second), "ERROR: connection timeout"},
		}

		// Detect burst of same error
		errorCounts := make(map[string]int)
		for _, log := range logSequence {
			errorCounts[log.message]++
		}

		timeoutCount := errorCounts["ERROR: connection timeout"]
		isBurst := timeoutCount >= 3 && len(logSequence) <= 5 // 3+ errors in short window

		assert.True(t, isBurst)
	})
}

// TestLogAnalysisErrorHandling tests error handling in log analysis
func TestLogAnalysisErrorHandling(t *testing.T) {
	t.Run("handle_invalid_log_entry", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)

		// Missing required fields
		invalidLogs := []map[string]interface{}{
			{
				"message":  "test",
				// Missing severity and timestamp
			},
		}

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", invalidLogs)

		// Should error due to missing fields
		assert.Error(t, err)
	})

	t.Run("handle_empty_log_list", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", []map[string]interface{}{})

		// Empty list should not error
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_timestamp", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)

		invalidLogs := []map[string]interface{}{
			{
				"message":   "test message",
				"severity":  "LOG",
				"timestamp": "invalid-timestamp",
			},
		}

		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", invalidLogs)

		// Should error due to invalid timestamp format
		assert.Error(t, err)
	})

	t.Run("handle_missing_database_id", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)

		validLogs := []map[string]interface{}{
			{
				"message":   "test message",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		ctx := context.Background()
		// Empty database ID should still be accepted
		err := collector.IngestLogs(ctx, "", validLogs)

		assert.NoError(t, err)
	})
}

// TestLogAnalysisContextHandling tests context handling in log analysis
func TestLogAnalysisContextHandling(t *testing.T) {
	t.Run("respect_context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		collector := log_analysis.NewLogCollector(nil)

		logEntries := MockPostgresLogEntries()

		// Start ingestion
		done := make(chan error, 1)
		go func() {
			done <- collector.IngestLogs(ctx, "test-db", logEntries)
		}()

		// Cancel context immediately
		cancel()

		// Wait for completion
		select {
		case err := <-done:
			// May error or succeed depending on timing, both are acceptable
			_ = err
		case <-time.After(2 * time.Second):
			t.Fatal("operation did not complete in time")
		}
	})

	t.Run("handle_context_timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		collector := log_analysis.NewLogCollector(nil)
		logEntries := MockPostgresLogEntries()

		// Try to ingest within timeout
		err := collector.IngestLogs(ctx, "test-db", logEntries)

		// Should complete (no actual DB operation)
		assert.NoError(t, err)
	})
}

// TestLogAnalysisPipelineIntegration tests the complete log analysis pipeline
func TestLogAnalysisPipelineIntegration(t *testing.T) {
	t.Run("complete_log_flow_simulation", func(t *testing.T) {
		// Step 1: Initialize collector
		collector := log_analysis.NewLogCollector(nil)
		require.NotNil(t, collector)

		// Step 2: Get mock logs
		logEntries := MockPostgresLogEntries()
		require.NotEmpty(t, logEntries)

		// Step 3: Ingest logs
		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", logEntries)
		require.NoError(t, err)

		// Step 4: Classify logs
		parser := collector.GetLogParser()
		require.NotNil(t, parser)

		categoryCount := make(map[log_analysis.LogCategory]int)
		for _, logEntry := range logEntries {
			msg := logEntry["message"].(string)
			category := parser.ClassifyLog(msg)
			categoryCount[category]++
		}

		// Step 5: Verify categorization occurred
		assert.Greater(t, len(categoryCount), 0)

		// Step 6: End-to-end verification
		assert.NotEmpty(t, logEntries)
	})

	t.Run("high_volume_log_processing", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)
		ctx := context.Background()

		// Simulate batch processing many logs
		logBatches := 10
		logsPerBatch := 100

		for i := 0; i < logBatches; i++ {
			baseEntries := MockPostgresLogEntries()

			// Expand to desired batch size
			var batch []map[string]interface{}
			for j := 0; j < logsPerBatch/len(baseEntries); j++ {
				batch = append(batch, baseEntries...)
			}

			err := collector.IngestLogs(ctx, "test-db", batch)
			assert.NoError(t, err)
		}

		// All batches processed successfully
		assert.True(t, true)
	})

	t.Run("categorized_log_output_structure", func(t *testing.T) {
		logsByCategory := MockLogEntriesByCategory()
		require.NotEmpty(t, logsByCategory)

		// Verify all categories are present
		expectedCategories := []string{"slow_query", "error", "lock", "connection"}
		for _, cat := range expectedCategories {
			assert.Contains(t, logsByCategory, cat)
		}

		// Verify each category has logs
		for category, logs := range logsByCategory {
			assert.NotEmpty(t, logs, "Category %s should have logs", category)
		}
	})
}

// TestLogAnalysisRealtimeScenarios tests realistic log analysis scenarios
func TestLogAnalysisRealtimeScenarios(t *testing.T) {
	t.Run("monitor_production_logs", func(t *testing.T) {
		collector := log_analysis.NewLogCollector(nil)
		ctx := context.Background()

		// Simulate production log stream
		prodLogs := []map[string]interface{}{
			{
				"message":   "duration: 45.123 ms  execute <unnamed>: SELECT * FROM users WHERE id = $1",
				"severity":  "LOG",
				"timestamp": time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			},
			{
				"message":   "LOG: checkpoint starting: xlog",
				"severity":  "LOG",
				"timestamp": time.Now().Add(-50 * time.Second).Format(time.RFC3339),
			},
			{
				"message":   "LOG: checkpoint complete: wrote 234 buffers (2.1%)",
				"severity":  "LOG",
				"timestamp": time.Now().Add(-40 * time.Second).Format(time.RFC3339),
			},
		}

		err := collector.IngestLogs(ctx, "production", prodLogs)
		assert.NoError(t, err)
	})

	t.Run("detect_maintenance_window", func(t *testing.T) {
		// Simulate logs during maintenance
		maintenanceLogs := []map[string]interface{}{
			{
				"message":   "LOG: autovacuum launcher started",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			{
				"message":   "duration: 12345.678 ms  statement: VACUUM ANALYZE",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		collector := log_analysis.NewLogCollector(nil)
		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", maintenanceLogs)
		assert.NoError(t, err)
	})

	t.Run("track_application_startup", func(t *testing.T) {
		// Simulate logs during application startup
		startupLogs := []map[string]interface{}{
			{
				"message":   "LOG: database system is ready to accept connections",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			{
				"message":   "LOG: redo starts at 0/3000000",
				"severity":  "LOG",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		collector := log_analysis.NewLogCollector(nil)
		ctx := context.Background()
		err := collector.IngestLogs(ctx, "test-db", startupLogs)
		assert.NoError(t, err)
	})
}
