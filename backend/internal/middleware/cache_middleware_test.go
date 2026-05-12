package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"go.uber.org/zap"
)

func TestCacheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("generates correct cache key from path and query params", func(t *testing.T) {
		// Create cache manager
		cacheManager := cache.NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer cacheManager.Close()

		// Create gin router
		router := gin.New()

		// Add a test endpoint
		router.GET("/api/v1/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test?param1=value1&param2=value2", nil)

		// Generate cache key using helper
		config := CacheConfig{
			Enabled: true,
			TTL:     5 * time.Minute,
		}
		key := generateCacheKey(c, config)

		if key == "" {
			t.Error("expected non-empty cache key")
		}
	})

	t.Run("returns cached response on cache hit", func(t *testing.T) {
		cacheManager := cache.NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer cacheManager.Close()

		callCount := 0
		router := gin.New()
		router.GET("/api/v1/databases/:id/slow-queries", CacheMiddleware(cacheManager, logger), func(c *gin.Context) {
			callCount++
			c.JSON(http.StatusOK, gin.H{"data": "fresh", "call": callCount})
		})

		// First request - cache miss, handler called
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/databases/123/slow-queries", nil)
		router.ServeHTTP(w1, req1)

		if callCount != 1 {
			t.Errorf("expected handler to be called once, got %d", callCount)
		}

		// Second request - cache hit, handler NOT called again
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/databases/123/slow-queries", nil)
		router.ServeHTTP(w2, req2)

		if callCount != 1 {
			t.Errorf("expected handler to NOT be called again on cache hit, got callCount=%d", callCount)
		}

		// Both responses should be identical (from cache on second request)
		if w1.Body.String() != w2.Body.String() {
			t.Errorf("expected same response from cache, got w1=%s, w2=%s", w1.Body.String(), w2.Body.String())
		}
	})

	t.Run("calls next handler on cache miss", func(t *testing.T) {
		cacheManager := cache.NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer cacheManager.Close()

		handlerCalled := false
		router := gin.New()
		router.GET("/api/v1/databases/:id/slow-queries", CacheMiddleware(cacheManager, logger), func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"data": "fresh"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/databases/123/slow-queries", nil)
		router.ServeHTTP(w, req)

		if !handlerCalled {
			t.Error("expected next handler to be called on cache miss")
		}
	})

	t.Run("stores response on cache miss", func(t *testing.T) {
		cacheManager := cache.NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer cacheManager.Close()

		router := gin.New()
		router.GET("/api/v1/databases/:id/slow-queries", CacheMiddleware(cacheManager, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "stored"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/databases/456/slow-queries", nil)
		router.ServeHTTP(w, req)

		// Check metrics - should have 1 miss
		metrics := cacheManager.GetResponseMetrics()
		if metrics.Misses != 1 {
			t.Errorf("expected 1 miss, got %d", metrics.Misses)
		}
	})

	t.Run("skips caching for non-GET requests", func(t *testing.T) {
		cacheManager := cache.NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer cacheManager.Close()

		handlerCalled := false
		router := gin.New()
		router.POST("/api/v1/test", CacheMiddleware(cacheManager, logger), func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"message": "posted"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test", bytes.NewReader([]byte("{}")))
		router.ServeHTTP(w, req)

		if !handlerCalled {
			t.Error("expected handler to be called for POST request")
		}

		// Check that no cache operations happened
		metrics := cacheManager.GetResponseMetrics()
		if metrics.Hits != 0 || metrics.Misses != 0 {
			t.Errorf("expected no cache operations, got hits=%d, misses=%d", metrics.Hits, metrics.Misses)
		}
	})

	t.Run("respects TTL configuration per endpoint", func(t *testing.T) {
		// Check that different endpoints have different TTLs configured
		if EndpointCacheConfigs["/api/v1/databases/:id/slow-queries"].TTL != 5*time.Minute {
			t.Errorf("expected 5 minute TTL for slow-queries, got %v", EndpointCacheConfigs["/api/v1/databases/:id/slow-queries"].TTL)
		}
		if EndpointCacheConfigs["/api/v1/queries/:hash/timeline"].TTL != 10*time.Minute {
			t.Errorf("expected 10 minute TTL for timeline, got %v", EndpointCacheConfigs["/api/v1/queries/:hash/timeline"].TTL)
		}
		if EndpointCacheConfigs["/api/v1/system/pool-metrics"].TTL != 30*time.Second {
			t.Errorf("expected 30 second TTL for pool-metrics, got %v", EndpointCacheConfigs["/api/v1/system/pool-metrics"].TTL)
		}
	})
}
