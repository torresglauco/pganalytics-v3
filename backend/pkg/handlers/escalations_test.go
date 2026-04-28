package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// MockEscalationDB implements the EscalationDB interface for testing
type MockEscalationDB struct {
	policies map[int64]*models.EscalationPolicy
	states   map[int64]*models.EscalationState
	nextID   int64
}

func NewMockEscalationDB() *MockEscalationDB {
	return &MockEscalationDB{
		policies: make(map[int64]*models.EscalationPolicy),
		states:   make(map[int64]*models.EscalationState),
		nextID:   1,
	}
}

func (m *MockEscalationDB) CreatePolicy(policy *models.EscalationPolicy) error {
	policy.ID = m.nextID
	m.policies[m.nextID] = policy
	m.nextID++
	return nil
}

func (m *MockEscalationDB) GetPolicy(id int64) (*models.EscalationPolicy, error) {
	policy, exists := m.policies[id]
	if !exists {
		return nil, nil
	}
	return policy, nil
}

func (m *MockEscalationDB) UpdatePolicy(policy *models.EscalationPolicy) error {
	m.policies[policy.ID] = policy
	return nil
}

func (m *MockEscalationDB) ListPolicies() ([]*models.EscalationPolicy, error) {
	var result []*models.EscalationPolicy
	for _, p := range m.policies {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockEscalationDB) CreateEscalationState(state *models.EscalationState) error {
	state.ID = m.nextID
	m.states[m.nextID] = state
	m.nextID++
	return nil
}

func (m *MockEscalationDB) UpdateEscalationState(state *models.EscalationState) error {
	m.states[state.ID] = state
	return nil
}

func (m *MockEscalationDB) GetEscalationState(triggerID int64) (*models.EscalationState, error) {
	for _, state := range m.states {
		if state.AlertTriggerID == triggerID {
			return state, nil
		}
	}
	return nil, nil
}

func (m *MockEscalationDB) GetPendingEscalations() ([]*models.EscalationState, error) {
	var result []*models.EscalationState
	for _, state := range m.states {
		if state.Status == "pending" {
			result = append(result, state)
		}
	}
	return result, nil
}

// Mock Notifier for testing
type MockNotifier struct {
	notifications []*services.NotificationRequest
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		notifications: []*services.NotificationRequest{},
	}
}

func (m *MockNotifier) SendNotification(req *services.NotificationRequest) error {
	m.notifications = append(m.notifications, req)
	return nil
}

// Test POST /api/v1/escalation-policies - Create policy
func TestCreateEscalationPolicy(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name:        "Critical Alert Policy",
		Description: stringPtr("Policy for critical alerts"),
		IsActive:    true,
	}

	err := escalationService.CreatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}

	if policy.ID == 0 {
		t.Errorf("Policy ID should be set after creation")
	}

	// Verify policy was created
	retrieved, err := escalationService.GetPolicy(policy.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve policy: %v", err)
	}

	if retrieved.Name != "Critical Alert Policy" {
		t.Errorf("Expected policy name 'Critical Alert Policy', got %s", retrieved.Name)
	}
}

// Test GET /api/v1/escalation-policies - List policies
func TestListEscalationPolicies(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create multiple policies
	policy1 := &models.EscalationPolicy{
		Name:        "Policy 1",
		Description: stringPtr("First policy"),
		IsActive:    true,
	}
	escalationService.CreatePolicy(policy1)

	policy2 := &models.EscalationPolicy{
		Name:        "Policy 2",
		Description: stringPtr("Second policy"),
		IsActive:    true,
	}
	escalationService.CreatePolicy(policy2)

	// List all policies
	policies, err := escalationService.ListPolicies()
	if err != nil {
		t.Fatalf("Failed to list policies: %v", err)
	}

	if len(policies) != 2 {
		t.Fatalf("Expected 2 policies, got %d", len(policies))
	}

	if policies[0].Name != "Policy 1" {
		t.Errorf("Expected first policy name 'Policy 1', got %s", policies[0].Name)
	}
	if policies[1].Name != "Policy 2" {
		t.Errorf("Expected second policy name 'Policy 2', got %s", policies[1].Name)
	}
}

// Test GET /api/v1/escalation-policies/{id} - Get policy details
func TestGetEscalationPolicy(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name:        "Test Policy",
		Description: stringPtr("Test description"),
		IsActive:    true,
	}
	escalationService.CreatePolicy(policy)

	// Retrieve the policy
	retrieved, err := escalationService.GetPolicy(policy.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve policy: %v", err)
	}

	if retrieved == nil {
		t.Fatalf("Retrieved policy is nil")
	}

	if retrieved.Name != "Test Policy" {
		t.Errorf("Expected policy name 'Test Policy', got %s", retrieved.Name)
	}

	if retrieved.Description == nil || *retrieved.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %v", retrieved.Description)
	}
}

