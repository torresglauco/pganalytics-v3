package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/audit"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================================================
// AUDIT LOG ENDPOINTS
// ============================================================================

// AuditLogQuery represents query parameters for audit log filtering
type AuditLogQuery struct {
	UserID       *int   `form:"user_id"`
	Action       string `form:"action"`
	ResourceType string `form:"resource_type"`
	ResourceID   string `form:"resource_id"`
	StartDate    string `form:"start_date"` // ISO8601 format
	EndDate      string `form:"end_date"`   // ISO8601 format
	Limit        int    `form:"limit" binding:"max=1000"`
	Offset       int    `form:"offset"`
	Format       string `form:"format"` // json, csv
}

// @Summary Get Audit Logs
// @Description Retrieve audit logs with filtering (admin only)
// @Tags Audit
// @Produce json
// @Security Bearer
// @Param user_id query int false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource_type query string false "Filter by resource type"
// @Param resource_id query string false "Filter by resource ID"
// @Param start_date query string false "Start date (ISO8601)"
// @Param end_date query string false "End date (ISO8601)"
// @Param limit query int false "Results limit (default 100, max 1000)"
// @Param offset query int false "Results offset (default 0)"
// @Success 200 {array} audit.AuditLog
// @Failure 400 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/audit-logs [get]
func (s *Server) handleAuditLogs(c *gin.Context) {
	// Check admin permission
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Admin access required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var query AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errResp := apperrors.BadRequest("Invalid query parameters", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Set defaults
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	ctx := c.Request.Context()

	// Build filter
	filter := &audit.AuditFilter{
		UserID:       query.UserID,
		Action:       &query.Action,
		ResourceType: &query.ResourceType,
		ResourceID:   &query.ResourceID,
		Limit:        query.Limit,
		Offset:       query.Offset,
	}

	// Parse dates if provided
	if query.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, query.StartDate); err == nil {
			filter.DateFrom = &t
		}
	}
	if query.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, query.EndDate); err == nil {
			filter.DateTo = &t
		}
	}

	// Query audit logs
	logs, err := s.auditLogger.GetHistory(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to retrieve audit logs", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve audit logs", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Return format
	if query.Format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=\"audit-logs.csv\"")

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write header
		header := []string{"ID", "User ID", "Action", "Resource Type", "Resource ID", "IP Address", "User Agent", "Timestamp"}
		writer.Write(header)

		// Write rows
		for _, log := range logs {
			userIDStr := ""
			if log.UserID != nil {
				userIDStr = fmt.Sprintf("%d", *log.UserID)
			}
			resourceIDStr := ""
			if log.ResourceID != nil {
				resourceIDStr = *log.ResourceID
			}
			ipStr := ""
			if log.IPAddress != nil {
				ipStr = log.IPAddress.String()
			}
			userAgent := ""
			if log.UserAgent != nil {
				userAgent = *log.UserAgent
			}

			row := []string{
				fmt.Sprintf("%d", log.ID),
				userIDStr,
				string(log.Action),
				log.ResourceType,
				resourceIDStr,
				ipStr,
				userAgent,
				log.Timestamp.Format(time.RFC3339),
			}
			writer.Write(row)
		}
		return
	}

	c.JSON(http.StatusOK, logs)
}

// @Summary Get Audit Log Detail
// @Description Retrieve a specific audit log entry (admin only)
// @Tags Audit
// @Produce json
// @Security Bearer
// @Param id path int true "Audit log ID"
// @Success 200 {object} audit.AuditLog
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/audit-logs/{id} [get]
func (s *Server) handleAuditLogDetail(c *gin.Context) {
	// Check admin permission
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Admin access required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid audit log ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get audit log
	log, err := s.auditLogger.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve audit log", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve audit log", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if log == nil {
		errResp := apperrors.NotFound("Audit log not found", fmt.Sprintf("ID: %d", id))
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, log)
}

// AuditStatsResponse represents audit log statistics
type AuditStatsResponse struct {
	TotalLogs     int64                  `json:"total_logs"`
	LogsByAction  map[string]int64       `json:"logs_by_action"`
	LogsByUser    map[int]int64          `json:"logs_by_user"`
	LogsByResource map[string]int64      `json:"logs_by_resource"`
	LatestLogTime time.Time              `json:"latest_log_time"`
	OldestLogTime time.Time              `json:"oldest_log_time"`
}

