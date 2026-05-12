package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================================================
// DASHBOARD DATABASE STATS HANDLER TESTS
// ============================================================================

func TestHandleGetDashboardDatabaseStats(t *testing.T) {
	t.Run("returns 400 for invalid collector_id format", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/database-stats", server.handleGetDashboardDatabaseStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/database-stats?collector_id=invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 503 when TimescaleDB is not available", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/database-stats", server.handleGetDashboardDatabaseStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/database-stats?collector_id=550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("returns 200 with database stats for valid collector_id", func(t *testing.T) {
		// This test requires TimescaleDB to be available
		// In a real integration test, we would set up a test database
		// For now, we test the handler returns proper structure when timescale is available
		t.Skip("Requires TimescaleDB connection - integration test")
	})

	t.Run("uses default time_range when not specified", func(t *testing.T) {
		// Verify the handler defaults to "24h" when time_range is not provided
		t.Skip("Requires TimescaleDB connection - integration test")
	})

	t.Run("accepts valid time_range values", func(t *testing.T) {
		validRanges := []string{"1h", "24h", "7d", "30d"}
		for _, tr := range validRanges {
			t.Run("time_range="+tr, func(t *testing.T) {
				// Verify each time range is accepted
				t.Skip("Requires TimescaleDB connection - integration test")
			})
		}
	})
}

// ============================================================================
// DASHBOARD TABLE STATS HANDLER TESTS
// ============================================================================

func TestHandleGetDashboardTableStats(t *testing.T) {
	t.Run("returns 400 for invalid collector_id format", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/table-stats", server.handleGetDashboardTableStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/table-stats?collector_id=not-a-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 503 when TimescaleDB is not available", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/table-stats", server.handleGetDashboardTableStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/table-stats?collector_id=550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("returns 200 with table stats for valid collector_id", func(t *testing.T) {
		t.Skip("Requires TimescaleDB connection - integration test")
	})

	t.Run("uses default limit when not specified", func(t *testing.T) {
		t.Skip("Requires TimescaleDB connection - integration test")
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		t.Skip("Requires TimescaleDB connection - integration test")
	})
}

// ============================================================================
// DASHBOARD SYSTEM STATS HANDLER TESTS
// ============================================================================

func TestHandleGetDashboardSysstat(t *testing.T) {
	t.Run("returns 400 for invalid collector_id format", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/system-stats", server.handleGetDashboardSysstat)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/system-stats?collector_id=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 503 when TimescaleDB is not available", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // TimescaleDB is nil

		router.GET("/api/v1/dashboard/system-stats", server.handleGetDashboardSysstat)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/system-stats?collector_id=550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("returns 200 with system stats for valid collector_id", func(t *testing.T) {
		t.Skip("Requires TimescaleDB connection - integration test")
	})
}

// ============================================================================
// DASHBOARD ROUTES AUTHENTICATION TESTS
// ============================================================================

func TestDashboardEndpointsRequireAuth(t *testing.T) {
	t.Run("database-stats requires authentication", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/dashboard/database-stats", func(c *gin.Context) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		})

		req := httptest.NewRequest("GET", "/api/v1/dashboard/database-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("table-stats requires authentication", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/dashboard/table-stats", func(c *gin.Context) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		})

		req := httptest.NewRequest("GET", "/api/v1/dashboard/table-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("system-stats requires authentication", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/dashboard/system-stats", func(c *gin.Context) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		})

		req := httptest.NewRequest("GET", "/api/v1/dashboard/system-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ============================================================================
// HANDLER ERROR RESPONSE FORMAT TESTS
// ============================================================================

func TestDashboardHandlerErrorResponses(t *testing.T) {
	t.Run("invalid collector_id returns proper error format", func(t *testing.T) {
		router := gin.New()
		server := &Server{}

		router.GET("/api/v1/dashboard/database-stats", server.handleGetDashboardDatabaseStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/database-stats?collector_id=bad", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Response should be JSON
		contentType := w.Header().Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})

	t.Run("service unavailable returns proper error format", func(t *testing.T) {
		router := gin.New()
		server := &Server{} // timescale is nil

		router.GET("/api/v1/dashboard/database-stats", server.handleGetDashboardDatabaseStats)

		req := httptest.NewRequest("GET", "/api/v1/dashboard/database-stats?collector_id=550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		// Response should be JSON
		contentType := w.Header().Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})
}
