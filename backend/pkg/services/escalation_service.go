// backend/pkg/services/escalation_service.go
package services

import (
	"fmt"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// EscalationDB defines the database interface for escalation operations
type EscalationDB interface {
	CreatePolicy(policy *models.EscalationPolicy) error
	GetPolicy(id int64) (*models.EscalationPolicy, error)
	UpdatePolicy(policy *models.EscalationPolicy) error
	ListPolicies() ([]*models.EscalationPolicy, error)
	CreateEscalationState(state *models.EscalationState) error
	UpdateEscalationState(state *models.EscalationState) error
	GetEscalationState(triggerID int64) (*models.EscalationState, error)
	GetPendingEscalations() ([]*models.EscalationState, error)
}

// EscalationService manages multi-step notification escalation policies
type EscalationService struct {
	db       EscalationDB
	notifier Notifier
}

// NewEscalationService creates a new EscalationService
func NewEscalationService(db EscalationDB, notifier Notifier) *EscalationService {
	return &EscalationService{
		db:       db,
		notifier: notifier,
	}
}

// CreatePolicy creates a new escalation policy
// Validates that policy name is not empty before saving
func (es *EscalationService) CreatePolicy(policy *models.EscalationPolicy) error {
	// Validate policy name is not empty
	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// Set default values
	policy.IsActive = true
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	// Save to database
	if err := es.db.CreatePolicy(policy); err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// GetPolicy retrieves a policy with all its steps
func (es *EscalationService) GetPolicy(id int64) (*models.EscalationPolicy, error) {
	policy, err := es.db.GetPolicy(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve policy: %w", err)
	}

	if policy == nil {
		return nil, fmt.Errorf("policy with id %d not found", id)
	}

	return policy, nil
}

// UpdatePolicy updates an existing escalation policy
// Only updates name and description fields
func (es *EscalationService) UpdatePolicy(policy *models.EscalationPolicy) error {
	// Validate policy name is not empty
	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// Set update timestamp
	policy.UpdatedAt = time.Now()

	// Update in database
	if err := es.db.UpdatePolicy(policy); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

// ListPolicies returns all active escalation policies
func (es *EscalationService) ListPolicies() ([]*models.EscalationPolicy, error) {
	policies, err := es.db.ListPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, nil
}

// StartEscalation creates a new escalation state for a triggered alert
// Initializes with current_step=0, status='pending', ack_received=false
func (es *EscalationService) StartEscalation(triggerID int64, policyID int64) error {
	// Verify policy exists
	policy, err := es.db.GetPolicy(policyID)
	if err != nil {
		return fmt.Errorf("failed to verify policy: %w", err)
	}

	if policy == nil {
		return fmt.Errorf("policy with id %d not found", policyID)
	}

	// Create escalation state with initial values
	state := &models.EscalationState{
		AlertTriggerID: triggerID,
		PolicyID:       policyID,
		CurrentStep:    0,
		AckReceived:    false,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	if err := es.db.CreateEscalationState(state); err != nil {
		return fmt.Errorf("failed to create escalation state: %w", err)
	}

	return nil
}

// AcknowledgeAlert marks an alert as acknowledged and stops escalation
// Updates ack_received=true, ack_by, ack_at, status='acknowledged'
func (es *EscalationService) AcknowledgeAlert(triggerID int64, userID int) error {
	// Get current escalation state
	state, err := es.db.GetEscalationState(triggerID)
	if err != nil {
		return fmt.Errorf("failed to retrieve escalation state: %w", err)
	}

	if state == nil {
		return fmt.Errorf("escalation state for trigger %d not found", triggerID)
	}

	// Update state to acknowledged
	state.AckReceived = true
	state.AckBy = &userID
	now := time.Now()
	state.AckAt = &now
	state.Status = "acknowledged"
	state.UpdatedAt = time.Now()

	// Save updated state
	if err := es.db.UpdateEscalationState(state); err != nil {
		return fmt.Errorf("failed to update escalation state: %w", err)
	}

	return nil
}

// UpdateEscalationState updates the escalation state
func (es *EscalationService) UpdateEscalationState(state *models.EscalationState) error {
	// Update timestamp
	state.UpdatedAt = time.Now()

	// Update in database
	if err := es.db.UpdateEscalationState(state); err != nil {
		return fmt.Errorf("failed to update escalation state: %w", err)
	}

	return nil
}

// GetPendingEscalations returns escalations that are ready to be executed
// Returns states with status='pending' and next_escalation_at <= now
func (es *EscalationService) GetPendingEscalations() ([]*models.EscalationState, error) {
	escalations, err := es.db.GetPendingEscalations()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pending escalations: %w", err)
	}

	return escalations, nil
}

// GetEscalationState returns the current escalation state for a trigger
func (es *EscalationService) GetEscalationState(triggerID int64) (*models.EscalationState, error) {
	state, err := es.db.GetEscalationState(triggerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve escalation state: %w", err)
	}

	if state == nil {
		return nil, fmt.Errorf("escalation state for trigger %d not found", triggerID)
	}

	return state, nil
}
