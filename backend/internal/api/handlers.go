package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/metrics"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const version = "3.0.0-alpha"

// ============================================================================
// HEALTH & SYSTEM ENDPOINTS
// ============================================================================

// @Summary Health Check
// @Description Check system health status
// @Tags Health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /api/v1/health [get]
func (s *Server) handleHealth(c *gin.Context) {
	ctx := c.Request.Context()

	healthResp := &models.HealthResponse{
		Status:      "ok",
		Version:     version,
		Timestamp:   time.Now(),
		DatabaseOk:  s.postgres.Health(ctx),
		TimescaleOk: s.timescale.Health(ctx),
	}

	if !healthResp.DatabaseOk || !healthResp.TimescaleOk {
		healthResp.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, healthResp)
		return
	}

	c.JSON(http.StatusOK, healthResp)
}

// @Summary Get Version
// @Description Get API version
// @Tags System
// @Produce json
// @Success 200 {object} gin.H
// @Router /version [get]
func (s *Server) handleVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"api":     "1.0.0",
	})
}

// ============================================================================
// AUTHENTICATION ENDPOINTS
// ============================================================================

// @Summary User Login
// @Description Authenticate user and get JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/login [post]
// @Summary Create User (Admin Only)
// @Description Create a new user account (admin-only endpoint)
// @Tags Administration
// @Accept json
// @Produce json
// @Security Bearer
// @Param req body models.CreateUserRequest true "Create user request"
// @Success 201 {object} models.User
// @Failure 400 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/users [post]
func (s *Server) handleCreateUser(c *gin.Context) {
	// Get current user from context (added by middleware)
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	// Check if user is admin
	if user.Role != "admin" {
		errResp := apperrors.Forbidden("Only admins can create users", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind create user request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Debug("Create user attempt", zap.String("username", req.Username), zap.String("role", req.Role))

	// Hash password
	passwordHash, err := s.authService.PasswordManager.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create user in database
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	newUser, err := s.postgres.CreateUserWithRole(ctx, req.Username, req.Email, passwordHash, req.FullName, req.Role)
	if err != nil {
		s.logger.Error("Failed to create user",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User created by admin",
		zap.String("username", req.Username),
		zap.String("created_by", user.Username),
		zap.Int("user_id", newUser.ID),
	)

	c.JSON(http.StatusCreated, newUser)
}

// @Summary List Users
// @Description Get all users (admin only)
// @Tags Administration
// @Security Bearer
// @Produce json
// @Success 200 {array} models.User
// @Failure 403 {object} apperrors.AppError
// @Router /api/v1/users [get]
func (s *Server) handleListUsers(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can list users", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	users, err := s.postgres.ListUsers(ctx)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if users == nil {
		users = make([]*models.User, 0)
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Update User
// @Description Update user (admin only)
// @Tags Administration
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body gin.H true "Update data (role, is_active)"
// @Success 200 {object} models.User
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/users/{id} [put]
func (s *Server) handleUpdateUser(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can update users", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userID := c.Param("id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind update request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user to check if it's the default admin
	userIDInt, err := strconv.Atoi(userID)
	if err == nil {
		userToUpdate, err := s.postgres.GetUserByID(ctx, userIDInt)
		if err == nil && userToUpdate.Username == "admin" {
			// Prevent changing role or disabling the default admin
			if _, hasRole := req["role"]; hasRole {
				errResp := apperrors.Forbidden("Cannot change role of default admin user", "")
				c.JSON(errResp.StatusCode, errResp)
				return
			}
			if isActive, hasIsActive := req["is_active"]; hasIsActive {
				if boolVal, ok := isActive.(bool); ok && !boolVal {
					errResp := apperrors.Forbidden("Cannot disable default admin user", "")
					c.JSON(errResp.StatusCode, errResp)
					return
				}
			}
		}
	}

	updatedUser, err := s.postgres.UpdateUser(ctx, userID, req)
	if err != nil {
		s.logger.Error("Failed to update user", zap.String("user_id", userID), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User updated by admin",
		zap.String("user_id", userID),
		zap.String("updated_by", user.Username),
	)

	c.JSON(http.StatusOK, updatedUser)
}

// @Summary Delete User
// @Description Delete user (admin only, cannot delete default admin)
// @Tags Administration
// @Security Bearer
// @Param id path int true "User ID"
// @Success 204 {object} nil
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/users/{id} [delete]
func (s *Server) handleDeleteUser(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can delete users", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid user ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user to check if it's the default admin
	userToDelete, err := s.postgres.GetUserByID(ctx, userID)
	if err != nil {
		errResp := apperrors.NotFound("User not found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Prevent deletion of default admin
	if userToDelete.Username == "admin" {
		errResp := apperrors.Forbidden("Cannot delete default admin user", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Delete the user
	userIDStr := strconv.Itoa(userID)
	if deleteErr := s.postgres.DeleteUser(ctx, userIDStr); deleteErr != nil {
		s.logger.Error("Failed to delete user", zap.String("user_id", userIDStr), zap.Error(deleteErr))
		errResp := apperrors.ToAppError(deleteErr)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User deleted by admin",
		zap.String("user_id", userIDStr),
		zap.String("deleted_by", user.Username),
	)

	c.JSON(http.StatusNoContent, nil)
}

// @Summary Reset User Password (Admin Only)
// @Description Reset a user's password to a temporary value (admin only)
// @Tags Administration
// @Security Bearer
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.ResetPasswordResponse
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/users/{id}/reset-password [post]
func (s *Server) handleResetUserPassword(c *gin.Context) {
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
		errResp := apperrors.Forbidden("Only admins can reset passwords", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid user ID", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user to check if it exists
	userToReset, err := s.postgres.GetUserByID(ctx, userID)
	if err != nil {
		errResp := apperrors.NotFound("User not found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate temporary password (12 characters, mixed case + digits)
	tempPassword := generateTemporaryPassword(12)

	// Hash the temporary password
	passwordHash, err := s.authService.PasswordManager.HashPassword(tempPassword)
	if err != nil {
		s.logger.Error("Failed to hash temporary password", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Update user password in database
	if err := s.postgres.ResetUserPassword(ctx, userID, passwordHash); err != nil {
		s.logger.Error("Failed to reset user password",
			zap.Int("user_id", userID),
			zap.Error(err),
		)
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User password reset by admin",
		zap.String("user_id", userIDParam),
		zap.String("reset_by", user.Username),
	)

	resp := &models.ResetPasswordResponse{
		Username:    userToReset.Username,
		TempPassword: tempPassword,
		Message:     "Password reset. User must change it on next login",
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Change User Password
// @Description Change password for authenticated user
// @Tags Authentication
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Current and new passwords"
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/auth/change-password [post]
func (s *Server) handleChangePassword(c *gin.Context) {
	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind change password request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if new password is different from old password
	if req.OldPassword == req.NewPassword {
		errResp := apperrors.BadRequest("New password must be different from current password", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get current user with password hash
	userRecord, err := s.postgres.GetUserByID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to fetch user", zap.Int("user_id", user.ID), zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process request", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Verify old password
	if !s.authService.PasswordManager.VerifyPassword(userRecord.PasswordHash, req.OldPassword) {
		s.logger.Debug("Invalid old password", zap.Int("user_id", user.ID))
		errResp := apperrors.Unauthorized("Current password is incorrect", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Hash new password
	newPasswordHash, err := s.authService.PasswordManager.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Update password in database
	if err := s.postgres.ResetUserPassword(ctx, user.ID, newPasswordHash); err != nil {
		s.logger.Error("Failed to change password", zap.Int("user_id", user.ID), zap.Error(err))
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User changed password",
		zap.Int("user_id", user.ID),
		zap.String("username", user.Username),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// generateTemporaryPassword generates a random 12-character password with mixed case and digits
func generateTemporaryPassword(length int) string {
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
		all       = uppercase + lowercase + digits
	)

	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)

	// Ensure at least one character from each category
	categories := []string{uppercase, lowercase, digits}
	for i, category := range categories {
		password[i] = category[rand.Intn(len(category))]
	}

	// Fill the rest randomly
	for i := len(categories); i < length; i++ {
		password[i] = all[rand.Intn(len(all))]
	}

	// Shuffle the password
	for i := length - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

func (s *Server) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Failed to bind login request", zap.Error(err))
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Debug("Login attempt", zap.String("username", req.Username))

	// Authenticate user
	loginResp, err := s.authService.LoginUser(req.Username, req.Password)
	if err != nil {
		s.logger.Debug("Login failed",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("User login successful",
		zap.String("username", req.Username),
		zap.Int("user_id", loginResp.User.ID),
	)

	c.JSON(http.StatusOK, loginResp)
}

// @Summary User Logout
// @Description Logout user (invalidate token)
// @Tags Authentication
// @Security Bearer
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/auth/logout [post]
func (s *Server) handleLogout(c *gin.Context) {
	// Token invalidation would typically be done via token blacklist
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// @Summary Refresh Token
// @Description Refresh JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body gin.H true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (s *Server) handleRefreshToken(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	refreshToken, ok := req["refresh_token"].(string)
	if !ok || refreshToken == "" {
		errResp := apperrors.BadRequest("Missing refresh token", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Refresh user token
	loginResp, err := s.authService.RefreshUserToken(refreshToken)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Token refreshed successfully",
		zap.Int("user_id", loginResp.User.ID),
	)

	c.JSON(http.StatusOK, loginResp)
}

// ============================================================================
// COLLECTOR ENDPOINTS
// ============================================================================

// @Summary Register Collector
// @Description Register a new collector and get authentication credentials
// @Tags Collectors
// @Accept json
// @Produce json
// @Param request body models.CollectorRegisterRequest true "Collector info"
// @Success 200 {object} models.CollectorRegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /api/v1/collectors/register [post]
func (s *Server) handleCollectorRegister(c *gin.Context) {
	var req models.CollectorRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Verify registration secret from request header or fallback to config
	registrationSecret := c.GetHeader("X-Registration-Secret")
	if registrationSecret == "" {
		errResp := apperrors.Unauthorized("Invalid or missing registration secret", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate secret from database (new way)
	validSecret, err := s.postgres.ValidateRegistrationSecret(c.Request.Context(), registrationSecret)
	if err != nil {
		// Fallback to hardcoded config for backward compatibility
		if registrationSecret != s.config.RegistrationSecret {
			errResp := apperrors.Unauthorized("Invalid or missing registration secret", "")
			c.JSON(errResp.StatusCode, errResp)
			return
		}
	} else {
		// Record the usage in audit log
		go func() {
			ipAddress := c.ClientIP()
			_ = s.postgres.RecordRegistrationSecretUsage(
				context.Background(),
				validSecret.ID,
				"", // will be filled after registration
				req.Hostname,
				"pending", // status will be updated after registration completes
				nil,
				ipAddress,
			)
		}()
	}

	// Register collector
	if s.authService == nil {
		// Auth service not initialized - use direct database registration as fallback
		collector := &models.Collector{
			Hostname: req.Hostname,
			Status:   "active",
		}
		if err := s.postgres.CreateCollector(c.Request.Context(), collector); err != nil {
			errResp := apperrors.InternalServerError("Failed to register collector", err.Error())
			c.JSON(errResp.StatusCode, errResp)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"collector_id": collector.ID,
			"status":       "registered",
			"token":        "dev-token",
		})
		return
	}

	registerResp, err := s.authService.RegisterCollector(&req)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Collector registered successfully",
		zap.String("collector_id", registerResp.CollectorID.String()),
		zap.String("hostname", req.Hostname),
	)

	c.JSON(http.StatusOK, registerResp)
}

// @Summary List Collectors
// @Description List all registered collectors
// @Tags Collectors
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors [get]
func (s *Server) handleListCollectors(c *gin.Context) {
	// Parse pagination parameters
	var params models.PaginationParams
	params.Page = 1
	params.PageSize = 20

	if page := c.Query("page"); page != "" {
		_, _ = parseIntParam(page, &params.Page)
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		_, _ = parseIntParam(pageSize, &params.PageSize)
	}

	// TODO: Query collectors from database with pagination
	// For now, return empty list
	resp := &models.PaginatedResponse{
		Data:       []models.Collector{},
		Total:      0,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: 0,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get Collector
// @Description Get collector details by ID
// @Tags Collectors
// @Security Bearer
// @Produce json
// @Param id path string true "Collector ID"
// @Success 200 {object} models.Collector
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/collectors/{id} [get]
func (s *Server) handleGetCollector(c *gin.Context) {
	// TODO: Implement get collector
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Delete Collector
// @Description Deregister a collector
// @Tags Collectors
// @Security Bearer
// @Param id path string true "Collector ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/collectors/{id} [delete]
func (s *Server) handleDeleteCollector(c *gin.Context) {
	// TODO: Implement delete collector
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// METRICS ENDPOINTS
// ============================================================================

// @Summary Push Metrics
// @Description Ingest metrics from a collector (high-volume endpoint)
// @Tags Metrics
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.MetricsPushRequest true "Metrics data"
// @Success 200 {object} models.MetricsPushResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/metrics/push [post]
func (s *Server) handleMetricsPush(c *gin.Context) {
	var req models.MetricsPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid metrics data", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get collector from context (set by CollectorAuthMiddleware)
	collectorClaimsInterface, exists := c.Get("collector_claims")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	collectorClaims, ok := collectorClaimsInterface.(*auth.CollectorClaims)
	if !ok {
		errResp := apperrors.Unauthorized("Invalid authentication claims", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate collector ID matches request
	if collectorClaims.CollectorID != req.CollectorID {
		errResp := apperrors.Unauthorized("Collector ID mismatch", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Process metrics based on type
	startTime := time.Now()
	metricsInserted := 0

	s.logger.Info("Metrics push received",
		zap.Int("metrics_count", len(req.Metrics)),
		zap.String("collector_id", req.CollectorID),
	)

	if req.Metrics != nil && len(req.Metrics) > 0 {
		for _, metric := range req.Metrics {
			// Type assertion to access metric fields
			if metricMap, ok := metric.(map[string]interface{}); ok {
				metricType := ""
				if typeVal, exists := metricMap["type"]; exists {
					metricType = typeVal.(string)
				}

				s.logger.Debug("Processing metric",
					zap.String("metric_type", metricType),
				)

				// Handle query stats metrics
				if metricType == "pg_query_stats" {
					// Handle both individual database metrics and wrapped metrics with arrays
					metricsJSON, _ := json.Marshal(metric)

					// Log the raw metric JSON before unmarshaling
					metricStr := string(metricsJSON)
					if len(metricStr) > 1500 {
						metricStr = metricStr[:1500]
					}
					s.logger.Debug("Raw pg_query_stats JSON", zap.String("json", metricStr))

					// Try to parse as individual database metric first (has "database" field)
					var singleDB models.QueryStatsDB
					var timestamp time.Time
					var databases []models.QueryStatsDB

					// Check if this is an individual database metric or a wrapped one
					if _, hasDB := metricMap["database"]; hasDB {
						// Individual database metric format
						if err := json.Unmarshal(metricsJSON, &singleDB); err != nil {
							s.logger.Error("Failed to unmarshal individual pg_query_stats metric",
								zap.Error(err),
							)
							continue
						}
						databases = append(databases, singleDB)

						// Extract timestamp
						if tsVal, hasTS := metricMap["timestamp"]; hasTS {
							if tsStr, ok := tsVal.(string); ok {
								if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
									timestamp = ts
								}
							}
						}
						s.logger.Debug("Processing individual pg_query_stats metric",
							zap.String("database", singleDB.Database),
							zap.Int("queries_count", len(singleDB.Queries)),
							zap.Time("timestamp", timestamp),
						)
					} else if _, hasDBArray := metricMap["databases"]; hasDBArray {
						// Wrapped format with databases array
						var queryStatsReq models.QueryStatsRequest
						if err := json.Unmarshal(metricsJSON, &queryStatsReq); err != nil {
							s.logger.Error("Failed to unmarshal pg_query_stats metric",
								zap.Error(err),
							)
							continue
						}
						databases = queryStatsReq.Databases
						timestamp = queryStatsReq.Timestamp
						s.logger.Debug("Processing pg_query_stats metric",
							zap.Int("databases_count", len(databases)),
							zap.Time("timestamp", timestamp),
						)
					} else {
						s.logger.Error("Invalid pg_query_stats metric: missing both database and databases fields")
						continue
					}

					// Extract and store individual query statistics
					if databases != nil {
						for _, db := range databases {
							s.logger.Debug("Processing database",
								zap.String("database", db.Database),
								zap.Int("queries_count", len(db.Queries)),
							)
							for _, queryInfo := range db.Queries {
								// Parse collector ID - if it's not a valid UUID, try to look it up or use a placeholder
								collectorUUID := uuid.Nil
								if uid, err := uuid.Parse(req.CollectorID); err == nil {
									collectorUUID = uid
								} else {
									// For collector IDs like "col_demo_001", we'll create a deterministic UUID
									// by hashing the string
									hash := uuid.NewSHA1(uuid.Nil, []byte(req.CollectorID))
									collectorUUID = hash
								}

								stat := &models.QueryStats{
									Time:              timestamp,
									CollectorID:       collectorUUID,
									DatabaseName:      db.Database,
									UserName:          "system", // Set from query info if available
									QueryHash:         queryInfo.Hash,
									QueryText:         queryInfo.Text,
									Calls:             queryInfo.Calls,
									TotalTime:         queryInfo.TotalTime,
									MeanTime:          queryInfo.MeanTime,
									MinTime:           queryInfo.MinTime,
									MaxTime:           queryInfo.MaxTime,
									StddevTime:        queryInfo.StddevTime,
									Rows:              queryInfo.Rows,
									SharedBlksHit:     queryInfo.SharedBlksHit,
									SharedBlksRead:    queryInfo.SharedBlksRead,
									SharedBlksDirtied: queryInfo.SharedBlksDirtied,
									SharedBlksWritten: queryInfo.SharedBlksWritten,
									LocalBlksHit:      queryInfo.LocalBlksHit,
									LocalBlksRead:     queryInfo.LocalBlksRead,
									LocalBlksDirtied:  queryInfo.LocalBlksDirtied,
									LocalBlksWritten:  queryInfo.LocalBlksWritten,
									TempBlksRead:      queryInfo.TempBlksRead,
									TempBlksWritten:   queryInfo.TempBlksWritten,
									BlkReadTime:       queryInfo.BlkReadTime,
									BlkWriteTime:      queryInfo.BlkWriteTime,
									WalRecords:        queryInfo.WalRecords,
									WalFpi:            queryInfo.WalFpi,
									WalBytes:          queryInfo.WalBytes,
									QueryPlanTime:     queryInfo.QueryPlanTime,
									QueryExecTime:     queryInfo.QueryExecTime,
								}

								// Insert individual query stat
								if err := s.postgres.InsertQueryStats(c, req.CollectorID, []*models.QueryStats{stat}); err != nil {
									s.logger.Error("Failed to insert query stat",
										zap.Error(err),
										zap.String("query_hash", fmt.Sprintf("%d", queryInfo.Hash)),
										zap.String("database", db.Database),
									)
								} else {
									metricsInserted++

									// Phase 4.4.2: Trigger EXPLAIN plan capture for slow queries (>1000ms mean time)
									// Strategy: Capture EXPLAIN on first occurrence to minimize overhead
									// Note: Actual EXPLAIN execution happens on the collector side,
									// which sends back the plan via the /api/v1/internal/explain-plans endpoint
									if queryInfo.MeanTime > 1000.0 {
										// Check if EXPLAIN plan already exists for this query
										existingPlan, err := s.postgres.GetExplainPlan(c.Request.Context(), queryInfo.Hash)
										if err == nil && existingPlan != nil {
											// Plan already exists, skip to avoid duplicate EXPLAINs
											continue
										}

										// TODO: Queue EXPLAIN execution request on collector
										// The collector will execute EXPLAIN (ANALYZE, FORMAT JSON, BUFFERS)
										// and send back results via POST /api/v1/internal/explain-plans
										// For now, we log the slow query for manual investigation
										s.logger.Info("Slow query detected (candidate for EXPLAIN)",
											zap.String("query_hash", fmt.Sprintf("%d", queryInfo.Hash)),
											zap.String("database", db.Database),
											zap.Float64("mean_time_ms", queryInfo.MeanTime),
										)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	processingTimeMs := time.Since(startTime).Milliseconds()

	resp := &models.MetricsPushResponse{
		Status:             "success",
		CollectorID:        req.CollectorID,
		MetricsInserted:    metricsInserted,
		BytesReceived:      int(c.Request.ContentLength),
		ProcessingTimeMs:   processingTimeMs,
		NextConfigVersion:  1,
		NextCheckInSeconds: 300,
	}

	s.logger.Info("Metrics pushed successfully",
		zap.String("collector_id", req.CollectorID),
		zap.Int("metrics_inserted", metricsInserted),
		zap.Int64("processing_time_ms", processingTimeMs),
	)

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CONFIGURATION ENDPOINTS
// ============================================================================

// @Summary Get Cache Metrics
// @Description Get cache performance metrics
// @Tags Metrics
// @Security Bearer
// @Accept json
// @Produce json
// @Success 200 {object} metrics.CacheStatusResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/metrics/cache [get]
func (s *Server) handleCacheMetrics(c *gin.Context) {
	if s.cacheManager == nil {
		c.JSON(http.StatusOK, metrics.CacheStatusResponse{
			Enabled: false,
			MaxSize: 0,
			Metrics: nil,
			Message: "Cache is disabled",
		})
		return
	}

	snapshot := metrics.CalculateMetricsSnapshot(s.cacheManager)

	response := metrics.CacheStatusResponse{
		Enabled: true,
		MaxSize: s.config.CacheMaxSize,
		Metrics: snapshot,
		Message: "Cache performance metrics",
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get Collector Config
// @Description Get configuration for a collector (pulled by collector)
// @Tags Configuration
// @Security Bearer
// @Produce plain
// @Param collector_id path string true "Collector ID"
// @Success 200 {string} string "TOML configuration"
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/config/{collector_id} [get]
func (s *Server) handleGetConfig(c *gin.Context) {
	collectorID := c.Param("collector_id")
	ctx := c.Request.Context()

	// Get the latest collector configuration
	config, err := s.postgres.GetCollectorConfig(ctx, collectorID)
	if err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Return TOML format with version info in header
	c.Header("Content-Type", "text/plain")
	c.Header("X-Config-Version", fmt.Sprintf("%d", config.Version))
	c.String(http.StatusOK, config.Config)

	s.logger.Info("Collector config retrieved",
		zap.String("collector_id", collectorID),
		zap.Int("config_version", config.Version),
	)
}

// @Summary Update Collector Config
// @Description Update configuration for a collector (admin only)
// @Tags Configuration
// @Security Bearer
// @Accept plain
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param request body string true "TOML Configuration"
// @Success 200 {object} gin.H
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/config/{collector_id} [put]
func (s *Server) handleUpdateConfig(c *gin.Context) {
	collectorID := c.Param("collector_id")
	ctx := c.Request.Context()

	// Get user from context (set by AuthMiddleware)
	userInterface, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("No user found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		errResp := apperrors.Unauthorized("Invalid user claims", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if user is admin (optional - for now allow any authenticated user)
	// In a real implementation, this would check user.Role == "admin"

	// Read TOML content from request body
	tomlContent, err := c.GetRawData()
	if err != nil {
		errResp := apperrors.BadRequest("Failed to read configuration", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create new config version
	parsedUUID, err := uuid.Parse(collectorID)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID format", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	config := &models.CollectorConfig{
		CollectorID: parsedUUID,
		Config:      string(tomlContent),
		UpdatedBy:   &user.ID,
	}

	if err = s.postgres.CreateCollectorConfig(ctx, config); err != nil {
		errResp := apperrors.ToAppError(err)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Collector config updated",
		zap.String("collector_id", collectorID),
		zap.Int("config_version", config.Version),
		zap.Int("updated_by", user.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"collector_id": collectorID,
		"version":      config.Version,
		"message":      "Configuration updated successfully",
	})
}

// ============================================================================
// SERVER ENDPOINTS
// ============================================================================

// @Summary List Servers
// @Description List all monitored servers
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/servers [get]
func (s *Server) handleListServers(c *gin.Context) {
	// TODO: Implement list servers
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Server
// @Description Get server details by ID
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} models.Server
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/servers/{id} [get]
func (s *Server) handleGetServer(c *gin.Context) {
	// TODO: Implement get server
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Server Metrics
// @Description Get historical metrics for a server
// @Tags Servers
// @Security Bearer
// @Produce json
// @Param id path int true "Server ID"
// @Param metric_type query string false "Metric type" default(pg_stats)
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} gin.H
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/servers/{id}/metrics [get]
func (s *Server) handleGetServerMetrics(c *gin.Context) {
	// TODO: Implement get server metrics
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// ALERTS ENDPOINTS
// ============================================================================

// @Summary List Alerts
// @Description List alerts for monitored resources
// @Tags Alerts
// @Security Bearer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param severity query string false "Severity filter" default(all)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/alerts [get]
func (s *Server) handleListAlerts(c *gin.Context) {
	// TODO: Implement list alerts
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Get Alert
// @Description Get alert details by ID
// @Tags Alerts
// @Security Bearer
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} models.Alert
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/alerts/{id} [get]
func (s *Server) handleGetAlert(c *gin.Context) {
	// TODO: Implement get alert
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// @Summary Acknowledge Alert
// @Description Acknowledge an alert
// @Tags Alerts
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param request body gin.H true "Acknowledgment data"
// @Success 200 {object} gin.H
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/alerts/{id}/acknowledge [post]
func (s *Server) handleAcknowledgeAlert(c *gin.Context) {
	// TODO: Implement acknowledge alert
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// ============================================================================
// QUERY STATISTICS HANDLERS
// ============================================================================

// @Summary Get Top Slow Queries
// @Description Get the slowest queries from a collector
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(20)
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.TopQueriesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors/{collector_id}/queries/slow [get]
func (s *Server) handleGetSlowQueries(c *gin.Context) {
	collectorID := c.Param("collector_id")
	limitStr := c.DefaultQuery("limit", "20")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse parameters
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	queries, err := s.postgres.GetTopSlowQueries(c, collectorID, limit, since)
	if err != nil {
		s.logger.Error("Failed to query slow queries", zap.Error(err), zap.String("collector_id", collectorID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query slow queries",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.TopQueriesResponse{
		ServerID:  uuid.MustParse(collectorID),
		QueryType: "slow",
		Hours:     hours,
		Count:     len(queries),
		Queries:   queries,
	})
}

// @Summary Get Top Frequent Queries
// @Description Get the most frequently executed queries from a collector
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Result limit" default(20)
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.TopQueriesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/collectors/{collector_id}/queries/frequent [get]
func (s *Server) handleGetFrequentQueries(c *gin.Context) {
	collectorID := c.Param("collector_id")
	limitStr := c.DefaultQuery("limit", "20")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse parameters
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	queries, err := s.postgres.GetTopFrequentQueries(c, collectorID, limit, since)
	if err != nil {
		s.logger.Error("Failed to query frequent queries", zap.Error(err), zap.String("collector_id", collectorID))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query frequent queries",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.TopQueriesResponse{
		ServerID:  uuid.MustParse(collectorID),
		QueryType: "frequent",
		Hours:     hours,
		Count:     len(queries),
		Queries:   queries,
	})
}

// @Summary Get Query Timeline
// @Description Get time-series data for a specific query
// @Tags QueryStats
// @Security Bearer
// @Produce json
// @Param query_hash path int true "Query hash from pg_stat_statements"
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} models.QueryTimelineResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/timeline [get]
func (s *Server) handleGetQueryTimeline(c *gin.Context) {
	hashStr := c.Param("query_hash")
	hoursStr := c.DefaultQuery("hours", "24")

	// Parse query hash
	queryHash, err := strconv.ParseInt(hashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid query hash",
			Code:    http.StatusBadRequest,
		})
		return
	}

	hours := 24
	if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
		hours = h
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	stats, err := s.postgres.GetQueryTimeline(c, queryHash, since)
	if err != nil {
		s.logger.Error("Failed to query timeline", zap.Error(err), zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to query timeline",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.QueryTimelineResponse{
		QueryHash: queryHash,
		Data:      stats,
	})
}

// ============================================================================
// REGISTRATION SECRETS HANDLERS
// ============================================================================

// handleCreateRegistrationSecret creates a new registration secret (admin only)
func (s *Server) handleCreateRegistrationSecret(c *gin.Context) {
	// Check admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		errResp := apperrors.Forbidden("Only admins can create registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userId, _ := c.Get("user_id")
	userIdInt, ok := userId.(int)
	if !ok {
		errResp := apperrors.Unauthorized("Invalid user context", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req models.CreateRegistrationSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create the secret
	response, err := s.postgres.CreateRegistrationSecret(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.ExpiresAt,
		userIdInt,
	)

	if err != nil {
		s.logger.Error("Failed to create registration secret", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to create registration secret", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// handleListRegistrationSecrets lists all registration secrets (admin only)
func (s *Server) handleListRegistrationSecrets(c *gin.Context) {
	// Check admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		errResp := apperrors.Forbidden("Only admins can view registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	secrets, err := s.postgres.ListRegistrationSecrets(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to list registration secrets", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to list registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Don't return the actual secret values in list
	for i := range secrets {
		secrets[i].SecretValue = "" // Clear secret values
	}

	c.JSON(http.StatusOK, gin.H{
		"secrets": secrets,
		"count":   len(secrets),
	})
}

// handleGetRegistrationSecret gets a single registration secret (admin only)
func (s *Server) handleGetRegistrationSecret(c *gin.Context) {
	// Check admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		errResp := apperrors.Forbidden("Only admins can view registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	secretId := c.Param("id")
	if secretId == "" {
		errResp := apperrors.BadRequest("Secret ID is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	secret, err := s.postgres.GetRegistrationSecret(c.Request.Context(), secretId)
	if err != nil {
		s.logger.Debug("Registration secret not found", zap.String("id", secretId))
		errResp := apperrors.NotFound("Registration secret not found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Don't return the actual secret value
	secret.SecretValue = ""

	c.JSON(http.StatusOK, secret)
}

// handleUpdateRegistrationSecret updates a registration secret (admin only)
func (s *Server) handleUpdateRegistrationSecret(c *gin.Context) {
	// Check admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		errResp := apperrors.Forbidden("Only admins can update registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	secretId := c.Param("id")
	if secretId == "" {
		errResp := apperrors.BadRequest("Secret ID is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req models.UpdateRegistrationSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Update the secret
	secret, err := s.postgres.UpdateRegistrationSecret(
		c.Request.Context(),
		secretId,
		&req.Name,
		&req.Description,
		req.Active,
	)

	if err != nil {
		s.logger.Error("Failed to update registration secret", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to update registration secret", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Don't return the actual secret value
	secret.SecretValue = ""

	c.JSON(http.StatusOK, secret)
}

// handleDeleteRegistrationSecret deletes a registration secret (admin only)
func (s *Server) handleDeleteRegistrationSecret(c *gin.Context) {
	// Check admin role
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		errResp := apperrors.Forbidden("Only admins can delete registration secrets", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	secretId := c.Param("id")
	if secretId == "" {
		errResp := apperrors.BadRequest("Secret ID is required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	err := s.postgres.DeleteRegistrationSecret(c.Request.Context(), secretId)
	if err != nil {
		s.logger.Error("Failed to delete registration secret", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to delete registration secret", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration secret deleted successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// parseIntParam parses an integer parameter from string
func parseIntParam(value string, target *int) (int, error) {
	_, _ = value, target // Placeholder
	// TODO: Implement proper integer parsing with validation
	return 0, nil
}
