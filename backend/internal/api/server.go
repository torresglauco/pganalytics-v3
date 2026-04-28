package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/torresglauco/pganalytics-v3/backend/internal/audit"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"github.com/torresglauco/pganalytics-v3/backend/internal/crypto"
	"github.com/torresglauco/pganalytics-v3/backend/internal/ml"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/log_analysis"
	"github.com/torresglauco/pganalytics-v3/backend/internal/session"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/internal/timescale"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/handlers"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	config            *config.Config
	logger            *zap.Logger
	postgres          *storage.PostgresDB
	timescale         *timescale.TimescaleDB
	authService       *auth.AuthService
	jwtManager        *auth.JWTManager
	mlClient          *ml.Client
	featureExtractor  ml.IFeatureExtractor
	cacheManager      *cache.Manager
	rateLimiter       *RateLimiter
	secretManager     *crypto.SecretManager
	sessionManager    session.ISessionManager
	mfaManager        *auth.MFAManager
	auditLogger       *audit.AuditLogger
	wsManager         *services.ConnectionManager
	conditionHandler  *handlers.ConditionHandler
	silenceHandler    *handlers.SilenceHandler
	escalationHandler *handlers.EscalationHandler
	logCollector      *log_analysis.LogCollector
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
	rateLimiter := NewRateLimiter(1000) // Increased to handle high-volume metric pushes from collectors

	// Initialize services and handlers
	conditionValidator := services.NewConditionValidator()
	conditionHandler := handlers.NewConditionHandler(conditionValidator)

	// TODO: Initialize SilenceService with SilenceDB implementation
	var silenceHandler *handlers.SilenceHandler
	// silenceHandler = handlers.NewSilenceHandler(silenceService)

	// TODO: Initialize EscalationService with EscalationDB implementation and Notifier
	var escalationHandler *handlers.EscalationHandler
	// escalationHandler = handlers.NewEscalationHandler(escalationService)

	// Initialize session manager
	sessionManager := session.NewSessionManager(nil) // Redis client to be configured

	// Initialize log collector for log analysis and streaming
	var logCollectorDB *sql.DB
	if postgres != nil {
		logCollectorDB = postgres.GetDB()
	}
	logCollector := log_analysis.NewLogCollector(logCollectorDB)

	return &Server{
		config:            cfg,
		logger:            logger,
		postgres:          postgres,
		timescale:         timescale,
		authService:       authService,
		jwtManager:        jwtManager,
		mlClient:          mlClient,
		featureExtractor:  featureExtractor,
		cacheManager:      nil, // Set via SetCacheManager
		rateLimiter:       rateLimiter,
		secretManager:     secretManager,
		sessionManager:    sessionManager,
		wsManager:         services.NewConnectionManager(logger),
		conditionHandler:  conditionHandler,
		silenceHandler:    silenceHandler,
		escalationHandler: escalationHandler,
		logCollector:      logCollector,
	}
}

// SetCacheManager sets the cache manager for the server
func (s *Server) SetCacheManager(cm *cache.Manager) {
	s.cacheManager = cm
}

// SetSessionManager sets the session manager for the server
func (s *Server) SetSessionManager(sm session.ISessionManager) {
	s.sessionManager = sm
}

// ValidateAuthConfiguration validates all enabled authentication methods at startup
// This ensures that invalid or missing configurations fail fast before the server accepts requests
func (s *Server) ValidateAuthConfiguration() error {
	if s.config.LDAPEnabled {
		if err := s.validateLDAPConfiguration(); err != nil {
			return err
		}
	}

	if s.config.OAuthEnabled {
		if err := s.validateOAuthConfiguration(); err != nil {
			return err
		}
	}

	if s.config.SAMLEnabled {
		if err := s.validateSAMLConfiguration(); err != nil {
			return err
		}
	}

	return nil
}

