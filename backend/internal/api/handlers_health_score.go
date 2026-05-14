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
// HEALTH SCORE ENDPOINTS (HOST-04)
// ============================================================================

// @Summary Get host health score
// @Description Get the latest health score for a host with component breakdown
// @Tags Hosts
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.HealthScoreResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/hosts/{id}/health [get]
func (s *Server) handleGetHealthScore(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get latest health score
	score, err := s.postgres.GetLatestHealthScore(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Get latest metrics for context
	metrics, err := s.postgres.GetHostMetrics(ctx, collectorID, "1h", 1)
	if err != nil {
		// Don't fail if metrics aren't available, just return score without context
		c.JSON(http.StatusOK, &models.HealthScoreResponse{HealthScore: score})
		return
	}

	var latestMetrics *models.HostMetrics
	if len(metrics) > 0 {
		latestMetrics = metrics[0]
	}

	c.JSON(http.StatusOK, &models.HealthScoreResponse{
		HealthScore:   score,
		LatestMetrics: latestMetrics,
	})
}

// @Summary Get host health score history
// @Description Get historical health scores for a host with pagination
// @Tags Hosts
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.HealthScoreHistoryResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/hosts/{id}/health/history [get]
func (s *Server) handleGetHealthScoreHistory(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Parse query parameters
	timeRange := c.DefaultQuery("time_range", "24h")

	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "24h"
	}

	ctx := c.Request.Context()

	// Get health score history
	scores, err := s.postgres.GetHealthScoreHistory(ctx, collectorID, timeRange, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, &models.HealthScoreHistoryResponse{
		Scores: scores,
		Pagination: models.PaginationParams{
			Page:     (offset / limit) + 1,
			PageSize: limit,
		},
	})
}

// @Summary Calculate and store host health score
// @Description Trigger immediate health score calculation from latest metrics
// @Tags Hosts
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.HealthScore
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/hosts/{id}/health/calculate [post]
func (s *Server) handleCalculateHealthScore(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Calculate and store health score
	score, err := s.postgres.CalculateAndStoreHealthScore(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, score)
}
