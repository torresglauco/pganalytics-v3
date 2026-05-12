package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/internal/metrics"
	"go.uber.org/zap"
)

// CacheConfig defines caching behavior for an endpoint
type CacheConfig struct {
	Enabled    bool
	TTL        time.Duration
	CacheByKey []string // Query params to include in cache key
}

// EndpointCacheConfigs maps endpoint patterns to cache configurations
var EndpointCacheConfigs = map[string]CacheConfig{
	"/api/v1/databases/:id/slow-queries": {
		Enabled: true,
		TTL:     5 * time.Minute,
	},
	"/api/v1/queries/:hash/timeline": {
		Enabled: true,
		TTL:     10 * time.Minute,
	},
	"/api/v1/databases/:id/index-stats": {
		Enabled: true,
		TTL:     10 * time.Minute,
	},
	"/api/v1/system/pool-metrics": {
		Enabled: true,
		TTL:     30 * time.Second,
	},
}

// CacheMiddleware creates a caching middleware for API responses
func CacheMiddleware(cacheManager *cache.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Check if endpoint should be cached
		path := c.FullPath()
		config, shouldCache := EndpointCacheConfigs[path]
		if !shouldCache || !config.Enabled {
			c.Next()
			return
		}

		// Generate cache key
		cacheKey := generateCacheKey(c, config)
		startTime := time.Now()

		// Try to get from cache
		if cachedResponse, found := cacheManager.GetResponseCache(cacheKey); found {
			latency := time.Since(startTime)
			metrics.RecordCacheHit("response")
			metrics.RecordCacheLatency("response", "get", latency)
			logger.Debug("Cache hit",
				zap.String("path", path),
				zap.String("cache_key", cacheKey),
				zap.Duration("latency", latency),
			)
			c.Data(http.StatusOK, "application/json", cachedResponse)
			c.Abort()
			return
		}

		latency := time.Since(startTime)
		metrics.RecordCacheMiss("response")
		metrics.RecordCacheLatency("response", "get", latency)
		logger.Debug("Cache miss",
			zap.String("path", path),
			zap.String("cache_key", cacheKey),
			zap.Duration("latency", latency),
		)

		// Capture response
		responseWriter := &responseCaptureWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		c.Next()

		// Cache successful responses only
		if c.Writer.Status() == http.StatusOK && responseWriter.body.Len() > 0 {
			cacheManager.SetResponseCache(cacheKey, responseWriter.body.Bytes())
			metrics.RecordCacheLatency("response", "set", time.Since(startTime))
			logger.Debug("Response cached",
				zap.String("path", path),
				zap.String("cache_key", cacheKey),
			)
		}
	}
}

// generateCacheKey creates a unique cache key from request
func generateCacheKey(c *gin.Context, config CacheConfig) string {
	hasher := sha256.New()
	hasher.Write([]byte(c.FullPath()))
	hasher.Write([]byte(c.Request.URL.RawQuery))

	// Include specific query params if configured
	for _, param := range config.CacheByKey {
		if val := c.Query(param); val != "" {
			hasher.Write([]byte(param + "=" + val))
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// responseCaptureWriter captures response body for caching
type responseCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
