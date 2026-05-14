package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// TenantContextConfig holds configuration for tenant context middleware
type TenantContextConfig struct {
	Store  *storage.PostgresDB
	Logger *zap.Logger
}

// TenantContextMiddleware creates a middleware that sets tenant context for RLS
// This middleware must be used after AuthMiddleware to have access to user_id
func TenantContextMiddleware(store *storage.PostgresDB, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by AuthMiddleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			// No user_id - skip tenant context (public endpoints)
			c.Next()
			return
		}

		// Type assert to uuid.UUID
		userID, ok := userIDInterface.(uuid.UUID)
		if !ok {
			logger.Error("Invalid user_id type in context",
				zap.Any("user_id", userIDInterface))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user context",
			})
			c.Abort()
			return
		}

		// Get tenant for user
		tenant, err := store.GetTenantByUserID(c.Request.Context(), userID)
		if err != nil {
			logger.Warn("Failed to get tenant for user",
				zap.String("user_id", userID.String()),
				zap.Error(err))

			// Return 403 Forbidden - user not associated with any tenant
			c.JSON(http.StatusForbidden, gin.H{
				"error": "No tenant associated with user",
			})
			c.Abort()
			return
		}

		// Set tenant context for RLS policies
		err = store.SetTenantSessionVariable(c.Request.Context(), tenant.ID)
		if err != nil {
			logger.Error("Failed to set tenant context for RLS",
				zap.String("tenant_id", tenant.ID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to set tenant context",
			})
			c.Abort()
			return
		}

		// Set tenant_id in gin context for downstream handlers
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_slug", tenant.Slug)

		logger.Debug("Tenant context set",
			zap.String("tenant_id", tenant.ID.String()),
			zap.String("tenant_slug", tenant.Slug),
			zap.String("user_id", userID.String()))

		c.Next()
	}
}

// GetTenantIDFromContext extracts tenant_id from gin context
func GetTenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	tenantIDInterface, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, false
	}

	tenantID, ok := tenantIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return tenantID, true
}

// GetTenantSlugFromContext extracts tenant_slug from gin context
func GetTenantSlugFromContext(c *gin.Context) (string, bool) {
	tenantSlugInterface, exists := c.Get("tenant_slug")
	if !exists {
		return "", false
	}

	tenantSlug, ok := tenantSlugInterface.(string)
	if !ok {
		return "", false
	}

	return tenantSlug, true
}

// SetTenantContext manually sets the tenant context for special cases
// This can be used for system operations that need to act on behalf of a tenant
func SetTenantContext(c *gin.Context, store *storage.PostgresDB, tenantID uuid.UUID) error {
	err := store.SetTenantSessionVariable(c.Request.Context(), tenantID)
	if err != nil {
		return err
	}

	c.Set("tenant_id", tenantID)
	return nil
}

// RequireTenant is a helper middleware that ensures tenant context is set
// Use this for endpoints that absolutely require tenant isolation
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant context required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
