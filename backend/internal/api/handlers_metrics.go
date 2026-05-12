package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/internal/metrics"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// SCHEMA METRICS ENDPOINTS
// ============================================================================

// @Summary Get Schema Metrics
// @Description Get schema information for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/schema [get]
func (s *Server) handleGetSchemaMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get query parameters
	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	// Get schema metrics
	metrics, err := s.postgres.GetSchemaMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_schema",
		Count:      len(metrics.Tables),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// LOCK METRICS ENDPOINTS
// ============================================================================

// @Summary Get Lock Metrics
// @Description Get lock information for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/locks [get]
func (s *Server) handleGetLockMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetLockMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_locks",
		Count:      len(metrics.ActiveLocks),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// BLOAT METRICS ENDPOINTS
// ============================================================================

// @Summary Get Bloat Metrics
// @Description Get table and index bloat metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/bloat [get]
func (s *Server) handleGetBloatMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetBloatMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_bloat",
		Count:      len(metrics.TableBloat),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CACHE METRICS ENDPOINTS
// ============================================================================

// @Summary Get Cache Hit Metrics
// @Description Get cache hit ratio metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/cache-hits [get]
func (s *Server) handleGetCacheMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetCacheMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_cache",
		Count:      len(metrics.TableCacheHit),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// CONNECTION METRICS ENDPOINTS
// ============================================================================

// @Summary Get Connection Metrics
// @Description Get connection tracking metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/connections [get]
func (s *Server) handleGetConnectionMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetConnectionMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_connections",
		Count:      len(metrics.ConnectionSummary),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// EXTENSION METRICS ENDPOINTS
// ============================================================================

