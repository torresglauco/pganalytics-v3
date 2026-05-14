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
// STREAMING REPLICATION ENDPOINTS
// ============================================================================

// @Summary Get Replication Metrics
// @Description Get streaming replication status with lag metrics
// @Tags Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/replication [get]
func (s *Server) handleGetReplicationMetrics(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	metrics, err := s.postgres.GetReplicationMetrics(ctx, collectorID, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_replication",
		Count:      len(metrics.ReplicationStatus),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Replication Slots
// @Description Get replication slots with WAL retention information
// @Tags Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/replication-slots [get]
func (s *Server) handleGetReplicationSlots(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	slots, err := s.postgres.GetReplicationSlots(ctx, collectorID, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_replication_slots",
		Count:      len(slots),
		Timestamp:  time.Now(),
		Data:       slots,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// LOGICAL REPLICATION ENDPOINTS
// ============================================================================

// @Summary Get Logical Subscriptions
// @Description Get logical replication subscriptions with state and LSN information
// @Tags Logical Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/logical-subscriptions [get]
func (s *Server) handleGetLogicalSubscriptions(c *gin.Context) {
	collectorIDStr := c.Param("id")
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

	subs, err := s.postgres.GetLogicalSubscriptions(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_logical_subscriptions",
		Count:      len(subs),
		Timestamp:  time.Now(),
		Data:       subs,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Publications
// @Description Get logical replication publications with table ownership
// @Tags Logical Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/publications [get]
func (s *Server) handleGetPublications(c *gin.Context) {
	collectorIDStr := c.Param("id")
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

	pubs, err := s.postgres.GetPublications(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_publications",
		Count:      len(pubs),
		Timestamp:  time.Now(),
		Data:       pubs,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Replication Topology
// @Description Get cascading replication topology showing primary -> standby -> standby chains
// @Tags Logical Replication
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Success 200 {object} models.ReplicationTopology
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/topology [get]
func (s *Server) handleGetReplicationTopology(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()
	topology, err := s.postgres.GetReplicationTopology(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"topology": topology})
}
