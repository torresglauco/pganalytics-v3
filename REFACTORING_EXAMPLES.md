# PGAnalytics V3 - Refactoring Examples & Solutions

This document provides concrete refactoring examples for the issues identified in the Code Quality Report.

---

## 1. CRITICAL: Fix Circuit Breaker Logic

### Problem
```go
// BROKEN: Returns true when circuit is CLOSED (normal operation)
func (cb *CircuitBreaker) IsOpen() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    if cb.state == StateClosed {
        return true  // BUG: Should return false
    }
    // ... rest
}
```

### Solution
```go
// File: /backend/internal/ml/circuit_breaker.go

// IsOpen checks if the circuit is open (blocking requests)
func (cb *CircuitBreaker) IsOpen() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    if cb.state == StateClosed {
        return false  // Circuit is closed (operational)
    }

    if cb.state == StateOpen {
        // Check if timeout has elapsed to try recovery
        if time.Since(cb.lastFailureTime) > cb.timeout {
            // Upgrade to half-open (need write lock)
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.failureCount = 0
            cb.successCount = 0
            cb.logger.Info("Circuit breaker transitioned to half-open - testing recovery")
            cb.mu.Unlock()
            cb.mu.RLock()
            return false  // Allow test request
        }
        return true  // Still blocked
    }

    // Half-open state allows requests to test recovery
    return false
}

// IsHealthy is a convenience method that's clearer
func (cb *CircuitBreaker) IsHealthy() bool {
    return !cb.IsOpen()  // More intuitive naming
}
```

### Usage Update
```go
// File: /backend/internal/ml/client.go

func (c *Client) TrainPerformanceModel(ctx context.Context, req *TrainingRequest) (*TrainingResponse, error) {
    // BEFORE (with bug):
    // if !c.circuitBreaker.IsOpen() {
    //     c.logger.Debug("Circuit breaker is open for ML service")
    //     return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
    // }

    // AFTER (fixed):
    if c.circuitBreaker.IsOpen() {
        c.logger.Debug("Circuit breaker is open for ML service")
        return nil, fmt.Errorf("ML service unavailable (circuit breaker open)")
    }

    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    resp, err := c.doRequest(ctx, "POST", "/api/train/performance-model", body)
    if err != nil {
        c.circuitBreaker.RecordFailure()
        return nil, err
    }
    defer func() { _ = resp.Body.Close() }()

    if resp.StatusCode >= 400 {
        c.circuitBreaker.RecordFailure()
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ML service error: %d - %s", resp.StatusCode, string(bodyBytes))
    }

    c.circuitBreaker.RecordSuccess()

    var trainingResp TrainingResponse
    if err := json.NewDecoder(resp.Body).Decode(&trainingResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &trainingResp, nil
}
```

---

## 2. Refactor Duplicate Metrics Handlers

### Problem: Massive Code Duplication

```go
// handlers_metrics.go - Repeated pattern in 6+ handlers
func (s *Server) handleGetSchemaMetrics(c *gin.Context) {
    collectorIDStr := c.Param("collector_id")
    collectorID, err := uuid.Parse(collectorIDStr)
    if err != nil {
        errResp := apperrors.BadRequest("Invalid collector ID", err.Error())
        c.JSON(errResp.StatusCode, errResp)
        return
    }

    database := c.Query("database")
    limit := 100
    if l, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && l > 0 && l <= 1000 {
        limit = l
    }

    offset := 0
    if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
        offset = o
    }

    ctx := c.Request.Context()
    var dbPtr *string
    if database != "" {
        dbPtr = &database
    }

    metrics, err := s.postgres.GetSchemaMetrics(ctx, collectorID, dbPtr, limit, offset)
    if err != nil {
        c.JSON(err.(*apperrors.AppError).StatusCode, err)
        return
    }

    resp := &models.MetricsResponse{
        MetricType: "pg_schema",
        Count:      len(metrics.Tables),
        Timestamp:  time.Now(),
        Data:       metrics,
    }

    c.JSON(http.StatusOK, resp)
}

// Same pattern repeated for lock, bloat, cache, connection, extension metrics...
```

