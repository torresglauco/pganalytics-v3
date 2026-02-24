# pgAnalytics v3 - Detailed Implementation Roadmap

**Date**: February 22, 2026
**Project**: pganalytics-v3 (torresglauco)
**Focus**: PostgreSQL 18 Support + AI Intelligence Integration

---

## Phase 1: PostgreSQL 18 Support (High Priority - Start Immediately)

### 1.1 Backend Driver Migration (lib/pq → pgx)

#### Task 1.1.1: Audit Current Setup
- [ ] Check current lib/pq version
- [ ] Identify all database operations
- [ ] List PostgreSQL features used
- [ ] Create compatibility matrix

#### Task 1.1.2: Add pgx Dependency
```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/stdlib  # For sql.DB compatibility
```

#### Task 1.1.3: Migration Plan
```go
// Gradual migration strategy:
// 1. Keep sql.DB interface (no breaking changes)
// 2. Use pgx stdlib adapter
// 3. Slowly migrate to native pgx for performance

// New connection initialization:
config, err := pgx.ParseConfig(connString)
config.IncludeErrorDetails = true
config.DefaultQueryExecMode = pgx.QueryExecModeExec

connPool, err := pgxpool.NewWithConfig(ctx, config)
defer connPool.Close()

// Direct pgx usage for performance-critical paths
conn, err := connPool.Acquire(ctx)
defer conn.Release()

rows, err := conn.Query(ctx, query, args...)
```

**Files to Update**:
- `backend/internal/storage/postgres.go` - SQL operations
- `backend/cmd/pganalytics-api/main.go` - Connection setup
- `go.mod` - Dependencies

### 1.2 PostgreSQL 18 Features Integration

#### Task 1.2.1: JSON Subscripting Support
```sql
-- Enable queries with new JSON syntax (PG 14+)
SELECT
    query_id,
    query_params -> 'filter' ->> 'table' AS table_name,
    query_params -> 'limit' AS limit_value,
    JSON_QUERY(execution_plan, '$.Plan.Plans[*].Node Type') AS nodes
FROM query_executions
WHERE query_params IS NOT NULL;
```

**Implementation**:
- [ ] Create migration for JSON subscripting queries
- [ ] Add helper functions for JSON path queries
- [ ] Test with PostgreSQL 18

#### Task 1.2.2: Advanced Query Analysis (PG 18)
```sql
-- MERGE statement support (PG 15+, enhanced in PG 18)
MERGE INTO query_stats AS target
USING new_stats AS source
ON target.query_hash = source.query_hash
WHEN MATCHED THEN
    UPDATE SET
        call_count = call_count + source.calls,
        total_time = total_time + source.time,
        updated_at = NOW()
WHEN NOT MATCHED THEN
    INSERT (query_hash, call_count, total_time)
    VALUES (source.query_hash, source.calls, source.time);
```

**Implementation**:
- [ ] Implement MERGE for query stats upserting
- [ ] Test performance improvement vs current approach
- [ ] Create migration scripts

#### Task 1.2.3: Performance Monitoring Enhancements
```sql
-- pg_stat_statements enhancements
CREATE OR REPLACE FUNCTION analyze_query_performance()
RETURNS TABLE (
    query_id UUID,
    mean_time FLOAT8,
    stddev_time FLOAT8,
    is_anomalous BOOLEAN,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        query_hash,
        mean_exec_time,
        stddev_exec_time,
        stddev_exec_time > mean_exec_time * 0.5,
        CASE
            WHEN total_time > 60000 THEN 'Consider optimization'
            WHEN calls > 10000 AND mean_time > 100 THEN 'Frequent slow query'
            ELSE NULL
        END
    FROM pg_stat_statements;
END;
$$ LANGUAGE plpgsql;
```

**Implementation**:
- [ ] Create analysis functions for PG 18
- [ ] Add performance views
- [ ] Integrate with collector

### 1.3 Collector Updates (C++)

#### Task 1.3.1: libpq Version Check
- [ ] Verify CMakeLists.txt for libpq version
- [ ] Update to require libpq >= 18.0
- [ ] Test with PostgreSQL 18

