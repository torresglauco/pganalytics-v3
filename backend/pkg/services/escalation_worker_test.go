// backend/pkg/services/escalation_worker_test.go
package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// MockWorkerDB implements the EscalationWorkerDB interface for testing
type MockWorkerDB struct {
	policies map[int64]*models.EscalationPolicy
	states   map[int64]*models.EscalationState
	nextID   int64
	mu       sync.Mutex
}

// NewMockWorkerDB creates a new MockWorkerDB
func NewMockWorkerDB() *MockWorkerDB {
	return &MockWorkerDB{
		policies: make(map[int64]*models.EscalationPolicy),
		states:   make(map[int64]*models.EscalationState),
		nextID:   1,
	}
}

// GetPendingEscalations returns all escalations with status != resolved
func (m *MockWorkerDB) GetPendingEscalations() ([]*models.EscalationState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var pending []*models.EscalationState
	for _, state := range m.states {
		if state.Status != "resolved" {
			pending = append(pending, state)
		}
	}
	return pending, nil
}

// GetPolicy retrieves a policy by ID
func (m *MockWorkerDB) GetPolicy(id int64) (*models.EscalationPolicy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy, exists := m.policies[id]
	if !exists {
		return nil, fmt.Errorf("policy not found")
	}
	return policy, nil
}

// UpdateEscalationState updates an escalation state
func (m *MockWorkerDB) UpdateEscalationState(state *models.EscalationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.states[state.ID]; !exists {
		return fmt.Errorf("escalation state not found")
	}
	m.states[state.ID] = state
	return nil
}

// Helper methods for tests

// CreatePolicy creates a test policy
func (m *MockWorkerDB) CreatePolicy(policy *models.EscalationPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy.ID = m.nextID
	m.nextID++
	m.policies[policy.ID] = policy
	return nil
}

// CreateState creates a test escalation state
func (m *MockWorkerDB) CreateState(state *models.EscalationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state.ID == 0 {
		state.ID = m.nextID
		m.nextID++
	}
	m.states[state.ID] = state
	return nil
}

// MockWorkerNotifier tracks sent notifications
type MockWorkerNotifier struct {
	sentNotifications []*NotificationRequest
	mu                sync.Mutex
}

// SendNotification records the notification
func (m *MockWorkerNotifier) SendNotification(req *NotificationRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentNotifications = append(m.sentNotifications, req)
	return nil
}

// GetSentNotifications returns all sent notifications
func (m *MockWorkerNotifier) GetSentNotifications() []*NotificationRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy
	copied := make([]*NotificationRequest, len(m.sentNotifications))
	copy(copied, m.sentNotifications)
	return copied
}

// Test: TestWorkerProcessesReadyEscalations - Processes escalations where next_escalation_at <= NOW()
func TestWorkerProcessesReadyEscalations(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 2 steps
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: map[string]interface{}{"recipient": "user1@example.com"},
				DelayMinutes:  5,
			},
			{
				StepOrder:     1,
				ChannelType:   "slack",
				ChannelConfig: map[string]interface{}{"webhook": "https://slack.com/hook"},
				DelayMinutes:  10,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state ready for processing (next_escalation_at in past)
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process
	err := worker.Process()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify notification was sent
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifications))
	}

	// Verify state was updated
	states, _ := db.GetPendingEscalations()
	if len(states) != 1 {
		t.Errorf("Expected 1 pending state, got %d", len(states))
	}

	updatedState := states[0]
	if updatedState.CurrentStep != 1 {
		t.Errorf("Expected current_step 1, got %d", updatedState.CurrentStep)
	}
}

// Test: TestWorkerSkipsAcknowledgedEscalations - Skips escalations with status="acknowledged"
func TestWorkerSkipsAcknowledgedEscalations(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: map[string]interface{}{},
				DelayMinutes:  5,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state with status="acknowledged"
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "acknowledged",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process
	err := worker.Process()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no notification was sent
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 0 {
		t.Errorf("Expected 0 notifications, got %d", len(notifications))
	}
}

// Test: TestWorkerSkipsNotReadyEscalations - Skips escalations where next_escalation_at > NOW()
func TestWorkerSkipsNotReadyEscalations(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: map[string]interface{}{},
				DelayMinutes:  5,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state with future next_escalation_at
	futureTime := time.Now().Add(10 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &futureTime,
	}
	db.CreateState(state)

	// Process
	err := worker.Process()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no notification was sent
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 0 {
		t.Errorf("Expected 0 notifications, got %d", len(notifications))
	}
}

