// backend/pkg/services/escalation_service_test.go
package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// MockEscalationDB implements the EscalationDB interface for testing
type MockEscalationDB struct {
	policies         map[int64]*models.EscalationPolicy
	escalationStates map[int64]*models.EscalationState
	nextPolicyID     int64
	nextStateID      int64
	mu               sync.Mutex
}

// NewMockEscalationDB creates a new MockEscalationDB
func NewMockEscalationDB() *MockEscalationDB {
	return &MockEscalationDB{
		policies:         make(map[int64]*models.EscalationPolicy),
		escalationStates: make(map[int64]*models.EscalationState),
		nextPolicyID:     1,
		nextStateID:      1,
	}
}

// CreatePolicy creates a new escalation policy in the mock database
func (m *MockEscalationDB) CreatePolicy(policy *models.EscalationPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy.ID = m.nextPolicyID
	m.nextPolicyID++
	m.policies[policy.ID] = policy
	return nil
}

// GetPolicy retrieves a policy by ID
func (m *MockEscalationDB) GetPolicy(id int64) (*models.EscalationPolicy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy, exists := m.policies[id]
	if !exists {
		return nil, nil
	}
	return policy, nil
}

// UpdatePolicy updates a policy
func (m *MockEscalationDB) UpdatePolicy(policy *models.EscalationPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.policies[policy.ID]; !exists {
		return fmt.Errorf("policy not found")
	}
	m.policies[policy.ID] = policy
	return nil
}

// ListPolicies returns all policies
func (m *MockEscalationDB) ListPolicies() ([]*models.EscalationPolicy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var policies []*models.EscalationPolicy
	for _, policy := range m.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}

// CreateEscalationState creates a new escalation state in the mock database
func (m *MockEscalationDB) CreateEscalationState(state *models.EscalationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state.ID = m.nextStateID
	m.nextStateID++
	m.escalationStates[state.AlertTriggerID] = state
	return nil
}

// UpdateEscalationState updates an escalation state
func (m *MockEscalationDB) UpdateEscalationState(state *models.EscalationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.escalationStates[state.AlertTriggerID]; !exists {
		return fmt.Errorf("escalation state not found")
	}
	m.escalationStates[state.AlertTriggerID] = state
	return nil
}

// GetEscalationState retrieves an escalation state by trigger ID
func (m *MockEscalationDB) GetEscalationState(triggerID int64) (*models.EscalationState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.escalationStates[triggerID]
	if !exists {
		return nil, nil
	}
	return state, nil
}

// GetPendingEscalations returns escalations ready to be executed
func (m *MockEscalationDB) GetPendingEscalations() ([]*models.EscalationState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var pending []*models.EscalationState

	for _, state := range m.escalationStates {
		// Include states with status='pending' and no next escalation time set,
		// or where next escalation time has passed
		if state.Status == "pending" {
			if state.NextEscalationAt == nil || state.NextEscalationAt.Before(now) || state.NextEscalationAt.Equal(now) {
				pending = append(pending, state)
			}
		}
	}

	return pending, nil
}

// MockNotifier implements the Notifier interface for testing
type MockNotifier struct {
	notifications []*NotificationRequest
	mu            sync.Mutex
}

// SendNotification sends a notification and records it
func (m *MockNotifier) SendNotification(req *NotificationRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.notifications = append(m.notifications, req)
	return nil
}

// GetNotifications returns all recorded notifications
func (m *MockNotifier) GetNotifications() []*NotificationRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.notifications
}

// Test: TestCreatePolicy - Creates policy successfully
func TestCreatePolicy(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	policy := &models.EscalationPolicy{
		Name:        "Critical Alert Escalation",
		Description: pointToString("Policy for critical alerts"),
	}

	err := service.CreatePolicy(policy)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if policy.ID != 1 {
		t.Errorf("Expected policy ID 1, got %d", policy.ID)
	}

	if !policy.IsActive {
		t.Error("Expected policy to be active")
	}
}

// Test: TestCreatePolicy_NoName - Rejects policy without name
func TestCreatePolicy_NoName(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	policy := &models.EscalationPolicy{
		Name: "",
	}

	err := service.CreatePolicy(policy)
	if err == nil {
		t.Error("Expected error when creating policy without name")
	}
}

// Test: TestGetPolicy - Retrieves policy with steps
func TestGetPolicy(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name:        "Test Policy",
		Description: pointToString("Test Description"),
	}

	err := service.CreatePolicy(policy)
	if err != nil {
		t.Errorf("Failed to create policy: %v", err)
	}

	// Retrieve the policy
	retrieved, err := service.GetPolicy(policy.ID)
	if err != nil {
		t.Errorf("Failed to retrieve policy: %v", err)
	}

	if retrieved.ID != policy.ID {
		t.Errorf("Expected policy ID %d, got %d", policy.ID, retrieved.ID)
	}

	if retrieved.Name != policy.Name {
		t.Errorf("Expected policy name '%s', got '%s'", policy.Name, retrieved.Name)
	}
}

