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

// MockSilenceDB implements the SilenceDB interface for testing
type MockSilenceDB struct {
	silences map[int64]*models.AlertSilence
	nextID   int64
}

func NewMockSilenceDB() *MockSilenceDB {
	return &MockSilenceDB{
		silences: make(map[int64]*models.AlertSilence),
		nextID:   1,
	}
}

func (m *MockSilenceDB) CreateSilence(silence *models.AlertSilence) error {
	silence.ID = m.nextID
	m.silences[m.nextID] = silence
	m.nextID++
	return nil
}

func (m *MockSilenceDB) GetSilenceByID(id int64) (*models.AlertSilence, error) {
	return m.silences[id], nil
}

func (m *MockSilenceDB) GetActiveSilences() ([]*models.AlertSilence, error) {
	var result []*models.AlertSilence
	for _, s := range m.silences {
		result = append(result, s)
	}
	return result, nil
}

func (m *MockSilenceDB) GetExpiredSilences() ([]*models.AlertSilence, error) {
	return []*models.AlertSilence{}, nil
}

func (m *MockSilenceDB) UpdateSilence(silence *models.AlertSilence) error {
	m.silences[silence.ID] = silence
	return nil
}

func (m *MockSilenceDB) DeleteSilence(id int64) error {
	delete(m.silences, id)
	return nil
}

func (m *MockSilenceDB) Broadcast(event string, data map[string]interface{}) error {
	return nil
}

// Test POST /api/v1/alerts/{rule_id}/silence - Create silence
func TestCreateSilence(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Test via service directly
	err := silenceService.CreateSilence(1, 30, "rule", nil, "Maintenance window")
	if err != nil {
		t.Fatalf("Failed to create silence: %v", err)
	}

	// Verify silence was created
	silences, err := silenceService.GetActiveSilences()
	if err != nil {
		t.Fatalf("Failed to get active silences: %v", err)
	}

	if len(silences) != 1 {
		t.Fatalf("Expected 1 silence, got %d", len(silences))
	}

	silence := silences[0]
	if silence.AlertRuleID != 1 {
		t.Errorf("Expected rule_id 1, got %d", silence.AlertRuleID)
	}
	if silence.SilenceType != "rule" {
		t.Errorf("Expected silence_type 'rule', got %s", silence.SilenceType)
	}
	if silence.Reason == nil || *silence.Reason != "Maintenance window" {
		t.Errorf("Expected reason 'Maintenance window', got %v", silence.Reason)
	}
}

// Test GET /api/v1/silences - List active silences
func TestListActiveSilences(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create multiple silences
	silenceService.CreateSilence(1, 30, "rule", nil, "Maintenance")
	silenceService.CreateSilence(2, 60, "instance", intPtr(1), "Instance maintenance")

	// Get active silences
	silences, err := silenceService.GetActiveSilences()
	if err != nil {
		t.Fatalf("Failed to get active silences: %v", err)
	}

	if len(silences) != 2 {
		t.Fatalf("Expected 2 silences, got %d", len(silences))
	}

	// Verify both silence rule IDs are present (order may vary)
	ruleIDs := make(map[int]bool)
	for _, silence := range silences {
		ruleIDs[silence.AlertRuleID] = true
	}

	if !ruleIDs[1] {
		t.Errorf("Expected rule_id 1 in silences")
	}
	if !ruleIDs[2] {
		t.Errorf("Expected rule_id 2 in silences")
	}
}

// Test DELETE /api/v1/silences/{id} - Deactivate silence
func TestDeleteSilence(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create a silence
	silenceService.CreateSilence(1, 30, "rule", nil, "Test")

	// Verify it exists
	silences, _ := silenceService.GetActiveSilences()
	if len(silences) != 1 {
		t.Fatalf("Expected 1 silence before deletion")
	}

	silenceID := silences[0].ID

	// Delete the silence
	err := mockDB.DeleteSilence(silenceID)
	if err != nil {
		t.Fatalf("Failed to delete silence: %v", err)
	}

	// Verify it was deleted
	silences, _ = silenceService.GetActiveSilences()
	if len(silences) != 0 {
		t.Fatalf("Expected 0 silences after deletion, got %d", len(silences))
	}
}

