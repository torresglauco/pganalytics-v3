package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// VERSION HEALTH CHECK OPERATIONS (VER-03)
// ============================================================================

// GetHealthChecksForVersion retrieves all health checks applicable to a specific PostgreSQL version
func (p *PostgresDB) GetHealthChecksForVersion(ctx context.Context, pgVersion int) ([]*models.VersionHealthCheck, error) {
	query := `
		SELECT id, min_version, max_version, check_name, check_query, expected_result,
			   severity, description, remediation, category, created_at, updated_at
		FROM postgres_health_checks
		WHERE min_version <= $1
		  AND (max_version IS NULL OR max_version >= $1)
		ORDER BY severity DESC, category
	`

	rows, err := p.db.QueryContext(ctx, query, pgVersion)
	if err != nil {
		return nil, apperrors.DatabaseError("query health checks for version", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var checks []*models.VersionHealthCheck
	for rows.Next() {
		check := &models.VersionHealthCheck{}
		var maxVersion sql.NullInt64

		err := rows.Scan(
			&check.ID, &check.MinVersion, &maxVersion, &check.CheckName, &check.CheckQuery,
			&check.ExpectedResult, &check.Severity, &check.Description, &check.Remediation,
			&check.Category, &check.CreatedAt, &check.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan health check", err.Error())
		}

		// Handle NULL max_version (0 means no upper limit)
		if maxVersion.Valid {
			check.MaxVersion = int(maxVersion.Int64)
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// GetHealthCheckByID retrieves a single health check by ID
func (p *PostgresDB) GetHealthCheckByID(ctx context.Context, id int) (*models.VersionHealthCheck, error) {
	query := `
		SELECT id, min_version, max_version, check_name, check_query, expected_result,
			   severity, description, remediation, category, created_at, updated_at
		FROM postgres_health_checks
		WHERE id = $1
	`

	check := &models.VersionHealthCheck{}
	var maxVersion sql.NullInt64

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&check.ID, &check.MinVersion, &maxVersion, &check.CheckName, &check.CheckQuery,
		&check.ExpectedResult, &check.Severity, &check.Description, &check.Remediation,
		&check.Category, &check.CreatedAt, &check.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("health check not found", fmt.Sprintf("id: %d", id))
		}
		return nil, apperrors.DatabaseError("query health check by id", err.Error())
	}

	// Handle NULL max_version
	if maxVersion.Valid {
		check.MaxVersion = int(maxVersion.Int64)
	}

	return check, nil
}

// GetAllHealthChecks retrieves all health checks ordered by version and name
func (p *PostgresDB) GetAllHealthChecks(ctx context.Context) ([]*models.VersionHealthCheck, error) {
	query := `
		SELECT id, min_version, max_version, check_name, check_query, expected_result,
			   severity, description, remediation, category, created_at, updated_at
		FROM postgres_health_checks
		ORDER BY min_version, check_name
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.DatabaseError("query all health checks", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var checks []*models.VersionHealthCheck
	for rows.Next() {
		check := &models.VersionHealthCheck{}
		var maxVersion sql.NullInt64

		err := rows.Scan(
			&check.ID, &check.MinVersion, &maxVersion, &check.CheckName, &check.CheckQuery,
			&check.ExpectedResult, &check.Severity, &check.Description, &check.Remediation,
			&check.Category, &check.CreatedAt, &check.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan health check", err.Error())
		}

		// Handle NULL max_version
		if maxVersion.Valid {
			check.MaxVersion = int(maxVersion.Int64)
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// RunHealthCheck executes a health check query and returns the result
// This method runs the check query against the monitored database and builds a result
func (p *PostgresDB) RunHealthCheck(ctx context.Context, collectorID uuid.UUID, check *models.VersionHealthCheck) (*models.HealthCheckResult, error) {
	result := &models.HealthCheckResult{
		CheckID:        check.ID,
		CheckName:      check.CheckName,
		Severity:       check.Severity,
		ExpectedResult: check.ExpectedResult,
		Remediation:    check.Remediation,
		CheckedAt:      time.Now(),
	}

	// Execute the check query
	// Note: This is a simplified implementation that runs the query on the pganalytics database
	// In a real implementation, this would connect to the monitored database via the collector
	var actualResult string
	err := p.db.QueryRowContext(ctx, check.CheckQuery).Scan(&actualResult)
	if err != nil {
		// Query failed - mark as not passed with error message
		result.Passed = false
		result.ActualResult = "Query execution failed"
		result.Message = err.Error()
		return result, nil
	}

	result.ActualResult = actualResult

	// Simple pass/fail determination based on query result
	// In a real implementation, this would use more sophisticated comparison logic
	// For now, we consider the check passed if the query executed successfully
	result.Passed = true
	result.Message = "Check executed successfully"

	return result, nil
}

// StoreHealthCheckResult stores a health check result in the database
func (p *PostgresDB) StoreHealthCheckResult(ctx context.Context, collectorID uuid.UUID, result *models.HealthCheckResult) error {
	query := `
		INSERT INTO postgres_health_check_results (collector_id, check_id, passed, actual_result, checked_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := p.db.ExecContext(ctx, query,
		collectorID, result.CheckID, result.Passed, result.ActualResult, result.CheckedAt,
	)
	if err != nil {
		return apperrors.DatabaseError("store health check result", err.Error())
	}

	return nil
}

// GetRecentHealthCheckResults retrieves recent health check results for a collector
func (p *PostgresDB) GetRecentHealthCheckResults(ctx context.Context, collectorID uuid.UUID, limit int) ([]*models.HealthCheckResult, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT r.check_id, hc.check_name, hc.severity, r.passed, r.actual_result,
			   hc.expected_result, hc.remediation, r.checked_at
		FROM postgres_health_check_results r
		JOIN postgres_health_checks hc ON r.check_id = hc.id
		WHERE r.collector_id = $1
		ORDER BY r.checked_at DESC
		LIMIT $2
	`

	rows, err := p.db.QueryContext(ctx, query, collectorID, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("query recent health check results", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var results []*models.HealthCheckResult
	for rows.Next() {
		result := &models.HealthCheckResult{}
		err := rows.Scan(
			&result.CheckID, &result.CheckName, &result.Severity, &result.Passed,
			&result.ActualResult, &result.ExpectedResult, &result.Remediation, &result.CheckedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan health check result", err.Error())
		}
		results = append(results, result)
	}

	return results, nil
}