// @Summary Get Extension Metrics
// @Description Get extension inventory metrics for a collector
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param database query string false "Database name"
// @Param limit query int false "Result limit" default(100)
// @Param offset query int false "Result offset" default(0)
// @Success 200 {object} models.MetricsResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/collectors/{collector_id}/extensions [get]
func (s *Server) handleGetExtensionMetrics(c *gin.Context) {
	collectorIDStr := c.Param("collector_id")
	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	database := c.Query("database")
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	ctx := c.Request.Context()
	var dbPtr *string
	if database != "" {
		dbPtr = &database
	}

	metrics, err := s.postgres.GetExtensionMetrics(ctx, collectorID, dbPtr, limit, offset)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	resp := &models.MetricsResponse{
		MetricType: "pg_extensions",
		Count:      len(metrics.Extensions),
		Timestamp:  time.Now(),
		Data:       metrics,
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// GENERAL METRICS ENDPOINTS (for frontend dashboard)
// ============================================================================

// @Summary Get General Metrics
// @Description Get aggregated metrics across all collectors
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param instance_id query string false "Instance ID"
// @Param time_range query string false "Time range (24h, 7d, 30d)" default(24h)
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/metrics [get]
func (s *Server) handleGetMetrics(c *gin.Context) {
	// Return mock/empty metrics data for frontend
	c.JSON(http.StatusOK, gin.H{
		"topErrors":    []gin.H{},
		"errorCount":   0,
		"warningCount": 0,
		"infoCount":    0,
	})
}

// @Summary Get Error Trend
// @Description Get error trend data over time
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param instance_id query string false "Instance ID"
// @Param hours query int false "Hours back" default(24)
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/metrics/error-trend [get]
func (s *Server) handleGetErrorTrend(c *gin.Context) {
	// Return mock error trend data for frontend
	c.JSON(http.StatusOK, []gin.H{})
}

// @Summary Get Log Distribution
// @Description Get log distribution by level
// @Tags Metrics
// @Produce json
// @Security Bearer
// @Param instance_id query string false "Instance ID"
// @Param time_range query string false "Time range (24h, 7d, 30d)" default(24h)
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/metrics/log-distribution [get]
func (s *Server) handleGetLogDistribution(c *gin.Context) {
	// Return mock log distribution data for frontend
	c.JSON(http.StatusOK, []gin.H{})
}

// ============================================================================
// QUERY PERFORMANCE METRICS ENDPOINTS
// ============================================================================

// handleGetQueryStats returns query performance statistics
// GET /api/v1/metrics/query-stats
func (s *Server) handleGetQueryStats(c *gin.Context) {
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
}

// handleGetHistogramBuckets returns histogram bucket configuration
// GET /api/v1/metrics/histogram-buckets
func (s *Server) handleGetHistogramBuckets(c *gin.Context) {
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
}

// handleGetMetricsSummary returns a combined metrics summary
// GET /api/v1/metrics/summary
func (s *Server) handleGetMetricsSummary(c *gin.Context) {
	queryStats := metrics.GetGlobalQueryStats()
	poolMetrics := gin.H{}

	if s.postgres != nil {
		poolMetrics["postgres"] = s.postgres.GetAllPoolMetrics()
	}
	if s.timescale != nil {
		poolMetrics["timescale"] = s.timescale.GetPoolMetrics()
	}

	c.JSON(http.StatusOK, gin.H{
		"query_stats": gin.H{
			"count":           queryStats.Count,
			"min_duration":    queryStats.MinDuration.String(),
			"max_duration":    queryStats.MaxDuration.String(),
			"avg_duration":    queryStats.AvgDuration.String(),
			"p50":             queryStats.P50.String(),
			"p95":             queryStats.P95.String(),
			"p99":             queryStats.P99.String(),
			"min_duration_ms": queryStats.MinDuration.Milliseconds(),
			"max_duration_ms": queryStats.MaxDuration.Milliseconds(),
			"avg_duration_ms": queryStats.AvgDuration.Milliseconds(),
			"p50_ms":          queryStats.P50.Milliseconds(),
			"p95_ms":          queryStats.P95.Milliseconds(),
			"p99_ms":          queryStats.P99.Milliseconds(),
		},
		"pool_metrics": poolMetrics,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}

// ============================================================================
// APPLICATION CACHE METRICS AND INVALIDATION ENDPOINTS
// ============================================================================

// handleAppCacheMetrics returns application cache performance metrics
// GET /api/v1/metrics/cache
func (s *Server) handleAppCacheMetrics(c *gin.Context) {
	if s.cacheManager == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Caching is disabled",
		})
		return
	}

	cacheMetrics := s.cacheManager.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"response_cache": gin.H{
			"hits":      cacheMetrics.ResponseCacheMetrics.Hits,
			"misses":    cacheMetrics.ResponseCacheMetrics.Misses,
			"evictions": cacheMetrics.ResponseCacheMetrics.Evictions,
			"hit_rate":  calculateHitRate(cacheMetrics.ResponseCacheMetrics),
		},
		"feature_cache": gin.H{
			"hits":      cacheMetrics.FeatureCacheMetrics.Hits,
			"misses":    cacheMetrics.FeatureCacheMetrics.Misses,
			"evictions": cacheMetrics.FeatureCacheMetrics.Evictions,
			"hit_rate":  calculateHitRate(cacheMetrics.FeatureCacheMetrics),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleClearCache clears all caches
// DELETE /api/v1/system/cache
func (s *Server) handleClearCache(c *gin.Context) {
	if s.cacheManager == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Caching is disabled, nothing to clear",
		})
		return
	}

	s.cacheManager.Clear()
	s.logger.Info("Cache cleared via API")

	c.JSON(http.StatusOK, gin.H{
		"message": "All caches cleared successfully",
	})
}

// calculateHitRate computes cache hit rate percentage
func calculateHitRate(m cache.CacheMetrics) float64 {
	total := m.Hits + m.Misses
	if total == 0 {
		return 0
	}
	return float64(m.Hits) / float64(total) * 100
}