**CMakeLists.txt Changes**:
```cmake
# Find PostgreSQL with version constraint
find_package(PostgreSQL 18.0 REQUIRED)
include_directories(${PostgreSQL_INCLUDE_DIRS})
target_link_libraries(pganalytics-collector ${PostgreSQL_LIBRARIES})
```

#### Task 1.3.2: Enhanced Data Collection
- [ ] Collect new PG 18 metrics
- [ ] Implement advanced wait sampling
- [ ] Add query plan caching

### 1.4 Version Compatibility Testing

#### Task 1.4.1: Test Matrix
- [ ] PostgreSQL 15 (LTS)
- [ ] PostgreSQL 16 (LTS)
- [ ] PostgreSQL 17
- [ ] PostgreSQL 18 (Latest)

**Tests to Create**:
```go
func TestPostgreSQL15Compatibility(t *testing.T)   {}
func TestPostgreSQL16Compatibility(t *testing.T)   {}
func TestPostgreSQL17Compatibility(t *testing.T)   {}
func TestPostgreSQL18Compatibility(t *testing.T)   {}
func TestPostgreSQL18Features(t *testing.T)        {}
```

---

## Phase 2: AI Foundation Layer (Core ML Infrastructure)

### 2.1 Feature Extraction Pipeline

#### Task 2.1.1: Time-Series Feature Engineering
```go
// File: backend/internal/ai/features/timeseries.go

type TimeSeriesFeature struct {
    Timestamp  time.Time
    Value      float64
    MovingAvg  float64    // 5-min average
    MovingStd  float64    // 5-min std dev
    RateChange float64    // % change from last point
    Trend      string     // "up", "down", "stable"
}

type TimeSeriesAnalyzer struct {
    window time.Duration
    metrics map[string]*TimeSeries
}

func (tsa *TimeSeriesAnalyzer) ExtractFeatures(
    queryHash int64,
    lookbackPeriod time.Duration,
) map[string]float64 {
    features := make(map[string]float64)

    // Aggregate metrics
    features["mean_execution_time"] = calculateMean(...)
    features["stddev_execution_time"] = calculateStdDev(...)
    features["p95_execution_time"] = calculatePercentile(...)
    features["trend_direction"] = calculateTrend(...)
    features["volatility"] = calculateVolatility(...)

    return features
}
```

**Files to Create**:
- `backend/internal/ai/features/timeseries.go`
- `backend/internal/ai/features/query.go`
- `backend/internal/ai/features/system.go`

#### Task 2.1.2: Feature Caching
- [ ] Cache computed features in Redis
- [ ] 5-minute TTL for feature vectors
- [ ] Implement feature invalidation

### 2.2 Statistical Anomaly Detection

#### Task 2.2.1: Implement Z-Score Based Detection
```go
// File: backend/internal/ai/anomaly/detector.go

type AnomalyDetector struct {
    metrics map[string]*MetricStats
    mu      sync.RWMutex
}

type MetricStats struct {
    Mean       float64
    StdDev     float64
    LastUpdate time.Time
}

func (ad *AnomalyDetector) Detect(metric string, value float64) *Anomaly {
    ad.mu.RLock()
    stats := ad.metrics[metric]
    ad.mu.RUnlock()

    if stats == nil {
        return nil // Not enough data yet
    }

    zscore := (value - stats.Mean) / stats.StdDev

    if math.Abs(zscore) > 3.0 {
        return &Anomaly{
            MetricName:      metric,
            Value:           value,
            ExpectedValue:   stats.Mean,
            ZScore:          zscore,
            Severity:        determineSeverity(zscore),
            DetectedAt:      time.Now(),
            Description:     fmt.Sprintf("Value %.2f is %.1f standard deviations from mean", value, zscore),
        }
    }

    return nil
}

func (ad *AnomalyDetector) UpdateStats(metric string, value float64) {
    // Use Welford's online algorithm for efficient std dev calculation
    ad.mu.Lock()
    defer ad.mu.Unlock()

    if ad.metrics[metric] == nil {
        ad.metrics[metric] = &MetricStats{
            Mean:       value,
            StdDev:     0,
            LastUpdate: time.Now(),
        }
    } else {
        stats := ad.metrics[metric]
        // Incremental mean update
        stats.Mean = (stats.Mean + value) / 2
        // Incremental std dev update
        // ... implement Welford's algorithm
    }
}
```

