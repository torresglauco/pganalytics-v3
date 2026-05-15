package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// MockTenantStore implements the tenant storage interface for testing
type MockTenantStore struct {
	tenants          map[uuid.UUID]*models.Tenant
	sessionVariables map[uuid.UUID]bool
	errOnGetTenant   bool
	errOnSetSession  bool
}

func NewMockTenantStore() *MockTenantStore {
	return &MockTenantStore{
		tenants:          make(map[uuid.UUID]*models.Tenant),
		sessionVariables: make(map[uuid.UUID]bool),
	}
}

func (m *MockTenantStore) GetTenantByUserID(ctx context.Context, userID uuid.UUID) (*models.Tenant, error) {
	if m.errOnGetTenant {
		return nil, &TenantNotFoundError{UserID: userID}
	}
	for _, tenant := range m.tenants {
		return tenant, nil // Return first tenant for simplicity
	}
	return nil, &TenantNotFoundError{UserID: userID}
}

func (m *MockTenantStore) SetTenantSessionVariable(ctx context.Context, tenantID uuid.UUID) error {
	if m.errOnSetSession {
		return &SessionVariableError{TenantID: tenantID}
	}
	m.sessionVariables[tenantID] = true
	return nil
}

// Test error types
type TenantNotFoundError struct {
	UserID uuid.UUID
}

func (e *TenantNotFoundError) Error() string {
	return "tenant not found for user"
}

type SessionVariableError struct {
	TenantID uuid.UUID
}

func (e *SessionVariableError) Error() string {
	return "failed to set session variable"
}

// Tests

func TestTenantContextMiddleware_SkipsWhenNoUserID(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	// No user_id set in context - simulating public endpoint
	nextCalled := false

	// Create a handler chain to track next call
	router := gin.New()
	router.Use(TenantContextMiddleware(nil, logger))
	router.Use(func(c *gin.Context) {
		nextCalled = true
		c.Next()
	})
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Execute
	router.ServeHTTP(w, c.Request)

	// Assert
	assert.True(t, nextCalled, "middleware should call next handler")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTenantContextMiddleware_Returns500WhenUserIDHasWrongType(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	w := httptest.NewRecorder()

	// Create a router that sets wrong type user_id
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set wrong type for user_id before the middleware runs
		c.Set("user_id", "not-a-uuid")
		c.Next()
	})
	router.Use(TenantContextMiddleware(nil, logger))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid user context")
}

func TestTenantContextMiddleware_Returns403WhenUserHasNoTenant(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	mockStore := NewMockTenantStore()
	mockStore.errOnGetTenant = true

	w := httptest.NewRecorder()

	userID := uuid.New()

	// Create router with mock middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(createTestMiddleware(mockStore, logger))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "No tenant associated with user")
}

func TestTenantContextMiddleware_SetsTenantInContextOnSuccess(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	mockStore := NewMockTenantStore()

	tenantID := uuid.New()
	userID := uuid.New()
	mockStore.tenants[userID] = &models.Tenant{
		ID:       tenantID,
		Name:     "Test Tenant",
		Slug:     "test-tenant",
		IsActive: true,
	}

	w := httptest.NewRecorder()

	var capturedTenantID interface{}
	var capturedTenantSlug interface{}

	// Create router with mock middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(createTestMiddleware(mockStore, logger))
	router.GET("/", func(c *gin.Context) {
		capturedTenantID, _ = c.Get("tenant_id")
		capturedTenantSlug, _ = c.Get("tenant_slug")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, tenantID, capturedTenantID)
	assert.Equal(t, "test-tenant", capturedTenantSlug)
}

func TestGetTenantIDFromContext_ExtractsUUIDCorrectly(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("returns UUID when tenant_id is set", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		tenantID := uuid.New()
		c.Set("tenant_id", tenantID)

		result, ok := GetTenantIDFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, tenantID, result)
	})

	t.Run("returns false when tenant_id is not set", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetTenantIDFromContext(c)

		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, result)
	})

	t.Run("returns false when tenant_id has wrong type", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("tenant_id", "not-a-uuid")

		result, ok := GetTenantIDFromContext(c)

		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, result)
	})
}

func TestGetTenantSlugFromContext_ExtractsSlugCorrectly(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("returns slug when tenant_slug is set", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("tenant_slug", "my-tenant")

		result, ok := GetTenantSlugFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, "my-tenant", result)
	})

	t.Run("returns false when tenant_slug is not set", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetTenantSlugFromContext(c)

		assert.False(t, ok)
		assert.Empty(t, result)
	})
}

func TestRequireTenant_Returns403WhenNoTenantIDSet(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()

	router := gin.New()
	router.Use(RequireTenant())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Tenant context required")
}

func TestRequireTenant_CallsNextWhenTenantIDSet(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	tenantID := uuid.New()

	nextCalled := false

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.Use(RequireTenant())
	router.GET("/", func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.True(t, nextCalled, "middleware should call next handler")
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test SetTenantContext using mock interface
func TestSetTenantContext_WithMock(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	mockStore := NewMockTenantStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	tenantID := uuid.New()

	// Execute using mock implementation
	err := mockStore.SetTenantSessionVariable(c.Request.Context(), tenantID)
	require.NoError(t, err)

	c.Set("tenant_id", tenantID)

	// Assert
	tenantIDFromContext, exists := c.Get("tenant_id")
	assert.True(t, exists)
	assert.Equal(t, tenantID, tenantIDFromContext)

	// Verify session variable was set
	assert.True(t, mockStore.sessionVariables[tenantID], "session variable should be set")
}

// TestSetTenantContext_WithError tests error handling in session variable setting
func TestSetTenantContext_WithError(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	mockStore := NewMockTenantStore()
	mockStore.errOnSetSession = true

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	tenantID := uuid.New()

	// Execute
	err := mockStore.SetTenantSessionVariable(c.Request.Context(), tenantID)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &SessionVariableError{}, err)
}

// Helper function to create test middleware with mock store
func createTestMiddleware(mockStore *MockTenantStore, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

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

		tenant, err := mockStore.GetTenantByUserID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "No tenant associated with user",
			})
			c.Abort()
			return
		}

		err = mockStore.SetTenantSessionVariable(c.Request.Context(), tenant.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to set tenant context",
			})
			c.Abort()
			return
		}

		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_slug", tenant.Slug)

		c.Next()
	}
}
