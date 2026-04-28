// backend/pkg/services/silence_service.go
package services

import (
	"fmt"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// SilenceDB defines the database interface for silence operations
type SilenceDB interface {
	CreateSilence(silence *models.AlertSilence) error
	GetSilenceByID(id int64) (*models.AlertSilence, error)
	GetActiveSilences() ([]*models.AlertSilence, error)
	GetExpiredSilences() ([]*models.AlertSilence, error)
	UpdateSilence(silence *models.AlertSilence) error
	DeleteSilence(id int64) error
	Broadcast(event string, data map[string]interface{}) error
}

// SilenceService manages alert suppression with auto-expiration
type SilenceService struct {
	db SilenceDB
}

// NewSilenceService creates a new SilenceService
func NewSilenceService(db SilenceDB) *SilenceService {
	return &SilenceService{
		db: db,
	}
}

// CreateSilence creates a new silence for an alert rule or instance
// durationMinutes: duration in minutes for the silence
// silenceType: 'rule' (all instances), 'instance' (specific instance), or 'all' (global)
func (s *SilenceService) CreateSilence(ruleID int64, durationMinutes int, silenceType string, instanceID *int, reason string) error {
	// Validate durationMinutes > 0
	if durationMinutes <= 0 {
		return fmt.Errorf("duration_minutes must be greater than 0, got %d", durationMinutes)
	}

	// Validate silenceType
	validTypes := map[string]bool{
		"rule":     true,
		"instance": true,
		"all":      true,
	}
	if !validTypes[silenceType] {
		return fmt.Errorf("invalid silence_type '%s'. Valid types are: rule, instance, all", silenceType)
	}

	// Validate instanceID based on silenceType
	if silenceType == "instance" && instanceID == nil {
		return fmt.Errorf("instance_id is required when silence_type is 'instance'")
	}

	// Create AlertSilence record
	silence := &models.AlertSilence{
		AlertRuleID:   int(ruleID),
		SilencedUntil: time.Now().Add(time.Duration(durationMinutes) * time.Minute),
		SilenceType:   silenceType,
		CreatedAt:     time.Now(),
	}

	// Set InstanceID if provided
	if instanceID != nil {
		silence.InstanceID = *instanceID
	}

	// Set Reason if provided
	if reason != "" {
		silence.Reason = &reason
	}

	// Save to database
	if err := s.db.CreateSilence(silence); err != nil {
		return fmt.Errorf("failed to create silence: %w", err)
	}

	// Broadcast via WebSocket if available
	event := map[string]interface{}{
		"type":           "silence_created",
		"rule_id":        ruleID,
		"instance_id":    instanceID,
		"silence_type":   silenceType,
		"silenced_until": silence.SilencedUntil,
		"reason":         reason,
		"created_at":     silence.CreatedAt,
	}
	_ = s.db.Broadcast("silence_event", event) // Ignore broadcast error

	return nil
}

// IsSilenced checks if an alert rule/instance is currently silenced
// Returns true if an active silence exists, false otherwise
func (s *SilenceService) IsSilenced(ruleID int64, instanceID *int) bool {
	activeSilences, err := s.db.GetActiveSilences()
	if err != nil {
		return false
	}

	now := time.Now()

	for _, silence := range activeSilences {
		// Check if silence has expired
		if silence.SilencedUntil.Before(now) {
			continue
		}

		// Check rule match
		if silence.AlertRuleID != int(ruleID) {
			continue
		}

		// Check silence type and instance match
		switch silence.SilenceType {
		case "all":
			// Global silence applies to all rules and instances
			return true
		case "rule":
			// Rule-level silence applies to all instances of this rule
			return true
		case "instance":
			// Instance-level silence only applies to specific instance
			if instanceID != nil && silence.InstanceID == *instanceID {
				return true
			}
		}
	}

	return false
}

// GetActiveSilences returns all non-expired silences
func (s *SilenceService) GetActiveSilences() ([]*models.AlertSilence, error) {
	silences, err := s.db.GetActiveSilences()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active silences: %w", err)
	}

	now := time.Now()
	var active []*models.AlertSilence

	for _, silence := range silences {
		// Only include silences that haven't expired
		if silence.SilencedUntil.After(now) {
			active = append(active, silence)
		}
	}

	return active, nil
}

// ExpireSilences marks expired silences for cleanup
// This is called as part of a periodic cleanup job (e.g., hourly)
func (s *SilenceService) ExpireSilences() error {
	expiredSilences, err := s.db.GetExpiredSilences()
	if err != nil {
		return fmt.Errorf("failed to retrieve expired silences: %w", err)
	}

	// Delete expired silences
	for _, silence := range expiredSilences {
		if err := s.db.DeleteSilence(silence.ID); err != nil {
			return fmt.Errorf("failed to delete expired silence %d: %w", silence.ID, err)
		}

		// Broadcast expiration event
		event := map[string]interface{}{
			"type":    "silence_expired",
			"id":      silence.ID,
			"rule_id": silence.AlertRuleID,
		}
		_ = s.db.Broadcast("silence_event", event) // Ignore broadcast error
	}

	return nil
}
