package log_analysis

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// LogCollector manages the ingestion and streaming of PostgreSQL logs
type LogCollector struct {
	db     *sql.DB
	parser *LogParser
}

// NewLogCollector creates a new LogCollector instance
func NewLogCollector(db *sql.DB) *LogCollector {
	return &LogCollector{
		db:     db,
		parser: NewLogParser(),
	}
}

// IngestLogs processes a batch of logs and stores them in the database
// It classifies each log using the LogParser and extracts metadata
func (lc *LogCollector) IngestLogs(ctx context.Context, databaseID string, logs []map[string]interface{}) error {
	if len(logs) == 0 {
		return nil
	}

	for _, logEntry := range logs {
		// Extract fields from log entry
		message, ok := logEntry["message"].(string)
		if !ok {
			return fmt.Errorf("invalid log entry: missing or invalid message field")
		}

		severity, ok := logEntry["severity"].(string)
		if !ok {
			return fmt.Errorf("invalid log entry: missing or invalid severity field")
		}

		timestampStr, ok := logEntry["timestamp"].(string)
		if !ok {
			return fmt.Errorf("invalid log entry: missing or invalid timestamp field")
		}

		// Parse timestamp
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}

		// Classify log and extract metadata
		category := lc.parser.ClassifyLog(message)
		metadata := lc.parser.ExtractMetadata(message)

		// Extract optional metadata fields
		duration := metadata["duration"]
		table := metadata["table"]

		// If database is nil (no database connection), skip actual insertion
		// This allows testing without a real database
		if lc.db == nil {
			continue
		}

		// Insert log into database
		query := `
			INSERT INTO logs (database_id, log_timestamp, category, severity, message, duration, table_affected)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err = lc.db.ExecContext(ctx, query,
			databaseID,
			timestamp,
			category,
			severity,
			message,
			duration,
			table,
		)

		if err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}
	}

	return nil
}

// StreamLogs continuously polls the database for new logs and sends them through a channel
// It uses a ticker for polling at regular intervals and respects context cancellation
func (lc *LogCollector) StreamLogs(ctx context.Context, databaseID string, ch chan<- map[string]interface{}) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// If database is nil, return error
			if lc.db == nil {
				return fmt.Errorf("database connection not available")
			}

			rows, err := lc.db.QueryContext(ctx, `
				SELECT id, log_timestamp, category, severity, message
				FROM logs
				WHERE database_id = $1
				ORDER BY created_at DESC
				LIMIT 10
			`, databaseID)

			if err != nil {
				return fmt.Errorf("failed to query logs: %w", err)
			}

			// Process results
			logCount := 0
			for rows.Next() {
				var id int64
				var timestamp time.Time
				var category, severity, message string

				err := rows.Scan(&id, &timestamp, &category, &severity, &message)
				if err != nil {
					rows.Close()
					return fmt.Errorf("failed to scan log row: %w", err)
				}

				logEntry := map[string]interface{}{
					"id":        id,
					"timestamp": timestamp,
					"category":  category,
					"severity":  severity,
					"message":   message,
				}

				// Non-blocking send to channel
				select {
				case <-ctx.Done():
					rows.Close()
					return ctx.Err()
				case ch <- logEntry:
					logCount++
				}
			}

			// Close rows to avoid resource leaks
			err = rows.Err()
			rows.Close()
			if err != nil {
				return fmt.Errorf("error iterating rows: %w", err)
			}
		}
	}
}

// GetLogParser returns the underlying LogParser instance
// Useful for direct parsing operations
func (lc *LogCollector) GetLogParser() *LogParser {
	return lc.parser
}
