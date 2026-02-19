package api

import (
	"github.com/dextra/pganalytics-v3/backend/internal/config"
	"github.com/dextra/pganalytics-v3/backend/internal/storage"
	"github.com/dextra/pganalytics-v3/backend/internal/timescale"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the API server
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	postgres   *storage.PostgresDB
	timescale  *timescale.TimescaleDB
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	postgres *storage.PostgresDB,
	timescale *timescale.TimescaleDB,
) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		postgres:  postgres,
		timescale: timescale,
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
	}

	s.logger.Info("API routes registered")
}