// validateLDAPConfiguration validates LDAP configuration
func (s *Server) validateLDAPConfiguration() error {
	if s.config.LDAPServerURL == "" {
		return fmt.Errorf("LDAP enabled but server URL not configured (set LDAP_SERVER_URL)")
	}

	// Validate JSON parsing of LDAP group mappings
	var ldapGroupMapping map[string]string
	if err := json.Unmarshal([]byte(s.config.LDAPGroupToRoleJSON), &ldapGroupMapping); err != nil {
		return fmt.Errorf("invalid LDAP group mapping JSON (LDAP_GROUP_TO_ROLE_MAPPING): %w", err)
	}

	if len(ldapGroupMapping) == 0 {
		return fmt.Errorf("LDAP group mappings are empty (set LDAP_GROUP_TO_ROLE_MAPPING with valid JSON mappings)")
	}

	s.logger.Info("LDAP configuration validated",
		zap.String("server_url", s.config.LDAPServerURL),
		zap.Int("group_mappings", len(ldapGroupMapping)))

	return nil
}

// validateOAuthConfiguration validates OAuth provider configuration
func (s *Server) validateOAuthConfiguration() error {
	var oauthConfigs []auth.OAuthProviderConfig
	if err := json.Unmarshal([]byte(s.config.OAuthProvidersJSON), &oauthConfigs); err != nil {
		return fmt.Errorf("invalid OAuth providers JSON (OAUTH_PROVIDERS): %w", err)
	}

	if len(oauthConfigs) == 0 {
		return fmt.Errorf("OAuth enabled but no providers configured (set OAUTH_PROVIDERS with valid JSON array)")
	}

	// Validate each provider's required fields
	for _, cfg := range oauthConfigs {
		if cfg.Name == "" {
			return fmt.Errorf("OAuth provider missing 'name' field")
		}

		if cfg.ClientID == "" {
			return fmt.Errorf("OAuth provider '%s' missing ClientID (provider.client_id)", cfg.Name)
		}

		if cfg.ClientSecret == "" {
			return fmt.Errorf("OAuth provider '%s' missing ClientSecret (provider.client_secret)", cfg.Name)
		}

		// For custom OIDC providers, validate URLs
		if cfg.Name == "custom" || cfg.Name == "oidc" {
			if cfg.AuthURL == "" {
				return fmt.Errorf("OAuth custom provider '%s' missing auth_url", cfg.Name)
			}
			if cfg.TokenURL == "" {
				return fmt.Errorf("OAuth custom provider '%s' missing token_url", cfg.Name)
			}
		}
	}

	s.logger.Info("OAuth configuration validated",
		zap.Int("providers", len(oauthConfigs)),
		zap.Strings("provider_names", getOAuthProviderNames(oauthConfigs)))

	return nil
}

// validateSAMLConfiguration validates SAML configuration
func (s *Server) validateSAMLConfiguration() error {
	if s.config.SAMLCertPath == "" {
		return fmt.Errorf("SAML enabled but certificate path not configured (set SAML_CERT_PATH)")
	}

	if s.config.SAMLKeyPath == "" {
		return fmt.Errorf("SAML enabled but key path not configured (set SAML_KEY_PATH)")
	}

	// Verify certificate file exists
	if _, err := os.Stat(s.config.SAMLCertPath); err != nil {
		return fmt.Errorf("SAML certificate file not found at '%s' (SAML_CERT_PATH): %w", s.config.SAMLCertPath, err)
	}

	// Verify key file exists
	if _, err := os.Stat(s.config.SAMLKeyPath); err != nil {
		return fmt.Errorf("SAML key file not found at '%s' (SAML_KEY_PATH): %w", s.config.SAMLKeyPath, err)
	}

	if s.config.SAMLIDPMetadataURL == "" {
		return fmt.Errorf("SAML enabled but IdP metadata URL not configured (set SAML_IDP_METADATA_URL)")
	}

	if s.config.SAMLEntityID == "" {
		return fmt.Errorf("SAML enabled but entity ID not configured (set SAML_ENTITY_ID)")
	}

	s.logger.Info("SAML configuration validated",
		zap.String("cert_path", s.config.SAMLCertPath),
		zap.String("key_path", s.config.SAMLKeyPath),
		zap.String("idp_metadata_url", s.config.SAMLIDPMetadataURL),
		zap.String("entity_id", s.config.SAMLEntityID))

	return nil
}

