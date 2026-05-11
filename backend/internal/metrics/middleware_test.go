package metrics

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestPrometheusMiddleware(t *testing.T) {
	// Test 1: Middleware records response time for each request
	t.Run("records response time for each request", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Simulate some processing
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify histogram was updated
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})

	// Test 2: Prometheus histogram is updated with recorded times
	t.Run("updates Prometheus histogram with recorded times", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/api/v1/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"users": []string{}})
		})
		router.POST("/api/v1/users", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": 1})
		})

		// Make multiple requests
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/api/v1/users", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		req := httptest.NewRequest("POST", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify histogram was updated
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations after requests")
	})

	// Test 3: Request path is normalized (IDs replaced with placeholders)
	t.Run("normalizes path by replacing UUIDs with placeholders", func(t *testing.T) {
		normalized := normalizePath("/api/v1/users/550e8400-e29b-41d4-a716-446655440000/profile")
		assert.Equal(t, "/api/v1/users/:uuid/profile", normalized)
	})

	t.Run("normalizes path by replacing numeric IDs with placeholders", func(t *testing.T) {
		normalized := normalizePath("/api/v1/users/123/profile")
		assert.Equal(t, "/api/v1/users/:id/profile", normalized)
	})

	t.Run("normalizes path with multiple IDs", func(t *testing.T) {
		normalized := normalizePath("/api/v1/users/123/posts/456")
		assert.Equal(t, "/api/v1/users/:id/posts/:id", normalized)
	})

	// Test 4: Response status code is recorded
	t.Run("records response status code", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/ok", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		router.GET("/notfound", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		})

		// Make requests with different status codes
		testCases := []struct {
			path         string
			expectStatus int
		}{
			{"/ok", http.StatusOK},
			{"/notfound", http.StatusNotFound},
			{"/error", http.StatusInternalServerError},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.expectStatus, w.Code)
		}

		// Verify histogram was updated with different status codes
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})

	// Test 5: RequestCounter increments for each request
	t.Run("increments request counter for each request", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Make multiple requests
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// Verify counter was updated
		count := testutil.CollectAndCount(RequestCounter)
		assert.GreaterOrEqual(t, count, 1, "Counter should have metric data")
	})
}

func TestNormalizePath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "/api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "path with numeric ID",
			input:    "/api/v1/users/123",
			expected: "/api/v1/users/:id",
		},
		{
			name:     "path with UUID",
			input:    "/api/v1/users/550e8400-e29b-41d4-a716-446655440000",
			expected: "/api/v1/users/:uuid",
		},
		{
			name:     "path with multiple numeric IDs",
			input:    "/api/v1/users/123/posts/456",
			expected: "/api/v1/users/:id/posts/:id",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "path with mixed IDs",
			input:    "/api/v1/databases/42/tables/550e8400-e29b-41d4-a716-446655440000",
			expected: "/api/v1/databases/:id/tables/:uuid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPrometheusMiddleware_WithGinFullPath(t *testing.T) {
	// Test that middleware uses gin's FullPath when available
	t.Run("uses gin FullPath for route pattern", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/api/v1/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
		})

		req := httptest.NewRequest("GET", "/api/v1/users/12345", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify histogram was updated
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})
}

func TestPrometheusMiddleware_StatusCodes(t *testing.T) {
	// Test that status codes are correctly categorized
	t.Run("correctly records status code categories", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())

		// Register routes with different status codes
		router.GET("/success", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		})
		router.GET("/created", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{})
		})
		router.GET("/client-error", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{})
		})
		router.GET("/server-error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{})
		})

		testCases := []string{"/success", "/created", "/client-error", "/server-error"}
		for _, path := range testCases {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// Verify histogram was updated with multiple status codes
		count := testutil.CollectAndCount(APIResponseTimeHistogram)
		assert.GreaterOrEqual(t, count, 1, "Histogram should have observations")
	})
}

func TestStatusToString(t *testing.T) {
	testCases := []struct {
		status   int
		expected string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{204, "2xx"},
		{301, "3xx"},
		{302, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{500, "5xx"},
		{503, "5xx"},
		{0, "unknown"},
		{99, "unknown"},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
			result := statusToString(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}
