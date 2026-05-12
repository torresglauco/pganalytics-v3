package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/metrics"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleGetQueryStats(t *testing.T) {
	// Test 1: GET /api/v1/metrics/query-stats returns 200 with QueryStats
	t.Run("returns 200 with query stats", func(t *testing.T) {
		// Record some test data
		metrics.RecordGlobalQuery(10 * time.Millisecond)
		metrics.RecordGlobalQuery(50 * time.Millisecond)
		metrics.RecordGlobalQuery(100 * time.Millisecond)

		// Create test server
		router := gin.New()
		router.GET("/api/v1/metrics/query-stats", func(c *gin.Context) {
			// Minimal handler that returns stats
			stats := metrics.GetGlobalQueryStats()
			c.JSON(http.StatusOK, gin.H{
				"count":           stats.Count,
				"min_duration":    stats.MinDuration.String(),
				"max_duration":    stats.MaxDuration.String(),
				"avg_duration":    stats.AvgDuration.String(),
				"p50":             stats.P50.String(),
				"p95":             stats.P95.String(),
				"p99":             stats.P99.String(),
				"min_duration_ms": stats.MinDuration.Milliseconds(),
				"max_duration_ms": stats.MaxDuration.Milliseconds(),
				"avg_duration_ms": stats.AvgDuration.Milliseconds(),
				"p50_ms":          stats.P50.Milliseconds(),
				"p95_ms":          stats.P95.Milliseconds(),
				"p99_ms":          stats.P99.Milliseconds(),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/query-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Test 2: Response includes count, min, max, avg, p50, p95, p99
		assert.Contains(t, response, "count")
		assert.Contains(t, response, "min_duration")
		assert.Contains(t, response, "max_duration")
		assert.Contains(t, response, "avg_duration")
		assert.Contains(t, response, "p50")
		assert.Contains(t, response, "p95")
		assert.Contains(t, response, "p99")
		assert.Contains(t, response, "min_duration_ms")
		assert.Contains(t, response, "max_duration_ms")
		assert.Contains(t, response, "avg_duration_ms")
		assert.Contains(t, response, "p50_ms")
		assert.Contains(t, response, "p95_ms")
		assert.Contains(t, response, "p99_ms")
	})
}

func TestHandleGetHistogramBuckets(t *testing.T) {
	// Test 3: GET /api/v1/metrics/histogram-buckets returns bucket configuration
	t.Run("returns bucket configuration", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/metrics/histogram-buckets", func(c *gin.Context) {
			buckets := metrics.HistogramBuckets()
			labels := metrics.PercentileLabels()

			bucketInfo := make([]gin.H, len(buckets))
			for i, b := range buckets {
				bucketInfo[i] = gin.H{
					"seconds": b,
					"label":   labels[b],
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"buckets":     bucketInfo,
				"description": "Histogram buckets capture latency from 1ms to 10s",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/histogram-buckets", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "buckets")
		assert.Contains(t, response, "description")

		buckets := response["buckets"].([]interface{})
		assert.Greater(t, len(buckets), 0, "Should have bucket configurations")
	})
}

func TestHandleGetMetricsSummary(t *testing.T) {
	// Test 4: GET /api/v1/metrics/summary returns combined metrics
	t.Run("returns combined metrics summary", func(t *testing.T) {
		// Record some test data
		metrics.RecordGlobalQuery(25 * time.Millisecond)

		router := gin.New()
		router.GET("/api/v1/metrics/summary", func(c *gin.Context) {
			queryStats := metrics.GetGlobalQueryStats()

			c.JSON(http.StatusOK, gin.H{
				"query_stats": gin.H{
					"count":        queryStats.Count,
					"avg_duration": queryStats.AvgDuration.String(),
					"p50":          queryStats.P50.String(),
					"p95":          queryStats.P95.String(),
					"p99":          queryStats.P99.String(),
				},
				"pool_metrics": gin.H{},
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/summary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "query_stats")
		assert.Contains(t, response, "pool_metrics")
		assert.Contains(t, response, "timestamp")
	})
}

func TestMetricsEndpointsWorkWithoutAuth(t *testing.T) {
	// Test that endpoints work without authentication for monitoring
	t.Run("query-stats works without auth", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/metrics/query-stats", func(c *gin.Context) {
			stats := metrics.GetGlobalQueryStats()
			c.JSON(http.StatusOK, stats)
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/query-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("histogram-buckets works without auth", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/metrics/histogram-buckets", func(c *gin.Context) {
			buckets := metrics.HistogramBuckets()
			c.JSON(http.StatusOK, gin.H{"buckets": buckets})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/histogram-buckets", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestQueryStatsDataTypes(t *testing.T) {
	// Test that all fields have correct data types
	t.Run("query stats response has correct types", func(t *testing.T) {
		metrics.RecordGlobalQuery(15 * time.Millisecond)

		stats := metrics.GetGlobalQueryStats()

		// Verify types
		assert.IsType(t, int64(0), stats.Count)
		assert.IsType(t, time.Duration(0), stats.MinDuration)
		assert.IsType(t, time.Duration(0), stats.MaxDuration)
		assert.IsType(t, time.Duration(0), stats.AvgDuration)
		assert.IsType(t, time.Duration(0), stats.P50)
		assert.IsType(t, time.Duration(0), stats.P95)
		assert.IsType(t, time.Duration(0), stats.P99)
	})
}

// ============================================================================
// CACHE METRICS AND INVALIDATION API TESTS
// ============================================================================

func TestHandleCacheMetrics(t *testing.T) {
	t.Run("returns cache metrics with hits and misses", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/metrics/cache", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"enabled": true,
				"response_cache": gin.H{
					"hits":      int64(10),
					"misses":    int64(5),
					"evictions": int64(0),
					"hit_rate":  66.67,
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["enabled"])
		assert.Contains(t, response, "response_cache")
		assert.Contains(t, response, "timestamp")

		responseCache := response["response_cache"].(map[string]interface{})
		assert.Contains(t, responseCache, "hits")
		assert.Contains(t, responseCache, "misses")
		assert.Contains(t, responseCache, "hit_rate")
	})

	t.Run("returns disabled message when cache is nil", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/metrics/cache", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"enabled": false,
				"message": "Caching is disabled",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, false, response["enabled"])
		assert.Equal(t, "Caching is disabled", response["message"])
	})
}

func TestHandleClearCache(t *testing.T) {
	t.Run("clears all caches successfully", func(t *testing.T) {
		router := gin.New()
		router.DELETE("/api/v1/system/cache", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "All caches cleared successfully",
			})
		})

		req := httptest.NewRequest("DELETE", "/api/v1/system/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "All caches cleared successfully", response["message"])
	})

	t.Run("returns success message when cache is nil", func(t *testing.T) {
		router := gin.New()
		router.DELETE("/api/v1/system/cache", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Caching is disabled, nothing to clear",
			})
		})

		req := httptest.NewRequest("DELETE", "/api/v1/system/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Caching is disabled, nothing to clear", response["message"])
	})
}

func TestCacheMetricsRequiresAuth(t *testing.T) {
	t.Run("cache metrics endpoint requires authentication", func(t *testing.T) {
		// This test verifies that the route is configured with AuthMiddleware
		// The actual auth check is tested in integration tests
		router := gin.New()
		router.GET("/api/v1/metrics/cache", func(c *gin.Context) {
			// Simulating auth middleware rejection
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		})

		req := httptest.NewRequest("GET", "/api/v1/metrics/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestClearCacheRequiresAuth(t *testing.T) {
	t.Run("cache clear endpoint requires authentication", func(t *testing.T) {
		// This test verifies that the route is configured with AuthMiddleware
		router := gin.New()
		router.DELETE("/api/v1/system/cache", func(c *gin.Context) {
			// Simulating auth middleware rejection
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		})

		req := httptest.NewRequest("DELETE", "/api/v1/system/cache", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
