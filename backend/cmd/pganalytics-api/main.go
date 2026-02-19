package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title pgAnalytics v3.0 API
// @version 1.0.0
// @description Modern PostgreSQL monitoring platform API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.pganalytics.local/support
// @contact.email support@pganalytics.local

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL environment variable not set")
	}

	timescaleURL := os.Getenv("TIMESCALE_URL")
	if timescaleURL == "" {
		logger.Fatal("TIMESCALE_URL environment variable not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Warn("JWT_SECRET not set, using default (insecure for production)")
		jwtSecret = "default-insecure-secret"
	}

	// Log startup
	logger.Info("pgAnalytics v3.0 API Starting",
		zap.String("port", port),
		zap.String("environment", getEnv("ENVIRONMENT", "development")),
	)

	// Initialize Gin
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "3.0.0",
			"message": "pgAnalytics API is running",
		})
	})

	// Placeholder endpoints (to be implemented in Phase 2)
	v1 := router.Group("/api/v1")
	{
		// Collectors
		v1.POST("/collectors/register", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "not implemented yet"})
		})
		v1.GET("/collectors", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "not implemented yet"})
		})

		// Metrics
		v1.POST("/metrics/push", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "not implemented yet"})
		})

		// Configuration
		v1.GET("/config/:collector_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "not implemented yet"})
		})
	}

	// Start server
	logger.Info("Starting HTTP server", zap.String("address", ":"+port))
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