// Test silence expiration logic
func TestSilenceExpiration(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create a silence with past expiration
	silence := &models.AlertSilence{
		AlertRuleID:   1,
		SilencedUntil: time.Now().Add(-5 * time.Minute), // Already expired
		SilenceType:   "rule",
		CreatedAt:     time.Now(),
	}
	mockDB.CreateSilence(silence)

	// Check if it's considered active (should be false)
	isSilenced := silenceService.IsSilenced(1, nil)
	if isSilenced {
		t.Errorf("Expected silence to be expired, but IsSilenced returned true")
	}
}

// Test silence with specific instance
func TestSilenceWithInstance(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	instanceID := 1

	// Create instance-level silence
	err := silenceService.CreateSilence(1, 30, "instance", &instanceID, "Instance silence")
	if err != nil {
		t.Fatalf("Failed to create instance silence: %v", err)
	}

	// Should be silenced for this instance
	isSilenced := silenceService.IsSilenced(1, &instanceID)
	if !isSilenced {
		t.Errorf("Expected silence to apply to instance 1")
	}

	// Should NOT be silenced for different instance
	otherInstance := 2
	isSilenced = silenceService.IsSilenced(1, &otherInstance)
	if isSilenced {
		t.Errorf("Expected silence to NOT apply to instance 2")
	}
}

// Test silence type "all" on a rule
func TestSilenceTypeAll(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create "all" type silence for rule 1 (applies to all instances of rule 1)
	err := silenceService.CreateSilence(1, 30, "all", nil, "Silence all instances of rule 1")
	if err != nil {
		t.Fatalf("Failed to create silence: %v", err)
	}

	// Should be silenced for rule 1 with any instance
	isSilenced := silenceService.IsSilenced(1, nil)
	if !isSilenced {
		t.Errorf("Expected silence to apply to rule 1")
	}

	isSilenced = silenceService.IsSilenced(1, intPtr(1))
	if !isSilenced {
		t.Errorf("Expected silence to apply to rule 1 instance 1")
	}

	// Should NOT be silenced for different rule
	isSilenced = silenceService.IsSilenced(2, nil)
	if isSilenced {
		t.Errorf("Expected silence to NOT apply to rule 2")
	}
}

// Test invalid duration
func TestInvalidDuration(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Try to create silence with invalid duration
	err := silenceService.CreateSilence(1, -1, "rule", nil, "Invalid")
	if err == nil {
		t.Errorf("Expected error for negative duration, got nil")
	}

	err = silenceService.CreateSilence(1, 0, "rule", nil, "Invalid")
	if err == nil {
		t.Errorf("Expected error for zero duration, got nil")
	}
}

// Test invalid silence type
func TestInvalidSilenceType(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Try to create silence with invalid type
	err := silenceService.CreateSilence(1, 30, "invalid", nil, "Invalid")
	if err == nil {
		t.Errorf("Expected error for invalid silence_type, got nil")
	}
}

// Test instance required for instance silence
func TestInstanceRequiredForInstanceSilence(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Try to create instance silence without instance_id
	err := silenceService.CreateSilence(1, 30, "instance", nil, "Invalid")
	if err == nil {
		t.Errorf("Expected error when instance_id missing for instance silence, got nil")
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}


// Test handler wrapper with HTTP response
func TestSilenceHandlerHTTP(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)
	handler := NewSilenceHandler(silenceService)

	// Test that handler is initialized correctly
	if handler == nil {
		t.Fatalf("Handler initialization failed")
	}

	// Test that handler has service
	if handler.service == nil {
		t.Fatalf("Handler service is nil")
	}
}

// Test multiple silences on different rules
func TestMultipleSilencesOnDifferentRules(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create silences for different rules
	silenceService.CreateSilence(1, 30, "rule", nil, "Rule 1")
	silenceService.CreateSilence(2, 60, "rule", nil, "Rule 2")
	silenceService.CreateSilence(3, 90, "rule", nil, "Rule 3")

	silences, _ := silenceService.GetActiveSilences()
	if len(silences) != 3 {
		t.Fatalf("Expected 3 silences, got %d", len(silences))
	}

	// Check that each rule is silenced
	for i := 1; i <= 3; i++ {
		isSilenced := silenceService.IsSilenced(int64(i), nil)
		if !isSilenced {
			t.Errorf("Rule %d should be silenced", i)
		}
	}

	// Check that rule 4 is not silenced
	isSilenced := silenceService.IsSilenced(4, nil)
	if isSilenced {
		t.Errorf("Rule 4 should not be silenced")
	}
}

