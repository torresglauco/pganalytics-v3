package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================================================
// INDEX ADVISOR ENDPOINTS
// ============================================================================

// handleGetIndexRecommendations returns recommended indexes for a database via the index-advisor endpoint
// GET /api/v1/index-advisor/database/:database_id/recommendations
// This endpoint returns a list of recommended indexes for performance optimization
func (s *Server) handleGetIndexAdvisorRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get database name/ID from URL parameter
	databaseID := c.Param("database_id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 20
	}

	// Query recommendations from database
	recommendations, err := s.postgres.GetIndexRecommendations(ctx, databaseID, limit)
	if err != nil {
		s.logger.Warn("Failed to get index recommendations", zap.Error(err), zap.String("database_id", databaseID))
		// Return empty list instead of error - recommendations may not be available for this database yet
		c.JSON(http.StatusOK, gin.H{
			"database_id":      databaseID,
			"recommendations":  []interface{}{},
			"count":            0,
		})
		return
	}

	if len(recommendations) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"database_id":      databaseID,
			"recommendations":  []interface{}{},
			"count":            0,
		})
		return
	}

	// Return response with recommendations
	c.JSON(http.StatusOK, gin.H{
		"database_id":      databaseID,
		"recommendations":  recommendations,
		"count":            len(recommendations),
	})
}

// handleCreateIndexFromRecommendation creates an index from a recommendation
// POST /api/v1/index-advisor/recommendation/:recommendation_id/create
// This endpoint executes the create statement for a recommended index and updates the status in the database
func (s *Server) handleCreateIndexFromRecommendation(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Parse recommendation ID from URL
	recommendationIDStr := c.Param("recommendation_id")
	recommendationID, err := strconv.ParseInt(recommendationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recommendation_id format"})
		return
	}

	// Get the recommendation details
	recommendation, err := s.postgres.GetIndexRecommendationByID(ctx, recommendationID)
	if err != nil {
		s.logger.Warn("Failed to get recommendation", zap.Error(err), zap.Int64("recommendation_id", recommendationID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Recommendation not found"})
		return
	}

	// Execute the create statement
	if recommendation.CreateStatement == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recommendation: no create statement"})
		return
	}

	// Execute the index creation against PostgreSQL
	_, err = s.postgres.ExecContext(ctx, recommendation.CreateStatement)
	if err != nil {
		s.logger.Error("Failed to create index", zap.Error(err), zap.String("statement", recommendation.CreateStatement))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create index"})
		return
	}

	// Mark recommendation as implemented (dismiss it since it's no longer a recommendation)
	dismissReason := "Index created successfully"
	err = s.postgres.DismissIndexRecommendation(ctx, recommendationID, &dismissReason)
	if err != nil {
		s.logger.Error("Failed to mark recommendation as implemented", zap.Error(err), zap.Int64("recommendation_id", recommendationID))
		// Don't fail the request - the index was created successfully
	}

	s.logger.Info("Index created from recommendation",
		zap.Int64("recommendation_id", recommendationID),
		zap.String("table", recommendation.TableName))

	c.JSON(http.StatusOK, gin.H{
		"recommendation_id": recommendationID,
		"status":           "created",
		"message":          "Index created successfully",
		"table":            recommendation.TableName,
		"columns":          recommendation.ColumnNames,
	})
}

// handleGetUnusedIndexes returns a list of unused indexes for a database
// GET /api/v1/index-advisor/database/:database_id/unused
// This endpoint returns indexes that are not being used and could potentially be removed
func (s *Server) handleGetUnusedIndexes(c *gin.Context) {
	// Get database name/ID from URL parameter
	databaseID := c.Param("database_id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 20
	}

	// Query unused indexes from database
	// This would call a method like GetUnusedIndexes if implemented in storage
	// For now, we return a structured response showing what would be returned
	unusedIndexes := []map[string]interface{}{}

	// TODO: Implement GetUnusedIndexes method in PostgresDB storage layer
	// This should query pg_stat_user_indexes for indexes with:
	// - idx_scan = 0 (never scanned)
	// - idx_tup_read = 0 and idx_tup_fetch = 0 (no tuples fetched)
	// - Exclude indexes that are part of constraints (primary key, unique, foreign key)
	// - Order by index size to prioritize removing large unused indexes

	// Placeholder query that would be executed against the target database:
	// SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch, pg_size_pretty(pg_relation_size(indexrelid))
	// FROM pg_stat_user_indexes
	// WHERE idx_scan = 0
	// ORDER BY pg_relation_size(indexrelid) DESC
	// LIMIT $1

	c.JSON(http.StatusOK, gin.H{
		"database_id":     databaseID,
		"unused_indexes":  unusedIndexes,
		"count":           len(unusedIndexes),
		"note":            "Implement GetUnusedIndexes in storage layer for full functionality",
	})
}

// registerIndexAdvisorRoutes registers all Index Advisor routes
// This function is called from RegisterRoutes in server.go
func (s *Server) registerIndexAdvisorRoutes(indexAdvisor *gin.RouterGroup) {
	// Get index recommendations for a database
	indexAdvisor.GET("/database/:database_id/recommendations", s.AuthMiddleware(), s.handleGetIndexAdvisorRecommendations)

	// Create index from recommendation
	indexAdvisor.POST("/recommendation/:recommendation_id/create", s.AuthMiddleware(), s.handleCreateIndexFromRecommendation)

	// Get unused indexes for a database
	indexAdvisor.GET("/database/:database_id/unused", s.AuthMiddleware(), s.handleGetUnusedIndexes)
}
