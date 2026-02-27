package api

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

// ============================================================================
// RDS INSTANCE MANAGEMENT ENDPOINTS
// ============================================================================

// @Summary Create Managed Instance
// @Description Register a new RDS PostgreSQL instance for monitoring (admin only)
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateManagedInstanceRequest true "Managed Instance details"
// @Success 201 {object} models.ManagedInstance
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/rds-instances [post]
func (s *Server) handleCreateManagedInstance(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can register Managed Instances", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Parse request
	var req models.CreateManagedInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind Managed Instance request", zap.Error(err))
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

	// Set defaults for optional fields
	if req.Port == 0 {
		req.Port = 5432 // Default PostgreSQL port
	}
	if req.SSLMode == "" {
		req.SSLMode = "require" // Default SSL mode
	}
	if req.MonitoringInterval == 0 {
		req.MonitoringInterval = 60 // Default to 60 seconds
	}
	if req.ConnectionTimeout == 0 {
		req.ConnectionTimeout = 30 // Default to 30 seconds
	}
	if !req.SSLEnabled {
		req.SSLEnabled = true // Enable SSL by default
	}

	// Default AWS region if not provided
	if req.AWSRegion == "" {
		req.AWSRegion = "us-east-1"
	}

	// Default environment if not provided
	if req.Environment == "" {
		req.Environment = "development"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Test connection to Managed Instance
	if err := testRDSConnection(ctx, req.RDSEndpoint, req.Port, req.MasterUsername, req.MasterPassword); err != nil {
		s.logger.Warn("RDS connection test failed", zap.String("endpoint", req.RDSEndpoint), zap.Error(err))
		// Don't fail here, allow registration even if connection fails (may be temporary)
	}

	// Store master password securely in secrets table
	var secretID *int
	if req.MasterPassword != "" {
		encryptedPassword, err := s.secretManager.Encrypt(req.MasterPassword)
		if err != nil {
			s.logger.Error("Failed to encrypt password", zap.Error(err))
			errResp := apperrors.BadRequest("Failed to encrypt password", err.Error())
			c.JSON(errResp.StatusCode, errResp)
			return
		}

		// Create secret in database
		id, err := s.postgres.CreateSecret(ctx, fmt.Sprintf("managed_instance_password_%s", req.Name), encryptedPassword)
		if err != nil {
			s.logger.Error("Failed to store encrypted password", zap.Error(err))
			errResp := apperrors.ToAppError(err)
			c.JSON(errResp.StatusCode, errResp)
			return
		}
		secretID = &id
	}

	// Create Managed Instance
	instance, err := s.postgres.CreateManagedInstance(ctx, &req, secretID, user.ID)
	if err != nil {
		s.logger.Error("Failed to create Managed Instance", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Managed Instance created",
		zap.String("instance_name", instance.Name),
		zap.String("endpoint", instance.RDSEndpoint),
		zap.String("created_by", user.Username),
	)

	c.JSON(201, instance)
}

// @Summary List Managed Instances
// @Description Get all active Managed Instances
// @Tags RDS Management
// @Produce json
// @Security Bearer
// @Success 200 {array} models.ManagedInstance
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/rds-instances [get]
func (s *Server) handleListManagedInstances(c *gin.Context) {
	// Get current user from context
	_, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	instances, err := s.postgres.ListManagedInstances(ctx)
	if err != nil {
		s.logger.Error("Failed to list Managed Instances", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if instances == nil {
		instances = make([]*models.ManagedInstance, 0)
	}

	c.JSON(200, instances)
}

// @Summary Get Managed Instance
// @Description Get a specific Managed Instance by ID
// @Tags RDS Management
// @Produce json
// @Security Bearer
// @Param id path int true "Managed Instance ID"
// @Success 200 {object} models.ManagedInstance
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id} [get]
func (s *Server) handleGetManagedInstance(c *gin.Context) {
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
		errResp := apperrors.BadRequest("Invalid Managed Instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	instance, err := s.postgres.GetManagedInstance(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get Managed Instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(200, instance)
}

// @Summary Update Managed Instance
// @Description Update an existing Managed Instance (admin only)
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Managed Instance ID"
// @Param request body models.UpdateManagedInstanceRequest true "Managed Instance details to update"
// @Success 200 {object} models.ManagedInstance
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id} [put]
func (s *Server) handleUpdateManagedInstance(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can update Managed Instances", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid Managed Instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Parse request
	var req models.UpdateManagedInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind Managed Instance request", zap.Error(err))
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

	// Set defaults for optional fields
	if req.Port == 0 {
		req.Port = 5432 // Default PostgreSQL port
	}
	if req.SSLMode == "" {
		req.SSLMode = "require" // Default SSL mode
	}
	if req.MonitoringInterval == 0 {
		req.MonitoringInterval = 60 // Default to 60 seconds
	}
	if req.ConnectionTimeout == 0 {
		req.ConnectionTimeout = 30 // Default to 30 seconds
	}
	if !req.SSLEnabled {
		req.SSLEnabled = true // Enable SSL by default
	}
	if req.Status == "" {
		req.Status = "registered" // Default status
	}

	// Default AWS region if not provided
	if req.AWSRegion == "" {
		req.AWSRegion = "us-east-1"
	}

	// Default environment if not provided
	if req.Environment == "" {
		req.Environment = "development"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Update Managed Instance
	instance, err := s.postgres.UpdateManagedInstance(ctx, id, &req, user.ID)
	if err != nil {
		s.logger.Error("Failed to update Managed Instance", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Managed Instance updated",
		zap.String("instance_name", instance.Name),
		zap.String("endpoint", instance.RDSEndpoint),
		zap.String("updated_by", user.Username),
	)

	c.JSON(200, instance)
}

// @Summary Delete Managed Instance
// @Description Delete (soft delete) an Managed Instance (admin only)
// @Tags RDS Management
// @Security Bearer
// @Param id path int true "Managed Instance ID"
// @Success 204 {object} nil
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id} [delete]
func (s *Server) handleDeleteManagedInstance(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can delete Managed Instances", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid Managed Instance ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := s.postgres.DeleteManagedInstance(ctx, id); err != nil {
		s.logger.Error("Failed to delete Managed Instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Managed Instance deleted",
		zap.Int("id", id),
		zap.String("deleted_by", user.Username),
	)

	c.JSON(204, nil)
}

// @Summary Test RDS Connection (Direct)
// @Description Test connection to an Managed Instance without creating it (for forms)
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.TestManagedInstanceConnectionRequest true "Connection test parameters"
// @Success 200 {object} models.TestConnectionResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/rds-instances/test-connection-direct [post]
func (s *Server) handleTestManagedInstanceConnectionDirect(c *gin.Context) {
	// Get current user from context
	_, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req models.TestManagedInstanceConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind test connection request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate required fields
	if req.RDSEndpoint == "" {
		errResp := apperrors.BadRequest("RDS endpoint is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	if req.Port == 0 {
		req.Port = 5432
	}
	if req.Username == "" {
		errResp := apperrors.BadRequest("Username is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	if req.Password == "" {
		errResp := apperrors.BadRequest("Password is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Test connection
	testErr := testRDSConnection(ctx, req.RDSEndpoint, req.Port, req.Username, req.Password)

	response := &models.TestConnectionResponse{
		Success: testErr == nil,
	}

	if testErr != nil {
		response.Error = fmt.Sprintf("Connection test failed - Endpoint: %s:%d - Error: %s", req.RDSEndpoint, req.Port, testErr.Error())
	} else {
		response.Error = ""
	}

	c.JSON(200, response)
}

// @Summary Test RDS Connection (Existing Instance)
// @Description Test connection to an existing registered Managed Instance
// @Tags RDS Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Managed Instance ID"
// @Param request body models.TestConnectionRequest true "Connection test parameters"
// @Success 200 {object} models.TestConnectionResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/rds-instances/{id}/test-connection [post]
func (s *Server) handleTestManagedInstanceConnection(c *gin.Context) {
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
		errResp := apperrors.BadRequest("Invalid Managed Instance ID", "")
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

	// Get Managed Instance
	instance, err := s.postgres.GetManagedInstance(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get Managed Instance", zap.Int("id", id), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get username from request or from stored instance
	username := req.Username
	if username == "" {
		username = instance.MasterUsername
	}

	// Get password from request or from stored secret
	password := req.Password

	// If no password provided in request and we have a secret, decrypt it
	if password == "" && instance.SecretID != nil {
		secret, err := s.postgres.GetSecret(ctx, *instance.SecretID)
		if err != nil {
			s.logger.Error("Failed to retrieve stored password", zap.Int("secret_id", *instance.SecretID), zap.Error(err))
			errResp := apperrors.BadRequest("Failed to retrieve stored password", "")
			c.JSON(errResp.StatusCode, errResp)
			return
		}

		// Decrypt password
		decryptedPassword, err := s.secretManager.Decrypt(string(secret.SecretEncrypted))
		if err != nil {
			s.logger.Error("Failed to decrypt password", zap.Error(err))
			errResp := apperrors.BadRequest("Failed to decrypt password", "")
			c.JSON(errResp.StatusCode, errResp)
			return
		}
		password = decryptedPassword
	}

	// Test connection
	testErr := testRDSConnection(ctx, instance.RDSEndpoint, instance.Port, username, password)

	response := &models.TestConnectionResponse{
		Success: testErr == nil,
	}

	if testErr != nil {
		response.Error = fmt.Sprintf("Connection test failed - Endpoint: %s:%d - Error: %s", instance.RDSEndpoint, instance.Port, testErr.Error())
	} else {
		response.Error = ""
	}

	c.JSON(200, response)
}

// Helper function to test RDS connection
func testRDSConnection(ctx context.Context, endpoint string, port int, username, password string) error {
	// Build PostgreSQL connection string
	// Try with SSL require first (for RDS), fallback to disable if that fails
	for _, sslMode := range []string{"require", "prefer", "disable"} {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s connect_timeout=5",
			endpoint, port, username, password, sslMode)

		// Attempt to open connection (validates credentials and connectivity)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			continue
		}

		// Test the actual connection by pinging the database
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr := db.PingContext(pingCtx)
		cancel()
		db.Close()

		if pingErr == nil {
			return nil // Success!
		}

		// If it's an SSL error, try next mode
		if sslMode == "require" && pingErr != nil {
			continue
		}

		// For other errors on the last attempt, return the error
		if sslMode == "disable" {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", pingErr)
		}
	}

	return fmt.Errorf("failed to connect to PostgreSQL: all SSL modes attempted")
}