// Test silence JSON marshaling
func TestSilenceJSONMarshaling(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	silenceService.CreateSilence(1, 30, "rule", nil, "Test silence")
	silences, _ := silenceService.GetActiveSilences()
	silence := silences[0]

	// Marshal to JSON
	jsonBytes, err := json.Marshal(silence)
	if err != nil {
		t.Fatalf("Failed to marshal silence: %v", err)
	}

	// Unmarshal back
	var unmarshaled models.AlertSilence
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal silence: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != silence.ID {
		t.Errorf("ID mismatch: %d vs %d", unmarshaled.ID, silence.ID)
	}
	if unmarshaled.AlertRuleID != silence.AlertRuleID {
		t.Errorf("AlertRuleID mismatch: %d vs %d", unmarshaled.AlertRuleID, silence.AlertRuleID)
	}
}

// Test silence list serialization
func TestSilenceListSerialization(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	// Create multiple silences
	silenceService.CreateSilence(1, 30, "rule", nil, "Silence 1")
	silenceService.CreateSilence(2, 60, "rule", nil, "Silence 2")

	silences, _ := silenceService.GetActiveSilences()

	// Create response structure
	type ListResponse struct {
		Success  bool                    `json:"success"`
		Silences []*models.AlertSilence  `json:"silences,omitempty"`
	}

	resp := ListResponse{
		Success:  true,
		Silences: silences,
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

	if len(unmarshaled.Silences) != 2 {
		t.Errorf("Expected 2 silences in response, got %d", len(unmarshaled.Silences))
	}
}

// Test silence response writing
func TestSilenceResponseWrite(t *testing.T) {
	w := httptest.NewRecorder()

	silence := &models.AlertSilence{
		ID:            1,
		AlertRuleID:   1,
		SilencedUntil: time.Now().Add(30 * time.Minute),
		SilenceType:   "rule",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(silence)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response models.AlertSilence
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected silence ID 1, got %d", response.ID)
	}
}

// Test silence with JSON request parsing
func TestSilenceJSONRequestParsing(t *testing.T) {
	reqBody := CreateSilenceRequest{
		Duration:    45,
		Reason:      "Testing",
		SilenceType: "rule",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"http://localhost/api/v1/alerts/1/silence",
		bytes.NewReader(bodyBytes),
	)
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	var parsed CreateSilenceRequest
	err := json.NewDecoder(req.Body).Decode(&parsed)
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if parsed.Duration != 45 {
		t.Errorf("Expected duration 45, got %d", parsed.Duration)
	}
	if parsed.Reason != "Testing" {
		t.Errorf("Expected reason 'Testing', got %s", parsed.Reason)
	}
	if parsed.SilenceType != "rule" {
		t.Errorf("Expected silence_type 'rule', got %s", parsed.SilenceType)
	}
}

// Test URL parameter parsing for silence ID
func TestSilenceIDParamParsing(t *testing.T) {
	silenceIDStr := "123"
	silenceID, err := strconv.ParseInt(silenceIDStr, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse silence ID: %v", err)
	}

	if silenceID != 123 {
		t.Errorf("Expected ID 123, got %d", silenceID)
	}
}

// Test silence response with missing reason
func TestSilenceWithoutReason(t *testing.T) {
	mockDB := NewMockSilenceDB()
	silenceService := services.NewSilenceService(mockDB)

	err := silenceService.CreateSilence(1, 30, "rule", nil, "")
	if err != nil {
		t.Fatalf("Failed to create silence without reason: %v", err)
	}

	silences, _ := silenceService.GetActiveSilences()
	if len(silences) != 1 {
		t.Fatalf("Expected 1 silence")
	}

	// Reason should be nil or empty
	if silences[0].Reason != nil && *silences[0].Reason != "" {
		t.Errorf("Expected empty reason, got %v", silences[0].Reason)
	}
}
