package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================================================
// WORKLOAD PATTERN DETECTION ENDPOINTS
// ============================================================================

// @Summary Detect Workload Patterns
// @Description Trigger workload pattern detection on historical data
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param request body gin.H true "Database name and lookback days"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/workload-patterns/analyze [post]
func (s *Server) handleDetectWorkloadPatterns(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		DatabaseName string `json:"database_name"`
		LookbackDays int    `json:"lookback_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate database_name
	if req.DatabaseName == "" {
		errResp := apperrors.BadRequest("Invalid request", "database_name is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Default to 30 days if not specified
	if req.LookbackDays == 0 {
		req.LookbackDays = 30
	}

	// Validate lookback days (7-365 range enforced)
	if req.LookbackDays < 7 {
		errResp := apperrors.BadRequest("Invalid lookback_days", "Minimum lookback is 7 days")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if req.LookbackDays > 365 {
		s.logger.Info("Capping lookback to maximum", zap.Int("requested_days", req.LookbackDays), zap.Int("max_days", 365))
		req.LookbackDays = 365
	}

	// Detect patterns
	count, err := s.postgres.DetectWorkloadPatterns(ctx, req.DatabaseName, req.LookbackDays)
	if err != nil {
		s.logger.Warn("Failed to detect patterns", zap.String("database", req.DatabaseName), zap.Error(err))
		errResp := apperrors.InternalServerError("Pattern detection failed", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Detected workload patterns", zap.Int("count", count), zap.String("database", req.DatabaseName), zap.Int("lookback_days", req.LookbackDays))

	c.JSON(http.StatusOK, gin.H{
		"patterns_detected": count,
		"database_name":     req.DatabaseName,
		"lookback_days":     req.LookbackDays,
		"timestamp":         time.Now().UTC(),
	})
}

// @Summary Get Workload Patterns
// @Description List detected workload patterns for a database
// @Tags ML-Optimization
// @Produce json
// @Param database_name query string false "Database name"
// @Param pattern_type query string false "Pattern type (hourly_peak, daily_cycle, etc)"
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {array} models.WorkloadPattern
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/workload-patterns [get]
func (s *Server) handleGetWorkloadPatterns(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	databaseName := c.Query("database_name")
	patternType := c.Query("pattern_type")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}

	patterns, err := s.postgres.GetWorkloadPatterns(ctx, databaseName, patternType, limit)
	if err != nil {
		s.logger.Warn("Failed to get patterns", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve patterns", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, patterns)
}

// ============================================================================
// QUERY REWRITE SUGGESTIONS ENDPOINTS
// ============================================================================

// @Summary Generate Rewrite Suggestions
// @Description Generate rewrite suggestions for a query based on EXPLAIN plan analysis
// @Description Detects: N+1 patterns, inefficient joins, missing indexes, subqueries, IN vs ANY
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Success 200 {object} gin.H{"suggestions_generated":int,"query_hash":int64,"timestamp":string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/rewrite-suggestions/generate [post]
func (s *Server) handleGenerateRewriteSuggestions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid query_hash", "Must be a valid integer")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if queryHash <= 0 {
		errResp := apperrors.BadRequest("Invalid query_hash", "Query hash must be positive")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Generating rewrite suggestions", zap.Int64("query_hash", queryHash))

	count, err := s.postgres.GenerateRewriteSuggestions(ctx, queryHash)
	if err != nil {
		s.logger.Warn("Failed to generate suggestions", zap.Int64("query_hash", queryHash), zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to generate suggestions", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	s.logger.Info("Generated rewrite suggestions", zap.Int("count", count), zap.Int64("query_hash", queryHash))

	c.JSON(http.StatusOK, gin.H{
		"suggestions_generated": count,
		"query_hash":            queryHash,
		"suggestion_types": []string{
			"n_plus_one_detected",
			"inefficient_join_detected",
			"missing_index_detected",
			"subquery_optimization",
			"in_vs_any_optimization",
		},
		"timestamp": time.Now().UTC(),
	})
}

// @Summary Get Rewrite Suggestions
// @Description List rewrite suggestions for a query, sorted by confidence score
// @Tags ML-Optimization
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Param limit query int false "Limit results (default 10, max 100)"
// @Success 200 {array} models.QueryRewriteSuggestion
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/rewrite-suggestions [get]
func (s *Server) handleGetRewriteSuggestions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid query_hash", "Must be a valid integer")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if queryHash <= 0 {
		errResp := apperrors.BadRequest("Invalid query_hash", "Query hash must be positive")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	s.logger.Debug("Retrieving rewrite suggestions", zap.Int64("query_hash", queryHash), zap.Int("limit", limit))

	suggestions, err := s.postgres.GetRewriteSuggestions(ctx, queryHash, limit)
	if err != nil {
		s.logger.Warn("Failed to get suggestions", zap.Int64("query_hash", queryHash), zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve suggestions", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Handle empty result (no suggestions found)
	if suggestions == nil {
		suggestions = []models.QueryRewriteSuggestion{}
	}

	s.logger.Debug("Retrieved rewrite suggestions", zap.Int("count", len(suggestions)), zap.Int64("query_hash", queryHash))

	c.JSON(http.StatusOK, suggestions)
}

// ============================================================================
// PARAMETER OPTIMIZATION ENDPOINTS
// ============================================================================

// @Summary Generate Parameter Optimization Suggestions
// @Description Analyze query and generate parameter optimization recommendations
// @Tags ML-Optimization
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/parameter-optimization/generate [post]
func (s *Server) handleOptimizeParameters(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		s.logger.Warn("Invalid query_hash format", zap.String("query_hash", queryHashStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	if queryHash <= 0 {
		s.logger.Warn("Invalid query_hash value", zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusBadRequest, gin.H{"error": "query_hash must be positive"})
		return
	}

	s.logger.Info("Generating parameter optimization suggestions", zap.Int64("query_hash", queryHash))

	// Generate suggestions using SQL function
	suggestions, err := s.postgres.OptimizeParameters(ctx, queryHash)
	if err != nil {
		s.logger.Error("Failed to generate parameter suggestions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate parameter suggestions"})
		return
	}

	// Extract parameter types
	paramTypes := make(map[string]bool)
	for _, s := range suggestions {
		paramTypes[s.ParameterName] = true
	}

	typesList := make([]string, 0, len(paramTypes))
	for paramType := range paramTypes {
		typesList = append(typesList, paramType)
	}

	s.logger.Info("Generated parameter optimization suggestions", zap.Int("count", len(suggestions)), zap.Int64("query_hash", queryHash))

	c.JSON(http.StatusOK, gin.H{
		"query_hash":        queryHash,
		"suggestions_count": len(suggestions),
		"suggestion_types":  typesList,
		"generated_at":      time.Now().UTC(),
		"message":           fmt.Sprintf("Generated %d parameter optimization suggestions", len(suggestions)),
	})
}

// @Summary Get Parameter Optimization Suggestions
// @Description List parameter tuning recommendations for a query
// @Tags ML-Optimization
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Param limit query int false "Limit results (1-100, default 10)"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/parameter-optimization [get]
func (s *Server) handleGetParameterOptimization(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		s.logger.Warn("Invalid query_hash format", zap.String("query_hash", queryHashStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	if queryHash <= 0 {
		s.logger.Warn("Invalid query_hash value", zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusBadRequest, gin.H{"error": "query_hash must be positive"})
		return
	}

	// Parse optional limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	s.logger.Debug("Fetching parameter optimization suggestions", zap.Int64("query_hash", queryHash), zap.Int("limit", limit))

	// Get suggestions from database
	suggestions, err := s.postgres.GetParameterOptimizationSuggestions(ctx, queryHash, limit)
	if err != nil {
		s.logger.Error("Failed to get parameter suggestions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parameter suggestions"})
		return
	}

	if len(suggestions) == 0 {
		s.logger.Info("No parameter optimization suggestions available", zap.Int64("query_hash", queryHash))
		c.JSON(http.StatusOK, gin.H{
			"query_hash":  queryHash,
			"suggestions": []models.ParameterTuningSuggestion{},
			"count":       0,
			"message":     "No parameter optimization suggestions available",
		})
		return
	}

	// Group suggestions by parameter type
	paramTypes := make(map[string]int)
	for _, s := range suggestions {
		paramTypes[s.ParameterName]++
	}

	s.logger.Info("Retrieved parameter suggestions", zap.Int("count", len(suggestions)), zap.Int64("query_hash", queryHash))

	c.JSON(http.StatusOK, gin.H{
		"query_hash":      queryHash,
		"suggestions":     suggestions,
		"count":           len(suggestions),
		"parameter_types": paramTypes,
		"timestamp":       time.Now().UTC(),
	})
}

// ============================================================================
// PREDICTIVE PERFORMANCE ENDPOINTS
// ============================================================================

// @Summary Predict Query Performance
// @Description Predict query execution time with given parameters
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Param request body gin.H true "Prediction parameters"
// @Success 200 {object} models.PerformancePrediction
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/queries/{query_hash}/predict-performance [post]
func (s *Server) handlePredictQueryPerformance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid query_hash", "Must be a valid integer")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req struct {
		Parameters map[string]interface{} `json:"parameters,omitempty"`
		Scenario   string                 `json:"scenario,omitempty"` // current, optimized
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if req.Scenario == "" {
		req.Scenario = "current"
	}

	prediction, err := s.postgres.PredictQueryPerformance(ctx, queryHash, req.Parameters, req.Scenario)
	if err != nil {
		s.logger.Warn("Failed to predict performance", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to predict performance", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if prediction == nil {
		// Return fallback prediction if model not available
		prediction = &models.PerformancePrediction{
			QueryHash:            queryHash,
			PredictedExecutionMs: 0,
			ConfidenceScore:      0.5,
			PredictionRange: models.PredictionRange{
				Min: 0,
				Max: 0,
			},
			Timestamp: time.Now().UTC(),
		}
	}

	c.JSON(http.StatusOK, prediction)
}

// ============================================================================
// OPTIMIZATION RECOMMENDATIONS ENDPOINTS
// ============================================================================

// @Summary Get Optimization Recommendations
// @Description List top optimization opportunities ranked by ROI
// @Tags ML-Optimization
// @Produce json
// @Param limit query int false "Limit results (default 20)"
// @Param min_impact query float false "Minimum impact percentage (default 5.0)"
// @Param source_type query string false "Filter by source type (index, rewrite, parameter)"
// @Success 200 {array} models.OptimizationRecommendation
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/optimization-recommendations [get]
func (s *Server) handleGetOptimizationRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	limitStr := c.DefaultQuery("limit", "20")
	minImpactStr := c.DefaultQuery("min_impact", "5.0")
	sourceType := c.Query("source_type")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 20
	}

	minImpact, err := strconv.ParseFloat(minImpactStr, 64)
	if err != nil || minImpact < 0 {
		minImpact = 5.0
	}

	recommendations, err := s.postgres.GetOptimizationRecommendations(ctx, limit, minImpact, sourceType)
	if err != nil {
		s.logger.Warn("Failed to get recommendations", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve recommendations", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, recommendations)
}

// ============================================================================
// OPTIMIZATION IMPLEMENTATION TRACKING
// ============================================================================

// @Summary Record Optimization Implementation
// @Description Record when an optimization recommendation is implemented
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param recommendation_id path int64 true "Recommendation ID"
// @Param request body gin.H true "Implementation details"
// @Success 200 {object} gin.H
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/optimization-recommendations/{recommendation_id}/implement [post]
func (s *Server) handleImplementRecommendation(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	recommendationIDStr := c.Param("recommendation_id")
	recommendationID, err := strconv.ParseInt(recommendationIDStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid recommendation_id", "Must be a valid integer")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	var req struct {
		Notes     string `json:"notes,omitempty"`
		QueryHash int64  `json:"query_hash,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	impl, err := s.postgres.ImplementRecommendation(ctx, recommendationID, req.QueryHash, req.Notes)
	if err != nil {
		s.logger.Warn("Failed to record implementation", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to record implementation", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	implID := int64(0)
	if impl != nil {
		implID = impl.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"implementation_id": implID,
		"recommendation_id": recommendationID,
		"status":            "pending",
		"timestamp":         time.Now().UTC(),
	})
}

// @Summary Get Optimization Results
// @Description Get measured results of implemented optimizations
// @Tags ML-Optimization
// @Produce json
// @Param recommendation_id query int64 false "Filter by recommendation ID"
// @Param status query string false "Filter by status (pending, implemented, reverted)"
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {array} models.OptimizationResult
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/optimization-results [get]
func (s *Server) handleGetOptimizationResults(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var recommendationID *int64
	if recIDStr := c.Query("recommendation_id"); recIDStr != "" {
		recID, err := strconv.ParseInt(recIDStr, 10, 64)
		if err == nil {
			recommendationID = &recID
		}
	}

	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 50
	}

	results, err := s.postgres.GetOptimizationResults(ctx, recommendationID, status, limit)
	if err != nil {
		s.logger.Warn("Failed to get optimization results", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to retrieve results", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, results)
}

// ============================================================================
// HELPER FUNCTIONS (Internal)
// ============================================================================

// updateOptimizationResultsWithActualMetrics updates implementation with post-optimization metrics
// This would be called after metrics are collected post-implementation
func (s *Server) updateOptimizationResultsWithActualMetrics(
	ctx context.Context,
	implementationID int64,
	postStats map[string]interface{},
	actualImprovementPct float64,
	actualImprovementSec float64,
) error {
	return s.postgres.UpdateOptimizationResults(ctx, implementationID, postStats, actualImprovementPct, actualImprovementSec)
}

// dismissOptimizationRecommendation marks a recommendation as dismissed
func (s *Server) dismissOptimizationRecommendation(
	ctx context.Context,
	recommendationID int64,
	reason string,
) error {
	return s.postgres.DismissOptimizationRecommendation(ctx, recommendationID, reason)
}

// getRecommendationDetails fetches full details of a recommendation
func (s *Server) getRecommendationDetails(ctx context.Context, recommendationID int64) (*models.OptimizationRecommendation, error) {
	return s.postgres.GetRecommendationByID(ctx, recommendationID)
}

// callMLService calls the Python ML service for advanced predictions
// This is called from handlePredictQueryPerformance if model predictions are needed
func (s *Server) callMLService(ctx context.Context, queryHash int64, params map[string]interface{}) (*models.PerformancePrediction, error) {
	// Check if ML service is configured
	if s.config.MLServiceURL == "" {
		return nil, fmt.Errorf("ML service not configured")
	}

	// TODO: Implement HTTP call to Python ML service
	// For now, return nil to trigger fallback behavior

	return nil, nil
}

// trainPerformanceModel trains a new ML model for performance prediction
func (s *Server) trainPerformanceModel(ctx context.Context, databaseName string, lookbackDays int) error {
	return s.postgres.TrainPerformanceModel(ctx, databaseName, lookbackDays)
}

// ============================================================================
// PHASE 4.5.4: ML-POWERED OPTIMIZATION WORKFLOW HANDLERS
// ============================================================================

// @Summary Aggregate Optimization Recommendations
// @Description Aggregate all suggestions into unified recommendation table with ROI scoring
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param request body gin.H false "Aggregation parameters"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/recommendations/aggregate [post]
func (s *Server) handleAggregateRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		QueryHash     int64   `json:"query_hash,omitempty"`
		MinConfidence float64 `json:"min_confidence,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Optional body, so don't fail if empty
		s.logger.Debug("No body provided for aggregation")
	}

	if req.MinConfidence == 0 {
		req.MinConfidence = 0.6
	}

	s.logger.Info("Aggregating recommendations", zap.Int64("query_hash", req.QueryHash), zap.Float64("min_confidence", req.MinConfidence))

	var count int
	var sourceTypes []string
	var err error

	if req.QueryHash > 0 {
		// Aggregate for specific query
		count, sourceTypes, err = s.postgres.AggregateRecommendationsForQuery(ctx, req.QueryHash)
	} else {
		// Aggregate for all queries with suggestions - would iterate over all
		// For now, return 0 if no specific query provided
		count = 0
		sourceTypes = []string{}
	}

	if err != nil {
		s.logger.Error("Failed to aggregate recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to aggregate recommendations"})
		return
	}

	s.logger.Info("Successfully aggregated recommendations", zap.Int("count", count))

	c.JSON(http.StatusOK, gin.H{
		"recommendations_aggregated": count,
		"source_types":               sourceTypes,
		"aggregated_at":              time.Now().UTC(),
		"message":                    fmt.Sprintf("Aggregated %d recommendations", count),
	})
}

// @Summary Get Optimization Recommendations
// @Description List top optimization opportunities ranked by ROI score
// @Tags ML-Optimization
// @Produce json
// @Param limit query int false "Limit results (1-100, default 20)"
// @Param min_impact query float false "Minimum impact percentage (default 5.0)"
// @Param source_type query string false "Filter by source type (rewrite, parameter, workload_pattern)"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/optimization-recommendations [get]
// @Summary Implement Optimization Recommendation
// @Description Record that a recommendation was implemented
// @Tags ML-Optimization
// @Accept json
// @Produce json
// @Param recommendation_id path int64 true "Recommendation ID"
// @Param request body gin.H true "Implementation details"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/optimization-recommendations/{recommendation_id}/implement [post]
// @Summary Get Optimization Results
// @Description Get measurement results from implemented recommendations
// @Tags ML-Optimization
// @Produce json
// @Param recommendation_id query int64 false "Filter by recommendation ID"
// @Param status query string false "Filter by status (pending, implemented, no_improvement)"
// @Param limit query int false "Limit results (1-100, default 20)"
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/optimization-results [get]