// Test: TestWorkerSendsCorrectChannel - Sends notification to correct channel
func TestWorkerSendsCorrectChannel(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with email step
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: map[string]interface{}{"recipient": "admin@example.com"},
				DelayMinutes:  5,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process
	worker.Process()

	// Verify correct channel was used
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifications))
		return
	}

	if notifications[0].Channel != "email" {
		t.Errorf("Expected channel 'email', got '%s'", notifications[0].Channel)
	}

	if notifications[0].StepNumber != 0 {
		t.Errorf("Expected step number 0, got %d", notifications[0].StepNumber)
	}
}

// Test: TestWorkerSchedulesNextStep - Updates next_escalation_at for next step
func TestWorkerSchedulesNextStep(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 2 steps (5 min, 10 min delay)
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: map[string]interface{}{},
				DelayMinutes:  5,
			},
			{
				StepOrder:     1,
				ChannelType:   "slack",
				ChannelConfig: map[string]interface{}{},
				DelayMinutes:  10,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	beforeProcess := time.Now()
	worker.Process()
	afterProcess := time.Now()

	// Verify next_escalation_at is scheduled correctly (10 minutes from now)
	states, _ := db.GetPendingEscalations()
	if len(states) != 1 {
		t.Errorf("Expected 1 state, got %d", len(states))
		return
	}

	updatedState := states[0]
	if updatedState.NextEscalationAt == nil {
		t.Error("Expected NextEscalationAt to be set, got nil")
		return
	}

	expectedMin := beforeProcess.Add(10 * time.Minute)
	expectedMax := afterProcess.Add(10 * time.Minute)

	if updatedState.NextEscalationAt.Before(expectedMin) || updatedState.NextEscalationAt.After(expectedMax) {
		t.Errorf("Expected next_escalation_at ~10 minutes from now, got %v", updatedState.NextEscalationAt)
	}
}

// Test: TestWorkerAdvancesStep - Increments current_step after sending
func TestWorkerAdvancesStep(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 3 steps
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 1, ChannelType: "slack", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 2, ChannelType: "sms", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state at step 0
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process - should move to step 1
	worker.Process()

	states, _ := db.GetPendingEscalations()
	if states[0].CurrentStep != 1 {
		t.Errorf("Expected current_step 1 after first process, got %d", states[0].CurrentStep)
	}

	// Update state for second process
	states[0].NextEscalationAt = &pastTime
	db.UpdateEscalationState(states[0])

	// Process again - should move to step 2
	worker.Process()

	states, _ = db.GetPendingEscalations()
	if states[0].CurrentStep != 2 {
		t.Errorf("Expected current_step 2 after second process, got %d", states[0].CurrentStep)
	}
}

// Test: TestWorkerMarksExhausted - Sets status="exhausted" when all steps sent
func TestWorkerMarksExhausted(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 2 steps
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 1, ChannelType: "slack", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state at last step
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      2, // Already at step 2 (out of bounds)
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process
	worker.Process()

	// Verify status is set to exhausted
	states, _ := db.GetPendingEscalations()
	if len(states) == 0 {
		t.Error("Expected state to still exist (not resolved)")
		return
	}

	if states[0].Status != "exhausted" {
		t.Errorf("Expected status 'exhausted', got '%s'", states[0].Status)
	}
}

// Test: TestWorkerHandlesErrors - Continues processing on notification errors
func TestWorkerHandlesErrors(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &ErrorNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 2 steps
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 1, ChannelType: "slack", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create two escalation states, first will fail notification
	pastTime := time.Now().Add(-1 * time.Minute)
	state1 := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state1)

	state2 := &models.EscalationState{
		AlertTriggerID:   2,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state2)

	// Process - should not return error even though notifier fails
	err := worker.Process()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Both states should be updated despite notification failure
	states, _ := db.GetPendingEscalations()
	if len(states) != 2 {
		t.Errorf("Expected 2 states, got %d", len(states))
	}

	for _, state := range states {
		if state.CurrentStep != 1 {
			t.Errorf("Expected current_step 1, got %d", state.CurrentStep)
		}
	}
}

// ErrorNotifier always fails to send
type ErrorNotifier struct{}

func (e *ErrorNotifier) SendNotification(req *NotificationRequest) error {
	return fmt.Errorf("notification failed")
}