**Files to Create**:
- `backend/internal/ai/anomaly/detector.go`
- `backend/internal/ai/anomaly/models.go`

#### Task 2.2.2: Query Performance Anomalies
- [ ] Detect execution time spikes
- [ ] Identify cache hit rate drops
- [ ] Monitor lock contention changes
- [ ] Track connection pool saturation

**Anomalies to Detect**:
```go
type AnomalyType string

const (
    ExecutionTimeSpike    AnomalyType = "execution_time_spike"
    CacheHitRateDrop      AnomalyType = "cache_hit_rate_drop"
    LockContentionIncrease AnomalyType = "lock_contention_increase"
    ConnectionPoolSaturation AnomalyType = "connection_pool_saturation"
    HighIOWait            AnomalyType = "high_io_wait"
    HighCPUUsage          AnomalyType = "high_cpu_usage"
    MemoryPressure        AnomalyType = "memory_pressure"
)
```

### 2.3 Correlation Analysis Engine

#### Task 2.3.1: Implement Correlation Matrix
```go
// File: backend/internal/ai/correlation/engine.go

type CorrelationEngine struct {
    timeSeries map[string][]float64
    window     time.Duration
}

func (ce *CorrelationEngine) CalculateCorrelations() map[string]map[string]float64 {
    correlations := make(map[string]map[string]float64)

    metrics := []string{"execution_time", "lock_wait", "memory_usage", "cache_hits"}

    for i := 0; i < len(metrics); i++ {
        for j := i + 1; j < len(metrics); j++ {
            corr := calculatePearsonCorrelation(
                ce.timeSeries[metrics[i]],
                ce.timeSeries[metrics[j]],
            )

            if math.Abs(corr) > 0.7 { // Strong correlation threshold
                if correlations[metrics[i]] == nil {
                    correlations[metrics[i]] = make(map[string]float64)
                }
                correlations[metrics[i]][metrics[j]] = corr
            }
        }
    }

    return correlations
}

func (ce *CorrelationEngine) FindTimeLagCorrelations() map[string][]TimeLagCorrelation {
    // For each metric pair, find if one leads the other
    lags := map[string][]TimeLagCorrelation{}

    // Test lags from 1 second to 5 minutes
    for lag := 1*time.Second; lag < 5*time.Minute; lag += 10*time.Second {
        // Calculate correlation with lag
    }

    return lags
}
```

**Files to Create**:
- `backend/internal/ai/correlation/engine.go`
- `backend/internal/ai/correlation/models.go`

### 2.4 Real-Time Metrics Pipeline

#### Task 2.4.1: Metrics Aggregation Service
```go
// File: backend/internal/ai/pipeline/aggregator.go

type MetricsAggregator struct {
    db        *pgxpool.Pool
    cache     *cache.Manager
    intervals []time.Duration // 1m, 5m, 15m, 1h
}

func (ma *MetricsAggregator) AggregateMetrics(ctx context.Context) error {
    // 1-minute aggregation
    if time.Now().Second() == 0 {
        ma.aggregate1Min(ctx)
    }

    // 5-minute aggregation (every 5 minutes)
    if time.Now().Minute()%5 == 0 && time.Now().Second() == 0 {
        ma.aggregate5Min(ctx)
    }

    return nil
}

func (ma *MetricsAggregator) aggregate1Min(ctx context.Context) error {
    query := `
    INSERT INTO metrics.metrics_1min (time_bucket, metric, value)
    SELECT
        time_bucket('1 minute', timestamp) as time_bucket,
        metric_name,
        AVG(value) as value
    FROM raw_metrics
    WHERE timestamp > NOW() - INTERVAL '1 minute'
    GROUP BY time_bucket, metric_name
    ON CONFLICT DO NOTHING;`

    return ma.db.QueryRow(ctx, query).Scan()
}
```

---

## Phase 3: Intelligent Analysis (Root Cause & Prediction)

### 3.1 Root Cause Analysis Module

