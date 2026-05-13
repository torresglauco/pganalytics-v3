package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/torresglauco/pganalytics-v3/backend/internal/services/index_advisor"
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
			"database_id":     databaseID,
			"recommendations": []interface{}{},
			"count":           0,
		})
		return
	}

	if len(recommendations) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"database_id":     databaseID,
			"recommendations": []interface{}{},
			"count":           0,
		})
		return
	}

	// Return response with recommendations
	c.JSON(http.StatusOK, gin.H{
		"database_id":     databaseID,
		"recommendations": recommendations,
		"count":           len(recommendations),
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
		"status":            "created",
		"message":           "Index created successfully",
		"table":             recommendation.TableName,
		"columns":           recommendation.ColumnNames,
	})
}

// handleGetUnusedIndexes returns a list of unused indexes for a database
// GET /api/v1/index-advisor/database/:database_id/unused
// This endpoint returns indexes that are not being used and could potentially be removed
func (s *Server) handleGetUnusedIndexes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get database ID from URL parameter
	databaseIDStr := c.Param("database_id")
	if databaseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	databaseID, err := strconv.Atoi(databaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid database_id format"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get connection string for the monitored database
	var connectionString *string
	err = s.postgres.QueryRowContext(ctx,
		`SELECT connection_string FROM pganalytics.postgresql_instances WHERE id = $1 AND is_active = true`,
		databaseID,
	).Scan(&connectionString)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "database not found"})
			return
		}
		s.logger.Error("Failed to get database connection info", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection info"})
		return
	}

	if connectionString == nil || *connectionString == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "database connection not configured"})
		return
	}

	// Connect to the monitored database
	monitoredDB, err := sql.Open("postgres", *connectionString)
	if err != nil {
		s.logger.Error("Failed to connect to monitored database", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to monitored database"})
		return
	}
	defer monitoredDB.Close()

	// Use UnusedIndexDetector to find unused indexes
	detector := index_advisor.NewUnusedIndexDetector(monitoredDB)
	indexes, err := detector.FindUnused(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to find unused indexes", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve unused indexes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database_id":    databaseIDStr,
		"unused_indexes": indexes,
		"count":          len(indexes),
	})
}

// handleEstimateIndexImpact estimates the impact of creating an index
// POST /api/v1/index-advisor/database/:database_id/estimate-impact
// This endpoint uses hypopg to estimate the cost improvement of creating an index
func (s *Server) handleEstimateIndexImpact(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Get database ID from URL parameter
	databaseIDStr := c.Param("database_id")
	if databaseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	databaseID, err := strconv.Atoi(databaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid database_id format"})
		return
	}

	// Parse request body
	var req struct {
		TableName string   `json:"table_name" binding:"required"`
		Columns   []string `json:"columns" binding:"required"`
		QueryText string   `json:"query_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Get connection string for the monitored database
	var connectionString *string
	err = s.postgres.QueryRowContext(ctx,
		`SELECT connection_string FROM pganalytics.postgresql_instances WHERE id = $1 AND is_active = true`,
		databaseID,
	).Scan(&connectionString)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "database not found"})
			return
		}
		s.logger.Error("Failed to get database connection info", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection info"})
		return
	}

	if connectionString == nil || *connectionString == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "database connection not configured"})
		return
	}

	// Connect to the monitored database
	monitoredDB, err := sql.Open("postgres", *connectionString)
	if err != nil {
		s.logger.Error("Failed to connect to monitored database", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to monitored database"})
		return
	}
	defer monitoredDB.Close()

	// Use HypoIndexTester to estimate impact
	tester := index_advisor.NewHypoIndexTester(monitoredDB, s.logger)
	impact, err := tester.EstimateImpact(ctx, req.QueryText, req.TableName, req.Columns)
	if err != nil {
		// Check if hypopg is not available
		if err.Error() == "hypopg extension not installed" {
			c.JSON(http.StatusOK, gin.H{
				"error":           "hypopg extension not installed on monitored database",
				"fallback_note":   "Install hypopg extension with: CREATE EXTENSION hypopg;",
				"table_name":      req.TableName,
				"columns":         req.Columns,
				"improvement_pct": 0,
			})
			return
		}
		s.logger.Error("Failed to estimate index impact", zap.Error(err), zap.Int("database_id", databaseID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to estimate index impact"})
		return
	}

	c.JSON(http.StatusOK, impact)
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

	// Estimate index impact using hypopg
	indexAdvisor.POST("/database/:database_id/estimate-impact", s.AuthMiddleware(), s.handleEstimateIndexImpact)
}