// getOAuthProviderNames returns the names of OAuth providers for logging
func getOAuthProviderNames(configs []auth.OAuthProviderConfig) []string {
	names := make([]string, len(configs))
	for i, cfg := range configs {
		names[i] = cfg.Name
	}
	return names
}

// RegisterRoutes registers all API routes
func (s *Server) RegisterRoutes(router *gin.Engine) {
	// Apply global middleware
	router.Use(s.RequestIDMiddleware())
	router.Use(s.SecurityHeadersMiddleware())

	// Health check (no auth required)
	router.GET("/api/v1/health", s.handleHealth)
	router.GET("/version", s.handleVersion)

	// WebSocket route (JWT auth required, handled in handler)
	router.GET("/api/v1/ws", s.handleWebSocket)

	// API v1 routes
	api := router.Group("/api/v1")
	api.Use(s.RateLimitMiddleware())
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			// Public endpoints (no auth required)
			auth.POST("/login", s.handleLogin)
			auth.POST("/logout", s.handleLogout)
			auth.POST("/refresh", s.handleRefreshToken)
			auth.POST("/setup", s.handleSetupFirstUser) // Create initial admin user (no auth required)

			// Protected endpoints (auth required)
			auth.GET("/me", s.AuthMiddleware(), s.handleGetCurrentUser)
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

			// Token refresh (collector auth required)
			collectors.POST("/refresh-token", s.CollectorAuthMiddleware(), s.handleRefreshCollectorToken)

			// Protected routes
			collectors.GET("", s.AuthMiddleware(), s.handleListCollectors)
			collectors.GET("/:id", s.AuthMiddleware(), s.handleGetCollector)
			collectors.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteCollector)

			// Query Statistics routes
			collectors.GET("/:id/queries/slow", s.AuthMiddleware(), s.handleGetSlowQueries)
			collectors.GET("/:id/queries/frequent", s.AuthMiddleware(), s.handleGetFrequentQueries)

			// ================================================================
			// Metrics Collection Routes (Phase 1 & 2)
			// ================================================================
			collectors.GET("/:id/schema", s.AuthMiddleware(), s.handleGetSchemaMetrics)
			collectors.GET("/:id/locks", s.AuthMiddleware(), s.handleGetLockMetrics)
			collectors.GET("/:id/bloat", s.AuthMiddleware(), s.handleGetBloatMetrics)
			collectors.GET("/:id/cache-hits", s.AuthMiddleware(), s.handleGetCacheMetrics)
			collectors.GET("/:id/connections", s.AuthMiddleware(), s.handleGetConnectionMetrics)
			collectors.GET("/:id/extensions", s.AuthMiddleware(), s.handleGetExtensionMetrics)
		}

		// Registration Secrets routes (admin only)
		secrets := api.Group("/registration-secrets")
		secrets.Use(s.AuthMiddleware())
		{
			secrets.POST("", s.handleCreateRegistrationSecret)
			secrets.GET("", s.handleListRegistrationSecrets)
			secrets.GET("/:id", s.handleGetRegistrationSecret)
			secrets.PUT("/:id", s.handleUpdateRegistrationSecret)
			secrets.DELETE("/:id", s.handleDeleteRegistrationSecret)
		}

		// Metrics routes
		metrics := api.Group("/metrics")
		{
			// High-volume endpoint - requires collector authentication
			metrics.POST("/push", s.CollectorAuthMiddleware(), s.handleMetricsPush)
			// Cache metrics (protected)
			metrics.GET("/cache", s.AuthMiddleware(), s.handleCacheMetrics)
			// General metrics endpoints for frontend dashboard
			metrics.GET("", s.AuthMiddleware(), s.handleGetMetrics)
			metrics.GET("/error-trend", s.AuthMiddleware(), s.handleGetErrorTrend)
			metrics.GET("/log-distribution", s.AuthMiddleware(), s.handleGetLogDistribution)
		}

		// Log Ingest routes
		logs := api.Group("/logs")
		{
			// High-volume endpoint for log ingestion - requires API token auth
			logs.POST("/ingest", s.handleIngestLogs)
			// Frontend log viewer endpoints
			logs.GET("", s.AuthMiddleware(), s.handleGetLogs)
			logs.GET("/:logId", s.AuthMiddleware(), s.handleGetLogDetails)
			// Log analysis endpoints (collector logs)
			logs.GET("/collector/:collector_id", s.AuthMiddleware(), s.handleGetCollectorLogs)
			// WebSocket endpoint for streaming logs in real-time
			logs.GET("/stream/:collector_id", s.handleLogStream)
		}

		// ========================================================================
		// PHASE 4.6: ALERT RULES, SILENCES, AND ESCALATIONS ROUTES
		// ========================================================================

		// Alert Rule Validation routes
		alertRules := api.Group("/alert-rules")
		{
			alertRules.POST("/validate", s.AuthMiddleware(), s.handleValidateAlertCondition)
		}

		// Silences management routes
		silences := api.Group("/silences")
		{
			silences.GET("", s.AuthMiddleware(), s.handleListActiveSilences)
			silences.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteSilence)
		}

		// Escalation Policies routes
		escalationPolicies := api.Group("/escalation-policies")
		{
			escalationPolicies.POST("", s.AuthMiddleware(), s.handleCreateEscalationPolicy)
			escalationPolicies.GET("/:policy_id", s.AuthMiddleware(), s.handleGetEscalationPolicy)
			escalationPolicies.PUT("/:id", s.AuthMiddleware(), s.handleUpdateEscalationPolicy)
		}

		// Alert routes merged below to avoid route conflicts

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

		// Notification Channels routes
		channels := api.Group("/channels")
		{
			channels.GET("", s.AuthMiddleware(), s.handleListChannels)
			channels.POST("", s.AuthMiddleware(), s.handleCreateChannel)
			channels.PUT("/:id", s.AuthMiddleware(), s.handleUpdateChannel)
			channels.DELETE("/:id", s.AuthMiddleware(), s.handleDeleteChannel)
			channels.POST("/:id/test", s.AuthMiddleware(), s.handleTestChannel)
		}

		// Alerts routes
		alerts := api.Group("/alerts")
		{
			alerts.GET("", s.AuthMiddleware(), s.handleListAlerts)
			alerts.GET("/:id", s.AuthMiddleware(), s.handleGetAlert)
			alerts.POST("/:id/acknowledge", s.AuthMiddleware(), s.handleAcknowledgeAlert)
			alerts.POST("/:id/silence", s.AuthMiddleware(), s.handleCreateSilence)
			alerts.POST("/:id/acknowledge-escalation", s.AuthMiddleware(), s.handleAcknowledgeAlertEscalation)
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

		// Query Performance routes
		performanceRoutes := api.Group("/queries")
		{
			performanceRoutes.GET("/:query_hash/performance", s.AuthMiddleware(), s.handleGetQueryPerformance)
		}

		// Query Performance Database routes (per-database endpoint)
		queryPerformanceRoutes := api.Group("/query-performance")
		{
			queryPerformanceRoutes.GET("/database/:database_id", s.AuthMiddleware(), s.handleGetDatabaseQueryPerformance)
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

		// Index Advisor routes (new endpoints for index analysis)
		indexAdvisor := api.Group("/index-advisor")
		{
			indexAdvisor.GET("/database/:database_id/recommendations", s.AuthMiddleware(), s.handleGetIndexAdvisorRecommendations)
			indexAdvisor.POST("/recommendation/:recommendation_id/create", s.AuthMiddleware(), s.handleCreateIndexFromRecommendation)
			indexAdvisor.GET("/database/:database_id/unused", s.AuthMiddleware(), s.handleGetUnusedIndexes)
		}

		// VACUUM Advisor routes (new endpoints for VACUUM recommendations)
		vacuumAdvisor := api.Group("/vacuum-advisor")
		{
			vacuumAdvisor.GET("/database/:database_id/recommendations", s.AuthMiddleware(), s.handleGetVacuumRecommendations)
			vacuumAdvisor.GET("/database/:database_id/table/:table_name", s.AuthMiddleware(), s.handleGetVacuumTableRecommendation)
			vacuumAdvisor.GET("/database/:database_id/autovacuum-config", s.AuthMiddleware(), s.handleGetAutovacuumConfig)
			vacuumAdvisor.POST("/recommendation/:recommendation_id/execute", s.AuthMiddleware(), s.handleExecuteVacuum)
			vacuumAdvisor.GET("/database/:database_id/tune-suggestions", s.AuthMiddleware(), s.handleGetVacuumTuningSuggestions)
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
			ml.POST("/predict-latency", s.AuthMiddleware(), s.handlePredictQueryLatency)

			// Pattern detection (requires auth)
			ml.POST("/patterns/detect", s.AuthMiddleware(), s.handleMLDetectPatterns)

			// Feature extraction (requires auth, for debugging)
			ml.GET("/features/:query_hash", s.AuthMiddleware(), s.handleMLGetFeatures)
		}
	}

	s.logger.Info("API routes registered")
}

