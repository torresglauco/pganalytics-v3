package api

import (
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"github.com/torresglauco/pganalytics-v3/backend/internal/crypto"
	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/internal/timescale"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	config           *config.Config
	logger           *zap.Logger
	postgres         *storage.PostgresDB
	timescale        *timescale.TimescaleDB
	authService      *auth.AuthService
	jwtManager       *auth.JWTManager
	mlClient         *ml.Client
	featureExtractor ml.IFeatureExtractor
	cacheManager     *cache.Manager
	rateLimiter      *RateLimiter
	secretManager    *crypto.SecretManager
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	postgres *storage.PostgresDB,
	timescale *timescale.TimescaleDB,
	authService *auth.AuthService,
	jwtManager *auth.JWTManager,
	secretManager *crypto.SecretManager,
) *Server {
	// Initialize ML client if enabled
	var mlClient *ml.Client
	var featureExtractor ml.IFeatureExtractor

	if cfg.MLServiceEnabled {
		mlClient = ml.NewClient(cfg.MLServiceURL, cfg.MLServiceTimeout, logger)
		baseExtractor := ml.NewFeatureExtractor(postgres, logger)
		// Wrap with caching if cache is enabled
		if cfg.CacheEnabled {
			featureExtractor = ml.NewCachedFeatureExtractor(
				baseExtractor,
				cfg.FeatureCacheTTL,
				cfg.CacheMaxSize,
				logger,
			)
		} else {
			featureExtractor = baseExtractor
		}
	}

	// Initialize rate limiter (100 req/min per user, 1000 req/min per collector)
	rateLimiter := NewRateLimiter(100) // Will be increased for collectors in middleware

	return &Server{
		config:           cfg,
		logger:           logger,
		postgres:         postgres,
		timescale:        timescale,
		authService:      authService,
		jwtManager:       jwtManager,
		mlClient:         mlClient,
		featureExtractor: featureExtractor,
		cacheManager:     nil, // Set via SetCacheManager
		rateLimiter:      rateLimiter,
		secretManager:    secretManager,
	}
}

// SetCacheManager sets the cache manager for the server
func (s *Server) SetCacheManager(cm *cache.Manager) {
	s.cacheManager = cm
}

