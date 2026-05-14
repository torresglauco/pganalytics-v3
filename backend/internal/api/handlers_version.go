package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// VERSION INFORMATION ENDPOINTS
// ============================================================================

// @Summary Get Version Information
// @Description Get PostgreSQL version information and capabilities for a collector
// @Tags Version
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Success 200 {object} models.VersionInfoResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/version [get]
func (s *Server) handleGetVersionInfo(c *gin.Context) {
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

	// Get capabilities for this version
	capabilities, err := s.postgres.GetVersionCapabilities(ctx, version.Major)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Get collector mode
	mode, err := s.postgres.GetCollectorMode(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.VersionInfoResponse{
		CollectorID:  collectorID,
		Version:      *version,
		Capabilities: *capabilities,
		Mode:         *mode,
		Timestamp:    time.Now(),
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Supported Versions
// @Description Get list of all supported PostgreSQL versions with EOL information
// @Tags Version
// @Produce json
// @Security Bearer
// @Success 200 {object} models.SupportedVersionsResponse
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/versions/supported [get]
func (s *Server) handleGetSupportedVersions(c *gin.Context) {
	versions := s.postgres.GetAllSupportedVersions()

	resp := &models.SupportedVersionsResponse{
		Versions: versions,
		Count:    len(versions),
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Collector Mode
// @Description Get the deployment mode configuration for a collector (decentralized vs centralized)
// @Tags Version
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Success 200 {object} models.CollectorModeResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/mode [get]
func (s *Server) handleGetCollectorMode(c *gin.Context) {
	collectorIDStr := c.Param("id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get collector mode configuration
	config, err := s.postgres.GetCollectorMode(ctx, collectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.CollectorModeResponse{
		CollectorID:    config.CollectorID,
		Mode:           config.Mode,
		ConnectionType: config.ConnectionType,
		UseTLS:         config.UseTLS,
		TLSEnabled:     config.UseTLS && config.TLSConfig.CertFile != "",
	}

	c.JSON(http.StatusOK, resp)
}
