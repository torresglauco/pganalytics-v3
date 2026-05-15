package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// AlertRulesHandler handles alert rule CRUD operations
type AlertRulesHandler struct {
	repo      *storage.AlertRulesRepository
	validator *services.ConditionValidator
}

// NewAlertRulesHandler creates a new AlertRulesHandler
func NewAlertRulesHandler(repo *storage.AlertRulesRepository, validator *services.ConditionValidator) *AlertRulesHandler {
	return &AlertRulesHandler{
		repo:      repo,
		validator: validator,
	}
}

// CreateAlertRuleRequest is the request body for POST /api/v1/alert-rules
type CreateAlertRuleRequest struct {
	Rule *storage.AlertRule `json:"rule"`
}

// CreateAlertRuleResponse is the response body
type CreateAlertRuleResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
	Rule    *storage.AlertRule `json:"rule,omitempty"`
}

// ListAlertRulesResponse is the response for GET /api/v1/alert-rules
type ListAlertRulesResponse struct {
	Success bool                 `json:"success"`
	Rules   []*storage.AlertRule `json:"rules,omitempty"`
	Total   int                  `json:"total"`
	Error   string               `json:"error,omitempty"`
}

// GetAlertRuleResponse is the response for GET /api/v1/alert-rules/:id
type GetAlertRuleResponse struct {
	Success bool               `json:"success"`
	Rule    *storage.AlertRule `json:"rule,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// UpdateAlertRuleRequest is the request body for PUT /api/v1/alert-rules/:id
type UpdateAlertRuleRequest struct {
	Rule *storage.AlertRule `json:"rule"`
}

// UpdateAlertRuleResponse is the response body
type UpdateAlertRuleResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
	Rule    *storage.AlertRule `json:"rule,omitempty"`
}

// DeleteAlertRuleResponse is the response for DELETE /api/v1/alert-rules/:id
type DeleteAlertRuleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AlertHistoryResponse is the response for GET /api/v1/alerts/history
type AlertHistoryResponse struct {
	Success  bool                    `json:"success"`
	Triggers []*storage.AlertTrigger `json:"triggers,omitempty"`
	Total    int                     `json:"total"`
	Error    string                  `json:"error,omitempty"`
}

// CreateAlertRule handles POST /api/v1/alert-rules
func (h *AlertRulesHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req CreateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Malformed request body",
		})
		return
	}

	// Validate rule
	if req.Rule == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Rule cannot be null",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Rule.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Rule name is required",
		})
		return
	}

	if req.Rule.RuleType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Rule type is required",
		})
		return
	}

	// Validate rule type
	validRuleTypes := map[string]bool{
		"threshold": true,
		"change":    true,
		"anomaly":   true,
		"composite": true,
	}
	if !validRuleTypes[req.Rule.RuleType] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Invalid rule type. Valid types: threshold, change, anomaly, composite",
		})
		return
	}

	// Validate severity
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	if req.Rule.AlertSeverity == "" {
		req.Rule.AlertSeverity = "medium"
	}
	if !validSeverities[req.Rule.AlertSeverity] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Invalid severity. Valid values: low, medium, high, critical",
		})
		return
	}

	// Validate condition JSON using ConditionValidator if condition is present
	if len(req.Rule.Condition) > 0 {
		var condition models.AlertCondition
		if err := json.Unmarshal(req.Rule.Condition, &condition); err == nil {
			if err := h.validator.Validate(condition); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(CreateAlertRuleResponse{
					Success: false,
					Error:   "Invalid condition: " + err.Error(),
				})
				return
			}
		}
	}

	// Set defaults
	if req.Rule.EvaluationInterval == 0 {
		req.Rule.EvaluationInterval = 300 // 5 minutes default
	}
	if req.Rule.IsEnabled {
		// Already set
	} else {
		req.Rule.IsEnabled = true
	}

	// TODO: Extract user ID from context/auth (for now use 1 as placeholder)
	req.Rule.UserID = 1

	// Create rule
	id, err := h.repo.CreateRule(req.Rule)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateAlertRuleResponse{
			Success: false,
			Error:   "Failed to create rule: " + err.Error(),
		})
		return
	}

	req.Rule.ID = id

	// Return success response with created rule
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateAlertRuleResponse{
		Success: true,
		Message: "Alert rule created successfully",
		Rule:    req.Rule,
	})
}

// ListAlertRules handles GET /api/v1/alert-rules
func (h *AlertRulesHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// TODO: Extract user ID from context/auth (for now use 1 as placeholder)
	userID := 1

	// List rules
	rules, err := h.repo.ListRules(userID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListAlertRulesResponse{
			Success: false,
			Error:   "Failed to list rules: " + err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListAlertRulesResponse{
		Success: true,
		Rules:   rules,
		Total:   len(rules),
	})
}

// GetAlertRule handles GET /api/v1/alert-rules/:id
func (h *AlertRulesHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract id from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetAlertRuleResponse{
			Success: false,
			Error:   "Missing id in URL path",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetAlertRuleResponse{
			Success: false,
			Error:   "Invalid id format",
		})
		return
	}

	// Get rule
	rule, err := h.repo.GetRuleByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GetAlertRuleResponse{
			Success: false,
			Error:   "Alert rule not found: " + err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetAlertRuleResponse{
		Success: true,
		Rule:    rule,
	})
}

// UpdateAlertRule handles PUT /api/v1/alert-rules/:id
func (h *AlertRulesHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
			Success: false,
			Error:   "Missing id in URL path",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
			Success: false,
			Error:   "Invalid id format",
		})
		return
	}

	// Parse request body
	var req UpdateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
			Success: false,
			Error:   "Malformed request body",
		})
		return
	}

	// Validate rule
	if req.Rule == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
			Success: false,
			Error:   "Rule cannot be null",
		})
		return
	}

	// Set ID for update
	req.Rule.ID = id

	// Validate condition if provided
	if len(req.Rule.Condition) > 0 {
		var condition models.AlertCondition
		if err := json.Unmarshal(req.Rule.Condition, &condition); err == nil {
			if err := h.validator.Validate(condition); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
					Success: false,
					Error:   "Invalid condition: " + err.Error(),
				})
				return
			}
		}
	}

	// Update rule
	err = h.repo.UpdateRule(req.Rule)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
			Success: false,
			Error:   "Failed to update rule: " + err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdateAlertRuleResponse{
		Success: true,
		Message: "Alert rule updated successfully",
		Rule:    req.Rule,
	})
}

// DeleteAlertRule handles DELETE /api/v1/alert-rules/:id
func (h *AlertRulesHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(DeleteAlertRuleResponse{
			Success: false,
			Error:   "Missing id in URL path",
		})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(DeleteAlertRuleResponse{
			Success: false,
			Error:   "Invalid id format",
		})
		return
	}

	// TODO: Extract user ID from context/auth (for now use 1 as placeholder)
	userID := 1

	// Delete rule
	err = h.repo.DeleteRule(id, userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(DeleteAlertRuleResponse{
			Success: false,
			Error:   "Failed to delete rule: " + err.Error(),
		})
		return
	}

	// Return success response (204 No Content is more appropriate, but we'll use JSON for consistency)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteAlertRuleResponse{
		Success: true,
		Message: "Alert rule deleted successfully",
	})
}

// GetAlertHistory handles GET /api/v1/alerts/history
func (h *AlertRulesHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	ruleIDStr := r.URL.Query().Get("rule_id")
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// If rule_id is specified, get history for that rule
	if ruleIDStr != "" {
		ruleID, err := strconv.ParseInt(ruleIDStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AlertHistoryResponse{
				Success: false,
				Error:   "Invalid rule_id format",
			})
			return
		}

		triggers, err := h.repo.GetAlertHistory(ruleID, limit, offset)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AlertHistoryResponse{
				Success: false,
				Error:   "Failed to get alert history: " + err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AlertHistoryResponse{
			Success:  true,
			Triggers: triggers,
			Total:    len(triggers),
		})
		return
	}

	// No rule_id specified - return empty list for now
	// In a full implementation, this would query all triggers across all rules
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AlertHistoryResponse{
		Success:  true,
		Triggers: []*storage.AlertTrigger{},
		Total:    0,
	})
}
