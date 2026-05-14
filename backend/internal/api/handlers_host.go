package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// HOST STATUS ENDPOINTS (HOST-01)
// ============================================================================

// @Summary Get Host Status
// @Description Get host up/down status based on collector last_seen timestamp
// @Tags Host
// @Produce json
// @Security Bearer
// @Param id path string false "Collector ID (optional - if missing, returns all hosts)"
// @Param threshold query int false "Down threshold in seconds" default(300)
// @Success 200 {object} models.HostStatusResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/hosts [get]
// @Router /api/v1/hosts/{id}/status [get]
func (s *Server) handleGetHostStatus(c *gin.Context) {
	// Get threshold from query param (default 300 seconds = 5 minutes)
	threshold := 300
	if t, err := strconv.Atoi(c.DefaultQuery("threshold", "300")); err == nil && t > 0 {
		threshold = t
	}

	ctx := c.Request.Context()

	// Check if collector ID is provided in path
	collectorIDStr := c.Param("id")
	if collectorIDStr == "" {
		// No ID provided - return all hosts status
		statuses, err := s.postgres.GetAllHostStatuses(ctx, threshold)
		if err != nil {
			c.JSON(err.(*apperrors.AppError).StatusCode, err)
			return
		}

		resp := &models.HostStatusResponse{
			Count:  len(statuses),
			Status: statuses,
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	// Single host status
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	status, err := s.postgres.GetHostStatus(ctx, collectorID, threshold)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.HostStatusResponse{
		Count:  1,
		Status: []*models.HostStatus{status},
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// HOST METRICS ENDPOINTS (HOST-02)
// ============================================================================

// @Summary Get Host Metrics
// @Description Get OS metrics (CPU, memory, disk I/O, load average) for a host
// @Tags Host
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Param limit query int false "Result limit" default(100)
// @Success 200 {object} models.HostMetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/hosts/{id}/metrics [get]
func (s *Server) handleGetHostMetrics(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get time_range query param (default 24h)
	timeRange := c.DefaultQuery("time_range", "24h")

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "24h" // Default to 24h if invalid
	}

	// Get limit (default 100, max 1000)
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	ctx := c.Request.Context()

	metrics, err := s.postgres.GetHostMetrics(ctx, collectorID, timeRange, limit)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.HostMetricsResponse{
		MetricType: "host_metrics",
		Count:      len(metrics),
		TimeRange:  timeRange,
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// HOST INVENTORY ENDPOINTS (HOST-03)
// ============================================================================

// @Summary Get Host Inventory
// @Description Get host inventory (OS version, hardware specs, PostgreSQL configuration)
// @Tags Host
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.HostInventoryResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/hosts/{id}/inventory [get]
func (s *Server) handleGetHostInventory(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	inventory, err := s.postgres.GetHostInventory(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.HostInventoryResponse{
		MetricType: "host_inventory",
		Data:       inventory,
	}

	c.JSON(http.StatusOK, resp)
}