### Solution: Create Generic Handler Factory

```go
// File: /backend/internal/api/handlers_metrics_refactored.go

package api

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
    "net/http"
    "strconv"
    "time"
)

// MetricsQuery represents common metrics query parameters
type MetricsQuery struct {
    CollectorID uuid.UUID
    Database    *string
    Limit       int
    Offset      int
}

// ParseMetricsQuery extracts and validates common metrics query parameters
func (s *Server) ParseMetricsQuery(c *gin.Context) (*MetricsQuery, error) {
    collectorIDStr := c.Param("collector_id")
    collectorID, err := uuid.Parse(collectorIDStr)
    if err != nil {
        return nil, apperrors.BadRequest("Invalid collector ID", err.Error())
    }

    limit := parseLimit(c.DefaultQuery("limit", "100"))
    offset := parseOffset(c.DefaultQuery("offset", "0"))

    database := c.Query("database")
    var dbPtr *string
    if database != "" {
        dbPtr = &database
    }

    return &MetricsQuery{
        CollectorID: collectorID,
        Database:    dbPtr,
        Limit:       limit,
        Offset:      offset,
    }, nil
}

// parseLimit safely parses limit parameter with bounds
func parseLimit(limitStr string) int {
    limit := 100
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
        limit = l
    }
    return limit
}

// parseOffset safely parses offset parameter
func parseOffset(offsetStr string) int {
    offset := 0
    if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
        offset = o
    }
    return offset
}

// MetricsHandler function type for metrics retrieval
type MetricsHandler func(ctx context.Context, query *MetricsQuery) (interface{}, int, error)

// createMetricsEndpoint creates a handler for a metrics endpoint
func (s *Server) createMetricsEndpoint(handler MetricsHandler, metricType string, countFunc func(interface{}) int) gin.HandlerFunc {
    return func(c *gin.Context) {
        query, err := s.ParseMetricsQuery(c)
        if err != nil {
            appErr := apperrors.ToAppError(err)
            c.JSON(appErr.StatusCode, appErr)
            return
        }

        ctx := c.Request.Context()
        metrics, statusCode, err := handler(ctx, query)
        if err != nil {
            appErr := apperrors.ToAppError(err)
            c.JSON(appErr.StatusCode, appErr)
            return
        }

        resp := &models.MetricsResponse{
            MetricType: metricType,
            Count:      countFunc(metrics),
            Timestamp:  time.Now(),
            Data:       metrics,
        }

        c.JSON(statusCode, resp)
    }
}

// Specific handler implementations
func (s *Server) handleGetSchemaMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()  // You may want to pass ctx as parameter
    metrics, err := s.postgres.GetSchemaMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

func (s *Server) handleGetLockMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()
    metrics, err := s.postgres.GetLockMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

func (s *Server) handleGetBloatMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()
    metrics, err := s.postgres.GetBloatMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

func (s *Server) handleGetCacheMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()
    metrics, err := s.postgres.GetCacheMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

func (s *Server) handleGetConnectionMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()
    metrics, err := s.postgres.GetConnectionMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

func (s *Server) handleGetExtensionMetrics(query *MetricsQuery) (interface{}, int, error) {
    ctx := context.Background()
    metrics, err := s.postgres.GetExtensionMetrics(ctx, query.CollectorID, query.Database, query.Limit, query.Offset)
    if err != nil {
        return nil, 0, err
    }
    return metrics, http.StatusOK, nil
}

// Registration in Router
func (s *Server) RegisterMetricsRoutes(router *gin.Engine) {
    // Schema metrics
    router.GET("/api/v1/collectors/:collector_id/schema",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetSchemaMetrics(q)
            },
            "pg_schema",
            func(data interface{}) int {
                return len(data.(*models.SchemaMetricsResponse).Tables)
            },
        ),
    )

    // Lock metrics
    router.GET("/api/v1/collectors/:collector_id/locks",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetLockMetrics(q)
            },
            "pg_locks",
            func(data interface{}) int {
                return len(data.(*models.LockMetricsResponse).ActiveLocks)
            },
        ),
    )

    // Bloat metrics
    router.GET("/api/v1/collectors/:collector_id/bloat",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetBloatMetrics(q)
            },
            "pg_bloat",
            func(data interface{}) int {
                return len(data.(*models.BloatMetricsResponse).TableBloat)
            },
        ),
    )

    // Cache metrics
    router.GET("/api/v1/collectors/:collector_id/cache-hits",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetCacheMetrics(q)
            },
            "pg_cache",
            func(data interface{}) int {
                return len(data.(*models.CacheMetricsResponse).TableCacheHit)
            },
        ),
    )

    // Connection metrics
    router.GET("/api/v1/collectors/:collector_id/connections",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetConnectionMetrics(q)
            },
            "pg_connections",
            func(data interface{}) int {
                return len(data.(*models.ConnectionMetricsResponse).ConnectionSummary)
            },
        ),
    )

    // Extension metrics
    router.GET("/api/v1/collectors/:collector_id/extensions",
        s.createMetricsEndpoint(
            func(ctx context.Context, q *MetricsQuery) (interface{}, int, error) {
                return s.handleGetExtensionMetrics(q)
            },
            "pg_extensions",
            func(data interface{}) int {
                return len(data.(*models.ExtensionMetricsResponse).Extensions)
            },
        ),
    )
}
```