// Test: TestGetPolicy_NotFound - Returns error for non-existent policy
func TestGetPolicy_NotFound(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	_, err := service.GetPolicy(999)
	if err == nil {
		t.Error("Expected error when retrieving non-existent policy")
	}
}

// Test: TestUpdatePolicy - Updates policy fields
func TestUpdatePolicy(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name:        "Original Name",
		Description: pointToString("Original Description"),
	}

	err := service.CreatePolicy(policy)
	if err != nil {
		t.Errorf("Failed to create policy: %v", err)
	}

	originalID := policy.ID

	// Update the policy
	policy.Name = "Updated Name"
	policy.Description = pointToString("Updated Description")

	err = service.UpdatePolicy(policy)
	if err != nil {
		t.Errorf("Failed to update policy: %v", err)
	}

	// Verify update
	retrieved, _ := service.GetPolicy(originalID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}

	if *retrieved.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", *retrieved.Description)
	}
}

// Test: TestListPolicies - Returns all policies
func TestListPolicies(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create multiple policies
	for i := 1; i <= 3; i++ {
		policy := &models.EscalationPolicy{
			Name: fmt.Sprintf("Policy %d", i),
		}
		service.CreatePolicy(policy)
	}

	// List all policies
	policies, err := service.ListPolicies()
	if err != nil {
		t.Errorf("Failed to list policies: %v", err)
	}

	if len(policies) != 3 {
		t.Errorf("Expected 3 policies, got %d", len(policies))
	}
}

// Test: TestStartEscalation - Creates escalation_state with initial values
func TestStartEscalation(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create a policy first
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)

	// Start escalation
	err := service.StartEscalation(100, policy.ID)
	if err != nil {
		t.Errorf("Failed to start escalation: %v", err)
	}

	// Verify escalation state
	state, _ := service.GetEscalationState(100)
	if state.AlertTriggerID != 100 {
		t.Errorf("Expected alert trigger ID 100, got %d", state.AlertTriggerID)
	}

	if state.PolicyID != policy.ID {
		t.Errorf("Expected policy ID %d, got %d", policy.ID, state.PolicyID)
	}
}

// Test: TestStartEscalation_InitialValues - Verifies initial_step=0, status='pending', ack_received=false
func TestStartEscalation_InitialValues(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)

	// Start escalation
	service.StartEscalation(200, policy.ID)

	// Verify initial values
	state, _ := service.GetEscalationState(200)
	if state.CurrentStep != 0 {
		t.Errorf("Expected current_step=0, got %d", state.CurrentStep)
	}

	if state.Status != "pending" {
		t.Errorf("Expected status='pending', got '%s'", state.Status)
	}

	if state.AckReceived {
		t.Error("Expected ack_received=false")
	}

	if state.AckBy != nil {
		t.Error("Expected ack_by=nil")
	}

	if state.AckAt != nil {
		t.Error("Expected ack_at=nil")
	}
}

// Test: TestStartEscalation_PolicyNotFound - Returns error when policy doesn't exist
func TestStartEscalation_PolicyNotFound(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	err := service.StartEscalation(100, 999)
	if err == nil {
		t.Error("Expected error when policy doesn't exist")
	}
}

// Test: TestAcknowledgeAlert - Marks alert acknowledged and stops escalation
func TestAcknowledgeAlert(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)
	service.StartEscalation(300, policy.ID)

	// Acknowledge alert
	userID := 42
	err := service.AcknowledgeAlert(300, userID)
	if err != nil {
		t.Errorf("Failed to acknowledge alert: %v", err)
	}

	// Verify acknowledgment
	state, _ := service.GetEscalationState(300)
	if !state.AckReceived {
		t.Error("Expected ack_received=true")
	}

	if state.AckBy == nil || *state.AckBy != userID {
		t.Errorf("Expected ack_by=%d, got %v", userID, state.AckBy)
	}

	if state.AckAt == nil {
		t.Error("Expected ack_at to be set")
	}

	if state.Status != "acknowledged" {
		t.Errorf("Expected status='acknowledged', got '%s'", state.Status)
	}
}

// Test: TestAcknowledgeAlert_VerifyFields - Verifies ack_by, ack_at, status='acknowledged'
func TestAcknowledgeAlert_VerifyFields(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)
	service.StartEscalation(400, policy.ID)

	// Acknowledge before timestamp
	beforeAck := time.Now()
	userID := 99
	service.AcknowledgeAlert(400, userID)
	afterAck := time.Now()

	// Verify fields
	state, _ := service.GetEscalationState(400)

	if state.AckBy == nil || *state.AckBy != userID {
		t.Errorf("Expected ack_by=%d, got %v", userID, state.AckBy)
	}

	if state.AckAt == nil {
		t.Error("Expected ack_at to be set")
	} else if state.AckAt.Before(beforeAck) || state.AckAt.After(afterAck.Add(1*time.Second)) {
		t.Errorf("Expected ack_at to be within the test window")
	}

	if state.Status != "acknowledged" {
		t.Errorf("Expected status='acknowledged', got '%s'", state.Status)
	}
}

