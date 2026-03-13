// backend/pkg/services/escalation_worker.go
package services

import (
	"fmt"
	"log"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// EscalationWorkerDB defines the database interface for escalation worker operations
type EscalationWorkerDB interface {
	GetPendingEscalations() ([]*models.EscalationState, error)
	GetPolicy(id int64) (*models.EscalationPolicy, error)
	UpdateEscalationState(state *models.EscalationState) error
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	SendNotification(req *NotificationRequest) error
}

// NotificationRequest represents a notification to be sent for an escalation step
type NotificationRequest struct {
	AlertTriggerID int64
	Channel        string
	Config         map[string]interface{}
	StepNumber     int
}

// EscalationWorker executes escalation steps for pending alerts
type EscalationWorker struct {
	db       EscalationWorkerDB
	notifier Notifier
	logger   *log.Logger
}

// NewEscalationWorker creates a new escalation worker
func NewEscalationWorker(db EscalationWorkerDB, notifier Notifier) *EscalationWorker {
	return &EscalationWorker{
		db:       db,
		notifier: notifier,
		logger:   log.New(log.Writer(), "[EscalationWorker] ", log.LstdFlags),
	}
}

// Process is the main worker loop that executes pending escalations
// Called periodically (e.g., every 30 seconds)
func (ew *EscalationWorker) Process() error {
	// Get all pending escalations
	escalations, err := ew.db.GetPendingEscalations()
	if err != nil {
		ew.logger.Printf("Error getting pending escalations: %v", err)
		return fmt.Errorf("failed to get pending escalations: %w", err)
	}

	now := time.Now()

	for _, state := range escalations {
		// Skip if next_escalation_at is in future (not time yet)
		if state.NextEscalationAt != nil && state.NextEscalationAt.After(now) {
			continue
		}

		// Skip if status = "acknowledged" (user ack'd)
		if state.Status == "acknowledged" {
			continue
		}

		// Process this escalation
		if err := ew.processEscalation(state); err != nil {
			ew.logger.Printf("Error processing escalation %d: %v", state.ID, err)
			// Continue with next escalation, don't return on error
			continue
		}
	}

	return nil
}

// processEscalation processes a single escalation state
func (ew *EscalationWorker) processEscalation(state *models.EscalationState) error {
	// Get policy
	policy, err := ew.db.GetPolicy(state.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to get policy %d: %w", state.PolicyID, err)
	}

	if policy == nil {
		return fmt.Errorf("policy %d not found", state.PolicyID)
	}

	// Check if all steps exhausted (current_step >= len(policy.Steps))
	if state.CurrentStep >= len(policy.Steps) {
		state.Status = "exhausted"
		if err := ew.db.UpdateEscalationState(state); err != nil {
			ew.logger.Printf("Failed to update escalation state to exhausted: %v", err)
		}
		ew.logger.Printf("Escalation %d marked as exhausted (all steps completed)", state.ID)
		return nil
	}

	// Get current step
	step := policy.Steps[state.CurrentStep]

	// Send notification via notifier
	notifReq := &NotificationRequest{
		AlertTriggerID: state.AlertTriggerID,
		Channel:        step.ChannelType,
		Config:         step.ChannelConfig,
		StepNumber:     state.CurrentStep,
	}

	if err := ew.notifier.SendNotification(notifReq); err != nil {
		ew.logger.Printf("Failed to send notification for escalation %d step %d: %v", state.ID, state.CurrentStep, err)
		// Continue even if notification fails, update state anyway
	}

	// Update state: current_step++, last_escalated_at=now
	now := time.Now()
	state.CurrentStep++
	state.LastEscalatedAt = &now

	// Calculate next escalation time if there are more steps
	if state.CurrentStep < len(policy.Steps) {
		nextStep := policy.Steps[state.CurrentStep]
		nextEscalationTime := now.Add(time.Duration(nextStep.DelayMinutes) * time.Minute)
		state.NextEscalationAt = &nextEscalationTime
	} else {
		// No more steps, clear next escalation time
		state.NextEscalationAt = nil
	}

	// Save updated state
	if err := ew.db.UpdateEscalationState(state); err != nil {
		return fmt.Errorf("failed to update escalation state: %w", err)
	}

	ew.logger.Printf("Escalation %d processed successfully: sent step %d, next at %v", state.ID, state.CurrentStep-1, state.NextEscalationAt)

	return nil
}
