package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// Broadcaster defines the interface for WebSocket broadcasting
// This avoids import cycles with the services package
type Broadcaster interface {
	Broadcast(event string, data map[string]interface{})
}

// SilenceRepository implements the SilenceDB interface for silence operations
type SilenceRepository struct {
	db         *sql.DB
	broadcaster Broadcaster
}

// NewSilenceRepository creates a new SilenceRepository
func NewSilenceRepository(db *sql.DB, broadcaster Broadcaster) *SilenceRepository {
	return &SilenceRepository{
		db:         db,
		broadcaster: broadcaster,
	}
}

// CreateSilence inserts a new silence record
func (r *SilenceRepository) CreateSilence(silence *models.AlertSilence) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO alert_silences (
			alert_rule_id, instance_id, silenced_until, silence_type,
			reason, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		silence.AlertRuleID,
		silence.InstanceID,
		silence.SilencedUntil,
		silence.SilenceType,
		silence.Reason,
		silence.CreatedBy,
		silence.CreatedAt,
	).Scan(&silence.ID)

	if err != nil {
		return fmt.Errorf("create silence: %w", err)
	}

	return nil
}

// GetSilenceByID retrieves a silence by its ID
func (r *SilenceRepository) GetSilenceByID(id int64) (*models.AlertSilence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_rule_id, instance_id, silenced_until, silence_type,
			   reason, created_by, created_at
		FROM alert_silences
		WHERE id = $1
	`

	silence := &models.AlertSilence{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&silence.ID,
		&silence.AlertRuleID,
		&silence.InstanceID,
		&silence.SilencedUntil,
		&silence.SilenceType,
		&silence.Reason,
		&silence.CreatedBy,
		&silence.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("silence not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get silence by id: %w", err)
	}

	return silence, nil
}

// GetActiveSilences retrieves all silences that have not expired yet
func (r *SilenceRepository) GetActiveSilences() ([]*models.AlertSilence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_rule_id, instance_id, silenced_until, silence_type,
			   reason, created_by, created_at
		FROM alert_silences
		WHERE silenced_until > NOW()
		ORDER BY silenced_until DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get active silences: %w", err)
	}
	defer rows.Close()

	silences := make([]*models.AlertSilence, 0)
	for rows.Next() {
		silence := &models.AlertSilence{}
		if err := rows.Scan(
			&silence.ID,
			&silence.AlertRuleID,
			&silence.InstanceID,
			&silence.SilencedUntil,
			&silence.SilenceType,
			&silence.Reason,
			&silence.CreatedBy,
			&silence.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan silence: %w", err)
		}
		silences = append(silences, silence)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate silences: %w", err)
	}

	return silences, nil
}

// GetExpiredSilences retrieves all silences that have expired
func (r *SilenceRepository) GetExpiredSilences() ([]*models.AlertSilence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_rule_id, instance_id, silenced_until, silence_type,
			   reason, created_by, created_at
		FROM alert_silences
		WHERE silenced_until <= NOW()
		ORDER BY silenced_until DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get expired silences: %w", err)
	}
	defer rows.Close()

	silences := make([]*models.AlertSilence, 0)
	for rows.Next() {
		silence := &models.AlertSilence{}
		if err := rows.Scan(
			&silence.ID,
			&silence.AlertRuleID,
			&silence.InstanceID,
			&silence.SilencedUntil,
			&silence.SilenceType,
			&silence.Reason,
			&silence.CreatedBy,
			&silence.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan silence: %w", err)
		}
		silences = append(silences, silence)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate silences: %w", err)
	}

	return silences, nil
}

// UpdateSilence updates an existing silence record
func (r *SilenceRepository) UpdateSilence(silence *models.AlertSilence) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE alert_silences SET
			alert_rule_id = $2,
			instance_id = $3,
			silenced_until = $4,
			silence_type = $5,
			reason = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		silence.ID,
		silence.AlertRuleID,
		silence.InstanceID,
		silence.SilencedUntil,
		silence.SilenceType,
		silence.Reason,
	)

	if err != nil {
		return fmt.Errorf("update silence: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("silence not found: %d", silence.ID)
	}

	return nil
}

// DeleteSilence removes a silence by ID
func (r *SilenceRepository) DeleteSilence(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM alert_silences WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete silence: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("silence not found: %d", id)
	}

	return nil
}

// Broadcast sends a WebSocket event via the broadcaster
func (r *SilenceRepository) Broadcast(event string, data map[string]interface{}) error {
	if r.broadcaster == nil {
		// No broadcaster configured, skip broadcast
		return nil
	}

	r.broadcaster.Broadcast(event, data)
	return nil
}