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
func (mr *MigrationRunner) createVersionsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS pganalytics.schema_versions (
		id SERIAL PRIMARY KEY,
		version VARCHAR(100) NOT NULL UNIQUE,
		description TEXT,
		executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		execution_time_ms INT
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
	// Split on semicolons followed by newline to avoid splitting inside strings/clauses
	statements := strings.Split(migration.Content, ";\n")
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Re-add semicolon if not the last statement
		if i < len(statements)-1 && !strings.HasSuffix(stmt, ";") {
			stmt += ";"
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

// splitStatements splits SQL content into individual statements
// This is a simple implementation that splits on semicolons
// For complex SQL with strings containing semicolons, a more sophisticated parser would be needed
func splitStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	escaped := false

	for _, char := range content {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' && inString {
			current.WriteRune(char)
			escaped = true
			continue
		}

		if char == '\'' {
			current.WriteRune(char)
			inString = !inString
			continue
		}

		if char == ';' && !inString {
			stmt := current.String()
			statements = append(statements, stmt)
			current.Reset()
			continue
		}

		current.WriteRune(char)
	}

	// Add any remaining content
	if remaining := current.String(); strings.TrimSpace(remaining) != "" {
		statements = append(statements, remaining)
	}

	return statements
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