// Test PUT /api/v1/escalation-policies/{id} - Update policy
func TestUpdateEscalationPolicy(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create a policy
	policy := &models.EscalationPolicy{
		Name:        "Original Name",
		Description: stringPtr("Original description"),
		IsActive:    true,
	}
	escalationService.CreatePolicy(policy)

	// Update the policy
	policy.Name = "Updated Name"
	policy.Description = stringPtr("Updated description")

	err := escalationService.UpdatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Verify update
	retrieved, _ := escalationService.GetPolicy(policy.ID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Description == nil || *retrieved.Description != "Updated description" {
		t.Errorf("Expected updated description, got %v", retrieved.Description)
	}
}

// Test policy validation - empty name
func TestEscalationPolicyEmptyName(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Try to create policy with empty name
	policy := &models.EscalationPolicy{
		Name:     "",
		IsActive: true,
	}

	err := escalationService.CreatePolicy(policy)
	if err == nil {
		t.Errorf("Expected error for empty policy name, got nil")
	}
}

// Test escalation state creation
func TestStartEscalation(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create a policy first
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy)

	// Start escalation for an alert trigger
	err := escalationService.StartEscalation(1, policy.ID)
	if err != nil {
		t.Fatalf("Failed to start escalation: %v", err)
	}

	// Verify escalation state was created
	state, err := escalationService.GetEscalationState(1)
	if err != nil {
		t.Fatalf("Failed to retrieve escalation state: %v", err)
	}

	if state == nil {
		t.Fatalf("Escalation state is nil")
	}

	if state.AlertTriggerID != 1 {
		t.Errorf("Expected trigger ID 1, got %d", state.AlertTriggerID)
	}
	if state.PolicyID != policy.ID {
		t.Errorf("Expected policy ID %d, got %d", policy.ID, state.PolicyID)
	}
	if state.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", state.Status)
	}
	if state.AckReceived {
		t.Errorf("Expected AckReceived to be false")
	}
}

// Test acknowledge alert
func TestAcknowledgeAlert(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy)
	escalationService.StartEscalation(1, policy.ID)

	// Acknowledge the alert
	err := escalationService.AcknowledgeAlert(1, 42)
	if err != nil {
		t.Fatalf("Failed to acknowledge alert: %v", err)
	}

	// Verify state was updated
	state, _ := escalationService.GetEscalationState(1)
	if !state.AckReceived {
		t.Errorf("Expected AckReceived to be true after acknowledgment")
	}
	if state.AckBy == nil || *state.AckBy != 42 {
		t.Errorf("Expected AckBy to be 42, got %v", state.AckBy)
	}
	if state.Status != "acknowledged" {
		t.Errorf("Expected status 'acknowledged', got %s", state.Status)
	}
}

// Test non-existent policy
func TestGetNonExistentPolicy(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Try to get non-existent policy
	policy, err := escalationService.GetPolicy(999)
	if err == nil {
		t.Errorf("Expected error for non-existent policy, got nil")
	}
	if policy != nil {
		t.Errorf("Expected nil policy for non-existent ID")
	}
}

// Test escalation for non-existent policy
func TestEscalationWithNonExistentPolicy(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Try to start escalation with non-existent policy
	err := escalationService.StartEscalation(1, 999)
	if err == nil {
		t.Errorf("Expected error for non-existent policy, got nil")
	}
}

// Test get pending escalations
func TestGetPendingEscalations(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create policy
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy)

	// Start multiple escalations
	escalationService.StartEscalation(1, policy.ID)
	escalationService.StartEscalation(2, policy.ID)

	// Acknowledge one
	escalationService.AcknowledgeAlert(1, 42)

	// Get pending escalations
	pending, err := escalationService.GetPendingEscalations()
	if err != nil {
		t.Fatalf("Failed to get pending escalations: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("Expected 1 pending escalation, got %d", len(pending))
	}

	if pending[0].AlertTriggerID != 2 {
		t.Errorf("Expected pending escalation for trigger 2, got %d", pending[0].AlertTriggerID)
	}
}

