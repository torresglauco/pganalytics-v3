package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// EscalationRepository implements the EscalationDB interface for escalation operations
type EscalationRepository struct {
	db *sql.DB
}

// NewEscalationRepository creates a new EscalationRepository
func NewEscalationRepository(db *sql.DB) *EscalationRepository {
	return &EscalationRepository{db: db}
}

// CreatePolicy creates a new escalation policy with its steps
func (r *EscalationRepository) CreatePolicy(policy *models.EscalationPolicy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert policy
	policyQuery := `
		INSERT INTO escalation_policies (name, description, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, policyQuery,
		policy.Name,
		policy.Description,
		policy.IsActive,
		policy.CreatedBy,
		policy.CreatedAt,
		policy.UpdatedAt,
	).Scan(&policy.ID)

	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}

	// Insert steps
	for _, step := range policy.Steps {
		stepQuery := `
			INSERT INTO escalation_policy_steps (policy_id, step_order, channel_type, channel_config, delay_minutes, requires_acknowledgment)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

		channelConfigJSON, err := json.Marshal(step.ChannelConfig)
		if err != nil {
			return fmt.Errorf("marshal channel config: %w", err)
		}

		_, err = tx.ExecContext(ctx, stepQuery,
			policy.ID,
			step.StepOrder,
			step.ChannelType,
			channelConfigJSON,
			step.DelayMinutes,
			step.RequiresAcknowledgment,
		)

		if err != nil {
			return fmt.Errorf("create policy step: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetPolicy retrieves a policy by ID with its steps
func (r *EscalationRepository) GetPolicy(id int64) (*models.EscalationPolicy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get policy
	policyQuery := `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM escalation_policies
		WHERE id = $1
	`

	policy := &models.EscalationPolicy{}
	err := r.db.QueryRowContext(ctx, policyQuery, id).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.IsActive,
		&policy.CreatedBy,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get policy: %w", err)
	}

	// Get steps
	stepsQuery := `
		SELECT id, policy_id, step_order, channel_type, channel_config, delay_minutes, requires_acknowledgment
		FROM escalation_policy_steps
		WHERE policy_id = $1
		ORDER BY step_order
	`

	rows, err := r.db.QueryContext(ctx, stepsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get policy steps: %w", err)
	}
	defer rows.Close()

	policy.Steps = make([]*models.EscalationPolicyStep, 0)
	for rows.Next() {
		step := &models.EscalationPolicyStep{}
		var channelConfigJSON []byte

		if err := rows.Scan(
			&step.ID,
			&step.PolicyID,
			&step.StepOrder,
			&step.ChannelType,
			&channelConfigJSON,
			&step.DelayMinutes,
			&step.RequiresAcknowledgment,
		); err != nil {
			return nil, fmt.Errorf("scan policy step: %w", err)
		}

		if err := json.Unmarshal(channelConfigJSON, &step.ChannelConfig); err != nil {
			return nil, fmt.Errorf("unmarshal channel config: %w", err)
		}

		policy.Steps = append(policy.Steps, step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policy steps: %w", err)
	}

	return policy, nil
}

// UpdatePolicy updates an existing policy and its steps
func (r *EscalationRepository) UpdatePolicy(policy *models.EscalationPolicy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update policy
	policyQuery := `
		UPDATE escalation_policies SET
			name = $2,
			description = $3,
			is_active = $4,
			updated_at = $5
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, policyQuery,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.IsActive,
		policy.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("update policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("policy not found: %d", policy.ID)
	}

	// Delete existing steps
	deleteStepsQuery := `DELETE FROM escalation_policy_steps WHERE policy_id = $1`
	if _, err := tx.ExecContext(ctx, deleteStepsQuery, policy.ID); err != nil {
		return fmt.Errorf("delete policy steps: %w", err)
	}

	// Insert new steps
	for _, step := range policy.Steps {
		stepQuery := `
			INSERT INTO escalation_policy_steps (policy_id, step_order, channel_type, channel_config, delay_minutes, requires_acknowledgment)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

		channelConfigJSON, err := json.Marshal(step.ChannelConfig)
		if err != nil {
			return fmt.Errorf("marshal channel config: %w", err)
		}

		_, err = tx.ExecContext(ctx, stepQuery,
			policy.ID,
			step.StepOrder,
			step.ChannelType,
			channelConfigJSON,
			step.DelayMinutes,
			step.RequiresAcknowledgment,
		)

		if err != nil {
			return fmt.Errorf("create policy step: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ListPolicies retrieves all active escalation policies
func (r *EscalationRepository) ListPolicies() ([]*models.EscalationPolicy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM escalation_policies
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	policies := make([]*models.EscalationPolicy, 0)
	for rows.Next() {
		policy := &models.EscalationPolicy{}
		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.IsActive,
			&policy.CreatedBy,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policies: %w", err)
	}

	return policies, nil
}

// CreateEscalationState creates a new escalation state record
func (r *EscalationRepository) CreateEscalationState(state *models.EscalationState) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO escalation_state (
			alert_trigger_id, policy_id, current_step, ack_received, ack_by, ack_at,
			last_escalated_at, next_escalation_at, status, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var metadataJSON []byte
	var err error
	if state.Metadata != nil {
		metadataJSON, err = json.Marshal(state.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	err = r.db.QueryRowContext(ctx, query,
		state.AlertTriggerID,
		state.PolicyID,
		state.CurrentStep,
		state.AckReceived,
		state.AckBy,
		state.AckAt,
		state.LastEscalatedAt,
		state.NextEscalationAt,
		state.Status,
		metadataJSON,
		state.CreatedAt,
		state.UpdatedAt,
	).Scan(&state.ID)

	if err != nil {
		return fmt.Errorf("create escalation state: %w", err)
	}

	return nil
}

// UpdateEscalationState updates an existing escalation state
func (r *EscalationRepository) UpdateEscalationState(state *models.EscalationState) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE escalation_state SET
			current_step = $2,
			ack_received = $3,
			ack_by = $4,
			ack_at = $5,
			last_escalated_at = $6,
			next_escalation_at = $7,
			status = $8,
			metadata = $9,
			updated_at = $10
		WHERE id = $1
	`

	var metadataJSON []byte
	var err error
	if state.Metadata != nil {
		metadataJSON, err = json.Marshal(state.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	result, err := r.db.ExecContext(ctx, query,
		state.ID,
		state.CurrentStep,
		state.AckReceived,
		state.AckBy,
		state.AckAt,
		state.LastEscalatedAt,
		state.NextEscalationAt,
		state.Status,
		metadataJSON,
		state.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("update escalation state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("escalation state not found: %d", state.ID)
	}

	return nil
}

// GetEscalationState retrieves the escalation state for a given trigger ID
func (r *EscalationRepository) GetEscalationState(triggerID int64) (*models.EscalationState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_trigger_id, policy_id, current_step, ack_received, ack_by, ack_at,
			   last_escalated_at, next_escalation_at, status, metadata, created_at, updated_at
		FROM escalation_state
		WHERE alert_trigger_id = $1
	`

	state := &models.EscalationState{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, triggerID).Scan(
		&state.ID,
		&state.AlertTriggerID,
		&state.PolicyID,
		&state.CurrentStep,
		&state.AckReceived,
		&state.AckBy,
		&state.AckAt,
		&state.LastEscalatedAt,
		&state.NextEscalationAt,
		&state.Status,
		&metadataJSON,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("escalation state not found for trigger: %d", triggerID)
	}
	if err != nil {
		return nil, fmt.Errorf("get escalation state: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &state.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return state, nil
}

// GetPendingEscalations retrieves escalations ready for execution
func (r *EscalationRepository) GetPendingEscalations() ([]*models.EscalationState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_trigger_id, policy_id, current_step, ack_received, ack_by, ack_at,
			   last_escalated_at, next_escalation_at, status, metadata, created_at, updated_at
		FROM escalation_state
		WHERE status = 'pending'
		  AND next_escalation_at <= NOW()
		ORDER BY next_escalation_at
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get pending escalations: %w", err)
	}
	defer rows.Close()

	states := make([]*models.EscalationState, 0)
	for rows.Next() {
		state := &models.EscalationState{}
		var metadataJSON []byte

		if err := rows.Scan(
			&state.ID,
			&state.AlertTriggerID,
			&state.PolicyID,
			&state.CurrentStep,
			&state.AckReceived,
			&state.AckBy,
			&state.AckAt,
			&state.LastEscalatedAt,
			&state.NextEscalationAt,
			&state.Status,
			&metadataJSON,
			&state.CreatedAt,
			&state.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan escalation state: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &state.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate escalation states: %w", err)
	}

	return states, nil
}