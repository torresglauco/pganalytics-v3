package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// DATA CLASSIFICATION OPERATIONS (DATA-01, DATA-02, DATA-03)
// ============================================================================

// StoreClassificationResults inserts a batch of classification results into the database
func (p *PostgresDB) StoreClassificationResults(ctx context.Context, results []*models.DataClassificationResult) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_data_classification (
			time, collector_id, database_name, schema_name, table_name, column_name,
			pattern_type, category, confidence, match_count, sample_values, regulation_mapping
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (time, collector_id, database_name, schema_name, table_name, column_name)
		DO UPDATE SET
			confidence = EXCLUDED.confidence,
			match_count = EXCLUDED.match_count,
			sample_values = EXCLUDED.sample_values,
			regulation_mapping = EXCLUDED.regulation_mapping
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare classification insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, r := range results {
		// Convert sample_values to JSONB
		samplesJSON, err := json.Marshal(r.SampleValues)
		if err != nil {
			return apperrors.DatabaseError("marshal sample values", err.Error())
		}

		// Convert regulation_mapping to JSONB
		regMapJSON, err := json.Marshal(r.RegulationMapping)
		if err != nil {
			return apperrors.DatabaseError("marshal regulation mapping", err.Error())
		}

		_, err = stmt.ExecContext(ctx,
			r.Time, r.CollectorID, r.DatabaseName, r.SchemaName, r.TableName, r.ColumnName,
			r.PatternType, r.Category, r.Confidence, r.MatchCount, samplesJSON, regMapJSON,
		)
		if err != nil {
			return apperrors.DatabaseError("insert classification result", err.Error())
		}
	}

	return tx.Commit()
}

// GetClassificationResults retrieves classification results with filtering
func (p *PostgresDB) GetClassificationResults(ctx context.Context, collectorID uuid.UUID, filter models.ClassificationFilter) ([]*models.DataClassificationResult, error) {
	// Build time filter based on timeRange
	var timeFilter string
	switch filter.TimeRange {
	case "1h":
		timeFilter = "AND time > NOW() - INTERVAL '1 hour'"
	case "7d":
		timeFilter = "AND time > NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "AND time > NOW() - INTERVAL '30 days'"
	default: // "24h"
		timeFilter = "AND time > NOW() - INTERVAL '24 hours'"
	}

	query := fmt.Sprintf(`
		SELECT time, collector_id, database_name, schema_name, table_name, column_name,
			pattern_type, category, confidence, match_count, sample_values, regulation_mapping
		FROM metrics_data_classification
		WHERE collector_id = $1 %s
	`, timeFilter)

	args := []interface{}{collectorID}
	argNum := 2

	if filter.DatabaseName != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *filter.DatabaseName)
		argNum++
	}

	if filter.SchemaName != nil {
		query += fmt.Sprintf(" AND schema_name = $%d", argNum)
		args = append(args, *filter.SchemaName)
		argNum++
	}

	if filter.TableName != nil {
		query += fmt.Sprintf(" AND table_name = $%d", argNum)
		args = append(args, *filter.TableName)
		argNum++
	}

	if filter.PatternType != nil {
		query += fmt.Sprintf(" AND pattern_type = $%d", argNum)
		args = append(args, *filter.PatternType)
		argNum++
	}

	if filter.Category != nil {
		query += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, *filter.Category)
		argNum++
	}

	// Add limit and offset
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query += fmt.Sprintf(" ORDER BY confidence DESC, match_count DESC, time DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query classification results", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var results []*models.DataClassificationResult
	for rows.Next() {
		r := &models.DataClassificationResult{}
		var samplesJSON, regMapJSON []byte

		err := rows.Scan(
			&r.Time, &r.CollectorID, &r.DatabaseName, &r.SchemaName, &r.TableName, &r.ColumnName,
			&r.PatternType, &r.Category, &r.Confidence, &r.MatchCount, &samplesJSON, &regMapJSON,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan classification result", err.Error())
		}

		// Parse JSONB fields
		if len(samplesJSON) > 0 {
			if err := json.Unmarshal(samplesJSON, &r.SampleValues); err != nil {
				return nil, apperrors.DatabaseError("unmarshal sample values", err.Error())
			}
		}

		if len(regMapJSON) > 0 {
			if err := json.Unmarshal(regMapJSON, &r.RegulationMapping); err != nil {
				return nil, apperrors.DatabaseError("unmarshal regulation mapping", err.Error())
			}
		}

		results = append(results, r)
	}

	return results, nil
}