// ============================================================================
// DASHBOARD AGGREGATE METRICS ENDPOINTS
// ============================================================================

// @Summary Get Dashboard Database Stats
// @Description Get pre-computed database statistics from TimescaleDB aggregates
// @Tags Dashboard
// @Produce json
// @Security Bearer
// @Param collector_id query string true "Collector ID"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} timescale.DatabaseStatsAggregate
// @Failure 400 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/dashboard/database-stats [get]
func (s *Server) handleGetDashboardDatabaseStats(c *gin.Context) {
	collectorIDStr := c.Query("collector_id")
	if collectorIDStr == "" {
		errResp := apperrors.BadRequest("Missing collector_id", "collector_id query parameter is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if TimescaleDB is available
	if s.timescale == nil {
		errResp := apperrors.ServiceUnavailable("TimescaleDB not available", "Dashboard aggregates require TimescaleDB connection")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get time range with default
	timeRange := c.DefaultQuery("time_range", "24h")

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "24h" // Default to 24h if invalid
	}

	ctx := c.Request.Context()
	stats, err := s.timescale.GetDashboardDatabaseStats(ctx, collectorID, timeRange)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"time_range": timeRange,
		"count":      len(stats),
	})
}

// @Summary Get Dashboard Table Stats
// @Description Get pre-computed table statistics from TimescaleDB aggregates
// @Tags Dashboard
// @Produce json
// @Security Bearer
// @Param collector_id query string true "Collector ID"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Param limit query int false "Result limit" default(100)
// @Success 200 {object} timescale.TableStatsAggregate
// @Failure 400 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/dashboard/table-stats [get]
func (s *Server) handleGetDashboardTableStats(c *gin.Context) {
	collectorIDStr := c.Query("collector_id")
	if collectorIDStr == "" {
		errResp := apperrors.BadRequest("Missing collector_id", "collector_id query parameter is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if TimescaleDB is available
	if s.timescale == nil {
		errResp := apperrors.ServiceUnavailable("TimescaleDB not available", "Dashboard aggregates require TimescaleDB connection")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get time range with default
	timeRange := c.DefaultQuery("time_range", "24h")

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "24h" // Default to 24h if invalid
	}

	// Get limit with default and max
	limit := 100
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	ctx := c.Request.Context()
	stats, err := s.timescale.GetDashboardTableStats(ctx, collectorID, timeRange, limit)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"time_range": timeRange,
		"count":      len(stats),
		"limit":      limit,
	})
}

// @Summary Get Dashboard System Stats
// @Description Get pre-computed system statistics from TimescaleDB aggregates
// @Tags Dashboard
// @Produce json
// @Security Bearer
// @Param collector_id query string true "Collector ID"
// @Param time_range query string false "Time range (1h, 24h, 7d, 30d)" default(24h)
// @Success 200 {object} timescale.SysstatAggregate
// @Failure 400 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/dashboard/system-stats [get]
func (s *Server) handleGetDashboardSysstat(c *gin.Context) {
	collectorIDStr := c.Query("collector_id")
	if collectorIDStr == "" {
		errResp := apperrors.BadRequest("Missing collector_id", "collector_id query parameter is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	collectorID, err := uuid.Parse(collectorIDStr)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Check if TimescaleDB is available
	if s.timescale == nil {
		errResp := apperrors.ServiceUnavailable("TimescaleDB not available", "Dashboard aggregates require TimescaleDB connection")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get time range with default
	timeRange := c.DefaultQuery("time_range", "24h")

	// Validate time range
	validRanges := map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}
	if !validRanges[timeRange] {
		timeRange = "24h" // Default to 24h if invalid
	}

	ctx := c.Request.Context()
	stats, err := s.timescale.GetDashboardSysstat(ctx, collectorID, timeRange)
	if err != nil {
		c.JSON(err.(*apperrors.AppError).StatusCode, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"time_range": timeRange,
		"count":      len(stats),
	})
}