// handleWebSocket is a Gin wrapper for the WebSocket handler
func (s *Server) handleWebSocket(c *gin.Context) {
	handler := WebSocketHandler(s.wsManager, s.jwtManager)
	handler(c.Writer, c.Request)
}

// handleIngestLogs is a Gin wrapper for the log ingest handler
func (s *Server) handleIngestLogs(c *gin.Context) {
	handler := handlers.IngestLogs(s.postgres, s.wsManager)
	handler(c.Writer, c.Request)
}

// ========================================================================
// PHASE 4.6: ALERT CONDITIONS, SILENCES, AND ESCALATIONS HANDLERS
// ========================================================================

// handleValidateAlertCondition is a Gin wrapper for condition validation
func (s *Server) handleValidateAlertCondition(c *gin.Context) {
	if s.conditionHandler == nil {
		c.JSON(500, gin.H{"error": "Condition handler not initialized"})
		return
	}
	s.conditionHandler.ValidateCondition(c.Writer, c.Request)
}

// handleCreateSilence is a Gin wrapper for creating a silence
func (s *Server) handleCreateSilence(c *gin.Context) {
	if s.silenceHandler == nil {
		c.JSON(500, gin.H{"error": "Silence handler not initialized"})
		return
	}
	s.silenceHandler.CreateSilence(c.Writer, c.Request)
}

