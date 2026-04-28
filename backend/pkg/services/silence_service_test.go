// backend/pkg/services/silence_service_test.go
package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// MockSilenceDB implements the SilenceDB interface for testing
type MockSilenceDB struct {
	silences map[int64]*models.AlertSilence
	nextID   int64
	mu       sync.Mutex
}

// NewMockSilenceDB creates a new MockSilenceDB
func NewMockSilenceDB() *MockSilenceDB {
	return &MockSilenceDB{
		silences: make(map[int64]*models.AlertSilence),
		nextID:   1,
	}
}

// CreateSilence creates a new silence in the mock database
func (m *MockSilenceDB) CreateSilence(silence *models.AlertSilence) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	silence.ID = m.nextID
	m.nextID++
	m.silences[silence.ID] = silence
	return nil
}

// GetSilenceByID retrieves a silence by ID
func (m *MockSilenceDB) GetSilenceByID(id int64) (*models.AlertSilence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	silence, exists := m.silences[id]
	if !exists {
		return nil, fmt.Errorf("silence not found")
	}
	return silence, nil
}

// GetActiveSilences returns all silences (expired or not)
// The service layer will filter expired ones
func (m *MockSilenceDB) GetActiveSilences() ([]*models.AlertSilence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var silences []*models.AlertSilence
	for _, silence := range m.silences {
		silences = append(silences, silence)
	}
	return silences, nil
}

// GetExpiredSilences returns silences that have expired
func (m *MockSilenceDB) GetExpiredSilences() ([]*models.AlertSilence, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expired []*models.AlertSilence
	for _, silence := range m.silences {
		if silence.SilencedUntil.Before(now) {
			expired = append(expired, silence)
		}
	}
	return expired, nil
}

// UpdateSilence updates a silence
func (m *MockSilenceDB) UpdateSilence(silence *models.AlertSilence) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.silences[silence.ID]; !exists {
		return fmt.Errorf("silence not found")
	}
	m.silences[silence.ID] = silence
	return nil
}

// DeleteSilence deletes a silence
func (m *MockSilenceDB) DeleteSilence(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.silences[id]; !exists {
		return fmt.Errorf("silence not found")
	}
	delete(m.silences, id)
	return nil
}

// Broadcast broadcasts a WebSocket event
func (m *MockSilenceDB) Broadcast(event string, data map[string]interface{}) error {
	// Mock implementation - just return nil
	return nil
}

// Test: TestCreateSilence - Creates silence successfully
func TestCreateSilence(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	err := service.CreateSilence(1, 60, "rule", nil, "Maintenance window")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify silence was created
	silences, _ := db.GetActiveSilences()
	if len(silences) != 1 {
		t.Errorf("Expected 1 silence, got %d", len(silences))
	}

	if silences[0].AlertRuleID != 1 {
		t.Errorf("Expected rule_id 1, got %d", silences[0].AlertRuleID)
	}

	if silences[0].SilenceType != "rule" {
		t.Errorf("Expected silence_type 'rule', got '%s'", silences[0].SilenceType)
	}
}

// Test: TestIsSilenced_RuleSilence - Rule-level silence applies to all instances
func TestIsSilenced_RuleSilence(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Create a rule-level silence
	err := service.CreateSilence(1, 60, "rule", nil, "Rule maintenance")
	if err != nil {
		t.Errorf("Failed to create silence: %v", err)
	}

	// Check if rule is silenced (with and without instance)
	if !service.IsSilenced(1, nil) {
		t.Error("Expected rule 1 to be silenced (no instance)")
	}

	instance1 := 100
	if !service.IsSilenced(1, &instance1) {
		t.Error("Expected rule 1 to be silenced for instance 100")
	}

	instance2 := 200
	if !service.IsSilenced(1, &instance2) {
		t.Error("Expected rule 1 to be silenced for instance 200")
	}

	// Different rule should not be silenced
	if service.IsSilenced(2, nil) {
		t.Error("Expected rule 2 to NOT be silenced")
	}
}

// Test: TestIsSilenced_ExpiredSilence - Expired silence doesn't apply
func TestIsSilenced_ExpiredSilence(t *testing.T) {
	db := NewMockSilenceDB()

	// Manually create an expired silence
	expiredSilence := &models.AlertSilence{
		AlertRuleID:   1,
		InstanceID:    0,
		SilencedUntil: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		SilenceType:   "rule",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}
	db.CreateSilence(expiredSilence)

	service := NewSilenceService(db)

	// Rule should NOT be silenced (silence expired)
	if service.IsSilenced(1, nil) {
		t.Error("Expected expired silence to not apply")
	}
}

