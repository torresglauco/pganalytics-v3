package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// HEALTH SCORE OPERATIONS (HOST-04)
// ============================================================================

// StoreHealthScore inserts a health score into the database
func (p *PostgresDB) StoreHealthScore(ctx context.Context, score *models.HealthScore) error {
	if score == nil {
		return nil
	}

	query := `
		INSERT INTO metrics_host_health_scores (
			time, collector_id, health_score, status,
			cpu_score, memory_score, disk_score, load_score,
			calculation_details
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (time, collector_id) DO UPDATE SET
			health_score = EXCLUDED.health_score,
			status = EXCLUDED.status,
			cpu_score = EXCLUDED.cpu_score,
			memory_score = EXCLUDED.memory_score,
			disk_score = EXCLUDED.disk_score,
			load_score = EXCLUDED.load_score,
			calculation_details = EXCLUDED.calculation_details
	`

	_, err := p.db.ExecContext(ctx, query,
		score.Time, score.CollectorID, score.HealthScore, score.Status,
		score.CpuScore, score.MemoryScore, score.DiskScore, score.LoadScore,
		score.CalculationDetails,
	)
	if err != nil {
		return apperrors.DatabaseError("insert health score", err.Error())
	}

	return nil
}

// GetLatestHealthScore retrieves the latest health score for a collector
func (p *PostgresDB) GetLatestHealthScore(ctx context.Context, collectorID uuid.UUID) (*models.HealthScore, error) {
	query := `
		SELECT time, collector_id, health_score, status,
			cpu_score, memory_score, disk_score, load_score,
			calculation_details
		FROM metrics_host_health_scores
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT 1
	`

	score := &models.HealthScore{}
	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(
		&score.Time, &score.CollectorID, &score.HealthScore, &score.Status,
		&score.CpuScore, &score.MemoryScore, &score.DiskScore, &score.LoadScore,
		&score.CalculationDetails,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("query latest health score", err.Error())
	}

	return score, nil
}

// GetHealthScoreHistory retrieves health score history for a collector
func (p *PostgresDB) GetHealthScoreHistory(ctx context.Context, collectorID uuid.UUID, timeRange string, limit, offset int) ([]*models.HealthScore, error) {
	// Build time filter based on timeRange
	var timeFilter string
	switch timeRange {
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
		SELECT time, collector_id, health_score, status,
			cpu_score, memory_score, disk_score, load_score,
			calculation_details
		FROM metrics_host_health_scores
		WHERE collector_id = $1 %s
		ORDER BY time DESC
		LIMIT $2 OFFSET $3
	`, timeFilter)

	rows, err := p.db.QueryContext(ctx, query, collectorID, limit, offset)
	if err != nil {
		return nil, apperrors.DatabaseError("query health score history", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var scores []*models.HealthScore
	for rows.Next() {
		score := &models.HealthScore{}
		err := rows.Scan(
			&score.Time, &score.CollectorID, &score.HealthScore, &score.Status,
			&score.CpuScore, &score.MemoryScore, &score.DiskScore, &score.LoadScore,
			&score.CalculationDetails,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan health score", err.Error())
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// CalculateAndStoreHealthScore fetches latest metrics, calculates health score, and stores it
func (p *PostgresDB) CalculateAndStoreHealthScore(ctx context.Context, collectorID uuid.UUID) (*models.HealthScore, error) {
	// Fetch latest host metrics (limit 1 for most recent)
	metrics, err := p.GetHostMetrics(ctx, collectorID, "1h", 1)
	if err != nil {
		return nil, apperrors.DatabaseError("fetch metrics for health score", err.Error())
	}

	if len(metrics) == 0 {
		return nil, apperrors.NotFound("No metrics found for collector", collectorID.String())
	}

	// Calculate health score with default weights
	score := services.CalculateHealthScoreWithDetails(metrics[0], models.DefaultHealthScoreWeights)
	if score == nil {
		return nil, apperrors.InternalServerError("Failed to calculate health score", "nil result")
	}

	// Store the score
	if err := p.StoreHealthScore(ctx, score); err != nil {
		return nil, err
	}

	return score, nil
}

// GetHealthScoreCount returns the total count of health scores for a collector
func (p *PostgresDB) GetHealthScoreCount(ctx context.Context, collectorID uuid.UUID, timeRange string) (int, error) {
	// Build time filter based on timeRange
	var timeFilter string
	switch timeRange {
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
		SELECT COUNT(*)
		FROM metrics_host_health_scores
		WHERE collector_id = $1 %s
	`, timeFilter)

	var count int
	err := p.db.QueryRowContext(ctx, query, collectorID).Scan(&count)
	if err != nil {
		return 0, apperrors.DatabaseError("count health scores", err.Error())
	}

	return count, nil
}

// DeleteHealthScoresOlderThan deletes health scores older than the specified duration
func (p *PostgresDB) DeleteHealthScoresOlderThan(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	query := `
		DELETE FROM metrics_host_health_scores
		WHERE time < $1
	`

	result, err := p.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, apperrors.DatabaseError("delete old health scores", err.Error())
	}

	return result.RowsAffected()
}
