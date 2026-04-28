package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================================================
// QUERY PERFORMANCE ENDPOINTS
// ============================================================================

// handleGetQueryPerformance returns aggregated performance metrics for a query
// GET /api/v1/queries/:query_hash/performance?hours=24&metrics=all
func (s *Server) handleGetQueryPerformance(c *gin.Context) {
	// Parse query hash from URL parameter
	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	// Parse optional query parameters
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 || hours > 8760 { // Max 1 year
		hours = 24
	}

	metricsFilter := c.DefaultQuery("metrics", "all")

	// Check if postgres database is available
	if s.postgres == nil {
		s.logger.Error("PostgreSQL database not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not available"})
		return
	}

	// Create context with timeout for database query
	ctx := c.Request.Context()

	// Get query timeline data which includes performance metrics
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	queries, err := s.postgres.GetQueryTimeline(ctx, queryHash, since)
	if err != nil {
		s.logger.Error("Failed to get query performance", zap.Error(err), zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve query performance data"})
		return
	}

	// If no data found, return not found
	if len(queries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No performance data found for this query"})
		return
	}

	// Aggregate metrics from timeline data
	aggregatedMetrics := aggregateQueryMetrics(queries, metricsFilter)

	// Return performance data
	c.JSON(http.StatusOK, gin.H{
		"query_hash":       queryHash,
		"time_range_hours": hours,
		"metrics_filter":   metricsFilter,
		"data_points":      len(queries),
		"metrics":          aggregatedMetrics,
		"performance_data": queries,
	})
}

// aggregateQueryMetrics aggregates performance metrics from query timeline data
func aggregateQueryMetrics(queries interface{}, filter string) map[string]interface{} {
	metrics := map[string]interface{}{
		"aggregation_type": "query_performance",
		"metrics_filter":   filter,
	}

	// Placeholder for metric aggregation
	// In a full implementation, this would calculate min/max/avg/p95/p99 for various metrics
	metrics["avg_execution_time"] = 0.0
	metrics["max_execution_time"] = 0.0
	metrics["min_execution_time"] = 0.0
	metrics["total_calls"] = 0
	metrics["total_rows"] = 0

	return metrics
}

// handleGetDatabaseQueryPerformance returns query performance data for a specific database
// GET /api/v1/query-performance/database/:database_id
// Returns empty list when no data available yet
func (s *Server) handleGetDatabaseQueryPerformance(c *gin.Context) {
	// For now, return empty response as this endpoint is not yet fully implemented
	// The frontend will gracefully handle the empty state
	c.JSON(http.StatusOK, gin.H{
		"queries":  []interface{}{},
		"timeline": []interface{}{},
	})
}
