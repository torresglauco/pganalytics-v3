package api

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================================================
// RDS INSTANCE MANAGEMENT ENDPOINTS
// ============================================================================

// @Summary Create RDS Instance
// @Description Register a new RDS PostgreSQL instance for monitoring (admin only)
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateRDSInstanceRequest true "RDS instance details"
// @Success 201 {object} models.RDSInstance
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/rds-instances [post]
func (s *Server) handleCreateRDSInstance(c *gin.Context) {
	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	// Check if user is admin
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Only admins can register RDS instances", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Parse request
	var req models.CreateRDSInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind RDS instance request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate request
	if req.Name == "" {
		errResp := apperrors.BadRequest("Instance name is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if req.RDSEndpoint == "" {
		errResp := apperrors.BadRequest("RDS endpoint is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if req.Port == 0 {
		req.Port = 5432 // Default PostgreSQL port
	}

	if req.AWSRegion == "" {
		errResp := apperrors.BadRequest("AWS region is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Test connection to RDS instance (if credentials provided)
	if req.MasterUsername != "" && req.MasterPassword != "" {
		if err := testRDSConnection(ctx, req.RDSEndpoint, req.Port, req.MasterUsername, req.MasterPassword); err != nil {
			s.logger.Warn("RDS connection test failed", zap.String("endpoint", req.RDSEndpoint), zap.Error(err))
			// Don't fail here, allow registration even if connection fails (may be temporary)
		}
	}

	// TODO: Store master password in secrets table and get secret_id
	var secretID int
	if req.MasterPassword != "" {
		// For now, placeholder - in production, encrypt and store credentials
		secretID = 0
	}

	// Create RDS instance
	instance, err := s.postgres.CreateRDSInstance(ctx, &req, secretID, user.ID)
	if err != nil {
		s.logger.Error("Failed to create RDS instance", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("RDS instance created",
		zap.String("instance_name", instance.Name),
		zap.String("endpoint", instance.RDSEndpoint),
		zap.String("created_by", user.Username),
	)

	c.JSON(201, instance)
}

// @Summary List RDS Instances
// @Description Get all active RDS instances
// @Tags RDS Management
// @Produce json
// @Security Bearer
// @Success 200 {array} models.RDSInstance
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/rds-instances [get]
func (s *Server) handleListRDSInstances(c *gin.Context) {
	// Get current user from context
	_, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	instances, err := s.postgres.ListRDSInstances(ctx)
	if err != nil {
		s.logger.Error("Failed to list RDS instances", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if instances == nil {
		instances = make([]*models.RDSInstance, 0)
	}

	c.JSON(200, instances)
}

// @Summary Get RDS Instance
// @Description Get a specific RDS instance by ID
// @Tags RDS Management
// @Produce json
// @Security Bearer
// @Param id path int true "RDS Instance ID"
// @Success 200 {object} models.RDSInstance
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id} [get]
func (s *Server) handleGetRDSInstance(c *gin.Context) {
	// Get current user from context
	_, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid RDS instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	instance, err := s.postgres.GetRDSInstance(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get RDS instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(200, instance)
}

// @Summary Delete RDS Instance
// @Description Delete (soft delete) an RDS instance (admin only)
// @Tags RDS Management
// @Security Bearer
// @Param id path int true "RDS Instance ID"
// @Success 204 {object} nil
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id} [delete]
func (s *Server) handleDeleteRDSInstance(c *gin.Context) {
	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	// Check if user is admin
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Only admins can delete RDS instances", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid RDS instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := s.postgres.DeleteRDSInstance(ctx, id); err != nil {
		s.logger.Error("Failed to delete RDS instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("RDS instance deleted",
		zap.Int("id", id),
		zap.String("deleted_by", user.Username),
	)

	c.JSON(204, nil)
}

// @Summary Test RDS Connection
// @Description Test connection to an RDS instance
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "RDS Instance ID"
// @Param request body models.TestConnectionRequest true "Connection test parameters"
// @Success 200 {object} models.TestConnectionResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id}/test-connection [post]
func (s *Server) handleTestRDSConnection(c *gin.Context) {
	// Get current user from context
	_, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid RDS instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req models.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind test connection request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get RDS instance
	instance, err := s.postgres.GetRDSInstance(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get RDS instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Test connection
	testErr := testRDSConnection(ctx, instance.RDSEndpoint, instance.Port, req.Username, req.Password)

	response := &models.TestConnectionResponse{
		Success: testErr == nil,
	}

	if testErr != nil {
		response.Error = testErr.Error()
	}

	c.JSON(200, response)
}

// Helper function to test RDS connection
func testRDSConnection(ctx context.Context, endpoint string, port int, username, password string) error {
	// Test TCP connection to RDS instance
	addr := fmt.Sprintf("%s:%d", endpoint, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to RDS instance: %w", err)
	}
	defer conn.Close()

	// TODO: Test actual PostgreSQL connection with provided credentials
	// This would require a PostgreSQL driver and would test the actual database connection

	return nil
}