// GetClassificationReport returns an aggregated classification report
func (p *PostgresDB) GetClassificationReport(ctx context.Context, collectorID uuid.UUID) (*models.ClassificationReportResponse, error) {
	// Get total counts per database
	query := `
		SELECT
			COUNT(DISTINCT database_name) as total_databases,
			COUNT(DISTINCT database_name || '.' || schema_name || '.' || table_name) as total_tables,
			COUNT(DISTINCT database_name || '.' || schema_name || '.' || table_name || '.' || column_name) as total_columns,
			SUM(CASE WHEN category = 'PII' THEN 1 ELSE 0 END) as pii_columns,
			SUM(CASE WHEN category = 'PCI' THEN 1 ELSE 0 END) as pci_columns,
			SUM(CASE WHEN category = 'SENSITIVE' THEN 1 ELSE 0 END) as sensitive_columns,
			SUM(CASE WHEN category = 'CUSTOM' THEN 1 ELSE 0 END) as custom_columns
		FROM (
			SELECT DISTINCT database_name, schema_name, table_name, column_name, category
			FROM metrics_data_classification
			WHERE collector_id = $1
		) distinct_columns
	`

	report := &models.ClassificationReportResponse{
		PatternBreakdown:  make(map[string]int64),
		CategoryBreakdown: make(map[string]int64),
	}

	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(
		&report.TotalDatabases, &report.TotalTables, &report.TotalColumns,
		&report.PiiColumns, &report.PciColumns, &report.SensitiveColumns, &report.CustomColumns,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, apperrors.DatabaseError("query classification report totals", err.Error())
	}

	// Get pattern breakdown
	patternQuery := `
		SELECT pattern_type, COUNT(*) as count
		FROM (
			SELECT DISTINCT database_name, schema_name, table_name, column_name, pattern_type
			FROM metrics_data_classification
			WHERE collector_id = $1
		) distinct_patterns
		GROUP BY pattern_type
	`

	rows, err := p.db.QueryContext(ctx, patternQuery, collectorID)
	if err != nil {
		return nil, apperrors.DatabaseError("query pattern breakdown", err.Error())
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var patternType string
		var count int64
		if err := rows.Scan(&patternType, &count); err != nil {
			return nil, apperrors.DatabaseError("scan pattern breakdown", err.Error())
		}
		report.PatternBreakdown[patternType] = count
	}

	// Get category breakdown
	categoryQuery := `
		SELECT category, COUNT(*) as count
		FROM (
			SELECT DISTINCT database_name, schema_name, table_name, column_name, category
			FROM metrics_data_classification
			WHERE collector_id = $1
		) distinct_categories
		GROUP BY category
	`

	rows, err = p.db.QueryContext(ctx, categoryQuery, collectorID)
	if err != nil {
		return nil, apperrors.DatabaseError("query category breakdown", err.Error())
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, apperrors.DatabaseError("scan category breakdown", err.Error())
		}
		report.CategoryBreakdown[category] = count
	}

	return report, nil
}

// ============================================================================
// CUSTOM PATTERN OPERATIONS (DATA-04)
// ============================================================================

// GetCustomPatterns retrieves custom patterns (global + tenant-specific)
func (p *PostgresDB) GetCustomPatterns(ctx context.Context, tenantID *uuid.UUID) ([]*models.CustomPattern, error) {
	query := `
		SELECT id, tenant_id, pattern_name, pattern_regex, category,
			validation_algorithm, description, enabled, created_at, updated_at
		FROM data_classification_patterns
		WHERE enabled = TRUE
	`

	args := []interface{}{}

	// Filter by tenant: include global patterns (tenant_id IS NULL) + tenant-specific
	if tenantID != nil {
		query += " AND (tenant_id IS NULL OR tenant_id = $1)"
		args = append(args, *tenantID)
	} else {
		query += " AND tenant_id IS NULL"
	}

	query += " ORDER BY pattern_name"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query custom patterns", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var patterns []*models.CustomPattern
	for rows.Next() {
		p := &models.CustomPattern{}
		var tenantID sql.NullTime
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&p.ID, &p.TenantID, &p.PatternName, &p.PatternRegex, &p.Category,
			&p.ValidationAlgorithm, &p.Description, &p.Enabled,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan custom pattern", err.Error())
		}

		if createdAt.Valid {
			t := createdAt.Time
			p.CreatedAt = &t
		}
		if updatedAt.Valid {
			t := updatedAt.Time
			p.UpdatedAt = &t
		}

		// Handle tenant_id scan
		_ = tenantID // The tenantID variable is for future use if needed

		patterns = append(patterns, p)
	}

	return patterns, nil
}