**Benefits:**
- Single source of truth for parameter parsing
- Consistent error handling
- DRY (Don't Repeat Yourself)
- Easier to add validation logic
- Simpler to test

---

## 3. Fix Error Handling in Alert Worker

### Problem
```go
func (aw *AlertWorker) Start(ctx context.Context) {
    go func() {
        // Run immediately on start
        aw.evaluateAlerts(ctx)  // Errors ignored

        for {
            select {
            case <-aw.ticker.C:
                aw.evaluateAlerts(ctx)  // Errors ignored
            case <-aw.done:
                return
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

### Solution
```go
// File: /backend/pkg/services/alert_worker.go

package services

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/torresglauco/pganalytics-v3/backend/internal/storage"
    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
    "go.uber.org/zap"
)

type AlertWorker struct {
    db        *storage.PostgresDB
    wsManager *ConnectionManager
    logger    *zap.Logger
    ticker    *time.Ticker
    done      chan bool
    wg        sync.WaitGroup  // Add WaitGroup
    stopped   bool            // Add flag to prevent double-stop
    mu        sync.Mutex      // Protect stopped flag
}

func NewAlertWorker(db *storage.PostgresDB, wsManager *ConnectionManager, logger *zap.Logger) *AlertWorker {
    if logger == nil {
        logger, _ = zap.NewProduction()
    }
    return &AlertWorker{
        db:        db,
        wsManager: wsManager,
        logger:    logger,
        ticker:    time.NewTicker(60 * time.Second),
        done:      make(chan bool, 1),  // Buffered to prevent deadlock
    }
}

// Start begins the alert evaluation loop
func (aw *AlertWorker) Start(ctx context.Context) {
    aw.wg.Add(1)
    go func() {
        defer aw.wg.Done()

        aw.logger.Info("Alert worker started")

        // Run immediately on start
        if err := aw.evaluateAlerts(ctx); err != nil {
            aw.logger.Error("Initial alert evaluation failed",
                zap.Error(err),
                zap.Time("timestamp", time.Now()),
            )
        }

        for {
            select {
            case <-aw.ticker.C:
                if err := aw.evaluateAlerts(ctx); err != nil {
                    aw.logger.Error("Alert evaluation error",
                        zap.Error(err),
                        zap.Time("timestamp", time.Now()),
                    )
                    // Don't return on error - keep trying
                }

            case <-aw.done:
                aw.logger.Info("Alert worker stopping")
                return

            case <-ctx.Done():
                aw.logger.Info("Alert worker context cancelled",
                    zap.Error(ctx.Err()),
                )
                return
            }
        }
    }()
}