// Test: TestWorkerMultiplePending - Processes multiple pending escalations
func TestWorkerMultiplePending(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 1, ChannelType: "slack", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create 3 escalation states
	pastTime := time.Now().Add(-1 * time.Minute)
	for i := 1; i <= 3; i++ {
		state := &models.EscalationState{
			AlertTriggerID:   int64(i),
			PolicyID:         policy.ID,
			CurrentStep:      0,
			Status:           "active",
			NextEscalationAt: &pastTime,
		}
		db.CreateState(state)
	}

	// Process
	err := worker.Process()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all 3 were processed
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 3 {
		t.Errorf("Expected 3 notifications, got %d", len(notifications))
	}

	// Verify all have different alert trigger IDs
	triggerIDs := make(map[int64]bool)
	for _, notif := range notifications {
		triggerIDs[notif.AlertTriggerID] = true
	}

	if len(triggerIDs) != 3 {
		t.Errorf("Expected 3 unique alert trigger IDs, got %d", len(triggerIDs))
	}
}

// Test: TestWorkerLastEscalatedAtSet - last_escalated_at is set after notification
func TestWorkerLastEscalatedAtSet(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
		LastEscalatedAt:  nil,
	}
	db.CreateState(state)

	beforeProcess := time.Now()
	worker.Process()
	afterProcess := time.Now()

	// Verify last_escalated_at is set
	states, _ := db.GetPendingEscalations()
	if states[0].LastEscalatedAt == nil {
		t.Error("Expected LastEscalatedAt to be set, got nil")
		return
	}

	if states[0].LastEscalatedAt.Before(beforeProcess) || states[0].LastEscalatedAt.After(afterProcess) {
		t.Errorf("Expected LastEscalatedAt to be recent, got %v", states[0].LastEscalatedAt)
	}
}

// Test: TestWorkerSendsNotificationConfig - Notification contains correct config
func TestWorkerSendsNotificationConfig(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with specific config
	config := map[string]interface{}{
		"recipient": "alerts@example.com",
		"template":  "critical_alert",
	}
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{
				StepOrder:     0,
				ChannelType:   "email",
				ChannelConfig: config,
				DelayMinutes:  5,
			},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state
	pastTime := time.Now().Add(-1 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      0,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Process
	worker.Process()

	// Verify config was passed
	notifications := notifier.GetSentNotifications()
	if len(notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifications))
		return
	}

	if len(notifications[0].Config) == 0 {
		t.Error("Expected config to be passed to notification")
	}

	if notifications[0].Config["recipient"] != "alerts@example.com" {
		t.Errorf("Expected recipient 'alerts@example.com', got %v", notifications[0].Config["recipient"])
	}
}

// Test: TestWorkerNoNextEscalationAtWhenAllStepsSent - no next_escalation_at when all steps sent
func TestWorkerNoNextEscalationAtWhenAllStepsSent(t *testing.T) {
	db := NewMockWorkerDB()
	notifier := &MockWorkerNotifier{}
	worker := NewEscalationWorker(db, notifier)

	// Create policy with 2 steps
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
		Steps: []*models.EscalationPolicyStep{
			{StepOrder: 0, ChannelType: "email", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
			{StepOrder: 1, ChannelType: "slack", ChannelConfig: map[string]interface{}{}, DelayMinutes: 5},
		},
	}
	db.CreatePolicy(policy)

	// Create escalation state at last step (step 1 of 2 steps = index 1)
	pastTime := time.Now().Add(-1 * time.Minute)
	nextTime := time.Now().Add(5 * time.Minute)
	state := &models.EscalationState{
		AlertTriggerID:   1,
		PolicyID:         policy.ID,
		CurrentStep:      1,
		Status:           "active",
		NextEscalationAt: &pastTime,
	}
	db.CreateState(state)

	// Verify state before processing
	statesBefore, _ := db.GetPendingEscalations()
	if statesBefore[0].NextEscalationAt == nil {
		statesBefore[0].NextEscalationAt = &nextTime
		db.UpdateEscalationState(statesBefore[0])
	}

	// Process - should send last notification and not schedule next
	worker.Process()

	states, _ := db.GetPendingEscalations()
	if states[0].CurrentStep != 2 {
		t.Errorf("Expected current_step 2, got %d", states[0].CurrentStep)
	}

	// NextEscalationAt should be nil since no more steps to escalate to
	if states[0].NextEscalationAt != nil {
		t.Errorf("Expected NextEscalationAt to be nil, got %v", states[0].NextEscalationAt)
	}
}
