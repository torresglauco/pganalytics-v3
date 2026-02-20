package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dextra/pganalytics-v3/backend/internal/auth"
	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

	// Process metrics based on type
	startTime := time.Now()
	metricsInserted := 0

	if req.Metrics != nil && len(req.Metrics) > 0 {
		for _, metric := range req.Metrics {
			// Type assertion to access metric fields
			if metricMap, ok := metric.(map[string]interface{}); ok {
				metricType := ""
				if typeVal, exists := metricMap["type"]; exists {
					metricType = typeVal.(string)
				}

				// Handle query stats metrics
				if metricType == "pg_query_stats" {
					// Convert metric to QueryStatsRequest
					var queryStatsReq models.QueryStatsRequest
					metricsJSON, _ := json.Marshal(metric)
					json.Unmarshal(metricsJSON, &queryStatsReq)

					// Extract and store individual query statistics
					if queryStatsReq.Databases != nil {
						for _, db := range queryStatsReq.Databases {
							for _, queryInfo := range db.Queries {
								stat := &models.QueryStats{
									Time:              queryStatsReq.Timestamp,
									CollectorID:       uuid.MustParse(req.CollectorID),
									DatabaseName:      db.Database,
									UserName:          "system", // Set from query info if available
									QueryHash:         queryInfo.Hash,
									QueryText:         queryInfo.Text,
									Calls:             queryInfo.Calls,
									TotalTime:         queryInfo.TotalTime,
									MeanTime:          queryInfo.MeanTime,
									MinTime:           queryInfo.MinTime,
									MaxTime:           queryInfo.MaxTime,
									StddevTime:        queryInfo.StddevTime,
									Rows:              queryInfo.Rows,
									SharedBlksHit:     queryInfo.SharedBlksHit,
									SharedBlksRead:    queryInfo.SharedBlksRead,
									SharedBlksDirtied: queryInfo.SharedBlksDirtied,
									SharedBlksWritten: queryInfo.SharedBlksWritten,
									LocalBlksHit:      queryInfo.LocalBlksHit,
									LocalBlksRead:     queryInfo.LocalBlksRead,
									LocalBlksDirtied:  queryInfo.LocalBlksDirtied,
									LocalBlksWritten:  queryInfo.LocalBlksWritten,
									TempBlksRead:      queryInfo.TempBlksRead,
									TempBlksWritten:   queryInfo.TempBlksWritten,
									BlkReadTime:       queryInfo.BlkReadTime,
									BlkWriteTime:      queryInfo.BlkWriteTime,
									WalRecords:        queryInfo.WalRecords,
									WalFpi:            queryInfo.WalFpi,
									WalBytes:          queryInfo.WalBytes,
									QueryPlanTime:     queryInfo.QueryPlanTime,
									QueryExecTime:     queryInfo.QueryExecTime,
								}

								// Insert individual query stat
								if err := s.db.InsertQueryStats(c, req.CollectorID, []*models.QueryStats{stat}); err != nil {
									s.logger.Error("Failed to insert query stat",
										zap.Error(err),
										zap.String("query_hash", fmt.Sprintf("%d", queryInfo.Hash)),
										zap.String("database", db.Database),
									)
								} else {
									metricsInserted++

									// Phase 4.4.2: Trigger EXPLAIN plan capture for slow queries (>1000ms mean time)
									// Strategy: Capture EXPLAIN on first occurrence to minimize overhead
									// Note: Actual EXPLAIN execution happens on the collector side,
									// which sends back the plan via the /api/v1/internal/explain-plans endpoint
									if queryInfo.MeanTime > 1000.0 {
										// Check if EXPLAIN plan already exists for this query
										existingPlan, err := s.db.GetLatestExplainPlan(c, queryInfo.Hash)
										if err == nil && existingPlan != nil {
											// Plan already exists, skip to avoid duplicate EXPLAINs
											continue
										}

										// TODO: Queue EXPLAIN execution request on collector
										// The collector will execute EXPLAIN (ANALYZE, FORMAT JSON, BUFFERS)
										// and send back results via POST /api/v1/internal/explain-plans
										// For now, we log the slow query for manual investigation
										s.logger.Info("Slow query detected (candidate for EXPLAIN)",
											zap.String("query_hash", fmt.Sprintf("%d", queryInfo.Hash)),
											zap.String("database", db.Database),
											zap.Float64("mean_time_ms", queryInfo.MeanTime),
										)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	processingTimeMs := time.Since(startTime).Milliseconds()

	resp := &models.MetricsPushResponse{
		Status:             "success",
		CollectorID:        req.CollectorID,
		MetricsInserted:    metricsInserted,
		BytesReceived:      c.Request.ContentLength,
		ProcessingTimeMs:   processingTimeMs,
		NextConfigVersion:  1,
		NextCheckInSeconds: 300,
	}

	s.logger.Info("Metrics pushed successfully",
		"collector_id", req.CollectorID,
		"metrics_inserted", metricsInserted,
		"processing_time_ms", processingTimeMs,
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
// QUERY STATISTICS HANDLERS
// ============================================================================

// @Summary Get Top Slow Queries
// @Description Get the slowest queries from a collector
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(20)
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.TopQueriesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors/{collector_id}/queries/slow [get]
func (s *Server) handleGetSlowQueries(c *gin.Context) {
	collectorID := c.Param("collector_id")
	limitStr := c.DefaultQuery("limit", "20")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse parameters
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	queries, err := s.db.GetTopSlowQueries(c, collectorID, limit, since)
	if err != nil {
		s.logger.Error("Failed to query slow queries", zap.Error(err), zap.String("collector_id", collectorID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query slow queries",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.TopQueriesResponse{
		ServerID:  uuid.MustParse(collectorID),
		QueryType: "slow",
		Hours:     hours,
		Count:     len(queries),
		Queries:   queries,
	})
}

// @Summary Get Top Frequent Queries
// @Description Get the most frequently executed queries from a collector
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(20)
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.TopQueriesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors/{collector_id}/queries/frequent [get]
func (s *Server) handleGetFrequentQueries(c *gin.Context) {
	collectorID := c.Param("collector_id")
	limitStr := c.DefaultQuery("limit", "20")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse parameters
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	queries, err := s.db.GetTopFrequentQueries(c, collectorID, limit, since)
	if err != nil {
		s.logger.Error("Failed to query frequent queries", zap.Error(err), zap.String("collector_id", collectorID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query frequent queries",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.TopQueriesResponse{
		ServerID:  uuid.MustParse(collectorID),
		QueryType: "frequent",
		Hours:     hours,
		Count:     len(queries),
		Queries:   queries,
	})
}

// @Summary Get Query Timeline
// @Description Get time-series data for a specific query
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param query_hash path int true "Query hash from pg_stat_statements"
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.QueryTimelineResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/timeline [get]
func (s *Server) handleGetQueryTimeline(c *gin.Context) {
	hashStr := c.Param("query_hash")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse query hash
	queryHash, err := strconv.ParseInt(hashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid query hash",
			Code:    http.StatusBadRequest,
		})
		return
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	stats, err := s.db.GetQueryTimeline(c, queryHash, since)
	if err != nil {
		s.logger.Error("Failed to query timeline", zap.Error(err), zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query timeline",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.QueryTimelineResponse{
		QueryHash: queryHash,
		Data:      stats,
	})
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