#### Task 3.1.1: Query Performance Analysis
```go
// File: backend/internal/ai/analysis/query_analyzer.go

type QueryAnalyzer struct {
    db     *pgxpool.Pool
    cache  *cache.Manager
}

func (qa *QueryAnalyzer) AnalyzeSlowQuery(ctx context.Context, queryHash int64) *RootCauseAnalysis {
    analysis := &RootCauseAnalysis{
        QueryHash: queryHash,
        Causes:    []CausePrediction{},
    }

    // Check 1: Missing Index
    if missingIndex := qa.checkMissingIndex(ctx, queryHash); missingIndex != nil {
        analysis.Causes = append(analysis.Causes, CausePrediction{
            Cause:           "missing_index",
            Confidence:      0.95,
            Evidence:        []string{"Full table scan detected", "100k rows scanned"},
            Recommendations: []string{"CREATE INDEX idx_user_id ON orders(user_id)"},
        })
    }

    // Check 2: Suboptimal Query Plan
    if planIssue := qa.checkQueryPlan(ctx, queryHash); planIssue != nil {
        analysis.Causes = append(analysis.Causes, planIssue)
    }

    // Check 3: Lock Contention
    if locks := qa.checkLockContention(ctx, queryHash); len(locks) > 0 {
        analysis.Causes = append(analysis.Causes, CausePrediction{
            Cause:           "lock_contention",
            Confidence:      0.85,
            Evidence:        locks,
            Recommendations: []string{"Investigate blocking queries", "Consider SERIALIZABLE isolation"},
        })
    }

    // Sort by confidence
    sort.Slice(analysis.Causes, func(i, j int) bool {
        return analysis.Causes[i].Confidence > analysis.Causes[j].Confidence
    })

    return analysis
}
```

**Files to Create**:
- `backend/internal/ai/analysis/query_analyzer.go`
- `backend/internal/ai/analysis/models.go`

### 3.2 Prediction Engine

#### Task 3.2.1: Time-Series Forecasting
```go
// File: backend/internal/ai/prediction/forecaster.go

type Forecaster struct {
    historyDays int
    db          *pgxpool.Pool
}

func (f *Forecaster) ForecastMetric(
    ctx context.Context,
    metric string,
    queryHash int64,
    daysAhead int,
) *Forecast {
    // Get historical data
    history := f.getHistoricalData(ctx, metric, queryHash, f.historyDays)

    // Simple exponential smoothing (can be replaced with Prophet)
    forecast := f.exponentialSmoothing(history, daysAhead)

    // Check for concerning trends
    if forecast.ExceedsThreshold {
        forecast.Alert = &Alert{
            Level:       "warning",
            Message:    fmt.Sprintf("%s will exceed SLA in %d days", metric, forecast.DaysToThreshold),
            Recommendation: "Consider proactive optimization",
        }
    }

    return forecast
}

func (f *Forecaster) exponentialSmoothing(
    data []float64,
    periods int,
) *Forecast {
    if len(data) < 2 {
        return nil
    }

    alpha := 0.3 // Smoothing factor
    lastValue := data[len(data)-1]
    forecast := make([]float64, periods)

    for i := 0; i < periods; i++ {
        forecast[i] = alpha*lastValue + (1-alpha)*forecast[i-1]
        if i > 0 {
            lastValue = forecast[i]
        }
    }

    return &Forecast{
        Values:     forecast,
        Confidence: 0.85,
    }
}
```

---

## Phase 4: Modern API & Real-Time Features

### 4.1 API v2 with AI Endpoints