// Test policy JSON marshaling
func TestPolicyJSONMarshaling(t *testing.T) {
	policy := &models.EscalationPolicy{
		ID:          1,
		Name:        "Test Policy",
		Description: stringPtr("Test description"),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("Failed to marshal policy: %v", err)
	}

	// Unmarshal back
	var unmarshaled models.EscalationPolicy
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal policy: %v", err)
	}

	if unmarshaled.Name != "Test Policy" {
		t.Errorf("Expected name 'Test Policy', got %s", unmarshaled.Name)
	}
	if unmarshaled.ID != 1 {
		t.Errorf("Expected ID 1, got %d", unmarshaled.ID)
	}
}

// Test policy list serialization
func TestPolicyListSerialization(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create multiple policies
	policy1 := &models.EscalationPolicy{
		Name:     "Policy 1",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy1)

	policy2 := &models.EscalationPolicy{
		Name:     "Policy 2",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy2)

	policies, _ := escalationService.ListPolicies()

	// Create response structure
	type ListResponse struct {
		Success  bool                       `json:"success"`
		Policies []*models.EscalationPolicy `json:"policies,omitempty"`
	}

	resp := ListResponse{
		Success:  true,
		Policies: policies,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal back
	var unmarshaled ListResponse
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(unmarshaled.Policies) != 2 {
		t.Errorf("Expected 2 policies in response, got %d", len(unmarshaled.Policies))
	}
}

// Test policy response writing
func TestPolicyResponseWrite(t *testing.T) {
	w := httptest.NewRecorder()

	policy := &models.EscalationPolicy{
		ID:          1,
		Name:        "Test Policy",
		Description: stringPtr("Test"),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response models.EscalationPolicy
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "Test Policy" {
		t.Errorf("Expected policy name 'Test Policy', got %s", response.Name)
	}
}

// Test policy JSON request parsing
func TestPolicyJSONRequestParsing(t *testing.T) {
	reqBody := CreatePolicyRequest{
		Policy: &models.EscalationPolicy{
			Name:        "Test Policy",
			Description: stringPtr("Test"),
			IsActive:    true,
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"http://localhost/api/v1/escalation-policies",
		bytes.NewReader(bodyBytes),
	)
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	var parsed CreatePolicyRequest
	err := json.NewDecoder(req.Body).Decode(&parsed)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if parsed.Policy.Name != "Test Policy" {
		t.Errorf("Expected policy name 'Test Policy', got %s", parsed.Policy.Name)
	}
}

// Test URL parameter parsing for policy ID
func TestPolicyIDParamParsing(t *testing.T) {
	policyIDStr := "456"
	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse policy ID: %v", err)
	}

	if policyID != 456 {
		t.Errorf("Expected ID 456, got %d", policyID)
	}
}

// Test handler initialization
func TestEscalationHandlerHTTP(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)
	handler := NewEscalationHandler(escalationService)

	// Test that handler is initialized correctly
	if handler == nil {
		t.Fatalf("Handler initialization failed")
	}

	// Test that handler has service
	if handler.service == nil {
		t.Fatalf("Handler service is nil")
	}
}

// Test policy is created as active
func TestPolicyCreatedAsActive(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create policy (even if we try to set IsActive to false, it will be overridden to true)
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: false, // This will be overridden
	}
	escalationService.CreatePolicy(policy)

	// Retrieve and verify - should be active
	retrieved, _ := escalationService.GetPolicy(policy.ID)
	if !retrieved.IsActive {
		t.Errorf("Expected policy to be active after creation")
	}
}

// Test multiple escalations for same trigger
func TestAcknowledgeAlert_AlreadyAcknowledged(t *testing.T) {
	mockDB := NewMockEscalationDB()
	mockNotifier := NewMockNotifier()
	escalationService := services.NewEscalationService(mockDB, mockNotifier)

	// Create policy and start escalation
	policy := &models.EscalationPolicy{
		Name:     "Test Policy",
		IsActive: true,
	}
	escalationService.CreatePolicy(policy)
	escalationService.StartEscalation(1, policy.ID)

	// Acknowledge the alert
	err := escalationService.AcknowledgeAlert(1, 42)
	if err != nil {
		t.Fatalf("Failed to acknowledge alert: %v", err)
	}

	// Acknowledge again (should not error, just update)
	err = escalationService.AcknowledgeAlert(1, 43)
	if err != nil {
		t.Fatalf("Failed to acknowledge alert second time: %v", err)
	}

	// Verify latest ack_by is 43
	state, _ := escalationService.GetEscalationState(1)
	if state.AckBy == nil || *state.AckBy != 43 {
		t.Errorf("Expected latest AckBy to be 43, got %v", state.AckBy)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