// RegisterRoutes registers all API routes
func (s *Server) RegisterRoutes(router *gin.Engine) {
	// Apply global middleware
	router.Use(s.SecurityHeadersMiddleware())

	// Health check (no auth required)
	router.GET("/api/v1/health", s.handleHealth)
	router.GET("/version", s.handleVersion)

	// API v1 routes
	api := router.Group("/api/v1")
	api.Use(s.RateLimitMiddleware())
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", s.handleLogin)
			auth.POST("/logout", s.handleLogout)
			auth.POST("/refresh", s.handleRefreshToken)
			auth.POST("/change-password", s.AuthMiddleware(), s.handleChangePassword)
		}

		// User Management routes (admin only)
		users := api.Group("/users")
		users.Use(s.AuthMiddleware())
		{
			users.POST("", s.handleCreateUser)
			users.GET("", s.handleListUsers)
			users.PUT("/:id", s.handleUpdateUser)
			users.DELETE("/:id", s.handleDeleteUser)
			users.POST("/:id/reset-password", s.handleResetUserPassword)
		}

		// Managed Instance Management routes (admin only)
		managedInstances := api.Group("/managed-instances")
		managedInstances.Use(s.AuthMiddleware())
		{
			// Exact path routes first
			managedInstances.POST("/test-connection-direct", s.handleTestManagedInstanceConnectionDirect)
			// Then CRUD routes
			managedInstances.POST("", s.handleCreateManagedInstance)
			managedInstances.GET("", s.handleListManagedInstances)
			managedInstances.GET("/:id", s.handleGetManagedInstance)
			managedInstances.PUT("/:id", s.handleUpdateManagedInstance)
			managedInstances.DELETE("/:id", s.handleDeleteManagedInstance)
			managedInstances.POST("/:id/test-connection", s.handleTestManagedInstanceConnection)
		}

		// Collector routes will be defined below

		collectors := api.Group("/collectors")
		{
			// Registration (no auth required)
			collectors.POST("/register", s.handleCollectorRegister)

			// Protected routes
			collectors.GET("", s.AuthMiddleware(), s.handleListCollectors)
			collectors.GET("/:id", s.AuthMiddleware(), s.handleGetCollector)
			collectors.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteCollector)

			// Query Statistics routes
			collectors.GET("/:id/queries/slow", s.AuthMiddleware(), s.handleGetSlowQueries)
			collectors.GET("/:id/queries/frequent", s.AuthMiddleware(), s.handleGetFrequentQueries)
		}

		// Metrics routes
		metrics := api.Group("/metrics")
		{
			// High-volume endpoint - requires collector authentication
			metrics.POST("/push", s.CollectorAuthMiddleware(), s.handleMetricsPush)
			// Cache metrics (protected)
			metrics.GET("/cache", s.AuthMiddleware(), s.handleCacheMetrics)
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

		// ========================================================================
		// PHASE 4.5: ML-BASED QUERY OPTIMIZATION SUGGESTIONS ROUTES
		// ========================================================================

		// Workload Pattern Detection routes
		patterns := api.Group("/workload-patterns")
		{
			patterns.POST("/analyze", s.AuthMiddleware(), s.handleDetectWorkloadPatterns)
			patterns.GET("", s.AuthMiddleware(), s.handleGetWorkloadPatterns)
		}

		// Query Rewrite Suggestions routes
		rewriteRoutes := api.Group("/queries")
		{
			rewriteRoutes.POST("/:query_hash/rewrite-suggestions/generate", s.AuthMiddleware(), s.handleGenerateRewriteSuggestions)
			rewriteRoutes.GET("/:query_hash/rewrite-suggestions", s.AuthMiddleware(), s.handleGetRewriteSuggestions)
			rewriteRoutes.POST("/:query_hash/parameter-optimization/generate", s.AuthMiddleware(), s.handleOptimizeParameters)
			rewriteRoutes.GET("/:query_hash/parameter-optimization", s.AuthMiddleware(), s.handleGetParameterOptimization)
			rewriteRoutes.POST("/:query_hash/predict-performance", s.AuthMiddleware(), s.handlePredictQueryPerformance)
		}

		// Recommendations Aggregation routes
		recommendationsRoutes := api.Group("/recommendations")
		{
			recommendationsRoutes.POST("/aggregate", s.AuthMiddleware(), s.handleAggregateRecommendations)
		}

		// Optimization Recommendations routes
		optimization := api.Group("/optimization-recommendations")
		{
			optimization.GET("", s.AuthMiddleware(), s.handleGetOptimizationRecommendations)
			optimization.POST("/:recommendation_id/implement", s.AuthMiddleware(), s.handleImplementRecommendation)
		}

		// Optimization Results routes
		results := api.Group("/optimization-results")
		{
			results.GET("", s.AuthMiddleware(), s.handleGetOptimizationResults)
		}

		// ========================================================================
		// PHASE 4.5.8: ML SERVICE INTEGRATION ROUTES
		// ========================================================================

		// ML Service integration routes (no auth required for health, auth required for operations)
		ml := api.Group("/ml")
		{
			// Health and status endpoints (no auth)
			ml.GET("/health", s.handleMLHealth)
			ml.GET("/circuit-breaker", s.handleMLCircuitBreakerStatus)

			// Model training (requires auth)
			ml.POST("/train", s.AuthMiddleware(), s.handleMLTrain)
			ml.GET("/train/:job_id", s.AuthMiddleware(), s.handleMLTrainingStatus)

			// Prediction and validation (requires auth)
			ml.POST("/predict", s.AuthMiddleware(), s.handleMLPredict)
			ml.POST("/validate", s.AuthMiddleware(), s.handleMLValidate)

			// Pattern detection (requires auth)
			ml.POST("/patterns/detect", s.AuthMiddleware(), s.handleMLDetectPatterns)

			// Feature extraction (requires auth, for debugging)
			ml.GET("/features/:query_hash", s.AuthMiddleware(), s.handleMLGetFeatures)
		}
	}

	s.logger.Info("API routes registered")
}
