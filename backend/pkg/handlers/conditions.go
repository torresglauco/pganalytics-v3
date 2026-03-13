package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// ConditionHandler handles alert condition validation
type ConditionHandler struct {
	validator *services.ConditionValidator
}

// NewConditionHandler creates a new ConditionHandler
func NewConditionHandler(validator *services.ConditionValidator) *ConditionHandler {
	return &ConditionHandler{
		validator: validator,
	}
}

// ValidateConditionRequest is the request body for POST /api/v1/alert-rules/validate
type ValidateConditionRequest struct {
	Condition models.AlertCondition `json:"condition"`
}

// ValidateConditionResponse is the response body
type ValidateConditionResponse struct {
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	DisplayText string `json:"display_text,omitempty"`
}

// ValidateCondition handles POST /api/v1/alert-rules/validate
func (ch *ConditionHandler) ValidateCondition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ValidateConditionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateConditionResponse{
			Valid: false,
			Error: "Malformed request body",
		})
		return
	}

	// Validate condition
	err := ch.validator.Validate(req.Condition)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateConditionResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	// Generate display text
	displayText := ch.validator.ToDisplayText(req.Condition)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateConditionResponse{
		Valid:       true,
		DisplayText: displayText,
	})
}
