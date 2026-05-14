package api

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/middleware"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// ============================================================================
// TENANT MANAGEMENT ENDPOINTS (SCALE-01, SCALE-02, SCALE-03, SCALE-04)
// ============================================================================

// slugRegex validates URL-safe slugs (alphanumeric and hyphens only)
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// @Summary Get user's tenants
// @Description Get all tenants the authenticated user belongs to
// @Tags Tenants
// @Produce json
// @Security Bearer
// @Success 200 {object} models.TenantListResponse
// @Failure 401 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/tenants [get]
func (s *Server) handleGetTenants(c *gin.Context) {
	// Get user_id from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "no user_id in context")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		errResp := apperrors.InternalServerError("Invalid user context", "user_id type assertion failed")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get all tenants for user
	tenants, err := s.postgres.GetTenantsByUserID(ctx, userID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Convert to response format
	var tenantResponses []*models.TenantResponse
	for _, t := range tenants {
		tenantResponses = append(tenantResponses, &models.TenantResponse{
			ID:        t.ID,
			Name:      t.Name,
			Slug:      t.Slug,
			IsActive:  t.IsActive,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}

	resp := &models.TenantListResponse{
		Count:   len(tenantResponses),
		Tenants: tenantResponses,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Create a new tenant
// @Description Create a new tenant and add the authenticated user as admin
// @Tags Tenants
// @Accept json
// @Produce json
// @Security Bearer
// @Param tenant body models.TenantCreateRequest true "Tenant details"
// @Success 201 {object} models.TenantResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/tenants [post]
func (s *Server) handleCreateTenant(c *gin.Context) {
	// Get user_id from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "no user_id in context")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		errResp := apperrors.InternalServerError("Invalid user context", "user_id type assertion failed")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Bind request body
	var req models.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate name
	if req.Name == "" {
		errResp := apperrors.BadRequest("Name is required", "tenant name cannot be empty")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate slug format (URL-safe)
	if !slugRegex.MatchString(req.Slug) {
		errResp := apperrors.BadRequest("Invalid slug format",
			"slug must be lowercase alphanumeric with hyphens only (e.g., my-tenant)")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Create tenant
	tenant := &models.Tenant{
		Name:     req.Name,
		Slug:     req.Slug,
		IsActive: true,
	}

	err := s.postgres.CreateTenant(ctx, tenant)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	// Add user as admin to the new tenant
	err = s.postgres.AddUserToTenant(ctx, tenant.ID, userID, "admin")
	if err != nil {
		s.logger.Error("Failed to add user as admin to tenant",
			zap.String("tenant_id", tenant.ID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		// Continue - tenant was created successfully
	}

	resp := &models.TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		IsActive:  tenant.IsActive,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	c.JSON(http.StatusCreated, resp)
}

// @Summary Get collectors for tenant
// @Description Get all collectors assigned to a specific tenant
// @Tags Tenants
// @Produce json
// @Security Bearer
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/tenants/{id}/collectors [get]
func (s *Server) handleGetTenantCollectors(c *gin.Context) {
	// Get tenant ID from path
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid tenant ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Verify user has access to this tenant
	userTenantID, hasTenant := middleware.GetTenantIDFromContext(c)
	if !hasTenant {
		errResp := apperrors.Forbidden("Tenant context required",
			"user must be associated with a tenant")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// User can only access their own tenant's collectors
	if userTenantID != tenantID {
		errResp := apperrors.Forbidden("Access denied",
			"cannot access another tenant's data")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Get collectors for tenant
	collectors, err := s.postgres.GetCollectorsByTenantID(ctx, tenantID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      len(collectors),
		"collectors": collectors,
	})
}

// @Summary Assign collector to tenant
// @Description Assign a collector to a specific tenant (admin only)
// @Tags Tenants
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Tenant ID"
// @Param assignment body models.TenantCollectorAssignmentRequest true "Collector assignment"
// @Success 200 {object} map[string]string
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 403 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Failure 500 {object} apperrors.AppError
// @Router /api/v1/tenants/{id}/collectors [post]
func (s *Server) handleAssignCollectorToTenant(c *gin.Context) {
	// Get tenant ID from path
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid tenant ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Verify user is admin of this tenant
	userIDInterface, _ := c.Get("user_id")
	userID := userIDInterface.(uuid.UUID)

	ctx := c.Request.Context()

	role, err := s.postgres.GetUserRoleInTenant(ctx, tenantID, userID)
	if err != nil {
		errResp := apperrors.Forbidden("Access denied", "user is not a member of this tenant")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Only admins can assign collectors
	if role != "admin" {
		errResp := apperrors.Forbidden("Admin required",
			"only tenant admins can assign collectors")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Bind request body
	var req models.TenantCollectorAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request body", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Assign collector to tenant
	err = s.postgres.AssignCollectorToTenant(ctx, tenantID, req.CollectorID)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Collector assigned successfully",
		"tenant_id":    tenantID.String(),
		"collector_id": req.CollectorID.String(),
	})
}
