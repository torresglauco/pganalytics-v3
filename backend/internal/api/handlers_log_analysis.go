package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"go.uber.org/zap"
)

// ============================================================================
// LOG ANALYSIS API ENDPOINTS
// ============================================================================

// @Summary Get Logs for Collector
// @Description Fetch recent logs for a specific collector instance
// @Tags Logs
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Param limit query int false "Limit results (default: 100)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param log_level query string false "Filter by log level (DEBUG, INFO, WARNING, ERROR, FATAL)"
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 404 {object} apperrors.AppError
// @Router /api/v1/logs/collector/{collector_id} [get]
func (s *Server) handleGetCollectorLogs(c *gin.Context) {
	collectorID := c.Param("collector_id")

	// Parse optional query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	logLevel := c.Query("log_level")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	s.logger.Debug("Fetching logs for collector",
		zap.String("collector_id", collectorID),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("log_level", logLevel))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Fetch logs from storage
	var logs interface{}
	var total int64
	var err2 error

	if logLevel != "" {
		// Fetch logs filtered by level
		logs, err2 = s.postgres.GetPostgresqlLogsByLevel(ctx, 0, logLevel, limit, offset)
	} else {
		// Fetch all recent logs
		logs, err2 = s.postgres.GetPostgresqlLogs(ctx, 0, limit, offset)
	}

	if err2 != nil {
		s.logger.Error("Failed to fetch logs",
			zap.String("collector_id", collectorID),
			zap.Error(err2))
		errResp := apperrors.InternalServerError("Failed to fetch logs", err2.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Calculate pagination info
	totalPages := 1
	if total > 0 && limit > 0 {
		totalPages = int(total) / limit
		if int(total)%limit > 0 {
			totalPages++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         logs,
		"total":        total,
		"page":         (offset / limit) + 1,
		"page_size":    limit,
		"total_pages":  totalPages,
		"collector_id": collectorID,
	})
}

// @Summary Stream Logs via WebSocket
// @Description Stream real-time logs for a collector via WebSocket
// @Tags Logs
// @Produce json
// @Security Bearer
// @Param collector_id path string true "Collector ID"
// @Success 101 {string} string "WebSocket Upgrade"
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/logs/stream/{collector_id} [get]
func (s *Server) handleLogStream(c *gin.Context) {
	collectorID := c.Param("collector_id")

	s.logger.Debug("WebSocket connection attempt for logs stream", zap.String("collector_id", collectorID))

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed",
			zap.String("collector_id", collectorID),
			zap.Error(err))
		return
	}
	defer conn.Close()

	s.logger.Info("WebSocket connection established for logs stream", zap.String("collector_id", collectorID))

	// Create a channel for receiving logs
	logChan := make(chan map[string]interface{}, 10)
	defer close(logChan)

	// Create a context for cancellation
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Start streaming logs in a goroutine
	go func() {
		if s.logCollector == nil {
			s.logger.Error("LogCollector not initialized")
			return
		}

		// StreamLogs will continuously poll the database and send logs through the channel
		if err := s.logCollector.StreamLogs(ctx, collectorID, logChan); err != nil {
			s.logger.Debug("Log streaming ended",
				zap.String("collector_id", collectorID),
				zap.Error(err))
		}
	}()

	// Main loop: receive logs from channel and send via WebSocket
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("Context cancelled for logs stream", zap.String("collector_id", collectorID))
			return

		case log, ok := <-logChan:
			if !ok {
				s.logger.Debug("Log channel closed", zap.String("collector_id", collectorID))
				return
			}

			// Send log via WebSocket
			if err := conn.WriteJSON(log); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.Debug("WebSocket error", zap.Error(err))
				}
				return
			}
		}
	}
}