// Stop gracefully stops the alert worker
func (aw *AlertWorker) Stop() {
    aw.mu.Lock()
    if aw.stopped {
        aw.mu.Unlock()
        return
    }
    aw.stopped = true
    aw.mu.Unlock()

    aw.ticker.Stop()
    select {
    case aw.done <- true:
    default:
        // Channel already has value
    }
}

// Wait waits for the worker to finish with timeout
func (aw *AlertWorker) Wait(timeout time.Duration) error {
    done := make(chan struct{})
    go func() {
        aw.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("alert worker shutdown timeout")
    }
}

// evaluateAlerts checks all active alerts and creates triggers if conditions met
func (aw *AlertWorker) evaluateAlerts(ctx context.Context) error {
    aw.logger.Debug("Starting alert evaluation")

    // Fetch all active alert rules from database
    alerts, err := aw.db.GetActiveAlertRules(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch alert rules: %w", err)
    }

    aw.logger.Debug("Fetched alert rules for evaluation",
        zap.Int("count", len(alerts)),
    )

    successCount := 0
    errorCount := 0

    for _, alert := range alerts {
        // Check if already triggered recently (within 5 minutes)
        if aw.recentlyTriggered(ctx, int64(alert.ID)) {
            aw.logger.Debug("Alert triggered recently, skipping",
                zap.Int("alert_id", alert.ID),
            )
            continue
        }

        // Get instances for each alert to evaluate
        instances, err := aw.db.GetAlertInstances(ctx, int64(alert.ID))
        if err != nil {
            aw.logger.Error("Error fetching instances",
                zap.Error(err),
                zap.Int("alert_id", alert.ID),
            )
            errorCount++
            continue  // Continue with next alert
        }

        for _, instance := range instances {
            // Evaluate alert condition for this instance
            shouldTrigger, currentValue, err := aw.evaluateConditions(ctx, alert, instance)
            if err != nil {
                aw.logger.Error("Error evaluating condition",
                    zap.Error(err),
                    zap.Int("alert_id", alert.ID),
                    zap.Int("instance_id", instance.ID),
                )
                errorCount++
                continue  // Continue with next instance
            }

            if shouldTrigger {
                // Create alert trigger
                trigger := &models.AlertTrigger{
                    AlertID:     int64(alert.ID),
                    // ... rest of trigger creation
                }
                successCount++
            }
        }
    }

    aw.logger.Info("Alert evaluation completed",
        zap.Int("successful", successCount),
        zap.Int("errors", errorCount),
    )

    return nil
}

// evaluateConditions evaluates if an alert should trigger
func (aw *AlertWorker) evaluateConditions(ctx context.Context, alert *models.AlertRule, instance *models.AlertInstance) (bool, float64, error) {
    // Implementation...
    return false, 0, nil
}

// recentlyTriggered checks if alert was triggered recently
func (aw *AlertWorker) recentlyTriggered(ctx context.Context, alertID int64) bool {
    // Implementation...
    return false
}
```

**Usage in main.go:**
```go
// In main.go
alertWorker := services.NewAlertWorker(db, wsManager, logger)
alertWorker.Start(context.Background())

// On shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

<-sigChan
logger.Info("Shutdown signal received, stopping alert worker...")

alertWorker.Stop()
if err := alertWorker.Wait(10 * time.Second); err != nil {
    logger.Error("Alert worker shutdown timeout", zap.Error(err))
}
```

---

## 4. Fix Database Query Construction

### Problem
```go
query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
args = append(args, limit, offset)
```

### Solution
```go
// Helper function to manage query parameters safely
func appendPaginationParams(query string, args []interface{}, limit, offset int) (string, []interface{}) {
    paramNum := len(args) + 1
    query += fmt.Sprintf(" ORDER BY time DESC LIMIT $%d OFFSET $%d", paramNum, paramNum+1)
    args = append(args, limit, offset)
    return query, args
}

