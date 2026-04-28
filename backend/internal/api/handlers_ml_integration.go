package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"go.uber.org/zap"
)

// ============================================================================
// ML SERVICE INTEGRATION ENDPOINTS
// ============================================================================

// @Summary Get ML Service Health
// @Description Check if the ML service is available
// @Tags ML-Service
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/ml/health [get]
func (s *Server) handleMLHealth(c *gin.Context) {
	if s.mlClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "ML service not enabled",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	healthy := s.mlClient.IsHealthy(ctx)
	state := s.mlClient.GetCircuitBreakerState()

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          status,
		"circuit_breaker": state,
		"timestamp":       time.Now().UTC(),
	})
}

// @Summary Train Performance Model
// @Description Trigger ML service to train a performance prediction model
// @Tags ML-Service
// @Accept json
// @Produce json
// @Param request body gin.H true "Training request"
// @Success 202 {object} ml.TrainingResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/train [post]
func (s *Server) handleMLTrain(c *gin.Context) {
	if s.mlClient == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		DatabaseURL  string `json:"database_url" binding:"required"`
		LookbackDays int    `json:"lookback_days"`
		ModelType    string `json:"model_type"`
		ForceRetrain bool   `json:"force_retrain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Set defaults
	if req.LookbackDays == 0 {
		req.LookbackDays = 90
	}
	if req.ModelType == "" {
		req.ModelType = "random_forest"
	}

	// Call ML service
	trainingReq := &ml.TrainingRequest{
		DatabaseURL:  req.DatabaseURL,
		LookbackDays: req.LookbackDays,
		ModelType:    req.ModelType,
		ForceRetrain: req.ForceRetrain,
	}

	resp, err := s.mlClient.TrainPerformanceModel(ctx, trainingReq)
	if err != nil {
		s.logger.Warn("ML training request failed", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("ML service unavailable", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// @Summary Get Training Status
// @Description Get the status of a model training job
// @Tags ML-Service
// @Produce json
// @Param job_id path string true "Training job ID"
// @Success 200 {object} ml.TrainingStatusResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/train/{job_id} [get]
func (s *Server) handleMLTrainingStatus(c *gin.Context) {
	if s.mlClient == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	jobID := c.Param("job_id")
	if jobID == "" {
		errResp := apperrors.BadRequest("Invalid request", "job_id is required")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	status, err := s.mlClient.GetTrainingStatus(ctx, jobID)
	if err != nil {
		s.logger.Warn("Failed to get training status", zap.String("job_id", jobID), zap.Error(err))
		errResp := apperrors.ServiceUnavailable("ML service unavailable", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary Predict Query Execution Time
// @Description Get ML prediction for query execution time
// @Tags ML-Service
// @Accept json
// @Produce json
// @Param request body gin.H true "Prediction request"
// @Success 200 {object} ml.PredictionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/predict [post]
func (s *Server) handleMLPredict(c *gin.Context) {
	if s.mlClient == nil || s.featureExtractor == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var req struct {
		QueryHash  int64                  `json:"query_hash" binding:"required"`
		Parameters map[string]interface{} `json:"parameters"`
		Scenario   string                 `json:"scenario"`
		ModelID    *int64                 `json:"model_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Extract features from database
	features, err := s.featureExtractor.ExtractQueryFeatures(ctx, req.QueryHash)
	if err != nil {
		s.logger.Warn("Failed to extract features", zap.Int64("query_hash", req.QueryHash), zap.Error(err))
		errResp := apperrors.BadRequest("Query not found", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Build prediction request with features
	predReq := &ml.PredictionRequest{
		QueryHash:  req.QueryHash,
		Features:   features.FeatureMap,
		Parameters: req.Parameters,
		Scenario:   req.Scenario,
		ModelID:    req.ModelID,
	}

	// Call ML service
	pred, err := s.mlClient.PredictQueryExecution(ctx, predReq)
	if err != nil {
		s.logger.Warn("ML prediction request failed", zap.Int64("query_hash", req.QueryHash), zap.Error(err))

		// Return fallback prediction based on historical data
		c.JSON(http.StatusOK, gin.H{
			"query_hash":                  req.QueryHash,
			"predicted_execution_time_ms": features.MeanExecutionTimeMs,
			"confidence":                  0.5,
			"range": gin.H{
				"min": features.MinExecutionTimeMs,
				"max": features.MaxExecutionTimeMs,
			},
			"source":  "fallback",
			"warning": "ML service unavailable, using historical baseline",
		})
		return
	}

	c.JSON(http.StatusOK, pred)
}

// @Summary Validate Prediction Accuracy
// @Description Record actual query execution and validate prediction
// @Tags ML-Service
// @Accept json
// @Produce json
// @Param request body gin.H true "Validation request"
// @Success 200 {object} ml.ValidationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/validate [post]
func (s *Server) handleMLValidate(c *gin.Context) {
	if s.mlClient == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req struct {
		PredictionID             string  `json:"prediction_id" binding:"required"`
		QueryHash                int64   `json:"query_hash" binding:"required"`
		PredictedExecutionTimeMs float64 `json:"predicted_execution_time_ms" binding:"required"`
		ActualExecutionTimeMs    float64 `json:"actual_execution_time_ms" binding:"required"`
		ModelVersion             *string `json:"model_version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate actual > 0
	if req.ActualExecutionTimeMs <= 0 {
		errResp := apperrors.BadRequest("Invalid request", "actual_execution_time_ms must be positive")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Call ML service
	valReq := &ml.ValidationRequest{
		PredictionID:             req.PredictionID,
		QueryHash:                req.QueryHash,
		PredictedExecutionTimeMs: req.PredictedExecutionTimeMs,
		ActualExecutionTimeMs:    req.ActualExecutionTimeMs,
		ModelVersion:             req.ModelVersion,
	}

	validation, err := s.mlClient.ValidatePrediction(ctx, valReq)
	if err != nil {
		s.logger.Warn("ML validation request failed", zap.String("prediction_id", req.PredictionID), zap.Error(err))
		errResp := apperrors.ServiceUnavailable("ML service unavailable", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, validation)
}

// @Summary Detect Workload Patterns
// @Description Trigger ML service pattern detection
// @Tags ML-Service
// @Accept json
// @Produce json
// @Param request body gin.H true "Pattern detection request"
// @Success 200 {object} ml.PatternResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/patterns/detect [post]
func (s *Server) handleMLDetectPatterns(c *gin.Context) {
	if s.mlClient == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req struct {
		DatabaseURL  string `json:"database_url" binding:"required"`
		LookbackDays int    `json:"lookback_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if req.LookbackDays == 0 {
		req.LookbackDays = 30
	}

	patternReq := &ml.PatternRequest{
		DatabaseURL:  req.DatabaseURL,
		LookbackDays: req.LookbackDays,
	}

	patterns, err := s.mlClient.DetectWorkloadPatterns(ctx, patternReq)
	if err != nil {
		s.logger.Warn("ML pattern detection failed", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("ML service unavailable", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, patterns)
}

// @Summary Get Query Features
// @Description Extract ML features for a query (for debugging/analysis)
// @Tags ML-Service
// @Produce json
// @Param query_hash path int64 true "Query hash"
// @Success 200 {object} ml.QueryFeatures
// @Failure 404 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/features/{query_hash} [get]
func (s *Server) handleMLGetFeatures(c *gin.Context) {
	if s.featureExtractor == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		errResp := apperrors.BadRequest("Invalid query_hash", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	features, err := s.featureExtractor.ExtractQueryFeatures(ctx, queryHash)
	if err != nil {
		s.logger.Warn("Failed to extract features", zap.Int64("query_hash", queryHash), zap.Error(err))
		errResp := apperrors.NotFound("Query not found", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, features)
}

// @Summary Get Circuit Breaker Status
// @Description Get the status of the ML service circuit breaker
// @Tags ML-Service
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/ml/circuit-breaker [get]
func (s *Server) handleMLCircuitBreakerStatus(c *gin.Context) {
	if s.mlClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "ML service not enabled",
		})
		return
	}

	state := s.mlClient.GetCircuitBreakerState()
	c.JSON(http.StatusOK, gin.H{
		"state":     state,
		"timestamp": time.Now().UTC(),
	})
}

// PredictLatencyRequest represents a request for query latency prediction
type PredictLatencyRequest struct {
	JoinCount     int    `json:"join_count" binding:"required"`
	ScanType      string `json:"scan_type" binding:"required"`
	RowCount      int    `json:"row_count" binding:"required"`
	FilterCount   int    `json:"filter_count"`
	SubqueryCount int    `json:"subquery_count"`
	AggregateType string `json:"aggregate_type"`
}

// @Summary Predict Query Latency
// @Description Predict query execution latency based on query characteristics
// @Tags ML-Service
// @Accept json
// @Produce json
// @Param request body PredictLatencyRequest true "Query latency prediction request"
// @Success 200 {object} gin.H{"predicted_latency_ms":float64,"confidence":float64,"recommendations":[]string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/v1/ml/predict-latency [post]
func (s *Server) handlePredictQueryLatency(c *gin.Context) {
	if s.mlClient == nil {
		errResp := apperrors.ServiceUnavailable("ML service not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req PredictLatencyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate scan type
	validScanTypes := map[string]bool{"seq_scan": true, "index_scan": true, "bitmap_scan": true}
	if !validScanTypes[req.ScanType] {
		errResp := apperrors.BadRequest("Invalid scan_type", "Must be one of: seq_scan, index_scan, bitmap_scan")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate row count
	if req.RowCount < 0 {
		errResp := apperrors.BadRequest("Invalid row_count", "Must be non-negative")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Build feature map for ML service
	features := map[string]interface{}{
		"join_count":     req.JoinCount,
		"scan_type":      scanTypeValue(req.ScanType),
		"row_count":      float64(req.RowCount) / 1000.0, // Normalize row count
		"filter_count":   req.FilterCount,
		"subquery_count": req.SubqueryCount,
		"aggregate_type": aggregateTypeValue(req.AggregateType),
	}

	// Create ML prediction request
	predReq := &ml.PredictionRequest{
		Features:  features,
		QueryHash: 0, // Not using query hash for this endpoint
	}

	// Get prediction from ML service
	pred, err := s.mlClient.PredictQueryExecution(ctx, predReq)
	if err != nil {
		s.logger.Warn("ML latency prediction failed", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("ML prediction failed", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate recommendations based on latency
	recommendations := generateLatencyRecommendations(pred.PredictedExecutionMs, &req)

	c.JSON(http.StatusOK, gin.H{
		"predicted_latency_ms": pred.PredictedExecutionMs,
		"confidence":           pred.ConfidenceScore,
		"recommendations":      recommendations,
		"timestamp":            time.Now().UTC(),
	})
}

// scanTypeValue converts scan type string to numeric feature
func scanTypeValue(st string) float64 {
	switch st {
	case "seq_scan":
		return 0.0
	case "index_scan":
		return 1.0
	case "bitmap_scan":
		return 2.0
	default:
		return 0.0
	}
}

// aggregateTypeValue converts aggregate type string to numeric feature
func aggregateTypeValue(at string) float64 {
	switch at {
	case "none":
		return 0.0
	case "sum", "count":
		return 1.0
	case "group_by":
		return 2.0
	default:
		return 0.0
	}
}

// generateLatencyRecommendations generates optimization recommendations based on predicted latency
func generateLatencyRecommendations(latency float64, req *PredictLatencyRequest) []string {
	var recs []string

	if latency > 500 {
		recs = append(recs, "High latency (>500ms): Consider adding an index on frequently filtered columns")
	}

	if latency > 1000 {
		recs = append(recs, "Very high latency (>1000ms): Consider query optimization or table refactoring")
	}

	if req.JoinCount > 3 {
		recs = append(recs, "Multiple joins detected: Review join order and ensure proper indexing on join keys")
	}

	if req.ScanType == "seq_scan" && req.RowCount > 10000 {
		recs = append(recs, "Sequential scan on large table: Add an index to improve scan performance")
	}

	if req.SubqueryCount > 2 {
		recs = append(recs, "Complex subqueries detected: Consider using CTEs or rewriting with joins for better optimization")
	}

	if len(recs) == 0 {
		recs = append(recs, "Query performance is acceptable for current structure")
	}

	return recs
}
