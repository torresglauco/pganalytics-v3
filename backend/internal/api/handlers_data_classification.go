package api

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// DATA CLASSIFICATION ENDPOINTS (DATA-01, DATA-02, DATA-03, DATA-05)
// ============================================================================

// @Summary Get data classification results
// @Description Get PII/PCI detection results for a collector with filtering options
// @Tags Data Classification
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param schema query string false "Schema name filter"
// @Param table query string false "Table name filter"
// @Param pattern_type query string false "Pattern type filter (CPF, CNPJ, EMAIL, PHONE, CREDIT_CARD, CUSTOM)"
// @Param category query string false "Category filter (PII, PCI, SENSITIVE, CUSTOM)"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.ClassificationMetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/classification [get]
func (s *Server) handleGetClassificationResults(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Build filter from query parameters
	filter := models.ClassificationFilter{
		TimeRange: c.DefaultQuery("time_range", "24h"),
	}

	if database := c.Query("database"); database != "" {
		filter.DatabaseName = &database
	}
	if schema := c.Query("schema"); schema != "" {
		filter.SchemaName = &schema
	}
	if table := c.Query("table"); table != "" {
		filter.TableName = &table
	}
	if patternType := c.Query("pattern_type"); patternType != "" {
		filter.PatternType = &patternType
	}
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[filter.TimeRange] {
		filter.TimeRange = "24h"
	}

	// Parse limit and offset
	filter.Limit = 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		filter.Limit = l
	}

	filter.Offset = 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		filter.Offset = o
	}

	ctx := c.Request.Context()

	// Get classification results
	results, err := s.postgres.GetClassificationResults(ctx, collectorID, filter)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.ClassificationMetricsResponse{
		MetricType: "data_classification",
		Count:      len(results),
		TimeRange:  filter.TimeRange,
		Data:       results,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get classification report
// @Description Get aggregated classification report with counts and breakdowns by database/table
// @Tags Data Classification
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.ClassificationReportResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/classification/report [get]
func (s *Server) handleGetClassificationReport(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get classification report
	report, err := s.postgres.GetClassificationReport(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// ============================================================================
// CUSTOM PATTERN ENDPOINTS (DATA-04)
// ============================================================================

// @Summary Get custom patterns
// @Description Get custom detection patterns (global and tenant-specific)
// @Tags Data Classification
// @Produce json
// @Security Bearer
// @Success 200 {object} models.CustomPatternResponse
// @Failure 401 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/classification/patterns [get]
func (s *Server) handleGetCustomPatterns(c *gin.Context) {
	ctx := c.Request.Context()

	// Get tenant_id from context (set by auth middleware)
	var tenantID *uuid.UUID
	if tid, exists := c.Get("tenant_id"); exists {
		if tidUUID, ok := tid.(uuid.UUID); ok {
			tenantID = &tidUUID
		}
	}

	// Get custom patterns
	patterns, err := s.postgres.GetCustomPatterns(ctx, tenantID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.CustomPatternResponse{
		Count:    len(patterns),
		Patterns: patterns,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Create custom pattern
// @Description Create a new custom detection pattern
// @Tags Data Classification
// @Accept json
// @Produce json
// @Security Bearer
// @Param pattern body models.CustomPattern true "Custom pattern definition"
// @Success 201 {object} models.CustomPattern
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/classification/patterns [post]
func (s *Server) handleCreateCustomPattern(c *gin.Context) {
	var pattern models.CustomPattern
	if err := c.ShouldBindJSON(&pattern); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate required fields
	if pattern.PatternName == "" {
		errResp := apperrors.BadRequest("Missing required field", "pattern_name is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	if pattern.PatternRegex == "" {
		errResp := apperrors.BadRequest("Missing required field", "pattern_regex is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	if pattern.Category == "" {
		errResp := apperrors.BadRequest("Missing required field", "category is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate regex pattern
	if _, err := regexp.Compile(pattern.PatternRegex); err != nil {
		errResp := apperrors.BadRequest("Invalid regex pattern", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate category
	validCategories := map[string]bool{"PII": true, "PCI": true, "SENSITIVE": true, "CUSTOM": true}
	if !validCategories[pattern.Category] {
		errResp := apperrors.BadRequest("Invalid category", "must be one of: PII, PCI, SENSITIVE, CUSTOM")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get tenant_id from context
	if tid, exists := c.Get("tenant_id"); exists {
		if tidUUID, ok := tid.(uuid.UUID); ok {
			pattern.TenantID = uuid.NullUUID{UUID: tidUUID, Valid: true}
		}
	}

	ctx := c.Request.Context()

	// Create pattern
	if err := s.postgres.CreateCustomPattern(ctx, &pattern); err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusCreated, pattern)
}

// @Summary Update custom pattern
// @Description Update an existing custom detection pattern
// @Tags Data Classification
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Pattern ID"
// @Param pattern body models.CustomPattern true "Pattern updates"
// @Success 200 {object} models.CustomPattern
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/classification/patterns/{id} [put]
func (s *Server) handleUpdateCustomPattern(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid pattern ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var pattern models.CustomPattern
	if err := c.ShouldBindJSON(&pattern); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate regex if provided
	if pattern.PatternRegex != "" {
		if _, err := regexp.Compile(pattern.PatternRegex); err != nil {
			errResp := apperrors.BadRequest("Invalid regex pattern", err.Error())
			c.JSON(errResp.StatusCode, errResp)
			return
		}
	}

	// Validate category if provided
	if pattern.Category != "" {
		validCategories := map[string]bool{"PII": true, "PCI": true, "SENSITIVE": true, "CUSTOM": true}
		if !validCategories[pattern.Category] {
			errResp := apperrors.BadRequest("Invalid category", "must be one of: PII, PCI, SENSITIVE, CUSTOM")
			c.JSON(errResp.StatusCode, errResp)
			return
		}
	}

	ctx := c.Request.Context()

	// Update pattern
	if err := s.postgres.UpdateCustomPattern(ctx, id, &pattern); err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Set the ID for response
	pattern.ID = id
	c.JSON(http.StatusOK, pattern)
}

// @Summary Delete custom pattern
// @Description Delete a custom detection pattern
// @Tags Data Classification
// @Security Bearer
// @Param id path int true "Pattern ID"
// @Success 204 "No Content"
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/classification/patterns/{id} [delete]
func (s *Server) handleDeleteCustomPattern(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid pattern ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Delete pattern
	if err := s.postgres.DeleteCustomPattern(ctx, id); err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.Status(http.StatusNoContent)
}
