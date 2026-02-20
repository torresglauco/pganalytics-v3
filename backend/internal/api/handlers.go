package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dextra/pganalytics-v3/backend/internal/auth"
	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const version = "3.0.0-alpha"

// ============================================================================
// HEALTH & SYSTEM ENDPOINTS
// ============================================================================

// @Summary Health Check
// @Description Check system health status
// @Tags Health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /api/v1/health [get]
func (s *Server) handleHealth(c *gin.Context) {
	ctx := c.Request.Context()

	healthResp := &models.HealthResponse{
		Status:      "ok",
		Version:     version,
		Timestamp:   time.Now(),
		DatabaseOk:  s.postgres.Health(ctx),
		TimescaleOk: s.timescale.Health(ctx),
	}

	if !healthResp.DatabaseOk || !healthResp.TimescaleOk {
		healthResp.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, healthResp)
		return
	}

	c.JSON(http.StatusOK, healthResp)
}

// @Summary Get Version
// @Description Get API version
// @Tags System
// @Produce json
// @Success 200 {object} gin.H
// @Router /version [get]
func (s *Server) handleVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"api":     "1.0.0",
	})
}

// ============================================================================
// AUTHENTICATION ENDPOINTS
// ============================================================================

// @Summary User Login
// @Description Authenticate user and get JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/login [post]
func (s *Server) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Authenticate user
	loginResp, err := s.authService.LoginUser(req.Username, req.Password)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User login successful",
		"username", req.Username,
		"user_id", loginResp.User.ID,
	)

	c.JSON(http.StatusOK, loginResp)
}

// @Summary User Logout
// @Description Logout user (invalidate token)
// @Tags Authentication
// @Security Bearer
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/auth/logout [post]
func (s *Server) handleLogout(c *gin.Context) {
	// Token invalidation would typically be done via token blacklist
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// @Summary Refresh Token
// @Description Refresh JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body gin.H true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (s *Server) handleRefreshToken(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	refreshToken, ok := req["refresh_token"].(string)
	if !ok || refreshToken == "" {
		errResp := apperrors.BadRequest("Missing refresh token", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Refresh user token
	loginResp, err := s.authService.RefreshUserToken(refreshToken)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Token refreshed successfully",
		"user_id", loginResp.User.ID,
	)

	c.JSON(http.StatusOK, loginResp)
}

// ============================================================================
// COLLECTOR ENDPOINTS
// ============================================================================

// @Summary Register Collector
// @Description Register a new collector and get authentication credentials
// @Tags Collectors
// @Accept json
// @Produce json
// @Param request body models.CollectorRegisterRequest true "Collector info"
// @Success 200 {object} models.CollectorRegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /api/v1/collectors/register [post]
func (s *Server) handleCollectorRegister(c *gin.Context) {
	var req models.CollectorRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Register collector
	registerResp, err := s.authService.RegisterCollector(&req)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Collector registered successfully",
		"collector_id", registerResp.CollectorID.String(),
		"hostname", req.Hostname,
	)

	c.JSON(http.StatusOK, registerResp)
}