// Test: TestUpdateEscalationState - State updates persist
func TestUpdateEscalationState(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)
	service.StartEscalation(500, policy.ID)

	// Get state and modify it
	state, _ := service.GetEscalationState(500)
	state.CurrentStep = 2
	state.Status = "escalating"

	// Update state
	err := service.UpdateEscalationState(state)
	if err != nil {
		t.Errorf("Failed to update escalation state: %v", err)
	}

	// Verify update
	updated, _ := service.GetEscalationState(500)
	if updated.CurrentStep != 2 {
		t.Errorf("Expected current_step=2, got %d", updated.CurrentStep)
	}

	if updated.Status != "escalating" {
		t.Errorf("Expected status='escalating', got '%s'", updated.Status)
	}
}

// Test: TestGetPendingEscalations - Returns pending escalations only
func TestGetPendingEscalations(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)

	// Create multiple escalation states
	// Escalation 1: pending, no next escalation time set
	service.StartEscalation(600, policy.ID)

	// Escalation 2: pending, next escalation time in the past
	service.StartEscalation(601, policy.ID)
	state2, _ := db.GetEscalationState(601)
	pastTime := time.Now().Add(-1 * time.Hour)
	state2.NextEscalationAt = &pastTime
	db.UpdateEscalationState(state2)

	// Escalation 3: acknowledged (should not be included)
	service.StartEscalation(602, policy.ID)
	service.AcknowledgeAlert(602, 1)

	// Escalation 4: pending, next escalation time in the future
	service.StartEscalation(603, policy.ID)
	state4, _ := db.GetEscalationState(603)
	futureTime := time.Now().Add(1 * time.Hour)
	state4.NextEscalationAt = &futureTime
	db.UpdateEscalationState(state4)

	// Get pending escalations
	pending, err := service.GetPendingEscalations()
	if err != nil {
		t.Errorf("Failed to get pending escalations: %v", err)
	}

	// Should return 2 pending escalations (600 and 601)
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending escalations, got %d", len(pending))
	}

	// Verify they are the correct ones
	pendingTriggerIDs := make(map[int64]bool)
	for _, p := range pending {
		pendingTriggerIDs[p.AlertTriggerID] = true
	}

	if !pendingTriggerIDs[600] || !pendingTriggerIDs[601] {
		t.Error("Expected pending escalations for triggers 600 and 601")
	}
}

// Test: TestGetEscalationState - Retrieves current state of escalation
func TestGetEscalationState(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)
	service.StartEscalation(700, policy.ID)

	// Get escalation state
	state, err := service.GetEscalationState(700)
	if err != nil {
		t.Errorf("Failed to get escalation state: %v", err)
	}

	if state.AlertTriggerID != 700 {
		t.Errorf("Expected trigger ID 700, got %d", state.AlertTriggerID)
	}
}

// Test: TestGetEscalationState_NotFound - Returns error for non-existent escalation
func TestGetEscalationState_NotFound(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	_, err := service.GetEscalationState(999)
	if err == nil {
		t.Error("Expected error when escalation state doesn't exist")
	}
}

// Test: TestMultipleEscalations - Multiple escalations work correctly
func TestMultipleEscalations(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name: "Test Policy",
	}
	service.CreatePolicy(policy)

	// Start multiple escalations
	for i := 1; i <= 5; i++ {
		triggerID := int64(800 + i)
		err := service.StartEscalation(triggerID, policy.ID)
		if err != nil {
			t.Errorf("Failed to start escalation %d: %v", triggerID, err)
		}
	}

	// Verify all escalations exist
	for i := 1; i <= 5; i++ {
		triggerID := int64(800 + i)
		state, err := service.GetEscalationState(triggerID)
		if err != nil {
			t.Errorf("Failed to get escalation state for trigger %d: %v", triggerID, err)
		}

		if state.AlertTriggerID != triggerID {
			t.Errorf("Expected trigger ID %d, got %d", triggerID, state.AlertTriggerID)
		}
	}
}

// Test: TestUpdatePolicy_EmptyName - Rejects update with empty name
func TestUpdatePolicy_EmptyName(t *testing.T) {
	db := NewMockEscalationDB()
	notifier := &MockNotifier{}
	service := NewEscalationService(db, notifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name: "Original Name",
	}
	service.CreatePolicy(policy)

	// Try to update with empty name
	policy.Name = ""
	err := service.UpdatePolicy(policy)
	if err == nil {
		t.Error("Expected error when updating policy with empty name")
	}
}

// Helper function to create a string pointer
func pointToString(s string) *string {
	return &s
}
