package api

import (
	"github.com/dextra/pganalytics-v3/backend/internal/auth"
	"github.com/dextra/pganalytics-v3/backend/internal/config"
	"github.com/dextra/pganalytics-v3/backend/internal/storage"
	"github.com/dextra/pganalytics-v3/backend/internal/timescale"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	config      *config.Config
	logger      *zap.Logger
	postgres    *storage.PostgresDB
	timescale   *timescale.TimescaleDB
	authService *auth.AuthService
	jwtManager  *auth.JWTManager
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	postgres *storage.PostgresDB,
	timescale *timescale.TimescaleDB,
	authService *auth.AuthService,
	jwtManager *auth.JWTManager,
) *Server {
	return &Server{
		config:      cfg,
		logger:      logger,
		postgres:    postgres,
		timescale:   timescale,
		authService: authService,
		jwtManager:  jwtManager,
	}
}

// RegisterRoutes registers all API routes
func (s *Server) RegisterRoutes(router *gin.Engine) {
	// Health check (no auth required)
	router.GET("/api/v1/health", s.handleHealth)
	router.GET("/version", s.handleVersion)

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", s.handleLogin)
			auth.POST("/logout", s.handleLogout)
			auth.POST("/refresh", s.handleRefreshToken)
		}

		// Collector routes
		collectors := api.Group("/collectors")
		{
			// Registration (no auth required)
			collectors.POST("/register", s.handleCollectorRegister)

			// Protected routes
			collectors.GET("", s.AuthMiddleware(), s.handleListCollectors)
			collectors.GET("/:id", s.AuthMiddleware(), s.handleGetCollector)
			collectors.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteCollector)
		}

		// Metrics routes
		metrics := api.Group("/metrics")
		{
			// High-volume endpoint (mTLS + JWT)
			metrics.POST("/push", s.MTLSMiddleware(), s.AuthMiddleware(), s.handleMetricsPush)
		}

		// Internal analysis routes (collector -> backend for analyzed data like EXPLAIN plans)
		internal := api.Group("/internal")
		{
			internal.POST("/explain-plans", s.MTLSMiddleware(), s.AuthMiddleware(), s.handleStoreExplainPlan)
		}

		// Configuration routes
		config := api.Group("/config")
		{
			config.GET("/:collector_id", s.MTLSMiddleware(), s.AuthMiddleware(), s.handleGetConfig)
			config.PUT("/:collector_id", s.AuthMiddleware(), s.handleUpdateConfig)
		}

		// Servers routes
		servers := api.Group("/servers")
		{
			servers.GET("", s.AuthMiddleware(), s.handleListServers)
			servers.GET("/:id", s.AuthMiddleware(), s.handleGetServer)
			servers.GET("/:id/metrics", s.AuthMiddleware(), s.handleGetServerMetrics)
		}

		// Alerts routes
		alerts := api.Group("/alerts")
		{
			alerts.GET("", s.AuthMiddleware(), s.handleListAlerts)
			alerts.GET("/:id", s.AuthMiddleware(), s.handleGetAlert)
			alerts.POST("/:id/acknowledge", s.AuthMiddleware(), s.handleAcknowledgeAlert)
		}

		// Query Statistics routes
		queries := api.Group("/collectors/:collector_id/queries")
		{
			queries.GET("/slow", s.AuthMiddleware(), s.handleGetSlowQueries)
			queries.GET("/frequent", s.AuthMiddleware(), s.handleGetFrequentQueries)
		}

		// Query timeline routes
		timeline := api.Group("/queries")
		{
			timeline.GET("/:query_hash/timeline", s.AuthMiddleware(), s.handleGetQueryTimeline)
		}

		// ========================================================================
		// PHASE 4.4: ADVANCED QUERY ANALYSIS ROUTES
		// ========================================================================

		// Query Fingerprinting routes
		fingerprints := api.Group("/queries/fingerprints")
		{
			fingerprints.GET("", s.AuthMiddleware(), s.handleGetQueryFingerprints)
			fingerprints.GET("/:fingerprint_hash/queries", s.AuthMiddleware(), s.handleGetQueriesByFingerprint)
		}

		// EXPLAIN Plan routes
		explainRoutes := api.Group("/queries")
		{
			explainRoutes.GET("/:query_hash/explain", s.AuthMiddleware(), s.handleGetExplainPlan)
			explainRoutes.GET("/:query_hash/explain/history", s.AuthMiddleware(), s.handleGetExplainPlanHistory)
		}

		// Index Recommendations routes
		indexRecommendations := api.Group("/databases/:database_name/index-recommendations")
		{
			indexRecommendations.GET("", s.AuthMiddleware(), s.handleGetIndexRecommendations)
			indexRecommendations.POST("/generate", s.AuthMiddleware(), s.handleGenerateIndexRecommendations)
		}

		recommendations := api.Group("/index-recommendations")
		{
			recommendations.POST("/:id/dismiss", s.AuthMiddleware(), s.handleDismissIndexRecommendation)
		}

		// Anomaly Detection routes
		anomalies := api.Group("/queries")
		{
			anomalies.GET("/:query_hash/anomalies", s.AuthMiddleware(), s.handleGetQueryAnomalies)
		}

		anomaliesBySeverity := api.Group("/anomalies")
		{
			anomaliesBySeverity.GET("", s.AuthMiddleware(), s.handleGetAnomaliesBySeverity)
			anomaliesBySeverity.POST("/detect", s.AuthMiddleware(), s.handleDetectAnomalies)
			anomaliesBySeverity.POST("/:id/resolve", s.AuthMiddleware(), s.handleResolveAnomaly)
		}

		// Performance Snapshots routes
		snapshots := api.Group("/snapshots")
		{
			snapshots.POST("", s.AuthMiddleware(), s.handleCreatePerformanceSnapshot)
			snapshots.GET("", s.AuthMiddleware(), s.handleGetPerformanceSnapshots)
		}

		snapshotComparison := api.Group("/queries")
		{
			snapshotComparison.GET("/comparison", s.AuthMiddleware(), s.handleCompareSnapshots)
			snapshotComparison.GET("/:query_hash/comparison", s.AuthMiddleware(), s.handleGetSnapshotComparison)
		}
	}

	s.logger.Info("API routes registered")
}
