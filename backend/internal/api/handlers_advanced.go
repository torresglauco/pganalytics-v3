package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
)

// ============================================================================
// PHASE 4.4: ADVANCED QUERY ANALYSIS HANDLERS
// ============================================================================

// ============================================================================
// 4.4.1: QUERY FINGERPRINTING HANDLERS
// ============================================================================

// handleGetQueryFingerprints returns grouped queries by fingerprint
// GET /api/v1/queries/fingerprints?limit=50&min_calls=100
func (s *Server) handleGetQueryFingerprints(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Query fingerprints from database
	fingerprints, err := s.db.GetQueryFingerprints(ctx, limit)
	if err != nil {
		s.logger.Warnf("Failed to get fingerprints: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve fingerprints"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"data":  fingerprints,
		"count": len(fingerprints),
	})
}

// handleGetQueriesByFingerprint returns all individual queries for a specific fingerprint
// GET /api/v1/queries/fingerprints/:fingerprint_hash/queries?limit=50
func (s *Server) handleGetQueriesByFingerprint(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse fingerprint hash from URL
	fingerprintHashStr := c.Param("fingerprint_hash")
	fingerprintHash, err := strconv.ParseInt(fingerprintHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fingerprint_hash format"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Query individual queries for this fingerprint
	queries, err := s.db.GetQueriesByFingerprint(ctx, fingerprintHash, limit)
	if err != nil {
		s.logger.Warnf("Failed to get queries by fingerprint: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve queries"})
		return
	}

	if len(queries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No queries found for this fingerprint"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"fingerprint_hash": fingerprintHash,
		"queries":          queries,
		"count":            len(queries),
	})
}

// ============================================================================
// 4.4.2: EXPLAIN PLAN HANDLERS
// ============================================================================

// handleGetExplainPlan returns the latest EXPLAIN plan for a query
// GET /api/v1/queries/:query_hash/explain
func (s *Server) handleGetExplainPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query hash from URL
	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	// Query EXPLAIN plan from database
	plan, err := s.db.GetExplainPlan(ctx, queryHash)
	if err != nil {
		s.logger.Warnf("Failed to get explain plan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve explain plan"})
		return
	}

	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No EXPLAIN plan found for this query"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"query_hash": queryHash,
		"plan":       plan,
	})
}

// handleGetExplainPlanHistory returns the last N EXPLAIN plans for a query
// GET /api/v1/queries/:query_hash/explain/history?limit=10
func (s *Server) handleGetExplainPlanHistory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query hash from URL
	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Query EXPLAIN plan history from database
	plans, err := s.db.GetExplainPlanHistory(ctx, queryHash, limit)
	if err != nil {
		s.logger.Warnf("Failed to get explain plan history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve explain history"})
		return
	}

	if len(plans) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No EXPLAIN plan history found for this query"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"query_hash": queryHash,
		"plans":      plans,
		"count":      len(plans),
	})
}

// ============================================================================
// 4.4.3: INDEX RECOMMENDATION HANDLERS
// ============================================================================