// Usage
func (p *PostgresDB) GetSchemaMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.SchemaMetricsResponse, error) {
    resp := &models.SchemaMetricsResponse{}

    query := `SELECT database_name, schema_name, table_name, table_type
              FROM metrics_pg_schema_tables
              WHERE collector_id = $1`
    args := []interface{}{collectorID}

    if database != nil {
        args = append(args, *database)
        query += fmt.Sprintf(" AND database_name = $%d", len(args))
    }

    // Use helper function for consistent pagination
    query, args = appendPaginationParams(query, args, limit, offset)

    rows, err := p.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, apperrors.DatabaseError("query schema tables", err.Error())
    }
    defer func() { _ = rows.Close() }()

    // Check for row iteration errors
    for rows.Next() {
        t := &models.SchemaTable{CollectorID: collectorID}
        if err := rows.Scan(&t.DatabaseName, &t.SchemaName, &t.TableName, &t.TableType); err != nil {
            return nil, apperrors.DatabaseError("scan schema table", err.Error())
        }
        resp.Tables = append(resp.Tables, t)
    }

    // Check for iteration errors
    if err = rows.Err(); err != nil {
        return nil, apperrors.DatabaseError("iterate schema tables", err.Error())
    }

    return resp, nil
}
```

---

## 5. Implement Graceful Shutdown

### Solution
```go
// File: /backend/cmd/pganalytics-api/main.go

package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"
)

func main() {
    // ... initialization code ...

    // Create alert worker
    alertWorker := services.NewAlertWorker(postgresDB, wsManager, logger)
    alertWorker.Start(context.Background())

    // Setup signal handling for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

    // Start server in goroutine
    go func() {
        if err := router.Run(":8080"); err != nil {
            logger.Fatal("Server error", zap.Error(err))
        }
    }()

    // Wait for shutdown signal
    sig := <-sigChan
    logger.Info("Shutdown signal received",
        zap.String("signal", sig.String()),
        zap.Time("time", time.Now()),
    )

    // Create shutdown context with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Shutdown components in order
    logger.Info("Stopping alert worker...")
    alertWorker.Stop()
    if err := alertWorker.Wait(10 * time.Second); err != nil {
        logger.Error("Alert worker shutdown failed", zap.Error(err))
    }

    logger.Info("Closing database connections...")
    if postgresDB != nil {
        if err := postgresDB.Close(); err != nil {
            logger.Error("Database close error", zap.Error(err))
        }
    }

    logger.Info("Shutdown complete")
}
```

---

## 6. Implement Comprehensive Health Check

### Solution
```go
// File: /backend/internal/api/handlers_health.go