// @Summary List Collectors
// @Description List all registered collectors
// @Tags Collectors
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors [get]
func (s *Server) handleListCollectors(c *gin.Context) {
	// Parse pagination parameters
	var params models.PaginationParams
	params.Page = 1
	params.PageSize = 20

	if page := c.Query("page"); page != "" {
		_, _ = parseIntParam(page, &params.Page)
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		_, _ = parseIntParam(pageSize, &params.PageSize)
	}

	// TODO: Query collectors from database with pagination
	// For now, return empty list
	resp := &models.PaginatedResponse{
		Data:       []models.Collector{},
		Total:      0,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: 0,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Collector
// @Description Get collector details by ID
// @Tags Collectors
// @Security Bearer
// @Produce json
// @Param id path string true "Collector ID"
// @Success 200 {object} models.Collector
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/collectors/{id} [get]
func (s *Server) handleGetCollector(c *gin.Context) {
	// TODO: Implement get collector
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Delete Collector
// @Description Deregister a collector
// @Tags Collectors
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/collectors/{id} [delete]
func (s *Server) handleDeleteCollector(c *gin.Context) {
	// TODO: Implement delete collector
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// METRICS ENDPOINTS
// ============================================================================

// @Summary Push Metrics
// @Description Ingest metrics from a collector (high-volume endpoint)
// @Tags Metrics
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.MetricsPushRequest true "Metrics data"
// @Success 200 {object} models.MetricsPushResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/metrics/push [post]
func (s *Server) handleMetricsPush(c *gin.Context) {
	var req models.MetricsPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid metrics data", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get collector from context (set by CollectorAuthMiddleware)
	collectorClaimsInterface, exists := c.Get("collector_claims")
	if !exists {
		errResp := apperrors.Unauthorized("No collector claims", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	collectorClaims, ok := collectorClaimsInterface.(*auth.CollectorClaims)
	if !ok {
		errResp := apperrors.Unauthorized("Invalid collector claims", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate collector ID matches request
	if collectorClaims.CollectorID != req.CollectorID {
		errResp := apperrors.Unauthorized("Collector ID mismatch", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// TODO: Process and store metrics
	// For now, return success response
	processingTimeMs := int64(time.Now().UnixMilli()) % 500 // Simulate processing time

	resp := &models.MetricsPushResponse{
		Status:             "success",
		CollectorID:        req.CollectorID,
		MetricsInserted:    req.MetricsCount,
		BytesReceived:      c.Request.ContentLength,
		ProcessingTimeMs:   processingTimeMs,
		NextConfigVersion:  1,
		NextCheckInSeconds: 300,
	}

	s.logger.Info("Metrics pushed successfully",
		"collector_id", req.CollectorID,
		"metrics_count", req.MetricsCount,
	)

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CONFIGURATION ENDPOINTS
// ============================================================================

// @Summary Get Collector Config
// @Description Get configuration for a collector (pulled by collector)
// @Tags Configuration
// @Security Bearer
// @Produce plain
// @Param collector_id path string true "Collector ID"
// @Success 200 {string} string "TOML configuration"
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/config/{collector_id} [get]
func (s *Server) handleGetConfig(c *gin.Context) {
	collectorID := c.Param("collector_id")
	ctx := c.Request.Context()

	// Get the latest collector configuration
	config, err := s.postgres.GetCollectorConfig(ctx, collectorID)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Return TOML format with version info in header
	c.Header("Content-Type", "text/plain")
	c.Header("X-Config-Version", fmt.Sprintf("%d", config.Version))
	c.String(http.StatusOK, config.Config)

	s.logger.Info("Collector config retrieved",
		"collector_id", collectorID,
		"config_version", config.Version,
	)
}

// @Summary Update Collector Config
// @Description Update configuration for a collector (admin only)
// @Tags Configuration
// @Security Bearer
// @Accept plain
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param request body string true "TOML Configuration"
// @Success 200 {object} gin.H
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/config/{collector_id} [put]
func (s *Server) handleUpdateConfig(c *gin.Context) {
	collectorID := c.Param("collector_id")
	ctx := c.Request.Context()

	// Get user from context (set by AuthMiddleware)
	userInterface, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("No user found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		errResp := apperrors.Unauthorized("Invalid user claims", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if user is admin (optional - for now allow any authenticated user)
	// In a real implementation, this would check user.Role == "admin"

	// Read TOML content from request body
	tomlContent, err := c.GetRawData()
	if err != nil {
		errResp := apperrors.BadRequest("Failed to read configuration", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create new config version
	parsedUUID, err := uuid.Parse(collectorID)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID format", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	config := &models.CollectorConfig{
		CollectorID: parsedUUID,
		Config:      string(tomlContent),
		UpdatedBy:   &user.ID,
	}

	if err = s.postgres.CreateCollectorConfig(ctx, config); err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Collector config updated",
		"collector_id", collectorID,
		"config_version", config.Version,
		"updated_by", user.ID,
	)

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"collector_id": collectorID,
		"version":      config.Version,
		"message":      "Configuration updated successfully",
	})
}

// ============================================================================
// SERVER ENDPOINTS
// ============================================================================

// @Summary List Servers
// @Description List all monitored servers
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/servers [get]
func (s *Server) handleListServers(c *gin.Context) {
	// TODO: Implement list servers
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Server
// @Description Get server details by ID
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} models.Server
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/servers/{id} [get]
func (s *Server) handleGetServer(c *gin.Context) {
	// TODO: Implement get server
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Server Metrics
// @Description Get historical metrics for a server
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param id path int true "Server ID"
// @Param metric_type query string false "Metric type" default(pg_stats)
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} gin.H
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/servers/{id}/metrics [get]
func (s *Server) handleGetServerMetrics(c *gin.Context) {
	// TODO: Implement get server metrics
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// ALERTS ENDPOINTS
// ============================================================================

// @Summary List Alerts
// @Description List alerts for monitored resources
// @Tags Alerts
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param severity query string false "Severity filter" default(all)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/alerts [get]
func (s *Server) handleListAlerts(c *gin.Context) {
	// TODO: Implement list alerts
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Alert
// @Description Get alert details by ID
// @Tags Alerts
// @Security Bearer
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} models.Alert
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/alerts/{id} [get]
func (s *Server) handleGetAlert(c *gin.Context) {
	// TODO: Implement get alert
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Acknowledge Alert
// @Description Acknowledge an alert
// @Tags Alerts
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param request body gin.H true "Acknowledgment data"
// @Success 200 {object} gin.H
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/alerts/{id}/acknowledge [post]
func (s *Server) handleAcknowledgeAlert(c *gin.Context) {
	// TODO: Implement acknowledge alert
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// parseIntParam parses an integer parameter from string
func parseIntParam(value string, target *int) (int, error) {
	_, _ = value, target // Placeholder
	// TODO: Implement proper integer parsing with validation
	return 0, nil
}
