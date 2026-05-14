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
// TABLE INVENTORY ENDPOINTS (INV-01)
// ============================================================================

// @Summary Get Table Inventory
// @Description Get table inventory with sizes and row counts for a collector
// @Tags Inventory
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param schema query string false "Schema name filter"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/inventory/tables [get]
func (s *Server) handleGetTableInventory(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")
	schema := c.Query("schema")

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

	var schemaPtr *string
	if schema != "" {
		schemaPtr = &schema
	}

	// Get table inventory
	tables, err := s.postgres.GetTableInventory(ctx, collectorID, dbPtr, schemaPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "table_inventory",
		Count:      len(tables),
		Timestamp:  time.Now(),
		Data:       tables,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// COLUMN INVENTORY ENDPOINTS (INV-02)
// ============================================================================

// @Summary Get Column Inventory
// @Description Get column inventory with data types for a collector
// @Tags Inventory
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param table query string false "Table name filter"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/inventory/columns [get]
func (s *Server) handleGetColumnInventory(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")
	table := c.Query("table")

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

	var tablePtr *string
	if table != "" {
		tablePtr = &table
	}

	// Get column inventory
	columns, err := s.postgres.GetColumnInventory(ctx, collectorID, dbPtr, tablePtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "column_inventory",
		Count:      len(columns),
		Timestamp:  time.Now(),
		Data:       columns,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// INDEX INVENTORY ENDPOINTS (INV-03)
// ============================================================================

// @Summary Get Index Inventory
// @Description Get index inventory with usage statistics for a collector
// @Tags Inventory
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param table query string false "Table name filter"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/inventory/indexes [get]
func (s *Server) handleGetIndexInventory(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")
	table := c.Query("table")

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

	var tablePtr *string
	if table != "" {
		tablePtr = &table
	}

	// Get index inventory
	indexes, err := s.postgres.GetIndexInventory(ctx, collectorID, dbPtr, tablePtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "index_inventory",
		Count:      len(indexes),
		Timestamp:  time.Now(),
		Data:       indexes,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// EXTENSION INVENTORY ENDPOINTS (INV-04)
// ============================================================================

// @Summary Get Extension Inventory
// @Description Get extension inventory with versions for a collector
// @Tags Inventory
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/inventory/extensions [get]
func (s *Server) handleGetExtensionInventory(c *gin.Context) {
	collectorIDStr := c.Param("id")
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

	// Get extension inventory
	extensions, err := s.postgres.GetExtensionInventory(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "extension_inventory",
		Count:      len(extensions),
		Timestamp:  time.Now(),
		Data:       extensions,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// SCHEMA VERSIONS ENDPOINTS (INV-05)
// ============================================================================

// @Summary Get Schema Versions
// @Description Get schema change history for a collector
// @Tags Inventory
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param database query string false "Database name filter"
// @Param limit query int false "Result limit" default(50)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/inventory/schema-versions [get]
func (s *Server) handleGetSchemaVersions(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")

	limit := 50
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	ctx := c.Request.Context()

	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	// Get schema versions
	versions, err := s.postgres.GetSchemaVersions(ctx, collectorID, dbPtr, limit)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "schema_versions",
		Count:      len(versions),
		Timestamp:  time.Now(),
		Data:       versions,
	}

	c.JSON(http.StatusOK, resp)
}
