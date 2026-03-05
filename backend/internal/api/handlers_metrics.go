package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// SCHEMA METRICS ENDPOINTS
// ============================================================================

// @Summary Get Schema Metrics
// @Description Get schema information for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/schema [get]
func (s *Server) handleGetSchemaMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	// Get schema metrics
	metrics, err := s.postgres.GetSchemaMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_schema",
		Count:      len(metrics.Tables),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// LOCK METRICS ENDPOINTS
// ============================================================================

// @Summary Get Lock Metrics
// @Description Get lock information for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/locks [get]
func (s *Server) handleGetLockMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetLockMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_locks",
		Count:      len(metrics.ActiveLocks),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// BLOAT METRICS ENDPOINTS
// ============================================================================

// @Summary Get Bloat Metrics
// @Description Get table and index bloat metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/bloat [get]
func (s *Server) handleGetBloatMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetBloatMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_bloat",
		Count:      len(metrics.TableBloat),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CACHE METRICS ENDPOINTS
// ============================================================================

// @Summary Get Cache Hit Metrics
// @Description Get cache hit ratio metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/cache-hits [get]
func (s *Server) handleGetCacheMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetCacheMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_cache",
		Count:      len(metrics.TableCacheHit),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CONNECTION METRICS ENDPOINTS
// ============================================================================

// @Summary Get Connection Metrics
// @Description Get connection tracking metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/connections [get]
func (s *Server) handleGetConnectionMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetConnectionMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_connections",
		Count:      len(metrics.ConnectionSummary),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// EXTENSION METRICS ENDPOINTS
// ============================================================================

// @Summary Get Extension Metrics
// @Description Get extension inventory metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/extensions [get]
func (s *Server) handleGetExtensionMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetExtensionMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_extensions",
		Count:      len(metrics.Extensions),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}