// handleGetIndexRecommendations returns recommended indexes for a database
// GET /api/v1/databases/:database_name/index-recommendations?limit=20
func (s *Server) handleGetIndexRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get database name from URL
	databaseName := c.Param("database_name")
	if databaseName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_name is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 20
	}

	// Query recommendations from database
	recommendations, err := s.db.GetIndexRecommendations(ctx, databaseName, limit)
	if err != nil {
		s.logger.Warnf("Failed to get index recommendations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recommendations"})
		return
	}

	if len(recommendations) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"database":        databaseName,
			"recommendations": []interface{}{},
			"count":           0,
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"database":        databaseName,
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// handleDismissIndexRecommendation marks a recommendation as dismissed
// POST /api/v1/index-recommendations/:id/dismiss
func (s *Server) handleDismissIndexRecommendation(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Parse recommendation ID from URL
	idStr := c.Param("id")
	recommendationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recommendation ID"})
		return
	}

	// Parse request body
	var req struct {
		Reason *string `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Dismiss the recommendation in database
	err = s.db.DismissIndexRecommendation(ctx, recommendationID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dismiss recommendation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     recommendationID,
		"status": "dismissed",
		"reason": req.Reason,
	})
}

// handleGenerateIndexRecommendations triggers analysis of EXPLAIN plans for index recommendations
// POST /api/v1/databases/:database_name/index-recommendations/generate
func (s *Server) handleGenerateIndexRecommendations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	databaseName := c.Param("database_name")
	if databaseName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_name is required"})
		return
	}

	// Optional: Get collector ID from query parameter
	collectorID := c.Query("collector_id")

	// Generate recommendations from recent EXPLAIN plans
	count, err := s.db.GenerateIndexRecommendations(ctx, databaseName, &collectorID)
	if err != nil {
		s.logger.Warnf("Failed to generate index recommendations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database":                  databaseName,
		"recommendations_generated": count,
		"status":                    "success",
	})
}

// ============================================================================
// 4.4.4: ANOMALY DETECTION HANDLERS
// ============================================================================

// handleGetQueryAnomalies returns detected anomalies for a specific query
// GET /api/v1/queries/:query_hash/anomalies?days=7
func (s *Server) handleGetQueryAnomalies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query hash from URL
	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	// Parse query parameters
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 30 {
		days = 7
	}

	// Query anomalies from database
	anomalies, err := s.db.GetQueryAnomalies(ctx, queryHash, days)
	if err != nil {
		s.logger.Warnf("Failed to get anomalies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve anomalies"})
		return
	}

	if len(anomalies) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"query_hash": queryHash,
			"anomalies":  []interface{}{},
			"count":      0,
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"query_hash": queryHash,
		"anomalies":  anomalies,
		"count":      len(anomalies),
		"days":       days,
	})
}

// handleGetAnomaliesBySeverity returns high-severity anomalies across all queries
// GET /api/v1/anomalies?severity=high&limit=50
func (s *Server) handleGetAnomaliesBySeverity(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get severity filter from query params
	severity := c.DefaultQuery("severity", "high")
	if severity != "low" && severity != "medium" && severity != "high" {
		severity = "high"
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// Query anomalies from database
	anomalies, err := s.db.GetAnomaliesBySeverity(ctx, severity, limit)
	if err != nil {
		s.logger.Warnf("Failed to get anomalies by severity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve anomalies"})
		return
	}

	if len(anomalies) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"severity":  severity,
			"anomalies": []interface{}{},
			"count":     0,
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"severity":  severity,
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

// handleDetectAnomalies triggers the anomaly detection process
// POST /api/v1/anomalies/detect
func (s *Server) handleDetectAnomalies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Execute anomaly detection calculation
	err := s.db.CalculateBaselineAndDetectAnomalies(ctx)
	if err != nil {
		s.logger.Warnf("Failed to detect anomalies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run anomaly detection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Anomaly detection completed",
	})
}

// handleResolveAnomaly marks an anomaly as resolved
// POST /api/v1/anomalies/:id/resolve
func (s *Server) handleResolveAnomaly(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Parse anomaly ID from URL
	idStr := c.Param("id")
	anomalyID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anomaly ID"})
		return
	}

	// Resolve the anomaly
	err = s.db.ResolveAnomaly(ctx, anomalyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve anomaly"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     anomalyID,
		"status": "resolved",
	})
}

// ============================================================================
// 4.4.5: PERFORMANCE SNAPSHOT HANDLERS
// ============================================================================

// handleCreatePerformanceSnapshot creates a new baseline snapshot
// POST /api/v1/snapshots
func (s *Server) handleCreatePerformanceSnapshot(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second) // Longer timeout for snapshot capture
	defer cancel()

	// Parse request body
	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description,omitempty"`
		SnapshotType string `json:"snapshot_type,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate snapshot type
	snapshotType := req.SnapshotType
	if snapshotType == "" {
		snapshotType = "manual"
	}
	if snapshotType != "manual" && snapshotType != "scheduled" && snapshotType != "pre_deploy" && snapshotType != "post_deploy" {
		snapshotType = "manual"
	}

	// Create snapshot
	snapshotID, err := s.db.CreatePerformanceSnapshot(ctx, req.Name, &req.Description, snapshotType, nil)
	if err != nil {
		s.logger.Warnf("Failed to create snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snapshot"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"id":            snapshotID,
		"name":          req.Name,
		"description":   req.Description,
		"snapshot_type": snapshotType,
		"created_at":    time.Now(),
	})
}

// handleGetPerformanceSnapshots returns all performance snapshots
// GET /api/v1/snapshots?limit=20
func (s *Server) handleGetPerformanceSnapshots(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Query snapshots from database
	snapshots, err := s.db.GetPerformanceSnapshots(ctx, limit)
	if err != nil {
		s.logger.Warnf("Failed to get snapshots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve snapshots"})
		return
	}

	if len(snapshots) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"snapshots": []interface{}{},
			"count":     0,
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}

// handleCompareSnapshots compares metrics between two snapshots
// GET /api/v1/queries/comparison?before_snapshot=1&after_snapshot=2&limit=50
func (s *Server) handleCompareSnapshots(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Parse query parameters
	beforeStr := c.Query("before_snapshot")
	afterStr := c.Query("after_snapshot")

	if beforeStr == "" || afterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "before_snapshot and after_snapshot are required"})
		return
	}

	beforeID, err := strconv.ParseInt(beforeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid before_snapshot ID"})
		return
	}

	afterID, err := strconv.ParseInt(afterStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid after_snapshot ID"})
		return
	}

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// Compare snapshots
	comparisons, err := s.db.CompareSnapshots(ctx, beforeID, afterID, limit)
	if err != nil {
		s.logger.Warnf("Failed to compare snapshots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare snapshots"})
		return
	}

	if len(comparisons) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"before_snapshot": beforeID,
			"after_snapshot":  afterID,
			"comparisons":     []interface{}{},
			"count":           0,
		})
		return
	}

	// Calculate summary statistics
	improvedCount := 0
	degradedCount := 0
	unchangedCount := 0

	for _, comp := range comparisons {
		switch comp.ImprovementStatus {
		case "improved":
			improvedCount++
		case "degraded":
			degradedCount++
		case "unchanged":
			unchangedCount++
		}
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"before_snapshot": beforeID,
		"after_snapshot":  afterID,
		"comparisons":     comparisons,
		"count":           len(comparisons),
		"improved_count":  improvedCount,
		"degraded_count":  degradedCount,
		"unchanged_count": unchangedCount,
	})
}

// handleGetSnapshotComparison compares a specific query between two snapshots
// GET /api/v1/queries/:query_hash/comparison?before_snapshot=1&after_snapshot=2
func (s *Server) handleGetSnapshotComparison(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse query hash from URL
	queryHashStr := c.Param("query_hash")
	queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
		return
	}

	// Parse query parameters
	beforeStr := c.Query("before_snapshot")
	afterStr := c.Query("after_snapshot")

	if beforeStr == "" || afterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "before_snapshot and after_snapshot are required"})
		return
	}

	beforeID, err := strconv.ParseInt(beforeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid before_snapshot ID"})
		return
	}

	afterID, err := strconv.ParseInt(afterStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid after_snapshot ID"})
		return
	}

	// Compare snapshots (limit=1 for specific query)
	comparisons, err := s.db.CompareSnapshots(ctx, beforeID, afterID, 1)
	if err != nil {
		s.logger.Warnf("Failed to compare snapshots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compare snapshots"})
		return
	}

	// Find the matching query
	var comparison *models.SnapshotComparison
	for _, comp := range comparisons {
		if comp.QueryHash == queryHash {
			comparison = comp
			break
		}
	}

	if comparison == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query not found in both snapshots"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"query_hash":      queryHash,
		"before_snapshot": beforeID,
		"after_snapshot":  afterID,
		"comparison":      comparison,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// validateDatabaseName validates a database name
func validateDatabaseName(name string) error {
	if name == "" {
		return apperrors.ValidationError("database_name", "database name is required")
	}
	if len(name) > 63 {
		return apperrors.ValidationError("database_name", "database name too long (max 63 characters)")
	}
	return nil
}

// validateSnapshotName validates a snapshot name
func validateSnapshotName(name string) error {
	if name == "" {
		return apperrors.ValidationError("name", "snapshot name is required")
	}
	if len(name) > 255 {
		return apperrors.ValidationError("name", "snapshot name too long (max 255 characters)")
	}
	return nil
}

// ============================================================================
// EXPLAIN PLAN STORAGE HANDLER (Internal)
// ============================================================================

// handleStoreExplainPlan stores EXPLAIN plans from collector
// POST /api/v1/internal/explain-plans (internal endpoint, no auth shown here)
func (s *Server) handleStoreExplainPlan(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Parse request body
	var req struct {
		QueryHash            int64       `json:"query_hash" binding:"required"`
		QueryFingerprintHash *int64      `json:"query_fingerprint_hash,omitempty"`
		CollectedAt          time.Time   `json:"collected_at"`
		PlanJSON             interface{} `json:"plan_json"`
		PlanText             *string     `json:"plan_text,omitempty"`
		RowsExpected         *int64      `json:"rows_expected,omitempty"`
		RowsActual           *int64      `json:"rows_actual,omitempty"`
		PlanDurationMs       *float64    `json:"plan_duration_ms,omitempty"`
		ExecutionDurationMs  *float64    `json:"execution_duration_ms,omitempty"`
		HasSeqScan           bool        `json:"has_seq_scan"`
		HasIndexScan         bool        `json:"has_index_scan"`
		HasBitmapScan        bool        `json:"has_bitmap_scan"`
		HasNestedLoop        bool        `json:"has_nested_loop"`
		TotalBuffersRead     *int64      `json:"total_buffers_read,omitempty"`
		TotalBuffersHit      *int64      `json:"total_buffers_hit,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create model
	plan := &models.ExplainPlan{
		QueryHash:            req.QueryHash,
		QueryFingerprintHash: req.QueryFingerprintHash,
		CollectedAt:          req.CollectedAt,
		PlanJSON:             req.PlanJSON,
		PlanText:             req.PlanText,
		RowsExpected:         req.RowsExpected,
		RowsActual:           req.RowsActual,
		PlanDurationMs:       req.PlanDurationMs,
		ExecutionDurationMs:  req.ExecutionDurationMs,
		HasSeqScan:           req.HasSeqScan,
		HasIndexScan:         req.HasIndexScan,
		HasBitmapScan:        req.HasBitmapScan,
		HasNestedLoop:        req.HasNestedLoop,
		TotalBuffersRead:     req.TotalBuffersRead,
		TotalBuffersHit:      req.TotalBuffersHit,
	}

	// Store in database
	err := s.db.StoreExplainPlan(ctx, plan)
	if err != nil {
		s.logger.Warnf("Failed to store explain plan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store explain plan"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"query_hash": req.QueryHash,
		"status":     "stored",
	})
}

// handleGetLatestExplainPlans returns recently captured EXPLAIN plans
// GET /api/v1/collectors/:collector_id/explain-plans?limit=10
func (s *Server) handleGetLatestExplainPlans(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get collector ID from URL
	collectorID := c.Param("collector_id")
	if collectorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collector_id is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	// Query EXPLAIN plans from database
	plans, err := s.db.GetLatestExplainPlans(ctx, collectorID, limit)
	if err != nil {
		s.logger.Warnf("Failed to get latest explain plans: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve explain plans"})
		return
	}

	if len(plans) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"collector_id": collectorID,
			"plans":        []interface{}{},
			"count":        0,
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"collector_id": collectorID,
		"plans":        plans,
		"count":        len(plans),
	})
}