// Test: TestIsSilenced_InstanceSilence - Instance-level silence applies only to that instance
func TestIsSilenced_InstanceSilence(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	instance1 := 100
	instance2 := 200

	// Create instance-level silence for instance 100
	err := service.CreateSilence(1, 60, "instance", &instance1, "Instance maintenance")
	if err != nil {
		t.Errorf("Failed to create silence: %v", err)
	}

	// Instance 100 should be silenced
	if !service.IsSilenced(1, &instance1) {
		t.Error("Expected rule 1 instance 100 to be silenced")
	}

	// Instance 200 should NOT be silenced
	if service.IsSilenced(1, &instance2) {
		t.Error("Expected rule 1 instance 200 to NOT be silenced")
	}

	// Rule-level check without instance should NOT be silenced
	if service.IsSilenced(1, nil) {
		t.Error("Expected rule 1 (no instance) to NOT be silenced with instance-level silence")
	}
}

// Test: TestIsSilenced_DifferentRule - Silence for rule A doesn't affect rule B
func TestIsSilenced_DifferentRule(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Create silence for rule 1
	err := service.CreateSilence(1, 60, "rule", nil, "Rule 1 maintenance")
	if err != nil {
		t.Errorf("Failed to create silence: %v", err)
	}

	// Rule 1 should be silenced
	if !service.IsSilenced(1, nil) {
		t.Error("Expected rule 1 to be silenced")
	}

	// Rule 2 should NOT be silenced
	if service.IsSilenced(2, nil) {
		t.Error("Expected rule 2 to NOT be silenced")
	}
}

// Test: TestGetActiveSilences - Returns only non-expired silences
func TestGetActiveSilences(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Create an active silence
	err := service.CreateSilence(1, 60, "rule", nil, "Active silence")
	if err != nil {
		t.Errorf("Failed to create active silence: %v", err)
	}

	// Manually create an expired silence
	expiredSilence := &models.AlertSilence{
		AlertRuleID:   2,
		InstanceID:    0,
		SilencedUntil: time.Now().Add(-1 * time.Hour),
		SilenceType:   "rule",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}
	db.CreateSilence(expiredSilence)

	// Get active silences
	activeSilences, err := service.GetActiveSilences()
	if err != nil {
		t.Errorf("Failed to get active silences: %v", err)
	}

	// Should only return 1 active silence
	if len(activeSilences) != 1 {
		t.Errorf("Expected 1 active silence, got %d", len(activeSilences))
	}

	if activeSilences[0].AlertRuleID != 1 {
		t.Errorf("Expected rule_id 1, got %d", activeSilences[0].AlertRuleID)
	}
}

// Test: TestCreateSilence_InvalidDuration - Negative/zero duration fails
func TestCreateSilence_InvalidDuration(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	testCases := []struct {
		name     string
		duration int
	}{
		{"zero duration", 0},
		{"negative duration", -1},
		{"large negative duration", -100},
	}

	for _, tc := range testCases {
		err := service.CreateSilence(1, tc.duration, "rule", nil, "Test")
		if err == nil {
			t.Errorf("[%s] Expected error, got nil", tc.name)
		}
	}
}

// Test: TestCreateSilence_InvalidSilenceType - Invalid type fails
func TestCreateSilence_InvalidSilenceType(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	testCases := []struct {
		name        string
		silenceType string
		shouldFail  bool
	}{
		{"valid type: rule", "rule", false},
		{"valid type: instance", "instance", false},
		{"valid type: all", "all", false},
		{"invalid type: temporary", "temporary", true},
		{"invalid type: permanent", "permanent", true},
		{"invalid type: invalid", "invalid", true},
	}

	for _, tc := range testCases {
		instanceID := 100
		err := service.CreateSilence(1, 60, tc.silenceType, &instanceID, "Test")

		if tc.shouldFail {
			if err == nil {
				t.Errorf("[%s] Expected error, got nil", tc.name)
			}
		} else {
			if err != nil {
				t.Errorf("[%s] Expected no error, got: %v", tc.name, err)
			}
		}
	}
}

// Test: TestCreateSilence_InstanceRequired - Instance type requires instanceID
func TestCreateSilence_InstanceRequired(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Try to create instance-level silence without instance ID
	err := service.CreateSilence(1, 60, "instance", nil, "Test")
	if err == nil {
		t.Error("Expected error when creating instance silence without instance_id")
	}
}

