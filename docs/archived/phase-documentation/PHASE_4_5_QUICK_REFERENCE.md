# Phase 4.5: ML-Based Query Optimization - Quick Reference Guide

## Database Schema Quick Lookup

### Tables
| Table | Purpose | Key Columns | Indexes |
|-------|---------|-------------|---------|
| `workload_patterns` | Pattern detection results | pattern_type, confidence, detection_timestamp | db+type, time |
| `query_rewrite_suggestions` | SQL rewrite recommendations | suggestion_type, estimated_improvement_percent, confidence_score | query_hash, confidence |
| `parameter_tuning_suggestions` | Parameter recommendations | parameter_name, recommended_value, confidence_score | query_hash, confidence |
| `optimization_recommendations` | Aggregated recommendations | roi_score, estimated_improvement_percent, source_type | roi, query_hash, source, dismissed |
| `optimization_implementations` | Implementation tracking | status, actual_improvement_percent, measured_at | rec_id, query, status, measured |
| `query_performance_models` | ML model storage | model_type, r_squared, is_active | database+active+updated |

### API Endpoints
| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| POST | /api/v1/workload-patterns/analyze | handleDetectWorkloadPatterns | Trigger pattern detection |
| GET | /api/v1/workload-patterns | handleGetWorkloadPatterns | List detected patterns |
| POST | /api/v1/queries/{hash}/rewrite-suggestions/generate | handleGenerateRewriteSuggestions | Generate suggestions |
| GET | /api/v1/queries/{hash}/rewrite-suggestions | handleGetRewriteSuggestions | List suggestions |
| GET | /api/v1/queries/{hash}/parameter-optimization | handleGetParameterOptimization | Get parameter recommendations |
| POST | /api/v1/queries/{hash}/predict-performance | handlePredictQueryPerformance | Predict execution time |
| GET | /api/v1/optimization-recommendations | handleGetOptimizationRecommendations | List top recommendations |
| POST | /api/v1/optimization-recommendations/{id}/implement | handleImplementRecommendation | Record implementation |
| GET | /api/v1/optimization-results | handleGetOptimizationResults | Get measured results |

### Storage Methods
| Method | Location | Purpose |
|--------|----------|---------|
| DetectWorkloadPatterns | postgres.go:1378 | Trigger pattern detection |
| GetWorkloadPatterns | postgres.go:1390 | Retrieve patterns |
| GenerateRewriteSuggestions | postgres.go:1430 | Generate rewrite suggestions |
| GetRewriteSuggestions | postgres.go:1445 | Retrieve rewrite suggestions |
| GetParameterOptimizationSuggestions | postgres.go:1480 | Get parameter recommendations |
| PredictQueryPerformance | postgres.go:1510 | Predict execution time (STUB) |
| GetOptimizationRecommendations | postgres.go:1520 | Get top recommendations |
| ImplementRecommendation | postgres.go:1570 | Record implementation |
| UpdateOptimizationResults | postgres.go:1595 | Update with measured results |
| GetOptimizationResults | postgres.go:1620 | Get implementation results |
| DismissOptimizationRecommendation | postgres.go:1670 | Dismiss recommendation |
| GetRecommendationByID | postgres.go:1690 | Get single recommendation |
| TrainPerformanceModel | postgres.go:1715 | Train ML model (STUB) |

## Code Patterns

### Handler Pattern
```go
func (s *Server) handleSomething(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
    defer cancel()

    // 1. Parse input
    var req struct { ... }
    if err := c.ShouldBindJSON(&req); err != nil {
        errResp := apperrors.BadRequest("message", err.Error())
        c.JSON(errResp.StatusCode, errResp)
        return
    }

    // 2. Validate
    if invalid {
        c.JSON(http.StatusBadRequest, gin.H{"error": "..."})
        return
    }

    // 3. Call storage layer
    result, err := s.postgres.DoSomething(ctx, args...)
    if err != nil {
        s.logger.Warnf("Error: %v", err)
        errResp := apperrors.InternalServerError("message")
        c.JSON(errResp.StatusCode, errResp)
        return
    }

    // 4. Return response
    c.JSON(http.StatusOK, result)
}
```

### Storage Method Pattern
```go
func (p *PostgresDB) DoSomething(ctx context.Context, arg1 string, arg2 int) (Result, error) {
    query := `SELECT ... FROM table WHERE col1 = $1 AND col2 = $2`

    rows, err := p.db.QueryContext(ctx, query, arg1, arg2)
    if err != nil {
        return nil, apperrors.DatabaseError("operation name", err.Error())
    }
    defer rows.Close()

    var results []Result
    for rows.Next() {
        var r Result
        err := rows.Scan(&r.Field1, &r.Field2, ...)
        if err != nil {
            return nil, apperrors.DatabaseError("scan row", err.Error())
        }
        results = append(results, r)
    }

    return results, rows.Err()
}
```

### Model Struct Pattern
```go
type ModelName struct {
    ID          int64       `db:"id" json:"id"`
    Name        string      `db:"name" json:"name"`
    Optional    *string     `db:"optional" json:"optional,omitempty"`
    Metadata    map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
    CreatedAt   time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
}
```

## Common Tasks

### Add a New Query Pattern for Rewrite Detection
1. Update `query_rewrite_suggestions` table comment in migration 005
2. Add detection logic to `GenerateRewriteSuggestions()` function in postgres.go
3. Add response field to `QueryRewriteSuggestion` struct if needed
4. Test with handler: POST /api/v1/queries/{hash}/rewrite-suggestions/generate

