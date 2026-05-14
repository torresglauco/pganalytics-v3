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
// VERSION HEALTH CHECK ENDPOINTS (VER-03)
// ============================================================================

// @Summary Get version-specific health checks for collector
// @Description Get all health checks applicable to the PostgreSQL version of a collector
// @Tags HealthChecks
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.VersionHealthCheckResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/health-checks [get]
func (s *Server) handleGetVersionHealthChecks(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get PostgreSQL version for the collector
	version, err := s.postgres.GetPostgreSQLVersion(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Get health checks applicable to this version
	checks, err := s.postgres.GetHealthChecksForVersion(ctx, version.Major)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Build response
	resp := &models.VersionHealthCheckResponse{
		CollectorID:             collectorID,
		PostgreSQLVersion:       version.Major,
		PostgreSQLVersionString: version.FullVersion,
		Results:                 []*models.HealthCheckResult{}, // Empty results - just definitions
		Summary: models.HealthCheckSummary{
			TotalChecks: len(checks),
		},
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Execute version-specific health checks
// @Description Run all applicable health checks for a collector and return results
// @Tags HealthChecks
// @Produce json
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 200 {object} models.VersionHealthCheckResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{id}/health-checks/run [post]
func (s *Server) handleRunVersionHealthChecks(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get PostgreSQL version for the collector
	version, err := s.postgres.GetPostgreSQLVersion(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Get health checks applicable to this version
	checks, err := s.postgres.GetHealthChecksForVersion(ctx, version.Major)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Execute each health check
	var results []*models.HealthCheckResult
	summary := models.HealthCheckSummary{
		TotalChecks: len(checks),
	}

	for _, check := range checks {
		result, err := s.postgres.RunHealthCheck(ctx, collectorID, check)
		if err != nil {
			// Log error but continue with other checks
			result = &models.HealthCheckResult{
				CheckID:   check.ID,
				CheckName: check.CheckName,
				Severity:  check.Severity,
				Passed:    false,
				Message:   err.Error(),
				CheckedAt: check.CreatedAt,
			}
		}

		// Update summary counts
		if result.Passed {
			summary.PassedChecks++
		} else {
			switch result.Severity {
			case "critical":
				summary.FailedCritical++
			case "warning":
				summary.FailedWarning++
			case "info":
				summary.FailedInfo++
			}
		}

		results = append(results, result)
	}

	// Build response
	resp := &models.VersionHealthCheckResponse{
		CollectorID:             collectorID,
		PostgreSQLVersion:       version.Major,
		PostgreSQLVersionString: version.FullVersion,
		Results:                 results,
		Summary:                 summary,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get all defined health checks
// @Description Get all health check definitions, optionally filtered by PostgreSQL version
// @Tags HealthChecks
// @Produce json
// @Security Bearer
// @Param version query int false "Filter by PostgreSQL major version"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} apperrors.AppError
// @Router /api/v1/health-checks [get]
func (s *Server) handleGetAllHealthChecks(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if version filter is provided
	versionStr := c.Query("version")
	if versionStr != "" {
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			errResp := apperrors.BadRequest("Invalid version parameter", err.Error())
			c.JSON(errResp.StatusCode, errResp)
			return
		}

		// Get checks for specific version
		checks, err := s.postgres.GetHealthChecksForVersion(ctx, version)
		if err != nil {
			c.JSON(err.(*apperrors.AppError).StatusCode, err)
			return
		}

		resp := map[string]interface{}{
			"version": version,
			"count":   len(checks),
			"checks":  checks,
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	// Get all health checks
	checks, err := s.postgres.GetAllHealthChecks(ctx)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := map[string]interface{}{
		"count":  len(checks),
		"checks": checks,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get health check definition
// @Description Get a single health check definition by ID
// @Tags HealthChecks
// @Produce json
// @Security Bearer
// @Param id path int true "Health Check ID"
// @Success 200 {object} models.VersionHealthCheck
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/health-checks/{id} [get]
func (s *Server) handleGetHealthCheckByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid health check ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	check, err := s.postgres.GetHealthCheckByID(ctx, id)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, check)
}
