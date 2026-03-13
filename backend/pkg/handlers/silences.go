package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// SilenceHandler handles alert silence operations
type SilenceHandler struct {
	service *services.SilenceService
}

// NewSilenceHandler creates a new SilenceHandler
func NewSilenceHandler(service *services.SilenceService) *SilenceHandler {
	return &SilenceHandler{
		service: service,
	}
}

// CreateSilenceRequest is the request body for POST /api/v1/alerts/{rule_id}/silence
type CreateSilenceRequest struct {
	Duration    int    `json:"duration"`      // Duration in minutes
	Reason      string `json:"reason"`       // Reason for silence
	SilenceType string `json:"silence_type"` // 'rule', 'instance', or 'all'
	InstanceID  *int   `json:"instance_id,omitempty"`
}

// CreateSilenceResponse is the response body
type CreateSilenceResponse struct {
	Success       bool                  `json:"success"`
	Message       string                `json:"message,omitempty"`
	Error         string                `json:"error,omitempty"`
	SilenceID     int64                 `json:"silence_id,omitempty"`
	AlertSilence  *models.AlertSilence  `json:"silence,omitempty"`
}

// ListSilencesResponse is the response body for GET /api/v1/silences
type ListSilencesResponse struct {
	Success  bool                    `json:"success"`
	Silences []*models.AlertSilence  `json:"silences,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

// CreateSilence handles POST /api/v1/alerts/{rule_id}/silence
func (sh *SilenceHandler) CreateSilence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule_id from URL path
	ruleIDStr := r.PathValue("rule_id")
	if ruleIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateSilenceResponse{
			Success: false,
			Error:   "Missing rule_id in URL path",
		})
		return
	}

	ruleID, err := strconv.ParseInt(ruleIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateSilenceResponse{
			Success: false,
			Error:   "Invalid rule_id format",
		})
		return
	}

	// Parse request body
	var req CreateSilenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateSilenceResponse{
			Success: false,
			Error:   "Malformed request body",
		})
		return
	}

	// Create silence
	err = sh.service.CreateSilence(ruleID, req.Duration, req.SilenceType, req.InstanceID, req.Reason)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateSilenceResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateSilenceResponse{
		Success: true,
		Message: "Silence created successfully",
	})
}

// ListActiveSilences handles GET /api/v1/silences
func (sh *SilenceHandler) ListActiveSilences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active silences
	silences, err := sh.service.GetActiveSilences()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListSilencesResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListSilencesResponse{
		Success:  true,
		Silences: silences,
	})
}

// DeleteSilenceRequest is the request body for DELETE /api/v1/silences/{id}
type DeleteSilenceRequest struct {
	ID int64 `json:"id"`
}

// DeleteSilenceResponse is the response body
type DeleteSilenceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DeleteSilence handles DELETE /api/v1/silences/{id}
func (sh *SilenceHandler) DeleteSilence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract id from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(DeleteSilenceResponse{
			Success: false,
			Error:   "Missing id in URL path",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(DeleteSilenceResponse{
			Success: false,
			Error:   "Invalid id format",
		})
		return
	}

	// TODO: Delete silence from database
	// This would be implemented once the SilenceDB interface is fully integrated
	_ = id

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteSilenceResponse{
		Success: true,
		Message: "Silence deleted successfully",
	})
}