// handleListActiveSilences is a Gin wrapper for listing active silences
func (s *Server) handleListActiveSilences(c *gin.Context) {
	if s.silenceHandler == nil {
		c.JSON(500, gin.H{"error": "Silence handler not initialized"})
		return
	}
	s.silenceHandler.ListActiveSilences(c.Writer, c.Request)
}

// handleDeleteSilence is a Gin wrapper for deleting a silence
func (s *Server) handleDeleteSilence(c *gin.Context) {
	if s.silenceHandler == nil {
		c.JSON(500, gin.H{"error": "Silence handler not initialized"})
		return
	}
	s.silenceHandler.DeleteSilence(c.Writer, c.Request)
}

// handleCreateEscalationPolicy is a Gin wrapper for creating an escalation policy
func (s *Server) handleCreateEscalationPolicy(c *gin.Context) {
	if s.escalationHandler == nil {
		c.JSON(500, gin.H{"error": "Escalation handler not initialized"})
		return
	}
	s.escalationHandler.CreatePolicy(c.Writer, c.Request)
}

// handleGetEscalationPolicy is a Gin wrapper for retrieving an escalation policy
func (s *Server) handleGetEscalationPolicy(c *gin.Context) {
	if s.escalationHandler == nil {
		c.JSON(500, gin.H{"error": "Escalation handler not initialized"})
		return
	}
	s.escalationHandler.GetPolicy(c.Writer, c.Request)
}

// handleUpdateEscalationPolicy is a Gin wrapper for updating an escalation policy
func (s *Server) handleUpdateEscalationPolicy(c *gin.Context) {
	if s.escalationHandler == nil {
		c.JSON(500, gin.H{"error": "Escalation handler not initialized"})
		return
	}
	s.escalationHandler.UpdatePolicy(c.Writer, c.Request)
}

// handleAcknowledgeAlertEscalation is a Gin wrapper for acknowledging an alert
func (s *Server) handleAcknowledgeAlertEscalation(c *gin.Context) {
	if s.escalationHandler == nil {
		c.JSON(500, gin.H{"error": "Escalation handler not initialized"})
		return
	}
	s.escalationHandler.AcknowledgeAlert(c.Writer, c.Request)
}
