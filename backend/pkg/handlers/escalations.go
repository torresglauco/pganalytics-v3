package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// EscalationHandler handles escalation policy and alert acknowledgment operations
type EscalationHandler struct {
	service *services.EscalationService
}

// NewEscalationHandler creates a new EscalationHandler
func NewEscalationHandler(service *services.EscalationService) *EscalationHandler {
	return &EscalationHandler{
		service: service,
	}
}

// CreatePolicyRequest is the request body for POST /api/v1/escalation-policies
type CreatePolicyRequest struct {
	Policy *models.EscalationPolicy `json:"policy"`
}

// CreatePolicyResponse is the response body
type CreatePolicyResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message,omitempty"`
	Error   string                   `json:"error,omitempty"`
	Policy  *models.EscalationPolicy `json:"policy,omitempty"`
}

// GetPolicyResponse is the response body for GET /api/v1/escalation-policies/{policy_id}
type GetPolicyResponse struct {
	Success bool                     `json:"success"`
	Policy  *models.EscalationPolicy `json:"policy,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// UpdatePolicyRequest is the request body for PUT /api/v1/escalation-policies/{id}
type UpdatePolicyRequest struct {
	Policy *models.EscalationPolicy `json:"policy"`
}

// UpdatePolicyResponse is the response body
type UpdatePolicyResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message,omitempty"`
	Error   string                   `json:"error,omitempty"`
	Policy  *models.EscalationPolicy `json:"policy,omitempty"`
}

// AcknowledgeAlertRequest is the request body for POST /api/v1/alerts/{trigger_id}/acknowledge
type AcknowledgeAlertRequest struct {
	// Empty body - trigger_id comes from URL
}

// AcknowledgeAlertResponse is the response body
type AcknowledgeAlertResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreatePolicy handles POST /api/v1/escalation-policies
func (eh *EscalationHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreatePolicyResponse{
			Success: false,
			Error:   "Malformed request body",
		})
		return
	}

	// Validate policy
	if req.Policy == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreatePolicyResponse{
			Success: false,
			Error:   "Policy cannot be null",
		})
		return
	}

	// Create policy
	err := eh.service.CreatePolicy(req.Policy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreatePolicyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response with generated ID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreatePolicyResponse{
		Success: true,
		Message: "Policy created successfully",
		Policy:  req.Policy,
	})
}

// GetPolicy handles GET /api/v1/escalation-policies/{policy_id}
func (eh *EscalationHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract policy_id from URL path
	policyIDStr := r.PathValue("policy_id")
	if policyIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetPolicyResponse{
			Success: false,
			Error:   "Missing policy_id in URL path",
		})
		return
	}

	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetPolicyResponse{
			Success: false,
			Error:   "Invalid policy_id format",
		})
		return
	}

	// Get policy
	policy, err := eh.service.GetPolicy(policyID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GetPolicyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetPolicyResponse{
		Success: true,
		Policy:  policy,
	})
}

// UpdatePolicy handles PUT /api/v1/escalation-policies/{id}
func (eh *EscalationHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract id from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdatePolicyResponse{
			Success: false,
			Error:   "Missing id in URL path",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdatePolicyResponse{
			Success: false,
			Error:   "Invalid id format",
		})
		return
	}

	// Parse request body
	var req UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdatePolicyResponse{
			Success: false,
			Error:   "Malformed request body",
		})
		return
	}

	// Validate policy
	if req.Policy == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdatePolicyResponse{
			Success: false,
			Error:   "Policy cannot be null",
		})
		return
	}

	// Set ID for update
	req.Policy.ID = id

	// Update policy
	err = eh.service.UpdatePolicy(req.Policy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdatePolicyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdatePolicyResponse{
		Success: true,
		Message: "Policy updated successfully",
		Policy:  req.Policy,
	})
}

// AcknowledgeAlert handles POST /api/v1/alerts/{trigger_id}/acknowledge
func (eh *EscalationHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract trigger_id from URL path
	triggerIDStr := r.PathValue("trigger_id")
	if triggerIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcknowledgeAlertResponse{
			Success: false,
			Error:   "Missing trigger_id in URL path",
		})
		return
	}

	triggerID, err := strconv.ParseInt(triggerIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcknowledgeAlertResponse{
			Success: false,
			Error:   "Invalid trigger_id format",
		})
		return
	}

	// TODO: Extract user ID from context/auth (for now use 1 as placeholder)
	userID := 1

	// Acknowledge alert
	err = eh.service.AcknowledgeAlert(triggerID, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcknowledgeAlertResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AcknowledgeAlertResponse{
		Success: true,
		Status:  "acknowledged",
		Message: "Alert acknowledged successfully",
	})
}