// @Summary Get Audit Log Statistics
// @Description Get audit log statistics and summary (admin only)
// @Tags Audit
// @Produce json
// @Security Bearer
// @Success 200 {object} AuditStatsResponse
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/audit-logs/stats [get]
func (s *Server) handleAuditStats(c *gin.Context) {
	// Check admin permission
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Admin access required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	stats, err := s.auditLogger.GetStats(ctx)
	if err != nil {
		s.logger.Error("Failed to retrieve audit stats", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve audit statistics", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportAuditLogsRequest represents request to export audit logs
type ExportAuditLogsRequest struct {
	StartDate    string `json:"start_date"` // ISO8601 format
	EndDate      string `json:"end_date"`   // ISO8601 format
	Format       string `json:"format"`     // json, csv
	IncludeData  bool   `json:"include_data"`
}

// @Summary Export Audit Logs
// @Description Export audit logs in specified format (admin only)
// @Tags Audit
// @Accept json
// @Produce application/json,text/csv
// @Security Bearer
// @Param request body ExportAuditLogsRequest true "Export parameters"
// @Success 200 {file} string "Exported audit logs"
// @Failure 400 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/audit-logs/export [post]
func (s *Server) handleExportAuditLogs(c *gin.Context) {
	// Check admin permission
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Admin access required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req ExportAuditLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Build filter
	filter := &audit.AuditFilter{
		Limit: 10000,
	}

	// Parse dates
	if req.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			filter.StartTime = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			filter.EndTime = &t
		}
	}

	// Query audit logs
	logs, err := s.auditLogger.GetHistory(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to retrieve audit logs for export", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to export audit logs", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Export in specified format
	if req.Format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"audit-logs-%d.csv\"", time.Now().Unix()))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write header
		header := []string{"ID", "User ID", "Action", "Resource Type", "Resource ID", "IP Address", "User Agent", "Timestamp"}
		if req.IncludeData {
			header = append(header, "Changes Before", "Changes After", "Additional Data")
		}
		writer.Write(header)

		// Write rows
		for _, log := range logs {
			userIDStr := ""
			if log.UserID != nil {
				userIDStr = fmt.Sprintf("%d", *log.UserID)
			}
			resourceIDStr := ""
			if log.ResourceID != nil {
				resourceIDStr = *log.ResourceID
			}
			ipStr := ""
			if log.IPAddress != nil {
				ipStr = log.IPAddress.String()
			}
			userAgent := ""
			if log.UserAgent != nil {
				userAgent = *log.UserAgent
			}

			row := []string{
				fmt.Sprintf("%d", log.ID),
				userIDStr,
				string(log.Action),
				log.ResourceType,
				resourceIDStr,
				ipStr,
				userAgent,
				log.Timestamp.Format(time.RFC3339),
			}

			if req.IncludeData {
				changesBefore := ""
				if log.ChangesBefore != nil {
					changesBefore = string(log.ChangesBefore)
				}
				changesAfter := ""
				if log.ChangesAfter != nil {
					changesAfter = string(log.ChangesAfter)
				}
				additionalData := ""
				if log.AdditionalData != nil {
					additionalData = string(log.AdditionalData)
				}

				row = append(row, changesBefore, changesAfter, additionalData)
			}

			writer.Write(row)
		}
		return
	}

	// JSON format (default)
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"audit-logs-%d.json\"", time.Now().Unix()))

	data := gin.H{
		"exported_at": time.Now(),
		"count":       len(logs),
		"logs":        logs,
	}

	encoder := json.NewEncoder(c.Writer)
	encoder.Encode(data)
}

// RegisterAuditHandlers registers all audit handlers
func (s *Server) RegisterAuditHandlers(engine *gin.Engine) {
	auditGroup := engine.Group("/api/v1/audit-logs")
	auditGroup.Use(s.AuthMiddleware(), s.RoleMiddleware("admin"))
	{
		auditGroup.GET("", s.handleAuditLogs)
		auditGroup.GET("/:id", s.handleAuditLogDetail)
		auditGroup.GET("/stats", s.handleAuditStats)
		auditGroup.POST("/export", s.handleExportAuditLogs)
	}
}