#### Task 4.1.1: AI Analysis Endpoints
```go
// File: backend/internal/api/handlers_ai.go

// POST /api/v2/ai/analyze/query/{hash}
func (s *Server) handleAIAnalyzeQuery(c *gin.Context) {
    queryHash := c.Param("hash")

    analysis := s.queryAnalyzer.AnalyzeSlowQuery(c.Request.Context(), queryHash)

    c.JSON(http.StatusOK, analysis)
}

// GET /api/v2/ai/predict/{metric}
func (s *Server) handleAIPredictMetric(c *gin.Context) {
    metric := c.Param("metric")
    days := c.DefaultQuery("days", "7")

    forecast := s.forecaster.ForecastMetric(c.Request.Context(), metric, days)

    c.JSON(http.StatusOK, forecast)
}

// POST /api/v2/ai/anomalies/detect
func (s *Server) handleAIDetectAnomalies(c *gin.Context) {
    var req struct {
        Metrics []string `json:"metrics"`
        Period  string   `json:"period"` // "1h", "24h", etc
    }

    c.ShouldBindJSON(&req)

    anomalies := s.anomalyDetector.DetectAll(c.Request.Context(), req.Metrics, req.Period)

    c.JSON(http.StatusOK, anomalies)
}

// POST /api/v2/ai/correlations/analyze
func (s *Server) handleAIAnalyzeCorrelations(c *gin.Context) {
    correlations := s.correlationEngine.CalculateCorrelations()

    c.JSON(http.StatusOK, gin.H{
        "correlations": correlations,
        "insights":     deriveInsights(correlations),
    })
}
```

**Endpoints to Create**:
- `POST /api/v2/ai/analyze/query/{hash}` - Root cause analysis
- `GET /api/v2/ai/predict/{metric}` - Forecasting
- `POST /api/v2/ai/anomalies/detect` - Anomaly detection
- `POST /api/v2/ai/correlations/analyze` - Correlation analysis
- `GET /api/v2/ai/recommendations` - Action recommendations

### 4.2 WebSocket for Real-Time Updates

#### Task 4.2.1: Real-Time Metrics Stream
```go
// File: backend/internal/api/websocket.go

func (s *Server) handleMetricsStream(c *gin.Context) {
    ws, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer ws.Close()

    // Send metrics every 5 seconds
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-c.Request.Context().Done():
            return
        case <-ticker.C:
            metrics := s.getRealtimeMetrics()
            ws.WriteJSON(metrics)
        }
    }
}
```

---

## Phase 5: Advanced Features

### 5.1 Machine Learning Model Training

#### Task 5.1.1: Offline Model Training
```go
// File: backend/internal/ai/training/trainer.go

type ModelTrainer struct {
    db *pgxpool.Pool
}

func (mt *ModelTrainer) TrainAnomalyModel(ctx context.Context) error {
    // Get historical data
    trainingData := mt.collectTrainingData(ctx)

    // Train model on patterns
    model := trainStatisticalModel(trainingData)

    // Save model
    return mt.saveModel(ctx, model)
}

func (mt *ModelTrainer) TrainRootCauseModel(ctx context.Context) error {
    // Get past incidents with known root causes
    incidents := mt.getHistoricalIncidents(ctx)

    // Train classifier
    model := trainClassifier(incidents)

    // Save model
    return mt.saveModel(ctx, model)
}
```

---

## Task Distribution by Priority

### Immediate (This Week)
1. **Task 1.1**: Audit PostgreSQL 18 support status
2. **Task 1.1.2**: Add pgx dependency
3. **Task 2.1.1**: Create feature extraction pipeline
4. **Task 2.2.1**: Implement basic anomaly detection

### Short-Term (Next Week)
1. **Task 1.1.3**: Migrate backend to pgx
2. **Task 2.3.1**: Build correlation engine
3. **Task 3.1.1**: Query analyzer
4. **Task 4.1.1**: API v2 endpoints

### Medium-Term (Weeks 3-4)
1. **Task 1.2**: PostgreSQL 18 features
2. **Task 1.3**: Collector updates
3. **Task 3.2**: Prediction engine
4. **Task 4.2**: WebSocket support

---

## Success Criteria

### PostgreSQL 18
- [ ] All tests pass on PG 15-18
- [ ] Performance improved by 10%+
- [ ] All PG 18 features supported

### AI Features
- [ ] Anomaly detection accuracy > 95%
- [ ] Root cause identification > 80% accurate
- [ ] Real-time processing < 100ms latency
- [ ] Prediction RMSE < 15%

### Modern Features
- [ ] Real-time dashboard updates < 1s
- [ ] API response times < 200ms (p95)
- [ ] Support 10k+ databases
- [ ] 1M+ metrics/second throughput

---

**Status**: Ready for Implementation
**Start Date**: February 22, 2026
**Target Completion**: April 2026