// Test: TestCreateSilence_AllSilence - Global silence affects the specific rule
func TestCreateSilence_AllSilence(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Create a global silence for rule 1 (affects all instances of rule 1)
	err := service.CreateSilence(1, 60, "all", nil, "Global maintenance")
	if err != nil {
		t.Errorf("Failed to create global silence: %v", err)
	}

	// Rule 1 should be silenced for all instances
	if !service.IsSilenced(1, nil) {
		t.Error("Expected rule 1 to be silenced by global silence")
	}

	instance := 100
	if !service.IsSilenced(1, &instance) {
		t.Error("Expected rule 1 instance 100 to be silenced by global silence")
	}

	// Different rule should not be silenced
	if service.IsSilenced(2, nil) {
		t.Error("Expected rule 2 to NOT be silenced")
	}
}

// Test: TestExpireSilences - Cleanup job removes expired silences
func TestExpireSilences(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	// Create an active silence
	err := service.CreateSilence(1, 60, "rule", nil, "Active")
	if err != nil {
		t.Errorf("Failed to create active silence: %v", err)
	}

	// Manually create an expired silence
	expiredSilence := &models.AlertSilence{
		AlertRuleID:   2,
		InstanceID:    0,
		SilencedUntil: time.Now().Add(-1 * time.Hour),
		SilenceType:   "rule",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}
	db.CreateSilence(expiredSilence)

	// Before expiration: 2 silences
	allSilences, _ := db.GetActiveSilences()
	if len(allSilences) != 2 {
		t.Errorf("Expected 2 silences before cleanup, got %d", len(allSilences))
	}

	// Run expiration cleanup
	err = service.ExpireSilences()
	if err != nil {
		t.Errorf("Failed to expire silences: %v", err)
	}

	// After expiration: 1 silence
	allSilences, _ = db.GetActiveSilences()
	if len(allSilences) != 1 {
		t.Errorf("Expected 1 silence after cleanup, got %d", len(allSilences))
	}
}

// Test: TestMultipleSilences - Multiple silences work correctly
func TestMultipleSilences(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	instance1 := 100
	instance2 := 200

	// Create multiple silences
	testCases := []struct {
		ruleID     int64
		duration   int
		sType      string
		instanceID *int
		reason     string
	}{
		{1, 60, "rule", nil, "Rule 1 maintenance"},
		{2, 30, "instance", &instance1, "Instance 100 maintenance"},
		{3, 120, "instance", &instance2, "Instance 200 maintenance"},
	}

	for _, tc := range testCases {
		err := service.CreateSilence(tc.ruleID, tc.duration, tc.sType, tc.instanceID, tc.reason)
		if err != nil {
			t.Errorf("Failed to create silence: %v", err)
		}
	}

	// Verify silences
	if !service.IsSilenced(1, nil) {
		t.Error("Expected rule 1 to be silenced")
	}

	if !service.IsSilenced(2, &instance1) {
		t.Error("Expected rule 2 instance 100 to be silenced")
	}

	if !service.IsSilenced(3, &instance2) {
		t.Error("Expected rule 3 instance 200 to be silenced")
	}

	// Rule 2 instance 200 should NOT be silenced
	if service.IsSilenced(2, &instance2) {
		t.Error("Expected rule 2 instance 200 to NOT be silenced")
	}
}

// Test: TestSilenceWithoutReason - Silence without reason works
func TestSilenceWithoutReason(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	err := service.CreateSilence(1, 60, "rule", nil, "")
	if err != nil {
		t.Errorf("Failed to create silence without reason: %v", err)
	}

	silences, _ := db.GetActiveSilences()
	if len(silences) != 1 {
		t.Errorf("Expected 1 silence, got %d", len(silences))
	}

	// Reason should be nil
	if silences[0].Reason != nil {
		t.Errorf("Expected reason to be nil, got: %v", silences[0].Reason)
	}
}

// Test: TestSilenceDuration - Duration is correctly set
func TestSilenceDuration(t *testing.T) {
	db := NewMockSilenceDB()
	service := NewSilenceService(db)

	beforeCreate := time.Now()
	err := service.CreateSilence(1, 60, "rule", nil, "Test")
	afterCreate := time.Now()

	if err != nil {
		t.Errorf("Failed to create silence: %v", err)
	}

	silences, _ := db.GetActiveSilences()
	if len(silences) != 1 {
		t.Errorf("Expected 1 silence, got %d", len(silences))
	}

	// SilencedUntil should be approximately 60 minutes from now
	expectedMin := beforeCreate.Add(59 * time.Minute)
	expectedMax := afterCreate.Add(61 * time.Minute)

	if silences[0].SilencedUntil.Before(expectedMin) || silences[0].SilencedUntil.After(expectedMax) {
		t.Errorf("Expected silenced_until to be ~60 minutes from now, got %v", silences[0].SilencedUntil)
	}
}