// CreateCustomPattern creates a new custom detection pattern
func (p *PostgresDB) CreateCustomPattern(ctx context.Context, pattern *models.CustomPattern) error {
	// Validate pattern_regex is valid regex
	if _, err := regexp.Compile(pattern.PatternRegex); err != nil {
		return apperrors.BadRequest("Invalid regex pattern", err.Error())
	}

	now := time.Now()
	query := `
		INSERT INTO data_classification_patterns (
			tenant_id, pattern_name, pattern_regex, category,
			validation_algorithm, description, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := p.db.QueryRowContext(ctx, query,
		pattern.TenantID, pattern.PatternName, pattern.PatternRegex, pattern.Category,
		pattern.ValidationAlgorithm, pattern.Description, pattern.Enabled, now, now,
	).Scan(&pattern.ID)

	if err != nil {
		return apperrors.DatabaseError("insert custom pattern", err.Error())
	}

	pattern.CreatedAt = &now
	pattern.UpdatedAt = &now

	return nil
}

// UpdateCustomPattern updates an existing custom pattern
func (p *PostgresDB) UpdateCustomPattern(ctx context.Context, id int, pattern *models.CustomPattern) error {
	// Validate pattern_regex is valid regex
	if pattern.PatternRegex != "" {
		if _, err := regexp.Compile(pattern.PatternRegex); err != nil {
			return apperrors.BadRequest("Invalid regex pattern", err.Error())
		}
	}

	now := time.Now()
	query := `
		UPDATE data_classification_patterns
		SET pattern_regex = COALESCE(NULLIF($2, ''), pattern_regex),
			category = COALESCE(NULLIF($3, ''), category),
			enabled = COALESCE($4, enabled),
			description = COALESCE($5, description),
			updated_at = $6
		WHERE id = $1
	`

	result, err := p.db.ExecContext(ctx, query,
		id, pattern.PatternRegex, pattern.Category, pattern.Enabled, pattern.Description, now,
	)
	if err != nil {
		return apperrors.DatabaseError("update custom pattern", err.Error())
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check update result", err.Error())
	}
	if affected == 0 {
		return apperrors.NotFound("Custom pattern not found", fmt.Sprintf("ID: %d", id))
	}

	pattern.UpdatedAt = &now

	return nil
}

// DeleteCustomPattern deletes a custom pattern
func (p *PostgresDB) DeleteCustomPattern(ctx context.Context, id int) error {
	query := `DELETE FROM data_classification_patterns WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.DatabaseError("delete custom pattern", err.Error())
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check delete result", err.Error())
	}
	if affected == 0 {
		return apperrors.NotFound("Custom pattern not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}

// GetRegulationMappings retrieves all regulation mappings for classification
func (p *PostgresDB) GetRegulationMappings(ctx context.Context) (map[string]map[string][]string, error) {
	query := `SELECT pattern_type, regulations FROM regulation_mappings`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.DatabaseError("query regulation mappings", err.Error())
	}
	defer func() { _ = rows.Close() }()

	mappings := make(map[string]map[string][]string)
	for rows.Next() {
		var patternType string
		var regulationsJSON []byte

		if err := rows.Scan(&patternType, &regulationsJSON); err != nil {
			return nil, apperrors.DatabaseError("scan regulation mapping", err.Error())
		}

		var regulations map[string][]string
		if err := json.Unmarshal(regulationsJSON, &regulations); err != nil {
			return nil, apperrors.DatabaseError("unmarshal regulations", err.Error())
		}

		mappings[patternType] = regulations
	}

	return mappings, nil
}
