package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db     *sql.DB
	logger *zap.Logger
}

// Migration represents a single database migration
type Migration struct {
	Name    string
	Content string
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB, logger *zap.Logger) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		logger: logger,
	}
}

// Run executes all pending migrations
func (mr *MigrationRunner) Run(ctx context.Context) error {
	// Create schema_versions table if it doesn't exist
	if err := mr.createVersionsTable(ctx); err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	// Load migrations from embedded files
	migrations, err := mr.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	mr.logger.Info("Found migrations", zap.Int("count", len(migrations)))

	// Execute each migration that hasn't been run yet
	for _, migration := range migrations {
		if err := mr.executeMigration(ctx, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
	}

	mr.logger.Info("All migrations completed successfully")
	return nil
}

// createVersionsTable creates the schema_versions table if it doesn't exist
// IMPORTANT: This also creates the pganalytics schema if it doesn't exist
// This must run BEFORE any migrations, so migrations can rely on the schema existing
func (mr *MigrationRunner) createVersionsTable(ctx context.Context) error {
	// First, ensure the pganalytics schema exists
	if _, err := mr.db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS pganalytics"); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS pganalytics.schema_versions (
		version VARCHAR(100) PRIMARY KEY,
		description TEXT,
		executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		execution_time_ms INTEGER
	);
	`

	_, err := mr.db.ExecContext(ctx, query)
	return err
}

// loadMigrations loads migration files from the migrations directory and returns them sorted
// It looks for migrations in common locations and supports mounted volumes
func (mr *MigrationRunner) loadMigrations() ([]Migration, error) {
	var migrations []Migration

	// Try different locations for migrations directory
	// Priority: environment variable, then relative paths
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// Try common locations
		possiblePaths := []string{
			"/app/migrations",              // Docker container mounted path
			"./migrations",                 // Current directory
			"../migrations",                // Parent directory
			"../../migrations",             // Two levels up
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				migrationsPath = path
				break
			}
		}
	}

	if migrationsPath == "" {
		mr.logger.Warn("No migrations directory found - skipping migrations")
		return migrations, nil
	}

	mr.logger.Debug("Loading migrations from", zap.String("path", migrationsPath))

	// Read migration files from directory
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Only process .sql files
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Skip disabled migrations
		if strings.HasSuffix(file.Name(), ".disabled") {
			mr.logger.Debug("Skipping disabled migration", zap.String("file", file.Name()))
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsPath, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			Name:    file.Name(),
			Content: string(content),
		})
	}

	// Sort migrations by filename (should already be in order, but ensure it)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

// executeMigration executes a single migration if it hasn't been run yet
func (mr *MigrationRunner) executeMigration(ctx context.Context, migration Migration) error {
	// Check if migration has already been executed
	var executed bool
	err := mr.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM pganalytics.schema_versions WHERE version = $1)`,
		migration.Name,
	).Scan(&executed)

	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	// If already executed, skip
	if executed {
		mr.logger.Debug("Migration already executed", zap.String("name", migration.Name))
		return nil
	}

	mr.logger.Info("Executing migration", zap.String("name", migration.Name))

	// Measure execution time
	startTime := time.Now()

	// Execute migration in a transaction
	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Execute the migration SQL
	// Split statements carefully, respecting string literals and dollar-quoted blocks
	statements := splitSQLStatements(migration.Content)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			mr.logger.Error("Migration statement failed",
				zap.String("migration", migration.Name),
				zap.String("statement_preview", truncateString(stmt, 100)),
				zap.Error(err),
			)
			return fmt.Errorf("failed to execute statement in migration %s: %w", migration.Name, err)
		}
	}

	// Record migration in schema_versions table
	executionTime := int(time.Since(startTime).Milliseconds())
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO pganalytics.schema_versions (version, description, execution_time_ms)
		 VALUES ($1, $2, $3)`,
		migration.Name,
		fmt.Sprintf("Executed at %s", time.Now().Format(time.RFC3339)),
		executionTime,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	mr.logger.Info("Migration executed successfully",
		zap.String("name", migration.Name),
		zap.Int("execution_time_ms", executionTime),
	)

	return nil
}

// splitSQLStatements splits SQL content into individual statements
// Properly handles:
// - Single-quoted strings ('...')
// - Dollar-quoted strings ($$...$$, $tag$...$tag$)
// - SQL comments (-- and /* */)
// - Escaped characters
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	runes := []rune(content)
	i := 0

	for i < len(runes) {
		// Skip whitespace and comments at statement boundaries
		for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n' || runes[i] == '\r') {
			current.WriteRune(runes[i])
			i++
		}

		// Check for line comment
		if i < len(runes)-1 && runes[i] == '-' && runes[i+1] == '-' {
			for i < len(runes) && runes[i] != '\n' {
				current.WriteRune(runes[i])
				i++
			}
			if i < len(runes) {
				current.WriteRune(runes[i]) // Add newline
				i++
			}
			continue
		}

		// Check for block comment
		if i < len(runes)-1 && runes[i] == '/' && runes[i+1] == '*' {
			current.WriteRune(runes[i])
			current.WriteRune(runes[i+1])
			i += 2
			for i < len(runes)-1 {
				if runes[i] == '*' && runes[i+1] == '/' {
					current.WriteRune(runes[i])
					current.WriteRune(runes[i+1])
					i += 2
					break
				}
				current.WriteRune(runes[i])
				i++
			}
			continue
		}

		// Check for dollar-quoted string
		if runes[i] == '$' {
			// Find the end of the tag
			tagStart := i
			i++
			for i < len(runes) && (isAlphaNum(runes[i])) {
				i++
			}
			if i < len(runes) && runes[i] == '$' {
				// Found complete opening tag
				tag := string(runes[tagStart : i+1])
				for j := tagStart; j <= i; j++ {
					current.WriteRune(runes[j])
				}
				i++

				// Find closing tag
				for i < len(runes) {
					// Check if remaining content starts with tag
					if i+len(tag) <= len(runes) && string(runes[i:i+len(tag)]) == tag {
						for j := i; j < i+len(tag); j++ {
							current.WriteRune(runes[j])
						}
						i += len(tag)
						break
					}
					current.WriteRune(runes[i])
					i++
				}
				continue
			}
		}

		// Check for single-quoted string
		if runes[i] == '\'' {
			current.WriteRune(runes[i])
			i++
			for i < len(runes) {
				if runes[i] == '\'' {
					if i+1 < len(runes) && runes[i+1] == '\'' {
						// Escaped single quote
						current.WriteRune(runes[i])
						current.WriteRune(runes[i+1])
						i += 2
					} else {
						// End of string
						current.WriteRune(runes[i])
						i++
						break
					}
				} else {
					current.WriteRune(runes[i])
					i++
				}
			}
			continue
		}

		// Check for statement terminator
		if runes[i] == ';' {
			current.WriteRune(runes[i])
			i++
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		// Regular character
		current.WriteRune(runes[i])
		i++
	}

	// Add any remaining statement
	if stmt := current.String(); strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetExecutedMigrations returns a list of all executed migrations
func (mr *MigrationRunner) GetExecutedMigrations(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := mr.db.QueryContext(
		ctx,
		`SELECT version, description, executed_at, execution_time_ms
		 FROM pganalytics.schema_versions
		 ORDER BY executed_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var migrations []map[string]interface{}
	for rows.Next() {
		var version, description string
		var executedAt time.Time
		var executionTimeMs int

		if err := rows.Scan(&version, &description, &executedAt, &executionTimeMs); err != nil {
			return nil, err
		}

		migrations = append(migrations, map[string]interface{}{
			"version":             version,
			"description":         description,
			"executed_at":         executedAt,
			"execution_time_ms":   executionTimeMs,
		})
	}

	return migrations, rows.Err()
}
