package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dextra/pganalytics-v3/backend/internal/api"
	"github.com/dextra/pganalytics-v3/backend/internal/config"
	"github.com/dextra/pganalytics-v3/backend/internal/storage"
	"github.com/dextra/pganalytics-v3/backend/internal/timescale"
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

const version = "3.0.0-alpha"

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize logger
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Log startup
	logger.Info("pgAnalytics v3.0 API Starting",
		zap.String("version", version),
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Port),
	)

	// Initialize PostgreSQL database
	postgresDB, err := storage.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgresDB.Close()

	logger.Info("Connected to PostgreSQL")

	// Initialize TimescaleDB database
	timescaleDB, err := timescale.NewTimescaleDB(cfg.TimescaleURL)
	if err != nil {
		logger.Fatal("Failed to initialize TimescaleDB", zap.Error(err))
	}
	defer timescaleDB.Close()

	logger.Info("Connected to TimescaleDB")

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Create API server
	apiServer := api.NewServer(cfg, logger, postgresDB, timescaleDB)

	// Register routes
	apiServer.RegisterRoutes(router)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:           ":" + getEnvInt("PORT", "8080"),
		Handler:        router,
		ReadTimeout:    cfg.RequestTimeout,
		WriteTimeout:   cfg.RequestTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Server shutdown complete")
}

func getEnvInt(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