### Add a New Parameter Tuning Rule
1. Add rule to `GetParameterOptimizationSuggestions()` in postgres.go
2. Update database migration if new columns needed
3. Add confidence calculation logic
4. Test with handler: GET /api/v1/queries/{hash}/parameter-optimization

### Implement ML Service Integration
1. In `handlePredictQueryPerformance()`, call `callMLService()`
2. Implement `callMLService()` HTTP call to Python service
3. Add ML_SERVICE_URL to config
4. Add circuit breaker for fallback behavior
5. Test: POST /api/v1/queries/{hash}/predict-performance

### Add Recommendation Filtering
1. Update `GetOptimizationRecommendations()` to add new filter parameter
2. Add parameter to handler query parsing
3. Add WHERE clause to SQL query
4. Test: GET /api/v1/optimization-recommendations?new_filter=value

## Configuration

### Environment Variables (for ML service)
```bash
ML_SERVICE_URL=http://ml-service:8081
ML_SERVICE_TIMEOUT=5s
ML_SERVICE_ENABLED=true
CIRCUIT_BREAKER_THRESHOLD=5
```

### Database Parameters
```sql
-- Minimum lookback days for pattern detection
SET min_lookback_days = 7;

-- Maximum pattern detection window
SET max_lookback_days = 365;

-- ROI scoring thresholds
SET min_roi_score = 10;
SET min_confidence_score = 0.5;
```

## Testing Commands

### Database Tests
```bash
# Test migration
psql -U user -d pganalytics < backend/migrations/005_ml_optimization.sql

# Verify tables
SELECT table_name FROM information_schema.tables
WHERE table_name LIKE '%workload%' OR '%rewrite%' OR '%optimization%';

# Test function
SELECT * FROM detect_workload_patterns('mydb', 30);
```

### API Tests
```bash
# Get patterns
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/workload-patterns

# Generate rewrite suggestions
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/v1/queries/12345/rewrite-suggestions/generate

# Get recommendations
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/optimization-recommendations?limit=20&min_impact=5"
```

### Build & Test
```bash
# Build backend
cd backend && go build -o pganalytics-api cmd/pganalytics-api/main.go

# Run tests
go test ./...

# Run specific test
go test -v -run TestHandleGetOptimizationRecommendations ./internal/api/
```

## File Locations

### Code Files
- Migration: `backend/migrations/005_ml_optimization.sql`
- Handlers: `backend/internal/api/handlers_ml.go`
- Storage: `backend/internal/storage/postgres.go` (end of file)
- Models: `backend/pkg/models/models.go` (end of file)
- Routes: `backend/internal/api/server.go` (line 177-207)

### Documentation
- Implementation Plan: `PHASE_4_5_IMPLEMENTATION_PLAN.md`
- Foundation Complete: `PHASE_4_5_FOUNDATION_COMPLETE.md`
- Quick Reference: `PHASE_4_5_QUICK_REFERENCE.md` (this file)

## Debugging Tips

### Check Database Connection
```go
if !p.postgres.Health(ctx) {
    // Database connection failed
}
```

### Log Query Details
```go
s.logger.Infof("Query: %s, Args: %v", query, args)
```

### Test Handler in Isolation
```go
req, _ := http.NewRequest("GET", "/api/v1/optimization-recommendations", nil)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)
// Check w.Code and w.Body
```

### Check Model Serialization
```go
data, _ := json.Marshal(recommendation)
fmt.Println(string(data))
```

## Performance Considerations

### Query Optimization
- Always use indexes: roi_score, query_hash, confidence_score
- LIMIT results in handlers (max 1000)
- Use prepared statements (already done with parameterized queries)

### Caching Opportunities
- Cache pattern detection results (30-day window)
- Cache model predictions (1-hour TTL)
- Cache top recommendations (5-minute TTL)

### Scaling Considerations
- Pattern detection can run as background job
- Model training should be async
- API calls to ML service should have timeouts
- Use connection pooling for database

## Common Errors & Fixes

### "UNIQUE constraint violation on query_hash, suggestion_type"
- Suggestion already exists for query
- Check dismissed flag before creating new suggestion
- Or update existing suggestion instead of insert

### "Foreign key violation: query_hash not found"
- Query doesn't exist in metrics_pg_stats_query
- Ensure query has been executed and metrics collected
- Check database name matches

### "ML service unavailable"
- Check ML_SERVICE_URL configuration
- Verify ML service is running
- Check circuit breaker: should fallback gracefully
- Log error and return empty prediction

### "NULL value not allowed in required field"
- Check model struct tags
- Ensure optional fields use pointers (*type)
- Verify database insert includes all required columns

## Next Phase Integration

### Phase 4.5.1: Workload Pattern Detection
- Implement `detect_workload_patterns()` SQL function details
- Add time-series bucketing logic
- Add autocorrelation algorithm

### Phase 4.5.2: Query Rewrite Suggestions
- Implement EXPLAIN plan parsing
- Add anti-pattern detection rules
- Add template library for rewrites

### Phase 4.5.3: Parameter Optimization
- Implement parameter correlation analysis
- Add recommendation rules
- Add confidence calculation

### Phase 4.5.4: ML-Powered Workflow
- Implement recommendation aggregation
- Add learning loop: predict vs actual
- Dashboard integration

### Phase 4.5.5: Python ML Service
- Create Flask app
- Implement model training
- Deploy as Docker container

### Phase 4.5.6: Predictive Modeling
- Implement `PredictQueryPerformance()` stub
- Add HTTP client for ML service
- Add confidence interval calculation

---

**Last Updated**: February 20, 2026
**Status**: Foundation Complete ✅
**Next Phase**: 4.5.1 Workload Pattern Detection