package api

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type ComponentHealth struct {
    Status    string `json:"status"`  // "healthy", "degraded", "unhealthy"
    Latency   int64  `json:"latency_ms"`
    Error     string `json:"error,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

type HealthResponse struct {
    Status        string                      `json:"status"`
    Timestamp     time.Time                   `json:"timestamp"`
    Components    map[string]ComponentHealth  `json:"components"`
    OverallHealth string                      `json:"overall_health"`
}

func (s *Server) handleHealth(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    health := &HealthResponse{
        Status:     "healthy",
        Timestamp:  time.Now(),
        Components: make(map[string]ComponentHealth),
    }

    // Check database
    start := time.Now()
    if dbHealth := s.checkDatabase(ctx); dbHealth != nil {
        health.Components["database"] = *dbHealth
    }

    // Check TimescaleDB
    if tsHealth := s.checkTimescaleDB(ctx); tsHealth != nil {
        health.Components["timescaledb"] = *tsHealth
    }

    // Check ML service
    if mlHealth := s.checkMLService(ctx); mlHealth != nil {
        health.Components["ml_service"] = *mlHealth
    }

    // Check cache
    if cacheHealth := s.checkCache(ctx); cacheHealth != nil {
        health.Components["cache"] = *cacheHealth
    }

    // Check auth service
    if authHealth := s.checkAuth(ctx); authHealth != nil {
        health.Components["auth"] = *authHealth
    }

    // Determine overall health
    unhealthyCount := 0
    degradedCount := 0
    for _, comp := range health.Components {
        if comp.Status == "unhealthy" {
            unhealthyCount++
        } else if comp.Status == "degraded" {
            degradedCount++
        }
    }

    if unhealthyCount > 0 {
        health.Status = "unhealthy"
        health.OverallHealth = "Critical: Some components are unhealthy"
        c.JSON(http.StatusServiceUnavailable, health)
        return
    }

    if degradedCount > 0 {
        health.Status = "degraded"
        health.OverallHealth = "Warning: Some components are degraded"
        c.JSON(http.StatusOK, health)
        return
    }

    health.Status = "healthy"
    health.OverallHealth = "All systems operational"
    c.JSON(http.StatusOK, health)
}

func (s *Server) checkDatabase(ctx context.Context) *ComponentHealth {
    start := time.Now()
    health := &ComponentHealth{Timestamp: time.Now()}

    if s.postgres == nil {
        health.Status = "unhealthy"
        health.Error = "Database not initialized"
        return health
    }

    if !s.postgres.Health(ctx) {
        health.Status = "unhealthy"
        health.Error = "Database ping failed"
        health.Latency = time.Since(start).Milliseconds()
        return health
    }

    health.Status = "healthy"
    health.Latency = time.Since(start).Milliseconds()
    return health
}

func (s *Server) checkTimescaleDB(ctx context.Context) *ComponentHealth {
    start := time.Now()
    health := &ComponentHealth{Timestamp: time.Now()}

    if s.timescale == nil {
        health.Status = "degraded"
        health.Error = "TimescaleDB not initialized"
        return health
    }

    if !s.timescale.Health(ctx) {
        health.Status = "degraded"
        health.Error = "TimescaleDB ping failed"
        health.Latency = time.Since(start).Milliseconds()
        return health
    }

    health.Status = "healthy"
    health.Latency = time.Since(start).Milliseconds()
    return health
}

func (s *Server) checkMLService(ctx context.Context) *ComponentHealth {
    start := time.Now()
    health := &ComponentHealth{Timestamp: time.Now()}

    if s.mlClient == nil {
        health.Status = "degraded"
        health.Error = "ML service not configured"
        return health
    }

    if !s.mlClient.IsHealthy(ctx) {
        health.Status = "degraded"
        health.Error = "ML service unavailable"
        health.Latency = time.Since(start).Milliseconds()
        return health
    }

    cbState := s.mlClient.GetCircuitBreakerState()
    if cbState != "closed" {
        health.Status = "degraded"
        health.Error = fmt.Sprintf("Circuit breaker: %s", cbState)
        return health
    }

    health.Status = "healthy"
    health.Latency = time.Since(start).Milliseconds()
    return health
}

func (s *Server) checkCache(ctx context.Context) *ComponentHealth {
    health := &ComponentHealth{Timestamp: time.Now()}

    if s.cacheManager == nil {
        health.Status = "degraded"
        health.Error = "Cache not initialized"
        return health
    }

    // Check if cache is operational
    if !s.cacheManager.IsHealthy() {
        health.Status = "degraded"
        health.Error = "Cache unavailable"
        return health
    }

    health.Status = "healthy"
    return health
}

func (s *Server) checkAuth(ctx context.Context) *ComponentHealth {
    health := &ComponentHealth{Timestamp: time.Now()}

    if s.authService == nil {
        health.Status = "unhealthy"
        health.Error = "Auth service not initialized"
        return health
    }

    health.Status = "healthy"
    return health
}
```

**Registration:**
```go
// In /backend/internal/api/server.go SetupRoutes() method
public.GET("/health", s.handleHealth)
public.GET("/api/v1/health", s.handleHealth)
```

---

## Summary of Changes

| Issue | Effort | Impact | Priority |
|-------|--------|--------|----------|
| Circuit Breaker Logic | 5 min | Critical Bug Fix | P0 |
| Hardcoded Credentials | 15 min | Security | P0 |
| Handler Duplication | 2-4 hrs | Maintenance | P1 |
| Error Handling | 2-3 hrs | Operational | P1 |
| Query Building | 2-3 hrs | Code Quality | P2 |
| Graceful Shutdown | 1-2 hrs | Reliability | P2 |
| Health Checks | 2-3 hrs | Operational | P2 |

---

**Generated:** April 14, 2026
